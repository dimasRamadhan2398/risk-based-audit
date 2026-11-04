import { defineStore } from 'pinia'
import { ref, computed, reactive } from 'vue'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { QAStatus, QAType, type QAReport } from '~/types/quality-assurance'
import { useCompanyApi } from '~/composables/useCompanyApi'
import type { Company } from '~/types/master'
import { extractErrorMessage } from '~/utils/error'
import { getFiscalYearStrings } from '~/composables/useFiscalYear'
import { viewDocumentInNewTab, openDocumentTab, renderDocumentInTab } from '~/utils/documentViewer'
import { useI18n } from '~/composables/useI18n'

export const useQualityAssuranceStore = defineStore('quality-assurance', () => {
  const loading = ref(false)
  const errorMsg = ref('')
  const toast = useToastNotification()
  const { t } = useI18n()

  const getMasterServiceBaseUrl = () => {
    const config = useRuntimeConfig()
    return config.public.masterServiceBaseUrl || config.public.apiBase || '/api/v1'
  }

  const columns = computed(() => [
    { accessorKey: 'type', header: t('qualityAssurance.dashboard.table.columns.type'), sortable: true },
    { accessorKey: 'period', header: t('qualityAssurance.dashboard.table.columns.period'), sortable: true },
    { accessorKey: 'reportName', header: t('qualityAssurance.dashboard.table.columns.reportName'), sortable: true },
    { accessorKey: 'result', header: t('qualityAssurance.dashboard.table.columns.result'), sortable: true },
    { accessorKey: 'status', header: t('qualityAssurance.dashboard.table.columns.status'), sortable: true },
    { accessorKey: 'conductedBy', header: t('qualityAssurance.dashboard.table.columns.conductedBy'), sortable: true },
    { accessorKey: 'viewReport', header: t('qualityAssurance.dashboard.table.columns.viewReport') },
    { accessorKey: 'actions', header: t('qualityAssurance.dashboard.table.columns.actions') }
  ])

  const getStatusColor = (status: QAStatus) => {
    switch (status) {
      case QAStatus.COMPLETED: return 'bg-emerald-500'
      case QAStatus.VERIFIED: return 'bg-emerald-500'
      case QAStatus.IN_PROGRESS: return 'bg-amber-400'
      case QAStatus.PLANNED: return 'bg-gray-300'
      default: return 'bg-gray-300'
    }
  }

  const matchQAType = (type: string, target: QAType | string): boolean => {
    if (!type || !target) return false
    if (type === target) return true

    const t = type.toLowerCase().trim()
    const targetStr = target.toLowerCase().trim()

    if (target === QAType.REGULAR || targetStr.includes('regular') || targetStr.includes('rsa')) {
      return t.includes('regular') || t.includes('rsa')
    }
    if (target === QAType.QAR || targetStr.includes('qar') || targetStr.includes('quality assurance review')) {
      return t.includes('qar') || t.includes('quality assurance review')
    }
    if (target === QAType.SAIV || targetStr.includes('saiv') || targetStr.includes('independent validation')) {
      return t.includes('saiv')
    }
    if (target === QAType.IACM || targetStr.includes('iacm') || targetStr.includes('bumn')) {
      return t.includes('iacm')
    }

    return false
  }

  const getTypeIconColor = (type: string) => {
    if (matchQAType(type, QAType.REGULAR)) return 'bg-amber-400'
    if (matchQAType(type, QAType.SAIV)) return 'bg-blue-500'
    if (matchQAType(type, QAType.QAR)) return 'bg-gray-100 border'
    if (matchQAType(type, QAType.IACM)) return 'bg-indigo-500'
    return 'bg-gray-300'
  }

  const page = ref(1)
  const pageCount = 10
  const items = computed(() => {
    return filteredReports.value
  })

  const qaTypes = Object.values(QAType)
  const qaStatuses = Object.values(QAStatus)
  const periods = getFiscalYearStrings()

  const parsePeriodWeight = (period?: string, createdAt?: string): number => {
    let year = 0
    let quarter = 0

    if (period) {
      const yearMatch = period.match(/\b(20\d{2})\b/)
      if (yearMatch && yearMatch[1]) {
        year = parseInt(yearMatch[1], 10)
      }
      const qMatch = period.match(/Q([1-4])/i)
      if (qMatch && qMatch[1]) {
        quarter = parseInt(qMatch[1], 10)
      }
    }

    if (year === 0 && createdAt) {
      const d = new Date(createdAt)
      if (!isNaN(d.getTime())) {
        year = d.getFullYear()
        quarter = Math.floor(d.getMonth() / 3) + 1
      }
    }

    return year * 10 + quarter
  }

  const compareReportsDesc = (a: QAReport, b: QAReport): number => {
    const weightA = parsePeriodWeight(a.period, a.created_at)
    const weightB = parsePeriodWeight(b.period, b.created_at)

    if (weightB !== weightA) {
      return weightB - weightA
    }

    if (a.created_at && b.created_at) {
      const timeA = new Date(a.created_at).getTime()
      const timeB = new Date(b.created_at).getTime()
      if (!isNaN(timeA) && !isNaN(timeB) && timeB !== timeA) {
        return timeB - timeA
      }
    }

    return (b.period || '').localeCompare(a.period || '')
  }

  // Master Data Perusahaan (Company Management Sync)
  const companies = ref<Company[]>([])
  const companiesLoading = ref(false)
  const companiesLoaded = ref(false)
  const companiesError = ref('')

  const formatCompanyLabel = (c: any): string => {
    if (!c) return ''
    const name = c.company_name || c.name || ''
    const code = c.company_code || c.code || ''
    if (name && code) return `${name} (${code})`
    return name || code || c.id || ''
  }

  const resolveCompanyOption = (val?: string): string => {
    if (!val) return ''
    const trimmed = String(val).trim()
    if (!trimmed) return ''
    const match = companies.value.find(
      c => c.id === trimmed ||
        c.company_code === trimmed ||
        (c as any).code === trimmed ||
        c.company_name === trimmed ||
        (c as any).name === trimmed ||
        formatCompanyLabel(c) === trimmed ||
        `${c.company_name || (c as any).name} (${c.company_code || (c as any).code})` === trimmed
    )
    if (match) {
      return formatCompanyLabel(match)
    }
    return trimmed
  }

  const formatCompanyName = (value?: string): string => {
    if (!value) return '-'
    const resolved = resolveCompanyOption(value)
    return resolved || '-'
  }

  const fetchCompanies = async (force = false) => {
    if (companiesLoaded.value && !force) return
    if (companiesLoading.value) return
    companiesLoading.value = true
    companiesError.value = ''
    try {
      const api = useCompanyApi()
      const list = await api.getAllCompanies()
      companies.value = (Array.isArray(list) ? list : [])
        .filter(c => (c?.company_name || (c as any)?.name || c?.company_code || (c as any)?.code) && c.is_active !== false)
        .sort((a, b) => {
          const nameA = a.company_name || (a as any).name || ''
          const nameB = b.company_name || (b as any).name || ''
          return nameA.localeCompare(nameB, 'id', { sensitivity: 'base' })
        })
      companiesLoaded.value = true

      if (newReport.conductedBy) {
        newReport.conductedBy = resolveCompanyOption(newReport.conductedBy)
      }
    } catch (error: any) {
      console.error('Failed to fetch companies for quality assurance:', error)
      companiesError.value = extractErrorMessage(error, 'Gagal memuat data perusahaan.')
      companiesLoaded.value = false
      companies.value = []
    } finally {
      companiesLoading.value = false
    }
  }

  const companyOptions = computed(() => {
    const list = companies.value
      .filter(c => c && c.is_active !== false)
      .map(c => formatCompanyLabel(c))
      .filter(Boolean)
    if (newReport.conductedBy && !list.includes(newReport.conductedBy)) {
      list.unshift(newReport.conductedBy)
    }
    return Array.from(new Set(list)).sort((a, b) => a.localeCompare(b, 'id', { sensitivity: 'base' }))
  })

  const conductedByOptions = companyOptions

  // Form State
  const newReport = reactive({
    type: QAType.REGULAR,
    assessmentTitle: '',
    periodYear: '',
    status: QAStatus.IN_PROGRESS,
    conductedBy: '',
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
      periodYear: '',
      status: QAStatus.IN_PROGRESS,
      conductedBy: '',
      result: '',
      internalEvaluator: '',
      attachment: null
    })
  }

  const reports = ref<QAReport[]>([])

  const fetchReports = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getMasterServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/quality-assurance`, {
        method: 'GET'
      })
      if (Array.isArray(response)) {
        reports.value = response
      } else {
        reports.value = []
      }
      if (!companiesLoaded.value && !companiesLoading.value) {
        fetchCompanies().catch(() => {})
      }
    } catch (error: any) {
      console.error('Failed to fetch QA reports:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to fetch quality assurance reports.')
      reports.value = []
    } finally {
      loading.value = false
    }
  }

  // Load on init
  fetchReports()
  fetchCompanies().catch(() => {})


  const saveReport = async () => {
    if (!newReport.assessmentTitle || !newReport.result) return
    loading.value = true
    errorMsg.value = ''

    const reportData = {
      type: newReport.type,
      period: newReport.periodYear || new Date().getFullYear().toString(),
      reportName: newReport.assessmentTitle,
      result: newReport.result,
      status: newReport.status,
      conductedBy: newReport.conductedBy,
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
        toast.showSuccess(t('qualityAssurance.dashboard.updateSuccess'))
      } else {
        await $fetch(`${baseUrl}/quality-assurance`, {
          method: 'POST',
          body: reportData
        })
        toast.showSuccess(t('qualityAssurance.dashboard.createSuccess'))
      }
      resetForm()
      isEditing.value = false
      closeForm()
      await fetchReports()
    } catch (error: any) {
      console.error('Failed to save QA report:', error)
      errorMsg.value = extractErrorMessage(error, t('qualityAssurance.dashboard.saveFailed'))
      toast.showError(t('qualityAssurance.dashboard.saveFailed'))
    } finally {
      loading.value = false
    }
  }

  const searchQuery = ref('')
  const selectedType = ref('')
  const selectedPeriod = ref('')
  const selectedStatus = ref('')

  const resetFilters = () => {
    searchQuery.value = ''
    selectedType.value = ''
    selectedPeriod.value = ''
    selectedStatus.value = ''
    page.value = 1
  }

  const isFormOpen = ref(false)
  const isImportOpen = ref(false)
  const isDetailOpen = ref(false)
  const selectedReport = ref<QAReport | null>(null)

  const openImportModal = () => {
    isImportOpen.value = true
    errorMsg.value = ''
  }

  const closeImportModal = () => {
    isImportOpen.value = false
  }

  const importQARReport = async (payload: {
    assessmentTitle: string
    type: string
    periodYear: string
    result: string
    status: string
    conductedBy: string
    validator: string
    fileName: string
    fileType: string
    file: File
  }) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getMasterServiceBaseUrl()
      const formData = new FormData()
      formData.append('assessmentTitle', payload.assessmentTitle)
      formData.append('type', payload.type)
      formData.append('periodYear', payload.periodYear)
      formData.append('result', payload.result)
      formData.append('status', payload.status)
      formData.append('conductedBy', payload.conductedBy)
      formData.append('validator', payload.validator)
      formData.append('fileName', payload.fileName)
      formData.append('fileType', payload.fileType)
      formData.append('file', payload.file)
      const response: any = await $fetch(`${baseUrl}/quality-assurance/import`, {
        method: 'POST',
        body: formData
      })
      await fetchReports()
      closeImportModal()
      toast.showSuccess(t('qualityAssurance.dashboard.importSuccess'))
      return response
    } catch (error: any) {
      console.error('Failed to import QAR report:', error)
      errorMsg.value = extractErrorMessage(error, t('qualityAssurance.dashboard.importFailed'))
      throw error
    } finally {
      loading.value = false
    }
  }

  const downloadAttachment = async (id: string, fileName: string) => {
    const report = reports.value.find(r => r.id === id)
    if (report?.attachment?.filePath && report.attachment.filePath.startsWith('blob:')) {
      const link = document.createElement('a')
      link.href = report.attachment.filePath
      link.download = fileName
      link.click()
      return
    }

    try {
      const baseUrl = getMasterServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/quality-assurance/${id}/download`, {
        responseType: 'blob'
      })
      const blob = new Blob([response], { type: response.type || 'application/octet-stream' })
      const link = document.createElement('a')
      link.href = window.URL.createObjectURL(blob)
      link.download = fileName
      link.click()
      window.URL.revokeObjectURL(link.href)
    } catch (error: any) {
      console.error('Failed to download QAR attachment:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to download QAR attachment.')
    }
  }

  const viewDocument = async (id: string, fileName?: string, title?: string) => {
    const report = reports.value.find(r => r.id === id)
    const resolvedName = fileName || report?.attachment?.name || 'document.pdf'
    const resolvedFileName = fileName || report?.attachment?.name || 'document.pdf'

    if (report?.attachment?.filePath && report.attachment.filePath.startsWith('blob:')) {
      const tab = openDocumentTab(resolvedName)
      try {
        const response = await fetch(report.attachment.filePath)
        const blob = await response.blob()
        renderDocumentInTab(tab, blob, resolvedName, resolvedFileName)
      } catch (err) {
        if (tab) tab.close()
        console.error('Failed to view blob document:', err)
      }
      return
    }

    try {
      const baseUrl = getMasterServiceBaseUrl()
      await viewDocumentInNewTab(
        async () => {
          try {
            return await $fetch(`${baseUrl}/quality-assurance/${id}/view`, {
              responseType: 'blob'
            })
          } catch (_err) {
            return await $fetch(`${baseUrl}/quality-assurance/${id}/download`, {
              responseType: 'blob'
            })
          }
        },
        resolvedName,
        resolvedFileName
      )
    } catch (error: any) {
      console.error('Failed to view QAR attachment:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to view QAR attachment.')
    }
  }

  const filteredReports = computed(() => {
    return reports.value
      .filter(report => {
        if (report.isImported) return false

        const matchesSearch = (report.reportName || '').toLowerCase().includes(searchQuery.value.toLowerCase()) ||
          (report.assessmentTitle || '').toLowerCase().includes(searchQuery.value.toLowerCase())
        const matchesType = !selectedType.value || matchQAType(report.type, selectedType.value)
        const matchesPeriod = !selectedPeriod.value || (report.period || '').includes(selectedPeriod.value)
        const matchesStatus = !selectedStatus.value || report.status === selectedStatus.value
        return matchesSearch && matchesType && matchesPeriod && matchesStatus
      })
      .sort(compareReportsDesc)
  })

  const importedReports = computed(() => {
    return reports.value.filter(report => report.isImported)
  })

  const regularImportedReports = computed(() => {
    return reports.value.filter(report => report.isImported && matchQAType(report.type, QAType.REGULAR))
  })

  const saivImportedReports = computed(() => {
    return reports.value.filter(report => report.isImported && matchQAType(report.type, QAType.SAIV))
  })

  const qarImportedReports = computed(() => {
    return reports.value.filter(report => report.isImported && matchQAType(report.type, QAType.QAR))
  })

  const iacmImportedReports = computed(() => {
    return reports.value.filter(report => report.isImported && matchQAType(report.type, QAType.IACM))
  })

  const summary = computed(() => {
    // Only non-imported reports (the exact reports displayed in QATable)
    const tableReports = reports.value.filter(r => !r.isImported)

    const getLatestForType = (targetType: QAType) => {
      // 1. If matching in filteredReports (respects period, search, status filters)
      const inFiltered = filteredReports.value.filter(r => matchQAType(r.type, targetType))
      if (inFiltered.length > 0) {
        return inFiltered.slice().sort(compareReportsDesc)[0]
      }

      // 2. If filtered out by selectedType filter, check matching tableReports with other filters
      const inTable = tableReports.filter(r => matchQAType(r.type, targetType))
      const withOtherFilters = inTable.filter(r => {
        const matchesSearch = !searchQuery.value ||
          (r.reportName || '').toLowerCase().includes(searchQuery.value.toLowerCase()) ||
          (r.assessmentTitle || '').toLowerCase().includes(searchQuery.value.toLowerCase())
        const matchesPeriod = !selectedPeriod.value || (r.period || '').includes(selectedPeriod.value)
        const matchesStatus = !selectedStatus.value || r.status === selectedStatus.value
        return matchesSearch && matchesPeriod && matchesStatus
      })
      if (withOtherFilters.length > 0) {
        return withOtherFilters.slice().sort(compareReportsDesc)[0]
      }

      // 3. Fallback to latest non-imported report of this type
      if (inTable.length > 0) {
        return inTable.slice().sort(compareReportsDesc)[0]
      }

      return null
    }

    const regular = getLatestForType(QAType.REGULAR)
    const qar = getLatestForType(QAType.QAR)
    const saiv = getLatestForType(QAType.SAIV)
    const iacm = getLatestForType(QAType.IACM)

    return {
      regular: regular || { result: '-', period: '-', status: QAStatus.PLANNED },
      qar: qar || { result: '-', period: '-', status: QAStatus.PLANNED },
      saiv: saiv || { result: '-', period: '-', status: QAStatus.PLANNED },
      iacm: iacm || { result: '-', period: '-', status: QAStatus.PLANNED }
    }
  })

  const isEditing = ref(false)

  const openForm = () => {
    isEditing.value = false
    resetForm()
    if (!companiesLoaded.value && !companiesLoading.value) {
      fetchCompanies().catch(() => {})
    }
    isFormOpen.value = true
  }

  const editReport = (report?: QAReport) => {
    if (report) {
      selectedReport.value = report
    }

    if (!selectedReport.value) return

    isEditing.value = true

    Object.assign(newReport, {
      type: selectedReport.value.type,
      assessmentTitle: selectedReport.value.assessmentTitle,
      periodYear: selectedReport.value.period || new Date().getFullYear().toString(),
      status: selectedReport.value.status,
      conductedBy: resolveCompanyOption(selectedReport.value.conductedBy),
      result: selectedReport.value.result,
      internalEvaluator: selectedReport.value.internalEvaluator || '',
      attachment: null
    })

    if (!companiesLoaded.value && !companiesLoading.value) {
      fetchCompanies().catch(() => {})
    }

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

  const deleteReport = async (report?: QAReport) => {
    const targetReport = report || selectedReport.value

    if (!targetReport) return

    loading.value = true
    errorMsg.value = ''

    try {
      const baseUrl = getMasterServiceBaseUrl()

      await $fetch(`${baseUrl}/quality-assurance/${targetReport.id}`, {
        method: 'DELETE'
      })

      if (selectedReport.value?.id === targetReport.id) {
        isDetailOpen.value = false
        selectedReport.value = null
      }

      await fetchReports()
      toast.showSuccess(t('qualityAssurance.dashboard.deleteSuccess'))
    } catch (error: any) {
      console.error('Failed to delete QA report:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to delete Quality Assurance report.')
      toast.showError('Gagal menghapus data laporan QA')
    } finally {
      loading.value = false
    }
  }

  return {
    reports, searchQuery, selectedType, selectedPeriod, selectedStatus, isFormOpen, isImportOpen, isDetailOpen, columns,
    selectedReport, filteredReports, importedReports, regularImportedReports, saivImportedReports, qarImportedReports, iacmImportedReports, summary, periods, qaStatuses, qaTypes, page, pageCount, items, newReport,
    handleFileUpload, saveReport, openForm, closeForm, openImportModal, closeImportModal, importQARReport, downloadAttachment, viewDocument, getMasterServiceBaseUrl, openDetail, closeDetail, getStatusColor, getTypeIconColor, matchQAType,
    editReport, isEditing, deleteReport, fetchReports, resetFilters, loading, errorMsg,
    companies, companiesLoading, companiesLoaded, companiesError, fetchCompanies, resolveCompanyOption, formatCompanyName, companyOptions, conductedByOptions
  }
})
export { QAType, QAStatus }
