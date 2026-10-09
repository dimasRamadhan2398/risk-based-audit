/**
 * Audit Charter (/audit-charter): GET /audit-charters pages its results (page_size defaults
 * to 10, max 100). The store used to call it without page/page_size, so only the 10 newest
 * charters loaded and the history table never had a second page. The store now loads every
 * page; the history table pages client-side and keeps its page across refetches.
 */
import { describe, it, expect, beforeEach, afterEach, vi, type Mock } from 'vitest'
import { createApp, defineComponent, h, nextTick, type App } from 'vue'
import { setActivePinia, createPinia } from 'pinia'
import {
  useCharterStore,
  extractCharterTotalPages,
  CHARTER_FETCH_PAGE_SIZE,
  CHARTER_HISTORY_PAGE_SIZE
} from '~/stores/charter'
import AuditCharterCard from '~/components/audit-charter/AuditCharterCard.vue'

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({ showSuccess: vi.fn(), showError: vi.fn(), showWarning: vi.fn(), showInfo: vi.fn() })
}))
vi.mock('~/composables/useRbac', () => ({ useRbac: () => ({ canManageCharter: { value: true } }) }))
vi.mock('~/stores/global-modal', () => ({ useGlobalModalStore: () => ({ confirmDelete: vi.fn(async () => true) }) }))

type FetchOpts = { method?: string, query?: Record<string, string | number>, cache?: string, body?: unknown }
type FetchFn = (url: string, opts?: FetchOpts) => Promise<unknown>
const g = globalThis as unknown as { $fetch: Mock<FetchFn>, [name: string]: unknown }

const flush = () => new Promise(r => setTimeout(r, 0))

type Row = { id: string, title: string, version: string, is_active: boolean, created_at: string }

/** n charters v1.0, v1.1, ... created in that order; the newest one is active. */
const charterRows = (n: number): Row[] =>
  Array.from({ length: n }, (_, i) => ({
    id: `c${i + 1}`,
    title: `Charter ${i + 1}`,
    version: (1 + i * 0.1).toFixed(1),
    is_active: i === n - 1,
    created_at: new Date(Date.UTC(2024, 0, 1 + i)).toISOString()
  }))

/**
 * Fake audit-service with the contract of controllers/audit_charter ListCharters:
 * page defaults to 1, page_size to 10 (capped at 100), newest first, and
 * { success, data: { charters, pagination: { page, page_size, total_count, total_pages } } }.
 * PUT and DELETE change the rows the next GET returns.
 */
const backend = (db: { rows: Row[] }) =>
  vi.fn<FetchFn>(async (url, opts = {}) => {
    const method = (opts.method ?? 'GET').toUpperCase()
    const id = url.match(/audit-charters\/([^/?]+)/)?.[1]
    if (method === 'DELETE' && id) {
      db.rows = db.rows.filter(r => r.id !== id)
      return { success: true }
    }
    if (method === 'PUT' && id) return { success: true }
    if (method !== 'GET') throw new Error(`unexpected ${method} ${url}`)

    const page = Math.max(1, Number(opts.query?.page) || 1)
    let pageSize = Number(opts.query?.page_size) || 10
    if (pageSize > 100) pageSize = 100
    const sorted = [...db.rows].sort((a, b) => b.created_at.localeCompare(a.created_at))
    const total = sorted.length
    return {
      success: true,
      message: 'Audit charters retrieved successfully',
      data: {
        charters: sorted.slice((page - 1) * pageSize, page * pageSize),
        pagination: { page, page_size: pageSize, total_count: total, total_pages: Math.ceil(total / pageSize) }
      }
    }
  })

const gets = () => g.$fetch.mock.calls.filter(([, o]) => (o?.method ?? 'GET') === 'GET')

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  g.getAuditServiceBaseUrl = () => '/api/v1'
  g.useAuthStore = () => ({ token: 'token' })
})

describe('charter store: loads every charter', () => {
  it('sends page and page_size, uncached, and makes one request when everything fits', async () => {
    g.$fetch = backend({ rows: charterRows(23) })
    const store = useCharterStore()
    await store.fetchCharters()

    expect(gets()).toHaveLength(1)
    const [url, opts] = gets()[0]!
    expect(url).toBe('/api/v1/audit-charters')
    expect(opts?.query).toEqual({ page: 1, page_size: CHARTER_FETCH_PAGE_SIZE })
    expect(opts?.cache).toBe('no-store')
    // Was 10 before: the backend's default page size.
    expect(store.charters).toHaveLength(23)
  })

  it('follows total_pages and fetches the remaining pages', async () => {
    g.$fetch = backend({ rows: charterRows(230) })
    const store = useCharterStore()
    await store.fetchCharters()

    expect(gets().map(([, o]) => o?.query?.page).sort()).toEqual([1, 2, 3])
    expect(store.charters).toHaveLength(230)
    expect(new Set(store.charters.map(c => c.id)).size).toBe(230)
    // Newest first across pages.
    expect(store.charters[0]?.version).toBe('23.9')
    expect(store.charters.at(-1)?.version).toBe('1.0')
  })

  it('splits the active charter from the history and bases nextVersion on the newest', async () => {
    g.$fetch = backend({ rows: charterRows(23) })
    const store = useCharterStore()
    await store.fetchCharters()

    expect(store.activeCharter?.version).toBe('3.2')
    expect(store.historyCharters).toHaveLength(22)
    expect(store.historyCharters.some(c => c.isActive)).toBe(false)
    expect(store.historyPageCount).toBe(3)
    expect(store.nextVersion).toBe('3.3')
  })

  it('a failed load shows the error and leaves no rows', async () => {
    g.$fetch = vi.fn<FetchFn>(async () => {
      throw Object.assign(new Error('[500]'), { statusCode: 500, data: { success: false, error: { message: 'database unavailable' } } })
    })
    const store = useCharterStore()
    await store.fetchCharters()

    expect(store.errorMsg).toBe('database unavailable')
    expect(store.charters).toEqual([])
    expect(store.loading).toBe(false)
  })
})

describe('extractCharterTotalPages', () => {
  it('reads total_pages from data.pagination', () => {
    expect(extractCharterTotalPages({ data: { charters: [], pagination: { total_pages: 4 } } })).toBe(4)
  })

  it('computes the page count from total_count and page_size when total_pages is missing', () => {
    expect(extractCharterTotalPages({ data: { pagination: { total_count: 201, page_size: 100 } } })).toBe(3)
  })

  it('falls back to one page without pagination meta', () => {
    expect(extractCharterTotalPages({ data: { charters: [] } })).toBe(1)
    expect(extractCharterTotalPages([])).toBe(1)
    expect(extractCharterTotalPages({ data: { pagination: { total_count: 0, total_pages: 0 } } })).toBe(1)
  })
})

describe('charter store: history page across refetches', () => {
  it('an edit keeps the page; a delete that empties the last page moves back one page', async () => {
    const db = { rows: charterRows(22) } // 21 history rows: pages of 10, 10, 1
    g.$fetch = backend(db)
    const store = useCharterStore()
    await store.fetchCharters()
    expect(store.historyPageCount).toBe(3)

    store.historyPage = 3
    await store.updateCharter('c1', { ...store.form, title: 'Renamed' })
    expect(store.historyPage).toBe(3)

    await store.deleteCharter('c1')
    await nextTick()
    expect(store.historyCharters).toHaveLength(20)
    expect(store.historyPage).toBe(2)
  })

  it('a create goes back to page 1, where the replaced charter now is', async () => {
    const db = { rows: charterRows(22) }
    const base = backend(db)
    g.$fetch = vi.fn<FetchFn>(async (url, opts = {}) => {
      if ((opts.method ?? 'GET').toUpperCase() === 'POST') return { success: true }
      return base(url, opts)
    })
    const store = useCharterStore()
    await store.fetchCharters()
    store.historyPage = 3

    await store.addCharter({ ...store.form, title: 'New', file: null })
    expect(store.historyPage).toBe(1)
  })
})

// --- AuditCharterCard with the real TableEntities; Nuxt UI primitives are stubbed ---

const stub = (name: string) => defineComponent({
  name,
  inheritAttrs: false,
  setup: (_p, { slots }) => () => h('div', { 'data-stub': name }, slots.default?.())
})

/** Renders the rows it gets so the visible page can be checked. */
const TableStub = defineComponent({
  props: { data: { type: Array, default: () => [] } },
  setup: p => () => h('div', { 'data-stub': 'UTable', 'data-versions': JSON.stringify((p.data as Row[]).map(r => r.version)) })
})

/** Same disabled rule as UPagination's next control: the last page has no next. */
const PaginationStub = defineComponent({
  props: { page: Number, itemsPerPage: Number, total: Number },
  emits: ['update:page'],
  setup: (p, { emit }) => () => {
    const pages = Math.max(1, Math.ceil((p.total ?? 0) / (p.itemsPerPage ?? 10)))
    const page = p.page ?? 1
    return h('div', { 'data-stub': 'UPagination', 'data-total': p.total, 'data-items-per-page': p.itemsPerPage, 'data-page': page }, [
      h('button', { 'data-next': '', 'disabled': page >= pages, 'onClick': () => emit('update:page', page + 1) })
    ])
  }
})

describe('AuditCharterCard: history table pagination', () => {
  let app: App | undefined
  let container: HTMLElement

  afterEach(() => {
    app?.unmount()
    container?.remove()
  })

  const mountCard = async (rows: Row[]) => {
    g.$fetch = backend({ rows })
    container = document.createElement('div')
    document.body.appendChild(container)
    app = createApp({ render: () => h(AuditCharterCard) })
    for (const name of ['UAlert', 'USkeleton', 'UIcon', 'UButton', 'UCard', 'UBadge', 'UTooltip', 'TeamMembersBadge']) {
      app.component(name, stub(name))
    }
    app.component('UTable', TableStub)
    app.component('UPagination', PaginationStub)
    const pinia = createPinia()
    setActivePinia(pinia)
    app.use(pinia)
    app.mount(container)
    await flush()
    await nextTick()
    return useCharterStore()
  }

  const pagination = () => container.querySelector('[data-stub="UPagination"]') as HTMLElement | null
  const nextButton = () => container.querySelector('[data-next]') as HTMLButtonElement
  const visibleVersions = () => JSON.parse((container.querySelector('[data-stub="UTable"]') as HTMLElement).dataset.versions!)

  it('enables next when the history is longer than one page and pages through it', async () => {
    const store = await mountCard(charterRows(23))

    expect(pagination()?.dataset.total).toBe('22')
    expect(pagination()?.dataset.itemsPerPage).toBe(String(CHARTER_HISTORY_PAGE_SIZE))
    expect(nextButton().disabled).toBe(false)
    expect(visibleVersions()).toHaveLength(10)
    expect(visibleVersions()[0]).toBe('3.1')

    nextButton().click()
    await nextTick()
    expect(store.historyPage).toBe(2)
    expect(visibleVersions()).toHaveLength(10)

    nextButton().click()
    await nextTick()
    expect(store.historyPage).toBe(3)
    expect(visibleVersions()).toEqual(['1.1', '1.0'])
    expect(nextButton().disabled).toBe(true)
  })

  it('keeps next disabled when the history fits on one page', async () => {
    await mountCard(charterRows(10))

    expect(pagination()?.dataset.total).toBe('9')
    expect(nextButton().disabled).toBe(true)
  })

  it('stays on the same page after a refetch remounts the table', async () => {
    const store = await mountCard(charterRows(23))
    nextButton().click()
    await nextTick()
    expect(store.historyPage).toBe(2)

    // The card swaps the table for a skeleton while loading.
    await store.fetchCharters()
    await nextTick()
    expect(pagination()?.dataset.page).toBe('2')
    expect(visibleVersions()[0]).toBe('2.1')
  })
})
