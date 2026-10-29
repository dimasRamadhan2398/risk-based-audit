<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useCaattAnalytics } from '~/composables/useCaattAnalytics'
import { useI18n } from '~/composables/useI18n'
import CaattHeader from '~/components/analytics/CaattHeader.vue'

const { t } = useI18n()
const { loading, policyState, fetchCaattAnalytics } = useCaattAnalytics()

onMounted(() => {
  fetchCaattAnalytics()
})

const mobilePage = ref(1)
const mobilePageSize = ref(10)

const paginatedRecords = computed<any[]>(() => {
  const start = (mobilePage.value - 1) * mobilePageSize.value
  return (policyState.value?.records || []).slice(start, start + mobilePageSize.value)
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
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
        <!-- Card 1: Total Violations -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.common.violationCount') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ policyState.summary.totalViolations }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-indigo-50 p-2.5 flex items-center justify-center dark:bg-indigo-500/10">
              <UIcon name="i-lucide-shield-alert" class="text-indigo-600 size-5 dark:text-indigo-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-indigo-500 font-semibold">{{ t('analytics.caattUi.common.cases') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.policy.violatedRules') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 2: Critical Level -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.policy.criticalLevel') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ policyState.summary.critical }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-rose-50 p-2.5 flex items-center justify-center dark:bg-rose-500/10">
              <UIcon name="i-lucide-alert-octagon" class="text-rose-600 size-5 dark:text-rose-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-rose-500 font-semibold">{{ t('analytics.caattUi.common.cases') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.policy.criticalHint') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 3: High Level -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.policy.highLevel') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ policyState.summary.high }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-amber-50 p-2.5 flex items-center justify-center dark:bg-amber-500/10">
              <UIcon name="i-lucide-alert-triangle" class="text-amber-600 size-5 dark:text-amber-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-amber-500 font-semibold">{{ t('analytics.caattUi.common.cases') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.policy.highHint') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 4: Affected Branches -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.common.affectedBranches') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
                {{ policyState.summary.branchesAffected }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-emerald-50 p-2.5 flex items-center justify-center dark:bg-emerald-500/10">
              <UIcon name="i-lucide-building-2" class="text-emerald-600 size-5 dark:text-emerald-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-emerald-500 font-semibold">{{ t('analytics.caattUi.common.branches') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.policy.branchHint') }}</span>
            </p>
          </div>
        </div>
      </div>

      <!-- Table -->
      <UCard>
        <template #header>
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-table-cells" class="text-primary-500" />
            <h3 class="font-bold">{{ t('analytics.caattUi.policy.findings') }}</h3>
          </div>
        </template>
        <!-- Desktop View (>= md): Full Table -->
        <div class="hidden md:block overflow-x-auto">
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
                <td class="py-3 px-3 font-bold text-primary-600 dark:text-primary-400">{{ r.ruleName }}</td>
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

        <!-- Mobile & Tablet View (< md): Stacked Cards List -->
        <div class="block md:hidden">
          <!-- Empty State -->
          <div
            v-if="!policyState.records.length"
            class="py-8 text-center border border-dashed border-gray-200 dark:border-gray-800 rounded-xl bg-[var(--bg-surface)]"
          >
            <UIcon name="i-heroicons-table-cells" class="size-8 mx-auto text-gray-400 mb-2" />
            <p class="text-xs text-[var(--text-muted)]">{{ t('analytics.caattUi.policy.empty') || 'Tidak ada data temuan kepatuhan' }}</p>
          </div>

          <!-- Cards List -->
          <div v-else class="space-y-3">
            <div
              v-for="(r, index) in paginatedRecords"
              :key="r.id || index"
              class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
            >
              <!-- Card Header: ID, Severity & Status Badge -->
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <span class="inline-flex items-center gap-1 font-mono text-[11px] font-bold text-slate-700 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded-md border border-slate-200/80 dark:border-slate-700">
                    <UIcon name="i-lucide-shield-alert" class="size-3 text-primary-500" />
                    {{ r.id }}
                  </span>
                  <UBadge :color="r.severity === 'Critical' ? 'error' : r.severity === 'High' ? 'warning' : 'info'" variant="solid" size="sm">
                    {{ r.severity }}
                  </UBadge>
                </div>
                <UBadge :color="r.status === 'OPEN' ? 'error' : r.status === 'RESOLVED' ? 'success' : 'warning'" variant="subtle" size="sm">
                  {{ r.status }}
                </UBadge>
              </div>

              <!-- Card Body: Rule Name & Customer/Branch Info -->
              <div class="space-y-1">
                <div class="font-bold text-sm text-primary-600 dark:text-primary-400 leading-snug">
                  {{ r.ruleName }}
                </div>
                <div class="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 flex-wrap gap-1">
                  <span class="font-semibold text-gray-900 dark:text-white">{{ r.customerName }}</span>
                  <div class="flex items-center gap-1">
                    <UIcon name="i-lucide-building-2" class="size-3.5 text-slate-400 shrink-0" />
                    <span>{{ r.branchName }}</span>
                  </div>
                </div>
              </div>

              <!-- Violation Description -->
              <p class="text-xs text-slate-600 dark:text-slate-300 leading-relaxed dashboard-subcard p-2.5">
                {{ r.description }}
              </p>

              <!-- Card Footer: Index & Date -->
              <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800/80 text-xs text-slate-500 dark:text-slate-400">
                <span class="text-xs text-slate-400 font-mono">
                  #{{ (mobilePage - 1) * mobilePageSize + Number(index) + 1 }}
                </span>
                <div class="flex items-center gap-1.5">
                  <UIcon name="i-lucide-calendar" class="size-3.5 text-slate-400 shrink-0" />
                  <span>{{ r.testDate }}</span>
                </div>
              </div>
            </div>

            <!-- Mobile Pagination Controls -->
            <div
              v-if="policyState.records.length > mobilePageSize"
              class="flex flex-col sm:flex-row items-center justify-between gap-3 pt-3 border-t border-slate-100 dark:border-slate-800"
            >
              <span class="text-xs text-slate-500 dark:text-slate-400">
                {{ (mobilePage - 1) * mobilePageSize + 1 }}-{{ Math.min(mobilePage * mobilePageSize, policyState.records.length) }} of {{ policyState.records.length }}
              </span>
              <UPagination
                v-model:page="mobilePage"
                :items-per-page="mobilePageSize"
                :total="policyState.records.length"
                size="md"
                active-color="primary"
                color="neutral"
                variant="outline"
                class="max-w-full overflow-x-auto"
              />
            </div>
          </div>
        </div>
      </UCard>
    </div>
  </div>
</template>
