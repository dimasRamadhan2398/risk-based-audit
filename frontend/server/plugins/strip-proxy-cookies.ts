import { isProxiedApiPath } from '../../utils/cookieCleanup'

/**
 * Strip the incoming browser Cookie header before Nitro proxies requests to Kong / Python AI.
 *
 * Every AuditSphere backend authenticates strictly via `Authorization: Bearer <token>` and
 * refuses cookie auth. Forwarding the browser's entire cookie jar (including third-party
 * localhost dev cookies and UI state cookies) serves no purpose and triggers HTTP 400
 * ("client sent too long header line") in Kong/Nginx whenever the Cookie header exceeds 8 KB.
 */
export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('request', (event) => {
    const path = event.path || event.node.req.url || ''
    if (isProxiedApiPath(path) && event.node.req.headers.cookie) {
      delete event.node.req.headers.cookie
    }
  })
})
