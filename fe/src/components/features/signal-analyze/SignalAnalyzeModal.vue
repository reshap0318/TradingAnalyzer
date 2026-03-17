<script setup lang="ts">
import { computed } from 'vue'
import { useSignalAnalyzeStore } from '@/stores/signal-analyze.store'
import { useStrategiesStore } from '@/stores/strategies.store'
import { getValidationErrors } from '@/helpers/validation'
import { UiButton, UiInput, UiModal } from '@/components/common'
import { PhFlask } from '@phosphor-icons/vue'

interface SignalAnalyzeModalProps {
  show: boolean
}

const props = defineProps<SignalAnalyzeModalProps>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  'analyzed': []
}>()

const signalStore = useSignalAnalyzeStore()
const strategiesStore = useStrategiesStore()

// Access validation dari store
const v$ = signalStore.analyzeReqValid
const isLoading = computed(() => signalStore.loading)

// Get strategies untuk dropdown
const strategies = computed(() => strategiesStore.strategies)

// Computed untuk v-model
const showAnalyzeModal = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

// Submit handler
const handleSubmit = async () => {
  // Call action dari store (validasi sudah ada di dalam analyzeSignal)
  const success = await signalStore.analyzeSignal()

  if (success) {
    // Close modal dan notify parent
    emit('update:show', false)
    emit('analyzed')
  }
}

// Handle Enter key
const handleKeyPress = (e: KeyboardEvent) => {
  if (e.key === 'Enter') {
    handleSubmit()
  }
}

// Handle close
const handleClose = () => {
  emit('update:show', false)
}
</script>

<template>
  <UiModal
    v-model="showAnalyzeModal"
    title="New Signal Analysis"
    size="md"
    @close="handleClose"
  >
    <form @submit.prevent="handleSubmit" @keypress="handleKeyPress">
      <!-- Symbol Input -->
      <UiInput
        v-model="signalStore.analyzeReq.symbol"
        label="Symbol"
        placeholder="e.g., BTCUSDT"
        autocomplete="off"
        :error="v$?.symbol?.$error"
        :error-message="v$?.symbol?.$errors?.length ? getValidationErrors(v$.symbol).join(', ') : ''"
        class="mb-4"
      />

      <!-- Strategy Select -->
      <div class="mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          Strategy <span class="text-red-500">*</span>
        </label>
        <select
          v-model="signalStore.analyzeReq.strategy_id"
          class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
          :class="{ 'border-red-500': v$?.strategy_id?.$error }"
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
          v-if="v$?.strategy_id?.$error"
          class="mt-1 text-sm text-red-600"
        >
          {{ getValidationErrors(v$.strategy_id).join(', ') }}
        </p>
      </div>

      <!-- Capital Input -->
      <div class="mb-6">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          Capital (USD) <span class="text-red-500">*</span>
        </label>
        <input
          v-model.number="signalStore.analyzeReq.capital"
          type="number"
          placeholder="50"
          class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
          :class="{ 'border-red-500': v$?.capital?.$error }"
          :disabled="isLoading"
        />
        <p
          v-if="v$?.capital?.$error"
          class="mt-1 text-sm text-red-600"
        >
          {{ getValidationErrors(v$.capital).join(', ') }}
        </p>
      </div>

      <!-- Info Box -->
      <div class="mb-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
        <div class="flex items-start gap-3">
          <PhFlask :size="20" class="text-blue-600 mt-0.5 flex-shrink-0" />
          <div class="text-sm text-blue-800">
            <p class="font-semibold mb-1">Analysis Information</p>
            <ul class="text-xs space-y-1 text-blue-700">
              <li>• Chart will use the primary timeframe from selected strategy</li>
              <li>• TP/SL calculated based on strategy's money management</li>
              <li>• Real-time data from Binance</li>
            </ul>
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center gap-3">
        <UiButton
          type="button"
          variant="outline"
          :disabled="isLoading"
          @click="handleClose"
          class="flex-1"
        >
          Cancel
        </UiButton>
        <UiButton
          type="submit"
          variant="primary"
          :loading="isLoading"
          class="flex-1"
        >
          {{ isLoading ? 'Analyzing...' : 'Analyze Signal' }}
        </UiButton>
      </div>
    </form>
  </UiModal>
</template>
