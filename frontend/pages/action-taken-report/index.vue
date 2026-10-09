<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { ActionTakenReport } from '~/types/audit'
import { useActionTakenReportStore, type AtrPicCandidate } from '~/stores/action-taken-report'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import { useAtrActions } from '~/composables/useAtrActions'
import {
  ATR_SLICE_DOT_CLASS,
  ATR_STATUSES,
  atrFindingCategoryColor,
  atrFindingCategoryI18nKey,
  atrStatusI18nKey,
  atrVisibleRows,
  formatAtrDate,
  type AtrScope,
  type AtrSliceKey
} from '~/utils/actionTakenReport'
import TableEntities from '~/components/shared/TableEntities.vue'
import ReusableSelectMenu from '~/components/shared/ReusableSelectMenu.vue'
import ATRDetail from '~/components/action-taken-report/ATRDetail.vue'
import ATRActionPlanForm from '~/components/action-taken-report/ATRActionPlanForm.vue'
import ATRAssignmentModal from '~/components/action-taken-report/ATRAssignmentModal.vue'
import ATRReviewModal from '~/components/action-taken-report/ATRReviewModal.vue'
import ATRCancelModal from '~/components/action-taken-report/ATRCancelModal.vue'
import AtrStatusBadge from '~/components/action-taken-report/AtrStatusBadge.vue'

definePageMeta({ middleware: 'auth' })

const store = useActionTakenReportStore()
const { t, locale } = useI18n()
const { canViewAllAtr, isAtrPic, getCurrentUserId } = useRbac()
const { availableActions } = useAtrActions()
const route = useRoute()

// Filtering, search and paging happen on the backend (GET /action-taken-reports); the page holds the inputs.
const query = computed(() => store.query)
const pagination = computed(() => store.pagination)

// "My actions": the server returns only the user's items (pic_user_id = me). As a guard, rows assigned to
// someone else are never shown in this view even if the server sent them.
const rows = computed<ActionTakenReport[]>(() => atrVisibleRows(store.items, query.value.scope, getCurrentUserId()))

const lhaFromRoute = () => (typeof route.query.lha === 'string' ? route.query.lha : '')

onMounted(() => {
  store.initList({ scope: canViewAllAtr.value ? 'all' : 'mine', lhaId: lhaFromRoute() })
  store.fetchSummary()
  store.fetchLhaOptions()
})

// Coming from the LHA page while already here (?lha= changes).
watch(() => route.query.lha, () => {
  const lhaId = lhaFromRoute()
  if (lhaId !== query.value.lhaId) store.setFilters({ lhaId })
})

// --- Scope ---------------------------------------------------------------------------------------
const scopeItems = computed(() => [
  { label: t('actionTakenReport.scope.all'), value: 'all', icon: 'i-lucide-list' },
  { label: t('actionTakenReport.scope.mine'), value: 'mine', icon: 'i-lucide-user-check' }
])
const scope = computed<string>({
  get: () => query.value.scope,
  set: (value) => { store.setFilters({ scope: value as AtrScope, picUserId: '' }) }
})

// --- Filters -------------------------------------------------------------------------------------
// Live text in the box; the store debounces it before it becomes part of the query.
const search = ref(query.value.search)
watch(search, value => store.setSearch(value))

const statusLabel = (status: string) => {
  const key = atrStatusI18nKey(status)
  return key ? t(key) : status
}
const statusItems = computed(() => ATR_STATUSES.map(value => ({ label: statusLabel(value), value })))
const status = computed<string | undefined>({
  get: () => query.value.status || undefined,
  set: (value) => { store.setFilters({ status: value ?? '' }) }
})

const lhaItems = computed(() => store.lhaOptions.map(lha => ({
  label: lha.reportTitle ? `${lha.reportNumber} — ${lha.reportTitle}` : lha.reportNumber,
  value: lha.id
})))
const lhaId = computed<string | undefined>({
  get: () => query.value.lhaId || undefined,
  set: (value) => { store.setFilters({ lhaId: value ?? '' }) }
})

const overdue = computed<boolean>({
  get: () => query.value.overdue,
  set: (value) => { store.setFilters({ overdue: value }) }
})

// PIC filter (users who can view all, in the "all" view): server-side search on the assignable users.
const picCandidates = ref<AtrPicCandidate[]>([])
const picSearch = ref('')
const picLoading = ref(false)
const picUnavailable = ref(false)
let picTimer: ReturnType<typeof setTimeout> | null = null
let picRequest = 0
const loadPics = async (term: string) => {
  const id = ++picRequest
  picLoading.value = true
  const result = await store.fetchPicCandidates(term)
  if (id !== picRequest) return
  picCandidates.value = result.users
  picUnavailable.value = result.forbidden
  picLoading.value = false
}
watch(picSearch, (term) => {
  if (picTimer) clearTimeout(picTimer)
  picTimer = setTimeout(() => {
    picTimer = null
    loadPics(term)
  }, 300)
})
onMounted(() => {
  if (canViewAllAtr.value) loadPics('')
})
onBeforeUnmount(() => {
  if (picTimer) clearTimeout(picTimer)
})
const picItems = computed(() => picCandidates.value.map(u => ({
  label: [u.full_name, [u.position, u.department].filter(Boolean).join(', ')].filter(Boolean).join(' — '),
  value: u.id
})))
const picUserId = computed<string | undefined>({
  get: () => query.value.picUserId || undefined,
  set: (value) => { store.setFilters({ picUserId: value ?? '' }) }
})
const showPicFilter = computed(() => canViewAllAtr.value && query.value.scope === 'all' && !picUnavailable.value)

const hasFilters = computed(() => !!(query.value.search || query.value.status || query.value.lhaId || query.value.picUserId || query.value.overdue))

const resetFilters = () => {
  search.value = ''
  store.resetFilters()
}

const rangeText = computed(() => {
  const p = pagination.value
  if (p.total <= 0) return t('actionTakenReport.table.showingNone')
  const from = (p.page - 1) * p.page_size + 1
  if (from > p.total) return t('actionTakenReport.table.showingNone')
  return t('actionTakenReport.table.showing', { from, to: Math.min(p.page * p.page_size, p.total), total: p.total })
})

// --- Summary -------------------------------------------------------------------------------------
const sliceLabelKeys: Record<AtrSliceKey, string> = {
  completed: 'actionTakenReport.summary.done',
  pendingReview: 'actionTakenReport.summary.pendingReview',
  inProgress: 'actionTakenReport.summary.inProgress',
  planned: 'actionTakenReport.summary.planned',
  overdue: 'actionTakenReport.summary.overdue',
  cancelled: 'actionTakenReport.summary.cancelled'
}

// --- Table ---------------------------------------------------------------------------------------
const columns = computed(() => [
  {
    accessorKey: 'finding_title',
    header: t('actionTakenReport.table.columns.finding'),
    meta: { class: { th: 'min-w-[260px]', td: 'min-w-[260px] max-w-[360px] whitespace-normal' } }
  },
  {
    accessorKey: 'report_number',
    header: t('actionTakenReport.table.columns.lha'),
    meta: { class: { th: 'min-w-[200px]', td: 'min-w-[200px] max-w-[260px] whitespace-normal' } }
  },
  { accessorKey: 'pic_name', header: t('actionTakenReport.table.columns.pic') },
  { accessorKey: 'due_date', header: t('actionTakenReport.table.columns.dueDate') },
  { accessorKey: 'progress', header: t('actionTakenReport.table.columns.progress') },
  { accessorKey: 'status', header: t('actionTakenReport.table.columns.status') },
  { accessorKey: 'actions', header: t('actionTakenReport.table.columns.actions') }
])

const categoryLabel = (value: string) => {
  const key = atrFindingCategoryI18nKey(value)
  return key ? t(key) : value
}

const filterByLha = (id: string) => {
  store.closeModal()
  store.setFilters({ lhaId: id })
}
</script>

<template>
  <div class="p-4 sm:p-6 space-y-6 min-w-0">
    <div class="flex flex-col gap-1">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
        {{ t('actionTakenReport.page.title') }}
      </h1>
      <p class="text-gray-500 dark:text-gray-400">
        {{ t('actionTakenReport.page.subtitle') }}
      </p>
    </div>

    <!-- Status summary over every ATR the user can see -->
    <UCard :ui="{ body: 'p-4 space-y-2' }">
      <div class="flex items-center gap-2">
        <p class="text-sm font-bold">
          {{ t('actionTakenReport.statusSummary') }}
        </p>
        <UIcon
          v-if="store.summaryLoading"
          name="i-lucide-loader-2"
          class="w-4 h-4 animate-spin text-primary-500"
        />
      </div>
      <p
        v-if="store.summaryError"
        class="text-sm text-error-600 dark:text-error-400"
      >
        {{ store.summaryError }}
      </p>
      <div
        v-else
        class="flex flex-wrap items-center gap-3 sm:gap-6"
      >
        <div
          v-for="slice in store.stats.breakdown"
          :key="slice.key"
          class="flex items-center gap-2"
        >
          <span :class="['w-3.5 h-3.5 rounded-full', ATR_SLICE_DOT_CLASS[slice.key]]" />
          <span class="text-sm font-semibold">{{ t(sliceLabelKeys[slice.key], { percent: slice.percent }) }}</span>
        </div>
      </div>
      <p class="text-xs text-gray-500 dark:text-gray-400">
        {{ t('actionTakenReport.summaryHint') }}
      </p>
    </UCard>

    <!-- Scope: users who can see every ATR switch between all and their own -->
    <UTabs
      v-if="canViewAllAtr"
      v-model="scope"
      :items="scopeItems"
      :content="false"
      size="sm"
      class="w-full sm:w-auto"
    />
    <div
      v-else
      class="flex items-center gap-2"
    >
      <UBadge
        color="primary"
        variant="subtle"
        icon="i-lucide-user-check"
        :label="t('actionTakenReport.scope.mine')"
      />
      <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('actionTakenReport.scope.mineHint') }}</span>
    </div>

    <!-- Filters -->
    <div class="flex flex-col sm:flex-row flex-wrap items-stretch sm:items-center gap-3">
      <UInput
        v-model="search"
        icon="i-lucide-search"
        :placeholder="t('actionTakenReport.filter.searchPlaceholder')"
        class="w-full sm:w-64"
      />
      <div class="w-full sm:w-48">
        <ReusableSelectMenu
          v-model="status"
          :items="statusItems"
          value-key="value"
          :placeholder="t('actionTakenReport.filter.statusPlaceholder')"
        />
      </div>
      <div class="w-full sm:w-72">
        <ReusableSelectMenu
          v-model="lhaId"
          :items="lhaItems"
          value-key="value"
          :loading="store.lhaOptionsLoading"
          :placeholder="t('actionTakenReport.filter.lhaPlaceholder')"
        />
      </div>
      <div
        v-if="showPicFilter"
        class="w-full sm:w-60"
      >
        <ReusableSelectMenu
          v-model="picUserId"
          v-model:search-term="picSearch"
          :items="picItems"
          value-key="value"
          ignore-filter
          :loading="picLoading"
          :placeholder="t('actionTakenReport.filter.picPlaceholder')"
          :search-input="{ placeholder: t('actionTakenReport.assignment.picSearch') }"
        />
      </div>
      <USwitch
        v-model="overdue"
        :label="t('actionTakenReport.filter.overdueOnly')"
        color="error"
      />
      <UButton
        :label="t('actionTakenReport.filter.reset')"
        icon="i-lucide-rotate-ccw"
        color="neutral"
        variant="outline"
        class="w-full sm:w-auto justify-center"
        :disabled="!hasFilters && !search"
        @click="resetFilters"
      />
    </div>

    <div class="space-y-2">
      <p class="text-sm text-gray-500 font-semibold">
        {{ rangeText }}
      </p>
      <TableEntities
        :data="rows"
        :columns="columns"
        :loading="store.loading"
        :server-side="true"
        :total="pagination.total"
        :items-per-page="pagination.page_size"
        :page="pagination.page"
        @update:page="(p: number) => store.setPage(p)"
        @update:items-per-page="(size: number) => store.setFilters({ pageSize: size })"
      >
        <template #finding_title-cell="{ row }">
          <div class="space-y-1.5">
            <p class="font-semibold text-gray-900 dark:text-white whitespace-normal break-words">
              {{ row.original.finding_title || '-' }}
            </p>
            <UBadge
              v-if="row.original.finding_category"
              :color="atrFindingCategoryColor(row.original.finding_category)"
              variant="subtle"
              size="sm"
              :label="categoryLabel(row.original.finding_category)"
            />
          </div>
        </template>

        <template #report_number-cell="{ row }">
          <div class="space-y-0.5">
            <p class="font-mono text-sm font-semibold text-primary-600 dark:text-primary-400">
              {{ row.original.report_number || '-' }}
            </p>
            <p
              v-if="row.original.report_title"
              class="text-xs text-gray-500 dark:text-gray-400 whitespace-normal break-words"
            >
              {{ row.original.report_title }}
            </p>
          </div>
        </template>

        <template #pic_name-cell="{ row }">
          <div class="flex items-center gap-1.5">
            <span
              v-if="row.original.pic_user_id"
              class="font-medium"
            >{{ row.original.pic_name || '-' }}</span>
            <span
              v-else
              class="italic text-gray-400"
            >{{ t('actionTakenReport.table.unassigned') }}</span>
            <UBadge
              v-if="isAtrPic(row.original)"
              color="primary"
              variant="soft"
              size="sm"
              :label="t('actionTakenReport.table.you')"
            />
          </div>
        </template>

        <template #due_date-cell="{ row }">
          <div class="space-y-0.5 whitespace-nowrap">
            <p :class="row.original.is_overdue ? 'font-semibold text-error-600 dark:text-error-400' : ''">
              {{ formatAtrDate(row.original.due_date, locale) }}
            </p>
            <p
              v-if="row.original.is_overdue"
              class="flex items-center gap-1 text-xs font-medium text-error-600 dark:text-error-400"
            >
              <UIcon
                name="i-lucide-alarm-clock"
                class="w-3.5 h-3.5"
              />
              {{ t('actionTakenReport.daysOverdue', { days: row.original.overdue_days }) }}
            </p>
          </div>
        </template>

        <template #progress-cell="{ row }">
          <div class="flex items-center gap-2 min-w-[120px]">
            <UProgress
              :model-value="row.original.progress"
              :max="100"
              size="sm"
              class="flex-1"
            />
            <span class="text-xs font-semibold tabular-nums w-9 text-right">{{ row.original.progress }}%</span>
          </div>
        </template>

        <template #status-cell="{ row }">
          <AtrStatusBadge :status="row.original.status" />
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-1">
            <UTooltip :text="t('actionTakenReport.actions.view')">
              <UButton
                icon="i-lucide-eye"
                color="neutral"
                variant="ghost"
                :aria-label="t('actionTakenReport.actions.view')"
                @click="store.openDetail(row.original)"
              />
            </UTooltip>
            <UTooltip
              v-if="availableActions(row.original).actionPlan"
              :text="t('actionTakenReport.actions.fillActionPlan')"
            >
              <UButton
                icon="i-lucide-pencil-line"
                color="primary"
                variant="ghost"
                :aria-label="t('actionTakenReport.actions.fillActionPlan')"
                @click="store.openActionPlan(row.original)"
              />
            </UTooltip>
            <UTooltip
              v-if="availableActions(row.original).assign"
              :text="row.original.pic_user_id ? t('actionTakenReport.actions.reassign') : t('actionTakenReport.actions.assign')"
            >
              <UButton
                icon="i-lucide-user-plus"
                color="neutral"
                variant="ghost"
                :aria-label="t('actionTakenReport.actions.assign')"
                @click="store.openAssignment(row.original)"
              />
            </UTooltip>
            <UTooltip
              v-if="availableActions(row.original).review"
              :text="t('actionTakenReport.actions.review')"
            >
              <UButton
                icon="i-lucide-clipboard-check"
                color="success"
                variant="ghost"
                :aria-label="t('actionTakenReport.actions.review')"
                @click="store.openReview(row.original)"
              />
            </UTooltip>
            <UTooltip
              v-if="availableActions(row.original).cancel"
              :text="t('actionTakenReport.actions.cancel')"
            >
              <UButton
                icon="i-lucide-ban"
                color="error"
                variant="ghost"
                :aria-label="t('actionTakenReport.actions.cancel')"
                @click="store.openCancel(row.original)"
              />
            </UTooltip>
          </div>
        </template>

        <template #empty>
          <div
            v-if="store.errorMsg"
            class="flex flex-col items-center justify-center gap-2 py-8 text-center"
          >
            <UIcon
              name="i-lucide-alert-triangle"
              class="w-8 h-8 text-error-500"
            />
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('actionTakenReport.table.errorTitle') }}
            </p>
            <p class="text-sm text-gray-500">
              {{ store.errorMsg }}
            </p>
            <UButton
              :label="t('actionTakenReport.table.retry')"
              icon="i-lucide-refresh-cw"
              color="neutral"
              variant="outline"
              size="sm"
              @click="store.fetchReports()"
            />
          </div>
          <div
            v-else
            class="flex flex-col items-center justify-center gap-2 py-8 px-4 text-center"
          >
            <UIcon
              :name="hasFilters ? 'i-lucide-search-x' : 'i-lucide-clipboard-list'"
              class="w-10 h-10 text-gray-400 opacity-70"
            />
            <p class="text-sm font-medium text-gray-600 dark:text-gray-300">
              {{ hasFilters ? t('actionTakenReport.table.emptyFiltered') : (query.scope === 'mine' ? t('actionTakenReport.table.emptyMine') : t('actionTakenReport.table.empty')) }}
            </p>
            <p
              v-if="!hasFilters"
              class="text-xs text-gray-500 dark:text-gray-400 max-w-md"
            >
              {{ t('actionTakenReport.table.emptyHint') }}
            </p>
          </div>
        </template>
      </TableEntities>
    </div>

    <ATRDetail @filter-lha="filterByLha" />
    <ATRActionPlanForm />
    <ATRAssignmentModal />
    <ATRReviewModal />
    <ATRCancelModal />
  </div>
</template>
