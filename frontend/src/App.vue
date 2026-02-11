<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import HelloWorld from './components/HelloWorld.vue'
</script>

<template>
  <div class="app">
    <aside class="sidebar">
      <img alt="book organizer logo" class="logo" src="@/assets/book-organizer-logo.png" />

      <nav>
        <RouterLink to="/">Library</RouterLink>
        <RouterLink to="/downloads">Downloads</RouterLink>
        <RouterLink to="/about">About</RouterLink>
      </nav>
    </aside>

    <main class="content">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.app {
  display: flex;
  height: 100vh;
}

.sidebar {
  width: 150px;
  border-right: 1px solid grey;
  padding: 1rem;
}

.content {
  padding: 1.5rem;
  overflow-y: auto;
  width: 100%;
}

.logo {
  object-fit: contain;
  height: auto;
  width: 100%;
}

nav {
  display: flex;
  flex-direction: column;
  margin-top: 2rem;
}

nav a {
  padding: 0.5rem 0;
  text-decoration: none;
  color: var(--color-text)
}

nav a.router-link-exact-active {
  font-weight: 600;
}
</style>
