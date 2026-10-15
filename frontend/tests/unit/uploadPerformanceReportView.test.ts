// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useUploadPerformanceReportStore } from '~/stores/upload-performance-report'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      auditServiceBaseUrl: 'http://localhost:8080/api/v1'
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

describe('Upload Performance Report Store - View Document Tab Title', () => {
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

  it('viewDocument calls the backend /view endpoint and sets browser tab title to uploaded document name', async () => {
    const store = useUploadPerformanceReportStore()
    const mockBlob = new Blob(['%PDF-1.4 mock content'], { type: 'application/pdf' })
    global.$fetch.mockResolvedValueOnce(mockBlob)

    const docId = 'rep-123'
    const fileName = 'Laporan_Kinerja_Audit_Q1_2026.pdf'
    const docTitle = 'Q1 2026 - Laporan Kinerja'

    await store.viewDocument(docId, fileName, docTitle)

    // 1. Should have opened the tab synchronously
    expect(window.open).toHaveBeenCalledWith('', '_blank')

    // 2. Should call backend view endpoint
    expect(global.$fetch).toHaveBeenCalledWith(
      expect.stringContaining(`/uploaded-performance-reports/${docId}/view`),
      expect.objectContaining({ responseType: 'blob' })
    )

    // 3. Tab title must be set to the uploaded document name (not a hash/UUID)
    expect(mockTab.document.title).toBe(fileName)

    // 4. Tab document should write HTML page containing the document title and iframe
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining(`<title>${fileName}</title>`)
    )
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining('<iframe src="blob:http://localhost/mock-uuid-blob"></iframe>')
    )
  })

  it('viewDocument uses fileName and does not use docTitle as tab title', async () => {
    const store = useUploadPerformanceReportStore()
    const mockBlob = new Blob(['%PDF-1.4 mock content'], { type: 'application/pdf' })
    global.$fetch.mockResolvedValueOnce(mockBlob)

    const docId = 'rep-456'
    const fileName = 'Actual_Uploaded_File.pdf'
    const docTitle = 'Annual Performance Report 2026'

    await store.viewDocument(docId, fileName, docTitle)

    expect(mockTab.document.title).toBe(fileName)
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining(`<title>${fileName}</title>`)
    )
  })

  it('viewDocument falls back to /download if /view returns error', async () => {
    const store = useUploadPerformanceReportStore()
    const mockBlob = new Blob(['%PDF-1.4 mock content'], { type: 'application/pdf' })
    global.$fetch
      .mockRejectedValueOnce(new Error('404 Not Found'))
      .mockResolvedValueOnce(mockBlob)

    const docId = 'rep-789'
    const fileName = 'Fallback_Doc.pdf'

    await store.viewDocument(docId, fileName)

    expect(global.$fetch).toHaveBeenCalledWith(
      expect.stringContaining(`/uploaded-performance-reports/${docId}/view`),
      expect.objectContaining({ responseType: 'blob' })
    )
    expect(global.$fetch).toHaveBeenCalledWith(
      expect.stringContaining(`/uploaded-performance-reports/${docId}/download`),
      expect.objectContaining({ responseType: 'blob' })
    )
    expect(mockTab.document.title).toBe(fileName)
    expect(mockTab.document.write).toHaveBeenCalledWith(
      expect.stringContaining(`<title>${fileName}</title>`)
    )
  })
})
