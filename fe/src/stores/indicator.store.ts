import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { type IApiResponse, get, post, put, del } from '@/lib/axios'
import { showSuccess, showError } from '@/lib/sweetalert'
import useVuelidate from '@vuelidate/core'
import { required, minValue, maxValue } from '@vuelidate/validators'
import Swal from 'sweetalert2'

const BASE_URL = '/indicators'

// Interfaces
export interface IIndicator {
  id: number
  name: string
  indicator: string
  description: string
  params: Record<string, any>
  is_active: boolean
  weight: number
  order_view: number
  created_at: string
}

export interface IIndicatorRequest {
  name: string
  indicator: string
  description: string
  params: Record<string, any>
  is_active: boolean
  weight: number
  order_view: number
}

export const useIndicatorStore = defineStore('indicator', () => {
  // State
  const items = ref<IIndicator[]>([])
  const loading = ref(false)
  const currentId = ref<number | null>(null)

  // Form State
  const indicatorReq = ref<IIndicatorRequest>({
    name: '',
    indicator: '',
    description: '',
    params: {},
    is_active: true,
    weight: 1.0,
    order_view: 1
  })

  // JSON Editor State
  const paramsJson = ref<string>('{}')
  const paramsJsonError = ref<string | null>(null)
  const paramsJsonValid = ref<boolean>(true)

  // Validation Rules
  const indicatorRules = ref({
    name: { required },
    indicator: { required },
    weight: { required, minValue: minValue(0), maxValue: maxValue(1) },
    order_view: { required, minValue: minValue(1) }
  })

  // Vuelidate instance
  const indicatorReqValid = useVuelidate(indicatorRules, indicatorReq)

  // Getters
  const sortedItems = computed(() =>
    [...items.value].sort((a, b) => a.order_view - b.order_view)
  )

  const activeIndicators = computed(() =>
    items.value.filter(i => i.is_active)
  )

  // Actions
  async function fetchIndicators() {
    loading.value = true
    try {
      const response = await get<IApiResponse<IIndicator[]>>(BASE_URL)
      items.value = response.data.data
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to fetch indicators')
    } finally {
      loading.value = false
    }
  }

  function validateParamsJson(): boolean {
    try {
      JSON.parse(paramsJson.value)
      paramsJsonValid.value = true
      paramsJsonError.value = null
      return true
    } catch (error: any) {
      paramsJsonValid.value = false
      paramsJsonError.value = error.message
      return false
    }
  }

  function formatParamsJson() {
    try {
      const parsed = JSON.parse(paramsJson.value)
      paramsJson.value = JSON.stringify(parsed, null, 2)
      paramsJsonError.value = null
      paramsJsonValid.value = true
    } catch (error: any) {
      paramsJsonError.value = 'Invalid JSON: ' + error.message
    }
  }

  function loadParamsToEditor(params: Record<string, any>) {
    paramsJson.value = JSON.stringify(params, null, 2)
    paramsJsonError.value = null
    paramsJsonValid.value = true
  }

  async function createIndicator(): Promise<boolean> {
    // Validate basic fields
    const valid = await indicatorReqValid.value.$validate()
    if (!valid) return false

    // Validate JSON params
    if (!validateParamsJson()) {
      showError('Invalid JSON', 'Please fix the JSON parameters first')
      return false
    }

    // Parse params
    indicatorReq.value.params = JSON.parse(paramsJson.value)

    loading.value = true
    try {
      await post<IApiResponse<IIndicator>>(BASE_URL, indicatorReq.value)
      showSuccess('Success', 'Indicator created successfully')
      resetForm()
      fetchIndicators()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to create indicator')
      return false
    } finally {
      loading.value = false
    }
  }

  async function updateIndicator(id: number): Promise<boolean> {
    // Validate basic fields
    const valid = await indicatorReqValid.value.$validate()
    if (!valid) return false

    // Validate JSON params
    if (!validateParamsJson()) {
      showError('Invalid JSON', 'Please fix the JSON parameters first')
      return false
    }

    // Parse params
    indicatorReq.value.params = JSON.parse(paramsJson.value)

    loading.value = true
    try {
      await put<IApiResponse<IIndicator>>(`${BASE_URL}/${id}`, indicatorReq.value)
      showSuccess('Success', 'Indicator updated successfully')
      resetForm()
      currentId.value = null
      fetchIndicators()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to update indicator')
      return false
    } finally {
      loading.value = false
    }
  }

  async function deleteIndicator(id: number): Promise<boolean> {
    // Show confirmation dialog
    const indicator = items.value.find(i => i.id === id)
    const result = await Swal.fire({
      title: 'Delete Indicator?',
      text: `Are you sure you want to delete "${indicator?.name}"?`,
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
      await del<IApiResponse<IIndicator>>(`${BASE_URL}/${id}`)
      showSuccess('Deleted', 'Indicator deleted successfully')
      fetchIndicators()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to delete indicator')
      return false
    } finally {
      loading.value = false
    }
  }

  function resetForm() {
    indicatorReq.value = {
      name: '',
      indicator: '',
      description: '',
      params: {},
      is_active: true,
      weight: 1.0,
      order_view: 1
    }
    paramsJson.value = '{}'
    paramsJsonError.value = null
    paramsJsonValid.value = true
    currentId.value = null
    indicatorReqValid.value.$reset()
  }

  function setEditMode(indicator: IIndicator) {
    currentId.value = indicator.id
    indicatorReq.value = {
      name: indicator.name,
      indicator: indicator.indicator,
      description: indicator.description,
      params: indicator.params,
      is_active: indicator.is_active,
      weight: indicator.weight,
      order_view: indicator.order_view
    }
    loadParamsToEditor(indicator.params)
  }

  return {
    // State
    items,
    loading,
    currentId,
    indicatorReq,
    indicatorReqValid,
    paramsJson,
    paramsJsonError,
    paramsJsonValid,

    // Getters
    sortedItems,
    activeIndicators,

    // Actions
    fetchIndicators,
    createIndicator,
    updateIndicator,
    deleteIndicator,
    resetForm,
    setEditMode,
    validateParamsJson,
    formatParamsJson,
    loadParamsToEditor
  }
})
