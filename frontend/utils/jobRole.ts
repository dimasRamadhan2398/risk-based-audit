// Job role helpers: position types, access-role options and the position -> access role suggestion.
import type { JobPositionType } from '~/types/master'

export const JOB_POSITION_TYPES: JobPositionType[] = [
  'STAFF',
  'SUPERVISOR',
  'MANAGER',
  'SENIOR_MANAGER',
  'GENERAL_MANAGER',
  'DIRECTOR',
  'COMMISSIONER',
  'AUDIT_COMMITTEE',
  'CHIEF_AUDIT_EXECUTIVE',
  'EXTERNAL_AUDITOR'
]

/** Used when GET /job-roles/access-roles is unavailable. */
export const FALLBACK_ACCESS_ROLES: string[] = [
  'EXECUTIVE',
  'AUDITOR',
  'AUDITEE',
  'DEPARTMENT_HEAD',
  'VIEWER'
]

const SUGGESTED_ACCESS_ROLE: Record<JobPositionType, string> = {
  STAFF: 'AUDITEE',
  SUPERVISOR: 'AUDITEE',
  MANAGER: 'AUDITEE',
  SENIOR_MANAGER: 'EXECUTIVE',
  GENERAL_MANAGER: 'EXECUTIVE',
  DIRECTOR: 'EXECUTIVE',
  COMMISSIONER: 'EXECUTIVE',
  AUDIT_COMMITTEE: 'EXECUTIVE',
  CHIEF_AUDIT_EXECUTIVE: 'EXECUTIVE',
  EXTERNAL_AUDITOR: ''
}

/** Access role to pre-select for a position type ('' = no suggestion). The user can override it. */
export const suggestAccessRole = (positionType?: string | null): string =>
  (positionType && SUGGESTED_ACCESS_ROLE[positionType as JobPositionType]) || ''

/**
 * Normalise the access-roles helper response: accepts `["A"]`, `[{name}]` or `{ access_roles | roles: [...] }`,
 * optionally wrapped in `data`. Returns [] when nothing usable is found.
 */
export const normalizeAccessRoles = (payload: any): string[] => {
  const body = payload?.data ?? payload
  const list = Array.isArray(body) ? body : (body?.access_roles ?? body?.roles ?? [])
  if (!Array.isArray(list)) return []
  const names = list
    .map((r: any) => (typeof r === 'string' ? r : (r?.name ?? r?.role_name ?? r?.value ?? '')))
    .map((n: string) => String(n).trim())
    .filter(Boolean)
  return Array.from(new Set(names))
}

/** Translate `key`, falling back to the raw value when the key has no translation (e.g. a new backend role). */
export const labelOrRaw = (t: (key: string) => string, key: string, raw: string): string => {
  const label = t(key)
  return label === key ? raw : label
}

export const accessRoleColor = (role?: string | null): 'primary' | 'info' | 'success' | 'warning' | 'neutral' => {
  switch (role) {
    case 'EXECUTIVE': return 'primary'
    case 'AUDITOR': return 'info'
    case 'AUDITEE': return 'success'
    case 'DEPARTMENT_HEAD': return 'warning'
    default: return 'neutral'
  }
}
