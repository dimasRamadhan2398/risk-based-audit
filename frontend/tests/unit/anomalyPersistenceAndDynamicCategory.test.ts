import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  useAiAnalytics,
  getPersistedCustomAnomalies,
  getPersistedCustomScatter,
  savePersistedCustomAnomaly
} from '~/composables/useAiAnalytics'
import { buildAnomalyScatterPoints } from '~/composables/useAnomalyScatter'

describe('Anomaly Persistence & Unlimited Dynamic Categories', () => {
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

  it('persists a new custom category and anomaly to localStorage and reactive state', () => {
    const { isolationState, addCustomAnomalyRecord } = useAiAnalytics()

    const newRecord = {
      id: 'ANM-CUSTOM-001',
      entity: 'Divisi Digital Banking',
      type: 'Cryptocurrency Settlement',
      anomalyScore: 0.95,
      description: 'Transaksi settlement kripto di luar jam kliring',
      severity: 'Critical',
      date: '2026-11-10',
      xMetric: 850.5,
      riskLevel: 'CRITICAL'
    }

    addCustomAnomalyRecord(newRecord)

    // Check localStorage
    const persisted = getPersistedCustomAnomalies()
    expect(persisted.some(a => a.id === 'ANM-CUSTOM-001')).toBe(true)

    const persistedScatter = getPersistedCustomScatter()
    expect(persistedScatter.some(s => s.label === 'ANM-CUSTOM-001')).toBe(true)

    // Check Vue state
    const inState = isolationState.value.anomalies.find((a: any) => a.id === 'ANM-CUSTOM-001')
    expect(inState).toBeDefined()
    expect(inState.type).toBe('Cryptocurrency Settlement')
    expect(inState.entity).toBe('Divisi Digital Banking')
  })

  it('keeps normalPoints empty ([]) for newly added categories without real normal baseline', () => {
    const { isolationState } = useAiAnalytics()

    const { normalPoints, anomalyPoints } = buildAnomalyScatterPoints(
      isolationState.value.anomalies,
      isolationState.value.scatterData,
      'Cryptocurrency Settlement'
    )

    // Per user rule: Only real normal data is plotted, empty if none exists, no dummy data
    expect(normalPoints).toEqual([])
    expect(anomalyPoints.length).toBeGreaterThanOrEqual(1)
    expect(anomalyPoints[0].x).toBe(850.5)
  })

  it('persists custom category records across browser refreshes (F5 simulation)', () => {
    // 1. Save custom anomaly
    savePersistedCustomAnomaly({
      id: 'ANM-F5-TEST',
      entity: 'Kantor Cabang Jayapura',
      type: 'Papua Special Branch',
      anomalyScore: 0.92,
      xMetric: 420.0,
      riskLevel: 'HIGH'
    }, {
      x: 420.0,
      y: 65,
      type: 'Papua Special Branch',
      isAnomaly: true,
      label: 'ANM-F5-TEST'
    })

    // 2. Simulate refresh by retrieving persisted items
    const restored = getPersistedCustomAnomalies()
    expect(restored.some(a => a.id === 'ANM-F5-TEST')).toBe(true)

    const restoredScatter = getPersistedCustomScatter()
    expect(restoredScatter.some(s => s.label === 'ANM-F5-TEST')).toBe(true)
  })

  it('removes custom anomaly correctly when deleted', () => {
    const { isolationState, addCustomAnomalyRecord, removeCustomAnomalyRecord } = useAiAnalytics()

    addCustomAnomalyRecord({
      id: 'ANM-TO-DELETE',
      entity: 'Unit Kerja Sementara',
      type: 'Temporary Category',
      anomalyScore: 0.75,
      xMetric: 100,
      riskLevel: 'MODERATE'
    })

    expect(isolationState.value.anomalies.some((a: any) => a.id === 'ANM-TO-DELETE')).toBe(true)

    removeCustomAnomalyRecord('ANM-TO-DELETE')

    expect(isolationState.value.anomalies.some((a: any) => a.id === 'ANM-TO-DELETE')).toBe(false)
    const persisted = getPersistedCustomAnomalies()
    expect(persisted.some(a => a.id === 'ANM-TO-DELETE')).toBe(false)
  })

  it('correctly maps and persists normal transaction distribution points from connected table', () => {
    const { isolationState, addCustomScatterBatch } = useAiAnalytics()

    const normalBaselinePoints = [
      { x: 3.99, y: 35.0, type: 'GL Transaction', isAnomaly: false, label: 'GL-NORM-0001' },
      { x: 12.14, y: 56.5, type: 'GL Transaction', isAnomaly: false, label: 'GL-NORM-0006' },
      { x: 25.00, y: 50.0, type: 'GL Transaction', isAnomaly: false, label: 'GL-NORM-0013' }
    ]

    addCustomScatterBatch(normalBaselinePoints)

    const persistedScatter = getPersistedCustomScatter()
    expect(persistedScatter.some(s => s.label === 'GL-NORM-0001')).toBe(true)
    expect(persistedScatter.some(s => s.label === 'GL-NORM-0013')).toBe(true)
    // Canonicalized to 'Transaction' so no duplicate category tab is created
    expect(persistedScatter.find(s => s.label === 'GL-NORM-0001')?.type).toBe('Transaction')

    const { normalPoints } = buildAnomalyScatterPoints(
      isolationState.value.anomalies,
      isolationState.value.scatterData,
      'Transaction'
    )

    expect(normalPoints.length).toBeGreaterThanOrEqual(3)
    expect(normalPoints.some(p => p.x === 3.99 && p.y === 35.0)).toBe(true)
  })

  it('unifies anomalies and normal distribution points for all connected database tables without duplicating categories', async () => {
    const { listAnomalyTypes, getCanonicalCategory } = await import('~/composables/useAnomalyScatter')
    
    // Connected tables from database
    const tables = [
      'gl_transactions',
      'loan_accounts',
      'customer_accounts',
      'user_access_audit_logs',
      'compliance_incident_records',
      'vendor_invoices'
    ]

    const mappedCategories = tables.map(getCanonicalCategory)
    expect(mappedCategories).toEqual([
      'Transaction',
      'Lending',
      'Funding',
      'IT Control',
      'KYC',
      'Procurement'
    ])

    // Verify listAnomalyTypes never produces 'GL Transaction' or duplicate tabs
    const anomalies = [
      { id: 'A1', type: 'gl_transactions' },
      { id: 'A2', type: 'Transaction' },
      { id: 'A3', type: 'GL Transaction' },
      { id: 'A4', type: 'loan_accounts' },
      { id: 'A5', type: 'Lending' }
    ]
    const scatter = [
      { label: 'S1', type: 'gl_transactions', isAnomaly: false },
      { label: 'S2', type: 'Transaction', isAnomaly: false },
      { label: 'S3', type: 'GL Transaction', isAnomaly: false },
      { label: 'S4', type: 'loan_accounts', isAnomaly: false }
    ]

    const tabs = listAnomalyTypes(anomalies, scatter)
    expect(tabs.includes('GL Transaction')).toBe(false)
    expect(tabs.includes('gl_transactions')).toBe(false)
    expect(tabs.includes('loan_accounts')).toBe(false)
    expect(tabs.includes('Transaction')).toBe(true)
    expect(tabs.includes('Lending')).toBe(true)
    // Check no duplicates in tabs
    const uniqueTabs = new Set(tabs)
    expect(uniqueTabs.size).toBe(tabs.length)
  })

  it('dynamically generates normal distribution and anomaly points for any brand new database table', async () => {
    const { getCanonicalCategory, getAnomalyTypeConfig, buildAnomalyScatterPoints, listAnomalyTypes } = await import('~/composables/useAnomalyScatter')
    const { isolationState, addCustomAnomalyRecord, addCustomScatterBatch } = useAiAnalytics()

    // Completely new table never seen before
    const newTableName = 'digital_asset_custody_logs'
    const canonicalCat = getCanonicalCategory(newTableName)
    expect(canonicalCat).toBe('Digital Asset Custody')

    // Dynamic config generation
    const config = getAnomalyTypeConfig((k: string) => k, canonicalCat)
    expect(config.xAxisTitle).toContain('Digital Asset Custody')
    expect(config.colors).toBeDefined()
    expect(config.colors.style).toBeDefined()

    // Add anomaly
    const newAnomaly = {
      id: 'ANOM-DAC-001',
      entity: 'Sentra Aset Digital',
      refNumber: 'DAC-9901',
      accountNumber: 'ACC-CRYPTO-01',
      amount: 450000000,
      type: canonicalCat,
      category: canonicalCat,
      anomalyScore: 0.95,
      description: 'Transfer custody wallet di luar jam operasional',
      riskLevel: 'CRITICAL',
      xMetric: 450.0,
      date: '2026-11-10'
    }
    const anomalyScatter = {
      x: 450.0,
      y: 55.0,
      type: canonicalCat,
      isAnomaly: true,
      label: 'ANOM-DAC-001'
    }
    addCustomAnomalyRecord(newAnomaly, anomalyScatter)

    // Add dynamic normal distribution points
    const dynamicNormalPoints = Array.from({ length: 15 }, (_, i) => ({
      x: 20.0 + i * 5,
      y: 30.0 + (i % 5) * 4,
      type: canonicalCat,
      isAnomaly: false,
      label: `DAC-NORM-${i + 1}`
    }))
    addCustomScatterBatch(dynamicNormalPoints)

    // Verify presence in listAnomalyTypes
    const availableTypes = listAnomalyTypes(isolationState.value.anomalies, isolationState.value.scatterData)
    expect(availableTypes.includes('Digital Asset Custody')).toBe(true)

    // Verify scatter points build both normal & anomaly under the SAME category
    const { normalPoints, anomalyPoints } = buildAnomalyScatterPoints(
      isolationState.value.anomalies,
      isolationState.value.scatterData,
      'Digital Asset Custody'
    )
    expect(normalPoints.length).toBe(15)
    expect(anomalyPoints.length).toBeGreaterThanOrEqual(1)
    expect(anomalyPoints.some(p => p.x === 450.0 && p.y === 55.0)).toBe(true)
  })

  it('scales anomalies and normal distribution points additively without limits across multiple batches', () => {
    const { isolationState, addCustomAnomalyRecord, addCustomScatterBatch } = useAiAnalytics()

    const initialAnomalyCount = (isolationState.value.anomalies || []).length
    const initialScatterCount = (isolationState.value.scatterData || []).length

    // Simulate Batch 1: Ingesting 6 anomalies and 60 normal points from gl_transactions
    const batch1Anomalies = Array.from({ length: 6 }, (_, i) => ({
      id: `ANOM-B1-${i + 1}`,
      entity: `Branch ${i + 1}`,
      type: 'Transaction',
      xMetric: 100 + i * 50,
      riskLevel: 'HIGH',
      description: `Anomaly batch 1 record ${i + 1} gl_transactions`
    }))
    batch1Anomalies.forEach((a, i) => {
      addCustomAnomalyRecord(a, {
        x: a.xMetric,
        y: 40 + i * 3,
        type: 'Transaction',
        isAnomaly: true,
        label: a.id
      })
    })

    const batch1Normal = Array.from({ length: 60 }, (_, i) => ({
      x: 5.0 + i * 1.5,
      y: 25.0 + (i % 7) * 4,
      type: 'Transaction',
      isAnomaly: false,
      label: `GL-NORM-B1-${i + 1}`
    }))
    addCustomScatterBatch(batch1Normal)

    expect(isolationState.value.anomalies.length).toBe(initialAnomalyCount + 6)
    expect(isolationState.value.scatterData.length).toBe(initialScatterCount + 66) // 6 anomalies + 60 normal

    // Simulate Batch 2: Ingesting additional batch (additive, no blocking limit)
    const batch2Anomalies = Array.from({ length: 6 }, (_, i) => ({
      id: `ANOM-B2-${i + 1}`,
      entity: `Branch B2-${i + 1}`,
      type: 'Transaction',
      xMetric: 200 + i * 60,
      riskLevel: 'CRITICAL',
      description: `Anomaly batch 2 record ${i + 1} gl_transactions`
    }))
    batch2Anomalies.forEach((a, i) => {
      addCustomAnomalyRecord(a, {
        x: a.xMetric,
        y: 45 + i * 3,
        type: 'Transaction',
        isAnomaly: true,
        label: a.id
      })
    })

    const batch2Normal = Array.from({ length: 60 }, (_, i) => ({
      x: 8.0 + i * 1.5,
      y: 28.0 + (i % 7) * 4,
      type: 'Transaction',
      isAnomaly: false,
      label: `GL-NORM-B2-${i + 1}`
    }))
    addCustomScatterBatch(batch2Normal)

    // Verify unlimited cumulative growth (6 + 6 = 12 anomalies, 60 + 60 = 120 normal points added)
    expect(isolationState.value.anomalies.length).toBe(initialAnomalyCount + 12)
    expect(isolationState.value.scatterData.length).toBe(initialScatterCount + 132)

    const { normalPoints, anomalyPoints } = buildAnomalyScatterPoints(
      isolationState.value.anomalies,
      isolationState.value.scatterData,
      'Transaction'
    )
    expect(normalPoints.length).toBeGreaterThanOrEqual(120)
    expect(anomalyPoints.length).toBeGreaterThanOrEqual(12)
  })
})

