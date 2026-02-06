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
    <div class="mb-8">
        <h2 class="text-2xl font-bold text-white mb-2 tracking-tight">
            Let's get started
        </h2>
        <p class="text-gray-400 text-sm">
            Choose a name and how you want to create your server.
        </p>
    </div>

    <div class="space-y-6">
        <div>
            <label
                for="name"
                class="block text-xs font-bold text-gray-500 uppercase tracking-widest mb-2 pl-1"
                >Server Name</label
            >
            <input
                type="text"
                id="name"
                bind:value={name}
                class="w-full bg-black/20 border border-white/10 text-white rounded-xl px-4 py-3.5 focus:outline-none focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500/50 transition-all text-base placeholder-gray-700"
                placeholder="My Awesome Server"
                autoFocus
                on:keydown={(e) => e.key === "Enter" && handleNext()}
            />
        </div>

        <div class="grid grid-cols-2 gap-4">
            <button
                class="relative p-5 rounded-2xl border transition-all text-left group
                {!importMode
                    ? 'bg-indigo-500/10 border-indigo-500/50 ring-1 ring-indigo-500/20'
                    : 'bg-white/5 border-white/5 hover:bg-white/10 hover:border-white/10'}"
                on:click={() => (importMode = false)}
            >
                <div
                    class="w-10 h-10 rounded-xl {!importMode
                        ? 'bg-indigo-500 text-white shadow-lg shadow-indigo-500/20'
                        : 'bg-white/10 text-gray-400 group-hover:bg-white/20 group-hover:text-white'} flex items-center justify-center mb-4 transition-colors"
                >
                    <Server size={20} />
                </div>
                <h3 class="text-white font-semibold mb-1">New Server</h3>
                <p
                    class="text-xs text-gray-400 group-hover:text-gray-300 transition-colors leading-relaxed"
                >
                    Install fresh from template
                </p>

                {#if !importMode}
                    <div class="absolute top-4 right-4 text-indigo-400">
                        <Check size={16} />
                    </div>
                {/if}
            </button>

            <button
                class="relative p-5 rounded-2xl border transition-all text-left group
                {importMode
                    ? 'bg-emerald-500/10 border-emerald-500/50 ring-1 ring-emerald-500/20'
                    : 'bg-white/5 border-white/5 hover:bg-white/10 hover:border-white/10'}"
                on:click={() => (importMode = true)}
            >
                <div
                    class="w-10 h-10 rounded-xl {importMode
                        ? 'bg-emerald-500 text-white shadow-lg shadow-emerald-500/20'
                        : 'bg-white/10 text-gray-400 group-hover:bg-white/20 group-hover:text-white'} flex items-center justify-center mb-4 transition-colors"
                >
                    <FolderInput size={20} />
                </div>
                <h3 class="text-white font-semibold mb-1">Import Existing</h3>
                <p
                    class="text-xs text-gray-400 group-hover:text-gray-300 transition-colors leading-relaxed"
                >
                    From local folder
                </p>

                {#if importMode}
                    <div class="absolute top-4 right-4 text-emerald-400">
                        <Check size={16} />
                    </div>
                {/if}
            </button>
        </div>
    </div>

    <div class="mt-10 flex justify-end">
        <button
            on:click={handleNext}
            class="bg-white hover:bg-gray-200 text-gray-900 px-6 py-3 rounded-xl font-bold transition-all flex items-center gap-2 transform active:scale-95"
        >
            Next Step <ArrowRight size={18} />
        </button>
    </div>
</div>
