// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'
import {
  ENTITY_RISK_LEVELS,
  normalizeEntityRiskLevel,
  toEntityRiskCode
} from '~/utils/riskLevelBadge'
import {
  useAiAnalytics,
  getPersistedCustomEntityPredictions,
  savePersistedCustomEntityPredictions
} from '~/composables/useAiAnalytics'
import { useAuditUniverseStore } from '~/stores/audit-universe'

const CANONICAL_5_LABELS = ['High', 'Moderate to High', 'Moderate', 'Low to Moderate', 'Low']
const CANONICAL_5_CODES = ['HIGH', 'MODERATE_HIGH', 'MODERATE', 'LOW_MODERATE', 'LOW']

const RISK_SCORING_PAGE = resolve(__dirname, '../../pages/analytics/ai/risk-scoring.vue')
const MAPPED_BANNER = resolve(__dirname, '../../components/common/MappedSourceDataBanner.vue')

describe('Entity Prediction 5-Tier Risk Level Enforcement', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    global.$fetch = vi.fn(async () => null)
    const { xgboostState } = useAiAnalytics()
    xgboostState.value.predictions = []
  })

  it('exposes strictly the 5 canonical Entity Prediction risk levels and codes', () => {
    expect([...ENTITY_RISK_LEVELS]).toEqual(CANONICAL_5_LABELS)
    expect(CANONICAL_5_LABELS).toHaveLength(5)
  })

  it('normalizes legacy and external DB custom risk levels onto strictly the 5 canonical levels', () => {
    const cases: Array<[unknown, string, string]> = [
      ['High', 'High', 'HIGH'],
      ['Moderate to High', 'Moderate to High', 'MODERATE_HIGH'],
      ['Moderate', 'Moderate', 'MODERATE'],
      ['Low to Moderate', 'Low to Moderate', 'LOW_MODERATE'],
      ['Low', 'Low', 'LOW'],
      ['Critical', 'High', 'HIGH'],
      ['CRITICAL', 'High', 'HIGH'],
      ['Extreme', 'High', 'HIGH'],
      ['Very High', 'High', 'HIGH'],
      ['Catastrophic', 'High', 'HIGH'],
      ['Severe', 'High', 'HIGH'],
      ['Blocker', 'High', 'HIGH'],
      ['Fatal', 'High', 'HIGH'],
      ['Risiko Tinggi', 'High', 'HIGH'],
      ['Sangat Tinggi', 'High', 'HIGH'],
      ['Macet', 'High', 'HIGH'],
      ['KOL-5', 'High', 'HIGH'],
      ['P1', 'High', 'HIGH'],
      ['Medium to High', 'Moderate to High', 'MODERATE_HIGH'],
      ['MEDIUM_TO_HIGH', 'Moderate to High', 'MODERATE_HIGH'],
      ['Major', 'Moderate to High', 'MODERATE_HIGH'],
      ['Elevated', 'Moderate to High', 'MODERATE_HIGH'],
      ['Significant', 'Moderate to High', 'MODERATE_HIGH'],
      ['KOL-3', 'Moderate to High', 'MODERATE_HIGH'],
      ['Medium', 'Moderate', 'MODERATE'],
      ['MEDIUM', 'Moderate', 'MODERATE'],
      ['Watch', 'Moderate', 'MODERATE'],
      ['Not Scored', 'Moderate', 'MODERATE'],
      ['Unknown', 'Moderate', 'MODERATE'],
      ['Low to Medium', 'Low to Moderate', 'LOW_MODERATE'],
      ['LOW_TO_MEDIUM', 'Low to Moderate', 'LOW_MODERATE'],
      ['Minor', 'Low to Moderate', 'LOW_MODERATE'],
      ['Guarded', 'Low to Moderate', 'LOW_MODERATE'],
      ['DPK', 'Low to Moderate', 'LOW_MODERATE'],
      ['KOL-2', 'Low to Moderate', 'LOW_MODERATE'],
      ['Insignificant', 'Low', 'LOW'],
      ['Negligible', 'Low', 'LOW'],
      ['Minimal', 'Low', 'LOW'],
      ['Trivial', 'Low', 'LOW'],
      ['Safe', 'Low', 'LOW'],
      ['Lancar', 'Low', 'LOW'],
      ['KOL-1', 'Low', 'LOW'],
      [25, 'High', 'HIGH'],
      [16, 'Moderate to High', 'MODERATE_HIGH'],
      [12, 'Moderate', 'MODERATE'],
      [6, 'Low to Moderate', 'LOW_MODERATE'],
      [2, 'Low to Moderate', 'LOW_MODERATE'],
      [1, 'Low', 'LOW'],
      ['', 'Moderate', 'MODERATE'],
      [null, 'Moderate', 'MODERATE']
    ]

    const { getRiskConfig } = useAiAnalytics()
    for (const [input, expectedLabel, expectedCode] of cases) {
      const label = normalizeEntityRiskLevel(input)
      const code = toEntityRiskCode(input)
      const cfg = getRiskConfig(String(input ?? ''))

      expect(label).toBe(expectedLabel)
      expect(code).toBe(expectedCode)
      expect(CANONICAL_5_LABELS).toContain(cfg.label)
      expect(cfg.label).toBe(expectedLabel)
    }
  })

  it('normalizes external DB items passed to syncEntityPredictionWithRiskProfile onto the 5 canonical levels', async () => {
    const { syncEntityPredictionWithRiskProfile, xgboostState, getRiskConfig } = useAiAnalytics()

    const externalItems = [
      { entity: 'Cabang A (External)', current_risk_level: 'Critical', risk_category: 'Financial', source_id: 'external_db:a' },
      { entity: 'Cabang B (External)', current_risk_level: 'Extreme', risk_category: 'Operational', source_id: 'external_db:b' },
      { entity: 'Unit C (External)', current_risk_level: 'Major', risk_category: 'Technology', source_id: 'external_db:c' },
      { entity: 'Unit D (External)', current_risk_level: 'Minor', risk_category: 'Compliance', source_id: 'external_db:d' },
      { entity: 'Unit E (External)', current_risk_level: 'Insignificant', risk_category: 'Operational', source_id: 'external_db:e' },
      { entity: 'Unit F (External)', current_risk_level: 'Not Scored', risk_category: 'Operational', source_id: 'external_db:f' },
      { entity: 'Unit G (External)', current_risk_level: 'MEDIUM_TO_HIGH', risk_category: 'Financial', source_id: 'external_db:g' }
    ]

    await syncEntityPredictionWithRiskProfile(externalItems, false)

    const preds = xgboostState.value.predictions
    expect(preds).toHaveLength(7)

    const actualByEntity = Object.fromEntries(preds.map((p: any) => [p.entity, p.actualRiskLevel]))
    expect(actualByEntity['Cabang A (External)']).toBe('HIGH')
    expect(actualByEntity['Cabang B (External)']).toBe('HIGH')
    expect(actualByEntity['Unit C (External)']).toBe('MODERATE_HIGH')
    expect(actualByEntity['Unit D (External)']).toBe('LOW_MODERATE')
    expect(actualByEntity['Unit E (External)']).toBe('LOW')
    expect(actualByEntity['Unit F (External)']).toBe('MODERATE')
    expect(actualByEntity['Unit G (External)']).toBe('MODERATE_HIGH')

    for (const p of preds) {
      expect(CANONICAL_5_CODES).toContain(p.actualRiskLevel)
      expect(CANONICAL_5_CODES).toContain(p.predictedRiskLevel)
      expect(CANONICAL_5_LABELS).toContain(getRiskConfig(p.actualRiskLevel).label)
      expect(CANONICAL_5_LABELS).toContain(getRiskConfig(p.predictedRiskLevel).label)
      expect(p.currentLikelihood).toBeGreaterThanOrEqual(1)
      expect(p.currentLikelihood).toBeLessThanOrEqual(5)
      expect(p.currentImpact).toBeGreaterThanOrEqual(1)
      expect(p.currentImpact).toBeLessThanOrEqual(5)
    }
  })

  it('normalizes Audit Universe sync outputs without MEDIUM_TO_HIGH / MEDIUM / LOW_TO_MEDIUM / Critical', async () => {
    const store = useAuditUniverseStore()
    store.yearlyUniverse = [
      { id: '1', corporate_audit_universe_id: 'c1', corporate_audit_universe: { id: 'c1', name: 'Unit Kritis' }, risk_index: 92, risk_level: 'Critical', audit_priority: true },
      { id: '2', corporate_audit_universe_id: 'c2', corporate_audit_universe: { id: 'c2', name: 'Unit MedHigh' }, risk_index: 68, risk_level: 'Medium to High', audit_priority: true },
      { id: '3', corporate_audit_universe_id: 'c3', corporate_audit_universe: { id: 'c3', name: 'Unit Med' }, risk_index: 48, risk_level: 'Medium', audit_priority: false },
      { id: '4', corporate_audit_universe_id: 'c4', corporate_audit_universe: { id: 'c4', name: 'Unit LowMed' }, risk_index: 25, risk_level: 'Low to Medium', audit_priority: false },
      { id: '5', corporate_audit_universe_id: 'c5', corporate_audit_universe: { id: 'c5', name: 'Unit Insignif' }, risk_index: 8, risk_level: 'Insignificant', audit_priority: false }
    ]
    store.corporateUniverse = [
      {
        id: 'root',
        name: 'Holding',
        children: [
          { id: 'c1', name: 'Unit Kritis' },
          { id: 'c2', name: 'Unit MedHigh' },
          { id: 'c3', name: 'Unit Med' },
          { id: 'c4', name: 'Unit LowMed' },
          { id: 'c5', name: 'Unit Insignif' }
        ]
      }
    ]

    const { syncEntityPredictionWithAuditUniverse, xgboostState } = useAiAnalytics()
    await syncEntityPredictionWithAuditUniverse(undefined, false)

    for (const p of xgboostState.value.predictions) {
      expect(CANONICAL_5_CODES).toContain(p.actualRiskLevel)
      expect(CANONICAL_5_CODES).toContain(p.predictedRiskLevel)
      expect(['MEDIUM_TO_HIGH', 'MEDIUM', 'LOW_TO_MEDIUM', 'CRITICAL', 'EXTREME', 'INSIGNIFICANT']).not.toContain(p.actualRiskLevel)
      expect(['MEDIUM_TO_HIGH', 'MEDIUM', 'LOW_TO_MEDIUM', 'CRITICAL', 'EXTREME', 'INSIGNIFICANT']).not.toContain(p.predictedRiskLevel)
    }

    const byName = Object.fromEntries(xgboostState.value.predictions.map((p: any) => [p.entity, p.actualRiskLevel]))
    expect(byName['Unit Kritis']).toBe('HIGH')
    expect(byName['Unit MedHigh']).toBe('MODERATE_HIGH')
    expect(byName['Unit Med']).toBe('MODERATE')
    expect(byName['Unit LowMed']).toBe('LOW_MODERATE')
    expect(byName['Unit Insignif']).toBe('LOW')
  })

  it('sanitizes stale non-canonical risk levels in persisted custom storage and backend batch hydration', async () => {
    savePersistedCustomEntityPredictions([
      {
        entity: 'Stale Critical Branch',
        riskCategory: 'Financial',
        actualRiskLevel: 'Critical',
        predictedRiskLevel: 'Extreme',
        currentLikelihood: 8,
        currentImpact: 9
      },
      {
        entity: 'Stale Medium Unit',
        riskCategory: 'Operational',
        actualRiskLevel: 'MEDIUM_TO_HIGH',
        predictedRiskLevel: 'LOW_TO_MEDIUM'
      }
    ])

    const persisted = getPersistedCustomEntityPredictions()
    expect(persisted).toHaveLength(2)
    expect(persisted.find((p: any) => p.entity === 'Stale Critical Branch')?.actualRiskLevel).toBe('HIGH')
    expect(persisted.find((p: any) => p.entity === 'Stale Critical Branch')?.predictedRiskLevel).toBe('HIGH')
    expect(persisted.find((p: any) => p.entity === 'Stale Medium Unit')?.actualRiskLevel).toBe('MODERATE_HIGH')
    expect(persisted.find((p: any) => p.entity === 'Stale Medium Unit')?.predictedRiskLevel).toBe('LOW_MODERATE')

    global.$fetch = vi.fn(async (url: string) => {
      if (url.includes('/risk-score/batch')) {
        return {
          predictions: [
            {
              entity: 'API Legacy Unit',
              type: 'Branch',
              risk_category: 'Financial',
              actual_risk_level: 'Insignificant',
              risk_level: 'Catastrophic'
            }
          ]
        }
      }
      return null
    })

    const { fetchAiAnalytics, xgboostState } = useAiAnalytics()
    await fetchAiAnalytics(true, { skipAutoRetrain: true })

    for (const p of xgboostState.value.predictions) {
      expect(CANONICAL_5_CODES).toContain(p.actualRiskLevel)
      expect(CANONICAL_5_CODES).toContain(p.predictedRiskLevel)
    }
    const apiUnit = xgboostState.value.predictions.find((p: any) => p.entity === 'API Legacy Unit')
    expect(apiUnit?.actualRiskLevel).toBe('LOW')
    expect(apiUnit?.predictedRiskLevel).toBe('HIGH')
  })

  it('verifies risk-scoring.vue tooltips and MappedSourceDataBanner.vue use canonical 5-tier normalization', () => {
    const pageSrc = readFileSync(RISK_SCORING_PAGE, 'utf-8')
    expect(pageSrc).toContain('getRiskConfig(item.predictedRiskLevel).label')
    expect(pageSrc).toContain('getRiskConfig(item.actualRiskLevel).label')

    const bannerSrc = readFileSync(MAPPED_BANNER, 'utf-8')
    expect(bannerSrc).toContain('normalizeEntityRiskLevel')
    expect(bannerSrc).toContain('toEntityRiskCode')
    expect(bannerSrc).toContain('extractExternalRiskRaw')
  })
})
