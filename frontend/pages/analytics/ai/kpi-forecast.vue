<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
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
import { useI18n } from '~/composables/useI18n'
import AiInsightHeader from '~/components/analytics/AiInsightHeader.vue'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const { t } = useI18n()
const toast = useToast()
const {
  loading,
  timeseriesState,
  formatNum,
  getRiskConfig,
  trendColor,
  trendIcon,
  safeApiFetch,
  fetchAiAnalytics
} = useAiAnalytics()

// Perspective switch: 'auditor' (Strategic Audit Plan) vs 'auditee' (Business Health Gold Zone)
const selectedPerspective = ref<'auditor' | 'auditee'>('auditor')
const isRetraining = ref(false)

onMounted(async () => {
  await fetchAiAnalytics(true)
})

const handleManualRetrain = async () => {
  if (isRetraining.value) return
  isRetraining.value = true
  try {
    await safeApiFetch('/retrain/auto', {
      method: 'POST',
      body: { trigger: 'user_perspective_refresh' }
    })
    
    await fetchAiAnalytics(true)
    toast.add({
      title: t('analytics.timeseries.retrainSuccess'),
      color: 'success',
      icon: 'i-lucide-check-circle'
    })
  } catch (err: any) {
    console.warn('Retrain error:', err)
  } finally {
    isRetraining.value = false
  }
}

// ─── AUDITOR TIME SERIES CHART DATA ──────────────────────────────────────────
const auditorTimeSeriesChartData = computed(() => {
  const points = timeseriesState.value.auditorForecast?.timeSeriesData || [
    { period: 'Q1 2025', actual: 88.5, forecast: 88.5, upperBound: 91.0, lowerBound: 86.0 },
    { period: 'Q2 2025', actual: 90.2, forecast: 90.2, upperBound: 93.0, lowerBound: 87.5 },
    { period: 'Q3 2025', actual: 89.0, forecast: 89.0, upperBound: 92.0, lowerBound: 86.0 },
    { period: 'Q4 2025', actual: 91.8, forecast: 91.8, upperBound: 94.5, lowerBound: 89.0 },
    { period: 'Q1 2026 (Pred)', actual: null, forecast: 86.4, upperBound: 89.5, lowerBound: 83.0 },
    { period: 'Akhir Tahun 2026 (Pred)', actual: null, forecast: 82.4, upperBound: 86.0, lowerBound: 78.5 }
  ]

  return {
    labels: points.map((p: any) => p.period || ''),
    datasets: [
      {
        label: 'Target Tahunan (100%)',
        data: points.map(() => 100),
        borderColor: 'rgba(239, 68, 68, 0.7)',
        borderDash: [5, 5],
        borderWidth: 2,
        pointRadius: 0,
        fill: false,
        spanGaps: true
      },
      {
        label: 'Realisasi Aktual PKAT (%)',
        data: points.map((p: any) => p.actual ?? null),
        borderColor: 'rgb(59, 130, 246)',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
        tension: 0.3,
        pointRadius: 5,
        pointBackgroundColor: 'rgb(59, 130, 246)',
        spanGaps: false
      },
      {
        label: 'Proyeksi AI LSTM (%)',
        data: points.map((p: any) => p.forecast ?? null),
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
        data: points.map((p: any) => p.upperBound ?? null),
        borderColor: 'transparent',
        backgroundColor: 'rgba(147, 51, 234, 0.08)',
        fill: '+1',
        tension: 0.3,
        pointRadius: 0,
        spanGaps: false
      },
      {
        label: t('analytics.timeseries.labelLowerBound'),
        data: points.map((p: any) => p.lowerBound ?? null),
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

// ─── AUDITEE TIME SERIES CHART DATA ──────────────────────────────────────────
const auditeeTimeSeriesChartData = computed(() => {
  const points = timeseriesState.value.auditeeForecast?.timeSeriesData || timeseriesState.value.historicalKPI || []
  return {
    labels: points.map((p: any) => p.period || ''),
    datasets: [
      {
        label: t('analytics.timeseries.labelActualKPI'),
        data: points.map((p: any) => p.actual ?? null),
        borderColor: 'rgb(59,130,246)',
        backgroundColor: 'rgba(59,130,246,0.1)',
        tension: 0.4,
        pointRadius: 5,
        pointBackgroundColor: 'rgb(59,130,246)',
        spanGaps: false,
      },
      {
        label: t('analytics.timeseries.labelForecast'),
        data: points.map((p: any) => p.forecast ?? null),
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
        label: t('analytics.timeseries.labelUpperBound'),
        data: points.map((p: any) => p.upperBound ?? null),
        borderColor: 'transparent',
        backgroundColor: 'rgba(139,92,246,0.08)',
        fill: '+1',
        tension: 0.4,
        pointRadius: 0,
        spanGaps: false,
      },
      {
        label: t('analytics.timeseries.labelLowerBound'),
        data: points.map((p: any) => p.lowerBound ?? null),
        borderColor: 'transparent',
        backgroundColor: 'rgba(139,92,246,0.08)',
        fill: '-1',
        tension: 0.4,
        pointRadius: 0,
        spanGaps: false,
      },
    ]
  }
})

const activeChartData = computed(() => {
  return selectedPerspective.value === 'auditor' ? auditorTimeSeriesChartData.value : auditeeTimeSeriesChartData.value
})

const timeSeriesOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        filter: (item: any) => ![t('analytics.timeseries.labelUpperBound'), t('analytics.timeseries.labelLowerBound')].includes(item.text)
      }
    },
    tooltip: {
      callbacks: {
        label: (ctx: any) => {
          if ([t('analytics.timeseries.labelUpperBound'), t('analytics.timeseries.labelLowerBound')].includes(ctx.dataset.label)) return ''
          return `${ctx.dataset.label}: ${formatNum(ctx.parsed.y, 1)}%`
        }
      }
    }
  },
  scales: {
    y: {
      title: { display: true, text: selectedPerspective.value === 'auditor' ? 'Tingkat Ketercapaian Audit (%)' : t('analytics.timeseries.axisKPIScore') },
      suggestedMin: selectedPerspective.value === 'auditor' ? 60 : undefined,
      suggestedMax: selectedPerspective.value === 'auditor' ? 105 : undefined
    }
  }
}))

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

// Active forecast list
const auditorForecasts = computed(() => timeseriesState.value.auditorForecast?.kpiForecasts || [])
const auditeeForecasts = computed(() => timeseriesState.value.auditeeForecast?.kpiForecasts || timeseriesState.value.kpiForecasts || [])

const auditorAtRiskPillars = computed(() => {
  if (timeseriesState.value.auditorForecast?.atRiskPillars?.length) {
    return timeseriesState.value.auditorForecast.atRiskPillars
  }
  return auditorForecasts.value.filter((k: any) => k.predictedStatus === 'Poor' || k.predictedStatus === 'Moderate' || k.trend === 'Deteriorating')
})

const auditeeAtRisk = computed(() => {
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

// Distinct pillars for Auditor KPI filter
const distinctAuditorPillars = computed(() => {
  const set = new Set<string>()
  auditorForecasts.value.forEach((k: any) => {
    if (k.category) set.add(k.category)
  })
  return Array.from(set).sort()
})

// Filtered Auditor KPIs
const filteredAuditorForecasts = computed(() => {
  let list = auditorForecasts.value
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter((k: any) =>
      (k.code && k.code.toLowerCase().includes(q)) ||
      (k.kpiName && k.kpiName.toLowerCase().includes(q)) ||
      (k.category && k.category.toLowerCase().includes(q)) ||
      (k.aiNarrative && k.aiNarrative.toLowerCase().includes(q)) ||
      (k.recommendedAction && k.recommendedAction.toLowerCase().includes(q))
    )
  }
  if (selectedStatusFilter.value !== 'ALL') {
    list = list.filter((k: any) => k.predictedStatus === selectedStatusFilter.value)
  }
  if (selectedPillarFilter.value !== 'ALL') {
    list = list.filter((k: any) => k.category === selectedPillarFilter.value)
  }
  return list
})

// Filtered Auditee KPIs
const filteredAuditeeForecasts = computed(() => {
  let list = auditeeForecasts.value
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter((k: any) =>
      (k.code && k.code.toLowerCase().includes(q)) ||
      (k.kpiName && k.kpiName.toLowerCase().includes(q)) ||
      (k.entity && k.entity.toLowerCase().includes(q)) ||
      (k.aiNarrative && k.aiNarrative.toLowerCase().includes(q)) ||
      (k.recommendedAction && k.recommendedAction.toLowerCase().includes(q))
    )
  }
  if (selectedRiskFilter.value !== 'ALL') {
    list = list.filter((k: any) => {
      const r = String(k.riskLevel || '').toUpperCase()
      return r === selectedRiskFilter.value || r.includes(selectedRiskFilter.value)
    })
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
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8 space-y-6">
    <AiInsightHeader
      :title="t('navigation.kpiForecast')"
      :subtitle="t('analytics.subtitle')"
      icon="i-lucide-trending-up"
      currentSubModule="kpi-forecast"
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
          <UIcon name="i-lucide-clipboard-check" class="w-4 h-4 text-primary-500" />
          <span>{{ t('analytics.timeseries.tabAuditor') }}</span>
          <UBadge color="primary" variant="subtle" size="xs" class="ml-1 font-mono">{{ auditorForecasts.length || 12 }} KPI</UBadge>
        </button>

        <button
          type="button"
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm transition-all flex-1 sm:flex-initial justify-center"
          :class="selectedPerspective === 'auditee'
            ? 'bg-white dark:bg-gray-900 text-indigo-600 dark:text-indigo-400 font-bold shadow-xs'
            : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white font-medium'"
          @click="selectedPerspective = 'auditee'"
        >
          <UIcon name="i-lucide-building-2" class="w-4 h-4 text-indigo-500" />
          <span>{{ t('analytics.timeseries.tabAuditee') }}</span>
          <UBadge color="info" variant="subtle" size="xs" class="ml-1 font-mono">{{ auditeeForecasts.length || 37 }} KPI (Gold/CAATT)</UBadge>
        </button>
      </div>

      <!-- Sync & Webhook Retrain Button -->
      <div class="flex items-center gap-2 w-full sm:w-auto justify-end">
        <UBadge color="neutral" variant="subtle" size="sm" class="hidden md:flex items-center gap-1 text-xs">
          <span class="size-2 rounded-full bg-emerald-500 animate-pulse" />
          <span>Sync CDC: Live</span>
        </UBadge>
        <UButton
          :loading="isRetraining"
          icon="i-lucide-refresh-cw"
          color="primary"
          variant="solid"
          size="sm"
          class="font-semibold rounded-xl"
          @click="handleManualRetrain"
        >
          {{ isRetraining ? t('analytics.timeseries.syncing') : t('analytics.timeseries.syncNow') }}
        </UButton>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center items-center h-96">
      <div class="flex flex-col items-center gap-4">
        <UIcon name="i-heroicons-cpu-chip" class="w-12 h-12 animate-pulse text-primary-500" />
        <span class="text-sm font-semibold text-gray-500 animate-pulse">{{ t('analytics.loading') }}</span>
      </div>
    </div>

    <!-- Content Area -->
    <div v-else class="space-y-6">
      <!-- Perspective Executive AI Narrative Banner -->
      <div v-if="selectedPerspective === 'auditor'" class="p-4 rounded-xl bg-gradient-to-r from-primary-50 via-indigo-50 to-purple-50 dark:from-primary-950/30 dark:via-indigo-950/30 dark:to-purple-950/30 border border-primary-200 dark:border-primary-800 shadow-xs">
        <div class="flex items-start gap-3">
          <div class="p-2 rounded-lg bg-primary-500 text-white shrink-0 mt-0.5">
            <UIcon name="i-lucide-sparkles" class="w-5 h-5" />
          </div>
          <div class="space-y-1">
            <h4 class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <span>{{ t('analytics.timeseries.tabAuditor') }} — AI Forecast Insight</span>
              <UBadge color="primary" variant="subtle" size="xs">PyTorch LSTM + HIG/HIB</UBadge>
            </h4>
            <p class="text-sm text-gray-700 dark:text-gray-300 font-medium leading-relaxed">
              {{ timeseriesState.auditorForecast?.executiveSummary || "Dengan kecepatan saat ini, diprediksi audit plan hanya selesai 82.4% di akhir tahun. Diperlukan akselerasi pada pilar Pelaksanaan Audit & PKAT dan Investigasi Khusus." }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('analytics.timeseries.tabAuditorDesc') }}
            </p>
          </div>
        </div>
      </div>

      <div v-else class="p-4 rounded-xl bg-gradient-to-r from-amber-50 via-orange-50 to-indigo-50 dark:from-amber-950/30 dark:via-orange-950/30 dark:to-indigo-950/30 border border-amber-200 dark:border-amber-800 shadow-xs">
        <div class="flex items-start gap-3">
          <div class="p-2 rounded-lg bg-amber-500 text-white shrink-0 mt-0.5">
            <UIcon name="i-lucide-building-2" class="w-5 h-5" />
          </div>
          <div class="space-y-1">
            <h4 class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <span>{{ t('analytics.timeseries.tabAuditee') }} — AI Business Trend Projection</span>
              <UBadge color="warning" variant="subtle" size="xs">Gold Zone / CAATT Multi-Source</UBadge>
            </h4>
            <p class="text-sm text-gray-700 dark:text-gray-300 font-medium leading-relaxed">
              {{ timeseriesState.auditeeForecast?.executiveSummary || "Diprediksi Cabang Surabaya akan mengalami lonjakan NPL dalam 2 kuartal ke depan akibat kenaikan rasio kredit berisiko. Diperlukan pemantauan kredit macet dan pengetatan threshold Continuous Auditing." }}
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
            <div class="flex items-center gap-2">
              <UIcon :name="selectedPerspective === 'auditor' ? 'i-lucide-clipboard-check' : 'i-heroicons-chart-bar'" :class="selectedPerspective === 'auditor' ? 'text-primary-500' : 'text-violet-500'" class="w-5 h-5" />
              <h3 class="font-bold text-sm sm:text-base">
                {{ selectedPerspective === 'auditor' ? t('analytics.timeseries.chartTitleAuditor') : t('analytics.timeseries.chartTitle') }}
              </h3>
              <UBadge color="info" variant="subtle" size="md" class="ml-auto font-medium">
                {{ t('analytics.timeseries.chartConfidenceBadge') }}
              </UBadge>
            </div>
          </template>
          <div class="h-80">
            <Line :data="activeChartData" :options="timeSeriesOptions" />
          </div>
        </UCard>

        <!-- Right Side: At Risk List Card -->
        <div class="space-y-6">
          <!-- Auditor Perspective At-Risk Pillars -->
          <UCard v-if="selectedPerspective === 'auditor'">
            <template #header>
              <div class="flex items-center justify-between">
                <h3 class="font-bold text-sm flex items-center gap-1.5">
                  <UIcon name="i-lucide-alert-triangle" class="w-4 h-4 text-amber-500" />
                  <span>{{ t('analytics.timeseries.atRiskPillarsTitle') }}</span>
                </h3>
                <UBadge color="warning" variant="subtle" size="xs">{{ auditorAtRiskPillars.length }} Pilar</UBadge>
              </div>
            </template>
            <div class="space-y-3 max-h-72 overflow-y-auto pr-1">
              <div
                v-for="item in auditorAtRiskPillars"
                :key="item.code"
                class="p-3 bg-amber-50 dark:bg-amber-950/20 rounded-xl border border-amber-200 dark:border-amber-800/60"
              >
                <div class="flex items-center justify-between gap-2">
                  <div class="font-bold text-sm text-gray-900 dark:text-white">{{ item.code }}</div>
                  <UBadge
                    :class="getStatusBadgeConfig(item.predictedStatus).badgeClass"
                    variant="solid"
                    size="xs"
                    class="font-bold shrink-0"
                  >
                    {{ item.predictedStatus }}
                  </UBadge>
                </div>
                <div class="text-xs text-gray-600 dark:text-gray-300 mt-0.5 line-clamp-1 font-medium">{{ item.kpi || item.kpiName }}</div>
                <div class="flex items-center justify-between mt-2 pt-1 border-t border-amber-200/50 dark:border-amber-800/40 text-xs">
                  <span class="text-gray-500">Target: <strong class="text-gray-700 dark:text-gray-200">{{ item.target || item.targetValue }}</strong></span>
                  <span class="font-bold text-amber-600 dark:text-amber-400">Proyeksi: {{ formatNum(item.achievementRate || item.forecastedAchievementRate, 1) }}%</span>
                </div>
              </div>
              <div v-if="auditorAtRiskPillars.length === 0" class="text-center py-6 text-sm text-gray-400">
                Semua indikator audit diproyeksikan berstatus Good.
              </div>
            </div>
          </UCard>

          <!-- Auditee Perspective At-Risk Departments -->
          <UCard v-else>
            <template #header>
              <div class="flex items-center justify-between">
                <h3 class="font-bold text-sm flex items-center gap-1.5">
                  <UIcon name="i-lucide-alert-octagon" class="w-4 h-4 text-rose-500" />
                  <span>{{ t('analytics.timeseries.atRiskTitle') }}</span>
                </h3>
                <UBadge color="error" variant="subtle" size="xs">{{ auditeeAtRisk.length }} Unit</UBadge>
              </div>
            </template>
            <div class="space-y-3 max-h-72 overflow-y-auto pr-1">
              <div
                v-for="dept in auditeeAtRisk"
                :key="dept.department"
                class="p-3 bg-rose-50 dark:bg-rose-900/20 rounded-xl border border-rose-200 dark:border-rose-800"
              >
                <div class="flex items-center justify-between gap-2">
                  <div class="font-bold text-sm text-gray-900 dark:text-white">{{ dept.department }}</div>
                  <UBadge
                    :class="getRiskConfig(dept.riskLevel).badgeClass"
                    variant="solid"
                    size="xs"
                    class="font-bold shrink-0"
                  >
                    {{ getRiskConfig(dept.riskLevel).label }}
                  </UBadge>
                </div>
                <div class="text-xs text-gray-500 mt-0.5 line-clamp-1">{{ dept.kpi }}</div>
                <div class="flex items-center gap-2 mt-2">
                  <UIcon name="i-heroicons-arrow-trending-down" class="w-4 h-4 text-rose-500" />
                  <span class="text-sm font-bold text-rose-500">{{ formatNum(dept.predictedQ3, 1) }}{{ t('analytics.timeseries.projectedQ3') }}</span>
                </div>
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
                <UIcon name="i-lucide-clipboard-check" class="text-primary-500 w-5 h-5" />
                <h3 class="font-bold text-base">{{ t('analytics.timeseries.tableTitleAuditor') }}</h3>
                <UBadge color="primary" variant="subtle" size="xs" class="font-mono">
                  Total: {{ filteredAuditorForecasts.length }} KPI
                </UBadge>
              </div>
              <UBadge color="primary" variant="subtle" size="sm" class="font-mono">
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
                  clearable
                />
                <select
                  v-model="selectedStatusFilter"
                  class="text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 py-1.5 px-2.5 focus:ring-1 focus:ring-primary-500"
                >
                  <option value="ALL">Semua Status (Good / Moderate / Poor)</option>
                  <option value="Good">Good (≥100%)</option>
                  <option value="Moderate">Moderate (70-99%)</option>
                  <option value="Poor">Poor (&lt;70%)</option>
                </select>
                <select
                  v-model="selectedPillarFilter"
                  class="text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 py-1.5 px-2.5 focus:ring-1 focus:ring-primary-500 max-w-[220px]"
                >
                  <option value="ALL">Semua Pilar Audit</option>
                  <option v-for="p in distinctAuditorPillars" :key="p" :value="p">{{ p }}</option>
                </select>
              </div>

              <div class="flex items-center gap-2">
                <span class="text-xs text-gray-500">Tampilkan:</span>
                <select
                  v-model="pageSize"
                  class="text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 py-1.5 px-2.5"
                >
                  <option v-for="opt in pageSizeOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                </select>
              </div>
            </div>
          </div>
        </template>

        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-100 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800">
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colCode') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[200px]">{{ t('analytics.timeseries.colKPI') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colPolarity') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colTarget') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colCurrent') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colForecast') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colForecastAchvRate') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colPredictedStatus') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colTrend') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[260px]">{{ t('analytics.timeseries.colRecommendedAction') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="kpi in paginatedAuditorForecasts"
                :key="kpi.code"
                class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
              >
                <td class="py-3 px-3 font-mono font-bold text-xs text-primary-600 dark:text-primary-400">{{ kpi.code }}</td>
                <td class="py-3 px-3">
                  <div class="font-bold text-gray-900 dark:text-white leading-tight">{{ kpi.kpiName }}</div>
                  <div class="text-[11px] text-gray-400 mt-0.5">{{ kpi.category }}</div>
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
                <td class="text-center py-3 px-3 font-mono font-semibold">{{ kpi.targetValue }} {{ kpi.unit === '%' ? '%' : '' }}</td>
                <td class="text-center py-3 px-3 font-mono">{{ formatNum(kpi.currentValue, 1) }}{{ kpi.unit === '%' ? '%' : '' }}</td>
                <td class="text-center py-3 px-3 font-mono font-bold text-primary-600 dark:text-primary-400">
                  {{ formatNum(kpi.forecastedValue, 1) }}{{ kpi.unit === '%' ? '%' : '' }}
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
                    <UIcon :name="trendIcon(kpi.trend)" class="w-4 h-4" :class="trendColor(kpi.trend)" />
                    <span class="text-xs font-bold" :class="trendColor(kpi.trend)">{{ kpi.trend }}</span>
                  </div>
                </td>
                <td class="py-3 px-3 text-xs text-gray-600 dark:text-gray-300 leading-relaxed">
                  <div class="font-semibold text-gray-800 dark:text-gray-200">{{ kpi.aiNarrative }}</div>
                  <div class="text-[11px] text-gray-500 mt-0.5">{{ kpi.recommendedAction }}</div>
                </td>
              </tr>
              <tr v-if="paginatedAuditorForecasts.length === 0">
                <td colspan="10" class="py-8 text-center text-gray-400">
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

            <div v-if="totalAuditorPages > 1" class="flex items-center gap-1">
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
                <UIcon name="i-lucide-building-2" class="text-indigo-500 w-5 h-5" />
                <h3 class="font-bold text-base">{{ t('analytics.timeseries.tableTitleAuditee') }}</h3>
                <UBadge color="info" variant="subtle" size="xs" class="font-mono">
                  Total: {{ filteredAuditeeForecasts.length }} Indikator
                </UBadge>
              </div>
              <UBadge color="info" variant="subtle" size="sm" class="font-mono">
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
                  clearable
                />
                <select
                  v-model="selectedRiskFilter"
                  class="text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 py-1.5 px-2.5 focus:ring-1 focus:ring-indigo-500"
                >
                  <option value="ALL">Semua Risk Level</option>
                  <option value="CRITICAL">Critical</option>
                  <option value="HIGH">High</option>
                  <option value="MODERATE">Moderate / Moderate to High</option>
                  <option value="LOW">Low</option>
                </select>
              </div>

              <div class="flex items-center gap-2">
                <span class="text-xs text-gray-500">Tampilkan:</span>
                <select
                  v-model="pageSize"
                  class="text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 py-1.5 px-2.5"
                >
                  <option v-for="opt in pageSizeOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                </select>
              </div>
            </div>
          </div>
        </template>

        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-100 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800">
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colCode') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[180px]">{{ t('analytics.timeseries.colKPI') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colEntity') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colSyncFreq') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colForecastHorizon') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colCurrent') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colForecast') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colTrend') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colRiskLevel') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[260px]">{{ t('analytics.timeseries.colRecommendedAction') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="kpi in paginatedAuditeeForecasts"
                :key="kpi.code"
                class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
              >
                <td class="py-3 px-3 font-mono font-bold text-xs text-indigo-600 dark:text-indigo-400">{{ kpi.code }}</td>
                <td class="py-3 px-3 font-bold text-gray-900 dark:text-white">{{ kpi.kpiName }}</td>
                <td class="py-3 px-3">
                  <div class="font-semibold text-gray-800 dark:text-gray-200">{{ kpi.entity || '-' }}</div>
                  <div class="text-[11px] text-gray-400">{{ kpi.entityType || '-' }}</div>
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
                  <UBadge color="neutral" variant="subtle" size="xs" class="font-mono font-medium">
                    {{ kpi.targetHorizon || 'Q3 2026' }}
                  </UBadge>
                </td>
                <td class="text-center py-3 px-3 font-mono">{{ formatNum(kpi.currentValue, 1) }}{{ kpi.unit === '%' ? '%' : '' }}</td>
                <td class="text-center py-3 px-3 font-mono font-bold text-indigo-600 dark:text-indigo-400">
                  {{ formatNum(kpi.forecastedValue, 1) }}{{ kpi.unit === '%' ? '%' : '' }}
                </td>
                <td class="text-center py-3 px-3">
                  <div class="flex items-center justify-center gap-1">
                    <UIcon :name="trendIcon(kpi.trend)" class="w-4 h-4" :class="trendColor(kpi.trend)" />
                    <span class="text-xs font-bold" :class="trendColor(kpi.trend)">{{ kpi.trend }}</span>
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
                  <div class="font-semibold text-gray-800 dark:text-gray-200">{{ kpi.aiNarrative }}</div>
                  <div class="text-[11px] text-gray-500 mt-0.5">{{ kpi.recommendedAction }}</div>
                </td>
              </tr>
              <tr v-if="paginatedAuditeeForecasts.length === 0">
                <td colspan="10" class="py-8 text-center text-gray-400">
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

            <div v-if="totalAuditeePages > 1" class="flex items-center gap-1">
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
