<script setup lang="ts">
import type { IWatchlist } from '@/stores/watchlist.store'
import { PhPencilSimple, PhTrash, PhCheckCircle, PhXCircle } from '@phosphor-icons/vue'

const props = defineProps<{
  watchlist: IWatchlist
}>()

defineEmits<{
  edit: [watchlist: IWatchlist]
  delete: [id: number]
}>()
</script>

<template>
  <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5 transition-all duration-200 hover:shadow-md hover:-translate-y-0.5">
    <!-- Header -->
    <div class="flex items-start justify-between mb-4">
      <div class="flex-1">
        <h3 class="text-xl font-bold text-gray-900 font-mono">{{ watchlist.symbol }}</h3>
      </div>
      <span
        class="flex items-center gap-1 px-2 py-1 text-xs font-semibold rounded-full"
        :class="watchlist.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'"
      >
        <PhCheckCircle v-if="watchlist.is_active" :size="12" weight="fill" />
        <PhXCircle v-else :size="12" weight="fill" />
        {{ watchlist.is_active ? 'Active' : 'Inactive' }}
      </span>
    </div>

    <!-- Meta -->
    <div class="text-xs text-gray-400 mb-4">
      Added: {{ new Date(watchlist.created_at).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' }) }}
    </div>

    <!-- Actions -->
    <div class="flex gap-2">
      <button
        type="button"
        @click="$emit('edit', watchlist)"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-500 hover:text-white transition-all duration-200"
      >
        <PhPencilSimple :size="16" weight="bold" />
        Edit
      </button>
      <button
        type="button"
        @click="$emit('delete', watchlist.id)"
        class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg bg-red-50 text-red-600 hover:bg-red-500 hover:text-white transition-all duration-200"
      >
        <PhTrash :size="16" weight="bold" />
        Delete
      </button>
    </div>
  </div>
</template>
