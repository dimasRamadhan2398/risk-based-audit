// @ts-nocheck
/**
 * Dashboard "Risk Profiles" card > "Registered Risks" table vs the Corporate
 * Risk Profile (CRP) page (/risk-profile → components/risk-profile/RiskHeatMap.vue).
 *
 * Both read the same Pinia store (stores/risk-profile.ts → GET /risks), so the
 * underlying rows and the year/quarter mapping agree.
 *
 * Intended dashboard table (product decision):
 *  - no ID column;
 *  - Score = impact × likelihood (deliberately NOT the CRP matrix score);
 *  - Risk Level column = store.getRiskLevel(likelihood, impact), with the same
 *    label (riskProfile.riskLevelLabels.*) and colour (riskLevelConfig) as CRP.
 *
 * Still open (out of scope, the last three tests fail on purpose): the
 * dashboard takes the first five rows in API order (risk-service FindAll has
 * no ORDER BY), while the CRP lists risks sorted by score (Priority tab).
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'
import { useRiskProfileStore, riskLevelConfig } from '~/stores/risk-profile'
import { RiskLevel } from '~/types/risk'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: { masterServiceBaseUrl: 'http://localhost:8080/api/v1' }
  })
}))

const DASHBOARD = resolve(__dirname, '../../pages/dashboard/index.vue')
const EN = resolve(__dirname, '../../locales/en/common.json')
const ID = resolve(__dirname, '../../locales/id/common.json')

const okResponse = (data: any) => ({ success: true, message: 'ok', data })

const assessment = (impact: number, likelihood: number) => ({
  id: `ast-${impact}${likelihood}`,
  year: 2026,
  impact_q1: impact, impact_q2: impact, impact_q3: impact, impact_q4: impact,
  likelihood_q1: likelihood, likelihood_q2: likelihood, likelihood_q3: likelihood, likelihood_q4: likelihood
})

// GET /risks as risk-service returns it: unordered (no ORDER BY in FindAll),
// severity = inherent_score (the CRP form always sends 50 for new risks).
const backendRisks = [
  { id: 'r-low', name: 'Office supplies shortage', impact: 1, likelihood: 2, severity: 50, category: 'Operations', branch: 'Head Office', assessments: [assessment(1, 2)] },
  { id: 'r-mod', name: 'Vendor delay', impact: 3, likelihood: 3, severity: 50, category: 'Operations', branch: 'Head Office', assessments: [assessment(3, 3)] },
  { id: 'r-lowmod', name: 'Minor policy gap', impact: 2, likelihood: 2, severity: 50, category: 'Compliance', branch: 'Bali Branch', assessments: [assessment(2, 2)] },
  { id: 'r-low2', name: 'Printer downtime', impact: 1, likelihood: 1, severity: 50, category: 'Technology', branch: 'Head Office', assessments: [assessment(1, 1)] },
  { id: 'r-fin', name: 'Revenue target missed', impact: 5, likelihood: 4, severity: 98, category: 'Financial', branch: 'Head Office', assessments: [assessment(5, 4)] },
  { id: 'r-cyber', name: 'Cyber security breach', impact: 5, likelihood: 5, severity: 88, category: 'Technology', branch: 'Head Office', assessments: [assessment(5, 5)] },
  { id: 'r-fraud', name: 'Fraud', impact: 4, likelihood: 5, severity: 92, category: 'Financial', branch: 'Surabaya Branch', assessments: [assessment(4, 5)] }
]

async function loadStore(risks = backendRisks) {
  global.$fetch = vi.fn(async (url: string) =>
    String(url).includes('/locations') ? okResponse([]) : okResponse(risks)
  )
  const store = useRiskProfileStore()
  await store.fetchRisks()
  return store
}

// --- Dashboard formulas, replicated from pages/dashboard/index.vue (SFC not importable) ---
//  registeredRiskHeatMap = riskProfileStore.risks.slice(0, 5)          (~line 1419)
//  Score column  = risk.impact * risk.likelihood                       (~line 1438)
//  Level column  = riskProfileStore.getRiskLevel(likelihood, impact),
//                  badge colour riskLevelConfig[level].color           (~line 1441)
const dashboardRow = (store, r) => {
  const level = store.getRiskLevel(r.likelihood, r.impact)
  return {
    name: r.name,
    score: r.impact * r.likelihood,
    level,
    color: riskLevelConfig[level]?.color
  }
}
const dashboardRegisteredRisks = store =>
  store.risks.slice(0, 5).map(r => dashboardRow(store, r))

// --- CRP formulas, replicated from components/risk-profile/RiskHeatMap.vue ---
//  filteredRisks: store.selectedBranch filter                         (~line 746)
//  priorityRisks (default tab 'priority'): level.priority, sorted by
//    getRiskScore desc then severity desc                             (~line 752)
//  list score    = getRiskScore(risk.likelihood, risk.impact)         (~line 362)
const crpFiltered = store =>
  store.selectedBranch === 'All Branches'
    ? store.risks
    : store.risks.filter(r => r.branch === store.selectedBranch)

const crpPriorityList = store =>
  crpFiltered(store)
    .filter(r => riskLevelConfig[store.getRiskLevel(r.likelihood, r.impact)].priority)
    .sort((a, b) => {
      const diff = store.getRiskScore(b.likelihood, b.impact) - store.getRiskScore(a.likelihood, a.impact)
      return diff !== 0 ? diff : b.severity - a.severity
    })
    .map(r => ({
      name: r.name,
      score: store.getRiskScore(r.likelihood, r.impact)
    }))

// CRP list badge: getRiskLevel(risk.likelihood, risk.impact), coloured with
// riskLevelConfig[level].color                                         (~line 335)
const crpLevel = (store, r) => store.getRiskLevel(r.likelihood, r.impact)

// The registeredRiskColumns block of the real dashboard SFC.
const dashboardColumnsSource = () => {
  const src = readFileSync(DASHBOARD, 'utf8')
  const start = src.indexOf('const registeredRiskColumns = [')
  const end = src.indexOf('\n];', start)
  return src.slice(start, end)
}

describe('Dashboard Registered Risks vs Corporate Risk Profile', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  // --- Shared source: these pass, the rows themselves are the same ---

  it('both sides read the same store rows for the selected year/quarter', async () => {
    const store = await loadStore()
    expect(store.risks).toHaveLength(backendRisks.length)
    store.selectedPeriod = 'Q3'
    const fin = store.risks.find(r => r.id === 'r-fin')
    expect([fin.impact, fin.likelihood]).toEqual([5, 4])
  })

  // --- Intended table columns: these pass ---

  it('has no ID column, a Score and a Risk Level column', () => {
    const cols = dashboardColumnsSource()
    expect(cols).not.toContain('getFormattedId')
    expect(cols).not.toMatch(/header:\s*"ID"/)
    expect(cols).toContain('rawObject.impact * rawObject.likelihood')
    expect(cols).toContain('riskProfileStore.getRiskLevel(risk.likelihood, risk.impact)')
    expect(cols).toContain('getRiskLevelColorClass(level)')
    expect(cols).toContain('t("dashboard.registeredRisks.columns.level")')
    expect(cols).toContain('t(`riskProfile.riskLevelLabels.${level}`)')
  })

  it('has the Risk Level header in both locales', () => {
    for (const file of [EN, ID]) {
      const json = JSON.parse(readFileSync(file, 'utf8'))
      expect(json.dashboard?.registeredRisks?.columns?.level).toBeTruthy()
    }
  })

  it('scores a risk as impact × likelihood', async () => {
    const store = await loadStore()
    const dash = Object.fromEntries(dashboardRegisteredRisks(store).map(r => [r.name, r.score]))
    // Q1 2026 values from the fixture assessments.
    expect(dash).toEqual({
      'Office supplies shortage': 2, // I1 × L2
      'Vendor delay': 9, // I3 × L3
      'Minor policy gap': 4, // I2 × L2
      'Printer downtime': 1, // I1 × L1
      'Revenue target missed': 20 // I5 × L4
    })
  })

  it('shows the same risk level (and colour) as the CRP for each risk', async () => {
    const store = await loadStore()
    for (const r of store.risks.slice(0, 5)) {
      const row = dashboardRow(store, r)
      expect({ name: row.name, level: row.level }).toEqual({ name: r.name, level: crpLevel(store, r) })
      expect(row.color).toBe(riskLevelConfig[crpLevel(store, r)].color)
    }
    const byName = Object.fromEntries(dashboardRegisteredRisks(store).map(r => [r.name, r.level]))
    expect(byName['Revenue target missed']).toBe(RiskLevel.HIGH)
    expect(byName['Vendor delay']).toBe(RiskLevel.MODERATE)
    expect(byName['Printer downtime']).toBe(RiskLevel.LOW)
  })

  it('passes likelihood and impact to getRiskLevel in the right order', async () => {
    // Asymmetric cell: L5/I1 is Low-Moderate, L1/I5 would be High.
    const store = await loadStore([
      { id: 'r-asym', name: 'Frequent but trivial', impact: 1, likelihood: 5, severity: 50, category: 'Operations', branch: 'Head Office', assessments: [assessment(1, 5)] }
    ])
    const [row] = dashboardRegisteredRisks(store)
    expect(row.level).toBe(RiskLevel.LOW_MODERATE)
    expect(row.level).toBe(crpLevel(store, store.risks[0]))
  })

  // --- Row selection / order: still open (out of scope), these fail on purpose ---

  it('lists the same top risks, in the same order, as the CRP priority list', async () => {
    const store = await loadStore()
    const dash = dashboardRegisteredRisks(store).map(r => r.name)
    const crp = crpPriorityList(store).slice(0, 5).map(r => r.name)
    // CRP: Cyber (25), Revenue (24), Fraud (22), Vendor (13).
    // Dashboard: first five rows in API order, incl. Low risks the CRP never prioritises.
    expect(dash).toEqual(crp)
  })

  it('does not depend on the order the API returns rows in', async () => {
    const a = dashboardRegisteredRisks(await loadStore(backendRisks)).map(r => r.name)
    setActivePinia(createPinia())
    const b = dashboardRegisteredRisks(await loadStore([...backendRisks].reverse())).map(r => r.name)
    expect(new Set(a)).toEqual(new Set(b))
  })

  it('does not show Low risks while higher-scoring CRP risks are left out', async () => {
    const store = await loadStore()
    const shownLevels = store.risks
      .slice(0, 5)
      .map(r => store.getRiskLevel(r.likelihood, r.impact))
    const hidden = store.risks.slice(5)
    const hiddenPriority = hidden.filter(r => riskLevelConfig[store.getRiskLevel(r.likelihood, r.impact)].priority)
    // With the fixture above: Low/Low-Moderate rows shown, Cyber (High) and Fraud (High) hidden.
    if (hiddenPriority.length > 0) {
      expect(shownLevels).not.toContain(RiskLevel.LOW)
    }
  })
})
