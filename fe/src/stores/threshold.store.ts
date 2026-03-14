import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { type IApiResponse, get, post, put, del } from '@/lib/axios'
import { showSuccess, showError } from '@/lib/sweetalert'
import useVuelidate from '@vuelidate/core'
import { required, minValue, maxValue, integer } from '@vuelidate/validators'
import Swal from 'sweetalert2'

const BASE_URL = '/thresholds'

// Interfaces
export interface IThreshold {
  id: number
  category: string
  min_value: number
  max_value: number
  action: 'BUY' | 'SELL' | 'WAIT'
  color: string
  order_display: number
  created_at: string
}

export interface IThresholdRequest {
  category: string
  min_value: number
  max_value: number
  action: 'BUY' | 'SELL' | 'WAIT'
  color: string
  order_display: number
}

export const useThresholdStore = defineStore('threshold', () => {
  // State
  const items = ref<IThreshold[]>([])
  const loading = ref(false)
  const currentId = ref<number | null>(null)

  // Form State
  const thresholdReq = ref<IThresholdRequest>({
    category: '',
    min_value: -100,
    max_value: 100,
    action: 'WAIT',
    color: 'gray',
    order_display: 1
  })

  // Validation Rules
  const thresholdRules = ref({
    category: { required },
    min_value: { required, minValue: minValue(-100), maxValue: maxValue(100), integer },
    max_value: { required, minValue: minValue(-100), maxValue: maxValue(100), integer },
    action: { required },
    color: { required },
    order_display: { required, minValue: minValue(1), integer }
  })

  // Vuelidate instance
  const thresholdReqValid = useVuelidate(thresholdRules, thresholdReq)

  // Getters
  const sortedItems = computed(() =>
    [...items.value].sort((a, b) => a.order_display - b.order_display)
  )

  const sortedByValue = computed(() =>
    [...items.value].sort((a, b) => a.min_value - b.min_value)
  )

  // Actions
  async function fetchThresholds() {
    loading.value = true
    try {
      const response = await get<IApiResponse<IThreshold[]>>(BASE_URL)
      items.value = response.data.data
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to fetch thresholds')
    } finally {
      loading.value = false
    }
  }

  async function createThreshold(): Promise<boolean> {
    const valid = await thresholdReqValid.value.$validate()
    if (!valid) return false

    // Validate min < max
    if (thresholdReq.value.min_value >= thresholdReq.value.max_value) {
      showError('Validation Error', 'Min value must be less than max value')
      return false
    }

    loading.value = true
    try {
      await post<IApiResponse<IThreshold>>(BASE_URL, thresholdReq.value)
      showSuccess('Success', 'Threshold created successfully')
      resetForm()
      fetchThresholds()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to create threshold')
      return false
    } finally {
      loading.value = false
    }
  }

  async function updateThreshold(id: number): Promise<boolean> {
    const valid = await thresholdReqValid.value.$validate()
    if (!valid) return false

    // Validate min < max
    if (thresholdReq.value.min_value >= thresholdReq.value.max_value) {
      showError('Validation Error', 'Min value must be less than max value')
      return false
    }

    loading.value = true
    try {
      await put<IApiResponse<IThreshold>>(`${BASE_URL}/${id}`, thresholdReq.value)
      showSuccess('Success', 'Threshold updated successfully')
      resetForm()
      currentId.value = null
      fetchThresholds()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to update threshold')
      return false
    } finally {
      loading.value = false
    }
  }

  async function deleteThreshold(id: number): Promise<boolean> {
    // Show confirmation dialog
    const threshold = items.value.find(t => t.id === id)
    const result = await Swal.fire({
      title: 'Delete Threshold?',
      text: `Are you sure you want to delete threshold "${threshold?.category}"?`,
      icon: 'warning',
      showCancelButton: true,
      confirmButtonColor: '#ef4444',
      cancelButtonColor: '#6b7280',
      confirmButtonText: 'Yes, delete it!',
      cancelButtonText: 'Cancel'
    })

    if (!result.isConfirmed) {
      return false
    }

    loading.value = true
    try {
      await del<IApiResponse<IThreshold>>(`${BASE_URL}/${id}`)
      showSuccess('Deleted', 'Threshold deleted successfully')
      fetchThresholds()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to delete threshold')
      return false
    } finally {
      loading.value = false
    }
  }

  function resetForm() {
    thresholdReq.value = {
      category: '',
      min_value: -100,
      max_value: 100,
      action: 'WAIT',
      color: 'gray',
      order_display: 1
    }
    thresholdReqValid.value.$reset()
    currentId.value = null
  }

  function setEditMode(threshold: IThreshold) {
    currentId.value = threshold.id
    thresholdReq.value = {
      category: threshold.category,
      min_value: threshold.min_value,
      max_value: threshold.max_value,
      action: threshold.action,
      color: threshold.color,
      order_display: threshold.order_display
    }
  }

  return {
    // State
    items,
    loading,
    currentId,
    thresholdReq,
    thresholdReqValid,

    // Getters
    sortedItems,
    sortedByValue,

    // Actions
    fetchThresholds,
    createThreshold,
    updateThreshold,
    deleteThreshold,
    resetForm,
    setEditMode
  }
})
