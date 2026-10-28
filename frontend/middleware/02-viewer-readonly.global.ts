/**
 * VIEWER read-only policy: master data pages and import/upload screens are not reachable for VIEWER,
 * including by typing the URL. Redirects to the dashboard. The backend enforces the same rules.
 * Runs after 00-auth-init so the session (and role) is already restored.
 */
import { isViewerBlockedPath } from '~/utils/viewerPolicy'

export default defineNuxtRouteMiddleware((to) => {
  const authStore = useAuthStore()
  if (!authStore.isAuthenticated) return

  const { isViewer } = useRbac()
  if (isViewer.value && isViewerBlockedPath(to.path)) {
    return navigateTo('/dashboard')
  }
})
