<script setup lang="ts">
import { ref, watch } from 'vue'
import KpiSummaryCards from '~/components/kpi-performance/KpiSummaryCards.vue'
import KpiCharts from '~/components/kpi-performance/KpiCharts.vue'
import KpiDetailedTable from '~/components/kpi-performance/KpiDetailedTable.vue'

import { usePerformanceStore } from '~/stores/performance'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'

const perfStore = usePerformanceStore()
const spStore = useStrategicPlanStore()

const year = ref('2024')
const yearOptions = ['2024', '2025', '2026', '2027']

perfStore.fetchKPIAchievements(parseInt(year.value))
perfStore.fetchWorkPlanRealizations(parseInt(year.value))
perfStore.fetchDashboardSummary(parseInt(year.value))
perfStore.fetchMonthlyTrends(parseInt(year.value))
spStore.fetchStrategicPlans()

watch(year, (newYear) => {
  const yr = parseInt(newYear)
  perfStore.fetchKPIAchievements(yr)
  perfStore.fetchWorkPlanRealizations(yr)
  perfStore.fetchDashboardSummary(yr)
  perfStore.fetchMonthlyTrends(yr)
  spStore.fetchStrategicPlans()
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
    <div class="flex items-start justify-between print:hidden">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">KPI Performance</h1>
        <p class="text-sm font-semibold text-gray-500 mt-1">Monitor and track internal audit performance metrics</p>
      </div>
      <div class="flex items-center gap-4">
        <USelectMenu
          v-model="year"
          :items="yearOptions"
          class="w-32"
        />
        <UButton
          label="Export PDF"
          icon="i-lucide-download"
          color="warning"
          class="font-bold shadow-sm"
          @click="exportPDF"
        />
      </div>
    </div>

    <!-- Summary Cards -->
    <KpiSummaryCards :year="parseInt(year)" />

    <!-- Charts -->
    <KpiCharts :year="parseInt(year)" />

    <!-- Detailed Table -->
    <KpiDetailedTable :year="parseInt(year)" />

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
