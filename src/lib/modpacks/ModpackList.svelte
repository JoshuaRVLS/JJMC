<script>
    import { Download } from "lucide-svelte";
    import InstallModpackModal from "./InstallModpackModal.svelte";

    /**
     * @typedef {Object} Modpack
     * @property {string} project_id
     * @property {string} title
     * @property {string} author
     * @property {string} description
     * @property {string} icon_url
     * @property {number} downloads
     */

    /** @type {Modpack[]} */
    export let results = [];

    /** @type {Modpack | null} */
    let selectedModpack = null;
    let showInstallModal = false;

    /**
     * @param {Modpack} modpack
     */
    function handleInstall(modpack) {
        selectedModpack = modpack;
        showInstallModal = true;
    }
</script>

{#if results.length === 0}
    <div class="h-full flex flex-col items-center justify-center text-gray-500">
        <span>No modpacks found. Try adjusting your filters.</span>
    </div>
{:else}
    <div
        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 overflow-y-auto h-full pb-4"
    >
        {#each results as modpack}
            <div
                class="bg-gray-800/40 border border-white/5 rounded-xl p-4 hover:bg-gray-800/60 transition-colors flex flex-col"
            >
                <div class="flex items-start gap-4 mb-3">
                    {#if modpack.icon_url}
                        <img
                            src={modpack.icon_url}
                            alt={modpack.title}
                            class="w-12 h-12 rounded-lg object-cover bg-gray-900"
                        />
                    {:else}
                        <div
                            class="w-12 h-12 rounded-lg bg-gray-900 flex items-center justify-center text-lg font-bold text-indigo-500"
                        >
                            {modpack.title.charAt(0)}
                        </div>
                    {/if}
                    <div class="flex-1 min-w-0">
                        <h3
                            class="font-bold text-white truncate text-sm"
                            title={modpack.title}
                        >
                            {modpack.title}
                        </h3>
                        <p class="text-xs text-gray-400 truncate">
                            by {modpack.author}
                        </p>
                    </div>
                </div>

                <p class="text-xs text-gray-400 line-clamp-2 mb-4 flex-1">
                    {modpack.description}
                </p>

                <div class="flex items-center justify-between mt-auto">
                    <div
                        class="text-[10px] text-gray-500 bg-gray-900/50 px-2 py-1 rounded"
                    >
                        {modpack.downloads.toLocaleString()} downloads
                    </div>
                    <button
                        on:click={() => handleInstall(modpack)}
                        class="bg-indigo-600 hover:bg-indigo-500 text-white px-3 py-1.5 rounded-lg text-xs font-bold transition-colors flex items-center gap-1.5"
                    >
                        <Download class="w-3.5 h-3.5" />
                        Install
                    </button>
                </div>
            </div>
        {/each}
    </div>
{/if}

<InstallModpackModal bind:open={showInstallModal} modpack={selectedModpack} />
