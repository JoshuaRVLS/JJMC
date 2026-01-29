<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import { askConfirm } from "$lib/stores/confirm";
    import { askInput } from "$lib/stores/input";

    export let instanceId;

    let files = [];
    let currentPath = ".";
    let loading = false;
    let viewingFile = null;
    let fileContent = "";
    let isEditing = false;
    let breadcrumbs = [];

    async function loadFiles(path = ".") {
        loading = true;
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

    function formatSize(bytes) {
        if (bytes === 0) return "0 B";
        const k = 1024;
        const sizes = ["B", "KB", "MB", "GB", "TB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
    }

    function formatDate(ms) {
        return new Date(ms).toLocaleString();
    }

    async function openFile(file) {
        if (file.isDir) {
            const newPath =
                currentPath === "." ? file.name : `${currentPath}/${file.name}`;
            loadFiles(newPath);
        } else {
            // Check if binary or too large? For now just try to read text
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
                    isEditing = false;
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
                isEditing = false;
            } else {
                addToast("Failed to save file", "error");
            }
        } catch (e) {
            addToast("Error saving file", "error");
        }
    }

    async function deleteFile(file) {
        const confirmed = await askConfirm({
            title: "Delete File",
            message: `Are you sure you want to delete ${file.name}?`,
            confirmText: "Delete",
            dangerous: true,
        });

        if (!confirmed) return;

        try {
            const path =
                currentPath === "." ? file.name : `${currentPath}/${file.name}`;
            const res = await fetch(
                `/api/instances/${instanceId}/files?path=${encodeURIComponent(
                    path,
                )}`,
                { method: "DELETE" },
            );

            if (res.ok) {
                addToast("Deleted successfully", "success");
                loadFiles(currentPath);
            } else {
                addToast("Failed to delete", "error");
            }
        } catch (e) {
            addToast("Error deleting", "error");
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

    async function handleFileUpload(e) {
        const fileList = e.target.files;
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
            e.target.value = ""; // Reset input
        }
    }

    onMount(() => {
        loadFiles();
    });
</script>

<div
    class="h-full flex flex-col bg-gray-900/50 rounded-xl overflow-hidden border border-white/5"
>
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
                        <svg
                            class="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                            ><path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M10 19l-7-7m0 0l7-7m-7 7h18"
                            /></svg
                        >
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
                    <svg
                        class="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        ><path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M5 10l7-7m0 0l7 7m-7-7v18"
                        /></svg
                    >
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
                <button
                    on:click={createFile}
                    class="p-2 hover:bg-white/10 rounded-lg text-gray-400 hover:text-white transition-colors"
                    title="New File"
                >
                    <svg
                        class="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        ><path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M9 13h6m-3-3v6m5 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                        /></svg
                    >
                </button>
                <button
                    on:click={createDirectory}
                    class="p-2 hover:bg-white/10 rounded-lg text-gray-400 hover:text-white transition-colors"
                    title="New Folder"
                >
                    <svg
                        class="w-5 h-5"
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
                    <svg
                        class="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        ><path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
                        /></svg
                    >
                </label>
            </div>
        </div>

        <!-- File List -->
        <div class="flex-1 overflow-y-auto">
            {#if loading}
                <div class="flex items-center justify-center h-40">
                    <svg
                        class="animate-spin h-8 w-8 text-indigo-500"
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
                            <th class="px-4 py-3">Name</th>
                            <th class="px-4 py-3 w-32">Size</th>
                            <th class="px-4 py-3 w-48">Modified</th>
                            <th class="px-4 py-3 w-20 text-right">Action</th>
                        </tr>
                    </thead>
                    <tbody
                        class="divide-y divide-white/5 text-sm text-gray-300"
                    >
                        {#each files as file}
                            <tr
                                class="hover:bg-white/5 transition-colors group"
                            >
                                <td class="px-4 py-2">
                                    <button
                                        class="flex items-center gap-3 hover:text-white w-full text-left truncate"
                                        on:click={() => openFile(file)}
                                    >
                                        {#if file.isDir}
                                            <svg
                                                class="w-5 h-5 text-yellow-500 flex-shrink-0"
                                                fill="currentColor"
                                                viewBox="0 0 20 20"
                                                ><path
                                                    d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"
                                                /></svg
                                            >
                                        {:else}
                                            <svg
                                                class="w-5 h-5 text-gray-500 flex-shrink-0"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                                ><path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"
                                                /></svg
                                            >
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
                                <td class="px-4 py-2 text-right">
                                    <button
                                        on:click={() => deleteFile(file)}
                                        class="text-gray-600 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-all"
                                        title="Delete"
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
                                                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                                            /></svg
                                        >
                                    </button>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            {/if}
        </div>
    {/if}
</div>
