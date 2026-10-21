import { describe, it, expect, vi, afterEach } from 'vitest'
import { useExecutiveSummarySources } from '~/composables/useExecutiveSummarySources'
import {
  followUpRowsFromStats,
  formatSyncedAt,
  parseCompilationStats,
  riskBreakdownTotal
} from '~/utils/executiveSummaryCompilation'

// Body of GET /api/v1/executive-summary/compilation-stats (contract of the audit-service handler)
const statsBody = (overrides: Record<string, unknown> = {}, dataOverrides: Record<string, unknown> = {}) => ({
  success: true,
  message: 'Compilation statistics retrieved successfully',
  data: {
    year: 2026,
    quarter: 1,
    syncedAt: '2026-04-10T08:30:00+07:00',
    jumlahLaporan: 1,
    jumlahRekomendasi: 7,
    totalTemuan: 5,
    risikoHigh: 1,
    risikoModerateHigh: 1,
    risikoModerate: 2,
    risikoLowModerate: 0,
    risikoLow: 1,
    risikoLainnya: 0,
    followUpTable: [
      { status: 'Closed', jumlah: 2 },
      { status: 'In Progress', jumlah: 1 },
      { status: 'Overdue', jumlah: 1 }
    ],
    followUpCancelled: 0,
    ...dataOverrides
  },
  ...overrides
})

describe('parseCompilationStats', () => {
  it('maps one LHA response to the statistics', () => {
    const stats = parseCompilationStats(statsBody())
    expect(stats.year).toBe(2026)
    expect(stats.quarter).toBe(1)
    expect(stats.jumlahLaporan).toBe(1)
    expect(stats.jumlahRekomendasi).toBe(7)
    expect(stats.totalTemuan).toBe(5)
    expect(stats.riskBreakdown).toEqual({ high: 1, moderateHigh: 1, moderate: 2, lowModerate: 0, low: 1 })
    expect(riskBreakdownTotal(stats.riskBreakdown)).toBe(stats.totalTemuan)
    expect(stats.followUp).toEqual({ closed: 2, inProgress: 1, overdue: 1 })
  })

  it('accepts the empty case (0 LHA): all counts zero and three follow-up rows', () => {
    const stats = parseCompilationStats(statsBody({}, {
      jumlahLaporan: 0,
      jumlahRekomendasi: 0,
      totalTemuan: 0,
      risikoHigh: 0,
      risikoModerateHigh: 0,
      risikoModerate: 0,
      risikoLowModerate: 0,
      risikoLow: 0,
      followUpTable: [
        { status: 'Closed', jumlah: 0 },
        { status: 'In Progress', jumlah: 0 },
        { status: 'Overdue', jumlah: 0 }
      ]
    }))
    expect(stats.jumlahLaporan).toBe(0)
    expect(stats.totalTemuan).toBe(0)
    expect(stats.followUp).toEqual({ closed: 0, inProgress: 0, overdue: 0 })
  })

  it('accepts more than 20 LHA and keeps risikoLainnya out of totalTemuan', () => {
    const stats = parseCompilationStats(statsBody({}, {
      jumlahLaporan: 23,
      totalTemuan: 96,
      risikoHigh: 40,
      risikoModerateHigh: 20,
      risikoModerate: 20,
      risikoLowModerate: 10,
      risikoLow: 6,
      risikoLainnya: 3,
      followUpTable: [
        { status: 'Closed', jumlah: 50 },
        { status: 'In Progress', jumlah: 30 },
        { status: 'Overdue', jumlah: 16 }
      ],
      followUpCancelled: 4
    }))
    expect(stats.jumlahLaporan).toBe(23)
    expect(stats.totalTemuan).toBe(96)
    expect(riskBreakdownTotal(stats.riskBreakdown)).toBe(96)
    expect(stats.risikoLainnya).toBe(3)
    expect(stats.followUpCancelled).toBe(4)
  })

  it('keeps the quarter that was requested (0 means the whole year)', () => {
    expect(parseCompilationStats(statsBody({}, { quarter: 3 })).quarter).toBe(3)
    expect(parseCompilationStats(statsBody({}, { quarter: 0 })).quarter).toBe(0)
  })

  it('rejects a response with a missing or non-integer count, so no partial figures reach the form', () => {
    expect(() => parseCompilationStats(statsBody({}, { jumlahRekomendasi: undefined }))).toThrow(/jumlahRekomendasi/)
    expect(() => parseCompilationStats(statsBody({}, { totalTemuan: 1.5 }))).toThrow(/totalTemuan/)
    expect(() => parseCompilationStats(statsBody({}, { risikoLow: -1 }))).toThrow(/risikoLow/)
    expect(() => parseCompilationStats(statsBody({}, { risikoHigh: '2' }))).toThrow(/risikoHigh/)
  })

  it('rejects a follow-up table that is not exactly the three statuses', () => {
    expect(() => parseCompilationStats(statsBody({}, { followUpTable: [{ status: 'Closed', jumlah: 1 }] }))).toThrow(/followUpTable/)
    expect(() => parseCompilationStats(statsBody({}, {
      followUpTable: [{ status: 'Closed', jumlah: 1 }, { status: 'Closed', jumlah: 1 }, { status: 'Overdue', jumlah: 0 }]
    }))).toThrow(/followUpTable/)
    expect(() => parseCompilationStats(statsBody({}, {
      followUpTable: [{ status: 'constructor', jumlah: 1 }, { status: 'Closed', jumlah: 0 }, { status: 'Overdue', jumlah: 0 }]
    }))).toThrow(/followUpTable/)
  })

  it('rejects a missing body or an invalid syncedAt', () => {
    expect(() => parseCompilationStats(null)).toThrow()
    expect(() => parseCompilationStats({ success: true })).toThrow()
    expect(() => parseCompilationStats(statsBody({}, { syncedAt: 'not-a-date' }))).toThrow(/syncedAt/)
  })
})

describe('followUpRowsFromStats', () => {
  it('derives the percentages from the counts and keeps the user notes by status', () => {
    const rows = followUpRowsFromStats(
      { closed: 2, inProgress: 1, overdue: 1 },
      [
        { status: 'Closed', jumlah: 0, persentase: 0, keterangan: 'Catatan user' },
        { status: 'Overdue', jumlah: 0, persentase: 0, keterangan: '' }
      ]
    )
    expect(rows.map(r => r.status)).toEqual(['Closed', 'In Progress', 'Overdue'])
    expect(rows.map(r => r.jumlah)).toEqual([2, 1, 1])
    expect(rows.map(r => r.persentase)).toEqual([50, 25, 25])
    expect(rows.map(r => r.keterangan)).toEqual(['Catatan user', '', ''])
  })

  it('writes no generated note text and zero percentages when there are no follow-ups', () => {
    const rows = followUpRowsFromStats({ closed: 0, inProgress: 0, overdue: 0 })
    expect(rows).toEqual([
      { status: 'Closed', jumlah: 0, persentase: 0, keterangan: '' },
      { status: 'In Progress', jumlah: 0, persentase: 0, keterangan: '' },
      { status: 'Overdue', jumlah: 0, persentase: 0, keterangan: '' }
    ])
  })
})

describe('formatSyncedAt', () => {
  it('formats a valid RFC3339 timestamp and returns an empty string otherwise', () => {
    expect(formatSyncedAt('2026-04-10T08:30:00+07:00', 'id')).not.toBe('')
    expect(formatSyncedAt('2026-04-10T08:30:00+07:00', 'en')).not.toBe('')
    expect(formatSyncedAt(undefined)).toBe('')
    expect(formatSyncedAt('garbage')).toBe('')
  })
})

describe('useExecutiveSummarySources.fetchCompilationStats', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('calls the compilation-stats endpoint once with year and quarter', async () => {
    const calls: { url: string, params?: Record<string, unknown> }[] = []
    vi.stubGlobal('$fetch', vi.fn(async (url: string, opts?: { params?: Record<string, unknown> }) => {
      calls.push({ url, params: opts?.params })
      return statsBody()
    }))

    const { fetchCompilationStats, loading, error } = useExecutiveSummarySources()
    const pending = fetchCompilationStats({ year: 2026, quarter: 1 })
    expect(loading.value).toBe(true)
    const stats = await pending

    expect(calls).toHaveLength(1)
    expect(calls[0]!.url.endsWith('/executive-summary/compilation-stats')).toBe(true)
    expect(calls[0]!.params).toEqual({ year: 2026, quarter: 1 })
    expect(stats?.jumlahRekomendasi).toBe(7)
    expect(loading.value).toBe(false)
    expect(error.value).toBe('')
  })

  it('returns null and an error message on a network failure', async () => {
    vi.stubGlobal('$fetch', vi.fn(async () => {
      throw new Error('Network down')
    }))
    vi.spyOn(console, 'error').mockImplementation(() => {})

    const { fetchCompilationStats, loading, error } = useExecutiveSummarySources()
    const stats = await fetchCompilationStats({ year: 2026, quarter: 2 })

    expect(stats).toBeNull()
    expect(error.value).toBe('Network down')
    expect(loading.value).toBe(false)
  })

  it('returns null and an error when the response is malformed (no partial figures)', async () => {
    vi.stubGlobal('$fetch', vi.fn(async () => statsBody({}, { risikoLow: undefined })))
    vi.spyOn(console, 'error').mockImplementation(() => {})

    const { fetchCompilationStats, error } = useExecutiveSummarySources()
    expect(await fetchCompilationStats({ year: 2026, quarter: 1 })).toBeNull()
    expect(error.value).toContain('risikoLow')
  })

  it('ignores a superseded request so an older quarter cannot overwrite a newer one', async () => {
    let resolveFirst: (value: unknown) => void = () => {}
    vi.stubGlobal('$fetch', vi.fn((_url: string, opts?: { params?: Record<string, unknown> }) => {
      if (opts?.params?.quarter === 1) {
        return new Promise((resolve) => {
          resolveFirst = resolve
        })
      }
      return Promise.resolve(statsBody({}, { quarter: 2, jumlahLaporan: 4 }))
    }))

    const { fetchCompilationStats } = useExecutiveSummarySources()
    const first = fetchCompilationStats({ year: 2026, quarter: 1 })
    const second = await fetchCompilationStats({ year: 2026, quarter: 2 })
    resolveFirst(statsBody({}, { quarter: 1, jumlahLaporan: 9 }))

    expect(second?.quarter).toBe(2)
    expect(await first).toBeNull()
  })
})
