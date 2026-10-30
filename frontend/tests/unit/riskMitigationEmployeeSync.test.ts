import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useMitigationStore } from '~/stores/mitigation-risk'
import * as employeeApiModule from '~/composables/useEmployeeApi'
import type { Employee } from '~/types/master'

const FORM_SFC = resolve(__dirname, '../../components/risk-mitigation/MitigationForm.vue')
const TABLE_SFC = resolve(__dirname, '../../components/risk-mitigation/MitigationTable.vue')

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

describe('Risk Mitigation - Supervisor & PIC Synchronized with Master Employee Management', () => {
  let store: ReturnType<typeof useMitigationStore>
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
    store = useMitigationStore()
  })

  it('fetches employees from Employee Management API and filters out inactive employees', async () => {
    await store.fetchEmployees()

    expect(getAllEmployeesMock).toHaveBeenCalled()
    expect(store.employees).toHaveLength(3)
    expect(store.employees.some(e => e.id === 'emp-inactive')).toBe(false)
    expect(store.employees[0]?.full_name).toBe('Anton Hermawan') // Sorted alphabetically
  })

  it('generates supervisorOptions and picOptions with employee full_name and employee_code', async () => {
    await store.fetchEmployees()

    expect(store.supervisorOptions).toContain('Anton Hermawan (EMP-003)')
    expect(store.supervisorOptions).toContain('Budi Hartanto (EMP-001)')
    expect(store.supervisorOptions).toContain('Carolina Wijaya (EMP-002)')

    expect(store.picOptions).toContain('Anton Hermawan (EMP-003)')
    expect(store.picOptions).toContain('Budi Hartanto (EMP-001)')
    expect(store.picOptions).toContain('Carolina Wijaya (EMP-002)')
  })

  it('preserves existing supervisor and pic values when opening form for edit', async () => {
    await store.fetchEmployees()

    // Open form with existing legacy name
    store.openForm({
      id: 'mit-1',
      riskId: 'risk-100',
      riskControlId: 'CTL-RSK-001',
      riskEvent: 'Event 1',
      mitigationPlan: 'Plan 1',
      supervisor: 'Budi Hartanto',
      pic: 'Carolina Wijaya',
      unitInCharge: 'Financial',
      start_date: '2026-01-01T00:00:00Z',
      end_date: '2026-12-31T00:00:00Z',
      notes: ''
    })

    expect(store.isEditing).toBe(true)
    // Matches and resolves to Employee Management format
    expect(store.form.supervisor).toBe('Budi Hartanto (EMP-001)')
    expect(store.form.pic).toBe('Carolina Wijaya (EMP-002)')
  })

  it('preserves custom/unmatched supervisor or pic when editing and includes it in options', async () => {
    await store.fetchEmployees()

    store.openForm({
      id: 'mit-custom',
      riskId: 'risk-100',
      riskControlId: 'CTL-RSK-002',
      riskEvent: 'Event 2',
      mitigationPlan: 'Plan 2',
      supervisor: 'External Consultant',
      pic: 'External Auditor',
      unitInCharge: 'Operasional',
      start_date: '2026-01-01T00:00:00Z',
      end_date: '2026-12-31T00:00:00Z',
      notes: ''
    })

    expect(store.form.supervisor).toBe('External Consultant')
    expect(store.form.pic).toBe('External Auditor')
    expect(store.supervisorOptions).toContain('External Consultant')
    expect(store.picOptions).toContain('External Auditor')
  })

  it('resets supervisor and pic on openForm in create/add mode', async () => {
    await store.fetchEmployees()

    store.openForm()

    expect(store.isEditing).toBe(false)
    expect(store.form.supervisor).toBe('')
    expect(store.form.pic).toBe('')
  })

  it('formats employee name in table correctly', async () => {
    await store.fetchEmployees()

    // By ID
    expect(store.formatEmployeeName('emp-1')).toBe('Budi Hartanto (EMP-001)')
    // By code
    expect(store.formatEmployeeName('EMP-002')).toBe('Carolina Wijaya (EMP-002)')
    // By plain text already formatted
    expect(store.formatEmployeeName('Anton Hermawan')).toBe('Anton Hermawan')
    // Empty
    expect(store.formatEmployeeName('')).toBe('-')
    expect(store.formatEmployeeName(undefined)).toBe('-')
  })

  it('MitigationForm.vue binds supervisor and pic to store options with employee loading state', () => {
    const sfc = readFileSync(FORM_SFC, 'utf-8')

    // Supervisor field
    expect(sfc).toMatch(/<UFormField[^>]*?:label="t\('riskMitigation\.form\.supervisor'\)"/)
    expect(sfc).toContain(':items="store.supervisorOptions"')
    expect(sfc).toContain('v-model="store.form.supervisor"')

    // PIC field
    expect(sfc).toMatch(/<UFormField[^>]*?:label="t\('riskMitigation\.form\.pic'\)"/)
    expect(sfc).toContain(':items="store.picOptions"')
    expect(sfc).toContain('v-model="store.form.pic"')

    // Loading states & empty templates
    expect(sfc).toContain(':loading="store.employeesLoading"')
    expect(sfc).toContain('masterData.employee.empty')
  })

  it('MitigationTable.vue uses store.formatEmployeeName for supervisor and pic display', () => {
    const sfc = readFileSync(TABLE_SFC, 'utf-8')

    expect(sfc).toContain('store.formatEmployeeName(row.pic)')
    expect(sfc).toContain('store.formatEmployeeName(row.supervisor)')
  })

  it('deduplicates fetchEmployees calls to prevent continuous loading loops', async () => {
    // Calling fetchEmployees multiple times concurrently or sequentially
    await Promise.all([
      store.fetchEmployees(),
      store.fetchEmployees(),
      store.fetchEmployees()
    ])

    expect(getAllEmployeesMock).toHaveBeenCalledTimes(1)
    expect(store.employeesLoaded).toBe(true)
    expect(store.employeesLoading).toBe(false)

    // Force fetch re-runs
    await store.fetchEmployees(true)
    expect(getAllEmployeesMock).toHaveBeenCalledTimes(2)
  })

  it('MitigationTable.vue uses watch instead of watchEffect to avoid infinite reactive loops', () => {
    const sfc = readFileSync(TABLE_SFC, 'utf-8')

    expect(sfc).not.toContain('watchEffect')
    expect(sfc).toMatch(/watch\s*\(\s*\(\)\s*=>\s*props\.currentRiskId/)
  })

  it('getMitigationsByRiskId handles case-insensitive UUID matching and fallbacks', () => {
    store.mitigations = [
      {
        id: 'mit-1',
        riskId: 'A1B2-C3D4',
        riskControlId: 'CTL-1',
        riskEvent: 'Event 1',
        mitigationPlan: 'Plan 1',
        supervisor: 'Budi',
        pic: 'Anton',
        unitInCharge: 'Financial',
        start_date: '2026-01-01',
        end_date: '2026-02-01'
      }
    ]

    // Uppercase query
    expect(store.getMitigationsByRiskId('A1B2-C3D4')).toHaveLength(1)
    // Lowercase query
    expect(store.getMitigationsByRiskId('a1b2-c3d4')).toHaveLength(1)
    // Undefined returns all
    expect(store.getMitigationsByRiskId(undefined)).toHaveLength(1)
    // Different risk returns empty
    expect(store.getMitigationsByRiskId('other-risk')).toHaveLength(0)
  })
})
