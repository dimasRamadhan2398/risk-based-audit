import { ref, computed } from 'vue'
import {
  useXGBoostData,
  useIsolationForestData,
  useIndoBERTData,
  useTimeSeriesData,
  useAnalyticsSummary,
  type RiskScorePrediction,
  type AnomalyRecord,
  type NLPDocumentResult,
  type KPIForecast
} from '~/composables/useAnalyticsData'
import {
  ENTITY_RISK_LEVELS,
  NLP_RISK_LEVELS,
  normalizeRiskLevel,
  normalizeAnomalyRiskLevel,
  normalizeEntityRiskLevel,
  normalizeNlpRiskLevel,
  toEntityRiskCode,
  toNlpRiskCode,
  getRiskLevelColorClass
} from '~/utils/riskLevelBadge'
import { getAnalyticsServiceBaseUrl, getPythonAiBaseUrl, getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { getCanonicalCategory } from '~/composables/useAnomalyScatter'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { buildAuditableEntityOptions } from '~/utils/auditableEntityOptions'
import { isReadableHumanText } from '~/utils/documentTextExtractor'
import {
  buildAuditorAggregateTimeSeries,
  buildAuditorLocalForecastsFromItems,
  computeAuditeeCustomFingerprint,
  computeAuditorFingerprint,
  dedupAndBuildAuditorPayloadItems,
  type AuditorPayloadItem
} from '~/utils/auditorKpiSync'

const CUSTOM_ANOMALIES_KEY = 'auditsphere_custom_anomaly_records_v1'
const CUSTOM_SCATTER_KEY = 'auditsphere_custom_scatter_records_v1'
const CUSTOM_NLP_DOCS_KEY = 'auditsphere_custom_nlp_docs_v3'
const CUSTOM_NLP_TAXONOMY_KEY = 'auditsphere_custom_nlp_taxonomy_v2'
const AUTOMATED_ANOMALY_RULES_KEY = 'auditsphere_automated_anomaly_rules_v1'
const CUSTOM_ENTITY_PREDICTIONS_KEY = 'auditsphere_custom_entity_predictions_v1'
const CUSTOM_AUDITEE_KPIS_KEY = 'auditsphere_custom_auditee_kpis_v1'

const CANONICAL_ENTITY_UI_PAIRS: Record<string, [number, number]> = {
  HIGH: [4.7, 4.8],
  MODERATE_HIGH: [3.8, 4.0],
  MODERATE: [3.0, 3.2],
  LOW_MODERATE: [2.2, 2.4],
  LOW: [1.2, 1.4]
}

const clampEntityRating = (v: unknown, fallback: number): number => {
  const n = Number(v)
  if (!Number.isFinite(n) || n <= 0) return fallback
  return Math.min(5, Math.max(1, n))
}

const matrixCalc5Tier = (lik: number, imp: number): string => {
  const m: Record<number, Record<number, string>> = {
    5: { 1: 'LOW_MODERATE', 2: 'MODERATE', 3: 'MODERATE_HIGH', 4: 'HIGH', 5: 'HIGH' },
    4: { 1: 'LOW', 2: 'LOW_MODERATE', 3: 'MODERATE', 4: 'MODERATE_HIGH', 5: 'HIGH' },
    3: { 1: 'LOW', 2: 'LOW_MODERATE', 3: 'MODERATE', 4: 'MODERATE_HIGH', 5: 'HIGH' },
    2: { 1: 'LOW', 2: 'LOW_MODERATE', 3: 'LOW_MODERATE', 4: 'MODERATE_HIGH', 5: 'HIGH' },
    1: { 1: 'LOW', 2: 'LOW', 3: 'LOW_MODERATE', 4: 'MODERATE', 5: 'HIGH' }
  }
  const lInt = Math.min(5, Math.max(1, Math.round(Number(lik) || 3)))
  const iInt = Math.min(5, Math.max(1, Math.round(Number(imp) || 3)))
  return m[lInt]?.[iInt] || 'MODERATE'
}

export const sanitizeEntityPredictionRecord = (it: any) => {
  if (!it || !it.entity) return null
  const rawActual = it.actualRiskLevel ?? it.actual_risk_level ?? it.currentRiskLevel ?? it.current_risk_level
  const rawCurLik = it.currentLikelihood ?? it.current_likelihood ?? it.inherent_likelihood
  const rawCurImp = it.currentImpact ?? it.current_impact ?? it.inherent_impact
  const actualCode = rawActual
    ? toEntityRiskCode(rawActual)
    : matrixCalc5Tier(clampEntityRating(rawCurLik, 3), clampEntityRating(rawCurImp, 3))
  const [defLik, defImp] = CANONICAL_ENTITY_UI_PAIRS[actualCode] || [3.0, 3.2]
  const curLik = clampEntityRating(rawCurLik, defLik)
  const curImp = clampEntityRating(rawCurImp, defImp)
  const rawPred = it.predictedRiskLevel ?? it.risk_level ?? it.riskLevel
  const predLik = clampEntityRating(it.predictedLikelihood ?? it.predicted_likelihood, Math.round(curLik))
  const predImp = clampEntityRating(it.predictedImpact ?? it.predicted_impact, Math.round(curImp))
  const predCode = rawPred ? toEntityRiskCode(rawPred) : matrixCalc5Tier(predLik, predImp)
  return {
    ...it,
    type: it.type === 'Department' ? 'Auditable Entity' : (it.type || 'Auditable Entity'),
    currentLikelihood: curLik,
    currentImpact: curImp,
    predictedLikelihood: predLik,
    predictedImpact: predImp,
    predictedRiskLevel: predCode,
    actualRiskLevel: actualCode
  }
}

const CANONICAL_AUDITEE_SCORE_PAIRS: Record<string, [number, number]> = {
  HIGH: [5, 5],
  MODERATE_HIGH: [4, 4],
  MODERATE: [3, 3],
  LOW_MODERATE: [2, 2],
  LOW: [1, 1]
}

const clampAuditeeIntScore = (v: unknown, fallback: number): number => {
  const n = Number(v)
  if (!Number.isFinite(n) || n <= 0) return fallback
  return Math.min(5, Math.max(1, Math.round(n)))
}

export const sanitizeAuditeeKpiForecastRecord = (it: any) => {
  if (!it || !it.kpiName) return null
  const rawLevel = it.riskLevel ?? it.risk_level ?? it.severity ?? it.riskRating ?? it.tingkat_risiko
  const hasExplicit = rawLevel !== undefined && rawLevel !== null && String(rawLevel).trim() !== '' && !['not scored', 'none', 'nan', 'null'].includes(String(rawLevel).trim().toLowerCase())
  const rawImp = it.impactScore ?? it.impact
  const rawLik = it.likelihoodScore ?? it.likelihood
  let imp = clampAuditeeIntScore(rawImp, 3)
  let lik = clampAuditeeIntScore(rawLik, 3)

  const riskCode = hasExplicit ? toEntityRiskCode(rawLevel) : matrixCalc5Tier(lik, imp)
  if (hasExplicit && (rawImp == null || rawLik == null || matrixCalc5Tier(lik, imp) !== riskCode)) {
    const [defImp, defLik] = CANONICAL_AUDITEE_SCORE_PAIRS[riskCode] || [3, 3]
    imp = defImp
    lik = defLik
  }

  return {
    ...it,
    impactScore: imp,
    likelihoodScore: lik,
    riskLevel: riskCode
  }
}

export const sanitizeAuditeeAtRiskRecord = (d: any) => {
  if (!d) return d
  const rawLevel = d.riskLevel ?? d.risk_level ?? d.severity
  return {
    ...d,
    riskLevel: toEntityRiskCode(rawLevel || 'HIGH')
  }
}

export const getPersistedCustomEntityPredictions = (): any[] => {
  if (typeof window === 'undefined') return []
  try {
    const raw = localStorage.getItem(CUSTOM_ENTITY_PREDICTIONS_KEY)
    const list = raw ? JSON.parse(raw) : []
    return Array.isArray(list)
      ? list.map(sanitizeEntityPredictionRecord).filter(Boolean)
      : []
  } catch {
    return []
  }
}

export const savePersistedCustomEntityPredictions = (items: any[]) => {
  if (typeof window === 'undefined' || !Array.isArray(items)) return
  try {
    const existing = getPersistedCustomEntityPredictions()
    const map = new Map<string, any>()
    for (const rawIt of items) {
      const it = sanitizeEntityPredictionRecord(rawIt)
      if (it?.entity) map.set(`${String(it.entity).toLowerCase()}__${String(it.riskCategory || '').toLowerCase()}`, it)
    }
    for (const rawEx of existing) {
      const ex = sanitizeEntityPredictionRecord(rawEx)
      const k = `${String(ex?.entity || '').toLowerCase()}__${String(ex?.riskCategory || '').toLowerCase()}`
      if (ex?.entity && !map.has(k)) map.set(k, ex)
    }
    localStorage.setItem(CUSTOM_ENTITY_PREDICTIONS_KEY, JSON.stringify(Array.from(map.values())))
  } catch (e) {
    console.warn('Failed to save persisted custom entity predictions', e)
  }
}

const mergeEntityPredictions = (baseList: any[], customList: any[]): any[] => {
  const seen = new Set<string>()
  const merged: any[] = []
  for (const rawItem of [...(customList || []), ...(baseList || [])]) {
    const item = sanitizeEntityPredictionRecord(rawItem)
    if (!item?.entity) continue
    const key = `${String(item.entity).trim().toLowerCase()}__${String(item.riskCategory || '').trim().toLowerCase()}`
    if (!seen.has(key)) {
      seen.add(key)
      merged.push(item)
    }
  }
  return merged
}

export const getPersistedCustomAuditeeKpis = (): any[] => {
  if (typeof window === 'undefined') return []
  try {
    const raw = localStorage.getItem(CUSTOM_AUDITEE_KPIS_KEY)
    const list = raw ? JSON.parse(raw) : []
    return Array.isArray(list) ? list.map(sanitizeAuditeeKpiForecastRecord).filter(Boolean) : []
  } catch {
    return []
  }
}

export const savePersistedCustomAuditeeKpis = (items: any[]) => {
  if (typeof window === 'undefined' || !Array.isArray(items)) return
  try {
    const existing = getPersistedCustomAuditeeKpis()
    const map = new Map<string, any>()
    for (const rawIt of items) {
      const it = sanitizeAuditeeKpiForecastRecord(rawIt)
      if (it?.kpiName) map.set(`${String(it.code || '').toLowerCase()}__${String(it.kpiName).toLowerCase()}`, it)
    }
    for (const ex of existing) {
      const k = `${String(ex?.code || '').toLowerCase()}__${String(ex?.kpiName || '').toLowerCase()}`
      if (ex?.kpiName && !map.has(k)) map.set(k, ex)
    }
    localStorage.setItem(CUSTOM_AUDITEE_KPIS_KEY, JSON.stringify(Array.from(map.values())))
  } catch (e) {
    console.warn('Failed to save persisted custom auditee KPIs', e)
  }
}

const mergeAuditeeKpiForecasts = (baseList: any[], customList: any[]): any[] => {
  const seen = new Set<string>()
  const merged: any[] = []
  for (const item of [...(customList || []), ...(baseList || [])]) {
    if (!item?.kpiName) continue
    const key = `${String(item.code || '').trim().toLowerCase()}__${String(item.kpiName).trim().toLowerCase()}`
    if (!seen.has(key)) {
      seen.add(key)
      merged.push(item)
    }
  }
  return merged
}

export const getPersistedAutomatedRules = (): any[] => {
  if (typeof window === 'undefined') return []
  try {
    const raw = localStorage.getItem(AUTOMATED_ANOMALY_RULES_KEY)
    const list = raw ? JSON.parse(raw) : []
    return Array.isArray(list)
      ? list.map((r: any) => ({
        ...r,
        severity: normalizeAnomalyRiskLevel(r?.severity || r?.riskLevel || 'High')
      }))
      : []
  } catch {
    return []
  }
}

export const savePersistedAutomatedRules = (rules: any[]) => {
  if (typeof window === 'undefined') return
  try {
    const normalized = Array.isArray(rules)
      ? rules.map((r: any) => ({
        ...r,
        severity: normalizeAnomalyRiskLevel(r?.severity || r?.riskLevel || 'High')
      }))
      : []
    localStorage.setItem(AUTOMATED_ANOMALY_RULES_KEY, JSON.stringify(normalized))
  } catch (e) {
    console.warn('Failed to save persisted automated anomaly rules', e)
  }
}

export const getPersistedCustomAnomalies = (): any[] => {
  if (typeof window === 'undefined') return []
  try {
    const raw = localStorage.getItem(CUSTOM_ANOMALIES_KEY)
    const list = raw ? JSON.parse(raw) : []
    return Array.isArray(list) ? list.map(a => {
      const normalizedLevel = normalizeAnomalyRiskLevel(a.riskLevel || a.risk_level || a.severity)
      return {
        ...a,
        type: getCanonicalCategory(a.type),
        category: getCanonicalCategory(a.category || a.type),
        severity: normalizedLevel,
        riskLevel: normalizedLevel
      }
    }) : []
  } catch {
    return []
  }
}

export const getPersistedCustomScatter = (): any[] => {
  if (typeof window === 'undefined') return []
  try {
    const raw = localStorage.getItem(CUSTOM_SCATTER_KEY)
    const list = raw ? JSON.parse(raw) : []
    return Array.isArray(list) ? list.map(s => ({
      ...s,
      type: getCanonicalCategory(s.type)
    })) : []
  } catch {
    return []
  }
}

export const sanitizeNlpDocumentRecord = (d: any) => {
  if (!d || typeof d !== 'object') return d
  const rawRisk = d.riskLevel ?? d.risk_level ?? d.risk_level_label ?? d.riskLevelLabel ?? d.severity
  const hasExplicit =
    rawRisk !== undefined &&
    rawRisk !== null &&
    String(rawRisk).trim() !== '' &&
    !['not scored', 'none', 'nan', 'null'].includes(String(rawRisk).trim().toLowerCase())
  const isPos =
    String(d.sentiment || '').toLowerCase() === 'positive' ||
    d.controlIndicator === 'Adequate Control' ||
    d.control_indicator === 'Adequate Control'
  const fallbackLabel = isPos
    ? 'Low'
    : (d.severityScore !== undefined ? normalizeNlpRiskLevel(d.severityScore, 'Moderate') : 'Moderate')
  const canonicalLabel = hasExplicit
    ? normalizeNlpRiskLevel(rawRisk, fallbackLabel)
    : (d.severityScore !== undefined ? normalizeNlpRiskLevel(d.severityScore, fallbackLabel) : fallbackLabel)
  const canonicalCode = toNlpRiskCode(canonicalLabel, fallbackLabel)
  return {
    ...d,
    riskLevel: canonicalCode,
    riskLevelLabel: canonicalLabel
  }
}

export const getPersistedCustomNlpDocs = (): any[] => {
  if (typeof window === 'undefined') return []
  try {
    const raw = localStorage.getItem(CUSTOM_NLP_DOCS_KEY)
    const list = raw ? JSON.parse(raw) : []
    return Array.isArray(list) ? list.map(sanitizeNlpDocumentRecord).filter(Boolean) : []
  } catch {
    return []
  }
}

export const savePersistedCustomNlpDoc = (doc: any) => {
  if (typeof window === 'undefined' || !doc) return
  try {
    const sanitized = sanitizeNlpDocumentRecord(doc)
    const current = getPersistedCustomNlpDocs()
    const targetId = sanitized.docId
    const existsIdx = current.findIndex(d => d.docId === targetId)
    if (existsIdx >= 0) {
      current[existsIdx] = sanitized
    } else {
      current.unshift(sanitized)
    }
    localStorage.setItem(CUSTOM_NLP_DOCS_KEY, JSON.stringify(current))
  } catch (e) {
    console.warn('Failed to save persisted custom NLP doc', e)
  }
}

export const deletePersistedCustomNlpDoc = (docId: string) => {
  if (typeof window === 'undefined') return
  try {
    const current = getPersistedCustomNlpDocs().filter(d => d.docId !== docId)
    localStorage.setItem(CUSTOM_NLP_DOCS_KEY, JSON.stringify(current))
  } catch (e) {
    console.warn('Failed to delete persisted custom NLP doc', e)
  }
}

export const clearAllPersistedCustomNlpDocs = () => {
  if (typeof window === 'undefined') return
  try {
    localStorage.removeItem(CUSTOM_NLP_DOCS_KEY)
  } catch (e) {
    console.warn('Failed to clear persisted custom NLP docs', e)
  }
}

export const getPersistedCustomNlpTaxonomy = (): { risk_categories: any[], control_indicators: any[] } => {
  if (typeof window === 'undefined') return { risk_categories: [], control_indicators: [] }
  try {
    const raw = localStorage.getItem(CUSTOM_NLP_TAXONOMY_KEY)
    const parsed = raw ? JSON.parse(raw) : null
    if (parsed && typeof parsed === 'object') {
      return {
        risk_categories: Array.isArray(parsed.risk_categories) ? parsed.risk_categories : [],
        control_indicators: Array.isArray(parsed.control_indicators) ? parsed.control_indicators : []
      }
    }
    return { risk_categories: [], control_indicators: [] }
  } catch {
    return { risk_categories: [], control_indicators: [] }
  }
}

export const mergeNlpTaxonomy = (baseTax: any = {}, customTax: any = {}): { risk_categories: any[], control_indicators: any[] } => {
  const baseCats = Array.isArray(baseTax?.risk_categories) ? baseTax.risk_categories : []
  const customCats = Array.isArray(customTax?.risk_categories) ? customTax.risk_categories : []
  const catMap = new Map<string, any>()
  baseCats.forEach((c: any) => {
    const name = (c.name || c).toLowerCase()
    catMap.set(name, typeof c === 'string' ? { name: c, is_default: true } : c)
  })
  customCats.forEach((c: any) => {
    const name = (c.name || c).toLowerCase()
    catMap.set(name, typeof c === 'string' ? { name: c, is_default: false } : c)
  })

  const baseInds = Array.isArray(baseTax?.control_indicators) ? baseTax.control_indicators : []
  const customInds = Array.isArray(customTax?.control_indicators) ? customTax.control_indicators : []
  const indMap = new Map<string, any>()
  baseInds.forEach((i: any) => {
    const name = (i.name || i).toLowerCase()
    indMap.set(name, typeof i === 'string' ? { name: i, is_default: true } : i)
  })
  customInds.forEach((i: any) => {
    const name = (i.name || i).toLowerCase()
    indMap.set(name, typeof i === 'string' ? { name: i, is_default: false } : i)
  })

  return {
    risk_categories: Array.from(catMap.values()),
    control_indicators: Array.from(indMap.values())
  }
}

export const savePersistedCustomNlpTaxonomy = (taxonomy: any) => {
  if (typeof window === 'undefined' || !taxonomy) return
  try {
    const current = getPersistedCustomNlpTaxonomy()
    const merged = mergeNlpTaxonomy(taxonomy, current)
    localStorage.setItem(CUSTOM_NLP_TAXONOMY_KEY, JSON.stringify(merged))
  } catch (e) {
    console.warn('Failed to save persisted custom NLP taxonomy', e)
  }
}

export const mergeNlpDocuments = (baseDocs: any[] = [], customDocs: any[] = []): any[] => {
  const sanitizedCustom = (customDocs || []).map(sanitizeNlpDocumentRecord).filter(Boolean)
  const sanitizedBase = (baseDocs || []).map(sanitizeNlpDocumentRecord).filter(Boolean)
  const customIds = new Set(sanitizedCustom.map(c => c.docId))
  const filteredBase = sanitizedBase.filter(b => !customIds.has(b.docId))
  return [...sanitizedCustom, ...filteredBase]
}

export const savePersistedCustomAnomaly = (anom: any, scatter?: any) => {
  if (typeof window === 'undefined') return
  try {
    const normalizedLevel = normalizeAnomalyRiskLevel(anom.riskLevel || anom.risk_level || anom.severity)
    const normalizedAnom = {
      ...anom,
      type: getCanonicalCategory(anom.type),
      category: getCanonicalCategory(anom.category || anom.type),
      severity: normalizedLevel,
      riskLevel: normalizedLevel
    }
    const current = getPersistedCustomAnomalies()
    const existsIdx = current.findIndex(a => a.id === normalizedAnom.id)
    if (existsIdx >= 0) {
      current[existsIdx] = normalizedAnom
    } else {
      current.unshift(normalizedAnom)
    }
    localStorage.setItem(CUSTOM_ANOMALIES_KEY, JSON.stringify(current))

    if (scatter) {
      const normalizedScatter = {
        ...scatter,
        type: getCanonicalCategory(scatter.type)
      }
      const currentScatter = getPersistedCustomScatter()
      const sIdx = currentScatter.findIndex(s => s.label === normalizedScatter.label)
      if (sIdx >= 0) {
        currentScatter[sIdx] = normalizedScatter
      } else {
        currentScatter.push(normalizedScatter)
      }
      localStorage.setItem(CUSTOM_SCATTER_KEY, JSON.stringify(currentScatter))
    }
  } catch (e) {
    console.warn('Failed to save persisted custom anomaly', e)
  }
}

export const savePersistedCustomScatterBatch = (points: any[]) => {
  if (typeof window === 'undefined' || !Array.isArray(points) || points.length === 0) return
  try {
    const currentScatter = getPersistedCustomScatter()
    for (const rawPt of points) {
      const pt = {
        ...rawPt,
        type: getCanonicalCategory(rawPt.type)
      }
      const sIdx = currentScatter.findIndex(s => s.label === pt.label)
      if (sIdx >= 0) {
        currentScatter[sIdx] = pt
      } else {
        currentScatter.push(pt)
      }
    }
    localStorage.setItem(CUSTOM_SCATTER_KEY, JSON.stringify(currentScatter))
  } catch (e) {
    console.warn('Failed to save persisted custom scatter batch', e)
  }
}

// Module-level singletons so all AI Insight sub-pages share data & connection status
const initialXGB = useXGBoostData()
const initialIso = useIsolationForestData()
const initialNLP = useIndoBERTData()
const initialTS = useTimeSeriesData()
const summary = ref(useAnalyticsSummary())

const loading = ref(false)
const error = ref('')
const isAiConnected = ref(true)
const usingCachedRealData = ref(false)
const lastSyncedTime = ref<string | null>(null)
const hasFetchedOnce = ref(false)

const initialCustomAnoms = typeof window !== 'undefined' ? getPersistedCustomAnomalies() : []
const initialCustomScatter = typeof window !== 'undefined' ? getPersistedCustomScatter() : []
const initialCustomDocs = typeof window !== 'undefined' ? getPersistedCustomNlpDocs() : []
const initialCustomTax = typeof window !== 'undefined' ? getPersistedCustomNlpTaxonomy() : { risk_categories: [], control_indicators: [] }
const initialAutomatedRules = typeof window !== 'undefined' ? getPersistedAutomatedRules() : []
const automatedAnomalyRules = ref<any[]>(initialAutomatedRules)

const mergedInitialDocs = mergeNlpDocuments(initialNLP.documents, initialCustomDocs)
const mergedInitialTax = mergeNlpTaxonomy((initialNLP as any).taxonomy || { risk_categories: [], control_indicators: [] }, initialCustomTax)

const initialCustomEntityPredictions = getPersistedCustomEntityPredictions()
const xgboostState = ref<any>({
  predictions: mergeEntityPredictions([...initialXGB.predictions], initialCustomEntityPredictions),
  featureImportance: [...initialXGB.featureImportance],
  modelMetrics: { ...initialXGB.modelMetrics }
})

const normalizeIsolationAnomaliesList = (list: any[] = []) =>
  (list || []).map((a: any) => {
    const normalizedLevel = normalizeAnomalyRiskLevel(a.riskLevel || a.risk_level || a.severity)
    return {
      ...a,
      type: getCanonicalCategory(a.type),
      category: getCanonicalCategory(a.category || a.type),
      severity: normalizedLevel,
      riskLevel: normalizedLevel
    }
  })

const isolationState = ref<any>({
  anomalies: normalizeIsolationAnomaliesList([...initialCustomAnoms, ...initialIso.anomalies]),
  scatterData: [...initialCustomScatter, ...initialIso.scatterData],
  summary: { ...initialIso.summary }
})

const nlpState = ref<any>({
  documents: mergedInitialDocs,
  categoryDistribution: { ...initialNLP.categoryDistribution },
  sentimentDistribution: { ...initialNLP.sentimentDistribution },
  controlIndicatorDistribution: { ...((initialNLP as any).controlIndicatorDistribution || {}) },
  sourceTypeDistribution: { ...((initialNLP as any).sourceTypeDistribution || {}) },
  taxonomy: mergedInitialTax
})

export const recalculateNlpDistributions = () => {
  const docs = nlpState.value.documents || []
  const taxCats = (nlpState.value.taxonomy?.risk_categories || []).map((c: any) => c.name || c)
  const baseCats = taxCats.length > 0 ? taxCats : ['Financial', 'Operational', 'Compliance', 'Technology']
  const catDist: Record<string, number> = {}
  baseCats.forEach((c: string) => { catDist[c] = 0 })

  const sentDist = { positive: 0, neutral: 0, negative: 0 }
  const indDist: Record<string, number> = {}
  const srcDist: Record<string, number> = {
    'Imported Document': 0,
    'Internal Note': 0,
    'External Database': 0
  }

  docs.forEach((d: any) => {
    const cat = d.autoCategory || d.risk_category || 'Operational'
    catDist[cat] = (catDist[cat] || 0) + 1

    const s = String(d.sentiment || 'Neutral').toLowerCase()
    if (s === 'positive') sentDist.positive++
    else if (s === 'negative') sentDist.negative++
    else sentDist.neutral++

    const ind = d.controlIndicator || d.control_indicator || 'Internal Control Weakness'
    indDist[ind] = (indDist[ind] || 0) + 1

    const st = d.sourceType || 'Imported Document'
    srcDist[st] = (srcDist[st] || 0) + 1
  })

  nlpState.value.categoryDistribution = catDist
  nlpState.value.sentimentDistribution = sentDist
  nlpState.value.controlIndicatorDistribution = indDist
  nlpState.value.sourceTypeDistribution = srcDist
}

if (initialCustomDocs.length > 0 || initialCustomTax.risk_categories.length > 0 || initialCustomTax.control_indicators.length > 0) {
  recalculateNlpDistributions()
}

const liveAuditorPayloadItems = ref<AuditorPayloadItem[]>([])
const customAuditeeVersion = ref(0)
const checkingRetrainStatus = ref(false)
const lastRetrainScheduledAt = ref(0)
const retrainStatus = ref({
  status: 'idle' as 'idle' | 'running' | 'completed' | 'failed',
  isRunning: false,
  trigger: 'none',
  startedAt: null as string | null,
  finishedAt: null as string | null,
  durationSeconds: null as number | null,
  trainedAuditorHash: '',
  trainedAuditeePoolHash: '',
  trainedAuditeeCustomHash: '',
  trainedAuditorCount: 0,
  trainedAuditeeCount: 0,
  hasUntrainedAuditeePoolData: false,
  modelUpdatedAt: null as string | null,
  message: '',
  lastError: null as string | null,
  statusFetchedOnce: false
})

const createDefaultModelEntry = (key: string, label: string) => ({
  key,
  label,
  state: 'trained' as 'never_trained' | 'trained' | 'untrained_data' | 'queued' | 'training' | 'failed',
  has_new_data: false,
  new_rows: 0,
  current_rows: 0,
  trained_rows: 0,
  trained_at: null as string | null,
  duration_seconds: null as number | null,
  samples: 0,
  tier: 'T1' as string | null,
  metrics: null as Record<string, any> | null,
  last_result: 'promoted' as string | null,
  message: `Model ${label} siap digunakan.`,
  last_error: null as string | null,
  progress: null as { phase?: string; batch?: number; total_batches?: number; percent?: number; tier?: string } | null,
  queue_position: null as number | null,
  job_id: null as string | null,
  trigger: null as string | null,
  waiting_reason: null as string | null,
  peak_ram_mb: null as number | null,
  version: 1
})

const modelStatuses = ref<Record<string, ReturnType<typeof createDefaultModelEntry>>>({
  entity: createDefaultModelEntry('entity', 'Entity Prediction'),
  anomaly: createDefaultModelEntry('anomaly', 'Anomaly Detection'),
  nlp: createDefaultModelEntry('nlp', 'NLP Analysis'),
  kpi_auditee: createDefaultModelEntry('kpi_auditee', 'KPI Forecast (Auditee)'),
  kpi_auditor: createDefaultModelEntry('kpi_auditor', 'KPI Forecast (Auditor)')
})
const retrainQueue = ref<any[]>([])
const retrainResource = ref<Record<string, any> | null>(null)

const initialCustomAuditeeKpis = getPersistedCustomAuditeeKpis()
const initialAuditeeKpis = mergeAuditeeKpiForecasts(
  initialTS.kpiForecasts.map(sanitizeAuditeeKpiForecastRecord).filter(Boolean),
  initialCustomAuditeeKpis
)
const initialAuditeeAtRisk = initialTS.atRiskDepartments.map(sanitizeAuditeeAtRiskRecord)
const timeseriesState = ref<any>({
  historicalKPI: [...initialTS.historicalKPI],
  kpiForecasts: initialAuditeeKpis,
  atRiskDepartments: initialAuditeeAtRisk,
  forecastAccuracy: { ...initialTS.forecastAccuracy },
  auditorForecast: {
    kpiForecasts: [] as any[],
    atRiskPillars: [] as any[],
    timeSeriesData: [] as any[],
    executiveSummary: 'Memuat proyeksi KPI Auditor dari fitur Strategic Audit Plan dan Internal Audit Performance...'
  },
  auditeeForecast: {
    kpiForecasts: initialAuditeeKpis,
    atRiskDepartments: initialAuditeeAtRisk,
    timeSeriesData: [...initialTS.historicalKPI],
    executiveSummary: 'Diprediksi Cabang Surabaya akan mengalami lonjakan NPL dalam 2 kuartal ke depan akibat kenaikan rasio kredit berisiko. Diperlukan pemantauan kredit macet dan pengetatan threshold Continuous Auditing.'
  }
})

const CACHE_KEY = 'auditsphere_real_analytics_cache_v20'

const clearLegacyAnalyticsCaches = () => {
  if (typeof window === 'undefined') return
  const legacyKeys = [
    'auditsphere_real_analytics_cache_v19',
    'auditsphere_real_analytics_cache_v18',
    'auditsphere_real_analytics_cache_v17',
    'auditsphere_real_analytics_cache_v16',
    'auditsphere_real_analytics_cache_v15',
    'auditsphere_real_analytics_cache_v14',
    'auditsphere_real_analytics_cache_v13',
    'auditsphere_real_analytics_cache_v12',
    'auditsphere_real_analytics_cache_v11',
    'auditsphere_real_analytics_cache_v10',
    'auditsphere_real_analytics_cache_v9',
    'auditsphere_real_analytics_cache_v8',
    'auditsphere_real_analytics_cache_v6',
    'auditsphere_real_analytics_cache_v5',
    'auditsphere_real_analytics_cache_v4',
    'auditsphere_real_analytics_cache_v3',
    'auditsphere_real_analytics_cache',
    'auditsphere_custom_nlp_docs_v2',
    'auditsphere_custom_nlp_docs_v1'
  ]
  legacyKeys.forEach(k => localStorage.removeItem(k))
}

const saveRealCache = () => {
  if (typeof window === 'undefined') return
  try {
    const timeStr = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
    const cacheObj = {
      time: timeStr,
      xgboost: xgboostState.value,
      isolation: isolationState.value,
      nlp: nlpState.value,
      timeseries: timeseriesState.value
    }
    localStorage.setItem(CACHE_KEY, JSON.stringify(cacheObj))
    clearLegacyAnalyticsCaches()
    lastSyncedTime.value = timeStr
  } catch (e) {
    console.warn('Failed to save real analytics cache', e)
  }
}

const loadRealCache = (): boolean => {
  if (typeof window === 'undefined') return false
  try {
    clearLegacyAnalyticsCaches()

    const raw = localStorage.getItem(CACHE_KEY)
    if (!raw) return false
    const parsed = JSON.parse(raw)

    const bankCats = ['Funding', 'Lending', 'Treasury', 'Payment', 'KYC', 'IT Control']
    const anomList = parsed.isolation?.anomalies || []
    // Verify that the cached isolation has all 150 anomalies and covers all 6 banking categories (at least 20 per category)
    const hasAllBank = bankCats.every((cat) => anomList.filter((a: any) => a.type === cat).length >= 20)
    if (!hasAllBank || anomList.length < 150) {
      localStorage.removeItem(CACHE_KEY)
      // Restore from full 150 seed dataset
      const freshIso = useIsolationForestData()
      isolationState.value = {
        anomalies: normalizeIsolationAnomaliesList([...freshIso.anomalies]),
        scatterData: [...freshIso.scatterData],
        summary: { ...freshIso.summary }
      }
      return false
    }

    if (parsed.xgboost) {
      if (Array.isArray(parsed.xgboost.predictions) && parsed.xgboost.predictions.length > 0) {
        if (!xgboostState.value.predictions || xgboostState.value.predictions.length <= parsed.xgboost.predictions.length) {
          const cy = new Date().getFullYear()
          parsed.xgboost.predictions = parsed.xgboost.predictions
            .filter((p: any) =>
              p.entity &&
              !String(p.entity).toLowerCase().includes('legal & compliance') &&
              p.riskCategory !== 'Legal'
            )
            .map((p: any) => sanitizeEntityPredictionRecord({
              ...p,
              targetTimeline: p.targetTimeline && !p.targetTimeline.startsWith('Q') ? p.targetTimeline : `FY ${cy}`
            }))
            .filter(Boolean)
          xgboostState.value = parsed.xgboost
        }
      }
    }
    if (parsed.isolation) {
      isolationState.value = {
        ...parsed.isolation,
        anomalies: normalizeIsolationAnomaliesList(parsed.isolation.anomalies || [])
      }
    }

    // Merge persisted custom anomalies so they are never lost
    const customAnoms = getPersistedCustomAnomalies()
    if (customAnoms.length > 0) {
      const existingIds = new Set((isolationState.value.anomalies || []).map((a: any) => a.id))
      customAnoms.forEach((ca: any) => {
        if (!existingIds.has(ca.id)) {
          isolationState.value.anomalies.unshift(ca)
          existingIds.add(ca.id)
        }
      })
    }
    const customScatter = getPersistedCustomScatter()
    if (customScatter.length > 0) {
      const existingScatterLabels = new Set((isolationState.value.scatterData || []).map((s: any) => s.label || `${s.x}_${s.y}`))
      customScatter.forEach((cs: any) => {
        const key = cs.label || `${cs.x}_${cs.y}`
        if (!existingScatterLabels.has(key)) {
          isolationState.value.scatterData.push(cs)
          existingScatterLabels.add(key)
        }
      })
    }
    if (parsed.nlp) {
      const customDocs = getPersistedCustomNlpDocs()
      const customTax = getPersistedCustomNlpTaxonomy()
      nlpState.value = {
        ...parsed.nlp,
        documents: mergeNlpDocuments(parsed.nlp.documents || [], customDocs),
        taxonomy: mergeNlpTaxonomy(parsed.nlp.taxonomy || { risk_categories: [], control_indicators: [] }, customTax)
      }
      recalculateNlpDistributions()
    }
    if (parsed.timeseries) {
      const sanitizedTopKpis = Array.isArray(parsed.timeseries.kpiForecasts)
        ? parsed.timeseries.kpiForecasts.map(sanitizeAuditeeKpiForecastRecord).filter(Boolean)
        : timeseriesState.value.kpiForecasts
      const sanitizedTopAtRisk = Array.isArray(parsed.timeseries.atRiskDepartments)
        ? parsed.timeseries.atRiskDepartments.map(sanitizeAuditeeAtRiskRecord)
        : timeseriesState.value.atRiskDepartments
      const sanitizedAuditeeForecast = parsed.timeseries.auditeeForecast
        ? {
          ...parsed.timeseries.auditeeForecast,
          kpiForecasts: Array.isArray(parsed.timeseries.auditeeForecast.kpiForecasts)
            ? parsed.timeseries.auditeeForecast.kpiForecasts.map(sanitizeAuditeeKpiForecastRecord).filter(Boolean)
            : sanitizedTopKpis,
          atRiskDepartments: Array.isArray(parsed.timeseries.auditeeForecast.atRiskDepartments)
            ? parsed.timeseries.auditeeForecast.atRiskDepartments.map(sanitizeAuditeeAtRiskRecord)
            : sanitizedTopAtRisk
        }
        : undefined
      timeseriesState.value = {
        ...timeseriesState.value,
        ...parsed.timeseries,
        kpiForecasts: sanitizedTopKpis,
        atRiskDepartments: sanitizedTopAtRisk,
        ...(sanitizedAuditeeForecast ? { auditeeForecast: sanitizedAuditeeForecast } : {}),
        auditorForecast: {
          kpiForecasts: Array.isArray(parsed.timeseries.auditorForecast?.kpiForecasts)
            ? parsed.timeseries.auditorForecast.kpiForecasts
            : [],
          atRiskPillars: Array.isArray(parsed.timeseries.auditorForecast?.atRiskPillars)
            ? parsed.timeseries.auditorForecast.atRiskPillars
            : [],
          timeSeriesData: Array.isArray(parsed.timeseries.auditorForecast?.timeSeriesData)
            ? parsed.timeseries.auditorForecast.timeSeriesData
            : [],
          executiveSummary: parsed.timeseries.auditorForecast?.executiveSummary || ''
        }
      }
    }
    if (parsed.time) lastSyncedTime.value = parsed.time
    return true
  } catch (e) {
    console.warn('Failed to load real analytics cache', e)
    return false
  }
}

export const useAiAnalytics = () => {
  const config = useRuntimeConfig()

  const getAnalyticsUrl = () => getAnalyticsServiceBaseUrl()
  const getPythonAiUrl = () => getPythonAiBaseUrl()

  const safeApiFetch = async (endpoint: string, options: any = {}): Promise<any> => {
    const analyticsUrl = getAnalyticsUrl()
    const pyUrl = getPythonAiUrl()
    try {
      return await $fetch(`${analyticsUrl}${endpoint}`, options)
    } catch (e1) {
      let pyEndpoint = endpoint
      if (endpoint === '/risk-score') pyEndpoint = '/predict/risk-score'
      else if (endpoint === '/anomaly') pyEndpoint = '/predict/anomaly'
      else if (endpoint === '/text-analysis') pyEndpoint = '/predict/text-analysis'
      else if (endpoint === '/performance-trend') pyEndpoint = '/predict/performance-trend'
      else if (endpoint === '/risk-score/batch') pyEndpoint = '/predict/risk-score/batch'
      else if (endpoint === '/anomaly/batch') pyEndpoint = '/predict/anomaly/batch'
      else if (endpoint === '/text-analysis/batch') pyEndpoint = '/predict/text-analysis/batch'
      else if (endpoint === '/text-analysis/taxonomy') pyEndpoint = '/taxonomy/nlp'
      else if (endpoint === '/text-analysis/sync-sources') pyEndpoint = '/predict/text-analysis/sync-sources'
      else if (endpoint.startsWith('/performance-trend/batch')) pyEndpoint = '/predict/performance-trend/batch'
      else if (endpoint === '/performance-trend/auditor-forecast') pyEndpoint = '/predict/auditor-kpi-forecast'
      else if (endpoint.startsWith('/retrain/status')) pyEndpoint = endpoint
      else if (endpoint.startsWith('/retrain/auto')) pyEndpoint = endpoint
      else if (endpoint.startsWith('/retrain/model/')) pyEndpoint = endpoint

      try {
        const pyOptions = {
          ...options,
          headers: {
            ...(options.headers || {}),
            'X-API-Key': 'dev-ai-api-key'
          }
        }
        return await $fetch(`${pyUrl}${pyEndpoint}`, pyOptions)
      } catch (e2) {
        return null
      }
    }
  }

  const formatNum = (val: any, decimals = 1): string => {
    if (val === null || val === undefined || isNaN(Number(val))) return '0'
    return Number(val).toFixed(decimals)
  }

  // Strictly map ANY risk level (including legacy 'Critical' or custom external DB ratings)
  // onto the 5 canonical levels: High, Moderate to High, Moderate, Low to Moderate, Low.
  const getRiskConfig = (level: string) => {
    const canonicalLabel = normalizeAnomalyRiskLevel(level || 'Moderate')
    const key = normalizeRiskLevel(canonicalLabel)
    const badgeClass = getRiskLevelColorClass(key)
    switch (key) {
      case 'High': return { label: 'High', color: '#F44336', bg: '#B71C1C', badgeClass }
      case 'Medium to High': return { label: 'Moderate to High', color: '#FF9800', bg: '#E65100', badgeClass }
      case 'Medium': return { label: 'Moderate', color: '#FFC107', bg: '#FF6F00', badgeClass }
      case 'Low to Medium': return { label: 'Low to Moderate', color: '#8BC34A', bg: '#33691E', badgeClass }
      case 'Low': return { label: 'Low', color: '#4CAF50', bg: '#1B5E20', badgeClass }
      default: return { label: 'Moderate', color: '#FFC107', bg: '#FF6F00', badgeClass: getRiskLevelColorClass('Medium') }
    }
  }

  type BadgeColor = 'error' | 'primary' | 'warning' | 'success' | 'info' | 'neutral'

  const sentimentColor = (s: string): BadgeColor => {
    const map: Record<string, BadgeColor> = { Positive: 'success', Neutral: 'warning', Negative: 'error' }
    return map[s] || 'neutral'
  }

  const riskCategoryColor = (cat: string): BadgeColor => {
    const map: Record<string, BadgeColor> = {
      'Financial': 'error',
      'Technology': 'info',
      'Operational': 'warning',
      'Operations': 'warning',
      'Compliance': 'success',
      'Strategic': 'primary',
      'Reputational': 'neutral',
      'Governance': 'primary',
      'Human Resources': 'info',
      'Legal': 'warning',
      'Security': 'error',
      'Cybersecurity': 'info',
      'ESG': 'success',
      'Fraud & Security': 'error'
    }
    if (map[cat]) return map[cat]
    const dynamicPalette: BadgeColor[] = ['primary', 'info', 'warning', 'success', 'error']
    const hash = Array.from(cat || '').reduce((acc, ch) => acc + ch.charCodeAt(0), 0)
    return dynamicPalette[hash % dynamicPalette.length]
  }

  const controlIndicatorColor = (ind?: string): BadgeColor => {
    const val = (ind || '').trim()
    const map: Record<string, BadgeColor> = {
      'Fraud Indication': 'error',
      'Segregation of Duties (SoD) Violation': 'error',
      'IT & Access Control Weakness': 'warning',
      'Internal Control Weakness': 'warning',
      'Compliance Deviation': 'info',
      'Adequate Control': 'success'
    }
    if (map[val]) return map[val]
    if (val.toLowerCase().includes('fraud') || val.toLowerCase().includes('violation') || val.toLowerCase().includes('breach')) return 'error'
    if (val.toLowerCase().includes('adequate') || val.toLowerCase().includes('effective')) return 'success'
    const dynamicPalette: BadgeColor[] = ['warning', 'info', 'primary', 'error']
    const hash = Array.from(val).reduce((acc, ch) => acc + ch.charCodeAt(0), 0)
    return dynamicPalette[hash % dynamicPalette.length]
  }

  const recalculateNlpDistributions = () => {
    const docs = nlpState.value.documents || []
    const taxCats = (nlpState.value.taxonomy?.risk_categories || []).map((c: any) => c.name || c)
    const baseCats = taxCats.length > 0 ? taxCats : ['Financial', 'Operational', 'Compliance', 'Technology']
    const catDist: Record<string, number> = {}
    baseCats.forEach((c: string) => { catDist[c] = 0 })

    const sentDist = { positive: 0, neutral: 0, negative: 0 }
    const indDist: Record<string, number> = {}
    const srcDist: Record<string, number> = {
      'Imported Document': 0,
      'Internal Note': 0,
      'External Database': 0
    }

    docs.forEach((d: any) => {
      const cat = d.autoCategory || d.risk_category || 'Operational'
      catDist[cat] = (catDist[cat] || 0) + 1

      const s = String(d.sentiment || 'Neutral').toLowerCase()
      if (s === 'positive') sentDist.positive++
      else if (s === 'negative') sentDist.negative++
      else sentDist.neutral++

      const ind = d.controlIndicator || d.control_indicator || 'Internal Control Weakness'
      indDist[ind] = (indDist[ind] || 0) + 1

      const st = d.sourceType || 'Imported Document'
      srcDist[st] = (srcDist[st] || 0) + 1
    })

    nlpState.value.categoryDistribution = catDist
    nlpState.value.sentimentDistribution = sentDist
    nlpState.value.controlIndicatorDistribution = indDist
    nlpState.value.sourceTypeDistribution = srcDist
  }

  const trendIcon = (t: string) => {
    const map: Record<string, string> = {
      Improving: 'i-heroicons-arrow-trending-up',
      Declining: 'i-heroicons-arrow-trending-down',
      Deteriorating: 'i-heroicons-arrow-trending-down',
      Stable: 'i-heroicons-minus',
      up: 'i-heroicons-arrow-trending-up',
      down: 'i-heroicons-arrow-trending-down',
      stable: 'i-heroicons-minus'
    }
    return map[t] || 'i-heroicons-minus'
  }

  const trendColor = (t: string) => {
    if (['Improving', 'down'].includes(t)) return 'text-emerald-500'
    if (['Declining', 'up', 'Deteriorating'].includes(t)) return 'text-rose-500'
    return 'text-gray-400'
  }

  const currentAuditorHash = computed(() => computeAuditorFingerprint(liveAuditorPayloadItems.value))
  const currentAuditeeCustomHash = computed(() => {
    void customAuditeeVersion.value
    return computeAuditeeCustomFingerprint(getPersistedCustomAuditeeKpis())
  })

  const hasUntrainedAuditorData = computed(() => {
    if (!retrainStatus.value.statusFetchedOnce) return false
    if (liveAuditorPayloadItems.value.length === 0) return false
    return Boolean(currentAuditorHash.value && currentAuditorHash.value !== retrainStatus.value.trainedAuditorHash)
  })

  const hasUntrainedAuditeeCustomData = computed(() => {
    if (!retrainStatus.value.statusFetchedOnce) return false
    return Boolean(currentAuditeeCustomHash.value && currentAuditeeCustomHash.value !== retrainStatus.value.trainedAuditeeCustomHash)
  })

  const hasUntrainedAuditeeData = computed(() =>
    Boolean(retrainStatus.value.hasUntrainedAuditeePoolData || hasUntrainedAuditeeCustomData.value)
  )

  const hasUntrainedNewData = computed(() =>
    Boolean(hasUntrainedAuditorData.value || hasUntrainedAuditeeData.value)
  )

  const untrainedDataReasons = computed<string[]>(() => {
    const reasons: string[] = []
    if (hasUntrainedAuditorData.value) {
      reasons.push(`KPI Auditor (${liveAuditorPayloadItems.value.length} indikator Strategic Audit Plan / Performance)`)
    }
    if (hasUntrainedAuditeeCustomData.value) {
      reasons.push(`KPI Auditee External DB (${getPersistedCustomAuditeeKpis().length} indikator)`)
    }
    if (retrainStatus.value.hasUntrainedAuditeePoolData) {
      reasons.push('KPI Auditee Zona Gold / Datalake')
    }
    return reasons
  })

  const canManualRetrain = computed(() =>
    retrainStatus.value.statusFetchedOnce && !retrainStatus.value.isRunning && hasUntrainedNewData.value
  )

  const applyRetrainPayload = (raw: any) => {
    const d = raw?.data || raw
    if (!d || typeof d !== 'object') return
    const st = String(d.status || retrainStatus.value.status || 'idle')
    const running = typeof d.is_running === 'boolean'
      ? d.is_running
      : (st === 'running' || st === 'scheduled' || st === 'already_running')

    retrainStatus.value = {
      status: running ? 'running' : (st === 'failed' ? 'failed' : (st === 'completed' ? 'completed' : retrainStatus.value.status)),
      isRunning: running,
      trigger: d.trigger || retrainStatus.value.trigger,
      startedAt: d.started_at ?? retrainStatus.value.startedAt,
      finishedAt: d.finished_at ?? retrainStatus.value.finishedAt,
      durationSeconds: d.duration_seconds !== undefined ? Number(d.duration_seconds) : retrainStatus.value.durationSeconds,
      trainedAuditorHash: d.trained_auditor_hash !== undefined ? String(d.trained_auditor_hash || '') : retrainStatus.value.trainedAuditorHash,
      trainedAuditeePoolHash: d.trained_auditee_pool_hash !== undefined ? String(d.trained_auditee_pool_hash || '') : retrainStatus.value.trainedAuditeePoolHash,
      trainedAuditeeCustomHash: d.trained_auditee_custom_hash !== undefined ? String(d.trained_auditee_custom_hash || '') : retrainStatus.value.trainedAuditeeCustomHash,
      trainedAuditorCount: d.trained_auditor_count !== undefined ? Number(d.trained_auditor_count || 0) : retrainStatus.value.trainedAuditorCount,
      trainedAuditeeCount: d.trained_auditee_count !== undefined ? Number(d.trained_auditee_count || 0) : retrainStatus.value.trainedAuditeeCount,
      hasUntrainedAuditeePoolData: d.has_untrained_auditee_pool_data !== undefined
        ? Boolean(d.has_untrained_auditee_pool_data)
        : (running ? retrainStatus.value.hasUntrainedAuditeePoolData : false),
      modelUpdatedAt: d.model_updated_at ?? retrainStatus.value.modelUpdatedAt,
      message: d.message ?? retrainStatus.value.message,
      lastError: d.last_error ?? null,
      statusFetchedOnce: true
    }

    if (d.models && typeof d.models === 'object') {
      const updated = { ...modelStatuses.value }
      for (const [mk, mv] of Object.entries(d.models)) {
        if (mv && typeof mv === 'object') {
          updated[mk] = {
            ...(updated[mk] || createDefaultModelEntry(mk, mk)),
            ...(mv as any)
          }
        }
      }
      modelStatuses.value = updated
    }
    if (Array.isArray(d.queue)) {
      retrainQueue.value = d.queue
    }
    if (d.resource && typeof d.resource === 'object') {
      retrainResource.value = d.resource
    }
  }

  const retrainModel = async (
    modelKey: 'entity' | 'anomaly' | 'nlp' | 'kpi' | 'kpi_auditee' | 'kpi_auditor',
    options: { force?: boolean; wait?: boolean; trigger?: string } = {}
  ) => {
    const customAuditee = getPersistedCustomAuditeeKpis()
    const isKpi = modelKey === 'kpi' || modelKey === 'kpi_auditee' || modelKey === 'kpi_auditor'
    const payload: Record<string, any> = {
      force: options.force ?? true,
      wait: options.wait ?? false,
      trigger: options.trigger || `manual_${modelKey}`
    }
    if (isKpi) {
      payload.auditor_items = liveAuditorPayloadItems.value
      payload.auditee_items = customAuditee
    }
    const qs = new URLSearchParams()
    if (payload.force) qs.set('force', 'true')
    if (payload.wait) qs.set('wait', 'true')
    const res: any = await safeApiFetch(`/retrain/model/${modelKey}?${qs.toString()}`, {
      method: 'POST',
      body: payload
    })
    const d = res?.data || res
    if (d) {
      applyRetrainPayload(d)
    }
    await fetchRetrainStatus().catch(() => null)
    return d
  }

  const fetchRetrainStatus = async () => {
    checkingRetrainStatus.value = true
    try {
      const qs = new URLSearchParams()
      if (currentAuditorHash.value) qs.set('auditor_hash', currentAuditorHash.value)
      if (currentAuditeeCustomHash.value) qs.set('auditee_custom_hash', currentAuditeeCustomHash.value)
      qs.set('_ts', String(Date.now()))
      const res: any = await safeApiFetch(`/retrain/status?${qs.toString()}`)
      if (res) {
        applyRetrainPayload(res)
        return retrainStatus.value
      }
      retrainStatus.value.statusFetchedOnce = false
      // Preserve recently scheduled background run flag within 15s to avoid race conditions during startup
      if (!(retrainStatus.value.isRunning && Date.now() - lastRetrainScheduledAt.value < 15000)) {
        retrainStatus.value.isRunning = false
      }
      return null
    } catch {
      retrainStatus.value.statusFetchedOnce = false
      if (!(retrainStatus.value.isRunning && Date.now() - lastRetrainScheduledAt.value < 15000)) {
        retrainStatus.value.isRunning = false
      }
      return null
    } finally {
      checkingRetrainStatus.value = false
    }
  }

  const fetchAiAnalytics = async (force = false, options: { skipAutoRetrain?: boolean } = {}) => {
    if (hasFetchedOnce.value && !force) return

    try {
      loading.value = true

      // Silent background auto-retrain trigger (unless skipped right after manual force sync)
      if (!options.skipAutoRetrain) {
        safeApiFetch('/retrain/auto', { method: 'POST' })
          .then((res: any) => {
            const d = res?.data || res
            if (d?.status === 'scheduled' || d?.status === 'already_running' || d?.is_running === true) {
              lastRetrainScheduledAt.value = Date.now()
              retrainStatus.value = {
                ...retrainStatus.value,
                status: 'running',
                isRunning: true,
                trigger: d.trigger || 'auto_background',
                startedAt: d.started_at || retrainStatus.value.startedAt || new Date().toISOString(),
                message: d.message || 'Model AI sedang dilatih secara otomatis di belakang layar.'
              }
            }
          })
          .catch(() => { })
      }

      const perfEndpoint = force ? `/performance-trend/batch?_ts=${Date.now()}` : '/performance-trend/batch'
      const [resRiskBatch, resAnomalyBatch, resTextBatch, resPerfBatch]: any[] = await Promise.all([
        safeApiFetch('/risk-score/batch'),
        safeApiFetch('/anomaly/batch'),
        safeApiFetch('/text-analysis/batch'),
        safeApiFetch(perfEndpoint)
      ])

      const hasRealData = !!(resRiskBatch || resAnomalyBatch || resTextBatch || resPerfBatch)

      if (hasRealData) {
        isAiConnected.value = true
        usingCachedRealData.value = false

        if (resRiskBatch?.data || resRiskBatch?.predictions) {
          const bData = resRiskBatch.data || resRiskBatch
          if (bData.predictions && bData.predictions.length > 0) {
            // Do not downgrade if current predictions already contains a larger set from Audit Universe
            if (!xgboostState.value.predictions || xgboostState.value.predictions.length <= bData.predictions.length) {
              const now = new Date()
              const cy = now.getFullYear()
              xgboostState.value.predictions = mergeEntityPredictions(
                bData.predictions
                  .filter((p: any) => p.entity && !String(p.entity).toLowerCase().includes('legal & compliance') && p.risk_category !== 'Legal')
                  .map((p: any) => {
                    const isBr = ['cabang', 'kcp', 'regional', 'kc', 'outlet', 'branch'].some(w => String(p.entity).toLowerCase().includes(w))
                    const typ = (p.type && p.type !== 'Department') ? p.type : (isBr ? 'Branch' : 'Auditable Entity')
                    const defTimeline = `FY ${cy}`
                    return {
                      entity: p.entity || 'Unknown Entity',
                      type: typ,
                      riskCategory: p.risk_category || p.riskCategory || 'Financial',
                      targetTimeline: p.target_timeline && !p.target_timeline.startsWith('Q') ? p.target_timeline : (p.targetTimeline && !p.targetTimeline.startsWith('Q') ? p.targetTimeline : defTimeline),
                      currentLikelihood: p.current_likelihood ?? p.actual_likelihood ?? p.inherent_likelihood ?? p.currentLikelihood ?? 3,
                      currentImpact: p.current_impact ?? p.actual_impact ?? p.inherent_impact ?? p.currentImpact ?? 3,
                      predictedLikelihood: p.predicted_likelihood ?? p.predictedLikelihood ?? 3,
                      predictedImpact: p.predicted_impact ?? p.predictedImpact ?? 3,
                      predictedScore: p.predicted_score ?? p.predictedScore ?? 9,
                      actualScore: p.actual_score ?? p.actualScore ?? 9,
                      delta: p.delta ?? 0,
                      trend: p.trend || 'stable',
                      predictedRiskLevel: toEntityRiskCode(p.risk_level || p.predictedRiskLevel || 'MODERATE'),
                      actualRiskLevel: toEntityRiskCode(p.actual_risk_level || p.actualRiskLevel || 'MODERATE')
                    }
                  }),
                getPersistedCustomEntityPredictions()
              )
            }
          }
        }

        if (resAnomalyBatch?.data || resAnomalyBatch?.anomalies) {
          const bData = resAnomalyBatch.data || resAnomalyBatch
          if (bData.anomalies && bData.anomalies.length > 0) {
            const fetchedAnomalies = bData.anomalies.map((a: any) => {
              const normalizedLevel = normalizeAnomalyRiskLevel(a.risk_level || a.riskLevel || a.severity)
              return {
                id: a.id || 'ANM-001',
                entity: a.entity || a.accountNumber || a.refNumber || 'Kantor Pusat / GL',
                type: a.type || a.category || 'Funding',
                anomalyScore: a.anomaly_score ?? a.anomalyScore ?? 0.85,
                description: a.description || a.reason || a.notes || 'Transaksi Anomali',
                severity: normalizedLevel,
                date: a.date || a.timestamp?.slice(0, 10) || '2026-06-01',
                amount: a.amount ?? 0,
                xMetric: a.xMetric ?? (a.amount ? a.amount / 1000000 : 0),
                hour: a.hour ?? a.hour_of_day ?? null,
                isAnomaly: a.is_anomaly !== undefined ? a.is_anomaly : true,
                riskLevel: normalizedLevel
              }
            })

            // Merge persisted custom anomalies so dynamic categories are NEVER wiped on refresh/fetch
            const customAnoms = getPersistedCustomAnomalies()
            const customIds = new Set(customAnoms.map((ca: any) => ca.id))
            const filteredFetched = fetchedAnomalies.filter((fa: any) => !customIds.has(fa.id))
            isolationState.value.anomalies = [...customAnoms, ...filteredFetched]
          }
          if (bData.scatter_data && bData.scatter_data.length > 0) {
            const fetchedScatter = bData.scatter_data.map((s: any) => ({
              x: s.x ?? 0,
              y: s.y ?? 0,
              type: s.type,
              isAnomaly: s.is_anomaly ?? false,
              label: s.label || ''
            }))
            const customScatter = getPersistedCustomScatter()
            const customLabels = new Set(customScatter.map((cs: any) => cs.label || `${cs.x}_${cs.y}`))
            const filteredFetchedScatter = fetchedScatter.filter((fs: any) => !customLabels.has(fs.label || `${fs.x}_${fs.y}`))
            isolationState.value.scatterData = [...customScatter, ...filteredFetchedScatter]
          }
          if (bData.summary) {
            isolationState.value.summary = {
              ...bData.summary,
              anomaliesFound: isolationState.value.anomalies.length
            }
            summary.value.anomaliesDetected = isolationState.value.anomalies.length
          }
        }

        if (resTextBatch?.data || resTextBatch?.documents) {
          const bData = resTextBatch.data || resTextBatch
          if (bData.documents && bData.documents.length > 0) {
            const fetchedDocs = bData.documents.map((d: any) => sanitizeNlpDocumentRecord({
              docId: d.docId || 'WP-001',
              title: d.title || 'Document Title',
              source: d.source || 'Working Paper',
              sourceType: d.sourceType || (String(d.source || '').toLowerCase().includes('external') ? 'External Database' : (String(d.source || '').toLowerCase().includes('note') ? 'Internal Note' : 'Imported Document')),
              fileName: d.fileName || d.source || '',
              fileType: d.fileType || 'PDF',
              extractionMethod: d.extractionMethod || 'PDF Text & OCR Engine',
              autoCategory: d.risk_category || d.autoCategory || 'Operational',
              confidence: d.confidence ?? 0.92,
              sentiment: d.sentiment || 'Negative',
              severityScore: d.severityScore ?? 80,
              controlIndicator: d.control_indicator || d.controlIndicator || (d.sentiment === 'Positive' ? 'Adequate Control' : 'Internal Control Weakness'),
              extractedFinding: d.extracted_finding || d.extractedFinding || d.excerpt || '',
              extractedRecommendation: d.extracted_recommendation || d.extractedRecommendation || 'Perkuat pengendalian internal dan lakukan verifikasi berjenjang.',
              date: d.date || '2026-06-15',
              excerpt: d.excerpt || '',
              riskLevel: d.risk_level || d.riskLevel || d.risk_level_label || d.severityScore || 'HIGH'
            }))
            const customDocs = getPersistedCustomNlpDocs()
            nlpState.value.documents = mergeNlpDocuments(fetchedDocs, customDocs)
          }
          if (bData.taxonomy) {
            const customTax = getPersistedCustomNlpTaxonomy()
            nlpState.value.taxonomy = mergeNlpTaxonomy(bData.taxonomy, customTax)
          }
          recalculateNlpDistributions()
        }

        if (resPerfBatch?.data || resPerfBatch?.kpi_forecasts || resPerfBatch?.auditee_forecast || resPerfBatch?.auditor_forecast) {
          const bData = resPerfBatch.data || resPerfBatch

          // Auditee perspective (Gold Zone / CAATT + External DB Mapped KPIs)
          const auditeeSource = bData.auditee_forecast || bData
          if (auditeeSource.kpi_forecasts && auditeeSource.kpi_forecasts.length > 0) {
            const mappedAuditee = auditeeSource.kpi_forecasts
              .map((f: any) => sanitizeAuditeeKpiForecastRecord({
                kpiName: f.kpiName || f.kpi_name,
                code: f.code || 'KPI-001',
                unit: f.unit || '%',
                entity: f.entity || 'Finance Dept',
                entityType: f.entityType || f.entity_type || 'Department',
                syncFrequency: f.syncFrequency || f.sync_frequency || 'Real-time (CDC)',
                lastSyncedAt: f.lastSyncedAt || f.last_synced_at || '',
                targetHorizon: f.targetHorizon || f.target_horizon || 'Q3 2026',
                currentValue: f.currentValue ?? f.current_value ?? 0,
                forecastedValue: f.forecastedValue ?? f.forecasted_value ?? 0,
                deltaPct: f.deltaPct ?? f.delta_pct ?? 0,
                trend: f.trend || 'Stable',
                impactScore: f.impactScore ?? f.impact ?? 3,
                likelihoodScore: f.likelihoodScore ?? f.likelihood ?? 3,
                riskLevel: f.riskLevel || f.risk_level || 'LOW',
                aiNarrative: f.aiNarrative || f.ai_narrative || '',
                recommendedAction: f.recommendedAction || f.recommended_action || '',
                lastActualPeriod: f.lastActualPeriod || f.last_actual_period || '',
                history: f.history || [],
                forecast: f.forecast || []
              }))
              .filter(Boolean)

            const customAuditee = getPersistedCustomAuditeeKpis()
            const mergedAuditee = mergeAuditeeKpiForecasts(mappedAuditee, customAuditee)
            const customAtRisk = customAuditee
              .filter((k: any) => k.riskLevel === 'HIGH' || k.riskLevel === 'MODERATE_HIGH' || k.trend === 'Deteriorating')
              .map((k: any) => sanitizeAuditeeAtRiskRecord({
                department: k.entity || 'Divisi Kredit',
                kpi: k.kpiName,
                predictedDrop: `${k.deltaPct >= 0 ? '+' : ''}${k.deltaPct}% (${k.forecastedValue}${k.unit === '%' ? '%' : ' ' + k.unit})`,
                riskLevel: k.riskLevel || 'HIGH'
              }))
            const baseAtRisk = (auditeeSource.at_risk_departments || []).map(sanitizeAuditeeAtRiskRecord)
            const mergedAtRisk = [...customAtRisk, ...baseAtRisk.filter((b: any) => !customAtRisk.some((c: any) => c.kpi === b.kpi))]

            timeseriesState.value.auditeeForecast = {
              kpiForecasts: mergedAuditee,
              atRiskDepartments: mergedAtRisk,
              timeSeriesData: auditeeSource.time_series_data || auditeeSource.historical_kpi || [],
              executiveSummary: auditeeSource.executive_summary || ''
            }
            // backward compatibility
            timeseriesState.value.kpiForecasts = mergedAuditee
            timeseriesState.value.atRiskDepartments = mergedAtRisk
            timeseriesState.value.historicalKPI = auditeeSource.time_series_data || auditeeSource.historical_kpi || []
            if (auditeeSource.forecast_accuracy || bData.forecast_accuracy) {
              const acc = auditeeSource.forecast_accuracy || bData.forecast_accuracy
              timeseriesState.value.forecastAccuracy = {
                mape: acc.mape ?? 3.82,
                rmse: acc.rmse ?? 1.45,
                r2Score: acc.r2Score ?? acc.r2_score ?? 0.942
              }
            }
          }
        }

        // Always synchronize Auditor KPI perspective strictly with live Strategic Audit Plan & Internal Audit Performance
        await syncAuditorKpiWithStrategicPlan()
        await fetchRetrainStatus().catch(() => null)

        saveRealCache()
      } else {
        isAiConnected.value = false
        const loaded = loadRealCache()
        usingCachedRealData.value = loaded
        await syncAuditorKpiWithStrategicPlan().catch(() => null)
        await fetchRetrainStatus().catch(() => null)
      }

      hasFetchedOnce.value = true
    } catch (err: any) {
      error.value = err.message || 'Error loading AI analytics'
      isAiConnected.value = false
      const loaded = loadRealCache()
      usingCachedRealData.value = loaded
    } finally {
      loading.value = false
    }
  }

  const analyzeDocumentOrText = async (payload: {
    title?: string
    text?: string
    source?: string
    sourceType?: string
    fileName?: string
    fileType?: string
    fileContentBase64?: string
    riskCategory?: string
    riskLevel?: string
    severity?: string
  }) => {
    const normalizedExtRiskCode = payload.riskLevel
      ? toNlpRiskCode(payload.riskLevel)
      : (payload.severity ? toNlpRiskCode(payload.severity) : undefined)
    const reqBody = {
      ...payload,
      ...(normalizedExtRiskCode ? { riskLevel: normalizedExtRiskCode } : {}),
      persist: true
    }
    const res = await safeApiFetch('/text-analysis', {
      method: 'POST',
      body: reqBody
    })

    const rawDoc = res?.data || res
    let mappedDoc: any

    const resolveDynamicCategory = (text: string, title?: string, fileName?: string, explicitCat?: string): string => {
      if (explicitCat) return explicitCat
      const txt = `${title || ''} ${fileName || ''} ${text || ''}`.toLowerCase()
      if (
        txt.includes('keamanan') || txt.includes('security') || txt.includes('server') ||
        txt.includes('jaringan') || txt.includes('network') || txt.includes('cyber') ||
        txt.includes('akses') || txt.includes('access') || txt.includes('firewall') ||
        txt.includes('database') || txt.includes('patch') || txt.includes('malware') ||
        txt.includes('sistem') || txt.includes('system') || txt.includes('infrastruktur') ||
        txt.includes('infrastructure') || txt.includes('it control') || /\b(it|ti)\b/.test(txt) ||
        txt.includes('spesifikasi') || txt.includes('specification') || txt.includes('resource') ||
        txt.includes('arsitektur') || txt.includes('architecture') || txt.includes('hardware') ||
        txt.includes('software') || txt.includes('cpu') || txt.includes('ram') ||
        txt.includes('storage') || txt.includes('kapasitas') || txt.includes('capacity') ||
        txt.includes('auditsphere') || txt.includes('cloud') || txt.includes('konfigurasi')
      ) {
        return 'Technology'
      }
      if (
        txt.includes('pojk') || txt.includes('ojk') || txt.includes('regulasi') ||
        txt.includes('kepatuhan') || txt.includes('compliance') || txt.includes('sop') ||
        txt.includes('ppatk') || txt.includes('aml') || txt.includes('kyc') ||
        txt.includes('hukum') || txt.includes('legal') || /\b(bi)\b/.test(txt)
      ) {
        return 'Compliance'
      }
      if (
        txt.includes('kas') || txt.includes('keuangan') || txt.includes('kredit') ||
        txt.includes('pembayaran') || txt.includes('invoice') || txt.includes('jurnal') ||
        txt.includes('buku besar') || txt.includes('pajak') || txt.includes('financial') ||
        txt.includes('voucher') || txt.includes('disbursement') || txt.includes('piutang') ||
        txt.includes('utang')
      ) {
        return 'Financial'
      }
      return 'Operational'
    }

    const resolveDynamicRecommendation = (cat: string, indicator: string, isFraud = false, isSoD = false): string => {
      if (isFraud || indicator === 'Fraud Indication') {
        return 'Lakukan audit investigatif khusus (forensik), amankan jejak audit (audit trail), bekukan sementara transaksi terkait, dan eskalasikan ke Unit Anti-Fraud.'
      }
      if (isSoD || indicator === 'Segregation of Duties (SoD) Violation') {
        return 'Terapkan pemisahan fungsi (Maker-Checker) secara sistemik, cabut hak override tunggal, dan terapkan otorisasi berjenjang.'
      }
      if (cat === 'Technology' || indicator === 'IT & Access Control Weakness') {
        if (combinedTxt.includes('spesifikasi') || combinedTxt.includes('resource') || combinedTxt.includes('kapasitas')) {
          return 'Lakukan capacity planning dan uji beban berkala, pastikan alokasi resource CPU, RAM, & storage sesuai proyeksi beban kerja, serta optimalkan arsitektur server.'
        }
        return 'Lakukan User Access Review (UAR) berkala, nonaktifkan akun dormant/shared account, terapkan patch keamanan terbaru, serta aktifkan pemantauan log akses real-time.'
      }
      if (cat === 'Compliance' || indicator === 'Compliance Deviation') {
        return 'Lakukan pemenuhan regulasi dan SOP terkait segera, lengkapi dokumen kepatuhan, serta perbarui template pelaporan berkala.'
      }
      if (cat === 'Financial') {
        return 'Perkuat verifikasi bukti pendukung transaksi keuangan, lakukan rekonsiliasi independen secara berkala, dan patuhi limit otorisasi bertingkat.'
      }
      return 'Tingkatkan pengendalian internal operasional, lakukan evaluasi berkala atas proses bisnis, dan dokumentasikan tindak lanjut perbaikan.'
    }

    const isBadFallbackFinding = (str?: string): boolean => {
      if (!str) return false
      if (!isReadableHumanText(str)) return true
      const s = str.toLowerCase()
      return s.includes('kelemahan pengendalian internal pada otorisasi kas') ||
        s.includes('otorisasi kas') ||
        s.includes('perkuat verifikasi dokumen pendukung dan terapkan otorisasi berjenjang') ||
        s.includes('berhasil diekstrak dan dianalisis dalam kategori') ||
        s.includes('berhasil diekstrak dan dianalisis')
    }

    const isBadTitle = (t?: string): boolean => {
      if (!t) return true
      const tl = t.trim().toLowerCase()
      return isBadFallbackFinding(tl) || tl.startsWith('ditemukan indikasi') || tl.startsWith('hasil ekstraksi')
    }

    const cleanTitle = (!isBadTitle(rawDoc?.title) ? rawDoc?.title : '') ||
      payload.title?.trim() ||
      payload.fileName?.replace(/\.[^/.]+$/, '').replace(/[_-]/g, ' ') ||
      'Dokumen Audit Teranalisis'

    const combinedTxt = `${cleanTitle} ${payload.fileName || ''} ${payload.text || ''}`.toLowerCase()
    const isFraud = combinedTxt.includes('fiktif') || combinedTxt.includes('fraud') || combinedTxt.includes('manipulasi') || combinedTxt.includes('penggelapan') || combinedTxt.includes('split') || combinedTxt.includes('mencurigakan')
    const isSoD = combinedTxt.includes('dual control') || combinedTxt.includes('maker-checker') || combinedTxt.includes('override') || combinedTxt.includes('pejabat yang sama')
    const isPos = combinedTxt.includes('memadai') || combinedTxt.includes('efektif') || combinedTxt.includes('sesuai prosedur')

    const heuristicCategory = resolveDynamicCategory(payload.text || '', cleanTitle, payload.fileName, payload.riskCategory)
    let resolvedCategory = heuristicCategory

    if (payload.riskCategory) {
      resolvedCategory = payload.riskCategory
    } else if (rawDoc?.risk_category && !isBadFallbackFinding(rawDoc.extracted_finding || rawDoc.title) && rawDoc.risk_category !== 'Financial') {
      resolvedCategory = rawDoc.risk_category
    } else if (rawDoc?.risk_category && rawDoc.risk_category === 'Financial') {
      if (heuristicCategory === 'Financial') {
        resolvedCategory = 'Financial'
      } else {
        resolvedCategory = heuristicCategory
      }
    }

    const defaultIndicator = isPos
      ? 'Adequate Control'
      : (isFraud
        ? 'Fraud Indication'
        : (isSoD
          ? 'Segregation of Duties (SoD) Violation'
          : (resolvedCategory === 'Technology'
            ? 'IT & Access Control Weakness'
            : (resolvedCategory === 'Compliance' ? 'Compliance Deviation' : 'Internal Control Weakness'))))

    let resolvedIndicator = defaultIndicator
    if (rawDoc?.control_indicator && !isBadFallbackFinding(rawDoc.control_indicator)) {
      resolvedIndicator = rawDoc.control_indicator
    } else if (rawDoc?.controlIndicator && !isBadFallbackFinding(rawDoc.controlIndicator)) {
      resolvedIndicator = rawDoc.controlIndicator
    }

    const defaultRec = resolveDynamicRecommendation(resolvedCategory, resolvedIndicator, isFraud, isSoD)

    // Guaranteed Unique Document ID to prevent overwriting
    const uniqueDocId = `DOC-${Date.now()}-${Math.floor(Math.random() * 90000 + 10000)}`
    let resolvedDocId = uniqueDocId
    if (rawDoc?.docId && rawDoc.docId !== 'WP-2026-99' && !nlpState.value.documents.some((d: any) => d.docId === rawDoc.docId)) {
      resolvedDocId = rawDoc.docId
    }

    const extractFindingClause = (docText: string, cat: string): string => {
      const clean = docText.replace(/\s+/g, ' ').trim()
      const sentences = clean
        .split(/(?<=[.!?])\s+|\n+/)
        .map(s => s.trim())
        .filter(s => s.length >= 15)
      if (sentences.length === 0) return clean.slice(0, 260)

      const markers = [
        'temuan', 'kondisi', 'ditemukan', 'terdapat', 'kelemahan', 'indikasi', 'belum',
        'tidak', 'selisih', 'penyimpangan', 'pelanggaran', 'keterlambatan', 'fiktif',
        'spesifikasi', 'resource', 'kapasitas', 'utilisasi', 'arsitektur', 'server',
        'cpu', 'ram', 'memory', 'storage', 'override', 'anomali', 'bypass', 'warning'
      ]

      for (const sent of sentences) {
        const sl = sent.toLowerCase()
        if (markers.some(m => sl.includes(m))) {
          if (cat === 'Technology' && (sl.includes('spesifikasi') || sl.includes('resource') || sl.includes('kapasitas')) && !sl.startsWith('temuan') && !sl.startsWith('observasi')) {
            return `Observasi Teknis & Spesifikasi: ${sent}`
          }
          return sent
        }
      }
      return sentences[0]
    }

    let findingText = ''
    if (rawDoc?.extracted_finding && !isBadFallbackFinding(rawDoc.extracted_finding)) {
      findingText = rawDoc.extracted_finding
    } else if (rawDoc?.extractedFinding && !isBadFallbackFinding(rawDoc.extractedFinding)) {
      findingText = rawDoc.extractedFinding
    } else if (payload.text && payload.text.trim()) {
      findingText = extractFindingClause(payload.text, resolvedCategory)
    } else if (rawDoc?.excerpt && !isBadFallbackFinding(rawDoc.excerpt)) {
      findingText = rawDoc.excerpt
    }

    if (!findingText || isBadFallbackFinding(findingText)) {
      if (resolvedCategory === 'Technology') {
        if (combinedTxt.includes('spesifikasi') || combinedTxt.includes('resource') || combinedTxt.includes('kapasitas')) {
          findingText = `Observasi Teknis & Spesifikasi: Tinjauan alokasi parameter resource, kapasitas sistem, dan arsitektur infrastruktur pada dokumen ${cleanTitle}.`
        } else {
          findingText = `Ditemukan beberapa indikasi kelemahan konfigurasi hak akses pengguna dan parameter keamanan sistem pada dokumen ${cleanTitle}.`
        }
      } else if (resolvedCategory === 'Compliance') {
        findingText = `Observasi Kepatuhan & Regulasi: Tinjauan pemenuhan standar prosedur operasional dan kepatuhan regulasi pada dokumen ${cleanTitle}.`
      } else if (isFraud) {
        findingText = `Indikasi Fraud & Anomali: Tinjauan transaksi mencurigakan dan potensi manipulasi data pada dokumen ${cleanTitle}.`
      } else if (resolvedCategory === 'Financial') {
        findingText = `Observasi Pengendalian Keuangan: Tinjauan keandalan bukti transaksi dan batas otorisasi keuangan pada dokumen ${cleanTitle}.`
      } else {
        findingText = `Observasi Operasional: Tinjauan alur proses bisnis dan kepatuhan pengendalian internal pada dokumen ${cleanTitle}.`
      }
    }

    let recText = defaultRec
    if (rawDoc?.extracted_recommendation && !isBadFallbackFinding(rawDoc.extracted_recommendation)) {
      recText = rawDoc.extracted_recommendation
    } else if (rawDoc?.extractedRecommendation && !isBadFallbackFinding(rawDoc.extractedRecommendation)) {
      recText = rawDoc.extractedRecommendation
    }

    let finalExcerpt = ''
    if (rawDoc?.excerpt && !isBadFallbackFinding(rawDoc.excerpt)) {
      finalExcerpt = rawDoc.excerpt
    } else if (payload.text && !isBadFallbackFinding(payload.text)) {
      finalExcerpt = payload.text
    } else if (rawDoc?.text && !isBadFallbackFinding(rawDoc.text)) {
      finalExcerpt = rawDoc.text
    } else {
      finalExcerpt = findingText
    }

    const fallbackRiskCode = isPos ? 'LOW' : (isFraud || isSoD || resolvedCategory === 'Technology' || resolvedCategory === 'Financial' ? 'HIGH' : 'MODERATE_HIGH')
    const rawRiskInput = rawDoc?.risk_level ?? rawDoc?.riskLevel ?? rawDoc?.risk_level_label ?? rawDoc?.severity ?? normalizedExtRiskCode ?? fallbackRiskCode

    mappedDoc = sanitizeNlpDocumentRecord({
      docId: resolvedDocId,
      title: cleanTitle,
      source: rawDoc?.source || payload.source || 'Working Paper',
      sourceType: rawDoc?.sourceType || payload.sourceType || 'Imported Document',
      fileName: rawDoc?.fileName || payload.fileName || `${cleanTitle.replace(/\s+/g, '_')}.${payload.fileType?.toLowerCase() || 'pdf'}`,
      fileType: rawDoc?.fileType || payload.fileType || 'PDF',
      extractionMethod: rawDoc?.extractionMethod || (payload.fileType === 'DOCX' ? 'Client Word (DOCX) Parser' : (payload.fileType === 'XLSX' ? 'Client Excel (XLSX) Parser' : (payload.fileContentBase64 ? 'Client PDF Stream & Text Engine' : 'Direct Text Input'))),
      autoCategory: resolvedCategory,
      confidence: rawDoc?.confidence ?? 0.94,
      sentiment: rawDoc?.sentiment || (isPos ? 'Positive' : 'Negative'),
      severityScore: rawDoc?.severityScore ?? (isPos ? 24 : (isFraud ? 92 : (resolvedCategory === 'Technology' ? 84 : 80))),
      controlIndicator: resolvedIndicator,
      extractedFinding: findingText,
      extractedRecommendation: recText,
      date: rawDoc?.date || new Date().toISOString().slice(0, 10),
      excerpt: finalExcerpt.slice(0, 340),
      riskLevel: toNlpRiskCode(rawRiskInput, isPos ? 'Low' : 'High')
    })

    nlpState.value.documents = [
      mappedDoc,
      ...(nlpState.value.documents || []).filter((d: any) => d.docId !== mappedDoc.docId)
    ]
    savePersistedCustomNlpDoc(mappedDoc)
    recalculateNlpDistributions()
    saveRealCache()
    return mappedDoc
  }

  const addNlpTaxonomyItem = async (payload: {
    itemType: 'risk_category' | 'control_indicator'
    name: string
    keywords: string[]
    description?: string
    severityBoost?: number
  }) => {
    const cleanName = payload.name.trim()
    if (!cleanName) return null

    const res = await safeApiFetch('/text-analysis/taxonomy', {
      method: 'POST',
      body: payload
    })

    const updatedTax = res?.data?.taxonomy || res?.taxonomy
    if (updatedTax) {
      nlpState.value.taxonomy = updatedTax
    } else {
      if (!nlpState.value.taxonomy) {
        nlpState.value.taxonomy = { risk_categories: [], control_indicators: [] }
      }
      if (payload.itemType === 'risk_category') {
        const exists = (nlpState.value.taxonomy.risk_categories || []).some((c: any) => (c.name || c).toLowerCase() === cleanName.toLowerCase())
        if (!exists) {
          nlpState.value.taxonomy.risk_categories.push({
            name: cleanName,
            is_default: false,
            description: payload.description || '',
            keywords: payload.keywords
          })
        }
      } else {
        const exists = (nlpState.value.taxonomy.control_indicators || []).some((c: any) => (c.name || c).toLowerCase() === cleanName.toLowerCase())
        if (!exists) {
          nlpState.value.taxonomy.control_indicators.push({
            name: cleanName,
            is_default: false,
            description: payload.description || '',
            keywords: payload.keywords,
            severity_boost: payload.severityBoost ?? 12
          })
        }
      }
    }

    savePersistedCustomNlpTaxonomy(nlpState.value.taxonomy)
    recalculateNlpDistributions()
    saveRealCache()
    return nlpState.value.taxonomy
  }

  const syncNlpMultiSources = async () => {
    const res = await safeApiFetch('/text-analysis/sync-sources', { method: 'POST' })
    const bData = res?.data || res
    if (bData?.documents && bData.documents.length > 0) {
      const mappedSynced = bData.documents.map((d: any) => sanitizeNlpDocumentRecord({
        docId: d.docId || 'WP-001',
        title: d.title || 'Document Title',
        source: d.source || 'Working Paper',
        sourceType: d.sourceType || 'Imported Document',
        fileName: d.fileName || d.source || '',
        fileType: d.fileType || 'PDF',
        extractionMethod: d.extractionMethod || 'PDF Text & OCR Engine',
        autoCategory: d.risk_category || d.autoCategory || 'Operational',
        confidence: d.confidence ?? 0.92,
        sentiment: d.sentiment || 'Negative',
        severityScore: d.severityScore ?? 80,
        controlIndicator: d.control_indicator || d.controlIndicator || 'Internal Control Weakness',
        extractedFinding: d.extracted_finding || d.extractedFinding || d.excerpt || '',
        extractedRecommendation: d.extracted_recommendation || d.extractedRecommendation || '',
        date: d.date || '2026-06-15',
        excerpt: d.excerpt || '',
        riskLevel: d.risk_level || d.riskLevel || d.risk_level_label || d.severityScore || 'HIGH'
      }))
      const customDocs = getPersistedCustomNlpDocs()
      nlpState.value.documents = mergeNlpDocuments(mappedSynced, customDocs)
      if (bData.taxonomy) {
        const customTax = getPersistedCustomNlpTaxonomy()
        nlpState.value.taxonomy = mergeNlpTaxonomy(bData.taxonomy, customTax)
      }
      recalculateNlpDistributions()
      saveRealCache()
    } else {
      await fetchAiAnalytics(true)
    }
    await fetchRetrainStatus().catch(() => null)
    return bData
  }

  const deleteNlpDocument = async (docId: string) => {
    if (!docId) return
    if (nlpState.value?.documents) {
      nlpState.value.documents = nlpState.value.documents.filter((d: any) => d.docId !== docId)
    }
    deletePersistedCustomNlpDoc(docId)
    recalculateNlpDistributions()
    saveRealCache()
    try {
      await safeApiFetch(`/text-analysis/document/${docId}`, { method: 'DELETE' })
    } catch {
      // local deletion takes effect immediately
    }
  }

  const clearAllCustomNlpDocuments = async () => {
    clearAllPersistedCustomNlpDocs()
    clearLegacyAnalyticsCaches()
    const baselineIds = [
      'WP-2026-001', 'LHA-2026-002', 'EXT-DB-003', 'EXT-DB-004', 'NOTE-FW-005', 'PKPT-2026-006', 'QA-2026-007',
      'WP-001', 'LHA-002', 'PKPT-003', 'NOTE-004'
    ]
    if (nlpState.value?.documents) {
      nlpState.value.documents = nlpState.value.documents.filter((d: any) =>
        baselineIds.includes(d.docId) || (d.docId && !d.docId.startsWith('DOC-') && !d.docId.startsWith('CUSTOM-'))
      )
    }
    recalculateNlpDistributions()
    saveRealCache()
    try {
      await safeApiFetch('/text-analysis/reset', { method: 'POST' })
    } catch {
      // local reset takes effect immediately
    }
  }

  const addCustomAnomalyRecord = (anom: any, scatter?: any) => {
    if (!anom) return null

    const normalizedLevel = normalizeAnomalyRiskLevel(anom.riskLevel || anom.risk_level || anom.severity || 'High')
    const record = {
      id: anom.id || `ANM-${Date.now().toString().slice(-4)}`,
      entity: anom.entity || 'Kantor Pusat / GL',
      type: anom.type || anom.category || 'Custom',
      category: anom.type || anom.category || 'Custom',
      anomalyScore: anom.anomalyScore ?? 0.88,
      description: anom.description || anom.reason || 'Transaksi Anomali Terdeteksi',
      severity: normalizedLevel,
      date: anom.date || new Date().toISOString().slice(0, 10),
      amount: anom.amount ?? (anom.xMetric ? anom.xMetric * 1000000 : 0),
      xMetric: anom.xMetric ?? (anom.amount ? anom.amount / 1000000 : 15),
      hour: anom.hour ?? anom.hour_of_day ?? null,
      hour_of_day: anom.hour_of_day ?? anom.hour ?? null,
      isAnomaly: true,
      riskLevel: normalizedLevel
    }

    const scatterPoint = scatter || {
      x: record.xMetric,
      y: Math.min(80, Math.max(20, Math.round((record.anomalyScore || 0.85) * 60 + 10))),
      type: record.type,
      isAnomaly: true,
      label: record.id
    }

    savePersistedCustomAnomaly(record, scatterPoint)

    if (isolationState.value) {
      if (!Array.isArray(isolationState.value.anomalies)) {
        isolationState.value.anomalies = []
      }
      const existingIdx = isolationState.value.anomalies.findIndex((a: any) => a.id === record.id)
      if (existingIdx >= 0) {
        isolationState.value.anomalies[existingIdx] = record
      } else {
        isolationState.value.anomalies.unshift(record)
      }

      if (!Array.isArray(isolationState.value.scatterData)) {
        isolationState.value.scatterData = []
      }
      const sExistingIdx = isolationState.value.scatterData.findIndex((s: any) => s.label === scatterPoint.label)
      if (sExistingIdx >= 0) {
        isolationState.value.scatterData[sExistingIdx] = scatterPoint
      } else {
        isolationState.value.scatterData.push(scatterPoint)
      }

      if (isolationState.value.summary) {
        isolationState.value.summary.anomaliesFound = isolationState.value.anomalies.length
        isolationState.value.summary.totalScanned = (isolationState.value.summary.totalScanned || 1200) + 1
      }
      summary.value.anomaliesDetected = isolationState.value.anomalies.length
    }

    saveRealCache()
    return record
  }

  const removeCustomAnomalyRecord = (id: string) => {
    if (typeof window !== 'undefined') {
      const anomList = getPersistedCustomAnomalies().filter(a => a.id !== id)
      localStorage.setItem(CUSTOM_ANOMALIES_KEY, JSON.stringify(anomList))
      const scatterList = getPersistedCustomScatter().filter(s => s.label !== id)
      localStorage.setItem(CUSTOM_SCATTER_KEY, JSON.stringify(scatterList))
    }
    if (isolationState.value?.anomalies) {
      isolationState.value.anomalies = isolationState.value.anomalies.filter((a: any) => a.id !== id)
    }
    if (isolationState.value?.scatterData) {
      isolationState.value.scatterData = isolationState.value.scatterData.filter((s: any) => s.label !== id)
    }
    if (isolationState.value?.summary) {
      isolationState.value.summary.anomaliesFound = isolationState.value.anomalies.length
    }
    summary.value.anomaliesDetected = isolationState.value?.anomalies?.length || 0
    saveRealCache()
  }

  const addCustomScatterBatch = (points: any[]) => {
    savePersistedCustomScatterBatch(points)
    if (isolationState.value?.scatterData && Array.isArray(points)) {
      for (const pt of points) {
        const idx = isolationState.value.scatterData.findIndex((s: any) => s.label === pt.label)
        if (idx >= 0) {
          isolationState.value.scatterData[idx] = pt
        } else {
          isolationState.value.scatterData.push(pt)
        }
      }
    }
    saveRealCache()
  }

  const trainAnomalyWithRules = async (payload: {
    connection_name?: string
    table_name: string
    category?: string
    rules: any[]
    sample_records?: any[]
  }) => {
    loading.value = true
    try {
      // 1. Persist rules locally (strictly normalized to the 5 risk levels) and update reactive state
      const existingRules = getPersistedAutomatedRules()
      const sanitizedRules = (payload.rules || []).map((r: any, idx: number) => {
        const baseRule = typeof r === 'string'
          ? { id: `rule-${Date.now()}-${idx}`, name: r, severity: 'High' }
          : r
        return {
          ...baseRule,
          severity: normalizeAnomalyRiskLevel(baseRule.severity || baseRule.riskLevel || 'High'),
          table_name: payload.table_name,
          connection_name: payload.connection_name,
          category: payload.category || 'Transaction',
          trained_at: new Date().toISOString()
        }
      })

      // Also normalize any external database sample records' risk level/severity before sending to AI training
      const sanitizedSamples = Array.isArray(payload.sample_records)
        ? payload.sample_records.map((rec: any) => {
          if (!rec || typeof rec !== 'object') return rec
          const extRisk = rec.risk_level ?? rec.riskLevel ?? rec.severity ?? rec.risk_rating ?? rec.tingkat_risiko ?? rec.severity_score
          if (extRisk === undefined || extRisk === null || extRisk === '') return rec
          const normExt = normalizeAnomalyRiskLevel(extRisk)
          return { ...rec, risk_level: normExt, severity: normExt }
        })
        : payload.sample_records

      // Filter out duplicate rules by name & table
      const filteredExisting = existingRules.filter(
        (er: any) => !(er.table_name === payload.table_name && sanitizedRules.some(nr => nr.name === er.name))
      )
      const mergedRules = [...sanitizedRules, ...filteredExisting]
      automatedAnomalyRules.value = mergedRules
      savePersistedAutomatedRules(mergedRules)

      // 2. Call backend AI service to retrain with 5-tier normalized rules & samples
      const res = await safeApiFetch('/predict/anomaly/train-rules', {
        method: 'POST',
        body: {
          ...payload,
          rules: sanitizedRules,
          sample_records: sanitizedSamples
        }
      })

      // 3. If response returns fresh anomalies and scatter points, merge them
      const bData = res?.data || res
      if (bData?.anomalies && bData.anomalies.length > 0) {
        const fetchedAnomalies = bData.anomalies.map((a: any) => {
          const normalizedLevel = normalizeAnomalyRiskLevel(a.risk_level || a.riskLevel || a.severity)
          return {
            id: a.id || 'ANM-001',
            entity: a.entity || 'Kantor Operasional / Core Banking',
            refNumber: a.id || 'TRX-001',
            accountNumber: a.accountNumber || a.id || 'ACC-881923',
            amount: a.amount || 0,
            type: getCanonicalCategory(a.type),
            category: getCanonicalCategory(a.category || a.type),
            anomalyScore: a.anomaly_score || 0.85,
            description: a.description || 'Transaksi anomali terdeteksi',
            reason: a.reason || a.description || 'Penyimpangan aturan operasional',
            date: a.date || new Date().toISOString().slice(0, 10),
            severity: normalizedLevel,
            riskLevel: normalizedLevel,
            xMetric: a.xMetric ?? (a.amount ? a.amount / 1000000 : 25.0),
            hour: a.hour ?? a.hour_of_day ?? 12,
            hour_of_day: a.hour ?? a.hour_of_day ?? 12,
            isAnomaly: true
          }
        })

        // Save generated rule anomalies to persistent state
        fetchedAnomalies
          .filter((fa: any) => fa.id && fa.id.startsWith('RULE-'))
          .forEach((ra: any) => {
            savePersistedCustomAnomaly(ra)
          })

        const customAnoms = getPersistedCustomAnomalies()
        const customIds = new Set(customAnoms.map((ca: any) => ca.id))
        const filteredFetched = fetchedAnomalies.filter((fa: any) => !customIds.has(fa.id))
        isolationState.value.anomalies = [...customAnoms, ...filteredFetched]
      }

      if (bData?.scatter_data && bData.scatter_data.length > 0) {
        const fetchedScatter = bData.scatter_data.map((s: any) => ({
          x: s.x ?? 0,
          y: s.y ?? 0,
          type: getCanonicalCategory(s.type),
          isAnomaly: s.is_anomaly || false,
          label: s.label || `SCATTER-${s.x}_${s.y}`
        }))
        const customScatter = getPersistedCustomScatter()
        const customLabels = new Set(customScatter.map((cs: any) => cs.label || `${cs.x}_${cs.y}`))
        const filteredFetchedScatter = fetchedScatter.filter((fs: any) => !customLabels.has(fs.label || `${fs.x}_${fs.y}`))
        isolationState.value.scatterData = [...customScatter, ...filteredFetchedScatter]
      }

      if (bData?.summary) {
        isolationState.value.summary = {
          ...isolationState.value.summary,
          ...bData.summary,
          anomaliesFound: isolationState.value.anomalies.length
        }
      }

      saveRealCache()
      await fetchRetrainStatus().catch(() => null)
      return {
        success: true,
        message: bData?.message || `Berhasil melatih model AI dengan ${payload.rules.length} aturan anomali.`,
        data: bData
      }
    } catch (err: any) {
      console.error('Failed to train anomaly with rules:', err)
      return {
        success: false,
        message: err.message || 'Gagal melatih model AI dengan aturan anomali.'
      }
    } finally {
      loading.value = false
    }
  }

  const fetchPaginatedAuditItems = async (baseEndpoint: string): Promise<{ items: any[]; ok: boolean }> => {
    const extractItems = (res: any): any[] => {
      if (Array.isArray(res?.data?.items)) return res.data.items
      if (Array.isArray(res?.items)) return res.items
      if (Array.isArray(res?.data)) return res.data
      if (Array.isArray(res)) return res
      return []
    }
    try {
      const separator = baseEndpoint.includes('?') ? '&' : '?'
      const page1: any = await $fetch(`${baseEndpoint}${separator}page=1&page_size=100`)
      if (!page1) return { items: [], ok: false }
      const collected = [...extractItems(page1)]
      const totalPages = Math.min(20, Number(page1?.data?.pagination?.total_pages || page1?.pagination?.total_pages || 1))
      if (totalPages > 1) {
        const restPages = await Promise.all(
          Array.from({ length: totalPages - 1 }, (_, i) =>
            $fetch(`${baseEndpoint}${separator}page=${i + 2}&page_size=100`).catch(() => null)
          )
        )
        for (const pRes of restPages) {
          if (pRes) collected.push(...extractItems(pRes))
        }
      }
      return { items: collected, ok: true }
    } catch {
      return { items: [], ok: false }
    }
  }

  const syncAuditorKpiWithStrategicPlan = async () => {
    try {
      const auditBase = getAuditServiceBaseUrl()
      const currentYear = new Date().getFullYear()

      // Safely sync Pinia store if initialized (unit tests may invoke composable without Pinia)
      let storePlans: any[] = []
      try {
        const strategicStore = useStrategicPlanStore()
        if (!strategicStore.strategicObjectives?.length && typeof strategicStore.fetchStrategicPlans === 'function') {
          await strategicStore.fetchStrategicPlans().catch(() => { })
        }
        if (Array.isArray(strategicStore.strategicObjectives)) {
          storePlans = [...strategicStore.strategicObjectives]
        }
      } catch {
        storePlans = []
      }

      const [plansResult, perfResult] = await Promise.all([
        fetchPaginatedAuditItems(`${auditBase}/strategic-plans?order=code%20ASC`),
        fetchPaginatedAuditItems(`${auditBase}/performance/kpi`)
      ])

      const rawPlans = plansResult.ok
        ? (plansResult.items.length > 0 ? plansResult.items : (storePlans.length > 0 && storePlans.length <= 100 ? [] : storePlans))
        : storePlans
      const rawPerfAchievements = perfResult.ok ? perfResult.items : []

      const payloadItems = dedupAndBuildAuditorPayloadItems(rawPlans, rawPerfAchievements, currentYear)
      liveAuditorPayloadItems.value = payloadItems

      if (payloadItems.length === 0) {
        timeseriesState.value.auditorForecast = {
          kpiForecasts: [],
          atRiskPillars: [],
          timeSeriesData: [],
          executiveSummary: 'Belum ada data KPI (dengan kolom Target & Actual) dari fitur Strategic Audit Plan ataupun Internal Audit Performance untuk diproyeksikan.'
        }
        saveRealCache()
        return timeseriesState.value.auditorForecast
      }

      // Evaluate live items through the pre-trained Auditor PyTorch LSTM model
      const aiRes: any = await safeApiFetch('/performance-trend/auditor-forecast', {
        method: 'POST',
        body: { items: payloadItems }
      }).catch(() => null)

      const aiData = aiRes?.data || aiRes
      if (aiData && Array.isArray(aiData.kpi_forecasts) && aiData.kpi_forecasts.length > 0) {
        const sourceByCode = new Map(payloadItems.map(p => [p.code, p.source]))
        const kpiForecasts = aiData.kpi_forecasts.map((k: any, idx: number) => ({
          code: k.code || `SO-IA${String(idx + 1).padStart(2, '0')}`,
          kpiName: k.kpiName || k.kpi_name || 'Indikator Audit',
          category: k.category || 'Pelaksanaan Audit & PKAT',
          strategicObjective: k.strategicObjective || k.strategic_objective || k.kpiName || '',
          hibHig: String(k.hibHig || k.hib_hig || 'HIG').toUpperCase() === 'HIB' ? 'HIB' : 'HIG',
          unit: k.unit || '%',
          source: k.source || sourceByCode.get(k.code) || 'Strategic Audit Plan',
          targetHorizon: k.targetHorizon || k.target_horizon || `Akhir Tahun ${currentYear}`,
          targetValue: Number(k.targetValue ?? k.target_value ?? 100),
          currentValue: Number(k.currentValue ?? k.current_value ?? 0),
          forecastedValue: Number(k.forecastedValue ?? k.forecasted_value ?? 0),
          achievementRate: Number(k.achievementRate ?? k.achievement_rate ?? 0),
          forecastedAchievementRate: Number(k.forecastedAchievementRate ?? k.forecasted_achievement_rate ?? 0),
          currentStatus: k.currentStatus || k.current_status || 'Moderate',
          predictedStatus: k.predictedStatus || k.predicted_status || 'Moderate',
          trend: k.trend || 'Improving',
          confidenceLower: Number(k.confidenceLower ?? k.confidence_lower ?? 0),
          confidenceUpper: Number(k.confidenceUpper ?? k.confidence_upper ?? 0),
          history: Array.isArray(k.history) ? k.history : [],
          forecast: Array.isArray(k.forecast) ? k.forecast : [],
          aiNarrative: k.aiNarrative || k.ai_narrative || '',
          recommendedAction: k.recommendedAction || k.recommended_action || ''
        }))

        const atRiskPillars = Array.isArray(aiData.at_risk_pillars)
          ? aiData.at_risk_pillars
          : kpiForecasts
            .filter((k: any) => k.predictedStatus === 'Poor' || k.predictedStatus === 'Moderate' || k.trend === 'Deteriorating')
            .map((k: any) => ({
              pillar: k.category,
              code: k.code,
              kpi: k.kpiName,
              target: k.targetValue,
              predicted: k.forecastedValue,
              achievementRate: k.forecastedAchievementRate,
              predictedStatus: k.predictedStatus
            }))

        const timeSeriesData = Array.isArray(aiData.time_series_data) && aiData.time_series_data.length > 0
          ? aiData.time_series_data
          : buildAuditorAggregateTimeSeries(kpiForecasts)

        timeseriesState.value.auditorForecast = {
          kpiForecasts,
          atRiskPillars,
          timeSeriesData,
          executiveSummary: aiData.executive_summary || ''
        }
        saveRealCache()
        return timeseriesState.value.auditorForecast
      }

      // Pure live-data fallback when AI engine is offline (never static CSV/mock)
      const localResult = buildAuditorLocalForecastsFromItems(payloadItems, currentYear)
      timeseriesState.value.auditorForecast = {
        kpiForecasts: localResult.kpiForecasts,
        atRiskPillars: localResult.atRiskPillars,
        timeSeriesData: localResult.timeSeriesData,
        executiveSummary: localResult.executiveSummary
      }
      saveRealCache()
      return timeseriesState.value.auditorForecast
    } catch (err) {
      console.warn('syncAuditorKpiWithStrategicPlan error:', err)
      return null
    }
  }

  const triggerManualForceRetrain = async () => {
    if (retrainStatus.value.isRunning) {
      return { ...retrainStatus.value, status: 'already_running' }
    }
    if (!hasUntrainedNewData.value && retrainStatus.value.statusFetchedOnce) {
      return { ...retrainStatus.value, status: 'up_to_date' }
    }

    retrainStatus.value = {
      ...retrainStatus.value,
      status: 'running',
      isRunning: true,
      trigger: 'user_manual_force_sync',
      message: 'Memulai Manual / Force Sync & melatih ulang model PyTorch LSTM dengan data terbaru...'
    }

    try {
      const customAuditee = getPersistedCustomAuditeeKpis()
      const res: any = await safeApiFetch('/retrain/auto?force=true&wait=true', {
        method: 'POST',
        body: {
          force: true,
          wait: true,
          scope: 'kpi',
          trigger: 'user_manual_force_sync',
          auditor_items: liveAuditorPayloadItems.value,
          auditee_items: customAuditee
        }
      })

      if (!res) {
        retrainStatus.value.isRunning = false
        retrainStatus.value.status = 'failed'
        retrainStatus.value.lastError = 'Layanan AI retrain tidak dapat dihubungi.'
        return { ...retrainStatus.value, status: 'failed' }
      }

      applyRetrainPayload(res)
      const d = res?.data || res
      if (d?.status !== 'already_running') {
        await fetchAiAnalytics(true, { skipAutoRetrain: true })
      }
      return { ...retrainStatus.value, status: d?.status || retrainStatus.value.status }
    } catch (err: any) {
      retrainStatus.value.isRunning = false
      retrainStatus.value.status = 'failed'
      retrainStatus.value.lastError = err?.message || 'Gagal melatih ulang model AI'
      return { ...retrainStatus.value, status: 'failed' }
    }
  }

  const addCustomAuditeeKpiForecasts = (newKpis: any[]) => {
    if (!Array.isArray(newKpis) || newKpis.length === 0) return []
    const sanitizedNewKpis = newKpis.map(sanitizeAuditeeKpiForecastRecord).filter(Boolean)
    if (sanitizedNewKpis.length === 0) return []
    savePersistedCustomAuditeeKpis(sanitizedNewKpis)
    customAuditeeVersion.value++
    const allCustom = getPersistedCustomAuditeeKpis()
    const currentBase = timeseriesState.value.auditeeForecast?.kpiForecasts || timeseriesState.value.kpiForecasts || []
    const mergedKpis = mergeAuditeeKpiForecasts(currentBase, allCustom)

    const customAtRisk = allCustom
      .filter((k: any) => k.riskLevel === 'HIGH' || k.riskLevel === 'MODERATE_HIGH' || k.trend === 'Deteriorating')
      .map((k: any) => sanitizeAuditeeAtRiskRecord({
        department: k.entity || 'Divisi Kredit & Portofolio',
        kpi: k.kpiName,
        predictedDrop: `${k.deltaPct >= 0 ? '+' : ''}${k.deltaPct}% (${k.forecastedValue}${k.unit === '%' ? '%' : ' ' + k.unit})`,
        riskLevel: k.riskLevel || 'HIGH'
      }))
    const existingAtRisk = (timeseriesState.value.auditeeForecast?.atRiskDepartments || timeseriesState.value.atRiskDepartments || []).map(sanitizeAuditeeAtRiskRecord)
    const mergedAtRisk = [...customAtRisk, ...existingAtRisk.filter((b: any) => !customAtRisk.some((c: any) => c.kpi === b.kpi))]

    timeseriesState.value.auditeeForecast = {
      ...(timeseriesState.value.auditeeForecast || {}),
      kpiForecasts: mergedKpis,
      atRiskDepartments: mergedAtRisk,
      executiveSummary: `Proyeksi KPI Auditee diperbarui dengan ${sanitizedNewKpis.length} indikator portofolio dari database eksternal (${sanitizedNewKpis[0]?.entity || 'Unit Operasional'}). Ditemukan ${mergedAtRisk.length} indikator/unit berisiko yang memerlukan pemantauan Continuous Auditing.`
    }
    timeseriesState.value.kpiForecasts = mergedKpis
    timeseriesState.value.atRiskDepartments = mergedAtRisk
    saveRealCache()
    return mergedKpis
  }

  const syncEntityPredictionWithRiskProfile = async (customItems?: any[], triggerRetrain: boolean = true) => {
    const currentYear = new Date().getFullYear()
    const isCustomExternalSync = Array.isArray(customItems) && customItems.length > 0
    let itemsToSync: any[] = []

    if (isCustomExternalSync) {
      itemsToSync = customItems!
        .filter((it: any) =>
          it.entity &&
          !String(it.entity).toLowerCase().includes('legal & compliance') &&
          it.risk_category !== 'Legal'
        )
        .map((it: any) => {
          const rawLvl = it.current_risk_level ?? it.risk_level ?? it.riskLevel ?? it.severity
          const lvlCode = rawLvl ? toEntityRiskCode(rawLvl) : null
          const [defLik, defImp] = lvlCode ? (CANONICAL_ENTITY_UI_PAIRS[lvlCode] || [3.0, 3.2]) : [3.0, 3.2]
          const lik = clampEntityRating(it.inherent_likelihood, defLik)
          const imp = clampEntityRating(it.inherent_impact, defImp)
          const resolvedCode = lvlCode || matrixCalc5Tier(lik, imp)
          return {
            ...it,
            entity_type: it.entity_type === 'Department' ? 'Auditable Entity' : (it.entity_type || 'Auditable Entity'),
            inherent_likelihood: lik,
            inherent_impact: imp,
            previous_risk_score: Number(it.previous_risk_score) > 0 ? Number(it.previous_risk_score) : +(lik * imp).toFixed(1),
            current_risk_level: resolvedCode
          }
        })
    } else {
      try {
        const auditUniverseStore = useAuditUniverseStore()

        if (!auditUniverseStore.yearlyUniverse || auditUniverseStore.yearlyUniverse.length === 0) {
          await auditUniverseStore.fetchYearlyUniverse(currentYear).catch(() => { })
        }
        if (!auditUniverseStore.corporateUniverse || auditUniverseStore.corporateUniverse.length === 0) {
          await auditUniverseStore.fetchCorporateUniverse().catch(() => { })
        }

        const entityMap = new Map<string, any>()

        // 1. Ingest Auditable Entities strictly from Audit Universe & Audit Priority
        const auditableEntities = buildAuditableEntityOptions(
          auditUniverseStore.yearlyUniverse,
          auditUniverseStore.corporateUniverse
        )

        for (const opt of auditableEntities) {
          if (!opt.name || String(opt.name).toLowerCase().includes('legal & compliance')) continue

          const cleanName = opt.name.trim()
          const lowerName = cleanName.toLowerCase()
          const isBranch = ['cabang', 'kcp', 'regional', 'kc', 'outlet', 'branch'].some(w => lowerName.includes(w))

          // Infer domain category from auditable entity name
          let category = 'Operational'
          if (
            lowerName.includes('keuangan') || lowerName.includes('kas') || lowerName.includes('pajak') ||
            lowerName.includes('akuntansi') || lowerName.includes('finance') || lowerName.includes('disbursement') ||
            lowerName.includes('treasury') || lowerName.includes('kredit') || lowerName.includes('lending') ||
            lowerName.includes('funding')
          ) {
            category = 'Financial'
          } else if (
            lowerName.includes('teknologi') || lowerName.includes('erp') || lowerName.includes('data') ||
            lowerName.includes('it') || lowerName.includes('ti') || lowerName.includes('akses') ||
            lowerName.includes('digital') || lowerName.includes('security') || lowerName.includes('cyber') ||
            lowerName.includes('server')
          ) {
            category = 'Technology'
          } else if (
            lowerName.includes('kepatuhan') || lowerName.includes('compliance') || lowerName.includes('regulasi') ||
            lowerName.includes('hukum') || lowerName.includes('sop')
          ) {
            category = 'Compliance'
          } else if (
            lowerName.includes('holding') || lowerName.includes('gcg') || lowerName.includes('governance') ||
            lowerName.includes('tata kelola')
          ) {
            category = 'Governance'
          } else if (lowerName.includes('sdm') || lowerName.includes('hr') || lowerName.includes('human')) {
            category = 'Human Resources'
          }

          // Map riskLevel and riskIndex strictly onto the 5 canonical Entity levels:
          // High, Moderate to High, Moderate, Low to Moderate, Low.
          const hasExplicitLevel = Boolean(opt.riskLevel && String(opt.riskLevel).trim())
          const rIndex = typeof opt.riskIndex === 'number' ? opt.riskIndex : Number(opt.riskIndex) || 0
          let canonicalLabel = hasExplicitLevel
            ? normalizeEntityRiskLevel(opt.riskLevel)
            : 'Moderate'

          if (rIndex >= 80) canonicalLabel = 'High'
          else if (rIndex >= 60 && (!hasExplicitLevel || canonicalLabel === 'Moderate')) canonicalLabel = 'Moderate to High'
          else if (!hasExplicitLevel) {
            if (rIndex >= 40) canonicalLabel = 'Moderate'
            else if (rIndex >= 20) canonicalLabel = 'Low to Moderate'
            else if (rIndex > 0) canonicalLabel = 'Low'
          }

          const lvlKey = toEntityRiskCode(canonicalLabel)
          const [lik, imp] = CANONICAL_ENTITY_UI_PAIRS[lvlKey] || [3.0, 3.2]
          const score = +(lik * imp).toFixed(1)
          entityMap.set(lowerName, {
            entity: cleanName,
            entity_type: isBranch ? 'Branch' : 'Auditable Entity',
            risk_category: category,
            target_timeline: `FY ${currentYear}`,
            inherent_likelihood: lik,
            inherent_impact: imp,
            audit_findings_count: opt.auditPriority ? 2 : 0,
            kpi_below_target_count: opt.auditPriority ? 1 : 0,
            kpi_volatility: 0.1,
            previous_risk_score: score,
            assessment_month: 1,
            current_risk_level: lvlKey,
            source_id: `audit_universe:${opt.id || cleanName}`
          })
        }

        // Only Audit Universe data is used: Corporate Risk Profile data is intentionally excluded
        // because branch entities in Audit Universe already represent the branches.
        itemsToSync = Array.from(entityMap.values())
      } catch (err) {
        console.warn('Could not read from auditUniverseStore:', err)
      }
    }

    if (itemsToSync.length === 0) return null

    const buildLocalPredictions = (list: any[]) => list.map((it: any) => {
      const lik = clampEntityRating(it.inherent_likelihood, 3)
      const imp = clampEntityRating(it.inherent_impact, 3)
      const actualScore = +(lik * imp).toFixed(1)
      const predLik = Math.min(5, Math.max(1, +(lik + ((it.audit_findings_count ?? 2) >= 2 ? 0.3 : -0.1)).toFixed(1)))
      const predImp = Math.min(5, Math.max(1, +(imp + ((it.kpi_below_target_count ?? 1) >= 1 ? 0.3 : -0.1)).toFixed(1)))
      const predScore = +(predLik * predImp).toFixed(1)
      const predLvl = toEntityRiskCode(matrixCalc5Tier(predLik, predImp))
      const rawActual = it.current_risk_level ?? it.currentRiskLevel ?? it.risk_level ?? it.riskLevel ?? it.severity
      const actualLvl = rawActual ? toEntityRiskCode(rawActual) : toEntityRiskCode(matrixCalc5Tier(lik, imp))
      const isBr = ['cabang', 'kcp', 'regional', 'kc', 'outlet', 'branch'].some(w => String(it.entity).toLowerCase().includes(w))
      const defTimeline = it.target_timeline && !String(it.target_timeline).startsWith('Q') ? it.target_timeline : `FY ${currentYear}`
      return {
        entity: it.entity,
        type: it.entity_type === 'Department' ? (isBr ? 'Branch' : 'Auditable Entity') : (it.entity_type || (isBr ? 'Branch' : 'Auditable Entity')),
        riskCategory: it.risk_category || 'Operational',
        targetTimeline: defTimeline,
        currentLikelihood: lik,
        currentImpact: imp,
        predictedLikelihood: predLik,
        predictedImpact: predImp,
        predictedScore: predScore,
        actualScore: actualScore,
        delta: +(predScore - actualScore).toFixed(1),
        trend: predScore > actualScore ? 'up' : (predScore < actualScore ? 'down' : 'stable'),
        predictedRiskLevel: predLvl,
        actualRiskLevel: actualLvl,
        source: it.source_id || 'external_db'
      }
    })

    // If customItems were passed from Data Source Mapping ("Prediksi Entitas AI"), persist & merge them directly into Entity Prediction!
    if (isCustomExternalSync) {
      const customPreds = buildLocalPredictions(itemsToSync)
      savePersistedCustomEntityPredictions(customPreds)
      xgboostState.value.predictions = mergeEntityPredictions(
        xgboostState.value.predictions || [],
        getPersistedCustomEntityPredictions()
      )
      saveRealCache()
    }

    try {
      const payload = {
        items: itemsToSync,
        trigger_retrain: triggerRetrain
      }

      let res: any = await safeApiFetch('/predict/risk-score/sync-profile', {
        method: 'POST',
        body: payload
      }, 'python').catch(() => null)

      if (!res) {
        res = await safeApiFetch('/api/analytics/risk-score/sync-profile', {
          method: 'POST',
          body: payload
        }, 'go').catch(() => null)
      }

      const resData = res?.data || res
      if (resData?.predictions && resData.predictions.length > 0) {
        const mappedApiPreds = resData.predictions
          .filter((p: any) => p.entity && !String(p.entity).toLowerCase().includes('legal & compliance') && p.risk_category !== 'Legal')
          .map((p: any) => {
            const isBr = ['cabang', 'kcp', 'regional', 'kc', 'outlet', 'branch'].some(w => String(p.entity).toLowerCase().includes(w))
            let typ = p.type || (isBr ? 'Branch' : 'Auditable Entity')
            if (typ === 'Department') typ = isBr ? 'Branch' : 'Auditable Entity'
            const defTimeline = `FY ${currentYear}`
            return {
              entity: p.entity || 'Unknown Entity',
              type: typ,
              riskCategory: p.risk_category || p.riskCategory || 'Operational',
              targetTimeline: p.target_timeline && !p.target_timeline.startsWith('Q') ? p.target_timeline : (p.targetTimeline && !p.targetTimeline.startsWith('Q') ? p.targetTimeline : defTimeline),
              currentLikelihood: p.current_likelihood ?? p.actual_likelihood ?? p.inherent_likelihood ?? p.currentLikelihood ?? 3,
              currentImpact: p.current_impact ?? p.actual_impact ?? p.inherent_impact ?? p.currentImpact ?? 3,
              predictedLikelihood: p.predicted_likelihood ?? p.predictedLikelihood ?? 3,
              predictedImpact: p.predicted_impact ?? p.predictedImpact ?? 3,
              predictedScore: p.predicted_score ?? p.predictedScore ?? 9,
              actualScore: p.actual_score ?? p.actualScore ?? 9,
              delta: p.delta ?? 0,
              trend: p.trend || 'stable',
              predictedRiskLevel: toEntityRiskCode(p.risk_level || p.predictedRiskLevel || 'MODERATE'),
              actualRiskLevel: toEntityRiskCode(p.actual_risk_level || p.actualRiskLevel || 'MODERATE')
            }
          })

        if (isCustomExternalSync) {
          savePersistedCustomEntityPredictions(mappedApiPreds)
          xgboostState.value.predictions = mergeEntityPredictions(
            xgboostState.value.predictions || [],
            getPersistedCustomEntityPredictions()
          )
        } else {
          xgboostState.value.predictions = mergeEntityPredictions(
            mappedApiPreds,
            getPersistedCustomEntityPredictions()
          )
        }
        if (resData.feature_importance) {
          xgboostState.value.featureImportance = resData.feature_importance
        }
        if (resData.metrics) {
          xgboostState.value.modelMetrics = resData.metrics
        }
        saveRealCache()
        await fetchRetrainStatus().catch(() => null)
        return { ...resData, predictions: xgboostState.value.predictions }
      }

      if (itemsToSync.length > 0) {
        const computedPreds = buildLocalPredictions(itemsToSync)
        if (isCustomExternalSync) {
          savePersistedCustomEntityPredictions(computedPreds)
          xgboostState.value.predictions = mergeEntityPredictions(
            xgboostState.value.predictions || [],
            getPersistedCustomEntityPredictions()
          )
        } else {
          xgboostState.value.predictions = mergeEntityPredictions(
            computedPreds,
            getPersistedCustomEntityPredictions()
          )
        }
        saveRealCache()
        return { predictions: xgboostState.value.predictions }
      }
    } catch (e) {
      console.warn('syncEntityPredictionWithRiskProfile failed:', e)
    }
    return { predictions: xgboostState.value.predictions }
  }

  return {
    loading,
    error,
    isAiConnected,
    usingCachedRealData,
    lastSyncedTime,
    summary,
    xgboostState,
    syncEntityPredictionWithRiskProfile,
    syncEntityPredictionWithAuditUniverse: syncEntityPredictionWithRiskProfile,
    syncAuditorKpiWithStrategicPlan,
    addCustomAuditeeKpiForecasts,
    sanitizeAuditeeKpiForecastRecord,
    isolationState,
    nlpState,
    timeseriesState,
    retrainStatus,
    modelStatuses,
    retrainQueue,
    retrainResource,
    checkingRetrainStatus,
    hasUntrainedAuditorData,
    hasUntrainedAuditeeData,
    hasUntrainedNewData,
    untrainedDataReasons,
    canManualRetrain,
    fetchRetrainStatus,
    retrainModel,
    triggerManualForceRetrain,
    automatedAnomalyRules,
    trainAnomalyWithRules,
    getPersistedAutomatedRules,
    savePersistedAutomatedRules,
    formatNum,
    getRiskConfig,
    ENTITY_RISK_LEVELS,
    NLP_RISK_LEVELS,
    normalizeAnomalyRiskLevel,
    normalizeEntityRiskLevel,
    normalizeNlpRiskLevel,
    toEntityRiskCode,
    toNlpRiskCode,
    sanitizeNlpDocumentRecord,
    sentimentColor,
    riskCategoryColor,
    controlIndicatorColor,
    trendIcon,
    trendColor,
    safeApiFetch,
    fetchAiAnalytics,
    analyzeDocumentOrText,
    addNlpTaxonomyItem,
    syncNlpMultiSources,
    deleteNlpDocument,
    clearAllCustomNlpDocuments,
    deletePersistedCustomNlpDoc,
    clearAllPersistedCustomNlpDocs,
    addCustomAnomalyRecord,
    removeCustomAnomalyRecord,
    addCustomScatterBatch,
    savePersistedCustomScatterBatch
  }
}
