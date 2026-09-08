<template>
  <!-- Interview Modal (View / Add / Edit) -->
  <UModal 
    v-model:open="store.showInterviewModal"
    :dismissible="false"
    :ui="{ content: 'sm:max-w-2xl w-full bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' }"
  >
    <template #content>
      <div class="relative flex flex-col max-h-[90vh] transition-colors duration-300">
        <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
          <div class="flex items-center gap-2">
            <div
              class="p-2 rounded-lg"
              :class="store.isReadOnlyInterview ? 'bg-blue-500/10 text-blue-500' : 'bg-primary-500/10 text-primary-500'"
            >
              <UIcon
                :name="store.isReadOnlyInterview ? 'i-heroicons-eye' : (store.isEditingInterview ? 'i-heroicons-pencil-square' : 'i-heroicons-plus-circle')"
                class="w-5 h-5"
              />
            </div>
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ store.isReadOnlyInterview ? (t('auditFieldwork.interview.modalView') || 'Detail Wawancara') : (store.isEditingInterview ? t('auditFieldwork.interview.modalEdit') : t('auditFieldwork.interview.modalAdd')) }}
            </h3>
          </div>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="() => { store.showInterviewModal = false }" />
        </div>

        <!-- Read-only Detail View -->
        <div v-if="store.isReadOnlyInterview" class="p-6 overflow-y-auto space-y-4">
          <!-- Assignment Letter & Date Header Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="p-2.5 rounded-lg bg-primary-500/10 text-primary-500">
                  <UIcon name="i-heroicons-document-text" class="w-6 h-6 text-primary-500" />
                </div>
                <div>
                  <p class="text-xs font-semibold tracking-wider text-[var(--text-muted)]">Surat Tugas</p>
                  <p class="text-sm font-bold text-[var(--text-main)]">{{ store.selectedAssignmentLetter || '-' }}</p>
                </div>
              </div>
              <UBadge color="primary" variant="subtle" size="md" class="font-semibold">
                <UIcon name="i-heroicons-calendar" class="w-4 h-4 mr-1.5" />
                {{ formatDate(store.interviewForm.date) || '-' }}
              </UBadge>
            </div>
          </UCard>

          <!-- Interviewee & Interviewer UCards -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Interviewee (Auditee) -->
            <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
              <template #header>
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                  <UIcon name="i-heroicons-user" class="w-4 h-4 text-primary-500" />
                  <span>{{ t('auditFieldwork.interview.interviewee') }}</span>
                </div>
              </template>
              <div class="space-y-1">
                <p class="text-base font-bold text-[var(--text-main)]">{{ store.interviewForm.interviewee || '-' }}</p>
                <UBadge color="primary" variant="subtle" size="sm" class="mt-1 font-medium">
                  {{ store.interviewForm.intervieweePosition || '-' }}
                </UBadge>
              </div>
            </UCard>

            <!-- Interviewer (Auditor) -->
            <UCard color="primary" variant="outline" class="border border-primary-500/20 shadow-xs">
              <template #header>
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                  <UIcon name="i-heroicons-user" class="w-4 h-4 text-primary-500" />
                  <span>{{ t('auditFieldwork.interview.interviewer') }}</span>
                </div>
              </template>
              <div class="space-y-1">
                <p class="text-base font-bold text-[var(--text-main)]">{{ store.interviewForm.interviewer || '-' }}</p>
                <UBadge color="primary" variant="subtle" size="sm" class="mt-1 font-medium">
                  {{ store.interviewForm.interviewerPosition || '-' }}
                </UBadge>
              </div>
            </UCard>
          </div>

          <!-- Topic UCard -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                <UIcon name="i-heroicons-chat-bubble-left-right" class="w-4 h-4 text-primary-500" />
                <span>{{ t('auditFieldwork.interview.topic') }}</span>
              </div>
            </template>
            <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
              {{ store.interviewForm.topic || '-' }}
            </p>
          </UCard>

          <!-- File Attachment UCard -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                <UIcon name="i-heroicons-paper-clip" class="w-4 h-4 text-primary-500" />
                <span>{{ t('auditFieldwork.interview.columns.file') }}</span>
              </div>
            </template>
            <div v-if="store.interviewForm.file || store.interviewForm.fileName" class="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-main)] border border-[var(--border-main)]">
              <div class="flex items-center gap-2.5 min-w-0">
                <UIcon name="i-heroicons-document-text" class="w-5 h-5 text-primary-500 shrink-0" />
                <span class="text-sm font-semibold text-[var(--text-main)] truncate">
                  {{ store.interviewForm.file?.name || store.interviewForm.fileName }}
                </span>
              </div>
              <UButton
                icon="i-heroicons-document-arrow-down"
                color="primary"
                variant="solid"
                size="xs"
                label="Download"
                @click="store.downloadInterviewFile(store.interviewForm)"
              />
            </div>
            <p v-else class="text-sm text-[var(--text-muted)] italic">Tidak ada berkas terlampir</p>
          </UCard>
        </div>

        <!-- Edit / Add Form View -->
        <div v-else class="p-6 overflow-y-auto space-y-5">
          <UForm @submit.prevent="store.saveInterview()" class="space-y-4">
            <!-- Interviewee (Auditee) -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField :label="t('auditFieldwork.interview.interviewee')" required>
                <UInput
                  v-model="store.interviewForm.interviewee"
                  :placeholder="t('auditFieldwork.interview.intervieweePlaceholder')"
                  class="w-full"
                  required
                />
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
                <AppDatePicker v-model="store.interviewForm.date" required class="w-full"/>
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
                <span class="font-bold text-sm">{{ store.interviewForm.file?.name || store.interviewForm.fileName }}</span>
              </div>
            </UFormField>
          </UForm>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-end gap-2">
          <template v-if="store.isReadOnlyInterview">
            <UButton color="neutral" variant="soft" :label="t('common.close') || 'Tutup'" @click="store.showInterviewModal = false" />
            <UButton
              color="primary"
              icon="i-heroicons-pencil-square"
              :label="t('common.edit') || 'Ubah'"
              @click="switchToEditMode()"
            />
          </template>
          <template v-else>
            <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="store.showInterviewModal = false" />
            <UButton color="primary" :label="store.isEditingInterview ? t('common.edit') : t('common.submit')" @click="store.saveInterview()" />
          </template>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import { formatDate } from '~/utils/dateConverter'

const store = useAuditFieldworkStore()
const { t } = useI18n()

const switchToEditMode = () => {
  store.isReadOnlyInterview = false
  store.isEditingInterview = true
}
</script>
