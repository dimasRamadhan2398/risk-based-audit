<script setup lang="ts">
import { computed } from 'vue'
import { usePerformanceStore } from '~/stores/performance'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement
} from 'chart.js'
import { Bar, Line } from 'vue-chartjs'

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
)

const props = defineProps({
  year: {
    type: Number,
    required: true
  }
})

const perfStore = usePerformanceStore()
const spStore = useStrategicPlanStore()

const findSpMetric = (keywords: string[]) => {
  if (!spStore.strategicObjectives || spStore.strategicObjectives.length === 0) return null
  return spStore.strategicObjectives.find((item: any) => {
    const kpiName = (item.kpi || item.strategicObjective || '').toLowerCase()
    return keywords.some(kw => kpiName.includes(kw.toLowerCase()))
  })
}

const barChartData = computed(() => {
  let labels = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun']
  let monthlyData = [85, 90, 95, 92, 98, 100]

  if (perfStore.monthlyTrends && perfStore.monthlyTrends.completion_rate_series && perfStore.monthlyTrends.completion_rate_series.length > 0) {
    labels = perfStore.monthlyTrends.labels
    monthlyData = perfStore.monthlyTrends.completion_rate_series
  } else {
    const spMonthly = findSpMetric(['monthly completion', 'completion rate', 'pkat'])
    const actualVal = parseFloat(spMonthly?.actual || '95')
    monthlyData = [
      Math.round(actualVal * 0.88),
      Math.round(actualVal * 0.92),
      Math.round(actualVal * 0.95),
      Math.round(actualVal * 0.93),
      Math.round(actualVal * 0.98),
      Math.min(100, Math.round(actualVal))
    ]
  }

  return {
    labels,
    datasets: [
      {
        label: 'Monthly Completion Rate',
        backgroundColor: '#4D00FF',
        borderRadius: 4,
        data: monthlyData,
        barPercentage: 0.6
      }
    ]
  }
})

const barChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false
    }
  },
  scales: {
    y: {
      beginAtZero: true,
      max: 100,
      ticks: {
        stepSize: 25,
        color: '#9CA3AF'
      },
      grid: {
        color: '#F3F4F6',
        drawBorder: false
      }
    },
    x: {
      grid: {
        display: false,
        drawBorder: false
      },
      ticks: {
        color: '#6B7280'
      }
    }
  }
}

const lineChartData = computed(() => {
  let labels = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun']
  let timelinessData = [85, 87, 90, 88, 92, 98]
  let csatData = [4.2, 4.3, 4.5, 4.4, 4.6, 4.7]

  if (perfStore.monthlyTrends && perfStore.monthlyTrends.timeliness_series && perfStore.monthlyTrends.timeliness_series.length > 0) {
    labels = perfStore.monthlyTrends.labels
    timelinessData = perfStore.monthlyTrends.timeliness_series
    csatData = perfStore.monthlyTrends.csat_series
  } else {
    const spTimeliness = findSpMetric(['report timeliness', 'timeliness', 'lha'])
    const timelinessActual = parseFloat(spTimeliness?.actual || '98')

    const spCsat = findSpMetric(['client satisfaction', 'auditee satisfaction', 'csat'])
    const csatActual = parseFloat(spCsat?.actual || '4.7')

    timelinessData = [
      Math.round(timelinessActual * 0.88),
      Math.round(timelinessActual * 0.90),
      Math.round(timelinessActual * 0.93),
      Math.round(timelinessActual * 0.91),
      Math.round(timelinessActual * 0.96),
      Math.min(100, Math.round(timelinessActual))
    ]

    csatData = [
      parseFloat((csatActual * 0.90).toFixed(1)),
      parseFloat((csatActual * 0.92).toFixed(1)),
      parseFloat((csatActual * 0.96).toFixed(1)),
      parseFloat((csatActual * 0.94).toFixed(1)),
      parseFloat((csatActual * 0.98).toFixed(1)),
      Math.min(5.0, parseFloat(csatActual.toFixed(1)))
    ]
  }

  return {
    labels,
    datasets: [
      {
        label: 'Timeliness (%)',
        borderColor: '#10B981',
        backgroundColor: '#10B981',
        pointBackgroundColor: '#10B981',
        pointBorderColor: '#10B981',
        pointBorderWidth: 2,
        pointRadius: 4,
        data: timelinessData,
        yAxisID: 'y'
      },
      {
        label: 'CSAT Score',
        borderColor: '#F97316',
        backgroundColor: '#F97316',
        pointBackgroundColor: '#F97316',
        pointBorderColor: '#F97316',
        pointBorderWidth: 2,
        pointRadius: 4,
        data: csatData,
        yAxisID: 'y1'
      }
    ]
  }
})

const lineChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false
    }
  },
  scales: {
    y: {
      beginAtZero: false,
      min: 50,
      max: 100,
      ticks: {
        stepSize: 10,
        color: '#9CA3AF'
      },
      grid: {
        color: '#F3F4F6',
        borderDash: [5, 5],
        drawBorder: false
      }
    },
    y1: {
      position: 'right' as const,
      beginAtZero: true,
      max: 5,
      ticks: {
        stepSize: 1,
        color: '#9CA3AF'
      },
      grid: {
        display: false,
        drawBorder: false
      }
    },
    x: {
      grid: {
        display: false,
        drawBorder: false
      },
      ticks: {
        color: '#6B7280'
      }
    }
  }
}
</script>

<template>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <!-- Bar Chart -->
    <UCard :ui="{ body: 'p-6' }" class="bg-white dark:bg-gray-900 shadow-sm border border-gray-100 dark:border-gray-800">
      <div class="flex items-center gap-2 mb-6">
        <UIcon name="i-lucide-bar-chart-2" class="w-5 h-5 text-gray-400" />
        <h3 class="text-base font-semibold text-gray-900 dark:text-white">Monthly Completion Rate</h3>
      </div>
      <div class="h-64">
        <Bar :data="barChartData" :options="barChartOptions" />
      </div>
    </UCard>

    <!-- Line Chart -->
    <UCard :ui="{ body: 'p-6' }" class="bg-white dark:bg-gray-900 shadow-sm border border-gray-100 dark:border-gray-800">
      <div class="flex flex-col mb-6">
        <div class="flex items-center gap-2">
          <UIcon name="i-lucide-activity" class="w-5 h-5 text-gray-400" />
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">CSAT & Timeliness Trend</h3>
        </div>
        <div class="flex items-center justify-center gap-6 mt-2">
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
            <span class="text-md font-semibold text-emerald-500">Timeliness (%)</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-orange-500"></span>
            <span class="text-md font-semibold text-orange-500">CSAT Score</span>
          </div>
        </div>
      </div>
      <div class="h-56">
        <Line :data="lineChartData" :options="lineChartOptions" />
      </div>
    </UCard>
  </div>
</template>
