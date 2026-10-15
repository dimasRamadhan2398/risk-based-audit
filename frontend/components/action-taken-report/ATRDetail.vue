<template>
    <UModal 
      v-model:open="store.showModal" 
      scrollable 
      :ui="{
        content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        header: 'border-b border-gray-100 dark:border-gray-800 p-4 sm:p-5 text-gray-900 dark:text-white font-bold shrink-0',
        body: 'p-4 sm:p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
        footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
    >
      <template #content>
        <UCard 
          class="flex flex-col h-full max-h-[90vh] overflow-hidden"
          :ui="{ 
            root: 'flex flex-col h-full max-h-[90vh] overflow-hidden',
            header: 'px-4 sm:px-6 py-4 border-b border-gray-100 dark:border-gray-800 shrink-0',
            body: 'px-4 sm:px-6 py-4 overflow-y-auto flex-1 min-h-0',
            footer: 'px-4 sm:px-6 py-4 border-t border-gray-100 dark:border-gray-800 shrink-0'
          }"
        >
          <template #header>
            <div class="flex items-center justify-between">
              <div class="space-y-1">
                <h3 class="text-xl font-bold">{{ t('actionTakenReport.detail.title') }}</h3>
                <p class="text-sm font-bold">
                  {{ t('actionTakenReport.detail.referenceStatus', { reference: store.selectedReport?.auditRef || '-' }) }} [ <span :class="modalStatusDetails.class">
                    {{ modalStatusDetails.icon }} {{ store.selectedReport?.status }}
                  </span> ]
                </p>
              </div>
              <UButton
                color="neutral"
                variant="ghost"
                icon="i-lucide-x"
                class="-my-1"
                @click="store.closeModal"
              />
            </div>
          </template>

          <div class="space-y-6 pb-6">
            <!-- Assignment Letter Info -->
            <div v-if="store.selectedReport?.assignmentLetter || store.selectedReport?.assignment_letter" class="p-4 sm:p-5 rounded-xl space-y-4 bg-primary-600 dark:bg-primary-800">
              <div class="flex items-center justify-between">
                <h4 class="text-lg font-bold text-white">{{ t('actionTakenReport.detail.assignmentInfo') }}</h4>
                <UBadge color="neutral" variant="solid" class="bg-white/20 text-white border-0">
                  {{ (store.selectedReport?.assignmentLetter || store.selectedReport?.assignment_letter)?.letterNumber }}
                </UBadge>
              </div>
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-3 text-sm">
                <div>
                  <p class="font-bold text-primary-100">{{ t('actionTakenReport.detail.auditTitle') }}</p>
                  <p class="font-medium text-white">{{ (store.selectedReport?.assignmentLetter || store.selectedReport?.assignment_letter)?.auditTitle || '-' }}</p>
                </div>
                <div>
                  <p class="font-bold text-primary-100">{{ t('actionTakenReport.detail.workingUnit') }}</p>
                  <p class="font-medium text-white">{{ (store.selectedReport?.assignmentLetter || store.selectedReport?.assignment_letter)?.workingUnit || '-' }}</p>
                </div>
                <div>
                  <p class="font-bold text-primary-100">{{ t('actionTakenReport.detail.auditLeader') }}</p>
                  <p class="font-medium text-white">{{ (store.selectedReport?.assignmentLetter || store.selectedReport?.assignment_letter)?.leader || '-' }}</p>
                </div>
                <div>
                  <p class="font-bold text-primary-100">{{ t('actionTakenReport.detail.period') }}</p>
                  <p class="font-medium text-white">{{ (store.selectedReport?.assignmentLetter || store.selectedReport?.assignment_letter)?.executionPeriod || '-' }}</p>
                </div>
              </div>
            </div>

            <!-- Category & Context -->
            <div class="p-4 sm:p-5 border border-gray-200 dark:border-gray-700 rounded-xl space-y-4 bg-white dark:bg-gray-900">
              <h4 class="text-lg font-bold">{{ t('actionTakenReport.detail.categoryContext') }}</h4>
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-4">
                <div class="space-y-1">
                  <p class="text-sm font-bold">{{ t('actionTakenReport.detail.auditObject') }}</p>
                  <p class="text-sm">{{ store.selectedReport?.auditObject || '-' }}</p>
                </div>
                <div class="space-y-1">
                  <p class="text-sm font-bold">{{ t('actionTakenReport.detail.findingsCategory') }}</p>
                  <p class="text-sm">{{ store.selectedReport?.findingCategory || '-' }}</p>
                </div>
              </div>
            </div>

            <!-- Detail Findings -->
            <div class="p-4 sm:p-5 border border-gray-200 dark:border-gray-700 rounded-xl space-y-4 bg-white dark:bg-gray-900">
              <h4 class="text-lg font-bold">{{ t('actionTakenReport.detail.findingsDetail') }}</h4>
              <div class="space-y-4">
                <div class="flex flex-col sm:flex-row gap-1 sm:gap-4">
                  <p class="w-full sm:w-1/3 text-sm font-bold">{{ t('actionTakenReport.detail.condition') }}</p>
                  <p class="w-full sm:w-2/3 text-sm">{{ store.selectedReport?.condition || '-' }}</p>
                </div>
                <div class="flex flex-col sm:flex-row gap-1 sm:gap-4">
                  <p class="w-full sm:w-1/3 text-sm font-bold">{{ t('actionTakenReport.detail.criteria') }}</p>
                  <p class="w-full sm:w-2/3 text-sm">{{ store.selectedReport?.criteria || '-' }}</p>
                </div>
              </div>
            </div>

            <!-- Corrective Action Plan -->
            <div class="p-4 sm:p-5 border border-gray-200 dark:border-gray-700 rounded-xl space-y-4 bg-white dark:bg-gray-900">
              <h4 class="text-lg font-bold">{{ t('actionTakenReport.detail.correctivePlan') }}</h4>
              <div class="space-y-4">
                <div class="flex flex-col sm:flex-row gap-1 sm:gap-4">
                  <p class="w-full sm:w-1/3 text-sm font-bold">{{ t('actionTakenReport.detail.recommendation') }}</p>
                  <p class="w-full sm:w-2/3 text-sm">{{ store.selectedReport?.recommendation || '-' }}</p>
                </div>
                <div class="flex flex-col sm:flex-row gap-1 sm:gap-4">
                  <p class="w-full sm:w-1/3 text-sm font-bold">{{ t('actionTakenReport.detail.deadline') }}</p>
                  <p class="w-full sm:w-2/3 text-sm">{{ store.selectedReport?.deadline || '-' }}</p>
                </div>
                <div class="flex flex-col sm:flex-row gap-1 sm:gap-4">
                  <p class="w-full sm:w-1/3 text-sm font-bold">{{ t('actionTakenReport.detail.pic') }}</p>
                  <p class="w-full sm:w-2/3 text-sm">{{ store.selectedReport?.pic || '-' }}</p>
                </div>
              </div>
            </div>

            <div class="border-t border-gray-300 dark:border-gray-600 my-2"></div>

            <!-- Latest Progress Update -->
            <div class="p-4 sm:p-5 border border-gray-200 dark:border-gray-700 rounded-xl space-y-4 bg-white dark:bg-gray-900">
              <h4 class="text-lg font-bold">{{ t('actionTakenReport.detail.latestProgress') }}</h4>
              <div class="space-y-4">
                <div class="flex flex-col sm:flex-row gap-1 sm:gap-4">
                  <p class="w-full sm:w-1/3 text-sm font-bold">{{ t('actionTakenReport.detail.attachment') }}</p>
                  <div class="w-full sm:w-2/3 flex items-center space-x-1 text-sm font-bold cursor-pointer">
                    <span>[ 📄 {{ store.selectedReport?.attachment || t('actionTakenReport.detail.noAttachment') }} ]</span>
                  </div>
                </div>
                <div class="flex flex-col sm:flex-row gap-1 sm:gap-4">
                  <p class="w-full sm:w-1/3 text-sm font-bold">{{ t('actionTakenReport.detail.description') }}</p>
                  <p class="w-full sm:w-2/3 text-sm">{{ store.selectedReport?.progressDescription || '-' }}</p>
                </div>
              </div>
            </div>
          </div>

          <template #footer>
            <div class="flex justify-end">
              <UButton
                color="primary"
                :label="t('actionTakenReport.detail.sendReminder')"
                class="w-full sm:w-auto px-6 py-2.5 font-bold"
                @click="store.closeModal"
              />
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { AuditStatus } from '~/types/audit'

const store = useActionTakenReportStore()
const { t } = useI18n()

const modalStatusDetails = computed(() => {
  const status = store.selectedReport?.status
  switch (status) {
    case AuditStatus.COMPLETED: return { class: 'success', icon: '🟢' }
    case AuditStatus.IN_PROGRESS: return { class: 'warning', icon: '🟡' }
    case AuditStatus.PLANNED: return { class: 'neutral', icon: '⚪' }
    case AuditStatus.CANCELLED: return { class: 'error', icon: '🔴' }
    default: return { class: 'neutral', icon: '' }
  }
})

</script>
