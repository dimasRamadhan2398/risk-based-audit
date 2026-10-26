// Pure helpers for saving a strategic plan (POST/PUT /strategic-plans).

import type { StrategicAuditPlan } from '~/types/audit'

/** Non-empty categories the backend accepts; "" means no category (existing rows). Anything else is a 400. */
export const STRATEGIC_PLAN_CATEGORIES = ['Operational', 'Financial', 'Quality', 'Issue', 'Efficiency'] as const
export type StrategicPlanCategory = typeof STRATEGIC_PLAN_CATEGORIES[number] | ''

// The fields the plan form edits. Anything else on a loaded row (id, created_at, updated_at, ...)
// is left out so a save never writes back server-managed columns.
const STRATEGIC_PLAN_FORM_FIELDS = [
  'code', 'goalId', 'strategicObjective', 'kpi', 'unit', 'category', 'hibHig', 'periodType',
  'selectedPeriod', 'yearStart', 'yearEnd', 'kpiTargets', 'kpiActuals', 'internalAuditSO',
  'actual', 'target', 'calculation', 'status'
] as const satisfies ReadonlyArray<keyof StrategicAuditPlan>

/** Body for POST/PUT: only the form's own fields; `category` is always sent ("" clears it). */
export const buildStrategicPlanPayload = (form: Partial<StrategicAuditPlan>): Partial<StrategicAuditPlan> => {
  const body: Record<string, unknown> = {}
  for (const key of STRATEGIC_PLAN_FORM_FIELDS) {
    if (form[key] !== undefined) body[key] = form[key]
  }
  body.category = (form.category ?? '').trim()
  return body as Partial<StrategicAuditPlan>
}

/**
 * goalId for a new plan: there is one corporate goal and no picker, so it is the first goal of the
 * active VMG, by id or else goal_code (the convention the old goal picker used). "" when there is none.
 */
export const defaultStrategicPlanGoalId = (goals?: ReadonlyArray<{ id?: string, goal_code?: string }> | null): string => {
  const goal = goals?.[0]
  return goal?.id || goal?.goal_code || ''
}

/** The form fields of a loaded plan, for editing; a missing category becomes "". */
export const strategicPlanFormFromPlan = (plan: Partial<StrategicAuditPlan>): Partial<StrategicAuditPlan> => {
  const form: Record<string, unknown> = { id: plan.id }
  for (const key of STRATEGIC_PLAN_FORM_FIELDS) {
    if (plan[key] !== undefined) form[key] = plan[key]
  }
  form.category = plan.category ?? ''
  return form as Partial<StrategicAuditPlan>
}

/**
 * What a KPI target/realization input keeps while typing: digits, one decimal point and a leading minus.
 * A comma is read as the decimal point; letters, spaces, "%" and "Rp" are dropped.
 */
export const sanitizeKpiNumberInput = (value: unknown): string => {
  const raw = String(value ?? '').replace(/,/g, '.')
  const negative = raw.trimStart().startsWith('-')
  const [whole = '', ...rest] = raw.replace(/[^\d.]/g, '').split('.')
  const number = rest.length ? `${whole}.${rest.join('')}` : whole
  return negative ? `-${number}` : number
}

/** A stored KPI target/realization: empty (not filled yet) or a plain number, as the backend accepts. */
export const isKpiNumber = (value: unknown): boolean => {
  const v = String(value ?? '').trim()
  return v === '' || /^-?\d+(\.\d+)?$/.test(v)
}

/** The first target/realization that is not a number, in period order; null when all are valid. */
export const firstInvalidKpiValue = (
  form: Pick<Partial<StrategicAuditPlan>, 'kpiTargets' | 'kpiActuals'>
): { field: 'target' | 'realization', period: string } | null => {
  for (const [field, values] of [['target', form.kpiTargets], ['realization', form.kpiActuals]] as const) {
    const periods = Object.keys(values ?? {}).sort()
    const period = periods.find(p => !isKpiNumber((values as Record<string, unknown>)[p]))
    if (period !== undefined) return { field, period }
  }
  return null
}

/**
 * Which form field a failed save belongs to: the backend answers an unknown category with
 * 400 "Invalid category ...". Null when the error is not tied to a field.
 */
export const strategicPlanErrorField = (error: unknown, message: string): 'category' | null => {
  const status = (error as { statusCode?: number, status?: number } | null)?.statusCode
    ?? (error as { status?: number } | null)?.status
  return status === 400 && /category/i.test(message) ? 'category' : null
}
