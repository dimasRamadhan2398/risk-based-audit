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
 * Which form field a failed save belongs to: the backend answers an unknown category with
 * 400 "Invalid category ...". Null when the error is not tied to a field.
 */
export const strategicPlanErrorField = (error: unknown, message: string): 'category' | null => {
  const status = (error as { statusCode?: number, status?: number } | null)?.statusCode
    ?? (error as { status?: number } | null)?.status
  return status === 400 && /category/i.test(message) ? 'category' : null
}
