<template>
  <div class="space-y-4">
    <!-- Header with Add Button -->
    <div class="flex justify-between items-center p-4">
      <div>
        <p class="text-sm text-gray-500">{{ t('auditFieldwork.interview.subtitle') }}</p>
        <h2 class="text-lg font-semibold">{{ t('auditFieldwork.interview.title') }}</h2>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" label="Add Interview" @click="store.openInterviewModal()" />
    </div>

    <!-- Interview List via TableEntities -->
    <TableEntities
      :data="store.interviews"
      :columns="columns"
      :empty-state="{
        icon: 'i-heroicons-microphone',
        label: t('auditFieldwork.interview.empty')
      }"
      class="w-full"
    >
      <template #interviewee-cell="{ row }">
        <div>
          <p class="font-semibold text-[var(--text-main)]">{{ row.original.interviewee }}</p>
          <p class="text-xs text-[var(--text-muted)]">{{ row.original.intervieweePosition }}</p>
        </div>
      </template>
      <template #interviewer-cell="{ row }">
        <div>
          <p class="font-semibold text-[var(--text-main)]">{{ row.original.interviewer }}</p>
          <p class="text-xs text-[var(--text-muted)]">{{ row.original.interviewerPosition }}</p>
        </div>
      </template>
      <template #date-cell="{ row }">
        <span class="text-sm text-[var(--text-main)]">{{ formatDate(row.original.date) }}</span>
      </template>
      <template #topic-cell="{ row }">
        <div
          class="line-clamp-2 overflow-hidden max-w-xs"
          :title="row.original.topic"
        >
          <UBadge color="primary" variant="subtle" class="font-medium whitespace-normal break-words">
            {{ row.original.topic }}
          </UBadge>
        </div>
      </template>
      <template #file-cell="{ row }">
        <UButton
          v-if="row.original.file || row.original.fileName || row.original.fileUrl || row.original.filePath"
          icon="i-heroicons-document-arrow-down"
          color="neutral"
          variant="ghost"
          size="sm"
          @click="store.downloadInterviewFile(row.original)"
        >
          {{ row.original.fileName || row.original.file?.name }}
        </UButton>
        <span v-else class="text-gray-400 text-sm">-</span>
      </template>
      <template #actions-cell="{ row }">
        <div class="flex items-center justify-center gap-1">
          <UButton
            icon="i-heroicons-eye"
            color="neutral"
            variant="ghost"
            size="sm"
            :title="t('common.actions.view') || 'Detail'"
            @click="store.viewInterview(row.original)"
          />
          <UButton
            icon="i-lucide-edit"
            color="warning"
            variant="ghost"
            size="sm"
            :title="t('common.actions.edit') || 'Ubah'"
            @click="store.editInterview(row.original)"
          />
          <UButton
            icon="i-lucide-trash-2"
            color="error"
            variant="ghost"
            size="sm"
            :title="t('common.actions.delete') || 'Hapus'"
            @click="store.deleteInterview(row.index)"
          />
        </div>
      </template>
    </TableEntities>

    <!-- Interview Modal (View / Add / Edit) -->
    <AuditFieldworkInterviewModal />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import TableEntities from '~/components/shared/TableEntities.vue'
import { formatDate } from '~/utils/dateConverter'
import AuditFieldworkInterviewModal from '~/components/audit-fieldwork/AuditFieldworkInterviewModal.vue'

const store = useAuditFieldworkStore()
const { t } = useI18n()

const columns = computed(() => [
  { accessorKey: 'interviewee', header: t('auditFieldwork.interview.columns.interviewee'), class: 'w-52 min-w-[180px]' },
  { accessorKey: 'interviewer', header: t('auditFieldwork.interview.columns.interviewer'), class: 'w-52 min-w-[180px]' },
  { accessorKey: 'date', header: t('auditFieldwork.interview.columns.date'), class: 'w-36 min-w-[120px] whitespace-nowrap' },
  { accessorKey: 'topic', header: t('auditFieldwork.interview.columns.topic'), class: 'w-64 min-w-[180px] max-w-xs' },
  { accessorKey: 'file', header: t('auditFieldwork.interview.columns.file'), class: 'w-48 min-w-[100px]' },
  { accessorKey: 'actions', header: t('auditFieldwork.interview.columns.actions'), class: 'w-24 min-w-[90px] whitespace-nowrap text-center' }
])
</script>
