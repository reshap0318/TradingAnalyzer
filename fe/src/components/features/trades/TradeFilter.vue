<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useWatchlistStore } from '@/stores/watchlist.store'
import { useTimeframeStore } from '@/stores/timeframe.store'
import { UiModal } from '@/components/common'
import {
  PhFunnel,
  PhCalendar,
  PhArrowCounterClockwise
} from '@phosphor-icons/vue'

// Filter state interface
export interface ITradeFilterState {
  status: string[]
  symbol: string[]
  interval: string
  side: string
  min_confidence: number
  date_start: string
  date_end: string
}

// Props
interface ITradeFilterProps {
  modelValue: ITradeFilterState
  loading?: boolean
}

const props = withDefaults(defineProps<ITradeFilterProps>(), {
  loading: false
})

// Emits
const emit = defineEmits<{
  'update:modelValue': [filter: ITradeFilterState]
  'apply': [filter: ITradeFilterState]
  'reset': []
}>()

// Local filter state
const localFilter = ref<ITradeFilterState>({ ...props.modelValue })
const showModal = ref(false)

// Watch for prop changes
watch(() => props.modelValue, (newValue) => {
  localFilter.value = { ...newValue }
}, { deep: true })

// Sync local state to parent
const updateFilter = (updates: Partial<ITradeFilterState>) => {
  localFilter.value = { ...localFilter.value, ...updates }
  emit('update:modelValue', localFilter.value)
}

// Status options
const statusOptions = [
  { value: 'ACTIVE', label: 'Active' },
  { value: 'CLOSED', label: 'Closed' },
  { value: 'CANCELLED', label: 'Cancelled' }
]

// Side options
const sideOptions = [
  { value: 'BUY', label: 'Buy/Long' },
  { value: 'SELL', label: 'Sell/Short' }
]

// Stores
const watchlistStore = useWatchlistStore()
const timeframeStore = useTimeframeStore()

// Get unique symbols from watchlist (using existing store)
const availableSymbols = computed(() => {
  return watchlistStore.activeWatchlists
    .map(w => w.symbol)
    .sort()
})

// Get timeframes from store (using existing store)
const availableTimeframes = computed(() => {
  return timeframeStore.sortedItems.map(tf => ({
    value: tf.name,
    label: tf.name
  }))
})

// Toggle status filter
const toggleStatus = (status: string) => {
  const current = localFilter.value.status
  const updated = current.includes(status)
    ? current.filter(s => s !== status)
    : [...current, status]
  updateFilter({ status: updated })
}

// Toggle symbol filter
const toggleSymbol = (symbol: string) => {
  const current = localFilter.value.symbol
  const updated = current.includes(symbol)
    ? current.filter(s => s !== symbol)
    : [...current, symbol]
  updateFilter({ symbol: updated })
}

// Clear all filters
const clearAllFilters = () => {
  localFilter.value = {
    status: [],
    symbol: [],
    interval: '',
    side: '',
    min_confidence: 0,
    date_start: '',
    date_end: ''
  }
  emit('update:modelValue', localFilter.value)
  emit('reset')
}

// Apply filters
const applyFilters = () => {
  emit('apply', localFilter.value)
  showModal.value = false
}

// Open modal
const openModal = () => {
  showModal.value = true
}

// Check if has active filters
const hasActiveFilters = computed(() => {
  return (
    localFilter.value.status.length > 0 ||
    localFilter.value.symbol.length > 0 ||
    localFilter.value.interval ||
    localFilter.value.side ||
    localFilter.value.min_confidence > 0
  )
})

// Get active filters count
const activeFiltersCount = computed(() => {
  let count = 0
  if (localFilter.value.status.length > 0) count += localFilter.value.status.length
  if (localFilter.value.symbol.length > 0) count += localFilter.value.symbol.length
  if (localFilter.value.interval) count += 1
  if (localFilter.value.side) count += 1
  if (localFilter.value.min_confidence > 0) count += 1
  return count
})
</script>

<template>
  <div class="inline-flex items-center gap-2">
    <!-- Filter Button with Badge -->
    <button
      @click="openModal"
      class="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white text-sm font-medium rounded-lg hover:bg-blue-600 transition-all relative"
      :disabled="loading"
    >
      <PhFunnel :size="16" />
      Filter
      <span
        v-if="activeFiltersCount > 0"
        class="absolute -top-2 -right-2 w-5 h-5 bg-red-500 text-white text-xs font-bold rounded-full flex items-center justify-center"
      >
        {{ activeFiltersCount }}
      </span>
    </button>

    <!-- Reset Button (shown when has filters) -->
    <button
      v-if="hasActiveFilters"
      @click="clearAllFilters"
      class="flex items-center gap-1 px-3 py-2 text-xs font-medium text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200 transition-all"
    >
      <PhArrowCounterClockwise :size="14" />
      Reset
    </button>

    <!-- Filter Modal -->
    <UiModal
      v-model="showModal"
      title="Filter Trades"
      size="xl"
    >
      <div class="space-y-4">
        <!-- Status Filter -->
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-2">Status</label>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="opt in statusOptions"
              :key="opt.value"
              @click="toggleStatus(opt.value)"
              class="px-3 py-1.5 text-sm rounded-lg border transition-all"
              :class="localFilter.status.includes(opt.value)
                ? 'bg-blue-500 text-white border-blue-500'
                : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>

        <!-- Symbol Filter -->
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-2">Symbols</label>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="symbol in availableSymbols"
              :key="symbol"
              @click="toggleSymbol(symbol)"
              class="px-3 py-1.5 text-sm rounded-lg border transition-all"
              :class="localFilter.symbol.includes(symbol)
                ? 'bg-purple-500 text-white border-purple-500'
                : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'"
            >
              {{ symbol }}
            </button>
          </div>
          <p v-if="availableSymbols.length === 0" class="text-sm text-gray-500 italic">
            No active watchlists. Add symbols to your watchlist first.
          </p>
        </div>

        <!-- Timeframe Filter -->
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-2">Timeframe</label>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="tf in availableTimeframes"
              :key="tf.value"
              @click="updateFilter({ interval: tf.value === localFilter.interval ? '' : tf.value })"
              class="px-3 py-1.5 text-sm rounded-lg border transition-all"
              :class="localFilter.interval === tf.value
                ? 'bg-green-500 text-white border-green-500'
                : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'"
            >
              {{ tf.label }}
            </button>
          </div>
          <p v-if="availableTimeframes.length === 0" class="text-sm text-gray-500 italic">
            No timeframes available.
          </p>
        </div>

        <!-- Side Filter -->
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-2">Side</label>
          <div class="flex gap-2">
            <button
              v-for="opt in sideOptions"
              :key="opt.value"
              @click="updateFilter({ side: opt.value === localFilter.side ? '' : opt.value })"
              class="px-4 py-2 text-sm rounded-lg border transition-all"
              :class="localFilter.side === opt.value
                ? 'bg-orange-500 text-white border-orange-500'
                : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>

        <!-- Min Confidence -->
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-2">
            Min Confidence: {{ localFilter.min_confidence }}%
          </label>
          <input
            v-model.number="localFilter.min_confidence"
            type="range"
            min="0"
            max="100"
            step="5"
            class="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
          />
          <div class="flex justify-between text-xs text-gray-500 mt-1">
            <span>0%</span>
            <span>50%</span>
            <span>100%</span>
          </div>
        </div>

        <!-- Date Range -->
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-semibold text-gray-700 mb-2">
              <PhCalendar :size="16" class="inline mr-1" />
              Start Date
            </label>
            <input
              v-model="localFilter.date_start"
              type="date"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
          <div>
            <label class="block text-sm font-semibold text-gray-700 mb-2">
              <PhCalendar :size="16" class="inline mr-1" />
              End Date
            </label>
            <input
              v-model="localFilter.date_end"
              type="date"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
        </div>

        <!-- Actions -->
        <div class="flex justify-end gap-3 pt-4 border-t border-gray-200">
          <button
            @click="clearAllFilters"
            class="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-all"
          >
            Clear All
          </button>
          <button
            @click="applyFilters"
            class="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-all"
          >
            Apply Filters
          </button>
        </div>
      </div>
    </UiModal>
  </div>
</template>

<style scoped>
</style>
