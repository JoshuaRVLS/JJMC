<script>
    import { addToast } from "$lib/stores/toast.js";
    // Wait, I don't have unplugin-icons set up probably.
    // I should use SVG directly or see if I can use a library.
    // The previous code had SVGs inline or no icons.
    // I will use Inline SVGs to be safe and dependency-free.

    export let status = "Offline"; // Received from parent
    export let instanceId;

    async function trigger(action) {
        console.log("Controls: Triggering", action);
        try {
            const res = await fetch(`/api/instances/${instanceId}/${action}`, {
                method: "POST",
            });
            if (!res.ok) throw new Error(await res.text());
            addToast(`Instance ${action}ed successfully`, "success");
        } catch (e) {
            console.error("Controls Trigger Error:", e);
            addToast("Failed to " + action + ": " + e.message, "error");
        }
    }

    // Computed
    $: isRunning = status === "Online" || status === "Starting"; // Approximate check
</script>

<div class="flex gap-2 items-center">
    <!-- Toggle Start/Stop -->
    {#if !isRunning}
        <button
            on:click={() => trigger("start")}
            class="cursor-pointer p-3 bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-colors hover:shadow-lg hover:shadow-blue-900/20 group relative z-50"
            title="Start Server"
        >
            <!-- Play Icon -->
            <svg class="w-5 h-5 fill-current" viewBox="0 0 24 24"
                ><path d="M8 5v14l11-7z" /></svg
            >
        </button>
    {:else}
        <button
            on:click={() => trigger("stop")}
            class="cursor-pointer p-3 bg-red-500/10 hover:bg-red-500/20 text-red-500 rounded-lg transition-colors group border border-red-500/20 relative z-50"
            title="Stop Server"
        >
            <!-- Square Icon -->
            <svg class="w-5 h-5 fill-current" viewBox="0 0 24 24"
                ><path d="M6 6h12v12H6z" /></svg
            >
        </button>
    {/if}

    <!-- Restart (Always visible) -->
    <button
        on:click={() => trigger("restart")}
        class="p-3 bg-gray-800 hover:bg-gray-700 text-gray-400 hover:text-white rounded-lg transition-colors border border-gray-700"
        title="Restart Server"
    >
        <!-- Refresh Icon -->
        <svg
            class="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            stroke-width="2"
        >
            <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
            />
        </svg>
    </button>
</div>
