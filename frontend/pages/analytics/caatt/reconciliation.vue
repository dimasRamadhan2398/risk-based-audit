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
const { loading, reconState, formatIDR, fetchCaattAnalytics } = useCaattAnalytics()

onMounted(() => {
  fetchCaattAnalytics()
})

const reconChartData = computed(() => ({
  labels: reconState.value.records.map((r: any) => `${r.systemA.split(' ')[0]} ↔ ${r.systemB.split(' ')[0]}`),
  datasets: [
    {
      label: t('analytics.caattUi.reconciliation.matchRateAxis'),
      backgroundColor: reconState.value.records.map((r: any) => r.matchRatePct >= 99.9 ? 'rgba(16,185,129,0.85)' : 'rgba(245,158,11,0.85)'),
      borderRadius: 4,
      data: reconState.value.records.map((r: any) => r.matchRatePct),
    },
  ],
}))

const reconChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (ctx: any) => `Match Rate: ${ctx.parsed.y}%`
      }
    }
  },
  scales: {
    y: { min: 95, max: 100, title: { display: true, text: t('analytics.caattUi.reconciliation.matchRateAxis') } }
  }
}))
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <CaattHeader
      :title="t('navigation.caattReconciliation')"
      :subtitle="t('analytics.caattUi.reconciliation.subtitle')"
      icon="i-lucide-git-compare"
      :badgeText="t('analytics.caattUi.reconciliation.badge')"
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
        icon="i-heroicons-arrows-right-left"
        color="primary"
        variant="subtle"
        :title="t('analytics.caattUi.reconciliation.title')"
        :description="t('analytics.caattUi.reconciliation.description')"
      />

      <!-- Summary Cards -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 sm:gap-6">
        <!-- Card 1: Average Match Rate -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.reconciliation.averageMatch') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ reconState.summary.avgMatchRate }}%
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-emerald-50 p-2.5 flex items-center justify-center dark:bg-emerald-500/10">
              <UIcon name="i-lucide-check-check" class="text-emerald-600 size-5 dark:text-emerald-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-emerald-500 font-semibold">{{ reconState.summary.avgMatchRate >= 99 ? 'SLA Passed' : 'Under Review' }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.reconciliation.sla') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 2: Unmatched Transactions -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.reconciliation.unmatched') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ reconState.summary.totalUnmatched }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-rose-50 p-2.5 flex items-center justify-center dark:bg-rose-500/10">
              <UIcon name="i-lucide-file-x" class="text-rose-600 size-5 dark:text-rose-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-rose-500 font-semibold">{{ t('analytics.caattUi.common.transactions') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.reconciliation.journalAdjustment') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 3: Total Difference -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.reconciliation.difference') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ formatIDR(reconState.summary.totalDifference) }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-amber-50 p-2.5 flex items-center justify-center dark:bg-amber-500/10">
              <UIcon name="i-lucide-scale" class="text-amber-600 size-5 dark:text-amber-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-amber-500 font-semibold">Selisih</span>
              <span class="ml-1">{{ t('analytics.caattUi.reconciliation.clearing') }}</span>
            </p>
          </div>
        </div>
      </div>

      <!-- Chart & Table Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Chart -->
        <UCard class="lg:col-span-1">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-chart-bar" class="text-indigo-500" />
              <h3 class="font-bold">{{ t('analytics.caattUi.reconciliation.integrationMatch') }}</h3>
            </div>
          </template>
          <div class="h-80">
            <Bar :data="reconChartData" :options="reconChartOptions" />
          </div>
        </UCard>

        <!-- Table -->
        <UCard class="lg:col-span-2">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-table-cells" class="text-indigo-500" />
              <h3 class="font-bold">{{ t('analytics.caattUi.reconciliation.status') }}</h3>
            </div>
          </template>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-gray-200 dark:border-gray-700">
                  <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.reconciliation.systems') }}</th>
                  <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('settings.permission.tableModule') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.reconciliation.transactionsA') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.reconciliation.transactionsB') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.reconciliation.matched') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.reconciliation.matchPercent') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.reconciliation.differenceCurrency') }}</th>
                  <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('strategicPlan.columns.status') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="r in reconState.records" :key="r.module" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                  <td class="py-3 px-3 font-bold">{{ r.systemA.split(' ')[0] }} ↔ {{ r.systemB.split(' ')[0] }}</td>
                  <td class="py-3 px-3 text-gray-600 dark:text-gray-300">{{ r.module }}</td>
                  <td class="text-right py-3 px-3 font-mono">{{ Number(r.totalRecordsA).toLocaleString('id-ID') }}</td>
                  <td class="text-right py-3 px-3 font-mono">{{ Number(r.totalRecordsB).toLocaleString('id-ID') }}</td>
                  <td class="text-right py-3 px-3 font-mono font-bold text-emerald-500">{{ Number(r.matchedRecords).toLocaleString('id-ID') }}</td>
                  <td class="text-right py-3 px-3 font-mono font-bold">{{ r.matchRatePct }}%</td>
                  <td class="text-right py-3 px-3 font-mono" :class="r.totalDifference > 0 ? 'text-rose-500 font-bold' : 'text-gray-400'">
                    {{ formatIDR(r.totalDifference) }}
                  </td>
                  <td class="text-center py-3 px-3">
                    <UBadge :color="r.status === 'PERFECT_MATCH' || r.status === 'BALANCED' ? 'success' : 'warning'" variant="solid" size="sm">
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
