<script>
     
    export let item;
     
    export let activeTab;
     
    export let installedIds;
     
    export let installingId;

    import { createEventDispatcher } from "svelte";
    const dispatch = createEventDispatcher();

    function getSideColor(side) {
        if (side === "required")
            return "bg-rose-500/20 text-rose-400 border-rose-500/30";
        if (side === "optional")
            return "bg-amber-500/20 text-amber-400 border-amber-500/30";
        return "bg-gray-500/20 text-gray-400 border-gray-500/30";
    }

    function viewVersions() {
        dispatch("viewVersions", item);
    }

    function install() {
        dispatch("install", item);
    }

    function uninstall() {
        dispatch("uninstall", item);
    }
</script>

<div
    class="bg-gray-900/40 border border-white/5 hover:border-indigo-500/30 rounded-xl p-4 transition-all group relative overflow-hidden flex flex-col"
>
    <div class="flex gap-4 mb-3">
        
        <div class="shrink-0">
            {#if item.icon_url}
                <img
                    src={item.icon_url}
                    alt={item.title}
                    class="w-16 h-16 rounded-lg object-cover bg-black/30"
                />
            {:else}
                <div
                    class="w-16 h-16 rounded-lg bg-indigo-500/20 flex items-center justify-center text-indigo-400"
                >
                    {#if activeTab === "mod" || activeTab === "plugin"}
                        <svg
                            class="w-8 h-8"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                            ><path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"
                            /></svg
                        >
                    {:else}
                        <svg
                            class="w-8 h-8"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                            ><path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"
                            /></svg
                        >
                    {/if}
                </div>
            {/if}
        </div>

        
        <div class="flex-1 min-w-0">
            <div class="flex justify-between items-start">
                <h3 class="text-white font-bold truncate pr-2 text-sm">
                    {item.title}
                </h3>
                <div class="flex gap-1">
                    {#if item.client_side !== "unsupported"}
                        <span
                            class="text-[9px] px-1.5 py-0.5 rounded border {getSideColor(
                                item.client_side,
                            )}"
                            title="Client Side: {item.client_side}">CL</span
                        >
                    {/if}
                    {#if item.server_side !== "unsupported"}
                        <span
                            class="text-[9px] px-1.5 py-0.5 rounded border {getSideColor(
                                item.server_side,
                            )}"
                            title="Server Side: {item.server_side}">SV</span
                        >
                    {/if}
                </div>
            </div>
            <div class="text-[10px] text-gray-500 mb-1 flex items-center gap-2">
                <span>by {item.author}</span>
                <span>•</span>
                <span
                    class="flex items-center gap-0.5"
                    title="{item.downloads} downloads"
                >
                    <svg
                        class="w-3 h-3"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        ><path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                        ></path></svg
                    >
                    {Intl.NumberFormat("en-US", {
                        notation: "compact",
                        maximumFractionDigits: 1,
                    }).format(item.downloads)}
                </span>
            </div>
            <p class="text-gray-400 text-xs line-clamp-2 leading-relaxed h-8">
                {item.description}
            </p>
        </div>
    </div>

    
    <div class="flex flex-wrap gap-1 mb-4">
        {#each (item.categories || []).slice(0, 3) as cat}
            <span
                class="bg-white/5 text-[10px] text-gray-400 px-2 py-0.5 rounded uppercase tracking-wider border border-white/5"
                >{cat}</span
            >
        {/each}
        {#if (item.categories || []).length > 3}
            <span
                class="bg-white/5 text-[10px] text-gray-400 px-2 py-0.5 rounded border border-white/5"
                >+{item.categories.length - 3}</span
            >
        {/if}
    </div>

    <slot name="versions"></slot>

    
    <div
        class="mt-auto flex justify-between items-center bg-black/20 -mx-4 -mb-4 px-4 py-2.5 border-t border-white/5"
    >
        <span class="text-[10px] text-gray-500"
            >Updated {new Date(item.date_modified).toLocaleDateString()}</span
        >
        <div class="flex gap-2">
            {#if activeTab !== "modpack"}
                <button
                    on:click|stopPropagation={viewVersions}
                    class="bg-white/5 hover:bg-white/10 text-gray-300 px-3 py-1.5 rounded-lg text-xs font-bold transition-all border border-white/5"
                >
                    Versions
                </button>
            {/if}

            <button
                on:click|stopPropagation={() => {
                    if (
                        (activeTab === "mod" || activeTab === "plugin") &&
                        installedIds.has(item.project_id)
                    ) {
                        uninstall();
                    } else {
                        install();
                    }
                }}
                disabled={installingId === item.project_id ||
                    installingId !== null}
                class="flex items-center gap-2 {(activeTab === 'mod' ||
                    activeTab === 'plugin') &&
                installedIds.has(item.project_id)
                    ? 'bg-red-500/10 text-red-400 border border-red-500/30 hover:bg-red-500/20'
                    : 'bg-indigo-500 hover:bg-indigo-600 text-white shadow-lg shadow-indigo-500/20'} px-3 py-1.5 rounded-lg text-xs font-bold transition-all disabled:opacity-50"
            >
                {#if installingId === item.project_id}
                    <svg
                        class="animate-spin w-3 h-3"
                        fill="none"
                        viewBox="0 0 24 24"
                        ><circle
                            class="opacity-25"
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            stroke-width="4"
                        ></circle><path
                            class="opacity-75"
                            fill="currentColor"
                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                        ></path></svg
                    >
                    Processing...
                {:else if (activeTab === "mod" || activeTab === "plugin") && installedIds.has(item.project_id)}
                    <svg
                        class="w-3 h-3"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        ><path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                        ></path></svg
                    >
                    Uninstall
                {:else}
                    <svg
                        class="w-3 h-3"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        ><path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                        ></path></svg
                    >
                    Install
                {/if}
            </button>
        </div>
    </div>
</div>
