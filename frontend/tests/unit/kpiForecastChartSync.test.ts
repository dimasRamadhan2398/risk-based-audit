// @ts-nocheck
/**
 * KPI Performance Forecast chart sync.
 *
 * The dashboard renders a "KPI Performance Forecasting" line chart
 * (pages/dashboard/index.vue) and the AI Insights tab renders the same chart
 * (pages/analytics/ai/kpi-forecast.vue). Both are supposed to show the same
 * series. These tests pin down where the two diverge.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { useTimeSeriesData } from '~/composables/useAnalyticsData'
import { useAiAnalytics } from '~/composables/useAiAnalytics'

const DASHBOARD_PAGE = resolve(__dirname, '../../pages/dashboard/index.vue')
const TAB_PAGE = resolve(__dirname, '../../pages/analytics/ai/kpi-forecast.vue')

/** Shape of /performance-trend/batch as returned by backend/python-ai/main.py */
const backendTimeSeries = [
  { period: 'Q1 2025', actual: 85.2, forecast: 85.2, upperBound: 88.0, lowerBound: 82.0 },
  { period: 'Q2 2025', actual: 82.0, forecast: 82.0, upperBound: 85.0, lowerBound: 79.0 },
  { period: 'Q3 2025', actual: 79.5, forecast: 79.5, upperBound: 82.5, lowerBound: 76.5 },
  { period: 'Q4 2025', actual: 76.0, forecast: 76.0, upperBound: 79.0, lowerBound: 73.0 },
  { period: 'Q1 2026 (Pred)', actual: null, forecast: 73.5, upperBound: 77.0, lowerBound: 70.0 },
  { period: 'Q2 2026 (Pred)', actual: null, forecast: 71.2, upperBound: 75.0, lowerBound: 67.5 }
]

const performanceTrendPayload = {
  kpi_forecasts: [
    {
      kpiName: 'Revenue Operational Cost',
      code: 'KPI-001',
      unit: '%',
      entity: 'Finance Dept',
      entityType: 'Department',
      targetHorizon: 'Q3 2026',
      currentValue: 41.2,
      forecastedValue: 37.9,
      trend: 'Deteriorating',
      recommendedAction: 'Audit biaya operasional',
      riskLevel: 'HIGH'
    }
  ],
  at_risk_departments: [
    { department: 'Finance Dept', kpi: 'Revenue Operational Cost', currentTrend: -3.5, predictedQ3: 37.9, riskLevel: 'HIGH' }
  ],
  time_series_data: backendTimeSeries,
  forecast_accuracy: { mape: 3.82, rmse: 1.45, r2Score: 0.942 }
}

/** Replica of the dataset builder in pages/analytics/ai/kpi-forecast.vue */
const buildTabChartData = (points: any[]) => ({
  labels: (points || []).map((p: any) => p.period || ''),
  actual: (points || []).map((p: any) => p.actual ?? null),
  forecast: (points || []).map((p: any) => p.forecast ?? null),
  upperBound: (points || []).map((p: any) => p.upperBound ?? null),
  lowerBound: (points || []).map((p: any) => p.lowerBound ?? null)
})

/** Replica of the dataset builder in pages/dashboard/index.vue */
const buildDashboardChartData = (points: any[]) => ({
  labels: points.map((p: any) => p.period),
  actual: points.map((p: any) => p.actual),
  forecast: points.map((p: any) => p.forecast),
  upperBound: points.map((p: any) => p.upperBound),
  lowerBound: points.map((p: any) => p.lowerBound)
})

describe('KPI Performance Forecast chart — dashboard vs AI Insights tab', () => {
  beforeEach(() => {
    localStorage.clear()
    global.$fetch = vi.fn(async (url: string) => {
      if (String(url).includes('performance-trend/batch')) return performanceTrendPayload
      return null
    })
  })

  it('both charts read from the same source of truth', async () => {
    const { timeseriesState, fetchAiAnalytics } = useAiAnalytics()
    await fetchAiAnalytics(true)

    // Both pages now map `useAiAnalytics().timeseriesState.historicalKPI`
    const points = timeseriesState.value.historicalKPI

    expect(buildDashboardChartData(points)).toEqual(buildTabChartData(points))
    // ...and that shared state is the backend series, not the static seed mock
    expect(buildDashboardChartData(points).labels).toEqual(backendTimeSeries.map((p) => p.period))
    expect(buildDashboardChartData(points)).not.toEqual(
      buildDashboardChartData(useTimeSeriesData().historicalKPI)
    )
  })

  it('the AI Insights tab picks up backend time_series_data with camelCase bound keys', async () => {
    const { timeseriesState, fetchAiAnalytics } = useAiAnalytics()
    await fetchAiAnalytics(true)

    const chart = buildTabChartData(timeseriesState.value.historicalKPI)
    expect(chart.labels).toEqual(backendTimeSeries.map((p) => p.period))
    expect(chart.upperBound).toEqual(backendTimeSeries.map((p) => p.upperBound))
    expect(chart.lowerBound).toEqual(backendTimeSeries.map((p) => p.lowerBound))
  })

  it('the dashboard chart is wired to the shared AI analytics state, not the static mock', () => {
    const src = readFileSync(DASHBOARD_PAGE, 'utf-8')
    const chartBlock = src.slice(src.indexOf('const kpiChartData'), src.indexOf('const kpiChartOptions'))

    expect(src.includes('useAiAnalytics'), 'dashboard/index.vue does not import useAiAnalytics').toBe(true)
    expect(
      /useTimeSeriesData/.test(src),
      'dashboard/index.vue still reads the static useTimeSeriesData() mock'
    ).toBe(false)
    expect(
      /historicalKPI/.test(chartBlock),
      'kpiChartData no longer maps any historicalKPI series'
    ).toBe(true)
  })

  it('the dashboard fetches forecast data so its chart can refresh', () => {
    const src = readFileSync(DASHBOARD_PAGE, 'utf-8')
    expect(src.includes('fetchAiAnalytics'), 'dashboard/index.vue never calls fetchAiAnalytics').toBe(true)
  })

  it('both charts label their series identically', () => {
    const dashboardSrc = readFileSync(DASHBOARD_PAGE, 'utf-8')
    const tabSrc = readFileSync(TAB_PAGE, 'utf-8')

    const labelKeys = [
      'analytics.timeseries.labelActualKPI',
      'analytics.timeseries.labelForecast',
      'analytics.timeseries.labelUpperBound',
      'analytics.timeseries.labelLowerBound'
    ]
    const missingInTab = labelKeys.filter((k) => !tabSrc.includes(k))
    const missingInDashboard = labelKeys.filter((k) => !dashboardSrc.includes(k))

    expect(missingInTab, 'i18n label keys missing in the AI Insights tab chart').toEqual([])
    expect(missingInDashboard, 'i18n label keys missing in the dashboard chart (labels hardcoded)').toEqual([])
  })
})
