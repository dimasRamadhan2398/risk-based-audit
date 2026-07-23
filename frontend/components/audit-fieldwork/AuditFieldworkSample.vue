<template>
  <div class="space-y-4">
    <!-- Header with Add Button -->
    <div class="flex justify-between items-center">
      <div>
        <h3 class="text-lg font-semibold">Sample Data</h3>
        <p class="text-sm text-gray-500">Manage sample data for audit testing</p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" label="Add Sample" @click="store.openSampleModal()" />
    </div>

    <!-- Sample List -->
    <UCard v-if="store.samples.length > 0" :ui="{ body: 'p-4' }">
      <UTable :data="store.samples" :columns="columns">
        <template #documentName-cell="{ row }">
          <span class="font-medium">{{ row.original.documentName }}</span>
        </template>
        <template #documentNumber-cell="{ row }">
          <UBadge color="neutral" variant="subtle">{{ row.original.documentNumber }}</UBadge>
        </template>
        <template #date-cell="{ row }">
          <span>{{ row.original.date }}</span>
        </template>
        <template #description-cell="{ row }">
          <span class="text-sm text-gray-600">{{ row.original.description }}</span>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex items-center">
            <UButton icon="i-heroicons-pencil-square" color="primary" variant="ghost" size="sm" @click="store.editSample(row.original)" />
            <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteSample(row.index)" />
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- Empty State -->
    <div v-else class="text-center py-8">
      <UIcon name="i-heroicons-table-cells" class="size-12 text-gray-300 mx-auto mb-2" />
      <p class="text-gray-500">No samples data yet</p>
      <UButton color="primary" variant="soft" class="mt-2" label="Add Sample" @click="store.openSampleModal()" />
    </div>

    <!-- Sample Modal -->
    <Teleport to="body">
      <div v-if="store.showSampleModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">{{ store.isEditingSample ? 'Edit Sample Data' : 'Sample Data Form' }}</h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="() => { store.showSampleModal = false }" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveSample()" class="space-y-4">
            <UFormField label="Document Name" required>
              <UInput v-model="store.sampleForm.documentName" placeholder="Example: Purchase Invoice" class="w-full" required />
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField label="Document Number" required>
                <UInput v-model="store.sampleForm.documentNumber" placeholder="Example: INV/2026/001" class="w-full"required />
              </UFormField>
              <UFormField label="Date" required>
                <UInput v-model="store.sampleForm.date" type="date" class="w-full" required />
              </UFormField>
            </div>

            <UFormField label="Description" required>
              <UTextarea v-model="store.sampleForm.description" placeholder="Describe the content or details of this sample" class="w-full" required />
            </UFormField>
          </UForm>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" label="Cancel" @click="() => {store.showSampleModal = false}" />
              <UButton color="primary" :label="store.isEditingSample ? 'Update' : 'Submit'" @click="store.saveSample()" />
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
  { accessorKey: 'documentNumber', header: 'Document Number' },
  { accessorKey: 'date', header: 'Date' },
  { accessorKey: 'description', header: 'Description' },
  { accessorKey: 'actions', header: 'Actions' }
]
</script>
