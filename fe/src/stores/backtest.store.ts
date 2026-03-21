import { defineStore } from 'pinia'
import { ref } from 'vue'
import { type IApiResponse, get, post, del } from '@/lib/axios'
import { showSuccess, showError, showConfirm } from '@/lib/sweetalert'

const BASE_URL = '/backtests'

// ==================== INTERFACES ====================

// Request interface
export interface ICreateBacktestRequest {
  name: string
  symbol: string
  strategy_id: number
  days: number
  capital: number
}

// Summary interface
export interface IBacktestSummary {
  initial_balance: number
  final_balance: number
  net_profit: number
  net_profit_percent: number
  win_rate_pct: number
  total_trades: number
  winning_trades: number
  losing_trades: number
  expired_trades: number
  cancelled_trades: number
  max_drawdown_pct: number
  profit_factor: number
  avg_win: number
  avg_loss: number
  largest_win: number
  largest_loss: number
}

// Equity curve point
export interface IEquityPoint {
  timestamp: number
  balance: number
  pnl: number
}

// OHLCV candle data
export interface ICandleData {
  timestamp: number
  open: number
  high: number
  low: number
  close: number
  volume: number
}

// Trade targets (TP/SL)
export interface ITradeTargets {
  tp_price: number
  sl_price: number
  ratio: number
}

// Trade entry
export interface ITradeEntry {
  entry_num: number
  type: 'MARKET' | 'LIMIT'
  price: number
  qty: number
  timestamp: number
  status: 'PENDING' | 'FILLED' | 'CANCELLED' | 'EXPIRED'
}

// Trade exit
export interface ITradeExit {
  reason: 'HIT_TP' | 'HIT_SL' | 'CLOSED_END' | 'DEAD_SIGNAL' | 'EXPIRED'
  price: number
  timestamp: number
}

// Daily stats snapshot
export interface IDailyStatsSnapshot {
  trade_count: number
  pnl: number
  consecutive_loss: number
}

// Backtest trade DTO
export interface IBacktestTrade {
  trade_id: number
  trade_num: number
  side: 'BUY' | 'SELL'
  signal: string
  confidence: number
  trading_mode: 'AGGRESSIVE' | 'CONSERVATIVE'
  status: 'ACTIVE' | 'CLOSED' | 'CANCELLED' | 'EXPIRED'
  targets: ITradeTargets
  entries: ITradeEntry[]
  exit: ITradeExit | null
  total_qty: number
  avg_entry_price: number
  total_capital: number
  pnl: number
  pnl_percent: number
  entry_time: string
  filled_time: string | null
  exit_time: string | null
  duration_minutes: number
  daily_stats: IDailyStatsSnapshot | null
}

// Strategy snapshot
export interface IStrategySnapshot {
  id: number
  strategy_name: string
  primary_tf: string
  is_active: boolean
  timeframes: IStrategyTimeframe[]
  indicator_weights: IStrategyIndicator[]
  money_management: IMoneyManagement
}

export interface IMoneyManagement {
  min_confidence: number
  max_daily_trades: number
  max_daily_loss_percent: number
  leverage: number
  is_agressive: boolean
  max_position_size?: number
  risk_reward_ratio?: number
  order_expiration_hours?: number
}

export interface IStrategyTimeframe {
  tf: string
  weight: number
  timeframe_detail?: {
    name: string
    in_minutes: number
  }
}

export interface IStrategyIndicator {
  indicator_id: number
  weight: number
  indicator_detail?: {
    id: number
    name: string
    indicator: string
  }
}

// Backtest summary (for list view)
export interface IBacktestSummaryItem {
  id: number
  name: string
  symbol: string
  strategy_name: string
  total_pnl: number
  total_pnl_percent: number
  win_rate: number
  total_trades: number
  status: 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED'
  created_at: string
}

// Full backtest response (for detail view)
export interface IBacktestDetail {
  id: number
  name: string
  symbol: string
  strategy_id: number
  start_time: string
  end_time: string
  capital: number
  summary: IBacktestSummary | null
  equity_curve: IEquityPoint[]
  ohlcv: ICandleData[]
  trades: IBacktestTrade[]
  status: 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED'
  error_message: string | null
  created_at: string
  completed_at: string | null
  strategy: IStrategySnapshot | null
}

// ==================== STORE ====================

export const useBacktestStore = defineStore('backtest', () => {
  // State
  const backtests = ref<IBacktestSummaryItem[]>([])
  const currentBacktest = ref<IBacktestDetail | null>(null)
  const loading = ref(false)
  const formLoading = ref(false)
  const pollingInterval = ref<number | null>(null)

  // Form State
  const createForm = ref<ICreateBacktestRequest>({
    name: '',
    symbol: 'BTCUSDT',
    strategy_id: 1,
    days: 30,
    capital: 1000
  })

  // ==================== CRUD ACTIONS ====================

  // Get all backtests (summary view)
  async function fetchBacktests() {
    loading.value = true
    try {
      const response = await get<IApiResponse<IBacktestSummaryItem[]>>(BASE_URL)
      backtests.value = response.data.data
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to fetch backtests')
    } finally {
      loading.value = false
    }
  }

  // Get backtest by ID (detail view) - silent version (no error alert)
  async function fetchBacktestDetail(id: number, silent: boolean = false): Promise<IBacktestDetail | null> {
    try {
      const response = await get<IApiResponse<IBacktestDetail>>(`${BASE_URL}/${id}`)
      currentBacktest.value = response.data.data
      return currentBacktest.value
    } catch (error: any) {
      if (!silent) {
        showError('Error', error.response?.data?.message || 'Failed to fetch backtest detail')
      }
      return null
    }
  }

  // Create backtest
  async function createBacktest(): Promise<number | null> {
    formLoading.value = true
    try {
      const response = await post<IApiResponse<IBacktestDetail>>(BASE_URL, createForm.value)
      const backtest = response.data.data
      
      showSuccess('Success', 'Backtest started successfully')
      
      // Start polling for completion
      startPolling(backtest.id)
      
      return backtest.id
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to create backtest')
      return null
    } finally {
      formLoading.value = false
    }
  }

  // Delete backtest
  async function deleteBacktest(id: number, name: string): Promise<boolean> {
    const result = await showConfirm(
      'Delete Backtest?',
      `Are you sure you want to delete "${name}"? This action cannot be undone.`
    )

    if (!result.isConfirmed) return false

    loading.value = true
    try {
      await del<IApiResponse<void>>(`${BASE_URL}/${id}`)
      showSuccess('Success', 'Backtest deleted successfully')
      await fetchBacktests()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to delete backtest')
      return false
    } finally {
      loading.value = false
    }
  }

  // ==================== POLLING ====================

  // Track current polling ID to avoid duplicate polling
  let currentPollingId: number | null = null

  // Start polling for backtest completion
  function startPolling(id: number) {
    // Don't start if already polling for the same ID
    if (currentPollingId === id) {
      return
    }

    // Clear existing interval
    stopPolling()

    currentPollingId = id

    // Poll every 2 seconds
    pollingInterval.value = window.setInterval(async () => {
      const backtest = await fetchBacktestDetail(id, true) // silent polling

      if (backtest) {
        if (backtest.status === 'COMPLETED') {
          stopPolling()
          // showSuccess('Success', 'Backtest completed successfully')
          await fetchBacktests() // Refresh list
        } else if (backtest.status === 'FAILED') {
          stopPolling()
          // showError('Error', backtest.error_message || 'Backtest failed')
        }
      }
    }, 2000)
  }

  // Stop polling
  function stopPolling() {
    if (pollingInterval.value) {
      clearInterval(pollingInterval.value)
      pollingInterval.value = null
      currentPollingId = null
    }
  }

  // ==================== UTILS ====================

  // Reset form
  function resetForm() {
    createForm.value = {
      name: '',
      symbol: 'BTCUSDT',
      strategy_id: 1,
      days: 30,
      capital: 1000
    }
  }

  // Load strategies for dropdown (helper)
  async function getStrategies() {
    const response = await get<IApiResponse<any[]>>('/strategies')
    return response.data.data.filter(s => s.is_active)
  }

  return {
    // State
    backtests,
    currentBacktest,
    loading,
    formLoading,
    createForm,

    // Actions
    fetchBacktests,
    fetchBacktestDetail,
    createBacktest,
    deleteBacktest,
    startPolling,
    stopPolling,
    resetForm,
    getStrategies
  }
})
