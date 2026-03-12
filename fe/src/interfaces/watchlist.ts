/**
 * Watchlist Module Interfaces
 */

/**
 * Watchlist entity
 */
export interface IWatchlist {
  id: number
  symbol: string
  is_active: boolean
  created_at: string
  updated_at?: string
}

/**
 * Watchlist creation request
 */
export interface IWatchlistRequest {
  symbol: string
  is_active?: boolean
}

/**
 * Watchlist update request (all fields optional)
 */
export interface IWatchlistUpdateRequest extends Partial<IWatchlistRequest> {}

/**
 * Watchlist query parameters
 */
export interface IWatchlistQueryParams {
  is_active?: boolean
  search?: string
  sort?: string
  order?: 'asc' | 'desc'
}

/**
 * Watchlist filters for store
 */
export interface IWatchlistFilters {
  isActive?: boolean
  searchQuery?: string
}
