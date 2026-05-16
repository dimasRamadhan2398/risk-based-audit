<template>
  <div class="space-y-4">
    <!-- Header with Add Button -->
    <div class="flex justify-between items-center">
      <div>
        <h3 class="text-lg font-semibold">Document Collections</h3>
        <p class="text-sm text-gray-500">Manage documents required for audit implementation</p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" label="Add Document" @click="store.openDocumentModal()" />
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
      <p class="text-gray-500">No documents collected yet</p>
      <UButton color="primary" variant="soft" class="mt-2" label="Add Document" @click="store.openDocumentModal()" />
    </div>

    <!-- Document Modal -->
    <Teleport to="body">
      <div v-if="store.showDocumentModal" class="fixed inset-0 bg-gray-900/60 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">{{ store.isEditingDocument ? 'Edit Document' : 'Document Collection Form' }}</h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="store.showDocumentModal = false" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveDocument()" class="space-y-4">
            <UFormField label="Document Name" required>
              <UInput v-model="store.documentForm.documentName" placeholder="Example: SOP Procurement" class="w-full" required />
            </UFormField>

            <UFormField label="Document Description" required>
              <UTextarea v-model="store.documentForm.description" placeholder="Describe the purpose and scope of this document" class="w-full" required />
            </UFormField>

            <UFormField label="Required Date" required>
              <UInput v-model="store.documentForm.requiredDate" type="date" class="w-full" required />
            </UFormField>
          
            <UFormField label="Upload File (PDF/DOCX/XLSX)">
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
              <UButton color="neutral" variant="soft" label="Cancel" @click="store.showDocumentModal = false" />
              <UButton color="primary" :label="store.isEditingDocument ? 'Update' : 'Submit'" @click="store.saveDocument()" />
            </div>
          </template>
        </UCard>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'

const store = useAuditFieldworkStore()

const columns = [
  { accessorKey: 'documentName', header: 'Document Name' },
  { accessorKey: 'description', header: 'Description' },
  { accessorKey: 'requiredDate', header: 'Required Date' },
  { accessorKey: 'file', header: 'File' },
  { accessorKey: 'actions', header: 'Actions' }
]
</script>
