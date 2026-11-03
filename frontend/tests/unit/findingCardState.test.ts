import { describe, it, expect } from 'vitest'
import { createFindingExpansion } from '~/utils/findingCardState'

describe('createFindingExpansion (significant finding cards)', () => {
  const existingA = { judulTemuan: 'A' }
  const existingB = { judulTemuan: 'B' }
  const existing = [existingA, existingB]

  it('starts with every existing card collapsed', () => {
    const state = createFindingExpansion()
    for (const f of existing) expect(state.isExpanded(f)).toBe(false)
  })

  it('expands only the newly added finding', () => {
    const state = createFindingExpansion()
    const added = { judulTemuan: '' }
    state.expand(added)
    expect(state.isExpanded(added)).toBe(true)
    for (const f of existing) expect(state.isExpanded(f)).toBe(false)
  })

  it('keeps an existing card open or closed as the user left it when another card is added', () => {
    const state = createFindingExpansion()
    state.setExpanded(existingA, true)
    state.expand({ judulTemuan: '' })
    expect(state.isExpanded(existingA)).toBe(true)
    expect(state.isExpanded(existingB)).toBe(false)
  })

  it('collapses a card when the user closes it, and reopens it on request', () => {
    const state = createFindingExpansion()
    const added = { judulTemuan: '' }
    state.expand(added)
    state.setExpanded(added, false)
    expect(state.isExpanded(added)).toBe(false)
    state.setExpanded(added, true)
    expect(state.isExpanded(added)).toBe(true)
  })

  it('does not move the open state to another card when an earlier card is removed', () => {
    const state = createFindingExpansion()
    const first = { judulTemuan: 'first' }
    const second = { judulTemuan: 'second' }
    state.expand(second)
    state.forget(first)
    expect(state.isExpanded(second)).toBe(true)
    expect(state.isExpanded(first)).toBe(false)
  })

  it('ignores expand() without a finding', () => {
    const state = createFindingExpansion()
    expect(() => state.expand(undefined)).not.toThrow()
  })
})
