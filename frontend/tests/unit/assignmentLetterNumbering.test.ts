import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import type { AssignmentLetter } from '~/types/audit'

const makeLetter = (letterNumber: string, auditTeam = 'SKAI', auditYear = '2026'): AssignmentLetter => ({
  id: `id-${letterNumber}`,
  letterNumber,
  status: 'Draft',
  auditTitle: `Audit ${letterNumber}`,
  leader: 'Tester',
  category: 'Assurance' as any,
  auditYear,
  auditTeam,
  startPeriod: `${auditYear}-01-01`,
  finishPeriod: `${auditYear}-01-31`,
  workingUnit: 'Finance',
  executionPeriod: `${auditYear}-01-01 to ${auditYear}-01-31`,
  auditPurpose: 'Audit',
  letterDate: `${auditYear}-01-01`,
  caeSignature: '',
  membersList: [],
  purposeList: [],
  scopeList: [],
  ccList: []
})

describe('Assignment Letter - Sequential Letter Number Per Audit Team & Year', () => {
  let store: ReturnType<typeof useAssignmentLetterStore>

  beforeEach(() => {
    setActivePinia(createPinia())
    store = useAssignmentLetterStore()
    store.assignmentLetterList = []
  })

  it('starts from ST-001 independently for every audit team (SKAI, SPI, CAE, custom)', () => {
    // Initially all teams start at ST-001
    expect(store.generateNomorSurat('SKAI', '2026')).toBe('ST-001/SKAI/2026')
    expect(store.generateNomorSurat('SPI', '2026')).toBe('ST-001/SPI/2026')
    expect(store.generateNomorSurat('CAE', '2026')).toBe('ST-001/CAE/2026')
    expect(store.generateNomorSurat('CUSTOM', '2026')).toBe('ST-001/CUSTOM/2026')

    // After creating ST-001/SKAI/2026, SPI and CAE must STILL start at ST-001
    store.assignmentLetterList = [makeLetter('ST-001/SKAI/2026', 'SKAI', '2026')]

    expect(store.generateNomorSurat('SKAI', '2026')).toBe('ST-002/SKAI/2026')
    expect(store.generateNomorSurat('SPI', '2026')).toBe('ST-001/SPI/2026')
    expect(store.generateNomorSurat('CAE', '2026')).toBe('ST-001/CAE/2026')
  })

  it('keeps SKAI, SPI, and CAE sequences strictly independent as letters grow', () => {
    store.assignmentLetterList = [
      makeLetter('ST-001/SKAI/2026', 'SKAI', '2026'),
      makeLetter('ST-002/SKAI/2026', 'SKAI', '2026'),
      makeLetter('ST-003/SKAI/2026', 'SKAI', '2026'),
      makeLetter('ST-004/SKAI/2026', 'SKAI', '2026'),
      makeLetter('ST-001/SPI/2026', 'SPI', '2026'),
      makeLetter('ST-001/CAE/2026', 'CAE', '2026'),
      makeLetter('ST-002/CAE/2026', 'CAE', '2026')
    ]

    expect(store.generateNomorSurat('SKAI', '2026')).toBe('ST-005/SKAI/2026')
    expect(store.generateNomorSurat('SPI', '2026')).toBe('ST-002/SPI/2026')
    expect(store.generateNomorSurat('CAE', '2026')).toBe('ST-003/CAE/2026')
  })

  it('isolates sequences by year as well as team', () => {
    store.assignmentLetterList = [
      makeLetter('ST-001/SKAI/2025', 'SKAI', '2025'),
      makeLetter('ST-002/SKAI/2025', 'SKAI', '2025'),
      makeLetter('ST-001/SKAI/2026', 'SKAI', '2026')
    ]

    expect(store.generateNomorSurat('SKAI', '2025')).toBe('ST-003/SKAI/2025')
    expect(store.generateNomorSurat('SKAI', '2026')).toBe('ST-002/SKAI/2026')
    expect(store.generateNomorSurat('SKAI', '2027')).toBe('ST-001/SKAI/2027')
    expect(store.generateNomorSurat('SPI', '2025')).toBe('ST-001/SPI/2025')
  })

  it('does not jump to 20 or 9995 when outlier or legacy letter numbers exist', () => {
    store.assignmentLetterList = [
      makeLetter('ST-001/SKAI/2026', 'SKAI', '2026'),
      makeLetter('ST-002/SKAI/2026', 'SKAI', '2026'),
      makeLetter('ST-003/SKAI/2026', 'SKAI', '2026'),
      makeLetter('ST-004/SKAI/2026', 'SKAI', '2026'),
      makeLetter('ST-020/SKAI/2026', 'SKAI', '2026'),
      makeLetter('ST-9995/SKAI/2026', 'SKAI', '2026'),
      makeLetter('020/ST/01/KSIAD/2023', 'SKAI', '2023')
    ]

    // Must continue sequentially at 005 instead of 021 or 9996
    expect(store.generateNomorSurat('SKAI', '2026')).toBe('ST-005/SKAI/2026')

    // SPI must still start at 001 (not assignmentLetterList.length + 1 = 8!)
    expect(store.generateNomorSurat('SPI', '2026')).toBe('ST-001/SPI/2026')
  })

  it('handles lowercase, whitespace, and malformed entries gracefully', () => {
    store.assignmentLetterList = [
      makeLetter(' st-001 / spi / 2026 ', 'SPI', '2026'),
      makeLetter('020/ST/01/KSIAD/2023', 'SKAI', '2023')
    ]

    expect(store.generateNomorSurat(' spi ', '2026-06-15')).toBe('ST-002/SPI/2026')
    expect(store.generateNomorSurat('SKAI', '2026')).toBe('ST-001/SKAI/2026')
  })

  it('fetchAssignmentLetters fetches all pages when total_pages > 1', async () => {
    const fetchMock = vi.fn(async (url: string) => {
      if (url.includes('/assignment-letters?page=2')) {
        return {
          data: {
            items: [makeLetter('ST-002/SKAI/2026', 'SKAI', '2026')],
            pagination: { page: 2, page_size: 100, total: 2, total_pages: 2 }
          }
        }
      }
      if (url.includes('/assignment-letters')) {
        return {
          data: {
            items: [makeLetter('ST-001/SKAI/2026', 'SKAI', '2026')],
            pagination: { page: 1, page_size: 100, total: 2, total_pages: 2 }
          }
        }
      }
      return { data: [] }
    })

    vi.stubGlobal('$fetch', fetchMock)
    await store.fetchAssignmentLetters()

    const letterCalls = fetchMock.mock.calls.filter(([url]) => String(url).includes('/assignment-letters'))
    expect(letterCalls).toHaveLength(2)
    expect(store.assignmentLetterList).toHaveLength(2)
    expect(store.generateNomorSurat('SKAI', '2026')).toBe('ST-003/SKAI/2026')
  })
})
