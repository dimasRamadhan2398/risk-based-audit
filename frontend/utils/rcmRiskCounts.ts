// Inherent / residual risk counts for the "Internal Control Effectiveness"
// card (RCM page and dashboard), read straight from the raw Corporate Risk
// Profile rows so the result does not depend on which period the CRP page has
// selected.
//
// - Inherent: risks of the year whose Q1 assessment is a priority level
//   (Moderate, Moderate to High, High). A Q1 pair that is missing or 0 falls
//   back to the risk register's inherent likelihood/impact.
// - Residual: inherent risks that are still a priority on the current calendar
//   quarter. A quarter pair that is missing or 0 falls back to the Q1 values
//   used for inherent (nothing reassessed yet = nothing mitigated yet).
// - Mitigation coverage: residual risks with at least one (non-deleted)
//   mitigation plan.

import { RiskLevel } from '~/types/risk'
import type { Quarter } from '~/utils/fiscalPeriod'
import { matchesBranchFilter } from '~/utils/rcmBranch'

export const RCM_ALL_DEPARTMENTS = 'All Departments'

const PRIORITY_LEVELS = new Set<string>([RiskLevel.MODERATE, RiskLevel.MODERATE_HIGH, RiskLevel.HIGH])

export type RiskLevelFn = (likelihood?: number, impact?: number) => RiskLevel | string

export interface RcmRiskCountOptions {
  year: number
  /** Calendar quarter residual is read from (normally currentQuarter()). */
  quarter: Quarter
  /** RCM_ALL_BRANCHES, RCM_UNASSIGNED_BRANCH or a Location master ID. */
  branch: string
  /** RCM_ALL_DEPARTMENTS or a branch/category name. */
  department: string
  getRiskLevel: RiskLevelFn
  /** Location master ID of a risk as read in the year (location_id, else matched branch name). */
  riskLocationId: (risk: any) => string | null
  /** Location master name for an ID, used to label the year's snapshot branch. */
  locationName?: (locationId: string) => string | undefined
}

export interface RcmRiskCounts {
  inherent: number
  residual: number
  inherentRiskIds: string[]
  residualRiskIds: string[]
}

export interface MitigationCoverage {
  covered: number
  total: number
}

const num = (v: unknown): number => {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}

/** A likelihood/impact pair, or null when either side is missing or 0. */
const pair = (likelihood: unknown, impact: unknown): { likelihood: number, impact: number } | null => {
  const l = num(likelihood)
  const i = num(impact)
  return l > 0 && i > 0 ? { likelihood: l, impact: i } : null
}

const isPriority = (getRiskLevel: RiskLevelFn, p: { likelihood: number, impact: number } | null) =>
  !!p && PRIORITY_LEVELS.has(String(getRiskLevel(p.likelihood, p.impact)))

/** Risk register inherent values: inherent_likelihood/impact, else the risk's own likelihood/impact. */
export const registerInherentPair = (risk: any) =>
  pair(risk?.inherent_likelihood, risk?.inherent_impact) ?? pair(risk?.likelihood, risk?.impact)

/**
 * Location/branch fields of a risk as read in the year, the same way the CRP
 * (mapRiskForYear) reads them: the assessment snapshot wins over the risk.
 */
const readForYear = (risk: any, assessment: any, locationName?: (id: string) => string | undefined) => ({
  ...risk,
  category: assessment.category || risk.category,
  location_id: assessment.location_id || risk.location_id,
  branch: assessment.location_id
    ? ((locationName && locationName(String(assessment.location_id))) ?? assessment.branch)
    : risk.branch
})

export const computeRcmRiskCounts = (rawRisks: any[] | null | undefined, opts: RcmRiskCountOptions): RcmRiskCounts => {
  const q = opts.quarter.toLowerCase()
  const inherentRiskIds: string[] = []
  const residualRiskIds: string[] = []

  for (const risk of rawRisks || []) {
    // Only risks assessed in the year belong to that year's CRP.
    const assessment = risk?.assessments?.find((a: any) => a?.year === opts.year)
    if (!assessment) continue

    const read = readForYear(risk, assessment, opts.locationName)
    if (!matchesBranchFilter(opts.riskLocationId(read), opts.branch)) continue
    if (opts.department && opts.department !== RCM_ALL_DEPARTMENTS
      && read.branch !== opts.department && read.category !== opts.department) continue

    const q1 = pair(assessment.likelihood_q1, assessment.impact_q1) ?? registerInherentPair(risk)
    if (!isPriority(opts.getRiskLevel, q1)) continue
    inherentRiskIds.push(String(risk.id))

    const current = pair(assessment[`likelihood_${q}`], assessment[`impact_${q}`]) ?? q1
    if (isPriority(opts.getRiskLevel, current)) residualRiskIds.push(String(risk.id))
  }

  return {
    inherent: inherentRiskIds.length,
    residual: residualRiskIds.length,
    inherentRiskIds,
    residualRiskIds
  }
}

const isDeletedPlan = (plan: any) =>
  !!(plan?.deleted_at || plan?.deletedAt || plan?.is_deleted || plan?.isDeleted)

/** Risk id a mitigation plan belongs to (API sends riskId; risk_id tolerated). */
const planRiskId = (plan: any): string => String(plan?.riskId ?? plan?.risk_id ?? '').trim().toLowerCase()

/** How many of the given risks have at least one non-deleted mitigation plan. */
export const computeMitigationCoverage = (riskIds: string[], plans: any[] | null | undefined): MitigationCoverage => {
  const withPlan = new Set<string>()
  for (const plan of plans || []) {
    if (isDeletedPlan(plan)) continue
    const id = planRiskId(plan)
    if (id) withPlan.add(id)
  }
  return {
    covered: riskIds.filter(id => withPlan.has(String(id).trim().toLowerCase())).length,
    total: riskIds.length
  }
}
