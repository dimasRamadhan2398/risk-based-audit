<script setup lang="ts">
import { usePerformanceStore } from '~/stores/performance'
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { buildKpiSummaryCards } from '~/utils/kpiPerformanceDisplay'

defineProps({
  year: {
    type: Number,
    required: true
  }
})

const { t } = useI18n()
const store = usePerformanceStore()

const ICONS: Record<string, string> = {
  audit_completion_rate: 'i-lucide-target',
  report_timeliness: 'i-lucide-clock',
  client_satisfaction: 'i-lucide-check-square',
  action_plan_closed: 'i-lucide-alert-circle'
}

// Only GET /performance/dashboard-summary data (titles and sub-metric texts translated, see
// utils/kpiPerformanceDisplay.ts). Without it each card shows "-" and a "no data" note, never sample numbers.
const cards = computed(() => buildKpiSummaryCards(t, store.dashboardCards).map(card => ({
  ...card,
  icon: ICONS[card.key] || 'i-lucide-target',
  iconColor: 'text-orange-500',
  iconBg: 'bg-orange-100 dark:bg-orange-900/30'
})))
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
    <UCard v-for="(card, index) in cards" :key="index" :ui="{ body: 'p-6'}" class="bg-white dark:bg-gray-900 shadow-sm border border-gray-100 dark:border-gray-800">
      <div class="flex justify-between items-start mb-4">
        <div :class="['p-2 rounded-lg', card.iconBg]">
          <UIcon :name="card.icon" :class="['w-6 h-6', card.iconColor]" />
        </div>
        <div
          v-if="card.hasData"
          class="bg-success-100 dark:bg-success-900/30 text-success-600 dark:text-success-400 text-md font-semibold px-2 py-1 rounded"
        >
          {{ t('kpiPerformance.summary.targetBadge', { value: card.target }) }}
        </div>
      </div>
      <div class="space-y-1">
        <p class="text-sm text-gray-500 dark:text-gray-400 font-medium">{{ card.title }}</p>
        <p class="text-4xl font-bold">{{ card.value }}</p>
        <p
          v-if="card.trend"
          :class="['text-sm font-semibold pt-2', card.trendUp === null ? 'text-gray-500 dark:text-gray-400' : card.trendUp ? 'text-success-600 dark:text-success-400' : 'text-rose-600 dark:text-rose-400']"
        >
          {{ card.trend }}
        </p>
        <p
          v-else-if="!card.hasData"
          class="text-sm font-medium pt-2 text-gray-500 dark:text-gray-400"
        >
          {{ t('kpiPerformance.summary.noData') }}
        </p>
      </div>

      <!-- Tiered sub-metrics details -->
      <div v-if="card.subMetrics && card.subMetrics.length > 0" class="mt-4 pt-4 border-t border-gray-100 dark:border-gray-800 space-y-3">
        <div v-for="(sub, sIdx) in card.subMetrics" :key="sIdx" class="space-y-1 bg-gray-50 dark:bg-gray-800/40 p-2.5 rounded-lg border border-gray-100 dark:border-gray-800/80">
          <div class="flex justify-between items-center text-md font-semibold text-gray-400 dark:text-gray-500 gap-2">
            <span class="truncate" :title="sub.title">{{ sub.title }}</span>
            <span class="bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 px-1.5 py-0.5 rounded text-[10px] shrink-0">{{ sub.target }}</span>
          </div>
          <div class="flex justify-between items-center gap-2">
            <span class="text-lg font-bold text-gray-800 dark:text-gray-100">{{ sub.value }}</span>
            <UTooltip v-if="sub.trend" :text="sub.trend">
              <UIcon name="i-lucide-info" class="w-4 h-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 cursor-pointer transition-colors shrink-0" />
            </UTooltip>
          </div>
        </div>
      </div>
    </UCard>
  </div>
</template>
