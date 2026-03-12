import {
  createRouter,
  createWebHistory,
  type RouteRecordRaw,
  type RouteLocationNormalized
} from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/pages/HomePage.vue')
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

// Navigation guards
router.beforeEach((to: RouteLocationNormalized, _from: RouteLocationNormalized, next: any) => {
  const token = localStorage.getItem('auth_token')
  const isAuthenticated = !!token

  // Check if route is for guests only (login, register)
  if (to.meta.guest && isAuthenticated) {
    next({ name: 'home' })
  }

  // Check if route requires authentication (add meta: { requiresAuth: true } to routes)
  else if (to.matched.some((record) => record.meta.requiresAuth) && !isAuthenticated) {
    next({ name: 'login' })
  }

  // Otherwise, allow navigation
  else {
    next()
  }
})

export default router
