// @ts-nocheck
/**
 * Add Annual Audit Plan modal, Timeline section: the Q1 workload validation.
 *
 * Issue: the error "Beban kerja Triwulan I terlalu tinggi (>40%). Mohon ratakan
 * jadwal." shown when Triwulan 1 (Q1) months were selected had to be removed.
 *
 * The rule was a frontend-only `quarterAlert` computed in stores/annual-audit.ts
 * (> 3 months selected and > 40% of them in Jan-Mar), rendered as a red box in
 * components/annual-audit/AnnualAuditForm.vue and used to disable the Save button.
 * It has been removed; the backend never had an equivalent rule (see
 * backend/audit-service/models/audit-annual-plan_test.go).
 *
 * These tests guard against the rule coming back, and check that the remaining
 * checks (year / at-least-one-month in validateForm, the utilization overload on
 * the Save button) still work. validateForm is not exported from the store, so it
 * is driven through handleSubmit() with a mocked $fetch.
 */
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

// Month indexes are 0-based: 0 = Jan ... 11 = Dec.
const [JAN, FEB, MAR, APR, MAY, , JUL, , , OCT] = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]

let store: ReturnType<typeof useAnnualPlanStore>

/** Fill every non-timeline field so validateForm() only has timeline rules left to fail. */
function fillValidForm(months: number[], year = '2026') {
  store.form.code = 'PKAT-2026-ASR-001'
  store.form.activities = [
    { name: 'Audit IT', category: 'Assurance', department: 'IT', riskName: 'Risk A' }
  ]
  store.form.supervisorId = 'S01'
  store.form.auditorCount = 2
  store.form.daysPerAuditor = 5
  store.form.year = year
  store.form.selectedMonths = [...months]
}

/** The Annual Audit Form template (everything before <script setup>). */
function readTemplate(): string {
  const src = readFileSync(FORM_SFC, 'utf-8')
  return src.slice(0, src.indexOf('<script'))
}

/** The `:disabled` expression of the type="submit" (Save/Update Plan) button. */
function saveButtonDisabledExpr(): string {
  const template = readTemplate()
  const button = template.match(/<UButton[^>]*?type="submit"[\s\S]*?\/>/)
  expect(button, 'submit UButton not found in AnnualAuditForm.vue').not.toBeNull()
  const disabled = button![0].match(/:disabled="([^"]*)"/)
  expect(disabled, 'submit UButton has no :disabled binding').not.toBeNull()
  return disabled![1]
}

/** Evaluate the real `:disabled` expression from the SFC against the real store. */
function isSaveDisabled(): boolean {
  // eslint-disable-next-line no-new-func
  return Boolean(new Function('store', `return (${saveButtonDisabledExpr()})`)(store))
}

beforeEach(() => {
  global.$fetch = vi.fn(async () => ({ success: true, data: { items: [], pagination: { total_pages: 1 } } }))
  setActivePinia(createPinia())
  store = useAnnualPlanStore()
})

describe('Annual plan form after removing the Q1 workload rule', () => {
  it('the store no longer exposes quarterAlert', () => {
    store.form.selectedMonths = [JAN, FEB, MAR, APR]
    expect('quarterAlert' in store).toBe(false)
  })

  it.each(['2026', '2027'])('a Q1-heavy plan for %s passes validation and is POSTed', async (year) => {
    fillValidForm([JAN, FEB, MAR, APR], year)

    await store.handleSubmit()

    expect(store.validationErrors.timeline).toBe('')
    expect(store.validationErrors.activityDetail).toBe('')
    expect(store.validationErrors.auditor).toBe('')
    const post = global.$fetch.mock.calls.find(([, opts]) => opts?.method === 'POST')
    expect(post, 'POST /annual-audit-plans was not called').toBeDefined()
    expect(post[1].body.selectedMonths).toEqual([JAN, FEB, MAR, APR])
  })

  it('toggleMonth can select all of Q1 plus more without any Q1 error', () => {
    ;[JAN, FEB, MAR, APR, JUL].forEach(m => store.toggleMonth(m))
    expect(store.form.selectedMonths).toEqual([JAN, FEB, MAR, APR, JUL])
    expect(isSaveDisabled()).toBe(false)
  })

  it('validateForm still has the year and at-least-one-month timeline rules', async () => {
    fillValidForm([], '')

    await store.handleSubmit()

    expect(store.validationErrors.timeline).toBe('Year wajib dipilih. Minimal 1 bulan pelaksanaan wajib dipilih.')
    expect(global.$fetch).not.toHaveBeenCalled()
  })

  it('the utilization overload (red) still disables Save', () => {
    store.form.selectedMonths = [JAN, FEB, MAR, APR]
    store.form.auditorCount = 1000
    store.form.daysPerAuditor = 1000
    expect(store.utilizationData.color).toBe('red')
    expect(isSaveDisabled()).toBe(true)
  })
})

describe('Q1 workload rule stays removed', () => {
  it.each([
    ['Jan-Apr', [JAN, FEB, MAR, APR]],
    ['Jan, Feb, Jul, Oct', [JAN, FEB, JUL, OCT]],
    ['Jan-Mar plus Jul', [JAN, FEB, MAR, JUL]]
  ])('a Q1-heavy selection (%s) produces no Q1 workload error from the store', (_label, months) => {
    store.form.selectedMonths = months
    expect(store.quarterAlert).toBeFalsy()
  })

  it('the store never emits the Triwulan I message, whatever the selection', () => {
    store.form.selectedMonths = [JAN, FEB, MAR, APR]
    expect(String(store.quarterAlert ?? '')).not.toContain('Triwulan I terlalu tinggi')
  })

  it('the Save button :disabled expression in AnnualAuditForm.vue no longer references quarterAlert', () => {
    expect(saveButtonDisabledExpr()).not.toContain('quarterAlert')
  })

  it('a Q1-heavy selection does not disable Save (SFC expression evaluated against the store)', () => {
    store.form.auditorCount = 2
    store.form.daysPerAuditor = 5 // utilization green
    store.form.selectedMonths = [JAN, FEB, MAR, APR]
    expect(store.utilizationData.color).toBe('green')
    expect(isSaveDisabled()).toBe(false)
  })

  it('the form no longer renders the Q1 workload message', () => {
    const template = readTemplate()
    expect(template.includes('quarterAlert'), 'template still references quarterAlert').toBe(false)
    expect(template.includes('Beban kerja Triwulan'), 'template still contains the Q1 message').toBe(false)
    expect(template.includes('terlalu tinggi'), 'template still contains "terlalu tinggi"').toBe(false)
  })

  it('editing an existing Q1-heavy plan (selectedMonths [0,1,2,3]) leaves Save enabled', () => {
    store.handleEdit({
      id: 'plan-1',
      code: 'PKAT-2026-ASR-001',
      status: 'Not Available',
      year: '2026',
      activities: [{ name: 'Audit IT', category: 'Assurance', department: 'IT', riskName: 'Risk A' }],
      selectedMonths: [0, 1, 2, 3],
      auditorCount: 2,
      daysPerAuditor: 5,
      supervisorId: 'S01',
      isActive: true
    })

    expect(store.isEditing).toBe(true)
    expect(store.showModal).toBe(true)
    expect(store.form.selectedMonths).toEqual([0, 1, 2, 3])
    expect(store.utilizationData.color).toBe('green') // keep the overload check out of the picture
    expect(isSaveDisabled()).toBe(false)
  })
})
