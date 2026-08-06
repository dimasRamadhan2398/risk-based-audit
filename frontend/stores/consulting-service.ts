import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'

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

  const isFormOpen = ref(false)
  const isDetailOpen = ref(false)
  const isImportOpen = ref(false)
  const isEditing = ref(false)
  const selectedService = ref<ConsultingService | null>(null)

  const categories = ['Operational Advisory', 'IT Advisory', 'Training & Workshops', 'Policy & SOP Review', 'Governance Assessment']
  const statuses = ['Planned', 'In Progress', 'Completed', 'Cancelled']
  const departments = ['IT', 'Finance', 'HR', 'Ops', 'Procurement', 'Legal', 'Sales']
  const quarters = ['Q1', 'Q2', 'Q3', 'Q4']
  const years = ['2024', '2025', '2026', '2027']

  const newService = reactive({
    title: '',
    category: 'Operational Advisory',
    requestorDept: 'IT',
    periodQuarter: 'Q1',
    periodYear: '2026',
    consultantName: '',
    status: 'Planned',
    notes: '',
    attachment: null as File | null | undefined
  })

  const resetForm = () => {
    Object.assign(newService, {
      title: '',
      category: 'Operational Advisory',
      requestorDept: 'IT',
      periodQuarter: 'Q1',
      periodYear: '2026',
      consultantName: '',
      status: 'Planned',
      notes: '',
      attachment: null
    })
  }

  const getMasterServiceBaseUrl = () => {
    return getAuditServiceBaseUrl()
  }

  const mockServices: ConsultingService[] = [
    {
      id: 'CS-001',
      title: 'Review & Assessment SOP Procurement & Purchasing',
      category: 'Policy & SOP Review',
      requestorDept: 'Procurement',
      period: 'Q1 2026',
      consultantName: 'Budi Hartanto',
      status: 'Completed',
      notes: 'Rekomendasi perbaikan kontrol persetujuan pembelian di atas Rp 100jt telah disetujui.',
      attachment: {
        name: 'SOP_Procurement_Review_Report.pdf',
        size: '1.2 MB',
        uploadedAt: '2026-03-10'
      }
    },
    {
      id: 'CS-002',
      title: 'IT Security Governance Assessment & Training',
      category: 'IT Advisory',
      requestorDept: 'IT',
      period: 'Q2 2026',
      consultantName: 'Wahyu Hidayat',
      status: 'In Progress',
      notes: 'Melakukan review celah keamanan dan melatih staf IT tentang standar ISO 27001.',
      attachment: {
        name: 'IT_Security_Audit_Scope.pdf',
        size: '850 KB',
        uploadedAt: '2026-05-18'
      }
    },
    {
      id: 'CS-003',
      title: 'Corporate Governance & Compliance Assessment',
      category: 'Governance Assessment',
      requestorDept: 'Legal',
      period: 'Q3 2026',
      consultantName: 'Carolina Wijaya',
      status: 'Planned',
      notes: 'Asesmen keselarasan kebijakan internal perusahaan terhadap regulasi POJK terbaru.'
    }
  ]

  const fetchServices = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getMasterServiceBaseUrl()
      const data: any = await $fetch(`${baseUrl}/consulting-services`)
      if (Array.isArray(data) && data.length > 0) {
        services.value = data
      } else {
        services.value = [...mockServices]
      }
    } catch (error) {
      console.error('Failed to fetch consulting services, falling back to mock:', error)
      errorMsg.value = 'Failed to load consulting services.'
      services.value = [...mockServices]
    } finally {
      loading.value = false
    }
  }

  const saveService = async () => {
    if (!newService.title || !newService.consultantName) {
      errorMsg.value = 'Title and Consultant Name are required.'
      return
    }
    loading.value = true
    errorMsg.value = ''

    const payload = {
      title: newService.title,
      category: newService.category,
      requestorDept: newService.requestorDept,
      period: `${newService.periodQuarter} ${newService.periodYear}`.trim(),
      consultantName: newService.consultantName,
      status: newService.status,
      notes: newService.notes,
      attachment: newService.attachment ? {
        name: newService.attachment.name,
        size: Math.round(newService.attachment.size / 1024) + ' KB',
        uploadedAt: new Date().toISOString().split('T')[0] || ''
      } : (isEditing.value ? selectedService.value?.attachment : undefined)
    }

    try {
      const baseUrl = getMasterServiceBaseUrl()
      if (isEditing.value && selectedService.value) {
        await $fetch(`${baseUrl}/consulting-services/${selectedService.value.id}`, {
          method: 'PUT',
          body: payload
        })
      } else {
        await $fetch(`${baseUrl}/consulting-services`, {
          method: 'POST',
          body: payload
        })
      }
      resetForm()
      isEditing.value = false
      isFormOpen.value = false
      await fetchServices()
    } catch (error: any) {
      console.error('Failed to save consulting service:', error)
      errorMsg.value = 'Failed to save consulting service.'
    } finally {
      loading.value = false
    }
  }

  const deleteService = async () => {
    if (!selectedService.value) return
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getMasterServiceBaseUrl()
      await $fetch(`${baseUrl}/consulting-services/${selectedService.value.id}`, {
        method: 'DELETE'
      })
      isDetailOpen.value = false
      selectedService.value = null
      await fetchServices()
    } catch (error) {
      console.error('Failed to delete consulting service:', error)
      errorMsg.value = 'Failed to delete consulting service.'
    } finally {
      loading.value = false
    }
  }

  const handleFileUpload = (files: FileList) => {
    if (files.length > 0) {
      const file = files[0]
      if (file && file.size > 10 * 1024 * 1024) {
        errorMsg.value = 'File is too large (max 10MB)'
        return
      }
      newService.attachment = file
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
    }
  }

  const openForm = () => {
    isEditing.value = false
    resetForm()
    isFormOpen.value = true
  }

  const editService = () => {
    if (!selectedService.value) return
    isEditing.value = true
    
    const periodParts = (selectedService.value.period || '').split(' ')
    Object.assign(newService, {
      title: selectedService.value.title,
      category: selectedService.value.category,
      requestorDept: selectedService.value.requestorDept,
      periodQuarter: periodParts[0] || 'Q1',
      periodYear: periodParts[1] || '2026',
      consultantName: selectedService.value.consultantName,
      status: selectedService.value.status,
      notes: selectedService.value.notes || '',
      attachment: null
    })
    
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
    fetchServices, saveService, deleteService, handleFileUpload, downloadAttachment,
    openForm, editService, closeForm, openDetail, closeDetail, getStatusColor
  }
})
