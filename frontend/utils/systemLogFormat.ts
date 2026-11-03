/**
 * Display helpers for the System Logs tab (components/settings/system-logs.vue).
 */

/**
 * Action codes with a translated label under settings.systemLogs.actions.<key>. Codes not listed here (or not
 * yet translated) show the raw code. Mirrors the Action* constants in backend/auth-service/pkg/activitylog/entry.go.
 */
export const KNOWN_SYSTEM_LOG_ACTIONS = [
  'user.login',
  'user.logout',
  'user.login_failed',
  'user.password_changed',
  'user.password_reset_by_admin',
  'mfa.enabled',
  'mfa.disabled',
  'mfa.verified',
  'mfa.failed',
  'role.created',
  'role.deleted',
  'role.permissions_assigned',
  'role.permissions_revoked',
  'role.assigned',
  'role.revoked',
  'device.trusted_added',
  'device.trusted_removed',
  'permission.denied',
  'security.step_up',
  'system_logs.viewed'
] as const

/** i18n-safe key for an action code: `auth.login-failed` -> `auth_login_failed` (dots would split the key path). */
export const actionKey = (action: string): string =>
  String(action || '').trim().toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_+|_+$/g, '')

/** Intl locale for the app locale cookie value. */
export const intlLocale = (locale: string): string => (locale === 'id' ? 'id-ID' : 'en-GB')

const UNITS: [Intl.RelativeTimeFormatUnit, number][] = [
  ['year', 31_536_000],
  ['month', 2_592_000],
  ['week', 604_800],
  ['day', 86_400],
  ['hour', 3_600],
  ['minute', 60],
  ['second', 1]
]

/** "3 minutes ago" / "3 menit yang lalu". Under 10 seconds reads as "now". */
export function formatRelativeTime(iso: string, now: number, locale: string): string {
  const ts = Date.parse(iso)
  if (Number.isNaN(ts)) return iso || ''
  const diffSec = Math.round((ts - now) / 1000)
  const rtf = new Intl.RelativeTimeFormat(intlLocale(locale), { numeric: 'auto' })
  if (Math.abs(diffSec) < 10) return rtf.format(0, 'second')
  for (const [unit, sec] of UNITS) {
    if (Math.abs(diffSec) >= sec) return rtf.format(Math.trunc(diffSec / sec), unit)
  }
  return rtf.format(diffSec, 'second')
}

/** Full local date and time, e.g. "4 Nov 2026, 14:03:22". */
export function formatAbsoluteTime(iso: string | Date, locale: string): string {
  const d = iso instanceof Date ? iso : new Date(iso)
  if (Number.isNaN(d.getTime())) return String(iso || '')
  return new Intl.DateTimeFormat(intlLocale(locale), { dateStyle: 'medium', timeStyle: 'medium' }).format(d)
}

/**
 * RFC3339 (UTC, whole seconds) for a `YYYY-MM-DD` date input, at the start or end of that local day.
 * Returns undefined for an empty or invalid value.
 */
export function dateInputToRfc3339(value: string | undefined | null, edge: 'start' | 'end'): string | undefined {
  if (!value || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return undefined
  const d = new Date(`${value}T${edge === 'start' ? '00:00:00' : '23:59:59'}`)
  if (Number.isNaN(d.getTime())) return undefined
  return d.toISOString().replace(/\.\d{3}Z$/, 'Z')
}

/** Details as sent, with a JSON-encoded string decoded. */
export function parseDetails(details: unknown): unknown {
  if (typeof details !== 'string') return details
  const s = details.trim()
  if (!s.startsWith('{') && !s.startsWith('[')) return details
  try {
    return JSON.parse(s)
  } catch {
    return details
  }
}

/** True when there is nothing to show for details. */
export function isEmptyDetails(details: unknown): boolean {
  const d = parseDetails(details)
  if (d === null || d === undefined || d === '') return true
  if (Array.isArray(d)) return d.length === 0
  return typeof d === 'object' && Object.keys(d as object).length === 0
}

/** Flat list of [key, value] for a details object; null when details should be shown as raw JSON instead. */
export function detailEntries(raw: unknown): [string, string][] | null {
  const details = parseDetails(raw)
  if (!details || typeof details !== 'object' || Array.isArray(details)) return null
  return Object.entries(details as Record<string, unknown>).map(([k, v]) => [
    k,
    v === null || v === undefined ? '-' : typeof v === 'object' ? JSON.stringify(v) : String(v)
  ])
}
