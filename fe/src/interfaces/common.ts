/**
 * Common/Generic Interfaces
 * Reusable interfaces used across the application
 */

/**
 * Standard API response structure
 */
export interface IApiResponse<TData> {
  code: number
  message: string
  data: TData
}

/**
 * Paginated response for list endpoints
 */
export interface IPaginatedResponse<TData> extends IApiResponse<TData[]> {
  pagination: {
    page: number
    limit: number
    total: number
    totalPages: number
  }
}

/**
 * Pagination request parameters
 */
export interface IPaginationParams {
  page?: number
  limit?: number
  sort?: string
  order?: 'asc' | 'desc'
}

/**
 * Response metadata
 */
export interface IMeta {
  timestamp: string
  version: string
  requestId?: string
}

/**
 * API response with metadata
 */
export interface IApiResponseWithMeta<TData> extends IApiResponse<TData> {
  meta: IMeta
}
