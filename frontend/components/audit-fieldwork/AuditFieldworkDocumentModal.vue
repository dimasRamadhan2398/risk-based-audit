<template>
  <!-- Document Modal (View / Add / Edit) -->
  <UModal 
    v-model:open="store.showDocumentModal"
    :ui="{
      content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
      body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
      footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
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
              :class="store.isReadOnlyDocument ? 'bg-blue-500/10 text-blue-500' : 'bg-primary-500/10 text-primary-500'"
            >
              <UIcon
                :name="store.isReadOnlyDocument ? 'i-heroicons-eye' : (store.isEditingDocument ? 'i-heroicons-pencil-square' : 'i-heroicons-plus-circle')"
                class="w-5 h-5"
              />
            </div>
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ store.isReadOnlyDocument ? (t('auditFieldwork.document.modalView') || 'Detail Pengumpulan Dokumen') : (store.isEditingDocument ? t('auditFieldwork.document.modalEdit') : t('auditFieldwork.document.modalAdd')) }}
            </h3>
          </div>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="store.showDocumentModal = false" />
        </div>

        <!-- Read-Only Detail View -->
        <div v-if="store.isReadOnlyDocument" class="p-6 overflow-y-auto space-y-4">
          <!-- Assignment Letter & Required Date Header Card -->
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
              <UBadge color="warning" variant="subtle" size="md" class="font-semibold">
                <UIcon name="i-heroicons-calendar" class="w-4 h-4 mr-1.5" />
                {{ formatDate(store.documentForm.requiredDate) || '-' }}
              </UBadge>
            </div>
          </UCard>

          <!-- Document Name Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                <UIcon name="i-heroicons-document-duplicate" class="w-4 h-4 text-primary-500" />
                <span>{{ t('auditFieldwork.document.name') }}</span>
              </div>
            </template>
            <p class="text-base font-bold text-[var(--text-main)]">{{ store.documentForm.documentName || '-' }}</p>
          </UCard>

          <!-- Description Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                <UIcon name="i-heroicons-information-circle" class="w-4 h-4 text-primary-500" />
                <span>{{ t('auditFieldwork.document.description') }}</span>
              </div>
            </template>
            <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
              {{ store.documentForm.description || '-' }}
            </p>
          </UCard>

          <!-- File Attachment Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                <UIcon name="i-heroicons-paper-clip" class="w-4 h-4 text-primary-500" />
                <span>{{ t('auditFieldwork.document.columns.file') }}</span>
              </div>
            </template>
            <div v-if="store.documentForm.file || store.documentForm.fileName" class="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-main)] border border-[var(--border-main)]">
              <div class="flex items-center gap-2.5 min-w-0">
                <UIcon name="i-heroicons-document-text" class="w-5 h-5 text-primary-500 shrink-0" />
                <span class="text-sm font-semibold text-[var(--text-main)] truncate">
                  {{ store.documentForm.file?.name || store.documentForm.fileName }}
                </span>
              </div>
              <UButton
                icon="i-heroicons-document-arrow-down"
                color="primary"
                variant="solid"
                size="xs"
                label="Download"
                @click="store.downloadDocumentFile(store.documentForm)"
              />
            </div>
            <p v-else class="text-sm text-[var(--text-muted)] italic">Tidak ada berkas terlampir</p>
          </UCard>
        </div>

        <!-- Add / Edit Form View -->
        <div v-else class="p-6 overflow-y-auto space-y-5">
          <UForm @submit.prevent="handleSaveDocument()" class="space-y-4">
            <UFormField :label="t('auditFieldwork.document.name')" required :error="errors.documentName">
              <UInput
                v-model="store.documentForm.documentName"
                :placeholder="t('auditFieldwork.document.namePlaceholder')"
                class="w-full"
                required
              />
            </UFormField>

            <UFormField :label="t('auditFieldwork.document.description')" required :error="errors.description">
              <UTextarea
                v-model="store.documentForm.description"
                :placeholder="t('auditFieldwork.document.descriptionPlaceholder')"
                :rows="3"
                class="w-full"
                required
              />
            </UFormField>

            <UFormField :label="t('auditFieldwork.document.requiredDate')" required :error="errors.requiredDate">
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
              <div v-if="store.documentForm.file || store.documentForm.fileName" class="mt-2 flex items-center gap-2">
                <UIcon name="i-heroicons-document" />
                <span class="font-bold text-sm">{{ store.documentForm.file?.name || store.documentForm.fileName }}</span>
              </div>
            </UFormField>
          </UForm>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-end gap-2">
          <template v-if="store.isReadOnlyDocument">
<<<<<<< HEAD
            <UButton color="neutral" variant="soft" :label="t('common.close') || 'Tutup'" @click="store.showDocumentModal = false" />
=======
            <UButton color="neutral" variant="soft" :label="t('common.close') || 'Tutup'" @click="() => { store.showDocumentModal = false; }" />
>>>>>>> upstream/main
            <UButton
              color="primary"
              icon="i-heroicons-pencil-square"
              :label="t('common.edit') || 'Ubah'"
              @click="switchToEditMode()"
            />
          </template>
          <template v-else>
            <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="store.showDocumentModal = false" />
            <UButton
              color="primary"
              :label="store.isEditingDocument ? t('common.edit') : t('common.submit')"
              :disabled="!isFormValid || store.loading"
              :loading="store.loading"
              @click="handleSaveDocument()"
            />
          </template>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import { formatDate } from '~/utils/dateConverter'

const store = useAuditFieldworkStore()
const { t } = useI18n()

const errors = reactive({
  documentName: '',
  description: '',
  requiredDate: ''
})

const isFormValid = computed(() => {
  return Boolean(
    store.documentForm.documentName?.trim() &&
    store.documentForm.description?.trim() &&
    store.documentForm.requiredDate?.trim()
  )
})

const validateForm = () => {
  let valid = true
  errors.documentName = ''
  errors.description = ''
  errors.requiredDate = ''

  if (!store.documentForm.documentName?.trim()) {
    errors.documentName = t('auditFieldwork.document.validation.nameRequired') || 'Nama dokumen wajib diisi'
    valid = false
  }
  if (!store.documentForm.description?.trim()) {
    errors.description = t('auditFieldwork.document.validation.descriptionRequired') || 'Deskripsi dokumen wajib diisi'
    valid = false
  }
  if (!store.documentForm.requiredDate?.trim()) {
    errors.requiredDate = t('auditFieldwork.document.validation.requiredDateRequired') || 'Tanggal dibutuhkan wajib diisi'
    valid = false
  }

  return valid
}

const handleSaveDocument = async () => {
  if (store.isReadOnlyDocument) return
  if (!validateForm()) {
    return
  }
  await store.saveDocument()
}

watch(() => store.documentForm.documentName, (val) => {
  if (val?.trim()) errors.documentName = ''
})
watch(() => store.documentForm.description, (val) => {
  if (val?.trim()) errors.description = ''
})
watch(() => store.documentForm.requiredDate, (val) => {
  if (val?.trim()) errors.requiredDate = ''
})
watch(() => store.showDocumentModal, (isOpen) => {
  if (!isOpen) {
    errors.documentName = ''
    errors.description = ''
    errors.requiredDate = ''
  }
})

const switchToEditMode = () => {
  store.isReadOnlyDocument = false
  store.isEditingDocument = true
}
</script>
