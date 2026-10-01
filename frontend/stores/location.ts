// Location (Branch) Store - Pinia State Management
//
// Branch master data: the backend serves /api/v1/locations as a flat list
// (no pagination), so search and paging are done client-side here.
import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'
import type { TableColumn } from '@nuxt/ui'
import type {
  Location,
  LocationFormState,
  CreateLocationRequest,
  UpdateLocationRequest
} from '~/types/master'
import { getUserErrorMessage } from '~/utils/error'
import { useI18n } from '~/composables/useI18n'
import { useLocationApi } from '~/composables/useLocationApi'

export const useLocationStore = defineStore('location', () => {
  // ============= State =============
  const locations = ref<Location[]>([])
  const loading = ref(false)
  const errorMsg = ref('')
  const { t } = useI18n()

  // Modal State
  const showModal = ref(false)
  const isEditing = ref(false)
  const editingId = ref<string | null>(null)

  const search = ref('')

  // Form State
  const form = reactive<LocationFormState>({
    name: '',
    address: '',
    city: '',
    province: '',
    postal_code: '',
    country: 'Indonesia',
    is_active: true
  })

  // Table Columns
  const columns: TableColumn<Location>[] = [
    { accessorKey: 'name', header: t('masterData.location.columns.name') },
    { accessorKey: 'city', header: t('masterData.location.columns.city') },
    { accessorKey: 'province', header: t('masterData.location.columns.province') },
    { accessorKey: 'address', header: t('masterData.location.columns.address') },
    { accessorKey: 'is_active', header: t('masterData.location.columns.status') },
    { accessorKey: 'actions', header: '' }
  ]

  // ============= Getters =============
  const filteredLocations = computed(() => {
    const q = search.value.trim().toLowerCase()
    if (!q) return locations.value
    return locations.value.filter((l) =>
      [l.name, l.city, l.province, l.address].some((field) =>
        (field || '').toLowerCase().includes(q)
      )
    )
  })

  const total = computed(() => filteredLocations.value.length)

  /** Branch names for risk/audit filters that group by branch. */
  const branchNames = computed(() =>
    locations.value.filter((l) => l.is_active).map((l) => l.name)
  )

  // ============= Actions =============

  const fetchLocations = async () => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useLocationApi()
      locations.value = await api.getLocations()
    } catch (error: any) {
      console.error('Failed to fetch locations:', error)
      errorMsg.value = getUserErrorMessage(error, t, { fallbackKey: 'masterData.errors.fetchLocations' })
      locations.value = []
    } finally {
      loading.value = false
    }
  }

  const createLocation = async (): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useLocationApi()
      const payload: CreateLocationRequest = {
        name: form.name,
        address: form.address,
        city: form.city,
        province: form.province || undefined,
        postal_code: form.postal_code || undefined,
        country: form.country || 'Indonesia',
        is_active: form.is_active
      }

      await api.createLocation(payload)
      await fetchLocations()
      return true
    } catch (error: any) {
      console.error('Failed to create location:', error)
      errorMsg.value = getUserErrorMessage(error, t, { fallbackKey: 'masterData.errors.createLocation' })
      return false
    } finally {
      loading.value = false
    }
  }

  const updateLocation = async (id: string): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useLocationApi()
      const payload: UpdateLocationRequest = {
        name: form.name || undefined,
        address: form.address || undefined,
        city: form.city || undefined,
        province: form.province || undefined,
        postal_code: form.postal_code || undefined,
        country: form.country || undefined,
        is_active: form.is_active
      }

      await api.updateLocation(id, payload)
      await fetchLocations()
      return true
    } catch (error: any) {
      console.error('Failed to update location:', error)
      errorMsg.value = getUserErrorMessage(error, t, { fallbackKey: 'masterData.errors.updateLocation' })
      return false
    } finally {
      loading.value = false
    }
  }

  const deleteLocation = async (id: string): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useLocationApi()
      await api.deleteLocation(id)
      await fetchLocations()
      return true
    } catch (error: any) {
      console.error('Failed to delete location:', error)
      errorMsg.value = getUserErrorMessage(error, t, { fallbackKey: 'masterData.errors.deleteLocation' })
      return false
    } finally {
      loading.value = false
    }
  }

  // ============= UI Actions =============

  const resetForm = () => {
    form.name = ''
    form.address = ''
    form.city = ''
    form.province = ''
    form.postal_code = ''
    form.country = 'Indonesia'
    form.is_active = true
  }

  const openCreateModal = () => {
    isEditing.value = false
    editingId.value = null
    resetForm()
    showModal.value = true
  }

  const handleEdit = (location: Location) => {
    isEditing.value = true
    editingId.value = location.id

    form.name = location.name
    form.address = location.address || ''
    form.city = location.city || ''
    form.province = location.province || ''
    form.postal_code = location.postal_code || ''
    form.country = location.country || 'Indonesia'
    form.is_active = location.is_active

    showModal.value = true
  }

  const closeModal = () => {
    showModal.value = false
    isEditing.value = false
    editingId.value = null
    errorMsg.value = ''
    resetForm()
  }

  const handleSubmit = async (): Promise<boolean> => {
    if (!form.name.trim()) {
      errorMsg.value = t('masterData.validation.locationNameRequired')
      return false
    }
    if (!form.address.trim()) {
      errorMsg.value = t('masterData.validation.locationAddressRequired')
      return false
    }
    if (!form.city.trim()) {
      errorMsg.value = t('masterData.validation.locationCityRequired')
      return false
    }

    const success = isEditing.value && editingId.value
      ? await updateLocation(editingId.value)
      : await createLocation()

    if (success) closeModal()
    return success
  }

  const handleDelete = async (location: Location): Promise<boolean> => {
    const confirmed = await useGlobalModalStore().confirmDelete({
      description: t('masterData.location.deleteConfirm', { name: location.name })
    })
    if (!confirmed) return false

    return await deleteLocation(location.id)
  }

  const setSearch = (value: string) => {
    search.value = value
  }

  // ============= Return =============
  return {
    // State
    locations,
    loading,
    errorMsg,
    showModal,
    isEditing,
    editingId,
    search,
    form,
    columns,

    // Getters
    filteredLocations,
    total,
    branchNames,

    // Actions
    fetchLocations,
    createLocation,
    updateLocation,
    deleteLocation,
    openCreateModal,
    handleEdit,
    handleDelete,
    closeModal,
    handleSubmit,
    resetForm,
    setSearch
  }
})
