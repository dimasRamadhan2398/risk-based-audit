import { describe, it, expect, beforeEach, vi } from 'vitest'
import fs from 'fs'
import path from 'path'
import { setActivePinia, createPinia } from 'pinia'
import {
  buildScoringRows,
  computeRiskIndexPreview,
  isUniverseYearPrioritized,
  isUniverseYearScored,
  isValidWeightTotal,
  normalizeWeightPercent,
  sortUniverseYearsByRisk,
  sumWeightPercents,
  weightFractionToPercent
} from '~/utils/riskScoring'
import { getRiskLevelForIndex, isPrioritizedRiskLevel } from '~/utils/riskLevelBadge'

vi.mock('~/composables/useApiUrl', () => ({
  getRiskServiceBaseUrl: () => 'http://localhost:8004/api/v1'
}))

const rootDir = path.resolve(__dirname, '../../')
const read = (rel: string) => fs.readFileSync(path.join(rootDir, rel), 'utf8')

const corporateFactors = [
  { id: 'cf-1', weight: 0.5, standard_risk_factor: { name: 'Financial' } },
  { id: 'cf-2', weight: 0.3, standard_risk_factor: { name: 'Compliance' } },
  { id: 'cf-3', weight: 0.2, standard_risk_factor: { name: 'IT' } }
]

describe('scoring rows: no default-3 fallback', () => {
  it('leaves a factor without a saved score row as null (not scored)', () => {
    const rows = buildScoringRows(corporateFactors, [{ corporate_risk_factor_id: 'cf-1', score: 4 }])
    expect(rows.map(r => r.score)).toEqual([4, null, null])
    expect(rows.some(r => r.score === 3)).toBe(false)
  })

  it('an entity with no score rows has every factor unscored', () => {
    expect(buildScoringRows(corporateFactors, undefined).every(r => r.score === null)).toBe(true)
  })

  it('the Risk Factors page no longer defaults a missing score to 3 or computes its own thresholds', () => {
    const page = read('pages/risk-profile/risk-factors/index.vue')
    expect(page).not.toMatch(/matchScore\s*\?\s*matchScore\.score\s*:\s*3/)
    expect(page).toContain('buildScoringRows(')
    expect(page).toContain('computeRiskIndexPreview(')
    expect(page).not.toMatch(/risk_index\s*>=\s*80/)
    expect(page).not.toContain('risk_level = \'High\'')
  })
})

describe('risk index preview mirrors the backend formula', () => {
  it('Σ(weight × score) / Σ(weight) / 5 × 100 over scored factors only', () => {
    const preview = computeRiskIndexPreview([
      { weight: 0.5, score: 5 },
      { weight: 0.3, score: null },
      { weight: 0.2, score: 2 }
    ])
    // (0.5×5 + 0.2×2) / (0.5 + 0.2) = 2.9 / 0.7 → 82.857% → 82.9; the unscored factor is left out, not counted as 0
    expect(preview.weightedSum).toBeCloseTo(2.9)
    expect(preview.riskIndex).toBe(82.9)
    expect(preview.riskLevel).toBe('High')
    expect(preview.auditPriority).toBe(true)
    expect(preview.scoredCount).toBe(2)
    expect(preview.totalCount).toBe(3)
  })

  it('no scored factor gives index 0', () => {
    expect(computeRiskIndexPreview([{ weight: 1, score: null }]).riskIndex).toBe(0)
  })

  it('rounds to 1 decimal before choosing the level', () => {
    // (0.998×4 + 0.002×3) / 1 = 3.998 → 79.96% → 80.0 → High
    const preview = computeRiskIndexPreview([{ weight: 0.998, score: 4 }, { weight: 0.002, score: 3 }])
    expect(preview.riskIndex).toBe(80)
    expect(preview.riskLevel).toBe('High')
    expect(preview.auditPriority).toBe(true)
  })

  it('shared thresholds: level by index and prioritized levels', () => {
    expect([100, 80, 79.9, 60, 59.9, 40, 39.9, 20, 19.9, 0].map(getRiskLevelForIndex)).toEqual([
      'High', 'High', 'Medium to High', 'Medium to High', 'Medium', 'Medium', 'Low to Medium', 'Low to Medium', 'Low', 'Low'
    ])
    expect(isPrioritizedRiskLevel('High')).toBe(true)
    expect(isPrioritizedRiskLevel('Medium to High')).toBe(true)
    expect(isPrioritizedRiskLevel('Medium')).toBe(false)
    expect(isPrioritizedRiskLevel('')).toBe(false)
  })
})

describe('not-scored rows (scored_at === null)', () => {
  const rows = [
    { id: 'a', scored_at: null, risk_index: 0, risk_level: '', audit_priority: false },
    { id: 'b', scored_at: '2026-10-01T00:00:00Z', risk_index: 45, risk_level: 'Medium', audit_priority: false },
    { id: 'c', scored_at: null, risk_index: 0, risk_level: '', audit_priority: true },
    { id: 'd', scored_at: '2026-10-01T00:00:00Z', risk_index: 85, risk_level: 'High', audit_priority: true },
    { id: 'e', risk_index: 15, risk_level: 'Low', audit_priority: false } // older backend: no scored_at
  ]

  it('treats only an explicit null as not scored', () => {
    expect(rows.map(isUniverseYearScored)).toEqual([false, true, false, true, true])
  })

  it('never counts an unscored row as prioritized', () => {
    expect(rows.filter(isUniverseYearPrioritized).map(r => r.id)).toEqual(['d'])
  })

  it('sorts scored rows by risk index (desc) and unscored rows after them', () => {
    expect(sortUniverseYearsByRisk(rows).map(r => r.id)).toEqual(['d', 'b', 'e', 'a', 'c'])
    expect(rows.map(r => r.id)).toEqual(['a', 'b', 'c', 'd', 'e']) // input not mutated
  })

  it('the Audit Priority page renders the Not scored badge and both locales have the keys', () => {
    const page = read('pages/risk-profile/audit-priority/index.vue')
    expect(page).toContain('t(\'riskFactors.priority.notScoredBadge\')')
    expect(page).toContain('isUniverseYearScored(row.original)')
    expect(page).toContain('isUniverseYearPrioritized(row.original)')

    const en = JSON.parse(read('locales/en/common.json'))
    const id = JSON.parse(read('locales/id/common.json'))
    expect(en.riskFactors.priority.notScoredBadge).toBe('Not scored')
    expect(id.riskFactors.priority.notScoredBadge).toBe('Belum dinilai')
    for (const key of ['notScored', 'notScoredHint', 'entityNotScored', 'previewLabel', 'previewHint', 'unscoredFactorsCount', 'noScoresToSave']) {
      expect(en.riskFactors.scoring[key], `en scoring.${key}`).toBeTruthy()
      expect(id.riskFactors.scoring[key], `id scoring.${key}`).toBeTruthy()
    }
  })
})

describe('corporate weights', () => {
  it('keeps up to 2 decimals instead of whole percent', () => {
    expect(weightFractionToPercent(0.1525)).toBe(15.25)
    expect(weightFractionToPercent(0.123456)).toBe(12.35)
    expect(normalizeWeightPercent('33.333')).toBe(33.33)
  })

  it('accepts a total within 99.9–100.1 like the backend', () => {
    expect(isValidWeightTotal(sumWeightPercents([33.33, 33.33, 33.33]))).toBe(true) // 99.99
    expect(isValidWeightTotal(sumWeightPercents([33.3, 33.3, 33.3]))).toBe(true) // 99.9
    expect(isValidWeightTotal(sumWeightPercents([50.05, 50.05]))).toBe(true) // 100.1
    expect(isValidWeightTotal(100)).toBe(true)
    expect(isValidWeightTotal(sumWeightPercents([33.3, 33.3, 33.29]))).toBe(false) // 99.89
    expect(isValidWeightTotal(100.11)).toBe(false)
    expect(isValidWeightTotal(NaN)).toBe(false)
  })

  it('sums without floating point noise', () => {
    expect(sumWeightPercents([0.1, 0.2, 99.7])).toBe(100)
    expect(sumWeightPercents(['10', '', null, 90])).toBe(100)
  })
})

describe('deleting a standard risk factor that still has a corporate weight (409)', () => {
  const fetchMock = vi.fn()

  beforeEach(() => {
    setActivePinia(createPinia())
    fetchMock.mockReset()
    vi.stubGlobal('$fetch', fetchMock)
  })

  it('keeps the factor and exposes the backend message', async () => {
    const { useRiskFactorsStore } = await import('~/stores/risk-factors')
    const store = useRiskFactorsStore()
    store.standardFactors = [{ id: 'sf-1', name: 'Financial' }, { id: 'sf-2', name: 'IT' }] as never
    store.corporateFactors = [{ id: 'cf-1', standard_risk_factor_id: 'sf-1', weight: 1 }] as never

    const conflict = Object.assign(new Error('409 Conflict'), {
      statusCode: 409,
      data: { success: false, error: { message: 'Risk factor is still weighted in the corporate risk factors; remove its weight first' } }
    })
    fetchMock.mockRejectedValueOnce(conflict)

    const ok = await store.deleteStandardFactor('sf-1')

    expect(ok).toBe(false)
    expect(store.errorMsg).toBe('Risk factor is still weighted in the corporate risk factors; remove its weight first')
    expect(store.standardFactors.map(f => f.id)).toEqual(['sf-1', 'sf-2'])
    expect(store.corporateFactors.map(f => f.id)).toEqual(['cf-1'])
    expect(fetchMock).toHaveBeenCalledTimes(1) // no refetch that could hide the failure
  })

  it('the page removes the factor from the weights list only inside the success branch and toasts store.errorMsg otherwise', () => {
    const page = read('pages/risk-profile/risk-factors/index.vue')
    const fn = page.slice(page.indexOf('const confirmDeleteFactor'), page.indexOf('const getScoreColor'))
    const successIdx = fn.indexOf('if (success) {')
    const elseIdx = fn.indexOf('} else {')
    const spliceIdx = fn.indexOf('selectedCorporateList.value.splice')
    expect(successIdx).toBeGreaterThan(-1)
    expect(spliceIdx).toBeGreaterThan(successIdx)
    expect(spliceIdx).toBeLessThan(elseIdx)
    expect(fn.slice(elseIdx)).toContain('toast.showError(store.errorMsg')
  })
})
