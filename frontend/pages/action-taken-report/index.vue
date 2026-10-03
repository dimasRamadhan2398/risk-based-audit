<script setup lang="ts">

import ATRDetail from '~/components/action-taken-report/ATRDetail.vue'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { AuditDepartment, AuditStatus } from '~/types/audit'
import { useI18n } from '~/composables/useI18n'
import { ATR_OVERDUE_FILTER, ATR_STATUSES, atrStatusI18nKey, type AtrSliceKey } from '~/utils/actionTakenReport'

const store = useActionTakenReportStore()
const { t } = useI18n()

interface TableColumn<T> {
  accessorKey: keyof T | string
  header: string
  sortable?: boolean
}

const columns = [
  { accessorKey: 'auditRef', header: 'Audit Reference', sortable: true },
  { accessorKey: 'title', header: 'Title', sortable: true },
  { accessorKey: 'department', header: 'Department', sortable: true },
  { accessorKey: 'deadline', header: 'Deadline', sortable: true },
  { accessorKey: 'status', header: 'Status', sortable: true },
  { accessorKey: 'actions', header: 'Actions' }
]

// Same palette for the summary dots and the table status dots; Overdue is reserved for the red badge.
const sliceColors: Record<AtrSliceKey, string> = {
  completed: 'bg-emerald-500',
  inProgress: 'bg-amber-400',
  planned: 'bg-sky-400',
  overdue: 'bg-rose-500',
  cancelled: 'bg-gray-400'
}

const sliceLabelKeys: Record<AtrSliceKey, string> = {
  completed: 'actionTakenReport.summary.done',
  inProgress: 'actionTakenReport.summary.inProgress',
  planned: 'actionTakenReport.summary.planned',
  overdue: 'actionTakenReport.summary.overdue',
  cancelled: 'actionTakenReport.summary.cancelled'
}

const getStatusColor = (status: string) => {
  switch (status) {
    case AuditStatus.COMPLETED: return sliceColors.completed
    case AuditStatus.IN_PROGRESS: return sliceColors.inProgress
    case AuditStatus.PLANNED: return sliceColors.planned
    case AuditStatus.CANCELLED: return sliceColors.cancelled
    default: return 'bg-gray-300'
  }
}

const getStatusLabel = (status: string) => {
  const key = atrStatusI18nKey(status)
  return key ? t(key) : status
}

const statusOptions = computed(() => [
  ...ATR_STATUSES.map(status => ({ label: getStatusLabel(status), value: status as string })),
  { label: t('actionTakenReport.status.overdue'), value: ATR_OVERDUE_FILTER }
])

const page = ref(1)
const pageCount = 5
const resetFilters = () => {
  store.searchQuery = ''
  store.selectedDepartment = ''
  store.selectedStatus = ''
  page.value = 1
}
watch(() => [store.searchQuery, store.selectedDepartment, store.selectedStatus], () => { page.value = 1 })
const items = computed(() => {
  return store.filteredReports.slice((page.value - 1) * pageCount, (page.value) * pageCount)
})
</script>

<template>
  <div class="p-4 sm:p-6 space-y-6 min-w-0">
    <div class="flex flex-col space-y-4">
      <h1 class="text-2xl font-bold">Action Taken Report</h1>
      
      <div class="space-y-2">
        <p class="text-sm font-bold">{{ t('actionTakenReport.statusSummary') }}</p>
        <div class="flex flex-wrap items-center gap-3 sm:gap-6">
          <div v-for="slice in store.stats.breakdown" :key="slice.key" class="flex items-center space-x-2">
            <div :class="['w-4 h-4 rounded-full', sliceColors[slice.key]]"></div>
            <span class="text-sm font-bold">{{ t(sliceLabelKeys[slice.key], { percent: slice.percent }) }}</span>
          </div>
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('actionTakenReport.summaryHint') }}</p>
      </div>
    </div>

    <div class="flex flex-wrap items-center gap-3 sm:gap-4">
      <UInput
        v-model="store.searchQuery"
        icon="i-lucide-search"
        placeholder="Search Finding"
        class="w-full sm:max-w-md"
        size="md"
      />
      <USelectMenu
        v-model="(store.selectedDepartment as any)"
        :items="Object.values(AuditDepartment)"
        placeholder="Choose Department"
        class="w-full sm:w-52"
        size="md"
      />
      <USelectMenu
        v-model="(store.selectedStatus as any)"
        :items="statusOptions"
        value-key="value"
        :placeholder="t('actionTakenReport.filter.statusPlaceholder')"
        class="w-full sm:w-52"
        size="md"
      />
      <UButton
        label="Reset Filter"
        icon="i-lucide-rotate-ccw"
        color="neutral"
        variant="outline"
        size="md"
        class="w-full sm:w-auto"
        @click="resetFilters"
      />
    </div>

    <UCard class="overflow-hidden border border-gray-200 dark:border-gray-800" :ui="{ body: 'p-0' }">
      <TableEntities
        :columns="columns"
        :data="items"
        class="w-full"
      >
        <template #status-cell="{ row }: { row: any }">
          <div class="flex items-center space-x-2">
            <div :class="['w-4 h-4 rounded-full', getStatusColor(row.original.status)]"></div>
            <span class="text-sm font-medium">{{ getStatusLabel(row.original.status) }}</span>
            <UBadge
              v-if="row.original.isOverdue"
              color="error"
              variant="subtle"
              size="sm"
              :label="t('actionTakenReport.status.overdue')"
            />
          </div>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center space-x-2">
            <UButton
              icon="i-lucide-eye"
              color="neutral"
              variant="ghost"
              @click="store.openDetail(row.original)"
            />
          </div>
        </template>
      </TableEntities>
    </UCard>

    <ATRDetail />
    
  </div>
</template>
