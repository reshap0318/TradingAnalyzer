import { defineStore } from 'pinia'
import { ref } from 'vue'
import { type IApiResponse, get, post, put, del } from '@/lib/axios'
import { showSuccess, showError, showConfirm } from '@/lib/sweetalert'
import useVuelidate from '@vuelidate/core'
import { required, minLength, minValue, maxValue } from '@vuelidate/validators'
import { useTimeframeStore } from '@/stores/timeframe.store'
import { useIndicatorStore } from '@/stores/indicator.store'
import { useConfigStore } from '@/stores/config.store'

const BASE_URL = '/strategies'

// Interfaces
export interface IStrategy {
  id: number
  strategy_name: string
  primary_tf: string
  is_active: boolean
  created_at: string
  updated_at: string
  timeframes: IStrategyTimeframe[]
  indicator_weights: IStrategyIndicator[]
  money_management: any[] | IMoneyManagement  // Can be array or object
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

export interface IMoneyManagement {
  min_confidence: number
  max_daily_trades: number
  max_daily_loss_percent: number
  max_position_size: number
  risk_reward_ratio: number
  leverage: number
  is_agressive: number  // 0 = Conservative, 1 = Aggressive
  order_expiration_hours: number
}

// Form interfaces
export interface IStrategyForm {
  strategy_name: string
  primary_tf: string
  is_active: boolean
  timeframes: { tf: string; weight: number }[]
  indicator_weights: { indicator_id: number; weight: number }[]
  money_management: IMoneyManagement
}

export const useStrategiesStore = defineStore('strategies', () => {
  // State
  const strategies = ref<IStrategy[]>([])
  const currentStrategy = ref<IStrategy | null>(null)
  const loading = ref(false)  // For page load (fetch strategies)
  const formLoading = ref(false)  // For form submit (create/update)

  // Form State
  const strategyForm = ref<IStrategyForm>({
    strategy_name: '',
    primary_tf: '15m',
    is_active: false,
    timeframes: [{ tf: '15m', weight: 0.6 }, { tf: '1h', weight: 0.4 }],
    indicator_weights: [],
    money_management: {
      min_confidence: 45,
      max_daily_trades: 10,
      max_daily_loss_percent: 5,
      max_position_size: 0.15,
      risk_reward_ratio: 1.5,
      leverage: 5,
      is_agressive: 0,  // 0 = Conservative, 1 = Aggressive
      order_expiration_hours: 4
    }
  })

  // Validation Rules
  const formRules = ref({
    strategy_name: { required, minLength: minLength(3) },
    primary_tf: { required },
    timeframes: { required },
    indicator_weights: { required },
    money_management: {
      min_confidence: { required, minValue: minValue(0), maxValue: maxValue(100) },
      max_daily_trades: { required, minValue: minValue(1) },
      max_daily_loss_percent: { required, minValue: minValue(0), maxValue: maxValue(100) },
      max_position_size: { required, minValue: minValue(0), maxValue: maxValue(1) },
      risk_reward_ratio: { required, minValue: minValue(0) },
      leverage: { required, minValue: minValue(1) },
      order_expiration_hours: { required, minValue: minValue(1) }
    }
  })

  const formValid = useVuelidate(formRules, strategyForm)

  // Helper to get master data from other stores
  const getTimeframes = () => {
    const timeframeStore = useTimeframeStore()
    return timeframeStore.items
  }

  const getIndicators = () => {
    const indicatorStore = useIndicatorStore()
    return indicatorStore.items.filter((ind: any) => ind.is_active)
  }

  const getMMConfigs = () => {
    const configStore = useConfigStore()
    return configStore.items.filter((config: any) => config.category === 'MONEY_MANAGEMENT')
  }

  // CRUD Actions
  async function fetchStrategies() {
    loading.value = true
    try {
      const response = await get<IApiResponse<IStrategy[]>>(BASE_URL)
      strategies.value = response.data.data
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to fetch strategies')
    } finally {
      loading.value = false
    }
  }

  async function fetchStrategy(id: number) {
    try {
      const response = await get<IApiResponse<IStrategy>>(`${BASE_URL}/${id}`)
      currentStrategy.value = response.data.data
      return currentStrategy.value
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to fetch strategy')
      return null
    }
  }

  async function createStrategy(): Promise<boolean> {
    // Validate form
    const valid = await formValid.value.$validate()
    if (!valid) return false

    formLoading.value = true
    try {
      // Convert money_management object to array format for backend
      const mmArray = Object.entries(strategyForm.value.money_management).map(([key, value]) => ({
        parameter: key.toUpperCase(),
        value: String(value)
      }))
      
      const payload = {
        ...strategyForm.value,
        money_management: mmArray
      }
      
      await post<IApiResponse<IStrategy>>(BASE_URL, payload)
      showSuccess('Success', 'Strategy created successfully')
      await fetchStrategies()
      resetForm()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to create strategy')
      return false
    } finally {
      formLoading.value = false
    }
  }

  async function updateStrategy(id: number): Promise<boolean> {
    // Validate form
    const valid = await formValid.value.$validate()
    if (!valid) return false

    formLoading.value = true
    try {
      // Convert money_management object to array format for backend
      const mmArray = Object.entries(strategyForm.value.money_management).map(([key, value]) => ({
        parameter: key.toUpperCase(),
        value: String(value)
      }))
      
      const payload = {
        ...strategyForm.value,
        money_management: mmArray
      }
      
      await put<IApiResponse<IStrategy>>(`${BASE_URL}/${id}`, payload)
      showSuccess('Success', 'Strategy updated successfully')
      await fetchStrategies()
      resetForm()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to update strategy')
      return false
    } finally {
      formLoading.value = false
    }
  }

  async function deleteStrategy(id: number, name: string): Promise<boolean> {
    const result = await showConfirm(
      'Delete Strategy?',
      `Are you sure you want to delete "${name}"? This action cannot be undone.`
    )

    if (!result.isConfirmed) return false

    loading.value = true
    try {
      await del<IApiResponse<void>>(`${BASE_URL}/${id}`)
      showSuccess('Success', 'Strategy deleted successfully')
      await fetchStrategies()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to delete strategy')
      return false
    } finally {
      loading.value = false
    }
  }

  async function activateStrategy(id: number, name: string): Promise<boolean> {
    const result = await showConfirm(
      'Activate Strategy?',
      `Are you sure you want to activate "${name}"? This will deactivate the current active strategy.`
    )

    if (!result.isConfirmed) return false

    loading.value = true
    try {
      await post<IApiResponse<IStrategy>>(`${BASE_URL}/${id}/activate`)
      showSuccess('Success', `"${name}" activated successfully`)
      await fetchStrategies()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to activate strategy')
      return false
    } finally {
      loading.value = false
    }
  }

  async function deactivateStrategy(id: number, name: string): Promise<boolean> {
    const result = await showConfirm(
      'Deactivate Strategy?',
      `Are you sure you want to deactivate "${name}"?`
    )

    if (!result.isConfirmed) return false

    loading.value = true
    try {
      await post<IApiResponse<IStrategy>>(`${BASE_URL}/${id}/deactivate`)
      showSuccess('Success', `"${name}" deactivated successfully`)
      await fetchStrategies()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to deactivate strategy')
      return false
    } finally {
      loading.value = false
    }
  }

  function resetForm() {
    strategyForm.value = {
      strategy_name: '',
      primary_tf: '15m',
      is_active: false,
      timeframes: [{ tf: '15m', weight: 0.6 }, { tf: '1h', weight: 0.4 }],
      indicator_weights: [],
      money_management: {
        min_confidence: 45,
        max_daily_trades: 10,
        max_daily_loss_percent: 5,
        max_position_size: 0.15,
        risk_reward_ratio: 1.5,
        leverage: 5,
        is_agressive: 0,
        order_expiration_hours: 4
      }
    }
    formValid.value.$reset()
  }

  function loadStrategyToForm(strategy: IStrategy) {
    const configStore = useConfigStore()
    const mmConfigs = configStore.items.filter(c => c.category === 'MONEY_MANAGEMENT')
    
    // Build default money management from configs
    const defaultMM: any = {}
    mmConfigs.forEach(config => {
      const fieldName = config.config_key.toLowerCase()
      const value = config.value
      
      // Convert to appropriate type
      if (config.config_key.toLowerCase().includes('is_')) {
        // Boolean field
        defaultMM[fieldName] = value.toLowerCase() === 'true' ? 1 : 0
      } else if (value.includes('.')) {
        // Decimal number
        defaultMM[fieldName] = parseFloat(value) || 0
      } else {
        // Integer
        defaultMM[fieldName] = parseInt(value) || 0
      }
    })
    
    // Convert money_management from array to object if needed
    let mmObject: any = {}
    if (Array.isArray(strategy.money_management)) {
      // Backend returns array format: [{parameter, value}, ...]
      strategy.money_management.forEach((item: any) => {
        const key = item.parameter.toLowerCase()
        let value: any = item.value
        
        // Convert to appropriate type
        if (key.includes('is_')) {
          value = value.toLowerCase() === 'true' ? 1 : 0
        } else if (value.includes('.')) {
          value = parseFloat(value) || 0
        } else {
          value = parseInt(value) || 0
        }
        
        mmObject[key] = value
      })
    } else if (strategy.money_management && typeof strategy.money_management === 'object') {
      // Already object format
      mmObject = strategy.money_management
    }
    
    strategyForm.value = {
      strategy_name: strategy.strategy_name,
      primary_tf: strategy.primary_tf,
      is_active: strategy.is_active,
      timeframes: strategy.timeframes.map(tf => ({
        tf: tf.tf,
        weight: tf.weight
      })),
      indicator_weights: strategy.indicator_weights.map(ind => ({
        indicator_id: ind.indicator_id,
        weight: ind.weight
      })),
      money_management: { 
        // Start with defaults from config
        ...defaultMM,
        // Override with strategy values if they exist
        ...mmObject,
        // Ensure is_agressive is number (0/1)
        is_agressive: mmObject.is_agressive === undefined || mmObject.is_agressive === null
          ? defaultMM.is_agressive ?? 0
          : typeof mmObject.is_agressive === 'boolean' 
            ? (mmObject.is_agressive ? 1 : 0)
            : mmObject.is_agressive
      }
    }
  }

  return {
    // State
    strategies,
    currentStrategy,
    strategyForm,
    formValid,
    loading,
    formLoading,

    // Actions
    fetchStrategies,
    fetchStrategy,
    createStrategy,
    updateStrategy,
    deleteStrategy,
    activateStrategy,
    deactivateStrategy,
    resetForm,
    loadStrategyToForm,
    getTimeframes,
    getIndicators,
    getMMConfigs
  }
})
