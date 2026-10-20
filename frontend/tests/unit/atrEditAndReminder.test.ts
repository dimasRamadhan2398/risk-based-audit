/**
 * ATR LHA fields merged into the detail, the edit form's partial update (PUT /:id), the reminder action
 * (POST /:id/remind) and the rules that hide the reminder on a cancelled ATR.
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '~/stores/auth'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useAtrActions } from '~/composables/useAtrActions'
import { atrCanRemindStatus, normalizeAtrItem } from '~/utils/actionTakenReport'

const showSuccess = vi.fn()
const showError = vi.fn()
const showWarning = vi.fn()
vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({ showSuccess, showError, showWarning, showInfo: vi.fn(), add: vi.fn() })
}))

const BASE = '/api/v1/action-taken-reports'

/** Installs a $fetch mock (the Nuxt global is typed; the mock is not). */
const mockFetch = <T extends (...args: never[]) => unknown>(impl: T) => {
  const fn = vi.fn(impl)
  global.$fetch = fn as unknown as typeof $fetch
  return fn
}

const raw = (extra = {}) => ({
  id: 'atr-1',
  audit_result_report_id: 'lha-1',
  report_number: 'LHA-021/SKAI/2026',
  report_title: 'Audit Kas',
  finding_id: 'f-1',
  finding_title: 'Kas tidak direkonsiliasi',
  finding_category: 'Significant',
  recommendation: 'Rekonsiliasi bulanan',
  action_plan: 'Plan',
  pic_user_id: 'pic-1',
  pic_name: 'Rina',
  due_date: '2027-01-01T00:00:00Z',
  progress: 20,
  status: 'IN_PROGRESS',
  is_overdue: false,
  overdue_days: 0,
  evidence: [],
  review_note: '',
  reviewed_by: '',
  reviewed_at: null,
  created_at: '2026-10-01T00:00:00Z',
  updated_at: '2026-10-01T00:00:00Z',
  ...extra
})

const item = (extra = {}) => normalizeAtrItem(raw(extra))

const USERS = {
  pic: { id: 'pic-1', roles: ['AUDITEE'] },
  auditor: { id: 'aud-1', roles: ['AUDITOR'] }
}

const loginAs = (who: keyof typeof USERS) => {
  const auth = useAuthStore()
  auth.user = USERS[who] as unknown as typeof auth.user
}

describe('ATR normalisation: LHA fields', () => {
  it('keeps the LHA header fields of the detail (nested `lha` object)', () => {
    const it1 = item({
      lha: {
        id: 'lha-1',
        status: 'Final',
        report_date: '2026-09-30T00:00:00Z',
        department: 'Finance',
        executive_summary: 'Ringkasan',
        prepared_by: 'Budi',
        reviewed_by: 'Sari',
        approved_by: 'Dewi',
        findings: []
      }
    })
    expect(it1.lha).toMatchObject({
      status: 'Final',
      report_date: '2026-09-30T00:00:00Z',
      department: 'Finance',
      executive_summary: 'Ringkasan',
      prepared_by: 'Budi',
      reviewed_by: 'Sari',
      approved_by: 'Dewi',
      scope: ''
    })
    // the ATR's own reviewed_by is kept separate from the LHA's
    expect(it1.reviewed_by).toBe('')
  })

  it('is null when the response has no LHA (list rows, LHA gone)', () => {
    expect(item().lha).toBeNull()
    expect(item({ lha: null }).lha).toBeNull()
  })
})

describe('reminder and edit rules (useAtrActions)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('auditor can send a reminder on an open follow-up', () => {
    loginAs('auditor')
    expect(useAtrActions().canRemind(item({ status: 'IN_PROGRESS' }))).toBe(true)
    expect(useAtrActions().canRemind(item({ status: 'PENDING_REVIEW' }))).toBe(true)
  })

  it('a cancelled follow-up never gets a reminder, for anyone', () => {
    expect(atrCanRemindStatus('CANCELLED')).toBe(false)
    loginAs('auditor')
    expect(useAtrActions().canRemind(item({ status: 'CANCELLED' }))).toBe(false)
  })

  it('the PIC (auditee) cannot send reminders', () => {
    loginAs('pic')
    expect(useAtrActions().canRemind(item({ status: 'IN_PROGRESS' }))).toBe(false)
  })

  it('the server allowed_actions must list "remind" when it sends the list', () => {
    loginAs('auditor')
    expect(useAtrActions().canRemind(item({ status: 'IN_PROGRESS', allowed_actions: ['assign'] }))).toBe(false)
    expect(useAtrActions().canRemind(item({ status: 'IN_PROGRESS', allowed_actions: ['remind'] }))).toBe(true)
  })

  it('edit fields follow the server allowed_actions (assign, update_action_plan) when the list is present', () => {
    loginAs('auditor')
    expect(useAtrActions().editableFields(item({ status: 'IN_PROGRESS', allowed_actions: ['update_action_plan'] })))
      .toEqual({ dueDate: false, actionPlan: false, progress: false })
    expect(useAtrActions().canEdit(item({ status: 'IN_PROGRESS', allowed_actions: [] }))).toBe(false)
    expect(useAtrActions().canEdit(item({ status: 'IN_PROGRESS', allowed_actions: ['assign'] }))).toBe(true)
  })

  it('PIC edits the action plan and progress only; auditor edits the due date only', () => {
    loginAs('pic')
    const pic = useAtrActions()
    expect(pic.editableFields(item({ status: 'IN_PROGRESS' }))).toEqual({ dueDate: false, actionPlan: true, progress: true })
    expect(pic.canEdit(item({ status: 'IN_PROGRESS' }))).toBe(true)
    loginAs('auditor')
    const auditor = useAtrActions()
    expect(auditor.editableFields(item({ status: 'PENDING_REVIEW' }))).toEqual({ dueDate: true, actionPlan: false, progress: false })
  })

  it('nothing is editable on a completed or cancelled follow-up', () => {
    loginAs('auditor')
    const { editableFields, canEdit } = useAtrActions()
    expect(canEdit(item({ status: 'COMPLETED' }))).toBe(false)
    expect(canEdit(item({ status: 'CANCELLED' }))).toBe(false)
    expect(editableFields(item({ status: 'CANCELLED' }))).toEqual({ dueDate: false, actionPlan: false, progress: false })
  })
})

describe('ATR store: reminder and partial update', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
    loginAs('auditor')
  })
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('sendReminder POSTs /:id/reminder, toasts the recipients and returns true', async () => {
    const fetchMock = mockFetch(async () => ({
      success: true,
      data: {
        sent_to: [{ role: 'auditee', name: 'Rina', email: 'rina@example.test' }, { role: 'department_head', name: 'Budi', email: 'budi@example.test' }],
        skipped: [],
        failed: []
      }
    }))
    const store = useActionTakenReportStore()
    expect(await store.sendReminder('atr-1')).toBe(true)
    expect(fetchMock).toHaveBeenCalledWith(`${BASE}/atr-1/reminder`, expect.objectContaining({ method: 'POST' }))
    expect(showSuccess).toHaveBeenCalledWith(expect.stringContaining('Rina, Budi'))
    expect(showError).not.toHaveBeenCalled()
    expect(showWarning).not.toHaveBeenCalled()
  })

  it('sendReminder warns about a role that got no e-mail', async () => {
    mockFetch(async () => ({
      success: true,
      data: { sent_to: [{ role: 'auditee', name: 'Rina', email: 'rina@example.test' }], skipped: [{ role: 'department_head', reason: 'no active department head' }], failed: [] }
    }))
    const store = useActionTakenReportStore()
    expect(await store.sendReminder('atr-1')).toBe(true)
    expect(showWarning).toHaveBeenCalled()
  })

  it('sendReminder toasts the failure and returns false (no rethrow)', async () => {
    mockFetch(async () => {
      throw Object.assign(new Error('boom'), { statusCode: 500 })
    })
    const store = useActionTakenReportStore()
    expect(await store.sendReminder('atr-1')).toBe(false)
    expect(showError).toHaveBeenCalled()
    expect(showSuccess).not.toHaveBeenCalled()
  })

  it('updateReport sends only the changed fields, due date as RFC3339 midnight UTC', async () => {
    const fetchMock = mockFetch(async () => ({ success: true, data: raw({ progress: 40 }) }))
    const store = useActionTakenReportStore()
    expect(await store.updateReport('atr-1', { progress: 40, due_date: '2027-02-15' })).toBe(true)
    const [url, opts] = fetchMock.mock.calls[0] as unknown as [string, { method: string, body: unknown }]
    expect(url).toBe(`${BASE}/atr-1`)
    expect(opts.method).toBe('PUT')
    expect(opts.body).toEqual({ progress: 40, due_date: '2027-02-15T00:00:00Z' })
  })
})
