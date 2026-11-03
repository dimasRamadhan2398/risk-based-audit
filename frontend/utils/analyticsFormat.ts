/**
 * Formatting and KPI semantics shared by the dashboard AI cards and the
 * /analytics/ai pages.
 *
 * - Rupiah amounts are written with Indonesian scale words (juta, miliar,
 *   triliun). "M" is never used: in Indonesian it reads as miliar, so
 *   "Rp 20,114.34M" would be misread as Rp 20 triliun.
 * - Each KPI declares what kind of number it is (percent, score, rupiah) and
 *   whether a higher value is better, so trend, units and "at risk" are
 *   decided in one place instead of assuming every KPI is a percentage.
 */

export type AppLocale = 'id' | 'en' | string

const numberLocale = (locale: AppLocale) => (locale === 'en' ? 'en-US' : 'id-ID')

const formatNumber = (value: number, locale: AppLocale, maxFractionDigits = 2, minFractionDigits = 0) =>
  new Intl.NumberFormat(numberLocale(locale), {
    maximumFractionDigits: maxFractionDigits,
    minimumFractionDigits: minFractionDigits,
  }).format(value)

const RUPIAH_SCALES = [
  { factor: 1e12, id: 'triliun', en: 'trillion' },
  { factor: 1e9, id: 'miliar', en: 'billion' },
  { factor: 1e6, id: 'juta', en: 'million' },
] as const

/**
 * Rp amount with a scale word: 20_114_340_000 → "Rp 20,11 miliar" (id) /
 * "Rp 20.11 billion" (en). Amounts under one million are written in full.
 */
export const formatRupiahCompact = (rupiah: number, locale: AppLocale = 'id', maxFractionDigits = 2): string => {
  if (rupiah === null || rupiah === undefined || Number.isNaN(Number(rupiah))) return '-'
  const value = Number(rupiah)
  const abs = Math.abs(value)
  const sign = value < 0 ? '-' : ''
  const scale = RUPIAH_SCALES.find((s) => abs >= s.factor)
  if (!scale) return `${sign}Rp ${formatNumber(abs, locale, 0)}`
  const word = locale === 'en' ? scale.en : scale.id
  return `${sign}Rp ${formatNumber(abs / scale.factor, locale, maxFractionDigits)} ${word}`
}

/** Same as formatRupiahCompact for a value expressed in millions of rupiah. */
export const formatRupiahFromMillions = (millions: number, locale: AppLocale = 'id', maxFractionDigits = 2) =>
  formatRupiahCompact(Number(millions) * 1e6, locale, maxFractionDigits)

/**
 * Parses a number written either way ("20114.34", "20,114.34", "66.662",
 * "20.114,34"). A lone separator followed by exactly three digits is read as a
 * thousands separator.
 */
export const parseLooseNumber = (raw: string): number => {
  const s = String(raw).trim()
  const hasDot = s.includes('.')
  const hasComma = s.includes(',')
  if (hasDot && hasComma) {
    // The separator that appears last is the decimal one.
    return s.lastIndexOf(',') > s.lastIndexOf('.')
      ? Number(s.replace(/\./g, '').replace(',', '.'))
      : Number(s.replace(/,/g, ''))
  }
  const sep = hasDot ? '.' : hasComma ? ',' : ''
  if (!sep) return Number(s)
  const thousands = new RegExp(`^\\d{1,3}(\\${sep}\\d{3})+$`)
  if (thousands.test(s)) return Number(s.split(sep).join(''))
  return Number(s.replace(sep, '.'))
}

/**
 * Rewrites rupiah amounts inside free text that came from the model or the
 * dataset ("Rp 20114.34 Juta", "Rp 20,114.34M") into the compact form.
 */
export const normalizeRupiahInText = (text: string, locale: AppLocale = 'id'): string => {
  if (!text) return text
  return String(text).replace(
    /Rp\s?([\d][\d.,]*)\s?(Juta|juta|jt|M)\b/g,
    (match, num: string) => {
      const millions = parseLooseNumber(num)
      return Number.isFinite(millions) ? formatRupiahFromMillions(millions, locale) : match
    }
  )
}

// ─── KPI semantics ─────────────────────────────────────────────────────────

export type KpiKind = 'percent' | 'score' | 'rupiah'

export interface KpiMeta {
  /** i18n slug under analytics.kpiAudit.* */
  key: string
  kind: KpiKind
  /** false when a lower value is the better outcome (turnover, acquisition cost). */
  higherIsBetter: boolean
  /** For rupiah KPIs: how many rupiah one unit of the dataset value stands for. */
  rupiahPerUnit?: number
}

/**
 * Units of the KPI training dataset (ai_model_training/kpi_prediction/kpi_data.csv).
 * The dataset stores every value in a column labelled "(%)", but only some of
 * them are percentages: NPS is a -100..+100 score and the monetary KPIs are
 * rupiah amounts. The rupiah scale per KPI is the dataset's convention and
 * lives here only.
 */
const KPI_META: Record<string, KpiMeta> = {
  'Net Promoter Score': { key: 'nps', kind: 'score', higherIsBetter: true },
  'Customer Lifetime Value': { key: 'clv', kind: 'rupiah', higherIsBetter: true, rupiahPerUnit: 1e6 },
  'Average Order Value': { key: 'aov', kind: 'rupiah', higherIsBetter: true, rupiahPerUnit: 1e3 },
  'Customer Acquisition Cost': { key: 'cac', kind: 'rupiah', higherIsBetter: false, rupiahPerUnit: 1e3 },
  'Monthly Recurring Revenue': { key: 'mrr', kind: 'rupiah', higherIsBetter: true, rupiahPerUnit: 1e9 },
  'Sales Conversion Rate': { key: 'conversion', kind: 'percent', higherIsBetter: true },
  'Gross Profit Margin': { key: 'gpm', kind: 'percent', higherIsBetter: true },
  'Customer Retention Rate': { key: 'retention', kind: 'percent', higherIsBetter: true },
  'Return on Investment': { key: 'roi', kind: 'percent', higherIsBetter: true },
  'Employee Turnover Rate': { key: 'turnover', kind: 'percent', higherIsBetter: false },
  'Revenue Operational Cost': { key: 'opCost', kind: 'percent', higherIsBetter: false },
  'Customer Satisfaction Index': { key: 'csi', kind: 'score', higherIsBetter: true },
  'Audit Response Time': { key: 'responseTime', kind: 'score', higherIsBetter: false },
  'Internal Control Effectiveness': { key: 'controlEffectiveness', kind: 'percent', higherIsBetter: true },
  'Compliance Adherence Rate': { key: 'compliance', kind: 'percent', higherIsBetter: true },
  'Risk Mitigation Completion': { key: 'mitigation', kind: 'percent', higherIsBetter: true },
}

export const getKpiMeta = (kpiName: string, fallbackUnit?: string): KpiMeta => {
  const known = KPI_META[String(kpiName || '').trim()]
  if (known) return known
  if (fallbackUnit && fallbackUnit !== '%') {
    return { key: '', kind: /rp|idr/i.test(fallbackUnit) ? 'rupiah' : 'score', higherIsBetter: true, rupiahPerUnit: 1 }
  }
  return { key: '', kind: 'percent', higherIsBetter: true }
}

/** "53,7%", "+53,7" (score) or "Rp 690,67 juta" depending on the KPI. */
export const formatKpiForecastValue = (value: number | null | undefined, meta: KpiMeta, locale: AppLocale = 'id'): string => {
  if (value === null || value === undefined || Number.isNaN(Number(value))) return '-'
  const v = Number(value)
  switch (meta.kind) {
    case 'percent':
      return `${formatNumber(v, locale, 1)}%`
    case 'score':
      return meta.key === 'nps' && v > 0 ? `+${formatNumber(v, locale, 1)}` : formatNumber(v, locale, 1)
    case 'rupiah':
      return formatRupiahCompact(v * (meta.rupiahPerUnit ?? 1), locale)
  }
}

export type KpiTrend = 'Improving' | 'Deteriorating' | 'Stable'

/** Changes smaller than this share of the current value count as stable. */
const STABLE_THRESHOLD = 0.005

/**
 * Direction-aware trend: a rising turnover rate is a deterioration even though
 * the number goes up.
 */
export const evaluateKpiTrend = (current: number, forecast: number, meta: KpiMeta): KpiTrend => {
  const c = Number(current)
  const f = Number(forecast)
  if (!Number.isFinite(c) || !Number.isFinite(f)) return 'Stable'
  const delta = f - c
  if (Math.abs(delta) <= Math.abs(c) * STABLE_THRESHOLD) return 'Stable'
  const better = meta.higherIsBetter ? delta > 0 : delta < 0
  return better ? 'Improving' : 'Deteriorating'
}

// ─── Periods ───────────────────────────────────────────────────────────────

const MONTHS_ID = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']
const MONTHS_EN = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

/** "2025-12" → "Des 2025"; other labels are returned unchanged. */
export const formatForecastPeriod = (period: string, locale: AppLocale = 'id'): string => {
  const m = /^(\d{4})-(\d{2})$/.exec(String(period || ''))
  if (!m) return String(period || '').replace(/\s*\(Pred\)\s*$/i, '')
  const names = locale === 'en' ? MONTHS_EN : MONTHS_ID
  return `${names[Number(m[2]) - 1]} ${m[1]}`
}

/** "Jan–Mar 2026", or "Nov 2026–Jan 2027" across a year boundary. */
export const formatForecastPeriodRange = (start: string, end: string, locale: AppLocale = 'id'): string => {
  if (!start) return ''
  if (!end || start === end) return formatForecastPeriod(start, locale)
  const a = formatForecastPeriod(start, locale)
  const b = formatForecastPeriod(end, locale)
  const [aMonth, aYear] = a.split(' ')
  const [, bYear] = b.split(' ')
  return aYear && aYear === bYear ? `${aMonth}–${b}` : `${a}–${b}`
}

/** Last day of a "YYYY-MM" period has passed. */
export const isPeriodPast = (period: string, now: Date = new Date()): boolean => {
  const m = /^(\d{4})-(\d{2})$/.exec(String(period || ''))
  if (!m) return false
  const endOfPeriod = new Date(Number(m[1]), Number(m[2]), 1) // first day of the next month
  return endOfPeriod.getTime() <= now.getTime()
}
