import { Writable } from 'svelte/store';

export interface User {
    id: string;
    email: string;
    username: string;
}

export interface AuthState {
    token: string | null;
    user: User | null;
}

export interface AuthStore extends Writable<AuthState> {
    login(token: string, user: User): void;
    logout(): void;
}

export const auth: AuthStore;
