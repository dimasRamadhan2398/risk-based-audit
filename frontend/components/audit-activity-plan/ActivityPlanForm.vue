<template>
  <UModal 
    v-model:open="store.isModalOpen" 
    :ui="{ content: 'sm:max-w-4xl w-full bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' }"
  >
    <template #content>
      <div class="relative flex flex-col max-h-[90vh] transition-colors duration-300">
        <!-- Header -->
        <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
          <div>
            <h3 class="text-xl font-bold text-[var(--text-main)]">
              {{ store.isEditMode ? t('auditActivityPlan.editPlan') : t('auditActivityPlan.form.createPlanTitle') }}
            </h3>
            <p class="text-xs text-[var(--text-muted)] mt-0.5">
              {{ steps[currentStep]?.label }} — {{ t('auditActivityPlan.form.stepNum', { current: currentStep + 1, total: steps.length }) }}
            </p>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark-20-solid"
            class="-my-1"
            @click="store.closeModal"
          />
        </div>

        <!-- Stepper Navigation Header (Full Width) -->
        <div class="w-full px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
          <div class="w-full flex items-center justify-between gap-2">
            <div 
              v-for="(step, idx) in steps" 
              :key="step.key"
              class="flex items-center flex-1 min-w-0 cursor-pointer group"
              @click="goToStep(idx)"
            >
              <div class="flex items-center gap-2.5 shrink-0">
                <div 
                  class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold transition-all duration-300"
                  :class="currentStep === idx 
                    ? 'bg-primary text-white shadow-md ring-4 ring-primary/20 scale-105' 
                    : currentStep > idx 
                      ? 'bg-secondary text-white shadow-sm' 
                      : 'bg-gray-100 dark:bg-gray-800 text-gray-400 group-hover:bg-gray-200 dark:group-hover:bg-gray-700'"
                >
                  <UIcon v-if="currentStep > idx" name="i-heroicons-check" class="w-4 h-4 stroke-[3]" />
                  <span v-else>{{ idx + 1 }}</span>
                </div>
                <span 
                  class="text-xs font-bold truncate transition-colors"
                  :class="currentStep === idx 
                    ? 'text-primary font-black' 
                    : currentStep > idx 
                      ? 'text-secondary font-bold' 
                      : 'text-gray-400 dark:text-gray-500'"
                >
                  {{ step.label }}
                </span>
              </div>
              <div 
                v-if="idx < steps.length - 1" 
                class="flex-1 h-1 mx-3 rounded transition-all duration-300"
                :class="currentStep > idx ? 'bg-secondary' : 'bg-gray-200 dark:bg-gray-800'"
              ></div>
            </div>
          </div>
        </div>

        <!-- Form Body -->
        <div class="flex-1 overflow-y-auto p-6 space-y-6">
          <!-- Step Error Alert Banner -->
          <Transition name="fade">
            <UAlert
              v-if="stepError"
              color="error"
              variant="solid"
              icon="i-heroicons-exclamation-triangle"
              :title="t('common.error')"
              :description="stepError"
              closable
              class="mb-2 shadow-md"
              @close="stepError = ''"
            />
          </Transition>

          <UForm :state="store.formState" @submit="handleFormSubmit" class="space-y-6">
            
            <!-- Step 1: Basic Information -->
            <div v-if="currentStep === 0" class="space-y-6">
              <UCard :ui="{ body: 'p-6' }" class="border border-[var(--border-main)]">
                <template #header>
                  <div class="flex items-center gap-2">
                    <UIcon name="i-heroicons-information-circle" class="w-5 h-5 text-amber-500" />
                    <h3 class="text-lg font-bold text-[var(--text-main)]">{{ t('auditActivityPlan.form.basicInfo') }}</h3>
                  </div>
                </template>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
                  <UFormField :label="t('auditActivityPlan.form.planTitle')" required>
                    <UInput v-model="store.formState.planTitle" placeholder="Audit of the Year 2026" class="w-full"/>
                  </UFormField>
                  <UFormField :label="t('auditActivityPlan.form.planYear')" required>
                    <UInput v-model="store.formState.planYear" class="w-full"/>
                  </UFormField>
                  <UFormField :label="t('auditActivityPlan.form.planPeriod')" required class="col-span-1 md:col-span-2 w-full">
                    <div class="flex items-center gap-3">
                      <UInput v-model="store.formState.planPeriodStart" type="date" class="flex-1 w-full" />
                      <span class="text-xs font-bold text-gray-400 uppercase">{{ t('auditActivityPlan.form.to') }}</span>
                      <UInput v-model="store.formState.planPeriodEnd" type="date" class="flex-1 w-full" />
                    </div>
                  </UFormField>
                  <UFormField :label="t('auditActivityPlan.form.department')" required>
                    <USelectMenu 
                      v-model="store.formState.department" 
                      placeholder="Risk Audit IT" 
                      class="w-full" 
                      :items="Object.values(AuditDepartment)" 
                    />
                  </UFormField>
                  <UFormField :label="t('auditActivityPlan.form.createdBy')">
                    <UInput v-model="store.formState.createdBy" class="w-full" />
                  </UFormField>
                </div>
                <UFormField :label="t('auditActivityPlan.form.creationDate')" class="pt-5">
                  <UInput v-model="store.formState.creationDate" type="date" class="w-full" readonly />
                </UFormField>
              </UCard>
            </div>

            <!-- Step 2: Planned Audit Activities -->
            <div v-else-if="currentStep === 1" class="space-y-6">
              <UCard :ui="{ body: 'p-6' }" class="border border-[var(--border-main)]">
                <template #header>
                  <div class="flex justify-between items-center">
                    <div class="flex items-center gap-2">
                      <UIcon name="i-heroicons-clipboard-document-check" class="w-5 h-5 text-amber-500" />
                      <h3 class="text-lg font-bold text-[var(--text-main)]">{{ t('auditActivityPlan.form.plannedActivities') }}</h3>
                    </div>
                    <UButton 
                      color="warning" 
                      variant="solid" 
                      icon="i-heroicons-plus" 
                      :label="t('auditActivityPlan.form.addAuditActivity')" 
                      class="font-bold shadow-sm"
                      @click="store.addPlannedActivity" 
                    />
                  </div>
                </template>

                <div class="space-y-6">
                  <div 
                    v-for="(activity, index) in store.formState.plannedActivities" 
                    :key="activity.id" 
                    class="border border-[var(--border-main)] rounded-xl p-5 relative bg-[var(--bg-surface)] shadow-sm space-y-4"
                  >
                    <UButton
                      color="error"
                      variant="ghost"
                      icon="i-heroicons-trash"
                      class="absolute top-3 right-3"
                      @click="store.removePlannedActivity(index)"
                    />
                    <h4 class="font-bold text-sm text-gray-800 dark:text-gray-200 border-b border-gray-100 dark:border-gray-800 pb-2">
                      {{ t('auditActivityPlan.form.activityNum', { num: index + 1 }) }}
                    </h4>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <UFormField :label="t('auditActivityPlan.form.activityTitle')" required>
                        <UInput v-model="activity.auditName" placeholder="Audit IT Q1" class="w-full"/>
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.auditee')" required>
                        <UInput v-model="activity.auditee" placeholder="Jamil" class="w-full"/>
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.category')" class="col-span-1 md:col-span-2">
                        <USelectMenu
                          v-model="activity.category"
                          :items="Object.values(AuditCategory)"
                          :placeholder="t('auditActivityPlan.form.selectCategory')"
                          class="w-full"
                        />
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.associatedRisk')" class="col-span-1 md:col-span-2">
                        <USelectMenu
                          :model-value="getFilteredRisksForDept(store.formState.department).find(r => r.name === activity.riskName)"
                          @update:model-value="(val: any) => {
                            activity.riskName = val ? val.name : '';
                            activity.riskLevel = val ? val.riskLevel : '';
                          }"
                          :items="getFilteredRisksForDept(store.formState.department)"
                          option-key="name"
                          :placeholder="t('auditActivityPlan.form.selectRiskProfile')"
                          class="w-full"
                        >
                          <template #item="{ item }">
                            <div class="flex items-center gap-2 max-w-full">
                              <span 
                                class="w-2.5 h-2.5 rounded-full shrink-0" 
                                :style="{ backgroundColor: getRiskLevelColorHex(item.riskLevel) }"
                              ></span>
                              <span class="text-[10px] font-bold text-gray-500 shrink-0">[{{ item.riskLevel }}]</span>
                              <span class="truncate text-md">{{ item.name }}</span>
                            </div>
                          </template>
                        </USelectMenu>
                      </UFormField>

                      <UFormField :label="t('auditActivityPlan.form.riskLevel')">
                        <USelectMenu
                          v-model="activity.riskLevel"
                          :items="store.riskLevelOptions"
                          value-key="value"
                          label-key="label"
                          :placeholder="t('auditActivityPlan.form.selectRiskLevel')"
                          class="w-full"
                        />
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.duration')">
                        <UInput v-model="activity.duration" type="number" class="w-full"/>
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.priority')">
                        <USelectMenu 
                          v-model="activity.priority" 
                          :items="store.priorityOptions" 
                          value-key="value" 
                          label-key="label" 
                          :placeholder="t('auditActivityPlan.form.selectPriority')" 
                          class="w-full"
                        />
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.auditorsNum')">
                        <UInput v-model="activity.numberOfAuditors" type="number" class="w-full"/>
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.estimatedSchedule')">
                        <UInput v-model="activity.estimatedSchedule" type="date" class="w-full"/>
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.budgetEstimation')">
                        <UInput v-model="activity.budgetEstimation" type="number" class="w-full"/>
                      </UFormField>
                    </div>
                  </div>
                  
                  <div v-if="store.formState.plannedActivities.length === 0" class="text-center text-gray-400 py-8 italic border border-dashed rounded-xl">
                    {{ t('auditActivityPlan.form.noActivitiesAdded') }}
                  </div>

                  <div class="mt-4 border-t border-[var(--border-main)] pt-4">
                    <h4 class="font-bold text-sm text-gray-800 dark:text-gray-200 mb-3">{{ t('auditActivityPlan.form.summaryOfActivities') }}</h4>
                    <div class="grid grid-cols-3 gap-4 text-center">
                      <div class="bg-red-500/10 text-red-600 dark:text-red-400 rounded-xl p-3 border border-red-500/20">
                        <div class="text-xs font-bold uppercase tracking-wider">{{ t('auditActivityPlan.form.highRisk') }}</div>
                        <div class="text-xl font-black mt-0.5">{{ store.formState.plannedActivities.filter((a: any) => String(a.riskLevel) === 'High').length }}</div>
                      </div>
                      <div class="bg-yellow-500/10 text-yellow-600 dark:text-yellow-400 rounded-xl p-3 border border-yellow-500/20">
                        <div class="text-xs font-bold uppercase tracking-wider">{{ t('auditActivityPlan.form.mediumRisk') }}</div>
                        <div class="text-xl font-black mt-0.5">{{ store.formState.plannedActivities.filter((a: any) => String(a.riskLevel) === 'Medium').length }}</div>
                      </div>
                      <div class="bg-green-500/10 text-green-600 dark:text-green-400 rounded-xl p-3 border border-green-500/20">
                        <div class="text-xs font-bold uppercase tracking-wider">{{ t('auditActivityPlan.form.lowRisk') }}</div>
                        <div class="text-xl font-black mt-0.5">{{ store.formState.plannedActivities.filter((a: any) => String(a.riskLevel) === 'Low').length }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </UCard>
            </div>

            <!-- Step 3: Resources & Budget -->
            <div v-else-if="currentStep === 2" class="space-y-6">
              <UCard :ui="{ body: 'p-6' }" class="border border-[var(--border-main)]">
                <template #header>
                  <div class="flex justify-between items-center">
                    <div class="flex items-center gap-2">
                      <UIcon name="i-heroicons-user-group" class="w-5 h-5 text-amber-500" />
                      <h3 class="text-lg font-bold text-[var(--text-main)]">{{ t('auditActivityPlan.form.resourcesAndBudget') }}</h3>
                    </div>
                    <UButton 
                      color="warning" 
                      variant="solid" 
                      icon="i-heroicons-plus" 
                      :label="t('auditActivityPlan.form.addAuditor')" 
                      class="font-bold shadow-sm"
                      @click="store.addResourceAuditor" 
                    />
                  </div>
                </template>
                
                <div class="space-y-6">
                  <div 
                    v-for="(auditor, index) in store.formState.resourceAuditors" 
                    :key="auditor.id" 
                    class="border border-[var(--border-main)] rounded-xl p-5 relative bg-[var(--bg-surface)] shadow-sm space-y-4"
                  >
                    <UButton
                      color="error"
                      variant="ghost"
                      icon="i-heroicons-trash"
                      class="absolute top-3 right-3"
                      @click="store.removeResourceAuditor(index)"
                    />
                    <h4 class="font-bold text-sm text-gray-800 dark:text-gray-200 border-b border-gray-100 dark:border-gray-800 pb-2">
                      {{ t('auditActivityPlan.form.auditorNum', { num: index + 1 }) }}
                    </h4>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <UFormField :label="t('auditActivityPlan.form.name')">
                        <UInput v-model="auditor.name" placeholder="Jamil" class="w-full"/>
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.position')">
                        <UInput v-model="auditor.position" placeholder="IT Auditor" class="w-full"/>
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.competence')">
                        <UInput v-model="auditor.competence" class="w-full"/>
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.availability')">
                        <UInput v-model="auditor.availability" placeholder="Available" class="w-full"/>
                      </UFormField>
                    </div>
                  </div>

                  <div class="mt-6 border-t border-[var(--border-main)] pt-5">
                    <h4 class="font-bold text-sm text-gray-800 dark:text-gray-200 mb-4 flex items-center gap-2">
                      <UIcon name="i-heroicons-banknotes" class="w-5 h-5 text-amber-500" />
                      {{ t('auditActivityPlan.form.budgetPlanning') }}
                    </h4>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <UFormField :label="t('auditActivityPlan.form.totalEstimatedCost')">
                        <UInput v-model="store.formState.budget.totalEstimatedCost" type="number" class="w-full"/>
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.totalAllocatedBudget')">
                        <UInput v-model="store.formState.budget.totalAllocatedBudget" type="number" class="w-full"/>
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.budgetNotes')" class="col-span-1 md:col-span-2">
                        <UTextarea v-model="store.formState.budget.budgetNotes" class="w-full" :rows="3"/>
                      </UFormField>
                    </div>
                  </div>
                </div>
              </UCard>
            </div>

            <!-- Step 4: Review & Approval -->
            <div v-else-if="currentStep === 3" class="space-y-6">
              <UCard :ui="{ body: 'p-6' }" class="border border-[var(--border-main)]">
                <template #header>
                  <div class="flex items-center gap-2">
                    <UIcon name="i-heroicons-check-badge" class="w-5 h-5 text-amber-500" />
                    <h3 class="text-lg font-bold text-[var(--text-main)]">{{ t('auditActivityPlan.form.reviewAndApproval') }}</h3>
                  </div>
                </template>
                <div class="space-y-6">
                  <div>
                    <h4 class="font-bold text-sm text-gray-800 dark:text-gray-200 mb-3">{{ t('auditActivityPlan.view.creator') }}</h4>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <UFormField :label="t('auditActivityPlan.form.creatorName')" required>
                        <UInput v-model="store.formState.review.creatorName" class="w-full"/>
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.creatorPosition')" required>
                        <UInput v-model="store.formState.review.creatorPosition" class="w-full"/>
                      </UFormField>
                    </div>
                  </div>

                  <div class="border-t border-[var(--border-main)] pt-5">
                    <h4 class="font-bold text-sm text-gray-800 dark:text-gray-200 mb-3">{{ t('auditActivityPlan.form.approvedBy') }}</h4>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <UFormField :label="t('auditActivityPlan.form.approverName')" required>
                        <UInput v-model="store.formState.review.approverName" class="w-full"/>
                      </UFormField>
                      <UFormField :label="t('auditActivityPlan.form.approverPosition')" required>
                        <UInput v-model="store.formState.review.approverPosition" class="w-full"/>
                      </UFormField>
                    </div>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4">
                      <UFormField :label="t('auditActivityPlan.form.approvalDate')" required>
                        <UInput v-model="store.formState.review.approvalDate" type="date" class="w-full"/>
                      </UFormField>
                    </div>
                    <UFormField :label="t('auditActivityPlan.form.additionalNotes')" class="pt-4">
                      <UTextarea v-model="store.formState.review.additionalNotes" class="w-full" :rows="3"/>
                    </UFormField>
                  </div>
                </div>
              </UCard>
            </div>

            <!-- Step 5: Attachments -->
            <div v-else-if="currentStep === 4" class="space-y-6">
              <UCard :ui="{ body: 'p-6' }" class="border border-[var(--border-main)]">
                <template #header>
                  <div class="flex items-center gap-2">
                    <UIcon name="i-heroicons-paper-clip" class="w-5 h-5 text-amber-500" />
                    <h3 class="text-lg font-bold text-[var(--text-main)]">{{ t('auditActivityPlan.form.attachment') }}</h3>
                  </div>
                </template>
                <div class="space-y-4">
                  <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <UFormField :label="t('auditActivityPlan.form.attachmentCategory')">
                      <USelectMenu v-model="store.formState.attachmentCategory" :items="store.attachmentCategoryOptions" class="w-full"/>
                    </UFormField>
                    <UFormField :label="t('auditActivityPlan.form.attachmentUploadedBy')">
                      <UInput v-model="store.formState.attachmentUploadedBy" placeholder="Example: Auditor" class="w-full" />
                    </UFormField>
                    <UFormField :label="t('auditActivityPlan.form.attachmentUploadDate')">
                      <UInput type="date" v-model="store.formState.attachmentUploadDate" class="w-full"/>
                    </UFormField>
                  </div>
                  <UFormField :label="t('auditActivityPlan.form.uploadAttachmentHere')" size="lg" class="pt-2">
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

          </UForm>
        </div>

        <!-- Stepper Navigation Footer -->
        <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-between items-center">
          <UButton 
            :label="t('auditActivityPlan.form.cancel')" 
            color="neutral" 
            variant="ghost" 
            @click="store.closeModal" 
          />

          <div class="flex items-center gap-3">
            <UButton 
              v-if="currentStep > 0"
              icon="i-heroicons-arrow-left"
              :label="t('auditActivityPlan.stepper.previous')" 
              color="neutral" 
              variant="outline"
              @click="currentStep--"
            />

            <UButton 
              v-if="currentStep < steps.length - 1"
              trailing-icon="i-heroicons-arrow-right"
              :label="t('auditActivityPlan.stepper.next')" 
              color="primary" 
              class="font-bold shadow-md shadow-primary/20"
              @click="handleNext"
            />

            <UButton 
              v-else
              icon="i-heroicons-check"
              :label="store.isEditMode ? t('auditActivityPlan.form.update') : t('auditActivityPlan.form.create')" 
              color="primary" 
              class="font-bold shadow-md shadow-primary/20"
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
import { useActivityPlanStore } from '~/stores/activity-plan'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { AuditCategory, AuditDepartment } from '~/types/audit';
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const store = useActivityPlanStore()
const riskStore = useRiskProfileStore()

const currentStep = ref(0)
const stepError = ref('')

const steps = computed(() => [
  { key: 'basic', label: t('auditActivityPlan.stepper.step1'), icon: 'i-heroicons-information-circle' },
  { key: 'activities', label: t('auditActivityPlan.stepper.step2'), icon: 'i-heroicons-clipboard-document-check' },
  { key: 'resources', label: t('auditActivityPlan.stepper.step3'), icon: 'i-heroicons-user-group' },
  { key: 'review', label: t('auditActivityPlan.stepper.step4'), icon: 'i-heroicons-check-badge' },
  { key: 'attachment', label: t('auditActivityPlan.stepper.step5'), icon: 'i-heroicons-paper-clip' }
])

watch(() => store.isModalOpen, (isOpen) => {
  if (isOpen) {
    currentStep.value = 0
    stepError.value = ''
  }
})

riskStore.fetchRisks()

const validateStep = (stepIdx: number): boolean => {
  stepError.value = ''

  // Step 0: Basic Planning Information
  if (stepIdx === 0) {
    const s = store.formState
    if (!s.planTitle?.trim() || !s.planYear || !s.planPeriodStart || !s.planPeriodEnd || !s.department) {
      stepError.value = t('auditActivityPlan.validation.basicInfo')
      return false
    }
  }

  // Step 1: Planned Audit Activities
  if (stepIdx === 1) {
    const acts = store.formState.plannedActivities
    if (!acts || acts.length === 0) {
      stepError.value = t('auditActivityPlan.validation.plannedActivities')
      return false
    }
    for (const act of acts) {
      if (!act.auditName?.trim() || !act.auditee?.trim()) {
        stepError.value = t('auditActivityPlan.validation.plannedActivities')
        return false
      }
    }
  }

  // Step 3: Review & Approval
  if (stepIdx === 3) {
    const rev = store.formState.review
    if (!rev.creatorName?.trim() || !rev.creatorPosition?.trim() || !rev.approverName?.trim() || !rev.approverPosition?.trim() || !rev.approvalDate) {
      stepError.value = t('auditActivityPlan.validation.reviewAndApproval')
      return false
    }
  }

  return true
}

const goToStep = (targetIdx: number) => {
  stepError.value = ''
  if (targetIdx <= currentStep.value) {
    currentStep.value = targetIdx
    return
  }

  for (let i = currentStep.value; i < targetIdx; i++) {
    if (!validateStep(i)) {
      currentStep.value = i
      return
    }
  }
  currentStep.value = targetIdx
}

const handleNext = () => {
  if (validateStep(currentStep.value)) {
    currentStep.value++
  }
}

const getFilteredRisksForDept = (dept: string) => {
  if (!riskStore.risks || riskStore.risks.length === 0) {
    return []
  }
  return riskStore.risks.filter(r => {
    if (dept === 'IT') return r.category === 'Technology'
    if (dept === 'Finance') return r.category === 'Financial'
    if (dept === 'HR') return r.category === 'Human Resources'
    if (dept === 'Ops') return ['Operations', 'Compliance', 'Strategic', 'Governance'].includes(r.category)
    return true
  })
}

const getRiskLevelColorHex = (level?: string) => {
  if (!level) return '#9E9E9E'
  const lvl = level.toLowerCase()
  if (lvl.includes('high')) return '#F44336'
  if (lvl.includes('moderate to high')) return '#FF9800'
  if (lvl.includes('moderate')) return '#FFC107'
  if (lvl.includes('low to moderate')) return '#8BC34A'
  if (lvl.includes('low')) return '#4CAF50'
  return '#9E9E9E'
}

const handleFormSubmit = () => {
  for (let i = 0; i < steps.value.length; i++) {
    if (!validateStep(i)) {
      currentStep.value = i
      return
    }
  }
  store.savePlan()
}
</script>
