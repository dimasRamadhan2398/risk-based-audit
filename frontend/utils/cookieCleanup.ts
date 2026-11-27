const LEGACY_DASHBOARD_COOKIE_RE = /^dashboard-(sidebar|panel)-/
const PROXIED_API_PREFIXES = ['/api/v1', '/api/analytics', '/api/python-ai', '/uploads']

export function isProxiedApiPath(rawPath: string | undefined): boolean {
  if (!rawPath) return false
  const path = rawPath.split('?')[0] || ''
  return PROXIED_API_PREFIXES.some(prefix => path === prefix || path.startsWith(`${prefix}/`))
}

/**
 * Expire legacy `@nuxt/ui` dashboard layout cookies (`dashboard-sidebar-*`, `dashboard-panel-*`)
 * that accumulated before `<UDashboardGroup storage="local">` was enabled.
 * Leaves auth (`auth-token`, `auth-user`), `locale`, and unrelated cookies untouched.
 */
export function purgeLegacyDashboardCookies(doc: Pick<Document, 'cookie'> | undefined = typeof document !== 'undefined' ? document : undefined): string[] {
  if (!doc || typeof doc.cookie !== 'string' || !doc.cookie) return []

  const removed: string[] = []
  for (const part of doc.cookie.split(';')) {
    const eqIdx = part.indexOf('=')
    const name = (eqIdx >= 0 ? part.slice(0, eqIdx) : part).trim()
    if (!name || !LEGACY_DASHBOARD_COOKIE_RE.test(name)) continue
    doc.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; path=/`
    removed.push(name)
  }
  return removed
}
