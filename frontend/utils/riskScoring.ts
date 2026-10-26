// Risk Factor scoring helpers shared by the Risk Factors, Audit Priority and PKAT pages.
// The backend (risk-service) is the source of truth for risk_index / risk_level / audit_priority;
// the preview below only mirrors its formula while the user is editing scores.
import { getRiskLevelForIndex, isPrioritizedRiskLevel, type PriorityRiskLevel } from './riskLevelBadge'

export interface ScoredUniverseYear {
  /** ISO timestamp of the last scoring, null when the entity was never scored. Missing on older backends. */
  scored_at?: string | null
  risk_index?: number | string | null
  risk_level?: string | null
  audit_priority?: boolean | null
}

/**
 * Whether a yearly audit universe row has been scored.
 * Only an explicit `scored_at: null` means "never scored"; older responses without the field count as scored.
 */
export const isUniverseYearScored = (row?: ScoredUniverseYear | null): boolean =>
  !!row && row.scored_at !== null

/** Prioritized only when scored and the backend flagged it. */
export const isUniverseYearPrioritized = (row?: ScoredUniverseYear | null): boolean =>
  isUniverseYearScored(row) && !!row?.audit_priority

/** Scored rows first by risk index (desc), unscored rows after them in their original order. */
export const sortUniverseYearsByRisk = <T extends ScoredUniverseYear>(rows: T[] | null | undefined): T[] =>
  [...(rows || [])].sort((a, b) => {
    const scoredA = isUniverseYearScored(a)
    const scoredB = isUniverseYearScored(b)
    if (scoredA !== scoredB) return scoredA ? -1 : 1
    if (!scoredA) return 0
    return (Number(b.risk_index) || 0) - (Number(a.risk_index) || 0)
  })

// ─── Scoring rows ───────────────────────────────────────────────────

export interface CorporateFactorLike {
  id: string
  /** fraction, e.g. 0.15 */
  weight: number
  standard_risk_factor?: { name?: string } | null
}

export interface SavedRiskScore {
  corporate_risk_factor_id: string
  score: number
}

export interface ScoringRow<R = unknown> {
  corporate_risk_factor_id: string
  factor_name?: string
  /** fraction, e.g. 0.15 */
  weight: number
  /** null = not scored; never defaulted */
  score: number | null
  rubric?: R
}

/** One row per corporate factor; a factor without a saved score row stays `score: null`. */
export const buildScoringRows = <F extends CorporateFactorLike>(
  corporateFactors: F[] | null | undefined,
  savedScores: SavedRiskScore[] | null | undefined
): ScoringRow<F['standard_risk_factor']>[] =>
  (corporateFactors || []).map((cf) => {
    const match = (savedScores || []).find(s => s.corporate_risk_factor_id === cf.id)
    return {
      corporate_risk_factor_id: cf.id,
      factor_name: cf.standard_risk_factor?.name,
      weight: Number(cf.weight) || 0,
      score: match && Number.isFinite(Number(match.score)) ? Number(match.score) : null,
      rubric: cf.standard_risk_factor
    }
  })

const roundDecimals = (value: number, decimals: number): number => {
  const factor = 10 ** decimals
  return Math.round((Number(value) || 0) * factor) / factor
}

export interface RiskIndexPreview {
  /** Σ(weight × score) over scored factors */
  weightedSum: number
  /** weightedSum / (Σ weight of scored factors) / 5 × 100, rounded to 1 decimal */
  riskIndex: number
  riskLevel: PriorityRiskLevel
  auditPriority: boolean
  scoredCount: number
  totalCount: number
}

/**
 * Same formula as the backend (risk-service models/audit_universe_scoring.go): over scored factors only,
 * Σ(weight × score) / Σ(weight) / 5 × 100, rounded to 1 decimal before choosing the level. An unscored
 * factor is left out rather than counted as 0, so with every factor scored this is Σ(weight × score) / 5 × 100.
 */
export const computeRiskIndexPreview = (rows: Pick<ScoringRow, 'weight' | 'score'>[] | null | undefined): RiskIndexPreview => {
  const list = rows || []
  let weightedSum = 0
  let scoredWeight = 0
  let scoredCount = 0
  for (const row of list) {
    if (row.score === null || row.score === undefined) continue
    const weight = Number(row.weight) || 0
    weightedSum += weight * Number(row.score)
    scoredWeight += weight
    scoredCount++
  }
  const riskIndex = scoredWeight > 0 ? roundDecimals((weightedSum / scoredWeight / 5) * 100, 1) : 0
  const riskLevel = getRiskLevelForIndex(riskIndex)
  return {
    weightedSum,
    riskIndex,
    riskLevel,
    auditPriority: isPrioritizedRiskLevel(riskLevel),
    scoredCount,
    totalCount: list.length
  }
}

// ─── Corporate weights ──────────────────────────────────────────────

/** The backend accepts a weight total within this range (percent). */
export const WEIGHT_SUM_MIN = 99.9
export const WEIGHT_SUM_MAX = 100.1
export const WEIGHT_DECIMALS = 2

/** Normalises a weight percentage to at most 2 decimals. */
export const normalizeWeightPercent = (value: unknown): number => roundDecimals(Number(value) || 0, WEIGHT_DECIMALS)

/** Stored fraction (0.1234) to a percentage for the form (12.34). */
export const weightFractionToPercent = (fraction: number): number => roundDecimals((Number(fraction) || 0) * 100, WEIGHT_DECIMALS)

/** Sum of weight percentages, each normalised to 2 decimals, without floating point noise. */
export const sumWeightPercents = (weights: unknown[]): number =>
  roundDecimals(weights.reduce<number>((sum, w) => sum + normalizeWeightPercent(w), 0), WEIGHT_DECIMALS)

export const isValidWeightTotal = (total: number): boolean =>
  Number.isFinite(total) && total >= WEIGHT_SUM_MIN && total <= WEIGHT_SUM_MAX
