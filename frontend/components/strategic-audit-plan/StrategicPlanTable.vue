<template>
  <div class="space-y-4">
    <!-- Goal Selection Tabs -->
    <div v-if="vmgStore.activeVmg?.goals?.length" class="flex gap-2 border-b border-[var(--border-main)] pb-2 overflow-x-auto">
      <UButton
        v-for="tab in goalTabs"
        :key="tab.value"
        :label="tab.label"
        :variant="selectedGoalId === tab.value ? 'solid' : 'ghost'"
        :color="selectedGoalId === tab.value ? 'primary' : 'neutral'"
        size="sm"
        @click="() => { selectedGoalId = tab.value }"
      />
    </div>

    <!-- Advanced Filter Controls Bar -->
    <UCard variant="soft" class="p-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <!-- Search & Filter Options -->
        <div class="flex flex-wrap items-center gap-3 flex-1">
          <!-- Search Input -->
          <div class="w-full sm:w-64">
            <UInput
              v-model="searchQuery"
              icon="i-lucide-search"
              :placeholder="t('strategicPlan.filters.searchPlaceholder')"
              class="w-full"
            />
          </div>

          <!-- Period Type Filter -->
          <div class="w-36">
            <USelectMenu
              v-model="selectedPeriodType"
              :items="periodTypeOptions"
              value-key="value"
              :placeholder="t('strategicPlan.filters.periodType')"
              class="w-full"
            />
          </div>

          <!-- Year Filter (shown strictly when Period Type is Yearly or ALL) -->
          <div v-if="selectedPeriodType === 'Yearly' || selectedPeriodType === 'ALL'" class="w-36">
            <USelectMenu
              v-model="selectedYear"
              :items="yearFilterOptions"
              value-key="value"
              :placeholder="t('strategicPlan.filters.selectYear')"
              class="w-full"
            />
          </div>

          <!-- Quartal Filter (shown strictly when Period Type is Quartal) -->
          <div v-if="selectedPeriodType === 'Quartal'" class="w-36">
            <USelectMenu
              v-model="selectedQuartal"
              :items="quartalFilterOptions"
              value-key="value"
              :placeholder="t('strategicPlan.filters.selectQuartal')"
              class="w-full"
            />
          </div>

          <!-- Reset Filters Button -->
          <UButton
            v-if="isFiltered"
            icon="i-lucide-rotate-ccw"
            color="neutral"
            variant="ghost"
            size="xs"
            class="rounded-xl font-semibold"
            @click="resetFilters"
          >
            {{ t('strategicPlan.filters.reset') }}
          </UButton>
        </div>
      </div>
    </UCard>

    <!-- Strategic KPI Table -->
    <UCard variant="soft">
      <UTable :data="filteredObjectives" :columns="columns">
        <!-- Action Buttons -->
        <template #actions-cell="{ row }">
          <div class="flex items-center gap-1">
            <!-- Direct View Details Button -->
            <UButton
              icon="i-lucide-eye"
              color="neutral"
              variant="ghost"
              size="xs"
              :title="t('strategicPlan.actions.viewDetails')"
              :aria-label="t('strategicPlan.actions.viewDetails')"
              @click="store.openViewModal(row.original)"
            />
            <!-- Dropdown Menu -->
            <UDropdownMenu
              :items="getActions(row.original)"
              aria-label="Actions"
            >
              <UButton
                icon="i-lucide-ellipsis-vertical"
                color="neutral"
                variant="ghost"
                size="xs"
                aria-label="Actions dropdown"
              />
            </UDropdownMenu>
          </div>
        </template>

        <!-- Custom cell for status with color coding -->
        <template #status-cell="{ row }">
          <span
            :class="{
              'text-emerald-600 dark:text-emerald-400 font-bold': row.original.status === 'Good',
              'text-amber-600 dark:text-amber-400 font-bold': row.original.status === 'Moderate',
              'text-rose-600 dark:text-rose-400 font-bold': row.original.status === 'Poor',
            }"
          >
            {{ formatStatus(row.original.status) }}
          </span>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useVisionMissionGoalsStore } from '~/stores/vision-mission-goals'
import { useI18n } from '~/composables/useI18n'
import type { StrategicAuditPlan } from '~/types/audit'

const { t } = useI18n()
const store = useStrategicPlanStore()
const vmgStore = useVisionMissionGoalsStore()

const selectedGoalId = ref('ALL')
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
  { accessorKey: 'code', header: t('strategicPlan.columns.code'), cell: (row: any) => row.getValue() },
  { accessorKey: 'strategicObjective', header: t('strategicPlan.columns.objective') },
  { accessorKey: 'kpi', header: t('strategicPlan.columns.kpi') },
  { accessorKey: 'unit', header: t('strategicPlan.columns.unit') },
  { accessorKey: 'selectedPeriod', header: t('strategicPlan.columns.period') },
  { accessorKey: 'target', header: t('strategicPlan.columns.target') },
  { accessorKey: 'actual', header: t('strategicPlan.columns.actual') },
  { accessorKey: 'calculation', header: t('strategicPlan.columns.calculation') },
  { accessorKey: 'status', header: t('strategicPlan.columns.status'), cell: 'status-cell' },
  { accessorKey: 'actions', header: t('strategicPlan.columns.actions'), cell: 'actions-cell' },
])

const periodTypeOptions = computed(() => [
  { label: t('strategicPlan.filters.allTypes'), value: 'ALL' },
  { label: t('strategicPlan.filters.yearly'), value: 'Yearly' },
  { label: t('strategicPlan.filters.quartal'), value: 'Quartal' },
])

const yearFilterOptions = computed(() => [
  { label: t('strategicPlan.filters.allYears'), value: 'ALL' },
  { label: '2024', value: '2024' },
  { label: '2025', value: '2025' },
  { label: '2026', value: '2026' },
  { label: '2027', value: '2027' },
  { label: '2028', value: '2028' },
])

const quartalFilterOptions = computed(() => [
  { label: t('strategicPlan.filters.allQuarters'), value: 'ALL' },
  { label: 'Q1', value: 'Q1' },
  { label: 'Q2', value: 'Q2' },
  { label: 'Q3', value: 'Q3' },
  { label: 'Q4', value: 'Q4' },
])

const isFiltered = computed(() => {
  return searchQuery.value !== '' || 
         selectedPeriodType.value !== 'ALL' || 
         selectedYear.value !== 'ALL' || 
         selectedQuartal.value !== 'ALL'
})

const resetFilters = () => {
  searchQuery.value = ''
  selectedPeriodType.value = 'ALL'
  selectedYear.value = 'ALL'
  selectedQuartal.value = 'ALL'
}

const formatStatus = (status: string) => {
  if (!status) return '-'
  const lower = status.toLowerCase()
  if (['good', 'moderate', 'poor', 'pending', 'planned'].includes(lower)) {
    return t(`strategicPlan.status.${lower}`)
  }
  return status
}

const goalTabs = computed(() => {
  const tabs = [{ label: t('strategicPlan.filters.allObjectives'), value: 'ALL' }]
  if (vmgStore.activeVmg?.goals) {
    vmgStore.activeVmg.goals.forEach(g => {
      tabs.push({
        label: `${g.goal_code} - ${g.goal_name}`,
        value: g.id || g.goal_code
      })
    })
  }
  return tabs
})

const filteredObjectives = computed(() => {
  let list = store.strategicObjectives

  // 1. Goal Filter
  if (selectedGoalId.value !== 'ALL') {
    const activeGoal = vmgStore.activeVmg?.goals?.find(g => g.id === selectedGoalId.value || g.goal_code === selectedGoalId.value)
    const activeGoalCode = activeGoal?.goal_code
    list = list.filter(item => item.goalId === selectedGoalId.value || (activeGoalCode && item.goalId === activeGoalCode))
  }

  // 2. Period Type Filter (Yearly vs Quartal)
  if (selectedPeriodType.value !== 'ALL') {
    list = list.filter(item => item.periodType === selectedPeriodType.value)
  }

  // 3. Year Filter
  if (selectedYear.value !== 'ALL') {
    const yStr = selectedYear.value
    list = list.filter(item => {
      if (item.kpiTargets && item.kpiTargets[yStr] !== undefined) return true
      if (item.selectedPeriod === yStr) return true
      if (item.yearStart && item.yearEnd) {
        const yNum = parseInt(yStr)
        return yNum >= item.yearStart && yNum <= item.yearEnd
      }
      return false
    })
  }

  // 4. Quartal Filter (Q1, Q2, Q3, Q4)
  if (selectedQuartal.value !== 'ALL') {
    list = list.filter(item => item.selectedPeriod === selectedQuartal.value)
  }

  // 5. Search Query
  if (searchQuery.value.trim() !== '') {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(item => {
      return (item.code && item.code.toLowerCase().includes(q)) ||
             (item.strategicObjective && item.strategicObjective.toLowerCase().includes(q)) ||
             (item.kpi && item.kpi.toLowerCase().includes(q))
    })
  }

  // Recalculate target/actual if specific year selected
  if (selectedYear.value !== 'ALL') {
    const yr = selectedYear.value
    return list.map(item => {
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

const getActions = (item: StrategicAuditPlan) => [
  [
    {
      type: "label" as const,
      label: t('strategicPlan.actions.title'),
    },
    {
      label: t('strategicPlan.actions.viewDetails'),
      icon: "i-lucide-eye",
      onSelect: () => store.openViewModal(item),
    },
    {
      label: t('strategicPlan.actions.edit'),
      icon: "i-lucide-edit",
      onSelect: () => store.handleEdit(item),
    },
    {
      label: t('strategicPlan.actions.delete'),
      icon: "i-lucide-trash-2",
      onSelect: () => store.handleDelete(item.id),
    },
  ],
]
</script>