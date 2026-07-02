/**
 * This middleware runs first (00- prefix ensures it runs before other middleware)
 * and waits for auth initialization to complete before allowing navigation.
 */
export default defineNuxtRouteMiddleware(async (to, from) => {
  const authStore = useAuthStore();

  // Wait for auth to be initialized from cookies
  if (!authStore._initialized) {
    await authStore.fetchUser();
  }
});
