<template>
  <div class="space-y-4">
    <!-- Header with Add Button -->
    <div class="flex justify-between items-center p-4">
      <div>
        <h2 class="text-lg font-semibold">{{ t('auditFieldwork.observation.title') }}</h2>
        <p class="text-sm text-gray-500">{{ t('auditFieldwork.observation.subtitle') }}</p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" :label="t('auditFieldwork.observation.addBtn')" @click="store.openObservationModal()" />
    </div>

    <!-- Observation List via TableEntities -->
    <TableEntities
      :data="store.observations"
      :columns="columns"
      :empty-state="{
        icon: 'i-heroicons-eye',
        label: t('auditFieldwork.observation.empty')
      }"
      :ui="{ td: '!whitespace-normal' }"
      class="w-full"
    >
      <!-- Activity Cell with Full Text Wrapping -->
      <template #activity-cell="{ row }">
        <div
          class="w-full min-w-0 whitespace-normal break-words leading-relaxed font-medium text-gray-900 dark:text-white"
          style="white-space: normal !important; word-break: break-word !important; overflow-wrap: anywhere !important;"
        >
          {{ row.original.activity || '-' }}
        </div>
      </template>
      <template #location-cell="{ row }">
        <UBadge color="neutral" variant="subtle">{{ row.original.location }}</UBadge>
      </template>
      <template #date-cell="{ row }">
        <span class="text-sm text-[var(--text-main)]">{{ formatDate(row.original.date) }}</span>
      </template>
      <template #observer-cell="{ row }">
        <span>{{ row.original.observer }}</span>
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
          <UButton icon="i-heroicons-eye" color="neutral" variant="ghost" size="sm" :title="t('common.actions.view') || 'Lihat'" @click="store.viewObservation(row.original)" />
          <UButton icon="i-heroicons-pencil-square" color="primary" variant="ghost" size="sm" :title="t('common.actions.edit') || 'Ubah'" @click="store.editObservation(row.original)" />
          <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" :title="t('common.actions.delete') || 'Hapus'" @click="store.deleteObservation(row.index)" />
        </div>
      </template>
    </TableEntities>

    <!-- Observation Modal (View / Add / Edit) -->
    <AuditFieldworkObservationModal />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import TableEntities from '~/components/shared/TableEntities.vue'
import { formatDate } from '~/utils/dateConverter'
import AuditFieldworkObservationModal from '~/components/audit-fieldwork/AuditFieldworkObservationModal.vue'

const store = useAuditFieldworkStore()
const { t } = useI18n()

const columns = computed(() => [
  {
    key: 'activity',
    accessorKey: 'activity',
    header: t('auditFieldwork.observation.columns.activity'),
    class: 'w-80 min-w-[240px] max-w-sm !whitespace-normal break-words',
    tdClass: '!whitespace-normal break-words'
  },
  { key: 'location', accessorKey: 'location', header: t('auditFieldwork.observation.columns.location'), class: 'w-52 min-w-[160px]' },
  { key: 'date', accessorKey: 'date', header: t('auditFieldwork.observation.columns.date'), class: 'w-36 min-w-[120px] whitespace-nowrap' },
  { key: 'observer', accessorKey: 'observer', header: t('auditFieldwork.observation.columns.observer'), class: 'w-52 min-w-[160px]' },
  { key: 'file', accessorKey: 'file', header: t('auditFieldwork.observation.columns.file'), class: 'w-48 min-w-[100px]' },
  { key: 'actions', accessorKey: 'actions', header: t('auditFieldwork.observation.columns.actions'), class: 'w-24 min-w-[90px] whitespace-nowrap text-center' }
])
</script>
