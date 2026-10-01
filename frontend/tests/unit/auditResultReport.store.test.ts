// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuditResultReportStore } from '~/stores/audit-result-report'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      auditServiceBaseUrl: 'http://localhost:8002/api/v1'
    }
  })
}))

vi.mock('#imports', () => ({
  useToast: () => ({
    add: vi.fn()
  })
}))

vi.mock('~/composables/useAppToast', () => ({
  useAppToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
    warning: vi.fn()
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

describe('Audit Result Report Store - Automated Findings Detection', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    // Default mock: fail fetch to trigger local fallback for deterministic testing
    vi.mocked($fetch).mockRejectedValue(new Error('Network error / fallback test'))
  })

  it('should auto-detect findings from Digital Working Paper (KKA) and Fieldwork for ST-001/SKAI/2026', async () => {
    const store = useAuditResultReportStore()
    const findings = await store.fetchAutoFindings('ST-001/SKAI/2026')

    expect(findings.length).toBeGreaterThan(0)
    const hasKkaOrFieldwork = findings.some(
      f => f.source?.includes('KKA') || f.source?.includes('Fieldwork') || f.source?.includes('Audit')
    )
    expect(hasKkaOrFieldwork).toBe(true)

    findings.forEach(f => {
      expect(['Very Significant', 'Significant', 'Quite Significant', 'Not Significant']).toContain(f.category)
      expect(f.title).toBeTruthy()
    })
  })

  it('should auto-populate findings and count when opening modal with a selected assignment letter', async () => {
    const store = useAuditResultReportStore()
    store.selectedAssignmentLetter = 'ST-001/SKAI/2026'

    await store.openModal()

    expect(store.showModal).toBe(true)
    expect(store.reportForm.assignmentLetterId).toBe('ST-001/SKAI/2026')
    expect(store.reportForm.reportTitle).toContain('Laporan Hasil Audit')
    expect(store.reportForm.findings.length).toBeGreaterThan(0)
    expect(store.reportForm.findingsCount).toBe(store.reportForm.findings.length)
  })

  it('should support runAutoDetectFindings in replace and merge mode', async () => {
    const store = useAuditResultReportStore()
    store.reportForm.assignmentLetterId = 'ST-001/SKAI/2026'
    store.reportForm.findings = [
      { title: 'Manual Custom Finding', category: 'Significant', action: 'Custom Action', source: 'Manual' }
    ]

    // Test merge mode
    await store.runAutoDetectFindings('merge')
    expect(store.reportForm.findings.length).toBeGreaterThan(1)
    expect(store.reportForm.findings.some(f => f.title === 'Manual Custom Finding')).toBe(true)

    // Test replace mode
    await store.runAutoDetectFindings('replace')
    expect(store.reportForm.findings.every(f => f.title !== 'Manual Custom Finding')).toBe(true)
  })

  it('should query backend auto-findings API when available', async () => {
    const store = useAuditResultReportStore()
    const mockBackendData = {
      success: true,
      data: {
        assignmentLetterId: 'ST-999',
        total: 1,
        findings: [
          {
            title: 'API Detected Ineffective Backup Control',
            category: 'Very Significant',
            action: 'Configure SMTP and Alerting',
            source: 'API Working Paper'
          }
        ]
      }
    }

    vi.mocked($fetch).mockResolvedValueOnce(mockBackendData)

    const findings = await store.fetchAutoFindings('ST-999')
    expect(findings.length).toBe(1)
    expect(findings[0].title).toBe('API Detected Ineffective Backup Control')
    expect(findings[0].category).toBe('Very Significant')
  })

  it('should dynamically generate Report Number (LHA ID) with auto-increment and dynamic month/year', async () => {
    const store = useAuditResultReportStore()
    
    // Test with specific date: 2026-10-15
    const dynamicNum = store.generateReportNumber('2026-10-15')
    expect(dynamicNum).toMatch(/^\d{3}\/LHA\/10\/KS IAD\/2026$/)

    // Test with different year and month: 2027-02-01
    const futureNum = store.generateReportNumber('2027-02-01')
    expect(futureNum).toMatch(/^\d{3}\/LHA\/02\/KS IAD\/2027$/)

    // Check that resetForm initializes reportForm.reportNumber dynamically
    store.resetForm()
    expect(store.reportForm.reportNumber).toBeTruthy()
    expect(store.reportForm.reportNumber).toMatch(/^\d{3}\/LHA\/\d{2}\/KS IAD\/\d{4}$/)
  })

  it('should dynamically generate Assignment Letter number with sanitized year and highest sequence', async () => {
    const { useAssignmentLetterStore } = await import('~/stores/assignment-letter')
    const alStore = useAssignmentLetterStore()

    // Test clean year
    const st1 = alStore.generateNomorSurat('SKAI', '2026')
    expect(st1).toMatch(/^ST-\d{3}\/SKAI\/2026$/)

    // Test full date format like 2026-05-20 from datepicker
    const st2 = alStore.generateNomorSurat('SKAI', '2026-05-20')
    expect(st2).toMatch(/^ST-\d{3}\/SKAI\/2026$/)

    // Test fallback to default SKAI and current year when empty
    const currentYear = new Date().getFullYear().toString()
    const stDefault = alStore.generateNomorSurat('', '')
    expect(stDefault).toBe(`ST-006/SKAI/${currentYear}`)
  })
})
