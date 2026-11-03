/**
 * Shared point builder for the Isolation Forest anomaly scatter chart.
 *
 * Both the dashboard card ("Transaction Anomaly Detection") and the AI Insights
 * tab (/analytics/ai/anomaly-detection) render this chart, so the mapping from
 * anomaly/scatter records to chart coordinates lives here — one implementation,
 * one set of coordinates.
 */

import { formatRupiahFromMillions, type AppLocale } from '~/utils/analyticsFormat'

export interface AnomalyTypeConfig {
  xAxisTitle: string
  unit: string
  formatX: (val: number) => string
  colors: { bg: string, border: string, style: string }
}

type Translate = (key: string, params?: any) => string

/**
 * Anomalies are always drawn in red, whatever the category: green reads as
 * "safe", which contradicts the red alert cards next to the chart. Categories
 * keep their own point shape so they stay distinguishable.
 */
export const ANOMALY_POINT_COLORS = { bg: 'rgba(220,38,38,0.85)', border: 'rgba(185,28,28,1)' }

const anomalyColors = (style: string) => ({ ...ANOMALY_POINT_COLORS, style })

/**
 * Per-category axis titles, units and point styling. Takes `t` so both pages
 * render the same labels in the active locale.
 */
export const createAnomalyTypeConfigs = (t: Translate, locale: AppLocale = 'id'): Record<string, AnomalyTypeConfig> => {
  // X values of the rupiah categories are in millions of rupiah.
  const rupiah = (val: number) => formatRupiahFromMillions(val, locale)
  return {
    'Funding': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.fundingXAxis'),
      unit: t('analytics.isolation.anomalyTypes.fundingUnit'),
      formatX: rupiah,
      colors: anomalyColors('circle')
    },
    'Lending': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.lendingXAxis'),
      unit: t('analytics.isolation.anomalyTypes.lendingUnit'),
      formatX: rupiah,
      colors: anomalyColors('triangle')
    },
    'Treasury': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.treasuryXAxis'),
      unit: t('analytics.isolation.anomalyTypes.treasuryUnit'),
      formatX: rupiah,
      colors: anomalyColors('rectRot')
    },
    'Payment': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.paymentXAxis'),
      unit: t('analytics.isolation.anomalyTypes.paymentUnit'),
      formatX: rupiah,
      colors: anomalyColors('rect')
    },
    'KYC': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.kycXAxis'),
      unit: t('analytics.isolation.anomalyTypes.kycUnit'),
      formatX: (val) => `${val}%`,
      colors: anomalyColors('star')
    },
    'IT Control': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.itControlXAxis'),
      unit: t('analytics.isolation.anomalyTypes.itControlUnit'),
      formatX: (val) => `${val} ${t('analytics.isolation.anomalyTypes.itControlUnit')}`,
      colors: anomalyColors('crossRot')
    },
    'Fieldwork': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.fieldworkXAxis'),
      unit: t('analytics.isolation.anomalyTypes.fieldworkUnit'),
      formatX: (val) => `${val} ${t('analytics.isolation.anomalyTypes.fieldworkUnit')}`,
      colors: anomalyColors('star')
    },
    'Access Pattern': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.accessPatternXAxis'),
      unit: t('analytics.isolation.anomalyTypes.accessPatternUnit'),
      formatX: (val) => `${t('analytics.isolation.anomalyTypes.accessPatternUnit')} ${val}:00`,
      colors: anomalyColors('rectRot')
    },
    'Data Access': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.dataAccessXAxis'),
      unit: t('analytics.isolation.anomalyTypes.dataAccessUnit'),
      formatX: (val) => `${val} ${t('analytics.isolation.anomalyTypes.dataAccessUnit')}`,
      colors: anomalyColors('rect')
    },
    'Inventory': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.inventoryXAxis'),
      unit: t('analytics.isolation.anomalyTypes.inventoryUnit'),
      formatX: rupiah,
      colors: anomalyColors('star')
    },
    'Expense Report': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.expenseReportXAxis'),
      unit: t('analytics.isolation.anomalyTypes.expenseReportUnit'),
      formatX: rupiah,
      colors: anomalyColors('rect')
    },
    'Travel Expense': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.travelExpenseXAxis'),
      unit: t('analytics.isolation.anomalyTypes.travelExpenseUnit'),
      formatX: rupiah,
      colors: anomalyColors('rectRot')
    },
    'Procurement': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.procurementXAxis'),
      unit: t('analytics.isolation.anomalyTypes.procurementUnit'),
      formatX: rupiah,
      colors: anomalyColors('triangle')
    },
    'Transaction': {
      xAxisTitle: t('analytics.isolation.anomalyTypes.transactionXAxis'),
      unit: t('analytics.isolation.anomalyTypes.transactionUnit'),
      formatX: rupiah,
      colors: anomalyColors('triangle')
    }
  }
}

/** Used when a category has no entry of its own (and as the typed fallback). */
export const getAnomalyTypeConfig = (tOrType: Translate | string, typeOrConfigs?: any, locale: AppLocale = 'id'): AnomalyTypeConfig => {
  if (typeof tOrType === 'function') {
    const t = tOrType
    const type = typeof typeOrConfigs === 'string' ? typeOrConfigs : 'Funding'
    const configs = createAnomalyTypeConfigs(t, locale)
    return (
      configs[type] ??
      configs['Funding'] ?? {
        xAxisTitle: t('analytics.isolation.anomalyTypes.fundingXAxis'),
        unit: t('analytics.isolation.anomalyTypes.fundingUnit'),
        formatX: (val: number) => formatRupiahFromMillions(val, locale),
        colors: anomalyColors('circle')
      }
    )
  } else {
    const type = String(tOrType || 'Funding')
    const configs = (typeOrConfigs && typeof typeOrConfigs === 'object') ? typeOrConfigs : {}
    return (
      configs[type] ??
      configs['Funding'] ?? {
        xAxisTitle: 'Amount (Rp)',
        unit: 'Rp',
        formatX: (val: number) => formatRupiahFromMillions(val, locale),
        colors: anomalyColors('circle')
      }
    )
  }
}

export interface AnomalyScatterPoint {
  x: number
  y: number
}

export interface AnomalyScatterSeries {
  normalPoints: AnomalyScatterPoint[]
  anomalyPoints: AnomalyScatterPoint[]
}

/** X fallbacks for categories whose metric is not a rupiah amount. */
const X_FALLBACK_BY_TYPE: Record<string, number> = {
  KYC: 25,
  'IT Control': 4
}

const X_FALLBACK_DEFAULT = 15

export const DEFAULT_ANOMALY_TYPES = [
  'Transaction',
  'Funding',
  'Lending',
  'Treasury',
  'Payment',
  'Procurement',
  'Expense Report'
]

/**
 * Category tabs to offer: every type present in the data, restricted to
 * `validTypes` when that yields anything, else whatever the data contains.
 */
export const listAnomalyTypes = (
  anomalies: any[] = [],
  scatterData: any[] = [],
  validTypes?: string[]
): string[] => {
  const types = new Set<string>()
  ;(anomalies || []).forEach((a: any) => { if (a?.type) types.add(a.type) })
  ;(scatterData || []).forEach((s: any) => { if (s?.type) types.add(s.type) })

  if (validTypes && validTypes.length > 0) {
    const filtered = Array.from(types).filter((t) => validTypes.includes(t))
    if (filtered.length > 0) return filtered
  }
  if (types.size > 0) return Array.from(types)
  return [...DEFAULT_ANOMALY_TYPES]
}

/** Keep `preferred` when the data has it, otherwise fall back to the first tab. */
export const resolveAnomalyType = (
  availableOrPreferred: string[] | string = [],
  preferredOrAvailable?: string[] | string
): string => {
  let available: string[] = []
  let preferred = 'Transaction'

  if (Array.isArray(availableOrPreferred)) {
    available = availableOrPreferred
    if (typeof preferredOrAvailable === 'string') preferred = preferredOrAvailable
  } else if (typeof availableOrPreferred === 'string') {
    preferred = availableOrPreferred
    if (Array.isArray(preferredOrAvailable)) available = preferredOrAvailable
  }

  if (available.includes(preferred)) return preferred
  return available[0] || preferred
}

/**
 * Chart coordinates for one category.
 *
 * X is the category's real metric (rupiah in millions, failed logins,
 * deviation %, …) and Y is the transaction hour (0-23) — never the anomaly
 * score and never the record's position in the array. An anomaly whose hour is
 * unknown is left off the chart (it is still listed in the alerts) rather than
 * plotted at an invented position.
 */
export const buildAnomalyScatterPoints = (
  anomalies: any[] = [],
  scatterData: any[] = [],
  typeFilter: string
): AnomalyScatterSeries => {
  const allAnomalies = anomalies || []
  const allScatter = scatterData || []

  const normalPoints = allScatter
    .filter((s: any) => !s.isAnomaly && (s.type === typeFilter || !s.type))
    .map((s: any) => ({ x: s.x ?? 0, y: s.y ?? 0 }))

  const anomalyPoints = allAnomalies
    .filter((a: any) => a.type === typeFilter)
    .flatMap((a: any) => {
      const matchedScatter = allScatter.find((s: any) => s.label === a.id)
      let xVal = matchedScatter?.x ?? a.xMetric
      if (xVal === undefined || xVal === null) {
        // Categories with a non-rupiah metric use their own fallback; the rest
        // derive the metric from the transaction amount.
        xVal = X_FALLBACK_BY_TYPE[typeFilter]
          ?? (a.amount ? a.amount / 1000000 : X_FALLBACK_DEFAULT)
      }
      const yVal = matchedScatter?.y ?? a.hour
      if (yVal === undefined || yVal === null) return []
      return [{ x: xVal, y: yVal }]
    })

  return { normalPoints, anomalyPoints }
}
