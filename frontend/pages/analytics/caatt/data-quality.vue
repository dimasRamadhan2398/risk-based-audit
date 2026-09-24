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
      label: 'Kelengkapan / Completeness (%)',
      backgroundColor: 'rgba(59,130,246,0.75)',
      borderRadius: 4,
      data: dqState.value.records.map((r: any) => r.completenessPct),
    },
    {
      label: 'Akurasi / Accuracy (%)',
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
    y: { min: 90, max: 100, title: { display: true, text: 'Skor Mutu (%)' } }
  }
}))
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <CaattHeader
      :title="t('navigation.caattDataQuality')"
      subtitle="Memantau kualitas dan reliabilitas data terintegrasi: kelengkapan (completeness), akurasi, timeliness, dan duplikasi di data lake."
      icon="i-lucide-gauge"
      badgeText="Data Lake Profiling"
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
        title="CAATT: Dashboard Kualitas Data Data Lake (Bronze → Silver → Gold)"
        description="Memantau kualitas dan reliabilitas data terintegrasi mencakup kelengkapan data (completeness), akurasi, timeliness (latensi sinkronisasi), dan keunikan (duplikasi) di setiap zona data lake."
      />

      <!-- Overall Quality Score Hero Banner -->
      <UCard class="bg-gradient-to-r from-indigo-500/10 via-purple-500/10 to-emerald-500/10 border border-indigo-200 dark:border-indigo-800">
        <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div class="flex items-center gap-4">
            <div class="w-16 h-16 rounded-2xl bg-indigo-600 flex items-center justify-center text-white shadow-lg shrink-0">
              <UIcon name="i-heroicons-check-badge" class="w-10 h-10" />
            </div>
            <div>
              <div class="text-xs font-bold uppercase tracking-widest text-indigo-500">Overall Data Quality Score</div>
              <div class="text-2xl md:text-3xl font-black text-gray-900 dark:text-white mt-0.5">{{ dqState.overallQualityScore }}% — EXCELLENT</div>
              <div class="text-xs text-gray-500">Berdasarkan profiling 6 tabel inti dan {{ Number(dqState.summary.totalRowsProfiled).toLocaleString('id-ID') }} baris data</div>
            </div>
          </div>
          <UBadge color="success" variant="solid" size="lg" class="px-4 py-2 font-bold text-sm shrink-0">
            Production Ready
          </UBadge>
        </div>
      </UCard>

      <!-- Summary Cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <UCard class="border border-indigo-100 dark:border-indigo-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Rata-Rata Kelengkapan</div>
          <div class="text-2xl font-black text-indigo-600 dark:text-indigo-400 mt-1">{{ dqState.summary.avgCompleteness }}%</div>
          <div class="text-[10px] text-gray-400 mt-0.5">Bebas dari null penting</div>
        </UCard>
        <UCard class="border border-emerald-100 dark:border-emerald-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Rata-Rata Akurasi</div>
          <div class="text-2xl font-black text-emerald-600 dark:text-emerald-400 mt-1">{{ dqState.summary.avgAccuracy }}%</div>
          <div class="text-[10px] text-emerald-500 mt-0.5">Validitas tipe data & relasi</div>
        </UCard>
        <UCard class="border border-amber-100 dark:border-amber-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Rata-Rata Keterlambatan</div>
          <div class="text-2xl font-black text-amber-600 dark:text-amber-400 mt-1">{{ dqState.summary.avgTimelinessDays }} Hari</div>
          <div class="text-[10px] text-amber-500 mt-0.5">Freshness sync pipeline</div>
        </UCard>
        <UCard class="border border-violet-100 dark:border-violet-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Tabel Diprofil</div>
          <div class="text-2xl font-black text-violet-600 dark:text-violet-400 mt-1">{{ dqState.summary.tablesProfiled }} Tabel</div>
          <div class="text-[10px] text-gray-400 mt-0.5">Zona Bronze, Silver, Gold</div>
        </UCard>
      </div>

      <!-- Chart & Table Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- DQ Chart -->
        <UCard class="lg:col-span-1">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-chart-bar" class="text-indigo-500" />
              <h3 class="font-bold">Kelengkapan vs Akurasi Tabel</h3>
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
              <h3 class="font-bold">Metrik Profiling Kualitas Data Lake Multi-Zona</h3>
            </div>
          </template>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-gray-200 dark:border-gray-700">
                  <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Nama Tabel</th>
                  <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Zona</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Total Baris</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Kelengkapan</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Akurasi</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Latensi</th>
                  <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Duplikat</th>
                  <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Status</th>
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
