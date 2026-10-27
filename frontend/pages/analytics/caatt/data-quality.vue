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
import { useCaattAnalytics } from '~/composables/useCaattAnalytics'
import { useI18n } from '~/composables/useI18n'
import CaattHeader from '~/components/analytics/CaattHeader.vue'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend)

const { t } = useI18n()
const { loading, dqState, fetchCaattAnalytics } = useCaattAnalytics()

onMounted(() => {
  fetchCaattAnalytics()
})

const formatFallbackName = (raw: string): string => {
  if (!raw) return '-'
  return raw
    .replace(/^(gold|silver|bronze|public|dbo)\./i, '')
    .replace(/^(fact_|dim_|cbs_|cb_)/i, '')
    .replace(/[._-]+/g, ' ')
    .trim()
    .split(' ')
    .map(w => w.charAt(0).toUpperCase() + w.slice(1).toLowerCase())
    .join(' ')
}

const getTableFriendlyName = (rawName: string): string => {
  if (!rawName) return '-'
  const i18nKey = `analytics.caattUi.dataQuality.tables.${rawName}`
  const translated = t(i18nKey)
  if (translated && translated !== i18nKey) {
    return translated
  }
  return formatFallbackName(rawName)
}

const getTableChartName = (rawName: string): string => {
  if (!rawName) return '-'
  const i18nKey = `analytics.caattUi.dataQuality.chartTables.${rawName}`
  const translated = t(i18nKey)
  if (translated && translated !== i18nKey) {
    return translated
  }
  return getTableFriendlyName(rawName)
}

const getTableIcon = (tableName: string): string => {
  if (!tableName) return 'i-heroicons-table-cells'
  if (tableName.includes('transactions')) return 'i-heroicons-banknotes'
  if (tableName.includes('loans')) return 'i-heroicons-credit-card'
  if (tableName.includes('accounts')) return 'i-heroicons-wallet'
  if (tableName.includes('customers')) return 'i-heroicons-user-group'
  if (tableName.includes('gl_entries') || tableName.includes('gl')) return 'i-heroicons-book-open'
  if (tableName.includes('branches')) return 'i-heroicons-building-office-2'
  if (tableName.includes('products')) return 'i-heroicons-cube'
  if (tableName.includes('failed_logins')) return 'i-heroicons-shield-exclamation'
  return 'i-heroicons-table-cells'
}

const dqChartData = computed(() => ({
  labels: dqState.value.records.map((r: any) => getTableChartName(r.tableName)),
  datasets: [
    {
      label: t('analytics.caattUi.dataQuality.completenessDataset'),
      backgroundColor: 'rgba(59,130,246,0.75)',
      borderRadius: 4,
      data: dqState.value.records.map((r: any) => r.completenessPct),
    },
    {
      label: t('analytics.caattUi.dataQuality.accuracyDataset'),
      backgroundColor: 'rgba(16,185,129,0.75)',
      borderRadius: 4,
      data: dqState.value.records.map((r: any) => r.accuracyPct),
    },
  ],
}))

const dqChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'top' as const },
    tooltip: {
      callbacks: {
        title: (items: any[]) => {
          if (!items.length) return ''
          const index = items[0].dataIndex
          const rec = dqState.value.records[index]
          if (rec) {
            return `${getTableFriendlyName(rec.tableName)} (${rec.tableName})`
          }
          return items[0].label
        }
      }
    }
  },
  scales: {
    x: {
      ticks: {
        autoSkip: false,
        maxRotation: 35,
        minRotation: 0,
        font: { size: 10 }
      }
    },
    y: { min: 90, max: 100, title: { display: true, text: t('analytics.caattUi.dataQuality.qualityAxis') } }
  }
}))
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <CaattHeader
      :title="t('navigation.caattDataQuality')"
      :subtitle="t('analytics.caattUi.dataQuality.subtitle')"
      icon="i-lucide-gauge"
      :badgeText="t('analytics.caattUi.dataQuality.badge')"
      badgeColor="success"
    />

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center items-center h-96">
      <div class="flex flex-col items-center gap-4">
        <UIcon name="i-heroicons-arrow-path" class="w-12 h-12 animate-spin text-indigo-500" />
        <span class="text-sm font-semibold text-gray-500 animate-pulse">{{ t('analytics.loading') }}</span>
      </div>
    </div>

    <!-- Content -->
    <div v-else class="space-y-6">
      <UAlert
        icon="i-heroicons-server-stack"
        color="primary"
        variant="subtle"
        :title="t('analytics.caattUi.dataQuality.title')"
        :description="t('analytics.caattUi.dataQuality.description')"
      />

      <!-- Overall Quality Score Hero Banner -->
      <UCard class="bg-gradient-to-r from-indigo-500/10 via-purple-500/10 to-emerald-500/10 border border-indigo-200 dark:border-indigo-800">
        <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div class="flex items-center gap-4">
            <div class="w-16 h-16 rounded-2xl bg-indigo-600 flex items-center justify-center text-white shadow-lg shrink-0">
              <UIcon name="i-heroicons-check-badge" class="w-10 h-10" />
            </div>
            <div>
              <div class="text-xs font-bold uppercase tracking-widest text-indigo-500">{{ t('analytics.caattUi.dataQuality.overallScore') }}</div>
              <div class="text-2xl md:text-3xl font-black text-gray-900 dark:text-white mt-0.5">{{ dqState.overallQualityScore }}{{ t('analytics.caattUi.dataQuality.excellent') }}</div>
              <div class="text-xs text-gray-500">{{ t('analytics.caattUi.dataQuality.profiledPrefix') }} {{ Number(dqState.summary.totalRowsProfiled).toLocaleString('id-ID') }} {{ t('analytics.caattUi.dataQuality.profiledSuffix') }}</div>
            </div>
          </div>
          <UBadge color="success" variant="solid" size="lg" class="px-4 py-2 font-bold text-sm shrink-0"> {{ t('analytics.caattUi.dataQuality.productionReady') }} </UBadge>
        </div>
      </UCard>

      <!-- Summary Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
        <!-- Card 1: Average Completeness -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.dataQuality.averageCompleteness') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ dqState.summary.avgCompleteness }}%
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-indigo-50 p-2.5 flex items-center justify-center dark:bg-indigo-500/10">
              <UIcon name="i-lucide-check-circle-2" class="text-indigo-600 size-5 dark:text-indigo-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-indigo-500 font-semibold">100% Valid</span>
              <span class="ml-1">{{ t('analytics.caattUi.dataQuality.noNulls') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 2: Average Accuracy -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.dataQuality.averageAccuracy') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ dqState.summary.avgAccuracy }}%
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-emerald-50 p-2.5 flex items-center justify-center dark:bg-emerald-500/10">
              <UIcon name="i-lucide-shield-check" class="text-emerald-600 size-5 dark:text-emerald-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-emerald-500 font-semibold">Akurat</span>
              <span class="ml-1">{{ t('analytics.caattUi.dataQuality.validity') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 3: Average Delay -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.dataQuality.averageDelay') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ dqState.summary.avgTimelinessDays }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-amber-50 p-2.5 flex items-center justify-center dark:bg-amber-500/10">
              <UIcon name="i-lucide-clock" class="text-amber-600 size-5 dark:text-amber-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-amber-500 font-semibold">{{ t('analytics.isolation.anomalyTypes.fieldworkUnit') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.dataQuality.freshness') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 4: Profiled Tables -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.dataQuality.profiledTables') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ dqState.summary.tablesProfiled }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-violet-50 p-2.5 flex items-center justify-center dark:bg-violet-500/10">
              <UIcon name="i-lucide-database" class="text-violet-600 size-5 dark:text-violet-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-violet-500 font-semibold">{{ t('analytics.caattUi.dataQuality.tableUnit') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.dataQuality.zones') }}</span>
            </p>
          </div>
        </div>
      </div>

      <!-- Chart & Table Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- DQ Chart -->
        <UCard class="lg:col-span-1">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-chart-bar" class="text-indigo-500" />
              <h3 class="font-bold">{{ t('analytics.caattUi.dataQuality.comparison') }}</h3>
            </div>
          </template>
          <div class="h-80">
            <Bar :data="dqChartData" :options="dqChartOptions" />
          </div>
        </UCard>

        <!-- DQ Table -->
        <UCard class="lg:col-span-2">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-table-cells" class="text-indigo-500" />
              <h3 class="font-bold">{{ t('analytics.caattUi.dataQuality.metrics') }}</h3>
            </div>
          </template>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-gray-200 dark:border-gray-700">
                  <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.dataQuality.tableName') }}</th>
                  <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.dataQuality.zone') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.dataQuality.totalRows') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.dataQuality.completeness') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.dataQuality.accuracy') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.dataQuality.latency') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.dataQuality.duplicates') }}</th>
                  <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('strategicPlan.columns.status') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="r in dqState.records" :key="r.tableName" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                  <td class="py-3 px-3">
                    <div class="font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                      <UIcon :name="getTableIcon(r.tableName)" class="w-4 h-4 text-indigo-500 shrink-0" />
                      <span>{{ getTableFriendlyName(r.tableName) }}</span>
                    </div>
                    <div class="text-[11px] text-gray-400 font-mono mt-0.5 pl-6" :title="`${t('analytics.caattUi.dataQuality.technicalTable')}: ${r.tableName}`">
                      {{ r.tableName }}
                    </div>
                  </td>
                  <td class="text-center py-3 px-3">
                    <UBadge :color="r.zone === 'Gold' ? 'warning' : r.zone === 'Silver' ? 'neutral' : 'primary'" variant="solid" size="sm">
                      {{ r.zone }}
                    </UBadge>
                  </td>
                  <td class="text-right py-3 px-3 font-mono">{{ Number(r.totalRows).toLocaleString('id-ID') }}</td>
                  <td class="text-right py-3 px-3 font-mono font-bold text-emerald-500">{{ r.completenessPct }}%</td>
                  <td class="text-right py-3 px-3 font-mono font-bold text-indigo-500">{{ r.accuracyPct }}%</td>
                  <td class="text-right py-3 px-3 font-mono text-gray-500">{{ r.timelinessDays }} h</td>
                  <td class="text-right py-3 px-3 font-mono" :class="r.duplicateCount > 0 ? 'text-rose-500 font-bold' : 'text-gray-400'">
                    {{ r.duplicateCount }}
                  </td>
                  <td class="text-center py-3 px-3">
                    <UBadge :color="r.status === 'EXCELLENT' ? 'success' : r.status === 'GOOD' ? 'primary' : 'warning'" variant="subtle" size="sm">
                      {{ r.status }}
                    </UBadge>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </UCard>
      </div>
    </div>
  </div>
</template>
