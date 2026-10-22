import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'

import { useAuditExecutionStore } from '~/stores/audit-execution'
import { createAutoSyncOnce } from '~/utils/auditExecutionAutoSync'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: { auditServiceBaseUrl: 'http://localhost:8002/api/v1' }
  })
}))

const LOCALES = ['en', 'id'].map(lang =>
  JSON.parse(readFileSync(resolve(__dirname, `../../locales/${lang}/common.json`), 'utf-8'))
)

const listPayload = (ids: string[]) => ({
  success: true,
  data: {
    items: ids.map((id, i) => ({ id, ref: `ST-00${i + 1}/SKAI/2026`, name: `Audit ${id}`, progress: 10, status: 'planning' }))
  }
})

const syncPayload = (synced: number, created: number, unmatched: Record<string, number> = {}) => ({
  success: true,
  data: { synced, created, items: [], unmatched_source_rows: unmatched }
})

/** Routes $fetch by method + path and records every call. */
const mockApi = (handlers: { sync: () => unknown, list: () => unknown }) => {
  const calls: string[] = []
  const fetchMock = vi.fn(async (url: string, opts?: { method?: string }) => {
    const method = opts?.method || 'GET'
    calls.push(`${method} ${url}`)
    if (method === 'POST' && /\/audit-executions\/sync$/.test(url)) return handlers.sync()
    if (method === 'GET' && /\/audit-executions$/.test(url)) return handlers.list()
    throw new Error(`unexpected ${method} ${url}`)
  })
  global.$fetch = fetchMock as unknown as typeof global.$fetch
  return calls
}

describe('audit execution store — syncFromAssignmentLetters', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  it('POSTs the sync, then refetches the list', async () => {
    const calls = mockApi({ sync: () => syncPayload(2, 0), list: () => listPayload(['a', 'b']) })

    const store = useAuditExecutionStore()
    const result = await store.syncFromAssignmentLetters()

    expect(calls).toHaveLength(2)
    expect(calls[0]).toMatch(/^POST \S*\/audit-executions\/sync$/)
    expect(calls[1]).toMatch(/^GET \S*\/audit-executions$/)
    expect(result).toEqual({ synced: 2, created: 0, unmatched: 0 })
    expect(store.auditExecutions.map(e => e.id)).toEqual(['a', 'b'])
    expect(store.syncing).toBe(false)
    expect(store.syncError).toBeNull()
  })

  it('reports how many executions were created from assignment letters', async () => {
    mockApi({ sync: () => syncPayload(7, 3), list: () => listPayload(['a', 'b', 'c']) })

    const store = useAuditExecutionStore()
    const result = await store.syncFromAssignmentLetters()

    expect(result?.created).toBe(3)
    expect(result?.synced).toBe(7)
    expect(store.auditExecutions).toHaveLength(3)
  })

  it('sums unmatched source rows and tolerates a response without created', async () => {
    mockApi({
      sync: () => ({ data: { synced: 4, unmatched_source_rows: { fieldwork_samples: 2, working_paper_headers: 1 } } }),
      list: () => listPayload(['a'])
    })

    const store = useAuditExecutionStore()
    const result = await store.syncFromAssignmentLetters()

    expect(result).toEqual({ synced: 4, created: 0, unmatched: 3 })
  })

  it('keeps the current list and sets syncError when the sync fails', async () => {
    let failSync = false
    const calls = mockApi({
      sync: () => {
        if (failSync) throw { data: { success: false, error: { code: 'SYNC_FAILED', message: 'Failed to sync audit executions' } } }
        return syncPayload(2, 0)
      },
      list: () => listPayload(['a', 'b'])
    })

    const store = useAuditExecutionStore()
    await store.fetchAuditExecutions()
    expect(store.auditExecutions).toHaveLength(2)

    failSync = true
    calls.length = 0
    const result = await store.syncFromAssignmentLetters()

    expect(result).toBeNull()
    expect(store.syncError).toBe('Failed to sync audit executions')
    expect(store.auditExecutions.map(e => e.id)).toEqual(['a', 'b'])
    // The list error belongs to the fetch; a failed sync must not mark the list as broken.
    expect(store.error).toBeNull()
    expect(store.syncing).toBe(false)
    // No refetch after a failed sync.
    expect(calls).toEqual([expect.stringMatching(/^POST /)])
  })

  it('ignores a second call while a sync is still running', async () => {
    let release!: () => void
    const gate = new Promise<void>((r) => {
      release = r
    })
    const calls = mockApi({
      sync: async () => {
        await gate
        return syncPayload(1, 1)
      },
      list: () => listPayload(['a'])
    })

    const store = useAuditExecutionStore()
    const first = store.syncFromAssignmentLetters()
    expect(store.syncing).toBe(true)
    const second = await store.syncFromAssignmentLetters()
    expect(second).toBeNull()
    expect(store.syncError).toBeNull()

    release()
    expect(await first).toMatchObject({ created: 1 })
    expect(calls.filter(c => c.startsWith('POST'))).toHaveLength(1)
  })
})

describe('createAutoSyncOnce — auto-sync guard', () => {
  const emptyOk = { canSync: true, hasError: false, count: 0 }

  it('runs once when the list is empty and the user may sync', async () => {
    const run = vi.fn(async () => ({ synced: 0, created: 2, unmatched: 0 }))
    const guard = createAutoSyncOnce(run)

    expect(await guard.maybeRun(emptyOk)).toMatchObject({ created: 2 })
    expect(guard.attempted).toBe(true)
    expect(run).toHaveBeenCalledTimes(1)
  })

  it('never runs a second time, even if the list is still empty (no letters yet)', async () => {
    const run = vi.fn(async () => ({ synced: 0, created: 0, unmatched: 0 }))
    const guard = createAutoSyncOnce(run)

    await guard.maybeRun(emptyOk)
    await guard.maybeRun(emptyOk)
    await guard.maybeRun(emptyOk)

    expect(run).toHaveBeenCalledTimes(1)
  })

  it('does not retry after a failed sync', async () => {
    const run = vi.fn(async () => null)
    const guard = createAutoSyncOnce(run)

    expect(await guard.maybeRun(emptyOk)).toBeNull()
    expect(await guard.maybeRun(emptyOk)).toBeUndefined()
    expect(run).toHaveBeenCalledTimes(1)
  })

  it('skips without permission, on a load error or when the list has rows', async () => {
    const run = vi.fn(async () => null)
    const guard = createAutoSyncOnce(run)

    await guard.maybeRun({ ...emptyOk, canSync: false })
    await guard.maybeRun({ ...emptyOk, hasError: true })
    await guard.maybeRun({ ...emptyOk, count: 3 })

    expect(run).not.toHaveBeenCalled()
    expect(guard.attempted).toBe(false)
  })

  it('still runs once after a retry fixes an initial load error', async () => {
    const run = vi.fn(async () => null)
    const guard = createAutoSyncOnce(run)

    await guard.maybeRun({ ...emptyOk, hasError: true })
    await guard.maybeRun(emptyOk)
    await guard.maybeRun(emptyOk)

    expect(run).toHaveBeenCalledTimes(1)
  })

  it('gives each page visit its own guard', async () => {
    const run = vi.fn(async () => null)
    await createAutoSyncOnce(run).maybeRun(emptyOk)
    await createAutoSyncOnce(run).maybeRun(emptyOk)

    expect(run).toHaveBeenCalledTimes(2)
  })
})

describe('audit execution sync — locales', () => {
  it('has the sync, empty state and load error keys in both en and id', () => {
    for (const locale of LOCALES) {
      const ae = locale.auditExecution
      for (const key of ['button', 'tooltip', 'successTitle', 'successCreated', 'successNoneCreated', 'autoCreatedTitle', 'autoCreatedDescription', 'unmatchedTitle', 'unmatchedWarning', 'failedTitle']) {
        expect(ae.sync[key], `sync.${key}`).toBeTruthy()
      }
      expect(ae.sync.successCreated).toContain('{created}')
      expect(ae.sync.successCreated).toContain('{synced}')
      expect(ae.sync.successNoneCreated).toContain('{synced}')
      expect(ae.sync.autoCreatedDescription).toContain('{created}')
      expect(ae.sync.unmatchedWarning).toContain('{count}')
      for (const key of ['title', 'description', 'descriptionReadOnly', 'goToAssignmentLetter', 'syncing']) {
        expect(ae.emptyState[key], `emptyState.${key}`).toBeTruthy()
      }
      expect(ae.loadError.title).toBeTruthy()
      expect(ae.loadError.retry).toBeTruthy()
    }
    expect(LOCALES[1].auditExecution.sync.button).toBe('Sinkronkan dari Surat Tugas')
  })
})
