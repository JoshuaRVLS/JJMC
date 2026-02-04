<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import { askConfirm } from "$lib/stores/confirm";
    import ModSearch from "./mods/ModSearch.svelte";
    import Modlist from "./mods/Modlist.svelte";

    /** @type {string} */
    export let instanceId;
    export let type = "";
    export let mode = "mod";

    let activeTab = mode === "plugin" ? "plugin" : "mod";

    $: if (mode) {
        if (mode === "plugin" && activeTab !== "plugin") {
            activeTab = "plugin";
        } else if (
            mode === "mod" &&
            activeTab !== "mod" &&
            activeTab !== "modpack"
        ) {
            activeTab = "mod";
        }
    }

    let query = "";

    /** @type {any[]} */
    let results = [];

    let installedIds = new Set();
    let loading = false;
    let loadingMore = false;

    /** @type {string | null} */
    let installingId = null;
    let offset = 0;
    let hasMore = true;
    let sortBy = "relevance";

    /** @type {string | null} */
    let viewingVersionsId = null;

    /** @type {any[]} */
    let versionsList = [];
    let loadingVersions = false;

    /** @type {IntersectionObserver} */
    let observer;

    /** @type {HTMLElement} */
    let sentinel;

    async function fetchInstalled() {
        try {
            const res = await fetch(`/api/instances/${instanceId}/mods`);
            if (res.ok) {
                const ids = await res.json();
                installedIds = new Set(ids);
            }
        } catch (e) {
            console.error("Failed to fetch installed mods:", e);
        }
    }

    /** @param {string} projectId */
    async function fetchVersions(projectId) {
        loadingVersions = true;
        versionsList = [];
        try {
            let typeParam = activeTab === "plugin" ? "plugin" : "mod";

            const res = await fetch(
                `/api/instances/${instanceId}/mods/${projectId}/versions?type=${typeParam}`,
            );
            if (res.ok) {
                versionsList = await res.json();
            } else {
                addToast("Failed to load versions", "error");
            }
        } catch (e) {
            const err = /** @type {Error} */ (e);
            addToast("Error loading versions: " + err.message, "error");
        } finally {
            loadingVersions = false;
        }
    }

    async function search(isNew = true) {
        if (isNew) {
            loading = true;
            offset = 0;
            results = [];
            hasMore = true;
        } else {
            loadingMore = true;
        }

        let typeParam = activeTab === "plugin" ? "plugin" : activeTab;
        if (typeParam === "mod" && mode === "plugin") typeParam = "plugin";

        try {
            const res = await fetch(
                `/api/instances/${instanceId}/mods/search?query=${encodeURIComponent(query)}&type=${typeParam}&offset=${offset}&sort=${sortBy}`,
            );
            if (res.ok) {
                const data = await res.json();
                if (isNew) {
                    results = data || [];
                } else {
                    results = [...results, ...(data || [])];
                }

                if (!data || data.length < 20) {
                    hasMore = false;
                }
            } else {
                const err = await res.json();
                addToast(
                    "Search failed: " + (err.error || "Unknown error"),
                    "error",
                );
            }
        } catch (e) {
            const msg = e instanceof Error ? e.message : String(e);
            addToast("Error searching: " + msg, "error");
        } finally {
            loading = false;
            loadingMore = false;
        }
    }

    function loadMore() {
        if (loading || loadingMore || !hasMore) return;
        offset += 20;
        search(false);
    }

    $: if (activeTab || sortBy) {
        search(true);

        viewingVersionsId = null;
    }

    onMount(() => {
        fetchInstalled();

        observer = new IntersectionObserver(
            (entries) => {
                if (entries[0].isIntersecting) {
                    loadMore();
                }
            },
            { threshold: 0.1 },
        );

        if (sentinel) observer.observe(sentinel);

        return () => {
            if (observer) observer.disconnect();
        };
    });

    /**
     * @param {string} projectId
     * @param {string} [versionId]
     */
    async function installMod(projectId, versionId = "") {
        installingId = projectId;
        try {
            let typeParam = activeTab === "plugin" ? "plugin" : "mod";

            const res = await fetch(`/api/instances/${instanceId}/mods`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    projectId,
                    resourceType: typeParam,
                    versionId,
                }),
            });
            if (res.ok) {
                addToast("Installed successfully", "success");
                fetchInstalled();
            } else {
                const err = await res.json();
                addToast(
                    "Install failed: " + (err.error || "Unknown error"),
                    "error",
                );
            }
        } catch (e) {
            const err = /** @type {Error} */ (e);
            addToast("Error installing: " + err.message, "error");
        } finally {
            installingId = null;
        }
    }

    /** @param {string} projectId */
    async function uninstallMod(projectId) {
        if (
            !(await askConfirm({
                title: "Uninstall Mod",
                message: "Are you sure you want to uninstall this mod?",
                dangerous: true,
                confirmText: "Uninstall",
            }))
        )
            return;
        installingId = projectId;
        try {
            let typeParam = activeTab === "plugin" ? "plugin" : "mod";

            const res = await fetch(`/api/instances/${instanceId}/mods`, {
                method: "DELETE",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    project_id: projectId,
                    resource_type: typeParam,
                }),
            });
            if (res.ok) {
                addToast("Uninstalled successfully", "success");
                fetchInstalled();
            } else {
                const err = await res.json();
                addToast(
                    "Uninstall failed: " + (err.error || "Unknown error"),
                    "error",
                );
            }
        } catch (e) {
            const err = /** @type {Error} */ (e);
            addToast("Error uninstalling: " + err.message, "error");
        } finally {
            installingId = null;
        }
    }

    /** @param {string} projectId */
    async function installModpack(projectId) {
        if (
            !(await askConfirm({
                title: "Install Modpack",
                message:
                    "Warning: Installing a modpack will DELETE all current mods in the 'mods' folder. Continue?",
                dangerous: true,
                confirmText: "Install Modpack",
            }))
        ) {
            return;
        }

        installingId = projectId;
        try {
            const res = await fetch(`/api/instances/${instanceId}/modpacks`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ projectId }),
            });
            if (res.ok) {
                addToast(
                    "Modpack installation started. Check console for progress.",
                    "success",
                );
            } else {
                const err = await res.json();
                addToast(
                    "Install failed: " + (err.error || "Unknown error"),
                    "error",
                );
            }
        } catch (e) {
            const err = /** @type {Error} */ (e);
            addToast("Error installing: " + err.message, "error");
        } finally {
            installingId = null;
        }
    }
</script>

<div class="h-full flex flex-col">
    <ModSearch
        {mode}
        bind:activeTab
        bind:sortBy
        bind:query
        {loading}
        {type}
        on:search={() => search(true)}
    />

    <Modlist
        {results}
        {loading}
        {loadingMore}
        {query}
        {hasMore}
        {installedIds}
        {installingId}
        {activeTab}
        bind:sentinel
        {versionsList}
        {loadingVersions}
        bind:viewingVersionsId
        on:viewVersions={(e) => {
            const item = e.detail;
            if (viewingVersionsId === item.project_id) {
                viewingVersionsId = null;
            } else {
                viewingVersionsId = item.project_id;
                fetchVersions(item.project_id);
            }
        }}
        on:closeVersions={() => (viewingVersionsId = null)}
        on:install={(e) => {
            const item = e.detail;
            if (activeTab === "modpack") {
                installModpack(item.project_id);
            } else {
                installMod(item.project_id);
            }
        }}
        on:installVersion={(e) =>
            installMod(e.detail.projectId, e.detail.versionId)}
        on:uninstall={(e) => uninstallMod(e.detail.project_id)}
    />
</div>
