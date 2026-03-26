<script setup lang="ts">
import type { ITrade } from '@/stores/tradebot.store'
import {
  getSideBgColor,
  getStatusColor,
  getTpSlStatusColor,
  getPnLColor,
  getEntryModeBadge,
  getEntryModeColor,
  getFilledOrdersCount
} from '@/helpers/trade'
import { formatPrice } from '@/helpers/formatters'
import {
  PhCurrencyBtc,
  PhTarget,
  PhStopCircle,
  PhChartLineUp,
  PhInfo,
  PhClock,
  PhArrowRight
} from '@phosphor-icons/vue'
import { useRouter } from 'vue-router'

interface ITradeCardProps {
  trade: ITrade
}

const props = defineProps<ITradeCardProps>()
const router = useRouter()

const handleViewSignal = () => {
  if (props.trade.signal_log_id) {
    router.push(`/signals/${props.trade.signal_log_id}`)
  }
}
</script>

<template>
  <div class="bg-gradient-to-br from-gray-50 to-white rounded-xl border border-gray-200 overflow-hidden">
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
                :class="getSideBgColor(trade.side)"
              >
                {{ trade.side.toUpperCase() }}
              </span>
            </div>
            <p class="text-xs text-gray-500">{{ trade.interval }} • Opened: {{ new Date(trade.created_at).toLocaleString() }}</p>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <span
            class="px-3 py-1.5 text-xs font-semibold rounded-lg"
            :class="getStatusColor(trade.status)"
          >
            {{ trade.status }}
          </span>
        </div>
      </div>
    </div>

    <!-- Card Body -->
    <div class="p-4 bg-white">
      <!-- Entry Mode & Confidence -->
      <div class="flex items-center justify-between mb-4">
        <span
          class="px-3 py-1 text-xs font-semibold rounded-full"
          :class="getEntryModeColor(trade.is_aggressive, trade.orders)"
        >
          {{ getEntryModeBadge(trade.is_aggressive, trade.orders) }}
        </span>
        <div 
          v-if="trade.signal_log_id"
          @click="handleViewSignal"
          class="flex items-center gap-2 cursor-pointer hover:bg-gray-100 px-3 py-1.5 rounded-lg transition-all duration-200 group"
        >
          <PhInfo :size="16" class="text-gray-400 group-hover:text-blue-500 transition-colors" />
          <span class="text-sm text-gray-600">Confidence: <span class="font-bold">{{ trade.confidence.toFixed(1) }}%</span></span>
          <PhArrowRight :size="16" class="text-gray-400 group-hover:text-blue-500 group-hover:translate-x-1 transition-all" weight="bold" />
        </div>
        <div v-else class="flex items-center gap-2">
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
          <p class="text-sm font-bold text-gray-900">{{ formatPrice(trade.avg_entry_price) }}</p>
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

      <!-- PnL Display -->
      <div class="p-4 bg-white rounded-lg border-2 mb-4" :class="getPnLColor(trade.pnl) ? 'border-green-200' : 'border-red-200'">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs text-gray-500 mb-1">PnL</p>
            <p class="text-2xl font-bold" :class="getPnLColor(trade.pnl)">
              {{ trade.pnl >= 0 ? '+' : '' }}{{ trade.pnl.toFixed(2) }} USDT
            </p>
          </div>
          <div class="text-right">
            <p class="text-xs text-gray-500 mb-1">PnL %</p>
            <p class="text-xl font-bold" :class="getPnLColor(trade.pnl_pct)">
              {{ trade.pnl_pct >= 0 ? '+' : '' }}{{ trade.pnl_pct.toFixed(2) }}%
            </p>
          </div>
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

      <!-- TP/SL Status Badge -->
      <div class="mt-4 pt-4 border-t border-gray-200 flex items-center justify-between">
        <div class="flex items-center gap-2 text-xs">
          <PhInfo :size="14" class="text-gray-400" />
          <span class="text-gray-600">Exit Reason:</span>
          <span
            class="font-semibold px-2 py-1 rounded bg-gray-100"
            :class="getTpSlStatusColor(trade.tp_sl_status)"
          >
            {{ trade.exit_reason || '-' }}
          </span>
        </div>
        <div class="flex items-center gap-4 text-xs text-gray-500">
          <span>Exit: {{ formatPrice(trade.exit_price || 0) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
