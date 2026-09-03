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
      class="w-full"
    >
      <template #documentName-cell="{ row }">
        <span class="font-medium">{{ row.original.documentName }}</span>
      </template>
      <template #description-cell="{ row }">
        <span class="text-sm text-gray-600">{{ row.original.description }}</span>
      </template>
      <template #requiredDate-cell="{ row }">
        <UBadge color="warning" variant="subtle">{{ formatDate(row.original.requiredDate) }}</UBadge>
      </template>
      <template #file-cell="{ row }">
        <UButton v-if="row.original.file" icon="i-heroicons-document-arrow-down" color="neutral" variant="ghost" size="sm">
          {{ row.original.file.name }}
        </UButton>
        <span v-else class="text-gray-400 text-sm">-</span>
      </template>
      <template #actions-cell="{ row }">
        <div class="flex items-center gap-1">
          <UButton icon="i-heroicons-pencil-square" color="primary" variant="ghost" size="sm" @click="store.editDocument(row.original)" />
          <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteDocument(row.index)" />
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
  { accessorKey: 'documentName', header: t('auditFieldwork.document.columns.name') },
  { accessorKey: 'description', header: t('auditFieldwork.document.columns.description') },
  { accessorKey: 'requiredDate', header: t('auditFieldwork.document.columns.requiredDate') },
  { accessorKey: 'file', header: t('auditFieldwork.document.columns.file') },
  { accessorKey: 'actions', header: t('auditFieldwork.document.columns.actions') }
])
</script>
