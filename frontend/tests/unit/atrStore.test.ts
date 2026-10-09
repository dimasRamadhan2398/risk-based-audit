// @ts-nocheck
/**
 * ATR store (stores/action-taken-report.ts): server-paged list without any sample data, filters sent to
 * GET /action-taken-reports, stale responses ignored, and one endpoint + payload per action.
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useAuthStore } from '~/stores/auth'

const showSuccess = vi.fn()
const showError = vi.fn()
vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({ showSuccess, showError, showWarning: vi.fn(), showInfo: vi.fn(), add: vi.fn() })
}))

const BASE = '/api/v1/action-taken-reports'

const row = (n, extra = {}) => ({
  id: `atr-${n}`,
  audit_result_report_id: 'lha-1',
  report_number: 'LHA-021/SKAI/2026',
  report_title: 'Audit Kas',
  finding_id: `f-${n}`,
  finding_title: `Finding ${n}`,
  finding_category: 'Significant',
  recommendation: 'Do it',
  action_plan: '',
  pic_user_id: 'me',
  pic_name: 'Me',
  due_date: '2027-01-01T00:00:00Z',
  progress: 0,
  status: 'PLANNED',
  is_overdue: false,
  overdue_days: 0,
  evidence: [],
  review_note: '',
  reviewed_by: '',
  reviewed_at: null,
  created_at: '2026-10-01T00:00:00Z',
  updated_at: '2026-10-01T00:00:00Z',
  ...extra
})

const page = (items, pagination) => ({ success: true, data: { items, pagination } })

/** A backend with `total` rows that honours page/page_size, and answers the action endpoints. */
const backend = (total = 3) => vi.fn(async (url, opts = {}) => {
  const method = (opts.method || 'GET').toUpperCase()
  if (method === 'GET' && url === BASE) {
    const p = Number(opts.params?.page ?? 1)
    const size = Number(opts.params?.page_size ?? 10)
    const start = (p - 1) * size
    const count = Math.max(0, Math.min(size, total - start))
    return page(Array.from({ length: count }, (_, i) => row(start + i + 1)), { page: p, page_size: size, total, total_pages: Math.ceil(total / size) })
  }
  if (method === 'GET' && url.startsWith(`${BASE}/`) && !url.includes('/evidence/')) return { success: true, data: row(1, { id: url.split('/').pop() }) }
  if (url.includes('/evidence/')) return new Blob(['pdf'])
  return { success: true, data: row(1, { status: 'IN_PROGRESS' }) }
})

const calls = (fn, predicate) => fn.mock.calls.filter(([url, opts = {}]) => predicate(url, (opts.method || 'GET').toUpperCase(), opts))

describe('ATR store: list', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
    useAuthStore().user = { id: 'me', roles: ['auditee'] }
  })
  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('starts empty and does not fetch on creation (no mock data)', () => {
    global.$fetch = backend()
    const store = useActionTakenReportStore()
    expect(store.items).toEqual([])
    expect(store.summaryItems).toEqual([])
    expect(store.stats.total).toBe(0)
    expect(global.$fetch).not.toHaveBeenCalled()
  })

  it('an empty response gives an empty list and no error', async () => {
    global.$fetch = vi.fn().mockResolvedValue(page([], { page: 1, page_size: 10, total: 0, total_pages: 0 }))
    const store = useActionTakenReportStore()
    await store.fetchReports()
    expect(store.items).toEqual([])
    expect(store.errorMsg).toBe('')
    expect(store.pagination.total).toBe(0)
  })

  it('an error gives an empty list plus the error state (list and summary)', async () => {
    global.$fetch = vi.fn().mockRejectedValue(Object.assign(new Error('boom'), { data: { message: 'Server down' } }))
    const store = useActionTakenReportStore()
    await store.fetchReports()
    await store.fetchSummary()
    expect(store.items).toEqual([])
    expect(store.errorMsg).toBeTruthy()
    expect(store.pagination).toEqual({ page: 1, page_size: 10, total: 0, total_pages: 0 })
    expect(store.summaryItems).toEqual([])
    expect(store.summaryError).toBeTruthy()
  })

  it('pagination comes only from the response metadata', async () => {
    global.$fetch = vi.fn().mockResolvedValue(page([row(11)], { page: 2, page_size: 25, total: 26, total_pages: 2 }))
    const store = useActionTakenReportStore()
    await store.setPage(2)
    expect(store.pagination).toEqual({ page: 2, page_size: 25, total: 26, total_pages: 2 })
    expect(store.query.page).toBe(2)
    expect(store.query.pageSize).toBe(25)
    expect(store.items.map(i => i.id)).toEqual(['atr-11'])
  })

  it('jumps to the last page when the requested one no longer exists', async () => {
    global.$fetch = backend(12)
    const store = useActionTakenReportStore()
    await store.setPage(5)
    expect(store.pagination.page).toBe(2)
    expect(store.items).toHaveLength(2)
  })

  it('sends status, LHA, PIC and overdue filters, and resets to page 1', async () => {
    global.$fetch = backend(30)
    const store = useActionTakenReportStore()
    await store.initList({ scope: 'all' })
    await store.setPage(3)
    await store.setFilters({ status: 'PENDING_REVIEW', lhaId: 'lha-9', picUserId: 'u-7', overdue: true })
    const [, opts] = global.$fetch.mock.calls.at(-1)
    expect(opts.params).toEqual({ page: 1, page_size: 10, status: 'PENDING_REVIEW', audit_result_report_id: 'lha-9', pic_user_id: 'u-7', overdue: true })
    await store.resetFilters()
    expect(global.$fetch.mock.calls.at(-1)[1].params).toEqual({ page: 1, page_size: 10 })
  })

  it('"My actions" sends the logged-in user as pic_user_id; ?lha= is applied on init', async () => {
    global.$fetch = backend()
    const store = useActionTakenReportStore()
    await store.initList({ scope: 'mine', lhaId: 'lha-3' })
    expect(global.$fetch.mock.calls.at(-1)[1].params).toEqual({ page: 1, page_size: 10, audit_result_report_id: 'lha-3', pic_user_id: 'me' })
  })

  it('debounces search and sends only the last term', async () => {
    vi.useFakeTimers()
    global.$fetch = backend()
    const store = useActionTakenReportStore()
    store.setSearch('ka')
    store.setSearch('kas ')
    await vi.advanceTimersByTimeAsync(299)
    expect(global.$fetch).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1)
    expect(global.$fetch).toHaveBeenCalledTimes(1)
    expect(global.$fetch.mock.calls[0][1].params).toEqual({ page: 1, page_size: 10, search: 'kas' })
  })

  it('ignores a stale response that arrives after a newer one', async () => {
    let releaseFirst
    global.$fetch = vi.fn()
      .mockImplementationOnce(() => new Promise((resolve) => { releaseFirst = () => resolve(page([row(1)], { page: 1, page_size: 10, total: 1, total_pages: 1 })) }))
      .mockImplementationOnce(async () => page([row(2), row(3)], { page: 1, page_size: 10, total: 2, total_pages: 1 }))
    const store = useActionTakenReportStore()
    const first = store.fetchReports()
    await store.setFilters({ status: 'IN_PROGRESS' })
    releaseFirst()
    await first
    expect(store.items.map(i => i.id)).toEqual(['atr-2', 'atr-3'])
    expect(store.pagination.total).toBe(2)
    expect(store.loading).toBe(false)
  })

  it('summary pages through every ATR for the dashboard stats', async () => {
    global.$fetch = backend(230)
    const store = useActionTakenReportStore()
    await store.fetchSummary()
    expect(store.summaryItems).toHaveLength(230)
    expect(calls(global.$fetch, url => url === BASE).map(([, o]) => o.params)).toEqual([
      { page: 1, page_size: 100 }, { page: 2, page_size: 100 }, { page: 3, page_size: 100 }
    ])
    expect(store.stats.total).toBe(230)
  })

  it('loads only approved LHAs for the LHA filter', async () => {
    global.$fetch = vi.fn().mockResolvedValue({
      data: { items: [
        { id: 'a', reportNumber: 'LHA-1', reportTitle: 'A', status: 'Final' },
        { id: 'b', report_number: 'LHA-2', report_title: 'B', status: 'Draft' }
      ] }
    })
    const store = useActionTakenReportStore()
    await store.fetchLhaOptions()
    expect(global.$fetch.mock.calls[0][0]).toBe('/api/v1/audit-result-reports')
    expect(store.lhaOptions).toEqual([{ id: 'a', reportNumber: 'LHA-1', reportTitle: 'A' }])
  })

  it('counts the ATRs of one LHA from pagination.total', async () => {
    global.$fetch = backend(7)
    const store = useActionTakenReportStore()
    expect(await store.countForLha('lha-1')).toBe(7)
    expect(global.$fetch.mock.calls[0][1].params).toEqual({ audit_result_report_id: 'lha-1', page: 1, page_size: 1 })
  })
})

describe('ATR store: PIC candidates', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })
  afterEach(() => vi.restoreAllMocks())

  it('searches GET /users/assignable on the auth service', async () => {
    global.$fetch = vi.fn().mockResolvedValue({
      success: true,
      data: { users: [{ id: 'u1', full_name: 'Rina', department: 'Finance', position: 'Manager' }], pagination: { page: 1, page_size: 20, total: 1, total_pages: 1 } }
    })
    const store = useActionTakenReportStore()
    const result = await store.fetchPicCandidates('rin')
    expect(global.$fetch.mock.calls[0][0]).toBe('/api/v1/users/assignable')
    expect(global.$fetch.mock.calls[0][1].params).toEqual({ search: 'rin', page: 1, page_size: 20 })
    expect(result).toEqual({ users: [{ id: 'u1', full_name: 'Rina', department: 'Finance', position: 'Manager' }], forbidden: false, error: '' })
  })

  it('reports 403 as forbidden instead of falling back to a list', async () => {
    global.$fetch = vi.fn().mockRejectedValue(Object.assign(new Error('Forbidden'), { statusCode: 403, status: 403 }))
    const store = useActionTakenReportStore()
    expect(await store.fetchPicCandidates('')).toEqual({ users: [], forbidden: true, error: '' })
  })
})

describe('ATR store: actions call the right endpoint with the right payload', () => {
  let store
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
    useAuthStore().user = { id: 'me', roles: ['auditee'] }
    global.$fetch = backend()
    store = useActionTakenReportStore()
  })
  afterEach(() => vi.restoreAllMocks())

  const lastWrite = () => calls(global.$fetch, (_u, m) => m !== 'GET').at(-1)

  it('assign: PUT /:id/assignment { pic_user_id, due_date (RFC3339) }', async () => {
    expect(await store.assign('atr-1', { pic_user_id: 'u5', due_date: '2026-12-31' })).toBe(true)
    const [url, opts] = lastWrite()
    expect(url).toBe(`${BASE}/atr-1/assignment`)
    expect(opts.method).toBe('PUT')
    expect(opts.body).toEqual({ pic_user_id: 'u5', due_date: '2026-12-31T00:00:00Z' })
    expect(opts.headers).toBeUndefined()
    expect(showSuccess).toHaveBeenCalled()
  })

  it('saveActionPlan: PUT /:id/action-plan with only { action_plan, progress }', async () => {
    await store.saveActionPlan('atr-1', { action_plan: '  Plan  ', progress: 42.4 })
    const [url, opts] = lastWrite()
    expect(url).toBe(`${BASE}/atr-1/action-plan`)
    expect(opts.method).toBe('PUT')
    expect(opts.body).toEqual({ action_plan: 'Plan', progress: 42 })
  })

  it('submit: POST /:id/submit', async () => {
    await store.submit('atr-1')
    const [url, opts] = lastWrite()
    expect(url).toBe(`${BASE}/atr-1/submit`)
    expect(opts.method).toBe('POST')
    expect(opts.body).toBeUndefined()
  })

  it('review: POST /:id/review { decision, note }', async () => {
    await store.review('atr-1', { decision: 'reject', note: ' Add evidence ' })
    const [url, opts] = lastWrite()
    expect(url).toBe(`${BASE}/atr-1/review`)
    expect(opts).toMatchObject({ method: 'POST', body: { decision: 'reject', note: 'Add evidence' } })
  })

  it('cancel: POST /:id/cancel { note }', async () => {
    await store.cancel('atr-1', 'Duplicate')
    const [url, opts] = lastWrite()
    expect(url).toBe(`${BASE}/atr-1/cancel`)
    expect(opts).toMatchObject({ method: 'POST', body: { note: 'Duplicate' } })
  })

  it('evidence upload: one multipart POST per file, field "file", no Content-Type header', async () => {
    const files = [new File(['a'], 'a.pdf'), new File(['b'], 'b.png')]
    expect(await store.uploadEvidence('atr-1', files)).toBe(2)
    const posts = calls(global.$fetch, (u, m) => m === 'POST' && u === `${BASE}/atr-1/evidence`)
    expect(posts).toHaveLength(2)
    for (const [, opts] of posts) {
      expect(opts.body).toBeInstanceOf(FormData)
      expect(opts.headers).toBeUndefined()
    }
    expect(posts[0][1].body.get('file').name).toBe('a.pdf')
    expect(posts[1][1].body.get('file').name).toBe('b.png')
  })

  it('evidence download: GET /:id/evidence/:evidenceId as a blob (token added by the auth plugin)', async () => {
    URL.createObjectURL = vi.fn(() => 'blob:x')
    URL.revokeObjectURL = vi.fn()
    await store.downloadEvidence('atr-1', { id: 'ev-9', file_name: 'proof.pdf', uploaded_at: '', uploaded_by: '' })
    const [url, opts] = calls(global.$fetch, u => u.includes('/evidence/'))[0]
    expect(url).toBe(`${BASE}/atr-1/evidence/ev-9`)
    expect(opts).toMatchObject({ method: 'GET', responseType: 'blob' })
    expect(opts.headers).toBeUndefined()
  })

  it('refreshes the list and the summary after a change', async () => {
    await store.submit('atr-1')
    const listCalls = calls(global.$fetch, (u, m) => m === 'GET' && u === BASE)
    expect(listCalls.length).toBeGreaterThanOrEqual(2)
  })

  it('a failed action returns false, toasts the error and keeps going', async () => {
    global.$fetch = vi.fn().mockRejectedValue(Object.assign(new Error('nope'), { statusCode: 409, data: { message: 'Not allowed in this status' } }))
    expect(await store.submit('atr-1')).toBe(false)
    expect(showError).toHaveBeenCalled()
    expect(store.saving).toBe(false)
  })
})
