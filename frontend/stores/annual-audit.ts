// stores/annual-plan.ts
import type { TableColumn } from '@nuxt/ui'
import { defineStore } from 'pinia'
import { ref, reactive, computed, watch } from 'vue'
import { AnnualAuditPlanStatus, AuditCategory, type AnnualAuditAttachment, type AnnualAuditPlan, type AnnualPlanForm, type AuditActivities } from '~/types/audit'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { useI18n } from '~/composables/useI18n'
import type { TablePagination } from '~/types/common'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { extractErrorMessage } from '~/utils/error'
import { getFiscalYearStrings } from '~/composables/useFiscalYear'
import { useDepartmentApi } from '~/composables/useDepartmentApi'
import { buildAnnualPlanPayload, sumActivityMandays, toCount } from '~/utils/annualPlanPayload'
import { parseApiError, translateApiError } from '~/utils/activityPlanForm'

/** Outcome of deletePlan: deleted, deactivated (the plan is in use), or null when nothing was done. */
export type DeletePlanResult = 'deleted' | 'deactivated' | null

/** 409 codes of /annual-audit-plans tied to the Audit Activity Plan; translated at annualAudit.errors.<CODE>. */
const ANNUAL_PLAN_ERROR_CODES = new Set(['ANNUAL_PLAN_IN_USE', 'ANNUAL_PLAN_ACTIVITY_IN_USE', 'ANNUAL_PLAN_ACTIVITY_TYPE_CONFLICT'])

export const useAnnualPlanStore = defineStore('annual-audit', () => {
  const showModal = ref(false)
  const errorMsg = ref('')
  const loading = ref(false)
  const validationErrors = reactive({
    activityDetail: '',
    timeline: '',
    auditor: ''
  })
  const toast = useToastNotification()
  const { t } = useI18n()

  // Canonical status vocabulary (API values) -> i18n key under annualAudit.status.*
  const statusLabelKeys: Record<AnnualAuditPlanStatus, string> = {
    [AnnualAuditPlanStatus.DRAFT]: 'annualAudit.status.draft',
    [AnnualAuditPlanStatus.PENDING_REVIEW]: 'annualAudit.status.pendingReview',
    [AnnualAuditPlanStatus.PENDING_APPROVAL]: 'annualAudit.status.pendingApproval',
    [AnnualAuditPlanStatus.APPROVED]: 'annualAudit.status.approved',
    [AnnualAuditPlanStatus.REVISED]: 'annualAudit.status.revised'
  }

  // Constants
  const monthsList = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  const attachmentCategoryOptions = ['Plan', 'Evidence', 'Charter', 'Other']
  const paginations = ref<TablePagination>({
    pageSize: 10,
    currentIndex: 1,
    totalPages: 1
  })

  const approvalStepperItems = computed(() => [
    {
      slot: 'draft' as const,
      value: 'draft',
      title: t('annualAudit.status.draft'),
      label: t('annualAudit.status.draft'),
      description: t('annualAudit.approvalRole.staff'),
      icon: 'i-lucide-check',
      disabled: true
    }, {
      slot: 'review' as const,
      value: 'review',
      title: t('annualAudit.status.pendingReview'),
      label: t('annualAudit.status.pendingReview'),
      description: t('annualAudit.approvalRole.manager'),
      icon: 'i-lucide-clock',
      disabled: true
    }, {
      slot: 'approval' as const,
      value: 'approval',
      title: t('annualAudit.status.pendingApproval'),
      label: t('annualAudit.status.pendingApproval'),
      description: t('annualAudit.approvalRole.chief'),
      icon: 'i-lucide-clock',
      disabled: true
    }, {
      slot: 'approved' as const,
      value: 'approved',
      title: t('annualAudit.status.approved'),
      label: t('annualAudit.status.approved'),
      description: t('annualAudit.approvalRole.system'),
      icon: 'i-lucide-check-circle',
      disabled: true
    }
  ]) as any

  const handleDownload = (plan: any) => {
    if (!plan) return

    const dataStr = JSON.stringify(plan, null, 2)
    const blob = new Blob([dataStr], { type: 'application/json' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')

    a.href = url
    a.download = `${plan.code}-annual-audit-plan.json`
    document.body.appendChild(a)
    a.click()

    document.body.removeChild(a)
    window.URL.revokeObjectURL(url)
  }

  const columns = computed<(TableColumn<AnnualAuditPlan> & { class?: string })[]>(() => [
    { accessorKey: 'activity', header: t('annualAudit.table.columns.activity'), class: 'w-44 whitespace-nowrap' },
    { accessorKey: 'department', header: t('annualAudit.detail.department'), class: 'w-36' },
    { accessorKey: 'riskName', header: t('annualAudit.table.columns.riskName'), class: 'w-52' },
    { accessorKey: 'riskLevel', header: t('annualAudit.table.columns.riskLevel'), class: 'w-28 whitespace-nowrap' },
    { accessorKey: 'status', header: t('annualAudit.form.status'), class: 'w-20 whitespace-nowrap text-center' },
    { accessorKey: 'actions', header: t('common.columns.actions'), class: 'w-36 whitespace-nowrap text-center' }
  ])

  const isEditing = ref(false)
  const editingId = ref<string | null>(null)
  // --- STATE UNTUK MODAL VIEW ---
  const showViewModal = ref(false)
  const selectedPlan = ref<any>(null)

  // --- STATE UNTUK FILTER ---
  const searchCode = ref('')
  const selectedDepartment = ref<string | undefined>(undefined)
  const selectedStatus = ref<string | undefined>(undefined)

  // --- OPSI UNTUK DROPDOWN FILTER ---
  const yearOptions = getFiscalYearStrings()
  // Department names from master-service (no fixed list). Empty on failure: no mock fallback.
  const departmentOptions = ref<string[]>([])
  // Master-data departments with their ids: the activity row stores departmentId.
  const departments = ref<Array<{ id: string, name: string }>>([])
  const departmentOptionsLoading = ref(false)
  // Filter options: value stays the canonical API string, label comes from i18n.
  const statusOptions = computed(() => Object.values(AnnualAuditPlanStatus).map((value): { value: string, label: string } => ({
    value,
    label: t(statusLabelKeys[value])
  })))

  const fetchDepartmentOptions = async () => {
    departmentOptionsLoading.value = true
    try {
      const list = await useDepartmentApi().getAllDepartments()
      departments.value = list
        .filter(d => d.id && d.department_name)
        .map(d => ({ id: d.id, name: d.department_name }))
      departmentOptions.value = departments.value.map(d => d.name)
    } finally {
      departmentOptionsLoading.value = false
    }
  }

  // riskName, riskLevel and department are server snapshots: they are read as they come
  // and never rewritten here, so a refreshed value from Kafka shows as is.
  // activityCode, justification, auditObjective and auditFocus are legacy values (moved to the
  // Audit Activity Plan): read for display only and never sent back.
  const normalizeActivity = (a: any): AuditActivities => ({
    ...a,
    id: a.id,
    itemNumber: String(a.itemNumber ?? a.item_number ?? ''),
    activityCode: a.activityCode || a.activity_code || '',
    name: a.name || a.title || '',
    category: a.category || AuditCategory.ASSURANCE,
    departmentId: a.departmentId || a.department_id || '',
    department: a.department || '',
    auditUniverseId: a.auditUniverseId || a.audit_universe_id || '',
    riskName: a.riskName || a.risk_name || '',
    riskLevel: a.riskLevel || a.risk_level || '',
    justification: a.justification || '',
    auditObjective: a.auditObjective || a.audit_objective || '',
    auditFocus: a.auditFocus || a.audit_focus || '',
    timelineText: a.timelineText || a.timeline_text || '',
    auditorCount: toCount(a.auditorCount ?? a.auditor_count, 0),
    totalMandays: toCount(a.totalMandays ?? a.total_mandays, 0),
    supervisorName: a.supervisorName || a.supervisor_name || '',
    activityPlanRef: a.activityPlanRef && typeof a.activityPlanRef === 'object' ? { ...a.activityPlanRef } : null
  })

  const normalizePlans = (rawPlans: any[]): AnnualAuditPlan[] => {
    if (!Array.isArray(rawPlans)) return []
    return rawPlans.map(p => ({
      ...p,
      code: p.code || p.plan_code || '',
      title: p.title || p.plan_title || '',
      activityType: p.activityType || p.activity_type || '',
      // Set by the server when an Audit Activity Plan or audit activity uses the plan.
      isUsed: p.isUsed === true || p.is_used === true,
      activities: Array.isArray(p.activities)
        ? p.activities.map(normalizeActivity)
        : [],
      selectedMonths: Array.isArray(p.selectedMonths) ? p.selectedMonths : (Array.isArray(p.selected_months) ? p.selected_months : []),
      quarters: Array.isArray(p.quarters) ? p.quarters : []
    }))
  }

  // Starts empty: the list only ever holds what GET /annual-audit-plans returned.
  const plans = ref<AnnualAuditPlan[]>([])

  // --- COMPUTED: FILTER DATA ---
  const filteredPlans = computed(() => {
    return (plans.value || []).filter((plan) => {
      const matchCode = !searchCode.value
        || (plan.code || '').toLowerCase().includes(searchCode.value.toLowerCase())
        || (plan.activities || []).some(act => (act.name || '').toLowerCase().includes(searchCode.value.toLowerCase()))

      const matchDept = !selectedDepartment.value
        || (plan.activities || []).some(act => act.department === selectedDepartment.value)

      const matchStatus = !selectedStatus.value || plan.status === selectedStatus.value

      return matchCode && matchDept && matchStatus
    })
  })

  // Fungsi Reset Filter
  const clearFilters = () => {
    searchCode.value = ''
    selectedDepartment.value = undefined
    selectedStatus.value = undefined
  }

  // --- ACTIONS UNTUK MODAL VIEW ---
  const openViewModal = (plan: any) => {
    selectedPlan.value = plan
    showViewModal.value = true
  }

  const closeViewModal = () => {
    showViewModal.value = false
    setTimeout(() => { selectedPlan.value = null }, 200)
  }

  watch(showViewModal, (isOpen) => {
    if (!isOpen) {
      setTimeout(() => { selectedPlan.value = null }, 200)
    }
  })

  const handleEditFromView = (plan: any) => {
    closeViewModal()
    handleEdit(plan)
  }

  const getStatusLabel = (status?: string) => {
    const key = statusLabelKeys[status as AnnualAuditPlanStatus]
    return key ? t(key) : (status || '-')
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case AnnualAuditPlanStatus.APPROVED: return 'bg-green-500'
      case AnnualAuditPlanStatus.PENDING_REVIEW:
      case AnnualAuditPlanStatus.PENDING_APPROVAL: return 'bg-amber-500'
      case AnnualAuditPlanStatus.REVISED: return 'bg-red-500'
      case AnnualAuditPlanStatus.DRAFT: return 'bg-gray-400'
      default: return 'bg-gray-200'
    }
  }

  const getStatusBadgeColor = (status?: string): 'success' | 'warning' | 'error' | 'neutral' => {
    switch (status) {
      case AnnualAuditPlanStatus.APPROVED:
        return 'success'
      case AnnualAuditPlanStatus.PENDING_REVIEW:
      case AnnualAuditPlanStatus.PENDING_APPROVAL:
        return 'warning'
      case AnnualAuditPlanStatus.REVISED:
        return 'error'
      case AnnualAuditPlanStatus.DRAFT:
      default:
        return 'neutral'
    }
  }

  const getRiskLevelColor = (level?: string) => {
    if (!level) return 'neutral'
    const lvl = level.toLowerCase()
    if (lvl.includes('high')) return 'error'
    if (lvl.includes('mod') || lvl.includes('medium')) return 'warning'
    if (lvl.includes('low')) return 'success'
    return 'neutral'
  }

  // State untuk menyimpan ID baris yang sedang "Read More"
  const expandedNotesRows = ref(new Set<string>())
  const expandedDetailRows = ref(new Set<string>())

  const toggleNotesReadMore = (id: string) => {
    if (expandedNotesRows.value.has(id)) {
      expandedNotesRows.value.delete(id)
    } else {
      expandedNotesRows.value.add(id)
    }
  }

  const toggleDetailReadMore = (id: string) => {
    if (expandedDetailRows.value.has(id)) {
      expandedDetailRows.value.delete(id)
    } else {
      expandedDetailRows.value.add(id)
    }
  }

  // Helper untuk mengecek apakah teks lebih dari 100 karakter
  const isNotesLongText = (text: string) => text && text.length > 100
  const isDetailLongText = (text: string) => text && text.length > 100

  // One empty PKAT row. Jumlah Tim defaults to 1. Kode Penugasan, Justifikasi, Tujuan and
  // Fokus are filled on the Audit Activity Plan.
  const createEmptyActivity = (): AuditActivities => ({
    name: '',
    category: AuditCategory.ASSURANCE,
    departmentId: '',
    department: '',
    auditUniverseId: '',
    riskName: '',
    riskLevel: '',
    timelineText: '',
    auditorCount: 1,
    totalMandays: 0,
    supervisorName: ''
  })

  // Form State
  const form = reactive<AnnualPlanForm>({
    code: '',
    title: '',
    version: 'v1.0',
    revisionHistory: [],
    activities: [createEmptyActivity()],
    status: AnnualAuditPlanStatus.DRAFT,
    selectedMonths: [],
    attachmentCategory: '',
    attachmentUploadedBy: '',
    attachmentUploadDate: '',
    file: [],
    notes: '',
    staffApprovalNote: '',
    managerApprovalNote: '',
    chiefApprovalNote: '',
    isActive: true,
    year: ''
  })

  // --- COMPUTED LOGIC (Real-time Validation) ---
  // Team mandays are the sum of the activity rows; there is no plan-level total.
  const totalMandays = computed(() => sumActivityMandays(form.activities))
  const utilizationData = computed(() => checkUtilization(totalMandays.value))
  const computedQuarters = computed(() => calculateQuarters(form.selectedMonths))
  const scheduleWarning = computed(() => checkScheduleGaps(form.selectedMonths))

  const toggleMonth = (idx: number) => {
    if (form.selectedMonths.includes(idx)) {
      form.selectedMonths = form.selectedMonths.filter(m => m !== idx)
    } else {
      form.selectedMonths.push(idx)
    }
  }

  const addActivity = () => {
    form.activities.push(createEmptyActivity())
  }

  const removeActivity = (index: number) => {
    // A row an Audit Activity Plan uses is removed there first (the server answers 409 otherwise).
    if (form.activities[index]?.activityPlanRef) return
    if (form.activities.length > 1) {
      form.activities.splice(index, 1)
    }
  }

  const openModal = () => {
    isEditing.value = false
    editingId.value = null

    clearValidationErrors()
    const defaultYear = new Date().getFullYear().toString()

    Object.assign(form, {
      code: '',
      title: '',
      version: 'v1.0',
      revisionHistory: [],
      activities: [createEmptyActivity()],
      status: AnnualAuditPlanStatus.DRAFT,
      selectedMonths: [],
      notes: '',
      staffApprovalNote: '',
      managerApprovalNote: '',
      chiefApprovalNote: '',
      file: [],
      attachmentCategory: '',
      attachmentUploadedBy: '',
      attachmentUploadDate: '',
      year: defaultYear
    })

    showModal.value = true
  }

  const closeModal = () => {
    showModal.value = false
  }

  watch(
    () => filteredPlans.value.length,
    (total) => {
      const size = paginations.value.pageSize || 10
      const pages = Math.ceil(total / size) || 1
      paginations.value.totalPages = pages
      if (paginations.value.currentIndex > pages) {
        paginations.value.currentIndex = Math.max(1, pages)
      }
    },
    { immediate: true }
  )

  // The crud List handler pages its results (page_size defaults to 20, max 100) and
  // returns { data: { items, pagination: { page, page_size, total, total_pages } } }.
  const PLAN_PAGE_SIZE = 100

  const extractPlanItems = (response: unknown): unknown[] => {
    const res = response as Record<string, unknown> | unknown[] | null | undefined
    if (Array.isArray(res)) return res
    const data = res?.data as Record<string, unknown> | unknown[] | undefined
    if (Array.isArray(data)) return data
    if (data && Array.isArray(data.items)) return data.items as unknown[]
    if (res && Array.isArray(res.items)) return res.items as unknown[]
    return []
  }

  const extractTotalPages = (response: unknown): number => {
    const res = response as Record<string, unknown> | null | undefined
    const data = res?.data as Record<string, unknown> | undefined
    const meta = (data && !Array.isArray(data) ? data.pagination : undefined) ?? res?.pagination
    const totalPages = Number((meta as Record<string, unknown> | undefined)?.total_pages)
    return Number.isFinite(totalPages) && totalPages > 1 ? totalPages : 1
  }

  // Loads every annual audit plan: the first page tells us total_pages, the rest are
  // fetched in parallel. Both /annual-audit (client-side table paging and filters) and
  // the dashboard coverage card need the full list, not just the first API page.
  const fetchPlans = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const url = `${getAuditServiceBaseUrl()}/annual-audit-plans`
      const fetchPage = (page: number) => $fetch<unknown>(url, {
        method: 'GET',
        query: { page, page_size: PLAN_PAGE_SIZE }
      })

      const first = await fetchPage(1)
      const totalPages = extractTotalPages(first)
      const rest = totalPages > 1
        ? await Promise.all(Array.from({ length: totalPages - 1 }, (_, i) => fetchPage(i + 2)))
        : []

      // De-duplicate by id in case rows shift between page requests.
      const seen = new Set<string>()
      const items = [first, ...rest].flatMap(extractPlanItems).filter((item) => {
        const id = (item as { id?: unknown })?.id
        if (id === undefined || id === null) return true
        const key = String(id)
        if (seen.has(key)) return false
        seen.add(key)
        return true
      })

      // An empty response really means "no plans" - no mock fallback.
      plans.value = normalizePlans(items)
    } catch (error: unknown) {
      // Keep whatever was loaded before (empty on first load); never show mock rows.
      console.warn('Failed to fetch annual audit plans:', error)
      errorMsg.value = extractErrorMessage(error, 'Gagal memuat rencana audit tahunan.')
    } finally {
      // The /annual-audit table pages the full list client-side, so keep its own page size.
      const pageSize = paginations.value.pageSize || 10
      const pages = Math.ceil(filteredPlans.value.length / pageSize) || 1
      paginations.value = {
        pageSize,
        currentIndex: Math.min(paginations.value.currentIndex || 1, pages),
        totalPages: pages
      }
      loading.value = false
    }
  }

  const clearValidationErrors = () => {
    validationErrors.activityDetail = ''
    validationErrors.timeline = ''
    validationErrors.auditor = ''
  }

  const validateForm = () => {
    clearValidationErrors()

    const activityErrors: string[] = []
    const timelineErrors: string[] = []
    const auditorErrors: string[] = []

    // 1. Activity Detail
    // New plans get their code from the server on save; only an existing plan must still carry one.
    if (isEditing.value && !form.code?.trim()) {
      activityErrors.push('Activity Code wajib diisi.')
    }

    if (!form.activities.length) {
      activityErrors.push('Minimal harus ada 1 aktivitas.')
    }

    form.activities.forEach((activity, index) => {
      const number = index + 1

      if (!activity.name?.trim()) {
        activityErrors.push(t('annualAudit.form.validation.objectRequired', { number }))
      }

      if (!activity.category) {
        activityErrors.push(t('annualAudit.form.validation.categoryRequired', { number }))
      }

      if (!activity.departmentId) {
        activityErrors.push(t('annualAudit.form.validation.departmentRequired', { number }))
      }

      if (!activity.auditUniverseId) {
        activityErrors.push(t('annualAudit.form.validation.universeRequired', { number }))
      }

      // Team and mandays are per row now: each row must carry its own team.
      const auditors = Number(activity.auditorCount)
      if (!Number.isInteger(auditors) || auditors < 1 || auditors > 10) {
        auditorErrors.push(t('annualAudit.form.validation.auditorCountRange', { number }))
      }

      const mandays = Number(activity.totalMandays)
      if (!Number.isFinite(mandays) || mandays < 1) {
        auditorErrors.push(t('annualAudit.form.validation.mandaysMin', { number }))
      }

      if (!activity.supervisorName?.trim()) {
        auditorErrors.push(t('annualAudit.form.validation.leadRequired', { number }))
      }
    })

    // 2. Timeline
    if (!form.year) {
      timelineErrors.push('Year wajib dipilih.')
    }

    if (form.selectedMonths.length === 0) {
      timelineErrors.push('Minimal 1 bulan pelaksanaan wajib dipilih.')
    }

    validationErrors.activityDetail = activityErrors.join(' ')
    validationErrors.timeline = timelineErrors.join(' ')
    validationErrors.auditor = auditorErrors.join(' ')

    return (
      activityErrors.length === 0
      && timelineErrors.length === 0
      && auditorErrors.length === 0
    )
  }

  const handleSubmit = async () => {
    if (!validateForm()) {
      toast.showError('Mohon lengkapi data pada Activity Detail, Timeline, dan Auditor.')
      return
    }

    try {
      if (isEditing.value && editingId.value) {
        await updatePlan(editingId.value, { ...form })
        toast.showSuccess('Data Rencana Audit Berhasil Diperbarui!')
      } else {
        const createdCode = await addPlan({ ...form })
        if (createdCode) form.code = createdCode
        toast.showSuccess('Data Rencana Audit Berhasil Disimpan!')
      }

      closeModal()
    } catch {
      // addPlan/updatePlan already showed the error toast (and the form message for a 409).
    }
  }

  const handleEdit = (plan: any) => {
    isEditing.value = true
    editingId.value = plan.id

    clearValidationErrors()

    form.code = plan.code
    form.version = plan.version || 'v1.0'
    form.revisionHistory = plan.revisionHistory ? [...plan.revisionHistory] : []
    // Rows keep their id (PUT matches on it) and the server snapshots as they came back.
    form.activities = plan.activities.map((act: any) => ({ ...act }))
    form.status = plan.status
    form.selectedMonths = [...plan.selectedMonths]
    form.title = plan.title || ''
    form.notes = plan.notes || ''
    form.file = []
    form.attachmentCategory = plan.attachmentCategory || ''
    form.attachmentUploadedBy = plan.attachmentUploadedBy || ''
    form.attachmentUploadDate = plan.attachmentUploadDate || ''
    form.isActive = plan.isActive
    form.year = plan.year || ''

    showModal.value = true
  }

  const handleDelete = async (id: string | undefined) => {
    if (!id) return
    let result: DeletePlanResult
    try {
      result = await deletePlan(id)
    } catch {
      // deletePlan already showed the error toast.
      return
    }
    // Cancelled in the confirmation (or the plan is gone): nothing happened, so no toast.
    if (!result) return
    toast.showSuccess(t(result === 'deactivated' ? 'annualAudit.delete.deactivated' : 'annualAudit.delete.success'))
    closeViewModal()
  }

  const handleFileChange = (event: Event) => {
    const target = event.target as HTMLInputElement
    const file = target.files?.[0]
    if (!file) return

    if (file.size > 5 * 1024 * 1024) {
      errorMsg.value = t('common.fileInput.tooLarge', { size: 5 })
      form.file = null
      target.value = ''
      return
    }

    const allowedTypes = [
      'application/pdf',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      'application/msword'
    ]

    if (!allowedTypes.includes(file.type)) {
      errorMsg.value = t('common.fileInput.invalidPdfDocx')
      form.file = null
      target.value = ''
      return
    }

    errorMsg.value = ''
    form.file = [file]
  }

  const TEAM_CAPACITY_PER_YEAR = 10 * 220 * 0.7

  const calculateQuarters = (months: number[]) => {
    const qSet = new Set<string>()
    months.forEach((m) => {
      if (m >= 0 && m <= 2) qSet.add('Q1')
      else if (m >= 3 && m <= 5) qSet.add('Q2')
      else if (m >= 6 && m <= 8) qSet.add('Q3')
      else qSet.add('Q4')
    })
    return Array.from(qSet).sort()
  }

  const checkUtilization = (additionalMandays: number) => {
    // The plan being edited is replaced by the form, so its stored rows are not counted twice.
    const currentUsed = plans.value
      .filter(p => !(isEditing.value && p.id === editingId.value))
      .reduce((sum, p) => sum + sumActivityMandays(p.activities), 0)
    const totalAfterAdd = currentUsed + additionalMandays
    const utilization = (totalAfterAdd / TEAM_CAPACITY_PER_YEAR) * 100

    if (utilization > 95) return { color: 'red', percent: utilization, msg: 'OVERLOAD (>95%)' }
    if (utilization > 80) return { color: 'yellow', percent: utilization, msg: 'High Load (80-95%)' }
    return { color: 'green', percent: utilization, msg: 'Optimal (60-80%)' }
  }

  const checkScheduleGaps = (months: number[]) => {
    if (months.length === 0) return 'Wajib pilih minimal 1 bulan.'
    const sorted = [...months].sort((a, b) => a - b)
    for (let i = 0; i < sorted.length - 1; i++) {
      if (sorted[i + 1]! - sorted[i]! > 1) {
        return 'Warning: Ada gap bulan kosong. Pastikan continuous coverage jika diperlukan.'
      }
    }
    return null
  }

  // New files only; the server stores them under the plan.
  const toNewAttachments = (files: File[] | null | undefined): AnnualAuditAttachment[] =>
    (files || []).map(f => ({
      name: f.name,
      size: Math.round(f.size / 1024) + ' KB',
      url: '#'
    }))

  /**
   * Message for a failed write: the Audit Activity Plan 409s are translated with their details
   * (row, plan title); anything else keeps the backend message.
   */
  const describeWriteError = (error: unknown, fallback: string): string => {
    const { code } = parseApiError(error)
    return code && ANNUAL_PLAN_ERROR_CODES.has(code)
      ? translateApiError(error, t, 'annualAudit.errors')
      : extractErrorMessage(error, fallback)
  }

  /** A 409 about a row used by an Audit Activity Plan is also shown in the form, next to the rows. */
  const showRowConflictInForm = (error: unknown, message: string) => {
    const { code } = parseApiError(error)
    if (code === 'ANNUAL_PLAN_ACTIVITY_IN_USE' || code === 'ANNUAL_PLAN_ACTIVITY_TYPE_CONFLICT') {
      validationErrors.activityDetail = message
    }
  }

  const addPlan = async (form: AnnualPlanForm) => {
    const quarters = calculateQuarters(form.selectedMonths)

    // No code is sent: the server generates PKAT-<Tahun>-<No.Urut> in the create transaction.
    const payload = buildAnnualPlanPayload(form, {
      quarters,
      attachments: toNewAttachments(form.file)
    })

    const baseUrl = getAuditServiceBaseUrl()
    try {
      const response: any = await $fetch(`${baseUrl}/annual-audit-plans`, {
        method: 'POST',
        body: payload
      })
      await fetchPlans()
      // Read from the create response; shape not yet confirmed against the backend.
      const createdCode = response?.data?.code ?? response?.code
      return typeof createdCode === 'string' ? createdCode : ''
    } catch (error: any) {
      console.error('Failed to create annual plan:', error)
      const detail = describeWriteError(error, 'Gagal menambahkan rencana audit.')
      errorMsg.value = detail
      toast.showError('Gagal menambahkan rencana audit.', detail)
      throw error
    }
  }

  const updatePlan = async (id: string, updatedData: AnnualPlanForm) => {
    const baseUrl = getAuditServiceBaseUrl()
    const quarters = calculateQuarters(updatedData.selectedMonths)

    const payload = {
      code: updatedData.code,
      ...buildAnnualPlanPayload(updatedData, {
        quarters,
        attachments: (updatedData.attachments || []).concat(toNewAttachments(updatedData.file))
      })
    }

    try {
      await $fetch(`${baseUrl}/annual-audit-plans/${id}`, {
        method: 'PUT',
        body: payload
      })
      await fetchPlans()
    } catch (error: any) {
      console.error('Failed to update annual plan:', error)
      const detail = describeWriteError(error, 'Gagal memperbarui rencana audit.')
      errorMsg.value = detail
      showRowConflictInForm(error, detail)
      toast.showError('Gagal memperbarui rencana audit.', detail)
      throw error
    }
  }

  /**
   * Asks for confirmation, then deletes the plan (or deactivates it when it is already used).
   * Resolves to what happened, or null when the user cancelled: no request is sent then.
   * A failed request shows the error toast and rethrows.
   */
  const deletePlan = async (id: string): Promise<DeletePlanResult> => {
    const plan = plans.value.find(a => a.id === id)
    if (!plan) return null

    const planLine = plan.code ? [t('annualAudit.delete.planLabel', { code: plan.code })] : []
    const confirmed = plan.isUsed
      ? await useGlobalModalStore().confirmDelete({
          description: t('annualAudit.delete.deactivateTitle'),
          body: [...planLine, t('annualAudit.delete.deactivateBody')],
          confirmLabel: t('annualAudit.delete.deactivateConfirm')
        })
      : await useGlobalModalStore().confirmDelete({
          description: t('annualAudit.delete.confirmTitle'),
          body: [...planLine, t('annualAudit.delete.confirmBody')]
        })
    if (!confirmed) return null

    const baseUrl = getAuditServiceBaseUrl()
    // PUT is partial: only the field that changes is sent.
    const deactivate = () => $fetch(`${baseUrl}/annual-audit-plans/${id}`, {
      method: 'PUT',
      body: { isActive: false }
    })
    let result: DeletePlanResult = plan.isUsed ? 'deactivated' : 'deleted'
    try {
      if (plan.isUsed) {
        await deactivate()
      } else {
        try {
          await $fetch(`${baseUrl}/annual-audit-plans/${id}`, {
            method: 'DELETE'
          })
        } catch (error) {
          // The server refuses to delete a plan an Audit Activity Plan uses (409 ANNUAL_PLAN_IN_USE),
          // also when the list did not know it yet: offer to deactivate it instead.
          if (parseApiError(error).code !== 'ANNUAL_PLAN_IN_USE' || plan.isActive === false) throw error
          const deactivateConfirmed = await useGlobalModalStore().confirmDelete({
            description: t('annualAudit.delete.deactivateTitle'),
            body: [...planLine, t('annualAudit.delete.deactivateBody')],
            confirmLabel: t('annualAudit.delete.deactivateConfirm')
          })
          if (!deactivateConfirmed) {
            // Refresh so isUsed now shows the plan as in use.
            await fetchPlans()
            return null
          }
          await deactivate()
          result = 'deactivated'
        }
      }
    } catch (error: any) {
      console.error('Failed to delete annual plan:', error)
      const detail = describeWriteError(error, t('annualAudit.delete.error'))
      errorMsg.value = detail
      toast.showError(t('annualAudit.delete.error'), detail)
      throw error
    }
    await fetchPlans()
    return result
  }

  const createRevision = async (planId: string, changesNote: string, user: string) => {
    const plan = plans.value.find(p => p.id === planId)
    if (!plan) return

    const currentVersion = plan.version || 'v1.0'
    const parts = currentVersion.replace('v', '').split('.')
    const minor = parseInt(parts[1] || '0')
    const newVersion = `v${parts[0]}.${minor + 1}`

    const historyEntry = {
      date: new Date().toISOString().split('T')[0] || '',
      version: newVersion,
      changes: changesNote,
      user: user
    }

    const revisionHistory = plan.revisionHistory ? [...plan.revisionHistory] : []
    revisionHistory.unshift(historyEntry)

    // PUT is partial: only the fields that change are sent.
    const payload = {
      version: newVersion,
      status: AnnualAuditPlanStatus.REVISED,
      revisionHistory: revisionHistory
    }

    const baseUrl = getAuditServiceBaseUrl()
    try {
      await $fetch(`${baseUrl}/annual-audit-plans/${planId}`, {
        method: 'PUT',
        body: payload
      })
      await fetchPlans()
      toast.showSuccess(`Revised RKAT created (Version ${newVersion}).`)
    } catch (error: any) {
      console.error('Failed to create revision:', error)
      const detail = extractErrorMessage(error, 'Gagal membuat revisi RKAT.')
      errorMsg.value = detail
      toast.showError('Gagal membuat revisi RKAT.', detail)
      throw error
    }
  }

  const updatePlanStatus = async (id: string, status: AnnualAuditPlanStatus, extraPayload: Record<string, any> = {}) => {
    const plan = plans.value.find(p => p.id === id)
    if (!plan) return

    const baseUrl = getAuditServiceBaseUrl()
    // PUT is partial: only the status and optional notes are sent.
    const payload = { status: status, ...extraPayload }
    try {
      await $fetch(`${baseUrl}/annual-audit-plans/${id}`, {
        method: 'PUT',
        body: payload
      })
      await fetchPlans()
      if (selectedPlan.value && selectedPlan.value.id === id) {
        selectedPlan.value = plans.value.find(p => p.id === id) || { ...selectedPlan.value, status, ...extraPayload }
      }
      toast.showSuccess(`Status Rencana Audit berhasil diperbarui menjadi ${getStatusLabel(status)}`)
    } catch (error: any) {
      console.error('Failed to update plan status:', error)
      const detail = extractErrorMessage(error, 'Gagal memperbarui status rencana audit.')
      errorMsg.value = detail
      toast.showError('Gagal memperbarui status rencana audit.', detail)
      throw error
    }
  }

  // Transitions: Draft|Revised -> submit (auditor) -> Pending Review
  const handleStaffApprove = async () => {
    if (selectedPlan.value) {
      await updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.PENDING_REVIEW, {
        staffApprovalNote: selectedPlan.value.staffApprovalNote || ''
      })
      selectedPlan.value = plans.value.find(p => p.id === selectedPlan.value.id) || selectedPlan.value
    }
  }

  const handleStaffReject = () => {
    if (selectedPlan.value) {
      selectedPlan.value.staffApprovalNote = ''
    }
  }

  const handleManagerApprove = async () => {
    if (selectedPlan.value) {
      await updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.PENDING_APPROVAL, {
        managerApprovalNote: selectedPlan.value.managerApprovalNote || ''
      })
      selectedPlan.value = plans.value.find(p => p.id === selectedPlan.value.id) || selectedPlan.value
    }
  }

  // Pending Review -> manager reject -> Revised
  const handleManagerReject = async () => {
    if (selectedPlan.value) {
      await updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.REVISED, {
        managerApprovalNote: selectedPlan.value.managerApprovalNote || ''
      })
      selectedPlan.value = plans.value.find(p => p.id === selectedPlan.value.id) || selectedPlan.value
    }
  }

  // Pending Approval -> CAE approve -> Approved
  const handleChiefApprove = async () => {
    if (selectedPlan.value) {
      await updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.APPROVED, {
        chiefApprovalNote: selectedPlan.value.chiefApprovalNote || ''
      })
      selectedPlan.value = plans.value.find(p => p.id === selectedPlan.value.id) || selectedPlan.value
    }
  }

  // Pending Approval -> CAE reject -> Revised
  const handleChiefReject = async () => {
    if (selectedPlan.value) {
      await updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.REVISED, {
        chiefApprovalNote: selectedPlan.value.chiefApprovalNote || ''
      })
      selectedPlan.value = plans.value.find(p => p.id === selectedPlan.value.id) || selectedPlan.value
    }
  }

  return {
    plans, monthsList, yearOptions, attachmentCategoryOptions,
    showModal, isEditing, editingId, showViewModal, selectedPlan, approvalStepperItems,
    searchCode, selectedDepartment, selectedStatus, form, columns, errorMsg, loading, validationErrors,
    filteredPlans, totalMandays, scheduleWarning, utilizationData,
    computedQuarters, paginations, pagination: paginations,
    clearFilters, openViewModal, closeViewModal, departmentOptions, departments, departmentOptionsLoading, fetchDepartmentOptions, statusOptions,
    toggleMonth, addActivity, removeActivity, handleDownload,
    openModal, closeModal, handleSubmit, handleEdit, handleEditFromView, handleDelete,
    getStatusColor, getStatusBadgeColor, getStatusLabel, handleFileChange, fetchPlans, fetchAnnualPlans: fetchPlans,
    handleStaffApprove, handleStaffReject, handleManagerApprove, handleManagerReject, handleChiefApprove, handleChiefReject,
    createRevision, getRiskLevelColor
  }
})
