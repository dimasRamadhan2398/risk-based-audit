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

  const saveReport = () => {
    if (isEditing.value && editingId.value) {
      const index = reportList.value.findIndex(r => r.id === editingId.value)
      if (index !== -1) {
        reportList.value[index] = {
          ...reportList.value[index]!,
          ...reportForm,
          id: editingId.value
        } as AuditResultReport
      }
    } else {
      reportList.value.unshift({
        id: crypto.randomUUID(),
        ...reportForm
      })
    }
    closeModal()
  }

  const editReport = (report: AuditResultReport) => {
    Object.assign(reportForm, report)
    isEditing.value = true
    editingId.value = report.id
    showModal.value = true
  }

  const deleteReport = (id: string) => {
    reportList.value = reportList.value.filter(r => r.id !== id)
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
    deleteReport
  }
})
