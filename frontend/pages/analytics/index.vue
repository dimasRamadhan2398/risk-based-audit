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
} from '~/composables/useAnalyticsData'
import type {
  RiskScorePrediction,
  AnomalyRecord,
  NLPDocumentResult,
  KPIForecast,
} from '~/composables/useAnalyticsData'
import { useRiskProfileStore, riskLevelConfig } from '~/stores/risk-profile'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement, ArcElement, Title, Tooltip, Legend, Filler)

// ─── Data ───────────────────────────────────────────────────────────────────
const xgboost = useXGBoostData()
const isolation = useIsolationForestData()
const nlp = useIndoBERTData()
const timeseries = useTimeSeriesData()
const summary = useAnalyticsSummary()

const loading = ref(true)

onMounted(() => {
  // Simulate loading
  setTimeout(() => { loading.value = false }, 600)
})

// ─── Tab Navigation ─────────────────────────────────────────────────────────
const activeTab = ref('xgboost')
const tabItems = [
  { key: 'xgboost', label: 'Risk Scoring', icon: 'i-heroicons-chart-bar-square' },
  { key: 'isolation', label: 'Anomaly Detection', icon: 'i-heroicons-shield-exclamation' },
  { key: 'nlp', label: 'NLP Analysis', icon: 'i-heroicons-document-magnifying-glass' },
  { key: 'timeseries', label: 'KPI Forecast', icon: 'i-heroicons-arrow-trending-up' },
]

// ─── Tab 1: XGBoost Charts ─────────────────────────────────────────────────
const xgboostBarData = computed(() => ({
  labels: xgboost.chartLabels,
  datasets: [
    {
      label: 'Predicted Score',
      backgroundColor: 'rgba(99,102,241,0.7)',
      borderColor: 'rgb(99,102,241)',
      borderWidth: 1,
      borderRadius: 4,
      data: xgboost.predictedScores,
    },
    {
      label: 'Actual Score',
      backgroundColor: 'rgba(16,185,129,0.7)',
      borderColor: 'rgb(16,185,129)',
      borderWidth: 1,
      borderRadius: 4,
      data: xgboost.actualScores,
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
        label: (ctx: any) => `${ctx.dataset.label}: ${ctx.parsed.y.toFixed(1)}`
      }
    }
  },
  scales: {
    y: { beginAtZero: true, title: { display: true, text: 'Risk Score' } },
    x: { ticks: { maxRotation: 45 } }
  }
}

const featureBarData = computed(() => ({
  labels: xgboost.featureImportance.map(f => f.feature),
  datasets: [{
    label: 'Feature Importance',
    backgroundColor: [
      'rgba(239,68,68,0.7)', 'rgba(249,115,22,0.7)', 'rgba(234,179,8,0.7)',
      'rgba(34,197,94,0.7)', 'rgba(59,130,246,0.7)', 'rgba(139,92,246,0.7)',
      'rgba(236,72,153,0.7)', 'rgba(107,114,128,0.7)'
    ],
    borderWidth: 0,
    borderRadius: 4,
    data: xgboost.featureImportance.map(f => +(f.importance * 100).toFixed(1)),
  }]
}))

const featureBarOptions = {
  indexAxis: 'y' as const,
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: { callbacks: { label: (ctx: any) => `${ctx.parsed.x}%` } }
  },
  scales: {
    x: { beginAtZero: true, max: 35, title: { display: true, text: 'Importance (%)' } }
  }
}

// ─── Tab 2: Isolation Forest Charts ─────────────────────────────────────────
const scatterChartData = computed(() => {
  const normal = isolation.scatterData.filter(p => !p.isAnomaly)
  const anomalous = isolation.scatterData.filter(p => p.isAnomaly)
  return {
    datasets: [
      {
        label: 'Normal Transactions',
        data: normal.map(p => ({ x: p.x, y: p.y })),
        backgroundColor: 'rgba(59,130,246,0.4)',
        borderColor: 'rgba(59,130,246,0.6)',
        pointRadius: 4,
      },
      {
        label: 'Anomalies',
        data: anomalous.map(p => ({ x: p.x, y: p.y })),
        backgroundColor: 'rgba(239,68,68,0.7)',
        borderColor: 'rgba(239,68,68,1)',
        pointRadius: 7,
        pointStyle: 'triangle',
      },
    ]
  }
})

const scatterOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'top' as const },
    tooltip: {
      callbacks: {
        label: (ctx: any) => `Amount: ${ctx.parsed.x}, Frequency: ${ctx.parsed.y}`
      }
    }
  },
  scales: {
    x: { title: { display: true, text: 'Transaction Amount (Rp M)' } },
    y: { title: { display: true, text: 'Frequency' } }
  }
}

// ─── Tab 3: NLP Charts ──────────────────────────────────────────────────────
const categoryDoughnutData = computed(() => ({
  labels: Object.keys(nlp.categoryDistribution),
  datasets: [{
    data: Object.values(nlp.categoryDistribution),
    backgroundColor: [
      '#EF4444', '#8B5CF6', '#F59E0B', '#3B82F6',
      '#10B981', '#6366F1', '#EC4899',
    ],
    borderWidth: 2,
    borderColor: 'rgba(255,255,255,0.1)',
  }]
}))

const sentimentDoughnutData = computed(() => ({
  labels: ['Positive', 'Neutral', 'Negative'],
  datasets: [{
    data: [nlp.sentimentDistribution.positive, nlp.sentimentDistribution.neutral, nlp.sentimentDistribution.negative],
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

// ─── Tab 4: Time-Series Chart ───────────────────────────────────────────────
const timeSeriesChartData = computed(() => ({
  labels: timeseries.historicalKPI.map(p => p.period),
  datasets: [
    {
      label: 'Actual KPI',
      data: timeseries.historicalKPI.map(p => p.actual),
      borderColor: 'rgb(59,130,246)',
      backgroundColor: 'rgba(59,130,246,0.1)',
      tension: 0.4,
      pointRadius: 5,
      pointBackgroundColor: 'rgb(59,130,246)',
      spanGaps: false,
    },
    {
      label: 'Forecast',
      data: timeseries.historicalKPI.map(p => p.forecast),
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
      data: timeseries.historicalKPI.map(p => p.upperBound),
      borderColor: 'transparent',
      backgroundColor: 'rgba(139,92,246,0.08)',
      fill: '+1',
      tension: 0.4,
      pointRadius: 0,
      spanGaps: false,
    },
    {
      label: 'Lower Bound',
      data: timeseries.historicalKPI.map(p => p.lowerBound),
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
          return `${ctx.dataset.label}: ${ctx.parsed.y?.toFixed(1) ?? 'N/A'}`
        }
      }
    }
  },
  scales: {
    y: { title: { display: true, text: 'KPI Score (%)' } }
  }
}

// ─── Helpers ────────────────────────────────────────────────────────────────
type BadgeColor = 'error' | 'primary' | 'secondary' | 'success' | 'info' | 'warning' | 'neutral'

const severityColor = (s: string): BadgeColor => {
  const map: Record<string, BadgeColor> = { Critical: 'error', High: 'warning', Medium: 'info', Low: 'success' }
  return map[s] || 'neutral'
}

const sentimentColor = (s: string): BadgeColor => {
  const map: Record<string, BadgeColor> = { Positive: 'success', Neutral: 'warning', Negative: 'error' }
  return map[s] || 'neutral'
}

const alertColor = (l: string): BadgeColor => {
  const map: Record<string, BadgeColor> = { Critical: 'error', Warning: 'warning', Watch: 'info', None: 'success' }
  return map[l] || 'neutral'
}

const trendIcon = (t: string) => {
  const map: Record<string, string> = { Improving: 'i-heroicons-arrow-trending-up', Declining: 'i-heroicons-arrow-trending-down', Stable: 'i-heroicons-minus', up: 'i-heroicons-arrow-trending-up', down: 'i-heroicons-arrow-trending-down', stable: 'i-heroicons-minus' }
  return map[t] || 'i-heroicons-minus'
}

const trendColor = (t: string) => {
  if (['Improving', 'down'].includes(t)) return 'text-emerald-500'
  if (['Declining', 'up'].includes(t)) return 'text-rose-500'
  return 'text-gray-400'
}

const pct = (v: number) => `${(v * 100).toFixed(1)}%`

const keywordSizeClass = (count: number) => {
  if (count >= 20) return 'text-lg font-black'
  if (count >= 15) return 'text-base font-bold'
  if (count >= 10) return 'text-sm font-semibold'
  return 'text-md font-medium'
}

const keywordColor = (category: string): BadgeColor => {
  const map: Record<string, BadgeColor> = {
    'Financial': 'error', 'Technology': 'info', 'Operations': 'warning',
    'Compliance': 'success', 'Human Resources': 'neutral', 'Governance': 'primary', 'Strategic': 'secondary'
  }
  return map[category] || 'neutral'
}
</script>

<template>
  <div class="max-w-[1440px] mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <!-- ═══ Header ═══ -->
    <div class="relative mb-8 p-6 rounded-2xl bg-gradient-to-br from-indigo-600 via-violet-600 to-purple-700 shadow-2xl shadow-indigo-500/20 overflow-hidden">
      <div class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PGRlZnM+PHBhdHRlcm4gaWQ9ImciIHdpZHRoPSI2MCIgaGVpZ2h0PSI2MCIgcGF0dGVyblVuaXRzPSJ1c2VyU3BhY2VPblVzZSI+PHBhdGggZD0iTTAgMGg2MHY2MEgweiIgZmlsbD0ibm9uZSIvPjxwYXRoIGQ9Ik0zMCAwdjYwTTAgMzBoNjAiIHN0cm9rZT0icmdiYSgyNTUsMjU1LDI1NSwwLjA1KSIgc3Ryb2tlLXdpZHRoPSIxIi8+PC9wYXR0ZXJuPjwvZGVmcz48cmVjdCBmaWmdPSJ1cmwoI2cpIiB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIi8+PC9zdmc+')] opacity-50" />
      <div class="relative flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div class="flex items-center gap-4">
          <div class="w-14 h-14 rounded-xl bg-white/15 backdrop-blur-sm flex items-center justify-center">
            <UIcon name="i-heroicons-cpu-chip" class="w-8 h-8 text-white" />
          </div>
          <div>
            <h1 class="text-2xl font-extrabold text-white tracking-tight">AI-Powered Risk Analytics</h1>
            <p class="text-sm text-indigo-100/80">Machine Learning insights for risk-based internal audit</p>
          </div>
        </div>
        <div class="flex items-center gap-2 text-md text-indigo-200/80 bg-white/10 backdrop-blur-sm rounded-lg px-3 py-2">
          <UIcon name="i-heroicons-clock" class="w-4 h-4" />
          <span>Last model run: {{ new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) }}</span>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center items-center h-96">
      <div class="flex flex-col items-center gap-4">
        <UIcon name="i-heroicons-cpu-chip" class="w-12 h-12 animate-pulse text-indigo-500" />
        <span class="text-sm font-semibold text-gray-500 animate-pulse">Loading AI models...</span>
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
              <div class="text-2xl font-black text-indigo-600 dark:text-indigo-400">{{ summary.totalEntitiesScored }}</div>
            </div>
          </div>
        </UCard>
        <UCard class="border border-rose-100 dark:border-rose-900/50">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-rose-500/10 flex items-center justify-center">
              <UIcon name="i-heroicons-shield-exclamation" class="w-5 h-5 text-rose-500" />
            </div>
            <div>
              <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Anomalies</div>
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
              <div class="text-2xl font-black text-amber-600 dark:text-amber-400">{{ summary.documentsAnalyzed }}</div>
            </div>
          </div>
        </UCard>
        <UCard class="border border-violet-100 dark:border-violet-900/50">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-violet-500/10 flex items-center justify-center">
              <UIcon name="i-heroicons-bell-alert" class="w-5 h-5 text-violet-500" />
            </div>
            <div>
              <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">KPI Alerts</div>
              <div class="text-2xl font-black text-violet-600 dark:text-violet-400">{{ summary.kpiAlerts }}</div>
            </div>
          </div>
        </UCard>
      </div>

      <!-- ═══ Tabs ═══ -->
      <UTabs v-model="activeTab" :items="tabItems" class="w-full">
        <template #content="{ item }">
          <!-- ═══════════════════════════════════════════════════════════════ -->
          <!-- TAB 1: XGBOOST RISK SCORING                                   -->
          <!-- ═══════════════════════════════════════════════════════════════ -->
          <div v-if="item.key === 'xgboost'" class="space-y-6 py-6">
            <UAlert
              icon="i-heroicons-cpu-chip"
              color="primary"
              variant="subtle"
              title="XGBoost Risk Scoring Model"
              description="Predicts Likelihood and Impact scores for each entity based on historical KPI achievement, prior audit findings, transaction volume, and master data features."
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

              <!-- Model Metrics
              <UCard>
                <template #header>
                  <div class="flex items-center gap-2">
                    <UIcon name="i-heroicons-beaker" class="text-emerald-500" />
                    <h3 class="font-bold">Model Performance</h3>
                  </div>
                </template>
                <div class="space-y-5">
                  <div v-for="(value, label) in { 'Accuracy': xgboost.modelMetrics.accuracy, 'Precision': xgboost.modelMetrics.precision, 'Recall': xgboost.modelMetrics.recall, 'F1-Score': xgboost.modelMetrics.f1Score, 'AUC-ROC': xgboost.modelMetrics.auc }" :key="label">
                    <div class="flex justify-between items-center mb-1">
                      <span class="text-md font-bold text-gray-500">{{ label }}</span>
                      <span class="text-sm font-black">{{ pct(value) }}</span>
                    </div>
                    <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                      <div
                        class="h-2 rounded-full transition-all duration-700"
                        :class="value >= 0.9 ? 'bg-emerald-500' : value >= 0.8 ? 'bg-indigo-500' : 'bg-amber-500'"
                        :style="{ width: `${value * 100}%` }"
                      />
                    </div>
                  </div>
                </div>
              </UCard> -->
            

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
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Pred. Likelihood</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Pred. Impact</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Pred. Score</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Pred. Risk Level</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Actual</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Actual Risk Level</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Confidence</th>
                      <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">Trend</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="row in xgboost.predictions" :key="row.entity" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                      <td class="py-3 px-4 font-bold">{{ row.entity }}</td>
                      <td class="py-3 px-4"><UBadge :color="row.type === 'Branch' ? 'primary' : 'warning'" variant="subtle" size="md">{{ row.type }}</UBadge></td>
                      <td class="text-center py-3 px-4 font-mono">{{ row.predictedLikelihood.toFixed(1) }}</td>
                      <td class="text-center py-3 px-4 font-mono">{{ row.predictedImpact.toFixed(1) }}</td>
                      <td class="text-center py-3 px-4 font-mono font-bold">{{ row.predictedScore.toFixed(1) }}</td>
                      <td class="text-center py-3 px-4">
                        <UBadge
                          :style="{ backgroundColor: riskLevelConfig[row.predictedRiskLevel]?.color || '#9E9E9E', color: 'white' }"
                          variant="solid"
                          size="md"
                          class="font-bold"
                        >
                          {{ riskLevelConfig[row.predictedRiskLevel]?.label || row.predictedRiskLevel }}
                        </UBadge>
                      </td>
                      <td class="text-center py-3 px-4 font-mono">{{ row.actualScore }}</td>
                      <td class="text-center py-3 px-4">
                        <UBadge
                          :style="{ backgroundColor: riskLevelConfig[row.actualRiskLevel]?.color || '#9E9E9E', color: 'white' }"
                          variant="solid"
                          size="md"
                          class="font-bold"
                        >
                          {{ riskLevelConfig[row.actualRiskLevel]?.label || row.actualRiskLevel }}
                        </UBadge>
                      </td>
                      <td class="text-center py-3 px-4">
                        <span class="font-bold" :class="row.confidence >= 0.9 ? 'text-emerald-500' : 'text-amber-500'">{{ pct(row.confidence) }}</span>
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
          <!-- TAB 2: ISOLATION FOREST ANOMALY DETECTION                     -->
          <!-- ═══════════════════════════════════════════════════════════════ -->
          <div v-if="item.key === 'isolation'" class="space-y-6 py-6">
            <UAlert
              icon="i-heroicons-shield-exclamation"
              color="warning"
              variant="subtle"
              title="Isolation Forest Anomaly Detection"
              description="Scans transactions, fieldwork activities, and financial reports to identify unusual patterns that may indicate fraud, errors, or unidentified risks."
            />

            <!-- Summary Cards -->
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <UCard>
                <div class="text-center">
                  <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Records Scanned</div>
                  <div class="text-2xl font-black mt-1">{{ isolation.summary.totalScanned.toLocaleString() }}</div>
                </div>
              </UCard>
              <UCard>
                <div class="text-center">
                  <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Anomalies Found</div>
                  <div class="text-2xl font-black text-rose-500 mt-1">{{ isolation.summary.anomaliesFound }}</div>
                </div>
              </UCard>
              <UCard>
                <div class="text-center">
                  <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Contamination Rate</div>
                  <div class="text-2xl font-black text-amber-500 mt-1">{{ (isolation.summary.contaminationRate * 100).toFixed(1) }}%</div>
                </div>
              </UCard>
              <UCard>
                <div class="text-center">
                  <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Top Category</div>
                  <div class="text-2xl font-black text-indigo-500 mt-1">{{ isolation.summary.topCategory }}</div>
                </div>
              </UCard>
            </div>

            <!-- Critical Anomaly Alert -->
            <UAlert
              icon="i-heroicons-exclamation-triangle"
              color="error"
              variant="subtle"
              title="Critical Anomalies Detected"
              :description="`${isolation.anomalies.filter(a => a.severity === 'Critical').length} critical anomaly patterns require immediate investigation.`"
            />

            <!-- Scatter Plot -->
            <UCard>
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon name="i-heroicons-chart-bar" class="text-blue-500" />
                  <h3 class="font-bold">Transaction Pattern Distribution</h3>
                  <UBadge color="error" variant="subtle" size="md" class="ml-auto">▲ = Anomaly</UBadge>
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
                  <UIcon name="i-heroicons-flag" class="text-rose-500" />
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
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Score</th>
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[300px]">Description</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Severity</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Risk Level</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Date</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="a in isolation.anomalies" :key="a.id" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                      <td class="py-3 px-3 font-mono font-bold text-md">{{ a.id }}</td>
                      <td class="py-3 px-3 font-bold">{{ a.entity }}</td>
                      <td class="py-3 px-3"><UBadge color="neutral" variant="subtle" size="md">{{ a.type }}</UBadge></td>
                      <td class="text-center py-3 px-3 font-mono font-bold text-rose-500">{{ a.anomalyScore.toFixed(2) }}</td>
                      <td class="py-3 px-3 text-md leading-relaxed text-gray-600 dark:text-gray-300">{{ a.description }}</td>
                      <td class="text-center py-3 px-3"><UBadge :color="severityColor(a.severity)" variant="subtle" size="md">{{ a.severity }}</UBadge></td>
                      <td class="text-center py-3 px-3">
                        <UBadge
                          :style="{ backgroundColor: riskLevelConfig[a.riskLevel]?.color || '#9E9E9E', color: 'white' }"
                          variant="solid"
                          size="md"
                          class="font-bold"
                        >
                          {{ riskLevelConfig[a.riskLevel]?.label || a.riskLevel }}
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
              title="IndoBERT NLP Document Analysis"
              description="Analyzes Working Papers and Audit Result Reports in Bahasa Indonesia to auto-classify risk categories, detect sentiment, and assess finding severity."
            />

            <!-- Charts Row -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
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

              <!-- Keywords Cloud -->
              <UCard>
                <template #header>
                  <h3 class="font-bold text-sm">Top Keywords (Extracted)</h3>
                </template>
                <div class="flex flex-wrap gap-2 items-center justify-center py-2">
                  <UBadge
                    v-for="kw in nlp.topKeywords"
                    :key="kw.word"
                    :color="keywordColor(kw.category)"
                    variant="subtle"
                    :class="keywordSizeClass(kw.count)"
                    class="px-2.5 py-1"
                  >
                    {{ kw.word }}
                    <span class="ml-1 opacity-50 text-[9px]">{{ kw.count }}</span>
                  </UBadge>
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
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Conf.</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Sentiment</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Severity</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Risk Level</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="doc in nlp.documents" :key="doc.docId" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                      <td class="py-3 px-3 font-mono font-bold text-md">{{ doc.docId }}</td>
                      <td class="py-3 px-3">
                        <div class="font-bold text-md">{{ doc.title }}</div>
                        <div class="text-[11px] text-gray-400 mt-0.5 italic truncate max-w-[300px]">"{{ doc.excerpt }}"</div>
                      </td>
                      <td class="py-3 px-3"><UBadge :color="doc.source === 'Working Paper' ? 'primary' : 'warning'" variant="subtle" size="md">{{ doc.source }}</UBadge></td>
                      <td class="text-center py-3 px-3"><UBadge color="neutral" variant="subtle" size="md">{{ doc.autoCategory }}</UBadge></td>
                      <td class="text-center py-3 px-3 font-mono font-bold" :class="doc.confidence >= 0.9 ? 'text-emerald-500' : 'text-amber-500'">{{ pct(doc.confidence) }}</td>
                      <td class="text-center py-3 px-3"><UBadge :color="sentimentColor(doc.sentiment)" variant="subtle" size="md">{{ doc.sentiment }}</UBadge></td>
                      <td class="text-center py-3 px-3">
                        <div class="flex items-center justify-center gap-1.5">
                          <div class="w-12 bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
                            <div
                              class="h-1.5 rounded-full"
                              :class="doc.severityScore >= 70 ? 'bg-rose-500' : doc.severityScore >= 50 ? 'bg-amber-500' : 'bg-emerald-500'"
                              :style="{ width: `${doc.severityScore}%` }"
                            />
                          </div>
                          <span class="text-md font-bold">{{ doc.severityScore }}</span>
                        </div>
                      </td>
                      <td class="text-center py-3 px-3">
                        <UBadge
                          :style="{ backgroundColor: riskLevelConfig[doc.riskLevel]?.color || '#9E9E9E', color: 'white' }"
                          variant="solid"
                          size="md"
                          class="font-bold"
                        >
                          {{ riskLevelConfig[doc.riskLevel]?.label || doc.riskLevel }}
                        </UBadge>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </UCard>
          </div>

          <!-- ═══════════════════════════════════════════════════════════════ -->
          <!-- TAB 4: TIME-SERIES KPI FORECAST                               -->
          <!-- ═══════════════════════════════════════════════════════════════ -->
          <div v-if="item.key === 'timeseries'" class="space-y-6 py-6">
            <UAlert
              icon="i-heroicons-arrow-trending-up"
              color="info"
              variant="subtle"
              title="Time-Series KPI Forecasting"
              description="Predicts whether KPI performance for departments will decline next quarter, providing early warning signals for strategic audit planning."
            />

            <!-- At-Risk Alert -->
            <UAlert
              v-if="timeseries.atRiskDepartments.length > 0"
              icon="i-heroicons-exclamation-triangle"
              color="error"
              variant="subtle"
              title="Departments Predicted to Decline"
              :description="`${timeseries.atRiskDepartments.length} departments show declining KPI trends and require audit attention.`"
            />

            <!-- Forecast Chart + Model Info -->
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

              <!-- Model Accuracy + At-Risk Departments -->
              <div class="space-y-6">
                <!-- <UCard>
                  <template #header>
                    <h3 class="font-bold text-sm">Forecast Accuracy</h3>
                  </template>
                  <div class="space-y-4">
                    <div>
                      <div class="flex justify-between text-md mb-1">
                        <span class="text-gray-400 font-bold">MAPE</span>
                        <span class="font-black">{{ timeseries.forecastAccuracy.mape }}%</span>
                      </div>
                      <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
                        <div class="h-1.5 rounded-full bg-emerald-500" :style="{ width: `${100 - timeseries.forecastAccuracy.mape * 5}%` }" />
                      </div>
                    </div>
                    <div>
                      <div class="flex justify-between text-md mb-1">
                        <span class="text-gray-400 font-bold">RMSE</span>
                        <span class="font-black">{{ timeseries.forecastAccuracy.rmse }}</span>
                      </div>
                      <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
                        <div class="h-1.5 rounded-full bg-emerald-500" :style="{ width: `${100 - timeseries.forecastAccuracy.rmse * 5}%` }" />
                      </div>
                    </div>
                    <div>
                      <div class="flex justify-between text-md mb-1">
                        <span class="text-gray-400 font-bold">R² Score</span>
                        <span class="font-black">{{ timeseries.forecastAccuracy.r2Score }}</span>
                      </div>
                      <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
                        <div class="h-1.5 rounded-full bg-indigo-500" :style="{ width: `${timeseries.forecastAccuracy.r2Score * 100}%` }" />
                      </div>
                    </div>
                  </div>
                </UCard> -->

                <UCard>
                  <template #header>
                    <h3 class="font-bold text-sm">⚠️ At-Risk Departments</h3>
                  </template>
                  <div class="space-y-3">
                    <div v-for="dept in timeseries.atRiskDepartments" :key="dept.department" class="p-3 bg-rose-50 dark:bg-rose-900/20 rounded-lg border border-rose-200 dark:border-rose-800">
                      <div class="flex items-center justify-between gap-2">
                        <div class="font-bold text-md">{{ dept.department }}</div>
                        <UBadge
                          :style="{ backgroundColor: riskLevelConfig[dept.riskLevel]?.color || '#9E9E9E', color: 'white' }"
                          variant="solid"
                          size="md"
                          class="font-bold shrink-0"
                        >
                          {{ riskLevelConfig[dept.riskLevel]?.label || dept.riskLevel }}
                        </UBadge>
                      </div>
                      <div class="text-[10px] text-gray-500 mt-0.5">{{ dept.kpi }}</div>
                      <div class="flex items-center gap-2 mt-1.5">
                        <UIcon name="i-heroicons-arrow-trending-down" class="w-4 h-4 text-rose-500" />
                        <span class="text-md font-bold text-rose-500">{{ dept.predictedQ3.toFixed(1) }}% projected Q3</span>
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
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Current</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Forecast</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Trend</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Alert</th>
                      <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Risk Level</th>
                      <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[250px]">Recommended Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="kpi in timeseries.kpiForecasts" :key="kpi.code" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                      <td class="py-3 px-3 font-mono font-bold text-md">{{ kpi.code }}</td>
                      <td class="py-3 px-3 font-bold text-md">{{ kpi.kpiName }}</td>
                      <td class="text-center py-3 px-3 font-mono">{{ kpi.currentValue.toFixed(1) }}{{ kpi.unit === '%' ? '%' : '' }}</td>
                      <td class="text-center py-3 px-3 font-mono font-bold">{{ kpi.forecastedValue.toFixed(1) }}{{ kpi.unit === '%' ? '%' : '' }}</td>
                      <td class="text-center py-3 px-3">
                        <div class="flex items-center justify-center gap-1">
                          <UIcon :name="trendIcon(kpi.trend)" class="w-4 h-4" :class="trendColor(kpi.trend)" />
                          <span class="text-md font-bold" :class="trendColor(kpi.trend)">{{ kpi.trend }}</span>
                        </div>
                      </td>
                      <td class="text-center py-3 px-3"><UBadge :color="alertColor(kpi.alertLevel)" variant="subtle" size="md">{{ kpi.alertLevel }}</UBadge></td>
                      <td class="text-center py-3 px-3">
                        <UBadge
                          :style="{ backgroundColor: riskLevelConfig[kpi.riskLevel]?.color || '#9E9E9E', color: 'white' }"
                          variant="solid"
                          size="md"
                          class="font-bold"
                        >
                          {{ riskLevelConfig[kpi.riskLevel]?.label || kpi.riskLevel }}
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