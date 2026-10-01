// @ts-nocheck
/**
 * Transaction Anomaly Detection chart sync.
 *
 * The dashboard renders a "Transaction Anomaly Detection" scatter chart
 * (pages/dashboard/index.vue) and the AI Insights tab renders the same chart
 * for the selected category (pages/analytics/ai/anomaly-detection.vue → the
 * "Transaction" type tab). Both must plot the same points from the same state.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { useAiAnalytics } from '~/composables/useAiAnalytics'
import {
  buildAnomalyScatterPoints,
  listAnomalyTypes,
  resolveAnomalyType
} from '~/composables/useAnomalyScatter'

const DASHBOARD_PAGE = resolve(__dirname, '../../pages/dashboard/index.vue')
const TAB_PAGE = resolve(__dirname, '../../pages/analytics/ai/anomaly-detection.vue')

/** Shape of /anomaly/batch as returned by backend/python-ai/main.py */
const anomalyPayload = {
  anomalies: [
    {
      id: 'TRX-0001',
      entity: 'Cabang Medan',
      type: 'Transaction',
      anomaly_score: -0.93,
      description: 'Transfer beruntun ke rekening baru di luar jam operasional',
      severity: 'Critical',
      date: '2026-06-01',
      amount: 850,
      xMetric: 850,
      is_anomaly: true,
      risk_level: 'HIGH'
    },
    {
      id: 'TRX-0002',
      entity: 'Cabang Bali',
      type: 'Transaction',
      anomaly_score: -0.81,
      description: 'Setoran tunai bulat berulang mendekati threshold pelaporan',
      severity: 'High',
      date: '2026-06-02',
      amount: 500,
      xMetric: 500,
      is_anomaly: true,
      risk_level: 'HIGH'
    },
    {
      id: 'FND-0003',
      entity: 'Cabang Bandung',
      type: 'Funding',
      anomaly_score: -0.77,
      description: 'Suku bunga deposito di atas batas LPS',
      severity: 'High',
      date: '2026-06-03',
      amount: 21465,
      xMetric: 21465,
      is_anomaly: true,
      risk_level: 'HIGH'
    }
  ],
  scatter_data: [
    { x: 120, y: 30.0, type: 'Transaction', is_anomaly: false, label: 'TRX-1001' },
    { x: 95, y: 22.5, type: 'Transaction', is_anomaly: false, label: 'TRX-1002' },
    { x: 850, y: 47.5, type: 'Transaction', is_anomaly: true, label: 'TRX-0001' },
    { x: 60, y: 12.5, type: 'Funding', is_anomaly: false, label: 'FND-1003' },
    { x: 21465, y: 52.0, type: 'Funding', is_anomaly: true, label: 'FND-0003' }
  ],
  summary: {
    totalScanned: 80,
    anomaliesFound: 3,
    contaminationRate: 0.037,
    topCategory: 'Transaction'
  }
}

describe('buildAnomalyScatterPoints', () => {
  const anomalies = [
    { id: 'TRX-0001', type: 'Transaction', xMetric: 850, amount: 850 },
    { id: 'TRX-0002', type: 'Transaction', xMetric: 500, amount: 500 },
    { id: 'TRX-0003', type: 'Transaction', amount: 42_000_000 },
    { id: 'FND-0009', type: 'Funding', xMetric: 21465 }
  ]
  const scatterData = [
    { x: 120, y: 30, type: 'Transaction', isAnomaly: false, label: 'TRX-1001' },
    { x: 95, y: 22.5, type: 'Transaction', isAnomaly: false, label: 'TRX-1002' },
    { x: 850, y: 47.5, type: 'Transaction', isAnomaly: true, label: 'TRX-0001' },
    { x: 60, y: 12.5, type: 'Funding', isAnomaly: false, label: 'FND-1003' }
  ]

  it('plots only the normal points of the requested type', () => {
    const { normalPoints } = buildAnomalyScatterPoints(anomalies, scatterData, 'Transaction')
    expect(normalPoints).toEqual([
      { x: 120, y: 30 },
      { x: 95, y: 22.5 }
    ])
  })

  it('takes anomaly coordinates from the matching scatter point when present', () => {
    const { anomalyPoints } = buildAnomalyScatterPoints(anomalies, scatterData, 'Transaction')
    expect(anomalyPoints[0]).toEqual({ x: 850, y: 47.5 })
  })

  it('falls back to xMetric, then to amount in millions, with y defaulting to 20', () => {
    const { anomalyPoints } = buildAnomalyScatterPoints(anomalies, scatterData, 'Transaction')
    // TRX-0002 has no scatter point → xMetric
    expect(anomalyPoints[1]).toEqual({ x: 500, y: 20 })
    // TRX-0003 has neither scatter point nor xMetric → amount / 1_000_000
    expect(anomalyPoints[2]).toEqual({ x: 42, y: 20 })
  })

  it('never derives coordinates from the anomaly score or the array index', () => {
    const { anomalyPoints } = buildAnomalyScatterPoints(
      [{ id: 'A', type: 'Transaction', xMetric: 10, anomalyScore: -0.9 }],
      [],
      'Transaction'
    )
    expect(anomalyPoints).toEqual([{ x: 10, y: 20 }])
  })

  it('returns empty series for a type that has no data', () => {
    const { normalPoints, anomalyPoints } = buildAnomalyScatterPoints(anomalies, scatterData, 'KYC')
    expect(normalPoints).toEqual([])
    expect(anomalyPoints).toEqual([])
  })

  it('tolerates missing/empty inputs', () => {
    expect(buildAnomalyScatterPoints(undefined, undefined, 'Transaction')).toEqual({
      normalPoints: [],
      anomalyPoints: []
    })
  })
})

describe('anomaly type selection', () => {
  it('lists the types present in anomalies and scatter data', () => {
    const types = listAnomalyTypes(
      [{ type: 'Transaction' }, { type: 'Funding' }],
      [{ type: 'Treasury' }]
    )
    expect(types).toEqual(expect.arrayContaining(['Transaction', 'Funding', 'Treasury']))
  })

  it('prefers the requested type when available, else the first available', () => {
    expect(resolveAnomalyType(['Funding', 'Transaction'], 'Transaction')).toBe('Transaction')
    expect(resolveAnomalyType(['Funding', 'Treasury'], 'Transaction')).toBe('Funding')
    expect(resolveAnomalyType([], 'Transaction')).toBe('Transaction')
  })
})

describe('Transaction Anomaly Detection — dashboard vs AI Insights tab', () => {
  beforeEach(() => {
    localStorage.clear()
    global.$fetch = vi.fn(async (url: string) => {
      if (String(url).includes('anomaly/batch')) return anomalyPayload
      return null
    })
  })

  it('both charts plot the same points for the Transaction category', async () => {
    const { isolationState, fetchAiAnalytics } = useAiAnalytics()
    await fetchAiAnalytics(true)

    const { anomalies, scatterData } = isolationState.value
    const points = buildAnomalyScatterPoints(anomalies, scatterData, 'Transaction')

    // Backend data reached the shared state, and both pages map it through the
    // same builder, so the dashboard card and the tab's Transaction view agree.
    expect(points.anomalyPoints).toEqual([
      { x: 850, y: 47.5 }, // TRX-0001, matched by label
      { x: 500, y: 20 } // TRX-0002, no scatter point → xMetric
    ])
    expect(points.normalPoints).toEqual([
      { x: 120, y: 30 },
      { x: 95, y: 22.5 }
    ])
  })

  it('the dashboard scatter is wired to the shared AI analytics state', () => {
    const src = readFileSync(DASHBOARD_PAGE, 'utf-8')
    expect(src.includes('isolationState'), 'dashboard/index.vue does not read isolationState').toBe(true)
    expect(
      /useIsolationForestData/.test(src),
      'dashboard/index.vue still reads the static useIsolationForestData() mock'
    ).toBe(false)
  })

  it('both pages build the scatter through the shared builder', () => {
    const dashboardSrc = readFileSync(DASHBOARD_PAGE, 'utf-8')
    const tabSrc = readFileSync(TAB_PAGE, 'utf-8')
    expect(
      dashboardSrc.includes('buildAnomalyScatterPoints'),
      'dashboard/index.vue does not use the shared scatter builder'
    ).toBe(true)
    expect(
      tabSrc.includes('buildAnomalyScatterPoints'),
      'anomaly-detection.vue does not use the shared scatter builder'
    ).toBe(true)
  })

  it('the dashboard no longer fabricates coordinates from score and index', () => {
    const src = readFileSync(DASHBOARD_PAGE, 'utf-8')
    const chartBlock = src.slice(
      src.indexOf('const anomalyScatterChartData'),
      src.indexOf('const anomalyScatterOptions')
    )
    expect(
      /anomalyScore \* -100/.test(chartBlock),
      'dashboard scatter still plots x = anomalyScore * -100'
    ).toBe(false)
    expect(
      /\(idx \+ 1\) \*/.test(chartBlock),
      'dashboard scatter still derives y from the array index'
    ).toBe(false)
  })
})
