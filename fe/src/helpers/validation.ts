/**
 * Global error messages for validators
 * Override this object to customize error messages globally
 */
export const validationMessages: Record<string, string> = {
  required: 'Wajib diisi',
  email: 'Format email tidak valid',
  minLength: 'Panjang karakter minimal {min}',
  maxLength: 'Panjang karakter maksimal {max}',
  minValue: 'Nilai minimal {min}',
  maxValue: 'Nilai maksimal {max}',
  between: 'Nilai harus antara {min} dan {max}',
  alpha: 'Hanya boleh mengandung huruf',
  alphaNum: 'Hanya boleh mengandung huruf dan angka',
  numeric: 'Hanya boleh mengandung angka',
  url: 'Format URL tidak valid',
  sameAs: 'Harus sama dengan {other}'
}

/**
 * Generate error messages from Vuelidate validation result
 * @param fieldValidation - Validation result for a specific field (e.g., v$.username)
 * @returns Array of error messages
 * 
 * @example
 * ```typescript
 * // Usage in component
 * const errors = getValidationErrors(v$.username)
 * 
 * // Usage in template
 * <p v-if="getValidationErrors(v$.username).length > 0">
 *   {{ getValidationErrors(v$.username)[0] }}
 * </p>
 * ```
 */
export function getValidationErrors(fieldValidation: any): string[] {
  if (!fieldValidation?.$errors || fieldValidation.$errors.length === 0) {
    return []
  }

  return fieldValidation.$errors.map((error: any) => {
    const validatorName = error.$validator
    const message = validationMessages[validatorName as keyof typeof validationMessages]

    if (!message) {
      // Return default message if validator not found
      return error.$message || 'Validasi gagal'
    }

    // Replace placeholders with actual values
    let formattedMessage = message
    const params = error.$params as Record<string, any>

    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        // Handle Ref types by accessing .value
        const paramValue = value?.value ?? value
        formattedMessage = formattedMessage.replace(`{${key}}`, String(paramValue))
      })
    }

    return formattedMessage
  })
}

/**
 * Get first error message from validation result
 * @param fieldValidation - Validation result for a specific field
 * @returns First error message or empty string
 * 
 * @example
 * ```typescript
 * const firstError = getFirstError(v$.username)
 * ```
 */
export function getFirstError(fieldValidation: any): string {
  const errors = getValidationErrors(fieldValidation)
  return errors[0] || ''
}

/**
 * Check if field has validation errors
 * @param fieldValidation - Validation result for a specific field
 * @returns true if has errors
 * 
 * @example
 * ```typescript
 * const hasError = hasValidationErrors(v$.username)
 * ```
 */
export function hasValidationErrors(fieldValidation: any): boolean {
  return fieldValidation?.$errors?.length > 0
}

export default {
  validationMessages,
  getValidationErrors,
  getFirstError,
  hasValidationErrors
}
