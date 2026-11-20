// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'
import {
  NLP_RISK_LEVELS,
  normalizeNlpRiskLevel,
  toNlpRiskCode
} from '~/utils/riskLevelBadge'
import {
  useAiAnalytics,
  sanitizeNlpDocumentRecord,
  getPersistedCustomNlpDocs
} from '~/composables/useAiAnalytics'

const CANONICAL_CODES = ['HIGH', 'MODERATE_HIGH', 'MODERATE', 'LOW_MODERATE', 'LOW']
const CANONICAL_LABELS = ['High', 'Moderate to High', 'Moderate', 'Low to Moderate', 'Low']

describe('NLP Analysis 5-Tier Risk Level Enforcement (No Critical / Unmapped External Tiers)', () => {
  let localStorageMock: Record<string, string> = {}

  beforeEach(() => {
    setActivePinia(createPinia())
    localStorageMock = {}
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => localStorageMock[key] ?? null,
      setItem: (key: string, val: string) => { localStorageMock[key] = String(val) },
      removeItem: (key: string) => { delete localStorageMock[key] },
      clear: () => { localStorageMock = {} }
    })
  })

  it('defines strictly the 5 canonical risk levels for NLP Analysis', () => {
    expect(NLP_RISK_LEVELS).toEqual([
      'High',
      'Moderate to High',
      'Moderate',
      'Low to Moderate',
      'Low'
    ])
    expect(NLP_RISK_LEVELS).not.toContain('Critical')
    expect(NLP_RISK_LEVELS).not.toContain('Medium')
  })

  it.each([
    ['Critical', 'High', 'HIGH'],
    ['CRITICAL', 'High', 'HIGH'],
    ['Extreme', 'High', 'HIGH'],
    ['Very High', 'High', 'HIGH'],
    ['Catastrophic', 'High', 'HIGH'],
    ['Sangat Tinggi', 'High', 'HIGH'],
    ['Kritis', 'High', 'HIGH'],
    ['Major', 'Moderate to High', 'MODERATE_HIGH'],
    ['Elevated', 'Moderate to High', 'MODERATE_HIGH'],
    ['Significant', 'Moderate to High', 'MODERATE_HIGH'],
    ['Medium to High', 'Moderate to High', 'MODERATE_HIGH'],
    ['KOL-3', 'Moderate to High', 'MODERATE_HIGH'],
    ['Medium', 'Moderate', 'MODERATE'],
    ['MEDIUM', 'Moderate', 'MODERATE'],
    ['Watch', 'Moderate', 'MODERATE'],
    ['Not Scored', 'Moderate', 'MODERATE'],
    ['Minor', 'Low to Moderate', 'LOW_MODERATE'],
    ['Guarded', 'Low to Moderate', 'LOW_MODERATE'],
    ['Low to Medium', 'Low to Moderate', 'LOW_MODERATE'],
    ['KOL-2', 'Low to Moderate', 'LOW_MODERATE'],
    ['Insignificant', 'Low', 'LOW'],
    ['Minimal', 'Low', 'LOW'],
    ['Negligible', 'Low', 'LOW'],
    ['Safe', 'Low', 'LOW'],
    ['KOL-1', 'Low', 'LOW']
  ])('normalizes external/legacy NLP risk level %s -> %s (%s)', (raw, expectedLabel, expectedCode) => {
    expect(normalizeNlpRiskLevel(raw)).toBe(expectedLabel)
    expect(toNlpRiskCode(raw)).toBe(expectedCode)
    expect(CANONICAL_LABELS).toContain(normalizeNlpRiskLevel(raw))
    expect(CANONICAL_CODES).toContain(toNlpRiskCode(raw))
  })

  it('heals poisoned legacy CRITICAL / Extreme / Medium / Minor records from localStorage', () => {
    localStorage.setItem(
      'auditsphere_custom_nlp_docs_v3',
      JSON.stringify([
        { docId: 'EXT-POISON-1', title: 'External Critical Doc', riskLevel: 'CRITICAL' },
        { docId: 'EXT-POISON-2', title: 'External Major Doc', riskLevel: 'Major' },
        { docId: 'EXT-POISON-3', title: 'External Medium Doc', riskLevel: 'MEDIUM' },
        { docId: 'EXT-POISON-4', title: 'External Minor Doc', riskLevel: 'Minor' },
        { docId: 'EXT-POISON-5', title: 'External Insignificant Doc', riskLevel: 'Insignificant' }
      ])
    )

    const loaded = getPersistedCustomNlpDocs()
    expect(loaded).toHaveLength(5)
    expect(loaded.map(d => d.riskLevel)).toEqual([
      'HIGH',
      'MODERATE_HIGH',
      'MODERATE',
      'LOW_MODERATE',
      'LOW'
    ])
    expect(loaded.map(d => d.riskLevelLabel)).toEqual([
      'High',
      'Moderate to High',
      'Moderate',
      'Low to Moderate',
      'Low'
    ])
  })

  it('analyzeDocumentOrText never emits CRITICAL on Fraud documents and adapts external DB ratings', async () => {
    // Offline fallback mode ($fetch returns null) with fraud text
    vi.stubGlobal('$fetch', vi.fn().mockResolvedValue(null))
    const { analyzeDocumentOrText, getRiskConfig } = useAiAnalytics()

    const fraudDoc = await analyzeDocumentOrText({
      title: 'Indikasi Fraud Transaksi Fiktif',
      text: 'Ditemukan transaksi fiktif dan indikasi fraud pada pencairan vendor.',
      source: 'External DB',
      sourceType: 'External Database'
    })
    expect(fraudDoc.riskLevel).toBe('HIGH')
    expect(fraudDoc.riskLevel).not.toBe('CRITICAL')
    expect(getRiskConfig(fraudDoc.riskLevel).label).toBe('High')

    // Backend returning non-5-tier value from external system gets normalized immediately
    vi.stubGlobal('$fetch', vi.fn().mockResolvedValue({
      docId: 'EXT-901',
      title: 'Laporan SIEM Eksternal',
      risk_category: 'Technology',
      sentiment: 'Negative',
      risk_level: 'Extreme',
      extracted_finding: 'Percobaan intrusi firewall eksternal.',
      extracted_recommendation: 'Blokir IP sumber dan perketat aturan firewall.'
    }))
    const extExtremeDoc = await analyzeDocumentOrText({
      title: 'Laporan SIEM Eksternal',
      text: 'Percobaan intrusi firewall eksternal.',
      riskLevel: 'Extreme'
    })
    expect(extExtremeDoc.riskLevel).toBe('HIGH')
    expect(getRiskConfig(extExtremeDoc.riskLevel).label).toBe('High')

    const sanitizedMinor = sanitizeNlpDocumentRecord({ docId: 'D-2', risk_level: 'Minor' })
    expect(sanitizedMinor.riskLevel).toBe('LOW_MODERATE')
    expect(sanitizedMinor.riskLevelLabel).toBe('Low to Moderate')
  })

  it('syncNlpMultiSources normalizes non-5-tier external DB values strictly onto the 5 canonical levels', async () => {
    vi.stubGlobal('$fetch', vi.fn().mockImplementation((url: string) => {
      if (String(url).includes('sync-sources')) {
        return Promise.resolve({
          documents: [
            { docId: 'EXT-S1', title: 'External Critical', risk_level: 'CRITICAL', excerpt: 'Temuan kritis.' },
            { docId: 'EXT-S2', title: 'External Major', risk_level: 'Major', excerpt: 'Temuan signifikan.' },
            { docId: 'EXT-S3', title: 'External Medium', risk_level: 'MEDIUM', excerpt: 'Temuan sedang.' },
            { docId: 'EXT-S4', title: 'External Minor', risk_level: 'Minor', excerpt: 'Temuan minor.' },
            { docId: 'EXT-S5', title: 'External Safe', risk_level: 'Insignificant', excerpt: 'Kondisi aman.' }
          ]
        })
      }
      return Promise.resolve(null)
    }))

    const { syncNlpMultiSources, nlpState, getRiskConfig } = useAiAnalytics()
    await syncNlpMultiSources()

    const docs = nlpState.value.documents
    expect(docs.map((d: any) => d.riskLevel)).toEqual([
      'HIGH',
      'MODERATE_HIGH',
      'MODERATE',
      'LOW_MODERATE',
      'LOW'
    ])
    for (const d of docs) {
      expect(CANONICAL_CODES).toContain(d.riskLevel)
      expect(CANONICAL_LABELS).toContain(getRiskConfig(d.riskLevel).label)
    }
  })

  it('ensures no CRITICAL fallback remains in NLP composables/Go service and nlp.vue exposes 5-tier filter', () => {
    const useAiSrc = readFileSync(resolve(__dirname, '../../composables/useAiAnalytics.ts'), 'utf8')
    const nlpPageSrc = readFileSync(resolve(__dirname, '../../pages/analytics/ai/nlp.vue'), 'utf8')
    const bannerSrc = readFileSync(resolve(__dirname, '../../components/common/MappedSourceDataBanner.vue'), 'utf8')
    const goServiceSrc = readFileSync(resolve(__dirname, '../../../backend/analytics-service/services/analytics_service.go'), 'utf8')

    expect(useAiSrc).not.toContain("isFraud ? 'CRITICAL'")
    expect(useAiSrc).toContain('sanitizeNlpDocumentRecord')
    expect(nlpPageSrc).toContain('NLP_RISK_LEVELS')
    expect(nlpPageSrc).toContain('selectedRiskLevel')
    expect(bannerSrc).toContain('toNlpRiskCode')
    expect(goServiceSrc).not.toContain('riskLevel = "CRITICAL"')
    expect(goServiceSrc).toContain('normalizeCanonicalRiskLevelGo')
  })
})
