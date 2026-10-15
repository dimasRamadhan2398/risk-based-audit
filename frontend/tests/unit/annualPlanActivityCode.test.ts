// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'
import { useAnnualPlanStore } from '~/stores/annual-audit'

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn()
  })
}))

const FORM_SFC = resolve(__dirname, '../../components/annual-audit/AnnualAuditForm.vue')

const validRow = (over: Record<string, unknown> = {}) => ({
  name: 'Audit IT',
  category: 'Assurance',
  departmentId: 'dept-it',
  department: 'IT',
  auditUniverseId: 'au-it',
  riskName: 'Risk A',
  auditorCount: 2,
  totalMandays: 10,
  supervisorName: 'Budi Santoso, CIA',
  ...over
})

function fillValidNewPlan(store: ReturnType<typeof useAnnualPlanStore>) {
  store.form.activities = [validRow({ name: 'Audit Keuangan', departmentId: 'dept-fin', department: 'Finance', riskName: 'Risk Finance', auditUniverseId: 'au-fin' })]
  store.form.year = '2026'
  store.form.selectedMonths = [0]
}

describe('Annual plan activity code (PKAT-<Tahun>-<No.Urut>)', () => {
  let store: ReturnType<typeof useAnnualPlanStore>
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    fetchMock = vi.fn(async (url: string, opts?: any) => {
      if (opts?.method === 'POST') {
        return { success: true, data: { id: 'new-1', code: 'PKAT-2026-001' } }
      }
      return { success: true, data: { items: [], pagination: { total_pages: 1 } } }
    })
    global.$fetch = fetchMock
    setActivePinia(createPinia())
    store = useAnnualPlanStore()
  })

  it('opening a new plan does not generate a client-side code', () => {
    store.openModal()

    expect(store.form.code).toBe('')
    expect(store.showModal).toBe(true)
  })

  it('POST for a new plan carries no code and takes the code returned by the server', async () => {
    store.openModal()
    fillValidNewPlan(store)

    await store.handleSubmit()

    const postCall = fetchMock.mock.calls.find(([, opts]) => opts?.method === 'POST')
    expect(postCall).toBeDefined()
    expect(postCall[1].body).not.toHaveProperty('code')
    expect(store.showModal).toBe(false)
    expect(store.form.code).toBe('PKAT-2026-001')
  })

  it('a new plan is not blocked by an empty code', async () => {
    store.openModal()
    fillValidNewPlan(store)

    await store.handleSubmit()

    expect(store.validationErrors.activityDetail).toBe('')
    expect(fetchMock.mock.calls.some(([, opts]) => opts?.method === 'POST')).toBe(true)
  })

  it('editing keeps the stored code unchanged, including older formats', async () => {
    store.handleEdit({
      id: 'plan-old',
      code: 'PKAT-2026-ASR-003-SPE',
      status: 'Draft',
      year: '2026',
      activities: [{ ...validRow(), id: 'row-uuid-1' }],
      selectedMonths: [0],
      isActive: true
    })

    expect(store.form.code).toBe('PKAT-2026-ASR-003-SPE')

    store.form.activities[0].category = 'Consulting'
    store.form.year = '2027'
    store.form.activities[0].supervisorName = 'Budi Santoso, CIA'
    await store.handleSubmit()

    const putCall = fetchMock.mock.calls.find(([, opts]) => opts?.method === 'PUT')
    expect(putCall).toBeDefined()
    expect(putCall[1].body.code).toBe('PKAT-2026-ASR-003-SPE')
    // The row keeps its server id, so the backend updates it instead of adding a new one.
    expect(putCall[1].body.activities[0].id).toBe('row-uuid-1')
  })

  it('the code field is read-only and has a pending placeholder key', () => {
    const src = readFileSync(FORM_SFC, 'utf-8')
    const codeInput = src.match(/<UInput\s+:model-value="activityCodePreview"[\s\S]*?\/>/)
    expect(codeInput, 'activity code UInput not found').not.toBeNull()
    expect(codeInput![0]).toContain('disabled')
    expect(codeInput![0]).toContain('t(\'annualAudit.form.activityCodePending\')')
    expect(src).toContain('t(\'annualAudit.form.activityCodeHelp\')')
    expect(src).not.toContain('PKAT-Tahun-CodeCategory')
  })
})
