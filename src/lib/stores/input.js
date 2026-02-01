import { writable } from 'svelte/store';

 

 
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
