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

// The team lead (Ketua Tim) is a per-row field now: the plan has no supervisor of its own.
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

describe('Annual Audit Form - Team lead (Ketua Tim) input field', () => {
  let store: ReturnType<typeof useAnnualPlanStore>

  beforeEach(() => {
    global.$fetch = vi.fn(async () => ({ success: true, data: { items: [], pagination: { total_pages: 1 } } }))
    setActivePinia(createPinia())
    store = useAnnualPlanStore()
  })

  it('renders the team lead field as UInput and not USelectMenu in AnnualAuditForm.vue', () => {
    const src = readFileSync(FORM_SFC, 'utf-8')
    const leadFieldMatch = src.match(/<UFormField[^>]*?:label="t\('annualAudit\.form\.teamLead'\)"[\s\S]*?<\/UFormField>/)
    expect(leadFieldMatch, 'team lead UFormField not found').not.toBeNull()

    const leadBlock = leadFieldMatch![0]
    expect(leadBlock).toContain('<UInput')
    expect(leadBlock).toContain('v-model="activity.supervisorName"')
    expect(leadBlock).not.toContain('<USelectMenu')
  })

  it('has no plan-level supervisor field left in the form', () => {
    const src = readFileSync(FORM_SFC, 'utf-8')
    expect(src).not.toContain('store.form.supervisorId')
    expect(src).not.toContain('annualAudit.form.supervisor\'')
  })

  it('validates the team lead of each row: fails if empty or whitespace', async () => {
    store.form.activities = [validRow({ supervisorName: '   ' })]
    store.form.year = '2026'
    store.form.selectedMonths = [0]

    await store.handleSubmit()

    expect(store.validationErrors.auditor).not.toBe('')
    const postCall = global.$fetch.mock.calls.find(([, opts]) => opts?.method === 'POST')
    expect(postCall).toBeUndefined()
  })

  it('submits with the team lead on each row and no plan-level supervisor', async () => {
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

  it('keeps each row\'s team lead when editing an existing plan', () => {
    store.handleEdit({
      id: 'plan-123',
      code: 'PKAT-2026-ASR-001',
      version: 'v1.0',
      status: 'Approved',
      selectedMonths: [0, 1],
      isActive: true,
      activities: [{ ...validRow({ supervisorName: 'Budi Santoso (Mgr)' }), id: 'row-1' }]
    })

    expect(store.form.activities[0].supervisorName).toBe('Budi Santoso (Mgr)')
    expect(store.form.activities[0].id).toBe('row-1')
  })
})
