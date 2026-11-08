// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { RiskLevel } from '~/types/risk'
import { computeRcmRiskCounts, computeMitigationCoverage } from '~/utils/rcmRiskCounts'
import { RCM_ALL_BRANCHES, RCM_UNASSIGNED_BRANCH } from '~/utils/rcmBranch'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({ public: { riskServiceBaseUrl: 'http://localhost:8080/api/v1' } })
}))

// Same 5x5 matrix as stores/risk-profile.ts getRiskLevel.
const getRiskLevel = (likelihood?: number, impact?: number) => {
  if (!likelihood || !impact) return RiskLevel.LOW
  const m = {
    5: { 1: RiskLevel.LOW_MODERATE, 2: RiskLevel.MODERATE, 3: RiskLevel.MODERATE_HIGH, 4: RiskLevel.HIGH, 5: RiskLevel.HIGH },
    4: { 1: RiskLevel.LOW, 2: RiskLevel.LOW_MODERATE, 3: RiskLevel.MODERATE, 4: RiskLevel.MODERATE_HIGH, 5: RiskLevel.HIGH },
    3: { 1: RiskLevel.LOW, 2: RiskLevel.LOW_MODERATE, 3: RiskLevel.MODERATE, 4: RiskLevel.MODERATE_HIGH, 5: RiskLevel.HIGH },
    2: { 1: RiskLevel.LOW, 2: RiskLevel.LOW_MODERATE, 3: RiskLevel.LOW_MODERATE, 4: RiskLevel.MODERATE_HIGH, 5: RiskLevel.HIGH },
    1: { 1: RiskLevel.LOW, 2: RiskLevel.LOW, 3: RiskLevel.LOW_MODERATE, 4: RiskLevel.MODERATE, 5: RiskLevel.HIGH }
  }
  return m[likelihood]?.[impact] || RiskLevel.LOW
}

const locations = { 'loc-ho': 'Head Office', 'loc-bali': 'Bali Branch' }
const baseOpts = {
  year: 2026,
  quarter: 'Q4',
  branch: RCM_ALL_BRANCHES,
  department: 'All Departments',
  getRiskLevel,
  riskLocationId: r => r.location_id ?? null,
  locationName: id => locations[id]
}

// q1 / qN: [likelihood, impact]
const risk = (id, q1, q4, extra = {}, assessmentExtra = {}) => ({
  id,
  likelihood: 2,
  impact: 2,
  location_id: 'loc-ho',
  branch: 'Head Office',
  category: 'Financial',
  ...extra,
  assessments: [{
    year: 2026,
    likelihood_q1: q1?.[0] ?? 0,
    impact_q1: q1?.[1] ?? 0,
    likelihood_q4: q4?.[0] ?? 0,
    impact_q4: q4?.[1] ?? 0,
    ...assessmentExtra
  }]
})

describe('computeRcmRiskCounts', () => {
  it('inherent reads Q1 and residual reads the current quarter, not the other way round', () => {
    const risks = [
      risk('a', [4, 4], [2, 2]), // priority on Q1, mitigated by Q4
      risk('b', [4, 4], [4, 4]), // priority on both
      risk('c', [2, 2], [4, 4]), // low on Q1 -> never inherent, so never residual
      risk('d', [1, 1], [1, 1]) // low throughout
    ]
    const res = computeRcmRiskCounts(risks, baseOpts)
    expect(res.inherent).toBe(2)
    expect(res.residual).toBe(1)
    expect(res.residualRiskIds).toEqual(['b'])
  })

  it('uses the quarter it is given (residual follows the calendar quarter)', () => {
    const r = risk('a', [4, 4], [2, 2], {}, { likelihood_q2: 4, impact_q2: 4 })
    expect(computeRcmRiskCounts([r], { ...baseOpts, quarter: 'Q4' }).residual).toBe(0)
    expect(computeRcmRiskCounts([r], { ...baseOpts, quarter: 'Q2' }).residual).toBe(1)
  })

  it('falls back to the Q1 values when the current quarter is empty', () => {
    const r = risk('a', [4, 4], null) // Q4 not assessed yet
    expect(computeRcmRiskCounts([r], baseOpts)).toMatchObject({ inherent: 1, residual: 1 })
  })

  it('falls back to the register inherent likelihood/impact when Q1 is missing or 0', () => {
    const fromInherentFields = risk('a', null, [1, 1], { inherent_likelihood: 5, inherent_impact: 5 })
    const fromRiskFields = risk('b', null, null, { likelihood: 4, impact: 4 })
    const lowRegister = risk('c', null, null, { likelihood: 1, impact: 1 })
    const res = computeRcmRiskCounts([fromInherentFields, fromRiskFields, lowRegister], baseOpts)
    expect(res.inherentRiskIds).toEqual(['a', 'b'])
    // a: Q4 is low -> mitigated; b: Q4 empty -> falls back to its (register) Q1 -> still priority
    expect(res.residualRiskIds).toEqual(['b'])
  })

  it('does not count every risk as inherent when none is a priority', () => {
    const risks = [risk('a', [1, 1], [1, 1]), risk('b', [2, 2], [2, 2])]
    expect(computeRcmRiskCounts(risks, baseOpts)).toMatchObject({ inherent: 0, residual: 0 })
  })

  it('only counts risks assessed in the selected year', () => {
    const r = risk('a', [4, 4], [4, 4])
    expect(computeRcmRiskCounts([r], { ...baseOpts, year: 2027 }).inherent).toBe(0)
  })

  it('applies the branch and department filters on the year snapshot', () => {
    const risks = [
      risk('ho', [4, 4], [4, 4]),
      // Moved to Bali in the 2026 snapshot although the risk row says Head Office
      risk('moved', [4, 4], [4, 4], {}, { location_id: 'loc-bali', branch: 'Bali (old name)' }),
      risk('none', [4, 4], [2, 2], { location_id: null, branch: null, category: 'Compliance' })
    ]
    expect(computeRcmRiskCounts(risks, { ...baseOpts, branch: 'loc-bali' }).inherentRiskIds).toEqual(['moved'])
    expect(computeRcmRiskCounts(risks, { ...baseOpts, branch: 'loc-ho' }).inherentRiskIds).toEqual(['ho'])
    expect(computeRcmRiskCounts(risks, { ...baseOpts, branch: RCM_UNASSIGNED_BRANCH }).inherentRiskIds).toEqual(['none'])
    // Department matches the snapshot branch name (from the Location master) or the category
    expect(computeRcmRiskCounts(risks, { ...baseOpts, department: 'Bali Branch' }).inherentRiskIds).toEqual(['moved'])
    expect(computeRcmRiskCounts(risks, { ...baseOpts, department: 'Compliance' })).toMatchObject({ inherent: 1, residual: 0 })
  })

  it('reproduces the prod shape: 34 priority on Q1, 20 still priority on Q4', () => {
    const risks = [
      ...Array.from({ length: 20 }, (_, i) => risk(`keep-${i}`, [4, 4], [4, 3])),
      ...Array.from({ length: 14 }, (_, i) => risk(`mitigated-${i}`, [3, 3], [2, 2])),
      ...Array.from({ length: 2 }, (_, i) => risk(`low-${i}`, [2, 2], [2, 2]))
    ]
    expect(computeRcmRiskCounts(risks, baseOpts)).toMatchObject({ inherent: 34, residual: 20 })
  })
})

describe('computeMitigationCoverage', () => {
  it('counts residual risks with at least one non-deleted plan', () => {
    const plans = [
      { id: 'p1', riskId: 'a' },
      { id: 'p2', riskId: 'a' },
      { id: 'p3', risk_id: 'B' }, // snake_case and case-insensitive
      { id: 'p4', riskId: 'c', deleted_at: '2026-10-01T00:00:00Z' },
      { id: 'p5', riskId: 'not-residual' },
      { id: 'p6' } // no risk
    ]
    expect(computeMitigationCoverage(['a', 'b', 'c', 'd'], plans)).toEqual({ covered: 2, total: 4 })
    expect(computeMitigationCoverage([], plans)).toEqual({ covered: 0, total: 0 })
    expect(computeMitigationCoverage(['a'], null)).toEqual({ covered: 0, total: 1 })
  })
})

describe('RCM store wiring', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers({ toFake: ['Date'] })
    vi.setSystemTime(new Date('2026-11-04T10:00:00'))
    globalThis.$fetch = vi.fn(async () => ({ success: true, data: [] }))
  })

  it('counts from raw risks regardless of the CRP selected period and follows mitigation changes', async () => {
    const { useRiskProfileStore } = await import('~/stores/risk-profile')
    const { useMitigationStore } = await import('~/stores/mitigation-risk')
    const { useRCMStore } = await import('~/stores/rcm')
    const riskStore = useRiskProfileStore()
    const mitigationStore = useMitigationStore()
    const rcmStore = useRCMStore()
    await vi.waitFor(() => expect(riskStore.loading).toBe(false))

    riskStore.rawRisks = [
      { id: 'r1', likelihood: 3, impact: 3, assessments: [{ year: 2026, likelihood_q1: 4, impact_q1: 4, likelihood_q4: 4, impact_q4: 4 }] },
      { id: 'r2', likelihood: 3, impact: 3, assessments: [{ year: 2026, likelihood_q1: 4, impact_q1: 4, likelihood_q4: 1, impact_q4: 1 }] }
    ]
    rcmStore.selectedYear = 2026
    expect(rcmStore.residualQuarter).toBe('Q4')

    riskStore.selectedPeriod = 'Q4'
    expect(rcmStore.synchronizedRiskCounts).toEqual({ inherent: 2, residual: 1 })
    riskStore.selectedPeriod = 'Q1'
    expect(rcmStore.synchronizedRiskCounts).toEqual({ inherent: 2, residual: 1 })
    expect(rcmStore.internalControlEffectiveness).toBe(50)

    // Reassessing a quarter in the CRP updates the counts
    riskStore.rawRisks[1].assessments[0].likelihood_q4 = 5
    riskStore.rawRisks[1].assessments[0].impact_q4 = 5
    expect(rcmStore.synchronizedRiskCounts).toEqual({ inherent: 2, residual: 2 })

    // Mitigation coverage follows the store's list of all plans
    expect(rcmStore.mitigationCoverage).toEqual({ covered: 0, total: 2 })
    globalThis.$fetch = vi.fn(async () => ({ success: true, data: [{ id: 'p1', riskId: 'r1' }] }))
    await mitigationStore.fetchAllMitigations()
    expect(rcmStore.mitigationCoverage).toEqual({ covered: 1, total: 2 })

    // A plan added to r2 (fetchMitigations(riskId) after save) is merged in
    globalThis.$fetch = vi.fn(async () => ({ success: true, data: [{ id: 'p2', riskId: 'r2' }] }))
    await mitigationStore.fetchMitigations('r2')
    expect(rcmStore.mitigationCoverage).toEqual({ covered: 2, total: 2 })

    // Deleting r1's last plan (refetch for r1 returns nothing) removes its coverage
    globalThis.$fetch = vi.fn(async () => ({ success: true, data: [] }))
    await mitigationStore.fetchMitigations('r1')
    expect(rcmStore.mitigationCoverage).toEqual({ covered: 1, total: 2 })
  })

  it('reads residual from Q4 for a past year and from the current quarter otherwise', async () => {
    vi.setSystemTime(new Date('2027-05-10T10:00:00'))
    const { useRiskProfileStore } = await import('~/stores/risk-profile')
    const { useRCMStore } = await import('~/stores/rcm')
    const riskStore = useRiskProfileStore()
    const rcmStore = useRCMStore()
    await vi.waitFor(() => expect(riskStore.loading).toBe(false))

    riskStore.rawRisks = [
      { id: 'r1', assessments: [
        { year: 2026, likelihood_q1: 4, impact_q1: 4, likelihood_q2: 4, impact_q2: 4, likelihood_q4: 1, impact_q4: 1 },
        { year: 2027, likelihood_q1: 4, impact_q1: 4, likelihood_q2: 4, impact_q2: 4, likelihood_q4: 1, impact_q4: 1 }
      ] }
    ]

    rcmStore.selectedYear = 2026
    expect(rcmStore.residualIsYearEnd).toBe(true)
    expect(rcmStore.residualQuarter).toBe('Q4')
    expect(rcmStore.synchronizedRiskCounts).toEqual({ inherent: 1, residual: 0 })

    rcmStore.selectedYear = 2027
    expect(rcmStore.residualIsYearEnd).toBe(false)
    expect(rcmStore.residualQuarter).toBe('Q2')
    expect(rcmStore.synchronizedRiskCounts).toEqual({ inherent: 1, residual: 1 })
  })
})
