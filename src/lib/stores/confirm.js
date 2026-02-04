import { writable } from 'svelte/store';




/** 
 * @typedef {Object} ConfirmState
 * @property {boolean} active
 * @property {string} title
 * @property {string} message
 * @property {string} confirmText
 * @property {string} cancelText
 * @property {boolean} dangerous
 * @property {((value: any) => void) | null} resolve
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
 * @param {{ title: string, message: string, confirmText?: string, cancelText?: string, dangerous?: boolean }} params
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
