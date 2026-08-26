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
          <span v-else-if="row.original.fileName" class="text-gray-600 text-sm font-semibold">{{ row.original.fileName }}</span>
          <span v-else class="text-gray-400 text-sm">-</span>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex items-center gap-1">
            <UButton icon="i-heroicons-pencil-square" color="primary" variant="ghost" size="sm" @click="store.editObservation(row.original)" />
            <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteObservation(row.index)" />
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- Empty State -->
    <div v-else class="text-center py-8">
      <UIcon name="i-heroicons-eye" class="size-12 text-gray-300 mx-auto mb-2" />
      <p class="text-gray-500">{{ t('auditFieldwork.observation.empty') }}</p>
      <UButton color="primary" variant="soft" class="mt-2" :label="t('auditFieldwork.observation.addBtn')" @click="store.openObservationModal()" />
    </div>

    <!-- Observation Modal -->
    <UModal 
      v-model:open="store.showObservationModal"
      :ui="{ content: 'sm:max-w-2xl w-full bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' }"
    >
      <template #content>
        <div class="relative flex flex-col max-h-[90vh] transition-colors duration-300">
          <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ store.isEditingObservation ? t('auditFieldwork.observation.modalEdit') : t('auditFieldwork.observation.modalAdd') }}
            </h3>
            <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="store.showObservationModal = false" />
          </div>

          <div class="p-6 overflow-y-auto space-y-5">
            <UForm @submit.prevent="store.saveObservation()" class="space-y-4">
              <UFormField :label="t('auditFieldwork.observation.activity')" required>
                <UInput v-model="store.observationForm.activity" :placeholder="t('auditFieldwork.observation.activityPlaceholder')" required class="w-full" />
              </UFormField>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField :label="t('auditFieldwork.observation.location')" required>
                  <UInput v-model="store.observationForm.location" :placeholder="t('auditFieldwork.observation.locationPlaceholder')" required class="w-full" />
                </UFormField>
                <UFormField :label="t('auditFieldwork.observation.date')" required>
                  <UInput v-model="store.observationForm.date" type="date" required class="w-full" />
                </UFormField>
              </div>

              <UFormField :label="t('auditFieldwork.observation.observer')" required>
                <USelectMenu
                  v-model="store.observationForm.observer"
                  :items="store.memberOptions"
                  value-key="value"
                  label-key="label"
                  :placeholder="t('auditFieldwork.observation.observerPlaceholder')"
                  :disabled="!store.hasSelectedAssignmentLetter"
                  class="w-full"
                >
                  <template #item="{ item }">
                    <div class="flex items-center justify-between w-full gap-2">
                      <span class="font-medium text-sm">{{ item.label }}</span>
                      <UBadge v-if="item.role" color="primary" variant="subtle" size="sm">{{ item.role }}</UBadge>
                    </div>
                  </template>
                </USelectMenu>
              </UFormField>

              <UFormField :label="t('auditFieldwork.observation.uploadFile')">
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
          </div>

          <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-end gap-2">
            <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="store.showObservationModal = false" />
            <UButton color="primary" :label="store.isEditingObservation ? t('common.edit') : t('common.submit')" @click="store.saveObservation()" />
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import { computed } from 'vue'

const store = useAuditFieldworkStore()
const { t } = useI18n()

const columns = computed(() => [
  { accessorKey: 'activity', header: t('auditFieldwork.observation.columns.activity') },
  { accessorKey: 'location', header: t('auditFieldwork.observation.columns.location') },
  { accessorKey: 'date', header: t('auditFieldwork.observation.columns.date') },
  { accessorKey: 'observer', header: t('auditFieldwork.observation.columns.observer') },
  { accessorKey: 'file', header: t('auditFieldwork.observation.columns.file') },
  { accessorKey: 'actions', header: t('auditFieldwork.observation.columns.actions') }
])
</script>
