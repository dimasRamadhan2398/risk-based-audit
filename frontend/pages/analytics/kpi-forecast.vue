<script setup lang="ts">
import { useTimeSeriesData } from '~/composables/useAnalyticsData'

const initialTS = useTimeSeriesData()
</script>

<template>
  <div class="max-w-[1440px] mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <div class="mb-6 flex items-center justify-between">
      <div>
        <div class="flex items-center gap-2 text-md text-gray-400 mb-1">
          <NuxtLink to="/analytics" class="hover:underline">Analytics</NuxtLink>
          <span>/</span>
          <span class="text-indigo-600 font-bold">Sub-Feature 9</span>
        </div>
        <h1 class="text-2xl font-black text-gray-900 dark:text-white flex items-center gap-3">
          <UIcon name="i-heroicons-arrow-trending-up" class="text-violet-500 w-7 h-7" />
          KPI PyTorch LSTM Forecast
        </h1>
        <p class="text-md text-gray-500 mt-1">
          Model Deep Learning (PyTorch LSTM) memprediksi tren performa KPI masa depan berdasarkan deret waktu historis.
        </p>
      </div>
      <UBadge color="info" variant="subtle" size="lg">PyTorch LSTM Model</UBadge>
    </div>

    <UCard>
      <template #header>
        <h3 class="font-bold text-md">Ringkasan Forecast KPI (PyTorch LSTM)</h3>
      </template>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-200 dark:border-gray-700">
              <th class="text-left py-3 px-3 font-bold text-[10px] uppercase text-gray-400">Kode</th>
              <th class="text-left py-3 px-3 font-bold text-[10px] uppercase text-gray-400">KPI</th>
              <th class="text-center py-3 px-3 font-bold text-[10px] uppercase text-gray-400">Nilai Saat Ini</th>
              <th class="text-center py-3 px-3 font-bold text-[10px] uppercase text-gray-400">Nilai Forecast</th>
              <th class="text-center py-3 px-3 font-bold text-[10px] uppercase text-gray-400">Tren</th>
              <th class="text-left py-3 px-3 font-bold text-[10px] uppercase text-gray-400">Rekomendasi Tindakan Audit</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="kpi in initialTS.kpiForecasts" :key="kpi.code" class="border-b border-gray-100 dark:border-gray-800">
              <td class="py-3 px-3 font-mono font-bold text-violet-600">{{ kpi.code }}</td>
              <td class="py-3 px-3 font-bold">{{ kpi.kpiName }}</td>
              <td class="text-center py-3 px-3 font-mono">{{ kpi.currentValue }}%</td>
              <td class="text-center py-3 px-3 font-mono font-bold text-violet-600">{{ kpi.forecastedValue }}%</td>
              <td class="text-center py-3 px-3">
                <UBadge :color="kpi.trend === 'Improving' ? 'success' : 'error'" variant="subtle">{{ kpi.trend }}</UBadge>
              </td>
              <td class="py-3 px-3 text-md text-gray-600 dark:text-gray-300">{{ kpi.recommendedAction }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </UCard>
  </div>
</template>
