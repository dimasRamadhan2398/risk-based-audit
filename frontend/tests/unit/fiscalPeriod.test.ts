import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { currentFiscalYear, currentQuarter } from '~/utils/fiscalPeriod'
import { getFiscalYears } from '~/composables/useFiscalYear'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({ public: { masterServiceBaseUrl: 'http://localhost:8080/api/v1' } })
}))

const at = (y: number, m: number, d: number, h = 12) => new Date(y, m - 1, d, h, 0, 0)

describe('currentFiscalYear / currentQuarter', () => {
  it.each([
    [at(2026, 11, 4), 2026, 'Q4'],
    [at(2026, 12, 31, 23), 2026, 'Q4'],
    [at(2027, 1, 1, 0), 2027, 'Q1'],
    [at(2026, 3, 31, 23), 2026, 'Q1'],
    [at(2026, 4, 1, 0), 2026, 'Q2'],
    [at(2026, 6, 30), 2026, 'Q2'],
    [at(2026, 7, 1, 0), 2026, 'Q3'],
    [at(2026, 10, 1, 0), 2026, 'Q4']
  ])('%s -> %i/%s', (d, y, q) => {
    expect(currentFiscalYear(d)).toBe(y)
    expect(currentQuarter(d)).toBe(q)
  })

  it('year options contain the default year and keep earlier years', () => {
    const years = getFiscalYears(currentFiscalYear(at(2027, 1, 1)))
    expect(years).toContain(2027)
    expect(years).toContain(2024)
  })
})

describe('risk-profile store default period', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('initial state follows the current date', async () => {
    const { useRiskProfileStore } = await import('~/stores/risk-profile')
    const store = useRiskProfileStore()
    const now = new Date()
    expect(store.selectedPeriod).toBe(currentQuarter(now))
    expect(store.selectedYear).toBe(currentFiscalYear(now))
  })

  it('syncDefaultPeriod advances an untouched selection but keeps a user choice', async () => {
    const { useRiskProfileStore } = await import('~/stores/risk-profile')
    const store = useRiskProfileStore()
    store.syncDefaultPeriod(at(2027, 1, 2))
    expect(store.selectedYear).toBe(2027)
    expect(store.selectedPeriod).toBe('Q1')
    store.selectedPeriod = 'Q3'
    store.syncDefaultPeriod(at(2027, 4, 2))
    expect(store.selectedPeriod).toBe('Q3')
    expect(store.selectedYear).toBe(2027)
  })
})
