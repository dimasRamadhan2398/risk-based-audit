// @ts-nocheck
import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import * as employeeApiModule from '~/composables/useEmployeeApi'
import type { Employee } from '~/types/master'

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn()
  })
}))

const FORM_SFC = resolve(__dirname, '../../components/annual-audit/AnnualAuditForm.vue')
const DETAIL_SFC = resolve(__dirname, '../../components/annual-audit/AnnualAuditDetail.vue')

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

// The supervisor (Pengendali Teknis) is a per-row field: the plan has no supervisor of its own.
const validRow = (over: Record<string, unknown> = {}) => ({
  name: 'Audit Keuangan',
  category: 'Assurance',
  departmentId: 'dept-fin',
  department: 'Finance',
  auditUniverseId: 'au-fin',
  riskName: 'Risk Finance',
  auditorCount: 2,
  totalMandays: 10,
  supervisorName: 'Budi Santoso, CIA',
  ...over
})

describe('Annual Audit Plan (PKAT) - Supervisor Synchronized with Master Employee Management', () => {
  let store: ReturnType<typeof useAnnualPlanStore>
  let getAllEmployeesMock: Mock

  beforeEach(() => {
    global.$fetch = vi.fn(async () => ({ success: true, data: { items: [], pagination: { total_pages: 1 } } }))
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
    store = useAnnualPlanStore()
  })

  it('renders the supervisor field as USelectMenu bound to store.supervisorOptions in AnnualAuditForm.vue', () => {
    const src = readFileSync(FORM_SFC, 'utf-8')
    const leadFieldMatch = src.match(/<UFormField[^>]*?:label="t\('annualAudit\.form\.teamLead'\)"[\s\S]*?<\/UFormField>/)
    expect(leadFieldMatch, 'team lead UFormField not found').not.toBeNull()

    const leadBlock = leadFieldMatch![0]
    expect(leadBlock).toContain('<USelectMenu')
    expect(leadBlock).toContain('v-model="activity.supervisorName"')
    expect(leadBlock).toContain(':items="store.supervisorOptions"')
    expect(leadBlock).toContain(':loading="store.employeesLoading"')
    expect(leadBlock).toContain('masterData.employee.empty')
  })

  it('has no plan-level supervisor field left in the form', () => {
    const src = readFileSync(FORM_SFC, 'utf-8')
    expect(src).not.toContain('store.form.supervisorId')
    expect(src).not.toContain('annualAudit.form.supervisor\'')
  })

  it('fetches employees from Employee Management API and filters out inactive employees', async () => {
    await store.fetchEmployees()

    expect(getAllEmployeesMock).toHaveBeenCalled()
    expect(store.employees).toHaveLength(3)
    expect(store.employees.some(e => e.id === 'emp-inactive')).toBe(false)
    expect(store.employees[0]?.full_name).toBe('Anton Hermawan') // Sorted alphabetically
  })

  it('generates supervisorOptions with employee full_name and employee_code', async () => {
    await store.fetchEmployees()

    expect(store.supervisorOptions).toContain('Anton Hermawan (EMP-003)')
    expect(store.supervisorOptions).toContain('Budi Hartanto (EMP-001)')
    expect(store.supervisorOptions).toContain('Carolina Wijaya (EMP-002)')
  })

  it('preserves existing supervisor value and resolves to Master Data format when editing an existing plan (Update plan)', async () => {
    await store.fetchEmployees()

    store.handleEdit({
      id: 'plan-123',
      code: 'PKAT-2026-ASR-001',
      version: 'v1.0',
      status: 'Approved',
      selectedMonths: [0, 1],
      isActive: true,
      activities: [{ ...validRow({ supervisorName: 'Budi Hartanto' }), id: 'row-1' }]
    })

    expect(store.isEditing).toBe(true)
    expect(store.form.activities[0].supervisorName).toBe('Budi Hartanto (EMP-001)')
    expect(store.form.activities[0].id).toBe('row-1')
  })

  it('preserves custom/unmatched supervisor when editing and includes it in supervisorOptions', async () => {
    await store.fetchEmployees()

    store.handleEdit({
      id: 'plan-custom',
      code: 'PKAT-2026-ASR-002',
      version: 'v1.0',
      status: 'Approved',
      selectedMonths: [0],
      isActive: true,
      activities: [{ ...validRow({ supervisorName: 'External Auditor' }), id: 'row-2' }]
    })

    expect(store.form.activities[0].supervisorName).toBe('External Auditor')
    expect(store.supervisorOptions).toContain('External Auditor')
  })

  it('resets supervisor on openModal in create mode', async () => {
    await store.fetchEmployees()

    store.openModal()

    expect(store.isEditing).toBe(false)
    expect(store.form.activities[0].supervisorName).toBe('')
  })

  it('formats employee name with formatEmployeeName correctly', async () => {
    await store.fetchEmployees()

    // By ID
    expect(store.formatEmployeeName('emp-1')).toBe('Budi Hartanto (EMP-001)')
    // By code
    expect(store.formatEmployeeName('EMP-002')).toBe('Carolina Wijaya (EMP-002)')
    // By name
    expect(store.formatEmployeeName('Anton Hermawan')).toBe('Anton Hermawan (EMP-003)')
    // By already formatted name
    expect(store.formatEmployeeName('Anton Hermawan (EMP-003)')).toBe('Anton Hermawan (EMP-003)')
    // Unmatched
    expect(store.formatEmployeeName('Custom Consultant')).toBe('Custom Consultant')
    // Empty
    expect(store.formatEmployeeName('')).toBe('-')
    expect(store.formatEmployeeName(undefined)).toBe('-')
  })

  it('deduplicates fetchEmployees calls to prevent continuous loading loops', async () => {
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

  it('AnnualAuditDetail.vue uses store.formatEmployeeName for supervisor display', () => {
    const sfc = readFileSync(DETAIL_SFC, 'utf-8')
    expect(sfc).toContain('store.formatEmployeeName')
  })

  it('validates the supervisor of each row: fails if empty or whitespace', async () => {
    store.form.activities = [validRow({ supervisorName: '   ' })]
    store.form.year = '2026'
    store.form.selectedMonths = [0]

    await store.handleSubmit()

    expect(store.validationErrors.auditor).not.toBe('')
    const postCall = global.$fetch.mock.calls.find(([, opts]) => opts?.method === 'POST')
    expect(postCall).toBeUndefined()
  })

  it('submits with the supervisor on each row and no plan-level supervisor', async () => {
    store.form.activities = [validRow()]
    store.form.year = '2026'
    store.form.selectedMonths = [0]

    await store.handleSubmit()

    expect(store.validationErrors.auditor).toBe('')
    const postCall = global.$fetch.mock.calls.find(([, opts]) => opts?.method === 'POST')
    expect(postCall).toBeDefined()
    expect(postCall[1].body.activities[0].supervisorName).toBe('Budi Santoso, CIA')
    expect(postCall[1].body).not.toHaveProperty('supervisorId')
    expect(postCall[1].body).not.toHaveProperty('supervisorName')
  })
})
