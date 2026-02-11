<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import HelloWorld from './components/HelloWorld.vue'
</script>

<template>
  <div class="app">
    <header class="header">
      <h1>
        <img alt="book organizer logo" class="logo" src="@/assets/book-organizer-logo.png" />
        <RouterLink to="/">Book Organizer</RouterLink>
      </h1>
    </header>

    <body class="body">
      <aside class="sidebar">

        <nav>
          <RouterLink to="/">Library</RouterLink>
          <RouterLink to="/downloads">Downloads</RouterLink>
          <RouterLink to="/about">About</RouterLink>
        </nav>
      </aside>

      <main class="content">
        <RouterView />
      </main>
    </body>
  </div>
</template>

<style scoped>
.app {
  height: 100vh;
  width: 100%;
}

.header {
  display: flex;
  height: 100px;
  border-bottom: 1px solid grey;
  padding: 1rem;
}

.body {
  display: flex;
}

.sidebar {
  width: 120px;
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
  height: 100%;
  width: auto;
  vertical-align: middle;
  padding: 0.2rem;
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
