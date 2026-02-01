import { writable } from 'svelte/store';

 

 
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

 
export const closeConfirm = (result) => {
    confirmState.update(s => {
        if (s.resolve) s.resolve(result);
        return { ...s, active: false };
    });
};
