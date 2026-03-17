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
  limit: number
}

const DEFAULT_FILTER: ITradeFilter = {
  status: [],
  symbol: [],
  interval: '',
  side: '',
  min_confidence: 0,
  date_start: '',
  date_end: '',
  limit: 100
}

export const useTradeStore = defineStore('trade', () => {
  // State
  const trades = ref<ITrade[]>([])
  const loading = ref(false)
  const filter = ref<ITradeFilter>({ ...DEFAULT_FILTER })

  // Actions

  // Fetch trades from backend with current filter state as query params
  // This is the SINGLE SOURCE OF TRUTH — no local filtering needed
  async function fetchTrades() {
    loading.value = true
    try {
      // Build query params from current filter state
      const params = new URLSearchParams()

      filter.value.status.forEach(status => {
        params.append('status', status)
      })

      filter.value.symbol.forEach(symbol => {
        params.append('symbol', symbol)
      })

      if (filter.value.interval) {
        params.append('interval', filter.value.interval)
      }

      if (filter.value.side) {
        params.append('side', filter.value.side)
      }

      if (filter.value.min_confidence > 0) {
        params.append('min_confidence', filter.value.min_confidence.toString())
      }

      if (filter.value.date_start) {
        params.append('date_start', filter.value.date_start)
      }
      if (filter.value.date_end) {
        params.append('date_end', filter.value.date_end)
      }

      // Always send limit
      if (filter.value.limit > 0) {
        params.append('limit', filter.value.limit.toString())
      }

      const queryString = params.toString()
      const url = queryString ? `${BASE_URL}?${queryString}` : BASE_URL + "?"

      const response = await get<IApiResponse<ITrade[]>>(url)
      trades.value = response.data.data || []
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to fetch trades')
      trades.value = []
    } finally {
      loading.value = false
    }
  }

  function updateFilter(newFilter: Partial<ITradeFilter>) {
    filter.value = { ...filter.value, ...newFilter }
  }

  function resetFilter() {
    filter.value = { ...DEFAULT_FILTER }
  }

  return {
    // State
    trades,
    filter,
    loading,

    // Actions
    fetchTrades,
    updateFilter,
    resetFilter
  }
})
