/**
 * KPI forecast view model shared by the dashboard card and
 * /analytics/ai/kpi-forecast, so both show the same KPI, horizon, units,
 * at-risk list and uncertainty band.
 *
 * Horizon: the forecast periods always follow the last actual data point of
 * the KPI itself. The UI states "data up to <month>" and flags the forecast as
 * stale when its horizon is already in the past, instead of labelling a
 * finished quarter as a prediction.
 *
 * Band: the shaded area starts at the last actual point (zero width) and widens
 * over the forecast steps by the model's historical error (MAPE × √step). It is
 * labelled as an error-based range, not as a 95% confidence interval, and MAPE
 * itself is shown as the model's accuracy.
 */
import {
  evaluateKpiTrend,
  formatKpiForecastValue,
  formatForecastPeriod,
  formatForecastPeriodRange,
  getKpiMeta,
  isPeriodPast,
  type AppLocale,
  type KpiMeta,
  type KpiTrend,
} from '~/utils/analyticsFormat'

export interface KpiPoint {
  period: string
  value: number
}

export interface KpiForecastView {
  kpiName: string
  code: string
  entity: string
  entityType: string
  riskLevel: string
  recommendedAction: string
  currentValue: number
  forecastedValue: number
  meta: KpiMeta
  trend: KpiTrend
  /** Relative change from current to the end of the horizon, signed so that negative = worse. */
  changeTowardsBetter: number
  history: KpiPoint[]
  forecast: KpiPoint[]
  lastActualPeriod: string
  horizonStart: string
  horizonEnd: string
  /** The whole forecast horizon is already in the past. */
  isStale: boolean
}

const RISK_RANK: Record<string, number> = {
  HIGH: 5,
  MODERATE_HIGH: 4,
  'MODERATE-HIGH': 4,
  MODERATE: 3,
  LOW_MODERATE: 2,
  'LOW-MODERATE': 2,
  LOW: 1,
}

const riskRank = (level: string) => RISK_RANK[String(level || '').toUpperCase().replace(/\s+TO\s+/g, '_')] ?? 0

export const toKpiForecastView = (kpi: any): KpiForecastView => {
  const meta = getKpiMeta(kpi.kpiName, kpi.unit)
  const history: KpiPoint[] = Array.isArray(kpi.history) ? kpi.history : []
  const forecast: KpiPoint[] = Array.isArray(kpi.forecast) ? kpi.forecast : []
  const current = Number(kpi.currentValue)
  // Trend is judged at the end of the horizon, the same point the chart ends on.
  const horizonValue = forecast.length ? Number(forecast[forecast.length - 1]!.value) : Number(kpi.forecastedValue)
  const trend = evaluateKpiTrend(current, horizonValue, meta)
  const rel = current ? (horizonValue - current) / Math.abs(current) : 0
  const lastActualPeriod = kpi.lastActualPeriod || history[history.length - 1]?.period || ''
  const horizonStart = forecast[0]?.period || ''
  const horizonEnd = forecast[forecast.length - 1]?.period || ''

  return {
    kpiName: kpi.kpiName,
    code: kpi.code,
    entity: kpi.entity,
    entityType: kpi.entityType,
    riskLevel: kpi.riskLevel,
    recommendedAction: kpi.recommendedAction,
    currentValue: current,
    forecastedValue: Number(kpi.forecastedValue),
    meta,
    trend,
    changeTowardsBetter: meta.higherIsBetter ? rel : -rel,
    history,
    forecast,
    lastActualPeriod,
    horizonStart,
    horizonEnd,
    isStale: !!horizonEnd && isPeriodPast(horizonEnd),
  }
}

/** Only KPIs projected to get worse, most severe first. */
export const selectAtRiskKpis = (views: KpiForecastView[]): KpiForecastView[] =>
  views
    .filter((v) => v.trend === 'Deteriorating')
    .sort((a, b) => riskRank(b.riskLevel) - riskRank(a.riskLevel) || a.changeTowardsBetter - b.changeTowardsBetter)

/** KPIs with a concrete audit procedure under analytics.kpiAudit.actions.* */
const AUDIT_ACTION_KEYS = new Set(['nps', 'clv', 'aov', 'cac', 'mrr', 'conversion', 'gpm', 'retention', 'roi', 'turnover'])

export const kpiAuditActionKey = (meta: KpiMeta): string | null =>
  meta.key && AUDIT_ACTION_KEYS.has(meta.key) ? meta.key : null

/** Horizon shared by the card: the horizon of the selected KPI. */
export const horizonLabel = (view: KpiForecastView | undefined, locale: AppLocale) =>
  view ? formatForecastPeriodRange(view.horizonStart, view.horizonEnd, locale) : ''

type Translate = (key: string, params?: any) => string

export interface KpiChartStyle {
  actualColor?: string
  forecastColor?: string
  bandColor?: string
}

export const buildKpiForecastChart = (
  view: KpiForecastView | undefined,
  mape: number,
  t: Translate,
  locale: AppLocale,
  style: KpiChartStyle = {}
) => {
  const actualColor = style.actualColor ?? '#4f46e5'
  const forecastColor = style.forecastColor ?? '#8b5cf6'
  const bandColor = style.bandColor ?? 'rgba(139, 92, 246, 0.15)'

  if (!view) {
    return { data: { labels: [], datasets: [] }, options: { responsive: true, maintainAspectRatio: false } }
  }

  const history = view.history
  const forecast = view.forecast
  const lastIdx = history.length - 1
  const labels = [...history, ...forecast].map((p) => formatForecastPeriod(p.period, locale))
  const anchor = lastIdx >= 0 ? history[lastIdx]!.value : null
  const err = Math.max(0, Number(mape) || 0) / 100

  const actual = [...history.map((p) => p.value), ...forecast.map(() => null)]
  const forecastLine = [...history.map((_, i) => (i === lastIdx ? anchor : null)), ...forecast.map((p) => p.value)]
  const upper = [
    ...history.map((_, i) => (i === lastIdx ? anchor : null)),
    ...forecast.map((p, i) => p.value + Math.abs(p.value) * err * Math.sqrt(i + 1)),
  ]
  const lower = [
    ...history.map((_, i) => (i === lastIdx ? anchor : null)),
    ...forecast.map((p, i) => p.value - Math.abs(p.value) * err * Math.sqrt(i + 1)),
  ]

  const bandLabel = t('analytics.timeseries.labelBand')
  const lowerLabel = t('analytics.timeseries.labelLowerBound')
  const fmt = (v: number) => formatKpiForecastValue(v, view.meta, locale)

  return {
    data: {
      labels,
      datasets: [
        {
          label: t('analytics.timeseries.labelActualKPI'),
          data: actual,
          borderColor: actualColor,
          backgroundColor: actualColor,
          tension: 0.3,
          borderWidth: 2.5,
          pointRadius: 3,
          spanGaps: false,
        },
        {
          label: t('analytics.timeseries.labelForecast'),
          data: forecastLine,
          borderColor: forecastColor,
          backgroundColor: forecastColor,
          borderDash: [6, 4],
          tension: 0.3,
          borderWidth: 2.5,
          pointRadius: 4,
          pointStyle: 'rectRot',
          spanGaps: false,
        },
        {
          label: bandLabel,
          data: upper,
          borderColor: 'transparent',
          backgroundColor: bandColor,
          fill: '+1',
          tension: 0.3,
          pointRadius: 0,
          spanGaps: false,
        },
        {
          label: lowerLabel,
          data: lower,
          borderColor: 'transparent',
          backgroundColor: bandColor,
          fill: false,
          tension: 0.3,
          pointRadius: 0,
          spanGaps: false,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: { intersect: false, mode: 'index' as const },
      plugins: {
        legend: {
          position: 'top' as const,
          labels: {
            usePointStyle: true,
            boxWidth: 10,
            font: { size: 11 },
            filter: (item: any) => item.text !== lowerLabel,
          },
        },
        tooltip: {
          callbacks: {
            label: (ctx: any) => {
              if (ctx.parsed?.y === null || ctx.parsed?.y === undefined) return ''
              if (ctx.dataset.label === lowerLabel) return ''
              if (ctx.dataset.label === bandLabel) {
                const lo = lower[ctx.dataIndex]
                const hi = upper[ctx.dataIndex]
                if (lo === null || hi === null || lo === hi) return ''
                return `${bandLabel}: ${fmt(lo as number)} – ${fmt(hi as number)}`
              }
              return `${ctx.dataset.label}: ${fmt(ctx.parsed.y)}`
            },
          },
        },
      },
      scales: {
        x: { grid: { display: false }, ticks: { font: { size: 10 }, maxRotation: 0, autoSkip: true, maxTicksLimit: 8 } },
        y: {
          grid: { color: 'rgba(226, 232, 240, 0.6)' },
          ticks: { font: { size: 10 }, callback: (v: any) => fmt(Number(v)) },
          title: {
            display: true,
            text: t(`analytics.kpiAudit.axis.${view.meta.kind}`, { kpi: view.kpiName }),
            font: { size: 11, weight: 'bold' as const },
          },
        },
      },
    },
  }
}
