/**
 * KPI Detailed Breakdown (/kpi-performance): the table is paged, filtered and searched by
 * GET /performance/kpi-breakdown, and the footer/pagination come only from the response metadata.
 * The stores this page uses hold no sample data.
 */
import { describe, it, expect, beforeEach, afterEach, vi, type Mock } from 'vitest'
import { readFileSync } from 'fs'
import { resolve } from 'path'
import { setActivePinia, createPinia } from 'pinia'
import en from '~/locales/en/common.json'
import id from '~/locales/id/common.json'
import { usePerformanceStore } from '~/stores/performance'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { kpiValueLabel, type TranslateFn } from '~/utils/kpiPerformanceLabels'
import {
  formatKpiGap,
  formatKpiValue,
  kpiBreakdownRange,
  kpiBreakdownRangeText,
  kpiCategoryText,
  kpiGapClass,
  kpiStatusColor
} from '~/utils/kpiBreakdown'

const showInfo = vi.fn()
const showError = vi.fn()
vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({ showSuccess: vi.fn(), showError, showWarning: vi.fn(), showInfo })
}))

// Same lookup semantics as composables/useI18n.ts: missing key returns the key itself.
const makeT = (dict: object): TranslateFn => (key, params) => {
  let value: unknown = dict
  for (const k of key.split('.')) {
    if (value && typeof value === 'object' && k in value) value = (value as Record<string, unknown>)[k]
    else return key
  }
  if (typeof value !== 'string') return key
  return params ? value.replace(/\{(\w+)\}/g, (m, p) => params[p]?.toString() || m) : value
}
const tEn = makeT(en)
const tId = makeT(id)

const flush = () => new Promise(r => setTimeout(r, 0))

// The stores call the Nuxt auto-imported $fetch; tests replace it on globalThis.
type FetchOpts = { params?: Record<string, string | number> }
type FetchFn = (url: string, opts?: FetchOpts) => Promise<unknown>
const g = globalThis as unknown as { $fetch: Mock<FetchFn>, [name: string]: unknown }
const fake = (impl: FetchFn): Mock<FetchFn> => vi.fn(impl)
const failWith = (message: string) => fake(async () => {
  throw new Error(message)
})
const callParams = (index: number): Record<string, string | number> => g.$fetch.mock.calls.at(index)?.[1]?.params ?? {}
/** Element of a list that the test expects to exist. */
const nth = <T>(list: T[], index: number): T => {
  const value = list[index]
  if (value === undefined) throw new Error(`no element ${index}`)
  return value
}

const item = (n: number, extra: Record<string, unknown> = {}) => ({
  id: `id-${n}`,
  source: 'strategic_plan',
  metric: `KPI ${n}`,
  category: '',
  period: 'Q1',
  unit: '%',
  target: 90,
  actual: 92,
  gap: 2,
  gapIsPositive: true,
  achievementRate: 102.22,
  status: 'Exceeded',
  ...extra
})

/** A fake backend for a data set of `total` rows, honouring page / page_size like the real one. */
const pagedBackend = (total: number) => fake(async (_url: string, opts: FetchOpts = {}) => {
  const page = Number(opts.params?.page ?? 1)
  const pageSize = Number(opts.params?.page_size ?? 10)
  const start = (page - 1) * pageSize
  const count = Math.max(0, Math.min(pageSize, total - start))
  return {
    success: true,
    message: 'ok',
    data: {
      items: Array.from({ length: count }, (_, i) => item(start + i + 1)),
      pagination: { page, page_size: pageSize, total, total_pages: total ? Math.ceil(total / pageSize) : 0 }
    }
  }
})

const lastParams = () => callParams(-1)

describe('KPI breakdown store: query sent to /performance/kpi-breakdown', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  it('sends year, page and page_size, and the filters once they are set', async () => {
    g.$fetch = pagedBackend(30)
    const store = usePerformanceStore()

    await store.loadKpiBreakdown(2026)
    const [url] = g.$fetch.mock.calls[0] ?? []
    expect(url).toMatch(/\/performance\/kpi-breakdown$/)
    expect(callParams(0)).toEqual({ year: 2026, page: 1, page_size: 10 })

    await store.setKpiBreakdownFilters({ status: 'On Track', period: 'Q1-2026', category: 'Financial' })
    expect(lastParams()).toEqual({ year: 2026, page: 1, page_size: 10, status: 'On Track', period: 'Q1-2026', category: 'Financial' })
  })

  it('debounces search and sends only the last term', async () => {
    vi.useFakeTimers()
    g.$fetch = pagedBackend(30)
    const store = usePerformanceStore()
    await store.loadKpiBreakdown(2026)
    expect(g.$fetch).toHaveBeenCalledTimes(1)

    store.setKpiBreakdownSearch('aud')
    store.setKpiBreakdownSearch('audit ')
    await vi.advanceTimersByTimeAsync(299)
    expect(g.$fetch).toHaveBeenCalledTimes(1)

    await vi.advanceTimersByTimeAsync(1)
    expect(g.$fetch).toHaveBeenCalledTimes(2)
    expect(lastParams()).toEqual({ year: 2026, page: 1, page_size: 10, search: 'audit' })

    // Typing back to the term already applied sends nothing.
    store.setKpiBreakdownSearch('audit')
    await vi.advanceTimersByTimeAsync(500)
    expect(g.$fetch).toHaveBeenCalledTimes(2)
  })

  it('a filter, search, year or page-size change goes back to page 1', async () => {
    vi.useFakeTimers()
    g.$fetch = pagedBackend(104)
    const store = usePerformanceStore()
    await store.loadKpiBreakdown(2026)

    await store.setKpiBreakdownPage(4)
    expect(lastParams().page).toBe(4)
    await store.setKpiBreakdownFilters({ status: 'Needs Attention' })
    expect(lastParams()).toMatchObject({ page: 1, status: 'Needs Attention' })

    await store.setKpiBreakdownPage(3)
    store.setKpiBreakdownSearch('report')
    await vi.advanceTimersByTimeAsync(300)
    expect(lastParams()).toMatchObject({ page: 1, search: 'report' })

    await store.setKpiBreakdownPage(2)
    await store.setKpiBreakdownFilters({ pageSize: 25 })
    expect(lastParams()).toMatchObject({ page: 1, page_size: 25 })

    await store.setKpiBreakdownPage(2)
    await store.loadKpiBreakdown(2025)
    expect(lastParams()).toMatchObject({ year: 2025, page: 1 })
    expect(store.kpiBreakdownQuery.page).toBe(1)
  })

  it('a reset clears every filter, cancels a pending search and refetches once', async () => {
    vi.useFakeTimers()
    g.$fetch = pagedBackend(30)
    const store = usePerformanceStore()
    await store.loadKpiBreakdown(2026)
    await store.setKpiBreakdownFilters({ status: 'Exceeded', period: 'Q2' })
    store.setKpiBreakdownSearch('pending')
    const before = g.$fetch.mock.calls.length

    await store.resetKpiBreakdownFilters()
    await vi.advanceTimersByTimeAsync(500)
    expect(g.$fetch.mock.calls.length).toBe(before + 1)
    expect(lastParams()).toEqual({ year: 2026, page: 1, page_size: 10 })
  })
})

describe('KPI breakdown store: pagination from the response metadata', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('104 rows at 10 per page: "1-10 of 104" on page 1, "101-104 of 104" on page 11', async () => {
    g.$fetch = pagedBackend(104)
    const store = usePerformanceStore()

    await store.loadKpiBreakdown(2026)
    expect(store.kpiBreakdown).toHaveLength(10)
    expect(store.kpiBreakdownPagination).toEqual({ page: 1, page_size: 10, total: 104, total_pages: 11 })
    expect(kpiBreakdownRangeText(tEn, store.kpiBreakdownPagination)).toBe('Showing 1-10 of 104')
    expect(kpiBreakdownRangeText(tId, store.kpiBreakdownPagination)).toBe('Menampilkan 1-10 dari 104')

    await store.setKpiBreakdownPage(11)
    expect(lastParams()).toMatchObject({ page: 11, page_size: 10 })
    expect(store.kpiBreakdown).toHaveLength(4)
    expect(store.kpiBreakdownPagination.total_pages).toBe(11)
    expect(kpiBreakdownRangeText(tEn, store.kpiBreakdownPagination)).toBe('Showing 101-104 of 104')
  })

  it('uses the page size the server reports', async () => {
    g.$fetch = fake(async () => ({
      success: true,
      data: { items: [item(1)], pagination: { page: 1, page_size: 100, total: 1, total_pages: 1 } }
    }))
    const store = usePerformanceStore()
    await store.setKpiBreakdownFilters({ year: 2026, pageSize: 500 })
    expect(lastParams().page_size).toBe(100) // clamped to the documented max before sending
    expect(store.kpiBreakdownPagination.page_size).toBe(100)
    expect(store.kpiBreakdownQuery.pageSize).toBe(100)
  })

  it('a page past the end (empty items, total > 0) moves to the last page', async () => {
    g.$fetch = pagedBackend(12)
    const store = usePerformanceStore()
    await store.loadKpiBreakdown(2026)
    await store.setKpiBreakdownPage(5)
    expect(callParams(-2).page).toBe(5)
    expect(lastParams().page).toBe(2)
    expect(store.kpiBreakdownPagination).toEqual({ page: 2, page_size: 10, total: 12, total_pages: 2 })
    expect(store.kpiBreakdown).toHaveLength(2)
  })

  it('range helpers: no "1-0 of 0" when there is no data', () => {
    expect(kpiBreakdownRange({ page: 1, page_size: 10, total: 0, total_pages: 0 })).toEqual({ from: 0, to: 0 })
    expect(kpiBreakdownRangeText(tEn, { page: 1, page_size: 10, total: 0, total_pages: 0 })).toBe(en.kpiPerformance.table.showingNone)
    expect(kpiBreakdownRangeText(tId, { page: 1, page_size: 10, total: 0, total_pages: 0 })).toBe(id.kpiPerformance.table.showingNone)
    expect(kpiBreakdownRange({ page: 3, page_size: 25, total: 60, total_pages: 3 })).toEqual({ from: 51, to: 60 })
  })
})

describe('KPI breakdown store: stale, empty and error responses', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('ignores a response that arrives after a newer request was started', async () => {
    const pending: Array<{ status: string, resolve: (v: unknown) => void }> = []
    g.$fetch = fake((_url, opts) => new Promise((resolve) => {
      pending.push({ status: String(opts?.params?.status ?? ''), resolve })
    }))
    const store = usePerformanceStore()

    const first = store.setKpiBreakdownFilters({ year: 2026, status: 'Exceeded' })
    const second = store.setKpiBreakdownFilters({ status: 'Needs Attention' })
    expect(pending.map(p => p.status)).toEqual(['Exceeded', 'Needs Attention'])

    const respond = (status: string, total: number) => ({
      success: true,
      data: { items: [item(1, { status, metric: status })], pagination: { page: 1, page_size: 10, total, total_pages: 1 } }
    })
    nth(pending, 1).resolve(respond('Needs Attention', 1))
    await second
    nth(pending, 0).resolve(respond('Exceeded', 7))
    await first

    expect(store.kpiBreakdown.map(r => r.metric)).toEqual(['Needs Attention'])
    expect(store.kpiBreakdownPagination.total).toBe(1)
    expect(store.kpiBreakdownLoading).toBe(false)
  })

  it('an older failure does not overwrite newer data', async () => {
    const pending: Array<{ resolve: (v: unknown) => void, reject: (e: unknown) => void }> = []
    g.$fetch = fake(() => new Promise((resolve, reject) => {
      pending.push({ resolve, reject })
    }))
    const store = usePerformanceStore()
    const first = store.setKpiBreakdownFilters({ year: 2026, status: 'Exceeded' })
    const second = store.setKpiBreakdownFilters({ status: 'On Track' })
    nth(pending, 1).resolve({ success: true, data: { items: [item(1)], pagination: { page: 1, page_size: 10, total: 1, total_pages: 1 } } })
    await second
    nth(pending, 0).reject(new Error('timeout'))
    await first
    expect(store.kpiBreakdown).toHaveLength(1)
    expect(store.kpiBreakdownError).toBeNull()
  })

  it('an empty response gives no rows and a zero total', async () => {
    g.$fetch = fake(async () => ({ success: true, data: { items: [], pagination: { page: 1, page_size: 10, total: 0, total_pages: 0 } } }))
    const store = usePerformanceStore()
    await store.loadKpiBreakdown(2026)
    expect(store.kpiBreakdown).toEqual([])
    expect(store.kpiBreakdownPagination).toEqual({ page: 1, page_size: 10, total: 0, total_pages: 0 })
    expect(store.kpiBreakdownError).toBeNull()
  })

  it('an error gives no rows, the error state and a consistent zero pagination', async () => {
    g.$fetch = pagedBackend(104)
    const store = usePerformanceStore()
    await store.loadKpiBreakdown(2026)
    expect(store.kpiBreakdown).toHaveLength(10)

    g.$fetch = failWith('Network error')
    await store.setKpiBreakdownPage(2)
    expect(store.kpiBreakdown).toEqual([])
    expect(store.kpiBreakdownError).toBeTruthy()
    expect(store.kpiBreakdownPagination.total).toBe(0)
    expect(store.kpiBreakdownPagination.total_pages).toBe(0)
    expect(store.kpiBreakdownLoading).toBe(false)
  })

  it('keeps the category the API sends ("" stays "", shown as "-"); no category is invented', async () => {
    g.$fetch = fake(async () => ({
      success: true,
      data: {
        items: [item(1), item(2, { period: 'Tahunan' }), item(3, { period: '2026' }), item(4, { category: 'Financial' })],
        pagination: { page: 1, page_size: 10, total: 4, total_pages: 1 }
      }
    }))
    const store = usePerformanceStore()
    await store.loadKpiBreakdown(2026)
    expect(store.kpiBreakdown.map(r => r.category)).toEqual(['', '', '', 'Financial'])
    expect(store.kpiBreakdownCategories).toEqual(['Financial'])
    expect(store.kpiBreakdownPeriods).toEqual(['Q1', 'Tahunan', '2026'])
    expect(kpiCategoryText(kpiValueLabel(tEn, 'categories', ''))).toBe('-')
    expect(kpiCategoryText(kpiValueLabel(tId, 'categories', 'Financial'))).toBe('Keuangan')
  })

  it('with only "" categories there are no category filter options', async () => {
    g.$fetch = pagedBackend(5)
    const store = usePerformanceStore()
    await store.loadKpiBreakdown(2026)
    expect(store.kpiBreakdownCategories).toEqual([])
  })
})

describe('KPI breakdown display helpers', () => {
  it('"No Target" has a label in both languages and a neutral colour', () => {
    expect(kpiValueLabel(tEn, 'statuses', 'No Target')).toBe('No target')
    expect(kpiValueLabel(tId, 'statuses', 'No Target')).toBe('Tanpa Target')
    expect(kpiStatusColor('No Target')).toBe('bg-gray-400')
  })

  it('maps the raw status value to its colour', () => {
    expect(kpiStatusColor('Exceeded')).toBe('bg-secondary-500')
    expect(kpiStatusColor('On Track')).toBe('bg-emerald-500')
    expect(kpiStatusColor('Needs Attention')).toBe('bg-red-500')
    // The translated label is not a key.
    expect(kpiStatusColor('Perlu Perhatian')).toBe('bg-gray-400')
  })

  it('formats values with the unit and rounds float noise', () => {
    expect(formatKpiValue(92.5, '%')).toBe('92.5%')
    expect(formatKpiValue(100, '%')).toBe('100%')
    expect(formatKpiValue(4.7, 'Score')).toBe('4.7')
    expect(formatKpiValue(13.456, '')).toBe('13.46')
    expect(formatKpiValue(1500000, 'Rp')).toBe('Rp 1,500,000')
    expect(formatKpiValue(14, 'Day')).toBe('14 Day')
    expect(formatKpiValue(92.5, '%', 'id-ID')).toBe('92,5%')
  })

  it('shows the gap with a sign', () => {
    expect(formatKpiGap(0.10000000000000053, '%')).toBe('+0.1%')
    expect(formatKpiGap(-3, '')).toBe('-3')
    expect(formatKpiGap(-2.25, '%')).toBe('-2.25%')
    expect(formatKpiGap(0, '%')).toBe('0%')
    expect(formatKpiGap(0.0000001, '%')).toBe('0%')
  })

  it('colours the gap by gapIsPositive, neutral at zero or without a target', () => {
    expect(kpiGapClass({ gap: 2, gapIsPositive: true, status: 'Exceeded' })).toBe('text-emerald-500')
    expect(kpiGapClass({ gap: 2, gapIsPositive: false, status: 'Needs Attention' })).toBe('text-red-500')
    expect(kpiGapClass({ gap: 0, gapIsPositive: true, status: 'On Track' })).toContain('text-gray-900')
    expect(kpiGapClass({ gap: 5, gapIsPositive: true, status: 'No Target' })).toContain('text-gray-900')
  })

  it('the table no longer builds rows itself (no merge, index category or mock rows)', () => {
    const src = readFileSync(resolve(__dirname, '../../components/kpi-performance/KpiDetailedTable.vue'), 'utf8')
    expect(src).not.toMatch(/categories\.length/)
    expect(src).not.toMatch(/mockData/)
    expect(src).not.toMatch(/strategicObjectives/)
    expect(src).not.toMatch(/kpiAchievements/)
    expect(src).not.toMatch(/total:\s*50/)
  })
})

describe('Stores used by /kpi-performance hold no mock data', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    g.getAuditServiceBaseUrl = () => '/api/v1'
    g.useAppToast = () => ({ success: vi.fn(), error: vi.fn(), info: vi.fn(), warning: vi.fn() })
    g.useGlobalModalStore = () => ({ confirmDelete: vi.fn(async () => true) })
  })

  it('performance store: empty or failed KPI / realization responses stay empty', async () => {
    g.$fetch = fake(async () => ({ success: true, data: [] }))
    const store = usePerformanceStore()
    await store.fetchKPIAchievements(2026)
    await store.fetchWorkPlanRealizations(2026)
    expect(store.kpiAchievements).toEqual([])
    expect(store.workPlanRealizations).toEqual([])
    expect((store as unknown as Record<string, unknown>).mockKpis).toBeUndefined()

    g.$fetch = failWith('down')
    await store.fetchKPIAchievements(2026)
    expect(store.kpiAchievements).toEqual([])
    expect(store.error).toBeTruthy()
    await store.fetchWorkPlanRealizations(2026)
    expect(store.workPlanRealizations).toEqual([])
    expect(store.error).toBeTruthy()
  })

  it('strategic plan store starts empty and stays empty on an empty response, without a sample-data toast', async () => {
    let resolveList: (v: unknown) => void = () => {}
    g.$fetch = fake(() => new Promise((resolve) => {
      resolveList = resolve
    }))
    const store = useStrategicPlanStore()
    expect(store.strategicObjectives).toEqual([]) // before the first response
    resolveList({ success: true, data: { items: [], pagination: { total: 0 } } })
    await flush()
    expect(store.strategicObjectives).toEqual([])
    expect(store.errorMsg).toBe('')
    expect(showInfo).not.toHaveBeenCalled()
  })

  it('strategic plan store: an error gives [] plus the error state', async () => {
    g.$fetch = failWith('down')
    const store = useStrategicPlanStore()
    await flush()
    expect(store.strategicObjectives).toEqual([])
    expect(store.errorMsg).toBeTruthy()
  })

  it('strategic plan store fetches one plan by id for editing', async () => {
    g.$fetch = fake(async (url: string) =>
      url.endsWith('/strategic-plans/abc')
        ? { success: true, data: { id: 'abc', kpi: 'Report Timeliness' } }
        : { success: true, data: { items: [] } })
    const store = useStrategicPlanStore()
    const plan = await store.fetchStrategicPlanById('abc')
    expect(plan).toEqual({ id: 'abc', kpi: 'Report Timeliness' })

    g.$fetch = failWith('404')
    expect(await store.fetchStrategicPlanById('missing')).toBeNull()
    expect(showError).toHaveBeenCalled()
  })
})
