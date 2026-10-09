// Pure display-label helpers for the KPI performance page.
// Raw values (status, category, month labels, backend strings) stay untouched as data;
// these only produce the translated text shown to the user, falling back to the raw value.

export type TranslateFn = (key: string, params?: Record<string, string | number>) => string

// Our t() returns the key itself when it is missing.
const translateOr = (t: TranslateFn, path: string, fallback: string, params?: Record<string, string | number>) => {
  const label = t(path, params)
  return label === path ? fallback : label
}

const MONTH_KEYS = ['jan', 'feb', 'mar', 'apr', 'may', 'jun', 'jul', 'aug', 'sep', 'oct', 'nov', 'dec']

/** Translate a 3-letter English month abbreviation (as sent by the API); keep anything else as-is. */
export const kpiMonthLabel = (t: TranslateFn, label: string): string => {
  const raw = String(label)
  const key = raw.toLowerCase()
  return raw.length === 3 && MONTH_KEYS.includes(key) ? translateOr(t, `kpiPerformance.charts.months.${key}`, raw) : label
}

/** Label for a table category/status value; the value itself is what filters and colour mapping compare. */
export const kpiValueLabel = (t: TranslateFn, group: 'categories' | 'statuses', value?: string): string =>
  value ? translateOr(t, `kpiPerformance.table.${group}.${value}`, value) : ''

// The dashboard-summary API (audit-service performance_stats_controller.go) returns English display
// strings for sub-metrics. Map the known formats to translated text; unknown strings are shown raw.
const SUB_TITLE_KEYS: Record<string, string> = {
  'Operational Completion Rate (Started Audits)': 'operationalCompletion',
  'Avg Drafting Cycle-Time': 'avgDraftingCycle',
  'Survey Response Rate': 'surveyResponseRate',
  'Open / Pending Action Plans': 'openPendingActionPlans'
}
const SUB_TEXT_PATTERNS: Array<{ re: RegExp, key: string, params: string[] }> = [
  { re: /^(-?\d+(?:\.\d+)?) days$/, key: 'days', params: ['count'] },
  { re: /^(-?\d+(?:\.\d+)?) days vs target$/, key: 'daysVsTarget', params: ['count'] },
  { re: /^< (\d+(?:\.\d+)?) days$/, key: 'lessThanDays', params: ['count'] },
  { re: /^(\d+) open$/, key: 'openCount', params: ['count'] },
  { re: /^(\d+) overdue$/, key: 'overdueCount', params: ['count'] },
  { re: /^(\d+) completed \/ (\d+) started$/, key: 'completedOfStarted', params: ['completed', 'started'] },
  { re: /^(\d+) responses from (\d+) completed audits$/, key: 'responsesFrom', params: ['responses', 'audits'] },
  { re: /^(\d+) total recommendations registered$/, key: 'totalRecommendations', params: ['count'] }
]

export const kpiSubText = (t: TranslateFn, text: string): string => {
  for (const { re, key, params } of SUB_TEXT_PATTERNS) {
    const m = re.exec(text)
    if (m) {
      const values: Record<string, string> = {}
      params.forEach((name, i) => {
        values[name] = m[i + 1] ?? ''
      })
      return translateOr(t, `kpiPerformance.summary.text.${key}`, text, values)
    }
  }
  return text
}

export interface KpiSubMetric { title: string, value: string, target: string, trend: string }

export const kpiSubMetric = (t: TranslateFn, sub: KpiSubMetric): KpiSubMetric => {
  const titleKey = SUB_TITLE_KEYS[sub.title]
  return {
    title: titleKey ? translateOr(t, `kpiPerformance.summary.subTitles.${titleKey}`, sub.title) : sub.title,
    value: kpiSubText(t, sub.value),
    target: kpiSubText(t, sub.target),
    trend: kpiSubText(t, sub.trend)
  }
}
