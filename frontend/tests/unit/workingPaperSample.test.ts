// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useWorkingPaperStore, sampleSchema } from '~/stores/working-paper'

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

describe('Working Paper Sample Form & Store - Button & Modal Actions', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('sampleSchema should successfully validate numeric values and coercible strings', () => {
    const validData1 = {
      population: 100,
      sampleSize: 10,
      conclusion: 'Effective control'
    }
    const result1 = sampleSchema.safeParse(validData1)
    expect(result1.success).toBe(true)

    const validData2 = {
      population: '100 Transaksi',
      sampleSize: '10',
      conclusion: 'Effective control'
    }
    const result2 = sampleSchema.safeParse(validData2)
    expect(result2.success).toBe(true)
    if (result2.success) {
      expect(result2.data.population).toBe('100 Transaksi')
      expect(result2.data.sampleSize).toBe(10)
    }

    const invalidData = {
      population: '',
      sampleSize: 0,
      conclusion: ''
    }
    const result3 = sampleSchema.safeParse(invalidData)
    expect(result3.success).toBe(false)
  })

  it('openModalF03 should open modal and initialize sampleForm with 1 clean sample item', () => {
    const store = useWorkingPaperStore()
    expect(store.showModalF03).toBe(false)

    store.openModalF03()
    expect(store.showModalF03).toBe(true)
    expect(store.isEditingF03).toBe(false)
    expect(store.sampleForm.samples.length).toBe(1)
    expect(store.sampleForm.conclusion).toBe('')
    expect(store.sampleForm.population).toBe('')
  })

  it('addSample and removeSample should correctly modify sample list', () => {
    const store = useWorkingPaperStore()
    store.openModalF03()
    expect(store.sampleForm.samples.length).toBe(1)

    store.addSample()
    expect(store.sampleForm.samples.length).toBe(2)

    store.removeSample(0)
    expect(store.sampleForm.samples.length).toBe(1)
  })

  it('checkSampleStatus should return correct boolean status and never crash on undefined', () => {
    const store = useWorkingPaperStore()
    expect(store.checkSampleStatus(undefined)).toBe(false)
    expect(store.checkSampleStatus({ l1: 'Pass', l2: 'Pass', l3: 'Pass' })).toBe(true)
    expect(store.checkSampleStatus({ l1: 'Pass', l2: 'Fail', l3: 'Pass' })).toBe(false)
  })

  it('handleEditF03 should populate sampleForm and support array or json-string samples and string population', () => {
    const store = useWorkingPaperStore()
    
    // Test with array
    const sampleItemWithArray = {
      id: 'WP-S-001',
      population: '100 Transaksi',
      sampleSize: 5,
      samples: [{ id: 1, document: 'DOC-1', l1: 'Pass', l2: 'Pass', l3: 'Pass' }],
      conclusion: 'All good'
    }
    store.handleEditF03(sampleItemWithArray)
    expect(store.showModalF03).toBe(true)
    expect(store.isEditingF03).toBe(true)
    expect(store.sampleForm.population).toBe('100 Transaksi')
    expect(store.sampleForm.samples.length).toBe(1)
    expect(store.sampleForm.samples[0].document).toBe('DOC-1')

    // Test with string JSON
    const sampleItemWithString = {
      id: 'WP-S-002',
      population: 80,
      sampleSize: 8,
      samples: JSON.stringify([{ id: 2, document: 'DOC-2', l1: 'Pass', l2: 'Fail', l3: 'Fail' }]),
      conclusion: 'Needs fix'
    }
    store.handleEditF03(sampleItemWithString)
    expect(store.sampleForm.samples.length).toBe(1)
    expect(store.sampleForm.samples[0].document).toBe('DOC-2')
  })

  it('closeModalF03 should close the modal', () => {
    const store = useWorkingPaperStore()
    store.openModalF03()
    expect(store.showModalF03).toBe(true)

    store.closeModalF03()
    expect(store.showModalF03).toBe(false)
  })

  it('fieldworkSampleOptions should return available options for documents', () => {
    const store = useWorkingPaperStore()
    expect(Array.isArray(store.fieldworkSampleOptions)).toBe(true)
    expect(store.fieldworkSampleOptions.length).toBeGreaterThan(0)
    expect(store.fieldworkSampleOptions[0]).toHaveProperty('label')
    expect(store.fieldworkSampleOptions[0]).toHaveProperty('value')
  })

  it('should correctly support fieldworkDocument and step1-3 in sample items', () => {
    const store = useWorkingPaperStore()
    store.openModalF03()
    expect(store.sampleForm.samples[0]).toMatchObject({
      fieldworkDocument: '',
      document: '',
      step1: '',
      step2: '',
      step3: ''
    })

    const sampleWithSteps = {
      id: 'WP-S-003',
      population: 10,
      sampleSize: 2,
      samples: [
        {
          id: 101,
          fieldworkDocument: 'Procurement Invoice (INV-2025-0988)',
          document: 'PO-2026-001',
          step1: 'Verifikasi kelengkapan tanda tangan',
          l1: 'Pass',
          step2: 'Kesesuaian jumlah unit barang',
          l2: 'Pass',
          step3: 'Pengecekan tanggal faktur pajak',
          l3: 'Fail'
        }
      ],
      conclusion: 'Step 3 failed'
    }

    store.handleEditF03(sampleWithSteps)
    const item = store.sampleForm.samples[0]
    expect(item.fieldworkDocument).toBe('Procurement Invoice (INV-2025-0988)')
    expect(item.document).toBe('PO-2026-001')
    expect(item.step1).toBe('Verifikasi kelengkapan tanda tangan')
    expect(item.l1).toBe('Pass')
    expect(item.step2).toBe('Kesesuaian jumlah unit barang')
    expect(item.l2).toBe('Pass')
    expect(item.step3).toBe('Pengecekan tanggal faktur pajak')
    expect(store.checkSampleStatus(item)).toBe(false)
  })

  it('addF03 should send boolean-compatible l1, l2, l3 in payload to prevent 400 Bad Request', async () => {
    const store = useWorkingPaperStore()
    const sampleFormData = {
      population: 100,
      sampleSize: 10,
      samples: [
        {
          id: 1,
          fieldworkDocument: 'Procurement Invoice (INV-2025-0988)',
          document: 'Contoh: INV/2026/03/001 - Faktur Pembelian PT Maju Mundur',
          step1: 'Verifikasi dokumen',
          l1: 'Pass',
          step2: 'Cek approval',
          l2: 'Fail',
          step3: 'Validasi nominal',
          l3: 'Pass'
        }
      ],
      conclusion: 'Sampel diuji'
    }
    await store.addF03(sampleFormData)
    expect(global.$fetch).toHaveBeenCalledWith(
      expect.stringContaining('/working-papers/samples'),
      expect.objectContaining({
        method: 'POST',
        body: expect.objectContaining({
          samples: [
            expect.objectContaining({
              l1: true,
              l2: false,
              l3: true
            })
          ]
        })
      })
    )
  })

  it('handleEditF03 should normalize boolean values from backend to Pass/Fail', () => {
    const store = useWorkingPaperStore()
    const backendData = {
      id: 'WP-S-004',
      population: 20,
      sampleSize: 2,
      samples: [
        {
          id: 201,
          fieldworkDocument: 'Bank Statement Reconciliation (BR-2025-12)',
          document: 'BR-001',
          step1: 'Step 1',
          l1: true,
          step2: 'Step 2',
          l2: false,
          step3: 'Step 3',
          l3: true
        }
      ],
      conclusion: 'Test boolean normalization'
    }

    store.handleEditF03(backendData)
    const item = store.sampleForm.samples[0]
    expect(item.l1).toBe('Pass')
    expect(item.l2).toBe('Fail')
    expect(item.l3).toBe('Pass')
  })
})
