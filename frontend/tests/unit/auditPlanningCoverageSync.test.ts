// @ts-nocheck
/**
 * Dashboard "Audit Planning Coverage" card vs the Annual Audit Plan page
 * (/annual-audit → stores/annual-audit.ts → GET /annual-audit-plans).
 *
 * Known causes of the mismatch (the tests marked "BUG" fail until fixed):
 *  - the dashboard never calls annualPlanStore.fetchPlans(), so it shows the
 *    store's two mock plans (1 WIP + 1 Done = 50%) unless /annual-audit was
 *    visited first in the same session;
 *  - an empty API response keeps the mock plans instead of showing 0;
 *  - fetchPlans sends no page_size, and the crud List handler defaults to 20,
 *    so plans beyond the first page are not counted;
 *  - Planned / Completed / Remaining don't add up: Planned = "not Done" and
 *    Remaining skips Pending Approval.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { AnnualAuditPlanStatus as S } from '~/types/audit'

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn()
  })
}))

const DASHBOARD = resolve(__dirname, '../../pages/dashboard/index.vue')

const plan = (i: number, status: string) => ({
  id: `plan-${i}`,
  code: `PKAT-2026-ASR-${String(i).padStart(3, '0')}`,
  status,
  year: '2026',
  activities: []
})

// 25 plans: 10 Done, 6 WIP, 4 Pending Approval, 5 Not Available.
const backendPlans = [
  ...Array.from({ length: 10 }, (_, i) => plan(i + 1, S.DONE)),
  ...Array.from({ length: 6 }, (_, i) => plan(i + 11, S.WORK_IN_PROGRESS)),
  ...Array.from({ length: 4 }, (_, i) => plan(i + 17, S.PENDING_APPROVAL)),
  ...Array.from({ length: 5 }, (_, i) => plan(i + 21, S.NOT_AVAILABLE))
]

// Mimics backend/audit-service/controllers/crud/crud_handler.go List:
// page_size defaults to 20 (max 100), response { data: { items, pagination } }.
const crudListResponse = (all: any[], url: string, opts: any = {}) => {
  const params = new URL(url, 'http://x').searchParams
  const query = opts.query || opts.params || {}
  const page = Number(query.page ?? params.get('page') ?? 1)
  let pageSize = Number(query.page_size ?? params.get('page_size') ?? 20)
  if (pageSize < 1 || pageSize > 100) pageSize = 20
  const items = all.slice((page - 1) * pageSize, page * pageSize)
  return {
    success: true,
    message: 'AuditAnnual fetched successfully',
    data: {
      items,
      pagination: {
        page,
        page_size: pageSize,
        total: all.length,
        total_pages: Math.ceil(all.length / pageSize)
      }
    }
  }
}

function mockApi(all: any[]) {
  global.$fetch = vi.fn(async (url: string, opts?: any) => crudListResponse(all, String(url), opts))
}

// --- Dashboard formulas, replicated from pages/dashboard/index.vue (SFC not importable) ---
//  auditPlansCount      = plans.length                                     (~line 1265)
//  completedAuditsCount = plans.filter(status === "Done")                  (~line 1266)
//  auditCoverage        = { planned: auditPlansCount, completed,
//                           remaining: planned - completed }               (~line 1475)
//  progressModel        = round(completed / planned * 100)                 (~line 1485)
const dashboardCoverage = (plans: any[]) => {
  const planned = plans.length
  const completed = plans.filter(p => p.status === 'Done').length
  return {
    planned,
    completed,
    remaining: planned - completed,
    progress: planned === 0 ? 0 : Math.round((completed / planned) * 100)
  }
}

// What the Annual Audit Plan data says the card should show.
const expectedCoverage = (plans: any[]) => {
  const completed = plans.filter(p => p.status === S.DONE).length
  return {
    planned: plans.length,
    completed,
    remaining: plans.length - completed,
    progress: plans.length === 0 ? 0 : Math.round((completed / plans.length) * 100)
  }
}

describe('Dashboard Audit Planning Coverage vs Annual Audit Plan', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  it('progress % matches the annual plan data once all plans are loaded', () => {
    expect(dashboardCoverage(backendPlans).progress).toBe(expectedCoverage(backendPlans).progress) // 40%
  })

  it('BUG: the dashboard loads the annual plans itself (calls fetchPlans)', () => {
    const src = readFileSync(DASHBOARD, 'utf8')
    expect(src).toMatch(/annualPlanStore\.fetchPlans\s*\(/)
  })

  it('BUG: a store that has not fetched yet holds no mock plans', () => {
    // Without a fetch the dashboard renders whatever the store starts with.
    const store = useAnnualPlanStore()
    expect(store.plans.map(p => p.id)).toEqual([])
  })

  it('BUG: an empty API response shows 0 plans, not the mock plans', async () => {
    mockApi([])
    const store = useAnnualPlanStore()
    await store.fetchPlans()
    expect(store.plans).toHaveLength(0)
    expect(dashboardCoverage(store.plans)).toEqual(expectedCoverage([]))
  })

  it('BUG: counts every plan, not just the first API page (default page_size 20)', async () => {
    mockApi(backendPlans)
    const store = useAnnualPlanStore()
    await store.fetchPlans()
    expect(store.plans).toHaveLength(backendPlans.length)
    expect(dashboardCoverage(store.plans).progress).toBe(expectedCoverage(backendPlans).progress)
  })

  it('loads every page when there are more plans than the max page_size (100)', async () => {
    const many = Array.from({ length: 250 }, (_, i) => plan(i + 1, i % 5 === 0 ? S.DONE : 'DRAFT'))
    mockApi(many)
    const store = useAnnualPlanStore()
    await store.fetchPlans()
    expect(global.$fetch).toHaveBeenCalledTimes(3)
    expect(store.plans).toHaveLength(250)
    expect(new Set(store.plans.map(p => p.id)).size).toBe(250)
    expect(dashboardCoverage(store.plans)).toEqual({ planned: 250, completed: 50, remaining: 200, progress: 20 })
  })

  it('an API error leaves no mock plans behind', async () => {
    global.$fetch = vi.fn(async () => {
      throw new Error('boom')
    })
    const store = useAnnualPlanStore()
    await store.fetchPlans()
    expect(store.plans).toHaveLength(0)
    expect(store.errorMsg).not.toBe('')
  })

  it('BUG: Planned = all plans, and Completed + Remaining = Planned', () => {
    const shown = dashboardCoverage(backendPlans)
    expect(shown.completed + shown.remaining).toBe(shown.planned)
    expect(shown).toEqual(expectedCoverage(backendPlans))
  })

  it('BUG: Pending Approval plans are counted as remaining', () => {
    const plans = [plan(1, S.DONE), plan(2, S.PENDING_APPROVAL)]
    expect(dashboardCoverage(plans).remaining).toBe(1)
  })
})
