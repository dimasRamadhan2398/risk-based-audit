import { describe, it, expect, vi } from 'vitest'
import { horizontalScrollDelta, scrollStepIntoView } from '~/utils/stepperScroll'

describe('horizontalScrollDelta', () => {
  const view = { left: 100, right: 500 }

  it('returns 0 when the step is fully visible', () => {
    expect(horizontalScrollDelta(view, { left: 200, right: 300 })).toBe(0)
  })

  it('scrolls right when the step is past the right edge, keeping the padding', () => {
    expect(horizontalScrollDelta(view, { left: 480, right: 560 }, 16)).toBe(560 - (500 - 16))
  })

  it('scrolls left when the step is before the left edge, keeping the padding', () => {
    expect(horizontalScrollDelta(view, { left: 40, right: 90 }, 16)).toBe(40 - (100 + 16))
  })

  it('aligns the left edge first when the step is wider than the view', () => {
    expect(horizontalScrollDelta(view, { left: 40, right: 900 }, 0)).toBe(-60)
  })
})

describe('scrollStepIntoView', () => {
  const element = (rect: { left: number, right: number }) => ({
    getBoundingClientRect: () => ({ left: rect.left, right: rect.right, top: 0, bottom: 40, width: rect.right - rect.left, height: 40 })
  })

  it('scrolls only the horizontal axis of the container when the step is off-screen', () => {
    const container = { ...element({ left: 0, right: 300 }), scrollBy: vi.fn() }
    const step = element({ left: 320, right: 400 })
    scrollStepIntoView(container as unknown as HTMLElement, step as unknown as HTMLElement)
    expect(container.scrollBy).toHaveBeenCalledTimes(1)
    const arg = container.scrollBy.mock.calls[0]![0]
    expect(arg.top).toBe(0)
    expect(arg.left).toBeGreaterThan(0)
    expect(arg.behavior).toBe('smooth')
  })

  it('does not scroll when the step is already visible', () => {
    const container = { ...element({ left: 0, right: 300 }), scrollBy: vi.fn() }
    scrollStepIntoView(container as unknown as HTMLElement, element({ left: 50, right: 100 }) as unknown as HTMLElement)
    expect(container.scrollBy).not.toHaveBeenCalled()
  })
})
