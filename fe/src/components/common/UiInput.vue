<script setup lang="ts">
interface IUiInputProps {
  modelValue: string
  type?: string
  label?: string
  placeholder?: string
  error?: boolean
  errorMessage?: string
  disabled?: boolean
  autocomplete?: string
}

const props = withDefaults(defineProps<IUiInputProps>(), {
  type: 'text',
  label: '',
  placeholder: '',
  error: false,
  errorMessage: '',
  disabled: false,
  autocomplete: 'off'
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <div>
    <label v-if="label" class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
      {{ label }}
    </label>
    <input
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :autocomplete="autocomplete"
      class="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent dark:bg-gray-700 dark:text-white transition-all"
      :class="{
        'border-danger': error,
        'border-gray-300 dark:border-gray-600': !error,
        'opacity-50 cursor-not-allowed': disabled
      }"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <p v-if="error && errorMessage" class="mt-1 text-sm text-danger">{{ errorMessage }}</p>
  </div>
</template>
