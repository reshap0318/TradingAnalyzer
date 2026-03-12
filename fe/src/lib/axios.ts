import axios, {
  type AxiosInstance,
  type AxiosRequestConfig,
  type AxiosResponse,
  type InternalAxiosRequestConfig
} from 'axios'

// Storage key for auth token
const TOKEN_KEY = 'auth_token'

// Base API URL from environment or default
const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8000/api'

// Request timeout in milliseconds
const TIMEOUT = 30000

/**
 * Create axios instance with default config
 */
const apiClient: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: TIMEOUT,
  headers: {
    'Content-Type': 'application/json'
  }
})

/**
 * Get token from storage
 */
export const getToken = (): string | null => {
  if (typeof window === 'undefined') return null
  return localStorage.getItem(TOKEN_KEY)
}

/**
 * Set token to storage
 */
export const setToken = (token: string): void => {
  if (typeof window === 'undefined') return
  localStorage.setItem(TOKEN_KEY, token)
}

/**
 * Remove token from storage
 */
export const removeToken = (): void => {
  if (typeof window === 'undefined') return
  localStorage.removeItem(TOKEN_KEY)
}

/**
 * Request interceptor - Add Authorization header
 */
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getToken()

    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

/**
 * Response interceptor - Handle common errors
 */
apiClient.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error) => {
    // Handle 401 Unauthorized
    if (error.response?.status === 401) {
      // Optional: Auto logout or redirect to login
      // removeToken()
      // window.location.href = '/login'
      console.error('Unauthorized - Token may be expired')
    }

    // Handle 403 Forbidden
    if (error.response?.status === 403) {
      console.error('Forbidden - Insufficient permissions')
    }

    // Handle 500 Server Error
    if (error.response?.status === 500) {
      console.error('Server Error - Please try again later')
    }

    return Promise.reject(error)
  }
)

/**
 * Custom request method with simplified API
 */
export const request = async <T = any>(config: AxiosRequestConfig): Promise<AxiosResponse<T>> => {
  return apiClient.request<T>(config)
}

/**
 * GET request
 */
export const get = <T = any>(
  url: string,
  config?: AxiosRequestConfig
): Promise<AxiosResponse<T>> => {
  return apiClient.get<T>(url, config)
}

/**
 * POST request
 */
export const post = <T = any>(
  url: string,
  data?: any,
  config?: AxiosRequestConfig
): Promise<AxiosResponse<T>> => {
  return apiClient.post<T>(url, data, config)
}

/**
 * PUT request
 */
export const put = <T = any>(
  url: string,
  data?: any,
  config?: AxiosRequestConfig
): Promise<AxiosResponse<T>> => {
  return apiClient.put<T>(url, data, config)
}

/**
 * PATCH request
 */
export const patch = <T = any>(
  url: string,
  data?: any,
  config?: AxiosRequestConfig
): Promise<AxiosResponse<T>> => {
  return apiClient.patch<T>(url, data, config)
}

/**
 * DELETE request
 */
export const del = <T = any>(
  url: string,
  config?: AxiosRequestConfig
): Promise<AxiosResponse<T>> => {
  return apiClient.delete<T>(url, config)
}

/**
 * Set custom base URL (for runtime changes)
 */
export const setBaseURL = (url: string): void => {
  apiClient.defaults.baseURL = url
}

/**
 * Set custom timeout (for runtime changes)
 */
export const setTimeout = (timeout: number): void => {
  apiClient.defaults.timeout = timeout
}

export default apiClient
