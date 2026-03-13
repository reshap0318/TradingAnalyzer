import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/pages/HomePage.vue'),
    meta: {
      requiresAuth: true
    }
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/pages/LoginPage.vue'),
    meta: { guest: true } // Only for non-authenticated users
  }
  // Add more routes here
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guards - Vue Router 5 pattern (return instead of next())
router.beforeEach((to) => {
  const token = localStorage.getItem('auth_token')
  const isAuthenticated = !!token

  // Check if route is for guests only (login, register)
  if (to.meta.guest && isAuthenticated) {
    return { name: 'home' }
  }

  // Check if route requires authentication
  if (to.matched.some((record) => record.meta.requiresAuth) && !isAuthenticated) {
    return { name: 'login' }
  }

  // Allow navigation
  return true
})

export default router
