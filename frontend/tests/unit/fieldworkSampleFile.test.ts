// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      auditServiceBaseUrl: 'http://localhost:8080/api/v1'
    }
  })
}))

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    success: vi.fn(),
    showSuccess: vi.fn(),
    error: vi.fn(),
    showError: vi.fn(),
    warning: vi.fn(),
    showWarning: vi.fn(),
    info: vi.fn(),
    showInfo: vi.fn()
  })
}))

global.$fetch = vi.fn()

describe('Audit Fieldwork Sample Data - File Upload, View, and Download', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    window.open = vi.fn()
    window.URL.createObjectURL = vi.fn(() => 'blob:http://localhost/mock-blob-id')
    window.URL.revokeObjectURL = vi.fn()
  })

  it('handleSampleFileChange should set file and fileName on sampleForm', () => {
    const store = useAuditFieldworkStore()
    const mockFile = new File(['sample content'], 'faktur_pembelian.pdf', { type: 'application/pdf' })
    const mockEvent = {
      target: {
        files: [mockFile]
      }
    } as any

    store.handleSampleFileChange(mockEvent)
    expect(store.sampleForm.file).toStrictEqual(mockFile)
    expect(store.sampleForm.fileName).toBe('faktur_pembelian.pdf')
  })

  it('resetSampleForm and openSampleModal should clear file fields', () => {
    const store = useAuditFieldworkStore()
    store.selectedAssignmentLetter = 'ST-001/SKAI/2026'
    store.sampleForm.file = new File([''], 'test.pdf')
    store.sampleForm.fileName = 'test.pdf'
    store.sampleForm.filePath = '/path/to/test.pdf'

    store.openSampleModal()
    expect(store.sampleForm.file).toBeNull()
    expect(store.sampleForm.fileName).toBe('')
    expect(store.sampleForm.filePath).toBe('')
    expect(store.sampleForm.fileUrl).toBe('')
  })

  it('editSample should populate file information if present', () => {
    const store = useAuditFieldworkStore()
    const item = {
      id: 'SAMP-001',
      assignmentLetterId: 'ST-001/SKAI/2026',
      documentName: 'Faktur Pengadaan',
      documentNumber: 'INV-2026-001',
      date: '2026-03-01',
      description: 'Pengujian faktur',
      fileName: 'faktur.pdf',
      filePath: 'Auditsphere/fieldwork/samples/faktur.pdf',
      fileUrl: 'http://localhost:8080/uploads/faktur.pdf'
    }

    store.editSample(item)
    expect(store.sampleForm.documentName).toBe('Faktur Pengadaan')
    expect(store.sampleForm.fileName).toBe('faktur.pdf')
    expect(store.sampleForm.filePath).toBe('Auditsphere/fieldwork/samples/faktur.pdf')
    expect(store.sampleForm.fileUrl).toBe('http://localhost:8080/uploads/faktur.pdf')
  })

  it('viewSampleFile should open file in new window', async () => {
    const store = useAuditFieldworkStore()

    // 1. Test with in-memory File
    const mockFile = new File(['hello'], 'document.pdf', { type: 'application/pdf' })
    await store.viewSampleFile({ file: mockFile, fileName: 'document.pdf' })
    expect(window.open).toHaveBeenCalledWith('blob:http://localhost/mock-blob-id', '_blank')

    // 2. Test with external URL
    await store.viewSampleFile({ fileName: 'file.pdf', fileUrl: 'https://example.com/file.pdf' })
    expect(window.open).toHaveBeenCalledWith('https://example.com/file.pdf', '_blank')
  })

  it('downloadSampleFile should trigger download via downloadInterviewFile handler', async () => {
    const store = useAuditFieldworkStore()
    const mockFile = new File(['content'], 'po.pdf', { type: 'application/pdf' })
    
    // In-memory file triggers anchor click
    const clickSpy = vi.fn()
    vi.spyOn(document, 'createElement').mockReturnValue({
      click: clickSpy,
      setAttribute: vi.fn()
    } as any)
    vi.spyOn(document.body, 'appendChild').mockImplementation(() => {})
    vi.spyOn(document.body, 'removeChild').mockImplementation(() => {})

    await store.downloadSampleFile({ file: mockFile, fileName: 'po.pdf' })
    expect(clickSpy).toHaveBeenCalled()
  })
})
