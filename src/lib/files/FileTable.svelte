<script>
    import { Loader2, Folder, File } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";

    export let files;

    export let loading;

    export let selectedFiles;

    const dispatch = createEventDispatcher();

    /** @param {number} bytes */
    function formatSize(bytes) {
        if (bytes === 0) return "0 B";
        const k = 1024;
        const sizes = ["B", "KB", "MB", "GB", "TB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
    }

    /** @param {number} ms */
    function formatDate(ms) {
        return new Date(ms).toLocaleString();
    }

    function toggleAll() {
        dispatch("toggleAll");
    }

    /**
     * @param {any} file
     * @param {Event} event
     */
    function toggleSelection(file, event) {
        dispatch("toggleSelection", { file, event });
    }

    /** @param {any} file */
    function openFile(file) {
        dispatch("openFile", file);
    }
</script>

<div class="flex-1 overflow-y-auto">
    {#if loading}
        <div class="flex items-center justify-center h-40">
            <Loader2 class="animate-spin h-8 w-8 text-indigo-500" />
        </div>
    {:else if files.length === 0}
        <div
            class="flex flex-col items-center justify-center h-40 text-gray-500"
        >
            <span>Empty directory</span>
        </div>
    {:else}
        <table class="w-full text-left border-collapse">
            <thead
                class="bg-gray-900 text-xs uppercase text-gray-400 font-semibold sticky top-0 z-10 border-b border-white/5"
            >
                <tr>
                    <th class="px-4 py-3 w-8">
                        <input
                            type="checkbox"
                            class="rounded bg-black/20 border-white/10 text-indigo-500 focus:ring-0 cursor-pointer"
                            checked={files.length > 0 &&
                                selectedFiles.size === files.length}
                            on:change={toggleAll}
                        />
                    </th>
                    <th class="px-4 py-3">Name</th>
                    <th class="px-4 py-3 w-32">Size</th>
                    <th class="px-4 py-3 w-48">Modified</th>
                </tr>
            </thead>
            <tbody class="divide-y divide-white/5 text-sm text-gray-300">
                {#each files as file (file.name)}
                    <tr
                        class="hover:bg-white/5 transition-colors group {selectedFiles.has(
                            file.name,
                        )
                            ? 'bg-indigo-500/10'
                            : ''}"
                    >
                        <td class="px-4 py-2">
                            <input
                                type="checkbox"
                                class="rounded bg-black/20 border-white/10 text-indigo-500 focus:ring-0 cursor-pointer"
                                checked={selectedFiles.has(file.name)}
                                on:change={(e) => toggleSelection(file, e)}
                            />
                        </td>
                        <td class="px-4 py-2">
                            <button
                                class="flex items-center gap-3 hover:text-white w-full text-left truncate"
                                on:click={() => openFile(file)}
                            >
                                {#if file.isDir}
                                    <Folder
                                        class="w-5 h-5 text-yellow-500 flex-shrink-0"
                                    />
                                {:else}
                                    <File
                                        class="w-5 h-5 text-gray-500 shrink-0"
                                    />
                                {/if}
                                <span class="truncate">{file.name}</span>
                            </button>
                        </td>
                        <td class="px-4 py-2 text-gray-500 font-mono text-xs"
                            >{file.isDir ? "-" : formatSize(file.size)}</td
                        >
                        <td class="px-4 py-2 text-gray-500 text-xs"
                            >{formatDate(file.modTime)}</td
                        >
                    </tr>
                {/each}
            </tbody>
        </table>
    {/if}
</div>
