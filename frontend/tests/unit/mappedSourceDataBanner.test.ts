// @ts-nocheck
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import {
  resolveBannerModuleInfo,
  evaluateTableModuleCompatibility,
  buildStrategicPlanCandidatesFromTable,
  computeStrategicKpiMetrics,
  buildAnnualPlanDraftFromTable,
  buildRiskProfilePayloadsFromRows,
  buildAuditUniverseCandidatesFromTable,
  buildRealTableAnomalies,
  calculateQuartersFromMonths,
  isRowAnomalousOrFlagged,
  type MappedRecordInput,
} from '~/utils/mappedFeatureAdapters'
import { STRATEGIC_PLAN_CATEGORIES } from '~/utils/strategicPlanPayload'

function matchMappingToModule(mapping: { targetScope: string; targetModule: string; tableName: string }, moduleInfo: any) {
  if (!moduleInfo) return false
  const norm = (str: string) => (str || '').toLowerCase().replace(/[\s_-]+/g, '')
  const mapModNorm = norm(mapping.targetModule)
  const curModNorm = norm(moduleInfo.module)
  const curSlugNorm = norm(moduleInfo.slug)

  const scopeMatches = mapping.targetScope === moduleInfo.scope
  const moduleMatches = mapModNorm.includes(curModNorm) || curModNorm.includes(mapModNorm) || mapModNorm.includes(curSlugNorm)

  return moduleMatches || (scopeMatches && !mapping.targetModule)
}

describe('MappedSourceDataBanner & mappedFeatureAdapters Logic', () => {
  it('accurately resolves routes including Strategic Audit Plan and Annual Audit Plan', () => {
    const riskInfo = resolveBannerModuleInfo('/risk-profile')
    expect(riskInfo).not.toBeNull()
    expect(riskInfo?.module).toBe('Corporate Risk Profile')
    expect(riskInfo?.scope).toBe('risk_management')
    expect(riskInfo?.actionType).toBe('risk_profile')

    const strategicInfo = resolveBannerModuleInfo('/strategic-audit-plan')
    expect(strategicInfo).not.toBeNull()
    expect(strategicInfo?.module).toBe('Strategic Audit Plan')
    expect(strategicInfo?.scope).toBe('audit_features')
    expect(strategicInfo?.actionType).toBe('strategic_plan')
    expect(strategicInfo?.supported).toBe(true)

    const annualInfo = resolveBannerModuleInfo('/annual-audit')
    expect(annualInfo).not.toBeNull()
    expect(annualInfo?.module).toBe('Annual Audit Plan')
    expect(annualInfo?.scope).toBe('audit_features')
    expect(annualInfo?.actionType).toBe('annual_plan')
    expect(annualInfo?.supported).toBe(true)

    const anomalyInfo = resolveBannerModuleInfo('/analytics/ai/anomaly-detection')
    expect(anomalyInfo).not.toBeNull()
    expect(anomalyInfo?.module).toBe('AI: Anomaly Detection')
    expect(anomalyInfo?.actionType).toBe('ai_anomaly')

    const entityInfo = resolveBannerModuleInfo('/analytics/ai/risk-scoring')
    expect(entityInfo).not.toBeNull()
    expect(entityInfo?.module).toBe('AI: Entity Prediction')
    expect(entityInfo?.actionType).toBe('ai_entity_prediction')
  })

  it('matches user mapping of corporate_transactions to Strategic Audit Plan and Annual Audit Plan', () => {
    const strategicInfo = resolveBannerModuleInfo('/strategic-audit-plan')
    expect(
      matchMappingToModule(
        {
          tableName: 'corporate_transactions',
          targetScope: 'audit_features',
          targetModule: 'Strategic Audit Plan',
        },
        strategicInfo
      )
    ).toBe(true)

    const annualInfo = resolveBannerModuleInfo('/annual-audit')
    expect(
      matchMappingToModule(
        {
          tableName: 'corporate_transactions',
          targetScope: 'audit_features',
          targetModule: 'Annual Audit Plan',
        },
        annualInfo
      )
    ).toBe(true)
  })

  it('evaluates table-to-module compatibility across all 6 connected tables', () => {
    const transactionalTables: MappedRecordInput[] = [
      {
        connection_id: 'conn-1',
        connection_name: 'Core Banking DB',
        table_name: 'corporate_transactions',
        target_scope: 'audit_features',
        target_module: 'Strategic Audit Plan',
        audit_field: 'trx_id',
        total_rows: 3244,
        sample_records: [
          {
            trx_id: 2781,
            trx_ref_number: 'NSA-AFT8-00001',
            branch_name: 'KC Jakarta Sudirman',
            amount: 485000000,
            authorization_status: 'PENDING_OVERRIDE',
            is_suspicious: true,
          },
        ],
      },
      {
        connection_id: 'conn-1',
        connection_name: 'Core Banking DB',
        table_name: 'credit_loan_portfolio',
        target_scope: 'audit_features',
        target_module: 'Strategic Audit Plan',
        audit_field: 'loan_id',
        total_rows: 220,
        sample_records: [
          {
            loan_id: 201,
            loan_number: 'CR-OVR-9901',
            branch_name: 'KC Jakarta Sudirman',
            plafond: 9500000000,
            collectability: 4,
            days_past_due: 145,
          },
        ],
      },
      {
        connection_id: 'conn-1',
        connection_name: 'Core Banking DB',
        table_name: 'gl_journal_entries',
        target_scope: 'audit_features',
        target_module: 'Strategic Audit Plan',
        audit_field: 'gl_id',
        total_rows: 420,
        sample_records: [
          {
            gl_id: 415,
            gl_account_code: '6010',
            voucher_number: null,
            debit_amount: 125000000,
            branch_name: 'KC Medan Putri Hijau',
          },
        ],
      },
      {
        connection_id: 'conn-1',
        connection_name: 'Core Banking DB',
        table_name: 'vendor_contracts',
        target_scope: 'audit_features',
        target_module: 'Strategic Audit Plan',
        audit_field: 'contract_id',
        total_rows: 150,
        sample_records: [
          {
            contract_id: 145,
            contract_no: 'VND-SPL-8801',
            vendor_name: 'PT Mega Solusi Digital',
            branch_name: 'KC Jakarta Sudirman',
            contract_value: 495000000,
            approval_status: 'DIRECT_APPOINTMENT_OVERRIDE',
            is_flagged: true,
          },
        ],
      },
      {
        connection_id: 'conn-1',
        connection_name: 'Core Banking DB',
        table_name: 'audit_risk_incidents',
        target_scope: 'audit_features',
        target_module: 'Strategic Audit Plan',
        audit_field: 'incident_id',
        total_rows: 40,
        sample_records: [
          {
            incident_id: 35,
            incident_code: 'INC-2026-0035',
            entity_unit: 'KC Jakarta Sudirman',
            severity_score: 92,
            exposure_value: 650000000,
          },
        ],
      },
    ]

    for (const rec of transactionalTables) {
      expect(evaluateTableModuleCompatibility(rec, 'strategic_plan').compatible).toBe(true)
      expect(evaluateTableModuleCompatibility(rec, 'annual_plan').compatible).toBe(true)
      expect(evaluateTableModuleCompatibility(rec, 'ai_anomaly').compatible).toBe(true)
    }

    const branchesRecord: MappedRecordInput = {
      connection_id: 'conn-1',
      connection_name: 'Core Banking DB',
      table_name: 'branches',
      target_scope: 'audit_features',
      target_module: 'Strategic Audit Plan',
      audit_field: 'branch_id',
      total_rows: 8,
      profile: {
        has_numeric_measure: false,
        dimension_column: 'branch_name',
        amount_column: null,
        columns: ['branch_id', 'branch_code', 'branch_name', 'region', 'city', 'created_at'],
      },
      sample_records: [
        {
          branch_id: 1,
          branch_code: 'BR-JKT-01',
          branch_name: 'KC Jakarta Sudirman',
          region: 'Region 1 - Jakarta',
          city: 'Jakarta Selatan',
        },
      ],
    }

    // Master reference table `branches` is rejected for Strategic Plan and AI Anomaly Detection
    const branchesStrategicCompat = evaluateTableModuleCompatibility(branchesRecord, 'strategic_plan')
    expect(branchesStrategicCompat.compatible).toBe(false)
    expect(branchesStrategicCompat.reason).toContain('tabel referensi master')

    const branchesAnomalyCompat = evaluateTableModuleCompatibility(branchesRecord, 'ai_anomaly')
    expect(branchesAnomalyCompat.compatible).toBe(false)

    // Master reference table `branches` IS allowed for Audit Universe and Annual Audit Plan (dimension-based)
    expect(evaluateTableModuleCompatibility(branchesRecord, 'audit_universe').compatible).toBe(true)
    expect(evaluateTableModuleCompatibility(branchesRecord, 'annual_plan').compatible).toBe(true)
  })

  it('derives valid Strategic Audit Plan KPI candidates from mapped tables', () => {
    const corpRecord: MappedRecordInput = {
      connection_id: 'conn-1',
      connection_name: 'Core Banking DB',
      table_name: 'corporate_transactions',
      target_scope: 'audit_features',
      target_module: 'Strategic Audit Plan',
      audit_field: 'trx_id',
      total_rows: 3244,
      profile: {
        total_rows: 3244,
        flagged_rows: 464,
        has_numeric_measure: true,
        metrics: {
          suspicious_count: 464,
          unauthorized_count: 150,
        },
      },
      sample_records: [
        { trx_id: 2781, trx_ref_number: 'NSA-AFT8-00001', amount: 485000000, is_suspicious: true },
      ],
    }

    const candidates = buildStrategicPlanCandidatesFromTable(corpRecord, 2026)
    expect(candidates.length).toBeGreaterThanOrEqual(2)
    for (const c of candidates) {
      expect(STRATEGIC_PLAN_CATEGORIES).toContain(c.category)
      expect(Number.isFinite(parseFloat(c.actual))).toBe(true)
      expect(Number.isFinite(parseFloat(c.target))).toBe(true)
      const calc = computeStrategicKpiMetrics(c.actual, c.target, c.hibHig)
      expect(calc.calculation).toMatch(/%$/)
      expect(['Good', 'Moderate', 'Poor']).toContain(calc.status)
    }
  })

  it('builds an Annual Audit Plan draft grouped by branch/entity dimension with valid schema fields', () => {
    const corpRecord: MappedRecordInput = {
      connection_id: 'conn-1',
      connection_name: 'Core Banking DB',
      table_name: 'corporate_transactions',
      target_scope: 'audit_features',
      target_module: 'Annual Audit Plan',
      audit_field: 'trx_id',
      total_rows: 3244,
      profile: {
        total_rows: 3244,
        flagged_rows: 464,
        dimension_column: 'branch_name',
        dimension_breakdown: [
          {
            name: 'KC Jakarta Sudirman',
            row_count: 850,
            flagged_count: 120,
            total_amount: 14500000000,
            max_amount: 495000000,
          },
          {
            name: 'KC Surabaya Gubeng',
            row_count: 620,
            flagged_count: 80,
            total_amount: 9200000000,
            max_amount: 480000000,
          },
        ],
      },
      sample_records: [],
    }

    const draft = buildAnnualPlanDraftFromTable(corpRecord, {
      departments: [
        { id: 'dept-ops', name: 'Operational & Branch Banking' },
        { id: 'dept-fin', name: 'Finance & Accounting' },
      ],
      corporateUniverse: [
        {
          id: 'corp-root',
          name: 'PT AIFL Indonesia (Holding)',
          children: [
            { id: 'au-jkt', name: 'KC Jakarta Sudirman', department_id: 'dept-ops' },
          ],
        },
      ],
      supervisors: ['Budi Santoso'],
      year: 2026,
    })

    expect(draft.title).toContain('corporate_transactions')
    expect(draft.year).toBe('2026')
    expect(draft.selectedMonths).toEqual([0, 1, 2])
    expect(calculateQuartersFromMonths(draft.selectedMonths)).toEqual(['Q1'])
    expect(calculateQuartersFromMonths([0, 4, 8, 11])).toEqual(['Q1', 'Q2', 'Q3', 'Q4'])
    expect(draft.activities).toHaveLength(2)
    expect(draft.activities[0].dimensionName).toBe('KC Jakarta Sudirman')
    expect(draft.activities[0].auditUniverseId).toBe('au-jkt')
    expect(draft.activities[1].dimensionName).toBe('KC Surabaya Gubeng')
    expect(draft.activities[1].auditUniverseId).toBe('')
    expect(draft.activities[0].departmentId).toBe('dept-ops')
    expect(draft.activities[0].supervisorName).toBe('Budi Santoso')
    expect(draft.activities[0].auditorCount).toBeGreaterThanOrEqual(2)
    expect(draft.activities[0].totalMandays).toBeGreaterThanOrEqual(15)
  })

  it('keeps Annual Audit Plan confirmation modal aligned with PKAT form controls and free of Reka UI empty-string USelect traps', () => {
    const src = readFileSync(
      resolve(__dirname, '../../components/common/MappedSourceDataBanner.vue'),
      'utf8'
    )
    expect(src).not.toContain('Kuartal / Timeline')
    expect(src).not.toContain("value: ''")
    expect(src).toContain('annualStore.yearOptions')
    expect(src).toContain('Object.values(AuditCategory)')
    expect(src).toContain('buildAuditableEntityOptions(universeStore.yearlyUniverse, universeStore.corporateUniverse)')
    expect(src).toContain('closeAnnualModal')
    expect(src).toContain('unlockBodyPointerEvents')
    expect(src).toContain('buildRiskProfilePayloadsFromRows')
    expect(src).toContain('buildAuditUniverseCandidatesFromTable')
    expect(src).not.toContain('Math.min(5, records.length)')
  })

  it('builds one distinct Corporate Risk Profile payload per row without artificial limits', () => {
    const rows = Array.from({ length: 12 }, (_, idx) => ({
      trx_id: 1000 + idx,
      trx_ref_number: `NSA-AFT8-000${idx + 1}`,
      branch_name: idx % 2 === 0 ? 'KC Jakarta Sudirman' : 'KC Surabaya Gubeng',
      amount: 100000000 * (idx + 1),
      authorization_status: idx < 4 ? 'PENDING_OVERRIDE' : 'AUTHORIZED',
      is_suspicious: idx < 4,
      notes: idx < 4 ? 'Transaksi di luar jam operasional tanpa dual-control' : 'Transaksi normal',
    }))

    const corpRecord: MappedRecordInput = {
      connection_id: 'conn-1',
      connection_name: 'Core Banking DB',
      table_name: 'corporate_transactions',
      target_scope: 'risk_management',
      target_module: 'Corporate Risk Profile',
      audit_field: 'trx_id',
      total_rows: 3244,
      sample_records: rows,
    }

    const payloads = buildRiskProfilePayloadsFromRows(corpRecord, {
      year: 2026,
      locations: [
        { id: 'loc-jkt', name: 'KC Jakarta Sudirman' },
        { id: 'loc-sby', name: 'KC Surabaya Gubeng' },
      ],
    })

    expect(payloads).toHaveLength(12)
    expect(payloads[0].category).toBe('Financial')
    expect(payloads[0].refCode).toContain('NSA-AFT8-0001')
    expect(payloads[0].branch).toBe('KC Jakarta Sudirman')
    expect(payloads[0].location_id).toBe('loc-jkt')
    expect(payloads[0].assessments).toHaveLength(1)
    expect(payloads[0].assessments[0].year).toBe(2026)
    expect(payloads[0].likelihood).toBeGreaterThanOrEqual(1)
    expect(payloads[0].likelihood).toBeLessThanOrEqual(5)
    expect(payloads[0].impact).toBeGreaterThanOrEqual(1)
    expect(payloads[0].impact).toBeLessThanOrEqual(5)
    expect(payloads[1].branch).toBe('KC Surabaya Gubeng')
    expect(payloads[1].location_id).toBe('loc-sby')
  })

  it('builds Audit Universe candidates per row without the old 6-row limit', () => {
    const rows = Array.from({ length: 10 }, (_, idx) => ({
      trx_id: 2000 + idx,
      trx_ref_number: `TRX-UNIV-${idx + 1}`,
      branch_name: `KC Cabang ${idx + 1}`,
      amount: 50000000 * (idx + 1),
      is_suspicious: idx % 3 === 0,
    }))

    const corpRecord: MappedRecordInput = {
      connection_id: 'conn-1',
      connection_name: 'Core Banking DB',
      table_name: 'corporate_transactions',
      target_scope: 'audit_features',
      target_module: 'Audit Universe',
      audit_field: 'trx_id',
      total_rows: 3244,
      sample_records: rows,
    }

    const candidates = buildAuditUniverseCandidatesFromTable(corpRecord)
    expect(candidates).toHaveLength(10)
    expect(candidates[0].name).toContain('TRX-UNIV-1')
    expect(candidates[9].name).toContain('TRX-UNIV-10')
  })

  it('builds Annual Audit Plan activities per row without the old 5-activity limit', () => {
    const rows = Array.from({ length: 8 }, (_, idx) => ({
      trx_id: 3000 + idx,
      trx_ref_number: `TRX-PKAT-${idx + 1}`,
      branch_name: 'KC Jakarta Sudirman',
      amount: 250000000 * (idx + 1),
      authorization_status: 'PENDING_OVERRIDE',
      is_suspicious: true,
    }))

    const corpRecord: MappedRecordInput = {
      connection_id: 'conn-1',
      connection_name: 'Core Banking DB',
      table_name: 'corporate_transactions',
      target_scope: 'audit_features',
      target_module: 'Annual Audit Plan',
      audit_field: 'trx_id',
      total_rows: 3244,
      sample_records: rows,
    }

    const draft = buildAnnualPlanDraftFromTable(corpRecord, {
      departments: [{ id: 'dept-ops', name: 'Operational & Branch Banking' }],
      corporateUniverse: [
        {
          id: 'corp-root',
          name: 'PT AIFL Indonesia',
          children: [{ id: 'au-jkt', name: 'KC Jakarta Sudirman', department_id: 'dept-ops' }],
        },
      ],
      supervisors: ['Budi Santoso'],
      year: 2026,
      mode: 'per_row',
    })

    expect(draft.activities).toHaveLength(8)
    expect(draft.activities[0].dimensionName).toContain('TRX-PKAT-1')
    expect(draft.activities[0].auditUniverseId).toBe('au-jkt')
    expect(draft.activities[7].dimensionName).toContain('TRX-PKAT-8')
  })

  it('detects real row-level anomalies from sample_records without synthetic fallback templates', () => {
    const corpRecord: MappedRecordInput = {
      connection_id: 'conn-1',
      connection_name: 'Core Banking DB',
      table_name: 'corporate_transactions',
      target_scope: 'data_analytics',
      target_module: 'AI: Anomaly Detection',
      audit_field: 'trx_id',
      total_rows: 3244,
      sample_records: [
        {
          trx_id: 2781,
          trx_ref_number: 'NSA-AFT8-00001',
          trx_date: '2026-03-14 22:45:00',
          branch_name: 'KC Jakarta Sudirman',
          account_type: 'CORPORATE_CHECKING',
          amount: 485000000,
          authorization_status: 'PENDING_OVERRIDE',
          is_suspicious: true,
          notes: 'Transaksi di luar jam operasional tanpa dual-control',
        },
        {
          trx_id: 1,
          trx_ref_number: 'TRX-2026-00001',
          trx_date: '2026-01-10 10:15:00',
          branch_name: 'KC Bandung Asia Afrika',
          account_type: 'CORPORATE_CHECKING',
          amount: 12500000,
          authorization_status: 'AUTHORIZED',
          is_suspicious: false,
          notes: null,
        },
      ],
    }

    const result = buildRealTableAnomalies(corpRecord)
    expect(result.totalPopulation).toBe(3244)
    expect(result.anomalies).toHaveLength(1)
    expect(result.anomalies[0].refNumber).toBe('NSA-AFT8-00001')
    expect(result.anomalies[0].entity).toBe('KC Jakarta Sudirman')
    expect(result.anomalies[0].amount).toBe(485000000)
    expect(result.anomalies[0].description).toContain('Transaksi di luar jam operasional')
    expect(result.normalPoints).toHaveLength(1)
    expect(result.normalPoints[0].label).toBe('TRX-2026-00001')
  })
})
