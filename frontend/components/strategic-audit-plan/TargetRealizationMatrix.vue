<template>
  <div class="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl border border-gray-200 dark:border-gray-700 space-y-3">
    <!-- Header -->
    <div class="flex items-center justify-between flex-wrap gap-2">
      <label class="font-bold text-sm text-gray-900 dark:text-white flex items-center gap-2">
        <UIcon name="i-lucide-target" class="w-4 h-4 text-primary-500" />
        {{ t('strategicPlan.form.targetMatrixTitle', { start: startYear, end: endYear }) }}
        <span class="text-orange-500">*</span>
      </label>
      <span class="text-xs text-gray-500 dark:text-gray-400 font-medium">
        {{ t('strategicPlan.form.targetMatrixSubtitle', { duration: yearsList.length }) }}
      </span>
    </div>

    <!-- Flexible Year Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
      <div
        v-for="(yr, idx) in yearsList"
        :key="yr"
        class="space-y-2 p-3 rounded-xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 text-center shadow-xs transition-all hover:border-primary-400"
      >
        <div class="pb-1.5 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
          <span class="text-[11px] font-extrabold uppercase tracking-wide text-primary-600 dark:text-primary-400">
            {{ t('strategicPlan.form.yearLabel', { number: idx + 1 }) }}
          </span>
          <span class="text-[11px] font-mono font-bold text-gray-700 dark:text-gray-300">
            {{ yr }}
          </span>
        </div>

        <!-- Target Input -->
        <div class="space-y-1 text-left">
          <label class="text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase">Target</label>
          <UInput
            :model-value="(store.form.kpiTargets as any)?.[yr] || ''"
            placeholder="Target"
            size="xs"
            class="w-full text-center font-bold font-mono text-gray-900 dark:text-white"
            @update:model-value="(val: string) => updateYearTarget(yr, val)"
          />
        </div>

        <!-- Realisasi Input -->
        <div class="space-y-1 text-left">
          <label class="text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase">Realisasi</label>
          <UInput
            :model-value="(store.form.kpiActuals as any)?.[yr] || ''"
            placeholder="Actual"
            size="xs"
            class="w-full text-center font-bold font-mono text-gray-900 dark:text-white"
            @update:model-value="(val: string) => updateYearActual(yr, val)"
          />
        </div>

        <!-- Hitungan (Calculation / % Achievement) -->
        <div class="space-y-1 text-left">
          <label class="text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase">Hitungan</label>
          <div
            class="w-full text-center font-bold font-mono text-xs py-1 px-2 rounded border transition-colors"
            :class="getYearCalcColorClass(yr)"
          >
            {{ getYearCalculation(yr) }}
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

const yearsList = computed(() => {
  const start = startYear.value
  const end = endYear.value
  const years: string[] = []
  for (let y = start; y <= end; y++) {
    years.push(String(y))
  }
  return years
})

function updateYearTarget(year: string, value: string) {
  if (!store.form.kpiTargets) {
    store.form.kpiTargets = {}
  }
  ;(store.form.kpiTargets as any)[year] = value
  if (store.form.selectedPeriod === year || !store.form.target) {
    store.form.target = value
  }
}

function updateYearActual(year: string, value: string) {
  if (!store.form.kpiActuals) {
    store.form.kpiActuals = {}
  }
  ;(store.form.kpiActuals as any)[year] = value
  if (store.form.selectedPeriod === year || !store.form.actual) {
    store.form.actual = value
  }
}

function getYearCalculation(year: string): string {
  const targetStr = (store.form.kpiTargets as any)?.[year]
  const actualStr = (store.form.kpiActuals as any)?.[year]
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

function getYearCalcColorClass(year: string): string {
  const calc = getYearCalculation(year)
  if (calc === '-') return 'text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 border-gray-200 dark:border-gray-700'
  const val = parseFloat(calc)
  if (val >= 100) return 'text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/40 border-emerald-200 dark:border-emerald-800'
  if (val >= 70) return 'text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800'
  return 'text-rose-600 dark:text-rose-400 bg-rose-50 dark:bg-rose-950/40 border-rose-200 dark:border-rose-800'
}
</script>
