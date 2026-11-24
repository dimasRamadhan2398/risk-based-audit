import { normalizeRiskLevel, PRIORITY_RISK_LEVELS, type PriorityRiskLevel } from '~/utils/riskLevelBadge'

interface SummaryActivity {
  auditorCount?: number | string | null
  totalMandays?: number | string | null
  riskLevel?: string | null
  riskName?: string | null
  auditUniverseId?: string | null
  months?: unknown
}

export interface AnnualPlanSummary {
  activityCount: number
  totalMandays: number
  totalAuditors: number
  /** distinct month numbers (0 = Jan), sorted */
  months: number[]
  /** activities with an Audit Universe link or a risk name */
  withRisk: number
  riskLevels: { level: PriorityRiskLevel, count: number }[]
}

const num = (v: unknown): number => {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}

const isMonth = (m: number) => Number.isInteger(m) && m >= 0 && m <= 11

/**
 * Compact roll-up of a plan's activities for the list: a plan can hold dozens of rows, so the
 * table shows these figures instead of one element per activity. Empty values are skipped.
 */
export const summarizeAnnualPlan = (
  plan: { activities?: SummaryActivity[] | null, selectedMonths?: number[] | null } | null | undefined,
  resolveRiskName: (a: SummaryActivity) => string = a => a.riskName || ''
): AnnualPlanSummary => {
  const activities = plan?.activities || []
  const months = new Set<number>((plan?.selectedMonths || []).map(Number).filter(isMonth))
  if (!months.size) {
    activities.forEach((a) => {
      if (Array.isArray(a.months)) a.months.map(Number).filter(isMonth).forEach(m => months.add(m))
    })
  }
  const counts = new Map<PriorityRiskLevel, number>()
  let withRisk = 0
  activities.forEach((a) => {
    if (a.auditUniverseId || resolveRiskName(a)) withRisk++
    const key = normalizeRiskLevel(a.riskLevel)
    if (key) counts.set(key, (counts.get(key) || 0) + 1)
  })
  return {
    activityCount: activities.length,
    totalMandays: activities.reduce((s, a) => s + num(a.totalMandays), 0),
    totalAuditors: activities.reduce((s, a) => s + num(a.auditorCount), 0),
    months: [...months].sort((a, b) => a - b),
    withRisk,
    riskLevels: PRIORITY_RISK_LEVELS.filter(l => counts.has(l)).map(level => ({ level, count: counts.get(level)! }))
  }
}
