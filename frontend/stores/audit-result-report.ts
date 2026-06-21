import { defineStore } from 'pinia'
import { ref, computed, reactive } from 'vue'
import { useAssignmentLetterStore } from './assignment-letter'

export interface AuditResultReport {
  id: string
  assignmentLetterId: string
  reportTitle: string
  executiveSummary: string
  overallRating: 'Satisfactory' | 'Needs Improvement' | 'Unsatisfactory'
  reportDate: string
  status: 'Draft' | 'Final'
  findingsCount: number
}

export const useAuditResultReportStore = defineStore('audit-result-report', () => {
  const assignmentLetterStore = useAssignmentLetterStore()

  // State
  const selectedAssignmentLetter = ref<string>('')
  const reportList = ref<AuditResultReport[]>([
    {
        id: 'R-001',
        assignmentLetterId: 'ST-001/SKAI/2026',
        reportTitle: 'Audit Report - Financial Operations 2026',
        executiveSummary: 'The financial operations are generally effective with some minor findings in documentation.',
        overallRating: 'Satisfactory',
        reportDate: '2026-04-15',
        status: 'Final',
        findingsCount: 3
    }
  ])
  const showModal = ref(false)
  const isEditing = ref(false)
  const editingId = ref<string | null>(null)

  const reportForm = reactive({
    assignmentLetterId: '',
    reportTitle: '',
    executiveSummary: '',
    overallRating: 'Satisfactory' as 'Satisfactory' | 'Needs Improvement' | 'Unsatisfactory',
    reportDate: new Date().toISOString().split('T')[0] as string,
    status: 'Draft' as 'Draft' | 'Final',
    findingsCount: 0
  })

  // Computed
  const publishedAssignmentLetters = computed(() => {
    return assignmentLetterStore.assignmentLetterList
      .filter((st: any) => st.status === 'Published')
      .map((st: any) => st.letterNumber)
  })

  const filteredReports = computed(() => {
    if (!selectedAssignmentLetter.value) return reportList.value
    return reportList.value.filter(r => r.assignmentLetterId === selectedAssignmentLetter.value)
  })

  const hasSelectedAssignmentLetter = computed(() => !!selectedAssignmentLetter.value && selectedAssignmentLetter.value !== '')

  const getAuditServiceBaseUrl = () => {
    const config = useRuntimeConfig()
    return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
  }

  const loading = ref(false)
  const errorMsg = ref('')

  const mockReports: AuditResultReport[] = [
    {
      id: 'R-001',
      assignmentLetterId: 'ST-001/SKAI/2026',
      reportTitle: 'Audit Report - Financial Operations 2026',
      executiveSummary: 'The financial operations are generally effective with some minor findings in documentation.',
      overallRating: 'Satisfactory',
      reportDate: '2026-04-15',
      status: 'Final',
      findingsCount: 3
    }
  ]

  const fetchReports = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/audit-result-reports`, { method: 'GET' })
      let items: AuditResultReport[] = []
      if (response && Array.isArray(response.items)) {
        items = response.items
      } else if (Array.isArray(response)) {
        items = response
      }

      if (items.length > 0) {
        reportList.value = items
      } else {
        reportList.value = [...mockReports]
      }
    } catch (error) {
      console.error('Failed to fetch reports, falling back to mock data:', error)
      errorMsg.value = 'Failed to load audit result reports.'
      reportList.value = [...mockReports]
    } finally {
      loading.value = false
    }
  }

  // Fetch on initialization
  fetchReports()

  // Actions
  const openModal = () => {
    resetForm()
    isEditing.value = false
    editingId.value = null
    showModal.value = true
  }

  const closeModal = () => {
    showModal.value = false
  }

  const resetForm = () => {
    Object.assign(reportForm, {
      assignmentLetterId: selectedAssignmentLetter.value,
      reportTitle: '',
      executiveSummary: '',
      overallRating: 'Satisfactory',
      reportDate: new Date().toISOString().split('T')[0] as string,
      status: 'Draft',
      findingsCount: 0
    })
  }

  const saveReport = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const payload = {
        ...reportForm,
        findingsCount: Number(reportForm.findingsCount)
      }
      if (isEditing.value && editingId.value) {
        await $fetch(`${baseUrl}/audit-result-reports/${editingId.value}`, {
          method: 'PUT',
          body: payload
        })
      } else {
        await $fetch(`${baseUrl}/audit-result-reports`, {
          method: 'POST',
          body: payload
        })
      }
      closeModal()
      await fetchReports()
    } catch (error: any) {
      console.error('Failed to save report:', error)
      alert('Failed to save report: ' + error.message)
    } finally {
      loading.value = false
    }
  }

  const editReport = (report: AuditResultReport) => {
    Object.assign(reportForm, report)
    isEditing.value = true
    editingId.value = report.id
    showModal.value = true
  }

  const deleteReport = async (id: string) => {
    if (!confirm('Are you sure you want to delete this report?')) return
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/audit-result-reports/${id}`, {
        method: 'DELETE'
      })
      await fetchReports()
    } catch (error: any) {
      console.error('Failed to delete report:', error)
      alert('Failed to delete report: ' + error.message)
    } finally {
      loading.value = false
    }
  }

  return {
    selectedAssignmentLetter,
    reportList,
    showModal,
    isEditing,
    reportForm,
    publishedAssignmentLetters,
    filteredReports,
    hasSelectedAssignmentLetter,
    openModal,
    closeModal,
    saveReport,
    editReport,
    deleteReport,
    loading,
    errorMsg,
    fetchReports
  }
})
