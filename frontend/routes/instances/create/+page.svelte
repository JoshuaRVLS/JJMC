<script>
    import { goto } from "$app/navigation";
    import { createId } from "@paralleldrive/cuid2";
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import {
        ArrowLeft,
        Check,
        Box,
        BookOpen,
        Scroll,
        Layers,
        Cpu,
    } from "lucide-svelte";
    import { fade } from "svelte/transition";
    import DirectoryPicker from "$lib/components/DirectoryPicker.svelte";
    import StepBasicDetails from "$lib/creation/StepBasicDetails.svelte";
    import StepLoader from "$lib/creation/StepLoader.svelte";
    import StepVersion from "$lib/creation/StepVersion.svelte";

    // --- State ---
    let step = 1;
    let name = "";
    let importMode = false;
    let sourcePath = "";
    let showDirPicker = false;

    let type = "";
    let version = "";

    // Options
    let versionOptions = [];
    let loadingLoaders = false;
    let loadingVersions = false;
    let creating = false;
    let status = "";

    // Loaders (Static + Dynamic types)
    // We map backend types to friendly UI items
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
        fabric: {
            image: "/fabric.png",
            color: "text-amber-200",
            bg: "bg-amber-900/20",
            border: "border-amber-500/50",
        },
        quilt: {
            image: "/quilt.png",
            color: "text-emerald-200",
            bg: "bg-emerald-900/20",
            border: "border-emerald-500/50",
        },
        forge: {
            image: "/forge.png",
            color: "text-orange-200",
            bg: "bg-orange-900/20",
            border: "border-orange-500/50",
        },
        neoforge: {
            image: "/neoforge.png",
            color: "text-orange-200",
            bg: "bg-orange-900/20",
            border: "border-orange-500/50",
        },
        paper: {
            icon: Scroll,
            color: "text-blue-200",
            bg: "bg-blue-900/20",
            border: "border-blue-500/50",
        },
        spigot: {
            icon: Layers,
            color: "text-orange-400",
            bg: "bg-orange-900/20",
            border: "border-orange-500/50",
        },
        bukkit: {
            icon: Box,
            color: "text-yellow-400",
            bg: "bg-yellow-900/20",
            border: "border-yellow-500/50",
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

    // Enrich static options with meta styles
    $: typeOptions = typeOptions.map((opt) => ({
        ...opt,
        ...(LOADER_META[opt.value] || LOADER_META.custom),
    }));

    // --- Logic ---

    async function loadLoaders() {
        loadingLoaders = true;
        try {
            // If we wanted to fetch available loaders from backend, we could.
            // But we have a static list for now, we just want to verify connectivity?
            // Or maybe fetch latest versions for each loader later.
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
        version = ""; // Reset selection
        versionOptions = [];

        try {
            // For Paper/Spigot etc we check backend, for Fabric/Forge we might check different endpoints
            // simplified:
            const res = await fetch(`/api/versions/game?loader=${type}`);
            if (res.ok) {
                const data = await res.json();
                versionOptions = data.map((v) => ({
                    value: v,
                    label: v,
                }));
                if (versionOptions.length > 0) {
                    version = versionOptions[0].value;
                }
            } else {
                // heavy fallback or just show empty
                // addToast("Failed to load versions for " + type, "error");
                // For now mockup:
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
                // Check if path is entered
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

            // 1. Create
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
                // We will poll or wait. For now, assuming synchronous-ish for simple types,
                // or the user can go to the dashboard and see progress in console.
                // But typically we want at least initial files setup.

                // Actually, backend usually handles "install" logic async, but let's assume
                // we interpret "success" = "created".
                // If we want to trigger specific install logic:
                const installRes = await fetch(`/api/instances/${id}/install`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ version }),
                });

                if (!installRes.ok) {
                    // It might fail if e.g. version not found, but instance is created.
                    addToast(
                        "Instance created, but installation failed. Check console.",
                        "warning",
                    );
                }
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
            class="absolute -top-20 -left-20 w-96 h-96 bg-blue-500/10 rounded-full blur-[100px] animate-pulse"
        ></div>
        <div
            class="absolute top-1/2 -right-20 w-96 h-96 bg-purple-500/10 rounded-full blur-[100px] animate-pulse delay-700"
        ></div>
        <div
            class="absolute -bottom-20 left-1/3 w-96 h-96 bg-emerald-500/10 rounded-full blur-[100px] animate-pulse delay-1000"
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
