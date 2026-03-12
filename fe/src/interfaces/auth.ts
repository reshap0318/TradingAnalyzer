/**
 * Authentication Module Interfaces
 */

/**
 * User login request
 */
export interface ILoginRequest {
  email: string
  password: string
}

/**
 * Login response with token
 */
export interface ILoginResponse {
  token: string
  user?: IUser
}

/**
 * User information (if returned by backend)
 */
export interface IUser {
  id: number
  email: string
  name?: string
  created_at?: string
}
