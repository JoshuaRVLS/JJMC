<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import { askConfirm } from "$lib/stores/confirm";
    import { askInput } from "$lib/stores/input";
    import {
        Loader2,
        Folder,
        File,
        ArrowUp,
        Trash2,
        FilePlus,
        FolderPlus,
        Upload,
        X,
        Archive,
        ArchiveRestore,
    } from "lucide-svelte";

    /** @type {string} */
    export let instanceId;

    /**
     * @typedef {Object} FileEntry
     * @property {string} name
     * @property {number} size
     * @property {boolean} isDir
     * @property {number} modTime
     * @property {string} [fullPath]
     */

    /** @type {FileEntry[]} */
    let files = [];
    let currentPath = ".";
    let loading = false;
    /** @type {FileEntry | null} */
    let viewingFile = null;
    let fileContent = "";
    /** @type {Array<{name: string, path: string}>} */
    let breadcrumbs = [];

    // Bulk selection
    /** @type {Set<string>} */
    let selectedFiles = new Set();
    let lastSelectedFile = null;

    // Drag & Drop
    let isDraggingOver = false;

    /** @param {string} [path] */
    async function loadFiles(path = ".") {
        loading = true;
        selectedFiles = new Set(); // access cleared on Nav
        try {
            const res = await fetch(
                `/api/instances/${instanceId}/files?path=${encodeURIComponent(
                    path,
                )}`,
            );
            if (res.ok) {
                files = await res.json();
                currentPath = path;
                updateBreadcrumbs(path);
            } else {
                addToast("Failed to load files", "error");
            }
        } catch (e) {
            console.error(e);
            addToast("Error loading files", "error");
        } finally {
            loading = false;
        }
    }

    /** @param {string} path */
    function updateBreadcrumbs(path) {
        const parts = path === "." ? [] : path.split("/");
        let acc = "";
        breadcrumbs = [{ name: "Home", path: "." }];
        parts.forEach((part) => {
            if (part && part !== ".") {
                acc = acc ? `${acc}/${part}` : part;
                breadcrumbs.push({ name: part, path: acc });
            }
        });
    }

    /** @param {string} path */
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

    /** @param {FileEntry} file */
    async function openFile(file) {
        if (file.isDir) {
            const newPath =
                currentPath === "." ? file.name : `${currentPath}/${file.name}`;
            loadFiles(newPath);
        } else {
            // Check if binary or too large
            if (file.size > 1024 * 1024) {
                const proceed = await askConfirm({
                    title: "Large File",
                    message:
                        "This file is larger than 1MB. Opening it might freeze the browser. Continue?",
                    confirmText: "Open Anyway",
                    dangerous: true,
                });
                if (!proceed) return;
            }

            // View file
            try {
                const path =
                    currentPath === "."
                        ? file.name
                        : `${currentPath}/${file.name}`;
                const res = await fetch(
                    `/api/instances/${instanceId}/files/content?path=${encodeURIComponent(
                        path,
                    )}`,
                );
                if (res.ok) {
                    fileContent = await res.text();
                    viewingFile = { ...file, fullPath: path };
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
                        path: viewingFile.fullPath,
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

    async function deleteSelected() {
        const count = selectedFiles.size;
        if (count === 0) return;

        const confirmed = await askConfirm({
            title: "Delete Files",
            message: `Are you sure you want to delete ${count} item${count > 1 ? "s" : ""}?`,
            confirmText: "Delete",
            dangerous: true,
        });

        if (!confirmed) return;

        let successCount = 0;
        let errors = 0;

        for (const fileName of selectedFiles) {
            try {
                const path =
                    currentPath === "."
                        ? fileName
                        : `${currentPath}/${fileName}`;
                const res = await fetch(
                    `/api/instances/${instanceId}/files?path=${encodeURIComponent(
                        path,
                    )}`,
                    { method: "DELETE" },
                );

                if (res.ok) {
                    successCount++;
                } else {
                    errors++;
                }
            } catch (e) {
                errors++;
            }
        }

        if (successCount > 0)
            addToast(`Deleted ${successCount} items`, "success");
        if (errors > 0) addToast(`Failed to delete ${errors} items`, "error");

        loadFiles(currentPath);
    }

    /**
     * @param {FileEntry} file
     * @param {Event} event
     */
    function toggleSelection(file, event) {
        // Shift select logic could go here
        if (selectedFiles.has(file.name)) {
            selectedFiles.delete(file.name);
        } else {
            selectedFiles.add(file.name);
        }
        selectedFiles = selectedFiles; // refresh
    }

    function toggleAll() {
        if (selectedFiles.size === files.length) {
            selectedFiles = new Set();
        } else {
            selectedFiles = new Set(files.map((f) => f.name));
        }
    }

    async function createDirectory() {
        const name = await askInput({
            title: "Create Folder",
            placeholder: "Folder Name",
            confirmText: "Create",
        });
        if (!name) return;

        try {
            const path = currentPath === "." ? name : `${currentPath}/${name}`;
            const res = await fetch(
                `/api/instances/${instanceId}/files/mkdir`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ path }),
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
            const path = currentPath === "." ? name : `${currentPath}/${name}`;
            const res = await fetch(
                `/api/instances/${instanceId}/files/content`,
                {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        path: path,
                        content: "",
                    }),
                },
            );
            if (res.ok) {
                loadFiles(currentPath);
                addToast("File created", "success");
            } else {
                addToast("Failed to create file", "error");
            }
        } catch (e) {
            addToast("Error creating file", "error");
        }
    }

    /** @param {FileList} fileList */
    async function uploadFiles(fileList) {
        if (!fileList || fileList.length === 0) return;

        const formData = new FormData();
        for (let i = 0; i < fileList.length; i++) {
            formData.append("files", fileList[i]);
        }

        try {
            loading = true;
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
                addToast("Uploaded successfully", "success");
                loadFiles(currentPath);
            } else {
                addToast("Upload failed", "error");
            }
        } catch (e) {
            addToast("Error uploading", "error");
        } finally {
            loading = false;
        }
    }

    /** @param {Event} e */
    function handleFileUpload(e) {
        const target = /** @type {HTMLInputElement} */ (e.target);
        if (target.files) uploadFiles(target.files);
        if (target) target.value = ""; // Reset input
    }

    // Drag and Drop Handlers
    /** @param {DragEvent} e */
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

        // Ensure .zip extension
        const finalName = name.endsWith(".zip") ? name : name + ".zip";
        const filesToCompress = Array.from(selectedFiles).map((f) =>
            currentPath === "." ? f : `${currentPath}/${f}`,
        );
        const destPath =
            currentPath === "." ? finalName : `${currentPath}/${finalName}`;

        try {
            loading = true; // or just toast
            const res = await fetch(
                `/api/instances/${instanceId}/files/compress`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        files: filesToCompress,
                        destination: destPath,
                    }),
                },
            );
            if (res.ok) {
                addToast("Compressed successfully", "success");
                loadFiles(currentPath);
                selectedFiles = new Set();
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
                `/api/instances/${instanceId}/files/decompress`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        file: filePath,
                        destination: currentPath, // Extract to current folder
                    }),
                },
            );
            if (res.ok) {
                addToast("Extracted successfully", "success");
                loadFiles(currentPath);
                selectedFiles = new Set();
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

    /** @param {DragEvent} e */
    function onDragLeave(e) {
        e.preventDefault();
        isDraggingOver = false;
    }

    /** @param {DragEvent} e */
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
    <!-- Drop Overlay -->
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
        <!-- Toolbar -->
        <div
            class="flex items-center justify-between px-4 py-3 bg-white/5 border-b border-white/5"
        >
            <div
                class="flex items-center gap-2 text-sm text-gray-400 overflow-x-auto"
            >
                <button
                    on:click={navigateUp}
                    disabled={currentPath === "."}
                    class="p-1 hover:text-white disabled:opacity-30 disabled:hover:text-gray-400"
                >
                    <ArrowUp class="w-5 h-5" />
                </button>
                <div class="h-4 w-px bg-white/10 mx-1"></div>
                {#each breadcrumbs as crumb, i}
                    <button
                        class="hover:text-white transition-colors whitespace-nowrap {i ===
                        breadcrumbs.length - 1
                            ? 'text-white font-semibold'
                            : ''}"
                        on:click={() => navigate(crumb.path)}
                    >
                        {crumb.name}
                    </button>
                    {#if i < breadcrumbs.length - 1}
                        <span class="text-gray-600">/</span>
                    {/if}
                {/each}
            </div>

            <div class="flex items-center gap-2">
                {#if selectedFiles.size > 0}
                    <button
                        on:click={deleteSelected}
                        class="flex items-center gap-1 px-2 py-1 bg-red-500/10 text-red-400 hover:bg-red-500 hover:text-white rounded-lg transition-colors text-xs font-bold mr-2"
                    >
                        <Trash2 class="w-4 h-4" />
                        Delete ({selectedFiles.size})
                    </button>

                    <button
                        on:click={compressSelected}
                        class="flex items-center gap-1 px-2 py-1 bg-blue-500/10 text-blue-400 hover:bg-blue-500 hover:text-white rounded-lg transition-colors text-xs font-bold mr-2"
                    >
                        <Archive class="w-4 h-4" />
                        Compress
                    </button>

                    <div class="h-4 w-px bg-white/10 mx-1"></div>
                {/if}

                {#if selectedFiles.size === 1 && (Array.from(selectedFiles)[0].endsWith(".zip") || Array.from(selectedFiles)[0].endsWith(".jar"))}
                    <button
                        on:click={extractSelected}
                        class="flex items-center gap-1 px-2 py-1 bg-green-500/10 text-green-400 hover:bg-green-500 hover:text-white rounded-lg transition-colors text-xs font-bold mr-2"
                    >
                        <ArchiveRestore class="w-4 h-4" />
                        Extract
                    </button>
                    <div class="h-4 w-px bg-white/10 mx-1"></div>
                {/if}

                <button
                    on:click={createFile}
                    class="p-2 hover:bg-white/10 rounded-lg text-gray-400 hover:text-white transition-colors"
                    title="New File"
                >
                    <FilePlus class="w-5 h-5" />
                </button>
                <button
                    on:click={createDirectory}
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

        <!-- File List -->
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
                        class="bg-white/5 text-xs uppercase text-gray-400 font-semibold sticky top-0 z-10 backdrop-blur-md"
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
                    <tbody
                        class="divide-y divide-white/5 text-sm text-gray-300"
                    >
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
                                        on:change={(e) =>
                                            toggleSelection(file, e)}
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
                                                class="w-5 h-5 text-gray-500 flex-shrink-0"
                                            />
                                        {/if}
                                        <span class="truncate">{file.name}</span
                                        >
                                    </button>
                                </td>
                                <td
                                    class="px-4 py-2 text-gray-500 font-mono text-xs"
                                    >{file.isDir
                                        ? "-"
                                        : formatSize(file.size)}</td
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
    {/if}
</div>
