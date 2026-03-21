<script setup lang="ts">
import { ref, watch } from 'vue'
import { formatCurrency, formatPercent, getSideColor, getTradeStatusColor } from '@/helpers/backtest'
import type { IBacktestTrade } from '@/stores/backtest.store'
import UiModal from '@/components/common/UiModal.vue'
import { formatDate, formatDuration } from '@/helpers/formatters'

interface Props {
  show: boolean
  trade: IBacktestTrade | null
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  trade: null
})

const emit = defineEmits<{
  close: []
}>()

const isOpen = ref(false)

// Sync dengan props.show
watch(
  () => props.show,
  (newVal) => {
    isOpen.value = newVal
  }
)

// Emit close event
watch(
  () => isOpen.value,
  (newVal) => {
    if (!newVal) {
      emit('close')
    }
  }
)
</script>

<template>
  <UiModal v-model="isOpen" title="Trade Detail" size="lg">
    <template v-if="trade">
      <!-- Basic Info -->
      <div class="grid grid-cols-2 gap-4 mb-4">
        <div>
          <p class="text-xs text-gray-500">Side</p>
          <span
            class="inline-block mt-1 px-2 py-1 text-sm font-bold rounded"
            :class="getSideColor(trade.side)"
          >
            {{ trade.side }}
          </span>
        </div>
        <div class="text-end">
          <p class="text-xs text-gray-500">Status</p>
          <span
            class="inline-block mt-1 px-2 py-1 text-xs rounded-full"
            :class="getTradeStatusColor(trade.status)"
          >
            {{ trade.status }}
          </span>
        </div>
      </div>

      <!-- PnL Card -->
      <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-4 text-center mb-4">
        <div
          class="text-3xl font-bold"
          :class="trade.pnl >= 0 ? 'text-green-600' : 'text-red-600'"
        >
          {{ trade.pnl >= 0 ? '+' : '' }}{{ formatCurrency(trade.pnl) }}
        </div>
        <div
          class="text-sm font-medium mt-1"
          :class="trade.pnl_percent >= 0 ? 'text-green-600' : 'text-red-600'"
        >
          {{ formatPercent(trade.pnl_percent) }}
        </div>
      </div>

      <!-- Trade Details Card -->
      <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-4 mb-4">
        <h4 class="text-sm font-semibold text-gray-700 mb-3">Trade Details</h4>
        <div class="space-y-3">
          <div class="flex justify-between py-2 border-b border-gray-100 last:border-0">
            <span class="text-sm text-gray-500">Signal</span>
            <span class="text-sm font-medium text-gray-900">{{ trade.signal }}</span>
          </div>
          <div class="flex justify-between py-2 border-b border-gray-100 last:border-0">
            <span class="text-sm text-gray-500">Confidence</span>
            <span class="text-sm font-medium text-gray-900">{{ trade.confidence }}%</span>
          </div>
          <div class="flex justify-between py-2 border-b border-gray-100 last:border-0">
            <span class="text-sm text-gray-500">Trading Mode</span>
            <span class="text-sm font-medium text-gray-900">{{ trade.trading_mode }}</span>
          </div>
          <div class="flex justify-between py-2 border-b border-gray-100 last:border-0">
            <span class="text-sm text-gray-500">Avg Entry Price</span>
            <span class="text-sm font-medium text-gray-900">
              {{ trade.avg_entry_price.toFixed(2) }}
            </span>
          </div>
          <div class="flex justify-between py-2 border-b border-gray-100 last:border-0">
            <span class="text-sm text-gray-500">Total Qty</span>
            <span class="text-sm font-medium text-gray-900">{{ trade.total_qty }}</span>
          </div>
          <div class="flex justify-between py-2 border-b border-gray-100 last:border-0">
            <span class="text-sm text-gray-500">Total Capital</span>
            <span class="text-sm font-medium text-gray-900">
              {{ formatCurrency(trade.total_capital) }}
            </span>
          </div>
          <div class="flex justify-between py-2 border-b border-gray-100 last:border-0">
            <span class="text-sm text-gray-500">Duration</span>
            <span class="text-sm font-medium text-gray-900">
              {{ formatDuration(trade.duration_minutes * 60) }}
            </span>
          </div>
        </div>
      </div>

      <!-- Entries Card -->
      <div v-if="trade.entries && trade.entries.length > 0" class="bg-white border border-gray-200 rounded-xl shadow-sm p-4 mb-4">
        <h4 class="text-sm font-semibold text-gray-700 mb-3">Entries ({{ trade.entries.length }})</h4>
        <div class="space-y-3">
          <div
            v-for="entry in trade.entries"
            :key="entry.entry_num"
            class="border border-gray-200 rounded-lg p-3 bg-gray-50"
          >
            <div class="flex justify-between mb-2">
              <span class="text-sm font-medium text-gray-700">Entry #{{ entry.entry_num }}</span>
              <span class="text-sm font-semibold text-gray-900">{{ entry.type }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <div>
                <span class="text-gray-500">Price:</span>
                <span class="ml-1 font-medium text-gray-900">{{ entry.price.toFixed(2) }}</span>
              </div>
              <div>
                <span class="text-gray-500">Qty:</span>
                <span class="ml-1 font-medium text-gray-900">{{ entry.qty }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Exit Card -->
      <div v-if="trade.exit" class="bg-white border border-gray-200 rounded-xl shadow-sm p-4 mb-4">
        <h4 class="text-sm font-semibold text-gray-700 mb-3">Exit Information</h4>
        <div class="space-y-3">
          <div class="flex justify-between py-2 border-b border-gray-100 last:border-0">
            <span class="text-sm text-gray-500">Reason</span>
            <span class="text-sm font-medium text-gray-900">{{ trade.exit.reason }}</span>
          </div>
          <div class="flex justify-between py-2 border-b border-gray-100 last:border-0">
            <span class="text-sm text-gray-500">Price</span>
            <span class="text-sm font-medium text-gray-900">{{ trade.exit.price.toFixed(2) }}</span>
          </div>
          <div class="flex justify-between py-2">
            <span class="text-sm text-gray-500">Time</span>
            <span class="text-sm font-medium text-gray-900">
              {{ formatDate(trade.exit.timestamp) }}
            </span>
          </div>
        </div>
      </div>
    </template>
  </UiModal>
</template>
