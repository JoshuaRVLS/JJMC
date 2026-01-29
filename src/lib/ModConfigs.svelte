<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import { askConfirm } from "$lib/stores/confirm";

    export let instanceId;

    let files = [];
    let loading = false;
    let viewingFile = null;
    let fileContent = "";
    let parsedData = null;
    let query = "";
    let isGuiMode = true;
    let parseError = null;

    async function loadConfigs() {
        loading = true;
        try {
            // We specifically target the 'config' directory
            const res = await fetch(
                `/api/instances/${instanceId}/files?path=config`,
            );
            if (res.ok) {
                files = await res.json();
            } else {
                // If config dir doesn't exist, it might be empty
                files = [];
            }
        } catch (e) {
            console.error(e);
            addToast("Error loading configs", "error");
        } finally {
            loading = false;
        }
    }

    async function openFile(file) {
        if (file.isDir) return; // Only files for now in configs view

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
                attemptParse();
            } else {
                addToast("Failed to read file", "error");
            }
        } catch (e) {
            addToast("Error reading file", "error");
        }
    }

    function attemptParse() {
        parseError = null;
        try {
            // Try JSON first
            if (
                viewingFile.name.endsWith(".json") ||
                viewingFile.name.endsWith(".json5")
            ) {
                parsedData = JSON.parse(fileContent);
                isGuiMode = true;
            } else {
                // For other files, maybe just treat as text for now or try simple KV
                isGuiMode = false;
                parsedData = null;
            }
        } catch (e) {
            parseError = e.message;
            isGuiMode = false;
            parsedData = null;
        }
    }

    async function saveFile() {
        if (!viewingFile) return;

        let contentToSave = fileContent;
        if (isGuiMode && parsedData !== null) {
            contentToSave = JSON.stringify(parsedData, null, 4);
        }

        try {
            const res = await fetch(
                `/api/instances/${instanceId}/files/content`,
                {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        path: viewingFile.fullPath,
                        content: contentToSave,
                    }),
                },
            );
            if (res.ok) {
                addToast("Config saved", "success");
                fileContent = contentToSave; // Sync back
            } else {
                addToast("Failed to save config", "error");
            }
        } catch (e) {
            addToast("Error saving config", "error");
        }
    }

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

    onMount(() => {
        loadConfigs();
    });
</script>

{#snippet editorField(label, value, parent, key)}
    <div
        class="flex flex-col gap-2 p-3 bg-white/5 rounded-lg border border-white/5 hover:border-white/10 transition-colors"
    >
        <label
            class="text-[10px] font-bold text-gray-500 uppercase tracking-widest"
            >{label}</label
        >

        {#if typeof value === "boolean"}
            <div class="flex items-center gap-3">
                <button
                    class="w-10 h-5 rounded-full relative transition-colors {value
                        ? 'bg-indigo-500'
                        : 'bg-gray-700'}"
                    on:click={() => (parent[key] = !value)}
                >
                    <div
                        class="absolute top-1 left-1 w-3 h-3 bg-white rounded-full transition-transform {value
                            ? 'translate-x-5'
                            : ''}"
                    ></div>
                </button>
                <span class="text-xs text-gray-400 capitalize">{value}</span>
            </div>
        {:else if typeof value === "number"}
            <input
                type="number"
                bind:value={parent[key]}
                class="bg-black/40 border border-white/10 rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-indigo-500/50 transition-colors"
            />
        {:else if typeof value === "string"}
            <input
                type="text"
                bind:value={parent[key]}
                class="bg-black/40 border border-white/10 rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-indigo-500/50 transition-colors"
            />
        {/if}
    </div>
{/snippet}

{#snippet editorObject(obj, depth = 0)}
    <div class="flex flex-col gap-4">
        {#each Object.entries(obj) as [key, value]}
            {#if value !== null && typeof value === "object"}
                {#if !Array.isArray(value)}
                    <div
                        class="flex flex-col gap-3 {depth > 0
                            ? 'ml-4 pl-4 border-l border-white/5'
                            : ''}"
                    >
                        <div class="flex items-center gap-2">
                            <span
                                class="text-xs font-black text-indigo-400 uppercase tracking-tighter"
                                >{key}</span
                            >
                            <div class="h-px flex-1 bg-white/5"></div>
                        </div>
                        {@render editorObject(value, depth + 1)}
                    </div>
                {:else}
                    <!-- Array handling: simplified -->
                    <div
                        class="flex flex-col gap-3 {depth > 0
                            ? 'ml-4 pl-4 border-l border-white/5'
                            : ''}"
                    >
                        <div class="flex items-center justify-between">
                            <span
                                class="text-xs font-black text-emerald-400 uppercase tracking-tighter"
                                >{key} (Array)</span
                            >
                        </div>
                        <div class="grid grid-cols-1 gap-2">
                            {#each value as item, i}
                                {@render editorField(
                                    `Item ${i}`,
                                    item,
                                    value,
                                    i,
                                )}
                            {/each}
                        </div>
                    </div>
                {/if}
            {:else}
                {@render editorField(key, value, obj, key)}
            {/if}
        {/each}
    </div>
{/snippet}

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
                            >Configuration Editor</span
                        >
                    </div>

                    {#if parsedData}
                        <div class="flex bg-black/40 p-1 rounded-lg ml-2">
                            <button
                                class="px-3 py-1 rounded text-[10px] font-bold uppercase transition-all {isGuiMode
                                    ? 'bg-indigo-600 text-white'
                                    : 'text-gray-500 hover:text-white'}"
                                on:click={() => (isGuiMode = true)}
                            >
                                GUI
                            </button>
                            <button
                                class="px-3 py-1 rounded text-[10px] font-bold uppercase transition-all {!isGuiMode
                                    ? 'bg-indigo-600 text-white'
                                    : 'text-gray-500 hover:text-white'}"
                                on:click={() => (isGuiMode = false)}
                            >
                                Raw
                            </button>
                        </div>
                    {/if}
                </div>

                <button
                    on:click={saveFile}
                    class="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-lg transition-all shadow-lg shadow-indigo-600/20"
                >
                    Save Changes
                </button>
            </div>

            <div class="flex-1 overflow-y-auto custom-scrollbar">
                {#if isGuiMode && parsedData}
                    <div class="p-6 max-w-4xl mx-auto w-full">
                        {@render editorObject(parsedData)}
                    </div>
                {:else}
                    <textarea
                        class="w-full h-full bg-[#0b0e14] text-gray-300 font-mono text-sm p-6 focus:outline-none resize-none selection:bg-indigo-500/30"
                        bind:value={fileContent}
                        on:input={() => {
                            if (!isGuiMode) attemptParse();
                        }}
                        spellcheck="false"
                    ></textarea>
                {/if}
            </div>

            {#if parseError && !isGuiMode}
                <div
                    class="bg-rose-500/10 border-t border-rose-500/20 px-4 py-2 flex items-center gap-3"
                >
                    <div
                        class="w-1.5 h-1.5 rounded-full bg-rose-500 animate-pulse"
                    ></div>
                    <span class="text-[10px] font-mono text-rose-400 truncate"
                        >JSON Parse Error: {parseError}</span
                    >
                </div>
            {/if}
        </div>
    {:else}
        <div class="flex flex-col gap-4 h-full">
            <!-- Search -->
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

            <!-- List -->
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
