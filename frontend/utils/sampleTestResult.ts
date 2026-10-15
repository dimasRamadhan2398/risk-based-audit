import type { TestResult } from '~/types/audit'

/**
 * KKA F-03 sample test results (l1/l2/l3).
 *
 * The backend stores them as booleans (true = Pass, false = Fail, null = not tested).
 * Older rows and the edit form use the strings 'Pass' / 'Fail' / 'N/A'.
 * These helpers accept either shape.
 */

/** Any stored or form value as the form's TestResult ('Pass' | 'Fail' | 'N/A' | undefined). */
export const toTestResult = (value: unknown): TestResult => {
  if (value === true) return 'Pass'
  if (value === false) return 'Fail'
  if (typeof value === 'string') {
    const v = value.trim().toLowerCase()
    if (v === 'pass') return 'Pass'
    if (v === 'fail') return 'Fail'
    if (v === 'n/a') return 'N/A'
  }
  return undefined
}

/** Any stored or form value as the API's boolean (null when not tested or N/A). */
export const toTestResultBoolean = (value: unknown): boolean | null => {
  const result = toTestResult(value)
  if (result === 'Pass') return true
  if (result === 'Fail') return false
  return null
}

/** Display text: 'Pass', 'Fail', 'N/A' or '-' when there is no result. */
export const formatTestResult = (value: unknown): string => toTestResult(value) ?? '-'
