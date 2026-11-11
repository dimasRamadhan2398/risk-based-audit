/**
 * Shared point builder for the Isolation Forest anomaly scatter chart.
 *
 * Both the dashboard card ("Transaction Anomaly Detection") and the AI Insights
 * tab (/analytics/ai/anomaly-detection) render this chart, so the mapping from
 * anomaly/scatter records to chart coordinates lives here — one implementation,
 * one set of coordinates.
 */

export interface AnomalyTypeConfig {
  xAxisTitle: string
  unit: string
  formatX: (val: number) => string
  colors: { bg: string, border: string, style: string }
}

type Translate = (key: string, params?: any) => string

/**
 * Per-category axis titles, units and point styling. Takes `t` so both pages
 * render the same labels in the active locale.
 */
export const createAnomalyTypeConfigs = (t: Translate): Record<string, AnomalyTypeConfig> => ({
  'Funding': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.fundingXAxis'),
    unit: t('analytics.isolation.anomalyTypes.fundingUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(16,185,129,0.85)', border: 'rgba(16,185,129,1)', style: 'circle' }
  },
  'Lending': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.lendingXAxis'),
    unit: t('analytics.isolation.anomalyTypes.lendingUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(239,68,68,0.85)', border: 'rgba(239,68,68,1)', style: 'triangle' }
  },
  'Treasury': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.treasuryXAxis'),
    unit: t('analytics.isolation.anomalyTypes.treasuryUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(99,102,241,0.85)', border: 'rgba(99,102,241,1)', style: 'rectRot' }
  },
  'Payment': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.paymentXAxis'),
    unit: t('analytics.isolation.anomalyTypes.paymentUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(245,158,11,0.85)', border: 'rgba(245,158,11,1)', style: 'rect' }
  },
  'KYC': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.kycXAxis'),
    unit: t('analytics.isolation.anomalyTypes.kycUnit'),
    formatX: (val) => `${val}%`,
    colors: { bg: 'rgba(236,72,153,0.85)', border: 'rgba(236,72,153,1)', style: 'star' }
  },
  'IT Control': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.itControlXAxis'),
    unit: t('analytics.isolation.anomalyTypes.itControlUnit'),
    formatX: (val) => `${val} ${t('analytics.isolation.anomalyTypes.itControlUnit')}`,
    colors: { bg: 'rgba(20,184,166,0.85)', border: 'rgba(20,184,166,1)', style: 'crossRot' }
  },
  'Fieldwork': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.fieldworkXAxis'),
    unit: t('analytics.isolation.anomalyTypes.fieldworkUnit'),
    formatX: (val) => `${val} ${t('analytics.isolation.anomalyTypes.fieldworkUnit')}`,
    colors: { bg: 'rgba(139,92,246,0.85)', border: 'rgba(139,92,246,1)', style: 'star' }
  },
  'Access Pattern': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.accessPatternXAxis'),
    unit: t('analytics.isolation.anomalyTypes.accessPatternUnit'),
    formatX: (val) => `${t('analytics.isolation.anomalyTypes.accessPatternUnit')} ${val}:00`,
    colors: { bg: 'rgba(249,115,22,0.85)', border: 'rgba(249,115,22,1)', style: 'rectRot' }
  },
  'Data Access': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.dataAccessXAxis'),
    unit: t('analytics.isolation.anomalyTypes.dataAccessUnit'),
    formatX: (val) => `${val} ${t('analytics.isolation.anomalyTypes.dataAccessUnit')}`,
    colors: { bg: 'rgba(59,130,246,0.85)', border: 'rgba(59,130,246,1)', style: 'rect' }
  },
  'Inventory': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.inventoryXAxis'),
    unit: t('analytics.isolation.anomalyTypes.inventoryUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(107,114,128,0.85)', border: 'rgba(107,114,128,1)', style: 'star' }
  },
  'Expense Report': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.expenseReportXAxis'),
    unit: t('analytics.isolation.anomalyTypes.expenseReportUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(234,179,8,0.85)', border: 'rgba(234,179,8,1)', style: 'rect' }
  },
  'Travel Expense': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.travelExpenseXAxis'),
    unit: t('analytics.isolation.anomalyTypes.travelExpenseUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(16,185,129,0.85)', border: 'rgba(16,185,129,1)', style: 'rectRot' }
  },
  'Procurement': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.procurementXAxis'),
    unit: t('analytics.isolation.anomalyTypes.procurementUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(236,72,153,0.85)', border: 'rgba(236,72,153,1)', style: 'triangle' }
  },
  'Transaction': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.transactionXAxis'),
    unit: t('analytics.isolation.anomalyTypes.transactionUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(239,68,68,0.85)', border: 'rgba(239,68,68,1)', style: 'triangle' }
  }
})

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
export const getAnomalyTypeConfig = (tOrType: Translate | string, typeOrConfigs?: any): AnomalyTypeConfig => {
  if (typeof tOrType === 'function') {
    const t = tOrType
    const type = typeof typeOrConfigs === 'string' ? typeOrConfigs : 'Funding'
    const configs = createAnomalyTypeConfigs(t)
    if (configs[type]) return configs[type]
    const customStyle = getDeterministicCategoryStyle(type)
    return {
      xAxisTitle: `${type} Metric`,
      unit: 'Rp M',
      formatX: (val: number) => `Rp ${val}M`,
      colors: customStyle
    }
  } else {
    const type = String(tOrType || 'Funding')
    const configs = (typeOrConfigs && typeof typeOrConfigs === 'object') ? typeOrConfigs : {}
    if (configs[type]) return configs[type]
    const customStyle = getDeterministicCategoryStyle(type)
    return {
      xAxisTitle: `${type} Metric`,
      unit: 'Rp M',
      formatX: (val: number) => `Rp ${val}M`,
      colors: customStyle
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

/** Y coordinate used when an anomaly has no matching scatter point. */
const DEFAULT_FREQUENCY = 20

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
  ;(anomalies || []).forEach((a: any) => {
    if (a?.type) types.add(getCanonicalCategory(a.type))
  })
  ;(scatterData || []).forEach((s: any) => {
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
 * X is the category's real metric (rupiah in millions, hour of day, deviation %,
 * …) and Y is its observed frequency — never the anomaly score and never the
 * record's position in the array.
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
    .map((a: any) => {
      const matchedScatter = allScatter.find((s: any) => s.label === a.id)
      let xVal = matchedScatter?.x ?? a.xMetric
      if (xVal === undefined || xVal === null) {
        // Categories with a non-rupiah metric use their own fallback; the rest
        // derive the metric from the transaction amount.
        xVal = X_FALLBACK_BY_TYPE[typeFilter]
          ?? (a.amount ? a.amount / 1000000 : X_FALLBACK_DEFAULT)
      }
      return {
        x: xVal,
        y: matchedScatter?.y ?? DEFAULT_FREQUENCY
      }
    })

  // Rule: Only real normal points from the dataset are displayed.
  // If a category has no real normal data in the dataset, normalPoints remains empty ([])
  // without generating synthetic dummy points.
  return { normalPoints, anomalyPoints }
}
