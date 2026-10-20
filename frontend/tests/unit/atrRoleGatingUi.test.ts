// @ts-nocheck
/**
 * ATR UI: who sees which action (useRbac + status + the server's allowed_actions), the "My actions"
 * default for PICs, the PIC picker on 403, and source checks on the ATR page and modals
 * (TableEntities, shared modal shell, no assignment-letter fields, no mock data).
 */
import { describe, it, expect, beforeAll, afterAll, beforeEach, afterEach, vi } from 'vitest'
import { computed as vueComputed, createApp, defineComponent, h, nextTick } from 'vue'
import { readFileSync, readdirSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'
import en from '~/locales/en/common.json'
import { useAuthStore } from '~/stores/auth'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useAtrActions } from '~/composables/useAtrActions'
import { normalizeAtrItem } from '~/utils/actionTakenReport'
import AtrPage from '~/pages/action-taken-report/index.vue'
import ATRDetail from '~/components/action-taken-report/ATRDetail.vue'
import ATRAssignmentModal from '~/components/action-taken-report/ATRAssignmentModal.vue'

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({ showSuccess: vi.fn(), showError: vi.fn(), showWarning: vi.fn(), showInfo: vi.fn(), add: vi.fn() })
}))

// TableEntities stand-in: one <td data-col> per column, filled through the page's #<key>-cell slots.
vi.mock('~/components/shared/TableEntities.vue', async () => {
  const { defineComponent, h } = await import('vue')
  return {
    default: defineComponent({
      name: 'TableEntities',
      props: ['data', 'columns', 'loading', 'serverSide', 'total', 'itemsPerPage', 'page'],
      setup: (p, { slots }) => () => h('table', { 'data-stub': 'TableEntities' }, [
        h('thead', h('tr', p.columns.map(c => h('th', c.header)))),
        h('tbody', p.data.length
          ? p.data.map(row => h('tr', { 'data-row': row.id }, p.columns.map(c =>
            h('td', { 'data-col': c.accessorKey }, slots[`${c.accessorKey}-cell`]?.({ row: { original: row } }) ?? String(row[c.accessorKey] ?? ''))
          )))
          : h('tr', h('td', { 'data-empty': '' }, slots.empty?.())))
      ])
    })
  }
})

const A = en.actionTakenReport

// tests/setup.ts stubs the auto-imported `computed` with a plain { value } object, which a template does
// not unwrap (useRbac relies on the auto-import). Use Vue's real computed here, as in the app.
let setupComputed
beforeAll(() => {
  setupComputed = global.computed
  global.computed = vueComputed
})
afterAll(() => {
  global.computed = setupComputed
})

const item = (extra = {}) => normalizeAtrItem({
  id: 'atr-1',
  audit_result_report_id: 'lha-1',
  report_number: 'LHA-021/SKAI/2026',
  report_title: 'Audit Kas',
  finding_id: 'f-1',
  finding_title: 'Kas tidak direkonsiliasi',
  finding_category: 'Significant',
  recommendation: 'Rekonsiliasi bulanan',
  action_plan: '',
  pic_user_id: 'pic-1',
  pic_name: 'Rina',
  due_date: '2027-01-01T00:00:00Z',
  progress: 0,
  status: 'PLANNED',
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

const USERS = {
  pic: { id: 'pic-1', roles: ['AUDITEE'] },
  auditor: { id: 'aud-1', roles: ['AUDITOR'] },
  cae: { id: 'cae-1', roles: ['EXECUTIVE'] },
  admin: { id: 'adm-1', roles: ['ADMIN'] }
}

const loginAs = (who) => {
  useAuthStore().user = USERS[who]
}

describe('role gating (useAtrActions)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  const actionsFor = (who, extra) => {
    loginAs(who)
    return useAtrActions().availableActions(item(extra))
  }

  it('PIC (auditee): action plan on own PLANNED/IN_PROGRESS items, never assign/review/cancel', () => {
    expect(actionsFor('pic', { status: 'PLANNED' })).toEqual({ actionPlan: true, assign: false, review: false, cancel: false })
    expect(actionsFor('pic', { status: 'IN_PROGRESS' }).actionPlan).toBe(true)
    expect(actionsFor('pic', { status: 'PENDING_REVIEW' })).toEqual({ actionPlan: false, assign: false, review: false, cancel: false })
    // someone else's item
    expect(actionsFor('pic', { pic_user_id: 'other' }).actionPlan).toBe(false)
  })

  it('auditor: assign and review, but not the action plan form or cancel', () => {
    expect(actionsFor('auditor', { status: 'IN_PROGRESS' })).toEqual({ actionPlan: false, assign: true, review: false, cancel: false })
    expect(actionsFor('auditor', { status: 'PENDING_REVIEW' })).toEqual({ actionPlan: false, assign: true, review: true, cancel: false })
    expect(actionsFor('auditor', { status: 'COMPLETED' })).toEqual({ actionPlan: false, assign: false, review: false, cancel: false })
  })

  it('EXECUTIVE (CAE): review and cancel only, no assign', () => {
    expect(actionsFor('cae', { status: 'PENDING_REVIEW' })).toEqual({ actionPlan: false, assign: false, review: true, cancel: true })
    expect(actionsFor('cae', { status: 'CANCELLED' }).cancel).toBe(false)
  })

  it('admin: everything, including the action plan of an assigned item', () => {
    expect(actionsFor('admin', { status: 'IN_PROGRESS' })).toEqual({ actionPlan: true, assign: true, review: false, cancel: true })
    expect(actionsFor('admin', { status: 'PENDING_REVIEW' })).toMatchObject({ review: true, cancel: true })
  })

  it('a reviewer does not review an item they are the PIC of (segregation of duties)', () => {
    expect(actionsFor('auditor', { status: 'PENDING_REVIEW', pic_user_id: 'aud-1' }).review).toBe(false)
  })

  it('the server\'s allowed_actions can only narrow what the role allows', () => {
    expect(actionsFor('cae', { status: 'PENDING_REVIEW', allowed_actions: ['review'] }))
      .toEqual({ actionPlan: false, assign: false, review: true, cancel: false })
    // the server never widens: a PIC is not given review because the list says so
    expect(actionsFor('pic', { status: 'PENDING_REVIEW', allowed_actions: ['review', 'cancel'] }))
      .toEqual({ actionPlan: false, assign: false, review: false, cancel: false })
    expect(actionsFor('pic', { status: 'IN_PROGRESS', allowed_actions: ['update_action_plan', 'upload_evidence', 'submit'] }).actionPlan).toBe(true)
  })
})

// --- Mounted components ----------------------------------------------------------------------------

const slotStub = name => defineComponent({
  name,
  inheritAttrs: false,
  props: ['label', 'title', 'description', 'text', 'items'],
  setup: (p, { slots }) => () => h('div', { 'data-stub': name }, [
    p.label, p.title, p.description,
    ...(Array.isArray(p.items) ? p.items.map(i => h('span', { 'data-item': i.value }, i.label)) : []),
    slots.default?.(), slots.content?.()
  ])
})
const nullStub = name => defineComponent({ name, setup: () => () => h('span', { 'data-stub': name }) })
const ModalStub = defineComponent({
  name: 'UModal',
  props: ['open'],
  setup: (p, { slots }) => () => (p.open ? h('div', { 'data-stub': 'UModal' }, slots.content?.()) : null)
})
const ButtonStub = defineComponent({
  name: 'UButton',
  props: ['label', 'disabled', 'loading', 'to', 'icon'],
  setup: (p, { slots, attrs }) => () => h('button', {
    'data-stub': 'UButton',
    'disabled': p.disabled || undefined,
    'aria-label': attrs['aria-label'],
    'data-to': p.to,
    'onClick': attrs.onClick
  }, [p.label, slots.default?.()])
})

const flush = () => new Promise(r => setTimeout(r, 0))

/** GET list returns `rows`; /users/assignable answers with `users` (or the given error). */
const backend = ({ rows = [], users = [], usersError = null } = {}) => vi.fn(async (url, opts = {}) => {
  const method = (opts.method || 'GET').toUpperCase()
  if (url.endsWith('/users/assignable')) {
    if (usersError) throw usersError
    return { success: true, data: { users, pagination: { page: 1, page_size: 20, total: users.length, total_pages: 1 } } }
  }
  if (method === 'GET' && url.endsWith('/action-taken-reports')) {
    return { success: true, data: { items: rows, pagination: { page: 1, page_size: 10, total: rows.length, total_pages: rows.length ? 1 : 0 } } }
  }
  if (method === 'GET' && url.includes('/action-taken-reports/')) {
    const id = url.split('/').pop()
    return { success: true, data: rows.find(r => r.id === id) ?? item({ id }) }
  }
  if (url.endsWith('/audit-result-reports')) return { success: true, data: { items: [] } }
  return { success: true, data: null }
})

describe('mounted ATR UI', () => {
  let app
  let container

  beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    vi.spyOn(console, 'warn').mockImplementation(() => {})
  })
  afterEach(() => {
    app?.unmount()
    container?.remove()
    app = null
    vi.restoreAllMocks()
  })

  const mount = async (component, { who, fetch, prepare }) => {
    const pinia = createPinia()
    setActivePinia(pinia)
    global.$fetch = fetch
    loginAs(who)
    const store = useActionTakenReportStore()
    await prepare?.(store)
    container = document.createElement('div')
    document.body.appendChild(container)
    app = createApp({ render: () => h(component) })
    app.use(pinia)
    app.config.warnHandler = () => {}
    for (const name of ['UTooltip', 'UAlert', 'UBadge', 'UTabs', 'URadioGroup', 'UPopover']) app.component(name, slotStub(name))
    for (const name of ['UIcon', 'UProgress', 'UInput', 'UTextarea', 'USlider', 'UInputNumber', 'USwitch', 'USelectMenu']) app.component(name, nullStub(name))
    app.component('UModal', ModalStub)
    app.component('UButton', ButtonStub)
    app.mount(container)
    await flush()
    await nextTick()
    return store
  }

  const buttonLabels = () => Array.from(container.querySelectorAll('button')).map(b => (b.textContent || b.getAttribute('aria-label') || '').trim())

  const listCalls = () => global.$fetch.mock.calls.filter(([url, o = {}]) => url.endsWith('/action-taken-reports') && (o.method || 'GET') === 'GET')

  it('PIC: the page opens on "My actions" (pic_user_id = me) and offers the action plan, not review', async () => {
    const rows = [item({ status: 'IN_PROGRESS' })]
    await mount(AtrPage, { who: 'pic', fetch: backend({ rows }) })
    const [first] = listCalls()
    expect(first[1].params).toMatchObject({ page: 1, pic_user_id: 'pic-1' })
    expect(container.querySelector('[data-stub="UTabs"]')).toBeNull()
    expect(container.textContent).toContain(A.scope.mine)
    const actions = container.querySelector('td[data-col="actions"]')
    const labels = Array.from(actions.querySelectorAll('button')).map(b => b.getAttribute('aria-label'))
    expect(labels).toContain(A.actions.fillActionPlan)
    expect(labels).not.toContain(A.actions.review)
    expect(labels).not.toContain(A.actions.assign)
    // the PIC never asks for the user list
    expect(global.$fetch.mock.calls.some(([url]) => url.endsWith('/users/assignable'))).toBe(false)
  })

  it('auditor: the page opens on all follow-ups with the scope tabs; assign and review, no action plan', async () => {
    const rows = [item({ status: 'PENDING_REVIEW', action_plan: 'Done', progress: 100 })]
    await mount(AtrPage, { who: 'auditor', fetch: backend({ rows }) })
    const [first] = listCalls()
    expect(first[1].params).not.toHaveProperty('pic_user_id')
    expect(container.querySelector('[data-stub="UTabs"]')?.textContent).toContain(A.scope.mine)
    const labels = Array.from(container.querySelector('td[data-col="actions"]').querySelectorAll('button')).map(b => b.getAttribute('aria-label'))
    expect(labels).toEqual(expect.arrayContaining([A.actions.view, A.actions.assign, A.actions.review]))
    expect(labels).not.toContain(A.actions.fillActionPlan)
    expect(labels).not.toContain(A.actions.cancel)
  })

  it('the table shows finding + category, LHA, PIC, due date with overdue, progress and status', async () => {
    const rows = [item({ status: 'IN_PROGRESS', is_overdue: true, overdue_days: 4, progress: 40 })]
    await mount(AtrPage, { who: 'auditor', fetch: backend({ rows }) })
    const headers = Array.from(container.querySelectorAll('th')).map(th => th.textContent)
    expect(headers).toEqual([
      A.table.columns.finding, A.table.columns.lha, A.table.columns.pic, A.table.columns.dueDate,
      A.table.columns.progress, A.table.columns.status, A.table.columns.actions
    ])
    const cell = key => container.querySelector(`td[data-col="${key}"]`).textContent
    expect(cell('finding_title')).toContain('Kas tidak direkonsiliasi')
    expect(cell('finding_title')).toContain(en.actionTakenReport.findingCategory.significant)
    expect(cell('report_number')).toContain('LHA-021/SKAI/2026')
    expect(cell('pic_name')).toContain('Rina')
    expect(cell('due_date')).toContain('01 Jan 2027')
    expect(cell('due_date')).toContain(A.daysOverdue.replace('{days}', '4'))
    expect(cell('progress')).toContain('40%')
    expect(cell('status')).toContain(A.status.inProgress)
  })

  it('empty list: explains that ATRs are created when an LHA is approved', async () => {
    await mount(AtrPage, { who: 'auditor', fetch: backend() })
    expect(container.querySelector('[data-empty]').textContent).toContain(A.table.emptyHint)
  })

  it('detail as PIC: action plan button, no review/assign/cancel; no reminder button', async () => {
    const row = item({ status: 'IN_PROGRESS', action_plan: 'Plan', evidence: [{ id: 'e1', file_name: 'proof.pdf', uploaded_at: '2026-10-02T03:00:00Z', uploaded_by: 'rina' }] })
    await mount(ATRDetail, { who: 'pic', fetch: backend({ rows: [row] }), prepare: s => s.openDetail(row) })
    const labels = buttonLabels()
    expect(labels).toContain(A.actions.fillActionPlan)
    expect(labels).not.toContain(A.actions.review)
    expect(labels).not.toContain(A.actions.assign)
    expect(labels).not.toContain(A.actions.cancel)
    expect(container.textContent).not.toMatch(/Kirim Pengingat|reminder/i)
    expect(container.textContent).toContain('Rekonsiliasi bulanan')
    expect(container.textContent).toContain('proof.pdf')
    // a PIC filters the follow-ups of the LHA but does not get the LHA page link
    expect(labels).toContain(A.detail.showLhaFollowUps)
    expect(labels).not.toContain(A.detail.openLha)
  })

  it('detail as auditor on a submitted item: review and assign, no action plan', async () => {
    const row = item({ status: 'PENDING_REVIEW', action_plan: 'Plan' })
    await mount(ATRDetail, { who: 'auditor', fetch: backend({ rows: [row] }), prepare: s => s.openDetail(row) })
    const labels = buttonLabels()
    expect(labels).toEqual(expect.arrayContaining([A.actions.review, A.actions.reassign]))
    expect(labels).not.toContain(A.actions.fillActionPlan)
    const lhaLink = Array.from(container.querySelectorAll('button')).find(b => b.textContent.includes(A.detail.openLha))
    expect(lhaLink.getAttribute('data-to')).toBe('/audit-result-report?report=lha-1')
  })

  it('assignment modal: 403 on the user list shows "cannot list users" and no fallback list', async () => {
    const forbidden = Object.assign(new Error('Forbidden'), { statusCode: 403, status: 403 })
    const row = item({ pic_user_id: null, pic_name: '' })
    await mount(ATRAssignmentModal, { who: 'auditor', fetch: backend({ rows: [row], usersError: forbidden }), prepare: s => s.openAssignment(row) })
    await flush()
    expect(container.textContent).toContain(A.assignment.usersForbidden)
    const save = Array.from(container.querySelectorAll('button')).find(b => b.textContent.includes(A.assignment.save))
    expect(save.hasAttribute('disabled')).toBe(true)
  })

  it('assignment modal: searches GET /users/assignable and labels users "full_name — position, department"', async () => {
    const users = [{ id: 'u-9', full_name: 'Budi', department: 'Finance', position: 'Manager' }]
    const row = item({ pic_user_id: null, pic_name: '' })
    let picker
    const Capture = defineComponent({
      name: 'USelectMenu',
      props: ['items'],
      setup: (p) => {
        picker = p
        return () => h('span')
      }
    })
    const pinia = createPinia()
    setActivePinia(pinia)
    global.$fetch = backend({ rows: [row], users })
    loginAs('auditor')
    await useActionTakenReportStore().openAssignment(row)
    container = document.createElement('div')
    app = createApp({ render: () => h(ATRAssignmentModal) })
    app.use(pinia)
    app.config.warnHandler = () => {}
    app.component('UModal', ModalStub)
    app.component('UButton', ButtonStub)
    app.component('USelectMenu', Capture)
    for (const name of ['UTooltip', 'UAlert', 'UBadge', 'UPopover']) app.component(name, slotStub(name))
    app.component('UIcon', nullStub('UIcon'))
    app.mount(container)
    await flush()
    const call = global.$fetch.mock.calls.find(([url]) => url.endsWith('/users/assignable'))
    expect(call[1].params).toEqual({ search: undefined, page: 1, page_size: 20 })
    expect(picker.items).toEqual([{ label: 'Budi — Manager, Finance', value: 'u-9' }])
  })
})

// --- Source checks -----------------------------------------------------------------------------------

const ROOT = resolve(__dirname, '../..')
const read = path => readFileSync(resolve(ROOT, path), 'utf8')
const ATR_COMPONENTS = readdirSync(resolve(ROOT, 'components/action-taken-report')).map(f => `components/action-taken-report/${f}`)

describe('ATR source', () => {
  it('the ATR page uses the shared TableEntities with server-side paging', () => {
    const src = read('pages/action-taken-report/index.vue')
    expect(src).toContain("import TableEntities from '~/components/shared/TableEntities.vue'")
    expect(src).toMatch(/<TableEntities[\s\S]*:server-side="true"/)
    expect(src).not.toMatch(/<UTable\b/)
  })

  it('no assignment-letter fields, filters or columns in the ATR UI, store, utils or type', () => {
    const files = ['pages/action-taken-report/index.vue', 'stores/action-taken-report.ts', 'utils/actionTakenReport.ts', 'composables/useAtrActions.ts', ...ATR_COMPONENTS]
    for (const file of files) {
      expect(read(file), file).not.toMatch(/assignment[_\s-]?letter|surat[\s_-]?tugas|\bST-\d/i)
    }
    const type = read('types/audit.ts').match(/export interface ActionTakenReport \{[\s\S]*?\n\}/)[0]
    expect(type).not.toMatch(/assignment|auditRef|deadline/i)
    expect(JSON.stringify(en.actionTakenReport)).not.toMatch(/assignment letter|surat tugas/i)
  })

  it('no mock data and no hardcoded reminder text (the reminder calls the real endpoint, labels come from i18n)', () => {
    const files = ['pages/action-taken-report/index.vue', 'stores/action-taken-report.ts', ...ATR_COMPONENTS]
    for (const file of files) {
      const src = read(file)
      expect(src, file).not.toMatch(/Kirim Pengingat|mockData|dummy|sampleData/i)
    }
  })

  it('every ATR modal uses the shared modal shell; confirmations go through the global modal store', () => {
    for (const file of ['ATRDetail.vue', 'ATRActionPlanForm.vue', 'ATRAssignmentModal.vue', 'ATRReviewModal.vue', 'ATRCancelModal.vue']) {
      expect(read(`components/action-taken-report/${file}`), file).toContain('<AtrModalShell')
    }
    expect(read('components/action-taken-report/ATRActionPlanForm.vue')).toContain('useGlobalModalStore().confirmSubmit')
    // shell: fixed header/footer, only the body scrolls
    const shell = read('components/action-taken-report/AtrModalShell.vue')
    expect(shell).toContain('shrink-0')
    expect(shell).toContain('flex-1 min-h-0 overflow-y-auto')
  })
})
