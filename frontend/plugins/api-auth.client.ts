import {
  getAnalyticsServiceBaseUrl,
  getAuditServiceBaseUrl,
  getAuthServiceBaseUrl,
  getMasterServiceBaseUrl,
  getPythonAiBaseUrl,
  getRiskServiceBaseUrl,
} from '~/composables/useApiUrl'

/**
 * Attaches the logged-in user's token to every $fetch call made to the
 * AuditSphere APIs, so stores don't each have to pass the Authorization header
 * (most of them didn't, which only worked while the backend skipped auth).
 *
 * Only requests to our own API base URLs get the token; calls to any other
 * origin are left alone so the token never leaks to third parties.
 *
 * Client-only on purpose: the app runs with ssr: false, and replacing the
 * global $fetch on the server would share one user's token across requests.
 */
export default defineNuxtPlugin(() => {
  const authStore = useAuthStore()

  const apiBases = () => [
    getAuditServiceBaseUrl(),
    getMasterServiceBaseUrl(),
    getRiskServiceBaseUrl(),
    getAuthServiceBaseUrl(),
    getAnalyticsServiceBaseUrl(),
    getPythonAiBaseUrl(),
  ].filter(Boolean)

  const requestUrl = (request: RequestInfo | URL) =>
    typeof request === 'string' ? request : request instanceof URL ? request.href : request.url

  const isApiRequest = (url: string) => apiBases().some(base => url.startsWith(base))

  // Auth endpoints return 401 for a wrong password; that must not end the session
  const isAuthEndpoint = (url: string) => /\/auth\/(login|verify-mfa-login|change-password|register)\b/.test(url)

  globalThis.$fetch = $fetch.create({
    onRequest({ request, options }) {
      if (!authStore.token || !isApiRequest(requestUrl(request))) return

      const headers = new Headers(options.headers as HeadersInit | undefined)
      if (!headers.has('Authorization')) {
        headers.set('Authorization', `Bearer ${authStore.token}`)
        options.headers = headers
      }
    },
    async onResponseError({ request, response }) {
      const url = requestUrl(request)
      // Expired or revoked session: send the user back to login instead of
      // leaving every page failing with 401s
      if (response.status === 401 && authStore.isAuthenticated && isApiRequest(url) && !isAuthEndpoint(url)) {
        authStore._clearState()
        await navigateTo('/auth/login')
      }
    },
  }) as typeof $fetch
})
