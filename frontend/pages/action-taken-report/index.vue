<script setup lang="ts">

import ATRDetail from '~/components/action-taken-report/ATRDetail.vue'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { AuditDepartment, AuditStatus } from '~/types/audit'
import type { ActionTakenReport } from '~/types/audit'

const store = useActionTakenReportStore()

interface TableColumn<T> {
  accessorKey: keyof T | string
  header: string
  sortable?: boolean
}

const columns: TableColumn<ActionTakenReport>[] = [
  { accessorKey: 'auditRef', header: 'Audit Reference', sortable: true },
  { accessorKey: 'title', header: 'Title', sortable: true },
  { accessorKey: 'department', header: 'Department', sortable: true },
  { accessorKey: 'deadline', header: 'Deadline', sortable: true },
  { accessorKey: 'status', header: 'Status', sortable: true },
  { accessorKey: 'actions', header: 'Actions' }
]

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
  if (status === AuditStatus.CANCELLED) return 'Cancelled'
  if (status === AuditStatus.IN_PROGRESS) return 'In Progress'
  if (status === AuditStatus.COMPLETED) return 'Completed'
  if (status === AuditStatus.PLANNED) return 'Planned'
}

const page = ref(1)
const pageCount = 5
const items = computed(() => {
  return store.filteredReports.slice((page.value - 1) * pageCount, (page.value) * pageCount)
})
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col space-y-4">
      <h1 class="text-2xl font-bold">Action Taken Report</h1>
      
      <div class="space-y-2">
        <p class="text-sm font-bold">Status Summary:</p>
        <div class="flex items-center space-x-6">
          <div class="flex items-center space-x-2">
            <div class="w-4 h-4 rounded-full bg-emerald-500"></div>
            <span class="text-sm font-bold">{{ store.stats.donePercent }}% Done</span>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-4 h-4 rounded-full bg-amber-400"></div>
            <span class="text-sm font-bold">{{ store.stats.wipPercent }}% Work In Progress</span>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-4 h-4 rounded-full bg-rose-500"></div>
            <span class="text-sm font-bold">{{ store.stats.latePercent }}% Late</span>
          </div>
        </div>
      </div>
    </div>

    <div class="flex flex-wrap items-center gap-4">
      <UInput
        v-model="store.searchQuery"
        icon="i-lucide-search"
        placeholder="Search Finding"
        class="w-full max-w-xs"
        size="md"
      />
      <USelectMenu
        v-model="(store.selectedDepartment as any)"
        :items="Object.values(AuditDepartment)"
        placeholder="Choose Department"
        class="w-full max-w-[200px]"
        size="md"
      />
      <USelectMenu
        v-model="(store.selectedStatus as any)"
        :items="Object.values(AuditStatus)"
        placeholder="Choose Status"
        class="w-full max-w-[200px]"
        size="md"
      />
    </div>

    <UCard class="overflow-hidden border border-gray-200 dark:border-gray-800" :ui="{ body: 'p-0' }">
      <UTable
        :columns="columns"
        :data="items"
        @select="(_e, row) => store.openDetail(row as any)"
        class="w-full"
        :ui="{ 
          thead: 'bg-gray-100 dark:bg-gray-800/50',
          th: 'font-bold py-4',
          td: 'py-4'
        }"
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
      </UTable>

      <div class="flex items-center justify-between px-6 py-4 border-t border-gray-200 dark:border-gray-700">
        <div class="flex items-center space-x-4">
           <UButton
            icon="i-lucide-chevron-left"
            color="neutral"
            variant="ghost"
            :disabled="page === 1"
            @click="page--"
          />
          <span class="text-sm font-medium">{{ page }} / {{ Math.ceil(store.filteredReports.length / pageCount) }}</span>
          <UButton
            icon="i-lucide-chevron-right"
            color="neutral"
            variant="ghost"
            :disabled="page >= Math.ceil(store.filteredReports.length / pageCount)"
            @click="page++"
          />
        </div>
        <p class="text-sm font-bold">
          Showing {{ (page - 1) * pageCount + 1 }} - {{ Math.min(page * pageCount, store.filteredReports.length) }} of {{ store.filteredReports.length }} data
        </p>
      </div>
    </UCard>

    <ATRDetail />
    
  </div>
</template>
