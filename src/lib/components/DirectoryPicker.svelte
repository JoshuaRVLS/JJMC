<script>
    import { createEventDispatcher, onMount } from "svelte";

    export let open = false;
    export let title = "Select Directory";

    const dispatch = createEventDispatcher();

    let currentPath = "";
    /** @type {Array<{name: string, path: string, isDir: boolean}>} */
    let files = [];
    let loading = false;
    let error = "";

    async function loadPath(path = "") {
        loading = true;
        error = "";
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
        dispatch("select", currentPath);
        close();
    }

    function close() {
        dispatch("close");
        open = false;
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
            <!-- Header -->
            <div
                class="p-4 border-b border-gray-800 flex justify-between items-center bg-gray-800/50 rounded-t-lg"
            >
                <h3 class="text-lg font-semibold text-white">{title}</h3>
                <button on:click={close} class="text-gray-400 hover:text-white">
                    <svg
                        class="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        ><path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M6 18L18 6M6 6l12 12"
                        ></path></svg
                    >
                </button>
            </div>

            <!-- Path Bar -->
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

            <!-- Content -->
            <div class="flex-1 overflow-y-auto p-2 min-h-[300px]">
                {#if loading}
                    <div
                        class="flex items-center justify-center h-full text-gray-400"
                    >
                        <svg
                            class="animate-spin h-6 w-6 mr-2"
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
                                class="flex items-center gap-3 p-2 hover:bg-gray-800 rounded text-left group"
                                on:click={() => loadPath(file.path)}
                            >
                                <div class="text-yellow-500">
                                    <svg
                                        class="w-6 h-6"
                                        fill="currentColor"
                                        viewBox="0 0 24 24"
                                        ><path
                                            d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"
                                        ></path></svg
                                    >
                                </div>
                                <span
                                    class="flex-1 text-gray-300 group-hover:text-white truncate font-medium"
                                >
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

            <!-- Footer -->
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
                    Select Current Folder
                </button>
            </div>
        </div>
    </div>
{/if}
