import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import DownloadsView from '../views/DownloadsView.vue'
import LibraryView from '@/views/LibraryView.vue'
import BookDetailsView from '@/views/BookDetailsView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: LibraryView,
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
      path: '/downloads',
      name: 'downloads',
      component: DownloadsView
    },
    {
      path: '/books/:id',
      name: 'book-details',
      component: BookDetailsView
    }
  ],
})

export default router
