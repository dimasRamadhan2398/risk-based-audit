// Risk Management Types

export enum RiskLevel {
  LOW = 'low',
  LOW_TO_MODERATE = 'low_to_moderate',
  MODERATE = 'moderate',
  MODERATE_TO_HIGH = 'moderate_to_high',
  HIGH = 'high'
}

export enum RiskTaxonomy { 
  Strategic = 'Strategic',
  Operational = 'Operational',
  Financial = 'Financial',
  IT = 'IT'

}
export enum ImpactLevel {
  VERY_LOW = 1,
  LOW = 2,
  MODERATE = 3,
  HIGH = 4,
  VERY_HIGH = 5
}

export enum PossibilityLevel {
  VERY_RARE = 1,
  RARE = 2,
  POSSIBLE = 3,
  LIKELY = 4,
  VERY_LIKELY = 5
}

export interface RiskItem {
  id: string
  title: string
  description: string
  impactLevel: ImpactLevel
  possibilityLevel: PossibilityLevel
  riskLevel: RiskLevel
  quarter: number
  year: number
  department?: string
  createdAt: string
  updatedAt: string
}

export interface RiskProfile {
  id: string
  organizationId?: string
  departmentId?: string
  year: number
  quarter: number
  risks: RiskProfileItem[]
  createdAt: string
  updatedAt: string
}

export interface RiskProfileItem {
  riskId: string
  index: number
  impactLevel: ImpactLevel
  possibilityLevel: PossibilityLevel
  riskLevel: RiskLevel
}

export interface RiskMitigation {
  id: string
  riskId: string
  title: string
  description: string
  status: MitigationStatus
  responsiblePerson: string
  targetDate: string
  completionDate?: string
  createdAt: string
  updatedAt: string
}

export enum MitigationStatus {
  PLANNED = 'planned',
  IN_PROGRESS = 'in_progress',
  COMPLETED = 'completed',
  OVERDUE = 'overdue'
}
