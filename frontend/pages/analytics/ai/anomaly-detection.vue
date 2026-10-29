<script setup lang="ts">
import { ref, computed, watch, watchEffect, onMounted } from 'vue'
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
  createAnomalyTypeConfigs,
  listAnomalyTypes,
  type AnomalyTypeConfig
} from '~/composables/useAnomalyScatter'
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

const anomalyTypeConfigs = computed<Record<string, AnomalyTypeConfig>>(() => createAnomalyTypeConfigs(t))

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

const mobileAnomalyPage = ref(1)
const mobileAnomalyPageSize = ref(10)

const paginatedMobileAnomalies = computed<any[]>(() => {
  const start = (mobileAnomalyPage.value - 1) * mobileAnomalyPageSize.value
  return filteredTableAnomalies.value.slice(start, start + mobileAnomalyPageSize.value)
})

watch(tableCategoryFilter, () => {
  mobileAnomalyPage.value = 1
})

const anomalyFilterOptions = computed(() => [
  {
    label: `${t('analytics.all')} (${isolationState.value.anomalies.length})`,
    value: 'All'
  },
  ...availableAnomalyTypes.value.map((cat: string) => ({
    label: `${cat} (${isolationState.value.anomalies.filter((a: any) => a.type === cat).length})`,
    value: cat
  }))
])

const getCategoryBadgeColor = (type: string) => {
  switch (type) {
    case 'Funding': return 'success'
    case 'Lending': return 'error'
    case 'Treasury': return 'primary'
    case 'Payment': return 'warning'
    case 'KYC': return 'secondary'
    case 'IT Control': return 'info'
    case 'Transaction': return 'error'
    case 'Procurement': return 'primary'
    case 'Expense Report': return 'warning'
    case 'Travel Expense': return 'success'
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

watch(selectedAnomalyType, (newType) => {
  if (newType) {
    tableCategoryFilter.value = newType
  }
})

const selectAnomalyType = (type: string): void => {
  selectedAnomalyType.value = type
}

const getScatterChartForType = (typeFilter: string) => {
  const { normalPoints, anomalyPoints } = buildAnomalyScatterPoints(
    isolationState.value.anomalies,
    isolationState.value.scatterData,
    typeFilter
  )

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
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
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
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-chart-bar" class="text-blue-500" />
              <h3 class="font-bold">{{ t('analytics.isolation.chartTitle') }}</h3>
            </div>
            
            <!-- Mobile & Tablet View (< md): Dropdown USelect -->
            <div class="block md:hidden w-full sm:w-auto">
              <USelect
                v-model="selectedAnomalyType"
                :items="availableAnomalyTypes"
                size="md"
                class="w-full sm:w-56"
              />
            </div>

            <!-- Desktop View (>= md): Anomaly Type Selector Buttons -->
            <div class="hidden md:flex flex-wrap gap-1">
              <UButton
                v-for="at in availableAnomalyTypes"
                :key="at"
                :color="selectedAnomalyType === at ? 'primary' : 'neutral'"
                :variant="selectedAnomalyType === at ? 'solid' : 'ghost'"
                size="md"
                @click="selectAnomalyType(at)"
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
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-table-cells" class="text-rose-500" />
              <h3 class="font-bold">{{ t('analytics.isolation.tableTitle') }}</h3>
            </div>
            
            <!-- Mobile & Tablet View (< md): Dropdown USelect -->
            <div class="block md:hidden w-full sm:w-auto">
              <USelect
                v-model="tableCategoryFilter"
                :items="anomalyFilterOptions"
                value-key="value"
                size="md"
                class="w-full sm:w-56"
              />
            </div>

            <!-- Desktop View (>= md): Interactive Filter Pills -->
            <div class="hidden md:flex items-center gap-1.5 flex-wrap">
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
        <!-- Desktop View (>= md): Full Table -->
        <div class="hidden md:block overflow-x-auto">
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

        <!-- Mobile & Tablet View (< md): Stacked Cards List -->
        <div class="block md:hidden">
          <!-- Empty State -->
          <div
            v-if="!filteredTableAnomalies.length"
            class="py-8 text-center border border-dashed border-gray-200 dark:border-gray-800 rounded-xl bg-[var(--bg-surface)]"
          >
            <UIcon name="i-heroicons-table-cells" class="size-8 mx-auto text-gray-400 mb-2" />
            <p class="text-xs text-[var(--text-muted)]">{{ t('analytics.isolation.empty') || 'Tidak ada data anomali' }}</p>
          </div>

          <!-- Cards List -->
          <div v-else class="space-y-3">
            <div
              v-for="(a, index) in paginatedMobileAnomalies"
              :key="a.id || index"
              class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
            >
              <!-- Card Header: ID & Type Badge + Risk Level Badge -->
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <span class="inline-flex items-center gap-1 font-mono text-[11px] font-bold text-slate-700 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded-md border border-slate-200/80 dark:border-slate-700">
                    <UIcon name="i-lucide-fingerprint" class="size-3 text-rose-500" />
                    {{ a.id }}
                  </span>
                  <UBadge :color="getCategoryBadgeColor(a.type)" variant="subtle" size="sm">
                    {{ a.type }}
                  </UBadge>
                </div>
                <UBadge
                  :class="getRiskConfig(a.riskLevel).badgeClass"
                  variant="solid"
                  size="sm"
                  class="font-bold"
                >
                  {{ getRiskConfig(a.riskLevel).label }}
                </UBadge>
              </div>

              <!-- Card Body: Entity & Description -->
              <div class="space-y-1">
                <div class="flex items-center gap-1.5 text-xs text-slate-500 dark:text-slate-400">
                  <UIcon name="i-lucide-building-2" class="size-3.5 text-slate-400 shrink-0" />
                  <span class="font-bold text-gray-900 dark:text-white text-sm truncate">{{ a.entity }}</span>
                </div>
                <p class="text-xs text-gray-600 dark:text-gray-300 leading-relaxed pt-1">
                  {{ a.description }}
                </p>
              </div>

              <!-- Card Footer: Index & Date -->
              <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800/80 text-xs text-slate-500 dark:text-slate-400">
                <span class="text-xs text-slate-400 font-mono">
                  #{{ (mobileAnomalyPage - 1) * mobileAnomalyPageSize + Number(index) + 1 }}
                </span>
                <div class="flex items-center gap-1.5">
                  <UIcon name="i-lucide-calendar" class="size-3.5 text-slate-400 shrink-0" />
                  <span>{{ a.date }}</span>
                </div>
              </div>
            </div>

            <!-- Mobile Pagination Controls -->
            <div
              v-if="filteredTableAnomalies.length > mobileAnomalyPageSize"
              class="flex flex-col sm:flex-row items-center justify-between gap-3 pt-3 border-t border-slate-100 dark:border-slate-800"
            >
              <span class="text-xs text-slate-500 dark:text-slate-400">
                {{ (mobileAnomalyPage - 1) * mobileAnomalyPageSize + 1 }}-{{ Math.min(mobileAnomalyPage * mobileAnomalyPageSize, filteredTableAnomalies.length) }} of {{ filteredTableAnomalies.length }}
              </span>
              <UPagination
                v-model:page="mobileAnomalyPage"
                :items-per-page="mobileAnomalyPageSize"
                :total="filteredTableAnomalies.length"
                size="md"
                active-color="primary"
                color="neutral"
                variant="outline"
                class="max-w-full overflow-x-auto"
              />
            </div>
          </div>
        </div>
      </UCard>
    </div>
  </div>
</template>
