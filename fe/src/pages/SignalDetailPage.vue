<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useSignalStore } from '@/stores/signal.store'
import {
  PhArrowLeft,
  PhTrash,
  PhChartLine,
  PhTrendUp,
  PhClock,
  PhCurrencyBtc
} from '@phosphor-icons/vue'
import {
  createChart,
  CandlestickSeries,
  type IChartApi,
  type ISeriesApi
} from 'lightweight-charts'
import DefaultLayout from '@/layouts/DefaultLayout.vue'

const router = useRouter()
const route = useRoute()
const signalStore = useSignalStore()

const loading = ref(true)
const chartContainerRef = ref<HTMLElement | null>(null)
let chart: IChartApi | null = null
let candlestickSeries: ISeriesApi<'Candlestick'> | null = null

// Computed
const signal = computed(() => signalStore.currentSignal)
const isLoading = computed(() => signalStore.loading)

// Group indicators by timeframe
const groupedIndicators = computed(() => {
  if (!signal.value?.strategy_snapshot?.indicator_weights) return {}

  const groups: Record<string, any[]> = {}
  const globalIndicators: any[] = []

  // Separate global indicators
  signal.value.strategy_snapshot.indicator_weights.forEach((indicator: any) => {
    if (!indicator.timeframe || indicator.timeframe === 'All') {
      globalIndicators.push(indicator)
    } else {
      const tf = indicator.timeframe
      if (!groups[tf]) {
        groups[tf] = []
      }
      groups[tf].push(indicator)
    }
  })

  // Merge global indicators into each timeframe
  const result: Record<string, any[]> = {}
  Object.keys(groups).forEach((tf) => {
    const tfIndicators = groups[tf] || []
    result[tf] = [...globalIndicators, ...tfIndicators]
  })

  // If no timeframe-specific indicators, show global only
  if (Object.keys(groups).length === 0 && globalIndicators.length > 0) {
    result['All Timeframes'] = globalIndicators
  }

  return result
})

// Format functions
function formatNumber(num: number, decimals: number = 2): string {
  return num.toLocaleString('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  })
}

function formatKey(key: string | number): string {
  const strKey = typeof key === 'number' ? String(key) : key
  return strKey
    .toUpperCase()
    .replace(/_/g, ' ')
    .replace(/([A-Z])/g, ' $1')
    .trim()
}

function formatValue(value: any): string {
  if (typeof value === 'boolean') {
    return value ? 'Yes' : 'No'
  }
  if (typeof value === 'number') {
    if (value >= 1000) {
      return value.toLocaleString('en-US')
    }
    return value.toFixed(2)
  }
  return String(value)
}

function getRoleBadge(role?: string) {
  if (!role) role = 'DRIVER'
  const upper = role.toUpperCase()
  switch (upper) {
    case 'DRIVER':
      return { label: 'D', bg: 'bg-blue-100', text: 'text-blue-700', title: 'Driver' }
    case 'FILTER':
      return { label: 'F', bg: 'bg-amber-100', text: 'text-amber-700', title: 'Filter' }
    case 'BOOSTER':
      return { label: 'B', bg: 'bg-purple-100', text: 'text-purple-700', title: 'Booster' }
    default:
      return { label: '?', bg: 'bg-gray-100', text: 'text-gray-700', title: 'Unknown' }
  }
}

// Get badge color for signal category
function getCategoryBadgeClass(category: string): string {
  const colors: Record<string, string> = {
    STRONG_BUY: 'bg-green-100 text-green-800 border-green-300',
    BUY: 'bg-blue-100 text-blue-800 border-blue-300',
    WAIT: 'bg-gray-100 text-gray-800 border-gray-300',
    SELL: 'bg-orange-100 text-orange-800 border-orange-300',
    STRONG_SELL: 'bg-red-100 text-red-800 border-red-300'
  }
  return colors[category] || 'bg-gray-100 text-gray-800 border-gray-300'
}

// Initialize chart
function initChart() {
  if (!chartContainerRef.value || !signal.value?.ohlc_snapshot) {
    console.warn('[Chart] Cannot initialize: missing container or snapshot')
    return
  }

  // Clear previous chart
  if (chartContainerRef.value) {
    chartContainerRef.value.innerHTML = ''
  }

  // Create chart
  chart = createChart(chartContainerRef.value, {
    width: chartContainerRef.value.clientWidth,
    height: 400,
    layout: {
      background: { color: '#ffffff' },
      textColor: '#374151'
    },
    grid: {
      vertLines: { color: '#f0f0f0' },
      horzLines: { color: '#f0f0f0' }
    },
    crosshair: {
      mode: 1,
      vertLine: {
        width: 1,
        color: '#94a3b8',
        style: 3,
        labelBackgroundColor: '#94a3b8'
      },
      horzLine: {
        width: 1,
        color: '#94a3b8',
        style: 3,
        labelBackgroundColor: '#94a3b8'
      }
    },
    rightPriceScale: {
      borderColor: '#e5e7eb',
      scaleMargins: {
        top: 0.1,
        bottom: 0.1
      }
    },
    timeScale: {
      borderColor: '#e5e7eb',
      timeVisible: true,
      secondsVisible: false
    }
  })

  // Create candlestick series - lightweight-charts v5 API
  candlestickSeries = chart.addSeries(CandlestickSeries, {
    upColor: '#22c55e',
    downColor: '#ef4444',
    borderUpColor: '#22c55e',
    borderDownColor: '#ef4444',
    wickUpColor: '#22c55e',
    wickDownColor: '#ef4444'
  })

  // Set candle data
  const candles = signal.value.ohlc_snapshot.candles || []
  if (candles.length === 0) {
    console.warn('[Chart] No candle data available')
    return
  }

  const TIMEZONE_OFFSET = 7 * 60 * 60 * 1000 // 7 jam dalam milliseconds
  const candleData = candles.map((c: any) => ({
    time: Math.floor((c.time * 1000 + TIMEZONE_OFFSET) / 1000) as any,
    open: c.open,
    high: c.high,
    low: c.low,
    close: c.close
  }))

  candlestickSeries.setData(candleData as any)

  // Add TP/SL lines using createPriceLine
  updatePriceLines()

  // Fit content
  chart.timeScale().fitContent()

  // Handle resize
  const handleResize = () => {
    if (chart && chartContainerRef.value) {
      chart.applyOptions({ width: chartContainerRef.value.clientWidth })
    }
  }

  window.addEventListener('resize', handleResize)
}

// Update TP/SL lines
function updatePriceLines() {
  if (!candlestickSeries || !signal.value) return

  const formatPrice = (price: number): string => {
    if (price >= 1000) {
      return price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
    }
    return price.toFixed(4)
  }

  // TP Line - Green Dashed
  if (signal.value.tp_price) {
    candlestickSeries.createPriceLine({
      price: signal.value.tp_price,
      color: '#22c55e',
      lineWidth: 2,
      lineStyle: 2,
      axisLabelVisible: true,
      title: `TP: ${formatPrice(signal.value.tp_price)}`,
      axisLabelColor: '#22c55e',
      axisLabelTextColor: '#ffffff'
    })
  }

  // SL Line - Red Dashed
  if (signal.value.sl_price) {
    candlestickSeries.createPriceLine({
      price: signal.value.sl_price,
      color: '#ef4444',
      lineWidth: 2,
      lineStyle: 2,
      axisLabelVisible: true,
      title: `SL: ${formatPrice(signal.value.sl_price)}`,
      axisLabelColor: '#ef4444',
      axisLabelTextColor: '#ffffff'
    })
  }

  // Entry Lines - Blue Solid
  if (signal.value.entry_levels) {
    signal.value.entry_levels.forEach((entry: any) => {
      candlestickSeries?.createPriceLine({
        price: entry.entry_price,
        color: '#6366f1',
        lineWidth: 2,
        lineStyle: 0,
        axisLabelVisible: true,
        title: `Entry ${entry.entry_number}: ${formatPrice(entry.entry_price)}`,
        axisLabelColor: '#6366f1',
        axisLabelTextColor: '#ffffff'
      })
    })
  }
}

// Delete signal
async function handleDelete() {
  if (!signal.value) return

  const success = await signalStore.deleteSignalById(signal.value.id)
  if (success) {
    router.push('/signals')
  }
}

// Go back
function goBack() {
  router.back()
}

// Watch for signal changes and update chart
watch(
  () => signal.value?.ohlc_snapshot,
  (newSnapshot) => {
    if (newSnapshot && chartContainerRef.value && !chart) {
      nextTick(() => {
        initChart()
      })
    }
  },
  { immediate: false }
)

// Cleanup chart on unmount
onUnmounted(() => {
  if (chart) {
    chart.remove()
    chart = null
    candlestickSeries = null
  }
})

onMounted(async () => {
  const signalId = Number(route.params.id)
  if (!signalId) {
    router.push('/signals')
    return
  }

  try {
    await signalStore.fetchSignalDetail(signalId)
    loading.value = false

    // Initialize chart after data is loaded
    await nextTick()
    if (signal.value?.ohlc_snapshot) {
      initChart()
    }
  } catch (error) {
    console.error('Error loading signal:', error)
    loading.value = false
  }
})
</script>

<template>
  <DefaultLayout>
    <template #header-title>Signal #{{ signal?.id }} - {{ signal?.symbol }}</template>

    <!-- Loading State -->
    <div v-if="loading || isLoading" class="max-w-7xl mx-auto px-4 py-12">
      <div class="flex flex-col items-center justify-center">
        <PhClock :size="48" class="text-blue-500 animate-spin mb-4" />
        <p class="text-gray-500 text-lg font-medium">Loading signal details...</p>
      </div>
    </div>

    <!-- Main Content -->
    <div v-else-if="signal" class="">
      <!-- Action Buttons -->
      <div class="flex justify-between items-center gap-3 mb-6">
        <button
          @click="goBack"
          class="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
        >
          <PhArrowLeft :size="16" />
          Back
        </button>
        <button
          @click="handleDelete"
          class="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-red-700 bg-white border border-red-300 rounded-lg hover:bg-red-50"
        >
          <PhTrash :size="16" />
          Delete
        </button>
      </div>

      <!-- Signal Header Cards -->
      <div class="grid grid-cols-1 md:grid-cols-5 gap-4 mb-6">
        <!-- Signal Category -->
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5">
          <div class="text-sm font-medium text-gray-500 mb-2">Signal</div>
          <span
            :class="getCategoryBadgeClass(signal?.signal_category || 'WAIT')"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold border"
          >
            <PhCurrencyBtc :size="18" />
            {{ signal?.signal_category?.replace('_', ' ') || 'Wait' }}
          </span>
          <div class="mt-3 flex items-center gap-2">
            <span
              :class="
                signal.signal_valid ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
              "
              class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
            >
              {{ signal.signal_valid ? '✓ Valid' : '✗ Invalid' }}
            </span>
          </div>
        </div>

        <!-- Confidence -->
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5">
          <div class="text-sm font-medium text-gray-500 mb-2">Confidence</div>
          <div class="flex items-center gap-3">
            <div
              class="text-3xl font-bold"
              :class="{
                'text-green-600': signal.confidence >= 70,
                'text-yellow-600': signal.confidence >= 50 && signal.confidence < 70,
                'text-red-600': signal.confidence < 50
              }"
            >
              {{ signal.confidence.toFixed(1) }}%
            </div>
          </div>
          <div class="mt-2 bg-gray-200 rounded-full h-2">
            <div
              :class="{
                'bg-green-500': signal.confidence >= 70,
                'bg-yellow-500': signal.confidence >= 50 && signal.confidence < 70,
                'bg-red-500': signal.confidence < 50
              }"
              class="h-2 rounded-full"
              :style="{ width: `${Math.min(signal.confidence, 100)}%` }"
            ></div>
          </div>
        </div>

        <!-- Score -->
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5">
          <div class="text-sm font-medium text-gray-500 mb-2">Total Score</div>
          <div
            class="text-3xl font-bold"
            :class="{
              'text-green-600': signal.total_score >= 50,
              'text-red-600': signal.total_score <= -50,
              'text-gray-600': signal.total_score > -50 && signal.total_score < 50
            }"
          >
            {{ signal.total_score > 0 ? '+' : '' }}{{ signal.total_score.toFixed(2) }}
          </div>
          <div class="mt-2 text-xs text-gray-500">Based on multi-timeframe analysis</div>
        </div>

        <!-- Current Price -->
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5">
          <div class="text-sm font-medium text-gray-500 mb-2">Current Price</div>
          <div class="text-3xl font-bold text-gray-900">
            ${{ formatNumber(signal.current_price) }}
          </div>
          <div class="mt-2 text-xs text-gray-500">Timeframe: {{ signal.primary_timeframe }}</div>
        </div>

        <!-- Risk/Reward -->
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5">
          <div class="text-sm font-medium text-gray-500 mb-2">Risk/Reward</div>
          <div
            class="text-3xl font-bold"
            :class="{
              'text-green-600': signal.risk_reward_ratio >= 2,
              'text-yellow-600': signal.risk_reward_ratio >= 1.5 && signal.risk_reward_ratio < 2,
              'text-red-600': signal.risk_reward_ratio < 1.5
            }"
          >
            1:{{ signal.risk_reward_ratio.toFixed(2) }}
          </div>
          <div class="mt-2 text-xs text-gray-500">Mode: {{ signal.entry_mode }}</div>
        </div>
      </div>

      <!-- Chart -->
      <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5 mb-6">
        <div class="flex items-center gap-2 mb-4">
          <PhChartLine :size="20" class="text-blue-600" />
          <h2 class="text-lg font-semibold text-gray-900">Price Chart</h2>
        </div>
        <div
          ref="chartContainerRef"
          class="w-full rounded-lg overflow-hidden border border-gray-200"
          style="height: 400px"
        ></div>
        <div class="mt-4 flex items-center gap-4 text-xs">
          <div class="flex items-center gap-2">
            <div class="w-6 h-0" style="border-top: 2px dashed #22c55e"></div>
            <span class="text-gray-600">Take Profit</span>
          </div>
          <div class="flex items-center gap-2">
            <div class="w-6 h-0" style="border-top: 2px dashed #ef4444"></div>
            <span class="text-gray-600">Stop Loss</span>
          </div>
          <div class="flex items-center gap-2">
            <div class="w-6 h-0" style="border-top: 2px solid #6366f1"></div>
            <span class="text-gray-600">Entry</span>
          </div>
        </div>
      </div>

      <!-- Trading Plan -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <!-- TP/SL Info -->
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5">
          <div class="flex items-center gap-2 mb-4">
            <PhTrendUp :size="20" class="text-green-600" />
            <h2 class="text-lg font-semibold text-gray-900">Take Profit & Stop Loss</h2>
          </div>
          <div class="space-y-4">
            <div class="flex items-center justify-between p-3 bg-green-50 rounded-lg">
              <div>
                <div class="text-sm font-medium text-green-900">Take Profit</div>
                <div class="text-xs text-green-700">Target profit level</div>
              </div>
              <div class="text-right">
                <div class="text-lg font-bold text-green-900">
                  ${{ formatNumber(signal.tp_price) }}
                </div>
                <div class="text-xs text-green-700">
                  +{{ signal.target_profit_percent.toFixed(2) }}%
                </div>
              </div>
            </div>

            <div class="flex items-center justify-between p-3 bg-red-50 rounded-lg">
              <div>
                <div class="text-sm font-medium text-red-900">Stop Loss</div>
                <div class="text-xs text-red-700">Maximum loss level</div>
              </div>
              <div class="text-right">
                <div class="text-lg font-bold text-red-900">
                  ${{ formatNumber(signal.sl_price) }}
                </div>
                <div class="text-xs text-red-700">-{{ signal.max_risk_percent.toFixed(2) }}%</div>
              </div>
            </div>

            <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
              <div>
                <div class="text-sm font-medium text-gray-900">Max Risk</div>
                <div class="text-xs text-gray-700">Potential loss</div>
              </div>
              <div class="text-right">
                <div class="text-lg font-bold text-gray-900">
                  ${{ formatNumber(signal.max_risk_usdt) }}
                </div>
                <div class="text-xs text-gray-700">
                  {{ signal.max_risk_percent.toFixed(2) }}% of position
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Entry Levels -->
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5">
          <div class="flex items-center gap-2 mb-4">
            <PhCurrencyBtc :size="20" class="text-blue-600" />
            <h2 class="text-lg font-semibold text-gray-900">Entry Levels</h2>
          </div>
          <div class="space-y-3">
            <div
              v-for="entry in signal.entry_levels"
              :key="entry.entry_number"
              class="flex items-center justify-between p-3 bg-blue-50 rounded-lg"
            >
              <div>
                <div class="text-sm font-medium text-blue-900">Entry #{{ entry.entry_number }}</div>
                <div class="text-xs text-blue-700">
                  {{ entry.position_size_percent }}% of position
                </div>
              </div>
              <div class="text-right">
                <div class="text-lg font-bold text-blue-900">
                  ${{ formatNumber(entry.entry_price) }}
                </div>
                <div class="text-xs text-blue-700">
                  {{ formatNumber(entry.position_qty, 6) }} coins
                </div>
              </div>
            </div>

            <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
              <div>
                <div class="text-sm font-medium text-gray-900">Average Entry</div>
                <div class="text-xs text-gray-700">Weighted average</div>
              </div>
              <div class="text-right">
                <div class="text-lg font-bold text-gray-900">
                  ${{ formatNumber(signal.avg_entry_price) }}
                </div>
                <div class="text-xs text-gray-700">
                  Total: ${{ formatNumber(signal.total_position_value) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Support & Resistance -->
      <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5 mb-6">
        <div class="flex items-center gap-2 mb-4">
          <PhTrendUp :size="20" class="text-purple-600" />
          <h2 class="text-lg font-semibold text-gray-900">Support & Resistance</h2>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="p-4 bg-green-50 rounded-lg">
            <div class="text-sm font-medium text-green-900 mb-1">Support Level</div>
            <div class="text-2xl font-bold text-green-900">
              ${{ formatNumber(signal.support_price) }}
            </div>
            <div class="text-xs text-green-700 mt-1">
              {{
                (
                  ((signal.current_price - signal.support_price) / signal.current_price) *
                  100
                ).toFixed(2)
              }}% below current price
            </div>
          </div>
          <div class="p-4 bg-red-50 rounded-lg">
            <div class="text-sm font-medium text-red-900 mb-1">Resistance Level</div>
            <div class="text-2xl font-bold text-red-900">
              ${{ formatNumber(signal.resistance_price) }}
            </div>
            <div class="text-xs text-red-700 mt-1">
              {{
                (
                  ((signal.resistance_price - signal.current_price) / signal.current_price) *
                  100
                ).toFixed(2)
              }}% above current price
            </div>
          </div>
        </div>
      </div>

      <!-- Strategy Configuration -->
      <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5 mb-6">
        <h2 class="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
          <PhChartLine :size="20" class="text-indigo-600" />
          Strategy Configuration
        </h2>

        <div class="space-y-4">
          <!-- Strategy Overview -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div
              class="bg-gradient-to-br from-blue-50 to-indigo-50 rounded-lg p-3 border border-blue-100"
            >
              <p class="text-xs text-blue-600 font-medium mb-1">Strategy Name</p>
              <p class="text-sm font-bold text-gray-900 truncate">
                {{ signal.strategy_snapshot?.name || '-' }}
              </p>
            </div>
            <div class="bg-gray-50 rounded-lg p-3">
              <p class="text-xs text-gray-500 mb-1">Primary Timeframe</p>
              <p class="text-sm font-bold text-gray-900">
                {{ signal.strategy_snapshot?.primary_timeframe || signal.primary_timeframe || '-' }}
              </p>
            </div>
            <div class="bg-gray-50 rounded-lg p-3">
              <p class="text-xs text-gray-500 mb-1">Trading Mode</p>
              <span
                class="inline-block px-2 py-1 text-xs font-bold rounded-full"
                :class="
                  signal.entry_mode === 'AGGRESSIVE'
                    ? 'bg-red-100 text-red-700'
                    : 'bg-blue-100 text-blue-700'
                "
              >
                {{ signal.entry_mode || 'Conservative' }}
              </span>
            </div>
          </div>

          <!-- Money Management -->
          <div>
            <h3 class="text-sm font-bold text-gray-700 mb-3 flex items-center gap-2">
              <span class="w-1 h-4 bg-green-500 rounded-full"></span>
              Money Management
            </h3>
            <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-3">
              <div
                v-for="(value, key) in signal.strategy_snapshot?.mm_config"
                :key="key"
                class="bg-gray-50 rounded-lg p-3"
              >
                <p class="text-xs text-gray-500 mb-1">{{ formatKey(key) }}</p>
                <template v-if="key === 'is_aggressive'">
                  <span
                    class="inline-block mt-1 px-2 py-1 text-xs font-bold rounded-full"
                    :class="value ? 'bg-red-100 text-red-700' : 'bg-blue-100 text-blue-700'"
                  >
                    {{ value ? 'Aggressive' : 'Conservative' }}
                  </span>
                </template>
                <template v-else-if="key.includes('percent') || key.includes('size')">
                  <p class="text-sm font-bold text-gray-900">
                    {{ (Number(value) * 100).toFixed(0) }}%
                  </p>
                </template>
                <template v-else-if="key.includes('ratio') || key.includes('target')">
                  <p class="text-sm font-bold text-gray-900">1:{{ Number(value).toFixed(2) }}</p>
                </template>
                <template v-else>
                  <p class="text-sm font-bold text-gray-900">
                    {{ formatValue(value) }}
                  </p>
                </template>
              </div>
            </div>
          </div>

          <!-- Timeframes -->
          <div
            v-if="
              signal.strategy_snapshot?.timeframes && signal.strategy_snapshot.timeframes.length > 0
            "
          >
            <h3 class="text-sm font-bold text-gray-700 mb-3 flex items-center gap-2">
              <span class="w-1 h-4 bg-purple-500 rounded-full"></span>
              Timeframes
            </h3>
            <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-2">
              <div
                v-for="(tf, index) in signal.strategy_snapshot.timeframes"
                :key="index"
                class="bg-gradient-to-br from-purple-50 to-indigo-50 rounded-lg p-3 border border-purple-100"
              >
                <div class="flex items-center justify-between">
                  <p class="text-sm font-bold text-gray-900">{{ tf.name }}</p>
                  <span class="text-xs font-bold text-purple-600">
                    {{ (tf.weight * 100).toFixed(0) }}%
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Indicators -->
          <div
            v-if="
              signal.strategy_snapshot?.indicator_weights &&
              signal.strategy_snapshot.indicator_weights.length > 0
            "
          >
            <h3 class="text-sm font-bold text-gray-700 mb-3 flex items-center gap-2">
              <span class="w-1 h-4 bg-blue-500 rounded-full"></span>
              Indicators
            </h3>

            <!-- Group indicators by timeframe -->
            <div class="space-y-4">
              <div
                v-for="(indicators, timeframe) in groupedIndicators"
                :key="timeframe"
                class="bg-gradient-to-br from-blue-50/50 to-indigo-50/50 rounded-xl p-4 border border-blue-100"
              >
                <!-- Timeframe Header -->
                <div class="flex items-center gap-3 mb-3 pb-3 border-b border-blue-100">
                  <div
                    class="flex items-center justify-center px-3 py-1.5 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-lg shadow-sm"
                  >
                    <span class="text-sm font-bold tracking-wide">{{
                      timeframe === 'All' ? 'Global' : timeframe
                    }}</span>
                  </div>
                  <div class="flex-1">
                    <p class="text-xs font-medium text-gray-600">Timeframe Analysis</p>
                    <p class="text-xs text-gray-400">{{ indicators.length }} indicator(s)</p>
                  </div>
                </div>

                <!-- Indicators Grid -->
                <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2">
                  <div
                    v-for="(indicator, idx) in indicators"
                    :key="indicator.indicator || idx"
                    class="bg-white rounded-lg p-2.5 border border-gray-200 hover:border-blue-300 hover:shadow-sm transition-all"
                  >
                    <div class="flex items-center gap-2">
                      <!-- Role Badge -->
                      <span
                        class="shrink-0 w-6 h-6 flex items-center justify-center rounded-md text-[10px] font-bold shadow-sm"
                        :class="
                          getRoleBadge(indicator.role).bg + ' ' + getRoleBadge(indicator.role).text
                        "
                        :title="getRoleBadge(indicator.role).title"
                      >
                        {{ getRoleBadge(indicator.role).label }}
                      </span>

                      <!-- Indicator Name & Weight -->
                      <div class="flex-1 min-w-0 flex items-center justify-between gap-2">
                        <p class="text-sm font-bold text-gray-900 truncate flex-1">
                          {{ indicator.indicator || indicator.name || 'Indicator' }}
                        </p>
                        <span
                          class="shrink-0 text-xs font-bold px-1.5 py-0.5 rounded bg-blue-50 text-blue-700 whitespace-nowrap"
                        >
                          {{
                            indicator.role === 'DRIVER'
                              ? (indicator.weight * 100).toFixed(0) + '%'
                              : indicator.weight + 'x'
                          }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </DefaultLayout>
</template>
