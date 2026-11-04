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

const mobilePage = ref(1)
const mobilePageSize = ref(10)

const paginatedKpiForecasts = computed<KpiForecastView[]>(() => {
  const start = (mobilePage.value - 1) * mobilePageSize.value
  return kpiViews.value.slice(start, start + mobilePageSize.value)
})
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
        <!-- Desktop View (>= md): Full Table -->
        <div class="hidden md:block overflow-x-auto">
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

        <!-- Mobile & Tablet View (< md): Stacked Cards List -->
        <div class="block md:hidden">
          <!-- Empty State -->
          <div
            v-if="!kpiViews.length"
            class="py-8 text-center border border-dashed border-gray-200 dark:border-gray-800 rounded-xl bg-[var(--bg-surface)]"
          >
            <UIcon name="i-heroicons-table-cells" class="size-8 mx-auto text-gray-400 mb-2" />
            <p class="text-xs text-[var(--text-muted)]">{{ t('analytics.timeseries.empty') || 'Tidak ada data perkiraan KPI' }}</p>
          </div>

          <!-- Cards List -->
          <div v-else class="space-y-3">
            <div
              v-for="(kpi, index) in paginatedKpiForecasts"
              :key="kpi.code || index"
              class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
            >
              <!-- Card Header: Code & Risk Level Badge -->
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <span class="inline-flex items-center gap-1 font-mono text-[11px] font-bold text-slate-700 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded-md border border-slate-200/80 dark:border-slate-700">
                    <UIcon name="i-heroicons-chart-bar" class="size-3 text-indigo-500" />
                    {{ kpi.code }}
                  </span>
                  <UBadge :color="kpi.entityType === 'Branch' ? 'primary' : 'warning'" variant="subtle" size="sm">
                    {{ kpi.entityType || '-' }}
                  </UBadge>
                  <UBadge color="info" variant="subtle" size="sm" class="font-bold">
                    {{ horizonOf(kpi) || '-' }}
                  </UBadge>
                </div>
                <UBadge
                  :class="getRiskConfig(kpi.riskLevel).badgeClass"
                  variant="solid"
                  size="sm"
                  class="font-bold shrink-0"
                >
                  {{ getRiskConfig(kpi.riskLevel).label }}
                </UBadge>
              </div>

              <!-- Card Body: KPI Name & Entity -->
              <div class="space-y-1">
                <div class="font-bold text-sm text-gray-900 dark:text-white leading-snug">
                  {{ kpi.kpiName }}
                </div>
                <div class="flex items-center gap-1.5 text-xs text-slate-500 dark:text-slate-400">
                  <UIcon name="i-lucide-building-2" class="size-3.5 text-slate-400 shrink-0" />
                  <span>{{ kpi.entity || '-' }}</span>
                </div>
              </div>

              <!-- Comparison Grid: Current vs Forecast -->
              <div class="grid grid-cols-2 gap-2 pt-1">
                <div class="p-2.5 rounded-lg bg-gray-50 dark:bg-gray-800/60 border border-gray-100 dark:border-gray-800 space-y-1">
                  <span class="text-[10px] uppercase font-bold tracking-wider text-gray-400 block truncate text-center">{{ t('analytics.timeseries.colCurrent') }}</span>
                  <div class="font-mono font-bold text-sm text-neutral-900 dark:text-neutral-100 text-center">
                    {{ formatKpiForecastValue(kpi.currentValue, kpi.meta, locale) }}
                  </div>
                </div>
                <div class="p-2.5 rounded-lg bg-indigo-50/50 dark:bg-indigo-950/20 border border-indigo-100 dark:border-indigo-900/40 space-y-1">
                  <span class="text-[10px] uppercase font-bold tracking-wider text-secondary-700 dark:text-secondary-400 block truncate text-center">{{ t('analytics.timeseries.colForecast') }}</span>
                  <div class="font-mono font-bold text-sm text-secondary-700 dark:text-secondary-400 text-center">
                    {{ formatKpiForecastValue(horizonValue(kpi), kpi.meta, locale) }}
                  </div>
                </div>
              </div>

              <!-- Trend Row -->
              <div class="flex items-center justify-between px-2.5 py-1.5 bg-gray-50/80 dark:bg-gray-800/40 rounded-md border border-gray-100 dark:border-gray-800 text-xs">
                <span class="text-[11px] text-gray-500 dark:text-gray-400 font-medium">{{ t('analytics.timeseries.colTrend') }}</span>
                <div class="flex items-center gap-1.5 font-bold" :class="trendColor(kpi.trend)">
                  <UIcon :name="trendIcon(kpi.trend)" class="w-4 h-4 shrink-0" />
                  <span>{{ trendText(kpi.trend) }}</span>
                </div>
              </div>

              <!-- Recommended Action -->
              <div class="p-2.5 rounded-lg bg-amber-50/60 dark:bg-amber-950/20 border border-amber-200/60 dark:border-amber-800/40 text-xs text-amber-900 dark:text-amber-300 leading-relaxed">
                <span class="font-bold text-amber-950 dark:text-amber-200">{{ t('analytics.timeseries.colRecommendedAction') }}: </span>
                <span>{{ auditAction(kpi) }}</span>
              </div>

              <!-- Card Footer: Index & Horizon -->
              <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800/80 text-xs text-slate-500 dark:text-slate-400">
                <span class="text-xs text-slate-400 font-mono">
                  #{{ (mobilePage - 1) * mobilePageSize + Number(index) + 1 }}
                </span>
                <span class="text-[11px] text-gray-400">{{ horizonOf(kpi) }}</span>
              </div>
            </div>

            <!-- Mobile Pagination Controls -->
            <div
              v-if="kpiViews.length > mobilePageSize"
              class="flex flex-col sm:flex-row items-center justify-between gap-3 pt-3 border-t border-slate-100 dark:border-slate-800"
            >
              <span class="text-xs text-slate-500 dark:text-slate-400">
                {{ (mobilePage - 1) * mobilePageSize + 1 }}-{{ Math.min(mobilePage * mobilePageSize, kpiViews.length) }} of {{ kpiViews.length }}
              </span>
              <UPagination
                v-model:page="mobilePage"
                :items-per-page="mobilePageSize"
                :total="kpiViews.length"
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
