import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import * as departmentApiModule from '~/composables/useDepartmentApi'
import * as employeeApiModule from '~/composables/useEmployeeApi'
import type { Department } from '~/types/master'

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

describe('Audit Fieldwork - Test Controls Unit Division Synchronized with Master Department Management', () => {
  let store: ReturnType<typeof useAuditFieldworkStore>
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
    store = useAuditFieldworkStore()
    store.selectedAssignmentLetter = 'ST-001/SKAI/2026'
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

  it('generates unitDivisionOptions with department_name and department_code sorted alphabetically', async () => {
    await store.fetchDepartments()

    expect(store.unitDivisionOptions).toEqual([
      'Financial (FIN)',
      'Operasional (OPS)',
      'Product Development (DEV)'
    ])
  })

  it('preserves existing unitDivision values when opening test control for edit and resolves to department format', async () => {
    await store.fetchDepartments()

    store.editTestControl({
      id: 'tc-1',
      assignmentLetterId: 'ST-001/SKAI/2026',
      controlName: 'Otorisasi PO',
      controlDescription: 'Persetujuan pesanan pembelian',
      controlType: 'Preventive',
      unitDivision: 'Operasional',
      testProcedure: 'Periksa 10 sampel PO',
      testResult: 'Effective',
      finding: '',
      recommendation: '',
      mitigationPlan: 'Pertahankan prosedur',
      pic: '',
      dueDate: '2026-12-31'
    })

    expect(store.isEditingTestControl).toBe(true)
    expect(store.showTestControlModal).toBe(true)
    // Matches and resolves to Master Data Department format
    expect(store.testControlForm.unitDivision).toBe('Operasional (OPS)')
  })

  it('preserves custom/unmatched unitDivision when editing and includes it in unitDivisionOptions', async () => {
    await store.fetchDepartments()

    store.editTestControl({
      id: 'tc-custom',
      assignmentLetterId: 'ST-001/SKAI/2026',
      controlName: 'Kontrol Eksternal',
      controlDescription: 'Pengujian pihak ketiga',
      controlType: 'Detective',
      unitDivision: 'External Taskforce',
      testProcedure: 'Verifikasi laporan vendor',
      testResult: 'Partially Effective',
      finding: '',
      recommendation: '',
      mitigationPlan: '',
      pic: '',
      dueDate: '2026-11-30'
    })

    expect(store.testControlForm.unitDivision).toBe('External Taskforce')
    expect(store.unitDivisionOptions).toContain('External Taskforce')
  })

  it('resets unitDivision on openTestControlModal in create mode', async () => {
    await store.fetchDepartments()

    store.testControlForm.unitDivision = 'Operasional (OPS)'
    store.openTestControlModal()

    expect(store.isEditingTestControl).toBe(false)
    expect(store.testControlForm.unitDivision).toBe('')
    expect(getAllDepartmentsMock).toHaveBeenCalled()
  })

  it('formats department name in table and view correctly', async () => {
    await store.fetchDepartments()

    // By ID
    expect(store.formatDepartmentName('dept-1')).toBe('Financial (FIN)')
    // By code
    expect(store.formatDepartmentName('OPS')).toBe('Operasional (OPS)')
    // By name
    expect(store.formatDepartmentName('Product Development')).toBe('Product Development (DEV)')
    // Already formatted string
    expect(store.formatDepartmentName('Financial (FIN)')).toBe('Financial (FIN)')
    // Custom/unmatched string preserved
    expect(store.formatDepartmentName('Special Division')).toBe('Special Division')
    // Empty / undefined
    expect(store.formatDepartmentName('')).toBe('-')
    expect(store.formatDepartmentName(undefined)).toBe('-')
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

    // Force fetch re-runs
    await store.fetchDepartments(true)
    expect(getAllDepartmentsMock).toHaveBeenCalledTimes(2)
  })

  it('AuditFieldworkTestControlsModal.vue binds unitDivision to store.unitDivisionOptions with department loading state, create-item and empty fallback', () => {
    const sfc = readFileSync(MODAL_SFC, 'utf-8')

    // unitDivision field binds to USelectMenu
    expect(sfc).toMatch(/<UFormField[^>]*?:label="t\('auditFieldwork\.testControls\.unitDivision'\)"/)
    expect(sfc).toContain(':items="store.unitDivisionOptions"')
    expect(sfc).toContain('v-model="store.testControlForm.unitDivision"')
    expect(sfc).toContain(':loading="store.departmentsLoading"')
    expect(sfc).toContain('create-item')
    expect(sfc).toContain('masterData.department.empty')

    // Read-only detail view formats unitDivision with store.formatDepartmentName
    expect(sfc).toContain('store.formatDepartmentName(store.testControlForm.unitDivision)')
  })

  it('AuditFieldworkTestControls.vue displays formatted unitDivision and triggers fetchDepartments', () => {
    const sfc = readFileSync(TABLE_SFC, 'utf-8')

    expect(sfc).toContain('store.formatDepartmentName(row.original.unitDivision)')
    expect(sfc).toContain('store.fetchDepartments')
  })
})

