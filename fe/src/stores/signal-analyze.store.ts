import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { type IApiResponse, post } from '@/lib/axios'
import { showSuccess, showError, showWarning } from '@/lib/sweetalert'
import useVuelidate from '@vuelidate/core'
import { required, minValue } from '@vuelidate/validators'

const BASE_URL = '/signal'

// ============== REQUEST DTOs ==============
export interface ISignalAnalyzeRequest {
  symbol: string
  strategy_id?: number
  capital?: number
}

export interface ISignalRawRequest {
  symbol: string
  timeframes: {
    name: string
    limit: number
  }[]
}

// ============== RESPONSE DTOs ==============
export interface ISignalAnalyzeResponse {
  symbol: string
  primary_timeframe: string
  timestamp: string
  signal: ISignalInfo
  scoring: IScoringBreakdown
}

export interface ISignalInfo {
  valid: boolean
  signal: string
  current_price?: number
  trading_plan?: ITradingPlan
}

export interface ITradingPlan {
  mode: 'CONSERVATIVE' | 'AGGRESSIVE'
  entries: ITradingPlanEntry[]
  take_profit: number
  stop_loss: number
  resistance: number
  support: number
  risk_reward_ratio: number
  buffer_percent: number
  summary?: ITradingPlanSummary
}

export interface ITradingPlanEntry {
  entry_number: number
  entry_price: number
  position_size: string
  position_value: number
  position_qty: number
}

export interface ITradingPlanSummary {
  capital_used: number
  total_entries: number
  total_position_value: number
  total_position_qty: number
  avg_entry_price: number
  max_risk_usdt: number
  max_risk_percent: number
  risk_from_capital: number
  target_profit_usdt: number
  target_profit_percent: number
  profit_from_capital: number
  effective_leverage: number
}

export interface IScoringBreakdown {
  totalScore: number
  confidence: number
  breakdown: ITimeframeSignalData[]
}

export interface ITimeframeSignalData {
  timeframe: string
  trend: string
  rawSignal: number
  weight: number
  contribution: number
  indicator: IIndicatorBreakdown[]
}

export interface IIndicatorBreakdown {
  name: string
  role: string
  rawSignal: number
  weight: number
  contribution: number
  details: string[]
  value?: any
  zone?: string
}

// Raw Data Response (untuk chart)
export interface ISignalRawResponse {
  symbol: string
  timeframes: ITimeframeRawData[]
}

export interface ITimeframeRawData {
  timeframe: string
  raws: IRawData[]
}

export interface IRawData {
  timestamp: number
  open: number
  high: number
  low: number
  close: number
  volume: number
}

// ============== STORE ==============
export const useSignalAnalyzeStore = defineStore('signalAnalyze', () => {
  // State
  const loading = ref(false)
  const chartLoading = ref(false)
  const result = ref<ISignalAnalyzeResponse | null>(null)
  const chartData = ref<ISignalRawResponse | null>(null)

  // Form State
  const analyzeReq = ref<ISignalAnalyzeRequest>({
    symbol: '',
    strategy_id: undefined,
    capital: 50
  })

  // Validation Rules
  const analyzeRules = ref({
    symbol: { required },
    strategy_id: { required },
    capital: { required, minValue: minValue(1) }
  })

  // Vuelidate instance
  const analyzeReqValid = useVuelidate(analyzeRules, analyzeReq as any)

  // Getters
  const hasResult = computed(() => result.value !== null)
  const primaryTimeframe = computed(() => result.value?.primary_timeframe || '')
  const currentPrice = computed(() => result.value?.signal.current_price || 0)
  const signalAction = computed(() => result.value?.signal.signal || '')

  // Reset State
  function resetState() {
    loading.value = false
    chartLoading.value = false
    result.value = null
    chartData.value = null
    analyzeReq.value = {
      symbol: '',
      strategy_id: undefined,
      capital: 50
    }
    // Reset validation state
    analyzeReqValid.value.$reset()
  }

  // Execute Trade
  async function executeTrade(signalResult: ISignalAnalyzeResponse): Promise<boolean> {
    try {
      const payload = {
        symbol: analyzeReq.value.symbol.toUpperCase(),
        strategy_id: analyzeReq.value.strategy_id,
        capital: analyzeReq.value.capital,
        signal_data: signalResult
      }

      const { data } = await post<IApiResponse<any>>(
        `/trade/execute`,
        payload
      )
      
      if (data.data.execution_info?.executed) {
        showSuccess('Trade Executed', `Trade for ${payload.symbol} has been executed successfully`)
      }
      else {
        showWarning('Trade Executed', `Trade for ${payload.symbol} has can't exucuted because signal not valid`)
      }
      return true
    } catch (error: any) {
      const errorMsg = error.response?.data?.message || 'Failed to execute trade'
      showError('Execution Error', errorMsg)
      return false
    }
  }
  const isSignalValid = computed(() => result.value?.signal.valid || false)
  const confidence = computed(() => result.value?.scoring.confidence || 0)

  // Actions
  async function analyzeSignal(): Promise<boolean> {
    // Validate form
    const valid = await analyzeReqValid.value.$validate()
    if (!valid) return false

    loading.value = true

    try {
      const payload: ISignalAnalyzeRequest = {
        symbol: analyzeReq.value.symbol.toUpperCase(),
        strategy_id: analyzeReq.value.strategy_id,
        capital: analyzeReq.value.capital
      }

      const response = await post<IApiResponse<ISignalAnalyzeResponse>>(
        `${BASE_URL}/analyze`,
        payload
      )

      result.value = response.data.data

      // Auto fetch raw data untuk chart setelah analyze berhasil
      if (result.value?.primary_timeframe) {
        await fetchRawData(result.value.primary_timeframe, 300)
      }

      showSuccess('Analysis Complete', `Signal for ${payload.symbol} has been analyzed`)
      return true
    } catch (error: any) {
      const errorMsg = error.response?.data?.message || 'Failed to analyze signal'
      showError('Analysis Failed', errorMsg)
      return false
    } finally {
      loading.value = false
    }
  }

  async function fetchRawData(timeframe: string, limit: number = 300): Promise<boolean> {
    chartLoading.value = true

    try {
      const payload: ISignalRawRequest = {
        symbol: analyzeReq.value.symbol.toUpperCase(),
        timeframes: [
          {
            name: timeframe,
            limit: limit
          }
        ]
      }

      const response = await post<IApiResponse<ISignalRawResponse>>(
        `${BASE_URL}/raws`,
        payload
      )

      chartData.value = response.data.data
      return true
    } catch (error: any) {
      const errorMsg = error.response?.data?.message || 'Failed to fetch chart data'
      showError('Chart Error', errorMsg)
      return false
    } finally {
      chartLoading.value = false
    }
  }

  function resetResult() {
    result.value = null
    chartData.value = null
    analyzeReqValid.value.$reset()
  }

  function resetForm() {
    analyzeReq.value = {
      symbol: '',
      strategy_id: undefined,
      capital: 50
    }
    analyzeReqValid.value.$reset()
    resetResult()
  }

  return {
    // State
    loading,
    chartLoading,
    result,
    chartData,
    analyzeReq,
    analyzeReqValid,

    // Getters
    hasResult,
    primaryTimeframe,
    currentPrice,
    signalAction,
    isSignalValid,
    confidence,

    // Actions
    analyzeSignal,
    fetchRawData,
    resetResult,
    resetForm,
    resetState,
    executeTrade
  }
})
