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
  const yr = parseInt(year.value)
  perfStore.fetchKPIAchievements(yr, selectedPeriod.value)
  perfStore.fetchWorkPlanRealizations(yr)
  perfStore.fetchDashboardSummary(yr)
  perfStore.fetchMonthlyTrends(yr)
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
  const config = useRuntimeConfig()
  const auditBaseUrl = config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
  const reportUrl = `${auditBaseUrl}/performance/export-pdf?year=${year.value}`

  useToast().add({
    title: 'Generating Executive PDF Report...',
    description: `Opening official KPI Performance PDF Report for Year ${year.value}...`,
    color: 'success'
  })

  window.open(reportUrl, '_blank')
}
</script>

<template>
  <div class="p-6 space-y-8 print:p-0 print:space-y-4 print:bg-white print:text-black">
    <!-- Printable Document Header (Visible only during PDF Print) -->
    <div class="hidden print:block border-b-2 border-primary-600 pb-4 mb-6">
      <div class="flex justify-between items-center">
        <div>
          <h1 class="text-xl font-bold uppercase tracking-wider text-gray-900">
            INTERNAL AUDIT DIVISION
          </h1>
          <h2 class="text-lg font-semibold text-primary-700">
            KPI Performance Report - Year {{ year }}
          </h2>
          <p class="text-xs text-gray-500 mt-0.5">
            Generated on: {{ new Date().toLocaleDateString('id-ID', { dateStyle: 'full' }) }}
          </p>
        </div>
        <div class="text-right text-xs text-gray-500">
          <span class="font-bold text-gray-800">Risk-Based Audit System</span>
          <br />
          <span>Confidential - Internal Use Only</span>
        </div>
      </div>
    </div>

    <!-- Screen Header (Hidden during PDF Print) -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 print:hidden">
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
          color="warning"
          variant="outline"
          class="font-bold shadow-sm"
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
    <KpiSummaryCards :year="parseInt(year)" />

    <!-- Charts -->
    <KpiCharts :year="parseInt(year)" />

    <!-- Detailed Table -->
    <KpiDetailedTable :year="parseInt(year)" />

    <TimelinessQuestionnaireModal
      v-model="isQuestionnaireOpen"
      :year="parseInt(year)"
      :period="selectedPeriod"
      @saved="loadData"
    />

    <!-- Printable Sign-off Footer (Visible only during PDF Print) -->
    <div class="hidden print:grid grid-cols-3 gap-8 pt-8 mt-8 border-t border-gray-300 text-center text-xs">
      <div>
        <p class="font-bold text-gray-700">Prepared By:</p>
        <div class="h-16"></div>
        <p class="font-semibold text-gray-900 border-t border-gray-400 pt-1">Internal Audit Specialist</p>
      </div>
      <div>
        <p class="font-bold text-gray-700">Reviewed By:</p>
        <div class="h-16"></div>
        <p class="font-semibold text-gray-900 border-t border-gray-400 pt-1">Audit Quality Manager</p>
      </div>
      <div>
        <p class="font-bold text-gray-700">Approved By:</p>
        <div class="h-16"></div>
        <p class="font-semibold text-gray-900 border-t border-gray-400 pt-1">Chief Audit Executive (CAE)</p>
      </div>
    </div>
  </div>
</template>

<style>
@media print {
  body {
    background: white !important;
    color: black !important;
  }
  aside, header, nav, button, .print\:hidden {
    display: none !important;
  }
  .print\:block {
    display: block !important;
  }
  .print\:grid {
    display: grid !important;
  }
  @page {
    size: A4 portrait;
    margin: 1.2cm;
  }
}
</style>
