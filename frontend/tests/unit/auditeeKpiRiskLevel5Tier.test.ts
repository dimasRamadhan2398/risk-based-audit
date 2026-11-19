// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'
import { toEntityRiskCode, normalizeEntityRiskLevel } from '~/utils/riskLevelBadge'
import {
  useAiAnalytics,
  getPersistedCustomAuditeeKpis,
  sanitizeAuditeeKpiForecastRecord,
  sanitizeAuditeeAtRiskRecord
} from '~/composables/useAiAnalytics'
import { toKpiForecastView, selectAtRiskKpis } from '~/composables/useKpiForecast'

const CANONICAL_5_CODES = ['HIGH', 'MODERATE_HIGH', 'MODERATE', 'LOW_MODERATE', 'LOW']
const CANONICAL_5_LABELS = ['High', 'Moderate to High', 'Moderate', 'Low to Moderate', 'Low']

const KPI_FORECAST_PAGE = resolve(__dirname, '../../pages/analytics/ai/kpi-forecast.vue')
const MAPPED_BANNER = resolve(__dirname, '../../components/common/MappedSourceDataBanner.vue')

describe('Auditee KPI Forecast 5-Tier Risk Level Enforcement', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    global.$fetch = vi.fn(async () => null)
    const { timeseriesState } = useAiAnalytics()
    timeseriesState.value.kpiForecasts = []
    timeseriesState.value.atRiskDepartments = []
    timeseriesState.value.auditeeForecast = {
      kpiForecasts: [],
      atRiskDepartments: [],
      timeSeriesData: [],
      executiveSummary: ''
    }
  })

  it('sanitizes arbitrary and external non-canonical risk levels onto strictly the 5 canonical levels', () => {
    const samples: Array<[string, string, string]> = [
      ['CRITICAL', 'HIGH', 'High'],
      ['Critical', 'HIGH', 'High'],
      ['Extreme', 'HIGH', 'High'],
      ['Very High', 'HIGH', 'High'],
      ['MEDIUM_TO_HIGH', 'MODERATE_HIGH', 'Moderate to High'],
      ['Major', 'MODERATE_HIGH', 'Moderate to High'],
      ['Elevated', 'MODERATE_HIGH', 'Moderate to High'],
      ['KOL-3', 'MODERATE_HIGH', 'Moderate to High'],
      ['MEDIUM', 'MODERATE', 'Moderate'],
      ['Watch', 'MODERATE', 'Moderate'],
      ['Not Scored', 'MODERATE', 'Moderate'],
      ['LOW_TO_MEDIUM', 'LOW_MODERATE', 'Low to Moderate'],
      ['Minor', 'LOW_MODERATE', 'Low to Moderate'],
      ['KOL-2', 'LOW_MODERATE', 'Low to Moderate'],
      ['Very Low', 'LOW', 'Low'],
      ['Insignificant', 'LOW', 'Low'],
      ['Safe', 'LOW', 'Low']
    ]

    const { getRiskConfig } = useAiAnalytics()
    for (const [raw, expectedCode, expectedLabel] of samples) {
      const sanitized = sanitizeAuditeeKpiForecastRecord({
        kpiName: `KPI (${raw})`,
        riskLevel: raw,
        impactScore: 3,
        likelihoodScore: 3
      })
      expect(CANONICAL_5_CODES).toContain(sanitized.riskLevel)
      expect(sanitized.riskLevel).toBe(expectedCode)
      expect(getRiskConfig(sanitized.riskLevel).label).toBe(expectedLabel)
      expect(CANONICAL_5_LABELS).toContain(normalizeEntityRiskLevel(raw))
    }
  })

  it('heals legacy CRITICAL and MEDIUM_TO_HIGH records stored in custom Auditee KPI localStorage and addCustomAuditeeKpiForecasts', () => {
    localStorage.setItem(
      'auditsphere_custom_auditee_kpis_v1',
      JSON.stringify([
        { kpiName: 'Legacy Critical NPL', code: 'KPI-EXT-NPL', riskLevel: 'CRITICAL', deltaPct: 12, forecastedValue: 5.5, unit: '%' },
        { kpiName: 'Legacy MediumToHigh Yield', code: 'KPI-EXT-YLD', riskLevel: 'MEDIUM_TO_HIGH', deltaPct: -4, forecastedValue: 8.9, unit: '%' }
      ])
    )

    const loaded = getPersistedCustomAuditeeKpis()
    expect(loaded).toHaveLength(2)
    expect(loaded[0].riskLevel).toBe('HIGH')
    expect(loaded[1].riskLevel).toBe('MODERATE_HIGH')

    const { addCustomAuditeeKpiForecasts, timeseriesState } = useAiAnalytics()
    addCustomAuditeeKpiForecasts([
      {
        kpiName: 'External DB Extreme Anomaly Ratio',
        code: 'KPI-EXT-TRX',
        entity: 'Unit Operasional',
        riskLevel: 'Extreme',
        trend: 'Deteriorating',
        deltaPct: 15,
        forecastedValue: 9.2,
        unit: '%',
        currentValue: 8.0
      }
    ])

    const allForecasts = timeseriesState.value.auditeeForecast?.kpiForecasts || []
    const allAtRisk = timeseriesState.value.auditeeForecast?.atRiskDepartments || []
    for (const row of allForecasts) {
      expect(CANONICAL_5_CODES).toContain(row.riskLevel)
    }
    for (const dep of allAtRisk) {
      expect(CANONICAL_5_CODES).toContain(dep.riskLevel)
    }
  })

  it('normalizes fetched batch Auditee forecasts & at-risk departments onto the 5 canonical levels', async () => {
    global.$fetch = vi.fn(async (url: string) => {
      if (String(url).includes('performance-trend/batch')) {
        return {
          auditee_forecast: {
            kpi_forecasts: [
              { kpiName: 'External NPL', code: 'KPI-001', currentValue: 5.1, forecastedValue: 5.9, trend: 'Deteriorating', riskLevel: 'CRITICAL' },
              { kpiName: 'External Yield', code: 'KPI-002', currentValue: 9.2, forecastedValue: 8.8, trend: 'Deteriorating', riskLevel: 'MEDIUM_TO_HIGH' },
              { kpiName: 'External Cost', code: 'KPI-003', currentValue: 55.0, forecastedValue: 55.2, trend: 'Improving', riskLevel: 'MEDIUM' },
              { kpiName: 'External Ops', code: 'KPI-004', currentValue: 2.2, forecastedValue: 2.1, trend: 'Improving', riskLevel: 'Minor' },
              { kpiName: 'External LCR', code: 'KPI-005', currentValue: 150.0, forecastedValue: 152.0, trend: 'Improving', riskLevel: 'Safe' }
            ],
            at_risk_departments: [
              { department: 'Portofolio Kredit', kpi: 'External NPL', riskLevel: 'CRITICAL' },
              { department: 'Komersial', kpi: 'External Yield', riskLevel: 'Major' }
            ]
          }
        }
      }
      return null
    })

    const { fetchAiAnalytics, timeseriesState, getRiskConfig } = useAiAnalytics()
    await fetchAiAnalytics(true, { skipAutoRetrain: true })

    const list = timeseriesState.value.auditeeForecast?.kpiForecasts || []
    const atRisk = timeseriesState.value.auditeeForecast?.atRiskDepartments || []
    expect(list.map((k: any) => k.riskLevel)).toEqual([
      'HIGH',
      'MODERATE_HIGH',
      'MODERATE',
      'LOW_MODERATE',
      'LOW'
    ])
    for (const k of list) {
      expect(CANONICAL_5_LABELS).toContain(getRiskConfig(k.riskLevel).label)
    }
    expect(atRisk.map((d: any) => d.riskLevel)).toEqual(['HIGH', 'MODERATE_HIGH'])

    const viewModels = list.map(toKpiForecastView)
    expect(selectAtRiskKpis(viewModels).every((v) => CANONICAL_5_CODES.includes(v.riskLevel))).toBe(true)
    expect(sanitizeAuditeeAtRiskRecord({ kpi: 'x', riskLevel: 'CRITICAL' }).riskLevel).toBe('HIGH')
  })

  it('enforces exact 5-tier filter options and removes CRITICAL/MEDIUM_TO_HIGH from kpi-forecast.vue and MappedSourceDataBanner.vue', () => {
    const pageSrc = readFileSync(KPI_FORECAST_PAGE, 'utf-8')
    const bannerSrc = readFileSync(MAPPED_BANNER, 'utf-8')

    expect(pageSrc).not.toContain('value="CRITICAL"')
    expect(pageSrc).not.toContain('>Critical<')
    expect(pageSrc).not.toContain('Moderate / Moderate to High')
    expect(pageSrc).toContain('value="MODERATE_HIGH"')
    expect(pageSrc).toContain('value="LOW_MODERATE"')
    expect(pageSrc).toContain('toEntityRiskCode(k.riskLevel')

    expect(bannerSrc).not.toContain("riskLevel: foreNpl >= 5 ? 'CRITICAL'")
    expect(bannerSrc).not.toContain("riskLevel: 'MEDIUM_TO_HIGH'")
    expect(bannerSrc).toContain('resolveExternalAuditeeRisk')
    expect(toEntityRiskCode('Critical')).toBe('HIGH')
  })
})
