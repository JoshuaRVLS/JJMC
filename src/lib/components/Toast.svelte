<script>
    import { toasts, dismissToast } from "../stores/toast.js";
    import { flip } from "svelte/animate";
    import { fade, fly } from "svelte/transition";
</script>

<div
    class="fixed bottom-4 right-4 z-50 flex flex-col gap-2 p-4 max-w-sm w-full pointer-events-none"
>
    {#each $toasts as toast (toast.id)}
        <div
            animate:flip
            in:fly={{ y: 20, duration: 300 }}
            out:fade
            class="pointer-events-auto flex items-center gap-3 px-4 py-3 rounded-xl shadow-2xl border backdrop-blur-md transform transition-all
            {toast.type === 'success'
                ? 'bg-emerald-900/80 border-emerald-500/30 text-emerald-100'
                : toast.type === 'error'
                  ? 'bg-red-900/80 border-red-500/30 text-red-100'
                  : 'bg-gray-800/80 border-gray-600/30 text-gray-100'}"
        >
            <div class="flex-1 text-sm font-medium">{toast.message}</div>
            <button
                class="opacity-50 hover:opacity-100 transition-opacity"
                on:click={() => dismissToast(toast.id)}
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
                        d="M6 18L18 6M6 6l12 12"
                    /></svg
                >
            </button>
        </div>
    {/each}
</div>
