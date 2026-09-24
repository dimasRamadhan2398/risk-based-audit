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
      label: 'Frekuensi Aktual (%)',
      backgroundColor: benfordState.value.records.map((r: any) => r.isSignificant ? 'rgba(239,68,68,0.85)' : 'rgba(99,102,241,0.75)'),
      borderColor: benfordState.value.records.map((r: any) => r.isSignificant ? 'rgb(239,68,68)' : 'rgb(99,102,241)'),
      borderWidth: 1.5,
      borderRadius: 4,
      data: benfordState.value.records.map((r: any) => r.actualPct),
    },
    {
      label: 'Kurva Benford (%)',
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
    y: { beginAtZero: true, max: 35, title: { display: true, text: 'Distribusi Frekuensi (%)' } }
  }
}))
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <CaattHeader
      :title="t('navigation.caattBenford')"
      subtitle="Menguji sebaran frekuensi angka pertama dari seluruh transaksi perbankan untuk mendeteksi indikasi rekayasa data."
      icon="i-lucide-calculator"
      badgeText="Uji Forensik Benford"
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
        title="CAATT: Analisis Hukum Benford (First-Digit Frequency Distribution)"
        description="Menguji sebaran frekuensi angka pertama dari seluruh transaksi perbankan. Deviasi signifikan mengindikasikan kemungkinan rekayasa data atau transaksi terstruktur (structuring) di bawah threshold pelaporan."
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
              <h3 class="font-bold">Perbandingan Distribusi Digit: Aktual vs Hukum Benford</h3>
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
              <UIcon name="i-heroicons-table-cells" class="text-indigo-500" />
              <h3 class="font-bold">Uji Deviasi Chi-Square Per Digit (1 - 9)</h3>
            </div>
          </template>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-gray-200 dark:border-gray-700">
                  <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Digit</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Jumlah Trx</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">% Aktual</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">% Benford</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Deviasi</th>
                  <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Status</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="r in benfordState.records" :key="r.digit" :class="r.isSignificant ? 'bg-rose-50/50 dark:bg-rose-950/20' : ''" class="border-b border-gray-100 dark:border-gray-800">
                  <td class="text-center py-3 px-3 font-bold text-lg text-indigo-600 dark:text-indigo-400">{{ r.digit }}</td>
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
        </UCard>
      </div>
    </div>
  </div>
</template>
