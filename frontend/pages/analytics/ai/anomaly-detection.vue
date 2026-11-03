<script setup lang="ts">
import { ref, computed, watchEffect, onMounted } from 'vue'
import { Scatter } from 'vue-chartjs'
import {
  Chart as ChartJS,
  LinearScale,
  PointElement,
  Tooltip,
  Legend
} from 'chart.js'
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import {
  buildAnomalyScatterPoints,
  ANOMALY_POINT_COLORS,
  createAnomalyTypeConfigs,
  listAnomalyTypes,
  type AnomalyTypeConfig
} from '~/composables/useAnomalyScatter'
import { formatRupiahFromMillions, normalizeRupiahInText } from '~/utils/analyticsFormat'
import { useI18n } from '~/composables/useI18n'
import AiInsightHeader from '~/components/analytics/AiInsightHeader.vue'

ChartJS.register(LinearScale, PointElement, Tooltip, Legend)

const { t, locale } = useI18n()
const {
  loading,
  isolationState,
  summary,
  formatNum,
  getRiskConfig,
  fetchAiAnalytics
} = useAiAnalytics()

onMounted(() => {
  fetchAiAnalytics()
})

const anomalyTypeConfigs = computed<Record<string, AnomalyTypeConfig>>(() => createAnomalyTypeConfigs(t, locale.value))

const selectedAnomalyType = ref('Funding')

const availableAnomalyTypes = computed(() =>
  listAnomalyTypes(
    isolationState.value.anomalies,
    isolationState.value.scatterData,
    Object.keys(anomalyTypeConfigs.value)
  )
)

const tableCategoryFilter = ref('All')

const filteredTableAnomalies = computed(() => {
  if (tableCategoryFilter.value === 'All') {
    return isolationState.value.anomalies
  }
  return isolationState.value.anomalies.filter((a: any) => a.type === tableCategoryFilter.value)
})

const getCategoryBadgeColor = (type: string) => {
  switch (type) {
    case 'Funding': return 'info'
    case 'Lending': return 'error'
    case 'Treasury': return 'primary'
    case 'Payment': return 'warning'
    case 'KYC': return 'secondary'
    case 'IT Control': return 'info'
    case 'Transaction': return 'error'
    case 'Procurement': return 'primary'
    case 'Expense Report': return 'warning'
    case 'Travel Expense': return 'neutral'
    case 'Inventory': return 'neutral'
    case 'Fieldwork': return 'secondary'
    default: return 'neutral'
  }
}

watchEffect(() => {
  if ((selectedAnomalyType.value === 'All' || !selectedAnomalyType.value || !availableAnomalyTypes.value.includes(selectedAnomalyType.value)) && availableAnomalyTypes.value.length > 0) {
    const firstType = availableAnomalyTypes.value[0]
    if (firstType) {
      selectedAnomalyType.value = firstType
    }
  }
})

const getScatterChartForType = (typeFilter: string) => {
  const { normalPoints, anomalyPoints } = buildAnomalyScatterPoints(
    isolationState.value.anomalies,
    isolationState.value.scatterData,
    typeFilter
  )

  const defaultConfig: AnomalyTypeConfig = {
    xAxisTitle: t('analytics.isolation.anomalyTypes.fundingXAxis'),
    unit: t('analytics.isolation.anomalyTypes.fundingUnit'),
    formatX: (val) => formatRupiahFromMillions(val, locale.value),
    colors: { ...ANOMALY_POINT_COLORS, style: 'circle' }
  }
  const config = anomalyTypeConfigs.value[typeFilter] || anomalyTypeConfigs.value['Funding'] || defaultConfig
  const colors = config.colors

  return {
    datasets: [
      {
        label: t('analytics.isolation.labelNormalData', { type: typeFilter }),
        data: normalPoints,
        backgroundColor: 'rgba(59,130,246,0.3)',
        borderColor: 'rgba(59,130,246,0.5)',
        pointRadius: 4,
      },
      {
        label: t('analytics.isolation.labelAnomalies', { type: typeFilter }),
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
  const currentConfig = anomalyTypeConfigs.value[selectedAnomalyType.value] || {
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
            return `${metricName}: ${formattedX}, ${t('analytics.isolation.hourValue', { hour: ctx.parsed.y })}`
          }
        }
      }
    },
    scales: {
      x: {
        title: { display: true, text: currentConfig.xAxisTitle },
        ticks: { maxTicksLimit: 6, callback: (v: any) => currentConfig.formatX ? currentConfig.formatX(Number(v)) : v }
      },
      // Y = transaction hour from the data (0-23).
      y: { min: 0, max: 23, ticks: { stepSize: 3 }, title: { display: true, text: t('analytics.isolation.axisHour') } }
    }
  }
})
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <AiInsightHeader
      :title="t('navigation.anomalyDetection')"
      :subtitle="t('analytics.subtitle')"
      icon="i-lucide-shield-alert"
      currentSubModule="anomaly-detection"
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
        icon="i-heroicons-shield-exclamation"
        color="warning"
        variant="subtle"
        :title="t('analytics.isolation.alertTitle')"
        :description="t('analytics.isolation.alertDesc')"
      />

      <!-- Summary Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <UCard>
          <div class="text-center">
            <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.recordsScanned') }}</div>
            <div class="text-2xl font-black mt-1 tabular-nums">{{ (isolationState.summary?.totalScanned || 0).toLocaleString() }}</div>
          </div>
        </UCard>
        <UCard>
          <div class="text-center">
            <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.anomaliesFound') }}</div>
            <div class="text-2xl font-black text-rose-500 mt-1 tabular-nums">{{ isolationState.summary?.anomaliesFound || summary.anomaliesDetected || 150 }}</div>
          </div>
        </UCard>
        <UCard>
          <div class="text-center">
            <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.contaminationRate') }}</div>
            <div class="text-2xl font-black text-amber-500 mt-1 tabular-nums">{{ formatNum((isolationState.summary?.contaminationRate || 0.125) * 100, 1) }}%</div>
          </div>
        </UCard>
        <UCard>
          <div class="text-center">
            <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.topCategory') }}</div>
            <div class="text-2xl font-black text-indigo-500 mt-1">{{ isolationState.summary?.topCategory || 'Funding' }}</div>
          </div>
        </UCard>
      </div>

      <!-- Scatter Chart Card with Type Filter Tabs -->
      <UCard>
        <template #header>
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-chart-bar" class="text-blue-500" />
              <h3 class="font-bold">{{ t('analytics.isolation.chartTitle') }}</h3>
            </div>
            <!-- Anomaly Type Selector Buttons -->
            <div class="flex flex-wrap gap-1">
              <UButton
                v-for="at in availableAnomalyTypes"
                :key="at"
                :color="selectedAnomalyType === at ? 'primary' : 'neutral'"
                :variant="selectedAnomalyType === at ? 'solid' : 'ghost'"
                size="md"
                @click="() => { selectedAnomalyType = at; tableCategoryFilter = at; }"
              >
                {{ at }}
              </UButton>
            </div>
          </div>
        </template>
        <div class="h-80">
          <Scatter :data="scatterChartData" :options="scatterOptions" />
        </div>
      </UCard>

      <!-- Anomalies Table -->
      <UCard class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm ring-0 divide-gray-200 dark:divide-gray-800">
        <template #header>
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-table-cells" class="text-rose-500" />
              <h3 class="font-bold">{{ t('analytics.isolation.tableTitle') }}</h3>
            </div>
            <!-- Filter Kategori Tabel -->
            <div class="flex flex-wrap items-center gap-1.5">
              <span class="text-xs text-gray-400 font-medium mr-1">{{ t('analytics.anomalyFilter') }}</span>
              <UButton
                size="xs"
                :color="tableCategoryFilter === 'All' ? 'primary' : 'neutral'"
                :variant="tableCategoryFilter === 'All' ? 'solid' : 'ghost'"
                @click="() => { tableCategoryFilter = 'All' }"
              >
                {{ t('analytics.all') }} ({{ isolationState.anomalies.length }})
              </UButton>
              <UButton
                v-for="cat in availableAnomalyTypes"
                :key="cat"
                size="xs"
                :color="tableCategoryFilter === cat ? 'primary' : 'neutral'"
                :variant="tableCategoryFilter === cat ? 'solid' : 'ghost'"
                @click="() => { tableCategoryFilter = cat }"
              >
                {{ cat }} ({{ isolationState.anomalies.filter((a: any) => a.type === cat).length }})
              </UButton>
            </div>
          </div>
        </template>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-100 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800">
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.colID') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.colEntity') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.colType') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[300px]">{{ t('analytics.isolation.colDescription') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.colRiskLevel') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.colDate') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="a in filteredTableAnomalies" :key="a.id" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                <td class="py-3 px-3 font-mono font-bold text-md">{{ a.id }}</td>
                <td class="py-3 px-3 font-bold">{{ a.entity }}</td>
                <td class="py-3 px-3">
                  <UBadge :color="getCategoryBadgeColor(a.type)" variant="subtle" size="md">{{ a.type }}</UBadge>
                </td>
                <td class="py-3 px-3 text-md leading-relaxed text-gray-600 dark:text-gray-300">{{ normalizeRupiahInText(a.description, locale) }}</td>
                <td class="text-center py-3 px-3">
                  <UBadge
                    :class="getRiskConfig(a.riskLevel).badgeClass"
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
  </div>
</template>
