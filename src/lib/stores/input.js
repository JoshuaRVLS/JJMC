import { writable } from 'svelte/store';

/**
 * @typedef {Object} InputState
 * @property {boolean} active
 * @property {string} title
 * @property {string} message
 * @property {string} value
 * @property {string} placeholder
 * @property {string} confirmText
 * @property {string} cancelText
 * @property {((value: string|null) => void) | null} resolve
 */

/** @type {InputState} */
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

export const inputState = writable(initial);

/**
 * prompts the user for input.
 * @param {Object} params
 * @param {string} params.title
 * @param {string} [params.message]
 * @param {string} [params.value]
 * @param {string} [params.placeholder]
 * @param {string} [params.confirmText]
 * @param {string} [params.cancelText]
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

/** @param {string|null} result */
export const closeInput = (result) => {
    inputState.update(s => {
        if (s.resolve) s.resolve(result);
        return { ...s, active: false };
    });
};
