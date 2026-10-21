/**
 * Executive summary compilation helpers shared by the compilation form, the executive summary store,
 * the list pages and the unit tests.
 *
 * Risk breakdown: five levels, the same as the working paper risks.
 *   high | moderate-high | moderate | low-moderate | low
 * Persisted as risikoHigh / risikoModerateHigh / risikoModerate / risikoLowModerate / risikoLow.
 * The risk data is calculated by the system; the legacy three-level fields are no longer sent.
 *
 * Statistics come from GET /executive-summary/compilation-stats (parseCompilationStats). The backend
 * does the counting; the frontend only maps the response onto the form fields.
 */

export const RISK_LEVELS = ['high', 'moderateHigh', 'moderate', 'lowModerate', 'low'] as const

export type RiskLevel = typeof RISK_LEVELS[number]

export type RiskBreakdown = Record<RiskLevel, number>

export type RiskField = 'risikoHigh' | 'risikoModerateHigh' | 'risikoModerate' | 'risikoLowModerate' | 'risikoLow'

export type RiskFields = Record<RiskField, number>

export const RISK_FIELD: Record<RiskLevel, RiskField> = {
  high: 'risikoHigh',
  moderateHigh: 'risikoModerateHigh',
  moderate: 'risikoModerate',
  lowModerate: 'risikoLowModerate',
  low: 'risikoLow'
}

/** i18n key per level ('executiveSummaryCompilation.riskLevels.<level>' / '.riskLevelsShort.<level>'). */
export const riskLevelI18nKey = (level: RiskLevel, short = false) =>
  `executiveSummaryCompilation.${short ? 'riskLevelsShort' : 'riskLevels'}.${level}`

/** Tailwind classes per level: accent border for inputs, chip for compact counts. */
export const RISK_LEVEL_STYLE: Record<RiskLevel, { border: string, chip: string }> = {
  high: { border: 'border-error-500', chip: 'text-error-600 bg-error-50 dark:bg-error-950/50' },
  moderateHigh: { border: 'border-orange-500', chip: 'text-orange-600 bg-orange-50 dark:bg-orange-950/50' },
  moderate: { border: 'border-warning-500', chip: 'text-warning-600 bg-warning-50 dark:bg-warning-950/50' },
  lowModerate: { border: 'border-lime-500', chip: 'text-lime-600 bg-lime-50 dark:bg-lime-950/50' },
  low: { border: 'border-success-500', chip: 'text-success-600 bg-success-50 dark:bg-success-950/50' }
}

export const emptyRiskBreakdown = (): RiskBreakdown => ({ high: 0, moderateHigh: 0, moderate: 0, lowModerate: 0, low: 0 })

export const riskBreakdownTotal = (breakdown: RiskBreakdown): number =>
  RISK_LEVELS.reduce((sum, level) => sum + breakdown[level], 0)

/**
 * Display buckets used by the compilation statistics (High / Medium / Low).
 * The five audit-universe levels stay the stored source; only the display is three buckets:
 *   High = High, Moderate to High | Medium = Moderate | Low = Low, Low to Moderate
 */
export type RiskBucket = 'high' | 'medium' | 'low'

export const RISK_BUCKET_BY_LEVEL: Record<RiskLevel, RiskBucket> = {
  high: 'high',
  moderateHigh: 'high',
  moderate: 'medium',
  lowModerate: 'low',
  low: 'low'
}

export const riskBucketBreakdown = (breakdown: RiskBreakdown): Record<RiskBucket, number> => {
  const buckets: Record<RiskBucket, number> = { high: 0, medium: 0, low: 0 }
  for (const level of RISK_LEVELS) {
    buckets[RISK_BUCKET_BY_LEVEL[level]] += breakdown[level]
  }
  return buckets
}

export const riskBucketTotal = (buckets: Record<RiskBucket, number>): number =>
  buckets.high + buckets.medium + buckets.low

const toCount = (value: unknown): number => {
  const n = Number(value)
  return Number.isFinite(n) && n > 0 ? n : 0
}

export const riskBreakdownToFields = (breakdown: RiskBreakdown): RiskFields => ({
  risikoHigh: breakdown.high,
  risikoModerateHigh: breakdown.moderateHigh,
  risikoModerate: breakdown.moderate,
  risikoLowModerate: breakdown.lowModerate,
  risikoLow: breakdown.low
})

export const riskFieldsToBreakdown = (fields: Partial<Record<RiskField, unknown>>): RiskBreakdown => ({
  high: toCount(fields.risikoHigh),
  moderateHigh: toCount(fields.risikoModerateHigh),
  moderate: toCount(fields.risikoModerate),
  lowModerate: toCount(fields.risikoLowModerate),
  low: toCount(fields.risikoLow)
})

export type FollowUpStatus = 'Closed' | 'In Progress' | 'Overdue'

export interface CompilationFollowUp {
  closed: number
  inProgress: number
  overdue: number
}

/** Statistics of GET /executive-summary/compilation-stats (data payload), already validated. */
export interface CompilationStats {
  year: number
  /** 1-4, or 0 for the whole year */
  quarter: number
  /** RFC3339 time at which the backend counted the figures */
  syncedAt: string
  jumlahLaporan: number
  jumlahRekomendasi: number
  /** Sum of the five risk levels (risikoLainnya is not part of it) */
  totalTemuan: number
  riskBreakdown: RiskBreakdown
  /** Findings with an empty or unrecognised risk level; shown as a note, not counted in totalTemuan */
  risikoLainnya: number
  followUp: CompilationFollowUp
  followUpCancelled: number
}

const FOLLOW_UP_STATUS_KEYS = new Map<string, keyof CompilationFollowUp>([
  ['Closed', 'closed'],
  ['In Progress', 'inProgress'],
  ['Overdue', 'overdue']
])

const COUNT_FIELDS = [
  'year',
  'quarter',
  'jumlahLaporan',
  'jumlahRekomendasi',
  'totalTemuan',
  'risikoHigh',
  'risikoModerateHigh',
  'risikoModerate',
  'risikoLowModerate',
  'risikoLow',
  'risikoLainnya',
  'followUpCancelled'
] as const

const isCount = (value: unknown): value is number =>
  typeof value === 'number' && Number.isInteger(value) && value >= 0

/**
 * Validates the body of GET /executive-summary/compilation-stats and maps it to CompilationStats.
 * Throws when a count is missing or not a non-negative integer, or when the follow-up table does not
 * have exactly the three statuses, so that a partial response is never written to the form.
 */
export const parseCompilationStats = (body: unknown): CompilationStats => {
  const data = (body as { data?: Record<string, unknown> } | null | undefined)?.data
  if (!data || typeof data !== 'object') throw new Error('Invalid compilation statistics response')

  for (const field of COUNT_FIELDS) {
    if (!isCount(data[field])) throw new Error(`Invalid compilation statistics field: ${field}`)
  }
  const syncedAt = data.syncedAt
  if (typeof syncedAt !== 'string' || Number.isNaN(Date.parse(syncedAt))) {
    throw new Error('Invalid compilation statistics field: syncedAt')
  }
  if (!Array.isArray(data.followUpTable)) throw new Error('Invalid compilation statistics field: followUpTable')

  const followUp: CompilationFollowUp = { closed: 0, inProgress: 0, overdue: 0 }
  const seen = new Set<string>()
  for (const row of data.followUpTable as { status?: unknown, jumlah?: unknown }[]) {
    const key = typeof row?.status === 'string' ? FOLLOW_UP_STATUS_KEYS.get(row.status) : undefined
    if (!key || !isCount(row.jumlah) || seen.has(row.status as string)) {
      throw new Error('Invalid compilation statistics field: followUpTable')
    }
    seen.add(row.status as string)
    followUp[key] = row.jumlah
  }
  if (seen.size !== FOLLOW_UP_STATUS_KEYS.size) throw new Error('Invalid compilation statistics field: followUpTable')

  const count = (field: typeof COUNT_FIELDS[number]) => data[field] as number
  return {
    year: count('year'),
    quarter: count('quarter'),
    syncedAt,
    jumlahLaporan: count('jumlahLaporan'),
    jumlahRekomendasi: count('jumlahRekomendasi'),
    totalTemuan: count('totalTemuan'),
    riskBreakdown: {
      high: count('risikoHigh'),
      moderateHigh: count('risikoModerateHigh'),
      moderate: count('risikoModerate'),
      lowModerate: count('risikoLowModerate'),
      low: count('risikoLow')
    },
    risikoLainnya: count('risikoLainnya'),
    followUp,
    followUpCancelled: count('followUpCancelled')
  }
}

/** Local date and time of an RFC3339 timestamp for the "last synced" label; '' when missing or invalid. */
export const formatSyncedAt = (iso: string | null | undefined, locale: string = 'id'): string => {
  if (!iso) return ''
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat(locale === 'en' ? 'en-GB' : 'id-ID', { dateStyle: 'medium', timeStyle: 'short' }).format(date)
}

export interface CompilationFollowUpRow {
  status: FollowUpStatus
  jumlah: number
  persentase: number
  keterangan: string
}

/**
 * Follow-up table rows from the counts. Percentages are derived from the counts; the notes (keterangan)
 * are kept from the previous rows of the same status and are never generated text.
 */
export const followUpRowsFromStats = (
  followUp: CompilationFollowUp,
  previous: readonly CompilationFollowUpRow[] = []
): CompilationFollowUpRow[] => {
  const total = followUp.closed + followUp.inProgress + followUp.overdue
  const statuses: [FollowUpStatus, number][] = [
    ['Closed', followUp.closed],
    ['In Progress', followUp.inProgress],
    ['Overdue', followUp.overdue]
  ]
  return statuses.map(([status, jumlah]) => ({
    status,
    jumlah,
    persentase: total > 0 ? (jumlah / total) * 100 : 0,
    keterangan: previous.find(row => row.status === status)?.keterangan ?? ''
  }))
}
