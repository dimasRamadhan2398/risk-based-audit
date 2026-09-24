<template>
  <div>
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Audit Result Report</h1>
        <p class="text-gray-500">Finalize and publish audit results and findings</p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <UButton
          v-if="canImportPlanDocs"
          color="neutral"
          variant="outline"
          icon="i-lucide-upload"
          label="Import LHA Document"
          to="/audit-result-report/upload"
          class="font-bold shadow"
        />
      </div>
    </div>

    <!-- Assignment Letter Selector -->
    <UCard class="mb-6" :ui="{ body: 'p-4' }">
      <div class="flex flex-col md:flex-row md:items-center gap-4">
        <div class="flex-1">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
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
        <div v-else-if="store.selectedAssignmentLetter" class="text-sm text-green-600 flex items-center gap-2">
          <UIcon name="i-heroicons-check-circle" class="size-4 inline" />
          <span>Audit: <strong>{{ store.selectedAssignmentLetter }}</strong></span>
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
        <TableEntities :data="store.filteredReports" :columns="columns">
          <template #reportNumber-cell="{ row }">
            <span class="font-mono text-md font-semibold text-primary-600 dark:text-primary-400">
              {{ row.original.reportNumber || (row.original as any).report_number || '-' }}
            </span>
          </template>
          <template #reportDate-cell="{ row }">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
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
              <div v-for="group in getGroupedFindings(row.original.findings)" :key="group.category" class="h-6 flex items-center justify-center bg-gray-50 dark:bg-gray-800 rounded">
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
                      <h4 class="text-md font-bold text-gray-900 dark:text-white mb-2 border-b pb-1">{{ group.category }} Findings</h4>
                      <ul class="list-disc pl-4 text-md text-gray-700 dark:text-gray-300 space-y-1">
                        <li v-for="(item, idx) in group.items" :key="idx" class="leading-relaxed">
                          {{ item.title }}
                          <span v-if="item.source" class="text-[10px] text-gray-400 block italic">({{ item.source }})</span>
                        </li>
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
                      <h4 class="text-md font-bold text-gray-900 dark:text-white mb-2 border-b pb-1">{{ group.category }} Actions</h4>
                      <ul class="list-disc pl-4 text-md text-gray-700 dark:text-gray-300 space-y-1">
                        <li v-for="(item, idx) in group.items" :key="idx" class="leading-relaxed">
                          <span class="font-semibold block mb-0.5 text-gray-800 dark:text-gray-200">{{ item.title }}:</span>
                          <span class="text-primary-700 dark:text-primary-400 block mb-1">{{ item.action || 'No action defined' }}</span>
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
              <UTooltip text="Sync Temuan Otomatis dari KKA & Fieldwork">
                <UButton
                  color="primary"
                  variant="soft"
                  icon="i-heroicons-sparkles"
                  size="sm"
                  label="Sync"
                  @click="syncReportFindings(row.original as any)"
                />
              </UTooltip>
              <UTooltip text="Generate LHA (.docx)">
                <UButton
                  color="success"
                  variant="soft"
                  icon="i-heroicons-arrow-down-tray"
                  size="sm"
                  label="Docx"
                  @click="store.downloadDocx((row.original as any).id, (row.original as any).reportNumber)"
                />
              </UTooltip>
              <UTooltip text="Edit">
                <UButton
                  color="warning"
                  variant="ghost"
                  icon="i-lucide-edit"
                  size="md"
                  @click="store.editReport(row.original as any)"
                />
              </UTooltip>
              <UTooltip text="Cetak">
                <UButton
                  color="neutral"
                  variant="ghost"
                  icon="i-lucide-printer"
                  size="md"
                  @click="printReport(row.original as any)"
                />
              </UTooltip>
            </div>
          </template>
        </TableEntities>
      </UCard>

      <div v-else class="text-center py-16 bg-gray-50 dark:bg-gray-850/50 rounded-xl border-2 border-dashed border-gray-200 dark:border-gray-800">
        <div class="p-4 bg-primary-50 dark:bg-primary-950/40 rounded-full w-fit mx-auto mb-4 text-primary-600">
          <UIcon name="i-heroicons-sparkles" class="size-12" />
        </div>
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">Belum Ada Laporan Hasil Audit (LHA)</h3>
        <p class="text-gray-500 mt-2 max-w-md mx-auto mb-6 text-sm">
          Belum ada laporan yang dibuat untuk surat tugas ini. Anda dapat membuat laporan dengan temuan audit yang langsung terisi otomatis dari modul KKA dan Fieldwork.
        </p>
        <div class="flex justify-center gap-3">
          <UButton
            color="primary"
            icon="i-heroicons-sparkles"
            label="Buat Laporan dengan Temuan Otomatis"
            size="lg"
            class="font-bold shadow"
            @click="store.openModal"
          />
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-16">
      <UIcon name="i-heroicons-document-magnifying-glass" class="size-20 text-gray-200 dark:text-gray-800 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-700 dark:text-gray-300 text-center">Select Assignment Letter</h3>
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
import { useRbac } from '~/composables/useRbac'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'

const store = useAuditResultReportStore()
const { canImportPlanDocs } = useRbac()
const toast = useToastNotification()

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

const syncReportFindings = async (report: any) => {
  store.editReport(report)
  await store.runAutoDetectFindings('merge')
}

const printReport = (report: any) => {
  toast.showSuccess(`Printing report: ${report.reportTitle}`)
}
</script>

