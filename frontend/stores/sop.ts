import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'

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



  const fetchSops = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const authStore = useAuthStore()
      const response: any = await $fetch(`${baseUrl}/audit-sops?page_size=100`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })
      if (response && response.data && Array.isArray(response.data.items)) {
        sops.value = response.data.items
      } else {
        sops.value = []
      }
    } catch (err: any) {
      console.error('Failed to fetch sops:', err)
      errorMsg.value = 'Gagal mengambil data Petunjuk Teknis/SOP.'
    } finally {
      loading.value = false
    }
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

      if (response && response.status === 'success' && response.data) {
        return {
          fileUrl: response.data.filePath,
          fileName: response.data.fileName,
          fileSize: response.data.fileSize
        }
      }
      return null
    } catch (err) {
      console.error('Failed to upload file:', err)
      errorMsg.value = 'Gagal mengupload file dokumen.'
      throw err
    }
  }

  const addSop = async () => {
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

      await fetchSops()
      closeModal()
    } catch (err: any) {
      console.error('Failed to add sop:', err)
      errorMsg.value = 'Gagal menambahkan Petunjuk Teknis/SOP.'
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

      await fetchSops()
      closeModal()
    } catch (err: any) {
      console.error('Failed to update sop:', err)
      errorMsg.value = 'Gagal memperbarui Petunjuk Teknis/SOP.'
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

      await fetchSops()
    } catch (err: any) {
      console.error('Failed to delete sop:', err)
      errorMsg.value = 'Gagal menghapus Petunjuk Teknis/SOP.'
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
    showModal,
    isEditing,
    form,
    fetchSops,
    handleFileChange,
    addSop,
    updateSop,
    deleteSop,
    handleEdit,
    closeModal
  }
})
