<script>
    import { onMount, onDestroy } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import { Globe, Terminal, Loader2, Key, Info } from "lucide-svelte";

    export let instanceId = "";

    let loading = false;
    let config = {
        provider: "playit", // 'playit' or 'ngrok'
        token: "",
    };

    let status = {
        running: false,
        provider: "",
        public_address: "",
        log: "",
        config: { provider: "", token: "" },
    };

    /** @type {ReturnType<typeof setInterval> | undefined} */
    let pollInterval;

    async function loadStatus() {
        try {
            const res = await fetch(`/api/instances/${instanceId}/tunnel`);
            if (res.ok) {
                status = await res.json();

                // Load saved config if we have it
                if (status.config) {
                    if (status.config.provider) {
                        // Only override if not already running (or to sync state)
                        if (!config.provider || !status.running) {
                            config.provider = status.config.provider;
                        }
                    }
                    if (status.config.token) {
                        config.token = status.config.token;
                    }
                }
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function toggleTunnel() {
        if (loading) return;
        loading = true;

        try {
            if (status.running) {
                // Stop
                const res = await fetch(
                    `/api/instances/${instanceId}/tunnel/stop`,
                    { method: "POST" },
                );
                if (!res.ok) throw await res.text();
                addToast("Tunnel stopped", "success");
            } else {
                // Start
                if (!config.token && config.provider === "playit") {
                    // Ngrok might not need token if configured globally, but playit usually does via CLI
                    return addToast("Please enter a token/secret", "error");
                }

                const res = await fetch(
                    `/api/instances/${instanceId}/tunnel/start`,
                    {
                        method: "POST",
                        headers: { "Content-Type": "application/json" },
                        body: JSON.stringify(config),
                    },
                );
                if (!res.ok) throw await res.text();
                addToast("Tunnel started", "success");
            }
            await loadStatus();
        } catch (e) {
            addToast(`Error: ${e}`, "error");
        } finally {
            loading = false;
        }
    }

    onMount(() => {
        loadStatus();
        pollInterval = setInterval(loadStatus, 2000);
    });

    onDestroy(() => {
        if (pollInterval) clearInterval(pollInterval);
    });
</script>

<div class="h-full flex flex-col gap-6 p-6">
    <div
        class="bg-gray-900/60 backdrop-blur-xl border border-white/5 rounded-2xl p-6 shadow-xl"
    >
        <div class="flex items-start gap-4 mb-8">
            <div
                class="w-12 h-12 rounded-xl bg-cyan-500/10 flex items-center justify-center text-cyan-400"
            >
                <Globe class="w-6 h-6" />
            </div>
            <div>
                <h2 class="text-xl font-bold text-white mb-1">
                    Port Forwarding
                </h2>
                <p class="text-gray-400 text-sm">
                    Expose your server to the internet using tunneling services.
                </p>
                <div class="flex gap-2 mt-2">
                    <span
                        class="px-2 py-0.5 rounded text-[10px] font-bold border border-emerald-500/20 text-emerald-400 bg-emerald-500/5"
                    >
                        MANAGED AUTOMATICALLY
                    </span>
                </div>
            </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
            <!-- Configuration -->
            <div class="space-y-6">
                <div>
                    <label class="block text-sm font-medium text-gray-400 mb-2"
                        >Provider</label
                    >
                    <div class="grid grid-cols-2 gap-3">
                        <button
                            class="px-4 py-3 rounded-xl border text-sm font-bold transition-all
                            {config.provider === 'playit'
                                ? 'bg-cyan-500/20 border-cyan-500/50 text-cyan-400'
                                : 'bg-gray-800/50 border-gray-700 text-gray-400 hover:bg-gray-800'}"
                            disabled={status.running}
                            on:click={() => (config.provider = "playit")}
                        >
                            Playit.gg
                        </button>
                        <button
                            class="px-4 py-3 rounded-xl border text-sm font-bold transition-all
                            {config.provider === 'ngrok'
                                ? 'bg-indigo-500/20 border-indigo-500/50 text-indigo-400'
                                : 'bg-gray-800/50 border-gray-700 text-gray-400 hover:bg-gray-800'}"
                            disabled={status.running}
                            on:click={() => (config.provider = "ngrok")}
                        >
                            Ngrok
                        </button>
                    </div>
                </div>

                <div>
                    <label class="block text-sm font-medium text-gray-400 mb-2">
                        {config.provider === "playit"
                            ? "Secret Key (Requires 'playit' binary)"
                            : "Auth Token (Optional)"}
                    </label>
                    <div class="relative">
                        <div
                            class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-gray-500"
                        >
                            <Key class="w-5 h-5" />
                        </div>
                        <input
                            type="password"
                            bind:value={config.token}
                            disabled={status.running}
                            class="w-full bg-gray-950 border border-gray-800 rounded-xl py-2.5 pl-10 pr-4 text-white focus:outline-none focus:border-cyan-500/50 focus:ring-1 focus:ring-cyan-500/50 transition-colors placeholder-gray-700"
                            placeholder={config.provider === "playit"
                                ? "Enter your playit agent secret..."
                                : "Enter ngrok authtoken (auto-installed if missing)..."}
                        />
                    </div>
                    <p class="text-xs text-gray-500 mt-2 flex gap-1">
                        <Info class="w-3 h-3 translate-y-px" />
                        {config.provider === "playit"
                            ? "Get your secret from the playit.gg dashboard."
                            : "We will automatically install ngrok if it's not found."}
                    </p>
                </div>

                <div class="pt-4">
                    <button
                        on:click={toggleTunnel}
                        disabled={loading}
                        class="w-full py-3 rounded-xl font-bold flex items-center justify-center gap-2 transition-all shadow-lg
                        {status.running
                            ? 'bg-red-500/10 text-red-400 border border-red-500/50 hover:bg-red-500/20 shadow-red-900/20'
                            : 'bg-cyan-600 text-white hover:bg-cyan-500 shadow-cyan-900/20'}"
                    >
                        {#if loading}
                            <Loader2 class="w-5 h-5 animate-spin" />
                        {:else if status.running}
                            Stop Tunnel
                        {:else}
                            Start Tunnel
                        {/if}
                    </button>
                </div>
            </div>

            <!-- Status -->
            <div
                class="bg-black/40 rounded-xl border border-white/5 p-4 flex flex-col min-h-[300px]"
            >
                <div
                    class="flex items-center justify-between mb-4 pb-4 border-b border-white/5"
                >
                    <h3 class="font-bold text-gray-300">Live Status</h3>
                    <div class="flex items-center gap-2">
                        <div
                            class="w-2 h-2 rounded-full {status.running
                                ? 'bg-emerald-500 animate-pulse'
                                : 'bg-gray-600'}"
                        ></div>
                        <span class="text-xs font-mono uppercase text-gray-500">
                            {status.running ? "Active" : "Inactive"}
                        </span>
                    </div>
                </div>

                {#if status.running}
                    <div class="mb-4">
                        <div
                            class="text-xs text-gray-500 uppercase font-bold mb-1"
                        >
                            Public Address
                        </div>
                        <div
                            class="bg-gray-900 rounded-lg p-3 text-emerald-400 font-mono text-sm break-all border border-emerald-500/20"
                        >
                            {status.public_address || "Waiting for address..."}
                        </div>
                    </div>
                {/if}

                <div class="flex-1 flex flex-col min-h-0">
                    <div class="text-xs text-gray-500 uppercase font-bold mb-1">
                        Logs
                    </div>
                    <div
                        class="flex-1 bg-gray-950 rounded-lg p-3 text-xs font-mono text-gray-400 overflow-y-auto custom-scrollbar"
                    >
                        <pre class="whitespace-pre-wrap">{status.log ||
                                "No logs available."}</pre>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

<style>
    /* Custom Scrollbar */
    .custom-scrollbar::-webkit-scrollbar {
        width: 6px;
    }
    .custom-scrollbar::-webkit-scrollbar-track {
        background: rgba(0, 0, 0, 0.2);
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
