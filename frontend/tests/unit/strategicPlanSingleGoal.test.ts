/**
 * Strategic Audit Plan with a single corporate goal: the Add/Edit form has no Corporate Goal picker,
 * the table has no goal tabs and always lists every objective, and a new plan's goalId is filled in
 * from the active VMG's (only) goal. An edit keeps the stored goalId; without a VMG/goal it saves "".
 */
import { describe, it, expect, beforeEach, afterEach, vi, type Mock } from 'vitest'
import { readFileSync } from 'fs'
import { resolve } from 'path'
import { createApp, defineComponent, h, nextTick, type App } from 'vue'
import { setActivePinia, createPinia, type Pinia } from 'pinia'
import en from '~/locales/en/common.json'
import id from '~/locales/id/common.json'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useVisionMissionGoalsStore } from '~/stores/vision-mission-goals'
import { defaultStrategicPlanGoalId } from '~/utils/strategicPlanPayload'
import StrategicPlanForm from '~/components/strategic-audit-plan/StrategicPlanForm.vue'
import StrategicPlanTable from '~/components/strategic-audit-plan/StrategicPlanTable.vue'
import type { VisionMissionGoals } from '~/types/master'

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({ showSuccess: vi.fn(), showError: vi.fn(), showWarning: vi.fn(), showInfo: vi.fn() })
}))
// The KPI matrix is not under test here.
vi.mock('~/components/strategic-audit-plan/TargetRealizationMatrix.vue', () => ({ default: { render: () => null } }))

type FetchOpts = { method?: string, body?: Record<string, unknown> }
type FetchFn = (url: string, opts?: FetchOpts) => Promise<unknown>
const g = globalThis as unknown as { $fetch: Mock<FetchFn>, [name: string]: unknown }

const flush = () => new Promise(r => setTimeout(r, 0))
const read = (path: string) => readFileSync(resolve(__dirname, '../../', path), 'utf-8')

const vmg = (goals: Array<{ id?: string, goal_code: string, goal_name: string }>) =>
  ({ id: 'vmg-1', company_id: 'c-1', period: '2026-2030', vision: 'V', mission: 'M', status: 'PUBLISHED', goals }) as unknown as VisionMissionGoals

const plan = (pid: string, extra: Record<string, unknown> = {}) => ({
  id: pid,
  code: `SO-${pid}`,
  goalId: '',
  strategicObjective: `Objective ${pid}`,
  kpi: `KPI ${pid}`,
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
  ...extra
})

/** GET /strategic-plans returns `items`; writes succeed. */
const backend = (items: unknown[]) => vi.fn<FetchFn>(async (_url, opts = {}) => {
  if ((opts.method ?? 'GET').toUpperCase() === 'GET') return { success: true, data: { items } }
  return { success: true }
})
const writes = () => g.$fetch.mock.calls.filter(([, o]) => (o?.method ?? 'GET') !== 'GET')
/** Element of a list that the test expects to exist. */
const nth = <T>(list: T[], index: number): T => {
  const value = list[index]
  if (value === undefined) throw new Error(`no element ${index}`)
  return value
}

beforeEach(() => {
  vi.clearAllMocks()
  g.getAuditServiceBaseUrl = () => '/api/v1'
  g.getMasterServiceBaseUrl = () => '/api/v1'
  g.useGlobalModalStore = () => ({ confirmDelete: vi.fn(async () => true) })
  // Nuxt auto-imports the table uses without an import.
  g.useRbac = () => ({ canManageStrategicPlan: { value: true } })
  g.getFiscalYears = () => [2026, 2027, 2028]
})

describe('source: the goal picker and goal tabs are gone', () => {
  it('the form has no Corporate Goal field and no VMG store', () => {
    const src = read('components/strategic-audit-plan/StrategicPlanForm.vue')
    expect(src).not.toMatch(/corporateGoal|selectCorporateGoal|goalOptions|vmgStore|form\.goalId/)
  })

  it('the table has no goal tabs or goal filter', () => {
    const src = read('components/strategic-audit-plan/StrategicPlanTable.vue')
    expect(src).not.toMatch(/goalTabs|selectedGoalId|allObjectives|vmgStore|\.goalId/)
  })

  it('the removed keys are gone from both locales and nothing references them', () => {
    for (const dict of [en, id] as Array<{ strategicPlan: { form: Record<string, unknown>, filters: Record<string, unknown> } }>) {
      expect(dict.strategicPlan.form).not.toHaveProperty('corporateGoal')
      expect(dict.strategicPlan.form).not.toHaveProperty('selectCorporateGoal')
      expect(dict.strategicPlan.filters).not.toHaveProperty('allObjectives')
    }
  })
})

describe('goalId for a new plan comes from the single corporate goal', () => {
  it('helper: id first, else goal_code, else ""', () => {
    expect(defaultStrategicPlanGoalId([{ id: 'uuid-1', goal_code: 'G-001' }])).toBe('uuid-1')
    expect(defaultStrategicPlanGoalId([{ id: '', goal_code: 'G-001' }])).toBe('G-001')
    expect(defaultStrategicPlanGoalId([])).toBe('')
    expect(defaultStrategicPlanGoalId(undefined)).toBe('')
    expect(defaultStrategicPlanGoalId(null)).toBe('')
  })

  describe('store', () => {
    beforeEach(() => {
      setActivePinia(createPinia())
    })

    const createWith = async (activeVmg: VisionMissionGoals | null) => {
      g.$fetch = backend([plan('a')])
      useVisionMissionGoalsStore().activeVmg = activeVmg
      const store = useStrategicPlanStore()
      await flush()
      store.openModal()
      store.form.strategicObjective = 'New objective'
      await store.handleSubmit()
      expect(writes()).toHaveLength(1)
      const [url, opts] = nth(writes(), 0)
      expect(url).toBe('/api/v1/strategic-plans')
      expect(opts?.method).toBe('POST')
      return { store, body: opts?.body }
    }

    it('a create sends the active goal id', async () => {
      const { body, store } = await createWith(vmg([{ id: 'goal-uuid', goal_code: 'G-001', goal_name: 'Single goal' }]))
      expect(body).toMatchObject({ strategicObjective: 'New objective', goalId: 'goal-uuid' })
      expect(store.isAddModalOpen).toBe(false)
    })

    it('a create falls back to goal_code when the goal has no id', async () => {
      const { body } = await createWith(vmg([{ goal_code: 'G-001', goal_name: 'Single goal' }]))
      expect(body).toMatchObject({ goalId: 'G-001' })
    })

    it('a create without a VMG (or without goals) still saves, with goalId ""', async () => {
      const none = await createWith(null)
      expect(none.body).toMatchObject({ strategicObjective: 'New objective', goalId: '' })
      expect(none.store.isAddModalOpen).toBe(false)

      setActivePinia(createPinia())
      vi.clearAllMocks()
      const empty = await createWith(vmg([]))
      expect(empty.body).toMatchObject({ goalId: '' })
    })

    it('an edit keeps the stored goalId, also when it differs from the active goal or is empty', async () => {
      g.$fetch = backend([plan('a', { goalId: 'G-003' }), plan('b', { goalId: '' })])
      useVisionMissionGoalsStore().activeVmg = vmg([{ id: 'goal-uuid', goal_code: 'G-001', goal_name: 'Single goal' }])
      const store = useStrategicPlanStore()
      await flush()

      store.handleEdit(nth(store.strategicObjectives, 0))
      store.form.kpi = 'Edited'
      await store.handleSubmit()
      store.handleEdit(nth(store.strategicObjectives, 1))
      store.form.kpi = 'Edited too'
      await store.handleSubmit()

      const [first, second] = writes()
      expect(first?.[0]).toBe('/api/v1/strategic-plans/a')
      expect(first?.[1]?.method).toBe('PUT')
      expect(first?.[1]?.body).toMatchObject({ kpi: 'Edited', goalId: 'G-003' })
      expect(second?.[0]).toBe('/api/v1/strategic-plans/b')
      expect(second?.[1]?.body).toMatchObject({ kpi: 'Edited too', goalId: '' })
    })
  })
})

// Minimal stand-ins for the Nuxt UI components; labels are exposed as attributes/text so the DOM can be queried.
const slotStub = (name: string) => defineComponent({
  name,
  setup: (_p, { slots }) => () => h('div', { 'data-stub': name }, [slots.body?.(), slots.default?.(), slots.footer?.()])
})
const FormFieldStub = defineComponent({
  props: ['label'],
  setup: (p, { slots }) => () => h('div', { 'data-stub': 'UFormField', 'data-label': p.label }, slots.default?.())
})
const ButtonStub = defineComponent({
  props: ['label'],
  setup: (p, { slots }) => () => h('button', { 'data-stub': 'UButton' }, [p.label, slots.default?.()])
})
const InputStub = defineComponent({
  props: ['modelValue', 'placeholder'],
  emits: ['update:modelValue'],
  setup: (p, { emit }) => () => h('input', {
    'data-stub': 'UInput',
    'placeholder': p.placeholder,
    'value': p.modelValue,
    'onInput': (e: Event) => emit('update:modelValue', (e.target as HTMLInputElement).value)
  })
})
const SelectMenuStub = defineComponent({
  props: ['modelValue', 'items', 'placeholder'],
  setup: p => () => h('div', { 'data-stub': 'USelectMenu', 'data-placeholder': p.placeholder })
})
// Renders one row per item it is given, with the objective text.
const TableStub = defineComponent({
  props: ['data', 'columns'],
  setup: p => () => h('div', { 'data-stub': 'TableEntities' },
    (p.data as Array<{ id: string, strategicObjective: string }>).map(row => h('div', { 'data-row': row.id }, row.strategicObjective)))
})

describe('mounted components', () => {
  let app: App | null = null
  let container: HTMLElement
  afterEach(() => {
    app?.unmount()
    app = null
    document.body.innerHTML = ''
  })

  const mount = async (component: object, items: unknown[]) => {
    g.$fetch = backend(items)
    const pinia: Pinia = createPinia()
    setActivePinia(pinia)
    // An active VMG with goals: before this change it produced the picker and the tabs.
    useVisionMissionGoalsStore().activeVmg = vmg([
      { id: 'goal-1', goal_code: 'G-001', goal_name: 'Single goal' },
      { id: 'goal-2', goal_code: 'G-002', goal_name: 'Legacy second goal' }
    ])
    container = document.createElement('div')
    document.body.appendChild(container)
    app = createApp({ render: () => h(component) })
    app.use(pinia)
    for (const name of ['UModal', 'UForm', 'UCard', 'UTooltip', 'UAlert', 'URadioGroup', 'UTextarea']) app.component(name, slotStub(name))
    app.component('UFormField', FormFieldStub)
    app.component('UButton', ButtonStub)
    app.component('UInput', InputStub)
    app.component('USelectMenu', SelectMenuStub)
    app.component('TableEntities', TableStub)
    app.mount(container)
    await flush()
    await nextTick()
  }

  it('the form renders its fields but no Corporate Goal field', async () => {
    await mount(StrategicPlanForm, [])
    const labels = Array.from(container.querySelectorAll('[data-stub="UFormField"]')).map(el => el.getAttribute('data-label'))
    expect(labels).toContain(en.strategicPlan.form.objective)
    expect(labels).toContain(en.strategicPlan.form.kpi)
    expect(labels).not.toContain('Corporate Goal')
    expect(labels).not.toContain('strategicPlan.form.corporateGoal')
    expect(container.textContent).not.toMatch(/Corporate Goal|G-001/)
    const placeholders = Array.from(container.querySelectorAll('[data-stub="USelectMenu"]')).map(el => el.getAttribute('data-placeholder'))
    expect(placeholders).not.toContain('Select Corporate Goal')
    expect(placeholders).not.toContain('strategicPlan.form.selectCorporateGoal')
  })

  it('the table has no goal tabs and lists every objective, whatever its goalId', async () => {
    await mount(StrategicPlanTable, [
      plan('a', { goalId: 'goal-1' }),
      plan('b', { goalId: 'G-002' }),
      plan('c', { goalId: '' }),
      plan('d', { goalId: 'unknown-goal' })
    ])
    const buttons = Array.from(container.querySelectorAll('[data-stub="UButton"]')).map(el => el.textContent ?? '')
    expect(buttons.some(text => /G-00\d|All Strategic|strategicPlan\.filters\.allObjectives/.test(text))).toBe(false)
    expect(Array.from(container.querySelectorAll('[data-row]')).map(el => el.getAttribute('data-row'))).toEqual(['a', 'b', 'c', 'd'])
  })

  it('search still filters the table', async () => {
    await mount(StrategicPlanTable, [plan('a'), plan('b', { kpi: 'Findings closed on time' }), plan('c')])
    const search = container.querySelector(`input[placeholder="${en.strategicPlan.filters.searchPlaceholder}"]`) as HTMLInputElement
    expect(search).not.toBeNull()
    search.value = 'findings closed'
    search.dispatchEvent(new Event('input'))
    await nextTick()
    expect(Array.from(container.querySelectorAll('[data-row]')).map(el => el.getAttribute('data-row'))).toEqual(['b'])
  })
})
