// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { useAuditUniverseStore } from '~/stores/audit-universe'

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn()
  })
}))

// What GET /annual-audit-plans returns after the Kafka refresh: snapshots filled by the server.
const serverPlan = {
  id: 'plan-1',
  code: 'PKAT-2026-001',
  title: 'Rencana Audit 2026',
  activityType: 'Assurance',
  status: 'Draft',
  year: '2026',
  selectedMonths: [0, 1],
  isActive: true,
  activities: [{
    id: 'row-1',
    activityCode: 'ASR-2026-001',
    name: 'Keuangan Cabang',
    category: 'Assurance',
    departmentId: 'dept-fin',
    department: 'Finance (refreshed)',
    auditUniverseId: 'au-fin',
    riskName: 'Risiko Kas (refreshed)',
    riskLevel: 'Moderate',
    auditorCount: 2,
    totalMandays: 12,
    supervisorName: 'Budi'
  }]
}

const pagedResponse = items => ({ success: true, data: { items, pagination: { total_pages: 1 } } })

describe('Annual plan store: server snapshots and partial PUT bodies', () => {
  let store: ReturnType<typeof useAnnualPlanStore>
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    fetchMock = vi.fn(async (url: string, opts?: any) => {
      if (opts?.method === 'PUT' || opts?.method === 'POST') return { success: true, data: {} }
      return pagedResponse([serverPlan])
    })
    global.$fetch = fetchMock
    setActivePinia(createPinia())
    store = useAnnualPlanStore()
  })

  it('keeps the server snapshots of riskName, riskLevel and department as they arrive', async () => {
    await store.fetchPlans()

    const row = store.plans[0].activities[0]
    expect(row.riskName).toBe('Risiko Kas (refreshed)')
    expect(row.riskLevel).toBe('Moderate')
    expect(row.department).toBe('Finance (refreshed)')
    expect(row.departmentId).toBe('dept-fin')
    expect(row.auditUniverseId).toBe('au-fin')
  })

  it('shows the snapshot names when the server sent them', async () => {
    await store.fetchPlans()

    const row = store.plans[0].activities[0]
    expect(store.activityDepartmentName(row)).toBe('Finance (refreshed)')
    expect(store.activityRiskName(row)).toBe('Risiko Kas (refreshed)')
  })

  it('resolves blank snapshots from departmentId and auditUniverseId, without changing the row', async () => {
    // What the list returns while the server lookup has no department / entity names yet.
    const blankRow = { ...serverPlan.activities[0], department: '', riskName: '', departmentId: 'DEPT-FIN', auditUniverseId: 'au-tax' }
    fetchMock.mockImplementation(async (url: string, opts?: { method?: string }) => {
      if (opts?.method === 'PUT' || opts?.method === 'POST') return { success: true, data: {} }
      return pagedResponse([{ ...serverPlan, activities: [blankRow] }])
    })
    await store.fetchPlans()
    const row = store.plans[0].activities[0]

    // Nothing to resolve against yet.
    expect(store.activityDepartmentName(row)).toBe('')
    expect(store.activityRiskName(row)).toBe('')

    store.departments = [{ id: 'dept-fin', name: 'Finance' }]
    useAuditUniverseStore().corporateUniverse = [
      { id: 'au-fin', name: 'Keuangan', children: [{ id: 'au-tax', name: 'Pajak', children: [] }] }
    ]

    expect(store.activityDepartmentName(row)).toBe('Finance')
    expect(store.activityRiskName(row)).toBe('Pajak')
    expect(row.department).toBe('')
    expect(row.riskName).toBe('')

    // The department filter matches the resolved name.
    store.selectedDepartment = 'Finance'
    expect(store.filteredPlans).toHaveLength(1)
  })

  it('an edit PUT keeps the row id and does not resend the snapshots', async () => {
    await store.fetchPlans()
    store.handleEdit(store.plans[0])
    store.form.activities[0].name = 'Keuangan Cabang Revisi'

    await store.handleSubmit()

    const putCall = fetchMock.mock.calls.find(([, opts]) => opts?.method === 'PUT')
    expect(putCall).toBeDefined()
    const [row] = putCall[1].body.activities
    expect(row.id).toBe('row-1')
    expect(row.name).toBe('Keuangan Cabang Revisi')
    expect(row).not.toHaveProperty('riskName')
    expect(row).not.toHaveProperty('riskLevel')
    expect(row).not.toHaveProperty('department')
  })

  it('a revision PUT changes only the version, status and history', async () => {
    await store.fetchPlans()

    await store.createRevision('plan-1', 'Revisi karena perubahan ruang lingkup', 'Auditor')

    const putCall = fetchMock.mock.calls.find(([, opts]) => opts?.method === 'PUT')
    expect(Object.keys(putCall[1].body).sort()).toEqual(['revisionHistory', 'status', 'version'])
  })

  it('the utilization counts the rows of the other plans and not the plan being edited', async () => {
    await store.fetchPlans()
    store.handleEdit(store.plans[0])

    // Only the rows of the form count: the stored 12 mandays of this plan are replaced, not added.
    expect(store.totalMandays).toBe(12)
    store.form.activities[0].totalMandays = 20
    expect(store.totalMandays).toBe(20)
  })
})
