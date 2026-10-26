/**
 * Strategic plan form rules (stores/strategic-audit-plan.ts, utils/strategicPlanPayload.ts):
 * the end year is never before the start year, targets/realizations are numbers, and the code
 * of a new plan is left to the server (it used to be "SO-IA{loaded rows + 1}", which collided).
 */
import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { nextTick } from 'vue'
import { setActivePinia, createPinia } from 'pinia'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { firstInvalidKpiValue, isKpiNumber, sanitizeKpiNumberInput } from '~/utils/strategicPlanPayload'

const showWarning = vi.fn()
vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({ showSuccess: vi.fn(), showError: vi.fn(), showWarning, showInfo: vi.fn() })
}))

type FetchOpts = { method?: string, body?: Record<string, unknown> }
type FetchFn = (url: string, opts?: FetchOpts) => Promise<unknown>
const g = globalThis as unknown as { $fetch: Mock<FetchFn>, [name: string]: unknown }
const flush = () => new Promise(r => setTimeout(r, 0))
const writes = () => g.$fetch.mock.calls.filter(([, o]) => (o?.method ?? 'GET') !== 'GET')

// 100 loaded rows, as after the seeder (SO-IA01..SO-IA100) with the list's page_size of 100.
const seeded = Array.from({ length: 100 }, (_, i) => ({ id: `id-${i}`, code: `SO-IA${String(i + 1).padStart(2, '0')}` }))

const newStore = async () => {
  g.$fetch = vi.fn<FetchFn>(async (_url, opts = {}) =>
    (opts.method ?? 'GET') === 'GET' ? { success: true, data: { items: seeded } } : { success: true })
  const store = useStrategicPlanStore()
  await flush()
  store.openModal()
  store.form.strategicObjective = 'Objective'
  return store
}

describe('strategic plan form validation', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    g.getAuditServiceBaseUrl = () => '/api/v1'
  })

  it('a new plan is POSTed without a code: the server assigns it', async () => {
    const store = await newStore()
    await store.handleSubmit()
    expect(writes()).toHaveLength(1)
    expect(writes()[0]?.[1]?.body?.code ?? '').toBe('')
  })

  it('an end year before the start year is not saved', async () => {
    const store = await newStore()
    store.form.yearStart = 2028
    store.form.yearEnd = 2026
    await store.handleSubmit()
    expect(writes()).toHaveLength(0)
    expect(showWarning).toHaveBeenCalledTimes(1)
  })

  it('moving the start year past the end year moves the end year along; "to" lists no earlier year', async () => {
    const store = await newStore()
    const start = store.form.yearStart as number
    store.form.yearEnd = start
    await nextTick()
    store.form.yearStart = start + 2
    await nextTick()
    expect(store.form.yearEnd).toBe(start + 2)
    expect(store.yearEndOptions.every(o => o.value >= start + 2)).toBe(true)
  })

  it('a target or realization that is not a number is not saved', async () => {
    const store = await newStore()
    const year = String(store.form.yearStart)
    store.form.kpiTargets = { [year]: '90' }
    store.form.kpiActuals = { [year]: '88%' }
    await store.handleSubmit()
    expect(writes()).toHaveLength(0)
    expect(showWarning).toHaveBeenCalledTimes(1)
  })
})

describe('KPI number helpers', () => {
  it('sanitizeKpiNumberInput keeps digits, one decimal point and a leading minus', () => {
    expect(sanitizeKpiNumberInput('85a')).toBe('85')
    expect(sanitizeKpiNumberInput('85,5')).toBe('85.5')
    expect(sanitizeKpiNumberInput('1.2.3')).toBe('1.23')
    expect(sanitizeKpiNumberInput('Rp 1000')).toBe('1000')
    expect(sanitizeKpiNumberInput('88%')).toBe('88')
    expect(sanitizeKpiNumberInput('-2')).toBe('-2')
    expect(sanitizeKpiNumberInput('2-')).toBe('2')
    expect(sanitizeKpiNumberInput('')).toBe('')
  })

  it('isKpiNumber accepts empty and plain numbers only', () => {
    for (const ok of ['', ' ', '0', '85', '85.5', '-1.25']) expect(isKpiNumber(ok)).toBe(true)
    for (const bad of ['85%', 'abc', '1.', '.5', '1e3', '1.000.000']) expect(isKpiNumber(bad)).toBe(false)
  })

  it('firstInvalidKpiValue names the first bad period, targets before realizations', () => {
    expect(firstInvalidKpiValue({ kpiTargets: { 2026: '90' }, kpiActuals: { 2026: '' } })).toBeNull()
    expect(firstInvalidKpiValue({ kpiTargets: { 2027: 'x', 2026: 'y' }, kpiActuals: { 2026: 'z' } }))
      .toEqual({ field: 'target', period: '2026' })
    expect(firstInvalidKpiValue({ kpiTargets: {}, kpiActuals: { 'Q2-2026': '5%' } }))
      .toEqual({ field: 'realization', period: 'Q2-2026' })
  })
})
