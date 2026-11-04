import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useConsultingServiceStore, type ConsultingService } from '~/stores/consulting-service'
import * as departmentApiModule from '~/composables/useDepartmentApi'
import type { Department } from '~/types/master'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      auditServiceBaseUrl: 'http://localhost:8080/api/v1',
      masterServiceBaseUrl: 'http://localhost:8080/api/v1'
    }
  })
}))

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

;(globalThis as any).$fetch = vi.fn().mockResolvedValue([])

const PAGE_SFC = resolve(__dirname, '../../pages/consulting-service/index.vue')

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

describe('Consulting Service - Requestor Department Synchronized with Master Department Management', () => {
  let store: ReturnType<typeof useConsultingServiceStore>
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
    store = useConsultingServiceStore()
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

  it('generates departmentOptions with department_name and department_code sorted alphabetically', async () => {
    await store.fetchDepartments()

    expect(store.departmentOptions).toEqual([
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
    expect(store.resolveDepartmentOption('Custom Department')).toBe('Custom Department')
    expect(store.resolveDepartmentOption('')).toBe('')
    expect(store.resolveDepartmentOption(undefined)).toBe('')
  })

  it('formats department name for table and detail display correctly', async () => {
    await store.fetchDepartments()

    expect(store.formatDepartmentName('dept-1')).toBe('Financial (FIN)')
    expect(store.formatDepartmentName('OPS')).toBe('Operasional (OPS)')
    expect(store.formatDepartmentName('Product Development')).toBe('Product Development (DEV)')
    expect(store.formatDepartmentName('Financial (FIN)')).toBe('Financial (FIN)')
    expect(store.formatDepartmentName('Custom Department')).toBe('Custom Department')
    expect(store.formatDepartmentName('')).toBe('-')
    expect(store.formatDepartmentName(undefined)).toBe('-')
  })

  it('preserves existing requestorDept and resolves it when opening edit form', async () => {
    await store.fetchDepartments()

    const mockService: ConsultingService = {
      id: 'cs-1',
      title: 'Consulting Advisory IT',
      category: 'IT Advisory',
      requestorDept: 'Operasional',
      period: 'Q1 2026',
      consultantName: 'John Doe',
      status: 'Planned'
    }

    store.editService(mockService)

    expect(store.isEditing).toBe(true)
    expect(store.isFormOpen).toBe(true)
    expect(store.newService.requestorDept).toBe('Operasional (OPS)')
  })

  it('preserves custom/unmatched requestorDept when editing and includes it in departmentOptions', async () => {
    await store.fetchDepartments()

    const mockService: ConsultingService = {
      id: 'cs-custom',
      title: 'Consulting Adhoc',
      category: 'Governance Assessment',
      requestorDept: 'External Advisory Committee',
      period: 'Q2 2026',
      consultantName: 'Jane Smith',
      status: 'Planned'
    }

    store.editService(mockService)

    expect(store.newService.requestorDept).toBe('External Advisory Committee')
    expect(store.departmentOptions).toContain('External Advisory Committee')
  })

  it('resets requestorDept on openForm in create mode and triggers department fetch', async () => {
    await store.fetchDepartments()

    store.newService.requestorDept = 'Operasional (OPS)'
    store.openForm()

    expect(store.isEditing).toBe(false)
    expect(store.isFormOpen).toBe(true)
    expect(store.newService.requestorDept).toBe('')
  })

  it('handles API errors gracefully when fetching departments', async () => {
    getAllDepartmentsMock.mockRejectedValueOnce(new Error('Network error'))

    await store.fetchDepartments()

    expect(store.departmentsLoaded).toBe(false)
    expect(store.departmentsLoading).toBe(false)
    expect(store.departmentsError).toBe('Network error')
  })

  it('does not re-fetch departments if already loaded unless force=true', async () => {
    await Promise.all([
      store.fetchDepartments(),
      store.fetchDepartments(),
      store.fetchDepartments()
    ])

    expect(getAllDepartmentsMock).toHaveBeenCalledTimes(1)
    expect(store.departmentsLoaded).toBe(true)
    expect(store.departmentsLoading).toBe(false)

    await store.fetchDepartments(true)
    expect(getAllDepartmentsMock).toHaveBeenCalledTimes(2)
  })

  it('consulting-service index.vue binds requestorDept to store.departmentOptions with USelectMenu, loading, create-item, and empty fallback', () => {
    const sfc = readFileSync(PAGE_SFC, 'utf-8')

    expect(sfc).toMatch(/<UFormField[^>]*?:label="t\('consultingService\.dashboard\.form\.requestorDepartment'\)"/)
    expect(sfc).toContain(':items="store.departmentOptions"')
    expect(sfc).toContain('v-model="store.newService.requestorDept"')
    expect(sfc).toContain(':loading="store.departmentsLoading"')
    expect(sfc).toContain('create-item')
    expect(sfc).toContain('@create=')
    expect(sfc).toContain('masterData.department.empty')
  })

  it('consulting-service index.vue displays formatted requestorDept in table and detail modal', () => {
    const sfc = readFileSync(PAGE_SFC, 'utf-8')

    // Table cell formatting
    expect(sfc).toContain('#requestorDept-cell')
    expect(sfc).toContain('store.formatDepartmentName((row.original as any).requestorDept)')

    // Detail modal formatting
    expect(sfc).toContain('store.formatDepartmentName(store.selectedService.requestorDept)')
  })

  it('consulting-service index.vue pre-fetches departments on mount and validates requestorDept in validateStep', () => {
    const sfc = readFileSync(PAGE_SFC, 'utf-8')

    expect(sfc).toContain('store.fetchDepartments()')
    expect(sfc).toContain('store.newService.requestorDept')
    expect(sfc).toContain('consultingService.dashboard.form.requestorDepartmentRequired')
  })
})

