import { describe, expect, it } from 'vitest'
import { buildInitiativeRows, buildSoRows, formatTargetCell, templateYears } from '~/utils/strategicPlanTemplate'
import { buildStrategicPlanPayload, strategicPlanFormFromPlan } from '~/utils/strategicPlanPayload'
import { computeStrategicKpiMetrics } from '~/utils/mappedFeatureAdapters'

const plan = (id: string, code: string, extra: Record<string, unknown> = {}) => ({
  id, code, strategicObjective: `SO ${code}`, kpi: `KPI ${id}`, hibHig: 'HIG', yearStart: 2026,
  kpiTargets: { 2026: '75.5', 2027: '80' }, ...extra
}) as never

describe('strategic plan template helpers', () => {
  it('formats target cells with locale decimals and notes', () => {
    expect(formatTargetCell('75', '', 'id')).toBe('75')
    expect(formatTargetCell('75.5', 'PC', 'id')).toBe('75,5 (PC)')
    expect(formatTargetCell('75.5', 'PC', 'en')).toBe('75.5 (PC)')
    expect(formatTargetCell('', '', 'id')).toBe('-')
  })

  it('derives Year 1..5 from the data', () => {
    expect(templateYears([plan('a', 'X')])).toEqual([2026, 2027, 2028, 2029, 2030])
    expect(templateYears([])).toEqual([2026, 2027, 2028, 2029, 2030])
  })

  it('merges SO cells for KPIs sharing a code but keeps one row per id', () => {
    const rows = buildSoRows([plan('a', 'SO-IA04'), plan('b', 'SO-IA04'), plan('c', 'SO-IA04'), plan('d', 'SO-IA05')], [2026, 2027], 'id')
    expect(rows.map(r => r.id)).toEqual(['a', 'b', 'c', 'd'])
    expect(rows.map(r => r.span)).toEqual([3, 0, 0, 1])
    expect(rows[0]!.cells).toEqual(['75,5', '80'])
  })

  it('builds initiative rows only for plans with an initiative', () => {
    const rows = buildInitiativeRows([plan('a', 'X', { initiative: 'Prog', initiativeYears: { 2026: '√', 2028: '-' } }), plan('b', 'Y')], [2026, 2027, 2028])
    expect(rows).toHaveLength(1)
    expect(rows[0]!.cells).toEqual(['√', '-', '-'])
  })

  it('keeps initiative, initiativeYears and targetNotes through the payload helpers', () => {
    const p = plan('a', 'X', { initiative: 'Prog', initiativeYears: { 2026: '√' }, targetNotes: { 2029: 'PC' } })
    expect(buildStrategicPlanPayload(p)).toMatchObject({ initiative: 'Prog', initiativeYears: { 2026: '√' }, targetNotes: { 2029: 'PC' } })
    expect(strategicPlanFormFromPlan(p)).toMatchObject({ initiative: 'Prog', targetNotes: { 2029: 'PC' } })
  })

  it('scores HIG as actual/target and HIB as (1 + (1 - actual/target)) x 100', () => {
    expect(computeStrategicKpiMetrics('90', '100', 'HIG').calculation).toBe('90.00%')
    expect(computeStrategicKpiMetrics('90', '100', 'HIB').calculation).toBe('110.00%')
    expect(computeStrategicKpiMetrics('120', '100', 'HIB').calculation).toBe('80.00%')
  })
})
