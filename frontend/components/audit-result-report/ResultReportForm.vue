<template>
  <UModal
    :open="store.showModal"
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-5xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
    @update:open="(open: boolean) => { if (!open) store.closeModal() }"
  >
    <template #content>
      <div class="relative flex max-h-[92vh] flex-col transition-colors duration-300">
        <!-- Header -->
        <div class="flex items-center justify-between border-b border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-primary-50 dark:bg-primary-950/40 text-primary-600 dark:text-primary-400">
              <UIcon
                name="i-heroicons-document-text"
                class="size-6"
              />
            </div>
            <div>
              <h3 class="text-lg sm:text-xl font-bold text-[var(--text-main)]">
                {{ store.isEditing ? t('auditResultReport.editor.form.titleEdit') : t('auditResultReport.editor.form.titleCreate') }}
              </h3>
              <p class="mt-0.5 text-xs text-[var(--text-muted)]">
                {{ steps[currentStep]?.label }} ({{ t('auditActivityPlan.form.stepNum', { current: currentStep + 1, total: steps.length }) }})
                <span class="mx-1.5 opacity-60">·</span>
                {{ t('auditResultReport.editor.form.assignmentLetter') }}
                <span class="font-semibold text-primary-600 dark:text-primary-400">{{ store.reportForm.assignmentLetterId || store.selectedAssignmentLetter || '-' }}</span>
              </p>
            </div>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark-20-solid"
            class="-my-1"
            :aria-label="t('auditResultReport.editor.form.cancel')"
            @click="store.closeModal"
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
            :title="t('common.error')"
            :description="formErrorMessage"
            close
            class="shadow-sm"
            @update:open="(open: boolean) => { if (!open) formErrorMessage = '' }"
          />

          <!-- Step 1: Report Information -->
          <div v-show="currentStep === 0" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-document-text"
                    class="h-5 w-5 text-primary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('auditResultReport.editor.form.stepper.reportInfo') }}
                  </h3>
                </div>
              </template>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-5 sm:gap-6">
                <UFormField :label="t('auditResultReport.editor.form.reportNumber')" name="reportNumber" required class="md:col-span-2">
                  <div class="flex gap-2 items-center">
                    <UInput
                      v-model="store.reportForm.reportNumber"
                      :placeholder="t('auditResultReport.editor.form.reportNumberExample')"
                      class="w-full font-medium"
                      required
                      disabled
                    />
                  </div>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    {{ t('auditResultReport.editor.form.automaticFormat') }}
                    <span class="text-primary-600 dark:text-primary-400 font-medium">{{ t('auditResultReport.editor.form.formatExample') }}</span>
                  </p>
                </UFormField>

                <UFormField :label="t('auditResultReport.editor.form.company')" name="companyId" class="md:col-span-2">
                  <USelectMenu
                    v-model="store.reportForm.companyId"
                    :items="companySelectItems"
                    value-key="value"
                    :placeholder="t('auditResultReport.editor.form.companyPlaceholder')"
                    class="w-full"
                    :clear="true"
                    :loading="masterOptions.loading"
                    @update:model-value="onCompanyChange"
                  />
                </UFormField>

                <UFormField :label="t('auditResultReport.dashboard.columns.reportTitle')" name="reportTitle" required class="md:col-span-2">
                  <UInput
                    v-model="store.reportForm.reportTitle"
                    :placeholder="t('auditResultReport.editor.form.titleExample')"
                    class="w-full"
                    maxlength="100"
                    @invalid="($event.target as any)?.setCustomValidity(t('auditResultReport.editor.form.reportTitleInvalid'))"
                    @input="($event.target as any)?.setCustomValidity('')"
                  />
                  <div class="text-xs text-gray-500 dark:text-gray-400 mt-1 text-right">
                    {{ store.reportForm.reportTitle ? store.reportForm.reportTitle.length : 0 }}/100
                  </div>
                </UFormField>

                <UFormField
                  :label="t('auditResultReport.editor.form.reportDate')"
                  name="reportDate"
                  required
                >
                  <AppDatePicker
                    v-model="store.reportForm.reportDate"
                    class="w-full"
                  />
                </UFormField>

                <UFormField
                  :label="t('strategicPlan.columns.status')"
                  name="status"
                >
                  <USelectMenu
                    v-model="store.reportForm.status"
                    :items="availableStatuses"
                    value-key="value"
                    class="w-full"
                    :disabled="!canPublishLha"
                  />
                  <p v-if="!canPublishLha" class="text-[11px] text-amber-600 dark:text-amber-400 mt-1">
                    {{ t('auditResultReport.editor.form.statusLocked') }}
                  </p>
                </UFormField>
              </div>
            </UCard>
          </div>

          <!-- Step 2: Audit Findings -->
          <div v-show="currentStep === 1" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
                  <div class="flex items-center gap-2">
                    <UIcon
                      name="i-heroicons-list-bullet"
                      class="h-5 w-5 text-secondary-500"
                    />
                    <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                      {{ t('auditResultReport.editor.form.stepper.auditFindings') }}
                    </h3>
                    <UBadge
                      color="secondary"
                      variant="subtle"
                      size="sm"
                    >
                      {{ store.reportForm.findings?.length || 0 }}
                    </UBadge>
                  </div>

                  <div class="flex items-center gap-2 w-full sm:w-auto justify-end">
                    <UTooltip :text="t('auditResultReport.editor.form.syncDescription')">
                      <UButton
                        :label="t('auditResultReport.editor.form.sync')"
                        size="sm"
                        color="secondary"
                        variant="solid"
                        :loading="store.isAutoDetecting"
                        @click="store.runAutoDetectFindings('replace')"
                      />
                    </UTooltip>
                    <UButton
                      :label="t('auditResultReport.editor.form.addManual')"
                      size="sm"
                      color="neutral"
                      variant="outline"
                      icon="i-heroicons-plus"
                      @click="addManualFinding"
                    />
                  </div>
                </div>
              </template>

              <!-- Findings Category Summary Badge Bar -->
              <div
                v-if="store.reportForm.findings && store.reportForm.findings.length > 0"
                class="flex flex-wrap items-center gap-2 p-3 bg-secondary-50/40 dark:bg-secondary-950/20 rounded-lg border border-secondary-200/60 dark:border-secondary-900/40 text-xs"
              >
                <span class="font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-1 mr-1">
                  <UIcon
                    name="i-heroicons-chart-pie"
                    class="size-4 text-secondary-600 dark:text-secondary-400"
                  />
                  {{ t('auditResultReport.editor.form.categoryDistribution') }}
                </span>
                <UBadge
                  color="error"
                  variant="subtle"
                  size="md"
                >
                  {{ t('auditResultReport.editor.form.verySignificant') }} {{ countCategory('Very Significant') }}
                </UBadge>
                <UBadge
                  color="warning"
                  variant="subtle"
                  size="md"
                >
                  {{ t('auditResultReport.editor.form.significant') }} {{ countCategory('Significant') }}
                </UBadge>
                <UBadge
                  color="info"
                  variant="subtle"
                  size="md"
                >
                  {{ t('auditResultReport.editor.form.quiteSignificant') }} {{ countCategory('Quite Significant') }}
                </UBadge>
                <UBadge
                  color="success"
                  variant="subtle"
                  size="md"
                >
                  {{ t('auditResultReport.editor.form.notSignificant') }} {{ countCategory('Not Significant') }}
                </UBadge>
              </div>

              <!-- Findings Cards -->
              <div class="space-y-3">
                <div
                  v-for="(finding, idx) in store.reportForm.findings"
                  :key="idx"
                  class="flex items-start gap-3 bg-secondary-50/30 dark:bg-secondary-950/20 p-4 rounded-xl border border-secondary-200/70 dark:border-secondary-900/40 transition shadow-sm hover:border-secondary-300 dark:hover:border-secondary-700"
                >
                  <div class="flex-1 space-y-3">
                    <!-- Card Header (Index & Source Badge) -->
                    <div class="flex justify-between items-center border-b border-secondary-100 dark:border-secondary-900/40 pb-2">
                      <span class="text-md font-bold text-secondary-700 dark:text-secondary-300">
                        {{ t('auditResultReport.editor.form.findingNumber') }}{{ idx + 1 }}
                      </span>
                      <UBadge
                        :color="getSourceColor(finding.source)"
                        variant="subtle"
                        size="md"
                        class="flex items-center gap-1 font-medium"
                      >
                        <UIcon
                          :name="getSourceIcon(finding.source)"
                          class="size-3"
                        />
                        {{ finding.source || 'Manual' }}
                      </UBadge>
                    </div>

                    <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                      <UFormField
                        :label="t('auditResultReport.editor.form.category')"
                        size="sm"
                        class="md:col-span-1"
                      >
                        <USelectMenu
                          v-model="finding.category"
                          :items="categoryItems"
                          value-key="value"
                          class="w-full"
                        />
                      </UFormField>

                      <UFormField
                        :label="t('auditResultReport.editor.form.findingDescription')"
                        size="sm"
                        class="md:col-span-1"
                      >
                        <UInput
                          v-model="finding.title"
                          :placeholder="t('auditResultReport.editor.form.findingExample')"
                          required
                          class="w-full"
                          maxlength="255"
                        />
                        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1 text-right">
                          {{ finding.title ? finding.title.length : 0 }}/255
                        </div>
                      </UFormField>

                      <UFormField
                        :label="t('auditResultReport.editor.form.action')"
                        size="sm"
                        class="md:col-span-2"
                      >
                        <UTextarea
                          v-model="finding.action"
                          :placeholder="t('auditResultReport.editor.form.actionExample')"
                          :rows="2"
                          class="w-full"
                        />
                      </UFormField>
                    </div>
                  </div>

                  <UTooltip :text="t('auditResultReport.editor.form.deleteFinding')">
                    <UButton
                      color="error"
                      variant="ghost"
                      icon="i-heroicons-trash"
                      size="sm"
                      class="mt-1"
                      @click="removeFinding(idx)"
                    />
                  </UTooltip>
                </div>

                <!-- Empty State in Findings List -->
                <div
                  v-if="!store.reportForm.findings || store.reportForm.findings.length === 0"
                  class="text-center py-8 bg-secondary-50/40 dark:bg-secondary-950/30 rounded-xl border border-dashed border-secondary-200 dark:border-secondary-900/50 space-y-2"
                >
                  <div class="p-3 bg-secondary-100 dark:bg-secondary-900/40 rounded-full w-fit mx-auto text-secondary-600 dark:text-secondary-400">
                    <UIcon
                      name="i-heroicons-sparkles"
                      class="size-6"
                    />
                  </div>
                  <p class="text-sm font-semibold text-gray-700 dark:text-gray-200">
                    {{ t('auditResultReport.editor.form.empty') }}
                  </p>
                  <p class="text-xs text-gray-500 dark:text-gray-400 max-w-md mx-auto">
                    {{ t('auditResultReport.editor.form.click') }} <strong>{{ t('auditResultReport.editor.form.sync') }}</strong> {{ t('auditResultReport.editor.form.emptyInstruction') }}
                  </p>
                  <div class="flex justify-center gap-2 pt-2">
                    <UButton
                      :label="t('auditResultReport.editor.form.sync')"
                      size="sm"
                      color="secondary"
                      icon="i-heroicons-sparkles"
                      :loading="store.isAutoDetecting"
                      @click="store.runAutoDetectFindings('replace')"
                    />
                    <UButton
                      :label="t('auditResultReport.editor.form.addManual')"
                      size="sm"
                      color="neutral"
                      variant="outline"
                      icon="i-heroicons-plus"
                      @click="addManualFinding"
                    />
                  </div>
                </div>
              </div>
            </UCard>
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="flex flex-col-reverse sm:flex-row items-center justify-between gap-3 border-t border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
          <div class="flex w-full items-center justify-start gap-2 sm:w-auto">
            <UButton
              :label="t('auditResultReport.editor.form.cancel')"
              color="neutral"
              variant="ghost"
              class="w-full sm:w-auto"
              @click="store.closeModal"
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
              :icon="isFinalStatus ? 'i-heroicons-pencil-square' : 'i-heroicons-check'"
              :label="submitButtonLabel"
              color="primary"
              class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
              :loading="store.loading"
              @click="onFormSubmit"
            />
          </div>
        </div>
      </div>
    </template>
  </UModal>

  <!-- Modal Input Tanda Tangan Team Member (Opens upon clicking Save Report) -->
  <ResultReportSignatureModal
    v-model:open="showSignatureModal"
    :assignment-letter-id="store.reportForm.assignmentLetterId || store.selectedAssignmentLetter"
    :report-number="store.reportForm.reportNumber"
    :company-name="store.reportForm.companyName"
    :initial-place="store.reportForm.signaturePlace"
    :initial-date="store.reportForm.signatureDate || store.reportForm.reportDate"
    :existing-signatures="store.reportForm.signatures"
    :team-members="currentTeamMembers"
    :status="store.reportForm.status"
    :loading="store.loading"
    @confirm="onSignatureConfirm"
  />
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useAuditResultReportStore, type MemberSignature } from '~/stores/audit-result-report'
import { useMasterOptionsStore } from '~/stores/master-options'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import ResultReportSignatureModal from './ResultReportSignatureModal.vue'

const { t } = useI18n()
const store = useAuditResultReportStore()
const masterOptions = useMasterOptionsStore()
const { canPublishLha } = useRbac()

const currentStep = ref(0)
const bodyRef = ref<HTMLElement | null>(null)
const formErrorMessage = ref('')
const showSignatureModal = ref(false)

const steps = computed(() => [
  { key: 'reportInfo', label: t('auditResultReport.editor.form.stepper.reportInfo') || 'Informasi Laporan' },
  { key: 'auditFindings', label: t('auditResultReport.editor.form.stepper.auditFindings') || 'Temuan Audit' }
])

const availableStatuses = computed(() => {
  const draft = { label: t('auditResultReport.dashboard.status.draft'), value: 'Draft' }
  const final = { label: t('auditResultReport.dashboard.status.final'), value: 'Final' }
  return canPublishLha.value ? [draft, final] : [draft]
})

const categoryItems = computed(() => [
  { label: t('auditResultReport.dashboard.categories.verySignificant'), value: 'Very Significant' },
  { label: t('auditResultReport.dashboard.categories.significant'), value: 'Significant' },
  { label: t('auditResultReport.dashboard.categories.quiteSignificant'), value: 'Quite Significant' },
  { label: t('auditResultReport.dashboard.categories.notSignificant'), value: 'Not Significant' }
])

const isFinalStatus = computed(() => {
  const s = (store.reportForm.status || '').toLowerCase().trim()
  return s === 'final' || s === 'approved' || s === 'published'
})

const submitButtonLabel = computed(() => {
  if (isFinalStatus.value) {
    return t('auditResultReport.editor.form.proceedToSignature') || t('auditResultReport.editor.form.signReport')
  }
  return store.isEditing ? t('common.save') : (t('auditResultReport.editor.form.saveDraft') || 'Simpan Draf Laporan')
})

const currentTeamMembers = computed(() => {
  const letter = store.reportForm.assignmentLetterId || store.selectedAssignmentLetter
  return store.getTeamMembersForLetter(letter)
})

const validateStep = (idx: number): boolean => {
  if (idx === 0) {
    const letter = String(store.reportForm.assignmentLetterId || store.selectedAssignmentLetter || '').trim()
    if (!letter) {
      store.formErrors.assignmentLetterId = 'auditResultReport.form.assignmentLetterRequired'
      formErrorMessage.value = t('auditResultReport.form.assignmentLetterRequired') || 'Assignment letter is required'
      return false
    }
    store.formErrors.assignmentLetterId = ''

    if (!store.reportForm.reportTitle?.trim()) {
      formErrorMessage.value = t('auditResultReport.editor.form.reportTitleInvalid') || 'Report Title is required'
      return false
    }
    formErrorMessage.value = ''
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

const openSignatureModal = () => {
  const letter = String(store.reportForm.assignmentLetterId || store.selectedAssignmentLetter || '').trim()
  if (!letter) {
    store.formErrors.assignmentLetterId = 'auditResultReport.form.assignmentLetterRequired'
    formErrorMessage.value = t('auditResultReport.form.assignmentLetterRequired') || 'Assignment letter is required'
    currentStep.value = 0
    return
  }
  showSignatureModal.value = true
}

const onFormSubmit = async () => {
  if (!validateStep(0)) {
    currentStep.value = 0
    return
  }
  formErrorMessage.value = ''
  if (isFinalStatus.value) {
    openSignatureModal()
  } else {
    await store.saveReport()
  }
}

const onSignatureConfirm = async (payload: {
  signaturePlace: string
  signatureDate: string
  companyName: string
  signatures: MemberSignature[]
}) => {
  store.reportForm.signaturePlace = payload.signaturePlace
  store.reportForm.signatureDate = payload.signatureDate
  store.reportForm.companyName = payload.companyName
  store.reportForm.signatures = payload.signatures
  showSignatureModal.value = false
  await store.saveReport()
}

const companySelectItems = computed(() => [
  { label: t('auditResultReport.editor.form.companyNone'), value: '__none__' },
  ...masterOptions.companyOptions
])

const onCompanyChange = (val: string) => {
  if (!val || val === '__none__') {
    store.reportForm.companyId = ''
    store.reportForm.companyName = ''
    return
  }
  const comp = masterOptions.companyOptions.find(c => c.value === val)
  store.reportForm.companyName = comp ? comp.label : ''
}

watch(
  () => store.showModal,
  (open) => {
    if (open) {
      currentStep.value = 0
      formErrorMessage.value = ''
      masterOptions.fetchAll()
    }
  },
  { immediate: true }
)

const countCategory = (cat: string) => {
  if (!store.reportForm.findings) return 0
  return store.reportForm.findings.filter((f: any) => f.category === cat).length
}

const getSourceColor = (source?: string) => {
  if (!source) return 'neutral'
  if (source.includes('KKA') || source.includes('Working Paper')) return 'primary'
  if (source.includes('Fieldwork')) return 'info'
  if (source.includes('ATR') || source.includes('Action Taken')) return 'warning'
  return 'neutral'
}

const getSourceIcon = (source?: string) => {
  if (!source) return 'i-heroicons-pencil'
  if (source.includes('KKA') || source.includes('Working Paper')) return 'i-heroicons-document-text'
  if (source.includes('Fieldwork')) return 'i-heroicons-shield-check'
  if (source.includes('ATR') || source.includes('Action Taken')) return 'i-heroicons-arrow-path'
  return 'i-heroicons-pencil'
}

const addManualFinding = () => {
  if (!store.reportForm.findings) store.reportForm.findings = []
  store.reportForm.findings.push({
    title: '',
    category: 'Significant',
    action: '',
    source: 'Manual'
  })
  store.reportForm.findingsCount = store.reportForm.findings.length
}

const removeFinding = (idx: number) => {
  if (store.reportForm.findings) {
    store.reportForm.findings.splice(idx, 1)
    store.reportForm.findingsCount = store.reportForm.findings.length
  }
}
</script>
