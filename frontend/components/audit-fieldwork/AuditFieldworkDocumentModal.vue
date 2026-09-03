<template>
  <!-- Document Modal (Add / Edit) -->
  <UModal 
    v-model:open="store.showDocumentModal"
    :ui="{ content: 'sm:max-w-2xl w-full bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' }"
  >
    <template #content>
      <div class="relative flex flex-col max-h-[90vh] transition-colors duration-300">
        <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
          <h3 class="text-lg font-bold text-[var(--text-main)]">
            {{ store.isEditingDocument ? t('auditFieldwork.document.modalEdit') : t('auditFieldwork.document.modalAdd') }}
          </h3>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="store.showDocumentModal = false" />
        </div>

        <div class="p-6 overflow-y-auto space-y-5">
          <UForm @submit.prevent="store.saveDocument()" class="space-y-4">
            <UFormField :label="t('auditFieldwork.document.name')" required>
              <UInput v-model="store.documentForm.documentName" :placeholder="t('auditFieldwork.document.namePlaceholder')" class="w-full" required />
            </UFormField>

            <UFormField :label="t('auditFieldwork.document.description')" required>
              <UTextarea v-model="store.documentForm.description" :placeholder="t('auditFieldwork.document.descriptionPlaceholder')" class="w-full" required />
            </UFormField>

            <UFormField :label="t('auditFieldwork.document.requiredDate')" required>
              <AppDatePicker v-model="store.documentForm.requiredDate" class="w-full" required />
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
        </div>

        <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-end gap-2">
          <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="store.showDocumentModal = false" />
          <UButton color="primary" :label="store.isEditingDocument ? t('common.edit') : t('common.submit')" @click="store.saveDocument()" />
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'

const store = useAuditFieldworkStore()
const { t } = useI18n()
</script>
