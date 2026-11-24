import { describe, expect, it } from 'vitest'
import { summarizeAnnualPlan } from '~/utils/annualPlanSummary'
import { buildAnnualActivityPayload } from '~/utils/annualPlanPayload'

describe('summarizeAnnualPlan', () => {
  it('rolls up 40 sparse activities without per-row output', () => {
    const activities = Array.from({ length: 40 }, (_, i) => ({ itemNumber: String(i + 1), auditorCount: 2, totalMandays: 5, riskLevel: '', auditUniverseId: '' }))
    const s = summarizeAnnualPlan({ activities, selectedMonths: [] })
    expect(s).toMatchObject({ activityCount: 40, totalMandays: 200, totalAuditors: 80, withRisk: 0, riskLevels: [], months: [] })
  })

  it('counts risk levels and linked risks, and falls back to activity months', () => {
    const s = summarizeAnnualPlan({
      activities: [
        { riskLevel: 'High', riskName: 'R1', months: [2, 0] },
        { riskLevel: 'High', auditUniverseId: 'u1' },
        { riskLevel: 'Low' }
      ]
    })
    expect(s.withRisk).toBe(2)
    expect(s.riskLevels).toEqual([{ level: 'High', count: 2 }, { level: 'Low', count: 1 }])
    expect(s.months).toEqual([0, 2])
  })
})

describe('buildAnnualActivityPayload extras', () => {
  it('keeps months, durationDays and remarks', () => {
    const p = buildAnnualActivityPayload({ name: 'x', months: [1], durationDays: 3, remarks: 'r' } as never, 0)
    expect(p).toMatchObject({ itemNumber: '1', months: [1], durationDays: 3, remarks: 'r' })
  })
})
