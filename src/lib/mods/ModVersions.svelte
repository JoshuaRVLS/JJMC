<script>
    /** @type {any} */
    export let item;

    /** @type {any[]} */
    export let versionsList;

    /** @type {boolean} */
    export let loadingVersions;

    /** @type {string} */
    export let activeTab;

    import { createEventDispatcher } from "svelte";
    const dispatch = createEventDispatcher();

    function close() {
        dispatch("close");
    }

    /**
     * @param {string} projectId
     * @param {string} versionId
     */
    function install(projectId, versionId) {
        dispatch("install", { projectId, versionId });
    }
</script>

<div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm animate-in fade-in duration-200"
    role="button"
    tabindex="-1"
    on:click|stopPropagation={close}
    on:keydown={(e) => e.key === "Escape" && close()}
    aria-label="Close modal"
>
    <div
        class="bg-gray-900 border border-white/10 rounded-2xl shadow-2xl w-full max-w-lg max-h-[80vh] flex flex-col m-4 animate-in zoom-in-95 duration-200 cursor-auto"
        role="dialog"
        tabindex="-1"
        aria-modal="true"
        on:click|stopPropagation
        on:keydown|stopPropagation
    >
        <div
            class="p-4 border-b border-white/10 flex justify-between items-center bg-black/20 rounded-t-2xl"
        >
            <div>
                <h3 class="text-white font-bold text-lg">
                    {item.title}
                </h3>
                <p class="text-gray-400 text-xs">Select a version to install</p>
            </div>
            <button
                class="text-gray-400 hover:text-white bg-white/5 hover:bg-white/10 p-2 rounded-lg transition-colors"
                aria-label="Close versions"
                on:click={close}
            >
                <svg
                    class="w-5 h-5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M6 18L18 6M6 6l12 12"
                    />
                </svg>
            </button>
        </div>

        <div class="flex-1 overflow-y-auto custom-scrollbar p-2 space-y-2">
            {#if loadingVersions}
                <div
                    class="flex flex-col items-center justify-center py-12 text-gray-400"
                >
                    <svg
                        class="animate-spin w-8 h-8 text-indigo-500 mb-3"
                        fill="none"
                        viewBox="0 0 24 24"
                    >
                        <circle
                            class="opacity-25"
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            stroke-width="4"
                        ></circle>
                        <path
                            class="opacity-75"
                            fill="currentColor"
                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                        ></path>
                    </svg>
                    <span class="text-sm font-medium">Loading versions...</span>
                </div>
            {:else if versionsList.length === 0}
                <div
                    class="flex flex-col items-center justify-center py-12 text-gray-500"
                >
                    <svg
                        class="w-12 h-12 mb-3 bg-white/5 p-2 rounded-xl"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="1.5"
                            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                        />
                    </svg>
                    <p class="text-sm font-medium">
                        No compatible versions found.
                    </p>
                    <p class="text-xs mt-1">
                        Try changing your Minecraft version filter.
                    </p>
                </div>
            {:else}
                {#each versionsList as ver}
                    <div
                        class="group flex items-center justify-between p-3 rounded-xl bg-black/20 hover:bg-indigo-500/10 border border-white/5 hover:border-indigo-500/30 transition-all"
                    >
                        <div class="min-w-0 flex-1 mr-4">
                            <div class="flex items-center gap-2 mb-0.5">
                                <span
                                    class="text-sm font-bold text-gray-200 truncate"
                                    >{ver.version_number}</span
                                >
                                <span
                                    class="text-[10px] px-1.5 py-0.5 rounded bg-white/10 text-gray-400 capitalize"
                                    >{ver.version_type || "release"}</span
                                >
                            </div>
                            <div
                                class="flex items-center gap-2 text-xs text-gray-500"
                            >
                                <span
                                    >{new Date(
                                        ver.date_published || Date.now(),
                                    ).toLocaleDateString(undefined, {
                                        year: "numeric",
                                        month: "short",
                                        day: "numeric",
                                    })}</span
                                >
                                {#if ver.loaders}
                                    <span>•</span>
                                    <span class="truncate max-w-[150px]"
                                        >{ver.loaders.join(", ")}</span
                                    >
                                {/if}
                            </div>
                        </div>

                        <button
                            on:click|stopPropagation={() =>
                                install(item.project_id, ver.id)}
                            class="shrink-0 bg-white/5 hover:bg-indigo-600 hover:text-white text-gray-300 px-4 py-2 rounded-lg text-xs font-bold transition-all flex items-center gap-2 group-hover:shadow-lg group-hover:shadow-indigo-500/20"
                        >
                            <svg
                                class="w-4 h-4"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                                />
                            </svg>
                            Install
                        </button>
                    </div>
                {/each}
            {/if}
        </div>

        <div
            class="p-3 bg-black/40 text-center text-[10px] text-gray-500 border-t border-white/5 rounded-b-2xl"
        >
            Showing compatible versions for {activeTab === "plugin"
                ? "Plugins"
                : "Mods"}
        </div>
    </div>
</div>
