<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  PhCheckCircle,
  PhClock,
  PhXCircle,
  PhGear,
  PhTrendUp,
  PhTrendDown,
  PhCurrencyDollar,
  PhChartBar
} from '@phosphor-icons/vue'
import type { IBacktestDetail, IBacktestTrade } from '@/stores/backtest.store'
import BacktestTradeDetailModal from './BacktestTradeDetailModal.vue'
import { formatDuration } from '@/helpers/formatters'

interface Props {
  backtest: IBacktestDetail | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'view-trade': [trade: IBacktestTrade]
}>()

const showTradeDetail = ref(false)
const selectedTrade = ref<IBacktestTrade | null>(null)

const tradeStats = computed(() => {
  const trades = props.backtest?.trades || []

  return {
    total: trades.length,
    closed: trades.filter(t => t.status === 'CLOSED').length,
    active: trades.filter(t => t.status === 'ACTIVE').length,
    cancelled: trades.filter(t => t.status === 'CANCELLED').length,
    expired: trades.filter(t => t.status === 'EXPIRED').length
  }
})

const getTradeStatusColor = (status: string) => {
  switch (status) {
    case 'CLOSED':
      return 'bg-green-100 text-green-700'
    case 'ACTIVE':
      return 'bg-blue-100 text-blue-700'
    case 'CANCELLED':
      return 'bg-gray-100 text-gray-700'
    case 'EXPIRED':
      return 'bg-yellow-100 text-yellow-700'
    default:
      return 'bg-gray-100 text-gray-700'
  }
}

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2
  }).format(value)
}

const formatPercent = (value: number) => {
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`
}

const viewTradeDetail = (trade: IBacktestTrade) => {
  selectedTrade.value = trade
  showTradeDetail.value = true
  emit('view-trade', trade)
}

const closeTradeDetail = () => {
  showTradeDetail.value = false
  selectedTrade.value = null
}
</script>

<template>
  <div class="space-y-4">
    <!-- Trade Statistics -->
    <div class="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
      <div class="bg-gradient-to-br from-gray-50 to-gray-100 border border-gray-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-gray-500 rounded-lg flex items-center justify-center">
            <PhChartBar :size="16" class="text-white" weight="fill" />
          </div>
          <span class="text-xs font-medium text-gray-600">Total</span>
        </div>
        <p class="text-2xl font-bold text-gray-900">{{ tradeStats.total }}</p>
      </div>
      <div class="bg-gradient-to-br from-green-50 to-emerald-100 border border-green-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-green-500 rounded-lg flex items-center justify-center">
            <PhCheckCircle :size="16" class="text-white" weight="fill" />
          </div>
          <span class="text-xs font-medium text-green-600">Closed</span>
        </div>
        <p class="text-2xl font-bold text-green-800">{{ tradeStats.closed }}</p>
      </div>
      <div class="bg-gradient-to-br from-blue-50 to-indigo-100 border border-blue-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-blue-500 rounded-lg flex items-center justify-center">
            <PhGear :size="16" class="text-white" weight="fill" />
          </div>
          <span class="text-xs font-medium text-blue-600">Active</span>
        </div>
        <p class="text-2xl font-bold text-blue-800">{{ tradeStats.active }}</p>
      </div>
      <div class="bg-gradient-to-br from-gray-50 to-gray-200 border border-gray-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-gray-400 rounded-lg flex items-center justify-center">
            <PhXCircle :size="16" class="text-white" weight="fill" />
          </div>
          <span class="text-xs font-medium text-gray-600">Cancelled</span>
        </div>
        <p class="text-2xl font-bold text-gray-800">{{ tradeStats.cancelled }}</p>
      </div>
      <div class="bg-gradient-to-br from-yellow-50 to-amber-100 border border-yellow-200 rounded-xl p-4">
        <div class="flex items-center gap-2 mb-2">
          <div class="w-8 h-8 bg-yellow-500 rounded-lg flex items-center justify-center">
            <PhClock :size="16" class="text-white" weight="fill" />
          </div>
          <span class="text-xs font-medium text-yellow-600">Expired</span>
        </div>
        <p class="text-2xl font-bold text-yellow-800">{{ tradeStats.expired }}</p>
      </div>
    </div>

    <!-- Trade List -->
    <div v-if="backtest?.trades && backtest.trades.length > 0" class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div
        v-for="trade in backtest.trades"
        :key="trade.trade_id"
        class="bg-white border border-gray-200 rounded-xl p-4 hover:shadow-lg hover:border-primary/30 transition-all cursor-pointer group"
        @click="viewTradeDetail(trade)"
      >
        <!-- Header -->
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-2">
            <div
              class="w-10 h-10 rounded-xl flex items-center justify-center"
              :class="trade.side === 'BUY' ? 'bg-green-500' : 'bg-red-500'"
            >
              <component
                :is="trade.side === 'BUY' ? PhTrendUp : PhTrendDown"
                :size="20"
                class="text-white"
                weight="fill"
              />
            </div>
            <div>
              <p class="text-sm font-bold text-gray-900">Trade #{{ trade.trade_num }}</p>
              <p class="text-xs text-gray-500">{{ trade.side }}</p>
            </div>
          </div>
          <div class="flex items-center gap-1">
            <!-- <component
              :is="getTradeStatusIcon(trade.status)"
              :size="16"
              :weight="trade.status === 'CLOSED' ? 'fill' : 'regular'"
              :class="getTradeStatusColor(trade.status).split(' ')[1]"
            /> -->
            <span
              class="px-2 py-1 text-xs font-medium rounded-full"
              :class="getTradeStatusColor(trade.status)"
            >
              {{ trade.status }}
            </span>
          </div>
        </div>

        <!-- PnL Display -->
        <div
          class="rounded-xl p-3 mb-3 text-center"
          :class="trade.pnl >= 0 ? 'bg-gradient-to-r from-green-50 to-emerald-50' : 'bg-gradient-to-r from-red-50 to-rose-50'"
        >
          <div
            class="text-xl font-bold"
            :class="trade.pnl >= 0 ? 'text-green-700' : 'text-red-700'"
          >
            {{ trade.pnl >= 0 ? '+' : '' }}{{ formatCurrency(trade.pnl) }}
          </div>
          <div
            class="text-xs font-medium mt-1"
            :class="trade.pnl_percent >= 0 ? 'text-green-600' : 'text-red-600'"
          >
            {{ formatPercent(trade.pnl_percent) }}
          </div>
        </div>

        <!-- Details Grid -->
        <div class="grid grid-cols-2 gap-2">
          <div class="bg-gray-50 rounded-lg p-2">
            <div class="flex items-center gap-1 mb-1">
              <PhCurrencyDollar :size="12" class="text-gray-400" />
              <span class="text-xs text-gray-500">Capital</span>
            </div>
            <p class="text-sm font-bold text-gray-900">{{ formatCurrency(trade.total_capital) }}</p>
          </div>
          <div class="bg-gray-50 rounded-lg p-2">
            <div class="flex items-center gap-1 mb-1">
              <PhClock :size="12" class="text-gray-400" />
              <span class="text-xs text-gray-500">Duration</span>
            </div>
            <p class="text-sm font-bold text-gray-900">{{ formatDuration(trade.duration_minutes * 60) }}</p>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="text-center py-20">
      <PhClock :size="48" class="mx-auto text-gray-300 mb-4" />
      <p class="text-gray-500">No trades yet</p>
    </div>

    <!-- Trade Detail Modal -->
    <BacktestTradeDetailModal
      v-model:show="showTradeDetail"
      :trade="selectedTrade"
      @close="closeTradeDetail"
    />
  </div>
</template>
