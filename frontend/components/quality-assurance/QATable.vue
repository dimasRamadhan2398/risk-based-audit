<template>
    <UCard class="overflow-hidden border border-gray-200 dark:border-gray-800" :ui="{ body: 'p-0' }">
      <UTable
        :columns="store.columns"
        :data="store.items"
        class="w-full"
        :ui="{ 
          thead: 'bg-gray-100 dark:bg-gray-800/50',
          th: 'font-bold py-4',
          td: 'py-4'
        }"
      >
        <template #type-cell="{ row }: { row: any }">
          <div class="flex items-center space-x-2">
            <div :class="['w-4 h-4 rounded-full', store.getTypeIconColor(row.original.type)]"></div>
            <span class="text-sm font-medium">
              {{ formatTypeLabel(row.original.type) }}
            </span>
          </div>
        </template>

        <template #period-cell="{ row }: { row: any }">
          <span class="font-bold">{{ row.original.period }}</span>
        </template>

        <template #reportName-cell="{ row }: { row: any }">
          <span class="font-bold">{{ row.original.reportName }}</span>
        </template>

        <template #result-cell="{ row }: { row: any }">
          <span class="font-bold">
            {{ (store.matchQAType(row.original.type, QAType.QAR) || store.matchQAType(row.original.type, QAType.SAIV)) ? formatOverallConclusion(row.original.result) : row.original.result }}
          </span>
        </template>

        <template #status-cell="{ row }: { row: any }">
          <div class="flex items-center space-x-2">
            <div :class="['w-4 h-4 rounded-full', store.getStatusColor(row.original.status)]"></div>
            <span class="text-sm font-medium">{{ row.original.status }}</span>
          </div>
        </template>

        <template #conductedBy-cell="{ row }">
          <span class="font-medium text-gray-700 dark:text-gray-300">{{ row.original.conductedBy || '-' }}</span>
        </template>

        <template #viewReport-cell="{ row }">
          <UButton
            v-if="row.original.attachment?.filePath"
            icon="i-lucide-file-text"
            label="View"
            color="neutral"
            variant="subtle"
            size="sm"
            class="font-bold border border-gray-200 dark:border-gray-700"
            @click="openReportFile(row.original)"
          />
          <span v-else class="text-md text-gray-400 font-medium italic">No File</span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-1">
            <UButton
              icon="i-lucide-eye"
              label="Details"
              color="neutral"
              variant="ghost"
              size="sm"
              class="font-bold"
              @click="store.openDetail(row.original)"
            />

            <UButton
              icon="i-lucide-edit"
              label="Edit"
              color="warning"
              variant="ghost"
              size="sm"
              class="font-bold"
              @click="store.editReport(row.original)"
            />

            <UButton
              icon="i-lucide-trash"
              label="Delete"
              color="error"
              variant="ghost"
              size="sm"
              class="font-bold"
              @click="store.deleteReport(row.original)"
            />
          </div>
        </template>
      </UTable>

      <!-- Pagination -->
      <div class="flex items-center justify-between px-6 py-4 border-t border-gray-200 dark:border-gray-700">
        <div class="flex items-center space-x-4">
           <UButton
            icon="i-lucide-chevron-left"
            color="neutral"
            variant="ghost"
            :disabled="store.page === 1"
            @click= "() => {store.page--}"
          />
          <span class="text-sm font-medium">{{ store.page }} / {{ Math.ceil(store.filteredReports.length / store.pageCount) }}</span>
          <UButton
            icon="i-lucide-chevron-right"
            color="neutral"
            variant="ghost"
            :disabled="store.page >= Math.ceil(store.filteredReports.length / store.pageCount)"
            @click="() => {store.page++}"
          />
        </div>
        <div class="space-y-1 text-right">
           <p class="text-md font-bold text-gray-500 uppercase">*G/C: Generally Conforms</p>
           <p class="text-sm font-bold">
            Showing {{ (store.page - 1) * store.pageCount + 1 }} - {{ Math.min(store.page * store.pageCount, store.filteredReports.length) }} of {{ store.filteredReports.length }} data
          </p>
        </div>
      </div>
    </UCard>
</template>

<script setup lang="ts">
import { useQualityAssuranceStore, QAType } from '~/stores/quality-assurance'

const store = useQualityAssuranceStore()

const formatTypeLabel = (type: string) => {
  if (store.matchQAType(type, QAType.REGULAR)) return 'Regular'
  if (store.matchQAType(type, QAType.SAIV)) return 'SAIV'
  if (store.matchQAType(type, QAType.IACM)) return 'BUMN IACM'
  return 'QAR'
}

const openReportFile = (row: any) => {
  if (row.attachment && row.attachment.filePath) {
    const baseUrl = store.getMasterServiceBaseUrl()
    window.open(`${baseUrl}/quality-assurance/${row.id}/download`, '_blank')
  }
}

const formatOverallConclusion = (result: string) => {
  if (!result) return '-'
  const res = result.trim().toLowerCase()
  if (res === 'g/c*' || res === 'gc' || res.includes('generally')) {
    return 'Generally Conformed'
  }
  if (res === 'fc' || res.includes('fully') || res.includes('conformance')) {
    return 'Fully Conformance'
  }
  if (res === 'pc' || res.includes('partially')) {
    return 'Partially Conform'
  }
  if (res === 'dnc' || res.includes('does not') || res.includes('doesnot')) {
    return 'Does not Conform'
  }
  return result
}
</script>