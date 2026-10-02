// @ts-nocheck
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { AuditStatus } from '~/types/audit'

/**
 * Dashboard "Action Taken Report" donut vs Audit Result Report > Action Taken Report feature.
 * Bug: backend returns UPPER_SNAKE statuses (COMPLETED, IN_PROGRESS, ...), the store compares
 * against AuditStatus ('Completed', 'In Progress', ...), so stats are always 0.
 */

const NOW = '2026-10-01T00:00:00Z'

const item = (id: number, status: string, deadline: string) => ({
  id: String(id),
  auditRef: `ST-00${id}/SKAI/2026`,
  title: `Finding ${id}`,
  department: 'Operations',
  deadline,
  status
})

// Realistic backend payload (production shape): 3 COMPLETED, 1 IN_PROGRESS, 1 PLANNED, 1 CANCELLED
const backendItems = (statuses = {
  c: 'COMPLETED', w: 'IN_PROGRESS', p: 'PLANNED', x: 'CANCELLED'
}) => [
  item(1, statuses.c, '2026-04-15'),
  item(2, statuses.c, '2026-04-20'),
  item(3, statuses.c, '2026-05-10'),
  item(4, statuses.w, '2026-05-15'),
  item(5, statuses.p, '2026-06-01'),
  item(6, statuses.x, '2026-06-15')
]

const frontendCasing = {
  c: AuditStatus.COMPLETED, w: AuditStatus.IN_PROGRESS, p: AuditStatus.PLANNED, x: AuditStatus.CANCELLED
}

async function loadStore(items) {
  global.$fetch = vi.fn().mockResolvedValue({ data: { items } })
  const store = useActionTakenReportStore()
  await store.fetchReports()
  return store
}

// Dashboard formulas replicated from frontend/pages/dashboard/index.vue (not importable from a SFC).
// Updated to the fixed dashboard, which only reads the store's stats (utils/actionTakenReport.ts):
//  openFindingsCount    = stats.counts.open   (not COMPLETED and not CANCELLED)
//  atrCompliancePercent = stats.compliance    (completed / total)
//  atrDonutData         = stats.breakdown: Completed | In Progress (on time) | Planned (on time) | Overdue | Cancelled
const dashboard = (store) => ({
  openFindingsCount: store.stats.counts.open,
  atrCompliancePercent: store.stats.compliance,
  atrDonutData: store.stats.breakdown.map(s => ({ name: s.key, value: s.percent }))
})

describe('ATR dashboard donut vs ATR feature sync', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(NOW)
    vi.spyOn(console, 'error').mockImplementation(() => {})
    setActivePinia(createPinia())
    global.$fetch = vi.fn().mockResolvedValue({ data: { items: [] } })
  })
  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('stats reflect backend data: done 50%, wip 17%', async () => {
    const store = await loadStore(backendItems())
    expect(store.reportList).toHaveLength(6)
    expect(store.stats.donePercent).toBe(50)
    // Changed: wipPercent is the non-overlapping "In Progress (on time)" slice; the only IN_PROGRESS
    // item (deadline 2026-05-15) is overdue, so it lands in Overdue. The raw status count is still 1 (17%).
    expect(store.stats.counts.inProgress).toBe(1)
    expect(store.stats.wipPercent).toBe(0)
  })

  it('stat categories cover every item (PLANNED accounted for, slices sum to ~100)', async () => {
    const store = await loadStore(backendItems())
    const total = store.reportList.length
    const s = store.stats
    // Changed: originally assumed "late" == CANCELLED. Partition is now done + wip(on time) + planned(on time)
    // + overdue + cancelled = 50 + 0 + 0 + 33 + 17 = 100; accept 99..101 for rounding
    const sum = s.donePercent + s.wipPercent + s.plannedPercent + s.overduePercent + s.cancelledPercent
    expect(total).toBe(6)
    expect(sum).toBeGreaterThanOrEqual(99)
    expect(sum).toBeLessThanOrEqual(101)
  })

  it('status filter finds items for each status present in the data', async () => {
    const store = await loadStore(backendItems())
    const expected = { COMPLETED: 3, IN_PROGRESS: 1, PLANNED: 1, CANCELLED: 1 }
    // the ATR page builds filter options from Object.values(AuditStatus)
    const optionValues = Object.values(AuditStatus)
    for (const [raw, count] of Object.entries(expected)) {
      const option = optionValues.find(v => v.toLowerCase().replace(/[\s_]/g, '') === raw.toLowerCase().replace(/[\s_]/g, ''))
      store.selectedStatus = option
      expect(store.filteredReports, `status ${raw} via option ${option}`).toHaveLength(count)
    }
  })

  // Changed: agreed rule is Open Findings = not COMPLETED and not CANCELLED -> IN_PROGRESS + PLANNED = 2
  it('dashboard open findings == not completed and not cancelled (2)', async () => {
    const store = await loadStore(backendItems())
    expect(dashboard(store).openFindingsCount).toBe(2)
  })

  it('dashboard ATR compliance == completed/total (0.5)', async () => {
    const store = await loadStore(backendItems())
    expect(dashboard(store).atrCompliancePercent).toBeCloseTo(3 / 6, 5)
  })

  it('donut Completed slice equals the ATR page "% Done"', async () => {
    const store = await loadStore(backendItems())
    expect(dashboard(store).atrDonutData[0].value).toBe(50)
  })

  // BUSINESS RULE (needs product confirmation): "Overdue" = deadline passed (vs. fixed system time
  // 2026-10-01) AND not finished (not COMPLETED, not CANCELLED). Past-deadline IN_PROGRESS + PLANNED = 2/6 = 33%.
  // CANCELLED items must not be counted as overdue.
  it('Overdue slice = past deadline and not completed/cancelled, not CANCELLED', async () => {
    const items = [...backendItems(), item(7, 'IN_PROGRESS', '2027-01-01')] // future deadline: not overdue
    const store = await loadStore(items)
    // Changed: the donut now has 5 slices, so look Overdue up by key instead of index 2
    expect(dashboard(store).atrDonutData.find(s => s.name === 'overdue').value).toBe(Math.round((2 / 7) * 100)) // 29
  })

  it('control: with frontend-cased statuses, stats are correct (casing mismatch is the root cause)', async () => {
    const store = await loadStore(backendItems(frontendCasing))
    expect(store.stats.donePercent).toBe(50)
    // Changed: same values as the backend-cased fixture under the agreed rules. CANCELLED is no longer
    // "late"; overdue = past-deadline IN_PROGRESS + PLANNED (2/6 = 33%), wip on time = 0, open = 2.
    expect(store.stats.wipPercent).toBe(0)
    expect(store.stats.overduePercent).toBe(33)
    expect(dashboard(store).openFindingsCount).toBe(2)
    store.selectedStatus = AuditStatus.IN_PROGRESS
    expect(store.filteredReports).toHaveLength(1)
  })
})
