// Locale-aware thousands separators for number/currency inputs and displays.
// The model always holds a canonical raw string ("-1234567.5", "." as decimal point, no grouping);
// only the text shown to the user is localised.

export type NumberLocale = 'en' | 'id'

export interface NumberSeparators { group: string, decimal: string }

/** en: 1,234,567.5; id: 1.234.567,5. Unknown locales fall back to en. */
export const separatorsFor = (locale?: string): NumberSeparators =>
  locale === 'id' ? { group: '.', decimal: ',' } : { group: ',', decimal: '.' }

/** Currency units the unit selector can store, with the symbol shown in tables. */
export const CURRENCY_UNITS: Readonly<Record<string, { symbol: string, code: string, spaced: boolean }>> = {
  Rp: { symbol: 'Rp', code: 'IDR', spaced: true },
  USD: { symbol: '$', code: 'USD', spaced: false },
  GBP: { symbol: '£', code: 'GBP', spaced: false }
}

export const isCurrencyUnit = (unit?: string | null): boolean => !!unit && unit in CURRENCY_UNITS

/** i18n key (under strategicPlan.form.units) for a stored unit value; undefined for free-text legacy units. */
export const UNIT_LABEL_KEYS: Readonly<Record<string, string>> = {
  '%': 'percentage', 'Rp': 'rupiah', 'USD': 'usd', 'GBP': 'gbp', 'Amount': 'quantity', 'Score': 'score', 'Hour': 'hour', 'Day': 'day'
}

/** A plain number as the backend stores it: optional minus, digits, optional ".digits". */
const RAW_NUMBER = /^-?\d+(\.\d+)?$/

/**
 * Typed or pasted text -> canonical raw string. Keeps digits, a leading minus and the first
 * locale decimal separator; group separators and anything else are dropped.
 * Intermediate states survive so typing works: "-", "12,", "0,".
 */
export const parseNumberInput = (text: unknown, locale?: string, opts: { allowNegative?: boolean } = {}): string => {
  const { decimal } = separatorsFor(locale)
  const input = String(text ?? '').trim()
  const negative = (opts.allowNegative ?? true) && input.startsWith('-')
  let int = ''
  let frac = ''
  let seenDecimal = false
  for (const ch of input) {
    if (ch >= '0' && ch <= '9') {
      if (seenDecimal) frac += ch
      else int += ch
    } else if (ch === decimal && !seenDecimal) {
      seenDecimal = true
    }
  }
  int = int.replace(/^0+(?=\d)/, '')
  if (seenDecimal && int === '') int = '0'
  const body = seenDecimal ? `${int}.${frac}` : int
  if (body === '') return negative ? '-' : ''
  return negative ? `-${body}` : body
}

/** Canonical raw string -> text with the locale's separators. Non-numeric values are returned as-is. */
export const formatNumberInput = (raw: unknown, locale?: string): string => {
  const value = String(raw ?? '').trim()
  if (value === '') return ''
  if (value === '-') return '-'
  const match = /^(-?)(\d+)(\.(\d*))?$/.exec(value)
  if (!match) return value
  const { group, decimal } = separatorsFor(locale)
  const [, sign, int, dot, frac] = match
  const grouped = int!.replace(/\B(?=(\d{3})+(?!\d))/g, group)
  return `${sign}${grouped}${dot ? decimal + (frac ?? '') : ''}`
}

/** Characters that carry meaning in an input (everything but group separators). */
const significantCount = (text: string, locale?: string): number => {
  const { group } = separatorsFor(locale)
  let n = 0
  for (const ch of text) if (ch !== group) n++
  return n
}

/**
 * Caret position in `formatted` that sits after the same number of significant characters as the
 * caret had in the text the user just edited. Keeps the caret next to the digit being typed.
 */
export const caretAfterFormat = (editedText: string, editedCaret: number, formatted: string, locale?: string): number => {
  const target = significantCount(editedText.slice(0, editedCaret), locale)
  const { group } = separatorsFor(locale)
  let seen = 0
  for (let i = 0; i < formatted.length; i++) {
    if (seen === target) return i
    if (formatted[i] !== group) seen++
  }
  return formatted.length
}

/** Raw number string for display: "1234567.5" -> "1,234,567.5". Anything non-numeric is shown unchanged. */
export const formatDisplayNumber = (value: unknown, locale?: string): string => {
  const raw = String(value ?? '').trim()
  return RAW_NUMBER.test(raw) ? formatNumberInput(raw, locale) : raw
}

/**
 * A KPI target/realization for a table or modal.
 * Currency units get their symbol ("Rp 1.234.567", "$1,234", "£1,234"); other units are appended
 * ("85 %"); "Amount" (shown as Quantity), "Score" and unit-less values are just the grouped number.
 * Empty values give "-".
 */
export const formatUnitValue = (value: unknown, unit?: string | null, locale?: string): string => {
  const raw = String(value ?? '').trim()
  if (raw === '' || raw === '-') return '-'
  const text = formatDisplayNumber(raw, locale)
  if (!RAW_NUMBER.test(raw)) return text
  const currency = unit ? CURRENCY_UNITS[unit] : undefined
  if (currency) {
    if (text.startsWith('-')) return `-${currency.symbol}${currency.spaced ? ' ' : ''}${text.slice(1)}`
    return `${currency.symbol}${currency.spaced ? ' ' : ''}${text}`
  }
  if (!unit || unit === 'Amount' || unit.toLowerCase() === 'score') return text
  return `${text} ${unit}`
}
