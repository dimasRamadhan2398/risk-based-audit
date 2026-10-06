// Pure view-model helpers for the KPI summary cards and charts on /kpi-performance.
// Only what the API returned is shown: with no data a card reads "-" and a chart has no series.

import { kpiSubMetric, type KpiSubMetric, type TranslateFn } from '~/utils/kpiPerformanceLabels'

/** Card as sent by GET /performance/dashboard-summary (subset used here). */
export interface KpiSummaryCardInput {
  title: string
  key: string
  value: string
  target: string
  gap: string
  trend: string
  trend_up: boolean
  sub_metrics?: KpiSubMetric[]
}

export interface KpiSummaryCardView {
  key: string
  title: string
  value: string
  target: string
  /** "" when there is nothing to compare. */
  trend: string
  /** null when there is no data (neutral colour). */
  trendUp: boolean | null
  subMetrics: KpiSubMetric[]
  hasData: boolean
}

/** The four cards the page always lays out, in this order, before/without API data. */
export const KPI_SUMMARY_CARD_KEYS = ['audit_completion_rate', 'report_timeliness', 'client_satisfaction', 'action_plan_closed'] as const

const NO_VALUE = '-'

const translateOr = (t: TranslateFn, path: string, fallback: string) => {
  const label = t(path)
  return label === path ? fallback : label
}

const textOrDash = (value: unknown) => (typeof value === 'string' && value.trim() ? value : NO_VALUE)

/**
 * Cards to render: the API's cards with translated titles/sub-metrics, or, when the API returned
 * none (not loaded, empty or failed), the four card titles with "-" and no trend or sub-metrics.
 */
export const buildKpiSummaryCards = (t: TranslateFn, apiCards: KpiSummaryCardInput[] | null | undefined): KpiSummaryCardView[] => {
  if (Array.isArray(apiCards) && apiCards.length > 0) {
    return apiCards.map(c => ({
      key: c.key,
      title: translateOr(t, `kpiPerformance.summary.cards.${c.key}`, c.title),
      value: textOrDash(c.value),
      target: textOrDash(c.target),
      // `gap` is the signed delta (e.g. "+2.0%"); fall back to the API's own text if it is missing.
      trend: c.gap ? t('kpiPerformance.summary.vsTarget', { gap: c.gap }) : (c.trend || ''),
      trendUp: typeof c.trend_up === 'boolean' ? c.trend_up : null,
      subMetrics: (c.sub_metrics ?? []).map(sub => kpiSubMetric(t, sub)),
      hasData: true
    }))
  }
  return KPI_SUMMARY_CARD_KEYS.map(key => ({
    key,
    title: translateOr(t, `kpiPerformance.summary.cards.${key}`, key),
    value: NO_VALUE,
    target: NO_VALUE,
    trend: '',
    trendUp: null,
    subMetrics: [],
    hasData: false
  }))
}

/** Monthly series as sent by GET /performance/monthly-trends (subset used here). */
export interface KpiMonthlyTrendsInput {
  labels?: unknown
  completion_rate_series?: unknown
  timeliness_series?: unknown
  csat_series?: unknown
}

export interface KpiChartSeries {
  labels: string[]
  completion: Array<number | null>
  timeliness: Array<number | null>
  csat: Array<number | null>
}

// Non-numeric points become null (a gap in the chart) rather than 0.
const numberSeries = (value: unknown): Array<number | null> =>
  Array.isArray(value) ? value.map(v => (v === null || v === '' || !Number.isFinite(Number(v)) ? null : Number(v))) : []

/** The API's series as-is; anything missing is an empty series, never a generated one. */
export const kpiChartSeries = (trends: KpiMonthlyTrendsInput | null | undefined): KpiChartSeries => ({
  labels: Array.isArray(trends?.labels) ? trends.labels.map(String) : [],
  completion: numberSeries(trends?.completion_rate_series),
  timeliness: numberSeries(trends?.timeliness_series),
  csat: numberSeries(trends?.csat_series)
})

/** A series has something to draw only if at least one point is a number. */
export const hasSeriesData = (...series: Array<Array<number | null>>): boolean =>
  series.some(s => s.some(v => v !== null))
