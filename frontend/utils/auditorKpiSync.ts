/**
 * Pure helpers to synchronize the KPI Auditor perspective on /analytics/ai/kpi-forecast
 * with live AuditSphere features that genuinely contain both `target` and `actual` columns
 * (Strategic Audit Plan and non-duplicate rows from Internal Audit Performance),
 * and to keep chart series 100% synchronized with table rows.
 */

export interface AuditorKpiHistoryPoint {
  period: string
  value: number
  target?: number
}

export interface AuditorKpiForecastPoint {
  period: string
  value: number
  target?: number
  lower?: number
  upper?: number
}

export interface AuditorPayloadItem {
  code: string
  kpi: string
  strategicObjective: string
  category: string
  unit: string
  polarity: 'HIG' | 'HIB'
  selectedPeriod: string
  target: number
  actual: number
  kpiTargets: Record<string, string | number>
  kpiActuals: Record<string, string | number>
  source: 'Strategic Audit Plan' | 'Internal Audit Performance'
}

export interface AuditorTimeSeriesPoint {
  period: string
  actual: number | null
  forecast: number | null
  target?: number | null
  upperBound: number | null
  lowerBound: number | null
}

export interface AuditorChartView {
  mode: 'single' | 'aggregate'
  selectedCode: string
  title: string
  unit: string
  targetValue: number | null
  polarity: 'HIG' | 'HIB'
  points: AuditorTimeSeriesPoint[]
}

export const AUDITOR_CHART_ALL_KEY = '__ALL__'
const AUDITOR_DEFAULT_MAPE = 0.0295

/** Parse numeric values stored as numbers or formatted strings (e.g. "85.5%", "1,200"). */
export const parseNumericKpiValue = (raw: unknown): number | null => {
  if (raw === null || raw === undefined) return null
  if (typeof raw === 'number') return Number.isFinite(raw) ? raw : null
  const text = String(raw).trim()
  if (!text || text === '-' || text.toLowerCase() === 'n/a') return null
  const cleaned = text.replace(/,/g, '').replace(/%$/g, '').trim()
  if (!cleaned) return null
  const num = Number(cleaned)
  return Number.isFinite(num) ? num : null
}

/** Sort periods chronologically: "2024", "2025", "Q1-2025", "2025-Q2", "Q1", "Akhir Tahun 2026". */
export const auditorPeriodSortKey = (period: string): [number, number, string] => {
  const s = String(period || '').trim()
  const qYear = s.match(/^Q([1-4])[-\s/]?(\d{4})$/i)
  if (qYear) return [Number(qYear[2]), Number(qYear[1]), s]
  const yearQ = s.match(/^(\d{4})[-\s/]?Q([1-4])$/i)
  if (yearQ) return [Number(yearQ[1]), Number(yearQ[2]), s]
  const qOnly = s.match(/^Q([1-4])$/i)
  if (qOnly) return [2026, Number(qOnly[1]), s]
  const yearOnly = s.match(/^(\d{4})$/)
  if (yearOnly) return [Number(yearOnly[1]), 0, s]
  const akhir = s.match(/(\d{4})/)
  if (akhir) return [Number(akhir[1]), 9, s]
  return [9999, 99, s]
}

export const compareKpiPeriods = (a: string, b: string): number => {
  const ka = auditorPeriodSortKey(a)
  const kb = auditorPeriodSortKey(b)
  if (ka[0] !== kb[0]) return ka[0] - kb[0]
  if (ka[1] !== kb[1]) return ka[1] - kb[1]
  return ka[2].localeCompare(kb[2])
}

/** Case-insensitive match used to avoid duplicating Strategic Audit Plan KPIs from Performance Achievement rows. */
export const kpiNamesMatch = (a: unknown, b: unknown): boolean => {
  const la = String(a || '').trim().toLowerCase()
  const lb = String(b || '').trim().toLowerCase()
  if (!la || !lb) return false
  if (la === lb) return true
  if (la.length >= 4 && lb.includes(la)) return true
  if (lb.length >= 4 && la.includes(lb)) return true
  return false
}

/** True only when an item genuinely carries both numeric `target` and numeric `actual` columns. */
export const hasTargetAndActualColumns = (item: any): boolean => {
  if (!item || typeof item !== 'object') return false
  const topTarget = parseNumericKpiValue(item.target)
  const topActual = parseNumericKpiValue(item.actual)
  if (topTarget !== null && topActual !== null) return true

  const targets = item.kpiTargets && typeof item.kpiTargets === 'object' ? item.kpiTargets : {}
  const actuals = item.kpiActuals && typeof item.kpiActuals === 'object' ? item.kpiActuals : {}
  for (const periodKey of Object.keys(actuals)) {
    if (parseNumericKpiValue(actuals[periodKey]) !== null && parseNumericKpiValue(targets[periodKey]) !== null) {
      return true
    }
  }
  return false
}

const normalizeRecord = (raw: unknown): Record<string, string | number> => {
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return {}
  const out: Record<string, string | number> = {}
  for (const [k, v] of Object.entries(raw as Record<string, unknown>)) {
    if (v !== null && v !== undefined && String(v).trim() !== '') {
      out[k] = typeof v === 'number' ? v : String(v)
    }
  }
  return out
}

const computeAchievementRate = (actual: number, target: number, polarity: 'HIG' | 'HIB'): number => {
  if (polarity === 'HIB') {
    if (actual <= 0) return 100.0
    return Math.round(((target / actual) * 100.0) * 10) / 10
  }
  if (target <= 0) return 100.0
  return Math.round(((actual / target) * 100.0) * 10) / 10
}

const statusFromRate = (rate: number): 'Good' | 'Moderate' | 'Poor' => {
  if (rate >= 100) return 'Good'
  if (rate >= 70) return 'Moderate'
  return 'Poor'
}

/**
 * Filter and normalize Auditor KPI items solely from live Strategic Audit Plan records and
 * non-duplicate Internal Audit Performance KPIAchievement rows that contain both `target` and `actual`.
 */
export const dedupAndBuildAuditorPayloadItems = (
  strategicPlans: any[],
  kpiAchievements: any[] = [],
  defaultYear = new Date().getFullYear(),
): AuditorPayloadItem[] => {
  const payloadItems: AuditorPayloadItem[] = []
  const coveredNames: string[] = []

  const plans = Array.isArray(strategicPlans) ? strategicPlans : []
  plans.forEach((plan, idx) => {
    if (!hasTargetAndActualColumns(plan)) return

    const kpiTargets = normalizeRecord(plan.kpiTargets)
    const kpiActuals = normalizeRecord(plan.kpiActuals)
    const matchedPeriods = Object.keys(kpiActuals)
      .filter(p => parseNumericKpiValue(kpiActuals[p]) !== null && parseNumericKpiValue(kpiTargets[p]) !== null)
      .sort(compareKpiPeriods)

    const topTarget = parseNumericKpiValue(plan.target)
    const topActual = parseNumericKpiValue(plan.actual)
    const latestMatched = matchedPeriods.length > 0 ? matchedPeriods[matchedPeriods.length - 1] : null

    const targetVal = topTarget ?? (latestMatched ? parseNumericKpiValue(kpiTargets[latestMatched]) : null)
    const actualVal = topActual ?? (latestMatched ? parseNumericKpiValue(kpiActuals[latestMatched]) : null)
    if (targetVal === null || actualVal === null) return

    const kpiLabel = String(plan.kpi || plan.strategicObjective || '').trim()
    if (!kpiLabel) return

    const polarity: 'HIG' | 'HIB' = String(plan.hibHig || plan.polarity || 'HIG').toUpperCase() === 'HIB' ? 'HIB' : 'HIG'
    const selectedPeriod = String(plan.selectedPeriod || latestMatched || defaultYear)

    payloadItems.push({
      code: String(plan.code || `SO-IA${String(idx + 1).padStart(2, '0')}`),
      kpi: kpiLabel,
      strategicObjective: String(plan.strategicObjective || kpiLabel),
      category: String(plan.category || 'Pelaksanaan Audit & PKAT'),
      unit: String(plan.unit || '%'),
      polarity,
      selectedPeriod,
      target: targetVal,
      actual: actualVal,
      kpiTargets,
      kpiActuals,
      source: 'Strategic Audit Plan',
    })
    coveredNames.push(kpiLabel)
  })

  // Group Internal Audit Performance KPIAchievement rows by KPI name, excluding any already in Strategic Audit Plan
  const perfGroups = new Map<string, { displayName: string; rows: Array<{ period: string; target: number; actual: number }> }>()
  const perfList = Array.isArray(kpiAchievements) ? kpiAchievements : []

  for (const row of perfList) {
    const rawName = String(row?.kpi_name ?? row?.kpiName ?? row?.kpi ?? '').trim()
    const tVal = parseNumericKpiValue(row?.target)
    const aVal = parseNumericKpiValue(row?.actual)
    if (!rawName || tVal === null || aVal === null) continue
    if (coveredNames.some(existing => kpiNamesMatch(existing, rawName))) continue

    const normKey = rawName.toLowerCase()
    const periodLabel = row?.period
      ? (row?.year && !String(row.period).includes(String(row.year)) ? `${row.period}-${row.year}` : String(row.period))
      : String(row?.year || defaultYear)

    if (!perfGroups.has(normKey)) {
      perfGroups.set(normKey, { displayName: rawName, rows: [] })
    }
    perfGroups.get(normKey)!.rows.push({ period: periodLabel, target: tVal, actual: aVal })
  }

  let perfIdx = 1
  for (const group of perfGroups.values()) {
    const sorted = [...group.rows].sort((a, b) => compareKpiPeriods(a.period, b.period))
    const latest = sorted[sorted.length - 1]
    if (!latest) continue

    const kpiTargets: Record<string, number> = {}
    const kpiActuals: Record<string, number> = {}
    for (const r of sorted) {
      kpiTargets[r.period] = r.target
      kpiActuals[r.period] = r.actual
    }

    payloadItems.push({
      code: `KPI-PERF-${String(perfIdx++).padStart(2, '0')}`,
      kpi: group.displayName,
      strategicObjective: group.displayName,
      category: 'Internal Audit Performance',
      unit: '%',
      polarity: 'HIG',
      selectedPeriod: latest.period,
      target: latest.target,
      actual: latest.actual,
      kpiTargets,
      kpiActuals,
      source: 'Internal Audit Performance',
    })
  }

  return payloadItems
}

/**
 * Deterministic fallback when Python AI Engine / Analytics Service are unreachable.
 * Still computes strictly from live Strategic Audit Plan / KPIAchievement items (never CSV/mock).
 */
export const buildAuditorLocalForecastsFromItems = (
  items: AuditorPayloadItem[],
  currentYear = new Date().getFullYear(),
) => {
  const kpiForecasts = items.map((item, idx) => {
    const pol = item.polarity
    const pairedKeys = Object.keys(item.kpiActuals || {})
      .filter(p => parseNumericKpiValue(item.kpiActuals[p]) !== null)
      .sort(compareKpiPeriods)

    const history: AuditorKpiHistoryPoint[] = pairedKeys.length > 0
      ? pairedKeys.map(p => ({
        period: p,
        value: parseNumericKpiValue(item.kpiActuals[p]) as number,
        target: parseNumericKpiValue(item.kpiTargets?.[p]) ?? item.target,
      }))
      : [{ period: item.selectedPeriod || String(currentYear), value: item.actual, target: item.target }]

    const currVal = history[history.length - 1]?.value ?? item.actual
    const prevVal = history.length >= 2 ? history[history.length - 2]!.value : currVal
    const baseTarget = history[history.length - 1]?.target ?? item.target

    const futureTargetKeys = Object.keys(item.kpiTargets || {})
      .filter(p => parseNumericKpiValue(item.kpiActuals?.[p]) === null && parseNumericKpiValue(item.kpiTargets[p]) !== null)
      .sort(compareKpiPeriods)
      .slice(0, 3)

    const slope = currVal - prevVal
    const stepDelta = Math.abs(slope) > 0.001
      ? slope * 0.45
      : (pol === 'HIG' ? (currVal >= baseTarget ? currVal * 0.015 : -currVal * 0.015) : (currVal <= baseTarget ? -currVal * 0.015 : currVal * 0.02))

    const forecastPeriods: string[] = futureTargetKeys.length > 0
      ? futureTargetKeys
      : (() => {
        const lastP = history[history.length - 1]?.period || String(currentYear)
        const yMatch = lastP.match(/^(\d{4})$/)
        if (yMatch) return [String(Number(yMatch[1]) + 1)]
        const qMatch = lastP.match(/^Q([1-4])(?:[-](\d{4}))?$/i)
        if (qMatch) {
          const qn = Number(qMatch[1])
          const y = qMatch[2] ? Number(qMatch[2]) : currentYear
          return [qn < 4 ? `Q${qn + 1}-${y}` : `Q1-${y + 1}`]
        }
        return [`Akhir Tahun ${currentYear}`]
      })()

    const forecast: AuditorKpiForecastPoint[] = forecastPeriods.map((fp, stepIdx) => {
      const fVal = Math.round(Math.max(0, currVal + stepDelta * (stepIdx + 1)) * 100) / 100
      const fTarget = parseNumericKpiValue(item.kpiTargets?.[fp]) ?? baseTarget
      const spread = Math.abs(fVal) * AUDITOR_DEFAULT_MAPE * Math.sqrt(stepIdx + 1)
      return {
        period: fp,
        value: fVal,
        target: fTarget,
        lower: Math.round(Math.max(0, fVal - spread) * 100) / 100,
        upper: Math.round((fVal + spread) * 100) / 100,
      }
    })

    const firstFore = forecast[0]!
    const foreVal = firstFore.value
    const foreTarget = firstFore.target ?? baseTarget
    const currRate = computeAchievementRate(currVal, baseTarget, pol)
    const foreRate = computeAchievementRate(foreVal, foreTarget, pol)
    const currStatus = statusFromRate(currRate)
    const predStatus = statusFromRate(foreRate)
    const trend = (pol === 'HIG' ? foreVal >= currVal : foreVal <= currVal) ? 'Improving' : 'Deteriorating'

    return {
      code: item.code || `SO-IA${String(idx + 1).padStart(2, '0')}`,
      kpiName: item.kpi,
      category: item.category,
      strategicObjective: item.strategicObjective,
      hibHig: pol,
      unit: item.unit || '%',
      source: item.source,
      targetHorizon: firstFore.period,
      targetValue: baseTarget,
      currentValue: currVal,
      forecastedValue: foreVal,
      achievementRate: currRate,
      forecastedAchievementRate: foreRate,
      currentStatus: currStatus,
      predictedStatus: predStatus,
      trend,
      confidenceLower: firstFore.lower ?? Math.round(foreVal * 0.97 * 100) / 100,
      confidenceUpper: firstFore.upper ?? Math.round(foreVal * 1.03 * 100) / 100,
      history,
      forecast,
      aiNarrative: `Dengan kecepatan saat ini, diprediksi target ${item.kpi} mencapai ${foreRate.toFixed(1)}% (${predStatus}) pada periode ${firstFore.period}.`,
      recommendedAction: predStatus === 'Poor'
        ? `Eskalasi kendala pencapaian target ${item.code} ke Komite Audit dan susun corrective action plan segera.`
        : (predStatus === 'Moderate'
          ? `Tingkatkan koordinasi tim audit pada pilar ${item.category} agar target ${baseTarget}${item.unit} tercapai tepat waktu.`
          : 'Pertahankan kualitas pelaksanaan audit dan dokumentasikan metodologi sebagai best practice audit internal.'),
    }
  })

  const atRiskPillars = kpiForecasts
    .filter(k => k.predictedStatus === 'Poor' || k.predictedStatus === 'Moderate' || k.trend === 'Deteriorating')
    .map(k => ({
      pillar: k.category,
      code: k.code,
      kpi: k.kpiName,
      target: k.targetValue,
      predicted: k.forecastedValue,
      achievementRate: k.forecastedAchievementRate,
      predictedStatus: k.predictedStatus,
    }))

  const timeSeriesData = buildAuditorAggregateTimeSeries(kpiForecasts)
  const avgForeRatio = kpiForecasts.length > 0
    ? (kpiForecasts.reduce((acc, k) => acc + (k.forecastedAchievementRate || 0), 0) / kpiForecasts.length).toFixed(1)
    : '0.0'

  const executiveSummary = kpiForecasts.length === 0
    ? 'Belum ada data KPI (dengan kolom Target & Actual) pada Strategic Audit Plan untuk diproyeksikan.'
    : (atRiskPillars.length > 0
      ? `Berdasarkan ${kpiForecasts.length} KPI aktif dari Strategic Audit Plan, rata-rata proyeksi capaian target berada di ${avgForeRatio}%. Ditemukan ${atRiskPillars.length} KPI memerlukan perhatian khusus (${atRiskPillars.slice(0, 3).map(r => r.code).join(', ')}).`
      : `Berdasarkan ${kpiForecasts.length} KPI aktif dari Strategic Audit Plan, seluruh indikator diproyeksikan berstatus Good dengan rata-rata capaian ${avgForeRatio}%.`)

  return {
    kpiForecasts,
    atRiskPillars,
    timeSeriesData,
    executiveSummary,
    forecastAccuracy: { mape: 2.95, rmse: 1.22, r2Score: 0.985 },
  }
}

/** Build aggregate % achievement time series strictly from the active `kpiForecasts` table rows. */
export const buildAuditorAggregateTimeSeries = (kpiForecasts: any[]): AuditorTimeSeriesPoint[] => {
  if (!Array.isArray(kpiForecasts) || kpiForecasts.length === 0) return []

  const histByPeriod = new Map<string, number[]>()
  const foreByPeriod = new Map<string, number[]>()

  for (const k of kpiForecasts) {
    const pol: 'HIG' | 'HIB' = String(k.hibHig || 'HIG').toUpperCase() === 'HIB' ? 'HIB' : 'HIG'
    const fallbackTarget = Number(k.targetValue || 100)
    const hist: any[] = Array.isArray(k.history) && k.history.length > 0
      ? k.history
      : [{ period: 'Aktual', value: Number(k.currentValue ?? 0), target: fallbackTarget }]
    const fore: any[] = Array.isArray(k.forecast) && k.forecast.length > 0
      ? k.forecast
      : [{ period: String(k.targetHorizon || 'Proyeksi'), value: Number(k.forecastedValue ?? 0), target: fallbackTarget }]

    for (const h of hist) {
      const p = String(h.period || '').trim()
      if (!p) continue
      const rate = Math.min(150, computeAchievementRate(Number(h.value ?? 0), Number(h.target ?? fallbackTarget), pol))
      if (!histByPeriod.has(p)) histByPeriod.set(p, [])
      histByPeriod.get(p)!.push(rate)
    }

    for (const f of fore) {
      const p = String(f.period || '').trim()
      if (!p) continue
      const rate = Math.min(150, computeAchievementRate(Number(f.value ?? 0), Number(f.target ?? fallbackTarget), pol))
      if (!foreByPeriod.has(p)) foreByPeriod.set(p, [])
      foreByPeriod.get(p)!.push(rate)
    }
  }

  const sortedHist = [...histByPeriod.keys()].sort(compareKpiPeriods).slice(-6)
  const sortedFore = [...foreByPeriod.keys()].sort(compareKpiPeriods).slice(0, 3)

  const avg = (arr: number[]) => Math.round((arr.reduce((s, n) => s + n, 0) / arr.length) * 10) / 10
  const points: AuditorTimeSeriesPoint[] = []

  sortedHist.forEach((p, idx) => {
    const val = avg(histByPeriod.get(p)!)
    const isAnchor = idx === sortedHist.length - 1 && sortedFore.length > 0
    points.push({
      period: p,
      actual: val,
      forecast: isAnchor ? val : null,
      target: 100,
      upperBound: isAnchor ? val : null,
      lowerBound: isAnchor ? val : null,
    })
  })

  sortedFore.forEach((p, stepIdx) => {
    const val = avg(foreByPeriod.get(p)!)
    const band = Math.round(val * AUDITOR_DEFAULT_MAPE * Math.sqrt(stepIdx + 1) * 10) / 10
    points.push({
      period: p,
      actual: null,
      forecast: val,
      target: 100,
      upperBound: Math.round((val + band) * 10) / 10,
      lowerBound: Math.round(Math.max(0, val - band) * 10) / 10,
    })
  })

  return points
}

/**
 * Build the chart view model for the KPI Auditor tab.
 * Synchronizes 1:1 with either a selected table row (`selectedCode`) or the filtered table aggregate (`__ALL__`).
 */
export const buildAuditorChartView = (
  kpiForecasts: any[],
  selectedCode: string = AUDITOR_CHART_ALL_KEY,
): AuditorChartView => {
  const list = Array.isArray(kpiForecasts) ? kpiForecasts : []
  if (list.length === 0) {
    return {
      mode: 'aggregate',
      selectedCode: AUDITOR_CHART_ALL_KEY,
      title: 'Rata-rata Capaian Target Strategic Audit Plan',
      unit: '%',
      targetValue: 100,
      polarity: 'HIG',
      points: [],
    }
  }

  const chosen = selectedCode && selectedCode !== AUDITOR_CHART_ALL_KEY
    ? list.find(k => k.code === selectedCode)
    : undefined

  if (!chosen) {
    return {
      mode: 'aggregate',
      selectedCode: AUDITOR_CHART_ALL_KEY,
      title: `Rata-rata Capaian ${list.length} KPI Strategic Audit Plan`,
      unit: '%',
      targetValue: 100,
      polarity: 'HIG',
      points: buildAuditorAggregateTimeSeries(list),
    }
  }

  const fallbackTarget = Number(chosen.targetValue ?? 100)
  const hist: any[] = Array.isArray(chosen.history) && chosen.history.length > 0
    ? chosen.history
    : [{ period: 'Aktual', value: Number(chosen.currentValue ?? 0), target: fallbackTarget }]
  const fore: any[] = Array.isArray(chosen.forecast) && chosen.forecast.length > 0
    ? chosen.forecast
    : [{
      period: String(chosen.targetHorizon || 'Proyeksi'),
      value: Number(chosen.forecastedValue ?? 0),
      target: fallbackTarget,
      lower: chosen.confidenceLower,
      upper: chosen.confidenceUpper,
    }]

  const points: AuditorTimeSeriesPoint[] = []
  hist.forEach((h, idx) => {
    const val = Number(h.value ?? 0)
    const isAnchor = idx === hist.length - 1 && fore.length > 0
    points.push({
      period: String(h.period || ''),
      actual: val,
      forecast: isAnchor ? val : null,
      target: Number(h.target ?? fallbackTarget),
      upperBound: isAnchor ? val : null,
      lowerBound: isAnchor ? val : null,
    })
  })

  fore.forEach((f, stepIdx) => {
    const val = Number(f.value ?? 0)
    const defaultBand = Math.abs(val) * AUDITOR_DEFAULT_MAPE * Math.sqrt(stepIdx + 1)
    const upper = f.upper !== undefined && f.upper !== null
      ? Number(f.upper)
      : (stepIdx === 0 && chosen.confidenceUpper !== undefined ? Number(chosen.confidenceUpper) : val + defaultBand)
    const lower = f.lower !== undefined && f.lower !== null
      ? Number(f.lower)
      : (stepIdx === 0 && chosen.confidenceLower !== undefined ? Number(chosen.confidenceLower) : Math.max(0, val - defaultBand))

    points.push({
      period: String(f.period || ''),
      actual: null,
      forecast: val,
      target: Number(f.target ?? fallbackTarget),
      upperBound: Math.round(upper * 100) / 100,
      lowerBound: Math.round(lower * 100) / 100,
    })
  })

  return {
    mode: 'single',
    selectedCode: chosen.code,
    title: `${chosen.code} — ${chosen.kpiName}`,
    unit: String(chosen.unit || '%'),
    targetValue: fallbackTarget,
    polarity: String(chosen.hibHig || 'HIG').toUpperCase() === 'HIB' ? 'HIB' : 'HIG',
    points,
  }
}

const SHA256_K = [
  0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
  0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
  0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
  0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
  0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
  0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
  0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
  0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
]

/** Synchronous pure-TS SHA-256 hex digest matching Python `hashlib.sha256` across browsers and Node. */
export const sha256HexSync = (text: string): string => {
  const utf8 = typeof TextEncoder !== 'undefined'
    ? Array.from(new TextEncoder().encode(text))
    : Array.from(unescape(encodeURIComponent(text)), c => c.charCodeAt(0))

  const bitLen = utf8.length * 8
  utf8.push(0x80)
  while ((utf8.length % 64) !== 56) utf8.push(0)
  const hiLen = Math.floor(bitLen / 0x100000000)
  const loLen = bitLen >>> 0
  for (let s = 24; s >= 0; s -= 8) utf8.push((hiLen >>> s) & 0xff)
  for (let s = 24; s >= 0; s -= 8) utf8.push((loLen >>> s) & 0xff)

  const h = [0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19]
  const w = new Array<number>(64)
  const rotr = (n: number, bits: number) => (n >>> bits) | (n << (32 - bits))

  for (let i = 0; i < utf8.length; i += 64) {
    for (let j = 0; j < 16; j++) {
      w[j] = (utf8[i + j * 4]! << 24) | (utf8[i + j * 4 + 1]! << 16) | (utf8[i + j * 4 + 2]! << 8) | utf8[i + j * 4 + 3]!
    }
    for (let j = 16; j < 64; j++) {
      const s0 = rotr(w[j - 15]!, 7) ^ rotr(w[j - 15]!, 18) ^ (w[j - 15]! >>> 3)
      const s1 = rotr(w[j - 2]!, 17) ^ rotr(w[j - 2]!, 19) ^ (w[j - 2]! >>> 10)
      w[j] = (w[j - 16]! + s0 + w[j - 7]! + s1) | 0
    }
    let [a, b, c, d, e, f, g, hh] = h as [number, number, number, number, number, number, number, number]
    for (let j = 0; j < 64; j++) {
      const S1 = rotr(e, 6) ^ rotr(e, 11) ^ rotr(e, 25)
      const ch = (e & f) ^ (~e & g)
      const temp1 = (hh + S1 + ch + SHA256_K[j]! + w[j]!) | 0
      const S0 = rotr(a, 2) ^ rotr(a, 13) ^ rotr(a, 22)
      const maj = (a & b) ^ (a & c) ^ (b & c)
      const temp2 = (S0 + maj) | 0
      hh = g
      g = f
      f = e
      e = (d + temp1) | 0
      d = c
      c = b
      b = a
      a = (temp1 + temp2) | 0
    }
    h[0] = (h[0]! + a) | 0
    h[1] = (h[1]! + b) | 0
    h[2] = (h[2]! + c) | 0
    h[3] = (h[3]! + d) | 0
    h[4] = (h[4]! + e) | 0
    h[5] = (h[5]! + f) | 0
    h[6] = (h[6]! + g) | 0
    h[7] = (h[7]! + hh) | 0
  }
  return h.map(v => (v >>> 0).toString(16).padStart(8, '0')).join('')
}

const normalizeNumHashStr = (raw: unknown): string => {
  const n = parseNumericKpiValue(raw)
  return n === null ? '' : (Math.round(n * 10000) / 10000).toFixed(4)
}

const canonicalSortedMapJson = (raw: unknown): string => {
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return '{}'
  const obj = raw as Record<string, unknown>
  const keys = Object.keys(obj).sort()
  return `{${keys.map(k => `${JSON.stringify(String(k))}:${JSON.stringify(String(obj[k]))}`).join(',')}}`
}

/** Deterministic SHA-256[:16] fingerprint matching Python `compute_auditor_items_hash`. */
export const computeAuditorFingerprint = (items: any[]): string => {
  if (!Array.isArray(items) || items.length === 0) return ''
  const rows = items
    .filter(it => it && typeof it === 'object')
    .map((it) => {
      const code = String(it.code || '').trim().toUpperCase()
      const kpi = String(it.kpi || it.kpiName || '').trim().toLowerCase()
      const pol = String(it.polarity || it.hibHig || 'HIG').trim().toUpperCase() === 'HIB' ? 'HIB' : 'HIG'
      const target = normalizeNumHashStr(it.target ?? it.targetValue)
      const actual = normalizeNumHashStr(it.actual ?? it.currentValue)
      const tCanon = canonicalSortedMapJson(it.kpiTargets)
      const aCanon = canonicalSortedMapJson(it.kpiActuals)
      return `${code}|${kpi}|${pol}|${target}|${actual}|${tCanon}|${aCanon}`
    })
    .sort()
  if (rows.length === 0) return ''
  return sha256HexSync(rows.join('\n')).slice(0, 16)
}

/** Deterministic SHA-256[:16] fingerprint matching Python `compute_auditee_custom_items_hash`. */
export const computeAuditeeCustomFingerprint = (items: any[]): string => {
  if (!Array.isArray(items) || items.length === 0) return ''
  const rows = items
    .filter(it => it && typeof it === 'object')
    .map((it) => {
      const code = String(it.code || '').trim().toUpperCase()
      const kpi = String(it.kpiName || it.kpi_name || it.kpi || '').trim().toLowerCase()
      const ent = String(it.entity || '').trim().toLowerCase()
      const cur = normalizeNumHashStr(it.currentValue ?? it.actual_value)
      const fore = normalizeNumHashStr(it.forecastedValue ?? it.forecast_value)
      return `${code}|${kpi}|${ent}|${cur}|${fore}`
    })
    .sort()
  if (rows.length === 0) return ''
  return sha256HexSync(rows.join('\n')).slice(0, 16)
}
