// @ts-nocheck
/**
 * KPI Performance Forecast — dashboard card vs /analytics/ai/kpi-forecast.
 *
 * Both render the chart and the at-risk list through the same view model
 * (composables/useKpiForecast.ts) from the same shared state
 * (useAiAnalytics().timeseriesState.kpiForecasts).
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { useAiAnalytics } from '~/composables/useAiAnalytics'
import {
  buildKpiForecastChart,
  selectAtRiskKpis,
  toKpiForecastView
} from '~/composables/useKpiForecast'

const DASHBOARD_PAGE = resolve(__dirname, '../../pages/dashboard/index.vue')
const TAB_PAGE = resolve(__dirname, '../../pages/analytics/ai/kpi-forecast.vue')

const t = (key: string, params?: any) => (params ? `${key}:${JSON.stringify(params)}` : key)

/** Shape of /performance-trend/batch as returned by backend/python-ai/main.py */
const performanceTrendPayload = {
  kpi_forecasts: [
    {
      kpiName: 'Net Promoter Score', code: 'KPI-001', unit: 'score', entity: 'Jakarta Branch', entityType: 'Branch',
      targetHorizon: '2026-01..2026-03', currentValue: 47.7, forecastedValue: 47.17, trend: 'Deteriorating',
      recommendedAction: 'Survey NPS', riskLevel: 'MODERATE', lastActualPeriod: '2025-12',
      history: [{ period: '2025-11', value: 47.0 }, { period: '2025-12', value: 47.7 }],
      forecast: [{ period: '2026-01', value: 47.17 }, { period: '2026-02', value: 46.64 }, { period: '2026-03', value: 46.11 }]
    },
    {
      kpiName: 'Customer Retention Rate', code: 'KPI-006', unit: '%', entity: 'Bandung Branch', entityType: 'Branch',
      targetHorizon: '2026-01..2026-03', currentValue: 91.7, forecastedValue: 92.1, trend: 'Improving',
      recommendedAction: 'NPS membaik — teruskan inisiatif', riskLevel: 'LOW', lastActualPeriod: '2025-12',
      history: [{ period: '2025-11', value: 91.5 }, { period: '2025-12', value: 91.7 }],
      forecast: [{ period: '2026-01', value: 92.1 }, { period: '2026-02', value: 92.5 }, { period: '2026-03', value: 92.87 }]
    },
    {
      // Lower is better: a rising turnover is a deterioration even though the number goes up.
      kpiName: 'Employee Turnover Rate', code: 'KPI-007', unit: '%', entity: 'HR Dept', entityType: 'Department',
      targetHorizon: '2026-01..2026-03', currentValue: 12.0, forecastedValue: 12.6, trend: 'Improving',
      recommendedAction: '', riskLevel: 'MODERATE_HIGH', lastActualPeriod: '2025-12',
      history: [{ period: '2025-12', value: 12.0 }],
      forecast: [{ period: '2026-01', value: 12.6 }, { period: '2026-02', value: 13.1 }, { period: '2026-03', value: 13.5 }]
    }
  ],
  at_risk_departments: [],
  time_series_data: [],
  forecast_accuracy: { mape: 3.82, rmse: 1.45, r2Score: 0.942 }
}

describe('KPI forecast view model', () => {
  beforeEach(() => {
    localStorage.clear()
    global.$fetch = vi.fn(async (url: string) => {
      if (String(url).includes('performance-trend/batch')) return performanceTrendPayload
      return null
    })
  })

  it('keeps per-KPI history, forecast and the model accuracy from the backend', async () => {
    const { timeseriesState, fetchAiAnalytics } = useAiAnalytics()
    await fetchAiAnalytics(true)
    const nps = timeseriesState.value.kpiForecasts[0]
    expect(nps.history).toHaveLength(2)
    expect(nps.forecast.map((p) => p.period)).toEqual(['2026-01', '2026-02', '2026-03'])
    expect(nps.lastActualPeriod).toBe('2025-12')
    expect(timeseriesState.value.forecastAccuracy.mape).toBe(3.82)
  })

  it('the horizon follows the last actual month, and a past horizon is flagged stale', () => {
    const view = toKpiForecastView(performanceTrendPayload.kpi_forecasts[0])
    expect(view.horizonStart).toBe('2026-01')
    expect(view.horizonEnd).toBe('2026-03')
    expect(view.isStale).toBe(true) // March 2026 is before "now" (Nov 2026)
  })

  it('at-risk lists only deteriorating KPIs, judged by KPI direction', () => {
    const views = performanceTrendPayload.kpi_forecasts.map(toKpiForecastView)
    const atRisk = selectAtRiskKpis(views).map((v) => v.kpiName)
    expect(atRisk).toEqual(['Employee Turnover Rate', 'Net Promoter Score'])
    expect(atRisk).not.toContain('Customer Retention Rate')
  })

  it('the band starts at the last actual point with zero width and only covers forecast points', () => {
    const view = toKpiForecastView(performanceTrendPayload.kpi_forecasts[0])
    const { data } = buildKpiForecastChart(view, 3.82, t, 'id')
    const [actual, forecast, upper, lower] = data.datasets.map((d) => d.data)
    // 2 actual + 3 forecast periods
    expect(data.labels).toEqual(['Nov 2025', 'Des 2025', 'Jan 2026', 'Feb 2026', 'Mar 2026'])
    expect(actual).toEqual([47.0, 47.7, null, null, null])
    expect(forecast).toEqual([null, 47.7, 47.17, 46.64, 46.11])
    expect(upper[0]).toBeNull()
    expect(upper[1]).toBe(47.7)
    expect(lower[1]).toBe(47.7)
    expect(upper[2]).toBeGreaterThan(47.17)
    expect(lower[2]).toBeLessThan(47.17)
    // widens with the horizon
    expect(upper[4] - lower[4]).toBeGreaterThan(upper[2] - lower[2])
  })

  it('names the KPI on the y axis instead of a generic "KPI score (%)"', () => {
    const view = toKpiForecastView(performanceTrendPayload.kpi_forecasts[0])
    const { options } = buildKpiForecastChart(view, 3.82, t, 'id')
    expect(options.scales.y.title.text).toContain('analytics.kpiAudit.axis.score')
    expect(options.scales.y.title.text).toContain('Net Promoter Score')
  })
})

describe('dashboard card and AI Insights tab share the view model', () => {
  const dashboardSrc = readFileSync(DASHBOARD_PAGE, 'utf-8')
  const tabSrc = readFileSync(TAB_PAGE, 'utf-8')

  it.each([
    ['dashboard', dashboardSrc],
    ['AI Insights tab', tabSrc]
  ])('%s builds the chart and at-risk list through useKpiForecast', (_, src) => {
    expect(src).toContain('buildKpiForecastChart')
    expect(src).toContain('selectAtRiskKpis')
    expect(src).toContain('toKpiForecastView')
    expect(src).toContain('fetchAiAnalytics')
    expect(src).not.toMatch(/useTimeSeriesData/)
  })

  it('no hardcoded horizon, confidence level or MAPE remains in the dashboard card', () => {
    expect(dashboardSrc).not.toMatch(/Q3 2026/)
    expect(dashboardSrc).not.toMatch(/95%/)
    expect(dashboardSrc).not.toMatch(/4[.,]8%/)
  })
})
