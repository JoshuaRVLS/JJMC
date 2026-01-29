<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";

    export let instanceId;

    let properties = [];
    let loading = true;

    // Helper to determine input type
    function getInputType(key, value) {
        if (value === "true" || value === "false") return "boolean";
        if (!isNaN(value) && value.trim() !== "") return "number";
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

    function parseProperties(text) {
        const lines = text.split("\n");
        const parsed = [];

        lines.forEach((line) => {
            line = line.trim();
            if (!line || line.startsWith("#")) return; // Skip comments and empty lines for now or store them?
            // Storing comments complicates things for a simple UI editor. Let's stick to key-values.

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

        // Sort effectively? Maybe active ones first, or alphabetical
        parsed.sort((a, b) => a.key.localeCompare(b.key));
        properties = parsed;
    }

    async function saveProperties() {
        try {
            // Reconstruct file
            // Note: This wipes comments. A true editor would preserve them.
            // For this version (User Request: "change to setting server.properties"), we focus on functionality.
            // If we want to preserve header we can add standard header.

            let content =
                "#Minecraft server properties\n#" +
                new Date().toISOString() +
                "\n";
            properties.forEach((p) => {
                content += `${p.key}=${p.value}\n`;
            });

            const res = await fetch(
                `/api/instances/${instanceId}/files/content`,
                {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        path: "server.properties",
                        content: content,
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
        <div class="text-sm font-bold text-gray-300">server.properties</div>
        <button
            on:click={saveProperties}
            class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-lg transition-colors"
        >
            Save Changes
        </button>
    </div>

    <!-- Scrollable Content -->
    <div class="flex-1 overflow-y-auto p-4">
        {#if loading}
            <div class="flex items-center justify-center h-full text-gray-500">
                Loading...
            </div>
        {:else if properties.length === 0}
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
        {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {#each properties as prop}
                    <div
                        class="bg-white/5 p-3 rounded-lg border border-white/5 hover:border-white/10 transition-colors"
                    >
                        <label
                            class="block text-xs font-mono text-gray-400 mb-1 truncate"
                            title={prop.key}
                        >
                            {prop.key}
                        </label>

                        {#if prop.type === "boolean"}
                            <div class="flex items-center gap-2 mt-1">
                                <button
                                    class="w-10 h-5 rounded-full relative transition-colors {prop.value ===
                                    'true'
                                        ? 'bg-indigo-500'
                                        : 'bg-gray-700'}"
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
                                    />
                                </button>
                                <span class="text-xs text-gray-300"
                                    >{prop.value}</span
                                >
                            </div>
                        {:else if prop.type === "number"}
                            <input
                                type="number"
                                class="w-full bg-black/20 border border-white/10 rounded px-2 py-1 text-sm text-gray-200 focus:outline-none focus:border-indigo-500/50"
                                bind:value={prop.value}
                            />
                        {:else}
                            <input
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
