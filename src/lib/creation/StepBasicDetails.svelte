<script>
    import { fade } from "svelte/transition";
    import { Server, FolderInput, Check, ArrowRight } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";

    /** @type {string} */
    export let name;

    /** @type {boolean} */
    export let importMode;

    const dispatch = createEventDispatcher();

    function handleNext() {
        dispatch("next");
    }
</script>

<div
    in:fade={{ duration: 200, delay: 100 }}
    class="flex-1 flex flex-col justify-center max-w-lg mx-auto w-full"
>
    <h2 class="text-3xl font-bold text-white mb-6">Let's start.</h2>

    <div class="space-y-6">
        <div>
            <label
                for="name"
                class="block text-sm font-medium text-gray-400 mb-2"
                >Server Name</label
            >
            <input
                type="text"
                id="name"
                bind:value={name}
                class="w-full bg-gray-800 border border-gray-700 text-white rounded-xl px-4 py-3 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-lg placeholder-gray-600"
                placeholder="My Awesome Server"
                on:keydown={(e) => e.key === "Enter" && handleNext()}
            />
        </div>

        <div class="grid grid-cols-2 gap-4">
            <button
                class="relative p-4 rounded-xl border transition-all text-left group
                {!importMode
                    ? 'bg-blue-600/10 border-blue-500 ring-1 ring-blue-500/50'
                    : 'bg-gray-800/50 border-gray-700 hover:bg-gray-800 hover:border-gray-600'}"
                on:click={() => (importMode = false)}
            >
                <div
                    class="p-2 w-10 h-10 rounded-lg {!importMode
                        ? 'bg-blue-500 text-white'
                        : 'bg-gray-700 text-gray-400'} flex items-center justify-center mb-3 transition-colors"
                >
                    <Server size={20} />
                </div>
                <h3 class="text-white font-medium mb-1">New Server</h3>
                <p class="text-xs text-gray-400 leading-relaxed">
                    Install a fresh Minecraft server from a template.
                </p>

                {#if !importMode}
                    <div class="absolute top-4 right-4 text-blue-400">
                        <Check size={16} />
                    </div>
                {/if}
            </button>

            <button
                class="relative p-4 rounded-xl border transition-all text-left group
                {importMode
                    ? 'bg-emerald-600/10 border-emerald-500 ring-1 ring-emerald-500/50'
                    : 'bg-gray-800/50 border-gray-700 hover:bg-gray-800 hover:border-gray-600'}"
                on:click={() => (importMode = true)}
            >
                <div
                    class="p-2 w-10 h-10 rounded-lg {importMode
                        ? 'bg-emerald-500 text-white'
                        : 'bg-gray-700 text-gray-400'} flex items-center justify-center mb-3 transition-colors"
                >
                    <FolderInput size={20} />
                </div>
                <h3 class="text-white font-medium mb-1">Import Existing</h3>
                <p class="text-xs text-gray-400 leading-relaxed">
                    Add a server that already exists on disk.
                </p>

                {#if importMode}
                    <div class="absolute top-4 right-4 text-emerald-400">
                        <Check size={16} />
                    </div>
                {/if}
            </button>
        </div>
    </div>

    <div class="mt-8 flex justify-end">
        <button
            on:click={handleNext}
            class="bg-white text-gray-900 px-6 py-3 rounded-xl font-bold hover:bg-gray-200 transition-colors flex items-center gap-2 shadow-lg shadow-white/10"
        >
            Next Step <ArrowRight size={18} />
        </button>
    </div>
</div>
