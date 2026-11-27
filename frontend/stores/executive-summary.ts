import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { extractErrorMessage } from '~/utils/error'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { useAuditResultReportStore } from './audit-result-report'
import { useI18n } from '~/composables/useI18n'
import {
  buildDefaultNarrative,
  buildSyncedNarrative,
  deriveNarrativeSyncData,
  formatQuarterPeriod,
  isAutoSyncableNarrative,
  type NarrativeSourceInput
} from '~/utils/executiveSummaryNarrative'
import {
  canGenerateExecutiveSummaryDocx,
  executiveSummaryDocxFilename,
  generateExecutiveSummaryDocxBytes,
  isCompilationDocNumber,
  resolveExecutiveSummaryVariant,
  type ExecutiveSummaryDocxContext,
  type ExecutiveSummaryDocxVariant
} from '~/utils/executiveSummaryDocx'
import type { ExecutiveSummaryApprovePayload } from '~/types/signature'
import { useDepartmentApi } from '~/composables/useDepartmentApi'
import type { Department } from '~/types/master'
import {
  followUpStatusParams,
  parseFollowUpStatus,
  type FollowUpStatusQuery,
  type FollowUpStatusResult,
  type NoActionReason
} from '~/utils/executiveSummaryFollowUp'

/** Follow-up status of a finding, from the bucket of its Action Taken Report (ATR). */
export type FollowUpStatusValue = 'Closed' | 'In Progress' | 'Overdue' | 'No Action'

/** Default follow-up table: one row per ATR bucket. Documents saved before No Action existed keep their three rows. */
export const defaultFollowUpTable = (): FollowUpRow[] => [
  { status: 'Closed', jumlah: 0, persentase: 0, keterangan: '' },
  { status: 'In Progress', jumlah: 0, persentase: 0, keterangan: '' },
  { status: 'Overdue', jumlah: 0, persentase: 0, keterangan: '' },
  { status: 'No Action', jumlah: 0, persentase: 0, keterangan: '' }
]

/**
 * Follow-up rows of a draft completed with any missing status (count 0), in the default order.
 * The backend freezes four rows on approval only when the saved draft already has a No Action row.
 */
export const withAllFollowUpRows = (rows: FollowUpRow[] | null | undefined): FollowUpRow[] => {
  const list = Array.isArray(rows) ? rows : []
  return defaultFollowUpTable().map(def => list.find(r => r.status === def.status) ?? def)
}

export interface FollowUpRow {
  status: FollowUpStatusValue
  jumlah: number
  persentase: number
  keterangan: string
}

export interface SignificantFinding {
  unitDivision: string
  judulTemuan: string
  risiko: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant' | 'Tinggi' | 'Sedang' | 'Rendah' | string
  category?: string
  significanceLevel?: string
  statusTL: FollowUpStatusValue
  usulan: string
  /** LHA finding the row was taken from; when set, statusTL follows the ATR of that finding */
  findingId?: string
  atrId?: string | null
  noActionReason?: NoActionReason | null
}

export interface MatriksRow {
  nomor: string
  division: string
  unitKerja: string
  prosesBisnis: string
  judulTemuan: string
  nilaiRisiko: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant' | 'Tinggi' | 'Sedang' | 'Rendah' | string
  rekomendasi: string
  dueDate: string
  picUnit: string
  progres: number
  status: FollowUpStatusValue
  buktiTL: string
  findingId?: string
  atrId?: string | null
  noActionReason?: NoActionReason | null
  auditResultReportId?: string
}

export interface ExecutiveSummary {
  id: string
  assignmentLetterId?: string
  quarter: number // 1, 2, 3, 4
  periodeBulan: string // e.g. "Maret"
  tahun: number // default 2026
  nomorDokumen: string
  dokumenPath: string
  status: 'Draft' | 'Approved' | 'Rejected'
  executiveNote?: string
  /** Approval time (backend); the follow-up figures of an approved summary are frozen as of this date */
  approvedAt?: string

  // Section I
  narrative: string

  // Section II: five-level risk breakdown (Total Temuan = sum of the five)
  jumlahLaporan: number
  risikoHigh: number
  risikoModerateHigh: number
  risikoModerate: number
  risikoLowModerate: number
  risikoLow: number
  jumlahRekomendasi: number

  // Section III (JSON string stored in DB, array in store)
  followUpTable: FollowUpRow[]

  // Section IV (JSON string stored in DB, array in store)
  topFindings: SignificantFinding[]

  // Section VIII (JSON string stored in DB, array in store)
  matriksKompilasi: MatriksRow[]

  // Section V & VII
  akarMasalah: string
  kesimpulan: string

  // Signatures
  signatureTempat: string
  signatureTanggal: string
  signatureNamaKepala: string
  signatureNIK: string

  created_at?: string
  updated_at?: string
}

export const ES_PERSISTENCE_KEY = 'risk_based_audit_executive_summaries_persisted_v2'
export const ES_DELETED_DOCS_KEY = 'risk_based_audit_executive_summaries_deleted_docs_v1'

const loadDeletedDocNumbers = (): string[] => {
  if (typeof window === 'undefined') return []
  try {
    const raw = localStorage.getItem(ES_DELETED_DOCS_KEY)
    const parsed = raw ? JSON.parse(raw) : []
    return Array.isArray(parsed) ? parsed.filter((v): v is string => typeof v === 'string' && v.trim() !== '') : []
  } catch {
    return []
  }
}

const saveDeletedDocNumbers = (list: readonly string[]) => {
  if (typeof window === 'undefined') return
  try {
    localStorage.setItem(ES_DELETED_DOCS_KEY, JSON.stringify(Array.from(new Set(list))))
  } catch {
    // ignore storage errors
  }
}

export const loadPersistedExecutiveSummaryOverrides = (): Record<string, Partial<ExecutiveSummary>> => {
  if (typeof window === 'undefined') return {}
  try {
    // Purge legacy contaminated v1 storage if present
    if (localStorage.getItem('risk_based_audit_executive_summaries_persisted_v1')) {
      localStorage.removeItem('risk_based_audit_executive_summaries_persisted_v1')
    }
    const raw = localStorage.getItem(ES_PERSISTENCE_KEY)
    const current = raw ? JSON.parse(raw) : {}
    // Delete dummy or non-standard legacy entries
    delete current['DOC-EXSUM-Q1-2026']
    delete current['020/LHA/01/KS IAD/2023']
    delete current['019/LHA/01/KS IAD/2025']
    delete current['LHA-020/SKAI/2023']
    delete current['LHA-019/SKAI/2025']
    return current
  } catch (e) {
    console.warn('Failed to read persisted executive summaries from localStorage:', e)
    return {}
  }
}

export const savePersistedExecutiveSummaryOverride = (keys: (string | undefined)[], data: Partial<ExecutiveSummary>) => {
  if (typeof window === 'undefined') return
  try {
    const current = loadPersistedExecutiveSummaryOverrides()
    // Extract only safe modifiable fields - NEVER overwrite nomorDokumen or ID across cards!
    const safeData: Partial<ExecutiveSummary> = {}
    if (data.status) safeData.status = data.status
    if (data.executiveNote !== undefined) safeData.executiveNote = data.executiveNote
    if (data.narrative) safeData.narrative = data.narrative
    if (data.jumlahRekomendasi !== undefined) safeData.jumlahRekomendasi = data.jumlahRekomendasi
    if (data.risikoHigh !== undefined) safeData.risikoHigh = data.risikoHigh
    if (data.risikoModerateHigh !== undefined) safeData.risikoModerateHigh = data.risikoModerateHigh
    if (data.risikoModerate !== undefined) safeData.risikoModerate = data.risikoModerate
    if (data.risikoLowModerate !== undefined) safeData.risikoLowModerate = data.risikoLowModerate
    if (data.risikoLow !== undefined) safeData.risikoLow = data.risikoLow
    if (data.topFindings) safeData.topFindings = data.topFindings
    if (data.matriksKompilasi) safeData.matriksKompilasi = data.matriksKompilasi
    if (data.akarMasalah) safeData.akarMasalah = data.akarMasalah
    if (data.kesimpulan) safeData.kesimpulan = data.kesimpulan
    if (data.dokumenPath) safeData.dokumenPath = data.dokumenPath
    if (data.assignmentLetterId) safeData.assignmentLetterId = data.assignmentLetterId

    keys.filter((k): k is string => !!k && typeof k === 'string').forEach(k => {
      if (k === 'DOC-EXSUM-Q1-2026') return
      current[k] = { ...(current[k] || {}), ...safeData }
    })
    localStorage.setItem(ES_PERSISTENCE_KEY, JSON.stringify(current))
  } catch (e) {
    console.warn('Failed to save persisted executive summary override:', e)
  }
}

export const isValidUuid = (val?: string): boolean => {
  if (!val) return false
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(val)
}

export const useExecutiveSummaryStore = defineStore('executive-summary', () => {
  const summaryList = ref<ExecutiveSummary[]>([])
  const currentSummary = ref<ExecutiveSummary | null>(null)

  const showModal = ref(false)
  const isEditing = ref(false)
  const isViewing = ref(false)
  const loading = ref(false)
  const errorMsg = ref('')
  const toast = useToastNotification()
  const { t } = useI18n()

  // Default narrative template, starting with the period ("Periode Januari 2026, SPI telah menerbitkan ...").
  // period is the full label (e.g. "Januari – Maret 2026", "Kuartal 4 2026"); an optional year is appended to it.
  // Without a period the narrative starts with the synchronized body derived from the active form and linked LHA.
  const resolveFormNarrativeSource = (extra?: Partial<NarrativeSourceInput>): NarrativeSourceInput => {
    const isCompilation = extra?.isCompilation ?? isCompilationDocNumber(form.nomorDokumen)
    let linkedLha = extra?.linkedLha
    if (!isCompilation && !linkedLha && form.nomorDokumen) {
      try {
        const auditReportStore = useAuditResultReportStore()
        linkedLha = auditReportStore.reportList.find(
          r => (r.reportNumber || (r as any).report_number) === form.nomorDokumen
        ) as any
      } catch {
        // ignore when report store is unavailable
      }
    }
    return {
      isCompilation,
      nomorDokumen: form.nomorDokumen,
      jumlahLaporan: form.jumlahLaporan,
      risikoHigh: form.risikoHigh,
      risikoModerateHigh: form.risikoModerateHigh,
      risikoModerate: form.risikoModerate,
      risikoLowModerate: form.risikoLowModerate,
      risikoLow: form.risikoLow,
      jumlahRekomendasi: form.jumlahRekomendasi,
      followUpTable: form.followUpTable,
      topFindings: form.topFindings,
      matriksKompilasi: form.matriksKompilasi,
      akarMasalah: form.akarMasalah,
      linkedLha,
      formatDepartmentName,
      ...extra
    }
  }

  const defaultNarrativeTemplate = (period: string = '', tahun?: number | string, extra?: Partial<NarrativeSourceInput>) => {
    const fullPeriod = [period, tahun].filter(Boolean).join(' ')
    const syncData = deriveNarrativeSyncData(resolveFormNarrativeSource(extra))
    return buildSyncedNarrative(t, fullPeriod, syncData)
  }

  const syncFormNarrative = (period: string = '', extra?: Partial<NarrativeSourceInput>, force: boolean = false) => {
    if (!force && !isAutoSyncableNarrative(t, form.narrative)) return
    form.narrative = defaultNarrativeTemplate(period, undefined, extra)
  }

  // Active form data
  const form = reactive<Omit<ExecutiveSummary, 'id'>>({
    assignmentLetterId: '',
    quarter: 1,
    periodeBulan: 'Januari',
    tahun: 2026,
    nomorDokumen: '',
    dokumenPath: '',
    status: 'Draft',
    executiveNote: '',
    approvedAt: '',
    narrative: '',
    jumlahLaporan: 0,
    risikoHigh: 0,
    risikoModerateHigh: 0,
    risikoModerate: 0,
    risikoLowModerate: 0,
    risikoLow: 0,
    jumlahRekomendasi: 0,
    followUpTable: defaultFollowUpTable(),
    topFindings: [],
    matriksKompilasi: [],
    akarMasalah: '',
    kesimpulan: '',
    signatureTempat: '',
    signatureTanggal: '',
    signatureNamaKepala: '',
    signatureNIK: ''
  })

  // Master Data Departemen (Department Management Sync for Unit Division)
  const departments = ref<Department[]>([])
  const departmentsLoading = ref(false)
  const departmentsLoaded = ref(false)
  const departmentsError = ref('')

  const resolveDepartmentOption = (val?: string): string => {
    if (!val) return ''
    const match = departments.value.find(
      d => d.id === val || d.department_code === val || d.department_name === val || `${d.department_name} (${d.department_code})` === val
    )
    if (match) {
      return match.department_code ? `${match.department_name} (${match.department_code})` : match.department_name
    }
    return val
  }

  const formatDepartmentName = (value?: string): string => {
    if (!value) return '-'
    const found = departments.value.find(
      d => d.id === value || d.department_code === value || d.department_name === value || `${d.department_name} (${d.department_code})` === value
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

      // Resolve any existing topFindings unitDivision
      if (form.topFindings && form.topFindings.length > 0) {
        form.topFindings.forEach(f => {
          if (f.unitDivision) {
            f.unitDivision = resolveDepartmentOption(f.unitDivision)
          }
        })
      }
    } catch (error: any) {
      console.error('Failed to fetch departments for executive summary:', error)
      departmentsError.value = extractErrorMessage(error, t('toasts.common.loadDepartmentsFailed'))
      if (!departmentsLoaded.value) {
        departments.value = []
      }
    } finally {
      departmentsLoading.value = false
    }
  }

  const unitDivisionOptions = computed(() => {
    const list = departments.value.map(d => d.department_code ? `${d.department_name} (${d.department_code})` : d.department_name)
    if (form.topFindings && form.topFindings.length > 0) {
      form.topFindings.forEach(f => {
        if (f.unitDivision && !list.includes(f.unitDivision)) {
          list.push(f.unitDivision)
        }
      })
    }
    return Array.from(new Set(list)).sort((a, b) => a.localeCompare(b, 'id', { sensitivity: 'base' }))
  })

  const divisionOptions = unitDivisionOptions



  const fetchSummaries = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const extractItems = (res: any): any[] => {
        if (res && res.data && Array.isArray(res.data.items)) return res.data.items
        if (res && Array.isArray(res.items)) return res.items
        if (Array.isArray(res)) return res
        return []
      }

      const firstPage: any = await $fetch(`${baseUrl}/executive-summaries`, {
        method: 'GET',
        params: { page: 1, page_size: 100 }
      })
      const items: any[] = [...extractItems(firstPage)]
      const totalPages = Number(firstPage?.data?.pagination?.total_pages || firstPage?.pagination?.total_pages || 1)
      if (totalPages > 1) {
        for (let page = 2; page <= totalPages; page++) {
          const nextRes: any = await $fetch(`${baseUrl}/executive-summaries`, {
            method: 'GET',
            params: { page, page_size: 100 }
          })
          items.push(...extractItems(nextRes))
        }
      }

      const isExcluded = (doc?: string) =>
        !doc ||
        doc === 'DOC-EXSUM-Q1-2026' ||
        doc === '020/LHA/01/KS IAD/2023' ||
        doc === '019/LHA/01/KS IAD/2025' ||
        doc === 'LHA-020/SKAI/2023' ||
        doc === 'LHA-019/SKAI/2025'

      summaryList.value = items
        .map(item => parseSummaryFromBackend(item))
        .filter(b => !isExcluded(b.nomorDokumen))
    } catch (error) {
      console.error('Failed to fetch executive summaries:', error)
      errorMsg.value = extractErrorMessage(error, t('toasts.executiveSummary.loadFailed'))
      summaryList.value = []
    } finally {
      // Overlay persisted changes from localStorage (status approvals and notes)
      const overrides = loadPersistedExecutiveSummaryOverrides()
      const isExcluded = (doc?: string) =>
        !doc ||
        doc === 'DOC-EXSUM-Q1-2026' ||
        doc === '020/LHA/01/KS IAD/2023' ||
        doc === '019/LHA/01/KS IAD/2025' ||
        doc === 'LHA-020/SKAI/2023' ||
        doc === 'LHA-019/SKAI/2025'

      summaryList.value = summaryList.value
        .map(item => {
          const override = overrides[item.nomorDokumen] || overrides[item.id]
          return override ? { ...item, ...override } : item
        })
        .filter(s => !isExcluded(s.nomorDokumen))
      loading.value = false
    }
  }

  const parseSummaryFromBackend = (item: any): ExecutiveSummary => {
    const parsed: ExecutiveSummary = {
      ...item,
      assignmentLetterId: item.assignmentLetterId || item.assignment_letter_id || '',
      executiveNote: item.executiveNote || item.executive_note || '',
      approvedAt: item.approvedAt || item.approved_at || '',
      quarter: Number(item.quarter),
      tahun: Number(item.tahun),
      jumlahLaporan: Number(item.jumlahLaporan),
      // Five levels. Legacy-only rows are copied into these columns by the backend on startup.
      risikoHigh: Number(item.risikoHigh) || 0,
      risikoModerateHigh: Number(item.risikoModerateHigh) || 0,
      risikoModerate: Number(item.risikoModerate) || 0,
      risikoLowModerate: Number(item.risikoLowModerate) || 0,
      risikoLow: Number(item.risikoLow) || 0,
      jumlahRekomendasi: Number(item.jumlahRekomendasi),
      followUpTable: item.followUpTable ? (typeof item.followUpTable === 'string' ? JSON.parse(item.followUpTable) : item.followUpTable) : [],
      topFindings: item.topFindings ? (typeof item.topFindings === 'string' ? JSON.parse(item.topFindings) : item.topFindings) : [],
      matriksKompilasi: item.matriksKompilasi ? (typeof item.matriksKompilasi === 'string' ? JSON.parse(item.matriksKompilasi) : item.matriksKompilasi) : []
    }
    if (!parsed.narrative || parsed.narrative.includes('....') || parsed.narrative.includes('….')) {
      const isComp = isCompilationDocNumber(parsed.nomorDokumen)
      const period = isComp
        ? formatQuarterPeriod(t, parsed.quarter || 1, parsed.tahun || 2026)
        : [parsed.periodeBulan, parsed.periodeBulan?.includes(String(parsed.tahun)) ? '' : parsed.tahun].filter(Boolean).join(' ')
      parsed.narrative = buildSyncedNarrative(t, period, deriveNarrativeSyncData({
        isCompilation: isComp,
        ...parsed,
        formatDepartmentName
      }))
    }
    return parsed
  }

  const serializeSummaryForBackend = (data: Omit<ExecutiveSummary, 'id'> | ExecutiveSummary) => {
    const rawId = (data as any).id
    const payload: any = {
      ...data,
      assignmentLetterId: (data as any).assignmentLetterId || '',
      quarter: Number(data.quarter) || 1,
      tahun: Number(data.tahun) || 2026,
      jumlahLaporan: Number(data.jumlahLaporan) || 0,
      risikoHigh: Number(data.risikoHigh) || 0,
      risikoModerateHigh: Number(data.risikoModerateHigh) || 0,
      risikoModerate: Number(data.risikoModerate) || 0,
      risikoLowModerate: Number(data.risikoLowModerate) || 0,
      risikoLow: Number(data.risikoLow) || 0,
      jumlahRekomendasi: Number(data.jumlahRekomendasi) || 0,
      followUpTable: typeof data.followUpTable === 'string' ? data.followUpTable : JSON.stringify(data.followUpTable || []),
      topFindings: typeof data.topFindings === 'string' ? data.topFindings : JSON.stringify(data.topFindings || []),
      matriksKompilasi: typeof data.matriksKompilasi === 'string' ? data.matriksKompilasi : JSON.stringify(data.matriksKompilasi || [])
    }
    if (!isValidUuid(rawId)) {
      delete payload.id
    }
    // Strip server-managed timestamp and approval fields: sending "" for approvedAt/created_at/updated_at
    // breaks Go's RFC3339 time.Time JSON unmarshaler ("parsing time \"\" as \"2006-01-02T15:04:05Z07:00\"").
    delete payload.approvedAt
    delete payload.approved_at
    delete payload.created_at
    delete payload.updated_at
    delete payload.deleted_at
    delete payload.signatureImage
    delete payload.signatureName
    delete payload.signatureRole
    delete payload.approvedBy
    delete payload.approvedByUserId
    delete payload.approvedByRole
    return payload
  }

  const generateEscDocNumber = (year: number = 2026): string => {
    let maxSeq = 0
    const escRegex = new RegExp(`^ESC-(\\d+)/${year}$`, 'i')

    for (const s of summaryList.value) {
      if (s.nomorDokumen) {
        const match = s.nomorDokumen.match(escRegex)
        if (match && match[1]) {
          const val = parseInt(match[1], 10)
          if (!isNaN(val) && val > maxSeq) {
            maxSeq = val
          }
        }
      }
    }

    try {
      const overrides = loadPersistedExecutiveSummaryOverrides()
      for (const key of Object.keys(overrides)) {
        const matchKey = key.match(escRegex)
        if (matchKey && matchKey[1]) {
          const val = parseInt(matchKey[1], 10)
          if (!isNaN(val) && val > maxSeq) {
            maxSeq = val
          }
        }
        const item = overrides[key]
        if (item && item.nomorDokumen) {
          const matchDoc = item.nomorDokumen.match(escRegex)
          if (matchDoc && matchDoc[1]) {
            const val = parseInt(matchDoc[1], 10)
            if (!isNaN(val) && val > maxSeq) {
              maxSeq = val
            }
          }
        }
      }
    } catch (e) {
      // ignore localstorage errors
    }

    const nextSeq = (maxSeq + 1).toString().padStart(3, '0')
    return `ESC-${nextSeq}/${year}`
  }

  const openNewForm = (quarterNum: number = 1, assignmentLetterId: string = '', isCompilation: boolean = false) => {
    isEditing.value = false
    isViewing.value = false
    currentSummary.value = null

    // Set default month based on selected quarter
    let defaultMonth = 'Januari'
    if (quarterNum === 2) defaultMonth = 'April'
    if (quarterNum === 3) defaultMonth = 'Juli'
    if (quarterNum === 4) defaultMonth = 'Oktober'

    const targetYear = 2026
    const autoDocNumber = isCompilation ? generateEscDocNumber(targetYear) : ''

    Object.assign(form, {
      assignmentLetterId: assignmentLetterId || 'ST-001/SKAI/2026',
      quarter: quarterNum || 1,
      periodeBulan: defaultMonth,
      tahun: targetYear,
      nomorDokumen: autoDocNumber,
      dokumenPath: '',
      status: 'Draft',
      executiveNote: '',
      approvedAt: '',
      // Compilation: the selected quarter. Individual: no period until an LHA is selected (the form adds the audit period).
      narrative: isCompilation ? defaultNarrativeTemplate(formatQuarterPeriod(t, quarterNum || 1, targetYear)) : defaultNarrativeTemplate(),
      jumlahLaporan: 0,
      risikoHigh: 0,
      risikoModerateHigh: 0,
      risikoModerate: 0,
      risikoLowModerate: 0,
      risikoLow: 0,
      jumlahRekomendasi: 0,
      followUpTable: defaultFollowUpTable(),
      topFindings: [],
      matriksKompilasi: [],
      akarMasalah: '',
      kesimpulan: '',
      signatureTempat: 'Jakarta',
      signatureTanggal: new Date().toISOString().split('T')[0],
      signatureNamaKepala: '',
      signatureNIK: ''
    })
    if (!departmentsLoaded.value && !departmentsLoading.value) {
      fetchDepartments().catch(() => {})
    }
    showModal.value = true
  }

  const openNewCompilationForm = (quarterNum: number = 1, targetYear: number = 2026) => {
    openNewForm(quarterNum, '', true)
    form.tahun = targetYear
    form.nomorDokumen = generateEscDocNumber(targetYear)
    form.narrative = defaultNarrativeTemplate(formatQuarterPeriod(t, form.quarter, targetYear))
  }

  const openEditForm = (summary: ExecutiveSummary) => {
    isEditing.value = true
    isViewing.value = false
    currentSummary.value = summary
    Object.assign(form, JSON.parse(JSON.stringify(summary)))
    if (!form.assignmentLetterId && summary.assignmentLetterId) {
      form.assignmentLetterId = summary.assignmentLetterId
    }
    if (form.topFindings && form.topFindings.length > 0) {
      form.topFindings.forEach(f => {
        if (f.unitDivision) {
          f.unitDivision = resolveDepartmentOption(f.unitDivision)
        }
        if (f.risiko === 'Tinggi') f.risiko = 'Significant'
        else if (f.risiko === 'Sedang') f.risiko = 'Quite Significant'
        else if (f.risiko === 'Rendah') f.risiko = 'Not Significant'
      })
    }
    if (!form.narrative || form.narrative.includes('....') || form.narrative.includes('….')) {
      const isComp = isCompilationDocNumber(form.nomorDokumen)
      const period = isComp
        ? formatQuarterPeriod(t, form.quarter || 1, form.tahun || 2026)
        : [form.periodeBulan, form.periodeBulan?.includes(String(form.tahun)) ? '' : form.tahun].filter(Boolean).join(' ')
      syncFormNarrative(period, { isCompilation: isComp }, true)
    }
    if (!departmentsLoaded.value && !departmentsLoading.value) {
      fetchDepartments().catch(() => {})
    }
    showModal.value = true
  }

  const openView = (summary: ExecutiveSummary) => {
    isEditing.value = false
    isViewing.value = true
    currentSummary.value = summary
    Object.assign(form, JSON.parse(JSON.stringify(summary)))
    if (!form.assignmentLetterId && summary.assignmentLetterId) {
      form.assignmentLetterId = summary.assignmentLetterId
    }
    if (form.topFindings && form.topFindings.length > 0) {
      form.topFindings.forEach(f => {
        if (f.unitDivision) {
          f.unitDivision = resolveDepartmentOption(f.unitDivision)
        }
        if (f.risiko === 'Tinggi') f.risiko = 'Significant'
        else if (f.risiko === 'Sedang') f.risiko = 'Quite Significant'
        else if (f.risiko === 'Rendah') f.risiko = 'Not Significant'
      })
    }
    if (!form.narrative || form.narrative.includes('....') || form.narrative.includes('….')) {
      const isComp = isCompilationDocNumber(form.nomorDokumen)
      const period = isComp
        ? formatQuarterPeriod(t, form.quarter || 1, form.tahun || 2026)
        : [form.periodeBulan, form.periodeBulan?.includes(String(form.tahun)) ? '' : form.tahun].filter(Boolean).join(' ')
      syncFormNarrative(period, { isCompilation: isComp }, true)
    }
    if (!departmentsLoaded.value && !departmentsLoading.value) {
      fetchDepartments().catch(() => {})
    }
    showModal.value = true
  }

  const saveForm = async (opts?: { requireLha?: boolean }): Promise<boolean> => {
    if (opts?.requireLha && (!form.nomorDokumen || isCompilationDocNumber(form.nomorDokumen))) {
      const msg = t('executiveSummaryForm.validation.lhaRequired')
      errorMsg.value = msg
      toast.showError(msg)
      return false
    }

    loading.value = true
    errorMsg.value = ''
    try {
      if (!form.nomorDokumen) {
        form.nomorDokumen = generateEscDocNumber(form.tahun || 2026)
      }
      if (form.status !== 'Approved') {
        form.followUpTable = withAllFollowUpRows(form.followUpTable)
      }
      const baseUrl = getAuditServiceBaseUrl()
      const payload = serializeSummaryForBackend(form)

      let savedId = ''
      if (isEditing.value && currentSummary.value) {
        const existingByDoc = summaryList.value.find(
          s => !!form.nomorDokumen && s.nomorDokumen === form.nomorDokumen && isValidUuid(s.id)
        )
        const targetId = isValidUuid(currentSummary.value.id) ? currentSummary.value.id : existingByDoc?.id
        if (targetId) {
          const res: any = await $fetch(`${baseUrl}/executive-summaries/${targetId}`, {
            method: 'PUT',
            body: payload
          })
          savedId = res?.data?.id || targetId
          currentSummary.value = { ...currentSummary.value, ...JSON.parse(JSON.stringify(form)), id: savedId }
        } else {
          const res: any = await $fetch(`${baseUrl}/executive-summaries`, {
            method: 'POST',
            body: payload
          })
          if (res?.data?.id) {
            savedId = res.data.id
            currentSummary.value = parseSummaryFromBackend(res.data)
          }
        }
      } else {
        // Prevent duplicate creation: if summary with same nomorDokumen exists, update it!
        const existingItem = summaryList.value.find(s => form.nomorDokumen && s.nomorDokumen === form.nomorDokumen)
        if (existingItem && isValidUuid(existingItem.id)) {
          const res: any = await $fetch(`${baseUrl}/executive-summaries/${existingItem.id}`, {
            method: 'PUT',
            body: payload
          })
          savedId = res?.data?.id || existingItem.id
          currentSummary.value = { ...existingItem, ...JSON.parse(JSON.stringify(form)), id: savedId }
        } else {
          const res: any = await $fetch(`${baseUrl}/executive-summaries`, {
            method: 'POST',
            body: payload
          })
          if (res?.data?.id) {
            savedId = res.data.id
            currentSummary.value = parseSummaryFromBackend(res.data)
          }
        }
      }

      savePersistedExecutiveSummaryOverride(
        [form.nomorDokumen, savedId || currentSummary.value?.id],
        { ...form }
      )
      if (form.nomorDokumen && deletedDocNumbers.value.includes(form.nomorDokumen)) {
        deletedDocNumbers.value = deletedDocNumbers.value.filter(d => d !== form.nomorDokumen)
        saveDeletedDocNumbers(deletedDocNumbers.value)
      }

      showModal.value = false
      await fetchSummaries()

      // 2-Way Sync: Update matching AuditResultReport item in Result Reports store
      try {
        const auditReportStore = useAuditResultReportStore()
        if (form.nomorDokumen && form.narrative) {
          await auditReportStore.syncExecutiveSummaryField(form.nomorDokumen, form.narrative, form.jumlahRekomendasi)
        }
      } catch (errSync) {
        console.warn('Sync to AuditResultReport failed:', errSync)
      }

      toast.showSuccess(t('toasts.executiveSummary.saved'))
      return true
    } catch (error: any) {
      console.error('Failed to save summary to backend:', error)
      const detail = extractErrorMessage(error, t('toasts.executiveSummary.saveFailed'))
      errorMsg.value = detail
      toast.showError(t('toasts.executiveSummary.saveFailed'), detail)
      return false
    } finally {
      loading.value = false
    }
  }

  const deletedDocNumbers = ref<string[]>(loadDeletedDocNumbers())

  const deleteSummary = async (id: string, docNum?: string) => {
    if (!await useGlobalModalStore().confirmDelete({ description: 'Apakah Anda yakin ingin menghapus Laporan Eksekutif ini?' })) return
    loading.value = true
    const item = summaryList.value.find(s => s.id === id || s.nomorDokumen === docNum)
    const targetDocNum = docNum || (item ? item.nomorDokumen : '')

    if (targetDocNum && !deletedDocNumbers.value.includes(targetDocNum)) {
      deletedDocNumbers.value.push(targetDocNum)
      saveDeletedDocNumbers(deletedDocNumbers.value)
    }

    // DELETE /executive-summaries/:id only accepts the UUID of an executive_summaries row,
    // like GET/PUT. Cards that the Individual page derives from an LHA without a summary row
    // carry `ES-<auditResultReportId>`, and offline saves get `ES-Q<n>-...`; neither exists
    // in the backend. Resolve the real row (by id, else by document number) like the PUT
    // paths do, and only remove locally when there is none.
    const backendId = isValidUuid(id)
      ? id
      : summaryList.value.find(s => !!targetDocNum && s.nomorDokumen === targetDocNum && isValidUuid(s.id))?.id

    try {
      if (backendId) {
        const baseUrl = getAuditServiceBaseUrl()
        await $fetch(`${baseUrl}/executive-summaries/${backendId}`, { method: 'DELETE' })
        await fetchSummaries()
      }
      toast.showSuccess(t('toasts.executiveSummary.deleted'))
    } catch (error: any) {
      console.error('Failed to delete on backend, simulating local deletion:', error)
      const detail = extractErrorMessage(error, t('toasts.executiveSummary.deleteFailed'))
      errorMsg.value = detail
      toast.showError(t('toasts.executiveSummary.deleteFailed'), detail)
    } finally {
      summaryList.value = summaryList.value.filter(s => s.id !== id && s.nomorDokumen !== targetDocNum)
      if (targetDocNum) {
        try {
          const auditReportStore = useAuditResultReportStore()
          await auditReportStore.clearExecutiveSummaryField(targetDocNum)
        } catch (e) {
          console.warn('Failed to clear executive summary on report store:', e)
        }
      }
      loading.value = false
    }
  }

  const saveExecutiveNote = async () => {
    if (!currentSummary.value) return

    loading.value = true
    errorMsg.value = ''
    const note = form.executiveNote?.trim() || ''
    const updated: ExecutiveSummary = {
      ...currentSummary.value,
      ...JSON.parse(JSON.stringify(form)),
      executiveNote: note
    }

    // Persist immediately in localStorage so page refresh never loses the note
    savePersistedExecutiveSummaryOverride(
      [currentSummary.value.id, currentSummary.value.nomorDokumen, form.nomorDokumen],
      { executiveNote: note }
    )

    try {
      const baseUrl = getAuditServiceBaseUrl()
      if (isValidUuid(currentSummary.value.id)) {
        await $fetch(`${baseUrl}/executive-summaries/${currentSummary.value.id}`, {
          method: 'PUT',
          body: serializeSummaryForBackend(updated)
        })
      } else {
        const realItem = summaryList.value.find(s => (s.id === currentSummary.value!.id || s.nomorDokumen === updated.nomorDokumen) && isValidUuid(s.id))
        if (realItem) {
          await $fetch(`${baseUrl}/executive-summaries/${realItem.id}`, {
            method: 'PUT',
            body: serializeSummaryForBackend(updated)
          })
          updated.id = realItem.id
        } else {
          try {
            const res: any = await $fetch(`${baseUrl}/executive-summaries`, {
              method: 'POST',
              body: serializeSummaryForBackend(updated)
            })
            if (res && res.data && res.data.id) {
              updated.id = res.data.id
            }
          } catch (postErr) {
            console.warn('POST executive summary note failed:', postErr)
          }
        }
      }
    } catch (error: any) {
      console.warn('Failed to save executive note on backend, saving to local store and storage:', error)
    } finally {
      const index = summaryList.value.findIndex(summary => 
        summary.id === currentSummary.value?.id || 
        (summary.nomorDokumen && summary.nomorDokumen === currentSummary.value?.nomorDokumen)
      )
      if (index !== -1) {
        summaryList.value[index] = { ...summaryList.value[index], ...updated }
      } else {
        summaryList.value.push(updated)
      }
      summaryList.value = [...summaryList.value]
      currentSummary.value = updated
      form.executiveNote = note

      // Sync note to matching report in AuditResultReport store for Auditor view
      try {
        const auditReportStore = useAuditResultReportStore()
        const targetDocNum = updated.nomorDokumen
        const matchedReport = auditReportStore.reportList.find(r => r.reportNumber === targetDocNum)
        if (matchedReport) {
          (matchedReport as any).executiveNote = note
        }
      } catch (e) {
        console.warn('Sync note to AuditResultReport failed:', e)
      }

      toast.showSuccess(t('toasts.executiveSummary.noteSaved'))
      loading.value = false
    }
  }

  // Ensures the current Draft summary has a real UUID row on the server before approving.
  // Needed for virtual Individual cards derived from an LHA (id = "ES-<lhaId>") that have
  // not yet been persisted to the executive_summaries table.
  const ensureSavedOnServer = async (): Promise<boolean> => {
    const target = currentSummary.value
    if (!target) return false
    if (isValidUuid(target.id)) return true

    const existing = summaryList.value.find(
      s => !!target.nomorDokumen && s.nomorDokumen === target.nomorDokumen && isValidUuid(s.id)
    )
    if (existing) {
      currentSummary.value = { ...target, id: existing.id }
      return true
    }

    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const draftData: Omit<ExecutiveSummary, 'id'> = {
        ...JSON.parse(JSON.stringify(form)),
        status: 'Draft',
        followUpTable: withAllFollowUpRows(form.followUpTable)
      }
      const res: any = await $fetch(`${baseUrl}/executive-summaries`, {
        method: 'POST',
        body: serializeSummaryForBackend(draftData)
      })
      const createdId = res?.data?.id
      if (!isValidUuid(createdId)) {
        toast.showError(t('executiveSummaryForm.approval.errors.saveFirst'))
        return false
      }
      const created: ExecutiveSummary = res?.data ? parseSummaryFromBackend(res.data) : { ...target, ...draftData, id: createdId }
      currentSummary.value = created
      await fetchSummaries()
      return true
    } catch (error: any) {
      const fallback = t('toasts.executiveSummary.saveFailed')
      const detail = extractErrorMessage(error, fallback)
      errorMsg.value = detail
      toast.showError(fallback, detail)
      return false
    } finally {
      loading.value = false
    }
  }

  // Approval with a signature: the only way to move a Draft to Approved. The backend refuses a PUT that sets
  // status Approved (APPROVE_REQUIRES_SIGNATURE), so the status change happens in POST /:id/approve instead.
  const approveWithSignature = async (payload: ExecutiveSummaryApprovePayload): Promise<boolean> => {
    const target = currentSummary.value
    if (!target) return false
    const resolvedId = isValidUuid(target.id)
      ? target.id
      : summaryList.value.find(s => !!target.nomorDokumen && s.nomorDokumen === target.nomorDokumen && isValidUuid(s.id))?.id
    if (!resolvedId) {
      toast.showError(t('executiveSummaryForm.approval.errors.saveFirst'))
      return false
    }

    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/executive-summaries/${resolvedId}/approve`, {
        method: 'POST',
        body: payload
      })
    } catch (error: any) {
      const fallback = t('executiveSummaryForm.approval.errors.failed')
      const detail = extractErrorMessage(error, fallback)
      errorMsg.value = detail
      toast.showError(fallback, detail)
      return false
    } finally {
      loading.value = false
    }

    // The backend accepted the approval: mirror it locally. Nothing is written on failure, so a Draft is never shown as Approved.
    const approvedFields: Partial<ExecutiveSummary> = {
      status: 'Approved',
      approvedAt: new Date().toISOString(),
      signatureTempat: payload.signaturePlace,
      signatureTanggal: payload.signatureDate,
      signatureNamaKepala: payload.signerName
    }
    const idx = summaryList.value.findIndex(s => s.id === resolvedId)
    if (idx !== -1) {
      summaryList.value[idx] = { ...summaryList.value[idx]!, ...approvedFields }
    }
    summaryList.value = [...summaryList.value]
    currentSummary.value = { ...target, id: resolvedId, ...approvedFields }
    form.status = 'Approved'
    form.approvedAt = approvedFields.approvedAt
    form.signatureTempat = payload.signaturePlace
    form.signatureTanggal = payload.signatureDate
    form.signatureNamaKepala = payload.signerName
    savePersistedExecutiveSummaryOverride([target.nomorDokumen, resolvedId], { status: 'Approved' })

    // The backend freezes the follow-up figures on approval; show its snapshot instead of the local draft values
    try {
      await fetchSummaries()
      const frozen = summaryList.value.find(s => s.id === resolvedId)
      if (frozen && currentSummary.value?.id === resolvedId) {
        if (frozen.approvedAt) form.approvedAt = frozen.approvedAt
        if (frozen.followUpTable?.length) form.followUpTable = JSON.parse(JSON.stringify(frozen.followUpTable))
        if (frozen.topFindings?.length) form.topFindings = JSON.parse(JSON.stringify(frozen.topFindings))
        if (frozen.matriksKompilasi?.length) form.matriksKompilasi = JSON.parse(JSON.stringify(frozen.matriksKompilasi))
        currentSummary.value = { ...currentSummary.value, ...frozen }
      }
    } catch (e) {
      console.warn('Reloading the approved executive summary failed:', e)
    }

    try {
      const auditReportStore = useAuditResultReportStore()
      const matchedReport = auditReportStore.reportList.find(r => r.reportNumber === target.nomorDokumen)
      if (matchedReport) {
        matchedReport.status = 'Final'
      }
    } catch (e) {
      console.warn('Sync status to AuditResultReport failed:', e)
    }

    toast.showSuccess(t('executiveSummaryForm.approval.success'))
    return true
  }

  const updateStatus = async (id: string, newStatus: 'Draft' | 'Approved' | 'Rejected') => {
    if (newStatus === 'Approved') {
      // Approval must carry a signature; see approveWithSignature. A plain status change is refused here.
      toast.showError(t('executiveSummaryForm.approval.errors.signatureRequired'))
      return
    }
    loading.value = true
    errorMsg.value = ''
    const item = summaryList.value.find(s => s.id === id)
      || (currentSummary.value?.id === id ? currentSummary.value : null)
    const updated: ExecutiveSummary = item
      ? { ...item, status: newStatus }
      : { ...(currentSummary.value || form), id, status: newStatus } as ExecutiveSummary

    // Persist immediately in localStorage so page refresh never loses the approved status
    savePersistedExecutiveSummaryOverride(
      [id, updated.nomorDokumen, item?.nomorDokumen, form.nomorDokumen],
      { status: newStatus }
    )

    try {
      const baseUrl = getAuditServiceBaseUrl()
      if (isValidUuid(id)) {
        const payload = serializeSummaryForBackend(updated)
        await $fetch(`${baseUrl}/executive-summaries/${id}`, {
          method: 'PUT',
          body: payload
        })
      } else {
        const realItem = summaryList.value.find(s => (s.id === id || s.nomorDokumen === updated.nomorDokumen) && isValidUuid(s.id))
        if (realItem) {
          await $fetch(`${baseUrl}/executive-summaries/${realItem.id}`, {
            method: 'PUT',
            body: serializeSummaryForBackend(updated)
          })
          updated.id = realItem.id
        } else {
          try {
            const res: any = await $fetch(`${baseUrl}/executive-summaries`, {
              method: 'POST',
              body: serializeSummaryForBackend(updated)
            })
            if (res && res.data && res.data.id) {
              updated.id = res.data.id
            }
          } catch (postErr) {
            console.warn('POST executive summary status failed:', postErr)
          }
        }
      }
    } catch (error: any) {
      console.warn('Failed to update status on backend, simulating local update:', error)
    } finally {
      const idx = summaryList.value.findIndex(s => 
        s.id === id || 
        (updated.nomorDokumen && s.nomorDokumen === updated.nomorDokumen)
      )
      if (idx !== -1) {
        summaryList.value[idx] = { ...summaryList.value[idx], ...updated }
      } else {
        summaryList.value.push(updated)
      }
      summaryList.value = [...summaryList.value]

      if (currentSummary.value) {
        currentSummary.value = { ...currentSummary.value, ...updated }
      }
      form.status = newStatus

      // Sync status to matching report in AuditResultReport store
      try {
        const auditReportStore = useAuditResultReportStore()
        const targetDocNum = updated.nomorDokumen
        const matchedReport = auditReportStore.reportList.find(r => r.reportNumber === targetDocNum)
        if (matchedReport) {
          // Approved never reaches here (see the guard above); Draft and Rejected map to a Draft report
          matchedReport.status = 'Draft'
        }
      } catch (e) {
        console.warn('Sync status to AuditResultReport failed:', e)
      }

      toast.showSuccess(t('toasts.executiveSummary.statusUpdated'))
      loading.value = false
    }
  }

  /**
   * Live follow-up status from the Action Taken Reports:
   * GET /executive-summaries/follow-up-status?audit_result_report_id= | ?report_number= | ?year=&quarter=
   * Throws on a failed request or an unusable response; callers show the error.
   */
  const fetchFollowUpStatus = async (query: FollowUpStatusQuery): Promise<FollowUpStatusResult> => {
    const baseUrl = getAuditServiceBaseUrl()
    const response = await $fetch<unknown>(`${baseUrl}/executive-summaries/follow-up-status`, {
      method: 'GET',
      params: followUpStatusParams(query)
    })
    return parseFollowUpStatus(response)
  }

  /**
   * Generates and downloads a Microsoft Word (.docx) document for an Approved Executive Summary:
   * - Compilation: "Executive Summary Compilation Hasil Audit Triwulanan" template (Sections I–VII + Lampiran Matriks Induk)
   * - Individual: Executive Summary Individual template adapted to the individual LHA data
   * Refuses generation when the summary status is not 'Approved' (e.g. still 'Draft').
   */
  const downloadDocx = async (
    targetSummary?: ExecutiveSummary | Omit<ExecutiveSummary, 'id'> | null,
    context: ExecutiveSummaryDocxContext = {}
  ): Promise<boolean> => {
    const rawTarget = targetSummary || currentSummary.value || form
    if (!canGenerateExecutiveSummaryDocx(rawTarget)) {
      const msg = t('executiveSummaryForm.docx.onlyApproved')
      errorMsg.value = msg
      toast.showError(msg)
      return false
    }

    const summary: ExecutiveSummary = {
      id: (rawTarget as ExecutiveSummary).id || currentSummary.value?.id || '',
      ...JSON.parse(JSON.stringify(rawTarget))
    }
    const variant: ExecutiveSummaryDocxVariant = resolveExecutiveSummaryVariant(summary, context.variant)
    const resolvedContext: ExecutiveSummaryDocxContext = {
      ...context,
      variant,
      formatDepartmentName
    }

    let linkedLha: any = null
    try {
      const auditReportStore = useAuditResultReportStore()
      if (summary.nomorDokumen) {
        linkedLha = auditReportStore.reportList.find(
          r => (r.reportNumber || (r as any).report_number) === summary.nomorDokumen
        )
      }
    } catch {
      // ignore if store not initialized
    }

    if (linkedLha) {
      if (!resolvedContext.reportTitle) resolvedContext.reportTitle = linkedLha.reportTitle
      if (!resolvedContext.auditObject) resolvedContext.auditObject = (linkedLha as any).auditObject || linkedLha.reportTitle
      if (!resolvedContext.department) resolvedContext.department = linkedLha.department
      if (!summary.assignmentLetterId && linkedLha.assignmentLetterId) {
        summary.assignmentLetterId = linkedLha.assignmentLetterId
      }
    }

    if (!summary.narrative || summary.narrative.includes('....') || summary.narrative.includes('….')) {
      const period = variant === 'compilation'
        ? formatQuarterPeriod(t, summary.quarter || 1, summary.tahun || 2026)
        : [summary.periodeBulan, summary.periodeBulan?.includes(String(summary.tahun)) ? '' : summary.tahun].filter(Boolean).join(' ')
      summary.narrative = buildSyncedNarrative(t, period, deriveNarrativeSyncData({
        isCompilation: variant === 'compilation',
        ...summary,
        linkedLha,
        formatDepartmentName
      }))
    }

    const filename = executiveSummaryDocxFilename(summary, variant)
    let blob: Blob | null = null

    const backendId = isValidUuid(summary.id)
      ? summary.id
      : summaryList.value.find(s => !!summary.nomorDokumen && s.nomorDokumen === summary.nomorDokumen && isValidUuid(s.id))?.id

    if (backendId) {
      try {
        const baseUrl = getAuditServiceBaseUrl()
        const fetched = await $fetch<Blob>(`${baseUrl}/executive-summaries/${backendId}/download-docx`, {
          method: 'GET',
          params: { variant },
          responseType: 'blob'
        })
        if (fetched instanceof Blob && fetched.size > 0) {
          blob = fetched
        }
      } catch {
        // Fallback to local docx builder when offline or when approval is persisted in localStorage
      }
    }

    try {
      if (!blob) {
        const bytes = generateExecutiveSummaryDocxBytes(summary, resolvedContext)
        blob = new Blob([bytes], {
          type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
        })
      }
      if (typeof window !== 'undefined' && typeof document !== 'undefined' && window.URL?.createObjectURL) {
        const url = window.URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = filename
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        window.URL.revokeObjectURL(url)
      }
      return true
    } catch (err: any) {
      console.error('Failed to generate/download Executive Summary docx:', err)
      const detail = extractErrorMessage(err, t('executiveSummaryForm.docx.downloadFailed'))
      errorMsg.value = detail
      toast.showError(t('executiveSummaryForm.docx.downloadFailed'), detail)
      return false
    }
  }

  // Load initially
  fetchSummaries()

  return {
    summaryList,
    deletedDocNumbers,
    currentSummary,
    showModal,
    isEditing,
    isViewing,
    loading,
    errorMsg,
    form,
    openNewForm,
    openNewCompilationForm,
    openEditForm,
    openView,
    saveForm,
    saveExecutiveNote,
    deleteSummary,
    updateStatus,
    ensureSavedOnServer,
    approveWithSignature,
    fetchFollowUpStatus,
    fetchSummaries,
    defaultNarrativeTemplate,
    syncFormNarrative,
    downloadDocx,
    generateEscDocNumber,
    departments,
    departmentsLoading,
    departmentsLoaded,
    departmentsError,
    unitDivisionOptions,
    divisionOptions,
    fetchDepartments,
    resolveDepartmentOption,
    formatDepartmentName
  }
})
