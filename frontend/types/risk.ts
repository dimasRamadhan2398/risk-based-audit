export enum RiskLevel {
  LOW = 'low',
  LOW_MODERATE = 'low-moderate',
  MODERATE = 'moderate',
  MODERATE_HIGH = 'moderate-high',
  HIGH = 'high'
}

export enum RiskTaxonomy {
  Strategic = 'Strategic',
  Operational = 'Operational',
  Financial = 'Financial',
  IT = 'IT'

}

export interface Risk {
  id: string
  risk_name: string
  category: string
  impact: number // 1-5
  likelihood: number // 1-5
  severity: number // 1-100
  description: string
  department?: string
}

export interface RiskForm {
  id: string
  risk_name: string
  category: string
  impact: number // 1-5
  likelihood: number // 1-5
  severity: number // 1-100
  description: string
  department?: string
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

export interface RiskMitigation {
  id: string
  riskId: string
  riskEvent: string
  mitigationPlan: string
  supervisor: string
  pic: string
  unitInCharge: string
  start_date: string
  end_date: string
  notes?: string
}

export interface RiskMitigationForm {
  riskEvent: string
  mitigationPlan: string
  supervisor: string
  pic: string
  unitInCharge: string
  start_date: string
  end_date: string
  notes: string
}