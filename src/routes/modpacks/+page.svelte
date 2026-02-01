<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import { Loader2, Search, Download, ExternalLink } from "lucide-svelte";
    import ModpackList from "$lib/modpacks/ModpackList.svelte";
    import Select from "$lib/components/Select.svelte";

    let query = "";
    let version = "";
    let type = "";

    let versions = [];
    let types = [
        { value: "", label: "Any Loader" },
        { value: "fabric", label: "Fabric" },
        { value: "forge", label: "Forge" },
        { value: "quilt", label: "Quilt" },
        { value: "neoforge", label: "NeoForge" },
    ];

    let loading = false;
    let results = [];
    let searchTimeout;

    async function loadVersions() {
        try {
            const res = await fetch("/api/versions/game");
            if (res.ok) {
                const data = await res.json();
                versions = [
                    { value: "", label: "Any Version" },
                    ...data
                        .filter((v) => v.version_type === "release")
                        .map((v) => ({ value: v.version, label: v.version })),
                ];
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function search() {
        loading = true;
        try {
            const params = new URLSearchParams({
                query,
                version,
                loader: type,
                sort: "relevance",
            });

            const res = await fetch(`/api/modpacks/search?${params}`);
            if (res.ok) {
                results = await res.json();
            } else {
                addToast("Failed to search modpacks", "error");
            }
        } catch (e) {
            console.error(e);
            addToast("Error searching modpacks", "error");
        } finally {
            loading = false;
        }
    }

    function handleInput() {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(search, 500);
    }

    onMount(() => {
        loadVersions();
        search();
    });
</script>

<div class="h-full flex flex-col p-8">
    <header class="flex justify-between items-center mb-8">
        <div>
            <h2 class="text-2xl font-bold text-white tracking-tight">
                Browse Modpacks
            </h2>
            <div class="text-sm text-gray-400 mt-1">
                Discover and install modpacks from Modrinth
            </div>
        </div>
    </header>

    <div
        class="bg-gray-900/60 backdrop-blur-xl border border-white/5 rounded-2xl p-6 mb-6"
    >
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div class="md:col-span-2 relative">
                <Search class="absolute left-3 top-3 w-5 h-5 text-gray-500" />
                <input
                    type="text"
                    bind:value={query}
                    on:input={handleInput}
                    placeholder="Search modpacks..."
                    class="w-full bg-gray-950 border border-gray-800 rounded-xl py-2.5 pl-10 pr-4 text-white focus:outline-none focus:border-indigo-500/50 focus:ring-1 focus:ring-indigo-500/50"
                />
            </div>
            <div>
                <Select
                    options={versions}
                    bind:value={version}
                    on:change={search}
                    placeholder="Game Version"
                    className="w-full"
                />
            </div>
            <div>
                <Select
                    options={types}
                    bind:value={type}
                    on:change={search}
                    placeholder="Mod Loader"
                    className="w-full"
                />
            </div>
        </div>
    </div>

    <div class="flex-1 min-h-0 relative">
        {#if loading}
            <div class="absolute inset-0 flex items-center justify-center">
                <Loader2 class="w-8 h-8 text-indigo-500 animate-spin" />
            </div>
        {:else}
            <ModpackList {results} />
        {/if}
    </div>
</div>
