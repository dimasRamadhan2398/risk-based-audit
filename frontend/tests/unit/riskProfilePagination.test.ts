// @ts-nocheck
/**
 * Corporate Risk Profile history tables (stores/risk-profile.ts): 10 per page,
 * server-side filters in the query, page reset on filter change, page kept on refresh.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useRiskProfileStore, RISK_PAGE_SIZE, UNASSIGNED_BRANCH } from '~/stores/risk-profile'
import { paginateList } from '~/utils/clientPagination'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({ public: { masterServiceBaseUrl: 'http://localhost:8080/api/v1' } })
}))

const YEAR = 2026
const row = n => ({
  id: `r-${n}`, name: `Risk ${n}`, category: 'Financial', impact: 3, likelihood: 3, severity: 50,
  assessments: [{ year: YEAR, impact_q1: 3, likelihood_q1: 3, risk_level_q1: 'moderate' }]
})

/** Backend with `total` rows honouring page/page_size; records every GET /risks?page call. */
const backend = (total = 25) => {
  const calls = []
  const fn = vi.fn(async (url, opts = {}) => {
    const u = String(url)
    if (u.endsWith('/risks/categories')) return { success: true, data: ['Technology', 'Financial', 'Financial'] }
    if (u.includes('/locations')) return { success: true, data: [] }
    if (u.endsWith('/risks') && opts.query?.page) {
      calls.push(opts.query)
      fn.cacheModes.push(opts.cache)
      const { page, page_size } = opts.query
      const items = Array.from({ length: total }, (_, i) => row(i + 1)).slice((page - 1) * page_size, page * page_size)
      return { success: true, message: 'ok', data: items, pagination: { page, page_size, total, total_pages: Math.ceil(total / page_size) } }
    }
    return { success: true, data: [] }
  })
  fn.pageCalls = calls
  fn.cacheModes = []
  return fn
}

describe('risk profile pagination', () => {
  let store
  beforeEach(() => {
    setActivePinia(createPinia())
    global.$fetch = backend()
    store = useRiskProfileStore()
    store.selectedYear = YEAR
    store.selectedPeriod = 'Q1'
  })

  it('requests page 1 with 10 rows and the filters, omitting empty ones', async () => {
    await store.fetchRiskPage('all')
    expect(global.$fetch.pageCalls[0]).toEqual({ page: 1, page_size: 10, fiscal_year: YEAR })
    expect(store.riskPages.all.items).toHaveLength(10)
    expect(store.riskPages.all.total).toBe(25)
    expect(store.riskPages.all.totalPages).toBe(3)
  })

  it('sends category, branch_id and search when set', async () => {
    store.selectedCategory = 'Technology'
    store.selectedBranch = 'loc-1'
    store.searchQuery = '  fraud '
    await store.fetchRiskPage('progress')
    expect(global.$fetch.pageCalls[0]).toMatchObject({ category: 'Technology', branch_id: 'loc-1', search: 'fraud', page: 1 })
  })

  it('loads the requested page and keeps it on refresh', async () => {
    await store.fetchRiskPage('all', 3)
    expect(store.riskPages.all.page).toBe(3)
    expect(store.riskPages.all.items).toHaveLength(5)
    await store.refreshRiskPages()
    expect(global.$fetch.pageCalls.at(-1).page).toBe(3)
    expect(store.riskPages.all.page).toBe(3)
  })

  it('goes back to the last page when the current one became empty', async () => {
    await store.fetchRiskPage('all', 3)
    global.$fetch = backend(20)
    await store.refreshRiskPages()
    expect(store.riskPages.all.page).toBe(2)
    expect(store.riskPages.all.items).toHaveLength(10)
  })

  it.each([
    ['selectedBranch', 'loc-2'],
    ['selectedCategory', 'Technology'],
    ['selectedPeriod', 'Q2'],
    ['selectedYear', 2025],
    ['searchQuery', 'x']
  ])('resets every table to page 1 when %s changes', async (field, value) => {
    await store.fetchRiskPage('all', 2)
    await store.fetchRiskPage('progress', 3)
    store.priorityPage = 2
    store[field] = value
    expect(store.riskPages.all.page).toBe(1)
    expect(store.riskPages.progress.page).toBe(1)
    expect(store.priorityPage).toBe(1)
  })

  it('shows an empty state without a page when there are no rows', async () => {
    global.$fetch = backend(0)
    await store.fetchRiskPage('all')
    expect(store.riskPages.all.items).toEqual([])
    expect(store.riskPages.all.total).toBe(0)
  })

  it('records an error and clears the rows when the request fails', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    global.$fetch = vi.fn(async () => {
      throw new Error('boom')
    })
    await store.fetchRiskPage('all')
    expect(store.riskPages.all.error).not.toBe('')
    expect(store.riskPages.all.items).toEqual([])
  })

  it('sends branch_id=unassigned to the server for the Unassigned branch', async () => {
    store.selectedBranch = UNASSIGNED_BRANCH
    await store.fetchRiskPage('all')
    expect(global.$fetch.pageCalls[0].branch_id).toBe('unassigned')
  })

  it('never sends period (not a server filter) and still applies the chosen quarter client-side', async () => {
    store.selectedPeriod = 'Q1'
    await store.fetchRiskPage('all')
    expect(global.$fetch.pageCalls[0]).not.toHaveProperty('period')
    expect(store.riskPages.all.items[0].riskLevel).toBe('moderate')
  })

  it('bypasses the browser cache so a refetch after a mutation is fresh', async () => {
    await store.fetchRiskPage('all', 2)
    await store.refreshRiskPages()
    expect(global.$fetch.cacheModes.every(m => m === 'no-store')).toBe(true)
  })

  it('still reads the older { data: { items, pagination } } shape', async () => {
    global.$fetch = vi.fn(async () => ({ success: true, data: { items: [row(1)], pagination: { page: 1, page_size: 10, total: 1, total_pages: 1 } } }))
    await store.fetchRiskPage('all')
    expect(store.riskPages.all.items).toHaveLength(1)
    expect(store.riskPages.all.total).toBe(1)
  })

  it('loads distinct category names from /risks/categories', async () => {
    await store.fetchCategories()
    expect(store.categories).toEqual(['Financial', 'Technology'])
  })

  it('leaves categories empty (page falls back to loaded data) when the call fails', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    global.$fetch = vi.fn(async () => {
      throw new Error('404')
    })
    await store.fetchCategories()
    expect(store.categories).toEqual([])
    expect(store.categoriesError).not.toBe('')
  })
})

describe('paginateList (Priority table, client-side)', () => {
  const list = Array.from({ length: 23 }, (_, i) => i)
  it('slices 10 per page and counts pages', () => {
    const p = paginateList(list, 3, RISK_PAGE_SIZE)
    expect(p.items).toEqual([20, 21, 22])
    expect(p.totalPages).toBe(3)
    expect(p.total).toBe(23)
  })
  it('clamps a page past the end and handles an empty list', () => {
    expect(paginateList(list, 9, 10).page).toBe(3)
    expect(paginateList([], 4, 10)).toMatchObject({ items: [], page: 1, total: 0, totalPages: 0 })
  })
})
