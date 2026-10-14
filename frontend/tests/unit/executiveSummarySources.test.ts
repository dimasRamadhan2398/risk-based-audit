import { describe, it, expect, vi, afterEach } from 'vitest'
import { summarizeAtrFollowUp, useExecutiveSummarySources } from '~/composables/useExecutiveSummarySources'
import {
  legacyRiskFields,
  mapWorkingPaperRiskLevel,
  quarterOfDate,
  resolveRiskFields,
  riskBreakdownTotal,
  riskFieldsToBreakdown,
  selectCompilationReports
} from '~/utils/executiveSummaryCompilation'

describe('mapWorkingPaperRiskLevel', () => {
  it('maps the five working paper levels one-to-one', () => {
    expect(mapWorkingPaperRiskLevel('high')).toBe('high')
    expect(mapWorkingPaperRiskLevel('moderate-high')).toBe('moderateHigh')
    expect(mapWorkingPaperRiskLevel('moderate')).toBe('moderate')
    expect(mapWorkingPaperRiskLevel('low-moderate')).toBe('lowModerate')
    expect(mapWorkingPaperRiskLevel('low')).toBe('low')
  })

  it('maps legacy values case-insensitively', () => {
    for (const raw of ['HIGH', 'Critical', 'TINGGI', 'tinggi']) expect(mapWorkingPaperRiskLevel(raw)).toBe('high')
    for (const raw of ['MEDIUM', 'Moderate', 'SEDANG', 'sedang']) expect(mapWorkingPaperRiskLevel(raw)).toBe('moderate')
    for (const raw of ['LOW', 'Rendah', 'RENDAH']) expect(mapWorkingPaperRiskLevel(raw)).toBe('low')
  })

  it('accepts capitalised, spaced, underscored and label variants', () => {
    expect(mapWorkingPaperRiskLevel('Moderate High')).toBe('moderateHigh')
    expect(mapWorkingPaperRiskLevel('MODERATE_HIGH')).toBe('moderateHigh')
    expect(mapWorkingPaperRiskLevel('Moderate to High')).toBe('moderateHigh')
    expect(mapWorkingPaperRiskLevel(' Low Moderate ')).toBe('lowModerate')
    expect(mapWorkingPaperRiskLevel('Low to Moderate')).toBe('lowModerate')
  })

  it('returns null for unknown or empty levels so they are not counted', () => {
    expect(mapWorkingPaperRiskLevel('')).toBeNull()
    expect(mapWorkingPaperRiskLevel(undefined)).toBeNull()
    expect(mapWorkingPaperRiskLevel(null)).toBeNull()
    expect(mapWorkingPaperRiskLevel('extreme')).toBeNull()
  })
})

describe('resolveRiskFields (legacy fallback)', () => {
  it('uses the five new fields when any of them is set', () => {
    const fields = resolveRiskFields({
      risikoHigh: 1, risikoModerateHigh: 2, risikoModerate: 0, risikoLowModerate: 3, risikoLow: 0,
      risikoTinggi: 9, risikoSedang: 9, risikoRendah: 9
    })
    expect(fields).toEqual({ risikoHigh: 1, risikoModerateHigh: 2, risikoModerate: 0, risikoLowModerate: 3, risikoLow: 0 })
  })

  it('falls back to Tinggi -> High, Sedang -> Moderate, Rendah -> Low when all five are 0 or missing', () => {
    expect(resolveRiskFields({ risikoTinggi: 3, risikoSedang: 2, risikoRendah: 1 }))
      .toEqual({ risikoHigh: 3, risikoModerateHigh: 0, risikoModerate: 2, risikoLowModerate: 0, risikoLow: 1 })
    expect(resolveRiskFields({
      risikoHigh: 0, risikoModerateHigh: 0, risikoModerate: 0, risikoLowModerate: 0, risikoLow: 0,
      risikoTinggi: '4', risikoSedang: null, risikoRendah: undefined
    })).toEqual({ risikoHigh: 4, risikoModerateHigh: 0, risikoModerate: 0, risikoLowModerate: 0, risikoLow: 0 })
  })

  it('returns zeros for an empty record and ignores invalid numbers', () => {
    expect(resolveRiskFields({})).toEqual({ risikoHigh: 0, risikoModerateHigh: 0, risikoModerate: 0, risikoLowModerate: 0, risikoLow: 0 })
    expect(resolveRiskFields({ risikoHigh: 'x', risikoLow: -2 })).toEqual({ risikoHigh: 0, risikoModerateHigh: 0, risikoModerate: 0, risikoLowModerate: 0, risikoLow: 0 })
  })

  it('total is the sum of the five levels', () => {
    expect(riskBreakdownTotal(riskFieldsToBreakdown({ risikoHigh: 1, risikoModerateHigh: 2, risikoModerate: 3, risikoLowModerate: 4, risikoLow: 5 }))).toBe(15)
  })

  it('derives the legacy fields on save so a reload falls back consistently', () => {
    expect(legacyRiskFields({ risikoHigh: 1, risikoModerateHigh: 2, risikoModerate: 3, risikoLowModerate: 4, risikoLow: 5 }))
      .toEqual({ risikoTinggi: 3, risikoSedang: 3, risikoRendah: 9 })
    // Round trip of a legacy record: load fallback, then save
    const loaded = resolveRiskFields({ risikoTinggi: 3, risikoSedang: 2, risikoRendah: 1 })
    expect(legacyRiskFields(loaded)).toEqual({ risikoTinggi: 3, risikoSedang: 2, risikoRendah: 1 })
  })
})

describe('summarizeAtrFollowUp (overdue by due date)', () => {
  // 2026-05-10 10:00 in Asia/Jakarta (03:00 UTC)
  const now = new Date('2026-05-10T03:00:00Z')

  it('counts open ATRs with a due date before today (Jakarta) as overdue, ignoring is_overdue', () => {
    const rows = [
      { status: 'IN_PROGRESS', due_date: '2026-05-09T00:00:00Z' }, // overdue
      { status: 'PLANNED', due_date: '2026-05-01', is_overdue: false }, // overdue (flag ignored)
      { status: 'PENDING_REVIEW', due_date: '2026-05-10T00:00:00Z' }, // due today -> in progress
      { status: 'IN_PROGRESS', due_date: '2026-06-01', is_overdue: true }, // future -> in progress (flag ignored)
      { status: 'IN_PROGRESS' }, // no due date -> in progress
      { status: 'COMPLETED', due_date: '2026-01-01' }, // closed even if past due
      { status: 'CANCELLED', due_date: '2026-01-01' } // excluded
    ]
    expect(summarizeAtrFollowUp(rows, now)).toEqual({ closed: 1, inProgress: 3, overdue: 2 })
  })

  it('uses the Asia/Jakarta date, not UTC', () => {
    // 2026-05-09 18:00 UTC is already 2026-05-10 01:00 in Jakarta, so a 2026-05-09 due date is overdue
    const lateUtc = new Date('2026-05-09T18:00:00Z')
    expect(summarizeAtrFollowUp([{ status: 'IN_PROGRESS', due_date: '2026-05-09' }], lateUtc)).toEqual({ closed: 0, inProgress: 0, overdue: 1 })
    // 2026-05-09 16:00 UTC is 23:00 on 2026-05-09 in Jakarta: due today, not overdue yet
    const earlierUtc = new Date('2026-05-09T16:00:00Z')
    expect(summarizeAtrFollowUp([{ status: 'IN_PROGRESS', due_date: '2026-05-09' }], earlierUtc)).toEqual({ closed: 0, inProgress: 1, overdue: 0 })
  })

  it('normalises status spelling and returns zeros for no rows', () => {
    expect(summarizeAtrFollowUp([{ status: 'Completed' }, { status: 'canceled' }, { status: 'in progress', due_date: '2026-04-30' }], now))
      .toEqual({ closed: 1, inProgress: 0, overdue: 1 })
    expect(summarizeAtrFollowUp([], now)).toEqual({ closed: 0, inProgress: 0, overdue: 0 })
  })
})

describe('selectCompilationReports (no fallback)', () => {
  const reports = [
    { id: 'a', assignmentLetterId: 'ST-001/SKAI/2026', reportDate: '2026-02-15' },
    { id: 'b', assignmentLetterId: 'ST-002/SKAI/2026', reportDate: '2026-05-20' },
    { id: 'c', assignmentLetterId: 'ST-003/SKAI/2026', reportDate: '' }
  ]

  it('matches by assignment letter or by report-date quarter', () => {
    expect(selectCompilationReports(reports, { assignmentLetterId: 'ST-003/SKAI/2026', quarter: 1 }).map(r => r.id)).toEqual(['a', 'c'])
    expect(selectCompilationReports(reports, { assignmentLetterId: '', quarter: 2 }).map(r => r.id)).toEqual(['b'])
  })

  it('returns nothing when nothing matches instead of falling back to all LHAs', () => {
    expect(selectCompilationReports(reports, { assignmentLetterId: 'ST-999/SKAI/2026', quarter: 4 })).toEqual([])
    expect(selectCompilationReports(reports, { assignmentLetterId: 'All Assignment Letters', quarter: 3 })).toEqual([])
  })

  it('does not put reports without a date into quarter 1', () => {
    expect(quarterOfDate('')).toBeNull()
    expect(quarterOfDate(undefined)).toBeNull()
    expect(quarterOfDate('2026-12-01')).toBe(4)
    expect(selectCompilationReports([reports[2]!], { assignmentLetterId: '', quarter: 1 })).toEqual([])
  })
})

describe('useExecutiveSummarySources.fetchCompilationSources', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('reads the recommendation total from the summary endpoint in one call', async () => {
    const calls: { url: string, params?: Record<string, unknown> }[] = []
    vi.stubGlobal('$fetch', vi.fn(async (url: string, opts?: { params?: Record<string, unknown> }) => {
      calls.push({ url, params: opts?.params })
      if (url.endsWith('/fieldwork/test-controls/recommendations/summary')) {
        return { data: { total: 7, items: [{ auditResultReportId: 'lha-1', assignmentLetterId: 'ST-1', count: 7 }] } }
      }
      if (url.endsWith('/working-papers/risks')) {
        return { data: [
          { workingPaperId: 'ST-1', riskLevel: 'moderate-high' },
          { workingPaperId: 'ST-1', riskLevel: 'HIGH' },
          { workingPaperId: 'ST-other', riskLevel: 'high' }
        ] }
      }
      return { data: { items: [], pagination: { page: 1, page_size: 100, total: 0, total_pages: 1 } } }
    }))

    const { fetchCompilationSources } = useExecutiveSummarySources()
    const sources = await fetchCompilationSources({ assignmentLetterIds: ['ST-1'], lhaIds: ['lha-1', 'lha-2', 'lha-1'] })

    const summaryCalls = calls.filter(c => c.url.endsWith('/recommendations/summary'))
    expect(summaryCalls).toHaveLength(1)
    expect(summaryCalls[0]!.params).toEqual({ audit_result_report_ids: 'lha-1,lha-2' })
    expect(calls.some(c => c.url.endsWith('/fieldwork/test-controls'))).toBe(false)
    expect(sources?.recommendationCount).toBe(7)
    expect(sources?.riskBreakdown).toEqual({ high: 1, moderateHigh: 1, moderate: 0, lowModerate: 0, low: 0 })
  })
})
