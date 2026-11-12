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

const PALETTE_COLORS = [
  { bg: 'rgba(59,130,246,0.85)', border: 'rgba(59,130,246,1)', style: 'circle' },
  { bg: 'rgba(168,85,247,0.85)', border: 'rgba(168,85,247,1)', style: 'triangle' },
  { bg: 'rgba(236,72,153,0.85)', border: 'rgba(236,72,153,1)', style: 'rect' },
  { bg: 'rgba(14,165,233,0.85)', border: 'rgba(14,165,233,1)', style: 'star' },
  { bg: 'rgba(16,185,129,0.85)', border: 'rgba(16,185,129,1)', style: 'rectRot' },
  { bg: 'rgba(249,115,22,0.85)', border: 'rgba(249,115,22,1)', style: 'crossRot' },
]

export const getDeterministicCategoryStyle = (categoryName: string) => {
  let hash = 0
  for (let i = 0; i < categoryName.length; i++) {
    hash = (hash << 5) - hash + categoryName.charCodeAt(i)
    hash |= 0
  }
  const idx = Math.abs(hash) % PALETTE_COLORS.length
  return PALETTE_COLORS[idx]
}

/** Used when a category has no entry of its own (and as the typed fallback). */
export const getAnomalyTypeConfig = (tOrType: Translate | string, typeOrConfigs?: any, locale: AppLocale = 'id'): AnomalyTypeConfig => {
  const rupiah = (val: number) => formatRupiahFromMillions(val, locale)
  if (typeof tOrType === 'function') {
    const t = tOrType
    const type = typeof typeOrConfigs === 'string' ? typeOrConfigs : 'Funding'
    const configs = createAnomalyTypeConfigs(t, locale)
    if (configs[type]) return configs[type]
    const customShape = getDeterministicCategoryStyle(type).style
    return {
      xAxisTitle: `${type} Metric`,
      unit: 'Rupiah',
      formatX: rupiah,
      colors: anomalyColors(customShape)
    }
  } else {
    const type = String(tOrType || 'Funding')
    const configs = (typeOrConfigs && typeof typeOrConfigs === 'object') ? typeOrConfigs : {}
    if (configs[type]) return configs[type]
    const customShape = getDeterministicCategoryStyle(type).style
    return {
      xAxisTitle: `${type} Metric`,
      unit: 'Rupiah',
      formatX: rupiah,
      colors: anomalyColors(customShape)
    }
  }
}

export interface AnomalyScatterPoint {
  x: number
  y: number
  id?: string
  entity?: string
  description?: string
  type?: string
  amount?: number
  xMetric?: number
  riskLevel?: string
  isAnomaly?: boolean
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
  'Funding',
  'Lending',
  'Treasury',
  'Payment',
  'KYC',
  'IT Control',
  'Transaction',
  'Procurement',
  'Expense Report'
]

/**
 * Returns canonical category name to ensure anomalies and normal distribution
 * baseline points are ALWAYS unified in the exact same category, preventing
 * duplicate categories (such as 'Transaction' vs 'GL Transaction' or 'gl_transactions').
 */
export const getCanonicalCategory = (rawCategoryOrTableName: any): string => {
  const s = String(rawCategoryOrTableName || '').trim()
  if (!s) return 'Transaction'
  const clean = s.toLowerCase().replace(/[\s_-]+/g, '')
  if (clean === 'gltransaction' || clean === 'gltransactions' || clean === 'transaction' || clean === 'transactions') {
    return 'Transaction'
  }
  if (clean === 'loanaccounts' || clean === 'loans' || clean === 'loan' || clean === 'lending') return 'Lending'
  if (clean === 'customeraccounts' || clean === 'accounts' || clean === 'funding' || clean === 'deposits' || clean === 'savings') {
    return 'Funding'
  }
  if (clean === 'useraccessauditlogs' || clean === 'accesslogs' || clean === 'itcontrol' || clean === 'auditlogs') return 'IT Control'
  if (clean === 'complianceincidentrecords' || clean === 'complianceincidents' || clean === 'kyc' || clean === 'aml') return 'KYC'
  if (clean === 'vendorinvoices' || clean === 'invoices' || clean === 'procurement') return 'Procurement'
  if (clean === 'treasury' || clean === 'treasurytransactions') return 'Treasury'
  if (clean === 'payment' || clean === 'payments') return 'Payment'
  if (clean === 'expensereport' || clean === 'expensereports') return 'Expense Report'
  if (clean === 'inventory') return 'Inventory'

  // Format any custom/new table name cleanly as Title Case
  const words = s.replace(/_tbl$|_records$|_logs$/i, '').split(/[\s_-]+/).filter(Boolean)
  if (words.length > 0) {
    return words.map(w => w.charAt(0).toUpperCase() + w.slice(1).toLowerCase()).join(' ')
  }
  return s
}

/**
 * Category tabs to offer: every type present in the data, preserving
 * any known `validTypes` ordering first, followed by dynamic categories.
 * All types are canonicalized so no duplicate categories ever appear.
 */
export const listAnomalyTypes = (
  anomalies: any[] = [],
  scatterData: any[] = [],
  validTypes?: string[]
): string[] => {
  const types = new Set<string>()
    ; (anomalies || []).forEach((a: any) => {
      if (a?.type) types.add(getCanonicalCategory(a.type))
    })
    ; (scatterData || []).forEach((s: any) => {
      if (s?.type) types.add(getCanonicalCategory(s.type))
    })

  if (types.size === 0) {
    return [...DEFAULT_ANOMALY_TYPES]
  }

  const allTypes = Array.from(types)
  if (validTypes && validTypes.length > 0) {
    const canonicalValid = validTypes.map(getCanonicalCategory)
    const known = canonicalValid.filter((t, idx) => canonicalValid.indexOf(t) === idx && types.has(t))
    const custom = allTypes.filter((t) => !known.includes(t)).sort()
    return [...known, ...custom]
  }

  return allTypes
}

/** Keep `preferred` when the data has it, otherwise fall back to the first tab. */
export const resolveAnomalyType = (
  availableOrPreferred: string[] | string = [],
  preferredOrAvailable?: string[] | string
): string => {
  let available: string[] = []
  let preferred = 'Funding'

  if (Array.isArray(availableOrPreferred)) {
    available = availableOrPreferred.map(getCanonicalCategory)
    if (typeof preferredOrAvailable === 'string') preferred = getCanonicalCategory(preferredOrAvailable)
  } else if (typeof availableOrPreferred === 'string') {
    preferred = getCanonicalCategory(availableOrPreferred)
    if (Array.isArray(preferredOrAvailable)) available = preferredOrAvailable.map(getCanonicalCategory)
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
export const isCategoryMatching = (typeVal: any, filterVal: string) => {
  const t = getCanonicalCategory(typeVal).toLowerCase()
  const f = getCanonicalCategory(filterVal).toLowerCase()
  return t === f
}

export const buildAnomalyScatterPoints = (
  anomalies: any[] = [],
  scatterData: any[] = [],
  typeFilter: string
): AnomalyScatterSeries => {
  const allAnomalies = anomalies || []
  const allScatter = scatterData || []
  const targetType = (typeFilter || '').trim().toLowerCase()

  const normalPoints: AnomalyScatterPoint[] = allScatter
    .filter((s: any) => !s.isAnomaly && isCategoryMatching(s.type, targetType))
    .map((s: any) => ({
      x: s.x ?? 0,
      y: s.y ?? 0
    }))

  const anomalyPoints: AnomalyScatterPoint[] = allAnomalies
    .filter((a: any) => isCategoryMatching(a.type, targetType))
    .flatMap((a: any) => {
      const matchedScatter = allScatter.find((s: any) => s.label === a.id)
      let xVal = matchedScatter?.x ?? a.xMetric
      if (xVal === undefined || xVal === null) {
        // Categories with a non-rupiah metric use their own fallback; the rest
        // derive the metric from the transaction amount.
        xVal = X_FALLBACK_BY_TYPE[typeFilter]
          ?? (a.amount ? a.amount / 1000000 : X_FALLBACK_DEFAULT)
      }
      const yVal = matchedScatter?.y ?? a.hour ?? a.hour_of_day ?? a.y
      if (yVal === undefined || yVal === null) return []
      return [{ x: xVal, y: yVal }]
    })

  // Rule: Only real normal points from the dataset are displayed.
  // If a category has no real normal data in the dataset, normalPoints remains empty ([])
  // without generating synthetic dummy points.
  return { normalPoints, anomalyPoints }
}
