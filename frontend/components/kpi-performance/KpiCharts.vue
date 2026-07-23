<script setup lang="ts">

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

const barChartData = {
  labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
  datasets: [
    {
      label: 'Monthly Completion Rate',
      backgroundColor: '#4D00FF', // matches secondary color from previous tasks
      borderRadius: 4,
      data: [85, 90, 95, 92, 98, 100],
      barPercentage: 0.6
    }
  ]
}

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

const lineChartData = {
  labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
  datasets: [
    {
      label: 'Timeliness (%)',
      borderColor: '#10B981', // emerald-500
      backgroundColor: '#10B981',
      pointBackgroundColor: '#10B981',
      pointBorderColor: '#FF5C02',
      pointBorderWidth: 2,
      pointRadius: 4,
      data: [85, 87, 90, 88, 92, 95], 
      yAxisID: 'y'
    }
  ]
}

const lineChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false // We will build a custom legend to match the image
    }
  },
  scales: {
    y: {
      beginAtZero: false,
      min: 80,
      max: 100,
      ticks: {
        stepSize: 5,
        color: '#9CA3AF'
      },
      grid: {
        color: '#F3F4F6',
        borderDash: [5, 5],
        drawBorder: false
      }
    },
    y1: {
      position: 'right',
      beginAtZero: true,
      max: 5,
      ticks: {
        stepSize: 2.5, // To show 0, 2, 5
        color: '#9CA3AF',
        callback: function(value: any) {
          if (value === 2.5) return '2';
          return value;
        }
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
        <Line :data="lineChartData" :items="lineChartOptions" />
      </div>
    </UCard>
  </div>
</template>
