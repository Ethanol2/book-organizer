import { defineStore } from "pinia";
import { useNotificationsStore } from "./notifications";
import api from "@/services/api";
import { isAxiosError } from "axios";

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
        isAuthenticated: (state) => !!state.user || !state.needsAuth,
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

            // console.log(this.user)
            // console.log(this.isAuthenticated, this.needsAuth)
        },
        async refreshSession() {
            try {
                await api.post('/api/auth/refresh')
            }
            catch (err) {
                console.error('Error refreshing session:', err);
                useNotificationsStore().notifyError("The session could not be refreshed")
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
        async login(params: UserParams): Promise<boolean> {
            try {
                const resp = await api.post('/api/auth/login', params)
                this.user = resp.data.user as User;
                useNotificationsStore().notifySuccess('Logged in successfully!')
                return true;
            }
            catch (err) {

                if (!isAxiosError(err)) {
                    console.error('Error logging in:', err);
                    useNotificationsStore().notifyError("Something went wrong with the authentication server");
                    return false;
                }

                if (err.response) {
                    if (err.response.status === 401) {
                        useNotificationsStore().notifyError('Invalid password')
                    }
                    else if (err.response.status === 404) {
                        useNotificationsStore().notifyError('User not found')
                    }
                    else {
                        console.error('Error logging in:', err);
                        useNotificationsStore().notifyError("Something went wrong with the authentication server")
                    }
                }
                else {
                    console.error('Error logging in:', err);
                    useNotificationsStore().notifyError("Something went wrong with the authentication server");
                }
            }
            return false;
        },
        async logout() {
            try {
                await api.post('/api/auth/logout')
                this.user = null;
                useNotificationsStore().notifySuccess('Logged out successfully!')
                window.location.reload();
            }
            catch (err) {
                console.error('Error logging out:', err);
                useNotificationsStore().notifyError("Something went wrong with the authentication server")
            }
        },
        async changePassword(password: string) {

            if (this.user === null) {
                return;
            }

            try {
                await api.post('/api/auth/users/' + this.user.id + '/reset-password', {
                    password: password
                });
            }
            catch (err) {
                console.error('Error changing password:', err);
                useNotificationsStore().notifyError("Something went wrong with the authentication server")
            }
        }
    }
});
