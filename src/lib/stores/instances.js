import { writable, get } from 'svelte/store';
import { addToast } from "$lib/stores/toast";

/**
 * @typedef {Object} Instance
 * @property {string} id
 * @property {string} name
 * @property {string} type
 * @property {string} version
 * @property {string} status
 * @property {string} [folderId]
 * @property {number} [maxMemory]
 * @property {string} [javaArgs]
 * @property {string} [jarFile]
 * @property {string} [javaPath]
 * @property {string} [webhookUrl]
 * @property {string} [group]
 */

/**
 * @typedef {Object} Folder
 * @property {string} id
 * @property {string} name
 */

/** @type {import('svelte/store').Writable<Instance[]>} */
export const instances = writable([]);

/** @type {import('svelte/store').Writable<Folder[]>} */
export const folders = writable([]);

export const loading = writable(true);

export const actions = {
    async load() {
        loading.set(true);
        try {
            const [instRes, folderRes] = await Promise.all([
                fetch("/api/instances"),
                fetch("/api/folders"),
            ]);

            if (instRes.ok) instances.set(await instRes.json());
            if (folderRes.ok) folders.set(await folderRes.json());
        } catch (e) {
            console.error("Failed to load data", e);
            addToast("Failed to load data", "error");
        } finally {
            loading.set(false);
        }
    },

    /**
     * @param {string} instanceId
     * @param {string} folderId
     */
    async moveInstance(instanceId, folderId) {
        // Optimistic update
        instances.update(all => all.map(i => {
            if (i.id === instanceId) {
                return { ...i, folderId: folderId };
            }
            return i;
        }));

        try {
            // Find the instance to get its current properties
            const allInstances = get(instances);
            const inst = allInstances.find(i => i.id === instanceId);

            if (!inst) return;

            const payload = {
                maxMemory: inst.maxMemory,
                javaArgs: inst.javaArgs,
                jarFile: inst.jarFile,
                javaPath: inst.javaPath,
                webhookUrl: inst.webhookUrl,
                group: inst.group,
                folderId: folderId,
            };

            const res = await fetch(`/api/instances/${instanceId}`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });

            if (!res.ok) throw new Error(await res.text());
            addToast("Instance moved", "success");
            // No need to reload, we updated optimistically. 
            // But we can reload to be safe eventually or if we want to sync status.
            this.load();
        } catch (e) {
            console.error("Move failed", e);
            addToast("Failed to move instance", "error");
            this.load(); // Revert
        }
    },

    /**
     * @param {string} name 
     */
    async createFolder(name) {
        try {
            const res = await fetch("/api/folders", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ name }),
            });
            if (res.ok) {
                addToast("Folder created", "success");
                this.load();
            } else {
                addToast("Failed to create folder", "error");
            }
        } catch (e) {
            addToast("Error creating folder", "error");
        }
    },

    /**
     * @param {string} id 
     */
    async deleteFolder(id) {
        try {
            // Optimistic removal from list
            folders.update(all => all.filter(f => f.id !== id));
            // Optimistic move of instances to Uncategorized (null folderId)
            instances.update(all => all.map(i => i.folderId === id ? { ...i, folderId: "" } : i));

            const res = await fetch(`/api/folders/${id}`, {
                method: "DELETE",
            });
            if (res.ok) {
                addToast("Folder deleted", "success");
                this.load(); // Sync
            } else {
                throw new Error("Failed to delete");
            }
        } catch (e) {
            addToast("Error deleting folder", "error");
            this.load(); // Revert
        }
    }
};
