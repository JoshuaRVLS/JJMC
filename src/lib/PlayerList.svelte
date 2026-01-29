<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";

    /** @type {string} */
    export let instanceId;
    export let type = "whitelist"; // 'whitelist' or 'ops'

    /** @type {Array<{uuid: string, name: string}>} */
    let players = [];
    let loading = true;
    let newPlayerName = "";

    $: fileName = type === "ops" ? "ops.json" : "whitelist.json";
    $: title = type === "ops" ? "Operators" : "Whitelist";

    async function loadPlayers() {
        loading = true;
        try {
            const res = await fetch(
                `/api/instances/${instanceId}/files/content?path=${fileName}`,
            );
            if (res.ok) {
                const text = await res.text();
                // If empty or new file, it might be empty string
                if (text.trim() === "") {
                    players = [];
                } else {
                    players = JSON.parse(text);
                }
            } else {
                players = [];
            }
        } catch (e) {
            console.error(e);
            addToast(`Error loading ${title}`, "error");
        } finally {
            loading = false;
        }
    }

    async function savePlayers() {
        try {
            const res = await fetch(
                `/api/instances/${instanceId}/files/content`,
                {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        path: fileName,
                        content: JSON.stringify(players, null, 2),
                    }),
                },
            );

            if (res.ok) {
                addToast(`${title} saved`, "success");
            } else {
                addToast("Failed to save", "error");
            }
        } catch (e) {
            addToast("Error saving", "error");
        }
    }

    async function checkOnlineMode() {
        try {
            const res = await fetch(
                `/api/instances/${instanceId}/files/content?path=server.properties`,
            );
            if (res.ok) {
                const text = await res.text();
                // Simple parsing for online-mode=false
                // Default is true if not found, but usually it's there
                const match = text.match(/online-mode\s*=\s*(false|true)/);
                if (match && match[1] === "false") {
                    return false;
                }
            }
        } catch (e) {
            console.error("Failed to check online mode", e);
        }
        return true;
    }

    async function addPlayer() {
        if (!newPlayerName.trim()) return;

        try {
            const isOnline = await checkOnlineMode();
            const res = await fetch(
                `/api/system/uuid?name=${newPlayerName}&offline=${!isOnline}`,
            );
            if (!res.ok) throw new Error("User not found or API error");
            const data = await res.json();

            const newEntry = {
                uuid: data.uuid,
                name: data.username,
                ...(type === "ops"
                    ? { level: 4, bypassesPlayerLimit: false }
                    : {}),
            };

            // Check duplicate
            if (players.find((p) => p.uuid === newEntry.uuid)) {
                addToast("Player already in list", "warning");
                return;
            }

            players = [...players, newEntry];
            newPlayerName = "";
            savePlayers();
        } catch (e) {
            console.error(e);
            addToast("Error: " + e.message + ". Check Online Mode?", "error");
        }
    }

    /** @param {string} uuid */
    function removePlayer(uuid) {
        players = players.filter((p) => p.uuid !== uuid);
        savePlayers();
    }

    onMount(() => {
        loadPlayers();
    });
</script>

<div
    class="h-full flex flex-col bg-gray-900/50 rounded-xl overflow-hidden border border-white/5"
>
    <!-- Toolbar -->
    <div
        class="flex items-center justify-between px-4 py-3 bg-white/5 border-b border-white/5"
    >
        <div class="text-sm font-bold text-gray-300">{title} Manager</div>
        <div class="flex gap-2">
            <input
                type="text"
                placeholder="Username"
                bind:value={newPlayerName}
                on:keydown={(e) => e.key === "Enter" && addPlayer()}
                class="bg-black/20 border border-white/10 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-indigo-500/50"
            />
            <button
                on:click={addPlayer}
                class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-lg transition-colors"
            >
                Add
            </button>
        </div>
    </div>

    <!-- List -->
    <div class="flex-1 overflow-y-auto p-4">
        {#if loading}
            <div class="flex items-center justify-center h-full text-gray-500">
                Loading...
            </div>
        {:else if players.length === 0}
            <div
                class="flex flex-col items-center justify-center h-full text-gray-500 gap-2"
            >
                <div>No players in {type}.</div>
            </div>
        {:else}
            <div class="flex flex-col gap-2">
                {#each players as player}
                    <div
                        class="flex items-center justify-between bg-white/5 p-3 rounded-lg border border-white/5 hover:border-white/10 transition-colors"
                    >
                        <div class="flex items-center gap-3">
                            <img
                                src={`https://minotar.net/helm/${player.name}/24.png`}
                                alt={player.name}
                                class="w-6 h-6 rounded"
                            />
                            <div class="flex flex-col">
                                <span class="text-sm font-bold text-gray-200"
                                    >{player.name}</span
                                >
                                <span
                                    class="text-[10px] font-mono text-gray-500"
                                    >{player.uuid}</span
                                >
                            </div>
                        </div>
                        <button
                            on:click={() => removePlayer(player.uuid)}
                            class="text-red-400 hover:text-red-300 p-1 hover:bg-red-500/10 rounded transition-colors"
                            title="Remove"
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
                                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                                /></svg
                            >
                        </button>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>
