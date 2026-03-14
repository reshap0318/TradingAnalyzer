<script setup lang="ts">
import { PhX } from '@phosphor-icons/vue'

defineProps<{
  modelValue: boolean
  title?: string
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
}>()

defineEmits<{
  'update:modelValue': [value: boolean]
}>()

// Map size to max-width class
const sizeClasses = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  full: 'max-w-4xl'
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="modelValue"
        class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4"
        @click.self="$emit('update:modelValue', false)"
      >
        <div
          :class="sizeClasses[size || 'md']"
          class="bg-white rounded-2xl shadow-2xl w-full max-h-[90vh] overflow-hidden flex flex-col
                 transform transition-all duration-300
                 modal-enter-from:opacity-0 modal-enter-from:scale-95 modal-enter-from:-translate-y-2
                 modal-leave-to:opacity-0 modal-leave-to:scale-95 modal-leave-to:-translate-y-2"
        >
          <!-- Header -->
          <header
            v-if="title"
            class="flex items-center justify-between px-6 py-5 border-b border-gray-200"
          >
            <h2 class="text-xl font-semibold text-gray-900">{{ title }}</h2>
            <button
              @click="$emit('update:modelValue', false)"
              class="text-gray-400 hover:text-gray-600 transition-colors p-1 rounded-lg hover:bg-gray-100"
              aria-label="Close"
            >
              <PhX :size="24" />
            </button>
          </header>

          <!-- Content -->
          <div class="px-6 py-5 overflow-y-auto">
            <slot />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
/* Modal Transitions - Only keep transition classes that Tailwind can't handle inline */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease;
}

.modal-enter-active .modal-container,
.modal-leave-active .modal-container {
  transition: transform 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  transform: scale(0.95) translateY(-10px);
}
</style>
