import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { type IApiResponse, get, post, put, del } from '@/lib/axios'
import { showSuccess, showError } from '@/lib/sweetalert'
import useVuelidate from '@vuelidate/core'
import { required, maxLength } from '@vuelidate/validators'
import Swal from 'sweetalert2'

const BASE_URL = '/watchlists'

// Interfaces
export interface IWatchlist {
  id: number
  symbol: string
  is_active: boolean
  created_at: string
}

export interface IWatchlistRequest {
  symbol: string
  is_active: boolean
}

export const useWatchlistStore = defineStore('watchlist', () => {
  // State
  const items = ref<IWatchlist[]>([])
  const loading = ref(false)
  const currentId = ref<number | null>(null)

  // Form State
  const watchlistReq = ref<IWatchlistRequest>({
    symbol: '',
    is_active: true
  })

  // Validation Rules
  const watchlistRules = ref({
    symbol: { required, maxLength: maxLength(20) },
    is_active: {}
  })

  // Vuelidate instance
  const watchlistReqValid = useVuelidate(watchlistRules, watchlistReq)

  // Getters
  const sortedItems = computed(() =>
    [...items.value].sort((a, b) => a.symbol.localeCompare(b.symbol))
  )

  const activeWatchlists = computed(() =>
    items.value.filter(w => w.is_active)
  )

  const inactiveWatchlists = computed(() =>
    items.value.filter(w => !w.is_active)
  )

  // Actions
  async function fetchWatchlists() {
    loading.value = true
    try {
      const response = await get<IApiResponse<IWatchlist[]>>(BASE_URL)
      items.value = response.data.data
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to fetch watchlists')
    } finally {
      loading.value = false
    }
  }

  async function createWatchlist(): Promise<boolean> {
    const valid = await watchlistReqValid.value.$validate()
    if (!valid) return false

    loading.value = true
    try {
      await post<IApiResponse<IWatchlist>>(BASE_URL, watchlistReq.value)
      showSuccess('Success', 'Watchlist created successfully')
      resetForm()
      fetchWatchlists()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to create watchlist')
      return false
    } finally {
      loading.value = false
    }
  }

  async function updateWatchlist(id: number): Promise<boolean> {
    const valid = await watchlistReqValid.value.$validate()
    if (!valid) return false

    loading.value = true
    try {
      await put<IApiResponse<IWatchlist>>(`${BASE_URL}/${id}`, watchlistReq.value)
      showSuccess('Success', 'Watchlist updated successfully')
      resetForm()
      currentId.value = null
      fetchWatchlists()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to update watchlist')
      return false
    } finally {
      loading.value = false
    }
  }

  async function deleteWatchlist(id: number): Promise<boolean> {
    // Show confirmation dialog
    const watchlist = items.value.find(w => w.id === id)
    const result = await Swal.fire({
      title: 'Delete Watchlist?',
      text: `Are you sure you want to delete "${watchlist?.symbol}"?`,
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
      await del<IApiResponse<IWatchlist>>(`${BASE_URL}/${id}`)
      showSuccess('Deleted', 'Watchlist deleted successfully')
      fetchWatchlists()
      return true
    } catch (error: any) {
      showError('Error', error.response?.data?.message || 'Failed to delete watchlist')
      return false
    } finally {
      loading.value = false
    }
  }

  function resetForm() {
    watchlistReq.value = {
      symbol: '',
      is_active: true
    }
    watchlistReqValid.value.$reset()
    currentId.value = null
  }

  function setEditMode(watchlist: IWatchlist) {
    currentId.value = watchlist.id
    watchlistReq.value = {
      symbol: watchlist.symbol,
      is_active: watchlist.is_active
    }
  }

  return {
    // State
    items,
    loading,
    currentId,
    watchlistReq,
    watchlistReqValid,

    // Getters
    sortedItems,
    activeWatchlists,
    inactiveWatchlists,

    // Actions
    fetchWatchlists,
    createWatchlist,
    updateWatchlist,
    deleteWatchlist,
    resetForm,
    setEditMode
  }
})
