// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { useRiskProfileStore } from '~/stores/risk-profile'

const RISK_SCORING_PAGE = resolve(__dirname, '../../pages/analytics/ai/risk-scoring.vue')

describe('Entity Prediction Audit Universe & Pure Auditable Entity Sync', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    global.$fetch = vi.fn(async () => null)
  })

  it('verifies risk-scoring.vue integrates with Audit Universe and removes Type, Risk Category, and their filters', () => {
    const sfc = readFileSync(RISK_SCORING_PAGE, 'utf-8')

    // Must import and use useAuditUniverseStore
    expect(sfc).toContain('useAuditUniverseStore')
    expect(sfc).toContain('fetchYearlyUniverse')
    expect(sfc).toContain('fetchCorporateUniverse')

    // Must remove Type and Risk Category columns from table headers
    expect(sfc).not.toContain("t('analytics.xgboost.colType')")
    expect(sfc).not.toContain("t('analytics.xgboost.colRiskCategory')")

    // Must remove Tipe Entitas and Kategori Risiko filters
    expect(sfc).not.toContain('Tipe Entitas:')
    expect(sfc).not.toContain('Kategori Risiko:')
    expect(sfc).not.toContain('selectedType')
    expect(sfc).not.toContain('selectedCategory')

    // Description must mention Audit Universe & Audit Priority
    expect(sfc).toContain('Audit Universe')
    expect(sfc).toContain('Audit Priority')
  })

  it('syncs pure auditable entities strictly from Audit Universe and sets FY timeline without Corporate Risk Profile', async () => {
    const auditUniverseStore = useAuditUniverseStore()
    const riskStore = useRiskProfileStore()

    // Populate audit universe with yearly established entities having risk level from audit priority
    auditUniverseStore.yearlyUniverse = [
      {
        id: 'y-1',
        corporate_audit_universe_id: 'corp-fin',
        corporate_audit_universe: { id: 'corp-fin', name: 'Divisi Keuangan & Akuntansi' },
        risk_index: 85.5,
        risk_level: 'High',
        audit_priority: true
      },
      {
        id: 'y-2',
        corporate_audit_universe_id: 'corp-it',
        corporate_audit_universe: { id: 'corp-it', name: 'Divisi Teknologi Informasi' },
        risk_index: 62.0,
        risk_level: 'Medium to High',
        audit_priority: true
      }
    ]

    // Populate corporate universe with branches included as auditable entity nodes
    auditUniverseStore.corporateUniverse = [
      {
        id: 'corp-root',
        name: 'PT AIFL Indonesia (Holding)',
        children: [
          { id: 'corp-fin', name: 'Divisi Keuangan & Akuntansi' },
          { id: 'corp-it', name: 'Divisi Teknologi Informasi' },
          {
            id: 'corp-ops',
            name: 'Divisi Operasional & Rantai Pasok',
            children: [
              { id: 'corp-b-1', name: 'Cabang Bandung' },
              { id: 'corp-b-2', name: 'Cabang Surabaya' }
            ]
          }
        ]
      }
    ]

    // Risk Profile has risks that should NOT be ingested into Entity Prediction
    riskStore.rawRisks = [
      {
        id: 'r-isolated',
        name: 'Risk Exclusively in Corporate Risk Profile',
        entity: 'Excluded CRP Entity',
        branch: 'Excluded Branch',
        category: 'Financial',
        likelihood: 5,
        impact: 5
      }
    ]

    const { syncEntityPredictionWithAuditUniverse, syncEntityPredictionWithRiskProfile, xgboostState } = useAiAnalytics()

    await (syncEntityPredictionWithAuditUniverse || syncEntityPredictionWithRiskProfile)(undefined, false)

    const predictions = xgboostState.value.predictions
    expect(predictions.length).toBeGreaterThan(0)

    // 1. Verify entity names are PURE entities from Audit Universe
    const entityNames = predictions.map(p => p.entity)
    expect(entityNames).toContain('Divisi Keuangan & Akuntansi')
    expect(entityNames).toContain('Divisi Teknologi Informasi')
    expect(entityNames).toContain('Cabang Bandung')
    expect(entityNames).toContain('Cabang Surabaya')

    // Must NOT contain any CRP-only entities
    expect(entityNames).not.toContain('Excluded CRP Entity')
    expect(entityNames).not.toContain('Excluded Branch')
    expect(entityNames).not.toContain('Risk Exclusively in Corporate Risk Profile')

    // 2. Verify all predictions have FY target timeline (no Q4 / quarterly timeline)
    for (const p of predictions) {
      expect(p.type).not.toBe('Department')
      expect(['Branch', 'Auditable Entity']).toContain(p.type)
      expect(p.targetTimeline).toMatch(/^FY \d{4}$/)
      expect(p.targetTimeline).not.toContain('Q')
    }

    // 3. Verify risk level from Audit Priority was mapped
    const finEntity = predictions.find(p => p.entity === 'Divisi Keuangan & Akuntansi')
    expect(finEntity).toBeDefined()
    expect(finEntity?.type).toBe('Auditable Entity')
    expect(finEntity?.riskCategory).toBe('Financial')
    expect(finEntity?.predictedScore).toBeGreaterThanOrEqual(15) // High risk mapped from Audit Priority
  })

  it('persists predictions and does not revert to small stale batch on refresh or fetchAiAnalytics', async () => {
    const auditUniverseStore = useAuditUniverseStore()

    auditUniverseStore.yearlyUniverse = [
      {
        id: 'y-1',
        corporate_audit_universe_id: 'corp-fin',
        corporate_audit_universe: { id: 'corp-fin', name: 'Divisi Keuangan & Akuntansi' },
        risk_index: 85.5,
        risk_level: 'High',
        audit_priority: true
      }
    ]

    auditUniverseStore.corporateUniverse = [
      {
        id: 'corp-root',
        name: 'PT AIFL Indonesia (Holding)',
        children: [
          { id: 'corp-fin', name: 'Divisi Keuangan & Akuntansi' },
          { id: 'corp-it', name: 'Divisi Teknologi Informasi' }
        ]
      }
    ]

    const { syncEntityPredictionWithAuditUniverse, fetchAiAnalytics, xgboostState } = useAiAnalytics()

    // 1. Sync Audit Universe
    await syncEntityPredictionWithAuditUniverse(undefined, false)
    const initialCount = xgboostState.value.predictions.length
    expect(initialCount).toBeGreaterThanOrEqual(2)

    // 2. Simulate subsequent fetchAiAnalytics returning empty or smaller batch
    global.$fetch = vi.fn(async (url: string) => {
      if (url.includes('/risk-score/batch')) {
        return { predictions: [] }
      }
      return null
    })

    await fetchAiAnalytics(true)

    // Predictions must NOT be wiped or degraded to empty or old small batch
    expect(xgboostState.value.predictions.length).toBe(initialCount)
    for (const p of xgboostState.value.predictions) {
      expect(p.targetTimeline).toMatch(/^FY \d{4}$/)
    }
  })
})
