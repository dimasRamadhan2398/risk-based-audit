import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import {
  buildAnnualActivityPayload,
  buildAnnualPlanPayload,
  sumActivityMandays,
  toCount
} from '~/utils/annualPlanPayload'
import { ANNUAL_ACTIVITY_TYPE_OPTIONS, AnnualAuditPlanStatus, type AnnualPlanForm } from '~/types/audit'

const row = (over: Record<string, unknown> = {}) => ({
  id: 'row-uuid-1',
  activityCode: 'ASR-2026-001',
  name: 'Keuangan Cabang',
  category: 'Assurance',
  departmentId: 'dept-fin',
  department: 'Finance',
  auditUniverseId: 'au-fin',
  riskName: 'Risk Finance (server snapshot)',
  riskLevel: 'High',
  justification: 'Nilai transaksi tinggi',
  auditObjective: 'Menilai kontrol kas',
  auditFocus: 'Rekonsiliasi bank',
  timelineText: 'Q1 2026',
  auditorCount: 3,
  totalMandays: 30,
  supervisorName: 'Budi Santoso, CIA',
  ...over
})

const form = (over: Partial<AnnualPlanForm> = {}): AnnualPlanForm => ({
  code: 'PKAT-2026-001',
  title: '  Rencana Audit 2026  ',
  activityType: 'Assurance',
  version: 'v1.0',
  revisionHistory: [],
  activities: [row()] as AnnualPlanForm['activities'],
  status: AnnualAuditPlanStatus.DRAFT,
  selectedMonths: [5, 0, 2],
  notes: '',
  attachmentCategory: 'Plan',
  attachmentUploadedBy: 'Auditor',
  attachmentUploadDate: '2026-01-02',
  attachments: [],
  file: null,
  staffApprovalNote: '',
  managerApprovalNote: '',
  chiefApprovalNote: '',
  isActive: true,
  year: '2026',
  ...over
})

describe('Formulir Rencana Kegiatan Audit: activity row payload', () => {
  it('sends every template field of a row under its backend name', () => {
    const payload = buildAnnualActivityPayload(row() as never, 0)

    expect(payload).toEqual({
      id: 'row-uuid-1',
      itemNumber: '1',
      activityCode: 'ASR-2026-001',
      name: 'Keuangan Cabang',
      category: 'Assurance',
      departmentId: 'dept-fin',
      auditUniverseId: 'au-fin',
      justification: 'Nilai transaksi tinggi',
      auditObjective: 'Menilai kontrol kas',
      auditFocus: 'Rekonsiliasi bank',
      timelineText: 'Q1 2026',
      auditorCount: 3,
      totalMandays: 30,
      supervisorName: 'Budi Santoso, CIA'
    })
  })

  it('never sends the server snapshots (riskName, riskLevel, department name) or dropped fields', () => {
    const payload = buildAnnualActivityPayload(row({ involvedDepartments: [{ name: 'IT' }], groupTitle: 'G', notesObjective: 'N' }) as never, 2)

    expect(payload).not.toHaveProperty('riskName')
    expect(payload).not.toHaveProperty('riskLevel')
    expect(payload).not.toHaveProperty('department')
    expect(payload).not.toHaveProperty('involvedDepartments')
    expect(payload).not.toHaveProperty('groupTitle')
    expect(payload).not.toHaveProperty('notesObjective')
    expect(payload.itemNumber).toBe('3')
  })

  it('keeps the row id on edit and omits it for a new row', () => {
    expect(buildAnnualActivityPayload(row({ id: undefined }) as never, 0)).not.toHaveProperty('id')
    expect(buildAnnualActivityPayload(row() as never, 0).id).toBe('row-uuid-1')
  })

  it('does not blank departmentId or auditUniverseId when a loaded row has none', () => {
    const payload = buildAnnualActivityPayload(row({ departmentId: '', auditUniverseId: '' }) as never, 0)

    expect(payload).not.toHaveProperty('departmentId')
    expect(payload).not.toHaveProperty('auditUniverseId')
  })

  it('sends team size and mandays as integers', () => {
    const payload = buildAnnualActivityPayload(row({ auditorCount: '4', totalMandays: '' }) as never, 0)

    expect(payload.auditorCount).toBe(4)
    expect(payload.totalMandays).toBe(0)
  })
})

describe('Formulir Rencana Kegiatan Audit: plan payload', () => {
  it('sends the plan header (title, activityType, year) and the rows', () => {
    const payload = buildAnnualPlanPayload(form(), { quarters: ['Q1', 'Q3'], attachments: [] })

    expect(payload.title).toBe('Rencana Audit 2026')
    expect(payload.activityType).toBe('Assurance')
    expect(payload.year).toBe(2026)
    expect(payload.selectedMonths).toEqual([0, 2, 5])
    expect(payload.quarters).toEqual(['Q1', 'Q3'])
    expect(payload.activities).toHaveLength(1)
    expect(payload.activities[0]).toHaveProperty('auditUniverseId', 'au-fin')
  })

  it('does not send plan-level auditor, days, supervisor or total mandays', () => {
    const payload = buildAnnualPlanPayload(form({ activities: [row(), row({ id: 'r2' })] as never }), { quarters: [], attachments: [] })

    for (const key of ['auditorCount', 'daysPerAuditor', 'supervisorId', 'supervisorName', 'totalMandays']) {
      expect(payload).not.toHaveProperty(key)
    }
  })

  it('allows an empty activityType for now', () => {
    const payload = buildAnnualPlanPayload(form({ activityType: '' }), { quarters: [], attachments: [] })

    expect(payload.activityType).toBe('')
  })

  it('offers exactly the four Jenis Kegiatan Audit values', () => {
    expect([...ANNUAL_ACTIVITY_TYPE_OPTIONS]).toEqual(['Assurance', 'Audit Khusus', 'Audit Tujuan Tertentu', 'Audit Investigasi'])
  })
})

describe('row totals and number parsing', () => {
  it('sums Total Mandays over the rows', () => {
    expect(sumActivityMandays([{ totalMandays: 10 }, { totalMandays: 5 }, {}])).toBe(15)
    expect(sumActivityMandays([])).toBe(0)
    expect(sumActivityMandays(undefined)).toBe(0)
  })

  it('toCount parses number inputs and falls back on blanks', () => {
    expect(toCount('7')).toBe(7)
    expect(toCount('')).toBe(0)
    expect(toCount(null, 1)).toBe(1)
    expect(toCount(undefined, 1)).toBe(1)
  })
})

describe('the form template matches the Formulir Rencana Kegiatan Audit fields', () => {
  const src = readFileSync(resolve(__dirname, '../../components/annual-audit/AnnualAuditForm.vue'), 'utf-8')

  it.each([
    'annualAudit.form.year',
    'annualAudit.form.activityType',
    'annualAudit.form.planTitle',
    'annualAudit.form.assignmentCode',
    'annualAudit.form.objectName',
    'annualAudit.form.auditType',
    'annualAudit.form.auditableEntityLabel',
    'annualAudit.form.justification',
    'annualAudit.form.auditObjective',
    'annualAudit.form.auditFocus',
    'annualAudit.form.scheduleText',
    'annualAudit.form.teamCount',
    'annualAudit.form.totalMandays',
    'annualAudit.form.teamLead'
  ])('has the %s field', (key) => {
    expect(src).toContain(`t('${key}')`)
  })

  it('binds the Audit Universe picker and the row fields to the payload names', () => {
    expect(src).toContain('@update:model-value="(val: any) => handleSelectAuditableEntity(activity, val)"')
    expect(src).toContain('v-model="activity.justification"')
    expect(src).toContain('v-model="activity.auditObjective"')
    expect(src).toContain('v-model="activity.auditFocus"')
    expect(src).toContain('v-model.number="activity.auditorCount"')
    expect(src).toContain('v-model.number="activity.totalMandays"')
    expect(src).toContain('v-model="activity.supervisorName"')
    expect(src).toContain('v-model="activity.departmentId"')
  })
})
