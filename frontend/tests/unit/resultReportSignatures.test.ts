// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      auditServiceBaseUrl: 'http://localhost:8002/api/v1'
    }
  })
}))

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn()
  })
}))

global.$fetch = vi.fn()

describe('Audit Result Report - Signatures Workflow', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.mocked($fetch).mockResolvedValue({ success: true, data: {} })
  })

  it('pulls Team Members from the selected Assignment Letter correctly', () => {
    const alStore = useAssignmentLetterStore()
    const arrStore = useAuditResultReportStore()

    // ST-001 has Zeta Ramadhani, Budi Santoso, Rina Wulandari, Andi Firmansyah, Dewi Kusumawati
    const membersST001 = arrStore.getTeamMembersForLetter('ST-001/SKAI/2026')
    expect(membersST001.length).toBe(5)
    expect(membersST001.map(m => m.name)).toContain('Zeta Ramadhani')
    expect(membersST001.map(m => m.name)).toContain('Budi Santoso')
    expect(membersST001.map(m => m.name)).toContain('Rina Wulandari')

    // ST-002 has Andi Firmansyah, Budi Santoso, Dedi Prasetyo
    const membersST002 = arrStore.getTeamMembersForLetter('ST-002/SKAI/2026')
    expect(membersST002.length).toBe(3)
    expect(membersST002.map(m => m.name)).toContain('Andi Firmansyah')
    expect(membersST002.map(m => m.name)).toContain('Dedi Prasetyo')
  })

  it('saves report with complete signature data (place, date, company, team member signatures)', async () => {
    const arrStore = useAuditResultReportStore()
    arrStore.reportForm.assignmentLetterId = 'ST-001/SKAI/2026'
    arrStore.reportForm.reportTitle = 'LHA Keuangan & Operasional'
    arrStore.reportForm.companyName = 'PT AIFL Indonesia'
    arrStore.reportForm.signaturePlace = 'Jakarta'
    arrStore.reportForm.signatureDate = '2026-10-09'
    arrStore.reportForm.signatures = [
      { name: 'Zeta Ramadhani', role: 'Chairperson', signature: 'data:image/png;base64,mockZeta' },
      { name: 'Budi Santoso', role: 'Supervisor', signature: 'data:image/png;base64,mockBudi' },
      { name: 'Rina Wulandari', role: 'Member', signature: 'data:image/png;base64,mockRina' }
    ]

    await arrStore.saveReport()

    const postCall = vi.mocked($fetch).mock.calls.find(c => c[1]?.method === 'POST')
    expect(postCall).toBeTruthy()
    const body = postCall![1].body

    expect(body.assignmentLetterId).toBe('ST-001/SKAI/2026')
    expect(body.signaturePlace).toBe('Jakarta')
    expect(body.signatureDate).toBe('2026-10-09')
    expect(body.companyName).toBe('PT AIFL Indonesia')
    expect(body.signatures).toHaveLength(3)
    expect(body.signatures[0]).toMatchObject({
      name: 'Zeta Ramadhani',
      role: 'Chairperson',
      signature: 'data:image/png;base64,mockZeta'
    })
  })

  it('populates and resets signature state cleanly', () => {
    const arrStore = useAuditResultReportStore()

    arrStore.editReport({
      id: 'report-123',
      reportNumber: 'LHA-001/SKAI/2026',
      assignmentLetterId: 'ST-001/SKAI/2026',
      reportTitle: 'Test Report',
      executiveSummary: 'Summary',
      reportDate: '2026-10-01',
      status: 'Final',
      findingsCount: 0,
      signaturePlace: 'Surabaya',
      signatureDate: '2026-10-05',
      companyName: 'PT Petrokimia',
      signatures: [
        { name: 'Budi Santoso', role: 'Supervisor', signature: 'data:image/png;base64,abc' }
      ]
    })

    expect(arrStore.reportForm.signaturePlace).toBe('Surabaya')
    expect(arrStore.reportForm.signatureDate).toBe('2026-10-05')
    expect(arrStore.reportForm.companyName).toBe('PT Petrokimia')
    expect(arrStore.reportForm.signatures).toHaveLength(1)

    arrStore.resetForm()
    expect(arrStore.reportForm.signaturePlace).toBe('Jakarta')
    expect(arrStore.reportForm.signatures).toEqual([])
  })
})
