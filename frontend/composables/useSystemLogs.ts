// System Logs (admin user activity log) composable
// Every call to GET /api/v1/system-logs (auth-service) lives in this file, so a contract change is a one-file fix.
import { computed, getCurrentScope, onScopeDispose, ref, shallowRef } from 'vue'
import { getAuthServiceBaseUrl } from '~/composables/useApiUrl'
import { useRbac } from '~/composables/useRbac'
import { useAuthStore } from '~/stores/auth'
import { extractErrorMessage } from '~/utils/error'

export const SYSTEM_LOGS_PERMISSION = 'view_system_logs'

export type SystemLogResult = 'success' | 'failure'

export interface SystemLogEntry {
  id: string
  occurred_at: string
  actor_user_id: string | null
  actor_username: string | null
  action: string
  entity_type: string | null
  entity_id: string | null
  result: SystemLogResult | string
  /** May already be masked by the backend. */
  ip_address: string | null
  user_agent: string | null
  summary: string | null
  details: unknown
}

export interface SystemLogPage {
  success?: boolean
  data: SystemLogEntry[]
  next_cursor: string | null
  latest_cursor: string | null
}

export interface SystemLogFilters {
  action?: string
  actor_user_id?: string
  result?: SystemLogResult | ''
  /** RFC3339 */
  from?: string
  /** RFC3339 */
  to?: string
  q?: string
}

export interface SystemLogQuery extends SystemLogFilters {
  limit?: number
  cursor?: string
  since_cursor?: string
}

export type SystemLogFetcher = (query: SystemLogQuery) => Promise<SystemLogPage>

/** HTTP status of a $fetch / FetchError, or undefined. */
export const errorStatus = (err: unknown): number | undefined => {
  const e = err as { status?: number, statusCode?: number, response?: { status?: number } } | null | undefined
  return e?.status ?? e?.statusCode ?? e?.response?.status
}

/** Drop empty filter values so they are not sent as `?action=`. */
const cleanQuery = (query: SystemLogQuery): Record<string, string | number> => {
  const out: Record<string, string | number> = {}
  for (const [k, v] of Object.entries(query)) {
    if (v === undefined || v === null || v === '') continue
    out[k] = typeof v === 'string' ? v.trim() : v
  }
  return out
}

/** GET /api/v1/system-logs. The api-auth plugin attaches the token. */
export const fetchSystemLogs: SystemLogFetcher = async (query) => {
  const res = await $fetch<Partial<SystemLogPage> | null>(`${getAuthServiceBaseUrl()}/system-logs`, {
    method: 'GET',
    query: cleanQuery(query)
  })
  return {
    success: res?.success,
    data: Array.isArray(res?.data) ? res.data : [],
    next_cursor: res?.next_cursor ?? null,
    latest_cursor: res?.latest_cursor ?? null
  }
}

const byNewest = (a: SystemLogEntry, b: SystemLogEntry) =>
  (Date.parse(b.occurred_at) || 0) - (Date.parse(a.occurred_at) || 0)

export interface UseSystemLogsOptions {
  fetcher?: SystemLogFetcher
  /** Poll interval when healthy (ms). */
  intervalMs?: number
  /** Upper bound for the error back-off (ms). */
  maxBackoffMs?: number
  pageSize?: number
  /** Load the first page and start polling right away. */
  immediate?: boolean
}

/**
 * Loads the first page, polls every `intervalMs` with `since_cursor` and prepends new rows (deduped by id),
 * loads older pages with `cursor`. Polling pauses while the tab is hidden, when the user pauses it, after a 403,
 * and when the owning component/scope is disposed. Failed polls back off exponentially (interval * 2^n, capped).
 */
export function useSystemLogs(options: UseSystemLogsOptions = {}) {
  const fetcher = options.fetcher ?? fetchSystemLogs
  const intervalMs = options.intervalMs ?? 10_000
  const maxBackoffMs = options.maxBackoffMs ?? 120_000
  const pageSize = options.pageSize ?? 50

  const rows = shallowRef<SystemLogEntry[]>([])
  const filters = ref<SystemLogFilters>({})
  const nextCursor = ref<string | null>(null)
  const latestCursor = ref<string | null>(null)

  const loading = ref(false)
  const loadingMore = ref(false)
  const loaded = ref(false)
  /** First-page error (shown as the error state). */
  const error = ref<string | null>(null)
  /** Older-page error (shown next to "Load more"). */
  const loadMoreError = ref<string | null>(null)
  /** Last poll error; rows stay on screen while polling retries. */
  const pollError = ref<string | null>(null)
  const forbidden = ref(false)
  const lastUpdated = ref<Date | null>(null)
  const consecutiveFailures = ref(0)

  /** User toggle. */
  const paused = ref(false)
  const visible = ref(typeof document === 'undefined' || document.visibilityState !== 'hidden')
  let disposed = false
  let timer: ReturnType<typeof setTimeout> | null = null
  let polling = false
  // Bumped on every reload so responses for an older filter set are dropped
  let generation = 0

  const hasMore = computed(() => !!nextCursor.value)
  const isEmpty = computed(() => loaded.value && !error.value && !forbidden.value && rows.value.length === 0)
  /** True while the poll loop is allowed to run. */
  const live = computed(() => !paused.value && visible.value && !forbidden.value && !disposed && loaded.value && !error.value)

  const nextDelay = () =>
    consecutiveFailures.value === 0
      ? intervalMs
      : Math.min(intervalMs * 2 ** consecutiveFailures.value, maxBackoffMs)

  const clearTimer = () => {
    if (timer) clearTimeout(timer)
    timer = null
  }

  const schedule = () => {
    clearTimer()
    if (!live.value) return
    timer = setTimeout(() => {
      void poll()
    }, nextDelay())
  }

  const prepend = (incoming: SystemLogEntry[]) => {
    if (!incoming.length) return
    const seen = new Set(rows.value.map(r => r.id))
    const fresh: SystemLogEntry[] = []
    for (const row of incoming) {
      if (row?.id == null || seen.has(row.id)) continue
      seen.add(row.id)
      fresh.push(row)
    }
    if (fresh.length) rows.value = [...fresh.sort(byNewest), ...rows.value]
  }

  const append = (incoming: SystemLogEntry[]) => {
    const seen = new Set(rows.value.map(r => r.id))
    const older = incoming.filter(r => r?.id != null && !seen.has(r.id))
    if (older.length) rows.value = [...rows.value, ...older]
  }

  const handleForbidden = (err: unknown) => {
    if (errorStatus(err) !== 403) return false
    forbidden.value = true
    clearTimer()
    return true
  }

  /** Load (or reload after a filter change) the newest page. */
  const load = async () => {
    const gen = ++generation
    clearTimer()
    loading.value = true
    error.value = null
    loadMoreError.value = null
    pollError.value = null
    consecutiveFailures.value = 0
    try {
      const page = await fetcher({ ...filters.value, limit: pageSize })
      if (gen !== generation || disposed) return
      forbidden.value = false
      rows.value = page.data
      nextCursor.value = page.next_cursor
      latestCursor.value = page.latest_cursor
      lastUpdated.value = new Date()
    } catch (err) {
      if (gen !== generation || disposed) return
      if (!handleForbidden(err)) error.value = extractErrorMessage(err, 'Failed to load system logs')
      rows.value = []
      nextCursor.value = null
    } finally {
      if (gen === generation) {
        loading.value = false
        loaded.value = true
      }
    }
    if (gen === generation) schedule()
  }

  /**
   * Fetch events newer than latest_cursor and prepend them. Never polls without since_cursor: a cursor-less
   * request is a first-page load, which the backend records as a system_logs.viewed event.
   */
  const poll = async () => {
    if (polling || !live.value) return
    if (!latestCursor.value) {
      schedule()
      return
    }
    polling = true
    const gen = generation
    try {
      const page = await fetcher({ ...filters.value, limit: pageSize, since_cursor: latestCursor.value })
      if (gen !== generation || disposed) return
      prepend(page.data)
      if (page.latest_cursor) latestCursor.value = page.latest_cursor
      consecutiveFailures.value = 0
      pollError.value = null
      lastUpdated.value = new Date()
    } catch (err) {
      if (gen !== generation || disposed) return
      if (handleForbidden(err)) return
      consecutiveFailures.value++
      pollError.value = extractErrorMessage(err, 'Failed to refresh system logs')
    } finally {
      polling = false
    }
    if (gen === generation) schedule()
  }

  /** Load the next older page with `cursor`. */
  const loadMore = async () => {
    if (!nextCursor.value || loadingMore.value) return
    const gen = generation
    loadingMore.value = true
    loadMoreError.value = null
    try {
      const page = await fetcher({ ...filters.value, limit: pageSize, cursor: nextCursor.value })
      if (gen !== generation || disposed) return
      append(page.data)
      nextCursor.value = page.next_cursor
    } catch (err) {
      if (gen !== generation || disposed) return
      if (!handleForbidden(err)) loadMoreError.value = extractErrorMessage(err, 'Failed to load older logs')
    } finally {
      if (gen === generation) loadingMore.value = false
    }
  }

  const setFilters = (next: SystemLogFilters) => {
    filters.value = { ...next }
    return load()
  }

  const pause = () => {
    paused.value = true
    clearTimer()
  }

  /** Resume polling and fetch right away. */
  const resume = () => {
    paused.value = false
    consecutiveFailures.value = 0
    void poll()
  }

  const togglePaused = () => (paused.value ? resume() : pause())

  const onVisibilityChange = () => {
    visible.value = document.visibilityState !== 'hidden'
    if (!visible.value) clearTimer()
    else void poll()
  }

  const stop = () => {
    disposed = true
    clearTimer()
    if (typeof document !== 'undefined') document.removeEventListener('visibilitychange', onVisibilityChange)
  }

  if (typeof document !== 'undefined') document.addEventListener('visibilitychange', onVisibilityChange)
  if (getCurrentScope()) onScopeDispose(stop)
  if (options.immediate !== false) void load()

  return {
    rows,
    filters,
    nextCursor,
    latestCursor,
    loading,
    loadingMore,
    loaded,
    error,
    loadMoreError,
    pollError,
    forbidden,
    lastUpdated,
    consecutiveFailures,
    paused,
    visible,
    live,
    hasMore,
    isEmpty,
    load,
    poll,
    loadMore,
    setFilters,
    pause,
    resume,
    togglePaused,
    stop
  }
}

// One probe per session token, shared by every caller
const accessProbes = new Map<string, Promise<boolean>>()

/**
 * Asks auth-service whether the caller holds view_system_logs, without loading or logging anything.
 * A plain first-page request would record a system_logs.viewed event, so the probe sends an invalid
 * since_cursor instead: the controller checks the permission first (403 = no permission) and only then rejects
 * the cursor (400 = has permission); poll requests are never recorded. 2xx also means allowed.
 * Returns null when the answer is unknown (network error, 5xx) so the caller can retry later.
 * TODO(contract): replace with `permissions` on the login / auth/me user or a dedicated access endpoint.
 */
export const probeSystemLogsAccess = async (fetcher: SystemLogFetcher = fetchSystemLogs): Promise<boolean | null> => {
  try {
    await fetcher({ limit: 1, since_cursor: 'access-check' })
    return true
  } catch (err) {
    const status = errorStatus(err)
    if (status === 400) return true
    if (status === undefined || status >= 500) return null
    // 403 no permission, 404 endpoint not deployed, anything else: hide
    return false
  }
}

/**
 * Whether the System Logs tab is shown. Permission-based, never by role name:
 *  1. VIEWER-only users never see it.
 *  2. When the server sent a permission list (login or /auth/me `permissions`), it decides.
 *  3. Otherwise probeSystemLogsAccess once per session token (unknown results are not cached).
 * Hiding the tab is UI only; the backend enforces view_system_logs on every request.
 */
export function useSystemLogAccess() {
  const { isViewer, getUserPermissions } = useRbac()
  const authStore = useAuthStore()
  const probed = ref<boolean | null>(null)

  const runProbe = () => {
    const token = authStore.token
    if (!token || isViewer.value || getUserPermissions() !== null) return
    let probe = accessProbes.get(token)
    if (!probe) {
      probe = probeSystemLogsAccess().then((ok) => {
        if (ok === null) accessProbes.delete(token)
        return ok === true
      })
      accessProbes.set(token, probe)
    }
    void probe.then((ok) => {
      probed.value = ok
    })
  }

  const canViewSystemLogs = computed(() => {
    if (isViewer.value) return false
    const perms = getUserPermissions()
    if (perms !== null) return perms.includes(SYSTEM_LOGS_PERMISSION)
    return probed.value === true
  })

  runProbe()

  return { canViewSystemLogs, recheck: runProbe }
}
