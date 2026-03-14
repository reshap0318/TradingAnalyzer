<script setup lang="ts">
import type { IThreshold } from '@/stores/threshold.store'
import { PhPencilSimple, PhTrash } from '@phosphor-icons/vue'

const props = defineProps<{
  threshold: IThreshold
}>()

defineEmits<{
  edit: [threshold: IThreshold]
  delete: [id: number]
}>()

// Map color name to Tailwind class
const getColorClass = (color: string): string => {
  const colorMap: Record<string, string> = {
    green: 'bg-green-500',
    'light-green': 'bg-lime-400',
    gray: 'bg-gray-400',
    red: 'bg-red-500',
    'dark-red': 'bg-red-700',
    blue: 'bg-blue-500',
    yellow: 'bg-yellow-500',
    purple: 'bg-purple-500',
    orange: 'bg-orange-500'
  }
  return colorMap[color] || 'bg-gray-400'
}
</script>

<template>
  <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5 transition-all duration-200 hover:shadow-md hover:-translate-y-0.5">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-3">
        <div :class="getColorClass(threshold.color)" class="w-4 h-4 rounded"></div>
        <h3 class="text-lg font-semibold text-gray-900">{{ threshold.category }}</h3>
      </div>
      <span class="text-xs font-medium text-gray-500">Order: {{ threshold.order_display }}</span>
    </div>

    <!-- Content -->
    <div class="space-y-3 mb-4">
      <!-- Range -->
      <div class="flex items-center justify-between text-sm">
        <span class="text-gray-600">Range:</span>
        <div class="flex items-center gap-2">
          <span class="px-2 py-1 bg-gray-100 rounded text-gray-700 font-medium">{{ threshold.min_value }}</span>
          <span class="text-gray-400">to</span>
          <span class="px-2 py-1 bg-gray-100 rounded text-gray-700 font-medium">{{ threshold.max_value }}</span>
        </div>
      </div>

      <!-- Action -->
      <div class="flex items-center justify-between text-sm">
        <span class="text-gray-600">Action:</span>
        <span
          class="px-3 py-1 rounded-full text-xs font-semibold"
          :class="{
            'bg-green-100 text-green-700': threshold.action === 'BUY',
            'bg-red-100 text-red-700': threshold.action === 'SELL',
            'bg-gray-100 text-gray-700': threshold.action === 'WAIT'
          }"
        >
          {{ threshold.action }}
        </span>
      </div>

      <!-- Color -->
      <div class="flex items-center justify-between text-sm">
        <span class="text-gray-600">Color:</span>
        <div class="flex items-center gap-2">
          <div :class="getColorClass(threshold.color)" class="w-6 h-4 rounded"></div>
          <span class="text-gray-500 capitalize">{{ threshold.color }}</span>
        </div>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex gap-2">
      <button
        type="button"
        @click="$emit('edit', threshold)"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-500 hover:text-white transition-all duration-200"
      >
        <PhPencilSimple :size="16" weight="bold" />
        Edit
      </button>
      <button
        type="button"
        @click="$emit('delete', threshold.id)"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg bg-red-50 text-red-600 hover:bg-red-500 hover:text-white transition-all duration-200"
      >
        <PhTrash :size="16" weight="bold" />
        Delete
      </button>
    </div>
  </div>
</template>
