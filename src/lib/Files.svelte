<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import { askConfirm } from "$lib/stores/confirm";
    import { askInput } from "$lib/stores/input";
    import { X, Upload } from "lucide-svelte";
    import FileToolbar from "./files/FileToolbar.svelte";
    import FileTable from "./files/FileTable.svelte";

    /** @type {string} */
    export let instanceId;

    /**
     * @typedef {Object} FileItem
     * @property {string} name
     * @property {boolean} isDir
     * @property {number} size
     * @property {string} mode
     * @property {string} modTime
     */

    /** @type {FileItem[]} */
    let files = [];
    let loading = true;
    let currentPath = ".";

    /** @type {FileItem | null} */
    let viewingFile = null;
    let fileContent = "";

    /** @type {{name: string, path: string}[]} */
    let breadcrumbs = [];

    /** @type {Set<string>} */
    let selectedFiles = new Set();

    let isDraggingOver = false;

    async function loadFiles(path = ".") {
        loading = true;
        selectedFiles = new Set();
        try {
            const res = await fetch(
                `/api/instances/${instanceId}/files?path=${encodeURIComponent(
                    path,
                )}`,
            );
            if (res.ok) {
                files = await res.json();

                files.sort((a, b) => {
                    if (a.isDir && !b.isDir) return -1;
                    if (!a.isDir && b.isDir) return 1;
                    return a.name.localeCompare(b.name);
                });
                currentPath = path;

                const parts = path === "." ? [] : path.split("/");
                breadcrumbs = [{ name: "Root", path: "." }];
                let acc = "";
                parts.forEach((part) => {
                    if (part !== ".") {
                        acc = acc ? `${acc}/${part}` : part;
                        breadcrumbs.push({ name: part, path: acc });
                    }
                });
            } else {
                addToast("Failed to load files", "error");
            }
        } catch (e) {
            addToast("Error loading files", "error");
        } finally {
            loading = false;
        }
    }

    /**
     * @param {string} path
     */
    function navigate(path) {
        loadFiles(path);
    }

    function navigateUp() {
        if (currentPath === "." || currentPath === "") return;
        const parts = currentPath.split("/");
        parts.pop();
        const newPath = parts.length === 0 ? "." : parts.join("/");
        loadFiles(newPath);
    }

    /**
     * @param {FileItem} file
     */
    async function openFile(file) {
        if (file.isDir) {
            const newPath =
                currentPath === "." ? file.name : `${currentPath}/${file.name}`;
            loadFiles(newPath);
        } else {
            try {
                const filePath =
                    currentPath === "."
                        ? file.name
                        : `${currentPath}/${file.name}`;
                const res = await fetch(
                    `/api/instances/${instanceId}/files/content?path=${encodeURIComponent(
                        filePath,
                    )}`,
                );
                if (res.ok) {
                    fileContent = await res.text();
                    viewingFile = file;
                } else {
                    addToast("Failed to read file", "error");
                }
            } catch (e) {
                addToast("Error reading file", "error");
            }
        }
    }

    async function saveFile() {
        if (!viewingFile) return;
        try {
            const res = await fetch(
                `/api/instances/${instanceId}/files/content`,
                {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        path:
                            currentPath === "."
                                ? viewingFile.name
                                : `${currentPath}/${viewingFile.name}`,
                        content: fileContent,
                    }),
                },
            );

            if (res.ok) {
                addToast("File saved", "success");
            } else {
                addToast("Failed to save file", "error");
            }
        } catch (e) {
            addToast("Error saving file", "error");
        }
    }

    function toggleAll() {
        if (selectedFiles.size === files.length) {
            selectedFiles.clear();
        } else {
            files.forEach((f) => selectedFiles.add(f.name));
        }
        selectedFiles = selectedFiles;
    }

    async function deleteSelected() {
        const count = selectedFiles.size;
        if (count === 0) return;

        const confirmed = await askConfirm({
            title: "Delete Files",
            message: `Are you sure you want to delete ${count} item(s)? This cannot be undone.`,
            confirmText: "Delete",
            dangerous: true,
        });

        if (!confirmed) return;

        let errors = 0;
        for (const name of selectedFiles) {
            try {
                const path =
                    currentPath === "." ? name : `${currentPath}/${name}`;
                const res = await fetch(
                    `/api/instances/${instanceId}/files?path=${encodeURIComponent(
                        path,
                    )}`,
                    {
                        method: "DELETE",
                    },
                );
                if (!res.ok) errors++;
            } catch (e) {
                errors++;
            }
        }

        if (errors > 0) addToast(`Failed to delete ${errors} items`, "error");

        loadFiles(currentPath);
    }

    /**
     * @param {FileItem} file
     * @param {Event} event
     */
    function toggleSelection(file, event) {
        if (selectedFiles.has(file.name)) {
            selectedFiles.delete(file.name);
        } else {
            selectedFiles.add(file.name);
        }
        selectedFiles = selectedFiles;
    }

    async function createDirectory() {
        const name = await askInput({
            title: "New Folder",
            placeholder: "Folder Name",
            confirmText: "Create",
        });
        if (!name) return;

        try {
            const newPath =
                currentPath === "." ? name : `${currentPath}/${name}`;

            const res = await fetch(
                `/api/instances/${instanceId}/files/directory`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ path: newPath }),
                },
            );

            if (res.ok) {
                loadFiles(currentPath);
            } else {
                addToast("Failed to create folder", "error");
            }
        } catch (e) {
            addToast("Error creating folder", "error");
        }
    }

    async function createFile() {
        const name = await askInput({
            title: "Create File",
            placeholder: "File Name (e.g., settings.txt)",
            confirmText: "Create",
        });
        if (!name) return;

        try {
            const filePath =
                currentPath === "." ? name : `${currentPath}/${name}`;
            const res = await fetch(
                `/api/instances/${instanceId}/files/content`,
                {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        path: filePath,
                        content: "",
                    }),
                },
            );

            if (res.ok) {
                loadFiles(currentPath);
            } else {
                addToast("Failed to create file", "error");
            }
        } catch (e) {
            addToast("Error creating file", "error");
        }
    }

    /**
     * @param {FileList} fileList
     */
    async function uploadFiles(fileList) {
        if (!fileList || fileList.length === 0) return;

        const formData = new FormData();
        for (let i = 0; i < fileList.length; i++) {
            formData.append("files", fileList[i]);
        }

        try {
            addToast(`Uploading ${fileList.length} files...`, "info");
            const res = await fetch(
                `/api/instances/${instanceId}/files/upload?path=${encodeURIComponent(
                    currentPath,
                )}`,
                {
                    method: "POST",
                    body: formData,
                },
            );

            if (res.ok) {
                addToast("Upload complete", "success");
                loadFiles(currentPath);
            } else {
                addToast("Upload failed", "error");
            }
        } catch (e) {
            addToast("Error uploading files", "error");
        }
    }

    /**
     * @param {DragEvent} e
     */
    function onDragOver(e) {
        e.preventDefault();
        isDraggingOver = true;
    }

    async function compressSelected() {
        const name = await askInput({
            title: "Compress Files",
            placeholder: "Archive Name (e.g. backup.zip)",
            confirmText: "Compress",
            value: "archive.zip",
        });
        if (!name) return;

        try {
            loading = true;
            const filesToCompress = Array.from(selectedFiles);
            const res = await fetch(
                `/api/instances/${instanceId}/files/compress`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        files: filesToCompress,
                        cwd: currentPath,
                        destination: name,
                    }),
                },
            );

            if (res.ok) {
                addToast("Compression started/finished", "success");
                loadFiles(currentPath);
                selectedFiles.clear();
                selectedFiles = selectedFiles;
            } else {
                const err = await res.json();
                addToast(err.error || "Compression failed", "error");
            }
        } catch (e) {
            addToast("Error compressing", "error");
        } finally {
            loading = false;
        }
    }

    async function extractSelected() {
        const file = Array.from(selectedFiles)[0];
        if (!file) return;

        const filePath = currentPath === "." ? file : `${currentPath}/${file}`;

        try {
            loading = true;
            const res = await fetch(
                `/api/instances/${instanceId}/files/extract`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        path: filePath,
                    }),
                },
            );

            if (res.ok) {
                addToast("Extraction started", "success");

                setTimeout(() => loadFiles(currentPath), 2000);
            } else {
                const err = await res.json();
                addToast(err.error || "Extraction failed", "error");
            }
        } catch (e) {
            addToast("Error extracting", "error");
        } finally {
            loading = false;
        }
    }

    /**
     * @param {DragEvent} e
     */
    function onDragLeave(e) {
        e.preventDefault();
        isDraggingOver = false;
    }

    /**
     * @param {DragEvent} e
     */
    function onDrop(e) {
        e.preventDefault();
        isDraggingOver = false;
        if (e.dataTransfer && e.dataTransfer.files) {
            uploadFiles(e.dataTransfer.files);
        }
    }

    onMount(() => {
        loadFiles();
    });
</script>

<div
    class="h-full flex flex-col bg-gray-900/50 rounded-xl overflow-hidden border border-white/5 relative"
    on:dragover={onDragOver}
    on:dragleave={onDragLeave}
    on:drop={onDrop}
    role="region"
    aria-label="File Browser"
>
    {#if isDraggingOver}
        <div
            class="absolute inset-0 z-50 bg-indigo-500/20 backdrop-blur-sm border-2 border-dashed border-indigo-400 flex flex-col items-center justify-center text-white pointer-events-none"
        >
            <Upload class="w-16 h-16 mb-4 text-indigo-300 animate-bounce" />
            <span class="text-xl font-bold">Drop files to upload</span>
        </div>
    {/if}

    {#if viewingFile}
        <div class="flex flex-col h-full">
            <div
                class="flex items-center justify-between px-4 py-3 bg-white/5 border-b border-white/5"
            >
                <div class="flex items-center gap-2">
                    <button
                        on:click={() => (viewingFile = null)}
                        class="text-gray-400 hover:text-white"
                    >
                        <X class="w-5 h-5" />
                    </button>
                    <span class="font-mono text-sm text-gray-200"
                        >{viewingFile.name}</span
                    >
                </div>
                <div class="flex items-center gap-2">
                    <button
                        on:click={saveFile}
                        class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-lg transition-colors"
                    >
                        Save
                    </button>
                </div>
            </div>
            <textarea
                class="flex-1 w-full bg-[#1e1e1e] text-gray-300 font-mono text-sm p-4 focus:outline-none resize-none"
                bind:value={fileContent}
                spellcheck="false"
            ></textarea>
        </div>
    {:else}
        <FileToolbar
            {currentPath}
            {selectedFiles}
            {breadcrumbs}
            on:navigateUp={navigateUp}
            on:navigate={(e) => navigate(e.detail)}
            on:deleteSelected={deleteSelected}
            on:compressSelected={compressSelected}
            on:extractSelected={extractSelected}
            on:createFile={createFile}
            on:createDirectory={createDirectory}
            on:uploadFiles={(e) => uploadFiles(e.detail)}
        />

        <FileTable
            {files}
            {loading}
            {selectedFiles}
            on:toggleAll={toggleAll}
            on:toggleSelection={(e) =>
                toggleSelection(e.detail.file, e.detail.event)}
            on:openFile={(e) => openFile(e.detail)}
        />
    {/if}
</div>
