<script setup lang="ts">
import { computed, onMounted } from 'vue'
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
const {
  loading,
  timeseriesState,
  formatNum,
  getRiskConfig,
  trendColor,
  trendIcon,
  fetchAiAnalytics
} = useAiAnalytics()

onMounted(() => {
  fetchAiAnalytics()
})

const timeSeriesChartData = computed(() => ({
  labels: (timeseriesState.value.historicalKPI || []).map((p: any) => p.period || ''),
  datasets: [
    {
      label: t('analytics.timeseries.labelActualKPI'),
      data: (timeseriesState.value.historicalKPI || []).map((p: any) => p.actual ?? null),
      borderColor: 'rgb(59,130,246)',
      backgroundColor: 'rgba(59,130,246,0.1)',
      tension: 0.4,
      pointRadius: 5,
      pointBackgroundColor: 'rgb(59,130,246)',
      spanGaps: false,
    },
    {
      label: t('analytics.timeseries.labelForecast'),
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
      label: t('analytics.timeseries.labelUpperBound'),
      data: (timeseriesState.value.historicalKPI || []).map((p: any) => p.upperBound ?? null),
      borderColor: 'transparent',
      backgroundColor: 'rgba(139,92,246,0.08)',
      fill: '+1',
      tension: 0.4,
      pointRadius: 0,
      spanGaps: false,
    },
    {
      label: t('analytics.timeseries.labelLowerBound'),
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
    y: { title: { display: true, text: t('analytics.timeseries.axisKPIScore') } }
  }
}))
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <AiInsightHeader
      :title="t('navigation.kpiForecast')"
      :subtitle="t('analytics.subtitle')"
      icon="i-lucide-trending-up"
      currentSubModule="kpi-forecast"
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
        icon="i-heroicons-arrow-trending-up"
        color="info"
        variant="subtle"
        :title="t('analytics.timeseries.alertTitle')"
        :description="t('analytics.timeseries.alertDesc')"
      />

      <!-- Forecast Chart -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <UCard class="lg:col-span-2">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-chart-bar" class="text-violet-500" />
              <h3 class="font-bold">{{ t('analytics.timeseries.chartTitle') }}</h3>
              <UBadge color="info" variant="subtle" size="md" class="ml-auto">{{ t('analytics.timeseries.chartConfidenceBadge') }}</UBadge>
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
              <h3 class="font-bold text-sm">{{ t('analytics.timeseries.atRiskTitle') }}</h3>
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
                  <span class="text-md font-bold text-rose-500">{{ formatNum(dept.predictedQ3, 1) }}{{ t('analytics.timeseries.projectedQ3') }}</span>
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
            <h3 class="font-bold">{{ t('analytics.timeseries.tableTitle') }}</h3>
          </div>
        </template>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colCode') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colKPI') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colEntity') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colType') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colForecastHorizon') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colCurrent') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colForecast') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colTrend') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.timeseries.colRiskLevel') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[250px]">{{ t('analytics.timeseries.colRecommendedAction') }}</th>
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
  </div>
</template>
