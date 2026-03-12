<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'

const router = useRouter()
const authStore = useAuthStore()

const user = computed(() => authStore.user)
const isAuthenticated = computed(() => authStore.isAuthenticated)

const handleLogout = async () => {
  await authStore.logoutAction()
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-primary/10 via-primary/5 to-transparent">
    <!-- Header -->
    <header class="bg-white dark:bg-gray-800 shadow-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex justify-between items-center">
        <h1 class="text-2xl font-bold text-primary">TradingAnalyzer</h1>
        <div class="flex items-center gap-4">
          <div v-if="user" class="text-sm text-gray-600 dark:text-gray-400">
            <span class="font-semibold">{{ user.name }}</span>
          </div>
          <button
            @click="handleLogout"
            class="px-4 py-2 text-sm font-semibold text-danger border border-danger rounded-lg hover:bg-danger hover:text-white transition-all"
          >
            Logout
          </button>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div v-if="isAuthenticated" class="text-center">
        <h2 class="text-4xl font-bold text-gray-900 dark:text-white mb-4">
          Welcome back, {{ user?.name }}! 🎉
        </h2>
        <p class="text-lg text-gray-600 dark:text-gray-400 mb-8">
          You are successfully logged in. Dashboard coming soon...
        </p>
        
        <!-- Feature placeholders -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mt-12">
          <div class="bg-white dark:bg-gray-800 p-6 rounded-xl shadow-lg">
            <div class="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Analytics</h3>
            <p class="text-sm text-gray-600 dark:text-gray-400">Real-time market analytics and insights</p>
          </div>

          <div class="bg-white dark:bg-gray-800 p-6 rounded-xl shadow-lg">
            <div class="w-12 h-12 bg-success/10 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-6 h-6 text-success" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Watchlists</h3>
            <p class="text-sm text-gray-600 dark:text-gray-400">Track your favorite trading pairs</p>
          </div>

          <div class="bg-white dark:bg-gray-800 p-6 rounded-xl shadow-lg">
            <div class="w-12 h-12 bg-info/10 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-6 h-6 text-info" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Auto Trading</h3>
            <p class="text-sm text-gray-600 dark:text-gray-400">Automated trading strategies</p>
          </div>
        </div>
      </div>

      <div v-else class="text-center">
        <p class="text-lg text-gray-600 dark:text-gray-400">Please login to continue</p>
        <button
          @click="router.push('/login')"
          class="mt-4 px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition-all"
        >
          Go to Login
        </button>
      </div>
    </main>
  </div>
</template>
