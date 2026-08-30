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
        <UBadge color="warning" variant="subtle">{{ row.original.requiredDate }}</UBadge>
      </template>
      <template #file-cell="{ row }">
        <UButton v-if="row.original.file" icon="i-heroicons-document-arrow-down" color="neutral" variant="ghost" size="sm">
          {{ row.original.file.name }}
        </UButton>
        <span v-else class="text-gray-400 text-sm">-</span>
      </template>
      <template #actions-cell="{ row }">
        <div class="flex items-center gap-1">
          <UButton icon="i-lucide-edit" color="warning" variant="ghost" size="md" @click="store.editDocument(row.original)" />
          <UButton icon="i-lucide-trash-2" color="error" variant="ghost" size="md" @click="store.deleteDocument(row.index)" />
        </div>
      </template>
    </TableEntities>

    <!-- Document Modal -->
    <UModal v-model:open="store.showDocumentModal" scrollable :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' }">
      <template #content>
        <div class="relative flex flex-col max-h-[90vh]">
          <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
            <h3 class="text-lg font-bold text-[var(--text-main)]">{{ store.isEditingDocument ? t('auditFieldwork.document.modalEdit') : t('auditFieldwork.document.modalAdd') }}</h3>
            <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="() => { store.showDocumentModal = false }" />
          </div>

          <div class="p-6 overflow-y-auto space-y-5">
            <UForm @submit.prevent="store.saveDocument()" class="space-y-4">
            <UFormField :label="t('auditFieldwork.document.name')" required>
              <UInput v-model="store.documentForm.documentName" :placeholder="t('auditFieldwork.document.namePlaceholder')" class="w-full" required />
            </UFormField>

            <UFormField :label="t('auditFieldwork.document.description')" required>
              <UTextarea v-model="store.documentForm.description" :placeholder="t('auditFieldwork.document.descriptionPlaceholder')" class="w-full" required />
            </UFormField>

            <UFormField :label="t('auditFieldwork.document.requiredDate')" required>
              <UInput v-model="store.documentForm.requiredDate" type="date" class="w-full" required />
            </UFormField>
          
            <UFormField :label="t('auditFieldwork.document.uploadFile')">
              <UInput
                type="file"
                icon="i-heroicons-paper-clip"
                @change="store.handleDocumentFileChange"
                accept=".pdf,.docx,.doc,.xlsx,.xls"
                class="w-full"
              />
              <div v-if="store.documentForm.file" class="mt-2 flex items-center gap-2">
                <UIcon name="i-heroicons-document" />
                <span class="font-bold text-sm">{{ store.documentForm.file.name }}</span>
              </div>
            </UFormField>
            </UForm>
          </div>

          <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-end gap-2">
            <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="() => { store.showDocumentModal = false }" />
            <UButton color="primary" :label="store.isEditingDocument ? t('common.edit') : t('common.submit')" @click="store.saveDocument()" />
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import TableEntities from '~/components/shared/TableEntities.vue'

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
