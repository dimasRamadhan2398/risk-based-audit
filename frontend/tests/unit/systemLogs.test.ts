import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope } from 'vue'
import { probeSystemLogsAccess, useSystemLogs, type SystemLogEntry, type SystemLogPage, type SystemLogQuery } from '../../composables/useSystemLogs'
import { getSettingsTabs, resolveSettingsTab } from '../../utils/viewerPolicy'
import { actionKey, dateInputToRfc3339, detailEntries, isEmptyDetails } from '../../utils/systemLogFormat'

const entry = (id: string, minutesAgo = 0): SystemLogEntry => ({
  id,
  occurred_at: new Date(Date.UTC(2026, 10, 4, 12, 0, 0) - minutesAgo * 60_000).toISOString(),
  actor_user_id: 'u1',
  actor_username: 'admin',
  action: 'auth.login',
  entity_type: 'user',
  entity_id: 'u1',
  result: 'success',
  ip_address: '10.0.x.x',
  user_agent: 'test',
  summary: `event ${id}`,
  details: null
})

const page = (data: SystemLogEntry[], next: string | null = null, latest: string | null = null): SystemLogPage =>
  ({ success: true, data, next_cursor: next, latest_cursor: latest })

const httpError = (status: number) => Object.assign(new Error(`HTTP ${status}`), { status, data: { message: `HTTP ${status}` } })

let visibility: DocumentVisibilityState = 'visible'
const setVisibility = (v: DocumentVisibilityState) => {
  visibility = v
  document.dispatchEvent(new Event('visibilitychange'))
}

/** Run the composable inside a scope so onScopeDispose (component unmount) can be exercised. */
const mount = (fetcher: (q: SystemLogQuery) => Promise<SystemLogPage>, opts: Record<string, unknown> = {}) => {
  const scope = effectScope()
  const api = scope.run(() => useSystemLogs({ fetcher, intervalMs: 10_000, maxBackoffMs: 60_000, ...opts }))!
  return { api, scope }
}

/** Let pending promise callbacks run without advancing timers. */
const flush = async () => {
  for (let i = 0; i < 5; i++) await Promise.resolve()
}

beforeEach(() => {
  vi.useFakeTimers()
  visibility = 'visible'
  Object.defineProperty(document, 'visibilityState', { configurable: true, get: () => visibility })
})

afterEach(() => {
  vi.useRealTimers()
})

describe('useSystemLogs: first page and polling', () => {
  it('loads the first page, then polls every 10s with since_cursor and prepends new rows', async () => {
    const fetcher = vi.fn()
      .mockResolvedValueOnce(page([entry('b', 1), entry('a', 2)], 'older-1', 'latest-1'))
      .mockResolvedValueOnce(page([entry('c', 0)], null, 'latest-2'))
      .mockResolvedValue(page([], null, null))
    const { api, scope } = mount(fetcher)
    await flush()

    expect(fetcher).toHaveBeenNthCalledWith(1, { limit: 50 })
    expect(api.rows.value.map(r => r.id)).toEqual(['b', 'a'])
    expect(api.live.value).toBe(true)

    await vi.advanceTimersByTimeAsync(9_999)
    expect(fetcher).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(1)
    expect(fetcher).toHaveBeenNthCalledWith(2, { limit: 50, since_cursor: 'latest-1' })
    expect(api.rows.value.map(r => r.id)).toEqual(['c', 'b', 'a'])
    expect(api.latestCursor.value).toBe('latest-2')

    // An empty poll keeps the previous latest cursor
    await vi.advanceTimersByTimeAsync(10_000)
    expect(fetcher).toHaveBeenNthCalledWith(3, { limit: 50, since_cursor: 'latest-2' })
    expect(api.latestCursor.value).toBe('latest-2')
    scope.stop()
  })

  it('dedupes rows by id and puts the newest first when a poll overlaps', async () => {
    const fetcher = vi.fn()
      .mockResolvedValueOnce(page([entry('b', 1), entry('a', 2)], null, 'l1'))
      // Overlapping and out of order: 'b' again, 'd' newer than 'c'
      .mockResolvedValueOnce(page([entry('c', 0.5), entry('b', 1), entry('d', 0)], null, 'l2'))
    const { api, scope } = mount(fetcher)
    await flush()
    await vi.advanceTimersByTimeAsync(10_000)

    expect(api.rows.value.map(r => r.id)).toEqual(['d', 'c', 'b', 'a'])
    expect(new Set(api.rows.value.map(r => r.id)).size).toBe(api.rows.value.length)
    scope.stop()
  })

  it('passes filters to every request and reloads from the first page when they change', async () => {
    const fetcher = vi.fn().mockResolvedValue(page([entry('a')], null, 'l1'))
    const { api, scope } = mount(fetcher)
    await flush()

    await api.setFilters({ action: 'auth.login', result: 'failure', q: 'bob' })
    expect(fetcher).toHaveBeenLastCalledWith({ action: 'auth.login', result: 'failure', q: 'bob', limit: 50 })

    await vi.advanceTimersByTimeAsync(10_000)
    expect(fetcher).toHaveBeenLastCalledWith({ action: 'auth.login', result: 'failure', q: 'bob', limit: 50, since_cursor: 'l1' })
    scope.stop()
  })

  it('drops a response for an older filter set', async () => {
    let resolveFirst!: (p: SystemLogPage) => void
    const fetcher = vi.fn()
      .mockImplementationOnce(() => new Promise<SystemLogPage>((r) => { resolveFirst = r }))
      .mockResolvedValueOnce(page([entry('filtered')], null, 'l2'))
    const { api, scope } = mount(fetcher)

    await api.setFilters({ result: 'failure' })
    resolveFirst(page([entry('stale')], 'x', 'l1'))
    await flush()

    expect(api.rows.value.map(r => r.id)).toEqual(['filtered'])
    expect(api.latestCursor.value).toBe('l2')
    scope.stop()
  })
})

describe('useSystemLogs: pausing', () => {
  it('stops polling while the browser tab is hidden and polls right away when visible again', async () => {
    const fetcher = vi.fn()
      .mockResolvedValueOnce(page([entry('a')], null, 'l1'))
      .mockResolvedValue(page([entry('b')], null, 'l2'))
    const { api, scope } = mount(fetcher)
    await flush()

    setVisibility('hidden')
    expect(api.live.value).toBe(false)
    await vi.advanceTimersByTimeAsync(60_000)
    expect(fetcher).toHaveBeenCalledTimes(1)

    setVisibility('visible')
    await flush()
    expect(fetcher).toHaveBeenCalledTimes(2)
    expect(fetcher).toHaveBeenLastCalledWith({ limit: 50, since_cursor: 'l1' })
    expect(api.rows.value.map(r => r.id)).toEqual(['b', 'a'])

    // Back on the normal schedule
    await vi.advanceTimersByTimeAsync(10_000)
    expect(fetcher).toHaveBeenCalledTimes(3)
    scope.stop()
  })

  it('pause/resume toggle stops polling and resumes with an immediate fetch', async () => {
    const fetcher = vi.fn().mockResolvedValue(page([], null, 'l1'))
    const { api, scope } = mount(fetcher)
    await flush()

    api.togglePaused()
    expect(api.paused.value).toBe(true)
    await vi.advanceTimersByTimeAsync(60_000)
    expect(fetcher).toHaveBeenCalledTimes(1)

    api.togglePaused()
    await flush()
    expect(api.paused.value).toBe(false)
    expect(fetcher).toHaveBeenCalledTimes(2)
    scope.stop()
  })

  it('stops polling and removes the visibility listener when the scope is disposed (unmount)', async () => {
    const removeSpy = vi.spyOn(document, 'removeEventListener')
    const fetcher = vi.fn().mockResolvedValue(page([], null, 'l1'))
    const { scope } = mount(fetcher)
    await flush()

    scope.stop()
    expect(removeSpy).toHaveBeenCalledWith('visibilitychange', expect.any(Function))
    await vi.advanceTimersByTimeAsync(60_000)
    setVisibility('visible')
    await flush()
    expect(fetcher).toHaveBeenCalledTimes(1)
    removeSpy.mockRestore()
  })
})

describe('useSystemLogs: load more', () => {
  it('appends older pages with cursor, dedupes, and stops when next_cursor is null', async () => {
    const fetcher = vi.fn()
      .mockResolvedValueOnce(page([entry('c', 1), entry('b', 2)], 'cur-1', 'l1'))
      .mockResolvedValueOnce(page([entry('b', 2), entry('a', 3)], null, 'l1'))
    const { api, scope } = mount(fetcher)
    await flush()
    expect(api.hasMore.value).toBe(true)

    await api.loadMore()
    expect(fetcher).toHaveBeenLastCalledWith({ limit: 50, cursor: 'cur-1' })
    expect(api.rows.value.map(r => r.id)).toEqual(['c', 'b', 'a'])
    expect(api.hasMore.value).toBe(false)

    await api.loadMore()
    expect(fetcher).toHaveBeenCalledTimes(2)
    scope.stop()
  })

  it('keeps the rows and the cursor when loading an older page fails', async () => {
    const fetcher = vi.fn()
      .mockResolvedValueOnce(page([entry('b')], 'cur-1', 'l1'))
      .mockRejectedValueOnce(httpError(500))
    const { api, scope } = mount(fetcher)
    await flush()

    await api.loadMore()
    expect(api.loadMoreError.value).toBeTruthy()
    expect(api.rows.value.map(r => r.id)).toEqual(['b'])
    expect(api.nextCursor.value).toBe('cur-1')
    scope.stop()
  })
})

describe('useSystemLogs: errors', () => {
  it('backs off exponentially on poll errors (capped) and resets after a success', async () => {
    const fetcher = vi.fn()
      .mockResolvedValueOnce(page([entry('a')], null, 'l1'))
      .mockRejectedValueOnce(httpError(502)) // at +10s
      .mockRejectedValueOnce(httpError(502)) // +20s later
      .mockRejectedValueOnce(httpError(502)) // +40s later
      .mockRejectedValueOnce(httpError(502)) // +60s later (cap)
      .mockResolvedValue(page([entry('b')], null, 'l2')) // +60s later (cap)
    const { api, scope } = mount(fetcher)
    await flush()

    await vi.advanceTimersByTimeAsync(10_000)
    expect(fetcher).toHaveBeenCalledTimes(2)
    expect(api.consecutiveFailures.value).toBe(1)
    expect(api.pollError.value).toBeTruthy()
    expect(api.rows.value.map(r => r.id)).toEqual(['a'])

    await vi.advanceTimersByTimeAsync(19_999)
    expect(fetcher).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(1)
    expect(fetcher).toHaveBeenCalledTimes(3)

    await vi.advanceTimersByTimeAsync(40_000)
    expect(fetcher).toHaveBeenCalledTimes(4)

    await vi.advanceTimersByTimeAsync(59_999)
    expect(fetcher).toHaveBeenCalledTimes(4)
    await vi.advanceTimersByTimeAsync(1)
    expect(fetcher).toHaveBeenCalledTimes(5)

    await vi.advanceTimersByTimeAsync(60_000)
    expect(fetcher).toHaveBeenCalledTimes(6)
    expect(api.consecutiveFailures.value).toBe(0)
    expect(api.pollError.value).toBeNull()
    expect(api.rows.value.map(r => r.id)).toEqual(['b', 'a'])

    // Healthy again: back to the 10s interval
    await vi.advanceTimersByTimeAsync(10_000)
    expect(fetcher).toHaveBeenCalledTimes(7)
    scope.stop()
  })

  it('shows the error state when the first page fails, without polling', async () => {
    const fetcher = vi.fn().mockRejectedValueOnce(httpError(500)).mockResolvedValue(page([entry('a')], null, 'l1'))
    const { api, scope } = mount(fetcher)
    await flush()

    expect(api.error.value).toBeTruthy()
    expect(api.live.value).toBe(false)
    await vi.advanceTimersByTimeAsync(60_000)
    expect(fetcher).toHaveBeenCalledTimes(1)

    await api.load()
    expect(api.error.value).toBeNull()
    expect(api.rows.value.map(r => r.id)).toEqual(['a'])
    scope.stop()
  })

  it('a 403 sets forbidden and stops polling for good', async () => {
    const fetcher = vi.fn()
      .mockResolvedValueOnce(page([entry('a')], null, 'l1'))
      .mockRejectedValueOnce(httpError(403))
    const { api, scope } = mount(fetcher)
    await flush()
    await vi.advanceTimersByTimeAsync(10_000)

    expect(api.forbidden.value).toBe(true)
    expect(api.live.value).toBe(false)
    await vi.advanceTimersByTimeAsync(120_000)
    expect(fetcher).toHaveBeenCalledTimes(2)
    scope.stop()
  })
})

describe('useSystemLogs: never re-records a view', () => {
  it('does not poll without since_cursor (a cursor-less request is a recorded first-page view)', async () => {
    const fetcher = vi.fn().mockResolvedValue(page([], null, null))
    const { scope } = mount(fetcher)
    await flush()
    await vi.advanceTimersByTimeAsync(60_000)

    expect(fetcher).toHaveBeenCalledTimes(1)
    for (const [q] of fetcher.mock.calls.slice(1)) expect(q.since_cursor).toBeTruthy()
    scope.stop()
  })
})

describe('probeSystemLogsAccess', () => {
  it('maps the probe status to the permission without a first-page load', async () => {
    const calls: SystemLogQuery[] = []
    const failWith = (status: number) => async (q: SystemLogQuery) => {
      calls.push(q)
      throw httpError(status)
    }
    expect(await probeSystemLogsAccess(failWith(403))).toBe(false)
    expect(await probeSystemLogsAccess(failWith(404))).toBe(false)
    // Permission passed; only the deliberately invalid cursor was rejected
    expect(await probeSystemLogsAccess(failWith(400))).toBe(true)
    expect(await probeSystemLogsAccess(failWith(503))).toBeNull()
    expect(await probeSystemLogsAccess(async () => page([]))).toBe(true)
    expect(await probeSystemLogsAccess(async () => {
      throw new Error('network')
    })).toBeNull()

    // Always a poll request, never a cursor-less (recorded) one
    for (const q of calls) expect(q.since_cursor).toBeTruthy()
  })
})

describe('System Logs tab gate', () => {
  it('is never offered to VIEWER, even with the permission', () => {
    expect(getSettingsTabs(true, { systemLogs: true })).not.toContain('systemlogs')
    expect(resolveSettingsTab('systemlogs', true, { systemLogs: true })).toBe('profile')
  })

  it('is added before FAQ only when the permission check passes', () => {
    expect(getSettingsTabs(false)).not.toContain('systemlogs')
    const tabs = getSettingsTabs(false, { systemLogs: true })
    expect(tabs.slice(-2)).toEqual(['systemlogs', 'faq'])
    expect(resolveSettingsTab('systemlogs', false)).toBe('profile')
    expect(resolveSettingsTab('systemlogs', false, { systemLogs: true })).toBe('systemlogs')
  })
})

describe('systemLogFormat', () => {
  it('turns action codes into i18n-safe keys', () => {
    expect(actionKey('auth.login')).toBe('auth_login')
    expect(actionKey('Auth.Login-Failed')).toBe('auth_login_failed')
  })

  it('converts date inputs to RFC3339 day bounds', () => {
    expect(dateInputToRfc3339('', 'start')).toBeUndefined()
    expect(dateInputToRfc3339('2026-11-04', 'start')).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/)
    const start = Date.parse(dateInputToRfc3339('2026-11-04', 'start')!)
    const end = Date.parse(dateInputToRfc3339('2026-11-04', 'end')!)
    expect(end - start).toBe(86_399_000)
  })

  it('reads details objects and JSON strings', () => {
    expect(detailEntries({ a: 1, b: { c: 2 }, d: null })).toEqual([['a', '1'], ['b', '{"c":2}'], ['d', '-']])
    expect(detailEntries('{"x":"y"}')).toEqual([['x', 'y']])
    expect(detailEntries([1, 2])).toBeNull()
    expect(isEmptyDetails(null)).toBe(true)
    expect(isEmptyDetails({})).toBe(true)
    expect(isEmptyDetails({ a: 1 })).toBe(false)
  })
})
