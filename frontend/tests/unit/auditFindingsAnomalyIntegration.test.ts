import { describe, expect, it } from 'vitest'
import {
  buildAnomalyScatterPoints,
  createAnomalyTypeConfigs,
  DEFAULT_ANOMALY_TYPES,
  getCanonicalCategory,
  listAnomalyTypes
} from '~/composables/useAnomalyScatter'
import {
  ANOMALY_RISK_LEVELS,
  normalizeAnomalyRiskLevel,
  toAnomalyRiskCode
} from '~/utils/riskLevelBadge'

describe('Audit Findings (Action Taken Report) Anomaly Detection Integration', () => {
  it('canonicalizes all finding and ATR aliases to Audit Findings', () => {
    expect(getCanonicalCategory('Audit Findings')).toBe('Audit Findings')
    expect(getCanonicalCategory('audit_findings')).toBe('Audit Findings')
    expect(getCanonicalCategory('finding')).toBe('Audit Findings')
    expect(getCanonicalCategory('findings')).toBe('Audit Findings')
    expect(getCanonicalCategory('temuan')).toBe('Audit Findings')
    expect(getCanonicalCategory('temuan_audit')).toBe('Audit Findings')
    expect(getCanonicalCategory('atr')).toBe('Audit Findings')
    expect(getCanonicalCategory('action_taken_report')).toBe('Audit Findings')
    expect(getCanonicalCategory('action_taken_reports')).toBe('Audit Findings')
  })

  it('includes Audit Findings in DEFAULT_ANOMALY_TYPES and createAnomalyTypeConfigs', () => {
    expect(DEFAULT_ANOMALY_TYPES).toContain('Audit Findings')

    const tMock = (k: string) => k
    const idConfigs = createAnomalyTypeConfigs(tMock, 'id')
    expect(idConfigs['Audit Findings']).toBeDefined()
    expect(idConfigs['Audit Findings']?.unit).toBe('Hari')
    expect(idConfigs['Audit Findings']?.formatX(27)).toBe('27 Hari')

    const enConfigs = createAnomalyTypeConfigs(tMock, 'en')
    expect(enConfigs['Audit Findings']).toBeDefined()
    expect(enConfigs['Audit Findings']?.unit).toBe('Days')
    expect(enConfigs['Audit Findings']?.formatX(27)).toBe('27 Days')
  })

  it('builds normal baseline (blue) and anomalous (red) scatter points for Audit Findings with 5-tier Risk Levels', () => {
    const atrAnomalies = [
      {
        id: 'ATR-38D28B58',
        entity: 'IT Infrastructure',
        type: 'Audit Findings',
        finding_significance: 'Very Significant',
        significance_score: 4,
        report_number: 'LHA-003/SKAI/2026',
        status: 'PLANNED',
        progress: 0,
        overdue_days: 27,
        xMetric: 27,
        hour: 4.4,
        risk_level: 'HIGH',
        severity: 'High'
      },
      {
        id: 'ATR-6238E163',
        entity: 'Supply Chain & Gudang',
        type: 'Audit Findings',
        finding_significance: 'Significant',
        significance_score: 3,
        report_number: 'LHA-002/SKAI/2026',
        status: 'IN_PROGRESS',
        progress: 15,
        overdue_days: 19,
        xMetric: 19,
        hour: 6.2,
        risk_level: 'MODERATE_HIGH',
        severity: 'Moderate to High'
      }
    ]

    const atrScatter = [
      { x: 5, y: 18.2, type: 'Audit Findings', isAnomaly: false, label: 'ATR-A948E8F6' },
      { x: 8, y: 19.0, type: 'Audit Findings', isAnomaly: false, label: 'ATR-EEA6BE8A' },
      { x: 27, y: 4.4, type: 'Audit Findings', isAnomaly: true, label: 'ATR-38D28B58' },
      { x: 19, y: 6.2, type: 'Audit Findings', isAnomaly: true, label: 'ATR-6238E163' }
    ]

    const tabs = listAnomalyTypes(atrAnomalies, atrScatter)
    expect(tabs).toContain('Audit Findings')

    const series = buildAnomalyScatterPoints(atrAnomalies, atrScatter, 'Audit Findings')
    expect(series.normalPoints).toHaveLength(2)
    expect(series.anomalyPoints).toHaveLength(2)
    expect(series.anomalyPoints[0]).toEqual({ x: 27, y: 4.4 })
    expect(series.anomalyPoints[1]).toEqual({ x: 19, y: 6.2 })

    for (const item of atrAnomalies) {
      const normalized = normalizeAnomalyRiskLevel(item.risk_level || item.severity)
      expect(ANOMALY_RISK_LEVELS).toContain(normalized)
      expect(normalized).not.toBe('Critical')
      expect(toAnomalyRiskCode(normalized)).not.toBe('CRITICAL')
    }
  })
})
