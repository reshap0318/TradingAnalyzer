<script setup lang="ts">
import { computed } from 'vue'
import { useThresholdStore } from '@/stores/threshold.store'
import { getValidationErrors } from '@/helpers/validation'
import { UiModal, UiButton, UiInput } from '@/components/common'

const store = useThresholdStore()

const props = defineProps<{
  modelValue: boolean
  isEditMode: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submitted': []
}>()

const title = computed(() =>
  props.isEditMode ? 'Edit Threshold' : 'Add Threshold'
)

const handleSubmit = async () => {
  const valid = await store.thresholdReqValid.$validate()
  if (!valid) return

  let success = false
  if (props.isEditMode && store.currentId) {
    success = await store.updateThreshold(store.currentId)
  } else {
    success = await store.createThreshold()
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

// Color options
const colorOptions = [
  { value: 'green', label: 'Green', class: 'bg-green-500' },
  { value: 'light-green', label: 'Light Green', class: 'bg-lime-400' },
  { value: 'gray', label: 'Gray', class: 'bg-gray-400' },
  { value: 'red', label: 'Red', class: 'bg-red-500' },
  { value: 'dark-red', label: 'Dark Red', class: 'bg-red-700' },
  { value: 'blue', label: 'Blue', class: 'bg-blue-500' },
  { value: 'yellow', label: 'Yellow', class: 'bg-yellow-500' },
  { value: 'purple', label: 'Purple', class: 'bg-purple-500' },
  { value: 'orange', label: 'Orange', class: 'bg-orange-500' }
]

// Action options
const actionOptions = [
  { value: 'BUY', label: 'BUY' },
  { value: 'SELL', label: 'SELL' },
  { value: 'WAIT', label: 'WAIT' }
]
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
        <UiInput
          v-model="store.thresholdReq.category"
          label="Category"
          placeholder="e.g., STRONG_BUY"
          :error="store.thresholdReqValid.category.$error"
          :error-message="getValidationErrors(store.thresholdReqValid.category).join(', ')"
          autocomplete="off"
        />

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Min Value (-100 to 100)</label>
            <input
              v-model.number="store.thresholdReq.min_value"
              type="number"
              min="-100"
              max="100"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary"
              :class="{ 'border-danger': store.thresholdReqValid.min_value.$error }"
            />
            <p v-if="store.thresholdReqValid.min_value.$error" class="mt-1 text-sm text-danger">
              {{ getValidationErrors(store.thresholdReqValid.min_value).join(', ') }}
            </p>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Max Value (-100 to 100)</label>
            <input
              v-model.number="store.thresholdReq.max_value"
              type="number"
              min="-100"
              max="100"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary"
              :class="{ 'border-danger': store.thresholdReqValid.max_value.$error }"
            />
            <p v-if="store.thresholdReqValid.max_value.$error" class="mt-1 text-sm text-danger">
              {{ getValidationErrors(store.thresholdReqValid.max_value).join(', ') }}
            </p>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">Action</label>
          <div class="flex gap-2">
            <button
              v-for="action in actionOptions"
              :key="action.value"
              type="button"
              @click="store.thresholdReq.action = action.value as 'BUY' | 'SELL' | 'WAIT'"
              class="flex-1 px-4 py-2 rounded-lg text-sm font-medium transition-all border-2"
              :class="
                store.thresholdReq.action === action.value
                  ? 'border-blue-500 bg-blue-50 text-blue-700'
                  : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300'
              "
            >
              {{ action.label }}
            </button>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">Color</label>
          <div class="grid grid-cols-5 gap-2">
            <button
              v-for="color in colorOptions"
              :key="color.value"
              type="button"
              @click="store.thresholdReq.color = color.value"
              class="h-10 rounded-lg transition-all border-2 flex items-center justify-center"
              :class="
                store.thresholdReq.color === color.value
                  ? 'border-blue-500 ring-2 ring-blue-200'
                  : 'border-gray-200'
              "
            >
              <div :class="color.class" class="w-6 h-6 rounded"></div>
            </button>
          </div>
          <p class="mt-1 text-xs text-gray-500">Selected: {{ store.thresholdReq.color }}</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Order Display</label>
          <input
            v-model.number="store.thresholdReq.order_display"
            type="number"
            min="1"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary"
            :class="{ 'border-danger': store.thresholdReqValid.order_display.$error }"
          />
          <p v-if="store.thresholdReqValid.order_display.$error" class="mt-1 text-sm text-danger">
            {{ getValidationErrors(store.thresholdReqValid.order_display).join(', ') }}
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
