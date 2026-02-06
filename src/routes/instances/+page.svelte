<script>
    import { onMount, onDestroy } from "svelte";
    import { askConfirm } from "$lib/stores/confirm.js";
    import { askInput } from "$lib/stores/input.js";
    import { addToast } from "$lib/stores/toast.js";
    import {
        instances,
        folders,
        loading,
        actions,
    } from "$lib/stores/instances.js";
    import {
        Folder,
        HardDrive,
        LayoutGrid,
        List,
        Plus,
        Search,
        ArrowLeft,
        MoreVertical,
        Trash2,
    } from "lucide-svelte";

    /** @type {ReturnType<typeof setInterval> | undefined} */
    let pollInterval;

    // View State
    /** @type {string | null} */
    let currentFolderId = null; // null represents Root
    let viewMode = "grid"; // 'grid' | 'list'
    let searchQuery = "";

    /** @type {string | null} */
    let draggingInstanceId = null;

    onMount(() => {
        actions.load(); // First load is loud (shows spinner)
        pollInterval = setInterval(() => actions.load(true), 2000); // subsequent are silent
    });

    onDestroy(() => {
        if (pollInterval) clearInterval(pollInterval);
    });

    // Computed
    $: currentFolder = $folders.find((f) => f.id === currentFolderId);

    // Breadcrumbs
    $: breadcrumbs = [
        { id: null, name: "Home" },
        ...(currentFolder
            ? [{ id: currentFolder.id, name: currentFolder.name }]
            : []),
    ];

    // Filtered Items (Folders & Instances)
    $: displayedFolders =
        currentFolderId === null
            ? $folders.filter((f) =>
                  f.name.toLowerCase().includes(searchQuery.toLowerCase()),
              )
            : [];

    $: displayedInstances = $instances.filter((i) => {
        const matchesSearch =
            i.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            i.id.includes(searchQuery);
        const matchesFolder =
            currentFolderId === null
                ? !i.folderId || !$folders.find((f) => f.id === i.folderId) // Uncategorized if at root
                : i.folderId === currentFolderId;

        return matchesSearch && matchesFolder;
    });

    // Actions Wrapper
    async function createFolder() {
        const name = await askInput({
            title: "Create Folder",
            placeholder: "Folder Name",
            confirmText: "Create",
        });
        if (name) actions.createFolder(name);
    }

    /** @param {string} id */
    async function deleteFolder(id) {
        if (
            await askConfirm({
                title: "Delete Folder",
                message:
                    "Delete this folder? Instances inside will be moved to Home.",
                dangerous: true,
                confirmText: "Delete",
            })
        ) {
            actions.deleteFolder(id);
        }
    }

    /** @param {string} id */
    async function deleteInstance(id) {
        if (
            await askConfirm({
                title: "Delete Instance",
                message:
                    "Are you sure? This will permanently delete all files.",
                confirmText: "Delete",
                dangerous: true,
            })
        ) {
            try {
                const res = await fetch(`/api/instances/${id}`, {
                    method: "DELETE",
                });
                if (res.ok) {
                    addToast("Instance deleted", "success");
                    actions.load();
                } else {
                    addToast("Failed to delete", "error");
                }
            } catch (e) {
                addToast("Error deleting instance", "error");
            }
        }
    }

    /**
     * @param {string} id
     * @param {string} action
     */
    async function triggerInstanceAction(id, action) {
        try {
            const res = await fetch(`/api/instances/${id}/${action}`, {
                method: "POST",
            });
            if (!res.ok) throw new Error(await res.text());
            addToast(`Instance ${action}ed`, "success");
            actions.load();
        } catch (e) {
            addToast(`Failed to ${action}`, "error");
        }
    }

    // Drag & Drop
    /**
     * @param {DragEvent} event
     * @param {string} id
     */
    function handleDragStart(event, id) {
        draggingInstanceId = id;
        if (event.dataTransfer) {
            event.dataTransfer.effectAllowed = "move";
            event.dataTransfer.setData("text/plain", id);
        }
    }

    /**
     * @param {DragEvent} event
     * @param {string} folderId
     */
    function handleDropOnFolder(event, folderId) {
        event.preventDefault();
        const id = event.dataTransfer?.getData("text/plain");
        if (id && id !== draggingInstanceId) return; // Basic check

        if (id) {
            actions.moveInstance(id, folderId);
            draggingInstanceId = null;
        }
    }

    /** @param {DragEvent} event */
    function handleDropOnRoot(event) {
        event.preventDefault();
        const id = event.dataTransfer?.getData("text/plain");
        if (id) {
            actions.moveInstance(id, ""); // "" for root/uncategorized
            draggingInstanceId = null;
        }
    }
</script>

<div class="h-full flex flex-col bg-[#0b0e14] text-gray-100">
    <!-- Header / Toolbar -->
    <header
        class="flex items-center justify-between px-6 py-4 border-b border-white/5 bg-gray-900/50 backdrop-blur-md sticky top-0 z-20"
    >
        <div class="flex items-center gap-4">
            <h1
                class="text-xl font-bold tracking-tight flex items-center gap-2"
            >
                <HardDrive class="w-6 h-6 text-indigo-500" />
                Storage
            </h1>

            <!-- Breadcrumbs -->
            <div
                class="flex items-center gap-2 text-sm text-gray-400 bg-black/20 px-3 py-1.5 rounded-lg border border-white/5"
            >
                {#each breadcrumbs as crumb, i}
                    <button
                        class="hover:text-white transition-colors {i ===
                        breadcrumbs.length - 1
                            ? 'font-semibold text-white'
                            : ''}"
                        on:click={() => (currentFolderId = crumb.id)}
                    >
                        {crumb.name}
                    </button>
                    {#if i < breadcrumbs.length - 1}
                        <span class="text-gray-600">/</span>
                    {/if}
                {/each}
            </div>
        </div>

        <div class="flex items-center gap-3">
            <!-- Search -->
            <div class="relative group">
                <Search
                    class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-500 group-focus-within:text-indigo-400 transition-colors"
                />
                <input
                    type="text"
                    placeholder="Search..."
                    bind:value={searchQuery}
                    class="bg-black/20 border border-white/5 rounded-lg pl-9 pr-4 py-2 text-sm focus:outline-none focus:border-indigo-500/50 focus:ring-1 focus:ring-indigo-500/20 w-64 transition-all"
                />
            </div>

            <div class="h-6 w-px bg-white/10 mx-1"></div>

            <button
                on:click={createFolder}
                class="bg-white/5 hover:bg-white/10 text-gray-300 hover:text-white px-3 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2 border border-white/5"
            >
                <Folder class="w-4 h-4" />
                New Folder
            </button>

            <a
                href="/instances/create"
                class="bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg text-sm font-bold transition-all shadow-lg shadow-indigo-500/20 flex items-center gap-2"
            >
                <Plus class="w-4 h-4" />
                New Instance
            </a>
        </div>
    </header>

    <!-- Main Content -->
    <main
        class="flex-1 overflow-y-auto p-6"
        on:dragover|preventDefault
        on:drop={(e) => {
            // Only handle drop on background if we are inside a folder to move it "out" ??
            // Actually, maybe better to have a specific drop zone or just rely on sidebar if we had one.
            // For now, dropping on the background of Root View does nothing specific unless we implement "move to root".
            // Let's implement move to root if dropping on empty space in Root View? No, usually drag to "Home" breadcrumb.
            // Let's allow dropping on "Home" breadcrumb.
        }}
    >
        {#if $loading && $instances.length === 0}
            <div
                class="h-full flex flex-col items-center justify-center text-gray-500 animate-pulse"
            >
                <HardDrive class="w-12 h-12 mb-4 opacity-20" />
                <p>Loading your digital universe...</p>
            </div>
        {:else}
            <!-- FOLDERS SECTION -->
            {#if displayedFolders.length > 0}
                <div class="mb-8">
                    <h2
                        class="text-xs font-bold text-gray-500 uppercase tracking-widest mb-4 flex items-center gap-2"
                    >
                        <Folder class="w-3 h-3" /> Folders
                    </h2>
                    <div
                        class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4"
                    >
                        {#each displayedFolders as folder (folder.id)}
                            <div
                                class="group bg-gray-800/40 hover:bg-gray-800/80 border border-white/5 hover:border-indigo-500/30 rounded-xl p-4 cursor-pointer transition-all hover:shadow-lg hover:shadow-indigo-500/10 flex flex-col justify-between aspect-square md:aspect-auto md:h-32 relative"
                                on:click={() => (currentFolderId = folder.id)}
                                on:dragover|preventDefault
                                on:drop|stopPropagation={(e) =>
                                    handleDropOnFolder(e, folder.id)}
                                role="button"
                                tabindex="0"
                                on:keydown={(e) =>
                                    e.key === "Enter" &&
                                    (currentFolderId = folder.id)}
                            >
                                <div class="flex justify-between items-start">
                                    <Folder
                                        class="w-8 h-8 text-indigo-400 group-hover:text-indigo-300 transition-colors"
                                    />
                                    <button
                                        class="p-1 hover:bg-white/10 rounded text-gray-500 hover:text-white transition-colors opacity-0 group-hover:opacity-100"
                                        on:click|stopPropagation={() =>
                                            deleteFolder(folder.id)}
                                        title="Delete Folder"
                                    >
                                        <Trash2 class="w-4 h-4" />
                                    </button>
                                </div>
                                <div>
                                    <div
                                        class="font-medium text-gray-200 truncate group-hover:text-white text-sm"
                                    >
                                        {folder.name}
                                    </div>
                                    <div class="text-xs text-gray-600 mt-1">
                                        {$instances.filter(
                                            (i) => i.folderId === folder.id,
                                        ).length} items
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </div>
            {/if}

            <!-- INSTANCES SECTION -->
            <div>
                <h2
                    class="text-xs font-bold text-gray-500 uppercase tracking-widest mb-4 flex items-center gap-2"
                >
                    <LayoutGrid class="w-3 h-3" /> Instances
                </h2>

                {#if displayedInstances.length === 0}
                    <div
                        class="flex flex-col items-center justify-center py-20 border-2 border-dashed border-white/5 rounded-2xl bg-white/[0.02]"
                    >
                        <HardDrive class="w-12 h-12 text-gray-700 mb-4" />
                        <p class="text-gray-500 font-medium">
                            No instances found here
                        </p>
                        {#if currentFolderId === null}
                            <a
                                href="/instances/create"
                                class="mt-4 text-indigo-400 hover:text-indigo-300 text-sm"
                                >Create one now &rarr;</a
                            >
                        {/if}
                    </div>
                {:else}
                    <div
                        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4"
                    >
                        {#each displayedInstances as inst (inst.id)}
                            <div
                                class="group bg-gray-900/40 hover:bg-gray-800/60 border border-white/5 hover:border-white/10 rounded-xl overflow-hidden transition-all hover:shadow-xl hover:shadow-black/20 flex flex-col"
                                draggable="true"
                                on:dragstart={(e) =>
                                    handleDragStart(e, inst.id)}
                            >
                                <!-- Status Strip -->
                                <div
                                    class="h-1 w-full bg-gray-800 relative overflow-hidden"
                                >
                                    {#if inst.status === "Online"}
                                        <div
                                            class="absolute inset-0 bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.5)]"
                                        ></div>
                                    {:else if inst.status === "Starting"}
                                        <div
                                            class="absolute inset-0 bg-amber-500 animate-pulse"
                                        ></div>
                                    {/if}
                                </div>

                                <div class="p-4 flex-1 flex flex-col">
                                    <div
                                        class="flex justify-between items-start mb-3"
                                    >
                                        <div>
                                            <a
                                                href="/instances/{inst.id}"
                                                class="font-bold text-gray-200 hover:text-white transition-colors line-clamp-1 block"
                                                title={inst.name}
                                            >
                                                {inst.name}
                                            </a>
                                            <div
                                                class="flex items-center gap-2 mt-1"
                                            >
                                                <span
                                                    class="text-[10px] uppercase font-bold px-1.5 py-0.5 rounded bg-white/5 text-gray-400 border border-white/5"
                                                >
                                                    {inst.type}
                                                </span>
                                                <span
                                                    class="text-[10px] text-gray-600 font-mono"
                                                >
                                                    {inst.version}
                                                </span>
                                            </div>
                                        </div>

                                        <div class="relative">
                                            {#if inst.status === "Online"}
                                                <div
                                                    class="w-2 h-2 rounded-full bg-emerald-500 shadow-lg shadow-emerald-500/50 animate-pulse"
                                                ></div>
                                            {:else}
                                                <div
                                                    class="w-2 h-2 rounded-full bg-gray-700"
                                                ></div>
                                            {/if}
                                        </div>
                                    </div>

                                    <div
                                        class="mt-auto pt-4 flex items-center justify-between border-t border-white/5 opacity-60 group-hover:opacity-100 transition-opacity"
                                    >
                                        <div class="flex gap-2">
                                            {#if inst.status === "Online" || inst.status === "Starting"}
                                                <button
                                                    on:click={() =>
                                                        triggerInstanceAction(
                                                            inst.id,
                                                            "stop",
                                                        )}
                                                    class="p-1.5 hover:bg-red-500/20 text-red-400 rounded-md transition-colors"
                                                    title="Stop"
                                                >
                                                    <div
                                                        class="w-3 h-3 bg-current rounded-xs"
                                                    ></div>
                                                </button>
                                            {:else}
                                                <button
                                                    on:click={() =>
                                                        triggerInstanceAction(
                                                            inst.id,
                                                            "start",
                                                        )}
                                                    class="p-1.5 hover:bg-emerald-500/20 text-emerald-400 rounded-md transition-colors"
                                                    title="Start"
                                                >
                                                    <svg
                                                        class="w-3 h-3 fill-current"
                                                        viewBox="0 0 24 24"
                                                        ><path
                                                            d="M8 5v14l11-7z"
                                                        /></svg
                                                    >
                                                </button>
                                            {/if}
                                        </div>

                                        <div class="flex gap-2">
                                            <button
                                                on:click={() =>
                                                    deleteInstance(inst.id)}
                                                class="text-xs font-medium text-red-400 hover:text-red-300 px-2 py-1 rounded hover:bg-red-500/10 transition-colors"
                                                title="Delete Instance"
                                            >
                                                Delete
                                            </button>
                                            <a
                                                href="/instances/{inst.id}"
                                                class="text-xs font-medium text-gray-400 hover:text-white px-2 py-1 rounded hover:bg-white/5 transition-colors"
                                            >
                                                Manage
                                            </a>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                {/if}
            </div>
        {/if}
    </main>
</div>
