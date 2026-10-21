/**
 * Audit Activity Plan (Formulir Rencana Kegiatan Audit): mapping between the audit-service
 * /activity-plans API and the form, the save payload, client validation (mirrors the server)
 * and the error-code mapping. Contract: docs/audit-activity-plan-rework.md, section 6.
 */
import {
  ACTIVITY_PLAN_ACTIVITY_TYPES,
  ACTIVITY_PLAN_FORM_VERSION,
  type ActivityPlanActivityType,
  type ActivityPlanAnnualPlan,
  type ActivityPlanAuditUniverse,
  type ActivityPlanDetail,
  type ActivityPlanErrorDetails,
  type ActivityPlanFormState,
  type ActivityPlanInput,
  type ActivityPlanNamedRef,
  type ActivityPlanRow,
  type ActivityPlanRowForm,
  type ActivityPlanSummary,
  type ActivityPlanTotals,
  type AnnualActivityOption,
  type AnnualAuditAttachment,
  type LegacyActivityPlan
} from '~/types/audit'
import {
  NUMBERED_LIST_MAX_ITEMS,
  NUMBERED_LIST_MAX_ITEM_LENGTH,
  checkNumberedList,
  normalizeListItems,
  toListItems
} from '~/utils/numberedList'
import { getErrorCode, getErrorStatus, getUserErrorMessage, type TranslateFn } from '~/utils/error'

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

export const ACTIVITY_PLAN_TITLE_MAX = 255
export const ACTIVITY_PLAN_MAX_ROWS = 200
export const AUDIT_KIND_MAX = 100
export const ASSIGNMENT_CODE_MAX = 50
export const ASSIGNMENT_CODE_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._/-]*$/

/** i18n key (under auditActivityPlan.activityTypes) per Jenis Kegiatan Audit. */
export const ACTIVITY_TYPE_LABEL_KEYS: Record<ActivityPlanActivityType, string> = {
  'Assurance': 'auditActivityPlan.activityTypes.assurance',
  'Audit Khusus': 'auditActivityPlan.activityTypes.specialAudit',
  'Audit Tujuan Tertentu': 'auditActivityPlan.activityTypes.specificPurpose',
  'Audit Investigasi': 'auditActivityPlan.activityTypes.investigation'
}

/**
 * Jenis Audit suggestions. The stored value is the Indonesian term of the paper form; the label is
 * translated. Other values may be typed in.
 */
export const AUDIT_KIND_SUGGESTIONS: ReadonlyArray<{ value: string, labelKey: string }> = [
  { value: 'Operasional', labelKey: 'auditActivityPlan.auditKinds.operational' },
  { value: 'Keuangan', labelKey: 'auditActivityPlan.auditKinds.financial' },
  { value: 'Kepatuhan', labelKey: 'auditActivityPlan.auditKinds.compliance' },
  { value: 'Teknologi Informasi', labelKey: 'auditActivityPlan.auditKinds.it' },
  { value: 'Kinerja', labelKey: 'auditActivityPlan.auditKinds.performance' }
]

// ---------------------------------------------------------------------------
// Header type <-> PKAT category (section 4.3)
// ---------------------------------------------------------------------------

/** Type code of a PKAT category, like audit-service activitycode.TypeCode for the known values. */
const CATEGORY_TYPE_CODES: Record<string, string> = {
  'ASSURANCE': 'ASR',
  'SPECIAL AUDIT': 'SPC',
  'SPECIAL': 'SPC',
  'SPECIFIC REASON': 'SPR',
  'INVESTIGATION': 'INV'
}

const ACTIVITY_TYPE_CODES: Record<ActivityPlanActivityType, string> = {
  'Assurance': 'ASR',
  'Audit Khusus': 'SPC',
  'Audit Tujuan Tertentu': 'SPR',
  'Audit Investigasi': 'INV'
}

/** Normalised type code of a PKAT row category ("Assurance", legacy "ASSURANCE", ...); "" when not mapped. */
export const categoryTypeCode = (category: string | null | undefined): string => {
  const norm = String(category ?? '').toUpperCase().replace(/[^A-Z0-9]+/g, ' ').trim()
  return CATEGORY_TYPE_CODES[norm] ?? ''
}

/** Whether a PKAT row of `category` can be part of an Audit Activity Plan of `activityType`. */
export const categoryMatchesActivityType = (category: string | null | undefined, activityType: string | null | undefined): boolean => {
  const code = ACTIVITY_TYPE_CODES[activityType as ActivityPlanActivityType]
  return !!code && categoryTypeCode(category) === code
}

export const isActivityPlanActivityType = (value: unknown): value is ActivityPlanActivityType =>
  (ACTIVITY_PLAN_ACTIVITY_TYPES as readonly unknown[]).includes(value)

// ---------------------------------------------------------------------------
// API -> client normalisation
// ---------------------------------------------------------------------------

type Raw = Record<string, any>

const str = (value: unknown): string => (value === null || value === undefined ? '' : String(value))
const num = (value: unknown): number => {
  const n = typeof value === 'number' ? value : Number.parseFloat(str(value))
  return Number.isFinite(n) ? n : 0
}

const toNamedRef = (raw: unknown): ActivityPlanNamedRef | null => {
  if (!raw || typeof raw !== 'object') return null
  const r = raw as Raw
  if (!r.id && !r.name) return null
  return { id: str(r.id), name: str(r.name) }
}

const toAuditUniverse = (raw: unknown): ActivityPlanAuditUniverse | null => {
  if (!raw || typeof raw !== 'object') return null
  const r = raw as Raw
  if (!r.id && !r.name) return null
  return { id: str(r.id), name: str(r.name), riskLevel: str(r.riskLevel) }
}

const toAnnualPlan = (raw: unknown, fallbackYear = 0): ActivityPlanAnnualPlan | null => {
  if (!raw || typeof raw !== 'object') return null
  const r = raw as Raw
  return {
    id: str(r.id),
    code: str(r.code),
    title: str(r.title),
    year: num(r.year) || fallbackYear,
    status: str(r.status),
    isActive: r.isActive !== false
  }
}

const toTotals = (raw: unknown): ActivityPlanTotals => {
  const r = (raw && typeof raw === 'object' ? raw : {}) as Raw
  return { rowCount: num(r.rowCount), teamSize: num(r.teamSize), totalMandays: num(r.totalMandays) }
}

const toActivityType = (value: unknown): ActivityPlanActivityType | '' =>
  isActivityPlanActivityType(value) ? value : ''

/** A plan that was not saved with the RKA form is legacy (view and delete only). */
export const isLegacyFormVersion = (formVersion: unknown): boolean => num(formVersion) !== ACTIVITY_PLAN_FORM_VERSION

export const normalizeActivityPlanSummary = (raw: Raw): ActivityPlanSummary => {
  const formVersion = num(raw.formVersion) || 1
  return {
    id: str(raw.id),
    formVersion,
    isLegacy: isLegacyFormVersion(formVersion),
    year: num(raw.year),
    activityType: toActivityType(raw.activityType),
    title: str(raw.title),
    annualPlan: toAnnualPlan(raw.annualPlan, num(raw.year)),
    totals: toTotals(raw.totals),
    status: str(raw.status),
    attachmentCount: num(raw.attachmentCount),
    createdBy: str(raw.createdBy),
    createdAt: str(raw.createdAt),
    updatedAt: str(raw.updatedAt)
  }
}

const toAttachments = (raw: unknown): AnnualAuditAttachment[] =>
  Array.isArray(raw)
    ? raw.filter(f => f && typeof f === 'object').map((f: Raw) => ({ name: str(f.name), size: str(f.size), url: str(f.url) }))
    : []

export const normalizeActivityPlanRow = (raw: Raw, index: number): ActivityPlanRow => {
  const leader = raw.teamLeader && typeof raw.teamLeader === 'object' && (raw.teamLeader.employeeId || raw.teamLeader.name)
    ? { employeeId: str(raw.teamLeader.employeeId), name: str(raw.teamLeader.name) }
    : null
  return {
    id: str(raw.id),
    no: num(raw.no) || index + 1,
    annualActivityId: str(raw.annualActivityId),
    assignmentCode: str(raw.assignmentCode),
    assignmentCodeEditable: raw.assignmentCodeEditable !== false,
    auditObject: str(raw.auditObject),
    auditCategory: str(raw.auditCategory),
    auditKind: str(raw.auditKind),
    auditUniverse: toAuditUniverse(raw.auditUniverse),
    department: toNamedRef(raw.department),
    justification: toListItems(raw.justification),
    objectives: toListItems(raw.objectives),
    focus: toListItems(raw.focus),
    schedule: str(raw.schedule),
    teamSize: num(raw.teamSize),
    totalMandays: num(raw.totalMandays),
    teamLeader: leader,
    pkatSupervisorName: str(raw.pkatSupervisorName),
    status: str(raw.status),
    inUse: raw.inUse === true,
    source: raw.source === 'snapshot' ? 'snapshot' : 'pkat'
  }
}

const toLegacy = (raw: unknown): LegacyActivityPlan | null => {
  if (!raw || typeof raw !== 'object') return null
  const r = raw as Raw
  const budget = (r.budget && typeof r.budget === 'object' ? r.budget : {}) as Raw
  const review = (r.review && typeof r.review === 'object' ? r.review : {}) as Raw
  return {
    planTitle: str(r.planTitle),
    planYear: str(r.planYear),
    planPeriodStart: str(r.planPeriodStart),
    planPeriodEnd: str(r.planPeriodEnd),
    department: str(r.department),
    creationDate: str(r.creationDate),
    plannedActivities: Array.isArray(r.plannedActivities) ? r.plannedActivities : [],
    resourceAuditors: Array.isArray(r.resourceAuditors) ? r.resourceAuditors : [],
    budget: {
      totalEstimatedCost: num(budget.totalEstimatedCost),
      totalAllocatedBudget: num(budget.totalAllocatedBudget),
      budgetNotes: str(budget.budgetNotes)
    },
    review: {
      creatorName: str(review.creatorName),
      creatorPosition: str(review.creatorPosition),
      approverName: str(review.approverName),
      approverPosition: str(review.approverPosition),
      approvalDate: str(review.approvalDate),
      additionalNotes: str(review.additionalNotes)
    }
  }
}

export const normalizeActivityPlanDetail = (raw: Raw): ActivityPlanDetail => {
  const formVersion = num(raw.formVersion) || 1
  const rows = Array.isArray(raw.rows) ? raw.rows.map((r: Raw, i: number) => normalizeActivityPlanRow(r, i)) : []
  return {
    id: str(raw.id),
    formVersion,
    isLegacy: isLegacyFormVersion(formVersion),
    annualPlanId: raw.annualPlanId ? str(raw.annualPlanId) : null,
    annualPlan: toAnnualPlan(raw.annualPlan, num(raw.year)),
    year: num(raw.year),
    activityType: toActivityType(raw.activityType),
    title: str(raw.title),
    status: str(raw.status),
    rows,
    totals: raw.totals ? toTotals(raw.totals) : sumRowTotals(rows),
    attachmentCategory: str(raw.attachmentCategory),
    attachments: toAttachments(raw.attachments),
    attachmentUploadedBy: str(raw.attachmentUploadedBy),
    attachmentUploadDate: str(raw.attachmentUploadDate),
    createdBy: str(raw.createdBy),
    createdAt: str(raw.createdAt),
    updatedAt: str(raw.updatedAt),
    legacy: toLegacy(raw.legacy)
  }
}

export const normalizeAnnualActivityOption = (raw: Raw): AnnualActivityOption => {
  const prefill = (raw.prefill && typeof raw.prefill === 'object' ? raw.prefill : {}) as Raw
  const usedBy = raw.usedBy && typeof raw.usedBy === 'object'
    ? { activityPlanId: str(raw.usedBy.activityPlanId), activityPlanTitle: str(raw.usedBy.activityPlanTitle), assignmentCode: str(raw.usedBy.assignmentCode) }
    : null
  const existingRow = raw.existingRow && typeof raw.existingRow === 'object'
    ? { id: str(raw.existingRow.id), assignmentCode: str(raw.existingRow.assignmentCode), status: str(raw.existingRow.status) }
    : null
  return {
    annualActivityId: str(raw.annualActivityId),
    itemNumber: str(raw.itemNumber),
    auditObject: str(raw.auditObject),
    auditCategory: str(raw.auditCategory),
    auditUniverse: toAuditUniverse(raw.auditUniverse),
    department: toNamedRef(raw.department),
    schedule: str(raw.schedule),
    teamSize: num(raw.teamSize),
    totalMandays: num(raw.totalMandays),
    pkatSupervisorName: str(raw.pkatSupervisorName),
    available: raw.available !== false && !usedBy,
    usedBy,
    prefill: {
      assignmentCode: str(prefill.assignmentCode),
      justification: toListItems(prefill.justification),
      objectives: toListItems(prefill.objectives),
      focus: toListItems(prefill.focus)
    },
    existingRow
  }
}

// ---------------------------------------------------------------------------
// Form rows
// ---------------------------------------------------------------------------

let rowKeySeq = 0
const nextRowKey = () => `row-${Date.now().toString(36)}-${++rowKeySeq}`

/** A list input starts with one empty item so there is something to type into. */
const editableList = (items: string[]): string[] => (items.length ? [...items] : [''])

/** Form row for a PKAT row picked in the Annual Activity Picker; its legacy values are prefilled once. */
export const rowFromOption = (option: AnnualActivityOption): ActivityPlanRowForm => {
  const existingStatus = option.existingRow?.status || ''
  const started = !!existingStatus && existingStatus.toUpperCase() !== 'PLANNED'
  return {
    key: nextRowKey(),
    annualActivityId: option.annualActivityId,
    // A linked legacy audit activity keeps its code; the server adopts that row on save.
    assignmentCode: option.prefill.assignmentCode || option.existingRow?.assignmentCode || '',
    assignmentCodeEditable: !started,
    auditKind: '',
    justification: editableList(option.prefill.justification),
    objectives: editableList(option.prefill.objectives),
    focus: editableList(option.prefill.focus),
    teamLeaderEmployeeId: null,
    teamLeaderName: '',
    inUse: started,
    source: 'pkat',
    status: existingStatus,
    pkat: {
      itemNumber: option.itemNumber,
      auditObject: option.auditObject,
      auditCategory: option.auditCategory,
      auditUniverse: option.auditUniverse,
      department: option.department,
      schedule: option.schedule,
      teamSize: option.teamSize,
      totalMandays: option.totalMandays,
      pkatSupervisorName: option.pkatSupervisorName
    }
  }
}

/** Form row for a saved row. */
export const rowFromDetail = (row: ActivityPlanRow): ActivityPlanRowForm => ({
  key: row.id || nextRowKey(),
  id: row.id || undefined,
  annualActivityId: row.annualActivityId,
  assignmentCode: row.assignmentCode,
  assignmentCodeEditable: row.assignmentCodeEditable,
  auditKind: row.auditKind,
  justification: editableList(row.justification),
  objectives: editableList(row.objectives),
  focus: editableList(row.focus),
  teamLeaderEmployeeId: row.teamLeader?.employeeId || null,
  teamLeaderName: row.teamLeader?.name || '',
  inUse: row.inUse,
  source: row.source,
  status: row.status,
  pkat: {
    itemNumber: '',
    auditObject: row.auditObject,
    auditCategory: row.auditCategory,
    auditUniverse: row.auditUniverse,
    department: row.department,
    schedule: row.schedule,
    teamSize: row.teamSize,
    totalMandays: row.totalMandays,
    pkatSupervisorName: row.pkatSupervisorName
  }
})

export const emptyActivityPlanForm = (): ActivityPlanFormState => ({
  annualPlanId: '',
  annualPlan: null,
  activityType: '',
  title: '',
  rows: [],
  attachmentCategory: '',
  attachments: [],
  file: []
})

/** Form state to edit a saved plan. */
export const formFromDetail = (detail: ActivityPlanDetail): ActivityPlanFormState => ({
  id: detail.id,
  annualPlanId: detail.annualPlanId || detail.annualPlan?.id || '',
  annualPlan: detail.annualPlan,
  activityType: detail.activityType,
  title: detail.title,
  updatedAt: detail.updatedAt || undefined,
  rows: detail.rows.map(rowFromDetail),
  attachmentCategory: detail.attachmentCategory,
  attachments: [...detail.attachments],
  file: []
})

export const sumRowTotals = (rows: ReadonlyArray<{ teamSize: number, totalMandays: number }>): ActivityPlanTotals => ({
  rowCount: rows.length,
  teamSize: rows.reduce((sum, r) => sum + num(r.teamSize), 0),
  totalMandays: rows.reduce((sum, r) => sum + num(r.totalMandays), 0)
})

export const sumFormTotals = (rows: readonly ActivityPlanRowForm[]): ActivityPlanTotals =>
  sumRowTotals(rows.map(r => r.pkat))

// ---------------------------------------------------------------------------
// Save payload
// ---------------------------------------------------------------------------

/**
 * Body of POST/PUT /activity-plans. Rows go in display order (the server numbers them). Only the
 * fields the Audit Activity Plan owns are sent: the PKAT-owned ones (object, universe, schedule,
 * team, mandays, supervisor) are resolved by the server from the PKAT row.
 */
export const buildActivityPlanPayload = (
  form: ActivityPlanFormState,
  attachments: AnnualAuditAttachment[] = form.attachments
): ActivityPlanInput => {
  const payload: ActivityPlanInput = {
    annualPlanId: form.annualPlanId,
    activityType: form.activityType as ActivityPlanActivityType,
    title: form.title.trim(),
    rows: form.rows.map((row) => {
      const leaderId = row.teamLeaderEmployeeId || null
      return {
        ...(row.id ? { id: row.id } : {}),
        annualActivityId: row.annualActivityId,
        assignmentCode: row.assignmentCode.trim(),
        auditKind: row.auditKind.trim(),
        justification: normalizeListItems(row.justification),
        objectives: normalizeListItems(row.objectives),
        focus: normalizeListItems(row.focus),
        teamLeaderEmployeeId: leaderId,
        teamLeaderName: leaderId ? row.teamLeaderName.trim() : ''
      }
    }),
    attachmentCategory: form.attachmentCategory || '',
    attachments
  }
  if (form.id && form.updatedAt) payload.updatedAt = form.updatedAt
  return payload
}

// ---------------------------------------------------------------------------
// Client validation (blocks the step; mirrors the server rules)
// ---------------------------------------------------------------------------

export interface ActivityPlanIssue {
  /** i18n key. */
  key: string
  params?: Record<string, string | number>
  /** 1-based row the issue is about. */
  row?: number
}

/** Step 1: PKAT, Jenis Kegiatan Audit and Judul. */
export const validateActivityPlanHeader = (form: ActivityPlanFormState): ActivityPlanIssue[] => {
  const issues: ActivityPlanIssue[] = []
  if (!form.annualPlanId) issues.push({ key: 'auditActivityPlan.validation.annualPlanRequired' })
  if (!isActivityPlanActivityType(form.activityType)) issues.push({ key: 'auditActivityPlan.validation.activityTypeRequired' })
  const title = form.title.trim()
  if (!title || title.length > ACTIVITY_PLAN_TITLE_MAX) issues.push({ key: 'auditActivityPlan.validation.titleRequired' })
  return issues
}

const LIST_FIELDS = [
  { field: 'justification', labelKey: 'auditActivityPlan.columns.justification' },
  { field: 'objectives', labelKey: 'auditActivityPlan.columns.objectives' },
  { field: 'focus', labelKey: 'auditActivityPlan.columns.focus' }
] as const

/** Step 2: the rows. `t` resolves the column names used in the messages. */
export const validateActivityPlanRows = (form: ActivityPlanFormState, t: TranslateFn): ActivityPlanIssue[] => {
  const issues: ActivityPlanIssue[] = []
  if (!form.rows.length) return [{ key: 'auditActivityPlan.validation.rowsRequired' }]
  if (form.rows.length > ACTIVITY_PLAN_MAX_ROWS) {
    issues.push({ key: 'auditActivityPlan.validation.rowsTooMany', params: { max: ACTIVITY_PLAN_MAX_ROWS } })
  }

  const seenActivities = new Set<string>()
  const codeRows = new Map<string, number>()
  form.rows.forEach((row, index) => {
    const num = index + 1
    if (seenActivities.has(row.annualActivityId)) {
      issues.push({ key: 'auditActivityPlan.validation.rowDuplicate', params: { num }, row: num })
    }
    seenActivities.add(row.annualActivityId)

    if (form.activityType && !categoryMatchesActivityType(row.pkat.auditCategory, form.activityType)) {
      issues.push({ key: 'auditActivityPlan.validation.rowTypeMismatch', params: { num }, row: num })
    }

    const kind = row.auditKind.trim()
    if (!kind || kind.length > AUDIT_KIND_MAX) {
      issues.push({ key: 'auditActivityPlan.validation.rowAuditKind', params: { num }, row: num })
    }

    for (const { field, labelKey } of LIST_FIELDS) {
      const check = checkNumberedList(row[field])
      if (check.empty) {
        issues.push({ key: 'auditActivityPlan.validation.rowListRequired', params: { num, field: t(labelKey) }, row: num })
      } else if (check.tooMany || check.tooLong) {
        issues.push({
          key: 'auditActivityPlan.validation.rowListLimits',
          params: { num, field: t(labelKey), max: NUMBERED_LIST_MAX_ITEMS, maxLength: NUMBERED_LIST_MAX_ITEM_LENGTH },
          row: num
        })
      }
    }

    const code = row.assignmentCode.trim()
    if (code) {
      if (code.length > ASSIGNMENT_CODE_MAX || !ASSIGNMENT_CODE_PATTERN.test(code)) {
        issues.push({ key: 'auditActivityPlan.validation.assignmentCodeFormat', params: { num }, row: num })
      } else if (codeRows.has(code)) {
        issues.push({ key: 'auditActivityPlan.validation.assignmentCodeDuplicate', params: { num }, row: num })
      } else {
        codeRows.set(code, num)
      }
    }
  })
  return issues
}

// ---------------------------------------------------------------------------
// Server errors
// ---------------------------------------------------------------------------

export interface ParsedApiError {
  code?: string
  status?: number
  details: ActivityPlanErrorDetails
}

/** `rows[2].annualActivityId` -> 3 (1-based). */
const rowFromField = (field: unknown): number | undefined => {
  const match = /^rows\[(\d+)\]/.exec(str(field))
  return match ? Number(match[1]) + 1 : undefined
}

/**
 * Code, status and details of a $fetch error. audit-service sends `error.details` as a
 * JSON-encoded string; an object is accepted too. `row` is derived from `field` when missing.
 */
export const parseApiError = (error: unknown): ParsedApiError => {
  const data = (error && typeof error === 'object' ? (error as { data?: unknown }).data : undefined) as Raw | undefined
  const rawDetails = data?.error && typeof data.error === 'object' ? (data.error as Raw).details : undefined
  let details: ActivityPlanErrorDetails = {}
  if (typeof rawDetails === 'string' && rawDetails.trim().startsWith('{')) {
    try {
      const parsed = JSON.parse(rawDetails)
      if (parsed && typeof parsed === 'object') details = parsed as ActivityPlanErrorDetails
    } catch {
      details = {}
    }
  } else if (rawDetails && typeof rawDetails === 'object') {
    details = rawDetails as ActivityPlanErrorDetails
  }
  const row = Number(details.row)
  details = { ...details, row: Number.isInteger(row) && row > 0 ? row : rowFromField(details.field) }
  if (details.row === undefined) delete details.row
  return { code: getErrorCode(error), status: getErrorStatus(error), details }
}

/**
 * Translated message of a coded error: `<prefix>.<CODE>` interpolated with the details. Falls back
 * to the generic message for the status (getUserErrorMessage) when the code has no key.
 */
export const translateApiError = (error: unknown, t: TranslateFn, prefix: string, fallbackKey?: string): string => {
  const { code, details } = parseApiError(error)
  if (code) {
    const key = `${prefix}.${code}`
    const params: Record<string, string | number> = {}
    for (const [name, value] of Object.entries(details)) {
      if (typeof value === 'string' || typeof value === 'number') params[name] = value
    }
    const message = t(key, params)
    if (message && message !== key) {
      // A detail the server did not send leaves its placeholder: drop it instead of showing "{row}".
      return message.replace(/\{\w+\}/g, '').replace(/\s{2,}/g, ' ').replace(/\s+([:.,])/g, '$1').trim()
    }
  }
  return getUserErrorMessage(error, t, fallbackKey ? { fallbackKey } : {})
}

/** Form step (0 header, 1 rows, 2 attachments) a server error belongs to; undefined to stay. */
export const stepForApiError = (parsed: ParsedApiError): number | undefined => {
  switch (parsed.code) {
    case 'ANNUAL_PLAN_NOT_FOUND':
    case 'ANNUAL_PLAN_INACTIVE':
    case 'ACTIVITY_PLAN_DUPLICATE':
      return 0
    case 'ANNUAL_ACTIVITY_NOT_FOUND':
    case 'ANNUAL_ACTIVITY_TYPE_MISMATCH':
    case 'ANNUAL_ACTIVITY_TAKEN':
    case 'ASSIGNMENT_CODE_TAKEN':
    case 'ASSIGNMENT_CODE_LOCKED':
    case 'ACTIVITY_ROW_IN_USE':
      return 1
    case 'VALIDATION_ERROR': {
      const field = str(parsed.details.field)
      if (field.startsWith('rows')) return 1
      if (field.startsWith('attachment')) return 2
      return field ? 0 : undefined
    }
    default:
      return parsed.details.row ? 1 : undefined
  }
}

// ---------------------------------------------------------------------------
// Form sheet (paper layout) data
// ---------------------------------------------------------------------------

export interface ActivityPlanSheetRow {
  key: string
  no: number
  assignmentCode: string
  auditObject: string
  auditKind: string
  auditUniverse: ActivityPlanAuditUniverse | null
  justification: string[]
  objectives: string[]
  focus: string[]
  schedule: string
  teamSize: number
  totalMandays: number
  teamLeaderName: string
}

export interface ActivityPlanSheet {
  year: number
  activityType: ActivityPlanActivityType | ''
  title: string
  rows: ActivityPlanSheetRow[]
  totals: ActivityPlanTotals
}

export const sheetFromDetail = (detail: ActivityPlanDetail): ActivityPlanSheet => ({
  year: detail.annualPlan?.year || detail.year,
  activityType: detail.activityType,
  title: detail.title,
  rows: detail.rows.map((row, index) => ({
    key: row.id || String(index),
    no: index + 1,
    assignmentCode: row.assignmentCode,
    auditObject: row.auditObject,
    auditKind: row.auditKind,
    auditUniverse: row.auditUniverse,
    justification: row.justification,
    objectives: row.objectives,
    focus: row.focus,
    schedule: row.schedule,
    teamSize: row.teamSize,
    totalMandays: row.totalMandays,
    teamLeaderName: row.teamLeader?.name || ''
  })),
  totals: detail.totals
})

export const sheetFromForm = (form: ActivityPlanFormState, year: number): ActivityPlanSheet => ({
  year,
  activityType: form.activityType,
  title: form.title.trim(),
  rows: form.rows.map((row, index) => ({
    key: row.key,
    no: index + 1,
    assignmentCode: row.assignmentCode.trim(),
    auditObject: row.pkat.auditObject,
    auditKind: row.auditKind.trim(),
    auditUniverse: row.pkat.auditUniverse,
    justification: normalizeListItems(row.justification),
    objectives: normalizeListItems(row.objectives),
    focus: normalizeListItems(row.focus),
    schedule: row.pkat.schedule,
    teamSize: row.pkat.teamSize,
    totalMandays: row.pkat.totalMandays,
    teamLeaderName: row.teamLeaderEmployeeId ? row.teamLeaderName : ''
  })),
  totals: sumFormTotals(form.rows)
})

/** "PLANNED" / "in_progress" -> "Planned" / "In Progress". */
export const formatActivityStatus = (status: string | null | undefined): string =>
  str(status).split(/[_\s]+/).filter(Boolean).map(w => w.charAt(0).toUpperCase() + w.slice(1).toLowerCase()).join(' ')

