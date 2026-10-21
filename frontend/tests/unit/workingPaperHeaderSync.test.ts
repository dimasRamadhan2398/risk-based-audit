import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { nextTick } from 'vue'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useWorkingPaperStore } from '~/stores/working-paper'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useLocationStore } from '~/stores/location'
import { useRiskProfileStore } from '~/stores/risk-profile'

describe('Working Paper Header - Assignment Letter Team Synchronization', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    const letterStore = useAssignmentLetterStore()
    letterStore.assignmentLetterList = [
      {
        id: 'mock-1',
        letterNumber: 'ST-001/SKAI/2026',
        auditPurpose: 'Annual Audit',
        membersList: [
          { name: 'Zeta Ramadhani', role: 'Chairperson' },
          { name: 'Budi Santoso', role: 'Supervisor' },
          { name: 'Rina Wulandari', role: 'Member' },
          { name: 'Andi Firmansyah', role: 'Member' },
          { name: 'Dewi Kusumawati', role: 'Person in Charge' }
        ]
      },
      {
        id: 'mock-2',
        letterNumber: 'ST-002/SKAI/2026',
        auditPurpose: 'IT Security Audit',
        membersList: [
          { name: 'Andi Firmansyah', role: 'Chairperson' },
          { name: 'Budi Santoso', role: 'Supervisor' },
          { name: 'Dedi Prasetyo', role: 'Member' }
        ]
      },
      {
        id: 'mock-3',
        letterNumber: 'ST-003/SKAI/2026',
        auditPurpose: 'Operational Audit',
        membersList: [
          { name: 'Rina Wulandari', role: 'Chairperson' },
          { name: 'Budi Santoso', role: 'Supervisor' }
        ]
      }
    ] as any
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
    const letterStore = useAssignmentLetterStore()
    letterStore.assignmentLetterList = [
      {
        id: 'mock-1',
        letterNumber: 'ST-001/SKAI/2026',
        auditPurpose: 'Annual Audit',
        membersList: [
          { name: 'Zeta Ramadhani', role: 'Chairperson' },
          { name: 'Budi Santoso', role: 'Supervisor' },
          { name: 'Rina Wulandari', role: 'Member' },
          { name: 'Andi Firmansyah', role: 'Member' },
          { name: 'Dewi Kusumawati', role: 'Person in Charge' }
        ]
      },
      {
        id: 'mock-2',
        letterNumber: 'ST-002/SKAI/2026',
        auditPurpose: 'IT Security Audit',
        membersList: [
          { name: 'Andi Firmansyah', role: 'Chairperson' },
          { name: 'Budi Santoso', role: 'Supervisor' },
          { name: 'Dedi Prasetyo', role: 'Member' }
        ]
      },
      {
        id: 'mock-3',
        letterNumber: 'ST-003/SKAI/2026',
        auditPurpose: 'Operational Audit',
        membersList: [
          { name: 'Rina Wulandari', role: 'Chairperson' },
          { name: 'Budi Santoso', role: 'Supervisor' }
        ]
      }
    ] as any
  })

  it('falls back to purposeList when the letter has no auditPurpose', () => {
    const letterStore = useAssignmentLetterStore()
    letterStore.assignmentLetterList.push(uiCreatedLetter() as any)
    const wpStore = useWorkingPaperStore()

    wpStore.syncFromAssignmentLetter('ST-006/SKAI/2026')

    expect(wpStore.headerForm.auditPurpose).toBe('Assess cash controls\nReview procurement approvals')
  })

  it('matches the letter by id as well as by letter number', () => {
    const letterStore = useAssignmentLetterStore()
    letterStore.assignmentLetterList.push(uiCreatedLetter() as any)
    const wpStore = useWorkingPaperStore()

    wpStore.syncFromAssignmentLetter('uuid-ui-6')

    expect(wpStore.headerForm.auditPurpose).toBe('Assess cash controls\nReview procurement approvals')
  })

  it('populates Audit Purpose when opening the create modal for a UI-created letter', () => {
    const letterStore = useAssignmentLetterStore()
    letterStore.assignmentLetterList.push(uiCreatedLetter() as any)
    const fieldworkStore = useAuditFieldworkStore()
    fieldworkStore.selectedAssignmentLetter = 'ST-006/SKAI/2026'
    const wpStore = useWorkingPaperStore()

    wpStore.openModalF01()

    expect(wpStore.headerForm.auditPurpose).toBe('Assess cash controls\nReview procurement approvals')
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

    expect(wpStore.headerForm.auditPurpose).toBe('Assess cash controls\nReview procurement approvals')
    expect(wpStore.headerForm.teamMembers.map(m => m.name)).toEqual(['Sari Dewi'])
  })
})

describe('Working Paper Header - Location Synchronization with Branch / Location Master Data', () => {
  const STORE_PATH = resolve(__dirname, '../../stores/working-paper.ts')

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('does not have hardcoded dummy locations in options.location', () => {
    const code = readFileSync(STORE_PATH, 'utf-8')
    expect(code).not.toContain("location: ['Head Office', 'Jakarta Branch', 'Bandung Branch']")
  })

  it('provides options.location based on active locations from Location master store', () => {
    const locStore = useLocationStore()
    locStore.locations = [
      { id: '1', name: 'AIFL Headquarters', is_active: true, address: 'Sudirman', city: 'Jakarta' },
      { id: '2', name: 'Surabaya Branch', is_active: true, address: 'A Yani', city: 'Surabaya' },
      { id: '3', name: 'Medan Branch (Closed)', is_active: false, address: 'Putri Hijau', city: 'Medan' }
    ] as any

    const wpStore = useWorkingPaperStore()
    expect(wpStore.options.location).toEqual(['AIFL Headquarters', 'Surabaya Branch'])
  })

  it('includes existing saved location in options.location even if not in active branch master', () => {
    const locStore = useLocationStore()
    locStore.locations = [
      { id: '1', name: 'AIFL Headquarters', is_active: true, address: 'Sudirman', city: 'Jakarta' }
    ] as any

    const wpStore = useWorkingPaperStore()
    wpStore.headerForm.location = 'Legacy Branch'

    expect(wpStore.options.location).toEqual(['Legacy Branch', 'AIFL Headquarters'])
  })

  it('fetches locations when openModalF01 is called and locations are empty', () => {
    const fieldworkStore = useAuditFieldworkStore()
    fieldworkStore.selectedAssignmentLetter = 'ST-001/SKAI/2026'

    const locStore = useLocationStore()
    const fetchSpy = vi.spyOn(locStore, 'fetchLocations').mockResolvedValue(undefined as any)

    const wpStore = useWorkingPaperStore()
    wpStore.openModalF01()

    expect(fetchSpy).toHaveBeenCalled()
  })

  it('fetches locations when handleEditF01 is called and locations are empty', () => {
    const locStore = useLocationStore()
    const fetchSpy = vi.spyOn(locStore, 'fetchLocations').mockResolvedValue(undefined as any)

    const wpStore = useWorkingPaperStore()
    wpStore.handleEditF01({
      id: 'wp-10',
      assignmentLetterId: 'ST-001/SKAI/2026',
      auditPurpose: 'Test Purpose',
      businessProcess: 'Procurement',
      period: '2026-01-01 s/d 2026-01-31',
      location: 'Surabaya Branch',
      teamMembers: [],
      activities: []
    })

    expect(fetchSpy).toHaveBeenCalled()
    expect(wpStore.headerForm.location).toBe('Surabaya Branch')
  })
})

describe('Working Paper Risk - Corporate Risk Profile Synchronization by Header Location', () => {
  const STORE_PATH = resolve(__dirname, '../../stores/working-paper.ts')

  const sampleCrpRisks = [
    {
      id: 'crp-1',
      name: 'Cyber security breach',
      category: 'Technology',
      impact: 5,
      likelihood: 5,
      branch: 'Head Office',
      location_id: 'loc-1'
    },
    {
      id: 'crp-2',
      name: 'Vendor delay',
      category: 'Operations',
      impact: 3,
      likelihood: 3,
      branch: 'Head Office',
      location_id: 'loc-1'
    },
    {
      id: 'crp-3',
      name: 'Fraud in Surabaya',
      category: 'Financial',
      impact: 4,
      likelihood: 5,
      branch: 'Surabaya Branch',
      location_id: 'loc-2'
    },
    {
      id: 'crp-4',
      name: 'Inventory Damage',
      category: 'Operations',
      impact: 1,
      likelihood: 2,
      branch: 'Surabaya Branch',
      location_id: 'loc-2'
    },
    {
      id: 'crp-5',
      name: 'General Policy Violation',
      category: 'Compliance',
      impact: 2,
      likelihood: 2,
      branch: 'Bali Branch',
      location_id: 'loc-3'
    }
  ]

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('does not have hardcoded dummy risk strings in options.risk', () => {
    const code = readFileSync(STORE_PATH, 'utf-8')
    expect(code).not.toContain('FIN-001: Fluktuasi Nilai Tukar Mata Uang')
    expect(code).not.toContain('OPS-001: Gangguan Sistem IT')
  })

  it('filters options.risk by the location specified in the header tab', () => {
    const fieldworkStore = useAuditFieldworkStore()
    fieldworkStore.selectedAssignmentLetter = 'ST-001/SKAI/2026'

    const riskProfileStore = useRiskProfileStore()
    riskProfileStore.rawRisks = [...sampleCrpRisks] as any

    const wpStore = useWorkingPaperStore()
    // Saved header with Surabaya Branch
    wpStore.dataF01 = [
      {
        id: 'h-1',
        assignmentLetterId: 'ST-001/SKAI/2026',
        businessProcess: 'Warehouse',
        period: '2026-01-01 s/d 2026-01-31',
        location: 'Surabaya Branch',
        teamMembers: [],
        activities: []
      }
    ] as any

    expect(wpStore.activeHeaderLocation).toBe('Surabaya Branch')
    expect(wpStore.options.risk).toEqual(['Fraud in Surabaya', 'Inventory Damage'])
  })

  it('updates options.risk dynamically when header location changes', () => {
    const fieldworkStore = useAuditFieldworkStore()
    fieldworkStore.selectedAssignmentLetter = 'ST-001/SKAI/2026'

    const riskProfileStore = useRiskProfileStore()
    riskProfileStore.rawRisks = [...sampleCrpRisks] as any

    const wpStore = useWorkingPaperStore()
    wpStore.dataF01 = [
      {
        id: 'h-1',
        assignmentLetterId: 'ST-001/SKAI/2026',
        businessProcess: 'IT',
        period: '2026-01-01 s/d 2026-01-31',
        location: 'Head Office',
        teamMembers: [],
        activities: []
      }
    ] as any

    expect(wpStore.options.risk).toEqual(['Cyber security breach', 'Vendor delay'])

    // Change location to Surabaya Branch in headerForm
    wpStore.headerForm.location = 'Surabaya Branch'
    wpStore.dataF01[0].location = 'Surabaya Branch'

    expect(wpStore.options.risk).toEqual(['Fraud in Surabaya', 'Inventory Damage'])
  })

  it('automatically populates Risk Category and Risk Level when selecting a risk', async () => {
    const riskProfileStore = useRiskProfileStore()
    riskProfileStore.rawRisks = [...sampleCrpRisks] as any

    const wpStore = useWorkingPaperStore()
    wpStore.openModalF02()

    expect(wpStore.riskForm.taxonomy).toBe('')
    expect(wpStore.riskForm.riskLevel).toBe('')

    // Pick 'Cyber security breach' (impact 5, likelihood 5 -> High, Technology)
    wpStore.riskForm.risk = 'Cyber security breach'
    await nextTick()

    expect(wpStore.riskForm.taxonomy).toBe('Technology')
    expect(wpStore.riskForm.riskLevel).toBe('High')

    // Switch to 'Vendor delay' (impact 3, likelihood 3 -> Moderate, Operations)
    wpStore.riskForm.risk = 'Vendor delay'
    await nextTick()

    expect(wpStore.riskForm.taxonomy).toBe('Operations')
    expect(wpStore.riskForm.riskLevel).toBe('Moderate')

    // Switch to 'Inventory Damage' (impact 1, likelihood 2 -> Low, Operations)
    wpStore.riskForm.risk = 'Inventory Damage'
    await nextTick()

    expect(wpStore.riskForm.taxonomy).toBe('Operations')
    expect(wpStore.riskForm.riskLevel).toBe('Low')
  })

  it('keeps saved risk in options.risk when editing a risk even if not matching current header location', () => {
    const fieldworkStore = useAuditFieldworkStore()
    fieldworkStore.selectedAssignmentLetter = 'ST-001/SKAI/2026'

    const riskProfileStore = useRiskProfileStore()
    riskProfileStore.rawRisks = [...sampleCrpRisks] as any

    const wpStore = useWorkingPaperStore()
    wpStore.dataF01 = [
      {
        id: 'h-1',
        assignmentLetterId: 'ST-001/SKAI/2026',
        location: 'Surabaya Branch'
      }
    ] as any

    // Edit an existing risk that was from Head Office
    wpStore.handleEditF02({
      id: 'r-saved-1',
      risk: 'Cyber security breach',
      taxonomy: 'Technology',
      riskLevel: 'High',
      controlDescription: 'Firewall rules'
    })

    expect(wpStore.options.risk).toContain('Cyber security breach')
    expect(wpStore.options.risk).toContain('Fraud in Surabaya')
    expect(wpStore.options.risk).toContain('Inventory Damage')
  })

  it('triggers fetchRisks and fetchBranches on openModalF02 when rawRisks is empty', () => {
    const riskProfileStore = useRiskProfileStore()
    riskProfileStore.rawRisks = []
    const fetchRisksSpy = vi.spyOn(riskProfileStore, 'fetchRisks').mockResolvedValue(undefined as any)
    const fetchBranchesSpy = vi.spyOn(riskProfileStore, 'fetchBranches').mockResolvedValue(undefined as any)

    const wpStore = useWorkingPaperStore()
    wpStore.openModalF02()

    expect(fetchRisksSpy).toHaveBeenCalled()
    expect(fetchBranchesSpy).toHaveBeenCalled()
  })
})


