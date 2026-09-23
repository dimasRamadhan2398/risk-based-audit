import { format, parseISO, isValid } from 'date-fns'

/**
 * Convert date string to formatted string (defaults to 'dd/MM/yyyy')
 */
export const formatDate = (
  date: string | Date | null | undefined,
  formatStr: string = 'dd/MM/yyyy'
): string => {
  if (!date) return ''
  try {
    if (typeof date === 'string' && /^\d{2}\/\d{2}\/\d{4}$/.test(date.trim()) && formatStr === 'dd/MM/yyyy') {
      return date.trim()
    }
    let dateObj = typeof date === 'string' ? parseISO(date) : date
    if (!isValid(dateObj) && typeof date === 'string') {
      dateObj = new Date(date)
    }
    return isValid(dateObj) ? format(dateObj, formatStr) : String(date)
  } catch {
    return typeof date === 'string' ? date : ''
  }
}

export const formatTime = (date: Date): string => {
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

/**
 * Convert date to display format (dd/MM/yyyy)
 */
export const toDisplayDate = (
  date: string | Date | null | undefined,
  formatStr: string = 'dd/MM/yyyy'
): string => {
  return formatDate(date, formatStr)
}

/**
 * Convert date to ISO string
 */
export const toISODate = (date: Date): string => {
  return date.toISOString()
}

/**
 * Convert date to localized full date string (e.g. "10 Februari 2026" or "10 February 2026")
 */
export const formatDateLocale = (
  date: string | Date | null | undefined,
  locale: string = 'id'
): string => {
  if (!date) return ''
  try {
    const dateObj = typeof date === 'string' ? parseISO(date) : date
    if (!dateObj || !isValid(dateObj)) return ''
    const intlLocale = locale === 'id' ? 'id-ID' : 'en-GB'
    return new Intl.DateTimeFormat(intlLocale, {
      day: 'numeric',
      month: 'long',
      year: 'numeric'
    }).format(dateObj)
  } catch {
    return ''
  }
}

/**
 * Format a start and end date range into a localized period string.
 * Example (id): "10 Februari 2026 - 20 Maret 2026"
 * Example (en): "10 February 2026 - 20 March 2026"
 */
export const formatPeriod = (
  startDate?: string | Date | null,
  endDate?: string | Date | null,
  locale: string = 'id',
  separator: string = ' - '
): string => {
  const startStr = formatDateLocale(startDate, locale)
  const endStr = formatDateLocale(endDate, locale)

  if (startStr && endStr) {
    return `${startStr}${separator}${endStr}`
  }
  return startStr || endStr || '-'
}
