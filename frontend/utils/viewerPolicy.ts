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
type DeptTarget = string | { departmentId?: string | null, department_id?: string | null, department?: string | { id?: string | null, name?: string | null } | null } | null | undefined

const norm = (v: unknown) => String(v ?? '').trim().toLowerCase()

/**
 * True when the user's department matches the row's department. Matches by id when the row carries one
 * (department_id / departmentId / department.id), otherwise by trimmed, case-insensitive name. A missing
 * department on either side never matches.
 */
export function isOwnDepartment(user: DeptUser, target: DeptTarget): boolean {
  if (!user || target === null || target === undefined) return false
  const ids: string[] = []
  const names: string[] = []
  if (typeof target === 'string') {
    ids.push(norm(target))
    names.push(norm(target))
  } else {
    ids.push(norm(target.department_id), norm(target.departmentId))
    const d = target.department
    if (typeof d === 'string') names.push(norm(d))
    else if (d) {
      ids.push(norm(d.id))
      names.push(norm(d.name))
    }
  }
  const myId = norm(user.departmentId)
  const myName = norm(user.department)
  if (myId && ids.includes(myId)) return true
  return !!myName && names.includes(myName)
}

/** QA import pages and the consulting upload page (writes limited to QA_CONSULTING_WRITE_ROLES). */
export function isQaConsultingWritePath(path: string): boolean {
  const p = (path || '').split('?')[0]!.replace(/\/+$/, '')
  return p === '/consulting-service/upload' || /^\/quality-assurance\/import(-[a-z-]+)?$/.test(p)
}
