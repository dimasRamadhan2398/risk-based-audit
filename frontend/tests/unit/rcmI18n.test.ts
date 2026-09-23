import { describe, it, expect } from 'vitest'
import en from '~/locales/en/common.json'
import id from '~/locales/id/common.json'

/**
 * Helper to recursively collect all leaf keys from a nested object.
 * Returns an array of dot-separated key paths.
 */
function collectKeys(obj: any, prefix = ''): string[] {
  const keys: string[] = []
  for (const key in obj) {
    const fullKey = prefix ? `${prefix}.${key}` : key
    if (typeof obj[key] === 'object' && obj[key] !== null && !Array.isArray(obj[key])) {
      keys.push(...collectKeys(obj[key], fullKey))
    } else {
      keys.push(fullKey)
    }
  }
  return keys
}

/**
 * Helper to get a nested value from an object by dot-separated key.
 */
function getNestedValue(obj: any, keyPath: string): any {
  const parts = keyPath.split('.')
  let current = obj
  for (const part of parts) {
    if (current === undefined || current === null) return undefined
    current = current[part]
  }
  return current
}

describe('Risk Control Matrix (RCM) i18n Translations', () => {
  const rcmEn = (en as any).rcm
  const rcmId = (id as any).rcm

  it('should have rcm section in both locales', () => {
    expect(rcmEn).toBeDefined()
    expect(rcmId).toBeDefined()
  })

  it('should have all required top-level sections in EN', () => {
    const expectedSections = [
      'breadcrumb',
      'header',
      'filters',
      'cards',
      'dimensions',
      'interpretationTable',
      'table',
      'modal',
      'toasts',
      'confirmDelete',
      'defaults'
    ]
    for (const section of expectedSections) {
      expect(rcmEn[section], `EN missing section: rcm.${section}`).toBeDefined()
    }
  })

  it('should have all required top-level sections in ID', () => {
    const expectedSections = [
      'breadcrumb',
      'header',
      'filters',
      'cards',
      'dimensions',
      'interpretationTable',
      'table',
      'modal',
      'toasts',
      'confirmDelete',
      'defaults'
    ]
    for (const section of expectedSections) {
      expect(rcmId[section], `ID missing section: rcm.${section}`).toBeDefined()
    }
  })

  describe('EN ↔ ID parity', () => {
    const enKeys = collectKeys(rcmEn)
    const idKeys = collectKeys(rcmId)

    it('every EN rcm key should have an ID counterpart', () => {
      const missingInId: string[] = []
      for (const key of enKeys) {
        if (getNestedValue(rcmId, key) === undefined) {
          missingInId.push(key)
        }
      }
      expect(missingInId, `Keys in EN but missing in ID: ${missingInId.join(', ')}`).toEqual([])
    })

    it('every ID rcm key should have an EN counterpart', () => {
      const missingInEn: string[] = []
      for (const key of idKeys) {
        if (getNestedValue(rcmEn, key) === undefined) {
          missingInEn.push(key)
        }
      }
      expect(missingInEn, `Keys in ID but missing in EN: ${missingInEn.join(', ')}`).toEqual([])
    })
  })

  describe('COSO 5 Dimensions & Interpretation Ratings', () => {
    it('should have all 5 dimensions defined with short name, full name, and description', () => {
      const dims = ['design', 'operating', 'coverage', 'timeliness', 'automation']
      for (const dim of dims) {
        expect(rcmEn.dimensions[dim]).toBeDefined()
        expect(rcmId.dimensions[dim]).toBeDefined()
        expect(rcmEn.dimensions[`${dim}FullName`]).toBeDefined()
        expect(rcmId.dimensions[`${dim}FullName`]).toBeDefined()
        expect(rcmEn.dimensions[`${dim}Desc`]).toBeDefined()
        expect(rcmId.dimensions[`${dim}Desc`]).toBeDefined()
      }
    })

    it('should have all 5 effectiveness ratings and descriptions', () => {
      const ratings = ['highlyEffective', 'effective', 'moderatelyEffective', 'weak', 'ineffective']
      for (const rating of ratings) {
        expect(rcmEn.interpretationTable.ratings[rating]).toBeDefined()
        expect(rcmId.interpretationTable.ratings[rating]).toBeDefined()
        expect(rcmEn.interpretationTable.descriptions[rating]).toBeDefined()
        expect(rcmId.interpretationTable.descriptions[rating]).toBeDefined()
      }
    })
  })

  describe('Table and Modal elements', () => {
    it('should have all table columns translated', () => {
      const columns = [
        'riskCode',
        'controlCode',
        'department',
        'design',
        'operating',
        'coverage',
        'timeliness',
        'automation',
        'totalScore',
        'rating',
        'actions'
      ]
      for (const col of columns) {
        expect(rcmEn.table.columns[col]).toBeDefined()
        expect(rcmId.table.columns[col]).toBeDefined()
      }
    })

    it('should have all toasts translated', () => {
      const toasts = ['validationWarning', 'updateSuccess', 'addSuccess', 'deleteSuccess']
      for (const toast of toasts) {
        expect(rcmEn.toasts[toast]).toBeDefined()
        expect(rcmId.toasts[toast]).toBeDefined()
      }
    })
  })
})
