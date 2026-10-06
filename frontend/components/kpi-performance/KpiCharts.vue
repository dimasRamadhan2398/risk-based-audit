<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { kpiMonthLabel } from '~/utils/kpiPerformanceLabels'
import { hasSeriesData, kpiChartSeries } from '~/utils/kpiPerformanceDisplay'
import { usePerformanceStore } from '~/stores/performance'
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

const { t } = useI18n()
const perfStore = usePerformanceStore()

// Month labels may come from the API as English abbreviations; translate known ones, keep others as-is.
const monthLabel = (label: string) => kpiMonthLabel(t, label)

// Only GET /performance/monthly-trends data; with no series the chart shows an empty state instead.
const series = computed(() => kpiChartSeries(perfStore.monthlyTrends))
const hasCompletionData = computed(() => hasSeriesData(series.value.completion))
const hasTrendData = computed(() => hasSeriesData(series.value.timeliness, series.value.csat))

const barChartData = computed(() => ({
  labels: series.value.labels.map(monthLabel),
  datasets: [
    {
      label: t('kpiPerformance.charts.monthlyCompletionRate'),
      backgroundColor: '#4D00FF',
      borderRadius: 4,
      data: series.value.completion,
      barPercentage: 0.6
    }
  ]
}))

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

const lineChartData = computed(() => ({
  labels: series.value.labels.map(monthLabel),
  datasets: [
    {
      label: t('kpiPerformance.charts.timeliness'),
      borderColor: '#10B981',
      backgroundColor: '#10B981',
      pointBackgroundColor: '#10B981',
      pointBorderColor: '#10B981',
      pointBorderWidth: 2,
      pointRadius: 4,
      data: series.value.timeliness,
      yAxisID: 'y'
    },
    {
      label: t('kpiPerformance.charts.csatScore'),
      borderColor: '#F97316',
      backgroundColor: '#F97316',
      pointBackgroundColor: '#F97316',
      pointBorderColor: '#F97316',
      pointBorderWidth: 2,
      pointRadius: 4,
      data: series.value.csat,
      yAxisID: 'y1'
    }
  ]
}))

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
        <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('kpiPerformance.charts.monthlyCompletionRate') }}</h3>
      </div>
      <div class="h-64">
        <Bar
          v-if="hasCompletionData"
          :data="barChartData"
          :options="barChartOptions"
        />
        <div
          v-else
          class="h-full flex flex-col items-center justify-center gap-2 text-center"
        >
          <UIcon
            :name="perfStore.loading ? 'i-lucide-loader-2' : 'i-lucide-bar-chart-2'"
            :class="['w-8 h-8 text-gray-400', perfStore.loading && 'animate-spin']"
          />
          <p class="text-sm font-medium text-gray-500">
            {{ perfStore.loading ? t('kpiPerformance.table.loading') : t('kpiPerformance.charts.noData', { year: props.year }) }}
          </p>
        </div>
      </div>
    </UCard>

    <!-- Line Chart -->
    <UCard :ui="{ body: 'p-6' }" class="bg-white dark:bg-gray-900 shadow-sm border border-gray-100 dark:border-gray-800">
      <div class="flex flex-col mb-6">
        <div class="flex items-center gap-2">
          <UIcon name="i-lucide-activity" class="w-5 h-5 text-gray-400" />
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('kpiPerformance.charts.trendTitle') }}</h3>
        </div>
        <div class="flex items-center justify-center gap-6 mt-2">
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
            <span class="text-md font-semibold text-emerald-500">{{ t('kpiPerformance.charts.timeliness') }}</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-orange-500"></span>
            <span class="text-md font-semibold text-orange-500">{{ t('kpiPerformance.charts.csatScore') }}</span>
          </div>
        </div>
      </div>
      <div class="h-56">
        <Line
          v-if="hasTrendData"
          :data="lineChartData"
          :options="lineChartOptions"
        />
        <div
          v-else
          class="h-full flex flex-col items-center justify-center gap-2 text-center"
        >
          <UIcon
            :name="perfStore.loading ? 'i-lucide-loader-2' : 'i-lucide-activity'"
            :class="['w-8 h-8 text-gray-400', perfStore.loading && 'animate-spin']"
          />
          <p class="text-sm font-medium text-gray-500">
            {{ perfStore.loading ? t('kpiPerformance.table.loading') : t('kpiPerformance.charts.noData', { year: props.year }) }}
          </p>
        </div>
      </div>
    </UCard>
  </div>
</template>
