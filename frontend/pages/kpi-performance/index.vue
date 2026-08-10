<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import KpiSummaryCards from '~/components/kpi-performance/KpiSummaryCards.vue'
import KpiCharts from '~/components/kpi-performance/KpiCharts.vue'
import KpiDetailedTable from '~/components/kpi-performance/KpiDetailedTable.vue'
import TimelinessQuestionnaireModal from '~/components/kpi-performance/TimelinessQuestionnaireModal.vue'

import { usePerformanceStore } from '~/stores/performance'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useUploadPerformanceReportStore } from '~/stores/upload-performance-report'

const perfStore = usePerformanceStore()
const spStore = useStrategicPlanStore()
const uploadStore = useUploadPerformanceReportStore()

const year = ref('2026')
const selectedPeriod = ref('Semua')
const yearOptions = ['2024', '2025', '2026', '2027', '2028']
const periodOptions = ['Semua', 'Q1', 'Q2', 'Q3', 'Q4', 'Tahunan']

const isQuestionnaireOpen = ref(false)

const loadData = () => {
  perfStore.fetchKPIAchievements(parseInt(year.value), selectedPeriod.value)
  perfStore.fetchWorkPlanRealizations(parseInt(year.value))
  spStore.fetchStrategicPlans()
  uploadStore.fetchUploadedReports(selectedPeriod.value, parseInt(year.value))
}

onMounted(() => {
  loadData()
})

watch([year, selectedPeriod], () => {
  loadData()
})

const exportPDF = () => {
  useToast().add({
    title: 'Exporting...',
    description: 'KPI Performance report is being exported to PDF.',
    color: 'success'
  })
}
</script>

<template>
  <div class="p-6 space-y-8">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">KPI Performance</h1>
        <p class="text-sm font-semibold text-gray-500 mt-1">Monitor and track internal audit performance metrics & Laporan Kinerja (Q1, Q2, Q3, Q4, Tahunan)</p>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <!-- Period Selector -->
        <USelect
          v-model="selectedPeriod"
          :items="periodOptions"
          class="w-32"
          placeholder="Periode"
        />
        <!-- Year Selector -->
        <USelect
          v-model="year"
          :items="yearOptions"
          class="w-28"
        />
        <!-- Upload Laporan Kinerja Button -->
        <UButton
          label="Impor Laporan Kinerja"
          icon="i-lucide-upload"
          color="primary"
          to="/kpi-performance/upload"
        />
        <!-- Isi Kuesioner Button -->
        <UButton
          v-if="selectedPeriod !== 'Semua'"
          label="Isi Kuesioner Ketepatan Waktu"
          icon="i-lucide-clipboard-check"
          color="info"
          variant="solid"
          @click="isQuestionnaireOpen = true"
        />
        <!-- Export Button -->
        <UButton
          label="Export PDF"
          icon="i-lucide-download"
          color="neutral"
          variant="outline"
          @click="exportPDF"
        />
      </div>
    </div>

    <!-- Status Banner Laporan Kinerja Terimpor -->
    <UCard v-if="uploadStore.uploadedReports.length > 0" class="border-l-4 border-l-primary bg-primary/5 dark:bg-primary/10">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-lg bg-primary/20 text-primary">
            <UIcon name="i-lucide-file-check-2" class="w-5 h-5" />
          </div>
          <div>
            <div class="text-sm font-bold text-gray-900 dark:text-white">
              Dokumen Laporan Kinerja Terimpor ({{ uploadStore.uploadedReports.length }} Laporan)
            </div>
            <div class="text-xs text-gray-600 dark:text-gray-400 mt-0.5">
              Laporan Kinerja aktif: <span class="font-bold">{{ uploadStore.uploadedReports[0]?.title }}</span> ({{ uploadStore.uploadedReports[0]?.period }} {{ uploadStore.uploadedReports[0]?.year }})
            </div>
          </div>
        </div>
        <UButton
          label="Kelola Dokumen Laporan"
          icon="i-lucide-arrow-right"
          color="primary"
          variant="subtle"
          size="xs"
          to="/kpi-performance/upload"
        />
      </div>
    </UCard>

    <!-- Summary Cards -->
    <KpiSummaryCards />

    <!-- Charts -->
    <KpiCharts />

    <!-- Detailed Table -->
    <KpiDetailedTable />

    <TimelinessQuestionnaireModal
      v-model="isQuestionnaireOpen"
      :year="parseInt(year)"
      :period="selectedPeriod"
      @saved="loadData"
    />
  </div>
</template>
