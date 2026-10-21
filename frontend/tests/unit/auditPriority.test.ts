import { describe, it, expect } from 'vitest'
import fs from 'fs'
import path from 'path'

describe('Audit Priority Feature Restructuring', () => {
  const rootDir = path.resolve(__dirname, '../../')

  it('verifies Audit Priority tab is removed from risk-factors/index.vue', () => {
    const filePath = path.join(rootDir, 'pages/risk-profile/risk-factors/index.vue')
    const content = fs.readFileSync(filePath, 'utf8')

    expect(content).not.toContain("slot: 'priority'")
    expect(content).not.toContain('<template #priority>')
    expect(content).toContain("slot: 'weighting'")
    expect(content).toContain("slot: 'scoring'")
  })

  it('verifies Audit Priority tab is removed from audit-universe/index.vue', () => {
    const filePath = path.join(rootDir, 'pages/risk-profile/audit-universe/index.vue')
    const content = fs.readFileSync(filePath, 'utf8')

    expect(content).not.toContain("value: 'priority'")
    expect(content).not.toContain("slot: 'priority'")
    expect(content).not.toContain('<template #priority>')
    expect(content).toContain("value: 'library'")
    expect(content).toContain("value: 'establish'")
  })

  it('verifies default.vue submenu navigation points to /risk-profile/audit-priority', () => {
    const filePath = path.join(rootDir, 'layouts/default.vue')
    const content = fs.readFileSync(filePath, 'utf8')

    expect(content).toContain("label: t('navigation.auditPriority')")
    expect(content).toContain("to: '/risk-profile/audit-priority'")
    expect(content).not.toContain("to: '/risk-profile/audit-universe?tab=priority'")
  })

  it('verifies dedicated Audit Priority page exists with core features', () => {
    const filePath = path.join(rootDir, 'pages/risk-profile/audit-priority/index.vue')
    expect(fs.existsSync(filePath)).toBe(true)

    const content = fs.readFileSync(filePath, 'utf8')

    // Year selection & Refresh
    expect(content).toContain('selectedYear')
    expect(content).toContain('loadData')

    // Table & Columns: priority and sorting go through the shared scoring helpers (unscored rows last, never prioritized)
    expect(content).toContain('filteredYearlyUniverse')
    expect(content).toContain('isUniverseYearPrioritized(ent)')
    expect(content).toContain('sortUniverseYearsByRisk(')
    expect(content).toContain('ent.risk_level')
    expect(content).toContain('riskFactors.priority.notScoredBadge')

    // Risk Index Level Reference comes from the one shared threshold table, not a local copy
    expect(content).toContain('levelInfoTitle')
    expect(content).toContain('RISK_LEVEL_THRESHOLDS')
    expect(content).toContain('formatRiskLevelRange(row)')
    expect(content).not.toMatch(/>=\s*80/)

    // Store integration
    expect(content).toContain('useAuditUniverseStore')
  })

  it('keeps the reference ranges of the shared thresholds', async () => {
    const { RISK_LEVEL_THRESHOLDS, formatRiskLevelRange } = await import('~/utils/riskLevelBadge')
    expect(RISK_LEVEL_THRESHOLDS.map(formatRiskLevelRange)).toEqual(['80 - 100%', '60 - 79%', '40 - 59%', '20 - 39%', '0 - 19%'])
    expect(RISK_LEVEL_THRESHOLDS.filter(th => th.prioritized).map(th => th.level)).toEqual(['High', 'Medium to High'])
  })

  it('verifies localization keys for auditPriority exist in en and id locales', () => {
    const enCommon = JSON.parse(fs.readFileSync(path.join(rootDir, 'locales/en/common.json'), 'utf8'))
    const idCommon = JSON.parse(fs.readFileSync(path.join(rootDir, 'locales/id/common.json'), 'utf8'))

    expect(enCommon.auditPriority).toBeDefined()
    expect(enCommon.auditPriority.title).toBe('Audit Priority')

    expect(idCommon.auditPriority).toBeDefined()
    expect(idCommon.auditPriority.title).toBe('Audit Priority')
  })
})
