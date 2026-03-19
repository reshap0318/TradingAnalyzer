/**
 * Trade Helper Functions
 * Utility functions for trade-related calculations and formatting
 */

import type { ITradeOrder } from '@/stores/tradebot.store'

/**
 * Get filled orders count
 */
export const getFilledOrdersCount = (orders: ITradeOrder[]): number => {
  return orders?.filter(o => o.status === 'FILLED').length || 0
}

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
 * Get side color class
 */
export const getSideColor = (side: string): string => {
  return side.toUpperCase() === 'BUY' || side.toUpperCase() === 'LONG' ? 'text-green-600' : 'text-red-600'
}

/**
 * Get side background color class
 */
export const getSideBgColor = (side: string): string => {
  return side.toUpperCase() === 'BUY' || side.toUpperCase() === 'LONG' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
}

/**
 * Get status color class
 */
export const getStatusColor = (status: string): string => {
  switch (status.toUpperCase()) {
    case 'COMPLETED':
      return 'text-green-600 bg-green-50'
    case 'ACTIVE':
      return 'text-blue-600 bg-blue-50'
    case 'CANCELLED':
      return 'text-gray-100 bg-gray-800'
    default:
      return 'text-orange-600 bg-orange-50'
  }
}

/**
 * Get TP/SL status badge text
 */
export const getTpSlStatusBadge = (status: string): string => {
  switch (status?.toUpperCase()) {
    case 'TP_HIT':
      return 'Take Profit Hit'
    case 'SL_HIT':
      return 'Stop Loss Hit'
    case 'ACTIVE':
      return 'Active'
    default:
      return status || '-'
  }
}

/**
 * Get TP/SL status color class
 */
export const getTpSlStatusColor = (status: string): string => {
  status = status?.toUpperCase();
  if (status.includes("TP_HIT")) return 'text-green-600';
  else if (status.includes("TP_HIT")) return 'text-red-600';
  else return 'text-gray-600'
}

/**
 * Get PnL color class
 */
export const getPnLColor = (pnl: number): string => {
  if (pnl > 0) return 'text-green-600'
  if (pnl < 0) return 'text-red-600'
  return 'text-gray-600'
}

/**
 * Get entry mode badge text
 */
export const getEntryModeBadge = (isAggressive: boolean, orders: ITradeOrder[]): string => {
  const filledCount = getFilledOrdersCount(orders)
  if (!isAggressive) return 'Conservative (1 Entry)'
  return filledCount === 1 ? 'Aggressive (1/2 Filled)' : 'Aggressive (2/2 Filled)'
}

/**
 * Get entry mode color class
 */
export const getEntryModeColor = (isAggressive: boolean, orders: ITradeOrder[]): string => {
  const filledCount = getFilledOrdersCount(orders)
  if (!isAggressive) return 'bg-blue-100 text-blue-700'
  if (filledCount === 1) return 'bg-orange-100 text-orange-700'
  return 'bg-purple-100 text-purple-700'
}
