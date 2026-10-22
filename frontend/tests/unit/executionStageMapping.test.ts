/* eslint-disable @typescript-eslint/no-explicit-any -- test doubles for $fetch and list payloads */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'

import { useAuditExecutionStore } from '~/stores/audit-execution'
import {
  describeDraftBand,
  describeExecutionSources,
  getDisplayProgress,
  getExecutionStageIndex,
  getProgressValue,
  getReportingSubStage,
  LHA_WAITING_APPROVAL_PROGRESS,
  progressToStageStatus,
  toSourceCount
} from '~/utils/executionSources'
import { EXECUTION_PHASES, getExecutionPhase } from '~/types/audit'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: { auditServiceBaseUrl: 'http://localhost:8002/api/v1' }
  })
}))

const ROOT = resolve(__dirname, '../..')
const LOCALES = ['en', 'id'].map(lang =>
  JSON.parse(readFileSync(resolve(ROOT, `locales/${lang}/common.json`), 'utf-8'))
)
const SOURCE_FILES = [
  'stores/audit-execution.ts',
  'pages/audit-execution-status/index.vue',
  'pages/dashboard/index.vue',
  'components/audit-execution/AuditExecutionDetailModal.vue',
  'utils/executionSources.ts'
].map(f => ({ path: f, src: readFileSync(resolve(ROOT, f), 'utf-8') }))

const approvedSources = {
  annual_plan: true,
  assignment_letter: true,
  fieldwork_records: 3,
  test_controls: 4,
  test_controls_tested: 4,
  test_controls_with_finding: 2,
  mitigation_plans: 2,
  working_paper_records: 2,
  working_paper_causes: 2,
  working_paper_plans: 2,
  result_report: true,
  result_report_approved: true,
  result_report_in_review: false
}

const waitingApprovalSources = {
  ...approvedSources,
  result_report_approved: false,
  result_report_in_review: true
}

describe('stage mapping at each boundary', () => {
  const cases: Array<[number, string, number]> = [
    [0, 'planning', 0],
    [1, 'entry meeting', 1],
    [25, 'entry meeting', 1],
    [26, 'fieldwork', 2],
    [50, 'fieldwork', 2],
    [51, 'draft findings', 3],
    [75, 'draft findings', 3],
    [76, 'reporting', 4],
    [89, 'reporting', 4],
    [90, 'reporting', 4],
    [99, 'reporting', 4],
    [100, 'completed', 5]
  ]

  it.each(cases)('progress %i maps to %s (phase index %i)', (progress, stage, index) => {
    expect(progressToStageStatus(progress)).toBe(stage)
    expect(getExecutionStageIndex(progress)).toBe(index)
  })

  it.each(cases)('progress %i maps to the matching phase step', (progress, _stage, index) => {
    expect(getExecutionPhase(progress).step).toBe(EXECUTION_PHASES[index]!.step)
  })

  it('splits reporting into drafting (76–87) and approval (88–99)', () => {
    expect(getReportingSubStage(76)).toBe('drafting')
    expect(getReportingSubStage(87)).toBe('drafting')
    expect(getReportingSubStage(88)).toBe('approval')
    expect(getReportingSubStage(90)).toBe('approval')
    expect(getReportingSubStage(99)).toBe('approval')
    expect(getReportingSubStage(75)).toBeNull()
    expect(getReportingSubStage(100)).toBeNull()
  })

  it('keeps the phase table bands in line with the stage thresholds', () => {
    expect(EXECUTION_PHASES.map(p => p.minProgress)).toEqual([0, 1, 26, 51, 76, 100])
    expect(EXECUTION_PHASES.map(p => p.maxProgress)).toEqual([0, 25, 50, 75, 99, 100])
  })
})

describe('progress as an integer 0–100', () => {
  it('coerces numeric strings and rounds', () => {
    expect(getProgressValue('40')).toBe(40)
    expect(getProgressValue('40.6')).toBe(41)
    expect(getProgressValue(' 75 ')).toBe(75)
  })

  it('clamps and falls back to 0 for values that are not numbers', () => {
    expect(getProgressValue(150)).toBe(100)
    expect(getProgressValue(-5)).toBe(0)
    expect(getProgressValue('abc')).toBe(0)
    expect(getProgressValue(undefined)).toBe(0)
    expect(getProgressValue(null)).toBe(0)
    expect(getProgressValue(NaN)).toBe(0)
  })
})

describe('waiting approval LHA', () => {
  it('shows about 90 (reporting approval) when the LHA is in review and progress is still a draft value', () => {
    expect(getDisplayProgress(76, waitingApprovalSources)).toBe(LHA_WAITING_APPROVAL_PROGRESS)
    expect(getProgressValue(getDisplayProgress(76, waitingApprovalSources))).toBe(90)
    expect(progressToStageStatus(getDisplayProgress(76, waitingApprovalSources))).toBe('reporting')
  })

  it('never shows 100 for an LHA that is not approved', () => {
    expect(getDisplayProgress(100, { ...approvedSources, result_report_approved: false })).toBe(90)
    expect(progressToStageStatus(getDisplayProgress(100, { ...approvedSources, result_report_approved: false }))).toBe('reporting')
  })

  it('keeps a server value that is already inside the approval band', () => {
    expect(getDisplayProgress(88, waitingApprovalSources)).toBe(88)
    expect(getDisplayProgress(95, waitingApprovalSources)).toBe(95)
  })

  it('shows 100 only when the LHA is approved', () => {
    expect(getDisplayProgress(100, approvedSources)).toBe(100)
    expect(progressToStageStatus(getDisplayProgress(100, approvedSources))).toBe('completed')
  })

  it('leaves progress alone when the review flags are missing (older payloads)', () => {
    const { result_report_in_review: _ignored, ...withoutInReview } = waitingApprovalSources
    expect(getDisplayProgress(76, withoutInReview)).toBe(76)
    expect(getDisplayProgress(100, { result_report: true })).toBe(100)
    expect(getDisplayProgress(76, undefined)).toBe(76)
  })

  it('moves a waiting-approval execution from completed to reporting in the store', async () => {
    setActivePinia(createPinia())
    global.$fetch = vi.fn(async () => ({
      data: {
        items: [
          { id: 'w1', ref: 'AUD-1', name: 'Waiting', progress: 100, status: 'completed', sources: { ...approvedSources, result_report_approved: false } },
          { id: 'w2', ref: 'AUD-2', name: 'Waiting in review', progress: '76', status: 'reporting', sources: waitingApprovalSources },
          { id: 'w3', ref: 'AUD-3', name: 'Approved', progress: '100', status: 'completed', sources: approvedSources }
        ]
      }
    })) as any

    const store = useAuditExecutionStore()
    await store.fetchAuditExecutions()

    const [waiting, inReview, approved] = store.auditExecutions
    expect(waiting).toMatchObject({ progress: 90, status: 'reporting' })
    expect(inReview).toMatchObject({ progress: 90, status: 'reporting' })
    expect(approved).toMatchObject({ progress: 100, status: 'completed' })
  })
})

describe('draft findings band requirements', () => {
  it('shows the mitigation and working paper plan counts from sources', () => {
    const band = describeDraftBand(approvedSources)
    expect(band.plans).toBe(2)
    expect(band.mitigations).toBe(2)
    expect(band.findings).toBe(2)
    expect(band.complete).toBe(true)
    expect(band.requirements.map(r => r.key)).toEqual(['findings', 'working_paper_plans', 'mitigation'])
  })

  it('reports the missing requirement when mitigation is not recorded', () => {
    const band = describeDraftBand({ ...approvedSources, mitigation_plans: 0 })
    expect(band.complete).toBe(false)
    expect(band.requirements.find(r => r.key === 'mitigation')!.done).toBe(false)
  })

  it('counts test controls with a finding when no causes are recorded', () => {
    expect(describeDraftBand({ working_paper_causes: 0, test_controls_with_finding: 3 }).findings).toBe(3)
  })
})

describe('missing or malformed sources', () => {
  it('shows 0, not NaN, for every count when sources are missing', () => {
    for (const sources of [undefined, null, {}]) {
      const rows = describeExecutionSources(sources as any)
      for (const row of rows) {
        for (const value of [row.count, row.tested, row.total, row.mitigations, row.plans]) {
          expect(Number.isNaN(value)).toBe(false)
          expect(value).toBe(0)
        }
      }
      const band = describeDraftBand(sources as any)
      expect([band.findings, band.plans, band.mitigations].every(n => n === 0)).toBe(true)
      expect(band.complete).toBe(false)
    }
  })

  it('treats a sources object with only some fields as partial, not broken', () => {
    const rows = describeExecutionSources({ fieldwork_records: 2, test_controls_tested: 1 } as any)
    const fieldwork = rows.find(r => r.key === 'fieldwork')!
    expect(fieldwork).toMatchObject({ done: true, count: 2, tested: 1, total: 0, mitigations: 0 })
    expect(rows.find(r => r.key === 'result_report')).toMatchObject({ done: false, approved: false })
  })

  it('coerces numeric strings and turns junk into 0', () => {
    expect(toSourceCount('3')).toBe(3)
    expect(toSourceCount('abc')).toBe(0)
    expect(toSourceCount(-2)).toBe(0)
    expect(toSourceCount(undefined)).toBe(0)
    const rows = describeExecutionSources({ fieldwork_records: '2', mitigation_plans: 'n/a' } as any)
    expect(rows.find(r => r.key === 'fieldwork')).toMatchObject({ count: 2, mitigations: 0 })
  })

  it('does not throw when the list item has no sources at all', async () => {
    setActivePinia(createPinia())
    global.$fetch = vi.fn(async () => ({
      data: { items: [{ id: 'n1', ref: 'AUD-9', name: 'No sources', progress: '30', status: '' }] }
    })) as any

    const store = useAuditExecutionStore()
    await store.fetchAuditExecutions()

    expect(store.error).toBeNull()
    expect(store.auditExecutions[0]).toMatchObject({ progress: 30, status: 'fieldwork' })
  })
})

describe('audit execution list does not call the sync endpoint', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('fetching the list issues one GET to /audit-executions and nothing else', async () => {
    const calls: string[] = []
    global.$fetch = vi.fn(async (url: string, opts: any) => {
      calls.push(`${opts?.method ?? 'GET'} ${url}`)
      return { data: { items: [{ id: 'a', ref: 'R', name: 'A', progress: 0, status: 'planning' }] } }
    }) as any

    await useAuditExecutionStore().fetchAuditExecutions()

    expect(calls).toHaveLength(1)
    expect(calls[0]).toMatch(/^GET \S*\/audit-executions$/)
    expect(calls.some(c => /sync/i.test(c))).toBe(false)
  })

  // The sync (create executions from assignment letters + recompute) is an explicit store action;
  // nothing else talks to the endpoint, and the list fetch above stays read-only.
  it('only the store action calls POST /audit-executions/sync', () => {
    for (const { path, src } of SOURCE_FILES) {
      if (path === 'stores/audit-execution.ts') continue
      expect(src, `${path} references the sync endpoint`).not.toMatch(/audit-executions\/sync/)
    }
    const store = SOURCE_FILES.find(f => f.path === 'stores/audit-execution.ts')!.src
    expect(store).toMatch(/async syncFromAssignmentLetters\(/)
  })

  it('the list page syncs only through the store action, auto-sync guarded to once per visit', () => {
    const page = SOURCE_FILES.find(f => f.path === 'pages/audit-execution-status/index.vue')!.src
    expect(page).not.toMatch(/\$fetch\(/)
    const syncCalls = page.match(/store\.sync\w*\s*\(/g) ?? []
    expect(syncCalls.every(c => c.startsWith('store.syncFromAssignmentLetters'))).toBe(true)
    expect(page).toMatch(/createAutoSyncOnce\(\(\) => store\.syncFromAssignmentLetters\(\)\)/)
  })
})

describe('locales for the new draft band and waiting approval copy', () => {
  const keys = [
    'auditExecution.detailModal.waitingApproval',
    'auditExecution.detailModal.draftBand.title',
    'auditExecution.detailModal.draftBand.description',
    'auditExecution.detailModal.draftBand.findings.label',
    'auditExecution.detailModal.draftBand.findings.detail',
    'auditExecution.detailModal.draftBand.working_paper_plans.label',
    'auditExecution.detailModal.draftBand.working_paper_plans.detail',
    'auditExecution.detailModal.draftBand.mitigation.label',
    'auditExecution.detailModal.draftBand.mitigation.detail'
  ]

  it.each(keys)('%s exists in en and id', (key) => {
    for (const locale of LOCALES) {
      const value = key.split('.').reduce<any>((node, part) => node?.[part], locale)
      expect(typeof value === 'string' && value.length > 0, `${key} missing`).toBe(true)
    }
  })

  it('the count placeholder is present in the detail strings', () => {
    for (const locale of LOCALES) {
      expect(locale.auditExecution.detailModal.draftBand.mitigation.detail).toContain('{count}')
      expect(locale.auditExecution.detailModal.draftBand.working_paper_plans.detail).toContain('{count}')
    }
  })
})
