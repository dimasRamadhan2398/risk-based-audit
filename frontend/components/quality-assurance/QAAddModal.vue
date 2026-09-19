<template>
  <UModal
    v-model:open="store.isFormOpen"
    scrollable
    :ui="{
      content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
      body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
      footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
  >
    <template #content>
      <UCard :ui="{ root: 'flex flex-col max-h-[90vh] overflow-hidden', header: 'shrink-0 px-6 py-4 bg-[var(--bg-main)] border-b border-[var(--border-main)]', body: 'px-6 py-6 overflow-y-auto flex-1', footer: 'shrink-0 px-6 py-4 border-t border-[var(--border-main)]' }">
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-xl font-bold">
              {{ store.isEditing ? t('qualityAssurance.modal.editAssessment') : t('qualityAssurance.modal.addNewAssessment') }}
            </h3>
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-lucide-x"
              :aria-label="t('common.close')"
              @click="store.closeForm"
            />
          </div>
        </template>

        <div class="space-y-8">
          <!-- Section 1 -->
          <div class="space-y-4">
            <h4 class="font-bold text-gray-700 dark:text-gray-200">
              {{ t('qualityAssurance.modal.selectType') }}
            </h4>
            <URadioGroup
              v-model="store.newReport.type"
              :items="qaTypeOptions"
              variant="card"
              color="primary"
              class="w-full"
              :ui="{
                fieldset: 'space-y-3',
                item: 'p-4 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900/40 transition-all cursor-pointer has-data-[state=checked]:border-primary-500 has-data-[state=checked]:bg-primary-500/10 dark:has-data-[state=checked]:bg-primary-500/20 dark:has-data-[state=checked]:border-primary-500 hover:bg-gray-50 dark:hover:bg-gray-800/40',
                label: 'font-bold text-sm text-gray-900 dark:text-gray-100 cursor-pointer',
                description: 'text-sm text-gray-500 dark:text-gray-400 mt-0.5'
              }"
            />
          </div>

          <!-- Section 2 -->
          <div class="space-y-4">
            <h4 class="font-bold text-gray-700 dark:text-gray-200">
              {{ t('qualityAssurance.modal.generalInfo') }}
            </h4>
            <div class="space-y-4">
              <UFormField
                :label="t('qualityAssurance.modal.assessmentTitle')"
                required
                :error="errors.assessmentTitle ? t('qualityAssurance.modal.titleRequired') : ''"
              >
                <UInput
                  v-model="store.newReport.assessmentTitle"
                  :placeholder="t('qualityAssurance.modal.titlePlaceholder')"
                  class="w-full"
                  maxlength="100"
                  @invalid="($event.target as any)?.setCustomValidity('Judul maksimal 100 karakter dan wajib diisi')"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.newReport.assessmentTitle ? store.newReport.assessmentTitle.length : 0 }}/100
                </div>
              </UFormField>
              <UFormField :label="t('qualityAssurance.modal.executionPeriod')">
                <USelectMenu
                  v-model="store.newReport.periodYear"
                  :items="store.periods"
                  :placeholder="t('qualityAssurance.modal.selectYear')"
                  class="w-full"
                />
              </UFormField>
            </div>
          </div>

          <!-- Section 3 -->
          <div class="space-y-4">
            <h4 class="font-bold text-gray-700 dark:text-gray-200">
              {{ t('qualityAssurance.modal.resultsAndStatus') }}
            </h4>
            <div class="grid grid-cols-2 gap-4">
              <UFormField
                :label="t('qualityAssurance.modal.status')"
                required
                :error="errors.status ? t('qualityAssurance.modal.statusRequired') : ''"
              >
                <USelectMenu
                  v-model="store.newReport.status"
                  :items="statusOptions"
                  value-key="value"
                  label-key="label"
                  :placeholder="t('qualityAssurance.modal.selectStatus')"
                  class="w-full"
                />
              </UFormField>
              <UFormField
                :label="t('qualityAssurance.modal.resultScore')"
                required
                :error="errors.result ? t('qualityAssurance.modal.resultRequired') : ''"
              >
                <USelectMenu
                  v-if="store.newReport.type === QAType.IACM"
                  v-model="store.newReport.result"
                  :items="['1', '2', '3', '4', '5']"
                  :placeholder="t('qualityAssurance.modal.selectScore')"
                  class="w-full"
                />
                <USelectMenu
                  v-else-if="store.newReport.type === QAType.QAR || store.newReport.type === QAType.SAIV"
                  v-model="store.newReport.result"
                  :items="conformanceOptions"
                  value-key="value"
                  label-key="label"
                  :placeholder="t('qualityAssurance.modal.selectConformance')"
                  class="w-full"
                />
                <UInput
                  v-else-if="store.newReport.type === QAType.REGULAR"
                  v-model="store.newReport.result"
                  :placeholder="t('qualityAssurance.modal.regularResultPlaceholder')"
                  class="w-full"
                  type="number"
                />
                <UInput
                  v-else
                  v-model="store.newReport.result"
                  :placeholder="t('qualityAssurance.modal.defaultResultPlaceholder')"
                  class="w-full"
                />
              </UFormField>
            </div>
          </div>

          <!-- Section 4 -->
          <div class="space-y-4">
            <h4 class="font-bold text-gray-700 dark:text-gray-200">
              {{ t('qualityAssurance.modal.specialDetails') }}
            </h4>
            <div class="grid grid-cols-2 gap-4">
              <UFormField :label="t('qualityAssurance.modal.conductedBy')">
                <UInput
                  v-model="store.newReport.conductedBy"
                  :placeholder="t('qualityAssurance.modal.conductedByPlaceholder')"
                  class="w-full"
                  maxlength="100"
                  @invalid="($event.target as any)?.setCustomValidity('Judul maksimal 100 karakter dan wajib diisi')"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.newReport.conductedBy ? store.newReport.conductedBy.length : 0 }}/100
                </div>
              </UFormField>
              <UFormField :label="t('qualityAssurance.modal.internalEvaluator')">
                <UInput
                  v-model="store.newReport.internalEvaluator"
                  :placeholder="t('qualityAssurance.modal.evaluatorPlaceholder')"
                  class="w-full"
                  maxlength="100"
                  @invalid="($event.target as any)?.setCustomValidity('Judul maksimal 100 karakter dan wajib diisi')"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.newReport.internalEvaluator ? store.newReport.internalEvaluator.length : 0 }}/100
                </div>
              </UFormField>
            </div>
          </div>

          <!-- Section 5 -->
          <div class="space-y-4">
            <h4 class="font-bold text-gray-700 dark:text-gray-200">
              {{ t('qualityAssurance.modal.supportingDocuments') }}
            </h4>
            <UFileUpload
              v-model="store.newReport.attachment"
              :label="t('qualityAssurance.modal.uploadLabel')"
              :description="t('qualityAssurance.modal.uploadDesc')"
              accept=".pdf,.docx,.doc"
              :max-size="10 * 1024 * 1024"
              :icon="'i-lucide-file-up'"
              :file-icon="'i-lucide-file-text'"
              :file-delete="{ color: 'neutral', variant: 'link' }"
              class="w-full"
            />
          </div>
        </div>

        <template #footer>
          <div class="flex justify-end gap-3">
            <UButton
              :label="t('qualityAssurance.modal.cancel')"
              variant="ghost"
              color="neutral"
              @click="store.closeForm"
            />
            <UButton
              :label="store.isEditing ? t('qualityAssurance.modal.updateReport') : t('qualityAssurance.modal.saveReport')"
              color="primary"
              class="px-8 font-bold"
              @click="validateAndSave"
            />
          </div>
        </template>
      </UCard>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useQualityAssuranceStore } from '~/stores/quality-assurance'
import { QAType, QAStatus } from '~/types/quality-assurance'
import { useI18n } from '~/composables/useI18n'

const toast = useToast()
const store = useQualityAssuranceStore()
const { t } = useI18n()

const errors = ref({
  assessmentTitle: false,
  status: false,
  result: false
})

const getTypeLabel = (type: string) => {
  if (store.matchQAType(type, QAType.REGULAR)) return t('qualityAssurance.types.regular')
  if (store.matchQAType(type, QAType.SAIV)) return t('qualityAssurance.types.saiv')
  if (store.matchQAType(type, QAType.QAR)) return t('qualityAssurance.types.qar')
  if (store.matchQAType(type, QAType.IACM)) return t('qualityAssurance.types.iacm')
  return type
}

const getTypeDescription = (type: string) => {
  if (store.matchQAType(type, QAType.REGULAR)) return t('qualityAssurance.modal.regularDesc')
  if (store.matchQAType(type, QAType.SAIV)) return t('qualityAssurance.modal.saivDesc')
  if (store.matchQAType(type, QAType.QAR)) return t('qualityAssurance.modal.qarDesc')
  if (store.matchQAType(type, QAType.IACM)) return t('qualityAssurance.modal.iacmDesc')
  return ''
}

const qaTypeOptions = computed(() =>
  store.qaTypes.map((type) => ({
    value: type,
    label: getTypeLabel(type),
    description: getTypeDescription(type)
  }))
)

const statusOptions = computed(() => [
  { label: t('qualityAssurance.statuses.inProgress'), value: QAStatus.IN_PROGRESS },
  { label: t('qualityAssurance.statuses.completed'), value: QAStatus.COMPLETED },
  { label: t('qualityAssurance.statuses.verified'), value: QAStatus.VERIFIED },
  { label: t('qualityAssurance.statuses.planned'), value: QAStatus.PLANNED }
])

const conformanceOptions = computed(() => [
  { label: t('qualityAssurance.conformance.doesNotConform'), value: 'Does not Conform' },
  { label: t('qualityAssurance.conformance.partiallyConform'), value: 'Partially Conform' },
  { label: t('qualityAssurance.conformance.generallyConformed'), value: 'Generally Conformed' },
  { label: t('qualityAssurance.conformance.fullyConformance'), value: 'Fully Conformance' }
])

const validateAndSave = () => {
  errors.value.assessmentTitle = !store.newReport.assessmentTitle
  errors.value.status = !store.newReport.status
  errors.value.result = !store.newReport.result

  if (errors.value.assessmentTitle || errors.value.status || errors.value.result) {
    toast.add({
      title: t('qualityAssurance.modal.validationErrorTitle'),
      description: t('qualityAssurance.modal.validationErrorDesc'),
      color: 'error',
      icon: 'i-heroicons-exclamation-circle'
    })
    return
  }

  store.saveReport()
}
</script>
