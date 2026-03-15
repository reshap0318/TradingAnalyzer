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
import { formatPrice, formatPnL } from '@/helpers/formatters'
import {
  PhTrendUp,
  PhTarget,
  PhStopCircle,
  PhCurrencyBtc,
  PhInfo
} from '@phosphor-icons/vue'

interface ITradeExecutedCardProps {
  trades: ITrade[]
}

const props = withDefaults(defineProps<ITradeExecutedCardProps>(), {
  trades: () => [],
})

// Modal state
const showModal = ref(false)

// Calculate total PnL from all executed trades
const totalPnL = computed(() => {
  return props.trades.reduce((sum, trade) => sum + (trade.pnl || 0), 0)
})

// Calculate win rate
const winRate = computed(() => {
  if (props.trades.length === 0) return 0
  const winningTrades = props.trades.filter(t => t.pnl > 0).length
  return (winningTrades / props.trades.length) * 100
})

const handleClose = () => {
  showModal.value = false
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
      <span class="text-sm text-gray-600">Trades Executed</span>
    </div>
    <div class="text-3xl font-bold text-gray-900">{{ trades.length }}</div>
    <p v-if="trades.length > 0" class="text-xs text-gray-500 mt-1">
      This session
    </p>
  </div>

  <!-- Modal - Full trade details -->
  <UiModal
    :model-value="showModal"
    title="Executed Trades"
    size="full"
    @update:model-value="handleClose"
  >
    <!-- Empty State -->
    <div v-if="trades.length === 0" class="text-center py-12">
      <PhCurrencyBtc :size="64" class="mx-auto text-gray-300 mb-4" />
      <p class="text-gray-500 text-lg font-medium">No executed trades yet</p>
      <p class="text-gray-400 text-sm mt-1">Trades will appear here when the bot executes them</p>
    </div>

    <!-- Trades List -->
    <div v-else class="space-y-6">
      <!-- Summary Stats -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 p-4 bg-gray-50 rounded-xl">
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
          <p class="text-2xl font-bold" :class="getPnLColor(totalPnL)">
            {{ totalPnL >= 0 ? '+' : '' }}{{ totalPnL.toFixed(2) }} USDT
          </p>
        </div>
      </div>

      <!-- Trades Table Header -->
      <div class="hidden md:grid grid-cols-12 gap-4 px-4 py-2 bg-gray-100 rounded-lg text-xs font-semibold text-gray-600">
        <div class="col-span-2">Symbol</div>
        <div class="col-span-1">Side</div>
        <div class="col-span-2">Entry → Exit</div>
        <div class="col-span-2">TP / SL</div>
        <div class="col-span-1">Status</div>
        <div class="col-span-2 text-right">PnL</div>
        <div class="col-span-2 text-right">Time</div>
      </div>

      <!-- Trade Items -->
      <div
        v-for="trade in trades"
        :key="trade.id"
        class="border border-gray-200 rounded-xl p-4 hover:shadow-md transition-all"
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
              <p class="font-semibold">{{ formatPrice(trade.exit_price) }}</p>
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
            <span>{{ trade.interval }}</span>
            <span>{{ new Date(trade.created_at).toLocaleString() }}</span>
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
                <span v-if="trade.exit_price > 0" class="text-gray-400">→</span>
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
              <p class="text-xs text-gray-500">Opened</p>
              <p class="font-medium text-gray-700">{{ new Date(trade.created_at).toLocaleTimeString() }}</p>
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
          </div>
        </div>
      </div>
    </div>
  </UiModal>
</template>

<style scoped>
</style>
