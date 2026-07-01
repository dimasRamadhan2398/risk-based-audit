<script setup lang="ts">
import { useAuditExecutionStore } from '~/stores/audit-execution'
import AuditExecutionDetailModal from '~/components/audit-execution/AuditExecutionDetailModal.vue'
import { AuditCategory, AuditDepartment, AuditStatus } from '~/types/audit'

const store = useAuditExecutionStore()
store.fetchAuditExecutions()
const { auditExecutions, getSummary } = storeToRefs(store)

const search = ref('')
const quarter = ref('')
const status = ref<AuditStatus | undefined>(undefined)
const department = ref<AuditDepartment | undefined>(undefined)

const quarters = ['Quarter I', 'Quarter II', 'Quarter III', 'Quarter IV']

const columns = [
  { accessorKey: 'name', header: 'Audit Name' },
  { accessorKey: 'progress', header: 'Execution Progress' },
  { accessorKey: 'lead_auditor', header: 'Lead Auditor' },
  { accessorKey: 'actions', header: 'Action' }
]

const filteredAudits = computed(() => {
  return auditExecutions.value.filter(audit => {
    const matchesSearch = audit.name.toLowerCase().includes(search.value.toLowerCase()) || 
                          audit.category.toLowerCase().includes(search.value.toLowerCase())
    const matchesDept = !AuditDepartment.FINANCE || audit.category === AuditCategory.ASSURANCE
    const matchesStatus = !status.value || audit.status === status.value
    return matchesSearch && matchesDept && matchesStatus
  })
})

const isDetailOpen = ref(false)
const selectedAudit = ref<any>(undefined)

const openDetail = (audit: any) => {
  selectedAudit.value = audit.original
  isDetailOpen.value = true
}

const handleRemind = (audit: any) => {
  useToast().add({
    title: 'Reminder Sent',
    description: `Reminder for audit ${audit.nama_audit} has been sent to ${audit.lead_auditor}`,
    color: 'success'
  })
}
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Audit Execution Status</h1>
    </div>

    <!-- Summary Section -->
    <div class="bg-white dark:bg-gray-900 rounded-xl p-4 border border-gray-200 dark:border-gray-800 shadow-sm">
      <p class="text-sm font-semibold text-gray-500 mb-3">Summary Status:</p>
      <div class="flex items-center gap-6">
        <div class="flex items-center gap-2">
          <span class="w-3 h-3 rounded-full bg-emerald-500"></span>
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ getSummary.completed }} Completed</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="w-3 h-3 rounded-full bg-blue-500"></span>
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ getSummary.inProgress }} In Progress</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="w-3 h-3 rounded-full bg-gray-300"></span>
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ getSummary.planned }} Planned</span>
        </div>
      </div>
    </div>

    <!-- Filters Section -->
    <div class="flex flex-wrap items-center gap-4">
      <UInput
        v-model="search"
        icon="i-lucide-search"
        placeholder="Search Audit"
        class="w-48"
      />
      <USelectMenu
        v-model="quarter"
        :items="quarters"
        placeholder="Quarter"
        class="w-48"
      />
      <USelectMenu
        v-model="department"
        :items="Object.values(AuditDepartment)"
        placeholder="Category"
        class="w-48"
      />
      <USelectMenu
        v-model="status"
        :items="Object.values(AuditStatus)"
        placeholder="Status"
        class="w-48"
      />
    </div>

    <!-- Table Section -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
      <UTable :columns="columns" :data="filteredAudits" :ui="{ th: 'bg-gray-50 dark:bg-gray-800/50' }">
        <template #nama_audit-cell="{ row }">
          <div class="flex flex-col">
            <span class="font-bold text-gray-900 dark:text-white">{{ row.original.name }}</span>
            <span class="text-xs text-gray-500">({{ row.original.category }})</span>
          </div>
        </template>

        <template #progress-cell="{ row }">
          <div class="flex items-center gap-3 min-w-[200px]">
            <UProgress :value="row.original.progress" color="secondary" class="flex-1" />
            <span class="text-sm font-bold text-secondary">{{ row.original.progress }} %</span>
          </div>
        </template>

        <template #lead_auditor-data="{ row }">
          <span class="text-sm text-gray-700 dark:text-gray-300">{{ row.original.lead_auditor }}</span>
        </template>

        <template #actions-cell="{ row }">
          <UButton
            icon="i-lucide-eye"
            variant="ghost"
            color="neutral"
            @click="openDetail(row)"
          />
        </template>
      </UTable>

      <!-- Pagination Placeholder -->
      <div class="px-4 py-3 border-t border-gray-200 dark:border-gray-800 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <UButton icon="i-lucide-chevron-left" variant="ghost" color="neutral" size="xs" />
          <span class="text-sm font-medium">1 / 10</span>
          <UButton icon="i-lucide-chevron-right" variant="ghost" color="neutral" size="xs" />
        </div>
        <span class="text-xs text-gray-500">Showing 1 - {{ filteredAudits.length }} of {{ auditExecutions.length }} data</span>
      </div>
    </div>

    <!-- Detail Modal -->
    <AuditExecutionDetailModal
      v-model:open="isDetailOpen"
      :audit="selectedAudit"
      @remind="handleRemind"
    />
  </div>
</template>
