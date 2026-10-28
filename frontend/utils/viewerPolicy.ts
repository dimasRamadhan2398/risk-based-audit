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
