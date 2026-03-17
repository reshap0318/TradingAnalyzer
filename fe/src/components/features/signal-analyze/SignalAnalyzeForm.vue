<script setup lang="ts">
import { computed } from 'vue'
import { useSignalAnalyzeStore } from '@/stores/signal-analyze.store'
import { useStrategiesStore } from '@/stores/strategies.store'
import { getValidationErrors } from '@/helpers/validation'
import { UiInput, UiButton } from '@/components/common'
import { PhCurrencyBtc } from '@phosphor-icons/vue'

const signalStore = useSignalAnalyzeStore()
const strategiesStore = useStrategiesStore()

// Access form state dan validation dari store
const analyzeReq = signalStore.analyzeReq
const v$ = signalStore.analyzeReqValid
const isLoading = computed(() => signalStore.loading)

// Get strategies untuk dropdown
const strategies = computed(() => strategiesStore.strategies)

// Submit handler
const handleSubmit = async () => {
  // Call action dari store (validasi sudah ada di dalam analyzeSignal)
  await signalStore.analyzeSignal()
}

// Handle Enter key
const handleKeyPress = (e: KeyboardEvent) => {
  if (e.key === 'Enter') {
    handleSubmit()
  }
}
</script>

<template>
  <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
    <div class="flex items-center gap-3 mb-6">
      <div class="p-3 bg-blue-50 rounded-xl">
        <PhCurrencyBtc :size="28" class="text-blue-600" weight="fill" />
      </div>
      <div>
        <h2 class="text-xl font-bold text-gray-900">Signal Analyze</h2>
        <p class="text-sm text-gray-500">Analyze trading signal based on your strategy</p>
      </div>
    </div>

    <form @submit.prevent="handleSubmit" @keypress="handleKeyPress">
      <!-- Symbol Input -->
      <UiInput
        v-model="analyzeReq.symbol"
        label="Symbol"
        placeholder="e.g., BTCUSDT"
        autocomplete="off"
        :error="v$.symbol.$error"
        :error-message="getValidationErrors(v$.symbol).join(', ')"
      />

      <!-- Strategy Select -->
      <div class="mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          Strategy <span class="text-red-500">*</span>
        </label>
        <select
          v-model="analyzeReq.strategy_id"
          class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
          :class="{ 'border-red-500': v$.strategy_id.$error }"
          :disabled="isLoading"
        >
          <option value="">Select a strategy...</option>
          <option
            v-for="strategy in strategies"
            :key="strategy.id"
            :value="strategy.id"
          >
            {{ strategy.strategy_name }} ({{ strategy.primary_tf }})
            {{ !strategy.is_active ? ' - Inactive' : '' }}
          </option>
        </select>
        <p
          v-if="v$.strategy_id.$error"
          class="mt-1 text-sm text-red-600"
        >
          {{ getValidationErrors(v$.strategy_id).join(', ') }}
        </p>
      </div>

      <!-- Capital Input -->
      <div class="mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          Capital (USD) <span class="text-red-500">*</span>
        </label>
        <input
          v-model.number="analyzeReq.capital"
          type="number"
          placeholder="50"
          class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
          :class="{ 'border-red-500': v$.capital.$error }"
          :disabled="isLoading"
        />
        <p
          v-if="v$.capital.$error"
          class="mt-1 text-sm text-red-600"
        >
          {{ getValidationErrors(v$.capital).join(', ') }}
        </p>
      </div>

      <!-- Submit Button -->
      <UiButton
        type="submit"
        variant="primary"
        :loading="isLoading"
        full-width
        class="mt-6"
      >
        {{ isLoading ? 'Analyzing...' : 'Analyze Signal' }}
      </UiButton>

      <!-- Info Text -->
      <p class="mt-4 text-xs text-gray-500 text-center">
        <PhCurrencyBtc :size="14" class="inline mr-1" />
        Analysis will use the primary timeframe from selected strategy
      </p>
    </form>
  </div>
</template>
