<script setup lang="ts">
const cards = [
  {
    title: 'Audit Completion Rate',
    value: '97%',
    target: '90%',
    trend: '6.9% from last month',
    trendUp: true,
    icon: 'i-lucide-target',
    iconColor: 'text-orange-500',
    iconBg: 'bg-orange-100 dark:bg-orange-900/30'
  },
  {
    title: 'Report Timeliness',
    value: '98%',
    target: '90%',
    trend: '6.9% from last month',
    trendUp: true,
    icon: 'i-lucide-clock',
    iconColor: 'text-orange-500',
    iconBg: 'bg-orange-100 dark:bg-orange-900/30'
  },
  {
    title: 'Client Satisfaction',
    value: '4.7 / 5.0',
    target: '90%',
    trend: '6.9% from last month',
    trendUp: true,
    icon: 'i-lucide-check-square',
    iconColor: 'text-orange-500',
    iconBg: 'bg-orange-100 dark:bg-orange-900/30'
  },
  {
    title: 'Action Plan Closed',
    value: '87%',
    target: '90%',
    trend: '-3.0 from last month',
    trendUp: false,
    icon: 'i-lucide-alert-circle',
    iconColor: 'text-orange-500',
    iconBg: 'bg-orange-100 dark:bg-orange-900/30'
  }
]
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
    <UCard v-for="(card, index) in cards" :key="index" :ui="{ body: 'p-6'}" class="bg-white dark:bg-gray-900 shadow-sm border border-gray-100 dark:border-gray-800">
      <div class="flex justify-between items-start mb-4">
        <div :class="['p-2 rounded-lg', card.iconBg]">
          <UIcon :name="card.icon" :class="['w-6 h-6', card.iconColor]" />
        </div>
        <div class="bg-success-100 dark:bg-success-900/30 text-success-600 dark:text-success-400 text-xs font-semibold px-2 py-1 rounded">
          Target: {{ card.target }}
        </div>
      </div>
      <div class="space-y-1">
        <p class="text-sm text-gray-500 dark:text-gray-400 font-medium">{{ card.title }}</p>
        <p class="text-4xl font-bold">{{ card.value }}</p>
        <p class="['text-lg font-bold pt-2', card.trendUp ? 'text-success-500' : 'text-gray-400']">
          {{ card.trend }}
        </p>
      </div>
    </UCard>
  </div>
</template>
