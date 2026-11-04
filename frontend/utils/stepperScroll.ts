/**
 * Horizontal auto-scroll for the executive summary stepper header.
 * Keeps the active step inside the visible strip of its scroll container.
 */

export interface HorizontalBounds {
  left: number
  right: number
}

/**
 * Pixels to scroll the container so the target is visible with `padding` to spare on each side.
 * Returns 0 when the target is already visible. The left edge wins when the target is wider than the view.
 */
export const horizontalScrollDelta = (view: HorizontalBounds, target: HorizontalBounds, padding = 16): number => {
  if (target.left < view.left + padding) return target.left - (view.left + padding)
  if (target.right > view.right - padding) return target.right - (view.right - padding)
  return 0
}

/**
 * Scrolls only the horizontal axis of `container` so `step` is visible. Never touches vertical scroll or ancestors.
 */
export const scrollStepIntoView = (container: HTMLElement, step: HTMLElement, behavior: ScrollBehavior = 'smooth'): void => {
  const view = container.getBoundingClientRect()
  const target = step.getBoundingClientRect()
  const delta = horizontalScrollDelta(
    { left: view.left, right: view.right },
    { left: target.left, right: target.right }
  )
  if (delta !== 0) container.scrollBy({ left: delta, top: 0, behavior })
}
