import { AuditStatus } from '~/types/audit'

/**
 * Action Taken Report (ATR) helpers shared by the ATR store, the ATR page and the dashboard.
 *
 * Status: the backend stores UPPER_SNAKE (PLANNED, IN_PROGRESS, COMPLETED, CANCELLED); the UI
 * works with the AuditStatus values ('Planned', 'In Progress', 'Completed', 'Cancelled').
 * Everything loaded into the store is normalized to the AuditStatus form, and converted back
 * with `toBackendAtrStatus` before it is sent to the API.
 *
 * Overdue: derived, never stored. An item is overdue when its deadline (date only) is before
 * today in Asia/Jakarta and it is neither COMPLETED nor CANCELLED. The backend returns
 * `isOverdue`/`daysOverdue`; `computeAtrOverdue` is the client-side fallback with the same rule.
 */

export type AtrStatus
  = | AuditStatus.PLANNED
    | AuditStatus.IN_PROGRESS
    | AuditStatus.COMPLETED
    | AuditStatus.CANCELLED

export type AtrBackendStatus = 'PLANNED' | 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED'

export const ATR_STATUSES: AtrStatus[] = [
  AuditStatus.PLANNED,
  AuditStatus.IN_PROGRESS,
  AuditStatus.COMPLETED,
  AuditStatus.CANCELLED
]

/** Value used by the ATR status filter to select overdue items (not a real status). */
export const ATR_OVERDUE_FILTER = 'overdue'

const statusKey = (value: string) => value.toLowerCase().replace(/[\s_-]/g, '')

const STATUS_BY_KEY: Record<string, AtrStatus> = {
  planned: AuditStatus.PLANNED,
  inprogress: AuditStatus.IN_PROGRESS,
  completed: AuditStatus.COMPLETED,
  cancelled: AuditStatus.CANCELLED,
  canceled: AuditStatus.CANCELLED
}

const BACKEND_STATUS: Record<AtrStatus, AtrBackendStatus> = {
  [AuditStatus.PLANNED]: 'PLANNED',
  [AuditStatus.IN_PROGRESS]: 'IN_PROGRESS',
  [AuditStatus.COMPLETED]: 'COMPLETED',
  [AuditStatus.CANCELLED]: 'CANCELLED'
}

/** Maps 'IN_PROGRESS', 'In Progress', 'in-progress', ... to AuditStatus. Unknown values are returned trimmed. */
export const normalizeAtrStatus = (raw: unknown): AtrStatus | string => {
  if (typeof raw !== 'string') return ''
  return STATUS_BY_KEY[statusKey(raw)] ?? raw.trim()
}

/** Converts any accepted status spelling to the backend's UPPER_SNAKE value (for create/update payloads). */
export const toBackendAtrStatus = (status: unknown): AtrBackendStatus | string => {
  const normalized = normalizeAtrStatus(status)
  return BACKEND_STATUS[normalized as AtrStatus] ?? normalized
}

const STATUS_I18N_KEY: Record<AtrStatus, string> = {
  [AuditStatus.PLANNED]: 'planned',
  [AuditStatus.IN_PROGRESS]: 'inProgress',
  [AuditStatus.COMPLETED]: 'completed',
  [AuditStatus.CANCELLED]: 'cancelled'
}

/** i18n key ('actionTakenReport.status.*') for a status, or null for unknown values (show them raw). */
export const atrStatusI18nKey = (status: unknown): string | null => {
  const key = STATUS_I18N_KEY[normalizeAtrStatus(status) as AtrStatus]
  return key ? `actionTakenReport.status.${key}` : null
}

export const isAtrClosed = (status: unknown) => {
  const normalized = normalizeAtrStatus(status)
  return normalized === AuditStatus.COMPLETED || normalized === AuditStatus.CANCELLED
}

// Asia/Jakarta is UTC+7 all year (no DST), so a fixed offset is exact and needs no ICU data.
const JAKARTA_OFFSET_MS = 7 * 60 * 60 * 1000
const DAY_MS = 24 * 60 * 60 * 1000

/** Today's date in Asia/Jakarta as 'YYYY-MM-DD'. */
export const jakartaToday = (now: Date = new Date()) =>
  new Date(now.getTime() + JAKARTA_OFFSET_MS).toISOString().slice(0, 10)

/** Date part of a 'YYYY-MM-DD' or ISO datetime deadline; null when the deadline is not in that form. */
export const atrDeadlineDate = (deadline: unknown): string | null => {
  if (typeof deadline !== 'string') return null
  const match = /^(\d{4}-\d{2}-\d{2})/.exec(deadline.trim())
  return match?.[1] ?? null
}

/** Client-side mirror of the backend `isOverdue`/`daysOverdue` rule. */
export const computeAtrOverdue = (
  deadline: unknown,
  status: unknown,
  now: Date = new Date()
): { isOverdue: boolean, daysOverdue: number } => {
  const due = atrDeadlineDate(deadline)
  if (!due || isAtrClosed(status)) return { isOverdue: false, daysOverdue: 0 }
  const today = jakartaToday(now)
  if (due >= today) return { isOverdue: false, daysOverdue: 0 }
  const days = Math.round((Date.parse(`${today}T00:00:00Z`) - Date.parse(`${due}T00:00:00Z`)) / DAY_MS)
  return { isOverdue: true, daysOverdue: days }
}

/** Normalizes one ATR item from the API: canonical status, assignment letter alias, overdue flags. */
// eslint-disable-next-line @typescript-eslint/no-explicit-any -- raw API rows (snake_case + camelCase mix)
export const normalizeAtrItem = <T extends Record<string, any>>(item: T, now: Date = new Date()) => {
  const status = normalizeAtrStatus(item.status)
  const computed = computeAtrOverdue(item.deadline, status, now)
  const apiHasOverdue = typeof item.isOverdue === 'boolean'
  const isOverdue = apiHasOverdue ? item.isOverdue as boolean : computed.isOverdue
  let daysOverdue = 0
  if (isOverdue) {
    daysOverdue = typeof item.daysOverdue === 'number' ? item.daysOverdue : computed.daysOverdue
  }
  return {
    ...item,
    assignmentLetter: item.assignment_letter || item.assignmentLetter,
    status,
    isOverdue,
    daysOverdue
  }
}

export type AtrSliceKey = 'completed' | 'inProgress' | 'planned' | 'overdue' | 'cancelled'

export interface AtrStats {
  total: number
  counts: {
    completed: number
    inProgress: number
    planned: number
    cancelled: number
    /** Overlay: not completed/cancelled and past the deadline (any status). */
    overdue: number
    /** Not completed and not cancelled. */
    open: number
    inProgressOnTime: number
    plannedOnTime: number
  }
  /** Non-overlapping donut slices over all items (Completed, In Progress, Planned, Overdue, Cancelled). */
  breakdown: { key: AtrSliceKey, count: number, percent: number }[]
  donePercent: number
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
 *   Completed | In Progress (on time) | Planned (on time) | Overdue (open + past deadline) | Cancelled
 */
export const computeAtrStats = (
  items: { status?: unknown, isOverdue?: boolean }[]
): AtrStats => {
  const counts = {
    completed: 0,
    inProgress: 0,
    planned: 0,
    cancelled: 0,
    overdue: 0,
    open: 0,
    inProgressOnTime: 0,
    plannedOnTime: 0
  }

  for (const item of items) {
    const status = normalizeAtrStatus(item.status)
    if (status === AuditStatus.COMPLETED) {
      counts.completed++
      continue
    }
    if (status === AuditStatus.CANCELLED) {
      counts.cancelled++
      continue
    }
    counts.open++
    if (status === AuditStatus.IN_PROGRESS) counts.inProgress++
    if (status === AuditStatus.PLANNED) counts.planned++
    if (item.isOverdue) {
      counts.overdue++
    } else if (status === AuditStatus.IN_PROGRESS) {
      counts.inProgressOnTime++
    } else if (status === AuditStatus.PLANNED) {
      counts.plannedOnTime++
    }
  }

  const total = items.length
  const pct = (n: number) => (total === 0 ? 0 : Math.round((n / total) * 100))

  const breakdown: AtrStats['breakdown'] = [
    { key: 'completed', count: counts.completed, percent: pct(counts.completed) },
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
    wipPercent: pct(counts.inProgressOnTime),
    plannedPercent: pct(counts.plannedOnTime),
    overduePercent: pct(counts.overdue),
    cancelledPercent: pct(counts.cancelled),
    compliance: total === 0 ? 0 : counts.completed / total
  }
}
