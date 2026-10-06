import { ref, computed } from 'vue'

/**
 * Returns a dynamic 5-year fiscal year range starting from last year
 * up to 3 years ahead from the current year.
 * Example for 2026: [2025, 2026, 2027, 2028, 2029]
 * Example for 2027: [2026, 2027, 2028, 2029, 2030]
 */
export function getFiscalYears(baseYear: number = new Date().getFullYear()): number[] {
  return [
    baseYear - 1,
    baseYear,
    baseYear + 1,
    baseYear + 2,
    baseYear + 3
  ]
}

/**
 * Returns the dynamic 5-year fiscal year range as an array of strings.
 * Example for 2026: ['2025', '2026', '2027', '2028', '2029']
 */
export function getFiscalYearStrings(baseYear: number = new Date().getFullYear()): string[] {
  return getFiscalYears(baseYear).map(String)
}

/**
 * Returns the dynamic 5-year fiscal year range formatted for select components.
 * Example: [{ label: '2025', value: 2025 }, ...] or with prefix like [{ label: 'Tahun 2025', value: 2025 }, ...]
 */
export function getFiscalYearSelectOptions(baseYear: number = new Date().getFullYear(), prefix: string = '') {
  return getFiscalYears(baseYear).map(y => ({
    label: prefix ? `${prefix} ${y}` : String(y),
    value: y
  }))
}

// Module-level shared ref for synchronizing selected fiscal year across the Risk Profile module suite
const sharedFiscalYear = ref<number>(new Date().getFullYear())

export function useFiscalYear() {
  const currentYear = computed(() => new Date().getFullYear())
  const fiscalYears = computed(() => getFiscalYears(currentYear.value))
  const fiscalYearStrings = computed(() => getFiscalYearStrings(currentYear.value))

  return {
    currentYear,
    selectedFiscalYear: sharedFiscalYear,
    fiscalYears,
    fiscalYearStrings,
    getFiscalYears,
    getFiscalYearStrings,
    getFiscalYearSelectOptions
  }
}
