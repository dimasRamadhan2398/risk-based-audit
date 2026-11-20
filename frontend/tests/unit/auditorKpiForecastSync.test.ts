// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { useAiAnalytics } from '~/composables/useAiAnalytics'
import { buildKpiForecastChart, selectAtRiskKpis, toKpiForecastView } from '~/composables/useKpiForecast'
import {
  AUDITOR_CHART_ALL_KEY,
  buildAuditorAggregateTimeSeries,
  buildAuditorChartView,
  buildAuditorLocalForecastsFromItems,
  computeAuditeeCustomFingerprint,
  computeAuditorFingerprint,
  dedupAndBuildAuditorPayloadItems,
  hasTargetAndActualColumns
} from '~/utils/auditorKpiSync'

const KPI_FORECAST_PAGE = resolve(__dirname, '../../pages/analytics/ai/kpi-forecast.vue')
const USE_AI_ANALYTICS = resolve(__dirname, '../../composables/useAiAnalytics.ts')

const t = (key: string, params?: any) => (params ? `${key}:${JSON.stringify(params)}` : key)

describe('Auditor KPI strict Target & Actual ingestion (no CSV/mock fallback)', () => {
  it('accepts only items that genuinely carry both numeric target and actual columns', () => {
    expect(hasTargetAndActualColumns({ kpi: 'PKAT', target: '90', actual: '82' })).toBe(true)
    expect(hasTargetAndActualColumns({
      kpi: 'Response Time',
      kpiTargets: { '2025': '5' },
      kpiActuals: { '2025': '4' }
    })).toBe(true)

    // Missing actual or non-numeric target/actual -> excluded
    expect(hasTargetAndActualColumns({ kpi: 'Vision Goal Only', target: '100', actual: '' })).toBe(false)
    expect(hasTargetAndActualColumns({ kpi: 'No Target', actual: '88' })).toBe(false)
    expect(hasTargetAndActualColumns({ kpi: 'Incomplete Map', kpiTargets: { '2026': '90' }, kpiActuals: {} })).toBe(false)
  })

  it('merges Strategic Audit Plan with non-duplicate KPIAchievement rows and skips duplicates', () => {
    const strategicPlans = [
      {
        code: 'SO-IA01',
        strategicObjective: 'Penyelesaian PKAT',
        kpi: 'Realisasi Pelaksanaan PKAT',
        category: 'Pelaksanaan Audit & PKAT',
        unit: '%',
        hibHig: 'HIG',
        selectedPeriod: '2025',
        target: '90',
        actual: '84',
        kpiTargets: { '2024': '85', '2025': '90', '2026': '95' },
        kpiActuals: { '2024': '82', '2025': '84' }
      },
      {
        code: 'SO-IA02',
        strategicObjective: 'Draft Only Without Actual',
        kpi: 'Draft Indicator',
        target: '100',
        actual: ''
      }
    ]

    const kpiAchievements = [
      { kpi_name: 'Realisasi Pelaksanaan PKAT', year: 2025, period: 'Q4', target: 90, actual: 84 },
      { kpi_name: 'Ketepatan Waktu LHA', year: 2024, period: 'FY', target: 90, actual: 85 },
      { kpi_name: 'Ketepatan Waktu LHA', year: 2025, period: 'FY', target: 95, actual: 92 },
      { kpi_name: 'Missing Target Perf', year: 2025, period: 'FY', actual: 70 }
    ]

    const items = dedupAndBuildAuditorPayloadItems(strategicPlans, kpiAchievements, 2026)

    expect(items).toHaveLength(2)
    expect(items[0].code).toBe('SO-IA01')
    expect(items[0].source).toBe('Strategic Audit Plan')
    expect(items[1].kpi).toBe('Ketepatan Waktu LHA')
    expect(items[1].source).toBe('Internal Audit Performance')
    expect(items[1].kpiActuals).toEqual({ 'FY-2024': 85, 'FY-2025': 92 })
  })
})

describe('syncAuditorKpiWithStrategicPlan pagination, AI scoring & empty state', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('returns empty auditorForecast instead of mock CSV items when Strategic Plan & Performance DB return empty', async () => {
    global.$fetch = vi.fn(async (url: string) => {
      const u = String(url)
      if (u.includes('performance-trend/batch')) {
        return {
          auditee_forecast: {
            kpi_forecasts: [
              { code: 'KPI-001', kpiName: 'NPL Ratio', unit: '%', entity: 'Surabaya Branch', riskLevel: 'HIGH', currentValue: 3.1, forecastedValue: 4.2, trend: 'Deteriorating', history: [{ period: '2026-01', value: 3.1 }], forecast: [{ period: '2026-02', value: 4.2 }] }
            ],
            at_risk_departments: [],
            time_series_data: [],
            forecast_accuracy: { mape: 3.82, rmse: 1.45, r2Score: 0.942 }
          },
          auditor_forecast: { kpi_forecasts: [], at_risk_pillars: [], time_series_data: [] }
        }
      }
      if (u.includes('/strategic-plans')) return { status: 'success', data: { items: [], pagination: { page: 1, page_size: 100, total: 0, total_pages: 1 } } }
      if (u.includes('/performance/kpi')) return { status: 'success', data: { items: [], pagination: { page: 1, page_size: 100, total: 0, total_pages: 1 } } }
      return null
    })

    const { timeseriesState, fetchAiAnalytics } = useAiAnalytics()
    await fetchAiAnalytics(true)

    expect(timeseriesState.value.auditorForecast.kpiForecasts).toEqual([])
    expect(timeseriesState.value.auditorForecast.timeSeriesData).toEqual([])
    expect(timeseriesState.value.auditorForecast.atRiskPillars).toEqual([])
    expect(timeseriesState.value.auditeeForecast.kpiForecasts).toHaveLength(1)
  })

  it('paginates >100 Strategic Audit Plan rows, scores them via /performance-trend/auditor-forecast, and keeps no riskLevel', async () => {
    const page1Items = Array.from({ length: 100 }, (_, i) => ({
      code: `SO-IA${String(i + 1).padStart(2, '0')}`,
      strategicObjective: `Objective ${i + 1}`,
      kpi: `Audit KPI ${i + 1}`,
      category: 'Pelaksanaan Audit',
      unit: '%',
      hibHig: 'HIG',
      selectedPeriod: '2025',
      target: '90',
      actual: String(80 + (i % 15))
    }))
    const page2Items = [
      {
        code: 'SO-IA101',
        strategicObjective: 'Objective 101',
        kpi: 'Audit KPI 101 (Page 2)',
        category: 'Investigasi',
        unit: 'Day',
        hibHig: 'HIB',
        selectedPeriod: '2025',
        target: '5',
        actual: '4'
      }
    ]

    let postedAuditorCount = 0
    global.$fetch = vi.fn(async (url: string, opts?: any) => {
      const u = String(url)
      if (u.includes('/strategic-plans') && u.includes('page=1')) {
        return { status: 'success', data: { items: page1Items, pagination: { page: 1, page_size: 100, total: 101, total_pages: 2 } } }
      }
      if (u.includes('/strategic-plans') && u.includes('page=2')) {
        return { status: 'success', data: { items: page2Items, pagination: { page: 2, page_size: 100, total: 101, total_pages: 2 } } }
      }
      if (u.includes('/performance/kpi')) {
        return { status: 'success', data: { items: [], pagination: { page: 1, page_size: 100, total: 0, total_pages: 1 } } }
      }
      if (u.includes('/performance-trend/auditor-forecast')) {
        postedAuditorCount = opts?.body?.items?.length || 0
        return {
          status: 'success',
          data: buildAuditorLocalForecastsFromItems(opts.body.items, 2026)
        }
      }
      return null
    })

    const { timeseriesState, syncAuditorKpiWithStrategicPlan } = useAiAnalytics()
    const result = await syncAuditorKpiWithStrategicPlan()

    expect(postedAuditorCount).toBe(101)
    expect(result.kpiForecasts).toHaveLength(101)
    expect(result.kpiForecasts[100].code).toBe('SO-IA101')
    expect(result.kpiForecasts[100].hibHig).toBe('HIB')
    expect(result.kpiForecasts.some((k: any) => 'riskLevel' in k)).toBe(false)
    expect(timeseriesState.value.auditorForecast.kpiForecasts).toHaveLength(101)
    expect(timeseriesState.value.auditorForecast.timeSeriesData.length).toBeGreaterThan(0)
  })
})

describe('100% Chart <-> Table Synchronization for both Auditor and Auditee tabs', () => {
  it('synchronizes Auditor chart series and native units for both __ALL__ aggregate and clicked table row', () => {
    const items = dedupAndBuildAuditorPayloadItems([
      {
        code: 'SO-IA01',
        kpi: 'Realisasi PKAT',
        category: 'Execution',
        unit: '%',
        hibHig: 'HIG',
        target: '90',
        actual: '84',
        kpiTargets: { '2024': '85', '2025': '90', '2026': '95' },
        kpiActuals: { '2024': '80', '2025': '84' }
      },
      {
        code: 'SO-IA02',
        kpi: 'Siklus Penerbitan LHA',
        category: 'Quality',
        unit: 'Day',
        hibHig: 'HIB',
        target: '10',
        actual: '12',
        kpiTargets: { '2024': '10', '2025': '10', '2026': '8' },
        kpiActuals: { '2024': '11', '2025': '12' }
      }
    ], [], 2026)
    const { kpiForecasts } = buildAuditorLocalForecastsFromItems(items, 2026)

    // Aggregate view reflects ALL filtered rows in % achievement
    const aggView = buildAuditorChartView(kpiForecasts, AUDITOR_CHART_ALL_KEY)
    expect(aggView.mode).toBe('aggregate')
    expect(aggView.unit).toBe('%')
    expect(aggView.points.some(p => p.period === '2025' && p.actual !== null)).toBe(true)
    expect(aggView.points.some(p => p.period === '2026' && p.forecast !== null)).toBe(true)

    // Narrowing table rows dynamically updates aggregate chart points
    const filteredOnlyGood = [kpiForecasts[0]]
    const filteredAgg = buildAuditorAggregateTimeSeries(filteredOnlyGood)
    expect(filteredAgg.find(p => p.period === '2025')?.actual).toBe(kpiForecasts[0].achievementRate)

    // Selecting row SO-IA02 plots its exact native unit ('Day'), history, forecast, and dynamic targets
    const rowView = buildAuditorChartView(kpiForecasts, 'SO-IA02')
    expect(rowView.mode).toBe('single')
    expect(rowView.unit).toBe('Day')
    expect(rowView.targetValue).toBe(10)
    expect(rowView.points.map(p => p.period)).toEqual(['2024', '2025', '2026'])
    expect(rowView.points[0].actual).toBe(11)
    expect(rowView.points[1].actual).toBe(12)
    expect(rowView.points[1].forecast).toBe(12)
    expect(rowView.points[2].actual).toBeNull()
    expect(rowView.points[2].forecast).toBe(kpiForecasts[1].forecastedValue)
    expect(rowView.points[2].target).toBe(8)
  })

  it('synchronizes Auditee chart strictly with the selected filtered Auditee row via useKpiForecast', () => {
    const tableRows = [
      {
        code: 'KPI-010',
        kpiName: 'Customer Lifetime Value',
        unit: '%',
        entity: 'Retail Lending',
        riskLevel: 'HIGH',
        currentValue: 1200,
        forecastedValue: 1120,
        lastActualPeriod: '2026-02',
        history: [{ period: '2026-01', value: 1250 }, { period: '2026-02', value: 1200 }],
        forecast: [{ period: '2026-03', value: 1150 }, { period: '2026-04', value: 1120 }]
      }
    ]
    const views = tableRows.map(toKpiForecastView)
    expect(selectAtRiskKpis(views)).toHaveLength(1)
    const chart = buildKpiForecastChart(views[0], 3.82, t, 'id')
    expect(chart.data.labels).toEqual(['Jan 2026', 'Feb 2026', 'Mar 2026', 'Apr 2026'])
    expect(chart.data.datasets[0].data).toEqual([1250, 1200, null, null])
    expect(chart.data.datasets[1].data).toEqual([null, 1200, 1150, 1120])
  })

  it('removes legacy hardcoded Auditor mock numbers/fallbacks from kpi-forecast.vue and useAiAnalytics.ts', () => {
    const pageSrc = readFileSync(KPI_FORECAST_PAGE, 'utf-8')
    const composableSrc = readFileSync(USE_AI_ANALYTICS, 'utf-8')

    expect(pageSrc).not.toContain('Q1 2025')
    expect(pageSrc).not.toContain('Akhir Tahun 2026 (Pred)')
    expect(pageSrc).not.toContain('auditorForecasts.length || 12')
    expect(pageSrc).not.toContain('auditeeForecasts.length || 37')
    expect(pageSrc).toContain('selectedAuditorKpiCode')
    expect(pageSrc).toContain('buildAuditorChartView')

    expect(composableSrc).not.toContain('initialTS.auditorKPIForecasts')
    expect(composableSrc).not.toContain('initialTS.auditorTimeSeries')
    expect(composableSrc).toContain('/performance-trend/auditor-forecast')
  })
})

describe('Manual / Force Sync & Anti-Spam Retrain Controls', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('produces bit-identical SHA-256[:16] fingerprints to Python retrain_engine.py', () => {
    const items = [
      {
        code: 'SO-IA01',
        kpi: 'Realisasi PKAT',
        polarity: 'HIG',
        target: 90,
        actual: 85,
        kpiTargets: { '2024': '85', '2025': '90' },
        kpiActuals: { '2024': '82', '2025': '85' }
      }
    ]
    const ext = [
      { code: 'KPI-EXT-01', kpiName: 'NPL Ratio Ext', entity: 'Surabaya', currentValue: 3.5, forecastedValue: 4.1 }
    ]

    expect(computeAuditorFingerprint(items)).toBe('ddc1b411ca70fddf')
    expect(computeAuditeeCustomFingerprint(ext)).toBe('8d6e301d45861744')
  })

  it('disables manual retrain when AI is already running in background and shows status; enables when untrained data exists and disables again after manual force sync', async () => {
    const strategicPlanRows = [
      {
        code: 'SO-IA01',
        kpi: 'Realisasi PKAT',
        category: 'Execution',
        unit: '%',
        hibHig: 'HIG',
        target: '90',
        actual: '85'
      }
    ]
    const auditorHash = computeAuditorFingerprint(dedupAndBuildAuditorPayloadItems(strategicPlanRows, [], 2026))
    let serverRunning = true
    let serverTrainedAuditorHash = ''
    let forceCalls = 0
    let lastForceScope = ''
    let backgroundCallsAfterForce = 0

    global.$fetch = vi.fn(async (url: string, opts?: any) => {
      const u = String(url)
      if (u.includes('performance-trend/batch')) {
        return {
          auditee_forecast: {
            kpi_forecasts: [{ code: 'KPI-001', kpiName: 'NPL', entity: 'Surabaya', riskLevel: 'HIGH', currentValue: 3.1, forecastedValue: 3.5 }],
            at_risk_departments: [],
            time_series_data: [],
            forecast_accuracy: { mape: 3.82, rmse: 1.45, r2Score: 0.942 }
          },
          auditor_forecast: { kpi_forecasts: [] }
        }
      }
      if (u.includes('/strategic-plans')) {
        return { status: 'success', data: { items: strategicPlanRows, pagination: { page: 1, page_size: 100, total: 1, total_pages: 1 } } }
      }
      if (u.includes('/performance/kpi')) {
        return { status: 'success', data: { items: [], pagination: { page: 1, page_size: 100, total: 0, total_pages: 1 } } }
      }
      if (u.includes('/performance-trend/auditor-forecast')) {
        return { status: 'success', data: buildAuditorLocalForecastsFromItems(opts.body.items, 2026) }
      }
      if (u.includes('/retrain/status')) {
        return {
          status: serverRunning ? 'running' : 'completed',
          is_running: serverRunning,
          trained_auditor_hash: serverTrainedAuditorHash,
          trained_auditee_pool_hash: 'pool123',
          trained_auditee_custom_hash: '',
          has_untrained_auditee_pool_data: false,
          trained_auditor_count: serverTrainedAuditorHash ? 1 : 0,
          trained_auditee_count: 100
        }
      }
      if (u.includes('/retrain/auto') && (u.includes('force=true') || opts?.body?.force === true)) {
        forceCalls++
        lastForceScope = opts?.body?.scope || ''
        serverRunning = false
        serverTrainedAuditorHash = computeAuditorFingerprint(opts.body.auditor_items || [])
        return {
          status: 'completed',
          is_running: false,
          duration_seconds: 1.8,
          trained_auditor_hash: serverTrainedAuditorHash,
          trained_auditee_pool_hash: 'pool123',
          trained_auditee_custom_hash: '',
          trained_auditor_count: 1,
          trained_auditee_count: 100,
          message: 'Model selesai dilatih ulang.'
        }
      }
      if (u.includes('/retrain/auto')) {
        if (forceCalls > 0) backgroundCallsAfterForce++
        return { status: serverRunning ? 'already_running' : 'skipped', is_running: serverRunning }
      }
      return null
    })

    const {
      retrainStatus,
      hasUntrainedNewData,
      untrainedDataReasons,
      canManualRetrain,
      fetchRetrainStatus,
      fetchAiAnalytics,
      triggerManualForceRetrain
    } = useAiAnalytics()

    // 1. While background training is active: button must stay disabled (cannot spam)
    await fetchAiAnalytics(true)
    expect(retrainStatus.value.isRunning).toBe(true)
    expect(canManualRetrain.value).toBe(false)

    // 2. Background finished (e.g., throttled or missed new data): untrained new data alert turns on & button enables
    serverRunning = false
    await fetchRetrainStatus()
    expect(retrainStatus.value.isRunning).toBe(false)
    expect(hasUntrainedNewData.value).toBe(true)
    expect(untrainedDataReasons.value.some(r => r.includes('KPI Auditor'))).toBe(true)
    expect(canManualRetrain.value).toBe(true)

    // 3. Click Manual / Force Sync: trains with live items (scope=kpi), skips re-triggering auto-retrain on refresh, and disables button afterward
    const res = await triggerManualForceRetrain()
    expect(res.status).toBe('completed')
    expect(forceCalls).toBe(1)
    expect(lastForceScope).toBe('kpi')
    expect(backgroundCallsAfterForce).toBe(0)
    expect(retrainStatus.value.trainedAuditorHash).toBe(auditorHash)
    expect(hasUntrainedNewData.value).toBe(false)
    expect(canManualRetrain.value).toBe(false)

    // 4. Subsequent click attempt is blocked by anti-spam guard
    const spamAttempt = await triggerManualForceRetrain()
    expect(spamAttempt.status).toBe('up_to_date')
    expect(forceCalls).toBe(1)
  })

  it('wires anti-spam disabled binding, aligned button/tooltip labels across all 4 states, and cleanup in kpi-forecast.vue', () => {
    const pageSrc = readFileSync(KPI_FORECAST_PAGE, 'utf-8')
    expect(pageSrc).toContain(':disabled="!isManualRetrainAllowed"')
    expect(pageSrc).toContain(':title="retrainButtonTooltip"')
    expect(pageSrc).toContain('Sedang Melatih Model AI...')
    expect(pageSrc).toContain('Model Sudah Terlatih')
    expect(pageSrc).toContain('Status AI Belum Terhubung')
    expect(pageSrc).toContain('Model AI Sedang Dilatih di Belakang Layar')
    expect(pageSrc).toContain('Data Baru Belum Dilatih oleh Model AI')
    expect(pageSrc).toContain('Semua Data Baru Sudah Dilatih oleh Model AI')
    expect(pageSrc).toContain('Status Pelatihan AI Belum Terhubung')
    expect(pageSrc).toContain('onUnmounted')
  })
})
