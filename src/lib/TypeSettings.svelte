<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import Select from "$lib/components/Select.svelte";
    import { askConfirm } from "$lib/stores/confirm";

    /** @type {string} */
    export let instanceId;

    /** @type {any[]} */
    let typeOptions = [];
    let loadingLoaders = false;
    let type = "";
    let version = "";
    let changing = false;
    let loading = true;
    let currentType = "";
    let currentVersion = "";

    /** @type {any[]} */
    let versionOptions = [];
    let loadingVersions = false;

    const SUPPORTED_LOADERS = ["fabric", "quilt", "forge", "neoforge"];
    const MANUAL_LOADERS = [{ value: "spigot", label: "Spigot" }];

    async function loadInstanceDetails() {
        try {
            const res = await fetch(`/api/instances/${instanceId}`);
            if (res.ok) {
                const data = await res.json();
                currentType = data.type;
                currentVersion = data.version;

                if (!type) {
                    type = currentType;
                    version = currentVersion;
                }
            }
        } catch (e) {
            console.error("Failed to load instance details", e);
        }
    }

    async function loadLoaders() {
        loadingLoaders = true;
        try {
            const res = await fetch("/api/versions/loader");
            if (res.ok) {
                const data = await res.json();
                const available = data
                    .filter((/** @type {any} */ l) =>
                        SUPPORTED_LOADERS.includes(l.name),
                    )
                    .map((/** @type {any} */ l) => ({
                        value: l.name,
                        label: l.name.charAt(0).toUpperCase() + l.name.slice(1),
                    }));

                typeOptions = [
                    ...available,
                    ...MANUAL_LOADERS,
                    { value: "custom", label: "Custom Jar" },
                ];
            }
        } catch (e) {
            console.error("Failed to load loaders", e);
            addToast("Failed to load server types", "error");

            typeOptions = [
                { value: "fabric", label: "Fabric" },
                { value: "quilt", label: "Quilt" },
                { value: "forge", label: "Forge" },
                { value: "neoforge", label: "NeoForge" },
                ...MANUAL_LOADERS,
                { value: "custom", label: "Custom Jar" },
            ];
        } finally {
            loadingLoaders = false;
        }
    }

    async function loadVersions() {
        if (type === "custom") {
            versionOptions = [];
            return;
        }
        loadingVersions = true;
        try {
            const res = await fetch("/api/versions/game");
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

    $: if (type) {
        loadVersions();
    }

    async function changeType() {
        const confirmed = await askConfirm({
            title: "Change Server Type?",
            message:
                "This action will DELETE all mods, configs, plugins, and server jar files. This cannot be undone. Are you sure you want to proceed?",
            confirmText: "Yes, Change Type",
            dangerous: true,
        });

        if (!confirmed) return;

        changing = true;
        try {
            const res = await fetch(`/api/instances/${instanceId}/type`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    type,
                    version: type === "custom" ? "" : version,
                }),
            });

            if (res.ok) {
                addToast("Server type changed successfully!", "success");

                await loadInstanceDetails();
            } else {
                const data = await res.json();
                addToast(`Failed: ${data.error}`, "error");
            }
        } catch (e) {
            console.error("Change type failed", e);
            addToast("Failed to change server type", "error");
        } finally {
            changing = false;
        }
    }

    onMount(async () => {
        await Promise.all([loadInstanceDetails(), loadLoaders()]);
        loading = false;
    });

    $: isChanged = type !== currentType || version !== currentVersion;
</script>

<div class="h-full overflow-y-auto pr-2">
    {#if loading}
        <div class="text-center text-gray-500 py-10">Loading settings...</div>
    {:else}
        <div
            class="bg-gray-900/60 backdrop-blur-xl border border-white/5 rounded-2xl p-6 shadow-lg max-w-2xl mx-auto space-y-8"
        >
            <div>
                <h2 class="text-xl font-bold text-white mb-2">
                    Change Server Type
                </h2>
                <div
                    class="p-4 bg-yellow-500/10 border border-yellow-500/20 rounded-lg text-yellow-200 text-sm"
                >
                    <strong>Warning:</strong> Changing the server type will
                    <span class="text-yellow-400 font-bold underline"
                        >delete all server files</span
                    >
                    (mods, configs, plugins, libraries). Your world data and server.properties
                    will be preserved.
                </div>
            </div>

            <div class="space-y-6">
                <div class="grid grid-cols-2 gap-4">
                    <div
                        class="bg-black/20 p-4 rounded-lg border border-white/5"
                    >
                        <div
                            class="text-xs text-gray-500 uppercase tracking-widest font-bold mb-1"
                        >
                            Current Type
                        </div>
                        <div class="text-lg text-white font-medium capitalize">
                            {currentType || "Unknown"}
                        </div>
                    </div>
                    <div
                        class="bg-black/20 p-4 rounded-lg border border-white/5"
                    >
                        <div
                            class="text-xs text-gray-500 uppercase tracking-widest font-bold mb-1"
                        >
                            Current Version
                        </div>
                        <div class="text-lg text-white font-medium">
                            {currentVersion || "Unknown"}
                        </div>
                    </div>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                        <Select
                            label="New Server Type"
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

                <div class="pt-4 flex justify-end border-t border-white/5">
                    <button
                        on:click={changeType}
                        disabled={changing || !type || !isChanged}
                        class="bg-red-600 hover:bg-red-500 text-white px-6 py-2 rounded-lg font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                    >
                        {#if changing}
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
                            Changing...
                        {:else}
                            Switch Server Type
                        {/if}
                    </button>
                </div>
            </div>
        </div>
    {/if}
</div>
