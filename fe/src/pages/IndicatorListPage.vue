<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useIndicatorStore, type IIndicator } from '@/stores/indicator.store'
import { DefaultLayout } from '@/layouts'
import { UiButton } from '@/components/common'
import { IndicatorCard, IndicatorFormModal } from '@/components/features/indicator'
import { PhPlus, PhArrowCounterClockwise } from '@phosphor-icons/vue'

const store = useIndicatorStore()

const indicators = computed(() => store.sortedItems)
const loading = computed(() => store.loading)
const showModal = ref(false)
const isEditMode = ref(false)

const openAddModal = () => {
  store.resetForm()
  isEditMode.value = false
  showModal.value = true
}

const openEditModal = (indicator: IIndicator) => {
  store.setEditMode(indicator)
  isEditMode.value = true
  showModal.value = true
}

const handleDelete = async (id: number) => {
  await store.deleteIndicator(id)
}

const handleModalSubmitted = () => {
  showModal.value = false
  store.fetchIndicators()
}

onMounted(() => {
  store.fetchIndicators()
})
</script>

<template>
  <DefaultLayout>
    <template #header-title>Indicators Management</template>

    <div>
      <!-- Actions -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <h1 class="text-2xl font-bold text-gray-900">Indicators</h1>
        <div class="flex items-center gap-2">
          <UiButton @click="openAddModal" variant="primary">
            <PhPlus :size="18" weight="bold" />
            <span>Add Indicator</span>
          </UiButton>
          <UiButton @click="store.fetchIndicators" variant="outline" :disabled="loading">
            <PhArrowCounterClockwise :size="18" :class="{ 'animate-spin': loading }" />
            <span class="hidden sm:inline">Refresh</span>
          </UiButton>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <p class="mt-4 text-gray-600">Loading indicators...</p>
      </div>

      <template v-else>
        <!-- Stats Summary -->
        <div v-if="indicators.length > 0" class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
          <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
            <div class="text-sm text-gray-500 mb-1">Total Indicators</div>
            <div class="text-2xl font-bold text-gray-900">{{ indicators.length }}</div>
          </div>
          <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
            <div class="text-sm text-gray-500 mb-1">Active</div>
            <div class="text-2xl font-bold text-green-600">{{ store.activeIndicators.length }}</div>
          </div>
          <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
            <div class="text-sm text-gray-500 mb-1">Inactive</div>
            <div class="text-2xl font-bold text-gray-600">{{ indicators.length - store.activeIndicators.length }}</div>
          </div>
        </div>

        <!-- Empty State -->
        <div v-if="indicators.length === 0" class="empty-state">
          <div class="text-center py-12">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
            <p class="mt-4 text-lg text-gray-600">No indicators found</p>
            <p class="mt-2 text-sm text-gray-500">Click "Add Indicator" to create one</p>
          </div>
        </div>

        <!-- Cards Grid -->
        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <IndicatorCard
            v-for="indicator in indicators"
            :key="indicator.id"
            :indicator="indicator"
            @edit="openEditModal"
            @delete="handleDelete"
          />
        </div>
      </template>

      <!-- Modal Form -->
      <IndicatorFormModal
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
