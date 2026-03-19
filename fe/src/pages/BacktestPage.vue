<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useBacktestStore } from '@/stores/backtest.store'
import { useStrategiesStore } from '@/stores/strategies.store'
import { DefaultLayout } from '@/layouts'
import {
  PhFlask,
  PhPlus,
  PhTrash,
  PhMagnifyingGlass,
  PhEye,
  PhClock,
  PhCheckCircle,
  PhXCircle,
  PhGear
} from '@phosphor-icons/vue'
import BacktestCreateModal from '@/components/features/backtest/BacktestCreateModal.vue'

const router = useRouter()
const store = useBacktestStore()
const strategiesStore = useStrategiesStore()

const searchQuery = ref('')
const showCreateModal = ref(false)

const filteredBacktests = computed(() => {
  if (!searchQuery.value) return store.backtests

  return store.backtests.filter(backtest =>
    backtest.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    backtest.symbol.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    backtest.strategy_name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const getStatusColor = (status: string) => {
  switch (status) {
    case 'COMPLETED':
      return 'bg-green-100 text-green-700'
    case 'RUNNING':
      return 'bg-blue-100 text-blue-700'
    case 'PENDING':
      return 'bg-yellow-100 text-yellow-700'
    case 'FAILED':
      return 'bg-red-100 text-red-700'
    default:
      return 'bg-gray-100 text-gray-700'
  }
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'COMPLETED':
      return PhCheckCircle
    case 'RUNNING':
      return PhGear
    case 'PENDING':
      return PhClock
    case 'FAILED':
      return PhXCircle
    default:
      return PhClock
  }
}

const handleCreate = () => {
  store.resetForm()
  showCreateModal.value = true
  strategiesStore.fetchStrategies()
}

const handleViewDetail = (id: number) => {
  router.push(`/backtest/${id}`)
}

const handleDelete = async (id: number, name: string) => {
  await store.deleteBacktest(id, name)
}

const handleCloseCreateModal = () => {
  showCreateModal.value = false
  store.resetForm()
}

const isPageLoading = computed(() => store.loading)

onMounted(() => {
  store.fetchBacktests()
  strategiesStore.fetchStrategies()
})
</script>

<template>
  <DefaultLayout>
    <template #header-title>Backtest</template>

    <div class="">
      <!-- Loading State -->
      <div v-if="isPageLoading" class="flex items-center justify-center py-20">
        <div class="relative">
          <div class="animate-spin rounded-full h-16 w-16 border-b-2 border-primary"></div>
          <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2">
            <PhFlask :size="24" class="text-primary" />
          </div>
        </div>
      </div>

      <template v-else>
        <!-- Header Actions -->
        <div class="flex items-center justify-between mb-6">
          <div class="flex-1 max-w-md">
            <div class="relative">
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Search backtests..."
                class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              />
              <PhMagnifyingGlass
                :size="20"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
              />
            </div>
          </div>

          <button
            @click="handleCreate"
            class="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-all"
          >
            <PhPlus :size="20" weight="bold" />
            New Backtest
          </button>
        </div>

        <!-- Stats Cards -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
            <p class="text-sm text-gray-500 mb-1">Total Backtests</p>
            <p class="text-2xl font-bold text-gray-900">{{ store.backtests.length }}</p>
          </div>
          <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
            <p class="text-sm text-gray-500 mb-1">Completed</p>
            <p class="text-2xl font-bold text-green-600">
              {{ store.backtests.filter(b => b.status === 'COMPLETED').length }}
            </p>
          </div>
          <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
            <p class="text-sm text-gray-500 mb-1">Running</p>
            <p class="text-2xl font-bold text-blue-600">
              {{ store.backtests.filter(b => b.status === 'RUNNING').length }}
            </p>
          </div>
          <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
            <p class="text-sm text-gray-500 mb-1">Failed</p>
            <p class="text-2xl font-bold text-red-600">
              {{ store.backtests.filter(b => b.status === 'FAILED').length }}
            </p>
          </div>
        </div>

        <!-- Backtests Grid -->
        <div v-if="filteredBacktests.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div
            v-for="backtest in filteredBacktests"
            :key="backtest.id"
            class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6 hover:shadow-xl transition-all duration-200"
            :class="{ 'ring-2 ring-primary': backtest.status === 'COMPLETED' }"
          >
            <!-- Header -->
            <div class="flex items-start justify-between mb-4">
              <div class="flex-1">
                <div class="flex items-center gap-2 mb-2">
                  <h3 class="text-lg font-bold text-gray-900 truncate">{{ backtest.name }}</h3>
                </div>
                <p class="text-sm text-gray-500">{{ backtest.symbol }}</p>
              </div>
              <div class="flex items-center gap-1">
                <component
                  :is="getStatusIcon(backtest.status)"
                  :size="16"
                  :weight="backtest.status === 'COMPLETED' ? 'fill' : 'regular'"
                  :class="getStatusColor(backtest.status).split(' ')[1]"
                />
                <span
                  class="px-2 py-1 text-xs font-medium rounded-full"
                  :class="getStatusColor(backtest.status)"
                >
                  {{ backtest.status }}
                </span>
              </div>
            </div>

            <!-- Strategy -->
            <div class="mb-4">
              <p class="text-xs text-gray-500 mb-1">Strategy</p>
              <p class="text-sm font-medium text-gray-900 truncate">{{ backtest.strategy_name }}</p>
            </div>

            <!-- Stats -->
            <div class="grid grid-cols-2 gap-3 mb-4 text-sm">
              <div>
                <p class="text-gray-500 text-xs">Net PnL</p>
                <p
                  class="font-bold text-base"
                  :class="backtest.total_pnl >= 0 ? 'text-green-600' : 'text-red-600'"
                >
                  {{ backtest.total_pnl >= 0 ? '+' : '' }}{{ backtest.total_pnl.toFixed(2) }}
                  <span class="text-xs">USDT</span>
                </p>
                <p
                  class="text-xs font-medium"
                  :class="backtest.total_pnl_percent >= 0 ? 'text-green-600' : 'text-red-600'"
                >
                  ({{ backtest.total_pnl_percent >= 0 ? '+' : '' }}{{ backtest.total_pnl_percent.toFixed(2) }}%)
                </p>
              </div>
              <div>
                <p class="text-gray-500 text-xs">Win Rate</p>
                <p class="font-bold text-base text-gray-900">{{ backtest.win_rate.toFixed(1) }}%</p>
                <p class="text-xs text-gray-500">{{ backtest.total_trades }} trades</p>
              </div>
            </div>

            <!-- Additional Info -->
            <div class="pt-3 border-t border-gray-100 mb-4">
              <div class="flex items-center justify-between text-xs text-gray-500">
                <span>Created: {{ new Date(backtest.created_at).toLocaleDateString() }}</span>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex items-center gap-2 pt-4 border-t border-gray-100">
              <button
                @click="handleViewDetail(backtest.id)"
                class="flex-1 flex items-center justify-center gap-1 px-3 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-all text-sm font-medium"
                :disabled="backtest.status !== 'COMPLETED'"
                :class="backtest.status !== 'COMPLETED' ? 'opacity-50 cursor-not-allowed' : ''"
              >
                <PhEye :size="16" weight="bold" />
                View Detail
              </button>

              <button
                @click="handleDelete(backtest.id, backtest.name)"
                class="flex items-center justify-center gap-1 px-3 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-all text-sm font-medium"
                aria-label="Delete backtest"
              >
                <PhTrash :size="16" />
              </button>
            </div>
          </div>
        </div>

        <!-- Empty State -->
        <div v-else class="text-center py-20">
          <PhFlask :size="64" class="mx-auto text-gray-300 mb-4" />
          <p class="text-gray-500 text-lg mb-2">
            {{ searchQuery ? 'No backtests found' : 'No backtests yet' }}
          </p>
          <p v-if="!searchQuery" class="text-gray-400 text-sm mb-4">
            Create your first backtest to test your trading strategy
          </p>
          <button
            v-if="!searchQuery"
            @click="handleCreate"
            class="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-all"
          >
            <PhPlus :size="20" weight="bold" />
            Create Backtest
          </button>
        </div>
      </template>
    </div>

    <!-- Create Backtest Modal -->
    <BacktestCreateModal
      v-model:show="showCreateModal"
      @submit="store.fetchBacktests"
      @close="handleCloseCreateModal"
    />
  </DefaultLayout>
</template>
