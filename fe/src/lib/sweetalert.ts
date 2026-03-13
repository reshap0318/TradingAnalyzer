import Swal, { type SweetAlertOptions, type SweetAlertResult } from 'sweetalert2'
import 'sweetalert2/dist/sweetalert2.css'

/**
 * SweetAlert2 Wrapper with TailwindCSS Integration
 * Pre-configured with project theme
 */

// Default configuration with Tailwind-compatible styles
const defaultConfig: SweetAlertOptions = {
  confirmButtonColor: '#3b82f6', // primary
  cancelButtonColor: '#6b7280', // gray-500
  confirmButtonText: 'Confirm',
  cancelButtonText: 'Cancel',
  showCancelButton: false,
  reverseButtons: true,
  width: '32rem', // w-96
  padding: '1.5rem', // p-6
  backdrop: `
    rgba(0, 0, 0, 0.5)
  `,
  customClass: {
    popup: 'rounded-xl shadow-2xl',
    container: 'backdrop-blur-sm',
    // header: 'mb-4',
    title: 'text-xl font-bold text-gray-900 dark:text-gray-100',
    htmlContainer: 'text-gray-600 dark:text-gray-300',
    closeButton: 'hover:bg-gray-100 dark:hover:bg-gray-800',
    icon: 'mb-4',
    input:
      'border border-gray-300 dark:border-gray-600 rounded-lg px-4 py-2 focus:ring-2 focus:ring-primary focus:border-transparent',
    inputLabel: 'text-gray-700 dark:text-gray-300 mb-2',
    validationMessage: 'text-danger mt-2',
    actions: 'flex gap-2 mt-6',
    confirmButton:
      'px-6 py-2.5 rounded-lg font-semibold transition-all hover:opacity-90 active:scale-95',
    cancelButton:
      'px-6 py-2.5 rounded-lg font-semibold transition-all hover:opacity-90 active:scale-95 ml-2',
    loader: 'text-primary'
  },
  buttonsStyling: true
}

/**
 * Show success alert
 */
export const showSuccess = (title: string, text?: string): Promise<SweetAlertResult> => {
  return Swal.fire({
    ...defaultConfig,
    icon: 'success',
    title,
    text,
    timer: 3000,
    showConfirmButton: false,
    iconColor: '#22c55e' // success
  })
}

/**
 * Show error alert
 */
export const showError = (title: string, text?: string): Promise<SweetAlertResult> => {
  return Swal.fire({
    ...defaultConfig,
    icon: 'error',
    title,
    text,
    confirmButtonText: 'Close',
    iconColor: '#ef4444' // danger
  })
}

/**
 * Show warning alert
 */
export const showWarning = (title: string, text?: string): Promise<SweetAlertResult> => {
  return Swal.fire({
    ...defaultConfig,
    icon: 'warning',
    title,
    text,
    confirmButtonColor: '#f59e0b', // warning
    iconColor: '#f59e0b'
  })
}

/**
 * Show info alert
 */
export const showInfo = (title: string, text?: string): Promise<SweetAlertResult> => {
  return Swal.fire({
    ...defaultConfig,
    icon: 'info',
    title,
    text,
    confirmButtonColor: '#06b6d4', // info
    iconColor: '#06b6d4'
  })
}

/**
 * Show confirmation dialog
 */
export const showConfirm = (
  title: string,
  text?: string,
  confirmText: string = 'Yes',
  cancelText: string = 'No'
): Promise<SweetAlertResult> => {
  return Swal.fire({
    ...defaultConfig,
    icon: 'question',
    title,
    text,
    showCancelButton: true,
    confirmButtonText: confirmText,
    cancelButtonText: cancelText,
    confirmButtonColor: '#2563eb', // primary-dark
    cancelButtonColor: '#6b7280', // gray-500
    iconColor: '#6b7280'
  })
}

/**
 * Show custom alert with HTML content
 */
export const showCustom = (options: SweetAlertOptions): Promise<SweetAlertResult> => {
  return Swal.fire({
    ...defaultConfig,
    ...options
  } as SweetAlertOptions)
}

/**
 * Show loading toast
 */
export const showLoading = (title: string): Promise<SweetAlertResult> => {
  return Swal.fire({
    ...defaultConfig,
    title,
    allowOutsideClick: false,
    allowEscapeKey: false,
    didOpen: () => {
      Swal.showLoading()
    }
  })
}

/**
 * Close current alert
 */
export const close = () => {
  Swal.close()
}

/**
 * Update loading title
 */
export const updateLoadingTitle = (title: string) => {
  const titleElement = Swal.getTitle()
  if (titleElement) {
    titleElement.textContent = title
  }
}

export default {
  showSuccess,
  showError,
  showWarning,
  showInfo,
  showConfirm,
  showCustom,
  showLoading,
  close,
  updateLoadingTitle
}
