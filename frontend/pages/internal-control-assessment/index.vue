<template>
  <div class="space-y-6 min-w-0">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
          <UIcon name="i-lucide-shield-check" class="size-6 sm:size-7 text-primary-500 shrink-0" />
          {{ t('icAssessment.page.title') }}
        </h1>
        <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-1">
          {{ t('icAssessment.page.description') }}
        </p>
      </div>
      <div v-if="caps.canLockPeriod" class="flex gap-2">
        <UButton
          v-if="!locked"
          color="neutral"
          variant="outline"
          icon="i-lucide-lock"
          :label="t('icAssessment.period.lock')"
          @click="lockOpen = true"
        />
        <UButton
          v-else-if="lockedRow"
          color="neutral"
          variant="outline"
          icon="i-lucide-lock-open"
          :label="t('icAssessment.period.unlock')"
          @click="unlockOpen = true"
        />
      </div>
    </div>

    <!-- Filters -->
    <UCard :ui="{ root: 'bg-transparent!' }">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
        <UFormField :label="t('icAssessment.filters.year')">
          <USelect v-model="year" :items="yearItems" class="w-full" />
        </UFormField>
        <UFormField :label="t('icAssessment.filters.quarter')">
          <USelect v-model="quarter" :items="quarterItems" class="w-full" />
        </UFormField>
        <UFormField :label="t('icAssessment.filters.department')">
          <USelectMenu
            v-model="departmentFilter"
            :items="departmentItems"
            value-key="value"
            :search-input="{ placeholder: t('common.search') }"
            class="w-full"
          />
        </UFormField>
        <UFormField :label="t('icAssessment.filters.branch')">
          <USelect v-model="branchFilter" :items="branchItems" class="w-full" />
        </UFormField>
        <UFormField :label="t('icAssessment.filters.status')">
          <USelect v-model="statusFilter" :items="statusItems" class="w-full" />
        </UFormField>
      </div>
    </UCard>

    <!-- Period lock banner -->
    <div
      v-if="locked"
      class="p-4 rounded-xl bg-amber-50/90 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-900/60 flex items-start gap-3"
      role="status"
    >
      <UIcon name="i-lucide-lock" class="size-5 shrink-0 text-amber-700 dark:text-amber-400 mt-0.5" />
      <div class="text-sm text-amber-900 dark:text-amber-200">
        <p class="font-semibold">{{ t('icAssessment.period.lockedBanner') }}</p>
        <p v-if="lockedRow?.lockedBy">{{ t('icAssessment.period.lockedBy', { name: lockedRow.lockedBy.name || '-', date: formatDate(lockedRow.lockedAt) }) }}</p>
        <p v-if="lockedRow?.lockNote" class="whitespace-pre-line">{{ lockedRow.lockNote }}</p>
      </div>
    </div>

    <p v-if="!caps.fullRead" class="text-sm text-gray-600 dark:text-gray-300 flex items-center gap-1.5">
      <UIcon name="i-lucide-info" class="size-4 shrink-0" />
      {{ t('icAssessment.grid.scopedNote') }}
    </p>
    <p v-if="caps.fullRead && departmentsFailed" class="text-sm text-amber-800 dark:text-amber-300 flex items-center gap-1.5">
      <UIcon name="i-lucide-triangle-alert" class="size-4 shrink-0" />
      {{ t('icAssessment.grid.departmentsUnavailable') }}
    </p>

    <!-- Department x component grid -->
    <UCard :ui="{ root: 'bg-transparent!', body: 'p-0 sm:p-0' }" class="overflow-hidden">
      <div v-if="store.listLoading" class="p-4 space-y-3">
        <USkeleton v-for="i in 5" :key="i" class="h-10 w-full" />
      </div>
      <div v-else-if="listFailed" class="p-6 text-sm text-red-700 dark:text-red-400 flex items-center gap-2">
        <UIcon name="i-lucide-circle-alert" class="size-4" />
        {{ t('icAssessment.grid.loadError') }}
        <UButton size="xs" variant="link" :label="t('icAssessment.actions.reload')" @click="load" />
      </div>
      <div v-else-if="!rows.length" class="p-10 text-center text-sm text-gray-500 dark:text-gray-400">
        {{ caps.fullRead ? t('icAssessment.grid.empty') : t('icAssessment.grid.emptyScoped') }}
      </div>
      <div v-else class="overflow-x-auto">
        <table class="min-w-full text-sm">
          <thead class="bg-gray-50 dark:bg-gray-800/60 text-left text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
            <tr>
              <th class="px-4 py-3 min-w-[200px]">{{ t('icAssessment.grid.department') }}</th>
              <th v-for="code in COSO_COMPONENT_CODES" :key="code" class="px-3 py-3 min-w-[130px]">
                <span class="font-mono">{{ code }}</span>
                <span class="block normal-case font-medium text-gray-600 dark:text-gray-300">{{ componentLabel(t, code) }}</span>
              </th>
              <th class="px-3 py-3 min-w-[150px]">{{ t('icAssessment.grid.overall') }}</th>
              <th class="px-3 py-3 text-right">{{ t('icAssessment.grid.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr v-for="row in rows" :key="row.key" class="hover:bg-gray-50/60 dark:hover:bg-gray-800/30">
              <td class="px-4 py-3">
                <p class="font-semibold text-gray-900 dark:text-white">{{ row.departmentName }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  <span v-if="row.departmentCode" class="font-mono">{{ row.departmentCode }}</span>
                  <span v-if="row.item?.locationName"> &middot; {{ row.item.locationName }}</span>
                </p>
                <UBadge
                  v-if="row.item"
                  :color="ASSESSMENT_STATUS_COLOR[row.item.status] || 'neutral'"
                  variant="subtle"
                  size="sm"
                  class="mt-1"
                >
                  {{ t(`icAssessment.status.${row.item.status}`) }}
                </UBadge>
                <UBadge v-else color="neutral" variant="outline" size="sm" class="mt-1">
                  {{ t('icAssessment.grid.notStarted') }}
                </UBadge>
              </td>
              <td v-for="code in COSO_COMPONENT_CODES" :key="code" class="px-3 py-3">
                <template v-if="row.item">
                  <IcStatusText :status="cell(row.item, code)?.finalStatus ?? null" short />
                  <p class="text-xs text-gray-600 dark:text-gray-300 tabular-nums mt-0.5">
                    {{ cell(row.item, code)?.finalRating ? t('icAssessment.grid.rating', { rating: cell(row.item, code)!.finalRating! }) : t('icAssessment.grid.noRating') }}
                  </p>
                </template>
                <span v-else class="text-gray-400 dark:text-gray-500">-</span>
              </td>
              <td class="px-3 py-3">
                <UBadge
                  v-if="row.item && (row.item.finalConclusion || row.item.derivedConclusion)"
                  :color="CONCLUSION_COLOR[row.item.finalConclusion || row.item.derivedConclusion!] || 'neutral'"
                  variant="subtle"
                >
                  {{ t(`icAssessment.conclusion.${row.item.finalConclusion || row.item.derivedConclusion}`) }}
                </UBadge>
                <span v-else class="text-gray-400 dark:text-gray-500">-</span>
              </td>
              <td class="px-3 py-3 text-right whitespace-nowrap">
                <UButton
                  v-if="row.item"
                  size="xs"
                  color="primary"
                  variant="soft"
                  icon="i-lucide-arrow-right"
                  :label="t('icAssessment.grid.open')"
                  :to="`/internal-control-assessment/${row.item.id}`"
                />
                <UButton
                  v-else-if="caps.canAssess && !locked && row.departmentId"
                  size="xs"
                  color="primary"
                  icon="i-lucide-plus"
                  :label="t('icAssessment.grid.create')"
                  :loading="creatingFor === row.departmentId"
                  :disabled="!!creatingFor"
                  @click="createFor(row)"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </UCard>

    <IcReasonModal
      v-model:open="lockOpen"
      :title="t('icAssessment.period.lockTitle', { period: currentPeriodLabel })"
      :description="t('icAssessment.period.lockDescription')"
      :label="t('icAssessment.period.lockNote')"
      :confirm-label="t('icAssessment.period.lock')"
      :loading="store.saving"
      @confirm="doLock"
    />
    <IcReasonModal
      v-model:open="unlockOpen"
      :title="t('icAssessment.period.unlockTitle', { period: lockedRow ? periodLabel(t, lockedRow.year, lockedRow.quarter) : currentPeriodLabel })"
      :label="t('icAssessment.period.unlockReason')"
      required
      :confirm-label="t('icAssessment.period.unlock')"
      :loading="store.saving"
      @confirm="doUnlock"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import { useAppToast } from '~/composables/useAppToast'
import { useDepartmentApi } from '~/composables/useDepartmentApi'
import { useLocationApi } from '~/composables/useLocationApi'
import { useIcAssessmentStore } from '~/stores/ic-assessment'
import { useGlobalModalStore } from '~/stores/global-modal'
import type { Department, Location } from '~/types/master'
import IcStatusText from '~/components/ic-assessment/IcStatusText.vue'
import IcReasonModal from '~/components/ic-assessment/IcReasonModal.vue'
import { ASSESSMENT_STATUSES, COSO_COMPONENT_CODES, isLockedFor, type ComponentCode, type IcAssessmentListItem } from '~/utils/icAssessment'
import { ASSESSMENT_STATUS_COLOR, CONCLUSION_COLOR, componentLabel, periodLabel } from '~/utils/icAssessmentLabels'
import { currentFiscalYear } from '~/utils/fiscalPeriod'
import { getUserErrorMessage } from '~/utils/error'

definePageMeta({ middleware: 'auth' })

const { t, locale } = useI18n()
const { internalControlCaps } = useRbac()
const caps = computed(() => internalControlCaps.value)
const store = useIcAssessmentStore()
const modal = useGlobalModalStore()
const toast = useAppToast()
const route = useRoute()
const router = useRouter()

// Select items cannot carry '' as a value, so "all" / "department-wide" use sentinels
const ALL = '__all'
const DEPT_WIDE = '__dept'
const NOT_STARTED = 'NOT_STARTED'

const thisYear = currentFiscalYear()
const qNum = (v: unknown, fallback: number) => {
  const n = Number(Array.isArray(v) ? v[0] : v)
  return Number.isInteger(n) ? n : fallback
}
const qStr = (v: unknown, fallback: string) => {
  const s = Array.isArray(v) ? v[0] : v
  return typeof s === 'string' && s ? s : fallback
}

// Filters live in the URL so the back button from a detail page returns to the same view
const year = ref<number>(qNum(route.query.year, thisYear))
const quarter = ref<number>(Math.min(4, Math.max(0, qNum(route.query.quarter, 0))))
const departmentFilter = ref<string>(qStr(route.query.department_id, ALL))
const branchFilter = ref<string>(qStr(route.query.location_id, DEPT_WIDE))
const statusFilter = ref<string>(qStr(route.query.status, ALL))

const departments = ref<Department[]>([])
const departmentsFailed = ref(false)
const locations = ref<Location[]>([])
const listFailed = ref(false)
const creatingFor = ref<string | null>(null)
const lockOpen = ref(false)
const unlockOpen = ref(false)

const yearItems = computed(() => Array.from({ length: 7 }, (_, i) => thisYear + 1 - i).map(y => ({ label: String(y), value: y })))
const quarterItems = computed(() => [
  { label: t('icAssessment.filters.annual'), value: 0 },
  ...[1, 2, 3, 4].map(n => ({ label: t('icAssessment.filters.quarterN', { n }), value: n }))
])
const departmentItems = computed(() => [
  { label: t('icAssessment.filters.allDepartments'), value: ALL },
  ...departments.value
    .filter(d => d.is_active !== false)
    .map(d => ({ label: d.department_code ? `${d.department_name} (${d.department_code})` : d.department_name, value: d.id }))
])
const branchItems = computed(() => [
  { label: t('icAssessment.filters.allBranches'), value: DEPT_WIDE },
  ...locations.value.filter(l => l.is_active !== false).map(l => ({ label: l.name, value: l.id }))
])
const statusItems = computed(() => [
  { label: t('icAssessment.filters.allStatuses'), value: ALL },
  ...ASSESSMENT_STATUSES.map(s => ({ label: t(`icAssessment.status.${s}`), value: s })),
  ...(caps.value.fullRead ? [{ label: t('icAssessment.filters.notStarted'), value: NOT_STARTED }] : [])
])

const locationId = computed(() => (branchFilter.value === DEPT_WIDE ? null : branchFilter.value))
const currentPeriodLabel = computed(() => periodLabel(t, year.value, quarter.value))
const locked = computed(() => isLockedFor(store.periods, year.value, quarter.value))
// The row that causes the lock: the year row wins over the quarter row
const lockedRow = computed(() =>
  store.periods.find(p => p.year === year.value && p.quarter === 0 && p.status === 'LOCKED')
  ?? store.periods.find(p => p.year === year.value && p.quarter === quarter.value && p.status === 'LOCKED')
  ?? null
)

interface GridRow {
  key: string
  departmentId: string
  departmentName: string
  departmentCode: string
  item: IcAssessmentListItem | null
}

/** Items of the selected scope (department-wide or the selected branch). */
const scopedItems = computed(() => store.items.filter(i => (i.locationId || null) === locationId.value))

const rows = computed<GridRow[]>(() => {
  const status = statusFilter.value
  const fromItem = (i: IcAssessmentListItem): GridRow => ({
    key: i.id,
    departmentId: i.departmentId,
    departmentName: i.departmentName || i.departmentCode || i.departmentId,
    departmentCode: i.departmentCode,
    item: i
  })
  // DEPARTMENT_HEAD / AUDITEE: only what the server returned (approved, own department)
  if (!caps.value.fullRead) return store.items.map(fromItem)

  const byDept = new Map(scopedItems.value.map(i => [i.departmentId, i]))
  const deptList = departments.value
    .filter(d => d.is_active !== false)
    .filter(d => departmentFilter.value === ALL || d.id === departmentFilter.value)
  const out: GridRow[] = deptList.map(d => ({
    key: d.id,
    departmentId: d.id,
    departmentName: d.department_name,
    departmentCode: d.department_code,
    item: byDept.get(d.id) ?? null
  }))
  // Assessments of departments missing from the master list (inactive, or the list failed)
  const known = new Set(deptList.map(d => d.id))
  for (const i of scopedItems.value) if (!known.has(i.departmentId)) out.push(fromItem(i))

  return out
    .filter(r => status === ALL || (status === NOT_STARTED ? !r.item : r.item?.status === status))
    .sort((a, b) => a.departmentName.localeCompare(b.departmentName))
})

const cell = (item: IcAssessmentListItem, code: ComponentCode) => item.components.find(c => c.code === code)

const load = async () => {
  listFailed.value = false
  const ok = await store.fetchList({
    year: year.value,
    quarter: quarter.value,
    department_id: departmentFilter.value === ALL ? undefined : departmentFilter.value,
    location_id: locationId.value ?? undefined,
    status: statusFilter.value === ALL || statusFilter.value === NOT_STARTED ? undefined : statusFilter.value
  })
  listFailed.value = !ok
  await store.fetchPeriods(year.value)
}

const loadMaster = async () => {
  const [deps, locs] = await Promise.all([
    caps.value.fullRead ? useDepartmentApi().getAllDepartments() : Promise.resolve([] as Department[]),
    useLocationApi().getAllLocations()
  ])
  departments.value = deps
  departmentsFailed.value = caps.value.fullRead && !deps.length
  locations.value = locs
}

watch([year, quarter, departmentFilter, branchFilter, statusFilter], () => {
  router.replace({
    query: {
      ...route.query,
      year: String(year.value),
      quarter: String(quarter.value),
      department_id: departmentFilter.value === ALL ? undefined : departmentFilter.value,
      location_id: locationId.value ?? undefined,
      status: statusFilter.value === ALL ? undefined : statusFilter.value
    }
  })
  load()
})

onMounted(() => {
  loadMaster()
  load()
})

const createFor = async (row: GridRow) => {
  const ok = await modal.confirmSubmit({
    description: t('icAssessment.grid.create') + ': ' + row.departmentName + ', ' + currentPeriodLabel.value,
    body: [locationId.value ? (locations.value.find(l => l.id === locationId.value)?.name || '') : t('icAssessment.filters.allBranches')].filter(Boolean),
    confirmLabel: t('icAssessment.grid.create')
  })
  if (!ok) return
  creatingFor.value = row.departmentId
  const created = await store.create({ department_id: row.departmentId, year: year.value, quarter: quarter.value, location_id: locationId.value })
  creatingFor.value = null
  if (!created) {
    toast.error(t('icAssessment.toast.actionFailed'), getUserErrorMessage(store.lastError, t))
    if (store.conflict) load()
    return
  }
  toast.success(t('icAssessment.toast.created'))
  router.push(`/internal-control-assessment/${created.id}`)
}

const doLock = async (note: string) => {
  const ok = await store.lockPeriod(year.value, quarter.value, note)
  if (!ok) return toast.error(t('icAssessment.toast.actionFailed'), getUserErrorMessage(store.lastError, t))
  lockOpen.value = false
  toast.success(t('icAssessment.period.locked'))
}

const doUnlock = async (reason: string) => {
  const row = lockedRow.value
  const ok = await store.unlockPeriod(row?.year ?? year.value, row?.quarter ?? quarter.value, reason)
  if (!ok) return toast.error(t('icAssessment.toast.actionFailed'), getUserErrorMessage(store.lastError, t))
  unlockOpen.value = false
  toast.success(t('icAssessment.period.unlocked'))
}

const formatDate = (iso: string | null) => {
  if (!iso) return '-'
  const d = new Date(iso)
  return isNaN(d.getTime()) ? iso : d.toLocaleDateString(locale.value === 'en' ? 'en-US' : 'id-ID', { dateStyle: 'medium' })
}
</script>
