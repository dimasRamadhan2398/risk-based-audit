/**
 * Resets the scroll position to the top of the page.
 * Handles window/document scrolling as well as Nuxt UI UDashboardPanel's
 * inner scrollable container ([data-slot="body"]).
 */
export const resetScrollPosition = () => {
  if (typeof window === 'undefined') return

  // 1. Reset standard window and document root scroll
  window.scrollTo({ top: 0, left: 0, behavior: 'instant' })
  if (document.documentElement) {
    document.documentElement.scrollTop = 0
    document.documentElement.scrollLeft = 0
  }
  if (document.body) {
    document.body.scrollTop = 0
    document.body.scrollLeft = 0
  }

  // 2. Reset UDashboardPanel body containers (data-slot="body")
  const dashboardBodies = document.querySelectorAll('[data-slot="body"]')
  dashboardBodies.forEach((el) => {
    if (el instanceof HTMLElement) {
      el.scrollTop = 0
      el.scrollLeft = 0
    }
  })

  // 3. Reset any main content wrappers
  const mainContainers = document.querySelectorAll('main, [role="main"]')
  mainContainers.forEach((el) => {
    if (el instanceof HTMLElement) {
      el.scrollTop = 0
      el.scrollLeft = 0
    }
  })
}

/**
 * Triggers a multi-stage scroll reset (immediate, nextTick, RAF, and short timeout)
 * to ensure scroll is reset even with asynchronous DOM updates, route transitions, and layout renders.
 */
export const triggerScrollReset = () => {
  if (typeof window === 'undefined') return

  // 1. Immediate reset
  resetScrollPosition()

  // 2. Next animation frame
  requestAnimationFrame(() => {
    resetScrollPosition()
  })

  // 3. Short timeouts to handle async content / transition render
  setTimeout(() => {
    resetScrollPosition()
  }, 50)

  setTimeout(() => {
    resetScrollPosition()
  }, 150)
}
