<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Bar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend
} from 'chart.js'
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import { useI18n } from '~/composables/useI18n'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import AiInsightHeader from '~/components/analytics/AiInsightHeader.vue'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend)

const { t } = useI18n()
const toast = useToast()
const {
  loading,
  xgboostState,
  formatNum,
  getRiskConfig,
  fetchAiAnalytics,
  syncEntityPredictionWithRiskProfile,
  syncEntityPredictionWithAuditUniverse
} = useAiAnalytics()

const isSyncing = ref(false)
const searchQuery = ref('')
const chartLimit = ref<'10' | '15' | 'ALL'>('10')

const handleSyncWithProfile = async () => {
  isSyncing.value = true
  try {
    const auditUniverseStore = useAuditUniverseStore()
    const currentYear = new Date().getFullYear()

    await Promise.all([
      auditUniverseStore.fetchYearlyUniverse(currentYear).catch(() => {}),
      auditUniverseStore.fetchCorporateUniverse().catch(() => {})
    ])

    const res = await (syncEntityPredictionWithAuditUniverse || syncEntityPredictionWithRiskProfile)()
    if (res && res.predictions) {
      toast.success(
        'Sinkronisasi Berhasil',
        `Model AI berhasil melatih ulang ${res.predictions.length} auditable entity langsung dari Audit Universe (Audit Priority) (Zero CSV, Tanpa Batas).`
      )
    } else {
      toast.info('Sinkronisasi Selesai', 'Data auditable entity telah diperbarui dan diselaraskan.')
    }
  } catch (err: any) {
    toast.error('Gagal Sinkronisasi', err?.message || String(err))
  } finally {
    isSyncing.value = false
  }
}

onMounted(async () => {
  const auditUniverseStore = useAuditUniverseStore()
  const currentYear = new Date().getFullYear()

  // 1. Fetch other AI analytics (anomalies, NLP, etc.)
  await fetchAiAnalytics().catch(() => {})

  // 2. Fetch and synchronize Audit Universe entities so data persists across refreshes
  await Promise.all([
    auditUniverseStore.fetchYearlyUniverse(currentYear).catch(() => {}),
    auditUniverseStore.fetchCorporateUniverse().catch(() => {})
  ])

  // Automatically sync with live Audit Universe so real auditable entities are immediately trained & predicted without limits
  await (syncEntityPredictionWithAuditUniverse || syncEntityPredictionWithRiskProfile)().catch(() => {})
})

const filteredPredictions = computed(() => {
  let list = xgboostState.value.predictions || []
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim()
    list = list.filter((p: any) =>
      p.entity && String(p.entity).toLowerCase().includes(q)
    )
  }
  return list
})

const chartPredictions = computed(() => {
  const sorted = [...filteredPredictions.value].sort((a: any, b: any) => {
    const scoreA = Math.max(a.predictedScore ?? 0, a.actualScore ?? 0)
    const scoreB = Math.max(b.predictedScore ?? 0, b.actualScore ?? 0)
    return scoreB - scoreA
  })
  if (chartLimit.value === '10') return sorted.slice(0, 10)
  if (chartLimit.value === '15') return sorted.slice(0, 15)
  return sorted
})

const xgboostBarData = computed(() => ({
  labels: chartPredictions.value.map((p: any) => {
    const name = p.entity || 'Entity'
    return name.length > 20 ? name.slice(0, 18) + '…' : name
  }),
  datasets: [
    {
      label: t('analytics.xgboost.labelPredictedScore'),
      backgroundColor: 'rgba(99,102,241,0.85)',
      borderColor: 'rgb(99,102,241)',
      borderWidth: 1.5,
      borderRadius: 6,
      maxBarThickness: 32,
      data: chartPredictions.value.map((p: any) => p.predictedScore ?? 0),
    },
    {
      label: t('analytics.xgboost.labelActualScore'),
      backgroundColor: 'rgba(16,185,129,0.85)',
      borderColor: 'rgb(16,185,129)',
      borderWidth: 1.5,
      borderRadius: 6,
      maxBarThickness: 32,
      data: chartPredictions.value.map((p: any) => p.actualScore ?? 0),
    },
  ],
}))

const xgboostBarOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        boxWidth: 12,
        usePointStyle: true,
        pointStyle: 'circle'
      }
    },
    tooltip: {
      callbacks: {
        title: (items: any[]) => {
          if (!items.length) return ''
          const idx = items[0].dataIndex
          const item = chartPredictions.value[idx]
          return item ? `${item.entity} (${item.targetTimeline || ''})` : ''
        },
        label: (ctx: any) => {
          const idx = ctx.dataIndex
          const item = chartPredictions.value[idx]
          if (!item) return `${ctx.dataset.label}: ${formatNum(ctx.parsed.y, 1)}`
          if (ctx.datasetIndex === 0) {
            return `Prediksi: ${formatNum(item.predictedScore, 1)} [Likelihood: ${formatNum(item.predictedLikelihood, 1)}, Impact: ${formatNum(item.predictedImpact, 1)}] - ${item.predictedRiskLevel}`
          } else {
            return `Current: ${formatNum(item.actualScore, 1)} [Likelihood: ${formatNum(item.currentLikelihood ?? 3, 1)}, Impact: ${formatNum(item.currentImpact ?? 3, 1)}] - ${item.actualRiskLevel}`
          }
        }
      }
    }
  },
  scales: {
    y: {
      beginAtZero: true,
      max: 25,
      title: { display: true, text: t('analytics.xgboost.axisRiskScore') },
      grid: {
        color: 'rgba(156, 163, 175, 0.1)'
      }
    },
    x: {
      grid: { display: false },
      ticks: {
        maxRotation: chartPredictions.value.length > 8 ? 35 : 0,
        autoSkip: false,
        font: { size: 11 }
      }
    }
  }
}))

const featureBarData = computed(() => ({
  labels: xgboostState.value.featureImportance.map((f: any) => f.feature || 'Feature'),
  datasets: [{
    label: t('analytics.xgboost.labelFeatureImportance'),
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

const featureBarOptions = computed(() => ({
  indexAxis: 'y' as const,
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: { callbacks: { label: (ctx: any) => `${formatNum(ctx.parsed.x, 1)}%` } }
  },
  scales: {
    x: { beginAtZero: true, max: 40, title: { display: true, text: t('analytics.xgboost.axisImportancePct') } }
  }
}))
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <AiInsightHeader
      :title="t('navigation.entityPrediction') || t('navigation.riskScoringPrediction')"
      :subtitle="t('analytics.subtitle')"
      icon="i-lucide-binary"
      currentSubModule="risk-scoring"
    />

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center items-center h-96">
      <div class="flex flex-col items-center gap-4">
        <UIcon name="i-heroicons-cpu-chip" class="w-12 h-12 animate-pulse text-indigo-500" />
        <span class="text-sm font-semibold text-gray-500 animate-pulse">{{ t('analytics.loading') }}</span>
      </div>
    </div>

    <!-- Content -->
    <div v-else class="space-y-6">
      <!-- Sync & Pipeline Status Bar -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-xl border border-indigo-500/20 bg-indigo-50/50 dark:bg-indigo-950/20">
        <div class="flex items-center gap-3">
          <div class="flex size-10 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 shrink-0">
            <UIcon name="i-lucide-sparkles" class="size-5" />
          </div>
          <div>
            <div class="flex items-center gap-2">
              <h4 class="text-sm font-bold text-gray-900 dark:text-white">
                Live Data Pipeline (Zero CSV Dependency)
              </h4>
              <UBadge color="success" variant="subtle" size="xs">
                Direct Memory-Mapped Parquet
              </UBadge>
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              Model AI dilatih langsung dari data Auditable Entity pada Audit Universe (Audit Priority) secara realtime tanpa membebani sistem I/O CSV.
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2 self-start sm:self-center shrink-0">
          <UButton
            color="primary"
            variant="solid"
            icon="i-lucide-refresh-cw"
            :loading="isSyncing"
            @click="handleSyncWithProfile"
          >
            Sinkronkan & Latih Ulang AI
          </UButton>
        </div>
      </div>

      <UAlert
        icon="i-heroicons-cpu-chip"
        color="primary"
        variant="subtle"
        :title="t('analytics.xgboost.alertTitle')"
        :description="t('analytics.xgboost.alertDesc')"
      />

      <!-- Search & Filters -->
      <div class="flex flex-wrap items-center justify-between gap-3 bg-white dark:bg-gray-900 p-3.5 rounded-xl border border-gray-200 dark:border-gray-800 shadow-xs">
        <div class="flex items-center gap-2 flex-1 min-w-[240px]">
          <UInput
            v-model="searchQuery"
            icon="i-heroicons-magnifying-glass"
            placeholder="Cari auditable entity atau cabang..."
            size="sm"
            class="w-full max-w-sm"
          />
        </div>

        <div class="flex items-center gap-2.5">
          <UBadge color="neutral" variant="subtle" size="sm">
            {{ filteredPredictions.length }} / {{ xgboostState.predictions.length }} Entitas
          </UBadge>
        </div>
      </div>

      <!-- Charts Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Predicted vs Actual Chart -->
        <UCard class="lg:col-span-2">
          <template #header>
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div class="flex items-center gap-2">
                <UIcon name="i-heroicons-chart-bar" class="text-indigo-500" />
                <h3 class="font-bold">{{ t('analytics.xgboost.chartPredictedVsActual') }}</h3>
                <UBadge color="primary" variant="subtle" size="xs">
                  {{ chartPredictions.length }} Entitas
                </UBadge>
              </div>
              <div class="flex items-center gap-1.5 bg-gray-100 dark:bg-gray-800 p-1 rounded-lg">
                <button
                  type="button"
                  class="px-2.5 py-1 text-xs font-semibold rounded-md transition-all"
                  :class="chartLimit === '10' ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-400 shadow-xs' : 'text-gray-500 hover:text-gray-900 dark:hover:text-white'"
                  @click="chartLimit = '10'"
                >
                  Top 10
                </button>
                <button
                  type="button"
                  class="px-2.5 py-1 text-xs font-semibold rounded-md transition-all"
                  :class="chartLimit === '15' ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-400 shadow-xs' : 'text-gray-500 hover:text-gray-900 dark:hover:text-white'"
                  @click="chartLimit = '15'"
                >
                  Top 15
                </button>
                <button
                  type="button"
                  class="px-2.5 py-1 text-xs font-semibold rounded-md transition-all"
                  :class="chartLimit === 'ALL' ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-400 shadow-xs' : 'text-gray-500 hover:text-gray-900 dark:hover:text-white'"
                  @click="chartLimit = 'ALL'"
                >
                  Semua
                </button>
              </div>
            </div>
          </template>
          <div class="h-80 overflow-x-auto">
            <div :style="{ minWidth: chartLimit === 'ALL' && chartPredictions.length > 12 ? `${chartPredictions.length * 56}px` : '100%', height: '100%' }">
              <Bar :data="xgboostBarData" :options="xgboostBarOptions" />
            </div>
          </div>
        </UCard>

        <!-- Feature Importance -->
        <UCard class="lg:col-span-1">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-adjustments-horizontal" class="text-orange-500" />
              <h3 class="font-bold">{{ t('analytics.xgboost.chartFeatureImportance') }}</h3>
            </div>
          </template>
          <div class="h-80">
            <Bar :data="featureBarData" :options="featureBarOptions" />
          </div>
        </UCard>
      </div>

      <!-- Predictions Table -->
      <UCard class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm ring-0 divide-gray-200 dark:divide-gray-800">
        <template #header>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-table-cells" class="text-blue-500" />
              <h3 class="font-bold">{{ t('analytics.xgboost.tableTitle') }}</h3>
            </div>
            <span class="text-xs text-gray-500">
              Sinkronisasi Audit Universe (Audit Priority) & 5x5 Matrix Corporate Risk Profile
            </span>
          </div>
        </template>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-100 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800">
                <th class="text-left py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colEntity') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colCurrentLikelihood') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colCurrentImpact') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colActualRiskLevel') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colTargetPeriod') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colPredLikelihood') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colPredImpact') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colPredRiskLevel') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in filteredPredictions" :key="row.entity" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                <td class="py-3 px-4 font-bold">{{ row.entity }}</td>
                <td class="text-center py-3 px-4 tabular-nums font-medium">{{ formatNum(row.currentLikelihood ?? row.actualLikelihood ?? 3, 1) }}</td>
                <td class="text-center py-3 px-4 tabular-nums font-medium">{{ formatNum(row.currentImpact ?? row.actualImpact ?? 3, 1) }}</td>
                <td class="text-center py-3 px-4">
                  <UBadge
                    :class="getRiskConfig(row.actualRiskLevel).badgeClass"
                    variant="solid"
                    size="md"
                    class="font-bold"
                  >
                    {{ getRiskConfig(row.actualRiskLevel).label }}
                  </UBadge>
                </td>
                <td class="text-center py-3 px-4"><UBadge color="info" variant="subtle" size="md" class="font-bold">{{ row.targetTimeline || 'FY 2026' }}</UBadge></td>
                <td class="text-center py-3 px-4 tabular-nums font-semibold">{{ formatNum(row.predictedLikelihood, 1) }}</td>
                <td class="text-center py-3 px-4 tabular-nums font-semibold">{{ formatNum(row.predictedImpact, 1) }}</td>
                <td class="text-center py-3 px-4">
                  <UBadge
                    :class="getRiskConfig(row.predictedRiskLevel).badgeClass"
                    variant="solid"
                    size="md"
                    class="font-bold"
                  >
                    {{ getRiskConfig(row.predictedRiskLevel).label }}
                  </UBadge>
                </td>
              </tr>
              <tr v-if="filteredPredictions.length === 0">
                <td colspan="8" class="text-center py-8 text-gray-500">
                  Tidak ada entitas yang sesuai dengan kriteria pencarian.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </UCard>
    </div>
  </div>
</template>
