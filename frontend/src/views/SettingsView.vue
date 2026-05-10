<script setup lang="ts">
import { ref } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useNotificationsStore } from '@/stores/notifications';

const authStore = useAuthStore();

const username = ref('');
const password = ref('');
const passwordConfirm = ref('');

function checkUserParams(): boolean {

    if (username.value.length < 3) {
        useNotificationsStore().notifyError('Username must be at least 3 characters long');
        return false;
    }

    if (password.value.length < 3) {
        useNotificationsStore().notifyError('Password must be at least 8 characters long');
        return false;
    }

    if (password.value !== passwordConfirm.value) {
        useNotificationsStore().notifyError('Passwords do not match');
        return false;
    }
    return true;
}

async function registerUser() {

    if (!checkUserParams()) return;

    console.log(username.value, password.value)
    
    await authStore.register({
        username: username.value,
        password: password.value
    });
}

async function changePassword() {
    if (!checkUserParams()) return;
    await authStore.changePassword(password.value);
}

</script>

<template>
    <div class="settings-view">
        <h2>Settings</h2>

        <section id="Users" class="card">
            <h3>Users</h3>
            <div v-if="authStore.needsAuth && authStore.user">
                <h4>Change Password</h4>
                <form>
                    <input v-model="password" type="password" placeholder="New Password" />
                    <input v-model="passwordConfirm" type="password" placeholder="Confirm Password" />
                    <button type="submit" @click="changePassword">Change Password</button>
                </form>
            </div>
            <div v-else>
                <h4>Create User</h4>
                Creating a user will force you to be logged in to use the app. Recommended if you plan on having the app be publically accessible.
                <form @submit.prevent="registerUser">
                    <input v-model="username" type="text" placeholder="Username" />
                    <input v-model="password" type="password" placeholder="Password" />
                    <input v-model="passwordConfirm" type="password" placeholder="Confirm Password" />
                    <button type="submit">Create User</button>
                </form>
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

</style>