<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSignalStore } from '@/stores/signal.store'
import { formatPrice, formatDate } from '@/helpers/formatters'
import {
  PhTrash,
  PhMagnifyingGlass,
  PhArrowCounterClockwise,
  PhX,
  PhFunnel,
  PhTarget,
  PhStopCircle,
  PhChartLineUp,
  PhCurrencyBtc,
  PhClock
} from '@phosphor-icons/vue'
import DefaultLayout from '@/layouts/DefaultLayout.vue'
import { showConfirmWithHtml, showError } from '@/lib/sweetalert'

const router = useRouter()
const signalStore = useSignalStore()

// Search & filters
const searchSymbol = ref('')
const showFilters = ref(false)
const selectedCategory = ref('')
const signalValidFilter = ref('')

// Computed
const isLoading = computed(() => signalStore.loading)
const signals = computed(() => signalStore.signals)
const pagination = computed(() => signalStore.pagination)
const validSignalsCount = computed(() => signalStore.validSignalsCount)

// Get badge color for signal category
function getCategoryBadgeClass(category: string): string {
  const colors: Record<string, string> = {
    STRONG_BUY: 'bg-green-100 text-green-800 border-green-300',
    BUY: 'bg-blue-100 text-blue-800 border-blue-300',
    WAIT: 'bg-gray-100 text-gray-800 border-gray-300',
    SELL: 'bg-orange-100 text-orange-800 border-orange-300',
    STRONG_SELL: 'bg-red-100 text-red-800 border-red-300'
  }
  return colors[category] || 'bg-gray-100 text-gray-800 border-gray-300'
}

// Get icon for signal category
function getCategoryIcon(category: string): string {
  const icons: Record<string, string> = {
    STRONG_BUY: '⬆️',
    BUY: '↑',
    WAIT: '⏸',
    SELL: '↓',
    STRONG_SELL: '⬇'
  }
  return icons[category] || '•'
}

// Handle search
function handleSearch() {
  signalStore.updateFiltersAndFetch({
    symbol: searchSymbol.value.toUpperCase(),
    page: 1
  })
}

// Handle filter change
function handleFilterChange() {
  const filters: any = { page: 1 }

  if (selectedCategory.value) {
    filters.signal_category = selectedCategory.value
  }

  if (signalValidFilter.value !== '') {
    filters.signal_valid = signalValidFilter.value === 'valid'
  }

  signalStore.updateFiltersAndFetch(filters)
}

// Clear filters
function clearFilters() {
  searchSymbol.value = ''
  selectedCategory.value = ''
  signalValidFilter.value = ''
  signalStore.updateFiltersAndFetch({
    symbol: '',
    signal_category: '',
    signal_valid: undefined,
    page: 1
  })
}

// View signal detail
function viewDetail(signalId: number) {
  router.push(`/signals/${signalId}`)
}

// Delete signal
async function handleDelete(signalId: number, event: Event) {
  event.stopPropagation()
  await signalStore.deleteSignalById(signalId)
}

// Open cleanup modal with SweetAlert
async function openCleanupModal() {
  const htmlContent = `
    <div class="text-left space-y-4">
      <p class="text-sm text-gray-600">Delete all signals older than specified hours. This action cannot be undone.</p>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">Hours (1-720)</label>
        <input
          type="number"
          id="cleanup-hours"
          min="1"
          max="720"
          value="720"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
        />
        <p class="mt-2 text-xs text-gray-500">Recommended: 720 hours (30 days)</p>
      </div>
    </div>
  `

  const result = await showConfirmWithHtml(
    'Cleanup Old Signals',
    htmlContent,
    'warning',
    'Delete',
    'Cancel'
  )

  if (result.value) {
    const hours = result.value
    if (hours >= 1 && hours <= 720) {
      await signalStore.cleanupOldSignals(hours)
    } else {
      showError('Invalid Input', 'Please enter a value between 1 and 720')
    }
  }
}

// Refresh data
function handleRefresh() {
  signalStore.refreshSignals()
}

// Handle page change
function handlePageChange(page: number) {
  signalStore.changePage(page)
}

onMounted(() => {
  signalStore.fetchSignals()
})
</script>

<template>
  <DefaultLayout>
    <template #header-title>Trading Signals</template>

    <div class="">
      <!-- Action Buttons -->
      <div class="flex justify-end items-center gap-3 mb-6">
        <button
          @click="handleRefresh"
          :disabled="isLoading"
          class="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <PhArrowCounterClockwise :size="16" :class="{ 'animate-spin': isLoading }" />
          Refresh
        </button>
        <button
          @click="openCleanupModal"
          class="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-red-700 bg-white border border-red-300 rounded-lg hover:bg-red-50"
        >
          <PhTrash :size="16" />
          Cleanup
        </button>
      </div>

      <!-- Stats Cards -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5">
          <div class="text-sm font-medium text-gray-500">Total Signals</div>
          <div class="mt-2 text-3xl font-bold text-gray-900">{{ pagination.total_items }}</div>
        </div>
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5">
          <div class="text-sm font-medium text-gray-500">Valid Signals</div>
          <div class="mt-2 text-3xl font-bold text-green-600">{{ validSignalsCount }}</div>
        </div>
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5">
          <div class="text-sm font-medium text-gray-500">Current Page</div>
          <div class="mt-2 text-3xl font-bold text-blue-600">{{ pagination.page }}</div>
        </div>
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5">
          <div class="text-sm font-medium text-gray-500">Total Pages</div>
          <div class="mt-2 text-3xl font-bold text-purple-600">{{ pagination.total_pages }}</div>
        </div>
      </div>

      <!-- Search & Filters -->
      <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5 mb-6">
        <div class="flex flex-col md:flex-row md:items-center gap-4">
          <!-- Search -->
          <div class="flex-1">
            <div class="relative">
              <input
                v-model="searchSymbol"
                @keyup.enter="handleSearch"
                type="text"
                placeholder="Search by symbol (e.g., BTCUSDT)"
                class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
              <PhMagnifyingGlass
                :size="20"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
              />
            </div>
          </div>

          <div class="flex items-center gap-3">
            <button
              @click="handleSearch"
              class="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700"
            >
              <PhMagnifyingGlass :size="16" weight="bold" />
              Search
            </button>

            <button
              @click="showFilters = !showFilters"
              class="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
            >
              <PhFunnel :size="16" />
              Filters
              <span
                v-if="selectedCategory || signalValidFilter"
                class="w-2 h-2 bg-blue-600 rounded-full"
              ></span>
            </button>

            <button
              @click="clearFilters"
              class="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
            >
              <PhX :size="16" />
              Clear
            </button>
          </div>
        </div>

        <!-- Expanded Filters -->
        <div v-if="showFilters" class="mt-4 pt-4 border-t border-gray-200">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1"> Signal Category </label>
              <select
                v-model="selectedCategory"
                @change="handleFilterChange"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="">All Categories</option>
                <option value="STRONG_BUY">Strong Buy</option>
                <option value="BUY">Buy</option>
                <option value="WAIT">Wait</option>
                <option value="SELL">Sell</option>
                <option value="STRONG_SELL">Strong Sell</option>
              </select>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1"> Signal Validity </label>
              <select
                v-model="signalValidFilter"
                @change="handleFilterChange"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="">All</option>
                <option value="valid">Valid Only</option>
                <option value="invalid">Invalid Only</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      <!-- Signals Grid - Card Layout -->
      <div v-if="!isLoading && signals.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
        <div
          v-for="signal in signals"
          :key="signal.id"
          @click="viewDetail(signal.id)"
          class="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden hover:shadow-md hover:-translate-y-0.5 transition-all duration-200 cursor-pointer"
        >
          <!-- Card Header -->
          <div class="p-4 bg-gradient-to-r from-gray-50 to-gray-100 border-b border-gray-200">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="p-2 bg-white rounded-lg shadow-sm">
                  <PhCurrencyBtc :size="20" class="text-blue-600" weight="fill" />
                </div>
                <div>
                  <h4 class="text-lg font-bold text-gray-900">{{ signal.symbol }}</h4>
                  <p class="text-xs text-gray-500">{{ signal.primary_timeframe }}</p>
                </div>
              </div>
              <span
                :class="getCategoryBadgeClass(signal.signal_category)"
                class="inline-flex items-center gap-1 px-3 py-1.5 rounded-full text-xs font-medium border"
              >
                <span>{{ getCategoryIcon(signal.signal_category) }}</span>
                {{ signal.signal_category.replace('_', ' ') }}
              </span>
            </div>
          </div>

          <!-- Card Body -->
          <div class="p-4 bg-white">
            <!-- Score & Validity -->
            <div class="flex items-center justify-between mb-4">
              <div>
                <div class="text-xs text-gray-500 mb-1">Total Score</div>
                <div
                  :class="{
                    'text-green-600 font-bold': signal.total_score >= 50,
                    'text-red-600 font-bold': signal.total_score <= -50,
                    'text-gray-600 font-semibold': signal.total_score > -50 && signal.total_score < 50
                  }"
                  class="text-lg"
                >
                  {{ signal.total_score > 0 ? '+' : '' }}{{ signal.total_score.toFixed(2) }}
                </div>
              </div>
              <span
                :class="signal.signal_valid ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'"
                class="px-3 py-1.5 text-xs font-semibold rounded-full"
              >
                {{ signal.signal_valid ? '✓ Valid' : '✗ Invalid' }}
              </span>
            </div>

            <!-- Price Info -->
            <div class="grid grid-cols-3 gap-2 mb-4">
              <div class="text-center p-2 bg-blue-50 rounded-lg">
                <div class="text-xs text-gray-500 mb-1">Current</div>
                <div class="text-sm font-bold text-gray-900">${{ formatPrice(signal.current_price) }}</div>
              </div>
              <div class="text-center p-2 bg-green-50 rounded-lg">
                <div class="flex items-center justify-center gap-1 text-xs text-gray-500 mb-1">
                  <PhTarget :size="12" />
                  TP
                </div>
                <div class="text-sm font-bold text-green-700">${{ formatPrice(signal.tp_price) }}</div>
              </div>
              <div class="text-center p-2 bg-red-50 rounded-lg">
                <div class="flex items-center justify-center gap-1 text-xs text-gray-500 mb-1">
                  <PhStopCircle :size="12" />
                  SL
                </div>
                <div class="text-sm font-bold text-red-700">${{ formatPrice(signal.sl_price) }}</div>
              </div>
            </div>

            <!-- R:R Ratio -->
            <div class="mb-4">
              <div class="flex items-center justify-between">
                <span class="text-xs text-gray-500">Risk/Reward Ratio</span>
                <span class="text-sm font-bold text-gray-900">{{ signal.risk_reward_ratio.toFixed(2) }}</span>
              </div>
            </div>

            <!-- Created At -->
            <div class="flex items-center gap-2 text-xs text-gray-500">
              <PhClock :size="14" />
              <span>{{ formatDate(signal.created_at) }}</span>
            </div>
          </div>

          <!-- Card Footer -->
          <div class="px-4 py-3 bg-gray-50 border-t border-gray-200 flex items-center justify-between">
            <div class="flex items-center gap-2">
              <PhChartLineUp :size="16" class="text-gray-400" />
              <span class="text-xs text-gray-600">Click to view details</span>
            </div>
            <button
              @click.stop="handleDelete(signal.id, $event)"
              class="text-red-600 hover:text-red-900 p-2 hover:bg-red-50 rounded-lg transition-colors"
              title="Delete signal"
            >
              <PhTrash :size="16" />
            </button>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="!isLoading && signals.length === 0" class="bg-white border border-gray-200 rounded-xl shadow-sm p-12 mb-6">
        <div class="flex flex-col items-center">
          <PhMagnifyingGlass :size="48" class="text-gray-300 mb-4" />
          <p class="text-gray-500 text-lg font-medium">No signals found</p>
          <p class="text-gray-400 text-sm mt-1">Try adjusting your search or filters</p>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="isLoading" class="bg-white border border-gray-200 rounded-xl shadow-sm p-12 mb-6">
        <div class="flex flex-col items-center">
          <PhArrowCounterClockwise :size="48" class="text-blue-500 animate-spin mb-4" />
          <p class="text-gray-500 text-lg font-medium">Loading signals...</p>
        </div>
      </div>

      <!-- Pagination -->
      <div
        v-if="pagination.total_pages > 1"
        class="bg-white border border-gray-200 rounded-xl shadow-sm px-6 py-4 mb-6"
      >
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-700">
            Showing
            <span class="font-medium">{{
              (pagination.page - 1) * pagination.page_size + 1
            }}</span>
            to
            <span class="font-medium">{{
              Math.min(pagination.page * pagination.page_size, pagination.total_items)
            }}</span>
            of <span class="font-medium">{{ pagination.total_items }}</span> results
          </div>
          <div class="flex items-center gap-2">
            <button
              @click="handlePageChange(pagination.page - 1)"
              :disabled="pagination.page <= 1"
              class="px-3 py-1 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>

            <div class="flex items-center gap-1">
              <button
                v-for="page in [pagination.page - 1, pagination.page, pagination.page + 1].filter(
                  (p) => p >= 1 && p <= pagination.total_pages
                )"
                :key="page"
                @click="handlePageChange(page)"
                :class="{
                  'bg-blue-600 text-white border-blue-600': page === pagination.page,
                  'bg-white text-gray-700 border-gray-300 hover:bg-gray-50':
                    page !== pagination.page
                }"
                class="px-3 py-1 text-sm font-medium border rounded-lg transition-colors"
              >
                {{ page }}
              </button>
            </div>

            <button
              @click="handlePageChange(pagination.page + 1)"
              :disabled="pagination.page >= pagination.total_pages"
              class="px-3 py-1 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
  </DefaultLayout>
</template>

<style scoped>
</style>
