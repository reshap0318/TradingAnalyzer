import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { type IApiResponse, get, post, put, del } from '@/lib/axios'
import { showSuccess, showError } from '@/lib/sweetalert'
import useVuelidate from '@vuelidate/core'
import { required, maxLength } from '@vuelidate/validators'
import Swal from 'sweetalert2'

const BASE_URL = '/configs'

// Interfaces
export interface IConfig {
  id: number
  config_key: string
  value: string
  category: string
  created_at: string
}

export interface IConfigRequest {
  config_key: string
  value: string
  category: string
}

export const useConfigStore = defineStore('config', () => {
  // State
  const items = ref<IConfig[]>([])
  const loading = ref(false)
  const currentId = ref<number | null>(null)

  // Form State
  const configReq = ref<IConfigRequest>({
    config_key: '',
    value: '',
    category: ''
  })

  // Validation Rules
  const configRules = ref({
    config_key: { required, maxLength: maxLength(100) },
    value: { required },
    category: { required, maxLength: maxLength(50) }
  })

  // Vuelidate instance
  const configReqValid = useVuelidate(configRules, configReq)

  // Getters
  const sortedItems = computed(() =>
    [...items.value].sort((a, b) => {
      // Sort by category first, then by config_key
      if (a.category !== b.category) {
        return a.category.localeCompare(b.category)
      }
      return a.config_key.localeCompare(b.config_key)
    })
  )

  const groupedByCategory = computed(() => {
    const groups: Record<string, IConfig[]> = {}
    items.value.forEach(config => {
      if (!groups[config.category]) {
        groups[config.category] = []
      }
      groups[config.category]?.push(config)
    })
    return groups
  })

  const categories = computed(() => {
    return [...new Set(items.value.map(c => c.category))].sort()
  })

  // Actions
  async function fetchConfigs() {
    loading.value = true
    try {
      const response = await get<IApiResponse<IConfig[]>>(BASE_URL)
      items.value = response.data.data
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to fetch configs')
    } finally {
      loading.value = false
    }
  }

  async function createConfig(): Promise<boolean> {
    const valid = await configReqValid.value.$validate()
    if (!valid) return false

    loading.value = true
    try {
      await post<IApiResponse<IConfig>>(BASE_URL, configReq.value)
      showSuccess('Success', 'Config created successfully')
      resetForm()
      fetchConfigs()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to create config')
      return false
    } finally {
      loading.value = false
    }
  }

  async function updateConfig(id: number): Promise<boolean> {
    const valid = await configReqValid.value.$validate()
    if (!valid) return false

    loading.value = true
    try {
      await put<IApiResponse<IConfig>>(`${BASE_URL}/${id}`, configReq.value)
      showSuccess('Success', 'Config updated successfully')
      resetForm()
      currentId.value = null
      fetchConfigs()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to update config')
      return false
    } finally {
      loading.value = false
    }
  }

  async function deleteConfig(id: number): Promise<boolean> {
    // Show confirmation dialog
    const config = items.value.find(c => c.id === id)
    const result = await Swal.fire({
      title: 'Delete Config?',
      text: `Are you sure you want to delete config "${config?.config_key}"?`,
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
      await del<IApiResponse<IConfig>>(`${BASE_URL}/${id}`)
      showSuccess('Deleted', 'Config deleted successfully')
      fetchConfigs()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to delete config')
      return false
    } finally {
      loading.value = false
    }
  }

  function resetForm() {
    configReq.value = {
      config_key: '',
      value: '',
      category: ''
    }
    configReqValid.value.$reset()
    currentId.value = null
  }

  function setEditMode(config: IConfig) {
    currentId.value = config.id
    configReq.value = {
      config_key: config.config_key,
      value: config.value,
      category: config.category
    }
  }

  return {
    // State
    items,
    loading,
    currentId,
    configReq,
    configReqValid,

    // Getters
    sortedItems,
    groupedByCategory,
    categories,

    // Actions
    fetchConfigs,
    createConfig,
    updateConfig,
    deleteConfig,
    resetForm,
    setEditMode
  }
})
