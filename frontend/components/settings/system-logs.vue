<template>
  <div class="wrapper-vertical-expanded space-y-4 min-w-0">
    <!-- No view_system_logs (backend 403) -->
    <UCard
      v-if="forbidden"
      class="w-full dark:bg-slate-900 dark:ring-slate-800"
    >
      <div class="flex flex-col items-center justify-center text-center py-10 px-4">
        <UIcon
          name="i-lucide-shield-x"
          class="size-10 text-gray-400 dark:text-gray-500 mb-3"
        />
        <p class="text-sm font-semibold text-gray-900 dark:text-white">
          {{ t('settings.systemLogs.forbiddenTitle') }}
        </p>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1 max-w-md">
          {{ t('settings.systemLogs.forbiddenDesc') }}
        </p>
      </div>
    </UCard>

    <UCard
      v-else
      class="w-full dark:bg-slate-900 dark:ring-slate-800"
      :ui="{ body: 'p-0 sm:p-0' }"
    >
      <template #header>
        <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div class="flex items-start gap-3 min-w-0">
            <div class="rounded-lg bg-primary/10 text-primary p-2 shrink-0">
              <UIcon
                name="i-lucide-activity"
                class="size-5"
              />
            </div>
            <div class="min-w-0">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('settings.systemLogs.title') }}
              </h3>
              <p class="text-sm text-gray-600 dark:text-gray-400">
                {{ t('settings.systemLogs.subtitle') }}
              </p>
            </div>
          </div>

          <!-- Live indicator + pause/resume -->
          <div class="flex items-center gap-2 shrink-0">
            <div
              class="flex items-center gap-2 px-2.5 py-1 rounded-lg border text-xs font-medium"
              :class="liveBadgeClass"
              role="status"
              aria-live="polite"
            >
              <span>{{ liveLabel }}</span>
              <span
                v-if="lastUpdated"
                class="tabular-nums text-gray-500 dark:text-gray-400 hidden sm:inline"
              >
                {{ t('settings.systemLogs.updatedAt', { time: formatClock(lastUpdated) }) }}
              </span>
            </div>
            <UTooltip :text="paused ? t('settings.systemLogs.resume') : t('settings.systemLogs.pause')">
              <UButton
                :icon="paused ? 'i-lucide-play' : 'i-lucide-pause'"
                color="neutral"
                variant="ghost"
                size="sm"
                :aria-label="paused ? t('settings.systemLogs.resume') : t('settings.systemLogs.pause')"
                :disabled="loading || !!error"
                @click="togglePaused"
              />
            </UTooltip>
            <UTooltip :text="t('settings.systemLogs.refresh')">
              <UButton
                icon="i-lucide-refresh-cw"
                color="primary"
                variant="ghost"
                size="sm"
                :aria-label="t('settings.systemLogs.refresh')"
                :loading="loading"
                @click="load"
              />
            </UTooltip>
          </div>
        </div>

        <!-- Filters -->
        <div class="mt-4 grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-[minmax(0,2fr)_minmax(0,1.3fr)_minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)_auto] gap-2 items-end">
          <UFormField
            :label="t('settings.systemLogs.filters.search')"
            class="sm:col-span-2 xl:col-span-1"
          >
            <UInput
              v-model="search"
              icon="i-lucide-search"
              :placeholder="t('settings.systemLogs.filters.searchPlaceholder')"
              class="w-full"
            />
          </UFormField>
          <UFormField :label="t('settings.systemLogs.filters.action')">
            <USelect
              v-model="actionFilter"
              :items="actionOptions"
              class="w-full"
            />
          </UFormField>
          <UFormField :label="t('settings.systemLogs.filters.result')">
            <USelect
              v-model="resultFilter"
              :items="resultOptions"
              class="w-full"
            />
          </UFormField>
          <UFormField :label="t('settings.systemLogs.filters.from')">
            <UInput
              v-model="fromDate"
              type="date"
              :max="toDate || undefined"
              class="w-full"
            />
          </UFormField>
          <UFormField :label="t('settings.systemLogs.filters.to')">
            <UInput
              v-model="toDate"
              type="date"
              :min="fromDate || undefined"
              class="w-full"
            />
          </UFormField>
          <UButton
            icon="i-lucide-x"
            color="neutral"
            variant="ghost"
            :disabled="!hasActiveFilters"
            class="justify-center"
            @click="clearFilters"
          >
            {{ t('settings.systemLogs.filters.clear') }}
          </UButton>
        </div>
        <p
          v-if="dateRangeInvalid"
          class="mt-2 text-xs text-error"
        >
          {{ t('settings.systemLogs.filters.invalidRange') }}
        </p>
      </template>

      <!-- Loading -->
      <div
        v-if="loading && !rows.length"
        class="divide-y divide-gray-100 dark:divide-gray-800"
        :aria-label="t('settings.systemLogs.loading')"
      >
        <div
          v-for="i in 6"
          :key="i"
          class="px-4 py-3 flex items-center gap-4"
        >
          <USkeleton class="h-4 w-24" />
          <USkeleton class="h-4 w-28" />
          <USkeleton class="h-4 flex-1" />
          <USkeleton class="h-5 w-16 rounded-full" />
        </div>
      </div>

      <!-- Error (first page) -->
      <div
        v-else-if="error"
        class="flex flex-col items-center justify-center text-center py-10 px-4"
      >
        <UIcon
          name="i-lucide-circle-alert"
          class="size-10 text-error mb-3"
        />
        <p class="text-sm font-semibold text-gray-900 dark:text-white">
          {{ t('settings.systemLogs.errorTitle') }}
        </p>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1 max-w-md break-words">
          {{ error }}
        </p>
        <UButton
          class="mt-4"
          icon="i-lucide-refresh-cw"
          color="primary"
          variant="soft"
          @click="load"
        >
          {{ t('settings.systemLogs.retry') }}
        </UButton>
      </div>

      <!-- Empty -->
      <div
        v-else-if="isEmpty"
        class="flex flex-col items-center justify-center text-center py-10 px-4"
      >
        <UIcon
          name="i-lucide-history"
          class="size-10 text-gray-400 dark:text-gray-500 mb-3 opacity-70"
        />
        <p class="text-sm font-semibold text-gray-900 dark:text-white">
          {{ hasActiveFilters ? t('settings.systemLogs.emptyFilteredTitle') : t('settings.systemLogs.emptyTitle') }}
        </p>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
          {{ hasActiveFilters ? t('settings.systemLogs.emptyFilteredDesc') : t('settings.systemLogs.emptyDesc') }}
        </p>
        <UButton
          v-if="hasActiveFilters"
          class="mt-4"
          icon="i-lucide-x"
          color="neutral"
          variant="soft"
          @click="clearFilters"
        >
          {{ t('settings.systemLogs.filters.clear') }}
        </UButton>
      </div>

      <!-- Rows -->
      <div
        v-else
        class="min-w-0"
      >
        <!-- Desktop header -->
        <div
          class="hidden lg:grid gap-3 px-4 py-2.5 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/40 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400"
          :class="gridCols"
        >
          <span>{{ t('settings.systemLogs.columns.time') }}</span>
          <span>{{ t('settings.systemLogs.columns.user') }}</span>
          <span>{{ t('settings.systemLogs.columns.action') }}</span>
          <span>{{ t('settings.systemLogs.columns.entity') }}</span>
          <span>{{ t('settings.systemLogs.columns.result') }}</span>
          <span class="hidden xl:block">{{ t('settings.systemLogs.columns.ip') }}</span>
          <span>{{ t('settings.systemLogs.columns.summary') }}</span>
        </div>

        <ul class="divide-y divide-gray-100 dark:divide-gray-800">
          <li
            v-for="row in rows"
            :key="row.id"
          >
            <button
              type="button"
              class="w-full text-left px-4 py-3 hover:bg-gray-50 dark:hover:bg-gray-800/40 focus-visible:outline-none focus-visible:bg-gray-50 dark:focus-visible:bg-gray-800/40 transition-colors"
              :aria-expanded="expandedId === row.id"
              :aria-controls="`syslog-details-${row.id}`"
              @click="toggleRow(row.id)"
            >
              <!-- Desktop row -->
              <div
                class="hidden lg:grid gap-3 items-center text-sm"
                :class="gridCols"
              >
                <UTooltip :text="formatAbsoluteTime(row.occurred_at, locale)">
                  <span class="tabular-nums text-gray-700 dark:text-gray-300 truncate">
                    {{ formatRelativeTime(row.occurred_at, now, locale) }}
                  </span>
                </UTooltip>
                <span
                  class="truncate font-medium text-gray-900 dark:text-white"
                  :title="row.actor_username || ''"
                >
                  {{ row.actor_username || t('settings.systemLogs.systemActor') }}
                </span>
                <span
                  class="truncate text-gray-800 dark:text-gray-200"
                  :title="row.action"
                >
                  {{ actionLabel(row.action) }}
                </span>
                <span
                  class="truncate text-gray-600 dark:text-gray-400"
                  :title="entityTitle(row)"
                >
                  {{ row.entity_type || '-' }}
                </span>
                <span>
                  <UBadge
                    :label="resultLabel(row.result)"
                    :color="row.result === 'success' ? 'success' : 'error'"
                    variant="soft"
                    size="sm"
                  />
                </span>
                <span
                  class="hidden xl:block truncate font-mono text-xs text-gray-600 dark:text-gray-400"
                  :title="row.ip_address || ''"
                >
                  {{ row.ip_address || '-' }}
                </span>
                <span class="flex items-center gap-2 min-w-0">
                  <span
                    class="truncate text-gray-700 dark:text-gray-300"
                    :title="row.summary || ''"
                  >
                    {{ row.summary || '-' }}
                  </span>
                  <UIcon
                    :name="expandedId === row.id ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
                    class="size-4 text-gray-400 shrink-0 ml-auto"
                  />
                </span>
              </div>

              <!-- Mobile / tablet stacked row -->
              <div class="lg:hidden space-y-1.5 min-w-0">
                <div class="flex items-start justify-between gap-2">
                  <span class="text-sm font-semibold text-gray-900 dark:text-white break-words min-w-0">
                    {{ actionLabel(row.action) }}
                  </span>
                  <UBadge
                    :label="resultLabel(row.result)"
                    :color="row.result === 'success' ? 'success' : 'error'"
                    variant="soft"
                    size="sm"
                    class="shrink-0"
                  />
                </div>
                <div class="flex flex-wrap items-center gap-x-2 gap-y-0.5 text-xs text-gray-500 dark:text-gray-400">
                  <span class="font-medium text-gray-700 dark:text-gray-300">{{ row.actor_username || t('settings.systemLogs.systemActor') }}</span>
                  <span aria-hidden="true">·</span>
                  <span
                    class="tabular-nums"
                    :title="formatAbsoluteTime(row.occurred_at, locale)"
                  >{{ formatRelativeTime(row.occurred_at, now, locale) }}</span>
                  <template v-if="row.entity_type">
                    <span aria-hidden="true">·</span>
                    <span class="break-all">{{ row.entity_type }}</span>
                  </template>
                </div>
                <p
                  v-if="row.summary"
                  class="text-sm text-gray-700 dark:text-gray-300 break-words"
                >
                  {{ row.summary }}
                </p>
              </div>
            </button>

            <!-- Expanded details -->
            <div
              v-if="expandedId === row.id"
              :id="`syslog-details-${row.id}`"
              class="px-4 pb-4 pt-1 bg-gray-50/70 dark:bg-gray-900/30"
            >
              <dl class="grid grid-cols-1 sm:grid-cols-[10rem_minmax(0,1fr)] gap-x-4 gap-y-2 text-sm pt-3">
                <dt class="text-gray-500 dark:text-gray-400">
                  {{ t('settings.systemLogs.detail.time') }}
                </dt>
                <dd class="tabular-nums text-gray-900 dark:text-white">
                  {{ formatAbsoluteTime(row.occurred_at, locale) }}
                </dd>
                <dt class="text-gray-500 dark:text-gray-400">
                  {{ t('settings.systemLogs.detail.action') }}
                </dt>
                <dd class="font-mono text-xs text-gray-900 dark:text-white break-all self-center">
                  {{ row.action }}
                </dd>
                <template v-if="row.entity_id">
                  <dt class="text-gray-500 dark:text-gray-400">
                    {{ t('settings.systemLogs.detail.entity') }}
                  </dt>
                  <dd class="text-gray-900 dark:text-white break-words">
                    {{ row.entity_type || '-' }}
                    <span class="font-mono text-xs text-gray-600 dark:text-gray-400 break-all">{{ row.entity_id }}</span>
                  </dd>
                </template>
                <template v-if="row.actor_user_id">
                  <dt class="text-gray-500 dark:text-gray-400">
                    {{ t('settings.systemLogs.detail.userId') }}
                  </dt>
                  <dd class="font-mono text-xs text-gray-900 dark:text-white break-all self-center">
                    {{ row.actor_user_id }}
                  </dd>
                </template>
                <dt class="text-gray-500 dark:text-gray-400">
                  {{ t('settings.systemLogs.detail.ip') }}
                </dt>
                <dd class="font-mono text-xs text-gray-900 dark:text-white break-all self-center">
                  {{ row.ip_address || '-' }}
                </dd>
                <dt class="text-gray-500 dark:text-gray-400">
                  {{ t('settings.systemLogs.detail.userAgent') }}
                </dt>
                <dd class="text-gray-900 dark:text-white break-words">
                  {{ row.user_agent || '-' }}
                </dd>
                <template v-if="row.summary">
                  <dt class="text-gray-500 dark:text-gray-400">
                    {{ t('settings.systemLogs.detail.summary') }}
                  </dt>
                  <dd class="text-gray-900 dark:text-white break-words">
                    {{ row.summary }}
                  </dd>
                </template>
                <dt class="text-gray-500 dark:text-gray-400">
                  {{ t('settings.systemLogs.detail.details') }}
                </dt>
                <dd class="min-w-0">
                  <span
                    v-if="isEmptyDetails(row.details)"
                    class="text-gray-500 dark:text-gray-400"
                  >{{ t('settings.systemLogs.detail.noDetails') }}</span>
                  <dl
                    v-else-if="detailEntries(row.details)"
                    class="grid grid-cols-1 sm:grid-cols-[minmax(0,12rem)_minmax(0,1fr)] gap-x-3 gap-y-1 rounded-lg border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] p-3"
                  >
                    <template
                      v-for="[key, value] in detailEntries(row.details)"
                      :key="key"
                    >
                      <dt class="text-xs text-gray-500 dark:text-gray-400 break-all">
                        {{ key }}
                      </dt>
                      <dd class="text-xs text-gray-900 dark:text-white break-all">
                        {{ value }}
                      </dd>
                    </template>
                  </dl>
                  <pre
                    v-else
                    class="text-xs whitespace-pre-wrap break-all rounded-lg border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] p-3 text-gray-900 dark:text-white"
                  >{{ rawDetails(row.details) }}</pre>
                </dd>
              </dl>
            </div>
          </li>
        </ul>

        <!-- Load more -->
        <div class="flex flex-col items-center gap-2 px-4 py-4 border-t border-gray-100 dark:border-gray-800">
          <p
            v-if="loadMoreError"
            class="text-xs text-error text-center break-words"
          >
            {{ loadMoreError }}
          </p>
          <UButton
            v-if="hasMore"
            icon="i-lucide-chevrons-down"
            color="neutral"
            variant="soft"
            :loading="loadingMore"
            @click="loadMore"
          >
            {{ t('settings.systemLogs.loadMore') }}
          </UButton>
          <p
            v-else
            class="text-xs text-gray-400 dark:text-gray-500"
          >
            {{ t('settings.systemLogs.endOfList') }}
          </p>
        </div>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useSystemLogs, type SystemLogEntry, type SystemLogFilters } from '~/composables/useSystemLogs'
import {
  KNOWN_SYSTEM_LOG_ACTIONS,
  actionKey,
  dateInputToRfc3339,
  detailEntries,
  formatAbsoluteTime,
  formatRelativeTime,
  intlLocale,
  isEmptyDetails,
  parseDetails
} from '~/utils/systemLogFormat'

const { t, locale } = useI18n()

const {
  rows,
  loading,
  loadingMore,
  error,
  loadMoreError,
  pollError,
  forbidden,
  lastUpdated,
  paused,
  visible,
  live,
  hasMore,
  isEmpty,
  load,
  loadMore,
  setFilters,
  togglePaused
} = useSystemLogs()

// Columns: time | user | action | entity | result | IP (xl only) | summary. minmax(0, …) keeps it inside the panel.
const gridCols = 'lg:grid-cols-[7rem_minmax(0,1fr)_minmax(0,1.3fr)_minmax(0,1fr)_5.5rem_minmax(0,2fr)] xl:grid-cols-[8rem_minmax(0,1fr)_minmax(0,1.3fr)_minmax(0,1fr)_5.5rem_minmax(0,0.9fr)_minmax(0,2fr)]'

// Relative times refresh on their own, independent of polling
const now = ref(Date.now())
let clock: ReturnType<typeof setInterval> | null = null
const tick = () => {
  now.value = Date.now()
}
onMounted(() => {
  clock = setInterval(tick, 15_000)
})
onBeforeUnmount(() => {
  if (clock) clearInterval(clock)
})
watch(lastUpdated, tick)

const expandedId = ref<string | null>(null)
const toggleRow = (id: string) => {
  expandedId.value = expandedId.value === id ? null : id
}

// Filters
const ALL = 'all'
const search = ref('')
const actionFilter = ref<string>(ALL)
const resultFilter = ref<string>(ALL)
const fromDate = ref('')
const toDate = ref('')

const dateRangeInvalid = computed(() => !!fromDate.value && !!toDate.value && fromDate.value > toDate.value)
const hasActiveFilters = computed(() =>
  !!search.value.trim() || actionFilter.value !== ALL || resultFilter.value !== ALL || !!fromDate.value || !!toDate.value
)

const buildFilters = (): SystemLogFilters => ({
  q: search.value.trim() || undefined,
  action: actionFilter.value !== ALL ? actionFilter.value : undefined,
  result: resultFilter.value === 'success' || resultFilter.value === 'failure' ? resultFilter.value : undefined,
  from: dateInputToRfc3339(fromDate.value, 'start'),
  to: dateInputToRfc3339(toDate.value, 'end')
})

// Set while clearFilters resets every field, so the watchers below don't fire a load per field
let suppressWatchers = false

const applyFilters = () => {
  if (dateRangeInvalid.value) return
  expandedId.value = null
  void setFilters(buildFilters())
}

let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(search, () => {
  if (suppressWatchers) return
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(applyFilters, 350)
})
onBeforeUnmount(() => {
  if (searchTimer) clearTimeout(searchTimer)
})
watch([actionFilter, resultFilter, fromDate, toDate], () => {
  if (!suppressWatchers) applyFilters()
})

const clearFilters = async () => {
  if (searchTimer) clearTimeout(searchTimer)
  suppressWatchers = true
  search.value = ''
  actionFilter.value = ALL
  resultFilter.value = ALL
  fromDate.value = ''
  toDate.value = ''
  applyFilters()
  // Pre-flush watchers run before nextTick resolves
  await nextTick()
  suppressWatchers = false
}

const actionLabel = (action: string): string => {
  const path = `settings.systemLogs.actions.${actionKey(action)}`
  const label = t(path)
  return label === path ? action : label
}

// Known actions plus any action seen in the loaded rows, so new backend codes are filterable too
const actionOptions = computed(() => {
  const codes = new Set<string>(KNOWN_SYSTEM_LOG_ACTIONS)
  for (const r of rows.value) if (r.action) codes.add(r.action)
  if (actionFilter.value !== ALL) codes.add(actionFilter.value)
  const items = [...codes]
    .map(code => ({ label: actionLabel(code), value: code }))
    .sort((a, b) => a.label.localeCompare(b.label, intlLocale(locale.value)))
  return [{ label: t('settings.systemLogs.filters.allActions'), value: ALL }, ...items]
})

const resultOptions = computed(() => [
  { label: t('settings.systemLogs.filters.allResults'), value: ALL },
  { label: t('settings.systemLogs.result.success'), value: 'success' },
  { label: t('settings.systemLogs.result.failure'), value: 'failure' }
])

const resultLabel = (result: string) =>
  result === 'success' || result === 'failure' ? t(`settings.systemLogs.result.${result}`) : (result || '-')

const entityTitle = (row: SystemLogEntry) => [row.entity_type, row.entity_id].filter(Boolean).join(' ')

const rawDetails = (details: unknown) => {
  const d = parseDetails(details)
  return typeof d === 'string' ? d : JSON.stringify(d, null, 2)
}

const formatClock = (d: Date) =>
  new Intl.DateTimeFormat(intlLocale(locale.value), { timeStyle: 'medium' }).format(d)

// Live indicator
const liveState = computed<'live' | 'paused' | 'hidden' | 'retrying'>(() => {
  if (paused.value) return 'paused'
  if (!visible.value) return 'hidden'
  if (pollError.value) return 'retrying'
  return live.value ? 'live' : 'paused'
})
const liveLabel = computed(() => t(`settings.systemLogs.live.${liveState.value}`))
const liveDotClass = computed(() => ({
  live: 'bg-emerald-500',
  paused: 'bg-gray-400',
  hidden: 'bg-gray-400',
  retrying: 'bg-amber-500'
}[liveState.value]))
const liveBadgeClass = computed(() => liveState.value === 'live'
  ? 'border-emerald-200 dark:border-emerald-800/60 bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-400'
  : liveState.value === 'retrying'
    ? 'border-amber-200 dark:border-amber-800/60 bg-amber-50 dark:bg-amber-950/40 text-amber-700 dark:text-amber-400'
    : 'border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 text-gray-600 dark:text-gray-300')
</script>
