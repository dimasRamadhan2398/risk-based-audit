import { defineStore } from 'pinia'
import { ref, computed, reactive } from 'vue'
import { QAStatus, QAType, type QAReport } from '~/types/quality-assurance'

export const useQualityAssuranceStore = defineStore('quality-assurance', () => {
  const loading = ref(false)
  const errorMsg = ref('')

  const getMasterServiceBaseUrl = () => {
    const config = useRuntimeConfig()
    return config.public.masterServiceBaseUrl || 'http://localhost:8003/api/v1'
  }

  const columns = [
    { accessorKey: 'type', header: 'QA Type', sortable: true },
    { accessorKey: 'period', header: 'Period', sortable: true },
    { accessorKey: 'reportName', header: 'Report Name', sortable: true },
    { accessorKey: 'result', header: 'Result/Score', sortable: true },
    { accessorKey: 'status', header: 'Status', sortable: true },
    { accessorKey: 'actions', header: 'Action' }
  ]

  const getStatusColor = (status: QAStatus) => {
    switch (status) {
      case QAStatus.COMPLETED: return 'bg-emerald-500'
      case QAStatus.VERIFIED: return 'bg-emerald-500'
      case QAStatus.IN_PROGRESS: return 'bg-amber-400'
      case QAStatus.PLANNED: return 'bg-gray-300'
      default: return 'bg-gray-300'
    }
  }

  const getTypeIconColor = (type: QAType) => {
    switch (type) {
      case QAType.REGULAR: return 'bg-amber-400'
      case QAType.SAIV: return 'bg-blue-500'
      case QAType.QAR: return 'bg-gray-100 border'
      default: return 'bg-gray-300'
    }
  }

  const page = ref(1)
  const pageCount = 5
  const items = computed(() => {
    return filteredReports.value.slice((page.value - 1) * pageCount, (page.value) * pageCount)
  })

  const qaTypes = Object.values(QAType)
  const qaStatuses = Object.values(QAStatus)
  const periods = ['2025', '2024', '2023']

  // Form State
  const newReport = reactive({
    type: QAType.REGULAR,
    assessmentTitle: '',
    periodQuarter: '',
    periodYear: '',
    status: QAStatus.IN_PROGRESS,
    result: '',
    internalEvaluator: '',
    attachment: null as File | null | undefined
  })

  const handleFileUpload = (files: FileList | null) => {
    if (files && files.length > 0) {
      newReport.attachment = files[0]
    }
  }

  const resetForm = () => {
    Object.assign(newReport, {
      type: QAType.REGULAR,
      assessmentTitle: '',
      periodQuarter: '',
      periodYear: '',
      status: QAStatus.IN_PROGRESS,
      result: '',
      internalEvaluator: '',
      attachment: null
    })
  }

  const mockReports: QAReport[] = [
    {
      id: '1',
      type: QAType.REGULAR,
      period: 'Q3 2025',
      reportName: 'Operational Efficiency Q3',
      result: '8.7/10',
      status: QAStatus.COMPLETED,
      assessmentTitle: 'RSA - Audit 2025 Q3',
      internalEvaluator: 'Internal Audit Team A',
      attachment: {
        name: 'Report_RSA_Q3_2025.pdf',
        size: '1.5 MB',
        uploadedAt: '2025-10-01'
      }
    },
    {
      id: '2',
      type: QAType.SAIV,
      period: 'Cycle 2025',
      reportName: 'Self Assessment GIAS \'22-24',
      result: '92%',
      status: QAStatus.COMPLETED,
      assessmentTitle: 'SAIV - Cycle 2025',
      validator: 'PT Independent Consultant X',
      attachment: {
        name: 'Certificate_SAIV_2025.pdf',
        size: '2.1 MB',
        uploadedAt: '2025-11-15'
      }
    },
    {
      id: '3',
      type: QAType.QAR,
      period: 'Year 2025',
      reportName: 'External QAR (IPPF 2027)',
      result: 'G/C*',
      status: QAStatus.COMPLETED,
      assessmentTitle: 'QAR - Year 2025',
      validator: 'Deloitte Independent Consultant',
      attachment: {
        name: 'Report_External_QAR_2025.pdf',
        size: '4.2 MB',
        uploadedAt: '2025-12-20'
      }
    },
    {
      id: '4',
      type: QAType.REGULAR,
      period: 'Q2 2025',
      reportName: 'Operational Efficiency Q2',
      result: '8.3/10',
      status: QAStatus.COMPLETED,
      assessmentTitle: 'RSA - Audit 2025 Q2'
    },
    {
      id: '5',
      type: QAType.REGULAR,
      period: 'Q1 2025',
      reportName: 'Operational Efficiency Q1',
      result: '6.9/10',
      status: QAStatus.COMPLETED,
      assessmentTitle: 'RSA - Audit 2025 Q1'
    }
  ]

  const fetchReports = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getMasterServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/quality-assurance`, {
        method: 'GET'
      })
      if (Array.isArray(response) && response.length > 0) {
        reports.value = response
      } else {
        reports.value = mockReports
      }
    } catch (error: any) {
      console.error('Failed to fetch QA reports:', error)
      errorMsg.value = 'Failed to fetch quality assurance reports.'
      reports.value = mockReports
    } finally {
      loading.value = false
    }
  }

  // Load on init
  fetchReports()


  const saveReport = async () => {
    if (!newReport.assessmentTitle || !newReport.result) return
    loading.value = true
    errorMsg.value = ''

    const reportData = {
      type: newReport.type,
      period: `${newReport.periodQuarter} ${newReport.periodYear}`.trim() || '2025',
      reportName: newReport.assessmentTitle,
      result: newReport.result,
      status: newReport.status,
      assessmentTitle: newReport.assessmentTitle,
      internalEvaluator: newReport.internalEvaluator,
      attachment: newReport.attachment ? {
        name: newReport.attachment.name,
        size: Math.round(newReport.attachment.size / 1024) + ' KB',
        uploadedAt: new Date().toISOString().split('T')[0] || ''
      } : (isEditing.value ? selectedReport.value?.attachment : undefined)
    }

    try {
      const baseUrl = getMasterServiceBaseUrl()
      if (isEditing.value && selectedReport.value) {
        await $fetch(`${baseUrl}/quality-assurance/${selectedReport.value.id}`, {
          method: 'PUT',
          body: reportData
        })
      } else {
        await $fetch(`${baseUrl}/quality-assurance`, {
          method: 'POST',
          body: reportData
        })
      }
      resetForm()
      isEditing.value = false
      closeForm()
      await fetchReports()
    } catch (error: any) {
      console.error('Failed to save QA report:', error)
      errorMsg.value = 'Failed to save Quality Assurance report.'
    } finally {
      loading.value = false
    }
  }

  const reports = ref<QAReport[]>([...mockReports])

  const searchQuery = ref('')
  const selectedType = ref('')
  const selectedPeriod = ref('')
  const selectedStatus = ref('')

  const isFormOpen = ref(false)
  const isDetailOpen = ref(false)
  const selectedReport = ref<QAReport | null>(null)

  const filteredReports = computed(() => {
    return reports.value.filter(report => {
      const matchesSearch = (report.reportName || '').toLowerCase().includes(searchQuery.value.toLowerCase()) ||
        (report.assessmentTitle || '').toLowerCase().includes(searchQuery.value.toLowerCase())
      const matchesType = !selectedType.value || report.type === selectedType.value
      const matchesPeriod = !selectedPeriod.value || (report.period || '').includes(selectedPeriod.value)
      const matchesStatus = !selectedStatus.value || report.status === selectedStatus.value
      return matchesSearch && matchesType && matchesPeriod && matchesStatus
    })
  })

  const summary = computed(() => {
    const regular = reports.value.filter(r => r.type === QAType.REGULAR).sort((a, b) => (b.period || '').localeCompare(a.period || ''))[0]
    const qar = reports.value.filter(r => r.type === QAType.QAR).sort((a, b) => (b.period || '').localeCompare(a.period || ''))[0]
    const saiv = reports.value.filter(r => r.type === QAType.SAIV).sort((a, b) => (b.period || '').localeCompare(a.period || ''))[0]

    return {
      regular: regular || { result: '-', period: '-', status: QAStatus.PLANNED },
      qar: qar || { result: '-', period: '-', status: QAStatus.PLANNED },
      saiv: saiv || { result: '-', period: '-', status: QAStatus.PLANNED }
    }
  })

  const isEditing = ref(false)

  const openForm = () => {
    isEditing.value = false
    resetForm()
    isFormOpen.value = true
  }

  const editReport = () => {
    if (!selectedReport.value) return
    isEditing.value = true
    
    // Populate form
    const periodParts = (selectedReport.value.period || '').split(' ')
    Object.assign(newReport, {
      type: selectedReport.value.type,
      assessmentTitle: selectedReport.value.assessmentTitle,
      periodQuarter: periodParts[0] || '',
      periodYear: periodParts[1] || '2025',
      status: selectedReport.value.status,
      result: selectedReport.value.result,
      internalEvaluator: selectedReport.value.internalEvaluator,
      attachment: null
    })
    
    isDetailOpen.value = false
    isFormOpen.value = true
  }

  const closeForm = () => {
    isFormOpen.value = false
  }

  const openDetail = (report: QAReport) => {
    selectedReport.value = report
    isDetailOpen.value = true
  }

  const closeDetail = () => {
    isDetailOpen.value = false
    selectedReport.value = null
  }

  const deleteReport = async () => {
    if (!selectedReport.value) return
    if (!confirm('Apakah Anda yakin ingin menghapus laporan Quality Assurance ini?')) return
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getMasterServiceBaseUrl()
      await $fetch(`${baseUrl}/quality-assurance/${selectedReport.value.id}`, {
        method: 'DELETE'
      })
      closeDetail()
      await fetchReports()
    } catch (error: any) {
      console.error('Failed to delete QA report:', error)
      errorMsg.value = 'Failed to delete Quality Assurance report.'
    } finally {
      loading.value = false
    }
  }

  return {
    reports, searchQuery, selectedType, selectedPeriod, selectedStatus, isFormOpen, isDetailOpen, columns,
    selectedReport, filteredReports, summary, periods, qaStatuses, qaTypes, page, pageCount, items, newReport,
    handleFileUpload, saveReport, openForm, closeForm, openDetail, closeDetail, getStatusColor, getTypeIconColor,
    editReport, isEditing, deleteReport, fetchReports, loading, errorMsg
  }
})
export { QAType, QAStatus }
