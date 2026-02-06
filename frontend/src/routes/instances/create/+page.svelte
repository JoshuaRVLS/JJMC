<script>
    import { goto } from "$app/navigation";
    import { createId } from "@paralleldrive/cuid2";
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import { ArrowLeft, Check, Box, Scroll, Layers, Cpu } from "lucide-svelte";
    import { fade } from "svelte/transition";
    import DirectoryPicker from "$lib/components/DirectoryPicker.svelte";
    import StepBasicDetails from "$lib/creation/StepBasicDetails.svelte";
    import StepLoader from "$lib/creation/StepLoader.svelte";
    import StepVersion from "$lib/creation/StepVersion.svelte";

    let step = 1;
    let name = "";
    let importMode = false;
    let sourcePath = "";
    let showDirPicker = false;

    let type = "";
    let version = "";

    /** @type {{value: string, label: string}[]} */
    let versionOptions = [];
    let loadingLoaders = false;
    let loadingVersions = false;
    let creating = false;
    let status = "";

    let typeOptions = [
        { value: "fabric", label: "Fabric", image: "/fabric.png" },
        { value: "quilt", label: "Quilt", image: "/quilt.png" },
        { value: "forge", label: "Forge", image: "/forge.png" },
        { value: "neoforge", label: "NeoForge", image: "/neoforge.png" },
        { value: "vanilla", label: "Vanilla", icon: Box },
        { value: "paper", label: "Paper", icon: Scroll },
        { value: "spigot", label: "Spigot", icon: Layers },
        { value: "bukkit", label: "CraftBukkit", icon: Box },
        { value: "custom", label: "Custom", icon: Cpu },
    ];

    const LOADER_META = {
        fabric: { image: "/fabric.png", color: "text-amber-200" },
        quilt: { image: "/quilt.png", color: "text-emerald-200" },
        forge: { image: "/forge.png", color: "text-orange-200" },
        neoforge: { image: "/neoforge.png", color: "text-orange-200" },
        paper: { icon: Scroll, color: "text-blue-200" },
        spigot: { icon: Layers, color: "text-orange-400" },
        bukkit: { icon: Box, color: "text-yellow-400" },
        vanilla: { icon: Box, color: "text-green-400" },
        custom: { icon: Cpu, color: "text-gray-400" },
    };

    /** @type {Record<string, any>} */
    const LOADER_META_TYPED = LOADER_META;

    $: typeOptions = typeOptions.map((opt) => ({
        ...opt,
        ...(LOADER_META_TYPED[opt.value] || LOADER_META.custom),
    }));

    async function loadLoaders() {
        loadingLoaders = true;
        try {
        } catch (e) {
            console.error(e);
            addToast("Failed to load server types", "error");
        } finally {
            loadingLoaders = false;
        }
    }

    async function loadVersions() {
        if (!type || type === "custom") return;

        loadingVersions = true;
        version = "";
        versionOptions = [];

        try {
            const res = await fetch(`/api/versions/game?loader=${type}`);
            if (res.ok) {
                const data = await res.json();
                versionOptions = data
                    .filter(
                        (/** @type {any} */ v) => v.version_type === "release",
                    )
                    .map((/** @type {any} */ v) => ({
                        value: v.version,
                        label: v.version,
                    }));
                if (versionOptions.length > 0) {
                    version = versionOptions[0].value;
                }
            } else {
            }
        } catch (e) {
            console.error(e);
            addToast("Error loading versions", "error");
        } finally {
            loadingVersions = false;
        }
    }

    $: if (type && !importMode) {
        loadVersions();
    }

    function handleNext() {
        if (step === 1) {
            if (!name) return addToast("Please enter a server name", "error");
            step = 2;
        } else if (step === 2) {
            if (importMode) {
                if (!sourcePath) return addToast("Select a folder", "error");
                finish();
                return;
            }
            if (!type) return addToast("Select a server software", "error");
            step = 3;
        }
    }

    function handleBack() {
        step = Math.max(1, step - 1);
    }

    async function finish() {
        if (importMode) {
            createImport();
        } else {
            createFresh();
        }
    }

    async function createImport() {
        if (!sourcePath) return addToast("Path required", "error");

        creating = true;
        status = "Importing server...";
        try {
            const id = createId();
            const res = await fetch("/api/instances/import", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    id,
                    name,
                    sourcePath,
                }),
            });

            if (res.ok) {
                addToast("Imported successfully", "success");
                await goto(`/instances/${id}`);
            } else {
                throw await res.text();
            }
        } catch (e) {
            addToast(`Import failed: ${e}`, "error");
        } finally {
            creating = false;
        }
    }

    async function createFresh() {
        if (!type || (!version && type !== "custom"))
            return addToast("Missing configuration", "error");

        creating = true;
        status = "Initializing instance...";

        try {
            const id = createId();
            const payload = {
                id,
                name,
                type,
                version: type === "custom" ? "" : version,
            };

            const res = await fetch("/api/instances", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });

            if (!res.ok) throw await res.text();

            if (type !== "custom") {
                status = `Installing ${type} ${version} server...`;

                const installRes = await fetch(`/api/instances/${id}/install`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ version, type }),
                });

                if (!installRes.ok) {
                    addToast(
                        "Instance created, but installation failed. Check console.",
                        "warning",
                    );
                }
            }

            status = "Finalizing settings...";

            await new Promise((r) => setTimeout(r, 500));

            await goto(`/instances/${id}`);
        } catch (e) {
            addToast(`Creation failed: ${e}`, "error");
        } finally {
            creating = false;
        }
    }

    onMount(() => {
        loadLoaders();
        loadVersions();
    });
</script>

<DirectoryPicker
    bind:open={showDirPicker}
    on:select={(e) => (sourcePath = e.detail)}
    on:close={() => (showDirPicker = false)}
/>

<div
    class="min-h-screen bg-[#050510] text-gray-100 flex flex-col items-center justify-center p-4 relative font-sans"
>
    <div
        class="w-full max-w-4xl bg-gray-900/40 backdrop-blur-xl border border-white/5 rounded-3xl shadow-2xl relative z-10 overflow-hidden flex flex-col md:flex-row min-h-[550px]"
    >
        <div
            class="w-full md:w-1/3 bg-[#0a0a14]/60 border-b md:border-b-0 md:border-r border-white/5 p-8 flex flex-col justify-between"
        >
            <div>
                <button
                    on:click={() => goto("/instances")}
                    class="flex items-center gap-2 text-gray-500 hover:text-white mb-8 transition-colors text-xs font-semibold uppercase tracking-wider"
                >
                    <ArrowLeft size={14} /> Cancel
                </button>

                <h1 class="text-xl font-bold text-white mb-2 tracking-tight">
                    Create Server
                </h1>
                <p class="text-gray-400 text-sm mb-10 leading-relaxed">
                    Deploy a new Minecraft instance ready for your community.
                </p>

                <div class="space-y-8 relative">
                    <!-- Vertical Line -->
                    <div
                        class="absolute left-[15px] top-4 bottom-4 w-px bg-white/5 -z-10"
                    ></div>

                    <!-- Step 1 -->
                    <div class="flex items-start gap-4 group">
                        <div class="relative">
                            <div
                                class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold ring-4 ring-[#0a0a14]/60 bg-gray-800 text-gray-500 transition-all duration-300
                                {step >= 1 ? '!bg-indigo-500 !text-white' : ''}"
                            >
                                {#if step > 1}
                                    <Check size={14} strokeWidth={3} />
                                {:else}
                                    1
                                {/if}
                            </div>
                        </div>
                        <div
                            class="pt-1.5 transition-opacity duration-300 {step >=
                            1
                                ? 'opacity-100'
                                : 'opacity-40'}"
                        >
                            <h3
                                class="text-sm font-semibold text-white leading-none"
                            >
                                Identity
                            </h3>
                            <p
                                class="text-[10px] text-gray-500 mt-1 uppercase tracking-wider font-bold"
                            >
                                Name & Type
                            </p>
                        </div>
                    </div>

                    <!-- Step 2 -->
                    <div class="flex items-start gap-4">
                        <div class="relative">
                            <div
                                class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold ring-4 ring-[#0a0a14]/60 bg-gray-800 text-gray-500 transition-all duration-300
                                {step >= 2
                                    ? importMode
                                        ? '!bg-emerald-500 !text-white'
                                        : '!bg-indigo-500 !text-white'
                                    : ''}"
                            >
                                {#if step > 2}
                                    <Check size={14} strokeWidth={3} />
                                {:else}
                                    2
                                {/if}
                            </div>
                        </div>
                        <div
                            class="pt-1.5 transition-opacity duration-300 {step >=
                            2
                                ? 'opacity-100'
                                : 'opacity-40'}"
                        >
                            <h3
                                class="text-sm font-semibold text-white leading-none"
                            >
                                {importMode ? "Source" : "Software"}
                            </h3>
                            <p
                                class="text-[10px] text-gray-500 mt-1 uppercase tracking-wider font-bold"
                            >
                                {importMode ? "Local Folder" : "Server Type"}
                            </p>
                        </div>
                    </div>

                    <!-- Step 3 -->
                    {#if !importMode}
                        <div class="flex items-start gap-4">
                            <div class="relative">
                                <div
                                    class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold ring-4 ring-[#0a0a14]/60 bg-gray-800 text-gray-500 transition-all duration-300
                                    {step >= 3
                                        ? '!bg-indigo-500 !text-white'
                                        : ''}"
                                >
                                    3
                                </div>
                            </div>
                            <div
                                class="pt-1.5 transition-opacity duration-300 {step >=
                                3
                                    ? 'opacity-100'
                                    : 'opacity-40'}"
                            >
                                <h3
                                    class="text-sm font-semibold text-white leading-none"
                                >
                                    Version
                                </h3>
                                <p
                                    class="text-[10px] text-gray-500 mt-1 uppercase tracking-wider font-bold"
                                >
                                    Target Release
                                </p>
                            </div>
                        </div>
                    {/if}
                </div>
            </div>

            <div class="text-[10px] text-gray-700 font-mono mt-8">
                JJMC SERVER MANAGER
            </div>
        </div>

        <div class="flex-1 p-8 md:p-12 relative flex flex-col">
            {#if creating}
                <div
                    class="absolute inset-0 z-20 bg-gray-900/90 backdrop-blur-md flex flex-col items-center justify-center p-8 text-center"
                    transition:fade
                >
                    <div class="relative w-16 h-16 mb-6">
                        <div
                            class="absolute inset-0 rounded-full border-4 border-white/10"
                        ></div>
                        <div
                            class="absolute inset-0 rounded-full border-4 border-t-indigo-500 border-r-transparent border-b-transparent border-l-transparent animate-spin"
                        ></div>
                    </div>
                    <h2 class="text-lg font-bold text-white mb-2">{status}</h2>
                    <p class="text-gray-400 text-sm">
                        Please do not close this window.
                    </p>
                </div>
            {/if}

            <div class="flex-1 flex flex-col">
                {#if step === 1}
                    <StepBasicDetails
                        bind:name
                        bind:importMode
                        on:next={handleNext}
                    />
                {:else if step === 2}
                    <StepLoader
                        bind:importMode
                        bind:sourcePath
                        bind:showDirPicker
                        bind:type
                        {typeOptions}
                        on:back={handleBack}
                        on:next={handleNext}
                        on:finishImport={createImport}
                    />
                {:else if step === 3}
                    <StepVersion
                        {type}
                        {versionOptions}
                        bind:version
                        on:back={handleBack}
                        on:finish={finish}
                    />
                {/if}
            </div>
        </div>
    </div>
</div>

<style>
    .custom-scrollbar::-webkit-scrollbar {
        width: 6px;
    }
    .custom-scrollbar::-webkit-scrollbar-track {
        background: transparent;
    }
    .custom-scrollbar::-webkit-scrollbar-thumb {
        background: rgba(255, 255, 255, 0.1);
        border-radius: 3px;
    }
    .custom-scrollbar::-webkit-scrollbar-thumb:hover {
        background: rgba(255, 255, 255, 0.2);
    }
</style>
