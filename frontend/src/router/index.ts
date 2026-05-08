import { createRouter, createWebHistory } from 'vue-router'
import LibraryView from '@/views/LibraryView.vue'
import LoginView from '@/views/LoginView.vue'

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
    }
  ],
});

router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth) {
    const token = localStorage.getItem('token');
    if (token) {
      next(); // Proceed to route if user is authenticated 
    } else {
      next('/login'); // Redirect to login if not authenticated
    }
  } else {
    next(); // Proceed to route if it's not protected
  }
});

export default router
