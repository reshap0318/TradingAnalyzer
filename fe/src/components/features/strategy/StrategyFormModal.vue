<script setup lang="ts">
import { computed, watch } from 'vue'
import { useStrategiesStore } from '@/stores/strategies.store'
import { useTimeframeStore } from '@/stores/timeframe.store'
import { useIndicatorStore } from '@/stores/indicator.store'
import { useConfigStore } from '@/stores/config.store'
import { getValidationErrors } from '@/helpers/validation'
import {
  configKeyToLabel,
  isBooleanField,
  getConfigNumericValue,
  getInputAttrs,
  formatConfigValue
} from '@/helpers/config'
import { PhX, PhFlask, PhPlus } from '@phosphor-icons/vue'

interface Props {
  show: boolean
  editing: boolean
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  editing: false
})

const emit = defineEmits<{
  submit: []
  close: []
}>()

const strategiesStore = useStrategiesStore()
const timeframeStore = useTimeframeStore()
const indicatorStore = useIndicatorStore()
const configStore = useConfigStore()

// Get MM configs from store
const mmConfigs = computed(() =>
  configStore.items.filter(c => c.category === 'MONEY_MANAGEMENT')
)

// Helper: Format default value display using config helper
const formatDefaultValue = (config: any): string => {
  const configKey = config.config_key
  const value = config.value

  if (isBooleanField(configKey)) {
    return value.toLowerCase() === 'true' ? 'Aggressive' : 'Conservative'
  }

  const numericValue = getConfigNumericValue(config)
  return formatConfigValue(configKey, numericValue)
}

const modalTitle = computed(() => props.editing ? 'Edit Strategy' : 'Create Strategy')
const submitButtonText = computed(() => props.editing ? 'Update Strategy' : 'Create Strategy')
const submitButtonLoading = computed(() => strategiesStore.formLoading)

// Watch modal state and toggle body scroll
watch(
  () => props.show,
  (isOpen) => {
    if (isOpen) {
      // Disable body scroll when modal is open
      document.body.style.overflow = 'hidden'
    } else {
      // Re-enable body scroll when modal is closed
      document.body.style.overflow = ''
    }
  },
  { immediate: true }
)

const handleBackdropClick = (event: MouseEvent) => {
  if (event.target === event.currentTarget) {
    emit('close')
  }
}

const handleSubmit = () => {
  emit('submit')
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="show"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4"
      @click="handleBackdropClick"
    >
      <div class="bg-white rounded-2xl shadow-2xl max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <!-- Header -->
        <div class="flex items-center justify-between p-6 border-b border-gray-200">
          <div class="flex items-center gap-3">
            <div class="p-2 bg-blue-50 rounded-xl">
              <PhFlask :size="24" class="text-blue-600" weight="fill" />
            </div>
            <div>
              <h2 class="text-xl font-bold text-gray-900">{{ modalTitle }}</h2>
              <p class="text-sm text-gray-500">Configure your trading strategy</p>
            </div>
          </div>
          <button
            @click="$emit('close')"
            class="p-2 hover:bg-gray-100 rounded-lg transition-all"
          >
            <PhX :size="24" class="text-gray-500" />
          </button>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-y-auto p-6">
          <!-- 1. Basic Information -->
          <div class="mb-8">
            <h3 class="text-lg font-semibold text-gray-900 mb-4">1. Basic Information</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  Strategy Name <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="strategiesStore.strategyForm.strategy_name"
                  type="text"
                  placeholder="e.g., Day Trading Pro"
                  class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  :class="{ 'border-red-500': strategiesStore.formValid.strategy_name?.$error }"
                />
                <p v-if="strategiesStore.formValid.strategy_name?.$error" class="mt-1 text-sm text-red-500">
                  {{ getValidationErrors(strategiesStore.formValid.strategy_name).join(', ') }}
                </p>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  Primary Timeframe <span class="text-red-500">*</span>
                </label>
                <select
                  v-model="strategiesStore.strategyForm.primary_tf"
                  class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  :class="{ 'border-red-500': strategiesStore.formValid.primary_tf?.$error }"
                >
                  <option
                    v-for="timeframe in timeframeStore.items"
                    :key="timeframe.name"
                    :value="timeframe.name"
                  >
                    {{ timeframe.name }} ({{ timeframe.in_minutes }} min)
                  </option>
                </select>
                <p v-if="strategiesStore.formValid.primary_tf?.$error" class="mt-1 text-sm text-red-500">
                  {{ getValidationErrors(strategiesStore.formValid.primary_tf).join(', ') }}
                </p>
              </div>
            </div>

            <!-- Is Active Checkbox -->
            <div class="mt-4 p-4 bg-blue-50 rounded-lg border border-blue-200">
              <label class="flex items-start gap-3 cursor-pointer">
                <input
                  v-model="strategiesStore.strategyForm.is_active"
                  type="checkbox"
                  class="w-5 h-5 text-primary rounded border-gray-300 focus:ring-2 focus:ring-primary mt-0.5"
                />
                <div>
                  <span class="text-sm font-medium text-gray-900">Active Strategy</span>
                  <p class="text-xs text-gray-600 mt-1">
                    Enable this strategy to be used by the trading bot. Only one strategy can be active at a time.
                  </p>
                </div>
              </label>
            </div>
          </div>

          <!-- 2. Timeframes -->
          <div class="mb-8">
            <h3 class="text-lg font-semibold text-gray-900 mb-4">2. Timeframes</h3>
            <div class="bg-gray-50 rounded-xl p-4 border border-gray-200">
              <div class="space-y-3">
                <div
                  v-for="(tf, index) in strategiesStore.strategyForm.timeframes"
                  :key="index"
                  class="flex items-center gap-4"
                >
                  <select
                    v-model="tf.tf"
                    class="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  >
                    <option
                      v-for="timeframe in timeframeStore.items"
                      :key="timeframe.name"
                      :value="timeframe.name"
                    >
                      {{ timeframe.name }} ({{ timeframe.in_minutes }} min)
                    </option>
                  </select>

                  <input
                    v-model="tf.weight"
                    type="number"
                    step="0.1"
                    min="0"
                    max="1"
                    placeholder="Weight (0-1)"
                    class="w-32 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  />

                  <button
                    @click="strategiesStore.strategyForm.timeframes.splice(index, 1)"
                    class="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-all"
                  >
                    <PhX :size="20" />
                  </button>
                </div>
              </div>

              <button
                @click="strategiesStore.strategyForm.timeframes.push({ tf: '15m', weight: 0.5 })"
                class="mt-4 w-full py-2 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-primary hover:text-primary transition-all flex items-center justify-center gap-2"
              >
                <PhPlus :size="20" />
                Add Timeframe
              </button>
            </div>
            <p v-if="strategiesStore.formValid.timeframes?.$error" class="mt-1 text-sm text-red-500">
              {{ getValidationErrors(strategiesStore.formValid.timeframes).join(', ') }}
            </p>
          </div>

          <!-- 3. Indicators -->
          <div class="mb-8">
            <h3 class="text-lg font-semibold text-gray-900 mb-4">3. Indicators</h3>
            <div class="bg-gray-50 rounded-xl p-4 border border-gray-200">
              <div class="space-y-3">
                <div
                  v-for="(ind, index) in strategiesStore.strategyForm.indicator_weights"
                  :key="index"
                  class="flex items-center gap-4"
                >
                  <select
                    v-model.number="ind.indicator_id"
                    class="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  >
                    <option value="" disabled>Select Indicator</option>
                    <option
                      v-for="indicator in indicatorStore.items"
                      :key="indicator.id"
                      :value="indicator.id"
                    >
                      {{ indicator.name }}
                    </option>
                  </select>

                  <input
                    v-model.number="ind.weight"
                    type="number"
                    step="0.1"
                    min="0"
                    max="1"
                    placeholder="Weight (0-1)"
                    class="w-32 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  />

                  <button
                    @click="strategiesStore.strategyForm.indicator_weights.splice(index, 1)"
                    class="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-all"
                  >
                    <PhX :size="20" />
                  </button>
                </div>
              </div>

              <button
                @click="strategiesStore.strategyForm.indicator_weights.push({ indicator_id: 0, weight: 0.5 })"
                class="mt-4 w-full py-2 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-primary hover:text-primary transition-all flex items-center justify-center gap-2"
              >
                <PhPlus :size="20" />
                Add Indicator
              </button>
            </div>
            <p v-if="strategiesStore.formValid.indicator_weights?.$error" class="mt-1 text-sm text-red-500">
              {{ getValidationErrors(strategiesStore.formValid.indicator_weights).join(', ') }}
            </p>
          </div>

          <!-- 4. Money Management -->
          <div>
            <h3 class="text-lg font-semibold text-gray-900 mb-4">4. Money Management</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <!-- Dynamic fields from config -->
              <div v-for="config in mmConfigs" :key="config.config_key">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  {{ configKeyToLabel(config.config_key) }}
                  <span v-if="!isBooleanField(config.config_key)" class="text-xs text-gray-500">
                    (Default: {{ formatDefaultValue(config) }})
                  </span>
                </label>
                
                <!-- Boolean field (radio buttons) -->
                <div v-if="isBooleanField(config.config_key)" class="flex items-center gap-4 mt-5.5">
                  <label class="inline-flex items-center gap-2 cursor-pointer">
                    <input
                      v-model.number="(strategiesStore.strategyForm.money_management as any)[config.config_key.toLowerCase()]"
                      type="radio"
                      :value="0"
                      class="w-4 h-4 text-primary"
                    />
                    <span class="text-sm text-gray-700">Conservative</span>
                  </label>
                  <label class="inline-flex items-center gap-2 cursor-pointer">
                    <input
                      v-model.number="(strategiesStore.strategyForm.money_management as any)[config.config_key.toLowerCase()]"
                      type="radio"
                      :value="1"
                      class="w-4 h-4 text-primary"
                    />
                    <span class="text-sm text-gray-700">Aggressive</span>
                  </label>
                </div>
                
                <!-- Number field -->
                <input
                  v-else
                  v-model.number="(strategiesStore.strategyForm.money_management as any)[config.config_key.toLowerCase()]"
                  v-bind="getInputAttrs(config)"
                  class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="flex items-center justify-end gap-3 p-6 border-t border-gray-200 bg-gray-50">
          <button
            @click="$emit('close')"
            type="button"
            class="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-100 transition-all"
          >
            Cancel
          </button>
          <button
            @click="handleSubmit"
            type="button"
            :disabled="submitButtonLoading"
            class="px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            <PhFlask v-if="!submitButtonLoading" :size="20" weight="fill" />
            {{ submitButtonLoading ? 'Saving...' : submitButtonText }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
