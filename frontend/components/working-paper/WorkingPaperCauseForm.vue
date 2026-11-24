<template>
  <UModal 
    :open="store.showModalF04" 
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
    @update:open="(open: boolean) => { if (!open) store.closeModalF04() }"
  >
    <template #content>
      <div class="relative flex max-h-[92vh] flex-col transition-colors duration-300">
        <!-- Header -->
        <div class="flex items-center justify-between border-b border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-primary-50 dark:bg-primary-950/40 text-primary-600 dark:text-primary-400">
              <UIcon
                name="i-heroicons-document-magnifying-glass"
                class="size-6"
              />
            </div>
            <div>
              <h3 class="text-lg sm:text-xl font-bold text-[var(--text-main)]">
                {{ store.isEditingF04 ? (t('workingPaper.causeForm.editTitle') || t('workingPaper.causeForm.title')) : t('workingPaper.causeForm.title') }}
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
            @click="store.closeModalF04"
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

          <!-- Step 1: Condition sampai Impact -->
          <div v-show="currentStep === 0" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-exclamation-circle"
                    class="h-5 w-5 text-primary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('workingPaper.causeForm.stepper.conditionAndImpact') }}
                  </h3>
                </div>
              </template>

              <!-- Condition (Field Impact) -->
              <UFormField 
                :label="t('workingPaper.causeForm.condition')" 
                name="condition" 
                required
              >
                <UTextarea 
                  v-model="store.causeForm.condition" 
                  :rows="3" 
                  :placeholder="t('workingPaper.causeForm.conditionPlaceholder')" 
                  class="w-full"
                  maxlength="200"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.causeForm.condition ? store.causeForm.condition.length : 0 }}/200
                </div>
              </UFormField>

              <!-- 4 Field Tambahan pada Section Kondisi -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField
                  :label="t('workingPaper.causeForm.riskEvent')"
                  name="riskEvent"
                >
                  <UTextarea
                    v-model="store.causeForm.riskEvent"
                    :rows="3"
                    :placeholder="t('workingPaper.causeForm.riskEventPlaceholder')"
                    class="w-full"
                  />
                </UFormField>

                <UFormField
                  :label="t('workingPaper.causeForm.cause')"
                  name="cause"
                >
                  <UTextarea
                    v-model="store.causeForm.cause"
                    :rows="3"
                    :placeholder="t('workingPaper.causeForm.causePlaceholder')"
                    class="w-full"
                  />
                </UFormField>

                <UFormField
                  :label="t('workingPaper.causeForm.riskLevel')"
                  name="riskLevel"
                >
                  <USelectMenu
                    v-model="store.causeForm.riskLevel"
                    :items="riskLevelItems"
                    value-key="value"
                    :placeholder="t('workingPaper.causeForm.riskLevelPlaceholder')"
                    class="w-full"
                  />
                </UFormField>

                <UFormField
                  :label="t('workingPaper.causeForm.mitigationPlanOption')"
                  name="mitigationPlanOption"
                >
                  <UTextarea
                    v-model="store.causeForm.mitigationPlanOption"
                    :rows="3"
                    :placeholder="t('workingPaper.causeForm.mitigationPlanOptionPlaceholder')"
                    class="w-full"
                  />
                </UFormField>
              </div>

              <!-- Upload Evidence -->
              <div class="space-y-2">
                <label class="block font-semibold text-sm text-[var(--text-main)]">
                  {{ t('workingPaper.causeForm.uploadEvidence') }}
                </label>
                <div @click="store.triggerUpload">
                  <LocalizedFileInput
                    accept="image/png, image/jpeg"
                    @change="store.onFileChange"
                  />
                </div>
              </div>

              <!-- Criteria (Related Rules/SOP) -->
              <UFormField 
                :label="t('workingPaper.causeForm.criteria')" 
                name="criteria" 
                required
              >
                <USelectMenu
                  v-model="store.causeForm.criteria"
                  :items="store.criteriaOptions"
                  value-key="value"
                  :loading="store.loadingSopOptions"
                  :placeholder="t('workingPaper.causeForm.criteriaPlaceholder')"
                  class="w-full"
                />
              </UFormField>

              <!-- Impact (Risk Impact) -->
              <UFormField 
                :label="t('workingPaper.causeForm.impact')" 
                name="impact" 
                required
              >
                <UTextarea
                  v-model="store.causeForm.impact"
                  :rows="3"
                  :placeholder="t('workingPaper.causeForm.impactPlaceholder')"
                  class="w-full"
                />
              </UFormField>
            </UCard>
          </div>

          <!-- Step 2: Root Cause Analysis -->
          <div v-show="currentStep === 1" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
                  <div class="flex items-center gap-2">
                    <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                      {{ t('workingPaper.causeForm.rootCauseAnalysis') }}
                    </h3>
                    <UBadge
                      color="secondary"
                      variant="subtle"
                      size="sm"
                    >
                      {{ store.causeForm.rootCause?.length || 0 }}
                    </UBadge>
                  </div>
                  <UButton
                    color="primary"
                    icon="i-heroicons-plus"
                    variant="soft"
                    size="sm"
                    :label="t('workingPaper.causeForm.addRootCause')"
                    @click="store.addRootCause"
                  />
                </div>
              </template>

              <!-- Empty state if no root cause added yet -->
              <div
                v-if="!store.causeForm.rootCause || store.causeForm.rootCause.length === 0"
                class="text-center py-8 bg-secondary-50/40 dark:bg-secondary-950/30 rounded-xl border border-dashed border-secondary-200 dark:border-secondary-900/50 space-y-2"
              >
                <p class="text-sm font-semibold text-gray-700 dark:text-gray-200">
                  {{ t('workingPaper.causeForm.emptyRootCause') || 'Belum ada analisis akar masalah' }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400 max-w-md mx-auto">
                  {{ t('workingPaper.causeForm.emptyRootCauseHint') || 'Klik tombol Tambah Analisa Akar Masalah di bawah untuk mulai menambahkan analisis.' }}
                </p>
                <div class="pt-2">
                  <UButton
                    color="primary"
                    icon="i-heroicons-plus"
                    variant="soft"
                    size="sm"
                    :label="t('workingPaper.causeForm.addRootCause')"
                    @click="store.addRootCause"
                  />
                </div>
              </div>

              <!-- Root cause items list -->
              <div v-else class="space-y-4">
                <div
                  v-for="(rca, index) in store.causeForm.rootCause"
                  :key="rca.id || index"
                  class="border border-gray-200 dark:border-gray-800 rounded-xl p-4 sm:p-5 bg-gray-50/50 dark:bg-gray-800/40 space-y-4"
                >
                  <div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-800 pb-3">
                    <span class="text-sm font-bold text-[var(--text-main)]">
                      {{ t('workingPaper.causeForm.analysisIndex', { index: index + 1 }) }}
                    </span>
                    <UButton
                      icon="i-heroicons-trash"
                      color="error"
                      variant="ghost"
                      size="sm"
                      @click="store.removeRootCause(index)"
                    />
                  </div>

                  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <UFormField :label="t('workingPaper.causeForm.category')">
                      <USelectMenu
                        v-model="rca.method"
                        :items="store.options.rootCauseMethod"
                        :placeholder="t('workingPaper.causeForm.categoryPlaceholder')"
                        class="w-full"
                      />
                    </UFormField>

                    <UFormField :label="t('workingPaper.causeForm.why1')" required>
                      <UInput
                        v-model="rca.w1"
                        :placeholder="t('workingPaper.causeForm.why1Placeholder')"
                        required
                        maxlength="100"
                        class="w-full"
                        @invalid="($event.target as any)?.setCustomValidity(t('workingPaper.causeForm.why1Validation'))"
                        @input="($event.target as any)?.setCustomValidity('')"
                      />
                    </UFormField>
                  </div>
                </div>
              </div>
            </UCard>

            <!-- Section Rekomendasi (lebih dari satu rekomendasi) -->
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
                  <div class="flex items-center gap-2">
                    <UIcon
                      name="i-heroicons-light-bulb"
                      class="h-5 w-5 text-primary-500"
                    />
                    <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                      {{ t('workingPaper.causeForm.recommendations') }}
                    </h3>
                    <UBadge
                      color="secondary"
                      variant="subtle"
                      size="sm"
                    >
                      {{ store.causeForm.recommendations?.length || 0 }}
                    </UBadge>
                  </div>
                  <UButton
                    color="primary"
                    icon="i-heroicons-plus"
                    variant="soft"
                    size="sm"
                    :label="t('workingPaper.causeForm.addRecommendation')"
                    @click="store.addRecommendation"
                  />
                </div>
              </template>

              <div
                v-if="!store.causeForm.recommendations || store.causeForm.recommendations.length === 0"
                class="text-center py-6 bg-secondary-50/40 dark:bg-secondary-950/30 rounded-xl border border-dashed border-secondary-200 dark:border-secondary-900/50 space-y-2"
              >
                <p class="text-sm font-semibold text-gray-700 dark:text-gray-200">
                  {{ t('workingPaper.causeForm.emptyRecommendations') }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400 max-w-md mx-auto">
                  {{ t('workingPaper.causeForm.emptyRecommendationsHint') }}
                </p>
                <div class="pt-2">
                  <UButton
                    color="primary"
                    icon="i-heroicons-plus"
                    variant="soft"
                    size="sm"
                    :label="t('workingPaper.causeForm.addRecommendation')"
                    @click="store.addRecommendation"
                  />
                </div>
              </div>

              <div v-else class="space-y-3">
                <div
                  v-for="(rec, index) in store.causeForm.recommendations"
                  :key="rec.id || index"
                  class="border border-gray-200 dark:border-gray-800 rounded-xl p-4 bg-gray-50/50 dark:bg-gray-800/40 space-y-3"
                >
                  <div class="flex items-center justify-between">
                    <span class="text-sm font-bold text-[var(--text-main)]">
                      {{ t('workingPaper.causeForm.recommendationIndex', { index: index + 1 }) }}
                    </span>
                    <UButton
                      icon="i-heroicons-trash"
                      color="error"
                      variant="ghost"
                      size="sm"
                      @click="store.removeRecommendation(index)"
                    />
                  </div>
                  <UTextarea
                    v-model="rec.recommendation"
                    :rows="2"
                    :placeholder="t('workingPaper.causeForm.recommendationPlaceholder')"
                    class="w-full"
                  />
                </div>
              </div>
            </UCard>
          </div>

          <!-- Step 3: Tab Baru — Tanggapan Auditee & Rencana Aksi -->
          <div v-show="currentStep === 2" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-clipboard-document-check"
                    class="h-5 w-5 text-secondary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('workingPaper.causeForm.stepper.responseAndActionPlan') }}
                  </h3>
                </div>
              </template>

              <!-- Tanggapan Auditee -->
              <UFormField
                :label="t('workingPaper.causeForm.auditeeResponse')"
                name="auditeeResponse"
              >
                <UTextarea
                  v-model="store.causeForm.auditeeResponse"
                  :rows="3"
                  :placeholder="t('workingPaper.causeForm.auditeeResponsePlaceholder')"
                  class="w-full"
                />
              </UFormField>

              <!-- Rencana Aksi -->
              <UFormField
                :label="t('workingPaper.causeForm.actionPlan')"
                name="actionPlan"
              >
                <UTextarea
                  v-model="store.causeForm.actionPlan"
                  :rows="3"
                  :placeholder="t('workingPaper.causeForm.actionPlanPlaceholder')"
                  class="w-full"
                />
              </UFormField>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- Target Penyelesaian -->
                <UFormField
                  :label="t('workingPaper.causeForm.targetDate')"
                  name="targetDate"
                >
                  <AppDatePicker
                    v-model="store.causeForm.targetDate"
                    class="w-full"
                  />
                </UFormField>

                <!-- Evidence -->
                <UFormField
                  :label="t('workingPaper.causeForm.actionEvidence')"
                  name="actionEvidence"
                >
                  <UInput
                    v-model="store.causeForm.actionEvidence"
                    :placeholder="t('workingPaper.causeForm.actionEvidencePlaceholder')"
                    class="w-full"
                  />
                </UFormField>

                <!-- PIC (diisi Manager/Supervisor) -->
                <UFormField
                  :label="t('workingPaper.causeForm.pic')"
                  name="pic"
                >
                  <USelectMenu
                    v-model="store.causeForm.pic"
                    icon="i-heroicons-magnifying-glass"
                    :items="store.picOptions"
                    :loading="store.employeesLoading"
                    :placeholder="t('workingPaper.causeForm.picPlaceholder')"
                    class="w-full"
                    create-item
                    @create="(val: string) => { store.causeForm.pic = val.trim() }"
                  >
                    <template #empty>
                      {{ store.employeesLoading ? t('common.loading') : (store.employeesError || t('masterData.employee.empty')) }}
                    </template>
                  </USelectMenu>
                </UFormField>

                <!-- Final Approver (diisi GM/Manager Unit/VP/EVP/Direktur) -->
                <UFormField
                  :label="t('workingPaper.causeForm.finalApprover')"
                  name="finalApprover"
                >
                  <USelectMenu
                    v-model="store.causeForm.finalApprover"
                    icon="i-heroicons-magnifying-glass"
                    :items="store.finalApproverOptions"
                    :loading="store.employeesLoading"
                    :placeholder="t('workingPaper.causeForm.finalApproverPlaceholder')"
                    class="w-full"
                    create-item
                    @create="(val: string) => { store.causeForm.finalApprover = val.trim() }"
                  >
                    <template #empty>
                      {{ store.employeesLoading ? t('common.loading') : (store.employeesError || t('masterData.employee.empty')) }}
                    </template>
                  </USelectMenu>
                </UFormField>
              </div>
            </UCard>
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="flex flex-col-reverse sm:flex-row items-center justify-between gap-3 border-t border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
          <div class="flex w-full items-center justify-start gap-2 sm:w-auto">
            <UButton
              :label="t('common.cancel')"
              color="neutral"
              variant="ghost"
              class="w-full sm:w-auto"
              @click="store.closeModalF04"
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
              :label="store.isEditingF04 ? t('common.updateData') : t('common.submit')"
              color="primary"
              class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
              @click="handleSave"
            />
          </div>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { useWorkingPaperStore } from '~/stores/working-paper'
import LocalizedFileInput from '~/components/shared/LocalizedFileInput.vue'
import AppDatePicker from '~/components/shared/AppDatePicker.vue'

const { t } = useI18n()
const store = useWorkingPaperStore()

const currentStep = ref(0)
const bodyRef = ref<HTMLElement | null>(null)
const formErrorMessage = ref('')

const riskLevelItems = computed(() => [
  { label: t('workingPaper.causeForm.riskLevelOptions.level1'), value: '1' },
  { label: t('workingPaper.causeForm.riskLevelOptions.level2'), value: '2' },
  { label: t('workingPaper.causeForm.riskLevelOptions.level3'), value: '3' },
  { label: t('workingPaper.causeForm.riskLevelOptions.level4'), value: '4' },
  { label: t('workingPaper.causeForm.riskLevelOptions.level5'), value: '5' }
])

const steps = computed(() => [
  { key: 'conditionAndImpact', label: t('workingPaper.causeForm.stepper.conditionAndImpact') || 'Kondisi & Dampak' },
  { key: 'rootCauseAnalysis', label: t('workingPaper.causeForm.stepper.rootCauseAnalysis') || 'Analisis Akar Masalah' },
  { key: 'responseAndActionPlan', label: t('workingPaper.causeForm.stepper.responseAndActionPlan') || 'Tanggapan & Rencana Aksi' }
])

const validateStep = (idx: number): boolean => {
  if (idx === 0) {
    if (!store.causeForm.condition?.trim()) {
      formErrorMessage.value = t('workingPaper.causeForm.conditionRequired') || 'Kondisi (Dampak di Lapangan) wajib diisi'
      return false
    }
    if (!store.causeForm.criteria?.trim()) {
      formErrorMessage.value = t('workingPaper.causeForm.criteriaRequired') || 'Kriteria (Aturan/SOP Terkait) wajib dipilih'
      return false
    }
    if (!store.causeForm.impact?.trim()) {
      formErrorMessage.value = t('workingPaper.causeForm.impactRequired') || 'Dampak (Risk Impact) wajib diisi'
      return false
    }
    formErrorMessage.value = ''
    return true
  }
  if (idx === 1) {
    if (store.causeForm.rootCause?.length > 0) {
      for (let i = 0; i < store.causeForm.rootCause.length; i++) {
        const rca = store.causeForm.rootCause[i]
        if (!rca.w1?.trim()) {
          formErrorMessage.value = t('workingPaper.causeForm.why1Validation') || 'Why maksimal 100 karakter dan wajib diisi'
          return false
        }
      }
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

const handleSave = () => {
  if (!validateStep(0)) {
    currentStep.value = 0
    return
  }
  if (!validateStep(1)) {
    currentStep.value = 1
    return
  }
  formErrorMessage.value = ''
  store.handleSubmitF04()
}

onMounted(() => {
  store.fetchEmployees?.()
})

watch(
  () => store.showModalF04,
  (open) => {
    if (open) {
      currentStep.value = 0
      formErrorMessage.value = ''
      store.fetchEmployees?.()
    }
  }
)
</script>
