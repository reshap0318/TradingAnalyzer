<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useWatchlistStore, type IWatchlist } from '@/stores/watchlist.store'
import { DefaultLayout } from '@/layouts'
import { UiButton } from '@/components/common'
import { WatchlistCard, WatchlistFormModal } from '@/components/features/watchlist'
import { PhPlus, PhArrowCounterClockwise } from '@phosphor-icons/vue'

const store = useWatchlistStore()

const watchlists = computed(() => store.sortedItems)
const activeCount = computed(() => store.activeWatchlists.length)
const inactiveCount = computed(() => store.inactiveWatchlists.length)
const loading = computed(() => store.loading)
const showModal = ref(false)
const isEditMode = ref(false)

const openAddModal = () => {
  store.resetForm()
  isEditMode.value = false
  showModal.value = true
}

const openEditModal = (watchlist: IWatchlist) => {
  store.setEditMode(watchlist)
  isEditMode.value = true
  showModal.value = true
}

const handleDelete = async (id: number) => {
  await store.deleteWatchlist(id)
}

const handleModalSubmitted = () => {
  showModal.value = false
  store.fetchWatchlists()
}

onMounted(() => {
  store.fetchWatchlists()
})
</script>

<template>
  <DefaultLayout>
    <template #header-title>Watchlists Management</template>

    <div>
      <!-- Actions -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <h1 class="text-2xl font-bold text-gray-900">Watchlists</h1>
        <div class="flex items-center gap-2">
          <UiButton @click="openAddModal" variant="primary">
            <PhPlus :size="18" weight="bold" />
            <span>Add Symbol</span>
          </UiButton>
          <UiButton @click="store.fetchWatchlists" variant="outline" :disabled="loading">
            <PhArrowCounterClockwise :size="18" :class="{ 'animate-spin': loading }" />
            <span class="hidden sm:inline">Refresh</span>
          </UiButton>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <p class="mt-4 text-gray-600">Loading watchlists...</p>
      </div>

      <template v-else>
        <!-- Stats Summary -->
        <div v-if="watchlists.length > 0" class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
          <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
            <div class="text-sm text-gray-500 mb-1">Total Symbols</div>
            <div class="text-2xl font-bold text-gray-900">{{ watchlists.length }}</div>
          </div>
          <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
            <div class="text-sm text-gray-500 mb-1">Active</div>
            <div class="text-2xl font-bold text-green-600">{{ activeCount }}</div>
          </div>
          <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
            <div class="text-sm text-gray-500 mb-1">Inactive</div>
            <div class="text-2xl font-bold text-gray-600">{{ inactiveCount }}</div>
          </div>
        </div>

        <!-- Empty State -->
        <div v-if="watchlists.length === 0" class="empty-state">
          <div class="text-center py-12">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
            <p class="mt-4 text-lg text-gray-600">No watchlists found</p>
            <p class="mt-2 text-sm text-gray-500">Click "Add Symbol" to create one</p>
          </div>
        </div>

        <!-- Cards Grid -->
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          <WatchlistCard
            v-for="watchlist in watchlists"
            :key="watchlist.id"
            :watchlist="watchlist"
            @edit="openEditModal"
            @delete="handleDelete"
          />
        </div>
      </template>

      <!-- Modal Form -->
      <WatchlistFormModal
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
