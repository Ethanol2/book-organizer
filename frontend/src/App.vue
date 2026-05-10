<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import Notifications from '@/components/Notifications.vue'
import { useAuthStore } from './stores/auth';

const menuOpen = ref(false)
const authStore = useAuthStore();

</script>

<template>
  <div class="app">
    <header class="header">
      <button class="hamburger" @click="menuOpen = !menuOpen" aria-label="Toggle menu">
        <span></span>
        <span></span>
        <span></span>
      </button>
      <h1>
        <img alt="book organizer logo" class="logo" src="@/assets/book-organizer-logo.png" />
        <RouterLink to="/">Book Organizer</RouterLink>
      </h1>
    </header>

    <section class="body">
      <div v-if="menuOpen" class="backdrop" @click="menuOpen = false"></div>
      <aside class="sidebar" :class="{ open: menuOpen }">
        <nav>
          <template class="login" v-if="!authStore.isAuthenticated">
            <RouterLink to="/login" @click="menuOpen = false">Login</RouterLink>
          </template>
          <template v-else>
            <RouterLink to="/" @click="menuOpen = false">Library</RouterLink>
            <RouterLink to="/add-book" @click="menuOpen = false">Add Book</RouterLink>
            <RouterLink to="/downloads" @click="menuOpen = false">Downloads</RouterLink>
            <RouterLink to="/settings" @click="menuOpen = false">Settings</RouterLink>
          </template>
          <RouterLink to="/about" @click="menuOpen = false">About</RouterLink>
        </nav>
      </aside>

      <main class="content">
        <RouterView />
      </main>
    </section>

    <Notifications />
  </div>
</template>

<style scoped>
.app {
  height: 100vh;
  width: 100%;
}

.header {
  display: flex;
  align-items: center;
  height: 80px;
  background-color: var(--vt-c-nav);
  padding: 1rem;
  font-family: "Libre Baskerville", serif;
  box-shadow: 5px 5px 5px rgba(0, 0, 0, 0.4);
  gap: 1rem;
}

.header h1 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  height: 100%;
}

.hamburger {
  display: none;
  flex-direction: column;
  width: 30px;
  height: 30px;
  background: none;
  border: none;
  cursor: pointer;
  gap: 5px;
}

.hamburger span {
  width: 100%;
  height: 3px;
  background-color: var(--color-text);
  border-radius: 2px;
  transition: all 0.3s ease;
}

.body {
  display: flex;
  margin: 0;
  height: calc(100vh - 80px);
  overflow: hidden;
}

.sidebar {
  width: 120px;
  background-color: var(--vt-c-nav);
  padding: 1rem;
  box-shadow: 5px 5px 5px rgba(0, 0, 0, 0.4);
}

.sidebar.login {
  padding-bottom: 2rem;
}

.content {
  padding: 1.5rem;
  overflow-y: auto;
  width: 100%;
}

.logo {
  object-fit: contain;
  height: 100%;
  width: auto;
  vertical-align: middle;
  padding: 0.2rem;
}

nav {
  display: flex;
  flex-direction: column;
  margin-top: 1rem;
}

nav a {
  padding: 1rem 0;
  text-decoration: none;
  color: var(--color-text)
}

nav a.router-link-exact-active {
  font-weight: 600;
}

.backdrop {
  display: none;
  position: fixed;
  top: 80px;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 99;
}

/* Mobile styles */
@media (max-width: 768px) {
  .hamburger {
    display: flex;
  }

  .backdrop {
    display: block;
  }

  .sidebar {
    position: fixed;
    top: 80px;
    left: 0;
    height: calc(100vh - 80px);
    width: 200px;
    transform: translateX(-100%);
    transition: transform 0.3s ease;
    z-index: 100;
  }

  .sidebar.open {
    transform: translateX(0);
  }

  .body {
    position: relative;
  }

  .content {
    padding: 1rem;
  }

  nav a {
    padding: 0.8rem 0;
  }

  h1 {
    font-size: 1.2rem;
  }
}

/* Desktop styles */
@media (min-width: 769px) {
  .backdrop {
    display: none !important;
  }
}
</style>
