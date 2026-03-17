<script setup lang="ts">
import type { IScoringBreakdown } from '@/stores/signal-analyze.store'
import { PhChartBar, PhTrendUp, PhTrendDown, PhEquals } from '@phosphor-icons/vue'

interface ScoringBreakdownTableProps {
  scoring: IScoringBreakdown | null
}

const props = defineProps<ScoringBreakdownTableProps>()

// Helper untuk format score (data dari BE sudah dalam bentuk 0-100)
const formatScore = (score: number): string => {
  return score.toFixed(1)
}

// Helper untuk weight percent (data dari BE sudah dalam bentuk 0-1)
const formatWeight = (weight: number): string => {
  return `${(weight * 100).toFixed(0)}%`
}

// Trend icon dan color
const getTrendInfo = (trend: string) => {
  const trendUpper = trend.toUpperCase()
  if (trendUpper === 'BULLISH') {
    return { icon: PhTrendUp, color: 'text-green-600', bg: 'bg-green-50' }
  }
  if (trendUpper === 'BEARISH') {
    return { icon: PhTrendDown, color: 'text-red-600', bg: 'bg-red-50' }
  }
  return { icon: PhEquals, color: 'text-gray-600', bg: 'bg-gray-50' }
}
</script>

<template>
  <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
    <div class="flex items-center gap-3 mb-6">
      <div class="p-3 bg-orange-50 rounded-xl">
        <PhChartBar :size="28" class="text-orange-600" weight="fill" />
      </div>
      <div>
        <h2 class="text-xl font-bold text-gray-900">Scoring Breakdown</h2>
        <p class="text-sm text-gray-500">Signal calculation per indicator</p>
      </div>
    </div>

    <div v-if="scoring && scoring.breakdown.length > 0" class="space-y-6">
      <!-- Overall Score -->
      <div class="p-4 bg-gradient-to-br from-blue-50 to-blue-100 rounded-xl border border-blue-200">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-blue-700 mb-1">Total Score</p>
            <p class="text-3xl font-bold text-blue-900">{{ formatScore(scoring.totalScore) }} / 100</p>
          </div>
          <div class="text-right">
            <p class="text-sm font-medium text-blue-700 mb-1">Confidence</p>
            <p class="text-3xl font-bold text-blue-900">{{ scoring.confidence.toFixed(1) }}%</p>
          </div>
        </div>
      </div>

      <!-- Timeframes -->
      <div v-for="(tf, index) in scoring.breakdown" :key="index" class="space-y-3">
        <!-- Timeframe Header -->
        <div class="flex items-center justify-between pb-2.5 border-b-2 border-gray-200">
          <div class="flex items-center gap-3">
            <!-- Timeframe Badge -->
            <div class="flex items-center justify-center px-3 py-1.5 bg-gradient-to-br from-blue-500 to-blue-600 text-white rounded-lg shadow-sm">
              <span class="text-sm font-bold tracking-wide">{{ tf.timeframe }}</span>
            </div>
            
            <!-- Trend Icon -->
            <div class="flex items-center justify-center w-8 h-8 rounded-lg"
                 :class="{
                   'bg-green-100': tf.trend.toUpperCase().includes('BULL'),
                   'bg-red-100': tf.trend.toUpperCase().includes('BEAR'),
                   'bg-gray-100': !tf.trend.toUpperCase().includes('BULL') && !tf.trend.toUpperCase().includes('BEAR')
                 }">
              <component
                :is="getTrendInfo(tf.trend).icon"
                :size="18"
                :class="getTrendInfo(tf.trend).color"
                weight="fill"
              />
            </div>
            
            <!-- Trend Name -->
            <div>
              <p class="text-base font-bold text-gray-900 leading-tight">{{ tf.trend }}</p>
              <p class="text-xs text-gray-500 mt-0.5">Timeframe Analysis</p>
            </div>
          </div>
          
          <!-- Contribution -->
          <div class="text-right">
            <p class="text-xs text-gray-500 font-medium uppercase tracking-wider mb-0.5">Contribution</p>
            <p class="text-2xl font-bold text-gray-900 leading-none">{{ tf.contribution.toFixed(2) }}</p>
          </div>
        </div>

        <!-- Indicators - Compact Grid Card Layout -->
        <div class="grid grid-cols-3 gap-2">
          <div
            v-for="(indicator, idx) in tf.indicator"
            :key="idx"
            class="p-2.5 bg-gray-50 rounded border border-gray-200"
          >
            <!-- Name + Contribution (Same Row) -->
            <div class="flex items-center justify-between mb-3">
              <p class="font-bold text-gray-900 text-base truncate flex-1" :title="indicator.name">
                {{ indicator.name }}
              </p>
              <span
                class="px-2 py-0.5 rounded text-sm font-bold ml-2 flex-shrink-0"
                :class="indicator.contribution >= 0 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'"
              >
                {{ indicator.contribution.toFixed(2) }}
              </span>
            </div>

            <!-- Weight Bar -->
            <div class="relative h-2.5 bg-gray-200 rounded-full overflow-hidden mb-2.5">
              <div
                class="absolute top-0 left-0 h-full rounded-full bg-blue-500"
                :style="{ width: `${indicator.weight * 100}%` }"
              ></div>
            </div>

            <!-- Score + Zone + Weight (Below Bar) -->
            <div class="flex items-center justify-between text-xs">
              <div>
                <span class="text-gray-500">Score:</span>
                <span class="font-semibold text-gray-900 ml-1">{{ formatScore(indicator.rawSignal) }}</span>
              </div>
              <div class="flex items-center gap-2">
                <span
                  v-if="indicator.zone"
                  class="px-1.5 py-0.5 rounded text-xs font-medium"
                  :class="{
                    'bg-green-100 text-green-700': indicator.zone.toLowerCase().includes('bull') || indicator.zone.toLowerCase().includes('buy'),
                    'bg-red-100 text-red-700': indicator.zone.toLowerCase().includes('bear') || indicator.zone.toLowerCase().includes('sell'),
                    'bg-gray-100 text-gray-700': !indicator.zone.toLowerCase().includes('bull') && !indicator.zone.toLowerCase().includes('bear') && !indicator.zone.toLowerCase().includes('buy') && !indicator.zone.toLowerCase().includes('sell')
                  }"
                >
                  {{ indicator.zone }}
                </span>
                <span class="text-gray-500">W:<span class="font-semibold text-gray-900 ml-0.5">{{ formatWeight(indicator.weight) }}</span></span>
              </div>
            </div>

            <!-- Details (Below, if exists) -->
            <div v-if="indicator.details && indicator.details.length > 0" class="mt-2 pt-2 border-t border-dashed border-gray-300">
              <p class="text-xs text-gray-500 line-clamp-2">
                {{ indicator.details.join(', ') }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-12">
      <PhChartBar :size="48" class="mx-auto text-gray-300 mb-3" />
      <p class="text-gray-500 text-sm">No scoring data yet</p>
      <p class="text-gray-400 text-xs mt-1">Run an analysis to see the breakdown</p>
    </div>
  </div>
</template>
