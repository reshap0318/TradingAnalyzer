import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { type IApiResponse, post } from '@/lib/axios'
import { showSuccess, showError } from '@/lib/sweetalert'
import { destroySession, getToken, setToken } from '@/lib/storage'

const BASE_URL = '/auth'

interface ILoginRequest {
  email: string
  password: string
}

export interface ILoginResponse {
  token: string
  user?: IUser
}

export interface IUser {
  id: number
  email: string
  name?: string
  created_at?: string
}

/**
 * Authentication Store
 * Manages user authentication state and actions
 * Centralized token management (no need to reload page)
 */
export const useAuthStore = defineStore('auth', () => {
  // State
  const loading = ref(false)
  const user = ref<IUser | null>(null)
  const token = ref<string | null>(getToken())

  // Getters
  const isAuthenticated = computed(() => !!token.value)

  async function loginAction(param: ILoginRequest): Promise<boolean> {
    loading.value = true

    try {
      const response = await post<IApiResponse<ILoginResponse>>(`${BASE_URL}/login`, param)

      // Save token to store and localStorage
      setToken(response.data.data.token)

      // Set user data if returned
      if (response.data.data.user) {
        user.value = response.data.data.user
      }

      showSuccess(
        'Welcome back!',
        `Hello, ${response.data.data.user?.name || response.data.data.user?.email || 'User'}`
      )
      return true
    } catch (err: any) {
      const error = err.response?.data?.message || 'Login failed. Please check your credentials.'
      showError('Login Failed', error)
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
        console.warn('Logout endpoint not available, skipping API call')
      })
    } finally {
      // Always clear local state
      destroySession()
      user.value = null
      token.value = null
      showSuccess('Logged out', 'You have been successfully logged out.')
    }
  }

  return {
    // State
    user,
    token,
    loading,

    // Getters (computed)
    isAuthenticated,

    // Actions
    loginAction,
    logoutAction
  }
})
