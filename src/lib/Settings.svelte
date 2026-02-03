<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";

    /** @type {string} */
    export let instanceId;

    let maxMemory = 2048;
    let javaArgs = "";
    let jarFile = "server.jar";
    let javaPath = "";
    let webhookUrl = "";
    let group = "";

    /**
     * @typedef {Object} FileEntry
     * @property {string} name
     * @property {boolean} [isDir]
     */

    /** @type {FileEntry[]} */
    let jarFiles = [];
    let installedRuntimes = [];
    let selectedJavaMode = "";
    let loading = true;
    let saving = false;

    async function loadRuntimes() {
        try {
            const res = await fetch("/api/java/installed");
            if (res.ok) {
                installedRuntimes = (await res.json()) || [];
            }
        } catch (e) {
            console.error("Failed to load runtimes", e);
        }
    }

    async function loadSettings() {
        loading = true;
        try {
            await loadRuntimes();

            const res = await fetch(`/api/instances/${instanceId}`);
            if (res.ok) {
                const data = await res.json();
                maxMemory = data.maxMemory || 2048;
                javaArgs = data.javaArgs || "";
                jarFile = data.jarFile || "server.jar";
                javaPath = data.javaPath || "";
                webhookUrl = data.webhookUrl || "";
                group = data.group || "";

                // Determine mode
                const knownRuntime = installedRuntimes.find(
                    (r) => r.path === javaPath,
                );
                if (javaPath === "") {
                    selectedJavaMode = "";
                } else if (knownRuntime) {
                    selectedJavaMode = javaPath;
                } else {
                    selectedJavaMode = "custom";
                }

                const mediaRes = await fetch(
                    `/api/instances/${instanceId}/files?path=.`,
                );
                if (mediaRes.ok) {
                    /** @type {FileEntry[]} */
                    const files = await mediaRes.json();
                    jarFiles = files.filter(
                        (f) => !f.isDir && f.name.endsWith(".jar"),
                    );

                    if (!jarFiles.find((f) => f.name === jarFile)) {
                        jarFiles = [...jarFiles, { name: jarFile }];
                    }
                }
            }
        } catch (/** @type {any} */ e) {
            addToast("Failed to load settings", "error");
        } finally {
            loading = false;
        }
    }

    $: if (selectedJavaMode !== "custom") {
        javaPath = selectedJavaMode;
    }

    async function saveSettings() {
        saving = true;
        try {
            const res = await fetch(`/api/instances/${instanceId}`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    maxMemory: parseInt(String(maxMemory)),
                    javaArgs: javaArgs,
                    jarFile: jarFile,
                    javaPath: javaPath,
                    webhookUrl: webhookUrl,
                    group: group,
                }),
            });
            if (!res.ok) throw new Error(await res.text());
            addToast("Settings saved", "success");
        } catch (/** @type {any} */ e) {
            addToast("Failed to save settings: " + e.message, "error");
        } finally {
            saving = false;
        }
    }

    onMount(loadSettings);
</script>

<div class="h-full overflow-y-auto pr-2 custom-scrollbar">
    {#if loading}
        <div class="flex flex-col items-center justify-center py-20 gap-4">
            <div
                class="w-8 h-8 rounded-full border-2 border-indigo-500 border-t-transparent animate-spin"
            ></div>
            <div class="text-gray-500 font-medium">Loading settings...</div>
        </div>
    {:else}
        <div class="max-w-6xl mx-auto space-y-6 pb-10">
            <div class="flex justify-between items-center">
                <div>
                    <h2 class="text-2xl font-bold text-white tracking-tight">
                        Server Settings
                    </h2>
                    <p class="text-gray-400 text-sm mt-1">
                        Configure your instance environment and startup options.
                    </p>
                </div>
                <button
                    on:click={saveSettings}
                    disabled={saving}
                    class="bg-indigo-600 hover:bg-indigo-500 focus:ring-4 focus:ring-indigo-500/20 text-white px-6 py-2.5 rounded-xl font-medium transition-all shadow-lg hover:shadow-indigo-500/25 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                >
                    {#if saving}
                        <div
                            class="w-4 h-4 rounded-full border-2 border-white/50 border-t-white animate-spin"
                        ></div>
                        <span>Saving...</span>
                    {:else}
                        <svg
                            class="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                            ><path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4"
                            /></svg
                        >
                        <span>Save Changes</span>
                    {/if}
                </button>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Startup Configuration -->
                <div
                    class="bg-gray-900/60 backdrop-blur-xl border border-white/5 rounded-2xl p-6 shadow-xl flex flex-col gap-6"
                >
                    <div
                        class="flex items-center gap-3 border-b border-white/5 pb-4"
                    >
                        <div
                            class="w-10 h-10 rounded-full bg-emerald-500/10 flex items-center justify-center text-emerald-400"
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
                                    d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                                /><path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                                /></svg
                            >
                        </div>
                        <div>
                            <h3 class="text-lg font-bold text-white">
                                Startup Configuration
                            </h3>
                            <p class="text-xs text-gray-500">
                                Manage how the server launches
                            </p>
                        </div>
                    </div>

                    <div class="space-y-4">
                        <div class="space-y-2">
                            <label
                                for="jar"
                                class="block text-sm font-medium text-gray-300"
                                >Target JAR File</label
                            >
                            <div class="relative">
                                <select
                                    id="jar"
                                    bind:value={jarFile}
                                    class="w-full bg-black/30 border border-white/10 rounded-xl px-4 py-3 text-white focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500/50 focus:outline-none transition-all appearance-none"
                                >
                                    {#each jarFiles as jar}
                                        <option value={jar.name}
                                            >{jar.name}</option
                                        >
                                    {/each}
                                </select>
                                <div
                                    class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-gray-500"
                                >
                                    <svg
                                        class="w-4 h-4"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                        ><path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M19 9l-7 7-7-7"
                                        /></svg
                                    >
                                </div>
                            </div>
                            <p class="text-xs text-gray-500">
                                The specific .jar file that initiates the server
                                process.
                            </p>
                        </div>
                    </div>
                </div>

                <!-- Java Environment -->
                <div
                    class="bg-gray-900/60 backdrop-blur-xl border border-white/5 rounded-2xl p-6 shadow-xl flex flex-col gap-6 row-span-2"
                >
                    <div
                        class="flex items-center gap-3 border-b border-white/5 pb-4"
                    >
                        <div
                            class="w-10 h-10 rounded-full bg-orange-500/10 flex items-center justify-center text-orange-400"
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
                                    d="M13 10V3L4 14h7v7l9-11h-7z"
                                /></svg
                            >
                        </div>
                        <div>
                            <h3 class="text-lg font-bold text-white">
                                Java Environment
                            </h3>
                            <p class="text-xs text-gray-500">
                                Runtime and memory allocation
                            </p>
                        </div>
                    </div>

                    <div class="space-y-6">
                        <div class="space-y-2">
                            <label
                                for="javapath"
                                class="block text-sm font-medium text-gray-300"
                                >Java Runtime</label
                            >
                            <div class="relative">
                                <select
                                    bind:value={selectedJavaMode}
                                    class="w-full bg-black/30 border border-white/10 rounded-xl px-4 py-3 text-white focus:ring-2 focus:ring-orange-500/50 focus:border-orange-500/50 focus:outline-none transition-all appearance-none"
                                >
                                    <option value="">System Default</option>
                                    {#each installedRuntimes as runtime}
                                        <option value={runtime.path}
                                            >Java {runtime.version} ({runtime.name})</option
                                        >
                                    {/each}
                                    <option value="custom"
                                        >Custom Path...</option
                                    >
                                </select>
                                <div
                                    class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-gray-500"
                                >
                                    <svg
                                        class="w-4 h-4"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                        ><path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M19 9l-7 7-7-7"
                                        /></svg
                                    >
                                </div>
                            </div>

                            {#if selectedJavaMode === "custom"}
                                <div class="mt-2">
                                    <input
                                        id="javapath"
                                        type="text"
                                        bind:value={javaPath}
                                        placeholder="/usr/bin/java"
                                        class="w-full bg-black/30 border border-white/10 rounded-xl px-4 py-2 text-white font-mono text-sm focus:ring-2 focus:ring-orange-500/50 focus:border-orange-500/50 focus:outline-none transition-all"
                                    />
                                </div>
                            {/if}
                        </div>

                        <div class="space-y-2">
                            <label
                                for="memory"
                                class="block text-sm font-medium text-gray-300"
                                >Max Memory Allocation</label
                            >
                            <div class="relative">
                                <input
                                    id="memory"
                                    type="number"
                                    bind:value={maxMemory}
                                    class="w-full bg-black/30 border border-white/10 rounded-xl px-4 py-3 text-white focus:ring-2 focus:ring-orange-500/50 focus:border-orange-500/50 focus:outline-none transition-all font-mono"
                                />
                                <div
                                    class="absolute right-4 top-1/2 -translate-y-1/2 text-gray-500 text-sm font-medium"
                                >
                                    MB
                                </div>
                            </div>
                            <div
                                class="flex justify-between text-xs text-gray-500 px-1"
                            >
                                <span>1024 MB = 1 GB</span>
                                <span>Recommended: > 2048 MB</span>
                            </div>
                        </div>

                        <div class="space-y-2">
                            <label
                                for="args"
                                class="block text-sm font-medium text-gray-300"
                                >JVM Flags</label
                            >
                            <textarea
                                id="args"
                                bind:value={javaArgs}
                                rows="3"
                                placeholder="-XX:+UseG1GC -Dfile.encoding=UTF-8"
                                class="w-full bg-black/30 border border-white/10 rounded-xl px-4 py-3 text-white font-mono text-sm focus:ring-2 focus:ring-orange-500/50 focus:border-orange-500/50 focus:outline-none transition-all"
                            ></textarea>
                            <p class="text-xs text-gray-500">
                                Advanced startup flags for the JVM.
                            </p>
                        </div>
                    </div>
                </div>

                <!-- Integrations -->
                <div
                    class="bg-gray-900/60 backdrop-blur-xl border border-white/5 rounded-2xl p-6 shadow-xl flex flex-col gap-6"
                >
                    <div
                        class="flex items-center gap-3 border-b border-white/5 pb-4"
                    >
                        <div
                            class="w-10 h-10 rounded-full bg-blue-500/10 flex items-center justify-center text-blue-400"
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
                                    d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                                /></svg
                            >
                        </div>
                        <div>
                            <h3 class="text-lg font-bold text-white">
                                Organization
                            </h3>
                            <p class="text-xs text-gray-500">
                                Group and categorize this instance
                            </p>
                        </div>
                    </div>
                    <div class="space-y-4">
                        <div class="space-y-2">
                            <label
                                for="group"
                                class="block text-sm font-medium text-gray-300"
                                >Folder / Group</label
                            >
                            <input
                                id="group"
                                type="text"
                                bind:value={group}
                                placeholder="e.g. Lobby Servers"
                                class="w-full bg-black/30 border border-white/10 rounded-xl px-4 py-3 text-white font-mono text-sm focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500/50 focus:outline-none transition-all"
                            />
                            <p class="text-xs text-gray-500">
                                Instances with the same folder name will be
                                grouped together.
                            </p>
                        </div>
                    </div>
                </div>

                <!-- Integrations -->
                <div
                    class="bg-gray-900/60 backdrop-blur-xl border border-white/5 rounded-2xl p-6 shadow-xl flex flex-col gap-6"
                >
                    <div
                        class="flex items-center gap-3 border-b border-white/5 pb-4"
                    >
                        <div
                            class="w-10 h-10 rounded-full bg-indigo-500/10 flex items-center justify-center text-indigo-400"
                        >
                            <svg
                                class="w-5 h-5"
                                fill="currentColor"
                                viewBox="0 0 24 24"
                                ><path
                                    d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037 13.434 13.434 0 0 0-.58 1.192 18.281 18.281 0 0 0-5.546 0 13.435 13.435 0 0 0-.58-1.192.074.074 0 0 0-.079-.037 19.796 19.796 0 0 0-4.884 1.515.064.064 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.897 19.897 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.093 14.093 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z"
                                /></svg
                            >
                        </div>
                        <div>
                            <h3 class="text-lg font-bold text-white">
                                Integrations
                            </h3>
                            <p class="text-xs text-gray-500">
                                Connect with external services
                            </p>
                        </div>
                    </div>

                    <div class="space-y-4">
                        <div class="space-y-2">
                            <label
                                for="webhook"
                                class="block text-sm font-medium text-gray-300"
                                >Discord Webhook URL</label
                            >
                            <input
                                id="webhook"
                                type="text"
                                bind:value={webhookUrl}
                                placeholder="https://discord.com/api/webhooks/..."
                                class="w-full bg-black/30 border border-white/10 rounded-xl px-4 py-3 text-white font-mono text-sm focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500/50 focus:outline-none transition-all"
                            />
                            <p class="text-xs text-gray-500">
                                Send notifications for server events (Start,
                                Stop, Crash).
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    {/if}
</div>
