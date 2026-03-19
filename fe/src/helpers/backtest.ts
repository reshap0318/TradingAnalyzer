/**
 * Backtest Helpers
 * Utility functions for backtest data formatting and calculations
 */

import type { IBacktestTrade } from '@/stores/backtest.store'

/**
 * Format currency value
 */
export function formatCurrency(value: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2
  }).format(value)
}

/**
 * Format percentage value
 */
export function formatPercent(value: number): string {
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`
}

/**
 * Format date/timestamp to readable string
 * @param timestamp - Unix timestamp or date string
 * @returns Formatted date string
 */
export function formatDate(timestamp: number | string): string {
  return new Date(timestamp).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

/**
 * Get side color class
 */
export function getSideColor(side: string): string {
  return side === 'BUY' ? 'text-green-600 bg-green-50' : 'text-red-600 bg-red-50'
}

/**
 * Get trade status color class
 */
export function getTradeStatusColor(status: string): string {
  switch (status) {
    case 'CLOSED':
      return 'bg-green-100 text-green-700'
    case 'ACTIVE':
      return 'bg-blue-100 text-blue-700'
    case 'CANCELLED':
      return 'bg-gray-100 text-gray-700'
    case 'EXPIRED':
      return 'bg-yellow-100 text-yellow-700'
    default:
      return 'bg-gray-100 text-gray-700'
  }
}

/**
 * Calculate trade statistics
 */
export function calculateTradeStats(trades: IBacktestTrade[]) {
  return {
    total: trades.length,
    closed: trades.filter(t => t.status === 'CLOSED').length,
    active: trades.filter(t => t.status === 'ACTIVE').length,
    cancelled: trades.filter(t => t.status === 'CANCELLED').length,
    expired: trades.filter(t => t.status === 'EXPIRED').length,
    winRate: trades.length > 0 
      ? (trades.filter(t => t.pnl > 0).length / trades.length) * 100 
      : 0
  }
}
