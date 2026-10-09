<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { usePerformanceStore } from '~/stores/performance'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import { kpiValueLabel } from '~/utils/kpiPerformanceLabels'
import {
  formatKpiGap,
  formatKpiValue,
  kpiBreakdownRangeText,
  kpiCategoryText,
  kpiFilterMenuValues,
  kpiGapClass,
  kpiStatusColor,
  KPI_BREAKDOWN_PAGE_SIZES,
  KPI_BREAKDOWN_STATUSES,
  type KpiBreakdownItem
} from '~/utils/kpiBreakdown'
import StrategicPlanForm from '~/components/strategic-audit-plan/StrategicPlanForm.vue'

const props = defineProps({
  year: {
    type: Number,
    required: true
  }
})

const { t, locale } = useI18n()
const store = useStrategicPlanStore()
const perfStore = usePerformanceStore()
const { canManageStrategicPlan } = useRbac()

// Filtering, search, paging, gap and status all happen on the backend (GET /performance/kpi-breakdown);
// this component only holds the inputs and renders the page it gets back.
const query = computed(() => perfStore.kpiBreakdownQuery)
const pagination = computed(() => perfStore.kpiBreakdownPagination)
const rows = computed(() => perfStore.kpiBreakdown)

watch(() => props.year, year => perfStore.loadKpiBreakdown(year), { immediate: true })

// Live text in the box; the store debounces it before it becomes part of the query.
const search = ref(query.value.search)
watch(search, value => perfStore.setKpiBreakdownSearch(value))

type FilterKey = 'category' | 'status' | 'period'
const filterModel = (key: FilterKey) => computed<string | undefined>({
  get: () => query.value[key] || undefined,
  set: (value) => { perfStore.setKpiBreakdownFilters({ [key]: value ?? '' }) }
})
const category = filterModel('category')
const period = filterModel('period')
const status = filterModel('status')

const pageSize = computed<number>({
  get: () => query.value.pageSize,
  set: (value) => { perfStore.setKpiBreakdownFilters({ pageSize: Number(value) }) }
})
const page = computed<number>({
  get: () => query.value.page,
  set: (value) => { perfStore.setKpiBreakdownPage(value) }
})

const hasFilters = computed(() => !!(query.value.search || query.value.category || query.value.status || query.value.period))

const resetFilters = () => {
  search.value = ''
  perfStore.resetKpiBreakdownFilters()
}

// Raw values are what the API filters on and what the colour mapping uses; only the label is translated.
// Status, gap and achievement are shown exactly as the API computed them (incl. HIB plans); nothing is recomputed here.
const categoryLabel = (value?: string) => kpiCategoryText(kpiValueLabel(t, 'categories', value))
const statusLabel = (value?: string) => kpiValueLabel(t, 'statuses', value)
// Achievement rows use 'Tahunan' for the annual period (see the page-level period selector).
const periodLabel = (value: string) => value === 'Tahunan' ? t('kpiPerformance.upload.annual') : value

// Menu options come from the API's `data.filters` for the year (already sorted there). Without it
// (older backend) or with no values, the menu is hidden; an active value stays listed so it can be cleared.
const categoryItems = computed(() =>
  kpiFilterMenuValues(perfStore.kpiBreakdownFilters.categories, query.value.category).map(value => ({ label: categoryLabel(value), value }))
)
const periodItems = computed(() =>
  kpiFilterMenuValues(perfStore.kpiBreakdownFilters.periods, query.value.period).map(value => ({ label: periodLabel(value), value }))
)
const statusItems = computed(() => KPI_BREAKDOWN_STATUSES.map((value): { label: string, value: string } => ({ label: statusLabel(value), value })))
const pageSizeItems = KPI_BREAKDOWN_PAGE_SIZES.map(value => ({ label: String(value), value }))

const numberLocale = computed(() => locale.value === 'id' ? 'id-ID' : 'en-US')
const formatValue = (value: number, unit: string) => formatKpiValue(value, unit, numberLocale.value)
const formatGap = (item: KpiBreakdownItem) => formatKpiGap(item.gap, item.unit, numberLocale.value)

const rangeText = computed(() => kpiBreakdownRangeText(t, pagination.value))

const columns = computed(() => [
  // Same width as the KPI column in StrategicPlanTable; long names wrap onto more lines.
  {
    accessorKey: 'metric',
    header: t('kpiPerformance.table.columns.metric'),
    meta: {
      class: {
        th: 'w-[130px] min-w-[130px] max-w-[130px] sm:w-[220px] sm:min-w-[220px] sm:max-w-[220px] whitespace-normal',
        td: 'w-[130px] min-w-[130px] max-w-[130px] sm:w-[220px] sm:min-w-[220px] sm:max-w-[220px] whitespace-normal break-words'
      }
    }
  },
  { accessorKey: 'category', header: t('kpiPerformance.table.columns.category') },
  { accessorKey: 'target', header: t('kpiPerformance.table.columns.target') },
  { accessorKey: 'actual', header: t('kpiPerformance.table.columns.actual') },
  { accessorKey: 'gap', header: t('kpiPerformance.table.columns.gap') },
  { accessorKey: 'status', header: t('kpiPerformance.table.columns.status') },
  ...(canManageStrategicPlan.value ? [{ accessorKey: 'actions', header: t('kpiPerformance.table.columns.actions') }] : [])
])

// Strategic-plan rows are deleted on the server (the store confirms, calls DELETE and toasts), then
// the current page is reloaded so totals and paging stay right. Achievement rows come from uploaded
// performance reports and cannot be deleted here.
const deletingPlanId = ref<string | null>(null)
async function deleteKpiTarget(item: KpiBreakdownItem) {
  if (item.source !== 'strategic_plan' || deletingPlanId.value) return
  deletingPlanId.value = item.id
  try {
    await store.handleDelete(item.id)
    await perfStore.fetchKpiBreakdown()
  } finally {
    deletingPlanId.value = null
  }
}

// Strategic-plan rows open the plan form with a fresh copy of that plan: the form PUTs every field
// it edits, so editing a stale list entry could overwrite newer data. Achievement rows come from
// uploaded performance reports and are not editable here.
const openingPlanId = ref<string | null>(null)
async function editKpiTarget(item: KpiBreakdownItem) {
  if (item.source !== 'strategic_plan' || openingPlanId.value) return
  openingPlanId.value = item.id
  try {
    const plan = await store.fetchStrategicPlanById(item.id)
    if (plan) store.handleEdit(plan)
  } finally {
    openingPlanId.value = null
  }
}

// The form closes itself after a save; reload the current page so edited/new targets show up.
watch(() => store.isAddModalOpen, (open, wasOpen) => {
  if (wasOpen && !open) perfStore.fetchKpiBreakdown()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h2 class="text-xl font-bold text-gray-900 dark:text-white">
        {{ t('kpiPerformance.table.title') }}
      </h2>
      <UButton
        v-if="canManageStrategicPlan"
        :label="t('kpiPerformance.table.setTargets')"
        icon="i-lucide-plus"
        color="primary"
        size="sm"
        class="font-bold rounded-xl"
        @click="store.openModal()"
      />
    </div>

    <div class="flex flex-col sm:flex-row flex-wrap items-stretch sm:items-center gap-3 sm:gap-4 w-full">
      <UInput
        v-model="search"
        icon="i-lucide-search"
        :placeholder="t('kpiPerformance.table.searchPlaceholder')"
        class="w-full sm:w-64"
      />
      <USelectMenu
        v-if="categoryItems.length"
        v-model="category"
        :items="categoryItems"
        value-key="value"
        :placeholder="t('kpiPerformance.table.selectCategory')"
        class="w-full sm:w-48"
      />
      <USelectMenu
        v-if="periodItems.length"
        v-model="period"
        :items="periodItems"
        value-key="value"
        :placeholder="t('kpiPerformance.table.selectPeriod')"
        class="w-full sm:w-48"
      />
      <USelectMenu
        v-model="status"
        :items="statusItems"
        value-key="value"
        :placeholder="t('kpiPerformance.table.selectStatus')"
        class="w-full sm:w-48"
      />
      <UButton
        :label="t('kpiPerformance.table.resetFilter')"
        icon="i-lucide-rotate-ccw"
        color="neutral"
        variant="outline"
        class="w-full sm:w-auto justify-center"
        :disabled="!hasFilters && !search"
        @click="resetFilters"
      />
    </div>

    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-x-auto shadow-sm">
      <UTable
        :columns="columns"
        :data="rows"
        :loading="perfStore.kpiBreakdownLoading"
        :ui="{ th: 'bg-gray-100 dark:bg-gray-800/50' }"
      >
        <template #metric-cell="{ row }">
          <span class="block font-bold text-gray-900 dark:text-white whitespace-normal break-words">{{ row.original.metric }}</span>
        </template>

        <template #category-cell="{ row }">
          <span class="font-semibold text-gray-900 dark:text-white">{{ categoryLabel(row.original.category) }}</span>
        </template>

        <template #target-cell="{ row }">
          <span class="font-bold text-gray-900 dark:text-white whitespace-nowrap">{{ formatValue(row.original.target, row.original.unit) }}</span>
        </template>

        <template #actual-cell="{ row }">
          <span class="font-bold text-gray-900 dark:text-white whitespace-nowrap">{{ formatValue(row.original.actual, row.original.unit) }}</span>
        </template>

        <template #gap-cell="{ row }">
          <span :class="['font-bold whitespace-nowrap', kpiGapClass(row.original)]">
            {{ formatGap(row.original) }}
          </span>
        </template>

        <template #status-cell="{ row }">
          <div class="flex items-center gap-2">
            <span :class="['w-3 h-3 rounded-full shrink-0', kpiStatusColor(row.original.status)]" />
            <span class="font-semibold text-gray-900 dark:text-white">{{ statusLabel(row.original.status) }}</span>
          </div>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-1">
            <UTooltip
              :text="row.original.source === 'strategic_plan' ? t('kpiPerformance.table.editTarget') : t('kpiPerformance.table.editUnavailable')"
            >
              <UButton
                color="warning"
                variant="ghost"
                size="md"
                icon="i-lucide-edit"
                :disabled="row.original.source !== 'strategic_plan'"
                :loading="openingPlanId === row.original.id"
                :aria-label="t('kpiPerformance.table.editTarget')"
                @click="editKpiTarget(row.original)"
              />
            </UTooltip>
            <UTooltip
              :text="row.original.source === 'strategic_plan' ? t('kpiPerformance.table.deleteTarget') : t('kpiPerformance.table.deleteUnavailable')"
            >
              <UButton
                color="error"
                variant="ghost"
                size="md"
                icon="i-lucide-trash-2"
                :disabled="row.original.source !== 'strategic_plan'"
                :loading="deletingPlanId === row.original.id"
                :aria-label="t('kpiPerformance.table.deleteTarget')"
                @click="deleteKpiTarget(row.original)"
              />
            </UTooltip>
          </div>
        </template>

        <template #loading>
          <div class="flex items-center justify-center gap-2 py-8 text-sm text-gray-500">
            <UIcon
              name="i-lucide-loader-2"
              class="w-5 h-5 animate-spin"
            />
            {{ t('kpiPerformance.table.loading') }}
          </div>
        </template>

        <template #empty>
          <div
            v-if="perfStore.kpiBreakdownError"
            class="flex flex-col items-center justify-center gap-2 py-8 text-center"
          >
            <UIcon
              name="i-lucide-alert-triangle"
              class="w-8 h-8 text-red-500"
            />
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('kpiPerformance.table.errorTitle') }}
            </p>
            <p class="text-sm text-gray-500">
              {{ perfStore.kpiBreakdownError }}
            </p>
            <UButton
              :label="t('kpiPerformance.table.retry')"
              icon="i-lucide-refresh-cw"
              color="neutral"
              variant="outline"
              size="sm"
              @click="perfStore.fetchKpiBreakdown()"
            />
          </div>
          <div
            v-else
            class="flex flex-col items-center justify-center gap-2 py-8 text-center"
          >
            <UIcon
              name="i-lucide-database"
              class="w-8 h-8 text-gray-400"
            />
            <p class="text-sm font-medium text-gray-500">
              {{ hasFilters ? t('kpiPerformance.table.emptyFiltered') : t('kpiPerformance.table.empty', { year: props.year }) }}
            </p>
          </div>
        </template>
      </UTable>

      <div class="px-4 py-3 border-t border-gray-200 dark:border-gray-800 flex flex-col sm:flex-row items-center justify-between gap-3">
        <span class="text-md text-gray-500 font-semibold">{{ rangeText }}</span>
        <div
          v-if="pagination.total > 0"
          class="flex flex-wrap items-center justify-center gap-3"
        >
          <div class="flex items-center gap-2">
            <span class="text-sm text-gray-500">{{ t('kpiPerformance.table.rowsPerPage') }}</span>
            <USelect
              v-model="pageSize"
              :items="pageSizeItems"
              size="sm"
              class="w-20"
            />
          </div>
          <UPagination
            v-if="pagination.total_pages > 1"
            v-model:page="page"
            :items-per-page="pagination.page_size"
            :total="pagination.total"
            size="sm"
            active-color="primary"
            color="neutral"
            variant="outline"
          />
        </div>
      </div>
    </div>

    <!-- Strategic Plan / KPI Target Form Modal -->
    <StrategicPlanForm />
  </div>
</template>
