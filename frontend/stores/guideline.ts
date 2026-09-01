import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'

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
      errorMsg.value = 'Gagal mengambil data Pedoman Audit.'
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
      formData.append('folder', 'audit')

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
      errorMsg.value = 'Gagal mengupload file dokumen.'
      throw err
    }
  }

  const addGuideline = async () => {
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

      await fetchGuidelines()
      closeModal()
    } catch (err: any) {
      console.error('Failed to add guideline:', err)
      errorMsg.value = 'Gagal menambahkan Pedoman Audit.'
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

      await fetchGuidelines()
      closeModal()
    } catch (err: any) {
      console.error('Failed to update guideline:', err)
      errorMsg.value = 'Gagal memperbarui Pedoman Audit.'
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

      await fetchGuidelines()
    } catch (err: any) {
      console.error('Failed to delete guideline:', err)
      errorMsg.value = 'Gagal menghapus Pedoman Audit.'
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
