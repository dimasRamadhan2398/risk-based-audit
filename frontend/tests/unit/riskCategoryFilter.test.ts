import { describe, it, expect } from 'vitest'
import {
  ALL_RISK_CATEGORIES,
  filterRisksByCategory,
  getDistinctRiskCategories,
  getRiskCategoryI18nKey
} from '~/utils/riskCategoryFilter'

const risks = [
  { id: '1', name: 'Fraud in procurement', category: 'Financial' },
  { id: '2', name: 'Data breach', category: 'Technology' },
  { id: '3', name: 'Cash shortfall', category: 'Financial' },
  { id: '4', name: 'Unknown bucket', category: '  ' },
  { id: '5', name: 'Legacy row', category: 'Human Resources' },
  { id: '6', name: 'Padded name', category: ' Technology ' },
  { id: '7', name: 'No category' }
]

describe('riskCategoryFilter', () => {
  describe('getDistinctRiskCategories', () => {
    it('returns unique, trimmed, non-empty categories sorted for display', () => {
      expect(getDistinctRiskCategories(risks)).toEqual(['Financial', 'Human Resources', 'Technology'])
    })

    it('returns an empty list when there are no risks', () => {
      expect(getDistinctRiskCategories([])).toEqual([])
    })
  })

  describe('filterRisksByCategory', () => {
    it('keeps every row for the All option', () => {
      expect(filterRisksByCategory(risks, ALL_RISK_CATEGORIES)).toHaveLength(risks.length)
    })

    it('keeps every row when no category is selected', () => {
      expect(filterRisksByCategory(risks, null)).toHaveLength(risks.length)
      expect(filterRisksByCategory(risks, undefined)).toHaveLength(risks.length)
      expect(filterRisksByCategory(risks, '')).toHaveLength(risks.length)
    })

    it('returns only risks with the exact selected category', () => {
      expect(filterRisksByCategory(risks, 'Financial').map(r => r.id)).toEqual(['1', '3'])
      expect(filterRisksByCategory(risks, 'Human Resources').map(r => r.id)).toEqual(['5'])
    })

    it('matches categories ignoring stored surrounding whitespace', () => {
      expect(filterRisksByCategory(risks, 'Technology').map(r => r.id)).toEqual(['2', '6'])
    })

    it('returns an empty list for a category not present in the data', () => {
      expect(filterRisksByCategory(risks, 'Governance')).toEqual([])
    })

    it('does not mutate the input array', () => {
      const copy = [...risks]
      filterRisksByCategory(risks, 'Financial')
      expect(risks).toEqual(copy)
    })

    it('returns a new array for the All option', () => {
      const result = filterRisksByCategory(risks, ALL_RISK_CATEGORIES)
      expect(result).not.toBe(risks)
    })
  })

  describe('getRiskCategoryI18nKey', () => {
    it('maps known categories to their translation key', () => {
      expect(getRiskCategoryI18nKey('Human Resources')).toBe('humanResources')
      expect(getRiskCategoryI18nKey('Financial')).toBe('financial')
    })

    it('returns null for unknown categories so the raw name is shown', () => {
      expect(getRiskCategoryI18nKey('Legal')).toBeNull()
    })
  })
})
