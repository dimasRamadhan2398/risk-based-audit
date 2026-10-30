import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import * as employeeApiModule from '~/composables/useEmployeeApi'
import type { Employee } from '~/types/master'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      auditServiceBaseUrl: 'http://localhost:8080/api/v1'
    }
  })
}))

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    success: vi.fn(),
    showSuccess: vi.fn(),
    error: vi.fn(),
    showError: vi.fn(),
    warning: vi.fn(),
    showWarning: vi.fn(),
    info: vi.fn(),
    showInfo: vi.fn()
  })
}))

;(globalThis as any).$fetch = vi.fn().mockResolvedValue({ data: [] })

const MODAL_SFC = resolve(__dirname, '../../components/audit-fieldwork/AuditFieldworkTestControlsModal.vue')
const TABLE_SFC = resolve(__dirname, '../../components/audit-fieldwork/AuditFieldworkTestControls.vue')

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

describe('Audit Fieldwork - Test Controls PIC Synchronized with Master Employee Management', () => {
  let store: ReturnType<typeof useAuditFieldworkStore>
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
    store = useAuditFieldworkStore()
    store.selectedAssignmentLetter = 'ST-001/SKAI/2026'
  })

  it('fetches employees from Employee Management API and filters out inactive employees', async () => {
    await store.fetchEmployees()

    expect(getAllEmployeesMock).toHaveBeenCalled()
    expect(store.employees).toHaveLength(3)
    expect(store.employees.some(e => e.id === 'emp-inactive')).toBe(false)
    expect(store.employees[0]?.full_name).toBe('Anton Hermawan') // Sorted alphabetically
  })

  it('generates picOptions with employee full_name and employee_code', async () => {
    await store.fetchEmployees()

    expect(store.picOptions).toContain('Anton Hermawan (EMP-003)')
    expect(store.picOptions).toContain('Budi Hartanto (EMP-001)')
    expect(store.picOptions).toContain('Carolina Wijaya (EMP-002)')
  })

  it('preserves existing PIC values when opening test control for edit and resolves to employee format', async () => {
    await store.fetchEmployees()

    store.editTestControl({
      id: 'tc-1',
      assignmentLetterId: 'ST-001/SKAI/2026',
      controlName: 'Otorisasi PO',
      controlDescription: 'Persetujuan pesanan pembelian di atas 50jt',
      controlType: 'Preventive',
      unitDivision: 'FIN',
      testProcedure: 'Periksa 10 sampel PO',
      testResult: 'Effective',
      finding: '',
      recommendation: '',
      mitigationPlan: 'Pertahankan prosedur',
      pic: 'Budi Hartanto',
      dueDate: '2026-12-31'
    })

    expect(store.isEditingTestControl).toBe(true)
    expect(store.showTestControlModal).toBe(true)
    // Matches and resolves to Employee Management format
    expect(store.testControlForm.pic).toBe('Budi Hartanto (EMP-001)')
  })

  it('preserves custom/unmatched PIC when editing and includes it in picOptions', async () => {
    await store.fetchEmployees()

    store.editTestControl({
      id: 'tc-custom',
      assignmentLetterId: 'ST-001/SKAI/2026',
      controlName: 'Kontrol Eksternal',
      controlDescription: 'Pengujian pihak ketiga',
      controlType: 'Detective',
      unitDivision: 'OP',
      testProcedure: 'Verifikasi laporan vendor',
      testResult: 'Partially Effective',
      finding: 'Ada keterlambatan laporan',
      recommendation: 'Perbaiki SLA',
      mitigationPlan: 'Review berkala dengan konsultan eksternal',
      pic: 'External Consultant',
      dueDate: '2026-11-30'
    })

    expect(store.testControlForm.pic).toBe('External Consultant')
    expect(store.picOptions).toContain('External Consultant')
  })

  it('resets pic on openTestControlModal in create mode', async () => {
    await store.fetchEmployees()

    store.testControlForm.pic = 'Budi Hartanto (EMP-001)'
    store.openTestControlModal()

    expect(store.isEditingTestControl).toBe(false)
    expect(store.testControlForm.pic).toBe('')
    expect(getAllEmployeesMock).toHaveBeenCalled()
  })

  it('formats employee name in table and view correctly', async () => {
    await store.fetchEmployees()

    // By ID
    expect(store.formatEmployeeName('emp-1')).toBe('Budi Hartanto (EMP-001)')
    // By code
    expect(store.formatEmployeeName('EMP-002')).toBe('Carolina Wijaya (EMP-002)')
    // By full name
    expect(store.formatEmployeeName('Anton Hermawan')).toBe('Anton Hermawan (EMP-003)')
    // Already formatted string
    expect(store.formatEmployeeName('Anton Hermawan (EMP-003)')).toBe('Anton Hermawan (EMP-003)')
    // Custom/unmatched string preserved
    expect(store.formatEmployeeName('External Auditor')).toBe('External Auditor')
    // Empty / undefined
    expect(store.formatEmployeeName('')).toBe('-')
    expect(store.formatEmployeeName(undefined)).toBe('-')
  })

  it('deduplicates fetchEmployees calls to prevent redundant network requests', async () => {
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

  it('AuditFieldworkTestControlsModal.vue binds PIC to store.picOptions with employee loading state and empty fallback', () => {
    const sfc = readFileSync(MODAL_SFC, 'utf-8')

    // PIC field binds to USelectMenu
    expect(sfc).toMatch(/<UFormField[^>]*?:label="t\('auditFieldwork\.testControls\.pic'\)"/)
    expect(sfc).toContain(':items="store.picOptions"')
    expect(sfc).toContain('v-model="store.testControlForm.pic"')
    expect(sfc).toContain(':loading="store.employeesLoading"')
    expect(sfc).toContain('create-item')
    expect(sfc).toContain('masterData.employee.empty')

    // Read-only detail view formats PIC with store.formatEmployeeName
    expect(sfc).toContain('store.formatEmployeeName(store.testControlForm.pic)')
  })

  it('AuditFieldworkTestControls.vue displays formatted PIC and triggers fetchEmployees', () => {
    const sfc = readFileSync(TABLE_SFC, 'utf-8')

    expect(sfc).toContain('store.formatEmployeeName(row.original.pic)')
    expect(sfc).toContain('store.fetchEmployees')
  })
})
