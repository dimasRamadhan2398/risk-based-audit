// @ts-nocheck
/**
 * Number semantics on the dashboard AI cards: rupiah scale words, KPI units
 * and direction-aware trends.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import {
  evaluateKpiTrend,
  formatKpiForecastValue,
  formatForecastPeriod,
  formatForecastPeriodRange,
  formatRupiahCompact,
  formatRupiahFromMillions,
  getKpiMeta,
  normalizeRupiahInText,
  parseLooseNumber
} from '~/utils/analyticsFormat'
import { createAnomalyTypeConfigs, ANOMALY_POINT_COLORS } from '~/composables/useAnomalyScatter'

const t = (k: string) => k

describe('rupiah amounts use Indonesian scale words, never "M"', () => {
  it('20,114.34 juta → Rp 20,11 miliar', () => {
    expect(formatRupiahCompact(20_114_340_000, 'id')).toBe('Rp 20,11 miliar')
    expect(formatRupiahFromMillions(20114.34, 'id')).toBe('Rp 20,11 miliar')
    expect(formatRupiahCompact(20_114_340_000, 'en')).toBe('Rp 20.11 billion')
  })

  it('picks juta / miliar / triliun by magnitude', () => {
    expect(formatRupiahFromMillions(450, 'id')).toBe('Rp 450 juta')
    expect(formatRupiahFromMillions(2_500_000, 'id')).toBe('Rp 2,5 triliun')
    expect(formatRupiahCompact(950_000, 'id')).toBe('Rp 950.000')
  })

  it('rewrites amounts inside model/dataset text', () => {
    expect(normalizeRupiahInText('Penarikan dana masif (Rp 20114.34 Juta) pada rekening dormant', 'id'))
      .toBe('Penarikan dana masif (Rp 20,11 miliar) pada rekening dormant')
    expect(normalizeRupiahInText('Transfer - Rp 20,114.34M (Jam 3:00)', 'id'))
      .toBe('Transfer - Rp 20,11 miliar (Jam 3:00)')
    expect(normalizeRupiahInText('kredit Rp 66.662 Juta', 'id')).toBe('kredit Rp 66,66 miliar')
  })

  it('parses both number conventions', () => {
    expect(parseLooseNumber('20114.34')).toBe(20114.34)
    expect(parseLooseNumber('20,114.34')).toBe(20114.34)
    expect(parseLooseNumber('20.114,34')).toBe(20114.34)
    expect(parseLooseNumber('66.662')).toBe(66662)
  })

  it('anomaly scatter tooltips/ticks for rupiah categories use the same format', () => {
    const cfg = createAnomalyTypeConfigs(t, 'id')
    expect(cfg.Funding.formatX(20114.34)).toBe('Rp 20,11 miliar')
  })

  it('no "Rp …M" template is left in the AI cards', () => {
    for (const rel of ['pages/dashboard/index.vue', 'pages/analytics/ai/anomaly-detection.vue', 'composables/useAnomalyScatter.ts']) {
      const src = readFileSync(resolve(__dirname, '../../', rel), 'utf-8')
      expect(src, rel).not.toMatch(/Rp \$\{[^}]+\}M|\}\}M\b|toLocaleString\(\) \}\}M/)
    }
  })
})

describe('KPI units', () => {
  it('NPS is a score, not a percentage', () => {
    const meta = getKpiMeta('Net Promoter Score')
    expect(formatKpiForecastValue(53.7, meta, 'id')).toBe('+53,7')
    expect(formatKpiForecastValue(53.7, meta, 'id')).not.toContain('%')
  })

  it('CLV is a rupiah amount, not a percentage', () => {
    const meta = getKpiMeta('Customer Lifetime Value')
    expect(formatKpiForecastValue(690.67, meta, 'id')).toBe('Rp 690,67 juta')
  })

  it('rates stay percentages', () => {
    expect(formatKpiForecastValue(91.7, getKpiMeta('Customer Retention Rate'), 'id')).toBe('91,7%')
  })
})

describe('direction-aware KPI trend', () => {
  it('a rising turnover rate deteriorates', () => {
    expect(evaluateKpiTrend(12, 13.5, getKpiMeta('Employee Turnover Rate'))).toBe('Deteriorating')
  })
  it('a falling NPS deteriorates, a rising one improves', () => {
    const meta = getKpiMeta('Net Promoter Score')
    expect(evaluateKpiTrend(47.7, 46.1, meta)).toBe('Deteriorating')
    expect(evaluateKpiTrend(47.7, 49, meta)).toBe('Improving')
  })
  it('tiny moves are stable', () => {
    expect(evaluateKpiTrend(90, 90.1, getKpiMeta('Customer Retention Rate'))).toBe('Stable')
  })
})

describe('periods', () => {
  it('formats months and ranges; drops legacy "(Pred)" suffixes', () => {
    expect(formatForecastPeriod('2025-12', 'id')).toBe('Des 2025')
    expect(formatForecastPeriodRange('2026-01', '2026-03', 'id')).toBe('Jan–Mar 2026')
    expect(formatForecastPeriodRange('2026-11', '2027-01', 'en')).toBe('Nov 2026–Jan 2027')
    expect(formatForecastPeriod('Q1 2026 (Pred)', 'id')).toBe('Q1 2026')
  })
})

describe('anomaly colours', () => {
  it('every category draws anomalies in red, never green', () => {
    const cfg = createAnomalyTypeConfigs(t, 'id')
    for (const [type, c] of Object.entries(cfg)) {
      expect(c.colors.bg, type).toBe(ANOMALY_POINT_COLORS.bg)
      expect(c.colors.bg, type).not.toMatch(/16,185,129/)
    }
  })
})
