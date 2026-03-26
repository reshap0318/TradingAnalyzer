<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useStrategiesStore } from '@/stores/strategies.store'
import { useTimeframeStore } from '@/stores/timeframe.store'
import { useIndicatorStore } from '@/stores/indicator.store'
import { useConfigStore } from '@/stores/config.store'
import { DefaultLayout } from '@/layouts'
import {
  PhFlask,
  PhPlus,
  PhPencilSimple,
  PhTrash,
  PhMagnifyingGlass
} from '@phosphor-icons/vue'
import StrategyFormModal from '@/components/features/strategy/StrategyFormModal.vue'

const store = useStrategiesStore()
const timeframeStore = useTimeframeStore()
const indicatorStore = useIndicatorStore()
const configStore = useConfigStore()

const searchQuery = ref('')
const showFormModal = ref(false)
const editingStrategy = ref(false)
const currentEditId = ref<number | null>(null)

const filteredStrategies = computed(() => {
  if (!searchQuery.value) return store.strategies
  
  return store.strategies.filter(strategy =>
    strategy.strategy_name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const handleCreate = () => {
  store.resetForm()
  editingStrategy.value = false
  currentEditId.value = null
  showFormModal.value = true
}

const handleEdit = async (id: number) => {
  const strategy = await store.fetchStrategy(id)
  if (strategy) {
    store.loadStrategyToForm(strategy)
    editingStrategy.value = true
    currentEditId.value = id
    showFormModal.value = true
  }
}

const handleDelete = async (id: number, name: string) => {
  await store.deleteStrategy(id, name)
}

const handleFormSubmit = async () => {
  if (editingStrategy.value && currentEditId.value) {
    const success = await store.updateStrategy(currentEditId.value)
    if (success) {
      showFormModal.value = false
    }
  } else {
    const success = await store.createStrategy()
    if (success) {
      showFormModal.value = false
    }
  }
}

const handleCloseModal = () => {
  showFormModal.value = false
  store.resetForm()
}

const isPageLoading = computed(() => store.loading)

onMounted(() => {
  store.fetchStrategies()
  timeframeStore.fetchTimeframes()
  indicatorStore.fetchIndicators()
  configStore.fetchConfigs()
})
</script>

<template>
  <DefaultLayout>
    <template #header-title>Strategies</template>

    <div class="">
      <!-- Loading State -->
      <div v-if="isPageLoading" class="flex items-center justify-center py-20">
        <div class="relative">
          <div class="animate-spin rounded-full h-16 w-16 border-b-2 border-primary"></div>
          <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2">
            <PhFlask :size="24" class="text-primary" />
          </div>
        </div>
      </div>

      <template v-else>
        <!-- Header Actions -->
        <div class="flex items-center justify-between mb-6">
          <div class="flex-1 max-w-md">
            <div class="relative">
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Search strategies..."
                class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              />
              <PhMagnifyingGlass
                :size="20"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
              />
            </div>
          </div>

          <button
            @click="handleCreate"
            class="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-all"
          >
            <PhPlus :size="20" weight="bold" />
            New Strategy
          </button>
        </div>

        <!-- Strategies Grid -->
        <div v-if="filteredStrategies.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div
            v-for="strategy in filteredStrategies"
            :key="strategy.id"
            class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6 hover:shadow-xl transition-all duration-200"
            :class="{ 'ring-2 ring-primary': strategy.is_active }"
          >
            <!-- Header -->
            <div class="flex items-start justify-between mb-4">
              <div class="flex-1">
                <div class="flex items-center gap-2 mb-2">
                  <h3 class="text-lg font-bold text-gray-900">{{ strategy.strategy_name }}</h3>
                  <span
                    v-if="strategy.is_active"
                    class="px-2 py-1 bg-blue-100 text-blue-700 text-xs font-medium rounded"
                  >
                    Active
                  </span>
                </div>
                <p class="text-sm text-gray-500">Primary: {{ strategy.primary_tf }}</p>
              </div>
            </div>

            <!-- Stats -->
            <div class="grid grid-cols-2 gap-3 mb-4 text-sm">
              <div>
                <p class="text-gray-500">Timeframes</p>
                <p class="font-semibold text-gray-900">{{ strategy.timeframes.length }}</p>
              </div>
              <div>
                <p class="text-gray-500">Indicators</p>
                <p class="font-semibold text-gray-900">{{ strategy.indicator_weights.length }}</p>
              </div>
              <div>
                <p class="text-gray-500">Leverage</p>
                <p class="font-semibold text-gray-900">{{ (strategy.money_management as any).leverage || '-' }}x</p>
              </div>
              <div>
                <p class="text-gray-500">Min Confidence</p>
                <p class="font-semibold text-gray-900">{{ (strategy.money_management as any).min_confidence || '-' }}%</p>
              </div>
            </div>

            <!-- Symbols -->
            <div v-if="(strategy as any).symbols?.length" class="mb-4">
              <p class="text-xs text-gray-500 mb-2">Symbols</p>
              <div class="flex flex-wrap gap-1.5">
                <span
                  v-for="sym in (strategy as any).symbols"
                  :key="sym.symbol"
                  class="px-2 py-0.5 text-xs font-medium rounded-full"
                  :class="sym.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'"
                >
                  {{ sym.symbol }}
                </span>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex items-center gap-2 pt-4 border-t border-gray-100">
              <button
                @click="handleEdit(strategy.id)"
                class="flex-1 flex items-center justify-center gap-1 px-3 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-all text-sm font-medium"
              >
                <PhPencilSimple :size="16" />
                Edit
              </button>

              <button
                @click="handleDelete(strategy.id, strategy.strategy_name)"
                class="flex items-center justify-center gap-1 px-3 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-all text-sm font-medium"
              >
                <PhTrash :size="16" />
              </button>
            </div>
          </div>
        </div>

        <!-- Empty State -->
        <div v-else class="text-center py-20">
          <PhFlask :size="64" class="mx-auto text-gray-300 mb-4" />
          <p class="text-gray-500 text-lg mb-2">
            {{ searchQuery ? 'No strategies found' : 'No strategies yet' }}
          </p>
          <p v-if="!searchQuery" class="text-gray-400 text-sm mb-4">
            Create your first trading strategy to get started
          </p>
          <button
            v-if="!searchQuery"
            @click="handleCreate"
            class="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-all"
          >
            <PhPlus :size="20" weight="bold" />
            Create Strategy
          </button>
        </div>
      </template>
    </div>

    <!-- Strategy Form Modal -->
    <StrategyFormModal
      v-model:show="showFormModal"
      :editing="editingStrategy"
      @submit="handleFormSubmit"
      @close="handleCloseModal"
    />
  </DefaultLayout>
</template>
