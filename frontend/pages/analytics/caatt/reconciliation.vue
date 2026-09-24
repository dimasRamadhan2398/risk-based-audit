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
      label: 'Tingkat Kesesuaian / Match Rate (%)',
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
    y: { min: 95, max: 100, title: { display: true, text: 'Match Rate (%)' } }
  }
}))
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <CaattHeader
      :title="t('navigation.caattReconciliation')"
      subtitle="Memvalidasi integritas data perbankan dengan membandingkan mutasi dan saldo antara Core Banking (CBS), GL, LOS, dan Switch."
      icon="i-lucide-git-compare"
      badgeText="3-Way Automated Reconciliation"
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
        title="CAATT: Rekonsiliasi Otomatis Lintas Sistem (3-Way Reconciliation)"
        description="Memvalidasi integritas data perbankan dengan membandingkan mutasi dan saldo transaksi antara Core Banking System (CBS), General Ledger (GL), Loan Origination System (LOS), dan ATM Switch."
      />

      <!-- Summary Cards -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <UCard class="border border-emerald-100 dark:border-emerald-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Rata-Rata Tingkat Kesesuaian</div>
          <div class="text-2xl font-black text-emerald-600 dark:text-emerald-400 mt-1">{{ reconState.summary.avgMatchRate }}%</div>
          <div class="text-[10px] text-emerald-500 mt-0.5">Standar SLA: 99.5%</div>
        </UCard>
        <UCard class="border border-rose-100 dark:border-rose-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Total Transaksi Unmatched</div>
          <div class="text-2xl font-black text-rose-600 dark:text-rose-400 mt-1">{{ reconState.summary.totalUnmatched }} Trx</div>
          <div class="text-[10px] text-rose-400 mt-0.5">Memerlukan penyesuaian jurnal</div>
        </UCard>
        <UCard class="border border-amber-100 dark:border-amber-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Total Selisih Nominal</div>
          <div class="text-2xl font-black text-amber-600 dark:text-amber-400 mt-1">{{ formatIDR(reconState.summary.totalDifference) }}</div>
          <div class="text-[10px] text-amber-400 mt-0.5">Discrepancy dalam proses kliring</div>
        </UCard>
      </div>

      <!-- Chart & Table Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Chart -->
        <UCard class="lg:col-span-1">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-chart-bar" class="text-indigo-500" />
              <h3 class="font-bold">Match Rate Per Jalur Integrasi</h3>
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
              <h3 class="font-bold">Status Rekonsiliasi Buku Besar & Sistem Operasional</h3>
            </div>
          </template>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-gray-200 dark:border-gray-700">
                  <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Sistem A ↔ Sistem B</th>
                  <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Modul</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Trx A</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Trx B</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Cocok</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Match %</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Selisih (Rp)</th>
                  <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Status</th>
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
