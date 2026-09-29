<template>
  <UModal
    v-model:open="store.showModal"
    :dismissible="false"
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
  >
    <template #content>
      <UForm @submit.prevent="handleSubmit" class="flex flex-col max-h-[90vh] overflow-hidden flex-1 min-h-0">
        <div class="px-4 py-4 sm:px-6 border-b border-gray-100 dark:border-gray-800 flex justify-between items-center shrink-0">
          <h3 class="text-lg leading-6 font-bold" id="modal-title">
            {{ store.isEditing ? t('auditCharter.sopForm.editTitle') : t('auditCharter.sopForm.addTitle') }}
          </h3>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-lucide-x"
            @click="store.closeModal"
          />
        </div>

        <div class="p-4 sm:p-6 space-y-4 overflow-y-auto flex-1 min-h-0">
            <!-- SOP Name -->
            <UFormField
              :label="t('auditCharter.sopForm.name')"
              class="block text-sm font-medium"
              size="lg"
              required
            >
              <UInput
                v-model="store.form.name"
                required
                type="text"
                maxlength="100"
                :placeholder="t('auditCharter.sopForm.namePlaceholder')"
                class="mt-1 block w-full rounded-md"
                @invalid="($event.target as any)?.setCustomValidity(t('auditCharter.sopForm.nameValidation'))"
                @input="($event.target as any)?.setCustomValidity('')"
              />
              <div class="text-xs text-gray-500 mt-1 text-right">
                {{ store.form.name ? store.form.name.length : 0 }}/100
              </div>
            </UFormField>

            <!-- Parent Guideline Selection -->
            <UFormField
              :label="t('auditCharter.sopForm.parentGuideline')"
              class="block text-sm font-medium"
              size="lg"
              required
            >
              <USelect
                v-model="store.form.guideline_id"
                required
                :items="guidelineOptions"
                :placeholder="t('auditCharter.sopForm.parentGuidelinePlaceholder')"
                class="mt-1 block w-full rounded-md"
              />
            </UFormField>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <!-- Status -->
              <UFormField
                :label="t('auditCharter.sopForm.status')"
                class="block text-sm font-medium"
                size="lg"
                required
              >
                <USelect
                  v-model="store.form.status"
                  required
                  :items="statusOptions"
                  class="mt-1 block w-full rounded-md"
                />
              </UFormField>

              <!-- Effective Date -->
              <UFormField
                :label="t('auditCharter.sopForm.effectiveDate')"
                class="block text-sm font-medium"
                size="lg"
                required
              >
                <UInput
                  v-model="store.form.effective_date"
                  required
                  type="month"
                  class="mt-1 block w-full rounded-md"
                />
              </UFormField>
            </div>

            <!-- File Upload -->
            <UFormField
              :label="t('auditCharter.sopForm.fileUpload')"
              class="block text-sm font-medium"
              size="lg"
              :required="!store.isEditing"
            >
              <div class="mt-1 flex items-center gap-4">
                <input
                  type="file"
                  accept="application/pdf"
                  @change="store.handleFileChange"
                  class="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-primary-50 file:text-primary-700 hover:file:bg-primary-100"
                  :required="!store.isEditing && !store.form.fileUrl"
                />
              </div>
              <p v-if="store.form.fileName" class="text-md text-gray-500 mt-1">
                {{ t('auditCharter.sopForm.selectedFile') }} <span class="font-semibold text-gray-700">{{ store.form.fileName }}</span>
              </p>
            </UFormField>

            <!-- Error message display -->
            <div v-if="store.errorMsg" class="p-3 bg-error-50 text-error-700 rounded-lg text-sm font-semibold">
              {{ store.errorMsg }}
            </div>
          </div>
        

        <div class="border-t border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-900 px-4 py-3 sm:px-6 flex flex-col-reverse sm:flex-row-reverse gap-3 rounded-b-2xl shrink-0">
          <UButton
            type="submit"
            :loading="store.loading"
            color="primary"
            variant="solid"
            size="md"
            class="w-full sm:w-auto font-bold"
          >
            {{ store.isEditing ? t('auditCharter.sopForm.saveChanges') : t('auditCharter.sopForm.addSop') }}
          </UButton>
          <UButton
            type="button"
            color="neutral"
            variant="outline"
            size="md"
            class="w-full sm:w-auto font-bold"
            @click="store.closeModal"
          >
            {{ t('common.cancel') }}
          </UButton>
        </div>
      </UForm>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSopStore } from '~/stores/sop'
import { useGuidelineStore } from '~/stores/guideline'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'

const { t } = useI18n()
const store = useSopStore()
const guidelineStore = useGuidelineStore()
const { canManageCharter } = useRbac()

const statusOptions = computed(() => [
  { label: t('auditCharter.sopForm.active'), value: 'Aktif' },
  { label: t('auditCharter.sopForm.underReview'), value: 'Sedang Diperbarui' }
])

const guidelineOptions = computed(() => {
  return guidelineStore.guidelines.map(g => ({
    value: g.id,
    label: g.name
  }))
})

const handleSubmit = async () => {
  if (!canManageCharter.value) return
  if (store.isEditing) {
    await store.updateSop()
  } else {
    await store.addSop()
  }
}
</script>
