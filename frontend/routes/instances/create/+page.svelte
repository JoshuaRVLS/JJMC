<script>
    import { goto } from "$app/navigation";
    import { createId } from "@paralleldrive/cuid2";
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import DirectoryPicker from "$lib/components/DirectoryPicker.svelte";
    import Select from "$lib/components/Select.svelte";
    import {
        ArrowRight,
        ArrowLeft,
        Server,
        FolderInput,
        Check,
        Box,
        Scroll,
        Hammer,
        Layers,
        Cpu,
        HardDrive,
        Download,
    } from "lucide-svelte";
    import { fade, slide } from "svelte/transition";

    // --- State ---
    let step = 1;
    let name = "";
    let importMode = false;
    let showDirPicker = false;
    let sourcePath = "";
    let creating = false;
    let status = "";

    // Options
    let typeOptions = [];
    let versionOptions = [];
    let loadingLoaders = false;
    let loadingVersions = false;

    // Selections
    let type = "";
    let version = "";

    // Constants
    const SUPPORTED_LOADERS = ["fabric", "quilt", "forge", "neoforge"];
    const MANUAL_LOADERS = [
        { value: "paper", label: "Paper", icon: Scroll },
        { value: "spigot", label: "Spigot", icon: Layers },
        { value: "bukkit", label: "CraftBukkit", icon: Box },
    ];

    // Map loader names to icons and colors
    const LOADER_META = {
        fabric: {
            image: "/fabric.png",
            color: "text-amber-200",
            bg: "bg-amber-900/20",
            border: "border-amber-500/50",
        },
        quilt: {
            image: "/quilt.png",
            color: "text-purple-400",
            bg: "bg-purple-900/20",
            border: "border-purple-500/50",
        },
        forge: {
            image: "/forge.png",
            color: "text-orange-400",
            bg: "bg-orange-900/20",
            border: "border-orange-500/50",
        },
        neoforge: {
            image: "/neoforge.png",
            color: "text-orange-500",
            bg: "bg-orange-900/20",
            border: "border-orange-600/50",
        },
        paper: {
            image: "/paper.png",
            color: "text-blue-400",
            bg: "bg-blue-900/20",
            border: "border-blue-500/50",
        },
        spigot: {
            image: "/spigot.png",
            color: "text-yellow-400",
            bg: "bg-yellow-900/20",
            border: "border-yellow-500/50",
        },
        bukkit: {
            image: "/bukkit.png",
            color: "text-red-400",
            bg: "bg-red-900/20",
            border: "border-red-500/50",
        },
        vanilla: {
            icon: Box,
            color: "text-green-400",
            bg: "bg-green-900/20",
            border: "border-green-500/50",
        },
        custom: {
            icon: Cpu,
            color: "text-gray-400",
            bg: "bg-gray-800",
            border: "border-gray-600",
        },
    };

    function getLoaderMeta(value) {
        return LOADER_META[value] || LOADER_META.custom;
    }

    // --- Logic ---

    async function loadLoaders() {
        loadingLoaders = true;
        try {
            const res = await fetch("/api/versions/loader");
            if (res.ok) {
                const data = await res.json();
                const available = data
                    .filter((l) => SUPPORTED_LOADERS.includes(l.name))
                    .map((l) => ({
                        value: l.name,
                        label: l.name.charAt(0).toUpperCase() + l.name.slice(1),
                        ...getLoaderMeta(l.name),
                    }));

                typeOptions = [
                    ...available,
                    ...MANUAL_LOADERS.map((l) => ({
                        ...l,
                        ...getLoaderMeta(l.value),
                    })),
                    {
                        value: "custom",
                        label: "Custom Jar",
                        ...getLoaderMeta("custom"),
                    },
                ];

                // Add Vanilla/Paper manually if not present?
                // For now sticking to what the API returns + Spigot/Custom
            }
        } catch (e) {
            console.error(e);
            addToast("Failed to load server types", "error");
        } finally {
            loadingLoaders = false;
        }
    }

    async function loadVersions() {
        loadingVersions = true;
        try {
            const res = await fetch("/api/versions/game");
            if (res.ok) {
                const data = await res.json();
                versionOptions = data
                    .filter((v) => v.version_type === "release")
                    .map((v) => ({
                        value: v.version,
                        label: v.version,
                    }));
                if (!version && versionOptions.length > 0)
                    version = versionOptions[0].value;
            }
        } catch (e) {
            console.error(e);
        } finally {
            loadingVersions = false;
        }
    }

    function handleNext() {
        if (step === 1) {
            if (!name.trim())
                return addToast("Please name your server", "error");
            step = 2;
            if (!importMode && typeOptions.length === 0) loadLoaders();
            if (!importMode && versionOptions.length === 0) loadVersions();
        } else if (step === 2) {
            // If we are selecting type, ensure type is selected
            if (!importMode && !type)
                return addToast("Select a server software", "error");
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
        if (!sourcePath) return addToast("Source path is required", "error");
        creating = true;
        status = "Importing...";

        try {
            const id = createId();
            const res = await fetch("/api/instances/import", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ id, name, sourcePath }),
            });

            if (res.ok) {
                addToast("Import successful!", "success");
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

            // 1. Create
            status = "Creating instance directory and config...";
            const res = await fetch("/api/instances", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });

            if (!res.ok) throw await res.text();

            // 2. Install (if not custom)
            if (type !== "custom") {
                status = `Installing ${type} ${version} server...`;

                // Note: The backend installation can take a while (e.g. BuildTools/Download)
                // We should probably optimize this to return async task ID,
                // but for now we rely on the long-polling request.
                // We update status to be informative.
                if (type === "spigot" || type === "bukkit") {
                    status = `Compiling ${type} (this may take several minutes)...`;
                }

                const installRes = await fetch(`/api/instances/${id}/install`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ version, type }),
                });

                if (!installRes.ok) {
                    addToast(
                        "Instance created but install failed. Check logs.",
                        "warning",
                    );
                } else {
                    addToast("Server created successfully!", "success");
                }
            } else {
                addToast("Server created! Please upload your jar.", "success");
            }

            status = "Finalizing settings...";
            // Slight delay to let user see "Finished" state if we wanted
            await new Promise((r) => setTimeout(r, 500));

            await goto(`/instances/${id}`);
        } catch (e) {
            addToast(`Creation failed: ${e}`, "error");
        } finally {
            creating = false;
        }
    }

    onMount(() => {
        // Preload
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
    class="min-h-screen bg-gray-950 text-gray-100 flex flex-col items-center justify-center p-4 relative overflow-hidden"
>
    <!-- Background Effects -->
    <div
        class="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none z-0"
    >
        <div
            class="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-900/20 rounded-full blur-[128px]"
        ></div>
        <div
            class="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-purple-900/20 rounded-full blur-[128px]"
        ></div>
    </div>

    <!-- Main Card -->
    <div
        class="w-full max-w-4xl bg-gray-900/60 backdrop-blur-xl border border-gray-800 rounded-2xl shadow-2xl relative z-10 overflow-hidden flex flex-col md:flex-row min-h-[500px]"
    >
        <!-- Left Sidebar / Stepper -->
        <div
            class="w-full md:w-1/3 bg-gray-900/80 border-b md:border-b-0 md:border-r border-gray-800 p-8 flex flex-col justify-between"
        >
            <div>
                <button
                    on:click={() => goto("/instances")}
                    class="flex items-center gap-2 text-gray-500 hover:text-white mb-8 transition-colors text-sm font-medium"
                >
                    <ArrowLeft size={16} /> Cancel
                </button>

                <h1
                    class="text-2xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-purple-400 mb-2"
                >
                    Create Server
                </h1>
                <p class="text-gray-400 text-sm mb-8">
                    Set up your new Minecraft instance in just a few steps.
                </p>

                <div class="space-y-6">
                    <!-- Step 1 Indicator -->
                    <div
                        class="flex items-start gap-4 transition-colors {step >=
                        1
                            ? 'opacity-100'
                            : 'opacity-40'}"
                    >
                        <div class="flex flex-col items-center gap-2">
                            <div
                                class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold {step >=
                                1
                                    ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/50'
                                    : 'bg-gray-800 text-gray-500'} transition-all duration-300"
                            >
                                {#if step > 1}
                                    <Check size={16} />
                                {:else}
                                    1
                                {/if}
                            </div>
                            <div class="w-0.5 h-10 bg-gray-800"></div>
                        </div>
                        <div class="pt-1">
                            <h3
                                class="font-medium {step >= 1
                                    ? 'text-white'
                                    : 'text-gray-500'}"
                            >
                                Identity
                            </h3>
                            <p class="text-xs text-gray-500 mt-1">
                                Name and Type
                            </p>
                        </div>
                    </div>

                    <!-- Step 2 Indicator -->
                    <div
                        class="flex items-start gap-4 transition-colors {step >=
                        2
                            ? 'opacity-100'
                            : 'opacity-40'}"
                    >
                        <div class="flex flex-col items-center gap-2">
                            <div
                                class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold {step >=
                                2
                                    ? (importMode
                                          ? 'bg-emerald-600'
                                          : 'bg-blue-600') +
                                      ' text-white shadow-lg'
                                    : 'bg-gray-800 text-gray-500'} transition-all duration-300"
                            >
                                {#if step > 2}
                                    <Check size={16} />
                                {:else}
                                    2
                                {/if}
                            </div>
                            <div class="w-0.5 h-10 bg-gray-800"></div>
                        </div>
                        <div class="pt-1">
                            <h3
                                class="font-medium {step >= 2
                                    ? 'text-white'
                                    : 'text-gray-500'}"
                            >
                                {importMode ? "Source" : "Platform"}
                            </h3>
                            <p class="text-xs text-gray-500 mt-1">
                                {importMode
                                    ? "Folder path"
                                    : "Software selection"}
                            </p>
                        </div>
                    </div>

                    <!-- Step 3 Indicator -->
                    {#if !importMode}
                        <div
                            class="flex items-start gap-4 transition-colors {step >=
                            3
                                ? 'opacity-100'
                                : 'opacity-40'}"
                        >
                            <div class="flex flex-col items-center gap-2">
                                <div
                                    class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold {step >=
                                    3
                                        ? 'bg-blue-600 text-white shadow-lg'
                                        : 'bg-gray-800 text-gray-500'} transition-all duration-300"
                                >
                                    3
                                </div>
                            </div>
                            <div class="pt-1">
                                <h3
                                    class="font-medium {step >= 3
                                        ? 'text-white'
                                        : 'text-gray-500'}"
                                >
                                    Version
                                </h3>
                                <p class="text-xs text-gray-500 mt-1">
                                    Game Version
                                </p>
                            </div>
                        </div>
                    {/if}
                </div>
            </div>

            <div class="text-xs text-gray-600 mt-8">
                JJMC Instance Manager v1.0
            </div>
        </div>

        <!-- Right Content Area -->
        <div class="flex-1 p-8 bg-black/20 relative">
            {#if creating}
                <div
                    class="absolute inset-0 z-20 bg-gray-900/80 backdrop-blur-sm flex flex-col items-center justify-center p-8 text-center"
                    transition:fade
                >
                    <div class="relative w-20 h-20 mb-6">
                        <div
                            class="absolute inset-0 rounded-full border-4 border-gray-800"
                        ></div>
                        <div
                            class="absolute inset-0 rounded-full border-4 border-t-blue-500 border-r-transparent border-b-transparent border-l-transparent animate-spin"
                        ></div>
                    </div>
                    <h2 class="text-xl font-bold text-white mb-2">{status}</h2>
                    <p class="text-gray-400 text-sm">
                        Please wait while we set up your server.
                    </p>
                </div>
            {/if}

            <div class="h-full flex flex-col">
                {#if step === 1}
                    <div
                        in:fade={{ duration: 200, delay: 100 }}
                        class="flex-1 flex flex-col justify-center max-w-lg mx-auto w-full"
                    >
                        <h2 class="text-3xl font-bold text-white mb-6">
                            Let's start.
                        </h2>

                        <div class="space-y-6">
                            <div>
                                <label
                                    for="name"
                                    class="block text-sm font-medium text-gray-400 mb-2"
                                    >Server Name</label
                                >
                                <input
                                    type="text"
                                    id="name"
                                    bind:value={name}
                                    class="w-full bg-gray-800 border border-gray-700 text-white rounded-xl px-4 py-3 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-lg placeholder-gray-600"
                                    placeholder="My Awesome Server"
                                    autofocus
                                    on:keydown={(e) =>
                                        e.key === "Enter" && handleNext()}
                                />
                            </div>

                            <div class="grid grid-cols-2 gap-4">
                                <button
                                    class="relative p-4 rounded-xl border transition-all text-left group
                                    {!importMode
                                        ? 'bg-blue-600/10 border-blue-500 ring-1 ring-blue-500/50'
                                        : 'bg-gray-800/50 border-gray-700 hover:bg-gray-800 hover:border-gray-600'}"
                                    on:click={() => (importMode = false)}
                                >
                                    <div
                                        class="p-2 w-10 h-10 rounded-lg {!importMode
                                            ? 'bg-blue-500 text-white'
                                            : 'bg-gray-700 text-gray-400'} flex items-center justify-center mb-3 transition-colors"
                                    >
                                        <Server size={20} />
                                    </div>
                                    <h3 class="text-white font-medium mb-1">
                                        New Server
                                    </h3>
                                    <p
                                        class="text-xs text-gray-400 leading-relaxed"
                                    >
                                        Install a fresh Minecraft server from a
                                        template.
                                    </p>

                                    {#if !importMode}
                                        <div
                                            class="absolute top-4 right-4 text-blue-400"
                                        >
                                            <Check size={16} />
                                        </div>
                                    {/if}
                                </button>

                                <button
                                    class="relative p-4 rounded-xl border transition-all text-left group
                                    {importMode
                                        ? 'bg-emerald-600/10 border-emerald-500 ring-1 ring-emerald-500/50'
                                        : 'bg-gray-800/50 border-gray-700 hover:bg-gray-800 hover:border-gray-600'}"
                                    on:click={() => (importMode = true)}
                                >
                                    <div
                                        class="p-2 w-10 h-10 rounded-lg {importMode
                                            ? 'bg-emerald-500 text-white'
                                            : 'bg-gray-700 text-gray-400'} flex items-center justify-center mb-3 transition-colors"
                                    >
                                        <FolderInput size={20} />
                                    </div>
                                    <h3 class="text-white font-medium mb-1">
                                        Import Existing
                                    </h3>
                                    <p
                                        class="text-xs text-gray-400 leading-relaxed"
                                    >
                                        Add a server that already exists on
                                        disk.
                                    </p>

                                    {#if importMode}
                                        <div
                                            class="absolute top-4 right-4 text-emerald-400"
                                        >
                                            <Check size={16} />
                                        </div>
                                    {/if}
                                </button>
                            </div>
                        </div>

                        <div class="mt-8 flex justify-end">
                            <button
                                on:click={handleNext}
                                class="bg-white text-gray-900 px-6 py-3 rounded-xl font-bold hover:bg-gray-200 transition-colors flex items-center gap-2 shadow-lg shadow-white/10"
                            >
                                Next Step <ArrowRight size={18} />
                            </button>
                        </div>
                    </div>
                {:else if step === 2}
                    <div
                        in:fade={{ duration: 200, delay: 100 }}
                        class="flex-1 flex flex-col justify-center max-w-lg mx-auto w-full"
                    >
                        {#if importMode}
                            <h2 class="text-3xl font-bold text-white mb-6">
                                Locate Server
                            </h2>
                            <div class="space-y-4">
                                <div>
                                    <label
                                        class="block text-sm font-medium text-gray-400 mb-2"
                                        >Folder Path</label
                                    >
                                    <div class="flex gap-2">
                                        <div class="relative flex-1">
                                            <div
                                                class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-gray-500"
                                            >
                                                <HardDrive size={18} />
                                            </div>
                                            <input
                                                type="text"
                                                bind:value={sourcePath}
                                                class="w-full bg-gray-800 border border-gray-700 text-white rounded-xl pl-10 pr-4 py-3 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all font-mono text-sm"
                                                placeholder="/absolute/path/to/server"
                                            />
                                        </div>
                                        <button
                                            on:click={() =>
                                                (showDirPicker = true)}
                                            class="bg-gray-800 hover:bg-gray-700 border border-gray-700 text-gray-300 px-4 rounded-xl transition-colors"
                                        >
                                            Browse
                                        </button>
                                    </div>
                                    <p class="text-xs text-gray-500 mt-2">
                                        Ensure this folder contains your <code
                                            >server.jar</code
                                        >
                                        and <code>server.properties</code>. We
                                        will copy it to the instances storage.
                                    </p>
                                </div>
                            </div>

                            <div class="mt-8 flex justify-between items-center">
                                <button
                                    on:click={handleBack}
                                    class="text-gray-400 hover:text-white px-4 py-2 font-medium transition-colors"
                                    >Back</button
                                >
                                <button
                                    on:click={finish}
                                    class="bg-emerald-500 text-white px-6 py-3 rounded-xl font-bold hover:bg-emerald-400 transition-colors flex items-center gap-2 shadow-lg shadow-emerald-500/20"
                                >
                                    <Download size={18} /> Import Server
                                </button>
                            </div>
                        {:else}
                            <h2 class="text-3xl font-bold text-white mb-6">
                                Choose Software
                            </h2>

                            <div
                                class="grid grid-cols-2 gap-3 max-h-[400px] overflow-y-auto pr-2 custom-scrollbar"
                            >
                                {#each typeOptions as option}
                                    <button
                                        class="relative p-4 rounded-xl border text-left transition-all
                                        {type === option.value
                                            ? `${option.bg} ${option.border} ring-1 ring-blue-500/30`
                                            : 'bg-gray-800/40 border-gray-700 hover:bg-gray-800 hover:border-gray-600'}"
                                        on:click={() => {
                                            type = option.value;
                                            handleNext();
                                        }}
                                    >
                                        <div
                                            class="flex items-center gap-3 mb-2"
                                        >
                                            {#if option.image}
                                                <img
                                                    src={option.image}
                                                    alt={option.label}
                                                    class="w-10 h-10 object-contain"
                                                />
                                            {:else}
                                                <div
                                                    class="p-2 rounded-lg bg-gray-900/50 {option.color}"
                                                >
                                                    <svelte:component
                                                        this={option.icon ||
                                                            Box}
                                                        size={20}
                                                    />
                                                </div>
                                            {/if}
                                            <h3
                                                class="text-white font-semibold"
                                            >
                                                {option.label}
                                            </h3>
                                        </div>
                                        {#if type === option.value}
                                            <div
                                                class="absolute top-1/2 right-4 -translate-y-1/2 text-blue-400"
                                            >
                                                <ArrowRight size={20} />
                                            </div>
                                        {/if}
                                    </button>
                                {/each}
                            </div>
                            <div class="mt-8 flex justify-between items-center">
                                <button
                                    on:click={handleBack}
                                    class="text-gray-400 hover:text-white px-4 py-2 font-medium transition-colors"
                                    >Back</button
                                >
                            </div>
                        {/if}
                    </div>
                {:else if step === 3}
                    <div
                        in:fade={{ duration: 200, delay: 100 }}
                        class="flex-1 flex flex-col justify-center max-w-lg mx-auto w-full"
                    >
                        <h2 class="text-3xl font-bold text-white mb-2">
                            Select Version
                        </h2>
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
                                    You will need to manually upload your server
                                    JAR file after creation.
                                </p>
                            </div>
                        {:else}
                            <div class="space-y-4">
                                <label
                                    class="block text-sm font-medium text-gray-400"
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
                {/if}
            </div>
        </div>
    </div>
</div>

<style>
    /* Clean scrollbar for list */
    .custom-scrollbar::-webkit-scrollbar {
        width: 6px;
    }
    .custom-scrollbar::-webkit-scrollbar-track {
        background: rgba(0, 0, 0, 0.1);
        border-radius: 3px;
    }
    .custom-scrollbar::-webkit-scrollbar-thumb {
        background: rgba(255, 255, 255, 0.1);
        border-radius: 3px;
    }
    .custom-scrollbar::-webkit-scrollbar-thumb:hover {
        background: rgba(255, 255, 255, 0.2);
    }
</style>
