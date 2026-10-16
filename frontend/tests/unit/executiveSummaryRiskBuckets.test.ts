import { describe, it, expect } from 'vitest'
import {
  RISK_BUCKET_BY_LEVEL,
  RISK_LEVELS,
  emptyRiskBreakdown,
  riskBucketBreakdown,
  riskBucketTotal,
  type RiskBreakdown
} from '~/utils/executiveSummaryCompilation'

describe('Compilation risk buckets (High / Medium / Low)', () => {
  it('maps the five audit-universe levels to the three display buckets', () => {
    expect(RISK_BUCKET_BY_LEVEL).toEqual({
      high: 'high',
      moderateHigh: 'high',
      moderate: 'medium',
      lowModerate: 'low',
      low: 'low'
    })
  })

  it('maps "Moderate to High" to High and "Low to Moderate" to Low', () => {
    expect(RISK_BUCKET_BY_LEVEL.moderateHigh).toBe('high')
    expect(RISK_BUCKET_BY_LEVEL.lowModerate).toBe('low')
    expect(RISK_BUCKET_BY_LEVEL.moderate).toBe('medium')
  })

  it('sums each level into its bucket', () => {
    const breakdown: RiskBreakdown = { high: 1, moderateHigh: 2, moderate: 4, lowModerate: 8, low: 16 }
    expect(riskBucketBreakdown(breakdown)).toEqual({ high: 3, medium: 4, low: 24 })
  })

  it('keeps the total equal to the five-level total', () => {
    const breakdown: RiskBreakdown = { high: 1, moderateHigh: 2, moderate: 4, lowModerate: 8, low: 16 }
    expect(riskBucketTotal(riskBucketBreakdown(breakdown))).toBe(31)
    expect(riskBucketTotal(riskBucketBreakdown(emptyRiskBreakdown()))).toBe(0)
  })

  it('covers every audit-universe level exactly once', () => {
    expect(Object.keys(RISK_BUCKET_BY_LEVEL).sort()).toEqual([...RISK_LEVELS].sort())
  })
})
