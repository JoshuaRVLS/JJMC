<script>
    import { onMount } from "svelte";
    import { addToast } from "$lib/stores/toast";
    import {
        Archive,
        RotateCcw,
        Trash2,
        Plus,
        Download,
        Loader2,
    } from "lucide-svelte";
    import ConfirmDialog from "$lib/components/ConfirmDialog.svelte";
    import { askConfirm } from "$lib/stores/confirm";

    export let instanceId;

    let backups = [];
    let loading = false;
    let creating = false;

    async function loadBackups() {
        loading = true;
        try {
            const res = await fetch(`/api/instances/${instanceId}/backups`);
            if (res.ok) {
                backups = (await res.json()) || [];
            } else {
                addToast("Failed to load backups", "error");
            }
        } catch (e) {
            console.error(e);
            addToast("Error loading backups", "error");
        } finally {
            loading = false;
        }
    }

    async function createBackup() {
        creating = true;
        try {
            const res = await fetch(`/api/instances/${instanceId}/backups`, {
                method: "POST",
            });
            if (res.ok) {
                addToast("Backup created successfully", "success");
                loadBackups();
            } else {
                const err = await res.json();
                addToast(err.error || "Failed to create backup", "error");
            }
        } catch (e) {
            addToast("Error creating backup", "error");
        } finally {
            creating = false;
        }
    }

    async function restoreBackup(filename) {
        if (
            await askConfirm({
                title: "Restore Backup?",
                message:
                    "This will overwrite your current server files. Make sure the server is offline. This action cannot be undone.",
                confirmText: "Restore",
                dangerous: true,
            })
        ) {
            try {
                const res = await fetch(
                    `/api/instances/${instanceId}/backups/${encodeURIComponent(filename)}/restore`,
                    { method: "POST" },
                );
                if (res.ok) {
                    addToast("Backup restored successfully", "success");
                } else {
                    const err = await res.json();
                    addToast(err.error || "Failed to restore backup", "error");
                }
            } catch (e) {
                addToast("Error restoring backup", "error");
            }
        }
    }

    async function deleteBackup(filename) {
        if (
            await askConfirm({
                title: "Delete Backup?",
                message: "Are you sure you want to delete this backup?",
                confirmText: "Delete",
                dangerous: true,
            })
        ) {
            try {
                const res = await fetch(
                    `/api/instances/${instanceId}/backups/${encodeURIComponent(filename)}`,
                    { method: "DELETE" },
                );
                if (res.ok) {
                    addToast("Backup deleted", "success");
                    loadBackups();
                } else {
                    addToast("Failed to delete backup", "error");
                }
            } catch (e) {
                addToast("Error deleting backup", "error");
            }
        }
    }

    function formatBytes(bytes, decimals = 2) {
        if (!+bytes) return "0 Bytes";
        const k = 1024;
        const dm = decimals < 0 ? 0 : decimals;
        const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
    }

    function formatDate(dateStr) {
        return new Date(dateStr).toLocaleString();
    }

    onMount(loadBackups);
</script>

<div class="h-full flex flex-col gap-4">
    <div class="flex justify-between items-center">
        <div>
            <h2 class="text-xl font-bold text-white flex items-center gap-2">
                <Archive class="w-6 h-6 text-indigo-400" /> Backups
            </h2>
            <p class="text-sm text-gray-400">Manage snapshots of your server</p>
        </div>
        <button
            on:click={createBackup}
            disabled={creating}
            class="bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
        >
            {#if creating}
                <Loader2 class="w-4 h-4 animate-spin" /> Creating...
            {:else}
                <Plus class="w-4 h-4" /> Create Backup
            {/if}
        </button>
    </div>

    <!-- Backup List -->
    <div
        class="bg-gray-900/50 border border-gray-800 rounded-xl overflow-hidden flex-1"
    >
        <div class="overflow-x-auto">
            <table class="w-full text-left text-sm">
                <thead>
                    <tr
                        class="bg-gray-900/80 border-b border-gray-800 text-gray-400 uppercase text-xs"
                    >
                        <th class="px-6 py-4 font-medium">Name</th>
                        <th class="px-6 py-4 font-medium">Size</th>
                        <th class="px-6 py-4 font-medium">Created At</th>
                        <th class="px-6 py-4 font-medium text-right">Actions</th
                        >
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-800">
                    {#if loading}
                        <tr>
                            <td
                                colspan="4"
                                class="px-6 py-8 text-center text-gray-500"
                            >
                                <Loader2
                                    class="w-6 h-6 animate-spin mx-auto mb-2"
                                />
                                Loading backups...
                            </td>
                        </tr>
                    {:else if backups.length === 0}
                        <tr>
                            <td
                                colspan="4"
                                class="px-6 py-12 text-center text-gray-500"
                            >
                                <div class="flex flex-col items-center gap-2">
                                    <Archive class="w-12 h-12 opacity-20" />
                                    <p>No backups found</p>
                                    <button
                                        on:click={createBackup}
                                        class="text-indigo-400 hover:text-indigo-300 text-sm mt-2 font-medium"
                                    >
                                        Create your first backup
                                    </button>
                                </div>
                            </td>
                        </tr>
                    {:else}
                        {#each backups as backup}
                            <tr
                                class="hover:bg-white/5 transition-colors group"
                            >
                                <td
                                    class="px-6 py-4 font-medium text-white flex items-center gap-3"
                                >
                                    <Archive class="w-4 h-4 text-indigo-400" />
                                    {backup.name}
                                </td>
                                <td class="px-6 py-4 text-gray-300"
                                    >{formatBytes(backup.size)}</td
                                >
                                <td class="px-6 py-4 text-gray-300"
                                    >{formatDate(backup.createdAt)}</td
                                >
                                <td class="px-6 py-4 text-right">
                                    <div
                                        class="flex justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity"
                                    >
                                        <button
                                            on:click={() =>
                                                restoreBackup(backup.name)}
                                            class="p-2 hover:bg-yellow-500/10 text-gray-400 hover:text-yellow-400 rounded-lg transition-colors"
                                            title="Restore"
                                        >
                                            <RotateCcw class="w-4 h-4" />
                                        </button>
                                        <button
                                            on:click={() =>
                                                deleteBackup(backup.name)}
                                            class="p-2 hover:bg-red-500/10 text-gray-400 hover:text-red-400 rounded-lg transition-colors"
                                            title="Delete"
                                        >
                                            <Trash2 class="w-4 h-4" />
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        {/each}
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>
