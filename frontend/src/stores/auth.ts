import { defineStore } from "pinia";
import { useNotificationsStore } from "./notifications";
import api from "@/services/api";

export type User = {
    id: string
    created_at: string
    updated_at: string
    username: string
}

export type UserParams = {
    username: string
    password: string
}

export const useAuthStore = defineStore('auth', {
    state: () => ({
        user: null as User | null,
        needsAuth: false,
        initialized: false,
    }),
    getters: {
        isAuthenticated: (state) => !!state.user,
    },
    actions: {
        async checkCurrentStatus() {
            try {
                const resp = await api.get('/api/auth/status')

                this.needsAuth = resp.data.user_count > 0;
                if (resp.data.user === null) {
                    this.user = null;
                }
                else {
                    this.user = resp.data.user as User;
                }

                this.initialized = true;

            } catch (err) {
                console.error('Error fetching auth status:', err);
                useNotificationsStore().notifyError("Something went wrong with the authentication server")
            }
        },
        async refreshSession() {
            try {
                await api.post('/api/auth/refresh')
            }
            catch (err) {
                console.error('Error refreshing session:', err);
                useNotificationsStore().notifyError("Something went wrong with the authentication server")
            }
        },
        async register(params: UserParams) {
            try {
                await api.post('/api/auth/register', params)
                useNotificationsStore().notifySuccess('User registered successfully!')
            }
            catch (err) {
                console.error('Error registering user:', err);
                useNotificationsStore().notifyError("Something went wrong with the authentication server")
            }
        },
        async login(params: UserParams) {
            try {
                const resp = await api.post('/api/auth/login', params)
                this.user = resp.data.user as User;
                useNotificationsStore().notifySuccess('Logged in successfully!')
            }
            catch (err) {
                console.error('Error logging in:', err);
                useNotificationsStore().notifyError("Something went wrong with the authentication server")
            }
        },
        async logout() {
            try {
                await api.post('/api/auth/logout')
                this.user = null;
                useNotificationsStore().notifySuccess('Logged out successfully!')
            }
            catch (err) {
                console.error('Error logging out:', err);
                useNotificationsStore().notifyError("Something went wrong with the authentication server")
            }
        }
    }
});
