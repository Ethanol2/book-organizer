<script setup lang="ts">
import { ref } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useNotificationsStore } from '@/stores/notifications';

const authStore = useAuthStore();

const username = ref('');
const oldPassword = ref('');
const newPassword = ref('');
const passwordConfirm = ref('');

const confirmDeleteUser = ref(false);

function checkUserParams(checkUsername: boolean = true): boolean {

    if (checkUsername) {
        if (username.value.length < 3) {
            useNotificationsStore().notifyError('Username must be at least 3 characters long');
            return false;
        }
    }

    if (newPassword.value.length < 3) {
        useNotificationsStore().notifyError('Password must be at least 3 characters long');
        return false;
    }

    if (newPassword.value !== passwordConfirm.value) {
        useNotificationsStore().notifyError('Passwords do not match');
        return false;
    }
    return true;
}

async function registerUser() {

    if (!checkUserParams()) return;

    console.log(username.value, newPassword.value)

    await authStore.register({
        username: username.value,
        password: newPassword.value
    });
}

async function changePassword() {
    if (!checkUserParams(false)) return;
    await authStore.changePassword(oldPassword.value, newPassword.value);
}

</script>

<template>
    <div class="settings-view">
        <h2>Settings</h2>

        <section id="Users">
            <h3>Users</h3>
            <div class="card">
                <div v-if="authStore.needsAuth && authStore.user">
                    <h4>Change Password</h4>
                    <form @submit.prevent="changePassword">
                        <input class="search-input" v-model="oldPassword" type="password" placeholder="Old Password" />
                        <input class="search-input" v-model="newPassword" type="password" placeholder="New Password" />
                        <input class="search-input" v-model="passwordConfirm" type="password"
                            placeholder="Confirm Password" />
                        <button class="search-button" type="submit">Change Password</button>
                    </form>
                </div>
                <div v-else>
                    <h4>Create User</h4>
                    Creating a user will force you to be logged in to use the app. Recommended if you plan on having the
                    app be publically accessible.
                    <form @submit.prevent="registerUser">
                        <input v-model="username" type="text" placeholder="Username" />
                        <input v-model="newPassword" type="password" placeholder="Password" />
                        <input v-model="passwordConfirm" type="password" placeholder="Confirm Password" />
                        <button type="submit">Create User</button>
                    </form>
                </div>
            </div>
            <div class="card" v-if="authStore.needsAuth && authStore.user">
                <h4>Remove User</h4>
                <button class="delete-button" type="button" @click="() => confirmDeleteUser = true">Delete User</button>
                <button v-if="confirmDeleteUser" class="delete-button" type="button"
                    @click="confirmDeleteUser = true">Delete Account</button>
            </div>
        </section>
    </div>
</template>

<style scoped>
.settings-view {
    display: block;
    overflow-y: auto;
    padding: 1rem;
    padding-bottom: 10rem;
    box-sizing: border-box;
}

.card {
    width: 100%;
    border: 1px solid var(--color-gray-700);
    border-radius: 6px;
    padding: 6px;
    margin-bottom: 1rem;
}

.card form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.settings-view h3 {
    margin: 0;
    margin-bottom: 0.5rem;
    font-size: 1.2rem;
}

.card h4 {
    margin: 0;
    margin-bottom: 0.5rem;
    font-size: 1rem;
}
</style>