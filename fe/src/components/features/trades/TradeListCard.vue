<script setup lang="ts">
import { ref, computed } from 'vue'
import { UiModal } from '@/components/common'
import type { ITrade } from '@/stores/tradebot.store'
import {
  getSideBgColor,
  getStatusColor,
  getTpSlStatusBadge,
  getTpSlStatusColor,
  getPnLColor
} from '@/helpers/trade'
import { formatPrice, formatPnL, formatDate } from '@/helpers/formatters'
import {
  PhTrendUp,
  PhTarget,
  PhStopCircle,
  PhCurrencyBtc,
  PhInfo
} from '@phosphor-icons/vue'

interface ITradeListCardProps {
  trades: ITrade[]
  loading?: boolean
}

const props = withDefaults(defineProps<ITradeListCardProps>(), {
  trades: () => [],
  loading: false
})

// Modal state
const showModal = ref(false)
const selectedTrade = ref<ITrade | null>(null)

// Calculate total PnL from all trades
const totalPnL = computed(() => {
  return props.trades.reduce((sum, trade) => sum + (trade.pnl || 0), 0)
})

// Calculate win rate
const winRate = computed(() => {
  if (props.trades.length === 0) return 0
  const winningTrades = props.trades.filter(t => t.pnl > 0).length
  return (winningTrades / props.trades.length) * 100
})

const winningTradesCount = computed(() => {
  return props.trades.filter(t => t.pnl > 0).length
})

const losingTradesCount = computed(() => {
  return props.trades.filter(t => t.pnl < 0).length
})

const handleClose = () => {
  showModal.value = false
  selectedTrade.value = null
}

// Open trade detail
const openTradeDetail = (trade: ITrade) => {
  selectedTrade.value = trade
  showModal.value = true
}
</script>

<template>
  <!-- Card - Clickable to open modal -->
  <div
    class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6 cursor-pointer hover:shadow-xl hover:border-blue-300 transition-all duration-200"
    @click="showModal = true"
  >
    <div class="flex items-center gap-3 mb-4">
      <div class="p-3 bg-blue-50 rounded-xl">
        <PhTrendUp :size="24" class="text-blue-600" weight="fill" />
      </div>
      <span class="text-sm text-gray-600">All Trades</span>
    </div>
    <div class="text-3xl font-bold text-gray-900">{{ trades.length }}</div>
    <p v-if="trades.length > 0" class="text-xs text-gray-500 mt-1">
      Total trades
    </p>
  </div>

  <!-- Modal - Full trade details -->
  <UiModal
    :model-value="showModal"
    title="Trade History"
    size="full"
    @update:model-value="handleClose"
  >
    <!-- Loading State -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <div class="relative">
        <div class="animate-spin rounded-full h-16 w-16 border-b-2 border-primary"></div>
        <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2">
          <PhCurrencyBtc :size="24" class="text-primary" />
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="trades.length === 0" class="text-center py-12">
      <PhCurrencyBtc :size="64" class="mx-auto text-gray-300 mb-4" />
      <p class="text-gray-500 text-lg font-medium">No trades yet</p>
      <p class="text-gray-400 text-sm mt-1">Trades will appear here when the bot executes them</p>
    </div>

    <!-- Trades List -->
    <div v-else class="space-y-6">
      <!-- Summary Stats -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 p-4 bg-gray-50 rounded-xl">
        <div class="text-center">
          <p class="text-xs text-gray-500 mb-1">Total Trades</p>
          <p class="text-2xl font-bold text-gray-900">{{ trades.length }}</p>
        </div>
        <div class="text-center">
          <p class="text-xs text-gray-500 mb-1">Win Rate</p>
          <p class="text-2xl font-bold" :class="winRate >= 50 ? 'text-green-600' : 'text-red-600'">
            {{ winRate.toFixed(1) }}%
          </p>
        </div>
        <div class="text-center">
          <p class="text-xs text-gray-500 mb-1">Total PnL</p>
          <p class="text-xl font-bold" :class="getPnLColor(totalPnL)">
            {{ totalPnL >= 0 ? '+' : '' }}{{ totalPnL.toFixed(2) }} USDT
          </p>
        </div>
        <div class="text-center">
          <p class="text-xs text-gray-500 mb-1">Win/Loss</p>
          <div class="flex items-center justify-center gap-1">
            <span class="text-lg font-bold text-green-600">{{ winningTradesCount }}</span>
            <span class="text-sm text-gray-400">/</span>
            <span class="text-lg font-bold text-red-600">{{ losingTradesCount }}</span>
          </div>
        </div>
      </div>

      <!-- Trade Items -->
      <div class="space-y-4">
        <div
          v-for="trade in trades"
          :key="trade.id"
          class="border border-gray-200 rounded-xl p-4 hover:shadow-md transition-all cursor-pointer"
          @click="openTradeDetail(trade)"
        >
          <!-- Mobile View -->
          <div class="md:hidden space-y-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <PhCurrencyBtc :size="20" class="text-blue-600" weight="fill" />
                <span class="font-bold text-gray-900">{{ trade.symbol }}</span>
                <span
                  class="px-2 py-0.5 text-xs font-bold rounded-full"
                  :class="getSideBgColor(trade.side)"
                >
                  {{ trade.side.toUpperCase() }}
                </span>
              </div>
              <span
                class="px-2 py-1 text-xs font-semibold rounded"
                :class="getStatusColor(trade.status)"
              >
                {{ trade.status }}
              </span>
            </div>

            <div class="grid grid-cols-2 gap-2 text-sm">
              <div>
                <p class="text-xs text-gray-500">Entry</p>
                <p class="font-semibold">{{ formatPrice(trade.avg_entry_price) }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500">Exit</p>
                <p class="font-semibold">{{ formatPrice(trade.exit_price || 0) }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500">TP</p>
                <p class="font-semibold text-green-600">{{ formatPrice(trade.tp_price) }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500">SL</p>
                <p class="font-semibold text-red-600">{{ formatPrice(trade.sl_price) }}</p>
              </div>
            </div>

            <div class="flex items-center justify-between pt-3 border-t border-gray-100">
              <div>
                <p class="text-xs text-gray-500">TP/SL Status</p>
                <p class="text-sm font-medium" :class="getTpSlStatusColor(trade.tp_sl_status)">
                  {{ getTpSlStatusBadge(trade.tp_sl_status) }}
                </p>
              </div>
              <div class="text-right">
                <p class="text-xs text-gray-500">PnL</p>
                <p class="text-lg font-bold" :class="getPnLColor(trade.pnl)">
                  {{ formatPnL(trade.pnl) }}
                </p>
              </div>
            </div>

            <div class="flex items-center justify-between text-xs text-gray-500">
              <span>{{ trade.interval }} • {{ trade.confidence.toFixed(0) }}% conf</span>
              <span>{{ new Date(trade.created_at).toLocaleDateString() }}</span>
            </div>
          </div>

          <!-- Desktop View -->
          <div class="hidden md:grid grid-cols-12 gap-4 items-center">
            <div class="col-span-2">
              <div class="flex items-center gap-2">
                <PhCurrencyBtc :size="20" class="text-blue-600" weight="fill" />
                <div>
                  <p class="font-bold text-gray-900">{{ trade.symbol }}</p>
                  <p class="text-xs text-gray-500">{{ trade.interval }}</p>
                </div>
              </div>
            </div>

            <div class="col-span-1">
              <span
                class="px-2 py-1 text-xs font-bold rounded-full"
                :class="getSideBgColor(trade.side)"
              >
                {{ trade.side.toUpperCase() }}
              </span>
            </div>

            <div class="col-span-2">
              <div class="text-sm">
                <p class="text-xs text-gray-500">Entry → Exit</p>
                <p class="font-semibold">
                  {{ formatPrice(trade.avg_entry_price) }}
                  <span v-if="trade.exit_price > 0" class="text-gray-400"> → </span>
                  <span v-if="trade.exit_price > 0" :class="trade.pnl >= 0 ? 'text-green-600' : 'text-red-600'">
                    {{ formatPrice(trade.exit_price) }}
                  </span>
                </p>
              </div>
            </div>

            <div class="col-span-2">
              <div class="text-sm">
                <p class="text-xs text-gray-500">TP / SL</p>
                <div class="flex items-center gap-2">
                  <PhTarget :size="14" class="text-green-600" />
                  <span class="font-medium text-green-600">{{ formatPrice(trade.tp_price) }}</span>
                </div>
                <div class="flex items-center gap-2">
                  <PhStopCircle :size="14" class="text-red-600" />
                  <span class="font-medium text-red-600">{{ formatPrice(trade.sl_price) }}</span>
                </div>
              </div>
            </div>

            <div class="col-span-1">
              <span
                class="px-2 py-1 text-xs font-semibold rounded"
                :class="getStatusColor(trade.status)"
              >
                {{ trade.status }}
              </span>
            </div>

            <div class="col-span-2 text-right">
              <div class="text-sm">
                <p
                  class="text-lg font-bold"
                  :class="getPnLColor(trade.pnl)"
                >
                  {{ trade.pnl >= 0 ? '+' : '' }}{{ trade.pnl.toFixed(2) }} USDT
                </p>
                <p class="text-xs" :class="getPnLColor(trade.pnl_pct)">
                  {{ trade.pnl_pct >= 0 ? '+' : '' }}{{ trade.pnl_pct.toFixed(2) }}%
                </p>
              </div>
            </div>

            <div class="col-span-2 text-right">
              <div class="text-sm">
                <p class="text-xs text-gray-500">Date</p>
                <p class="font-medium text-gray-700">{{ formatDate(trade.created_at).split(',')[0] }}</p>
                <p class="text-xs text-gray-500">{{ formatDate(trade.created_at).split(',')[1] }}</p>
              </div>
            </div>
          </div>

          <!-- TP/SL Status Badge -->
          <div class="mt-3 pt-3 border-t border-gray-100 flex items-center justify-between">
            <div class="flex items-center gap-2 text-xs">
              <PhInfo :size="14" class="text-gray-400" />
              <span class="text-gray-600">TP/SL Status:</span>
              <span
                class="font-semibold"
                :class="getTpSlStatusColor(trade.tp_sl_status)"
              >
                {{ getTpSlStatusBadge(trade.tp_sl_status) }}
              </span>
            </div>
            <div class="flex items-center gap-4 text-xs text-gray-500">
              <span>Leverage: {{ trade.leverage }}x</span>
              <span>Capital: {{ trade.capital_used.toFixed(2) }} USDT</span>
              <span>Confidence: {{ trade.confidence.toFixed(0) }}%</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Trade Detail Modal (Nested) -->
    <UiModal
      v-if="selectedTrade"
      :model-value="!!selectedTrade"
      :title="`${selectedTrade.symbol} - Trade Details`"
      size="lg"
      @update:model-value="selectedTrade = null"
    >
      <div class="space-y-4">
        <!-- Header -->
        <div class="flex items-center justify-between p-4 bg-gradient-to-r from-gray-50 to-gray-100 rounded-xl">
          <div class="flex items-center gap-3">
            <PhCurrencyBtc :size="32" class="text-blue-600" weight="fill" />
            <div>
              <h3 class="text-xl font-bold text-gray-900">{{ selectedTrade.symbol }}</h3>
              <p class="text-sm text-gray-500">{{ selectedTrade.interval }} • {{ selectedTrade.side.toUpperCase() }}</p>
            </div>
          </div>
          <div class="text-right">
            <p class="text-sm text-gray-500">Status</p>
            <span
              class="px-3 py-1 text-sm font-semibold rounded-full"
              :class="getStatusColor(selectedTrade.status)"
            >
              {{ selectedTrade.status }}
            </span>
          </div>
        </div>

        <!-- PnL Display -->
        <div class="p-6 bg-white rounded-xl border-2" :class="getPnLColor(selectedTrade.pnl) ? 'border-green-200' : 'border-red-200'">
          <p class="text-sm text-gray-500 mb-1">Total PnL</p>
          <p class="text-4xl font-bold" :class="getPnLColor(selectedTrade.pnl)">
            {{ formatPnL(selectedTrade.pnl) }}
          </p>
        </div>

        <!-- Trade Info Grid -->
        <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
          <div class="p-4 bg-blue-50 rounded-xl">
            <p class="text-xs text-gray-600 mb-1">Avg Entry</p>
            <p class="text-lg font-bold text-gray-900">{{ formatPrice(selectedTrade.avg_entry_price) }}</p>
          </div>
          <div class="p-4 bg-green-50 rounded-xl">
            <p class="text-xs text-gray-600 mb-1">Take Profit</p>
            <p class="text-lg font-bold text-green-700">{{ formatPrice(selectedTrade.tp_price) }}</p>
          </div>
          <div class="p-4 bg-red-50 rounded-xl">
            <p class="text-xs text-gray-600 mb-1">Stop Loss</p>
            <p class="text-lg font-bold text-red-700">{{ formatPrice(selectedTrade.sl_price) }}</p>
          </div>
          <div class="p-4 bg-gray-50 rounded-xl">
            <p class="text-xs text-gray-600 mb-1">Leverage</p>
            <p class="text-lg font-bold text-gray-900">{{ selectedTrade.leverage }}x</p>
          </div>
          <div class="p-4 bg-gray-50 rounded-xl">
            <p class="text-xs text-gray-600 mb-1">Capital</p>
            <p class="text-lg font-bold text-gray-900">{{ selectedTrade.capital_used.toFixed(2) }} USDT</p>
          </div>
          <div class="p-4 bg-gray-50 rounded-xl">
            <p class="text-xs text-gray-600 mb-1">Confidence</p>
            <p class="text-lg font-bold text-gray-900">{{ selectedTrade.confidence.toFixed(1) }}%</p>
          </div>
        </div>

        <!-- Additional Info -->
        <div class="p-4 bg-gray-50 rounded-xl">
          <p class="text-sm font-semibold text-gray-700 mb-2">Additional Information</p>
          <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span class="text-gray-500">TP/SL Status:</span>
              <span class="ml-2 font-semibold" :class="getTpSlStatusColor(selectedTrade.tp_sl_status)">
                {{ getTpSlStatusBadge(selectedTrade.tp_sl_status) }}
              </span>
            </div>
            <div>
              <span class="text-gray-500">Risk/Reward:</span>
              <span class="ml-2 font-semibold">{{ selectedTrade.risk_reward_ratio.toFixed(2) }}</span>
            </div>
            <div>
              <span class="text-gray-500">Total Qty:</span>
              <span class="ml-2 font-semibold">{{ selectedTrade.total_qty }}</span>
            </div>
            <div>
              <span class="text-gray-500">Mode:</span>
              <span class="ml-2 font-semibold">{{ selectedTrade.is_aggressive ? 'Aggressive' : 'Conservative' }}</span>
            </div>
          </div>
        </div>

        <!-- Orders -->
        <div v-if="selectedTrade.orders && selectedTrade.orders.length > 0">
          <p class="text-sm font-semibold text-gray-700 mb-3">Orders ({{ selectedTrade.orders.length }})</p>
          <div class="space-y-2">
            <div
              v-for="order in selectedTrade.orders"
              :key="order.entry_number"
              class="p-3 bg-white border border-gray-200 rounded-lg"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <div
                    class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold"
                    :class="{
                      'bg-blue-100 text-blue-700': order.status === 'PENDING',
                      'bg-green-100 text-green-700': order.status === 'FILLED',
                      'bg-gray-100 text-gray-700': order.status === 'CANCELLED',
                      'bg-red-100 text-red-700': order.status === 'REJECTED'
                    }"
                  >
                    {{ order.entry_number }}
                  </div>
                  <div>
                    <p class="text-sm font-semibold text-gray-900">
                      {{ order.type }} • {{ order.quantity }} {{ selectedTrade.symbol.replace('USDT', '') }}
                    </p>
                    <p class="text-xs text-gray-500">
                      @ {{ formatPrice(order.price) }} • {{ order.status }}
                    </p>
                  </div>
                </div>
                <div class="text-right">
                  <p class="text-sm font-bold text-gray-900">
                    {{ (order.price * order.quantity).toFixed(2) }} USDT
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Timestamps -->
        <div class="p-4 bg-gray-50 rounded-xl text-sm">
          <div class="flex items-center justify-between mb-2">
            <span class="text-gray-500">Created:</span>
            <span class="font-medium text-gray-900">{{ formatDate(selectedTrade.created_at) }}</span>
          </div>
          <div class="flex items-center justify-between mb-2">
            <span class="text-gray-500">Updated:</span>
            <span class="font-medium text-gray-900">{{ formatDate(selectedTrade.updated_at) }}</span>
          </div>
          <div v-if="selectedTrade.closed_at" class="flex items-center justify-between">
            <span class="text-gray-500">Closed:</span>
            <span class="font-medium text-gray-900">{{ formatDate(selectedTrade.closed_at) }}</span>
          </div>
        </div>

        <!-- Close Button -->
        <div class="flex justify-end pt-4 border-t border-gray-200">
          <button
            @click="selectedTrade = null"
            class="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-all"
          >
            Close
          </button>
        </div>
      </div>
    </UiModal>
  </UiModal>
</template>

<style scoped>
</style>
