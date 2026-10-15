import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'

import { useAuditExecutionStore } from '~/stores/audit-execution'
import { describeExecutionSources } from '~/utils/executionSources'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: { auditServiceBaseUrl: 'http://localhost:8002/api/v1' }
  })
}))

const LOCALES = ['en', 'id'].map((lang) =>
  JSON.parse(readFileSync(resolve(__dirname, `../../locales/${lang}/common.json`), 'utf-8'))
)

const fullSources = {
  annual_plan: true,
  assignment_letter: true,
  fieldwork_records: 3,
  test_controls: 4,
  test_controls_tested: 2,
  test_controls_with_finding: 1,
  working_paper_records: 2,
  working_paper_causes: 1,
  result_report: true,
  result_report_approved: true
}

const listPayload = {
  success: true,
  data: {
    items: [
      { id: 'exec-1', ref: 'ST-001/SKAI/2026', name: 'Audit Operasional Keuangan', category: 'Assurance', progress: 0, status: 'planning' },
      { id: 'exec-2', ref: 'ST-002/SKAI/2026', name: 'Audit IT', category: 'Assurance', progress: 0, status: 'planning' }
    ]
  }
}

const listWithSourcesPayload = {
  success: true,
  data: {
    items: [
      { id: 'exec-1', ref: 'ST-001/SKAI/2026', name: 'Audit Operasional Keuangan', progress: 40, status: 'fieldwork', sources: { ...fullSources, result_report: false, result_report_approved: false } },
      { id: 'exec-2', ref: 'ST-002/SKAI/2026', name: 'Audit IT', progress: 100, status: 'completed', sources: { ...fullSources } }
    ]
  }
}

describe('describeExecutionSources', () => {
  it('lists the five source modules in audit-cycle order', () => {
    expect(describeExecutionSources(fullSources).map((r) => r.key)).toEqual([
      'annual_plan',
      'assignment_letter',
      'fieldwork',
      'working_paper',
      'result_report'
    ])
  })

  it('marks every source done when the sync found data in all of them', () => {
    const rows = describeExecutionSources(fullSources)
    expect(rows.every((r) => r.done)).toBe(true)
    expect(rows.find((r) => r.key === 'result_report')!.approved).toBe(true)
  })

  it('treats a missing sources object as nothing synced yet', () => {
    const rows = describeExecutionSources(undefined)
    expect(rows.every((r) => !r.done && !r.approved)).toBe(true)
    expect(rows.every((r) => r.count === 0 && r.tested === 0 && r.total === 0)).toBe(true)
  })

  it('carries fieldwork test-control counts for the detail text', () => {
    const fieldwork = describeExecutionSources(fullSources).find((r) => r.key === 'fieldwork')!
    expect(fieldwork).toMatchObject({ done: true, count: 3, tested: 2, total: 4 })
  })

  it('treats a drafted but unapproved LHA as done and not approved', () => {
    const lha = describeExecutionSources({ ...fullSources, result_report_approved: false })
      .find((r) => r.key === 'result_report')!
    expect(lha).toMatchObject({ done: true, approved: false })
  })

  it('counts zero fieldwork records as not started even when test controls exist', () => {
    const fieldwork = describeExecutionSources({ ...fullSources, fieldwork_records: 0 })
      .find((r) => r.key === 'fieldwork')!
    expect(fieldwork.done).toBe(false)
  })
})

describe('audit execution store — read-only list with derived sources', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('reads the list only and never writes on read', async () => {
    const calls: string[] = []
    global.$fetch = vi.fn(async (url: string, opts: any) => {
      calls.push(`${opts?.method} ${url}`)
      return listPayload
    }) as any

    const store = useAuditExecutionStore()
    await store.fetchAuditExecutions()

    expect(calls).toHaveLength(1)
    expect(calls[0]).toMatch(/^GET \S*\/audit-executions$/)
  })

  it('keeps the sources each execution carries in the list response', async () => {
    global.$fetch = vi.fn(async () => listWithSourcesPayload) as any

    const store = useAuditExecutionStore()
    await store.fetchAuditExecutions()

    const [first, second] = store.auditExecutions
    expect(first!.sources).toMatchObject({ fieldwork_records: 3, result_report: false })
    expect(second!.sources?.result_report_approved).toBe(true)
  })

  it('lists executions without sources when the list has none', async () => {
    global.$fetch = vi.fn(async () => listPayload) as any

    const store = useAuditExecutionStore()
    await store.fetchAuditExecutions()

    expect(store.auditExecutions).toHaveLength(2)
    expect(store.error).toBeNull()
    expect(store.auditExecutions[0]!.sources).toBeUndefined()
  })

  it('coerces string progress to a number and derives the stage from it', async () => {
    global.$fetch = vi.fn(async () => ({
      data: { items: [{ id: 's1', ref: 'ST-1', name: 'A', progress: '40', status: '' }] }
    })) as any

    const store = useAuditExecutionStore()
    await store.fetchAuditExecutions()

    expect(store.auditExecutions[0]!.progress).toBe(40)
    expect(store.auditExecutions[0]!.status).toBe('fieldwork')
  })
})

describe('audit execution source section — locales', () => {
  it('has the source section keys in both en and id', () => {
    for (const locale of LOCALES) {
      const sources = locale.auditExecution.detailModal.sources
      expect(sources.title).toBeTruthy()
      for (const key of ['annual_plan', 'assignment_letter', 'fieldwork', 'working_paper', 'result_report']) {
        expect(sources[key].label, `${key}.label`).toBeTruthy()
      }
      expect(sources.fieldwork.detail).toContain('{tested}')
      expect(sources.fieldwork.detail).toContain('{mitigations}')
      expect(sources.working_paper.detail).toContain('{plans}')
      expect(locale.auditExecution.detailModal.missingRefWarning).toBeTruthy()
      expect(locale.auditExecution.detailModal.progressDerived).toBeTruthy()
    }
  })
})
