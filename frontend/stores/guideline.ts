import type { TableColumn } from '@nuxt/ui'
import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { useI18n } from '~/composables/useI18n'
import { extractErrorMessage } from '~/utils/error'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'

export interface AuditGuideline {
  id: string
  name: string
  status: string // "Aktif" | "Sedang Diperbarui"
  effective_date: string // "YYYY-MM"
  file_url: string
  file_name: string
  file_size: number
  created_at?: string
  updated_at?: string
}

export const useGuidelineStore = defineStore('guideline', () => {
  const guidelines = ref<AuditGuideline[]>([])
  const loading = ref(false)
  const errorMsg = ref('')
  const toast = useToastNotification()
  const { t } = useI18n()

  // Pagination State
  const pagination = ref({
    page: 1,
    page_size: 20,
    total: 0,
    total_pages: 0
  })

  // Modal & Form State
  const showModal = ref(false)
  const isEditing = ref(false)
  const editingId = ref<string | null>(null)

  const columns = computed<(TableColumn<AuditGuideline> & { class?: string })[]>(() => [
    { accessorKey: 'no', header: t('auditCharter.guideline.columns.no'), class: 'w-16 whitespace-nowrap text-center' },
    { accessorKey: 'name', header: t('auditCharter.guideline.columns.name'), class: 'min-w-[280px]' },
    { accessorKey: 'status', header: t('auditCharter.guideline.columns.status'), class: 'w-32 whitespace-nowrap text-center' },
    { accessorKey: 'effective_date', header: t('auditCharter.guideline.columns.effectiveDate'), class: 'w-40 whitespace-nowrap text-center' },
    { accessorKey: 'file_name', header: t('auditCharter.guideline.columns.fileName'), class: 'w-64 min-w-[220px]' },
    { accessorKey: 'actions', header: t('auditCharter.guideline.columns.actions'), class: 'w-28 whitespace-nowrap text-center' }
  ])

  const form = reactive({
    name: '',
    status: 'Aktif',
    effective_date: new Date().toISOString().slice(0, 7), // "YYYY-MM"
    file: null as File | null,
    fileName: '',
    fileUrl: '',
    fileSize: 0
  })



  const fetchGuidelines = async (page?: number, pageSize?: number) => {
    if (page !== undefined) pagination.value.page = page
    if (pageSize !== undefined) pagination.value.page_size = pageSize

    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const authStore = useAuthStore()
      const response: any = await $fetch(
        `${baseUrl}/audit-guidelines?page=${pagination.value.page}&page_size=${pagination.value.page_size}`,
        {
          method: 'GET',
          headers: {
            Authorization: `Bearer ${authStore.token}`
          }
        }
      )
      if (response && response.data && Array.isArray(response.data.items)) {
        guidelines.value = response.data.items
        if (response.data.pagination) {
          pagination.value = {
            ...pagination.value,
            ...response.data.pagination
          }
        }
      } else {
        guidelines.value = []
      }
    } catch (err: any) {
      console.error('Failed to fetch guidelines:', err)
      errorMsg.value = extractErrorMessage(err, 'Gagal mengambil data Pedoman Audit.')
    } finally {
      loading.value = false
    }
  }

  const setPage = (page: number) => {
    fetchGuidelines(page)
  }

  const setPageSize = (size: number) => {
    fetchGuidelines(1, size)
  }

  const handleFileChange = (event: Event) => {
    const target = event.target as HTMLInputElement
    const file = target.files?.[0]
    if (!file) return

    if (file.size > 10 * 1024 * 1024) {
      errorMsg.value = 'File terlalu besar! Maksimal 10MB.'
      form.file = null
      target.value = ''
      return
    }

    errorMsg.value = ''
    form.file = file
    form.fileName = file.name
    form.fileSize = file.size
  }

  const uploadFile = async (): Promise<{ fileUrl: string; fileName: string; fileSize: number } | null> => {
    if (!form.file) return null
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const authStore = useAuthStore()

      const formData = new FormData()
      formData.append('file', form.file)
      formData.append('folder', 'Auditsphere/guidelines')
      formData.append('feature_name', 'guidelines')

      const response: any = await $fetch(`${baseUrl}/media/upload`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        },
        body: formData
      })

      if (response?.success === true && response?.data) {
        const { filePath, fileName, fileSize } = response.data

        if (!filePath) {
          throw new Error('Upload response did not contain filePath')
        }

        return {
          fileUrl: filePath,
          fileName,
          fileSize
        }
      }

      throw new Error(
        response?.error?.message || 'Invalid media upload response'
      )
    } catch (err) {
      console.error('Failed to upload file:', err)
      const detail = extractErrorMessage(err, 'Gagal mengupload file dokumen.')
      errorMsg.value = detail
      toast.showError('Gagal mengupload file dokumen.', detail)
      throw err
    }
  }

  const addGuideline = async () => {
    if (!form.file && !form.fileUrl) {
      errorMsg.value = 'Mohon upload file dokumen Pedoman Audit.'
      return
    }

    loading.value = true
    errorMsg.value = ''
    try {
      let fileUrl = form.fileUrl
      let fileName = form.fileName
      let fileSize = form.fileSize

      if (form.file) {
        const uploadResult = await uploadFile()
        if (uploadResult) {
          fileUrl = uploadResult.fileUrl
          fileName = uploadResult.fileName
          fileSize = uploadResult.fileSize
        }
      }

      const baseUrl = getAuditServiceBaseUrl()
      const authStore = useAuthStore()

      await $fetch(`${baseUrl}/audit-guidelines`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: {
          name: form.name,
          status: form.status,
          effective_date: form.effective_date,
          file_url: fileUrl,
          file_name: fileName,
          file_size: fileSize
        }
      })

      toast.showSuccess('Pedoman Audit berhasil ditambahkan!')
      await fetchGuidelines()
      closeModal()
    } catch (err: any) {
      console.error('Failed to add guideline:', err)
      const detail = extractErrorMessage(err, 'Gagal menambahkan Pedoman Audit.')
      errorMsg.value = detail
      toast.showError('Gagal menambahkan Pedoman Audit.', detail)
    } finally {
      loading.value = false
    }
  }

  const updateGuideline = async () => {
    if (!editingId.value) return
    loading.value = true
    errorMsg.value = ''
    try {
      let fileUrl = form.fileUrl
      let fileName = form.fileName
      let fileSize = form.fileSize

      if (form.file) {
        const uploadResult = await uploadFile()
        if (uploadResult) {
          fileUrl = uploadResult.fileUrl
          fileName = uploadResult.fileName
          fileSize = uploadResult.fileSize
        }
      }

      const baseUrl = getAuditServiceBaseUrl()
      const authStore = useAuthStore()

      await $fetch(`${baseUrl}/audit-guidelines/${editingId.value}`, {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: {
          name: form.name,
          status: form.status,
          effective_date: form.effective_date,
          file_url: fileUrl,
          file_name: fileName,
          file_size: fileSize
        }
      })

      toast.showSuccess('Pedoman Audit berhasil diperbarui!')
      await fetchGuidelines()
      closeModal()
    } catch (err: any) {
      console.error('Failed to update guideline:', err)
      const detail = extractErrorMessage(err, 'Gagal memperbarui Pedoman Audit.')
      errorMsg.value = detail
      toast.showError('Gagal memperbarui Pedoman Audit.', detail)
    } finally {
      loading.value = false
    }
  }

  const deleteGuideline = async (id: string) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const authStore = useAuthStore()

      await $fetch(`${baseUrl}/audit-guidelines/${id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      toast.showSuccess('Pedoman Audit berhasil dihapus!')
      await fetchGuidelines()
    } catch (err: any) {
      console.error('Failed to delete guideline:', err)
      const detail = extractErrorMessage(err, 'Gagal menghapus Pedoman Audit.')
      errorMsg.value = detail
      toast.showError('Gagal menghapus Pedoman Audit.', detail)
    } finally {
      loading.value = false
    }
  }

  const handleEdit = (item: AuditGuideline) => {
    isEditing.value = true
    editingId.value = item.id
    showModal.value = true

    form.name = item.name
    form.status = item.status
    form.effective_date = item.effective_date
    form.file = null
    form.fileName = item.file_name
    form.fileUrl = item.file_url
    form.fileSize = item.file_size
    errorMsg.value = ''
  }

  const closeModal = () => {
    showModal.value = false
    isEditing.value = false
    editingId.value = null

    form.name = ''
    form.status = 'Aktif'
    form.effective_date = new Date().toISOString().slice(0, 7)
    form.file = null
    form.fileName = ''
    form.fileUrl = ''
    form.fileSize = 0
    errorMsg.value = ''
  }

  return {
    guidelines,
    loading,
    errorMsg,
    pagination,
    showModal,
    isEditing,
    columns,
    form,
    fetchGuidelines,
    setPage,
    setPageSize,
    handleFileChange,
    addGuideline,
    updateGuideline,
    deleteGuideline,
    handleEdit,
    closeModal
  }
})
