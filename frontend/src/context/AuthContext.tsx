import { create } from 'zustand';
import type {AuthState} from '../types';

const useAuthStore = create<AuthState>((set) => ({
    token: localStorage.getItem('token') || null,
    setToken: (token: string) => {
        localStorage.setItem('token', token);
        set({ token });
    },
    clearToken: () => {
        localStorage.removeItem('token');
        set({ token: null });
    },
}));

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => children;
export default useAuthStore;
