import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { type IApiResponse, get, post, put, del } from '@/lib/axios'
import { showSuccess, showError } from '@/lib/sweetalert'
import useVuelidate from '@vuelidate/core'
import { required, maxLength, minValue } from '@vuelidate/validators'
import Swal from 'sweetalert2'

const BASE_URL = '/timeframes'

// Interfaces
export interface ITimeframe {
  name: string
  in_minutes: number
  created_at: string
}

export interface ITimeframeRequest {
  name: string
  in_minutes: number
}

export const useTimeframeStore = defineStore('timeframe', () => {
  // State
  const items = ref<ITimeframe[]>([])
  const loading = ref(false)
  const currentName = ref<string | null>(null)

  // Form State
  const timeframeReq = ref<ITimeframeRequest>({
    name: '',
    in_minutes: 1
  })

  // Validation Rules
  const timeframeRules = ref({
    name: { required, maxLength: maxLength(5) },
    in_minutes: { required, minValue: minValue(1) }
  })

  // Vuelidate instance
  const timeframeReqValid = useVuelidate(timeframeRules, timeframeReq)

  // Getters
  const sortedItems = computed(() =>
    [...items.value].sort((a, b) => a.in_minutes - b.in_minutes)
  )

  // Actions
  async function fetchTimeframes() {
    loading.value = true
    try {
      const response = await get<IApiResponse<ITimeframe[]>>(BASE_URL)
      items.value = response.data.data
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to fetch timeframes')
    } finally {
      loading.value = false
    }
  }

  async function createTimeframe(): Promise<boolean> {
    const valid = await timeframeReqValid.value.$validate()
    if (!valid) return false

    loading.value = true
    try {
      await post<IApiResponse<ITimeframe>>(BASE_URL, timeframeReq.value)
      showSuccess('Success', 'Timeframe created successfully')
      resetForm()
      fetchTimeframes()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to create timeframe')
      return false
    } finally {
      loading.value = false
    }
  }

  async function updateTimeframe(name: string): Promise<boolean> {
    const valid = await timeframeReqValid.value.$validate()
    if (!valid) return false

    loading.value = true
    try {
      await put<IApiResponse<ITimeframe>>(`${BASE_URL}/${name}`, timeframeReq.value)
      showSuccess('Success', 'Timeframe updated successfully')
      resetForm()
      currentName.value = null
      fetchTimeframes()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to update timeframe')
      return false
    } finally {
      loading.value = false
    }
  }

  async function deleteTimeframe(name: string): Promise<boolean> {
    // Show confirmation dialog
    const result = await Swal.fire({
      title: 'Delete Timeframe?',
      text: `Are you sure you want to delete timeframe "${name}"?`,
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
      await del<IApiResponse<ITimeframe>>(`${BASE_URL}/${name}`)
      showSuccess('Deleted', 'Timeframe deleted successfully')
      fetchTimeframes()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to delete timeframe')
      return false
    } finally {
      loading.value = false
    }
  }

  function resetForm() {
    timeframeReq.value = { name: '', in_minutes: 1 }
    timeframeReqValid.value.$reset()
    currentName.value = null
  }

  function setEditMode(timeframe: ITimeframe) {
    currentName.value = timeframe.name
    timeframeReq.value = {
      name: timeframe.name,
      in_minutes: timeframe.in_minutes
    }
  }

  return {
    // State
    items,
    loading,
    currentName,
    timeframeReq,
    timeframeReqValid,

    // Getters
    sortedItems,

    // Actions
    fetchTimeframes,
    createTimeframe,
    updateTimeframe,
    deleteTimeframe,
    resetForm,
    setEditMode
  }
})
