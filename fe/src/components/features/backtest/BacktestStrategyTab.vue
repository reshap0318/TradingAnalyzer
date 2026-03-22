<script setup lang="ts">
import { computed } from 'vue'
import { PhChartLineUp, PhClock, PhCurrencyDollar, PhTrophy } from '@phosphor-icons/vue'
import { configKeyToLabel, formatConfigValue } from '@/helpers/config'
import type { IBacktestDetail } from '@/stores/backtest.store'

interface Props {
  backtest: IBacktestDetail | null
}

const props = defineProps<Props>()

// Calculate total weights
const totalTimeframeWeight = computed(() => {
  if (!props.backtest?.strategy?.timeframes) return 0
  return props.backtest.strategy.timeframes.reduce((sum, tf) => sum + tf.weight, 0)
})

const totalIndicatorWeight = computed(() => {
  if (!props.backtest?.strategy?.indicator_weights) return 0
  return props.backtest.strategy.indicator_weights.reduce((sum, ind) => sum + ind.weight, 0)
})

const groupedIndicators = computed(() => {
  if (!props.backtest?.strategy?.indicator_weights) return {}
  const groups: Record<string, typeof props.backtest.strategy.indicator_weights> = {}
  props.backtest.strategy.indicator_weights.forEach(ind => {
    const tfName = ind.tf || 'All Timeframes'
    if (!groups[tfName]) {
      groups[tfName] = []
    }
    groups[tfName].push(ind)
  })
  return groups
})

const getRoleBadge = (role?: string) => {
  if (!role) role = 'DRIVER'
  const upper = role.toUpperCase()
  switch (upper) {
    case 'DRIVER':  return { label: 'D', bg: 'bg-blue-100', text: 'text-blue-700', title: 'Driver' }
    case 'FILTER':  return { label: 'F', bg: 'bg-amber-100', text: 'text-amber-700', title: 'Filter' }
    case 'BOOSTER': return { label: 'B', bg: 'bg-purple-100', text: 'text-purple-700', title: 'Booster' }
    default:        return { label: '?', bg: 'bg-gray-100', text: 'text-gray-700', title: 'Unknown' }
  }
}
</script>

<template>
  <div class="space-y-4">
    <!-- Strategy Overview -->
    <div class="bg-white border border-gray-200 rounded-xl p-4">
      <h3 class="text-base font-bold text-gray-700 mb-3 flex items-center gap-2">
        <PhTrophy :size="24" class="text-blue-500" weight="fill" />
        Strategy Overview
      </h3>
      <div class="grid grid-cols-3 gap-3">
        <div class="bg-gray-50 rounded-lg p-3">
          <p class="text-xs text-gray-500 mb-1">Strategy Name</p>
          <p class="text-sm font-semibold text-gray-900 truncate">
            {{ backtest?.strategy?.strategy_name || '-' }}
          </p>
        </div>
        <div class="bg-gray-50 rounded-lg p-3">
          <p class="text-xs text-gray-500 mb-1">Primary Timeframe</p>
          <p class="text-sm font-semibold text-gray-900">
            {{ backtest?.strategy?.primary_tf || '-' }}
          </p>
        </div>
        <div class="bg-gray-50 rounded-lg p-3">
          <p class="text-xs text-gray-500 mb-1">Status</p>
          <span
            class="inline-block px-2 py-1 text-xs font-medium rounded-full"
            :class="backtest?.strategy?.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'"
          >
            {{ backtest?.strategy?.is_active ? 'Active' : 'Inactive' }}
          </span>
        </div>
      </div>
    </div>

    <!-- Money Management -->
    <div class="bg-white border border-gray-200 rounded-xl p-4">
      <h3 class="text-base font-bold text-gray-700 mb-3 flex items-center gap-2">
        <PhCurrencyDollar :size="24" class="text-green-500" weight="fill" />
        Money Management
      </h3>
      <div class="grid grid-cols-3 md:grid-cols-6 gap-3">
        <div
          v-for="(value, key) in backtest?.strategy?.money_management"
          :key="key"
          class="bg-gray-50 rounded-lg p-3"
        >
          <p class="text-xs text-gray-500 mb-1">{{ configKeyToLabel(key.toUpperCase()) }}</p>
          <template v-if="key === 'is_agressive'">
            <span
              class="inline-block mt-1 px-2 py-1 text-xs font-medium rounded-full"
              :class="value ? 'bg-red-100 text-red-700' : 'bg-blue-100 text-blue-700'"
            >
              {{ value ? 'Aggressive' : 'Conservative' }}
            </span>
          </template>
          <template v-else>
            <p class="text-sm font-semibold text-gray-900">
              {{ formatConfigValue(key.toUpperCase(), Number(value) || 0) }}
            </p>
          </template>
        </div>
      </div>
    </div>

    <!-- Timeframes -->
    <div class="bg-white border border-gray-200 rounded-xl p-4">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-base font-bold text-gray-900 flex items-center gap-2">
          <PhClock :size="24" class="text-purple-500" weight="fill" />
          Timeframes
        </h3>
        <div class="flex items-center gap-2">
          <span class="text-xs text-gray-500">Total:</span>
          <span
            class="px-2 py-1 text-xs font-bold rounded-full"
            :class="totalTimeframeWeight === 1 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'"
          >
            {{ (totalTimeframeWeight * 100).toFixed(0) }}%
          </span>
        </div>
      </div>
      <div v-if="backtest?.strategy?.timeframes && backtest.strategy.timeframes.length > 0" class="grid grid-cols-1 md:grid-cols-3 gap-3">
        <div
          v-for="(tf, index) in backtest.strategy.timeframes"
          :key="index"
          class="bg-gray-50 rounded-lg p-3"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <div class="w-7 h-7 bg-primary/10 rounded-lg flex items-center justify-center">
                <span class="text-xs font-bold text-primary">{{ index + 1 }}</span>
              </div>
              <p class="text-base font-bold text-gray-900">{{ tf.tf }}</p>
            </div>
            <span class="text-sm font-bold text-primary">{{ (tf.weight * 100).toFixed(0) }}%</span>
          </div>
          <!-- <p class="text-xs text-gray-500">
            {{ tf.timeframe_detail?.name || 'Timeframe' }}
            <span v-if="tf.timeframe_detail?.in_minutes">
              • {{ tf.timeframe_detail.in_minutes }} min
            </span>
          </p> -->
        </div>
      </div>
      <div v-else class="text-center py-8 text-gray-400">
        No timeframe data available
      </div>
    </div>

    <!-- Indicators -->
    <div class="bg-white border border-gray-200 rounded-xl p-4">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-base font-bold text-gray-900 flex items-center gap-2">
          <PhChartLineUp :size="20" class="text-blue-500" weight="fill" />
          Indicators
        </h3>
      </div>
      <div v-if="backtest?.strategy?.indicator_weights && backtest.strategy.indicator_weights.length > 0" class="space-y-6">
        <div v-for="(indicators, tf) in groupedIndicators" :key="tf" class="space-y-3">
          <!-- Timeframe Header matching SignalAnalyzePage -->
          <div class="flex items-center justify-between pb-2.5 border-b-2 border-gray-100">
            <div class="flex items-center gap-3">
              <div class="flex items-center justify-center px-3 py-1.5 bg-linear-to-br from-blue-500 to-blue-600 text-white rounded-lg shadow-sm">
                <span class="text-sm font-bold tracking-wide">{{ tf === 'All Timeframes' ? 'Global' : tf }}</span>
              </div>
              <div>
                <p class="text-sm font-bold text-gray-900 leading-tight">Timeframe Analysis</p>
                <p class="text-xs text-gray-500 mt-0.5">{{ indicators.length }} indicator(s)</p>
              </div>
            </div>
          </div>
          
          <!-- Indicators Grid matching ScoringBreakdownTable -->
          <div class="grid grid-cols-1 md:grid-cols-4 gap-2">
            <div
              v-for="(indicator, idx) in indicators"
              :key="idx"
              class="p-2.5 bg-gray-50 rounded border border-gray-200"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-1.5 flex-1 min-w-0">
                  <span
                    class="shrink-0 w-5 h-5 flex items-center justify-center rounded text-[10px] font-bold"
                    :class="[getRoleBadge(indicator.indicator_detail?.role).bg, getRoleBadge(indicator.indicator_detail?.role).text]"
                    :title="getRoleBadge(indicator.indicator_detail?.role).title"
                  >
                    {{ getRoleBadge(indicator.indicator_detail?.role).label }}
                  </span>
                  <p class="text-sm font-bold text-gray-900 truncate flex-1">
                    {{ indicator.indicator_detail?.name || indicator.indicator_detail?.indicator || 'Indicator' }}
                  </p>
                </div>
                <div class="text-right shrink-0 ml-2">
                  <span class="text-sm font-bold text-blue-600">
                    <template v-if="indicator.indicator_detail?.role == 'DRIVER' || !indicator.indicator_detail?.role">
                      {{ (indicator.weight * 100).toFixed(0) }}%
                    </template>
                    <template v-else>
                      {{ indicator.weight }}X
                    </template>
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="text-center py-8 text-gray-400">
        No indicator data available
      </div>
    </div>
  </div>
</template>
