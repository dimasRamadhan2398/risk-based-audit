<template>
  <div class="w-full min-w-0 p-4 sm:p-6 dashboard-card space-y-4">
    <!-- Header: Title, Count Badge, Reset Filter -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
      <div class="flex items-center gap-2">
        <h2 class="text-base font-bold text-slate-900 tracking-tight dark:text-white flex items-center gap-2">
          <UIcon name="i-lucide-target" class="size-5 text-primary-500" />
          {{ t('strategicPlan.title') }}
        </h2>
        <UBadge color="primary" variant="subtle" size="sm" class="font-semibold tabular-nums">
          {{ filteredObjectives.length }}
        </UBadge>
      </div>

      <!-- Reset Filters Button if filtered -->
      <UButton
        v-if="isFiltered"
        icon="i-lucide-rotate-ccw"
        color="neutral"
        variant="ghost"
        size="sm"
        class="text-xs font-semibold hover:bg-gray-100 dark:hover:bg-gray-800"
        @click="resetFilters"
      >
        {{ t('strategicPlan.filters.reset') }}
      </UButton>
    </div>

    <!-- Search & Filter Options Bar -->
    <div class="flex flex-wrap items-center gap-3 w-full">
      <!-- Search Input -->
      <div class="w-full sm:flex-1 sm:min-w-64">
        <CommonSearchInput
          v-model="searchQuery"
          :placeholder="t('strategicPlan.filters.searchPlaceholder')"
          class="w-full"
        />
      </div>

      <!-- Period Type Filter -->
      <div class="w-full sm:w-36">
        <USelectMenu
          v-model="selectedPeriodType"
          :items="periodTypeOptions"
          value-key="value"
          :placeholder="t('strategicPlan.filters.periodType')"
          class="w-full"
        />
      </div>

      <!-- Year Filter (shown strictly when Period Type is Yearly or ALL) -->
      <div
        v-if="selectedPeriodType === 'Yearly' || selectedPeriodType === 'ALL'"
        class="w-full sm:w-36"
      >
        <USelectMenu
          v-model="selectedYear"
          :items="yearFilterOptions"
          value-key="value"
          :placeholder="t('strategicPlan.filters.selectYear')"
          class="w-full"
        />
      </div>

      <!-- Quartal Filter (shown strictly when Period Type is Quartal) -->
      <div
        v-if="selectedPeriodType === 'Quartal'"
        class="w-full sm:w-36"
      >
        <USelectMenu
          v-model="selectedQuartal"
          :items="quartalFilterOptions"
          value-key="value"
          :placeholder="t('strategicPlan.filters.selectQuartal')"
          class="w-full"
        />
      </div>
    </div>

    <!-- Desktop View (>= md): Full Table Entities in UCard -->
    <div class="hidden md:block">
      <UCard class="overflow-hidden border border-gray-200 dark:border-gray-800" :ui="{ root: 'bg-transparent!', body: 'p-0 bg-transparent!' }">
        <TableEntities
          :data="filteredObjectives"
          :columns="columns"
          :loading="store.loading"
          :empty-state="{
            icon: 'i-lucide-target',
            label: t('strategicPlan.vmg.emptyTitle') || 'Tidak ada sasaran strategis',
            description: t('strategicPlan.vmg.emptySubtitle') || 'Belum ada data sasaran strategis yang ditemukan.'
          }"
          :ui="{ root: 'overflow-visible' }"
          class="w-full"
        >
          <!-- Strategic Objective Cell with Text Wrapping -->
          <template #strategicObjective-cell="{ row }">
            <div class="whitespace-normal break-words font-medium text-gray-900 dark:text-white py-1">
              <div v-if="row.original.code" class="mb-1">
                <span class="inline-flex items-center font-mono text-[10px] font-bold text-primary-600 dark:text-primary-400 bg-primary-50 dark:bg-primary-950/40 px-1.5 py-0.5 rounded border border-primary-200/50 dark:border-primary-800/40">
                  {{ row.original.code }}
                </span>
              </div>
              {{ row.original.strategicObjective || '-' }}
            </div>
          </template>

          <!-- KPI Name Cell with Text Wrapping -->
          <template #kpi-cell="{ row }">
            <div class="whitespace-normal break-words font-medium text-gray-900 dark:text-white py-1">
              {{ row.original.kpi || '-' }}
            </div>
          </template>

          <!-- Category -->
          <template #category-cell="{ row }">
            <UBadge color="neutral" variant="subtle" size="xs" class="font-medium">
              {{ categoryLabel(row.original.category) }}
            </UBadge>
          </template>

          <!-- Unit -->
          <template #unit-cell="{ row }">
            <span class="whitespace-nowrap text-xs text-gray-700 dark:text-gray-300 font-medium">{{ unitLabel(row.original.unit) }}</span>
          </template>

          <!-- Period -->
          <template #selectedPeriod-cell="{ row }">
            <span class="whitespace-nowrap font-mono text-xs text-slate-600 dark:text-slate-400 font-semibold">{{ row.original.selectedPeriod || row.original.periodType || '-' }}</span>
          </template>

          <!-- Target -->
          <template #target-cell="{ row }">
            <span class="whitespace-nowrap tabular-nums text-sm font-medium text-slate-700 dark:text-slate-300">{{ formatUnitValue(row.original.target, row.original.unit, locale) }}</span>
          </template>

          <!-- Actual -->
          <template #actual-cell="{ row }">
            <span class="whitespace-nowrap tabular-nums text-sm font-bold text-slate-900 dark:text-white">{{ formatUnitValue(row.original.actual, row.original.unit, locale) }}</span>
          </template>

          <!-- Calculation -->
          <template #calculation-cell="{ row }">
            <span class="whitespace-nowrap text-xs text-[var(--text-muted)]">{{ row.original.calculation || '-' }}</span>
          </template>

          <!-- Status with color coding -->
          <template #status-cell="{ row }">
            <UBadge
              :color="getStatusBadgeColor(row.original.status)"
              variant="subtle"
              size="xs"
              class="font-bold tabular-nums"
            >
              <UIcon :name="getStatusIcon(row.original.status)" class="size-3 mr-1" />
              {{ formatStatus(row.original.status) }}
            </UBadge>
          </template>

          <!-- Action Buttons -->
          <template #actions-cell="{ row }">
            <div class="flex items-center justify-center gap-1">
              <!-- Direct View Details Button -->
              <UTooltip :text="t('strategicPlan.actions.viewDetails')">
                <UButton
                  icon="i-lucide-eye"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  class="hover:bg-gray-100 dark:hover:bg-gray-800"
                  @click="store.openViewModal(row.original)"
                />
              </UTooltip>
              <UTooltip
                v-if="canWrite"
                :text="t('strategicPlan.actions.edit')"
              >
                <UButton
                  icon="i-lucide-edit"
                  color="warning"
                  variant="ghost"
                  size="xs"
                  class="hover:bg-warning-50 dark:hover:bg-warning-950/40"
                  @click="store.handleEdit(row.original)"
                />
              </UTooltip>
              <UTooltip
                v-if="canWrite"
                :text="t('strategicPlan.actions.delete')"
              >
                <UButton
                  icon="i-lucide-trash-2"
                  color="error"
                  variant="ghost"
                  size="xs"
                  class="hover:bg-error-50 dark:hover:bg-error-950/40"
                  @click="store.handleDelete(row.original.id)"
                />
              </UTooltip>
            </div>
          </template>
        </TableEntities>
      </UCard>
    </div>

    <!-- Mobile & Tablet View (< md): Stacked Cards List -->
    <div class="block md:hidden">
      <!-- Loading State Skeleton -->
      <div v-if="store.loading" class="space-y-3">
        <div v-for="i in 3" :key="i" class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] animate-pulse space-y-2.5">
          <div class="flex justify-between items-center">
            <div class="h-4 w-24 bg-gray-200 dark:bg-gray-700 rounded"></div>
            <div class="h-5 w-20 bg-gray-200 dark:bg-gray-700 rounded-full"></div>
          </div>
          <div class="h-4 w-3/4 bg-gray-200 dark:bg-gray-700 rounded"></div>
          <div class="h-10 bg-gray-100 dark:bg-gray-800 rounded-lg"></div>
        </div>
      </div>

      <!-- Empty State -->
      <div
        v-else-if="!filteredObjectives.length"
        class="py-8 text-center border border-dashed border-gray-200 dark:border-gray-800 rounded-xl bg-[var(--bg-surface)]"
      >
        <UIcon name="i-lucide-target" class="size-8 mx-auto text-gray-400 mb-2" />
        <p class="text-xs font-semibold text-[var(--text-main)]">{{ t('strategicPlan.vmg.emptyTitle') || 'Tidak ada sasaran strategis' }}</p>
        <p class="text-xs text-[var(--text-muted)] mt-1 max-w-sm mx-auto">{{ t('strategicPlan.vmg.emptySubtitle') || 'Belum ada data sasaran strategis yang ditemukan.' }}</p>
      </div>

      <!-- Cards List -->
      <div v-else class="space-y-3">
        <div
          v-for="item in filteredObjectives"
          :key="item.id"
          class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
        >
          <!-- Card Header: Code/Period Badge & Status Badge -->
          <div class="flex items-center justify-between gap-2">
            <div class="flex items-center gap-1.5 flex-wrap">
              <span v-if="item.code" class="inline-flex items-center font-mono text-[11px] font-bold text-primary-600 dark:text-primary-400 bg-primary-50 dark:bg-primary-950/40 px-2 py-0.5 rounded-md border border-primary-200/50 dark:border-primary-800/40 tabular-nums">
                {{ item.code }}
              </span>
              <span class="inline-flex items-center gap-1 text-[11px] font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded-md border border-slate-200/80 dark:border-slate-700">
                <UIcon name="i-lucide-calendar" class="size-3" />
                {{ item.selectedPeriod || item.periodType }}
              </span>
            </div>
            <UBadge :color="getStatusBadgeColor(item.status)" variant="subtle" size="xs" class="font-bold">
              <UIcon :name="getStatusIcon(item.status)" class="size-3 mr-1" />
              {{ formatStatus(item.status) }}
            </UBadge>
          </div>

          <!-- Card Body: Strategic Objective & KPI -->
          <div>
            <h4 class="text-sm font-bold text-gray-900 dark:text-white leading-snug line-clamp-2">
              {{ item.strategicObjective || '-' }}
            </h4>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-1 line-clamp-2">
              <span class="font-semibold text-slate-700 dark:text-slate-300">KPI:</span> {{ item.kpi || '-' }}
            </p>
          </div>

          <!-- Target vs Actual Metric Grid -->
          <div class="grid grid-cols-2 gap-2 p-2.5 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800 text-xs">
            <div>
              <span class="text-[10px] uppercase tracking-wider font-semibold text-slate-400 dark:text-slate-400 block">{{ t('strategicPlan.columns.target') }}</span>
              <span class="font-bold text-slate-800 dark:text-slate-200 tabular-nums text-sm">
                {{ formatUnitValue(item.target, item.unit, locale) }}
              </span>
              <span class="text-[10px] text-slate-400 ml-1">({{ unitLabel(item.unit) }})</span>
            </div>
            <div>
              <span class="text-[10px] uppercase tracking-wider font-semibold text-slate-400 dark:text-slate-400 block">{{ t('strategicPlan.columns.actual') }}</span>
              <span class="font-bold text-slate-900 dark:text-white tabular-nums text-sm">
                {{ formatUnitValue(item.actual, item.unit, locale) }}
              </span>
              <span class="text-[10px] text-slate-400 ml-1">({{ unitLabel(item.unit) }})</span>
            </div>
          </div>

          <!-- Metadata: Category, Calculation, and Actions -->
          <div class="flex flex-wrap items-center justify-between gap-2 pt-2 border-t border-gray-100 dark:border-gray-800/80 text-xs text-[var(--text-muted)]">
            <div class="flex items-center gap-1.5">
              <UBadge color="neutral" variant="subtle" size="xs" class="font-medium">
                <UIcon name="i-lucide-tag" class="size-3 mr-1" />
                {{ categoryLabel(item.category) }}
              </UBadge>
              <span v-if="item.calculation" class="text-[11px] text-slate-400">
                • {{ item.calculation }}
              </span>
            </div>

            <!-- Card Actions -->
            <div class="flex items-center gap-1 ml-auto">
              <UButton
                icon="i-lucide-eye"
                color="neutral"
                variant="outline"
                size="xs"
                :label="t('strategicPlan.actions.viewDetails')"
                class="font-semibold"
                @click="store.openViewModal(item)"
              />
              <UButton
                v-if="canWrite"
                icon="i-lucide-edit"
                color="warning"
                variant="ghost"
                size="xs"
                @click="store.handleEdit(item)"
              />
              <UButton
                v-if="canWrite"
                icon="i-lucide-trash-2"
                color="error"
                variant="ghost"
                size="xs"
                @click="store.handleDelete(item.id)"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import type { StrategicAuditPlan } from '~/types/audit'
import { formatUnitValue } from '~/utils/numberInput'
import { kpiValueLabel } from '~/utils/kpiPerformanceLabels'
import TableEntities from '~/components/shared/TableEntities.vue'

const { t, locale } = useI18n()
const store = useStrategicPlanStore()
const { canManageStrategicPlan, canWrite } = useRbac()

const searchQuery = ref('')
const selectedPeriodType = ref('ALL')
const selectedYear = ref('ALL')
const selectedQuartal = ref('ALL')

watch(selectedPeriodType, (newType) => {
  if (newType === 'Yearly') {
    selectedQuartal.value = 'ALL'
  } else if (newType === 'Quartal') {
    selectedYear.value = 'ALL'
  }
})

const columns = computed(() => [
  { key: 'strategicObjective', accessorKey: 'strategicObjective', label: t('strategicPlan.columns.objective'), header: t('strategicPlan.columns.objective'), class: 'w-[130px] min-w-[130px] max-w-[130px] sm:w-[220px] sm:min-w-[220px] sm:max-w-[220px] whitespace-normal break-words font-medium sticky left-0', thClass: 'z-20 !bg-[var(--bg-surface)]', tdClass: 'z-10 bg-[var(--bg-main)]' },
  { key: 'kpi', accessorKey: 'kpi', label: t('strategicPlan.columns.kpi'), header: t('strategicPlan.columns.kpi'), class: 'w-[130px] min-w-[130px] max-w-[130px] sm:w-[220px] sm:min-w-[220px] sm:max-w-[220px] whitespace-normal break-words font-medium sticky left-[130px] sm:left-[220px] border-r border-[var(--border-main)] shadow-[2px_0_4px_-2px_rgba(0,0,0,0.15)]', thClass: 'z-20 !bg-[var(--bg-surface)]', tdClass: 'z-10 bg-[var(--bg-main)]' },
  { key: 'category', accessorKey: 'category', label: t('strategicPlan.columns.category'), header: t('strategicPlan.columns.category'), class: 'w-28 whitespace-nowrap' },
  { key: 'unit', accessorKey: 'unit', label: t('strategicPlan.columns.unit'), header: t('strategicPlan.columns.unit'), class: 'w-20 text-center whitespace-nowrap' },
  { key: 'selectedPeriod', accessorKey: 'selectedPeriod', label: t('strategicPlan.columns.period'), header: t('strategicPlan.columns.period'), class: 'w-28 whitespace-nowrap' },
  { key: 'target', accessorKey: 'target', label: t('strategicPlan.columns.target'), header: t('strategicPlan.columns.target'), class: 'w-24 whitespace-nowrap' },
  { key: 'actual', accessorKey: 'actual', label: t('strategicPlan.columns.actual'), header: t('strategicPlan.columns.actual'), class: 'w-24 whitespace-nowrap' },
  { key: 'calculation', accessorKey: 'calculation', label: t('strategicPlan.columns.calculation'), header: t('strategicPlan.columns.calculation'), class: 'w-28 whitespace-nowrap' },
  { key: 'status', accessorKey: 'status', label: t('strategicPlan.columns.status'), header: t('strategicPlan.columns.status'), class: 'w-28 whitespace-nowrap' },
  { key: 'actions', accessorKey: 'actions', label: t('strategicPlan.columns.actions'), header: t('strategicPlan.columns.actions'), class: 'w-24 whitespace-nowrap text-center' }
])

const periodTypeOptions = computed(() => [
  { label: t('strategicPlan.filters.allTypes'), value: 'ALL' },
  { label: t('strategicPlan.filters.yearly'), value: 'Yearly' },
  { label: t('strategicPlan.filters.quartal'), value: 'Quartal' }
])

const yearFilterOptions = computed(() => [
  { label: t('strategicPlan.filters.allYears'), value: 'ALL' },
  ...getFiscalYears().map(y => ({ label: String(y), value: String(y) }))
])

const quartalFilterOptions = computed(() => [
  { label: t('strategicPlan.filters.allQuarters'), value: 'ALL' },
  { label: 'Q1', value: 'Q1' },
  { label: 'Q2', value: 'Q2' },
  { label: 'Q3', value: 'Q3' },
  { label: 'Q4', value: 'Q4' }
])

const isFiltered = computed(() => {
  return searchQuery.value !== ''
    || selectedPeriodType.value !== 'ALL'
    || selectedYear.value !== 'ALL'
    || selectedQuartal.value !== 'ALL'
})

const resetFilters = () => {
  searchQuery.value = ''
  selectedPeriodType.value = 'ALL'
  selectedYear.value = 'ALL'
  selectedQuartal.value = 'ALL'
}

const unitLabel = (unit?: string) => {
  if (!unit) return '-'
  return unit === 'Amount' ? t('strategicPlan.form.units.quantity') : unit
}

const categoryLabel = (value?: string) => kpiValueLabel(t, 'categories', value) || '-'

const formatStatus = (status: string) => {
  if (!status) return '-'
  const lower = status.toLowerCase()
  if (['good', 'moderate', 'poor', 'pending', 'planned'].includes(lower)) {
    return t(`strategicPlan.status.${lower}`)
  }
  return status
}

const getStatusBadgeColor = (status?: string) => {
  if (!status) return 'neutral'
  const lower = status.toLowerCase()
  if (lower === 'good') return 'success'
  if (lower === 'moderate') return 'warning'
  if (lower === 'poor') return 'error'
  return 'neutral'
}

const getStatusIcon = (status?: string) => {
  if (!status) return 'i-lucide-circle'
  const lower = status.toLowerCase()
  if (lower === 'good') return 'i-lucide-check-circle'
  if (lower === 'moderate') return 'i-lucide-alert-triangle'
  if (lower === 'poor') return 'i-lucide-alert-circle'
  return 'i-lucide-circle'
}

const filteredObjectives = computed(() => {
  let list = store.strategicObjectives

  // 1. Period Type Filter (Yearly vs Quartal)
  if (selectedPeriodType.value !== 'ALL') {
    list = list.filter(item => item.periodType === selectedPeriodType.value)
  }

  // 2. Year Filter
  if (selectedYear.value !== 'ALL') {
    const yStr = selectedYear.value
    list = list.filter((item) => {
      if (item.kpiTargets && item.kpiTargets[yStr] !== undefined) return true
      if (item.selectedPeriod === yStr) return true
      if (item.yearStart && item.yearEnd) {
        const yNum = parseInt(yStr)
        return yNum >= item.yearStart && yNum <= item.yearEnd
      }
      return false
    })
  }

  // 3. Quartal Filter (Q1, Q2, Q3, Q4)
  if (selectedQuartal.value !== 'ALL') {
    list = list.filter(item => item.selectedPeriod === selectedQuartal.value)
  }

  // 4. Search Query
  if (searchQuery.value.trim() !== '') {
    const q = searchQuery.value.toLowerCase()
    list = list.filter((item) => {
      return (item.strategicObjective && item.strategicObjective.toLowerCase().includes(q))
        || (item.kpi && item.kpi.toLowerCase().includes(q))
    })
  }

  // Recalculate target/actual if specific year selected
  if (selectedYear.value !== 'ALL') {
    const yr = selectedYear.value
    return list.map((item) => {
      const yearTarget = item.kpiTargets?.[yr] || item.target
      const yearActual = item.kpiActuals?.[yr] || item.actual
      return {
        ...item,
        target: yearTarget,
        actual: yearActual,
        selectedPeriod: `${yr} (${item.selectedPeriod || item.periodType})`
      }
    })
  }

  return list
})

const getActions = (item: StrategicAuditPlan) => {
  const actions: any[] = [
    {
      type: 'label' as const,
      label: t('strategicPlan.actions.title')
    },
    {
      label: t('strategicPlan.actions.viewDetails'),
      icon: 'i-lucide-eye',
      onSelect: () => store.openViewModal(item)
    }
  ]

  if (canManageStrategicPlan.value) {
    actions.push(
      {
        label: t('strategicPlan.actions.edit'),
        icon: 'i-lucide-edit',
        onSelect: () => store.handleEdit(item)
      },
      {
        label: t('strategicPlan.actions.delete'),
        icon: 'i-lucide-trash-2',
        onSelect: () => store.handleDelete(item.id)
      }
    )
  }

  return [actions]
}
</script>
