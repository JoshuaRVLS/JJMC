<script>
    import { createEventDispatcher } from "svelte";
    import ModCard from "./ModCard.svelte";
    import ModVersions from "./ModVersions.svelte";

    /** @type {Array<any>} */
    export let results;
    /** @type {boolean} */
    export let loading;
    /** @type {boolean} */
    export let loadingMore;
    /** @type {string} */
    export let query;
    /** @type {boolean} */
    export let hasMore;
    /** @type {Set<string>} */
    export let installedIds;
    /** @type {string | null} */
    export let installingId;
    /** @type {string} */
    export let activeTab;
    /** @type {string | null} */
    export let viewingVersionsId;
    /** @type {Array<any>} */
    export let versionsList;
    /** @type {boolean} */
    export let loadingVersions;
    /** @type {HTMLElement} */
    export let sentinel;

    const dispatch = createEventDispatcher();

    function viewVersions(item) {
        dispatch("viewVersions", item);
    }

    function install(item) {
        dispatch("install", item);
    }

    function uninstall(item) {
        dispatch("uninstall", item);
    }

    function closeVersions() {
        dispatch("closeVersions");
    }

    function installVersion(projectId, versionId) {
        dispatch("installVersion", { projectId, versionId });
    }
</script>

<div class="flex-1 overflow-y-auto pr-2 custom-scrollbar">
    {#if loading && results.length === 0}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {#each Array(6) as _}
                <div
                    class="bg-white/5 h-48 rounded-xl animate-pulse border border-white/5"
                ></div>
            {/each}
        </div>
    {:else if results.length > 0}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-4 pb-4">
            {#each results as item}
                <ModCard
                    {item}
                    {activeTab}
                    {installedIds}
                    {installingId}
                    on:viewVersions={(e) => viewVersions(e.detail)}
                    on:install={(e) => install(e.detail)}
                    on:uninstall={(e) => uninstall(e.detail)}
                >
                    <div slot="versions">
                        {#if viewingVersionsId === item.project_id}
                            <ModVersions
                                {item}
                                {versionsList}
                                {loadingVersions}
                                {activeTab}
                                on:close={closeVersions}
                                on:install={(e) =>
                                    installVersion(
                                        e.detail.projectId,
                                        e.detail.versionId,
                                    )}
                            />
                        {/if}
                    </div>
                </ModCard>
            {/each}
        </div>

        <!-- Sentinel for Infinite Scroll -->
        <div bind:this={sentinel} class="h-10 flex items-center justify-center">
            {#if loadingMore}
                <div
                    class="flex items-center gap-2 text-gray-500 text-xs font-medium bg-black/20 px-4 py-2 rounded-full"
                >
                    <svg class="animate-spin h-3 w-3" viewBox="0 0 24 24">
                        <circle
                            class="opacity-25"
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            stroke-width="4"
                            fill="none"
                        ></circle>
                        <path
                            class="opacity-75"
                            fill="currentColor"
                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                        ></path>
                    </svg>
                    Loading more...
                </div>
            {:else if !hasMore && results.length > 0}
                <span
                    class="text-xs text-gray-600 bg-black/10 px-4 py-1.5 rounded-full border border-white/5"
                    >You've reached the end of the list.</span
                >
            {/if}
        </div>
    {:else if query && !loading}
        <div
            class="text-center text-gray-500 py-20 flex flex-col items-center gap-4"
        >
            <div class="p-6 rounded-full bg-white/5 text-gray-600">
                <svg
                    class="w-12 h-12"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.5"
                        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                    />
                </svg>
            </div>
            <div class="space-y-1">
                <p class="text-white font-medium">No results found</p>
                <p class="text-xs">
                    Try adjusting your search or sorting criteria.
                </p>
            </div>
        </div>
    {/if}
</div>
