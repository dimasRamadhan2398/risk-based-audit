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

describe('Audit Charter i18n Translations', () => {
  const auditCharterEn = (en as any).auditCharter
  const auditCharterId = (id as any).auditCharter

  it('should have auditCharter section in both locales', () => {
    expect(auditCharterEn).toBeDefined()
    expect(auditCharterId).toBeDefined()
  })

  it('should have all required top-level sections in EN', () => {
    const expectedSections = ['tabs', 'card', 'form', 'guideline', 'guidelineForm', 'sopList', 'sopForm']
    for (const section of expectedSections) {
      expect(auditCharterEn[section], `EN missing section: auditCharter.${section}`).toBeDefined()
    }
  })

  it('should have all required top-level sections in ID', () => {
    const expectedSections = ['tabs', 'card', 'form', 'guideline', 'guidelineForm', 'sopList', 'sopForm']
    for (const section of expectedSections) {
      expect(auditCharterId[section], `ID missing section: auditCharter.${section}`).toBeDefined()
    }
  })

  describe('EN ↔ ID parity', () => {
    const enKeys = collectKeys(auditCharterEn)
    const idKeys = collectKeys(auditCharterId)

    it('every EN auditCharter key should have an ID counterpart', () => {
      const missingInId: string[] = []
      for (const key of enKeys) {
        if (getNestedValue(auditCharterId, key) === undefined) {
          missingInId.push(key)
        }
      }
      expect(missingInId, `Keys in EN but missing in ID: ${missingInId.join(', ')}`).toEqual([])
    })

    it('every ID auditCharter key should have an EN counterpart', () => {
      const missingInEn: string[] = []
      for (const key of idKeys) {
        if (getNestedValue(auditCharterEn, key) === undefined) {
          missingInEn.push(key)
        }
      }
      expect(missingInEn, `Keys in ID but missing in EN: ${missingInEn.join(', ')}`).toEqual([])
    })
  })

  describe('Tabs section', () => {
    it('should have charter, guideline, sop keys', () => {
      expect(auditCharterEn.tabs.charter).toBeDefined()
      expect(auditCharterEn.tabs.guideline).toBeDefined()
      expect(auditCharterEn.tabs.sop).toBeDefined()
      expect(auditCharterId.tabs.charter).toBeDefined()
      expect(auditCharterId.tabs.guideline).toBeDefined()
      expect(auditCharterId.tabs.sop).toBeDefined()
    })
  })

  describe('Card section', () => {
    it('should have column definitions for charter history table', () => {
      const requiredColumns = ['version', 'title', 'content', 'date', 'approvedBy', 'uploadedBy', 'actions']
      for (const col of requiredColumns) {
        expect(auditCharterEn.card.columns[col], `EN missing card.columns.${col}`).toBeDefined()
        expect(auditCharterId.card.columns[col], `ID missing card.columns.${col}`).toBeDefined()
      }
    })
  })

  describe('Form section', () => {
    it('should have validation keys', () => {
      expect(auditCharterEn.form.docTitleValidation).toBeDefined()
      expect(auditCharterEn.form.approvedByValidation).toBeDefined()
      expect(auditCharterId.form.docTitleValidation).toBeDefined()
      expect(auditCharterId.form.approvedByValidation).toBeDefined()
    })

    it('should have selectedFile key', () => {
      expect(auditCharterEn.form.selectedFile).toBeDefined()
      expect(auditCharterId.form.selectedFile).toBeDefined()
    })
  })

  describe('Guideline section', () => {
    it('should have all required guideline list keys', () => {
      const requiredKeys = ['title', 'subtitle', 'emptyTitle', 'emptyDesc', 'addGuideline', 'addGuidelineShort', 'emptyTable', 'deleteConfirm', 'statusActive', 'statusUnderReview']
      for (const key of requiredKeys) {
        expect(auditCharterEn.guideline[key], `EN missing guideline.${key}`).toBeDefined()
        expect(auditCharterId.guideline[key], `ID missing guideline.${key}`).toBeDefined()
      }
    })

    it('should have all column definitions', () => {
      const requiredColumns = ['no', 'name', 'status', 'effectiveDate', 'fileName', 'actions']
      for (const col of requiredColumns) {
        expect(auditCharterEn.guideline.columns[col], `EN missing guideline.columns.${col}`).toBeDefined()
        expect(auditCharterId.guideline.columns[col], `ID missing guideline.columns.${col}`).toBeDefined()
      }
    })
  })

  describe('Guideline Form section', () => {
    it('should have all required guidelineForm keys', () => {
      const requiredKeys = ['editTitle', 'addTitle', 'name', 'namePlaceholder', 'nameValidation', 'status', 'statusActive', 'statusUnderReview', 'effectiveDate', 'fileUpload', 'selectedFile', 'saveChanges', 'addGuideline', 'cancel']
      for (const key of requiredKeys) {
        expect(auditCharterEn.guidelineForm[key], `EN missing guidelineForm.${key}`).toBeDefined()
        expect(auditCharterId.guidelineForm[key], `ID missing guidelineForm.${key}`).toBeDefined()
      }
    })
  })

  describe('SOP List section', () => {
    it('should have all required sopList keys', () => {
      const requiredKeys = ['title', 'subtitle', 'emptyTitle', 'emptyDesc', 'addSop', 'addSopShort', 'emptyTable', 'deleteConfirm', 'statusActive', 'statusUnderReview']
      for (const key of requiredKeys) {
        expect(auditCharterEn.sopList[key], `EN missing sopList.${key}`).toBeDefined()
        expect(auditCharterId.sopList[key], `ID missing sopList.${key}`).toBeDefined()
      }
    })

    it('should have all column definitions', () => {
      const requiredColumns = ['no', 'name', 'guidelineName', 'status', 'effectiveDate', 'actions']
      for (const col of requiredColumns) {
        expect(auditCharterEn.sopList.columns[col], `EN missing sopList.columns.${col}`).toBeDefined()
        expect(auditCharterId.sopList.columns[col], `ID missing sopList.columns.${col}`).toBeDefined()
      }
    })
  })

  describe('SOP Form section', () => {
    it('should have all required sopForm keys', () => {
      const requiredKeys = ['editTitle', 'addTitle', 'name', 'namePlaceholder', 'nameValidation', 'parentGuideline', 'parentGuidelinePlaceholder', 'status', 'active', 'underReview', 'effectiveDate', 'fileUpload', 'selectedFile', 'saveChanges', 'addSop', 'cancel']
      for (const key of requiredKeys) {
        expect(auditCharterEn.sopForm[key], `EN missing sopForm.${key}`).toBeDefined()
        expect(auditCharterId.sopForm[key], `ID missing sopForm.${key}`).toBeDefined()
      }
    })
  })

  describe('No empty string values (except column actions)', () => {
    const enKeys = collectKeys(auditCharterEn)
    const idKeys = collectKeys(auditCharterId)

    it('EN values should not be empty (except *.actions)', () => {
      for (const key of enKeys) {
        if (key.endsWith('.actions') || key === 'actions') continue
        const val = getNestedValue(auditCharterEn, key)
        expect(typeof val === 'string' && val.length > 0, `EN key "${key}" is empty`).toBe(true)
      }
    })

    it('ID values should not be empty (except *.actions)', () => {
      for (const key of idKeys) {
        if (key.endsWith('.actions') || key === 'actions') continue
        const val = getNestedValue(auditCharterId, key)
        expect(typeof val === 'string' && val.length > 0, `ID key "${key}" is empty`).toBe(true)
      }
    })
  })
})
