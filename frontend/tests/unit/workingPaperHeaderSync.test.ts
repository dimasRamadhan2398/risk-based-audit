import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { nextTick } from 'vue'
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

describe('Working Paper Header - Audit Purpose from Assignment Letter', () => {
  // Letters created in the UI only fill purposeList; auditPurpose stays empty.
  const uiCreatedLetter = (overrides: Record<string, any> = {}) => ({
    id: 'uuid-ui-6',
    letterNumber: 'ST-006/SKAI/2026',
    status: 'Published',
    auditPurpose: '',
    purposeList: ['Assess cash controls', '  ', 'Review procurement approvals'],
    membersList: [{ name: 'Sari Dewi', role: 'Chairperson' }],
    scopeList: [],
    ccList: [],
    ...overrides
  })

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('falls back to purposeList when the letter has no auditPurpose', () => {
    const letterStore = useAssignmentLetterStore()
    letterStore.assignmentLetterList.push(uiCreatedLetter() as any)
    const wpStore = useWorkingPaperStore()

    wpStore.syncFromAssignmentLetter('ST-006/SKAI/2026')

    expect(wpStore.headerForm.auditPurpose).toBe('Assess cash controls; Review procurement approvals')
  })

  it('matches the letter by id as well as by letter number', () => {
    const letterStore = useAssignmentLetterStore()
    letterStore.assignmentLetterList.push(uiCreatedLetter() as any)
    const wpStore = useWorkingPaperStore()

    wpStore.syncFromAssignmentLetter('uuid-ui-6')

    expect(wpStore.headerForm.auditPurpose).toBe('Assess cash controls; Review procurement approvals')
  })

  it('populates Audit Purpose when opening the create modal for a UI-created letter', () => {
    const letterStore = useAssignmentLetterStore()
    letterStore.assignmentLetterList.push(uiCreatedLetter() as any)
    const fieldworkStore = useAuditFieldworkStore()
    fieldworkStore.selectedAssignmentLetter = 'ST-006/SKAI/2026'
    const wpStore = useWorkingPaperStore()

    wpStore.openModalF01()

    expect(wpStore.headerForm.auditPurpose).toBe('Assess cash controls; Review procurement approvals')
    expect(wpStore.headerForm.teamMembers.map(m => m.name)).toEqual(['Sari Dewi'])
  })

  it('does not keep the previous letter purpose when switching to a letter without one', () => {
    const letterStore = useAssignmentLetterStore()
    letterStore.assignmentLetterList.push(uiCreatedLetter({ purposeList: [] }) as any)
    const wpStore = useWorkingPaperStore()

    wpStore.headerForm.assignmentLetterId = 'ST-001/SKAI/2026'
    expect(wpStore.headerForm.auditPurpose).toBe('Annual Audit')

    wpStore.headerForm.assignmentLetterId = 'ST-006/SKAI/2026'
    expect(wpStore.headerForm.auditPurpose).toBe('')
  })

  it('updates Audit Purpose when the assignment letter changes on an open form', () => {
    const wpStore = useWorkingPaperStore()

    wpStore.headerForm.assignmentLetterId = 'ST-001/SKAI/2026'
    wpStore.headerForm.assignmentLetterId = 'ST-002/SKAI/2026'

    expect(wpStore.headerForm.auditPurpose).toBe('IT Security Audit')
  })

  it('editing a saved working paper with an empty auditPurpose falls back to its letter', async () => {
    const wpStore = useWorkingPaperStore()

    wpStore.handleEditF01({
      id: 'wp-1',
      assignmentLetterId: 'ST-002/SKAI/2026',
      auditPurpose: '',
      businessProcess: 'Access review',
      period: '2026-04-01 s/d 2026-04-30',
      location: 'Jakarta',
      teamMembers: [{ id: 1, name: 'Saved Member', role: 'Member' }],
      activities: []
    })
    await nextTick()

    expect(wpStore.headerForm.auditPurpose).toBe('IT Security Audit')
    // Saved team members are kept, not replaced by the letter's
    expect(wpStore.headerForm.teamMembers.map(m => m.name)).toEqual(['Saved Member'])
  })

  it('editing a saved working paper keeps its stored auditPurpose', async () => {
    const wpStore = useWorkingPaperStore()
    // Form previously held another letter, so the letter watcher fires on edit
    wpStore.headerForm.assignmentLetterId = 'ST-001/SKAI/2026'

    wpStore.handleEditF01({
      id: 'wp-2',
      assignmentLetterId: 'ST-002/SKAI/2026',
      auditPurpose: 'Stored purpose',
      businessProcess: 'Access review',
      period: '2026-04-01 s/d 2026-04-30',
      location: 'Jakarta',
      teamMembers: [],
      activities: []
    })
    await nextTick()

    expect(wpStore.headerForm.auditPurpose).toBe('Stored purpose')
  })

  it('fills Audit Purpose once the letter list loads after the modal opened', async () => {
    const fieldworkStore = useAuditFieldworkStore()
    fieldworkStore.selectedAssignmentLetter = 'ST-006/SKAI/2026'
    const letterStore = useAssignmentLetterStore()
    const wpStore = useWorkingPaperStore()

    wpStore.openModalF01()
    expect(wpStore.headerForm.auditPurpose).toBe('')

    letterStore.assignmentLetterList = [uiCreatedLetter() as any]
    await nextTick()

    expect(wpStore.headerForm.auditPurpose).toBe('Assess cash controls; Review procurement approvals')
    expect(wpStore.headerForm.teamMembers.map(m => m.name)).toEqual(['Sari Dewi'])
  })
})
