<template>
  <div class="p-4 sm:p-6 max-w-full mx-auto space-y-8 min-w-0">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-[var(--border-main)] pb-5">
      <div>
        <h1 class="text-2xl sm:text-3xl font-extrabold tracking-tight text-slate-900 dark:text-white font-space">
          {{ t('auditPriority.title') !== 'auditPriority.title' ? t('auditPriority.title') : t('navigation.auditPriority') }}
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
          {{ t('riskFactors.priority.subtitle') }}
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-2 sm:gap-3">
        <UButton
          to="/risk-profile"
          icon="i-lucide-arrow-left"
          color="neutral"
          variant="outline"
          class="w-full sm:w-auto"
        >
          {{ t('riskFactors.backToHeatmap') }}
        </UButton>
        <UButton
          to="/risk-profile/audit-universe"
          icon="i-lucide-globe"
          color="neutral"
          variant="outline"
          class="w-full sm:w-auto"
        >
          {{ t('auditPriority.goToUniverse') !== 'auditPriority.goToUniverse' ? t('auditPriority.goToUniverse') : 'Audit Universe' }}
        </UButton>
        <UButton
          to="/risk-profile/risk-factors"
          icon="i-lucide-activity"
          color="neutral"
          variant="outline"
          class="w-full sm:w-auto"
        >
          {{ t('auditPriority.goToScoring') !== 'auditPriority.goToScoring' ? t('auditPriority.goToScoring') : 'Risk Factors & Scoring' }}
        </UButton>
      </div>
    </div>

    <!-- Alert / Toast -->
    <Transition name="fade">
      <UAlert
        v-if="alertMessage"
        :color="alertType === 'success' ? 'success' : 'error'"
        variant="solid"
        :title="alertType === 'success' ? 'Success' : 'Error'"
        :description="alertMessage"
        icon="i-lucide-info"
        class="shadow-md"
        closable
        @close="alertMessage = ''"
      />
    </Transition>

    <!-- Top KPI & Year Controls -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <!-- Year Selector Card -->
      <UCard class="shadow-sm border border-[var(--border-main)] flex flex-col justify-center">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              {{ t('auditPriority.selectYear') !== 'auditPriority.selectYear' ? t('auditPriority.selectYear') : 'Audit Year' }}
            </p>
            <div class="flex items-center gap-2 mt-1.5">
              <USelect
                v-model.number="selectedYear"
                :items="fiscalYears"
                size="sm"
                class="w-28 font-bold"
              />
              <UButton
                icon="i-lucide-refresh-cw"
                color="neutral"
                variant="ghost"
                size="sm"
                :loading="loading"
                :title="t('auditPriority.refresh') !== 'auditPriority.refresh' ? t('auditPriority.refresh') : 'Refresh'"
                @click="loadData"
              />
            </div>
          </div>
          <div class="p-3 bg-primary-50 dark:bg-primary-950/40 rounded-xl text-primary-600 dark:text-primary-400">
            <UIcon name="i-lucide-calendar" class="w-6 h-6" />
          </div>
        </div>
      </UCard>

      <!-- Total Established Entities -->
      <UCard class="shadow-sm border border-[var(--border-main)]">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              {{ t('auditPriority.totalEntities') !== 'auditPriority.totalEntities' ? t('auditPriority.totalEntities') : 'Established Entities' }}
            </p>
            <h3 class="text-2xl font-bold text-slate-800 dark:text-slate-100 mt-1">
              {{ yearlyUniverse.length }}
            </h3>
            <p class="text-xs text-slate-400 mt-0.5">Year {{ selectedYear }}</p>
          </div>
          <div class="p-3 bg-slate-100 dark:bg-slate-800 rounded-xl text-slate-600 dark:text-slate-300">
            <UIcon name="i-lucide-layers" class="w-6 h-6" />
          </div>
        </div>
      </UCard>

      <!-- Prioritized Entities -->
      <UCard class="shadow-sm border border-[var(--border-main)]">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-emerald-600 dark:text-emerald-400 uppercase tracking-wider">
              {{ t('auditPriority.prioritizedEntities') !== 'auditPriority.prioritizedEntities' ? t('auditPriority.prioritizedEntities') : 'Prioritized' }}
            </p>
            <div class="flex items-baseline gap-2 mt-1">
              <h3 class="text-2xl font-bold text-emerald-600 dark:text-emerald-400">
                {{ prioritizedCount }}
              </h3>
              <span v-if="yearlyUniverse.length > 0" class="text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                ({{ ((prioritizedCount / yearlyUniverse.length) * 100).toFixed(0) }}%)
              </span>
            </div>
            <p class="text-xs text-slate-400 mt-0.5">
              {{ t('riskFactors.priority.prioritizedCount', { count: prioritizedCount }) }}
            </p>
          </div>
          <div class="p-3 bg-emerald-50 dark:bg-emerald-950/40 rounded-xl text-emerald-600 dark:text-emerald-400">
            <UIcon name="i-lucide-check-circle" class="w-6 h-6" />
          </div>
        </div>
      </UCard>

    </div>

    <!-- Main Table Card -->
    <UCard class="shadow-sm border border-[var(--border-main)]">
      <template #header>
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h2 class="text-lg font-bold text-slate-800 dark:text-slate-100 font-space">
              {{ t('riskFactors.priority.title', { year: selectedYear }) }}
            </h2>
            <p class="text-xs text-slate-500 mt-0.5">
              {{ t('riskFactors.priority.subtitle') }}
            </p>
          </div>
          <!-- Search & Filter Controls -->
          <div class="flex flex-wrap items-center gap-3 w-full sm:w-auto">
            <UInput
              v-model="searchQuery"
              icon="i-lucide-search"
              size="sm"
              :placeholder="t('auditPriority.searchPlaceholder') !== 'auditPriority.searchPlaceholder' ? t('auditPriority.searchPlaceholder') : 'Search auditable entities...'"
              class="w-full sm:w-64"
            />
            <USelect
              v-model="statusFilter"
              :items="statusFilterOptions"
              size="sm"
              class="w-full sm:w-44"
            />
            <UBadge color="success" variant="subtle" class="font-bold">
              {{ prioritizedCount }} Prioritized
            </UBadge>
          </div>
        </div>
      </template>

      <!-- Table Content -->
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-sm">
          <thead class="bg-slate-50 dark:bg-slate-850/50">
            <tr>
              <th scope="col" class="px-6 py-3 text-left font-semibold text-slate-700 dark:text-slate-300">
                {{ t('riskFactors.priority.no') }}
              </th>
              <th scope="col" class="px-6 py-3 text-left font-semibold text-slate-700 dark:text-slate-300">
                {{ t('riskFactors.priority.auditableEntity') }}
              </th>
              <th scope="col" class="px-6 py-3 text-center scope font-semibold text-slate-700 dark:text-slate-300">
                {{ t('riskFactors.priority.riskIndex') }}
              </th>
              <th scope="col" class="px-6 py-3 text-center scope font-semibold text-slate-700 dark:text-slate-300">
                {{ t('riskFactors.priority.riskLevel') }}
              </th>
              <th scope="col" class="px-6 py-3 text-center scope font-semibold text-slate-700 dark:text-slate-300">
                {{ t('riskFactors.priority.auditPriorityCol') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
            <tr
              v-for="(ent, idx) in filteredYearlyUniverse"
              :key="ent.id"
              class="hover:bg-slate-50/50 dark:hover:bg-slate-800/10 transition-colors"
              :class="ent.audit_priority ? 'bg-primary-50/10 dark:bg-primary-950/20' : ''"
            >
              <td class="px-6 py-4 text-slate-500 font-medium">
                {{ idx + 1 }}
              </td>
              <td class="px-6 py-4 font-bold text-slate-800 dark:text-slate-200">
                {{ ent.corporate_audit_universe?.name || 'Unnamed Entity' }}
              </td>
              <td class="px-6 py-4 text-center font-semibold text-slate-700 dark:text-slate-300">
                {{ typeof ent.risk_index === 'number' ? ent.risk_index.toFixed(1) + '%' : '-' }}
              </td>
              <td class="px-6 py-4 text-center">
                <span :class="getRiskLevelBadgeClass(ent.risk_level)">
                  {{ formatRiskLevel(ent.risk_level) }}
                </span>
              </td>
              <td class="px-6 py-4 text-center">
                <div v-if="ent.audit_priority" class="inline-flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400 font-bold">
                  <UIcon name="i-lucide-check-circle" class="w-5 h-5 text-emerald-500" />
                  <span>{{ t('riskFactors.priority.priorityBadge') }}</span>
                </div>
                <span v-else class="text-slate-400 text-sm">-</span>
              </td>
            </tr>

            <!-- Empty Search State -->
            <tr v-if="yearlyUniverse.length > 0 && filteredYearlyUniverse.length === 0">
              <td colspan="5" class="text-center py-10 text-slate-400 text-sm">
                <UIcon name="i-lucide-filter-x" class="w-8 h-8 mx-auto text-slate-300 dark:text-slate-600 mb-2" />
                <p>No auditable entities match your filter criteria.</p>
              </td>
            </tr>

            <!-- Empty State for Year -->
            <tr v-if="yearlyUniverse.length === 0 && !loading">
              <td colspan="5" class="text-center py-12 text-slate-400 text-sm">
                <UIcon name="i-lucide-inbox" class="w-10 h-10 mx-auto text-slate-300 dark:text-slate-600 mb-2" />
                <p class="font-medium text-slate-600 dark:text-slate-300">
                  {{ t('riskFactors.priority.noEntities', { year: selectedYear }) }}
                </p>
                <p class="text-xs text-slate-400 mt-1 mb-4">
                  Please establish active entities in the Yearly Establishment tab of Audit Universe first.
                </p>
                <UButton
                  to="/risk-profile/audit-universe?tab=establish"
                  icon="i-lucide-plus"
                  color="primary"
                  variant="subtle"
                  size="sm"
                >
                  Establish Universe for {{ selectedYear }}
                </UButton>
              </td>
            </tr>

            <!-- Loading State -->
            <tr v-if="loading">
              <td colspan="5" class="text-center py-12 text-slate-400 text-sm">
                <UIcon name="i-lucide-loader-2" class="w-8 h-8 mx-auto text-primary-500 animate-spin mb-2" />
                <p>Loading audit priority data for year {{ selectedYear }}...</p>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <template #footer>
        <div class="flex flex-col sm:flex-row items-center justify-between gap-2 text-xs text-slate-400">
          <span>{{ t('riskFactors.priority.sortedNote') }}</span>
          <span>Showing {{ filteredYearlyUniverse.length }} of {{ yearlyUniverse.length }} entities</span>
        </div>
      </template>
    </UCard>

    <!-- Corporate Risk Index Level Information Reference -->
    <UCard class="shadow-sm border border-[var(--border-main)] bg-slate-50/50 dark:bg-slate-900/30">
      <template #header>
        <h3 class="text-sm font-bold text-slate-800 dark:text-slate-100 flex items-center gap-2">
          <UIcon name="i-lucide-book-open" class="w-4 h-4 text-slate-500" />
          {{ t('riskFactors.priority.levelInfoTitle') }}
        </h3>
      </template>
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-sm text-center">
          <thead class="bg-slate-100 dark:bg-slate-800">
            <tr>
              <th class="px-4 py-2 font-semibold text-slate-700 dark:text-slate-300 text-center">
                {{ t('riskFactors.priority.riskIndex') }}
              </th>
              <th class="px-4 py-2 font-semibold text-slate-700 dark:text-slate-300 text-center">
                {{ t('riskFactors.priority.riskLevel') }}
              </th>
              <th class="px-4 py-2 font-semibold text-slate-700 dark:text-slate-300 text-center">
                {{ t('riskFactors.priority.auditPriorityCol') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
            <tr v-for="row in riskLevelReferenceRows" :key="row.level">
              <td class="px-4 py-2.5 text-slate-600 dark:text-slate-400 font-medium text-center">{{ row.range }}</td>
              <td class="px-4 py-2.5 text-center">
                <span :class="getRiskLevelBadgeClass(row.level)">
                  {{ formatRiskLevel(row.level) }}
                </span>
              </td>
              <td v-if="row.prioritized" class="px-4 py-2.5 text-center text-xs font-bold text-emerald-600 dark:text-emerald-400">
                √ {{ t('riskFactors.scoring.priorityYes') }}
              </td>
              <td v-else class="px-4 py-2.5 text-center text-xs text-slate-400">
                {{ t('riskFactors.scoring.priorityNo') }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { useI18n } from '~/composables/useI18n'
import { useFiscalYear } from '~/composables/useFiscalYear'
import {
  getRiskLevelBadgeClass,
  isPriorityRiskLevel,
  RISK_LEVEL_LABEL_KEYS,
  type PriorityRiskLevel
} from '~/utils/riskLevelBadge'

const store = useAuditUniverseStore()
const { t } = useI18n()

// State
const { fiscalYears, selectedFiscalYear } = useFiscalYear()
const selectedYear = selectedFiscalYear
const searchQuery = ref('')
const statusFilter = ref<'all' | 'prioritized' | 'non_prioritized'>('all')
const alertMessage = ref('')
const alertType = ref('success')

const statusFilterOptions = computed(() => [
  { label: t('auditPriority.allStatus') !== 'auditPriority.allStatus' ? t('auditPriority.allStatus') : 'All Entities', value: 'all' },
  { label: t('auditPriority.onlyPrioritized') !== 'auditPriority.onlyPrioritized' ? t('auditPriority.onlyPrioritized') : 'Prioritized Only', value: 'prioritized' },
  { label: t('auditPriority.onlyNonPrioritized') !== 'auditPriority.onlyNonPrioritized' ? t('auditPriority.onlyNonPrioritized') : 'Non-Prioritized', value: 'non_prioritized' }
])

const loading = computed(() => store.loading)
const yearlyUniverse = computed(() => store.yearlyUniverse || [])

const sortedYearlyUniverse = computed(() => {
  return [...yearlyUniverse.value].sort((a, b) => (b.risk_index || 0) - (a.risk_index || 0))
})

const prioritizedCount = computed(() => {
  return yearlyUniverse.value.filter(ent => ent.audit_priority).length
})

const filteredYearlyUniverse = computed(() => {
  return sortedYearlyUniverse.value.filter(ent => {
    // Search query filter
    if (searchQuery.value) {
      const q = searchQuery.value.toLowerCase().trim()
      const name = (ent.corporate_audit_universe?.name || '').toLowerCase()
      const level = (ent.risk_level || '').toLowerCase()
      if (!name.includes(q) && !level.includes(q)) {
        return false
      }
    }

    // Status filter
    if (statusFilter.value === 'prioritized' && !ent.audit_priority) {
      return false
    }
    if (statusFilter.value === 'non_prioritized' && ent.audit_priority) {
      return false
    }

    return true
  })
})

const loadData = async () => {
  try {
    await store.fetchYearlyUniverse(selectedYear.value)
  } catch (error: any) {
    showAlert(store.errorMsg || 'Failed to load audit priority data', 'error')
  }
}

watch(selectedYear, async (newYear) => {
  if (newYear) {
    await loadData()
  }
})

onMounted(async () => {
  await loadData()
})

const formatRiskLevel = (level?: string) => {
  if (!level) return 'N/A'
  return isPriorityRiskLevel(level) ? t(RISK_LEVEL_LABEL_KEYS[level]) : level
}

// Corporate Risk Index Level Information reference table; badges share getRiskLevelBadgeClass with the main table.
const riskLevelReferenceRows: { range: string, level: PriorityRiskLevel, prioritized: boolean }[] = [
  { range: '80 - 100%', level: 'High', prioritized: true },
  { range: '60 - 79%', level: 'Medium to High', prioritized: true },
  { range: '40 - 59%', level: 'Medium', prioritized: false },
  { range: '20 - 39%', level: 'Low to Medium', prioritized: false },
  { range: '0 - 19%', level: 'Low', prioritized: false }
]

const showAlert = (msg: string, type: string) => {
  alertMessage.value = msg
  alertType.value = type
  setTimeout(() => {
    alertMessage.value = ''
  }, 4000)
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
