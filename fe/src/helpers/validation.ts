import type { Validation } from '@vuelidate/core'

/**
 * Get all validation error messages from a field validation
 * @param fieldValidation - Vuelidate field validation object
 * @returns Array of error messages
 */
export function getValidationErrors(fieldValidation: Validation): string[] {
  if (!fieldValidation?.$errors) return []
  return fieldValidation.$errors.map((e) => e.$message.toString()).filter(Boolean)
}

/**
 * Get the first validation error message
 * @param fieldValidation - Vuelidate field validation object
 * @returns First error message or empty string
 */
export function getFirstError(fieldValidation: Validation): string {
  const errors = getValidationErrors(fieldValidation)
  return errors[0] || ''
}

/**
 * Check if field has validation errors
 * @param fieldValidation - Vuelidate field validation object
 * @returns true if has errors
 */
export function hasValidationErrors(fieldValidation: Validation): boolean {
  return fieldValidation?.$error ?? false
}
