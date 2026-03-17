<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted } from 'vue'
import { createChart, type IChartApi, type ISeriesApi, CandlestickSeries, type CandlestickData, type Time } from 'lightweight-charts'
import type { ITimeframeRawData, ITradingPlan } from '@/stores/signal-analyze.store'
import { PhChartLine, PhXCircle } from '@phosphor-icons/vue'

interface SignalChartProps {
  chartData: ITimeframeRawData | null
  tradingPlan: ITradingPlan | null
  timeframe: string
  isLoading: boolean
}

const props = defineProps<SignalChartProps>()

// Chart refs
const chartContainerRef = ref<HTMLDivElement | null>(null)
let chart: IChartApi | null = null
let candlestickSeries: ISeriesApi<'Candlestick'> | null = null

// Chart dimensions
const CHART_HEIGHT = 400

// Format price untuk display
const formatPrice = (price: number): string => {
  if (price >= 1000) {
    return price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  }
  return price.toFixed(4)
}

// Initialize chart
const initChart = () => {
  if (!chartContainerRef.value) return

  // Create chart
  chart = createChart(chartContainerRef.value, {
    width: chartContainerRef.value.clientWidth,
    height: CHART_HEIGHT,
    layout: {
      background: { color: '#ffffff' },
      textColor: '#374151'
    },
    grid: {
      vertLines: { color: '#f0f0f0' },
      horzLines: { color: '#f0f0f0' }
    },
    crosshair: {
      mode: 1, // Normal mode - shows OHLC tooltip on hover
      vertLine: {
        width: 1,
        color: '#94a3b8',
        style: 3, // Dotted
        labelBackgroundColor: '#94a3b8'
      },
      horzLine: {
        width: 1,
        color: '#94a3b8',
        style: 3, // Dotted
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
      secondsVisible: false,
      // Timezone offset untuk WIB (UTC+7)
      // Lightweight charts automatically handles timezone display based on timestamp
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
  
  // Trigger update jika data sudah ada saat series dibuat
  if (props.chartData) {
    updateCandlestickData()
  }

  // Handle resize
  const handleResize = () => {
    if (chart && chartContainerRef.value) {
      chart.applyOptions({
        width: chartContainerRef.value.clientWidth
      })
    }
  }

  window.addEventListener('resize', handleResize)

  // Cleanup on unmount
  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
    if (chart) {
      chart.remove()
      chart = null
    }
  })
}

// Update candlestick data
const updateCandlestickData = () => {
  if (!candlestickSeries || !props.chartData) return

  const rawData = props.chartData.raws || []

  // Convert data untuk lightweight charts dengan timezone adjustment (UTC+7 untuk WIB)
  const TIMEZONE_OFFSET = 7 * 60 * 60 * 1000 // 7 jam dalam milliseconds
  const candleData: CandlestickData<Time>[] = rawData.map(raw => ({
    time: Math.floor((raw.timestamp + TIMEZONE_OFFSET) / 1000) as Time, // Convert ms to seconds dengan offset
    open: raw.open,
    high: raw.high,
    low: raw.low,
    close: raw.close
  }))

  candlestickSeries.setData(candleData)

  // Fit content
  if (chart) {
    chart.timeScale().fitContent()
  }
  
  // Trigger price lines update setelah data di-set
  if (props.tradingPlan) {
    updatePriceLines()
  }
}

// Update TP/SL lines
const updatePriceLines = () => {
  if (!candlestickSeries || !props.tradingPlan) return

  const tpPrice = props.tradingPlan.take_profit
  const slPrice = props.tradingPlan.stop_loss
  const resistancePrice = props.tradingPlan.resistance
  const supportPrice = props.tradingPlan.support
  const entryPrices = props.tradingPlan.entries.map(e => e.entry_price)

  // TP Line - Green Dashed
  candlestickSeries.createPriceLine({
    price: tpPrice,
    color: '#22c55e',
    lineWidth: 2,
    lineStyle: 2, // Dashed
    axisLabelVisible: true,
    title: `TP: ${formatPrice(tpPrice)}`,
    axisLabelColor: '#22c55e',
    axisLabelTextColor: '#ffffff'
  })

  // SL Line - Red Dashed
  candlestickSeries.createPriceLine({
    price: slPrice,
    color: '#ef4444',
    lineWidth: 2,
    lineStyle: 2, // Dashed
    axisLabelVisible: true,
    title: `SL: ${formatPrice(slPrice)}`,
    axisLabelColor: '#ef4444',
    axisLabelTextColor: '#ffffff'
  })

  // Resistance Line - Orange Dotted
  candlestickSeries.createPriceLine({
    price: resistancePrice,
    color: '#f97316',
    lineWidth: 2,
    lineStyle: 3, // Dotted
    axisLabelVisible: true,
    title: `Resistance: ${formatPrice(resistancePrice)}`,
    axisLabelColor: '#f97316',
    axisLabelTextColor: '#ffffff'
  })

  // Support Line - Blue Dotted
  candlestickSeries.createPriceLine({
    price: supportPrice,
    color: '#3b82f6',
    lineWidth: 2,
    lineStyle: 3, // Dotted
    axisLabelVisible: true,
    title: `Support: ${formatPrice(supportPrice)}`,
    axisLabelColor: '#3b82f6',
    axisLabelTextColor: '#ffffff'
  })

  // Entry Lines - Blue Solid (bisa multiple)
  entryPrices.forEach((price, index) => {
    if (candlestickSeries) {
      candlestickSeries.createPriceLine({
        price: price,
        color: '#6366f1',
        lineWidth: 2,
        lineStyle: 0, // Solid
        axisLabelVisible: true,
        title: `Entry ${index + 1}: ${formatPrice(price)}`,
        axisLabelColor: '#6366f1',
        axisLabelTextColor: '#ffffff'
      })
    }
  })
}

// Watch untuk data changes (tanpa immediate, tunggu chart ready)
watch(
  () => props.chartData,
  (newData) => {
    if (newData && candlestickSeries) {
      updateCandlestickData()
    }
  }
)

watch(
  () => props.tradingPlan,
  () => {
    if (props.tradingPlan && candlestickSeries) {
      updatePriceLines()
    }
  },
  { immediate: true }
)

// Initialize chart on mount
onMounted(() => {
  initChart()
})
</script>

<template>
  <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
    <div class="flex items-center gap-3 mb-4">
      <div class="p-3 bg-cyan-50 rounded-xl">
        <PhChartLine :size="28" class="text-cyan-600" weight="fill" />
      </div>
      <div class="flex-1">
        <h2 class="text-xl font-bold text-gray-900">Price Chart</h2>
        <p class="text-sm text-gray-500">{{ timeframe }} • Entry, TP & SL levels</p>
      </div>

      <!-- Legend -->
      <div class="flex items-center gap-4 text-xs flex-wrap">
        <div class="flex items-center gap-2">
          <div class="w-6 h-0.5 bg-indigo-500"></div>
          <span class="text-gray-600">Entry</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-6 h-0.5 bg-green-500" style="border-top: 2px dashed #22c55e"></div>
          <span class="text-gray-600">TP</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-6 h-0.5 bg-red-500" style="border-top: 2px dashed #ef4444"></div>
          <span class="text-gray-600">SL</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-6 h-0.5 bg-orange-500" style="border-top: 2px dotted #f97316"></div>
          <span class="text-gray-600">Resistance</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-6 h-0.5 bg-blue-500" style="border-top: 2px dotted #3b82f6"></div>
          <span class="text-gray-600">Support</span>
        </div>
      </div>
    </div>

    <!-- Chart Container -->
    <div class="relative">
      <!-- Loading State -->
      <div
        v-if="isLoading"
        class="absolute inset-0 bg-white/80 backdrop-blur-sm z-10 flex items-center justify-center rounded-lg"
      >
        <div class="text-center">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-3"></div>
          <p class="text-gray-600 font-medium">Loading chart data...</p>
        </div>
      </div>

      <!-- Empty State -->
      <div
        v-else-if="!chartData || !chartData.raws || chartData.raws.length === 0"
        class="absolute inset-0 bg-white/80 backdrop-blur-sm z-10 flex items-center justify-center rounded-lg"
      >
        <div class="text-center">
          <PhXCircle :size="48" class="mx-auto text-gray-300 mb-3" />
          <p class="text-gray-500 font-medium">No chart data available</p>
          <p class="text-gray-400 text-sm mt-1">Run an analysis to load the chart</p>
        </div>
      </div>

      <!-- Chart -->
      <div
        ref="chartContainerRef"
        class="rounded-lg overflow-hidden border border-gray-200"
        style="height: 400px"
      ></div>
    </div>

    <!-- Info Text -->
    <div class="mt-4 p-3 bg-blue-50 rounded-lg border border-blue-200">
      <p class="text-xs text-blue-800">
        <span class="font-semibold">💡 Tip:</span>
        Hover on the chart to see price details. Use mouse wheel to zoom, drag to pan.
        TP/SL lines are automatically calculated based on your trading plan.
      </p>
    </div>
  </div>
</template>
