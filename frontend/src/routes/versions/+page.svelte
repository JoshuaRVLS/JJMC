<script>
    /**
     * @typedef {Object} Version
     * @property {string} name
     * @property {string} type
     * @property {string} version
     * @property {boolean} installed
     */

    /** @type {Version[]} */
    let versions = [
        {
            name: "Fabric 1.21.1",
            type: "fabric",
            version: "1.21.1",
            installed: false,
        },
    ];
    let activeVersion = "None";
    let isInstalling = false;

    /**
     * @param {Version} v
     */
    async function install(v) {
        if (isInstalling) return;
        isInstalling = true;
        try {
            const res = await fetch("/api/versions/install", {
                method: "POST",
                body: JSON.stringify(v),
            });
            if (res.ok) {
                alert("Installed " + v.name);
                v.installed = true;
                versions = [...versions];
            } else {
                alert("Install failed");
            }
        } catch (e) {
            /** @type {Error} */
            const err = /** @type {Error} */ (e);
            alert("Error: " + err.message);
        } finally {
            isInstalling = false;
        }
    }

    /**
     * @param {Version} v
     */
    function select(v) {
        activeVersion = v.name;
    }
</script>

<div class="h-full flex flex-col">
    <header
        class="flex justify-between items-center px-4 py-3 border-b border-black bg-white"
    >
        <h2 class="font-bold uppercase tracking-widest">Version Manager</h2>
        <div class="text-xs">
            ACTIVE: <span class="font-bold px-2 py-0.5 bg-black text-white"
                >{activeVersion}</span
            >
        </div>
    </header>

    <div class="p-6">
        <div class="border border-black">
            <div
                class="bg-black text-white px-4 py-2 font-bold uppercase text-xs flex justify-between"
            >
                <span>Available Versions</span>
                <span>Action</span>
            </div>

            {#each versions as v}
                <div
                    class="flex justify-between items-center px-4 py-3 border-t border-black hover:bg-gray-50 transition-colors"
                >
                    <div>
                        <div class="font-bold uppercase">{v.name}</div>
                        <div class="text-xs text-gray-500">
                            Type: {v.type} | Ver: {v.version}
                        </div>
                    </div>
                    <div class="flex gap-2">
                        {#if !v.installed}
                            <button
                                on:click={() => install(v)}
                                disabled={isInstalling}
                                class="border border-black px-3 py-1 text-xs uppercase hover:bg-black hover:text-white transition-colors disabled:opacity-50"
                            >
                                {isInstalling ? "Busy..." : "Install"}
                            </button>
                        {:else}
                            <button
                                on:click={() => select(v)}
                                class="bg-black text-white px-3 py-1 text-xs uppercase hover:opacity-80 transition-opacity"
                            >
                                Select
                            </button>
                        {/if}
                    </div>
                </div>
            {/each}
        </div>
    </div>
</div>
