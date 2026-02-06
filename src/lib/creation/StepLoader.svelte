<script>
    import { fade } from "svelte/transition";
    import { HardDrive, Download, Box, ArrowRight } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";

    /** @type {boolean} */
    export let importMode;

    /** @type {string} */
    export let sourcePath;

    /** @type {string} */
    export let type;

    /** @type {any[]} */
    export let typeOptions;
    export let showDirPicker = false;

    const dispatch = createEventDispatcher();

    function handleBack() {
        dispatch("back");
    }

    function handleNext() {
        dispatch("next");
    }

    function finishImport() {
        dispatch("finishImport");
    }
</script>

<div
    in:fade={{ duration: 200, delay: 100 }}
    class="flex-1 flex flex-col justify-center max-w-lg mx-auto w-full"
>
    {#if importMode}
        <div class="mb-8">
            <h2 class="text-2xl font-bold text-white mb-2 tracking-tight">
                Locate Server
            </h2>
            <p class="text-gray-400 text-sm">
                Select the folder containing your server files.
            </p>
        </div>

        <div class="space-y-4">
            <div>
                <label
                    for="server-path"
                    class="block text-xs font-bold text-gray-500 uppercase tracking-widest mb-2 pl-1"
                    >Folder Path</label
                >
                <div class="flex gap-2">
                    <div class="relative flex-1">
                        <div
                            class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-gray-500"
                        >
                            <HardDrive size={18} />
                        </div>
                        <input
                            type="text"
                            id="server-path"
                            bind:value={sourcePath}
                            class="w-full bg-black/20 border border-white/10 text-white rounded-xl pl-11 pr-4 py-3 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500/50 transition-all font-mono text-sm placeholder-gray-700"
                            placeholder="/absolute/path/to/server"
                        />
                    </div>
                    <button
                        on:click={() => (showDirPicker = true)}
                        class="bg-white/5 hover:bg-white/10 border border-white/5 text-gray-300 px-4 rounded-xl transition-colors text-sm font-medium"
                    >
                        Browse
                    </button>
                </div>
                <p class="text-xs text-gray-600 mt-2 pl-1">
                    Must contain <code>server.jar</code> and
                    <code>server.properties</code>.
                </p>
            </div>
        </div>

        <div class="mt-10 flex justify-between items-center">
            <button
                on:click={handleBack}
                class="text-gray-500 hover:text-white px-2 py-2 font-medium transition-colors text-sm"
                >Back</button
            >
            <button
                on:click={finishImport}
                class="bg-emerald-500 text-white px-6 py-3 rounded-xl font-bold hover:bg-emerald-400 transition-colors flex items-center gap-2 shadow-lg shadow-emerald-500/20 transform active:scale-95"
            >
                <Download size={18} /> Import Server
            </button>
        </div>
    {:else}
        <div class="mb-6">
            <h2 class="text-2xl font-bold text-white mb-2 tracking-tight">
                Select Software
            </h2>
            <p class="text-gray-400 text-sm">
                Choose the server software you want to run.
            </p>
        </div>

        <div
            class="grid grid-cols-2 gap-3 max-h-[400px] overflow-y-auto pr-2 custom-scrollbar"
        >
            {#each typeOptions as option}
                <button
                    class="relative p-4 rounded-xl border text-left transition-all group
                    {type === option.value
                        ? 'bg-indigo-500/10 border-indigo-500/50 ring-1 ring-indigo-500/20'
                        : 'bg-white/5 border-white/5 hover:bg-white/10 hover:border-white/10'}"
                    on:click={() => {
                        type = option.value;
                        handleNext();
                    }}
                >
                    <div class="flex items-center gap-3">
                        {#if option.image}
                            <img
                                src={option.image}
                                alt={option.label}
                                class="w-8 h-8 object-contain opacity-80 group-hover:opacity-100 transition-opacity"
                            />
                        {:else}
                            <div
                                class="p-2 rounded-lg bg-white/5 text-gray-400 group-hover:text-white transition-colors"
                            >
                                <svelte:component
                                    this={option.icon || Box}
                                    size={18}
                                />
                            </div>
                        {/if}
                        <h3
                            class="text-gray-300 group-hover:text-white font-medium text-sm"
                        >
                            {option.label}
                        </h3>
                    </div>
                </button>
            {/each}
        </div>
        <div class="mt-8 flex justify-between items-center">
            <button
                on:click={handleBack}
                class="text-gray-500 hover:text-white px-2 py-2 font-medium transition-colors text-sm"
                >Back</button
            >
        </div>
    {/if}
</div>
