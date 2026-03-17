<script setup lang="ts">
import { computed } from 'vue'
import type { ITradingPlan } from '@/stores/signal-analyze.store'
import {
  PhTarget,
  PhStopCircle,
  PhTrendUp,
  PhTrendDown,
  PhCurrencyDollar,
  PhChartLine
} from '@phosphor-icons/vue'

interface TradingPlanCardProps {
  tradingPlan: ITradingPlan | null
}

const props = defineProps<TradingPlanCardProps>()

// Helper untuk format price
const formatPrice = (price: number): string => {
  if (price >= 1000) {
    return price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  }
  return price.toFixed(4)
}

// Helper untuk format percent
const formatPercent = (percent: number): string => {
  return `${percent.toFixed(2)}%`
}

// Helper untuk format USD
const formatUSD = (amount: number): string => {
  return `$${amount.toFixed(2)}`
}

// Mode badge color
const modeBadgeClass = computed(() => {
  if (!props.tradingPlan) return ''
  return props.tradingPlan.mode === 'CONSERVATIVE'
    ? 'bg-blue-100 text-blue-700 border-blue-200'
    : 'bg-orange-100 text-orange-700 border-orange-200'
})

// Risk/Reward color
const rrColor = computed(() => {
  if (!props.tradingPlan) return 'text-gray-600'
  if (props.tradingPlan.risk_reward_ratio >= 1) return 'text-green-600'
  if (props.tradingPlan.risk_reward_ratio >= 0.5) return 'text-yellow-600'
  return 'text-red-600'
})
</script>

<template>
  <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-4 h-full flex flex-col">
    <div class="flex items-center gap-2 mb-3 flex-shrink-0">
      <div class="p-2 bg-green-50 rounded-lg">
        <PhChartLine :size="20" class="text-green-600" weight="fill" />
      </div>
      <div class="flex-1 min-w-0">
        <h2 class="text-base font-bold text-gray-900 truncate">Trading Plan</h2>
        <p class="text-xs text-gray-500">Entry, TP, SL and risk management</p>
      </div>
    </div>

    <div v-if="tradingPlan" class="flex-1 overflow-y-auto space-y-3 pr-1 -mr-1">
      <!-- Mode Badge & R/R -->
      <div class="flex items-center justify-between gap-2">
        <span
          class="px-2 py-1 text-xs font-semibold rounded border flex-shrink-0"
          :class="modeBadgeClass"
        >
          {{ tradingPlan.mode }}
        </span>
        <div class="flex items-center gap-2 text-xs flex-shrink-0">
          <span class="text-gray-500">R/R:</span>
          <span class="font-bold" :class="rrColor">{{ tradingPlan.risk_reward_ratio.toFixed(2) }}</span>
          <span class="text-gray-500">Buffer:</span>
          <span class="font-semibold text-gray-900">{{ formatPercent(tradingPlan.buffer_percent) }}</span>
        </div>
      </div>

      <!-- TP / SL - Compact -->
      <div class="grid grid-cols-2 gap-3">
        <!-- Take Profit -->
        <div class="p-3 bg-gradient-to-br from-green-50 to-green-100 rounded-lg border-2 border-green-200">
          <div class="flex items-center gap-1.5 mb-1.5">
            <PhTarget :size="16" class="text-green-600" weight="fill" />
            <span class="text-xs font-bold text-green-700 uppercase">TP</span>
          </div>
          <p class="text-lg font-bold text-green-900">{{ formatPrice(tradingPlan.take_profit) }}</p>
          <p class="text-xs text-green-600 mt-0.5">
            +{{ formatUSD(tradingPlan.summary?.target_profit_usdt || 0) }}
          </p>
        </div>

        <!-- Stop Loss -->
        <div class="p-3 bg-gradient-to-br from-red-50 to-red-100 rounded-lg border-2 border-red-200">
          <div class="flex items-center gap-1.5 mb-1.5">
            <PhStopCircle :size="16" class="text-red-600" weight="fill" />
            <span class="text-xs font-bold text-red-700 uppercase">SL</span>
          </div>
          <p class="text-lg font-bold text-red-900">{{ formatPrice(tradingPlan.stop_loss) }}</p>
          <p class="text-xs text-red-600 mt-0.5">
            -{{ formatUSD(tradingPlan.summary?.max_risk_usdt || 0) }}
          </p>
        </div>
      </div>

      <!-- Support & Resistance - Compact -->
      <div class="grid grid-cols-2 gap-2">
        <div class="p-2 bg-blue-50 rounded border border-blue-200">
          <div class="flex items-center gap-1">
            <PhTrendUp :size="14" class="text-blue-600" weight="bold" />
            <span class="text-xs font-medium text-blue-700">Resistance</span>
          </div>
          <p class="text-base font-bold text-blue-900 mt-0.5">{{ formatPrice(tradingPlan.resistance) }}</p>
        </div>
        <div class="p-2 bg-orange-50 rounded border border-orange-200">
          <div class="flex items-center gap-1">
            <PhTrendDown :size="14" class="text-orange-600" weight="bold" />
            <span class="text-xs font-medium text-orange-700">Support</span>
          </div>
          <p class="text-base font-bold text-orange-900 mt-0.5">{{ formatPrice(tradingPlan.support) }}</p>
        </div>
      </div>

      <!-- Entries - Dynamic Grid Layout -->
      <div v-if="tradingPlan.entries && tradingPlan.entries.length > 0">
        <h3 class="text-xs font-bold text-gray-700 mb-2 flex items-center gap-1.5">
          <PhCurrencyDollar :size="14" weight="bold" />
          Entry Points ({{ tradingPlan.entries.length }})
        </h3>
        <div class="grid gap-2" :class="{
          'grid-cols-1': tradingPlan.entries.length === 1,
          'grid-cols-2': tradingPlan.entries.length === 2,
          'grid-cols-3': tradingPlan.entries.length === 3
        }">
          <div
            v-for="entry in tradingPlan.entries"
            :key="entry.entry_number"
            class="p-2.5 bg-gray-50 rounded border border-gray-200"
          >
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2 min-w-0 flex-1">
                <div class="w-6 h-6 bg-blue-100 rounded-full flex items-center justify-center flex-shrink-0">
                  <span class="text-xs font-bold text-blue-700">{{ entry.entry_number }}</span>
                </div>
                <div class="min-w-0">
                  <p class="text-sm font-bold text-gray-900 truncate">{{ formatPrice(entry.entry_price) }}</p>
                  <p class="text-xs text-gray-500">{{ entry.position_size }}</p>
                </div>
              </div>
              <div class="text-right flex-shrink-0 ml-2">
                <p class="text-xs font-semibold text-gray-900">{{ entry.position_qty.toFixed(4) }}</p>
                <p class="text-xs text-gray-500">{{ formatUSD(entry.position_value) }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Summary - Compact Grid -->
      <div v-if="tradingPlan.summary" class="pt-3 border-t border-gray-100">
        <h3 class="text-xs font-bold text-gray-700 mb-2">Summary</h3>
        <div class="grid grid-cols-3 gap-2">
          <div class="p-2 bg-gray-50 rounded">
            <p class="text-xs text-gray-500 mb-0.5">Capital</p>
            <p class="text-sm font-bold text-gray-900">{{ formatUSD(tradingPlan.summary.capital_used) }}</p>
          </div>
          <div class="p-2 bg-gray-50 rounded">
            <p class="text-xs text-gray-500 mb-0.5">Position</p>
            <p class="text-sm font-bold text-gray-900">{{ formatUSD(tradingPlan.summary.total_position_value) }}</p>
          </div>
          <div class="p-2 bg-gray-50 rounded">
            <p class="text-xs text-gray-500 mb-0.5">Risk</p>
            <p class="text-sm font-bold text-red-600">{{ formatPercent(tradingPlan.summary.risk_from_capital) }}</p>
          </div>
          <div class="p-2 bg-gray-50 rounded">
            <p class="text-xs text-gray-500 mb-0.5">Profit</p>
            <p class="text-sm font-bold text-green-600">{{ formatPercent(tradingPlan.summary.profit_from_capital) }}</p>
          </div>
          <div class="p-2 bg-gray-50 rounded">
            <p class="text-xs text-gray-500 mb-0.5">Leverage</p>
            <p class="text-sm font-bold text-gray-900">{{ tradingPlan.summary.effective_leverage.toFixed(1) }}x</p>
          </div>
          <div class="p-2 bg-gray-50 rounded">
            <p class="text-xs text-gray-500 mb-0.5">Avg Entry</p>
            <p class="text-sm font-bold text-gray-900">{{ formatPrice(tradingPlan.summary.avg_entry_price) }}</p>
          </div>
        </div>
      </div>

      <!-- Risk Metrics -->
      <div v-if="tradingPlan.summary" class="pt-3 border-t border-gray-100">
        <h3 class="text-xs font-bold text-gray-700 mb-2">Risk Metrics</h3>
        
        <!-- Risk vs Profit Bar -->
        <div class="mb-3">
          <div class="flex items-center justify-between text-xs mb-1">
            <span class="text-gray-600">Risk/Reward Balance</span>
            <span class="font-semibold text-gray-900">
              {{ formatPercent(tradingPlan.summary?.risk_from_capital || 0) }} / {{ formatPercent(tradingPlan.summary?.profit_from_capital || 0) }}
            </span>
          </div>
          <div class="relative h-3 bg-gray-200 rounded-full overflow-hidden">
            <div
              class="absolute top-0 left-0 h-full bg-gradient-to-r from-red-500 via-yellow-500 to-green-500 rounded-full"
              :style="{ width: '100%' }"
            ></div>
            <div
              class="absolute top-0 h-full w-0.5 bg-white"
              :style="{ left: `${Math.min(((tradingPlan.summary?.risk_from_capital || 0) / ((tradingPlan.summary?.risk_from_capital || 0) + (tradingPlan.summary?.profit_from_capital || 1))) * 100, 100)}%` }"
            ></div>
          </div>
          <div class="flex justify-between text-xs mt-1">
            <span class="text-red-600 font-medium">Risk</span>
            <span class="text-green-600 font-medium">Reward</span>
          </div>
        </div>

        <!-- Position Size Breakdown -->
        <div class="grid grid-cols-2 gap-2">
          <div class="p-2 bg-blue-50 rounded border border-blue-200">
            <p class="text-xs text-blue-700 font-medium mb-1">Position Qty</p>
            <p class="text-sm font-bold text-blue-900">{{ tradingPlan.summary?.total_position_qty.toFixed(4) }} coins</p>
          </div>
          <div class="p-2 bg-purple-50 rounded border border-purple-200">
            <p class="text-xs text-purple-700 font-medium mb-1">Entry Count</p>
            <p class="text-sm font-bold text-purple-900">{{ tradingPlan.summary?.total_entries }} position(s)</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="flex-1 flex items-center justify-center py-8">
      <div class="text-center">
        <PhChartLine :size="40" class="mx-auto text-gray-300 mb-2" />
        <p class="text-gray-500 text-sm">No trading plan yet</p>
        <p class="text-gray-400 text-xs mt-1">Run an analysis to see the plan</p>
      </div>
    </div>
  </div>
</template>
