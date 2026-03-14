<script setup lang="ts">
import { computed } from 'vue'
import { useConfigStore } from '@/stores/config.store'
import { getValidationErrors } from '@/helpers/validation'
import { UiModal, UiButton, UiInput } from '@/components/common'

const store = useConfigStore()

const props = defineProps<{
  modelValue: boolean
  isEditMode: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submitted': []
}>()

const title = computed(() =>
  props.isEditMode ? 'Edit Config' : 'Add Config'
)

const handleSubmit = async () => {
  const valid = await store.configReqValid.$validate()
  if (!valid) return

  let success = false
  if (props.isEditMode && store.currentId) {
    success = await store.updateConfig(store.currentId)
  } else {
    success = await store.createConfig()
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

// Common category suggestions
const categorySuggestions = [
  'MONEY_MANAGEMENT',
  'BINANCE',
  'TRADING',
  'SYSTEM',
  'CUSTOM'
]
</script>

<template>
  <UiModal
    :model-value="modelValue"
    :title="title"
    @update:model-value="handleCancel"
  >
    <form @submit.prevent="handleSubmit">
      <div class="space-y-4">
        <UiInput
          v-model="store.configReq.config_key"
          label="Config Key"
          placeholder="e.g., MIN_CONFIDENCE"
          :error="store.configReqValid.config_key.$error"
          :error-message="getValidationErrors(store.configReqValid.config_key as any).join(', ')"
          autocomplete="off"
        />

        <UiInput
          v-model="store.configReq.value"
          label="Value"
          placeholder="e.g., 45"
          :error="store.configReqValid.value.$error"
          :error-message="getValidationErrors(store.configReqValid.value as any).join(', ')"
          autocomplete="off"
        />

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Category</label>
          <div class="flex flex-wrap gap-2 mb-2">
            <button
              v-for="cat in categorySuggestions"
              :key="cat"
              type="button"
              @click="store.configReq.category = cat"
              class="px-3 py-1 text-xs font-medium rounded-full border transition-all"
              :class="
                store.configReq.category === cat
                  ? 'border-blue-500 bg-blue-50 text-blue-700'
                  : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300'
              "
            >
              {{ cat }}
            </button>
          </div>
          <input
            v-model="store.configReq.category"
            type="text"
            placeholder="Or type custom category"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary"
            :class="{ 'border-danger': store.configReqValid.category.$error }"
          />
          <p v-if="store.configReqValid.category.$error" class="mt-1 text-sm text-danger">
            {{ getValidationErrors(store.configReqValid.category as any).join(', ') }}
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
