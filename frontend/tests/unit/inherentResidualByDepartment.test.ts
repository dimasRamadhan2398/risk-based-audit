// @ts-nocheck
/**
 * "Inherent vs Residual Risk by Department" chart (pages/dashboard/index.vue).
 *
 * The two series must come from the same data the Corporate Risk Profile (CRP)
 * and the Risk Control Matrix (RCM) show:
 *   - department  → the risk's branch/department, as used by RCM's department filter
 *   - inherent    → the start-of-year (Q1) assessment score, impact × likelihood
 *   - residual    → the end-of-year (Q4) assessment score, after mitigation
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'

import { useRCMStore, computeInherentVsResidualByDepartment } from '~/stores/rcm'
import { useRiskProfileStore } from '~/stores/risk-profile'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: { riskServiceBaseUrl: 'http://localhost:8080/api/v1' }
  })
}))

global.$fetch = vi.fn()

const DASHBOARD_PAGE = resolve(__dirname, '../../pages/dashboard/index.vue')

/** CRP risks: two departments, one risk without an assessment for 2026. */
const crpRisks = [
  {
    id: 1,
    name: 'Target pendapatan tidak tercapai',
    category: 'Financial',
    branch: 'Head Office',
    impact: 5,
    likelihood: 4,
    assessments: [
      { year: 2026, impact_q1: 5, impact_q2: 4, impact_q3: 3, impact_q4: 2, likelihood_q1: 4, likelihood_q2: 3, likelihood_q3: 2, likelihood_q4: 2 }
    ]
  },
  {
    id: 2,
    name: 'Ancaman cyber security',
    category: 'Technology',
    branch: 'Head Office',
    impact: 4,
    likelihood: 4,
    assessments: [
      { year: 2026, impact_q1: 4, impact_q2: 4, impact_q3: 3, impact_q4: 3, likelihood_q1: 4, likelihood_q2: 3, likelihood_q3: 2, likelihood_q4: 1 }
    ]
  },
  {
    id: 3,
    name: 'Keterlambatan layanan cabang',
    category: 'Operational',
    branch: 'Bali Branch',
    impact: 3,
    likelihood: 2,
    assessments: [] // no assessment → both series fall back to the base score
  }
]

describe('computeInherentVsResidualByDepartment', () => {
  it('groups by the department/branch present in CRP, not a hardcoded list', () => {
    const rows = computeInherentVsResidualByDepartment(crpRisks, 2026)
    expect(rows.map((r) => r.name)).toEqual(['Head Office', 'Bali Branch'])
  })

  it('averages the Q1 score for inherent and the Q4 score for residual', () => {
    const rows = computeInherentVsResidualByDepartment(crpRisks, 2026)
    const headOffice = rows.find((r) => r.name === 'Head Office')

    // Q1: (5×4) and (4×4) → (20 + 16) / 2 = 18
    expect(headOffice.inherentRisk).toBe(18)
    // Q4: (2×2) and (3×1) → (4 + 3) / 2 = 3.5
    expect(headOffice.residualRisk).toBe(3.5)
    expect(headOffice.riskCount).toBe(2)
  })

  it('falls back to the base impact × likelihood when the year has no assessment', () => {
    const rows = computeInherentVsResidualByDepartment(crpRisks, 2026)
    const bali = rows.find((r) => r.name === 'Bali Branch')

    expect(bali.inherentRisk).toBe(6) // 3 × 2
    expect(bali.residualRisk).toBe(6) // unmitigated → residual equals inherent
  })

  it('only counts assessments of the requested year', () => {
    const rows = computeInherentVsResidualByDepartment(crpRisks, 2025)
    const headOffice = rows.find((r) => r.name === 'Head Office')

    // 2025 has no assessment → base scores: (5×4) and (4×4) → 18 for both series
    expect(headOffice.inherentRisk).toBe(18)
    expect(headOffice.residualRisk).toBe(18)
  })

  it('never fabricates residual as a fixed percentage of inherent', () => {
    const rows = computeInherentVsResidualByDepartment(crpRisks, 2026)
    for (const row of rows) {
      expect(row.residualRisk).not.toBeCloseTo(row.inherentRisk * 0.6, 5)
    }
  })

  it('tolerates empty or missing input', () => {
    expect(computeInherentVsResidualByDepartment([], 2026)).toEqual([])
    expect(computeInherentVsResidualByDepartment(undefined, 2026)).toEqual([])
  })

  it('skips risks with no department rather than inventing one', () => {
    const rows = computeInherentVsResidualByDepartment(
      [{ id: 9, impact: 4, likelihood: 4, assessments: [] }],
      2026
    )
    expect(rows).toEqual([])
  })
})

describe('RCM store exposes the chart series', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    if (typeof localStorage !== 'undefined') localStorage.clear()
  })

  it('derives the series from CRP risks for the selected year', () => {
    const riskStore = useRiskProfileStore()
    riskStore.rawRisks = crpRisks

    const rcmStore = useRCMStore()
    rcmStore.selectedYear = 2026

    const rows = rcmStore.inherentVsResidualByDepartment
    expect(rows.map((r) => r.name)).toEqual(['Head Office', 'Bali Branch'])
    expect(rows.find((r) => r.name === 'Head Office').residualRisk).toBe(3.5)
  })

  it('produces a bar per department for the seeded CRP data, none of them empty', async () => {
    const riskStore = useRiskProfileStore()
    const rcmStore = useRCMStore()

    // The store loads its risks on creation (API, falling back to seed data).
    await vi.waitFor(() => expect(riskStore.rawRisks.length).toBeGreaterThan(0))

    const rows = rcmStore.inherentVsResidualByDepartment
    expect(rows.length).toBeGreaterThan(0)
    // Every department on the chart comes from CRP and carries real exposure —
    // the old hardcoded list left most bars at zero.
    const branches = new Set(riskStore.rawRisks.map((r: any) => r.branch))
    for (const row of rows) {
      expect(branches.has(row.name)).toBe(true)
      expect(row.inherentRisk).toBeGreaterThan(0)
      expect(row.riskCount).toBeGreaterThan(0)
    }
  })

  it('agrees with the store\'s own inherent/residual risk counts', () => {
    const riskStore = useRiskProfileStore()
    riskStore.rawRisks = crpRisks

    const rcmStore = useRCMStore()
    rcmStore.selectedYear = 2026

    // Mitigation lowered the scores, so the chart must show residual below
    // inherent — the same direction the RCM summary cards report.
    const totalInherent = rcmStore.inherentVsResidualByDepartment.reduce((s, r) => s + r.inherentRisk, 0)
    const totalResidual = rcmStore.inherentVsResidualByDepartment.reduce((s, r) => s + r.residualRisk, 0)
    expect(totalResidual).toBeLessThan(totalInherent)
    expect(rcmStore.totalResidualRisk).toBeLessThanOrEqual(rcmStore.totalInherentRisk)
  })
})

describe('dashboard chart wiring', () => {
  const src = () => readFileSync(DASHBOARD_PAGE, 'utf-8')

  it('builds the chart from the RCM/CRP series', () => {
    expect(
      src().includes('inherentVsResidualByDepartment'),
      'dashboard/index.vue does not use the shared RCM/CRP series'
    ).toBe(true)
  })

  it('no longer groups by a hardcoded department list', () => {
    expect(
      /\["Finance", "IT", "Operations", "Legal", "HR"\]/.test(src()),
      'dashboard/index.vue still groups by a hardcoded department list'
    ).toBe(false)
  })

  it('no longer mocks residual risk as 60% of inherent', () => {
    expect(
      /\* 0\.6/.test(src()),
      'dashboard/index.vue still derives residual risk as 60% of inherent'
    ).toBe(false)
  })
})
