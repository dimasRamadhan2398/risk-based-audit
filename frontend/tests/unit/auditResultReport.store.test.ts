// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useWorkingPaperStore } from '~/stores/working-paper'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'

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

// Real-shaped KKA / fieldwork records for one assignment letter, seeded into the
// working-paper and fieldwork stores as if they had been loaded from the API.
const LETTER = 'ST-001/SKAI/2026'

function seedAuditStores() {
  const wp = useWorkingPaperStore()
  wp.dataF02 = [{ id: 'R1', workingPaperId: LETTER, risk: 'Selisih kas', riskLevel: 'HIGH' }]
  wp.dataF04 = [{ id: 'C1', workingPaperId: LETTER, condition: 'Rekonsiliasi kas harian terlambat', criteria: 'SOP Kas', impact: 'Selisih kas' }]
  wp.dataF05 = [{ id: 'P1', workingPaperId: LETTER, recommendation: 'Rekonsiliasi harian', actionDescription: 'Otomatisasi rekonsiliasi' }]
  const fw = useAuditFieldworkStore()
  fw.fieldworkData = {
    [LETTER]: {
      interviews: [],
      observations: [],
      documents: [],
      samples: [],
      testControls: [
        { controlName: 'Otorisasi transaksi', testResult: 'Ineffective', finding: '', mitigationPlan: 'Dual approval' },
        { controlName: 'Backup harian', testResult: 'Effective', finding: '' }
      ]
    }
  }
}

describe('Audit Result Report Store - Automated Findings Detection', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    // Default: backend unreachable, so fetchAutoFindings aggregates from the loaded stores.
    vi.mocked($fetch).mockRejectedValue(new Error('Network error / fallback test'))
  })

  it('should auto-detect findings from loaded Digital Working Paper (KKA) and Fieldwork data', async () => {
    seedAuditStores()
    const store = useAuditResultReportStore()
    const findings = await store.fetchAutoFindings(LETTER)

    expect(findings.map(f => f.title)).toEqual([
      'Rekonsiliasi kas harian terlambat',
      'Kelemahan Kontrol: Otorisasi transaksi'
    ])
    expect(findings[0]).toMatchObject({ source: 'Digital Working Paper (KKA - AOI & RCA)', category: 'Very Significant', action: 'Otomatisasi rekonsiliasi' })
    expect(findings[1]).toMatchObject({ source: 'Audit Fieldwork (Test Controls)', category: 'Very Significant', action: 'Dual approval' })

    findings.forEach((f) => {
      expect(['Very Significant', 'Significant', 'Quite Significant', 'Not Significant']).toContain(f.category)
      expect(f.title).toBeTruthy()
    })
  })

  it('should not fall back to mock KKA / fieldwork data when nothing is loaded', async () => {
    const store = useAuditResultReportStore()
    const findings = await store.fetchAutoFindings(LETTER)
    expect(findings).toEqual([])
    // Auto-detection must not seed the fieldwork store with its mock records either.
    expect(useAuditFieldworkStore().fieldworkData).toEqual({})
  })

  it('should auto-populate findings and count when opening modal with a selected assignment letter', async () => {
    seedAuditStores()
    const store = useAuditResultReportStore()
    store.selectedAssignmentLetter = LETTER

    await store.openModal()

    expect(store.showModal).toBe(true)
    expect(store.reportForm.assignmentLetterId).toBe(LETTER)
    expect(store.reportForm.reportTitle).toContain('Laporan Hasil Audit')
    expect(store.reportForm.findings.length).toBe(2)
    expect(store.reportForm.findingsCount).toBe(store.reportForm.findings.length)
  })

  it('should support runAutoDetectFindings in replace and merge mode', async () => {
    seedAuditStores()
    const store = useAuditResultReportStore()
    store.reportForm.assignmentLetterId = LETTER
    store.reportForm.findings = [
      { title: 'Manual Custom Finding', category: 'Significant', action: 'Custom Action', source: 'Manual' }
    ]

    // Test merge mode
    await store.runAutoDetectFindings('merge')
    expect(store.reportForm.findings.length).toBe(3)
    expect(store.reportForm.findings.some(f => f.title === 'Manual Custom Finding')).toBe(true)

    // Test replace mode
    await store.runAutoDetectFindings('replace')
    expect(store.reportForm.findings.length).toBe(2)
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

describe('Audit Result Report Store - saveReport requires an assignment letter', () => {
  const REQUIRED_KEY = 'auditResultReport.form.assignmentLetterRequired'

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.mocked($fetch).mockResolvedValue({ success: true, data: { items: [] } })
  })

  it('blocks the save with a validation error and makes no API call when no letter is selected', async () => {
    const store = useAuditResultReportStore()
    await Promise.resolve()
    vi.mocked($fetch).mockClear() // ignore the fetchReports() fired on store creation

    store.selectedAssignmentLetter = ''
    store.reportForm.assignmentLetterId = ''
    store.reportForm.reportTitle = 'Laporan tanpa surat tugas'
    await store.saveReport()

    expect($fetch).not.toHaveBeenCalled()
    expect(store.formErrors.assignmentLetterId).toBe(REQUIRED_KEY)
  })

  it('has the validation message in both locales', async () => {
    const fs = await import('node:fs')
    for (const lang of ['en', 'id']) {
      const json = JSON.parse(fs.readFileSync(`locales/${lang}/common.json`, 'utf8'))
      expect(json.auditResultReport.form.assignmentLetterRequired).toBeTruthy()
    }
  })

  it('creates the report under the selected letter and clears the error', async () => {
    const store = useAuditResultReportStore()
    store.formErrors.assignmentLetterId = REQUIRED_KEY
    store.reportForm.assignmentLetterId = 'ST-010/SKAI/2026'
    await store.saveReport()

    const post = vi.mocked($fetch).mock.calls.find(c => c[1]?.method === 'POST')
    expect(post[1].body.assignmentLetterId).toBe('ST-010/SKAI/2026')
    expect(store.formErrors.assignmentLetterId).toBe('')
  })

  it('editing an existing report keeps its own letter, whatever the page filter is', async () => {
    const store = useAuditResultReportStore()
    store.selectedAssignmentLetter = 'ST-OTHER/SKAI/2026'
    store.editReport({
      id: 'r1',
      reportNumber: '',
      assignmentLetterId: 'ST-777/SKAI/2026',
      reportTitle: 'Existing',
      executiveSummary: '',
      reportDate: '2026-09-01',
      status: 'Draft',
      findingsCount: 0,
      findings: []
    })
    await store.saveReport()

    const put = vi.mocked($fetch).mock.calls.find(c => c[1]?.method === 'PUT')
    expect(String(put[0])).toContain('/audit-result-reports/r1')
    expect(put[1].body.assignmentLetterId).toBe('ST-777/SKAI/2026')
  })
})
