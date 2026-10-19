import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useWorkingPaperStore } from '~/stores/working-paper'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'

describe('Auditor Dashboard Store Integration & Navigation Safety', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('exposes getLocalAutoFindings function on useAuditResultReportStore', () => {
    const arrStore = useAuditResultReportStore()
    expect(typeof arrStore.getLocalAutoFindings).toBe('function')
  })

  it('getLocalAutoFindings returns findings safely without throwing', () => {
    const arrStore = useAuditResultReportStore()
    const wpStore = useWorkingPaperStore()
    const fwStore = useAuditFieldworkStore()

    // Seed dummy data in working paper store
    wpStore.dataF04 = [
      {
        id: 'c-1',
        workingPaperId: 'ST-001/SKAI/2026',
        condition: 'Sistem otentikasi belum memiliki multi-factor authentication (MFA)',
        criteria: 'Kebijakan Keamanan Sistem Informasi',
        impact: 'Risiko akses tidak sah'
      }
    ] as any

    const findings = arrStore.getLocalAutoFindings('ST-001/SKAI/2026')
    expect(Array.isArray(findings)).toBe(true)
    expect(findings.length).toBeGreaterThan(0)
    expect(findings[0].title).toContain('multi-factor authentication')
  })

  it('handles empty or non-existent assignment letters gracefully without errors', () => {
    const arrStore = useAuditResultReportStore()
    expect(() => arrStore.getLocalAutoFindings('')).not.toThrow()
    expect(arrStore.getLocalAutoFindings('')).toEqual([])
    expect(() => arrStore.getLocalAutoFindings('NON-EXISTENT')).not.toThrow()
    expect(arrStore.getLocalAutoFindings('NON-EXISTENT')).toEqual([])
  })

  it('safely updates selected assignment letter between fieldwork and working paper stores', () => {
    const fieldworkStore = useAuditFieldworkStore()
    const workingPaperStore = useWorkingPaperStore()

    const letter = 'ST-999/SKAI/2026'
    fieldworkStore.selectedAssignmentLetter = letter
    workingPaperStore.selectedAssignmentLetter = letter

    expect(fieldworkStore.selectedAssignmentLetter).toBe(letter)
    expect(workingPaperStore.hasAssignmentLetter).toBe(true)
  })
})
