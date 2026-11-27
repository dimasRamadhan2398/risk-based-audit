import { purgeLegacyDashboardCookies } from '~/utils/cookieCleanup'

export default defineNuxtPlugin(async () => {
  // Clean up accumulated @nuxt/ui dashboard-sidebar-* / dashboard-panel-* cookies
  // so browser requests never bloat past proxy/gateway header limits.
  purgeLegacyDashboardCookies()

  const authStore = useAuthStore()

  // Restore session from cookies on app initialization
  await authStore.fetchUser()

  // The avatar is not stored in the cookie; load it in the background
  void authStore.hydrateProfile()

  return {
    provide: {
      authStore,
    },
  }
})
