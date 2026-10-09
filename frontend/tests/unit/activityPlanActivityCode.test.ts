/**
 * Activity ID (plannedActivities[].activityCode) of an Audit Activity Plan. The backend assigns it
 * (audit-service pkg/activitycode) and keeps it on update when an activity sends back its stored `id`
 * or `activityCode`. The frontend must show it and must never send a code it made up.
 */
import { describe, it, expect, beforeEach, afterEach, vi, type Mock } from 'vitest'
import { createApp, defineComponent, h, nextTick, type App } from 'vue'
import { setActivePinia, createPinia, type Pinia } from 'pinia'
import en from '~/locales/en/common.json'
import id from '~/locales/id/common.json'
import { useActivityPlanStore } from '~/stores/activity-plan'
import ActivityPlanForm from '~/components/audit-activity-plan/ActivityPlanForm.vue'
import ActivityPlanTable from '~/components/audit-activity-plan/ActivityPlanTable.vue'
import ActivityPlanViewModal from '~/components/audit-activity-plan/ActivityPlanViewModal.vue'

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({ showSuccess: vi.fn(), showError: vi.fn(), showWarning: vi.fn(), showInfo: vi.fn() })
}))
vi.mock('~/composables/useDepartmentApi', () => ({
  useDepartmentApi: () => ({ getAllDepartments: async () => [] })
}))

type Body = Record<string, unknown> & { plannedActivities?: Array<Record<string, unknown>> }
type FetchOpts = { method?: string, body?: Body, query?: Record<string, unknown> }
type FetchFn = (url: string, opts?: FetchOpts) => Promise<unknown>
const g = globalThis as unknown as { $fetch: Mock<FetchFn>, [name: string]: unknown }

const flush = () => new Promise(r => setTimeout(r, 0))

const activity = (extra: Record<string, unknown> = {}) => ({
  auditName: 'Audit IT',
  auditee: 'IT Dept',
  category: 'Assurance',
  riskName: 'Data breach',
  riskLevel: 'high',
  duration: 10,
  priority: 'p1',
  numberOfAuditors: 2,
  estimatedSchedule: '2026-03-01',
  budgetEstimation: 1000,
  ...extra
})

/** A plan as GET /activity-plans returns it. */
const storedPlan = (plannedActivities: unknown[]) => ({
  id: 'plan-1',
  planTitle: 'Plan 2026',
  planYear: '2026',
  planPeriodStart: '2026-01-01',
  planPeriodEnd: '2026-12-31',
  department: 'IT',
  createdBy: 'Admin',
  creationDate: '2026-01-01',
  plannedActivities,
  resourceAuditors: [],
  budget: { totalEstimatedCost: 0, totalAllocatedBudget: 0, budgetNotes: '' },
  review: { creatorName: 'A', creatorPosition: 'B', approverName: 'C', approverPosition: 'D', approvalDate: '', additionalNotes: '' },
  attachments: [],
  status: 'draft',
  createdAt: '2026-01-01T00:00:00Z'
})

/** GET /activity-plans returns `items`, GET /risks nothing; writes succeed. */
const backend = (items: unknown[]) => vi.fn<FetchFn>(async (url, opts = {}) => {
  if ((opts.method ?? 'GET').toUpperCase() !== 'GET') return { success: true }
  if (url.endsWith('/activity-plans')) return { success: true, data: { items, pagination: { total_pages: 1 } } }
  return { success: true, data: [] }
})
const writes = () => g.$fetch.mock.calls.filter(([, o]) => (o?.method ?? 'GET') !== 'GET')
const lastWrite = () => {
  const call = writes().at(-1)
  if (!call) throw new Error('no write')
  return { url: call[0], method: call[1]?.method, body: call[1]?.body as Body }
}

beforeEach(() => {
  vi.clearAllMocks()
  g.useGlobalModalStore = () => ({ confirmDelete: vi.fn(async () => true) })
})

describe('store: save payload', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('fetchPlans keeps id and activityCode of each planned activity', async () => {
    g.$fetch = backend([storedPlan([activity({ id: 'a1', activityCode: 'ASR-2026-001' })])])
    const store = useActivityPlanStore()
    await store.fetchPlans()
    expect(store.plans[0]?.plannedActivities[0]).toMatchObject({ id: 'a1', activityCode: 'ASR-2026-001' })
    expect(store.filteredPlans[0]?.plannedActivities[0]).toMatchObject({ id: 'a1', activityCode: 'ASR-2026-001' })
  })

  it('a create sends no activityCode, also when a row somehow carries one', async () => {
    g.$fetch = backend([])
    const store = useActivityPlanStore()
    store.openModal()
    Object.assign(store.formState, { planTitle: 'New', planPeriodStart: '2026-01-01', planPeriodEnd: '2026-12-31', department: 'IT' })
    store.addPlannedActivity()
    store.addPlannedActivity()
    const [first, second] = store.formState.plannedActivities
    expect(first).not.toHaveProperty('activityCode')
    // A made-up code must not reach the server.
    second!.activityCode = 'ASR-2026-999'

    await store.savePlan()

    const { url, method, body } = lastWrite()
    expect(method).toBe('POST')
    expect(url).toMatch(/\/activity-plans$/)
    expect(body.plannedActivities).toHaveLength(2)
    for (const sent of body.plannedActivities!) expect(sent).not.toHaveProperty('activityCode')
    expect(body.plannedActivities![0]!.id).toBe(first!.id)
  })

  it('an edit keeps id and activityCode of stored rows; new rows go without activityCode', async () => {
    g.$fetch = backend([storedPlan([
      activity({ id: 'a1', activityCode: 'ASR-2026-001' }),
      // Legacy row without a client id: the backend matches it by its code.
      activity({ category: 'Special Audit', activityCode: 'SPC-2026-004' }),
      // Stored before codes existed: the backend gives it one on this save.
      activity({ id: 'a3', activityCode: '' })
    ])])
    const store = useActivityPlanStore()
    await store.fetchPlans()
    store.handleEdit(store.filteredPlans[0] as never)
    // Changing the category does not touch the code; the backend keeps it.
    store.formState.plannedActivities[0]!.category = 'Investigation' as never
    store.addPlannedActivity()
    const added = store.formState.plannedActivities[3]!

    await store.savePlan()

    const { url, method, body } = lastWrite()
    expect(method).toBe('PUT')
    expect(url).toMatch(/\/activity-plans\/plan-1$/)
    const sent = body.plannedActivities!
    expect(sent).toHaveLength(4)
    expect(sent[0]).toMatchObject({ id: 'a1', activityCode: 'ASR-2026-001', category: 'Investigation' })
    expect(sent[1]).toMatchObject({ activityCode: 'SPC-2026-004' })
    expect(sent[1]).not.toHaveProperty('id')
    expect(sent[2]).toMatchObject({ id: 'a3' })
    expect(sent[2]).not.toHaveProperty('activityCode')
    expect(sent[3]).toMatchObject({ id: added.id })
    expect(sent[3]).not.toHaveProperty('activityCode')
    // Table-only fields of filteredPlans are not sent.
    expect(body).not.toHaveProperty('period')
    expect(body).not.toHaveProperty('totalActivity')
    // The list itself was not edited in place (handleEdit deep-copies).
    expect(store.plans[0]?.plannedActivities[0]?.category).toBe('Assurance')
  })

  it('removing a row leaves the others with their own id and code', async () => {
    g.$fetch = backend([storedPlan([
      activity({ id: 'a1', activityCode: 'ASR-2026-001' }),
      activity({ id: 'a2', activityCode: 'ASR-2026-002' }),
      activity({ id: 'a3', activityCode: 'ASR-2026-003' })
    ])])
    const store = useActivityPlanStore()
    await store.fetchPlans()
    store.handleEdit(store.filteredPlans[0] as never)
    store.removePlannedActivity(1)

    await store.savePlan()

    expect(lastWrite().body.plannedActivities!.map(a => [a.id, a.activityCode])).toEqual([
      ['a1', 'ASR-2026-001'],
      ['a3', 'ASR-2026-003']
    ])
  })
})

describe('i18n', () => {
  it('both locales have the Activity ID keys', () => {
    for (const dict of [en, id]) {
      const p = dict.auditActivityPlan
      for (const value of [p.table.activityId, p.form.activityId, p.form.activityIdPending, p.view.columns.activityId]) {
        expect(typeof value).toBe('string')
        expect(value.length).toBeGreaterThan(0)
      }
    }
  })
})

// Minimal stand-ins for the Nuxt UI components.
const slotStub = (name: string) => defineComponent({
  name,
  setup: (_p, { slots }) => () => h('div', { 'data-stub': name }, [slots.header?.(), slots.content?.(), slots.body?.(), slots.default?.(), slots.footer?.()])
})
const nullStub = (name: string) => defineComponent({ name, setup: () => () => null })
const FormFieldStub = defineComponent({
  props: ['label'],
  setup: (p, { slots }) => () => h('div', { 'data-stub': 'UFormField', 'data-label': p.label }, slots.default?.())
})
const ButtonStub = defineComponent({
  props: ['label'],
  setup: (p, { slots }) => () => h('button', { 'data-stub': 'UButton' }, [p.label, slots.default?.()])
})
type Column = { accessorKey: string, header: string, cell?: (ctx: { row: { original: Record<string, unknown> } }) => unknown }
// UTable: a header row, then one row per item using the column's cell renderer when it has one.
const UTableStub = defineComponent({
  props: ['data', 'columns'],
  setup: p => () => h('table', { 'data-stub': 'UTable' }, [
    h('tr', (p.columns as Column[]).map(c => h('th', c.header))),
    ...(p.data as Array<Record<string, unknown>>).map(row => h('tr', { 'data-row': '' },
      (p.columns as Column[]).map(c => h('td', { 'data-col': c.accessorKey }, [c.cell ? c.cell({ row: { original: row } }) as never : String(row[c.accessorKey] ?? '')]))))
  ])
})
// TableEntities: a header row, then each row's `<accessorKey>-cell` slot.
const TableEntitiesStub = defineComponent({
  props: ['data', 'columns'],
  setup: (p, { slots }) => () => h('table', { 'data-stub': 'TableEntities' }, [
    h('tr', (p.columns as Column[]).map(c => h('th', c.header))),
    ...(p.data as Array<Record<string, unknown>>).map(original => h('tr', { 'data-row': '' },
      (p.columns as Column[]).map(c => h('td', { 'data-col': c.accessorKey }, slots[`${c.accessorKey}-cell`]?.({ row: { original } })))))
  ])
})

describe('mounted components', () => {
  let app: App | null = null
  let container: HTMLElement
  afterEach(() => {
    app?.unmount()
    app = null
    document.body.innerHTML = ''
  })

  const mount = async (component: object, items: unknown[], prepare?: (store: ReturnType<typeof useActivityPlanStore>) => void | Promise<void>) => {
    g.$fetch = backend(items)
    const pinia: Pinia = createPinia()
    setActivePinia(pinia)
    const store = useActivityPlanStore()
    await store.fetchPlans()
    await prepare?.(store)
    container = document.createElement('div')
    document.body.appendChild(container)
    app = createApp({ render: () => h(component) })
    app.use(pinia)
    for (const name of ['UModal', 'UForm', 'UCard', 'UTooltip', 'UAlert', 'UBadge']) app.component(name, slotStub(name))
    for (const name of ['UIcon', 'USelectMenu', 'UInput', 'UTextarea', 'UFileUpload', 'AppDatePicker']) app.component(name, nullStub(name))
    app.component('UFormField', FormFieldStub)
    app.component('UButton', ButtonStub)
    app.component('UTable', UTableStub)
    app.component('TableEntities', TableEntitiesStub)
    app.mount(container)
    await flush()
    await nextTick()
    return store
  }

  /** Text of each cell of a column, row by row. */
  const columnTexts = (key: string) =>
    Array.from(container.querySelectorAll(`tr[data-row] td[data-col="${key}"]`)).map(td => td.textContent?.trim())

  it('the form shows the stored code read-only and a placeholder for an unsaved row', async () => {
    const store = await mount(ActivityPlanForm, [storedPlan([
      activity({ id: 'a1', activityCode: 'ASR-2026-001' }),
      activity({ id: 'a2', activityCode: '' })
    ])], (s) => {
      s.handleEdit(s.filteredPlans[0] as never)
    })
    store.addPlannedActivity()
    // Go to step 2 (planned activities) through the stepper header; step 1 is valid.
    const stepHeaders = container.querySelectorAll<HTMLElement>('.cursor-pointer')
    stepHeaders[1]!.click()
    await nextTick()

    const cells = Array.from(container.querySelectorAll('[data-testid="planned-activity-code"]'))
    expect(cells).toHaveLength(3)
    expect(cells[0]!.textContent).toContain(en.auditActivityPlan.form.activityId)
    expect(cells[0]!.textContent).toContain('ASR-2026-001')
    expect(cells[0]!.textContent).not.toContain(en.auditActivityPlan.form.activityIdPending)
    // Stored before codes existed, and a new row: both get their code on save.
    expect(cells[1]!.textContent).toContain(en.auditActivityPlan.form.activityIdPending)
    expect(cells[2]!.textContent).toContain(en.auditActivityPlan.form.activityIdPending)
    // Read-only: no input for the code.
    expect(cells[0]!.querySelector('input')).toBeNull()
  })

  it('a new plan shows the placeholder for its rows', async () => {
    await mount(ActivityPlanForm, [], (s) => {
      s.openModal()
      Object.assign(s.formState, { planTitle: 'New', planPeriodStart: '2026-01-01', planPeriodEnd: '2026-12-31', department: 'IT' })
      s.addPlannedActivity()
    })
    container.querySelectorAll<HTMLElement>('.cursor-pointer')[1]!.click()
    await nextTick()
    const cells = Array.from(container.querySelectorAll('[data-testid="planned-activity-code"]'))
    expect(cells).toHaveLength(1)
    expect(cells[0]!.textContent).toContain(en.auditActivityPlan.form.activityIdPending)
    expect(cells[0]!.textContent).not.toMatch(/[A-Z]{3}-\d{4}-\d{3}/)
  })

  it('the view modal has an Activity ID column', async () => {
    await mount(ActivityPlanViewModal, [storedPlan([
      activity({ id: 'a1', activityCode: 'ASR-2026-001' }),
      activity({ id: 'a2', activityCode: '' })
    ])], (s) => {
      s.openViewModal(s.plans[0]!)
    })
    const headers = Array.from(container.querySelectorAll('[data-stub="UTable"] th')).map(th => th.textContent)
    expect(headers[0]).toBe(en.auditActivityPlan.view.columns.activityId)
    expect(columnTexts('activityCode')).toEqual(['ASR-2026-001', '-'])
  })

  it('the plan table lists the Activity ID of each planned activity', async () => {
    await mount(ActivityPlanTable, [storedPlan([
      activity({ id: 'a1', activityCode: 'ASR-2026-001' }),
      activity({ id: 'a2', category: 'Investigation', activityCode: 'INV-2026-002' }),
      activity({ id: 'a3' })
    ])])
    const headers = Array.from(container.querySelectorAll('[data-stub="TableEntities"] th')).map(th => th.textContent)
    expect(headers).toContain(en.auditActivityPlan.table.activityId)
    const cell = container.querySelector('tr[data-row] td[data-col="activityCode"]')!
    expect(Array.from(cell.querySelectorAll('span')).map(s => s.textContent?.trim())).toEqual(['ASR-2026-001', 'INV-2026-002', '-'])
  })
})
