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
import { useI18n } from '~/composables/useI18n'
import AiInsightHeader from '~/components/analytics/AiInsightHeader.vue'

ChartJS.register(LinearScale, PointElement, Tooltip, Legend)

const { t } = useI18n()
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

interface AnomalyTypeConfig {
  xAxisTitle: string
  unit: string
  formatX: (val: number) => string
  colors: { bg: string, border: string, style: string }
}

const anomalyTypeConfigs = computed<Record<string, AnomalyTypeConfig>>(() => ({
  'Funding': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.fundingXAxis'),
    unit: t('analytics.isolation.anomalyTypes.fundingUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(16,185,129,0.85)', border: 'rgba(16,185,129,1)', style: 'circle' }
  },
  'Lending': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.lendingXAxis'),
    unit: t('analytics.isolation.anomalyTypes.lendingUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(239,68,68,0.85)', border: 'rgba(239,68,68,1)', style: 'triangle' }
  },
  'Treasury': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.treasuryXAxis'),
    unit: t('analytics.isolation.anomalyTypes.treasuryUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(99,102,241,0.85)', border: 'rgba(99,102,241,1)', style: 'rectRot' }
  },
  'Payment': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.paymentXAxis'),
    unit: t('analytics.isolation.anomalyTypes.paymentUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(245,158,11,0.85)', border: 'rgba(245,158,11,1)', style: 'rect' }
  },
  'KYC': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.kycXAxis'),
    unit: t('analytics.isolation.anomalyTypes.kycUnit'),
    formatX: (val) => `${val}%`,
    colors: { bg: 'rgba(236,72,153,0.85)', border: 'rgba(236,72,153,1)', style: 'star' }
  },
  'IT Control': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.itControlXAxis'),
    unit: t('analytics.isolation.anomalyTypes.itControlUnit'),
    formatX: (val) => `${val} ${t('analytics.isolation.anomalyTypes.itControlUnit')}`,
    colors: { bg: 'rgba(20,184,166,0.85)', border: 'rgba(20,184,166,1)', style: 'crossRot' }
  },
  'Fieldwork': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.fieldworkXAxis'),
    unit: t('analytics.isolation.anomalyTypes.fieldworkUnit'),
    formatX: (val) => `${val} ${t('analytics.isolation.anomalyTypes.fieldworkUnit')}`,
    colors: { bg: 'rgba(139,92,246,0.85)', border: 'rgba(139,92,246,1)', style: 'star' }
  },
  'Access Pattern': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.accessPatternXAxis'),
    unit: t('analytics.isolation.anomalyTypes.accessPatternUnit'),
    formatX: (val) => `${t('analytics.isolation.anomalyTypes.accessPatternUnit')} ${val}:00`,
    colors: { bg: 'rgba(249,115,22,0.85)', border: 'rgba(249,115,22,1)', style: 'rectRot' }
  },
  'Data Access': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.dataAccessXAxis'),
    unit: t('analytics.isolation.anomalyTypes.dataAccessUnit'),
    formatX: (val) => `${val} ${t('analytics.isolation.anomalyTypes.dataAccessUnit')}`,
    colors: { bg: 'rgba(59,130,246,0.85)', border: 'rgba(59,130,246,1)', style: 'rect' }
  },
  'Inventory': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.inventoryXAxis'),
    unit: t('analytics.isolation.anomalyTypes.inventoryUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(107,114,128,0.85)', border: 'rgba(107,114,128,1)', style: 'star' }
  },
  'Expense Report': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.expenseReportXAxis'),
    unit: t('analytics.isolation.anomalyTypes.expenseReportUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(234,179,8,0.85)', border: 'rgba(234,179,8,1)', style: 'rect' }
  },
  'Travel Expense': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.travelExpenseXAxis'),
    unit: t('analytics.isolation.anomalyTypes.travelExpenseUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(16,185,129,0.85)', border: 'rgba(16,185,129,1)', style: 'rectRot' }
  },
  'Procurement': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.procurementXAxis'),
    unit: t('analytics.isolation.anomalyTypes.procurementUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(236,72,153,0.85)', border: 'rgba(236,72,153,1)', style: 'triangle' }
  },
  'Transaction': {
    xAxisTitle: t('analytics.isolation.anomalyTypes.transactionXAxis'),
    unit: t('analytics.isolation.anomalyTypes.transactionUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(239,68,68,0.85)', border: 'rgba(239,68,68,1)', style: 'triangle' }
  }
}))

const selectedAnomalyType = ref('Funding')

const availableAnomalyTypes = computed(() => {
  const types = new Set<string>()
  const anomalies = isolationState.value.anomalies || []
  const scatterData = isolationState.value.scatterData || []

  anomalies.forEach((a: any) => { if (a.type) types.add(a.type) })
  scatterData.forEach((s: any) => { if (s.type) types.add(s.type) })

  const validBankTypes = ['Funding', 'Lending', 'Treasury', 'Payment', 'KYC', 'IT Control']
  const filtered = Array.from(types).filter(t => validBankTypes.includes(t))
  return filtered.length > 0 ? filtered : validBankTypes
})

const tableCategoryFilter = ref('All')

const filteredTableAnomalies = computed(() => {
  if (tableCategoryFilter.value === 'All') {
    return isolationState.value.anomalies
  }
  return isolationState.value.anomalies.filter((a: any) => a.type === tableCategoryFilter.value)
})

const getCategoryBadgeColor = (type: string) => {
  switch (type) {
    case 'Funding': return 'success'
    case 'Lending': return 'error'
    case 'Treasury': return 'primary'
    case 'Payment': return 'warning'
    case 'KYC': return 'secondary'
    case 'IT Control': return 'info'
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
      if (typeFilter === 'KYC') xVal = 25
      else if (typeFilter === 'IT Control') xVal = 4
      else xVal = a.amount ? a.amount / 1000000 : 15
    }
    return {
      x: xVal,
      y: matchedScatter?.y ?? 20
    }
  })

  const defaultConfig: AnomalyTypeConfig = {
    xAxisTitle: t('analytics.isolation.anomalyTypes.fundingXAxis'),
    unit: t('analytics.isolation.anomalyTypes.fundingUnit'),
    formatX: (val) => `Rp ${val}M`,
    colors: { bg: 'rgba(16,185,129,0.85)', border: 'rgba(16,185,129,1)', style: 'circle' }
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
            return `${metricName}: ${formattedX}, ${t('analytics.isolation.axisFrequency')}: ${ctx.parsed.y}`
          }
        }
      }
    },
    scales: {
      x: { title: { display: true, text: currentConfig.xAxisTitle } },
      y: { title: { display: true, text: t('analytics.isolation.axisFrequency') } }
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
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <UCard>
          <div class="text-center">
            <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.recordsScanned') }}</div>
            <div class="text-2xl font-black mt-1">{{ (isolationState.summary?.totalScanned || 0).toLocaleString() }}</div>
          </div>
        </UCard>
        <UCard>
          <div class="text-center">
            <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.anomaliesFound') }}</div>
            <div class="text-2xl font-black text-rose-500 mt-1">{{ isolationState.summary?.anomaliesFound || summary.anomaliesDetected || 150 }}</div>
          </div>
        </UCard>
        <UCard>
          <div class="text-center">
            <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.contaminationRate') }}</div>
            <div class="text-2xl font-black text-amber-500 mt-1">{{ formatNum((isolationState.summary?.contaminationRate || 0.125) * 100, 1) }}%</div>
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
      <UCard>
        <template #header>
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-table-cells" class="text-rose-500" />
              <h3 class="font-bold">{{ t('analytics.isolation.tableTitle') }}</h3>
            </div>
            <!-- Filter Kategori Tabel -->
            <div class="flex flex-wrap items-center gap-1.5">
              <span class="text-xs text-gray-400 font-medium mr-1">Filter:</span>
              <UButton
                size="xs"
                :color="tableCategoryFilter === 'All' ? 'primary' : 'neutral'"
                :variant="tableCategoryFilter === 'All' ? 'solid' : 'ghost'"
                @click="() => { tableCategoryFilter = 'All' }"
              >
                Semua ({{ isolationState.anomalies.length }})
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
              <tr class="border-b border-gray-200 dark:border-gray-700">
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
  </div>
</template>
