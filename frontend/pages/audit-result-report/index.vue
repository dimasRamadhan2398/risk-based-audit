<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Audit Result Report</h1>
        <p class="text-gray-500">Finalize and publish audit results and findings</p>
      </div>
      <!-- <UButton
        v-if="store.hasSelectedAssignmentLetter"
        color="primary"
        icon="i-heroicons-plus"
        label="Create New Report"
        @click="store.openModal"
      /> -->
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
          <template #reportNumber-cell="{ row }">
            <span class="font-mono text-md font-semibold text-primary-600 dark:text-primary-400">
              {{ row.original.reportNumber || (row.original as any).report_number || '-' }}
            </span>
          </template>
          <template #reportDate-cell="{ row }">
            <span class="text-sm font-medium text-gray-700">
              {{ row.original.reportDate || (row.original as any).report_date?.split('T')[0] || '-' }}
            </span>
          </template>
          <template #status-cell="{ row }">
            <UBadge :color="row.original.status === 'Final' ? 'success' : 'neutral'" variant="soft">
              {{ row.original.status }}
            </UBadge>
          </template>
          <template #findingsCount-cell="{ row }">
            <div class="flex flex-col gap-2 min-w-[30px]">
              <div v-for="group in getGroupedFindings(row.original.findings)" :key="group.category" class="h-6 flex items-center justify-center bg-gray-50 rounded">
                <span class="text-md font-semibold">{{ group.count }}</span>
              </div>
            </div>
          </template>
          <template #category-cell="{ row }">
            <div class="flex flex-col gap-2">
              <div v-for="group in getGroupedFindings(row.original.findings)" :key="group.category" class="h-6 flex items-center">
                <UBadge :color="getRatingColor(group.category)" variant="subtle" size="sm" class="w-full justify-center">
                  {{ group.category }}
                </UBadge>
              </div>
            </div>
          </template>
          <template #listOfFinding-cell="{ row }">
            <div class="flex flex-col gap-2">
              <div v-for="group in getGroupedFindings(row.original.findings)" :key="group.category" class="h-6 flex items-center justify-center">
                <UPopover v-if="group.count > 0" mode="hover">
                  <UButton color="neutral" variant="soft" size="md" trailing-icon="i-heroicons-chevron-down" class="text-md">
                    View {{ group.count }} Findings
                  </UButton>
                  <template #content>
                    <div class="p-3 max-w-sm max-h-60 overflow-y-auto">
                      <h4 class="text-md font-bold text-gray-900 mb-2 border-b pb-1">{{ group.category }} Findings</h4>
                      <ul class="list-disc pl-4 text-md text-gray-700 space-y-1">
                        <li v-for="(item, idx) in group.items" :key="idx" class="leading-relaxed">{{ item.title }}</li>
                      </ul>
                    </div>
                  </template>
                </UPopover>
                <span v-else class="text-md text-gray-400">-</span>
              </div>
            </div>
          </template>
          <template #action-cell="{ row }">
            <div class="flex flex-col gap-2">
              <div v-for="group in getGroupedFindings(row.original.findings)" :key="group.category" class="h-6 flex items-center justify-center">
                <UPopover v-if="group.count > 0" mode="hover">
                  <UButton color="neutral" variant="soft" size="md" trailing-icon="i-heroicons-chevron-down" class="text-md">
                    View Actions
                  </UButton>
                  <template #content>
                    <div class="p-3 max-w-sm max-h-60 overflow-y-auto">
                      <h4 class="text-md font-bold text-gray-900 mb-2 border-b pb-1">{{ group.category }} Actions</h4>
                      <ul class="list-disc pl-4 text-md text-gray-700 space-y-1">
                        <li v-for="(item, idx) in group.items" :key="idx" class="leading-relaxed">
                          <span class="font-semibold block mb-0.5 text-gray-800">{{ item.title }}:</span>
                          <span class="text-primary-700 block mb-1">{{ item.action || 'No action defined' }}</span>
                        </li>
                      </ul>
                    </div>
                  </template>
                </UPopover>
                <span v-else class="text-md text-gray-400">-</span>
              </div>
            </div>
          </template>
          <template #actions-cell="{ row }">
            <div class="flex gap-2 items-center">
              <UButton
                color="success"
                variant="soft"
                icon="i-heroicons-arrow-down-tray"
                size="sm"
                label="Docx"
                title="Download LHA (.docx)"
                @click="store.downloadDocx((row.original as any).id, (row.original as any).reportNumber)"
              />
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
  { accessorKey: 'reportNumber', header: 'No. LHA / ID' },
  { accessorKey: 'reportTitle', header: 'Report Title' },
  { accessorKey: 'reportDate', header: 'Date' },
  { accessorKey: 'findingsCount', header: 'Findings' },
  { accessorKey: 'category', header: 'Category' },
  { accessorKey: 'listOfFinding', header: 'List of Finding' },
  { accessorKey: 'action', header: 'Action' },
  { accessorKey: 'status', header: 'Status' },
  { accessorKey: 'actions', header: '' }
]

const getRatingColor = (rating: any) => {
  switch (rating) {
    case 'Very Significant': return 'error'
    case 'Significant': return 'warning'
    case 'Quite Significant': return 'info'
    case 'Not Significant': return 'success'
    default: return 'neutral'
  }
}

const CATEGORY_ORDER = ['Very Significant', 'Significant', 'Quite Significant', 'Not Significant']

const getGroupedFindings = (findings: any[] | undefined) => {
  if (!findings || findings.length === 0) return CATEGORY_ORDER.map(c => ({ category: c, count: 0, items: [] }))
  
  return CATEGORY_ORDER.map(cat => {
    const items = findings.filter((f: any) => f.category === cat)
    return {
      category: cat,
      count: items.length,
      items
    }
  })
}

const printReport = (report: any) => {
  alert(`Printing report: ${report.reportTitle}`)
}
</script>