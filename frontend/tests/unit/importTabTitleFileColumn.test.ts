// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useUploadStrategicPlanDocumentStore } from '~/stores/upload-strategic-plan-document'
import { useUploadPlanDocumentStore } from '~/stores/upload-plan-document'
import { useUploadAnnualPlanStore } from '~/stores/upload-annual-plan'
import { useUploadAssignmentLetterStore } from '~/stores/upload-assignment-letter'
import { useImportWorkingPaperStore } from '~/stores/import-working-paper'
import { useUploadAuditResultReportStore } from '~/stores/upload-audit-result-report'
import { useUploadExecutiveSummaryStore } from '~/stores/upload-executive-summary'
import { useUploadExecutiveSummaryReportStore } from '~/stores/upload-executive-summary-report'
import { useUploadConsultingDocumentStore } from '~/stores/upload-consulting-document'
import { useQualityAssuranceStore } from '~/stores/quality-assurance'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      auditServiceBaseUrl: 'http://localhost:8080/api/v1',
      masterServiceBaseUrl: 'http://localhost:8080/api/v1'
    }
  })
}))

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn()
  })
}))

global.$fetch = vi.fn()

describe('All Import Features - Browser Tab Title Uses File Name from File Column', () => {
  let mockTab: any

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()

    mockTab = {
      document: {
        title: '',
        open: vi.fn(),
        write: vi.fn(),
        close: vi.fn(),
        body: {
          style: {},
          innerHTML: ''
        }
      },
      addEventListener: vi.fn(),
      close: vi.fn()
    }

    window.open = vi.fn(() => mockTab)
    window.URL.createObjectURL = vi.fn(() => 'blob:http://localhost/mock-uuid-blob')
    window.URL.revokeObjectURL = vi.fn()
  })

  it('Strategic Audit Plan uses fileName instead of document title', async () => {
    const store = useUploadStrategicPlanDocumentStore()
    global.$fetch.mockResolvedValueOnce(new Blob(['pdf'], { type: 'application/pdf' }))

    await store.viewDocument('id-1', 'Renstra_Audit_2025.pdf', 'Document Title Renstra')

    expect(mockTab.document.title).toBe('Renstra_Audit_2025.pdf')
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining('<title>Renstra_Audit_2025.pdf</title>')
    )
  })

  it('Audit Activity Plan uses fileName instead of document title', async () => {
    const store = useUploadPlanDocumentStore()
    global.$fetch.mockResolvedValueOnce(new Blob(['pdf'], { type: 'application/pdf' }))

    await store.viewDocument('id-2', 'PKAT_Activity_Plan_2025.pdf', 'Document Title Activity Plan')

    expect(mockTab.document.title).toBe('PKAT_Activity_Plan_2025.pdf')
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining('<title>PKAT_Activity_Plan_2025.pdf</title>')
    )
  })

  it('Annual Audit Plan uses fileName instead of document title', async () => {
    const store = useUploadAnnualPlanStore()
    global.$fetch.mockResolvedValueOnce(new Blob(['pdf'], { type: 'application/pdf' }))

    await store.viewDocument('id-3', 'Annual_Plan_2025.pdf', 'Document Title Annual Plan')

    expect(mockTab.document.title).toBe('Annual_Plan_2025.pdf')
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining('<title>Annual_Plan_2025.pdf</title>')
    )
  })

  it('Assignment Letter uses fileName instead of document title', async () => {
    const store = useUploadAssignmentLetterStore()
    global.$fetch.mockResolvedValueOnce(new Blob(['pdf'], { type: 'application/pdf' }))

    await store.viewDocument('id-4', 'Surat_Tugas_2025.pdf', 'Document Title Surat Tugas')

    expect(mockTab.document.title).toBe('Surat_Tugas_2025.pdf')
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining('<title>Surat_Tugas_2025.pdf</title>')
    )
  })

  it('Working Paper uses fileName instead of document title', async () => {
    const store = useImportWorkingPaperStore()
    global.$fetch.mockResolvedValueOnce(new Blob(['pdf'], { type: 'application/pdf' }))

    await store.viewDocument('id-5', 'Kertas_Kerja_Audit.pdf', 'Document Title KKA')

    expect(mockTab.document.title).toBe('Kertas_Kerja_Audit.pdf')
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining('<title>Kertas_Kerja_Audit.pdf</title>')
    )
  })

  it('Audit Result Report (LHA) uses fileName instead of document title', async () => {
    const store = useUploadAuditResultReportStore()
    global.$fetch.mockResolvedValueOnce(new Blob(['pdf'], { type: 'application/pdf' }))

    await store.viewDocument('id-6', 'LHA_Final_2025.pdf', 'Document Title LHA')

    expect(mockTab.document.title).toBe('LHA_Final_2025.pdf')
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining('<title>LHA_Final_2025.pdf</title>')
    )
  })

  it('Executive Summary uses fileName instead of document title', async () => {
    const store = useUploadExecutiveSummaryStore()
    global.$fetch.mockResolvedValueOnce(new Blob(['pdf'], { type: 'application/pdf' }))

    await store.viewDocument('id-7', 'Executive_Summary_Doc.pdf', 'Document Title Executive Summary')

    expect(mockTab.document.title).toBe('Executive_Summary_Doc.pdf')
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining('<title>Executive_Summary_Doc.pdf</title>')
    )
  })

  it('Executive Summary Report uses fileName instead of document title', async () => {
    const store = useUploadExecutiveSummaryReportStore()
    global.$fetch.mockResolvedValueOnce(new Blob(['pdf'], { type: 'application/pdf' }))

    await store.viewDocument('id-8', 'Executive_Summary_Report.pdf', 'Document Title Report')

    expect(mockTab.document.title).toBe('Executive_Summary_Report.pdf')
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining('<title>Executive_Summary_Report.pdf</title>')
    )
  })

  it('Consulting Document uses fileName instead of document title', async () => {
    const store = useUploadConsultingDocumentStore()
    global.$fetch.mockResolvedValueOnce(new Blob(['pdf'], { type: 'application/pdf' }))

    await store.viewDocument('id-9', 'Consulting_Service_Report.pdf', 'Document Title Consulting')

    expect(mockTab.document.title).toBe('Consulting_Service_Report.pdf')
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining('<title>Consulting_Service_Report.pdf</title>')
    )
  })

  it('Quality Assurance uses fileName/attachment.name instead of assessment title', async () => {
    const store = useQualityAssuranceStore()
    global.$fetch.mockResolvedValueOnce(new Blob(['pdf'], { type: 'application/pdf' }))

    await store.viewDocument('4c22f5d1-4442-482e-9591-d4615eb18321', 'Report_External_QAR_2025.pdf', 'External QAR Assessment')

    expect(mockTab.document.title).toBe('Report_External_QAR_2025.pdf')
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining('<title>Report_External_QAR_2025.pdf</title>')
    )
  })
})
