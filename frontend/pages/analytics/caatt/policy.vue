<script setup lang="ts">
import { onMounted } from 'vue'
import { useCaattAnalytics } from '~/composables/useCaattAnalytics'
import { useI18n } from '~/composables/useI18n'
import CaattHeader from '~/components/analytics/CaattHeader.vue'

const { t } = useI18n()
const { loading, policyState, fetchCaattAnalytics } = useCaattAnalytics()

onMounted(() => {
  fetchCaattAnalytics()
})
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <CaattHeader
      :title="t('navigation.caattPolicy')"
      :subtitle="t('analytics.caattUi.policy.subtitle')"
      icon="i-lucide-alert-triangle"
      :badgeText="t('analytics.caattUi.policy.badge')"
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
        icon="i-heroicons-shield-check"
        color="primary"
        variant="subtle"
        :title="t('analytics.caattUi.policy.title')"
        :description="t('analytics.caattUi.policy.description')"
      />

      <!-- Summary Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <UCard class="border border-indigo-100 dark:border-indigo-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.common.violationCount') }}</div>
          <div class="text-2xl font-black text-indigo-600 dark:text-indigo-400 mt-1">{{ policyState.summary.totalViolations }} {{ t('analytics.caattUi.common.cases') }}</div>
          <div class="text-[10px] text-gray-400 mt-0.5">{{ t('analytics.caattUi.policy.violatedRules') }}</div>
        </UCard>
        <UCard class="border border-rose-100 dark:border-rose-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.policy.criticalLevel') }}</div>
          <div class="text-2xl font-black text-rose-600 dark:text-rose-400 mt-1">{{ policyState.summary.critical }} {{ t('analytics.caattUi.common.cases') }}</div>
          <div class="text-[10px] text-rose-400 mt-0.5">{{ t('analytics.caattUi.policy.criticalHint') }}</div>
        </UCard>
        <UCard class="border border-amber-100 dark:border-amber-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.policy.highLevel') }}</div>
          <div class="text-2xl font-black text-amber-600 dark:text-amber-400 mt-1">{{ policyState.summary.high }} {{ t('analytics.caattUi.common.cases') }}</div>
          <div class="text-[10px] text-amber-400 mt-0.5">{{ t('analytics.caattUi.policy.highHint') }}</div>
        </UCard>
        <UCard class="border border-emerald-100 dark:border-emerald-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.common.affectedBranches') }}</div>
          <div class="text-2xl font-black text-emerald-600 dark:text-emerald-400 mt-1">{{ policyState.summary.branchesAffected }} {{ t('analytics.caattUi.common.branches') }}</div>
          <div class="text-[10px] text-gray-400 mt-0.5">{{ t('analytics.caattUi.policy.branchHint') }}</div>
        </UCard>
      </div>

      <!-- Table -->
      <UCard>
        <template #header>
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-table-cells" class="text-indigo-500" />
            <h3 class="font-bold">{{ t('analytics.caattUi.policy.findings') }}</h3>
          </div>
        </template>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">ID</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.isolation.colDate') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.policy.policyName') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.common.branch') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.policy.customerEntity') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.caattUi.policy.severity') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[260px]">{{ t('analytics.caattUi.policy.violationDescription') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('strategicPlan.columns.status') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in policyState.records" :key="r.id" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                <td class="py-3 px-3 font-mono font-bold">{{ r.id }}</td>
                <td class="py-3 px-3 text-gray-500 font-mono text-xs">{{ r.testDate }}</td>
                <td class="py-3 px-3 font-bold text-indigo-600 dark:text-indigo-400">{{ r.ruleName }}</td>
                <td class="py-3 px-3">{{ r.branchName }}</td>
                <td class="py-3 px-3 font-semibold">{{ r.customerName }}</td>
                <td class="text-center py-3 px-3">
                  <UBadge :color="r.severity === 'Critical' ? 'error' : r.severity === 'High' ? 'warning' : 'info'" variant="solid" size="sm">
                    {{ r.severity }}
                  </UBadge>
                </td>
                <td class="py-3 px-3 text-gray-600 dark:text-gray-300 leading-relaxed">{{ r.description }}</td>
                <td class="text-center py-3 px-3">
                  <UBadge :color="r.status === 'OPEN' ? 'error' : r.status === 'RESOLVED' ? 'success' : 'warning'" variant="subtle" size="sm">
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
