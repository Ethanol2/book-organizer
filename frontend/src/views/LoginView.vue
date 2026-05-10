<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useNotificationsStore } from '@/stores/notifications';
import { useAuthStore } from '@/stores/auth';

const router = useRouter();
const notificationsStore = useNotificationsStore();
const authStore = useAuthStore();
const username = ref('');
const password = ref('');

async function login() {
  if (username.value === '' || password.value === '') {
    notificationsStore.notifyError('Username and password are required');
    return;
  }

  try {
    const ok = await authStore.login({
      username: username.value,
      password: password.value,
    });

    if (ok) {
      router.push('/');
      window.location.reload();
    }
  } catch (error) {
    console.error('Error logging in:', error);
  }
}

</script>

<template>
  <div class="login-page">
    <div class="login-card">

      <div class="branding">
        <h1 class="title">Login</h1>
        <p class="subtitle">
          Organize your all your books in one cozy place.
        </p>
      </div>

      <form class="login-form" @submit.prevent="login">
        <div class="field">
          <label>Username</label>
          <input type="text" v-model="username" autocomplete="username" />
        </div>

        <div class="field">
          <label>Password</label>
          <input type="password" v-model="password" autocomplete="current-password" />
        </div>

        <button type="submit" class="search-button">
          Sign In
        </button>
      </form>

    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.login-card {
  width: 100%;
  max-width: 420px;

  background: var(--color-background-soft);
  border: 1px solid var(--color-border);

  border-radius: 18px;

  padding: 2.5rem;

  box-shadow:
    0 8px 30px rgba(0, 0, 0, 0.06);
}

.branding {
  margin-bottom: 2rem;
}

.title {
  font-family: var(--font-title);
  color: var(--color-title);

  font-size: 2.5rem;
  font-weight: 600;

  letter-spacing: 0.02em;

  margin-bottom: 0.5rem;
}

.subtitle {
  color: var(--color-text-subtle);

  line-height: 1.5;
  font-size: 0.95rem;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.field label {
  color: var(--color-heading);
  font-size: 0.9rem;
  font-weight: 500;
}

.field input {
  background: var(--color-background);

  border: 1px solid var(--color-border);

  border-radius: 10px;

  padding: 0.85rem 1rem;

  color: var(--color-text);

  font-size: 0.95rem;

  transition:
    border-color 0.15s ease,
    background 0.15s ease;
}

.field input::placeholder {
  color: var(--color-text-subtle);
}

.field input:focus {
  outline: none;

  border-color: var(--color-primary-green);

  background: var(--color-background-mute);
}

.login-button {
  margin-top: 0.5rem;

  background: var(--vt-c-nav);
  color: var(--color-heading);

  border: none;

  border-radius: 10px;

  padding: 0.9rem 1rem;

  font-size: 0.95rem;
  font-weight: 600;

  cursor: pointer;

  transition:
    background 0.15s ease,
    transform 0.1s ease;
}

.login-button:hover {
  background: var(--color-nav-hover-bg);
  color: var(--color-nav-hover-text);
}

.login-button:active {
  transform: translateY(1px);
}

.helper-text {
  text-align: center;

  color: var(--color-text-subtle);

  font-size: 0.85rem;

  margin-top: 0.5rem;
}

/* Dark mode shadow adjustment */
:root.dark .login-card {
  box-shadow:
    0 10px 40px rgba(0, 0, 0, 0.35);
}
</style>