<template>
    <UModal 
      v-model:open="store.showViewModal" 
      dismissible 
      :ui="{
        content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        header: 'border-b border-gray-100 dark:border-gray-800 p-4 sm:p-5 text-gray-900 dark:text-white font-bold shrink-0',
        body: 'p-4 sm:p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
        footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
      >
      <template #content>
        <div v-if="store.selectedPlan" class="relative rounded-xl shadow-2xl flex flex-col max-h-[90vh] overflow-y-auto p-4 sm:p-8">
          <template v-if="store.selectedPlan">
              <div class="flex justify-between items-center">
                <div class="flex items-center gap-4">
                  <UBadge color="primary" variant="subtle" class="text-xl font-bold text-gray-800 dark:text-gray-100">
                    {{ store.selectedPlan.code }} <span class="text-gray-500 dark:text-gray-400 ml-2">({{ store.selectedPlan.version || 'v1.0' }})</span>
                  </UBadge>
                  <div class="flex items-center gap-2">
                    <span class="w-4 h-4 rounded-full" :class="store.getStatusColor(store.selectedPlan.status)"></span>
                    <span class="font-bold text-lg text-gray-800 dark:text-white">{{ store.getStatusLabel(store.selectedPlan.status) }}</span>
                  </div>
                </div>

                <div class="flex items-center gap-2">
                  <UIcon name="i-heroicons-x-mark" @click="store.closeViewModal" class="text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-200 text-3xl ml-2"></UIcon>
                </div>
              </div>

            <USeparator class="pt-6"/>

            <div class="space-y-6 pt-6">


              <UCard class="border border-gray-200 dark:border-gray-800 rounded-lg p-6">
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white">{{ t('annualAudit.detail.activityDetail') }}</h3>
                </template>

                <div class="space-y-6">
                  <div class="flex items-center pb-4 border-b border-gray-100 dark:border-gray-800/60">
                    <span class="font-bold text-gray-700 dark:text-gray-300 w-48">{{ t('annualAudit.detail.activityCode') }}</span>
                    <UBadge color="primary" variant="subtle" size="lg" class="text-primary-600 dark:text-primary-400">{{ store.selectedPlan.code }}</UBadge>
                  </div>

                  <UCard
                    v-for="(activity, index) in (store.selectedPlan.activities || [])"
                    :key="index"
                    class="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg space-y-3 border border-gray-100 dark:border-gray-800"
                  >
                    <template #header>
                      <div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-700 pb-2">
                        <span class="text-md uppercase text-gray-500 dark:text-gray-400 tracking-wider">
                          {{ t('annualAudit.detail.subActivity', { number: Number(index) + 1 }) }}
                        </span>
                      </div>
                    </template>

                    <div class="flex items-start">
                      <span class="font-bold text-gray-600 dark:text-gray-400 w-44 text-sm">{{ t('annualAudit.detail.activityName') }}</span>
                      <span class="font-semibold text-gray-800 dark:text-gray-200 flex-1">{{ activity.name }}</span>
                    </div>

                    <div class="flex items-center">
                      <span class="font-bold text-gray-600 dark:text-gray-400 w-44 text-sm">{{ t('annualAudit.detail.category') }}</span>
                      <UBadge size="md" color="primary" variant="subtle" class="font-bold">
                        {{ activity.category }}
                      </UBadge>
                    </div>

                    <div class="flex items-center">
                      <span class="font-bold text-gray-600 dark:text-gray-400 w-44 text-sm">{{ t('annualAudit.detail.department') }}</span>
                      <span class="font-semibold text-gray-800 dark:text-gray-200 flex items-center gap-1">
                        {{ activity.department }}
                      </span>
                    </div>

                    <div class="flex items-start">
                      <span class="font-bold text-gray-600 dark:text-gray-400 w-44 text-sm">{{ t('annualAudit.detail.associatedRisk') }}</span>
                      <span class="font-semibold text-gray-800 dark:text-gray-200 flex-1">{{ activity.riskName || '-' }}</span>
                    </div>

                    <div class="flex items-center">
                      <span class="font-bold text-gray-600 dark:text-gray-400 w-44 text-sm">{{ t('annualAudit.detail.riskLevel') }}</span>
                      <span v-if="activity.riskLevel" class="inline-flex items-center px-2 py-1 rounded-md text-xs font-semibold" :class="getRiskLevelColorClass(activity.riskLevel)">
                        {{ activity.riskLevel }}
                      </span>
                      <span v-else class="text-gray-400 dark:text-gray-500 text-sm">-</span>
                    </div>

                    <div v-if="activity.activityPlanRef" class="flex items-start" data-testid="detail-used-in-activity-plan">
                      <span class="font-bold text-gray-600 dark:text-gray-400 w-44 text-sm shrink-0">{{ t('annualAudit.detail.activityPlan') }}</span>
                      <span class="font-semibold text-primary-700 dark:text-primary-300 flex-1">
                        {{ t('annualAudit.form.usedInActivityPlan', { title: activity.activityPlanRef.activityPlanTitle || '-', code: activity.activityPlanRef.assignmentCode || '-' }) }}
                      </span>
                    </div>
                  </UCard>

                  <UCard v-if="!store.selectedPlan.activities?.length" class="text-center py-4 text-gray-400 dark:text-gray-500 italic">
                    {{ t('annualAudit.detail.noActivityDetail') }}
                  </UCard>
                </div>
              </UCard>

              <UCard class="border border-gray-200 dark:border-gray-800 rounded-lg p-4 sm:p-6">
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-2">{{ t('annualAudit.detail.sectionTimeline') }}</h3>
                </template>
                <div class="space-y-4">
                  <div class="flex flex-col sm:flex-row sm:items-center py-2 border-b border-gray-100 dark:border-gray-800/60">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">{{ t('annualAudit.detail.year') }}</span>
                    <span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedPlan.year }}</span>
                  </div>
                  <div class="flex flex-col sm:flex-row sm:items-center py-2 border-b border-gray-100 dark:border-gray-800/60">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">{{ t('annualAudit.detail.quarterDistribution') }}</span>
                    <div class="flex flex-wrap items-center gap-1.5">
                      <template v-if="store.selectedPlan.quarters?.length">
                        <UBadge v-for="q in store.selectedPlan.quarters" :key="q" color="primary" variant="subtle" size="sm" class="font-semibold">
                          {{ q }}
                        </UBadge>
                      </template>
                      <span v-else class="text-gray-400 dark:text-gray-500 text-sm">-</span>
                    </div>
                  </div>
                  <div class="flex flex-col sm:flex-row sm:items-center py-2">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">{{ t('annualAudit.detail.executionMonth') }}</span>
                    <div class="flex flex-wrap items-center gap-1.5">
                      <template v-if="store.selectedPlan.selectedMonths?.length">
                        <UBadge 
                          v-for="m in store.selectedPlan.selectedMonths.slice().sort((a: number, b: number) => a - b)" 
                          :key="m" 
                          color="neutral" 
                          variant="subtle" 
                          size="sm"
                          class="font-medium"
                        >
                          {{ store.monthsList[m] || m }}
                        </UBadge>
                      </template>
                      <span v-else class="text-gray-400 dark:text-gray-500 text-sm">-</span>
                    </div>
                  </div>
                </div>
              </UCard>

              <UCard class="border border-gray-200 dark:border-gray-800 rounded-lg p-4 sm:p-6">
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-2">{{ t('annualAudit.detail.sectionAuditorResources') }}</h3>
                </template>
                <div class="space-y-4">
                  <div class="flex flex-col sm:flex-row sm:items-center py-2 border-b border-gray-100 dark:border-gray-800/60">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">{{ t('annualAudit.detail.supervisor') }}</span>
                    <span class="font-semibold text-gray-800 dark:text-gray-200">{{ teamLeadNames || '-' }}</span>
                  </div>
                  <div class="flex flex-col sm:flex-row sm:items-center py-2 border-b border-gray-100 dark:border-gray-800/60">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">{{ t('annualAudit.detail.timeAllocation') }}</span>
                    <div class="flex flex-wrap items-center gap-2">
                      <span class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-200 rounded-md text-sm font-medium">
                        <UIcon name="i-heroicons-user-group" class="w-4 h-4 text-primary-500" />
                        {{ t('annualAudit.detail.auditorsCount', { count: teamSize }) }}
                      </span>
                    </div>
                  </div>
                  <div class="flex flex-col sm:flex-row sm:items-center py-2 border-b border-gray-100 dark:border-gray-800/60">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">{{ t('annualAudit.detail.totalMandays') }}</span>
                    <span class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-primary-50 dark:bg-primary-950/60 text-primary-700 dark:text-primary-300 font-bold rounded-md text-sm border border-primary-200 dark:border-primary-800">
                      <UIcon name="i-heroicons-bolt" class="w-4 h-4 text-primary-500" />
                      {{ t('annualAudit.detail.mandaysValue', { count: teamMandays }) }}
                    </span>
                  </div>
                  <div class="flex flex-col sm:flex-row sm:items-center py-2">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">{{ t('annualAudit.detail.estimatedCapacity') }}</span>
                    <div class="flex flex-wrap items-center gap-2">
                      <UBadge color="success" variant="subtle" size="sm">{{ t('annualAudit.detail.capacityOptimal') }}</UBadge>
                      <span class="text-sm text-gray-600 dark:text-gray-300">{{ t('annualAudit.detail.totalLoad') }}</span>
                    </div>
                  </div>
                </div>
              </UCard>

              <UCard class="border border-gray-200 dark:border-gray-800 rounded-lg p-6">
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-6">{{ t('annualAudit.detail.sectionNotes') }}</h3>
                </template>
                <p class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedPlan.notes || '-' }}</p>
              </UCard>

              <UCard class="border border-gray-200 dark:border-gray-800 rounded-lg p-6">
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-6">{{ t('annualAudit.detail.sectionAttachment') }}</h3>
                </template>
                <div class="space-y-4">
                  <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">{{ t('annualAudit.detail.fileCategory') }}</span><span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedPlan.attachmentCategory }}</span></div>
                  <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">{{ t('annualAudit.detail.fileUploadedBy') }}</span><span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedPlan.attachmentUploadedBy }}</span></div>
                  <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">{{ t('annualAudit.detail.fileUploadDate') }}</span><span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedPlan.attachmentUploadDate }}</span></div>

                  <div class="pt-2">
                    <span class="font-bold text-gray-700 dark:text-gray-300">{{ t('annualAudit.detail.uploadedFiles') }}</span>
                    <ul v-if="store.selectedPlan.attachments?.length" class="list-disc list-inside mt-2 space-y-2">
                      <li v-for="(file, index) in store.selectedPlan.attachments" :key="index" class="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-800/50 rounded-md">
                        <div class="flex items-center gap-2">
                          <UIcon name="i-heroicons-document-text" class="text-gray-500 dark:text-gray-400" />
                          <span class="font-semibold text-gray-800 dark:text-gray-200">{{ file.name }}</span>
                          <UBadge color="neutral" variant="soft" size="md">{{ file.size }}</UBadge>
                        </div>
                        <UButton :to="file.url" target="_blank" icon="i-heroicons-arrow-down-tray" size="sm" color="primary" variant="link" :label="t('annualAudit.detail.download')" />
                      </li>
                    </ul>
                    <p v-else class="text-gray-500 dark:text-gray-400 italic mt-2">{{ t('annualAudit.detail.noFiles') }}</p>
                  </div>
                </div>
              </UCard>

              <UCard>
                <template #header>
                  <div class="flex items-center justify-between">
                    <h3 class="text-lg font-bold text-gray-800 dark:text-white">{{ t('annualAudit.detail.revisionHistory') }}</h3>
                    <UButton 
                      v-if="isApproved && !isCreatingRevision"
                      color="primary" 
                      variant="soft" 
                      icon="i-heroicons-document-duplicate" 
                      @click="() => { isCreatingRevision = true }"
                    >
                      {{ t('annualAudit.detail.createRevisedRkat') }}
                    </UButton>
                  </div>
                </template>
                
                <div v-if="isCreatingRevision" class="mb-6 p-4 border border-primary-200 dark:border-primary-800 bg-primary-50 dark:bg-primary-950/60 rounded-lg">
                  <UFormField :label="t('annualAudit.detail.revisionNoteLabel')">
                    <UTextarea v-model="revisionNote" :placeholder="t('annualAudit.detail.revisionNotePlaceholder')" class="w-full" />
                  </UFormField>
                  <div class="mt-4 flex justify-end gap-2">
                    <UButton color="neutral" variant="ghost" @click="() => { isCreatingRevision = false }">{{ t('common.cancel') }}</UButton>
                    <UButton color="primary" icon="i-heroicons-check" @click="submitRevision" :disabled="!revisionNote">{{ t('annualAudit.detail.submitRevision') }}</UButton>
                  </div>
                </div>

                <div v-if="store.selectedPlan.revisionHistory && store.selectedPlan.revisionHistory.length > 0">
                  <div class="space-y-4">
                    <div v-for="(rev, idx) in store.selectedPlan.revisionHistory" :key="idx" class="p-4 border dark:border-gray-700 rounded-lg bg-gray-50 dark:bg-gray-800/50">
                      <div class="flex items-center justify-between mb-2">
                        <div class="flex items-center gap-2">
                          <UBadge color="primary" variant="subtle">{{ rev.version }}</UBadge>
                          <span class="text-sm text-gray-500 dark:text-gray-400">{{ rev.date }}</span>
                        </div>
                        <span class="text-sm font-semibold text-gray-700 dark:text-gray-300">{{ rev.user }}</span>
                      </div>
                      <p class="text-sm text-gray-600 dark:text-gray-300">{{ rev.changes }}</p>
                    </div>
                  </div>
                </div>
                <div v-else class="text-center p-4 text-gray-500 dark:text-gray-400 italic">
                  {{ t('annualAudit.detail.noRevisions') }}
                </div>
              </UCard>

              <UCard>
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white">{{ t('annualAudit.detail.approvalStatus') }}</h3>
                </template>

                <UStepper 
                  :model-value="approvalStep"
                  :items="store.approvalStepperItems" 
                  value-key="value"
                  disabled
                  :ui="{ trigger: 'cursor-default pointer-events-none' }"
                  class="w-full"
                >
                  <template #draft>
                    <UFormField v-if="canApproveDraft" :label="t('annualAudit.detail.staffNoteLabel')" class="p-4 mt-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg text-center border dark:border-gray-700">
                      <UTextarea v-model="store.selectedPlan.staffApprovalNote" :placeholder="t('annualAudit.detail.approvalNotePlaceholder')" class="w-full"/>
                      <div class="mt-4 flex gap-2">
                        <UButton color="primary" icon="i-heroicons-check" @click="store.handleStaffApprove()">{{ t('annualAudit.detail.approve') }}</UButton>
                        <UButton color="error" variant="soft" icon="i-heroicons-x-mark" @click="store.handleStaffReject()">{{ t('annualAudit.detail.reject') }}</UButton>                  
                      </div>
                    </UFormField>
                    <div v-else class="p-4 mt-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg text-center border dark:border-gray-700 text-gray-500 dark:text-gray-400">
                      {{ t('annualAudit.detail.waitingStaff') }}
                    </div>
                  </template>

                  <template #review>
                    <UFormField v-if="canApproveReview" :label="t('annualAudit.detail.managerNoteLabel')" class="p-4 mt-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg text-center border dark:border-gray-700">
                      <UTextarea v-model="store.selectedPlan.managerApprovalNote" :placeholder="t('annualAudit.detail.approvalNotePlaceholder')" class="w-full"/>
                      <div class="mt-4 flex gap-2">
                        <UButton color="primary" icon="i-heroicons-check" @click="store.handleManagerApprove()">{{ t('annualAudit.detail.approve') }}</UButton>
                        <UButton color="error" variant="soft" icon="i-heroicons-x-mark" @click="store.handleManagerReject()">{{ t('annualAudit.detail.reject') }}</UButton>                  
                      </div>
                    </UFormField>
                    <div v-else class="p-4 mt-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg text-center border dark:border-gray-700 text-gray-500 dark:text-gray-400">
                      {{ t('annualAudit.detail.waitingManager') }}
                    </div>
                  </template>

                  <template #approval>
                    <UFormField v-if="canApproveChief" :label="t('annualAudit.detail.chiefNoteLabel')" class="p-4 mt-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg text-center border dark:border-gray-700">
                      <UTextarea v-model="store.selectedPlan.chiefApprovalNote" :placeholder="t('annualAudit.detail.approvalNotePlaceholder')" class="w-full"/>
                      <div class="mt-4 flex gap-2">
                        <UButton color="primary" icon="i-heroicons-check" @click="store.handleChiefApprove()">{{ t('annualAudit.detail.approve') }}</UButton>
                        <UButton color="error" variant="soft" icon="i-heroicons-x-mark" @click="store.handleChiefReject()">{{ t('annualAudit.detail.reject') }}</UButton>                  
                      </div>
                    </UFormField>
                    <div v-else class="p-4 mt-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg text-center border dark:border-gray-700 text-gray-500 dark:text-gray-400">
                      {{ t('annualAudit.detail.waitingChief') }}
                    </div>
                  </template>

                  <template #approved>
                    <UAlert
                      :title="t('annualAudit.detail.approvedAlert')"
                      :description="`${t('annualAudit.status.approved')} (${t('annualAudit.approvalRole.system')})`"
                      icon="i-heroicons-check-circle"
                      color="success"
                    />
                  </template>
                </UStepper>
              </UCard>
            </div>
          </template>
        </div>
      </template>
    </UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { getRiskLevelColorClass } from '~/utils/riskLevelBadge'
import { useAuthStore } from '~/stores/auth'
import { useRbac } from '~/composables/useRbac'
import { UserRole } from '~/types/auth'
import { AnnualAuditPlanStatus } from '~/types/audit'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useAnnualPlanStore()
const authStore = useAuthStore()
const { isAdmin, hasAnyRole } = useRbac()
const { t } = useI18n()

const canApproveDraft = computed(() =>
  isAdmin.value || hasAnyRole([UserRole.AUDITOR, 'auditstaff', 'auditor'])
)

const canApproveReview = computed(() =>
  isAdmin.value || hasAnyRole([UserRole.DEPARTMENT_HEAD, 'manager', 'auditmanager', 'audit_manager', 'dept_head'])
)

const canApproveChief = computed(() =>
  isAdmin.value || hasAnyRole([UserRole.EXECUTIVE, 'cae', 'chiefauditexecutive', 'chief_audit_executive', 'executive'])
)

// Team figures are the sum of the activity rows: the plan has no team-level totals.
const teamSize = computed(() => (store.selectedPlan?.activities || []).reduce((sum: number, a: { auditorCount?: number }) => sum + (Number(a.auditorCount) || 0), 0))
const teamMandays = computed(() => (store.selectedPlan?.activities || []).reduce((sum: number, a: { totalMandays?: number }) => sum + (Number(a.totalMandays) || 0), 0))
const teamLeadNames = computed(() => [...new Set((store.selectedPlan?.activities || []).map((a: { supervisorName?: string }) => a.supervisorName).filter(Boolean))].join(', '))

const isCreatingRevision = ref(false)
const revisionNote = ref('')

const isApproved = computed(() => store.selectedPlan?.status === AnnualAuditPlanStatus.APPROVED)

// Stepper slot for the current status. Draft and Revised both sit with Auditor (draft slot).
const approvalStep = computed(() => {
  switch (store.selectedPlan?.status) {
    case AnnualAuditPlanStatus.APPROVED: return 'approved'
    case AnnualAuditPlanStatus.PENDING_APPROVAL: return 'approval'
    case AnnualAuditPlanStatus.PENDING_REVIEW: return 'review'
    default: return 'draft'
  }
})

const submitRevision = () => {
  if (!revisionNote.value) return
  store.createRevision(store.selectedPlan.id, revisionNote.value, authStore.user?.fullName || 'System')
  isCreatingRevision.value = false
  revisionNote.value = ''
}

watch(() => store.showViewModal, (newVal) => {
  if (!newVal) {
    isCreatingRevision.value = false
    revisionNote.value = ''
  }
})

</script>