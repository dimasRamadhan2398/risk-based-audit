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
import { riskLevelConfig } from '~/stores/risk-profile'

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

const xgboostState = ref<any>({
  predictions: [...initialXGB.predictions],
  featureImportance: [...initialXGB.featureImportance],
  modelMetrics: { ...initialXGB.modelMetrics }
})

const isolationState = ref<any>({
  anomalies: [...initialIso.anomalies],
  scatterData: [...initialIso.scatterData],
  summary: { ...initialIso.summary }
})

const nlpState = ref<any>({
  documents: [...initialNLP.documents],
  categoryDistribution: { ...initialNLP.categoryDistribution },
  sentimentDistribution: { ...initialNLP.sentimentDistribution }
})

const timeseriesState = ref<any>({
  historicalKPI: [...initialTS.historicalKPI],
  kpiForecasts: [...initialTS.kpiForecasts],
  atRiskDepartments: [...initialTS.atRiskDepartments],
  forecastAccuracy: { ...initialTS.forecastAccuracy }
})

const CACHE_KEY = 'auditsphere_real_analytics_cache_v6'

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
    localStorage.removeItem('auditsphere_real_analytics_cache_v5')
    localStorage.removeItem('auditsphere_real_analytics_cache_v4')
    localStorage.removeItem('auditsphere_real_analytics_cache_v3')
    localStorage.removeItem('auditsphere_real_analytics_cache')
    lastSyncedTime.value = timeStr
  } catch (e) {
    console.warn('Failed to save real analytics cache', e)
  }
}

const loadRealCache = (): boolean => {
  if (typeof window === 'undefined') return false
  try {
    localStorage.removeItem('auditsphere_real_analytics_cache_v5')
    localStorage.removeItem('auditsphere_real_analytics_cache_v4')
    localStorage.removeItem('auditsphere_real_analytics_cache_v3')
    localStorage.removeItem('auditsphere_real_analytics_cache')

    const raw = localStorage.getItem(CACHE_KEY)
    if (!raw) return false
    const parsed = JSON.parse(raw)

    const bankCats = ['Funding', 'Lending', 'Treasury', 'Payment', 'KYC', 'IT Control']
    const hasBank = parsed.isolation?.anomalies?.some((a: any) => bankCats.includes(a.type))
    if (!hasBank) {
      localStorage.removeItem(CACHE_KEY)
      return false
    }

    if (parsed.xgboost) xgboostState.value = parsed.xgboost
    if (parsed.isolation) isolationState.value = parsed.isolation
    if (parsed.nlp) nlpState.value = parsed.nlp
    if (parsed.timeseries) timeseriesState.value = parsed.timeseries
    if (parsed.time) lastSyncedTime.value = parsed.time
    return true
  } catch (e) {
    console.warn('Failed to load real analytics cache', e)
    return false
  }
}

export const useAiAnalytics = () => {
  const config = useRuntimeConfig()

  const getAnalyticsUrl = () => config.public.analyticsApiBase || '/api/analytics'
  const getPythonAiUrl = () => config.public.pythonAiBaseUrl || '/api/python-ai'

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
      else if (endpoint === '/performance-trend/batch') pyEndpoint = '/predict/performance-trend/batch'
      else if (endpoint === '/retrain/auto') pyEndpoint = '/retrain/auto'

      try {
        return await $fetch(`${pyUrl}${pyEndpoint}`, options)
      } catch (e2) {
        return null
      }
    }
  }

  const formatNum = (val: any, decimals = 1): string => {
    if (val === null || val === undefined || isNaN(Number(val))) return '0'
    return Number(val).toFixed(decimals)
  }

  const getRiskConfig = (level: string) => {
    if (!level) return { label: 'Moderate', color: '#FFC107', bg: '#FF6F00' }
    const raw = String(level).trim().toLowerCase().replace(/[\s\-_]+/g, '')

    if (raw.includes('critical') || raw.includes('extreme') || raw === 'high' || raw.includes('veryhigh')) {
      return { label: level.toUpperCase() === 'HIGH' ? 'High' : level, color: '#F44336', bg: '#B71C1C' }
    }
    if (raw.includes('moderatehigh') || raw.includes('mediumhigh')) {
      return { label: 'Moderate to High', color: '#FF9800', bg: '#E65100' }
    }
    if (raw.includes('moderate') || raw.includes('medium') || raw.includes('watch')) {
      return { label: 'Moderate', color: '#FFC107', bg: '#FF6F00' }
    }
    if (raw.includes('lowmoderate')) {
      return { label: 'Low to Moderate', color: '#8BC34A', bg: '#33691E' }
    }
    if (raw.includes('low') || raw.includes('verylow')) {
      return { label: 'Low', color: '#4CAF50', bg: '#1B5E20' }
    }

    const mapped = (riskLevelConfig as any)[raw] || (riskLevelConfig as any)[level]
    if (mapped) return mapped

    return { label: level, color: '#FFC107', bg: '#FF6F00' }
  }

  type BadgeColor = 'error' | 'primary' | 'warning' | 'success' | 'info' | 'neutral'

  const sentimentColor = (s: string): BadgeColor => {
    const map: Record<string, BadgeColor> = { Positive: 'success', Neutral: 'warning', Negative: 'error' }
    return map[s] || 'neutral'
  }

  const riskCategoryColor = (cat: string): BadgeColor => {
    const map: Record<string, BadgeColor> = {
      'Financial': 'error', 'Technology': 'info', 'Operational': 'warning',
      'Compliance': 'success', 'Strategic': 'primary', 'Reputational': 'neutral',
      'Legal': 'warning', 'Security': 'error', 'Fraud & Security': 'error'
    }
    return map[cat] || 'neutral'
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

  const fetchAiAnalytics = async (force = false) => {
    if (hasFetchedOnce.value && !force) return

    try {
      loading.value = true

      // Silent background auto-retrain trigger
      safeApiFetch('/retrain/auto', { method: 'POST' }).catch(() => {})

      const [resRiskBatch, resAnomalyBatch, resTextBatch, resPerfBatch]: any[] = await Promise.all([
        safeApiFetch('/risk-score/batch'),
        safeApiFetch('/anomaly/batch'),
        safeApiFetch('/text-analysis/batch'),
        safeApiFetch('/performance-trend/batch')
      ])

      const hasRealData = !!(resRiskBatch || resAnomalyBatch || resTextBatch || resPerfBatch)

      if (hasRealData) {
        isAiConnected.value = true
        usingCachedRealData.value = false

        if (resRiskBatch?.data || resRiskBatch?.predictions) {
          const bData = resRiskBatch.data || resRiskBatch
          if (bData.predictions && bData.predictions.length > 0) {
            xgboostState.value.predictions = bData.predictions.map((p: any) => ({
              entity: p.entity || 'Unknown Entity',
              type: p.type || (p.entity && p.entity.includes('Dept') ? 'Department' : 'Branch'),
              riskCategory: p.risk_category || p.riskCategory || 'Financial',
              targetTimeline: p.target_timeline || p.targetTimeline || 'Q3 2026',
              predictedLikelihood: p.predicted_likelihood ?? p.predictedLikelihood ?? 3,
              predictedImpact: p.predicted_impact ?? p.predictedImpact ?? 3,
              predictedScore: p.predicted_score ?? p.predictedScore ?? 9,
              actualScore: p.actual_score ?? p.actualScore ?? 9,
              delta: p.delta ?? 0,
              trend: p.trend || 'stable',
              predictedRiskLevel: p.risk_level || p.predictedRiskLevel || 'MODERATE_HIGH',
              actualRiskLevel: p.actual_risk_level || p.actualRiskLevel || 'MODERATE_HIGH'
            }))
          }
        }

        if (resAnomalyBatch?.data || resAnomalyBatch?.anomalies) {
          const bData = resAnomalyBatch.data || resAnomalyBatch
          if (bData.anomalies && bData.anomalies.length > 0) {
            isolationState.value.anomalies = bData.anomalies.map((a: any) => ({
              id: a.id || 'ANM-001',
              entity: a.entity || 'Jakarta Branch',
              type: a.type || 'Funding',
              anomalyScore: a.anomaly_score ?? a.anomalyScore ?? 0.85,
              description: a.description || '',
              severity: a.severity || 'High',
              date: a.date || '2026-06-01',
              amount: a.amount ?? 0,
              xMetric: a.xMetric ?? (a.amount ? a.amount / 1000000 : 0),
              isAnomaly: a.is_anomaly !== undefined ? a.is_anomaly : true,
              riskLevel: a.risk_level || a.riskLevel || 'HIGH'
            }))
          }
          if (bData.scatter_data && bData.scatter_data.length > 0) {
            isolationState.value.scatterData = bData.scatter_data.map((s: any) => ({
              x: s.x ?? 0,
              y: s.y ?? 0,
              type: s.type,
              isAnomaly: s.is_anomaly ?? false,
              label: s.label || ''
            }))
          }
          if (bData.summary) {
            isolationState.value.summary = bData.summary
            summary.value.anomaliesDetected = bData.summary.anomaliesFound || isolationState.value.anomalies.length
          }
        }

        if (resTextBatch?.data || resTextBatch?.documents) {
          const bData = resTextBatch.data || resTextBatch
          if (bData.documents && bData.documents.length > 0) {
            nlpState.value.documents = bData.documents.map((d: any) => ({
              docId: d.docId || 'WP-001',
              title: d.title || 'Document Title',
              source: d.source || 'Working Paper',
              autoCategory: d.risk_category || d.autoCategory || 'Financial',
              confidence: d.confidence ?? 0.91,
              sentiment: d.sentiment || 'Negative',
              severityScore: d.severityScore ?? 80,
              date: d.date || '2026-06-01',
              excerpt: d.excerpt || '',
              riskLevel: d.risk_level || d.riskLevel || 'HIGH'
            }))
          }
          if (bData.category_distribution) nlpState.value.categoryDistribution = bData.category_distribution
          if (bData.sentiment_distribution) nlpState.value.sentimentDistribution = bData.sentiment_distribution
        }

        if (resPerfBatch?.data || resPerfBatch?.kpi_forecasts) {
          const bData = resPerfBatch.data || resPerfBatch
          if (bData.kpi_forecasts && bData.kpi_forecasts.length > 0) {
            timeseriesState.value.kpiForecasts = bData.kpi_forecasts.map((f: any) => ({
              kpiName: f.kpiName || f.kpi_name,
              code: f.code || 'KPI-001',
              unit: f.unit || '%',
              entity: f.entity || 'Finance Dept',
              entityType: f.entityType || f.entity_type || 'Department',
              targetHorizon: f.targetHorizon || f.target_horizon || 'Q3 2026',
              currentValue: f.currentValue ?? f.current_value ?? 0,
              forecastedValue: f.forecastedValue ?? f.forecasted_value ?? 0,
              trend: f.trend || 'Stable',
              recommendedAction: f.recommendedAction || f.recommended_action || '',
              riskLevel: f.riskLevel || f.risk_level || 'LOW'
            }))
          }
          if (bData.at_risk_departments) timeseriesState.value.atRiskDepartments = bData.at_risk_departments
          if (bData.time_series_data) timeseriesState.value.historicalKPI = bData.time_series_data
        }

        saveRealCache()
      } else {
        isAiConnected.value = false
        const loaded = loadRealCache()
        usingCachedRealData.value = loaded
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

  return {
    loading,
    error,
    isAiConnected,
    usingCachedRealData,
    lastSyncedTime,
    summary,
    xgboostState,
    isolationState,
    nlpState,
    timeseriesState,
    formatNum,
    getRiskConfig,
    sentimentColor,
    riskCategoryColor,
    trendIcon,
    trendColor,
    fetchAiAnalytics
  }
}
