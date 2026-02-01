<script>
    import { fade } from "svelte/transition";
    import { Cpu, Hammer } from "lucide-svelte";
    import Select from "$lib/components/Select.svelte";
    import { createEventDispatcher } from "svelte";

     
    export let type;
     
    export let version;
     
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
    <h2 class="text-3xl font-bold text-white mb-2">Select Version</h2>
    <p class="text-gray-400 mb-8">
        Which version of Minecraft do you want to install?
    </p>

    {#if type === "custom"}
        <div
            class="bg-gray-800/50 border border-gray-700 rounded-xl p-6 text-center"
        >
            <div
                class="inline-flex p-3 rounded-full bg-blue-900/30 text-blue-400 mb-4"
            >
                <Cpu size={32} />
            </div>
            <h3 class="text-lg font-medium text-white mb-2">
                Custom Server JAR
            </h3>
            <p class="text-sm text-gray-400">
                You will need to manually upload your server JAR file after
                creation.
            </p>
        </div>
    {:else}
        <div class="space-y-4">
            <label class="block text-sm font-medium text-gray-400"
                >Minecraft Version</label
            >
            <Select
                options={versionOptions}
                bind:value={version}
                placeholder="Loading versions..."
                className="w-full text-lg py-3"
            />
        </div>
    {/if}

    <div class="mt-8 flex justify-between items-center">
        <button
            on:click={handleBack}
            class="text-gray-400 hover:text-white px-4 py-2 font-medium transition-colors"
            >Back</button
        >
        <button
            on:click={finish}
            class="bg-blue-600 text-white px-6 py-3 rounded-xl font-bold hover:bg-blue-500 transition-colors flex items-center gap-2 shadow-lg shadow-blue-500/20"
        >
            <Hammer size={18} /> Create Server
        </button>
    </div>
</div>
