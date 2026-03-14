<script setup lang="ts">
import { computed } from 'vue'
import { useWatchlistStore } from '@/stores/watchlist.store'
import { getValidationErrors } from '@/helpers/validation'
import { UiModal, UiButton, UiInput } from '@/components/common'

const store = useWatchlistStore()

const props = defineProps<{
  modelValue: boolean
  isEditMode: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submitted': []
}>()

const title = computed(() =>
  props.isEditMode ? 'Edit Watchlist' : 'Add Watchlist'
)

const handleSubmit = async () => {
  const valid = await store.watchlistReqValid.$validate()
  if (!valid) return

  let success = false
  if (props.isEditMode && store.currentId) {
    success = await store.updateWatchlist(store.currentId)
  } else {
    success = await store.createWatchlist()
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
          v-model="store.watchlistReq.symbol"
          label="Symbol"
          placeholder="e.g., BTCUSDT"
          :error="store.watchlistReqValid.symbol.$error"
          :error-message="getValidationErrors(store.watchlistReqValid.symbol).join(', ')"
          autocomplete="off"
        />

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">Is Active</label>
          <label class="flex items-center cursor-pointer">
            <input
              v-model="store.watchlistReq.is_active"
              type="checkbox"
              class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500"
            />
            <span class="ml-2 text-sm text-gray-700">
              {{ store.watchlistReq.is_active ? 'Active' : 'Inactive' }}
            </span>
          </label>
          <p class="mt-1 text-xs text-gray-500">
            Active watchlists will be included in trading operations
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
