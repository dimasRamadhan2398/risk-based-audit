// @ts-nocheck
/**
 * Dashboard "Audit Execution Status" card vs the Audit Execution Status page.
 *
 * Both read `useAuditExecutionStore().auditExecutions`. The page calls
 * `fetchAuditExecutions()` on setup; these tests check whether the dashboard card
 * shows the same executions, with the same progress, from the same (fetched)
 * source.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'

import { useAuditExecutionStore } from '~/stores/audit-execution'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: { auditServiceBaseUrl: 'http://localhost:8002/api/v1' }
  })
}))

const DASHBOARD_PAGE = resolve(__dirname, '../../pages/dashboard/index.vue')
const STATUS_PAGE = resolve(__dirname, '../../pages/audit-execution-status/index.vue')

/** Realistic GET /audit-executions payload — different from the store's seed data. */
const backendPayload = {
  success: true,
  data: {
    items: [
      { id: 'be-1', ref: 'AUD-2026-101', name: 'Treasury & Liquidity', category: 'Assurance', progress: 100, status: 'completed', created_at: '2026-01-10T00:00:00Z' },
      { id: 'be-2', ref: 'AUD-2026-102', name: 'Credit Underwriting', category: 'Assurance', progress: 65, status: 'draft findings', created_at: '2026-04-02T00:00:00Z' },
      { id: 'be-3', ref: 'AUD-2026-103', name: 'Branch Cash Handling', category: 'Special Audit', progress: 30, status: 'fieldwork', created_at: '2026-07-11T00:00:00Z' },
      { id: 'be-4', ref: 'AUD-2026-104', name: 'IT General Controls', category: 'Assurance', progress: 10, status: 'entry meeting', created_at: '2026-09-01T00:00:00Z' }
    ]
  }
}

/** Replica of `dashboardExecutionStatus` in pages/dashboard/index.vue */
const buildDashboardCard = (executions: any[]) =>
  executions
    .map((e: any) => ({ name: e.name, percentage: e.progress }))
    .slice(0, 3)

/** Replica of `executionStatusPercent` in pages/dashboard/index.vue */
const buildExecutionStatusPercent = (executions: any[]) => {
  if (executions.length === 0) return 0
  return executions.reduce((sum: number, e: any) => sum + e.progress, 0) / (executions.length * 100)
}

/** What the Audit Execution Status page lists (no filters applied). */
const buildPageRows = (executions: any[]) =>
  executions.map((e: any) => ({ ref: e.ref, name: e.name, percentage: e.progress }))

describe('Audit Execution Status — dashboard card vs the feature page', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    global.$fetch = vi.fn(async () => backendPayload)
  })

  it('the dashboard card shows the data the page shows', async () => {
    const store = useAuditExecutionStore()

    // The dashboard reads the store without fetching, so capture that state...
    const cardBeforeFetch = buildDashboardCard(store.auditExecutions)

    // ...while the page fetches on setup.
    await store.fetchAuditExecutions()
    const pageRows = buildPageRows(store.auditExecutions)

    for (const item of cardBeforeFetch) {
      expect(
        pageRows.some((r) => r.name === item.name && r.percentage === item.percentage),
        `card row "${item.name} ${item.percentage}%" is not on the page`
      ).toBe(true)
    }
  })

  it('the dashboard fetches executions instead of rendering seed state', () => {
    const src = readFileSync(DASHBOARD_PAGE, 'utf-8')
    expect(
      src.includes('fetchAuditExecutions'),
      'dashboard/index.vue never calls fetchAuditExecutions'
    ).toBe(true)
  })

  it('the store seed state is not the backend data (so reading it unfetched is stale)', async () => {
    const store = useAuditExecutionStore()
    const seedRefs = store.auditExecutions.map((e: any) => e.ref)

    await store.fetchAuditExecutions()
    const fetchedRefs = store.auditExecutions.map((e: any) => e.ref)

    expect(seedRefs).not.toEqual(fetchedRefs)
  })

  it('the card lists every execution the page lists', async () => {
    const store = useAuditExecutionStore()
    await store.fetchAuditExecutions()

    expect(buildDashboardCard(store.auditExecutions)).toHaveLength(
      buildPageRows(store.auditExecutions).length
    )
  })

  it('card rows carry an identifier so same-named audits stay distinguishable', async () => {
    const store = useAuditExecutionStore()
    // Two different audits can share a name (the seed data has two
    // "Financial Operations" at 100% and 40%).
    const card = buildDashboardCard(store.auditExecutions)
    const names = card.map((r: any) => r.name)

    expect(
      new Set(names).size === names.length || card.every((r: any) => r.ref),
      'card rows are labelled by name only, so duplicates look contradictory'
    ).toBe(true)
  })

  it('the Execution Status tile survives a string progress from the API', async () => {
    global.$fetch = vi.fn(async () => ({
      data: {
        items: [
          { id: 's1', ref: 'AUD-2026-201', name: 'A', progress: '100', status: 'completed' },
          { id: 's2', ref: 'AUD-2026-202', name: 'B', progress: '40', status: 'fieldwork' }
        ]
      }
    }))

    const store = useAuditExecutionStore()
    await store.fetchAuditExecutions()

    const percent = buildExecutionStatusPercent(store.auditExecutions)
    expect(Number.isFinite(percent), `Execution Status computed as ${percent}`).toBe(true)
    expect(percent).toBeCloseTo(0.7, 5)
  })

  it('the card period badge is not hardcoded to a different year than the data', () => {
    const src = readFileSync(DASHBOARD_PAGE, 'utf-8')
    // Anchor on the card's own comment markers — the row comment above them
    // mentions both card titles, so searching for the titles alone matches it.
    const card = src.slice(
      src.indexOf('<!-- Audit Execution Status -->'),
      src.indexOf('<!-- Recent Finding Issues -->')
    )
    expect(card.length, 'could not locate the Audit Execution Status card').toBeGreaterThan(100)
    const hardcodedQuarter = card.match(/label="(Q[1-4]\s*\d{4})"/)

    expect(
      hardcodedQuarter,
      `card shows a hardcoded period badge ${hardcodedQuarter?.[1]} regardless of the data`
    ).toBeNull()
  })

  it('both views derive progress the same way', () => {
    const page = readFileSync(STATUS_PAGE, 'utf-8')
    const dashboard = readFileSync(DASHBOARD_PAGE, 'utf-8')

    // The page normalises progress through getProgressValue (handles strings);
    // the dashboard must not read e.progress raw if the page needs coercion.
    const pageCoerces = page.includes('parseFloat')
    const dashboardCoerces =
      dashboard.includes('getProgressValue') || dashboard.includes('Number(e.progress)')

    expect(
      !pageCoerces || dashboardCoerces,
      'the page coerces progress but the dashboard uses it raw'
    ).toBe(true)
  })
})
