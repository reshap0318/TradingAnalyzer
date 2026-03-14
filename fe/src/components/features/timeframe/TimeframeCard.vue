<script setup lang="ts">
import type { ITimeframe } from '@/stores/timeframe.store'
import { PhPencilSimple, PhTrash } from '@phosphor-icons/vue'

const props = defineProps<{
  timeframe: ITimeframe
}>()

defineEmits<{
  edit: [name: string]
  delete: [name: string]
}>()

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}
</script>

<template>
  <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5 transition-all duration-200 hover:shadow-md hover:-translate-y-0.5">
    <!-- Header -->
    <div class="flex items-center gap-3 mb-4">
      <span class="text-2xl">⏱️</span>
      <h3 class="text-lg font-semibold text-gray-900">{{ timeframe.name }}</h3>
    </div>

    <!-- Content -->
    <div class="space-y-2 mb-4">
      <p class="text-sm text-gray-600">
        <span class="font-medium text-gray-700">Minutes:</span> {{ timeframe.in_minutes }}
      </p>
      <p class="text-sm text-gray-600">
        <span class="font-medium text-gray-700">Created:</span> {{ formatDate(timeframe.created_at) }}
      </p>
    </div>

    <!-- Actions -->
    <div class="flex gap-2">
      <button
        type="button"
        @click="$emit('edit', timeframe.name)"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-500 hover:text-white transition-all duration-200"
      >
        <PhPencilSimple :size="16" weight="bold" />
        Edit
      </button>
      <button
        type="button"
        @click="$emit('delete', timeframe.name)"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg bg-red-50 text-red-600 hover:bg-red-500 hover:text-white transition-all duration-200"
      >
        <PhTrash :size="16" weight="bold" />
        Delete
      </button>
    </div>
  </div>
</template>
