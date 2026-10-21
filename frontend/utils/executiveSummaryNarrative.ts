/**
 * Executive summary narrative helpers shared by the executive summary store, the Individual and
 * Compilation forms and the unit tests.
 *
 * The generated narrative always starts with the period it covers:
 *   Individual:  "Periode Januari 2026, SPI telah menerbitkan ..." (audit period of the linked LHA)
 *   Compilation: "Periode Kuartal 4 2026, SPI telah menerbitkan ..." (selected quarter and year)
 * Texts come from i18n (executiveSummaryForm.narrative.*); month names from common.datePicker.months.
 * When no period is known the prefix is left out and the narrative starts with the template body.
 */

export type Translate = (key: string, params?: Record<string, string | number>) => string

export const NARRATIVE_KEYS = {
  periodPrefix: 'executiveSummaryForm.narrative.periodPrefix',
  templateBody: 'executiveSummaryForm.narrative.templateBody',
  quarterPeriod: 'executiveSummaryForm.narrative.quarterPeriod',
  quarterPeriodNoYear: 'executiveSummaryForm.narrative.quarterPeriodNoYear'
} as const

const MONTH_KEY = 'common.datePicker.months'
const RANGE_SEPARATOR = ' – '

export interface MonthYear {
  year: number
  /** 1-12 */
  month: number
}

export interface PeriodRange {
  start: MonthYear | null
  end: MonthYear | null
}

const toMonthYear = (year: number, month: number): MonthYear | null =>
  Number.isInteger(year) && year >= 1000 && Number.isInteger(month) && month >= 1 && month <= 12
    ? { year, month }
    : null

/** Reads the month and year of 'YYYY-MM', 'YYYY-MM-DD', an ISO datetime or 'DD/MM/YYYY' / 'DD-MM-YYYY'. */
export const parseMonthYear = (value?: string | null): MonthYear | null => {
  const text = String(value ?? '').trim()
  if (!text) return null
  const iso = text.match(/^(\d{4})-(\d{1,2})(?:-\d{1,2})?(?:[T\s].*)?$/)
  if (iso) return toMonthYear(Number(iso[1]), Number(iso[2]))
  const dmy = text.match(/^\d{1,2}[/-](\d{1,2})[/-](\d{4})$/)
  if (dmy) return toMonthYear(Number(dmy[2]), Number(dmy[1]))
  return null
}

/** Splits an assignment letter executionPeriod ("2026-01-05 to 2026-03-20", "... s/d ...") into its two dates. */
export const parseExecutionPeriod = (value?: string | null): PeriodRange => {
  const parts = String(value ?? '').split(/\s+(?:to|s\/d|sd|until|hingga|sampai|–|—|-)\s+/i)
  return { start: parseMonthYear(parts[0]), end: parseMonthYear(parts[1]) }
}

/** Audit period of an assignment letter: startPeriod/finishPeriod, else the executionPeriod text. */
export const auditPeriodFromAssignmentLetter = (letter?: {
  startPeriod?: string | null
  finishPeriod?: string | null
  executionPeriod?: string | null
} | null): PeriodRange => {
  if (!letter) return { start: null, end: null }
  const fromExecution = parseExecutionPeriod(letter.executionPeriod)
  return {
    start: parseMonthYear(letter.startPeriod) ?? fromExecution.start,
    end: parseMonthYear(letter.finishPeriod) ?? fromExecution.end
  }
}

const monthIndex = (value: MonthYear) => value.year * 12 + (value.month - 1)

/** Localized month name (id: Januari, en: January). */
export const monthName = (t: Translate, month: number): string => t(`${MONTH_KEY}.${month - 1}`)

/**
 * Formats an audit period as month(s) and year:
 *   same month           -> "Januari 2026"
 *   same year            -> "Januari – Maret 2026"
 *   different years      -> "Desember 2025 – Januari 2026"
 * Returns '' when neither date is known. A single known date is shown as one month.
 */
export const formatAuditPeriod = (t: Translate, range: PeriodRange): string => {
  let start = range.start ?? range.end
  let end = range.end ?? range.start
  if (!start || !end) return ''
  if (monthIndex(end) < monthIndex(start)) [start, end] = [end, start]
  const endLabel = `${monthName(t, end.month)} ${end.year}`
  if (monthIndex(start) === monthIndex(end)) return endLabel
  const startLabel = start.year === end.year
    ? monthName(t, start.month)
    : `${monthName(t, start.month)} ${start.year}`
  return `${startLabel}${RANGE_SEPARATOR}${endLabel}`
}

/** Compilation period: "Kuartal 4 2026" / "Q4 2026". */
export const formatQuarterPeriod = (t: Translate, quarter: number, year?: number | null): string => {
  const q = Number(quarter)
  if (!Number.isInteger(q) || q < 1 || q > 4) return ''
  const y = Number(year)
  return y > 0
    ? t(NARRATIVE_KEYS.quarterPeriod, { quarter: q, year: y })
    : t(NARRATIVE_KEYS.quarterPeriodNoYear, { quarter: q })
}

/** "Periode Januari 2026," / "For the period of January 2026,", or '' without a period. */
export const narrativePrefix = (t: Translate, period: string): string =>
  period ? t(NARRATIVE_KEYS.periodPrefix, { period }) : ''

/** Default narrative: the period prefix followed by the template body. */
export const buildDefaultNarrative = (t: Translate, period: string): string =>
  [narrativePrefix(t, period), t(NARRATIVE_KEYS.templateBody)].filter(Boolean).join(' ')

const escapeRegExp = (value: string) => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

// A generated prefix of the current locale. The period part must contain a digit (year or quarter),
// so ordinary text such as "Periode ini, ..." is not taken for a generated prefix.
const prefixPattern = (t: Translate): RegExp | null => {
  const template = t(NARRATIVE_KEYS.periodPrefix)
  const [before, after] = template.split('{period}')
  if (before === undefined || after === undefined) return null
  return new RegExp(`^\\s*${escapeRegExp(before)}[^,\\n]*\\d[^,\\n]*?${escapeRegExp(after)}\\s*`)
}

export const hasPeriodPrefix = (t: Translate, text: string): boolean =>
  prefixPattern(t)?.test(text ?? '') ?? false

export const stripPeriodPrefix = (t: Translate, text: string): string => {
  const pattern = prefixPattern(t)
  return pattern ? (text ?? '').replace(pattern, '') : (text ?? '')
}

// "Berdasarkan hasil ..." -> "berdasarkan hasil ..." after the prefix; acronyms ("SPI", "LHA") stay as they are.
const lowerFirstWord = (text: string) => text.replace(/^(\p{Lu})(?=\p{Ll})/u, c => c.toLowerCase())

/**
 * Puts the period prefix in front of a narrative, replacing a generated prefix that is already there.
 * Without a period, a generated prefix is removed and the rest is returned as it is.
 */
export const prefixNarrative = (t: Translate, text: string, period: string): string => {
  const body = stripPeriodPrefix(t, text ?? '').trimStart()
  if (!period) return body
  return body ? `${narrativePrefix(t, period)} ${lowerFirstWord(body)}` : narrativePrefix(t, period)
}

/**
 * Updates the period of a narrative after the period changed (quarter/year selection, linked LHA).
 * Only text that still starts with a generated prefix or with the template body is changed;
 * free text written by the user and an empty narrative are returned unchanged.
 */
export const refreshNarrativePeriod = (t: Translate, text: string, period: string): string => {
  const current = text ?? ''
  if (!current.trim()) return current
  const lead = (t(NARRATIVE_KEYS.templateBody).split('....')[0] ?? '').trim()
  const startsWithTemplate = !!lead && current.trimStart().startsWith(lead)
  if (!hasPeriodPrefix(t, current) && !startsWithTemplate) return current
  return prefixNarrative(t, current, period)
}
