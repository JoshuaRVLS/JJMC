<script>
    import { onMount } from "svelte";
    import { Plus, Trash2, Loader2, Download } from "lucide-svelte";
    import { addToast } from "$lib/stores/toast";

    let runtimes = [];
    let loading = true;
    let installing = false;
    let isModalOpen = false;

    let availableVersions = [8, 11, 17, 21];
    let selectedVersion = 17;

    async function loadRuntimes() {
        loading = true;
        try {
            const res = await fetch("/api/java/installed");
            if (res.ok) {
                runtimes = (await res.json()) || [];
            } else {
                addToast("Failed to load Java runtimes", "error");
            }
        } catch (e) {
            console.error(e);
            addToast("Error loading Java runtimes", "error");
        } finally {
            loading = false;
        }
    }

    async function installJava() {
        installing = true;
        try {
            const res = await fetch("/api/java/install", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ version: selectedVersion }),
            });

            if (res.ok) {
                addToast(
                    `Installing Java ${selectedVersion}... This may take a while.`,
                    "info",
                );
                isModalOpen = false;
                // Polling will pick up the new installation
            } else {
                addToast("Failed to start installation", "error");
            }
        } catch (e) {
            console.error(e);
            addToast("Error installing Java", "error");
        } finally {
            installing = false;
        }
    }

    async function deleteRuntime(name) {
        if (!confirm(`Are you sure you want to delete ${name}?`)) return;

        try {
            const res = await fetch(`/api/java/${name}`, {
                method: "DELETE",
            });
            if (res.ok) {
                addToast("Runtime deleted", "success");
                loadRuntimes();
            } else {
                addToast("Failed to delete runtime", "error");
            }
        } catch (e) {
            console.error(e);
            addToast("Error deleting runtime", "error");
        }
    }

    onMount(() => {
        loadRuntimes();
        // Start polling
        const interval = setInterval(loadRuntimes, 1000);
        return () => clearInterval(interval);
    });
</script>

<div class="h-full flex flex-col gap-6 p-6">
    <div class="flex justify-between items-center">
        <div>
            <h2 class="text-xl font-bold text-white">Java Versions</h2>
            <p class="text-gray-400 text-sm">Manage installed Java Runtimes</p>
        </div>
        <button
            on:click={() => (isModalOpen = true)}
            class="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-colors font-medium shadow-lg shadow-indigo-900/20"
        >
            <Download class="w-4 h-4" />
            Install Version
        </button>
    </div>

    {#if loading && (!runtimes || runtimes.length === 0)}
        <div class="flex-1 flex items-center justify-center">
            <Loader2 class="w-8 h-8 text-indigo-500 animate-spin" />
        </div>
    {:else if !runtimes || runtimes.length === 0}
        <div
            class="flex-1 flex flex-col items-center justify-center text-gray-500 gap-4"
        >
            <p>No managed Java versions installed</p>
        </div>
    {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {#each runtimes as runtime}
                <div
                    class="bg-gray-900/50 border border-white/5 p-4 rounded-xl flex justify-between items-center group hover:border-indigo-500/30 transition-colors"
                >
                    <div class="w-full">
                        <div class="flex justify-between items-start mb-2">
                            <h3 class="font-bold text-white">
                                Java {runtime.version}
                            </h3>
                            {#if runtime.status && runtime.status !== "Ready"}
                                <div
                                    class="text-xs font-mono text-indigo-400 animate-pulse"
                                >
                                    {runtime.status}
                                </div>
                            {:else}
                                <button
                                    on:click={() => deleteRuntime(runtime.name)}
                                    class="p-2 text-gray-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
                                >
                                    <Trash2 class="w-4 h-4" />
                                </button>
                            {/if}
                        </div>

                        {#if runtime.status && runtime.status !== "Ready"}
                            <div
                                class="w-full bg-gray-700 rounded-full h-1.5 mt-2 overflow-hidden"
                            >
                                <div
                                    class="bg-indigo-500 h-1.5 rounded-full transition-all duration-300"
                                    style="width: {runtime.progress}%"
                                ></div>
                            </div>
                            <div
                                class="flex justify-between text-xs text-gray-500 mt-1"
                            >
                                <span>{runtime.status}...</span>
                                <span>{runtime.progress}%</span>
                            </div>
                        {:else}
                            <p
                                class="text-xs text-gray-500 font-mono break-all"
                            >
                                {runtime.path}
                            </p>
                        {/if}
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>

{#if isModalOpen}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
    >
        <div
            class="bg-gray-900 border border-white/10 rounded-2xl w-full max-w-md shadow-2xl p-6"
        >
            <h3 class="text-xl font-bold text-white mb-6">
                Install Java Version
            </h3>

            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-400 mb-2"
                        >Select Version</label
                    >
                    <div class="grid grid-cols-2 gap-2">
                        {#each availableVersions as version}
                            <button
                                on:click={() => (selectedVersion = version)}
                                class="px-4 py-3 rounded-lg border transition-all text-sm font-bold {selectedVersion ===
                                version
                                    ? 'bg-indigo-500/10 border-indigo-500 text-indigo-400'
                                    : 'bg-black/20 border-white/5 text-gray-400 hover:border-white/10'}"
                            >
                                Java {version}
                            </button>
                        {/each}
                    </div>
                </div>
            </div>

            <div class="flex justify-end gap-3 mt-8">
                <button
                    on:click={() => (isModalOpen = false)}
                    class="px-4 py-2 text-gray-400 hover:text-white transition-colors"
                >
                    Cancel
                </button>
                <button
                    on:click={installJava}
                    disabled={installing}
                    class="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                >
                    {#if installing}
                        <Loader2 class="w-4 h-4 animate-spin" />
                        Installing...
                    {:else}
                        Install
                    {/if}
                </button>
            </div>
        </div>
    </div>
{/if}
