<script>
    import { addToast } from "$lib/stores/toast.js";
    import { Play, Square, RotateCcw } from "lucide-svelte";

    export let status = "Offline";

    /** @type {string} */
    export let instanceId;

    /** @param {string} action */
    async function trigger(action) {
        console.log("Controls: Triggering", action);
        try {
            const res = await fetch(`/api/instances/${instanceId}/${action}`, {
                method: "POST",
            });
            if (!res.ok) throw new Error(await res.text());
            addToast(`Instance ${action}ed successfully`, "success");
        } catch (e) {
            const message = e instanceof Error ? e.message : String(e);
            console.error("Controls Trigger Error:", e);
            addToast("Failed to " + action + ": " + message, "error");
        }
    }

    $: isRunning = status === "Online" || status === "Starting";
</script>

<div class="flex gap-2 items-center">
    {#if !isRunning}
        <button
            on:click={() => trigger("start")}
            class="cursor-pointer p-3 bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-colors hover:shadow-lg hover:shadow-blue-900/20 group relative z-50"
            title="Start Server"
        >
            <Play class="w-5 h-5 fill-current" />
        </button>
    {:else}
        <button
            on:click={() => trigger("stop")}
            class="cursor-pointer p-3 bg-red-500/10 hover:bg-red-500/20 text-red-500 rounded-lg transition-colors group border border-red-500/20 relative z-50"
            title="Stop Server"
        >
            <Square class="w-5 h-5 fill-current" />
        </button>
    {/if}

    <button
        on:click={() => trigger("restart")}
        class="p-3 bg-gray-800 hover:bg-gray-700 text-gray-400 hover:text-white rounded-lg transition-colors border border-gray-700"
        title="Restart Server"
    >
        <RotateCcw class="w-5 h-5" />
    </button>
</div>
