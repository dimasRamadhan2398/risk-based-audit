/**
 * Employee create now returns a login account result: toast on created/linked,
 * one-time temporary password surfaced then cleared, readable error mapping,
 * and the ADMIN-only POST /employees/:id/login-account composable method.
 */
import { describe, it, expect, beforeAll, beforeEach, vi, type Mock } from 'vitest'
import { ref, reactive, computed } from 'vue'
import { defineStore, setActivePinia, createPinia } from 'pinia'
import { useEmployeeApi } from '~/composables/useEmployeeApi'
import { getUserErrorMessage } from '~/utils/error'
import en from '~/locales/en/common.json'
import id from '~/locales/id/common.json'

type FetchOpts = { method?: string, body?: Record<string, unknown> }
type FetchFn = (url: string, opts?: FetchOpts) => Promise<unknown>
const g = globalThis as unknown as { $fetch: Mock<FetchFn>, [name: string]: unknown }

const toastSuccess = vi.fn()
type Tree = { [k: string]: Tree | string }
const lookup = (obj: Tree, path: string): string | undefined => {
  let node: Tree | string | undefined = obj
  for (const k of path.split('.')) node = typeof node === 'object' ? node[k] : undefined
  return typeof node === 'string' ? node : undefined
}
const t = (key: string, params?: Record<string, string | number>) => {
  let s = lookup(en, key) ?? key
  for (const [k, v] of Object.entries(params || {})) s = s.replace(`{${k}}`, String(v))
  return s
}

vi.mock('~/composables/useAppToast', () => ({
  useAppToast: () => ({ success: toastSuccess, error: vi.fn(), warning: vi.fn(), info: vi.fn() })
}))
vi.mock('~/composables/useI18n', () => ({ useI18n: () => ({ t }) }))

let useEmployeeStore: typeof import('~/stores/employee').useEmployeeStore

const created = (account?: unknown) => ({
  success: true,
  data: { id: 'emp-1', employee_code: 'EMP-0002', full_name: 'Siti', email: 's@example.com', account }
})

const fillForm = (store: ReturnType<typeof useEmployeeStore>) => {
  Object.assign(store.form, {
    full_name: 'Siti', email: 's@example.com', company_id: 'c', department_id: 'd', job_role_id: 'r'
  })
}

const fetchError = (statusCode: number, code: string) =>
  Object.assign(new Error('fail'), { statusCode, data: { success: false, error: { code, message: 'raw backend text' } } })

describe('employee create login account result', () => {
  beforeAll(async () => {
    g.defineStore = defineStore
    g.ref = ref
    g.reactive = reactive
    g.computed = computed
    g.useEmployeeApi = useEmployeeApi
    ;({ useEmployeeStore } = await import('~/stores/employee'))
  })

  let post: (() => unknown) | null
  beforeEach(() => {
    setActivePinia(createPinia())
    toastSuccess.mockClear()
    vi.spyOn(console, 'error').mockImplementation(() => {})
    post = null
    g.$fetch = vi.fn(async (_url: string, opts?: FetchOpts) => {
      if (opts?.method === 'POST') return post!()
      return { success: true, data: { employees: [], pagination: { page: 1, page_size: 10, total: 0, total_pages: 0 } } }
    }) as unknown as Mock<FetchFn>
  })

  it('toasts and surfaces the one-time temporary password when created', async () => {
    post = () => created({ status: 'created', username: 'siti', temporary_password: 'Tmp-Pass-1' })
    const store = useEmployeeStore()
    fillForm(store)
    await store.handleSubmit()

    expect(toastSuccess).toHaveBeenCalledTimes(1)
    expect(toastSuccess.mock.calls[0]!.join(' ')).toContain('siti')
    // the password itself must never go in a toast
    expect(JSON.stringify(toastSuccess.mock.calls)).not.toContain('Tmp-Pass-1')
    expect(store.createdAccount?.temporary_password).toBe('Tmp-Pass-1')
    expect(store.showModal).toBe(false)

    store.clearCreatedAccount()
    expect(store.createdAccount).toBeNull()
  })

  it('toasts for a linked account and keeps no password state', async () => {
    post = () => created({ status: 'linked', username: 'siti' })
    const store = useEmployeeStore()
    fillForm(store)
    await store.handleSubmit()

    expect(toastSuccess.mock.calls[0]!.join(' ')).toContain(en.masterData.employee.account.linkedToast.split('{')[0])
    expect(store.createdAccount).toBeNull()
  })

  it('maps login account errors to readable messages and keeps the form open', async () => {
    post = () => {
      throw fetchError(502, 'LOGIN_ACCOUNT_UNAVAILABLE')
    }
    const store = useEmployeeStore()
    store.openCreateModal()
    fillForm(store)
    await store.handleSubmit()

    expect(store.errorMsg).toBe(en.errors.codes.LOGIN_ACCOUNT_UNAVAILABLE)
    expect(store.errorMsg).not.toContain('raw backend text')
    expect(store.showModal).toBe(true)
    expect(store.createdAccount).toBeNull()
    expect(toastSuccess).not.toHaveBeenCalled()
  })

  it.each([
    'LOGIN_ACCOUNT_UNAVAILABLE', 'LOGIN_ACCOUNT_NOT_CONFIGURED', 'LOGIN_ACCOUNT_FORBIDDEN',
    'LOGIN_ACCOUNT_CONFLICT', 'LOGIN_ACCOUNT_LINKED_ELSEWHERE', 'LOGIN_EMAIL_TOO_LONG'
  ])('has en and id copy for %s', (code) => {
    const err = fetchError(409, code)
    expect(getUserErrorMessage(err, k => lookup(en, k) ?? k)).toBe((en.errors.codes as Record<string, string>)[code])
    expect(getUserErrorMessage(err, k => lookup(id, k) ?? k)).toBe((id.errors.codes as Record<string, string>)[code])
    expect((id.errors.codes as Record<string, string>)[code]).toBeTruthy()
  })

  it('keeps en/id key parity for the account block', () => {
    expect(Object.keys(id.masterData.employee.account).sort()).toEqual(Object.keys(en.masterData.employee.account).sort())
  })
})

describe('useEmployeeApi.createLoginAccount', () => {
  beforeEach(() => {
    g.$fetch = vi.fn(async () => ({ success: true, data: { account: { status: 'created', username: 'u', temporary_password: 'p' } } })) as unknown as Mock<FetchFn>
  })

  it('posts the access role and returns the account', async () => {
    const account = await useEmployeeApi().createLoginAccount('emp-1', 'AUDITOR')
    const [url, opts] = g.$fetch.mock.calls[0]!
    expect(url).toMatch(/\/employees\/emp-1\/login-account$/)
    expect(opts?.method).toBe('POST')
    expect(opts?.body).toEqual({ access_role: 'AUDITOR' })
    expect(account.status).toBe('created')
  })

  it('sends an empty body when no role is chosen', async () => {
    await useEmployeeApi().createLoginAccount('emp-1')
    expect(g.$fetch.mock.calls[0]![1]?.body).toEqual({})
  })
})
