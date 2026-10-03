import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useWorkingPaperStore } from '~/stores/working-paper'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'

describe('Working Paper Header - Assignment Letter Team Synchronization', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('automatically synchronizes Team Members and Audit Purpose when selecting ST-001/SKAI/2026', () => {
    const wpStore = useWorkingPaperStore()
    const letterStore = useAssignmentLetterStore()

    // Ensure ST-001 has distinct members in assignmentLetterList
    const st001 = letterStore.assignmentLetterList.find(l => l.letterNumber === 'ST-001/SKAI/2026')
    expect(st001).toBeDefined()
    expect(st001?.membersList.length).toBeGreaterThan(0)

    // Select ST-001/SKAI/2026
    wpStore.headerForm.assignmentLetterId = 'ST-001/SKAI/2026'
    wpStore.syncFromAssignmentLetter('ST-001/SKAI/2026')

    // Expect headerForm.teamMembers to match ST-001 members only
    expect(wpStore.headerForm.teamMembers.length).toBe(st001!.membersList.length)
    expect(wpStore.headerForm.teamMembers.map(m => m.name)).toEqual(st001!.membersList.map(m => m.name))
    expect(wpStore.headerForm.teamMembers.map(m => m.role)).toEqual(st001!.membersList.map(m => m.role))

    // Expect auditPurpose to match ST-001 audit purpose
    expect(wpStore.headerForm.auditPurpose).toBe(st001!.auditPurpose)
  })

  it('switches Team Members correctly when changing Assignment Letter to ST-002/SKAI/2026, without merging members', () => {
    const wpStore = useWorkingPaperStore()
    const letterStore = useAssignmentLetterStore()

    const st001 = letterStore.assignmentLetterList.find(l => l.letterNumber === 'ST-001/SKAI/2026')!
    const st002 = letterStore.assignmentLetterList.find(l => l.letterNumber === 'ST-002/SKAI/2026')!

    // First select ST-001
    wpStore.syncFromAssignmentLetter('ST-001/SKAI/2026')
    expect(wpStore.headerForm.teamMembers.length).toBe(st001.membersList.length)

    // Now change to ST-002
    wpStore.headerForm.assignmentLetterId = 'ST-002/SKAI/2026'
    wpStore.syncFromAssignmentLetter('ST-002/SKAI/2026')

    // Must strictly equal ST-002 members, NOT combined with ST-001
    expect(wpStore.headerForm.teamMembers.length).toBe(st002.membersList.length)
    expect(wpStore.headerForm.teamMembers.map(m => m.name)).toEqual(st002.membersList.map(m => m.name))
    expect(wpStore.headerForm.auditPurpose).toBe(st002.auditPurpose)
  })

  it('openModalF01 automatically populates Team Members from current selected assignment letter', () => {
    const fieldworkStore = useAuditFieldworkStore()
    fieldworkStore.selectedAssignmentLetter = 'ST-003/SKAI/2026'

    const wpStore = useWorkingPaperStore()
    const letterStore = useAssignmentLetterStore()
    const st003 = letterStore.assignmentLetterList.find(l => l.letterNumber === 'ST-003/SKAI/2026')!

    wpStore.openModalF01()

    expect(wpStore.headerForm.assignmentLetterId).toBe('ST-003/SKAI/2026')
    expect(wpStore.headerForm.teamMembers.length).toBe(st003.membersList.length)
    expect(wpStore.headerForm.teamMembers.map(m => m.name)).toEqual(st003.membersList.map(m => m.name))
  })

  it('clears Team Members if assignment letter is cleared', () => {
    const wpStore = useWorkingPaperStore()
    wpStore.syncFromAssignmentLetter('ST-001/SKAI/2026')
    expect(wpStore.headerForm.teamMembers.length).toBeGreaterThan(0)

    wpStore.syncFromAssignmentLetter('')
    expect(wpStore.headerForm.teamMembers).toEqual([])
    expect(wpStore.headerForm.auditPurpose).toBe('')
  })
})
