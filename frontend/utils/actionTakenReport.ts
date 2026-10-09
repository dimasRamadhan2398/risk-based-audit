import { z } from 'zod'
import type { ActionTakenReport, AtrEvidence } from '~/types/audit'

/**
 * Action Taken Report (ATR) helpers shared by the ATR store, the ATR page, the LHA page and the dashboard.
 *
 * An ATR is the follow-up of one Audit Result Report (LHA) finding. The backend creates one per finding
 * when the LHA is approved; there is no manual create.
 *
 * Status (backend UPPER_SNAKE, used as-is everywhere in the UI and in the API filters):
 *   PLANNED -> IN_PROGRESS (first action plan save) -> PENDING_REVIEW (PIC submits) -> COMPLETED (approved)
 *   rejecting a review goes back to IN_PROGRESS; CANCELLED can be set by manager/CAE/admin.
 *
 * Overdue: computed by the server (`is_overdue`, `overdue_days`). `computeAtrOverdue` is only a fallback
 * for rows without those fields: due date (date part) before today in Asia/Jakarta, and still open.
 */

export type AtrStatus = 'PLANNED' | 'IN_PROGRESS' | 'PENDING_REVIEW' | 'COMPLETED' | 'CANCELLED'

export const ATR_STATUSES: AtrStatus[] = ['PLANNED', 'IN_PROGRESS', 'PENDING_REVIEW', 'COMPLETED', 'CANCELLED']

const statusKey = (value: string) => value.toLowerCase().replace(/[\s_-]/g, '')

const STATUS_BY_KEY: Record<string, AtrStatus> = {
  planned: 'PLANNED',
  inprogress: 'IN_PROGRESS',
  pendingreview: 'PENDING_REVIEW',
  completed: 'COMPLETED',
  cancelled: 'CANCELLED',
  canceled: 'CANCELLED'
}

/** Maps 'IN_PROGRESS', 'In Progress', 'pending-review', ... to the backend value. Unknown values are returned trimmed. */
export const normalizeAtrStatus = (raw: unknown): AtrStatus | string => {
  if (typeof raw !== 'string') return ''
  return STATUS_BY_KEY[statusKey(raw)] ?? raw.trim()
}

const STATUS_I18N_KEY: Record<AtrStatus, string> = {
  PLANNED: 'planned',
  IN_PROGRESS: 'inProgress',
  PENDING_REVIEW: 'pendingReview',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled'
}

/** i18n key ('actionTakenReport.status.*') for a status, or null for unknown values (show them raw). */
export const atrStatusI18nKey = (status: unknown): string | null => {
  const key = STATUS_I18N_KEY[normalizeAtrStatus(status) as AtrStatus]
  return key ? `actionTakenReport.status.${key}` : null
}

export type AtrBadgeColor = 'info' | 'warning' | 'secondary' | 'success' | 'neutral'

const STATUS_BADGE_COLOR: Record<AtrStatus, AtrBadgeColor> = {
  PLANNED: 'info',
  IN_PROGRESS: 'warning',
  PENDING_REVIEW: 'secondary',
  COMPLETED: 'success',
  CANCELLED: 'neutral'
}

/** Nuxt UI badge colour for a status (neutral for unknown values). */
export const atrStatusBadgeColor = (status: unknown): AtrBadgeColor =>
  STATUS_BADGE_COLOR[normalizeAtrStatus(status) as AtrStatus] ?? 'neutral'

export const isAtrClosed = (status: unknown) => {
  const normalized = normalizeAtrStatus(status)
  return normalized === 'COMPLETED' || normalized === 'CANCELLED'
}

// ---------------------------------------------------------------------------
// Overdue (client fallback only)
// ---------------------------------------------------------------------------

// Asia/Jakarta is UTC+7 all year (no DST), so a fixed offset is exact and needs no ICU data.
const JAKARTA_OFFSET_MS = 7 * 60 * 60 * 1000
const DAY_MS = 24 * 60 * 60 * 1000

/** Today's date in Asia/Jakarta as 'YYYY-MM-DD'. */
export const jakartaToday = (now: Date = new Date()) =>
  new Date(now.getTime() + JAKARTA_OFFSET_MS).toISOString().slice(0, 10)

/** Date part of a 'YYYY-MM-DD' or RFC3339 value; null when the value is not in that form. */
export const atrDateOnly = (value: unknown): string | null => {
  if (typeof value !== 'string') return null
  const match = /^(\d{4}-\d{2}-\d{2})/.exec(value.trim())
  return match?.[1] ?? null
}

/** Client-side mirror of the server's is_overdue/overdue_days rule. */
export const computeAtrOverdue = (
  dueDate: unknown,
  status: unknown,
  now: Date = new Date()
): { isOverdue: boolean, overdueDays: number } => {
  const due = atrDateOnly(dueDate)
  if (!due || isAtrClosed(status)) return { isOverdue: false, overdueDays: 0 }
  const today = jakartaToday(now)
  if (due >= today) return { isOverdue: false, overdueDays: 0 }
  const days = Math.round((Date.parse(`${today}T00:00:00Z`) - Date.parse(`${due}T00:00:00Z`)) / DAY_MS)
  return { isOverdue: true, overdueDays: days }
}

const intlLocale = (locale: string) => (locale === 'id' ? 'id-ID' : 'en-GB')

/** Due date as '15 Apr 2026' (calendar date only, no time-zone shift); '-' when empty. */
export const formatAtrDate = (value: unknown, locale = 'en'): string => {
  const day = atrDateOnly(value)
  if (!day) return '-'
  return new Intl.DateTimeFormat(intlLocale(locale), { day: '2-digit', month: 'short', year: 'numeric', timeZone: 'UTC' })
    .format(new Date(`${day}T00:00:00Z`))
}

/** Timestamp as '15 Apr 2026, 14:05' in Asia/Jakarta; '-' when empty or unparseable. */
export const formatAtrDateTime = (value: unknown, locale = 'en'): string => {
  if (typeof value !== 'string' || !value.trim()) return '-'
  const ms = Date.parse(value)
  if (Number.isNaN(ms)) return value
  return new Intl.DateTimeFormat(intlLocale(locale), {
    day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit', hour12: false, timeZone: 'Asia/Jakarta'
  }).format(new Date(ms))
}

/** 'YYYY-MM-DD' from the date picker -> RFC3339 for the API (midnight UTC keeps the same calendar date in UTC and WIB). */
export const toAtrDueDatePayload = (date: string | null | undefined): string | null => {
  const day = atrDateOnly(date)
  return day ? `${day}T00:00:00Z` : null
}

// ---------------------------------------------------------------------------
// Normalisation of API rows
// ---------------------------------------------------------------------------

type RawRow = Record<string, unknown>

const str = (value: unknown): string => (typeof value === 'string' ? value : value == null ? '' : String(value))
const strOrNull = (value: unknown): string | null => (typeof value === 'string' && value.trim() ? value : null)

const clampProgress = (value: unknown): number => {
  const n = Number(value)
  if (!Number.isFinite(n)) return 0
  return Math.min(100, Math.max(0, Math.round(n)))
}

const normalizeEvidence = (raw: unknown): AtrEvidence[] => {
  if (!Array.isArray(raw)) return []
  return raw
    .filter((e): e is RawRow => !!e && typeof e === 'object')
    .map(e => ({
      id: str(e.id),
      file_name: str(e.file_name),
      uploaded_at: str(e.uploaded_at),
      uploaded_by: str(e.uploaded_by)
    }))
    .filter(e => e.id)
}

/** One ATR item from the API, with a canonical status, a 0-100 progress and server overdue flags (client fallback when absent). */
export const normalizeAtrItem = (item: RawRow, now: Date = new Date()): ActionTakenReport => {
  const status = normalizeAtrStatus(item.status)
  const dueDate = strOrNull(item.due_date)
  const fallback = computeAtrOverdue(dueDate, status, now)
  const isOverdue = typeof item.is_overdue === 'boolean' ? item.is_overdue : fallback.isOverdue
  let overdueDays = 0
  if (isOverdue) overdueDays = typeof item.overdue_days === 'number' ? item.overdue_days : fallback.overdueDays
  return {
    id: str(item.id),
    audit_result_report_id: str(item.audit_result_report_id),
    report_number: str(item.report_number),
    report_title: str(item.report_title),
    finding_id: str(item.finding_id),
    finding_title: str(item.finding_title),
    finding_category: str(item.finding_category),
    recommendation: str(item.recommendation),
    action_plan: str(item.action_plan),
    pic_user_id: strOrNull(item.pic_user_id),
    pic_name: str(item.pic_name),
    due_date: dueDate,
    progress: clampProgress(item.progress),
    status,
    is_overdue: isOverdue,
    overdue_days: overdueDays,
    evidence: normalizeEvidence(item.evidence),
    review_note: str(item.review_note),
    reviewed_by: str(item.reviewed_by),
    reviewed_at: strOrNull(item.reviewed_at),
    created_at: str(item.created_at),
    updated_at: str(item.updated_at)
  }
}

// ---------------------------------------------------------------------------
// List query (server-side paging and filters)
// ---------------------------------------------------------------------------

export type AtrScope = 'mine' | 'all'

export interface AtrQuery {
  page: number
  pageSize: number
  /** Applied (debounced) search term. */
  search: string
  status: string
  lhaId: string
  /** PIC filter for users who can view all; ignored in the "mine" scope. */
  picUserId: string
  overdue: boolean
  scope: AtrScope
}

export interface AtrPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export const ATR_PAGE_SIZES = [10, 25, 50]
export const ATR_DEFAULT_PAGE_SIZE = 10
export const ATR_MAX_PAGE_SIZE = 100

export const defaultAtrQuery = (scope: AtrScope = 'all'): AtrQuery => ({
  page: 1,
  pageSize: ATR_DEFAULT_PAGE_SIZE,
  search: '',
  status: '',
  lhaId: '',
  picUserId: '',
  overdue: false,
  scope
})

export const emptyAtrPagination = (pageSize = ATR_DEFAULT_PAGE_SIZE): AtrPagination =>
  ({ page: 1, page_size: pageSize, total: 0, total_pages: 0 })

/**
 * Query string for GET /action-taken-reports; empty filters are left out.
 * The "mine" scope always sends the current user's id as pic_user_id.
 */
export const buildAtrParams = (q: AtrQuery, currentUserId: string | null | undefined): Record<string, string | number | boolean> => {
  const params: Record<string, string | number | boolean> = { page: q.page, page_size: q.pageSize }
  const search = q.search.trim()
  if (search) params.search = search
  if (q.status) params.status = q.status
  if (q.lhaId) params.audit_result_report_id = q.lhaId
  if (q.scope === 'mine') {
    if (currentUserId) params.pic_user_id = currentUserId
  } else if (q.picUserId) {
    params.pic_user_id = q.picUserId
  }
  if (q.overdue) params.overdue = true
  return params
}

/**
 * Rows shown in the table. In "My actions" the server already returns only the user's items; as a guard,
 * rows assigned to anyone else (or to nobody) are never shown there.
 */
export const atrVisibleRows = <T extends { pic_user_id: string | null }>(
  rows: T[],
  scope: AtrScope,
  currentUserId: string | null | undefined
): T[] => {
  if (scope !== 'mine') return rows
  if (!currentUserId) return []
  return rows.filter(row => !!row.pic_user_id && String(row.pic_user_id) === String(currentUserId))
}

const toInt = (value: unknown, fallback: number): number => {
  const n = Number(value)
  return Number.isFinite(n) ? Math.trunc(n) : fallback
}

/** Pagination from `data.pagination`; total_pages is derived from total when the server leaves it out. */
export const parseAtrPagination = (
  raw: Partial<Record<keyof AtrPagination, unknown>> | null | undefined,
  requested: { page: number, pageSize: number }
): AtrPagination => {
  const pageSize = Math.max(1, toInt(raw?.page_size, requested.pageSize))
  const total = Math.max(0, toInt(raw?.total, 0))
  const totalPages = raw?.total_pages !== undefined ? Math.max(0, toInt(raw.total_pages, 0)) : Math.ceil(total / pageSize)
  return { page: Math.max(1, toInt(raw?.page, requested.page)), page_size: pageSize, total, total_pages: totalPages }
}

// ---------------------------------------------------------------------------
// What can be done with an item, by status (roles are checked in useRbac)
// ---------------------------------------------------------------------------

const isOpenForWork = (status: unknown) => {
  const s = normalizeAtrStatus(status)
  return s === 'PLANNED' || s === 'IN_PROGRESS'
}

/** PIC / due date can be (re)assigned while the work has not been submitted or closed. */
export const atrCanAssignStatus = (status: unknown) => isOpenForWork(status)
/** The action plan and evidence can be edited while PLANNED or IN_PROGRESS (also after a rejected review). */
export const atrCanEditPlanStatus = (status: unknown) => isOpenForWork(status)
/** Submitting needs an open item and an action plan. */
export const atrCanSubmit = (item: Pick<ActionTakenReport, 'status' | 'action_plan'>) =>
  isOpenForWork(item.status) && !!item.action_plan.trim()
export const atrCanReviewStatus = (status: unknown) => normalizeAtrStatus(status) === 'PENDING_REVIEW'
export const atrCanCancelStatus = (status: unknown) => !!normalizeAtrStatus(status) && !isAtrClosed(status)

// ---------------------------------------------------------------------------
// Form schemas (messages are i18n keys)
// ---------------------------------------------------------------------------

export const ATR_ACTION_PLAN_MAX = 5000
export const ATR_NOTE_MAX = 2000

export const atrActionPlanSchema = z.object({
  action_plan: z.string().trim()
    .min(1, 'actionTakenReport.validation.actionPlanRequired')
    .max(ATR_ACTION_PLAN_MAX, 'actionTakenReport.validation.actionPlanTooLong'),
  progress: z.number({ message: 'actionTakenReport.validation.progressRange' }).int()
    .min(0, 'actionTakenReport.validation.progressRange')
    .max(100, 'actionTakenReport.validation.progressRange')
})

export const atrAssignmentSchema = z.object({
  pic_user_id: z.string().trim().min(1, 'actionTakenReport.validation.picRequired'),
  due_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'actionTakenReport.validation.dueDateRequired')
})

export const atrReviewSchema = z.object({
  decision: z.enum(['approve', 'reject'], { message: 'actionTakenReport.validation.decisionRequired' }),
  note: z.string().trim().max(ATR_NOTE_MAX, 'actionTakenReport.validation.noteTooLong')
}).refine(v => v.decision !== 'reject' || v.note.length > 0, {
  message: 'actionTakenReport.validation.rejectNoteRequired',
  path: ['note']
})

export const atrCancelSchema = z.object({
  note: z.string().trim()
    .min(1, 'actionTakenReport.validation.cancelNoteRequired')
    .max(ATR_NOTE_MAX, 'actionTakenReport.validation.noteTooLong')
})

/** First error message (an i18n key) per field from a zod result; {} when valid. */
export const atrFieldErrors = (result: { success: boolean, error?: z.ZodError }): Record<string, string> => {
  if (result.success || !result.error) return {}
  const out: Record<string, string> = {}
  for (const issue of result.error.issues) {
    const field = String(issue.path[0] ?? '')
    if (field && !out[field]) out[field] = issue.message
  }
  return out
}

// ---------------------------------------------------------------------------
// Finding category (copied from the LHA finding; same colours as the LHA page)
// ---------------------------------------------------------------------------

const FINDING_CATEGORY: Record<string, { key: string, color: 'error' | 'warning' | 'info' | 'success' }> = {
  verysignificant: { key: 'verySignificant', color: 'error' },
  significant: { key: 'significant', color: 'warning' },
  quitesignificant: { key: 'quiteSignificant', color: 'info' },
  moderatelysignificant: { key: 'quiteSignificant', color: 'info' },
  notsignificant: { key: 'notSignificant', color: 'success' },
  insignificant: { key: 'notSignificant', color: 'success' }
}

const findingCategory = (raw: unknown) => (typeof raw === 'string' ? FINDING_CATEGORY[statusKey(raw)] : undefined)

export const atrFindingCategoryColor = (raw: unknown): 'error' | 'warning' | 'info' | 'success' | 'neutral' =>
  findingCategory(raw)?.color ?? 'neutral'

/** i18n key ('actionTakenReport.findingCategory.*') or null for unknown values (show them raw). */
export const atrFindingCategoryI18nKey = (raw: unknown): string | null => {
  const entry = findingCategory(raw)
  return entry ? `actionTakenReport.findingCategory.${entry.key}` : null
}

// ---------------------------------------------------------------------------
// LHA helpers
// ---------------------------------------------------------------------------

/** An LHA counts as approved once it is Final (the LHA form's approval state); 'Approved'/'Published' are accepted too. */
export const isLhaApproved = (status: unknown) => {
  const s = typeof status === 'string' ? status.trim().toLowerCase() : ''
  return s === 'final' || s === 'approved' || s === 'published'
}

// ---------------------------------------------------------------------------
// Stats (ATR page summary and dashboard)
// ---------------------------------------------------------------------------

export type AtrSliceKey = 'completed' | 'pendingReview' | 'inProgress' | 'planned' | 'overdue' | 'cancelled'

export interface AtrStats {
  total: number
  counts: {
    completed: number
    pendingReview: number
    inProgress: number
    planned: number
    cancelled: number
    /** Overlay: open (not completed/cancelled) and past the due date, any status. */
    overdue: number
    /** Not completed and not cancelled (PLANNED, IN_PROGRESS, PENDING_REVIEW). */
    open: number
    pendingReviewOnTime: number
    inProgressOnTime: number
    plannedOnTime: number
  }
  /** Non-overlapping slices over all items: Completed, Pending Review, In Progress, Planned, Overdue, Cancelled. */
  breakdown: { key: AtrSliceKey, count: number, percent: number }[]
  donePercent: number
  /** Pending review and not overdue. */
  pendingReviewPercent: number
  /** In Progress and not overdue. */
  wipPercent: number
  /** Planned and not overdue. */
  plannedPercent: number
  overduePercent: number
  cancelledPercent: number
  /** completed / total, 0..1 (same basis as donePercent). */
  compliance: number
}

/**
 * Single source of truth for ATR numbers.
 * Breakdown (sums to 100% up to rounding, denominator = all items):
 *   Completed | Pending Review (on time) | In Progress (on time) | Planned (on time) | Overdue (open + past due) | Cancelled
 */
export const computeAtrStats = (
  items: { status?: unknown, is_overdue?: boolean }[]
): AtrStats => {
  const counts = {
    completed: 0,
    pendingReview: 0,
    inProgress: 0,
    planned: 0,
    cancelled: 0,
    overdue: 0,
    open: 0,
    pendingReviewOnTime: 0,
    inProgressOnTime: 0,
    plannedOnTime: 0
  }

  for (const item of items) {
    const status = normalizeAtrStatus(item.status)
    if (status === 'COMPLETED') {
      counts.completed++
      continue
    }
    if (status === 'CANCELLED') {
      counts.cancelled++
      continue
    }
    counts.open++
    if (status === 'PENDING_REVIEW') counts.pendingReview++
    if (status === 'IN_PROGRESS') counts.inProgress++
    if (status === 'PLANNED') counts.planned++
    if (item.is_overdue) {
      counts.overdue++
    } else if (status === 'PENDING_REVIEW') {
      counts.pendingReviewOnTime++
    } else if (status === 'IN_PROGRESS') {
      counts.inProgressOnTime++
    } else if (status === 'PLANNED') {
      counts.plannedOnTime++
    }
  }

  const total = items.length
  const pct = (n: number) => (total === 0 ? 0 : Math.round((n / total) * 100))

  const breakdown: AtrStats['breakdown'] = [
    { key: 'completed', count: counts.completed, percent: pct(counts.completed) },
    { key: 'pendingReview', count: counts.pendingReviewOnTime, percent: pct(counts.pendingReviewOnTime) },
    { key: 'inProgress', count: counts.inProgressOnTime, percent: pct(counts.inProgressOnTime) },
    { key: 'planned', count: counts.plannedOnTime, percent: pct(counts.plannedOnTime) },
    { key: 'overdue', count: counts.overdue, percent: pct(counts.overdue) },
    { key: 'cancelled', count: counts.cancelled, percent: pct(counts.cancelled) }
  ]

  return {
    total,
    counts,
    breakdown,
    donePercent: pct(counts.completed),
    pendingReviewPercent: pct(counts.pendingReviewOnTime),
    wipPercent: pct(counts.inProgressOnTime),
    plannedPercent: pct(counts.plannedOnTime),
    overduePercent: pct(counts.overdue),
    cancelledPercent: pct(counts.cancelled),
    compliance: total === 0 ? 0 : counts.completed / total
  }
}

/** Tailwind dot colours for the ATR page summary and status column (same family as the existing ATR palette). */
export const ATR_SLICE_DOT_CLASS: Record<AtrSliceKey, string> = {
  completed: 'bg-emerald-500',
  pendingReview: 'bg-violet-500',
  inProgress: 'bg-amber-400',
  planned: 'bg-sky-400',
  overdue: 'bg-rose-500',
  cancelled: 'bg-gray-400'
}

/** Hex colours for the dashboard donut (brand purple/slate palette; Pending Review sits between Completed and Planned). */
export const ATR_SLICE_HEX: Record<AtrSliceKey, string> = {
  completed: '#4d00ff',
  pendingReview: '#8b5cf6',
  inProgress: '#94a3b8',
  planned: '#c4b5fd',
  overdue: '#ff5c02',
  cancelled: '#e2e8f0'
}

const SLICE_BY_STATUS: Record<AtrStatus, AtrSliceKey> = {
  PLANNED: 'planned',
  IN_PROGRESS: 'inProgress',
  PENDING_REVIEW: 'pendingReview',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled'
}

/** Dot colour for a status (gray for unknown values). */
export const atrStatusDotClass = (status: unknown): string => {
  const slice = SLICE_BY_STATUS[normalizeAtrStatus(status) as AtrStatus]
  return slice ? ATR_SLICE_DOT_CLASS[slice] : 'bg-gray-300'
}
