export default defineNuxtRouteMiddleware((to, from) => {
  const authStore = useAuthStore()

  if (!authStore._initialized) {
    return
  }

  if (!authStore.isAuthenticated) {
    return navigateTo('/auth/login')
  }

  // F-05: Redirect to confidentiality agreement if not yet accepted
  // (but allow access to the agreement page itself)
  if (authStore.needsConfidentialityAgreement && to.path !== '/auth/confidentiality') {
    return navigateTo('/auth/confidentiality')
  }

  return
})

