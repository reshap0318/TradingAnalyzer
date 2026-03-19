<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useBacktestStore } from '@/stores/backtest.store'
import { DefaultLayout } from '@/layouts'
import { PhArrowLeft, PhChartLineUp, PhFlask, PhList, PhSquaresFour, PhGear } from '@phosphor-icons/vue'
import BacktestSummaryTab from '@/components/features/backtest/BacktestSummaryTab.vue'
import BacktestStrategyTab from '@/components/features/backtest/BacktestStrategyTab.vue'
import BacktestTradesTab from '@/components/features/backtest/BacktestTradesTab.vue'
import BacktestChartTab from '@/components/features/backtest/BacktestChartTab.vue'

const route = useRoute()
const router = useRouter()
const store = useBacktestStore()

const activeTab = ref<'summary' | 'strategy' | 'trades' | 'chart'>('summary')
const backtestId = computed(() => Number(route.params.id))

const backtest = computed(() => store.currentBacktest)
const isLoading = computed(() => store.loading)

const backtestName = computed(() => {
  return backtest.value?.name || 'Backtest Detail'
})

// Fetch backtest detail when component mounts or route changes
const fetchBacktestDetail = () => {
  store.fetchBacktestDetail(backtestId.value)
  
  // Start polling if status is PENDING or RUNNING
  const status = backtest.value?.status
  if (status === 'PENDING' || status === 'RUNNING') {
    store.startPolling(backtestId.value)
  }
}

const handleBack = () => {
  router.push('/backtest')
}

const handleViewTrade = (trade: any) => {
  console.log('View trade:', trade)
}

// Watch for route changes
watch(
  () => route.params.id,
  () => {
    fetchBacktestDetail()
  }
)

onMounted(() => {
  fetchBacktestDetail()
})
</script>

<template>
  <DefaultLayout>
    <template #header-title>
      <div class="flex items-center gap-3">
        <button
          @click="handleBack"
          class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          aria-label="Go back"
        >
          <PhArrowLeft :size="20" />
        </button>
        <span>{{ backtestName }}</span>
      </div>
    </template>

    <div class="pb-8">
      <!-- Loading State -->
      <div v-if="isLoading && !backtest" class="flex items-center justify-center py-20">
        <div class="relative">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
        </div>
      </div>

      <template v-else-if="backtest">
        <!-- Header & Status -->
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-4 mb-4">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-lg font-bold text-gray-900">{{ backtest.name }}</h2>
              <p class="text-sm text-gray-500">
                {{ backtest.symbol }} • {{ backtest.strategy?.strategy_name || 'Strategy' }}
              </p>
            </div>
            <span
              v-if="backtest.status"
              class="px-3 py-1 text-xs font-medium rounded-full"
              :class="{
                'bg-green-100 text-green-700': backtest.status === 'COMPLETED',
                'bg-blue-100 text-blue-700': backtest.status === 'RUNNING',
                'bg-yellow-100 text-yellow-700': backtest.status === 'PENDING',
                'bg-red-100 text-red-700': backtest.status === 'FAILED'
              }"
            >
              {{ backtest.status }}
            </span>
          </div>
        </div>

        <!-- Tabs -->
        <div class="bg-white border border-gray-200 rounded-xl shadow-sm mb-4 overflow-hidden">
          <div class="grid grid-cols-4 border-b border-gray-200">
            <button
              @click="activeTab = 'summary'"
              class="py-3 flex flex-col items-center gap-1 transition-all border-b-2"
              :class="
                activeTab === 'summary'
                  ? 'border-primary bg-blue-50/50'
                  : 'border-transparent hover:bg-gray-50'
              "
            >
              <PhSquaresFour :size="20" :weight="activeTab === 'summary' ? 'fill' : 'regular'" 
                :class="activeTab === 'summary' ? 'text-primary' : 'text-gray-400'" />
              <span class="text-xs font-medium" :class="activeTab === 'summary' ? 'text-primary' : 'text-gray-500'">
                Summary
              </span>
            </button>
            <button
              @click="activeTab = 'strategy'"
              class="py-3 flex flex-col items-center gap-1 transition-all border-b-2"
              :class="
                activeTab === 'strategy'
                  ? 'border-primary bg-purple-50/50'
                  : 'border-transparent hover:bg-gray-50'
              "
            >
              <PhGear :size="20" :weight="activeTab === 'strategy' ? 'fill' : 'regular'" 
                :class="activeTab === 'strategy' ? 'text-purple-600' : 'text-gray-400'" />
              <span class="text-xs font-medium" :class="activeTab === 'strategy' ? 'text-purple-600' : 'text-gray-500'">
                Strategy
              </span>
            </button>
            <button
              @click="activeTab = 'trades'"
              class="py-3 flex flex-col items-center gap-1 transition-all border-b-2"
              :class="
                activeTab === 'trades'
                  ? 'border-primary bg-green-50/50'
                  : 'border-transparent hover:bg-gray-50'
              "
            >
              <PhList :size="20" :weight="activeTab === 'trades' ? 'fill' : 'regular'" 
                :class="activeTab === 'trades' ? 'text-green-600' : 'text-gray-400'" />
              <span class="text-xs font-medium" :class="activeTab === 'trades' ? 'text-green-600' : 'text-gray-500'">
                Trades
              </span>
            </button>
            <button
              @click="activeTab = 'chart'"
              class="py-3 flex flex-col items-center gap-1 transition-all border-b-2"
              :class="
                activeTab === 'chart'
                  ? 'border-primary bg-orange-50/50'
                  : 'border-transparent hover:bg-gray-50'
              "
            >
              <PhChartLineUp :size="20" :weight="activeTab === 'chart' ? 'fill' : 'regular'" 
                :class="activeTab === 'chart' ? 'text-orange-600' : 'text-gray-400'" />
              <span class="text-xs font-medium" :class="activeTab === 'chart' ? 'text-orange-600' : 'text-gray-500'">
                Chart
              </span>
            </button>
          </div>

          <!-- Tab Content -->
          <div class="p-4">
            <BacktestSummaryTab v-if="activeTab === 'summary'" :backtest="backtest" />
            <BacktestStrategyTab v-else-if="activeTab === 'strategy'" :backtest="backtest" />
            <BacktestTradesTab v-else-if="activeTab === 'trades'" :backtest="backtest" @view-trade="handleViewTrade" />
            <BacktestChartTab v-else-if="activeTab === 'chart'" :backtest="backtest" />
          </div>
        </div>
      </template>

      <!-- Empty State -->
      <div v-else class="text-center py-20">
        <PhFlask :size="64" class="mx-auto text-gray-300 mb-4" />
        <p class="text-gray-500 text-lg">Backtest not found</p>
      </div>
    </div>
  </DefaultLayout>
</template>
