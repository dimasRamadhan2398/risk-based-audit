import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { extractErrorMessage } from '~/utils/error'

export interface AuditSop {
  id: string
  name: string
  guideline_id: string
  guideline?: {
    id: string
    name: string
  }
  status: string // "Aktif" | "Sedang Diperbarui"
  effective_date: string // "YYYY-MM"
  file_url: string
  file_name: string
  file_size: number
  created_at?: string
  updated_at?: string
}

export const useSopStore = defineStore('sop', () => {
  const sops = ref<AuditSop[]>([])
  const loading = ref(false)
  const errorMsg = ref('')
  const toast = useToastNotification()

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

  const form = reactive({
    name: '',
    guideline_id: '',
    status: 'Aktif',
    effective_date: new Date().toISOString().slice(0, 7), // "YYYY-MM"
    file: null as File | null,
    fileName: '',
    fileUrl: '',
    fileSize: 0
  })

  const getAuditServiceBaseUrl = () => {
    const config = useRuntimeConfig()
    return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
  }

  const fetchSops = async (page?: number, pageSize?: number) => {
    if (page !== undefined) pagination.value.page = page
    if (pageSize !== undefined) pagination.value.page_size = pageSize

    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const authStore = useAuthStore()
      const response: any = await $fetch(
        `${baseUrl}/audit-sops?page=${pagination.value.page}&page_size=${pagination.value.page_size}`,
        {
          method: 'GET',
          headers: {
            Authorization: `Bearer ${authStore.token}`
          }
        }
      )
      if (response && response.data && Array.isArray(response.data.items)) {
        sops.value = response.data.items
        if (response.data.pagination) {
          pagination.value = {
            ...pagination.value,
            ...response.data.pagination
          }
        }
      } else {
        sops.value = []
      }
    } catch (err: any) {
      console.error('Failed to fetch sops:', err)
      errorMsg.value = extractErrorMessage(err, 'Gagal mengambil data Petunjuk Teknis/SOP.')
    } finally {
      loading.value = false
    }
  }

  const setPage = (page: number) => {
    fetchSops(page)
  }

  const setPageSize = (size: number) => {
    fetchSops(1, size)
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
      formData.append('folder', 'Auditsphere/sop')
      formData.append('feature_name', 'sop')

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

  const addSop = async () => {
    if (!form.file && !form.fileUrl) {
      errorMsg.value = 'Mohon upload file dokumen Petunjuk Teknis / SOP.'
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

      await $fetch(`${baseUrl}/audit-sops`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: {
          name: form.name,
          guideline_id: form.guideline_id,
          status: form.status,
          effective_date: form.effective_date,
          file_url: fileUrl,
          file_name: fileName,
          file_size: fileSize
        }
      })

      toast.showSuccess('Petunjuk Teknis/SOP berhasil ditambahkan!')
      await fetchSops()
      closeModal()
    } catch (err: any) {
      console.error('Failed to add sop:', err)
      const detail = extractErrorMessage(err, 'Gagal menambahkan Petunjuk Teknis/SOP.')
      errorMsg.value = detail
      toast.showError('Gagal menambahkan Petunjuk Teknis/SOP.', detail)
    } finally {
      loading.value = false
    }
  }

  const updateSop = async () => {
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

      await $fetch(`${baseUrl}/audit-sops/${editingId.value}`, {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: {
          name: form.name,
          guideline_id: form.guideline_id,
          status: form.status,
          effective_date: form.effective_date,
          file_url: fileUrl,
          file_name: fileName,
          file_size: fileSize
        }
      })

      toast.showSuccess('Petunjuk Teknis/SOP berhasil diperbarui!')
      await fetchSops()
      closeModal()
    } catch (err: any) {
      console.error('Failed to update sop:', err)
      const detail = extractErrorMessage(err, 'Gagal memperbarui Petunjuk Teknis/SOP.')
      errorMsg.value = detail
      toast.showError('Gagal memperbarui Petunjuk Teknis/SOP.', detail)
    } finally {
      loading.value = false
    }
  }

  const deleteSop = async (id: string) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const authStore = useAuthStore()

      await $fetch(`${baseUrl}/audit-sops/${id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      toast.showSuccess('Petunjuk Teknis/SOP berhasil dihapus!')
      await fetchSops()
    } catch (err: any) {
      console.error('Failed to delete sop:', err)
      const detail = extractErrorMessage(err, 'Gagal menghapus Petunjuk Teknis/SOP.')
      errorMsg.value = detail
      toast.showError('Gagal menghapus Petunjuk Teknis/SOP.', detail)
    } finally {
      loading.value = false
    }
  }

  const handleEdit = (item: AuditSop) => {
    isEditing.value = true
    editingId.value = item.id
    showModal.value = true

    form.name = item.name
    form.guideline_id = item.guideline_id
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
    form.guideline_id = ''
    form.status = 'Aktif'
    form.effective_date = new Date().toISOString().slice(0, 7)
    form.file = null
    form.fileName = ''
    form.fileUrl = ''
    form.fileSize = 0
    errorMsg.value = ''
  }

  return {
    sops,
    loading,
    errorMsg,
    pagination,
    showModal,
    isEditing,
    form,
    fetchSops,
    setPage,
    setPageSize,
    handleFileChange,
    addSop,
    updateSop,
    deleteSop,
    handleEdit,
    closeModal
  }
})
