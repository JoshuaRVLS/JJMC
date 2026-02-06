<script>
    import { fade } from "svelte/transition";
    import { Cpu, Hammer, ArrowRight } from "lucide-svelte";
    import Select from "$lib/components/Select.svelte";
    import { createEventDispatcher } from "svelte";

    /** @type {string} */
    export let type;

    /** @type {string} */
    export let version;

    /** @type {any[]} */
    export let versionOptions;

    const dispatch = createEventDispatcher();

    function handleBack() {
        dispatch("back");
    }

    function finish() {
        dispatch("finish");
    }
</script>

<div
    in:fade={{ duration: 200, delay: 100 }}
    class="flex-1 flex flex-col justify-center max-w-lg mx-auto w-full"
>
    <div class="mb-8">
        <h2 class="text-2xl font-bold text-white mb-2 tracking-tight">
            Select Version
        </h2>
        <p class="text-gray-400 text-sm">Target Minecraft release version.</p>
    </div>

    {#if type === "custom"}
        <div
            class="bg-white/5 border border-white/5 rounded-2xl p-8 text-center"
        >
            <div
                class="inline-flex p-4 rounded-full bg-white/5 text-gray-400 mb-4"
            >
                <Cpu size={32} />
            </div>
            <h3 class="text-lg font-bold text-white mb-2">Custom Server JAR</h3>
            <p class="text-sm text-gray-400 leading-relaxed max-w-xs mx-auto">
                You will need to manually upload your server JAR file via the
                file manager after creation.
            </p>
        </div>
    {:else}
        <div class="space-y-4">
            <label
                for="version-select"
                class="block text-xs font-bold text-gray-500 uppercase tracking-widest pl-1"
                >Minecraft Version</label
            >
            <Select
                id="version-select"
                options={versionOptions}
                bind:value={version}
                placeholder="Loading versions..."
                className="w-full text-lg py-3 bg-black/20 border-white/10"
            />
        </div>
    {/if}

    <div class="mt-10 flex justify-between items-center">
        <button
            on:click={handleBack}
            class="text-gray-500 hover:text-white px-2 py-2 font-medium transition-colors text-sm"
            >Back</button
        >
        <button
            on:click={finish}
            class="bg-white text-gray-900 px-6 py-3 rounded-xl font-bold hover:bg-gray-200 transition-colors flex items-center gap-2 shadow-lg shadow-white/10 transform active:scale-95"
        >
            <Hammer size={18} /> Create Server
        </button>
    </div>
</div>
