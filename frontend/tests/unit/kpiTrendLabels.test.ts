/**
 * KPI trend chart month labels (utils/kpiPerformanceLabels.ts): the API sends the last 6 months up to now,
 * so the window can cross a year boundary; then each label carries its year.
 */
import { describe, it, expect } from 'vitest'
import { kpiTrendLabels } from '~/utils/kpiPerformanceLabels'

const id: Record<string, string> = { 'kpiPerformance.charts.months.dec': 'Des', 'kpiPerformance.charts.months.may': 'Mei' }
const t = (key: string) => id[key] ?? key

describe('kpiTrendLabels', () => {
  it('a window within one year shows translated months only', () => {
    expect(kpiTrendLabels(t, ['May', 'Jun'], ['2026-05', '2026-06'])).toEqual(['Mei', 'Jun'])
  })

  it('a window crossing a year adds the year to every label', () => {
    expect(kpiTrendLabels(t, ['Dec', 'Jan'], ['2025-12', '2026-01'])).toEqual(['Des 2025', 'Jan 2026'])
  })

  it('without matching periods (older backend) the months are shown alone', () => {
    expect(kpiTrendLabels(t, ['Dec', 'Jan'])).toEqual(['Des', 'Jan'])
    expect(kpiTrendLabels(t, ['Dec', 'Jan'], ['2025-12'])).toEqual(['Des', 'Jan'])
    expect(kpiTrendLabels(t, ['Dec', 'Jan'], ['x', '2026-01'])).toEqual(['Des', 'Jan'])
  })
})
