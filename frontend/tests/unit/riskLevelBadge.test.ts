import { describe, it, expect } from 'vitest'
import fs from 'fs'
import path from 'path'
import {
  getControlEffectivenessColorClass,
  getRiskLevelBadgeClass,
  getRiskLevelColorClass,
  getRiskLevelDotClass,
  isPriorityRiskLevel,
  normalizeRiskLevel,
  PRIORITY_RISK_LEVELS,
  RISK_LEVEL_LABEL_KEYS
} from '~/utils/riskLevelBadge'

describe('Risk level badge colours', () => {
  const expectedBg: Record<string, string[]> = {
    'High': ['bg-red-400', 'dark:bg-red-700'],
    'Medium to High': ['bg-orange-400', 'dark:bg-orange-700'],
    'Medium': ['bg-yellow-400', 'dark:bg-yellow-700'],
    'Low to Medium': ['bg-lime-400', 'dark:bg-lime-700'],
    'Low': ['bg-green-400', 'dark:bg-green-700']
  }

  it('covers exactly the five backend risk_level values', () => {
    expect(PRIORITY_RISK_LEVELS).toEqual(['High', 'Medium to High', 'Medium', 'Low to Medium', 'Low'])
    expect(Object.keys(RISK_LEVEL_LABEL_KEYS).sort()).toEqual([...PRIORITY_RISK_LEVELS].sort())
  })

  it.each(Object.entries(expectedBg))('maps %s to its background classes', (level, bgClasses) => {
    const classes = getRiskLevelColorClass(level).split(' ')
    for (const bg of bgClasses) expect(classes).toContain(bg)
  })

  it('uses black text in light mode and white text in dark mode for every level', () => {
    for (const level of PRIORITY_RISK_LEVELS) {
      const classes = getRiskLevelColorClass(level).split(' ')
      expect(classes).toContain('text-black')
      expect(classes).toContain('dark:text-white')
      expect(classes).not.toContain('text-white')
    }
  })

  it('gives every level a distinct background', () => {
    const backgrounds = PRIORITY_RISK_LEVELS.map(level =>
      getRiskLevelColorClass(level).split(' ').filter(c => c.startsWith('bg-')).join(' ')
    )
    expect(new Set(backgrounds).size).toBe(PRIORITY_RISK_LEVELS.length)
  })

  it.each([undefined, null, '', 'Not Scored', 'Custom'])('falls back to a neutral badge for %s', (level) => {
    const classes = getRiskLevelColorClass(level).split(' ')
    expect(classes).toContain('bg-slate-200')
    expect(classes).toContain('dark:bg-slate-700')
    expect(classes).toContain('text-black')
    expect(classes).toContain('dark:text-white')
    expect(isPriorityRiskLevel(level)).toBe(false)
  })

  it('only treats the exact backend values as canonical', () => {
    expect(isPriorityRiskLevel('High')).toBe(true)
    expect(isPriorityRiskLevel('high')).toBe(false)
  })

  it('shares one badge shape across levels, including the fallback', () => {
    const shape = (level?: string) => getRiskLevelBadgeClass(level).replace(getRiskLevelColorClass(level), '').trim()
    const shapes = new Set([...PRIORITY_RISK_LEVELS, undefined, 'Unknown'].map(shape))
    expect(shapes.size).toBe(1)
    expect([...shapes][0]).toContain('rounded')
  })
})

describe('Risk level normalisation', () => {
  it.each([
    // backend risk_level
    ['High', 'High'], ['Medium to High', 'Medium to High'], ['Medium', 'Medium'], ['Low to Medium', 'Low to Medium'], ['Low', 'Low'],
    // RiskLevel enum (stores/risk-profile, heatmap, dashboard, risk appetite)
    ['high', 'High'], ['moderate-high', 'Medium to High'], ['moderate', 'Medium'], ['low-moderate', 'Low to Medium'], ['low', 'Low'],
    // UI / AI spellings
    ['Moderate to High', 'Medium to High'], ['Medium-High', 'Medium to High'], ['MODERATE_HIGH', 'Medium to High'],
    ['Low to Moderate', 'Low to Medium'], ['Moderate', 'Medium'], ['Watch', 'Medium'],
    ['Critical', 'High'], ['Extreme', 'High'], ['Very High', 'High'], ['Very Low', 'Low']
  ])('normalises %s to %s', (input, expected) => {
    expect(normalizeRiskLevel(input)).toBe(expected)
  })

  it('never collapses a combined level into a single-word level', () => {
    // The old includes()-based helpers turned these into High / Moderate.
    expect(getRiskLevelColorClass('Moderate to High')).toBe(getRiskLevelColorClass('Medium to High'))
    expect(getRiskLevelColorClass('Low to Moderate')).toBe(getRiskLevelColorClass('Low to Medium'))
    expect(getRiskLevelColorClass('Moderate to High')).not.toBe(getRiskLevelColorClass('High'))
    expect(getRiskLevelColorClass('Low to Moderate')).not.toBe(getRiskLevelColorClass('Medium'))
  })

  it.each([undefined, null, '', 'Not Scored', 'Custom', 'Unknown'])('returns null for %s', (input) => {
    expect(normalizeRiskLevel(input)).toBeNull()
  })

  it('gives dots the same background as the badge, without text classes', () => {
    for (const level of PRIORITY_RISK_LEVELS) {
      const dot = getRiskLevelDotClass(level)
      expect(getRiskLevelColorClass(level).startsWith(dot)).toBe(true)
      expect(dot).not.toContain('text-')
    }
  })
})

describe('Control effectiveness colours', () => {
  it.each([
    ['Highly Effective', 'Low'],
    ['Effective', 'Low to Medium'],
    ['Moderately Effective', 'Medium'],
    ['Partially Effective', 'Medium'],
    ['Weak', 'Medium to High'],
    ['Ineffective', 'High']
  ])('colours %s like risk level %s', (rating, level) => {
    expect(getControlEffectivenessColorClass(rating)).toBe(getRiskLevelColorClass(level))
  })

  it('ignores case and spacing', () => {
    expect(getControlEffectivenessColorClass('highly effective')).toBe(getControlEffectivenessColorClass('Highly Effective'))
    expect(getControlEffectivenessColorClass('partially-effective')).toBe(getControlEffectivenessColorClass('Partially Effective'))
  })

  it.each([undefined, null, '', 'Not Tested'])('falls back to neutral for %s', (rating) => {
    expect(getControlEffectivenessColorClass(rating)).toBe(getRiskLevelColorClass(null))
  })

  it('uses black text in light mode and white text in dark mode', () => {
    for (const rating of ['Highly Effective', 'Effective', 'Moderately Effective', 'Weak', 'Ineffective', 'Not Tested']) {
      const classes = getControlEffectivenessColorClass(rating).split(' ')
      expect(classes).toContain('text-black')
      expect(classes).toContain('dark:text-white')
    }
  })
})

describe('Features render risk level and effectiveness badges through the shared helper', () => {
  const read = (rel: string) => fs.readFileSync(path.resolve(__dirname, '../..', rel), 'utf8')

  it.each([
    ['components/risk-profile/RiskHeatMap.vue', ['getQLevelBadgeClass', 'getRiskLevelColorClass(getRiskLevel(risk.likelihood, risk.impact))', 'getRiskLevelDotClass(key)']],
    ['pages/dashboard/index.vue', ['getRiskLevelColorClass(level)', 'getControlEffectivenessColorClass(ratingLabel)']],
    ['pages/risk-appetite/index.vue', ['getRiskLevelBadgeColorClass(row.original)', 'getRiskLevelDotClass(']],
    ['pages/risk-profile/risk-control-matrix/index.vue', ['getControlEffectivenessColorClass(ratingLabel)', 'getControlEffectivenessColorClass(\'Highly Effective\')', 'getControlEffectivenessColorClass(\'Ineffective\')']],
    ['components/annual-audit/AnnualAuditTable.vue', ['getRiskLevelColorClass(act.riskLevel)']],
    ['components/annual-audit/AnnualAuditDetail.vue', ['getRiskLevelColorClass(activity.riskLevel)']],
    ['components/annual-audit/AnnualAuditForm.vue', ['getRiskLevelColorClass(activity.riskLevel)', 'getRiskLevelDotClass(item.riskLevel)']],
    ['components/audit-activity-plan/ActivityPlanTable.vue', ['getRiskLevelColorClass(act.riskLevel)']],
    ['components/audit-activity-plan/ActivityPlanForm.vue', ['getRiskLevelDotClass(item.riskLevel)']],
    ['components/audit-fieldwork/AuditFieldworkTestControls.vue', ['getControlEffectivenessColorClass(row.original.testResult)']],
    ['components/audit-fieldwork/AuditFieldworkTestControlsModal.vue', ['getControlEffectivenessColorClass(store.testControlForm.testResult)']],
    ['components/working-paper/WorkingPaperSampleTable.vue', ['getControlEffectivenessColorClass(']],
    ['pages/analytics/ai/anomaly-detection.vue', ['getRiskConfig(a.riskLevel).badgeClass']],
    ['pages/analytics/ai/kpi-forecast.vue', ['getRiskConfig(dept.riskLevel).badgeClass', 'getRiskConfig(kpi.riskLevel).badgeClass']],
    ['pages/analytics/ai/nlp.vue', ['getRiskConfig(doc.riskLevel).badgeClass']],
    ['pages/analytics/ai/risk-scoring.vue', ['getRiskConfig(row.actualRiskLevel).badgeClass', 'getRiskConfig(row.predictedRiskLevel).badgeClass']]
  ])('%s uses the shared colours', (rel, snippets) => {
    const content = read(rel)
    // AI pages get the class through useAiAnalytics().getRiskConfig, checked separately below.
    if (!rel.startsWith('pages/analytics/ai/')) expect(content).toContain('~/utils/riskLevelBadge')
    for (const snippet of snippets) expect(content).toContain(snippet)
  })

  it.each([
    'components/annual-audit/AnnualAuditForm.vue',
    'components/audit-activity-plan/ActivityPlanForm.vue',
    'pages/risk-appetite/index.vue',
    'pages/analytics/ai/anomaly-detection.vue',
    'pages/analytics/ai/kpi-forecast.vue',
    'pages/analytics/ai/nlp.vue',
    'pages/analytics/ai/risk-scoring.vue'
  ])('%s no longer paints level badges with hex colours and white text', (rel) => {
    const content = read(rel)
    expect(content).not.toContain('getRiskLevelColorHex')
    expect(content).not.toMatch(/backgroundColor: getRisk\w*\([^)]*\)(\.color)?, color: 'white'/)
    expect(content).not.toMatch(/bg-\[#(F44336|FF9800|FFC107|8BC34A|4CAF50)\]/i)
  })

  it('no longer uses semantic Nuxt UI colours for test results or effectiveness ratings', () => {
    expect(read('components/audit-fieldwork/AuditFieldworkTestControls.vue')).not.toContain('getResultColor')
    expect(read('components/audit-fieldwork/AuditFieldworkTestControlsModal.vue')).not.toContain('getResultColor')
    for (const rel of ['pages/dashboard/index.vue', 'pages/risk-profile/risk-control-matrix/index.vue']) {
      expect(read(rel)).not.toMatch(/bg-(emerald|sky|amber)-500 text-white/)
    }
  })

  it('the AI getRiskConfig returns the shared badge class', () => {
    const content = read('composables/useAiAnalytics.ts')
    expect(content).toContain('normalizeRiskLevel(')
    expect(content).toContain('badgeClass')
  })
})

describe('Audit Priority page uses the shared risk level badge', () => {
  const filePath = path.resolve(__dirname, '../../pages/risk-profile/audit-priority/index.vue')
  const content = fs.readFileSync(filePath, 'utf8')

  it('renders both the main table and the reference table through getRiskLevelBadgeClass', () => {
    expect(content).toContain('from \'~/utils/riskLevelBadge\'')
    expect(content).toContain(':class="getRiskLevelBadgeClass(ent.risk_level)"')
    expect(content).toContain(':class="getRiskLevelBadgeClass(row.level)"')
  })

  it('no longer hardcodes hex backgrounds or the semantic UBadge colour mapping', () => {
    expect(content).not.toMatch(/background-color:\s*#/i)
    for (const hex of ['#F44336', '#FF9800', '#FFC107', '#8BC34A', '#4CAF50']) {
      expect(content.toUpperCase()).not.toContain(hex)
    }
    expect(content).not.toContain('getRiskLevelBadgeColor')
  })

  it('keeps reference table cells as real table cells (no flex on td)', () => {
    expect(content).not.toMatch(/<td[^>]*class="[^"]*\bflex\b[^"]*"/)
  })
})
