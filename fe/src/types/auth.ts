/**
 * Authentication Module Types
 */

import type { ILoginRequest } from '@/interfaces/auth'

/**
 * Login form data (with UI helpers)
 */
export type TLoginFormData = ILoginRequest & {
  remember?: boolean
}

/**
 * Auth token storage
 */
export type TAuthToken = string
