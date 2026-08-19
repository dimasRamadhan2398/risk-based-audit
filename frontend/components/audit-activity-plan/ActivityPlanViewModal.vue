<template>
  <UModal v-model:open="store.isViewModalOpen" class="sm:max-w-4xl bg-[var(--bg-main)] border-[var(--border-main)]">
    <template #content>
    <div class="relative rounded-xl shadow-2xl flex flex-col max-h-[90vh] border">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 sticky top-0 rounded-t-xl z-10">
        <h3 class="text-xl font-bold text-[var(--text-main)]">
          {{ t('auditActivityPlan.view.title') }}
        </h3>
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-heroicons-x-mark-20-solid"
          class="-my-1"
          @click="store.closeViewModal"
        />
      </div>

      <!-- Content -->
      <div v-if="store.selectedPlan" class="flex-1 overflow-y-auto p-4 space-y-6">
        <!-- Basic Info -->
        <UCard>
            <template #header>
                <h4 class="text-lg font-medium">{{ t('auditActivityPlan.view.basicInfo') }}</h4>
            </template>
            <div class="space-y-3 text-sm">
                <div class="flex"><strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.planTitle') }}</strong> <span>{{ store.selectedPlan.planTitle }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.planYear') }}</strong> <span>{{ store.selectedPlan.planYear }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.planPeriod') }}</strong> <span>{{ store.selectedPlan.planPeriodStart }} {{ t('auditActivityPlan.form.to') }} {{ store.selectedPlan.planPeriodEnd }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.department') }}</strong> <UBadge variant="soft">{{ store.selectedPlan.department }}</UBadge></div>
                <div class="flex"><strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.createdBy') }}</strong> <span>{{ store.selectedPlan.createdBy }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.creationDate') }}</strong> <span>{{ new Date(store.selectedPlan.creationDate).toLocaleDateString(locale === 'id' ? 'id-ID' : 'en-US', { day: '2-digit', month: 'long', year: 'numeric' }) }}</span></div>
            </div>
        </UCard>

        <!-- Planned Activities -->
        <UCard>
            <template #header>
                <h4 class="text-lg font-medium">{{ t('auditActivityPlan.view.plannedActivities', { count: store.selectedPlan.plannedActivities.length }) }}</h4>
            </template>
            <UTable :data="store.selectedPlan.plannedActivities" :columns="plannedActivitiesColumns" />
        </UCard>

        <!-- Resource Auditors -->
        <UCard>
            <template #header>
                <h4 class="text-lg font-medium">{{ t('auditActivityPlan.view.resourceAuditors', { count: store.selectedPlan.resourceAuditors.length }) }}</h4>
            </template>
            <UTable :data="store.selectedPlan.resourceAuditors" :columns="resourceAuditorsColumns" />
        </UCard>

        <!-- Budget -->
        <UCard>
            <template #header>
                <h4 class="text-lg font-medium">{{ t('auditActivityPlan.view.budgetPlanning') }}</h4>
            </template>
            <div class="space-y-3 text-sm">
                <div class="flex"><strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.totalEstimatedCost') }}</strong> <span>{{ new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(store.selectedPlan.budget.totalEstimatedCost) }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.totalAllocatedBudget') }}</strong> <span>{{ new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(store.selectedPlan.budget.totalAllocatedBudget) }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.budgetNotes') }}</strong> <span class="flex-1">{{ store.selectedPlan.budget.budgetNotes || '-' }}</span></div>
            </div>
        </UCard>

        <!-- Review -->
        <UCard>
            <template #header>
                <h4 class="text-lg font-medium">{{ t('auditActivityPlan.view.reviewAndApproval') }}</h4>
            </template>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6 text-sm">
                <div class="space-y-2">
                    <p class="font-semibold border-b pb-1 mb-2">{{ t('auditActivityPlan.view.creator') }}</p>
                    <div class="flex"><strong class="w-24 shrink-0">{{ t('auditActivityPlan.view.name') }}</strong> <span>{{ store.selectedPlan.review.creatorName }}</span></div>
                    <div class="flex"><strong class="w-24 shrink-0">{{ t('auditActivityPlan.view.position') }}</strong> <span>{{ store.selectedPlan.review.creatorPosition }}</span></div>
                </div>
                <div class="space-y-2">
                    <p class="font-semibold border-b pb-1 mb-2">{{ t('auditActivityPlan.view.approver') }}</p>
                    <div class="flex"><strong class="w-24 shrink-0">{{ t('auditActivityPlan.view.name') }}</strong> <span>{{ store.selectedPlan.review.approverName }}</span></div>
                    <div class="flex"><strong class="w-24 shrink-0">{{ t('auditActivityPlan.view.position') }}</strong> <span>{{ store.selectedPlan.review.approverPosition }}</span></div>
                    <div class="flex"><strong class="w-24 shrink-0">{{ t('auditActivityPlan.view.date') }}</strong> <span>{{ new Date(store.selectedPlan.review.approvalDate).toLocaleDateString(locale === 'id' ? 'id-ID' : 'en-US', { day: '2-digit', month: 'long', year: 'numeric' }) }}</span></div>
                </div>
                <div class="col-span-2 mt-2">
                    <strong class="font-semibold">{{ t('auditActivityPlan.view.additionalNotes') }}</strong>
                    <p class="mt-1 text-gray-600 dark:text-gray-300">{{ store.selectedPlan.review.additionalNotes || '-' }}</p>
                </div>
            </div>
        </UCard>

        <!-- Attachments -->
        <UCard v-if="store.selectedPlan.attachmentCategory">
            <template #header>
                <h4 class="text-lg font-medium">{{ t('auditActivityPlan.view.attachment') }}</h4>
            </template>
            <div class="space-y-3 text-sm">
                <div class="flex"><strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.attachmentCategory') }}</strong> <span>{{ store.selectedPlan.attachmentCategory }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.uploadedBy') }}</strong> <span>{{ store.selectedPlan.attachmentUploadedBy }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.uploadDate') }}</strong> <span>{{ store.selectedPlan.attachmentUploadDate }}</span></div>
                
                <div class="pt-2">
                  <strong class="block mb-2 text-gray-700">{{ t('auditActivityPlan.view.uploadedFiles') }}</strong>
                  <ul v-if="store.selectedPlan.attachments?.length" class="list-disc list-inside space-y-2">
                    <li v-for="(file, index) in store.selectedPlan.attachments" :key="index" class="flex items-center justify-between p-2 bg-gray-50 rounded-md">
                      <div class="flex items-center gap-2">
                        <UIcon name="i-heroicons-document-text" class="text-gray-500" />
                        <span class="font-semibold text-gray-800">{{ file.name }}</span>
                        <UBadge color="neutral" variant="soft" size="md">{{ file.size }}</UBadge>
                      </div>
                      <UButton :to="file.url" target="_blank" icon="i-heroicons-arrow-down-tray" size="sm" color="primary" variant="link" :label="t('auditActivityPlan.view.download')" />
                    </li>
                  </ul>
                  <p v-else class="text-gray-500 italic mt-2">{{ t('auditActivityPlan.view.noFilesAttached') }}</p>
                </div>
            </div>
        </UCard>
      </div>
      <div v-else class="flex items-center justify-center h-32">
        <p class="text-gray-500">{{ t('auditActivityPlan.view.loading') }}</p>
      </div>

      <!-- Footer -->
      <div class="flex justify-end gap-3 p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] sticky bottom-0 rounded-b-xl">
        <UButton :label="t('auditActivityPlan.view.close')" color="neutral" variant="ghost" @click="store.closeViewModal" />
      </div>
    </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useActivityPlanStore } from '~/stores/activity-plan'
import { useI18n } from '~/composables/useI18n'

const { t, locale } = useI18n()
const store = useActivityPlanStore()

const plannedActivitiesColumns = computed(() => [
  { accessorKey: 'auditName', header: t('auditActivityPlan.view.columns.auditName') },
  { accessorKey: 'auditee', header: t('auditActivityPlan.view.columns.auditee') },
  { accessorKey: 'category', header: t('auditActivityPlan.view.columns.category') },
  { accessorKey: 'riskName', header: t('auditActivityPlan.view.columns.riskName') },
  { accessorKey: 'riskLevel', header: t('auditActivityPlan.view.columns.riskLevel') },
  { accessorKey: 'duration', header: t('auditActivityPlan.view.columns.duration') },
  { accessorKey: 'priority', header: t('auditActivityPlan.view.columns.priority') },
])

const resourceAuditorsColumns = computed(() => [
  { accessorKey: 'name', header: t('auditActivityPlan.view.columns.name') },
  { accessorKey: 'position', header: t('auditActivityPlan.view.columns.position') },
  { accessorKey: 'competence', header: t('auditActivityPlan.view.columns.competence') },
  { accessorKey: 'availability', header: t('auditActivityPlan.view.columns.availability') },
])
</script>