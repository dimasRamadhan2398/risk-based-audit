/**
 * User seats and employee account lifecycle helpers (Employee Management).
 *
 * A seat is any active login holding a role other than VIEWER; viewers are
 * unlimited. GET /users/seats (admin only) reports the usage. Kept free of
 * Vue/Nuxt so the rules can be unit-tested directly.
 */
import type { CreatedEmployee, UserSeats } from '~/types/master'
import { getErrorCode, getUserErrorMessage, type TranslateFn } from '~/utils/error'

/** Seats left before the indicator turns to the warning colour */
export const SEAT_WARNING_MARGIN = 5

export type SeatStatus = 'neutral' | 'warning' | 'error'

/**
 * Indicator colour: neutral below limit - 5 (45 of 50), warning from there up
 * to limit - 1 (49), error at or above the limit or when the backend reports
 * the tenant as over the limit.
 */
export const getSeatStatus = (seats: Pick<UserSeats, 'limit' | 'used'> & { over_limit?: boolean }): SeatStatus => {
  if (seats.over_limit || seats.used >= seats.limit) return 'error'
  if (seats.used >= seats.limit - SEAT_WARNING_MARGIN) return 'warning'
  return 'neutral'
}

/**
 * The create form warns that the role will be set to Viewer when no seat is
 * left, including tenants already above the limit. Never shown when editing
 * or when the seats are unknown (not loaded, or the user is not an admin).
 */
export const shouldShowSeatsFullAlert = (seats: UserSeats | null | undefined, isEditing: boolean): boolean => {
  if (isEditing || !seats) return false
  return seats.over_limit || seats.remaining <= 0 || seats.used >= seats.limit
}

/**
 * Whether the create response says the login was created as VIEWER because
 * the seats were full. Read from data.account (where the login account result
 * already lives), with a top-level fallback in case the backend flattens it.
 */
export const isSeatDowngraded = (created: CreatedEmployee | null | undefined): boolean => {
  if (!created) return false
  return created.account?.seat_downgraded === true || created.seat_downgraded === true
}

/** Employee mutations whose 409 responses get a specific message */
export type EmployeeAction = 'deactivate' | 'reactivate' | 'update' | 'delete' | 'hardDelete'

export interface EmployeeActionError {
  /** One-line message for a toast title or the form alert */
  message: string
  /** For EMPLOYEE_IN_USE: where the employee is still referenced */
  usages: string[]
}

const asRecord = (value: unknown): Record<string, unknown> | undefined =>
  value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, unknown> : undefined

/** error.data.error.details, also accepting a flat { details } body */
const getErrorDetails = (error: unknown): unknown => {
  const data = asRecord(asRecord(error)?.data)
  const nested = asRecord(data?.error)
  return nested?.details ?? data?.details
}

/** "audit_plans" -> "Audit plans"; human labels are kept as they are */
const humanize = (value: string): string => {
  const text = value.trim()
  if (!/^[a-z0-9_]+$/.test(text)) return text
  const words = text.replace(/_/g, ' ')
  return words.charAt(0).toUpperCase() + words.slice(1)
}

const usageLabel = (entry: unknown, key?: string): string | undefined => {
  if (typeof entry === 'string') return entry.trim() ? humanize(entry) : undefined
  if (typeof entry === 'number') return key ? `${humanize(key)} (${entry})` : undefined

  const item = asRecord(entry)
  if (!item) return key ? humanize(key) : undefined
  const nameField = ['label', 'name', 'module', 'entity', 'resource', 'table', 'type']
    .map((f) => item[f])
    .find((v): v is string => typeof v === 'string' && !!v.trim())
  const name = nameField ? humanize(nameField) : (key ? humanize(key) : undefined)
  if (!name) return undefined
  const count = item.count ?? item.total
  return typeof count === 'number' ? `${name} (${count})` : name
}

/**
 * Where an employee is still referenced, from the EMPLOYEE_IN_USE details.
 * The exact shape is not fixed yet, so this accepts a list of strings, a list
 * of { label|name|module|entity|table, count } objects, a { name: count } map,
 * or any of those wrapped in { references | usages | used_in | items }.
 */
export const getEmployeeUsages = (error: unknown): string[] => {
  let details = getErrorDetails(error)
  if (typeof details === 'string') return details.trim() ? [details.trim()] : []

  const wrapper = asRecord(details)
  const wrapped = wrapper && ['references', 'usages', 'used_in', 'items'].map((k) => wrapper[k]).find((v) => v != null)
  if (wrapped != null) details = wrapped

  if (Array.isArray(details)) {
    return details.map((entry) => usageLabel(entry)).filter((v): v is string => !!v)
  }
  const map = asRecord(details)
  if (map) {
    return Object.entries(map)
      .filter(([, v]) => !(typeof v === 'number' && v <= 0))
      .map(([k, v]) => usageLabel(v, k))
      .filter((v): v is string => !!v)
  }
  return []
}

const SEAT_LIMIT_FALLBACK = 50

const seatLimitFromError = (error: unknown): number => {
  const limit = asRecord(getErrorDetails(error))?.limit
  return typeof limit === 'number' && limit > 0 ? limit : SEAT_LIMIT_FALLBACK
}

const LAST_ADMIN_KEYS: Record<EmployeeAction, string> = {
  deactivate: 'masterData.employee.errors.lastAdminDeactivate',
  reactivate: 'masterData.employee.errors.lastAdminDeactivate',
  update: 'masterData.employee.errors.lastAdminUpdate',
  delete: 'masterData.employee.errors.lastAdminDelete',
  hardDelete: 'masterData.employee.errors.lastAdminDelete'
}

const FALLBACK_KEYS: Record<EmployeeAction, string> = {
  deactivate: 'masterData.employee.errors.statusChangeFailed',
  reactivate: 'masterData.employee.errors.statusChangeFailed',
  update: 'masterData.errors.updateEmployee',
  delete: 'masterData.errors.deleteEmployee',
  hardDelete: 'masterData.errors.deleteEmployee'
}

/**
 * Message for a failed employee mutation. The 409 codes of the seat / last
 * admin / in-use rules get action-specific text; anything else goes through
 * the generic getUserErrorMessage.
 */
export const getEmployeeActionError = (error: unknown, t: TranslateFn, action: EmployeeAction): EmployeeActionError => {
  const code = getErrorCode(error)

  if (code === 'SEAT_LIMIT_REACHED') {
    const key = action === 'update' ? 'masterData.employee.errors.seatLimitUpdate' : 'masterData.employee.errors.seatLimitReactivate'
    return { message: t(key, { limit: seatLimitFromError(error) }), usages: [] }
  }
  if (code === 'LAST_ADMIN') {
    return { message: t(LAST_ADMIN_KEYS[action]), usages: [] }
  }
  if (code === 'EMPLOYEE_IN_USE') {
    return { message: t('masterData.employee.errors.inUse'), usages: getEmployeeUsages(error) }
  }
  return { message: getUserErrorMessage(error, t, { fallbackKey: FALLBACK_KEYS[action] }), usages: [] }
}
