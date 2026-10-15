<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import KpiSummaryCards from '~/components/kpi-performance/KpiSummaryCards.vue'
import KpiCharts from '~/components/kpi-performance/KpiCharts.vue'
import KpiDetailedTable from '~/components/kpi-performance/KpiDetailedTable.vue'

import { usePerformanceStore } from '~/stores/performance'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useUploadPerformanceReportStore } from '~/stores/upload-performance-report'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { extractErrorMessage } from '~/utils/error'

const perfStore = usePerformanceStore()
const spStore = useStrategicPlanStore()
const uploadStore = useUploadPerformanceReportStore()
const { t, locale } = useI18n()

const year = ref('2026')
const selectedPeriod = ref('Semua')
const yearOptions = ['2024', '2025', '2026', '2027', '2028', '2029', '2030']
const periodOptions = computed(() => [
  { label: t('kpiPerformance.dashboard.allPeriods'), value: 'Semua' },
  ...['Q1', 'Q2', 'Q3', 'Q4'].map(value => ({ label: value, value })),
  { label: t('kpiPerformance.dashboard.annual'), value: 'Tahunan' }
])

const formattedCurrentDate = computed(() => new Date().toLocaleDateString(locale.value === 'id' ? 'id-ID' : 'en-US', { dateStyle: 'full' }))

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

const exportPDF = async () => {
  const auditBaseUrl = getAuditServiceBaseUrl()
  const reportUrl = `${auditBaseUrl}/performance/export-pdf?year=${year.value}`

  useToast().add({
    title: t('kpiPerformance.dashboard.messages.generating'),
    description: t('kpiPerformance.dashboard.messages.opening', { year: year.value }),
    color: 'success'
  })

  // Open the tab synchronously so the popup blocker allows it, then load the
  // PDF through $fetch (which carries the auth token) instead of a bare URL
  const tab = window.open('', '_blank')
  try {
    const pdf = await $fetch<Blob>(reportUrl, { responseType: 'blob' })
    const url = window.URL.createObjectURL(pdf)
    if (tab) {
      tab.location.href = url
    } else {
      window.open(url, '_blank')
    }
    setTimeout(() => window.URL.revokeObjectURL(url), 60000)
  } catch (error) {
    tab?.close()
    useToast().add({
      title: t('kpiPerformance.dashboard.messages.failed'),
      description: extractErrorMessage(error, t('kpiPerformance.dashboard.messages.retry')),
      color: 'error'
    })
  }
}
</script>

<template>
  <div class="p-4 sm:p-6 space-y-6 sm:space-y-8 print:p-0 print:space-y-4 print:bg-white print:text-black min-w-0">
    <!-- Printable Document Header (Visible only during PDF Print) -->
    <div class="hidden print:block border-b-2 border-primary-600 pb-4 mb-6">
      <div class="flex justify-between items-center">
        <div>
          <h1 class="text-xl font-bold uppercase tracking-wider text-gray-900">
            {{ t('kpiPerformance.dashboard.print.division') }}
          </h1>
          <h2 class="text-lg font-semibold text-primary-700">
            {{ t('kpiPerformance.dashboard.print.reportTitle', { year }) }}
          </h2>
          <p class="text-md text-gray-500 mt-0.5">
            {{ t('kpiPerformance.dashboard.print.generatedOn', { date: formattedCurrentDate }) }}
          </p>
        </div>
        <div class="text-right text-md text-gray-500">
          <span class="font-bold text-gray-800">{{ t('kpiPerformance.dashboard.print.system') }}</span>
          <br />
          <span>{{ t('kpiPerformance.dashboard.print.confidential') }}</span>
        </div>
      </div>
    </div>

    <!-- Screen Header (Hidden during PDF Print) -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 print:hidden">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white">{{ t('kpiPerformance.dashboard.title') }}</h1>
        <p class="text-xs sm:text-sm font-semibold text-gray-500 mt-1">{{ t('kpiPerformance.dashboard.subtitle') }}</p>
      </div>
      <div class="flex flex-col sm:flex-row flex-wrap items-stretch sm:items-center gap-2 sm:gap-3 w-full md:w-auto">
        <!-- Period Selector -->
        <USelect
          v-model="selectedPeriod"
          :items="periodOptions"
          value-key="value"
          class="w-full sm:w-32"
          :placeholder="t('kpiPerformance.dashboard.periodPlaceholder')"
        />
        <!-- Year Selector -->
        <USelect
          v-model="year"
          :items="yearOptions"
          class="w-full sm:w-28"
        />
        <!-- Upload Laporan Kinerja Button -->
        <UButton
          :label="t('kpiPerformance.dashboard.importReport')"
          icon="i-lucide-upload"
          color="primary"
          class="w-full sm:w-auto justify-center"
          to="/kpi-performance/upload"
        />

        <!-- Export Button -->
        <UButton
          :label="t('kpiPerformance.dashboard.exportPdf')"
          icon="i-lucide-download"
          color="warning"
          variant="outline"
          class="font-bold shadow-sm w-full sm:w-auto justify-center"
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
              {{ t('kpiPerformance.dashboard.imported.title', { count: uploadStore.uploadedReports.length }) }}
            </div>
            <div class="text-md text-gray-600 dark:text-gray-400 mt-0.5">
              {{ t('kpiPerformance.dashboard.imported.active', { title: uploadStore.uploadedReports[0]?.title || '-', period: uploadStore.uploadedReports[0]?.period || '-', year: uploadStore.uploadedReports[0]?.year || '-' }) }}
            </div>
          </div>
        </div>
        <UButton
          :label="t('kpiPerformance.dashboard.imported.manage')"
          icon="i-lucide-arrow-right"
          color="primary"
          variant="subtle"
          size="md"
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


    <!-- Printable Sign-off Footer (Visible only during PDF Print) -->
    <div class="hidden print:grid grid-cols-3 gap-8 pt-8 mt-8 border-t border-gray-300 text-center text-md">
      <div>
        <p class="font-bold text-gray-700">{{ t('kpiPerformance.dashboard.print.preparedBy') }}</p>
        <div class="h-16"></div>
        <p class="font-semibold text-gray-900 border-t border-gray-400 pt-1">{{ t('kpiPerformance.dashboard.print.specialist') }}</p>
      </div>
      <div>
        <p class="font-bold text-gray-700">{{ t('kpiPerformance.dashboard.print.reviewedBy') }}</p>
        <div class="h-16"></div>
        <p class="font-semibold text-gray-900 border-t border-gray-400 pt-1">{{ t('kpiPerformance.dashboard.print.qualityManager') }}</p>
      </div>
      <div>
        <p class="font-bold text-gray-700">{{ t('kpiPerformance.dashboard.print.approvedBy') }}</p>
        <div class="h-16"></div>
        <p class="font-semibold text-gray-900 border-t border-gray-400 pt-1">{{ t('kpiPerformance.dashboard.print.cae') }}</p>
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
