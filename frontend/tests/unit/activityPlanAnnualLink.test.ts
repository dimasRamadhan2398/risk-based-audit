// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'
import { useActivityPlanStore } from '~/stores/activity-plan'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { RiskLevel } from '~/types/risk'

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn()
  })
}))

const FORM_SFC = resolve(__dirname, '../../components/audit-activity-plan/ActivityPlanForm.vue')

const annualPlan = {
  id: 'plan-1',
  code: 'PKAT-2026-001',
  year: '2026',
  isActive: true,
  status: 'Approved',
  activities: [{
    id: 'row-1',
    activityCode: 'ASR-2026-001',
    name: 'Keuangan Cabang',
    category: 'Assurance',
    departmentId: 'dept-fin',
    department: 'Finance',
    auditUniverseId: 'au-fin',
    riskName: 'Risiko Kas',
    riskLevel: 'High',
    auditorCount: 3,
    totalMandays: 30
  }]
}

describe('Activity plan linked to the Annual Audit Plan', () => {
  let store: ReturnType<typeof useActivityPlanStore>
  let annual: ReturnType<typeof useAnnualPlanStore>
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    fetchMock = vi.fn(async () => ({ success: true, data: { items: [], pagination: { total_pages: 1 } } }))
    global.$fetch = fetchMock
    setActivePinia(createPinia())
    annual = useAnnualPlanStore()
    annual.plans = [annualPlan]
    store = useActivityPlanStore()
    store.openModal()
    store.addPlannedActivity()
  })

  it('offers the annual activities as link options', () => {
    expect(store.annualActivityOptions).toHaveLength(1)
    expect(store.annualActivityOptions[0].label).toContain('PKAT-2026-001')
    expect(store.annualActivityOptions[0].label).toContain('ASR-2026-001')
  })

  it('linking copies the annual values into the row and sets the plan year and department', () => {
    store.linkAnnualActivity(0, 'row-1')

    const row = store.formState.plannedActivities[0]
    expect(row.annualPlanId).toBe('plan-1')
    expect(row.annualActivityId).toBe('row-1')
    expect(row.category).toBe('Assurance')
    expect(row.riskName).toBe('Risiko Kas')
    expect(row.riskLevel).toBe(RiskLevel.HIGH)
    expect(row.numberOfAuditors).toBe(3)
    expect(row.duration).toBe(30)
    expect(store.formState.planYear).toBe('2026')
    expect(store.formState.department).toBe('Finance')
    expect(store.unlinkedActivityCount).toBe(0)
  })

  it('a row without a link is counted as unlinked', () => {
    expect(store.unlinkedActivityCount).toBe(1)
  })

  it('save sends the link on each planned activity', async () => {
    store.linkAnnualActivity(0, 'row-1')
    store.formState.planTitle = 'Audit Activity Plan 2026'
    store.formState.planPeriodStart = '2026-01-01'
    store.formState.planPeriodEnd = '2026-12-31'
    store.formState.plannedActivities[0].auditName = 'Audit kas'
    store.formState.plannedActivities[0].auditee = 'Kasir'

    await store.savePlan()

    const postCall = fetchMock.mock.calls.find(([, opts]) => opts?.method === 'POST')
    expect(postCall).toBeDefined()
    const [planned] = postCall[1].body.plannedActivities
    expect(planned.annualActivityId).toBe('row-1')
    expect(planned.annualPlanId).toBe('plan-1')
  })

  it('the form shows category, risk, department, year, team and mandays read-only', () => {
    const src = readFileSync(FORM_SFC, 'utf-8')

    expect(src).not.toContain('v-model="store.formState.planYear"')
    expect(src).not.toContain('v-model="store.formState.department"')
    expect(src).not.toContain('v-model="activity.category"')
    expect(src).not.toContain('v-model="activity.duration"')
    expect(src).not.toContain('v-model="activity.numberOfAuditors"')
    expect(src).not.toContain('v-model="activity.riskLevel"')
    // The fields that stay editable are the activity-level ones.
    expect(src).toContain('v-model="activity.auditName"')
    expect(src).toContain('v-model="activity.priority"')
    expect(src).toContain('v-model="activity.estimatedSchedule"')
    expect(src).toContain('v-model="activity.budgetEstimation"')
  })
})
