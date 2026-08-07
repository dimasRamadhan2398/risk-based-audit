<script setup lang="ts">
import { usePerformanceStore } from '~/stores/performance'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { computed } from 'vue'

const store = usePerformanceStore()
const spStore = useStrategicPlanStore()

const findSpMetric = (keywords: string[]) => {
  if (!spStore.strategicObjectives || spStore.strategicObjectives.length === 0) return null
  return spStore.strategicObjectives.find((item: any) => {
    const kpiName = (item.kpi || item.strategicObjective || '').toLowerCase()
    return keywords.some(kw => kpiName.includes(kw.toLowerCase()))
  })
}

const cards = computed(() => {
  // 1. Audit Completion Rate
  const spCompletion = findSpMetric(['audit completion', 'completion rate', 'pkat'])
  const completionVal = spCompletion?.actual ? `${spCompletion.actual}${spCompletion.unit || '%'}` : '97%'
  const completionTarget = spCompletion?.target ? `${spCompletion.target}${spCompletion.unit || '%'}` : '90%'
  const completionActualNum = parseFloat(spCompletion?.actual || '97')
  const completionTargetNum = parseFloat(spCompletion?.target || '90')
  const completionGap = (completionActualNum - completionTargetNum).toFixed(1)
  const completionTrend = `${parseFloat(completionGap) >= 0 ? '+' : ''}${completionGap}% vs target`
  const completionTrendUp = completionActualNum >= completionTargetNum

  // 2. Report Timeliness
  const spTimeliness = findSpMetric(['report timeliness', 'timeliness', 'lha'])
  const timelinessVal = spTimeliness?.actual ? `${spTimeliness.actual}${spTimeliness.unit || '%'}` : '98%'
  const timelinessTarget = spTimeliness?.target ? `${spTimeliness.target}${spTimeliness.unit || '%'}` : '90%'
  const timelinessActualNum = parseFloat(spTimeliness?.actual || '98')
  const timelinessTargetNum = parseFloat(spTimeliness?.target || '90')
  const timelinessGap = (timelinessActualNum - timelinessTargetNum).toFixed(1)
  const timelinessTrend = `${parseFloat(timelinessGap) >= 0 ? '+' : ''}${timelinessGap}% vs target`
  const timelinessTrendUp = timelinessActualNum >= timelinessTargetNum

  // 3. Client Satisfaction
  const spCsat = findSpMetric(['client satisfaction', 'auditee satisfaction', 'csat'])
  const csatVal = spCsat?.actual ? (spCsat.unit === 'Score' ? `${spCsat.actual} / 5.0` : `${spCsat.actual}${spCsat.unit}`) : '4.7 / 5.0'
  const csatTarget = spCsat?.target ? (spCsat.unit === 'Score' ? `${spCsat.target} / 5.0` : `${spCsat.target}${spCsat.unit}`) : '4.5 / 5.0'
  const csatActualNum = parseFloat(spCsat?.actual || '4.7')
  const csatTargetNum = parseFloat(spCsat?.target || '4.5')
  const csatGap = (csatActualNum - csatTargetNum).toFixed(1)
  const csatTrend = `${parseFloat(csatGap) >= 0 ? '+' : ''}${csatGap} vs target`
  const csatTrendUp = csatActualNum >= csatTargetNum

  // 4. Action Plan Closed
  const spActionPlan = findSpMetric(['action plan', 'tindak lanjut', 'recommendation'])
  const actionPlanVal = spActionPlan?.actual ? `${spActionPlan.actual}${spActionPlan.unit || '%'}` : '87%'
  const actionPlanTarget = spActionPlan?.target ? `${spActionPlan.target}${spActionPlan.unit || '%'}` : '90%'
  const actionPlanActualNum = parseFloat(spActionPlan?.actual || '87')
  const actionPlanTargetNum = parseFloat(spActionPlan?.target || '90')
  const actionPlanGap = (actionPlanActualNum - actionPlanTargetNum).toFixed(1)
  const actionPlanTrend = `${parseFloat(actionPlanGap) >= 0 ? '+' : ''}${actionPlanGap}% vs target`
  const actionPlanTrendUp = actionPlanActualNum >= actionPlanTargetNum

  return [
    {
      title: 'Audit Completion Rate',
      value: completionVal,
      target: completionTarget,
      trend: completionTrend,
      trendUp: completionTrendUp,
      icon: 'i-lucide-target',
      iconColor: 'text-orange-500',
      iconBg: 'bg-orange-100 dark:bg-orange-900/30'
    },
    {
      title: 'Report Timeliness',
      value: timelinessVal,
      target: timelinessTarget,
      trend: timelinessTrend,
      trendUp: timelinessTrendUp,
      icon: 'i-lucide-clock',
      iconColor: 'text-orange-500',
      iconBg: 'bg-orange-100 dark:bg-orange-900/30'
    },
    {
      title: 'Client Satisfaction',
      value: csatVal,
      target: csatTarget,
      trend: csatTrend,
      trendUp: csatTrendUp,
      icon: 'i-lucide-check-square',
      iconColor: 'text-orange-500',
      iconBg: 'bg-orange-100 dark:bg-orange-900/30'
    },
    {
      title: 'Action Plan Closed',
      value: actionPlanVal,
      target: actionPlanTarget,
      trend: actionPlanTrend,
      trendUp: actionPlanTrendUp,
      icon: 'i-lucide-alert-circle',
      iconColor: 'text-orange-500',
      iconBg: 'bg-orange-100 dark:bg-orange-900/30'
    }
  ]
})
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
    <UCard v-for="(card, index) in cards" :key="index" :ui="{ body: 'p-6'}" class="bg-white dark:bg-gray-900 shadow-sm border border-gray-100 dark:border-gray-800">
      <div class="flex justify-between items-start mb-4">
        <div :class="['p-2 rounded-lg', card.iconBg]">
          <UIcon :name="card.icon" :class="['w-6 h-6', card.iconColor]" />
        </div>
        <div class="bg-success-100 dark:bg-success-900/30 text-success-600 dark:text-success-400 text-md font-semibold px-2 py-1 rounded">
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
