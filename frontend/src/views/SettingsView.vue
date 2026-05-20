<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useNotificationsStore } from '@/stores/notifications';
import api from '@/services/api';
import LogsViewer from '@/components/LogsViewer.vue';

type UserSettings = {
    google_api_key: string,
    scan_frequency: number,
    port: number
}

const authStore = useAuthStore();

// User Settings
const userSettings = ref<UserSettings>({
    google_api_key: '',
    scan_frequency: 5,
    port: 8080
});
const port = ref(8080);
const portValid = ref(true);
const google_api_key = ref('');
const showKey = ref(false);
const scan_frequency = ref(5);
const scanFrequencyValid = ref(true);

const validatePort = () => port.value >= 1 && port.value <= 65535;
const validateScanFrequency = () => scan_frequency.value >= 1;

// Authentication Params
const username = ref('');
const oldPassword = ref('');
const newPassword = ref('');
const passwordConfirm = ref('');
const userParamsOk = ref(false);

const deleterUserConfirm = ref(false);

const usernameOk = () => username.value.length >= 3;
const passwordOk = () => newPassword.value.length >= 3;
const passwordConfirmOk = () => newPassword.value === passwordConfirm.value;

async function getUserSettings() {
    try {
        const resp = await api.get('/api/settings');
        userSettings.value = resp.data;

        port.value = userSettings.value.port;
        google_api_key.value = userSettings.value.google_api_key;
        scan_frequency.value = userSettings.value.scan_frequency;

    } catch (err) {
        console.error('Error fetching user settings:', err);
        useNotificationsStore().notifyError('Something went wrong while fetching user settings')
    }
}

async function postUserSettings() {
    const newSettings = {
        google_api_key: google_api_key.value,
        scan_frequency: scan_frequency.value,
        port: port.value
    }

    try {
        const resp = await api.post('/api/settings', newSettings);
        useNotificationsStore().notifySuccess('Settings saved successfully!');

        userSettings.value = newSettings;
    } catch (err) {
        console.error('Error saving user settings:', err);
        useNotificationsStore().notifyError('Something went wrong while saving user settings')
    }
}

function settingsChanged(): boolean {
    return port.value !== userSettings.value.port ||
        google_api_key.value !== userSettings.value.google_api_key ||
        scan_frequency.value !== userSettings.value.scan_frequency
}

function allSettingsValid(): boolean {
    return validatePort() && validateScanFrequency();
}

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

onMounted(async () => {
    await getUserSettings();
})

</script>

<template>
    <section class="settings-view">
        <h2 class="vue-heading">Settings</h2>

        <!-- General Settings -->
        <h3>General Settings</h3>
        <section id="general-settings">
            <div class="card">
                <div>
                    <h4>Port</h4>
                    <small>Server will have to restart for settings to take effect</small>
                    <input :class="'text-input ' + (portValid ? '' : 'invalid')" v-model="port" type="text"
                        placeholder="Port" name="port" @keyup="portValid = validatePort()" />
                </div>
                <div>
                    <h4>Google API Key</h4>
                    <small>Required to use Google Books as a metadata source. <a
                            href="https://developers.google.com/books/docs/v1/using" target="_blank">Google's
                            documentation</a></small>
                    <div class="hidden-field">
                        <input class="text-input" v-model="google_api_key" :type="showKey ? 'text' : 'password'"
                            placeholder=" Google API Key" name="google_api_key" />
                        <button class="toggle-button" @click="showKey = !showKey">{{ showKey ? "Hide" : "Show"
                            }}</button>
                    </div>
                </div>
                <div>
                    <h4>Scan Frequency</h4>
                    <p>How often to scan the downloads folder</p>
                    <small>Number in seconds. Must be larger than 0 seconds</small>
                    <input :class="'text-input ' + (scanFrequencyValid ? '' : 'invalid')"" v-model="scan_frequency"
                        type="text" placeholder="Scan Frequency" name="scan_frequency"
                        @keyup=" scanFrequencyValid = validateScanFrequency()" />
                </div>
                <button class="green-button" v-if="settingsChanged()" :disabled="!allSettingsValid()"
                    @click="postUserSettings()">Apply</button>
            </div>
        </section>

        <!-- Authentication -->
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
                    <button class="green-button" type="submit" :disabled="!userParamsOk">Change Password</button>
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
                    <button class="green-button" type="submit" :disabled="!userParamsOk">Create User</button>
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

        <!-- Logs -->
        <h3>Logs</h3>
        <section id="logs">
            <div class="card logs">
                <LogsViewer />
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

.settings-view h3 {
    margin: 0;
    margin-bottom: 0.5rem;
    font-size: 1.2rem;
}

.general-settings {
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
    align-items: center;
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
    gap: 1rem;
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

.card.logs {
    max-width: 800px;
    height: fit-content;
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

.text-input {
    height: auto;
}

.text-input.invalid {
    border-color: var(--color-error-border);
}

.hidden-field {
    display: flex;
    gap: 0.5rem;
    justify-content: space-between;
}

.hidden-field button {
    max-width: 100px;
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