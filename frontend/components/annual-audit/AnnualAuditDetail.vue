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
                  <UBadge color="primary" variant="subtle" class="text-xl font-bold text-gray-800 ">
                    {{ store.selectedPlan.code }} <span class="text-gray-500 ml-2">({{ store.selectedPlan.version || 'v1.0' }})</span>
                  </UBadge>
                  <div class="flex items-center gap-2">
                    <span class="w-4 h-4 rounded-full" :class="store.getStatusColor(store.selectedPlan.status)"></span>
                    <span class="font-bold text-lg text-gray-800 ">{{ store.selectedPlan.status }}</span>
                  </div>
                </div>
                
                <div class="flex items-center gap-2">
                  <UIcon name="i-heroicons-x-mark" @click="store.closeViewModal" class="text-gray-400 hover:text-gray-600 text-3xl ml-2"></UIcon>
                </div>
              </div>

            <USeparator class="pt-6"/>

            <div class="space-y-6 pt-6">
              

              <UCard class="border border-gray-200  rounded-lg p-6">
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800 ">Activity Status</h3>
                </template>
                <div class="flex items-center gap-8">
                  <span class="font-bold text-gray-700  w-32">Progress</span>
                  <div class="flex-1 max-w-xl">
                    <UProgress v-model="store.progressAudit" color="secondary" class="h-3" />
                  </div>
                  <span class="text-secondary-600 font-bold">50 %</span>
                </div>
              </UCard>

              <UCard class="border border-gray-200  rounded-lg p-6">
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800 ">Activity Detail</h3>
                </template>
                
                <div class="space-y-6">
                  <div class="flex items-center pb-4 border-b border-gray-100 ">
                    <span class="font-bold text-gray-700  w-48">Activity Code</span>
                    <UBadge color="primary" variant="subtle" size="lg" class="text-primary-600 ">{{ store.selectedPlan.code }}</UBadge>
                  </div>

                  <UCard 
                    v-for="(activity, index) in (store.selectedPlan.activities || [])" 
                    :key="index"
                    class="p-4 bg-gray-50  rounded-lg space-y-3 border border-gray-100 "
                  >
                    <template #header>
                      <div class="flex items-center justify-between border-b border-gray-200  pb-2">
                        <span class="text-md uppercase text-gray-500 tracking-wider">
                          Sub-Activity {{ Number(index) + 1 }}
                        </span>
                      </div>
                    </template>
                    
                    <div class="flex items-start">
                      <span class="font-bold text-gray-600  w-44 text-sm">Activity Name</span>
                      <span class="font-semibold text-gray-800  flex-1">{{ activity.name }}</span>
                    </div>

                    <div class="flex items-center">
                      <span class="font-bold text-gray-600  w-44 text-sm">Category</span>
                      <UBadge size="md" color="primary" variant="subtle" class="font-bold">
                        {{ activity.category }}
                      </UBadge>
                    </div>

                    <div class="flex items-center">
                      <span class="font-bold text-gray-600  w-44 text-sm">Department</span>
                      <span class="font-semibold text-gray-800  flex items-center gap-1">
                        {{ activity.department }}
                      </span>
                    </div>

                    <div class="flex items-start">
                      <span class="font-bold text-gray-600  w-44 text-sm">Associated Risk</span>
                      <span class="font-semibold text-gray-800  flex-1">{{ activity.riskName || '-' }}</span>
                    </div>

                    <div class="flex items-center">
                      <span class="font-bold text-gray-600  w-44 text-sm">Risk Level</span>
                      <span v-if="activity.riskLevel" class="inline-flex items-center px-2 py-1 rounded-md text-xs font-semibold" :class="getRiskLevelColorClass(activity.riskLevel)">
                        {{ activity.riskLevel }}
                      </span>
                      <span v-else class="text-gray-400 text-sm">-</span>
                    </div>
                  </UCard>

                  <UCard v-if="!store.selectedPlan.activities?.length" class="text-center py-4 text-gray-400 italic">
                    There is no activity detail registered.
                  </UCard>
                </div>
              </UCard>

              <UCard class="border border-gray-200 dark:border-gray-800 rounded-lg p-4 sm:p-6">
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-2">2. Timeline</h3>
                </template>
                <div class="space-y-4">
                  <div class="flex flex-col sm:flex-row sm:items-center py-2 border-b border-gray-100 dark:border-gray-800/60">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">Year</span>
                    <span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedPlan.year }}</span>
                  </div>
                  <div class="flex flex-col sm:flex-row sm:items-center py-2 border-b border-gray-100 dark:border-gray-800/60">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">Quarter Distribution</span>
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
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">Execution Month</span>
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
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-2">3. Auditor Resources</h3>
                </template>
                <div class="space-y-4">
                  <div class="flex flex-col sm:flex-row sm:items-center py-2 border-b border-gray-100 dark:border-gray-800/60">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">Supervisor</span>
                    <span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedPlan?.supervisorName || store.getSupervisorName(store.selectedPlan?.supervisorId) }}</span>
                  </div>
                  <div class="flex flex-col sm:flex-row sm:items-center py-2 border-b border-gray-100 dark:border-gray-800/60">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">Time Allocation</span>
                    <div class="flex flex-wrap items-center gap-2">
                      <span class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-200 rounded-md text-sm font-medium">
                        <UIcon name="i-heroicons-user-group" class="w-4 h-4 text-primary-500" />
                        {{ store.selectedPlan.auditorCount }} Auditor
                      </span>
                      <span class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-200 rounded-md text-sm font-medium">
                        <UIcon name="i-heroicons-clock" class="w-4 h-4 text-amber-500" />
                        {{ store.selectedPlan.daysPerAuditor }} Days Duration
                      </span>
                    </div>
                  </div>
                  <div class="flex flex-col sm:flex-row sm:items-center py-2 border-b border-gray-100 dark:border-gray-800/60">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">Total Mandays</span>
                    <span class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-primary-50 dark:bg-primary-950/60 text-primary-700 dark:text-primary-300 font-bold rounded-md text-sm border border-primary-200 dark:border-primary-800">
                      <UIcon name="i-heroicons-bolt" class="w-4 h-4 text-primary-500" />
                      {{ (store.selectedPlan.auditorCount || 0) * (store.selectedPlan.daysPerAuditor || 0) }} Mandays
                    </span>
                  </div>
                  <div class="flex flex-col sm:flex-row sm:items-center py-2">
                    <span class="font-bold text-gray-600 dark:text-gray-400 w-full sm:w-48 text-sm shrink-0 mb-1 sm:mb-0">Estimated Capacity</span>
                    <div class="flex flex-wrap items-center gap-2">
                      <UBadge color="success" variant="subtle" size="sm">Optimal (60-80%)</UBadge>
                      <span class="text-sm text-gray-600 dark:text-gray-300">Total Load: 0.6% from Annual Capacity</span>
                    </div>
                  </div>
                </div>
              </UCard>

              <UCard class="border border-gray-200  rounded-lg p-6">
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800  mb-6">4. Additional Notes</h3>
                </template>
                <p class="font-semibold text-gray-800 ">{{ store.selectedPlan.notes || '-' }}</p>
              </UCard>

              <UCard class="border border-gray-200  rounded-lg p-6">
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800  mb-6">5. Attachment</h3>
                </template>
                <div class="space-y-4">
                  <div class="flex items-center"><span class="font-bold text-gray-700  w-48">File Category</span><span class="font-semibold text-gray-800 ">{{ store.selectedPlan.attachmentCategory }}</span></div>
                  <div class="flex items-center"><span class="font-bold text-gray-700  w-48">File Uploaded By</span><span class="font-semibold text-gray-800 ">{{ store.selectedPlan.attachmentUploadedBy }}</span></div>
                  <div class="flex items-center"><span class="font-bold text-gray-700  w-48">File Upload Date</span><span class="font-semibold text-gray-800 ">{{ store.selectedPlan.attachmentUploadDate }}</span></div>
                  
                  <div class="pt-2">
                    <span class="font-bold text-gray-700 ">Uploaded Files:</span>
                    <ul v-if="store.selectedPlan.attachments?.length" class="list-disc list-inside mt-2 space-y-2">
                      <li v-for="(file, index) in store.selectedPlan.attachments" :key="index" class="flex items-center justify-between p-2 bg-gray-50 rounded-md">
                        <div class="flex items-center gap-2">
                          <UIcon name="i-heroicons-document-text" class="text-gray-500" />
                          <span class="font-semibold text-gray-800">{{ file.name }}</span>
                          <UBadge color="neutral" variant="soft" size="md">{{ file.size }}</UBadge>
                        </div>
                        <UButton :to="file.url" target="_blank" icon="i-heroicons-arrow-down-tray" size="sm" color="primary" variant="link" label="Download" />
                      </li>
                    </ul>
                    <p v-else class="text-gray-500 italic mt-2">No files attached.</p>
                  </div>
                </div>
              </UCard>

              <UCard>
                <template #header>
                  <div class="flex items-center justify-between">
                    <h3 class="text-lg font-bold text-gray-800 ">Revision History</h3>
                    <UButton 
                      v-if="store.selectedPlan.status === 'Done' && !isCreatingRevision"
                      color="primary" 
                      variant="soft" 
                      icon="i-heroicons-document-duplicate" 
                      @click="() => { isCreatingRevision = true }"
                    >
                      Create Revised RKAT
                    </UButton>
                  </div>
                </template>
                
                <div v-if="isCreatingRevision" class="mb-6 p-4 border border-primary-200 bg-primary-50 rounded-lg">
                  <UFormField label="Revision Note / Changes Description">
                    <UTextarea v-model="revisionNote" placeholder="Describe the reason for this mid-year revision..." class="w-full" />
                  </UFormField>
                  <div class="mt-4 flex justify-end gap-2">
                    <UButton color="neutral" variant="ghost" @click="() => { isCreatingRevision = false }">Cancel</UButton>
                    <UButton color="primary" icon="i-heroicons-check" @click="submitRevision" :disabled="!revisionNote">Submit Revision</UButton>
                  </div>
                </div>

                <div v-if="store.selectedPlan.revisionHistory && store.selectedPlan.revisionHistory.length > 0">
                  <div class="space-y-4">
                    <div v-for="(rev, idx) in store.selectedPlan.revisionHistory" :key="idx" class="p-4 border rounded-lg bg-gray-50">
                      <div class="flex items-center justify-between mb-2">
                        <div class="flex items-center gap-2">
                          <UBadge color="primary" variant="subtle">{{ rev.version }}</UBadge>
                          <span class="text-sm text-gray-500">{{ rev.date }}</span>
                        </div>
                        <span class="text-sm font-semibold text-gray-700">{{ rev.user }}</span>
                      </div>
                      <p class="text-sm text-gray-600">{{ rev.changes }}</p>
                    </div>
                  </div>
                </div>
                <div v-else class="text-center p-4 text-gray-500 italic">
                  No revisions yet.
                </div>
              </UCard>

              <UCard>
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800 ">Approval Status</h3>
                </template>

                <UStepper 
                  :model-value="store.selectedPlan.status === 'Done' ? 'approved' : store.selectedPlan.status === 'Pending Approval' ? 'approval' : store.selectedPlan.status === 'Work In Progress' ? 'review' : 'draft'" 
                  :items="store.approvalStepperItems" 
                  class="w-full"
                >
                  <template #draft>
                    <UFormField v-if="authStore.user?.roles?.includes(UserRole.AUDIT_STAFF)" label="Audit Staff Note / Reason" class="p-4 mt-4 bg-gray-50 rounded-lg text-center border">
                      <UTextarea v-model="store.selectedPlan.staffApprovalNote" placeholder="Write your note or reason to approve" class="w-full"/>
                      <div class="mt-4 flex gap-2">
                        <UButton color="primary" icon="i-heroicons-check" @click="store.handleStaffApprove()">Approve</UButton>
                        <UButton color="error" variant="soft" icon="i-heroicons-x-mark" @click="store.handleStaffReject()">Reject</UButton>                  
                      </div>
                    </UFormField>
                    <div v-else class="p-4 mt-4 bg-gray-50 rounded-lg text-center border text-gray-500">
                      Waiting for Audit Staff approval...
                    </div>
                  </template>

                  <template #review>
                    <UFormField v-if="authStore.user?.roles?.includes(UserRole.AUDIT_MANAGER)" label="Audit Manager Note / Reason" class="p-4 mt-4 bg-gray-50 rounded-lg text-center border">
                      <UTextarea v-model="store.selectedPlan.managerApprovalNote" placeholder="Write your note or reason to approve" class="w-full"/>
                      <div class="mt-4 flex gap-2">
                        <UButton color="primary" icon="i-heroicons-check" @click="store.handleManagerApprove()">Approve</UButton>
                        <UButton color="error" variant="soft" icon="i-heroicons-x-mark" @click="store.handleManagerReject()">Reject</UButton>                  
                      </div>
                    </UFormField>
                    <div v-else class="p-4 mt-4 bg-gray-50 rounded-lg text-center border text-gray-500">
                      Waiting for Audit Manager approval...
                    </div>
                  </template>

                  <template #approval>
                    <UFormField v-if="authStore.user?.roles?.includes(UserRole.CHIEF_AUDIT_EXECUTIVE)" label="Chief Audit Executive Note / Reason" class="p-4 mt-4 bg-gray-50 rounded-lg text-center border">
                      <UTextarea v-model="store.selectedPlan.chiefApprovalNote" placeholder="Write your note or reason to approve" class="w-full"/>
                      <div class="mt-4 flex gap-2">
                        <UButton color="primary" icon="i-heroicons-check" @click="store.handleChiefApprove()">Approve</UButton>
                        <UButton color="error" variant="soft" icon="i-heroicons-x-mark" @click="store.handleChiefReject()">Reject</UButton>                  
                      </div>
                    </UFormField>
                    <div v-else class="p-4 mt-4 bg-gray-50 rounded-lg text-center border text-gray-500">
                      Waiting for Chief Audit Executive approval...
                    </div>
                  </template>

                  <template #approved>
                    <UAlert title="This Document has been approved." icon="i-heroicons-check-circle" color="success" />
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
import { ref, watch } from 'vue'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { getRiskLevelColorClass } from '~/utils/riskLevelBadge'
import { useAuthStore } from '~/stores/auth'
import { UserRole } from '~/types/auth'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useAnnualPlanStore()
const authStore = useAuthStore()

const isCreatingRevision = ref(false)
const revisionNote = ref('')

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