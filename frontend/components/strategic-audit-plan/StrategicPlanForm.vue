<template>
  <!-- Add/Edit Modal: a 2-step wizard laid out like the audit activity plan form -->
  <UModal
    :open="store.isAddModalOpen"
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
    @update:open="(open: boolean) => { if (!open) store.closeModal() }"
  >
    <template #content>
      <div class="relative flex max-h-[90vh] flex-col transition-colors duration-300">
        <!-- Header -->
        <div class="flex items-center justify-between border-b border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
          <h3 class="text-lg sm:text-xl font-bold text-[var(--text-main)]">
            {{ t('strategicPlan.form.title') }}
          </h3>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark-20-solid"
            class="-my-1"
            :aria-label="t('strategicPlan.form.cancel')"
            @click="store.closeModal"
          />
        </div>

        <!-- Stepper -->
        <div class="w-full overflow-x-auto border-b border-[var(--border-main)] bg-[var(--bg-surface)] px-4 py-3 sm:px-6 sm:py-4 scrollbar-thin">
          <div class="flex min-w-max items-center justify-center gap-3 px-1 sm:gap-4">
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
                class="h-0.5 w-5 shrink-0 rounded transition-all duration-300 sm:w-8"
                :class="currentStep > idx ? 'bg-secondary' : 'bg-gray-200 dark:bg-gray-800'"
              />
            </template>
          </div>
        </div>

        <!-- Body -->
        <div
          ref="bodyRef"
          class="flex-1 overflow-y-auto p-4 sm:p-6 bg-white dark:bg-gray-900 text-gray-900 dark:text-white"
        >
          <UForm
            :state="store.form"
            @submit="onFormSubmit"
          >
            <div class="space-y-5">
              <!-- Error of the last failed save; the form keeps the input so it can be retried -->
              <UAlert
                v-if="store.formError"
                color="error"
                variant="subtle"
                icon="i-lucide-alert-triangle"
                :title="t('strategicPlan.toast.saveFailedTitle')"
                :description="store.formError"
              />

              <!-- Step 1: objective, KPI, unit and category -->
              <template v-if="currentStep === 0">
                <!-- Strategic Objective (no goal picker: the store links a new plan to the single corporate goal on save) -->
                <UFormField
                  :label="t('strategicPlan.form.objective')"
                  :error="stepErrors.strategicObjective"
                  required
                >
                  <UTextarea
                    v-model="store.form.strategicObjective"
                    :placeholder="t('strategicPlan.form.objectivePlaceholder')"
                    :rows="2"
                    class="w-full"
                    data-testid="objective-input"
                  />
                </UFormField>

                <!-- KPI Title -->
                <UFormField
                  :label="t('strategicPlan.form.kpi')"
                  required
                >
                  <UTextarea
                    v-model="store.form.kpi"
                    :placeholder="t('strategicPlan.form.kpiPlaceholder')"
                    :rows="2"
                    class="w-full"
                  />
                </UFormField>

                <!-- Unit -->
                <UFormField
                  :label="t('strategicPlan.form.unit')"
                  required
                >
                  <USelectMenu
                    v-model="store.form.unit"
                    :items="store.unitOptions"
                    value-key="value"
                    :placeholder="t('strategicPlan.form.selectUnit')"
                    class="w-full"
                  />
                </UFormField>

                <!-- Category (optional): the backend accepts only these values or none -->
                <UFormField
                  :label="t('strategicPlan.form.category')"
                  :error="store.formFieldErrors.category"
                >
                  <USelectMenu
                    v-model="category"
                    :items="categoryOptions"
                    value-key="value"
                    :placeholder="t('strategicPlan.form.selectCategory')"
                    :search-input="false"
                    clear
                    class="w-full"
                  />
                </UFormField>
              </template>

              <!-- Step 2: direction, period, target & realization matrix and status -->
              <template v-else>
                <!-- HIB/HIG Radio -->
                <div class="form-row">
                  <UFormField
                    :label="t('strategicPlan.form.hibHig')"
                    required
                  >
                    <URadioGroup
                      v-model="store.form.hibHig"
                      :items="[
                        { label: t('strategicPlan.form.higLabel'), value: 'HIG' },
                        { label: t('strategicPlan.form.hibLabel'), value: 'HIB' }
                      ]"
                      color="primary"
                    />
                  </UFormField>

                  <!-- Period Type Radio -->
                  <UFormField
                    :label="t('strategicPlan.form.periodType')"
                    required
                  >
                    <URadioGroup
                      v-model="store.form.periodType"
                      :items="[
                        { label: t('strategicPlan.form.quartal'), value: 'Quartal' },
                        { label: t('strategicPlan.form.yearly'), value: 'Yearly' }
                      ]"
                      color="primary"
                    />
                  </UFormField>
                </div>

                <!-- Periode: Dari Tahun / Sampai Tahun -->
                <div class="form-row">
                  <label class="form-label text-gray-700 dark:text-white">{{ t('strategicPlan.form.period') }}</label>
                  <div class="flex items-center gap-3 flex-wrap">
                    <span class="text-sm font-bold text-gray-800 dark:text-white">{{ t('strategicPlan.form.from') }}</span>
                    <USelectMenu
                      v-model="store.form.yearStart"
                      :items="store.yearOptions"
                      value-key="value"
                      :placeholder="t('strategicPlan.vmg.year')"
                      class="w-28"
                    />
                    <span class="text-sm font-bold text-gray-800 dark:text-white">{{ t('strategicPlan.form.to') }}</span>
                    <USelectMenu
                      v-model="store.form.yearEnd"
                      :items="store.yearEndOptions"
                      value-key="value"
                      :placeholder="t('strategicPlan.vmg.year')"
                      class="w-28"
                    />
                  </div>
                </div>

                <!-- Flexible Target & Realisasi KPI Matrix Component -->
                <TargetRealizationMatrix />

                <!-- Status (Read Only) -->
                <UFormField
                  :label="t('strategicPlan.form.status')"
                  required
                >
                  <UInput
                    :value="formatStatus(store.computedStatus)"
                    disabled
                  />
                </UFormField>
              </template>
            </div>
          </UForm>
        </div>

        <!-- Footer -->
        <div class="flex flex-col-reverse items-stretch justify-between gap-3 border-t border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:flex-row sm:items-center">
          <UButton
            :label="t('strategicPlan.form.cancel')"
            color="neutral"
            variant="ghost"
            class="w-full sm:w-auto"
            @click="store.closeModal"
          />

          <div class="flex w-full items-center justify-end gap-2 sm:w-auto sm:gap-3">
            <UButton
              v-if="currentStep > 0"
              icon="i-heroicons-arrow-left"
              :label="t('strategicPlan.form.stepper.previous')"
              color="neutral"
              variant="outline"
              class="flex-1 sm:flex-none"
              data-testid="form-previous"
              @click="() => { currentStep-- }"
            />
            <UButton
              v-if="currentStep < steps.length - 1"
              trailing-icon="i-heroicons-arrow-right"
              :label="t('strategicPlan.form.stepper.next')"
              color="primary"
              class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
              data-testid="form-next"
              @click="handleNext"
            />
            <UButton
              v-else
              icon="i-heroicons-check"
              :label="store.isEditMode ? t('strategicPlan.form.update') : t('strategicPlan.form.submit')"
              color="primary"
              class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
              :loading="store.saving"
              :disabled="store.saving"
              data-testid="form-submit"
              @click="handleFormSubmit"
            />
          </div>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useI18n } from '~/composables/useI18n'
import { kpiValueLabel } from '~/utils/kpiPerformanceLabels'
import { STRATEGIC_PLAN_CATEGORIES } from '~/utils/strategicPlanPayload'
import TargetRealizationMatrix from './TargetRealizationMatrix.vue'

const { t } = useI18n()
const store = useStrategicPlanStore()

// Raw value is what is saved; the label comes from the KPI table's category labels.
const categoryOptions = computed(() =>
  STRATEGIC_PLAN_CATEGORIES.map((value): { label: string, value: string } => ({ label: kpiValueLabel(t, 'categories', value), value }))
)
// "" (no category) shows the placeholder; clearing the menu stores "" again.
const category = computed<string | undefined>({
  get: () => store.form.category || undefined,
  set: (value) => {
    store.form.category = value ?? ''
    if (store.formFieldErrors.category) store.formFieldErrors = {}
  }
})

/**
 * Two steps: 1) objective, KPI, unit, category; 2) HIB/HIG, period, target & realization matrix, status.
 * Next checks only the step-1 rules; saving runs them again and then the store's own checks (period, numbers).
 */
const currentStep = ref(0)
const bodyRef = ref<HTMLElement | null>(null)
const stepErrors = ref<Partial<Record<'strategicObjective', string>>>({})

const steps = computed(() => [
  { key: 'objective', label: t('strategicPlan.form.stepper.objective') },
  { key: 'targets', label: t('strategicPlan.form.stepper.targets') }
])

// Same rule the store applies on save (the objective is the only required step-1 value it enforces).
const validateStep = (idx: number): boolean => {
  if (idx !== 0) return true
  const errors: typeof stepErrors.value = {}
  if (!store.form.strategicObjective?.trim()) errors.strategicObjective = t('strategicPlan.toast.objectiveRequired')
  stepErrors.value = errors
  return Object.keys(errors).length === 0
}

const showStep = (idx: number) => {
  currentStep.value = idx
  bodyRef.value?.scrollTo?.({ top: 0, behavior: 'smooth' })
}

const goToStep = (targetIdx: number) => {
  if (targetIdx <= currentStep.value) {
    currentStep.value = targetIdx
    return
  }
  for (let i = 0; i < targetIdx; i++) {
    if (!validateStep(i)) {
      showStep(i)
      return
    }
  }
  showStep(targetIdx)
}

const handleNext = () => goToStep(currentStep.value + 1)

const handleFormSubmit = async () => {
  if (store.saving) return
  for (let i = 0; i < steps.value.length; i++) {
    if (!validateStep(i)) {
      showStep(i)
      return
    }
  }
  await store.handleSubmit()
}

// Enter inside the form moves forward on step 1 instead of saving past it.
const onFormSubmit = () => {
  if (currentStep.value < steps.value.length - 1) handleNext()
  else handleFormSubmit()
}

// Fixing the objective clears its message.
watch(() => store.form.strategicObjective, (value) => {
  if (stepErrors.value.strategicObjective && value?.trim()) stepErrors.value = {}
})

// A rejected category (server-side) belongs to step 1: bring it forward.
watch(() => store.formFieldErrors.category, (error) => {
  if (error) showStep(0)
})

// Every open, close or create/edit switch starts at step 1 without old messages.
watch(() => [store.isAddModalOpen, store.isEditMode], () => {
  currentStep.value = 0
  stepErrors.value = {}
})

const formatStatus = (status: string) => {
  if (!status) return '-'
  const lower = status.toLowerCase()
  if (['good', 'moderate', 'poor', 'pending', 'planned'].includes(lower)) {
    return t(`strategicPlan.status.${lower}`)
  }
  return status
}
</script>

<style scoped>
.form-row {
  display: grid;
  grid-template-columns: 140px 1fr;
  align-items: start;
  gap: 1rem;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 600;
  padding-top: 0.5rem;
}

html.dark .form-label,
.dark .form-label {
  color: #ffffff !important;
}

.readonly-field {
  background-color: #f3f4f6;
  border-radius: 0.375rem;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  color: #111827;
  font-weight: 600;
  min-height: 2.25rem;
  display: flex;
  align-items: center;
}

html.dark .readonly-field,
.dark .readonly-field {
  background-color: #1f2937 !important;
  color: #ffffff !important;
  border: 1px solid #374151;
}

.period-tab {
  padding: 0.375rem 1.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.15s ease;
  cursor: pointer;
  border: none;
  outline: none;
}

.period-tab-active {
  border: 2px solid #f97316;
  color: #f97316;
  background-color: white;
}

html.dark .period-tab-active,
.dark .period-tab-active {
  background-color: #1f2937 !important;
  color: #f97316 !important;
}

.period-tab-inactive {
  color: #6b7280;
  background-color: white;
}

html.dark .period-tab-inactive,
.dark .period-tab-inactive {
  color: #ffffff !important;
  background-color: #111827 !important;
}

.period-tab-inactive:hover {
  background-color: #f9fafb;
}

html.dark .period-tab-inactive:hover,
.dark .period-tab-inactive:hover {
  background-color: #1f2937 !important;
}
</style>
