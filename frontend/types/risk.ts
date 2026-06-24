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

export interface RiskAssessment {
  id?: string
  risk_register_id?: string
  year: number
  impact_q1: number
  impact_q2: number
  impact_q3: number
  impact_q4: number
  likelihood_q1: number
  likelihood_q2: number
  likelihood_q3: number
  likelihood_q4: number
}

export interface Risk {
  id: string
  name: string
  category: string
  impact: number // 1-5
  likelihood: number // 1-5
  severity: number // 1-100
  description: string
  branch?: string
  assessments?: RiskAssessment[]
}

export interface RiskForm {
  id?: string
  name: string
  category: string
  impact: number // 1-5
  likelihood: number // 1-5
  severity: number // 1-100
  description: string
  branch?: string
  assessments?: RiskAssessment[]
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
  monitoring?: any[]
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
  monitoring?: any[]
}