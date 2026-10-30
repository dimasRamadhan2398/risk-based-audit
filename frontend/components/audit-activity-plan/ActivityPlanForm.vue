<template>
  <UModal
    :open="store.isModalOpen"
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-6xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
    @update:open="(open: boolean) => { if (!open) store.closeModal() }"
  >
    <template #content>
      <div class="relative flex max-h-[92vh] flex-col transition-colors duration-300">
        <!-- Header -->
        <div class="flex items-center justify-between border-b border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
          <div>
            <h3 class="text-lg sm:text-xl font-bold text-[var(--text-main)]">
              {{ store.isEditMode ? t('auditActivityPlan.editPlan') : t('auditActivityPlan.createPlanTitle') }}
            </h3>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark-20-solid"
            class="-my-1"
            :aria-label="t('auditActivityPlan.form.cancel')"
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
          class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6"
        >
          <!-- Loading the saved plan -->
          <div
            v-if="store.formLoading"
            class="flex h-40 items-center justify-center gap-2 text-[var(--text-muted)]"
          >
            <UIcon
              name="i-heroicons-arrow-path"
              class="h-5 w-5 animate-spin"
            />
            {{ t('auditActivityPlan.view.loading') }}
          </div>

          <UAlert
            v-else-if="store.formLoadError"
            color="error"
            variant="subtle"
            icon="i-heroicons-exclamation-triangle"
            :description="store.formLoadError"
            :actions="[{ label: t('auditActivityPlan.retry'), color: 'error', variant: 'outline', onClick: () => store.loadEditing() }]"
          />

          <template v-else>
            <!-- Blocking errors of the step / server -->
            <UAlert
              v-if="store.formError?.messages.length"
              color="error"
              variant="solid"
              icon="i-heroicons-exclamation-triangle"
              :title="t('common.error')"
              :actions="store.formError.code === 'ACTIVITY_PLAN_STALE'
                ? [{ label: t('auditActivityPlan.reload'), color: 'neutral', variant: 'solid', onClick: () => store.loadEditing() }]
                : []"
              close
              class="shadow-sm"
              data-testid="form-error"
              @update:open="(open: boolean) => { if (!open) store.formError = null }"
            >
              <template #description>
                <ul class="list-disc space-y-0.5 pl-4">
                  <li
                    v-for="(message, i) in store.formError.messages"
                    :key="i"
                  >
                    {{ message }}
                  </li>
                </ul>
              </template>
            </UAlert>

            <!-- Step 1: header of the form -->
            <div
              v-if="currentStep === 0"
              class="space-y-5"
            >
              <UCard
                :ui="{ body: 'p-4 sm:p-6' }"
                class="border border-[var(--border-main)] bg-transparent ring-0 shadow-none"
              >
                <template #header>
                  <div class="flex items-center gap-2">
                    <UIcon
                      name="i-heroicons-document-text"
                      class="h-5 w-5 text-amber-500"
                    />
                    <h3 class="text-lg font-bold text-[var(--text-main)]">
                      {{ t('auditActivityPlan.stepper.header') }}
                    </h3>
                  </div>
                </template>

                <div class="grid grid-cols-1 gap-5 md:grid-cols-3">
                  <UFormField
                    :label="t('auditActivityPlan.header.annualPlan')"
                    required
                    class="md:col-span-2"
                    :help="store.isEditMode ? t('auditActivityPlan.header.annualPlanLocked') : undefined"
                  >
                    <UInput
                      v-if="store.isEditMode"
                      :model-value="annualPlanLabel"
                      disabled
                      class="w-full"
                      data-testid="annual-plan-readonly"
                    />
                    <USelectMenu
                      v-else
                      :model-value="store.formState.annualPlanId || undefined"
                      :items="store.annualPlanOptions"
                      value-key="value"
                      label-key="label"
                      :loading="store.annualPlansLoading"
                      :placeholder="t('auditActivityPlan.header.annualPlanPlaceholder')"
                      class="w-full"
                      data-testid="annual-plan-select"
                      @update:model-value="(value: string | undefined) => store.setAnnualPlan(value)"
                    >
                      <template #empty>
                        {{ store.annualPlansLoading ? t('common.loading') : t('auditActivityPlan.header.annualPlanEmpty') }}
                      </template>
                    </USelectMenu>
                  </UFormField>

                  <UFormField
                    :label="t('auditActivityPlan.header.year')"
                    :hint="t('auditActivityPlan.form.derivedFromAnnual')"
                  >
                    <UInput
                      :model-value="store.formYear ? String(store.formYear) : ''"
                      disabled
                      class="w-full"
                      data-testid="header-year"
                    />
                  </UFormField>

                  <UAlert
                    v-if="store.annualPlanNotApproved"
                    color="warning"
                    variant="subtle"
                    icon="i-heroicons-information-circle"
                    :description="t('auditActivityPlan.header.annualPlanNotApproved', { status: annualPlanStatusLabel })"
                    class="md:col-span-3"
                    data-testid="annual-plan-not-approved"
                  />

                  <UFormField
                    :label="t('auditActivityPlan.header.activityType')"
                    required
                  >
                    <USelectMenu
                      :model-value="store.formState.activityType || undefined"
                      :items="store.activityTypeOptions"
                      value-key="value"
                      label-key="label"
                      :placeholder="t('auditActivityPlan.header.activityTypePlaceholder')"
                      class="w-full"
                      data-testid="activity-type-select"
                      @update:model-value="(value: ActivityPlanActivityType | undefined) => store.setActivityType(value)"
                    />
                  </UFormField>

                  <UFormField
                    :label="t('auditActivityPlan.header.title')"
                    required
                    class="md:col-span-2"
                  >
                    <UInput
                      v-model="store.formState.title"
                      :placeholder="t('auditActivityPlan.header.titlePlaceholder')"
                      :maxlength="ACTIVITY_PLAN_TITLE_MAX"
                      class="w-full"
                      data-testid="header-title"
                    />
                    <div class="flex justify-end">
                      <span class="text-xs text-gray-500 dark:text-gray-400">{{ store.formState.title.length }}/{{ ACTIVITY_PLAN_TITLE_MAX }}</span>
                    </div>
                  </UFormField>
                </div>
              </UCard>
            </div>

            <!-- Step 2: assignments -->
            <div
              v-else-if="currentStep === 1"
              class="space-y-5"
            >
              <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-clipboard-document-check"
                    class="h-5 w-5 text-amber-500"
                  />
                  <h3 class="text-lg font-bold text-[var(--text-main)]">
                    {{ t('auditActivityPlan.stepper.rows') }}
                  </h3>
                </div>
                <UTooltip
                  :text="canPick ? t('auditActivityPlan.rows.addFromPkat') : t('auditActivityPlan.rows.pickHeaderFirst')"
                >
                  <UButton
                    icon="i-heroicons-plus"
                    color="primary"
                    :label="t('auditActivityPlan.rows.addFromPkat')"
                    :disabled="!canPick"
                    data-testid="add-from-pkat"
                    @click="pickerOpen = true"
                  />
                </UTooltip>
              </div>

              <div
                v-if="!store.formState.rows.length"
                class="rounded-xl border border-dashed border-[var(--border-main)] py-10 text-center text-sm italic text-gray-400"
                data-testid="rows-empty"
              >
                {{ t('auditActivityPlan.rows.empty') }}
              </div>

              <ActivityPlanRowCard
                v-for="(row, index) in store.formState.rows"
                :key="row.key"
                :index="index"
                :total="store.formState.rows.length"
                :highlighted="highlightedRow === index + 1"
              />

              <!-- Totals -->
              <div
                v-if="store.formState.rows.length"
                class="grid grid-cols-1 gap-3 rounded-xl border border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:grid-cols-3"
                data-testid="rows-totals"
              >
                <div class="text-center">
                  <p class="text-xs font-semibold uppercase text-[var(--text-muted)]">
                    {{ t('auditActivityPlan.totals.rows') }}
                  </p>
                  <p class="text-2xl font-bold text-primary-600 dark:text-primary-400">
                    {{ store.formTotals.rowCount }}
                  </p>
                </div>
                <div class="text-center">
                  <p class="text-xs font-semibold uppercase text-[var(--text-muted)]">
                    {{ t('auditActivityPlan.totals.teamSize') }}
                  </p>
                  <p class="text-2xl font-bold text-primary-600 dark:text-primary-400">
                    {{ store.formTotals.teamSize }}
                  </p>
                </div>
                <div class="text-center">
                  <p class="text-xs font-semibold uppercase text-[var(--text-muted)]">
                    {{ t('auditActivityPlan.totals.mandays') }}
                  </p>
                  <p class="text-2xl font-bold text-primary-600 dark:text-primary-400">
                    {{ store.formTotals.totalMandays }}
                  </p>
                </div>
              </div>
            </div>

            <!-- Step 3: attachments -->
            <div
              v-else-if="currentStep === 2"
              class="space-y-6"
            >
              <UCard
                :ui="{ body: 'p-4 sm:p-6' }"
                class="border border-[var(--border-main)] bg-transparent ring-0 shadow-none"
              >
                <template #header>
                  <div class="flex items-center gap-2">
                    <UIcon
                      name="i-heroicons-paper-clip"
                      class="h-5 w-5 text-amber-500"
                    />
                    <h3 class="text-lg font-bold text-[var(--text-main)]">
                      {{ t('auditActivityPlan.form.attachment') }}
                    </h3>
                  </div>
                </template>
                <div class="space-y-4">
                  <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
                    <UFormField :label="t('auditActivityPlan.form.attachmentCategory')">
                      <USelectMenu
                        v-model="store.formState.attachmentCategory"
                        :items="store.attachmentCategoryOptions"
                        class="w-full"
                      />
                    </UFormField>
                  </div>
                  <div
                    v-if="store.formState.attachments.length"
                    class="space-y-2"
                  >
                    <p class="text-sm font-semibold text-[var(--text-main)]">
                      {{ t('auditActivityPlan.form.attachedFiles') }}
                    </p>
                    <ul class="space-y-1">
                      <li
                        v-for="(file, index) in store.formState.attachments"
                        :key="`${file.url}-${index}`"
                        class="flex items-center gap-2 text-sm text-[var(--text-muted)]"
                      >
                        <UIcon
                          name="i-heroicons-document-text"
                          class="h-4 w-4 shrink-0"
                        />
                        <span class="truncate">{{ file.name }}</span>
                        <span class="shrink-0 text-xs">{{ file.size }}</span>
                      </li>
                    </ul>
                  </div>
                  <UFormField
                    :label="t('auditActivityPlan.form.uploadAttachmentHere')"
                    size="lg"
                    class="pt-2"
                  >
                    <UFileUpload
                      v-model="store.formState.file"
                      layout="list"
                      multiple
                      :label="t('auditActivityPlan.form.dropAttachmentsHere')"
                      :description="t('auditActivityPlan.form.maxFileSizeDesc')"
                      class="w-full"
                      :ui="{ base: 'min-h-48' }"
                    />
                  </UFormField>
                </div>
              </UCard>
            </div>
          </template>
        </div>

        <!-- Footer -->
        <div class="flex flex-col-reverse items-stretch justify-between gap-3 border-t border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:flex-row sm:items-center">
          <div class="flex flex-col-reverse gap-2 sm:flex-row">
            <UButton
              :label="t('auditActivityPlan.form.cancel')"
              color="neutral"
              variant="ghost"
              class="w-full sm:w-auto"
              @click="store.closeModal"
            />
            <UButton
              icon="i-heroicons-eye"
              :label="t('auditActivityPlan.preview')"
              color="neutral"
              variant="outline"
              class="w-full sm:w-auto"
              :disabled="store.formLoading || !!store.formLoadError"
              data-testid="open-preview"
              @click="previewOpen = true"
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
              :disabled="store.formLoading || !!store.formLoadError"
              @click="handleNext"
            />
            <UButton
              v-else
              icon="i-heroicons-check"
              :label="store.isEditMode ? t('auditActivityPlan.form.update') : t('auditActivityPlan.form.create')"
              color="primary"
              class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
              :loading="store.saving"
              :disabled="store.formLoading || !!store.formLoadError"
              data-testid="form-submit"
              @click="handleFormSubmit"
            />
          </div>
        </div>
      </div>

      <AnnualActivityPicker v-model:open="pickerOpen" />

      <!-- Preview of the paper form -->
      <UModal
        v-model:open="previewOpen"
        :ui="{
          content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-7xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
          overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-sm'
        }"
      >
        <template #content>
          <div class="flex max-h-[90vh] flex-col">
            <div class="flex items-center justify-between border-b border-[var(--border-main)] bg-[var(--bg-surface)] p-4">
              <h3 class="text-lg font-bold text-[var(--text-main)]">
                {{ t('auditActivityPlan.preview') }}
              </h3>
              <UButton
                color="neutral"
                variant="ghost"
                icon="i-heroicons-x-mark-20-solid"
                :aria-label="t('auditActivityPlan.view.close')"
                @click="previewOpen = false"
              />
            </div>
            <div class="flex-1 overflow-y-auto p-4 sm:p-6">
              <ActivityPlanFormSheet :sheet="previewSheet" />
            </div>
          </div>
        </template>
      </UModal>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useActivityPlanStore } from '~/stores/activity-plan'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { useI18n } from '~/composables/useI18n'
import type { ActivityPlanActivityType } from '~/types/audit'
import { ACTIVITY_PLAN_TITLE_MAX, sheetFromForm } from '~/utils/activityPlanForm'
import ActivityPlanRowCard from '~/components/audit-activity-plan/ActivityPlanRowCard.vue'
import AnnualActivityPicker from '~/components/audit-activity-plan/AnnualActivityPicker.vue'
import ActivityPlanFormSheet from '~/components/audit-activity-plan/ActivityPlanFormSheet.vue'

/**
 * Formulir Rencana Kegiatan Audit: 1) header (PKAT, Jenis Kegiatan Audit, Judul), 2) assignments
 * picked from the PKAT, 3) attachments; plus a preview of the paper layout.
 */
const { t } = useI18n()
const store = useActivityPlanStore()
const annualPlanStore = useAnnualPlanStore()

const currentStep = ref(0)
const pickerOpen = ref(false)
const previewOpen = ref(false)
const highlightedRow = ref<number | null>(null)
const bodyRef = ref<HTMLElement | null>(null)

const steps = computed(() => [
  { key: 'header', label: t('auditActivityPlan.stepper.header') },
  { key: 'rows', label: t('auditActivityPlan.stepper.rows') },
  { key: 'attachments', label: t('auditActivityPlan.stepper.attachments') }
])

const canPick = computed(() => !!store.formState.annualPlanId && !!store.formState.activityType)

const annualPlanStatusLabel = computed(() => annualPlanStore.getStatusLabel(store.selectedAnnualPlan?.status))

const annualPlanLabel = computed(() => {
  const plan = store.selectedAnnualPlan
  if (!plan) return ''
  return [plan.year || '', plan.code, plan.title, annualPlanStatusLabel.value].filter(Boolean).join(' · ')
})

const previewSheet = computed(() => sheetFromForm(store.formState, store.formYear))

watch(() => store.isModalOpen, (isOpen) => {
  if (isOpen) {
    currentStep.value = 0
    highlightedRow.value = null
    pickerOpen.value = false
    previewOpen.value = false
  }
})

// A blocking error (client or server) brings its step forward and scrolls to its row.
watch(() => store.formError, async (error) => {
  highlightedRow.value = null
  if (!error) return
  if (error.step !== undefined) currentStep.value = error.step
  if (error.row && error.step === 1) {
    highlightedRow.value = error.row
    await nextTick()
    bodyRef.value?.querySelector(`#activity-plan-row-${error.row}`)?.scrollIntoView?.({ behavior: 'smooth', block: 'start' })
  } else {
    bodyRef.value?.scrollTo?.({ top: 0, behavior: 'smooth' })
  }
})

const validateUpTo = (targetIdx: number): boolean => {
  for (let i = 0; i < targetIdx; i++) {
    const error = store.validateStep(i)
    if (error) {
      store.formError = error
      return false
    }
  }
  return true
}

const goToStep = (targetIdx: number) => {
  if (store.formLoading || store.formLoadError) return
  if (targetIdx <= currentStep.value) {
    currentStep.value = targetIdx
    return
  }
  if (!validateUpTo(targetIdx)) return
  store.formError = null
  currentStep.value = targetIdx
}

const handleNext = () => goToStep(currentStep.value + 1)

const handleFormSubmit = () => {
  store.savePlan()
}
</script>
