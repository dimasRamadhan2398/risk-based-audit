<template>
  <div class="space-y-4">
    <!-- Header with Add Button -->
    <div class="flex justify-between items-center p-4">
      <div>
        <h2 class="text-lg font-semibold">{{ t('auditFieldwork.interview.title') }}</h2>
        <p class="text-sm text-gray-500">{{ t('auditFieldwork.interview.subtitle') }}</p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" label="Add Interview" @click="store.openInterviewModal()" />
    </div>

    <!-- Interview List via TableEntities -->
    <TableEntities
      :data="store.interviews"
      :columns="columns"
      :empty-state="{
        icon: 'i-heroicons-microphone',
        label: t('auditFieldwork.interview.empty')
      }"
      class="w-full"
    >
      <template #interviewee-cell="{ row }">
        <div>
          <p class="font-semibold text-[var(--text-main)]">{{ row.original.interviewee }}</p>
          <p class="text-xs text-[var(--text-muted)]">{{ row.original.intervieweePosition }}</p>
        </div>
      </template>
      <template #interviewer-cell="{ row }">
        <div>
          <p class="font-semibold text-[var(--text-main)]">{{ row.original.interviewer }}</p>
          <p class="text-xs text-[var(--text-muted)]">{{ row.original.interviewerPosition }}</p>
        </div>
      </template>
      <template #date-cell="{ row }">
        <span class="text-sm text-[var(--text-main)]">{{ row.original.date }}</span>
      </template>
      <template #topic-cell="{ row }">
        <UBadge color="primary" variant="subtle" class="font-medium">{{ row.original.topic }}</UBadge>
      </template>
      <template #file-cell="{ row }">
        <UButton
          v-if="row.original.fileName || row.original.file"
          icon="i-heroicons-document-arrow-down"
          color="neutral"
          variant="ghost"
          size="sm"
          @click="store.downloadFile(row.original.fileName)"
        >
          {{ row.original.fileName ? row.original.fileName.split('-').slice(1).join('-') : (row.original.file ? row.original.file.name : '') }}
        </UButton>
        <span v-else class="text-gray-400 text-sm">-</span>
      </template>
      <template #actions-cell="{ row }">
        <div class="flex items-center gap-1">
          <UTooltip text="Edit Interview">
            <UButton
              icon="i-lucide-edit"
              color="warning"
              variant="ghost"
              size="md"
              @click="store.editInterview(row.original)"
            />
          </UTooltip>
          <UTooltip text="Delete Interview">
            <UButton
              icon="i-lucide-trash-2"
              color="error"
              variant="ghost"
              size="md"
              @click="store.deleteInterview(row.index)"
            />
          </UTooltip>
        </div>
      </template>
    </TableEntities>

    <!-- Interview Modal -->
    <UModal 
      v-model:open="store.showInterviewModal"
      :ui="{
        content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
        body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
        footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0 bg-white dark:bg-gray-900',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
    >
      <template #content>
        <div class="relative flex flex-col max-h-[90vh] transition-colors duration-300">
          <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ store.isEditingInterview ? t('auditFieldwork.interview.modalEdit') : t('auditFieldwork.interview.modalAdd') }}
            </h3>
            <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="() => { store.showInterviewModal = false; }" />
          </div>

          <div class="p-6 overflow-y-auto space-y-5">
            <UForm @submit.prevent="store.saveInterview()" class="space-y-4">
              <!-- Interviewee (Auditee) -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField :label="t('auditFieldwork.interview.interviewee')" required>
                  <UInput
                    v-model="store.interviewForm.interviewee"
                    :placeholder="t('auditFieldwork.interview.intervieweePlaceholder')"
                    class="w-full"
                    required
                    maxlength="100"
                    @invalid="($event.target as any)?.setCustomValidity('Nama auditee maksimal 100 karakter dan wajib diisi')"
                    @input="($event.target as any)?.setCustomValidity('')"
                  />
                  <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ store.interviewForm.interviewee ? store.interviewForm.interviewee.length : 0 }}/100
                  </div>
                </UFormField>
                <UFormField :label="t('auditFieldwork.interview.intervieweePosition')" required>
                  <USelectMenu 
                    v-model="store.interviewForm.intervieweePosition" 
                    :items="store.options.positions" 
                    :placeholder="t('auditFieldwork.interview.intervieweePositionPlaceholder')" 
                    required 
                    class="w-full" 
                  />
                </UFormField>
              </div>

              <!-- Interviewer (Auditor) -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField :label="t('auditFieldwork.interview.interviewer')" required>
                  <USelectMenu
                    v-model="store.interviewForm.interviewer"
                    :items="store.memberOptions"
                    value-key="value"
                    label-key="label"
                    :placeholder="t('auditFieldwork.interview.interviewerPlaceholder')"
                    :disabled="!store.hasSelectedAssignmentLetter"
                    class="w-full"
                    @update:model-value="(val: any) => {
                      const found = store.memberOptions.find((m: any) => m.value === val || m.label === val)
                      if (found && found.role && !store.interviewForm.interviewerPosition) {
                        store.interviewForm.interviewerPosition = found.role
                      }
                    }"
                    required
                  >
                    <template #item="{ item }">
                      <div class="flex items-center justify-between w-full gap-2">
                        <span class="font-medium text-sm">{{ item.label }}</span>
                        <UBadge v-if="item.role" color="primary" variant="subtle" size="sm">{{ item.role }}</UBadge>
                      </div>
                    </template>
                  </USelectMenu>
                </UFormField>
                <UFormField :label="t('auditFieldwork.interview.interviewerPosition')" required>
                  <USelectMenu 
                    v-model="store.interviewForm.interviewerPosition" 
                    :items="['Chairperson', 'Supervisor', 'Member', 'Person in Charge', ...store.options.positions.filter(p => !['Chairperson', 'Supervisor', 'Member', 'Person in Charge'].includes(p))]" 
                    :placeholder="t('auditFieldwork.interview.interviewerPositionPlaceholder')" 
                    required 
                    class="w-full" 
                  />
                </UFormField>
              </div>

              <!-- Date and Topic -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField :label="t('auditFieldwork.interview.date')" required>
                  <UInput v-model="store.interviewForm.date" type="date" required class="w-full"/>
                </UFormField>
                <UFormField :label="t('auditFieldwork.interview.topic')" required>
                  <USelectMenu 
                    v-model="store.interviewForm.topic" 
                    :items="store.options.auditTopics" 
                    :placeholder="t('auditFieldwork.interview.topicPlaceholder')" 
                    required 
                    class="w-full" 
                  />
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
                <div v-if="store.interviewForm.file || store.interviewForm.fileName" class="mt-2 flex items-center gap-2">
                  <UIcon name="i-heroicons-document" />
                  <span class="font-bold text-sm">{{ store.interviewForm.file ? store.interviewForm.file.name : (store.interviewForm.fileName ? store.interviewForm.fileName.split('-').slice(1).join('-') : '') }}</span>
                </div>
              </UFormField>
            </UForm>
          </div>

          <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-end gap-2">
            <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="() => { store.showInterviewModal = false; }" />
            <UButton color="primary" :label="store.isEditingInterview ? t('common.edit') : t('common.submit')" @click="store.saveInterview()" />
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
  { accessorKey: 'interviewee', header: t('auditFieldwork.interview.columns.interviewee') },
  { accessorKey: 'interviewer', header: t('auditFieldwork.interview.columns.interviewer') },
  { accessorKey: 'date', header: t('auditFieldwork.interview.columns.date') },
  { accessorKey: 'topic', header: t('auditFieldwork.interview.columns.topic') },
  { accessorKey: 'file', header: t('auditFieldwork.interview.columns.file') },
  { accessorKey: 'actions', header: t('auditFieldwork.interview.columns.actions') }
])
</script>
