// @ts-nocheck
/**
 * Audit Fieldwork and Digital Working Paper (KKA) stores hold only data the API
 * returned: no sample records on an empty response, an error, a fresh letter
 * holder, or when nothing has been loaded yet.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { nextTick } from 'vue'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useWorkingPaperStore } from '~/stores/working-paper'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({ public: { auditServiceBaseUrl: 'http://localhost:8002/api/v1' } })
}))

// Both stores import the base URL from ~/composables/useApiUrl (not the Nuxt auto-import).
const API_BASE = 'http://localhost:8080/api/v1'
vi.mock('~/composables/useApiUrl', async importOriginal => ({
  ...(await importOriginal<typeof import('~/composables/useApiUrl')>()),
  getAuditServiceBaseUrl: () => API_BASE
}))

vi.mock('~/composables/useAppToast', () => ({
  useAppToast: () => ({ success: vi.fn(), error: vi.fn(), info: vi.fn(), warning: vi.fn() })
}))

const showError = vi.fn()
vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({ showSuccess: vi.fn(), showError, showWarning: vi.fn(), showInfo: vi.fn() })
}))

// A letter the old sample data was keyed by, so a mock leak would show up here.
const LETTER = 'ST-001/SKAI/2026'
const EMPTY = { interviews: [], observations: [], documents: [], samples: [], testControls: [] }
const flush = () => new Promise(r => setTimeout(r, 0))

const listResponse = (items: any[]) => ({ success: true, data: { items, pagination: { total: items.length } } })

describe('Audit Fieldwork store holds no mock data', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('does not expose a mock data set', () => {
    global.$fetch = vi.fn(async () => listResponse([]))
    const store = useAuditFieldworkStore()
    expect(store.mockFieldwork).toBeUndefined()
    expect(store.ensureDataExists).toBeUndefined()
  })

  it('an empty API response gives empty fieldwork data', async () => {
    global.$fetch = vi.fn(async () => listResponse([]))
    const store = useAuditFieldworkStore()
    await store.fetchAllFieldworkData(LETTER)
    expect(store.fieldworkData[LETTER]).toEqual(EMPTY)
    expect(store.errorMsg).toBe('')
  })

  it('an API error gives empty fieldwork data plus the error state and a toast', async () => {
    global.$fetch = vi.fn(async () => { throw new Error('Network error') })
    const store = useAuditFieldworkStore()
    await store.fetchAllFieldworkData(LETTER)
    expect(store.fieldworkData[LETTER]).toEqual(EMPTY)
    expect(store.errorMsg).toBeTruthy()
    expect(showError).toHaveBeenCalledTimes(1)
  })

  it('a partly failed load keeps the lists that loaded and still sets the error state and one toast', async () => {
    const tc = { id: 'tc1', assignmentLetterId: LETTER, controlName: 'Real control', testResult: 'Effective' }
    global.$fetch = vi.fn(async (url: string) => {
      if (String(url).includes('/interviews')) throw new Error('Interviews down')
      return listResponse(String(url).includes('/test-controls') ? [tc] : [])
    })
    const store = useAuditFieldworkStore()
    await store.fetchAllFieldworkData(LETTER)
    expect(store.fieldworkData[LETTER]).toEqual({ ...EMPTY, testControls: [tc] })
    expect(store.errorMsg).toBeTruthy()
    expect(showError).toHaveBeenCalledTimes(1)
  })

  it('selecting a letter (watch path) with the API down shows no records', async () => {
    global.$fetch = vi.fn(async () => { throw new Error('Network error') })
    const store = useAuditFieldworkStore()
    store.selectedAssignmentLetter = LETTER
    await nextTick()
    await flush()
    expect(store.interviews).toEqual([])
    expect(store.observations).toEqual([])
    expect(store.documents).toEqual([])
    expect(store.samples).toEqual([])
    expect(store.testControls).toEqual([])
  })

  it('shows exactly what the API returned', async () => {
    const tc = { id: 'tc1', assignmentLetterId: LETTER, controlName: 'Real control', testResult: 'Effective' }
    global.$fetch = vi.fn(async (url: string) => listResponse(String(url).includes('/test-controls') ? [tc] : []))
    const store = useAuditFieldworkStore()
    await store.fetchAllFieldworkData(LETTER)
    expect(store.fieldworkData[LETTER]).toEqual({ ...EMPTY, testControls: [tc] })
  })

  it('ensureFieldworkDataHolder creates empty structures and keeps existing data', () => {
    global.$fetch = vi.fn(async () => listResponse([]))
    const store = useAuditFieldworkStore()
    store.ensureFieldworkDataHolder(LETTER)
    expect(store.fieldworkData[LETTER]).toEqual(EMPTY)

    store.fieldworkData[LETTER].interviews.push({ id: 'i1' })
    store.ensureFieldworkDataHolder(LETTER)
    expect(store.fieldworkData[LETTER].interviews).toEqual([{ id: 'i1' }])
  })
})

describe('Digital Working Paper (KKA) store holds no mock data', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    global.$fetch = vi.fn(async () => listResponse([]))
  })

  const filtered = (wp: any) => [wp.filteredDataF01, wp.filteredDataF02, wp.filteredDataF03, wp.filteredDataF04, wp.filteredDataF05]

  it('does not expose mock F01–F05 data', () => {
    const wp = useWorkingPaperStore()
    for (const k of ['mockF01', 'mockF02', 'mockF03', 'mockF04', 'mockF05']) expect(wp[k]).toBeUndefined()
  })

  it('filtered F01–F05 are empty with no real data, with or without a selected letter', async () => {
    const wp = useWorkingPaperStore()
    filtered(wp).forEach(list => expect(list).toEqual([]))

    useAuditFieldworkStore().selectedAssignmentLetter = LETTER
    await nextTick()
    await flush()
    filtered(wp).forEach(list => expect(list).toEqual([]))
  })

  it('filtered F01–F05 are empty after an empty API load', async () => {
    const wp = useWorkingPaperStore()
    await wp.fetchAllData()
    expect(wp.errorMsg).toBe('')
    expect(vi.mocked($fetch)).toHaveBeenCalledWith(`${API_BASE}/working-papers/causes`, { method: 'GET' })
    filtered(wp).forEach(list => expect(list).toEqual([]))
  })

  it('fieldwork document options hold no placeholder documents', () => {
    const wp = useWorkingPaperStore()
    expect(wp.fieldworkSampleOptions).toEqual([])
    const src = readFileSync(resolve(__dirname, '../../stores/working-paper.ts'), 'utf8')
    expect(src).not.toMatch(/mockFieldwork/)
    expect(src).not.toMatch(/INV-2025-0988|BR-2025-12|LOG-ERP-2026-001|GI-2026-044/)
  })

  it('filtered data shows only real rows for the selected letter', async () => {
    const wp = useWorkingPaperStore()
    wp.dataF04 = [
      { id: 'C1', workingPaperId: LETTER, condition: 'Real cause' },
      { id: 'C2', workingPaperId: 'ST-OTHER', condition: 'Other letter' }
    ]
    useAuditFieldworkStore().selectedAssignmentLetter = LETTER
    await nextTick()
    await flush()
    expect(wp.filteredDataF04.map(c => c.id)).toEqual(['C1'])
  })

  it('filtered F01–F05 stay empty until an assignment letter is selected', async () => {
    const wp = useWorkingPaperStore()
    wp.dataF01 = [{ id: 'H1', assignmentLetterId: LETTER }]
    wp.dataF02 = [{ id: 'R1', workingPaperId: LETTER }]
    wp.dataF03 = [{ id: 'S1', workingPaperId: LETTER }]
    wp.dataF04 = [{ id: 'C1', workingPaperId: LETTER }]
    wp.dataF05 = [{ id: 'P1', workingPaperId: LETTER }]
    filtered(wp).forEach(list => expect(list).toEqual([]))

    useAuditFieldworkStore().selectedAssignmentLetter = LETTER
    await nextTick()
    await flush()
    filtered(wp).forEach(list => expect(list).toHaveLength(1))
  })
})
