import { writable } from 'svelte/store';
import { browser } from '$app/environment';

function createAuthStore() {
    let initialToken = null;
    let initialUser = null;

    if (browser) {
        initialToken = localStorage.getItem('token');
        const storedUser = localStorage.getItem('user');
        try {
            if (storedUser && storedUser !== 'undefined') {
                initialUser = JSON.parse(storedUser);
            }
        } catch (e) {
            console.error('Failed to parse stored user:', e);
            localStorage.removeItem('user');
        }
    }

    const { subscribe, set, update } = writable({
        token: initialToken,
        user: initialUser,
    });

    return {
        subscribe,
        login: (token, user) => {
            if (browser) {
                localStorage.setItem('token', token);
                localStorage.setItem('user', JSON.stringify(user));
            }
            set({ token, user });
        },
        logout: () => {
            if (browser) {
                localStorage.removeItem('token');
                localStorage.removeItem('user');
            }
            set({ token: null, user: null });
        },
    };
}

export const auth = createAuthStore();
