import { triggerScrollReset } from '~/utils/scroll'

export default defineNuxtPlugin((nuxtApp) => {
  const router = useRouter()

  // Reset scroll on every route navigation
  router.afterEach((to, from) => {
    if (to.path !== from.path || to.fullPath !== from.fullPath) {
      triggerScrollReset()
    }
  })

  // Hook into Nuxt page finish lifecycle event
  nuxtApp.hook('page:finish', () => {
    triggerScrollReset()
  })

  // Hook into Nuxt app mounted
  nuxtApp.hook('app:mounted', () => {
    triggerScrollReset()
  })
})
