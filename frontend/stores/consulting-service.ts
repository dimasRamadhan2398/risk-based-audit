import { defineStore } from 'pinia'
import { ref, reactive, computed, markRaw } from 'vue'
import { extractErrorMessage } from '~/utils/error'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { getFiscalYearStrings } from '~/composables/useFiscalYear'
import { useI18n } from '~/composables/useI18n'
import { useDepartmentApi } from '~/composables/useDepartmentApi'
import type { Department } from '~/types/master'

export interface ConsultingAttachment {
  name: string
  size: string
  uploadedAt: string
}

export interface ConsultingService {
  id: string
  title: string
  category: string
  requestorDept: string
  period: string
  consultantName: string
  status: string
  notes?: string
  attachment?: ConsultingAttachment
  created_at?: string
  updated_at?: string
}

export const useConsultingServiceStore = defineStore('consulting-service', () => {
  const services = ref<ConsultingService[]>([])
  const loading = ref(false)
  const errorMsg = ref('')
  const toast = useToastNotification()
  const { t } = useI18n()

  const isFormOpen = ref(false)
  const isDetailOpen = ref(false)
  const isImportOpen = ref(false)
  const isEditing = ref(false)
  const selectedService = ref<ConsultingService | null>(null)

  const categories = ['Operational Advisory', 'IT Advisory', 'Training & Workshops', 'Policy & SOP Review', 'Governance Assessment']
  const statuses = ['Planned', 'In Progress', 'Completed', 'Cancelled']
  const quarters = ['Q1', 'Q2', 'Q3', 'Q4']
  const years = getFiscalYearStrings()

  // Master Data Departemen (Department Management Sync)
  const departments = ref<Department[]>([])
  const departmentsLoading = ref(false)
  const departmentsLoaded = ref(false)
  const departmentsError = ref('')

  const resolveDepartmentOption = (val?: string): string => {
    if (!val) return ''
    const match = departments.value.find(
      d => d.id === val ||
        d.department_code === val ||
        d.department_name === val ||
        `${d.department_name} (${d.department_code})` === val
    )
    if (match) {
      return match.department_code ? `${match.department_name} (${match.department_code})` : match.department_name
    }
    return val
  }

  const formatDepartmentName = (value?: string): string => {
    if (!value) return '-'
    const found = departments.value.find(
      d => d.id === value ||
        d.department_code === value ||
        d.department_name === value ||
        `${d.department_name} (${d.department_code})` === value
    )
    if (found) {
      return found.department_code ? `${found.department_name} (${found.department_code})` : found.department_name
    }
    return value
  }

  const fetchDepartments = async (force = false) => {
    if (departmentsLoaded.value && !force) return
    if (departmentsLoading.value) return
    departmentsLoading.value = true
    departmentsError.value = ''
    try {
      const api = useDepartmentApi()
      const list = await api.getAllDepartments()
      departments.value = (Array.isArray(list) ? list : [])
        .filter(d => d?.department_name && d.is_active !== false)
        .sort((a, b) => a.department_name.localeCompare(b.department_name, 'id', { sensitivity: 'base' }))
      departmentsLoaded.value = true

      if (newService.requestorDept) {
        newService.requestorDept = resolveDepartmentOption(newService.requestorDept)
      }
    } catch (error: any) {
      console.error('Failed to fetch departments for consulting service:', error)
      departmentsError.value = extractErrorMessage(error, t('consultingService.dashboard.messages.departmentFetchFailed') || 'Gagal memuat data departemen.')
      if (!departmentsLoaded.value) {
        departments.value = []
      }
    } finally {
      departmentsLoading.value = false
    }
  }

  const departmentOptions = computed(() => {
    const list = departments.value.map(d =>
      d.department_code ? `${d.department_name} (${d.department_code})` : d.department_name
    )
    if (newService.requestorDept && !list.includes(newService.requestorDept)) {
      list.unshift(newService.requestorDept)
    }
    return Array.from(new Set(list)).sort((a, b) => a.localeCompare(b, 'id', { sensitivity: 'base' }))
  })

  const newService = reactive({
    title: '',
    category: 'Operational Advisory',
    requestorDept: '',
    periodQuarter: 'Q1',
    periodYear: new Date().getFullYear().toString(),
    consultantName: '',
    status: 'Planned',
    notes: '',
    attachment: null as File | null | undefined
  })

  const resetForm = () => {
    Object.assign(newService, {
      title: '',
      category: 'Operational Advisory',
      requestorDept: '',
      periodQuarter: 'Q1',
      periodYear: new Date().getFullYear().toString(),
      consultantName: '',
      status: 'Planned',
      notes: '',
      attachment: null
    })
  }

  const getMasterServiceBaseUrl = () => {
    return getAuditServiceBaseUrl()
  }

  const fetchServices = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getMasterServiceBaseUrl()
      const data: any = await $fetch(`${baseUrl}/consulting-services`)
      if (Array.isArray(data)) {
        services.value = data
      } else {
        services.value = []
      }
      if (!departmentsLoaded.value && !departmentsLoading.value) {
        fetchDepartments().catch(() => {})
      }
    } catch (error) {
      console.error('Failed to fetch consulting services:', error)
      errorMsg.value = extractErrorMessage(error, t('consultingService.dashboard.messages.fetchFailed'))
      services.value = []
    } finally {
      loading.value = false
    }
  }

  const toBase64 = (file: File) => new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = error => reject(error)
  })

  const saveService = async () => {
    if (!newService.title || !newService.consultantName) {
      errorMsg.value = t('consultingService.dashboard.messages.required')
      return
    }
    loading.value = true
    errorMsg.value = ''

    try {
      const baseUrl = getMasterServiceBaseUrl()
      if (isEditing.value && selectedService.value) {
        const payload = {
          title: newService.title,
          category: newService.category,
          requestorDept: newService.requestorDept,
          period: `${newService.periodQuarter} ${newService.periodYear}`.trim(),
          consultantName: newService.consultantName,
          status: newService.status,
          notes: newService.notes,
          attachment: newService.attachment && !(newService.attachment instanceof File)
            ? newService.attachment
            : selectedService.value?.attachment
        }
        await $fetch(`${baseUrl}/consulting-services/${selectedService.value.id}`, {
          method: 'PUT',
          body: payload
        })
        if (newService.attachment instanceof File) {
          toast.showError(t('consultingService.dashboard.messages.updateFileUnsupported'))
        }
      } else {
        if (newService.attachment && newService.attachment instanceof File) {
          const base64 = await toBase64(newService.attachment)
          const importPayload = {
            title: newService.title,
            category: newService.category,
            requestorDept: newService.requestorDept,
            period: `${newService.periodQuarter} ${newService.periodYear}`.trim(),
            consultantName: newService.consultantName,
            status: newService.status,
            notes: newService.notes,
            fileName: newService.attachment.name,
            fileType: newService.attachment.type || 'application/octet-stream',
            fileContent: base64
          }
          await $fetch(`${baseUrl}/consulting-services/import`, {
            method: 'POST',
            body: importPayload
          })
        } else {
          const payload = {
            title: newService.title,
            category: newService.category,
            requestorDept: newService.requestorDept,
            period: `${newService.periodQuarter} ${newService.periodYear}`.trim(),
            consultantName: newService.consultantName,
            status: newService.status,
            notes: newService.notes
          }
          await $fetch(`${baseUrl}/consulting-services`, {
            method: 'POST',
            body: payload
          })
        }
      }
      resetForm()
      isEditing.value = false
      isFormOpen.value = false
      toast.showSuccess(t('consultingService.dashboard.messages.saveSuccess'))
      await fetchServices()
    } catch (error: any) {
      console.error('Failed to save consulting service:', error)
      toast.showError(error.data?.message || t('consultingService.dashboard.messages.saveFailed'))
    } finally {
      loading.value = false
    }
  }

  const deleteService = async (service?: ConsultingService) => {
    const targetService = service || selectedService.value

    if (!targetService) return

    loading.value = true
    errorMsg.value = ''

    try {
      const baseUrl = getMasterServiceBaseUrl()

      await $fetch(`${baseUrl}/consulting-services/${targetService.id}`, {
        method: 'DELETE'
      })

      if (selectedService.value?.id === targetService.id) {
        isDetailOpen.value = false
        selectedService.value = null
      }
      toast.showSuccess(t('consultingService.dashboard.messages.deleteSuccess'))
      await fetchServices()
    } catch (error: any) {
      console.error('Failed to delete consulting service:', error)
      toast.showError(error.data?.message || t('consultingService.dashboard.messages.deleteFailed'))
    } finally {
      loading.value = false
    }
  }

  const handleFileUpload = (files: FileList) => {
    if (files.length > 0) {
      const file = files[0]
      if (!file) return

      if (file.size > 10 * 1024 * 1024) {
        errorMsg.value = t('consultingService.dashboard.messages.fileTooLarge')
        return
      }
      newService.attachment = markRaw(file) as any
      errorMsg.value = ''
    }
  }

  const downloadAttachment = async (id: string, fileName: string) => {
    try {
      const baseUrl = getMasterServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/consulting-services/${id}/download`, {
        responseType: 'blob'
      })
      const blob = new Blob([response], { type: response.type || 'application/octet-stream' })
      const link = document.createElement('a')
      link.href = window.URL.createObjectURL(blob)
      link.download = fileName
      link.click()
      window.URL.revokeObjectURL(link.href)
    } catch (error) {
      console.error('Failed to download consulting attachment:', error)
      errorMsg.value = extractErrorMessage(error, t('consultingService.dashboard.messages.downloadFailed'))
    }
  }

  const viewAttachment = async (id: string, fileName: string) => {
    try {
      const baseUrl = getMasterServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/consulting-services/${id}/download`, {
        responseType: 'blob'
      })
      const blob = new Blob([response], { type: response.type || 'application/pdf' })
      const url = window.URL.createObjectURL(blob)
      window.open(url, '_blank')
      setTimeout(() => window.URL.revokeObjectURL(url), 10000)
    } catch (error) {
      console.error('Failed to view consulting attachment:', error)
      toast.showError(t('consultingService.dashboard.messages.viewFailed'))
    }
  }

  const openForm = () => {
    isEditing.value = false
    resetForm()
    if (!departmentsLoaded.value && !departmentsLoading.value) {
      fetchDepartments().catch(() => {})
    }
    isFormOpen.value = true
  }

  const editService = (service?: ConsultingService) => {
    if (service) {
      selectedService.value = service
    }

    if (!selectedService.value) return

    isEditing.value = true

    const periodParts = (selectedService.value.period || '').split(' ')

    Object.assign(newService, {
      title: selectedService.value.title,
      category: selectedService.value.category,
      requestorDept: resolveDepartmentOption(selectedService.value.requestorDept),
      periodQuarter: periodParts[0] || 'Q1',
      periodYear: periodParts[1] || '2026',
      consultantName: selectedService.value.consultantName,
      status: selectedService.value.status,
      notes: selectedService.value.notes || '',
      attachment: selectedService.value.attachment ? { ...selectedService.value.attachment } : null
    })

    if (!departmentsLoaded.value && !departmentsLoading.value) {
      fetchDepartments().catch(() => {})
    }

    isDetailOpen.value = false
    isFormOpen.value = true
  }

  const closeForm = () => {
    isFormOpen.value = false
  }

  const openDetail = (service: ConsultingService) => {
    selectedService.value = service
    isDetailOpen.value = true
  }

  const closeDetail = () => {
    isDetailOpen.value = false
    selectedService.value = null
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'Completed': return 'success'
      case 'In Progress': return 'warning'
      case 'Planned': return 'primary'
      case 'Cancelled': return 'error'
      default: return 'neutral'
    }
  }

  return {
    services, loading, errorMsg, isFormOpen, isDetailOpen, isImportOpen, isEditing, selectedService,
    categories, statuses, departments, quarters, years, newService,
    departmentsLoading, departmentsLoaded, departmentsError,
    departmentOptions, requestorDeptOptions: departmentOptions,
    fetchDepartments, resolveDepartmentOption, formatDepartmentName,
    fetchServices, saveService, deleteService, handleFileUpload, downloadAttachment, viewAttachment,
    openForm, editService, closeForm, openDetail, closeDetail, getStatusColor
  }
})
