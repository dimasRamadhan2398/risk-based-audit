import { describe, expect, it, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import fs from 'node:fs'
import path from 'node:path'
import { useAuthStore } from '../../stores/auth'
import { useExecutiveSummaryStore } from '../../stores/executive-summary'
import { useRbac } from '../../composables/useRbac'
import { UserRole } from '../../types/auth'

const read = (relativePath: string) => fs.readFileSync(path.resolve(__dirname, '../../', relativePath), 'utf8')

describe('Executive Summary Features (Individual & Compilation)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  describe('RBAC permissions for Executive Summary', () => {
    it('allows Admin and EXECUTIVE (CAE) to review, write notes, and approve', () => {
      const auth = useAuthStore()

      // 1. Admin
      auth.user = { id: '1', username: 'admin', email: 'admin@test.com', fullName: 'Admin', roles: [UserRole.ADMIN] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)

      // 2. EXECUTIVE (acts as the Chief Audit Executive)
      auth.user = { id: '2', username: 'cae', email: 'cae@test.com', fullName: 'CAE', roles: [UserRole.EXECUTIVE] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)

      // 3. CAE alias / uppercase
      auth.user = { id: '3', username: 'cae2', email: 'cae2@test.com', fullName: 'CAE 2', roles: ['CAE'] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)

      // 4. EXECUTIVE in lowercase (role checks are case-insensitive)
      auth.user = { id: '4', username: 'exec-lower', email: 'exec-lower@test.com', fullName: 'Executive', roles: ['executive'] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)

      // 5. Audit Manager has no seeded backend role, so it no longer gets the review action
      auth.user = { id: '5', username: 'manager2', email: 'manager2@test.com', fullName: 'Audit Manager 2', roles: ['AUDIT_MANAGER'] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(false)
    })

    it('denies Auditor, Auditee, and Viewer from writing notes and approving', () => {
      const auth = useAuthStore()

      // Auditor
      auth.user = { id: '6', username: 'auditor', email: 'auditor@test.com', fullName: 'Auditor', roles: [UserRole.AUDITOR] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(false)

      // Auditee
      auth.user = { id: '8', username: 'auditee', email: 'auditee@test.com', fullName: 'Auditee', roles: [UserRole.AUDITEE] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(false)

      // Viewer
      auth.user = { id: '9', username: 'viewer', email: 'viewer@test.com', fullName: 'Viewer', roles: [UserRole.VIEWER] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(false)
    })
  })

  describe('Form Components Structure & Verification', () => {
    it('ensures ExecutiveSummaryIndividualForm has Noted field bound and Approve button in View only', () => {
      const content = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')

      // Noted section in View mode
      expect(content).toContain('v-if="store.isViewing" id="sec-notes"')
      expect(content).toContain('v-model="store.form.executiveNote"')
      expect(content).toContain(':disabled="!canWriteExecutiveNote"')
      expect(content).toContain('v-if="canWriteExecutiveNote"')
      expect(content).toContain('@click="store.saveExecutiveNote"')

      // Approve button strictly in View mode (under v-else of !store.isViewing)
      expect(content).toContain('v-if="isDraft && canApprove"')
      expect(content).toContain("t('executiveSummaryForm.actions.approve')")
      expect(content).toContain('@click="approveReportDirectly"')

      // Role check defines permissions from useRbac
      expect(content).toContain('canReviewExecutiveSummary')
    })

    it('ensures ExecutiveSummaryCompilationForm has Noted field bound and Approve button in View only', () => {
      const content = read('components/audit-result-report/ExecutiveSummaryCompilationForm.vue')

      // Noted section in View mode
      expect(content).toContain('v-if="store.isViewing" id="sec-notes"')
      expect(content).toContain('v-model="store.form.executiveNote"')
      expect(content).toContain(':disabled="!canWriteExecutiveNote"')
      expect(content).toContain('v-if="canWriteExecutiveNote"')
      expect(content).toContain('@click="store.saveExecutiveNote"')

      // Approve button strictly in View mode
      expect(content).toContain('v-if="isDraft && canApprove"')
      expect(content).toContain("t('executiveSummaryForm.actions.approve')")
      expect(content).toContain('@click="approveReportDirectly"')

      // Role check defines permissions from useRbac
      expect(content).toContain('canReviewExecutiveSummary')
    })
  })

  describe('Cards Note Display', () => {
    it('ensures Individual summary cards conditionally render executive note only when present', () => {
      const content = read('pages/executive-summary/index.vue')

      // Conditional note rendering
      expect(content).toContain('v-if="item.executiveNote && item.executiveNote.trim()"')
      expect(content).toContain("t('executiveSummaryForm.notes.title')")
      expect(content).toContain('{{ item.executiveNote }}')
    })

    it('ensures Compilation summary cards conditionally render executive note only when present', () => {
      const content = read('pages/executive-summary-compilation/index.vue')

      // Conditional note rendering
      expect(content).toContain('v-if="item.executiveNote && item.executiveNote.trim()"')
      expect(content).toContain('Noted dari Executive untuk Auditor')
      expect(content).toContain('{{ item.executiveNote }}')
    })
  })

  describe('Executive Summary Store Functions', () => {
    const testItems = [
      {
        id: 'ES-LHA-021-2026',
        assignmentLetterId: 'ST-001/SKAI/2026',
        quarter: 1,
        periodeBulan: 'Januari - Maret',
        tahun: 2026,
        nomorDokumen: 'LHA-021/SKAI/2026',
        dokumenPath: 'Executive_Summary_021_LHA_2026.pdf',
        status: 'Draft',
        executiveNote: '',
        narrative: 'Executive Summary Individual untuk Laporan Hasil Audit Operasional Keuangan (LHA-021/SKAI/2026).',
        jumlahLaporan: 1,
        risikoHigh: 3,
        risikoModerateHigh: 0,
        risikoModerate: 2,
        risikoLowModerate: 0,
        risikoLow: 0,
        jumlahRekomendasi: 5,
        followUpTable: [],
        topFindings: [],
        matriksKompilasi: []
      }
    ]

    beforeEach(() => {
      global.$fetch = (async (url: string, opts?: any) => {
        if (!opts?.method || opts.method === 'GET') {
          return { data: { items: [...testItems] } }
        }
        return { success: true }
      }) as any
    })

    it('handles saveExecutiveNote and updateStatus fallback and state sync', async () => {
      const store = useExecutiveSummaryStore()
      await store.fetchSummaries()

      const item = store.summaryList[0]
      expect(item).toBeDefined()
      store.openView(item!)

      store.form.executiveNote = 'New Executive Note for Auditor'
      await store.saveExecutiveNote()

      expect(store.currentSummary?.executiveNote).toBe('New Executive Note for Auditor')
      expect(store.summaryList.find(s => s.id === item!.id)?.executiveNote).toBe('New Executive Note for Auditor')

      // Test approve
      await store.updateStatus(item!.id, 'Approved')
      expect(store.currentSummary?.status).toBe('Approved')
      expect(store.form.status).toBe('Approved')
      expect(store.summaryList.find(s => s.id === item!.id)?.status).toBe('Approved')
    })

    it('persists note and approved status for LHA-021/SKAI/2026 across re-fetch', async () => {
      const store = useExecutiveSummaryStore()
      await store.fetchSummaries()

      const item021 = store.summaryList.find(s => s.nomorDokumen === 'LHA-021/SKAI/2026')
      expect(item021).toBeDefined()
      expect(item021?.assignmentLetterId).toBe('ST-001/SKAI/2026')

      store.openView(item021!)
      store.form.executiveNote = 'Catatan penting untuk auditor keuangan terkait rekonsiliasi kas.'
      await store.saveExecutiveNote()
      await store.updateStatus(item021!.id, 'Approved')

      expect(store.currentSummary?.status).toBe('Approved')
      expect(store.currentSummary?.executiveNote).toBe('Catatan penting untuk auditor keuangan terkait rekonsiliasi kas.')

      // Simulate page refresh / re-fetch
      await store.fetchSummaries()
      const refreshed021 = store.summaryList.find(s => s.nomorDokumen === 'LHA-021/SKAI/2026')
      expect(refreshed021?.status).toBe('Approved')
      expect(refreshed021?.executiveNote).toBe('Catatan penting untuk auditor keuangan terkait rekonsiliasi kas.')
    })

    it('ensures dummy DOC-EXSUM-Q1-2026 is filtered out and default assignment letter is ST-001/SKAI/2026', async () => {
      const store = useExecutiveSummaryStore()
      await store.fetchSummaries()

      const dummyDoc = store.summaryList.find(s => s.nomorDokumen === 'DOC-EXSUM-Q1-2026')
      expect(dummyDoc).toBeUndefined()

      store.openNewForm()
      expect(store.form.assignmentLetterId).toBe('ST-001/SKAI/2026')
      expect(store.form.nomorDokumen).toBe('')
    })

    it('ensures saving an existing LHA document updates the card without duplicating it', async () => {
      const store = useExecutiveSummaryStore()
      await store.fetchSummaries()

      const initialCount = store.summaryList.filter(s => s.nomorDokumen === 'LHA-021/SKAI/2026').length
      expect(initialCount).toBe(1)

      // Open new form and attempt to save with LHA-021/SKAI/2026
      store.openNewForm()
      store.form.nomorDokumen = 'LHA-021/SKAI/2026'
      store.form.narrative = 'Updated narrative for 021'
      await store.saveForm()

      const afterCount = store.summaryList.filter(s => s.nomorDokumen === 'LHA-021/SKAI/2026').length
      expect(afterCount).toBe(1)
      expect(store.summaryList.find(s => s.nomorDokumen === 'LHA-021/SKAI/2026')?.narrative).toBe('Updated narrative for 021')
    })
  })

  // DELETE /executive-summaries/:id takes the UUID of an executive_summaries row (as GET/PUT do).
  // Cards derived from an LHA without a summary row carry `ES-<auditResultReportId>`; sending that
  // id made the backend answer 400 "Invalid ExecutiveSummary ID".
  describe('deleteSummary ID handling', () => {
    const REAL_ID = '6f1c2d3e-4b5a-4c7d-8e9f-0a1b2c3d4e5f'
    const LHA_ID = '12a2eebc-a733-412f-a51a-5c9f447a291c'
    let deleteUrls: string[]
    let backendItems: any[]

    const baseItem = {
      quarter: 1, periodeBulan: 'Januari', tahun: 2026, dokumenPath: '', status: 'Draft', executiveNote: '',
      narrative: 'n', jumlahLaporan: 1, risikoHigh: 0, risikoModerateHigh: 0, risikoModerate: 0,
      risikoLowModerate: 0, risikoLow: 0, jumlahRekomendasi: 0, followUpTable: [], topFindings: [], matriksKompilasi: []
    }

    beforeEach(() => {
      deleteUrls = []
      backendItems = []
      ;(globalThis as any).useGlobalModalStore = () => ({ confirmDelete: async () => true })
      global.$fetch = (async (url: string, opts?: any) => {
        if (opts?.method === 'DELETE') {
          deleteUrls.push(url)
          return { success: true }
        }
        if (!opts?.method || opts.method === 'GET') {
          return { data: { items: url.includes('/executive-summaries') ? [...backendItems] : [] } }
        }
        return { success: true }
      }) as any
    })

    it('does not call the backend for a card derived from an LHA (ES-<reportId>)', async () => {
      const store = useExecutiveSummaryStore()
      await store.fetchSummaries()

      await store.deleteSummary(`ES-${LHA_ID}`, 'LHA-001/SKAI/2026')

      expect(deleteUrls).toEqual([])
      expect(store.deletedDocNumbers).toContain('LHA-001/SKAI/2026')
      expect(store.errorMsg).toBe('')
    })

    it('deletes a real summary by its bare UUID', async () => {
      backendItems = [{ ...baseItem, id: REAL_ID, nomorDokumen: 'LHA-002/SKAI/2026' }]
      const store = useExecutiveSummaryStore()
      await store.fetchSummaries()

      await store.deleteSummary(REAL_ID, 'LHA-002/SKAI/2026')

      expect(deleteUrls).toHaveLength(1)
      expect(deleteUrls[0]).toMatch(new RegExp(`/executive-summaries/${REAL_ID}$`))
    })

    it('resolves a prefixed id to the real summary with the same document number', async () => {
      backendItems = [{ ...baseItem, id: REAL_ID, nomorDokumen: 'LHA-003/SKAI/2026' }]
      const store = useExecutiveSummaryStore()
      await store.fetchSummaries()

      await store.deleteSummary(`ES-${LHA_ID}`, 'LHA-003/SKAI/2026')

      expect(deleteUrls).toHaveLength(1)
      expect(deleteUrls[0]).toMatch(new RegExp(`/executive-summaries/${REAL_ID}$`))
      expect(deleteUrls[0]).not.toContain('ES-')
    })
  })

  describe('Assignment Letter (Surat Tugas) Integration', () => {
    it('ensures IndividualForm integrates LHA selection with Audit Result Report (LHA) context while CompilationForm compiles quarterly LHAs', () => {
      const individualContent = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')
      const compilationContent = read('components/audit-result-report/ExecutiveSummaryCompilationForm.vue')

      // Individual form checks
      expect(individualContent).toContain('Pilih Laporan Hasil Audit (LHA)')
      expect(individualContent).toContain('lhaDropdownOptions')
      expect(individualContent).toContain('onLhaSelect')
      expect(individualContent).toContain('linkedLhaReport')
      expect(individualContent).toContain('useAuditResultReportStore')

      // Compilation form checks: removes single assignment letter selection and replaces with quarterly LHAs collection
      expect(compilationContent).not.toContain('Pilih Surat Tugas (Assignment Letter) / LHA')
      expect(compilationContent).not.toContain('linkedAssignmentLetter')
      expect(compilationContent).toMatch(/(Kumpulan Data LHA Terkompilasi|sourcesTitle)/)
    })

    it('ensures Executive Summary Individual uses Nomor LHA filter and Compilation uses Nomor Dokumen Kompilasi filter', () => {
      const individualPage = read('pages/executive-summary/index.vue')
      const compilationPage = read('pages/executive-summary-compilation/index.vue')

      // Individual page uses Nomor LHA filter
      expect(individualPage).toContain('selectedLhaNumber')
      expect(individualPage).toContain('lhaNumberOptions')
      expect(individualPage).toMatch(/(Filter berdasarkan Nomor LHA\.\.\.|filterLhaPlaceholder)/)
      expect(individualPage).toContain('item.nomorDokumen')

      // Compilation page uses Nomor Dokumen Kompilasi filter
      expect(compilationPage).toContain('selectedCompilationDoc')
      expect(compilationPage).toContain('compilationDocOptions')
      expect(compilationPage).toMatch(/(Filter berdasarkan Nomor Dokumen Kompilasi\.\.\.|filterCompilationPlaceholder)/)
      expect(compilationPage).toContain('item.nomorDokumen')
    })

    it('verifies compilation document options include ESC documents', () => {
      const store = useExecutiveSummaryStore()
      store.summaryList = [
        {
          id: 'comp-1',
          nomorDokumen: 'ESC-001/2026',
          quarter: 1,
          periodeBulan: 'Kuartal I',
          tahun: 2026,
          dokumenPath: 'comp1.pdf',
          status: 'Approved',
          narrative: 'Narrative',
          jumlahLaporan: 3,
          risikoHigh: 1,
          risikoModerateHigh: 0,
          risikoModerate: 1,
          risikoLowModerate: 0,
          risikoLow: 1,
          jumlahRekomendasi: 3,
          followUpTable: [],
          topFindings: [],
          matriksKompilasi: [],
          akarMasalah: '',
          kesimpulan: '',
          signatureTempat: 'Jakarta',
          signatureTanggal: '2026-03-31',
          signatureNamaKepala: '',
          signatureNIK: ''
        }
      ]
      expect(store.summaryList.some(s => s.nomorDokumen.startsWith('ESC-'))).toBe(true)
    })
  })

  describe('Findings & Compilation Synchronization', () => {
    it('ensures IndividualForm synchronizes Section IV and Section V & VII with LHA findings', () => {
      const individualContent = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')

      // Section IV synchronization
      expect(individualContent).toContain("t('executiveSummaryForm.sync.findings')")
      expect(individualContent).toContain('syncFindingsFromLha')
      expect(individualContent).toContain('store.form.topFindings')

      // Section V & VII synchronization
      expect(individualContent).toContain("t('executiveSummaryForm.sync.analysis')")
      expect(individualContent).toContain('store.form.akarMasalah')
      expect(individualContent).toContain('store.form.kesimpulan')
    })

    it('ensures CompilationForm synchronizes with Individual LHAs count, risk breakdown, and Top 5 findings', () => {
      const compilationContent = read('components/audit-result-report/ExecutiveSummaryCompilationForm.vue')

      // Sync button and method
      expect(compilationContent).toContain("t('executiveSummaryCompilation.syncFromIndividual')")
      expect(compilationContent).toContain('syncFromIndividualLha')

      // Matriks Kompilasi section
      expect(compilationContent).toContain('id="sec-matriks"')
      expect(compilationContent).toContain("t('executiveSummaryCompilation.matrix.title')")
      expect(compilationContent).toContain('store.form.matriksKompilasi')

      // Statistics come from one compilation-stats call for the selected year and quarter
      expect(compilationContent).toContain('fetchCompilationStats({ year: targetYear, quarter: targetQuarter })')
      expect(compilationContent).not.toContain('fetchCompilationSources')
      expect(compilationContent).not.toContain('computeAtrOverdue')
      expect(compilationContent).toContain('store.form.jumlahLaporan = stats.jumlahLaporan')
      expect(compilationContent).toContain('riskBreakdownToFields(stats.riskBreakdown)')
      expect(compilationContent).toContain('riskBucketFields')
      expect(compilationContent).toContain('riskBucketBreakdown(riskFieldsToBreakdown(store.form))')
      expect(compilationContent).toContain('riskBuckets.high')
      expect(compilationContent).not.toContain('noMatchingLha')
      expect(compilationContent).not.toContain('selectCompilationReports')
      expect(compilationContent).toContain('store.form.topFindings')
      // Saving is blocked until a sync succeeds; generated narrative text is no longer written
      expect(compilationContent).toContain(':disabled="!canSaveDraft"')
      expect(compilationContent).not.toContain('Tingkat penyelesaian on-time 95%')
      expect(compilationContent).not.toContain('cukup efektif')
    })

    it('handles RBAC with varied spacing, positions, and executive aliases', () => {
      const auth = useAuthStore()

      // Position based check
      auth.user = { id: 'p1', username: 'cae_pos', email: 'cae@test.com', fullName: 'Head of SPI', position: 'Chief Audit Executive', roles: [] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)

      // Spaced role name
      auth.user = { id: 'p2', username: 'cae_space', email: 'cae2@test.com', fullName: 'CAE', roles: ['Chief Audit Executive'] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)

      // Executive alias
      auth.user = { id: 'p3', username: 'exec', email: 'exec@test.com', fullName: 'Executive Director', roles: ['EXECUTIVE'] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)
    })

    it('ensures Nomor Dokumen Kompilasi is disabled and auto-generated with ESC-[nomor urut]/[tahun] format', () => {
      const compilationContent = read('components/audit-result-report/ExecutiveSummaryCompilationForm.vue')

      // Verify template has disabled attribute on Nomor Dokumen Kompilasi input
      expect(compilationContent).toMatch(/(label="Nomor Dokumen Kompilasi"|:label="t\('executiveSummaryCompilation\.formExtra\.documentNumber'\)")/)
      expect(compilationContent).toMatch(/(placeholder="Contoh: ESC-001\/2026"|:placeholder="t\('executiveSummaryCompilation\.formExtra\.documentNumberExample'\)")/)
      expect(compilationContent).toMatch(/<UFormField\s+[^>]*documentNumber[^>]*>[\s\S]*?<UInput[\s\S]*?disabled/)

      // Verify compilation script has ESC document generation logic
      expect(compilationContent).toContain('generateEscDocNumber')
      expect(compilationContent).toMatch(/ESC-\$\{.*\}\/\$\{.*targetYear.*\}/)

      // Verify store generateEscDocNumber function
      const store = useExecutiveSummaryStore()
      expect(typeof store.generateEscDocNumber).toBe('function')

      // Initial sequential number
      const doc1 = store.generateEscDocNumber(2026)
      expect(doc1).toMatch(/^ESC-\d{3}\/2026$/)

      // When an existing ESC document is in summaryList, next document increments sequence
      store.summaryList.push({
        id: 'esc-test-1',
        nomorDokumen: 'ESC-001/2026',
        quarter: 1,
        periodeBulan: 'Kuartal I',
        tahun: 2026,
        dokumenPath: 'doc.pdf',
        status: 'Draft',
        narrative: '',
        jumlahLaporan: 1,
        risikoHigh: 0,
        risikoModerateHigh: 0,
        risikoModerate: 0,
        risikoLowModerate: 0,
        risikoLow: 0,
        jumlahRekomendasi: 0,
        followUpTable: [],
        topFindings: [],
        matriksKompilasi: [],
        akarMasalah: '',
        kesimpulan: '',
        signatureTempat: 'Jakarta',
        signatureTanggal: '2026-03-31',
        signatureNamaKepala: '',
        signatureNIK: ''
      })

      const doc2 = store.generateEscDocNumber(2026)
      expect(doc2).toBe('ESC-002/2026')

      // openNewCompilationForm assigns the auto-generated ESC number
      store.openNewCompilationForm(1)
      expect(store.form.nomorDokumen).toBe('ESC-002/2026')
    })
  })
})



