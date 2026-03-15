/**
 * Formatter Helper Functions
 * General utility functions for formatting values
 */

/**
 * Format price with proper decimals based on value
 */
export const formatPrice = (price: number): string => {
  if (price === 0) return '-'
  if (price >= 1000) return price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  if (price >= 1) return price.toFixed(4)
  return price.toFixed(6)
}

/**
 * Format PnL with sign
 */
export const formatPnL = (pnl: number): string => {
  const sign = pnl >= 0 ? '+' : ''
  return `${sign}${pnl.toFixed(2)} USDT (${sign}${pnl.toFixed(2)}%)`
}

/**
 * Format percentage with fixed decimals
 */
export const formatPercent = (value: number, decimals: number = 2): string => {
  return `${value.toFixed(decimals)}%`
}

/**
 * Format number with commas
 */
export const formatNumber = (value: number, decimals: number = 2): string => {
  return value.toLocaleString('en-US', { minimumFractionDigits: decimals, maximumFractionDigits: decimals })
}

/**
 * Format date to locale string
 */
export const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleString()
}

/**
 * Format time to locale string
 */
export const formatTime = (dateString: string): string => {
  return new Date(dateString).toLocaleTimeString()
}

/**
 * Format duration in seconds to human readable
 */
export const formatDuration = (seconds: number): string => {
  if (seconds < 60) return `${Math.floor(seconds)}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${Math.floor(seconds % 60)}s`
  if (seconds < 86400) {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    return `${hours}h ${minutes}m`
  }
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  return `${days}d ${hours}h ${minutes}m`
}
