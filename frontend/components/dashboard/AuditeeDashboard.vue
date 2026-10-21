<template>
  <div class="space-y-6 sm:space-y-8 animate-fade-in">
    <!-- Header: Auditee profile & unit context -->
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4 p-5 sm:p-6 bg-gradient-to-r from-sky-900/10 via-primary-500/5 to-transparent border border-sky-500/20 rounded-2xl shadow-xs">
      <div class="flex items-start sm:items-center gap-4">
        <div class="size-14 rounded-2xl bg-sky-600/10 dark:bg-sky-500/20 border border-sky-500/30 flex items-center justify-center text-sky-600 dark:text-sky-400 shrink-0 shadow-xs">
          <UIcon
            name="i-lucide-user-check"
            class="size-7"
          />
        </div>
        <div>
          <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white tracking-tight">
            {{ t('dashboard.auditee.title') }}
          </h1>
          <p class="text-xs sm:text-sm text-gray-600 dark:text-gray-400 mt-1 max-w-3xl">
            {{ t('dashboard.auditee.welcome', { name: currentUserFullName }) }}
          </p>
          <p
            v-if="departmentName"
            class="text-[11px] text-gray-500 dark:text-gray-400 mt-1"
          >
            {{ t('dashboard.auditee.unitLabel', { unit: departmentName }) }}
          </p>
        </div>
      </div>

      <UButton
        variant="outline"
        color="neutral"
        size="sm"
        class="font-medium cursor-pointer self-start lg:self-center shrink-0"
        :loading="isRefreshing"
        @click="refreshAll"
      >
        <template #leading>
          <UIcon
            name="i-lucide-rotate-cw"
            :class="{ 'animate-spin': isRefreshing }"
            class="size-4"
          />
        </template>
        {{ t('dashboard.auditee.refresh') }}
      </UButton>
    </div>

    <!-- Notice -->
    <div class="flex items-start gap-3 p-4 rounded-xl bg-secondary-500/10 border border-secondary-500/20 text-secondary-900 dark:text-secondary-200 text-xs sm:text-sm">
      <UIcon
        name="i-lucide-info"
        class="size-5 text-secondary-600 dark:text-secondary-400 shrink-0"
      />
      <div>
        <span class="font-bold">{{ t('dashboard.auditee.notice.label') }}</span>
        <span class="ml-1 text-secondary-800 dark:text-secondary-300">{{ t('dashboard.auditee.notice.text') }}</span>
      </div>
    </div>

    <!-- Account link state (department is needed for the engagement widget) -->
    <AuditorWidgetState
      v-if="linkState !== 'ready'"
      :state="linkState === 'loading' ? 'loading' : 'error'"
      :error-text="t('dashboard.auditee.errors.link')"
      :retry-label="t('dashboard.auditee.states.retry')"
      @retry="loadEmployeeLink"
    />
    <div
      v-else-if="!hasDepartmentLink"
      role="status"
      class="flex items-start gap-3 p-4 rounded-xl bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-900/40 text-xs sm:text-sm text-amber-800 dark:text-amber-300"
    >
      <UIcon
        name="i-lucide-triangle-alert"
        class="size-5 shrink-0"
      />
      <span>{{ t('dashboard.auditee.errors.notLinked') }}</span>
    </div>

    <!-- 1. Audit statistics for own action plans (ATR) -->
    <section
      class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-5"
      aria-labelledby="auditee-stats-title"
    >
      <div>
        <h2
          id="auditee-stats-title"
          class="text-lg font-bold text-gray-900 dark:text-white"
        >
          {{ t('dashboard.auditee.stats.title') }}
        </h2>
        <p class="text-[11px] text-gray-400 dark:text-gray-500 mt-0.5">
          {{ t('dashboard.auditee.stats.scope') }}
        </p>
      </div>
      <AuditorWidgetState
        :state="atrState"
        :empty-text="t('dashboard.auditee.stats.empty')"
        :error-text="t('dashboard.auditee.states.loadError')"
        :retry-label="t('dashboard.auditee.states.retry')"
        @retry="reloadAtr"
      >
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
          <div
            v-for="card in statCards"
            :key="card.key"
            class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] border border-gray-200/80 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30"
          >
            <div class="flex items-start justify-between gap-2">
              <span class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">{{ card.label }}</span>
              <div
                class="p-2 rounded-lg shrink-0"
                :class="card.iconBg"
              >
                <UIcon
                  :name="card.icon"
                  class="size-4"
                  :class="card.iconClass"
                />
              </div>
            </div>
            <div class="mt-3">
              <div class="text-2xl font-black text-gray-900 dark:text-white">
                {{ card.value }}
              </div>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 mt-1">
                {{ card.hint }}
              </p>
            </div>
          </div>
        </div>
      </AuditorWidgetState>
    </section>

    <!-- 2. Engagements affecting the department -->
    <section
      class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4"
      aria-labelledby="auditee-engagements-title"
    >
      <div>
        <h2
          id="auditee-engagements-title"
          class="text-lg font-bold text-gray-900 dark:text-white"
        >
          {{ t('dashboard.auditee.engagements.title') }}
        </h2>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
          {{ t('dashboard.auditee.engagements.description') }}
        </p>
        <p class="text-[11px] text-gray-400 dark:text-gray-500 mt-0.5">
          {{ t('dashboard.auditee.engagements.scope') }}
        </p>
      </div>

      <AuditorWidgetState
        :state="engagementsState"
        :empty-text="hasDepartmentLink ? t('dashboard.auditee.engagements.empty') : t('dashboard.auditee.errors.notLinkedShort')"
        :error-text="t('dashboard.auditee.engagements.loadError')"
        :retry-label="t('dashboard.auditee.states.retry')"
        @retry="fetchEngagements"
      >
        <ul class="divide-y divide-gray-100 dark:divide-gray-800">
          <li
            v-for="row in engagements"
            :key="row.id"
            class="py-3 flex flex-col sm:flex-row sm:items-center justify-between gap-2"
          >
            <div class="min-w-0">
              <p class="text-sm font-semibold text-gray-900 dark:text-white truncate">
                {{ row.title || row.projectCode || '-' }}
              </p>
              <p class="text-[11px] text-gray-500 dark:text-gray-400">
                <span v-if="row.projectCode">{{ row.projectCode }} · </span>
                {{ t('dashboard.auditee.engagements.period', { start: formatDate(row.plannedStart), end: formatDate(row.plannedEnd) }) }}
              </p>
            </div>
            <UBadge
              :color="engagementStatusColor(row.status)"
              variant="subtle"
              size="sm"
              class="font-bold shrink-0 self-start sm:self-center"
            >
              {{ engagementStatusLabel(row.status) }}
            </UBadge>
          </li>
        </ul>
        <p
          v-if="engagementsTotal > engagements.length"
          class="pt-3 text-[11px] text-gray-500 dark:text-gray-400"
        >
          {{ t('dashboard.auditee.engagements.showing', { shown: engagements.length, total: engagementsTotal }) }}
        </p>
      </AuditorWidgetState>
    </section>

    <div class="grid grid-cols-1 xl:grid-cols-2 gap-6 sm:gap-8">
      <!-- 3. Tasks and deadlines (ATR action plans assigned to the user) -->
      <section
        class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4"
        aria-labelledby="auditee-tasks-title"
      >
        <div>
          <h2
            id="auditee-tasks-title"
            class="text-lg font-bold text-gray-900 dark:text-white"
          >
            {{ t('dashboard.auditee.tasks.title') }}
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {{ t('dashboard.auditee.tasks.description') }}
          </p>
          <p class="text-[11px] text-gray-400 dark:text-gray-500 mt-0.5">
            {{ t('dashboard.auditee.tasks.scope') }}
          </p>
        </div>

        <AuditorWidgetState
          :state="atrState"
          :empty-text="t('dashboard.auditee.tasks.empty')"
          :error-text="t('dashboard.auditee.states.loadError')"
          :retry-label="t('dashboard.auditee.states.retry')"
          @retry="reloadAtr"
        >
          <ul class="divide-y divide-gray-100 dark:divide-gray-800">
            <li
              v-for="item in openTasks"
              :key="item.id"
              class="py-3 flex items-start justify-between gap-3"
            >
              <div class="min-w-0">
                <p class="text-sm font-semibold text-gray-900 dark:text-white truncate">
                  {{ item.finding_title || item.report_number || '-' }}
                </p>
                <p class="text-[11px] text-gray-500 dark:text-gray-400">
                  {{ item.due_date ? t('dashboard.auditee.tasks.dueOn', { date: formatDate(item.due_date) }) : t('dashboard.auditee.tasks.noDueDate') }}
                </p>
              </div>
              <div class="flex flex-col items-end gap-1 shrink-0">
                <UBadge
                  :color="atrStatusBadgeColor(item.status)"
                  variant="subtle"
                  size="sm"
                  class="font-bold"
                >
                  {{ atrStatusLabel(item.status) }}
                </UBadge>
                <span
                  v-if="item.is_overdue"
                  class="text-[11px] font-bold text-red-600 dark:text-red-400"
                >
                  {{ t('dashboard.auditee.tasks.overdueDays', { days: item.overdue_days }) }}
                </span>
              </div>
            </li>
          </ul>
          <p
            v-if="openAtrItems.length > openTasks.length"
            class="pt-3 text-[11px] text-gray-500 dark:text-gray-400"
          >
            {{ t('dashboard.auditee.tasks.showing', { shown: openTasks.length, total: openAtrItems.length }) }}
          </p>
        </AuditorWidgetState>
      </section>

      <!-- 4. Requests needing a response (ATR revisions now; PBC and comments pending) -->
      <section
        class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4"
        aria-labelledby="auditee-requests-title"
      >
        <div>
          <h2
            id="auditee-requests-title"
            class="text-lg font-bold text-gray-900 dark:text-white"
          >
            {{ t('dashboard.auditee.requests.title') }}
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {{ t('dashboard.auditee.requests.description') }}
          </p>
        </div>

        <AuditorWidgetState
          :state="atrState"
          :empty-text="t('dashboard.auditee.requests.empty')"
          :error-text="t('dashboard.auditee.states.loadError')"
          :retry-label="t('dashboard.auditee.states.retry')"
          @retry="reloadAtr"
        >
          <ul
            v-if="revisionItems.length"
            class="divide-y divide-gray-100 dark:divide-gray-800"
          >
            <li
              v-for="item in revisionItems"
              :key="item.id"
              class="py-3 space-y-1"
            >
              <p class="text-sm font-semibold text-gray-900 dark:text-white truncate">
                {{ item.finding_title || item.report_number || '-' }}
              </p>
              <p class="text-xs text-amber-700 dark:text-amber-300">
                {{ t('dashboard.auditee.requests.revisionLabel') }}
              </p>
              <p
                v-if="item.review_note"
                class="text-[11px] text-gray-500 dark:text-gray-400 line-clamp-2"
              >
                {{ item.review_note }}
              </p>
            </li>
          </ul>
          <div
            v-else
            class="py-4 text-xs text-gray-500 dark:text-gray-400"
          >
            {{ t('dashboard.auditee.requests.noRevisions') }}
          </div>
        </AuditorWidgetState>

        <!-- TODO(B6, B7, B9): PBC requests and comments need a request workflow and a notifications endpoint.
             Planned: GET /api/v1/fieldwork/documents (status, due date), PUT /api/v1/fieldwork/documents/:id/respond,
             comment threads (B7) and GET /api/v1/notifications (B9). Not connected yet. -->
        <div class="flex items-start gap-2 p-3 rounded-xl bg-gray-50 dark:bg-gray-800/40 border border-dashed border-gray-300 dark:border-gray-700 text-[11px] text-gray-500 dark:text-gray-400">
          <UIcon
            name="i-lucide-construction"
            class="size-4 shrink-0"
          />
          <span>{{ t('dashboard.auditee.requests.placeholder') }}</span>
        </div>
      </section>
    </div>

    <!-- 5. Important dates (calendar) -->
    <section
      class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4"
      aria-labelledby="auditee-calendar-title"
    >
      <div>
        <h2
          id="auditee-calendar-title"
          class="text-lg font-bold text-gray-900 dark:text-white"
        >
          {{ t('dashboard.auditee.calendar.title') }}
        </h2>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
          {{ t('dashboard.auditee.calendar.description') }}
        </p>
      </div>
      <!-- TODO(B4): GET /api/v1/me/calendar?from&to, filtered to the user's department engagements and ATR due dates.
           Until it exists this widget only shows its placeholder state (no dates are derived here). -->
      <AuditorWidgetState
        state="empty"
        :empty-text="t('dashboard.auditee.calendar.placeholder')"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
// Auditee dashboard: the audits that affect the user's department and the action plans (ATR) the user must answer.
// Corporate-level metrics (annual plan, risk, analytics) stay out of this view.
import { computed, onMounted, ref } from 'vue'
import { useAuthStore } from '~/stores/auth'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useI18n } from '~/composables/useI18n'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { extractErrorMessage } from '~/utils/error'
import {
  atrStatusBadgeColor,
  atrStatusI18nKey,
  atrDateOnly,
  formatAtrDate,
  isAtrClosed,
  normalizeAtrStatus
} from '~/utils/actionTakenReport'
import type { ActionTakenReport } from '~/types/audit'

type WidgetState = 'loading' | 'error' | 'empty' | 'ready'

interface AuditeeEngagement {
  id: string
  title: string
  projectCode: string
  status: string
  plannedStart: string
  plannedEnd: string
}

const ENGAGEMENT_PAGE_SIZE = 50
const TASK_PREVIEW_LIMIT = 8
const ENGAGEMENT_PREVIEW_LIMIT = 10

const { t, locale } = useI18n()
const authStore = useAuthStore()
const atrStore = useActionTakenReportStore()

// --- Identity -----------------------------------------------------------------------------------
const currentUserFullName = computed(() => {
  return authStore.user?.fullName || authStore.user?.username || ''
})
const departmentName = computed(() => authStore.user?.department || '')
// Master-data department UUID. Null until /auth/me has been read, or when the account has no linked employee.
const departmentId = computed(() => authStore.user?.departmentId || null)
const hasDepartmentLink = computed(() => !!departmentId.value)

const linkState = ref<WidgetState>('loading')

/** Reads employee and department from GET /auth/me (spec B1). Never fakes a department. */
const loadEmployeeLink = async () => {
  linkState.value = 'loading'
  try {
    await authStore.fetchEmployeeLink()
    linkState.value = 'ready'
  } catch (error) {
    console.error('Failed to load employee link:', error)
    linkState.value = 'error'
  }
}

// --- Engagements affecting the department -------------------------------------------------------
const engagements = ref<AuditeeEngagement[]>([])
const engagementsTotal = ref(0)
const engagementsLoading = ref(false)
const engagementsError = ref('')
const engagementsDone = ref(false)

const normalizeEngagement = (raw: Record<string, unknown>): AuditeeEngagement => ({
  id: String(raw.id ?? ''),
  title: String(raw.title ?? ''),
  projectCode: String(raw.project_code ?? ''),
  status: String(raw.status ?? ''),
  plannedStart: String(raw.planned_start ?? ''),
  plannedEnd: String(raw.planned_end ?? '')
})

/**
 * Engagements whose target unit is the user's department.
 * Interim source: GET /audit-activities?target_unit_id (the filter is applied by the server, the list itself is not
 * role-scoped yet).
 * TODO(B3): replace with GET /api/v1/me/engagements, which scopes the rows on the server for AUDITEE.
 */
const fetchEngagements = async () => {
  if (!departmentId.value) {
    engagements.value = []
    engagementsTotal.value = 0
    engagementsDone.value = true
    return
  }
  engagementsLoading.value = true
  engagementsError.value = ''
  try {
    const response = await $fetch<{ data?: { items?: unknown[], pagination?: { total?: number } } }>(
      `${getAuditServiceBaseUrl()}/audit-activities`,
      {
        method: 'GET',
        params: { target_unit_id: departmentId.value, page: 1, page_size: ENGAGEMENT_PAGE_SIZE }
      }
    )
    const rows = Array.isArray(response?.data?.items) ? response.data.items : []
    engagements.value = rows
      .map(row => normalizeEngagement(row as Record<string, unknown>))
      .sort((a, b) => (b.plannedStart || '').localeCompare(a.plannedStart || ''))
      .slice(0, ENGAGEMENT_PREVIEW_LIMIT)
    engagementsTotal.value = Number(response?.data?.pagination?.total) || rows.length
  } catch (error) {
    console.error('Failed to fetch engagements for the department:', error)
    engagementsError.value = extractErrorMessage(error, t('dashboard.auditee.engagements.loadError'))
    engagements.value = []
    engagementsTotal.value = 0
  } finally {
    engagementsLoading.value = false
    engagementsDone.value = true
  }
}

const engagementsState = computed<WidgetState>(() => {
  if (linkState.value === 'loading') return 'loading'
  if (linkState.value === 'error') return 'error'
  if (!hasDepartmentLink.value) return 'empty'
  if (engagements.value.length === 0 && (engagementsLoading.value || !engagementsDone.value)) return 'loading'
  if (engagements.value.length === 0 && engagementsError.value) return 'error'
  return engagements.value.length === 0 ? 'empty' : 'ready'
})

const ENGAGEMENT_STATUS_KEYS: Record<string, string> = {
  PLANNED: 'planned',
  IN_PROGRESS: 'inProgress',
  REPORTING: 'reporting',
  COMPLETED: 'completed'
}

const engagementStatusLabel = (status: string) => {
  const key = ENGAGEMENT_STATUS_KEYS[status]
  return key ? t(`dashboard.auditee.engagements.status.${key}`) : status || '-'
}

const engagementStatusColor = (status: string): 'info' | 'warning' | 'success' | 'neutral' => {
  if (status === 'COMPLETED') return 'success'
  if (status === 'REPORTING') return 'warning'
  if (status === 'IN_PROGRESS') return 'info'
  return 'neutral'
}

// --- Action plans (ATR) assigned to the user --------------------------------------------------
// The ATR list is already scoped on the server to the PIC (the caller's own items), so no client filter is needed.
const atrDone = ref(false)

const atrState = computed<WidgetState>(() => {
  const items = atrStore.summaryItems
  if (items.length === 0 && (atrStore.summaryLoading || !atrDone.value)) return 'loading'
  if (items.length === 0 && atrStore.summaryError) return 'error'
  return items.length === 0 ? 'empty' : 'ready'
})

const reloadAtr = async () => {
  await atrStore.fetchSummary()
}

const openAtrItems = computed<ActionTakenReport[]>(() => {
  return atrStore.summaryItems
    .filter(item => !isAtrClosed(item.status))
    .sort((a, b) => {
      // Overdue first, then the earliest due date; items without a due date go last.
      if (a.is_overdue !== b.is_overdue) return a.is_overdue ? -1 : 1
      const ad = atrDateOnly(a.due_date) ?? '9999-12-31'
      const bd = atrDateOnly(b.due_date) ?? '9999-12-31'
      return ad.localeCompare(bd)
    })
})

const openTasks = computed(() => openAtrItems.value.slice(0, TASK_PREVIEW_LIMIT))

// An ATR returned by the reviewer comes back to IN_PROGRESS with review_note set (ATRActionPlanForm.vue).
// Derived from the ATR list only. TODO(B8): replace with the server revision state once LHA revisions exist.
const revisionItems = computed(() => {
  return openAtrItems.value.filter(item =>
    normalizeAtrStatus(item.status) === 'IN_PROGRESS' && !!item.review_note
  )
})

const statCards = computed(() => {
  const items = atrStore.summaryItems
  const total = items.length
  const stats = atrStore.stats
  const open = stats.counts.open
  const overdue = stats.counts.overdue
  const completed = stats.counts.completed
  const compliance = total > 0 ? Math.round((completed / total) * 100) : 0
  return [
    {
      key: 'open',
      label: t('dashboard.auditee.stats.open'),
      value: String(open),
      hint: t('dashboard.auditee.stats.openHint'),
      icon: 'i-lucide-list-todo',
      iconBg: 'bg-sky-500/10',
      iconClass: 'text-sky-600 dark:text-sky-400'
    },
    {
      key: 'overdue',
      label: t('dashboard.auditee.stats.overdue'),
      value: String(overdue),
      hint: t('dashboard.auditee.stats.overdueHint'),
      icon: 'i-lucide-alarm-clock',
      iconBg: 'bg-red-500/10',
      iconClass: 'text-red-600 dark:text-red-400'
    },
    {
      key: 'review',
      label: t('dashboard.auditee.stats.pendingReview'),
      value: String(stats.counts.pendingReview),
      hint: t('dashboard.auditee.stats.pendingReviewHint'),
      icon: 'i-lucide-hourglass',
      iconBg: 'bg-amber-500/10',
      iconClass: 'text-amber-600 dark:text-amber-400'
    },
    {
      key: 'compliance',
      label: t('dashboard.auditee.stats.compliance'),
      value: `${compliance}%`,
      hint: t('dashboard.auditee.stats.complianceHint'),
      icon: 'i-lucide-badge-check',
      iconBg: 'bg-emerald-500/10',
      iconClass: 'text-emerald-600 dark:text-emerald-400'
    }
  ]
})

const atrStatusLabel = (status: string) => {
  const key = atrStatusI18nKey(status)
  return key ? t(key) : status
}

const formatDate = (value: string) => formatAtrDate(value, locale.value)

// --- Refresh ----------------------------------------------------------------------------------
const isRefreshing = ref(false)

const refreshAll = async () => {
  isRefreshing.value = true
  try {
    await loadEmployeeLink()
    await Promise.all([fetchEngagements(), reloadAtr()])
  } finally {
    isRefreshing.value = false
  }
}

onMounted(async () => {
  // The department is needed before the engagement list can be filtered, so the link is read first.
  // The ATR list does not depend on it and loads in parallel.
  const atrLoad = reloadAtr().finally(() => {
    atrDone.value = true
  })
  await loadEmployeeLink()
  await fetchEngagements()
  await atrLoad
})
</script>
