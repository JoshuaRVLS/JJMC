<script>
    import { ArrowUp } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";

    export let currentPath;

    export let breadcrumbs;

    const dispatch = createEventDispatcher();

    function navigateUp() {
        dispatch("navigateUp");
    }

    /** @param {string} path */
    function navigate(path) {
        dispatch("navigate", path);
    }
</script>

<div class="flex items-center gap-2 text-sm text-gray-400 overflow-x-auto">
    <button
        on:click={navigateUp}
        disabled={currentPath === "." || currentPath === ""}
        class="p-1 hover:text-white disabled:opacity-30 disabled:hover:text-gray-400"
    >
        <ArrowUp class="w-5 h-5" />
    </button>
    <div class="h-4 w-px bg-white/10 mx-1"></div>
    {#each breadcrumbs as crumb, i}
        <button
            class="hover:text-white transition-colors whitespace-nowrap {i ===
            breadcrumbs.length - 1
                ? 'text-white font-semibold'
                : ''}"
            on:click={() => navigate(crumb.path)}
        >
            {crumb.name}
        </button>
        {#if i < breadcrumbs.length - 1}
            <span class="text-gray-600">/</span>
        {/if}
    {/each}
</div>
