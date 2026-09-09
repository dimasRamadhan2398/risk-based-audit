<template>
  <!-- Sample Modal (View / Add / Edit) -->
  <UModal 
    v-model:open="store.showSampleModal"
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
        <!-- Modal Header -->
        <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
          <div class="flex items-center gap-2">
            <div
              class="p-2 rounded-lg"
              :class="store.isReadOnlySample ? 'bg-blue-500/10 text-blue-500' : 'bg-primary-500/10 text-primary-500'"
            >
              <UIcon
                :name="store.isReadOnlySample ? 'i-heroicons-eye' : (store.isEditingSample ? 'i-heroicons-pencil-square' : 'i-heroicons-plus-circle')"
                class="w-5 h-5"
              />
            </div>
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ store.isReadOnlySample ? (t('auditFieldwork.sample.modalView') || 'Detail Data Sampel') : (store.isEditingSample ? t('auditFieldwork.sample.modalEdit') : t('auditFieldwork.sample.modalAdd')) }}
            </h3>
          </div>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="() => { store.showSampleModal = false; store.isReadOnlySample = false; store.isEditingSample = false; }" />
        </div>

        <!-- Read-Only Detail View -->
        <div v-if="store.isReadOnlySample" class="p-6 overflow-y-auto space-y-4">
          <!-- Header Info Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="p-2.5 rounded-lg bg-primary-500/10 text-primary-500">
                  <UIcon name="i-heroicons-table-cells" class="w-6 h-6 text-primary-500" />
                </div>
                <div>
                  <p class="text-xs font-semibold tracking-wider text-[var(--text-muted)]">Surat Tugas</p>
                  <p class="text-sm font-bold text-[var(--text-main)]">{{ store.selectedAssignmentLetter || '-' }}</p>
                </div>
              </div>
              <UBadge color="primary" variant="subtle" size="md" class="font-semibold">
                <UIcon name="i-heroicons-calendar" class="w-4 h-4 mr-1.5" />
                {{ formatDate(store.sampleForm.date) || '-' }}
              </UBadge>
            </div>
          </UCard>

          <!-- Document Name & Number Cards -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
              <template #header>
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                  <UIcon name="i-heroicons-document-text" class="w-4 h-4 text-primary-500" />
                  <span>{{ t('auditFieldwork.sample.name') }}</span>
                </div>
              </template>
              <p class="text-base font-bold text-[var(--text-main)]">{{ store.sampleForm.documentName || '-' }}</p>
            </UCard>

            <UCard color="primary" variant="outline" class="border border-primary-500/20 shadow-xs">
              <template #header>
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                  <UIcon name="i-heroicons-hashtag" class="w-4 h-4 text-primary-500" />
                  <span>{{ t('auditFieldwork.sample.number') }}</span>
                </div>
              </template>
              <p class="text-base font-bold text-[var(--text-main)]">{{ store.sampleForm.documentNumber || '-' }}</p>
            </UCard>
          </div>

          <!-- Description Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                <UIcon name="i-heroicons-information-circle" class="w-4 h-4 text-primary-500" />
                <span>{{ t('auditFieldwork.sample.description') }}</span>
              </div>
            </template>
            <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
              {{ store.sampleForm.description || '-' }}
            </p>
          </UCard>
        </div>

        <!-- Add / Edit Form View -->
        <div v-else class="p-6 overflow-y-auto space-y-5">
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
              <UTextarea v-model="store.sampleForm.description" :placeholder="t('auditFieldwork.sample.descriptionPlaceholder')" :rows="3" class="w-full" required />
            </UFormField>
          </UForm>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-end gap-2">
          <template v-if="store.isReadOnlySample">
            <UButton color="neutral" variant="soft" :label="t('common.close') || 'Tutup'" @click="() => {store.showSampleModal = false;}" />
            <UButton
              color="primary"
              icon="i-heroicons-pencil-square"
              :label="t('common.edit') || 'Ubah'"
              @click="switchToEditMode()"
            />
          </template>
          <template v-else>
            <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="() => {store.showSampleModal = false;}" />
            <UButton color="primary" :label="store.isEditingSample ? t('common.edit') : t('common.submit')" @click="store.saveSample()" />
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
  store.isReadOnlySample = false
  store.isEditingSample = true
}
</script>
