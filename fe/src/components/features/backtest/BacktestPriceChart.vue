<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted, computed } from 'vue'
import { createChart, CandlestickSeries, createSeriesMarkers, type IChartApi, type ISeriesApi, type Time } from 'lightweight-charts'
import type { ICandleData, IBacktestTrade } from '@/stores/backtest.store'

interface Props {
  ohlcv: ICandleData[]
  trades: IBacktestTrade[]
}

const props = withDefaults(defineProps<Props>(), {
  ohlcv: () => [],
  trades: () => []
})

// Chart refs
const chartContainerRef = ref<HTMLDivElement | null>(null)
let chart: IChartApi | null = null
let candlestickSeries: ISeriesApi<'Candlestick'> | null = null

// Chart dimensions
const CHART_HEIGHT = 320

// Prepare candlestick data for lightweight-charts
const candleData = computed(() => {
  return props.ohlcv.map(candle => ({
    time: (candle.timestamp / 1000) as Time,
    open: candle.open,
    high: candle.high,
    low: candle.low,
    close: candle.close
  }))
})

// Prepare trade markers
const tradeMarkers = computed(() => {
  const markers: Array<{
    time: Time
    position: 'aboveBar' | 'belowBar' | 'inBar'
    color: string
    shape: 'circle' | 'arrowUp' | 'arrowDown' | 'square'
    text: string
    tooltip?: string
  }> = []

  props.trades.forEach(trade => {
    if (!trade.entries || trade.entries.length === 0 || !trade.exit) return

    // Entry time
    const entryTime = new Date(trade.entry_time).getTime() / 1000
    const exitTime = trade.exit.timestamp 
      ? (typeof trade.exit.timestamp === 'number' ? trade.exit.timestamp : new Date(trade.exit.timestamp).getTime() / 1000)
      : null

    // Entry marker - ALWAYS BELOW BAR
    // Buy: Blue arrowUp, Sell: Red arrowDown
    const entryLabel = trade.side === 'BUY' ? 'Buy' : 'Sell'
    const entryColor = trade.side === 'BUY' ? '#3b82f6' : '#ef4444'
    const entryShape = trade.side === 'BUY' ? 'arrowUp' as const : 'arrowDown' as const
    
    markers.push({
      time: entryTime as Time,
      position: 'belowBar' as const,
      color: entryColor,
      shape: entryShape,
      text: entryLabel,
      tooltip: `${entryLabel} @ ${trade.avg_entry_price.toFixed(2)} (Trade #${trade.trade_num})`
    })

    // Exit marker - ALWAYS ABOVE BAR
    // Shape is OPPOSITE of entry shape
    // Buy entry → TP/SL uses arrowDown
    // Sell entry → TP/SL uses arrowUp
    if (exitTime) {
      const exitReason = trade.exit.reason || 'CLOSED'

      // Map exit reason to user-friendly label
      let exitLabel = 'TP'
      // Color based on entry side:
      // BUY entry: TP=Red (↓), SL=Green (↓)
      // SELL entry: TP=Green (↑), SL=Red (↑)
      let exitColor = trade.side === 'BUY' ? '#ef4444': '#22c55e'

      if (exitReason === 'EXPIRED_TP_HIT' || exitReason === 'EXPIRED_SL_HIT') {
        exitLabel = 'Exp'
        exitColor = '#6b7280'  // Gray for Expired (no fill, PnL=0)
      } else if (exitReason.includes('SL') || exitReason === 'HIT_SL') {
        exitLabel = 'SL'
        // Invert color for SL
        exitColor = trade.side === 'BUY' ? '#22c55e': '#ef4444'
      } else if (exitReason.includes('CLOSED')) {
        exitLabel = 'Close'
        exitColor = trade.pnl >= 0 ? '#22c55e' : '#ef4444'  // Green/Red based on PnL
      } else if (exitReason.includes('DEAD')) {
        exitLabel = 'Dead'
        exitColor = '#6b7280'  // Gray for Dead
      } else if (exitReason === 'REVERSE_SIGNAL') {
        exitLabel = 'Rev'
        exitColor = '#f59e0b'  // Amber for Reverse Signal
      }

      // Shape: OPPOSITE of entry
      // If BUY entry (arrowUp) → Exit uses arrowDown
      // If SELL entry (arrowDown) → Exit uses arrowUp
      const exitShape = trade.side === 'BUY' ? 'arrowDown' as const : 'arrowUp' as const

      markers.push({
        time: exitTime as Time,
        position: 'aboveBar' as const,
        color: exitColor,
        shape: exitShape,
        text: exitLabel,
        tooltip: `${exitLabel} @ ${trade.exit?.price.toFixed(2)} | PnL: ${trade.pnl >= 0 ? '+' : ''}${trade.pnl.toFixed(2)} USDT (${trade.pnl_percent >= 0 ? '+' : ''}${trade.pnl_percent.toFixed(2)}%)`
      })
    }
  })

  return markers
})

// Initialize chart
const initChart = () => {
  if (!chartContainerRef.value) return

  // Create chart
  chart = createChart(chartContainerRef.value, {
    width: chartContainerRef.value.clientWidth,
    height: CHART_HEIGHT,
    layout: {
      background: { color: '#f9fafb' },
      textColor: '#374151'
    },
    grid: {
      vertLines: { color: '#e5e7eb' },
      horzLines: { color: '#e5e7eb' }
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
        bottom: 0.2 // Space for markers
      }
    },
    timeScale: {
      borderColor: '#e5e7eb',
      timeVisible: true,
      secondsVisible: false
    }
  })

  // Create candlestick series
  candlestickSeries = chart.addSeries(CandlestickSeries, {
    upColor: '#22c55e',
    downColor: '#ef4444',
    borderUpColor: '#22c55e',
    borderDownColor: '#ef4444',
    wickUpColor: '#22c55e',
    wickDownColor: '#ef4444'
  })

  // Set data
  if (candleData.value.length > 0) {
    candlestickSeries.setData(candleData.value)
  }

  // Set markers
  if (tradeMarkers.value.length > 0) {
    createSeriesMarkers(candlestickSeries, tradeMarkers.value)
  }

  // Fit content
  chart.timeScale().fitContent()

  // Handle resize
  const handleResize = () => {
    if (chart && chartContainerRef.value) {
      chart.applyOptions({
        width: chartContainerRef.value.clientWidth
      })
    }
  }

  window.addEventListener('resize', handleResize)

  // Store cleanup function
  ;(chartContainerRef.value as any)._cleanup = () => {
    window.removeEventListener('resize', handleResize)
    if (chart) {
      chart.remove()
      chart = null
    }
  }
}

// Update chart data
const updateChartData = () => {
  if (!candlestickSeries || !chart) return

  if (candleData.value.length > 0) {
    candlestickSeries.setData(candleData.value)
    chart.timeScale().fitContent()
  }
}

// Update markers
const updateMarkers = () => {
  if (!candlestickSeries) return

  // Update markers dengan array baru
  createSeriesMarkers(candlestickSeries, tradeMarkers.value)
}

// Cleanup
const cleanup = () => {
  if (chartContainerRef.value && (chartContainerRef.value as any)._cleanup) {
    (chartContainerRef.value as any)._cleanup()
  }
}

// Watch for data changes
watch(() => props.ohlcv, () => {
  updateChartData()
}, { deep: true })

watch(() => props.trades, () => {
  updateMarkers()
}, { deep: true })

// Lifecycle
onMounted(() => {
  initChart()
})

onUnmounted(() => {
  cleanup()
})
</script>

<template>
  <div class="w-full mb-2">
    <div ref="chartContainerRef" class="w-full rounded-xl overflow-hidden"></div>
    
    <!-- Legend -->
    <div class="flex items-center justify-center gap-4 mt-3 text-xs flex-wrap">
      <div class="flex items-center gap-2">
        <div class="w-3 h-3 bg-green-500 rounded"></div>
        <span class="text-gray-600">Bullish</span>
      </div>
      <div class="flex items-center gap-2">
        <div class="w-3 h-3 bg-red-500 rounded"></div>
        <span class="text-gray-600">Bearish</span>
      </div>
      <div class="flex items-center gap-2">
        <div class="w-0 h-0 border-l-4 border-r-4 border-b-8 border-l-transparent border-r-transparent border-b-blue-500"></div>
        <span class="text-gray-600">Buy Entry</span>
      </div>
      <div class="flex items-center gap-2">
        <div class="w-0 h-0 border-l-4 border-r-4 border-t-8 border-l-transparent border-r-transparent border-t-red-500"></div>
        <span class="text-gray-600">Sell Entry</span>
      </div>
      <div class="flex items-center gap-2">
        <span class="text-gray-600">Exit markers show TP/SL/Close above entry</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Lightweight Charts container styles */
:deep(.tv-chart) {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Helvetica Neue', sans-serif !important;
}
</style>
