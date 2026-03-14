<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useTimeframeStore, type ITimeframe } from '@/stores/timeframe.store'
import { DefaultLayout } from '@/layouts'
import { UiButton } from '@/components/common'
import { TimeframeCard, TimeframeFormModal } from '@/components/features/timeframe'
import { PhPlus, PhArrowCounterClockwise } from '@phosphor-icons/vue'

const store = useTimeframeStore()

const timeframes = computed(() => store.sortedItems)
const loading = computed(() => store.loading)
const showModal = ref(false)
const isEditMode = ref(false)

const openAddModal = () => {
  store.resetForm()
  isEditMode.value = false
  showModal.value = true
}

const openEditModal = (timeframe: ITimeframe) => {
  store.setEditMode(timeframe)
  isEditMode.value = true
  showModal.value = true
}

const handleDelete = async (name: string) => {
  await store.deleteTimeframe(name)
}

const handleModalSubmitted = () => {
  showModal.value = false
  store.fetchTimeframes()
}

onMounted(() => {
  store.fetchTimeframes()
})
</script>

<template>
  <DefaultLayout>
    <template #header-title>Timeframes Management</template>

    <div>
      <!-- Actions -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <h1 class="text-2xl font-bold text-gray-900">Timeframes</h1>
        <div class="flex items-center gap-2">
          <UiButton @click="openAddModal" variant="primary">
            <PhPlus :size="18" weight="bold" />
            <span>Add Timeframe</span>
          </UiButton>
          <UiButton @click="store.fetchTimeframes" variant="outline" :disabled="loading">
            <PhArrowCounterClockwise :size="18" :class="{ 'animate-spin': loading }" />
            <span class="hidden sm:inline">Refresh</span>
          </UiButton>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <p class="mt-4 text-gray-600">Loading timeframes...</p>
      </div>

      <!-- Grid Cards -->
      <div
        v-else-if="timeframes.length > 0"
        class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4"
      >
        <TimeframeCard
          v-for="tf in timeframes"
          :key="tf.name"
          :timeframe="tf"
          @edit="() => openEditModal(tf)"
          @delete="handleDelete"
        />
      </div>

      <!-- Empty State -->
      <div v-else class="empty-state">
        <div class="text-center py-12">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p class="mt-4 text-lg text-gray-600">No timeframes found</p>
          <p class="mt-2 text-sm text-gray-500">Click "Add Timeframe" to create one</p>
        </div>
      </div>

      <!-- Modal Form -->
      <TimeframeFormModal
        v-model="showModal"
        :is-edit-mode="isEditMode"
        @submitted="handleModalSubmitted"
      />
    </div>
  </DefaultLayout>
</template>

<style scoped>
.empty-state {
  background-color: white;
  border: 2px dashed #e5e7eb;
  border-radius: 0.75rem;
  padding: 3rem;
  text-align: center;
}
</style>
