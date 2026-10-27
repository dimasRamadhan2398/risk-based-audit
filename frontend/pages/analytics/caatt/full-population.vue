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
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 sm:gap-6">
        <!-- Card 1: CBS Population Tested -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.fullPopulation.tested') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ Number(fptState.summary.totalTested).toLocaleString('id-ID') }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-indigo-50 p-2.5 flex items-center justify-center dark:bg-indigo-500/10">
              <UIcon name="i-lucide-database" class="text-indigo-600 size-5 dark:text-indigo-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-indigo-500 font-semibold">{{ t('analytics.caattUi.common.transactions') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.fullPopulation.cbsPopulation') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 2: Violations -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.common.violationCount') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ fptState.summary.totalViolations }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-rose-50 p-2.5 flex items-center justify-center dark:bg-rose-500/10">
              <UIcon name="i-lucide-shield-alert" class="text-rose-600 size-5 dark:text-rose-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-rose-500 font-semibold">{{ t('analytics.caattUi.common.cases') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.fullPopulation.followUp') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 3: Violation Ratio -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.fullPopulation.violationRatio') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ fptState.summary.avgViolationRate }}%
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-amber-50 p-2.5 flex items-center justify-center dark:bg-amber-500/10">
              <UIcon name="i-lucide-activity" class="text-amber-600 size-5 dark:text-amber-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-amber-500 font-semibold">Toleransi</span>
              <span class="ml-1">{{ t('analytics.caattUi.fullPopulation.tolerance') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 4: Branches Tested -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.fullPopulation.branchesTested') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ fptState.summary.branchesTested }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-emerald-50 p-2.5 flex items-center justify-center dark:bg-emerald-500/10">
              <UIcon name="i-lucide-building-2" class="text-emerald-600 size-5 dark:text-emerald-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-emerald-500 font-semibold">{{ t('analytics.caattUi.common.branches') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.fullPopulation.allBranches') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 5: Categories Tested -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.fullPopulation.categoriesTested') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ fptState.summary.categoriesTested }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-violet-50 p-2.5 flex items-center justify-center dark:bg-violet-500/10">
              <UIcon name="i-lucide-layers" class="text-violet-600 size-5 dark:text-violet-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-violet-500 font-semibold">Kategori</span>
              <span class="ml-1">{{ t('analytics.caattUi.fullPopulation.categories') }}</span>
            </p>
          </div>
        </div>
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
