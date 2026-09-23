<template>
  <UModal
    v-model:open="store.isDetailOpen"
    scrollable
    class="w-full sm:max-w-2xl bg-[var(--bg-main)] border-[var(--border-main)]"
  >
    <template #content>
      <UCard :ui="{ header: 'px-6 py-4', body: 'px-6 py-6', footer: 'px-6 py-4' }">
        <template #header>
          <div class="flex items-center space-x-4">
            <UButton
              icon="i-lucide-arrow-left"
              color="neutral"
              variant="ghost"
              :aria-label="t('common.close')"
              @click="store.closeDetail"
            />

            <h3 class="text-xl font-bold">
              {{ t('qualityAssurance.detailModal.title', { title: store.selectedReport?.assessmentTitle || '' }) }}
            </h3>
          </div>
        </template>

        <div class="space-y-10">
          <!-- Section 1 -->
          <div class="space-y-4">
            <h4 class="text-lg font-bold text-gray-900 dark:text-white">
              {{ t('qualityAssurance.detailModal.generalInfo') }}
            </h4>
            <div class="p-6 border border-gray-100 dark:border-gray-800 rounded-xl space-y-6">
              <div class="grid grid-cols-3 gap-4">
                <p class="font-bold text-gray-700">
                  {{ t('qualityAssurance.detailModal.assessmentTitle') }}
                </p>
                <p class="col-span-2 font-medium">
                  {{ store.selectedReport?.assessmentTitle }}
                </p>
              </div>
              <div class="grid grid-cols-3 gap-4">
                <p class="font-bold text-gray-700">
                  {{ t('qualityAssurance.detailModal.assessmentType') }}
                </p>
                <div class="col-span-2 flex items-center space-x-2">
                  <div :class="['w-4 h-4 rounded-full', store.getTypeIconColor(store.selectedReport?.type!)]" />
                  <p class="font-medium">
                    {{ getTypeLabel(store.selectedReport?.type!) }}
                  </p>
                </div>
              </div>
              <div class="grid grid-cols-3 gap-4">
                <p class="font-bold text-gray-700">
                  {{ t('qualityAssurance.detailModal.period') }}
                </p>
                <p class="col-span-2 font-medium">
                  {{ store.selectedReport?.period }}
                </p>
              </div>
            </div>
          </div>

          <!-- Section 2 -->
          <div class="space-y-4">
            <h4 class="text-lg font-bold text-gray-900 dark:text-white">
              {{ t('qualityAssurance.detailModal.results') }}
            </h4>
            <div class="p-6 border border-gray-100 dark:border-gray-800 rounded-xl space-y-6">
              <div class="grid grid-cols-3 gap-4">
                <p class="font-bold text-gray-700">
                  {{ t('qualityAssurance.detailModal.resultScore') }}
                </p>
                <p class="col-span-2 text-xl font-bold">
                  {{ (store.matchQAType(store.selectedReport?.type!, QAType.QAR) || store.matchQAType(store.selectedReport?.type!, QAType.SAIV)) ? formatOverallConclusion(store.selectedReport?.result!) : store.selectedReport?.result }}
                </p>
              </div>
              <div class="grid grid-cols-3 gap-4">
                <p class="font-bold text-gray-700">
                  {{ t('qualityAssurance.detailModal.status') }}
                </p>
                <div class="col-span-2 flex items-center space-x-2">
                  <div :class="['w-4 h-4 rounded-full', store.getStatusColor(store.selectedReport?.status!)]" />
                  <p class="font-bold">
                    {{ getStatusLabel(store.selectedReport?.status!) }} ({{ t('qualityAssurance.detailModal.finalProject') }})
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- Section 3 -->
          <div
            v-if="store.selectedReport?.validator || store.selectedReport?.internalEvaluator || store.selectedReport?.conductedBy"
            class="space-y-4"
          >
            <h4 class="text-lg font-bold text-gray-900 dark:text-white">
              {{ t('qualityAssurance.detailModal.specialDetails') }}
            </h4>
            <div class="p-6 border border-gray-100 dark:border-gray-800 rounded-xl space-y-6">
              <div
                v-if="store.selectedReport?.conductedBy"
                class="grid grid-cols-3 gap-4"
              >
                <p class="font-bold text-gray-700">
                  {{ t('qualityAssurance.detailModal.conductedBy') }}
                </p>
                <p class="col-span-2 font-bold text-orange-500">
                  {{ store.selectedReport?.conductedBy }}
                </p>
              </div>
              <div
                v-if="store.selectedReport?.validator || store.selectedReport?.internalEvaluator"
                class="grid grid-cols-3 gap-4"
              >
                <p class="font-bold text-gray-700">
                  {{ store.selectedReport?.validator ? t('qualityAssurance.detailModal.validator') : t('qualityAssurance.detailModal.internalEvaluator') }}
                </p>
                <p class="col-span-2 font-bold">
                  {{ store.selectedReport?.validator || store.selectedReport?.internalEvaluator }}
                </p>
              </div>
            </div>
          </div>

          <!-- Section 4 -->
          <div class="space-y-4">
            <h4 class="text-lg font-bold text-gray-900 dark:text-white">
              {{ t('qualityAssurance.detailModal.supportingDocuments') }}
            </h4>
            <div class="p-4 border border-gray-100 dark:border-gray-800 rounded-xl flex items-center justify-between bg-gray-50/50 dark:bg-gray-900/50">
              <div
                v-if="store.selectedReport?.attachment"
                class="flex items-center space-x-3"
              >
                <UIcon
                  name="i-lucide-file-text"
                  class="size-8 text-gray-400"
                />
                <div class="space-y-0.5">
                  <p class="font-bold text-sm">
                    {{ store.selectedReport.attachment.name }}
                  </p>
                  <p class="text-md text-gray-500">
                    {{ store.selectedReport.attachment.size }} • {{ t('qualityAssurance.detailModal.uploaded') }}
                  </p>
                </div>
              </div>
              <div
                v-else
                class="flex items-center space-x-3"
              >
                <UIcon
                  name="i-lucide-file-text"
                  class="size-8 text-gray-400"
                />
                <div class="space-y-0.5">
                  <p class="font-bold text-sm text-gray-400">
                    {{ t('qualityAssurance.detailModal.noAttachment') }}
                  </p>
                  <p class="text-md text-gray-400">
                    {{ t('qualityAssurance.detailModal.noAttachmentDesc') }}
                  </p>
                </div>
              </div>
              <div
                v-if="store.selectedReport?.attachment"
                class="flex items-center space-x-2"
              >
                <UButton
                  v-if="store.selectedReport.attachment.filePath"
                  icon="i-lucide-download"
                  :label="t('qualityAssurance.detailModal.download')"
                  color="neutral"
                  variant="ghost"
                  size="sm"
                  class="font-bold"
                  @click="store.downloadAttachment(store.selectedReport.id, store.selectedReport.attachment.name)"
                />
              </div>
            </div>
          </div>
        </div>
      </UCard>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useQualityAssuranceStore } from '~/stores/quality-assurance'
import type { QAStatus } from '~/types/quality-assurance'
import { QAType } from '~/types/quality-assurance'
import { useI18n } from '~/composables/useI18n'

const store = useQualityAssuranceStore()
const { t } = useI18n()

const getTypeLabel = (type: string) => {
  if (!type) return '-'
  if (store.matchQAType(type, QAType.REGULAR)) return t('qualityAssurance.types.regular')
  if (store.matchQAType(type, QAType.SAIV)) return t('qualityAssurance.types.saiv')
  if (store.matchQAType(type, QAType.QAR)) return t('qualityAssurance.types.qar')
  if (store.matchQAType(type, QAType.IACM)) return t('qualityAssurance.types.iacm')
  return type
}

const getStatusLabel = (status: QAStatus | string) => {
  if (!status) return '-'
  const s = status.toLowerCase()
  if (s.includes('complete') || s.includes('selesai')) {
    return t('qualityAssurance.statuses.completed')
  }
  if (s.includes('verif') || s.includes('terverifikasi')) {
    return t('qualityAssurance.statuses.verified')
  }
  if (s.includes('progress') || s.includes('berjalan')) {
    return t('qualityAssurance.statuses.inProgress')
  }
  if (s.includes('plan') || s.includes('rencana')) {
    return t('qualityAssurance.statuses.planned')
  }
  return status
}

const formatOverallConclusion = (result: string) => {
  if (!result) return '-'
  const res = result.trim().toLowerCase()
  if (res === 'g/c*' || res === 'gc' || res.includes('generally')) {
    return t('qualityAssurance.conformance.generallyConformed')
  }
  if (res === 'fc' || res.includes('fully') || res.includes('conformance')) {
    return t('qualityAssurance.conformance.fullyConformance')
  }
  if (res === 'pc' || res.includes('partially')) {
    return t('qualityAssurance.conformance.partiallyConform')
  }
  if (res === 'dnc' || res.includes('does not') || res.includes('doesnot')) {
    return t('qualityAssurance.conformance.doesNotConform')
  }
  return result
}
</script>
