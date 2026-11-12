<script setup lang="ts">
import { ref, computed, watch, watchEffect, onMounted, nextTick } from 'vue'
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
  getAnomalyTypeConfig,
  listAnomalyTypes,
  isCategoryMatching,
  getCanonicalCategory,
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
  automatedAnomalyRules,
  trainAnomalyWithRules,
  formatNum,
  getRiskConfig,
  fetchAiAnalytics
} = useAiAnalytics()

const isRetrainingRules = ref(false)

async function reSyncAllAutomatedRules() {
  if (!automatedAnomalyRules.value || automatedAnomalyRules.value.length === 0) return
  isRetrainingRules.value = true
  try {
    const rulesByTable: Record<string, any[]> = {}
    automatedAnomalyRules.value.forEach((r: any) => {
      const tbl = r.table_name || 'gl_transactions'
      if (!rulesByTable[tbl]) rulesByTable[tbl] = []
      rulesByTable[tbl].push(r)
    })

    for (const [tbl, rules] of Object.entries(rulesByTable)) {
      await trainAnomalyWithRules({
        table_name: tbl,
        connection_name: rules[0]?.connection_name || 'Core Banking',
        category: rules[0]?.category || 'Transaction',
        rules
      })
    }
  } finally {
    isRetrainingRules.value = false
  }
}

onMounted(() => {
  fetchAiAnalytics()
})

const anomalyTypeConfigs = computed<Record<string, AnomalyTypeConfig>>(() => createAnomalyTypeConfigs(t, locale.value))

const selectedAnomalyType = ref('Funding')
const categorySearch = ref('')

// Interactive point-to-table marker highlighting
const highlightedAnomalyId = ref<string | null>(null)
const highlightedPointInfo = ref<{
  id: string
  entity: string
  type: string
  metric: string
  riskLevel: string
} | null>(null)

const availableAnomalyTypes = computed(() =>
  listAnomalyTypes(
    isolationState.value.anomalies,
    isolationState.value.scatterData,
    Object.keys(anomalyTypeConfigs.value)
  )
)

const filteredAvailableAnomalyTypes = computed(() => {
  if (!categorySearch.value.trim()) return availableAnomalyTypes.value
  const q = categorySearch.value.toLowerCase().trim()
  return availableAnomalyTypes.value.filter(cat => cat.toLowerCase().includes(q))
})

const tableCategoryFilter = ref('All')
const tableSearch = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const pageSizeSelectItems = [
  { label: '10', value: 10 },
  { label: '25', value: 25 },
  { label: '50', value: 50 },
  { label: '100', value: 100 },
  { label: 'All', value: 999999 }
]

const filteredTableAnomalies = computed(() => {
  let list = isolationState.value.anomalies || []
  if (tableCategoryFilter.value !== 'All') {
    list = list.filter((a: any) => isCategoryMatching(a.type, tableCategoryFilter.value))
  }
  if (tableSearch.value.trim()) {
    const q = tableSearch.value.toLowerCase().trim()
    list = list.filter((a: any) =>
      (a.id && String(a.id).toLowerCase().includes(q)) ||
      (a.entity && String(a.entity).toLowerCase().includes(q)) ||
      (a.type && String(a.type).toLowerCase().includes(q)) ||
      (a.description && String(a.description).toLowerCase().includes(q))
    )
  }
  return list
})

const totalAnomalies = computed(() => filteredTableAnomalies.value.length)
const totalPages = computed(() => Math.ceil(totalAnomalies.value / pageSize.value) || 1)

const paginatedAnomalies = computed(() => {
  if (pageSize.value >= 999999) return filteredTableAnomalies.value
  const start = (currentPage.value - 1) * pageSize.value
  return filteredTableAnomalies.value.slice(start, start + pageSize.value)
})

const fromEntry = computed(() => (totalAnomalies.value === 0 ? 0 : (currentPage.value - 1) * pageSize.value + 1))
const toEntry = computed(() => Math.min(currentPage.value * pageSize.value, totalAnomalies.value))

watch([tableSearch, pageSize], () => {
  currentPage.value = 1
})

watch(tableCategoryFilter, (newCat) => {
  currentPage.value = 1
  if (newCat && newCat !== 'All' && availableAnomalyTypes.value.includes(newCat)) {
    selectedAnomalyType.value = newCat
  }
})

watch(selectedAnomalyType, (newType) => {
  if (newType && tableCategoryFilter.value !== 'All' && tableCategoryFilter.value !== newType) {
    tableCategoryFilter.value = newType
  }
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
    default: {
      const badgeColors = ['primary', 'secondary', 'success', 'warning', 'error', 'info']
      let hash = 0
      for (let i = 0; i < (type || '').length; i++) hash = (hash << 5) - hash + type.charCodeAt(i)
      return badgeColors[Math.abs(hash) % badgeColors.length] as any
    }
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

const formatMetricDisplay = (a: any): string => {
  const type = a.type || selectedAnomalyType.value
  const config = anomalyTypeConfigs.value[type] || getAnomalyTypeConfig(t, type)
  const val = a.xMetric ?? (a.amount ? a.amount / 1000000 : 0)
  return config.formatX ? config.formatX(val) : `${val} ${config.unit || ''}`
}

const selectAnomalyFromPlot = (id: string, type?: string) => {
  highlightedAnomalyId.value = id
  if (type && availableAnomalyTypes.value.includes(type)) {
    selectedAnomalyType.value = type
  }
  if (type && tableCategoryFilter.value !== 'All' && tableCategoryFilter.value !== type) {
    tableCategoryFilter.value = type
  }

  const allAnoms = isolationState.value.anomalies || []
  const found = allAnoms.find((a: any) => a.id === id)
  if (found) {
    highlightedPointInfo.value = {
      id: found.id,
      entity: found.entity || found.accountNumber || 'Kantor Pusat / GL',
      type: found.type,
      metric: formatMetricDisplay(found),
      riskLevel: found.riskLevel || 'High'
    }

    // Adjust page if necessary
    const filteredList = filteredTableAnomalies.value
    const idx = filteredList.findIndex((a: any) => a.id === id)
    if (idx >= 0 && pageSize.value < 999999) {
      currentPage.value = Math.floor(idx / pageSize.value) + 1
    }
  }

  nextTick(() => {
    const el = document.getElementById(`anomaly-row-${id}`)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  })
}

const focusOnPlot = (a: any) => {
  highlightedAnomalyId.value = a.id
  if (a.type && availableAnomalyTypes.value.includes(a.type)) {
    selectedAnomalyType.value = a.type
  }
  highlightedPointInfo.value = {
    id: a.id,
    entity: a.entity || a.accountNumber || 'Kantor Pusat / GL',
    type: a.type,
    metric: formatMetricDisplay(a),
    riskLevel: a.riskLevel || 'High'
  }
  nextTick(() => {
    const el = document.getElementById('anomaly-scatter-chart-card')
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }
  })
}

const scrollToTableRow = (id: string) => {
  const filteredList = filteredTableAnomalies.value
  const idx = filteredList.findIndex((a: any) => a.id === id)
  if (idx >= 0 && pageSize.value < 999999) {
    currentPage.value = Math.floor(idx / pageSize.value) + 1
  }
  nextTick(() => {
    const el = document.getElementById(`anomaly-row-${id}`)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  })
}

const clearHighlight = () => {
  highlightedAnomalyId.value = null
  highlightedPointInfo.value = null
}

const getPlottedAnomaliesForType = (typeFilter: string) => {
  const allScatter = isolationState.value.scatterData || []
  return (isolationState.value.anomalies || [])
    .filter((a: any) => isCategoryMatching(a.type, typeFilter))
    .filter((a: any) => {
      const matchedScatter = allScatter.find((s: any) => s.label === a.id)
      const yVal = matchedScatter?.y ?? a.hour ?? a.hour_of_day ?? a.y
      return yVal !== undefined && yVal !== null
    })
}

const getScatterChartForType = (typeFilter: string) => {
  const { normalPoints, anomalyPoints } = buildAnomalyScatterPoints(
    isolationState.value.anomalies,
    isolationState.value.scatterData,
    typeFilter
  )
  const plottedAnomalies = getPlottedAnomaliesForType(typeFilter)

  const config = anomalyTypeConfigs.value[typeFilter] || getAnomalyTypeConfig(t, typeFilter)
  const colors = config.colors

  const isSelected = (idx: number) => {
    const anom = plottedAnomalies[idx]
    return !!(anom && highlightedAnomalyId.value && anom.id === highlightedAnomalyId.value)
  }

  return {
    datasets: [
      {
        label: t('analytics.isolation.labelNormalData', { type: typeFilter }),
        data: normalPoints,
        backgroundColor: 'rgba(59,130,246,0.3)',
        borderColor: 'rgba(59,130,246,0.6)',
        pointRadius: 4,
        pointHoverRadius: 6,
      },
      {
        label: t('analytics.isolation.labelAnomalies', { type: typeFilter }),
        data: anomalyPoints,
        backgroundColor: anomalyPoints.map((_, idx) => isSelected(idx) ? '#FF0055' : colors.bg),
        borderColor: anomalyPoints.map((_, idx) => isSelected(idx) ? '#E11D48' : colors.border),
        pointRadius: anomalyPoints.map((_, idx) => isSelected(idx) ? 14 : 7),
        pointHoverRadius: 13,
        pointBorderWidth: anomalyPoints.map((_, idx) => isSelected(idx) ? 4 : 1.5),
        pointStyle: colors.style,
      }
    ]
  }
}

const scatterChartData = computed(() => getScatterChartForType(selectedAnomalyType.value))

const scatterOptions = computed(() => {
  const currentConfig = anomalyTypeConfigs.value[selectedAnomalyType.value] || getAnomalyTypeConfig(t, selectedAnomalyType.value)

  return {
    responsive: true,
    maintainAspectRatio: false,
    onClick: (_event: any, elements: any[]) => {
      if (!elements || elements.length === 0) return
      const firstElem = elements[0]
      if (firstElem.datasetIndex === 1) {
        const plottedAnomalies = getPlottedAnomaliesForType(selectedAnomalyType.value)
        const anom = plottedAnomalies[firstElem.index]
        if (anom && anom.id) {
          selectAnomalyFromPlot(anom.id, anom.type || selectedAnomalyType.value)
        }
      }
    },
    plugins: {
      legend: { position: 'top' as const },
      tooltip: {
        callbacks: {
          title: (items: any[]) => {
            const item = items[0]
            if (item && item.datasetIndex === 1) {
              const plottedAnomalies = getPlottedAnomaliesForType(selectedAnomalyType.value)
              const anom = plottedAnomalies[item.dataIndex]
              if (anom) return `[${anom.id}] ${anom.entity || ''}`
            }
            return item?.dataset?.label || ''
          },
          label: (ctx: any) => {
            const formattedX = currentConfig.formatX ? currentConfig.formatX(ctx.parsed.x) : `${ctx.parsed.x}`
            const metricName = currentConfig.xAxisTitle ? currentConfig.xAxisTitle.split(' (')[0] : selectedAnomalyType.value
            if (ctx.datasetIndex === 1) {
              const plottedAnomalies = getPlottedAnomaliesForType(selectedAnomalyType.value)
              const anom = plottedAnomalies[ctx.dataIndex]
              return [
                `Kategori: ${anom?.type || selectedAnomalyType.value}`,
                `Metrik (${metricName}): ${formattedX}`,
                `${t('analytics.isolation.axisFrequency')}: ${ctx.parsed.y}`,
                `Tingkat Risiko: ${anom?.riskLevel || 'High'}`,
                `👉 Klik titik ini untuk menyorot di Tabel`
              ]
            }
            return [
              `Data Normal (Baseline Real)`,
              `Metrik (${metricName}): ${formattedX}`,
              `${t('analytics.isolation.axisFrequency')}: ${ctx.parsed.y}`
            ]
          }
        }
      }
    },
    scales: {
      x: {
        title: { display: true, text: currentConfig.xAxisTitle },
        ticks: { maxTicksLimit: 6, callback: (v: any) => currentConfig.formatX ? currentConfig.formatX(Number(v)) : v }
      },
      // Y = transaction hour from the data (0-24).
      y: { suggestedMin: 0, suggestedMax: 24, ticks: { stepSize: 3 }, title: { display: true, text: t('analytics.isolation.axisHour') } }
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

      <!-- Automated Anomaly Rules (from Data Source Pipeline) Banner -->
      <UCard v-if="automatedAnomalyRules && automatedAnomalyRules.length > 0" class="border border-primary-200/60 dark:border-primary-900/40 bg-gradient-to-r from-primary-50/40 via-white to-primary-50/20 dark:from-primary-950/20 dark:via-gray-900 dark:to-primary-950/10">
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-3">
          <div class="space-y-1">
            <div class="flex items-center gap-2">
              <UIcon name="i-lucide-sparkles" class="text-primary-500 w-5 h-5 shrink-0" />
              <h4 class="font-bold text-sm text-gray-900 dark:text-white">
                Automated Anomaly Rules (Terkoneksi dari Data Source)
              </h4>
              <UBadge color="primary" variant="solid" size="xs">
                {{ automatedAnomalyRules.length }} Aturan Aktif
              </UBadge>
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              Model AI Anomaly Detection (Isolation Forest & XGBoost) secara otomatis dilatih berdasarkan kriteria dan aturan anomali berikut:
            </p>
          </div>
          <div class="flex items-center gap-2 shrink-0">
            <UButton
              size="xs"
              color="primary"
              variant="subtle"
              icon="i-lucide-refresh-cw"
              :loading="isRetrainingRules"
              class="font-semibold rounded-lg"
              @click="reSyncAllAutomatedRules"
            >
              Latih Ulang Model AI
            </UButton>
          </div>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2.5 mt-3 pt-3 border-t border-primary-100 dark:border-primary-900/40">
          <div
            v-for="rule in automatedAnomalyRules"
            :key="rule.id || rule.name"
            class="flex items-start justify-between p-2.5 rounded-xl border bg-white dark:bg-gray-800/80 border-gray-200 dark:border-gray-700/70 shadow-xs"
          >
            <div class="space-y-1 overflow-hidden pr-2">
              <div class="flex items-center gap-1.5 flex-wrap">
                <UBadge
                  :color="rule.severity === 'CRITICAL' ? 'error' : rule.severity === 'HIGH' ? 'warning' : 'neutral'"
                  size="xs"
                  variant="subtle"
                  class="font-mono text-[9px]"
                >
                  {{ rule.severity || 'HIGH' }}
                </UBadge>
                <span class="text-xs font-bold text-gray-900 dark:text-white truncate">{{ rule.name }}</span>
              </div>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 line-clamp-1">
                {{ rule.description || (rule.field ? `${rule.field} ${rule.operator} ${rule.threshold}` : 'Aturan operasional standar') }}
              </p>
              <div class="flex items-center gap-2 text-[10px] text-gray-400 font-mono">
                <span v-if="rule.table_name" class="flex items-center gap-1">
                  <UIcon name="i-lucide-database" class="w-3 h-3 text-primary-500" />
                  {{ rule.table_name }}
                </span>
                <span v-if="rule.category" class="bg-gray-100 dark:bg-gray-700 px-1.5 py-0.5 rounded">
                  {{ rule.category }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </UCard>

      <!-- Scatter Chart Card with Type Filter Tabs -->
      <UCard id="anomaly-scatter-chart-card">
        <template #header>
          <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3">
            <div class="flex items-center gap-2 shrink-0">
              <UIcon name="i-heroicons-chart-bar" class="text-blue-500" />
              <h3 class="font-bold">{{ t('analytics.isolation.chartTitle') }}</h3>
            </div>
            <!-- Anomaly Type Selector with Search and Scroll -->
            <div class="flex items-center gap-2 overflow-hidden max-w-full">
              <div class="flex items-center gap-1.5 overflow-x-auto pb-1 max-w-full scrollbar-thin">
                <UButton
                  v-for="at in filteredAvailableAnomalyTypes"
                  :key="at"
                  :color="selectedAnomalyType === at ? 'primary' : 'neutral'"
                  :variant="selectedAnomalyType === at ? 'solid' : 'ghost'"
                  size="xs"
                  class="whitespace-nowrap shrink-0"
                  @click="() => { selectedAnomalyType = at; tableCategoryFilter = at; }"
                >
                  {{ at }}
                </UButton>
              </div>
            </div>
          </div>
        </template>

        <!-- Interactive Selection Marker Pill / Banner -->
        <div
          v-if="highlightedPointInfo"
          class="mb-3 flex flex-wrap items-center justify-between gap-2 px-3 py-2 bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-900 rounded-lg text-xs"
        >
          <div class="flex flex-wrap items-center gap-2">
            <UIcon name="i-heroicons-cursor-arrow-rays" class="w-4 h-4 text-rose-500 animate-pulse shrink-0" />
            <span class="font-bold text-rose-700 dark:text-rose-300">
              {{ t('analytics.isolation.selectedPointBanner') }}: {{ highlightedPointInfo.id }}
            </span>
            <span class="text-gray-400">|</span>
            <span class="text-gray-700 dark:text-gray-200 font-semibold">{{ highlightedPointInfo.entity }}</span>
            <span class="text-gray-400">|</span>
            <span class="text-rose-600 dark:text-rose-400 font-mono font-medium">{{ highlightedPointInfo.metric }}</span>
            <UBadge :color="getCategoryBadgeColor(highlightedPointInfo.type)" variant="subtle" size="xs">
              {{ highlightedPointInfo.type }}
            </UBadge>
          </div>
          <div class="flex items-center gap-2 shrink-0">
            <UButton
              size="xs"
              color="error"
              variant="soft"
              icon="i-heroicons-arrow-down-circle"
              @click="scrollToTableRow(highlightedPointInfo.id)"
            >
              {{ t('analytics.isolation.viewInTable') }}
            </UButton>
            <UButton
              size="xs"
              color="neutral"
              variant="ghost"
              icon="i-heroicons-x-mark"
              @click="clearHighlight"
            >
              {{ t('analytics.isolation.clearSelection') }}
            </UButton>
          </div>
        </div>

        <div class="h-80">
          <Scatter
            :key="'scatter-chart-' + selectedAnomalyType + '-' + (highlightedAnomalyId || 'none')"
            :data="scatterChartData"
            :options="scatterOptions"
          />
        </div>
      </UCard>

      <!-- Anomalies Table -->
      <UCard class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm ring-0 divide-gray-200 dark:divide-gray-800">
        <template #header>
          <div class="flex flex-col gap-3">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div class="flex items-center gap-2">
                <UIcon name="i-heroicons-table-cells" class="text-rose-500" />
                <h3 class="font-bold">{{ t('analytics.isolation.tableTitle') }}</h3>
              </div>
              <!-- Table Search Box -->
              <div class="w-full sm:w-72">
                <UInput
                  v-model="tableSearch"
                  icon="i-heroicons-magnifying-glass"
                  :placeholder="t('analytics.isolation.searchAnomalies')"
                  size="sm"
                  class="w-full"
                />
              </div>
            </div>
            <!-- Filter Kategori Tabel (Scrollable Horizontal Pill Bar) -->
            <div class="flex items-center gap-1.5 overflow-x-auto pb-1 max-w-full scrollbar-thin">
              <span class="text-xs text-gray-400 font-medium mr-1 shrink-0">{{ t('analytics.anomalyFilter') }}</span>
              <UButton
                size="xs"
                :color="tableCategoryFilter === 'All' ? 'primary' : 'neutral'"
                :variant="tableCategoryFilter === 'All' ? 'solid' : 'ghost'"
                class="whitespace-nowrap shrink-0"
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
                class="whitespace-nowrap shrink-0"
                @click="() => { tableCategoryFilter = cat; selectedAnomalyType = cat; }"
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
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.colAction') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="paginatedAnomalies.length === 0">
                <td colspan="7" class="py-8 text-center text-gray-400 italic">
                  {{ t('analytics.isolation.noAnomaliesFound') }}
                </td>
              </tr>
              <tr
                v-for="a in paginatedAnomalies"
                :id="'anomaly-row-' + a.id"
                :key="a.id"
                :class="[
                  'border-b border-gray-100 dark:border-gray-800 transition-all duration-300 cursor-pointer',
                  highlightedAnomalyId === a.id
                    ? 'ring-2 ring-rose-500 bg-rose-50/90 dark:bg-rose-950/50 shadow-md font-semibold'
                    : 'hover:bg-gray-50 dark:hover:bg-gray-800/50'
                ]"
                @click="focusOnPlot(a)"
              >
                <td class="py-3 px-3">
                  <div class="flex items-center gap-1.5">
                    <span class="font-mono font-bold text-md">{{ a.id }}</span>
                    <UBadge
                      v-if="highlightedAnomalyId === a.id"
                      color="error"
                      variant="solid"
                      size="xs"
                      class="animate-pulse shrink-0"
                    >
                      🎯 {{ t('analytics.isolation.selectedOnPlot') }}
                    </UBadge>
                  </div>
                </td>
                <td class="py-3 px-3 font-bold">{{ a.entity || a.accountNumber || a.refNumber || 'Kantor Pusat / GL' }}</td>
                <td class="py-3 px-3">
                  <UBadge :color="getCategoryBadgeColor(a.type)" variant="subtle" size="md">{{ a.type }}</UBadge>
                </td>
                <td class="py-3 px-3 text-md leading-relaxed text-gray-600 dark:text-gray-300">
                  <div class="flex items-center gap-1.5 flex-wrap">
                    <UBadge
                      v-if="a.id?.startsWith('RULE-') || a.description?.includes('[Rule:') || a.reason?.includes('[Rule:')"
                      color="primary"
                      variant="subtle"
                      size="xs"
                      class="font-mono text-[10px]"
                    >
                      ⚡ Rule Match
                    </UBadge>
                    <span>{{ a.description || a.reason || a.notes || 'Transaksi Anomali' }}</span>
                  </div>
                </td>
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
                <td class="text-center py-3 px-3 text-md text-gray-400">{{ a.date || (a.timestamp ? a.timestamp.slice(0, 10) : '2026-06-01') }}</td>
                <td class="text-center py-3 px-3">
                  <UButton
                    size="xs"
                    :color="highlightedAnomalyId === a.id ? 'error' : 'neutral'"
                    :variant="highlightedAnomalyId === a.id ? 'solid' : 'ghost'"
                    icon="i-heroicons-arrow-trending-up"
                    class="whitespace-nowrap"
                    @click.stop="focusOnPlot(a)"
                  >
                    {{ t('analytics.isolation.viewOnPlot') }}
                  </UButton>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <!-- Table Pagination Footer -->
        <template #footer>
          <div class="flex flex-col sm:flex-row items-center justify-between gap-3 px-2 py-1 text-xs text-gray-500">
            <div>
              {{ t('analytics.isolation.showingEntries', { from: fromEntry, to: toEntry, total: totalAnomalies }) }}
            </div>
            <div class="flex flex-wrap items-center gap-3">
              <div class="flex items-center gap-1.5">
                <span>{{ t('analytics.isolation.rowsPerPage') }}</span>
                <USelect
                  v-model="pageSize"
                  :items="pageSizeSelectItems"
                  size="xs"
                  class="w-20"
                />
              </div>
              <UPagination
                v-if="totalPages > 1"
                v-model:page="currentPage"
                :items-per-page="pageSize"
                :total="totalAnomalies"
                :show-controls="true"
                :show-edges="true"
                size="xs"
                active-color="primary"
                color="neutral"
                variant="outline"
              />
            </div>
          </div>
        </template>
      </UCard>
    </div>
  </div>
</template>
