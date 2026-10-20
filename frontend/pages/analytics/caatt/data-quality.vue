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

const dqChartData = computed(() => ({
  labels: dqState.value.records.map((r: any) => r.tableName.replace('gold.', 'G:').replace('silver.', 'S:').replace('bronze.', 'B:')),
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
  },
  scales: {
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
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <UCard class="border border-indigo-100 dark:border-indigo-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.dataQuality.averageCompleteness') }}</div>
          <div class="text-2xl font-black text-indigo-600 dark:text-indigo-400 mt-1">{{ dqState.summary.avgCompleteness }}%</div>
          <div class="text-[10px] text-gray-400 mt-0.5">{{ t('analytics.caattUi.dataQuality.noNulls') }}</div>
        </UCard>
        <UCard class="border border-emerald-100 dark:border-emerald-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.dataQuality.averageAccuracy') }}</div>
          <div class="text-2xl font-black text-emerald-600 dark:text-emerald-400 mt-1">{{ dqState.summary.avgAccuracy }}%</div>
          <div class="text-[10px] text-emerald-500 mt-0.5">{{ t('analytics.caattUi.dataQuality.validity') }}</div>
        </UCard>
        <UCard class="border border-amber-100 dark:border-amber-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.dataQuality.averageDelay') }}</div>
          <div class="text-2xl font-black text-amber-600 dark:text-amber-400 mt-1">{{ dqState.summary.avgTimelinessDays }} {{ t('analytics.isolation.anomalyTypes.fieldworkUnit') }}</div>
          <div class="text-[10px] text-amber-500 mt-0.5">{{ t('analytics.caattUi.dataQuality.freshness') }}</div>
        </UCard>
        <UCard class="border border-violet-100 dark:border-violet-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.dataQuality.profiledTables') }}</div>
          <div class="text-2xl font-black text-violet-600 dark:text-violet-400 mt-1">{{ dqState.summary.tablesProfiled }} {{ t('analytics.caattUi.dataQuality.tableUnit') }}</div>
          <div class="text-[10px] text-gray-400 mt-0.5">{{ t('analytics.caattUi.dataQuality.zones') }}</div>
        </UCard>
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
                  <td class="py-3 px-3 font-mono font-bold">{{ r.tableName }}</td>
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
