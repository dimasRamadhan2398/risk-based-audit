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
const { loading, stratState, formatIDR, fetchCaattAnalytics } = useCaattAnalytics()

onMounted(() => {
  fetchCaattAnalytics()
})

const stratChartData = computed(() => ({
  labels: stratState.value.records.map((r: any) => r.stratumLabel),
  datasets: [
    {
      label: 'Total Nilai Akumulasi (Rp Miliar)',
      backgroundColor: 'rgba(139,92,246,0.75)',
      borderColor: 'rgb(139,92,246)',
      borderWidth: 1.5,
      borderRadius: 4,
      data: stratState.value.records.map((r: any) => +(r.totalAmount / 1000000000).toFixed(1)),
    },
  ],
}))

const stratChartOptions = computed(() => ({
  indexAxis: 'y' as const,
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (ctx: any) => `Total Akumulasi: Rp ${ctx.parsed.x} Miliar`
      }
    }
  },
  scales: {
    x: { beginAtZero: true, title: { display: true, text: 'Akumulasi Nilai (Miliar Rupiah)' } }
  }
}))
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <CaattHeader
      :title="t('navigation.caattStratification')"
      subtitle="Membagi populasi transaksi dan portofolio ke dalam tingkatan nominal (strata) untuk mengidentifikasi konsentrasi eksposur risiko."
      icon="i-lucide-layers"
      badgeText="Monetary Unit Sampling"
      badgeColor="neutral"
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
        icon="i-heroicons-bars-3-bottom-left"
        color="primary"
        variant="subtle"
        title="CAATT: Stratifikasi Transaksi & Portofolio Kredit"
        description="Membagi populasi transaksi dan portofolio ke dalam tingkatan nominal (strata) untuk mengidentifikasi konsentrasi eksposur risiko dan memilih sampel bernilai tinggi (Monetary Unit Sampling)."
      />

      <!-- Summary Cards -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <UCard class="border border-indigo-100 dark:border-indigo-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Total Transaksi Terstratifikasi</div>
          <div class="text-2xl font-black text-indigo-600 dark:text-indigo-400 mt-1">{{ Number(stratState.summary.totalTransactions).toLocaleString('id-ID') }} Trx</div>
          <div class="text-[10px] text-gray-400 mt-0.5">Seluruh kategori nasabah</div>
        </UCard>
        <UCard class="border border-violet-100 dark:border-violet-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Total Nilai Portofolio</div>
          <div class="text-2xl font-black text-violet-600 dark:text-violet-400 mt-1">{{ formatIDR(stratState.summary.totalAmount) }}</div>
          <div class="text-[10px] text-violet-400 mt-0.5">Eksposur nominal penuh</div>
        </UCard>
        <UCard class="border border-emerald-100 dark:border-emerald-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Jumlah Lapisan Strata</div>
          <div class="text-2xl font-black text-emerald-600 dark:text-emerald-400 mt-1">{{ stratState.summary.totalStrata }} Strata</div>
          <div class="text-[10px] text-gray-400 mt-0.5">Retail s/d High Value Wholesale</div>
        </UCard>
      </div>

      <!-- Chart & Table Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Strat Chart -->
        <UCard class="lg:col-span-1">
          <template #header>
            <div>
              <div class="flex items-center gap-2">
                <UIcon name="i-heroicons-chart-bar" class="text-indigo-500" />
                <h3 class="font-bold">Konsentrasi Nilai Per Strata</h3>
              </div>
              <p class="text-xs text-gray-400 mt-1">Akumulasi eksposur nilai (Rp Miliar) per rentang strata</p>
            </div>
          </template>
          <div class="h-80">
            <Bar :data="stratChartData" :options="stratChartOptions" />
          </div>
        </UCard>

        <!-- Strat Table -->
        <UCard class="lg:col-span-2">
          <template #header>
            <div>
              <div class="flex items-center gap-2">
                <UIcon name="i-heroicons-table-cells" class="text-indigo-500" />
                <h3 class="font-bold">Tabel Distribusi Strata Nominal Transaksi</h3>
              </div>
              <p class="text-xs text-gray-400 mt-1">Distribusi jumlah transaksi dan akumulasi nominal per lapisan rentang</p>
            </div>
          </template>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-gray-200 dark:border-gray-700">
                  <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Strata (Rentang Nilai Trx)</th>
                  <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Segmen</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Jumlah Trx</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Total Akumulasi Nominal</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">% Trx</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">% Nilai</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="r in stratState.records" :key="r.stratumLabel" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                  <td class="py-3 px-3 font-bold text-indigo-600 dark:text-indigo-400">{{ r.stratumLabel }}</td>
                  <td class="py-3 px-3"><UBadge color="neutral" variant="subtle" size="sm">{{ r.category }}</UBadge></td>
                  <td class="text-right py-3 px-3 font-mono">{{ Number(r.trxCount).toLocaleString('id-ID') }}</td>
                  <td class="text-right py-3 px-3 font-mono font-bold">{{ formatIDR(r.totalAmount) }}</td>
                  <td class="text-right py-3 px-3 font-mono">{{ r.pctCount }}%</td>
                  <td class="text-right py-3 px-3 font-mono font-bold text-violet-500">{{ r.pctAmount }}%</td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Interpretation Guide Note -->
          <div class="mt-4 p-3 rounded-lg bg-gray-50 dark:bg-gray-800/60 border border-gray-100 dark:border-gray-700/60 text-xs text-gray-500 dark:text-gray-400 space-y-1.5">
            <div class="flex items-center gap-1.5 font-semibold text-gray-700 dark:text-gray-300">
              <UIcon name="i-heroicons-information-circle" class="w-4 h-4 text-indigo-500" />
              <span>Panduan Interpretasi Kolom:</span>
            </div>
            <p>
              &bull; <strong class="text-gray-700 dark:text-gray-300">Strata (Rentang Nilai Trx):</strong> Merupakan batasan nominal per 1 lembar transaksi individu (contoh: transaksi ritel bernilai &lt; Rp 10 Juta hingga wholesale &gt; Rp 500 Juta).
            </p>
            <p>
              &bull; <strong class="text-gray-700 dark:text-gray-300">Total Akumulasi Nominal:</strong> Merupakan total kumulatif dari seluruh transaksi dalam strata tersebut (contoh: 8.420 transaksi bernilai &lt; Rp 10 Juta memiliki akumulasi total Rp 32,5 Miliar).
            </p>
          </div>
        </UCard>
      </div>
    </div>
  </div>
</template>
