import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { extractErrorMessage } from '~/utils/error'

export interface ImportedWorkingPaper {
  id: string
  title: string
  description: string
  fileName: string
  filePath: string
  fileSize: number
  fileType: string
  created_at: string
  updated_at: string
}

export const useImportWorkingPaperStore = defineStore('import-working-paper', () => {
  const importedPapers = ref<ImportedWorkingPaper[]>([])
  const loading = ref(false)
  const errorMsg = ref('')
  const toast = useToastNotification()

  const getAuditServiceBaseUrl = () => {
    const config = useRuntimeConfig()
    return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
  }

  const fetchImportedPapers = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/working-papers/imports`, { method: 'GET' })
      if (response && Array.isArray(response.data)) {
        importedPapers.value = response.data
      } else if (response && response.data && Array.isArray(response.data.items)) {
        importedPapers.value = response.data.items
      } else if (response && Array.isArray(response.items)) {
        importedPapers.value = response.items
      } else if (Array.isArray(response)) {
        importedPapers.value = response
      } else {
        importedPapers.value = []
      }
    } catch (error) {
      console.error('Failed to fetch imported working papers:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to load imported working papers.')
    } finally {
      loading.value = false
    }
  }

  const importWorkingPaper = async (payload: {
    title: string
    description: string
    fileName: string
    fileType: string
    file: File
  }) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const formData = new FormData()
      Object.keys(payload).forEach(key => {
        const val = (payload as any)[key]
        if (val !== undefined && val !== null) {
          formData.append(key, val)
        }
      })
      const response: any = await $fetch(`${baseUrl}/working-papers/imports`, {
        method: 'POST',
        body: formData
      })
      toast.showSuccess('Working paper imported successfully')
      await fetchImportedPapers()
      return response
    } catch (error: any) {
      console.error('Failed to upload working paper:', error)
      const detail = extractErrorMessage(error, 'Failed to upload working paper.')
      errorMsg.value = detail
      toast.showError('Failed to upload working paper.', detail)
      throw error
    } finally {
      loading.value = false
    }
  }

  const deleteImportedPaper = async (id: string) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/working-papers/imports/${id}`, {
        method: 'DELETE'
      })
      toast.showSuccess('Imported working paper deleted successfully')
      await fetchImportedPapers()
    } catch (error) {
      console.error('Failed to delete imported working paper:', error)
      const detail = extractErrorMessage(error, 'Failed to delete imported working paper.')
      errorMsg.value = detail
      toast.showError('Failed to delete imported working paper.', detail)
      throw error
    } finally {
      loading.value = false
    }
  }

  const viewDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/working-papers/imports/${id}/download`, {
        responseType: 'blob'
      })
      const blob = new Blob([response], { type: response.type || 'application/pdf' })
      const url = window.URL.createObjectURL(blob)
      window.open(url, '_blank')
      setTimeout(() => window.URL.revokeObjectURL(url), 10000)
    } catch (error) {
      console.error('Failed to view document:', error)
      const detail = extractErrorMessage(error, 'Failed to view document.')
      errorMsg.value = detail
      toast.showError('Failed to view document.', detail)
    }
  }

  const downloadImportedPaper = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/working-papers/imports/${id}/download`, {
        responseType: 'blob'
      })
      const blob = new Blob([response], { type: response.type || 'application/octet-stream' })
      const link = document.createElement('a')
      link.href = window.URL.createObjectURL(blob)
      link.download = fileName
      link.click()
      window.URL.revokeObjectURL(link.href)
    } catch (error) {
      console.error('Failed to download imported working paper:', error)
      const detail = extractErrorMessage(error, 'Failed to download file.')
      errorMsg.value = detail
      toast.showError('Failed to download file.', detail)
    }
  }

  return {
    importedPapers,
    loading,
    errorMsg,
    fetchImportedPapers,
    importWorkingPaper,
    deleteImportedPaper,
    downloadImportedPaper,
    viewDocument
  }
})
