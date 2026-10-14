/**
 * KPI detailed breakdown: after a KPI target is saved through the strategic plan form, the table
 * reloads and renders the new row for the year (store -> component -> UTable rows).
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { createApp, defineComponent, h, nextTick } from 'vue'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('~/composables/useApiUrl', () => ({
  getAuditServiceBaseUrl: () => 'http://api.test/api/v1',
  getMasterServiceBaseUrl: () => 'http://api.test/api/v1',
  getRiskServiceBaseUrl: () => 'http://api.test/api/v1',
  getAuthServiceBaseUrl: () => 'http://api.test/api/v1',
  getAnalyticsServiceBaseUrl: () => 'http://api.test/api/analytics',
  getPythonAiBaseUrl: () => 'http://api.test/api/ai'
}))

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({ showSuccess: vi.fn(), showError: vi.fn(), showWarning: vi.fn(), showInfo: vi.fn() })
}))

vi.mock('~/composables/useRbac', () => ({
  useRbac: () => ({ canManageStrategicPlan: { value: true } })
}))

vi.mock('~/components/strategic-audit-plan/StrategicPlanForm.vue', () => ({
  default: defineComponent({ name: 'StrategicPlanForm', render: () => h('div', { 'data-stub': 'form' }) })
}))

// Stand-in for UTable: one <tr> per data row, so the test sees exactly what the table would render.
const UTableStub = defineComponent({
  name: 'UTable',
  props: ['data', 'columns', 'loading'],
  setup: (p, { slots }) => () => h('table', [
    h('tbody', p.data && p.data.length
      ? p.data.map(row => h('tr', { 'data-row': row.id }, [h('td', row.metric)]))
      : [h('tr', [h('td', { 'data-empty': '' }, slots.empty ? slots.empty() : 'empty')])])
  ])
})

const passthrough = name => defineComponent({ name, setup: (_, { slots }) => () => h('div', slots.default ? slots.default() : []) })

const planRow = {
  id: 'plan-1',
  source: 'strategic_plan',
  metric: 'Audit coverage',
  category: 'Operational',
  period: '2026',
  unit: '%',
  target: 90,
  actual: 0,
  gap: -90,
  gapIsPositive: false,
  achievementRate: 0,
  status: 'Needs Attention'
}

const flush = async () => {
  for (let i = 0; i < 10; i++) {
    await new Promise(resolve => setTimeout(resolve, 0))
    await nextTick()
  }
}

const mountApp = (component, props, pinia, stubs) => {
  const el = document.createElement('div')
  document.body.appendChild(el)
  const app = createApp(component, props)
  app.use(pinia)
  for (const [name, comp] of Object.entries(stubs)) app.component(name, comp)
  app.mount(el)
  return { el, unmount: () => app.unmount() }
}

describe('KPI detailed breakdown reloads after saving a target', () => {
  // Auto-imported in the store (no explicit import), so it must be a global here.
  globalThis.getAuditServiceBaseUrl = () => 'http://api.test/api/v1'
  let breakdownItems
  let breakdownCalls

  beforeEach(() => {
    setActivePinia(createPinia())
    breakdownItems = []
    breakdownCalls = []
    globalThis.$fetch = vi.fn(async (url, opts = {}) => {
      if (String(url).endsWith('/performance/kpi-breakdown')) {
        breakdownCalls.push(opts.params)
        return {
          success: true,
          data: {
            items: breakdownItems,
            pagination: { page: 1, page_size: 10, total: breakdownItems.length, total_pages: breakdownItems.length ? 1 : 0 },
            filters: { categories: [], periods: [] }
          }
        }
      }
      if (String(url).endsWith('/strategic-plans') && opts.method === 'POST') {
        // The row exists on the server from here on.
        breakdownItems = [planRow]
        return { success: true, data: { id: 'plan-1' } }
      }
      if (String(url).endsWith('/strategic-plans')) return { success: true, data: { items: [] } }
      throw new Error(`unexpected request ${url}`)
    })
  })

  it('renders the saved row after the form closes', async () => {
    const { default: KpiDetailedTable } = await import('~/components/kpi-performance/KpiDetailedTable.vue')
    const { useStrategicPlanStore } = await import('~/stores/strategic-audit-plan')

    const pinia = createPinia()
    setActivePinia(pinia)
    const stubs = {
      UTable: UTableStub,
      UButton: passthrough('UButton'),
      UInput: passthrough('UInput'),
      USelectMenu: passthrough('USelectMenu'),
      USelect: passthrough('USelect'),
      UPagination: passthrough('UPagination'),
      UTooltip: passthrough('UTooltip'),
      UIcon: passthrough('UIcon')
    }
    const wrapper = mountApp(KpiDetailedTable, { year: 2026 }, pinia, stubs)
    await flush()
    expect(wrapper.el.querySelector('[data-empty]')).not.toBeNull()

    const store = useStrategicPlanStore(pinia)
    store.openModal()
    store.form.strategicObjective = 'Improve audit coverage'
    store.form.kpi = 'Audit coverage'
    store.form.unit = '%'
    store.form.kpiTargets = { 2026: '90' }
    await store.handleSubmit()
    await flush()

    expect(breakdownCalls.length).toBeGreaterThanOrEqual(2)
    expect(breakdownCalls[breakdownCalls.length - 1]).toMatchObject({ year: 2026, page: 1 })
    expect(wrapper.el.querySelector('[data-row="plan-1"]')).not.toBeNull()
    expect(wrapper.el.querySelector('[data-empty]')).toBeNull()
    wrapper.unmount()
  })
})
