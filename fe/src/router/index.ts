import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: () => import('@/pages/DashboardPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Dashboard'
    }
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/pages/LoginPage.vue'),
    meta: { guest: true } // Only for non-authenticated users
  },
  {
    path: '/timeframes',
    name: 'timeframes',
    component: () => import('@/pages/TimeframeListPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Timeframes'
    }
  },
  {
    path: '/thresholds',
    name: 'thresholds',
    component: () => import('@/pages/ThresholdListPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Thresholds'
    }
  },
  {
    path: '/configs',
    name: 'configs',
    component: () => import('@/pages/ConfigListPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Configs'
    }
  },
  {
    path: '/indicators',
    name: 'indicators',
    component: () => import('@/pages/IndicatorListPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Indicators'
    }
  },
  {
    path: '/watchlists',
    name: 'watchlists',
    component: () => import('@/pages/WatchlistListPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Watchlists'
    }
  },
  {
    path: '/bot-control',
    name: 'bot-control',
    component: () => import('@/pages/BotControlPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Bot Control'
    }
  },
  {
    path: '/trades',
    name: 'trades',
    component: () => import('@/pages/TradesPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Trade History'
    }
  },
  {
    path: '/strategies',
    name: 'strategies',
    component: () => import('@/pages/StrategiesPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Strategies'
    }
  },
  {
    path: '/signal-analyze',
    name: 'signal-analyze',
    component: () => import('@/pages/SignalAnalyzePage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Signal Analyze'
    }
  },
  {
    path: '/backtest',
    name: 'backtest',
    component: () => import('@/pages/BacktestPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Backtest'
    }
  },
  {
    path: '/backtest/:id',
    name: 'backtest-detail',
    component: () => import('@/pages/BacktestDetailPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Backtest Detail'
    }
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
