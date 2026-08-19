<template>
  <UModal v-model:open="store.isModalOpen" :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }">
    <div></div>
    <template #content>
    <div class="relative rounded-xl shadow-2xl flex flex-col max-h-[90vh] transition-colors duration-300">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 z-10 sticky top-0 rounded-t-xl transition-colors duration-300">
        <h3 class="text-xl font-bold text-[var(--text-main)]">
          {{ store.isEditMode ? t('auditActivityPlan.editPlan') : t('auditActivityPlan.form.createPlanTitle') }}
        </h3>
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-heroicons-x-mark-20-solid"
          class="-my-1"
          @click="store.closeModal"
        />
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-4 space-y-6">
        <UForm :state="store.formState" @submit="store.savePlan" class="space-y-6">
          
          <!-- Basic Planning Information -->
          <UCard :ui="{ body: 'px-4 py-5 sm:p-6' }">
            <template #header>
              <div class="flex justify-between items-center">
                <h3 class="text-lg font-medium">{{ t('auditActivityPlan.form.basicInfo') }}</h3>
              </div>
            </template>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField :label="t('auditActivityPlan.form.planTitle')" required>
                <UInput v-model="store.formState.planTitle" placeholder="Audit of the Year 2024" class="w-full"/>
              </UFormField>
              <UFormField :label="t('auditActivityPlan.form.planYear')" required>
                <UInput v-model="store.formState.planYear" class="w-full"/>
              </UFormField>
              <UFormField :label="t('auditActivityPlan.form.planPeriod')" required class="col-span-1 md:col-span-2 w-full">
                <div class="flex items-center gap-2">
                  <UInput v-model="store.formState.planPeriodStart" type="date" class="flex-1 w-full" />
                  <span class="text-gray-500">{{ t('auditActivityPlan.form.to') }}</span>
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
            <UFormField :label="t('auditActivityPlan.form.creationDate')" class="pt-4">
              <UInput v-model="store.formState.creationDate" type="date" class="w-full" readonly />
            </UFormField>
          </UCard>

          <!-- Planned Audit Activities -->
          <UCard :ui="{ body: 'px-4 py-5 sm:p-6' }">
            <template #header>
              <div class="flex justify-between items-center">
                <h3 class="text-lg font-medium">{{ t('auditActivityPlan.form.plannedActivities') }}</h3>
                <UButton color="warning" variant="solid" icon="i-heroicons-plus" :label="t('auditActivityPlan.form.addAuditActivity')" @click="store.addPlannedActivity" />
              </div>
            </template>

            <div class="space-y-6">
              <div v-for="(activity, index) in store.formState.plannedActivities" :key="activity.id" class="border border-[var(--border-main)] rounded-lg p-4 relative bg-[var(--bg-surface)] transition-all duration-300">
                <UButton
                  color="error"
                  variant="ghost"
                  icon="i-heroicons-trash"
                  class="absolute top-2 right-2"
                  @click="store.removePlannedActivity(index)"
                />
                <h4 class="font-medium mb-4">{{ t('auditActivityPlan.form.activityNum', { num: index + 1 }) }}</h4>
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
              
              <div v-if="store.formState.plannedActivities.length === 0" class="text-center text-gray-500 py-4">
                {{ t('auditActivityPlan.form.noActivitiesAdded') }}
              </div>

              <div class="mt-4 border-t pt-4">
                <h4 class="font-medium mb-2">{{ t('auditActivityPlan.form.summaryOfActivities') }}</h4>
                <div class="grid grid-cols-3 gap-4 text-center">
                  <div class="bg-red-50 text-red-600 rounded-lg p-2  ">
                    <div class="text-md font-semibold">{{ t('auditActivityPlan.form.highRisk') }}</div>
                    <div class="text-lg font-bold">{{ store.formState.plannedActivities.filter((a: any) => String(a.riskLevel) === 'High').length }}</div>
                  </div>
                  <div class="bg-yellow-50 text-yellow-600 rounded-lg p-2  ">
                    <div class="text-md font-semibold">{{ t('auditActivityPlan.form.mediumRisk') }}</div>
                    <div class="text-lg font-bold">{{ store.formState.plannedActivities.filter((a: any) => String(a.riskLevel) === 'Medium').length }}</div>
                  </div>
                  <div class="bg-green-50 text-green-600 rounded-lg p-2  ">
                    <div class="text-md font-semibold">{{ t('auditActivityPlan.form.lowRisk') }}</div>
                    <div class="text-lg font-bold">{{ store.formState.plannedActivities.filter((a: any) => String(a.riskLevel) === 'Low').length }}</div>
                  </div>
                </div>
              </div>
            </div>
          </UCard>

          <!-- Resources & Budget -->
          <UCard :ui="{ body: 'px-4 py-5 sm:p-6' }">
            <template #header>
              <div class="flex justify-between items-center">
                <h3 class="text-lg font-medium">{{ t('auditActivityPlan.form.resourcesAndBudget') }}</h3>
                <UButton color="warning" variant="solid" icon="i-heroicons-plus" :label="t('auditActivityPlan.form.addAuditor')" @click="store.addResourceAuditor" />
              </div>
            </template>
            
            <div class="space-y-6">
              <div v-for="(auditor, index) in store.formState.resourceAuditors" :key="auditor.id" class="border border-[var(--border-main)] rounded-lg p-4 relative bg-[var(--bg-surface)] transition-all duration-300">
                 <UButton
                  color="error"
                  variant="ghost"
                  icon="i-heroicons-trash"
                  class="absolute top-2 right-2"
                  @click="store.removeResourceAuditor(index)"
                />
                <h4 class="font-medium mb-4">{{ t('auditActivityPlan.form.auditorNum', { num: index + 1 }) }}</h4>
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

              <div class="mt-6 border-t pt-4">
                <h4 class="font-medium mb-4">{{ t('auditActivityPlan.form.budgetPlanning') }}</h4>
                <div class="grid grid-cols-1 gap-4">
                  <UFormField :label="t('auditActivityPlan.form.totalEstimatedCost')">
                    <UInput v-model="store.formState.budget.totalEstimatedCost" type="number" class="w-full"/>
                  </UFormField>
                  <UFormField :label="t('auditActivityPlan.form.totalAllocatedBudget')">
                    <UInput v-model="store.formState.budget.totalAllocatedBudget" type="number" class="w-full"/>
                  </UFormField>
                  <UFormField :label="t('auditActivityPlan.form.budgetNotes')">
                    <UTextarea v-model="store.formState.budget.budgetNotes" class="w-full"/>
                  </UFormField>
                </div>
              </div>
            </div>
          </UCard>

          <!-- Review & Approval -->
          <UCard :ui="{ body: 'px-4 py-5 sm:p-6' }">
            <template #header>
              <div class="flex justify-between items-center">
                <h3 class="text-lg font-medium">{{ t('auditActivityPlan.form.reviewAndApproval') }}</h3>
              </div>
            </template>
            <div class="space-y-6">
              <div>
                <h4 class="font-medium mb-2">{{ t('auditActivityPlan.view.creator') }}</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <UFormField :label="t('auditActivityPlan.form.creatorName')" required>
                    <UInput v-model="store.formState.review.creatorName" class="w-full"/>
                  </UFormField>
                  <UFormField :label="t('auditActivityPlan.form.creatorPosition')" required>
                    <UInput v-model="store.formState.review.creatorPosition" class="w-full"/>
                  </UFormField>
                </div>
              </div>
              <div class="border-t pt-4">
                <h4 class="font-medium mb-2">{{ t('auditActivityPlan.form.approvedBy') }}</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <UFormField :label="t('auditActivityPlan.form.approverName')" required>
                    <UInput v-model="store.formState.review.approverName" class="w-full"/>
                  </UFormField>
                  <UFormField :label="t('auditActivityPlan.form.approverPosition')" required>
                    <UInput v-model="store.formState.review.approverPosition" class="w-full"/>
                  </UFormField>
                </div>
                <UFormField :label="t('auditActivityPlan.form.approvalDate')" class="pt-4" required>
                  <UInput v-model="store.formState.review.approvalDate" type="date" class="w-full"/>
                </UFormField>
                <UFormField :label="t('auditActivityPlan.form.additionalNotes')" class="col-span-1 md:col-span-2 pt-4">
                  <UTextarea v-model="store.formState.review.additionalNotes" class="w-full"/>
                </UFormField>
                
              </div>
            </div>
          </UCard>

          <!-- Attachment -->
          <UCard :ui="{ body: 'px-4 py-5 sm:p-6' }">
            <template #header>
              <h3 class="text-lg font-medium">{{ t('auditActivityPlan.form.attachment') }}</h3>
            </template>
            <div class="space-y-4">
              <UFormField :label="t('auditActivityPlan.form.attachmentCategory')">
                <USelectMenu v-model="store.formState.attachmentCategory" :items="store.attachmentCategoryOptions" class="w-full"/>
              </UFormField>
              <UFormField :label="t('auditActivityPlan.form.attachmentUploadedBy')">
                <UInput v-model="store.formState.attachmentUploadedBy" placeholder="Example: Auditor" class="w-full" />
              </UFormField>
              <UFormField :label="t('auditActivityPlan.form.attachmentUploadDate')">
                <UInput type="date" v-model="store.formState.attachmentUploadDate" class="w-full"/>
              </UFormField>
              <UFormField :label="t('auditActivityPlan.form.uploadAttachmentHere')" size="lg">
                <UFileUpload
                  v-model="store.formState.file"
                  layout="list"
                  multiple
                  :label="t('auditActivityPlan.form.dropAttachmentsHere')"
                  :description="t('auditActivityPlan.form.maxFileSizeDesc')"
                  class="w-full"
                  :ui="{
                    base: 'min-h-48'
                  }"
                />
              </UFormField>
            </div>
          </UCard>

          <!-- Buttons -->
          <div class="flex justify-end gap-3 pb-6">
            <UButton :label="t('auditActivityPlan.form.cancel')" color="neutral" variant="ghost" @click="store.closeModal" />
            <UButton type="submit" :label="store.isEditMode ? t('auditActivityPlan.form.update') : t('auditActivityPlan.form.create')" color="warning" />
          </div>

        </UForm>
      </div>
    </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useActivityPlanStore } from '~/stores/activity-plan'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { AuditCategory, AuditDepartment } from '~/types/audit';
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const store = useActivityPlanStore()
const riskStore = useRiskProfileStore()

riskStore.fetchRisks()

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
</script>
