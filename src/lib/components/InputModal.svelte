<script>
    import { inputState, closeInput } from "../stores/input.js";
    import { fade, scale } from "svelte/transition";
    import { tick } from "svelte";

     
    let inputEl;

     
    $: if ($inputState.active && inputEl) {
        tick().then(() =>   (inputEl).focus());
    }

     
    function handleKeydown(e) {
        if (e.key === "Enter") {
            closeInput($inputState.value);
        } else if (e.key === "Escape") {
            closeInput(null);
        }
    }
</script>

{#if $inputState.active}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
        transition:fade={{ duration: 200 }}
    >
        <div
            class="w-full max-w-md bg-gray-900 border border-white/10 rounded-2xl shadow-2xl p-6"
            transition:scale={{ start: 0.95, duration: 200 }}
        >
            <h3 class="text-xl font-bold text-white mb-2">
                {$inputState.title}
            </h3>
            {#if $inputState.message}
                <p class="text-gray-400 mb-4 text-sm">{$inputState.message}</p>
            {/if}

            <input
                bind:this={inputEl}
                type="text"
                class="w-full bg-black/30 border border-white/10 rounded-lg px-4 py-2 text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500 transition-colors mb-6"
                placeholder={$inputState.placeholder}
                bind:value={$inputState.value}
                on:keydown={handleKeydown}
            />

            <div class="flex justify-end gap-3">
                <button
                    class="px-4 py-2 text-sm font-medium text-gray-400 hover:text-white transition-colors"
                    on:click={() => closeInput(null)}
                >
                    {$inputState.cancelText}
                </button>
                <button
                    class="px-4 py-2 text-sm font-bold text-white rounded-lg transition-all shadow-lg bg-indigo-600 hover:bg-indigo-500 shadow-indigo-500/20"
                    on:click={() => closeInput($inputState.value)}
                >
                    {$inputState.confirmText}
                </button>
            </div>
        </div>
    </div>
{/if}
