<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useTradeStore } from '@/stores/trade.store'
import { useWatchlistStore } from '@/stores/watchlist.store'
import { useTimeframeStore } from '@/stores/timeframe.store'
import { DefaultLayout } from '@/layouts'
import { TradeFilter, TradeCard } from '@/components/features/trades'
import {
  PhCurrencyBtc,
  PhTrendUp,
  PhTarget,
  PhInfo,
  PhX,
  PhChartLineUp
} from '@phosphor-icons/vue'
import {
  getPnLColor
} from '@/helpers/trade'
import type { ITradeFilterState } from '@/components/features/trades/TradeFilter.vue'

const tradeStore = useTradeStore()
const watchlistStore = useWatchlistStore()
const timeframeStore = useTimeframeStore()

// Filter state
const filter = ref<ITradeFilterState>({
  status: [],
  symbol: [],
  interval: '',
  side: '',
  min_confidence: 0,
  date_start: '',
  date_end: ''
})

// Load data on mount
onMounted(() => {
  tradeStore.fetchTrades()
  watchlistStore.fetchWatchlists()
  timeframeStore.fetchTimeframes()
})

// Calculate stats
const totalTrades = computed(() => tradeStore.filteredTrades.length)

const totalPnL = computed(() => {
  return tradeStore.filteredTrades.reduce((sum, trade) => sum + (trade.pnl || 0), 0)
})

const winRate = computed(() => {
  if (totalTrades.value === 0) return 0
  const winningTrades = tradeStore.filteredTrades.filter(t => t.pnl > 0).length
  return (winningTrades / totalTrades.value) * 100
})

const winningTradesCount = computed(() => {
  return tradeStore.filteredTrades.filter(t => t.pnl > 0).length
})

const losingTradesCount = computed(() => {
  return tradeStore.filteredTrades.filter(t => t.pnl < 0).length
})

// Apply filters when filter state changes
watch(() => filter.value, () => {
  // Update trade store filter
  tradeStore.updateFilter(filter.value)
  tradeStore.applyFilters()
}, { deep: true })

// Remove single filter
const removeFilter = (type: 'status' | 'symbol' | 'interval' | 'side' | 'min_confidence', value: any) => {
  if (type === 'status' || type === 'symbol') {
    const current = tradeStore.filter[type]
    const updated = current.filter(v => v !== value)
    tradeStore.updateFilter({ [type]: updated })
  } else {
    tradeStore.updateFilter({ [type]: type === 'min_confidence' ? 0 : '' })
  }
  tradeStore.applyFilters()
}

// Reset filters handler
const handleResetFilters = () => {
  tradeStore.resetFilter()
  tradeStore.applyFilters()
}

// Apply filters handler
const handleApplyFilters = (newFilter: ITradeFilterState) => {
  tradeStore.updateFilter(newFilter)
  tradeStore.applyFilters()
}
</script>

<template>
  <DefaultLayout>
    <template #header-title>Trade History</template>

    <div class="md:mx-auto sm:px-6">
      <!-- Loading State -->
      <div v-if="tradeStore.loading" class="flex items-center justify-center py-20">
        <div class="relative">
          <div class="animate-spin rounded-full h-16 w-16 border-b-2 border-primary"></div>
          <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2">
            <PhCurrencyBtc :size="24" class="text-primary" />
          </div>
        </div>
      </div>

      <template v-else>
        <!-- Stats Overview -->
        <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <!-- Total Trades -->
          <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-4">
            <div class="flex items-center gap-2 mb-2">
              <div class="p-2 bg-blue-50 rounded-lg">
                <PhTrendUp :size="20" class="text-blue-600" weight="fill" />
              </div>
              <span class="text-xs text-gray-600">Total Trades</span>
            </div>
            <p class="text-2xl font-bold text-gray-900">{{ totalTrades }}</p>
          </div>

          <!-- Win Rate -->
          <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-4">
            <div class="flex items-center gap-2 mb-2">
              <div class="p-2" :class="winRate >= 50 ? 'bg-green-50' : 'bg-red-50'">
                <PhTarget :size="20" class="text-gray-600" :class="winRate >= 50 ? 'text-green-600' : 'text-red-600'" weight="fill" />
              </div>
              <span class="text-xs text-gray-600">Win Rate</span>
            </div>
            <p class="text-2xl font-bold" :class="winRate >= 50 ? 'text-green-600' : 'text-red-600'">
              {{ winRate.toFixed(1) }}%
            </p>
          </div>

          <!-- Total PnL -->
          <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-4">
            <div class="flex items-center gap-2 mb-2">
              <div class="p-2" :class="totalPnL >= 0 ? 'bg-green-50' : 'bg-red-50'">
                <PhChartLineUp :size="20" class="text-gray-600" :class="totalPnL >= 0 ? 'text-green-600' : 'text-red-600'" weight="fill" />
              </div>
              <span class="text-xs text-gray-600">Total PnL</span>
            </div>
            <p class="text-xl font-bold" :class="getPnLColor(totalPnL)">
              {{ totalPnL >= 0 ? '+' : '' }}{{ totalPnL.toFixed(2) }} USDT
            </p>
          </div>

          <!-- Win/Loss -->
          <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-4">
            <div class="flex items-center gap-2 mb-2">
              <div class="p-2 bg-purple-50 rounded-lg">
                <PhInfo :size="20" class="text-purple-600" weight="fill" />
              </div>
              <span class="text-xs text-gray-600">Win/Loss</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="text-lg font-bold text-green-600">{{ winningTradesCount }}</span>
              <span class="text-sm text-gray-400">/</span>
              <span class="text-lg font-bold text-red-600">{{ losingTradesCount }}</span>
            </div>
          </div>
        </div>

        <!-- Filter Bar -->
        <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-4 mb-6">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3 flex-wrap">
              <div class="flex items-center gap-2">
                <PhChartLineUp :size="20" class="text-gray-600" />
                <span class="text-sm font-semibold text-gray-700">Active Filters:</span>
              </div>

              <!-- Active Filter Chips -->
              <div class="flex items-center gap-2 flex-wrap">
                <span
                  v-for="status in tradeStore.filter.status"
                  :key="status"
                  class="px-2 py-1 bg-blue-100 text-blue-700 text-xs font-medium rounded-full flex items-center gap-1"
                >
                  {{ status }}
                  <button @click="removeFilter('status', status)" class="hover:text-blue-900">
                    <PhX :size="12" />
                  </button>
                </span>

                <span
                  v-for="symbol in tradeStore.filter.symbol"
                  :key="symbol"
                  class="px-2 py-1 bg-purple-100 text-purple-700 text-xs font-medium rounded-full flex items-center gap-1"
                >
                  {{ symbol }}
                  <button @click="removeFilter('symbol', symbol)" class="hover:text-purple-900">
                    <PhX :size="12" />
                  </button>
                </span>

                <span
                  v-if="tradeStore.filter.interval"
                  class="px-2 py-1 bg-green-100 text-green-700 text-xs font-medium rounded-full flex items-center gap-1"
                >
                  {{ tradeStore.filter.interval }}
                  <button @click="tradeStore.updateFilter({ interval: '' })" class="hover:text-green-900">
                    <PhX :size="12" />
                  </button>
                </span>

                <span
                  v-if="tradeStore.filter.side"
                  class="px-2 py-1 bg-orange-100 text-orange-700 text-xs font-medium rounded-full flex items-center gap-1"
                >
                  {{ tradeStore.filter.side }}
                  <button @click="tradeStore.updateFilter({ side: '' })" class="hover:text-orange-900">
                    <PhX :size="12" />
                  </button>
                </span>

                <span
                  v-if="tradeStore.filter.min_confidence > 0"
                  class="px-2 py-1 bg-yellow-100 text-yellow-700 text-xs font-medium rounded-full flex items-center gap-1"
                >
                  ≥{{ tradeStore.filter.min_confidence }}% conf
                  <button @click="tradeStore.updateFilter({ min_confidence: 0 })" class="hover:text-yellow-900">
                    <PhX :size="12" />
                  </button>
                </span>
              </div>
            </div>

            <!-- Filter Component -->
            <TradeFilter
              v-model="filter"
              :loading="tradeStore.loading"
              @apply="handleApplyFilters"
              @reset="handleResetFilters"
            />
          </div>
        </div>

        <!-- Trades List -->
        <div v-if="tradeStore.filteredTrades.length === 0" class="text-center py-12 bg-white rounded-2xl shadow-lg border border-gray-100">
          <PhCurrencyBtc :size="64" class="mx-auto text-gray-300 mb-4" />
          <p class="text-gray-500 text-lg font-medium">No trades found</p>
          <p class="text-gray-400 text-sm mt-1">
            {{ tradeStore.trades.length === 0 ? 'Trades will appear here when they are executed' : 'Try adjusting your filters' }}
          </p>
        </div>

        <div v-else class="space-y-4 grid grid-cols-2 gap-2">
          <!-- Trade Cards -->
          <TradeCard
            v-for="trade in tradeStore.filteredTrades"
            :key="trade.id"
            :trade="trade"
          />
        </div>
      </template>
    </div>
  </DefaultLayout>
</template>

<style scoped>
</style>
