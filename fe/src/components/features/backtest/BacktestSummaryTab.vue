<script setup lang="ts">
import { computed } from 'vue'
import {
  PhCurrencyDollar,
  PhChartLineUp,
  PhTrendUp,
  PhTrendDown,
  PhTarget,
  PhMedal
} from '@phosphor-icons/vue'
import type { IBacktestDetail } from '@/stores/backtest.store'
import { formatCurrency, formatPercent, formatDate } from '@/helpers/backtest'

interface Props {
  backtest: IBacktestDetail | null
}

const props = defineProps<Props>()

const summary = computed(() => props.backtest?.summary)
</script>

<template>
  <div class="space-y-4">
    <!-- Key Metrics -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <div class="bg-gradient-to-br from-green-50 to-emerald-100 border border-green-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-green-500 rounded-lg flex items-center justify-center">
            <PhCurrencyDollar :size="16" class="text-white" weight="fill" />
          </div>
          <span class="text-xs font-medium text-green-700">Net Profit</span>
        </div>
        <div class="text-xl font-bold text-green-800">
          {{ formatCurrency(summary?.net_profit || 0) }}
        </div>
        <div class="text-xs font-medium mt-1 text-green-700">
          {{ formatPercent(summary?.net_profit_percent || 0) }}
        </div>
      </div>

      <div class="bg-gradient-to-br from-blue-50 to-indigo-100 border border-blue-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-blue-500 rounded-lg flex items-center justify-center">
            <PhChartLineUp :size="16" class="text-white" weight="fill" />
          </div>
          <span class="text-xs font-medium text-blue-700">Win Rate</span>
        </div>
        <div class="text-xl font-bold text-blue-800">
          {{ summary?.win_rate_pct?.toFixed(1) || '0' }}%
        </div>
        <div class="text-xs text-blue-600 mt-1">
          {{ summary?.winning_trades || 0 }}W / {{ summary?.losing_trades || 0 }}L
        </div>
      </div>

      <div class="bg-gradient-to-br from-purple-50 to-pink-100 border border-purple-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-purple-500 rounded-lg flex items-center justify-center">
            <PhTrendUp :size="16" class="text-white" weight="fill" />
          </div>
          <span class="text-xs font-medium text-purple-700">Profit Factor</span>
        </div>
        <div class="text-xl font-bold text-purple-800">
          {{ summary?.profit_factor?.toFixed(2) || '0' }}
        </div>
        <div class="text-xs text-purple-600 mt-1">
          Gross Profit / Loss
        </div>
      </div>

      <div class="bg-gradient-to-br from-red-50 to-rose-100 border border-red-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-red-500 rounded-lg flex items-center justify-center">
            <PhTrendDown :size="16" class="text-white" weight="fill" />
          </div>
          <span class="text-xs font-medium text-red-700">Max Drawdown</span>
        </div>
        <div class="text-xl font-bold text-red-800">
          {{ summary?.max_drawdown_pct?.toFixed(2) || '0' }}%
        </div>
        <div class="text-xs text-red-600 mt-1">
          Peak to Valley
        </div>
      </div>
    </div>

    <!-- Additional Stats -->
    <div class="bg-white border border-gray-200 rounded-xl p-4">
      <h3 class="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
        <PhTarget :size="16" class="text-blue-500" weight="fill" />
        Additional Statistics
      </h3>
      <div class="grid grid-cols-3 md:grid-cols-6 gap-3">
        <div class="text-center p-3 bg-gray-50 rounded-lg">
          <p class="text-xs text-gray-500 mb-1">Total</p>
          <p class="text-lg font-bold text-gray-900">{{ summary?.total_trades || 0 }}</p>
        </div>
        <div class="text-center p-3 bg-green-50 rounded-lg">
          <p class="text-xs text-green-600 mb-1">Winning</p>
          <p class="text-lg font-bold text-green-700">{{ summary?.winning_trades || 0 }}</p>
        </div>
        <div class="text-center p-3 bg-red-50 rounded-lg">
          <p class="text-xs text-red-600 mb-1">Losing</p>
          <p class="text-lg font-bold text-red-700">{{ summary?.losing_trades || 0 }}</p>
        </div>
        <div class="text-center p-3 bg-yellow-50 rounded-lg">
          <p class="text-xs text-yellow-600 mb-1">Expired</p>
          <p class="text-lg font-bold text-yellow-700">{{ summary?.expired_trades || 0 }}</p>
        </div>
        <div class="text-center p-3 bg-gray-100 rounded-lg">
          <p class="text-xs text-gray-600 mb-1">Cancelled</p>
          <p class="text-lg font-bold text-gray-700">{{ summary?.cancelled_trades || 0 }}</p>
        </div>
        <div class="text-center p-3 bg-indigo-50 rounded-lg">
          <p class="text-xs text-indigo-600 mb-1">Initial</p>
          <p class="text-lg font-bold text-indigo-700">{{ formatCurrency(summary?.initial_balance || 0) }}</p>
        </div>
      </div>
    </div>

    <!-- Backtest Info -->
    <div class="bg-white border border-gray-200 rounded-xl p-4">
      <h3 class="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
        <PhMedal :size="16" class="text-purple-500" weight="fill" />
        Backtest Information
      </h3>
      <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
        <div class="space-y-1">
          <p class="text-xs text-gray-500">Symbol</p>
          <p class="text-sm font-semibold text-gray-900 bg-gray-50 rounded px-2 py-1.5">{{ backtest?.symbol }}</p>
        </div>
        <div class="space-y-1">
          <p class="text-xs text-gray-500">Strategy</p>
          <p class="text-sm font-semibold text-gray-900 bg-gray-50 rounded px-2 py-1.5">{{ backtest?.strategy?.strategy_name || '-' }}</p>
        </div>
        <div class="space-y-1">
          <p class="text-xs text-gray-500">Capital</p>
          <p class="text-sm font-semibold text-green-700 bg-green-50 rounded px-2 py-1.5">{{ formatCurrency(backtest?.capital || 0) }}</p>
        </div>
        <div class="space-y-1">
          <p class="text-xs text-gray-500">Period</p>
          <p class="text-xs font-medium text-gray-900 bg-gray-50 rounded px-2 py-1.5">{{ formatDate(backtest?.start_time || '') }} - {{ formatDate(backtest?.end_time || '') }}</p>
        </div>
        <div class="space-y-1">
          <p class="text-xs text-gray-500">Created</p>
          <p class="text-xs font-medium text-gray-900 bg-gray-50 rounded px-2 py-1.5">{{ formatDate(backtest?.created_at || '') }}</p>
        </div>
        <div v-if="backtest?.completed_at" class="space-y-1">
          <p class="text-xs text-gray-500">Completed</p>
          <p class="text-xs font-medium text-green-700 bg-green-50 rounded px-2 py-1.5">{{ formatDate(backtest.completed_at) }}</p>
        </div>
      </div>
    </div>
  </div>
</template>
