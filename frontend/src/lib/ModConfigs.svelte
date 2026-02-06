<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import CodeEditor from "$lib/components/CodeEditor.svelte";

    /** @type {string} */
    export let instanceId;

    /** @type {any[]} */
    let files = [];
    let loading = false;

    /** @type {any} */
    let viewingFile = null;
    let fileContent = "";
    let query = "";

    async function loadConfigs() {
        loading = true;
        try {
            const res = await fetch(
                `/api/instances/${instanceId}/files?path=config`,
            );
            if (res.ok) {
                files = await res.json();
            } else {
                files = [];
            }
        } catch (e) {
            console.error(e);
            addToast("Error loading configs", "error");
        } finally {
            loading = false;
        }
    }

    /** @param {any} file */
    async function openFile(file) {
        if (file.isDir) return;

        try {
            const path = `config/${file.name}`;
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
                addToast("Config saved", "success");
            } else {
                addToast("Failed to save config", "error");
            }
        } catch (e) {
            addToast("Error saving config", "error");
        }
    }

    /** @param {number} bytes */
    function formatSize(bytes) {
        if (bytes === 0) return "0 B";
        const k = 1024;
        const sizes = ["B", "KB", "MB", "GB", "TB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
    }

    $: filteredFiles = files.filter(
        (f) => !f.isDir && f.name.toLowerCase().includes(query.toLowerCase()),
    );

    let language = "json";

    $: if (viewingFile) {
        if (
            viewingFile.name.endsWith(".json") ||
            viewingFile.name.endsWith(".json5")
        ) {
            language = "json";
        } else if (viewingFile.name.endsWith(".toml")) {
            language = "toml";
        } else if (
            viewingFile.name.endsWith(".properties") ||
            viewingFile.name.endsWith(".ini")
        ) {
            language = "properties";
        } else {
            language = "json";
        }
    }

    onMount(() => {
        loadConfigs();
    });
</script>

<div class="h-full flex flex-col gap-4">
    {#if viewingFile}
        <div
            class="h-full flex flex-col bg-gray-900/50 rounded-xl overflow-hidden border border-white/5"
        >
            <div
                class="flex items-center justify-between px-4 py-3 bg-white/5 border-b border-white/5"
            >
                <div class="flex items-center gap-4">
                    <button
                        on:click={() => (viewingFile = null)}
                        class="text-gray-400 hover:text-white transition-colors"
                        aria-label="Back to file list"
                    >
                        <svg
                            class="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M10 19l-7-7m0 0l7-7m-7 7h18"
                            />
                        </svg>
                    </button>
                    <div class="flex flex-col">
                        <span class="font-mono text-xs text-white leading-none"
                            >{viewingFile.name}</span
                        >
                        <span class="text-[10px] text-gray-500 uppercase mt-0.5"
                            >Configuration Editor (Raw)</span
                        >
                    </div>
                </div>

                <button
                    on:click={saveFile}
                    class="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-lg transition-all shadow-lg shadow-indigo-600/20"
                >
                    Save Changes
                </button>
            </div>

            <div class="flex-1 relative">
                <CodeEditor bind:value={fileContent} {language} />
            </div>
        </div>
    {:else}
        <div class="flex flex-col gap-4 h-full">
            <div class="relative">
                <input
                    type="text"
                    bind:value={query}
                    placeholder="Search configuration files..."
                    class="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-3 pl-11 text-white focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all"
                />
                <svg
                    class="w-5 h-5 text-gray-500 absolute left-3.5 top-3.5"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                    />
                </svg>
            </div>

            <div
                class="flex-1 bg-gray-900/50 rounded-xl border border-white/5 overflow-hidden flex flex-col"
            >
                <div class="overflow-y-auto flex-1 custom-scrollbar">
                    {#if loading}
                        <div class="flex items-center justify-center h-40">
                            <div
                                class="animate-spin h-8 w-8 border-4 border-indigo-500 border-t-transparent rounded-full"
                            ></div>
                        </div>
                    {:else if filteredFiles.length === 0}
                        <div
                            class="flex flex-col items-center justify-center h-64 text-gray-500 gap-4"
                        >
                            <div class="p-4 rounded-full bg-white/5">
                                <svg
                                    class="w-8 h-8 opacity-20"
                                    fill="none"
                                    stroke="currentColor"
                                    viewBox="0 0 24 24"
                                >
                                    <path
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        stroke-width="2"
                                        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                                    />
                                </svg>
                            </div>
                            <span class="text-xs"
                                >{query
                                    ? "No matching files found"
                                    : "No configuration files found in /config"}</span
                            >
                        </div>
                    {:else}
                        <div
                            class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 p-4"
                        >
                            {#each filteredFiles as file}
                                <button
                                    on:click={() => openFile(file)}
                                    class="flex flex-col gap-2 p-4 bg-white/5 hover:bg-white/10 border border-white/5 hover:border-indigo-500/30 rounded-xl transition-all text-left group"
                                >
                                    <div
                                        class="flex items-start justify-between"
                                    >
                                        <div
                                            class="p-2 rounded-lg bg-indigo-500/10 text-indigo-400 group-hover:bg-indigo-500 group-hover:text-white transition-all"
                                        >
                                            {#if file.name.endsWith(".json") || file.name.endsWith(".json5")}
                                                <svg
                                                    class="w-5 h-5"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    viewBox="0 0 24 24"
                                                >
                                                    <path
                                                        stroke-linecap="round"
                                                        stroke-linejoin="round"
                                                        stroke-width="2"
                                                        d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
                                                    />
                                                </svg>
                                            {:else}
                                                <svg
                                                    class="w-5 h-5"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    viewBox="0 0 24 24"
                                                >
                                                    <path
                                                        stroke-linecap="round"
                                                        stroke-linejoin="round"
                                                        stroke-width="2"
                                                        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                                                    />
                                                </svg>
                                            {/if}
                                        </div>
                                        <span
                                            class="text-[10px] font-mono text-gray-500 uppercase"
                                            >{file.name.split(".").pop()}</span
                                        >
                                    </div>
                                    <div class="mt-1">
                                        <div
                                            class="text-sm font-bold text-white truncate w-full"
                                            title={file.name}
                                        >
                                            {file.name}
                                        </div>
                                        <div
                                            class="text-[10px] text-gray-500 mt-0.5"
                                        >
                                            {formatSize(file.size)}
                                        </div>
                                    </div>
                                </button>
                            {/each}
                        </div>
                    {/if}
                </div>
            </div>
        </div>
    {/if}
</div>
