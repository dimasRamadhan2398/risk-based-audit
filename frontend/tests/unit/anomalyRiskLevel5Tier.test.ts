// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import {
  ANOMALY_RISK_LEVELS,
  normalizeAnomalyRiskLevel,
  toAnomalyRiskCode
} from '~/utils/riskLevelBadge'
import {
  useAiAnalytics,
  getPersistedCustomAnomalies
} from '~/composables/useAiAnalytics'
import { BANK_ANOMALIES_SEED } from '~/composables/bankAnomalyDataset'

describe('Anomaly Detection 5-Tier Risk Level Enforcement (No Critical)', () => {
  let localStorageMock: Record<string, string> = {}

  beforeEach(() => {
    localStorageMock = {}
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => localStorageMock[key] ?? null,
      setItem: (key: string, val: string) => { localStorageMock[key] = String(val) },
      removeItem: (key: string) => { delete localStorageMock[key] },
      clear: () => { localStorageMock = {} }
    })
  })

  it('defines strictly the 5 allowed risk levels and never Critical', () => {
    expect(ANOMALY_RISK_LEVELS).toEqual([
      'High',
      'Moderate to High',
      'Moderate',
      'Low to Moderate',
      'Low'
    ])
    expect(ANOMALY_RISK_LEVELS).not.toContain('Critical')
  })

  it.each([
    ['Critical', 'High'],
    ['CRITICAL', 'High'],
    ['kritis', 'High'],
    ['Extreme', 'High'],
    ['Very High', 'High'],
    ['Catastrophic', 'High'],
    ['High', 'High'],
    ['HIGH', 'High'],
    ['KOL-5', 'High'],
    ['Moderate to High', 'Moderate to High'],
    ['MODERATE_HIGH', 'Moderate to High'],
    ['Medium to High', 'Moderate to High'],
    ['Elevated', 'Moderate to High'],
    ['Major', 'Moderate to High'],
    ['Sedang ke Tinggi', 'Moderate to High'],
    ['KOL-3', 'Moderate to High'],
    ['Moderate', 'Moderate'],
    ['MODERATE', 'Moderate'],
    ['Medium', 'Moderate'],
    ['Sedang', 'Moderate'],
    ['Watch', 'Moderate'],
    ['Low to Moderate', 'Low to Moderate'],
    ['LOW_MODERATE', 'Low to Moderate'],
    ['Low to Medium', 'Low to Moderate'],
    ['Minor', 'Low to Moderate'],
    ['Guarded', 'Low to Moderate'],
    ['Rendah ke Sedang', 'Low to Moderate'],
    ['KOL-2', 'Low to Moderate'],
    ['Low', 'Low'],
    ['LOW', 'Low'],
    ['Very Low', 'Low'],
    ['Minimal', 'Low'],
    ['Insignificant', 'Low'],
    ['Negligible', 'Low'],
    ['Rendah', 'Low'],
    ['KOL-1', 'Low'],
    ['Unknown External Status', 'Moderate'],
    ['', 'Moderate'],
    [null, 'Moderate'],
    [95, 'High'],
    [70, 'Moderate to High'],
    [50, 'Moderate'],
    [30, 'Low to Moderate'],
    [25, 'High'],
    [16, 'Moderate to High'],
    [10, 'Moderate'],
    [6, 'Low to Moderate'],
    [0, 'Low'],
    [5, 'High'],
    [4, 'Moderate to High'],
    [3, 'Moderate'],
    [2, 'Low to Moderate'],
    [1, 'Low']
  ])('normalizes external/legacy value %s strictly onto %s', (input, expected) => {
    const out = normalizeAnomalyRiskLevel(input)
    expect(out).toBe(expected)
    expect(ANOMALY_RISK_LEVELS).toContain(out)
  })

  it('getRiskConfig never leaks CRITICAL or Critical or unknown external labels', () => {
    const { getRiskConfig } = useAiAnalytics()
    const inputs = ['CRITICAL', 'Critical', 'Extreme', 'Very High', 'Minor', 'Major', 'Insignificant', 'CUSTOM_TIER_X', 'MODERATE_HIGH', 'LOW_MODERATE']
    for (const inp of inputs) {
      const cfg = getRiskConfig(inp)
      expect(ANOMALY_RISK_LEVELS).toContain(cfg.label)
      expect(cfg.label.toLowerCase()).not.toContain('critical')
    }
    expect(getRiskConfig('CRITICAL').label).toBe('High')
    expect(getRiskConfig('Critical').label).toBe('High')
    expect(getRiskConfig('MODERATE_HIGH').label).toBe('Moderate to High')
    expect(getRiskConfig('LOW_MODERATE').label).toBe('Low to Moderate')
  })

  it('normalizes legacy Critical records in custom anomalies and bank seed dataset', () => {
    const allowedSet = new Set(ANOMALY_RISK_LEVELS)

    for (const seed of BANK_ANOMALIES_SEED) {
      expect(seed.severity.toLowerCase()).not.toBe('critical')
      expect(allowedSet.has(normalizeAnomalyRiskLevel(seed.riskLevel))).toBe(true)
      expect(allowedSet.has(seed.severity as any)).toBe(true)
    }

    const { isolationState, addCustomAnomalyRecord } = useAiAnalytics()
    const added = addCustomAnomalyRecord({
      id: 'ANM-EXT-991',
      entity: 'External DB Core',
      type: 'Transaction',
      severity: 'Critical',
      riskLevel: 'EXTREME',
      xMetric: 520
    })

    expect(added.riskLevel).toBe('High')
    expect(added.severity).toBe('High')

    const persisted = getPersistedCustomAnomalies()
    const match = persisted.find(a => a.id === 'ANM-EXT-991')
    expect(match.riskLevel).toBe('High')
    expect(match.severity).toBe('High')

    const stateItem = isolationState.value.anomalies.find((a: any) => a.id === 'ANM-EXT-991')
    expect(stateItem.riskLevel).toBe('High')
  })

  it('ensures static anomaly profiles and Data Source rule severity options exclude Critical', () => {
    const dsContent = readFileSync(resolve(__dirname, '../../components/settings/data-source.vue'), 'utf8')
    const bannerContent = readFileSync(resolve(__dirname, '../../components/common/MappedSourceDataBanner.vue'), 'utf8')
    const pageContent = readFileSync(resolve(__dirname, '../../pages/analytics/ai/anomaly-detection.vue'), 'utf8')

    expect(dsContent).not.toContain("{ label: 'CRITICAL', value: 'CRITICAL' }")
    expect(dsContent).toContain("{ label: 'Moderate to High', value: 'MODERATE_HIGH' }")
    expect(dsContent).toContain("{ label: 'Low to Moderate', value: 'LOW_MODERATE' }")
    expect(bannerContent).not.toMatch(/TABLE_ANOMALY_PROFILES[\s\S]*riskLevel:\s*'CRITICAL'/)
    expect(pageContent).not.toContain("rule.severity === 'CRITICAL'")
  })

  it('converts display levels to expected backend 5-tier risk codes', () => {
    expect(toAnomalyRiskCode('Critical')).toBe('HIGH')
    expect(toAnomalyRiskCode('High')).toBe('HIGH')
    expect(toAnomalyRiskCode('Moderate to High')).toBe('MODERATE_HIGH')
    expect(toAnomalyRiskCode('Moderate')).toBe('MODERATE')
    expect(toAnomalyRiskCode('Low to Moderate')).toBe('LOW_MODERATE')
    expect(toAnomalyRiskCode('Low')).toBe('LOW')
  })
})
