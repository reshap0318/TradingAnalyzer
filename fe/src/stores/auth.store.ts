import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { post } from '@/lib/axios'
import type { IApiResponse } from '@/interfaces/common'
import type { IUser, ILoginResponse } from '@/interfaces/auth'
import { setToken, removeToken, getToken } from '@/lib/axios'
import { showSuccess, showError } from '@/lib/sweetalert'

const BASE_URL = '/auth'

/**
 * Authentication Store
 * Manages user authentication state and actions
 *
 * NOTE: Only includes actions for endpoints that exist in backend API
 */
export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<IUser | null>(null)
  const loading = ref(false)
  const error = ref<string | undefined>()
  const isAuthenticated = computed(() => !!getToken())

  // Actions
  /**
   * Login user
   */
  async function loginAction(email: string, password: string, remember = false): Promise<boolean> {
    loading.value = true
    error.value = undefined

    try {
      const response = await post<IApiResponse<ILoginResponse>>(`${BASE_URL}/login`, { email, password })

      // Save token to localStorage
      setToken(response.data.data.token)

      // Set user data if returned
      if (response.data.data.user) {
        user.value = response.data.data.user
      }

      // Optionally persist for "remember me"
      if (remember) {
        localStorage.setItem('remember_email', email)
      } else {
        localStorage.removeItem('remember_email')
      }

      showSuccess(
        'Welcome back!',
        `Hello, ${response.data.data.user?.name || response.data.data.user?.email || 'User'}`
      )
      return true
    } catch (err: any) {
      error.value = err.response?.data?.message || 'Login failed. Please check your credentials.'
      showError('Login Failed', error.value)
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * Logout user
   */
  async function logoutAction(): Promise<void> {
    try {
      // Try to call logout API (if it exists)
      await post<IApiResponse<void>>(`${BASE_URL}/logout`).catch(() => {
        // Ignore if endpoint doesn't exist
        console.warn('Logout endpoint not available, skipping API call')
      })
    } finally {
      // Always clear local state
      removeToken()
      user.value = null
      showSuccess('Logged out', 'You have been successfully logged out.')
    }
  }

  /**
   * Clear user state (for cleanup)
   */
  function clearUserState(): void {
    user.value = null
    error.value = undefined
    loading.value = false
  }

  return {
    // State
    user,
    loading,
    error,
    isAuthenticated,

    // Getters (computed)
    userEmail: computed(() => user.value?.email ?? ''),
    userName: computed(() => user.value?.name ?? user.value?.email ?? 'User'),
    userId: computed(() => user.value?.id ?? null),

    // Actions
    loginAction,
    logoutAction,
    clearUserState
  }
})
