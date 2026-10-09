/**
 * Source-level guards for UI fixes that have no behavioural test and were lost
 * once already through merges/rebases of staging. Each block names the screen.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const read = (path: string) => readFileSync(resolve(__dirname, '../../', path), 'utf8')

describe('Dashboard: Registered Risks table and Risk Heat Map card', () => {
  const src = read('pages/dashboard/index.vue')
  const columns = src.slice(src.indexOf('const registeredRiskColumns'), src.indexOf('// Audit Coverage'))

  it('has Risk Name, Score (impact x likelihood) and Risk Level columns, and no ID column', () => {
    expect(columns).toMatch(/accessorKey: "name"/)
    expect(columns).toMatch(/rawObject\.impact \* rawObject\.likelihood/)
    expect(columns).toMatch(/id: "level"/)
    expect(columns).toMatch(/getRiskLevel\(risk\.likelihood, risk\.impact\)/)
    expect(columns).not.toMatch(/accessorKey: "id"/)
  })

  it('keeps Risk Name at a fixed width, clamped to 2 lines with an overflow tooltip', () => {
    expect(columns).toMatch(/w-64 min-w-64 max-w-64/)
    expect(columns).toMatch(/line-clamp-2/)
    expect(columns).toMatch(/h\(OverflowTooltip/)
  })

  it('stretches the heat map card to the table height, with the narrow Probability label box', () => {
    expect(src).toMatch(/lg:grid-cols-12[^"]*items-stretch/)
    expect(src).toMatch(/lg:col-span-5[^"]*h-full[^"]*flex-col/)
    expect(src).toMatch(/flex-1 flex flex-row[^"]*items-center/)
    expect(src).toMatch(/shrink-0 w-5">\s*<span[^>]*>\s*Probability/)
  })

  it('loads the cards from the stores instead of seed data', () => {
    expect(src).toMatch(/annualPlanStore\.fetchPlans\(\)/)
    expect(src).toMatch(/auditResultStore\.fetchRecentFindings\(5\)/)
    expect(src).toMatch(/rcmStore\.fetchRCMList\(\)/)
    expect(src).toMatch(/remainingAudits: plannedAudits - completedAudits/)
  })
})

describe('OverflowTooltip', () => {
  it('only enables the tooltip when the text is clipped', () => {
    const src = read('components/shared/OverflowTooltip.vue')
    expect(src).toMatch(/:disabled="!isOverflowing"/)
    expect(src).toMatch(/scrollWidth > node\.clientWidth \|\| node\.scrollHeight > node\.clientHeight/)
  })
})

describe('TableEntities slot columns', () => {
  it('does not generate a cell slot for columns with their own cell renderer', () => {
    const src = read('components/shared/TableEntities.vue')
    expect(src).toMatch(/v-for="col in slotColumns"/)
    expect(src).toMatch(/typeof col\.cell !== 'function' \|\| !!\$slots\[`\$\{col\.id\}-cell`\]/)
  })
})

describe('KPI Detailed Breakdown table', () => {
  it('keeps the KPI Metric column at 130px / sm:220px and lets it wrap', () => {
    const src = read('components/kpi-performance/KpiDetailedTable.vue')
    expect(src).toMatch(/w-\[130px\] min-w-\[130px\] max-w-\[130px\] sm:w-\[220px\] sm:min-w-\[220px\] sm:max-w-\[220px\] whitespace-normal/)
    expect(src).toMatch(/perfStore\.kpiBreakdownFilters\.categories/)
    expect(src).toMatch(/perfStore\.kpiBreakdownFilters\.periods/)
  })

  it('only deletes strategic-plan rows, through the store, then reloads the page', () => {
    const src = read('components/kpi-performance/KpiDetailedTable.vue')
    expect(src).toMatch(/item\.source !== 'strategic_plan'/)
    expect(src).toMatch(/store\.handleDelete\(item\.id\)/)
    expect(src).toMatch(/perfStore\.fetchKpiBreakdown\(\)/)
  })
})

describe('Risk Control Matrix modal', () => {
  const src = read('pages/risk-profile/risk-control-matrix/index.vue')

  it('has a fixed header and a scrolling body', () => {
    expect(src).toMatch(/class="shrink-0 flex items-center justify-between[^"]*border-b/)
    expect(src).toMatch(/class="flex-1 min-h-0 overflow-y-auto/)
  })

  it('has no branch dropdown and no "(Sinkron)" labels', () => {
    expect(src).not.toMatch(/branchModalOptions|selectedBranchInModal/)
    expect(src).not.toMatch(/\(Sinkron\)/)
  })

  it('catches delete failures', () => {
    expect(src).toMatch(/try \{\s*await rcmStore\.deleteRCMItem\(id\)/)
  })
})

describe('Stores that must hold no seed data', () => {
  it.each([
    ['stores/annual-audit.ts', /initialMockPlans/],
    ['stores/rcm.ts', /initialRCMData/],
    ['stores/strategic-audit-plan.ts', /mockObjectives/],
    ['stores/audit-result-report.ts', /ST-001\/SKAI\/2026/]
  ])('%s', (path, pattern) => {
    expect(read(path)).not.toMatch(pattern)
  })
})

describe('Executive Summary (Individual) sections', () => {
  const src = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')

  it('numbers the headings I to V and titles section V', () => {
    for (const n of ['I', 'II', 'III', 'IV', 'V']) {
      expect(src).toContain(`<span class="text-primary-500">${n}.</span>`)
    }
    expect(src).not.toContain('<span class="text-primary-500">VI.</span>')
    expect(src).toMatch(/V\.<\/span> Analisis Temuan Berulang & Kesimpulan/)
  })

  it('labels nav tab 5 "Temuan Berulang & Kesimpulan"', () => {
    expect(src).toMatch(/index: '5', title: 'Temuan Berulang & Kesimpulan'/)
  })
})
