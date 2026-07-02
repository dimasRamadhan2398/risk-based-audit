export default defineNuxtPlugin(async () => {
  const authStore = useAuthStore()

  // Restore session from cookies on app initialization
  await authStore.fetchUser()

  return {
    provide: {
      authStore,
    },
  }
})
