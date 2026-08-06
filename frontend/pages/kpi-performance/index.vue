<script setup lang="ts">
import { ref } from 'vue'
import KpiSummaryCards from '~/components/kpi-performance/KpiSummaryCards.vue'
import KpiCharts from '~/components/kpi-performance/KpiCharts.vue'
import KpiDetailedTable from '~/components/kpi-performance/KpiDetailedTable.vue'

import { usePerformanceStore } from '~/stores/performance'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { watch } from 'vue'

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
  useToast().add({
    title: 'Exporting...',
    description: 'KPI Performance report is being exported to PDF.',
    color: 'success'
  })
}
</script>

<template>
  <div class="p-6 space-y-8">
    <div class="flex items-start justify-between">
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
  </div>
</template>
