<template>
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Reguler Asesmen -->
      <UCard class="border-2 border-gray-100 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow">
        <div class="space-y-4">
          <div class="flex items-center space-x-3">
            <div class="w-4 h-4 rounded-full bg-amber-400"></div>
            <h3 class="text-lg font-bold">{{ t('qualityAssurance.dashboard.summary.regular') }}</h3>
          </div>
          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">{{ t('qualityAssurance.dashboard.summary.latestResult') }}</span>
              <span class="font-bold">{{ (store.summary.regular as any).result }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">{{ t('qualityAssurance.dashboard.summary.period') }}</span>
              <span class="font-bold">{{ (store.summary.regular as any).period }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">{{ t('qualityAssurance.dashboard.summary.status') }}</span>
              <div class="flex items-center space-x-2">
                <div :class="['w-3 h-3 rounded-full', store.getStatusColor((store.summary.regular as any).status)]"></div>
                <span class="font-bold text-sm">{{ formatStatus((store.summary.regular as any).status) }}</span>
              </div>
            </div>
          </div>
          <UButton
            variant="ghost"
            color="neutral"
            :label="t('qualityAssurance.dashboard.summary.viewDetail')"
            icon="i-lucide-chevron-right"
            trailing
            class="p-0 text-gray-400 hover:text-gray-900 dark:hover:text-white"
            @click="openCardDetail(store.summary.regular)"
          />
        </div>
      </UCard>

      <!-- QAR (PROFESSIONAL REVIEW) -->
      <UCard class="border-2 border-gray-100 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow">
        <div class="space-y-4">
          <div class="flex items-center space-x-3">
            <div class="w-4 h-4 rounded-full bg-white border border-gray-300"></div>
            <h3 class="text-lg font-bold">{{ t('qualityAssurance.dashboard.summary.qar') }}</h3>
          </div>
          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">{{ t('qualityAssurance.dashboard.summary.overallConclusion') }}</span>
              <span class="font-bold">{{ formatOverallConclusion((store.summary.qar as any).result) }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">{{ t('qualityAssurance.dashboard.summary.cycle') }}</span>
              <span class="font-bold">{{ (store.summary.qar as any).period }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">{{ t('qualityAssurance.dashboard.summary.status') }}</span>
              <div class="flex items-center space-x-2">
                <div :class="['w-3 h-3 rounded-full', store.getStatusColor((store.summary.qar as any).status)]"></div>
                <span class="font-bold text-sm">{{ formatStatus((store.summary.qar as any).status) }}</span>
              </div>
            </div>
          </div>
          <UButton
            variant="ghost"
            color="neutral"
            :label="t('qualityAssurance.dashboard.summary.viewDetail')"
            icon="i-lucide-certificate"
            trailing
            class="p-0 text-gray-400 hover:text-gray-900 dark:hover:text-white"
            @click="openCardDetail(store.summary.qar)"
          />
        </div>
      </UCard>

      <!-- SAIV (GIAS STANDARD) -->
      <UCard class="border-2 border-gray-100 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow">
        <div class="space-y-4">
          <div class="flex items-center space-x-3">
            <div class="w-4 h-4 rounded-full bg-blue-500"></div>
            <h3 class="text-lg font-bold">{{ t('qualityAssurance.dashboard.summary.saiv') }}</h3>
          </div>
          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">{{ t('qualityAssurance.dashboard.summary.overallConclusion') }}</span>
              <span class="font-bold">{{ formatOverallConclusion((store.summary.saiv as any).result) }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">{{ t('qualityAssurance.dashboard.summary.validator') }}</span>
              <span class="font-bold text-sm truncate max-w-[120px]">{{ (store.summary.saiv as any).validator || (store.summary.saiv as any).conductedBy || '-' }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">{{ t('qualityAssurance.dashboard.summary.status') }}</span>
              <div class="flex items-center space-x-2">
                <div :class="['w-3 h-3 rounded-full', store.getStatusColor((store.summary.saiv as any).status)]"></div>
                <span class="font-bold text-sm">{{ formatStatus((store.summary.saiv as any).status) }}</span>
              </div>
            </div>
          </div>
          <UButton
            variant="ghost"
            color="neutral"
            :label="t('qualityAssurance.dashboard.summary.viewDetail')"
            icon="i-lucide-chevron-right"
            trailing
            class="p-0 text-gray-400 hover:text-gray-900 dark:hover:text-white"
            @click="openCardDetail(store.summary.saiv)"
          />
        </div>
      </UCard>

      <!-- BUMN IACM (IACM SCALE) -->
      <UCard class="border-2 border-gray-100 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow">
        <div class="space-y-4">
          <div class="flex items-center space-x-3">
            <div class="w-4 h-4 rounded-full bg-indigo-500"></div>
            <h3 class="text-lg font-bold">{{ t('qualityAssurance.dashboard.summary.iacm') }}</h3>
          </div>
          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">{{ t('qualityAssurance.dashboard.summary.capabilityLevel') }}</span>
              <span class="font-bold text-indigo-600 text-lg">{{ formatIacmResult((store.summary.iacm as any).result) }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">{{ t('qualityAssurance.dashboard.summary.period') }}</span>
              <span class="font-bold">{{ (store.summary.iacm as any).period }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500 font-medium">{{ t('qualityAssurance.dashboard.summary.status') }}</span>
              <div class="flex items-center space-x-2">
                <div :class="['w-3 h-3 rounded-full', store.getStatusColor((store.summary.iacm as any).status)]"></div>
                <span class="font-bold text-sm">{{ formatStatus((store.summary.iacm as any).status) }}</span>
              </div>
            </div>
          </div>
          <UButton
            variant="ghost"
            color="neutral"
            :label="t('qualityAssurance.dashboard.summary.viewDetail')"
            icon="i-lucide-chevron-right"
            trailing
            class="p-0 text-gray-400 hover:text-gray-900 dark:hover:text-white"
            @click="openCardDetail(store.summary.iacm)"
          />
        </div>
      </UCard>
    </div>
</template>

<script setup lang="ts">
import { useQualityAssuranceStore, QAStatus } from '~/stores/quality-assurance'

const store = useQualityAssuranceStore()
const { t } = useI18n()

const formatStatus = (status: string) => {
  if (status === QAStatus.COMPLETED) return t('qualityAssurance.statuses.completed')
  if (status === QAStatus.VERIFIED) return t('qualityAssurance.statuses.verified')
  if (status === QAStatus.IN_PROGRESS) return t('qualityAssurance.statuses.inProgress')
  if (status === QAStatus.PLANNED) return t('qualityAssurance.statuses.planned')
  return status
}

const openCardDetail = (report: any) => {
  if (report && report.id) {
    store.openDetail(report)
  }
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

const formatIacmResult = (result?: string) => {
  if (!result || result === '-') return '-'
  const trimmed = result.trim()
  if (trimmed.includes('/')) return trimmed
  return `${trimmed} / 5`
}
</script>
