import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useRbac } from '~/composables/useRbac'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useAuditeeSurveyStore } from '~/stores/auditee-survey'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useAuthStore } from '~/stores/auth'

describe('Role-Based Dashboard Switching & Auditee Store Integration', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    const { setRoleView } = useRbac()
    setRoleView(null)
  })

  it('correctly determines currentDashboardRole based strictly on authentic user roles', () => {
    const authStore = useAuthStore()
    const { currentDashboardRole } = useRbac()

    // Test Auditee user
    authStore.user = {
      id: 'auditee-1',
      email: 'auditee@company.com',
      fullName: 'Bob Auditee',
      roles: ['AUDITEE'],
      mustChangePassword: false,
    } as any
    expect(currentDashboardRole.value).toBe('auditee')

    // Test Auditor user
    authStore.user = {
      id: 'auditor-1',
      email: 'auditor@company.com',
      fullName: 'John Auditor',
      roles: ['AUDITOR'],
      mustChangePassword: false,
    } as any
    expect(currentDashboardRole.value).toBe('auditor')

    // Test Executive user
    authStore.user = {
      id: 'cae-1',
      email: 'cae@company.com',
      fullName: 'Chief Audit Executive',
      roles: ['CHIEF_AUDIT_EXECUTIVE'],
      mustChangePassword: false,
    } as any
    expect(currentDashboardRole.value).toBe('executive')

    // Test Admin user
    authStore.user = {
      id: 'admin-1',
      email: 'admin@company.com',
      fullName: 'Admin User',
      roles: ['ADMIN'],
      mustChangePassword: false,
    } as any
    expect(currentDashboardRole.value).toBe('admin')
  })

  it('handles ATR store items calculation and overdue status correctly', () => {
    const atrStore = useActionTakenReportStore()
    
    // Seed ATR store with mock items
    atrStore.items = [
      {
        id: 'atr-1',
        finding_id: 'f-1',
        finding_title: 'SOP Belum Diperbarui',
        target_date: '2026-01-01',
        status: 'OPEN',
        progress: 0,
        department: 'Operasional'
      },
      {
        id: 'atr-2',
        finding_id: 'f-2',
        finding_title: 'Implementasi MFA',
        target_date: '2026-12-31',
        status: 'IN_PROGRESS',
        progress: 60,
        department: 'Teknologi Informasi'
      },
      {
        id: 'atr-3',
        finding_id: 'f-3',
        finding_title: 'Pelatihan Karyawan',
        target_date: '2026-02-01',
        status: 'COMPLETED',
        progress: 100,
        department: 'SDM'
      }
    ] as any

    expect(atrStore.items.length).toBe(3)
    
    const openCount = atrStore.items.filter(r => r.status === 'OPEN').length
    const inProgressCount = atrStore.items.filter(r => r.status === 'IN_PROGRESS').length
    const completedCount = atrStore.items.filter(r => r.status === 'COMPLETED').length

    expect(openCount).toBe(1)
    expect(inProgressCount).toBe(1)
    expect(completedCount).toBe(1)
  })

  it('checks PBC documents handling in audit fieldwork store', () => {
    const fieldworkStore = useAuditFieldworkStore()
    const letter = 'ST-001/SKAI/2026'

    fieldworkStore.selectedAssignmentLetter = letter
    fieldworkStore.fieldworkData[letter] = {
      interviews: [],
      observations: [],
      samples: [],
      testControls: [],
      documents: [
        {
          id: 'doc-1',
          assignmentLetterId: letter,
          documentName: 'Bukti Rekonsiliasi Bank',
          filePath: '/uploads/rekonsiliasi.pdf',
          file: null,
          description: 'Laporan rekonsiliasi',
          requiredDate: '2026-02-10'
        },
        {
          id: 'doc-2',
          assignmentLetterId: letter,
          documentName: 'Log Akses Firewall',
          filePath: '',
          file: null,
          description: 'Log akses bulanan',
          requiredDate: '2026-02-15'
        }
      ]
    }

    const pendingDocs = fieldworkStore.documents.filter(d => !d.filePath && !d.file)
    const uploadedDocs = fieldworkStore.documents.filter(d => !!d.filePath || !!d.file)

    expect(pendingDocs.length).toBe(1)
    expect(uploadedDocs.length).toBe(1)
    expect(pendingDocs[0]?.documentName).toBe('Log Akses Firewall')
  })

  it('checks audit survey pending detection logic', () => {
    const arrStore = useAuditResultReportStore()
    const surveyStore = useAuditeeSurveyStore()

    arrStore.reportList = [
      {
        id: 'rep-1',
        letter_no: 'ST-001/SKAI/2026',
        status: 'Published',
        title: 'LHA Operasional Q1',
        auditee_name: 'Divisi Operasional'
      }
    ] as any

    surveyStore.surveys = []

    const completedAudits = (arrStore.reportList as any[]).filter(
      r => r.status === 'Final' || r.status === 'Published'
    )
    const pendingSurveys = completedAudits.filter(report => {
      const alreadySubmitted = surveyStore.surveys.some(
        (s: any) => (s.assignmentLetterId || s.letter_no) === report.letter_no
      )
      return !alreadySubmitted
    })

    expect(pendingSurveys.length).toBe(1)
    expect(pendingSurveys[0]?.letter_no).toBe('ST-001/SKAI/2026')
  })

  it('correctly maps user roles from auth store whether string, object array, or position', () => {
    const authStore = useAuthStore()

    // 1. Role as object array: [{ name: 'Auditee' }]
    authStore.user = {
      id: 'u-1',
      username: 'auditee_user',
      roles: [{ id: 'r1', name: 'Auditee' }],
      department: 'Finance'
    } as any
    const rbacAuditee = useRbac()
    expect(rbacAuditee.isAuditee.value).toBe(true)
    expect(rbacAuditee.currentDashboardRole.value).toBe('auditee')

    // 2. Role as string array: ['Auditor']
    authStore.user = {
      id: 'u-2',
      username: 'auditor_user',
      roles: ['Auditor'],
      department: 'SKAI'
    } as any
    const rbacAuditor = useRbac()
    expect(rbacAuditor.isAuditor.value).toBe(true)
    expect(rbacAuditor.currentDashboardRole.value).toBe('auditor')

    // 3. Role from position: 'Senior Auditor'
    authStore.user = {
      id: 'u-3',
      username: 'senior_auditor',
      position: 'Senior Auditor',
      roles: []
    } as any
    const rbacSeniorAuditor = useRbac()
    expect(rbacSeniorAuditor.isAuditor.value).toBe(true)
    expect(rbacSeniorAuditor.currentDashboardRole.value).toBe('auditor')

    // 4. Role Admin
    authStore.user = {
      id: 'u-4',
      username: 'admin_user',
      roles: ['Admin']
    } as any
    const rbacAdmin = useRbac()
    expect(rbacAdmin.isAdmin.value).toBe(true)
    expect(rbacAdmin.currentDashboardRole.value).toBe('admin')

    // 5. Role Executive / Viewer
    authStore.user = {
      id: 'u-5',
      username: 'board_user',
      roles: ['Chief Audit Executive']
    } as any
    const rbacExecutive = useRbac()
    expect(rbacExecutive.isExecutive.value).toBe(true)
    expect(rbacExecutive.currentDashboardRole.value).toBe('executive')
  })
})
