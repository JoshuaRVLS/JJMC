<script>
    import { goto } from "$app/navigation";
    import { onMount } from "svelte";

    let name = "";
    let isCreating = false;
    /** @type {string[]} */
    let versions = [];
    let selectedVersion = "1.21";

    onMount(async () => {
        const res = await fetch("/api/versions/game");
        if (res.ok) {
            const data = await res.json();
            versions = data.map((/** @type {any} */ v) => v.version);

            if (versions.length > 0) {
                // Find if '1.21' exists, otherwise default to the first version
                const defaultVersion = versions.find((v) => v === "1.21");
                selectedVersion = defaultVersion || versions[0];
            }
        }
    });

    async function createNetwork() {
        if (!name) return;
        isCreating = true;

        const res = await fetch("/api/network/create", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                name,
                proxyType: "velocity",
                backendType: "paper",
                backendVersion: selectedVersion,
            }),
        });

        if (res.ok) {
            alert(
                "Network creation started! It may take a few minutes for the servers to appear.",
            );
            goto("/instances");
        } else {
            const data = await res.json();
            alert("Error: " + data.error);
            isCreating = false;
        }
    }
</script>

<div class="max-w-2xl mx-auto p-6">
    <h1 class="text-3xl font-bold mb-6 text-white">Create Network</h1>

    <div class="bg-gray-800/50 p-6 rounded-xl border border-white/10 space-y-6">
        <div>
            <label class="block text-sm font-medium text-gray-400 mb-2">
                Network Name
                <input
                    bind:value={name}
                    type="text"
                    class="w-full bg-gray-900/50 border border-white/10 rounded-lg p-3 text-white focus:border-blue-500 outline-none mt-2"
                    placeholder="My Network"
                />
            </label>
        </div>

        <div class="grid grid-cols-2 gap-4">
            <div class="p-4 bg-gray-900/30 rounded-lg border border-white/5">
                <div class="font-medium text-white mb-1">Proxy Type</div>
                <div class="text-sm text-gray-500">Velocity (Latest)</div>
            </div>
            <div
                class="p-4 bg-gray-900/30 rounded-lg border border-white/5 space-y-2"
            >
                <div class="font-medium text-white mb-1">Backend Type</div>
                <div class="text-sm text-gray-500 mb-2">
                    Paper + Velocity Support
                </div>
                <select
                    bind:value={selectedVersion}
                    class="w-full bg-gray-800/50 border border-white/10 rounded px-2 py-1 text-xs text-white outline-none"
                >
                    {#each versions as v}
                        <option value={v}>{v}</option>
                    {/each}
                </select>
            </div>
        </div>

        <div class="text-sm text-blue-400/80 bg-blue-500/10 p-4 rounded-lg">
            This will create 3 servers:
            <ul class="list-disc pl-5 mt-2 space-y-1">
                <li>
                    <strong>{name || "My Network"} Proxy</strong> (Velocity)
                </li>
                <li><strong>{name || "My Network"} Lobby</strong> (Paper)</li>
                <li>
                    <strong>{name || "My Network"} Survival</strong> (Paper)
                </li>
            </ul>
        </div>

        <button
            on:click={createNetwork}
            disabled={isCreating || !name}
            class="w-full bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium py-3 rounded-lg transition-colors"
        >
            {isCreating ? "Creating Network..." : "Create Network"}
        </button>
    </div>
</div>
