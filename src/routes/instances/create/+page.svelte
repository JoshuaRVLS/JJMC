<script>
    import { goto } from "$app/navigation";
    import { createId } from "@paralleldrive/cuid2";
    import Select from "$lib/components/Select.svelte";
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import DirectoryPicker from "$lib/components/DirectoryPicker.svelte";

    let step = 1;
    let name = "";

    /**
     * @typedef {Object} Option
     * @property {string} value
     * @property {string} label
     */

    /** @type {Option[]} */
    let typeOptions = [];
    let loadingLoaders = false;
    let type = "";
    let version = "";
    let creating = false;
    let status = "";

    /** @type {Option[]} */
    let versionOptions = [];
    let loadingVersions = false;

    // Loaders we currently support with automated installers
    const SUPPORTED_LOADERS = ["fabric", "quilt", "forge", "neoforge"];

    const MANUAL_LOADERS = [{ value: "spigot", label: "Spigot" }];

    async function loadLoaders() {
        loadingLoaders = true;
        try {
            const res = await fetch("/api/versions/loader");
            if (res.ok) {
                const data = await res.json();
                // Filter and map
                const available = data
                    .filter((/** @type {{name: string}} */ l) =>
                        SUPPORTED_LOADERS.includes(l.name),
                    )
                    .map((/** @type {{name: string}} */ l) => ({
                        value: l.name,
                        label: l.name.charAt(0).toUpperCase() + l.name.slice(1), // Capitalize
                    }));

                typeOptions = [
                    ...available,
                    ...MANUAL_LOADERS,
                    { value: "custom", label: "Custom Jar" },
                ];

                if (!type && typeOptions.length > 0) {
                    type = typeOptions[0].value;
                }
            }
        } catch (e) {
            console.error("Failed to load loaders", e);
            addToast("Failed to load server types", "error");
            // Fallback
            typeOptions = [
                { value: "fabric", label: "Fabric" },
                { value: "quilt", label: "Quilt" },
                { value: "forge", label: "Forge" },
                { value: "neoforge", label: "NeoForge" },
                ...MANUAL_LOADERS,
                { value: "custom", label: "Custom Jar" },
            ];
            if (!type) type = "fabric";
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
                    .filter(
                        (
                            /** @type {{version: string, version_type: string}} */ v,
                        ) => v.version_type === "release",
                    )
                    .map((/** @type {{version: string}} */ v) => ({
                        value: v.version,
                        label: v.version,
                    }));

                if (!version && versionOptions.length > 0) {
                    version = versionOptions[0].value;
                }
            }
        } catch (e) {
            console.error("Failed to load versions", e);
            addToast("Failed to load versions", "error");
        } finally {
            loadingVersions = false;
        }
    }

    function nextStep() {
        if (!name.trim()) {
            addToast("Please enter an instance name", "error");
            return;
        }
        step = 2;
        // Trigger load versions if not loaded
        if (versionOptions.length === 0) {
            loadVersions();
        }
    }

    let importMode = false;
    let showDirPicker = false;
    let sourcePath = "";

    async function create() {
        if (!name) return addToast("Name is required", "error");

        if (importMode) {
            if (!sourcePath)
                return addToast("Source path is required", "error");
            creating = true;
            status = "Importing instance...";
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
                    addToast("Instance imported successfully!", "success");
                    await goto(`/instances/${id}`);
                } else {
                    const err = await res.text();
                    addToast(`Import failed: ${err}`, "error");
                }
            } catch (e) {
                console.error("Import failed", e);
                addToast("Failed to import instance", "error");
            } finally {
                creating = false;
                status = "";
            }
            return;
        }

        if (!type) return addToast("Server type is required", "error");

        creating = true;
        status = "Creating instance...";

        try {
            // Generate ID
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

            if (res.ok) {
                // If not custom, trigger install
                if (type !== "custom") {
                    status = "Installing server components...";
                    const installPayload = {
                        version,
                        type,
                    };

                    const installRes = await fetch(
                        `/api/instances/${id}/install`,
                        {
                            method: "POST",
                            headers: { "Content-Type": "application/json" },
                            body: JSON.stringify(installPayload),
                        },
                    );

                    if (!installRes.ok) {
                        const err = await installRes.text();
                        // We don't block navigation, but we warn
                        addToast(
                            `Instance created but install failed: ${err}`,
                            "error",
                        );
                    } else {
                        addToast("Instance created and installed!", "success");
                    }
                } else {
                    addToast("Instance created!", "success");
                }

                await goto(`/instances/${id}`);
            } else {
                const err = await res.text();
                addToast(`Failed: ${err}`, "error");
            }
        } catch (e) {
            console.error("Creation failed", e);
            addToast("Failed to create instance", "error");
        } finally {
            creating = false;
            status = "";
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

<div class="h-full flex flex-col items-center justify-center p-6 bg-gray-950">
    <div
        class="w-full max-w-xl bg-gray-900 rounded-lg border border-gray-800 shadow-xl overflow-hidden"
    >
        <!-- Header -->
        <div
            class="px-6 py-4 border-b border-gray-800 flex justify-between items-center bg-gray-900/50"
        >
            <div class="flex items-center gap-4">
                <h2 class="text-lg font-semibold text-white">
                    Create Instance
                </h2>
                <div
                    class="flex bg-gray-800 rounded-lg p-1 text-xs font-medium"
                >
                    <button
                        class="px-3 py-1 rounded-md transition-all {!importMode
                            ? 'bg-blue-600 text-white shadow-lg'
                            : 'text-gray-400 hover:text-white'}"
                        on:click={() => (importMode = false)}>Create New</button
                    >
                    <button
                        class="px-3 py-1 rounded-md transition-all {importMode
                            ? 'bg-emerald-600 text-white shadow-lg'
                            : 'text-gray-400 hover:text-white'}"
                        on:click={() => (importMode = true)}
                        >Import Existing</button
                    >
                </div>
            </div>
            <a
                href="/instances"
                class="text-xs font-medium text-gray-500 hover:text-white transition-colors"
                >CANCEL</a
            >
        </div>

        <div class="p-6">
            <!-- Simple Steps -->
            <div class="flex items-center gap-2 mb-6 text-sm">
                <span
                    class={step === 1
                        ? importMode
                            ? "text-emerald-400 font-bold"
                            : "text-blue-400 font-bold"
                        : "text-gray-500"}>1. Name</span
                >
                <span class="text-gray-700">/</span>
                <span
                    class={step === 2
                        ? importMode
                            ? "text-emerald-400 font-bold"
                            : "text-blue-400 font-bold"
                        : "text-gray-500"}
                    >2. {importMode ? "Source" : "Configuration"}</span
                >
            </div>

            <div class="min-h-[200px]">
                {#if step === 1}
                    <div class="space-y-4">
                        <div>
                            <label
                                class="block text-xs font-medium text-gray-400 mb-1.5"
                                for="instance-name">Instance Name</label
                            >
                            <input
                                id="instance-name"
                                type="text"
                                bind:value={name}
                                class="w-full bg-gray-800 border border-gray-700 rounded-md p-2.5 text-white placeholder-gray-600 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors"
                                placeholder={importMode
                                    ? "My Imported Server"
                                    : "My Server"}
                            />
                        </div>
                        <div class="flex justify-end pt-2">
                            <button
                                type="button"
                                on:click={nextStep}
                                class="{importMode
                                    ? 'bg-emerald-600 hover:bg-emerald-500'
                                    : 'bg-blue-600 hover:bg-blue-500'} text-white px-5 py-2 rounded-md font-medium text-sm transition-colors hover:shadow-lg flex items-center gap-2"
                            >
                                Continue
                            </button>
                        </div>
                    </div>
                {:else if step === 2}
                    <div class="space-y-4">
                        {#if importMode}
                            <div>
                                <label
                                    class="block text-xs font-medium text-gray-400 mb-1.5"
                                    for="source-path">Server Folder Path</label
                                >
                                <div class="relative flex gap-2">
                                    <input
                                        id="source-path"
                                        type="text"
                                        bind:value={sourcePath}
                                        class="w-full bg-gray-800 border border-gray-700 rounded-md p-2.5 text-white placeholder-gray-600 focus:outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 transition-colors"
                                        placeholder="/path/to/existing/server"
                                    />
                                    <button
                                        on:click={() => (showDirPicker = true)}
                                        class="bg-gray-700 hover:bg-gray-600 text-white px-3 rounded-md border border-gray-600 transition-colors"
                                        title="Browse"
                                    >
                                        <svg
                                            class="w-5 h-5"
                                            fill="none"
                                            stroke="currentColor"
                                            viewBox="0 0 24 24"
                                            ><path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z"
                                            ></path></svg
                                        >
                                    </button>
                                </div>
                                <p class="text-[10px] text-gray-500 mt-2">
                                    Files will be copied from this location to
                                    the new instance folder.
                                </p>
                            </div>
                        {:else}
                            <div class="grid grid-cols-1 gap-4">
                                <div>
                                    <Select
                                        label="Server Type"
                                        options={typeOptions}
                                        bind:value={type}
                                        placeholder="Select Type"
                                    />
                                </div>

                                {#if type !== "custom"}
                                    <div>
                                        <Select
                                            label="Version"
                                            options={versionOptions}
                                            bind:value={version}
                                            placeholder="Select Version"
                                        />
                                    </div>
                                {/if}
                            </div>
                        {/if}

                        <div
                            class="flex justify-between items-center pt-6 border-t border-gray-800 mt-6"
                        >
                            <button
                                on:click={() => (step = 1)}
                                class="text-gray-500 hover:text-white text-sm font-medium transition-colors"
                            >
                                Back
                            </button>
                            <button
                                on:click={create}
                                disabled={creating}
                                class="{importMode
                                    ? 'bg-emerald-600 hover:bg-emerald-500'
                                    : 'bg-blue-600 hover:bg-blue-500'} text-white px-5 py-2 rounded-md font-medium text-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                            >
                                {#if creating}
                                    <svg
                                        class="animate-spin h-4 w-4 text-white"
                                        xmlns="http://www.w3.org/2000/svg"
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
                                    {status ||
                                        (importMode
                                            ? "Importing..."
                                            : "Creating...")}
                                {:else}
                                    {importMode
                                        ? "Import Server"
                                        : "Create Server"}
                                {/if}
                            </button>
                        </div>
                    </div>
                {/if}
            </div>
        </div>
    </div>
</div>
