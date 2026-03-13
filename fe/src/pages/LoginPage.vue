<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'
import { UiInput, UiButton, UiPassword } from '@/components/common'
import { getValidationErrors } from '@/helpers/validation'

const router = useRouter()
const authStore = useAuthStore()

// Computed
const isLoading = computed(() => authStore.loading)
const loginReq = authStore.loginReq
const v$ = authStore.loginReqValid

// Methods
const handleSubmit = async () => {
  const success = await authStore.loginAction()
  if (success) {
    router.push('/')
  }
}

const handleKeyPress = (e: KeyboardEvent) => {
  if (e.key === 'Enter') {
    handleSubmit()
  }
}
</script>

<template>
  <div
    class="min-h-screen flex items-center justify-center bg-linear-to-br from-primary/10 via-primary/5 to-transparent"
  >
    <div class="w-full max-w-md p-8">
      <!-- Logo & Title -->
      <div class="text-center mb-8">
        <h1 class="text-4xl font-bold text-primary mb-2">TradingAnalyzer</h1>
        <p class="text-gray-600 dark:text-gray-400">Sign in to your account</p>
      </div>

      <!-- Login Card -->
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-8">
        <form @submit.prevent="handleSubmit" @keypress="handleKeyPress">
          <!-- Username Field -->
          <div class="mb-6">
            <UiInput
              v-model="loginReq.username"
              label="Username"
              placeholder="Enter your username"
              autocomplete="username"
              :error="v$.username.$error"
              :error-message="getValidationErrors(v$.username).join(',')"
            />
          </div>

          <!-- Password Field -->
          <div class="mb-6">
            <UiPassword
              v-model="loginReq.password"
              label="Password"
              placeholder="Enter your password"
              :error="v$.password.$error"
              :error-message="getValidationErrors(v$.password).join(',')"
            />
          </div>

          <!-- Submit Button -->
          <UiButton type="submit" variant="primary" :loading="isLoading" full-width>
            {{ isLoading ? 'Signing in...' : 'Sign In' }}
          </UiButton>
        </form>
      </div>

      <!-- Footer -->
      <p class="text-center mt-8 text-sm text-gray-500 dark:text-gray-400">
        © 2025 TradingAnalyzer. All rights reserved.
      </p>
    </div>
  </div>
</template>
