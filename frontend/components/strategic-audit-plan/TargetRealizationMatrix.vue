<template>
  <div class="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl border border-gray-200 dark:border-gray-700 space-y-3">
    <!-- Header -->
    <div class="flex items-center justify-between flex-wrap gap-2">
      <label class="font-bold text-sm text-gray-900 dark:text-white flex items-center gap-2">
        <UIcon name="i-lucide-target" class="w-4 h-4 text-primary-500" />
        <template v-if="store.form.periodType === 'Quartal'">
          {{ t('strategicPlan.form.targetMatrixQuarterlyTitle') }}
        </template>
        <template v-else>
          {{ t('strategicPlan.form.targetMatrixTitle', { start: startYear, end: endYear }) }}
        </template>
        <span class="text-orange-500">*</span>
      </label>
      <span class="text-xs text-gray-500 dark:text-gray-400 font-medium">
        <template v-if="store.form.periodType === 'Quartal'">
          {{ t('strategicPlan.form.targetMatrixQuarterlySubtitle') }}
        </template>
        <template v-else>
          {{ t('strategicPlan.form.targetMatrixSubtitle', { duration: periodsList.length }) }}
        </template>
      </span>
    </div>

    <!-- Flexible Grid (Quarterly or Yearly) -->
    <div 
      class="grid gap-3"
      :class="store.form.periodType === 'Quartal' ? 'grid-cols-1 sm:grid-cols-2 md:grid-cols-4' : 'grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5'"
    >
      <div
        v-for="(item, idx) in periodsList"
        :key="item"
        class="space-y-2 p-3 rounded-xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 text-center shadow-md transition-all hover:border-primary-400"
      >
        <div class="pb-1.5 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
          <span class="text-[11px] font-extrabold uppercase tracking-wide text-primary-600 dark:text-primary-400">
            {{ store.form.periodType === 'Quartal' ? item.split('-')[0] : t('strategicPlan.form.yearLabel', { number: idx + 1 }) }}
          </span>
          <span class="text-[11px] font-mono font-bold text-gray-700 dark:text-gray-300">
            {{ store.form.periodType === 'Quartal' ? item.split('-')[1] : item }}
          </span>
        </div>

        <!-- Target Input -->
        <div class="space-y-1 text-left">
          <label class="text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase">Target</label>
          <UInput
            :model-value="(store.form.kpiTargets as any)?.[item] || ''"
            placeholder="Target"
            size="md"
            class="w-full text-center font-bold font-mono text-gray-900 dark:text-white"
            @update:model-value="(val: string) => updatePeriodTarget(item, val)"
          />
        </div>

        <!-- Realisasi Input -->
        <div class="space-y-1 text-left">
          <label class="text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase">Realisasi</label>
          <UInput
            :model-value="(store.form.kpiActuals as any)?.[item] || ''"
            placeholder="Actual"
            size="md"
            class="w-full text-center font-bold font-mono text-gray-900 dark:text-white"
            @update:model-value="(val: string) => updatePeriodActual(item, val)"
          />
        </div>

        <!-- Hitungan (Calculation / % Achievement) -->
        <div class="space-y-1 text-left">
          <label class="text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase">Hitungan</label>
          <div
            class="w-full text-center font-bold font-mono text-md py-1 px-2 rounded border transition-colors"
            :class="getPeriodCalcColorClass(item)"
          >
            {{ getPeriodCalculation(item) }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const store = useStrategicPlanStore()

const currentYear = new Date().getFullYear()

const startYear = computed(() => store.form.yearStart || currentYear)
const endYear = computed(() => {
  const start = startYear.value
  const end = store.form.yearEnd
  if (end && end >= start) return end
  return start + 4
})

const periodsList = computed(() => {
  const start = startYear.value
  const end = endYear.value
  if (store.form.periodType === 'Quartal') {
    // Generate Q1–Q4 for every year in the range
    // Key format: "Q1-2026", "Q2-2026", ..., "Q4-2030"
    const quarters: string[] = []
    for (let y = start; y <= end; y++) {
      quarters.push(`Q1-${y}`, `Q2-${y}`, `Q3-${y}`, `Q4-${y}`)
    }
    return quarters
  }
  const years: string[] = []
  for (let y = start; y <= end; y++) {
    years.push(String(y))
  }
  return years
})

function updatePeriodTarget(periodKey: string, value: string) {
  if (!store.form.kpiTargets) {
    store.form.kpiTargets = {}
  }
  ;(store.form.kpiTargets as any)[periodKey] = value
  if (store.form.selectedPeriod === periodKey || !store.form.target) {
    store.form.target = value
  }
}

function updatePeriodActual(periodKey: string, value: string) {
  if (!store.form.kpiActuals) {
    store.form.kpiActuals = {}
  }
  ;(store.form.kpiActuals as any)[periodKey] = value
  if (store.form.selectedPeriod === periodKey || !store.form.actual) {
    store.form.actual = value
  }
}

function getPeriodCalculation(periodKey: string): string {
  const targetStr = (store.form.kpiTargets as any)?.[periodKey]
  const actualStr = (store.form.kpiActuals as any)?.[periodKey]
  if (!targetStr || !actualStr) return '-'
  const target = parseFloat(targetStr)
  const actual = parseFloat(actualStr)
  if (isNaN(target) || isNaN(actual)) return '-'

  let result = 0
  if (store.form.hibHig === 'HIG') {
    if (target === 0) return '-'
    result = (actual / target) * 100
  } else {
    if (actual === 0) return '-'
    result = (target / actual) * 100
  }
  return `${result.toFixed(2)}%`
}

function getPeriodCalcColorClass(periodKey: string): string {
  const calc = getPeriodCalculation(periodKey)
  if (calc === '-') return 'text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 border-gray-200 dark:border-gray-700'
  const val = parseFloat(calc)
  if (val >= 100) return 'text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/40 border-emerald-200 dark:border-emerald-800'
  if (val >= 70) return 'text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800'
  return 'text-rose-600 dark:text-rose-400 bg-rose-50 dark:bg-rose-950/40 border-rose-200 dark:border-rose-800'
}
</script>
