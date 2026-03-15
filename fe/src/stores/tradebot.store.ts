import { defineStore } from 'pinia'
import { ref } from 'vue'
import { type IApiResponse, get, post } from '@/lib/axios'
import { showSuccess, showError, showConfirm } from '@/lib/sweetalert'

const BASE_URL = '/trade/bot'

// Interfaces
export interface IBotStatus {
  is_active: boolean
  strategy?: IStrategy | null
  bot_started_at?: string
  bot_running_duration?: string
  bot_running_seconds?: number
}

export interface IStrategy {
  id: number
  strategy_name: string
  primary_tf: string
  is_active: boolean
  timeframes?: IStrategyTimeframe[]
  indicator_weights?: IStrategyIndicator[]
  money_management?: IMoneyManagement
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
    name: string
    indicator: string
  }
}

export interface IMoneyManagement {
  min_confidence?: number
  max_daily_trades?: number
  max_daily_loss_percent?: number
  max_position_size?: number
  risk_reward_ratio?: number
  leverage?: number
  is_agressive?: number
  order_expiration_hours?: number
}

// Session data interfaces
export interface ISessionSummary {
  total_trades: number
  executed: number
  skipped: number
  success_rate: number
  total_pnl: number
  symbols_traded: string[]
  session_started: string
}

export interface ITradeOrder {
  entry_number: number
  binance_order_id: number
  price: number
  quantity: number
  type: string
  status: 'PENDING' | 'FILLED' | 'CANCELLED' | 'REJECTED'
}

export interface ITrade {
  id: number
  symbol: string
  interval: string
  side: string
  confidence: number
  total_score: number
  is_aggressive: boolean
  tp_price: number
  sl_price: number
  risk_reward_ratio: number
  avg_entry_price: number
  leverage: number
  capital_used: number
  total_qty: number
  status: string
  description: string
  tp_order_id: number
  sl_order_id: number
  tp_sl_status: string
  exit_price: number
  pnl: number
  pnl_pct: number
  created_at: string
  updated_at: string
  closed_at: string | null
  orders: ITradeOrder[]
}

export const useTradeBotStore = defineStore('tradebot', () => {
  // State
  const botStatus = ref<IBotStatus | null>(null)
  const strategy = ref<IStrategy | null>(null)
  const strategies = ref<IStrategy[]>([])
  const loading = ref(false)
  const toggling = ref(false)
  const strategiesLoaded = ref(false)
  
  // Session data state
  const sessionSummary = ref<ISessionSummary | null>(null)
  const activeTrades = ref<ITrade[]>([])
  const executedTrades = ref<ITrade[]>([])
  const summaryLoading = ref(false)

  // Actions
  async function fetchBotStatus() {
    loading.value = true
    try {
      const response = await get<IApiResponse<IBotStatus & { strategy?: IStrategy | null }>>(`${BASE_URL}/status`)
      botStatus.value = response.data.data
      
      // Clear strategy if bot is not active
      if (!botStatus.value?.is_active) {
        strategy.value = null
        return
      }
      
      // Assign strategy from response if exists
      if (response.data.data.strategy) {
        strategy.value = response.data.data.strategy
      }
    } catch (error: any) {
      // If 404 or bot not found, set to null (not active)
      if (error.response?.status === 404 || error.response?.data?.message?.includes('not found')) {
        botStatus.value = null
        strategy.value = null
      } else {
        showError('Error', error.response?.data?.message || 'Failed to fetch bot status')
      }
    } finally {
      loading.value = false
    }
  }

  async function fetchStrategies() {
    if (strategiesLoaded.value) return
    
    loading.value = true
    try {
      const response = await get<IApiResponse<IStrategy[]>>('/strategies')
      strategies.value = response.data.data
      strategiesLoaded.value = true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to fetch strategies')
    } finally {
      loading.value = false
    }
  }

  async function selectStrategy(strategyId: number | null): Promise<boolean> {
    const result = await showConfirm(
      strategyId ? 'Change Strategy?' : 'Remove Strategy?',
      strategyId
        ? 'Are you sure you want to use this strategy for the trading bot?'
        : 'Are you sure you want to run the bot without a specific strategy?'
    )

    if (!result.isConfirmed) return false

    loading.value = true
    try {
      const payload: any = {}
      if (strategyId !== null) {
        payload.strategy_id = strategyId
      }

      await post(`${BASE_URL}/activate`, payload)
      showSuccess('Success', strategyId ? 'Strategy updated successfully' : 'Bot will use default strategy')
      await fetchBotStatus()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to update strategy')
      return false
    } finally {
      loading.value = false
    }
  }

  async function activateBot(strategyId?: number): Promise<boolean> {
    toggling.value = true
    try {
      const payload: any = {}
      if (strategyId) {
        payload.strategy_id = strategyId
      }

      const response = await post<IApiResponse<IBotStatus>>(`${BASE_URL}/activate`, payload)
      botStatus.value = response.data.data
      
      showSuccess('Success', 'Trading bot activated successfully')
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to activate bot')
      return false
    } finally {
      toggling.value = false
    }
  }

  async function deactivateBot(): Promise<boolean> {
    toggling.value = true
    try {
      const response = await post<IApiResponse<IBotStatus>>(`${BASE_URL}/deactivate`)
      botStatus.value = response.data.data
      
      showSuccess('Success', 'Trading bot deactivated successfully')
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to deactivate bot')
      return false
    } finally {
      toggling.value = false
    }
  }

  async function toggleBot(): Promise<boolean> {
    let success = false
    if (botStatus.value?.is_active) {
      success = await deactivateBot()
    } else {
      success = await activateBot()
    }

    // Refresh status after toggle
    if (success) {
      await fetchBotStatus()
    }

    return success
  }

  // Session data actions
  async function fetchSessionData() {
    if (!botStatus.value?.is_active) return

    summaryLoading.value = true
    try {
      // Fetch all three endpoints in parallel
      const [summaryRes, activeRes, executedRes] = await Promise.all([
        get<IApiResponse<ISessionSummary>>(`${BASE_URL}/summary`),
        get<IApiResponse<ITrade[]>>(`${BASE_URL}/active`),
        get<IApiResponse<ITrade[]>>(`${BASE_URL}/`)
      ])

      // Set data from responses
      sessionSummary.value = summaryRes.data.data
      activeTrades.value = activeRes.data.data || []
      executedTrades.value = executedRes.data.data || []
    } catch (error) {
      // Ignore errors - if bot not running, endpoints will fail and that's OK
      console.warn('Failed to fetch session data (bot may not be running):', error)
    } finally {
      summaryLoading.value = false
    }
  }

  async function manualCloseTrade(tradeId: number): Promise<boolean> {
    loading.value = true
    try {
      await post<IApiResponse<any>>(`/trade/monitor/${tradeId}/close`)
      showSuccess('Success', 'Trade closed manually')
      await fetchSessionData() // Refresh data after close
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to close trade')
      return false
    } finally {
      loading.value = false
    }
  }

  function clearSessionData() {
    sessionSummary.value = null
    activeTrades.value = []
    executedTrades.value = []
  }

  return {
    // State
    botStatus,
    strategy,
    strategies,
    loading,
    toggling,
    strategiesLoaded,
    sessionSummary,
    activeTrades,
    executedTrades,
    summaryLoading,

    // Actions
    fetchBotStatus,
    activateBot,
    deactivateBot,
    toggleBot,
    fetchStrategies,
    selectStrategy,
    fetchSessionData,
    clearSessionData,
    manualCloseTrade
  }
})
