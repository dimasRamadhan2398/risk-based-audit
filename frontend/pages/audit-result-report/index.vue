<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Audit Result Report</h1>
        <p class="text-gray-500">Finalize and publish audit results and findings</p>
      </div>
      <UButton
        v-if="store.hasSelectedAssignmentLetter"
        color="primary"
        icon="i-heroicons-plus"
        label="Create New Report"
        @click="store.openModal"
      />
    </div>

    <!-- Assignment Letter Selector -->
    <UCard class="mb-6" :ui="{ body: 'p-4' }">
      <div class="flex flex-col md:flex-row md:items-center gap-4">
        <div class="flex-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Select Assignment Letter (Audit)
          </label>
          <USelectMenu
            v-model="store.selectedAssignmentLetter"
            :items="store.publishedAssignmentLetters"
            placeholder="Select Assignment Letter to audit"
            class="w-full md:w-96"
            :disabled="store.publishedAssignmentLetters.length === 0"
          >
            <template #leading>
              <UIcon name="i-heroicons-document-text" class="size-5" />
            </template>
          </USelectMenu>
        </div>
        <div v-if="store.publishedAssignmentLetters.length === 0" class="text-sm text-amber-600">
          <UIcon name="i-heroicons-exclamation-triangle" class="size-4 inline mr-1" />
          No Assignment Letter with Published status. Please create and publish an Assignment Letter first.
        </div>
        <div v-else-if="store.selectedAssignmentLetter" class="text-sm text-green-600">
          <UIcon name="i-heroicons-check-circle" class="size-4 inline mr-1" />
          Audit: {{ store.selectedAssignmentLetter }}
        </div>
        <div v-else class="text-sm text-gray-600">
          <UIcon name="i-heroicons-document-text" class="size-4 inline mr-1" />
          No Assignment Letter selected
        </div>
      </div>
    </UCard>

    <!-- Main Content -->
    <div v-if="store.hasSelectedAssignmentLetter">
      <UCard v-if="store.filteredReports.length > 0" class="overflow-hidden overflow-x-auto">
        <UTable :data="store.filteredReports" :columns="columns">
          <template #status-data="{ row }">
            <UBadge :color="row.original.status === 'Final' ? 'success' : 'neutral'" variant="soft">
              {{ row.original.status }}
            </UBadge>
          </template>
          <template #overallRating-data="{ row }">
            <UBadge :color="getRatingColor(row.original.overallRating)" variant="subtle">
              {{ row.original.overallRating }}
            </UBadge>
          </template>
          <template #actions-cell="{ row }">
            <div class="flex gap-2">
              <UButton
                color="primary"
                variant="ghost"
                icon="i-heroicons-pencil-square"
                size="sm"
                @click="store.editReport(row.original as any)"
              />
              <UButton
                color="error"
                variant="ghost"
                icon="i-heroicons-trash"
                size="sm"
                @click="store.deleteReport((row.original as any).id)"
              />
              <UButton
                color="neutral"
                variant="ghost"
                icon="i-heroicons-printer"
                size="sm"
                @click="printReport(row.original as any)"
              />
            </div>
          </template>
        </UTable>
      </UCard>

      <div v-else class="text-center py-16 bg-gray-50 rounded-xl border-2 border-dashed border-gray-200">
        <UIcon name="i-heroicons-document-plus" class="size-16 text-gray-300 mx-auto mb-4" />
        <h3 class="text-lg font-semibold text-gray-700">No Reports Created</h3>
        <p class="text-gray-500 mt-2 max-w-md mx-auto mb-6">
          No audit result reports have been created for this assignment letter yet.
        </p>
        <UButton
          color="primary"
          icon="i-heroicons-plus"
          label="Create First Report"
          @click="store.openModal"
        />
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-16">
      <UIcon name="i-heroicons-document-magnifying-glass" class="size-20 text-gray-200 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-700 text-center">Select Assignment Letter</h3>
      <p class="text-gray-500 mt-2 max-w-md mx-auto text-center">
        Please select an assignment letter to view or manage its audit result reports.
      </p>
    </div>

    <!-- Report Form Modal -->
    <ResultReportForm />
  </div>
</template>

<script setup lang="ts">
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import ResultReportForm from '~/components/audit-result-report/ResultReportForm.vue'

const store = useAuditResultReportStore()

const columns = [
  { accessorKey: 'reportTitle', header: 'Report Title' },
  { accessorKey: 'reportDate', header: 'Date' },
  { accessorKey: 'findingsCount', header: 'Findings' },
  { accessorKey: 'overallRating', header: 'Overall Rating' },
  { accessorKey: 'status', header: 'Status' },
  { accessorKey: 'actions', header: 'Actions' }
]

const getRatingColor = (rating: any) => {
  switch (rating) {
    case 'Satisfactory': return 'success'
    case 'Needs Improvement': return 'warning'
    case 'Unsatisfactory': return 'error'
    default: return 'neutral'
  }
}

const printReport = (report: any) => {
  alert(`Printing report: ${report.reportTitle}`)
}
</script>