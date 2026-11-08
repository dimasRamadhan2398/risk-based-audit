/**
 * Trusted device helpers for Settings > Activity (components/settings/activity.vue).
 *
 * GET /api/v1/devices (auth-service) answers `{ success, message, data: { data: [...] } }` with snake_case rows.
 * The list holds trusted devices and session-only devices (signed in with MFA without "trust this device");
 * `is_trusted` tells them apart. Older builds prepended one zero-value row (id 00000000-...) per device and
 * sent no `is_trusted` (everything listed was trusted); both are handled here.
 */

export const ZERO_UUID = '00000000-0000-0000-0000-000000000000'

export interface TrustedDeviceView {
  id: string
  fingerprint: string
  /** As stored at trust time, e.g. "Chrome on macOS". */
  name: string
  type: string
  userAgent: string
  /** First address of an X-Forwarded-For list. */
  ipAddress: string
  /** Not sent by the backend today; shown when present. */
  location: string
  createdAt: string
  expiresAt: string
  /** Last sign-in or MFA check from this device ('' when unknown) */
  lastSeenAt: string
  /** Skips the MFA step at sign-in */
  trusted: boolean
}

type Raw = Record<string, unknown>

const str = (v: unknown): string => (typeof v === 'string' ? v.trim() : '')

const pick = (row: Raw, ...keys: string[]): string => {
  for (const k of keys) {
    const v = str(row[k])
    if (v) return v
  }
  return ''
}

/** The device array from any of the envelopes the API has used. */
export function unwrapDeviceList(payload: unknown): unknown[] {
  let cur: unknown = payload
  for (let i = 0; i < 3; i++) {
    if (Array.isArray(cur)) return cur
    if (cur && typeof cur === 'object' && 'data' in (cur as Raw)) cur = (cur as Raw).data
    else break
  }
  return Array.isArray(cur) ? cur : []
}

export function normalizeTrustedDevice(raw: unknown): TrustedDeviceView | null {
  if (!raw || typeof raw !== 'object') return null
  const row = raw as Raw
  const id = pick(row, 'id', 'ID')
  if (!id || id === ZERO_UUID) return null
  const ip = pick(row, 'ip_address', 'ipAddress')
  return {
    id,
    fingerprint: pick(row, 'device_fingerprint', 'deviceFingerprint'),
    name: pick(row, 'device_name', 'deviceName'),
    type: pick(row, 'device_type', 'deviceType').toLowerCase(),
    userAgent: pick(row, 'user_agent', 'userAgent'),
    ipAddress: ip.split(',')[0]?.trim() ?? '',
    location: pick(row, 'location', 'city'),
    createdAt: pick(row, 'created_at', 'createdAt'),
    expiresAt: pick(row, 'expires_at', 'expiresAt'),
    lastSeenAt: pick(row, 'last_seen_at', 'lastSeenAt'),
    trusted: isTrustedRow(row)
  }
}

const isTrustedRow = (row: Raw): boolean => {
  const flag = row.is_trusted ?? row.isTrusted
  if (typeof flag === 'boolean') return flag
  const sessionOnly = row.session_only ?? row.sessionOnly
  return sessionOnly !== true
}

const time = (iso: string) => Date.parse(iso) || 0

/** Valid devices, most recently active first (last seen, else added), one row per id. */
export function normalizeTrustedDevices(payload: unknown): TrustedDeviceView[] {
  const seen = new Set<string>()
  const out: TrustedDeviceView[] = []
  for (const raw of unwrapDeviceList(payload)) {
    const d = normalizeTrustedDevice(raw)
    if (!d || seen.has(d.id)) continue
    seen.add(d.id)
    out.push(d)
  }
  const activeAt = (d: TrustedDeviceView) => time(d.lastSeenAt) || time(d.createdAt)
  return out.sort((a, b) => activeAt(b) - activeAt(a))
}

export interface DeviceDescription {
  browser: string
  os: string
  type: 'mobile' | 'tablet' | 'desktop'
}

const detectBrowser = (ua: string): string => {
  if (/Edg(e|A|iOS)?\//i.test(ua)) return 'Edge'
  if (/OPR\/|Opera/i.test(ua)) return 'Opera'
  if (/Firefox|FxiOS/i.test(ua)) return 'Firefox'
  if (/Chrome|CriOS|Chromium/i.test(ua)) return 'Chrome'
  if (/Safari/i.test(ua)) return 'Safari'
  return ''
}

const detectOs = (ua: string): string => {
  if (/iPhone|iPad|iPod/i.test(ua)) return 'iOS'
  if (/Android/i.test(ua)) return 'Android'
  if (/Windows/i.test(ua)) return 'Windows'
  if (/Mac OS X|Macintosh/i.test(ua)) return 'macOS'
  if (/CrOS/i.test(ua)) return 'ChromeOS'
  if (/Linux/i.test(ua)) return 'Linux'
  return ''
}

/**
 * Browser, OS and form factor from the user agent, falling back to the stored "<Browser> on <OS>" name
 * (what useDeviceFingerprint sends) and the stored device type. Unknown parts are ''.
 */
export function describeDevice(input: { userAgent?: string, name?: string, type?: string }): DeviceDescription {
  const ua = input.userAgent || ''
  const [nameBrowser = '', nameOs = ''] = (input.name || '').split(/\s+on\s+/i).map(s => s.trim())
  const browser = detectBrowser(ua) || (nameBrowser && !/^unknown/i.test(nameBrowser) ? nameBrowser : '')
  const os = detectOs(ua) || (nameOs && !/^unknown/i.test(nameOs) ? nameOs : '')

  const storedType = (input.type || '').toLowerCase()
  let type: DeviceDescription['type'] = 'desktop'
  if (/iPad|Tablet/i.test(ua) || storedType === 'tablet') type = 'tablet'
  else if (/Mobi|iPhone|Android/i.test(ua) || storedType === 'mobile') type = 'mobile'

  return { browser, os, type }
}

export const deviceIcon = (type: string): string =>
  type === 'mobile' ? 'i-lucide-smartphone' : type === 'tablet' ? 'i-lucide-tablet' : 'i-lucide-laptop'
