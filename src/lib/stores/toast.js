import { writable } from 'svelte/store';

export const toasts = writable([]);

export const addToast = (message, type = 'info', duration = 3000) => {
    const id = Math.random();
    toasts.update((all) => [...all, { id, message, type }]);

    if (duration) {
        setTimeout(() => {
            dismissToast(id);
        }, duration);
    }
};

export const dismissToast = (id) => {
    toasts.update((all) => all.filter((t) => t.id !== id));
};
