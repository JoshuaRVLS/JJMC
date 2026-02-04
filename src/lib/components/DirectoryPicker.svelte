<script>
    import { createEventDispatcher, onMount } from "svelte";
    import { X, Loader2, Folder } from "lucide-svelte";

    export let open = false;
    export let title = "Select Directory";

    const dispatch = createEventDispatcher();

    let currentPath = "";
    let selectedPath = "";

    /** @type {any[]} */
    let files = [];
    let loading = false;
    let error = "";

    async function loadPath(path = "") {
        loading = true;
        error = "";
        selectedPath = "";
        try {
            const res = await fetch(
                `/api/system/files?path=${encodeURIComponent(path)}`,
            );
            if (res.ok) {
                const data = await res.json();
                currentPath = data.path;
                files = data.files;
            } else {
                const err = await res.json();
                error = err.error || "Failed to load directory";
            }
        } catch (e) {
            console.error(e);
            error = "Network error";
        } finally {
            loading = false;
        }
    }

    function select() {
        dispatch("select", selectedPath || currentPath);
        close();
    }

    function close() {
        dispatch("close");
        open = false;
    }

    /** @param {KeyboardEvent} e */
    function handleKeydown(e) {
        if (e.key === "Enter" && selectedPath) {
            loadPath(selectedPath);
        }
    }

    onMount(() => {
        if (open) {
            loadPath();
        }
    });

    $: if (open && !currentPath) {
        loadPath();
    }
</script>

{#if open}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
    >
        <div
            class="bg-gray-900 border border-gray-700 rounded-lg shadow-2xl w-full max-w-2xl flex flex-col max-h-[80vh]"
        >
            <div
                class="p-4 border-b border-gray-800 flex justify-between items-center bg-gray-800/50 rounded-t-lg"
            >
                <h3 class="text-lg font-semibold text-white">{title}</h3>
                <button on:click={close} class="text-gray-400 hover:text-white">
                    <X class="w-5 h-5" />
                </button>
            </div>

            <div class="p-2 border-b border-gray-800 bg-gray-950/30 flex gap-2">
                <input
                    type="text"
                    value={currentPath}
                    on:change={(e) => loadPath(e.currentTarget.value)}
                    class="flex-1 bg-gray-800 border border-gray-700 rounded px-2 py-1 text-sm text-gray-300 font-mono"
                />
                <button
                    on:click={() => loadPath(currentPath)}
                    class="bg-gray-700 hover:bg-gray-600 text-white px-3 py-1 rounded text-sm"
                    >Go</button
                >
            </div>

            <div class="flex-1 overflow-y-auto p-2 min-h-[300px]">
                {#if loading}
                    <div
                        class="flex items-center justify-center h-full text-gray-400"
                    >
                        <Loader2 class="animate-spin h-6 w-6 mr-2" />
                        Loading...
                    </div>
                {:else if error}
                    <div
                        class="flex items-center justify-center h-full text-red-400"
                    >
                        {error}
                    </div>
                {:else}
                    <div class="grid grid-cols-1 gap-1">
                        {#each files as file}
                            <button
                                class="flex items-center gap-3 p-2 rounded text-left group transition-all {selectedPath ===
                                file.path
                                    ? 'bg-blue-600/20 ring-1 ring-blue-500 text-white'
                                    : 'hover:bg-gray-800 text-gray-300'}"
                                on:click={() => (selectedPath = file.path)}
                                on:dblclick={() => loadPath(file.path)}
                            >
                                <div
                                    class={selectedPath === file.path
                                        ? "text-blue-400"
                                        : "text-yellow-500"}
                                >
                                    <Folder class="w-6 h-6" />
                                </div>
                                <span class="flex-1 truncate font-medium">
                                    {file.name}
                                </span>
                            </button>
                        {/each}
                        {#if files.length === 0}
                            <div class="text-gray-500 text-center p-4 italic">
                                Empty directory
                            </div>
                        {/if}
                    </div>
                {/if}
            </div>

            <div
                class="p-4 border-t border-gray-800 flex justify-end gap-2 bg-gray-800/50 rounded-b-lg"
            >
                <button
                    on:click={close}
                    class="px-4 py-2 rounded text-gray-400 hover:text-white hover:bg-gray-700 transition-colors text-sm font-medium"
                    >Cancel</button
                >
                <button
                    on:click={select}
                    class="px-4 py-2 rounded bg-blue-600 hover:bg-blue-500 text-white shadow-lg transition-colors text-sm font-medium flex items-center gap-2"
                >
                    Select {selectedPath ? "Highlight" : "Current Folder"}
                </button>
            </div>
        </div>
    </div>
{/if}
