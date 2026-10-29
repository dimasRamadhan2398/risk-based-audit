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
const { loading, benfordState, fetchCaattAnalytics } = useCaattAnalytics()

onMounted(() => {
  fetchCaattAnalytics()
})

const benfordChartData = computed(() => ({
  labels: benfordState.value.records.map((r: any) => `Digit ${r.digit}`),
  datasets: [
    {
      label: t('analytics.caattUi.benford.actualFrequency'),
      backgroundColor: benfordState.value.records.map((r: any) => r.isSignificant ? 'rgba(239,68,68,0.85)' : 'rgba(99,102,241,0.75)'),
      borderColor: benfordState.value.records.map((r: any) => r.isSignificant ? 'rgb(239,68,68)' : 'rgb(99,102,241)'),
      borderWidth: 1.5,
      borderRadius: 4,
      data: benfordState.value.records.map((r: any) => r.actualPct),
    },
    {
      label: t('analytics.caattUi.benford.benfordCurve'),
      backgroundColor: 'rgba(16,185,129,0.75)',
      borderColor: 'rgb(16,185,129)',
      borderWidth: 1.5,
      borderRadius: 4,
      data: benfordState.value.records.map((r: any) => r.expectedPct),
    },
  ],
}))

const benfordChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'top' as const },
    tooltip: {
      callbacks: {
        label: (ctx: any) => `${ctx.dataset.label}: ${ctx.parsed.y}%`
      }
    }
  },
  scales: {
    y: { beginAtZero: true, max: 35, title: { display: true, text: t('analytics.caattUi.benford.frequencyDistribution') } }
  }
}))
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <CaattHeader
      :title="t('navigation.caattBenford')"
      :subtitle="t('analytics.caattUi.benford.subtitle')"
      icon="i-lucide-calculator"
      :badgeText="t('analytics.caattUi.benford.badge')"
      badgeColor="info"
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
        icon="i-heroicons-variable"
        color="primary"
        variant="subtle"
        :title="t('analytics.caattUi.benford.title')"
        :description="t('analytics.caattUi.benford.description')"
      />

      <!-- Conclusion Alert -->
      <UAlert
        icon="i-heroicons-information-circle"
        :color="benfordState.summary.significantDeviations > 0 ? 'warning' : 'success'"
        variant="solid"
        :title="benfordState.summary.significantDeviations > 0 ? 'Indikasi Deviasi Digit Pertama Terdeteksi' : 'Distribusi Angka Normal'"
        :description="benfordState.summary.conclusion"
      />

      <!-- Chart & Table Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Benford Chart -->
        <UCard>
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-chart-bar" class="text-indigo-500" />
              <h3 class="font-bold">{{ t('analytics.caattUi.benford.comparison') }}</h3>
            </div>
          </template>
          <div class="h-80">
            <Bar :data="benfordChartData" :options="benfordChartOptions" />
          </div>
        </UCard>

        <!-- Benford Table -->
        <UCard>
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-table-cells" class="text-primary-500" />
              <h3 class="font-bold">{{ t('analytics.caattUi.benford.chiSquare') }}</h3>
            </div>
          </template>
          <!-- Desktop View: Table -->
          <div class="hidden md:block overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-gray-200 dark:border-gray-700">
                  <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.benford.digit') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.common.transactionCount') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.benford.actualPercent') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.benford.benfordPercent') }}</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.benford.deviation') }}</th>
                  <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('strategicPlan.columns.status') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="r in benfordState.records" :key="r.digit" :class="r.isSignificant ? 'bg-rose-50/50 dark:bg-rose-950/20' : ''" class="border-b border-gray-100 dark:border-gray-800">
                  <td class="text-center py-3 px-3 font-bold text-lg text-primary-600 dark:text-primary-400">{{ r.digit }}</td>
                  <td class="text-right py-3 px-3 font-mono">{{ Number(r.actualCount).toLocaleString('id-ID') }}</td>
                  <td class="text-right py-3 px-3 font-mono font-bold">{{ r.actualPct }}%</td>
                  <td class="text-right py-3 px-3 font-mono text-gray-400">{{ r.expectedPct }}%</td>
                  <td class="text-right py-3 px-3 font-mono font-bold" :class="r.deviationPct > 0 ? 'text-rose-500' : 'text-emerald-500'">
                    {{ r.deviationPct > 0 ? '+' : '' }}{{ r.deviationPct }}%
                  </td>
                  <td class="text-center py-3 px-3">
                    <UBadge :color="r.isSignificant ? 'error' : 'success'" variant="subtle" size="sm">
                      {{ r.isSignificant ? 'Signifikan' : 'Normal' }}
                    </UBadge>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Mobile View: Card List -->
          <div class="block md:hidden">
            <div v-if="!benfordState.records.length" class="text-center py-8 text-gray-400">
              <UIcon name="i-heroicons-inbox" class="w-10 h-10 mx-auto mb-2 text-gray-300 dark:text-gray-600" />
              <p class="text-sm">{{ t('common.noData') }}</p>
            </div>
            <div v-else class="space-y-3">
              <div
                v-for="(r, index) in benfordState.records"
                :key="r.digit"
                class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
                :class="r.isSignificant ? 'ring-1 ring-rose-500/30 bg-rose-50/30 dark:bg-rose-950/10' : ''"
              >
                <!-- Card Header -->
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <span class="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-primary-50 dark:bg-primary-950/50 text-primary-600 dark:text-primary-400 font-bold text-base">
                      {{ r.digit }}
                    </span>
                    <div>
                      <span class="text-xs font-semibold text-gray-900 dark:text-gray-100">Digit {{ r.digit }}</span>
                      <span class="block text-[11px] text-gray-400 font-mono">{{ Number(r.actualCount).toLocaleString('id-ID') }} tx</span>
                    </div>
                  </div>
                  <UBadge :color="r.isSignificant ? 'error' : 'success'" variant="subtle" size="xs">
                    {{ r.isSignificant ? 'Signifikan' : 'Normal' }}
                  </UBadge>
                </div>

                <!-- Percentages Grid -->
                <div class="grid grid-cols-2 gap-2 text-xs">
                  <div class="dashboard-subcard p-2 space-y-0.5 min-w-0 overflow-hidden">
                    <span class="text-slate-400 block text-[10px] uppercase font-bold truncate">{{ t('analytics.caattUi.benford.actualPercent') }}</span>
                    <span class="font-mono font-bold text-xs sm:text-sm text-gray-900 dark:text-gray-100 break-words leading-tight block">{{ r.actualPct }}%</span>
                  </div>
                  <div class="dashboard-subcard p-2 space-y-0.5 min-w-0 overflow-hidden">
                    <span class="text-slate-400 block text-[10px] uppercase font-bold truncate">{{ t('analytics.caattUi.benford.benfordPercent') }}</span>
                    <span class="font-mono font-semibold text-xs sm:text-sm text-slate-500 dark:text-slate-400 break-words leading-tight block">{{ r.expectedPct }}%</span>
                  </div>
                </div>

                <!-- Deviation Row -->
                <div class="flex items-center justify-between text-xs px-2.5 py-1.5 dashboard-subcard min-w-0 overflow-hidden">
                  <span class="text-slate-400 text-[10px] uppercase font-bold shrink-0">{{ t('analytics.caattUi.benford.deviation') }}</span>
                  <span class="font-mono font-bold text-xs sm:text-sm ml-1" :class="r.deviationPct > 0 ? 'text-rose-500' : 'text-emerald-500'">
                    {{ r.deviationPct > 0 ? '+' : '' }}{{ r.deviationPct }}%
                  </span>
                </div>

                <!-- Card Footer -->
                <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800 text-[11px] text-gray-400">
                  <span>#{{ Number(index) + 1 }}</span>
                  <span>{{ r.isSignificant ? 'Chi-Sq Deviation' : 'Within Expected' }}</span>
                </div>
              </div>
            </div>
          </div>
        </UCard>
      </div>
    </div>
  </div>
</template>
