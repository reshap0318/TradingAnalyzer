<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useBacktestStore } from '@/stores/backtest.store'
import { useStrategiesStore } from '@/stores/strategies.store'
import { PhX } from '@phosphor-icons/vue'
import { getValidationErrors } from '@/helpers/validation'
import useVuelidate from '@vuelidate/core'
import { required, minLength, minValue, maxValue } from '@vuelidate/validators'

interface IProps {
  show: boolean
}

const props = withDefaults(defineProps<IProps>(), {
  show: false
})

const emit = defineEmits<{
  submit: []
  close: []
}>()

const store = useBacktestStore()
const strategiesStore = useStrategiesStore()

// Local form state
const form = ref({
  name: '',
  symbol: 'BTCUSDT',
  strategy_id: 1,
  days: 30,
  capital: 1000
})

// Validation rules
const rules = computed(() => ({
  name: { required, minLength: minLength(3) },
  symbol: { required, minLength: minLength(3) },
  strategy_id: { required, minValue: minValue(1) },
  days: { required, minValue: minValue(1), maxValue: maxValue(30) },
  capital: { required, minValue: minValue(10) }
}))

const v$ = useVuelidate(rules, form)

const isSubmitting = computed(() => store.formLoading)

// Watch show prop
watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      // Reset form when modal opens
      form.value = {
        name: '',
        symbol: 'BTCUSDT',
        strategy_id: strategiesStore.strategies[0]?.id || 1,
        days: 30,
        capital: 1000
      }
      v$.value.$reset()
      strategiesStore.fetchStrategies()
    }
  },
  { immediate: true }
)

const handleSubmit = async () => {
  const valid = await v$.value.$validate()
  if (!valid) return

  // Copy form data to store
  store.createForm = { ...form.value }

  const backtestId = await store.createBacktest()
  if (backtestId) {
    emit('submit')
    emit('close')
  }
}

const handleClose = () => {
  emit('close')
}

// Handle ESC key
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    handleClose()
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click="handleClose">
        <div class="modal-container" @click.stop @keydown="handleKeydown" tabindex="0">
          <!-- Header -->
          <div class="modal-header">
            <h2 class="text-xl font-bold text-gray-900">Create New Backtest</h2>
            <button
              @click="handleClose"
              class="text-gray-400 hover:text-gray-600 transition-colors"
              aria-label="Close modal"
            >
              <PhX :size="24" />
            </button>
          </div>

          <!-- Body -->
          <div class="modal-body">
            <form class="space-y-4">
              <!-- Name -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">
                  Backtest Name <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="form.name"
                  type="text"
                  placeholder="e.g., BTC Backtest January 2024"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  :class="{ 'border-red-500': v$.name.$error }"
                />
                <p v-if="v$.name.$error" class="mt-1 text-sm text-red-500">
                  {{ getValidationErrors(v$.name).join(', ') }}
                </p>
              </div>

              <!-- Symbol -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">
                  Symbol <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="form.symbol"
                  type="text"
                  placeholder="e.g., BTCUSDT"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent uppercase"
                  :class="{ 'border-red-500': v$.symbol.$error }"
                />
                <p v-if="v$.symbol.$error" class="mt-1 text-sm text-red-500">
                  {{ getValidationErrors(v$.symbol).join(', ') }}
                </p>
              </div>

              <!-- Strategy -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">
                  Strategy <span class="text-red-500">*</span>
                </label>
                <select
                  v-model="form.strategy_id"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  :class="{ 'border-red-500': v$.strategy_id.$error }"
                >
                  <option
                    v-for="strategy in strategiesStore.strategies"
                    :key="strategy.id"
                    :value="strategy.id"
                  >
                    {{ strategy.strategy_name }}
                  </option>
                </select>
                <p v-if="v$.strategy_id.$error" class="mt-1 text-sm text-red-500">
                  {{ getValidationErrors(v$.strategy_id).join(', ') }}
                </p>
              </div>

              <!-- Days -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">
                  Backtest Period (Days) <span class="text-red-500">*</span>
                </label>
                <input
                  v-model.number="form.days"
                  type="number"
                  min="1"
                  max="30"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  :class="{ 'border-red-500': v$.days.$error }"
                />
                <p v-if="v$.days.$error" class="mt-1 text-sm text-red-500">
                  {{ getValidationErrors(v$.days).join(', ') }}
                </p>
                <p class="mt-1 text-xs text-gray-500">Maximum 30 days</p>
              </div>

              <!-- Capital -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">
                  Initial Capital (USDT) <span class="text-red-500">*</span>
                </label>
                <input
                  v-model.number="form.capital"
                  type="number"
                  min="10"
                  step="10"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  :class="{ 'border-red-500': v$.capital.$error }"
                />
                <p v-if="v$.capital.$error" class="mt-1 text-sm text-red-500">
                  {{ getValidationErrors(v$.capital).join(', ') }}
                </p>
                <p class="mt-1 text-xs text-gray-500">Minimum 10 USDT</p>
              </div>
            </form>
          </div>

          <!-- Footer -->
          <div class="modal-footer">
            <button
              type="button"
              @click="handleClose"
              class="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
              :disabled="isSubmitting"
            >
              Cancel
            </button>
            <button
              type="button"
              @click="handleSubmit"
              class="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors flex items-center gap-2"
              :disabled="isSubmitting"
            >
              <span v-if="isSubmitting" class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></span>
              {{ isSubmitting ? 'Starting...' : 'Start Backtest' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
/* Overlay */
.modal-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 1rem;
}

/* Modal Container */
.modal-container {
  background: white;
  border-radius: 1rem;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  max-width: 32rem;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  outline: none;
}

/* Header */
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem 2rem;
  border-bottom: 1px solid #e5e7eb;
}

/* Body */
.modal-body {
  padding: 2rem;
}

/* Footer */
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.5rem 2rem;
  border-top: 1px solid #e5e7eb;
}

/* Transitions */
.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  opacity: 0;
  transform: scale(0.95) translateY(-10px);
}

.modal-enter-from .modal-overlay,
.modal-leave-to .modal-overlay {
  opacity: 0;
}
</style>
