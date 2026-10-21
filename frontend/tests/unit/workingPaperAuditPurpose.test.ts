/**
 * Audit Purpose of the Working Paper header: list <-> stored text, and the letter fallback.
 */
import { describe, it, expect } from 'vitest'
import {
  auditPurposeItemsFromLetter,
  auditPurposeItemsFromText,
  auditPurposeTextFromItems,
  resolveAuditPurposeText
} from '~/utils/workingPaperAuditPurpose'

describe('auditPurposeItemsFromText', () => {
  it('splits the stored text by line and strips "1." / "1)"', () => {
    expect(auditPurposeItemsFromText('1. Assess cash\n2) Review approvals\n')).toEqual([
      'Assess cash',
      'Review approvals'
    ])
  })

  it('keeps a single-line value as one item', () => {
    expect(auditPurposeItemsFromText('Annual Audit')).toEqual(['Annual Audit'])
  })

  it('returns no items for empty or missing text', () => {
    expect(auditPurposeItemsFromText('')).toEqual([])
    expect(auditPurposeItemsFromText(null)).toEqual([])
    expect(auditPurposeItemsFromText(undefined)).toEqual([])
  })
})

describe('auditPurposeTextFromItems', () => {
  it('joins trimmed, non-empty items with a newline', () => {
    expect(auditPurposeTextFromItems(['  Assess cash ', '', '2. Review approvals'])).toBe(
      'Assess cash\nReview approvals'
    )
  })

  it('returns an empty string when every item is blank', () => {
    expect(auditPurposeTextFromItems(['', '   '])).toBe('')
  })

  it('round-trips with auditPurposeItemsFromText', () => {
    const text = 'Assess cash\nReview approvals'
    expect(auditPurposeTextFromItems(auditPurposeItemsFromText(text))).toBe(text)
  })
})

describe('auditPurposeItemsFromLetter', () => {
  it('uses the saved single auditPurpose, split by line', () => {
    expect(
      auditPurposeItemsFromLetter({ auditPurpose: 'Annual Audit\nIT Review', purposeList: ['Ignored'] })
    ).toEqual(['Annual Audit', 'IT Review'])
  })

  it('falls back to purposeList, one item per entry, when auditPurpose is blank', () => {
    expect(
      auditPurposeItemsFromLetter({ auditPurpose: '  ', purposeList: ['Assess cash', ' ', 'Review procurement'] })
    ).toEqual(['Assess cash', 'Review procurement'])
  })

  it('keeps purposeList entries that contain a semicolon as separate items, not split', () => {
    expect(auditPurposeItemsFromLetter({ auditPurpose: '', purposeList: ['A; B'] })).toEqual(['A; B'])
  })

  it('returns no items for a letter without purpose', () => {
    expect(auditPurposeItemsFromLetter({ auditPurpose: '', purposeList: [] })).toEqual([])
    expect(auditPurposeItemsFromLetter(null)).toEqual([])
  })
})

describe('resolveAuditPurposeText', () => {
  it('stores one purpose per line for a UI-created letter', () => {
    expect(resolveAuditPurposeText({ auditPurpose: '', purposeList: ['Assess cash', 'Review procurement'] })).toBe(
      'Assess cash\nReview procurement'
    )
  })

  it('keeps a single seeded purpose unchanged', () => {
    expect(resolveAuditPurposeText({ auditPurpose: 'IT Security Audit', purposeList: [] })).toBe('IT Security Audit')
  })
})
