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

describe('Annual Audit Form - Supervisor Input Field', () => {
  let store: ReturnType<typeof useAnnualPlanStore>

  beforeEach(() => {
    global.$fetch = vi.fn(async () => ({ success: true, data: { items: [], pagination: { total_pages: 1 } } }))
    setActivePinia(createPinia())
    store = useAnnualPlanStore()
  })

  it('renders Supervisor field as UInput and not USelectMenu in AnnualAuditForm.vue', () => {
    const src = readFileSync(FORM_SFC, 'utf-8')
    const supervisorFieldMatch = src.match(/<UFormField[^>]*?label="Supervisor"[\s\S]*?<\/UFormField>/)
    expect(supervisorFieldMatch, 'Supervisor UFormField not found').not.toBeNull()

    const supervisorBlock = supervisorFieldMatch![0]
    expect(supervisorBlock).toContain('<UInput')
    expect(supervisorBlock).toContain('v-model="store.form.supervisorId"')
    expect(supervisorBlock).not.toContain('<USelectMenu')
  })

  it('validates supervisor: fails if empty or whitespace', async () => {
    store.form.code = 'PKAT-2026-ASR-001'
    store.form.activities = [
      { name: 'Audit Keuangan', category: 'Assurance', department: 'Finance', riskName: 'Risk Finance' }
    ]
    store.form.year = '2026'
    store.form.selectedMonths = [0]
    store.form.auditorCount = 2
    store.form.daysPerAuditor = 5
    store.form.supervisorId = '   '

    await store.handleSubmit()

    expect(store.validationErrors.auditor).toContain('Supervisor wajib diisi')
    const postCall = global.$fetch.mock.calls.find(([, opts]) => opts?.method === 'POST')
    expect(postCall).toBeUndefined()
  })

  it('submits successfully when supervisor name is provided as text input', async () => {
    store.form.code = 'PKAT-2026-ASR-001'
    store.form.activities = [
      { name: 'Audit Keuangan', category: 'Assurance', department: 'Finance', riskName: 'Risk Finance' }
    ]
    store.form.year = '2026'
    store.form.selectedMonths = [0]
    store.form.auditorCount = 2
    store.form.daysPerAuditor = 5
    store.form.supervisorId = 'Budi Santoso, CIA'

    await store.handleSubmit()

    expect(store.validationErrors.auditor).toBe('')
    const postCall = global.$fetch.mock.calls.find(([, opts]) => opts?.method === 'POST')
    expect(postCall).toBeDefined()
    expect(postCall[1].body.supervisorId).toBe('Budi Santoso, CIA')
    expect(postCall[1].body.supervisorName).toBe('Budi Santoso, CIA')
  })

  it('populates supervisorId correctly during handleEdit', () => {
    const existingPlan = {
      id: 'plan-123',
      code: 'PKAT-2026-ASR-001',
      version: 'v1.0',
      status: 'Done',
      selectedMonths: [0, 1],
      auditorCount: 3,
      daysPerAuditor: 4,
      supervisorId: 'S01',
      supervisorName: 'Budi Santoso (Mgr)',
      activities: []
    }

    store.handleEdit(existingPlan)

    expect(store.form.supervisorId).toBe('Budi Santoso (Mgr)')
  })
})
