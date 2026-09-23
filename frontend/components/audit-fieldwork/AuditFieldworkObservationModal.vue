<template>
  <!-- Observation Modal (View / Add / Edit) -->
  <UModal 
    v-model:open="store.showObservationModal"
    :ui="{ content: 'sm:max-w-2xl w-full bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' }"
  >
    <template #content>
      <div class="relative flex flex-col max-h-[90vh] transition-colors duration-300">
        <!-- Modal Header -->
        <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
          <div class="flex items-center gap-2">
            <div
              class="p-2 rounded-lg"
              :class="store.isReadOnlyObservation ? 'bg-blue-500/10 text-blue-500' : 'bg-primary-500/10 text-primary-500'"
            >
              <UIcon
                :name="store.isReadOnlyObservation ? 'i-heroicons-eye' : (store.isEditingObservation ? 'i-heroicons-pencil-square' : 'i-heroicons-plus-circle')"
                class="w-5 h-5"
              />
            </div>
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ store.isReadOnlyObservation ? (t('auditFieldwork.observation.modalView') || 'Detail Observasi') : (store.isEditingObservation ? t('auditFieldwork.observation.modalEdit') : t('auditFieldwork.observation.modalAdd')) }}
            </h3>
          </div>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="store.showObservationModal = false" />
        </div>

        <!-- Read-Only Detail View (using UCard for each section) -->
        <div v-if="store.isReadOnlyObservation" class="p-6 overflow-y-auto space-y-4">
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
                {{ formatDate(store.observationForm.date) || '-' }}
              </UBadge>
            </div>
          </UCard>

          <!-- Observer & Location Cards -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Location -->
            <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
              <template #header>
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                  <UIcon name="i-heroicons-map-pin" class="w-4 h-4 text-primary-500" />
                  <span>{{ t('auditFieldwork.observation.location') }}</span>
                </div>
              </template>
              <p class="text-base font-bold text-[var(--text-main)]">{{ store.observationForm.location || '-' }}</p>
            </UCard>

            <!-- Observer -->
            <UCard color="primary" variant="outline" class="border border-primary-500/20 shadow-xs">
              <template #header>
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                  <UIcon name="i-heroicons-user" class="w-4 h-4 text-primary-500" />
                  <span>{{ t('auditFieldwork.observation.observer') }}</span>
                </div>
              </template>
              <p class="text-base font-bold text-[var(--text-main)]">{{ store.observationForm.observer || '-' }}</p>
            </UCard>
          </div>

          <!-- Activity Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                <UIcon name="i-heroicons-clipboard-document-list" class="w-4 h-4 text-primary-500" />
                <span>{{ t('auditFieldwork.observation.activity') }}</span>
              </div>
            </template>
            <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
              {{ store.observationForm.activity || '-' }}
            </p>
          </UCard>

          <!-- File Attachment Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                <UIcon name="i-heroicons-paper-clip" class="w-4 h-4 text-primary-500" />
                <span>{{ t('auditFieldwork.observation.columns.file') }}</span>
              </div>
            </template>
            <div v-if="store.observationForm.file || store.observationForm.fileName" class="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-main)] border border-[var(--border-main)]">
              <div class="flex items-center gap-2.5 min-w-0">
                <UIcon name="i-heroicons-document-text" class="w-5 h-5 text-primary-500 shrink-0" />
                <span class="text-sm font-semibold text-[var(--text-main)] truncate">
                  {{ store.observationForm.file?.name || store.observationForm.fileName }}
                </span>
              </div>
              <div class="flex items-center gap-1.5 shrink-0">
                <UButton
                  icon="i-heroicons-eye"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  label="Lihat"
                  @click="store.previewInterviewFile(store.observationForm)"
                />
                <UButton
                  icon="i-heroicons-document-arrow-down"
                  color="primary"
                  variant="solid"
                  size="xs"
                  label="Download"
                  @click="store.downloadInterviewFile(store.observationForm)"
                />
              </div>
            </div>
            <p v-else class="text-sm text-[var(--text-muted)] italic">Tidak ada berkas terlampir</p>
          </UCard>
        </div>

        <!-- Add / Edit Form View -->
        <div v-else class="p-6 overflow-y-auto space-y-5">
          <UForm @submit.prevent="handleSaveObservation()" class="space-y-4">
            <!-- Activity Textarea similar to StrategicPlanForm -->
            <UFormField :label="t('auditFieldwork.observation.activity')" required :error="errors.activity">
              <UTextarea
                v-model="store.observationForm.activity"
                :placeholder="t('auditFieldwork.observation.activityPlaceholder')"
                :rows="2"
                required
                class="w-full"
              />
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField :label="t('auditFieldwork.observation.location')" required :error="errors.location">
                <UInput v-model="store.observationForm.location" :placeholder="t('auditFieldwork.observation.locationPlaceholder')" required class="w-full" />
              </UFormField>
              <UFormField :label="t('auditFieldwork.observation.date')" required :error="errors.date">
                <AppDatePicker v-model="store.observationForm.date" required class="w-full" />
              </UFormField>
            </div>

            <UFormField :label="t('auditFieldwork.observation.observer')" required :error="errors.observer">
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
              <div v-if="store.observationForm.file || store.observationForm.fileName" class="mt-2 flex items-center gap-2">
                <UIcon name="i-heroicons-document" />
                <span class="font-bold text-sm">{{ store.observationForm.file?.name || store.observationForm.fileName }}</span>
              </div>
            </UFormField>
          </UForm>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-end gap-2">
          <template v-if="store.isReadOnlyObservation">
            <UButton color="neutral" variant="soft" :label="t('common.close') || 'Tutup'" @click="store.showObservationModal = false" />
            <UButton
              color="primary"
              icon="i-heroicons-pencil-square"
              :label="t('common.edit') || 'Ubah'"
              @click="switchToEditMode()"
            />
          </template>
          <template v-else>
            <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="store.showObservationModal = false" />
            <UButton
              color="primary"
              :label="store.isEditingObservation ? t('common.edit') : t('common.submit')"
              :disabled="!isFormValid || store.loading"
              :loading="store.loading"
              @click="handleSaveObservation()"
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
  activity: '',
  location: '',
  date: '',
  observer: ''
})

const isFormValid = computed(() => {
  return Boolean(
    store.observationForm.activity?.trim() &&
    store.observationForm.location?.trim() &&
    store.observationForm.date?.trim() &&
    store.observationForm.observer?.trim()
  )
})

const validateForm = () => {
  let valid = true
  errors.activity = ''
  errors.location = ''
  errors.date = ''
  errors.observer = ''

  if (!store.observationForm.activity?.trim()) {
    errors.activity = t('auditFieldwork.observation.validation.activityRequired') || 'Aktivitas yang diobservasi wajib diisi'
    valid = false
  }
  if (!store.observationForm.location?.trim()) {
    errors.location = t('auditFieldwork.observation.validation.locationRequired') || 'Lokasi observasi wajib diisi'
    valid = false
  }
  if (!store.observationForm.date?.trim()) {
    errors.date = t('auditFieldwork.observation.validation.dateRequired') || 'Tanggal observasi wajib diisi'
    valid = false
  }
  if (!store.observationForm.observer?.trim()) {
    errors.observer = t('auditFieldwork.observation.validation.observerRequired') || 'Petugas observasi wajib dipilih'
    valid = false
  }

  return valid
}

const handleSaveObservation = async () => {
  if (store.isReadOnlyObservation) return
  if (!validateForm()) {
    return
  }
  await store.saveObservation()
}

watch(() => store.observationForm.activity, (val) => {
  if (val?.trim()) errors.activity = ''
})
watch(() => store.observationForm.location, (val) => {
  if (val?.trim()) errors.location = ''
})
watch(() => store.observationForm.date, (val) => {
  if (val?.trim()) errors.date = ''
})
watch(() => store.observationForm.observer, (val) => {
  if (val?.trim()) errors.observer = ''
})
watch(() => store.showObservationModal, (isOpen) => {
  if (!isOpen) {
    errors.activity = ''
    errors.location = ''
    errors.date = ''
    errors.observer = ''
  }
})

const switchToEditMode = () => {
  store.isReadOnlyObservation = false
  store.isEditingObservation = true
}
</script>
