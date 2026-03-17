<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useSignalAnalyzeStore, type ITimeframeRawData } from '@/stores/signal-analyze.store'
import { useStrategiesStore } from '@/stores/strategies.store'
import { getValidationErrors } from '@/helpers/validation'
import { showSuccess, showError } from '@/lib/sweetalert'
import { DefaultLayout } from '@/layouts'
import { PhFlask, PhCurrencyBtc, PhPlay } from '@phosphor-icons/vue'
import { UiButton } from '@/components/common'
import Swal from 'sweetalert2'

// Components
import SignalInfoCard from '@/components/features/signal-analyze/SignalInfoCard.vue'
import TradingPlanCard from '@/components/features/signal-analyze/TradingPlanCard.vue'
import ScoringBreakdownTable from '@/components/features/signal-analyze/ScoringBreakdownTable.vue'
import SignalChart from '@/components/features/signal-analyze/SignalChart.vue'

const signalStore = useSignalAnalyzeStore()
const strategiesStore = useStrategiesStore()

// Execute loading state
const isExecuting = ref(false)

// Access form state dan validation dari store
const v$ = signalStore.analyzeReqValid
const isLoading = computed(() => signalStore.loading)

// Check if can execute (must have analyzed data first)
const canExecute = computed(() => {
  return !!result.value && !!tradingPlan.value
})

// Get strategies untuk dropdown (tampilkan semua)
const strategies = computed(() => strategiesStore.strategies)

// Get data dari store
const result = computed(() => signalStore.result)
const chartData = computed((): ITimeframeRawData | null => {
  if (!signalStore.chartData || !signalStore.chartData.timeframes.length) return null
  return signalStore.chartData.timeframes[0] ?? null
})
const tradingPlan = computed(() => result.value?.signal.trading_plan || null)
const primaryTimeframe = computed(() => signalStore.primaryTimeframe)
const isChartLoading = computed(() => signalStore.chartLoading)

// Get symbol dan strategy info untuk ditampilkan di card
const currentSymbol = computed(() => signalStore.analyzeReq.symbol)
const currentStrategyName = computed(() => {
  const strategyId = signalStore.analyzeReq.strategy_id
  if (!strategyId) return ''
  const strategy = strategiesStore.strategies.find(s => s.id === strategyId)
  return strategy?.strategy_name || ''
})
const currentStrategyTF = computed(() => {
  const strategyId = signalStore.analyzeReq.strategy_id
  if (!strategyId) return ''
  const strategy = strategiesStore.strategies.find(s => s.id === strategyId)
  return strategy?.primary_tf || ''
})

// Submit handler
const handleSubmit = async () => {
  const success = await signalStore.analyzeSignal()
  if (success) {
    // Form sudah di-reset otomatis oleh store
  }
}

// Execute handler
const handleExecute = async () => {
  if (!canExecute.value || !result.value) return

  // Get strategy info for confirmation
  const strategyId = signalStore.analyzeReq.strategy_id
  const strategy = strategiesStore.strategies.find(s => s.id === strategyId)
  const strategyName = strategy?.strategy_name || 'Unknown'
  const symbol = signalStore.analyzeReq.symbol.toUpperCase()

  // Show confirmation dialog
  const confirmResult = await Swal.fire({
    title: 'Execute Trade?',
    html: `
      <div style="text-align: left; padding: 20px 0;">
        <p style="margin-bottom: 10px;">You are about to execute a trade with:</p>
        <div style="background: #f3f4f6; padding: 15px; border-radius: 8px;">
          <p style="margin: 5px 0;"><strong>📊 Symbol:</strong> ${symbol}</p>
          <p style="margin: 5px 0;"><strong>🎯 Strategy:</strong> ${strategyName}</p>
          <p style="margin: 5px 0;"><strong>💰 Capital:</strong> $${signalStore.analyzeReq.capital}</p>
        </div>
        <p style="margin-top: 15px; color: #f59e0b;">⚠️ This action will create a real trade in the database.</p>
      </div>
    `,
    icon: 'question',
    showCancelButton: true,
    confirmButtonText: 'Yes, Execute!',
    cancelButtonText: 'Cancel',
    confirmButtonColor: '#10b981',
    cancelButtonColor: '#6b7280',
    reverseButtons: true
  })

  if (!confirmResult.isConfirmed) return

  // Execute trade
  isExecuting.value = true
  try {
    const success = await signalStore.executeTrade(result.value)
    if (success) {
      showSuccess('Trade Executed', `Trade for ${symbol} has been executed successfully`)
    }
  } catch (error: any) {
    showError('Execution Failed', error.message || 'Failed to execute trade')
  } finally {
    isExecuting.value = false
  }
}

// Handle Enter key
const handleKeyPress = (e: KeyboardEvent) => {
  if (e.key === 'Enter') {
    handleSubmit()
  }
}

// Fetch strategies on mount
onMounted(() => {
  // Reset semua data saat pertama kali masuk halaman
  signalStore.resetState()
  
  strategiesStore.fetchStrategies().then(() => {
    // Set default strategy ke yang pertama yang active
    if (strategies.value.length > 0 && !signalStore.analyzeReq.strategy_id) {
      const activeStrategy = strategies.value.find(s => s.is_active)
      if (activeStrategy) {
        signalStore.analyzeReq.strategy_id = activeStrategy.id
      }
    }
  })
})

// Reset data saat leave halaman
onUnmounted(() => {
  signalStore.resetState()
})
</script>

<template>
  <DefaultLayout>
    <template #header-title>Signal Analyze</template>

    <div class="space-y-6">
      <!-- Loading State -->
      <div v-if="isLoading" class="flex items-center justify-center py-20">
        <div class="relative">
          <div class="animate-spin rounded-full h-16 w-16 border-b-2 border-primary"></div>
          <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2">
            <PhFlask :size="24" class="text-primary" />
          </div>
        </div>
      </div>

      <template v-else>
        <!-- Top Section: 2 Columns -->
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
          <!-- Left: Form + Signal Info (4 columns = 1/3) -->
          <div class="lg:col-span-4 space-y-6">
            <!-- Form Card -->
            <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="p-2.5 bg-blue-50 rounded-xl">
                  <PhCurrencyBtc :size="24" class="text-blue-600" weight="fill" />
                </div>
                <div>
                  <h2 class="text-lg font-bold text-gray-900">Analysis Setup</h2>
                  <p class="text-xs text-gray-500">Configure your analysis</p>
                </div>
              </div>

              <form @submit.prevent="handleSubmit" @keypress="handleKeyPress">
                <!-- Symbol & Capital Row (7:3 ratio) -->
                <div class="grid grid-cols-10 gap-3 mb-4">
                  <!-- Symbol Input (7/10 = 70%) -->
                  <div class="col-span-7">
                    <label class="block text-sm font-medium text-gray-700 mb-2">
                      Symbol <span class="text-red-500">*</span>
                    </label>
                    <input
                      v-model="signalStore.analyzeReq.symbol"
                      type="text"
                      placeholder="e.g., BTCUSDT"
                      class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent transition-all uppercase"
                      :class="{ 'border-red-500': v$?.symbol?.$error }"
                      :disabled="isLoading"
                    />
                    <p v-if="v$?.symbol?.$error" class="mt-1 text-sm text-red-600">
                      {{ getValidationErrors(v$.symbol).join(', ') }}
                    </p>
                  </div>

                  <!-- Capital Input (3/10 = 30%) -->
                  <div class="col-span-3">
                    <label class="block text-sm font-medium text-gray-700 mb-2">
                      Capital <span class="text-red-500">*</span>
                    </label>
                    <input
                      v-model.number="signalStore.analyzeReq.capital"
                      type="number"
                      placeholder="50"
                      class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
                      :class="{ 'border-red-500': v$?.capital?.$error }"
                      :disabled="isLoading"
                    />
                    <p v-if="v$?.capital?.$error" class="mt-1 text-sm text-red-600">
                      {{ getValidationErrors(v$.capital).join(', ') }}
                    </p>
                  </div>
                </div>

                <!-- Strategy Select -->
                <div class="mb-4">
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    Strategy <span class="text-red-500">*</span>
                  </label>
                  <select
                    v-model="signalStore.analyzeReq.strategy_id"
                    class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
                    :class="{ 'border-red-500': v$?.strategy_id?.$error }"
                    :disabled="isLoading"
                  >
                    <option value="">Select a strategy...</option>
                    <option
                      v-for="strategy in strategies"
                      :key="strategy.id"
                      :value="strategy.id"
                    >
                      {{ strategy.strategy_name }} ({{ strategy.primary_tf }})
                    </option>
                  </select>
                  <p v-if="v$?.strategy_id?.$error" class="mt-1 text-sm text-red-600">
                    {{ getValidationErrors(v$.strategy_id).join(', ') }}
                  </p>
                </div>

                <!-- Submit Button + Execute Button (8:2 ratio) -->
                <div class="grid grid-cols-10 gap-3">
                  <!-- Analyze Button (8/10 = 80%) -->
                  <div class="col-span-8">
                    <UiButton
                      type="submit"
                      variant="primary"
                      :loading="isLoading"
                      full-width
                    >
                      {{ isLoading ? 'Analyzing...' : 'Analyze Signal' }}
                    </UiButton>
                  </div>

                  <!-- Execute Button (2/10 = 20%) -->
                  <div class="col-span-2">
                    <UiButton
                      type="button"
                      variant="success"
                      :loading="isExecuting"
                      :disabled="!canExecute || isLoading"
                      full-width
                      @click="handleExecute"
                    >
                      <PhPlay :size="20" weight="fill" />
                    </UiButton>
                  </div>
                </div>
              </form>
            </div>

            <!-- Signal Info Card -->
            <SignalInfoCard
              :result="result"
              :symbol="currentSymbol"
              :strategy-name="currentStrategyName"
              :strategy-timeframe="currentStrategyTF"
            />
          </div>

          <!-- Right: Trading Plan (8 columns = 2/3) -->
          <div class="lg:col-span-8">
            <TradingPlanCard :trading-plan="tradingPlan" />
          </div>
        </div>

        <!-- Chart Section -->
        <SignalChart
          :chart-data="chartData"
          :trading-plan="tradingPlan"
          :timeframe="primaryTimeframe"
          :is-loading="isChartLoading"
        />

        <!-- Scoring Breakdown -->
        <ScoringBreakdownTable :scoring="result?.scoring || null" />
      </template>
    </div>
  </DefaultLayout>
</template>
