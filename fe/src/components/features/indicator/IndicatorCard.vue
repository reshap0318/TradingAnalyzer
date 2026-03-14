<script setup lang="ts">
import type { IIndicator } from '@/stores/indicator.store'
import { PhPencilSimple, PhTrash, PhCheckCircle, PhXCircle } from '@phosphor-icons/vue'

const props = defineProps<{
  indicator: IIndicator
}>()

defineEmits<{
  edit: [indicator: IIndicator]
  delete: [id: number]
}>()

// Count params
const paramsCount = (params: Record<string, any>): number => {
  return Object.keys(params).length
}

// Format params preview
const getParamsPreview = (params: Record<string, any>): string => {
  const keys = Object.keys(params)
  if (keys.length === 0) return 'No parameters'
  if (keys.length <= 3) return keys.join(', ')
  return `${keys.slice(0, 3).join(', ')} +${keys.length - 3} more`
}
</script>

<template>
  <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5 transition-all duration-200 hover:shadow-md hover:-translate-y-0.5">
    <!-- Header -->
    <div class="flex items-start justify-between mb-3">
      <div class="flex-1">
        <h3 class="text-lg font-semibold text-gray-900">{{ indicator.name }}</h3>
        <p class="text-xs text-gray-500 font-mono mt-1">{{ indicator.indicator }}</p>
      </div>
      <span
        class="flex items-center gap-1 px-2 py-1 text-xs font-semibold rounded-full"
        :class="indicator.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'"
      >
        <PhCheckCircle v-if="indicator.is_active" :size="12" weight="fill" />
        <PhXCircle v-else :size="12" weight="fill" />
        {{ indicator.is_active ? 'Active' : 'Inactive' }}
      </span>
    </div>

    <!-- Description -->
    <p class="text-sm text-gray-600 mb-3 line-clamp-2">{{ indicator.description || 'No description' }}</p>

    <!-- Params Preview -->
    <div class="bg-gray-50 rounded-lg p-3 mb-3">
      <div class="flex items-center justify-between mb-1">
        <span class="text-xs text-gray-500 font-medium">Parameters</span>
        <span class="text-xs text-gray-400">{{ paramsCount(indicator.params) }} keys</span>
      </div>
      <p class="text-xs font-mono text-gray-700 truncate">{{ getParamsPreview(indicator.params) }}</p>
    </div>

    <!-- Meta -->
    <div class="flex items-center gap-4 text-xs text-gray-400 mb-4">
      <div>
        <span class="text-gray-500">Weight:</span>
        <span class="font-medium text-gray-700 ml-1">{{ (indicator.weight * 100).toFixed(0) }}%</span>
      </div>
      <div>
        <span class="text-gray-500">Order:</span>
        <span class="font-medium text-gray-700 ml-1">{{ indicator.order_view }}</span>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex gap-2">
      <button
        type="button"
        @click="$emit('edit', indicator)"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-500 hover:text-white transition-all duration-200"
      >
        <PhPencilSimple :size="16" weight="bold" />
        Edit
      </button>
      <button
        type="button"
        @click="$emit('delete', indicator.id)"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg bg-red-50 text-red-600 hover:bg-red-500 hover:text-white transition-all duration-200"
      >
        <PhTrash :size="16" weight="bold" />
        Delete
      </button>
    </div>
  </div>
</template>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
