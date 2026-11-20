import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { useI18n } from '~/composables/useI18n'
import { extractErrorMessage } from '~/utils/error'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { viewDocumentInNewTab } from '~/utils/documentViewer'

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
  const { t } = useI18n()

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
      errorMsg.value = extractErrorMessage(error, t('toasts.document.loadFailed', { name: t('toasts.documentTypes.workingPaper') }))
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
      toast.showSuccess(t('toasts.document.uploaded', { name: t('toasts.documentTypes.workingPaper') }))
      await fetchImportedPapers()
      return response
    } catch (error: any) {
      console.error('Failed to upload working paper:', error)
      const detail = extractErrorMessage(error, t('toasts.document.uploadFailed', { name: t('toasts.documentTypes.workingPaper') }))
      errorMsg.value = detail
      toast.showError(t('toasts.document.uploadFailed', { name: t('toasts.documentTypes.workingPaper') }), detail)
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
      toast.showSuccess(t('toasts.document.deleted', { name: t('toasts.documentTypes.workingPaper') }))
      await fetchImportedPapers()
    } catch (error) {
      console.error('Failed to delete imported working paper:', error)
      const detail = extractErrorMessage(error, t('toasts.document.deleteFailed', { name: t('toasts.documentTypes.workingPaper') }))
      errorMsg.value = detail
      throw error
    } finally {
      loading.value = false
    }
  }

  const viewDocument = async (id: string, fileName?: string, title?: string) => {
    const doc = importedPapers.value.find(d => d.id === id)
    const resolvedName = fileName || doc?.fileName || 'document.pdf'
    const resolvedFileName = fileName || doc?.fileName || 'document.pdf'

    try {
      const baseUrl = getAuditServiceBaseUrl()
      await viewDocumentInNewTab(
        async () => {
          try {
            return await $fetch(`${baseUrl}/working-papers/imports/${id}/view`, {
              responseType: 'blob'
            })
          } catch (_err) {
            return await $fetch(`${baseUrl}/working-papers/imports/${id}/download`, {
              responseType: 'blob'
            })
          }
        },
        resolvedName,
        resolvedFileName
      )
    } catch (error) {
      console.error('Failed to view document:', error)
      const detail = extractErrorMessage(error, t('toasts.document.viewFailed'))
      errorMsg.value = detail
      toast.showError(t('toasts.document.viewFailed'), detail)
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
      const detail = extractErrorMessage(error, t('toasts.document.downloadFailed'))
      errorMsg.value = detail
      toast.showError(t('toasts.document.downloadFailed'), detail)
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
