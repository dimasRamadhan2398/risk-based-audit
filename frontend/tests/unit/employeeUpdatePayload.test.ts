/**
 * stores/employee.ts updateEmployee: the PUT body must carry company_id (the
 * backend rejects an empty company) and join_date. join_date used to be left
 * out, and the backend overwrote the stored date with 0001-01-01.
 */
import { describe, it, expect, beforeAll, beforeEach, vi, type Mock } from 'vitest'
import { ref, reactive, computed } from 'vue'
import { defineStore, setActivePinia, createPinia } from 'pinia'
import { useEmployeeApi } from '~/composables/useEmployeeApi'

type FetchOpts = { method?: string, body?: Record<string, unknown>, params?: Record<string, unknown> }
type FetchFn = (url: string, opts?: FetchOpts) => Promise<unknown>
const g = globalThis as unknown as { $fetch: Mock<FetchFn>, [name: string]: unknown }

const employee = {
  id: 'emp-1',
  employee_code: 'EMP-0001',
  full_name: 'Budi Santoso',
  email: 'budi@example.com',
  phone: '0812-3456',
  company_id: 'company-1',
  department_id: 'dept-1',
  job_role_id: 'role-1',
  level_grade: 3,
  work_location_id: 'loc-1',
  residence_address: '',
  residence_city: '',
  residence_province: '',
  residence_postal_code: '',
  manager_id: '',
  is_active: true,
  join_date: '2021-03-15T00:00:00Z',
  created_at: '2021-03-15T00:00:00Z',
  updated_at: '2021-03-15T00:00:00Z'
}

let useEmployeeStore: typeof import('~/stores/employee').useEmployeeStore

const putBody = () => {
  const call = g.$fetch.mock.calls.find(([, o]) => o?.method === 'PUT')
  expect(call).toBeDefined()
  expect(call?.[0]).toMatch(/\/employees\/emp-1$/)
  return call?.[1]?.body as Record<string, unknown>
}

describe('employee store updateEmployee payload', () => {
  beforeAll(async () => {
    // The store relies on Nuxt auto-imports; tests/setup.ts stubs ref/computed
    // with non-reactive fakes, so provide the real ones here.
    g.defineStore = defineStore
    g.ref = ref
    g.reactive = reactive
    g.computed = computed
    g.useEmployeeApi = useEmployeeApi
    ;({ useEmployeeStore } = await import('~/stores/employee'))
  })

  beforeEach(() => {
    setActivePinia(createPinia())
    g.$fetch = vi.fn(async (_url: string, opts?: FetchOpts) => {
      if (opts?.method === 'PUT') return { success: true, data: employee }
      return { success: true, data: { employees: [employee], pagination: { page: 1, page_size: 10, total: 1, total_pages: 1 } } }
    }) as unknown as Mock<FetchFn>
  })

  it('sends company_id and the stored join_date when editing', async () => {
    const store = useEmployeeStore()
    store.handleEdit(employee)
    expect(store.form.join_date).toBe('2021-03-15')

    expect(await store.updateEmployee('emp-1')).toBe(true)

    const body = putBody()
    expect(body.employee_code).toBe('EMP-0001')
    expect(body.company_id).toBe('company-1')
    expect(body.join_date).toBe('2021-03-15T00:00:00Z')
  })

  it('sends a join_date changed in the form as RFC 3339', async () => {
    const store = useEmployeeStore()
    store.handleEdit(employee)
    store.form.join_date = '2023-07-01'

    await store.updateEmployee('emp-1')

    expect(putBody().join_date).toBe('2023-07-01T00:00:00Z')
  })

  it('omits join_date when the stored date is the zero date, so the backend keeps it', async () => {
    const store = useEmployeeStore()
    store.handleEdit({ ...employee, join_date: '0001-01-01T00:00:00Z' })
    expect(store.form.join_date).toBe('')

    await store.updateEmployee('emp-1')

    const body = putBody()
    expect(body.company_id).toBe('company-1')
    expect(body.join_date).toBeUndefined()
  })
})
