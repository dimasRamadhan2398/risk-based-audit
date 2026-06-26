import { defineStore } from 'pinia'
import { ref } from 'vue'

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
      } else if (response && Array.isArray(response.items)) {
        importedPapers.value = response.items
      } else if (Array.isArray(response)) {
        importedPapers.value = response
      }
    } catch (error) {
      console.error('Failed to fetch imported working papers:', error)
      errorMsg.value = 'Failed to load imported working papers.'
    } finally {
      loading.value = false
    }
  }

  const importWorkingPaper = async (payload: {
    title: string
    description: string
    fileName: string
    fileType: string
    fileContent: string // base64
  }) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/working-papers/imports`, {
        method: 'POST',
        body: payload
      })
      await fetchImportedPapers()
      return response
    } catch (error: any) {
      console.error('Failed to import working paper:', error)
      errorMsg.value = error.data?.message || 'Failed to import working paper.'
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
      await fetchImportedPapers()
    } catch (error) {
      console.error('Failed to delete imported working paper:', error)
      errorMsg.value = 'Failed to delete imported working paper.'
      throw error
    } finally {
      loading.value = false
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
      errorMsg.value = 'Failed to download file.'
    }
  }

  return {
    importedPapers,
    loading,
    errorMsg,
    fetchImportedPapers,
    importWorkingPaper,
    deleteImportedPaper,
    downloadImportedPaper
  }
})
