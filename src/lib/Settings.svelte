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

<div class="h-full overflow-y-auto pr-2">
    {#if loading}
        <div class="text-center text-gray-500 py-10">Loading settings...</div>
    {:else}
        <div
            class="bg-gray-900/60 backdrop-blur-xl border border-white/5 rounded-2xl p-6 shadow-lg max-w-2xl mx-auto space-y-6"
        >
            <h2 class="text-xl font-bold text-white mb-4">Java Settings</h2>

            <div class="space-y-2">
                <label
                    for="jar"
                    class="block text-sm font-medium text-gray-400"
                >
                    Target JAR
                </label>
                <select
                    id="jar"
                    bind:value={jarFile}
                    class="bg-black/20 border border-white/10 rounded-lg px-4 py-2 text-white w-full focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all"
                >
                    {#each jarFiles as jar}
                        <option value={jar.name}>{jar.name}</option>
                    {/each}
                </select>
                <div class="text-xs text-gray-500">
                    Select the server JAR file to execute.
                </div>
            </div>

            <div class="space-y-2">
                <label
                    for="webhook"
                    class="block text-sm font-medium text-gray-400"
                >
                    Discord Webhook URL
                </label>
                <input
                    id="webhook"
                    type="text"
                    bind:value={webhookUrl}
                    placeholder="https://discord.com/api/webhooks/..."
                    class="bg-black/20 border border-white/10 rounded-lg px-4 py-2 text-white w-full focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all font-mono text-sm"
                />
                <div class="text-xs text-gray-500">
                    Send server status notifications (Start, Stop, Crash) to a
                    Discord channel.
                </div>
            </div>

            <div class="space-y-2">
                <label
                    for="memory"
                    class="block text-sm font-medium text-gray-400"
                >
                    Max Memory (MB)
                </label>
                <div class="flex gap-4 items-center">
                    <input
                        id="memory"
                        type="number"
                        bind:value={maxMemory}
                        class="bg-black/20 border border-white/10 rounded-lg px-4 py-2 text-white w-full focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all no-spin"
                    />
                    <span class="text-gray-500 text-sm whitespace-nowrap">
                        Example: 2048 = 2GB
                    </span>
                </div>
                <div class="text-xs text-gray-500">
                    Allocate RAM for the Minecraft server. Ensure you leave
                    enough for the OS.
                </div>
            </div>

            <div class="space-y-2">
                <label
                    for="args"
                    class="block text-sm font-medium text-gray-400"
                >
                    Java Flags
                </label>
                <input
                    id="args"
                    type="text"
                    bind:value={javaArgs}
                    placeholder="-XX:+UseG1GC -D..."
                    class="bg-black/20 border border-white/10 rounded-lg px-4 py-2 text-white w-full focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all font-mono text-sm"
                />
                <div class="text-xs text-gray-500">
                    Additional arguments to pass to the startup command.
                </div>
            </div>

            <div class="space-y-2">
                <label
                    for="javapath"
                    class="block text-sm font-medium text-gray-400"
                >
                    Java Runtime
                </label>
                <div class="flex flex-col gap-2">
                    <select
                        bind:value={selectedJavaMode}
                        class="bg-black/20 border border-white/10 rounded-lg px-4 py-2 text-white w-full focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all"
                    >
                        <option value="">System Default</option>
                        {#each installedRuntimes as runtime}
                            <option value={runtime.path}
                                >Java {runtime.version} ({runtime.name})</option
                            >
                        {/each}
                        <option value="custom">Custom Path...</option>
                    </select>

                    {#if selectedJavaMode === "custom"}
                        <input
                            id="javapath"
                            type="text"
                            bind:value={javaPath}
                            placeholder="/path/to/java"
                            class="bg-black/20 border border-white/10 rounded-lg px-4 py-2 text-white w-full focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all font-mono text-sm"
                        />
                    {/if}
                </div>
                <div class="text-xs text-gray-500">
                    Select the Java version to use for this server.
                </div>
            </div>

            <div class="pt-4 flex justify-end">
                <button
                    on:click={saveSettings}
                    disabled={saving}
                    class="bg-indigo-600 hover:bg-indigo-500 text-white px-6 py-2 rounded-lg font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    {saving ? "Saving..." : "Save Changes"}
                </button>
            </div>
        </div>
    {/if}
</div>
