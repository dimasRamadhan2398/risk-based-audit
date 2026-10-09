// @ts-nocheck
import { describe, it, expect } from 'vitest'
import en from '~/locales/en/common.json'
import id from '~/locales/id/common.json'
import {
  ATR_SLICE_DOT_CLASS,
  ATR_SLICE_HEX,
  ATR_STATUSES,
  atrActionPlanSchema,
  atrAssignmentSchema,
  atrCancelSchema,
  atrCanSubmit,
  atrFieldErrors,
  atrReviewSchema,
  atrStatusBadgeColor,
  atrStatusDotClass,
  atrStatusI18nKey,
  atrVisibleRows,
  buildAtrParams,
  computeAtrOverdue,
  computeAtrStats,
  defaultAtrQuery,
  isLhaApproved,
  normalizeAtrItem,
  normalizeAtrStatus,
  parseAtrPagination,
  toAtrDueDatePayload
} from '~/utils/actionTakenReport'

const lookup = (dict, key) => key.split('.').reduce((v, k) => (v && typeof v === 'object' ? v[k] : undefined), dict)

describe('ATR status mapping', () => {
  it('normalizes any spelling to the backend value, including PENDING_REVIEW', () => {
    expect(normalizeAtrStatus('IN_PROGRESS')).toBe('IN_PROGRESS')
    expect(normalizeAtrStatus('In Progress')).toBe('IN_PROGRESS')
    expect(normalizeAtrStatus('pending-review')).toBe('PENDING_REVIEW')
    expect(normalizeAtrStatus('Pending Review')).toBe('PENDING_REVIEW')
    expect(normalizeAtrStatus('Completed')).toBe('COMPLETED')
    expect(normalizeAtrStatus('Canceled')).toBe('CANCELLED')
    expect(normalizeAtrStatus('PLANNED')).toBe('PLANNED')
    expect(normalizeAtrStatus(undefined)).toBe('')
  })

  it('has an en and id label for every status and the overdue flag', () => {
    for (const status of ATR_STATUSES) {
      const key = atrStatusI18nKey(status)
      expect(key, status).toBeTruthy()
      expect(typeof lookup(en, key), `en ${key}`).toBe('string')
      expect(typeof lookup(id, key), `id ${key}`).toBe('string')
    }
    expect(lookup(en, 'actionTakenReport.status.pendingReview')).toBe('Pending Review')
    expect(lookup(id, 'actionTakenReport.status.pendingReview')).toBe('Menunggu Reviu')
    expect(atrStatusI18nKey('SOMETHING_ELSE')).toBeNull()
  })

  it('maps statuses to badge colours and dots (PENDING_REVIEW is secondary/violet)', () => {
    expect(atrStatusBadgeColor('PLANNED')).toBe('info')
    expect(atrStatusBadgeColor('IN_PROGRESS')).toBe('warning')
    expect(atrStatusBadgeColor('PENDING_REVIEW')).toBe('secondary')
    expect(atrStatusBadgeColor('COMPLETED')).toBe('success')
    expect(atrStatusBadgeColor('CANCELLED')).toBe('neutral')
    expect(atrStatusBadgeColor('???')).toBe('neutral')
    expect(atrStatusDotClass('PENDING_REVIEW')).toBe(ATR_SLICE_DOT_CLASS.pendingReview)
    expect(atrStatusDotClass('???')).toBe('bg-gray-300')
    // every slice has a page dot and a dashboard colour, all distinct
    const keys = ['completed', 'pendingReview', 'inProgress', 'planned', 'overdue', 'cancelled']
    expect(Object.keys(ATR_SLICE_HEX).sort()).toEqual([...keys].sort())
    expect(new Set(Object.values(ATR_SLICE_HEX)).size).toBe(keys.length)
    expect(new Set(Object.values(ATR_SLICE_DOT_CLASS)).size).toBe(keys.length)
  })
})

describe('ATR overdue (server value first, client fallback)', () => {
  it('fallback uses the Asia/Jakarta date (UTC+7)', () => {
    expect(computeAtrOverdue('2026-09-30T00:00:00Z', 'IN_PROGRESS', new Date('2026-09-30T16:59:00Z')))
      .toEqual({ isOverdue: false, overdueDays: 0 })
    expect(computeAtrOverdue('2026-09-30T00:00:00Z', 'IN_PROGRESS', new Date('2026-09-30T17:00:00Z')))
      .toEqual({ isOverdue: true, overdueDays: 1 })
  })

  it('never marks COMPLETED or CANCELLED, or items without a due date, as overdue', () => {
    const now = new Date('2026-10-01T05:00:00Z')
    expect(computeAtrOverdue('2026-01-01', 'COMPLETED', now).isOverdue).toBe(false)
    expect(computeAtrOverdue('2026-01-01', 'CANCELLED', now).isOverdue).toBe(false)
    expect(computeAtrOverdue(null, 'PLANNED', now).isOverdue).toBe(false)
  })

  it('normalizeAtrItem keeps the server is_overdue/overdue_days and fills safe defaults', () => {
    const now = new Date('2026-10-01T05:00:00Z')
    expect(normalizeAtrItem({ id: 'a', status: 'IN_PROGRESS', due_date: '2026-05-15T00:00:00Z', is_overdue: false, overdue_days: 0 }, now))
      .toMatchObject({ status: 'IN_PROGRESS', is_overdue: false, overdue_days: 0 })
    expect(normalizeAtrItem({ id: 'b', status: 'PLANNED', due_date: '2027-01-01T00:00:00Z', is_overdue: true, overdue_days: 4 }, now))
      .toMatchObject({ is_overdue: true, overdue_days: 4 })
    const bare = normalizeAtrItem({ id: 'c', status: 'pending_review', progress: 140, evidence: [{ id: 'e1', file_name: 'a.pdf' }, { file_name: 'no-id' }] }, now)
    expect(bare).toMatchObject({ status: 'PENDING_REVIEW', progress: 100, pic_user_id: null, due_date: null, is_overdue: false })
    expect(bare.evidence).toEqual([{ id: 'e1', file_name: 'a.pdf', uploaded_at: '', uploaded_by: '' }])
  })

  it('sends due dates as RFC3339 at midnight UTC', () => {
    expect(toAtrDueDatePayload('2026-11-30')).toBe('2026-11-30T00:00:00Z')
    expect(toAtrDueDatePayload('')).toBeNull()
  })
})

describe('ATR stats (dashboard and ATR page summary)', () => {
  it('returns zeros for an empty list', () => {
    const s = computeAtrStats([])
    expect(s.total).toBe(0)
    expect(s.compliance).toBe(0)
    expect(s.breakdown.every(b => b.percent === 0)).toBe(true)
  })

  it('PENDING_REVIEW is open and has its own non-overlapping slice', () => {
    const s = computeAtrStats([
      { status: 'COMPLETED', is_overdue: false },
      { status: 'PENDING_REVIEW', is_overdue: false },
      { status: 'PENDING_REVIEW', is_overdue: true },
      { status: 'IN_PROGRESS', is_overdue: false },
      { status: 'IN_PROGRESS', is_overdue: true },
      { status: 'PLANNED', is_overdue: false },
      { status: 'CANCELLED', is_overdue: false },
      { status: 'Completed', is_overdue: false }
    ])
    expect(Object.fromEntries(s.breakdown.map(b => [b.key, b.count])))
      .toEqual({ completed: 2, pendingReview: 1, inProgress: 1, planned: 1, overdue: 2, cancelled: 1 })
    expect(s.counts).toMatchObject({ open: 5, pendingReview: 2, inProgress: 2, planned: 1, overdue: 2, completed: 2, cancelled: 1 })
    expect(s.breakdown.reduce((n, b) => n + b.count, 0)).toBe(s.total)
    expect(s.compliance).toBeCloseTo(2 / 8, 5)
    expect(s.pendingReviewPercent).toBe(13)
  })
})

describe('ATR list query', () => {
  it('sends only the filters that are set', () => {
    expect(buildAtrParams(defaultAtrQuery('all'), 'u1')).toEqual({ page: 1, page_size: 10 })
    const q = { ...defaultAtrQuery('all'), page: 2, search: '  kas ', status: 'PENDING_REVIEW', lhaId: 'lha-1', picUserId: 'u9', overdue: true }
    expect(buildAtrParams(q, 'u1')).toEqual({
      page: 2, page_size: 10, search: 'kas', status: 'PENDING_REVIEW', audit_result_report_id: 'lha-1', pic_user_id: 'u9', overdue: true
    })
  })

  it('"My actions" always sends the current user as pic_user_id and ignores the PIC filter', () => {
    expect(buildAtrParams({ ...defaultAtrQuery('mine'), picUserId: 'someone-else' }, 'me')).toEqual({ page: 1, page_size: 10, pic_user_id: 'me' })
  })

  it('"My actions" never shows rows of other PICs or unassigned rows', () => {
    const rows = [{ id: '1', pic_user_id: 'me' }, { id: '2', pic_user_id: 'other' }, { id: '3', pic_user_id: null }]
    expect(atrVisibleRows(rows, 'mine', 'me').map(r => r.id)).toEqual(['1'])
    expect(atrVisibleRows(rows, 'mine', '')).toEqual([])
    expect(atrVisibleRows(rows, 'all', 'me')).toHaveLength(3)
  })

  it('pagination comes from the response metadata', () => {
    expect(parseAtrPagination({ page: 3, page_size: 25, total: 61, total_pages: 3 }, { page: 1, pageSize: 10 }))
      .toEqual({ page: 3, page_size: 25, total: 61, total_pages: 3 })
    expect(parseAtrPagination({ total: 21 }, { page: 2, pageSize: 10 })).toEqual({ page: 2, page_size: 10, total: 21, total_pages: 3 })
    expect(parseAtrPagination(undefined, { page: 1, pageSize: 10 })).toEqual({ page: 1, page_size: 10, total: 0, total_pages: 0 })
  })
})

describe('ATR form schemas', () => {
  it('action plan: required text and an integer progress 0-100', () => {
    expect(atrActionPlanSchema.safeParse({ action_plan: ' Fix it ', progress: 40 })).toMatchObject({ success: true, data: { action_plan: 'Fix it', progress: 40 } })
    expect(atrFieldErrors(atrActionPlanSchema.safeParse({ action_plan: '   ', progress: 101 })))
      .toEqual({ action_plan: 'actionTakenReport.validation.actionPlanRequired', progress: 'actionTakenReport.validation.progressRange' })
  })

  it('assignment: PIC and a YYYY-MM-DD due date are required', () => {
    expect(atrAssignmentSchema.safeParse({ pic_user_id: 'u1', due_date: '2026-12-01' }).success).toBe(true)
    expect(atrFieldErrors(atrAssignmentSchema.safeParse({ pic_user_id: '', due_date: '' })))
      .toEqual({ pic_user_id: 'actionTakenReport.validation.picRequired', due_date: 'actionTakenReport.validation.dueDateRequired' })
  })

  it('review: a note is required on reject only', () => {
    expect(atrReviewSchema.safeParse({ decision: 'approve', note: '' }).success).toBe(true)
    expect(atrFieldErrors(atrReviewSchema.safeParse({ decision: 'reject', note: '  ' })))
      .toEqual({ note: 'actionTakenReport.validation.rejectNoteRequired' })
    expect(atrFieldErrors(atrReviewSchema.safeParse({ decision: undefined, note: '' })))
      .toEqual({ decision: 'actionTakenReport.validation.decisionRequired' })
  })

  it('cancel: a reason is required', () => {
    expect(atrFieldErrors(atrCancelSchema.safeParse({ note: '' }))).toEqual({ note: 'actionTakenReport.validation.cancelNoteRequired' })
  })

  it('every validation message exists in en and id', () => {
    for (const key of Object.keys(en.actionTakenReport.validation)) {
      expect(typeof id.actionTakenReport.validation[key], key).toBe('string')
    }
  })

  it('submit needs an open item with an action plan', () => {
    expect(atrCanSubmit({ status: 'IN_PROGRESS', action_plan: 'x' })).toBe(true)
    expect(atrCanSubmit({ status: 'PLANNED', action_plan: '  ' })).toBe(false)
    expect(atrCanSubmit({ status: 'PENDING_REVIEW', action_plan: 'x' })).toBe(false)
  })
})

describe('LHA approval state', () => {
  it('Final (and Approved/Published) count as approved; Draft does not', () => {
    expect(isLhaApproved('Final')).toBe(true)
    expect(isLhaApproved('Approved')).toBe(true)
    expect(isLhaApproved('Draft')).toBe(false)
    expect(isLhaApproved(undefined)).toBe(false)
  })
})
