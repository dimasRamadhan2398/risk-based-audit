/**
 * COSO 2013 internal control component assessment (docs/coso-2013-component-assessment-plan.md).
 * Domain types, the 5 components / 17 principles catalogue, the roll-up rules of section 4.3,
 * the display rules of section 5.1 and the role gating of section 7.2.
 * The backend (audit-service /api/v1/ic-assessments) is authoritative; these rules only drive the UI
 * and give early feedback. Response parsing lives in utils/icAssessmentParse.ts.
 */

// ============= Catalogue =============

export type ComponentCode = 'CE' | 'RA' | 'CA' | 'IC' | 'MA'

export const COSO_COMPONENT_CODES: ComponentCode[] = ['CE', 'RA', 'CA', 'IC', 'MA']

/** Principles per component (COSO 2013, fixed by the standard; mirrors the risk-service seed). */
export const COSO_PRINCIPLES_BY_COMPONENT: Record<ComponentCode, string[]> = {
  CE: ['P01', 'P02', 'P03', 'P04', 'P05'],
  RA: ['P06', 'P07', 'P08', 'P09'],
  CA: ['P10', 'P11', 'P12'],
  IC: ['P13', 'P14', 'P15'],
  MA: ['P16', 'P17']
}

export const COSO_PRINCIPLE_CODES: string[] = COSO_COMPONENT_CODES.flatMap(c => COSO_PRINCIPLES_BY_COMPONENT[c])

export const componentOfPrinciple = (code: string): ComponentCode | null =>
  COSO_COMPONENT_CODES.find(c => COSO_PRINCIPLES_BY_COMPONENT[c].includes(code)) ?? null

// ============= Enums =============

export type AssessmentStatus = 'DRAFT' | 'SUBMITTED' | 'APPROVED' | 'REOPENED'
export type ComponentStatus = 'YES' | 'PARTIAL' | 'NO'
export type PrincipleStatus = 'NOT_ASSESSED' | ComponentStatus | 'NOT_APPLICABLE'
export type Severity = 'NONE' | 'DEFICIENCY' | 'SIGNIFICANT' | 'MAJOR'
export type DerivedConclusion = 'EFFECTIVE' | 'NOT_EFFECTIVE' | 'REQUIRES_JUDGMENT' | 'INCOMPLETE'
export type FinalConclusion = 'EFFECTIVE' | 'NOT_EFFECTIVE'
export type EvidenceState = 'OK' | 'NO_DATA' | 'NOT_APPLICABLE' | 'SOURCE_UNAVAILABLE'

export const ASSESSMENT_STATUSES: AssessmentStatus[] = ['DRAFT', 'SUBMITTED', 'APPROVED', 'REOPENED']
export const COMPONENT_STATUSES: ComponentStatus[] = ['YES', 'PARTIAL', 'NO']
export const PRINCIPLE_STATUSES: PrincipleStatus[] = ['NOT_ASSESSED', 'YES', 'PARTIAL', 'NO', 'NOT_APPLICABLE']
export const SEVERITIES: Severity[] = ['NONE', 'DEFICIENCY', 'SIGNIFICANT', 'MAJOR']
export const FINAL_CONCLUSIONS: FinalConclusion[] = ['EFFECTIVE', 'NOT_EFFECTIVE']

const SEVERITY_RANK: Record<Severity, number> = { NONE: 0, DEFICIENCY: 1, SIGNIFICANT: 2, MAJOR: 3 }

export const severityRank = (s: Severity | null | undefined): number => (s ? SEVERITY_RANK[s] ?? 0 : 0)

export const maxSeverity = (list: (Severity | null | undefined)[]): Severity =>
  list.reduce<Severity>((acc, s) => (s && severityRank(s) > severityRank(acc) ? s : acc), 'NONE')

/** Minimum override reason length (section 4.3). */
export const OVERRIDE_REASON_MIN = 20

/** Tenant threshold range (decision Q4). */
export const THRESHOLD_MIN = 1
export const THRESHOLD_MAX = 365
export const DEFAULT_THRESHOLDS = { lhaDueDays: 30, managementResponseDueDays: 14 } as const

/** Fixed signal weights per component (mirrors audit-service ic_assessment/scoring.go SignalDefs; display only). */
export const SIGNAL_WEIGHTS: Record<ComponentCode, { key: string, weight: number }[]> = {
  CE: [
    { key: 'ce.integrity_pact', weight: 60 },
    { key: 'ce.unit_accountability', weight: 20 },
    { key: 'ce.audit_charter', weight: 20 }
  ],
  RA: [
    { key: 'ra.quarterly_coverage', weight: 50 },
    { key: 'ra.above_appetite_mitigated', weight: 50 }
  ],
  CA: [{ key: 'ca.rcm_effectiveness', weight: 100 }],
  IC: [
    { key: 'ic.lha_timeliness', weight: 50 },
    { key: 'ic.management_response', weight: 50 }
  ],
  MA: [
    { key: 'ma.atr_completed_on_time', weight: 50 },
    { key: 'ma.atr_not_overdue', weight: 50 }
  ]
}

/** Minimum share of a component's weight that must have data to get a proposal (scoring.go MinCoveragePct). */
export const MIN_COVERAGE_PCT = 50

// ============= Model =============

export interface IcActor {
  id: string
  name: string
}

export interface IcComponent {
  code: ComponentCode
  proposedScorePct: number | null
  proposedRating: number | null
  proposedStatus: ComponentStatus | null
  proposedSeverity: Severity | null
  proposalDataCoveragePct: number | null
  finalRating: number | null
  finalStatus: ComponentStatus | null
  finalSeverity: Severity | null
  derivedStatus: ComponentStatus | null
  derivedSeverity: Severity | null
  conclusion: string
  overrideReason: string
  isOverride: boolean
  reviewedBy: IcActor | null
  reviewedAt: string | null
}

export interface IcPrinciple {
  code: string
  componentCode: ComponentCode
  taggedControlCount: number
  taggedControlAvgPct: number | null
  status: PrincipleStatus
  severity: Severity
  rationale: string
  updatedBy: IcActor | null
  updatedAt: string | null
}

export interface IcEvidence {
  id: string
  proposalRun: number
  componentCode: ComponentCode
  signalKey: string
  principleHint: string
  sourceService: string
  sourceRef: string
  state: EvidenceState
  valuePct: number | null
  numerator: number | null
  denominator: number | null
  weight: number
  scopeNote: string
  detail: Record<string, unknown>
  collectedAt: string | null
}

export interface IcThresholds {
  lhaDueDays: number
  managementResponseDueDays: number
}

export interface IcSignature {
  image: string
  name: string
  title: string
  signedAt: string
}

export interface IcAssessment {
  id: string
  departmentId: string
  departmentCode: string
  departmentName: string
  locationId: string | null
  locationName: string
  year: number
  quarter: number
  periodStart: string | null
  periodEnd: string | null
  asOfDate: string | null
  status: AssessmentStatus
  version: number
  proposalRun: number
  derivedConclusion: DerivedConclusion | null
  finalConclusion: FinalConclusion | null
  operateTogether: boolean | null
  operateTogetherNote: string
  conclusionText: string
  conclusionJustification: string
  createdBy: IcActor | null
  createdAt: string | null
  submittedBy: IcActor | null
  submittedAt: string | null
  approvedBy: IcActor | null
  approvedByRole: string
  approvedAt: string | null
  approvalNote: string
  approvalSignature: IcSignature | null
  reopenCount: number
  lastReopenReason: string
  updatedAt: string | null
  components: IcComponent[]
  principles: IcPrinciple[]
  /** Evidence of the latest proposal run (detail response). */
  evidence: IcEvidence[]
  /** Threshold values the latest proposal used (frozen with the evidence), when the server sends them. */
  thresholdsUsed: IcThresholds | null
  /** Server-computed drift flag on the detail response, when present. */
  dataChangedSinceApproval: boolean | null
  /** Period lock as reported on the detail response, when present. */
  periodLocked: boolean | null
}

/** One list row: header plus a compact per-component result. */
export interface IcAssessmentListItem {
  id: string
  departmentId: string
  departmentCode: string
  departmentName: string
  locationId: string | null
  locationName: string
  year: number
  quarter: number
  status: AssessmentStatus
  derivedConclusion: DerivedConclusion | null
  finalConclusion: FinalConclusion | null
  components: Pick<IcComponent, 'code' | 'finalStatus' | 'finalRating' | 'derivedStatus'>[]
  updatedAt: string | null
}

export interface IcEvent {
  id: string
  action: string
  fromStatus: string
  toStatus: string
  actorName: string
  actorRoles: string[]
  reason: string
  before: Record<string, unknown> | null
  after: Record<string, unknown> | null
  createdAt: string | null
}

export interface IcPeriod {
  year: number
  quarter: number
  status: 'OPEN' | 'LOCKED'
  lockedBy: IcActor | null
  lockedAt: string | null
  lockNote: string
}

export interface IcDriftSignal {
  componentCode: ComponentCode | null
  signalKey: string
  /** null = the backend could not compare this signal (e.g. a source was unavailable). */
  changed: boolean | null
  approvedValuePct: number | null
  currentValuePct: number | null
  approvedState: EvidenceState | null
  currentState: EvidenceState | null
}

export interface IcDrift {
  changed: boolean
  /** false when some signals could not be re-collected, so "unchanged" is not conclusive. */
  complete: boolean | null
  checkedAt: string | null
  signals: IcDriftSignal[]
}

// ============= Period =============

/** True when the year (quarter 0) or the given quarter is locked. Locking a year also locks its quarters. */
export const isLockedFor = (periods: IcPeriod[], year: number, quarter: number): boolean =>
  periods.some(p => p.year === year && p.status === 'LOCKED' && (p.quarter === 0 || p.quarter === quarter))

// ============= Score display (section 5.1) =============

/** RCM rating bands (GetEffectivenessRating). Null in, null out: missing data is never a 0. */
export const ratingFromPct = (pct: number | null | undefined): number | null => {
  if (pct === null || pct === undefined || !Number.isFinite(pct)) return null
  if (pct >= 90) return 5
  if (pct >= 80) return 4
  if (pct >= 70) return 3
  if (pct >= 60) return 2
  return 1
}

export type ProposalView
  = | { kind: 'notProposed' }
    | { kind: 'insufficient', coveragePct: number | null, reasons: IcEvidence[] }
    | { kind: 'score', pct: number, rating: number | null, coveragePct: number | null, partial: boolean }

/**
 * What the component card shows for the system proposal.
 * No proposal run yet -> notProposed; a NULL score -> insufficient (with the signals that had no value);
 * otherwise the score. A NULL score is never rendered as 0.
 */
export const proposalView = (
  component: Pick<IcComponent, 'code' | 'proposedScorePct' | 'proposedRating' | 'proposalDataCoveragePct'>,
  proposalRun: number,
  evidence: IcEvidence[] = []
): ProposalView => {
  if (!proposalRun || proposalRun < 1) return { kind: 'notProposed' }
  const own = evidence.filter(e => e.componentCode === component.code)
  const withoutValue = own.filter(e => e.state !== 'OK' || e.valuePct === null)
  const pct = component.proposedScorePct
  if (pct === null || pct === undefined || !Number.isFinite(pct)) {
    return { kind: 'insufficient', coveragePct: component.proposalDataCoveragePct, reasons: withoutValue }
  }
  return {
    kind: 'score',
    pct,
    rating: component.proposedRating ?? ratingFromPct(pct),
    coveragePct: component.proposalDataCoveragePct,
    partial: withoutValue.length > 0
  }
}

/** "87.5 %" style text, or null when there is no value (callers show "Insufficient data" / "-"). */
export const formatPct = (pct: number | null | undefined, digits = 1): string | null => {
  if (pct === null || pct === undefined || !Number.isFinite(pct)) return null
  return `${Number(pct.toFixed(digits))} %`
}

// ============= Roll-up (section 4.3) =============

/** Severities allowed per principle status (server validation in 4.2.3). */
export const ALLOWED_SEVERITIES: Record<PrincipleStatus, Severity[]> = {
  NOT_ASSESSED: ['NONE'],
  YES: ['NONE', 'DEFICIENCY'],
  PARTIAL: ['DEFICIENCY', 'SIGNIFICANT'],
  NO: ['MAJOR'],
  NOT_APPLICABLE: ['NONE']
}

/** Default severity when the status changes and the old one is no longer allowed. */
export const coerceSeverity = (status: PrincipleStatus, severity: Severity): Severity => {
  const allowed = ALLOWED_SEVERITIES[status]
  return allowed.includes(severity) ? severity : allowed[0]!
}

export const principleNeedsRationale = (p: Pick<IcPrinciple, 'status' | 'severity'>): boolean =>
  p.status === 'PARTIAL' || p.status === 'NO' || p.status === 'NOT_APPLICABLE' || (p.severity !== 'NONE' && p.status !== 'NOT_ASSESSED')

export type PrincipleIssue = 'severityNotAllowed' | 'rationaleRequired'

export const principleIssues = (p: Pick<IcPrinciple, 'status' | 'severity' | 'rationale'>): PrincipleIssue[] => {
  const issues: PrincipleIssue[] = []
  if (!ALLOWED_SEVERITIES[p.status].includes(p.severity)) issues.push('severityNotAllowed')
  if (principleNeedsRationale(p) && !String(p.rationale || '').trim()) issues.push('rationaleRequired')
  return issues
}

export interface DerivedComponent {
  status: ComponentStatus | null
  severity: Severity
  incomplete: boolean
  applicableCount: number
}

/**
 * Component roll-up from its principles. NOT_APPLICABLE is ignored; any NOT_ASSESSED marks it incomplete.
 * status: NO if any principle is NO or MAJOR; YES if every applicable principle is YES; else PARTIAL
 * (a NOT_ASSESSED principle is not YES). With no assessed applicable principle the status is null.
 */
export const deriveComponent = (principles: Pick<IcPrinciple, 'status' | 'severity'>[]): DerivedComponent => {
  const applicable = principles.filter(p => p.status !== 'NOT_APPLICABLE')
  const assessed = applicable.filter(p => p.status !== 'NOT_ASSESSED')
  const incomplete = applicable.length !== assessed.length
  const severity = maxSeverity(assessed.map(p => p.severity))
  let status: ComponentStatus | null
  if (assessed.some(p => p.status === 'NO' || p.severity === 'MAJOR')) status = 'NO'
  else if (!assessed.length) status = null
  else if (applicable.every(p => p.status === 'YES')) status = 'YES'
  else status = 'PARTIAL'
  return { status, severity, incomplete, applicableCount: applicable.length }
}

/** Override = a final rating that differs from the proposal, or any final rating without a proposal. */
export const isOverride = (c: Pick<IcComponent, 'finalRating' | 'proposedRating'>): boolean =>
  c.finalRating !== null && c.finalRating !== undefined && (c.proposedRating === null || c.proposedRating === undefined || c.finalRating !== c.proposedRating)

export const overrideReasonMissing = (c: Pick<IcComponent, 'finalRating' | 'proposedRating' | 'overrideReason'>): boolean =>
  isOverride(c) && String(c.overrideReason || '').trim().length < OVERRIDE_REASON_MIN

export type ComponentIssue
  = | 'finalMissing'
    | 'overrideReasonRequired'
    | 'severityBelowDerived'
    | 'statusYesWhenDerivedNo'
    | 'majorNeedsLowRating'
    | 'yesNeedsHighRating'

export const componentIssues = (c: IcComponent, derived: DerivedComponent): ComponentIssue[] => {
  const issues: ComponentIssue[] = []
  if (c.finalRating === null || !c.finalStatus || !c.finalSeverity) issues.push('finalMissing')
  if (overrideReasonMissing(c)) issues.push('overrideReasonRequired')
  if (c.finalSeverity && severityRank(c.finalSeverity) < severityRank(derived.severity)) issues.push('severityBelowDerived')
  if (c.finalStatus === 'YES' && derived.status === 'NO') issues.push('statusYesWhenDerivedNo')
  if (c.finalSeverity === 'MAJOR' && c.finalRating !== null && c.finalRating > 2) issues.push('majorNeedsLowRating')
  if (c.finalStatus === 'YES' && c.finalRating !== null && c.finalRating < 3) issues.push('yesNeedsHighRating')
  return issues
}

const principlesOf = (principles: IcPrinciple[], code: ComponentCode) => principles.filter(p => p.componentCode === code)

export const deriveAllComponents = (principles: IcPrinciple[]): Record<ComponentCode, DerivedComponent> =>
  Object.fromEntries(COSO_COMPONENT_CODES.map(code => [code, deriveComponent(principlesOf(principles, code))])) as Record<ComponentCode, DerivedComponent>

/** Overall derived conclusion from the final component values and the principles. */
export const deriveConclusion = (components: IcComponent[], principles: IcPrinciple[]): DerivedConclusion => {
  const byCode = new Map(components.map(c => [c.code, c]))
  const all = COSO_COMPONENT_CODES.map(code => byCode.get(code))
  if (principles.some(p => p.status === 'NOT_ASSESSED')) return 'INCOMPLETE'
  if (all.some(c => !c || c.finalRating === null || !c.finalStatus || !c.finalSeverity)) return 'INCOMPLETE'
  const finals = all as IcComponent[]
  const anyMajor = finals.some(c => c.finalSeverity === 'MAJOR') || principles.some(p => p.severity === 'MAJOR')
  if (anyMajor || finals.some(c => c.finalStatus === 'NO')) return 'NOT_EFFECTIVE'
  if (finals.every(c => c.finalStatus === 'YES' && severityRank(c.finalSeverity) <= SEVERITY_RANK.DEFICIENCY)) return 'EFFECTIVE'
  return 'REQUIRES_JUDGMENT'
}

export type ConclusionIssue = 'conclusionMissing' | 'effectiveNotAllowed' | 'justificationRequired'

export const conclusionNeedsJustification = (derived: DerivedConclusion, final: FinalConclusion | null): boolean =>
  !!final && (derived === 'REQUIRES_JUDGMENT' || derived !== final)

export const conclusionIssues = (
  a: Pick<IcAssessment, 'components' | 'principles' | 'finalConclusion' | 'conclusionJustification'>
): ConclusionIssue[] => {
  const issues: ConclusionIssue[] = []
  const derived = deriveConclusion(a.components, a.principles)
  if (!a.finalConclusion) issues.push('conclusionMissing')
  const blocked = a.components.some(c => c.finalSeverity === 'MAJOR' || c.finalStatus === 'NO') || a.principles.some(p => p.severity === 'MAJOR')
  if (a.finalConclusion === 'EFFECTIVE' && blocked) issues.push('effectiveNotAllowed')
  if (conclusionNeedsJustification(derived, a.finalConclusion) && !String(a.conclusionJustification || '').trim()) issues.push('justificationRequired')
  return issues
}

export interface SubmitBlocker {
  key: string
  params?: Record<string, string | number>
}

/** Everything that stops a submit (section 7.1). Empty list = ready. Keys live under icAssessment.blockers.*. */
export const submitBlockers = (a: IcAssessment): SubmitBlocker[] => {
  const out: SubmitBlocker[] = []
  if (!a.proposalRun || a.proposalRun < 1) out.push({ key: 'noProposal' })
  const notAssessed = a.principles.filter(p => p.status === 'NOT_ASSESSED').length
  if (notAssessed) out.push({ key: 'principlesNotAssessed', params: { count: notAssessed } })
  const badPrinciples = a.principles.filter(p => principleIssues(p).length).length
  if (badPrinciples) out.push({ key: 'principleIssues', params: { count: badPrinciples } })
  const derived = deriveAllComponents(a.principles)
  for (const c of a.components) {
    for (const issue of componentIssues(c, derived[c.code])) out.push({ key: `component.${issue}`, params: { component: c.code } })
  }
  for (const issue of conclusionIssues(a)) out.push({ key: `conclusion.${issue}` })
  return out
}

// ============= Thresholds =============

export const thresholdValid = (n: unknown): boolean =>
  typeof n === 'number' && Number.isInteger(n) && n >= THRESHOLD_MIN && n <= THRESHOLD_MAX

// ============= Roles (section 7.2, decisions Q1 and Q6) =============

export type IcRole = 'ADMIN' | 'CHIEF_AUDIT_EXECUTIVE' | 'AUDIT_MANAGER' | 'AUDITOR' | 'DEPARTMENT_HEAD' | 'AUDITEE' | 'VIEWER'

// Mirrors audit-service roleAliases (services/action_taken_report/rbac.go). A bare MANAGER is not an audit manager.
const ROLE_ALIASES: Record<string, IcRole> = {
  ADMIN: 'ADMIN',
  ADMINISTRATOR: 'ADMIN',
  SUPERADMIN: 'ADMIN',
  CHIEFAUDITEXECUTIVE: 'CHIEF_AUDIT_EXECUTIVE',
  CAE: 'CHIEF_AUDIT_EXECUTIVE',
  EXECUTIVE: 'CHIEF_AUDIT_EXECUTIVE',
  KEPALASPI: 'CHIEF_AUDIT_EXECUTIVE',
  HEADOFSKAI: 'CHIEF_AUDIT_EXECUTIVE',
  HEADOFINTERNALAUDIT: 'CHIEF_AUDIT_EXECUTIVE',
  AUDITMANAGER: 'AUDIT_MANAGER',
  MANAGERAUDIT: 'AUDIT_MANAGER',
  AUDITOR: 'AUDITOR',
  AUDITSTAFF: 'AUDITOR',
  STAFFAUDIT: 'AUDITOR',
  LEADAUDITOR: 'AUDITOR',
  DEPARTMENTHEAD: 'DEPARTMENT_HEAD',
  AUDITEE: 'AUDITEE',
  VIEWER: 'VIEWER'
}

/** Canonical roles of the user, from the actual role list (role preview mode is ignored on purpose). */
export const icCanonicalRoles = (roles: unknown[]): Set<IcRole> => {
  const out = new Set<IcRole>()
  for (const r of roles) {
    const key = String(r ?? '').toUpperCase().replace(/[\s_\-\t]+/g, '')
    const c = ROLE_ALIASES[key]
    if (c) out.add(c)
  }
  return out
}

const ASSESSOR_ROLES: IcRole[] = ['ADMIN', 'CHIEF_AUDIT_EXECUTIVE', 'AUDIT_MANAGER', 'AUDITOR']
const APPROVER_ROLES: IcRole[] = ['ADMIN', 'CHIEF_AUDIT_EXECUTIVE']
const FULL_READER_ROLES: IcRole[] = [...ASSESSOR_ROLES, 'VIEWER']
const SCOPED_READER_ROLES: IcRole[] = ['DEPARTMENT_HEAD', 'AUDITEE']

export interface IcRoleCaps {
  /** Sees the list at all (nav item and pages). */
  canRead: boolean
  /** Sees every assessment and status; false for DEPARTMENT_HEAD / AUDITEE (own approved only). */
  fullRead: boolean
  canAssess: boolean
  canApproveRole: boolean
  canLockPeriod: boolean
  canViewEvents: boolean
  isAdmin: boolean
}

export const icRoleCaps = (roles: unknown[]): IcRoleCaps => {
  const set = icCanonicalRoles(roles)
  const has = (list: IcRole[]) => list.some(r => set.has(r))
  const fullRead = has(FULL_READER_ROLES)
  return {
    canRead: fullRead || has(SCOPED_READER_ROLES),
    fullRead,
    canAssess: has(ASSESSOR_ROLES),
    canApproveRole: has(APPROVER_ROLES),
    canLockPeriod: has(APPROVER_ROLES),
    canViewEvents: fullRead,
    isAdmin: set.has('ADMIN')
  }
}

export interface IcActions {
  edit: boolean
  propose: boolean
  submit: boolean
  returnToDraft: boolean
  approve: boolean
  reopen: boolean
  delete: boolean
  /** Shown instead of Approve when the approver is the submitter. */
  approveBlockedBySoD: boolean
}

/**
 * Workflow actions available to the user on one assessment (section 7.1).
 * Approve is never offered to the submitter; a locked period blocks every write.
 */
export const icActions = (
  caps: IcRoleCaps,
  userId: string,
  a: Pick<IcAssessment, 'status' | 'submittedBy' | 'submittedAt' | 'createdBy'> | null,
  periodLocked: boolean
): IcActions => {
  const none: IcActions = { edit: false, propose: false, submit: false, returnToDraft: false, approve: false, reopen: false, delete: false, approveBlockedBySoD: false }
  if (!a) return none
  const open = !periodLocked
  const working = a.status === 'DRAFT' || a.status === 'REOPENED'
  const isSubmitter = !!userId && !!a.submittedBy?.id && a.submittedBy.id === userId
  const isCreator = !!userId && !!a.createdBy?.id && a.createdBy.id === userId
  const approverOnSubmitted = caps.canApproveRole && a.status === 'SUBMITTED' && open
  return {
    edit: caps.canAssess && working && open,
    propose: caps.canAssess && working && open,
    submit: caps.canAssess && working && open,
    returnToDraft: approverOnSubmitted,
    approve: approverOnSubmitted && !isSubmitter,
    approveBlockedBySoD: approverOnSubmitted && isSubmitter,
    reopen: caps.canApproveRole && a.status === 'APPROVED' && open,
    delete: caps.canAssess && a.status === 'DRAFT' && !a.submittedAt && open && (isCreator || caps.isAdmin)
  }
}
