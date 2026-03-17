<script setup lang="ts">
import { computed } from 'vue'
import type { ISignalAnalyzeResponse } from '@/stores/signal-analyze.store'
import {
  PhCheckCircle,
  PhXCircle,
  PhTrendUp,
  PhCurrencyBtc
} from '@phosphor-icons/vue'

interface SignalInfoCardProps {
  result: ISignalAnalyzeResponse | null
  symbol: string
  strategyName: string
  strategyTimeframe: string
}

const props = defineProps<SignalInfoCardProps>()

// Helper untuk menentukan warna signal
const getSignalColor = (signal: string): string => {
  const signalUpper = signal.toUpperCase()
  if (signalUpper.includes('STRONG_BUY') || signalUpper === 'BUY') return 'text-green-600 bg-green-50 border-green-200'
  if (signalUpper.includes('STRONG_SELL') || signalUpper === 'SELL') return 'text-red-600 bg-red-50 border-red-200'
  return 'text-gray-600 bg-gray-50 border-gray-200'
}

// Helper untuk format price
const formatPrice = (price: number): string => {
  if (price >= 1000) {
    return price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  }
  return price.toFixed(4)
}

// Get signal badge class
const signalBadgeClass = computed(() => {
  if (!props.result) return ''
  return getSignalColor(props.result.signal.signal)
})

// Confidence color
const confidenceColor = computed(() => {
  if (!props.result) return 'bg-gray-500'
  const conf = props.result.scoring.confidence
  if (conf >= 70) return 'bg-green-500'
  if (conf >= 50) return 'bg-yellow-500'
  return 'bg-red-500'
})
</script>

<template>
  <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-5">
    <div class="flex items-center gap-3 mb-4">
      <div class="p-2.5 bg-purple-50 rounded-xl">
        <PhCurrencyBtc :size="24" class="text-purple-600" weight="fill" />
      </div>
      <div class="min-w-0">
        <h2 class="text-lg font-bold text-gray-900 truncate">Signal Info</h2>
        <p class="text-xs text-gray-500">Current trading signal analysis</p>
      </div>
    </div>

    <div v-if="result" class="space-y-3">
      <!-- Signal Badge - Compact -->
      <div
        class="flex items-center justify-between p-3 rounded-lg border-2"
        :class="signalBadgeClass"
      >
        <div class="flex items-center gap-2">
          <PhCheckCircle
            v-if="result.signal.valid"
            :size="22"
            weight="fill"
            class="text-green-600"
          />
          <PhXCircle
            v-else
            :size="22"
            weight="fill"
            class="text-red-600"
          />
          <div>
            <p class="text-xs font-medium uppercase opacity-70">Signal</p>
            <p class="text-lg font-bold">{{ result.signal.signal }}</p>
          </div>
        </div>
        <div class="text-right">
          <p class="text-xs font-medium uppercase opacity-70">Valid</p>
          <p class="text-sm font-semibold">
            {{ result.signal.valid ? 'Yes' : 'No' }}
          </p>
        </div>
      </div>

      <!-- Price & Timeframe - Compact -->
      <div class="grid grid-cols-2 gap-2">
        <div class="p-2.5 bg-gradient-to-br from-blue-50 to-blue-100 rounded-lg border border-blue-200">
          <div class="flex items-center gap-1.5 mb-1">
            <PhCurrencyBtc :size="16" class="text-blue-600" weight="fill" />
            <span class="text-xs font-medium text-blue-700 uppercase">Price</span>
          </div>
          <p class="text-base font-bold text-blue-900">
            {{ result.signal.current_price ? formatPrice(result.signal.current_price) : '-' }}
          </p>
        </div>

        <!-- Analysis Timeframe -->
        <div class="p-2.5 bg-gradient-to-br from-indigo-50 to-indigo-100 rounded-lg border border-indigo-200">
          <div class="flex items-center gap-1.5 mb-1">
            <PhTrendUp :size="16" class="text-indigo-600" weight="bold" />
            <span class="text-xs font-medium text-indigo-700 uppercase">TF</span>
          </div>
          <p class="text-base font-bold text-indigo-900">
            {{ result.primary_timeframe }}
          </p>
        </div>
      </div>

      <!-- Confidence Meter - Compact -->
      <div class="p-3 bg-gray-50 rounded-lg border border-gray-200">
        <div class="flex items-center justify-between mb-1.5">
          <span class="text-xs font-medium text-gray-700">Confidence</span>
          <span class="text-sm font-bold text-gray-900">{{ result.scoring.confidence.toFixed(1) }}%</span>
        </div>
        <div class="relative h-2.5 bg-gray-200 rounded-full overflow-hidden">
          <div
            class="absolute top-0 left-0 h-full rounded-full transition-all duration-500"
            :class="confidenceColor"
            :style="{ width: `${result.scoring.confidence}%` }"
          ></div>
        </div>
      </div>

      <!-- Timestamp -->
      <div class="pt-2 border-t border-gray-100">
        <p class="text-xs text-gray-500 text-center mt-2">
          {{ new Date(result.timestamp).toLocaleString() }}
        </p>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="py-8">
      <div class="text-center">
        <PhCurrencyBtc :size="40" class="mx-auto text-gray-300 mb-2" />
        <p class="text-gray-500 text-sm">No signal data yet</p>
        <p class="text-gray-400 text-xs mt-1">Run an analysis to see the signal</p>
      </div>
    </div>
  </div>
</template>
