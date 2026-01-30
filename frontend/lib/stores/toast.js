import { writable } from 'svelte/store';

/** @type {import('svelte/store').Writable<Array<{id: number, message: string, type: string}>>} */
export const toasts = writable([]);

/**
 * @param {string} message
 * @param {string} [type='info']
 * @param {number} [duration=3000]
 */
export const addToast = (message, type = 'info', duration = 3000) => {
    const id = Math.random();
    toasts.update((/** @type {Array<any>} */ all) => [...all, { id, message, type }]);

    if (duration) {
        setTimeout(() => {
            dismissToast(id);
        }, duration);
    }
};

/** @param {number} id */
export const dismissToast = (id) => {
    toasts.update((/** @type {Array<{id: number, message: string, type: string}>} */ all) => all.filter((t) => t.id !== id));
};
