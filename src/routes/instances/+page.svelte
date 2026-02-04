<script>
    import { onMount } from "svelte";
    import { askConfirm } from "$lib/stores/confirm.js";
    import { askInput } from "$lib/stores/input.js";
    import { addToast } from "$lib/stores/toast.js";

    import { onDestroy } from "svelte";

    /**
     * @typedef {Object} Instance
     * @property {string} id
     * @property {string} name
     * @property {string} type
     * @property {string} version
     * @property {string} status
     * @property {string} [folderId]
     * @property {number} [maxMemory]
     * @property {string} [javaArgs]
     * @property {string} [jarFile]
     * @property {string} [javaPath]
     * @property {string} [webhookUrl]
     * @property {string} [group]
     */

    /**
     * @typedef {Object} Folder
     * @property {string} id
     * @property {string} name
     */

    /** @type {Instance[]} */
    let instances = [];
    /** @type {Folder[]} */
    let folders = [];
    let loading = true;

    /** @type {ReturnType<typeof setInterval> | undefined} */
    let pollInterval;
    /** @type {string | null} */
    let draggingInstanceId = null;

    async function loadData() {
        try {
            const [instRes, folderRes] = await Promise.all([
                fetch("/api/instances"),
                fetch("/api/folders"),
            ]);

            if (instRes.ok) instances = await instRes.json();
            if (folderRes.ok) folders = await folderRes.json();
        } catch (e) {
            console.error("Failed to load data", e);
        } finally {
            loading = false;
        }
    }

    onMount(() => {
        loadData();
        pollInterval = setInterval(loadData, 2000);
    });

    onDestroy(() => {
        if (pollInterval) clearInterval(pollInterval);
    });

    async function createFolder() {
        const name = await askInput({
            title: "Create Folder",
            placeholder: "Folder Name",
            confirmText: "Create",
        });
        if (!name) return;

        try {
            const res = await fetch("/api/folders", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ name }),
            });
            if (res.ok) {
                addToast("Folder created", "success");
                loadData();
            } else {
                addToast("Failed to create folder", "error");
            }
        } catch (e) {
            /** @type {Error} */
            const err = /** @type {Error} */ (e);
            addToast("Error creating folder: " + err.message, "error");
        }
    }

    /**
     * @param {string} id
     */
    async function deleteFolder(id) {
        if (
            !(await askConfirm({
                title: "Delete Folder",
                message:
                    "Delete this folder? Instances inside will be moved to Uncategorized.",
                dangerous: true,
                confirmText: "Delete",
            }))
        )
            return;
        try {
            const res = await fetch(`/api/folders/${id}`, {
                method: "DELETE",
            });
            if (res.ok) {
                addToast("Folder deleted", "success");
                loadData();
            } else {
                addToast("Failed to delete folder", "error");
            }
        } catch (e) {
            /** @type {Error} */
            const err = /** @type {Error} */ (e);
            addToast("Error deleting folder: " + err.message, "error");
        }
    }

    /**
     * @param {string} instanceId
     * @param {string} folderId
     */
    async function moveInstance(instanceId, folderId) {
        const inst = instances.find((i) => i.id === instanceId);
        if (!inst) return;

        // Optimistic update
        inst.folderId = folderId;
        instances = [...instances];

        try {
            // We need to fetch current settings to not overwrite other fields,
            // but for now we assume consistent state or just patch what we need if backend supported it.
            // Since existing UpdateSettings requires all fields, we fetch first.
            // Actually, let's just use the current instance state we have.

            const payload = {
                maxMemory: inst.maxMemory,
                javaArgs: inst.javaArgs,
                jarFile: inst.jarFile,
                javaPath: inst.javaPath,
                webhookUrl: inst.webhookUrl,
                group: inst.group, // Keep legacy group for now or clear it? Let's keep it.
                folderId: folderId,
            };

            const res = await fetch(`/api/instances/${instanceId}`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });

            if (!res.ok) throw new Error(await res.text());
            addToast("Instance moved", "success");
            loadData();
        } catch (e) {
            /** @type {Error} */
            const err = /** @type {Error} */ (e);
            addToast("Failed to move instance: " + err.message, "error");
            loadData(); // Revert
        }
    }

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
    function handleDrop(event, folderId) {
        event.preventDefault();
        if (event.dataTransfer) {
            const id = event.dataTransfer.getData("text/plain");
            if (id) {
                moveInstance(id, folderId);
            }
        }
        draggingInstanceId = null;
    }

    // Instance management functions...
    /**
     * @param {string} id
     */
    async function deleteInstance(id) {
        const confirmed = await askConfirm({
            title: "Delete Instance",
            message:
                "Are you sure you want to delete this instance? This action cannot be undone and will permanently delete all files.",
            confirmText: "Delete",
            dangerous: true,
        });

        if (!confirmed) return;

        try {
            const res = await fetch(`/api/instances/${id}`, {
                method: "DELETE",
            });
            if (res.ok) {
                addToast("Instance deleted successfully", "success");
                loadData();
            } else {
                addToast("Failed to delete instance", "error");
            }
        } catch (e) {
            /** @type {Error} */
            const err = /** @type {Error} */ (e);
            addToast("Error deleting instance: " + err.message, "error");
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
            addToast(`Instance ${action}ed successfully`, "success");

            loadData();
        } catch (e) {
            const message = e instanceof Error ? e.message : String(e);
            console.error(`Failed to ${action} instance`, e);
            addToast(`Failed to ${action}: ${message}`, "error");
        }
    }

    // Grouping logic
    /** @type {Record<string, Instance[]>} */
    $: groupedInstances = {
        Uncategorized: instances.filter(
            (i) => !i.folderId || !folders.find((f) => f.id === i.folderId),
        ),
    };

    $: {
        folders.forEach((f) => {
            groupedInstances[f.id] = instances.filter(
                (i) => i.folderId === f.id,
            );
        });
    }

    $: sortedGroups = ["Uncategorized", ...folders.map((f) => f.id)];

    /**
     * @param {string} groupId
     * @returns {Instance[]}
     */
    function getGroupInstances(groupId) {
        return groupedInstances[groupId] || [];
    }
</script>

<div class="h-full flex flex-col p-8">
    <header class="flex justify-between items-center mb-8">
        <div>
            <h2 class="text-2xl font-bold text-white tracking-tight">
                Your Instances
            </h2>
            <div class="text-sm text-gray-400 mt-1">
                Manage and deploy your Minecraft servers
            </div>
        </div>
        <div class="flex gap-3">
            <button
                on:click={createFolder}
                class="bg-gray-800 hover:bg-gray-700 text-white px-5 py-2.5 rounded-xl font-bold text-sm transition-all hover:shadow-lg hover:shadow-gray-500/20 flex items-center gap-2 border border-white/10"
            >
                <svg
                    class="w-4 h-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
                    /></svg
                >
                New Folder
            </button>
            <a
                href="/instances/create"
                class="bg-indigo-600 hover:bg-indigo-500 text-white px-5 py-2.5 rounded-xl font-bold text-sm transition-all hover:shadow-lg hover:shadow-indigo-500/20 flex items-center gap-2"
            >
                <svg
                    class="w-4 h-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M12 4v16m8-8H4"
                    /></svg
                >
                Create New Instance
            </a>
        </div>
    </header>

    <div
        class="bg-gray-900/60 backdrop-blur-xl border border-white/5 rounded-2xl overflow-hidden shadow-xl flex-1 flex flex-col min-h-0"
    >
        <div
            class="grid grid-cols-12 gap-4 px-6 py-4 border-b border-white/5 bg-white/5 text-xs font-bold text-gray-400 uppercase tracking-wider"
        >
            <div class="col-span-4">Instance Name</div>
            <div class="col-span-2">Type</div>
            <div class="col-span-2">Version</div>
            <div class="col-span-2">Status</div>
            <div class="col-span-2 text-right">Actions</div>
        </div>

        <div class="overflow-y-auto flex-1 p-2 space-y-4">
            {#if loading}
                <div
                    class="h-full flex flex-col items-center justify-center text-gray-500"
                >
                    <svg
                        class="animate-spin h-8 w-8 mb-4 text-indigo-500"
                        xmlns="http://www.w3.org/2000/svg"
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
                    <span>Loading instances...</span>
                </div>
            {:else}
                {#each sortedGroups as groupId}
                    {@const isFolder = groupId !== "Uncategorized"}
                    {@const folder = folders.find((f) => f.id === groupId)}
                    {@const groupName = folder ? folder.name : "Uncategorized"}
                    {@const groupInstances = getGroupInstances(groupId)}

                    <div
                        class="mb-2 transition-all"
                        on:dragover|preventDefault
                        on:drop={(e) => handleDrop(e, isFolder ? groupId : "")}
                        role="region"
                        aria-label={groupName}
                    >
                        <div
                            class="px-4 py-2 text-xs font-bold text-gray-500 uppercase tracking-wider flex items-center justify-between group/header"
                        >
                            <div class="flex items-center gap-2">
                                <svg
                                    class="w-3 h-3"
                                    fill="none"
                                    stroke="currentColor"
                                    viewBox="0 0 24 24"
                                >
                                    {#if isFolder}
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                                        />
                                    {:else}
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M4 6h16M4 10h16M4 14h16M4 18h16"
                                        />
                                    {/if}
                                </svg>
                                {groupName}
                                <span
                                    class="bg-gray-800 text-gray-500 px-1.5 py-0.5 rounded text-[10px]"
                                    >{groupInstances.length}</span
                                >
                            </div>

                            {#if isFolder}
                                <button
                                    on:click={() => deleteFolder(groupId)}
                                    class="text-gray-600 hover:text-red-400 opacity-0 group-hover/header:opacity-100 transition-opacity"
                                    title="Delete Folder"
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
                                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                                        /></svg
                                    >
                                </button>
                            {/if}
                        </div>

                        {#if groupInstances.length === 0}
                            <div
                                class="px-4 py-8 border-2 border-dashed border-white/5 rounded-lg text-center text-gray-600 text-xs mx-2"
                            >
                                Drop instances here
                            </div>
                        {:else}
                            <div class="space-y-1">
                                {#each groupInstances as inst (inst.id)}
                                    <div
                                        class="grid grid-cols-12 gap-4 items-center px-4 py-3 rounded-lg hover:bg-white/5 transition-colors group cursor-grab active:cursor-grabbing bg-transparent"
                                        draggable="true"
                                        role="button"
                                        tabindex="0"
                                        on:dragstart={(e) =>
                                            handleDragStart(e, inst.id)}
                                    >
                                        <div class="col-span-4 min-w-0">
                                            <div
                                                class="font-bold text-white truncate"
                                            >
                                                {inst.name}
                                            </div>
                                            <div
                                                class="text-[10px] text-gray-500 font-mono truncate"
                                            >
                                                ID: {inst.id}
                                            </div>
                                        </div>
                                        <div class="col-span-2">
                                            <div
                                                class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-800 text-gray-300 border border-white/5"
                                            >
                                                {inst.type || "Unknown"}
                                            </div>
                                        </div>
                                        <div class="col-span-2">
                                            <div
                                                class="text-xs text-gray-300 font-mono"
                                            >
                                                {inst.version || "Latest"}
                                            </div>
                                        </div>
                                        <div class="col-span-2">
                                            {#if inst.status === "Online"}
                                                <div
                                                    class="flex items-center gap-2"
                                                >
                                                    <div
                                                        class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"
                                                    ></div>
                                                    <span
                                                        class="text-emerald-400 text-xs font-bold"
                                                        >Online</span
                                                    >
                                                </div>
                                            {:else}
                                                <div
                                                    class="flex items-center gap-2"
                                                >
                                                    <div
                                                        class="w-2 h-2 rounded-full bg-gray-600"
                                                    ></div>
                                                    <span
                                                        class="text-gray-500 text-xs font-bold"
                                                        >Offline</span
                                                    >
                                                </div>
                                            {/if}
                                        </div>
                                        <div
                                            class="col-span-2 flex justify-end gap-2 opacity-60 group-hover:opacity-100 transition-opacity"
                                        >
                                            {#if inst.status === "Online" || inst.status === "Starting"}
                                                <button
                                                    on:click|stopPropagation={() =>
                                                        triggerInstanceAction(
                                                            inst.id,
                                                            "stop",
                                                        )}
                                                    class="px-2 py-1.5 rounded-lg bg-red-500/10 hover:bg-red-500/20 text-red-400 hover:text-red-300 transition-colors border border-red-500/20"
                                                    title="Stop Server"
                                                >
                                                    <svg
                                                        class="w-4 h-4 fill-current"
                                                        viewBox="0 0 24 24"
                                                        ><path
                                                            d="M6 6h12v12H6z"
                                                        /></svg
                                                    >
                                                </button>
                                            {:else}
                                                <button
                                                    on:click|stopPropagation={() =>
                                                        triggerInstanceAction(
                                                            inst.id,
                                                            "start",
                                                        )}
                                                    class="px-2 py-1.5 rounded-lg bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 hover:text-emerald-300 transition-colors border border-emerald-500/20"
                                                    title="Start Server"
                                                >
                                                    <svg
                                                        class="w-4 h-4 fill-current"
                                                        viewBox="0 0 24 24"
                                                        ><path
                                                            d="M8 5v14l11-7z"
                                                        /></svg
                                                    >
                                                </button>
                                            {/if}

                                            <a
                                                href="/instances/{inst.id}"
                                                class="px-3 py-1.5 rounded-lg bg-white/5 hover:bg-white/10 text-white text-xs font-bold transition-colors border border-white/5"
                                                >Manage</a
                                            >
                                            <button
                                                on:click={() =>
                                                    deleteInstance(inst.id)}
                                                class="px-3 py-1.5 rounded-lg bg-red-500/10 hover:bg-red-500/20 text-red-400 hover:text-red-300 text-xs font-bold transition-colors border border-red-500/20"
                                                >Delete</button
                                            >
                                        </div>
                                    </div>
                                {/each}
                            </div>
                        {/if}
                    </div>
                {/each}
            {/if}
        </div>
    </div>
</div>
