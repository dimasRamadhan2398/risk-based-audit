import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import {
  applyFollowUpToTopFindings,
  followUpClosedPercent,
  followUpStatusI18nKey,
  followUpStatusParams,
  matrixRowsFromFollowUp,
  noActionReasonI18nKey,
  parseFollowUpStatus
} from '~/utils/executiveSummaryFollowUp'
import { useExecutiveSummaryStore, withAllFollowUpRows } from '~/stores/executive-summary'
import { useAuthStore } from '~/stores/auth'
import { useRbac } from '~/composables/useRbac'
import { UserRole } from '~/types/auth'

// Body of GET /executive-summaries/follow-up-status (contract shared with the audit-service)
const body = (overrides: Record<string, unknown> = {}) => ({
  asOf: '2026-11-04T09:00:00+07:00',
  buckets: { closed: 1, inProgress: 1, overdue: 1, noAction: 1 },
  total: 4,
  findings: [
    { findingId: 'f1', auditResultReportId: 'r1', reportNumber: 'LHA-1', atrId: 'a1', title: 'Closed one', divisi: 'FIN', atrStatus: 'CLOSED', bucket: 'CLOSED', noActionReason: null, dueDate: '2026-10-01', progress: 100, picName: 'Budi', overdueDays: 0, evidence: [{ id: 'e1', fileName: 'bukti.pdf', uploadedAt: '2026-10-02T03:00:00.123Z' }] },
    { findingId: 'f2', auditResultReportId: 'r1', reportNumber: 'LHA-1', atrId: 'a2', title: 'Running', divisi: '', atrStatus: 'SOMETHING_NEW', evidence: [], bucket: 'IN_PROGRESS', noActionReason: null, dueDate: '2026-12-01', progress: 40, picName: 'Sari', overdueDays: 0 },
    { findingId: 'f3', auditResultReportId: 'r1', reportNumber: 'LHA-1', atrId: 'a3', title: 'Late', atrStatus: 'IN_PROGRESS', bucket: 'OVERDUE', noActionReason: null, dueDate: '2026-09-01', progress: 10, picName: 'Andi', overdueDays: 64 },
    { findingId: 'f4', auditResultReportId: 'r1', reportNumber: 'LHA-1', atrId: null, title: 'Nothing yet', atrStatus: null, bucket: 'NO_ACTION', noActionReason: 'NO_PIC', dueDate: null, progress: null, picName: null, overdueDays: 0 },
    { findingId: '', bucket: 'CLOSED' },
    { findingId: 'bad', bucket: 'CANCELLED' }
  ],
  ...overrides
})

describe('parseFollowUpStatus', () => {
  it('maps the four buckets and the valid findings, dropping malformed rows', () => {
    const r = parseFollowUpStatus(body())
    expect(r.buckets).toEqual({ closed: 1, inProgress: 1, overdue: 1, noAction: 1 })
    expect(r.total).toBe(4)
    expect(r.findingsRestricted).toBe(false)
    expect(r.findings.map(f => f.findingId)).toEqual(['f1', 'f2', 'f3', 'f4'])
    expect(r.findings[0]!.evidence).toBe('bukti.pdf')
    expect(r.findings[1]!.evidence).toBe('')
    expect(r.findings[3]!.noActionReason).toBe('NO_PIC')
    expect(r.findings[3]!.progress).toBeNull()
  })

  it('accepts the response wrapped in { data } and a missing noAction bucket', () => {
    const r = parseFollowUpStatus({ success: true, data: body({ buckets: { closed: 2, inProgress: 0, overdue: 0 }, total: undefined }) })
    expect(r.buckets.noAction).toBe(0)
    expect(r.total).toBe(2)
  })

  it('keeps findingsRestricted and throws when the bucket counts are missing', () => {
    expect(parseFollowUpStatus(body({ findingsRestricted: true, findings: [] })).findingsRestricted).toBe(true)
    expect(() => parseFollowUpStatus(body({ buckets: undefined }))).toThrow(/buckets/)
    expect(() => parseFollowUpStatus(body({ buckets: { closed: -1, inProgress: 0, overdue: 0 } }))).toThrow(/closed/)
    expect(() => parseFollowUpStatus(null)).toThrow()
  })
})

describe('followUpStatusParams', () => {
  it('sends the query parameter names of the contract', () => {
    expect(followUpStatusParams({ auditResultReportId: 'u' })).toEqual({ audit_result_report_id: 'u' })
    expect(followUpStatusParams({ reportNumber: '001/LHA' })).toEqual({ report_number: '001/LHA' })
    expect(followUpStatusParams({ year: 2026, quarter: 3 })).toEqual({ year: 2026, quarter: 3 })
  })
})

describe('applyFollowUpToTopFindings', () => {
  it('sets statusTL, atrId and noActionReason from the bucket of linked rows only', () => {
    const { findings } = parseFollowUpStatus(body())
    const rows = [
      { findingId: 'f3', statusTL: 'In Progress' },
      { findingId: 'f4', statusTL: 'In Progress' },
      { statusTL: 'Closed' }
    ] as Parameters<typeof applyFollowUpToTopFindings>[0]
    applyFollowUpToTopFindings(rows, findings)
    expect(rows[0]).toMatchObject({ statusTL: 'Overdue', atrId: 'a3', noActionReason: null })
    expect(rows[1]).toMatchObject({ statusTL: 'No Action', atrId: null, noActionReason: 'NO_PIC' })
    expect(rows[2]).toEqual({ statusTL: 'Closed' })
  })
})

describe('matrixRowsFromFollowUp', () => {
  it('builds one row per finding with the real PIC, progress, due date and evidence', () => {
    const rows = matrixRowsFromFollowUp(parseFollowUpStatus(body()), id => (id === 'f1' ? { category: 'Significant', action: 'Fix SOP', reportTitle: 'Audit Keuangan' } : undefined))
    expect(rows).toHaveLength(4)
    expect(rows[0]).toMatchObject({ nomor: 'LHA-1', status: 'Closed', picUnit: 'Budi', progres: 100, dueDate: '2026-10-01', buktiTL: 'bukti.pdf', nilaiRisiko: 'Significant', rekomendasi: 'Fix SOP', findingId: 'f1', atrId: 'a1' })
    expect(rows[3]).toMatchObject({ status: 'No Action', picUnit: '', progres: 0, dueDate: '', buktiTL: '', noActionReason: 'NO_PIC' })
    expect(rows.some(r => r.picUnit.startsWith('PIC ') || r.buktiTL.startsWith('LHA_'))).toBe(false)
  })

  it('is empty when the per-finding detail is restricted', () => {
    expect(matrixRowsFromFollowUp(parseFollowUpStatus(body({ findingsRestricted: true })))).toEqual([])
  })
})

describe('followUpClosedPercent', () => {
  it('works for old three-row and new four-row tables', () => {
    expect(followUpClosedPercent([{ status: 'Closed', jumlah: 1 }, { status: 'In Progress', jumlah: 1 }, { status: 'Overdue', jumlah: 0 }])).toBe(50)
    expect(followUpClosedPercent([{ status: 'Closed', jumlah: 1 }, { status: 'In Progress', jumlah: 1 }, { status: 'Overdue', jumlah: 1 }, { status: 'No Action', jumlah: 1 }])).toBe(25)
    expect(followUpClosedPercent([])).toBe(0)
    expect(followUpClosedPercent(undefined)).toBe(0)
  })
})

describe('follow-up i18n keys', () => {
  const locale = (lang: string) => JSON.parse(readFileSync(resolve(__dirname, `../../locales/${lang}/common.json`), 'utf8'))
  const get = (obj: unknown, key: string): unknown =>
    key.split('.').reduce<unknown>((o, k) => (o as Record<string, unknown> | undefined)?.[k], obj)

  it.each(['en', 'id'])('%s has every status and reason label', (lang) => {
    const json = locale(lang)
    for (const s of ['Closed', 'In Progress', 'Overdue', 'No Action']) {
      expect(typeof get(json, followUpStatusI18nKey(s))).toBe('string')
    }
    for (const r of ['NO_PIC', 'NO_PLAN_14D', 'NO_PLAN']) {
      expect(typeof get(json, noActionReasonI18nKey(r))).toBe('string')
    }
    expect(typeof get(json, 'executiveSummaryForm.followUp.statusAsOf')).toBe('string')
    expect(typeof get(json, 'executiveSummaryForm.followUp.current.title')).toBe('string')
  })

  it('uses the agreed Indonesian terms', () => {
    const id = locale('id')
    expect(get(id, 'executiveSummaryForm.followUp.statuses')).toEqual({
      closed: 'Selesai', inProgress: 'Dalam Proses', overdue: 'Terlambat', noAction: 'Belum Ditindaklanjuti'
    })
    expect(get(id, 'executiveSummaryForm.followUp.current.title')).toBe('Status terkini')
    expect(get(id, 'executiveSummaryForm.followUp.statusAsOf')).toContain('Status per')
  })
})

describe('store.fetchFollowUpStatus', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('calls the follow-up-status endpoint with the query and parses the body', async () => {
    const calls: { url: string, params?: Record<string, unknown> }[] = []
    vi.stubGlobal('$fetch', vi.fn(async (url: string, opts?: { params?: Record<string, unknown> }) => {
      calls.push({ url, params: opts?.params })
      if (url.endsWith('/executive-summaries')) return { data: { items: [] } }
      // The audit-service wraps the payload: { success, message, data }
      return { success: true, message: 'ok', data: body({ findingsRestricted: false }) }
    }))
    const store = useExecutiveSummaryStore()
    const result = await store.fetchFollowUpStatus({ year: 2026, quarter: 4 })
    const call = calls.find(c => c.url.endsWith('/executive-summaries/follow-up-status'))
    expect(call?.params).toEqual({ year: 2026, quarter: 4 })
    expect(result.buckets.noAction).toBe(1)
  })

  it('starts a new summary with the four follow-up rows', () => {
    vi.stubGlobal('$fetch', vi.fn(async () => ({ data: { items: [] } })))
    const store = useExecutiveSummaryStore()
    store.openNewForm(1)
    expect(store.form.followUpTable.map(r => r.status)).toEqual(['Closed', 'In Progress', 'Overdue', 'No Action'])
  })
})

describe('withAllFollowUpRows', () => {
  it('adds the missing No Action row to an old three-row draft and keeps existing rows', () => {
    const rows = withAllFollowUpRows([
      { status: 'Closed', jumlah: 2, persentase: 50, keterangan: 'ok' },
      { status: 'In Progress', jumlah: 1, persentase: 25, keterangan: '' },
      { status: 'Overdue', jumlah: 1, persentase: 25, keterangan: '' }
    ])
    expect(rows.map(r => r.status)).toEqual(['Closed', 'In Progress', 'Overdue', 'No Action'])
    expect(rows[0]).toMatchObject({ jumlah: 2, keterangan: 'ok' })
    expect(rows[3]).toMatchObject({ jumlah: 0 })
    expect(withAllFollowUpRows(undefined)).toHaveLength(4)
  })
})

describe('canViewLiveFollowUpStatus', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('is granted to the holders of view_all ATR and not to VIEWER', () => {
    const auth = useAuthStore()
    const as = (role: string) => {
      auth.user = { id: role, username: role, email: `${role}@test.com`, fullName: role, roles: [role] }
      return useRbac().canViewLiveFollowUpStatus.value
    }
    expect(as(UserRole.ADMIN)).toBe(true)
    expect(as(UserRole.EXECUTIVE)).toBe(true)
    expect(as(UserRole.AUDITOR)).toBe(true)
    expect(as(UserRole.AUDIT_MANAGER)).toBe(true)
    expect(as(UserRole.CHIEF_AUDIT_EXECUTIVE)).toBe(true)
    expect(as(UserRole.VIEWER)).toBe(false)
    expect(as(UserRole.AUDITEE)).toBe(false)
  })
})

describe('forms no longer fabricate follow-up data', () => {
  const read = (p: string) => readFileSync(resolve(__dirname, '../../', p), 'utf8')

  it('compilation matrix has no placeholder PIC, generated file names or LHA-status mapping', () => {
    const c = read('components/audit-result-report/ExecutiveSummaryCompilationForm.vue')
    expect(c).not.toContain('picUnit: `PIC ${div}`')
    expect(c).not.toContain('buktiTL: `LHA_')
    expect(c).not.toMatch(/status: r\.status === 'Final'/)
    expect(c).toContain('matrixRowsFromFollowUp')
  })

  it('individual Top 5 status comes from the ATR bucket', () => {
    const c = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')
    expect(c).toContain('BUCKET_STATUS[atr.bucket]')
    expect(c).toContain('findingId: f.id')
    expect(c).not.toContain('@input="recalculatePercentages"')
  })
})
