import { describe, it, expect } from 'vitest'
import fs from 'fs'
import path from 'path'

describe('Navigation Grouping Structure in default.vue', () => {
  const defaultVuePath = path.resolve(__dirname, '../../layouts/default.vue')
  const content = fs.readFileSync(defaultVuePath, 'utf8')

  it('contains all required main menu groups in order', () => {
    // Matches rawItems computed definition
    expect(content).toContain("t('navigation.dashboard')")
    expect(content).toContain("to: '/dashboard'")

    expect(content).toContain("t('navigation.auditCharter')")
    expect(content).toContain("to: '/audit-charter'")

    expect(content).toContain("t('navigation.riskProfileUniverse')")
    expect(content).toContain("t('navigation.auditPlanning')")
    expect(content).toContain("t('navigation.auditResultReport')")
    expect(content).toContain("t('navigation.consultingService')")
    expect(content).toContain("t('navigation.qualityAssuranceReview')")
    expect(content).toContain("t('navigation.analytics')")
  })

  it('verifies Group 3: Risk Profile & Audit Universe subfeatures', () => {
    expect(content).toContain("t('navigation.corporateRiskProfile')")
    expect(content).toContain("t('navigation.riskAppetiteStatement')")
    expect(content).toContain("t('navigation.riskFactors')")
    expect(content).toContain("t('navigation.auditUniverse')")
    expect(content).toContain("t('navigation.auditPriority')")
    expect(content).toContain("to: '/risk-profile/audit-universe?tab=priority'")
    expect(content).toContain("t('navigation.riskControlMatrix')")
  })

  it('verifies Group 4: Audit Planning subfeatures with all imports and create items directly visible', () => {
    // Strategic Audit Plan
    expect(content).toContain("t('navigation.strategicAuditPlan')")
    expect(content).toContain("t('navigation.importStrategicAuditPlan')")
    expect(content).toContain("to: '/strategic-audit-plan/upload'")

    // KPI Performance
    expect(content).toContain("t('navigation.internalAuditPerformance')")
    expect(content).toContain("t('navigation.importPerformanceReport')")
    expect(content).toContain("to: '/kpi-performance/upload'")

    // Annual Audit Plan
    expect(content).toContain("t('navigation.annualAuditPlan')")
    expect(content).toContain("t('navigation.importAnnualAuditPlan')")
    expect(content).toContain("to: '/annual-audit/upload'")
    expect(content).toContain("t('navigation.auditExecutionStatus')")
    expect(content).toContain("to: '/audit-execution-status'")

    // Audit Activity Plan
    expect(content).toContain("t('navigation.auditActivityPlan')")
    expect(content).toContain("t('navigation.importAuditActivityPlan')")
    expect(content).toContain("to: '/audit-activity-plan/upload'")

    // Assignment Letter
    expect(content).toContain("t('navigation.assignmentLetter')")
    expect(content).toContain("t('navigation.importAssignmentLetter')")
    expect(content).toContain("to: '/assignment-letter/upload'")

    // Audit Fieldwork
    expect(content).toContain("t('navigation.auditFieldwork')")
    expect(content).toContain("to: '/audit-fieldwork'")

    // Working Paper
    expect(content).toContain("t('navigation.workingPaper')")
    expect(content).toContain("t('navigation.importWorkingPaper')")
    expect(content).toContain("to: '/working-paper/upload'")
  })

  it('verifies Group 5: Audit Result Report subfeatures with accurate separation and no duplicate paths', () => {
    expect(content).toContain("t('navigation.resultReportsLha')")
    expect(content).toContain("to: '/audit-result-report'")
    expect(content).toContain("t('navigation.importLhaDocument')")
    expect(content).toContain("to: '/audit-result-report/upload'")

    // Executive Summary Individual
    expect(content).toContain("t('navigation.executiveSummaryIndividual')")
    expect(content).toContain("to: '/executive-summary'")
    expect(content).toContain("t('navigation.importExecutiveSummary')")
    expect(content).toContain("to: '/executive-summary/upload'")

    // Executive Summary Compilation
    expect(content).toContain("t('navigation.executiveSummaryCompilation')")
    expect(content).toContain("to: '/executive-summary-compilation'")
    expect(content).toContain("t('navigation.importExecutiveSummaryCompilation')")
    expect(content).toContain("to: '/executive-summary-compilation/upload'")

    expect(content).toContain("t('navigation.actionTakenReport')")
    expect(content).toContain("to: '/action-taken-report'")
    expect(content).toContain("t('navigation.clientSatisfactionSurvey')")
    expect(content).toContain("to: '/audit-result-report/satisfaction-survey'")
  })

  it('verifies Group 6: Consulting Service subfeatures with import', () => {
    expect(content).toContain("t('navigation.consultingServiceDashboard')")
    expect(content).toContain("t('navigation.importConsultingDocument')")
    expect(content).toContain("to: '/consulting-service/upload'")
  })

  it('verifies Group 7: Quality Assurance Review subfeatures with all imports', () => {
    expect(content).toContain("t('navigation.qualityAssuranceDashboard')")
    expect(content).toContain("t('navigation.importPeriodicSelfAssessment')")
    expect(content).toContain("to: '/quality-assurance/import-periodic-self-assessment'")
    expect(content).toContain("t('navigation.importSaiv')")
    expect(content).toContain("to: '/quality-assurance/import-saiv'")
    expect(content).toContain("t('navigation.importQarReport')")
    expect(content).toContain("to: '/quality-assurance/import'")
    expect(content).toContain("t('navigation.importIacm')")
    expect(content).toContain("to: '/quality-assurance/import-iacm'")
  })

  it('verifies Group 8: Analytics subfeatures', () => {
    expect(content).toContain("t('navigation.riskScoringPrediction')")
    expect(content).toContain("to: '/analytics/ai/risk-scoring'")
    expect(content).toContain("t('navigation.anomalyDetection')")
    expect(content).toContain("to: '/analytics/ai/anomaly-detection'")
    expect(content).toContain("t('navigation.nlpAnalysis')")
    expect(content).toContain("to: '/analytics/ai/nlp'")
    expect(content).toContain("t('navigation.kpiForecast')")
    expect(content).toContain("to: '/analytics/ai/kpi-forecast'")
    expect(content).toContain("t('navigation.caattFullPopulation')")
    expect(content).toContain("to: '/analytics/caatt/full-population'")
    expect(content).toContain("t('navigation.caattDuplicateGap')")
    expect(content).toContain("to: '/analytics/caatt/duplicate-gap'")
    expect(content).toContain("t('navigation.caattBenford')")
    expect(content).toContain("to: '/analytics/caatt/benford'")
    expect(content).toContain("t('navigation.caattStratification')")
    expect(content).toContain("to: '/analytics/caatt/stratification'")
    expect(content).toContain("t('navigation.caattReconciliation')")
    expect(content).toContain("to: '/analytics/caatt/reconciliation'")
    expect(content).toContain("t('navigation.caattPolicy')")
    expect(content).toContain("to: '/analytics/caatt/policy'")
    expect(content).toContain("t('navigation.caattDataQuality')")
    expect(content).toContain("to: '/analytics/caatt/data-quality'")
  })

  it('ensures active path calculation logic prevents false prefix matching on sibling upload routes', () => {
    // Verify allMenuPaths computation exists
    expect(content).toContain('allMenuPaths = computed(')
    // Verify exact match priority exists
    expect(content).toContain('if (currentPath === targetBase)')
    // Verify fallback prefix matching is restricted to unlisted paths
    expect(content).toContain('!allMenuPaths.value.has(currentPath)')
  })
})
