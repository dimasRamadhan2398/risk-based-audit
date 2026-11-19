<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import { toEntityRiskCode } from '~/utils/riskLevelBadge'
import { useI18n } from '~/composables/useI18n'
import {
  buildKpiForecastChart,
  selectAtRiskKpis,
  toKpiForecastView,
  type KpiForecastView
} from '~/composables/useKpiForecast'
import {
  AUDITOR_CHART_ALL_KEY,
  buildAuditorChartView,
  type AuditorTimeSeriesPoint
} from '~/utils/auditorKpiSync'
import { formatKpiForecastValue, formatRupiahCompact, getKpiMeta } from '~/utils/analyticsFormat'
import AiInsightHeader from '~/components/analytics/AiInsightHeader.vue'
import AiModelStatusBadge from '~/components/analytics/AiModelStatusBadge.vue'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

interface AuditorKpiItem {
  code: string
  kpiName: string
  category: string
  strategicObjective?: string
  hibHig: string
  unit: string
  targetHorizon?: string
  targetValue: number
  currentValue: number
  forecastedValue: number
  achievementRate?: number
  forecastedAchievementRate: number
  currentStatus?: string
  predictedStatus: string
  trend: string
  confidenceLower?: number
  confidenceUpper?: number
  aiNarrative?: string
  recommendedAction?: string
}

interface AuditeeKpiItem {
  code: string
  kpiName: string
  unit?: string
  entity?: string
  entityType?: string
  syncFrequency?: string
  targetHorizon?: string
  currentValue: number
  forecastedValue: number
  deltaPct?: number
  trend: string
  impactScore?: number
  likelihoodScore?: number
  riskLevel: string
  aiNarrative?: string
  recommendedAction?: string
}

interface AtRiskPillarItem {
  pillar?: string
  code: string
  kpi?: string
  kpiName?: string
  target?: number | string
  targetValue?: number | string
  predicted?: number
  achievementRate?: number
  forecastedAchievementRate?: number
  predictedStatus: string
}

interface AtRiskDepartmentItem {
  code?: string
  department: string
  kpi: string
  unit?: string
  currentTrend?: number
  predictedQ3?: number
  predictedValue?: number
  predictedDrop?: string
  riskLevel: string
}

const { t, locale } = useI18n()
const toast = useToast()
const {
  loading,
  timeseriesState,
  retrainStatus,
  checkingRetrainStatus,
  hasUntrainedNewData,
  untrainedDataReasons,
  canManualRetrain,
  formatNum,
  getRiskConfig,
  trendColor,
  trendIcon,
  fetchRetrainStatus,
  triggerManualForceRetrain,
  fetchAiAnalytics
} = useAiAnalytics()

// Perspective switch: 'auditor' (Strategic Audit Plan) vs 'auditee' (Business Health Gold Zone)
const selectedPerspective = ref<'auditor' | 'auditee'>('auditor')
const isRetraining = ref(false)
const localTrainingStartedAt = ref<number | null>(null)
const trainingElapsedSec = ref(0)
let retrainPollTimer: ReturnType<typeof setInterval> | null = null
let elapsedTickerTimer: ReturnType<typeof setInterval> | null = null
let statusRetryTimer: ReturnType<typeof setInterval> | null = null

const isTrainingActive = computed(() => Boolean(retrainStatus.value.isRunning || isRetraining.value))

const updateElapsedSeconds = () => {
  const serverStart = retrainStatus.value.startedAt ? Date.parse(retrainStatus.value.startedAt) : NaN
  const baseMs = Number.isFinite(serverStart) ? serverStart : localTrainingStartedAt.value
  trainingElapsedSec.value = baseMs ? Math.max(0, Math.round((Date.now() - baseMs) / 1000)) : 0
}

const stopElapsedTicker = () => {
  if (elapsedTickerTimer) {
    clearInterval(elapsedTickerTimer)
    elapsedTickerTimer = null
  }
  localTrainingStartedAt.value = null
  trainingElapsedSec.value = 0
}

const startElapsedTicker = () => {
  if (!localTrainingStartedAt.value) localTrainingStartedAt.value = Date.now()
  updateElapsedSeconds()
  if (elapsedTickerTimer || typeof window === 'undefined') return
  elapsedTickerTimer = setInterval(updateElapsedSeconds, 1000)
}

const stopRetrainPolling = () => {
  if (retrainPollTimer) {
    clearInterval(retrainPollTimer)
    retrainPollTimer = null
  }
}

const stopStatusRetry = () => {
  if (statusRetryTimer) {
    clearInterval(statusRetryTimer)
    statusRetryTimer = null
  }
}

const startRetrainPolling = () => {
  startElapsedTicker()
  if (retrainPollTimer || typeof window === 'undefined') return
  retrainPollTimer = setInterval(async () => {
    const st = await fetchRetrainStatus()
    if (st && !st.isRunning) {
      stopRetrainPolling()
      stopElapsedTicker()
      await fetchAiAnalytics(true, { skipAutoRetrain: true })
      if (st.status === 'completed') {
        toast.add({
          title: 'Pelatihan Otomatis Model AI Selesai',
          description: st.message || 'Model AI di belakang layar selesai diperbarui.',
          color: 'success',
          icon: 'i-lucide-check-circle'
        })
      }
    }
  }, 2500)
}

const startStatusReconnectRetries = () => {
  if (statusRetryTimer || typeof window === 'undefined') return
  let attempts = 0
  statusRetryTimer = setInterval(async () => {
    if (retrainStatus.value.statusFetchedOnce || attempts >= 3) {
      stopStatusRetry()
      return
    }
    attempts++
    const st = await fetchRetrainStatus()
    if (st) {
      stopStatusRetry()
      if (st.isRunning) startRetrainPolling()
    }
  }, 4000)
}

watch(isTrainingActive, (running) => {
  if (running) {
    startRetrainPolling()
  } else {
    stopRetrainPolling()
    stopElapsedTicker()
  }
})

onMounted(async () => {
  await fetchAiAnalytics(true)
  if (retrainStatus.value.isRunning) {
    startRetrainPolling()
  } else if (!retrainStatus.value.statusFetchedOnce) {
    startStatusReconnectRetries()
  }
})

onUnmounted(() => {
  stopRetrainPolling()
  stopElapsedTicker()
  stopStatusRetry()
})

watch(() => timeseriesState.value.auditeeForecast?.kpiForecasts?.length, (newLen, oldLen) => {
  if (oldLen !== undefined && newLen !== undefined && newLen > oldLen) {
    selectedPerspective.value = 'auditee'
  }
})

const isManualRetrainAllowed = computed(() => canManualRetrain.value && !isRetraining.value)

const retrainButtonLabel = computed(() => {
  if (isTrainingActive.value) {
    return trainingElapsedSec.value > 0
      ? `Sedang Melatih (${trainingElapsedSec.value} dtk)...`
      : 'Sedang Melatih Model AI...'
  }
  if (hasUntrainedNewData.value) return 'Manual / Force Sync & Retrain'
  if (retrainStatus.value.statusFetchedOnce) return 'Model Sudah Terlatih'
  if (checkingRetrainStatus.value) return 'Memeriksa Status AI...'
  return 'Status AI Belum Terhubung'
})

const retrainButtonTooltip = computed(() => {
  if (isTrainingActive.value) {
    return 'Model AI sedang aktif melatih ulang bobot PyTorch LSTM di server. Tunggu hingga proses pelatihan selesai.'
  }
  if (isManualRetrainAllowed.value) {
    const detail = untrainedDataReasons.value.length ? ` (${untrainedDataReasons.value.join(', ')})` : ''
    return `Data baru belum dilatih ke model AI${detail}. Klik untuk Manual / Force Sync & Retrain sekarang.`
  }
  if (retrainStatus.value.statusFetchedOnce) {
    return 'Semua data terbaru sudah dilatih ke model AI — tombol dinonaktifkan agar tidak terjadi spam pelatihan ulang.'
  }
  return 'Belum menerima status pelatihan dari layanan AI (/retrain/status). Pastikan layanan analytics-service / python-ai terbaru aktif, lalu klik Cek Ulang Status AI.'
})

const handleManualRetrain = async () => {
  if (!isManualRetrainAllowed.value) return
  isRetraining.value = true
  localTrainingStartedAt.value = Date.now()
  startElapsedTicker()
  try {
    const res: any = await triggerManualForceRetrain()
    if (res?.status === 'already_running') {
      toast.add({
        title: 'Model AI Sedang Berjalan di Belakang Layar',
        description: 'Pelatihan otomatis sedang berlangsung di server AI, tidak perlu Manual / Force Sync.',
        color: 'info',
        icon: 'i-lucide-cpu'
      })
      startRetrainPolling()
      return
    }
    if (res?.status === 'failed') {
      toast.add({
        title: 'Pelatihan Ulang Model Gagal',
        description: res.lastError || res.message || 'Terjadi kendala saat melatih ulang model AI.',
        color: 'error',
        icon: 'i-lucide-alert-triangle'
      })
      return
    }
    toast.add({
      title: 'Manual / Force Sync & Retrain Selesai',
      description: res?.message || `Model AI selesai dilatih ulang dalam ${res?.durationSeconds ?? 0} detik.`,
      color: 'success',
      icon: 'i-lucide-check-circle'
    })
  } catch (err: unknown) {
    console.warn('Retrain error:', err)
  } finally {
    isRetraining.value = false
    if (!retrainStatus.value.isRunning) {
      stopElapsedTicker()
    }
  }
}

// Active forecast list
const auditorForecasts = computed<AuditorKpiItem[]>(() => timeseriesState.value.auditorForecast?.kpiForecasts || [])
const auditeeForecasts = computed<AuditeeKpiItem[]>(() => timeseriesState.value.auditeeForecast?.kpiForecasts || timeseriesState.value.kpiForecasts || [])

const auditorAtRiskPillars = computed<AtRiskPillarItem[]>(() => {
  if (timeseriesState.value.auditorForecast?.atRiskPillars?.length) {
    return timeseriesState.value.auditorForecast.atRiskPillars
  }
  return auditorForecasts.value.filter((k: AuditorKpiItem) =>
    k.predictedStatus === 'Poor' || k.predictedStatus === 'Moderate' || k.trend === 'Deteriorating'
  )
})

const auditeeAtRisk = computed<AtRiskDepartmentItem[]>(() => {
  return timeseriesState.value.auditeeForecast?.atRiskDepartments || timeseriesState.value.atRiskDepartments || []
})

// Search, Filter, and Pagination Controls
const searchQuery = ref('')
const selectedStatusFilter = ref('ALL')
const selectedPillarFilter = ref('ALL')
const selectedRiskFilter = ref('ALL')
const currentPage = ref(1)
const pageSize = ref(10)
const pageSizeOptions = [
  { label: '10 per hal', value: 10 },
  { label: '25 per hal', value: 25 },
  { label: '50 per hal', value: 50 },
  { label: 'Semua', value: 9999 }
]

const distinctAuditorPillars = computed(() => {
  const set = new Set<string>()
  auditorForecasts.value.forEach((k: AuditorKpiItem) => {
    if (k.category) set.add(k.category)
  })
  return Array.from(set).sort()
})

// Filtered Auditor KPIs (table & chart share this exact list)
const filteredAuditorForecasts = computed(() => {
  let list = auditorForecasts.value
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter((k: AuditorKpiItem) =>
      Boolean(
        (k.code && k.code.toLowerCase().includes(q))
        || (k.kpiName && k.kpiName.toLowerCase().includes(q))
        || (k.category && k.category.toLowerCase().includes(q))
        || (k.aiNarrative && k.aiNarrative.toLowerCase().includes(q))
        || (k.recommendedAction && k.recommendedAction.toLowerCase().includes(q))
      )
    )
  }
  if (selectedStatusFilter.value !== 'ALL') {
    list = list.filter((k: AuditorKpiItem) => k.predictedStatus === selectedStatusFilter.value)
  }
  if (selectedPillarFilter.value !== 'ALL') {
    list = list.filter((k: AuditorKpiItem) => k.category === selectedPillarFilter.value)
  }
  return list
})

// Filtered Auditee KPIs (table & chart share this exact list)
const filteredAuditeeForecasts = computed(() => {
  let list = auditeeForecasts.value
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter((k: AuditeeKpiItem) =>
      Boolean(
        (k.code && k.code.toLowerCase().includes(q))
        || (k.kpiName && k.kpiName.toLowerCase().includes(q))
        || (k.entity && k.entity.toLowerCase().includes(q))
        || (k.aiNarrative && k.aiNarrative.toLowerCase().includes(q))
        || (k.recommendedAction && k.recommendedAction.toLowerCase().includes(q))
      )
    )
  }
  if (selectedRiskFilter.value !== 'ALL') {
    list = list.filter((k: AuditeeKpiItem) => toEntityRiskCode(k.riskLevel || 'MODERATE') === selectedRiskFilter.value)
  }
  return list
})

// Paginated lists
const paginatedAuditorForecasts = computed(() => {
  if (pageSize.value >= 9999) return filteredAuditorForecasts.value
  const start = (currentPage.value - 1) * pageSize.value
  return filteredAuditorForecasts.value.slice(start, start + pageSize.value)
})

const paginatedAuditeeForecasts = computed(() => {
  if (pageSize.value >= 9999) return filteredAuditeeForecasts.value
  const start = (currentPage.value - 1) * pageSize.value
  return filteredAuditeeForecasts.value.slice(start, start + pageSize.value)
})

const totalAuditorPages = computed(() => {
  if (pageSize.value >= 9999) return 1
  return Math.ceil(filteredAuditorForecasts.value.length / pageSize.value) || 1
})

const totalAuditeePages = computed(() => {
  if (pageSize.value >= 9999) return 1
  return Math.ceil(filteredAuditeeForecasts.value.length / pageSize.value) || 1
})

watch([selectedPerspective, searchQuery, selectedStatusFilter, selectedPillarFilter, selectedRiskFilter, pageSize], () => {
  currentPage.value = 1
})

// ─── Unit Formatters Shared by Charts & Tables ───────────────────────────────
const formatAuditorKpiVal = (val: number | null | undefined, unit = '%'): string => {
  if (val === null || val === undefined || Number.isNaN(Number(val))) return '-'
  const u = String(unit || '%').trim()
  if (u === '%') return `${formatNum(val, 1)}%`
  if (/^(rp|idr)$/i.test(u)) return formatRupiahCompact(Number(val), locale.value)
  return u ? `${formatNum(val, 1)} ${u}` : formatNum(val, 1)
}

const formatAuditeeKpiVal = (val: number | null | undefined, kpiName: string, unit?: string): string => {
  if (val === null || val === undefined || Number.isNaN(Number(val))) return '-'
  const meta = getKpiMeta(kpiName, unit)
  const base = formatKpiForecastValue(Number(val), meta, locale.value)
  const u = String(unit || '').trim()
  if (meta.kind === 'score' && u && !['%', 'score', 'Score', 'rp', 'Rp', 'IDR'].includes(u)) {
    return `${base} ${u}`
  }
  return base
}

const formatAuditeeAtRiskVal = (dept: AtRiskDepartmentItem): string => {
  if (dept.predictedDrop) return dept.predictedDrop
  const rawVal = dept.predictedValue ?? dept.predictedQ3
  const matched = auditeeForecasts.value.find(
    (k: any) => k.code === dept.code || (k.kpiName === dept.kpi && (!dept.department || k.entity === dept.department))
  )
  return formatAuditeeKpiVal(rawVal, dept.kpi, dept.unit || matched?.unit)
}

// ─── Shared View Model (useKpiForecast) Synced with Auditee Table ────────────
const kpiViews = computed<KpiForecastView[]>(() =>
  filteredAuditeeForecasts.value.map(toKpiForecastView)
)
const atRiskKpis = computed(() => selectAtRiskKpis(kpiViews.value))
const selectedKpiCode = ref('')
const selectedKpiName = ref('')
const selectedKpi = computed<KpiForecastView | undefined>(() =>
  (selectedKpiCode.value ? kpiViews.value.find(k => k.code === selectedKpiCode.value) : undefined)
  ?? kpiViews.value.find(k => k.kpiName === selectedKpiName.value)
  ?? atRiskKpis.value[0]
  ?? kpiViews.value[0]
)
watch(selectedKpi, (kpi) => {
  if (kpi) {
    if (kpi.code !== selectedKpiCode.value) selectedKpiCode.value = kpi.code
    if (kpi.kpiName !== selectedKpiName.value) selectedKpiName.value = kpi.kpiName
  }
}, { immediate: true })

const focusAuditeeKpi = (item: { code?: string; kpiName?: string; kpi?: string; department?: string }) => {
  const match = filteredAuditeeForecasts.value.find(
    (k: any) => (item.code && k.code === item.code)
      || (k.kpiName === (item.kpiName || item.kpi) && (!item.department || k.entity === item.department))
  )
  if (match) {
    selectedKpiCode.value = match.code
    selectedKpiName.value = match.kpiName
  }
}

const forecastMape = computed(() => Number(timeseriesState.value.forecastAccuracy?.mape ?? 0))
const kpiChart = computed(() => buildKpiForecastChart(selectedKpi.value, forecastMape.value, t, locale.value, {
  actualColor: 'rgb(59,130,246)',
  forecastColor: 'rgb(139,92,246)',
  bandColor: 'rgba(139,92,246,0.15)'
}))

// ─── AUDITOR TIME SERIES CHART DATA (100% Synced with Auditor Table) ─────────
const selectedAuditorKpiCode = ref<string>(AUDITOR_CHART_ALL_KEY)
watch(filteredAuditorForecasts, (list) => {
  if (selectedAuditorKpiCode.value !== AUDITOR_CHART_ALL_KEY && !list.some(k => k.code === selectedAuditorKpiCode.value)) {
    selectedAuditorKpiCode.value = AUDITOR_CHART_ALL_KEY
  }
})

const auditorChartView = computed(() =>
  buildAuditorChartView(filteredAuditorForecasts.value, selectedAuditorKpiCode.value)
)

const auditorTimeSeriesChartData = computed(() => {
  const view = auditorChartView.value
  const points: AuditorTimeSeriesPoint[] = view.points
  const uSuffix = view.unit === '%' ? '%' : (view.unit ? ` ${view.unit}` : '')
  const targetLabel = view.mode === 'single' && view.targetValue !== null
    ? `Target (${formatAuditorKpiVal(view.targetValue, view.unit)})`
    : 'Target Tahunan (100%)'

  return {
    labels: points.map(p => p.period || ''),
    datasets: [
      {
        label: targetLabel,
        data: points.map(p => p.target ?? view.targetValue ?? 100),
        borderColor: 'rgba(239, 68, 68, 0.75)',
        borderDash: [5, 5],
        borderWidth: 2,
        pointRadius: 0,
        fill: false,
        spanGaps: true
      },
      {
        label: view.mode === 'single' ? `Realisasi Aktual (${uSuffix.trim() || 'Nilai'})` : 'Realisasi Aktual KPI (%)',
        data: points.map(p => p.actual ?? null),
        borderColor: 'rgb(59, 130, 246)',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
        tension: 0.3,
        pointRadius: 5,
        pointBackgroundColor: 'rgb(59, 130, 246)',
        spanGaps: false
      },
      {
        label: view.mode === 'single' ? `Proyeksi AI LSTM (${uSuffix.trim() || 'Nilai'})` : 'Proyeksi AI LSTM (%)',
        data: points.map(p => p.forecast ?? null),
        borderColor: 'rgb(147, 51, 234)',
        backgroundColor: 'rgba(147, 51, 234, 0.1)',
        borderDash: [6, 4],
        tension: 0.3,
        pointRadius: 5,
        pointStyle: 'rectRot',
        pointBackgroundColor: 'rgb(147, 51, 234)',
        spanGaps: false
      },
      {
        label: t('analytics.timeseries.labelUpperBound'),
        data: points.map(p => p.upperBound ?? null),
        borderColor: 'transparent',
        backgroundColor: 'rgba(147, 51, 234, 0.08)',
        fill: '+1',
        tension: 0.3,
        pointRadius: 0,
        spanGaps: false
      },
      {
        label: t('analytics.timeseries.labelLowerBound'),
        data: points.map(p => p.lowerBound ?? null),
        borderColor: 'transparent',
        backgroundColor: 'rgba(147, 51, 234, 0.08)',
        fill: '-1',
        tension: 0.3,
        pointRadius: 0,
        spanGaps: false
      }
    ]
  }
})

// ─── AUDITEE TIME SERIES CHART DATA (100% Synced via useKpiForecast) ─────────
const auditeeTimeSeriesChartData = computed(() => {
  if (kpiChart.value.data.labels.length > 0) {
    return kpiChart.value.data
  }
  return { labels: [], datasets: [] }
})

const activeChartData = computed(() => {
  return selectedPerspective.value === 'auditor' ? auditorTimeSeriesChartData.value : auditeeTimeSeriesChartData.value
})

const activeChartHasData = computed(() => activeChartData.value.labels.length > 0)

const timeSeriesOptions = computed(() => {
  if (selectedPerspective.value === 'auditee' && kpiChart.value.data.labels.length > 0) {
    return kpiChart.value.options
  }
  const unit = auditorChartView.value.unit
  const points = auditorChartView.value.points
  const isPercent = unit === '%'
  const allVals = points.flatMap(p => [p.actual, p.forecast, p.target, p.upperBound, p.lowerBound].filter((n): n is number => typeof n === 'number' && Number.isFinite(n)))
  const minVal = allVals.length ? Math.min(...allVals) : 60
  const maxVal = allVals.length ? Math.max(...allVals) : 105

  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { intersect: false, mode: 'index' as const },
    plugins: {
      legend: {
        position: 'top' as const,
        labels: {
          filter: (item: { text?: string }) =>
            ![t('analytics.timeseries.labelUpperBound'), t('analytics.timeseries.labelLowerBound')].includes(item.text || '')
        }
      },
      tooltip: {
        callbacks: {
          label: (ctx: { dataset: { label?: string }, parsed: { y: number | null } }) => {
            if (ctx.parsed?.y === null || ctx.parsed?.y === undefined) return ''
            const label = ctx.dataset.label || ''
            if ([t('analytics.timeseries.labelUpperBound'), t('analytics.timeseries.labelLowerBound')].includes(label)) return ''
            return `${label}: ${formatAuditorKpiVal(ctx.parsed.y, unit)}`
          }
        }
      }
    },
    scales: {
      y: {
        title: {
          display: true,
          text: auditorChartView.value.mode === 'single'
            ? `${auditorChartView.value.title} (${unit || '%'})`
            : 'Tingkat Ketercapaian Target KPI Auditor (%)'
        },
        ticks: {
          callback: (v: any) => formatAuditorKpiVal(Number(v), unit)
        },
        suggestedMin: isPercent ? Math.min(60, Math.floor(minVal * 0.9)) : undefined,
        suggestedMax: isPercent ? Math.max(105, Math.ceil(maxVal * 1.05)) : undefined
      }
    }
  }
})

// Status Badge Config for Auditor KPI (Good / Moderate / Poor)
const getStatusBadgeConfig = (status: string) => {
  switch (status) {
    case 'Good':
      return { label: 'Good (≥100%)', color: 'success', badgeClass: 'bg-emerald-500 text-white dark:bg-emerald-600' }
    case 'Moderate':
      return { label: 'Moderate (70-99%)', color: 'warning', badgeClass: 'bg-amber-500 text-white dark:bg-amber-600' }
    case 'Poor':
      return { label: 'Poor (<70%)', color: 'error', badgeClass: 'bg-rose-500 text-white dark:bg-rose-600' }
    default:
      return { label: status || 'Unknown', color: 'neutral', badgeClass: 'bg-gray-500 text-white' }
  }
}
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8 space-y-6">
    <AiInsightHeader
      :title="t('navigation.kpiForecast')"
      :subtitle="t('analytics.subtitle')"
      icon="i-lucide-chart-line"
      :current-sub-module="'kpi-forecast'"
    />

    <!-- Perspective Tabs Switcher & Dynamic Sync Toolbar -->
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-2 bg-gray-50 dark:bg-gray-900/60 rounded-2xl border border-gray-200 dark:border-gray-800">
      <div class="flex items-center gap-1.5 p-1 bg-gray-200/70 dark:bg-gray-800 rounded-xl w-full sm:w-auto">
        <button
          type="button"
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm transition-all flex-1 sm:flex-initial justify-center"
          :class="selectedPerspective === 'auditor'
            ? 'bg-white dark:bg-gray-900 text-primary-600 dark:text-primary-400 font-bold shadow-xs'
            : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white font-medium'"
          @click="selectedPerspective = 'auditor'"
        >
          <UIcon
            name="i-lucide-clipboard-check"
            class="w-4 h-4 text-primary-500"
          />
          <span>{{ t('analytics.timeseries.tabAuditor') }}</span>
          <UBadge
            color="primary"
            variant="subtle"
            size="xs"
            class="ml-1 font-mono"
          >
            {{ auditorForecasts.length }} KPI
          </UBadge>
        </button>

        <button
          type="button"
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm transition-all flex-1 sm:flex-initial justify-center"
          :class="selectedPerspective === 'auditee'
            ? 'bg-white dark:bg-gray-900 text-indigo-600 dark:text-indigo-400 font-bold shadow-xs'
            : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white font-medium'"
          @click="selectedPerspective = 'auditee'"
        >
          <UIcon
            name="i-lucide-building-2"
            class="w-4 h-4 text-indigo-500"
          />
          <span>{{ t('analytics.timeseries.tabAuditee') }}</span>
          <UBadge
            color="info"
            variant="subtle"
            size="xs"
            class="ml-1 font-mono"
          >
            {{ auditeeForecasts.length }} KPI (Gold/CAATT)
          </UBadge>
        </button>
      </div>

      <!-- Sync & Manual Force Retrain Button (Disabled until untrained data appears) -->
      <div class="flex items-center gap-2 w-full sm:w-auto justify-end">
        <UBadge
          :color="isTrainingActive ? 'info' : (hasUntrainedNewData ? 'warning' : (retrainStatus.statusFetchedOnce ? 'success' : 'neutral'))"
          variant="subtle"
          size="sm"
          class="hidden md:flex items-center gap-1.5 text-xs"
        >
          <span
            class="size-2 rounded-full"
            :class="isTrainingActive ? 'bg-sky-500 animate-ping' : (hasUntrainedNewData ? 'bg-amber-500 animate-pulse' : (retrainStatus.statusFetchedOnce ? 'bg-emerald-500' : 'bg-gray-400'))"
          />
          <span v-if="isTrainingActive">Sedang Melatih Model AI</span>
          <span v-else-if="hasUntrainedNewData">Ada Data Belum Dilatih</span>
          <span v-else-if="retrainStatus.statusFetchedOnce">Sudah Terlatih</span>
          <span v-else>Status AI Belum Terhubung</span>
        </UBadge>
        <UButton
          :loading="isTrainingActive || checkingRetrainStatus"
          :disabled="!isManualRetrainAllowed"
          :icon="isManualRetrainAllowed ? 'i-lucide-refresh-cw' : (isTrainingActive ? 'i-lucide-cpu' : (retrainStatus.statusFetchedOnce ? 'i-lucide-check' : 'i-lucide-wifi-off'))"
          :color="isManualRetrainAllowed ? 'warning' : (retrainStatus.statusFetchedOnce ? 'success' : 'neutral')"
          :variant="isManualRetrainAllowed ? 'solid' : 'soft'"
          size="sm"
          class="font-semibold rounded-xl disabled:opacity-75 disabled:cursor-not-allowed"
          :title="retrainButtonTooltip"
          @click="handleManualRetrain"
        >
          {{ retrainButtonLabel }}
        </UButton>
      </div>
    </div>

    <!-- Per-Model Granular Training & Queue Status Badge -->
    <AiModelStatusBadge
      :model-key="selectedPerspective === 'auditor' ? 'kpi_auditor' : 'kpi_auditee'"
    />

    <!-- Real-Time AI Training & Sync Status Alert -->
    <UAlert
      v-if="isTrainingActive"
      color="info"
      variant="subtle"
      icon="i-lucide-cpu"
      :title="trainingElapsedSec > 0 ? `Model AI Sedang Dilatih di Belakang Layar (${trainingElapsedSec} dtk)` : 'Model AI Sedang Dilatih di Belakang Layar'"
      description="Proses pelatihan ulang model PyTorch LSTM sedang aktif berjalan di server AI. Anda tidak perlu menekan Manual / Force Sync — status akan otomatis berubah menjadi 'Model Sudah Terlatih' setelah pelatihan selesai."
    />
    <UAlert
      v-else-if="hasUntrainedNewData"
      color="warning"
      variant="subtle"
      icon="i-lucide-alert-circle"
      title="Data Baru Belum Dilatih oleh Model AI"
      :description="`Terdeteksi pembaruan pada: ${untrainedDataReasons.join(' • ')}. Klik tombol 'Manual / Force Sync & Retrain' untuk melatih ulang model PyTorch LSTM dengan data tersebut.`"
    />
    <UAlert
      v-else-if="retrainStatus.status === 'failed' && retrainStatus.lastError"
      color="error"
      variant="subtle"
      icon="i-lucide-alert-triangle"
      title="Pelatihan Ulang Model AI Gagal"
      :description="retrainStatus.lastError"
    />
    <UAlert
      v-else-if="retrainStatus.statusFetchedOnce"
      color="success"
      variant="subtle"
      icon="i-lucide-shield-check"
      title="Semua Data Baru Sudah Dilatih oleh Model AI"
      :description="`Model PyTorch LSTM sudah terlatih & mutakhir (${retrainStatus.trainedAuditorCount} KPI Auditor, ${retrainStatus.trainedAuditeeCount} baris pool Auditee${retrainStatus.durationSeconds ? `, durasi ${retrainStatus.durationSeconds} dtk` : ''}). Tombol dinonaktifkan hingga ada data baru.`"
    />
    <UAlert
      v-else-if="!loading"
      color="neutral"
      variant="subtle"
      icon="i-lucide-wifi-off"
      title="Status Pelatihan AI Belum Terhubung"
      description="Aplikasi belum menerima respons dari endpoint status pelatihan AI (/api/analytics/retrain/status). Pastikan service analytics-service & python-ai terbaru sudah berjalan."
    >
      <template #actions>
        <UButton
          size="xs"
          color="primary"
          variant="soft"
          icon="i-lucide-refresh-cw"
          :loading="checkingRetrainStatus"
          @click="fetchRetrainStatus()"
        >
          Cek Ulang Status AI
        </UButton>
      </template>
    </UAlert>

    <!-- Loading State -->
    <div
      v-if="loading"
      class="flex justify-center items-center h-96"
    >
      <div class="flex flex-col items-center gap-4">
        <UIcon
          name="i-heroicons-cpu-chip"
          class="w-12 h-12 animate-pulse text-primary-500"
        />
        <span class="text-sm font-semibold text-gray-500 animate-pulse">{{ t('analytics.loading') }}</span>
      </div>
    </div>

    <!-- Content Area -->
    <div
      v-else
      class="space-y-6"
    >
      <!-- Perspective Executive AI Narrative Banner -->
      <div
        v-if="selectedPerspective === 'auditor'"
        class="p-4 rounded-xl bg-gradient-to-r from-primary-50 via-indigo-50 to-purple-50 dark:from-primary-950/30 dark:via-indigo-950/30 dark:to-purple-950/30 border border-primary-200 dark:border-primary-800 shadow-xs"
      >
        <div class="flex items-start gap-3">
          <div class="p-2 rounded-lg bg-primary-500 text-white shrink-0 mt-0.5">
            <UIcon
              name="i-lucide-sparkles"
              class="w-5 h-5"
            />
          </div>
          <div class="space-y-1">
            <h4 class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <span>{{ t('analytics.timeseries.tabAuditor') }} — AI Forecast Insight</span>
              <UBadge
                color="primary"
                variant="subtle"
                size="xs"
              >
                PyTorch LSTM + HIG/HIB
              </UBadge>
            </h4>
            <p class="text-sm text-gray-700 dark:text-gray-300 font-medium leading-relaxed">
              {{ timeseriesState.auditorForecast?.executiveSummary || (auditorForecasts.length === 0 ? 'Belum ada data KPI (dengan kolom Target & Actual) pada fitur Strategic Audit Plan ataupun Internal Audit Performance untuk diproyeksikan.' : `Tersinkronisasi dengan ${auditorForecasts.length} KPI aktif dari Strategic Audit Plan dan Internal Audit Performance.`) }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('analytics.timeseries.tabAuditorDesc') }}
            </p>
          </div>
        </div>
      </div>

      <div
        v-else
        class="p-4 rounded-xl bg-gradient-to-r from-amber-50 via-orange-50 to-indigo-50 dark:from-amber-950/30 dark:via-orange-950/30 dark:to-indigo-950/30 border border-amber-200 dark:border-amber-800 shadow-xs"
      >
        <div class="flex items-start gap-3">
          <div class="p-2 rounded-lg bg-amber-500 text-white shrink-0 mt-0.5">
            <UIcon
              name="i-lucide-building-2"
              class="w-5 h-5"
            />
          </div>
          <div class="space-y-1">
            <h4 class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <span>{{ t('analytics.timeseries.tabAuditee') }} — AI Business Trend Projection</span>
              <UBadge
                color="warning"
                variant="subtle"
                size="xs"
              >
                Gold Zone / CAATT Multi-Source
              </UBadge>
            </h4>
            <p class="text-sm text-gray-700 dark:text-gray-300 font-medium leading-relaxed">
              {{ timeseriesState.auditeeForecast?.executiveSummary || (auditeeForecasts.length === 0 ? 'Belum ada proyeksi indikator bisnis auditee dari Datalake Zona Gold.' : `Menampilkan proyeksi AI LSTM untuk ${auditeeForecasts.length} indikator kesehatan unit bisnis dari Zona Gold / CAATT.`) }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('analytics.timeseries.tabAuditeeDesc') }}
            </p>
          </div>
        </div>
      </div>

      <!-- Forecast Chart & At-Risk Summary Cards -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Main Line Chart -->
        <UCard class="lg:col-span-2">
          <template #header>
            <div class="flex flex-col gap-2.5">
              <div class="flex flex-wrap items-center gap-2">
                <UIcon
                  :name="selectedPerspective === 'auditor' ? 'i-lucide-clipboard-check' : 'i-heroicons-chart-bar'"
                  :class="selectedPerspective === 'auditor' ? 'text-primary-500' : 'text-violet-500'"
                  class="w-5 h-5"
                />
                <h3 class="font-bold text-sm sm:text-base">
                  {{ selectedPerspective === 'auditor' ? t('analytics.timeseries.chartTitleAuditor') : t('analytics.timeseries.chartTitle') }}
                </h3>
                <UBadge
                  color="info"
                  variant="subtle"
                  size="md"
                  class="ml-auto font-medium"
                >
                  {{ t('analytics.timeseries.chartConfidenceBadge') }}
                </UBadge>
              </div>

              <!-- Chart <-> Table Sync Selector -->
              <div class="flex flex-wrap items-center gap-2 pt-1">
                <label class="text-[11px] font-semibold text-gray-500 dark:text-gray-400">
                  Grafik KPI Terpilih:
                </label>
                <select
                  v-if="selectedPerspective === 'auditor'"
                  v-model="selectedAuditorKpiCode"
                  aria-label="Pilih KPI Auditor untuk Grafik"
                  class="text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 py-1 px-2.5 max-w-md focus:ring-1 focus:ring-primary-500"
                >
                  <option :value="AUDITOR_CHART_ALL_KEY">
                    Semua KPI Tabel ({{ filteredAuditorForecasts.length }} KPI — Rata-rata Capaian %)
                  </option>
                  <option
                    v-for="item in filteredAuditorForecasts"
                    :key="item.code"
                    :value="item.code"
                  >
                    {{ item.code }} — {{ item.kpiName }} ({{ item.unit || '%' }})
                  </option>
                </select>
                <select
                  v-else
                  v-model="selectedKpiCode"
                  aria-label="Pilih KPI Auditee untuk Grafik"
                  class="text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 py-1 px-2.5 max-w-md focus:ring-1 focus:ring-indigo-500"
                >
                  <option
                    v-for="item in filteredAuditeeForecasts"
                    :key="item.code"
                    :value="item.code"
                  >
                    {{ item.code }} — {{ item.kpiName }} ({{ item.entity || 'Unit' }})
                  </option>
                </select>
                <UButton
                  v-if="selectedPerspective === 'auditor' && selectedAuditorKpiCode !== AUDITOR_CHART_ALL_KEY"
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  icon="i-lucide-rotate-ccw"
                  @click="selectedAuditorKpiCode = AUDITOR_CHART_ALL_KEY"
                >
                  Semua KPI
                </UButton>
              </div>
            </div>
          </template>
          <UAlert
            v-if="selectedPerspective === 'auditee' && selectedKpi?.isStale"
            color="warning"
            variant="subtle"
            icon="i-lucide-clock-alert"
            :description="t('analytics.timeseries.staleWarning')"
            class="mb-3"
          />
          <div
            v-if="activeChartHasData"
            class="h-80"
          >
            <Line
              :data="activeChartData"
              :options="timeSeriesOptions"
            />
          </div>
          <div
            v-else
            class="h-80 flex flex-col items-center justify-center text-center px-4 text-sm text-gray-400"
          >
            <UIcon
              name="i-lucide-line-chart"
              class="w-10 h-10 mb-2 text-gray-300 dark:text-gray-600"
            />
            <p v-if="selectedPerspective === 'auditor'">
              Belum ada data KPI (dengan kolom Target &amp; Actual) dari fitur Strategic Audit Plan untuk ditampilkan pada grafik.
            </p>
            <p v-else>
              Belum ada data indikator Auditee yang sesuai dengan filter untuk ditampilkan pada grafik.
            </p>
          </div>
          <p
            v-if="selectedPerspective === 'auditee' && selectedKpi?.meta?.key === 'nps'"
            class="text-xs text-gray-400 mt-2"
          >
            {{ t('analytics.kpiAudit.kindHint.nps') }}
          </p>
        </UCard>

        <!-- Right Side: At Risk List Card -->
        <div class="space-y-6">
          <!-- Auditor Perspective At-Risk Pillars -->
          <UCard v-if="selectedPerspective === 'auditor'">
            <template #header>
              <div class="flex items-center justify-between">
                <h3 class="font-bold text-sm flex items-center gap-1.5">
                  <UIcon
                    name="i-lucide-alert-triangle"
                    class="w-4 h-4 text-amber-500"
                  />
                  <span>{{ t('analytics.timeseries.atRiskPillarsTitle') }}</span>
                </h3>
                <UBadge
                  color="warning"
                  variant="subtle"
                  size="xs"
                >
                  {{ auditorAtRiskPillars.length }} Pilar
                </UBadge>
              </div>
            </template>
            <div class="space-y-3 max-h-72 overflow-y-auto pr-1">
              <div
                v-for="item in auditorAtRiskPillars"
                :key="item.code"
                class="p-3 bg-amber-50 dark:bg-amber-950/20 rounded-xl border border-amber-200 dark:border-amber-800/60 cursor-pointer hover:border-amber-400 transition-colors"
                :class="selectedAuditorKpiCode === item.code ? 'ring-2 ring-primary-500' : ''"
                @click="selectedAuditorKpiCode = item.code"
              >
                <div class="flex items-center justify-between gap-2">
                  <div class="font-bold text-sm text-gray-900 dark:text-white">
                    {{ item.code }}
                  </div>
                  <UBadge
                    :class="getStatusBadgeConfig(item.predictedStatus).badgeClass"
                    variant="solid"
                    size="xs"
                    class="font-bold shrink-0"
                  >
                    {{ item.predictedStatus }}
                  </UBadge>
                </div>
                <div class="text-xs text-gray-600 dark:text-gray-300 mt-0.5 line-clamp-1 font-medium">
                  {{ item.kpi || item.kpiName }}
                </div>
                <div class="flex items-center justify-between mt-2 pt-1 border-t border-amber-200/50 dark:border-amber-800/40 text-xs">
                  <span class="text-gray-500">Target: <strong class="text-gray-700 dark:text-gray-200">{{ item.target || item.targetValue }}</strong></span>
                  <span class="font-bold text-amber-600 dark:text-amber-400">Proyeksi: {{ formatNum(item.achievementRate || item.forecastedAchievementRate, 1) }}%</span>
                </div>
              </div>
              <div
                v-if="auditorAtRiskPillars.length === 0"
                class="text-center py-6 text-sm text-gray-400"
              >
                {{ auditorForecasts.length === 0 ? 'Belum ada data KPI Auditor untuk dievaluasi.' : 'Semua indikator audit diproyeksikan berstatus Good.' }}
              </div>
            </div>
          </UCard>

          <!-- Auditee Perspective At-Risk Departments -->
          <UCard v-else>
            <template #header>
              <div class="flex items-center justify-between">
                <h3 class="font-bold text-sm flex items-center gap-1.5">
                  <UIcon
                    name="i-lucide-alert-octagon"
                    class="w-4 h-4 text-rose-500"
                  />
                  <span>{{ t('analytics.timeseries.atRiskTitle') }}</span>
                </h3>
                <UBadge
                  color="error"
                  variant="subtle"
                  size="xs"
                >
                  {{ auditeeAtRisk.length }} Unit
                </UBadge>
              </div>
            </template>
            <div class="space-y-3 max-h-72 overflow-y-auto pr-1">
              <div
                v-for="dept in auditeeAtRisk"
                :key="dept.code || `${dept.department}-${dept.kpi}`"
                class="p-3 bg-rose-50 dark:bg-rose-900/20 rounded-xl border border-rose-200 dark:border-rose-800 cursor-pointer hover:border-rose-400 transition-colors"
                :class="selectedKpi?.kpiName === dept.kpi && selectedKpi?.entity === dept.department ? 'ring-2 ring-indigo-500' : ''"
                @click="focusAuditeeKpi(dept)"
              >
                <div class="flex items-center justify-between gap-2">
                  <div class="font-bold text-sm text-gray-900 dark:text-white">
                    {{ dept.department }}
                  </div>
                  <UBadge
                    :class="getRiskConfig(dept.riskLevel).badgeClass"
                    variant="solid"
                    size="xs"
                    class="font-bold shrink-0"
                  >
                    {{ getRiskConfig(dept.riskLevel).label }}
                  </UBadge>
                </div>
                <div class="text-xs text-gray-500 mt-0.5 line-clamp-1">
                  {{ dept.kpi }}
                </div>
                <div class="flex items-center gap-2 mt-2">
                  <UIcon
                    name="i-heroicons-arrow-trending-down"
                    class="w-4 h-4 text-rose-500"
                  />
                  <span class="text-sm font-bold text-rose-500">{{ formatAuditeeAtRiskVal(dept) }} {{ t('analytics.timeseries.projectedQ3') }}</span>
                </div>
              </div>
              <div
                v-if="auditeeAtRisk.length === 0"
                class="text-center py-6 text-sm text-gray-400"
              >
                {{ t('analytics.timeseries.noAtRisk') }}
              </div>
            </div>
          </UCard>
        </div>
      </div>

      <!-- ─── PERSPECTIVE 1: AUDITOR KPI TABLE (Efektivitas Audit Internal) ─── -->
      <UCard
        v-if="selectedPerspective === 'auditor'"
        class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm ring-0 divide-gray-200 dark:divide-gray-800"
      >
        <template #header>
          <div class="flex flex-col gap-3">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div class="flex items-center gap-2">
                <UIcon
                  name="i-lucide-clipboard-check"
                  class="text-primary-500 w-5 h-5"
                />
                <h3 class="font-bold text-base">
                  {{ t('analytics.timeseries.tableTitleAuditor') }}
                </h3>
                <UBadge
                  color="primary"
                  variant="subtle"
                  size="xs"
                  class="font-mono"
                >
                  Total: {{ filteredAuditorForecasts.length }} KPI
                </UBadge>
              </div>
              <UBadge
                color="primary"
                variant="subtle"
                size="sm"
                class="font-mono"
              >
                Model: PyTorch Auditor LSTM + HIG/HIB
              </UBadge>
            </div>

            <!-- Toolbar: Search & Filter -->
            <div class="flex flex-wrap items-center justify-between gap-3 pt-2 border-t border-gray-100 dark:border-gray-800">
              <div class="flex flex-wrap items-center gap-2 flex-1 min-w-[280px]">
                <UInput
                  v-model="searchQuery"
                  icon="i-lucide-search"
                  placeholder="Cari kode, nama KPI, atau rekomendasi..."
                  size="sm"
                  class="w-full sm:w-72"
                />
                <select
                  v-model="selectedStatusFilter"
                  class="text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 py-1.5 px-2.5 focus:ring-1 focus:ring-primary-500"
                >
                  <option value="ALL">
                    Semua Status (Good / Moderate / Poor)
                  </option>
                  <option value="Good">
                    Good (≥100%)
                  </option>
                  <option value="Moderate">
                    Moderate (70-99%)
                  </option>
                  <option value="Poor">
                    Poor (&lt;70%)
                  </option>
                </select>
                <select
                  v-model="selectedPillarFilter"
                  class="text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 py-1.5 px-2.5 focus:ring-1 focus:ring-primary-500 max-w-[220px]"
                >
                  <option value="ALL">
                    Semua Pilar Audit
                  </option>
                  <option
                    v-for="p in distinctAuditorPillars"
                    :key="p"
                    :value="p"
                  >
                    {{ p }}
                  </option>
                </select>
              </div>

              <div class="flex items-center gap-2">
                <span class="text-xs text-gray-500">Tampilkan:</span>
                <select
                  v-model="pageSize"
                  class="text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 py-1.5 px-2.5"
                >
                  <option
                    v-for="opt in pageSizeOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >
                    {{ opt.label }}
                  </option>
                </select>
              </div>
            </div>
          </div>
        </template>

        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-100 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800">
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colCode') }}
                </th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[200px]">
                  {{ t('analytics.timeseries.colKPI') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colPolarity') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colTarget') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colCurrent') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colForecast') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colForecastAchvRate') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colPredictedStatus') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colTrend') }}
                </th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[260px]">
                  {{ t('analytics.timeseries.colRecommendedAction') }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="kpi in paginatedAuditorForecasts"
                :key="kpi.code"
                class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors cursor-pointer"
                :class="selectedAuditorKpiCode === kpi.code ? 'bg-primary-50/70 dark:bg-primary-950/30 ring-1 ring-inset ring-primary-300 dark:ring-primary-700' : ''"
                @click="selectedAuditorKpiCode = kpi.code"
              >
                <td class="py-3 px-3 font-mono font-bold text-xs text-primary-600 dark:text-primary-400">
                  {{ kpi.code }}
                </td>
                <td class="py-3 px-3">
                  <div class="font-bold text-gray-900 dark:text-white leading-tight">
                    {{ kpi.kpiName }}
                  </div>
                  <div class="text-[11px] text-gray-400 mt-0.5">
                    {{ kpi.category }}
                  </div>
                </td>
                <td class="text-center py-3 px-3">
                  <UBadge
                    :color="kpi.hibHig === 'HIG' ? 'primary' : 'info'"
                    variant="subtle"
                    size="xs"
                    class="font-mono font-bold"
                    :title="kpi.hibHig === 'HIG' ? t('analytics.timeseries.polarityHIG') : t('analytics.timeseries.polarityHIB')"
                  >
                    {{ kpi.hibHig }}
                  </UBadge>
                </td>
                <td class="text-center py-3 px-3 font-mono font-semibold">
                  {{ formatAuditorKpiVal(kpi.targetValue, kpi.unit) }}
                </td>
                <td class="text-center py-3 px-3 font-mono">
                  {{ formatAuditorKpiVal(kpi.currentValue, kpi.unit) }}
                </td>
                <td class="text-center py-3 px-3 font-mono font-bold text-primary-600 dark:text-primary-400">
                  {{ formatAuditorKpiVal(kpi.forecastedValue, kpi.unit) }}
                </td>
                <td class="text-center py-3 px-3 font-mono font-bold">
                  <span :class="kpi.forecastedAchievementRate >= 100 ? 'text-emerald-600 dark:text-emerald-400' : (kpi.forecastedAchievementRate >= 70 ? 'text-amber-600 dark:text-amber-400' : 'text-rose-600 dark:text-rose-400')">
                    {{ formatNum(kpi.forecastedAchievementRate, 1) }}%
                  </span>
                </td>
                <td class="text-center py-3 px-3">
                  <UBadge
                    :class="getStatusBadgeConfig(kpi.predictedStatus).badgeClass"
                    variant="solid"
                    size="xs"
                    class="font-bold"
                  >
                    {{ kpi.predictedStatus }}
                  </UBadge>
                </td>
                <td class="text-center py-3 px-3">
                  <div class="flex items-center justify-center gap-1">
                    <UIcon
                      :name="trendIcon(kpi.trend)"
                      class="w-4 h-4"
                      :class="trendColor(kpi.trend)"
                    />
                    <span
                      class="text-xs font-bold"
                      :class="trendColor(kpi.trend)"
                    >{{ kpi.trend }}</span>
                  </div>
                </td>
                <td class="py-3 px-3 text-xs text-gray-600 dark:text-gray-300 leading-relaxed">
                  <div class="font-semibold text-gray-800 dark:text-gray-200">
                    {{ kpi.aiNarrative }}
                  </div>
                  <div class="text-[11px] text-gray-500 mt-0.5">
                    {{ kpi.recommendedAction }}
                  </div>
                </td>
              </tr>
              <tr v-if="paginatedAuditorForecasts.length === 0">
                <td
                  colspan="10"
                  class="py-8 text-center text-gray-400"
                >
                  Tidak ada data KPI Auditor yang sesuai dengan filter pencarian.
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <template #footer>
          <div class="flex flex-wrap items-center justify-between gap-3 text-xs text-gray-500 dark:text-gray-400">
            <div>
              Menampilkan
              <span class="font-bold text-gray-900 dark:text-white">
                {{ filteredAuditorForecasts.length > 0 ? (currentPage - 1) * pageSize + 1 : 0 }}
              </span>
              -
              <span class="font-bold text-gray-900 dark:text-white">
                {{ Math.min(currentPage * pageSize, filteredAuditorForecasts.length) }}
              </span>
              dari total
              <span class="font-bold text-gray-900 dark:text-white">{{ filteredAuditorForecasts.length }}</span>
              KPI Auditor
            </div>

            <div
              v-if="totalAuditorPages > 1"
              class="flex items-center gap-1"
            >
              <UButton
                icon="i-lucide-chevron-left"
                size="xs"
                color="neutral"
                variant="ghost"
                :disabled="currentPage <= 1"
                @click="currentPage--"
              />
              <span class="px-2 font-mono font-medium">Halaman {{ currentPage }} / {{ totalAuditorPages }}</span>
              <UButton
                icon="i-lucide-chevron-right"
                size="xs"
                color="neutral"
                variant="ghost"
                :disabled="currentPage >= totalAuditorPages"
                @click="currentPage++"
              />
            </div>
          </div>
        </template>
      </UCard>

      <!-- ─── PERSPECTIVE 2: AUDITEE KPI TABLE (Kesehatan Unit Bisnis - Gold Zone / CAATT) ─── -->
      <UCard
        v-else
        class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm ring-0 divide-gray-200 dark:divide-gray-800"
      >
        <template #header>
          <div class="flex flex-col gap-3">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div class="flex items-center gap-2">
                <UIcon
                  name="i-lucide-building-2"
                  class="text-indigo-500 w-5 h-5"
                />
                <h3 class="font-bold text-base">
                  {{ t('analytics.timeseries.tableTitleAuditee') }}
                </h3>
                <UBadge
                  color="info"
                  variant="subtle"
                  size="xs"
                  class="font-mono"
                >
                  Total: {{ filteredAuditeeForecasts.length }} Indikator
                </UBadge>
              </div>
              <UBadge
                color="info"
                variant="subtle"
                size="sm"
                class="font-mono"
              >
                Sumber: Datalake Zona Gold (Multi-Source CDC)
              </UBadge>
            </div>

            <!-- Toolbar: Search & Filter -->
            <div class="flex flex-wrap items-center justify-between gap-3 pt-2 border-t border-gray-100 dark:border-gray-800">
              <div class="flex flex-wrap items-center gap-2 flex-1 min-w-[280px]">
                <UInput
                  v-model="searchQuery"
                  icon="i-lucide-search"
                  placeholder="Cari indikator bisnis, cabang, atau rekomendasi..."
                  size="sm"
                  class="w-full sm:w-72"
                />
                <select
                  v-model="selectedRiskFilter"
                  class="text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 py-1.5 px-2.5 focus:ring-1 focus:ring-indigo-500"
                >
                  <option value="ALL">
                    Semua Risk Level
                  </option>
                  <option value="HIGH">
                    High
                  </option>
                  <option value="MODERATE_HIGH">
                    Moderate to High
                  </option>
                  <option value="MODERATE">
                    Moderate
                  </option>
                  <option value="LOW_MODERATE">
                    Low to Moderate
                  </option>
                  <option value="LOW">
                    Low
                  </option>
                </select>
              </div>

              <div class="flex items-center gap-2">
                <span class="text-xs text-gray-500">Tampilkan:</span>
                <select
                  v-model="pageSize"
                  class="text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 py-1.5 px-2.5"
                >
                  <option
                    v-for="opt in pageSizeOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >
                    {{ opt.label }}
                  </option>
                </select>
              </div>
            </div>
          </div>
        </template>

        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-100 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800">
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colCode') }}
                </th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[180px]">
                  {{ t('analytics.timeseries.colKPI') }}
                </th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colEntity') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colSyncFreq') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colForecastHorizon') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colCurrent') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colForecast') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colTrend') }}
                </th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">
                  {{ t('analytics.timeseries.colRiskLevel') }}
                </th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[260px]">
                  {{ t('analytics.timeseries.colRecommendedAction') }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="kpi in paginatedAuditeeForecasts"
                :key="kpi.code"
                class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors cursor-pointer"
                :class="selectedKpiCode === kpi.code ? 'bg-indigo-50/70 dark:bg-indigo-950/30 ring-1 ring-inset ring-indigo-300 dark:ring-indigo-700' : ''"
                @click="focusAuditeeKpi(kpi)"
              >
                <td class="py-3 px-3 font-mono font-bold text-xs text-indigo-600 dark:text-indigo-400">
                  {{ kpi.code }}
                </td>
                <td class="py-3 px-3 font-bold text-gray-900 dark:text-white">
                  {{ kpi.kpiName }}
                </td>
                <td class="py-3 px-3">
                  <div class="font-semibold text-gray-800 dark:text-gray-200">
                    {{ kpi.entity || '-' }}
                  </div>
                  <div class="text-[11px] text-gray-400">
                    {{ kpi.entityType || '-' }}
                  </div>
                </td>
                <td class="text-center py-3 px-3">
                  <UBadge
                    :color="kpi.syncFrequency === 'Real-time (CDC)' ? 'primary' : 'neutral'"
                    variant="subtle"
                    size="xs"
                    class="font-mono"
                  >
                    {{ kpi.syncFrequency || 'Real-time (CDC)' }}
                  </UBadge>
                </td>
                <td class="text-center py-3 px-3">
                  <UBadge
                    color="neutral"
                    variant="subtle"
                    size="xs"
                    class="font-mono font-medium"
                  >
                    {{ kpi.targetHorizon || 'Q3 2026' }}
                  </UBadge>
                </td>
                <td class="text-center py-3 px-3 font-mono">
                  {{ formatAuditeeKpiVal(kpi.currentValue, kpi.kpiName, kpi.unit) }}
                </td>
                <td class="text-center py-3 px-3 font-mono font-bold text-indigo-600 dark:text-indigo-400">
                  {{ formatAuditeeKpiVal(kpi.forecastedValue, kpi.kpiName, kpi.unit) }}
                </td>
                <td class="text-center py-3 px-3">
                  <div class="flex items-center justify-center gap-1">
                    <UIcon
                      :name="trendIcon(kpi.trend)"
                      class="w-4 h-4"
                      :class="trendColor(kpi.trend)"
                    />
                    <span
                      class="text-xs font-bold"
                      :class="trendColor(kpi.trend)"
                    >{{ kpi.trend }}</span>
                  </div>
                </td>
                <td class="text-center py-3 px-3">
                  <UBadge
                    :class="getRiskConfig(kpi.riskLevel).badgeClass"
                    variant="solid"
                    size="xs"
                    class="font-bold"
                  >
                    {{ getRiskConfig(kpi.riskLevel).label }}
                  </UBadge>
                </td>
                <td class="py-3 px-3 text-xs text-gray-600 dark:text-gray-300 leading-relaxed">
                  <div class="font-semibold text-gray-800 dark:text-gray-200">
                    {{ kpi.aiNarrative }}
                  </div>
                  <div class="text-[11px] text-gray-500 mt-0.5">
                    {{ kpi.recommendedAction }}
                  </div>
                </td>
              </tr>
              <tr v-if="paginatedAuditeeForecasts.length === 0">
                <td
                  colspan="10"
                  class="py-8 text-center text-gray-400"
                >
                  Tidak ada data Indikator Auditee yang sesuai dengan filter pencarian.
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <template #footer>
          <div class="flex flex-wrap items-center justify-between gap-3 text-xs text-gray-500 dark:text-gray-400">
            <div>
              Menampilkan
              <span class="font-bold text-gray-900 dark:text-white">
                {{ filteredAuditeeForecasts.length > 0 ? (currentPage - 1) * pageSize + 1 : 0 }}
              </span>
              -
              <span class="font-bold text-gray-900 dark:text-white">
                {{ Math.min(currentPage * pageSize, filteredAuditeeForecasts.length) }}
              </span>
              dari total
              <span class="font-bold text-gray-900 dark:text-white">{{ filteredAuditeeForecasts.length }}</span>
              Indikator Auditee
            </div>

            <div
              v-if="totalAuditeePages > 1"
              class="flex items-center gap-1"
            >
              <UButton
                icon="i-lucide-chevron-left"
                size="xs"
                color="neutral"
                variant="ghost"
                :disabled="currentPage <= 1"
                @click="currentPage--"
              />
              <span class="px-2 font-mono font-medium">Halaman {{ currentPage }} / {{ totalAuditeePages }}</span>
              <UButton
                icon="i-lucide-chevron-right"
                size="xs"
                color="neutral"
                variant="ghost"
                :disabled="currentPage >= totalAuditeePages"
                @click="currentPage++"
              />
            </div>
          </div>
        </template>
      </UCard>
    </div>
  </div>
</template>
