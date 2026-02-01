<script>
    import { fade } from "svelte/transition";
    import { HardDrive, Download, Box, ArrowRight } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";

     
    export let importMode;
     
    export let sourcePath;
     
    export let type;
     
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
        <h2 class="text-3xl font-bold text-white mb-6">Locate Server</h2>
        <div class="space-y-4">
            <div>
                <label class="block text-sm font-medium text-gray-400 mb-2"
                    >Folder Path</label
                >
                <div class="flex gap-2">
                    <div class="relative flex-1">
                        <div
                            class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-gray-500"
                        >
                            <HardDrive size={18} />
                        </div>
                        <input
                            type="text"
                            bind:value={sourcePath}
                            class="w-full bg-gray-800 border border-gray-700 text-white rounded-xl pl-10 pr-4 py-3 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all font-mono text-sm"
                            placeholder="/absolute/path/to/server"
                        />
                    </div>
                    <button
                        on:click={() => (showDirPicker = true)}
                        class="bg-gray-800 hover:bg-gray-700 border border-gray-700 text-gray-300 px-4 rounded-xl transition-colors"
                    >
                        Browse
                    </button>
                </div>
                <p class="text-xs text-gray-500 mt-2">
                    Ensure this folder contains your <code>server.jar</code>
                    and <code>server.properties</code>. We will copy it to the
                    instances storage.
                </p>
            </div>
        </div>

        <div class="mt-8 flex justify-between items-center">
            <button
                on:click={handleBack}
                class="text-gray-400 hover:text-white px-4 py-2 font-medium transition-colors"
                >Back</button
            >
            <button
                on:click={finishImport}
                class="bg-emerald-500 text-white px-6 py-3 rounded-xl font-bold hover:bg-emerald-400 transition-colors flex items-center gap-2 shadow-lg shadow-emerald-500/20"
            >
                <Download size={18} /> Import Server
            </button>
        </div>
    {:else}
        <h2 class="text-3xl font-bold text-white mb-6">Choose Software</h2>

        <div
            class="grid grid-cols-2 gap-3 max-h-[400px] overflow-y-auto pr-2 custom-scrollbar"
        >
            {#each typeOptions as option}
                <button
                    class="relative p-4 rounded-xl border text-left transition-all
                    {type === option.value
                        ? `${option.bg} ${option.border} ring-1 ring-blue-500/30`
                        : 'bg-gray-800/40 border-gray-700 hover:bg-gray-800 hover:border-gray-600'}"
                    on:click={() => {
                        type = option.value;
                        handleNext();
                    }}
                >
                    <div class="flex items-center gap-3 mb-2">
                        {#if option.image}
                            <img
                                src={option.image}
                                alt={option.label}
                                class="w-10 h-10 object-contain"
                            />
                        {:else}
                            <div
                                class="p-2 rounded-lg bg-gray-900/50 {option.color}"
                            >
                                <svelte:component
                                    this={option.icon || Box}
                                    size={20}
                                />
                            </div>
                        {/if}
                        <h3 class="text-white font-semibold">
                            {option.label}
                        </h3>
                    </div>
                    {#if type === option.value}
                        <div
                            class="absolute top-1/2 right-4 -translate-y-1/2 text-blue-400"
                        >
                            <ArrowRight size={20} />
                        </div>
                    {/if}
                </button>
            {/each}
        </div>
        <div class="mt-8 flex justify-between items-center">
            <button
                on:click={handleBack}
                class="text-gray-400 hover:text-white px-4 py-2 font-medium transition-colors"
                >Back</button
            >
        </div>
    {/if}
</div>
