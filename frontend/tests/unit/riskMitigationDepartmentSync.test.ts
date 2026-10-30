import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useMitigationStore } from '~/stores/mitigation-risk'
import * as departmentApiModule from '~/composables/useDepartmentApi'
import * as employeeApiModule from '~/composables/useEmployeeApi'
import type { Department } from '~/types/master'

const FORM_SFC = resolve(__dirname, '../../components/risk-mitigation/MitigationForm.vue')
const TABLE_SFC = resolve(__dirname, '../../components/risk-mitigation/MitigationTable.vue')

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

describe('Risk Mitigation - Division in Charge Synchronized with Master Department Management', () => {
  let store: ReturnType<typeof useMitigationStore>
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
    store = useMitigationStore()
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

  it('generates unitInChargeOptions with department_name and department_code sorted alphabetically', async () => {
    await store.fetchDepartments()

    expect(store.unitInChargeOptions).toEqual([
      'Financial (FIN)',
      'Operasional (OPS)',
      'Product Development (DEV)'
    ])
  })

  it('preserves existing unitInCharge values and resolves them when opening form for edit', async () => {
    await store.fetchDepartments()

    // Open form with department name
    store.openForm({
      id: 'mit-1',
      riskId: 'risk-100',
      riskControlId: 'CTL-RSK-001',
      riskEvent: 'Event 1',
      mitigationPlan: 'Plan 1',
      supervisor: '',
      pic: '',
      unitInCharge: 'Operasional',
      start_date: '2026-01-01T00:00:00Z',
      end_date: '2026-12-31T00:00:00Z',
      notes: ''
    })

    expect(store.isEditing).toBe(true)
    expect(store.form.unitInCharge).toBe('Operasional (OPS)')
  })

  it('preserves custom/unmatched unitInCharge when editing and includes it in options', async () => {
    await store.fetchDepartments()

    store.openForm({
      id: 'mit-custom',
      riskId: 'risk-100',
      riskControlId: 'CTL-RSK-002',
      riskEvent: 'Event 2',
      mitigationPlan: 'Plan 2',
      supervisor: '',
      pic: '',
      unitInCharge: 'External Taskforce',
      start_date: '2026-01-01T00:00:00Z',
      end_date: '2026-12-31T00:00:00Z',
      notes: ''
    })

    expect(store.form.unitInCharge).toBe('External Taskforce')
    expect(store.unitInChargeOptions).toContain('External Taskforce')
  })

  it('resets unitInCharge on openForm in create/add mode', async () => {
    await store.fetchDepartments()

    store.openForm()

    expect(store.isEditing).toBe(false)
    expect(store.form.unitInCharge).toBe('')
  })

  it('formats department name in table correctly', async () => {
    await store.fetchDepartments()

    // By ID
    expect(store.formatDepartmentName('dept-1')).toBe('Financial (FIN)')
    // By code
    expect(store.formatDepartmentName('OPS')).toBe('Operasional (OPS)')
    // By name
    expect(store.formatDepartmentName('Product Development')).toBe('Product Development (DEV)')
    // By already formatted string
    expect(store.formatDepartmentName('Financial (FIN)')).toBe('Financial (FIN)')
    // Empty
    expect(store.formatDepartmentName('')).toBe('-')
    expect(store.formatDepartmentName(undefined)).toBe('-')
  })

  it('MitigationForm.vue binds unitInCharge to store options with department loading state', () => {
    const sfc = readFileSync(FORM_SFC, 'utf-8')

    // Division in Charge field
    expect(sfc).toMatch(/<UFormField[^>]*?:label="t\('riskMitigation\.form\.unit'\)"/)
    expect(sfc).toContain(':items="store.unitInChargeOptions"')
    expect(sfc).toContain('v-model="store.form.unitInCharge"')

    // Loading states & empty templates
    expect(sfc).toContain(':loading="store.departmentsLoading"')
    expect(sfc).toContain('masterData.department.empty')
  })

  it('MitigationTable.vue uses store.formatDepartmentName for unitInCharge display', () => {
    const sfc = readFileSync(TABLE_SFC, 'utf-8')

    expect(sfc).toContain('store.formatDepartmentName(row.unitInCharge)')
  })

  it('deduplicates fetchDepartments calls to prevent redundant requests', async () => {
    await Promise.all([
      store.fetchDepartments(),
      store.fetchDepartments(),
      store.fetchDepartments()
    ])

    expect(getAllDepartmentsMock).toHaveBeenCalledTimes(1)
    expect(store.departmentsLoaded).toBe(true)
    expect(store.departmentsLoading).toBe(false)

    // Force fetch re-runs
    await store.fetchDepartments(true)
    expect(getAllDepartmentsMock).toHaveBeenCalledTimes(2)
  })
})

