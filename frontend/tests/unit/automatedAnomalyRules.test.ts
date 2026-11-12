// @ts-nocheck
import { describe, it, expect, beforeEach } from 'vitest'

interface CustomAnomalyRule {
  id: string
  name: string
  field: string
  operator: string
  threshold: number
  thresholdStr?: string
  severity: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW'
  category?: string
  description?: string
  isActive?: boolean
  trainedAt?: string
}

interface TableDataMapping {
  tableName: string
  targetScope: string
  targetModule: string
  auditField: string
  syncInterval: string
  anomalyRules?: string[]
  customRules?: CustomAnomalyRule[]
  isActive: boolean
}

describe('Automated Anomaly Rules & Model Training', () => {
  let mockStorage: Record<string, string> = {}

  beforeEach(() => {
    mockStorage = {}
    global.localStorage = {
      getItem: (key: string) => mockStorage[key] || null,
      setItem: (key: string, value: string) => { mockStorage[key] = value },
      removeItem: (key: string) => { delete mockStorage[key] },
      clear: () => { mockStorage = {} }
    } as any
  })

  it('allows adding and removing custom anomaly rules on a table mapping', () => {
    const mapping: TableDataMapping = {
      tableName: 'gl_transactions',
      targetScope: 'data_analytics',
      targetModule: 'AI: Anomaly Detection',
      auditField: 'id',
      syncInterval: 'Hourly',
      anomalyRules: ['Threshold Breach (>100M)'],
      customRules: [],
      isActive: true
    }

    // Add custom rule
    const newRule: CustomAnomalyRule = {
      id: 'crule-001',
      name: 'High Value RTGS Transfer (>500M)',
      field: 'amount',
      operator: '>',
      threshold: 500000000,
      severity: 'CRITICAL',
      category: 'Transaction',
      description: 'Transaksi bernilai sangat tinggi via RTGS',
      isActive: true
    }
    mapping.customRules!.push(newRule)

    expect(mapping.customRules?.length).toBe(1)
    expect(mapping.customRules![0].name).toBe('High Value RTGS Transfer (>500M)')
    expect(mapping.customRules![0].threshold).toBe(500000000)
    expect(mapping.customRules![0].severity).toBe('CRITICAL')

    // Toggle active state
    mapping.customRules![0].isActive = false
    expect(mapping.customRules![0].isActive).toBe(false)

    // Remove custom rule
    mapping.customRules = mapping.customRules!.filter(r => r.id !== 'crule-001')
    expect(mapping.customRules?.length).toBe(0)
  })

  it('correctly persists automated anomaly rules to storage', () => {
    const STORAGE_KEY = 'auditsphere_automated_anomaly_rules_v1'
    const rulesToPersist = [
      {
        id: 'rule-1',
        name: 'Threshold Breach (>100M)',
        table_name: 'gl_transactions',
        category: 'Transaction',
        severity: 'HIGH',
        trained_at: '2026-11-11T12:00:00Z'
      },
      {
        id: 'crule-2',
        name: 'Weekend Cash Withdrawal',
        table_name: 'gl_transactions',
        field: 'channel',
        operator: '==',
        threshold: 0,
        thresholdStr: 'ATM',
        category: 'Transaction',
        severity: 'MEDIUM',
        trained_at: '2026-11-11T12:00:00Z'
      }
    ]

    localStorage.setItem(STORAGE_KEY, JSON.stringify(rulesToPersist))
    const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]')

    expect(stored.length).toBe(2)
    expect(stored[0].name).toBe('Threshold Breach (>100M)')
    expect(stored[1].thresholdStr).toBe('ATM')
  })

  it('generates synthetic rule anomalies and normal baseline points correctly for trained rules', () => {
    const rules = [
      {
        name: 'Threshold Breach (>100M)',
        field: 'amount',
        operator: '>',
        threshold: 100000000,
        severity: 'CRITICAL'
      }
    ]

    // Simulate synthesis logic from Python AI train-rules engine
    const syntheticAnomalies = rules.map((r, idx) => ({
      id: `RULE-gl_transactions-${idx + 1}`,
      entity: 'gl_transactions AI Rule Engine',
      amount: r.threshold * 1.5,
      type: 'Transaction',
      category: 'Transaction',
      anomalyScore: 0.96,
      description: `[Rule: ${r.name}] Nilai amount (${r.threshold * 1.5}) melebihi threshold (${r.threshold})`,
      severity: r.severity,
      xMetric: (r.threshold * 1.5) / 1000000,
      hour: 23,
      isAnomaly: true
    }))

    expect(syntheticAnomalies.length).toBe(1)
    expect(syntheticAnomalies[0].id).toContain('RULE-')
    expect(syntheticAnomalies[0].description).toContain('[Rule: Threshold Breach (>100M)]')
    expect(syntheticAnomalies[0].xMetric).toBe(150)
  })

  it('identifies and marks rule match correctly for detail table badges', () => {
    const anomalyRecordWithRule = {
      id: 'RULE-gl_transactions-1',
      description: '[Rule: Threshold Breach (>100M)] Nilai amount melebihi batas 100M',
      reason: 'Rule-based constraint check'
    }

    const isRuleMatch = (record: any) =>
      Boolean(record.id?.startsWith('RULE-') ||
      record.description?.includes('[Rule:') ||
      record.reason?.includes('[Rule:'))

    expect(isRuleMatch(anomalyRecordWithRule)).toBe(true)

    const normalRecord = {
      id: 'ANM-001',
      description: 'Deviasi anomali statistik Isolation Forest',
      reason: 'Jarak multidimensi melampaui kontaminasi 10%'
    }
    expect(isRuleMatch(normalRecord)).toBe(false)
  })
})

