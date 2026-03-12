import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { type IApiResponse, post } from '@/lib/axios'
import { showSuccess, showError } from '@/lib/sweetalert'
import { destroySession, getToken, setToken } from '@/lib/storage'
import useVuelidate from '@vuelidate/core'
import { required } from '@vuelidate/validators'

const BASE_URL = '/auth'

export interface ILoginRequest {
  username: string
  password: string
}

export interface IUser {
  name: string
  token: string
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
  const loginReq = ref<ILoginRequest>({
    username: '',
    password: ''
  })
  const loginRules = ref({
    username: { required },
    password: { required }
  })
  const loginReqValid = useVuelidate(loginRules, loginReq, {})

  // Getters
  const isAuthenticated = computed(() => !!token.value)

  async function loginAction(): Promise<boolean> {
    loading.value = true

    const valid = await loginReqValid.value.$validate()
    if (!valid) return false

    try {
      const response = await post<IApiResponse<IUser>>(`${BASE_URL}/login`, loginReq.value)
      const data = response.data.data
      // Save token to store and localStorage
      setToken(data.token)

      showSuccess('Welcome back!', `Hello, ${data.name || 'User'}`)
      return true
    } catch (err: any) {
      const error = err.response?.data?.message || 'Login failed. Please check your credentials.'
      showError('Login Failed', error)
      return false
    } finally {
      loading.value = false
    }
  }

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
    loginReq,

    // Validation
    loginReqValid,

    // Getters
    isAuthenticated,

    // Actions
    loginAction,
    logoutAction
  }
})
