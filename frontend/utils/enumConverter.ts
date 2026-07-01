import type { SelectOption } from '~/types/common'
import { RiskLevel, ImpactLevel, PossibilityLevel } from '~/types/risk'
import { AuditCategory, AuditStatus, FindingSeverity } from '~/types/audit'

/**
 * Convert enum to select options
 */
export const enumToOptions = <T extends Record<string, string>>(
  enumObj: T,
  labelMap?: Record<string, string>
): SelectOption[] => {
  return Object.values(enumObj).map((value) => ({
    value,
    label: labelMap?.[value] || value.replace(/_/g, ' ').toLowerCase()
  }))
}

/**
 * Get risk level options
 */
export const getRiskLevelOptions = (): SelectOption[] => {
  return enumToOptions(RiskLevel, {
    [RiskLevel.LOW]: 'Low',
    [RiskLevel.LOW_MODERATE]: 'Low to Moderate',
    [RiskLevel.MODERATE]: 'Moderate',
    [RiskLevel.MODERATE_HIGH]: 'Moderate to High',
    [RiskLevel.HIGH]: 'High'
  })
}

/**
 * Get impact level options
 */
export const getImpactLevelOptions = (): SelectOption[] => {
  return [
    { value: ImpactLevel.VERY_LOW, label: 'Very Low' },
    { value: ImpactLevel.LOW, label: 'Low' },
    { value: ImpactLevel.MODERATE, label: 'Moderate' },
    { value: ImpactLevel.HIGH, label: 'High' },
    { value: ImpactLevel.VERY_HIGH, label: 'Very High' }
  ]
}

/**
 * Get possibility level options
 */
export const getPossibilityLevelOptions = (): SelectOption[] => {
  return [
    { value: PossibilityLevel.VERY_RARE, label: 'Very Rare' },
    { value: PossibilityLevel.RARE, label: 'Rare' },
    { value: PossibilityLevel.POSSIBLE, label: 'Possible' },
    { value: PossibilityLevel.LIKELY, label: 'Likely' },
    { value: PossibilityLevel.VERY_LIKELY, label: 'Very Likely' }
  ]
}

/**
 * Get audit type options
 */
export const getAuditCategoryOptions = (): SelectOption[] => {
  return enumToOptions(AuditCategory, {
    [AuditCategory.ASSURANCE]: 'Assurance',
    [AuditCategory.SPECIAL_AUDIT]: 'Special Audit',
    [AuditCategory.SPECIFIC_REASON]: 'Specific Reason',
    [AuditCategory.INVESTIGATION]: 'Investigation'
  })
}

/**
 * Get audit status options
 */
export const getAuditStatusOptions = (): SelectOption[] => {
  return enumToOptions(AuditStatus, {
    [AuditStatus.PLANNED]: 'Planned',
    [AuditStatus.IN_PROGRESS]: 'In Progress',
    [AuditStatus.FIELDWORK]: 'Fieldwork',
    [AuditStatus.REPORTING]: 'Reporting',
    [AuditStatus.COMPLETED]: 'Completed',
    [AuditStatus.CANCELLED]: 'Cancelled'
  })
}

/**
 * Get finding severity options
 */
export const getFindingSeverityOptions = (): SelectOption[] => {
  return enumToOptions(FindingSeverity, {
    [FindingSeverity.LOW]: 'Low',
    [FindingSeverity.MEDIUM]: 'Medium',
    [FindingSeverity.HIGH]: 'High',
    [FindingSeverity.CRITICAL]: 'Critical'
  })
}

/**
 * Calculate risk level from impact and possibility
 */
export const calculateRiskLevel = (
  impact: ImpactLevel,
  possibility: PossibilityLevel
): RiskLevel => {
  const score = Number(impact) * Number(possibility)

  if (score <= 4) return RiskLevel.LOW
  if (score <= 8) return RiskLevel.LOW_MODERATE
  if (score <= 12) return RiskLevel.MODERATE
  if (score <= 16) return RiskLevel.MODERATE_HIGH
  return RiskLevel.HIGH
}
