<script setup lang="ts">
import { onMounted } from 'vue'
import { useCaattAnalytics } from '~/composables/useCaattAnalytics'
import { useI18n } from '~/composables/useI18n'
import CaattHeader from '~/components/analytics/CaattHeader.vue'

const { t } = useI18n()
const { loading, dupGapState, formatIDR, fetchCaattAnalytics } = useCaattAnalytics()

onMounted(() => {
  fetchCaattAnalytics()
})
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <CaattHeader
      :title="t('navigation.caattDuplicateGap')"
      :subtitle="t('analytics.caattUi.duplicateGap.subtitle')"
      icon="i-lucide-copy-x"
      :badgeText="t('analytics.caattUi.duplicateGap.badge')"
      badgeColor="warning"
    />

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center items-center h-96">
      <div class="flex flex-col items-center gap-4">
        <UIcon name="i-heroicons-arrow-path" class="w-12 h-12 animate-spin text-indigo-500" />
        <span class="text-sm font-semibold text-gray-500 animate-pulse">{{ t('analytics.loading') }}</span>
      </div>
    </div>

    <!-- Content -->
    <div v-else class="space-y-6">
      <UAlert
        icon="i-heroicons-document-duplicate"
        color="primary"
        variant="subtle"
        :title="t('analytics.caattUi.duplicateGap.title')"
        :description="t('analytics.caattUi.duplicateGap.description')"
      />

      <!-- Summary Cards -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <UCard class="border border-rose-100 dark:border-rose-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.duplicateGap.duplicates') }}</div>
          <div class="text-2xl font-black text-rose-600 dark:text-rose-400 mt-1">{{ dupGapState.summary.totalDuplicates }} {{ t('analytics.caattUi.common.cases') }}</div>
          <div class="text-[10px] text-rose-400 mt-0.5">{{ t('analytics.caattUi.duplicateGap.duplicateHint') }}</div>
        </UCard>
        <UCard class="border border-amber-100 dark:border-amber-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.duplicateGap.gaps') }}</div>
          <div class="text-2xl font-black text-amber-600 dark:text-amber-400 mt-1">{{ dupGapState.summary.totalGaps }} {{ t('analytics.caattUi.duplicateGap.missingNumbers') }}</div>
          <div class="text-[10px] text-amber-400 mt-0.5">{{ t('analytics.caattUi.duplicateGap.gapHint') }}</div>
        </UCard>
        <UCard class="border border-indigo-100 dark:border-indigo-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.common.affectedBranches') }}</div>
          <div class="text-2xl font-black text-indigo-600 dark:text-indigo-400 mt-1">{{ dupGapState.summary.branchesAffected }} {{ t('analytics.caattUi.common.branches') }}</div>
          <div class="text-[10px] text-gray-400 mt-0.5">{{ t('analytics.caattUi.duplicateGap.branchHint') }}</div>
        </UCard>
      </div>

      <!-- Table -->
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-table-cells" class="text-indigo-500" />
              <h3 class="font-bold">{{ t('analytics.caattUi.duplicateGap.findings') }}</h3>
            </div>
          </div>
        </template>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">ID</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.xgboost.colType') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.common.branch') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.duplicateGap.reference') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.common.accountNumber') }}</th>
                <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.duplicateGap.amount') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[260px]">{{ t('analytics.caattUi.duplicateGap.findingDescription') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.duplicateGap.frequency') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('strategicPlan.columns.status') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in dupGapState.records" :key="r.id" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                <td class="py-3 px-3 font-mono font-bold">{{ r.id }}</td>
                <td class="text-center py-3 px-3">
                  <UBadge :color="r.resultType === 'DUPLICATE' ? 'error' : 'warning'" variant="solid" size="sm">
                    {{ r.resultType }}
                  </UBadge>
                </td>
                <td class="py-3 px-3 font-semibold">{{ r.branchName }}</td>
                <td class="py-3 px-3 font-mono font-bold">{{ r.referenceNo }}</td>
                <td class="py-3 px-3 font-mono text-xs text-gray-500">{{ r.accountId }}</td>
                <td class="py-3 px-3 text-right font-mono font-bold">{{ r.amount > 0 ? formatIDR(r.amount) : '-' }}</td>
                <td class="py-3 px-3 text-gray-600 dark:text-gray-300 leading-relaxed">{{ r.description }}</td>
                <td class="text-center py-3 px-3 font-mono font-bold">{{ r.occurrences }}x</td>
                <td class="text-center py-3 px-3">
                  <UBadge :color="r.status === 'OPEN' ? 'error' : r.status === 'INVESTIGATING' ? 'warning' : 'success'" variant="subtle" size="sm">
                    {{ r.status }}
                  </UBadge>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </UCard>
    </div>
  </div>
</template>
