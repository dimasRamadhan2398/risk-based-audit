<template>
  <div class="space-y-4">
    <!-- Header with Add Button -->
    <div class="flex justify-between items-center">
      <div>
        <h3 class="text-lg font-semibold">Observation Schedule</h3>
        <p class="text-sm text-gray-500">Manage observation schedule in the field</p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" label="Add Observation" @click="store.openObservationModal()" />
    </div>

    <!-- Observation List -->
    <UCard v-if="store.observations.length > 0" :ui="{ body: 'p-4' }">
      <UTable :data="store.observations" :columns="columns">
        <template #activity-cell="{ row }">
          <span class="font-medium">{{ row.original.activity }}</span>
        </template>
        <template #location-cell="{ row }">
          <UBadge color="neutral" variant="subtle">{{ row.original.location }}</UBadge>
        </template>
        <template #observer-cell="{ row }">
          <span>{{ row.original.observer }}</span>
        </template>
        <template #file-cell="{ row }">
          <UButton v-if="row.original.file" icon="i-heroicons-document-arrow-down" color="neutral" variant="ghost" size="sm">
            {{ row.original.file.name }}
          </UButton>
          <span v-else class="text-gray-400 text-sm">-</span>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex items-center">
            <UButton icon="i-heroicons-pencil-square" color="primary" variant="ghost" size="sm" @click="store.editObservation(row.original)" />
            <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteObservation(row.index)" />
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- Empty State -->
    <div v-else class="text-center py-8">
      <UIcon name="i-heroicons-eye" class="size-12 text-gray-300 mx-auto mb-2" />
      <p class="text-gray-500">No observation data yet</p>
      <UButton color="primary" variant="soft" class="mt-2" label="Add Observation" @click="store.openObservationModal()" />
    </div>

    <!-- Observation Modal -->
    <Teleport to="body">
      <div v-if="store.showObservationModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">{{ store.isEditingObservation ? 'Edit Observation' : 'Observation Form' }}</h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="() => { store.showObservationModal = false }" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveObservation()" class="space-y-4">
            <UFormField label="Activity Observed" required>
              <UInput v-model="store.observationForm.activity" placeholder="Example: Procurement Process" required />
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField label="Observation Location" required>
                <UInput v-model="store.observationForm.location" placeholder="Example: Central Warehouse" required />
              </UFormField>
              <UFormField label="Observation Date" required>
                <UInput v-model="store.observationForm.date" type="date" required />
              </UFormField>
            </div>

            <UFormField label="Observer" required>
              <UInput v-model="store.observationForm.observer" placeholder="Full name of observer" required />
            </UFormField>

            <UFormField label="Upload File (PDF/DOCX)">
              <UInput
                type="file"
                icon="i-heroicons-paper-clip"
                @change="store.handleObservationFileChange"
                accept=".pdf,.docx,.doc"
                class="w-full"
              />
              <div v-if="store.observationForm.file" class="mt-2 flex items-center gap-2">
                <UIcon name="i-heroicons-document" />
                <span class="font-bold text-sm">{{ store.observationForm.file.name }}</span>
              </div>
            </UFormField>
          </UForm>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" label="Cancel" @click="() => { store.showObservationModal = false }" />
              <UButton color="primary" :label="store.isEditingObservation ? 'Update' : 'Submit'" @click="store.saveObservation()" />
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
  { accessorKey: 'activity', header: 'Activity' },
  { accessorKey: 'location', header: 'Location' },
  { accessorKey: 'date', header: 'Date' },
  { accessorKey: 'observer', header: 'Observer' },
  { accessorKey: 'file', header: 'File' },
  { accessorKey: 'actions', header: 'Actions' }
]
</script>
