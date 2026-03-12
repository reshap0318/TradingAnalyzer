/**
 * TailwindCSS v4 Configuration
 * 
 * Tailwind v4 menggunakan CSS-based configuration via @theme directive
 * Lihat: src/style.css untuk custom theme
 */

// Re-export theme colors untuk penggunaan di TypeScript
export const colors = {
  primary: {
    DEFAULT: '#3b82f6',
    dark: '#2563eb',
    light: '#60a5fa',
  },
  success: {
    DEFAULT: '#22c55e',
    dark: '#16a34a',
    light: '#4ade80',
  },
  danger: {
    DEFAULT: '#ef4444',
    dark: '#dc2626',
    light: '#f87171',
  },
  warning: {
    DEFAULT: '#f59e0b',
    dark: '#d97706',
    light: '#fbbf24',
  },
  info: {
    DEFAULT: '#06b6d4',
    dark: '#0891b2',
    light: '#22d3ee',
  },
  signal: {
    strongBuy: '#16a34a',
    buy: '#22c55e',
    wait: '#6b7280',
    sell: '#ef4444',
    strongSell: '#dc2626',
  },
} as const

export type ThemeColor = keyof typeof colors
