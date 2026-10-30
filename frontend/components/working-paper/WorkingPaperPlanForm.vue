<template>
  <UModal 
    :open="store.showModalF05" 
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
    @update:open="(open: boolean) => { if (!open) store.closeModalF05() }"
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
                {{ store.isEditingF05 ? (t('workingPaper.planForm.editTitle') || t('workingPaper.planForm.title')) : t('workingPaper.planForm.title') }}
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
            @click="store.closeModalF05"
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

          <!-- Step 1: Auditor Recommendation sampai Auditee Response -->
          <div v-show="currentStep === 0" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-chat-bubble-bottom-center-text"
                    class="h-5 w-5 text-primary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('workingPaper.planForm.stepper.recommendationAndResponse') }}
                  </h3>
                </div>
              </template>

              <!-- Auditor Recommendation (Solution) -->
              <UFormField 
                :label="t('workingPaper.planForm.recommendation')" 
                name="recommendation" 
                required
              >
                <UTextarea 
                  v-model="store.planForm.recommendation" 
                  :rows="4" 
                  :placeholder="t('workingPaper.planForm.recommendationPlaceholder')" 
                  class="w-full" 
                />
              </UFormField>

              <!-- Auditee Response (Management Response) -->
              <UFormField 
                :label="t('workingPaper.planForm.response')" 
                name="response" 
                required
              >
                <UTextarea 
                  v-model="store.planForm.response" 
                  :rows="4" 
                  :placeholder="t('workingPaper.planForm.responsePlaceholder')" 
                  class="w-full" 
                />
              </UFormField>
            </UCard>
          </div>

          <!-- Step 2: Action Plan Details (Action Description sampai Planning Period) -->
          <div v-show="currentStep === 1" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-clipboard-document-list"
                    class="h-5 w-5 text-secondary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('workingPaper.planForm.actionPlanDetail') }}
                  </h3>
                </div>
              </template>

              <!-- Action Description -->
              <UFormField 
                :label="t('workingPaper.planForm.actionDescription')" 
                name="actionDescription" 
                required
              >
                <UTextarea 
                  v-model="store.planForm.actionDescription" 
                  :rows="3"
                  :placeholder="t('workingPaper.planForm.actionDescriptionPlaceholder')" 
                  class="w-full" 
                />
              </UFormField>

              <!-- PIC (Person In Charge) -->
              <UFormField 
                :label="t('workingPaper.planForm.pic')" 
                name="pic" 
                required
              >
                <USelectMenu 
                  v-model="store.planForm.pic" 
                  icon="i-heroicons-magnifying-glass" 
                  :items="store.picOptions" 
                  :loading="store.employeesLoading"
                  :placeholder="t('workingPaper.planForm.picPlaceholder')" 
                  class="w-full" 
                  create-item
                  @create="(val: string) => { store.planForm.pic = val.trim() }"
                >
                  <template #empty>
                    {{ store.employeesLoading ? t('common.loading') : (store.employeesError || t('masterData.employee.empty')) }}
                  </template>
                </USelectMenu>
              </UFormField>

              <!-- Planning Period -->
              <UFormField 
                :label="t('workingPaper.planForm.periodAction')" 
                name="periodAction" 
                required
              >
                <AppDatePicker 
                  v-model="store.planForm.periodAction" 
                  class="w-full" 
                />
              </UFormField>
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
              @click="store.closeModalF05"
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
              :label="store.isEditingF05 ? t('common.updateData') : t('common.submit')"
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
import AppDatePicker from '~/components/shared/AppDatePicker.vue'

const { t } = useI18n()
const store = useWorkingPaperStore()

const currentStep = ref(0)
const bodyRef = ref<HTMLElement | null>(null)
const formErrorMessage = ref('')

const steps = computed(() => [
  { key: 'recommendationAndResponse', label: t('workingPaper.planForm.stepper.recommendationAndResponse') || 'Rekomendasi & Tanggapan' },
  { key: 'actionPlanDetails', label: t('workingPaper.planForm.stepper.actionPlanDetails') || 'Detail Rencana Aksi' }
])

const validateStep = (idx: number): boolean => {
  if (idx === 0) {
    if (!store.planForm.recommendation?.trim()) {
      formErrorMessage.value = t('workingPaper.planForm.recommendationRequired') || 'Rekomendasi Auditor (Solusi) wajib diisi'
      return false
    }
    if (!store.planForm.response?.trim()) {
      formErrorMessage.value = t('workingPaper.planForm.responseRequired') || 'Tanggapan Auditi (Tanggapan Manajemen) wajib diisi'
      return false
    }
    formErrorMessage.value = ''
    return true
  }
  if (idx === 1) {
    if (!store.planForm.actionDescription?.trim()) {
      formErrorMessage.value = t('workingPaper.planForm.actionDescriptionRequired') || 'Deskripsi Aksi wajib diisi'
      return false
    }
    if (!store.planForm.pic?.trim()) {
      formErrorMessage.value = t('workingPaper.planForm.picRequired') || 'PIC (Penanggung Jawab) wajib dipilih'
      return false
    }
    if (!store.planForm.periodAction?.trim()) {
      formErrorMessage.value = t('workingPaper.planForm.periodActionRequired') || 'Periode Perencanaan wajib dipilih'
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
  store.handleSubmitF05()
}

onMounted(() => {
  store.fetchEmployees?.()
})

watch(
  () => store.showModalF05,
  (open) => {
    if (open) {
      currentStep.value = 0
      formErrorMessage.value = ''
      store.fetchEmployees?.()
    }
  }
)
</script>
