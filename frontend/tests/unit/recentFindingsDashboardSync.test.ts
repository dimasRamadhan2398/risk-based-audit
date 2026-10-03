// @ts-nocheck
/**
 * Dashboard "Recent Finding Issues" card vs Audit Result Report (ARR),
 * Audit Fieldwork and Digital Working Paper (KKA).
 *
 * The card (pages/dashboard/index.vue recentFindingsData) renders
 * auditResultStore.recentFindings, loaded on dashboard mount by
 * fetchRecentFindings(5) from GET /audit-result-reports/recent-findings?limit=5.
 * The backend merges saved ARR findings with live KKA / fieldwork findings,
 * dedupes, normalises the category and sorts newest first:
 *   { success, message, data: { items: [{ title, category, action, source,
 *     assignmentLetterId, reportId: string|null, date: RFC3339 }], total } }
 *
 * mockApi below fakes that endpoint from the same reports / auto-findings the
 * other ARR endpoints return, so the card and the ARR page see one data set.
 * No mock data may reach the card: before the fetch, on an empty response and
 * on an error it is empty.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuditResultReportStore } from '~/stores/audit-result-report'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: { auditServiceBaseUrl: 'http://localhost:8002/api/v1' }
  })
}))

vi.mock('~/composables/useAppToast', () => ({
  useAppToast: () => ({ success: vi.fn(), error: vi.fn(), info: vi.fn(), warning: vi.fn() })
}))

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn()
  })
}))

const finding = (title: string, category = 'Significant') => ({ title, category, action: 'Follow up' })

const report = (id: string, letter: string, date: string, findings: any[], extra: any = {}) => ({
  id,
  assignmentLetterId: letter,
  reportTitle: `Laporan Hasil Audit ${letter}`,
  report_date: `${date}T00:00:00Z`,
  created_at: `${date}T08:00:00Z`,
  status: 'DRAFT',
  findingsCount: findings.length,
  findings,
  ...extra
})

// Backend shape (crud.List): { success, data: { items, pagination } }.
const listResponse = (items: any[]) => ({
  success: true,
  message: 'AuditResultReport fetched successfully',
  data: { items, pagination: { page: 1, page_size: 20, total: items.length, total_pages: 1 } }
})

// Fake of the backend recent-findings merge: saved report findings + live
// auto-findings (KKA / fieldwork) not yet in a saved report, deduped per
// assignment letter by title, newest first.
function buildRecentFindings(reports: any[], autoFindings: Record<string, any[]>, limit: number) {
  const seen = new Set<string>()
  const items: any[] = []
  for (const r of reports) {
    for (const f of r.findings || []) {
      const key = `${r.assignmentLetterId}|${f.title.toLowerCase()}`
      if (seen.has(key)) continue
      seen.add(key)
      items.push({
        title: f.title,
        category: f.category,
        action: f.action || '',
        source: f.source || 'Audit Result Report',
        assignmentLetterId: r.assignmentLetterId,
        reportId: r.id,
        date: r.created_at
      })
    }
  }
  for (const [letter, list] of Object.entries(autoFindings)) {
    for (const f of list) {
      const key = `${letter}|${f.title.toLowerCase()}`
      if (seen.has(key)) continue
      seen.add(key)
      items.push({
        title: f.title,
        category: f.category,
        action: f.action || '',
        source: f.source,
        assignmentLetterId: letter,
        reportId: null,
        date: f.date
      })
    }
  }
  items.sort((a, b) => Date.parse(b.date) - Date.parse(a.date))
  return { success: true, message: 'ok', data: { items: items.slice(0, limit), total: items.length } }
}

// Routes /audit-result-reports, /audit-result-reports/auto-findings and
// /audit-result-reports/recent-findings.
function mockApi({ reports = [], autoFindings = {}, fail = false, recentItems = null }: any = {}) {
  global.$fetch = vi.fn(async (url: string) => {
    if (fail) throw new Error('Network error')
    const u = new URL(String(url), 'http://x')
    if (u.pathname.endsWith('/audit-result-reports/recent-findings')) {
      if (recentItems) return { success: true, message: 'ok', data: { items: recentItems, total: recentItems.length } }
      return buildRecentFindings(reports, autoFindings, Number(u.searchParams.get('limit') || 5))
    }
    if (u.pathname.endsWith('/audit-result-reports/auto-findings')) {
      const letter = u.searchParams.get('assignmentLetterId')
      const list = autoFindings[letter] || []
      return { success: true, data: { assignmentLetterId: letter, total: list.length, findings: list } }
    }
    if (u.pathname.endsWith('/audit-result-reports')) return listResponse(reports)
    return { success: true, data: [] }
  })
}

// Store + what the dashboard does on mount.
async function loadStore(api: any) {
  mockApi(api)
  const store = useAuditResultReportStore() // also fires fetchReports() on creation
  await store.fetchReports()
  await store.fetchRecentFindings(5)
  return store
}

// --- Dashboard formula, replicated from pages/dashboard/index.vue (SFC not importable) ---
//  recentFindingsData = recentFindings.map(f => ({ audit_finding, findings_category }))
const dashboardRecentFindings = (store: any) =>
  store.recentFindings.map((f: any) => ({ audit_finding: f.title, findings_category: f.category }))

const titles = (rows: any[]) => rows.map(r => r.audit_finding)

describe('Dashboard Recent Finding Issues vs Audit Result Report / Fieldwork / KKA', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  it('lists the findings saved on the audit result reports, with their category, from the recent-findings endpoint', async () => {
    const store = await loadStore({
      reports: [
        report('r1', 'ST-001', '2026-09-20', [finding('Kas tidak direkonsiliasi', 'Very Significant'), finding('Akses kasir aktif')]),
        report('r2', 'ST-002', '2026-09-10', [finding('HPS tanpa survei harga', 'Quite Significant')])
      ]
    })
    expect(dashboardRecentFindings(store)).toEqual([
      { audit_finding: 'Kas tidak direkonsiliasi', findings_category: 'Very Significant' },
      { audit_finding: 'Akses kasir aktif', findings_category: 'Significant' },
      { audit_finding: 'HPS tanpa survei harga', findings_category: 'Quite Significant' }
    ])
    const urls = vi.mocked($fetch).mock.calls.map(c => String(c[0]))
    expect(urls.some(u => u.endsWith('/audit-result-reports/recent-findings?limit=5'))).toBe(true)
  })

  it('a store that has not fetched yet holds no mock reports or findings', () => {
    global.$fetch = vi.fn(() => new Promise(() => {})) // request never resolves
    const store = useAuditResultReportStore()
    expect(store.reportList).toEqual([])
    expect(store.recentFindings).toEqual([])
    expect(dashboardRecentFindings(store)).toEqual([])
  })

  it('no audit result reports or findings in the API → the card shows no findings', async () => {
    const store = await loadStore({ reports: [] })
    expect(store.reportList).toHaveLength(0)
    expect(dashboardRecentFindings(store)).toEqual([])
    expect(store.recentFindingsError).toBe('')
    expect(store.recentFindingsLoading).toBe(false)
  })

  it('an API error does not fill the card (or the report list) with mock findings', async () => {
    const store = await loadStore({ fail: true })
    expect(dashboardRecentFindings(store)).toEqual([])
    expect(store.recentFindingsError).toBeTruthy()
    expect(store.reportList).toEqual([])
    expect(store.errorMsg).toBeTruthy()
  })

  it('findingsCount matches the findings actually saved on the report', async () => {
    // e.g. findingsCount saved as 8 when the report was created, findings later edited down to 2.
    const store = await loadStore({
      reports: [report('r1', 'ST-001', '2026-09-20', [finding('A'), finding('B')], { findingsCount: 8 })]
    })
    expect(store.reportList[0].findingsCount).toBe(store.reportList[0].findings.length)
    expect(store.reportList[0].findingsCount).toBe(2)
  })

  it('shows findings from the newest reports first, whatever order the report list returns', async () => {
    const older = report('r-old', 'ST-001', '2026-01-15', [finding('Temuan lama 1'), finding('Temuan lama 2'), finding('Temuan lama 3')])
    const newer = report('r-new', 'ST-009', '2026-09-25', [finding('Temuan baru 1'), finding('Temuan baru 2')])

    const store = await loadStore({ reports: [older, newer] })
    expect(titles(dashboardRecentFindings(store)).slice(0, 2)).toEqual(['Temuan baru 1', 'Temuan baru 2'])
    expect(dashboardRecentFindings(store)).toHaveLength(5)
  })

  it('keeps newest first even if the endpoint items arrive out of order', async () => {
    const store = await loadStore({
      recentItems: [
        { title: 'Lama', category: 'Significant', assignmentLetterId: 'ST-001', reportId: 'r1', date: '2026-01-01T00:00:00Z' },
        { title: 'Baru', category: 'Significant', assignmentLetterId: 'ST-002', reportId: null, date: '2026-09-01T00:00:00Z' }
      ]
    })
    expect(titles(dashboardRecentFindings(store))).toEqual(['Baru', 'Lama'])
  })

  it('a finding recorded in Fieldwork / KKA after the report was saved still shows on the card', async () => {
    // ARR for ST-001 was saved with one finding; an ineffective test control was
    // recorded in Audit Fieldwork afterwards (auto-findings now returns both).
    const saved = report('r1', 'ST-001', '2026-09-20', [finding('Kas tidak direkonsiliasi', 'Very Significant')])
    const store = await loadStore({
      reports: [saved],
      autoFindings: {
        'ST-001': [
          { title: 'Kas tidak direkonsiliasi', category: 'Very Significant', source: 'Digital Working Paper (KKA - AOI & RCA)', date: '2026-09-18T00:00:00Z' },
          { title: 'Kelemahan Kontrol: Otorisasi transaksi', category: 'Very Significant', source: 'Audit Fieldwork (Test Controls)', date: '2026-09-28T00:00:00Z' }
        ]
      }
    })

    // Sanity: the source data does contain the new fieldwork finding.
    const current = await store.fetchAutoFindings('ST-001')
    expect(titles(current.map(f => ({ audit_finding: f.title })))).toContain('Kelemahan Kontrol: Otorisasi transaksi')

    // It is on the card (newest, not tied to a saved report) and the saved finding is not duplicated.
    expect(titles(dashboardRecentFindings(store))).toEqual([
      'Kelemahan Kontrol: Otorisasi transaksi',
      'Kas tidak direkonsiliasi'
    ])
    expect(store.recentFindings[0].reportId).toBeNull()
  })
})
