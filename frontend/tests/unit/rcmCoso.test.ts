// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import {
  RCM_FILTER_ALL,
  RCM_FILTER_UNASSIGNED,
  matchesRcmCoso,
  matchesRcmDepartment,
  parseCosoCatalogue,
  parseRcmFieldErrors,
  selectablePrinciples,
  validateCosoMapping
} from '~/utils/rcmCoso'
import { parseDrift, parseSummary } from '~/utils/icAssessmentParse'
import { useRCMStore } from '~/stores/rcm'
import { rcmFormFromItem, createEmptyRcmForm } from '~/utils/rcmForm'

const catalogue = parseCosoCatalogue({
  success: true,
  data: {
    framework: 'COSO_2013',
    components: [
      { code: 'CE', sort_order: 1, name_en: 'Control Environment', principles: [
        { code: 'P01', number: 1, component_code: 'CE', name_en: 'Integrity', is_active: true },
        { code: 'P02', number: 2, component_code: 'CE', name_en: 'Oversight', is_active: false }
      ] },
      { code: 'CA', sort_order: 3, name_en: 'Control Activities', principles: [
        { code: 'P10', number: 10, component_code: 'CA', name_en: 'Selects controls', is_active: true }
      ] }
    ]
  }
})

describe('RCM COSO mapping helpers', () => {
  it('parses the catalogue and tolerates garbage', () => {
    expect(catalogue.map(c => c.code)).toEqual(['CE', 'CA'])
    expect(parseCosoCatalogue(null)).toEqual([])
    expect(parseCosoCatalogue({ data: {} })).toEqual([])
  })

  it('lists active principles, keeps inactive ones already tagged, and narrows by component', () => {
    expect(selectablePrinciples(catalogue).map(p => p.code)).toEqual(['P01', 'P10'])
    expect(selectablePrinciples(catalogue, { keep: ['P02'] }).map(p => p.code)).toEqual(['P01', 'P02', 'P10'])
    expect(selectablePrinciples(catalogue, { componentCode: 'CA' }).map(p => p.code)).toEqual(['P10'])
  })

  it('requires the component to come from a selected principle', () => {
    expect(validateCosoMapping(catalogue, null, [])).toBeNull()
    expect(validateCosoMapping(catalogue, 'CE', [])).toBe('componentWithoutPrinciples')
    expect(validateCosoMapping(catalogue, 'CE', ['P10'])).toBe('componentMismatch')
    expect(validateCosoMapping(catalogue, 'CE', ['P10', 'P01'])).toBeNull()
    // catalogue not loaded: the backend decides
    expect(validateCosoMapping([], 'CE', ['P10'])).toBeNull()
  })

  it('filters by master department, including unassigned legacy rows', () => {
    const legacy = {}
    const linked = { department_id: 'd1' }
    expect(matchesRcmDepartment(legacy, RCM_FILTER_ALL)).toBe(true)
    expect(matchesRcmDepartment(legacy, RCM_FILTER_UNASSIGNED)).toBe(true)
    expect(matchesRcmDepartment(linked, RCM_FILTER_UNASSIGNED)).toBe(false)
    expect(matchesRcmDepartment(linked, 'd1')).toBe(true)
    expect(matchesRcmDepartment(legacy, 'd1')).toBe(false)
  })

  it('filters by component and principle, and keeps untagged rows when unfiltered', () => {
    const row = { coso_component_code: 'CA', coso_principle_codes: ['P10'] }
    expect(matchesRcmCoso({}, catalogue, RCM_FILTER_ALL, RCM_FILTER_ALL)).toBe(true)
    expect(matchesRcmCoso(row, catalogue, 'CA', RCM_FILTER_ALL)).toBe(true)
    expect(matchesRcmCoso(row, catalogue, 'CE', RCM_FILTER_ALL)).toBe(false)
    expect(matchesRcmCoso(row, catalogue, RCM_FILTER_ALL, 'P10')).toBe(true)
    expect(matchesRcmCoso(row, catalogue, RCM_FILTER_ALL, 'P01')).toBe(false)
    expect(matchesRcmCoso({}, catalogue, 'CA', RCM_FILTER_ALL)).toBe(false)
  })

  it('reads 422 per-field messages', () => {
    const err = { data: { error: { code: 'VALIDATION_ERROR', fields: { coso_component_code: 'bad', x: 1 } } } }
    expect(parseRcmFieldErrors(err)).toEqual({ coso_component_code: 'bad' })
    expect(parseRcmFieldErrors(new Error('boom'))).toEqual({})
  })

  it('form from a legacy row has no department and no tags', () => {
    const form = rcmFormFromItem({ id: 'a', risk_code: 'R', department: 'Head Office' } as any)
    expect(form.department_id).toBeNull()
    expect(form.location_id).toBeNull()
    expect(form.coso_principle_codes).toEqual([])
    expect(createEmptyRcmForm(2026).coso_component_code).toBeNull()
  })
})

describe('IC summary and drift wire format', () => {
  it('parses the summary and zero-fills missing components', () => {
    const s = parseSummary({ success: true, data: { year: 2026, quarter: 0, total: 2, by_component: { CE: { YES: 2, PARTIAL: 0, NO: 0 } }, by_conclusion: { EFFECTIVE: 1, NOT_EFFECTIVE: 1 } } })
    expect(s?.total).toBe(2)
    expect(s?.byComponent.CE.YES).toBe(2)
    expect(s?.byComponent.MA).toEqual({ YES: 0, PARTIAL: 0, NO: 0 })
    expect(s?.byConclusion.NOT_EFFECTIVE).toBe(1)
    expect(parseSummary(null)).toBeNull()
  })

  it('parses drift signals from frozen/current', () => {
    const d = parseDrift({ data: { changed: true, complete: true, checked_at: '2026-01-01T00:00:00Z', signals: [
      { signal_key: 'ca.rcm_effectiveness', component_code: 'CA', changed: true, frozen: { state: 'OK', value_pct: 80 }, current: { state: 'OK', value_pct: 70 } }
    ] } })
    expect(d?.changed).toBe(true)
    expect(d?.signals[0]).toMatchObject({ signalKey: 'ca.rcm_effectiveness', changed: true, approvedValuePct: 80, currentValuePct: 70, approvedState: 'OK' })
    expect(parseDrift({ data: {} })).toBeNull()
  })
})

describe('RCM store COSO fields', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    global.$fetch = vi.fn()
  })

  it('sends department and tags, and clears the component when no principle is left', async () => {
    const store = useRCMStore()
    vi.mocked($fetch).mockResolvedValueOnce({ data: { id: 'x' } })
    await store.addRCMItem({
      risk_code: 'R', risk_event: 'E', control_code: 'C', control_description: 'D', control_owner: 'O', department: 'HO', year: 2026,
      department_id: '550e8400-e29b-41d4-a716-446655440000', coso_component_code: 'CA', coso_principle_codes: [],
      design_effectiveness_rating: 3
    } as any)
    const post = vi.mocked($fetch).mock.calls.find(c => (c[1] as any)?.method === 'POST')
    const body = (post![1] as any).body
    expect(body.department_id).toBe('550e8400-e29b-41d4-a716-446655440000')
    expect(body.coso_principle_codes).toEqual([])
    expect(body.coso_component_code).toBeNull()
    // no own branch picked: null = follow the linked risk
    expect(body.location_id).toBeNull()
  })

  it('sends the own branch (location_id) when picked', async () => {
    const store = useRCMStore()
    vi.mocked($fetch).mockResolvedValueOnce({ data: { id: 'x' } })
    await store.addRCMItem({
      risk_code: 'R', risk_event: 'E', control_code: 'C', control_description: 'D', control_owner: 'O', department: 'HO', year: 2026,
      location_id: '6ba7b810-9dad-11d1-80b4-00c04fd430c8', coso_principle_codes: []
    } as any)
    const post = vi.mocked($fetch).mock.calls.find(c => (c[1] as any)?.method === 'POST')
    expect((post![1] as any).body.location_id).toBe('6ba7b810-9dad-11d1-80b4-00c04fd430c8')
  })

  it('keeps 422 field errors for the form', async () => {
    const store = useRCMStore()
    vi.mocked($fetch).mockRejectedValueOnce({ data: { error: { code: 'VALIDATION_ERROR', message: 'Invalid request', fields: { coso_component_code: 'mismatch' } } } })
    await expect(store.addRCMItem({ risk_code: 'R', control_code: 'C' } as any)).rejects.toBeTruthy()
    expect(store.fieldErrors.coso_component_code).toBe('mismatch')
  })

  it('fetches the catalogue once', async () => {
    const store = useRCMStore()
    vi.mocked($fetch).mockResolvedValue({ data: { components: [{ code: 'CE', principles: [] }] } })
    await store.fetchCosoCatalogue()
    await store.fetchCosoCatalogue()
    expect(vi.mocked($fetch).mock.calls.filter(c => String(c[0]).endsWith('/rcm/coso-principles'))).toHaveLength(1)
    expect(store.cosoCatalogue).toHaveLength(1)
  })
})
