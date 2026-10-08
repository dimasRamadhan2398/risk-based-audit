import { defineStore } from 'pinia'
import { ref, computed, reactive, watch } from 'vue'
import { useAssignmentLetterStore } from './assignment-letter'
import { useWorkingPaperStore } from './working-paper'
import { useAuditFieldworkStore } from './audit-fieldwork'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { extractErrorMessage } from '~/utils/error'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'

export type FindingCategory = 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant'

export const FINDING_CATEGORIES: FindingCategory[] = ['Very Significant', 'Significant', 'Quite Significant', 'Not Significant']

export interface FindingItem {
  title: string
  category: FindingCategory
  action?: string
  source?: string
  impact?: string
  criteria?: string
}

export interface AuditResultReport {
  id: string
  reportNumber: string
  assignmentLetterId: string
  reportTitle: string
  executiveSummary: string
  findings?: FindingItem[]
  reportDate: string
  status: 'Draft' | 'Final'
  findingsCount: number
  category?: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant'
  department?: string
  companyId?: string
  companyName?: string
}

// GET /audit-result-reports/recent-findings item
export interface RecentFinding {
  title: string
  category: FindingCategory
  action: string
  source: string
  assignmentLetterId: string
  reportId: string | null
  date: string
}

export const normalizeReportNumber = (raw?: string | null): string => {
  if (!raw) return ''
  const trimmed = raw.trim()
  if (!trimmed) return ''

  // Already in target format: LHA-xxx/TEAM/yyyy
  if (/^LHA-\d+\/[^\/]+\/\d{4}$/i.test(trimmed)) {
    return trimmed
  }

  // Legacy full pattern: 021/LHA/01/KS IAD/2026
  const fullLegacy = trimmed.match(/^(\d+)\/LHA\/\d+\/(?:KS\s*IAD|[^\/]+)\/(\d{4})$/i)
  if (fullLegacy && fullLegacy[1] && fullLegacy[2]) {
    const seq = fullLegacy[1].padStart(3, '0')
    const year = fullLegacy[2]
    return `LHA-${seq}/SKAI/${year}`
  }

  // Generic /LHA/ pattern: (\d+)/LHA/...
  if (trimmed.includes('/LHA/')) {
    const parts = trimmed.split('/')
    if (parts.length >= 2) {
      const seq = (parts[0] || '').replace(/\D/g, '').padStart(3, '0')
      let year = new Date().getFullYear().toString()
      const lastPart = parts[parts.length - 1]
      if (lastPart && /^\d{4}$/.test(lastPart)) {
        year = lastPart
      }
      return `LHA-${seq}/SKAI/${year}`
    }
  }

  return trimmed
}

export const useAuditResultReportStore = defineStore('audit-result-report', () => {
  const assignmentLetterStore = useAssignmentLetterStore()
  const toast = useToastNotification()

  // State
  const selectedAssignmentLetter = ref<string>('')
  const reportList = ref<AuditResultReport[]>([])
  const showModal = ref(false)
  const isEditing = ref(false)
  const editingId = ref<string | null>(null)

  // Field validation errors for the ARR form, as i18n keys ('' = no error).
  const formErrors = reactive({ assignmentLetterId: '' })

  const reportForm = reactive({
    reportNumber: '',
    assignmentLetterId: '',
    reportTitle: '',
    reportDate: new Date().toISOString().split('T')[0] as string,
    status: 'Draft' as 'Draft' | 'Final',
    findingsCount: 0,
    findings: [] as FindingItem[],
    companyId: '',
    companyName: ''
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

  const loading = ref(false)
  const errorMsg = ref('')

  const normalizeFindingCategory = (raw: unknown): FindingCategory => {
    if (raw === 'Moderately Significant') return 'Quite Significant'
    if (raw === 'Insignificant') return 'Not Significant'
    if (FINDING_CATEGORIES.includes(raw as FindingCategory)) return raw as FindingCategory
    return 'Quite Significant'
  }

  const generateReportNumber = (dateStr?: string, auditTeam: string = 'SKAI'): string => {
    let year = new Date().getFullYear().toString()
    if (dateStr) {
      const cleanDate = (dateStr.split('T')[0] || '').trim()
      const parts = cleanDate.split('-')
      if (parts.length === 3 && parts[0]) {
        year = parts[0]
      } else {
        const d = new Date(dateStr)
        if (!isNaN(d.getFullYear())) {
          year = d.getFullYear().toString()
        }
      }
    }

    let team = auditTeam || 'SKAI'
    if (team === 'SKAI') {
      const letter = reportForm.assignmentLetterId || selectedAssignmentLetter.value
      if (letter) {
        const st = assignmentLetterStore.assignmentLetterList.find(
          (s: any) => s.letterNumber === letter || s.id === letter
        )
        if (st && (st as any).auditTeam) {
          team = (st as any).auditTeam
        } else if (letter.includes('/')) {
          const parts = letter.split('/')
          if (parts[1]) team = parts[1]
        }
      }
    }

    let maxSeq = 20
    const lhaRegex = /(?:LHA-|^)(\d+)/i
    for (const r of reportList.value) {
      const num = r.reportNumber || (r as any).report_number
      if (num) {
        const match = num.match(lhaRegex)
        if (match && match[1]) {
          const parsed = parseInt(match[1], 10)
          if (!isNaN(parsed) && parsed > maxSeq) {
            maxSeq = parsed
          }
        }
      }
    }
    const nextSeq = (maxSeq + 1).toString().padStart(3, '0')
    return `LHA-${nextSeq}/${team.toUpperCase()}/${year}`
  }

  const mapReportItem = (item: any): AuditResultReport => {
    let dateVal = item.reportDate || item.report_date || item.created_at || ''
    if (typeof dateVal === 'string' && dateVal.includes('T')) {
      dateVal = dateVal.split('T')[0]
    }

    const findingsArr = item.findings || item.Findings || []

    // Map legacy severity to category
    const mappedFindings = findingsArr.map((f: any) => ({
      ...f,
      category: normalizeFindingCategory(f.category || f.severity),
      source: f.source || 'Audit Features'
    }))

    const finalReportDate = dateVal || new Date().toISOString().split('T')[0]
    const rawNum = item.reportNumber || item.report_number || ''
    const normalizedNum = normalizeReportNumber(rawNum) || generateReportNumber(finalReportDate)

    return {
      ...item,
      reportNumber: normalizedNum,
      report_number: normalizedNum,
      findingsCount: item.findingsCount || item.findings_count || mappedFindings.length || 0,
      findings: mappedFindings,
      reportDate: finalReportDate,
      companyId: item.companyId || item.company_id || '',
      companyName: item.companyName || item.company_name || ''
    }
  }

  const fetchReports = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/audit-result-reports`, { method: 'GET' })
      let items: any[] = []
      if (response && response.data && Array.isArray(response.data.items)) {
        items = response.data.items
      } else if (response && Array.isArray(response.items)) {
        items = response.items
      } else if (Array.isArray(response)) {
        items = response
      }

      reportList.value = items.map(mapReportItem)
    } catch (error) {
      console.error('Failed to fetch reports:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to load audit result reports.')
      reportList.value = []
    } finally {
      loading.value = false
    }
  }

  // Fetch on initialization
  fetchReports()

  // Dashboard "Recent Finding Issues": saved ARR findings merged server-side with
  // live KKA / fieldwork findings, deduped and sorted newest first.
  const recentFindings = ref<RecentFinding[]>([])
  const recentFindingsLoading = ref(false)
  const recentFindingsError = ref('')

  const fetchRecentFindings = async (limit = 5) => {
    recentFindingsLoading.value = true
    recentFindingsError.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const res = await $fetch<{ data?: { items?: Partial<RecentFinding>[] } }>(
        `${baseUrl}/audit-result-reports/recent-findings?limit=${limit}`,
        { method: 'GET' }
      )
      const items = Array.isArray(res?.data?.items) ? res.data.items : []
      recentFindings.value = items
        .map((f): RecentFinding => ({
          title: f.title || '',
          category: normalizeFindingCategory(f.category),
          action: f.action || '',
          source: f.source || '',
          assignmentLetterId: f.assignmentLetterId || '',
          reportId: f.reportId ?? null,
          date: f.date || ''
        }))
        .filter(f => f.title)
        // The backend already sorts; keep newest first even if it ever doesn't.
        .sort((a, b) => (Date.parse(b.date) || 0) - (Date.parse(a.date) || 0))
        .slice(0, limit)
    } catch (error) {
      console.error('Failed to fetch recent findings:', error)
      recentFindingsError.value = extractErrorMessage(error, 'Failed to load recent findings.')
      recentFindings.value = []
    } finally {
      recentFindingsLoading.value = false
    }
  }

  const isAutoDetecting = ref(false)

  // Local fallback aggregation from Working Paper and Fieldwork stores
  const getLocalAutoFindings = (targetLetter: string): FindingItem[] => {
    const findings: FindingItem[] = []
    const existingTitles = new Set<string>()

    const wpStore = useWorkingPaperStore()
    const fwStore = useAuditFieldworkStore()

    // 1. Digital Working Paper (AOI & RCA F04 + Plan F05 + Risk F02) — loaded data only
    const allCauses: any[] = wpStore.dataF04 || []
    const allPlans: any[] = wpStore.dataF05 || []
    const allRisks: any[] = wpStore.dataF02 || []

    const stCauses = allCauses.filter((c: any) => (c.workingPaperId || c.assignmentLetterId) === targetLetter)
    const stPlans = allPlans.filter((p: any) => (p.workingPaperId || p.assignmentLetterId) === targetLetter)
    const stRisks = allRisks.filter((r: any) => (r.workingPaperId || r.assignmentLetterId) === targetLetter)

    let defaultCat: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant' = 'Significant'
    for (const r of stRisks) {
      const lvl = String(r.riskLevel || '').toUpperCase()
      if (lvl === 'HIGH' || lvl === 'CRITICAL') {
        defaultCat = 'Very Significant'
        break
      } else if (lvl === 'MODERATE' || lvl === 'MEDIUM') {
        defaultCat = 'Significant'
      } else if (lvl === 'LOW') {
        defaultCat = 'Quite Significant'
      }
    }

    stCauses.forEach((cause: any, idx: number) => {
      const cond = (cause.condition || '').trim()
      if (!cond) return
      const key = cond.toLowerCase()
      if (existingTitles.has(key)) return
      existingTitles.add(key)

      let action = ''
      if (stPlans[idx]) {
        action = stPlans[idx]?.actionDescription || stPlans[idx]?.recommendation || ''
      } else if (stPlans.length > 0 && stPlans[0]) {
        action = stPlans[0]?.actionDescription || stPlans[0]?.recommendation || ''
      }

      let cat = defaultCat
      const condLower = cond.toLowerCase()
      if (condLower.includes('kritis') || condLower.includes('overhaul') || condLower.includes('mfa') || condLower.includes('override') || condLower.includes('transisi')) {
        cat = 'Very Significant'
      }

      findings.push({
        title: cond,
        category: cat,
        action,
        source: 'Digital Working Paper (KKA - AOI & RCA)',
        impact: cause.impact || '',
        criteria: cause.criteria || ''
      })
    })

    // 2. Audit Fieldwork Test Controls — loaded data only
    const testControlsList: any[] = fwStore.fieldworkData?.[targetLetter]?.testControls || []

    testControlsList.forEach((tc: any) => {
      const findingText = (tc.finding || '').trim()
      const resultUpper = (tc.testResult || '').toUpperCase()

      if (findingText || resultUpper === 'INEFFECTIVE' || resultUpper === 'PARTIALLY EFFECTIVE') {
        const title = findingText || `Kelemahan Kontrol: ${tc.controlName || 'Internal Control'}`
        const key = title.toLowerCase()
        if (existingTitles.has(key)) return
        existingTitles.add(key)

        const action = (tc.mitigationPlan || tc.recommendation || '').trim()
        let cat: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant' = 'Significant'
        if (resultUpper === 'INEFFECTIVE') {
          cat = 'Very Significant'
        } else if (resultUpper === 'PARTIALLY EFFECTIVE') {
          cat = 'Significant'
        }

        findings.push({
          title,
          category: cat,
          action,
          source: 'Audit Fieldwork (Test Controls)'
        })
      }
    })

    // Fallback to pre-existing reports if still empty
    if (findings.length === 0) {
      const existingReport = reportList.value.find(r => r.assignmentLetterId === targetLetter)
      if (existingReport && existingReport.findings && existingReport.findings.length > 0) {
        return existingReport.findings.map(f => ({
          ...f,
          source: (f as any).source || 'Audit Record'
        }))
      }
    }

    return findings
  }

  // Fetch auto-findings from Backend with fallback to local stores
  const fetchAutoFindings = async (stNumber?: string): Promise<FindingItem[]> => {
    const targetLetter = stNumber || selectedAssignmentLetter.value || reportForm.assignmentLetterId
    if (!targetLetter) return []

    isAutoDetecting.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const res: any = await $fetch(`${baseUrl}/audit-result-reports/auto-findings?assignmentLetterId=${encodeURIComponent(targetLetter)}`, {
        method: 'GET'
      })
      if (res && res.data && Array.isArray(res.data.findings) && res.data.findings.length > 0) {
        return res.data.findings.map((f: any) => ({
          title: f.title,
          category: f.category || 'Significant',
          action: f.action || '',
          source: f.source || 'Audit Features',
          impact: f.impact || '',
          criteria: f.criteria || ''
        }))
      }
    } catch (err) {
      console.warn('Backend auto-findings API not reachable or returned empty, falling back to local audit stores:', err)
    } finally {
      isAutoDetecting.value = false
    }

    return getLocalAutoFindings(targetLetter)
  }

  const autoPopulateFindings = async (stNumber?: string) => {
    const targetLetter = stNumber || selectedAssignmentLetter.value || reportForm.assignmentLetterId
    if (!targetLetter) return
    const autoFindings = await fetchAutoFindings(targetLetter)
    if (autoFindings.length > 0) {
      reportForm.findings = JSON.parse(JSON.stringify(autoFindings))
      reportForm.findingsCount = autoFindings.length
    }
    const stData = assignmentLetterStore.assignmentLetterList.find(
      (st: any) => st.letterNumber === targetLetter
    )
    if (stData) {
      if (!reportForm.companyId && (stData as any).companyId) {
        reportForm.companyId = (stData as any).companyId
      }
      if (!reportForm.companyName && (stData as any).companyName) {
        reportForm.companyName = (stData as any).companyName
      }
    }
  }

  const runAutoDetectFindings = async (mode: 'replace' | 'merge' = 'replace') => {
    const targetLetter = reportForm.assignmentLetterId || selectedAssignmentLetter.value
    if (!targetLetter) {
      toast.showWarning('Peringatan', 'Silakan pilih Assignment Letter terlebih dahulu.')
      return
    }
    const detected = await fetchAutoFindings(targetLetter)
    if (detected.length === 0) {
      toast.showWarning('Informasi', `Tidak ada temuan audit baru yang terdeteksi untuk ${targetLetter}.`)
      return
    }

    if (mode === 'replace' || !reportForm.findings || reportForm.findings.length === 0) {
      reportForm.findings = JSON.parse(JSON.stringify(detected))
    } else {
      const existing = new Set(reportForm.findings.map(f => f.title.toLowerCase().trim()))
      detected.forEach((d) => {
        if (!existing.has(d.title.toLowerCase().trim())) {
          reportForm.findings.push(JSON.parse(JSON.stringify(d)))
          existing.add(d.title.toLowerCase().trim())
        }
      })
    }
    reportForm.findingsCount = reportForm.findings.length
    toast.showSuccess('Temuan Terisi Otomatis', `${detected.length} temuan berhasil ditarik dari modul KKA & Fieldwork.`)
  }

  // Actions
  const openModal = async () => {
    resetForm()
    isEditing.value = false
    editingId.value = null
    showModal.value = true

    if (selectedAssignmentLetter.value) {
      reportForm.assignmentLetterId = selectedAssignmentLetter.value
      const stData = assignmentLetterStore.assignmentLetterList.find(
        (st: any) => st.letterNumber === selectedAssignmentLetter.value
      )
      reportForm.reportNumber = generateReportNumber(reportForm.reportDate, stData?.auditTeam || 'SKAI')
      if (stData?.auditTitle) {
        reportForm.reportTitle = `Laporan Hasil Audit - ${stData.auditTitle}`
      } else {
        reportForm.reportTitle = `Laporan Hasil Audit - ${selectedAssignmentLetter.value}`
      }
      if (stData) {
        reportForm.companyId = (stData as any).companyId || ''
        reportForm.companyName = (stData as any).companyName || ''
      }
      await autoPopulateFindings(selectedAssignmentLetter.value)
    }
  }

  const closeModal = () => {
    showModal.value = false
  }

  const resetForm = () => {
    const defaultDate = new Date().toISOString().split('T')[0] as string
    formErrors.assignmentLetterId = ''
    Object.assign(reportForm, {
      reportNumber: generateReportNumber(defaultDate),
      assignmentLetterId: selectedAssignmentLetter.value || 'ST-001/SKAI/2026',
      reportTitle: '',
      reportDate: defaultDate,
      status: 'Draft',
      findingsCount: 0,
      findings: [],
      companyId: '',
      companyName: ''
    })
  }

  const saveReport = async () => {
    // A report always belongs to an assignment letter. Editing keeps the report's
    // own letter (the field is disabled); creating may fall back to the page filter.
    const letter = String(reportForm.assignmentLetterId || (isEditing.value ? '' : selectedAssignmentLetter.value) || '').trim()
    if (!letter) {
      formErrors.assignmentLetterId = 'auditResultReport.form.assignmentLetterRequired'
      return
    }
    formErrors.assignmentLetterId = ''
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      if (!reportForm.reportNumber) {
        reportForm.reportNumber = generateReportNumber(reportForm.reportDate)
      }
      const payload: any = {
        assignmentLetterId: letter,
        reportTitle: reportForm.reportTitle,
        reportDate: reportForm.reportDate,
        report_date: reportForm.reportDate,
        reportNumber: reportForm.reportNumber,
        findingsCount: Number(reportForm.findings?.length || reportForm.findingsCount || 0),
        findings: (reportForm.findings || []).map(f => ({
          title: f.title,
          category: f.category,
          action: f.action || '',
          source: (f as any).source || 'Audit Features'
        })),
        status: reportForm.status
      }
      if (reportForm.companyId) {
        payload.companyId = reportForm.companyId
        payload.company_id = reportForm.companyId
      }
      if (reportForm.companyName) {
        payload.companyName = reportForm.companyName
        payload.company_name = reportForm.companyName
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
      toast.showSuccess('Report saved successfully')
    } catch (error: any) {
      console.error('Failed to save report:', error)
      const detail = extractErrorMessage(error, 'Failed to save report.')
      errorMsg.value = detail
      toast.showError('Failed to save report', detail)
    } finally {
      loading.value = false
    }
  }

  const editReport = (report: AuditResultReport) => {
    formErrors.assignmentLetterId = ''
    Object.assign(reportForm, {
      ...report,
      reportNumber: report.reportNumber || (report as any).report_number || generateReportNumber(report.reportDate),
      companyId: report.companyId || '',
      companyName: report.companyName || '',
      findings: report.findings ? JSON.parse(JSON.stringify(report.findings)) : []
    })
    isEditing.value = true
    editingId.value = report.id
    showModal.value = true
  }

  const deleteReport = async (id: string) => {
    if (!await useGlobalModalStore().confirmDelete({ description: 'Are you sure you want to delete this report?' })) return
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/audit-result-reports/${id}`, {
        method: 'DELETE'
      })
      await fetchReports()
      toast.showSuccess('Report deleted successfully')
    } catch (error: any) {
      console.error('Failed to delete report:', error)
      const detail = extractErrorMessage(error, 'Failed to delete report.')
      errorMsg.value = detail
      toast.showError('Failed to delete report', detail)
    } finally {
      loading.value = false
    }
  }

  const downloadDocx = async (id: string, reportNumber: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const blob = await $fetch<Blob>(`${baseUrl}/audit-result-reports/${id}/download-docx`, {
        method: 'GET',
        responseType: 'blob'
      })
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      const cleanNum = (reportNumber || 'LHA').replace(/[\/\s]/g, '_')
      a.download = `LHA_${cleanNum}.docx`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      window.URL.revokeObjectURL(url)
    } catch (err: any) {
      console.error('Failed to download docx:', err)
      const detail = extractErrorMessage(err, 'Gagal mengunduh dokumen Word LHA.')
      toast.showError('Gagal mengunduh dokumen Word LHA', detail)
    }
  }

  const syncExecutiveSummaryField = async (reportNumber: string, narrative: string, findingsCount?: number) => {
    const normTarget = normalizeReportNumber(reportNumber)
    const found = reportList.value.find(r => {
      const num = r.reportNumber || (r as any).report_number
      return num === reportNumber || normalizeReportNumber(num) === normTarget
    })
    if (found) {
      found.executiveSummary = narrative
      if (typeof findingsCount === 'number' && findingsCount > 0) {
        found.findingsCount = findingsCount
      }
      try {
        const baseUrl = getAuditServiceBaseUrl()
        await $fetch(`${baseUrl}/audit-result-reports/${found.id}`, {
          method: 'PUT',
          body: {
            executive_summary: narrative,
            findingsCount: found.findingsCount
          }
        })
      } catch (e) {
        console.warn('Silent sync to backend failed:', e)
      }
    }
  }

  const clearExecutiveSummaryField = async (reportNumber: string) => {
    const normTarget = normalizeReportNumber(reportNumber)
    const found = reportList.value.find(r => {
      const num = r.reportNumber || (r as any).report_number
      return num === reportNumber || normalizeReportNumber(num) === normTarget
    })
    if (found) {
      found.executiveSummary = ''
      try {
        const baseUrl = getAuditServiceBaseUrl()
        await $fetch(`${baseUrl}/audit-result-reports/${found.id}`, {
          method: 'PUT',
          body: {
            executive_summary: ''
          }
        })
      } catch (e) {
        console.warn('Silent clear executive summary failed:', e)
      }
    }
  }

  return {
    selectedAssignmentLetter,
    reportList,
    showModal,
    isEditing,
    reportForm,
    formErrors,
    publishedAssignmentLetters,
    filteredReports,
    hasSelectedAssignmentLetter,
    openModal,
    closeModal,
    saveReport,
    editReport,
    deleteReport,
    downloadDocx,
    syncExecutiveSummaryField,
    clearExecutiveSummaryField,
    loading,
    errorMsg,
    fetchReports,
    recentFindings,
    recentFindingsLoading,
    recentFindingsError,
    fetchRecentFindings,
    isAutoDetecting,
    fetchAutoFindings,
    runAutoDetectFindings,
    autoPopulateFindings,
    generateReportNumber,
    normalizeReportNumber,
    resetForm
  }
})
