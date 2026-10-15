<script setup lang="ts">

import ATRDetail from '~/components/action-taken-report/ATRDetail.vue'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { AuditDepartment, AuditStatus } from '~/types/audit'
import type { ActionTakenReport } from '~/types/audit'

const store = useActionTakenReportStore()
const { t } = useI18n()

interface TableColumn<T> {
  accessorKey: keyof T | string
  header: string
  sortable?: boolean
}

const columns = computed(() => [
  { accessorKey: 'auditRef', header: t('actionTakenReport.columns.auditReference'), sortable: true },
  { accessorKey: 'title', header: t('actionTakenReport.columns.title'), sortable: true },
  { accessorKey: 'department', header: t('actionTakenReport.columns.department'), sortable: true },
  { accessorKey: 'deadline', header: t('actionTakenReport.columns.deadline'), sortable: true },
  { accessorKey: 'status', header: t('actionTakenReport.columns.status'), sortable: true },
  { accessorKey: 'actions', header: t('actionTakenReport.columns.actions') }
])

const getStatusColor = (status: string) => {
  switch (status) {
    case AuditStatus.COMPLETED: return 'bg-green-500'
    case AuditStatus.IN_PROGRESS: return 'bg-yellow-500'
    case AuditStatus.PLANNED: return 'bg-gray-500'
    case AuditStatus.CANCELLED: return 'bg-red-500'
    default: return 'bg-gray-500'
  }
}

const getStatusLabel = (status: string) => {
  if (status === AuditStatus.CANCELLED) return t('actionTakenReport.statuses.cancelled')
  if (status === AuditStatus.IN_PROGRESS) return t('actionTakenReport.statuses.inProgress')
  if (status === AuditStatus.COMPLETED) return t('actionTakenReport.statuses.completed')
  if (status === AuditStatus.PLANNED) return t('actionTakenReport.statuses.planned')
  return status
}

const statusOptions = computed(() => Object.values(AuditStatus).map(status => ({
  label: getStatusLabel(status),
  value: status
})))

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
      <h1 class="text-2xl font-bold">{{ t('actionTakenReport.title') }}</h1>
      
      <div class="space-y-2">
        <p class="text-sm font-bold">{{ t('actionTakenReport.statusSummary') }}</p>
        <div class="flex flex-wrap items-center gap-3 sm:gap-6">
          <div class="flex items-center space-x-2">
            <div class="w-4 h-4 rounded-full bg-emerald-500"></div>
            <span class="text-sm font-bold">{{ t('actionTakenReport.summary.done', { percent: store.stats.donePercent }) }}</span>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-4 h-4 rounded-full bg-amber-400"></div>
            <span class="text-sm font-bold">{{ t('actionTakenReport.summary.inProgress', { percent: store.stats.wipPercent }) }}</span>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-4 h-4 rounded-full bg-rose-500"></div>
            <span class="text-sm font-bold">{{ t('actionTakenReport.summary.late', { percent: store.stats.latePercent }) }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="flex flex-wrap items-center gap-3 sm:gap-4">
      <UInput
        v-model="store.searchQuery"
        icon="i-lucide-search"
        :placeholder="t('actionTakenReport.filters.searchPlaceholder')"
        class="w-full sm:max-w-md"
        size="md"
      />
      <USelectMenu
        v-model="(store.selectedDepartment as any)"
        :items="Object.values(AuditDepartment)"
        :placeholder="t('actionTakenReport.filters.departmentPlaceholder')"
        class="w-full sm:w-52"
        size="md"
      />
      <USelectMenu
        v-model="(store.selectedStatus as any)"
        :items="statusOptions"
        value-key="value"
        :placeholder="t('actionTakenReport.filters.statusPlaceholder')"
        class="w-full sm:w-52"
        size="md"
      />
      <UButton
        :label="t('actionTakenReport.filters.reset')"
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
