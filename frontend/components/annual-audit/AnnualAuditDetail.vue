<template>
    <UModal v-model:open="store.showViewModal" dismissible class="w-full sm:max-w-5xl bg-[var(--bg-main)] border-[var(--border-main)]">
      <template #content>
        <div v-if="store.selectedPlan" class="relative   rounded-xl shadow-2xl flex flex-col max-h-[95vh] overflow-y-auto p-8">
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
                      <UBadge v-if="activity.riskLevel" size="md" :color="store.getRiskLevelColor ? store.getRiskLevelColor(activity.riskLevel) : 'neutral'" variant="soft">
                        {{ activity.riskLevel }}
                      </UBadge>
                      <span v-else class="text-gray-400 text-sm">-</span>
                    </div>
                  </UCard>

                  <UCard v-if="!store.selectedPlan.activities?.length" class="text-center py-4 text-gray-400 italic">
                    There is no activity detail registered.
                  </UCard>
                </div>
              </UCard>

              <UCard class="border border-gray-200  rounded-lg p-6">
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800  mb-6">2. Timeline</h3>
                </template>
                <div class="space-y-4">
                  <div class="flex items-center"><span class="font-bold text-gray-700  w-48">Year</span><span class="font-semibold text-gray-800 ">{{ store.selectedPlan.year }}</span></div>
                  <div class="flex items-center"><span class="font-bold text-gray-700  w-48">Quarter Distribution</span><span class="font-semibold text-gray-800 ">{{ store.selectedPlan.quarters?.map((q: any) => `[ ${q} ]`).join(' ') || '-' }}</span></div>
                  <div class="flex items-center"><span class="font-bold text-gray-700  w-48">Execution Month</span><span class="font-semibold text-gray-800 ">{{ store.selectedPlan.selectedMonths?.slice().sort((a: number, b: number) => a - b).map((m: number) => `[ ${store.monthsList[m]} ]`).join(' ') || '-' }}</span></div>
                </div>
              </UCard>

              <UCard class="border border-gray-200  rounded-lg p-6">
                <template #header>
                  <h3 class="text-lg font-bold text-gray-800  mb-6">3. Auditor Resources</h3>
                </template>
                <div class="space-y-4">
                  <div class="flex items-center"><span class="font-bold text-gray-700  w-48">Supervisor</span><span class="font-semibold text-gray-800 ">{{ store.getSupervisorName(store.selectedPlan.supervisorId) }}</span></div>
                  <div class="flex items-center"><span class="font-bold text-gray-700  w-48">Time Allocation</span><span class="font-semibold text-gray-800 ">👥 {{ store.selectedPlan.auditorCount }} Auditor ⏱️ {{ store.selectedPlan.daysPerAuditor }} Days Duration</span></div>
                  <div class="flex items-center"><span class="font-bold text-gray-700  w-48">Total Mandays</span><span class="font-semibold text-gray-800 ">🔥 {{ store.selectedPlan.auditorCount * store.selectedPlan.daysPerAuditor }} Mandays</span></div>
                  <div class="flex items-center"><span class="font-bold text-gray-700  w-48">Estimated Capacity</span><span class="font-semibold text-gray-800 ">[🟢 Optimal (60-80%)] Total Load: 0.6% from Annual Capacity</span></div>
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