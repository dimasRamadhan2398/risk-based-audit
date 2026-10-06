/**
 * /kpi-performance summary cards and charts show only API data: with nothing from the API a card
 * reads "-" and a chart shows an empty state, never sample numbers or series scaled from a guess.
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { readFileSync } from 'fs'
import { resolve } from 'path'
import { createApp, defineComponent, h, nextTick, type App } from 'vue'
import { setActivePinia, createPinia } from 'pinia'
import en from '~/locales/en/common.json'
import id from '~/locales/id/common.json'
import type { TranslateFn } from '~/utils/kpiPerformanceLabels'
import { buildKpiSummaryCards, hasSeriesData, kpiChartSeries, KPI_SUMMARY_CARD_KEYS } from '~/utils/kpiPerformanceDisplay'
import { usePerformanceStore } from '~/stores/performance'
import KpiSummaryCards from '~/components/kpi-performance/KpiSummaryCards.vue'
import KpiCharts from '~/components/kpi-performance/KpiCharts.vue'

// Charts render as stubs that expose the data they were given.
vi.mock('vue-chartjs', () => {
  const chart = (name: string) => ({
    name,
    props: ['data', 'options'],
    setup: (p: { data: unknown }) => () => h('div', { 'data-chart': name, 'data-chart-data': JSON.stringify(p.data) })
  })
  return { Bar: chart('Bar'), Line: chart('Line') }
})

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

const g = globalThis as unknown as { $fetch: unknown, [name: string]: unknown }
const flush = () => new Promise(r => setTimeout(r, 0))

const apiCard = {
  title: 'Audit Completion Rate',
  key: 'audit_completion_rate',
  value: '75.0%',
  target: '90%',
  actual_number: 75,
  target_number: 90,
  gap: '-15.0%',
  trend: '',
  trend_up: false,
  unit: '%',
  sub_metrics: [{ title: 'Survey Response Rate', value: '50.0%', target: '80%', trend: '3 responses from 6 completed audits' }]
}

describe('summary cards view model', () => {
  it('no API cards: the four titles with "-", no target, trend or sub-metric numbers', () => {
    for (const input of [[], null, undefined]) {
      const cards = buildKpiSummaryCards(tEn, input)
      expect(cards.map(c => c.key)).toEqual([...KPI_SUMMARY_CARD_KEYS])
      for (const card of cards) {
        expect(card).toMatchObject({ value: '-', target: '-', trend: '', trendUp: null, subMetrics: [], hasData: false })
      }
      expect(JSON.stringify(cards)).not.toMatch(/\d/)
    }
    expect(buildKpiSummaryCards(tId, []).map(c => c.title)).toEqual(KPI_SUMMARY_CARD_KEYS.map(k => id.kpiPerformance.summary.cards[k]))
  })

  it('API cards are shown as returned, with translated titles and sub-metric texts', () => {
    const [card] = buildKpiSummaryCards(tId, [apiCard])
    expect(card).toMatchObject({
      key: 'audit_completion_rate',
      title: id.kpiPerformance.summary.cards.audit_completion_rate,
      value: '75.0%',
      target: '90%',
      trendUp: false,
      hasData: true
    })
    expect(card?.trend).toBe('-15.0% terhadap target')
    expect(card?.subMetrics).toEqual([{
      title: id.kpiPerformance.summary.subTitles.surveyResponseRate,
      value: '50.0%',
      target: '80%',
      trend: tId('kpiPerformance.summary.text.responsesFrom', { responses: 3, audits: 6 })
    }])
  })

  it('a blank value from the API is "-", not a number', () => {
    const [card] = buildKpiSummaryCards(tEn, [{ ...apiCard, value: '', target: '  ', gap: '', trend: '' }])
    expect(card).toMatchObject({ value: '-', target: '-', trend: '' })
  })
})

describe('chart series', () => {
  it('no trends: empty series and nothing to draw', () => {
    for (const input of [null, undefined, {}]) {
      const s = kpiChartSeries(input)
      expect(s).toEqual({ labels: [], completion: [], timeliness: [], csat: [] })
      expect(hasSeriesData(s.completion)).toBe(false)
      expect(hasSeriesData(s.timeliness, s.csat)).toBe(false)
    }
  })

  it('uses the API series as-is; a missing series stays empty instead of being generated', () => {
    const s = kpiChartSeries({ labels: ['Jan', 'Feb'], completion_rate_series: [40, 'x'], timeliness_series: [0, 55] })
    expect(s).toEqual({ labels: ['Jan', 'Feb'], completion: [40, null], timeliness: [0, 55], csat: [] })
    expect(hasSeriesData(s.completion)).toBe(true)
    expect(hasSeriesData(s.csat)).toBe(false)
    // Zeros are real values from the API.
    expect(hasSeriesData([0, 0])).toBe(true)
    expect(hasSeriesData([null, null])).toBe(false)
  })
})

describe('performance store: no stale or sample summary data', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('an empty or failed dashboard-summary / monthly-trends response leaves nothing to show', async () => {
    const store = usePerformanceStore()
    g.$fetch = vi.fn(async (url: string) => url.includes('dashboard-summary')
      ? { success: true, data: [apiCard] }
      : { success: true, data: { labels: ['Jan'], completion_rate_series: [10], timeliness_series: [20], csat_series: [3] } })
    await store.fetchDashboardSummary(2026)
    await store.fetchMonthlyTrends(2026)
    expect(store.dashboardCards).toHaveLength(1)
    expect(store.monthlyTrends).not.toBeNull()

    // Another year with no data must not keep showing the previous year's numbers.
    g.$fetch = vi.fn(async () => ({ success: true, data: null }))
    await store.fetchDashboardSummary(2025)
    await store.fetchMonthlyTrends(2025)
    expect(store.dashboardCards).toEqual([])
    expect(store.monthlyTrends).toBeNull()

    await store.fetchDashboardSummary(2026)
    await store.fetchMonthlyTrends(2026)
    g.$fetch = vi.fn(async () => {
      throw new Error('down')
    })
    await store.fetchDashboardSummary(2026)
    await store.fetchMonthlyTrends(2026)
    expect(store.dashboardCards).toEqual([])
    expect(store.monthlyTrends).toBeNull()
    expect(store.error).toBeTruthy()
  })
})

const UIStub = (name: string) => defineComponent({ name, setup: (_p, { slots }) => () => h('div', { 'data-stub': name }, slots.default?.()) })

describe('cards and charts rendered with no API data', () => {
  let app: App | null = null
  let container: HTMLElement
  beforeEach(() => {
    setActivePinia(createPinia())
  })
  afterEach(() => {
    app?.unmount()
    app = null
    document.body.innerHTML = ''
  })

  const mount = async (component: object, setup?: (store: ReturnType<typeof usePerformanceStore>) => void) => {
    container = document.createElement('div')
    document.body.appendChild(container)
    const pinia = createPinia()
    setActivePinia(pinia)
    setup?.(usePerformanceStore())
    app = createApp({ render: () => h(component, { year: 2026 }) })
    app.use(pinia)
    for (const name of ['UCard', 'UIcon', 'UTooltip']) app.component(name, UIStub(name))
    app.mount(container)
    await flush()
    await nextTick()
    return container
  }

  it('summary cards show "-" and "no data", without any number', async () => {
    const el = await mount(KpiSummaryCards)
    expect(el.querySelectorAll('[data-stub="UCard"]')).toHaveLength(4)
    expect(el.textContent).toContain(en.kpiPerformance.summary.noData)
    expect(el.textContent).not.toMatch(/\d/)
  })

  it('summary cards show the API values when present', async () => {
    const el = await mount(KpiSummaryCards, (store) => {
      store.dashboardCards = [apiCard]
    })
    expect(el.querySelectorAll('[data-stub="UCard"]')).toHaveLength(1)
    expect(el.textContent).toContain('75.0%')
    expect(el.textContent).not.toContain(en.kpiPerformance.summary.noData)
  })

  it('charts show the empty state instead of a series', async () => {
    const el = await mount(KpiCharts)
    expect(el.querySelectorAll('[data-chart]')).toHaveLength(0)
    expect(el.textContent).toContain(tEn('kpiPerformance.charts.noData', { year: 2026 }))
  })

  it('charts draw exactly the API series', async () => {
    const el = await mount(KpiCharts, (store) => {
      store.monthlyTrends = { labels: ['Jan', 'Feb'], completion_rate_series: [40, 60], timeliness_series: [70, 80], csat_series: [3.5, 4] }
    })
    const data = (name: string) => JSON.parse(el.querySelector(`[data-chart="${name}"]`)?.getAttribute('data-chart-data') || 'null')
    expect(data('Bar').datasets.map((d: { data: number[] }) => d.data)).toEqual([[40, 60]])
    expect(data('Line').datasets.map((d: { data: number[] }) => d.data)).toEqual([[70, 80], [3.5, 4]])
  })

  it('the components hold no sample numbers or strategic-plan guesses', () => {
    for (const file of ['KpiSummaryCards.vue', 'KpiCharts.vue']) {
      const src = readFileSync(resolve(__dirname, `../../components/kpi-performance/${file}`), 'utf8')
      expect(src).not.toMatch(/findSpMetric|strategicObjectives|useStrategicPlanStore/)
      expect(src).not.toMatch(/'(92|98|87|90)%'|4\.0 \/ 5\.0|\|\| '(95|98|4\.7|92|87|4\.0)'/)
      expect(src).not.toMatch(/\[\s*85,\s*90|\* 0\.\d\d/)
    }
  })
})
