<script>
    import { createEventDispatcher } from "svelte";
    import { X, Server } from "lucide-svelte";
    import { fade, scale } from "svelte/transition";
    import { addToast } from "$lib/stores/toast";
    import { goto } from "$app/navigation";
    import { createId } from "@paralleldrive/cuid2";

    export let open = false;
    export let modpack = null;

    let name = "";
    let loading = false;
    let error = "";

    $: if (open && modpack) {
        name = modpack.title;
    }

    async function install() {
        if (!name) return;

        loading = true;
        error = "";

        try {
            // 1. Create Instance
            const id = createId();

            // Assume fabric for now as most modpacks are fabric/quilt or forge.
            // Ideally we'd detect this from modpack metadata but for now we'll default to 'fabric'
            // or we could ask the user.
            // Actually, we can just create a 'custom' instance initially or try to infer.
            // But wait, the install endpoint for modpack `inst.InstallModpack` handles downloading.
            // We need to create a base instance first.

            // For simplicity, we'll create a "fabric" instance (safe bet for many modern packs)
            // or maybe "custom" and let the modpack installer handle the jar?
            // The backend `InstallModpack` implementation will need to be intelligent.
            // Let's create a "fabric" instance for now as a default container.

            // Wait, the plan says: "Install button on a modpack will import it as a new server."
            // We need to create an instance first.

            // Let's create as "fabric" (most common) and empty version,
            // the modpack installer will likely overwrite/setup jars.

            // NOTE: Current backend implementation of InstallModpack just runs `inst.InstallModpack(payload.ProjectID)`.
            // We need to check what `inst.InstallModpack` does. It's likely using `packwiz` or `mrpack-install`.

            const createRes = await fetch("/api/instances", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    id,
                    name,
                    type: "custom", // Avoid triggering template install with empty version
                    version: "", // Let installer handle it
                }),
            });

            if (!createRes.ok) throw await createRes.text();

            // 2. Install Modpack
            const installRes = await fetch(`/api/instances/${id}/modpacks`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    projectId: modpack.project_id,
                }),
            });

            if (!installRes.ok) throw await installRes.text();

            addToast("Modpack installation started", "success");
            open = false;
            goto(`/instances/${id}`);
        } catch (e) {
            console.error(e);
            error = e.toString();
        } finally {
            loading = false;
        }
    }
</script>

{#if open}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
        transition:fade
    >
        <div
            class="bg-gray-900 border border-white/10 rounded-2xl w-full max-w-md shadow-2xl overflow-hidden"
            transition:scale
        >
            <div
                class="p-6 border-b border-white/5 flex justify-between items-center"
            >
                <h3 class="text-lg font-bold text-white">Install Modpack</h3>
                <button
                    on:click={() => (open = false)}
                    class="text-gray-400 hover:text-white transition-colors"
                >
                    <X class="w-5 h-5" />
                </button>
            </div>

            <div class="p-6">
                {#if modpack}
                    <div class="flex items-center gap-4 mb-6">
                        {#if modpack.icon_url}
                            <img
                                src={modpack.icon_url}
                                alt={modpack.title}
                                class="w-16 h-16 rounded-xl object-cover bg-gray-800"
                            />
                        {:else}
                            <div
                                class="w-16 h-16 rounded-xl bg-gray-800 flex items-center justify-center text-2xl font-bold text-indigo-500"
                            >
                                {modpack.title.charAt(0)}
                            </div>
                        {/if}
                        <div>
                            <h4 class="font-bold text-white">
                                {modpack.title}
                            </h4>
                            <p class="text-sm text-gray-400">
                                by {modpack.author}
                            </p>
                        </div>
                    </div>
                {/if}

                <div class="mb-4">
                    <label
                        for="server-name"
                        class="block text-sm font-medium text-gray-400 mb-2"
                        >Server Name</label
                    >
                    <input
                        id="server-name"
                        type="text"
                        bind:value={name}
                        class="w-full bg-gray-950 border border-gray-800 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:border-indigo-500/50 focus:ring-1 focus:ring-indigo-500/50"
                    />
                </div>

                {#if error}
                    <div
                        class="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm mb-4"
                    >
                        {error}
                    </div>
                {/if}
            </div>

            <div
                class="p-6 border-t border-white/5 flex justify-end gap-3 bg-white/5"
            >
                <button
                    on:click={() => (open = false)}
                    class="px-4 py-2 rounded-xl text-sm font-medium text-gray-400 hover:text-white hover:bg-white/5 transition-colors"
                >
                    Cancel
                </button>
                <button
                    on:click={install}
                    disabled={loading || !name}
                    class="px-4 py-2 rounded-xl text-sm font-bold bg-indigo-600 hover:bg-indigo-500 text-white shadow-lg shadow-indigo-500/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2"
                >
                    {#if loading}
                        Installing...
                    {:else}
                        Create Server
                    {/if}
                </button>
            </div>
        </div>
    </div>
{/if}
