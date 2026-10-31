<template>
  <UModal
    :open="store.isFormOpen"
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
    @update:open="(open: boolean) => { if (!open) store.closeForm() }"
  >
    <template #content>
      <div class="relative flex max-h-[92vh] flex-col transition-colors duration-300">
        <!-- Header -->
        <div class="flex items-center justify-between border-b border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-primary-50 dark:bg-primary-950/40 text-primary-600 dark:text-primary-400">
              <UIcon
                name="i-heroicons-clipboard-document-check"
                class="size-6"
              />
            </div>
            <div>
              <h3 class="text-lg sm:text-xl font-bold text-[var(--text-main)]">
                {{ store.isEditing ? t('qualityAssurance.modal.editAssessment') : t('qualityAssurance.modal.addNewAssessment') }}
              </h3>
              <p class="mt-0.5 text-xs text-[var(--text-muted)]">
                {{ steps[currentStep]?.label }} ({{ t('auditActivityPlan.form.stepNum', { current: currentStep + 1, total: steps.length }) }})
              </p>
            </div>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark-20-solid"
            class="-my-1"
            :aria-label="t('common.close')"
            @click="store.closeForm"
          />
        </div>

        <!-- Stepper -->
        <div class="w-full overflow-x-auto border-b border-[var(--border-main)] bg-[var(--bg-surface)] px-4 py-3 sm:px-6 sm:py-4 scrollbar-thin">
          <div class="flex min-w-max items-center gap-3 px-1 sm:gap-4">
            <template
              v-for="(step, idx) in steps"
              :key="step.key"
            >
              <div
                class="group flex shrink-0 cursor-pointer select-none items-center gap-2 sm:gap-2.5"
                data-testid="form-step"
                @click="goToStep(idx)"
              >
                <div
                  class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-bold transition-all duration-300 sm:h-8 sm:w-8"
                  :class="currentStep === idx
                    ? 'bg-primary text-white shadow-md ring-4 ring-primary/20 scale-105'
                    : currentStep > idx
                      ? 'bg-secondary text-white shadow-sm'
                      : 'bg-gray-100 dark:bg-gray-800 text-gray-400 group-hover:bg-gray-200 dark:group-hover:bg-gray-700'"
                >
                  <UIcon
                    v-if="currentStep > idx"
                    name="i-heroicons-check"
                    class="h-3.5 w-3.5 stroke-[3]"
                  />
                  <span v-else>{{ idx + 1 }}</span>
                </div>
                <span
                  class="whitespace-nowrap text-xs font-semibold transition-colors"
                  :class="currentStep === idx
                    ? 'text-primary font-bold'
                    : currentStep > idx
                      ? 'text-secondary font-semibold'
                      : 'text-gray-400 dark:text-gray-500 group-hover:text-gray-600 dark:group-hover:text-gray-300'"
                >
                  {{ step.label }}
                </span>
              </div>
              <div
                v-if="idx < steps.length - 1"
                class="h-0.5 w-8 shrink-0 rounded transition-all duration-300 sm:w-12"
                :class="currentStep > idx ? 'bg-secondary' : 'bg-gray-200 dark:bg-gray-800'"
              />
            </template>
          </div>
        </div>

        <!-- Body -->
        <div
          ref="bodyRef"
          class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6"
        >
          <!-- Validation / Blocking Alert -->
          <UAlert
            v-if="formErrorMessage"
            color="error"
            variant="subtle"
            icon="i-heroicons-exclamation-triangle"
            :title="t('qualityAssurance.modal.validationErrorTitle')"
            :description="formErrorMessage"
            close
            class="shadow-sm"
            @update:open="(open: boolean) => { if (!open) formErrorMessage = '' }"
          />

          <!-- Step 1: Select Assessment Type -->
          <div v-show="currentStep === 0" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-4' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-clipboard-document-check"
                    class="h-5 w-5 text-primary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('qualityAssurance.modal.selectType') }}
                  </h3>
                </div>
              </template>

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
            </UCard>
          </div>

          <!-- Step 2: General Information dan Results & Status -->
          <div v-show="currentStep === 1" class="space-y-5">
            <!-- 2. General Information -->
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-4' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-information-circle"
                    class="h-5 w-5 text-primary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('qualityAssurance.modal.generalInfo') }}
                  </h3>
                </div>
              </template>

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
            </UCard>

            <!-- 3. Results & Status -->
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-4' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-chart-bar"
                    class="h-5 w-5 text-secondary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('qualityAssurance.modal.resultsAndStatus') }}
                  </h3>
                </div>
              </template>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
            </UCard>
          </div>

          <!-- Step 3: Special Details dan Supporting Documents -->
          <div v-show="currentStep === 2" class="space-y-5">
            <!-- 4. Special Details -->
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-4' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-user-group"
                    class="h-5 w-5 text-primary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('qualityAssurance.modal.specialDetails') }}
                  </h3>
                </div>
              </template>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <UFormField :label="t('qualityAssurance.modal.conductedBy')">
                  <USelectMenu
                    v-model="store.newReport.conductedBy"
                    :items="store.companyOptions"
                    :loading="store.companiesLoading"
                    :placeholder="t('qualityAssurance.modal.conductedByPlaceholder')"
                    class="w-full"
                    create-item
                    @create="(val: string) => { store.newReport.conductedBy = val.trim() }"
                  >
                    <template #empty>
                      {{ store.companiesLoading ? t('common.loading') : (store.companiesError || t('masterData.company.empty')) }}
                    </template>
                  </USelectMenu>
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
            </UCard>

            <!-- 5. Supporting Documents -->
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-4' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-paper-clip"
                    class="h-5 w-5 text-secondary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('qualityAssurance.modal.supportingDocuments') }}
                  </h3>
                </div>
              </template>

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
            </UCard>
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="flex flex-col-reverse sm:flex-row items-center justify-between gap-3 border-t border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
          <div class="flex w-full items-center justify-start gap-2 sm:w-auto">
            <UButton
              :label="t('qualityAssurance.modal.cancel')"
              variant="ghost"
              color="neutral"
              class="w-full sm:w-auto"
              @click="store.closeForm"
            />
          </div>

          <div class="flex w-full items-center justify-end gap-2 sm:w-auto sm:gap-3">
            <UButton
              v-if="currentStep > 0"
              icon="i-heroicons-arrow-left"
              :label="t('auditActivityPlan.stepper.previous')"
              color="neutral"
              variant="outline"
              class="flex-1 sm:flex-none"
              @click="() => { currentStep-- }"
            />
            <UButton
              v-if="currentStep < steps.length - 1"
              trailing-icon="i-heroicons-arrow-right"
              :label="t('auditActivityPlan.stepper.next')"
              color="primary"
              class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
              @click="handleNext"
            />
            <UButton
              v-else
              icon="i-heroicons-check"
              :label="store.isEditing ? t('qualityAssurance.modal.updateReport') : t('qualityAssurance.modal.saveReport')"
              color="primary"
              class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
              :loading="store.loading"
              @click="validateAndSave"
            />
          </div>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { useQualityAssuranceStore } from '~/stores/quality-assurance'
import { QAType, QAStatus } from '~/types/quality-assurance'
import { useI18n } from '~/composables/useI18n'

const toast = useToast()
const store = useQualityAssuranceStore()
const { t } = useI18n()

const currentStep = ref(0)
const bodyRef = ref<HTMLElement | null>(null)
const formErrorMessage = ref('')

const steps = computed(() => [
  { key: 'assessmentType', label: t('qualityAssurance.modal.stepper.assessmentType') || 'Jenis Penilaian' },
  { key: 'generalInfoAndStatus', label: t('qualityAssurance.modal.stepper.generalInfoAndStatus') || 'Informasi & Status' },
  { key: 'detailsAndDocuments', label: t('qualityAssurance.modal.stepper.detailsAndDocuments') || 'Detail & Dokumen' }
])

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

const validateStep = (idx: number): boolean => {
  if (idx === 0) {
    if (!store.newReport.type) {
      formErrorMessage.value = t('qualityAssurance.modal.selectType')
      return false
    }
    formErrorMessage.value = ''
    return true
  }
  if (idx === 1) {
    errors.value.assessmentTitle = !store.newReport.assessmentTitle?.trim()
    errors.value.status = !store.newReport.status
    errors.value.result = !store.newReport.result?.toString().trim()

    if (errors.value.assessmentTitle || errors.value.status || errors.value.result) {
      formErrorMessage.value = t('qualityAssurance.modal.validationErrorDesc')
      toast.add({
        title: t('qualityAssurance.modal.validationErrorTitle'),
        description: t('qualityAssurance.modal.validationErrorDesc'),
        color: 'error',
        icon: 'i-heroicons-exclamation-circle'
      })
      return false
    }
    formErrorMessage.value = ''
    return true
  }
  return true
}

const goToStep = (targetIdx: number) => {
  if (targetIdx <= currentStep.value) {
    currentStep.value = targetIdx
    formErrorMessage.value = ''
    return
  }
  for (let i = currentStep.value; i < targetIdx; i++) {
    if (!validateStep(i)) return
  }
  formErrorMessage.value = ''
  currentStep.value = targetIdx
  nextTick(() => {
    bodyRef.value?.scrollTo?.({ top: 0, behavior: 'smooth' })
  })
}

const handleNext = () => {
  if (!validateStep(currentStep.value)) return
  formErrorMessage.value = ''
  currentStep.value++
  nextTick(() => {
    bodyRef.value?.scrollTo?.({ top: 0, behavior: 'smooth' })
  })
}

const validateAndSave = () => {
  if (!validateStep(0)) {
    currentStep.value = 0
    return
  }
  if (!validateStep(1)) {
    currentStep.value = 1
    return
  }
  formErrorMessage.value = ''
  store.saveReport()
}

watch(
  () => store.isFormOpen,
  (open) => {
    if (open) {
      currentStep.value = 0
      formErrorMessage.value = ''
      errors.value.assessmentTitle = false
      errors.value.status = false
      errors.value.result = false
      if (!store.companiesLoaded && !store.companiesLoading) {
        store.fetchCompanies().catch(() => {})
      }
    }
  }
)

onMounted(() => {
  if (!store.companiesLoaded && !store.companiesLoading) {
    store.fetchCompanies().catch(() => {})
  }
})
</script>
