import { describe, it, expect } from 'vitest'
import { calculateRiskLevel } from '~/utils/enumConverter'
import { RiskLevel, ImpactLevel, PossibilityLevel } from '~/types/risk'

describe('enumConverter', () => {
  describe('calculateRiskLevel', () => {
    it('should calculate LOW risk level', () => {
      expect(calculateRiskLevel(ImpactLevel.VERY_LOW, PossibilityLevel.VERY_RARE)).toBe(RiskLevel.LOW)
      expect(calculateRiskLevel(ImpactLevel.LOW, PossibilityLevel.VERY_RARE)).toBe(RiskLevel.LOW)
    })

    it('should calculate LOW_MODERATE risk level', () => {
      expect(calculateRiskLevel(ImpactLevel.LOW, PossibilityLevel.RARE)).toBe(RiskLevel.LOW_MODERATE)
      expect(calculateRiskLevel(ImpactLevel.MODERATE, PossibilityLevel.VERY_RARE)).toBe(RiskLevel.LOW_MODERATE)
    })

    it('should calculate MODERATE risk level', () => {
      expect(calculateRiskLevel(ImpactLevel.MODERATE, PossibilityLevel.RARE)).toBe(RiskLevel.MODERATE)
      expect(calculateRiskLevel(ImpactLevel.LOW, PossibilityLevel.LIKELY)).toBe(RiskLevel.MODERATE)
    })

    it('should calculate MODERATE_HIGH risk level', () => {
      expect(calculateRiskLevel(ImpactLevel.HIGH, PossibilityLevel.RARE)).toBe(RiskLevel.MODERATE_HIGH)
      expect(calculateRiskLevel(ImpactLevel.MODERATE, PossibilityLevel.LIKELY)).toBe(RiskLevel.MODERATE_HIGH)
    })

    it('should calculate HIGH risk level', () => {
      expect(calculateRiskLevel(ImpactLevel.VERY_HIGH, PossibilityLevel.LIKELY)).toBe(RiskLevel.HIGH)
      expect(calculateRiskLevel(ImpactLevel.VERY_HIGH, PossibilityLevel.VERY_LIKELY)).toBe(RiskLevel.HIGH)
    })
  })
})
