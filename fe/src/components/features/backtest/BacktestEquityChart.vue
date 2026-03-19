<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted, computed } from 'vue'
import { createChart, LineSeries, type IChartApi, type ISeriesApi, type Time } from 'lightweight-charts'
import type { IEquityPoint } from '@/stores/backtest.store'

interface Props {
  equityCurve: IEquityPoint[]
}

const props = withDefaults(defineProps<Props>(), {
  equityCurve: () => []
})

// Chart refs
const chartContainerRef = ref<HTMLDivElement | null>(null)
let chart: IChartApi | null = null
let lineSeries: ISeriesApi<'Line'> | null = null

// Chart dimensions
const CHART_HEIGHT = 250

// Prepare equity curve data for lightweight-charts
const equityData = computed(() => {
  return props.equityCurve.map(point => ({
    time: (point.timestamp / 1000) as Time,
    value: point.balance
  }))
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
        bottom: 0.1
      }
    },
    timeScale: {
      borderColor: '#e5e7eb',
      timeVisible: true,
      secondsVisible: false
    }
  })

  // Create line series
  lineSeries = chart.addSeries(LineSeries, {
    color: '#3b82f6',
    lineWidth: 2,
    lastValueVisible: true,
    priceLineVisible: true,
    priceLineSource: 0,
    priceLineWidth: 1,
    priceLineColor: '#3b82f6',
    priceLineStyle: 2,
    baseLineVisible: false,
    baseLineWidth: 1,
    baseLineColor: '#3b82f6',
    baseLineStyle: 0
  })

  // Set data
  if (equityData.value.length > 0) {
    lineSeries.setData(equityData.value)
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
      lineSeries = null
    }
  }
}

// Update chart data
const updateChartData = () => {
  if (!lineSeries || !chart) return

  if (equityData.value.length > 0) {
    lineSeries.setData(equityData.value)
    chart.timeScale().fitContent()
  }
}

// Cleanup
const cleanup = () => {
  if (chartContainerRef.value && (chartContainerRef.value as any)._cleanup) {
    (chartContainerRef.value as any)._cleanup()
  }
}

// Watch for data changes
watch(() => props.equityCurve, () => {
  updateChartData()
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
  <div class="w-full">
    <div class="flex items-center gap-2 mb-3">
      <PhChartLineUp :size="20" class="text-blue-500" weight="fill" />
      <h3 class="text-base font-bold text-gray-900">Equity Curve</h3>
    </div>
    <div ref="chartContainerRef" class="w-full rounded-xl overflow-hidden"></div>
  </div>
</template>

<style scoped>
/* Lightweight Charts container styles */
:deep(.tv-chart) {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Helvetica Neue', sans-serif !important;
}
</style>
