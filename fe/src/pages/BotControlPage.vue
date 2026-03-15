<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useTradeBotStore } from '@/stores/tradebot.store'
import { DefaultLayout } from '@/layouts'
import {
  PhRobot,
  PhPower,
  PhStop,
  PhArrowCounterClockwise,
  PhChartLineUp,
  PhGearSix,
  PhFlask
} from '@phosphor-icons/vue'
import Swal from 'sweetalert2'
import { ActiveTradeCard, TradeExecutedCard } from '@/components/features/bot-control'

const store = useTradeBotStore()

const isActive = computed(() => store.botStatus?.is_active ?? false)
const strategy = computed(() => store.strategy)
const botRunningDuration = computed(() => store.botStatus?.bot_running_seconds || 0)

// Format bot running duration to human readable
const formattedBotDuration = computed(() => {
  const seconds = botRunningDuration.value
  if (seconds < 60) return `${Math.floor(seconds)}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${Math.floor(seconds % 60)}s`
  if (seconds < 86400) {
    // Less than 1 day
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    return `${hours}h ${minutes}m`
  }
  // More than 1 day
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  return `${days}d ${hours}h ${minutes}m`
})

// Session summary state
const sessionSummary = computed(() => store.sessionSummary)
const activeTrades = computed(() => store.activeTrades)
const summaryLoading = computed(() => store.summaryLoading)

// Calculate total indicator weight
const totalIndicatorWeight = computed(() => {
  if (!strategy.value?.indicator_weights || strategy.value.indicator_weights.length === 0) {
    return 0
  }
  return strategy.value.indicator_weights.reduce((sum, ind) => sum + ind.weight, 0) * 100
})

// Calculate total PnL from session
const totalPnL = computed(() => {
  if (!sessionSummary.value) return 0
  return sessionSummary.value.total_pnl || 0
})

// Background reload for session data (every 3 minutes when bot is active)
const RELOAD_INTERVAL_MS = 3 * 60 * 1000 // 3 minutes in milliseconds
let reloadIntervalId: ReturnType<typeof setInterval> | null = null

const startBackgroundReload = () => {
  // Stop existing interval if any
  stopBackgroundReload()

  // Fetch immediately first
  store.fetchSessionData()

  // Then set interval for subsequent reloads
  reloadIntervalId = setInterval(() => {
    store.fetchSessionData()
  }, RELOAD_INTERVAL_MS)
}

const stopBackgroundReload = () => {
  if (reloadIntervalId) {
    clearInterval(reloadIntervalId)
    reloadIntervalId = null
  }
}

// Watch bot active state to start/stop background reload
watch(() => isActive.value, async (newActive) => {
  if (newActive) {
    startBackgroundReload()
  } else {
    stopBackgroundReload()
    store.clearSessionData()
  }
}, { immediate: true })

// Strategy selection
const showStrategyModal = ref(false)
const selectedStrategyId = ref<number | null>(null)

const handleToggle = async () => {
  if (isActive.value && !strategy.value) {
    const result = await Swal.fire({
      title: 'No Strategy Selected',
      text: 'Bot will run without a specific strategy. Are you sure?',
      icon: 'warning',
      showCancelButton: true,
      confirmButtonText: 'Continue',
      cancelButtonText: 'Cancel'
    })

    if (!result.isConfirmed) return
  }

  await store.toggleBot()
}

const handleRefresh = async () => {
  await store.fetchBotStatus()
  // Also refresh session data if bot is active
  if (isActive.value) {
    await store.fetchSessionData()
  }
}

const openStrategySelector = () => {
  store.fetchStrategies()
  selectedStrategyId.value = strategy.value?.id || null
  showStrategyModal.value = true
}

const handleStrategySelect = async () => {
  const success = await store.selectStrategy(selectedStrategyId.value)
  if (success) {
    showStrategyModal.value = false
  }
}

onMounted(() => {
  store.fetchBotStatus()
})

onUnmounted(() => {
  stopBackgroundReload()
})
</script>

<template>
  <DefaultLayout>
    <template #header-title>Trading Bot Control</template>

    <div class="mx-auto sm:px-6">
      <!-- Loading State -->
      <div v-if="store.loading" class="flex items-center justify-center py-20">
        <div class="relative">
          <div class="animate-spin rounded-full h-16 w-16 border-b-2 border-primary"></div>
          <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2">
            <PhRobot :size="24" class="text-primary" />
          </div>
        </div>
      </div>

      <template v-else>
        <!-- Top Section: Bot Status & Strategy -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          <!-- Bot Status Card -->
          <div class="lg:col-span-2 bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 rounded-3xl shadow-2xl p-8 relative overflow-hidden">
            <!-- Background Pattern -->
            <div class="absolute inset-0 opacity-10">
              <div class="absolute top-10 left-10 w-32 h-32 bg-blue-500 rounded-full blur-3xl"></div>
              <div class="absolute bottom-10 right-10 w-40 h-40 bg-purple-500 rounded-full blur-3xl"></div>
            </div>

            <div class="relative z-10">
              <!-- Header -->
              <div class="flex items-center justify-between mb-6">
                <div class="flex items-center gap-4">
                  <div class="p-3 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl shadow-lg">
                    <PhRobot :size="40" class="text-white" weight="fill" />
                  </div>
                  <div>
                    <h2 class="text-2xl font-bold text-white">Trading Bot</h2>
                    <p class="text-gray-400 text-sm">Automated Trading System</p>
                  </div>
                </div>

                <button
                  @click="handleRefresh"
                  class="p-2 bg-gray-800 hover:bg-gray-700 rounded-xl transition-all duration-200 hover:scale-110"
                  :disabled="store.toggling"
                >
                  <PhArrowCounterClockwise :size="20" class="text-gray-300" :class="{ 'animate-spin': store.loading }" />
                </button>
              </div>

              <!-- Status -->
              <div class="flex items-center gap-3 mb-6">
                <div
                  class="w-3 h-3 rounded-full transition-all duration-500"
                  :class="isActive ? 'bg-green-500 animate-pulse' : 'bg-gray-500'"
                ></div>
                <span
                  class="text-base font-semibold transition-all duration-300"
                  :class="isActive ? 'text-green-400' : 'text-gray-400'"
                >
                  {{ isActive ? 'ACTIVE' : 'INACTIVE' }}
                </span>
              </div>

              <!-- Toggle Button -->
              <div class="flex items-center justify-center mb-6">
                <button
                  @click="handleToggle"
                  :disabled="store.toggling"
                  class="relative inline-flex items-center justify-center w-28 h-28 rounded-full transition-all duration-500"
                  :class="[
                    isActive
                      ? 'bg-gradient-to-br from-green-500 to-emerald-600 hover:from-green-400 hover:to-emerald-500 shadow-lg shadow-green-500/50'
                      : 'bg-gradient-to-br from-gray-600 to-gray-700 hover:from-gray-500 hover:to-gray-600 shadow-lg shadow-gray-500/30'
                  ]"
                >
                  <div v-if="store.toggling" class="animate-spin rounded-full h-10 w-10 border-4 border-white border-t-transparent"></div>
                  <PhStop v-else-if="isActive" :size="40" class="text-white" weight="fill" />
                  <PhPower v-else :size="40" class="text-white" weight="fill" />
                </button>
              </div>

              <!-- Status Label -->
              <p class="text-center text-gray-400 text-sm">
                {{ isActive ? 'Bot is Running' : 'Bot is Stopped' }}
              </p>
            </div>
          </div>

          <!-- Strategy Card -->
          <div class="bg-white rounded-3xl shadow-lg border border-gray-100 p-6">
            <div class="flex items-center justify-between mb-6">
              <div class="flex items-center gap-3">
                <div class="p-3 bg-blue-50 rounded-xl">
                  <PhChartLineUp :size="28" class="text-blue-600" weight="fill" />
                </div>
                <div>
                  <h3 class="text-lg font-bold text-gray-900">Strategy</h3>
                  <p class="text-xs text-gray-500">Active Configuration</p>
                </div>
              </div>
              <button
                @click="openStrategySelector"
                class="p-2 hover:bg-gray-100 rounded-lg transition-all"
                title="Change Strategy"
              >
                <PhGearSix :size="20" class="text-gray-600" />
              </button>
            </div>

            <div v-if="strategy" class="space-y-4">
              <!-- Strategy Name -->
              <div class="bg-gradient-to-r from-blue-50 to-purple-50 rounded-xl p-4">
                <p class="text-sm text-gray-600 mb-1">Current Strategy</p>
                <p class="text-lg font-bold text-gray-900">{{ strategy.strategy_name }}</p>
              </div>

              <!-- Primary Timeframe -->
              <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <span class="text-sm text-gray-600">Primary TF</span>
                <span class="px-3 py-1 bg-blue-100 text-blue-700 text-sm font-semibold rounded-full">{{ strategy.primary_tf }}</span>
              </div>

              <!-- Money Management -->
              <div v-if="strategy.money_management" class="space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-sm text-gray-600">Min Confidence</span>
                  <span class="text-sm font-bold text-gray-900">{{ strategy.money_management.min_confidence ?? '-' }}%</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-sm text-gray-600">Leverage</span>
                  <span class="text-sm font-bold text-gray-900">{{ strategy.money_management.leverage ?? '-' }}x</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-sm text-gray-600">Max Position</span>
                  <span class="text-sm font-bold text-gray-900">{{ ((strategy.money_management.max_position_size ?? 0) * 100).toFixed(0) }}%</span>
                </div>
              </div>
            </div>

            <div v-else class="text-center py-8">
              <PhChartLineUp :size="40" class="mx-auto text-gray-300 mb-2" />
              <p class="text-gray-500 text-sm mb-3">No strategy selected</p>
              <button
                @click="openStrategySelector"
                class="px-4 py-2 bg-blue-500 text-white text-sm font-medium rounded-lg hover:bg-blue-600 transition-all"
              >
                Select Strategy
              </button>
            </div>
          </div>
        </div>

        <!-- Stats Overview -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <!-- Trades Executed -->
          <TradeExecutedCard :trades="store.executedTrades" />

          <!-- Active Trades -->
          <ActiveTradeCard :trades="activeTrades" />

          <!-- Total PnL -->
          <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
            <div class="flex items-center gap-3 mb-4">
              <div class="p-3" :class="totalPnL >= 0 ? 'bg-green-50' : 'bg-red-50'">
                <PhChartLineUp :size="24" class="text-gray-600" :class="totalPnL >= 0 ? 'text-green-600' : 'text-red-600'" weight="fill" />
              </div>
              <span class="text-sm text-gray-600">Total PnL</span>
            </div>
            <div v-if="summaryLoading" class="animate-spin rounded-full h-8 w-8 border-4 border-gray-200 border-t-gray-600"></div>
            <p v-else class="text-2xl font-bold" :class="totalPnL >= 0 ? 'text-green-600' : 'text-red-600'">
              {{ totalPnL >= 0 ? '+' : '' }}{{ totalPnL.toFixed(2) }} USDT
            </p>
            <p v-if="sessionSummary" class="text-xs text-gray-500 mt-1">
              Success Rate: {{ sessionSummary.success_rate?.toFixed(1) ?? 0 }}%
            </p>
          </div>

          <!-- Bot Running Duration -->
          <div class="bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 rounded-2xl shadow-lg border border-gray-700 p-6">
            <div class="flex items-center gap-3 mb-4">
              <div class="p-3 bg-gray-700 rounded-xl">
                <PhRobot :size="24" :class="isActive ? 'text-green-400' : 'text-gray-400'" weight="fill" />
              </div>
              <span class="text-sm text-gray-400">Running Duration</span>
            </div>
            <div v-if="!isActive" class="text-2xl font-bold text-gray-500">-</div>
            <p v-else class="text-2xl font-bold text-green-400">{{ formattedBotDuration }}</p>
            <p v-if="isActive && store.botStatus?.bot_started_at" class="text-xs text-gray-500 mt-1">
              Since: {{ new Date(store.botStatus.bot_started_at).toLocaleTimeString() }}
            </p>
          </div>
        </div>

        <!-- Strategy Details -->
        <div v-if="strategy" class="bg-white rounded-3xl shadow-lg border border-gray-100 p-6 mb-8">
          <div class="flex items-center gap-3 mb-6">
            <div class="p-3 bg-purple-50 rounded-xl">
              <PhFlask :size="28" class="text-purple-600" weight="fill" />
            </div>
            <div>
              <h3 class="text-xl font-bold text-gray-900">Strategy Details</h3>
              <p class="text-sm text-gray-500">Complete configuration breakdown</p>
            </div>
          </div>

          <!-- Timeframes -->
          <div v-if="strategy.timeframes && strategy.timeframes.length > 0" class="mb-6">
            <h4 class="text-sm font-semibold text-gray-700 mb-3">Timeframes Analysis</h4>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <div
                v-for="tf in strategy.timeframes"
                :key="tf.tf"
                class="p-4 bg-gradient-to-r from-gray-50 to-gray-100 rounded-xl border border-gray-200"
              >
                <div class="flex items-center justify-between mb-2">
                  <span class="text-lg font-bold text-gray-900">{{ tf.tf }}</span>
                  <span class="text-xs font-medium text-gray-600 bg-white px-2 py-1 rounded">{{ (tf.weight * 100).toFixed(0) }}% weight</span>
                </div>
                <div class="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
                  <div
                    class="h-full bg-gradient-to-r from-blue-500 to-purple-500 rounded-full"
                    :style="{ width: `${tf.weight * 100}%` }"
                  ></div>
                </div>
                <div v-if="tf.timeframe_detail" class="mt-2 text-xs text-gray-500">
                  {{ tf.timeframe_detail.in_minutes }} minutes
                </div>
              </div>
            </div>
          </div>

          <!-- Indicators -->
          <div v-if="strategy.indicator_weights && strategy.indicator_weights.length > 0" class="mb-6">
            <div class="flex items-center justify-between mb-3">
              <h4 class="text-sm font-semibold text-gray-700">Indicator Weights</h4>
              <span class="text-xs font-medium text-blue-600 bg-blue-50 px-2 py-1 rounded">
                Total: {{ totalIndicatorWeight.toFixed(1) }}%
              </span>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <div
                v-for="ind in strategy.indicator_weights"
                :key="ind.indicator_id"
                class="p-4 bg-gray-50 rounded-xl border border-gray-200 flex items-center justify-between"
              >
                <span class="text-sm font-medium text-gray-900">{{ ind.indicator_detail?.name || 'Unknown' }}</span>
                <span class="text-xs font-bold text-blue-600 bg-blue-50 px-2 py-1 rounded">{{ (ind.weight * 100).toFixed(1) }}%</span>
              </div>
            </div>
          </div>

          <!-- Full Money Management -->
          <div v-if="strategy.money_management">
            <h4 class="text-sm font-semibold text-gray-700 mb-3">Money Management</h4>
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div class="p-4 bg-gray-50 rounded-xl">
                <p class="text-xs text-gray-500 mb-1">Min Confidence</p>
                <p class="text-lg font-bold text-gray-900">{{ strategy.money_management.min_confidence ?? '-' }}%</p>
              </div>
              <div class="p-4 bg-gray-50 rounded-xl">
                <p class="text-xs text-gray-500 mb-1">Max Daily Trades</p>
                <p class="text-lg font-bold text-gray-900">{{ strategy.money_management.max_daily_trades ?? '-' }}</p>
              </div>
              <div class="p-4 bg-gray-50 rounded-xl">
                <p class="text-xs text-gray-500 mb-1">Max Daily Loss</p>
                <p class="text-lg font-bold text-gray-900">{{ strategy.money_management.max_daily_loss_percent ?? '-' }}%</p>
              </div>
              <div class="p-4 bg-gray-50 rounded-xl">
                <p class="text-xs text-gray-500 mb-1">Risk/Reward</p>
                <p class="text-lg font-bold text-gray-900">{{ strategy.money_management.risk_reward_ratio ?? '-' }}</p>
              </div>
              <div class="p-4 bg-gray-50 rounded-xl">
                <p class="text-xs text-gray-500 mb-1">Max Position</p>
                <p class="text-lg font-bold text-gray-900">{{ ((strategy.money_management.max_position_size ?? 0) * 100).toFixed(0) }}%</p>
              </div>
              <div class="p-4 bg-gray-50 rounded-xl">
                <p class="text-xs text-gray-500 mb-1">Leverage</p>
                <p class="text-lg font-bold text-gray-900">{{ strategy.money_management.leverage ?? '-' }}x</p>
              </div>
              <div class="p-4 bg-gray-50 rounded-xl">
                <p class="text-xs text-gray-500 mb-1">Mode</p>
                <p class="text-lg font-bold" :class="strategy.money_management.is_agressive ? 'text-orange-600' : 'text-blue-600'">
                  {{ strategy.money_management.is_agressive ? 'Aggressive' : 'Conservative' }}
                </p>
              </div>
              <div class="p-4 bg-gray-50 rounded-xl">
                <p class="text-xs text-gray-500 mb-1">Order Expiry</p>
                <p class="text-lg font-bold text-gray-900">{{ strategy.money_management.order_expiration_hours ?? '-' }}h</p>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Strategy Selection Modal -->
    <div v-if="showStrategyModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div class="bg-white rounded-2xl shadow-2xl max-w-2xl w-full max-h-[80vh] overflow-hidden flex flex-col">
        <!-- Header -->
        <div class="p-6 border-b border-gray-200">
          <h3 class="text-xl font-bold text-gray-900">Select Strategy</h3>
          <p class="text-sm text-gray-500 mt-1">Choose which strategy to use for automated trading</p>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-y-auto p-6">
          <div class="space-y-3">
            <!-- No Strategy Option -->
            <label class="flex items-center gap-4 p-4 border-2 rounded-xl cursor-pointer transition-all hover:bg-gray-50"
                   :class="selectedStrategyId === null ? 'border-blue-500 bg-blue-50' : 'border-gray-200'">
              <input
                v-model="selectedStrategyId"
                type="radio"
                :value="null"
                class="w-4 h-4 text-blue-600"
              />
              <div class="flex-1">
                <p class="font-semibold text-gray-900">No Specific Strategy</p>
                <p class="text-sm text-gray-500">Bot will use default configuration</p>
              </div>
            </label>

            <!-- Strategy Options -->
            <label
              v-for="s in store.strategies"
              :key="s.id"
              class="flex items-center gap-4 p-4 border-2 rounded-xl cursor-pointer transition-all hover:bg-gray-50"
              :class="selectedStrategyId === s.id ? 'border-blue-500 bg-blue-50' : 'border-gray-200'"
            >
              <input
                v-model="selectedStrategyId"
                type="radio"
                :value="s.id"
                class="w-4 h-4 text-blue-600"
              />
              <div class="flex-1">
                <div class="flex items-center justify-between">
                  <p class="font-semibold text-gray-900">{{ s.strategy_name }}</p>
                  <span v-if="s.is_active" class="px-2 py-1 bg-blue-100 text-blue-700 text-xs font-medium rounded">Default</span>
                </div>
                <div class="flex items-center gap-3 mt-1 text-sm text-gray-500">
                  <span>Primary: {{ s.primary_tf }}</span>
                  <span v-if="s.timeframes && s.timeframes.length > 0">• {{ s.timeframes.length }} TFs</span>
                  <span v-if="s.money_management">• {{ s.money_management.leverage ?? '-' }}x leverage</span>
                </div>
              </div>
            </label>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-6 border-t border-gray-200 flex justify-end gap-3">
          <button
            @click="showStrategyModal = false"
            class="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-all"
          >
            Cancel
          </button>
          <button
            @click="handleStrategySelect"
            :disabled="store.loading"
            class="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-all disabled:opacity-50"
          >
            {{ store.loading ? 'Applying...' : 'Apply Strategy' }}
          </button>
        </div>
      </div>
    </div>
  </DefaultLayout>
</template>

<style scoped>
</style>
