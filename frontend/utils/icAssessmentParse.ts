/**
 * The ONE place that knows the wire format of audit-service /api/v1/ic-assessments
 * (docs/coso-2013-component-assessment-plan.md section 8.2). Everything else uses the camelCase
 * model in utils/icAssessment.ts. When the backend (task B3) lands, align field names here only.
 *
 * Assumptions (snake_case, `{success, data}` envelope, decimals possibly sent as strings):
 * - actors come as `<prefix>_by_id` + `<prefix>_by_name`, or a nested `<prefix>_by: {id, name}`.
 * - list: `data` is an array, or `{items|assessments: [...], pagination|total}`.
 * - evidence: `data` is an array, or `{run, thresholds, items|evidence: [...]}`.
 */
import {
  COSO_COMPONENT_CODES,
  componentOfPrinciple,
  type ComponentCode,
  type ComponentStatus,
  type DerivedConclusion,
  type EvidenceState,
  type FinalConclusion,
  type IcActor,
  type IcAssessment,
  type IcAssessmentListItem,
  type IcComponent,
  type IcDrift,
  type IcEvent,
  type IcEvidence,
  type IcPeriod,
  type IcPrinciple,
  type IcSignature,
  type IcThresholds,
  type PrincipleStatus,
  type Severity,
  type AssessmentStatus
} from '~/utils/icAssessment'

type Raw = Record<string, any>

const isObj = (v: unknown): v is Raw => !!v && typeof v === 'object' && !Array.isArray(v)

/** Unwraps `{success, data}`; returns the body itself when there is no envelope. */
export const unwrap = (res: unknown): any => (isObj(res) && 'data' in res ? res.data : res)

const str = (v: unknown): string => (v === null || v === undefined ? '' : String(v))
const strOrNull = (v: unknown): string | null => (v === null || v === undefined || v === '' ? null : String(v))

/** Number or null. Never turns a missing value into 0. */
export const numOrNull = (v: unknown): number | null => {
  if (v === null || v === undefined || v === '') return null
  const n = typeof v === 'number' ? v : Number(v)
  return Number.isFinite(n) ? n : null
}
const int = (v: unknown, fallback = 0): number => {
  const n = numOrNull(v)
  return n === null ? fallback : Math.trunc(n)
}
const boolOrNull = (v: unknown): boolean | null => (typeof v === 'boolean' ? v : v === null || v === undefined ? null : v === 'true' ? true : v === 'false' ? false : null)

const pick = (raw: Raw, ...keys: string[]): unknown => {
  for (const k of keys) if (raw[k] !== undefined) return raw[k]
  return undefined
}

const upperEnum = <T extends string>(v: unknown, allowed: readonly T[]): T | null => {
  const s = str(v).trim().toUpperCase()
  return (allowed as readonly string[]).includes(s) ? (s as T) : null
}

const COMPONENT_STATUS = ['YES', 'PARTIAL', 'NO'] as const
const PRINCIPLE_STATUS = ['NOT_ASSESSED', 'YES', 'PARTIAL', 'NO', 'NOT_APPLICABLE'] as const
const SEVERITY = ['NONE', 'DEFICIENCY', 'SIGNIFICANT', 'MAJOR'] as const
const DERIVED = ['EFFECTIVE', 'NOT_EFFECTIVE', 'REQUIRES_JUDGMENT', 'INCOMPLETE'] as const
const FINAL = ['EFFECTIVE', 'NOT_EFFECTIVE'] as const
const STATUS = ['DRAFT', 'SUBMITTED', 'APPROVED', 'REOPENED'] as const
const EVIDENCE_STATE = ['OK', 'NO_DATA', 'NOT_APPLICABLE', 'SOURCE_UNAVAILABLE'] as const

const componentCode = (v: unknown): ComponentCode | null => upperEnum(v, COSO_COMPONENT_CODES)

const actor = (raw: Raw, prefix: string): IcActor | null => {
  const nested = raw[`${prefix}_by`]
  if (isObj(nested)) {
    const id = str(nested.id ?? nested.user_id)
    return id || nested.name ? { id, name: str(nested.name ?? nested.full_name) } : null
  }
  const id = str(pick(raw, `${prefix}_by_id`, `${prefix}_by_user_id`))
  const name = str(pick(raw, `${prefix}_by_name`))
  return id || name ? { id, name } : null
}

const jsonObj = (v: unknown): Record<string, unknown> | null => {
  if (isObj(v)) return v
  if (typeof v === 'string' && v.trim().startsWith('{')) {
    try {
      const parsed = JSON.parse(v)
      return isObj(parsed) ? parsed : null
    } catch {
      return null
    }
  }
  return null
}

export const parseThresholds = (raw: unknown): IcThresholds | null => {
  if (!isObj(raw)) return null
  const lha = numOrNull(pick(raw, 'lha_due_days', 'lhaDueDays'))
  const mr = numOrNull(pick(raw, 'management_response_due_days', 'managementResponseDueDays'))
  if (lha === null && mr === null) return null
  return { lhaDueDays: lha ?? 0, managementResponseDueDays: mr ?? 0 }
}

export const parseComponent = (raw: Raw): IcComponent | null => {
  const code = componentCode(pick(raw, 'component_code', 'code'))
  if (!code) return null
  const finalRating = numOrNull(raw.final_rating)
  const proposedRating = numOrNull(raw.proposed_rating)
  return {
    code,
    proposedScorePct: numOrNull(raw.proposed_score_pct),
    proposedRating,
    proposedStatus: upperEnum<ComponentStatus>(raw.proposed_status, COMPONENT_STATUS),
    proposedSeverity: upperEnum<Severity>(raw.proposed_severity, SEVERITY),
    proposalDataCoveragePct: numOrNull(raw.proposal_data_coverage_pct),
    finalRating,
    finalStatus: upperEnum<ComponentStatus>(raw.final_status, COMPONENT_STATUS),
    finalSeverity: upperEnum<Severity>(raw.final_severity, SEVERITY),
    derivedStatus: upperEnum<ComponentStatus>(raw.derived_status, COMPONENT_STATUS),
    derivedSeverity: upperEnum<Severity>(raw.derived_severity, SEVERITY),
    conclusion: str(raw.conclusion),
    overrideReason: str(raw.override_reason),
    isOverride: typeof raw.is_override === 'boolean' ? raw.is_override : finalRating !== null && finalRating !== proposedRating,
    reviewedBy: actor(raw, 'reviewed'),
    reviewedAt: strOrNull(raw.reviewed_at)
  }
}

export const parsePrinciple = (raw: Raw): IcPrinciple | null => {
  const code = str(pick(raw, 'principle_code', 'code')).toUpperCase()
  const comp = componentCode(raw.component_code) ?? componentOfPrinciple(code)
  if (!code || !comp) return null
  return {
    code,
    componentCode: comp,
    taggedControlCount: int(raw.tagged_control_count),
    taggedControlAvgPct: numOrNull(raw.tagged_control_avg_pct),
    status: upperEnum<PrincipleStatus>(raw.status, PRINCIPLE_STATUS) ?? 'NOT_ASSESSED',
    severity: upperEnum<Severity>(raw.severity, SEVERITY) ?? 'NONE',
    rationale: str(raw.rationale),
    updatedBy: actor(raw, 'updated'),
    updatedAt: strOrNull(raw.updated_at)
  }
}

export const parseEvidence = (raw: Raw): IcEvidence | null => {
  const comp = componentCode(raw.component_code)
  if (!comp) return null
  return {
    id: str(raw.id),
    proposalRun: int(raw.proposal_run),
    componentCode: comp,
    signalKey: str(raw.signal_key),
    principleHint: str(raw.principle_hint),
    sourceService: str(raw.source_service),
    sourceRef: str(raw.source_ref),
    state: upperEnum<EvidenceState>(raw.state, EVIDENCE_STATE) ?? 'NO_DATA',
    valuePct: numOrNull(raw.value_pct),
    numerator: numOrNull(raw.numerator),
    denominator: numOrNull(raw.denominator),
    weight: numOrNull(raw.weight) ?? 0,
    scopeNote: str(raw.scope_note),
    detail: jsonObj(raw.detail) ?? {},
    collectedAt: strOrNull(raw.collected_at)
  }
}

const list = <T>(v: unknown, fn: (r: Raw) => T | null): T[] =>
  Array.isArray(v) ? v.filter(isObj).map(fn).filter((x): x is T => x !== null) : []

const parseSignature = (raw: Raw): IcSignature | null => {
  const sig = jsonObj(raw.approval_signature) ?? jsonObj(raw.signature)
  const image = str(sig?.image ?? raw.approval_signature_image)
  if (!image) return null
  return {
    image,
    name: str(sig?.name ?? raw.approval_signer_name),
    title: str(sig?.title ?? raw.approval_signer_title),
    signedAt: str(sig?.signed_at ?? raw.approval_signed_at)
  }
}

/** Header fields shared by detail and list rows. */
const parseHeader = (raw: Raw) => ({
  id: str(raw.id),
  departmentId: str(raw.department_id),
  departmentCode: str(raw.department_code),
  departmentName: str(raw.department_name),
  locationId: strOrNull(raw.location_id),
  locationName: str(raw.location_name),
  year: int(raw.year),
  quarter: int(raw.quarter),
  status: upperEnum<AssessmentStatus>(raw.status, STATUS) ?? 'DRAFT',
  derivedConclusion: upperEnum<DerivedConclusion>(raw.derived_conclusion, DERIVED),
  finalConclusion: upperEnum<FinalConclusion>(raw.final_conclusion, FINAL),
  updatedAt: strOrNull(raw.updated_at)
})

/** Ensures the 5 components appear in catalogue order (missing ones as empty rows). */
const completeComponents = (parsed: IcComponent[]): IcComponent[] => {
  const byCode = new Map(parsed.map(c => [c.code, c]))
  return COSO_COMPONENT_CODES.map(code => byCode.get(code) ?? emptyComponent(code))
}

export const emptyComponent = (code: ComponentCode): IcComponent => ({
  code,
  proposedScorePct: null,
  proposedRating: null,
  proposedStatus: null,
  proposedSeverity: null,
  proposalDataCoveragePct: null,
  finalRating: null,
  finalStatus: null,
  finalSeverity: null,
  derivedStatus: null,
  derivedSeverity: null,
  conclusion: '',
  overrideReason: '',
  isOverride: false,
  reviewedBy: null,
  reviewedAt: null
})

export const parseAssessment = (res: unknown): IcAssessment | null => {
  const body = unwrap(res)
  const raw: Raw | null = isObj(body) ? (isObj(body.assessment) ? { ...body, ...body.assessment } : body) : null
  if (!raw || !raw.id) return null
  const evidenceRaw = raw.evidence ?? raw.current_evidence
  const evidence = isObj(evidenceRaw) ? parseEvidenceList(evidenceRaw).items : list(evidenceRaw, parseEvidence)
  const thresholds = parseThresholds(raw.thresholds_used ?? raw.thresholds)
    ?? (isObj(evidenceRaw) ? parseEvidenceList(evidenceRaw).thresholds : null)
  return {
    ...parseHeader(raw),
    periodStart: strOrNull(raw.period_start),
    periodEnd: strOrNull(raw.period_end),
    asOfDate: strOrNull(raw.as_of_date),
    version: int(raw.version, 1),
    proposalRun: int(raw.proposal_run),
    operateTogether: boolOrNull(raw.operate_together),
    operateTogetherNote: str(raw.operate_together_note),
    conclusionText: str(raw.conclusion_text),
    conclusionJustification: str(raw.conclusion_justification),
    createdBy: actor(raw, 'created'),
    createdAt: strOrNull(raw.created_at),
    submittedBy: actor(raw, 'submitted'),
    submittedAt: strOrNull(raw.submitted_at),
    approvedBy: actor(raw, 'approved'),
    approvedByRole: str(raw.approved_by_role),
    approvedAt: strOrNull(raw.approved_at),
    approvalNote: str(raw.approval_note),
    approvalSignature: parseSignature(raw),
    reopenCount: int(raw.reopen_count),
    lastReopenReason: str(raw.last_reopen_reason),
    components: completeComponents(list(raw.components, parseComponent)),
    principles: list(raw.principles, parsePrinciple),
    evidence,
    thresholdsUsed: thresholds,
    dataChangedSinceApproval: boolOrNull(pick(raw, 'data_changed_since_approval', 'source_changed_since_approval')),
    periodLocked: boolOrNull(raw.period_locked)
  }
}

const parseListItem = (raw: Raw): IcAssessmentListItem | null => {
  if (!raw.id) return null
  return {
    ...parseHeader(raw),
    components: list(raw.components, parseComponent).map(c => ({
      code: c.code,
      finalStatus: c.finalStatus,
      finalRating: c.finalRating,
      derivedStatus: c.derivedStatus
    }))
  }
}

export const parseAssessmentList = (res: unknown): { items: IcAssessmentListItem[], total: number } => {
  const body = unwrap(res)
  const rows = Array.isArray(body) ? body : isObj(body) ? (body.items ?? body.assessments ?? body.data ?? []) : []
  const items = list(rows, parseListItem)
  const total = isObj(body) ? int(body.pagination?.total ?? body.total, items.length) : items.length
  return { items, total }
}

export const parseEvidenceList = (res: unknown): { run: number | null, thresholds: IcThresholds | null, items: IcEvidence[] } => {
  const body = unwrap(res)
  if (Array.isArray(body)) {
    const items = list(body, parseEvidence)
    return { run: items[0]?.proposalRun ?? null, thresholds: null, items }
  }
  if (!isObj(body)) return { run: null, thresholds: null, items: [] }
  const items = list(body.items ?? body.evidence ?? [], parseEvidence)
  return {
    run: numOrNull(body.run ?? body.proposal_run) ?? items[0]?.proposalRun ?? null,
    thresholds: parseThresholds(body.thresholds_used ?? body.thresholds),
    items
  }
}

export const parseEvents = (res: unknown): IcEvent[] => {
  const body = unwrap(res)
  const rows = Array.isArray(body) ? body : isObj(body) ? (body.items ?? body.events ?? []) : []
  return list(rows, (raw) => {
    const roles = raw.actor_roles
    return {
      id: str(raw.id),
      action: str(raw.action).toUpperCase(),
      fromStatus: str(raw.from_status),
      toStatus: str(raw.to_status),
      actorName: str(raw.actor_name),
      actorRoles: Array.isArray(roles) ? roles.map(String) : str(roles).split(',').map(s => s.trim()).filter(Boolean),
      reason: str(raw.reason),
      before: jsonObj(raw.before),
      after: jsonObj(raw.after),
      createdAt: strOrNull(raw.created_at)
    }
  })
}

export const parsePeriods = (res: unknown): IcPeriod[] => {
  const body = unwrap(res)
  const rows = Array.isArray(body) ? body : isObj(body) ? (body.items ?? body.periods ?? []) : []
  return list(rows, raw => ({
    year: int(raw.year),
    quarter: int(raw.quarter),
    status: str(raw.status).toUpperCase() === 'LOCKED' ? 'LOCKED' : 'OPEN',
    lockedBy: actor(raw, 'locked'),
    lockedAt: strOrNull(raw.locked_at),
    lockNote: str(raw.lock_note)
  }))
}

/**
 * GET /ic-assessments/:id/drift: `{ changed, complete, checked_at, signals: [{ signal_key, component_code,
 * changed: bool|null, frozen: { state, value_pct, ... }, current: { state, value_pct, ... } }] }`.
 * "frozen" is the evidence stored at approval, "current" a fresh read-only run of the collectors.
 */
export const parseDrift = (res: unknown): IcDrift | null => {
  const body = unwrap(res)
  if (!isObj(body)) return null
  const changed = boolOrNull(body.changed)
  if (changed === null) return null
  const side = (v: unknown): Raw => (isObj(v) ? v : {})
  return {
    changed,
    complete: boolOrNull(body.complete),
    checkedAt: strOrNull(body.checked_at),
    signals: list(body.signals ?? [], (raw) => {
      const frozen = side(raw.frozen)
      const current = side(raw.current)
      return {
        componentCode: componentCode(raw.component_code),
        signalKey: str(raw.signal_key),
        changed: boolOrNull(raw.changed),
        approvedValuePct: numOrNull(frozen.value_pct),
        currentValuePct: numOrNull(current.value_pct),
        approvedState: upperEnum<EvidenceState>(frozen.state, EVIDENCE_STATE),
        currentState: upperEnum<EvidenceState>(current.state, EVIDENCE_STATE)
      }
    })
  }
}

// ============= Requests =============

const COMPONENT_FIELDS: [keyof IcComponent, string][] = [
  ['finalRating', 'final_rating'],
  ['finalStatus', 'final_status'],
  ['finalSeverity', 'final_severity'],
  ['conclusion', 'conclusion'],
  ['overrideReason', 'override_reason']
]
const PRINCIPLE_FIELDS: [keyof IcPrinciple, string][] = [
  ['status', 'status'],
  ['severity', 'severity'],
  ['rationale', 'rationale']
]
const HEADER_FIELDS: [keyof IcAssessment, string][] = [
  ['finalConclusion', 'final_conclusion'],
  ['conclusionText', 'conclusion_text'],
  ['conclusionJustification', 'conclusion_justification']
]

export interface IcUpdatePayload {
  version: number
  components?: Record<string, unknown>[]
  principles?: Record<string, unknown>[]
  final_conclusion?: FinalConclusion | null
  conclusion_text?: string
  conclusion_justification?: string
}

/**
 * Body of PUT /ic-assessments/:id with only what changed (the backend update is partial):
 * changed components by component_code and changed principles by principle_code, each with only
 * its changed fields, plus the changed conclusion fields. Null when nothing changed.
 */
export const buildUpdatePayload = (original: IcAssessment, draft: IcAssessment): IcUpdatePayload | null => {
  const payload: IcUpdatePayload = { version: original.version }
  let changed = false

  const comps: Record<string, unknown>[] = []
  for (const d of draft.components) {
    const o = original.components.find(c => c.code === d.code)
    const row: Record<string, unknown> = { component_code: d.code }
    for (const [k, wire] of COMPONENT_FIELDS) if (!o || o[k] !== d[k]) row[wire] = d[k]
    if (Object.keys(row).length > 1) comps.push(row)
  }
  if (comps.length) {
    payload.components = comps
    changed = true
  }

  const prins: Record<string, unknown>[] = []
  for (const d of draft.principles) {
    const o = original.principles.find(p => p.code === d.code)
    const row: Record<string, unknown> = { principle_code: d.code }
    for (const [k, wire] of PRINCIPLE_FIELDS) if (!o || o[k] !== d[k]) row[wire] = d[k]
    if (Object.keys(row).length > 1) prins.push(row)
  }
  if (prins.length) {
    payload.principles = prins
    changed = true
  }

  for (const [k, wire] of HEADER_FIELDS) {
    if (original[k] !== draft[k]) {
      ;(payload as unknown as Record<string, unknown>)[wire] = draft[k]
      changed = true
    }
  }
  return changed ? payload : null
}

export interface IcApprovePayload {
  version: number
  operate_together: boolean
  operate_together_note?: string
  approval_note?: string
  signature?: { image: string, name: string, title: string, signed_at: string }
}

export const buildApprovePayload = (input: {
  version: number
  operateTogether: boolean
  operateTogetherNote?: string
  approvalNote?: string
  signature?: IcSignature | null
}): IcApprovePayload => {
  const body: IcApprovePayload = { version: input.version, operate_together: input.operateTogether }
  if (input.operateTogetherNote?.trim()) body.operate_together_note = input.operateTogetherNote.trim()
  if (input.approvalNote?.trim()) body.approval_note = input.approvalNote.trim()
  if (input.signature?.image) {
    body.signature = {
      image: input.signature.image,
      name: input.signature.name.trim(),
      title: input.signature.title.trim(),
      signed_at: input.signature.signedAt
    }
  }
  return body
}

export const buildThresholdsPayload = (original: IcThresholds | null, next: IcThresholds): Record<string, number> | null => {
  const body: Record<string, number> = {}
  if (!original || original.lhaDueDays !== next.lhaDueDays) body.lha_due_days = next.lhaDueDays
  if (!original || original.managementResponseDueDays !== next.managementResponseDueDays) body.management_response_due_days = next.managementResponseDueDays
  return Object.keys(body).length ? body : null
}

/**
 * GET /ic-assessments/summary?year=&quarter=: latest APPROVED assessment per department.
 * Wire: `{ year, quarter, total, by_component: { CE: { YES, PARTIAL, NO } ... }, by_conclusion: { EFFECTIVE, NOT_EFFECTIVE } }`.
 */
export interface IcSummary {
  year: number
  quarter: number
  /** Approved assessments counted (departments). 0 = nothing approved for the period. */
  total: number
  byComponent: Record<ComponentCode, { YES: number, PARTIAL: number, NO: number }>
  byConclusion: { EFFECTIVE: number, NOT_EFFECTIVE: number }
}

export const parseSummary = (res: unknown): IcSummary | null => {
  const raw = unwrap(res)
  if (!isObj(raw)) return null
  const byComponentRaw = isObj(raw.by_component) ? raw.by_component : {}
  const byComponent = Object.fromEntries(COSO_COMPONENT_CODES.map((code) => {
    const c = isObj(byComponentRaw[code]) ? byComponentRaw[code] as Raw : {}
    return [code, { YES: int(c.YES), PARTIAL: int(c.PARTIAL), NO: int(c.NO) }]
  })) as IcSummary['byComponent']
  const conclusion = isObj(raw.by_conclusion) ? raw.by_conclusion : {}
  const departments = Array.isArray(raw.departments) ? raw.departments.length : 0
  return {
    year: int(raw.year),
    quarter: int(raw.quarter),
    total: int(raw.total, departments),
    byComponent,
    byConclusion: { EFFECTIVE: int(conclusion.EFFECTIVE), NOT_EFFECTIVE: int(conclusion.NOT_EFFECTIVE) }
  }
}
