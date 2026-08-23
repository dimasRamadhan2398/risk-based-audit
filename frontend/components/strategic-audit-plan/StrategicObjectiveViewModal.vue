<template>
  <UModal
    v-model:open="store.isViewModalOpen"
    :title="t('strategicPlan.viewModal.title')"
    :ui="{
      content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-2xl overflow-hidden',
      header: 'border-b border-gray-100 dark:border-gray-800 p-5 font-bold text-gray-900 dark:text-white shrink-0',
      body: 'p-6 space-y-6 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
      footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0 bg-white dark:bg-gray-900',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
  >
    <template #body>
      <div v-if="objective" class="space-y-6">
        <!-- Top Metadata Summary Card -->
        <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700/60 flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div class="space-y-1">
            <div class="flex items-center gap-2">
              <span class="px-2.5 py-1 text-md font-mono font-bold rounded-lg bg-primary-100 dark:bg-primary-900/40 text-primary-700 dark:text-primary-300">
                {{ objective.code || 'SO-IA' }}
              </span>
              <span class="px-2.5 py-1 text-md font-semibold rounded-lg bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200">
                {{ objective.periodType || 'Yearly' }}
              </span>
              <span 
                :class="[
                  'px-2.5 py-1 text-md font-semibold rounded-lg',
                  objective.hibHig === 'HIG' ? 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-300' : 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-300'
                ]"
              >
                {{ objective.hibHig === 'HIG' ? t('strategicPlan.form.higLabel') : t('strategicPlan.form.hibLabel') }}
              </span>
            </div>
            <h2 class="text-lg font-bold text-gray-900 dark:text-white">
              {{ objective.strategicObjective }}
            </h2>
            <p v-if="objective.internalAuditSO" class="text-md text-gray-500 dark:text-gray-400">
              <span class="font-semibold text-gray-700 dark:text-gray-300">{{ t('strategicPlan.viewModal.iaObjective') }}</span> {{ objective.internalAuditSO }}
            </p>
          </div>

          <div class="flex items-center gap-3 shrink-0">
            <div class="text-right">
              <p class="text-md text-gray-500 dark:text-gray-400 uppercase font-medium">{{ t('strategicPlan.viewModal.overallStatus') }}</p>
              <span
                :class="{
                  'text-emerald-600 dark:text-emerald-400 font-bold text-sm': objective.status === 'Good',
                  'text-amber-600 dark:text-amber-400 font-bold text-sm': objective.status === 'Moderate',
                  'text-rose-600 dark:text-rose-400 font-bold text-sm': objective.status === 'Poor',
                }"
              >
                {{ formatStatus(objective.status) }}
              </span>
            </div>
            <div class="w-12 h-12 rounded-xl bg-primary-50 dark:bg-primary-950/40 flex items-center justify-center text-primary-600 dark:text-primary-400 font-bold text-sm">
              {{ objective.calculation || '-' }}
            </div>
          </div>
        </div>

        <!-- KPI Key Details Grid -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="p-3.5 rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
            <p class="text-md font-semibold text-gray-500 dark:text-gray-400 uppercase">{{ t('strategicPlan.viewModal.kpiName') }}</p>
            <p class="text-sm font-medium text-gray-900 dark:text-white mt-1">{{ objective.kpi || '-' }}</p>
          </div>
          <div class="p-3.5 rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
            <p class="text-md font-semibold text-gray-500 dark:text-gray-400 uppercase">{{ t('strategicPlan.viewModal.unitAndTargetPeriod') }}</p>
            <p class="text-sm font-medium text-gray-900 dark:text-white mt-1">
              {{ objective.unit || '-' }} <span class="text-gray-400 dark:text-gray-500">({{ objective.selectedPeriod || 'All' }})</span>
            </p>
          </div>
          <div class="p-3.5 rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
            <p class="text-md font-semibold text-gray-500 dark:text-gray-400 uppercase">{{ t('strategicPlan.viewModal.strategicPeriod') }}</p>
            <p class="text-sm font-medium text-gray-900 dark:text-white mt-1">
              {{ objective.yearStart || 2024 }} &ndash; {{ objective.yearEnd || 2028 }}
            </p>
          </div>
        </div>

        <!-- Multi-Year Target & Actual Matrix Table -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon name="i-lucide-table" class="w-4 h-4 text-primary-500" />
              {{ t('strategicPlan.viewModal.multiYearMatrixTitle') }}
            </h3>
            <span class="text-md text-gray-500 dark:text-gray-400">
              {{ t('strategicPlan.viewModal.unitLabel', { unit: objective.unit || '%' }) }}
            </span>
          </div>

          <div class="overflow-x-auto rounded-xl border border-gray-200 dark:border-gray-800">
            <table class="w-full text-left text-md">
              <thead class="bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 font-semibold border-b border-gray-200 dark:border-gray-700">
                <tr>
                  <th class="py-2.5 px-4">{{ t('strategicPlan.viewModal.periodYear') }}</th>
                  <th class="py-2.5 px-4 text-right">{{ t('strategicPlan.viewModal.target') }}</th>
                  <th class="py-2.5 px-4 text-right">{{ t('strategicPlan.viewModal.actual') }}</th>
                  <th class="py-2.5 px-4 text-right">{{ t('strategicPlan.viewModal.achievement') }}</th>
                  <th class="py-2.5 px-4 text-center">{{ t('strategicPlan.viewModal.status') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                <tr 
                  v-for="row in multiYearData" 
                  :key="row.year"
                  class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors"
                >
                  <td class="py-2.5 px-4 font-bold text-gray-900 dark:text-white">
                    {{ row.year }}
                    <span v-if="String(row.year) === String(currentYear)" class="ml-1.5 px-1.5 py-0.5 text-[10px] bg-primary-100 text-primary-700 dark:bg-primary-900/50 dark:text-primary-300 rounded font-semibold">
                      {{ t('strategicPlan.viewModal.current') }}
                    </span>
                  </td>
                  <td class="py-2.5 px-4 text-right font-medium text-gray-800 dark:text-gray-200">
                    {{ row.target !== '' ? row.target + ' ' + objective.unit : '-' }}
                  </td>
                  <td class="py-2.5 px-4 text-right font-medium text-gray-800 dark:text-gray-200">
                    {{ row.actual !== '' ? row.actual + ' ' + objective.unit : '-' }}
                  </td>
                  <td class="py-2.5 px-4 text-right font-mono font-bold text-gray-900 dark:text-white">
                    {{ row.achievement }}
                  </td>
                  <td class="py-2.5 px-4 text-center">
                    <span
                      :class="[
                        'inline-flex items-center px-2 py-0.5 rounded-md text-[11px] font-semibold',
                        row.status === 'Good' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300' :
                        row.status === 'Moderate' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950/60 dark:text-amber-300' :
                        row.status === 'Poor' ? 'bg-rose-100 text-rose-800 dark:bg-rose-950/60 dark:text-rose-300' :
                        'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
                      ]"
                    >
                      {{ formatStatus(row.status) }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Quarterly Breakdown Table (if Quartal) -->
        <div v-if="objective.periodType === 'Quartal'" class="space-y-4 pt-2 border-t border-gray-100 dark:border-gray-800">
          <div class="flex items-center justify-between flex-wrap gap-2">
            <h3 class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon name="i-lucide-calendar-days" class="w-4 h-4 text-primary-500" />
              {{ t('strategicPlan.viewModal.quarterlyBreakdownTitle') }}
            </h3>
            <span class="text-xs font-semibold text-gray-500 dark:text-gray-400">
              {{ quarterlyYears.length * 4 }} Quartals ({{ objective.yearStart }} - {{ objective.yearEnd }})
            </span>
          </div>

          <div v-for="yr in quarterlyYears" :key="yr" class="space-y-2">
            <div class="flex items-center gap-2">
              <span class="text-xs font-extrabold uppercase tracking-wider text-gray-600 dark:text-gray-400">
                Tahun {{ yr }}
              </span>
              <span v-if="String(yr) === String(currentYear)" class="px-1.5 py-0.5 text-[10px] bg-primary-100 text-primary-700 dark:bg-primary-900/50 dark:text-primary-300 rounded font-semibold">
                {{ t('strategicPlan.viewModal.current') }}
              </span>
            </div>
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
              <div 
                v-for="q in ['Q1', 'Q2', 'Q3', 'Q4']" 
                :key="`${q}-${yr}`"
                :class="[
                  'p-3 rounded-xl border transition-all',
                  isQuarterSelected(q, yr)
                    ? 'border-primary-500 bg-primary-50/50 dark:bg-primary-950/20 ring-1 ring-primary-500' 
                    : 'border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900'
                ]"
              >
                <div class="flex items-center justify-between">
                  <span class="font-bold text-md text-gray-900 dark:text-white">{{ q }}</span>
                  <span v-if="isQuarterSelected(q, yr)" class="text-[10px] font-bold text-primary-600 dark:text-primary-400">
                    {{ t('strategicPlan.viewModal.selected') }}
                  </span>
                </div>
                <div class="mt-2 space-y-1 text-md">
                  <div class="flex justify-between text-gray-500 dark:text-gray-400">
                    <span>{{ t('strategicPlan.viewModal.target') }}:</span>
                    <span class="font-semibold text-gray-800 dark:text-gray-200">
                      {{ getQuarterTarget(q, yr) !== '-' ? getQuarterTarget(q, yr) + ' ' + objective.unit : '-' }}
                    </span>
                  </div>
                  <div class="flex justify-between text-gray-500 dark:text-gray-400">
                    <span>{{ t('strategicPlan.viewModal.actual') }}:</span>
                    <span class="font-semibold text-gray-800 dark:text-gray-200">
                      {{ getQuarterActual(q, yr) !== '-' ? getQuarterActual(q, yr) + ' ' + objective.unit : '-' }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

      </div>
    </template>

    <template #footer>
      <div class="flex items-center justify-between w-full">
        <p class="text-md text-gray-500 dark:text-gray-400">{{ t('strategicPlan.viewModal.footerNote') }}</p>
        <UButton color="neutral" variant="ghost" class="rounded-xl font-bold" @click="store.closeViewModal">
          {{ t('strategicPlan.viewModal.close') }}
        </UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const store = useStrategicPlanStore()
const currentYear = new Date().getFullYear()

const objective = computed(() => store.selectedViewObjective)

const formatStatus = (status: string) => {
  if (!status) return '-'
  const lower = status.toLowerCase()
  if (['good', 'moderate', 'poor', 'pending', 'planned'].includes(lower)) {
    return t(`strategicPlan.status.${lower}`)
  }
  return status
}

const quarterlyYears = computed(() => {
  if (!objective.value) return []
  const startY = objective.value.yearStart || currentYear
  const endY = objective.value.yearEnd || (startY + 4)
  const years = []
  for (let yr = startY; yr <= endY; yr++) {
    years.push(yr)
  }
  return years
})

const getQuarterTarget = (q: string, yr: number) => {
  if (!objective.value) return '-'
  const key = `${q}-${yr}`
  const targets = objective.value.kpiTargets || {}
  if (targets[key] !== undefined && targets[key] !== '') return String(targets[key])
  if (targets[q] !== undefined && targets[q] !== '') return String(targets[q])
  return '-'
}

const getQuarterActual = (q: string, yr: number) => {
  if (!objective.value) return '-'
  const key = `${q}-${yr}`
  const actuals = objective.value.kpiActuals || {}
  if (actuals[key] !== undefined && actuals[key] !== '') return String(actuals[key])
  if (actuals[q] !== undefined && actuals[q] !== '') return String(actuals[q])
  return '-'
}

const isQuarterSelected = (q: string, yr: number) => {
  if (!objective.value) return false
  const key = `${q}-${yr}`
  return objective.value.selectedPeriod === key || objective.value.selectedPeriod === q
}

const multiYearData = computed(() => {
  if (!objective.value) return []
  const startY = objective.value.yearStart || currentYear
  const endY = objective.value.yearEnd || (startY + 4)
  const targets = objective.value.kpiTargets || {}
  const actuals = objective.value.kpiActuals || {}

  const rows = []
  for (let yr = startY; yr <= endY; yr++) {
    let targetVal = targets[yr] !== undefined && targets[yr] !== '' ? String(targets[yr]) : ''
    let actualVal = actuals[yr] !== undefined && actuals[yr] !== '' ? String(actuals[yr]) : ''

    if (objective.value.periodType === 'Quartal' && (!targetVal || !actualVal)) {
      const qKeys = [`Q1-${yr}`, `Q2-${yr}`, `Q3-${yr}`, `Q4-${yr}`]
      if (!targetVal) {
        for (const qk of qKeys) {
          if (targets[qk] !== undefined && targets[qk] !== '') {
            targetVal = String(targets[qk])
            break
          }
        }
      }
      if (!actualVal) {
        for (let i = qKeys.length - 1; i >= 0; i--) {
          const qk = qKeys[i]
          if (actuals[qk] !== undefined && actuals[qk] !== '') {
            actualVal = String(actuals[qk])
            break
          }
        }
      }
    }

    if (!targetVal && yr === currentYear) {
      targetVal = objective.value.target || ''
    }
    if (!actualVal && yr === currentYear) {
      actualVal = objective.value.actual || ''
    }

    const targetNum = parseFloat(targetVal)
    const actualNum = parseFloat(actualVal)

    let achievement = '-'
    let status = 'Pending'

    if (!isNaN(targetNum) && !isNaN(actualNum) && targetNum > 0) {
      let ratio = 0
      if (objective.value.hibHig === 'HIG') {
        ratio = (actualNum / targetNum) * 100
      } else {
        ratio = (targetNum / actualNum) * 100
      }
      achievement = ratio.toFixed(2) + '%'
      if (ratio >= 100) status = 'Good'
      else if (ratio >= 70) status = 'Moderate'
      else status = 'Poor'
    } else if (targetVal !== '' && (actualVal === '' || actualVal === '-')) {
      status = 'Planned'
    }

    rows.push({
      year: yr,
      target: targetVal,
      actual: actualVal,
      achievement,
      status
    })
  }
  return rows
})
</script>
