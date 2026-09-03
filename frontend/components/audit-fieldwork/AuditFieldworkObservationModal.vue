<template>
  <!-- Observation Modal (View / Add / Edit) -->
  <UModal 
    v-model:open="store.showObservationModal"
    :ui="{ content: 'sm:max-w-2xl w-full bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' }"
  >
    <template #content>
      <div class="relative flex flex-col max-h-[90vh] transition-colors duration-300">
        <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
          <h3 class="text-lg font-bold text-[var(--text-main)]">
            {{ store.isReadOnlyObservation ? (t('auditFieldwork.observation.modalView') || 'Detail Observasi') : (store.isEditingObservation ? t('auditFieldwork.observation.modalEdit') : t('auditFieldwork.observation.modalAdd')) }}
          </h3>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="store.showObservationModal = false" />
        </div>

        <div class="p-6 overflow-y-auto space-y-5">
          <UForm @submit.prevent="handleSaveObservation()" class="space-y-4">
            <!-- Activity Textarea similar to StrategicPlanForm -->
            <UFormField :label="t('auditFieldwork.observation.activity')" required :error="errors.activity">
              <UTextarea
                v-model="store.observationForm.activity"
                :placeholder="t('auditFieldwork.observation.activityPlaceholder')"
                :rows="2"
                :disabled="store.isReadOnlyObservation"
                required
                class="w-full"
              />
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField :label="t('auditFieldwork.observation.location')" required :error="errors.location">
                <UInput v-model="store.observationForm.location" :placeholder="t('auditFieldwork.observation.locationPlaceholder')" :disabled="store.isReadOnlyObservation" required class="w-full" />
              </UFormField>
              <UFormField :label="t('auditFieldwork.observation.date')" required :error="errors.date">
                <AppDatePicker v-model="store.observationForm.date" :disabled="store.isReadOnlyObservation" required class="w-full" />
              </UFormField>
            </div>

            <UFormField :label="t('auditFieldwork.observation.observer')" required :error="errors.observer">
              <USelectMenu
                v-model="store.observationForm.observer"
                :items="store.memberOptions"
                value-key="value"
                label-key="label"
                :placeholder="t('auditFieldwork.observation.observerPlaceholder')"
                :disabled="store.isReadOnlyObservation || !store.hasSelectedAssignmentLetter"
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

            <UFormField v-if="!store.isReadOnlyObservation || store.observationForm.file || store.observationForm.fileName" :label="t('auditFieldwork.observation.uploadFile')">
              <UInput
                v-if="!store.isReadOnlyObservation"
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
