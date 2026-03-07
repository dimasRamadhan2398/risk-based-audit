import { format, parseISO, isValid } from 'date-fns'

/**
 * Convert date string to formatted string
 */
export const formatDate = (date: string | Date, formatStr: string = 'yyyy-MM-dd'): string => {
  try {
    const dateObj = typeof date === 'string' ? parseISO(date) : date
    return isValid(dateObj) ? format(dateObj, formatStr) : ''
  } catch {
    return ''
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
 * Convert date to display format
 */
export const toDisplayDate = (date: string | Date): string => {
  return formatDate(date, 'dd MMM yyyy')
}

/**
 * Convert date to ISO string
 */
export const toISODate = (date: Date): string => {
  return date.toISOString()
}

/**
 * Parse ISO date string
 */
export const parseDate = (dateString: string): Date | null => {
  try {
    const date = parseISO(dateString)
    return isValid(date) ? date : null
  } catch {
    return null
  }
}
