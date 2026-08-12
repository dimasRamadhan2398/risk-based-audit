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

    <!-- Document List -->
    <UCard v-if="store.documents.length > 0" :ui="{ body: 'p-4' }">
      <UTable :data="store.documents" :columns="columns">
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
          <div class="flex items-center">
            <UButton icon="i-heroicons-pencil-square" color="primary" variant="ghost" size="sm" @click="store.editDocument(row.original)" />
            <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteDocument(row.index)" />
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- Empty State -->
    <div v-else class="text-center py-8">
      <UIcon name="i-heroicons-document-duplicate" class="size-12 text-gray-300 mx-auto mb-2" />
      <p class="text-gray-500">{{ t('auditFieldwork.document.empty') }}</p>
      <UButton color="primary" variant="soft" class="mt-2" :label="t('auditFieldwork.document.addBtn')" @click="store.openDocumentModal()" />
    </div>

    <!-- Document Modal -->
    <Teleport to="body">
      <div v-if="store.showDocumentModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">{{ store.isEditingDocument ? t('auditFieldwork.document.modalEdit') : t('auditFieldwork.document.modalAdd') }}</h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="() => { store.showDocumentModal = false }" />
            </div>
          </template>

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

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="() => { store.showDocumentModal = false }" />
              <UButton color="primary" :label="store.isEditingDocument ? t('common.edit') : t('common.submit')" @click="store.saveDocument()" />
            </div>
          </template>
        </UCard>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import { computed } from 'vue'

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
