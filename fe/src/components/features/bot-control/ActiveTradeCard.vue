<script setup lang="ts">
import { ref } from 'vue'
import { UiModal } from '@/components/common'
import type { ITrade, ITradeOrder } from '@/stores/tradebot.store'
import {
  PhWallet,
  PhTarget,
  PhStopCircle,
  PhClock,
  PhChartLineUp,
  PhCurrencyBtc,
  PhInfo
} from '@phosphor-icons/vue'

interface IActiveTradeCardProps {
  trades: ITrade[]
}

const props = withDefaults(defineProps<IActiveTradeCardProps>(), {
  trades: () => [],
})

// Modal state
const showModal = ref(false)

// Calculate average entry price from filled orders only
const calculateAvgEntry = (orders: ITradeOrder[]): number => {
  const filledOrders = orders.filter(o => o.status === 'FILLED')
  if (filledOrders.length === 0) return 0

  const totalValue = filledOrders.reduce((sum, order) => sum + (order.price * order.quantity), 0)
  const totalQty = filledOrders.reduce((sum, order) => sum + order.quantity, 0)

  return totalQty > 0 ? totalValue / totalQty : 0
}

// Get filled orders count
const getFilledOrdersCount = (orders: ITradeOrder[]): number => {
  return orders.filter(o => o.status === 'FILLED').length
}

// Format price with proper decimals
const formatPrice = (price: number): string => {
  if (price === 0) return '-'
  if (price >= 1000) return price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  if (price >= 1) return price.toFixed(4)
  return price.toFixed(6)
}

// Get side color
const getSideColor = (side: string): string => {
  return side.toUpperCase() === 'BUY' || side.toUpperCase() === 'LONG' ? 'text-green-600' : 'text-red-600'
}

// Get entry mode badge
const getEntryModeBadge = (isAggressive: boolean, orders: ITradeOrder[]): string => {
  const filledCount = getFilledOrdersCount(orders)
  if (!isAggressive) return 'Conservative (1 Entry)'
  return filledCount === 1 ? 'Aggressive (1/2 Filled)' : 'Aggressive (2/2 Filled)'
}

// Get entry mode color
const getEntryModeColor = (isAggressive: boolean, orders: ITradeOrder[]): string => {
  const filledCount = getFilledOrdersCount(orders)
  if (!isAggressive) return 'bg-blue-100 text-blue-700'
  if (filledCount === 1) return 'bg-orange-100 text-orange-700'
  return 'bg-purple-100 text-purple-700'
}

const handleClose = () => {
  showModal.value = false
}
</script>

<template>
  <!-- Card - Clickable to open modal -->
  <div
    class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6 cursor-pointer hover:shadow-xl hover:border-green-300 transition-all duration-200"
    @click="showModal = true"
  >
    <div class="flex items-center gap-3 mb-4">
      <div class="p-3 bg-green-50 rounded-xl">
        <PhWallet :size="24" class="text-green-600" weight="fill" />
      </div>
      <span class="text-sm text-gray-600">Active Trades</span>
    </div>
    <div class="text-3xl font-bold text-gray-900">{{ trades.length }}</div>
    <p v-if="trades.length > 0" class="text-xs text-gray-500 mt-1">
      From session
    </p>
    <p v-if="trades.length > 0" class="text-xs text-green-600 mt-2 font-medium">Click to view details →</p>
  </div>

  <!-- Modal - Full trade details -->
  <UiModal
    :model-value="showModal"
    title="Active Trades"
    size="full"
    @update:model-value="handleClose"
  >
    <!-- Empty State -->
    <div v-if="trades.length === 0" class="text-center py-12">
      <PhCurrencyBtc :size="64" class="mx-auto text-gray-300 mb-4" />
      <p class="text-gray-500 text-lg font-medium">No active trades</p>
      <p class="text-gray-400 text-sm mt-1">Trades will appear here when the bot is executing</p>
    </div>

    <!-- Trades List -->
    <div v-else class="space-y-4">
      <!-- Subtitle -->
      <div class="flex items-center justify-between mb-2">
        <p class="text-sm text-gray-600">
          <span class="font-semibold">{{ trades.length }}</span>
          {{ trades.length === 1 ? 'trade' : 'trades' }} currently active
        </p>
      </div>

      <div
        v-for="trade in trades"
        :key="trade.id"
        class="bg-gradient-to-br from-gray-50 to-white rounded-xl border border-gray-200 overflow-hidden"
      >
        <!-- Card Header -->
        <div class="p-4 bg-gradient-to-r from-gray-100 to-gray-50 border-b border-gray-200">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-white rounded-lg shadow-sm">
                <PhCurrencyBtc :size="24" class="text-blue-600" weight="fill" />
              </div>
              <div>
                <div class="flex items-center gap-2">
                  <h4 class="text-lg font-bold text-gray-900">{{ trade.symbol }}</h4>
                  <span
                    class="px-2 py-1 text-xs font-bold rounded-full"
                    :class="getSideColor(trade.side)"
                  >
                    {{ trade.side.toUpperCase() }}
                  </span>
                </div>
                <p class="text-xs text-gray-500">{{ trade.interval }} • Opened: {{ new Date(trade.created_at).toLocaleString() }}</p>
              </div>
            </div>
            <div class="text-right">
              <div class="text-sm font-semibold text-gray-600">Status</div>
              <div class="text-sm font-bold text-green-600">{{ trade.status }}</div>
            </div>
          </div>
        </div>

        <!-- Card Body -->
        <div class="p-4">
          <!-- Entry Mode & Confidence -->
          <div class="flex items-center justify-between mb-4">
            <span
              class="px-3 py-1 text-xs font-semibold rounded-full"
              :class="getEntryModeColor(trade.is_aggressive, trade.orders)"
            >
              {{ getEntryModeBadge(trade.is_aggressive, trade.orders) }}
            </span>
            <div class="flex items-center gap-2">
              <PhInfo :size="16" class="text-gray-400" />
              <span class="text-sm text-gray-600">Confidence: <span class="font-bold">{{ trade.confidence.toFixed(1) }}%</span></span>
            </div>
          </div>

          <!-- Price Info Grid -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3 mb-4">
            <!-- Average Entry -->
            <div class="p-3 bg-blue-50 rounded-lg">
              <div class="flex items-center gap-1 mb-1">
                <PhChartLineUp :size="14" class="text-blue-600" />
                <span class="text-xs text-gray-600">Avg Entry</span>
              </div>
              <p class="text-sm font-bold text-gray-900">{{ formatPrice(calculateAvgEntry(trade.orders) || trade.avg_entry_price) }}</p>
              <p class="text-xs text-gray-500 mt-1">{{ getFilledOrdersCount(trade.orders) }}/{{ trade.is_aggressive ? 2 : 1 }} filled</p>
            </div>

            <!-- Take Profit -->
            <div class="p-3 bg-green-50 rounded-lg">
              <div class="flex items-center gap-1 mb-1">
                <PhTarget :size="14" class="text-green-600" />
                <span class="text-xs text-gray-600">Take Profit</span>
              </div>
              <p class="text-sm font-bold text-green-700">{{ formatPrice(trade.tp_price) }}</p>
              <p class="text-xs text-gray-500 mt-1">TP Order #{{ trade.tp_order_id || '-' }}</p>
            </div>

            <!-- Stop Loss -->
            <div class="p-3 bg-red-50 rounded-lg">
              <div class="flex items-center gap-1 mb-1">
                <PhStopCircle :size="14" class="text-red-600" />
                <span class="text-xs text-gray-600">Stop Loss</span>
              </div>
              <p class="text-sm font-bold text-red-700">{{ formatPrice(trade.sl_price) }}</p>
              <p class="text-xs text-gray-500 mt-1">SL Order #{{ trade.sl_order_id || '-' }}</p>
            </div>
          </div>

          <!-- Trade Details -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mb-4">
            <div class="text-center p-2 bg-gray-50 rounded-lg">
              <p class="text-xs text-gray-500 mb-1">Leverage</p>
              <p class="text-lg font-bold text-gray-900">{{ trade.leverage }}x</p>
            </div>
            <div class="text-center p-2 bg-gray-50 rounded-lg">
              <p class="text-xs text-gray-500 mb-1">Capital</p>
              <p class="text-lg font-bold text-gray-900">{{ trade.capital_used.toFixed(2) }}</p>
              <p class="text-xs text-gray-500">USDT</p>
            </div>
            <div class="text-center p-2 bg-gray-50 rounded-lg">
              <p class="text-xs text-gray-500 mb-1">R:R Ratio</p>
              <p class="text-lg font-bold text-gray-900">{{ trade.risk_reward_ratio.toFixed(2) }}</p>
            </div>
            <div class="text-center p-2 bg-gray-50 rounded-lg">
              <p class="text-xs text-gray-500 mb-1">Total Qty</p>
              <p class="text-lg font-bold text-gray-900">{{ trade.total_qty }}</p>
            </div>
          </div>

          <!-- Orders List -->
          <div v-if="trade.orders && trade.orders.length > 0">
            <div class="flex items-center gap-2 mb-2">
              <PhClock :size="16" class="text-gray-500" />
              <span class="text-sm font-semibold text-gray-700">Orders ({{ trade.orders.length }})</span>
            </div>
            <div class="space-y-2">
              <div
                v-for="order in trade.orders"
                :key="order.entry_number"
                class="flex items-center justify-between p-3 bg-white border border-gray-200 rounded-lg"
              >
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
                      {{ order.type }} • {{ order.quantity }} {{ trade.symbol.replace('USDT', '') }}
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
                  <p class="text-xs text-gray-500">
                    Order #{{ order.binance_order_id }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </UiModal>
</template>

<style scoped>
</style>
