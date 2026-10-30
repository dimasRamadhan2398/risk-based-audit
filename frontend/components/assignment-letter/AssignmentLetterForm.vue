<template>
  <UModal
    v-model:open="store.isModalOpen"
    :dismissible="false"
    :ui="{
      content: 'z-50 w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      header: 'border-b border-gray-100 dark:border-gray-800 p-4 sm:p-5 text-gray-900 dark:text-white font-bold shrink-0',
      body: 'p-4 sm:p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
      footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
      overlay: 'z-40 bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
  >
    <div />
    <template #content>
      <UForm
        :state="store.form"
        @submit.prevent="onFormSubmit"
      >
        <div class="relative rounded-xl shadow-2xl flex flex-col max-h-[90vh]">
          <div class="flex justify-between items-center p-4 sm:p-6 border-b border-gray-100 dark:border-gray-800">
            <div>
              <h2 class="text-xl font-bold text-gray-800 dark:text-white flex items-center gap-2">
                <UIcon
                  :name="store.editingId ? 'i-heroicons-pencil-square' : 'i-heroicons-document-plus'"
                  class="w-6 h-6 text-orange-500"
                />
                {{ store.editingId ? 'Edit Assignment Letter' : 'Add Assignment Letter' }}
              </h2>
              <p class="text-xs text-gray-500 mt-1">
                {{ t('assignmentLetter.formExtra.number') }} <span class="font-mono font-semibold text-orange-600 dark:text-orange-400">{{ store.editingId ? (store.assignmentLetterList.find(l => l.id === store.editingId)?.letterNumber || store.generateNomorSurat(store.form.auditTeam, store.form.auditYear)) : store.generateNomorSurat(store.form.auditTeam, store.form.auditYear) }}</span>
              </p>
            </div>
            <UIcon
              name="i-heroicons-x-mark"
              size="xl"
              class="cursor-pointer"
              @click="store.closeModal"
            />
          </div>

          <!-- Stepper -->
          <div class="w-full overflow-x-auto border-b border-gray-100 dark:border-gray-800 px-4 py-3 sm:px-6 sm:py-4 scrollbar-thin">
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

          <div
            ref="bodyRef"
            class="p-4 sm:p-6 md:p-8 overflow-y-auto space-y-6 flex-1"
          >
            <!-- Step 1: audit information -->
            <UCard
              v-if="currentStep === 0"
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
                    {{ t('assignmentLetter.stepper.auditInfo') }}
                  </h3>
                </div>
              </template>
              <div class="space-y-6">
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                  <UFormField
                    :label="t('actionTakenReport.detail.auditTitle')"
                    class="md:col-span-4"
                    required
                  >
                    <UInput
                      v-model="store.form.auditTitle"
                      :placeholder="t('assignmentLetter.formExtra.titleExample')"
                      size="lg"
                      class="w-full"
                      required
                      maxlength="100"
                      @invalid="($event.target as any)?.setCustomValidity('Judul audit maksimal 100 karakter dan wajib diisi')"
                      @input="($event.target as any)?.setCustomValidity('')"
                    />
                    <div class="text-xs text-gray-500 mt-1 text-right">
                      {{ store.form.auditTitle ? store.form.auditTitle.length : 0 }}/100
                    </div>
                  </UFormField>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                  <UFormField
                    :label="t('assignmentLetter.formExtra.category')"
                    class="md:col-span-4"
                    required
                  >
                    <USelectMenu
                      v-model="store.form.category"
                      :items="categoryOptions"
                      size="lg"
                      class="w-full"
                      required
                    />
                  </UFormField>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                  <UFormField
                    :label="t('auditPriority.selectYear')"
                    class="md:col-span-4"
                    required
                  >
                    <USelectMenu
                      v-model="store.form.auditYear"
                      :items="yearOptions"
                      size="lg"
                      class="w-full"
                      required
                    />
                  </UFormField>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center border-y border-gray-100  py-4 my-6">
                  <label class="font-bold text-gray-700 ">{{ t('assignmentLetter.formExtra.team') }}</label>
                  <UFormField class="md:col-span-3">
                    <div class="flex flex-col gap-2">
                      <label
                        v-for="team in store.options.auditTeam"
                        :key="team.value"
                        class="inline-flex items-center gap-2 cursor-pointer"
                      >
                        <input
                          v-model="store.form.auditTeam"
                          type="radio"
                          :value="team.value"
                          class="accent-orange-500 w-4 h-4"
                          required
                        >
                        <span class="text-sm font-medium text-gray-800 dark:text-white">{{ team.label }}</span>
                      </label>
                    </div>
                  </UFormField>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                  <UFormField
                    :label="t('workingPaper.headerForm.auditPeriod')"
                    class="md:col-span-4"
                    required
                  >
                    <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 sm:gap-4">
                      <AppDatePicker
                        v-model="store.form.startPeriod"
                        size="lg"
                        class="flex-1"
                        :class="{ 'ring-red-500': store.dateError }"
                      />
                      <span class="font-bold text-gray-500 text-center">{{ t('assignmentLetter.formExtra.to') }}</span>
                      <AppDatePicker
                        v-model="store.form.finishPeriod"
                        size="lg"
                        class="flex-1"
                        :class="{ 'ring-red-500': store.dateError }"
                      />
                    </div>
                    <p
                      v-if="store.dateError"
                      class="text-red-500 text-sm font-semibold mt-1"
                    >
                      {{ store.dateError }}
                    </p>
                  </UFormField>
                </div>
              </div>
            </UCard>

            <!-- Step 2: work unit & team -->
            <UCard
              v-if="currentStep === 1"
              :ui="{ body: 'p-4 sm:p-6' }"
              class="border border-[var(--border-main)] bg-transparent ring-0 shadow-none"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-user-group"
                    class="h-5 w-5 text-amber-500"
                  />
                  <h3 class="text-lg font-bold text-[var(--text-main)]">
                    {{ t('assignmentLetter.stepper.unitTeam') }}
                  </h3>
                </div>
              </template>
              <div class="space-y-6">
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                  <UFormField
                    :label="t('assignmentLetter.formExtra.workUnit')"
                    class="md:col-span-4"
                    required
                  >
                    <USelectMenu
                      v-model="store.form.workingUnit"
                      :items="store.workUnitOptions"
                      :loading="store.departmentsLoading"
                      :placeholder="t('assignmentLetter.formExtra.workUnitPlaceholder')"
                      size="lg"
                      class="w-full"
                      :popper="{ strategy: 'absolute' }"
                      create-item
                      @create="(val: string) => { store.form.workingUnit = val.trim() }"
                      required
                    >
                      <template #empty>
                        {{ store.departmentsLoading ? t('common.loading') : (store.departmentsError || t('masterData.department.empty')) }}
                      </template>
                    </USelectMenu>
                  </UFormField>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                  <UFormField
                    :label="t('assignmentLetter.formExtra.company')"
                    class="md:col-span-4"
                  >
                    <USelectMenu
                      v-model="store.form.companyId"
                      :items="companySelectItems"
                      value-key="value"
                      :placeholder="t('assignmentLetter.formExtra.companyPlaceholder')"
                      size="lg"
                      class="w-full"
                      :clear="true"
                      :loading="masterOptions.loading"
                      @update:model-value="onCompanyChange"
                    />
                  </UFormField>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
                  <UFormField
                    :label="t('assignmentLetter.formExtra.members')"
                    class="md:col-span-4 space-y-3"
                    required
                  >
                    <div
                      v-for="(member, index) in store.form.membersList"
                      :key="index"
                      class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2"
                    >
                      <div class="flex-1">
                        <USelectMenu
                          v-model="member.name"
                          :items="store.getAvailableMemberOptions(index)"
                          :loading="store.employeesLoading"
                          :placeholder="t('assignmentLetter.formExtra.memberName')"
                          class="w-full"
                          create-item
                          @create="(val: string) => { member.name = val.trim() }"
                          required
                        >
                          <template #empty>
                            {{ store.employeesLoading ? t('common.loading') : (store.employeesError || t('masterData.employee.empty')) }}
                          </template>
                        </USelectMenu>
                        <p
                          v-if="store.isDuplicateMember(member.name, index)"
                          class="text-xs text-red-500 font-semibold mt-1"
                        >
                          {{ t('assignmentLetter.formExtra.memberDuplicate') }}
                        </p>
                      </div>
                      <div class="flex items-center gap-2">
                        <USelectMenu
                          v-model="member.role"
                          :items="store.options.role"
                          :placeholder="t('assignmentLetter.formExtra.role')"
                          class="w-full sm:w-44"
                          required
                        />
                        <UButton
                          icon="i-heroicons-trash"
                          color="error"
                          variant="ghost"
                          @click="store.removeItem(store.form.membersList, index)"
                        />
                      </div>
                    </div>
                    <UButton
                      class="mt-4"
                      color="primary"
                      variant="soft"
                      icon="i-heroicons-plus"
                      :label="t('workingPaper.headerForm.addMember')"
                      @click="store.addItem(store.form.membersList, { name: '', role: 'Member' })"
                    />
                    <p class="text-md text-orange-600  font-semibold">
                      {{ t('assignmentLetter.formExtra.memberHint') }}
                    </p>
                  </UFormField>
                </div>
              </div>
            </UCard>

            <!-- Step 3: purpose, scope & CC -->
            <UCard
              v-if="currentStep === 2"
              :ui="{ body: 'p-4 sm:p-6' }"
              class="border border-[var(--border-main)] bg-transparent ring-0 shadow-none"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-clipboard-document-list"
                    class="h-5 w-5 text-amber-500"
                  />
                  <h3 class="text-lg font-bold text-[var(--text-main)]">
                    {{ t('assignmentLetter.stepper.purposeDistribution') }}
                  </h3>
                </div>
              </template>
              <div class="space-y-6">
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
                  <UFormField
                    :label="t('workingPaper.headerForm.auditPurpose')"
                    class="md:col-span-4 space-y-3"
                  >
                    <div
                      v-for="(purpose, index) in store.form.purposeList"
                      :key="index"
                      class="flex items-start gap-2"
                    >
                      <span class="mt-2 font-bold text-gray-400">{{ index + 1 }}.</span>
                      <UTextarea
                        v-model="store.form.purposeList[index]"
                        :placeholder="t('assignmentLetter.formExtra.purposePlaceholder')"
                        class="flex-1"
                        :rows="2"
                      />
                      <UButton
                        icon="i-heroicons-trash"
                        color="error"
                        variant="ghost"
                        class="mt-1"
                        @click="store.removeItem(store.form.purposeList, index)"
                      />
                    </div>
                    <UButton
                      class="mt-4"
                      color="primary"
                      variant="soft"
                      icon="i-heroicons-plus"
                      :label="t('assignmentLetter.formExtra.addPurpose')"
                      @click="store.addItem(store.form.purposeList, '')"
                    />
                  </UFormField>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
                  <UFormField
                    :label="t('actionTakenReport.detail.scope')"
                    class="md:col-span-4 space-y-3"
                  >
                    <div
                      v-for="(scope, index) in store.form.scopeList"
                      :key="index"
                      class="flex items-start gap-2"
                    >
                      <UInput
                        v-model="store.form.scopeList[index]"
                        :placeholder="t('assignmentLetter.formExtra.scopePlaceholder')"
                        class="flex-1"
                      />
                      <UButton
                        icon="i-heroicons-trash"
                        color="error"
                        variant="ghost"
                        @click="store.removeItem(store.form.scopeList, index)"
                      />
                    </div>
                    <UButton
                      class="mt-4"
                      color="primary"
                      variant="soft"
                      icon="i-heroicons-plus"
                      :label="t('assignmentLetter.formExtra.addScope')"
                      @click="store.addItem(store.form.scopeList, '')"
                    />
                  </UFormField>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
                  <UFormField
                    :label="t('assignmentLetter.formExtra.cc')"
                    class="md:col-span-4 space-y-3"
                  >
                    <div
                      v-for="(cc, index) in store.form.ccList"
                      :key="index"
                      class="flex items-center gap-2"
                    >
                      <UInput
                        v-model="store.form.ccList[index]"
                        :placeholder="t('assignmentLetter.formExtra.ccPosition')"
                        class="flex-1"
                        required
                        maxlength="100"
                        @invalid="($event.target as any)?.setCustomValidity('CC Position maksimal 100 karakter dan wajib diisi')"
                        @input="($event.target as any)?.setCustomValidity('')"
                      />
                      <div class="text-xs text-gray-500 mt-1 text-right">
                        {{ store.form.ccList[index] ? store.form.ccList[index].length : 0 }}/100
                      </div>
                      <UButton
                        icon="i-heroicons-trash"
                        color="error"
                        variant="ghost"
                        @click="store.removeItem(store.form.ccList, index)"
                      />
                    </div>
                    <UButton
                      class="mt-4"
                      color="primary"
                      variant="soft"
                      icon="i-heroicons-plus"
                      :label="t('assignmentLetter.formExtra.addCc')"
                      @click="store.addItem(store.form.ccList, '')"
                    />
                  </UFormField>
                </div>
              </div>
            </UCard>
          </div>

          <!-- Error message display -->
          <UAlert
            v-if="store.errorMsg"
            color="error"
            variant="solid"
            icon="i-heroicons-exclamation-triangle"
            :description="store.errorMsg"
            class="m-6 mt-0 w-auto shadow-sm"
            data-testid="form-error"
          />

          <div class="p-4 sm:p-6 border-t border-gray-100 dark:border-gray-800 flex flex-col-reverse sm:flex-row justify-between items-stretch sm:items-center gap-3 sm:gap-4">
            <UButton
              :label="t('common.cancel')"
              color="neutral"
              variant="ghost"
              class="w-full sm:w-auto"
              @click="store.closeModal"
            />

            <div class="flex w-full items-center justify-end gap-2 sm:w-auto sm:gap-3">
              <UButton
                v-if="currentStep > 0"
                icon="i-heroicons-arrow-left"
                :label="t('assignmentLetter.stepper.previous')"
                color="neutral"
                variant="outline"
                size="lg"
                class="flex-1 sm:flex-none"
                @click="handleBack"
              />
              <UButton
                v-if="currentStep < steps.length - 1"
                trailing-icon="i-heroicons-arrow-right"
                :label="t('assignmentLetter.stepper.next')"
                color="primary"
                size="lg"
                class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
                data-testid="form-next"
                @click="handleNext"
              />
              <UButton
                v-else
                icon="i-heroicons-check"
                :label="store.editingId ? 'Update Assignment Letter' : 'Save Assignment Letter'"
                :loading="store.loading"
                :disabled="store.loading"
                color="primary"
                size="lg"
                class="flex-1 sm:flex-none font-bold px-8 shadow-md shadow-primary/20"
                data-testid="form-submit"
                @click="handleFormSubmit"
              />
            </div>
          </div>
        </div>
      </UForm>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import { useMasterOptionsStore } from '~/stores/master-options'
import { getFiscalYearStrings } from '~/composables/useFiscalYear'

const { t } = useI18n()

const store = useAssignmentLetterStore()
const masterOptions = useMasterOptionsStore()

const categoryOptions = ['Assurance', 'Special Audit', 'Specific Reason', 'Consulting Services', 'Follow-Up Audit', 'Investigation', 'Quality Assurance Review']
const yearOptions = computed(() => store.options.yearOptions || getFiscalYearStrings())

// Company name is data (the tenant default entity); only the surrounding text is translated.
const DEFAULT_COMPANY_NAME = 'PT AIFL Indonesia'
const noneOption = computed(() => ({
  label: t('assignmentLetter.formExtra.companyNone', { company: DEFAULT_COMPANY_NAME }),
  value: '__none__'
}))
const companySelectItems = computed(() => [
  noneOption.value,
  ...masterOptions.companyOptions
])

const onCompanyChange = (val: string) => {
  if (!val || val === '__none__') {
    store.form.companyId = ''
    store.form.companyName = ''
    return
  }
  const comp = masterOptions.companyOptions.find(c => c.value === val)
  store.form.companyName = comp ? comp.label : ''
}

/**
 * Wizard: 1) audit information (title → period), 2) work unit & team, 3) purpose, scope & CC.
 * Each step mirrors the checks in store.handleSubmit for its own fields; the final submit still
 * goes through store.handleSubmit unchanged.
 */
const currentStep = ref(0)
const bodyRef = ref<HTMLElement | null>(null)

const steps = computed(() => [
  { key: 'auditInfo', label: t('assignmentLetter.stepper.auditInfo') },
  { key: 'unitTeam', label: t('assignmentLetter.stepper.unitTeam') },
  { key: 'purposeDistribution', label: t('assignmentLetter.stepper.purposeDistribution') }
])

/** Returns the first error message of a step, or null when the step is valid. */
const validateStep = (idx: number): string | null => {
  const form = store.form
  if (idx === 0) {
    if (!form.auditTitle) return t('assignmentLetter.validation.titleRequired')
    if (!form.category) return t('assignmentLetter.validation.categoryRequired')
    if (!form.auditYear) return t('assignmentLetter.validation.yearRequired')
    if (!form.auditTeam) return t('assignmentLetter.validation.teamRequired')
    if (!form.startPeriod || !form.finishPeriod) return t('assignmentLetter.validation.periodRequired')
    if (store.dateError) return t('assignmentLetter.validation.periodOrder')
  } else if (idx === 1) {
    if (!form.workingUnit) return t('assignmentLetter.validation.workUnitRequired')
    if (!form.membersList.every(m => m.name && m.role)) return t('assignmentLetter.validation.membersInvalid')
    if (store.hasDuplicateMembers) return 'Nama anggota tim tidak boleh sama (duplikat).'
  } else if (idx === 2) {
    if (!form.ccList.every(cc => cc && cc.trim() !== '')) return t('assignmentLetter.validation.ccInvalid')
  }
  return null
}

/** Validates steps [0, targetIdx); on failure shows the error and jumps to the failing step. */
const validateUpTo = (targetIdx: number): boolean => {
  for (let i = 0; i < targetIdx; i++) {
    const error = validateStep(i)
    if (error) {
      store.errorMsg = error
      currentStep.value = i
      bodyRef.value?.scrollTo?.({ top: 0, behavior: 'smooth' })
      return false
    }
  }
  return true
}

const goToStep = (targetIdx: number) => {
  if (targetIdx <= currentStep.value) {
    currentStep.value = targetIdx
    return
  }
  if (!validateUpTo(targetIdx)) return
  store.errorMsg = ''
  currentStep.value = targetIdx
}

const handleNext = () => goToStep(currentStep.value + 1)
const handleBack = () => {
  if (currentStep.value > 0) currentStep.value--
}

const handleFormSubmit = () => {
  if (!validateUpTo(steps.value.length)) return
  store.handleSubmit()
}

// Enter inside an input submits the <form>: advance on earlier steps, save on the last one.
const onFormSubmit = () => {
  if (currentStep.value < steps.value.length - 1) handleNext()
  else handleFormSubmit()
}

watch(
  () => store.isModalOpen,
  (open) => {
    if (open) {
      currentStep.value = 0
      masterOptions.fetchAll()
      if (!store.employeesLoaded && !store.employeesLoading) {
        store.fetchEmployees().catch(() => {})
      }
      if (!store.departmentsLoaded && !store.departmentsLoading) {
        store.fetchDepartments().catch(() => {})
      }
    }
  },
  { immediate: true }
)

onMounted(() => {
  if (!store.employeesLoaded && !store.employeesLoading) {
    store.fetchEmployees().catch(() => {})
  }
  if (!store.departmentsLoaded && !store.departmentsLoading) {
    store.fetchDepartments().catch(() => {})
  }
})
</script>
