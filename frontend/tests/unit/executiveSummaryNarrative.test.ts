import { describe, expect, it, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import fs from 'node:fs'
import path from 'node:path'
import en from '../../locales/en/common.json'
import id from '../../locales/id/common.json'
import { useExecutiveSummaryStore } from '../../stores/executive-summary'
import {
  auditPeriodFromAssignmentLetter,
  buildDefaultNarrative,
  buildSyncedNarrative,
  deriveNarrativeSyncData,
  formatAuditPeriod,
  formatQuarterPeriod,
  isAutoSyncableNarrative,
  parseExecutionPeriod,
  parseMonthYear,
  prefixNarrative,
  refreshNarrativePeriod,
  type Translate
} from '../../utils/executiveSummaryNarrative'
import {
  buildExecutiveSummaryDocumentXml,
  canGenerateExecutiveSummaryDocx,
  executiveSummaryDocxFilename,
  generateExecutiveSummaryDocxBytes
} from '../../utils/executiveSummaryDocx'

const read = (relativePath: string) => fs.readFileSync(path.resolve(__dirname, '../../', relativePath), 'utf8')

// Same lookup and {param} replacement as composables/useI18n.ts, for a fixed locale
const translator = (messages: unknown): Translate => (key, params) => {
  let value: unknown = messages
  for (const k of key.split('.')) {
    if (value && typeof value === 'object' && k in value) value = (value as Record<string, unknown>)[k]
    else return key
  }
  if (typeof value !== 'string') return key
  return params ? value.replace(/\{(\w+)\}/g, (m, p) => params[p]?.toString() || m) : value
}

const tId = translator(id)
const tEn = translator(en)

const range = (start: string, end: string) => ({ start: parseMonthYear(start), end: parseMonthYear(end) })

describe('Executive summary narrative period', () => {
  describe('date parsing', () => {
    it('reads month and year from ISO dates, datetimes and DD/MM/YYYY', () => {
      expect(parseMonthYear('2026-01-15')).toEqual({ year: 2026, month: 1 })
      expect(parseMonthYear('2026-03')).toEqual({ year: 2026, month: 3 })
      expect(parseMonthYear('2025-12-01T00:00:00Z')).toEqual({ year: 2025, month: 12 })
      expect(parseMonthYear('20/02/2026')).toEqual({ year: 2026, month: 2 })
      expect(parseMonthYear('')).toBeNull()
      expect(parseMonthYear(undefined)).toBeNull()
      expect(parseMonthYear('Kuartal I')).toBeNull()
      expect(parseMonthYear('2026-13-01')).toBeNull()
    })

    it('splits an assignment letter executionPeriod', () => {
      expect(parseExecutionPeriod('2026-01-05 to 2026-03-20')).toEqual({
        start: { year: 2026, month: 1 },
        end: { year: 2026, month: 3 }
      })
      expect(parseExecutionPeriod('2026-01-05 s/d 2026-02-20').end).toEqual({ year: 2026, month: 2 })
      expect(parseExecutionPeriod('Q1 2026')).toEqual({ start: null, end: null })
    })

    it('prefers startPeriod/finishPeriod and falls back to executionPeriod', () => {
      expect(auditPeriodFromAssignmentLetter({ startPeriod: '2026-01-05', finishPeriod: '2026-02-10', executionPeriod: '2025-01-01 to 2025-01-31' }))
        .toEqual({ start: { year: 2026, month: 1 }, end: { year: 2026, month: 2 } })
      expect(auditPeriodFromAssignmentLetter({ executionPeriod: '2025-12-01 to 2026-01-31' }))
        .toEqual({ start: { year: 2025, month: 12 }, end: { year: 2026, month: 1 } })
      expect(auditPeriodFromAssignmentLetter(null)).toEqual({ start: null, end: null })
    })
  })

  describe('formatAuditPeriod', () => {
    it('shows a single month when the audit starts and ends in the same month', () => {
      expect(formatAuditPeriod(tId, range('2026-01-05', '2026-01-28'))).toBe('Januari 2026')
      expect(formatAuditPeriod(tEn, range('2026-01-05', '2026-01-28'))).toBe('January 2026')
    })

    it('shows a month range within one year', () => {
      expect(formatAuditPeriod(tId, range('2026-01-05', '2026-03-20'))).toBe('Januari – Maret 2026')
      expect(formatAuditPeriod(tEn, range('2026-01-05', '2026-03-20'))).toBe('January – March 2026')
    })

    it('shows both years when the audit crosses a year', () => {
      expect(formatAuditPeriod(tId, range('2025-12-01', '2026-01-31'))).toBe('Desember 2025 – Januari 2026')
      expect(formatAuditPeriod(tEn, range('2025-12-01', '2026-01-31'))).toBe('December 2025 – January 2026')
    })

    it('handles one known date, swapped dates and no dates', () => {
      expect(formatAuditPeriod(tId, range('2026-05-02', ''))).toBe('Mei 2026')
      expect(formatAuditPeriod(tId, range('', '2026-08-02'))).toBe('Agustus 2026')
      expect(formatAuditPeriod(tId, range('2026-03-01', '2026-01-01'))).toBe('Januari – Maret 2026')
      expect(formatAuditPeriod(tId, { start: null, end: null })).toBe('')
    })
  })

  describe('formatQuarterPeriod', () => {
    it('formats the selected quarter and year per locale', () => {
      expect(formatQuarterPeriod(tId, 4, 2026)).toBe('Kuartal 4 2026')
      expect(formatQuarterPeriod(tEn, 4, 2026)).toBe('Q4 2026')
      expect(formatQuarterPeriod(tId, 2)).toBe('Kuartal 2')
      expect(formatQuarterPeriod(tId, 0, 2026)).toBe('')
    })
  })

  describe('narrative text', () => {
    it('starts the Individual narrative with the audit period', () => {
      const text = buildDefaultNarrative(tId, formatAuditPeriod(tId, range('2026-01-05', '2026-01-28')))
      expect(text.startsWith('Periode Januari 2026, SPI telah menerbitkan .... LHA. Total temuan')).toBe(true)
    })

    it('starts the Compilation narrative with the quarter', () => {
      expect(buildDefaultNarrative(tId, formatQuarterPeriod(tId, 4, 2026)).startsWith('Periode Kuartal 4 2026, SPI telah menerbitkan .... LHA')).toBe(true)
      expect(buildDefaultNarrative(tEn, formatQuarterPeriod(tEn, 4, 2026)).startsWith('For the period of Q4 2026, SPI has issued .... LHA')).toBe(true)
    })

    it('leaves the prefix out when no period is known', () => {
      const text = buildDefaultNarrative(tId, '')
      expect(text.startsWith('SPI telah menerbitkan')).toBe(true)
      expect(text).not.toContain('undefined')
      expect(text).not.toContain('Periode')
    })

    it('prefixes an LHA executive summary once and keeps acronyms', () => {
      expect(prefixNarrative(tId, 'Berdasarkan hasil audit ditemukan 3 temuan.', 'Januari 2026'))
        .toBe('Periode Januari 2026, berdasarkan hasil audit ditemukan 3 temuan.')
      const once = prefixNarrative(tId, 'SPI menemukan 3 temuan.', 'Januari 2026')
      expect(once).toBe('Periode Januari 2026, SPI menemukan 3 temuan.')
      expect(prefixNarrative(tId, once, 'Januari 2026')).toBe(once)
      expect(prefixNarrative(tId, once, 'Februari – Maret 2026')).toBe('Periode Februari – Maret 2026, SPI menemukan 3 temuan.')
      expect(prefixNarrative(tId, once, '')).toBe('SPI menemukan 3 temuan.')
    })

    it('updates the period of generated text when the quarter changes, keeping edits after the prefix', () => {
      const q3 = buildDefaultNarrative(tId, formatQuarterPeriod(tId, 3, 2026))
      const q4Period = formatQuarterPeriod(tId, 4, 2026)
      expect(refreshNarrativePeriod(tId, q3, q4Period)).toBe(buildDefaultNarrative(tId, q4Period))

      const edited = 'Periode Kuartal 3 2026, SPI telah menerbitkan 12 LHA dengan catatan khusus.'
      expect(refreshNarrativePeriod(tId, edited, q4Period)).toBe('Periode Kuartal 4 2026, SPI telah menerbitkan 12 LHA dengan catatan khusus.')
    })

    it('adds the period to the untouched template once an audit period is known', () => {
      const noPeriod = buildDefaultNarrative(tId, '')
      expect(refreshNarrativePeriod(tId, noPeriod, 'Januari 2026')).toBe(buildDefaultNarrative(tId, 'Januari 2026'))
    })

    it('leaves free text and an empty narrative unchanged', () => {
      expect(refreshNarrativePeriod(tId, 'Ringkasan ditulis sendiri oleh auditor.', 'Kuartal 4 2026')).toBe('Ringkasan ditulis sendiri oleh auditor.')
      expect(refreshNarrativePeriod(tId, 'Periode ini, SPI fokus pada TI.', 'Kuartal 4 2026')).toBe('Periode ini, SPI fokus pada TI.')
      expect(refreshNarrativePeriod(tId, '', 'Kuartal 4 2026')).toBe('')
    })
  })

  describe('store and forms', () => {
    beforeEach(() => {
      setActivePinia(createPinia())
    })

    it('opens a new compilation with the selected quarter and year at the start of the narrative', () => {
      const store = useExecutiveSummaryStore()
      store.openNewCompilationForm(4, 2025)
      // Tests run with the default locale (en)
      expect(store.form.narrative.startsWith('For the period of Q4 2025, SPI has issued')).toBe(true)
    })

    it('opens a new individual summary without a period until an LHA is selected', () => {
      const store = useExecutiveSummaryStore()
      store.openNewForm(1)
      expect(store.form.narrative.startsWith('SPI has issued')).toBe(true)
      expect(store.defaultNarrativeTemplate('January – March 2026').startsWith('For the period of January – March 2026, SPI has issued')).toBe(true)
    })

    it('wires the period into both forms', () => {
      const individual = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')
      expect(individual).toContain('prefixNarrative(t, item.executiveSummary, auditPeriodForReport(item))')
      expect(individual).toContain('refreshNarrativePeriod(t, store.form.narrative, period)')
      expect(individual).toContain('store.defaultNarrativeTemplate(narrativePeriod.value)')

      const compilation = read('components/audit-result-report/ExecutiveSummaryCompilationForm.vue')
      expect(compilation).toContain('formatQuarterPeriod(t, Number(store.form.quarter) || 1, Number(store.form.tahun) || null)')
      expect(compilation).toContain('refreshNarrativePeriod(t, store.form.narrative, period)')
      expect(compilation).toContain('store.defaultNarrativeTemplate(narrativePeriod.value)')
    })

    it('keeps the narrative keys in both locales', () => {
      for (const messages of [en, id]) {
        const narrative = messages.executiveSummaryForm.narrative as Record<string, string>
        expect(narrative.periodPrefix).toContain('{period}')
        expect(narrative.quarterPeriod).toContain('{quarter}')
        expect(narrative.quarterPeriod).toContain('{year}')
        expect(narrative.quarterPeriodNoYear).toContain('{quarter}')
        expect(narrative.templateBody).toContain('SPI')
        expect(narrative.syncedBody).toContain('{lhaCount}')
        expect(narrative.syncedBody).toContain('{totalFindings}')
        expect(narrative.syncedBody).toContain('{recommendationCount}')
        expect(narrative.syncedBody).toContain('{onTimeRate}')
        expect(narrative.syncedBody).toContain('{overdueHighRiskCount}')
        expect(narrative.syncedBody).toContain('{overdueTopic}')
        expect(narrative.syncedBody).toContain('{recurringProcess}')
      }
    })

    it('synchronizes every narrative field automatically from Individual and Compilation data without dots', () => {
      const compData = deriveNarrativeSyncData({
        isCompilation: true,
        jumlahLaporan: 4,
        totalTemuan: 12,
        jumlahRekomendasi: 15,
        risikoHigh: 3,
        risikoModerate: 5,
        risikoLow: 4,
        followUpTable: [
          { status: 'Closed', jumlah: 9, persentase: 75 },
          { status: 'In Progress', jumlah: 2, persentase: 16.7 },
          { status: 'Overdue', jumlah: 1, persentase: 8.3 }
        ],
        matriksKompilasi: [
          {
            nomor: '001/SPI/2026',
            division: 'IT',
            unitKerja: 'Division IT',
            prosesBisnis: 'Keamanan Sistem & TI',
            judulTemuan: 'Celah Otentikasi Aplikasi RAS',
            nilaiRisiko: 'Very Significant',
            rekomendasi: 'Patch otentikasi',
            status: 'Overdue'
          },
          {
            nomor: '002/SPI/2026',
            division: 'IT',
            unitKerja: 'Division IT',
            prosesBisnis: 'Keamanan Sistem & TI',
            judulTemuan: 'Backup Server Terlambat',
            nilaiRisiko: 'Significant',
            rekomendasi: 'Otomasi backup',
            status: 'Closed'
          }
        ]
      })

      expect(compData).toEqual({
        lhaCount: 4,
        totalFindings: 12,
        recommendationCount: 15,
        onTimeRate: 75,
        overdueHighRiskCount: 1,
        overdueTopic: 'Celah Otentikasi Aplikasi RAS',
        recurringProcess: 'Keamanan Sistem & TI'
      })

      const enText = buildSyncedNarrative(tEn, 'October – November 2026', compData)
      expect(enText).toBe(
        'For the period of October – November 2026, SPI has issued 4 LHA. Total findings 12 with 15 recommendations. On-time completion rate 75%. There are 1 overdue/late high-risk findings related to Celah Otentikasi Aplikasi RAS that require direction from the Board of Directors. The most frequently recurring findings are in the Keamanan Sistem & TI process.'
      )
      expect(enText).not.toContain('....')
      expect(enText).not.toContain('….')

      const idText = buildSyncedNarrative(tId, 'Kuartal 4 2026', compData)
      expect(idText).toBe(
        'Periode Kuartal 4 2026, SPI telah menerbitkan 4 LHA. Total temuan 12 dengan 15 rekomendasi. Tingkat penyelesaian on-time 75%. Terdapat 1 temuan risiko tinggi overdue/terlambat terkait Celah Otentikasi Aplikasi RAS yang perlu arahan Direksi. Temuan berulang tertinggi pada proses Keamanan Sistem & TI.'
      )
      expect(isAutoSyncableNarrative(tEn, enText)).toBe(true)
      expect(isAutoSyncableNarrative(tId, idText)).toBe(true)
      expect(isAutoSyncableNarrative(tId, 'Catatan khusus auditor independen.')).toBe(false)
    })
  })

  describe('Word (.docx) generator for Approved Compilation and Individual summaries', () => {
    const baseSummary: any = {
      id: '6f1c2d3e-4b5a-4c7d-8e9f-0a1b2c3d4e5f',
      assignmentLetterId: 'ST-001/SKAI/2026',
      quarter: 1,
      periodeBulan: 'Maret',
      tahun: 2026,
      nomorDokumen: 'ESC-001/2026',
      dokumenPath: 'Compilation_Q1_2026.pdf',
      status: 'Approved',
      narrative: 'Periode Kuartal 1 2026, SPI telah menerbitkan 3 LHA. Total temuan 6 dengan 8 rekomendasi.',
      jumlahLaporan: 3,
      risikoHigh: 2,
      risikoModerateHigh: 0,
      risikoModerate: 3,
      risikoLowModerate: 0,
      risikoLow: 1,
      jumlahRekomendasi: 8,
      followUpTable: [
        { status: 'Closed', jumlah: 4, persentase: 66.7, keterangan: 'Telah divalidasi SPI' },
        { status: 'In Progress', jumlah: 1, persentase: 16.7, keterangan: 'On track' },
        { status: 'Overdue', jumlah: 1, persentase: 16.7, keterangan: '1 Risiko Tinggi' }
      ],
      topFindings: [
        {
          unitDivision: 'TI',
          judulTemuan: 'Celah Otentikasi Aplikasi RAS',
          risiko: 'Very Significant',
          statusTL: 'Overdue',
          usulan: 'Eskalasi Direksi'
        }
      ],
      matriksKompilasi: [
        {
          nomor: '001/SPI/2026',
          division: 'OP',
          unitKerja: 'Division Operation Personnel',
          prosesBisnis: 'O&M',
          judulTemuan: 'Kalibrasi Alat Ukur Gas Terlambat',
          nilaiRisiko: 'Significant',
          rekomendasi: 'Lakukan kalibrasi Q2',
          dueDate: '30/06/2026',
          picUnit: 'Manager O&M',
          progres: 60,
          status: 'In Progress',
          buktiTL: 'BA Kalibrasi 1'
        }
      ],
      akarMasalah: 'Tema berulang periode ini: 1. Pengelolaan Aset Tetap; 2. Kepatuhan Kontrak.',
      kesimpulan: 'Secara umum pengendalian membaik, indikator on-time di atas target.',
      signatureTempat: 'Jakarta',
      signatureTanggal: '2026-03-31',
      signatureNamaKepala: 'Budi Santoso, CIA',
      signatureNIK: 'SPI-77621'
    }

    it('blocks Word (.docx) generation when status is Draft and allows it only when Approved', async () => {
      setActivePinia(createPinia())
      const store = useExecutiveSummaryStore()

      const draftSummary = { ...baseSummary, status: 'Draft' }
      expect(canGenerateExecutiveSummaryDocx(draftSummary)).toBe(false)
      expect(() => buildExecutiveSummaryDocumentXml(draftSummary)).toThrow(/Approved/)

      const okDraft = await store.downloadDocx(draftSummary)
      expect(okDraft).toBe(false)

      expect(canGenerateExecutiveSummaryDocx(baseSummary)).toBe(true)
    })

    it('renders the Compilation Word template with Sections I-VII and Lampiran Matriks Induk Kompilasi Temuan', () => {
      const xml = buildExecutiveSummaryDocumentXml(baseSummary, { variant: 'compilation' })
      expect(xml).toContain('EXECUTIVE SUMMARY COMPILATION')
      expect(xml).toContain('HASIL AUDIT TRIWULANAN')
      expect(xml).toContain('II. STATISTIK KOMPILASI')
      expect(xml).toContain('Jumlah Laporan: Internal 3 LHA.')
      expect(xml).toContain('III. STATUS TINDAK LANJUT')
      expect(xml).toContain('IV. TOP 5 TEMUAN SIGNIFIKAN')
      expect(xml).toContain('Celah Otentikasi Aplikasi RAS')
      expect(xml).toContain('V. ANALISIS TEMUAN BERULANG')
      expect(xml).toContain('VI. TREN &amp; GRAFIK')
      expect(xml).toContain('VII. KESIMPULAN &amp; REKOMENDASI UNTUK MANAJEMEN')
      expect(xml).toContain('MATRIKS INDUK KOMPILASI TEMUAN')
      expect(xml).toContain('Kalibrasi Alat Ukur Gas Terlambat')
      expect(xml).toContain('Budi Santoso, CIA')

      const zipBytes = generateExecutiveSummaryDocxBytes(baseSummary, { variant: 'compilation' })
      // PK\x03\x04 ZIP magic bytes
      expect(zipBytes[0]).toBe(0x50)
      expect(zipBytes[1]).toBe(0x4b)
      expect(executiveSummaryDocxFilename(baseSummary, 'compilation')).toBe('Executive_Summary_Compilation_ESC-001_2026.docx')
    })

    it('renders the Individual Word template tailored to Individual LHA data without Compilation matrix attachment', () => {
      const indSummary = {
        ...baseSummary,
        nomorDokumen: 'LHA-001/SKAI/2026',
        jumlahLaporan: 1,
        matriksKompilasi: []
      }
      const xml = buildExecutiveSummaryDocumentXml(indSummary, {
        variant: 'individual',
        reportTitle: 'Laporan Hasil Audit Operasional Keuangan',
        department: 'Finance (FIN)'
      })
      expect(xml).toContain('EXECUTIVE SUMMARY INDIVIDUAL')
      expect(xml).toContain('LAPORAN HASIL AUDIT (LHA)')
      expect(xml).toContain('Nomor LHA: LHA-001/SKAI/2026')
      expect(xml).toContain('Nomor Surat Tugas: ST-001/SKAI/2026')
      expect(xml).toContain('Judul Laporan / Objek Audit: Laporan Hasil Audit Operasional Keuangan')
      expect(xml).toContain('II. INFORMASI &amp; STATISTIK LHA INDIVIDUAL')
      expect(xml).toContain('III. STATUS TINDAK LANJUT')
      expect(xml).toContain('IV. TOP 5 TEMUAN SIGNIFIKAN')
      expect(xml).toContain('V. ANALISIS TEMUAN BERULANG &amp; AKAR MASALAH')
      expect(xml).toContain('VI. KESIMPULAN &amp; REKOMENDASI UNTUK MANAJEMEN')
      expect(xml).not.toContain('MATRIKS INDUK KOMPILASI TEMUAN')
      expect(executiveSummaryDocxFilename(indSummary, 'individual')).toBe('Executive_Summary_Individual_LHA-001_SKAI_2026.docx')
    })

    it('only renders the Word generator buttons in UI when status is Approved', () => {
      const indPage = read('pages/executive-summary/index.vue')
      const compPage = read('pages/executive-summary-compilation/index.vue')
      const indForm = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')
      const compForm = read('components/audit-result-report/ExecutiveSummaryCompilationForm.vue')

      expect(indPage).toContain('v-if="item.status === \'Approved\' && canExport"')
      expect(compPage).toContain('v-if="item.status === \'Approved\' && canExport"')
      expect(indForm).toContain('v-if="isLocked && canExport"')
      expect(compForm).toContain('v-if="isLocked && canExport"')
    })
  })

  describe('Executive Summary form cleanup, LHA validation, persistence and approval', () => {
    it('removes redundant SPI Electronic Signature block from Recurring Findings & Conclusion in both forms', () => {
      const indForm = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')
      const compForm = read('components/audit-result-report/ExecutiveSummaryCompilationForm.vue')

      expect(indForm).not.toContain("t('executiveSummaryForm.signature.title')")
      expect(compForm).not.toContain("t('executiveSummaryForm.signature.title')")
      expect(indForm).not.toContain('v-model="store.form.signatureNIK"')
      expect(compForm).not.toContain('v-model="store.form.signatureNIK"')
    })

    it('enforces required Select Assignment Letter / LHA validation in Individual form and page header', () => {
      const indForm = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')
      const indPage = read('pages/executive-summary/index.vue')

      expect(indForm).toContain(':label="t(\'executiveSummaryForm.upload.lhaLabel\')"')
      expect(indForm).toContain(':error="lhaValidationError"')
      expect(indForm).toContain('t(\'executiveSummaryForm.validation.lhaRequired\')')
      expect(indForm).toContain('store.saveForm({ requireLha: true })')
      expect(indPage).toContain('ref="individualFormRef"')
      expect(indPage).not.toContain('@click="store.saveForm"')
    })

    it('strips empty approvedAt/timestamp fields on save so backend RFC3339 parsing never fails, and supports ensureSavedOnServer before approval', async () => {
      localStorage.clear()
      setActivePinia(createPinia())
      const sentBodies: any[] = []
      const createdIds = [
        '7a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d',
        '8b3c4d5e-6f7a-4b8c-9d0e-1f2a3b4c5d6e'
      ]
      let nextIdIdx = 0
      const savedRows: any[] = []

      ;(globalThis as any).$fetch = async (url: string, opts?: any) => {
        if (opts?.method === 'POST' && url.endsWith('/executive-summaries')) {
          sentBodies.push(opts.body)
          const id = createdIds[nextIdIdx++]!
          const row = { ...opts.body, id }
          savedRows.push(row)
          return { success: true, data: row }
        }
        const approveMatch = opts?.method === 'POST' && url.match(/\/executive-summaries\/([^/]+)\/approve$/)
        if (approveMatch) {
          const targetRow = savedRows.find(r => r.id === approveMatch[1])
          if (targetRow) targetRow.status = 'Approved'
          return { success: true }
        }
        return { data: { items: [...savedRows] } }
      }

      const store = useExecutiveSummaryStore()

      // Individual form save without LHA is rejected
      store.openNewForm(1)
      const rejected = await store.saveForm({ requireLha: true })
      expect(rejected).toBe(false)
      expect(sentBodies).toHaveLength(0)

      // Compilation form save strips approvedAt and persists cleanly
      store.openNewCompilationForm(2, 2026)
      store.form.dokumenPath = 'Compilation_Q2_2026.pdf'
      const savedOk = await store.saveForm()
      expect(savedOk).toBe(true)
      expect(sentBodies).toHaveLength(1)
      expect(sentBodies[0]).not.toHaveProperty('approvedAt')
      expect(sentBodies[0]).not.toHaveProperty('created_at')
      expect(sentBodies[0]).not.toHaveProperty('updated_at')
      expect(store.summaryList.some(s => s.id === createdIds[0] && s.nomorDokumen === 'ESC-001/2026')).toBe(true)

      // LHA-derived virtual Draft card (ES-...) gets persisted via ensureSavedOnServer before approve
      store.openView({
        ...store.form,
        id: 'ES-virtual-lha-id',
        nomorDokumen: 'LHA-099/SKAI/2026',
        status: 'Draft'
      })
      const ensured = await store.ensureSavedOnServer()
      expect(ensured).toBe(true)
      expect(store.currentSummary?.id).toBe(createdIds[1])
      const approved = await store.approveWithSignature({
        signaturePlace: 'Jakarta',
        signatureDate: '2026-06-30',
        signatureImage: 'data:image/png;base64,iVBORw0KGgo=',
        signerName: 'Kepala SPI',
        signerRole: 'CAE'
      })
      expect(approved).toBe(true)
      expect(store.currentSummary?.status).toBe('Approved')
    })
  })
})
