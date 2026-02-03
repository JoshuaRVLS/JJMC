<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import CodeEditor from "$lib/components/CodeEditor.svelte";

    export let instanceId;

    let content = "";
    let loading = true;

    async function loadConfig() {
        loading = true;
        try {
            const res = await fetch(
                `/api/instances/${instanceId}/files/content?path=velocity.toml`,
            );
            if (res.ok) {
                content = await res.text();
            } else {
                if (res.status === 404) {
                    addToast("velocity.toml not found", "error");
                    content =
                        "# velocity.toml not found. Start the server to generate it.";
                } else {
                    addToast("Failed to load configuration", "error");
                }
            }
        } catch (e) {
            console.error(e);
            addToast("Error loading configuration", "error");
        } finally {
            loading = false;
        }
    }

    async function saveConfig() {
        try {
            const res = await fetch(
                `/api/instances/${instanceId}/files/content`,
                {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        path: "velocity.toml",
                        content: content,
                    }),
                },
            );

            if (res.ok) {
                addToast("Configuration saved", "success");
            } else {
                addToast("Failed to save", "error");
            }
        } catch (e) {
            addToast("Error saving", "error");
        }
    }

    onMount(() => {
        loadConfig();
    });
</script>

<div
    class="h-full flex flex-col bg-gray-900/50 rounded-xl overflow-hidden border border-white/5"
>
    <div
        class="flex items-center justify-between px-4 py-3 bg-white/5 border-b border-white/5 gap-4"
    >
        <div class="flex items-center gap-4 flex-1">
            <div class="text-sm font-bold text-gray-300">velocity.toml</div>
            <div class="text-xs text-gray-500">Velocity Configuration</div>
        </div>

        <button
            on:click={saveConfig}
            class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-lg transition-colors shrink-0"
        >
            Save Changes
        </button>
    </div>

    <div class="flex-1 overflow-hidden relative">
        {#if loading}
            <div class="flex items-center justify-center h-full text-gray-500">
                Loading...
            </div>
        {:else}
            <CodeEditor bind:value={content} language="toml" />
        {/if}
    </div>
</div>
