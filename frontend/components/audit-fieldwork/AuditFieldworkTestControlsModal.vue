<template>
  <!-- Test Control Modal (View / Add / Edit) -->
  <UModal 
    :open="store.showTestControlModal"
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
    @update:open="(open: boolean) => { if (!open) store.showTestControlModal = false }"
  >
    <template #content>
      <div class="relative flex flex-col max-h-[92vh] transition-colors duration-300">
        <!-- Modal Header -->
        <div class="flex items-center justify-between p-4 sm:p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
          <div class="flex items-center gap-3">
            <div
              class="p-2 rounded-lg"
              :class="store.isReadOnlyTestControl ? 'bg-blue-500/10 text-blue-500' : 'bg-primary-500/10 text-primary-500'"
            >
              <UIcon
                :name="store.isReadOnlyTestControl ? 'i-heroicons-eye' : (store.isEditingTestControl ? 'i-heroicons-pencil-square' : 'i-heroicons-shield-check')"
                class="size-6"
              />
            </div>
            <div>
              <h3 class="text-lg sm:text-xl font-bold text-[var(--text-main)]">
                {{ store.isReadOnlyTestControl ? (t('auditFieldwork.testControls.modalView') || 'Detail Pengujian Pengendalian') : (store.isEditingTestControl ? t('auditFieldwork.testControls.modalEdit') : t('auditFieldwork.testControls.modalAdd')) }}
              </h3>
              <p class="mt-0.5 text-xs text-[var(--text-muted)]">
                <template v-if="store.isReadOnlyTestControl">
                  {{ t('auditExecution.detailModal.sources.assignment_letter.label') }}: <span class="font-semibold text-primary-600 dark:text-primary-400">{{ store.selectedAssignmentLetter || '-' }}</span>
                </template>
                <template v-else>
                  {{ steps[currentStep]?.label }} ({{ t('auditActivityPlan.form.stepNum', { current: currentStep + 1, total: steps.length }) }})
                  <span class="mx-1.5 opacity-60">·</span>
                  {{ store.selectedAssignmentLetter || '-' }}
                </template>
              </p>
            </div>
          </div>
          <UButton
            icon="i-heroicons-x-mark-20-solid"
            color="neutral"
            variant="ghost"
            class="-my-1"
            :aria-label="t('common.close')"
            @click="store.showTestControlModal = false"
          />
        </div>

        <!-- Stepper (Only in Form Mode) -->
        <div
          v-if="!store.isReadOnlyTestControl"
          class="w-full overflow-x-auto border-b border-[var(--border-main)] bg-[var(--bg-surface)] px-4 py-3 sm:px-6 sm:py-4 scrollbar-thin"
        >
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
                class="h-0.5 w-6 shrink-0 rounded transition-all duration-300 sm:w-10"
                :class="currentStep > idx ? 'bg-secondary' : 'bg-gray-200 dark:bg-gray-800'"
              />
            </template>
          </div>
        </div>

        <!-- Read-Only Detail View -->
        <div v-if="store.isReadOnlyTestControl" class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-4">
          <!-- Header Info Card: Assignment Letter, Due Date, and Test Result -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div class="flex items-center gap-3">
                <div class="p-2.5 rounded-lg bg-primary-500/10 text-primary-500">
                  <UIcon name="i-heroicons-document-text" class="w-6 h-6 text-primary-500" />
                </div>
                <div>
                  <p class="text-xs font-semibold tracking-wider text-[var(--text-muted)]">{{ t('auditExecution.detailModal.sources.assignment_letter.label') }}</p>
                  <p class="text-sm font-bold text-[var(--text-main)]">{{ store.selectedAssignmentLetter || '-' }}</p>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <span v-if="store.testControlForm.testResult" class="inline-flex items-center px-2 py-1 rounded-md text-xs font-semibold" :class="getControlEffectivenessColorClass(store.testControlForm.testResult)">
                  {{ store.testControlForm.testResult }}
                </span>
                <UBadge color="warning" variant="subtle" size="md" class="font-semibold">
                  <UIcon name="i-heroicons-calendar" class="w-4 h-4 mr-1.5" />
                  {{ formatDate(store.testControlForm.dueDate) || '-' }}
                </UBadge>
              </div>
            </div>
          </UCard>

          <!-- Control Information Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                  <UIcon name="i-heroicons-shield-check" class="w-4 h-4 text-primary-500" />
                  <span>{{ t('auditFieldwork.testControls.sectionControl') }}</span>
                </div>
                <UBadge :color="getControlTypeColor(store.testControlForm.controlType)" variant="subtle" size="sm" class="font-medium">
                  {{ store.testControlForm.controlType || '-' }}
                </UBadge>
              </div>
            </template>
            <div class="space-y-3">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-1">{{ t('auditFieldwork.testControls.name') }}</p>
                <p class="text-base font-bold text-[var(--text-main)]">{{ store.testControlForm.controlName || '-' }}</p>
              </div>
              <div>
                <p class="text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-1">{{ t('auditFieldwork.testControls.unitDivision') }}</p>
                <p class="text-sm font-bold text-[var(--text-main)]">{{ store.testControlForm.unitDivision || '-' }}</p>
              </div>
              <div>
                <p class="text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-1">{{ t('auditFieldwork.testControls.description') }}</p>
                <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
                  {{ store.testControlForm.controlDescription || '-' }}
                </p>
              </div>
            </div>
          </UCard>

          <!-- Test Procedure Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                <UIcon name="i-heroicons-clipboard-document-check" class="w-4 h-4 text-primary-500" />
                <span>{{ t('auditFieldwork.testControls.sectionProcedure') }}</span>
              </div>
            </template>
            <div class="space-y-3">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-1">{{ t('auditFieldwork.testControls.steps') }}</p>
                <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
                  {{ store.testControlForm.testProcedure || '-' }}
                </p>
              </div>
              <div class="flex items-center gap-2 pt-1">
                <span class="text-xs font-semibold text-[var(--text-muted)]">{{ t('auditFieldwork.testControls.result') }}:</span>
                <span class="inline-flex items-center px-1.5 py-1 rounded-md text-[10px]/3 font-medium" :class="getControlEffectivenessColorClass(store.testControlForm.testResult)">
                  {{ store.testControlForm.testResult || '-' }}
                </span>
              </div>
            </div>
          </UCard>

          <!-- Finding and Recommendation Card -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Finding -->
            <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
              <template #header>
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-amber-500">
                  <UIcon name="i-heroicons-exclamation-triangle" class="w-4 h-4 text-amber-500" />
                  <span>{{ t('auditFieldwork.testControls.finding') }}</span>
                </div>
              </template>
              <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
                {{ store.testControlForm.finding || '-' }}
              </p>
            </UCard>

            <!-- Recommendation -->
            <UCard color="primary" variant="outline" class="border border-primary-500/20 shadow-xs">
              <template #header>
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                  <UIcon name="i-heroicons-light-bulb" class="w-4 h-4 text-primary-500" />
                  <span>{{ t('auditFieldwork.testControls.recommendation') }}</span>
                </div>
              </template>
              <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
                {{ store.testControlForm.recommendation || '-' }}
              </p>
            </UCard>
          </div>

          <!-- Mitigation Plan Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                  <UIcon name="i-heroicons-wrench-screwdriver" class="w-4 h-4 text-primary-500" />
                  <span>{{ t('auditFieldwork.testControls.sectionMitigation') }}</span>
                </div>
                <div v-if="store.testControlForm.pic" class="flex items-center gap-1.5 text-xs text-[var(--text-muted)]">
                  <UIcon name="i-heroicons-user-circle" class="w-4 h-4" />
                  <span class="font-medium text-[var(--text-main)]">{{ t('auditFieldwork.extra.pic') }} {{ store.testControlForm.pic }}</span>
                </div>
              </div>
            </template>
            <div class="space-y-3">
              <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
                {{ store.testControlForm.mitigationPlan || '-' }}
              </p>
              <div class="flex items-center gap-4 text-xs text-[var(--text-muted)] pt-1">
                <span>{{ t('auditFieldwork.extra.pic') }} <strong class="text-[var(--text-main)]">{{ store.testControlForm.pic || '-' }}</strong></span>
                <span>{{ t('auditFieldwork.extra.dueDate') }} <strong class="text-[var(--text-main)]">{{ formatDate(store.testControlForm.dueDate) || '-' }}</strong></span>
              </div>
            </div>
          </UCard>
        </div>

        <!-- Add / Edit Form View (Multi-Step Stepper) -->
        <div v-else ref="bodyRef" class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
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

          <!-- Step 1: Control Information -->
          <div v-show="currentStep === 0" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-4' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-shield-check"
                    class="h-5 w-5 text-primary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('auditFieldwork.testControls.sectionControl') }}
                  </h3>
                </div>
              </template>

              <div class="space-y-4">
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <UFormField :label="t('auditFieldwork.testControls.name')" required>
                    <UInput
                      v-model="store.testControlForm.controlName"
                      :placeholder="t('auditFieldwork.testControls.namePlaceholder')"
                      class="w-full"
                      required
                    />
                  </UFormField>
                  <UFormField :label="t('auditFieldwork.testControls.type')" required>
                    <ReusableSelectMenu
                      v-model="store.testControlForm.controlType"
                      :items="store.options.controlTypes"
                      :placeholder="t('auditFieldwork.testControls.typePlaceholder')"
                      class="w-full"
                      required
                    />
                  </UFormField>
                  <UFormField :label="t('auditFieldwork.testControls.unitDivision')">
                    <ReusableSelectMenu
                      v-model="store.testControlForm.unitDivision"
                      :items="store.options.divisionOptions"
                      :placeholder="t('auditFieldwork.testControls.unitDivisionPlaceholder')"
                      class="w-full"
                    />
                  </UFormField>
                </div>
                <UFormField :label="t('auditFieldwork.testControls.description')" required>
                  <UTextarea
                    v-model="store.testControlForm.controlDescription"
                    :placeholder="t('auditFieldwork.testControls.descriptionPlaceholder')"
                    class="w-full"
                    :rows="3"
                    required
                  />
                </UFormField>
              </div>
            </UCard>
          </div>

          <!-- Step 2: Test Procedure -->
          <div v-show="currentStep === 1" class="space-y-5">
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
                    {{ t('auditFieldwork.testControls.sectionProcedure') }}
                  </h3>
                </div>
              </template>

              <div class="space-y-4">
                <UFormField :label="t('auditFieldwork.testControls.steps')" required>
                  <UTextarea
                    v-model="store.testControlForm.testProcedure"
                    :placeholder="t('auditFieldwork.testControls.stepsPlaceholder')"
                    class="w-full"
                    :rows="3"
                    required
                  />
                </UFormField>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <UFormField :label="t('auditFieldwork.testControls.result')" required>
                    <ReusableSelectMenu
                      v-model="store.testControlForm.testResult"
                      :items="store.options.testResults"
                      :placeholder="t('auditFieldwork.testControls.resultPlaceholder')"
                      class="w-full"
                      required
                    />
                  </UFormField>
                </div>
              </div>
            </UCard>
          </div>

          <!-- Step 3: Finding and Recommendation -->
          <div v-show="currentStep === 2" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-4' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-light-bulb"
                    class="h-5 w-5 text-primary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('auditFieldwork.testControls.sectionFinding') }}
                  </h3>
                </div>
              </template>

              <div class="space-y-4">
                <UFormField :label="t('auditFieldwork.testControls.finding')">
                  <UTextarea
                    v-model="store.testControlForm.finding"
                    :placeholder="t('auditFieldwork.testControls.findingPlaceholder')"
                    class="w-full"
                    :rows="3"
                  />
                </UFormField>
                <UFormField :label="t('auditFieldwork.testControls.recommendation')">
                  <UTextarea
                    v-model="store.testControlForm.recommendation"
                    :placeholder="t('auditFieldwork.testControls.recommendationPlaceholder')"
                    class="w-full"
                    :rows="3"
                  />
                </UFormField>
              </div>
            </UCard>
          </div>

          <!-- Step 4: Mitigation Plan -->
          <div v-show="currentStep === 3" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-4' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-wrench-screwdriver"
                    class="h-5 w-5 text-primary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('auditFieldwork.testControls.sectionMitigation') }}
                  </h3>
                </div>
              </template>

              <div class="space-y-4">
                <UFormField :label="t('auditFieldwork.testControls.mitigation')">
                  <UTextarea
                    v-model="store.testControlForm.mitigationPlan"
                    :placeholder="t('auditFieldwork.testControls.mitigationPlaceholder')"
                    class="w-full"
                    :rows="3"
                  />
                </UFormField>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <UFormField :label="t('auditFieldwork.testControls.pic')">
                    <UInput
                      v-model="store.testControlForm.pic"
                      :placeholder="t('auditFieldwork.testControls.picPlaceholder')"
                      class="w-full"
                    />
                  </UFormField>
                  <UFormField :label="t('auditFieldwork.testControls.dueDate')">
                    <AppDatePicker
                      v-model="store.testControlForm.dueDate"
                      class="w-full"
                    />
                  </UFormField>
                </div>
              </div>
            </UCard>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="flex flex-col-reverse sm:flex-row items-center justify-between gap-3 border-t border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
          <template v-if="store.isReadOnlyTestControl">
            <div class="flex w-full items-center justify-start gap-2 sm:w-auto">
              <UButton
                color="neutral"
                variant="ghost"
                :label="t('common.close') || 'Tutup'"
                class="w-full sm:w-auto"
                @click="store.showTestControlModal = false"
              />
            </div>
            <div class="flex w-full items-center justify-end gap-2 sm:w-auto">
              <UButton
                v-if="canWrite"
                color="primary"
                icon="i-heroicons-pencil-square"
                :label="t('common.edit') || 'Ubah'"
                class="w-full sm:w-auto font-bold shadow-md shadow-primary/20"
                @click="switchToEditMode()"
              />
            </div>
          </template>
          <template v-else>
            <div class="flex w-full items-center justify-start gap-2 sm:w-auto">
              <UButton
                color="neutral"
                variant="ghost"
                :label="t('common.cancel')"
                class="w-full sm:w-auto"
                @click="store.showTestControlModal = false"
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
                :label="store.isEditingTestControl ? (t('common.edit') || 'Ubah') : (t('common.submit') || 'Simpan')"
                color="primary"
                class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
                :loading="store.loading"
                @click="handleSave"
              />
            </div>
          </template>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRbac } from '~/composables/useRbac'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import { formatDate } from '~/utils/dateConverter'
import { getControlEffectivenessColorClass } from '~/utils/riskLevelBadge'

const store = useAuditFieldworkStore()
const { t } = useI18n()
const { canWrite } = useRbac()

const currentStep = ref(0)
const bodyRef = ref<HTMLElement | null>(null)
const formErrorMessage = ref('')

const steps = computed(() => [
  { key: 'controlInfo', label: t('auditFieldwork.testControls.stepper.controlInfo') || 'Informasi Pengendalian' },
  { key: 'testProcedure', label: t('auditFieldwork.testControls.stepper.testProcedure') || 'Prosedur Pengujian' },
  { key: 'findingRecommendation', label: t('auditFieldwork.testControls.stepper.findingRecommendation') || 'Temuan & Rekomendasi' },
  { key: 'mitigationPlan', label: t('auditFieldwork.testControls.stepper.mitigationPlan') || 'Rencana Mitigasi' }
])

const switchToEditMode = () => {
  store.isReadOnlyTestControl = false
  store.isEditingTestControl = true
  currentStep.value = 0
  formErrorMessage.value = ''
}

const getControlTypeColor = (type: string) => {
  const colors: Record<string, "success" | "warning" | "info" | "neutral" | "primary"> = {
    'Preventive': 'success',
    'Detective': 'warning',
    'Corrective': 'info',
    'Manual': 'neutral',
    'Automated': 'primary'
  }
  return colors[type] || 'neutral'
}

const validateStep = (idx: number): boolean => {
  if (idx === 0) {
    if (!store.testControlForm.controlName?.trim()) {
      formErrorMessage.value = t('auditFieldwork.testControls.nameRequired') || 'Nama pengendalian wajib diisi'
      return false
    }
    if (!store.testControlForm.controlType?.trim()) {
      formErrorMessage.value = t('auditFieldwork.testControls.typeRequired') || 'Jenis pengendalian wajib dipilih'
      return false
    }
    if (!store.testControlForm.controlDescription?.trim()) {
      formErrorMessage.value = t('auditFieldwork.testControls.descriptionRequired') || 'Deskripsi pengendalian wajib diisi'
      return false
    }
    formErrorMessage.value = ''
    return true
  }
  if (idx === 1) {
    if (!store.testControlForm.testProcedure?.trim()) {
      formErrorMessage.value = t('auditFieldwork.testControls.stepsRequired') || 'Langkah pengujian wajib diisi'
      return false
    }
    if (!store.testControlForm.testResult?.trim()) {
      formErrorMessage.value = t('auditFieldwork.testControls.resultRequired') || 'Hasil pengujian wajib dipilih'
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
  store.saveTestControl()
}

watch(
  () => store.showTestControlModal,
  (open) => {
    if (open) {
      currentStep.value = 0
      formErrorMessage.value = ''
    }
  }
)
</script>
