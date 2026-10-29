<template>
  <UModal 
    :open="store.showModalF01" 
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
    @update:open="(open: boolean) => { if (!open) store.closeModalF01() }"
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
                {{ t('workingPaper.headerForm.title') }}
              </h3>
              <p class="mt-0.5 text-xs text-[var(--text-muted)]">
                {{ steps[currentStep]?.label }} ({{ t('auditActivityPlan.form.stepNum', { current: currentStep + 1, total: steps.length }) }})
                <template v-if="store.headerForm.assignmentLetterId">
                  <span class="mx-1.5 opacity-60">·</span>
                  {{ store.headerForm.assignmentLetterId }}
                </template>
              </p>
            </div>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark-20-solid"
            class="-my-1"
            :aria-label="t('common.close')"
            @click="store.closeModalF01"
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

          <!-- Step 1: Audit Purpose sampai Activities -->
          <div v-show="currentStep === 0" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-clipboard-document-list"
                    class="h-5 w-5 text-primary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('workingPaper.headerForm.stepper.purposeAndActivities') }}
                  </h3>
                </div>
              </template>

              <!-- Audit Purpose -->
              <div class="space-y-1.5">
                <label class="block font-semibold text-sm text-[var(--text-main)]">
                  {{ t('workingPaper.headerForm.auditPurpose') }}
                </label>
                <NumberedListInput
                  v-model="auditPurposeItems"
                  :label="t('workingPaper.headerForm.auditPurpose')"
                  :disabled="!auditPurposeEditable"
                  test-id="list-audit-purpose"
                />
                <p class="mt-1 text-xs text-gray-400 dark:text-gray-500">
                  {{ t('workingPaper.headerForm.auditPurposePlaceholder') }}
                </p>
              </div>

              <!-- Business Process -->
              <UFormField 
                :label="t('workingPaper.headerForm.businessProcess')"
                name="businessProcess" 
                required 
              >
                <UInput 
                  v-model="store.headerForm.businessProcess" 
                  class="w-full"
                  :placeholder="t('workingPaper.headerForm.businessProcessPlaceholder')"
                  :maxlength="100"
                  @invalid="($event.target as any)?.setCustomValidity(t('workingPaper.headerForm.businessProcessValidation'))"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.headerForm.businessProcess ? store.headerForm.businessProcess.length : 0 }}/100
                </div>
              </UFormField>

              <!-- Activities -->
              <div class="space-y-2">
                <label class="block font-semibold text-sm text-[var(--text-main)]">
                  {{ t('workingPaper.headerForm.activities') }}
                </label>
                <div class="space-y-3">
                  <div v-for="(activity, index) in store.headerForm.activities" :key="activity.id" class="flex gap-2 items-start">
                    <div class="flex-1">
                      <UTextarea 
                        v-model="activity.name" 
                        :placeholder="t('workingPaper.headerForm.activityPlaceholder')"
                        autoresize
                        :rows="2"
                        :maxrows="5"
                        class="w-full"
                      />
                    </div>
                    <UButton 
                      v-if="store.headerForm.activities.length > 1"
                      icon="i-heroicons-trash" 
                      color="error" 
                      variant="ghost" 
                      class="mt-1"
                      @click="store.removeActivity(index)" 
                    />
                  </div>
                  <UButton 
                    color="primary" 
                    variant="soft"
                    icon="i-heroicons-plus" 
                    :label="t('workingPaper.headerForm.addActivity')"
                    @click="store.addActivity()"
                  />
                </div>
              </div>
            </UCard>
          </div>

          <!-- Step 2: Audit Period sampai Team -->
          <div v-show="currentStep === 1" class="space-y-5">
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-calendar-days"
                    class="h-5 w-5 text-secondary-500"
                  />
                  <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                    {{ t('workingPaper.headerForm.stepper.periodAndTeam') }}
                  </h3>
                </div>
              </template>

              <!-- Audit Period -->
              <UFormField 
                :label="t('workingPaper.headerForm.auditPeriod')"
                required 
              >
                <div class="flex flex-col sm:flex-row items-center gap-3 w-full">
                  <UFormField name="periodStart" class="w-full flex-1">
                    <AppDatePicker 
                      v-model="store.headerForm.periodStart" 
                      class="w-full"
                    />
                  </UFormField>
                  <span class="text-gray-500 font-bold whitespace-nowrap">{{ t('workingPaper.headerForm.until') }}</span>
                  <UFormField name="periodEnd" class="w-full flex-1">
                    <AppDatePicker 
                      v-model="store.headerForm.periodEnd" 
                      class="w-full"
                    />
                  </UFormField>
                </div>
              </UFormField>

              <!-- Location -->
              <UFormField 
                :label="t('workingPaper.headerForm.location')"
                name="location" 
                required 
              >
                <USelectMenu 
                  v-model="store.headerForm.location" 
                  :items="store.options.location" 
                  :loading="locationStore.loading"
                  :placeholder="t('workingPaper.headerForm.locationPlaceholder')"
                  class="w-full" 
                />
              </UFormField>

              <!-- Team -->
              <div class="space-y-2">
                <label class="block font-semibold text-sm text-[var(--text-main)]">
                  {{ t('workingPaper.headerForm.team') }}
                </label>
                <div class="space-y-3">
                  <div v-if="store.headerForm.teamMembers && store.headerForm.teamMembers.length > 0" class="space-y-2">
                    <div 
                      v-for="(member, index) in store.headerForm.teamMembers" 
                      :key="member.id || index"
                      class="flex items-center justify-between gap-3 p-3 bg-gray-50 dark:bg-gray-800/60 rounded-xl border border-gray-200 dark:border-gray-700/80 shadow-xs"
                    >
                      <div class="flex items-center gap-2.5 min-w-0">
                        <div class="size-8 rounded-full bg-primary-100 dark:bg-primary-950 flex items-center justify-center shrink-0">
                          <UIcon name="i-heroicons-user" class="size-4 text-primary-600 dark:text-primary-400" />
                        </div>
                        <span class="text-sm font-semibold text-gray-900 dark:text-white truncate">
                          {{ member.name || '-' }}
                        </span>
                      </div>
                      <UBadge 
                        color="primary" 
                        variant="subtle" 
                        size="md"
                        class="font-medium shrink-0"
                      >
                        {{ member.role || t('workingPaper.headerForm.member') }}
                      </UBadge>
                    </div>
                  </div>
                  <div 
                    v-else 
                    class="text-sm text-gray-400 dark:text-gray-500 italic p-3 bg-gray-50 dark:bg-gray-800/40 rounded-xl border border-dashed border-gray-200 dark:border-gray-700"
                  >
                    {{ t('workingPaper.headerForm.teamAutoFilled') }}
                  </div>
                  <p class="text-xs text-gray-400 dark:text-gray-500">
                    {{ t('workingPaper.headerForm.teamReadonly') }}
                  </p>
                </div>
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
              @click="store.closeModalF01"
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
              :label="store.isEditingF01 ? t('workingPaper.headerForm.update') : t('workingPaper.headerForm.submit')"
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
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useWorkingPaperStore } from '~/stores/working-paper'
import { useLocationStore } from '~/stores/location'
import { useI18n } from '~/composables/useI18n'
import NumberedListInput from '~/components/shared/NumberedListInput.vue'
import {
  auditPurposeItemsFromText,
  auditPurposeTextFromItems
} from '~/utils/workingPaperAuditPurpose'

const store = useWorkingPaperStore()
const locationStore = useLocationStore()
const { t } = useI18n()

const currentStep = ref(0)
const bodyRef = ref<HTMLElement | null>(null)
const formErrorMessage = ref('')

const steps = computed(() => [
  { key: 'purposeAndActivities', label: t('workingPaper.headerForm.stepper.purposeAndActivities') || 'Tujuan & Aktivitas' },
  { key: 'periodAndTeam', label: t('workingPaper.headerForm.stepper.periodAndTeam') || 'Periode & Tim' }
])

// Audit Purpose comes from the selected assignment letter, so it is read-only for now.
// Set to true to let the user add, edit and reorder purposes on the working paper.
const auditPurposeEditable = true

// The store keeps one text value (one purpose per line); the form edits it as a list.
const auditPurposeItems = computed<string[]>({
  get: () => auditPurposeItemsFromText(store.headerForm.auditPurpose),
  set: (items) => {
    store.headerForm.auditPurpose = auditPurposeTextFromItems(items)
  }
})

const validateStep = (idx: number): boolean => {
  if (idx === 0) {
    if (!store.headerForm.businessProcess?.trim()) {
      formErrorMessage.value = t('workingPaper.headerForm.businessProcessValidation') || 'Business Process Name is required'
      return false
    }
    formErrorMessage.value = ''
    return true
  }
  if (idx === 1) {
    if (!store.headerForm.periodStart || !store.headerForm.periodEnd) {
      formErrorMessage.value = t('workingPaper.headerForm.periodRequired') || 'Periode audit wajib diisi lengkap'
      return false
    }
    if (store.headerForm.periodStart > store.headerForm.periodEnd) {
      formErrorMessage.value = t('workingPaper.common.dateError') || 'Tanggal selesai tidak boleh sebelum tanggal mulai'
      return false
    }
    if (!store.headerForm.location) {
      formErrorMessage.value = t('workingPaper.headerForm.locationRequired') || 'Lokasi audit wajib dipilih'
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
  store.handleSubmitF01()
}

onMounted(() => {
  if (locationStore.locations.length === 0) {
    locationStore.fetchLocations().catch(() => {})
  }
})

watch(
  () => store.showModalF01,
  (open) => {
    if (open) {
      currentStep.value = 0
      formErrorMessage.value = ''
      if (locationStore.locations.length === 0) {
        locationStore.fetchLocations().catch(() => {})
      }
    }
  }
)
</script>
