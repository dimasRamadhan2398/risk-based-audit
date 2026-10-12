// @ts-nocheck
/**
 * LHA (Audit Result Report) side of the ATRs: after an LHA is approved the user is told how many ATRs
 * were created, with a link to the ATR list filtered by that LHA; and re-detected findings keep the ids
 * their ATRs point at.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import en from '~/locales/en/common.json'
import id from '~/locales/id/common.json'
import { useAuditResultReportStore } from '~/stores/audit-result-report'

const toast = vi.hoisted(() => ({
  showSuccess: vi.fn(), showError: vi.fn(), showWarning: vi.fn(), showInfo: vi.fn(), add: vi.fn()
}))
vi.mock('~/components/shared/ToastNotification.vue', () => ({ useToastNotification: () => toast }))

const LHA = 'lha-1'
const finding = (title, extra = {}) => ({ title, category: 'Significant', action: 'Fix', source: 'Audit Features', ...extra })
const storedReport = status => ({
  id: LHA,
  assignmentLetterId: 'letter-1',
  reportTitle: 'Audit Kas',
  reportNumber: 'LHA-021/SKAI/2026',
  reportDate: '2026-10-01',
  status,
  findings: [finding('Kas selisih', { id: 'f-1' }), finding('Otorisasi lemah', { id: 'f-2' })]
})

/** The LHA list holds the report in `status`; the ATR count for the LHA is `atrTotal` (or an error). */
const backend = ({ status = 'Draft', atrTotal = 2, atrError = false, autoFindings = [] } = {}) => vi.fn(async (url, opts = {}) => {
  const method = (opts.method || 'GET').toUpperCase()
  if (url.endsWith('/action-taken-reports')) {
    if (atrError) throw new Error('down')
    return { success: true, data: { items: [], pagination: { page: 1, page_size: 1, total: atrTotal, total_pages: atrTotal } } }
  }
  if (url.includes('/auto-findings')) return { success: true, data: { findings: autoFindings } }
  if (method === 'GET' && url.endsWith('/audit-result-reports')) return { success: true, data: { items: [storedReport(status)] } }
  return { success: true, data: { id: LHA } }
})

const atrCountCalls = () => global.$fetch.mock.calls.filter(([url]) => url.endsWith('/action-taken-reports'))

describe('LHA approval -> ATRs created', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  const approve = async (opts) => {
    global.$fetch = backend(opts)
    const store = useAuditResultReportStore()
    await store.fetchReports()
    store.editReport(store.reportList[0])
    store.reportForm.status = 'Final'
    await store.saveReport()
    return store
  }

  it('saving Draft -> Final keeps the finding ids and toasts "N ATRs created" with a link filtered by the LHA', async () => {
    await approve({ atrTotal: 2 })
    const put = global.$fetch.mock.calls.find(([, o]) => o?.method === 'PUT')
    expect(put[1].body.findings.map(f => f.id)).toEqual(['f-1', 'f-2'])
    expect(atrCountCalls()[0][1].params).toEqual({ audit_result_report_id: LHA, page: 1, page_size: 1 })

    expect(toast.add).toHaveBeenCalledTimes(1)
    const shown = toast.add.mock.calls[0][0]
    expect(shown.title).toBe(en.auditResultReport.followUp.createdTitle.replace('{count}', '2'))
    expect(shown.actions[0].label).toBe(en.auditResultReport.followUp.view.replace('{count}', '2'))
    shown.actions[0].onClick()
    expect(global.navigateTo).toHaveBeenCalledWith(`/action-taken-report?lha=${LHA}`)
  })

  it('no ATR found after approval: a warning instead of a success link', async () => {
    await approve({ atrTotal: 0 })
    expect(toast.add).not.toHaveBeenCalled()
    expect(toast.showWarning).toHaveBeenCalledWith(en.auditResultReport.followUp.noneTitle, en.auditResultReport.followUp.none)
  })

  it('count unavailable: an info toast, never a made-up number', async () => {
    await approve({ atrError: true })
    expect(toast.add).not.toHaveBeenCalled()
    expect(toast.showInfo).toHaveBeenCalled()
  })

  it('re-saving an already approved LHA does not announce ATRs again', async () => {
    global.$fetch = backend({ status: 'Final' })
    const store = useAuditResultReportStore()
    await store.fetchReports()
    store.editReport(store.reportList[0])
    await store.saveReport()
    expect(atrCountCalls()).toHaveLength(0)
    expect(toast.add).not.toHaveBeenCalled()
  })

  it('every follow-up string exists in en and id', () => {
    for (const key of Object.keys(en.auditResultReport.followUp)) {
      expect(typeof id.auditResultReport.followUp[key], key).toBe('string')
    }
  })
})

describe('re-detected findings keep their ids', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.spyOn(console, 'warn').mockImplementation(() => {})
  })

  it('a "replace" sync reuses the id of the saved finding with the same title; new findings get none', async () => {
    global.$fetch = backend({
      status: 'Final',
      autoFindings: [finding(' kas SELISIH '), finding('Temuan baru'), finding('Otorisasi lemah')]
    })
    const store = useAuditResultReportStore()
    await store.fetchReports()
    store.editReport(store.reportList[0])
    await store.runAutoDetectFindings('replace')
    expect(store.reportForm.findings.map(f => [f.title.trim(), f.id])).toEqual([
      ['kas SELISIH', 'f-1'],
      ['Temuan baru', undefined],
      ['Otorisasi lemah', 'f-2']
    ])
  })
})
