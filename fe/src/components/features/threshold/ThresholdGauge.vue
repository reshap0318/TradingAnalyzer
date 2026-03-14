<script setup lang="ts">
import type { IThreshold } from '@/stores/threshold.store'

defineProps<{
  thresholds: IThreshold[]
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

// Calculate segment width
const calculateWidth = (min: number, max: number): number => {
  return ((max - min) / 200) * 100
}
</script>

<template>
  <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
    <h3 class="text-lg font-semibold text-gray-900 mb-6">Threshold Gauge</h3>

    <!-- Gauge Container -->
    <div class="relative">
      <!-- Scale Markers -->
      <div class="flex justify-between text-xs text-gray-500 mb-2">
        <span>-100</span>
        <span>-50</span>
        <span>0</span>
        <span>50</span>
        <span>100</span>
      </div>

      <!-- Gauge Bar -->
      <div class="relative h-16 bg-gray-100 rounded-lg overflow-hidden flex">
        <!-- Threshold Segments -->
        <div
          v-for="threshold in thresholds"
          :key="threshold.id"
          :class="getColorClass(threshold.color)"
          :style="{
            width: `${calculateWidth(threshold.min_value, threshold.max_value)}%`
          }"
          :title="`${threshold.category}: ${threshold.min_value} to ${threshold.max_value}`"
          class="h-full transition-all duration-200 hover:opacity-80"
        ></div>
      </div>

      <!-- Center Marker (0) -->
      <div class="absolute top-0 bottom-0 left-1/2 w-px bg-gray-300 -translate-x-1/2">
        <div class="absolute -top-1 -translate-x-1/2 w-0.5 h-3 bg-gray-300"></div>
        <div class="absolute -bottom-1 -translate-x-1/2 w-0.5 h-3 bg-gray-300"></div>
      </div>
    </div>
  </div>
</template>
