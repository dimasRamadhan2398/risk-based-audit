<template>
  <!-- Sample Modal (Add / Edit) -->
  <UModal 
    v-model:open="store.showSampleModal"
    :ui="{ content: 'sm:max-w-2xl w-full bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' }"
  >
    <template #content>
      <div class="relative flex flex-col max-h-[90vh] transition-colors duration-300">
        <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
          <h3 class="text-lg font-bold text-[var(--text-main)]">
            {{ store.isEditingSample ? t('auditFieldwork.sample.modalEdit') : t('auditFieldwork.sample.modalAdd') }}
          </h3>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="store.showSampleModal = false" />
        </div>

        <div class="p-6 overflow-y-auto space-y-5">
          <UForm @submit.prevent="store.saveSample()" class="space-y-4">
            <UFormField :label="t('auditFieldwork.sample.name')" required>
              <UInput v-model="store.sampleForm.documentName" :placeholder="t('auditFieldwork.sample.namePlaceholder')" class="w-full" required />
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField :label="t('auditFieldwork.sample.number')" required>
                <UInput v-model="store.sampleForm.documentNumber" :placeholder="t('auditFieldwork.sample.numberPlaceholder')" class="w-full" required />
              </UFormField>
              <UFormField :label="t('auditFieldwork.sample.date')" required>
                <AppDatePicker v-model="store.sampleForm.date" class="w-full" required />
              </UFormField>
            </div>

            <UFormField :label="t('auditFieldwork.sample.description')" required>
              <UTextarea v-model="store.sampleForm.description" :placeholder="t('auditFieldwork.sample.descriptionPlaceholder')" class="w-full" required />
            </UFormField>
          </UForm>
        </div>

        <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-end gap-2">
          <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="store.showSampleModal = false" />
          <UButton color="primary" :label="store.isEditingSample ? t('common.edit') : t('common.submit')" @click="store.saveSample()" />
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
