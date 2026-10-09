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
    it('allows Admin, CAE, and Audit Manager to review, write notes, and approve', () => {
      const auth = useAuthStore()

      // 1. Admin
      auth.user = { id: '1', username: 'admin', email: 'admin@test.com', fullName: 'Admin', roles: [UserRole.ADMIN] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)

      // 2. Chief Audit Executive (lowercase)
      auth.user = { id: '2', username: 'cae', email: 'cae@test.com', fullName: 'CAE', roles: [UserRole.CHIEF_AUDIT_EXECUTIVE] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)

      // 3. CAE alias / uppercase
      auth.user = { id: '3', username: 'cae2', email: 'cae2@test.com', fullName: 'CAE 2', roles: ['CAE'] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)

      // 4. Audit Manager
      auth.user = { id: '4', username: 'manager', email: 'manager@test.com', fullName: 'Audit Manager', roles: [UserRole.AUDIT_MANAGER] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)

      // 5. Audit Manager uppercase
      auth.user = { id: '5', username: 'manager2', email: 'manager2@test.com', fullName: 'Audit Manager 2', roles: ['AUDIT_MANAGER'] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)
    })

    it('denies Auditor, Auditee, and Viewer from writing notes and approving', () => {
      const auth = useAuthStore()

      // Auditor
      auth.user = { id: '6', username: 'auditor', email: 'auditor@test.com', fullName: 'Auditor', roles: [UserRole.AUDITOR] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(false)

      // Audit Staff
      auth.user = { id: '7', username: 'staff', email: 'staff@test.com', fullName: 'Staff', roles: [UserRole.AUDIT_STAFF] }
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
      expect(content).toContain('label="Setujui (Approve)"')
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
      expect(content).toContain('label="Setujui (Approve)"')
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
      expect(content).toContain('Noted dari Executive untuk Auditor')
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

  describe('Assignment Letter (Surat Tugas) Integration', () => {
    it('ensures IndividualForm and CompilationForm integrate Assignment Letter selection and context banner', () => {
      const individualContent = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')
      const compilationContent = read('components/audit-result-report/ExecutiveSummaryCompilationForm.vue')

      // Individual form checks
      expect(individualContent).toContain('Pilih Surat Tugas (Assignment Letter)')
      expect(individualContent).toContain('assignmentLetterDropdownOptions')
      expect(individualContent).toContain('onAssignmentLetterSelect')
      expect(individualContent).toContain('linkedAssignmentLetter')
      expect(individualContent).toContain('useAssignmentLetterStore')

      // Compilation form checks
      expect(compilationContent).toContain('Pilih Surat Tugas (Assignment Letter)')
      expect(compilationContent).toContain('assignmentLetterDropdownOptions')
      expect(compilationContent).toContain('onAssignmentLetterSelect')
      expect(compilationContent).toContain('linkedAssignmentLetter')
      expect(compilationContent).toContain('useAssignmentLetterStore')
    })

    it('ensures Executive Summary pages have Assignment Letter filter and badges', () => {
      const individualPage = read('pages/executive-summary/index.vue')
      const compilationPage = read('pages/executive-summary-compilation/index.vue')

      // Individual page
      expect(individualPage).toContain('selectedAssignmentLetter')
      expect(individualPage).toContain('assignmentLetterOptions')
      expect(individualPage).toContain('item.assignmentLetterId')

      // Compilation page
      expect(compilationPage).toContain('selectedAssignmentLetter')
      expect(compilationPage).toContain('assignmentLetterOptions')
      expect(compilationPage).toContain('item.assignmentLetterId')
    })
  })

  describe('Findings & Compilation Synchronization', () => {
    it('ensures IndividualForm synchronizes Section IV and Section V & VII with LHA findings', () => {
      const individualContent = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')

      // Section IV synchronization
      expect(individualContent).toContain('Sinkronkan Temuan dari LHA')
      expect(individualContent).toContain('syncFindingsFromLha')
      expect(individualContent).toContain('store.form.topFindings')

      // Section V & VII synchronization
      expect(individualContent).toContain('Sinkronkan Analisis dari LHA')
      expect(individualContent).toContain('store.form.akarMasalah')
      expect(individualContent).toContain('store.form.kesimpulan')
    })

    it('ensures CompilationForm synchronizes with Individual LHAs count, risk breakdown, and Top 5 findings', () => {
      const compilationContent = read('components/audit-result-report/ExecutiveSummaryCompilationForm.vue')

      // Sync button and method
      expect(compilationContent).toContain('Sinkronkan dari LHA Individual')
      expect(compilationContent).toContain('syncFromIndividualLha')

      // Matriks Kompilasi section
      expect(compilationContent).toContain('id="sec-matriks"')
      expect(compilationContent).toContain('Section VIII: Matriks Kompilasi Temuan LHA Individual')
      expect(compilationContent).toContain('store.form.matriksKompilasi')

      // Synchronized stats
      expect(compilationContent).toContain('store.form.jumlahLaporan')
      expect(compilationContent).toContain('store.form.risikoTinggi')
      expect(compilationContent).toContain('store.form.risikoSedang')
      expect(compilationContent).toContain('store.form.risikoRendah')
      expect(compilationContent).toContain('store.form.topFindings')
    })

    it('handles RBAC with varied spacing, positions, and executive aliases', () => {
      const auth = useAuthStore()

      // Position based check
      auth.user = { id: 'p1', username: 'cae_pos', email: 'cae@test.com', fullName: 'Head of SPI', position: 'Chief Audit Executive', roles: [] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)

      // Spaced role name
      auth.user = { id: 'p2', username: 'mgr_space', email: 'mgr@test.com', fullName: 'Manager', roles: ['Audit Manager'] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)

      // Executive alias
      auth.user = { id: 'p3', username: 'exec', email: 'exec@test.com', fullName: 'Executive Director', roles: ['EXECUTIVE'] }
      expect(useRbac().canReviewExecutiveSummary.value).toBe(true)
    })
  })
})


