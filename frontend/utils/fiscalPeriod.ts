export type Quarter = 'Q1' | 'Q2' | 'Q3' | 'Q4'

/** Fiscal year = calendar year, from LOCAL date parts (never UTC ISO slicing). */
export function currentFiscalYear(now: Date = new Date()): number {
  return now.getFullYear()
}

/** Calendar quarter in local time: Jan-Mar Q1, Apr-Jun Q2, Jul-Sep Q3, Oct-Dec Q4. */
export function currentQuarter(now: Date = new Date()): Quarter {
  return `Q${Math.floor(now.getMonth() / 3) + 1}` as Quarter
}
