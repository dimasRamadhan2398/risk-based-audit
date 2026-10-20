<script setup lang="ts">
import { onMounted } from 'vue'
import { useCaattAnalytics } from '~/composables/useCaattAnalytics'
import { useI18n } from '~/composables/useI18n'
import CaattHeader from '~/components/analytics/CaattHeader.vue'

const { t } = useI18n()
const { loading, fptState, formatIDR, fetchCaattAnalytics } = useCaattAnalytics()

onMounted(() => {
  fetchCaattAnalytics()
})
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <CaattHeader
      :title="t('navigation.caattFullPopulation')"
      :subtitle="t('analytics.caattUi.fullPopulation.subtitle')"
      icon="i-lucide-check-check"
      :badgeText="t('analytics.caattUi.fullPopulation.badge')"
      badgeColor="primary"
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
        icon="i-heroicons-clipboard-document-check"
        color="primary"
        variant="subtle"
        :title="t('analytics.caattUi.fullPopulation.title')"
        :description="t('analytics.caattUi.fullPopulation.description')"
      />

      <!-- Summary Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        <UCard class="border border-indigo-100 dark:border-indigo-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.fullPopulation.tested') }}</div>
          <div class="text-2xl font-black text-indigo-600 dark:text-indigo-400 mt-1">{{ Number(fptState.summary.totalTested).toLocaleString('id-ID') }}</div>
          <div class="text-[10px] text-gray-400 mt-0.5">{{ t('analytics.caattUi.fullPopulation.cbsPopulation') }}</div>
        </UCard>
        <UCard class="border border-rose-100 dark:border-rose-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.common.violationCount') }}</div>
          <div class="text-2xl font-black text-rose-600 dark:text-rose-400 mt-1">{{ fptState.summary.totalViolations }}</div>
          <div class="text-[10px] text-rose-400 mt-0.5">{{ t('analytics.caattUi.fullPopulation.followUp') }}</div>
        </UCard>
        <UCard class="border border-amber-100 dark:border-amber-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.fullPopulation.violationRatio') }}</div>
          <div class="text-2xl font-black text-amber-600 dark:text-amber-400 mt-1">{{ fptState.summary.avgViolationRate }}%</div>
          <div class="text-[10px] text-amber-400 mt-0.5">{{ t('analytics.caattUi.fullPopulation.tolerance') }}</div>
        </UCard>
        <UCard class="border border-emerald-100 dark:border-emerald-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.fullPopulation.branchesTested') }}</div>
          <div class="text-2xl font-black text-emerald-600 dark:text-emerald-400 mt-1">{{ fptState.summary.branchesTested }}</div>
          <div class="text-[10px] text-gray-400 mt-0.5">{{ t('analytics.caattUi.fullPopulation.allBranches') }}</div>
        </UCard>
        <UCard class="border border-violet-100 dark:border-violet-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.fullPopulation.categoriesTested') }}</div>
          <div class="text-2xl font-black text-violet-600 dark:text-violet-400 mt-1">{{ fptState.summary.categoriesTested }}</div>
          <div class="text-[10px] text-gray-400 mt-0.5">{{ t('analytics.caattUi.fullPopulation.categories') }}</div>
        </UCard>
      </div>

      <!-- Table -->
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-table-cells" class="text-indigo-500" />
              <h3 class="font-bold">{{ t('analytics.caattUi.fullPopulation.findings') }}</h3>
            </div>
            <UBadge color="primary" variant="subtle" size="sm">{{ t('analytics.caattUi.fullPopulation.badge') }}</UBadge>
          </div>
        </template>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">ID</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.colDate') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.common.branch') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.common.accountNumber') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.common.customerName') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.nlp.colCategory') }}</th>
                <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.fullPopulation.amount') }}</th>
                <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.fullPopulation.limit') }}</th>
                <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.fullPopulation.excess') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[200px]">{{ t('analytics.caattUi.fullPopulation.violationType') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('strategicPlan.columns.status') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in fptState.records" :key="r.id" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                <td class="py-3 px-3 font-mono font-bold">{{ r.id }}</td>
                <td class="py-3 px-3 text-gray-500 font-mono text-xs">{{ r.testDate }}</td>
                <td class="py-3 px-3 font-semibold">{{ r.branchName }}</td>
                <td class="py-3 px-3 font-mono text-xs text-gray-500">{{ r.accountId }}</td>
                <td class="py-3 px-3 font-semibold">{{ r.customerName }}</td>
                <td class="py-3 px-3">
                  <UBadge color="neutral" variant="subtle" size="sm">{{ r.category }}</UBadge>
                </td>
                <td class="py-3 px-3 text-right font-mono font-bold">{{ formatIDR(r.transactionAmount) }}</td>
                <td class="py-3 px-3 text-right font-mono text-gray-400">{{ formatIDR(r.thresholdLimit) }}</td>
                <td class="py-3 px-3 text-right font-mono font-bold text-rose-500">+{{ formatIDR(r.excessAmount) }}</td>
                <td class="py-3 px-3 text-gray-600 dark:text-gray-300">{{ r.violationType }}</td>
                <td class="text-center py-3 px-3">
                  <UBadge :color="r.status === 'FLAGGED' ? 'error' : 'warning'" variant="solid" size="sm">
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
