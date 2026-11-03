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
// Layout follows the dashboard stat cards (components/dashboard/AdminDashboard.vue): title + icon tile,
// value, one secondary line.
const cards = computed(() => buildKpiSummaryCards(t, store.dashboardCards).map(card => ({
  ...card,
  icon: ICONS[card.key] || 'i-lucide-target'
})))

const trendClass = (trendUp: boolean | null) =>
  trendUp === null ? 'text-slate-500 dark:text-slate-300' : trendUp ? 'text-emerald-500' : 'text-red-500'
</script>

<template>
  <div class="grid grid-cols-1 gap-4 sm:gap-6 sm:grid-cols-2 lg:grid-cols-4">
    <div
      v-for="card in cards"
      :key="card.key"
      data-kpi-card
      class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card"
    >
      <div class="flex justify-between items-start gap-3">
        <span class="text-md font-semibold text-slate-500 dark:text-slate-200">{{ card.title }}</span>
        <div class="rounded-xl bg-primary-50 p-2.5 flex items-center justify-center shrink-0 dark:bg-primary-500/10">
          <UIcon
            :name="card.icon"
            class="text-primary-500 size-5 dark:text-primary-400"
          />
        </div>
      </div>
      <div class="mt-1">
        <h4 class="text-2xl font-bold text-slate-900 dark:text-white">
          {{ card.value }}
        </h4>
        <p class="flex flex-col gap-1 text-[11px] mt-1 text-slate-400 dark:text-slate-300 font-normal">
          <template v-if="card.hasData">
            <span
              v-if="card.trend"
              :class="['font-semibold', trendClass(card.trendUp)]"
            >{{ card.trend }}</span>
            <span>{{ t('kpiPerformance.summary.targetBadge', { value: card.target }) }}</span>
          </template>
          <span v-else>{{ t('kpiPerformance.summary.noData') }}</span>
        </p>
      </div>
    </div>
  </div>
</template>
