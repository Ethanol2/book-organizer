import { createRouter, createWebHistory } from 'vue-router'
import LibraryView from '@/views/LibraryView.vue'
import LoginView from '@/views/LoginView.vue'
import { useAuthStore } from '@/stores/auth';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: LibraryView,
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView
    },
    {
      path: '/about',
      name: 'about',
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import('../views/AboutView.vue'),
    },
    {
      path: '/add-book',
      name: 'add-book',
      component: () => import('../views/AddBookView.vue'),
    },
    {
      path: '/downloads',
      name: 'downloads',
      component: () => import('../views/DownloadsView.vue'),
    },
    {
      path: '/books/:id',
      name: 'book-details',
      component: () => import('../views/BookDetailsView.vue'),
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../views/SettingsView.vue'),
    }
  ],
});

router.beforeEach( async(to, from) => {
  const authStore = useAuthStore();

  if (!authStore.initialized) {
    await authStore.checkCurrentStatus();
  }

  if (!authStore.isAuthenticated) {
    if (to.name !== 'login' && to.name !== 'about') {
      return { name: 'login' };
    }
  }
  else if (to.name === 'login') {
    return { name: 'home' };
  }

  return true;
});

export default router
