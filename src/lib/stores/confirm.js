import { writable } from 'svelte/store';

/**
 * @typedef {Object} ConfirmState
 * @property {boolean} active
 * @property {string} title
 * @property {string} message
 * @property {string} confirmText
 * @property {string} cancelText
 * @property {boolean} dangerous
 * @property {((value: boolean) => void) | null} resolve
 */

/** @type {ConfirmState} */
const initial = {
    active: false,
    title: '',
    message: '',
    confirmText: 'Confirm',
    cancelText: 'Cancel',
    dangerous: false,
    resolve: null
};

export const confirmState = writable(initial);

/**
 * @param {Object} params
 * @param {string} params.title
 * @param {string} params.message
 * @param {string} [params.confirmText]
 * @param {string} [params.cancelText]
 * @param {boolean} [params.dangerous]
 * @returns {Promise<boolean>}
 */
export const askConfirm = ({ title, message, confirmText = 'Confirm', cancelText = 'Cancel', dangerous = false }) => {
    return new Promise((resolve) => {
        confirmState.set({
            active: true,
            title,
            message,
            confirmText,
            cancelText,
            dangerous,
            resolve
        });
    });
};

/** @param {boolean} result */
export const closeConfirm = (result) => {
    confirmState.update(s => {
        if (s.resolve) s.resolve(result);
        return { ...s, active: false };
    });
};
