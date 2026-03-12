/**
 * Common Type Aliases & Union Types
 */

/**
 * HTTP Methods
 */
export type THttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

/**
 * Generic ID type
 */
export type TId = number | string

/**
 * Timestamp string (ISO 8601)
 */
export type TTimestamp = string

/**
 * UUID format
 */
export type TUuid = `${string}-${string}-${string}-${string}-${string}`

/**
 * Percentage value (0-100)
 */
export type TPercentage = number

/**
 * Status types
 */
export type TStatus = 'active' | 'inactive' | 'pending' | 'error'

/**
 * Sort order
 */
export type TSortOrder = 'asc' | 'desc'

/**
 * Nullable type helper
 */
export type TNullable<T> = {
  [K in keyof T]: T[K] | null
}

/**
 * Partial type helper (all fields optional)
 */
export type TPartial<T> = {
  [K in keyof T]?: T[K]
}

/**
 * Readonly type helper
 */
export type TReadonly<T> = {
  readonly [K in keyof T]: T[K]
}
