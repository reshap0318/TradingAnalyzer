<script setup lang="ts">
import { computed, watch } from 'vue'
import { useIndicatorStore } from '@/stores/indicator.store'
import { getValidationErrors } from '@/helpers/validation'
import { UiModal, UiButton, UiInput } from '@/components/common'
import { PhCheckCircle, PhXCircle, PhCode, PhBroom } from '@phosphor-icons/vue'
import { showError } from '@/lib/sweetalert'

const store = useIndicatorStore()

const props = defineProps<{
  modelValue: boolean
  isEditMode: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submitted': []
}>()

const title = computed(() =>
  props.isEditMode ? 'Edit Indicator' : 'Add Indicator'
)

// Watch paramsJson for validation
watch(() => store.paramsJson, () => {
  if (store.paramsJson.trim()) {
    store.validateParamsJson()
  }
}, { immediate: true })

const handleSubmit = async () => {
  // Validate basic fields
  const valid = await store.indicatorReqValid.$validate()
  if (!valid) return

  // Validate JSON params
  if (!store.validateParamsJson()) {
    showError('Invalid JSON', 'Please fix the JSON parameters first')
    return
  }

  let success = false
  if (props.isEditMode && store.currentId) {
    success = await store.updateIndicator(store.currentId)
  } else {
    success = await store.createIndicator()
  }

  if (success) {
    emit('submitted')
    emit('update:modelValue', false)
  }
}

const handleCancel = () => {
  store.resetForm()
  emit('update:modelValue', false)
}

const handleFormatJson = () => {
  store.formatParamsJson()
}

const handleClearJson = () => {
  store.paramsJson = '{}'
  store.validateParamsJson()
}

// Import sample params
const loadSampleParams = () => {
  const sampleParams = {
    sma_periods: [20, 50, 200],
    ema_periods: [12, 26]
  }
  store.paramsJson = JSON.stringify(sampleParams, null, 2)
  store.validateParamsJson()
}
</script>

<template>
  <UiModal
    :model-value="modelValue"
    :title="title"
    size="xl"
    @update:model-value="handleCancel"
  >
    <form @submit.prevent="handleSubmit">
      <div class="space-y-4">
        <!-- Basic Info -->
        <div class="grid grid-cols-2 gap-4">
          <UiInput
            v-model="store.indicatorReq.name"
            label="Name"
            placeholder="e.g., Moving Average"
            :error="store.indicatorReqValid.name.$error"
            :error-message="getValidationErrors(store.indicatorReqValid.name as any).join(', ')"
            autocomplete="off"
          />

          <UiInput
            v-model="store.indicatorReq.indicator"
            label="Indicator Key"
            placeholder="e.g., moving_average"
            :error="store.indicatorReqValid.indicator.$error"
            :error-message="getValidationErrors(store.indicatorReqValid.indicator as any).join(', ')"
            autocomplete="off"
          />
        </div>

        <UiInput
          v-model="store.indicatorReq.description"
          label="Description"
          placeholder="Describe what this indicator does..."
          autocomplete="off"
        />

        <div class="grid grid-cols-3 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Weight (0-1)
            </label>
            <input
              v-model.number="store.indicatorReq.weight"
              type="number"
              step="0.01"
              min="0"
              max="1"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary"
              :class="{ 'border-danger': store.indicatorReqValid.weight.$error }"
            />
            <p v-if="store.indicatorReqValid.weight.$error" class="mt-1 text-sm text-danger">
              {{ getValidationErrors(store.indicatorReqValid.weight as any).join(', ') }}
            </p>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Order View
            </label>
            <input
              v-model.number="store.indicatorReq.order_view"
              type="number"
              min="1"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary"
              :class="{ 'border-danger': store.indicatorReqValid.order_view.$error }"
            />
            <p v-if="store.indicatorReqValid.order_view.$error" class="mt-1 text-sm text-danger">
              {{ getValidationErrors(store.indicatorReqValid.order_view as any).join(', ') }}
            </p>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Is Active
            </label>
            <div class="flex items-center h-[42px]">
              <label class="flex items-center cursor-pointer">
                <input
                  v-model="store.indicatorReq.is_active"
                  type="checkbox"
                  class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500"
                />
                <span class="ml-2 text-sm text-gray-700">Yes</span>
              </label>
            </div>
          </div>
        </div>

        <!-- JSON Parameters Section -->
        <div>
          <div class="flex items-center justify-between mb-2">
            <label class="block text-sm font-medium text-gray-700">
              Parameters (JSON)
            </label>
            <div class="flex gap-2">
              <button
                type="button"
                @click="loadSampleParams"
                class="px-2 py-1 text-xs font-medium text-blue-600 bg-blue-50 rounded hover:bg-blue-100 transition-colors"
                title="Load sample params"
              >
                <PhCode :size="14" class="inline" />
                Sample
              </button>
              <button
                type="button"
                @click="handleFormatJson"
                class="px-2 py-1 text-xs font-medium text-purple-600 bg-purple-50 rounded hover:bg-purple-100 transition-colors"
                title="Format JSON"
              >
                <PhBroom :size="14" class="inline" />
                Format
              </button>
              <button
                type="button"
                @click="handleClearJson"
                class="px-2 py-1 text-xs font-medium text-gray-600 bg-gray-100 rounded hover:bg-gray-200 transition-colors"
                title="Clear JSON"
              >
                Clear
              </button>
            </div>
          </div>

          <div class="relative">
            <textarea
              v-model="store.paramsJson"
              rows="10"
              spellcheck="false"
              class="w-full px-3 py-2 font-mono text-sm bg-gray-50 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary resize-y"
              :class="{
                'border-green-500 bg-green-50': store.paramsJsonValid && store.paramsJson.trim(),
                'border-danger bg-red-50': !store.paramsJsonValid
              }"
              placeholder='{"key": "value"}'
            ></textarea>

            <!-- Status Indicator -->
            <div class="absolute top-2 right-2">
              <PhCheckCircle
                v-if="store.paramsJsonValid && store.paramsJson.trim()"
                :size="20"
                class="text-green-500"
                weight="fill"
              />
              <PhXCircle
                v-else-if="!store.paramsJsonValid"
                :size="20"
                class="text-red-500"
                weight="fill"
              />
            </div>
          </div>

          <!-- Error Message -->
          <p
            v-if="!store.paramsJsonValid && store.paramsJsonError"
            class="mt-2 text-sm text-red-600 bg-red-50 p-2 rounded border border-red-200"
          >
            <strong>JSON Error:</strong> {{ store.paramsJsonError }}
          </p>

          <!-- Success Message -->
          <p
            v-else-if="store.paramsJsonValid && store.paramsJson.trim()"
            class="mt-2 text-sm text-green-600"
          >
            ✓ Valid JSON
          </p>

          <p class="mt-1 text-xs text-gray-500">
            Support nested objects, arrays, and all JSON types
          </p>
        </div>
      </div>

      <div class="flex justify-end gap-3 mt-6 pt-6 border-t border-gray-200">
        <UiButton
          type="button"
          variant="outline"
          :disabled="store.loading"
          @click="handleCancel"
        >
          Cancel
        </UiButton>
        <UiButton
          type="submit"
          variant="primary"
          :loading="store.loading"
        >
          {{ isEditMode ? 'Update' : 'Create' }}
        </UiButton>
      </div>
    </form>
  </UiModal>
</template>
