<template>
  <div class="space-y-4">
    <!-- Header with Add Button -->
    <div class="flex justify-between items-center p-4">
      <div>
        <h2 class="text-lg font-semibold">{{ t('auditFieldwork.interview.title') }}</h2>
        <p class="text-sm text-gray-500">{{ t('auditFieldwork.interview.subtitle') }}</p>
      </div>
    </div>

    <!-- Interview List -->
    <UCard v-if="store.interviews.length > 0" :ui="{ body: 'p-4' }">
      <UTable :data="store.interviews" :columns="columns">
        <template #interviewee-cell="{ row }">
          <div>
            <p class="font-medium">{{ row.original.interviewee }}</p>
            <p class="text-md text-gray-500">{{ row.original.intervieweePosition }}</p>
          </div>
        </template>
        <template #interviewer-cell="{ row }">
          <div>
            <p class="font-medium">{{ row.original.interviewer }}</p>
            <p class="text-md text-gray-500">{{ row.original.interviewerPosition }}</p>
          </div>
        </template>
        <template #topic-cell="{ row }">
          <UBadge color="primary" variant="subtle">{{ row.original.topic }}</UBadge>
        </template>
        <template #file-cell="{ row }">
          <UButton v-if="row.original.file" icon="i-heroicons-document-arrow-down" color="neutral" variant="ghost" size="sm">
            {{ row.original.file.name }}
          </UButton>
          <span v-else class="text-gray-400 text-sm">-</span>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex items-center">
            <UButton icon="i-heroicons-pencil-square" color="primary" variant="ghost" size="sm" @click="store.editInterview(row.original)" />
            <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteInterview(row.index)" />
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- Empty State -->
    <div v-else class="text-center py-8">
      <UIcon name="i-heroicons-microphone" class="size-12 text-gray-300 mx-auto mb-2" />
      <p class="text-gray-500">{{ t('auditFieldwork.interview.empty') }}</p>
      <UButton color="primary" variant="soft" class="mt-2" :label="t('auditFieldwork.interview.addBtn')" @click="store.openInterviewModal()" />
    </div>

    <!-- Interview Modal -->
    <Teleport to="body">
      <div v-if="store.showInterviewModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">{{ store.isEditingInterview ? t('auditFieldwork.interview.modalEdit') : t('auditFieldwork.interview.modalAdd') }}</h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="() => { store.showInterviewModal = false }" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveInterview()" class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField :label="t('auditFieldwork.interview.interviewee')" required>
                <USelectMenu
                  v-model="store.interviewForm.interviewee"
                  :items="store.memberOptions"
                  value-key="value"
                  :placeholder="t('auditFieldwork.interview.intervieweePlaceholder')"
                  :disabled="!store.hasSelectedAssignmentLetter"
                  class="w-full"
                  required
                />
              </UFormField>
              <UFormField :label="t('auditFieldwork.interview.intervieweePosition')" required>
                <USelectMenu v-model="store.interviewForm.intervieweePosition" :items="store.options.positions" :placeholder="t('auditFieldwork.interview.intervieweePositionPlaceholder')" required class="w-full" />
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField :label="t('auditFieldwork.interview.interviewer')" required>
                <USelectMenu
                  v-model="store.interviewForm.interviewer"
                  :items="store.memberOptions"
                  value-key="value"
                  :placeholder="t('auditFieldwork.interview.interviewerPlaceholder')"
                  :disabled="!store.hasSelectedAssignmentLetter"
                  class="w-full"
                  required
                />
              </UFormField>
              <UFormField :label="t('auditFieldwork.interview.interviewerPosition')" required>
                <USelectMenu v-model="store.interviewForm.interviewerPosition" :items="store.options.positions" :placeholder="t('auditFieldwork.interview.interviewerPositionPlaceholder')" required class="w-full" />
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField :label="t('auditFieldwork.interview.date')" required>
                <UInput v-model="store.interviewForm.date" type="date" required class="w-full"/>
              </UFormField>
              <UFormField :label="t('auditFieldwork.interview.topic')" required>
                <USelectMenu v-model="store.interviewForm.topic" :items="store.options.auditTopics" :placeholder="t('auditFieldwork.interview.topicPlaceholder')" required class="w-full" />
              </UFormField>
            </div>

            <UFormField :label="t('auditFieldwork.interview.uploadFile')">
              <UInput
                type="file"
                icon="i-heroicons-paper-clip"
                @change="store.handleInterviewFileChange"
                accept=".pdf,.docx,.doc"
                class="w-full"
              />
              <div v-if="store.interviewForm.file" class="mt-2 flex items-center gap-2">
                <UIcon name="i-heroicons-document" />
                <span class="font-bold text-sm">{{ store.interviewForm.file.name }}</span>
              </div>
            </UFormField>
          </UForm>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="() => { store.showInterviewModal = false }" />
              <UButton color="primary" :label="store.isEditingInterview ? t('common.edit') : t('common.submit')" @click="store.saveInterview()" />
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
  { accessorKey: 'interviewee', header: t('auditFieldwork.interview.columns.interviewee') },
  { accessorKey: 'intervieweePosition', header: t('auditFieldwork.interview.columns.intervieweePosition') },
  { accessorKey: 'interviewer', header: t('auditFieldwork.interview.columns.interviewer') },
  { accessorKey: 'interviewerPosition', header: t('auditFieldwork.interview.columns.interviewerPosition') },
  { accessorKey: 'date', header: t('auditFieldwork.interview.columns.date') },
  { accessorKey: 'topic', header: t('auditFieldwork.interview.columns.topic') },
  { accessorKey: 'file', header: t('auditFieldwork.interview.columns.file') },
  { accessorKey: 'actions', header: t('auditFieldwork.interview.columns.actions') }
])
</script>
