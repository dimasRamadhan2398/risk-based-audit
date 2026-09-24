<script setup lang="ts">
import { computed, onMounted } from 'vue'
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
import AiInsightHeader from '~/components/analytics/AiInsightHeader.vue'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend)

const { t } = useI18n()
const {
  loading,
  xgboostState,
  formatNum,
  getRiskConfig,
  riskCategoryColor,
  fetchAiAnalytics
} = useAiAnalytics()

onMounted(() => {
  fetchAiAnalytics()
})

const xgboostBarData = computed(() => ({
  labels: xgboostState.value.predictions.map((p: any) => p.entity || 'Entity'),
  datasets: [
    {
      label: t('analytics.xgboost.labelPredictedScore'),
      backgroundColor: 'rgba(99,102,241,0.75)',
      borderColor: 'rgb(99,102,241)',
      borderWidth: 1.5,
      borderRadius: 4,
      data: xgboostState.value.predictions.map((p: any) => p.predictedScore ?? 0),
    },
    {
      label: t('analytics.xgboost.labelActualScore'),
      backgroundColor: 'rgba(16,185,129,0.75)',
      borderColor: 'rgb(16,185,129)',
      borderWidth: 1.5,
      borderRadius: 4,
      data: xgboostState.value.predictions.map((p: any) => p.actualScore ?? 0),
    },
  ],
}))

const xgboostBarOptions = computed(() => ({
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
    y: { beginAtZero: true, title: { display: true, text: t('analytics.xgboost.axisRiskScore') } },
    x: { ticks: { maxRotation: 45 } }
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
      :title="t('navigation.riskScoringPrediction')"
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
      <UAlert
        icon="i-heroicons-cpu-chip"
        color="primary"
        variant="subtle"
        :title="t('analytics.xgboost.alertTitle')"
        :description="t('analytics.xgboost.alertDesc')"
      />

      <!-- Charts Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Predicted vs Actual Chart -->
        <UCard class="lg:col-span-2">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-chart-bar" class="text-indigo-500" />
              <h3 class="font-bold">{{ t('analytics.xgboost.chartPredictedVsActual') }}</h3>
            </div>
          </template>
          <div class="h-80">
            <Bar :data="xgboostBarData" :options="xgboostBarOptions" />
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
      <UCard>
        <template #header>
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-table-cells" class="text-blue-500" />
            <h3 class="font-bold">{{ t('analytics.xgboost.tableTitle') }}</h3>
          </div>
        </template>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="text-left py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colEntity') }}</th>
                <th class="text-left py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colType') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colRiskCategory') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colActual') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colActualRiskLevel') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colTargetPeriod') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colPredLikelihood') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colPredImpact') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colPredScore') }}</th>
                <th class="text-center py-3 px-4 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colPredRiskLevel') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in xgboostState.predictions" :key="row.entity" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                <td class="py-3 px-4 font-bold">{{ row.entity }}</td>
                <td class="py-3 px-4"><UBadge :color="row.type === 'Branch' ? 'primary' : 'warning'" variant="subtle" size="md">{{ row.type }}</UBadge></td>
                <td class="text-center py-3 px-4"><UBadge :color="riskCategoryColor(row.riskCategory)" variant="subtle" size="md">{{ row.riskCategory }}</UBadge></td>
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
              </tr>
            </tbody>
          </table>
        </div>
      </UCard>
    </div>
  </div>
</template>
