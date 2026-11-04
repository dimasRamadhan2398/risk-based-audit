/**
 * Audit Activity Plan after the rework (docs/audit-activity-plan-rework.md): rows link a PKAT
 * (Annual Audit Plan) row by annualActivityId. PKAT-owned values are copied into the row read-only
 * and never sent; the server resolves them. Replaces the tests of the old plannedActivities store.
 */
import { describe, it, expect } from 'vitest'
import type { AnnualActivityOption } from '~/types/audit'
import {
  buildActivityPlanPayload,
  categoryMatchesActivityType,
  emptyActivityPlanForm,
  rowFromOption,
  sumFormTotals,
  validateActivityPlanHeader,
  validateActivityPlanRows
} from '~/utils/activityPlanForm'

const t = (key: string) => key

const option = (extra: Partial<AnnualActivityOption> = {}): AnnualActivityOption => ({
  annualActivityId: 'aa-1',
  itemNumber: '1',
  auditObject: 'Audit IT',
  auditCategory: 'Assurance',
  auditUniverse: null,
  department: null,
  schedule: 'Q1',
  teamSize: 3,
  totalMandays: 20,
  pkatSupervisorName: 'Budi',
  available: true,
  usedBy: null,
  prefill: { assignmentCode: '', justification: [], objectives: [], focus: [] },
  existingRow: null,
  ...extra
} as AnnualActivityOption)

/** A form with valid rows built from `options`. */
const formWith = (...options: AnnualActivityOption[]) => {
  const form = emptyActivityPlanForm()
  Object.assign(form, { annualPlanId: 'pkat-1', activityType: 'Assurance', title: 'Plan 2026' })
  form.rows = options.map((o) => {
    const row = rowFromOption(o)
    row.auditKind = 'Operasional'
    row.justification = ['Reason']
    row.objectives = ['Objective']
    row.focus = ['Focus']
    return row
  })
  return form
}

describe('linking a PKAT row', () => {
  it('copies the PKAT values into the row read-only', () => {
    const row = rowFromOption(option())
    expect(row.annualActivityId).toBe('aa-1')
    expect(row.source).toBe('pkat')
    expect(row.pkat).toMatchObject({ auditObject: 'Audit IT', auditCategory: 'Assurance', schedule: 'Q1', teamSize: 3, totalMandays: 20, pkatSupervisorName: 'Budi' })
  })

  it('prefills the legacy values and keeps the code of a linked legacy activity', () => {
    const row = rowFromOption(option({
      prefill: { assignmentCode: '', justification: ['J1'], objectives: [], focus: [] },
      existingRow: { id: 'old-1', assignmentCode: 'ASR-2026-001', status: 'PLANNED' }
    }))
    expect(row.assignmentCode).toBe('ASR-2026-001')
    expect(row.assignmentCodeEditable).toBe(true)
    expect(row.justification).toEqual(['J1'])
    // Empty lists start with one empty item to type into.
    expect(row.objectives).toEqual([''])
  })

  it('locks the code of an activity that has already started', () => {
    const row = rowFromOption(option({ existingRow: { id: 'old-1', assignmentCode: 'ASR-2026-001', status: 'IN_PROGRESS' } }))
    expect(row.assignmentCodeEditable).toBe(false)
    expect(row.inUse).toBe(true)
  })

  it('sums team size and mandays over the rows', () => {
    const form = formWith(option(), option({ annualActivityId: 'aa-2', teamSize: 2, totalMandays: 5 }))
    expect(sumFormTotals(form.rows)).toEqual({ rowCount: 2, teamSize: 5, totalMandays: 25 })
  })
})

describe('save payload', () => {
  it('sends the link and the plan-owned fields only, never the PKAT-owned ones', () => {
    const form = formWith(option())
    form.rows[0]!.justification = [' Reason ', '']
    const payload = buildActivityPlanPayload(form)
    expect(payload).toMatchObject({ annualPlanId: 'pkat-1', activityType: 'Assurance', title: 'Plan 2026' })
    const [sent] = payload.rows
    expect(sent).toMatchObject({ annualActivityId: 'aa-1', auditKind: 'Operasional', justification: ['Reason'] })
    for (const pkatField of ['auditObject', 'schedule', 'teamSize', 'totalMandays', 'pkatSupervisorName', 'pkat']) {
      expect(sent).not.toHaveProperty(pkatField)
    }
  })

  it('sends the id of a saved row and none for a new one', () => {
    const form = formWith(option(), option({ annualActivityId: 'aa-2' }))
    form.rows[0]!.id = 'row-1'
    const payload = buildActivityPlanPayload(form)
    expect(payload.rows[0]!.id).toBe('row-1')
    expect(payload.rows[1]).not.toHaveProperty('id')
  })

  it('drops the team leader name when no leader is picked', () => {
    const form = formWith(option())
    form.rows[0]!.teamLeaderName = 'Stale'
    expect(buildActivityPlanPayload(form).rows[0]).toMatchObject({ teamLeaderEmployeeId: null, teamLeaderName: '' })
  })
})

describe('validation', () => {
  it('a valid form has no issues', () => {
    const form = formWith(option())
    expect(validateActivityPlanHeader(form)).toEqual([])
    expect(validateActivityPlanRows(form, t)).toEqual([])
  })

  it('the header needs a PKAT, an activity type and a title', () => {
    const keys = validateActivityPlanHeader(emptyActivityPlanForm()).map(i => i.key)
    expect(keys).toEqual([
      'auditActivityPlan.validation.annualPlanRequired',
      'auditActivityPlan.validation.activityTypeRequired',
      'auditActivityPlan.validation.titleRequired'
    ])
  })

  it('needs at least one row', () => {
    const form = formWith()
    expect(validateActivityPlanRows(form, t).map(i => i.key)).toEqual(['auditActivityPlan.validation.rowsRequired'])
  })

  it('rejects the same PKAT row twice and a duplicate assignment code', () => {
    const form = formWith(option(), option())
    form.rows[0]!.assignmentCode = 'ASR-1'
    form.rows[1]!.assignmentCode = 'ASR-1'
    const issues = validateActivityPlanRows(form, t)
    expect(issues).toContainEqual(expect.objectContaining({ key: 'auditActivityPlan.validation.rowDuplicate', row: 2 }))
    expect(issues).toContainEqual(expect.objectContaining({ key: 'auditActivityPlan.validation.assignmentCodeDuplicate', row: 2 }))
  })

  it('rejects a PKAT row whose category does not match the activity type', () => {
    const form = formWith(option({ auditCategory: 'Investigation' }))
    expect(validateActivityPlanRows(form, t)).toContainEqual(expect.objectContaining({ key: 'auditActivityPlan.validation.rowTypeMismatch', row: 1 }))
  })

  it('maps legacy PKAT categories to activity types', () => {
    expect(categoryMatchesActivityType('ASSURANCE', 'Assurance')).toBe(true)
    expect(categoryMatchesActivityType('Special Audit', 'Audit Khusus')).toBe(true)
    expect(categoryMatchesActivityType('Investigation', 'Audit Investigasi')).toBe(true)
    expect(categoryMatchesActivityType('Assurance', 'Audit Investigasi')).toBe(false)
    expect(categoryMatchesActivityType('Assurance', '')).toBe(false)
  })
})
