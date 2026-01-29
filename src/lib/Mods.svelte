<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";

    /** @type {string} */
    export let instanceId;
    export let type = "";

    let activeTab = "mod"; // 'mod' | 'modpack'
    let query = "";
    /** @type {Array<any>} */
    let results = [];
    /** @type {Set<string>} */
    let installedIds = new Set();
    let loading = false;
    let loadingMore = false;
    /** @type {string | null} */
    let installingId = null;
    let offset = 0;
    let hasMore = true;
    let sortBy = "relevance"; // relevance, downloads, follows, newest, updated

    /** @type {IntersectionObserver} */
    let observer;
    /** @type {HTMLElement} */
    let sentinel;

    async function fetchInstalled() {
        try {
            const res = await fetch(`/api/instances/${instanceId}/mods`);
            if (res.ok) {
                const ids = await res.json();
                installedIds = new Set(ids);
            }
        } catch (e) {
            console.error("Failed to fetch installed mods:", e);
        }
    }

    async function search(isNew = true) {
        if (isNew) {
            loading = true;
            offset = 0;
            results = [];
            hasMore = true;
        } else {
            loadingMore = true;
        }

        try {
            const res = await fetch(
                `/api/instances/${instanceId}/mods/search?query=${encodeURIComponent(query)}&type=${activeTab}&offset=${offset}&sort=${sortBy}`,
            );
            if (res.ok) {
                const data = await res.json();
                if (isNew) {
                    results = data || [];
                } else {
                    results = [...results, ...(data || [])];
                }

                if (!data || data.length < 20) {
                    hasMore = false;
                }
            } else {
                const err = await res.json();
                addToast(
                    "Search failed: " + (err.error || "Unknown error"),
                    "error",
                );
            }
        } catch (e) {
            addToast("Error searching: " + e.message, "error");
        } finally {
            loading = false;
            loadingMore = false;
        }
    }

    function loadMore() {
        if (loading || loadingMore || !hasMore) return;
        offset += 20;
        search(false);
    }

    // Trigger search when tab or sort changes
    $: if (activeTab || sortBy) {
        search(true);
    }

    onMount(() => {
        fetchInstalled();

        observer = new IntersectionObserver(
            (entries) => {
                if (entries[0].isIntersecting) {
                    loadMore();
                }
            },
            { threshold: 0.1 },
        );

        if (sentinel) observer.observe(sentinel);

        return () => {
            if (observer) observer.disconnect();
        };
    });

    /** @param {string} projectId */
    async function installMod(projectId) {
        installingId = projectId;
        try {
            const res = await fetch(`/api/instances/${instanceId}/mods`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ projectId }),
            });
            if (res.ok) {
                addToast("Mod installed successfully", "success");
                fetchInstalled();
            } else {
                const err = await res.json();
                addToast(
                    "Install failed: " + (err.error || "Unknown error"),
                    "error",
                );
            }
        } catch (e) {
            addToast(
                "Error installing: " + /** @type {Error} */ (e).message,
                "error",
            );
        } finally {
            installingId = null;
        }
    }

    /** @param {string} projectId */
    async function uninstallMod(projectId) {
        if (!confirm("Are you sure you want to uninstall this?")) return;
        installingId = projectId; // Re-using state var for loading
        try {
            const res = await fetch(`/api/instances/${instanceId}/mods`, {
                method: "DELETE",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ project_id: projectId }),
            });
            if (res.ok) {
                addToast("Uninstalled successfully", "success");
                fetchInstalled();
            } else {
                const err = await res.json();
                addToast(
                    "Uninstall failed: " + (err.error || "Unknown error"),
                    "error",
                );
            }
        } catch (e) {
            addToast(
                "Error uninstalling: " + /** @type {Error} */ (e).message,
                "error",
            );
        } finally {
            installingId = null;
        }
    }

    /** @param {string} projectId */
    async function installModpack(projectId) {
        if (
            !confirm(
                "Warning: Installing a modpack will DELETE all current mods in the 'mods' folder. Continue?",
            )
        ) {
            return;
        }

        installingId = projectId;
        try {
            const res = await fetch(`/api/instances/${instanceId}/modpacks`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ projectId }),
            });
            if (res.ok) {
                addToast(
                    "Modpack installation started. Check console for progress.",
                    "success",
                );
            } else {
                const err = await res.json();
                addToast(
                    "Install failed: " + (err.error || "Unknown error"),
                    "error",
                );
            }
        } catch (e) {
            addToast(
                "Error installing: " + /** @type {Error} */ (e).message,
                "error",
            );
        } finally {
            installingId = null;
        }
    }

    /** @param {KeyboardEvent} e */
    function handleKeydown(e) {
        if (e.key === "Enter") search(true);
    }

    /** @param {string} side */
    function getSideColor(side) {
        if (side === "required")
            return "bg-rose-500/20 text-rose-400 border-rose-500/30";
        if (side === "optional")
            return "bg-amber-500/20 text-amber-400 border-amber-500/30";
        return "bg-gray-500/20 text-gray-400 border-gray-500/30";
    }
</script>

<div class="h-full flex flex-col">
    <!-- Header / Tabs / Sorting -->
    <div
        class="flex flex-col md:flex-row md:items-center justify-between mb-4 gap-4 px-1"
    >
        <div class="flex bg-black/20 p-1 rounded-lg self-start">
            <button
                class="px-4 py-1.5 rounded-md text-sm font-medium transition-all {activeTab ===
                'mod'
                    ? 'bg-indigo-500 text-white shadow-lg'
                    : 'text-gray-400 hover:text-white'}"
                on:click={() => {
                    activeTab = "mod";
                }}
            >
                {type === "spigot" ? "Plugins" : "Mods"}
            </button>
            {#if type !== "spigot"}
                <button
                    class="px-4 py-1.5 rounded-md text-sm font-medium transition-all {activeTab ===
                    'modpack'
                        ? 'bg-indigo-500 text-white shadow-lg'
                        : 'text-gray-400 hover:text-white'}"
                    on:click={() => {
                        activeTab = "modpack";
                    }}
                >
                    Modpacks
                </button>
            {/if}
        </div>

        <div class="flex items-center gap-2">
            <span
                class="text-xs text-gray-500 font-medium uppercase tracking-wider"
                >Sort by:</span
            >
            <select
                bind:value={sortBy}
                class="bg-black/20 border border-white/10 rounded-lg px-2 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-indigo-500 transition-all"
            >
                <option value="relevance">Relevance</option>
                <option value="downloads">Downloads</option>
                <option value="follows">Follows</option>
                <option value="newest">Newest</option>
                <option value="updated">Recently Updated</option>
            </select>
        </div>
    </div>

    <!-- Search Bar -->
    <div class="flex gap-2 mb-6">
        <div class="relative flex-1">
            <input
                type="text"
                bind:value={query}
                on:keydown={handleKeydown}
                placeholder={activeTab === "mod"
                    ? type === "spigot"
                        ? "Search for plugins..."
                        : "Search for mods..."
                    : "Search for modpacks..."}
                class="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-3 pl-11 text-white focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all"
            />
            <svg
                class="w-5 h-5 text-gray-500 absolute left-3.5 top-3.5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                />
            </svg>
        </div>
        <button
            on:click={() => search(true)}
            disabled={loading}
            class="bg-indigo-600 hover:bg-indigo-500 text-white px-6 rounded-xl font-medium transition-colors disabled:opacity-50"
        >
            {loading ? "Searching..." : "Search"}
        </button>
    </div>

    <!-- Results Grid -->
    <div class="flex-1 overflow-y-auto pr-2 custom-scrollbar">
        {#if loading && results.length === 0}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {#each Array(6) as _}
                    <div
                        class="bg-white/5 h-48 rounded-xl animate-pulse border border-white/5"
                    ></div>
                {/each}
            </div>
        {:else if results.length > 0}
            <div
                class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-4 pb-4"
            >
                {#each results as item}
                    <div
                        class="bg-gray-900/40 border border-white/5 hover:border-indigo-500/30 rounded-xl p-4 transition-all group relative overflow-hidden flex flex-col"
                    >
                        <div class="flex gap-4 mb-3">
                            <!-- Icon -->
                            <div class="shrink-0">
                                {#if item.icon_url}
                                    <img
                                        src={item.icon_url}
                                        alt={item.title}
                                        class="w-16 h-16 rounded-lg object-cover bg-black/30"
                                    />
                                {:else}
                                    <div
                                        class="w-16 h-16 rounded-lg bg-indigo-500/20 flex items-center justify-center text-indigo-400"
                                    >
                                        {#if activeTab === "mod"}
                                            <svg
                                                class="w-8 h-8"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                                ><path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"
                                                /></svg
                                            >
                                        {:else}
                                            <svg
                                                class="w-8 h-8"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                                ><path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"
                                                /></svg
                                            >
                                        {/if}
                                    </div>
                                {/if}
                            </div>

                            <!-- Content -->
                            <div class="flex-1 min-w-0">
                                <div class="flex justify-between items-start">
                                    <h3
                                        class="text-white font-bold truncate pr-2 text-sm"
                                    >
                                        {item.title}
                                    </h3>
                                    <div class="flex gap-1">
                                        {#if item.client_side !== "unsupported"}
                                            <span
                                                class="text-[9px] px-1.5 py-0.5 rounded border {getSideColor(
                                                    item.client_side,
                                                )}"
                                                title="Client Side: {item.client_side}"
                                                >CL</span
                                            >
                                        {/if}
                                        {#if item.server_side !== "unsupported"}
                                            <span
                                                class="text-[9px] px-1.5 py-0.5 rounded border {getSideColor(
                                                    item.server_side,
                                                )}"
                                                title="Server Side: {item.server_side}"
                                                >SV</span
                                            >
                                        {/if}
                                    </div>
                                </div>
                                <div
                                    class="text-[10px] text-gray-500 mb-1 flex items-center gap-2"
                                >
                                    <span>by {item.author}</span>
                                    <span>•</span>
                                    <span
                                        class="flex items-center gap-0.5"
                                        title="{item.downloads} downloads"
                                    >
                                        <svg
                                            class="w-3 h-3"
                                            fill="none"
                                            stroke="currentColor"
                                            viewBox="0 0 24 24"
                                            ><path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                                            ></path></svg
                                        >
                                        {Intl.NumberFormat("en-US", {
                                            notation: "compact",
                                            maximumFractionDigits: 1,
                                        }).format(item.downloads)}
                                    </span>
                                </div>
                                <p
                                    class="text-gray-400 text-xs line-clamp-2 leading-relaxed h-8"
                                >
                                    {item.description}
                                </p>
                            </div>
                        </div>

                        <!-- Categories -->
                        <div class="flex flex-wrap gap-1 mb-4">
                            {#each (item.categories || []).slice(0, 3) as cat}
                                <span
                                    class="bg-white/5 text-[10px] text-gray-400 px-2 py-0.5 rounded uppercase tracking-wider border border-white/5"
                                    >{cat}</span
                                >
                            {/each}
                            {#if (item.categories || []).length > 3}
                                <span
                                    class="bg-white/5 text-[10px] text-gray-400 px-2 py-0.5 rounded border border-white/5"
                                    >+{item.categories.length - 3}</span
                                >
                            {/if}
                        </div>

                        <!-- Install Button -->
                        <div
                            class="mt-auto flex justify-between items-center bg-black/20 -mx-4 -mb-4 px-4 py-2.5 border-t border-white/5"
                        >
                            <span class="text-[10px] text-gray-500"
                                >Updated {new Date(
                                    item.date_modified,
                                ).toLocaleDateString()}</span
                            >
                            <button
                                on:click={() => {
                                    if (
                                        activeTab === "mod" &&
                                        installedIds.has(item.project_id)
                                    ) {
                                        uninstallMod(item.project_id);
                                    } else if (activeTab === "mod") {
                                        installMod(item.project_id);
                                    } else {
                                        installModpack(item.project_id);
                                    }
                                }}
                                disabled={installingId === item.project_id ||
                                    installingId !== null}
                                class="flex items-center gap-2 {activeTab ===
                                    'mod' && installedIds.has(item.project_id)
                                    ? 'bg-red-500/10 text-red-400 border border-red-500/30 hover:bg-red-500/20'
                                    : 'bg-indigo-500 hover:bg-indigo-600 text-white shadow-lg shadow-indigo-500/20'} px-3 py-1.5 rounded-lg text-xs font-bold transition-all disabled:opacity-50"
                            >
                                {#if installingId === item.project_id}
                                    <svg
                                        class="animate-spin w-3 h-3"
                                        fill="none"
                                        viewBox="0 0 24 24"
                                        ><circle
                                            class="opacity-25"
                                            cx="12"
                                            cy="12"
                                            r="10"
                                            stroke="currentColor"
                                            stroke-width="4"
                                        ></circle><path
                                            class="opacity-75"
                                            fill="currentColor"
                                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                                        ></path></svg
                                    >
                                    Processing...
                                {:else if activeTab === "mod" && installedIds.has(item.project_id)}
                                    <svg
                                        class="w-3 h-3"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                        ><path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                                        ></path></svg
                                    >
                                    Uninstall
                                {:else}
                                    <svg
                                        class="w-3 h-3"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                        ><path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                                        ></path></svg
                                    >
                                    Install
                                {/if}
                            </button>
                        </div>
                    </div>
                {/each}
            </div>

            <!-- Sentinel for Infinite Scroll -->
            <div
                bind:this={sentinel}
                class="h-10 flex items-center justify-center"
            >
                {#if loadingMore}
                    <div
                        class="flex items-center gap-2 text-gray-500 text-xs font-medium bg-black/20 px-4 py-2 rounded-full"
                    >
                        <svg class="animate-spin h-3 w-3" viewBox="0 0 24 24">
                            <circle
                                class="opacity-25"
                                cx="12"
                                cy="12"
                                r="10"
                                stroke="currentColor"
                                stroke-width="4"
                                fill="none"
                            ></circle>
                            <path
                                class="opacity-75"
                                fill="currentColor"
                                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                            ></path>
                        </svg>
                        Loading more...
                    </div>
                {:else if !hasMore && results.length > 0}
                    <span
                        class="text-xs text-gray-600 bg-black/10 px-4 py-1.5 rounded-full border border-white/5"
                        >You've reached the end of the list.</span
                    >
                {/if}
            </div>
        {:else if query && !loading}
            <div
                class="text-center text-gray-500 py-20 flex flex-col items-center gap-4"
            >
                <div class="p-6 rounded-full bg-white/5 text-gray-600">
                    <svg
                        class="w-12 h-12"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="1.5"
                            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                        />
                    </svg>
                </div>
                <div class="space-y-1">
                    <p class="text-white font-medium">No results found</p>
                    <p class="text-xs">
                        Try adjusting your search or sorting criteria.
                    </p>
                </div>
            </div>
        {/if}
    </div>
</div>
