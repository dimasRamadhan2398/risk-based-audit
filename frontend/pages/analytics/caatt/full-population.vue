<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useCaattAnalytics } from '~/composables/useCaattAnalytics'
import { useI18n } from '~/composables/useI18n'
import CaattHeader from '~/components/analytics/CaattHeader.vue'

const { t } = useI18n()
const { loading, fptState, formatIDR, fetchCaattAnalytics } = useCaattAnalytics()

onMounted(() => {
  fetchCaattAnalytics()
})

const mobilePage = ref(1)
const mobilePageSize = ref(10)

const paginatedRecords = computed<any[]>(() => {
  const start = (mobilePage.value - 1) * mobilePageSize.value
  return (fptState.value?.records || []).slice(start, start + mobilePageSize.value)
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
        <!-- Desktop View (>= md): Full Table -->
        <div class="hidden md:block overflow-x-auto">
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

        <!-- Mobile & Tablet View (< md): Stacked Cards List -->
        <div class="block md:hidden">
          <!-- Empty State -->
          <div
            v-if="!fptState.records.length"
            class="py-8 text-center border border-dashed border-gray-200 dark:border-gray-800 rounded-xl bg-[var(--bg-surface)]"
          >
            <UIcon name="i-heroicons-table-cells" class="size-8 mx-auto text-gray-400 mb-2" />
            <p class="text-xs text-[var(--text-muted)]">{{ t('analytics.caattUi.fullPopulation.empty') || 'Tidak ada data temuan' }}</p>
          </div>

          <!-- Cards List -->
          <div v-else class="space-y-3">
            <div
              v-for="(r, index) in paginatedRecords"
              :key="r.id || index"
              class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
            >
              <!-- Card Header: ID, Category & Status Badge -->
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <span class="inline-flex items-center gap-1 font-mono text-[11px] font-bold text-slate-700 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded-md border border-slate-200/80 dark:border-slate-700">
                    <UIcon name="i-lucide-file-text" class="size-3 text-primary-500" />
                    {{ r.id }}
                  </span>
                  <UBadge color="neutral" variant="subtle" size="sm">{{ r.category }}</UBadge>
                </div>
                <UBadge :color="r.status === 'FLAGGED' ? 'error' : 'warning'" variant="solid" size="sm">
                  {{ r.status }}
                </UBadge>
              </div>

              <!-- Card Body: Customer & Account Info -->
              <div class="space-y-1">
                <div class="font-bold text-sm text-gray-900 dark:text-white leading-snug">
                  {{ r.customerName }}
                </div>
                <div class="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 flex-wrap gap-1">
                  <div class="flex items-center gap-1">
                    <UIcon name="i-lucide-building-2" class="size-3.5 text-slate-400 shrink-0" />
                    <span>{{ r.branchName }}</span>
                  </div>
                  <span class="font-mono text-[11px] text-slate-400">Acc: {{ r.accountId }}</span>
                </div>
              </div>

              <!-- Stat Comparison Grid: Amount vs Limit -->
              <div class="grid grid-cols-2 gap-2 pt-1">
                <div class="dashboard-subcard p-2.5 space-y-0.5 min-w-0 overflow-hidden">
                  <span class="text-[10px] uppercase font-bold tracking-wider text-slate-400 block truncate">{{ t('analytics.caattUi.fullPopulation.amount') }}</span>
                  <div class="font-mono font-bold text-xs sm:text-sm text-gray-900 dark:text-white break-words leading-tight">
                    {{ formatIDR(r.transactionAmount) }}
                  </div>
                </div>
                <div class="dashboard-subcard p-2.5 space-y-0.5 min-w-0 overflow-hidden">
                  <span class="text-[10px] uppercase font-bold tracking-wider text-slate-400 block truncate">{{ t('analytics.caattUi.fullPopulation.limit') }}</span>
                  <div class="font-mono font-semibold text-xs sm:text-sm text-slate-500 dark:text-slate-400 break-words leading-tight">
                    {{ formatIDR(r.thresholdLimit) }}
                  </div>
                </div>
              </div>

              <!-- Excess Amount Row -->
              <div class="flex items-center justify-between px-2.5 py-1.5 bg-rose-50/70 dark:bg-rose-950/20 rounded-md border border-rose-100 dark:border-rose-900/40 text-xs min-w-0 overflow-hidden">
                <span class="text-[11px] text-rose-600 dark:text-rose-400 font-medium shrink-0">{{ t('analytics.caattUi.fullPopulation.excess') }}</span>
                <span class="font-mono font-bold text-xs sm:text-sm text-rose-600 dark:text-rose-400 break-words text-right">+{{ formatIDR(r.excessAmount) }}</span>
              </div>

              <!-- Violation Type -->
              <div v-if="r.violationType" class="text-xs text-slate-600 dark:text-slate-300 leading-relaxed dashboard-subcard p-2.5">
                <span class="font-semibold text-slate-700 dark:text-slate-200">{{ t('analytics.caattUi.fullPopulation.violationType') }}: </span>
                <span>{{ r.violationType }}</span>
              </div>

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
              v-if="fptState.records.length > mobilePageSize"
              class="flex flex-col sm:flex-row items-center justify-between gap-3 pt-3 border-t border-slate-100 dark:border-slate-800"
            >
              <span class="text-xs text-slate-500 dark:text-slate-400">
                {{ (mobilePage - 1) * mobilePageSize + 1 }}-{{ Math.min(mobilePage * mobilePageSize, fptState.records.length) }} of {{ fptState.records.length }}
              </span>
              <UPagination
                v-model:page="mobilePage"
                :items-per-page="mobilePageSize"
                :total="fptState.records.length"
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
