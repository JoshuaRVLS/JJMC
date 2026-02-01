<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";

    /** @type {string} */
    export let instanceId;
    export let type = "";
    export let mode = "mod"; // "mod" or "plugin"

    let activeTab = mode === "plugin" ? "plugin" : "mod"; // 'mod', 'plugin', 'modpack'

    // React to mode changes
    $: if (mode) {
        if (mode === "plugin" && activeTab !== "plugin") {
            activeTab = "plugin";
        } else if (
            mode === "mod" &&
            activeTab !== "mod" &&
            activeTab !== "modpack"
        ) {
            activeTab = "mod";
        }
    }

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

    /** @type {string | null} */
    let viewingVersionsId = null;
    /** @type {Array<any>} */
    let versionsList = [];
    let loadingVersions = false;

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

    /** @param {string} projectId */
    async function fetchVersions(projectId) {
        loadingVersions = true;
        versionsList = [];
        try {
            let typeParam = activeTab === "plugin" ? "plugin" : "mod";
            // If viewing modpack, treat as mod for now or disable?
            // Modpack versions are usually just the pack versions.
            // Modrinth "project_type" handles this.

            const res = await fetch(
                `/api/instances/${instanceId}/mods/${projectId}/versions?type=${typeParam}`,
            );
            if (res.ok) {
                versionsList = await res.json();
            } else {
                addToast("Failed to load versions", "error");
            }
        } catch (e) {
            addToast(
                "Error loading versions: " + /** @type {Error} */ (e).message,
                "error",
            );
        } finally {
            loadingVersions = false;
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

        // Map activeTab to backend 'type' param
        // If activeTab is 'mod', send 'mod'
        // If activeTab is 'modpack', send 'modpack'
        // If activeTab is 'plugin', send 'plugin'
        let typeParam = activeTab === "plugin" ? "plugin" : activeTab;
        if (typeParam === "mod" && mode === "plugin") typeParam = "plugin"; // fallback

        try {
            const res = await fetch(
                `/api/instances/${instanceId}/mods/search?query=${encodeURIComponent(query)}&type=${typeParam}&offset=${offset}&sort=${sortBy}`,
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
            const msg = e instanceof Error ? e.message : String(e);
            addToast("Error searching: " + msg, "error");
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
        // Reset version view when changing context
        viewingVersionsId = null;
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
    /** @param {string} [versionId] */
    async function installMod(projectId, versionId = "") {
        installingId = projectId; // We use project ID for loading state on card
        try {
            let typeParam = activeTab === "plugin" ? "plugin" : "mod";

            const res = await fetch(`/api/instances/${instanceId}/mods`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    projectId,
                    resourceType: typeParam,
                    versionId,
                }),
            });
            if (res.ok) {
                addToast("Installed successfully", "success");
                fetchInstalled();
                // If installed specific version, maybe close version list?
                // viewingVersionsId = null;
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
            let typeParam = activeTab === "plugin" ? "plugin" : "mod";

            const res = await fetch(`/api/instances/${instanceId}/mods`, {
                method: "DELETE",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    project_id: projectId,
                    resource_type: typeParam,
                }),
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
            {#if mode === "plugin"}
                <button
                    class="px-4 py-1.5 rounded-md text-sm font-medium transition-all bg-indigo-500 text-white shadow-lg"
                    on:click={() => {}}
                >
                    Plugins
                </button>
            {:else}
                <button
                    class="px-4 py-1.5 rounded-md text-sm font-medium transition-all {activeTab ===
                    'mod'
                        ? 'bg-indigo-500 text-white shadow-lg'
                        : 'text-gray-400 hover:text-white'}"
                    on:click={() => {
                        activeTab = "mod";
                    }}
                >
                    Mods
                </button>
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
                                        {#if activeTab === "mod" || activeTab === "plugin"}
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

                        <!-- Versions Modal -->
                        {#if viewingVersionsId === item.project_id}
                            <div
                                class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm animate-in fade-in duration-200"
                                role="button"
                                tabindex="0"
                                on:click|stopPropagation={() =>
                                    (viewingVersionsId = null)}
                                on:keydown={(e) =>
                                    e.key === "Escape" &&
                                    (viewingVersionsId = null)}
                                aria-label="Close modal"
                            >
                                <div
                                    class="bg-gray-900 border border-white/10 rounded-2xl shadow-2xl w-full max-w-lg max-h-[80vh] flex flex-col m-4 animate-in zoom-in-95 duration-200 cursor-auto"
                                    role="dialog"
                                    aria-modal="true"
                                    on:click|stopPropagation
                                    on:keydown|stopPropagation
                                >
                                    <!-- Header -->
                                    <div
                                        class="p-4 border-b border-white/10 flex justify-between items-center bg-black/20 rounded-t-2xl"
                                    >
                                        <div>
                                            <h3
                                                class="text-white font-bold text-lg"
                                            >
                                                {item.title}
                                            </h3>
                                            <p class="text-gray-400 text-xs">
                                                Select a version to install
                                            </p>
                                        </div>
                                        <button
                                            class="text-gray-400 hover:text-white bg-white/5 hover:bg-white/10 p-2 rounded-lg transition-colors"
                                            aria-label="Close versions"
                                            on:click={() =>
                                                (viewingVersionsId = null)}
                                        >
                                            <svg
                                                class="w-5 h-5"
                                                fill="none"
                                                viewBox="0 0 24 24"
                                                stroke="currentColor"
                                            >
                                                <path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M6 18L18 6M6 6l12 12"
                                                />
                                            </svg>
                                        </button>
                                    </div>

                                    <!-- Version List -->
                                    <div
                                        class="flex-1 overflow-y-auto custom-scrollbar p-2 space-y-2"
                                    >
                                        {#if loadingVersions}
                                            <div
                                                class="flex flex-col items-center justify-center py-12 text-gray-400"
                                            >
                                                <svg
                                                    class="animate-spin w-8 h-8 text-indigo-500 mb-3"
                                                    fill="none"
                                                    viewBox="0 0 24 24"
                                                >
                                                    <circle
                                                        class="opacity-25"
                                                        cx="12"
                                                        cy="12"
                                                        r="10"
                                                        stroke="currentColor"
                                                        stroke-width="4"
                                                    ></circle>
                                                    <path
                                                        class="opacity-75"
                                                        fill="currentColor"
                                                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                                                    ></path>
                                                </svg>
                                                <span
                                                    class="text-sm font-medium"
                                                    >Loading versions...</span
                                                >
                                            </div>
                                        {:else if versionsList.length === 0}
                                            <div
                                                class="flex flex-col items-center justify-center py-12 text-gray-500"
                                            >
                                                <svg
                                                    class="w-12 h-12 mb-3 bg-white/5 p-2 rounded-xl"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    viewBox="0 0 24 24"
                                                >
                                                    <path
                                                        stroke-linecap="round"
                                                        stroke-linejoin="round"
                                                        stroke-width="1.5"
                                                        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                                                    />
                                                </svg>
                                                <p class="text-sm font-medium">
                                                    No compatible versions
                                                    found.
                                                </p>
                                                <p class="text-xs mt-1">
                                                    Try changing your Minecraft
                                                    version filter.
                                                </p>
                                            </div>
                                        {:else}
                                            {#each versionsList as ver}
                                                <div
                                                    class="group flex items-center justify-between p-3 rounded-xl bg-black/20 hover:bg-indigo-500/10 border border-white/5 hover:border-indigo-500/30 transition-all"
                                                >
                                                    <div
                                                        class="min-w-0 flex-1 mr-4"
                                                    >
                                                        <div
                                                            class="flex items-center gap-2 mb-0.5"
                                                        >
                                                            <span
                                                                class="text-sm font-bold text-gray-200 truncate"
                                                                >{ver.version_number}</span
                                                            >
                                                            <span
                                                                class="text-[10px] px-1.5 py-0.5 rounded bg-white/10 text-gray-400 capitalize"
                                                                >{ver.version_type ||
                                                                    "release"}</span
                                                            >
                                                        </div>
                                                        <div
                                                            class="flex items-center gap-2 text-xs text-gray-500"
                                                        >
                                                            <span
                                                                >{new Date(
                                                                    ver.date_published ||
                                                                        Date.now(),
                                                                ).toLocaleDateString(
                                                                    undefined,
                                                                    {
                                                                        year: "numeric",
                                                                        month: "short",
                                                                        day: "numeric",
                                                                    },
                                                                )}</span
                                                            >
                                                            {#if ver.loaders}
                                                                <span>•</span>
                                                                <span
                                                                    class="truncate max-w-[150px]"
                                                                    >{ver.loaders.join(
                                                                        ", ",
                                                                    )}</span
                                                                >
                                                            {/if}
                                                        </div>
                                                    </div>

                                                    <button
                                                        on:click|stopPropagation={() =>
                                                            installMod(
                                                                item.project_id,
                                                                ver.id,
                                                            )}
                                                        class="shrink-0 bg-white/5 hover:bg-indigo-600 hover:text-white text-gray-300 px-4 py-2 rounded-lg text-xs font-bold transition-all flex items-center gap-2 group-hover:shadow-lg group-hover:shadow-indigo-500/20"
                                                    >
                                                        <svg
                                                            class="w-4 h-4"
                                                            fill="none"
                                                            stroke="currentColor"
                                                            viewBox="0 0 24 24"
                                                        >
                                                            <path
                                                                stroke-linecap="round"
                                                                stroke-linejoin="round"
                                                                stroke-width="2"
                                                                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                                                            />
                                                        </svg>
                                                        Install
                                                    </button>
                                                </div>
                                            {/each}
                                        {/if}
                                    </div>

                                    <!-- Footer -->
                                    <div
                                        class="p-3 bg-black/40 text-center text-[10px] text-gray-500 border-t border-white/5 rounded-b-2xl"
                                    >
                                        Showing compatible versions for {activeTab ===
                                        "plugin"
                                            ? "Plugins"
                                            : "Mods"}
                                    </div>
                                </div>
                            </div>
                        {/if}

                        <!-- Install Button -->
                        <div
                            class="mt-auto flex justify-between items-center bg-black/20 -mx-4 -mb-4 px-4 py-2.5 border-t border-white/5"
                        >
                            <span class="text-[10px] text-gray-500"
                                >Updated {new Date(
                                    item.date_modified,
                                ).toLocaleDateString()}</span
                            >
                            <div class="flex gap-2">
                                {#if activeTab !== "modpack"}
                                    <button
                                        on:click|stopPropagation={() => {
                                            if (
                                                viewingVersionsId ===
                                                item.project_id
                                            ) {
                                                viewingVersionsId = null;
                                            } else {
                                                viewingVersionsId =
                                                    item.project_id;
                                                fetchVersions(item.project_id);
                                            }
                                        }}
                                        class="bg-white/5 hover:bg-white/10 text-gray-300 px-3 py-1.5 rounded-lg text-xs font-bold transition-all border border-white/5"
                                    >
                                        Versions
                                    </button>
                                {/if}

                                <button
                                    on:click|stopPropagation={() => {
                                        if (
                                            (activeTab === "mod" ||
                                                activeTab === "plugin") &&
                                            installedIds.has(item.project_id)
                                        ) {
                                            uninstallMod(item.project_id);
                                        } else if (activeTab === "modpack") {
                                            installModpack(item.project_id);
                                        } else {
                                            installMod(item.project_id);
                                        }
                                    }}
                                    disabled={installingId ===
                                        item.project_id ||
                                        installingId !== null}
                                    class="flex items-center gap-2 {(activeTab ===
                                        'mod' ||
                                        activeTab === 'plugin') &&
                                    installedIds.has(item.project_id)
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
                                    {:else if (activeTab === "mod" || activeTab === "plugin") && installedIds.has(item.project_id)}
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
