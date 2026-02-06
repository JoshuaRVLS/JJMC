<script>
    /** @type {string} */
    export let mode = "mod";

    /** @type {string} */
    export let activeTab;

    /** @type {string} */
    export let sortBy;

    /** @type {string} */
    export let query;

    /** @type {boolean} */
    export let loading;

    /** @type {string} */
    export let type = "";

    import { createEventDispatcher } from "svelte";
    const dispatch = createEventDispatcher();

    function search() {
        dispatch("search");
    }

    /** @param {KeyboardEvent} e */
    function handleKeydown(e) {
        if (e.key === "Enter") search();
    }
</script>

<div
    class="flex flex-col md:flex-row md:items-center justify-between mb-4 gap-4 px-1"
>
    <div class="flex bg-black/20 p-1 rounded-lg self-start">
        {#if mode === "plugin"}
            <button
                class="px-4 py-1.5 rounded-md text-sm font-medium transition-all bg-indigo-500 text-white shadow-lg"
                on:click={() => {}}
            >
                Plugins
            </button>
        {:else}
            <button
                class="px-4 py-1.5 rounded-md text-sm font-medium transition-all {activeTab ===
                'mod'
                    ? 'bg-indigo-500 text-white shadow-lg'
                    : 'text-gray-400 hover:text-white'}"
                on:click={() => {
                    activeTab = "mod";
                }}
            >
                Mods
            </button>
            <button
                class="px-4 py-1.5 rounded-md text-sm font-medium transition-all {activeTab ===
                'modpack'
                    ? 'bg-indigo-500 text-white shadow-lg'
                    : 'text-gray-400 hover:text-white'}"
                on:click={() => {
                    activeTab = "modpack";
                }}
            >
                Modpacks
            </button>
        {/if}
    </div>

    <div class="flex items-center gap-2">
        <span class="text-xs text-gray-500 font-medium uppercase tracking-wider"
            >Sort by:</span
        >
        <select
            bind:value={sortBy}
            class="bg-black/20 border border-white/10 rounded-lg px-2 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-indigo-500 transition-all"
        >
            <option value="relevance">Relevance</option>
            <option value="downloads">Downloads</option>
            <option value="follows">Follows</option>
            <option value="newest">Newest</option>
            <option value="updated">Recently Updated</option>
        </select>
    </div>
</div>

<div class="flex gap-2 mb-6">
    <div class="relative flex-1">
        <input
            type="text"
            bind:value={query}
            on:keydown={handleKeydown}
            placeholder={activeTab === "mod"
                ? type === "spigot"
                    ? "Search for plugins..."
                    : "Search for mods..."
                : "Search for plugins..."}
            class="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-3 pl-11 text-white focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all"
        />
        <svg
            class="w-5 h-5 text-gray-500 absolute left-3.5 top-3.5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
        >
            <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
        </svg>
    </div>
    <button
        on:click={search}
        disabled={loading}
        class="bg-indigo-600 hover:bg-indigo-500 text-white px-6 rounded-xl font-medium transition-colors disabled:opacity-50"
    >
        {loading ? "Searching..." : "Search"}
    </button>
</div>
