// Pure helpers for the Strategic Plan Internal Audit template view
// (Sasaran Strategis table + Program/Kegiatan Utama table).

import type { StrategicAuditPlan } from '~/types/audit'
import { formatDisplayNumber } from '~/utils/numberInput'

export const TEMPLATE_YEAR_COUNT = 5

/** Year 1..5 of the template: the earliest yearStart (or kpiTargets year) of the plans, else the fallback year. */
export const templateYears = (plans: ReadonlyArray<Partial<StrategicAuditPlan>>, fallbackStart = 2026): number[] => {
  let start = Infinity
  for (const p of plans) {
    if (p.yearStart && p.yearStart < start) start = p.yearStart
    for (const k of Object.keys(p.kpiTargets ?? {})) {
      const y = Number(k)
      if (Number.isInteger(y) && y < start) start = y
    }
  }
  const first = Number.isFinite(start) ? start : fallbackStart
  return Array.from({ length: TEMPLATE_YEAR_COUNT }, (_, i) => first + i)
}

/** A target cell: localised number plus the note in brackets, e.g. "75,0 (PC)"; "-" when empty and no note. */
export const formatTargetCell = (value: unknown, note: unknown, locale?: string): string => {
  const raw = String(value ?? '').trim()
  const text = raw === '' || raw === '-' ? '' : formatDisplayNumber(raw, locale)
  const n = String(note ?? '').trim()
  if (text && n) return `${text} (${n})`
  if (text) return text
  return n ? `(${n})` : '-'
}

export interface SoRow {
  id: string
  code: string
  objective: string
  kpi: string
  hibHig: string
  /** Rows to span in the No SO / Strategic Objective cells; 0 = covered by the row above. */
  span: number
  cells: string[]
}

/** Rows in the order given; consecutive rows sharing a code (and objective) merge their SO cells. Keyed by id, not code. */
export const buildSoRows = (plans: ReadonlyArray<StrategicAuditPlan>, years: number[], locale?: string): SoRow[] => {
  const rows: SoRow[] = plans.map(p => ({
    id: p.id,
    code: p.code || '',
    objective: p.strategicObjective || '',
    kpi: p.kpi || '',
    hibHig: p.hibHig || '',
    span: 1,
    cells: years.map(y => formatTargetCell(p.kpiTargets?.[y], p.targetNotes?.[y], locale))
  }))
  let head: SoRow | null = null
  for (const r of rows) {
    if (head && r.code && r.code === head.code && r.objective === head.objective) {
      head.span++
      r.span = 0
    } else {
      head = r
    }
  }
  return rows
}

export interface InitiativeRow { id: string, code: string, initiative: string, cells: string[] }

/** One row per plan that has an initiative; marks come straight from initiativeYears ("√" / "-"). */
export const buildInitiativeRows = (plans: ReadonlyArray<StrategicAuditPlan>, years: number[]): InitiativeRow[] =>
  plans
    .filter(p => (p.initiative ?? '').trim() !== '')
    .map(p => ({
      id: p.id,
      code: p.code || '',
      initiative: (p.initiative ?? '').trim(),
      cells: years.map(y => String(p.initiativeYears?.[y] ?? '').trim() || '-')
    }))
