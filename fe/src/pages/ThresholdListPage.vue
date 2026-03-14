<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useThresholdStore, type IThreshold } from '@/stores/threshold.store'
import { DefaultLayout } from '@/layouts'
import { UiButton } from '@/components/common'
import {
  ThresholdCard,
  ThresholdFormModal,
  ThresholdGauge
} from '@/components/features/threshold'
import { PhPlus, PhArrowCounterClockwise } from '@phosphor-icons/vue'

const store = useThresholdStore()

const thresholds = computed(() => store.sortedItems)
const loading = computed(() => store.loading)
const showModal = ref(false)
const isEditMode = ref(false)

const openAddModal = () => {
  store.resetForm()
  isEditMode.value = false
  showModal.value = true
}

const openEditModal = (threshold: IThreshold) => {
  store.setEditMode(threshold)
  isEditMode.value = true
  showModal.value = true
}

const handleDelete = async (id: number) => {
  await store.deleteThreshold(id)
}

const handleModalSubmitted = () => {
  showModal.value = false
  store.fetchThresholds()
}

onMounted(() => {
  store.fetchThresholds()
})
</script>

<template>
  <DefaultLayout>
    <template #header-title>Thresholds Management</template>

    <div>
      <!-- Actions -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <h1 class="text-2xl font-bold text-gray-900">Thresholds</h1>
        <div class="flex items-center gap-2">
          <UiButton @click="openAddModal" variant="primary">
            <PhPlus :size="18" weight="bold" />
            <span>Add Threshold</span>
          </UiButton>
          <UiButton @click="store.fetchThresholds" variant="outline" :disabled="loading">
            <PhArrowCounterClockwise :size="18" :class="{ 'animate-spin': loading }" />
            <span class="hidden sm:inline">Refresh</span>
          </UiButton>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <p class="mt-4 text-gray-600">Loading thresholds...</p>
      </div>

      <template v-else>
        <!-- Gauge Visualization -->
        <div class="mb-8">
          <ThresholdGauge :thresholds="thresholds" />
        </div>

        <!-- Cards Grid -->
        <div v-if="thresholds.length > 0">
          <h2 class="text-lg font-semibold text-gray-900 mb-4">Threshold List</h2>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <ThresholdCard
              v-for="threshold in thresholds"
              :key="threshold.id"
              :threshold="threshold"
              @edit="openEditModal"
              @delete="handleDelete"
            />
          </div>
        </div>

        <!-- Empty State -->
        <div v-else class="empty-state">
          <div class="text-center py-12">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p class="mt-4 text-lg text-gray-600">No thresholds found</p>
            <p class="mt-2 text-sm text-gray-500">Click "Add Threshold" to create one</p>
          </div>
        </div>
      </template>

      <!-- Modal Form -->
      <ThresholdFormModal
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
