<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-2">{{ t('auditFieldwork.title') }}</h1>
    <p class="text-gray-500 mb-6">{{ t('auditFieldwork.subtitle') }}</p>

    <!-- Assignment Letter Selector -->
    <UCard class="mb-6" :ui="{ body: 'p-4' }">
      <div class="flex flex-col md:flex-row md:items-center gap-4">
        <div class="flex-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            {{ t('auditFieldwork.assignmentLetter.label') }}
          </label>
          <USelectMenu
            v-model="store.selectedAssignmentLetter"
            :items="store.publishedAssignmentLetters"
            :placeholder="t('auditFieldwork.assignmentLetter.placeholder')"
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
          {{ t('auditFieldwork.assignmentLetter.noPublished') }}
        </div>
        <div v-else-if="store.selectedAssignmentLetter" class="text-sm text-green-600">
          <UIcon name="i-heroicons-check-circle" class="size-4 inline mr-1" />
          {{ t('auditFieldwork.assignmentLetter.selectedAudit', { letter: store.selectedAssignmentLetter }) }}
        </div>
      </div>
    </UCard>

    <!-- Main Content - Only show when Assignment Letter is selected -->
    <div v-if="store.hasSelectedAssignmentLetter">
      <UTabs :items="tabs" class="w-full">
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
                <h2 class="text-lg font-semibold text-gray-900">{{ t('auditFieldwork.workingPapers.title') }}</h2>
                <p class="text-sm text-gray-500">{{ t('auditFieldwork.workingPapers.subtitle', { letter: store.selectedAssignmentLetter }) }}</p>
              </div>
            </div>

            <div v-if="filteredWorkingPapers.length > 0" class="mx-4">
              <UCard :ui="{ body: 'p-0', root: 'ring-1 ring-gray-200 shadow-sm' }">
                <UTable :data="filteredWorkingPapers" :columns="workingPaperColumns">
                  <template #actions-cell="{ row }">
                    <div class="flex gap-2">
                      <UButton
                        icon="i-lucide-eye"
                        size="md"
                        color="neutral"
                        variant="ghost"
                        :to="`/working-paper?id=${row.original.id}`"
                        :title="t('auditFieldwork.workingPapers.tooltips.view')"
                      />
                      <UButton
                        icon="i-lucide-edit"
                        size="md"
                        color="warning"
                        variant="ghost"
                        :to="`/working-paper?id=${row.original.id}&action=edit`"
                        :title="t('auditFieldwork.workingPapers.tooltips.edit')"
                      />
                      <UButton
                        icon="i-lucide-trash-2"
                        size="md"
                        color="error"
                        variant="ghost"
                        @click="wpStore.deleteF01(row.original.id!)"
                        :title="t('auditFieldwork.workingPapers.tooltips.delete')"
                      />
                    </div>
                  </template>
                </UTable>
              </UCard>
            </div>

            <div v-else class="text-center py-12 rounded-xl border-2 border-dashed border-gray-200 mx-4">
              <UIcon name="i-heroicons-document-text" class="size-16 text-gray-300 mx-auto mb-4" />
              <h3 class="text-lg font-semibold text-gray-700">{{ t('auditFieldwork.workingPapers.emptyTitle') }}</h3>
              <p class="text-gray-500 mt-2 max-w-md mx-auto mb-6">
                {{ t('auditFieldwork.workingPapers.emptyDesc', { letter: store.selectedAssignmentLetter }) }}
              </p>
              <UButton 
                color="primary" 
                icon="i-heroicons-plus" 
                :label="t('auditFieldwork.workingPapers.createFirst')" 
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
      <h3 class="text-lg font-semibold text-gray-700">{{ t('auditFieldwork.emptyState.title') }}</h3>
      <p class="text-gray-500 mt-2 max-w-md mx-auto">
        {{ t('auditFieldwork.emptyState.desc') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useWorkingPaperStore } from '~/stores/working-paper'
import { useI18n } from '~/composables/useI18n'
import { computed, watchEffect } from 'vue'

const store = useAuditFieldworkStore()
const wpStore = useWorkingPaperStore()
const { t } = useI18n()

watchEffect(() => {
  if (store.selectedAssignmentLetter) {
    store.fetchAllFieldworkData(store.selectedAssignmentLetter)
  }
})

const tabs = computed(() => [
  { label: t('auditFieldwork.tabs.interview'), slot: 'tab01', icon: 'i-heroicons-microphone' },
  { label: t('auditFieldwork.tabs.observation'), slot: 'tab02', icon: 'i-heroicons-eye' },
  { label: t('auditFieldwork.tabs.document'), slot: 'tab03', icon: 'i-heroicons-document-duplicate' },
  { label: t('auditFieldwork.tabs.sample'), slot: 'tab04', icon: 'i-heroicons-table-cells' },
  { label: t('auditFieldwork.tabs.testControls'), slot: 'tab05', icon: 'i-heroicons-shield-check' },
  { label: t('auditFieldwork.tabs.workingPapers'), slot: 'tab06', icon: 'i-heroicons-document-text' }
])

const filteredWorkingPapers = computed(() => {
  if (!store.selectedAssignmentLetter) return []
  return wpStore.dataF01.filter(wp => wp.assignmentLetterId === store.selectedAssignmentLetter)
})

const workingPaperColumns = computed(() => [
  { accessorKey: 'businessProcess', header: t('auditFieldwork.workingPapers.columns.businessProcess') },
  { accessorKey: 'period', header: t('auditFieldwork.workingPapers.columns.period') },
  { accessorKey: 'riskLevel', header: t('auditFieldwork.workingPapers.columns.riskLevel') },
  { accessorKey: 'rootCause', header: t('auditFieldwork.workingPapers.columns.rootCause') },
  { accessorKey: 'location', header: t('auditFieldwork.workingPapers.columns.location') },
  { accessorKey: 'actions', header: t('auditFieldwork.workingPapers.columns.actions') }
])
</script>
