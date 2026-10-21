/**
 * Numbered lists of the Formulir Rencana Kegiatan Audit (Justifikasi, Tujuan, Fokus).
 * Mirrors the audit-service normaliser (B2): trim, strip a leading "1." / "1)", drop empty items.
 */
import { describe, it, expect } from 'vitest'
import {
  NUMBERED_LIST_MAX_ITEMS,
  NUMBERED_LIST_MAX_ITEM_LENGTH,
  applyPastedText,
  checkNumberedList,
  joinListItems,
  normalizeListItem,
  normalizeListItems,
  splitNumberedText,
  toListItems
} from '~/utils/numberedList'

describe('normalise', () => {
  it('splits "1. a\\n2) b\\n\\n" into ["a","b"] like the server', () => {
    expect(splitNumberedText('1. a\n2) b\n\n')).toEqual(['a', 'b'])
  })

  it('trims, strips one leading number and drops empty items', () => {
    expect(normalizeListItems(['  1. BOD & BOC Concern ', '2)RKAP 2025', '', '   ', '10.  Top risk'])).toEqual([
      'BOD & BOC Concern',
      'RKAP 2025',
      'Top risk'
    ])
  })

  it('keeps numbers that are not a list marker', () => {
    expect(normalizeListItem('RKAP 2025')).toBe('RKAP 2025')
    expect(normalizeListItem('2025 target')).toBe('2025 target')
    expect(normalizeListItem('Top risk 2024-2025: kilang')).toBe('Top risk 2024-2025: kilang')
  })

  it('handles CRLF and null input', () => {
    expect(splitNumberedText('1. a\r\n2. b')).toEqual(['a', 'b'])
    expect(splitNumberedText(null)).toEqual([])
    expect(normalizeListItems(undefined)).toEqual([])
  })

  it('joins items with "\\n" as the server stores them', () => {
    expect(joinListItems(['1. a', '', 'b'])).toBe('a\nb')
  })

  it('reads an API value given as an array or as legacy joined text', () => {
    expect(toListItems(['a', ' ', '2) b'])).toEqual(['a', 'b'])
    expect(toListItems('1. a\n2. b')).toEqual(['a', 'b'])
    expect(toListItems(null)).toEqual([])
  })
})

describe('paste', () => {
  it('a multi-line paste into an empty item replaces it with the items', () => {
    expect(applyPastedText(['x', ''], 1, '1. a\n2. b')).toEqual(['x', 'a', 'b'])
  })

  it('a multi-line paste into a filled item inserts the items after it', () => {
    expect(applyPastedText(['x', 'y'], 0, '1) a\n2) b')).toEqual(['x', 'a', 'b', 'y'])
  })

  it('a single line pastes normally', () => {
    expect(applyPastedText([''], 0, 'just text')).toBeNull()
    expect(applyPastedText([''], 0, 'just text\n')).toBeNull()
  })
})

describe('limits', () => {
  it('flags an empty list', () => {
    expect(checkNumberedList(['', '  ']).empty).toBe(true)
    expect(checkNumberedList(['a']).empty).toBe(false)
  })

  it('flags more than 20 items and items over 1000 characters', () => {
    expect(NUMBERED_LIST_MAX_ITEMS).toBe(20)
    expect(NUMBERED_LIST_MAX_ITEM_LENGTH).toBe(1000)
    expect(checkNumberedList(Array.from({ length: 21 }, (_, i) => `item ${i}`)).tooMany).toBe(true)
    expect(checkNumberedList(Array.from({ length: 20 }, (_, i) => `item ${i}`)).tooMany).toBe(false)
    expect(checkNumberedList(['x'.repeat(1001)]).tooLong).toBe(true)
    expect(checkNumberedList(['x'.repeat(1000)]).tooLong).toBe(false)
  })
})
