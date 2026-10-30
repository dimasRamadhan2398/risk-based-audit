<script setup lang="ts">
import { normalizeAuditStatus, useAuditExecutionStore, type AuditExecutionSyncResult } from '~/stores/audit-execution'
import AuditExecutionDetailModal from '~/components/audit-execution/AuditExecutionDetailModal.vue'
import { AuditCategory, AuditStatus, EXECUTION_PHASES, getExecutionPhase, type AuditExecution } from '~/types/audit'
import { getProgressValue } from '~/utils/executionSources'
import { createAutoSyncOnce } from '~/utils/auditExecutionAutoSync'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'

const { t, locale } = useI18n()
const toast = useToast()
const store = useAuditExecutionStore()
const { auditExecutions, getSummary, loading, error, syncing } = storeToRefs(store)

// Executions are created from assignment letters, so syncing needs letter write access (the backend enforces its own rule).
const { canManageAssignmentLetter } = useRbac()
const canSyncAudits = computed(() => canManageAssignmentLetter.value)

const showSyncToast = (result: AuditExecutionSyncResult) => {
  toast.add({
    title: t('auditExecution.sync.successTitle'),
    description: result.created > 0
      ? t('auditExecution.sync.successCreated', { created: result.created, synced: result.synced })
      : t('auditExecution.sync.successNoneCreated', { synced: result.synced }),
    color: 'success'
  })
  if (result.unmatched > 0) {
    toast.add({
      title: t('auditExecution.sync.unmatchedTitle'),
      description: t('auditExecution.sync.unmatchedWarning', { count: result.unmatched }),
      color: 'warning'
    })
  }
}

const syncFromAssignmentLetters = async () => {
  const result = await store.syncFromAssignmentLetters()
  if (result) {
    showSyncToast(result)
  } else if (store.syncError) {
    toast.add({
      title: t('auditExecution.sync.failedTitle'),
      description: store.syncError,
      color: 'error'
    })
  }
}

// Once per visit, an empty list triggers a silent sync so letters that already exist show up
// without a manual click. Only a sync that created audits is announced; failures stay quiet.
const autoSync = createAutoSyncOnce(() => store.syncFromAssignmentLetters())

const loadAudits = async () => {
  // Always refetch on mount: the store is shared with the dashboard and outlives navigation.
  await store.fetchAuditExecutions()
  const result = await autoSync.maybeRun({
    canSync: canSyncAudits.value,
    hasError: !!store.error,
    count: store.auditExecutions.length
  })
  if (result && result.created > 0) {
    toast.add({
      title: t('auditExecution.sync.autoCreatedTitle'),
      description: t('auditExecution.sync.autoCreatedDescription', { created: result.created }),
      color: 'success'
    })
  }
}
loadAudits()

// While a sync runs (including its refetch) the empty card stays up with a spinner instead of flipping to the table.
const showEmptyState = computed(() =>
  !error.value && auditExecutions.value.length === 0 && (syncing.value || !loading.value)
)

const search = ref('')
const quarter = ref('')
const status = ref<string | undefined>(undefined)
const category = ref<AuditCategory | undefined>(undefined)

const isHelpModalOpen = ref(false)

const quarters = ['Quarter I', 'Quarter II', 'Quarter III', 'Quarter IV']

const statusOptions = computed(() => [
  { label: locale.value === 'id' ? 'Planning (Perencanaan & Persiapan)' : 'Planning (Planning & Preparation)', value: 'planning' },
  { label: locale.value === 'id' ? 'Entry Meeting (Entry Meeting & Penyelarasan Scope)' : 'Entry Meeting (Entry Meeting & Scope Alignment)', value: 'entry meeting' },
  { label: locale.value === 'id' ? 'Fieldwork (Fieldwork & Pengujian Pengendalian)' : 'Fieldwork (Fieldwork & Control Testing)', value: 'fieldwork' },
  { label: locale.value === 'id' ? 'Draft Findings (Draft Temuan & Rekomendasi)' : 'Draft Findings (Draft Findings & Recommendations)', value: 'draft findings' },
  { label: locale.value === 'id' ? 'Reporting (Pelaporan & Exit Meeting)' : 'Reporting (Reporting & Exit Meeting)', value: 'reporting' },
  { label: locale.value === 'id' ? 'Completed (Audit Selesai)' : 'Completed (Audit Completed)', value: 'completed' }
])

const hasFilters = computed(() => !!(search.value || quarter.value || category.value || status.value))

const resetFilters = () => {
  search.value = ''
  quarter.value = ''
  category.value = undefined
  status.value = undefined
}

const columns = computed(() => [
  { accessorKey: 'name', header: t('auditExecution.columns.name') },
  { accessorKey: 'phase', header: t('auditExecution.columns.phase') },
  { accessorKey: 'progress', header: t('auditExecution.columns.progress') },
  { accessorKey: 'lead_auditor', header: t('auditExecution.columns.leadAuditor') },
  { accessorKey: 'actions', header: t('auditExecution.columns.actions') }
])

// Filter values are canonical stage statuses (see statusOptions); the row's status is normalised the same way as the store.
const isStatusMatch = (audit: AuditExecution, filterStatus?: string): boolean => {
  if (!filterStatus) return true
  return normalizeAuditStatus(audit.status, getProgressValue(audit.progress)) === filterStatus
}

const getQuarterFromDate = (dateStr?: string): string => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return ''
  const month = date.getMonth() + 1 // 1-12
  if (month <= 3) return 'Quarter I'
  if (month <= 6) return 'Quarter II'
  if (month <= 9) return 'Quarter III'
  return 'Quarter IV'
}

const filteredAudits = computed(() => {
  return auditExecutions.value.filter((audit) => {
    const matchesSearch = !search.value
      || (audit.name && audit.name.toLowerCase().includes(search.value.toLowerCase()))
      || (audit.category && audit.category.toLowerCase().includes(search.value.toLowerCase()))
      || (audit.ref && audit.ref.toLowerCase().includes(search.value.toLowerCase()))
    const matchesCategory = !category.value
      || (audit.category && audit.category.toLowerCase() === category.value.toLowerCase())
    const matchesStatus = isStatusMatch(audit, status.value)
    const matchesQuarter = !quarter.value
      || getQuarterFromDate((audit as any).created_at || (audit as any).createdAt) === quarter.value
    return matchesSearch && matchesCategory && matchesStatus && matchesQuarter
  })
})

const phaseBreakdown = computed(() => {
  const counts: Record<number, number> = { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0 }
  auditExecutions.value.forEach((audit) => {
    const phase = getExecutionPhase(audit.progress)
    counts[phase.step] = (counts[phase.step] || 0) + 1
  })
  return counts
})

const rowProgress = (row: any): number => getProgressValue((row?.original || row)?.progress)

const getProgressColor = (val: number) => {
  if (val >= 100) return 'bg-secondary-600 shadow-sm shadow-secondary-600/40'
  if (val >= 76) return 'bg-indigo-500 shadow-sm shadow-indigo-500/40'
  if (val >= 51) return 'bg-purple-500 shadow-sm shadow-purple-500/40'
  if (val >= 26) return 'bg-violet-500 shadow-sm shadow-violet-500/40'
  if (val >= 1) return 'bg-blue-500 shadow-sm shadow-blue-500/40'
  return 'bg-sky-400 shadow-sm shadow-sky-400/40'
}

const isDetailOpen = ref(false)
const selectedAudit = ref<any>(undefined)

const openDetail = (row: any) => {
  selectedAudit.value = row?.original || row
  isDetailOpen.value = true
}

const handleRemind = (audit: any) => {
  useToast().add({
    title: t('auditExecution.detailModal.reminderSent'),
    description: t('auditExecution.detailModal.reminderDescription', {
      name: audit?.name || audit?.nama_audit || '',
      auditor: audit?.lead_auditor || ''
    }),
    color: 'success'
  })
}
</script>

<template>
  <div class="p-4 sm:p-6 space-y-6 min-w-0">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('auditExecution.title') }}
        </h1>
        <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-1">
          {{ t('auditExecution.subtitle') }}
        </p>
      </div>

      <div class="flex flex-col sm:flex-row gap-2 w-full sm:w-auto">
        <UTooltip
          v-if="canSyncAudits"
          :text="t('auditExecution.sync.tooltip')"
        >
          <UButton
            icon="i-lucide-refresh-cw"
            :label="t('auditExecution.sync.button')"
            color="primary"
            variant="outline"
            size="md"
            class="font-bold w-full sm:w-auto justify-center"
            :loading="syncing"
            :disabled="syncing || loading"
            @click="syncFromAssignmentLetters"
          />
        </UTooltip>
        <UButton
          icon="i-lucide-help-circle"
          :label="t('auditExecution.cycleGuideButton')"
          color="primary"
          variant="soft"
          size="md"
          class="font-bold w-full sm:w-auto justify-center"
          @click="() => { isHelpModalOpen = true }"
        />
      </div>
    </div>

    <UAlert
      v-if="error"
      :title="t('auditExecution.loadError.title')"
      :description="error"
      color="error"
      variant="soft"
      icon="i-lucide-alert-circle"
      :actions="[{ label: t('auditExecution.loadError.retry'), icon: 'i-lucide-rotate-ccw', color: 'error', variant: 'outline', loading: loading, onClick: loadAudits }]"
    />

    <!-- Summary Section -->
    <div class="bg-white dark:bg-gray-900 rounded-xl p-4 border border-gray-200 dark:border-gray-800 shadow-sm flex flex-wrap items-center justify-between gap-4">
      <div class="flex flex-wrap items-center gap-4 sm:gap-6">
        <span class="text-md font-bold text-gray-500 tracking-tight">{{ t('auditExecution.summary.title') }}</span>
        <div class="flex items-center gap-2">
          <span class="w-3 h-3 rounded-full bg-emerald-500" />
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('auditExecution.summary.completed', { count: getSummary.completed }) }}</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="w-3 h-3 rounded-full bg-sky-500" />
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('auditExecution.summary.inProgress', { count: getSummary.inProgress }) }}</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="w-3 h-3 rounded-full bg-gray-300 dark:bg-gray-600" />
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('auditExecution.summary.planned', { count: getSummary.planned }) }}</span>
        </div>
      </div>
    </div>

    <!-- Filters Section -->
    <div class="flex flex-col sm:flex-row flex-wrap items-stretch sm:items-center gap-3 sm:gap-4 w-full">
      <CommonSearchInput
        v-model="search"
        :placeholder="t('auditExecution.filters.searchPlaceholder')"
        class="w-full sm:flex-1 sm:min-w-48"
      />
      <USelectMenu
        v-model="quarter"
        :items="quarters"
        :placeholder="t('auditExecution.filters.quarterPlaceholder')"
        class="w-full sm:w-48"
      />
      <USelectMenu
        v-model="category"
        :items="Object.values(AuditCategory)"
        :placeholder="t('auditExecution.filters.categoryPlaceholder')"
        class="w-full sm:w-48"
      />
      <USelectMenu
        v-model="status"
        :items="statusOptions"
        value-key="value"
        :placeholder="t('auditExecution.filters.statusPlaceholder')"
        class="w-full sm:w-64"
      />
      <UButton
        v-if="hasFilters"
        :label="t('auditExecution.resetFilter')"
        icon="i-lucide-rotate-ccw"
        color="neutral"
        variant="outline"
        class="w-full sm:w-auto justify-center"
        @click="resetFilters"
      />
    </div>

    <!-- Empty State: executions come from assignment letters -->
    <div
      v-if="showEmptyState"
      class="bg-white dark:bg-gray-900 rounded-xl border border-dashed border-gray-300 dark:border-gray-700 shadow-sm px-6 py-12 flex flex-col items-center text-center"
    >
      <template v-if="syncing">
        <UIcon
          name="i-lucide-loader-2"
          class="w-10 h-10 text-primary-500 animate-spin mb-3"
        />
        <p class="text-sm font-medium text-gray-600 dark:text-gray-300">
          {{ t('auditExecution.emptyState.syncing') }}
        </p>
      </template>
      <template v-else>
        <div class="p-3 rounded-full bg-primary-50 dark:bg-primary-950/50 text-primary-600 dark:text-primary-400 mb-4">
          <UIcon
            name="i-lucide-file-text"
            class="w-8 h-8"
          />
        </div>
        <h2 class="text-lg font-bold text-gray-900 dark:text-white">
          {{ t('auditExecution.emptyState.title') }}
        </h2>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1 max-w-xl leading-relaxed">
          {{ canSyncAudits ? t('auditExecution.emptyState.description') : t('auditExecution.emptyState.descriptionReadOnly') }}
        </p>
        <div class="flex flex-col sm:flex-row gap-2 mt-5 w-full sm:w-auto">
          <UButton
            to="/assignment-letter"
            icon="i-lucide-file-signature"
            :label="t('auditExecution.emptyState.goToAssignmentLetter')"
            color="neutral"
            variant="outline"
            class="justify-center"
          />
          <UButton
            v-if="canSyncAudits"
            icon="i-lucide-refresh-cw"
            :label="t('auditExecution.sync.button')"
            color="primary"
            class="justify-center font-bold"
            :loading="syncing"
            :disabled="syncing"
            @click="syncFromAssignmentLetters"
          />
        </div>
      </template>
    </div>

    <!-- Table Section -->
    <div
      v-else
      class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm"
    >
      <TableEntities
        :columns="columns"
        :data="filteredAudits"
        :loading="loading"
      >
        <template #name-cell="{ row }">
          <div class="flex flex-col">
            <span class="font-bold text-gray-900 dark:text-white">{{ (row.original || row).name }}</span>
            <span class="text-md text-gray-500 font-medium">{{ t('auditExecution.table.ref') }} {{ (row.original || row).ref || '-' }} | {{ (row.original || row).category }}</span>
          </div>
        </template>

        <template #phase-cell="{ row }">
          <UBadge
            :class="getExecutionPhase(rowProgress(row)).badgeClass"
            variant="subtle"
            size="md"
          >
            <UIcon
              :name="getExecutionPhase(rowProgress(row)).icon"
              class="mr-1.5 inline-block text-md"
            />
            {{ t(`auditExecution.phases.${getExecutionPhase(rowProgress(row)).step}.shortLabel`) }}
          </UBadge>
        </template>

        <template #progress-cell="{ row }">
          <div class="flex items-center gap-3 min-w-[200px]">
            <div class="flex-1 h-3 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden border border-gray-200 dark:border-gray-700 shadow-inner">
              <div
                class="h-full rounded-full transition-all duration-500 ease-out"
                :class="getProgressColor(rowProgress(row))"
                :style="{ width: `${Math.min(100, Math.max(0, rowProgress(row)))}%` }"
              />
            </div>
            <span class="text-md font-bold text-gray-700 dark:text-white w-10 text-right">
              {{ rowProgress(row) }}%
            </span>
          </div>
        </template>

        <template #lead_auditor-cell="{ row }">
          <span class="text-sm text-gray-700 dark:text-white">{{ (row.original || row).lead_auditor || '-' }}</span>
        </template>

        <template #actions-cell="{ row }">
          <UTooltip :text="t('qualityAssurance.dashboard.summary.viewDetail')">
            <UButton
              icon="i-lucide-eye"
              variant="ghost"
              color="neutral"
              size="md"
              @click="openDetail(row)"
            />
          </UTooltip>
        </template>
      </TableEntities>
    </div>

    <!-- Detail Modal -->
    <AuditExecutionDetailModal
      v-model:open="isDetailOpen"
      :audit="selectedAudit"
      @remind="handleRemind"
    />

    <!-- Lifecycle Phase Guide Help Modal -->
    <UModal
      v-model:open="isHelpModalOpen"
      :ui="{
        content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        header: 'border-b border-gray-100 dark:border-gray-800 p-4 sm:p-5 text-gray-900 dark:text-white font-bold shrink-0',
        body: 'p-4 sm:p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
        footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
      scrollable
    >
      <template #content>
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2.5">
                <div class="p-2 rounded-lg bg-secondary-100 dark:bg-secondary-950/60 text-secondary-600 dark:text-secondary-400">
                  <UIcon
                    name="i-lucide-route"
                    class="text-xl"
                  />
                </div>
                <div>
                  <h3 class="text-lg font-bold text-gray-900 dark:text-white">
                    {{ t('auditExecution.guideModal.title') }}
                  </h3>
                  <p class="text-md text-gray-500">
                    {{ t('auditExecution.guideModal.subtitle') }}
                  </p>
                </div>
              </div>
              <UButton
                color="neutral"
                variant="ghost"
                icon="i-lucide-x"
                @click="() => { isHelpModalOpen = false }"
              />
            </div>
          </template>

          <div class="space-y-4">
            <p class="text-md text-gray-600 dark:text-gray-300 leading-relaxed">
              {{ t('auditExecution.guideModal.description') }}
            </p>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div
                v-for="phase in EXECUTION_PHASES"
                :key="phase.step"
                class="p-4 rounded-xl border space-y-2 transition-all duration-200 hover:shadow-md"
                :class="phase.cardClass"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div
                      class="w-6 h-6 rounded-full flex items-center justify-center text-md font-bold shrink-0 shadow-md"
                      :class="phase.numBgClass"
                    >
                      {{ phase.step }}
                    </div>
                    <span class="text-sm font-bold text-gray-900 dark:text-white">{{ t(`auditExecution.phases.${phase.step}.title`) }}</span>
                  </div>
                  <span
                    class="text-md font-bold px-2 py-0.5 rounded-full shrink-0"
                    :class="phase.badgeClass"
                  >
                    {{ phase.minProgress === phase.maxProgress ? `${phase.minProgress}%` : `${phase.minProgress}–${phase.maxProgress}%` }}
                  </span>
                </div>
                <p class="text-[11px] text-gray-600 dark:text-gray-400 pl-8 leading-relaxed">
                  {{ t(`auditExecution.phases.${phase.step}.description`) }}
                </p>
              </div>
            </div>
          </div>

          <template #footer>
            <div class="flex justify-end">
              <UButton
                :label="t('auditExecution.guideModal.close')"
                color="primary"
                class="px-5 font-bold"
                @click="() => { isHelpModalOpen = false }"
              />
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
  </div>
</template>
