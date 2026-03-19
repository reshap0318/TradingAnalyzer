<script setup lang="ts">
import { computed } from 'vue'
import {
  PhChartLineUp,
  PhChartBar,
  PhTrendUp,
  PhTrendDown,
  PhCurrencyDollar,
  PhWallet,
  PhPiggyBank,
  PhChartPie,
  PhClock,
  PhCalendar
} from '@phosphor-icons/vue'
import type { IBacktestDetail } from '@/stores/backtest.store'
import { formatCurrency, formatDate } from '@/helpers/backtest'
import BacktestPriceChart from './BacktestPriceChart.vue'
import BacktestEquityChart from './BacktestEquityChart.vue'

interface Props {
  backtest: IBacktestDetail | null
}

const props = defineProps<Props>()

// Chart statistics
const chartStats = computed(() => {
  const equityCurve = props.backtest?.equity_curve || []
  const ohlcv = props.backtest?.ohlcv || []
  const trades = props.backtest?.trades || []

  // Calculate price stats
  const prices = ohlcv.map(c => c.close)
  const minPrice = prices.length > 0 ? Math.min(...prices) : 0
  const maxPrice = prices.length > 0 ? Math.max(...prices) : 0
  const avgPrice = prices.length > 0 ? prices.reduce((a, b) => a + b, 0) / prices.length : 0

  // Calculate equity stats
  const balances = equityCurve.map(p => p.balance)
  const initialBalance = balances[0] || 0
  const finalBalance = balances[balances.length - 1] || 0
  const peakBalance = balances.length > 0 ? Math.max(...balances) : 0
  const lowestBalance = balances.length > 0 ? Math.min(...balances) : 0

  return {
    ohlcv: {
      count: ohlcv.length,
      minPrice,
      maxPrice,
      avgPrice,
      firstCandle: ohlcv[0],
      lastCandle: ohlcv[ohlcv.length - 1]
    },
    equity: {
      initialBalance,
      finalBalance,
      peakBalance,
      lowestBalance,
      totalGrowth: finalBalance - initialBalance,
      growthPercent: initialBalance > 0 ? ((finalBalance - initialBalance) / initialBalance * 100) : 0
    },
    trades: {
      long: trades.filter(t => t.side === 'BUY').length,
      short: trades.filter(t => t.side === 'SELL').length,
      totalPnl: trades.reduce((sum, t) => sum + t.pnl, 0)
    }
  }
})
</script>

<template>
  <div class="space-y-4">
    <!-- Equity Curve Chart -->
    <div class="bg-white border border-gray-200 rounded-xl p-6">
      <BacktestEquityChart :equityCurve="props.backtest?.equity_curve || []" />
    </div>

    <!-- Balance Statistics -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="bg-white border border-gray-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-purple-100 rounded-lg flex items-center justify-center">
            <PhWallet :size="16" class="text-purple-600" weight="fill" />
          </div>
          <span class="text-xs font-medium text-gray-600">Initial</span>
        </div>
        <p class="text-lg font-bold text-gray-900">{{ formatCurrency(chartStats.equity.initialBalance) }}</p>
      </div>
      <div class="bg-white border border-gray-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-green-100 rounded-lg flex items-center justify-center">
            <PhPiggyBank :size="16" class="text-green-600" weight="fill" />
          </div>
          <span class="text-xs font-medium text-gray-600">Final</span>
        </div>
        <p class="text-lg font-bold text-gray-900">{{ formatCurrency(chartStats.equity.finalBalance) }}</p>
      </div>
      <div class="bg-white border border-gray-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
            <PhTrendUp :size="16" class="text-blue-600" weight="fill" />
          </div>
          <span class="text-xs font-medium text-gray-600">Peak</span>
        </div>
        <p class="text-lg font-bold text-gray-900">{{ formatCurrency(chartStats.equity.peakBalance) }}</p>
      </div>
      <div class="bg-white border border-gray-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-red-100 rounded-lg flex items-center justify-center">
            <PhTrendDown :size="16" class="text-red-600" weight="fill" />
          </div>
          <span class="text-xs font-medium text-gray-600">Lowest</span>
        </div>
        <p class="text-lg font-bold text-gray-900">{{ formatCurrency(chartStats.equity.lowestBalance) }}</p>
      </div>
    </div>

    <!-- Growth Stats -->
    <div class="bg-white border border-gray-200 rounded-xl p-6">
      <h3 class="text-base font-bold text-gray-900 mb-4 flex items-center gap-2">
        <PhChartBar :size="20" class="text-gray-500" weight="fill" />
        Growth Statistics
      </h3>
      <div class="grid grid-cols-2 gap-4">
        <div class="bg-gray-50 rounded-lg p-4">
          <div class="flex items-center gap-2 mb-2">
            <PhCurrencyDollar :size="16" class="text-gray-400" />
            <span class="text-xs text-gray-500">Total Growth</span>
          </div>
          <p class="text-xl font-bold" :class="chartStats.equity.totalGrowth >= 0 ? 'text-green-600' : 'text-red-600'">
            {{ chartStats.equity.totalGrowth >= 0 ? '+' : '' }}{{ formatCurrency(chartStats.equity.totalGrowth) }}
          </p>
        </div>
        <div class="bg-gray-50 rounded-lg p-4">
          <div class="flex items-center gap-2 mb-2">
            <PhChartLineUp :size="16" class="text-gray-400" />
            <span class="text-xs text-gray-500">Growth %</span>
          </div>
          <p class="text-xl font-bold" :class="chartStats.equity.growthPercent >= 0 ? 'text-green-600' : 'text-red-600'">
            {{ chartStats.equity.growthPercent >= 0 ? '+' : '' }}{{ chartStats.equity.growthPercent.toFixed(2) }}%
          </p>
        </div>
      </div>
    </div>

    <!-- OHLCV Data -->
    <div class="bg-white border border-gray-200 rounded-xl p-6">
      <h3 class="text-base font-bold text-gray-900 mb-4 flex items-center gap-2">
        <PhChartPie :size="20" class="text-gray-500" weight="fill" />
        Price Chart (OHLCV)
      </h3>

      <!-- Price Chart with Lightweight Charts -->
      <BacktestPriceChart
        :ohlcv="props.backtest?.ohlcv || []"
        :trades="props.backtest?.trades || []"
      />

      <!-- OHLCV Stats Grid -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="bg-gray-50 rounded-lg p-3">
          <div class="flex items-center gap-2 mb-2">
            <PhChartBar :size="14" class="text-gray-400" />
            <span class="text-xs text-gray-500">Candles</span>
          </div>
          <p class="text-lg font-bold text-gray-900">{{ chartStats.ohlcv.count }}</p>
        </div>
        <div class="bg-gray-50 rounded-lg p-3">
          <div class="flex items-center gap-2 mb-2">
            <PhTrendDown :size="14" class="text-gray-400" />
            <span class="text-xs text-gray-500">Min</span>
          </div>
          <p class="text-lg font-bold text-gray-900">{{ formatCurrency(chartStats.ohlcv.minPrice) }}</p>
        </div>
        <div class="bg-gray-50 rounded-lg p-3">
          <div class="flex items-center gap-2 mb-2">
            <PhTrendUp :size="14" class="text-gray-400" />
            <span class="text-xs text-gray-500">Max</span>
          </div>
          <p class="text-lg font-bold text-gray-900">{{ formatCurrency(chartStats.ohlcv.maxPrice) }}</p>
        </div>
        <div class="bg-gray-50 rounded-lg p-3">
          <div class="flex items-center gap-2 mb-2">
            <PhChartLineUp :size="14" class="text-gray-400" />
            <span class="text-xs text-gray-500">Avg</span>
          </div>
          <p class="text-lg font-bold text-gray-900">{{ formatCurrency(chartStats.ohlcv.avgPrice) }}</p>
        </div>
      </div>
      <div class="mt-4 pt-4 border-t border-gray-200">
        <div class="grid grid-cols-2 gap-4">
          <div class="bg-gray-50 rounded-lg p-3">
            <div class="flex items-center gap-2 mb-1">
              <PhCalendar :size="14" class="text-gray-400" />
              <span class="text-xs text-gray-500">First Candle</span>
            </div>
            <p class="text-xs font-medium text-gray-900">
              {{ chartStats.ohlcv.firstCandle ? formatDate(chartStats.ohlcv.firstCandle.timestamp) : '-' }}
            </p>
          </div>
          <div class="bg-gray-50 rounded-lg p-3">
            <div class="flex items-center gap-2 mb-1">
              <PhClock :size="14" class="text-gray-400" />
              <span class="text-xs text-gray-500">Last Candle</span>
            </div>
            <p class="text-xs font-medium text-gray-900">
              {{ chartStats.ohlcv.lastCandle ? formatDate(chartStats.ohlcv.lastCandle.timestamp) : '-' }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Trade Summary -->
    <div class="bg-white border border-gray-200 rounded-xl p-6">
      <h3 class="text-base font-bold text-gray-900 mb-4 flex items-center gap-2">
        <PhChartLineUp :size="20" class="text-gray-500" weight="fill" />
        Trade Summary
      </h3>
      <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
        <div class="bg-gray-50 rounded-lg p-4">
          <div class="flex items-center gap-2 mb-2">
            <div class="w-8 h-8 bg-green-100 rounded-lg flex items-center justify-center">
              <PhTrendUp :size="16" class="text-green-600" weight="fill" />
            </div>
            <span class="text-xs text-gray-500">Long</span>
          </div>
          <p class="text-2xl font-bold text-gray-900">{{ chartStats.trades.long }}</p>
        </div>
        <div class="bg-gray-50 rounded-lg p-4">
          <div class="flex items-center gap-2 mb-2">
            <div class="w-8 h-8 bg-red-100 rounded-lg flex items-center justify-center">
              <PhTrendDown :size="16" class="text-red-600" weight="fill" />
            </div>
            <span class="text-xs text-gray-500">Short</span>
          </div>
          <p class="text-2xl font-bold text-gray-900">{{ chartStats.trades.short }}</p>
        </div>
        <div class="bg-gray-50 rounded-lg p-4">
          <div class="flex items-center gap-2 mb-2">
            <div class="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
              <PhCurrencyDollar :size="16" class="text-blue-600" weight="fill" />
            </div>
            <span class="text-xs text-gray-500">Total PnL</span>
          </div>
          <p class="text-xl font-bold" :class="chartStats.trades.totalPnl >= 0 ? 'text-green-600' : 'text-red-600'">
            {{ formatCurrency(chartStats.trades.totalPnl) }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
