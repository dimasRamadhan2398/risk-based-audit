/**
 * Follow-up status of an executive summary, driven by the Action Taken Reports (ATR).
 *
 * Source: GET /executive-summaries/follow-up-status
 *   ?audit_result_report_id=<uuid> | ?report_number=<nomor>   (Individual)
 *   ?year=YYYY&quarter=N                                       (Compilation)
 *
 * Four buckets: Closed / In Progress / Overdue / No Action. Cancelled ATRs are excluded by the backend.
 * An approved summary keeps the figures frozen at approval; these helpers only map a live response
 * onto the form rows, they never invent statuses, PIC names or evidence.
 */
import type { FollowUpStatus } from '~/utils/executiveSummaryCompilation'

export type FollowUpBucket = 'CLOSED' | 'IN_PROGRESS' | 'OVERDUE' | 'NO_ACTION'

export type NoActionReason = 'NO_PIC' | 'NO_PLAN_14D' | 'NO_PLAN'

export interface FollowUpBuckets {
  closed: number
  inProgress: number
  overdue: number
  noAction: number
}

export interface FollowUpFinding {
  findingId: string
  auditResultReportId: string
  reportNumber: string
  atrId: string | null
  title: string
  divisi: string
  atrStatus: string | null
  bucket: FollowUpBucket
  noActionReason: NoActionReason | null
  dueDate: string | null
  progress: number | null
  picName: string | null
  overdueDays: number
  /** Evidence file names, joined; '' when there is none */
  evidence: string
}

export interface FollowUpStatusResult {
  asOf: string
  buckets: FollowUpBuckets
  total: number
  findingsRestricted: boolean
  findings: FollowUpFinding[]
}

export type FollowUpStatusQuery
  = | { auditResultReportId: string }
    | { reportNumber: string }
    | { year: number, quarter: number }

export const FOLLOW_UP_STATUSES: readonly FollowUpStatus[] = ['Closed', 'In Progress', 'Overdue', 'No Action']

export const BUCKET_STATUS: Record<FollowUpBucket, FollowUpStatus> = {
  CLOSED: 'Closed',
  IN_PROGRESS: 'In Progress',
  OVERDUE: 'Overdue',
  NO_ACTION: 'No Action'
}

const NO_ACTION_REASONS: readonly NoActionReason[] = ['NO_PIC', 'NO_PLAN_14D', 'NO_PLAN']

/** i18n key of a follow-up status label ('executiveSummaryForm.followUp.statuses.<key>'). */
export const followUpStatusI18nKey = (status: string): string => {
  switch (status) {
    case 'Closed': return 'executiveSummaryForm.followUp.statuses.closed'
    case 'In Progress': return 'executiveSummaryForm.followUp.statuses.inProgress'
    case 'Overdue': return 'executiveSummaryForm.followUp.statuses.overdue'
    case 'No Action': return 'executiveSummaryForm.followUp.statuses.noAction'
    default: return ''
  }
}

/** i18n key of a No Action reason ('executiveSummaryForm.followUp.noActionReasons.<reason>'); '' when none. */
export const noActionReasonI18nKey = (reason: string | null | undefined): string => {
  switch (reason) {
    case 'NO_PIC': return 'executiveSummaryForm.followUp.noActionReasons.noPic'
    case 'NO_PLAN_14D': return 'executiveSummaryForm.followUp.noActionReasons.noPlan14d'
    case 'NO_PLAN': return 'executiveSummaryForm.followUp.noActionReasons.noPlan'
    default: return ''
  }
}

export const followUpStatusColor = (status: string): 'success' | 'info' | 'error' | 'warning' | 'neutral' => {
  switch (status) {
    case 'Closed': return 'success'
    case 'In Progress': return 'info'
    case 'Overdue': return 'error'
    case 'No Action': return 'warning'
    default: return 'neutral'
  }
}

const isCount = (value: unknown): value is number =>
  typeof value === 'number' && Number.isInteger(value) && value >= 0

const str = (value: unknown): string => (typeof value === 'string' ? value : '')
const strOrNull = (value: unknown): string | null => (typeof value === 'string' && value !== '' ? value : null)

/** Evidence of a finding as display text: a string, a file object or a list of them. Never a URL to open. */
const evidenceText = (value: unknown): string => {
  if (!value) return ''
  if (typeof value === 'string') return value
  if (Array.isArray(value)) return value.map(evidenceText).filter(Boolean).join(', ')
  if (typeof value === 'object') {
    const v = value as Record<string, unknown>
    return str(v.fileName) || str(v.file_name) || str(v.filename) || str(v.name) || str(v.originalName) || ''
  }
  return ''
}

const parseFinding = (raw: unknown): FollowUpFinding | null => {
  if (!raw || typeof raw !== 'object') return null
  const f = raw as Record<string, unknown>
  const bucket = f.bucket as FollowUpBucket
  if (!(bucket in BUCKET_STATUS)) return null
  const findingId = str(f.findingId)
  if (!findingId) return null
  const reason = NO_ACTION_REASONS.includes(f.noActionReason as NoActionReason) ? (f.noActionReason as NoActionReason) : null
  const progress = typeof f.progress === 'number' && Number.isFinite(f.progress) ? f.progress : null
  return {
    findingId,
    auditResultReportId: str(f.auditResultReportId),
    reportNumber: str(f.reportNumber),
    atrId: strOrNull(f.atrId),
    title: str(f.title),
    divisi: str(f.divisi),
    atrStatus: strOrNull(f.atrStatus),
    bucket,
    noActionReason: bucket === 'NO_ACTION' ? reason : null,
    dueDate: strOrNull(f.dueDate),
    progress,
    picName: strOrNull(f.picName),
    overdueDays: isCount(f.overdueDays) ? f.overdueDays : 0,
    evidence: evidenceText(f.evidence)
  }
}

/**
 * Validates the follow-up-status response (bare or wrapped in { data }) and maps it to FollowUpStatusResult.
 * Throws when the bucket counts are missing, so a partial response never reaches the form.
 * Malformed finding rows are dropped.
 */
export const parseFollowUpStatus = (body: unknown): FollowUpStatusResult => {
  const outer = body as Record<string, unknown> | null | undefined
  const data = (outer && typeof outer.data === 'object' && outer.data !== null ? outer.data : outer) as Record<string, unknown> | null | undefined
  if (!data || typeof data !== 'object') throw new Error('Invalid follow-up status response')

  const b = data.buckets as Record<string, unknown> | undefined
  if (!b || typeof b !== 'object') throw new Error('Invalid follow-up status field: buckets')
  for (const field of ['closed', 'inProgress', 'overdue'] as const) {
    if (!isCount(b[field])) throw new Error(`Invalid follow-up status field: buckets.${field}`)
  }
  const buckets: FollowUpBuckets = {
    closed: b.closed as number,
    inProgress: b.inProgress as number,
    overdue: b.overdue as number,
    noAction: isCount(b.noAction) ? b.noAction : 0
  }
  const bucketTotal = buckets.closed + buckets.inProgress + buckets.overdue + buckets.noAction
  const findings = Array.isArray(data.findings)
    ? (data.findings as unknown[]).map(parseFinding).filter((f): f is FollowUpFinding => f !== null)
    : []

  return {
    asOf: str(data.asOf),
    buckets,
    total: isCount(data.total) ? data.total : bucketTotal,
    findingsRestricted: data.findingsRestricted === true,
    findings
  }
}

/** Query string parameters of the endpoint for a query. */
export const followUpStatusParams = (query: FollowUpStatusQuery): Record<string, string | number> => {
  if ('auditResultReportId' in query) return { audit_result_report_id: query.auditResultReportId }
  if ('reportNumber' in query) return { report_number: query.reportNumber }
  return { year: query.year, quarter: query.quarter }
}

/** Fields of a Top-5 or matrix row that follow the ATR of its finding. */
export interface FollowUpLinkedRow {
  findingId?: string
  atrId?: string | null
  noActionReason?: NoActionReason | null
}

/**
 * Writes the ATR bucket of each linked finding onto the Top-5 rows (statusTL, atrId, noActionReason).
 * Rows without a findingId, or whose finding is not in the response, are left unchanged.
 */
export const applyFollowUpToTopFindings = <T extends FollowUpLinkedRow & { statusTL: string }>(
  rows: T[],
  findings: readonly FollowUpFinding[]
): void => {
  const byId = new Map(findings.map(f => [f.findingId, f]))
  for (const row of rows) {
    const f = row.findingId ? byId.get(row.findingId) : undefined
    if (!f) continue
    row.statusTL = BUCKET_STATUS[f.bucket]
    row.atrId = f.atrId
    row.noActionReason = f.noActionReason
  }
}

/** LHA finding fields used to complete a matrix row: risk category, recommendation, business process. */
export interface LhaFindingInfo {
  category?: string
  action?: string
  reportTitle?: string
  department?: string
}

export interface FollowUpMatrixRow extends FollowUpLinkedRow {
  nomor: string
  division: string
  unitKerja: string
  prosesBisnis: string
  judulTemuan: string
  nilaiRisiko: string
  rekomendasi: string
  dueDate: string
  picUnit: string
  progres: number
  status: FollowUpStatus
  buktiTL: string
  auditResultReportId?: string
}

/**
 * One matrix row per finding, with the status, progress, due date, PIC and evidence of its ATR.
 * Empty when the backend restricted the per-finding detail.
 */
export const matrixRowsFromFollowUp = (
  result: FollowUpStatusResult,
  lhaInfo: (findingId: string) => LhaFindingInfo | undefined = () => undefined
): FollowUpMatrixRow[] => {
  if (result.findingsRestricted) return []
  return result.findings.map((f) => {
    const info = lhaInfo(f.findingId)
    const division = f.divisi || info?.department || ''
    return {
      nomor: f.reportNumber,
      division,
      unitKerja: division,
      prosesBisnis: info?.reportTitle || '',
      judulTemuan: f.title,
      nilaiRisiko: info?.category || '',
      rekomendasi: info?.action || '',
      dueDate: f.dueDate || '',
      picUnit: f.picName || '',
      progres: f.progress ?? 0,
      status: BUCKET_STATUS[f.bucket],
      buktiTL: f.evidence,
      findingId: f.findingId,
      atrId: f.atrId,
      noActionReason: f.noActionReason,
      auditResultReportId: f.auditResultReportId
    }
  })
}

/**
 * Closed share of a saved follow-up table in percent (0-100), over every row present.
 * Works for old three-row documents and new four-row documents alike.
 */
export const followUpClosedPercent = (rows: readonly { status: string, jumlah: unknown }[] | null | undefined): number => {
  if (!Array.isArray(rows)) return 0
  const total = rows.reduce((sum, r) => sum + (Number(r?.jumlah) || 0), 0)
  const closed = Number(rows.find(r => r?.status === 'Closed')?.jumlah) || 0
  return total > 0 ? Math.round((closed / total) * 100) : 0
}

/** Local date of an RFC3339 or YYYY-MM-DD value for the "Status per" labels; '' when missing or invalid. */
export const formatFollowUpDate = (value: string | null | undefined, locale: string = 'id'): string => {
  if (!value) return ''
  const date = new Date(/^\d{4}-\d{2}-\d{2}$/.test(value) ? `${value}T00:00:00` : value)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat(locale === 'en' ? 'en-GB' : 'id-ID', { dateStyle: 'medium' }).format(date)
}
