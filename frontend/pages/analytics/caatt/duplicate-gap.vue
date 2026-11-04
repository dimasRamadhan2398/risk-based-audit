<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useCaattAnalytics } from '~/composables/useCaattAnalytics'
import { useI18n } from '~/composables/useI18n'
import CaattHeader from '~/components/analytics/CaattHeader.vue'

const { t } = useI18n()
const { loading, dupGapState, formatIDR, fetchCaattAnalytics } = useCaattAnalytics()

onMounted(() => {
  fetchCaattAnalytics()
})

const mobilePage = ref(1)
const mobilePageSize = ref(10)

const paginatedRecords = computed<any[]>(() => {
  const start = (mobilePage.value - 1) * mobilePageSize.value
  return (dupGapState.value?.records || []).slice(start, start + mobilePageSize.value)
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
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 sm:gap-6">
        <!-- Card 1: Duplicates -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.duplicateGap.duplicates') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5 tabular-nums">
                {{ dupGapState.summary.totalDuplicates }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-rose-50 p-2.5 flex items-center justify-center dark:bg-rose-500/10">
              <UIcon name="i-lucide-copy-x" class="text-rose-600 size-5 dark:text-rose-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-rose-500 font-semibold">{{ t('analytics.caattUi.common.cases') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.duplicateGap.duplicateHint') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 2: Gaps -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.duplicateGap.gaps') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5 tabular-nums">
                {{ dupGapState.summary.totalGaps }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-amber-50 p-2.5 flex items-center justify-center dark:bg-amber-500/10">
              <UIcon name="i-lucide-binary" class="text-amber-600 size-5 dark:text-amber-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-amber-500 font-semibold">{{ t('analytics.caattUi.duplicateGap.missingNumbers') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.duplicateGap.gapHint') }}</span>
            </p>
          </div>
        </div>

        <!-- Card 3: Affected Branches -->
        <div class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-card">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
                {{ t('analytics.caattUi.common.affectedBranches') }}
              </span>
              <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5 tabular-nums">
                {{ dupGapState.summary.branchesAffected }}
              </h4>
            </div>
            <div class="shrink-0 rounded-xl bg-indigo-50 p-2.5 flex items-center justify-center dark:bg-indigo-500/10">
              <UIcon name="i-lucide-building-2" class="text-indigo-600 size-5 dark:text-indigo-400" />
            </div>
          </div>
          <div class="mt-3">
            <p class="text-[11px] text-slate-400 dark:text-slate-300 font-normal leading-normal">
              <span class="text-indigo-500 font-semibold">{{ t('analytics.caattUi.common.branches') }}</span>
              <span class="ml-1">{{ t('analytics.caattUi.duplicateGap.branchHint') }}</span>
            </p>
          </div>
        </div>
      </div>

      <!-- Table -->
      <UCard class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm ring-0 divide-gray-200 dark:divide-gray-800">
        <template #header>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-table-cells" class="text-indigo-500" />
              <h3 class="font-bold">{{ t('analytics.caattUi.duplicateGap.findings') }}</h3>
            </div>
          </div>
        </template>
        <!-- Desktop View (>= md): Full Table -->
        <div class="hidden md:block overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-100 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800">
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
                <td class="py-3 px-3 num-currency font-bold">{{ r.amount > 0 ? formatIDR(r.amount) : '-' }}</td>
                <td class="py-3 px-3 text-gray-600 dark:text-gray-300 leading-relaxed">{{ r.description }}</td>
                <td class="text-center py-3 px-3 tabular-nums font-bold">{{ r.occurrences }}x</td>
                <td class="text-center py-3 px-3">
                  <UBadge :color="r.status === 'OPEN' ? 'error' : r.status === 'INVESTIGATING' ? 'warning' : 'success'" variant="subtle" size="sm">
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
            v-if="!dupGapState.records.length"
            class="py-8 text-center border border-dashed border-gray-200 dark:border-gray-800 rounded-xl bg-[var(--bg-surface)]"
          >
            <UIcon name="i-heroicons-table-cells" class="size-8 mx-auto text-gray-400 mb-2" />
            <p class="text-xs text-[var(--text-muted)]">{{ t('analytics.caattUi.duplicateGap.empty') || 'Tidak ada data temuan duplikasi/gap' }}</p>
          </div>

          <!-- Cards List -->
          <div v-else class="space-y-3">
            <div
              v-for="(r, index) in paginatedRecords"
              :key="r.id || index"
              class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
            >
              <!-- Card Header: ID, Result Type & Status Badge -->
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <span class="inline-flex items-center gap-1 font-mono text-[11px] font-bold text-slate-700 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded-md border border-slate-200/80 dark:border-slate-700">
                    <UIcon name="i-lucide-hash" class="size-3 text-primary-500" />
                    {{ r.id }}
                  </span>
                  <UBadge :color="r.resultType === 'DUPLICATE' ? 'error' : 'warning'" variant="solid" size="sm">
                    {{ r.resultType }}
                  </UBadge>
                </div>
                <UBadge :color="r.status === 'OPEN' ? 'error' : r.status === 'INVESTIGATING' ? 'warning' : 'success'" variant="subtle" size="sm">
                  {{ r.status }}
                </UBadge>
              </div>

              <!-- Card Body: Branch, Reference & Account -->
              <div class="space-y-1">
                <div class="flex items-center gap-1.5 text-xs text-slate-500 dark:text-slate-400">
                  <UIcon name="i-lucide-building-2" class="size-3.5 text-slate-400 shrink-0" />
                  <span class="font-bold text-gray-900 dark:text-white text-sm truncate">{{ r.branchName }}</span>
                </div>
                <div class="flex items-center justify-between text-xs font-mono text-gray-500 flex-wrap gap-1 pt-0.5">
                  <span class="bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded text-[11px] font-bold text-slate-700 dark:text-slate-300 border border-slate-200/60 dark:border-slate-700">
                    Ref: {{ r.referenceNo }}
                  </span>
                  <span class="text-[11px] text-slate-400">Acc: {{ r.accountId }}</span>
                </div>
              </div>

              <!-- Stat Grid: Amount vs Frequency -->
              <div class="grid grid-cols-2 gap-2 pt-1">
                <div class="dashboard-subcard p-2.5 space-y-0.5 min-w-0 overflow-hidden">
                  <span class="text-[10px] uppercase font-bold tracking-wider text-slate-400 block truncate">{{ t('analytics.caattUi.duplicateGap.amount') }}</span>
                  <div class="font-mono font-bold text-xs sm:text-sm text-gray-900 dark:text-white break-words leading-tight">
                    {{ r.amount > 0 ? formatIDR(r.amount) : '-' }}
                  </div>
                </div>
                <div class="dashboard-subcard p-2.5 space-y-0.5 min-w-0 overflow-hidden">
                  <span class="text-[10px] uppercase font-bold tracking-wider text-slate-400 block truncate">{{ t('analytics.caattUi.duplicateGap.frequency') }}</span>
                  <div class="font-mono font-bold text-xs sm:text-sm text-warning-600 dark:text-warning-400 break-words leading-tight">
                    {{ r.occurrences }}x
                  </div>
                </div>
              </div>

              <!-- Description -->
              <p class="text-xs text-slate-600 dark:text-slate-300 leading-relaxed dashboard-subcard p-2.5">
                {{ r.description }}
              </p>

              <!-- Card Footer: Index & Result Type -->
              <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800/80 text-xs text-slate-500 dark:text-slate-400">
                <span class="text-xs text-slate-400 font-mono">
                  #{{ (mobilePage - 1) * mobilePageSize + Number(index) + 1 }}
                </span>
                <span class="text-[11px] text-gray-400">{{ r.resultType }}</span>
              </div>
            </div>

            <!-- Mobile Pagination Controls -->
            <div
              v-if="dupGapState.records.length > mobilePageSize"
              class="flex flex-col sm:flex-row items-center justify-between gap-3 pt-3 border-t border-slate-100 dark:border-slate-800"
            >
              <span class="text-xs text-slate-500 dark:text-slate-400">
                {{ (mobilePage - 1) * mobilePageSize + 1 }}-{{ Math.min(mobilePage * mobilePageSize, dupGapState.records.length) }} of {{ dupGapState.records.length }}
              </span>
              <UPagination
                v-model:page="mobilePage"
                :items-per-page="mobilePageSize"
                :total="dupGapState.records.length"
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
