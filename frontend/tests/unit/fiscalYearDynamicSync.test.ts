import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import fs from 'fs'
import path from 'path'
import { getFiscalYears, getFiscalYearStrings, useFiscalYear } from '~/composables/useFiscalYear'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { useRCMStore } from '~/stores/rcm'

describe('Dynamic Fiscal Year & Multi-Module Synchronization', () => {
  const rootDir = path.resolve(__dirname, '../../')

  beforeEach(() => {
    setActivePinia(createPinia())
  })

  describe('1. Dynamic Fiscal Year 5-Year Range Formula', () => {
    it('generates 5 years from last year to 3 years ahead for 2026', () => {
      const years = getFiscalYears(2026)
      expect(years).toEqual([2025, 2026, 2027, 2028, 2029])
      expect(years).toHaveLength(5)
    })

    it('generates 5 years from last year to 3 years ahead when rolling over to 2027', () => {
      const years = getFiscalYears(2027)
      expect(years).toEqual([2026, 2027, 2028, 2029, 2030])
      expect(years).toHaveLength(5)
    })

    it('provides string equivalents matching the same range', () => {
      const strings = getFiscalYearStrings(2026)
      expect(strings).toEqual(['2025', '2026', '2027', '2028', '2029'])
    })

    it('composable returns reactive computed fiscal years for current year', () => {
      const { fiscalYears, fiscalYearStrings, currentYear } = useFiscalYear()
      const thisYear = new Date().getFullYear()
      expect(currentYear.value).toBe(thisYear)
      expect(fiscalYears.value).toEqual([
        thisYear - 1,
        thisYear,
        thisYear + 1,
        thisYear + 2,
        thisYear + 3
      ])
      expect(fiscalYearStrings.value).toEqual([
        String(thisYear - 1),
        String(thisYear),
        String(thisYear + 1),
        String(thisYear + 2),
        String(thisYear + 3)
      ])
    })
  })

  describe('2. State Synchronization across Risk Profile Modules', () => {
    it('synchronizes selected fiscal year between riskProfileStore and rcmStore', () => {
      const riskStore = useRiskProfileStore()
      const rcmStore = useRCMStore()

      // Both should match initially
      expect(rcmStore.selectedYear).toBe(riskStore.selectedYear)

      // Updating riskStore updates rcmStore
      riskStore.selectedYear = 2028
      expect(rcmStore.selectedYear).toBe(2028)

      // Updating rcmStore updates riskStore
      rcmStore.selectedYear = 2029
      expect(riskStore.selectedYear).toBe(2029)
    })

    it('verifies RiskHeatMap uses dynamic fiscalYears items', () => {
      const content = fs.readFileSync(path.join(rootDir, 'components/risk-profile/RiskHeatMap.vue'), 'utf8')
      expect(content).toContain(':items="fiscalYears"')
      expect(content).not.toContain(':items="[2025, 2026, 2027]"')
    })

    it('verifies Risk Factors uses dynamic fiscalYears and synchronizes selectedYear', () => {
      const content = fs.readFileSync(path.join(rootDir, 'pages/risk-profile/risk-factors/index.vue'), 'utf8')
      expect(content).toContain(':items="fiscalYears"')
      expect(content).toContain('selectedFiscalYear')
      expect(content).not.toContain(':items="[2025, 2026, 2027, 2028]"')
    })

    it('verifies Audit Universe uses dynamic fiscalYears and synchronizes selectedYear', () => {
      const content = fs.readFileSync(path.join(rootDir, 'pages/risk-profile/audit-universe/index.vue'), 'utf8')
      expect(content).toContain(':items="fiscalYears"')
      expect(content).toContain('selectedFiscalYear')
      expect(content).not.toContain(':items="[2025, 2026, 2027, 2028]"')
    })

    it('verifies Audit Priority uses dynamic fiscalYears select instead of number input', () => {
      const content = fs.readFileSync(path.join(rootDir, 'pages/risk-profile/audit-priority/index.vue'), 'utf8')
      expect(content).toContain(':items="fiscalYears"')
      expect(content).toContain('selectedFiscalYear')
      expect(content).not.toContain(':min="2020"')
    })

    it('verifies Risk Control Matrix yearOptions dynamically derives from fiscalYears', () => {
      const content = fs.readFileSync(path.join(rootDir, 'pages/risk-profile/risk-control-matrix/index.vue'), 'utf8')
      expect(content).toContain('fiscalYears')
      expect(content).toContain('yearOptions')
      expect(content).not.toContain("{ label: 'Tahun 2024', value: 2024 }")
    })

    it('empties risk profile and RCM data for the 3 future years until inputted by user', async () => {
      const riskStore = useRiskProfileStore()
      const rcmStore = useRCMStore()
      const currentYear = new Date().getFullYear()

      // Current year (e.g. 2026) has initial risks
      riskStore.selectedYear = currentYear
      expect(riskStore.risks.length).toBeGreaterThan(0)

      // 3 future years (2027, 2028, 2029) are empty by default because user must input them
      riskStore.selectedYear = currentYear + 1
      expect(riskStore.risks).toEqual([])
      expect(rcmStore.filteredRCMList).toEqual([])
      expect(rcmStore.synchronizedRiskCounts).toEqual({ inherent: 0, residual: 0 })

      riskStore.selectedYear = currentYear + 2
      expect(riskStore.risks).toEqual([])
      expect(rcmStore.filteredRCMList).toEqual([])
      expect(rcmStore.synchronizedRiskCounts).toEqual({ inherent: 0, residual: 0 })

      riskStore.selectedYear = currentYear + 3
      expect(riskStore.risks).toEqual([])
      expect(rcmStore.filteredRCMList).toEqual([])
      expect(rcmStore.synchronizedRiskCounts).toEqual({ inherent: 0, residual: 0 })

      // Adding a risk for future year 1 (e.g. 2027) populates only that year
      await riskStore.addRisk({
        name: 'New Future Risk 2027',
        category: 'Strategic',
        impact: 4,
        likelihood: 4,
        severity: 80,
        branch: 'Head Office',
        assessments: [
          {
            year: currentYear + 1,
            impact_q1: 4, impact_q2: 4, impact_q3: 4, impact_q4: 4,
            likelihood_q1: 4, likelihood_q2: 4, likelihood_q3: 4, likelihood_q4: 4,
            risk_level_q1: 'High', risk_level_q2: 'High', risk_level_q3: 'High', risk_level_q4: 'High'
          }
        ]
      })

      riskStore.selectedYear = currentYear + 1
      expect(riskStore.risks.length).toBe(1)
      expect(riskStore.risks[0].name).toBe('New Future Risk 2027')

      // Other future years (2028, 2029) still remain empty
      riskStore.selectedYear = currentYear + 2
      expect(riskStore.risks).toEqual([])

      riskStore.selectedYear = currentYear + 3
      expect(riskStore.risks).toEqual([])
    })
  })

  describe('3. Dynamic Fiscal Year Format across AuditSphere Features', () => {
    it('Strategic Audit Plan uses dynamic fiscal years in store and table filter', () => {
      const storeContent = fs.readFileSync(path.join(rootDir, 'stores/strategic-audit-plan.ts'), 'utf8')
      const tableContent = fs.readFileSync(path.join(rootDir, 'components/strategic-audit-plan/StrategicPlanTable.vue'), 'utf8')
      const vmgContent = fs.readFileSync(path.join(rootDir, 'stores/vision-mission-goals.ts'), 'utf8')

      expect(storeContent).toContain('getFiscalYears')
      expect(tableContent).toContain('getFiscalYears')
      expect(tableContent).not.toContain("{ label: '2024', value: '2024' }")
      expect(vmgContent).toContain('getFiscalYears')
    })

    it('Internal Audit Performance uses dynamic fiscal years in index and upload', () => {
      const indexContent = fs.readFileSync(path.join(rootDir, 'pages/kpi-performance/index.vue'), 'utf8')
      const uploadContent = fs.readFileSync(path.join(rootDir, 'pages/kpi-performance/upload.vue'), 'utf8')

      expect(indexContent).toContain('getFiscalYearStrings()')
      expect(indexContent).not.toContain("['2024', '2025', '2026', '2027', '2028', '2029', '2030']")
      expect(uploadContent).toContain('getFiscalYears()')
      expect(uploadContent).not.toContain('[2024, 2025, 2026, 2027, 2028]')
    })

    it('Annual Audit Plan store uses dynamic fiscal years for yearOptions', () => {
      const content = fs.readFileSync(path.join(rootDir, 'stores/annual-audit.ts'), 'utf8')
      expect(content).toContain('getFiscalYearStrings()')
      expect(content).not.toContain("['2026', '2027', '2028', '2029', '2030']")
    })

    it('Audit Activity Plan form uses dynamic fiscal years for yearOptions', () => {
      const content = fs.readFileSync(path.join(rootDir, 'components/audit-activity-plan/ActivityPlanForm.vue'), 'utf8')
      expect(content).toContain('getFiscalYearStrings()')
      expect(content).not.toContain('currentYear + 5')
    })

    it('Assignment Letter uses dynamic fiscal years in store options and form select', () => {
      const storeContent = fs.readFileSync(path.join(rootDir, 'stores/assignment-letter.ts'), 'utf8')
      const formContent = fs.readFileSync(path.join(rootDir, 'components/assignment-letter/AssignmentLetterForm.vue'), 'utf8')

      expect(storeContent).toContain('yearOptions: getFiscalYearStrings()')
      expect(formContent).toContain('USelectMenu')
      expect(formContent).toContain('store.form.auditYear')
      expect(formContent).not.toContain('type="date"')
    })

    it('Consulting Service uses dynamic fiscal years in store', () => {
      const content = fs.readFileSync(path.join(rootDir, 'stores/consulting-service.ts'), 'utf8')
      expect(content).toContain('getFiscalYearStrings()')
      expect(content).not.toContain("['2024', '2025', '2026', '2027']")
    })

    it('Quality Assurance Review uses dynamic fiscal years in store periods', () => {
      const content = fs.readFileSync(path.join(rootDir, 'stores/quality-assurance.ts'), 'utf8')
      expect(content).toContain('getFiscalYearStrings()')
      expect(content).not.toContain("['2026', '2025', '2024', '2023']")
    })
  })
})
