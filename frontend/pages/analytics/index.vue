<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Line, Bar, Doughnut } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, BarElement, ArcElement, Title, Tooltip, Legend } from 'chart.js'
import { format } from 'date-fns'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement, ArcElement, Title, Tooltip, Legend)

const config = useRuntimeConfig()

// Analytics Service Base URL
const ANALYTICS_API_URL = config.public.analyticsApiBase

// State
const loading = ref(true)
const error = ref('')

const reportData = ref<any>(null)
const predictData = ref<any>(null)

// Chart Configurations
const lineChartData = ref<{ labels: string[], datasets: any[] }>({
  labels: [],
  datasets: []
})

const doughnutChartData = ref<{ labels: string[], datasets: any[] }>({
  labels: ['Resolved', 'Open', 'Overdue Follow Up'],
  datasets: []
})

const predictChartData = ref<{ labels: string[], datasets: any[] }>({
  labels: [],
  datasets: []
})

const fetchAnalytics = async () => {
  try {
    loading.value = true

    // Dummy data fallback
    const dummyReportData = {
      total_findings: 145,
      resolved: 92,
      open: 38,
      overdue_follow_up: 15,
      finding_trends: [
        { month: 'Jan', count: 12 },
        { month: 'Feb', count: 19 },
        { month: 'Mar', count: 15 },
        { month: 'Apr', count: 22 },
        { month: 'May', count: 18 },
        { month: 'Jun', count: 25 }
      ],
      anomalies: [
        { severity: 'High', description: 'Unusual spike in financial operational cost findings', date: '2026-04-15' },
        { severity: 'Medium', description: 'Multiple recurring IT compliance issues in Q2', date: '2026-05-02' }
      ]
    }

    const dummyPredictData = {
      forecast: [
        { date: '2026-07-01T00:00:00Z', predicted_risk: 3.5 },
        { date: '2026-08-01T00:00:00Z', predicted_risk: 4.1 },
        { date: '2026-09-01T00:00:00Z', predicted_risk: 3.8 },
        { date: '2026-10-01T00:00:00Z', predicted_risk: 4.5 },
        { date: '2026-11-01T00:00:00Z', predicted_risk: 5.2 },
        { date: '2026-12-01T00:00:00Z', predicted_risk: 4.9 }
      ],
      trend_direction: 'Up',
      model_accuracy: 0.87
    }

    // Use Nuxt useFetch with fallback to dummy data
    const [resReport, resPredict] = await Promise.all([
      $fetch(`${ANALYTICS_API_URL}/report`).catch(e => {
        console.warn('Backend /report failed, using dummy data', e)
        return { data: dummyReportData }
      }),
      $fetch(`${ANALYTICS_API_URL}/predict`).catch(e => {
        console.warn('Backend /predict failed, using dummy data', e)
        return { data: dummyPredictData }
      })
    ])

    reportData.value = (resReport as any)?.data || dummyReportData
    predictData.value = (resPredict as any)?.data || dummyPredictData

    setupCharts()
  } catch (err: any) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const setupCharts = () => {
  if (reportData.value) {
    // Setup Trend Line Chart
    lineChartData.value = {
      labels: reportData.value.finding_trends.map((t: any) => t.month),
      datasets: [
        {
          label: 'Findings Trend',
          backgroundColor: '#3b82f6',
          borderColor: '#3b82f6',
          data: reportData.value.finding_trends.map((t: any) => t.count),
          tension: 0.4
        }
      ]
    }

    // Setup Status Doughnut Chart
    doughnutChartData.value = {
      labels: ['Resolved', 'Open', 'Overdue Follow Up'],
      datasets: [
        {
          backgroundColor: ['#10b981', '#f59e0b', '#ef4444'],
          data: [
            reportData.value.resolved,
            reportData.value.open,
            reportData.value.overdue_follow_up
          ]
        }
      ]
    }
  }

  if (predictData.value) {
    // Setup Prediction Line Chart
    predictChartData.value = {
      labels: predictData.value.forecast.map((f: any) => format(new Date(f.date), 'MMM yyyy')),
      datasets: [
        {
          label: 'Predicted Risk Trend',
          backgroundColor: '#8b5cf6',
          borderColor: '#8b5cf6',
          data: predictData.value.forecast.map((f: any) => f.predicted_risk.toFixed(2)),
          tension: 0.4,
          borderDash: [5, 5] // Dashed line for forecast
        }
      ]
    }
  }
}

onMounted(() => {
  fetchAnalytics()
})
</script>

<template>
  <div class="p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Data Analytics & Predictive AI</h1>
      <p class="text-gray-500 dark:text-gray-400">Risk-based audit insights and forecasting</p>
    </div>

    <div v-if="loading" class="flex justify-center items-center h-64">
      <UIcon name="i-lucide-loader-2" class="w-8 h-8 animate-spin text-primary-500" />
    </div>

    <div v-else-if="error" class="bg-red-50 text-red-500 p-4 rounded-lg">
      Error: {{ error }}. Please ensure the analytics-service is running.
    </div>

    <div v-else class="space-y-6">
      <!-- Top Stats Cards -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <UCard>
          <div class="text-gray-500 text-sm">Total Findings</div>
          <div class="text-3xl font-bold">{{ reportData?.total_findings }}</div>
        </UCard>
        <UCard>
          <div class="text-gray-500 text-sm">Resolved</div>
          <div class="text-3xl font-bold text-emerald-500">{{ reportData?.resolved }}</div>
        </UCard>
        <UCard>
          <div class="text-gray-500 text-sm">Open</div>
          <div class="text-3xl font-bold text-amber-500">{{ reportData?.open }}</div>
        </UCard>
        <UCard>
          <div class="text-gray-500 text-sm">Overdue Follow Up</div>
          <div class="text-3xl font-bold text-rose-500">{{ reportData?.overdue_follow_up }}</div>
        </UCard>
      </div>

      <!-- Charts Row 1: Historical Patterns -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <UCard>
          <template #header>
            <h3 class="font-semibold">Audit Finding Trends</h3>
          </template>
          <div class="h-64 flex justify-center">
            <Line v-if="lineChartData.datasets.length" :data="lineChartData" :options="{ responsive: true, maintainAspectRatio: false }" />
          </div>
        </UCard>

        <UCard>
          <template #header>
            <h3 class="font-semibold">Findings Status Distribution</h3>
          </template>
          <div class="h-64 flex justify-center">
            <Doughnut v-if="doughnutChartData.datasets.length" :data="doughnutChartData" :options="{ responsive: true, maintainAspectRatio: false }" />
          </div>
        </UCard>
      </div>

      <!-- Predictive AI Insights -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <!-- Forecast Chart -->
        <UCard class="md:col-span-2">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-lucide-sparkles" class="text-purple-500" />
              <h3 class="font-semibold">AI Predictive Risk Forecast</h3>
            </div>
          </template>
          <div class="h-64 flex justify-center">
            <Line v-if="predictChartData.datasets.length" :data="predictChartData" :options="{ responsive: true, maintainAspectRatio: false }" />
          </div>
        </UCard>

        <!-- Insights Sidebar -->
        <div class="space-y-6">
          <UCard>
            <template #header>
              <h3 class="font-semibold">Forecast Summary</h3>
            </template>
            <div class="space-y-4">
              <div>
                <span class="text-gray-500 text-sm block">Trend Direction</span>
                <div class="flex items-center gap-2 mt-1">
                  <UIcon
                    :name="predictData?.trend_direction === 'Up' ? 'i-lucide-trending-up' : predictData?.trend_direction === 'Down' ? 'i-lucide-trending-down' : 'i-lucide-minus'"
                    :class="predictData?.trend_direction === 'Up' ? 'text-rose-500' : predictData?.trend_direction === 'Down' ? 'text-emerald-500' : 'text-gray-500'"
                  />
                  <span class="font-bold">{{ predictData?.trend_direction }}</span>
                </div>
              </div>
              <div>
                <span class="text-gray-500 text-sm block">Model Accuracy</span>
                <span class="font-bold">{{ (predictData?.model_accuracy * 100).toFixed(0) }}%</span>
              </div>
            </div>
          </UCard>

          <UCard>
            <template #header>
              <h3 class="font-semibold">Detected Anomalies</h3>
            </template>
            <div class="space-y-3">
              <div v-for="(anomaly, idx) in reportData?.anomalies" :key="idx" class="border-l-4 p-3 bg-gray-50 dark:bg-gray-800" :class="anomaly.severity === 'High' ? 'border-rose-500' : 'border-amber-500'">
                <p class="text-sm font-semibold">{{ anomaly.severity }} Severity</p>
                <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">{{ anomaly.description }}</p>
                <p class="text-xs text-gray-400 mt-1">{{ anomaly.date }}</p>
              </div>
            </div>
          </UCard>
        </div>
      </div>
    </div>
  </div>
</template>