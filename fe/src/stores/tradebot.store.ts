import { defineStore } from 'pinia'
import { ref } from 'vue'
import { type IApiResponse, get, post } from '@/lib/axios'
import { showSuccess, showError, showConfirm } from '@/lib/sweetalert'

const BASE_URL = '/trade/bot'

// Interfaces
export interface IBotStatus {
  id: number
  is_active: boolean
  active_since?: string
  strategy_id?: number
  last_scan?: string
  trades_executed?: number
  scan_interval?: number
}

export interface IStrategyData {
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
  is_agressive?: boolean
  order_expiration_hours?: number
}

export const useTradeBotStore = defineStore('tradebot', () => {
  // State
  const botStatus = ref<IBotStatus | null>(null)
  const strategy = ref<IStrategyData | null>(null)
  const strategies = ref<IStrategyData[]>([])
  const loading = ref(false)
  const toggling = ref(false)
  const strategiesLoaded = ref(false)

  // Actions
  async function fetchBotStatus() {
    loading.value = true
    try {
      const response = await get<IApiResponse<IBotStatus & { strategy?: IStrategyData | null }>>(`${BASE_URL}/status`)
      botStatus.value = response.data.data
      
      // Clear strategy if bot is not active
      if (!botStatus.value?.is_active) {
        strategy.value = null
        return
      }
      
      // Assign strategy from response if exists
      if (response.data.data.strategy) {
        strategy.value = response.data.data.strategy
      } else if (botStatus.value.strategy_id) {
        // Fallback: fetch strategy if bot active but strategy not in response
        await fetchStrategy(botStatus.value.strategy_id)
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

  async function fetchStrategy(strategyId: number) {
    try {
      const response = await get<IApiResponse<IStrategyData>>(`/strategies/${strategyId}`)
      strategy.value = response.data.data
    } catch (error: any) {
      console.warn('Failed to fetch strategy:', error)
    }
  }

  async function fetchStrategies() {
    if (strategiesLoaded.value) return
    
    loading.value = true
    try {
      const response = await get<IApiResponse<IStrategyData[]>>('/strategies')
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

  return {
    // State
    botStatus,
    strategy,
    strategies,
    loading,
    toggling,
    strategiesLoaded,

    // Actions
    fetchBotStatus,
    activateBot,
    deactivateBot,
    toggleBot,
    fetchStrategies,
    selectStrategy
  }
})
