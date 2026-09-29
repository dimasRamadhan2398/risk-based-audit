/**
 * Keeps a user whose password was reset by an admin on the change password
 * page until they replace the temporary password. Runs after 00-auth-init so
 * the session is already restored from cookies.
 */
const ALLOWED_PATHS = ['/auth/change-password', '/auth/login', '/auth/confidentiality']

export default defineNuxtRouteMiddleware((to) => {
  const authStore = useAuthStore()

  if (!authStore.isAuthenticated || !authStore.user?.mustChangePassword) {
    return
  }

  if (!ALLOWED_PATHS.includes(to.path)) {
    return navigateTo('/auth/change-password')
  }
})
