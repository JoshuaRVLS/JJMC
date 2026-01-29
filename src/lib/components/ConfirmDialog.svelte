<script>
    import { confirmState, closeConfirm } from "../stores/confirm.js";
    import { fade, scale } from "svelte/transition";
</script>

{#if $confirmState.active}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
        transition:fade={{ duration: 200 }}
    >
        <div
            class="w-full max-w-md bg-gray-900 border border-white/10 rounded-2xl shadow-2xl p-6"
            transition:scale={{ start: 0.95, duration: 200 }}
        >
            <h3 class="text-xl font-bold text-white mb-2">
                {$confirmState.title}
            </h3>
            <p class="text-gray-400 mb-6 text-sm">{$confirmState.message}</p>

            <div class="flex justify-end gap-3">
                <button
                    class="px-4 py-2 text-sm font-medium text-gray-400 hover:text-white transition-colors"
                    on:click={() => closeConfirm(false)}
                >
                    {$confirmState.cancelText}
                </button>
                <button
                    class="px-4 py-2 text-sm font-bold text-white rounded-lg transition-all shadow-lg
                    {$confirmState.dangerous
                        ? 'bg-red-600 hover:bg-red-500 shadow-red-500/20'
                        : 'bg-indigo-600 hover:bg-indigo-500 shadow-indigo-500/20'}"
                    on:click={() => closeConfirm(true)}
                >
                    {$confirmState.confirmText}
                </button>
            </div>
        </div>
    </div>
{/if}
