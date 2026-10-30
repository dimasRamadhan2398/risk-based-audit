<template>
    <UCard class="overflow-hidden border border-gray-200 dark:border-gray-800" :ui="{ body: 'p-0' }">
      <TableEntities
        :columns="store.columns"
        :data="store.filteredReports"
        class="w-full"
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
            {{ formatResult(row.original) }}
          </span>
        </template>

        <template #status-cell="{ row }: { row: any }">
          <div class="flex items-center space-x-2">
            <div :class="['w-4 h-4 rounded-full', store.getStatusColor(row.original.status)]"></div>
            <span class="text-sm font-medium">{{ formatStatus(row.original.status) }}</span>
          </div>
        </template>

        <template #conductedBy-cell="{ row }">
          <span class="font-medium text-gray-700 dark:text-white">{{ row.original.conductedBy || '-' }}</span>
        </template>

        <template #viewReport-cell="{ row }">
          <UButton
            v-if="row.original.attachment?.filePath && canAccessFiles"
            icon="i-lucide-file-text"
            :label="t('qualityAssurance.dashboard.table.view')"
            color="neutral"
            variant="subtle"
            size="sm"
            class="font-bold border border-gray-200 dark:border-gray-700"
            @click="openReportFile(row.original)"
          />
          <span v-else-if="!row.original.attachment?.filePath" class="text-md text-gray-400 font-medium italic">{{ t('qualityAssurance.dashboard.table.noFile') }}</span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-1">
            <UTooltip :text="t('qualityAssurance.dashboard.table.viewDetail')">
            <UButton
              icon="i-lucide-eye"
              color="neutral"
              variant="ghost"
              size="md"
              @click="store.openDetail(row.original)"
            />
            </UTooltip>
            <UTooltip v-if="canWriteQaConsulting" :text="t('qualityAssurance.dashboard.table.editReport')">
            <UButton
              icon="i-lucide-edit"
              color="warning"
              variant="ghost"
              size="md"
              @click="store.editReport(row.original)"
            />
            </UTooltip>
            <UTooltip v-if="canWriteQaConsulting" :text="t('qualityAssurance.dashboard.table.deleteReport')">
            <UButton
              icon="i-lucide-trash-2"
              color="error"
              variant="ghost"
              size="md"
              @click="store.deleteReport(row.original)"
            />
            </UTooltip>
          </div>
        </template>
      </TableEntities>
    </UCard>
</template>

<script setup lang="ts">
import { useRbac } from '~/composables/useRbac'
import { useQualityAssuranceStore, QAStatus, QAType } from '~/stores/quality-assurance'

const store = useQualityAssuranceStore()
const { t } = useI18n()
const { canWriteQaConsulting, canAccessFiles } = useRbac()

const formatStatus = (status: string) => {
  if (status === QAStatus.COMPLETED) return t('qualityAssurance.statuses.completed')
  if (status === QAStatus.VERIFIED) return t('qualityAssurance.statuses.verified')
  if (status === QAStatus.IN_PROGRESS) return t('qualityAssurance.statuses.inProgress')
  if (status === QAStatus.PLANNED) return t('qualityAssurance.statuses.planned')
  return status
}

const formatTypeLabel = (type: string) => {
  if (!type) return '-'
  if (store.matchQAType(type, QAType.REGULAR)) return t('qualityAssurance.types.regular')
  if (store.matchQAType(type, QAType.SAIV)) return 'SAIV'
  if (store.matchQAType(type, QAType.IACM)) return 'BUMN IACM'
  if (store.matchQAType(type, QAType.QAR)) return 'QAR'
  return type
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
    return t('qualityAssurance.conformance.generallyConformed')
  }
  if (res === 'fc' || res.includes('fully') || res.includes('conformance')) {
    return t('qualityAssurance.conformance.fullyConformance')
  }
  if (res === 'pc' || res.includes('partially')) {
    return t('qualityAssurance.conformance.partiallyConform')
  }
  if (res === 'dnc' || res.includes('does not') || res.includes('doesnot')) {
    return t('qualityAssurance.conformance.doesNotConform')
  }
  return result
}

const formatIacmResult = (result?: string) => {
  if (!result || result === '-') return '-'
  const trimmed = result.trim()
  if (trimmed.includes('/')) return trimmed
  return `${trimmed} / 5`
}

const formatResult = (row: any) => {
  if (!row || !row.result) return '-'
  if (store.matchQAType(row.type, QAType.QAR) || store.matchQAType(row.type, QAType.SAIV)) {
    return formatOverallConclusion(row.result)
  }
  if (store.matchQAType(row.type, QAType.IACM)) {
    return formatIacmResult(row.result)
  }
  return row.result
}
</script>
