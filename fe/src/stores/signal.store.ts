import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { get, del, post } from '@/lib/axios'
import type { IApiResponse } from '@/lib/axios'
import { showSuccess, showError, showConfirm } from '@/lib/sweetalert'

// ============== Types ==============

export interface ISignal {
  id: number
  symbol: string
  strategy_id: number
  signal_category: 'BUY' | 'SELL' | 'STRONG_BUY' | 'STRONG_SELL' | 'WAIT'
  signal_valid: boolean
  total_score: number
  confidence: number
  current_price: number
  primary_timeframe: string
  tp_price: number
  sl_price: number
  support_price: number
  resistance_price: number
  risk_reward_ratio: number
  avg_entry_price: number
  entry_mode: 'CONSERVATIVE' | 'AGGRESSIVE'
  trading_capital: number
  total_position_value: number
  total_position_qty: number
  max_risk_usdt: number
  max_risk_percent: number
  target_profit_usdt: number
  target_profit_percent: number
  effective_leverage: number
  leverage: number
  buffer_percent: number
  created_at: string
  updated_at: string
  strategy_snapshot?: {
    name?: string
    primary_timeframe?: string
    timeframes?: any[]
    indicator_weights?: any[]
    mm_config?: Record<string, any>
  }
  ohlc_snapshot?: {
    timeframe?: string
    candles?: any[]
  }
  indicator_values?: any[]
  entry_levels?: any[]
}

export interface ISignalIndexRequest {
  symbol?: string
  strategy_id?: number
  signal_category?: string
  signal_valid?: boolean
  start_time?: string
  end_time?: string
  page?: number
  page_size?: number
}

export interface ISignalIndexResponse {
  signals: ISignal[]
  pagination: {
    page: number
    page_size: number
    total_items: number
    total_pages: number
  }
}

export interface ISignalCleanupResponse {
  deleted_count: number
  older_than_hours: number
  message: string
}

export const useSignalStore = defineStore('signal', () => {
  // ============== State ==============

  const signals = ref<ISignal[]>([])
  const currentSignal = ref<ISignal | null>(null)
  const loading = ref(false)
  const pagination = ref({
    page: 1,
    page_size: 12,
    total_items: 0,
    total_pages: 0
  })

  // Filters
  const filters = ref<ISignalIndexRequest>({
    symbol: '',
    strategy_id: undefined,
    signal_category: '',
    signal_valid: undefined,
    start_time: '',
    end_time: '',
    page: 1,
    page_size: 12
  })

  // ============== Computed ==============

  const totalPages = computed(() => pagination.value.total_pages)
  const currentPage = computed(() => pagination.value.page)
  const totalItems = computed(() => pagination.value.total_items)

  const validSignalsCount = computed(() => {
    return signals.value.filter((s) => s.signal_valid).length
  })

  const signalsByCategory = computed(() => {
    const categories: Record<string, number> = {}
    signals.value.forEach((signal) => {
      const category = signal.signal_category
      categories[category] = (categories[category] || 0) + 1
    })
    return categories
  })

  // ============== Actions ==============

  /**
   * Fetch signals with pagination and filters
   */
  async function fetchSignals(overrideFilters?: Partial<ISignalIndexRequest>) {
    loading.value = true
    try {
      const queryParams: ISignalIndexRequest = {
        ...filters.value,
        ...overrideFilters
      }

      const cleanParams = Object.fromEntries(
        Object.entries(queryParams).filter(([_, v]) => v !== '' && v !== undefined && v !== null)
      ) as ISignalIndexRequest

      const response = await get<IApiResponse<ISignalIndexResponse>>('/signal', {
        params: cleanParams
      })

      let backendData: ISignalIndexResponse | null = null
      backendData = response.data.data as unknown as ISignalIndexResponse

      if (backendData) {
        signals.value = backendData.signals || []
        pagination.value = backendData.pagination || {
          page: 1,
          page_size: 20,
          total_items: 0,
          total_pages: 0
        }
      }
    } catch (error: any) {
      showError('Error', error.message || 'Failed to fetch signals')
      signals.value = []
      pagination.value = {
        page: 1,
        page_size: 20,
        total_items: 0,
        total_pages: 0
      }
    } finally {
      loading.value = false
    }
  }

  /**
   * Fetch signal detail by ID
   */
  async function fetchSignalDetail(id: number) {
    loading.value = true
    try {
      const response = await get<IApiResponse<ISignal>>(`/signal/${id}`)
      currentSignal.value = response.data.data
    } catch (error: any) {
      showError('Error', error.message || 'Failed to fetch signal detail')
      currentSignal.value = null
      throw error
    } finally {
      loading.value = false
    }
  }

  /**
   * Delete signal by ID
   */
  async function deleteSignalById(id: number): Promise<boolean> {
    try {
      const confirmed = await showConfirm(
        'Delete Signal',
        'Are you sure you want to delete this signal? This action cannot be undone.'
      )

      if (!confirmed) return false

      await del<IApiResponse<null>>(`/signal/${id}`)
      showSuccess('Success', 'Signal deleted successfully')

      // Refresh list
      await fetchSignals()

      return true
    } catch (error: any) {
      showError('Error', error.message || 'Failed to delete signal')
      return false
    }
  }

  /**
   * Cleanup old signals
   */
  async function cleanupOldSignals(olderThanHours: number = 720): Promise<boolean> {
    try {
      const confirmed = await showConfirm(
        'Cleanup Old Signals',
        `Are you sure you want to delete signals older than ${olderThanHours} hours? This action cannot be undone.`
      )

      if (!confirmed) return false

      loading.value = true
      const response = await post<IApiResponse<ISignalCleanupResponse>>('/signal/cleanup', {
        older_than_hours: olderThanHours
      })

      if (response.data.data) {
        showSuccess(
          'Cleanup Complete',
          `Successfully deleted ${response.data.data.deleted_count} old signals`
        )

        // Refresh list
        await fetchSignals()
      }

      return true
    } catch (error: any) {
      showError('Error', error.message || 'Failed to cleanup signals')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * Update filters and fetch signals
   */
  async function updateFiltersAndFetch(newFilters: Partial<ISignalIndexRequest>) {
    filters.value = {
      ...filters.value,
      ...newFilters,
      page: 1 // Reset to first page when filters change
    }
    await fetchSignals()
  }

  /**
   * Change page and fetch signals
   */
  async function changePage(page: number) {
    filters.value.page = page
    await fetchSignals()
  }

  /**
   * Clear current signal
   */
  function clearCurrentSignal() {
    currentSignal.value = null
  }

  /**
   * Refresh signals list
   */
  async function refreshSignals() {
    await fetchSignals()
  }

  // ============== Return ==============

  return {
    // State
    signals,
    currentSignal,
    loading,
    pagination,
    filters,

    // Computed
    totalPages,
    currentPage,
    totalItems,
    validSignalsCount,
    signalsByCategory,

    // Actions
    fetchSignals,
    fetchSignalDetail,
    deleteSignalById,
    cleanupOldSignals,
    updateFiltersAndFetch,
    changePage,
    clearCurrentSignal,
    refreshSignals
  }
})
