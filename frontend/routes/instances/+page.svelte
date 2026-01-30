<script>
    import { onMount } from "svelte";
    import { askConfirm } from "$lib/stores/confirm.js";
    import { addToast } from "$lib/stores/toast.js";

    import { onDestroy } from "svelte";

    /**
     * @typedef {Object} Instance
     * @property {string} id
     * @property {string} name
     * @property {string} type
     * @property {string} version
     * @property {string} status
     */

    /** @type {Instance[]} */
    let instances = [];
    let loading = true;
    /** @type {ReturnType<typeof setInterval> | undefined} */
    let pollInterval;

    async function loadInstances() {
        // Don't set loading=true on subsequent polls to avoid flickering
        if (!instances.length) loading = true;

        try {
            const res = await fetch("/api/instances");
            if (res.ok) {
                instances = await res.json();
            }
        } catch (e) {
            console.error("Failed to load instances", e);
        } finally {
            loading = false;
        }
    }

    onMount(() => {
        loadInstances();
        pollInterval = setInterval(loadInstances, 2000);
    });

    onDestroy(() => {
        if (pollInterval) clearInterval(pollInterval);
    });

    /** @param {string} id */
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
                loadInstances();
            } else {
                addToast("Failed to delete instance", "error");
            }
        } catch (e) {
            addToast(
                "Error deleting instance: " + /** @type {Error} */ (e).message,
                "error",
            );
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
            // Refresh list to update status UI immediately (and let poll pick it up)
            loadInstances();
        } catch (e) {
            const message = e instanceof Error ? e.message : String(e);
            console.error(`Failed to ${action} instance`, e);
            addToast(`Failed to ${action}: ${message}`, "error");
        }
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
    </header>

    <div
        class="bg-gray-900/60 backdrop-blur-xl border border-white/5 rounded-2xl overflow-hidden shadow-xl flex-1 flex flex-col min-h-0"
    >
        <!-- Table Header -->
        <div
            class="grid grid-cols-12 gap-4 px-6 py-4 border-b border-white/5 bg-white/5 text-xs font-bold text-gray-400 uppercase tracking-wider"
        >
            <div class="col-span-4">Instance Name</div>
            <div class="col-span-2">Type</div>
            <div class="col-span-2">Version</div>
            <div class="col-span-2">Status</div>
            <div class="col-span-2 text-right">Actions</div>
        </div>

        <!-- Table Body -->
        <div class="overflow-y-auto flex-1 p-2 space-y-1">
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
            {:else if instances.length === 0}
                <div
                    class="h-full flex flex-col items-center justify-center text-gray-500"
                >
                    <div
                        class="w-16 h-16 bg-gray-800 rounded-full flex items-center justify-center mb-4"
                    >
                        <svg
                            class="w-8 h-8 text-gray-600"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                            ><path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
                            /></svg
                        >
                    </div>
                    <span>No instances found. Create one to get started.</span>
                </div>
            {:else}
                {#each instances as inst}
                    <div
                        class="grid grid-cols-12 gap-4 items-center px-4 py-3 rounded-lg hover:bg-white/5 transition-colors group"
                    >
                        <div class="col-span-4 min-w-0">
                            <div class="font-bold text-white truncate">
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
                            <div class="text-xs text-gray-300 font-mono">
                                {inst.version || "Latest"}
                            </div>
                        </div>
                        <div class="col-span-2">
                            {#if inst.status === "Online"}
                                <div class="flex items-center gap-2">
                                    <div
                                        class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"
                                    />
                                    <span
                                        class="text-emerald-400 text-xs font-bold"
                                        >Online</span
                                    >
                                </div>
                            {:else}
                                <div class="flex items-center gap-2">
                                    <div
                                        class="w-2 h-2 rounded-full bg-gray-600"
                                    />
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
                            <!-- Start/Stop Button -->
                            {#if inst.status === "Online" || inst.status === "Starting"}
                                <button
                                    on:click|stopPropagation={() =>
                                        triggerInstanceAction(inst.id, "stop")}
                                    class="px-2 py-1.5 rounded-lg bg-red-500/10 hover:bg-red-500/20 text-red-400 hover:text-red-300 transition-colors border border-red-500/20"
                                    title="Stop Server"
                                >
                                    <svg
                                        class="w-4 h-4 fill-current"
                                        viewBox="0 0 24 24"
                                        ><path d="M6 6h12v12H6z" /></svg
                                    >
                                </button>
                            {:else}
                                <button
                                    on:click|stopPropagation={() =>
                                        triggerInstanceAction(inst.id, "start")}
                                    class="px-2 py-1.5 rounded-lg bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 hover:text-emerald-300 transition-colors border border-emerald-500/20"
                                    title="Start Server"
                                >
                                    <svg
                                        class="w-4 h-4 fill-current"
                                        viewBox="0 0 24 24"
                                        ><path d="M8 5v14l11-7z" /></svg
                                    >
                                </button>
                            {/if}

                            <a
                                href="/instances/{inst.id}"
                                class="px-3 py-1.5 rounded-lg bg-white/5 hover:bg-white/10 text-white text-xs font-bold transition-colors border border-white/5"
                            >
                                Manage
                            </a>
                            <button
                                on:click={() => deleteInstance(inst.id)}
                                class="px-3 py-1.5 rounded-lg bg-red-500/10 hover:bg-red-500/20 text-red-400 hover:text-red-300 text-xs font-bold transition-colors border border-red-500/20"
                            >
                                Delete
                            </button>
                        </div>
                    </div>
                {/each}
            {/if}
        </div>
    </div>
</div>
