import type { AuditExecutionSources } from '~/types/audit'

/**
 * Single source of truth for Audit Execution progress and phase on the frontend.
 * Mirrors backend models.MapProgressToStatus (audit-service/models/audit-execution.go)
 * and the derived bands in audit-service/services/audit_execution_sync/derive.go:
 *
 *   0         Planning              no issued assignment letter, no fieldwork evidence
 *   1–25      Entry Meeting         assignment letter issued
 *   26–50     Fieldwork             fieldwork evidence, test controls being tested
 *   51–75     Draft Findings        findings + working paper action plans + mitigation (auditee follow-up)
 *   76–87     Reporting (drafting)  LHA exists in draft
 *   88–99     Reporting (approval)  LHA submitted, waiting review / CAE approval (~90)
 *   100       Completed             LHA approved / final only
 *
 * Progress is a whole number 0–100. The 51 stage boundary is the one in MapProgressToStatus;
 * derive.go starts its draft band at 55, but no derived value falls in 51–54.
 */

export type ExecutionSourceKey = 'annual_plan' | 'assignment_letter' | 'fieldwork' | 'working_paper' | 'result_report'

export type ExecutionStageStatus
  = | 'planning'
    | 'entry meeting'
    | 'fieldwork'
    | 'draft findings'
    | 'reporting'
    | 'completed'

export type ReportingSubStage = 'drafting' | 'approval'

/** Stage n (index 0..5) starts at EXECUTION_STAGE_MIN_PROGRESS[n]. Index-aligned with EXECUTION_STAGE_STATUSES. */
export const EXECUTION_STAGE_MIN_PROGRESS = [0, 1, 26, 51, 76, 100] as const

export const EXECUTION_STAGE_STATUSES: readonly ExecutionStageStatus[] = [
  'planning',
  'entry meeting',
  'fieldwork',
  'draft findings',
  'reporting',
  'completed'
]

/** Reporting progress from this value up is the LHA approval sub-stage (waiting approval). */
export const REPORTING_APPROVAL_MIN_PROGRESS = 88

/** Progress shown for an LHA that is waiting approval but not yet approved. */
export const LHA_WAITING_APPROVAL_PROGRESS = 90

export interface ExecutionSourceRow {
  key: ExecutionSourceKey
  done: boolean
  count: number
  tested: number
  total: number
  mitigations: number
  plans: number
  approved: boolean
}

export interface DraftBandRequirement {
  key: 'findings' | 'working_paper_plans' | 'mitigation'
  count: number
  done: boolean
}

export interface DraftBandSummary {
  /** Findings (working paper causes, or test controls with a finding when no causes are recorded). */
  findings: number
  /** Working paper action plans (`working_paper_plans`). */
  plans: number
  /** Auditee follow-up: fieldwork test controls with a mitigation plan (`mitigation_plans`). */
  mitigations: number
  /** The three requirements in the order the modal lists them. */
  requirements: DraftBandRequirement[]
  /** True when all three requirements have at least one record (the draft band is reachable). */
  complete: boolean
}

/**
 * Progress as a whole number from 0 to 100. Accepts numbers and numeric strings
 * (the API has sent both); anything that is not a number counts as 0.
 */
export const getProgressValue = (value: unknown): number => {
  const n = Number(value)
  if (!Number.isFinite(n)) return 0
  return Math.min(100, Math.max(0, Math.round(n)))
}

/**
 * Count from a source field. Missing, non-numeric or negative values count as 0,
 * so the UI never shows NaN.
 */
export const toSourceCount = (value: unknown): number => {
  const n = Number(value)
  return Number.isFinite(n) && n > 0 ? Math.floor(n) : 0
}

/** 0-based stage index (0 = planning … 5 = completed) for a progress value. */
export const getExecutionStageIndex = (progress: unknown): number => {
  const p = getProgressValue(progress)
  let index = 0
  for (let i = 0; i < EXECUTION_STAGE_MIN_PROGRESS.length; i++) {
    if (p >= EXECUTION_STAGE_MIN_PROGRESS[i]!) index = i
  }
  return index
}

/** Stage status for a progress value, e.g. 40 → 'fieldwork', 90 → 'reporting'. */
export const progressToStageStatus = (progress: unknown): ExecutionStageStatus =>
  EXECUTION_STAGE_STATUSES[getExecutionStageIndex(progress)]!

/** Sub-stage inside reporting: drafting (76–87) or approval (88–99). Null outside reporting. */
export const getReportingSubStage = (progress: unknown): ReportingSubStage | null => {
  const p = getProgressValue(progress)
  if (p < EXECUTION_STAGE_MIN_PROGRESS[4] || p >= 100) return null
  return p >= REPORTING_APPROVAL_MIN_PROGRESS ? 'approval' : 'drafting'
}

/**
 * Progress to display, reconciled with the LHA state in `sources`.
 * - Only an approved LHA can show 100. A not-approved LHA at 100 shows the waiting-approval value.
 * - An LHA in review below the approval band shows the waiting-approval value.
 * The check runs only when the flags are present as booleans, so older payloads keep their progress.
 */
export const getDisplayProgress = (progress: unknown, sources?: Partial<AuditExecutionSources> | null): number => {
  const p = getProgressValue(progress)
  if (!sources || typeof sources.result_report_approved !== 'boolean' || sources.result_report_approved) return p
  if (p >= 100) return LHA_WAITING_APPROVAL_PROGRESS
  if (sources.result_report_in_review === true && p < REPORTING_APPROVAL_MIN_PROGRESS) return LHA_WAITING_APPROVAL_PROGRESS
  return p
}

/**
 * One row per source module, in the order of the audit cycle, for the
 * "source data" section of the execution detail. Missing sources count as not done.
 */
export const describeExecutionSources = (s?: Partial<AuditExecutionSources> | null): ExecutionSourceRow[] => {
  const fieldwork = toSourceCount(s?.fieldwork_records)
  const workingPaper = toSourceCount(s?.working_paper_records)
  return [
    { key: 'annual_plan', done: !!s?.annual_plan, count: 0, tested: 0, total: 0, mitigations: 0, plans: 0, approved: false },
    { key: 'assignment_letter', done: !!s?.assignment_letter, count: 0, tested: 0, total: 0, mitigations: 0, plans: 0, approved: false },
    {
      key: 'fieldwork',
      done: fieldwork > 0,
      count: fieldwork,
      tested: toSourceCount(s?.test_controls_tested),
      total: toSourceCount(s?.test_controls),
      mitigations: toSourceCount(s?.mitigation_plans),
      plans: 0,
      approved: false
    },
    {
      key: 'working_paper',
      done: workingPaper > 0,
      count: workingPaper,
      tested: 0,
      total: 0,
      mitigations: 0,
      plans: toSourceCount(s?.working_paper_plans),
      approved: false
    },
    {
      key: 'result_report',
      done: !!s?.result_report,
      count: 0,
      tested: 0,
      total: 0,
      mitigations: 0,
      plans: 0,
      approved: !!s?.result_report_approved
    }
  ]
}

/**
 * The three requirements for the draft findings band (51–75%), with the counts the
 * modal shows. Mirrors draftReady in derive.go, which counts plans with an action and a
 * recommendation; `plans` here is the total number of working paper plans.
 */
export const describeDraftBand = (s?: Partial<AuditExecutionSources> | null): DraftBandSummary => {
  const causes = toSourceCount(s?.working_paper_causes)
  const findings = causes > 0 ? causes : toSourceCount(s?.test_controls_with_finding)
  const plans = toSourceCount(s?.working_paper_plans)
  const mitigations = toSourceCount(s?.mitigation_plans)
  const requirements: DraftBandRequirement[] = [
    { key: 'findings', count: findings, done: findings > 0 },
    { key: 'working_paper_plans', count: plans, done: plans > 0 },
    { key: 'mitigation', count: mitigations, done: mitigations > 0 }
  ]
  return {
    findings,
    plans,
    mitigations,
    requirements,
    complete: requirements.every(r => r.done)
  }
}
