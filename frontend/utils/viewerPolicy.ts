/**
 * VIEWER read-only policy (UI side). The backend is authoritative; this only decides what the UI shows.
 * VIEWER: view business features only; no create/edit/delete/approve/publish/import/upload, no file access
 * (download, preview, attachments), no exports, no master data, and only Profile/Activity/FAQ in Account Settings.
 */

const clean = (r: unknown) => String(r || '').toLowerCase().replace(/[\s_-]+/g, '')

// Any of these alongside VIEWER means the user is not "viewer only" (the more privileged role wins).
const NON_VIEWER_ROLES = [
  'admin', 'administrator', 'superadmin', 'executive', 'chiefauditexecutive', 'cae',
  'auditor', 'auditstaff', 'auditee', 'pic', 'departmenthead'
]

/** True when the role list contains VIEWER (case-insensitive, exact) and no other business role. */
export function isViewerOnly(roles: string[]): boolean {
  const cleaned = roles.map(clean).filter(Boolean)
  if (!cleaned.includes('viewer')) return false
  return !cleaned.some(r => NON_VIEWER_ROLES.includes(r))
}

export const SETTINGS_TABS_ALL = ['profile', 'settings', 'mfa', 'activity', 'permissions', 'datasource', 'faq'] as const
export const SETTINGS_TABS_VIEWER = ['profile', 'activity', 'faq'] as const

/** Account Settings tabs visible to the role. */
export function getSettingsTabs(viewer: boolean): string[] {
  return [...(viewer ? SETTINGS_TABS_VIEWER : SETTINGS_TABS_ALL)]
}

/** Resolve a ?tab= query to an allowed tab, falling back to profile. */
export function resolveSettingsTab(query: unknown, viewer: boolean): string {
  const tab = Array.isArray(query) ? query[0] : query
  return typeof tab === 'string' && getSettingsTabs(viewer).includes(tab) ? tab : 'profile'
}

/** Master data pages (/master/*). */
export function isMasterDataPath(path: string): boolean {
  const p = (path || '').split('?')[0]!.replace(/\/+$/, '')
  return p === '/master' || p.startsWith('/master/')
}

/** Paths a VIEWER must not open: master data and import/upload screens (write features). */
export function isViewerBlockedPath(path: string): boolean {
  const p = (path || '').split('?')[0]!.replace(/\/+$/, '')
  if (isMasterDataPath(p)) return true
  if (p === '/import-working-paper' || p.startsWith('/import-working-paper/')) return true
  if (p === '/audit-assignment/upload') return true
  return /\/(upload|import)(-[a-z-]+)?$/.test(p)
}

/**
 * Write rules per feature area. Rights are the UNION of a user's roles and VIEWER contributes nothing.
 * The backend is authoritative; these lists only decide what the UI shows. Role names are compared
 * case-insensitively through useRbac().hasAnyRole.
 */
/** Master data writes: companies, business units, departments, employees, job roles, locations, vision-mission-goals. */
export const MASTER_DATA_WRITE_ROLES = ['ADMIN']
/** QA reports and consulting services writes (add, edit, delete, import, upload). */
export const QA_CONSULTING_WRITE_ROLES = ['ADMIN', 'EXECUTIVE', 'CHIEF_AUDIT_EXECUTIVE', 'AUDIT_MANAGER']
/** Audit universe edits on everything (establish, sync from master, add, rename, remove). */
export const AUDIT_UNIVERSE_WRITE_ROLES = ['ADMIN', 'EXECUTIVE', 'CHIEF_AUDIT_EXECUTIVE']
/**
 * Roles that may edit audit universe and library rows ONLY when the row belongs to their own department
 * (never sync from master). A row with no department data stays read-only for these roles.
 */
export const AUDIT_UNIVERSE_OWN_DEPT_ROLES = ['DEPARTMENT_HEAD']

type DeptUser = { departmentId?: string | null, department?: string | null } | null | undefined
type DeptTarget = string | { departmentId?: string | null, department_id?: string | null, department_name?: string | null, department?: string | { id?: string | null, name?: string | null } | null } | null | undefined

const norm = (v: unknown) => String(v ?? '').trim().toLowerCase()

/**
 * True when the user's department matches the row's department. The department id decides when both sides have
 * one; the name (trimmed, case-insensitive) is used only when the id is missing on one side. A row with no
 * department (null id and no name) never matches, nor does a user without a department.
 */
export function isOwnDepartment(user: DeptUser, target: DeptTarget): boolean {
  if (!user || target === null || target === undefined) return false
  const myId = norm(user.departmentId)
  const myName = norm(user.department)
  if (typeof target === 'string') {
    const v = norm(target)
    return !!v && (v === myId || v === myName)
  }
  let rowId = norm(target.department_id) || norm(target.departmentId)
  let rowName = norm(target.department_name)
  const d = target.department
  if (typeof d === 'string') rowName = rowName || norm(d)
  else if (d) {
    rowId = rowId || norm(d.id)
    rowName = rowName || norm(d.name)
  }
  if (myId && rowId) return myId === rowId
  return !!myName && !!rowName && myName === rowName
}
/** QA import pages and the consulting upload page (writes limited to QA_CONSULTING_WRITE_ROLES). */
export function isQaConsultingWritePath(path: string): boolean {
  const p = (path || '').split('?')[0]!.replace(/\/+$/, '')
  return p === '/consulting-service/upload' || /^\/quality-assurance\/import(-[a-z-]+)?$/.test(p)
}
