export default defineNuxtPlugin(async () => {
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
