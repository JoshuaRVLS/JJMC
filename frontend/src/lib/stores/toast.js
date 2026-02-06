import { writable } from 'svelte/store';


/** @type {import('svelte/store').Writable<{id: number, message: string, type: string}[]>} */
export const toasts = writable([]);


/**
 * @param {string} message 
 * @param {string} type 
 * @param {number} duration 
 */
export const addToast = (message, type = 'info', duration = 3000) => {
    const id = Math.random();
    toasts.update((all) => [...all, { id, message, type }]);

    if (duration) {
        setTimeout(() => {
            dismissToast(id);
        }, duration);
    }
};


/** @param {number} id */
export const dismissToast = (id) => {
    toasts.update((all) => all.filter((t) => t.id !== id));
};
