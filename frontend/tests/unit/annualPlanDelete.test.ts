// @ts-nocheck
/**
 * Annual Audit Plan delete flow (stores/annual-audit.ts handleDelete/deletePlan).
 *
 * Bug: clicking Cancel in the delete confirmation still showed the success toast
 * "Data Rencana Audit Berhasil Dihapus!", because deletePlan returned nothing on
 * cancel and handleDelete toasted unconditionally after awaiting it.
 *
 * Also covers the shared confirmation store (stores/global-modal.ts): dismissing or
 * replacing a pending confirmation must resolve it with false.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { useGlobalModalStore } from '~/stores/global-modal'

// vi.hoisted/vi.mock run before the imports above, so the store gets these spies.
const toastSpies = vi.hoisted(() => ({
  showSuccess: vi.fn(),
  showError: vi.fn(),
  showWarning: vi.fn(),
  showInfo: vi.fn()
}))

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => toastSpies
}))

// Nuxt auto-imports it in the app; the store calls it without an import.
global.useGlobalModalStore = useGlobalModalStore

const plan = (over: Record<string, unknown> = {}) => ({
  id: 'plan-1',
  code: 'PKAT-2026-001',
  year: '2026',
  status: 'Draft',
  isUsed: false,
  activities: [],
  ...over
})

let store: ReturnType<typeof useAnnualPlanStore>
let modal: ReturnType<typeof useGlobalModalStore>
let fetchMock: ReturnType<typeof vi.fn>

/** Calls other than the list refresh (GET), i.e. the delete/deactivate requests. */
const writeCalls = () => fetchMock.mock.calls.filter(([, opts]) => opts?.method && opts.method !== 'GET')
const listCalls = () => fetchMock.mock.calls.filter(([, opts]) => !opts?.method || opts.method === 'GET')

/** Starts handleDelete, waits for the confirmation to open, answers it, and waits for the flow to end. */
async function deleteAndAnswer(id: string, answer: boolean) {
  const done = store.handleDelete(id)
  await vi.waitFor(() => expect(modal.isOpen).toBe(true))
  modal.resolve(answer)
  await done
}

describe('Annual Audit Plan delete', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    fetchMock = vi.fn(async () => ({ success: true, data: { items: [], pagination: { total: 0 } } }))
    global.$fetch = fetchMock
    store = useAnnualPlanStore()
    modal = useGlobalModalStore()
    store.plans = [plan(), plan({ id: 'plan-used', code: 'PKAT-2026-002', isUsed: true })]
    fetchMock.mockClear()
  })

  it('Cancel sends no request, shows no toast and does not refresh the list', async () => {
    await deleteAndAnswer('plan-1', false)

    expect(writeCalls()).toHaveLength(0)
    expect(listCalls()).toHaveLength(0)
    expect(toastSpies.showSuccess).not.toHaveBeenCalled()
    expect(toastSpies.showError).not.toHaveBeenCalled()
  })

  it('Cancel on a plan in use does not deactivate it', async () => {
    await deleteAndAnswer('plan-used', false)

    expect(writeCalls()).toHaveLength(0)
    expect(toastSpies.showSuccess).not.toHaveBeenCalled()
  })

  it('Confirm deletes, refreshes the list and shows one success toast', async () => {
    await deleteAndAnswer('plan-1', true)

    expect(writeCalls()).toHaveLength(1)
    expect(writeCalls()[0][0]).toMatch(/\/annual-audit-plans\/plan-1$/)
    expect(writeCalls()[0][1].method).toBe('DELETE')
    expect(listCalls().length).toBeGreaterThan(0)
    expect(toastSpies.showSuccess).toHaveBeenCalledTimes(1)
    expect(toastSpies.showError).not.toHaveBeenCalled()
  })

  it('Confirm on a plan in use sends only isActive=false and shows the deactivated toast', async () => {
    await deleteAndAnswer('plan-used', true)

    expect(writeCalls()).toHaveLength(1)
    expect(writeCalls()[0][1]).toMatchObject({ method: 'PUT', body: { isActive: false } })
    expect(toastSpies.showSuccess).toHaveBeenCalledTimes(1)
    const message = toastSpies.showSuccess.mock.calls[0][0]
    expect(typeof message).toBe('string')
    expect(message).not.toMatch(/^annualAudit\./) // resolved text, not a missing i18n key
  })

  it('A failed delete shows one error toast and no success toast', async () => {
    fetchMock.mockImplementation(async (_url, opts) => {
      if (opts?.method === 'DELETE') throw Object.assign(new Error('boom'), { data: { message: 'Server error' } })
      return { success: true, data: { items: [] } }
    })

    await deleteAndAnswer('plan-1', true)

    expect(toastSpies.showError).toHaveBeenCalledTimes(1)
    expect(toastSpies.showSuccess).not.toHaveBeenCalled()
  })

  it('The confirmation text is not hardcoded Indonesian', async () => {
    const done = store.handleDelete('plan-1')
    await vi.waitFor(() => expect(modal.isOpen).toBe(true))
    const text = [modal.options.description, ...(modal.options.body || [])].join(' ')
    modal.resolve(false)
    await done

    expect(text).not.toMatch(/Apakah Anda yakin/)
    expect(modal.options.type).toBe('delete')
  })
})

describe('Global confirmation store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('confirmDelete with an options object keeps type "delete"', () => {
    const modal = useGlobalModalStore()
    modal.confirmDelete({ description: 'x' })
    expect(modal.options.type).toBe('delete')
    modal.resolve(false)
  })

  it('a pending confirmation resolves false when another one opens', async () => {
    const modal = useGlobalModalStore()
    const first = modal.confirmDelete({ description: 'first' })
    const second = modal.confirmSubmit({ description: 'second' })

    await expect(first).resolves.toBe(false)
    modal.resolve(true)
    await expect(second).resolves.toBe(true)
  })
})
