import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import * as employeeApiModule from '~/composables/useEmployeeApi'
import * as departmentApiModule from '~/composables/useDepartmentApi'
import type { Employee } from '~/types/master'
import type { AssignmentLetter } from '~/types/audit'

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn(),
    success: vi.fn(),
    error: vi.fn()
  })
}))

const FORM_SFC = resolve(__dirname, '../../components/assignment-letter/AssignmentLetterForm.vue')
const TABLE_SFC = resolve(__dirname, '../../components/assignment-letter/AssignmentLetterTable.vue')

const mockEmployees: Employee[] = [
  {
    id: 'emp-1',
    employee_code: 'EMP-001',
    full_name: 'Budi Hartanto',
    email: 'budi@example.com',
    phone: '08123456789',
    company_id: 'comp-1',
    department_id: 'dept-1',
    job_role_id: 'role-1',
    level_grade: 1,
    is_active: true,
    join_date: '2025-01-01',
    created_at: '2025-01-01',
    updated_at: '2025-01-01'
  },
  {
    id: 'emp-2',
    employee_code: 'EMP-002',
    full_name: 'Carolina Wijaya',
    email: 'carolina@example.com',
    phone: '08123456780',
    company_id: 'comp-1',
    department_id: 'dept-2',
    job_role_id: 'role-2',
    level_grade: 2,
    is_active: true,
    join_date: '2025-02-01',
    created_at: '2025-02-01',
    updated_at: '2025-02-01'
  },
  {
    id: 'emp-3',
    employee_code: 'EMP-003',
    full_name: 'Anton Hermawan',
    email: 'anton@example.com',
    phone: '08123456781',
    company_id: 'comp-1',
    department_id: 'dept-3',
    job_role_id: 'role-3',
    level_grade: 3,
    is_active: true,
    join_date: '2025-03-01',
    created_at: '2025-03-01',
    updated_at: '2025-03-01'
  },
  {
    id: 'emp-inactive',
    employee_code: 'EMP-999',
    full_name: 'Inactive Employee',
    email: 'inactive@example.com',
    phone: '08123456799',
    company_id: 'comp-1',
    department_id: 'dept-1',
    job_role_id: 'role-1',
    level_grade: 1,
    is_active: false,
    join_date: '2024-01-01',
    created_at: '2024-01-01',
    updated_at: '2024-01-01'
  }
]

describe('Assignment Letter - Team Members Synchronized with Master Employee Management', () => {
  let store: ReturnType<typeof useAssignmentLetterStore>
  let getAllEmployeesMock: Mock

  beforeEach(() => {
    setActivePinia(createPinia())
    getAllEmployeesMock = vi.fn().mockResolvedValue(mockEmployees)
    vi.spyOn(employeeApiModule, 'useEmployeeApi').mockReturnValue({
      getAllEmployees: getAllEmployeesMock,
      getEmployees: vi.fn(),
      getEmployeeById: vi.fn(),
      getNextEmployeeCode: vi.fn(),
      createEmployee: vi.fn(),
      updateEmployee: vi.fn(),
      deleteEmployee: vi.fn()
    })
    vi.spyOn(departmentApiModule, 'useDepartmentApi').mockReturnValue({
      getAllDepartments: vi.fn().mockResolvedValue([]),
      getDepartments: vi.fn(),
      getDepartmentById: vi.fn(),
      createDepartment: vi.fn(),
      updateDepartment: vi.fn(),
      deleteDepartment: vi.fn()
    })
    store = useAssignmentLetterStore()
  })

  it('fetches employees from Employee Management API and filters out inactive employees', async () => {
    await store.fetchEmployees()

    expect(getAllEmployeesMock).toHaveBeenCalled()
    expect(store.employees).toHaveLength(3)
    expect(store.employees.some(e => e.id === 'emp-inactive')).toBe(false)
    expect(store.employees[0]?.full_name).toBe('Anton Hermawan') // Sorted alphabetically
  })

  it('generates employeeOptions with employee full_name and employee_code', async () => {
    await store.fetchEmployees()

    expect(store.employeeOptions).toContain('Anton Hermawan (EMP-003)')
    expect(store.employeeOptions).toContain('Budi Hartanto (EMP-001)')
    expect(store.employeeOptions).toContain('Carolina Wijaya (EMP-002)')
  })

  it('filters out already selected team members from other rows so they do not appear again', async () => {
    await store.fetchEmployees()

    store.form.membersList = [
      { name: 'Budi Hartanto (EMP-001)', role: 'Chairperson' },
      { name: '', role: 'Member' }
    ]

    // Row 0's dropdown contains Budi (its current value)
    const row0Options = store.getAvailableMemberOptions(0)
    expect(row0Options).toContain('Budi Hartanto (EMP-001)')
    expect(row0Options).toContain('Anton Hermawan (EMP-003)')
    expect(row0Options).toContain('Carolina Wijaya (EMP-002)')

    // Row 1's dropdown must NOT contain Budi (already chosen in row 0)
    const row1Options = store.getAvailableMemberOptions(1)
    expect(row1Options).not.toContain('Budi Hartanto (EMP-001)')
    expect(row1Options).toContain('Anton Hermawan (EMP-003)')
    expect(row1Options).toContain('Carolina Wijaya (EMP-002)')

    // Now Row 1 selects Carolina
    store.form.membersList[1].name = 'Carolina Wijaya (EMP-002)'

    // Row 0 now excludes Carolina
    expect(store.getAvailableMemberOptions(0)).not.toContain('Carolina Wijaya (EMP-002)')
    expect(store.getAvailableMemberOptions(0)).toContain('Budi Hartanto (EMP-001)')

    // Row 1 excludes Budi
    expect(store.getAvailableMemberOptions(1)).not.toContain('Budi Hartanto (EMP-001)')
    expect(store.getAvailableMemberOptions(1)).toContain('Carolina Wijaya (EMP-002)')

    // Add Row 2: should exclude both Budi and Carolina
    store.form.membersList.push({ name: '', role: 'Member' })
    const row2Options = store.getAvailableMemberOptions(2)
    expect(row2Options).not.toContain('Budi Hartanto (EMP-001)')
    expect(row2Options).not.toContain('Carolina Wijaya (EMP-002)')
    expect(row2Options).toContain('Anton Hermawan (EMP-003)')
  })

  it('filters out members selected in other rows even if entered as raw name or code', async () => {
    await store.fetchEmployees()

    store.form.membersList = [
      { name: 'Budi Hartanto', role: 'Chairperson' }, // raw name without code
      { name: '', role: 'Member' }
    ]

    // Row 1 should filter out 'Budi Hartanto (EMP-001)'
    const row1Options = store.getAvailableMemberOptions(1)
    expect(row1Options).not.toContain('Budi Hartanto (EMP-001)')
    expect(row1Options).toContain('Anton Hermawan (EMP-003)')
    expect(row1Options).toContain('Carolina Wijaya (EMP-002)')
  })

  it('detects duplicate members with isDuplicateMember and hasDuplicateMembers', async () => {
    await store.fetchEmployees()

    store.form.membersList = [
      { name: 'Budi Hartanto (EMP-001)', role: 'Chairperson' },
      { name: 'Anton Hermawan (EMP-003)', role: 'Member' }
    ]

    expect(store.hasDuplicateMembers).toBe(false)
    expect(store.isDuplicateMember('Budi Hartanto (EMP-001)', 0)).toBe(false)
    expect(store.isDuplicateMember('Anton Hermawan (EMP-003)', 1)).toBe(false)

    // Simulate duplicate in row 1
    store.form.membersList[1].name = 'Budi Hartanto (EMP-001)'
    expect(store.hasDuplicateMembers).toBe(true)
    expect(store.isDuplicateMember('Budi Hartanto (EMP-001)', 0)).toBe(true)
    expect(store.isDuplicateMember('Budi Hartanto (EMP-001)', 1)).toBe(true)
  })

  it('blocks submit in handleSubmit when duplicate team members are present', async () => {
    await store.fetchEmployees()

    store.form.auditTitle = 'Audit Keuangan'
    store.form.category = 'Assurance' as any
    store.form.auditYear = '2026'
    store.form.auditTeam = 'SKAI'
    store.form.workingUnit = 'Finance'
    store.form.startPeriod = '2026-01-01'
    store.form.finishPeriod = '2026-01-31'
    store.form.membersList = [
      { name: 'Budi Hartanto (EMP-001)', role: 'Chairperson' },
      { name: 'Budi Hartanto (EMP-001)', role: 'Member' },
      { name: 'Anton Hermawan (EMP-003)', role: 'Member' }
    ]

    await store.handleSubmit()

    expect(store.errorMsg).toBe('Nama anggota tim tidak boleh sama (duplikat).')
  })

  it('preserves existing member values when opening form for edit', async () => {
    await store.fetchEmployees()

    const mockLetter: AssignmentLetter = {
      id: 'letter-1',
      letterNumber: 'ST-001/SKAI/2026',
      status: 'Draft',
      auditTitle: 'Audit Operasional',
      leader: 'Budi Hartanto',
      category: 'Assurance' as any,
      auditYear: '2026',
      auditTeam: 'SKAI',
      startPeriod: '2026-01-01',
      finishPeriod: '2026-01-31',
      workingUnit: 'Finance',
      auditPurpose: 'Purpose',
      executionPeriod: '2026-01-01 to 2026-01-31',
      createdAt: '2026-01-01',
      membersList: [
        { name: 'Budi Hartanto', role: 'Person in Charge' },
        { name: 'Carolina Wijaya', role: 'Member' }
      ],
      purposeList: ['Purpose 1'],
      scopeList: ['Scope 1'],
      ccList: ['President Director']
    }

    store.openEditModal(mockLetter)

    expect(store.editingId).toBe('letter-1')
    expect(store.isModalOpen).toBe(true)
    // Matches and resolves to Employee Management format with code
    expect(store.form.membersList[0]?.name).toBe('Budi Hartanto (EMP-001)')
    expect(store.form.membersList[1]?.name).toBe('Carolina Wijaya (EMP-002)')
  })

  it('preserves custom/unmatched member names when editing and includes them in employeeOptions', async () => {
    await store.fetchEmployees()

    const mockLetter: AssignmentLetter = {
      id: 'letter-custom',
      letterNumber: 'ST-002/SKAI/2026',
      status: 'Draft',
      auditTitle: 'Audit Khusus',
      leader: 'External Consultant',
      category: 'Special Audit' as any,
      auditYear: '2026',
      auditTeam: 'SKAI',
      startPeriod: '2026-02-01',
      finishPeriod: '2026-02-28',
      workingUnit: 'Operations',
      auditPurpose: 'Custom Purpose',
      executionPeriod: '2026-02-01 to 2026-02-28',
      createdAt: '2026-02-01',
      membersList: [
        { name: 'External Consultant', role: 'Chairperson' },
        { name: 'External Auditor', role: 'Member' }
      ],
      purposeList: ['Purpose 1'],
      scopeList: ['Scope 1'],
      ccList: ['President Director']
    }

    store.openEditModal(mockLetter)

    expect(store.form.membersList[0]?.name).toBe('External Consultant')
    expect(store.form.membersList[1]?.name).toBe('External Auditor')
    expect(store.employeeOptions).toContain('External Consultant')
    expect(store.employeeOptions).toContain('External Auditor')
  })

  it('resets membersList on openModal in create/add mode and fetches employees', async () => {
    store.openModal()

    expect(store.editingId).toBeNull()
    expect(store.isModalOpen).toBe(true)
    expect(store.form.membersList).toHaveLength(1)
    expect(store.form.membersList[0]?.name).toBe('')
    expect(store.form.membersList[0]?.role).toBe('Chairperson')
    expect(getAllEmployeesMock).toHaveBeenCalled()
  })

  it('formats employee name and leader correctly', async () => {
    await store.fetchEmployees()

    // By ID
    expect(store.formatEmployeeName('emp-1')).toBe('Budi Hartanto (EMP-001)')
    // By code
    expect(store.formatEmployeeName('EMP-002')).toBe('Carolina Wijaya (EMP-002)')
    // By plain full name matching employee
    expect(store.formatEmployeeName('Anton Hermawan')).toBe('Anton Hermawan (EMP-003)')
    // Already formatted string
    expect(store.formatEmployeeName('Anton Hermawan (EMP-003)')).toBe('Anton Hermawan (EMP-003)')
    // Unmatched string preserved
    expect(store.formatEmployeeName('External Person')).toBe('External Person')
    // Empty
    expect(store.formatEmployeeName('')).toBe('-')
    expect(store.formatEmployeeName(undefined)).toBe('-')

    // formatLeaderDisplay helper
    expect(store.formatLeaderDisplay({ leader: 'emp-1' } as any)).toBe('Budi Hartanto (EMP-001)')
    expect(store.formatLeaderDisplay({ leader: '', membersList: [{ name: 'emp-2', role: 'Person in Charge' }] } as any)).toBe('Carolina Wijaya (EMP-002)')
    expect(store.formatLeaderDisplay({ leader: '', membersList: [] } as any)).toBe('-')
  })

  it('AssignmentLetterForm.vue binds member.name to USelectMenu with getAvailableMemberOptions and shows duplicate validation', () => {
    const sfc = readFileSync(FORM_SFC, 'utf-8')

    // Binds USelectMenu to member.name with per-row available options
    expect(sfc).toContain(':items="store.getAvailableMemberOptions(index)"')
    expect(sfc).toContain('v-model="member.name"')
    expect(sfc).toContain(':loading="store.employeesLoading"')
    expect(sfc).toContain('create-item')
    expect(sfc).toContain('store.fetchEmployees')
    expect(sfc).toContain("t('masterData.employee.empty')")
    expect(sfc).toContain('store.isDuplicateMember(member.name, index)')
    expect(sfc).toContain("t('assignmentLetter.formExtra.memberDuplicate')")
  })

  it('AssignmentLetterTable.vue displays formatted leader using store.formatLeaderDisplay', () => {
    const sfc = readFileSync(TABLE_SFC, 'utf-8')

    expect(sfc).toContain('store.formatLeaderDisplay(row.original)')
    expect(sfc).toContain('store.fetchEmployees')
  })
})
