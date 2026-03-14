<script setup lang="ts">
import type { IConfig } from '@/stores/config.store'
import { PhPencilSimple, PhTrash } from '@phosphor-icons/vue'

const props = defineProps<{
  config: IConfig
}>()

defineEmits<{
  edit: [config: IConfig]
  delete: [id: number]
}>()
</script>

<template>
  <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5 transition-all duration-200 hover:shadow-md hover:-translate-y-0.5">
    <!-- Header -->
    <div class="mb-4">
      <h3 class="text-lg font-semibold text-gray-900">{{ config.config_key }}</h3>
    </div>

    <!-- Content -->
    <div class="mb-4">
      <div class="bg-gray-50 rounded-lg p-3">
        <span class="text-xs text-gray-500 block mb-1">Value</span>
        <span class="text-sm font-mono text-gray-900 break-all">{{ config.value }}</span>
      </div>
    </div>

    <!-- Meta -->
    <div class="text-xs text-gray-400 mb-4">
      Created: {{ new Date(config.created_at).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' }) }}
    </div>

    <!-- Actions -->
    <div class="flex gap-2">
      <button
        type="button"
        @click="$emit('edit', config)"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-500 hover:text-white transition-all duration-200"
      >
        <PhPencilSimple :size="16" weight="bold" />
        Edit
      </button>
      <button
        type="button"
        @click="$emit('delete', config.id)"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg bg-red-50 text-red-600 hover:bg-red-500 hover:text-white transition-all duration-200"
      >
        <PhTrash :size="16" weight="bold" />
        Delete
      </button>
    </div>
  </div>
</template>
