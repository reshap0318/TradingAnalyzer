<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'
import { PhList, PhSignOut } from '@phosphor-icons/vue'

const router = useRouter()
const authStore = useAuthStore()

const user = computed(() => authStore.user)

const emit = defineEmits<{
  'menu-toggle': []
}>()

const handleLogout = async () => {
  await authStore.logoutAction()
  router.push('/login')
}
</script>

<template>
  <header class="fixed top-0 left-64 right-0 h-16 bg-white border-b border-gray-200 flex items-center justify-between px-8 z-30">
    <!-- Left: Hamburger + Page Title -->
    <div class="flex items-center gap-4">
      <button
        @click="emit('menu-toggle')"
        class="hidden md:flex items-center justify-center p-2 bg-transparent border-none cursor-pointer text-gray-500 rounded-lg hover:bg-gray-100 hover:text-gray-900 transition-all"
        aria-label="Toggle menu"
      >
        <PhList :size="24" />
      </button>
      <h2 class="text-xl font-semibold text-gray-900">
        <slot name="title">Dashboard</slot>
      </h2>
    </div>

    <!-- Right: User Info + Logout -->
    <div class="flex items-center gap-4">
      <div v-if="user" class="flex items-center gap-3 px-4 py-2 bg-gray-50 rounded-xl">
        <div class="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-600 text-white rounded-full flex items-center justify-center font-semibold text-sm">
          {{ user.name.charAt(0).toUpperCase() }}
        </div>
        <span class="text-sm font-medium text-gray-700 hidden sm:inline">{{ user.name }}</span>
      </div>
      <button
        @click="handleLogout"
        class="flex items-center gap-2 px-4 py-2 bg-transparent border border-gray-200 rounded-xl text-red-600 hover:bg-red-50 hover:border-red-600 transition-all text-sm font-medium"
        aria-label="Logout"
      >
        <PhSignOut :size="20" />
        <span class="hidden sm:inline">Logout</span>
      </button>
    </div>
  </header>
</template>
