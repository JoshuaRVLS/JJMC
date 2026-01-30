<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import CodeEditor from "$lib/components/CodeEditor.svelte";

    /** @type {string} */
    export let instanceId;

    /**
     * @typedef {Object} Property
     * @property {string} key
     * @property {string} value
     * @property {string} type
     * @property {string} originalValue
     */

    /** @type {Property[]} */
    let properties = [];
    let loading = true;
    let isRawMode = false;
    let rawContent = "";

    // Maintain raw content syncing
    let originalRawContent = "";

    // Helper to determine input type
    /**
     * @param {string} key
     * @param {string} value
     */
    function getInputType(key, value) {
        if (value === "true" || value === "false") return "boolean";
        if (
            !isNaN(parseFloat(value)) &&
            isFinite(Number(value)) &&
            value.trim() !== ""
        )
            return "number";
        return "text";
    }

    async function loadProperties() {
        loading = true;
        try {
            const res = await fetch(
                `/api/instances/${instanceId}/files/content?path=server.properties`,
            );
            if (res.ok) {
                const text = await res.text();
                rawContent = text;
                originalRawContent = text;
                parseProperties(text);
            } else {
                // If not found, maybe just empty or show error
                if (res.status === 404) {
                    addToast("server.properties not found", "error");
                } else {
                    addToast("Failed to load properties", "error");
                }
            }
        } catch (e) {
            console.error(e);
            addToast("Error loading properties", "error");
        } finally {
            loading = false;
        }
    }

    /** @param {string} text */
    function parseProperties(text) {
        const lines = text.split("\n");
        /** @type {Property[]} */
        const parsed = [];

        lines.forEach((line) => {
            line = line.trim();
            if (!line || line.startsWith("#")) return;

            const idx = line.indexOf("=");
            if (idx !== -1) {
                const key = line.substring(0, idx).trim();
                const value = line.substring(idx + 1).trim();
                parsed.push({
                    key,
                    value,
                    type: getInputType(key, value),
                    originalValue: value,
                });
            }
        });

        parsed.sort((a, b) => a.key.localeCompare(b.key));
        properties = parsed;
    }

    function syncToRaw() {
        // Reconstruct raw content from properties
        // We try to preserve the header from original if possible, removing old props
        // But for simplicity and correctness, we generate a fresh structure
        // A robust solution would parse the original AST and inject values.
        // Here we just regenerate.

        let content =
            "#Minecraft server properties\n#" + new Date().toISOString() + "\n";
        properties.forEach((p) => {
            content += `${p.key}=${p.value}\n`;
        });
        rawContent = content;
    }

    function syncToGui() {
        // Parse rawContent content back to properties
        parseProperties(rawContent);
    }

    async function saveProperties() {
        // Sync current View to Payload
        let contentToSend = "";
        if (isRawMode) {
            contentToSend = rawContent;
            // efficient: also sync back to GUI state
            syncToGui();
        } else {
            syncToRaw();
            contentToSend = rawContent;
        }

        try {
            const res = await fetch(
                `/api/instances/${instanceId}/files/content`,
                {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        path: "server.properties",
                        content: contentToSend,
                    }),
                },
            );

            if (res.ok) {
                addToast("Properties saved", "success");
            } else {
                addToast("Failed to save", "error");
            }
        } catch (e) {
            addToast("Error saving", "error");
        }
    }

    onMount(() => {
        loadProperties();
    });
</script>

<div
    class="h-full flex flex-col bg-gray-900/50 rounded-xl overflow-hidden border border-white/5"
>
    <!-- Toolbar -->
    <div
        class="flex items-center justify-between px-4 py-3 bg-white/5 border-b border-white/5"
    >
        <div class="flex items-center gap-4">
            <div class="text-sm font-bold text-gray-300">server.properties</div>

            <div class="flex bg-black/40 p-1 rounded-lg">
                <button
                    class="px-3 py-1 rounded text-[10px] font-bold uppercase transition-all {!isRawMode
                        ? 'bg-indigo-600 text-white'
                        : 'text-gray-500 hover:text-white'}"
                    on:click={() => {
                        if (isRawMode) syncToGui();
                        isRawMode = false;
                    }}
                >
                    GUI
                </button>
                <button
                    class="px-3 py-1 rounded text-[10px] font-bold uppercase transition-all {isRawMode
                        ? 'bg-indigo-600 text-white'
                        : 'text-gray-500 hover:text-white'}"
                    on:click={() => {
                        if (!isRawMode) syncToRaw();
                        isRawMode = true;
                    }}
                >
                    Raw
                </button>
            </div>
        </div>

        <button
            on:click={saveProperties}
            class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-lg transition-colors"
        >
            Save Changes
        </button>
    </div>

    <!-- Scrollable Content -->
    <div class="flex-1 overflow-y-auto relative">
        {#if loading}
            <div class="flex items-center justify-center h-full text-gray-500">
                Loading...
            </div>
        {:else if properties.length === 0 && !isRawMode}
            <div
                class="flex flex-col items-center justify-center h-full text-gray-500 gap-2"
            >
                <div>No properties found or file missing.</div>
                <button
                    on:click={loadProperties}
                    class="text-indigo-400 hover:text-indigo-300 text-sm underline"
                >
                    Retry
                </button>
            </div>
        {:else if isRawMode}
            <CodeEditor bind:value={rawContent} language="properties" />
        {:else}
            <div
                class="p-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
            >
                {#each properties as prop, i}
                    <div
                        class="bg-white/5 p-3 rounded-lg border border-white/5 hover:border-white/10 transition-colors"
                    >
                        <label
                            class="block text-xs font-mono text-gray-400 mb-1 truncate"
                            title={prop.key}
                            for="prop-{i}"
                        >
                            {prop.key}
                        </label>

                        {#if prop.type === "boolean"}
                            <div class="flex items-center gap-2 mt-1">
                                <button
                                    id="prop-{i}"
                                    class="w-10 h-5 rounded-full relative transition-colors {prop.value ===
                                    'true'
                                        ? 'bg-indigo-500'
                                        : 'bg-gray-700'}"
                                    aria-label="Toggle {prop.key}"
                                    on:click={() =>
                                        (prop.value =
                                            prop.value === "true"
                                                ? "false"
                                                : "true")}
                                >
                                    <div
                                        class="absolute top-1 left-1 w-3 h-3 bg-white rounded-full transition-transform {prop.value ===
                                        'true'
                                            ? 'translate-x-5'
                                            : ''}"
                                    ></div>
                                </button>
                                <span class="text-xs text-gray-300"
                                    >{prop.value}</span
                                >
                            </div>
                        {:else if prop.type === "number"}
                            <input
                                id="prop-{i}"
                                type="number"
                                class="w-full bg-black/20 border border-white/10 rounded px-2 py-1 text-sm text-gray-200 focus:outline-none focus:border-indigo-500/50"
                                bind:value={prop.value}
                            />
                        {:else}
                            <input
                                id="prop-{i}"
                                type="text"
                                class="w-full bg-black/20 border border-white/10 rounded px-2 py-1 text-sm text-gray-200 focus:outline-none focus:border-indigo-500/50"
                                bind:value={prop.value}
                            />
                        {/if}
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>
