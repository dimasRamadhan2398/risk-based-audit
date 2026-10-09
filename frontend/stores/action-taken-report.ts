import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ActionTakenReport, AtrEvidence } from '~/types/audit'
import { extractErrorMessage, getErrorStatus, getUserErrorMessage, parseBlobErrorBody } from '~/utils/error'
import { getAuditServiceBaseUrl, getAuthServiceBaseUrl } from '~/composables/useApiUrl'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import {
  ATR_DEFAULT_PAGE_SIZE,
  ATR_MAX_PAGE_SIZE,
  buildAtrParams,
  computeAtrStats,
  defaultAtrQuery,
  emptyAtrPagination,
  isLhaApproved,
  normalizeAtrItem,
  parseAtrPagination,
  toAtrDueDatePayload,
  type AtrPagination,
  type AtrQuery,
  type AtrScope
} from '~/utils/actionTakenReport'

/** Approved LHA offered in the ATR "LHA" filter. */
export interface AtrLhaOption {
  id: string
  reportNumber: string
  reportTitle: string
}

/** User that can be picked as PIC (GET /users/assignable on the auth service). */
export interface AtrPicCandidate {
  id: string
  full_name: string
  department: string
  position: string
}

type ApiEnvelope<T> = { success?: boolean, message?: string, data?: T }

const SEARCH_DEBOUNCE_MS = 300
// Pages fetched for the summary/dashboard stats (every ATR the user can see).
const SUMMARY_PAGE_SIZE = ATR_MAX_PAGE_SIZE
const SUMMARY_MAX_PAGES = 50

export const useActionTakenReportStore = defineStore('action-taken-report', () => {
  const { t } = useI18n()
  const toast = useToastNotification()
  const rbac = useRbac()

  const baseUrl = () => `${getAuditServiceBaseUrl()}/action-taken-reports`

  // --- List (server-side paged and filtered) -------------------------------------------------
  const items = ref<ActionTakenReport[]>([])
  const pagination = ref<AtrPagination>(emptyAtrPagination())
  // `search` here is the applied (debounced) term; the input box holds the live text.
  const query = ref<AtrQuery>(defaultAtrQuery())
  const loading = ref(false)
  const errorMsg = ref('')
  let requestId = 0
  let searchTimer: ReturnType<typeof setTimeout> | null = null

  const fetchReports = async () => {
    const id = ++requestId
    const q = { ...query.value }
    loading.value = true
    errorMsg.value = ''
    try {
      const response = await $fetch<ApiEnvelope<{ items?: unknown[], pagination?: Partial<AtrPagination> }>>(baseUrl(), {
        method: 'GET',
        params: buildAtrParams(q, rbac.getCurrentUserId())
      })
      // A newer request (filter/page change) was started meanwhile; its result wins.
      if (id !== requestId) return

      const rows = Array.isArray(response?.data?.items) ? response.data.items : []
      const meta = parseAtrPagination(response?.data?.pagination, q)

      // The requested page no longer exists (rows were closed/cancelled meanwhile): go to the last one.
      if (rows.length === 0 && meta.total > 0 && meta.total_pages > 0 && q.page > meta.total_pages) {
        query.value.page = meta.total_pages
        await fetchReports()
        return
      }

      const now = new Date()
      items.value = rows.map(row => normalizeAtrItem(row as Record<string, unknown>, now))
      pagination.value = meta
      query.value.page = meta.page
      query.value.pageSize = meta.page_size
    } catch (error) {
      if (id !== requestId) return
      console.error('Failed to fetch action taken reports:', error)
      errorMsg.value = extractErrorMessage(error, t('actionTakenReport.errors.load'))
      items.value = []
      pagination.value = emptyAtrPagination(q.pageSize)
    } finally {
      if (id === requestId) loading.value = false
    }
  }

  const cancelSearch = () => {
    if (searchTimer) {
      clearTimeout(searchTimer)
      searchTimer = null
    }
  }

  /** Change filters/scope/page size; anything that changes goes back to page 1 and refetches. */
  const setFilters = (changes: Partial<Omit<AtrQuery, 'page'>>) => {
    const current = query.value
    const next: AtrQuery = { ...current, ...changes, page: 1 }
    next.pageSize = Math.min(ATR_MAX_PAGE_SIZE, Math.max(1, Number(next.pageSize) || ATR_DEFAULT_PAGE_SIZE))
    next.search = (next.search ?? '').trim()
    next.status = next.status ?? ''
    next.lhaId = next.lhaId ?? ''
    next.picUserId = next.picUserId ?? ''
    next.overdue = !!next.overdue
    const changed = (Object.keys(changes) as Array<keyof AtrQuery>).some(k => next[k] !== current[k])
    if (!changed) return
    if ('search' in changes) cancelSearch()
    query.value = next
    return fetchReports()
  }

  /** Debounced search: only the last term typed within the window is sent. */
  const setSearch = (term: string) => {
    cancelSearch()
    const value = (term ?? '').trim()
    if (value === query.value.search) return
    searchTimer = setTimeout(() => {
      searchTimer = null
      setFilters({ search: value })
    }, SEARCH_DEBOUNCE_MS)
  }

  const setPage = (page: number) => {
    const target = Math.max(1, Math.trunc(Number(page)) || 1)
    if (target === query.value.page) return
    query.value.page = target
    return fetchReports()
  }

  const resetFilters = () => {
    cancelSearch()
    return setFilters({ search: '', status: '', lhaId: '', picUserId: '', overdue: false })
  }

  /**
   * First load of the ATR page. Users who cannot view every ATR always work in "My actions".
   * `lhaId` comes from ?lha= (the LHA page's "View follow-ups" link).
   */
  const initList = (options: { scope: AtrScope, lhaId?: string }) => {
    cancelSearch()
    query.value = { ...defaultAtrQuery(options.scope), pageSize: query.value.pageSize, lhaId: options.lhaId ?? '' }
    return fetchReports()
  }

  // --- Summary / dashboard stats (all ATRs the user can see) ----------------------------------
  const summaryItems = ref<ActionTakenReport[]>([])
  const summaryLoading = ref(false)
  const summaryError = ref('')
  let summaryRequestId = 0

  const fetchSummary = async () => {
    const id = ++summaryRequestId
    summaryLoading.value = true
    summaryError.value = ''
    try {
      const collected: unknown[] = []
      let page = 1
      let totalPages = 1
      do {
        const response = await $fetch<ApiEnvelope<{ items?: unknown[], pagination?: Partial<AtrPagination> }>>(baseUrl(), {
          method: 'GET',
          params: { page, page_size: SUMMARY_PAGE_SIZE }
        })
        if (id !== summaryRequestId) return
        const rows = Array.isArray(response?.data?.items) ? response.data.items : []
        collected.push(...rows)
        totalPages = parseAtrPagination(response?.data?.pagination, { page, pageSize: SUMMARY_PAGE_SIZE }).total_pages
        page++
      } while (page <= totalPages && page <= SUMMARY_MAX_PAGES)
      const now = new Date()
      summaryItems.value = collected.map(row => normalizeAtrItem(row as Record<string, unknown>, now))
    } catch (error) {
      if (id !== summaryRequestId) return
      console.error('Failed to fetch action taken report summary:', error)
      summaryError.value = extractErrorMessage(error, t('actionTakenReport.errors.load'))
      summaryItems.value = []
    } finally {
      if (id === summaryRequestId) summaryLoading.value = false
    }
  }

  // Single source of truth for the ATR page summary and the dashboard (see computeAtrStats).
  const stats = computed(() => computeAtrStats(summaryItems.value))

  // --- Filter options --------------------------------------------------------------------------
  const lhaOptions = ref<AtrLhaOption[]>([])
  const lhaOptionsLoading = ref(false)

  /** Approved LHAs (GET /audit-result-reports) for the LHA filter. */
  const fetchLhaOptions = async () => {
    lhaOptionsLoading.value = true
    try {
      const response = await $fetch<ApiEnvelope<{ items?: Record<string, unknown>[] }> | Record<string, unknown>[]>(
        `${getAuditServiceBaseUrl()}/audit-result-reports`,
        { method: 'GET' }
      )
      const rows = Array.isArray(response)
        ? response
        : (Array.isArray(response?.data?.items) ? response.data.items : [])
      lhaOptions.value = rows
        .filter(r => isLhaApproved(r.status))
        .map(r => ({
          id: String(r.id ?? ''),
          reportNumber: String(r.reportNumber ?? r.report_number ?? ''),
          reportTitle: String(r.reportTitle ?? r.report_title ?? '')
        }))
        .filter(r => r.id)
    } catch (error) {
      // The filter is optional; the list still works without it.
      console.error('Failed to fetch approved audit result reports:', error)
      lhaOptions.value = []
    } finally {
      lhaOptionsLoading.value = false
    }
  }

  /**
   * Users that can be assigned as PIC. Server-side search; 403 means the user may not list users
   * (`forbidden: true`), which the caller shows instead of a list.
   */
  const fetchPicCandidates = async (search = '', pageSize = 20): Promise<{ users: AtrPicCandidate[], forbidden: boolean, error: string }> => {
    try {
      const response = await $fetch<ApiEnvelope<{ users?: Record<string, unknown>[] }>>(`${getAuthServiceBaseUrl()}/users/assignable`, {
        method: 'GET',
        params: { search: search.trim() || undefined, page: 1, page_size: pageSize }
      })
      const users = (Array.isArray(response?.data?.users) ? response.data.users : [])
        .map(u => ({
          id: String(u.id ?? ''),
          full_name: String(u.full_name ?? ''),
          department: String(u.department ?? ''),
          position: String(u.position ?? '')
        }))
        .filter(u => u.id)
      return { users, forbidden: false, error: '' }
    } catch (error) {
      if (getErrorStatus(error) === 403) return { users: [], forbidden: true, error: '' }
      console.error('Failed to fetch assignable users:', error)
      return { users: [], forbidden: false, error: extractErrorMessage(error, t('actionTakenReport.errors.loadUsers')) }
    }
  }

  /** Number of ATRs of one LHA (pagination.total of a 1-row page); null when it cannot be read. */
  const countForLha = async (lhaId: string): Promise<number | null> => {
    if (!lhaId) return null
    try {
      const response = await $fetch<ApiEnvelope<{ pagination?: Partial<AtrPagination> }>>(baseUrl(), {
        method: 'GET',
        params: { audit_result_report_id: lhaId, page: 1, page_size: 1 }
      })
      return parseAtrPagination(response?.data?.pagination, { page: 1, pageSize: 1 }).total
    } catch (error) {
      console.error('Failed to count action taken reports:', error)
      return null
    }
  }

  // --- Current item and modals -----------------------------------------------------------------
  const current = ref<ActionTakenReport | null>(null)
  const detailLoading = ref(false)
  const saving = ref(false)
  const uploading = ref(false)
  const downloadingEvidenceId = ref<string | null>(null)

  const showDetail = ref(false)
  const showActionPlan = ref(false)
  const showAssignment = ref(false)
  const showReview = ref(false)
  const showCancel = ref(false)

  const closeAll = () => {
    showDetail.value = false
    showActionPlan.value = false
    showAssignment.value = false
    showReview.value = false
    showCancel.value = false
  }

  /** GET /action-taken-reports/:id; the modals work on this fresh copy, not on the list row. */
  const fetchReport = async (id: string): Promise<ActionTakenReport | null> => {
    detailLoading.value = true
    try {
      const response = await $fetch<ApiEnvelope<Record<string, unknown>>>(`${baseUrl()}/${id}`, { method: 'GET' })
      const item = response?.data ? normalizeAtrItem(response.data) : null
      if (item && current.value?.id === id) current.value = item
      return item
    } catch (error) {
      console.error('Failed to fetch action taken report:', error)
      toast.showError(t('actionTakenReport.errors.loadDetailTitle'), extractErrorMessage(error, t('actionTakenReport.errors.loadDetail')))
      return null
    } finally {
      detailLoading.value = false
    }
  }

  type ModalName = 'detail' | 'actionPlan' | 'assignment' | 'review' | 'cancel'
  const modalFlags: Record<ModalName, typeof showDetail> = {
    detail: showDetail,
    actionPlan: showActionPlan,
    assignment: showAssignment,
    review: showReview,
    cancel: showCancel
  }

  /** Open a modal for a row: show the row immediately, then replace it with the server copy. */
  const open = async (modal: ModalName, row: ActionTakenReport) => {
    closeAll()
    current.value = row
    modalFlags[modal].value = true
    await fetchReport(row.id)
  }

  const openDetail = (row: ActionTakenReport) => open('detail', row)
  const openActionPlan = (row: ActionTakenReport) => open('actionPlan', row)
  const openAssignment = (row: ActionTakenReport) => open('assignment', row)
  const openReview = (row: ActionTakenReport) => open('review', row)
  const openCancel = (row: ActionTakenReport) => open('cancel', row)

  const closeModal = () => {
    closeAll()
    current.value = null
  }

  /** After a change: take the server's item (or refetch it), and refresh the list and the summary. */
  const afterChange = async (id: string, data: unknown) => {
    const updated = data && typeof data === 'object' && (data as Record<string, unknown>).id
      ? normalizeAtrItem(data as Record<string, unknown>)
      : null
    if (updated) {
      if (current.value?.id === id) current.value = updated
      const index = items.value.findIndex(i => i.id === id)
      if (index >= 0) items.value[index] = updated
    } else if (current.value?.id === id) {
      await fetchReport(id)
    }
    await Promise.all([fetchReports(), fetchSummary()])
  }

  /**
   * Runs one action: toasts success/failure and refreshes. Returns true on success.
   * Errors are not rethrown so the modal stays open with the user's input.
   */
  const runAction = async (
    id: string,
    request: () => Promise<ApiEnvelope<unknown> | undefined>,
    successKey: string,
    errorKey: string
  ): Promise<boolean> => {
    saving.value = true
    try {
      const response = await request()
      await afterChange(id, response?.data)
      toast.showSuccess(t(successKey))
      return true
    } catch (error) {
      console.error(`ATR action failed (${successKey}):`, error)
      toast.showError(t(errorKey), extractErrorMessage(error, t('actionTakenReport.errors.generic')))
      return false
    } finally {
      saving.value = false
    }
  }

  /** PUT /:id/assignment { pic_user_id, due_date } (auditor/manager/admin). due_date is 'YYYY-MM-DD' from the date picker. */
  const assign = (id: string, payload: { pic_user_id: string, due_date: string }) =>
    runAction(
      id,
      () => $fetch<ApiEnvelope<unknown>>(`${baseUrl()}/${id}/assignment`, {
        method: 'PUT',
        body: { pic_user_id: payload.pic_user_id, due_date: toAtrDueDatePayload(payload.due_date) }
      }),
      'actionTakenReport.toast.assigned',
      'actionTakenReport.toast.assignFailed'
    )

  /** PUT /:id/action-plan { action_plan, progress } (assigned PIC). */
  const saveActionPlan = (id: string, payload: { action_plan: string, progress: number }) =>
    runAction(
      id,
      () => $fetch<ApiEnvelope<unknown>>(`${baseUrl()}/${id}/action-plan`, {
        method: 'PUT',
        body: { action_plan: payload.action_plan.trim(), progress: Math.round(payload.progress) }
      }),
      'actionTakenReport.toast.planSaved',
      'actionTakenReport.toast.planSaveFailed'
    )

  /** POST /:id/submit (assigned PIC): IN_PROGRESS -> PENDING_REVIEW. */
  const submit = (id: string) =>
    runAction(
      id,
      () => $fetch<ApiEnvelope<unknown>>(`${baseUrl()}/${id}/submit`, { method: 'POST' }),
      'actionTakenReport.toast.submitted',
      'actionTakenReport.toast.submitFailed'
    )

  /** POST /:id/review { decision, note } (auditor/manager/CAE/admin). */
  const review = (id: string, payload: { decision: 'approve' | 'reject', note: string }) =>
    runAction(
      id,
      () => $fetch<ApiEnvelope<unknown>>(`${baseUrl()}/${id}/review`, {
        method: 'POST',
        body: { decision: payload.decision, note: payload.note.trim() }
      }),
      payload.decision === 'approve' ? 'actionTakenReport.toast.approved' : 'actionTakenReport.toast.rejected',
      'actionTakenReport.toast.reviewFailed'
    )

  /** POST /:id/cancel { note } (manager/CAE/admin). */
  const cancel = (id: string, note: string) =>
    runAction(
      id,
      () => $fetch<ApiEnvelope<unknown>>(`${baseUrl()}/${id}/cancel`, { method: 'POST', body: { note: note.trim() } }),
      'actionTakenReport.toast.cancelled',
      'actionTakenReport.toast.cancelFailed'
    )

  /**
   * POST /:id/evidence, one multipart request per file (field `file`). No Content-Type header:
   * the browser sets the multipart boundary. Returns the number of files uploaded.
   */
  const uploadEvidence = async (id: string, files: File[]): Promise<number> => {
    if (!files.length) return 0
    uploading.value = true
    let uploaded = 0
    let lastResponse: ApiEnvelope<unknown> | undefined
    try {
      for (const file of files) {
        const form = new FormData()
        form.append('file', file)
        try {
          lastResponse = await $fetch<ApiEnvelope<unknown>>(`${baseUrl()}/${id}/evidence`, { method: 'POST', body: form })
          uploaded++
        } catch (error) {
          console.error('Failed to upload ATR evidence:', error)
          toast.showError(
            t('actionTakenReport.toast.uploadFailed', { name: file.name }),
            extractErrorMessage(error, t('actionTakenReport.errors.generic'))
          )
        }
      }
      if (uploaded > 0) {
        // The upload response may be the evidence row rather than the ATR, so always refetch the item.
        const data = lastResponse?.data as Record<string, unknown> | undefined
        await afterChange(id, data && Array.isArray(data.evidence) ? data : null)
        if (!(data && Array.isArray(data.evidence)) && current.value?.id === id) await fetchReport(id)
        toast.showSuccess(t('actionTakenReport.toast.uploaded', { count: uploaded }))
      }
      return uploaded
    } finally {
      uploading.value = false
    }
  }

  /** GET /:id/evidence/:evidenceId as a blob (the auth plugin adds the token) and save it under its file name. */
  const downloadEvidence = async (id: string, evidence: AtrEvidence) => {
    downloadingEvidenceId.value = evidence.id
    try {
      const blob = await $fetch<Blob>(`${baseUrl()}/${id}/evidence/${evidence.id}`, { method: 'GET', responseType: 'blob' })
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = evidence.file_name || 'evidence'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      window.URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to download ATR evidence:', error)
      // The blob request reads the error body as a Blob; parse it to get the status and message.
      const parsed = await parseBlobErrorBody(error)
      const detail = getErrorStatus(parsed) === 404
        ? t('actionTakenReport.errors.evidenceMissing')
        : getUserErrorMessage(parsed, t, { fallbackKey: 'actionTakenReport.errors.download' })
      toast.showError(t('actionTakenReport.errors.downloadTitle'), detail)
    } finally {
      downloadingEvidenceId.value = null
    }
  }

  return {
    // list
    items,
    pagination,
    query,
    loading,
    errorMsg,
    fetchReports,
    setFilters,
    setSearch,
    setPage,
    resetFilters,
    initList,
    // summary
    summaryItems,
    summaryLoading,
    summaryError,
    fetchSummary,
    stats,
    // options
    lhaOptions,
    lhaOptionsLoading,
    fetchLhaOptions,
    fetchPicCandidates,
    countForLha,
    // current item / modals
    current,
    detailLoading,
    saving,
    uploading,
    downloadingEvidenceId,
    showDetail,
    showActionPlan,
    showAssignment,
    showReview,
    showCancel,
    fetchReport,
    openDetail,
    openActionPlan,
    openAssignment,
    openReview,
    openCancel,
    closeModal,
    // actions
    assign,
    saveActionPlan,
    submit,
    review,
    cancel,
    uploadEvidence,
    downloadEvidence
  }
})
