<script setup lang="ts">
import { usePerformanceStore } from '~/stores/performance'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { computed } from 'vue'

const props = defineProps({
  year: {
    type: Number,
    required: true
  }
})

const store = usePerformanceStore()
const spStore = useStrategicPlanStore()

const findSpMetric = (keywords: string[]) => {
  if (!spStore.strategicObjectives || spStore.strategicObjectives.length === 0) return null
  return spStore.strategicObjectives.find((item: any) => {
    const kpiName = (item.kpi || item.strategicObjective || '').toLowerCase()
    return keywords.some(kw => kpiName.includes(kw.toLowerCase()))
  })
}

const getTargetValueForYear = (metric: any, defaultVal: number): number => {
  if (metric && metric.kpiTargets) {
    const tgt = metric.kpiTargets[props.year] || metric.kpiTargets[String(props.year)]
    if (tgt) return parseFloat(tgt)
  }
  return metric?.target ? parseFloat(metric.target) : defaultVal
}

const getTargetForYear = (metric: any, defaultVal: string): string => {
  if (metric && metric.kpiTargets) {
    const tgt = metric.kpiTargets[props.year] || metric.kpiTargets[String(props.year)]
    if (tgt) {
      if (metric.unit === 'Score') {
        return `${parseFloat(tgt).toFixed(1)} / 5.0`
      }
      return `${tgt}${metric.unit || '%'}`
    }
  }
  return metric?.target ? (metric.unit === 'Score' ? `${parseFloat(metric.target).toFixed(1)} / 5.0` : `${metric.target}${metric.unit || '%'}`) : defaultVal
}

const cards = computed(() => {
  const iconMap: Record<string, { icon: string; color: string; bg: string }> = {
    audit_completion_rate: { icon: 'i-lucide-target', color: 'text-orange-500', bg: 'bg-orange-100 dark:bg-orange-900/30' },
    report_timeliness: { icon: 'i-lucide-clock', color: 'text-orange-500', bg: 'bg-orange-100 dark:bg-orange-900/30' },
    client_satisfaction: { icon: 'i-lucide-check-square', color: 'text-orange-500', bg: 'bg-orange-100 dark:bg-orange-900/30' },
    action_plan_closed: { icon: 'i-lucide-alert-circle', color: 'text-orange-500', bg: 'bg-orange-100 dark:bg-orange-900/30' }
  }

  if (store.dashboardCards && store.dashboardCards.length > 0) {
    return store.dashboardCards.map(c => {
      const iconConfig = iconMap[c.key] || { icon: 'i-lucide-target', color: 'text-orange-500', bg: 'bg-orange-100 dark:bg-orange-900/30' }
      return {
        title: c.title,
        value: c.value,
        target: c.target,
        trend: c.trend,
        trendUp: c.trend_up,
        icon: iconConfig.icon,
        iconColor: iconConfig.color,
        iconBg: iconConfig.bg,
        subMetrics: c.sub_metrics
      }
    })
  }

  // 1. Audit Completion Rate
  const spCompletion = findSpMetric(['audit completion', 'completion rate', 'pkat'])
  const completionVal = spCompletion?.actual ? `${spCompletion.actual}${spCompletion.unit || '%'}` : '92%'
  const completionTarget = getTargetForYear(spCompletion, '90%')
  const completionActualNum = parseFloat(spCompletion?.actual || '92')
  const completionTargetNum = getTargetValueForYear(spCompletion, 90)
  const completionGap = (completionActualNum - completionTargetNum).toFixed(1)
  const completionTrend = `${parseFloat(completionGap) >= 0 ? '+' : ''}${completionGap}% vs target`
  const completionTrendUp = completionActualNum >= completionTargetNum

  // 2. Report Timeliness
  const spTimeliness = findSpMetric(['report timeliness', 'timeliness', 'lha'])
  const timelinessVal = spTimeliness?.actual ? `${spTimeliness.actual}${spTimeliness.unit || '%'}` : '98%'
  const timelinessTarget = getTargetForYear(spTimeliness, '90%')
  const timelinessActualNum = parseFloat(spTimeliness?.actual || '98')
  const timelinessTargetNum = getTargetValueForYear(spTimeliness, 90)
  const timelinessGap = (timelinessActualNum - timelinessTargetNum).toFixed(1)
  const timelinessTrend = `${parseFloat(timelinessGap) >= 0 ? '+' : ''}${timelinessGap}% vs target`
  const timelinessTrendUp = timelinessActualNum >= timelinessTargetNum

  // 3. Client Satisfaction
  const spCsat = findSpMetric(['client satisfaction', 'auditee satisfaction', 'csat'])
  const csatVal = spCsat?.actual ? (spCsat.unit === 'Score' ? `${spCsat.actual} / 5.0` : `${spCsat.actual}${spCsat.unit}`) : '4.0 / 5.0'
  const csatTarget = getTargetForYear(spCsat, '4.0 / 5.0')
  const csatActualNum = parseFloat(spCsat?.actual || '4.0')
  const csatTargetNum = getTargetValueForYear(spCsat, 4.0)
  const csatGap = (csatActualNum - csatTargetNum).toFixed(1)
  const csatTrend = `${parseFloat(csatGap) >= 0 ? '+' : ''}${csatGap} vs target`
  const csatTrendUp = csatActualNum >= csatTargetNum

  // 4. Action Plan Closed
  const spActionPlan = findSpMetric(['action plan', 'tindak lanjut', 'recommendation'])
  const actionPlanVal = spActionPlan?.actual ? `${spActionPlan.actual}${spActionPlan.unit || '%'}` : '87%'
  const actionPlanTarget = getTargetForYear(spActionPlan, '90%')
  const actionPlanActualNum = parseFloat(spActionPlan?.actual || '87')
  const actionPlanTargetNum = getTargetValueForYear(spActionPlan, 90)
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
      iconBg: 'bg-orange-100 dark:bg-orange-900/30',
      subMetrics: [
        {
          title: 'Operational Completion Rate (Started Audits)',
          value: '92.0%',
          target: '90%',
          trend: '11 completed / 12 started'
        }
      ]
    },
    {
      title: 'Report Timeliness',
      value: timelinessVal,
      target: timelinessTarget,
      trend: timelinessTrend,
      trendUp: timelinessTrendUp,
      icon: 'i-lucide-clock',
      iconColor: 'text-orange-500',
      iconBg: 'bg-orange-100 dark:bg-orange-900/30',
      subMetrics: [
        {
          title: 'Avg Drafting Cycle-Time',
          value: '12.5 days',
          target: '< 14 days',
          trend: '-1.5 days vs target'
        }
      ]
    },
    {
      title: 'Client Satisfaction',
      value: csatVal,
      target: csatTarget,
      trend: csatTrend,
      trendUp: csatTrendUp,
      icon: 'i-lucide-check-square',
      iconColor: 'text-orange-500',
      iconBg: 'bg-orange-100 dark:bg-orange-900/30',
      subMetrics: [
        {
          title: 'Survey Response Rate',
          value: '85.5%',
          target: '80%',
          trend: '12 responses from 14 completed audits'
        }
      ]
    },
    {
      title: 'Action Plan Closed',
      value: actionPlanVal,
      target: actionPlanTarget,
      trend: actionPlanTrend,
      trendUp: actionPlanTrendUp,
      icon: 'i-lucide-alert-circle',
      iconColor: 'text-orange-500',
      iconBg: 'bg-orange-100 dark:bg-orange-900/30',
      subMetrics: [
        {
          title: 'Open / Pending Action Plans',
          value: '4 open',
          target: '0 overdue',
          trend: '31 total recommendations registered'
        }
      ]
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
        <p :class="['text-sm font-semibold pt-2', card.trendUp ? 'text-success-600 dark:text-success-400' : 'text-rose-600 dark:text-rose-400']">
          {{ card.trend }}
        </p>
      </div>

      <!-- Tiered sub-metrics details -->
      <div v-if="card.subMetrics && card.subMetrics.length > 0" class="mt-4 pt-4 border-t border-gray-100 dark:border-gray-800 space-y-3">
        <div v-for="(sub, sIdx) in card.subMetrics" :key="sIdx" class="space-y-1 bg-gray-50 dark:bg-gray-800/40 p-2.5 rounded-lg border border-gray-100 dark:border-gray-800/80">
          <div class="flex justify-between items-center text-xs font-semibold text-gray-400 dark:text-gray-500 gap-2">
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
