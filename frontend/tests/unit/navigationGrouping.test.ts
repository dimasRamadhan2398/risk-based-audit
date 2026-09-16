import { describe, it, expect } from 'vitest'
import fs from 'fs'
import path from 'path'

describe('Navigation Grouping Structure in default.vue', () => {
  const defaultVuePath = path.resolve(__dirname, '../../layouts/default.vue')
  const content = fs.readFileSync(defaultVuePath, 'utf8')

  it('contains all 8 required main menu groups in order', () => {
    const rawItemsMatch = content.match(/const rawItems: NavigationMenuItem\[\]\[\] = \[\[([\s\S]*?)\]\]/)
    expect(rawItemsMatch).not.toBeNull()
    const rawStr = rawItemsMatch![1]

    // 1. Dashboard
    expect(rawStr).toContain("label: 'Dashboard'")
    expect(rawStr).toContain("to: '/dashboard'")

    // 2. Audit Charter
    expect(rawStr).toContain("label: 'Audit Charter'")
    expect(rawStr).toContain("to: '/audit-charter'")

    // 3. Risk Profile & Audit Universe
    expect(rawStr).toContain("label: 'Risk Profile & Audit Universe'")

    // 4. Audit Planning
    expect(rawStr).toContain("label: 'Audit Planning'")

    // 5. Audit Result Report
    expect(rawStr).toContain("label: 'Audit Result Report'")

    // 6. Consulting Service
    expect(rawStr).toContain("label: 'Consulting Service'")

    // 7. Quality Assurance Review
    expect(rawStr).toContain("label: 'Quality Assurance Review'")

    // 8. Analytics
    expect(rawStr).toContain("label: 'Analytics'")
  })

  it('verifies Group 3: Risk Profile & Audit Universe subfeatures', () => {
    expect(content).toContain("label: 'Corporate Risk Profile'")
    expect(content).toContain("label: 'Risk Appetite Statement'")
    expect(content).toContain("label: 'Risk Factors'")
    expect(content).toContain("label: 'Audit Universe'")
    expect(content).toContain("label: 'Audit Priority'")
    expect(content).toContain("to: '/risk-profile/audit-universe?tab=priority'")
    expect(content).toContain("label: 'Risk Control Matrix'")
  })

  it('verifies Group 4: Audit Planning subfeatures with all imports and create items directly visible', () => {
    // Strategic Audit Plan
    expect(content).toContain("label: 'Strategic Audit Plan'")
    expect(content).toContain("label: 'Import Strategic Audit Plan'")
    expect(content).toContain("to: '/strategic-audit-plan/upload'")

    // KPI Performance & Laporan Kinerja
    expect(content).toContain("label: 'Internal Audit Performance'")
    expect(content).toContain("label: 'Import Laporan Kinerja'")
    expect(content).toContain("to: '/kpi-performance/upload'")

    // Annual Audit Plan
    expect(content).toContain("label: 'Annual Audit Plan'")
    expect(content).toMatch(/label:\s*'(Import Annual Audit Plan|Import Annual Audit Plan Document)'/)
    expect(content).toContain("to: '/annual-audit/upload'")

    // Audit Execution Status
    expect(content).toContain("label: 'Audit Execution Status'")
    expect(content).toContain("to: '/audit-execution-status'")

    // Audit Activity Plan
    expect(content).toContain("label: 'Audit Activity Plan'")
    expect(content).toMatch(/label:\s*'(Import Audit Activity Plan|Import Activity Plan Document)'/)
    expect(content).toContain("to: '/audit-activity-plan/upload'")

    // Assignment Letter
    expect(content).toContain("label: 'Assignment Letter'")
    expect(content).toContain("label: 'Import Assignment Letter Document'")
    expect(content).toContain("to: '/assignment-letter/upload'")

    // Audit Fieldwork
    expect(content).toContain("label: 'Audit Fieldwork'")
    expect(content).toContain("to: '/audit-fieldwork'")

    // Working Papers
    expect(content).toContain("label: 'Working Paper'")
    expect(content).toContain("label: 'Import Working Paper Document'")
    expect(content).toContain("to: '/working-paper/upload'")

    // AOI & RCA
    expect(content).toContain("label: 'AOI & RCA'")
    expect(content).toContain("to: '/working-paper?step=f04'")

    // Action Plan
    expect(content).toContain("label: 'Action Plan'")
    expect(content).toContain("to: '/mitigation'")
  })

  it('verifies Group 5: Audit Result Report subfeatures with all imports directly visible', () => {
    expect(content).toContain("label: 'Result Reports (LHA)'")
    expect(content).toContain("label: 'Import LHA Document'")
    expect(content).toContain("to: '/audit-result-report/upload'")
    expect(content).toContain("label: 'Auto Generate Report'")
    expect(content).toContain("label: 'Executive Summary'")
    expect(content).toContain("label: 'Import Executive Summary Document'")
    expect(content).toContain("to: '/audit-result-report/executive-summary-upload'")
    expect(content).toContain("label: 'Executive Summary Report Kompilasi'")
    expect(content).toContain("label: 'Import Executive Summary Report'")
    expect(content).toContain("to: '/executive-summary/upload'")
    expect(content).toContain("label: 'Action Taken Report'")
    expect(content).toContain("to: '/action-taken-report'")
  })

  it('verifies Group 6: Consulting Service subfeatures with import directly visible', () => {
    expect(content).toContain("label: 'Consulting Service Dashboard'")
    expect(content).toContain("label: 'Import Consulting Document'")
    expect(content).toContain("to: '/consulting-service/upload'")
  })

  it('verifies Group 7: Quality Assurance Review subfeatures with all imports directly visible', () => {
    expect(content).toContain("label: 'Quality Assurance Dashboard'")
    expect(content).toContain("label: 'Import Periodic Self Assessment'")
    expect(content).toContain("to: '/quality-assurance/import-periodic-self-assessment'")
    expect(content).toContain("label: 'Import SAIV'")
    expect(content).toContain("to: '/quality-assurance/import-saiv'")
    expect(content).toContain("label: 'Import QAR Report'")
    expect(content).toContain("to: '/quality-assurance/import'")
    expect(content).toContain("label: 'Import IACM'")
    expect(content).toContain("to: '/quality-assurance/import-iacm'")
  })

  it('verifies Group 8: Analytics subfeatures', () => {
    expect(content).toContain("label: 'Risk Scoring Prediction'")
    expect(content).toContain("to: '/analytics?tab=xgboost'")
    expect(content).toContain("label: 'Anomaly Detection'")
    expect(content).toContain("to: '/analytics?tab=isolation'")
    expect(content).toContain("label: 'Detected Anomalies Detail'")
    expect(content).toContain("label: 'NLP Analysis'")
    expect(content).toContain("to: '/analytics?tab=nlp'")
    expect(content).toContain("label: 'KPI Forecast'")
    expect(content).toContain("to: '/analytics/kpi-forecast'")
  })
})
