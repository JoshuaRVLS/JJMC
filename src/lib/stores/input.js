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
 * @property {((value: any) => void) | null} resolve
 */

/** @type {InputState} */
const initial = {
    active: false,
    title: '',
    message: '',
    value: '',
    placeholder: '',
    confirmText: 'Confirm',
    cancelText: 'Cancel',
    resolve: null
};

export const inputState = writable(initial);


/**
 * @param {{ title: string, message?: string, value?: string, placeholder?: string, confirmText?: string, cancelText?: string }} params
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


/** @param {string | null} result */
export const closeInput = (result) => {
    inputState.update(s => {
        if (s.resolve) s.resolve(result);
        return { ...s, active: false };
    });
};
