<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Line, Bar, Doughnut, Scatter } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
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
import { RiskLevel } from '~/types/risk'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement, ArcElement, Title, Tooltip, Legend, Filler)

// ─── Data & Config ──────────────────────────────────────────────────────────
const getAnalyticsUrl = () => {
  const config = useRuntimeConfig()
  if (import.meta.client && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
    return 'https://api.auditsphere.app/api/analytics'
  }
  return config.public.analyticsApiBase || 'http://localhost:8084/api/analytics'
}

const getPythonAiUrl = () => {
  const config = useRuntimeConfig()
  if (import.meta.client && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
    return 'https://api.auditsphere.app/api/python-ai'
  }
  return config.public.pythonAiBaseUrl || 'http://localhost:8000'
}

const initialXGB = useXGBoostData()
const initialIso = useIsolationForestData()
const initialNLP = useIndoBERTData()
const initialTS = useTimeSeriesData()
const summary = ref(useAnalyticsSummary())

const loading = ref(true)
const error = ref('')

// Connection & Persistent Cache State
const isAiConnected = ref(true)
const usingCachedRealData = ref(false)
const lastSyncedTime = ref<string | null>(null)

// Reactive Model States driven by Real Model Data & Continuous Background Sync
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

// Safe Fetch Helper with Multi-tier Fallbacks
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
      console.warn(`API call failed for ${endpoint}, using fallback`, e2)
      return null
    }
  }
}

// Helper formatting functions to prevent template TypeError exceptions
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

// ─── LocalStorage Persistent Cache Helpers ──────────────────────────────
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
    localStorage.setItem('auditsphere_real_analytics_cache_v3', JSON.stringify(cacheObj))
    lastSyncedTime.value = timeStr
  } catch (e) {
    console.warn('Failed to save real analytics cache', e)
  }
}

const loadRealCache = (): boolean => {
  if (typeof window === 'undefined') return false
  try {
    const raw = localStorage.getItem('auditsphere_real_analytics_cache_v3') || localStorage.getItem('auditsphere_real_analytics_cache')
    if (!raw) return false
    const parsed = JSON.parse(raw)
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

// ─── Fetch Initial Batch Predictions & Automated Sync ───────────────────────
const fetchAnalytics = async () => {
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
            type: a.type || 'Transaction',
            anomalyScore: a.anomaly_score ?? a.anomalyScore ?? 0.85,
            description: a.description || '',
            severity: a.severity || 'High',
            date: a.date || '2026-06-01',
            amount: a.amount ?? 0,
            isAnomaly: a.is_anomaly !== undefined ? a.is_anomaly : true,
            riskLevel: a.risk_level || a.riskLevel || 'HIGH'
          }))
        }
        if (bData.scatter_data && bData.scatter_data.length > 0) {
          isolationState.value.scatterData = bData.scatter_data.map((s: any) => ({
            x: s.x ?? 0,
            y: s.y ?? 0,
            isAnomaly: s.is_anomaly ?? false,
            label: s.label || ''
          }))
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

      // Save real data to persistent storage
      saveRealCache()
    } else {
      isAiConnected.value = false
      const loaded = loadRealCache()
      usingCachedRealData.value = loaded
    }

  } catch (err: any) {
    error.value = err.message || 'Error loading analytics'
    isAiConnected.value = false
    const loaded = loadRealCache()
    usingCachedRealData.value = loaded
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchAnalytics()
})

// ─── Tab Navigation ─────────────────────────────────────────────────────────
const activeTab = ref('xgboost')
const tabItems = [
  { key: 'xgboost', label: 'Risk Scoring ML', icon: 'i-heroicons-chart-bar-square' },
  { key: 'isolation', label: 'Anomaly Detection ML', icon: 'i-heroicons-shield-exclamation' },
  { key: 'nlp', label: 'IndoBERT NLP', icon: 'i-heroicons-document-magnifying-glass' },
  { key: 'timeseries', label: 'KPI PyTorch LSTM', icon: 'i-heroicons-arrow-trending-up' },
]

// ─── Tab 1: Dynamic Computed Charts for Risk Scoring ───────────────────────
const xgboostBarData = computed(() => ({
  labels: xgboostState.value.predictions.map((p: any) => p.entity || 'Entity'),
  datasets: [
    {
      label: 'Predicted Score',
      backgroundColor: 'rgba(99,102,241,0.75)',
      borderColor: 'rgb(99,102,241)',
      borderWidth: 1.5,
      borderRadius: 4,
      data: xgboostState.value.predictions.map((p: any) => p.predictedScore ?? 0),
    },
    {
      label: 'Actual Score',
      backgroundColor: 'rgba(16,185,129,0.75)',
      borderColor: 'rgb(16,185,129)',
      borderWidth: 1.5,
      borderRadius: 4,
      data: xgboostState.value.predictions.map((p: any) => p.actualScore ?? 0),
    },
  ],
}))

const xgboostBarOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'top' as const },
    tooltip: {
      callbacks: {
        label: (ctx: any) => `${ctx.dataset.label}: ${formatNum(ctx.parsed.y, 1)}`
      }
    }
  },
  scales: {
    y: { beginAtZero: true, title: { display: true, text: 'Risk Score' } },
    x: { ticks: { maxRotation: 45 } }
  }
}

const featureBarData = computed(() => ({
  labels: xgboostState.value.featureImportance.map((f: any) => f.feature || 'Feature'),
  datasets: [{
    label: 'Feature Importance',
    backgroundColor: [
      'rgba(239,68,68,0.75)', 'rgba(249,115,22,0.75)', 'rgba(234,179,8,0.75)',
      'rgba(34,197,94,0.75)', 'rgba(59,130,246,0.75)', 'rgba(139,92,246,0.75)',
      'rgba(236,72,153,0.75)', 'rgba(107,114,128,0.75)'
    ],
    borderWidth: 0,
    borderRadius: 4,
    data: xgboostState.value.featureImportance.map((f: any) => +((f.importance ?? 0) * 100).toFixed(1)),
  }]
}))

const featureBarOptions = {
  indexAxis: 'y' as const,
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: { callbacks: { label: (ctx: any) => `${formatNum(ctx.parsed.x, 1)}%` } }
  },
  scales: {
    x: { beginAtZero: true, max: 35, title: { display: true, text: 'Importance (%)' } }
  }
}

// ─── Tab 2: Dynamic Computed Scatter Chart for Anomaly Detection (Separated) ─
interface AnomalyTypeConfig {
  xAxisTitle: string
  unit: string
  formatX: (val: number) => string
  colors: { bg: string, border: string, style: string }
}

const anomalyTypeConfigs: Record<string, AnomalyTypeConfig> = {
  'Fieldwork': {
    xAxisTitle: 'Durasi Pengerjaan Fieldwork (Hari)',
    unit: 'Hari',
    formatX: (val) => `${val} Hari`,
    colors: { bg: 'rgba(139,92,246,0.85)', border: 'rgba(139,92,246,1)', style: 'star' }
  },
  'Access Pattern': {
    xAxisTitle: 'Waktu Akses Sistem (Jam 0-23)',
    unit: 'Jam',
    formatX: (val) => `Jam ${val}:00`,
    colors: { bg: 'rgba(249,115,22,0.85)', border: 'rgba(249,115,22,1)', style: 'rectRot' }
  },
  'Data Access': {
    xAxisTitle: 'Volume Export Data (MB)',
    unit: 'MB',
    formatX: (val) => `${val} MB`,
    colors: { bg: 'rgba(59,130,246,0.85)', border: 'rgba(59,130,246,1)', style: 'rect' }
  },
  'Inventory': {
    xAxisTitle: 'Nilai Penyesuaian Stok (Rp Juta)',
    unit: 'Rp Juta',
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(107,114,128,0.85)', border: 'rgba(107,114,128,1)', style: 'star' }
  },
  'Expense Report': {
    xAxisTitle: 'Nominal Klaim Pengeluaran (Rp Juta)',
    unit: 'Rp Juta',
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(234,179,8,0.85)', border: 'rgba(234,179,8,1)', style: 'rect' }
  },
  'Travel Expense': {
    xAxisTitle: 'Nominal Biaya Perjalanan (Rp Juta)',
    unit: 'Rp Juta',
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(16,185,129,0.85)', border: 'rgba(16,185,129,1)', style: 'rectRot' }
  },
  'Procurement': {
    xAxisTitle: 'Nilai Pengadaan / Vendor (Rp Juta)',
    unit: 'Rp Juta',
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(236,72,153,0.85)', border: 'rgba(236,72,153,1)', style: 'triangle' }
  },
  'Transaction': {
    xAxisTitle: 'Nominal Transaksi (Rp Juta)',
    unit: 'Rp Juta',
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(239,68,68,0.85)', border: 'rgba(239,68,68,1)', style: 'triangle' }
  }
}

const selectedAnomalyType = ref('Expense Report')

const availableAnomalyTypes = computed(() => {
  const types = new Set<string>()
  const anomalies = isolationState.value.anomalies || []
  const scatterData = isolationState.value.scatterData || []

  anomalies.forEach((a: any) => { if (a.type) types.add(a.type) })
  scatterData.forEach((s: any) => { if (s.type) types.add(s.type) })

  if (types.size === 0) {
    ['Expense Report', 'Travel Expense', 'Procurement', 'Access Pattern', 'Fieldwork', 'Data Access', 'Inventory', 'Transaction'].forEach(t => types.add(t))
  }

  return Array.from(types).filter(t => t !== 'All')
})

watchEffect(() => {
  if ((selectedAnomalyType.value === 'All' || !selectedAnomalyType.value || !availableAnomalyTypes.value.includes(selectedAnomalyType.value)) && availableAnomalyTypes.value.length > 0) {
    selectedAnomalyType.value = availableAnomalyTypes.value[0]
  }
})

const getScatterChartForType = (typeFilter: string) => {
  const anomalies = isolationState.value.anomalies || []
  const scatterData = isolationState.value.scatterData || []

  const filteredAnomalies = anomalies.filter((a: any) => a.type === typeFilter)
  const normalPoints = scatterData
    .filter((s: any) => !s.isAnomaly && (s.type === typeFilter || !s.type))
    .map((s: any) => ({ x: s.x ?? 0, y: s.y ?? 0 }))
  
  const anomalyPoints = filteredAnomalies.map((a: any) => {
    const matchedScatter = scatterData.find((s: any) => s.label === a.id)
    let xVal = matchedScatter?.x ?? a.xMetric
    if (xVal === undefined || xVal === null) {
      if (typeFilter === 'Fieldwork') xVal = 24
      else if (typeFilter === 'Access Pattern') xVal = 2
      else if (typeFilter === 'Data Access') xVal = 450
      else xVal = a.amount ? a.amount / 1000000 : 15
    }
    return {
      x: xVal,
      y: matchedScatter?.y ?? 20
    }
  })

  const config = anomalyTypeConfigs[typeFilter] || anomalyTypeConfigs['Transaction']
  const colors = config.colors

  return {
    datasets: [
      {
        label: `Normal Data (${typeFilter})`,
        data: normalPoints,
        backgroundColor: 'rgba(59,130,246,0.3)',
        borderColor: 'rgba(59,130,246,0.5)',
        pointRadius: 4,
      },
      {
        label: `Anomalies (${typeFilter})`,
        data: anomalyPoints,
        backgroundColor: colors.bg,
        borderColor: colors.border,
        pointRadius: 8,
        pointStyle: colors.style,
      }
    ]
  }
}

const scatterChartData = computed(() => getScatterChartForType(selectedAnomalyType.value))

const scatterOptions = computed(() => {
  const currentConfig = anomalyTypeConfigs[selectedAnomalyType.value] || {
    xAxisTitle: 'Metrik Anomali (Nilai / Score)',
    unit: 'Value',
    formatX: (val: number) => `${val}`
  }

  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { position: 'top' as const },
      tooltip: {
        callbacks: {
          label: (ctx: any) => {
            const formattedX = currentConfig.formatX ? currentConfig.formatX(ctx.parsed.x) : `${ctx.parsed.x}`
            const metricName = currentConfig.xAxisTitle.split(' (')[0]
            return `${metricName}: ${formattedX}, Frequency/Occurrence: ${ctx.parsed.y}`
          }
        }
      }
    },
    scales: {
      x: { title: { display: true, text: currentConfig.xAxisTitle } },
      y: { title: { display: true, text: 'Frequency / Occurrence' } }
    }
  }
})

// ─── Tab 3: Dynamic Computed Doughnut Charts for IndoBERT NLP ─────────────
const categoryDoughnutData = computed(() => ({
  labels: Object.keys(nlpState.value.categoryDistribution || {}),
  datasets: [{
    data: Object.values(nlpState.value.categoryDistribution || {}) as number[],
    backgroundColor: [
      '#EF4444', '#8B5CF6', '#F59E0B', '#3B82F6',
      '#10B981', '#6366F1', '#EC4899',
    ],
    borderWidth: 2,
    borderColor: 'rgba(255,255,255,0.1)',
  }]
}))

const getSentimentCount = (key: string): number => {
  const dist = nlpState.value?.sentimentDistribution || {}
  const lower = key.toLowerCase()
  const upper = key.charAt(0).toUpperCase() + key.slice(1).toLowerCase()
  return (dist[lower] ?? dist[upper] ?? dist[key] ?? 0) as number
}

const sentimentDoughnutData = computed(() => ({
  labels: ['Positive', 'Neutral', 'Negative'],
  datasets: [{
    data: [
      getSentimentCount('positive'),
      getSentimentCount('neutral'),
      getSentimentCount('negative')
    ],
    backgroundColor: ['#10B981', '#F59E0B', '#EF4444'],
    borderWidth: 2,
    borderColor: 'rgba(255,255,255,0.1)',
  }]
}))

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'bottom' as const }
  }
}

// ─── Tab 4: Dynamic Computed Line Chart for PyTorch LSTM ───────────────────
const timeSeriesChartData = computed(() => ({
  labels: (timeseriesState.value.historicalKPI || []).map((p: any) => p.period || ''),
  datasets: [
    {
      label: 'Actual KPI (%)',
      data: (timeseriesState.value.historicalKPI || []).map((p: any) => p.actual ?? null),
      borderColor: 'rgb(59,130,246)',
      backgroundColor: 'rgba(59,130,246,0.1)',
      tension: 0.4,
      pointRadius: 5,
      pointBackgroundColor: 'rgb(59,130,246)',
      spanGaps: false,
    },
    {
      label: 'PyTorch LSTM Forecast',
      data: (timeseriesState.value.historicalKPI || []).map((p: any) => p.forecast ?? null),
      borderColor: 'rgb(139,92,246)',
      backgroundColor: 'rgba(139,92,246,0.1)',
      borderDash: [6, 4],
      tension: 0.4,
      pointRadius: 5,
      pointStyle: 'rectRot',
      pointBackgroundColor: 'rgb(139,92,246)',
      spanGaps: false,
    },
    {
      label: 'Upper Bound',
      data: (timeseriesState.value.historicalKPI || []).map((p: any) => p.upperBound ?? null),
      borderColor: 'transparent',
      backgroundColor: 'rgba(139,92,246,0.08)',
      fill: '+1',
      tension: 0.4,
      pointRadius: 0,
      spanGaps: false,
    },
    {
      label: 'Lower Bound',
      data: (timeseriesState.value.historicalKPI || []).map((p: any) => p.lowerBound ?? null),
      borderColor: 'transparent',
      backgroundColor: 'rgba(139,92,246,0.08)',
      fill: '-1',
      tension: 0.4,
      pointRadius: 0,
      spanGaps: false,
    },
  ]
}))

const timeSeriesOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        filter: (item: any) => !['Upper Bound', 'Lower Bound'].includes(item.text)
      }
    },
    tooltip: {
      callbacks: {
        label: (ctx: any) => {
          if (['Upper Bound', 'Lower Bound'].includes(ctx.dataset.label)) return ''
          return `${ctx.dataset.label}: ${formatNum(ctx.parsed.y, 1)}%`
        }
      }
    }
  },
  scales: {
    y: { title: { display: true, text: 'KPI Score (%)' } }
  }
}

// ─── Helpers ────────────────────────────────────────────────────────────────
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
  const map: Record<string, string> = { Improving: 'i-heroicons-arrow-trending-up', Declining: 'i-heroicons-arrow-trending-down', Deteriorating: 'i-heroicons-arrow-trending-down', Stable: 'i-heroicons-minus', up: 'i-heroicons-arrow-trending-up', down: 'i-heroicons-arrow-trending-down', stable: 'i-heroicons-minus' }
  return map[t] || 'i-heroicons-minus'
}

const trendColor = (t: string) => {
  if (['Improving', 'down'].includes(t)) return 'text-emerald-500'
  if (['Declining', 'up', 'Deteriorating'].includes(t)) return 'text-rose-500'
  return 'text-gray-400'
}

const pct = (v: number) => `${((v || 0) * 100).toFixed(1)}%`
</script>

<template>
  <div class="max-w-[1440px] mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <!-- Header -->
    <div class="mb-8 flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-3xl font-black tracking-tight text-gray-900 dark:text-white flex items-center gap-3">
          <UIcon name="i-heroicons-cpu-chip" class="text-indigo-600 dark:text-indigo-400 w-8 h-8" />
          AI & Analytics Center
        </h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">
          Integrasi Terpadu Model Machine Learning & Deep Learning (Department Risk ML, Anomaly Detection, IndoBERT NLP, PyTorch LSTM)
        </p>
      </div>
      <div class="flex items-center gap-2">
        <UBadge :color="isAiConnected ? 'primary' : 'warning'" variant="subtle" size="lg" class="px-3 py-1 font-bold">
          <UIcon :name="isAiConnected ? 'i-heroicons-check-circle' : 'i-heroicons-exclamation-triangle'" :class="isAiConnected ? 'text-emerald-500' : 'text-amber-500'" class="w-4 h-4 mr-1" />
          {{ isAiConnected ? '4 Trained Models Active & Synchronized' : (usingCachedRealData ? 'Offline - Displaying Saved Real Data' : 'AI Backend Disconnected') }}
        </UBadge>
      </div>
    </div>

    <!-- Connection Warning Alert Banner -->
    <UAlert
      v-if="!isAiConnected"
      icon="i-heroicons-exclamation-triangle"
      color="warning"
      variant="solid"
      class="mb-6 border border-amber-500 shadow-md font-medium"
      :title="usingCachedRealData ? '⚠️ Peringatan: Service AI Backend Tidak Terhubung (Menampilkan Data Real Terakhir)' : '⚠️ Peringatan: Fitur AI Analytics Belum Berjalan / Tersambung ke Backend'"
      :description="usingCachedRealData
        ? `Fitur AI Analytics saat ini menampilkan data real terakhir yang berhasil disinkronkan pada ${lastSyncedTime || 'sesi sebelumnya'} sebelum koneksi terputus. Mohon pastikan Service Python AI (Port 8000) atau Go Analytics (Port 8084) telah dinyalakan di backend.`
        : 'Fitur AI Analytics saat ini belum tersambung ke backend. Mohon pastikan Service Python AI (Port 8000) atau Go Analytics (Port 8084) telah dinyalakan di backend.'"
    />

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center items-center h-96">
      <div class="flex flex-col items-center gap-4">
        <UIcon name="i-heroicons-cpu-chip" class="w-12 h-12 animate-pulse text-indigo-500" />
        <span class="text-sm font-semibold text-gray-500 animate-pulse">Menghubungkan & Memuat Model AI...</span>
      </div>
    </div>

    <div v-else class="space-y-8">
      <!-- ═══ Summary Stats ═══ -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <UCard class="border border-indigo-100 dark:border-indigo-900/50">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-indigo-500/10 flex items-center justify-center">
              <UIcon name="i-heroicons-chart-bar-square" class="w-5 h-5 text-indigo-500" />
            </div>
            <div>
              <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Entities Scored</div>
              <div class="text-2xl font-black text-indigo-600 dark:text-indigo-400">{{ xgboostState.predictions.length }}</div>
            </div>
          </div>
        </UCard>
        <UCard class="border border-rose-100 dark:border-rose-900/50">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-rose-500/10 flex items-center justify-center">
              <UIcon name="i-heroicons-shield-exclamation" class="w-5 h-5 text-rose-500" />
            </div>
            <div>
              <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Anomalies Detected</div>
              <div class="text-2xl font-black text-rose-600 dark:text-rose-400">{{ summary.anomaliesDetected }}</div>
            </div>
          </div>
        </UCard>
        <UCard class="border border-amber-100 dark:border-amber-900/50">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-amber-500/10 flex items-center justify-center">
              <UIcon name="i-heroicons-document-magnifying-glass" class="w-5 h-5 text-amber-500" />
            </div>
            <div>
              <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Docs Analyzed</div>
              <div class="text-2xl font-black text-amber-600 dark:text-amber-400">{{ nlpState.documents.length }}</div>
            </div>
          </div>
        </UCard>
        <UCard class="border border-violet-100 dark:border-violet-900/50">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-violet-500/10 flex items-center justify-center">
              <UIcon name="i-heroicons-bell-alert" class="w-5 h-5 text-violet-500" />
            </div>
            <div>
              <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">KPI Forecasts</div>
              <div class="text-2xl font-black text-violet-600 dark:text-violet-400">{{ timeseriesState.kpiForecasts.length }}</div>
            </div>
          </div>
        </UCard>
      </div>

      <!-- ═══ Tabs ═══ -->
      <UTabs v-model="activeTab" :items="tabItems" class="w-full">
        <template #content="{ item }">
          <!-- ═══════════════════════════════════════════════════════════════ -->
          <!-- TAB 1: RISK SCORING ML (DEPARTMENT PREDICTION)                -->
          <!-- ═══════════════════════════════════════════════════════════════ -->
          <div v-if="item.key === 'xgboost'" class="space-y-6 py-6">
            <UAlert
              icon="i-heroicons-cpu-chip"
              color="primary"
              variant="subtle"
              title="Department & Entity Risk Scoring ML Model"
              description="Model Machine Learning memprediksi Impact dan Likelihood departemen/cabang secara dinamis berdasarkan achievement KPI, temuan audit lalu, dan volatilitas."
            />

            <!-- Predicted vs Actual Chart -->
            <UCard class="lg:col-span-2">
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon name="i-heroicons-chart-bar" class="text-indigo-500" />
                  <h3 class="font-bold">Predicted vs Actual Risk Score</h3>
                </div>
              </template>
              <div class="h-80">
                <Bar :data="xgboostBarData" :options="xgboostBarOptions" />
              </div>
            </UCard>

            <!-- Feature Importance -->
            <UCard>
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon name="i-heroicons-adjustments-horizontal" class="text-orange-500" />
                  <h3 class="font-bold">Feature Importance (Top Predictors)</h3>
                </div>
              </template>
              <div class="h-72">
                <Bar :data="featureBarData" :options="featureBarOptions" />
              </div>
            </UCard>

            <!-- Predictions Table -->
            <UCard>
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon name="i-heroicons-table-cells" class="text-blue-500" />
                  <h3 class="font-bold">Entity Risk Score Predictions</h3>
                </div>
              </template>
              <div class="overflow-x-auto">
                <table class="w-full text-sm">
                  <thead>
                    <tr class="border-b border-gray-200 dark:border-gray-700">
                      <th class="text-left py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Entity</th>
                      <th class="text-left py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Type</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Kategori Risiko</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Target Periode</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Pred. Likelihood</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Pred. Impact</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Pred. Score</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Pred. Risk Level</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Actual</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Actual Risk Level</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Trend</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="row in xgboostState.predictions" :key="row.entity" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                      <td class="py-3 px-4 font-bold">{{ row.entity }}</td>
                      <td class="py-3 px-4"><UBadge :color="row.type === 'Branch' ? 'primary' : 'warning'" variant="subtle" size="md">{{ row.type }}</UBadge></td>
                      <td class="text-center py-3 px-4"><UBadge :color="riskCategoryColor(row.riskCategory)" variant="subtle" size="md">{{ row.riskCategory }}</UBadge></td>
                      <td class="text-center py-3 px-4"><UBadge color="info" variant="subtle" size="md" class="font-bold">{{ row.targetTimeline || 'Q3 2026' }}</UBadge></td>
                      <td class="text-center py-3 px-4 font-mono">{{ formatNum(row.predictedLikelihood, 1) }}</td>
                      <td class="text-center py-3 px-4 font-mono">{{ formatNum(row.predictedImpact, 1) }}</td>
                      <td class="text-center py-3 px-4 font-mono font-bold">{{ formatNum(row.predictedScore, 1) }}</td>
                      <td class="text-center py-3 px-4">
                        <UBadge
                          :style="{ backgroundColor: getRiskConfig(row.predictedRiskLevel).color, color: 'white' }"
                          variant="solid"
                          size="md"
                          class="font-bold"
                        >
                          {{ getRiskConfig(row.predictedRiskLevel).label }}
                        </UBadge>
                      </td>
                      <td class="text-center py-3 px-4 font-mono">{{ formatNum(row.actualScore, 1) }}</td>
                      <td class="text-center py-3 px-4">
                        <UBadge
                          :style="{ backgroundColor: getRiskConfig(row.actualRiskLevel).color, color: 'white' }"
                          variant="solid"
                          size="md"
                          class="font-bold"
                        >
                          {{ getRiskConfig(row.actualRiskLevel).label }}
                        </UBadge>
                      </td>
                      <td class="text-center py-3 px-4">
                        <UIcon :name="trendIcon(row.trend)" class="w-5 h-5" :class="trendColor(row.trend)" />
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </UCard>
          </div>

          <!-- ═══════════════════════════════════════════════════════════════ -->
          <!-- TAB 2: ANOMALY DETECTION ML                                   -->
          <!-- ═══════════════════════════════════════════════════════════════ -->
          <div v-if="item.key === 'isolation'" class="space-y-6 py-6">
            <UAlert
              icon="i-heroicons-shield-exclamation"
              color="warning"
              variant="subtle"
              title="Transaction Anomaly Detection Model"
              description="Model Machine Learning memindai pola transaksi (jumlah, jam operasional, hari, penerima baru, pola pembulatan) untuk mendeteksi anomali fraud secara real-time."
            />

            <!-- Summary Cards -->
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <UCard>
                <div class="text-center">
                  <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Records Scanned</div>
                  <div class="text-2xl font-black mt-1">{{ (isolationState.summary?.totalScanned || 0).toLocaleString() }}</div>
                </div>
              </UCard>
              <UCard>
                <div class="text-center">
                  <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Anomalies Found</div>
                  <div class="text-2xl font-black text-rose-500 mt-1">{{ summary.anomaliesDetected }}</div>
                </div>
              </UCard>
              <UCard>
                <div class="text-center">
                  <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Contamination Rate</div>
                  <div class="text-2xl font-black text-amber-500 mt-1">{{ formatNum((isolationState.summary?.contaminationRate || 0) * 100, 1) }}%</div>
                </div>
              </UCard>
              <UCard>
                <div class="text-center">
                  <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Top Category</div>
                  <div class="text-2xl font-black text-indigo-500 mt-1">{{ isolationState.summary?.topCategory || 'Transaction' }}</div>
                </div>
              </UCard>
            </div>

            <!-- Scatter Chart Card with Type Filter Tabs -->
            <UCard>
              <template #header>
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div class="flex items-center gap-2">
                    <UIcon name="i-heroicons-chart-bar" class="text-blue-500" />
                    <h3 class="font-bold">Data Pattern Distribution per Anomaly Type</h3>
                  </div>
                  <!-- Anomaly Type Selector Buttons -->
                  <div class="flex flex-wrap gap-1">
                    <UButton
                      v-for="t in availableAnomalyTypes"
                      :key="t"
                      :color="selectedAnomalyType === t ? 'primary' : 'neutral'"
                      :variant="selectedAnomalyType === t ? 'solid' : 'ghost'"
                      size="xs"
                      @click="selectedAnomalyType = t"
                    >
                      {{ t }}
                    </UButton>
                  </div>
                </div>
              </template>
              <div class="h-80">
                <Scatter :data="scatterChartData" :options="scatterOptions" />
              </div>
            </UCard>

            <!-- Anomalies Table -->
            <UCard>
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon name="i-heroicons-table-cells" class="text-rose-500" />
                  <h3 class="font-bold">Detected Anomalies Detail</h3>
                </div>
              </template>
              <div class="overflow-x-auto">
                <table class="w-full text-sm">
                  <thead>
                    <tr class="border-b border-gray-200 dark:border-gray-700">
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">ID</th>
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Entity</th>
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Type</th>
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[300px]">Description</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Risk Level</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Date</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="a in isolationState.anomalies" :key="a.id" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                      <td class="py-3 px-3 font-mono font-bold text-md">{{ a.id }}</td>
                      <td class="py-3 px-3 font-bold">{{ a.entity }}</td>
                      <td class="py-3 px-3"><UBadge color="neutral" variant="subtle" size="md">{{ a.type }}</UBadge></td>
                      <td class="py-3 px-3 text-md leading-relaxed text-gray-600 dark:text-gray-300">{{ a.description }}</td>
                      <td class="text-center py-3 px-3">
                        <UBadge
                          :style="{ backgroundColor: getRiskConfig(a.riskLevel).color, color: 'white' }"
                          variant="solid"
                          size="md"
                          class="font-bold"
                        >
                          {{ getRiskConfig(a.riskLevel).label }}
                        </UBadge>
                      </td>
                      <td class="text-center py-3 px-3 text-md text-gray-400">{{ a.date }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </UCard>
          </div>

          <!-- ═══════════════════════════════════════════════════════════════ -->
          <!-- TAB 3: IndoBERT NLP ANALYSIS                                  -->
          <!-- ═══════════════════════════════════════════════════════════════ -->
          <div v-if="item.key === 'nlp'" class="space-y-6 py-6">
            <UAlert
              icon="i-heroicons-document-magnifying-glass"
              color="success"
              variant="subtle"
              title="IndoBERT NLP Document Analysis Model"
              description="Model IndoBERT & Natural Language Processing menganalisis kutipan Laporan Hasil Audit (Bahasa Indonesia) untuk mengklasifikasi kategori risiko, sentimen, impact, dan likelihood secara otomatis."
            />

            <!-- Charts Row -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <UCard>
                <template #header>
                  <h3 class="font-bold text-sm">Auto-Classified Risk Categories</h3>
                </template>
                <div class="h-60">
                  <Doughnut :data="categoryDoughnutData" :options="doughnutOptions" />
                </div>
              </UCard>
              <UCard>
                <template #header>
                  <h3 class="font-bold text-sm">Sentiment Distribution</h3>
                </template>
                <div class="h-60">
                  <Doughnut :data="sentimentDoughnutData" :options="doughnutOptions" />
                </div>
              </UCard>
            </div>

            <!-- Documents Table -->
            <UCard>
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon name="i-heroicons-document-text" class="text-violet-500" />
                  <h3 class="font-bold">Analyzed Documents</h3>
                </div>
              </template>
              <div class="overflow-x-auto">
                <table class="w-full text-sm">
                  <thead>
                    <tr class="border-b border-gray-200 dark:border-gray-700">
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Doc ID</th>
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Title</th>
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Source</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Category</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Sentiment</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Risk Level</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="doc in nlpState.documents" :key="doc.docId" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                      <td class="py-3 px-3 font-mono font-bold text-md">{{ doc.docId }}</td>
                      <td class="py-3 px-3">
                        <div class="font-bold text-md">{{ doc.title }}</div>
                        <div class="text-[11px] text-gray-400 mt-0.5 italic truncate max-w-[300px]">"{{ doc.excerpt }}"</div>
                      </td>
                      <td class="py-3 px-3"><UBadge :color="doc.source === 'Working Paper' ? 'primary' : 'warning'" variant="subtle" size="md">{{ doc.source }}</UBadge></td>
                      <td class="text-center py-3 px-3"><UBadge :color="riskCategoryColor(doc.autoCategory)" variant="subtle" size="md">{{ doc.autoCategory }}</UBadge></td>
                      <td class="text-center py-3 px-3"><UBadge :color="sentimentColor(doc.sentiment)" variant="subtle" size="md">{{ doc.sentiment }}</UBadge></td>
                      <td class="text-center py-3 px-3">
                        <UBadge
                          :style="{ backgroundColor: getRiskConfig(doc.riskLevel).color, color: 'white' }"
                          variant="solid"
                          size="md"
                          class="font-bold"
                        >
                          {{ getRiskConfig(doc.riskLevel).label }}
                        </UBadge>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </UCard>
          </div>

          <!-- ═══════════════════════════════════════════════════════════════ -->
          <!-- TAB 4: TIME-SERIES KPI PYTORCH LSTM                            -->
          <!-- ═══════════════════════════════════════════════════════════════ -->
          <div v-if="item.key === 'timeseries'" class="space-y-6 py-6">
            <UAlert
              icon="i-heroicons-arrow-trending-up"
              color="info"
              variant="subtle"
              title="Time-Series KPI PyTorch LSTM Forecasting Model"
              description="Model Deep Learning (PyTorch LSTM) memprediksi tren performa KPI masa depan berdasarkan deret waktu historis untuk peringatan dini perencanaan audit."
            />

            <!-- Forecast Chart -->
            <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
              <UCard class="lg:col-span-2">
                <template #header>
                  <div class="flex items-center gap-2">
                    <UIcon name="i-heroicons-chart-bar" class="text-violet-500" />
                    <h3 class="font-bold">KPI Performance Forecast</h3>
                    <UBadge color="info" variant="subtle" size="md" class="ml-auto">Shaded area = confidence interval</UBadge>
                  </div>
                </template>
                <div class="h-80">
                  <Line :data="timeSeriesChartData" :options="timeSeriesOptions" />
                </div>
              </UCard>

              <!-- At-Risk Departments -->
              <div class="space-y-6">
                <UCard>
                  <template #header>
                    <h3 class="font-bold text-sm">⚠️ At-Risk Departments</h3>
                  </template>
                  <div class="space-y-3">
                    <div v-for="dept in timeseriesState.atRiskDepartments" :key="dept.department" class="p-3 bg-rose-50 dark:bg-rose-900/20 rounded-lg border border-rose-200 dark:border-rose-800">
                      <div class="flex items-center justify-between gap-2">
                        <div class="font-bold text-md">{{ dept.department }}</div>
                        <UBadge
                          :style="{ backgroundColor: getRiskConfig(dept.riskLevel).color, color: 'white' }"
                          variant="solid"
                          size="md"
                          class="font-bold shrink-0"
                        >
                          {{ getRiskConfig(dept.riskLevel).label }}
                        </UBadge>
                      </div>
                      <div class="text-[10px] text-gray-500 mt-0.5">{{ dept.kpi }}</div>
                      <div class="flex items-center gap-2 mt-1.5">
                        <UIcon name="i-heroicons-arrow-trending-down" class="w-4 h-4 text-rose-500" />
                        <span class="text-md font-bold text-rose-500">{{ formatNum(dept.predictedQ3, 1) }}% projected Q3</span>
                      </div>
                    </div>
                  </div>
                </UCard>
              </div>
            </div>

            <!-- KPI Forecasts Table -->
            <UCard>
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon name="i-heroicons-table-cells" class="text-indigo-500" />
                  <h3 class="font-bold">KPI Forecast Summary</h3>
                </div>
              </template>
              <div class="overflow-x-auto">
                <table class="w-full text-sm">
                  <thead>
                    <tr class="border-b border-gray-200 dark:border-gray-700">
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Code</th>
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">KPI</th>
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Entity</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Type</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Horizon Prediksi</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Current</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Forecast</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Trend</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Risk Level</th>
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[250px]">Recommended Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="kpi in timeseriesState.kpiForecasts" :key="kpi.code" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                      <td class="py-3 px-3 font-mono font-bold text-md">{{ kpi.code }}</td>
                      <td class="py-3 px-3 font-bold text-md">{{ kpi.kpiName }}</td>
                      <td class="py-3 px-3 font-bold text-md">{{ kpi.entity || '-' }}</td>
                      <td class="text-center py-3 px-3"><UBadge :color="(kpi.entityType || kpi.type) === 'Branch' ? 'primary' : 'warning'" variant="subtle" size="md">{{ kpi.entityType || kpi.type || '-' }}</UBadge></td>
                      <td class="text-center py-3 px-3"><UBadge color="info" variant="subtle" size="md" class="font-bold">{{ kpi.targetHorizon || 'Q3 2026' }}</UBadge></td>
                      <td class="text-center py-3 px-3 font-mono">{{ formatNum(kpi.currentValue, 1) }}{{ kpi.unit === '%' ? '%' : '' }}</td>
                      <td class="text-center py-3 px-3 font-mono font-bold">{{ formatNum(kpi.forecastedValue, 1) }}{{ kpi.unit === '%' ? '%' : '' }}</td>
                      <td class="text-center py-3 px-3">
                        <div class="flex items-center justify-center gap-1">
                          <UIcon :name="trendIcon(kpi.trend)" class="w-4 h-4" :class="trendColor(kpi.trend)" />
                          <span class="text-md font-bold" :class="trendColor(kpi.trend)">{{ kpi.trend }}</span>
                        </div>
                      </td>
                      <td class="text-center py-3 px-3">
                        <UBadge
                          :style="{ backgroundColor: getRiskConfig(kpi.riskLevel).color, color: 'white' }"
                          variant="solid"
                          size="md"
                          class="font-bold"
                        >
                          {{ getRiskConfig(kpi.riskLevel).label }}
                        </UBadge>
                      </td>
                      <td class="py-3 px-3 text-md text-gray-600 dark:text-gray-300 leading-relaxed">{{ kpi.recommendedAction }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </UCard>
          </div>
        </template>
      </UTabs>
    </div>
  </div>
</template>