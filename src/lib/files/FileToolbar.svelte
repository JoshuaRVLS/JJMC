<script>
    import {
        Trash2,
        Archive,
        ArchiveRestore,
        FilePlus,
        FolderPlus,
        Upload,
    } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";
    import FileBreadcrumbs from "./FileBreadcrumbs.svelte";

     
    export let currentPath;
     
    export let selectedFiles;
     
    export let breadcrumbs;

    const dispatch = createEventDispatcher();

    function handleFileUpload(e) {
        if (e.target.files) {
            dispatch("uploadFiles", e.target.files);
        }
    }
</script>

<div
    class="flex items-center justify-between px-4 py-3 bg-white/5 border-b border-white/5"
>
    <FileBreadcrumbs
        {currentPath}
        {breadcrumbs}
        on:navigateUp={() => dispatch("navigateUp")}
        on:navigate={(e) => dispatch("navigate", e.detail)}
    />

    <div class="flex items-center gap-2">
        {#if selectedFiles.size > 0}
            <button
                on:click={() => dispatch("deleteSelected")}
                class="flex items-center gap-1 px-2 py-1 bg-red-500/10 text-red-400 hover:bg-red-500 hover:text-white rounded-lg transition-colors text-xs font-bold mr-2"
            >
                <Trash2 class="w-4 h-4" />
                Delete ({selectedFiles.size})
            </button>

            <button
                on:click={() => dispatch("compressSelected")}
                class="flex items-center gap-1 px-2 py-1 bg-blue-500/10 text-blue-400 hover:bg-blue-500 hover:text-white rounded-lg transition-colors text-xs font-bold mr-2"
            >
                <Archive class="w-4 h-4" />
                Compress
            </button>

            <div class="h-4 w-px bg-white/10 mx-1"></div>
        {/if}

        {#if selectedFiles.size === 1 && (Array.from(selectedFiles)[0].endsWith(".zip") || Array.from(selectedFiles)[0].endsWith(".jar"))}
            <button
                on:click={() => dispatch("extractSelected")}
                class="flex items-center gap-1 px-2 py-1 bg-green-500/10 text-green-400 hover:bg-green-500 hover:text-white rounded-lg transition-colors text-xs font-bold mr-2"
            >
                <ArchiveRestore class="w-4 h-4" />
                Extract
            </button>
            <div class="h-4 w-px bg-white/10 mx-1"></div>
        {/if}

        <button
            on:click={() => dispatch("createFile")}
            class="p-2 hover:bg-white/10 rounded-lg text-gray-400 hover:text-white transition-colors"
            title="New File"
        >
            <FilePlus class="w-5 h-5" />
        </button>
        <button
            on:click={() => dispatch("createDirectory")}
            class="p-2 hover:bg-white/10 rounded-lg text-gray-400 hover:text-white transition-colors"
            title="New Folder"
        >
            <FolderPlus class="w-5 h-5" />
        </button>
        <label
            class="p-2 hover:bg-white/10 rounded-lg text-gray-400 hover:text-white transition-colors cursor-pointer"
            title="Upload"
        >
            <input
                type="file"
                multiple
                class="hidden"
                on:change={handleFileUpload}
            />
            <Upload class="w-5 h-5" />
        </label>
    </div>
</div>
