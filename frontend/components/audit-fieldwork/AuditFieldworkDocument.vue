<template>
  <div class="space-y-4">
    <!-- Header with Add Button -->
    <div class="flex justify-between items-center">
      <div>
        <h3 class="text-lg font-semibold">{{ t('auditFieldwork.document.title') }}</h3>
        <p class="text-sm text-gray-500">{{ t('auditFieldwork.document.subtitle') }}</p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" :label="t('auditFieldwork.document.addBtn')" @click="store.openDocumentModal()" />
    </div>

    <!-- Document List via TableEntities -->
    <TableEntities
      :data="store.documents"
      :columns="columns"
      :empty-state="{
        icon: 'i-heroicons-document-duplicate',
        label: t('auditFieldwork.document.empty')
      }"
      :ui="{ td: '!whitespace-normal' }"
      class="w-full"
    >
      <template #documentName-cell="{ row }">
        <span class="font-medium text-gray-900 dark:text-white">{{ row.original.documentName }}</span>
      </template>
      <template #description-cell="{ row }">
        <div class="w-full min-w-0 whitespace-normal break-words leading-relaxed text-sm text-gray-600 dark:text-gray-300">
          {{ row.original.description || '-' }}
        </div>
      </template>
      <template #requiredDate-cell="{ row }">
        <UBadge color="warning" variant="subtle" size="md" class="font-semibold">
          <UIcon name="i-heroicons-calendar" class="w-3.5 h-3.5 mr-1" />
          {{ formatDate(row.original.requiredDate) }}
        </UBadge>
      </template>
      <template #file-cell="{ row }">
        <UButton
          v-if="row.original.file || row.original.fileName"
          icon="i-heroicons-document-arrow-down"
          color="neutral"
          variant="ghost"
          size="sm"
          @click="store.downloadDocumentFile(row.original)"
        >
          {{ row.original.file?.name || row.original.fileName }}
        </UButton>
        <span v-else class="text-gray-400 text-sm">-</span>
      </template>
      <template #actions-cell="{ row }">
        <div class="flex items-center gap-1 justify-center">
          <UTooltip :text="t('common.actions.view') || 'Lihat'">
            <UButton 
              icon="i-heroicons-eye" 
              color="neutral" 
              variant="ghost" 
              size="sm" 
              @click="store.viewDocument(row.original)" />
          </UTooltip>
          <UTooltip :text="t('common.actions.edit') || 'Ubah'">
            <UButton 
              icon="i-heroicons-pencil-square" 
              color="primary" 
              variant="ghost" 
              size="sm" 
              @click="store.editDocument(row.original)" />
          </UTooltip>
          <UTooltip :text="t('common.actions.delete') || 'Hapus'">
            <UButton 
              icon="i-heroicons-trash" 
              color="error" 
              variant="ghost" 
              size="sm" 
              @click="store.deleteDocument(row.index)" />
          </UTooltip>
        </div>
      </template>
    </TableEntities>

    <!-- Document Modal -->
    <AuditFieldworkDocumentModal />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import TableEntities from '~/components/shared/TableEntities.vue'
import { formatDate } from '~/utils/dateConverter'
import AuditFieldworkDocumentModal from '~/components/audit-fieldwork/AuditFieldworkDocumentModal.vue'

const store = useAuditFieldworkStore()
const { t } = useI18n()

const columns = computed(() => [
  { key: 'documentName', accessorKey: 'documentName', header: t('auditFieldwork.document.columns.name'), class: 'w-56 min-w-[180px]' },
  { key: 'description', accessorKey: 'description', header: t('auditFieldwork.document.columns.description'), class: 'max-w-[300px] whitespace-normal break-words' },
  { key: 'requiredDate', accessorKey: 'requiredDate', header: t('auditFieldwork.document.columns.requiredDate'), class: 'w-40 min-w-[130px] whitespace-nowrap' },
  { key: 'file', accessorKey: 'file', header: t('auditFieldwork.document.columns.file'), class: 'w-48 min-w-[100px]' },
  { key: 'actions', accessorKey: 'actions', header: t('auditFieldwork.document.columns.actions'), class: 'w-24 min-w-[90px] whitespace-nowrap text-center' }
])
</script>
