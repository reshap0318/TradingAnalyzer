<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useConfigStore } from '@/stores/config.store'
import { DefaultLayout } from '@/layouts'
import { UiButton } from '@/components/common'
import { ConfigCard, ConfigFormModal } from '@/components/features/config'
import { PhPlus, PhArrowCounterClockwise } from '@phosphor-icons/vue'

const store = useConfigStore()

const configsByCategory = computed(() => store.groupedByCategory)
const categories = computed(() => store.categories)
const loading = computed(() => store.loading)
const showModal = ref(false)
const isEditMode = ref(false)

const openAddModal = () => {
  store.resetForm()
  isEditMode.value = false
  showModal.value = true
}

const openEditModal = (config: any) => {
  store.setEditMode(config)
  isEditMode.value = true
  showModal.value = true
}

const handleDelete = async (id: number) => {
  await store.deleteConfig(id)
}

const handleModalSubmitted = () => {
  showModal.value = false
  store.fetchConfigs()
}

onMounted(() => {
  store.fetchConfigs()
})
</script>

<template>
  <DefaultLayout>
    <template #header-title>Configs Management</template>

    <div>
      <!-- Actions -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <h1 class="text-2xl font-bold text-gray-900">Configs</h1>
        <div class="flex items-center gap-2">
          <UiButton @click="openAddModal" variant="primary">
            <PhPlus :size="18" weight="bold" />
            <span>Add Config</span>
          </UiButton>
          <UiButton @click="store.fetchConfigs" variant="outline" :disabled="loading">
            <PhArrowCounterClockwise :size="18" :class="{ 'animate-spin': loading }" />
            <span class="hidden sm:inline">Refresh</span>
          </UiButton>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <p class="mt-4 text-gray-600">Loading configs...</p>
      </div>

      <template v-else>
        <!-- Empty State -->
        <div v-if="categories.length === 0" class="empty-state">
          <div class="text-center py-12">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
            <p class="mt-4 text-lg text-gray-600">No configs found</p>
            <p class="mt-2 text-sm text-gray-500">Click "Add Config" to create one</p>
          </div>
        </div>

        <!-- Grouped by Category -->
        <div v-else class="space-y-8">
          <div v-for="category in categories" :key="category">
            <div class="flex items-center gap-3 mb-4">
              <h2 class="text-lg font-semibold text-gray-900">{{ category }}</h2>
              <span class="px-2 py-1 text-xs font-medium bg-gray-100 text-gray-600 rounded-full">
                {{ configsByCategory[category]?.length ?? 0 }}
              </span>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              <ConfigCard
                v-for="config in configsByCategory[category]"
                :key="config.id"
                :config="config"
                @edit="openEditModal"
                @delete="handleDelete"
              />
            </div>
          </div>
        </div>
      </template>

      <!-- Modal Form -->
      <ConfigFormModal
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
