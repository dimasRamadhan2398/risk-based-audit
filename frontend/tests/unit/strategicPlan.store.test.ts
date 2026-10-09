/**
 * Strategic plan store (stores/strategic-audit-plan.ts): a save or delete only changes the list
 * when the server accepted it, and the form sends `category` plus only the fields it edits.
 */
import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { buildStrategicPlanPayload, strategicPlanErrorField, strategicPlanFormFromPlan } from '~/utils/strategicPlanPayload'

const showSuccess = vi.fn()
const showError = vi.fn()
vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({ showSuccess, showError, showWarning: vi.fn(), showInfo: vi.fn() })
}))

type FetchOpts = { method?: string, body?: Record<string, unknown>, params?: Record<string, unknown> }
type FetchFn = (url: string, opts?: FetchOpts) => Promise<unknown>
const g = globalThis as unknown as { $fetch: Mock<FetchFn>, [name: string]: unknown }

const flush = () => new Promise(r => setTimeout(r, 0))

const plan = (id: string, extra: Record<string, unknown> = {}) => ({
  id,
  code: `SO-${id}`,
  goalId: '',
  strategicObjective: `Objective ${id}`,
  kpi: `KPI ${id}`,
  unit: '%',
  category: '',
  hibHig: 'HIG',
  periodType: 'Yearly',
  selectedPeriod: '2026',
  yearStart: 2026,
  yearEnd: 2030,
  kpiTargets: { 2026: '90' },
  kpiActuals: { 2026: '80' },
  internalAuditSO: '',
  actual: '80',
  target: '90',
  calculation: '88.89%',
  status: 'Moderate',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-02T00:00:00Z',
  ...extra
})

/** HTTP error shaped like ofetch's FetchError with the Go envelope from pkg/response.Error. */
const httpError = (statusCode: number, message: string) =>
  Object.assign(new Error(`[${statusCode}]`), { statusCode, data: { success: false, error: { code: 'ERR', message, details: '' } } })

/**
 * Fake API: GET /strategic-plans returns `list.items`; the given handler answers writes.
 * A successful write can change what the next GET returns via `list.items`.
 */
const backend = (list: { items: unknown[] }, onWrite: (method: string, url: string, body?: Record<string, unknown>) => unknown) =>
  vi.fn<FetchFn>(async (url, opts = {}) => {
    const method = (opts.method ?? 'GET').toUpperCase()
    if (method === 'GET') return { success: true, data: { items: list.items } }
    const result = onWrite(method, url, opts.body)
    if (result instanceof Error) throw result
    return result
  })

const writes = () => g.$fetch.mock.calls.filter(([, o]) => (o?.method ?? 'GET') !== 'GET')
const gets = () => g.$fetch.mock.calls.filter(([, o]) => (o?.method ?? 'GET') === 'GET')
const snapshot = (v: unknown) => JSON.parse(JSON.stringify(v))

/** Store with the initial list loaded. */
const loadedStore = async () => {
  const store = useStrategicPlanStore()
  await flush()
  return store
}

describe('strategic plan store: failed saves and deletes change nothing', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    g.getAuditServiceBaseUrl = () => '/api/v1'
    g.useGlobalModalStore = () => ({ confirmDelete: vi.fn(async () => true) })
  })

  it('a failed create does not add the form to the list; the modal stays open with the error', async () => {
    const list = { items: [plan('a')] }
    g.$fetch = backend(list, () => httpError(500, 'database unavailable'))
    const store = await loadedStore()
    const before = snapshot(store.strategicObjectives)

    store.openModal()
    store.form.strategicObjective = 'New objective'
    store.form.kpi = 'New KPI'
    store.form.category = 'Financial'
    const getsBefore = gets().length
    await store.handleSubmit()

    expect(writes()).toHaveLength(1)
    expect(snapshot(store.strategicObjectives)).toEqual(before)
    expect(store.strategicObjectives.some(o => o.strategicObjective === 'New objective')).toBe(false)
    expect(showError).toHaveBeenCalledTimes(1)
    expect(showSuccess).not.toHaveBeenCalled()
    expect(store.isAddModalOpen).toBe(true)
    expect(store.formError).toBe('database unavailable')
    // The input is kept for a retry.
    expect(store.form.strategicObjective).toBe('New objective')
    expect(store.form.category).toBe('Financial')
    expect(gets().length).toBe(getsBefore)
    expect(store.saving).toBe(false)
  })

  it('a failed update does not overwrite the row', async () => {
    const list = { items: [plan('a'), plan('b')] }
    g.$fetch = backend(list, () => httpError(500, 'boom'))
    const store = await loadedStore()
    const before = snapshot(store.strategicObjectives)

    store.handleEdit(nth(store.strategicObjectives, 1))
    store.form.kpi = 'Edited KPI'
    await store.handleSubmit()

    expect(writes()).toHaveLength(1)
    expect(snapshot(store.strategicObjectives)).toEqual(before)
    expect(showError).toHaveBeenCalledTimes(1)
    expect(store.isAddModalOpen).toBe(true)
    expect(store.isEditMode).toBe(true)
    expect(store.form.kpi).toBe('Edited KPI')
  })

  it('a failed delete keeps the row and shows the error', async () => {
    const list = { items: [plan('a'), plan('b')] }
    g.$fetch = backend(list, () => httpError(403, 'forbidden'))
    const store = await loadedStore()
    const before = snapshot(store.strategicObjectives)

    await store.handleDelete('a')

    expect(writes()).toHaveLength(1)
    expect(snapshot(store.strategicObjectives)).toEqual(before)
    expect(showError).toHaveBeenCalledTimes(1)
    expect(showSuccess).not.toHaveBeenCalled()
    expect(store.errorMsg).toBe('forbidden')
  })

  it('a 400 about the category is shown on the category field', async () => {
    g.$fetch = backend({ items: [] }, () => httpError(400, 'Invalid category "Strategic": must be one of Operational, Financial, Quality, Issue, Efficiency or empty'))
    const store = await loadedStore()
    store.openModal()
    store.form.strategicObjective = 'X'
    store.form.category = 'Strategic'
    await store.handleSubmit()

    expect(store.formFieldErrors.category).toMatch(/Invalid category/)
    expect(store.formError).toMatch(/Invalid category/)
    expect(store.isAddModalOpen).toBe(true)

    // Opening the form again starts without the old error.
    store.openModal()
    expect(store.formError).toBe('')
    expect(store.formFieldErrors).toEqual({})
  })
})

describe('strategic plan store: successful saves come from the server', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    g.getAuditServiceBaseUrl = () => '/api/v1'
    g.useGlobalModalStore = () => ({ confirmDelete: vi.fn(async () => true) })
  })

  it('a create POSTs the form with its category and then shows the list the server returns', async () => {
    const list = { items: [plan('a')] as unknown[] }
    g.$fetch = backend(list, (method, url, body) => {
      list.items = [...list.items, plan('server-id', { strategicObjective: body?.strategicObjective, category: body?.category })]
      return { success: true, data: { id: 'server-id' } }
    })
    const store = await loadedStore()

    store.openModal()
    store.form.strategicObjective = 'New objective'
    store.form.category = 'Quality'
    await store.handleSubmit()

    const [url, opts] = nth(writes(), 0)
    expect(url).toBe('/api/v1/strategic-plans')
    expect(opts?.method).toBe('POST')
    expect(opts?.body).toMatchObject({ strategicObjective: 'New objective', category: 'Quality' })
    expect(opts?.body).not.toHaveProperty('id')

    expect(store.strategicObjectives.map(o => o.id)).toEqual(['a', 'server-id'])
    expect(store.strategicObjectives.every(o => !/^\d{13}$/.test(String(o.id)))).toBe(true) // no Date.now() ids
    expect(showSuccess).toHaveBeenCalledTimes(1)
    expect(store.isAddModalOpen).toBe(false)
  })

  it('an update PUTs only the form fields (with category) and refetches', async () => {
    const list = { items: [plan('a', { category: 'Financial' })] as unknown[] }
    g.$fetch = backend(list, (_method, _url, body) => {
      list.items = [plan('a', { ...body })]
      return { success: true }
    })
    const store = await loadedStore()

    store.handleEdit(nth(store.strategicObjectives, 0))
    expect(store.form.category).toBe('Financial')
    store.form.category = '' // cleared in the form
    store.form.kpi = 'Renamed KPI'
    const getsBefore = gets().length
    await store.handleSubmit()

    const [url, opts] = nth(writes(), 0)
    expect(url).toBe('/api/v1/strategic-plans/a')
    expect(opts?.method).toBe('PUT')
    expect(opts?.body).toMatchObject({ kpi: 'Renamed KPI', category: '' })
    for (const key of ['id', 'created_at', 'updated_at']) expect(opts?.body).not.toHaveProperty(key)

    expect(gets().length).toBe(getsBefore + 1)
    expect(nth(store.strategicObjectives, 0)).toMatchObject({ id: 'a', kpi: 'Renamed KPI', category: '' })
    expect(store.isAddModalOpen).toBe(false)
  })

  it('a successful delete removes the row and refetches', async () => {
    const list = { items: [plan('a'), plan('b')] as unknown[] }
    g.$fetch = backend(list, () => {
      list.items = [plan('b')]
      return { success: true }
    })
    const store = await loadedStore()
    await store.handleDelete('a')
    expect(store.strategicObjectives.map(o => o.id)).toEqual(['b'])
    expect(showSuccess).toHaveBeenCalledTimes(1)
  })
})

describe('strategic plan payload helpers', () => {
  it('always sends category ("" when unset) and leaves out server-managed fields', () => {
    const body = buildStrategicPlanPayload({ ...plan('a'), category: undefined } as never)
    expect(body.category).toBe('')
    expect(body).not.toHaveProperty('id')
    expect(body).not.toHaveProperty('created_at')
    expect(body).toMatchObject({ kpi: 'KPI a', kpiTargets: { 2026: '90' } })
  })

  it('loads an older row without a category as ""', () => {
    const { category: _omit, ...older } = plan('a')
    const form = strategicPlanFormFromPlan(older as never)
    expect(form.category).toBe('')
    expect(form.id).toBe('a')
    expect(form).not.toHaveProperty('updated_at')
  })

  it('only a 400 mentioning the category is a category field error', () => {
    expect(strategicPlanErrorField(httpError(400, 'x'), 'Invalid category "x"')).toBe('category')
    expect(strategicPlanErrorField(httpError(400, 'x'), 'kpi is required')).toBeNull()
    expect(strategicPlanErrorField(httpError(500, 'x'), 'Invalid category')).toBeNull()
    expect(strategicPlanErrorField(null, 'Invalid category')).toBeNull()
  })
})

/** Element of a list that the test expects to exist. */
function nth<T>(list: T[], index: number): T {
  const value = list[index]
  if (value === undefined) throw new Error(`no element ${index}`)
  return value
}
