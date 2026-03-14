<script setup lang="ts">
import { computed } from 'vue'
import { useTimeframeStore } from '@/stores/timeframe.store'
import { getValidationErrors } from '@/helpers/validation'
import { UiModal, UiButton, UiInput } from '@/components/common'

const store = useTimeframeStore()

const props = defineProps<{
  modelValue: boolean
  isEditMode: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submitted': []
}>()

const title = computed(() =>
  props.isEditMode ? 'Edit Timeframe' : 'Add Timeframe'
)

const handleSubmit = async () => {
  const valid = await store.timeframeReqValid.$validate()
  if (!valid) return

  let success = false
  if (props.isEditMode && store.currentName) {
    success = await store.updateTimeframe(store.currentName)
  } else {
    success = await store.createTimeframe()
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
          v-model="store.timeframeReq.name"
          label="Name"
          placeholder="e.g., 5m"
          :error="store.timeframeReqValid.name.$error"
          :error-message="getValidationErrors(store.timeframeReqValid.name).join(', ')"
          autocomplete="off"
        />

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">In Minutes</label>
          <input
            v-model.number="store.timeframeReq.in_minutes"
            type="number"
            min="1"
            placeholder="e.g., 5"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            :class="{ 'border-danger': store.timeframeReqValid.in_minutes.$error }"
          />
          <p v-if="store.timeframeReqValid.in_minutes.$error" class="mt-1 text-sm text-danger">
            {{ getValidationErrors(store.timeframeReqValid.in_minutes).join(', ') }}
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
