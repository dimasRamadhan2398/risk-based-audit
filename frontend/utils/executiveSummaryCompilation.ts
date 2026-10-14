/**
 * Executive summary compilation helpers shared by the compilation form, the executive summary store,
 * the list pages and the unit tests.
 *
 * Risk breakdown: five levels, the same as the working paper risks.
 *   high | moderate-high | moderate | low-moderate | low
 * Persisted as risikoHigh / risikoModerateHigh / risikoModerate / risikoLowModerate / risikoLow.
 * Records saved before the five-level breakdown only have risikoTinggi / risikoSedang / risikoRendah;
 * they are read as High / Moderate / Low (see resolveRiskFields).
 */

export const RISK_LEVELS = ['high', 'moderateHigh', 'moderate', 'lowModerate', 'low'] as const

export type RiskLevel = typeof RISK_LEVELS[number]

export type RiskBreakdown = Record<RiskLevel, number>

export type RiskField = 'risikoHigh' | 'risikoModerateHigh' | 'risikoModerate' | 'risikoLowModerate' | 'risikoLow'

export type RiskFields = Record<RiskField, number>

export interface LegacyRiskFields {
  risikoTinggi: number
  risikoSedang: number
  risikoRendah: number
}

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

const RISK_LEVEL_BY_KEY: Record<string, RiskLevel> = {
  'high': 'high',
  'critical': 'high',
  'tinggi': 'high',
  'moderate-high': 'moderateHigh',
  'moderate-to-high': 'moderateHigh',
  'moderate': 'moderate',
  'medium': 'moderate',
  'sedang': 'moderate',
  'low-moderate': 'lowModerate',
  'low-to-moderate': 'lowModerate',
  'low': 'low',
  'rendah': 'low'
}

/**
 * Working paper risk level -> breakdown level. Case-insensitive; spaces/underscores are read as '-'.
 * Legacy values: HIGH/CRITICAL/TINGGI -> high, MEDIUM/MODERATE/SEDANG -> moderate, LOW/RENDAH -> low.
 * Returns null for unknown or empty values so they are not counted.
 */
export const mapWorkingPaperRiskLevel = (raw: unknown): RiskLevel | null => {
  const key = String(raw ?? '').trim().toLowerCase().replace(/[\s_]+/g, '-')
  return RISK_LEVEL_BY_KEY[key] ?? null
}

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

/**
 * The five risk fields of a stored executive summary.
 * When all five are 0/missing (record saved before the five-level breakdown), falls back to the
 * legacy fields: Tinggi -> High, Sedang -> Moderate, Rendah -> Low.
 */
export const resolveRiskFields = (item: Partial<Record<RiskField | keyof LegacyRiskFields, unknown>>): RiskFields => {
  const breakdown = riskFieldsToBreakdown(item)
  if (riskBreakdownTotal(breakdown) > 0) return riskBreakdownToFields(breakdown)
  return riskBreakdownToFields({
    ...emptyRiskBreakdown(),
    high: toCount(item.risikoTinggi),
    moderate: toCount(item.risikoSedang),
    low: toCount(item.risikoRendah)
  })
}

/**
 * Legacy three-level fields derived from the five levels, sent on save so the old columns stay consistent:
 * Tinggi = High + Moderate to High, Sedang = Moderate, Rendah = Low to Moderate + Low.
 * (Inverse of the load fallback for records that only use High / Moderate / Low.)
 */
export const legacyRiskFields = (fields: Partial<Record<RiskField, unknown>>): LegacyRiskFields => {
  const b = riskFieldsToBreakdown(fields)
  return {
    risikoTinggi: b.high + b.moderateHigh,
    risikoSedang: b.moderate,
    risikoRendah: b.lowModerate + b.low
  }
}

/** Quarter (1-4) of a 'YYYY-MM-DD...' date, or null when the month cannot be read. */
export const quarterOfDate = (value: unknown): number | null => {
  const match = /^\d{4}-(\d{2})/.exec(String(value ?? '').trim())
  const month = match ? Number(match[1]) : 0
  if (month < 1 || month > 12) return null
  return Math.ceil(month / 3)
}

/**
 * Individual LHAs that belong to a compilation: same assignment letter, or report date in the target quarter.
 * No fallback: when nothing matches, the result is empty (the caller syncs nothing).
 */
export const selectCompilationReports = <T extends { assignmentLetterId?: string | null, reportDate?: string | null }>(
  reports: T[],
  target: { assignmentLetterId?: string | null, quarter: number }
): T[] => {
  const letter = target.assignmentLetterId && target.assignmentLetterId !== 'All Assignment Letters'
    ? target.assignmentLetterId
    : ''
  return reports.filter((report) => {
    if (letter && report.assignmentLetterId === letter) return true
    return quarterOfDate(report.reportDate) === target.quarter
  })
}
