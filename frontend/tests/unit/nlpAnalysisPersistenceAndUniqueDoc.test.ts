import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  useAiAnalytics,
  getPersistedCustomNlpDocs,
  savePersistedCustomNlpDoc,
  getPersistedCustomNlpTaxonomy,
  savePersistedCustomNlpTaxonomy,
  mergeNlpDocuments,
  mergeNlpTaxonomy
} from '~/composables/useAiAnalytics'

describe('NLP Analysis Fixes: Persistence, Unique DocIDs & Fallback Guard', () => {
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

  it('Issue 1: rejects stale cash fallback finding and maps technical document correctly', async () => {
    const { analyzeDocumentOrText } = useAiAnalytics()

    // Simulate backend returning stale fallback finding with cash
    vi.stubGlobal('$fetch', vi.fn().mockResolvedValue({
      docId: 'WP-2026-99',
      title: 'Ditemukan indikasi kelemahan pengendalian internal pada otorisasi kas.',
      risk_category: 'Financial',
      extracted_finding: 'Ditemukan indikasi kelemahan pengendalian internal pada otorisasi kas.',
      extracted_recommendation: 'Perkuat verifikasi dokumen pendukung dan terapkan otorisasi berjenjang.'
    }))

    const result = await analyzeDocumentOrText({
      title: 'Audit Keamanan Jaringan dan Firewall IT',
      fileName: 'Laporan_Keamanan_Jaringan_Server.pdf',
      text: 'Pemeriksaan konfigurasi firewall dan hak akses server database produksi.'
    })

    // Title should NOT be the cash finding
    expect(result.title).toBe('Audit Keamanan Jaringan dan Firewall IT')
    // Category should be Technology, not Financial
    expect(result.autoCategory).toBe('Technology')
    // Extracted finding should NOT contain otorisasi kas
    expect(result.extractedFinding).not.toContain('otorisasi kas')
    // Recommendation should NOT be the cash authorization recommendation
    expect(result.extractedRecommendation).not.toContain('otorisasi berjenjang')
    // DocID should NOT be the hardcoded WP-2026-99
    expect(result.docId).not.toBe('WP-2026-99')
  })

  it('Issue 2: persists uploaded documents and custom taxonomy across refresh', async () => {
    const { nlpState, addNlpTaxonomyItem } = useAiAnalytics()

    // 1. Add custom taxonomy
    await addNlpTaxonomyItem({
      itemType: 'risk_category',
      name: 'Cybersecurity & AI Governance',
      keywords: ['cyber', 'ai', 'model drift', 'prompt injection'],
      description: 'Risiko keamanan kecerdasan buatan'
    })

    const persistedTax = getPersistedCustomNlpTaxonomy()
    expect(persistedTax.risk_categories.some(c => (c.name || c) === 'Cybersecurity & AI Governance')).toBe(true)

    // 2. Add custom document directly to persistence
    const testDoc = {
      docId: 'DOC-CUSTOM-TEST-001',
      title: 'Laporan Audit Infrastruktur Cloud',
      fileName: 'Cloud_Audit_2026.pdf',
      autoCategory: 'Technology',
      controlIndicator: 'IT & Access Control Weakness'
    }
    savePersistedCustomNlpDoc(testDoc)

    const persistedDocs = getPersistedCustomNlpDocs()
    expect(persistedDocs.some(d => d.docId === 'DOC-CUSTOM-TEST-001')).toBe(true)

    // 3. Simulate simulated page refresh merge
    const fetchedDefaultDocs = [
      { docId: 'WP-001', title: 'Default Working Paper' },
      { docId: 'WP-002', title: 'Default Audit Report' }
    ]
    const mergedDocs = mergeNlpDocuments(fetchedDefaultDocs, persistedDocs)
    expect(mergedDocs.some(d => d.docId === 'DOC-CUSTOM-TEST-001')).toBe(true)
    expect(mergedDocs[0].docId).toBe('DOC-CUSTOM-TEST-001')
  })

  it('Issue 3: uploading multiple documents appends rows without overwriting previous entries', async () => {
    const { nlpState, analyzeDocumentOrText } = useAiAnalytics()

    const initialCount = nlpState.value.documents.length

    // Mock backend returning successful dynamic responses
    let counter = 1
    vi.stubGlobal('$fetch', vi.fn().mockImplementation(() => {
      counter++
      return Promise.resolve({
        docId: `DYNAMIC-DOC-${counter}`,
        title: `Dokumen Pengujian Ke-${counter}`,
        risk_category: 'Technology',
        extracted_finding: `Temuan pengujian nomor ${counter}`,
        extracted_recommendation: `Rekomendasi tindak lanjut nomor ${counter}`
      })
    }))

    // Upload doc 1
    const doc1 = await analyzeDocumentOrText({
      title: 'Pemeriksaan Server Cabang Bandung',
      fileName: 'Server_Bandung.pdf',
      text: 'Temuan akun dormant pada sistem cabang.'
    })

    // Upload doc 2
    const doc2 = await analyzeDocumentOrText({
      title: 'Pemeriksaan Database Pusat',
      fileName: 'Database_Pusat.pdf',
      text: 'Temuan hak akses berlebih pada database transaksi.'
    })

    // Upload doc 3
    const doc3 = await analyzeDocumentOrText({
      title: 'Pemeriksaan Keamanan API Gateway',
      fileName: 'API_Gateway_Log.json',
      text: 'Log anomali request pada endpoint gateway.'
    })

    // Verify row count increased by 3
    expect(nlpState.value.documents.length).toBe(initialCount + 3)

    // Verify all 3 uploaded documents exist simultaneously
    expect(nlpState.value.documents.some((d: any) => d.docId === doc1.docId)).toBe(true)
    expect(nlpState.value.documents.some((d: any) => d.docId === doc2.docId)).toBe(true)
    expect(nlpState.value.documents.some((d: any) => d.docId === doc3.docId)).toBe(true)

    // Verify doc2 did NOT overwrite doc1
    expect(doc1.docId).not.toBe(doc2.docId)
    expect(doc2.docId).not.toBe(doc3.docId)
  })

  it('Issue 4: accurately classifies "Spesifikasi Resource Auditsphere" as Technology and never outputs status text as finding or excerpt', async () => {
    const { analyzeDocumentOrText } = useAiAnalytics()

    // Simulate backend returning empty or status string
    vi.stubGlobal('$fetch', vi.fn().mockResolvedValue({
      docId: 'DOC-TEST-TECH-001',
      title: 'Spesifikasi Resource Auditsphere',
      risk_category: 'Technology',
      autoCategory: 'Technology',
      extracted_finding: 'Observasi Teknis & Spesifikasi: Kapasitas Server 16 Core vCPU, 64 GB RAM, 1 TB NVMe Storage.',
      extracted_recommendation: 'Lakukan capacity planning dan uji beban (stress test) berkala, pastikan alokasi resource CPU, RAM, & storage sesuai beban kerja.',
      excerpt: 'Spesifikasi Resource Auditsphere. Kapasitas Server 16 Core vCPU, 64 GB RAM, 1 TB NVMe Storage. Utilisasi CPU rata-rata mencapai 88% pada jam sibuk.'
    }))

    const result = await analyzeDocumentOrText({
      title: 'Spesifikasi Resource Auditsphere',
      fileName: 'Spesifikasi_Resource_Auditsphere.pdf',
      text: 'Spesifikasi Resource Auditsphere. Kapasitas Server 16 Core vCPU, 64 GB RAM, 1 TB NVMe Storage. Utilisasi CPU rata-rata mencapai 88% pada jam sibuk.'
    })

    // Must be classified as Technology
    expect(result.autoCategory).toBe('Technology')
    // Finding must NOT be the fake status string
    expect(result.extractedFinding).not.toContain('berhasil diekstrak dan dianalisis')
    expect(result.extractedFinding).toContain('Kapasitas Server 16 Core vCPU')
    // Excerpt must be original extracted text
    expect(result.excerpt).toContain('Spesifikasi Resource Auditsphere')
    expect(result.excerpt).not.toContain('berhasil diekstrak dan dianalisis')
    // Recommendation must be technology/capacity planning, NOT cash/financial
    expect(result.extractedRecommendation).toContain('capacity planning')
    expect(result.extractedRecommendation).not.toContain('otorisasi kas')
  })
})

