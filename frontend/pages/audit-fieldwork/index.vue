<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-2">Audit Fieldwork</h1>
    <p class="text-gray-500 mb-6">Manage audit execution from interview to working papers</p>

    <!-- Assignment Letter Selector -->
    <UCard class="mb-6" :ui="{ body: 'p-4' }">
      <div class="flex flex-col md:flex-row md:items-center gap-4">
        <div class="flex-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Select Assignment Letter (Audit)
          </label>
          <USelectMenu
            v-model="store.selectedAssignmentLetter"
            :items="store.publishedAssignmentLetters"
            placeholder="Select Assignment Letter to audit"
            class="w-full md:w-96"
            :disabled="store.publishedAssignmentLetters.length === 0"
          >
            <template #leading>
              <UIcon name="i-heroicons-document-text" class="size-5" />
            </template>
          </USelectMenu>
        </div>
        <div v-if="store.publishedAssignmentLetters.length === 0" class="text-sm text-amber-600">
          <UIcon name="i-heroicons-exclamation-triangle" class="size-4 inline mr-1" />
          No Assignment Letter with Published status. Please create and publish an Assignment Letter first.
        </div>
        <div v-else-if="store.selectedAssignmentLetter" class="text-sm text-green-600">
          <UIcon name="i-heroicons-check-circle" class="size-4 inline mr-1" />
          Audit: {{ store.selectedAssignmentLetter }}
        </div>
      </div>
    </UCard>

    <!-- Main Content - Only show when Assignment Letter is selected -->
    <div v-if="store.hasSelectedAssignmentLetter">
      <UTabs :items="store.tabs" class="w-full">
        <template #tab01>
          <AuditFieldworkInterview />
        </template>

        <template #tab02>
          <AuditFieldworkObservation />
        </template>

        <template #tab03>
          <AuditFieldworkDocument />
        </template>

        <template #tab04>
          <AuditFieldworkSample />
        </template>

        <template #tab05>
          <AuditFieldworkTestControls />
        </template>

        <template #tab06>
          <div class="space-y-4">
            <div class="flex justify-between items-center p-4">
              <div>
                <h2 class="text-lg font-semibold text-gray-900">Working Papers</h2>
                <p class="text-sm text-gray-500">Working papers associated with {{ store.selectedAssignmentLetter }}</p>
              </div>
            </div>

            <div v-if="filteredWorkingPapers.length > 0" class="mx-4">
              <UCard :ui="{ body: 'p-0', root: 'ring-1 ring-gray-200 shadow-sm' }">
                <UTable :data="filteredWorkingPapers" :columns="workingPaperColumns">
                  <template #actions-cell="{ row }">
                    <div class="flex gap-2">
                      <UButton
                        icon="i-heroicons-eye"
                        size="sm"
                        color="neutral"
                        variant="ghost"
                        :to="`/working-paper?id=${row.original.id}`"
                        title="View Details"
                      />
                      <UButton
                        icon="i-heroicons-pencil-square"
                        size="sm"
                        color="primary"
                        variant="ghost"
                        :to="`/working-paper?id=${row.original.id}&action=edit`"
                        title="Edit"
                      />
                      <UButton
                        icon="i-heroicons-trash"
                        size="sm"
                        color="error"
                        variant="ghost"
                        @click="wpStore.deleteF01(row.original.id!)"
                        title="Delete"
                      />
                    </div>
                  </template>
                </UTable>
              </UCard>
            </div>

            <div v-else class="text-center py-12 bg-gray-50 rounded-xl border-2 border-dashed border-gray-200 mx-4">
              <UIcon name="i-heroicons-document-text" class="size-16 text-gray-300 mx-auto mb-4" />
              <h3 class="text-lg font-semibold text-gray-700">No Working Papers Found</h3>
              <p class="text-gray-500 mt-2 max-w-md mx-auto mb-6">
                There are no working papers created for assignment letter <span class="font-medium text-gray-900">{{ store.selectedAssignmentLetter }}</span> yet.
              </p>
              <UButton 
                color="primary" 
                icon="i-heroicons-plus" 
                label="Create First Working Paper" 
                to="/working-paper?action=create" 
              />
            </div>
          </div>
        </template>
      </UTabs>
    </div>

    <!-- Empty State - When no Assignment Letter selected -->
    <div v-else class="text-center py-16">
      <UIcon name="i-heroicons-folder-open" class="size-20 text-gray-300 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-700">Select Assignment Letter</h3>
      <p class="text-gray-500 mt-2 max-w-md mx-auto">
        Please select an Assignment Letter (Audit) first to start managing fieldwork.
        Only Assignment Letters with "Published" status can be selected.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useWorkingPaperStore } from '~/stores/working-paper'
import { computed } from 'vue'

const store = useAuditFieldworkStore()
const wpStore = useWorkingPaperStore()

const filteredWorkingPapers = computed(() => {
  if (!store.selectedAssignmentLetter) return []
  return wpStore.dataF01.filter(wp => wp.assignmentLetterId === store.selectedAssignmentLetter)
})

const workingPaperColumns = [
  { accessorKey: 'businessProcess', header: 'Business Process' },
  { accessorKey: 'period', header: 'Period' },
  { accessorKey: 'riskLevel', header: 'Risk Level' },
  { accessorKey: 'rootCause', header: 'Root Cause' },
  { accessorKey: 'location', header: 'Location' },
  { accessorKey: 'actions', header: 'Actions' }
]
</script>
