<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import Notifications from '@/components/Notifications.vue'
</script>

<template>
  <div class="app">
    <header class="header">
      <h1>
        <img alt="book organizer logo" class="logo" src="@/assets/book-organizer-logo.png" />
        <RouterLink to="/">Book Organizer</RouterLink>
      </h1>
    </header>

    <section class="body">
      <aside class="sidebar">
        <nav>
          <RouterLink to="/">Library</RouterLink>
          <RouterLink to="/add-book">Add Book</RouterLink>
          <RouterLink to="/downloads">Downloads</RouterLink>
          <RouterLink to="/about">About</RouterLink>
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
  height: 80px;
  background-color: var(--vt-c-nav);
  padding: 1rem;
  font-family: "Libre Baskerville", serif;
  box-shadow: 5px 5px 5px rgba(0, 0, 0, 0.4);
}

.body {
  display: flex;
  margin: 0;
  height: 100%;
  overflow: hidden;
}

.sidebar {
  width: 120px;
  background-color: var(--vt-c-nav);
  padding: 1rem;
  box-shadow: 5px 5px 5px rgba(0, 0, 0, 0.4);
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
</style>
