// @ts-nocheck
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { AuditStatus } from '~/types/audit'
import {
  computeAtrOverdue,
  computeAtrStats,
  normalizeAtrItem,
  normalizeAtrStatus,
  toBackendAtrStatus
} from '~/utils/actionTakenReport'

describe('ATR status mapping', () => {
  it('normalizes backend UPPER_SNAKE and UI casing to AuditStatus', () => {
    expect(normalizeAtrStatus('IN_PROGRESS')).toBe(AuditStatus.IN_PROGRESS)
    expect(normalizeAtrStatus('In Progress')).toBe(AuditStatus.IN_PROGRESS)
    expect(normalizeAtrStatus('COMPLETED')).toBe(AuditStatus.COMPLETED)
    expect(normalizeAtrStatus('Canceled')).toBe(AuditStatus.CANCELLED)
    expect(normalizeAtrStatus('PLANNED')).toBe(AuditStatus.PLANNED)
  })

  it('converts back to the backend value for create/update payloads', () => {
    expect(toBackendAtrStatus(AuditStatus.IN_PROGRESS)).toBe('IN_PROGRESS')
    expect(toBackendAtrStatus(AuditStatus.PLANNED)).toBe('PLANNED')
    expect(toBackendAtrStatus('COMPLETED')).toBe('COMPLETED')
    expect(toBackendAtrStatus('Cancelled')).toBe('CANCELLED')
  })
})

describe('ATR overdue rule (client fallback)', () => {
  it('uses the Asia/Jakarta date (UTC+7)', () => {
    // 2026-09-30T16:59Z = 23:59 WIB on Sep 30: deadline Sep 30 not yet passed
    expect(computeAtrOverdue('2026-09-30', 'IN_PROGRESS', new Date('2026-09-30T16:59:00Z')))
      .toEqual({ isOverdue: false, daysOverdue: 0 })
    // 2026-09-30T17:00Z = 00:00 WIB on Oct 1: overdue by one day
    expect(computeAtrOverdue('2026-09-30', 'IN_PROGRESS', new Date('2026-09-30T17:00:00Z')))
      .toEqual({ isOverdue: true, daysOverdue: 1 })
  })

  it('takes the date part of ISO datetimes and ignores unparseable deadlines', () => {
    const now = new Date('2026-10-01T05:00:00Z')
    expect(computeAtrOverdue('2026-09-28T23:00:00Z', 'PLANNED', now)).toEqual({ isOverdue: true, daysOverdue: 3 })
    expect(computeAtrOverdue('15 Apr 2026', 'PLANNED', now).isOverdue).toBe(false)
    expect(computeAtrOverdue('', 'PLANNED', now).isOverdue).toBe(false)
  })

  it('never marks COMPLETED or CANCELLED as overdue', () => {
    const now = new Date('2026-10-01T05:00:00Z')
    expect(computeAtrOverdue('2026-01-01', 'COMPLETED', now).isOverdue).toBe(false)
    expect(computeAtrOverdue('2026-01-01', 'CANCELLED', now).isOverdue).toBe(false)
  })

  it('prefers isOverdue/daysOverdue from the API over the client computation', () => {
    const now = new Date('2026-10-01T05:00:00Z')
    expect(normalizeAtrItem({ status: 'IN_PROGRESS', deadline: '2026-05-15', isOverdue: false, daysOverdue: 0 }, now))
      .toMatchObject({ status: AuditStatus.IN_PROGRESS, isOverdue: false, daysOverdue: 0 })
    expect(normalizeAtrItem({ status: 'PLANNED', deadline: '2027-01-01', isOverdue: true, daysOverdue: 4 }, now))
      .toMatchObject({ isOverdue: true, daysOverdue: 4 })
    // fallback when the API fields are absent
    expect(normalizeAtrItem({ status: 'PLANNED', deadline: '2026-09-29' }, now))
      .toMatchObject({ isOverdue: true, daysOverdue: 2 })
  })
})

describe('ATR stats', () => {
  it('returns zeros for an empty list', () => {
    const s = computeAtrStats([])
    expect(s.total).toBe(0)
    expect(s.compliance).toBe(0)
    expect(s.breakdown.every(b => b.percent === 0)).toBe(true)
  })

  it('breakdown is non-overlapping and covers every item', () => {
    const s = computeAtrStats([
      { status: AuditStatus.COMPLETED, isOverdue: false },
      { status: AuditStatus.IN_PROGRESS, isOverdue: false },
      { status: AuditStatus.IN_PROGRESS, isOverdue: true },
      { status: AuditStatus.PLANNED, isOverdue: false },
      { status: AuditStatus.CANCELLED, isOverdue: false }
    ])
    expect(Object.fromEntries(s.breakdown.map(b => [b.key, b.count])))
      .toEqual({ completed: 1, inProgress: 1, planned: 1, overdue: 1, cancelled: 1 })
    expect(s.counts).toMatchObject({ inProgress: 2, open: 3, overdue: 1 })
    expect(s.compliance).toBeCloseTo(0.2, 5)
  })
})

describe('ATR store overdue filter', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime('2026-10-01T00:00:00Z')
    setActivePinia(createPinia())
  })
  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('selectedStatus \'overdue\' shows only overdue items', async () => {
    global.$fetch = vi.fn().mockResolvedValue({
      data: {
        items: [
          { id: '1', auditRef: 'A', title: 'a', deadline: '2026-05-01', status: 'IN_PROGRESS' },
          { id: '2', auditRef: 'B', title: 'b', deadline: '2027-05-01', status: 'IN_PROGRESS' },
          { id: '3', auditRef: 'C', title: 'c', deadline: '2026-05-01', status: 'CANCELLED' }
        ]
      }
    })
    const store = useActionTakenReportStore()
    await store.fetchReports()
    store.selectedStatus = 'overdue'
    expect(store.filteredReports.map(r => r.id)).toEqual(['1'])
  })
})
