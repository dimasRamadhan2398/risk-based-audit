<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
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
import {
  buildKpiForecastChart,
  horizonLabel,
  kpiAuditActionKey,
  selectAtRiskKpis,
  toKpiForecastView,
  type KpiForecastView
} from '~/composables/useKpiForecast'
import { formatKpiForecastValue, formatForecastPeriod } from '~/utils/analyticsFormat'
import AiInsightHeader from '~/components/analytics/AiInsightHeader.vue'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const { t, locale } = useI18n()
const {
  loading,
  timeseriesState,
  getRiskConfig,
  trendColor,
  trendIcon,
  fetchAiAnalytics
} = useAiAnalytics()

onMounted(() => {
  fetchAiAnalytics()
})

// Same view model as the dashboard card: units, direction-aware trend, horizon.
const kpiViews = computed<KpiForecastView[]>(() =>
  (timeseriesState.value.kpiForecasts || []).map(toKpiForecastView)
)
const atRiskKpis = computed(() => selectAtRiskKpis(kpiViews.value))

const selectedKpiName = ref('')
const selectedKpi = computed<KpiForecastView | undefined>(() =>
  kpiViews.value.find((k) => k.kpiName === selectedKpiName.value) ?? atRiskKpis.value[0] ?? kpiViews.value[0]
)
watch(selectedKpi, (kpi) => {
  if (kpi && kpi.kpiName !== selectedKpiName.value) selectedKpiName.value = kpi.kpiName
}, { immediate: true })
const kpiOptions = computed(() => kpiViews.value.map((k) => ({ label: k.kpiName, value: k.kpiName })))

const forecastMape = computed(() => Number(timeseriesState.value.forecastAccuracy?.mape ?? 0))
const mapeLabel = computed(() =>
  new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'id-ID', { maximumFractionDigits: 1 }).format(forecastMape.value)
)

const chart = computed(() => buildKpiForecastChart(selectedKpi.value, forecastMape.value, t, locale.value, {
  actualColor: 'rgb(59,130,246)',
  forecastColor: 'rgb(139,92,246)',
  bandColor: 'rgba(139,92,246,0.15)'
}))

const horizonOf = (kpi: KpiForecastView) => horizonLabel(kpi, locale.value)
const horizonValue = (kpi: KpiForecastView) =>
  kpi.forecast.length ? kpi.forecast[kpi.forecast.length - 1]!.value : kpi.forecastedValue

const trendText = (trend: string) => t(`dashboard.overview.trend.${String(trend).toLowerCase()}`)

const auditAction = (kpi: KpiForecastView) => {
  const key = kpiAuditActionKey(kpi.meta)
  if (kpi.trend !== 'Deteriorating') return kpi.recommendedAction || '-'
  if (key) return t(`analytics.kpiAudit.actions.${key}`)
  return kpi.recommendedAction || t('analytics.kpiAudit.actions.generic')
}
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <AiInsightHeader
      :title="t('navigation.kpiForecast')"
      :subtitle="t('analytics.subtitle')"
      icon="i-lucide-chart-line"
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
        icon="i-lucide-chart-line"
        color="info"
        variant="subtle"
        :title="t('analytics.timeseries.alertTitle')"
        :description="t('analytics.timeseries.alertDesc')"
      />

      <!-- Forecast Chart -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <UCard class="lg:col-span-2">
          <template #header>
            <div class="flex flex-wrap items-center gap-2">
              <UIcon name="i-heroicons-chart-bar" class="text-violet-500" />
              <h3 class="font-bold">{{ t('analytics.timeseries.chartTitle') }}</h3>
              <USelect
                v-if="kpiOptions.length"
                v-model="selectedKpiName"
                :items="kpiOptions"
                size="sm"
                class="w-60 ml-auto"
                :aria-label="t('analytics.timeseries.selectKpi')"
              />
            </div>
            <div v-if="selectedKpi" class="flex flex-wrap gap-x-4 gap-y-1 mt-2 text-xs text-gray-500">
              <span v-if="selectedKpi.lastActualPeriod">{{ t('analytics.timeseries.dataAsOf', { period: formatForecastPeriod(selectedKpi.lastActualPeriod, locale) }) }}</span>
              <span v-if="horizonOf(selectedKpi)">{{ t('analytics.timeseries.horizon', { range: horizonOf(selectedKpi) }) }}</span>
              <span>{{ t('analytics.timeseries.accuracy', { mape: mapeLabel }) }}</span>
              <span>{{ t('analytics.timeseries.chartConfidenceBadge') }}</span>
            </div>
          </template>
          <UAlert
            v-if="selectedKpi?.isStale"
            color="warning"
            variant="subtle"
            icon="i-lucide-clock-alert"
            :description="t('analytics.timeseries.staleWarning')"
            class="mb-3"
          />
          <div class="h-80">
            <Line :data="chart.data" :options="chart.options" />
          </div>
          <p v-if="selectedKpi?.meta.key === 'nps'" class="text-xs text-gray-400 mt-2">{{ t('analytics.kpiAudit.kindHint.nps') }}</p>
        </UCard>

        <!-- At-risk KPIs: only those projected to deteriorate -->
        <div class="space-y-6">
          <UCard>
            <template #header>
              <h3 class="font-bold text-sm">{{ t('analytics.timeseries.atRiskTitle') }}</h3>
            </template>
            <div class="space-y-3">
              <button
                v-for="kpi in atRiskKpis"
                :key="kpi.code"
                type="button"
                class="w-full text-left p-3 bg-rose-50 dark:bg-rose-900/20 rounded-lg border border-rose-200 dark:border-rose-800"
                @click="selectedKpiName = kpi.kpiName"
              >
                <div class="flex items-center justify-between gap-2">
                  <div class="font-bold text-md">{{ kpi.kpiName }}</div>
                  <UBadge
                    :class="getRiskConfig(kpi.riskLevel).badgeClass"
                    variant="solid"
                    size="md"
                    class="font-bold shrink-0"
                  >
                    {{ getRiskConfig(kpi.riskLevel).label }}
                  </UBadge>
                </div>
                <div class="text-[11px] text-gray-500 mt-0.5">{{ kpi.entity }}</div>
                <div class="flex items-center gap-2 mt-1.5 text-md">
                  <UIcon name="i-heroicons-arrow-trending-down" class="w-4 h-4 text-rose-500" />
                  <span class="font-mono text-gray-500">{{ formatKpiForecastValue(kpi.currentValue, kpi.meta, locale) }}</span>
                  <span class="text-gray-400">&rarr;</span>
                  <span class="font-mono font-bold text-rose-500">{{ formatKpiForecastValue(horizonValue(kpi), kpi.meta, locale) }}</span>
                  <span class="text-[11px] text-gray-400">{{ t('analytics.timeseries.projectedBy', { period: formatForecastPeriod(kpi.horizonEnd, locale) }) }}</span>
                </div>
                <p class="text-[11px] text-gray-600 dark:text-gray-300 mt-1.5 leading-relaxed">{{ auditAction(kpi) }}</p>
              </button>
              <p v-if="!atRiskKpis.length" class="text-xs text-gray-500">{{ t('analytics.timeseries.noAtRisk') }}</p>
            </div>
          </UCard>
        </div>
      </div>

      <!-- KPI Forecasts Table -->
      <UCard class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm ring-0 divide-gray-200 dark:divide-gray-800">
        <template #header>
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-table-cells" class="text-indigo-500" />
            <h3 class="font-bold">{{ t('analytics.timeseries.tableTitle') }}</h3>
          </div>
        </template>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-100 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800">
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
              <tr v-for="kpi in kpiViews" :key="kpi.code" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                <td class="py-3 px-3 font-mono font-bold text-md">{{ kpi.code }}</td>
                <td class="py-3 px-3 font-bold text-md">{{ kpi.kpiName }}</td>
                <td class="py-3 px-3 font-bold text-md">{{ kpi.entity || '-' }}</td>
                <td class="text-center py-3 px-3"><UBadge :color="kpi.entityType === 'Branch' ? 'primary' : 'warning'" variant="subtle" size="md">{{ kpi.entityType || '-' }}</UBadge></td>
                <td class="text-center py-3 px-3"><UBadge color="info" variant="subtle" size="md" class="font-bold whitespace-nowrap">{{ horizonOf(kpi) || '-' }}</UBadge></td>
                <td class="text-center py-3 px-3 font-mono whitespace-nowrap">{{ formatKpiForecastValue(kpi.currentValue, kpi.meta, locale) }}</td>
                <td class="text-center py-3 px-3 font-mono font-bold whitespace-nowrap">{{ formatKpiForecastValue(horizonValue(kpi), kpi.meta, locale) }}</td>
                <td class="text-center py-3 px-3">
                  <div class="flex items-center justify-center gap-1">
                    <UIcon :name="trendIcon(kpi.trend)" class="w-4 h-4" :class="trendColor(kpi.trend)" />
                    <span class="text-md font-bold" :class="trendColor(kpi.trend)">{{ trendText(kpi.trend) }}</span>
                  </div>
                </td>
                <td class="text-center py-3 px-3">
                  <UBadge
                    :class="getRiskConfig(kpi.riskLevel).badgeClass"
                    variant="solid"
                    size="md"
                    class="font-bold"
                  >
                    {{ getRiskConfig(kpi.riskLevel).label }}
                  </UBadge>
                </td>
                <td class="py-3 px-3 text-md text-gray-600 dark:text-gray-300 leading-relaxed">{{ auditAction(kpi) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </UCard>
    </div>
  </div>
</template>
