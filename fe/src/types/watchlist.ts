/**
 * Watchlist Module Types
 */

import type { IWatchlist, IWatchlistRequest } from '@/interfaces/watchlist'

/**
 * Watchlist creation input (excludes id, timestamps)
 */
export type TWatchlistCreateInput = Omit<IWatchlist, 'id' | 'created_at' | 'updated_at'>

/**
 * Watchlist update input (partial, excludes id, timestamps)
 */
export type TWatchlistUpdateInput = Partial<TWatchlistCreateInput>

/**
 * Watchlist form data (for UI forms)
 */
export type TWatchlistFormData = IWatchlistRequest & {
  confirm?: boolean
}

/**
 * Watchlist table row (extended for UI)
 */
export type TWatchlistTableRow = IWatchlist & {
  isSelected?: boolean
  isEditing?: boolean
}
