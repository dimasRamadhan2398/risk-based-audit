import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import * as departmentApiModule from '~/composables/useDepartmentApi'
import * as employeeApiModule from '~/composables/useEmployeeApi'
import type { Department } from '~/types/master'
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
const PAGE_SFC = resolve(__dirname, '../../pages/assignment-letter/index.vue')

const mockDepartments: Department[] = [
  {
    id: 'dept-1',
    department_code: 'FIN',
    department_name: 'Financial',
    level: 1,
    company_id: 'comp-1',
    is_active: true,
    created_at: '2025-01-01',
    updated_at: '2025-01-01'
  },
  {
    id: 'dept-2',
    department_code: 'OPS',
    department_name: 'Operasional',
    level: 1,
    company_id: 'comp-1',
    is_active: true,
    created_at: '2025-02-01',
    updated_at: '2025-02-01'
  },
  {
    id: 'dept-3',
    department_code: 'DEV',
    department_name: 'Product Development',
    level: 1,
    company_id: 'comp-1',
    is_active: true,
    created_at: '2025-03-01',
    updated_at: '2025-03-01'
  },
  {
    id: 'dept-inactive',
    department_code: 'OLD',
    department_name: 'Legacy Department',
    level: 1,
    company_id: 'comp-1',
    is_active: false,
    created_at: '2024-01-01',
    updated_at: '2024-01-01'
  }
]

describe('Assignment Letter - Work Unit Synchronized with Master Department Management', () => {
  let store: ReturnType<typeof useAssignmentLetterStore>
  let getAllDepartmentsMock: Mock

  beforeEach(() => {
    setActivePinia(createPinia())
    getAllDepartmentsMock = vi.fn().mockResolvedValue(mockDepartments)
    vi.spyOn(departmentApiModule, 'useDepartmentApi').mockReturnValue({
      getAllDepartments: getAllDepartmentsMock,
      getDepartments: vi.fn(),
      getDepartmentById: vi.fn(),
      createDepartment: vi.fn(),
      updateDepartment: vi.fn(),
      deleteDepartment: vi.fn()
    })
    vi.spyOn(employeeApiModule, 'useEmployeeApi').mockReturnValue({
      getAllEmployees: vi.fn().mockResolvedValue([]),
      getEmployees: vi.fn(),
      getEmployeeById: vi.fn(),
      getNextEmployeeCode: vi.fn(),
      createEmployee: vi.fn(),
      updateEmployee: vi.fn(),
      deleteEmployee: vi.fn()
    })
    store = useAssignmentLetterStore()
  })

  it('fetches departments from Department Management API and filters out inactive departments', async () => {
    await store.fetchDepartments()

    expect(getAllDepartmentsMock).toHaveBeenCalled()
    expect(store.departments).toHaveLength(3)
    expect(store.departments.some(d => d.id === 'dept-inactive')).toBe(false)
    expect(store.departments[0]?.department_name).toBe('Financial')
    expect(store.departments[1]?.department_name).toBe('Operasional')
    expect(store.departments[2]?.department_name).toBe('Product Development')
  })

  it('generates workUnitOptions with department_name and department_code sorted alphabetically', async () => {
    await store.fetchDepartments()

    expect(store.workUnitOptions).toEqual([
      'Financial (FIN)',
      'Operasional (OPS)',
      'Product Development (DEV)'
    ])
  })

  it('resolves department option from ID, code, name or formatted string', async () => {
    await store.fetchDepartments()

    expect(store.resolveDepartmentOption('dept-1')).toBe('Financial (FIN)')
    expect(store.resolveDepartmentOption('OPS')).toBe('Operasional (OPS)')
    expect(store.resolveDepartmentOption('Product Development')).toBe('Product Development (DEV)')
    expect(store.resolveDepartmentOption('Financial (FIN)')).toBe('Financial (FIN)')
    expect(store.resolveDepartmentOption('Custom Unit')).toBe('Custom Unit')
    expect(store.resolveDepartmentOption('')).toBe('')
    expect(store.resolveDepartmentOption(undefined)).toBe('')
  })

  it('formats department name for table display correctly', async () => {
    await store.fetchDepartments()

    expect(store.formatDepartmentName('dept-1')).toBe('Financial (FIN)')
    expect(store.formatDepartmentName('OPS')).toBe('Operasional (OPS)')
    expect(store.formatDepartmentName('Product Development')).toBe('Product Development (DEV)')
    expect(store.formatDepartmentName('Financial (FIN)')).toBe('Financial (FIN)')
    expect(store.formatDepartmentName('Custom Unit')).toBe('Custom Unit')
    expect(store.formatDepartmentName('')).toBe('-')
    expect(store.formatDepartmentName(undefined)).toBe('-')
  })

  it('preserves existing workingUnit and resolves it when opening edit modal', async () => {
    await store.fetchDepartments()

    const mockLetter: AssignmentLetter = {
      id: 'letter-1',
      letterNumber: 'ST-001/SKAI/2026',
      auditTitle: 'Audit Operasional',
      leader: 'Budi Hartanto',
      category: 'Assurance' as any,
      auditYear: '2026',
      auditTeam: 'SKAI',
      startPeriod: '2026-01-01',
      finishPeriod: '2026-01-31',
      workingUnit: 'Operasional', // Raw department name
      companyId: 'comp-1',
      companyName: 'Company 1',
      auditPurpose: 'Purpose',
      letterDate: '2026-01-01',
      caeSignature: 'CAE',
      status: 'Draft',
      membersList: [],
      purposeList: [],
      scopeList: [],
      ccList: []
    }

    store.openEditModal(mockLetter)

    expect(store.editingId).toBe('letter-1')
    expect(store.form.workingUnit).toBe('Operasional (OPS)')
  })

  it('preserves custom/unmatched workingUnit when editing and includes it in workUnitOptions', async () => {
    await store.fetchDepartments()

    const mockLetter: AssignmentLetter = {
      id: 'letter-custom',
      letterNumber: 'ST-002/SKAI/2026',
      auditTitle: 'Audit Ad-hoc',
      leader: 'Budi Hartanto',
      category: 'Assurance' as any,
      auditYear: '2026',
      auditTeam: 'SKAI',
      startPeriod: '2026-01-01',
      finishPeriod: '2026-01-31',
      workingUnit: 'Special Adhoc Unit',
      companyId: 'comp-1',
      companyName: 'Company 1',
      auditPurpose: 'Purpose',
      letterDate: '2026-01-01',
      caeSignature: 'CAE',
      status: 'Draft',
      membersList: [],
      purposeList: [],
      scopeList: [],
      ccList: []
    }

    store.openEditModal(mockLetter)

    expect(store.form.workingUnit).toBe('Special Adhoc Unit')
    expect(store.workUnitOptions).toContain('Special Adhoc Unit')
  })

  it('resets workingUnit on openModal in create mode', async () => {
    await store.fetchDepartments()
    store.form.workingUnit = 'Operasional (OPS)'

    store.openModal()

    expect(store.editingId).toBeNull()
    expect(store.form.workingUnit).toBe('')
  })

  it('deduplicates fetchDepartments calls to prevent redundant network requests', async () => {
    await Promise.all([
      store.fetchDepartments(),
      store.fetchDepartments(),
      store.fetchDepartments()
    ])

    expect(getAllDepartmentsMock).toHaveBeenCalledTimes(1)
    expect(store.departmentsLoaded).toBe(true)
    expect(store.departmentsLoading).toBe(false)

    // Force fetch should re-query
    await store.fetchDepartments(true)
    expect(getAllDepartmentsMock).toHaveBeenCalledTimes(2)
  })

  it('handles API errors gracefully when fetching departments', async () => {
    getAllDepartmentsMock.mockRejectedValueOnce(new Error('Network error'))

    await store.fetchDepartments()

    expect(store.departmentsLoaded).toBe(false)
    expect(store.departmentsLoading).toBe(false)
    expect(store.departmentsError).toBe('Network error')
  })

  it('AssignmentLetterForm.vue binds workingUnit to workUnitOptions with department loading and create-item', () => {
    const sfc = readFileSync(FORM_SFC, 'utf-8')

    // Work Unit field binding
    expect(sfc).toMatch(/<UFormField[^>]*?:label="t\('assignmentLetter\.formExtra\.workUnit'\)"/)
    expect(sfc).toContain(':items="store.workUnitOptions"')
    expect(sfc).toContain('v-model="store.form.workingUnit"')
    expect(sfc).toContain(':loading="store.departmentsLoading"')
    expect(sfc).toContain('create-item')
    expect(sfc).toContain('@create=')
    expect(sfc).toContain('masterData.department.empty')
  })

  it('AssignmentLetterTable.vue uses store.formatDepartmentName for workingUnit cell display', () => {
    const sfc = readFileSync(TABLE_SFC, 'utf-8')

    expect(sfc).toContain('store.formatDepartmentName(row.original.workingUnit)')
  })

  it('AssignmentLetterTable.vue and index.vue pre-fetch departments on mount', () => {
    const tableSfc = readFileSync(TABLE_SFC, 'utf-8')
    const pageSfc = readFileSync(PAGE_SFC, 'utf-8')

    expect(tableSfc).toContain('store.fetchDepartments()')
    expect(pageSfc).toContain('store.fetchDepartments()')
  })
})

