import { writable } from 'svelte/store';

const initial = {
    active: false,
    title: '',
    message: '',
    value: '', // Initial value
    placeholder: '',
    confirmText: 'Confirm',
    cancelText: 'Cancel',
    resolve: null
};

export const inputState = writable({ ...initial });

/**
 * prompts the user for input.
 * @returns {Promise<string|null>} The input value if confirmed, or null if cancelled.
 */
export const askInput = ({ title, message = '', value = '', placeholder = '', confirmText = 'Confirm', cancelText = 'Cancel' }) => {
    return new Promise((resolve) => {
        inputState.set({
            active: true,
            title,
            message,
            value,
            placeholder,
            confirmText,
            cancelText,
            resolve
        });
    });
};

export const closeInput = (result) => {
    inputState.update(s => {
        if (s.resolve) s.resolve(result);
        return { ...s, active: false };
    });
};
