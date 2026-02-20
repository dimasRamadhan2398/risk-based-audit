/**
 * Convert string to number safely
 */
export const toNumber = (value: string | number | null | undefined, defaultValue: number = 0): number => {
  if (typeof value === 'number') return value
  if (!value) return defaultValue

  const num = parseFloat(value)
  return isNaN(num) ? defaultValue : num
}

/**
 * Convert string to integer safely
 */
export const toInt = (value: string | number | null | undefined, defaultValue: number = 0): number => {
  if (typeof value === 'number') return Math.floor(value)
  if (!value) return defaultValue

  const num = parseInt(value, 10)
  return isNaN(num) ? defaultValue : num
}

/**
 * Convert value to string safely
 */
export const toString = (value: any): string => {
  if (value === null || value === undefined) return ''
  return String(value)
}

/**
 * Convert string to boolean
 */
export const toBoolean = (value: string | boolean | number | null | undefined): boolean => {
  if (typeof value === 'boolean') return value
  if (typeof value === 'number') return value !== 0
  if (!value) return false

  const str = String(value).toLowerCase().trim()
  return str === 'true' || str === '1' || str === 'yes' || str === 'on'
}

/**
 * Convert array to comma-separated string
 */
export const arrayToString = (arr: any[], separator: string = ', '): string => {
  return arr.filter(Boolean).join(separator)
}

/**
 * Convert comma-separated string to array
 */
export const stringToArray = (str: string, separator: string = ','): string[] => {
  return str.split(separator).map(s => s.trim()).filter(Boolean)
}

export const doubleToPercentage = (value: number): string => {
  return `${parseFloat((value * 100).toFixed(2))}%`
}
