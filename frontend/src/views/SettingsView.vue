<script setup lang="ts">
import { ref } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useNotificationsStore } from '@/stores/notifications';

const authStore = useAuthStore();

const username = ref('');
const oldPassword = ref('');
const newPassword = ref('');
const passwordConfirm = ref('');
const userParamsOk = ref(false);

const deleterUserConfirm = ref(false);

const usernameOk = () => username.value.length >= 3;
const passwordOk = () => newPassword.value.length >= 3;
const passwordConfirmOk = () => newPassword.value === passwordConfirm.value;

function checkUserParams(isRegistering: boolean = true, notifications: boolean = false): boolean {

    //console.log("Checking params:", username.value, oldPassword.value, newPassword.value, passwordConfirm.value);

    if (isRegistering) {
        if (!usernameOk()) {
            if (notifications) useNotificationsStore().notifyError('Username must be at least 3 characters long');
            return userParamsOk.value = false;
        }
    }
    else if (oldPassword.value === '') {
        if (notifications) useNotificationsStore().notifyError('You must enter your old password to change it');
        return userParamsOk.value = false;
    }

    if (!passwordOk()) {
        if (notifications) useNotificationsStore().notifyError('Password must be at least 3 characters long');
        return userParamsOk.value = false;
    }

    if (!passwordConfirmOk()) {
        if (notifications) useNotificationsStore().notifyError('Passwords do not match');
        return userParamsOk.value = false;
    }

    return userParamsOk.value = true;
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
    <section class="settings-view">
        <h2 class="vue-heading">Settings</h2>

        <h3>Authentication</h3>
        <section id="authentication"">
            <div class=" card">
            <div v-if="authStore.needsAuth && authStore.user">
                <h4>Change Password</h4>
                <form @submit.prevent="changePassword" @keypress="checkUserParams(false, false)">
                    <input class="text-input" v-model="oldPassword" type="password" placeholder="Old Password"
                        name="password" required />
                    <small>Password must be at least 3 characters long</small>
                    <input class="text-input" v-model="newPassword" type="password" placeholder="New Password"
                        name="password" required />
                    <input class="text-input" v-model="passwordConfirm" type="password" placeholder="Confirm Password"
                        name="passwordConfirm" required />
                    <button class="search-button" type="submit" :disabled="!userParamsOk">Change Password</button>
                </form>
            </div>
            <div v-else>
                <h4>Create User</h4>
                Creating a user will enable authentication, meaning you will need to log in before using the app.
                Recommended if you plan on having the
                app be publically accessible.
                <form @submit.prevent="registerUser" @keyup="checkUserParams(true, false)">
                    <small>Username must be at least 3 characters long</small>
                    <input class="text-input" v-model="username" type="text" placeholder="Username" name="username"
                        required />
                    <small>Password must be at least 3 characters long</small>
                    <input class="text-input" v-model="newPassword" type="password" placeholder="Password"
                        name="password" required />
                    <input class="text-input" v-model="passwordConfirm" type="password" placeholder="Confirm Password"
                        name="passwordConfirm" required />
                    <button class="search-button" type="submit" :disabled="!userParamsOk">Create User</button>
                </form>
            </div>
            </div>
            <div class="card" v-if="authStore.needsAuth && authStore.user">
                <h4>Remove User</h4>
                <p>Removing the user will remove all authentication from the app. You should not remove the user if your
                    app is publically accessible.</p>
                <div class="warning-box">
                    <label for="deleterUserConfirm">I understand <input type="checkbox" v-model="deleterUserConfirm"
                            id="deleterUserConfirm"></label>

                    <button class="delete-button" :disabled="!deleterUserConfirm" @click="authStore.deleteUser">Remove
                        User</button>
                </div>
            </div>
        </section>
    </section>
</template>

<style scoped>
.settings-view {
    display: block;
    overflow-y: auto;
    padding-bottom: 10rem;
    box-sizing: border-box;
}

.settings-view section {
    display: flex;
    justify-content: center;
    gap: 0.8rem;
    margin-top: 0.8rem;
    height: 100%;
}

.card {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    width: -webkit-fill-available;
    max-width: 600px;
    border: 1px solid var(--color-gray-700);
    border-radius: 6px;
    padding: 1rem;
    margin-bottom: 1rem;
}

.card form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-top: 1rem;
}

.card h4 {
    margin: 0;
    margin-bottom: 0.5rem;
    font-size: 1rem;
}

.settings-view h3 {
    margin: 0;
    margin-bottom: 0.5rem;
    font-size: 1.2rem;
}

.warning-box {
    display: flex;
    justify-content: space-between;
    background: var(--color-warning-background);
    border: 1px solid var(--color-warning-border);
    border-radius: 6px;
    padding: 1rem;
    margin-top: 1rem;
}

.warning-box label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

@media (max-width: 768px) {
    .settings-view section {
        display: flex;
        flex-direction: column;
        gap: 0.8rem;
        margin-bottom: 0.8rem;
        height: 100%;
    }

    .warning-box {
        flex-direction: column;
        align-content: center;
        gap: 1rem;
    }
}
</style>