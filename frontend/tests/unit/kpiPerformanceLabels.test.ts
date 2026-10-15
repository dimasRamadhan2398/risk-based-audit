import { describe, it, expect } from 'vitest'
import en from '~/locales/en/common.json'
import id from '~/locales/id/common.json'
import { kpiMonthLabel, kpiValueLabel, kpiSubText, kpiSubMetric, type TranslateFn } from '~/utils/kpiPerformanceLabels'

// Same lookup semantics as composables/useI18n.ts: missing key returns the key itself.
const makeT = (dict: object): TranslateFn => (key, params) => {
  let value: unknown = dict
  for (const k of key.split('.')) {
    if (value && typeof value === 'object' && k in value) value = (value as Record<string, unknown>)[k]
    else return key
  }
  if (typeof value !== 'string') return key
  return params ? value.replace(/\{(\w+)\}/g, (m, p) => params[p]?.toString() || m) : value
}
const tEn = makeT(en)
const tId = makeT(id)

describe('KPI performance labels', () => {
  it('translates API month abbreviations and falls back to the raw label', () => {
    expect(kpiMonthLabel(tId, 'May')).toBe(id.kpiPerformance.charts.months.may)
    expect(kpiMonthLabel(tEn, 'Jan')).toBe('Jan')
    expect(kpiMonthLabel(tId, 'January')).toBe('January')
    expect(kpiMonthLabel(tId, 'W01')).toBe('W01')
    expect(kpiMonthLabel(makeT({}), 'Feb')).toBe('Feb')
  })

  it('labels statuses/categories without changing the raw value, falling back for unknown values', () => {
    expect(kpiValueLabel(tId, 'statuses', 'On Track')).toBe('Sesuai Target')
    expect(kpiValueLabel(tId, 'categories', 'Financial')).toBe('Keuangan')
    expect(kpiValueLabel(tId, 'statuses', 'Unknown Status')).toBe('Unknown Status')
    expect(kpiValueLabel(tEn, 'categories', 'Brand New')).toBe('Brand New')
  })

  it('translates known backend sub-metric formats and leaves others raw', () => {
    expect(kpiSubText(tId, '11 completed / 12 started')).toBe('11 selesai / 12 dimulai')
    expect(kpiSubText(tId, '-1.5 days vs target')).toBe('-1.5 hari terhadap target')
    expect(kpiSubText(tId, '< 14 days')).toBe('< 14 hari')
    expect(kpiSubText(tId, '0 overdue')).toBe('0 terlambat')
    expect(kpiSubText(tEn, '12 responses from 14 completed audits')).toBe('12 responses from 14 completed audits')
    expect(kpiSubText(tId, '92.0%')).toBe('92.0%')
    expect(kpiSubText(tId, 'NaN days')).toBe('NaN days')
    expect(kpiSubText(tId, '3 open items')).toBe('3 open items')
  })

  it('translates known sub-metric titles and keeps unknown titles', () => {
    const out = kpiSubMetric(tId, { title: 'Survey Response Rate', value: '85.5%', target: '80%', trend: '4 open' })
    expect(out).toEqual({ title: 'Tingkat Respons Survei', value: '85.5%', target: '80%', trend: '4 terbuka' })
    expect(kpiSubMetric(tId, { title: 'New Metric', value: 'x', target: 'y', trend: 'z' }).title).toBe('New Metric')
  })
})
