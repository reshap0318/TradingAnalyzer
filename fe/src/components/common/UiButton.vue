<script setup lang="ts">
interface IUiButtonProps {
  type?: 'button' | 'submit' | 'reset'
  variant?: 'primary' | 'danger' | 'outline'
  loading?: boolean
  disabled?: boolean
  fullWidth?: boolean
}

const props = withDefaults(defineProps<IUiButtonProps>(), {
  type: 'button',
  variant: 'primary',
  loading: false,
  disabled: false,
  fullWidth: false
})

const emit = defineEmits<{
  click: []
}>()

const variantClasses = {
  primary: 'bg-primary hover:bg-primary-dark text-white',
  danger: 'bg-danger hover:bg-danger-dark text-white',
  outline: 'border-2 border-primary text-primary hover:bg-primary hover:text-white'
}
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    class="px-6 py-3 font-semibold rounded-lg transition-all duration-200 flex items-center justify-center disabled:cursor-not-allowed"
    :class="[
      variantClasses[variant],
      { 'w-full': fullWidth },
      { 'opacity-50': disabled || loading }
    ]"
    @click="emit('click')"
  >
    <svg
      v-if="loading"
      class="animate-spin -ml-1 mr-3 h-5 w-5"
      :class="{ 'text-white': variant !== 'outline' }"
      fill="none"
      viewBox="0 0 24 24"
    >
      <circle
        class="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        stroke-width="4"
      />
      <path
        class="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      />
    </svg>
    <slot />
  </button>
</template>
