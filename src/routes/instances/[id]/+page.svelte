<script>
    import { page } from "$app/stores";
    import { goto } from "$app/navigation";
    import { onMount, onDestroy } from "svelte";
    import Console from "$lib/Console.svelte";
    import Files from "$lib/Files.svelte";
    import ServerProperties from "$lib/ServerProperties.svelte";
    import PlayerList from "$lib/PlayerList.svelte";
    import Controls from "$lib/Controls.svelte";
    import Settings from "$lib/Settings.svelte";
    import Mods from "$lib/Mods.svelte";

    // Get instance ID from route params
    $: instanceId = $page.params.id;

    let status = "Offline";
    let interval;

    // Sync activeTab with URL
    $: activeTab = $page.url.searchParams.get("tab") || "console";

    function setTab(tab) {
        const url = new URL($page.url);
        url.searchParams.set("tab", tab);
        goto(url.toString(), {
            replaceState: true,
            noScroll: true,
            keepFocus: true,
        });
    }

    async function checkStatus() {
        if (!instanceId) return;
        try {
            const res = await fetch(`/api/instances/${instanceId}`);
            if (res.ok) {
                const data = await res.json();
                console.log("Status Poll:", data.status);
                status = data.status;
            }
        } catch (e) {
            console.error("Status check failed", e);
        }
    }

    onMount(() => {
        checkStatus();
        interval = setInterval(checkStatus, 2000);
    });

    onDestroy(() => {
        if (interval) clearInterval(interval);
    });
</script>

<div class="h-full flex flex-col p-6 gap-6">
    <!-- Header -->
    <header class="flex justify-between items-center">
        <!-- ... existing header content ... -->
        <div class="flex items-center gap-4">
            <a
                href="/instances"
                class="flex items-center justify-center w-10 h-10 rounded-xl bg-white/5 hover:bg-white/10 text-gray-400 hover:text-white transition-all border border-white/5"
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
                        d="M15 19l-7-7 7-7"
                    /></svg
                >
            </a>
            <div>
                <h2 class="text-xl font-bold text-white tracking-tight">
                    {instanceId}
                </h2>
                <!-- Status indicator -->
                <div
                    class="flex items-center gap-2 text-xs font-mono text-gray-400"
                >
                    <span class="relative flex h-2 w-2">
                        <span
                            class:animate-ping={status === "Starting" ||
                                status === "Stopping"}
                            class:bg-green-500={status === "Online"}
                            class:bg-red-500={status === "Offline" ||
                                status === "Error"}
                            class:bg-yellow-500={status === "Starting" ||
                                status === "Stopping"}
                            class="absolute inline-flex h-full w-full rounded-full opacity-75"
                        ></span>
                        <span
                            class:bg-green-500={status === "Online"}
                            class:bg-red-500={status === "Offline" ||
                                status === "Error"}
                            class:bg-yellow-500={status === "Starting" ||
                                status === "Stopping"}
                            class="relative inline-flex rounded-full h-2 w-2"
                        ></span>
                    </span>
                    {status}
                </div>
            </div>
        </div>

        <!-- Tabs -->
    </header>

    <!-- Main Grid -->
    <div class="flex-1 min-h-0 flex flex-col">
        {#if activeTab === "console"}
            <div class="flex-1 grid grid-rows-[1fr_auto] gap-6 min-h-0">
                <!-- Console Area -->
                <div class="min-h-0 relative">
                    <Console {instanceId} {status} />
                </div>

                <!-- Controls Area -->
                <div
                    class="relative z-10 bg-gray-900/60 backdrop-blur-xl border border-white/5 rounded-2xl p-6 flex justify-between items-center shadow-lg"
                >
                    <div class="flex items-center gap-3">
                        <div
                            class="w-10 h-10 rounded-full bg-indigo-500/10 flex items-center justify-center text-indigo-400"
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
                                    d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"
                                /></svg
                            >
                        </div>
                        <div>
                            <div class="text-sm font-bold text-white">
                                Power Controls
                            </div>
                            <div class="text-xs text-gray-500">
                                Manage server state
                            </div>
                        </div>
                    </div>
                    <div>
                        <Controls {instanceId} {status} />
                    </div>
                </div>
            </div>
        {:else if activeTab === "files"}
            <Files {instanceId} />
        {:else if activeTab === "properties"}
            <ServerProperties {instanceId} />
        {:else if activeTab === "whitelist"}
            <PlayerList {instanceId} type="whitelist" />
        {:else if activeTab === "ops"}
            <PlayerList {instanceId} type="ops" />
        {:else if activeTab === "settings"}
            <Settings {instanceId} />
        {:else if activeTab === "mods"}
            <Mods {instanceId} />
        {/if}
    </div>
</div>
