/**
 * VIEWER read-only policy: master data pages and import/upload screens are not reachable for VIEWER,
 * including by typing the URL. Redirects to the dashboard. The backend enforces the same rules.
 * Also keeps platform-admin pages (tenant provisioning) away from every other role.
 * Runs after 00-auth-init so the session (and role) is already restored.
 */
import { isViewerBlockedPath, isQaConsultingWritePath, isPlatformAdminPath } from '~/utils/viewerPolicy'

export default defineNuxtRouteMiddleware((to) => {
  const authStore = useAuthStore()
  if (!authStore.isAuthenticated) return

  const { isViewer, canWriteQaConsulting, isPlatformAdmin } = useRbac()
  if (isViewer.value && isViewerBlockedPath(to.path)) {
    return navigateTo('/dashboard')
  }
  // QA import and consulting upload are write screens limited to ADMIN / EXECUTIVE / CAE / AUDIT_MANAGER.
  if (!canWriteQaConsulting.value && isQaConsultingWritePath(to.path)) {
    return navigateTo('/dashboard')
  }
  // Site Generator (tenant provisioning) is platform admin only; auth-service checks it again.
  if (!isPlatformAdmin.value && isPlatformAdminPath(to.path)) {
    return navigateTo('/dashboard')
  }
})
