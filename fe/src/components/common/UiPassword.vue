<script setup lang="ts">
import { ref, computed } from 'vue'

interface IUiPasswordProps {
  modelValue: string
  label?: string
  placeholder?: string
  error?: boolean
  errorMessage?: string
  disabled?: boolean
  autocomplete?: string
}

const props = withDefaults(defineProps<IUiPasswordProps>(), {
  label: '',
  placeholder: '',
  error: false,
  errorMessage: '',
  disabled: false,
  autocomplete: 'current-password'
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const showPassword = ref(false)

const inputType = computed(() => (showPassword.value ? 'text' : 'password'))

const togglePassword = () => {
  showPassword.value = !showPassword.value
}
</script>

<template>
  <div>
    <label v-if="label" class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
      {{ label }}
    </label>
    <div class="relative">
      <input
        :type="inputType"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :autocomplete="autocomplete"
        class="w-full px-4 py-3 pr-12 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent dark:bg-gray-700 dark:text-white transition-all"
        :class="{
          'border-danger': error,
          'border-gray-300 dark:border-gray-600': !error,
          'opacity-50 cursor-not-allowed': disabled
        }"
        @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      />
      <button
        type="button"
        @click="togglePassword"
        class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
        :disabled="disabled"
        tabindex="-1"
      >
        <!-- Eye Open Icon -->
        <svg
          v-if="showPassword"
          class="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"
          />
        </svg>
        <!-- Eye Closed Icon -->
        <svg
          v-else
          class="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
          />
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
          />
        </svg>
      </button>
    </div>
    <p v-if="error && errorMessage" class="mt-1 text-sm text-danger">{{ errorMessage }}</p>
  </div>
</template>
