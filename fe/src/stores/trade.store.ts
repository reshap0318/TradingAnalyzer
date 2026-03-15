import { defineStore } from 'pinia'
import { ref } from 'vue'
import { type IApiResponse, get } from '@/lib/axios'
import { showError } from '@/lib/sweetalert'
import type { ITrade } from '@/stores/tradebot.store'

const BASE_URL = '/trade/bot'

// Filter interface
export interface ITradeFilter {
  status: string[]
  symbol: string[]
  interval: string
  side: string
  min_confidence: number
  date_start: string
  date_end: string
}

export const useTradeStore = defineStore('trade', () => {
  // State
  const trades = ref<ITrade[]>([])
  const loading = ref(false)
  const filter = ref<ITradeFilter>({
    status: [],
    symbol: [],
    interval: '',
    side: '',
    min_confidence: 0,
    date_start: '',
    date_end: ''
  })

  // Computed
  const filteredTrades = ref<ITrade[]>([])

  // Actions
  async function fetchTrades() {
    loading.value = true
    try {
      // Build query params
      const params = new URLSearchParams()

      // Add status filters (multi-select)
      filter.value.status.forEach(status => {
        params.append('status', status)
      })

      // Add symbol filters (multi-select)
      filter.value.symbol.forEach(symbol => {
        params.append('symbol', symbol)
      })

      // Add interval filter
      if (filter.value.interval) {
        params.append('interval', filter.value.interval)
      }

      // Add side filter
      if (filter.value.side) {
        params.append('side', filter.value.side)
      }

      // Add min confidence filter
      if (filter.value.min_confidence > 0) {
        params.append('min_confidence', filter.value.min_confidence.toString())
      }

      // Add date range filters
      if (filter.value.date_start) {
        params.append('date_start', filter.value.date_start)
      }
      if (filter.value.date_end) {
        params.append('date_end', filter.value.date_end)
      }

      // Build URL with query params
      const queryString = params.toString()
      const url = queryString ? `${BASE_URL}?${queryString}` : BASE_URL + "?"

      const response = await get<IApiResponse<ITrade[]>>(url)
      trades.value = response.data.data || []
      filteredTrades.value = trades.value
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to fetch trades')
      trades.value = []
      filteredTrades.value = []
    } finally {
      loading.value = false
    }
  }

  function updateFilter(newFilter: Partial<ITradeFilter>) {
    filter.value = { ...filter.value, ...newFilter }
  }

  function resetFilter() {
    filter.value = {
      status: [],
      symbol: [],
      interval: '',
      side: '',
      min_confidence: 0,
      date_start: '',
      date_end: ''
    }
  }

  function applyFilters() {
    let result = [...trades.value]

    // Filter by status
    if (filter.value.status.length > 0) {
      result = result.filter(trade =>
        filter.value.status.includes(trade.status.toUpperCase())
      )
    }

    // Filter by symbol
    if (filter.value.symbol.length > 0) {
      result = result.filter(trade =>
        filter.value.symbol.includes(trade.symbol)
      )
    }

    // Filter by interval
    if (filter.value.interval) {
      result = result.filter(trade => trade.interval === filter.value.interval)
    }

    // Filter by side
    if (filter.value.side) {
      result = result.filter(trade => trade.side.toUpperCase() === filter.value.side.toUpperCase())
    }

    // Filter by min confidence
    if (filter.value.min_confidence > 0) {
      result = result.filter(trade => trade.confidence >= filter.value.min_confidence)
    }

    // Filter by date range
    if (filter.value.date_start) {
      const startDate = new Date(filter.value.date_start)
      result = result.filter(trade => new Date(trade.created_at) >= startDate)
    }
    if (filter.value.date_end) {
      const endDate = new Date(filter.value.date_end)
      endDate.setHours(23, 59, 59, 999) // Include entire end date
      result = result.filter(trade => new Date(trade.created_at) <= endDate)
    }

    filteredTrades.value = result
  }

  return {
    // State
    trades,
    filteredTrades,
    filter,
    loading,

    // Actions
    fetchTrades,
    updateFilter,
    resetFilter,
    applyFilters
  }
})
