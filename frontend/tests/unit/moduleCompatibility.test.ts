// @ts-nocheck
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import {
  analyzeTableSchema,
  evaluateModuleCompatibility,
  getDefaultModuleForTableScope,
  getModuleOptionsForTableScope,
  getNextRecommendedMappingForTable,
  getRecommendedMappingsForTable,
  getScopeOptionsForTable,
  targetModuleToImportMode,
  type TableCompatibilityInput,
} from '~/utils/moduleCompatibility'
import {
  buildWorkingPaperSamplesFromTable,
  evaluateTableModuleCompatibility,
} from '~/utils/mappedFeatureAdapters'

const REAL_TABLES: Record<string, TableCompatibilityInput> = {
  branches: {
    tableName: 'branches',
    rowCount: 10,
    columns: [
      { name: 'cabang_id', dataType: 'serial', isPrimary: true, isNullable: false },
      { name: 'branch_code', dataType: 'varchar(16)', isPrimary: false, isNullable: false },
      { name: 'branch_name', dataType: 'varchar(100)', isPrimary: false, isNullable: false },
      { name: 'region', dataType: 'varchar(60)', isPrimary: false, isNullable: false },
      { name: 'city', dataType: 'varchar(60)', isPrimary: false, isNullable: false },
      { name: 'risk_tier', dataType: 'varchar(20)', isPrimary: false, isNullable: false },
      { name: 'updated_at', dataType: 'timestamp', isPrimary: false, isNullable: false },
    ],
  },
  corporate_transactions: {
    tableName: 'corporate_transactions',
    rowCount: 3250,
    columns: [
      { name: 'trx_id', dataType: 'serial', isPrimary: true, isNullable: false },
      { name: 'trx_ref_number', dataType: 'varchar(30)', isPrimary: false, isNullable: false },
      { name: 'trx_date', dataType: 'timestamp', isPrimary: false, isNullable: false },
      { name: 'branch_name', dataType: 'varchar(100)', isPrimary: false, isNullable: false },
      { name: 'customer_name', dataType: 'varchar(150)', isPrimary: false, isNullable: false },
      { name: 'category', dataType: 'varchar(50)', isPrimary: false, isNullable: false },
      { name: 'channel', dataType: 'varchar(40)', isPrimary: false, isNullable: false },
      { name: 'account_type', dataType: 'varchar(30)', isPrimary: false, isNullable: false },
      { name: 'amount', dataType: 'numeric(18,2)', isPrimary: false, isNullable: false },
      { name: 'authorization_status', dataType: 'varchar(30)', isPrimary: false, isNullable: false },
      { name: 'is_suspicious', dataType: 'boolean', isPrimary: false, isNullable: false },
      { name: 'notes', dataType: 'text', isPrimary: false, isNullable: true },
      { name: 'updated_at', dataType: 'timestamp', isPrimary: false, isNullable: false },
    ],
  },
  credit_loan_portfolio: {
    tableName: 'credit_loan_portfolio',
    rowCount: 220,
    columns: [
      { name: 'loan_id', dataType: 'serial', isPrimary: true, isNullable: false },
      { name: 'loan_number', dataType: 'varchar(24)', isPrimary: false, isNullable: false },
      { name: 'product_name', dataType: 'varchar(100)', isPrimary: false, isNullable: false },
      { name: 'customer_name', dataType: 'varchar(150)', isPrimary: false, isNullable: false },
      { name: 'branch_name', dataType: 'varchar(100)', isPrimary: false, isNullable: false },
      { name: 'plafond', dataType: 'numeric(18,2)', isPrimary: false, isNullable: false },
      { name: 'outstanding', dataType: 'numeric(18,2)', isPrimary: false, isNullable: false },
      { name: 'interest_rate', dataType: 'numeric(5,2)', isPrimary: false, isNullable: false },
      { name: 'tenor_months', dataType: 'integer', isPrimary: false, isNullable: false },
      { name: 'collateral_value', dataType: 'numeric(18,2)', isPrimary: false, isNullable: false },
      { name: 'collectability', dataType: 'integer', isPrimary: false, isNullable: false },
      { name: 'days_past_due', dataType: 'integer', isPrimary: false, isNullable: false },
      { name: 'status', dataType: 'varchar(20)', isPrimary: false, isNullable: false },
      { name: 'disbursed_at', dataType: 'date', isPrimary: false, isNullable: false },
      { name: 'updated_at', dataType: 'timestamp', isPrimary: false, isNullable: false },
    ],
  },
  gl_journal_entries: {
    tableName: 'gl_journal_entries',
    rowCount: 420,
    columns: [
      { name: 'gl_id', dataType: 'serial', isPrimary: true, isNullable: false },
      { name: 'gl_date', dataType: 'date', isPrimary: false, isNullable: false },
      { name: 'gl_account_code', dataType: 'varchar(20)', isPrimary: false, isNullable: false },
      { name: 'gl_account_name', dataType: 'varchar(100)', isPrimary: false, isNullable: false },
      { name: 'voucher_number', dataType: 'varchar(30)', isPrimary: false, isNullable: true },
      { name: 'debit_amount', dataType: 'numeric(18,2)', isPrimary: false, isNullable: false },
      { name: 'credit_amount', dataType: 'numeric(18,2)', isPrimary: false, isNullable: false },
      { name: 'branch_name', dataType: 'varchar(100)', isPrimary: false, isNullable: false },
      { name: 'posted_by', dataType: 'varchar(60)', isPrimary: false, isNullable: false },
      { name: 'description', dataType: 'text', isPrimary: false, isNullable: true },
      { name: 'updated_at', dataType: 'timestamp', isPrimary: false, isNullable: false },
    ],
  },
  vendor_contracts: {
    tableName: 'vendor_contracts',
    rowCount: 150,
    columns: [
      { name: 'contract_id', dataType: 'serial', isPrimary: true, isNullable: false },
      { name: 'contract_no', dataType: 'varchar(32)', isPrimary: false, isNullable: false },
      { name: 'vendor_name', dataType: 'varchar(120)', isPrimary: false, isNullable: false },
      { name: 'branch_name', dataType: 'varchar(100)', isPrimary: false, isNullable: false },
      { name: 'contract_value', dataType: 'numeric(18,2)', isPrimary: false, isNullable: false },
      { name: 'po_number', dataType: 'varchar(32)', isPrimary: false, isNullable: false },
      { name: 'approval_status', dataType: 'varchar(30)', isPrimary: false, isNullable: false },
      { name: 'approver_id', dataType: 'varchar(50)', isPrimary: false, isNullable: false },
      { name: 'signed_at', dataType: 'date', isPrimary: false, isNullable: false },
      { name: 'is_flagged', dataType: 'boolean', isPrimary: false, isNullable: false },
      { name: 'audit_note', dataType: 'text', isPrimary: false, isNullable: true },
      { name: 'updated_at', dataType: 'timestamp', isPrimary: false, isNullable: false },
    ],
  },
  audit_risk_incidents: {
    tableName: 'audit_risk_incidents',
    rowCount: 40,
    columns: [
      { name: 'incident_id', dataType: 'serial', isPrimary: true, isNullable: false },
      { name: 'incident_code', dataType: 'varchar(24)', isPrimary: false, isNullable: false },
      { name: 'entity_unit', dataType: 'varchar(100)', isPrimary: false, isNullable: false },
      { name: 'risk_category', dataType: 'varchar(60)', isPrimary: false, isNullable: false },
      { name: 'severity_score', dataType: 'integer', isPrimary: false, isNullable: false },
      { name: 'exposure_value', dataType: 'numeric(18,2)', isPrimary: false, isNullable: false },
      { name: 'status', dataType: 'varchar(24)', isPrimary: false, isNullable: false },
      { name: 'finding_summary', dataType: 'text', isPrimary: false, isNullable: false },
      { name: 'reported_at', dataType: 'timestamp', isPrimary: false, isNullable: false },
      { name: 'is_deleted', dataType: 'boolean', isPrimary: false, isNullable: false },
    ],
  },
}

describe('Table-to-Module Compatibility Matrix & Smart Mapping', () => {
  it('classifies master reference table branches accurately and disables Working Paper, Anomaly, Strategic, NLP, KPI Forecast, and Risk Profile with reasons', () => {
    const branches = REAL_TABLES.branches
    const signals = analyzeTableSchema(branches)
    expect(signals.archetype).toBe('reference')
    expect(signals.isReferenceTable).toBe(true)
    expect(signals.hasNumericMeasure).toBe(false)
    expect(signals.hasNarrativeText).toBe(false)

    // Allowed modules
    expect(evaluateModuleCompatibility(branches, 'Audit Universe').compatible).toBe(true)
    expect(evaluateModuleCompatibility(branches, 'Annual Audit Plan').compatible).toBe(true)
    expect(evaluateModuleCompatibility(branches, 'AI: Entity Prediction').compatible).toBe(true)

    // Disallowed modules must be disabled with clear Indonesian reason
    for (const mod of [
      'Corporate Risk Profile',
      'Strategic Audit Plan',
      'Working Paper Evidence',
      'AI: Anomaly Detection',
      'AI: NLP Document Analysis',
      'AI: KPI & Loss Forecast',
    ]) {
      const ev = evaluateModuleCompatibility(branches, mod)
      expect(ev.compatible).toBe(false)
      expect(ev.reason.length).toBeGreaterThan(15)
    }

    // Dropdown options for audit_features disable Working Paper & Strategic Plan with reason
    const auditOpts = getModuleOptionsForTableScope('audit_features', branches)
    const wpOpt = auditOpts.find(o => o.value === 'Working Paper Evidence')
    const annualOpt = auditOpts.find(o => o.value === 'Annual Audit Plan')
    expect(wpOpt?.disabled).toBe(true)
    expect(wpOpt?.label).toContain('Tidak cocok')
    expect(annualOpt?.disabled).toBe(false)
    expect(annualOpt?.label).toContain('Direkomendasikan')

    // Never defaults to Working Paper
    expect(getDefaultModuleForTableScope('audit_features', branches)).toBe('Annual Audit Plan')
    expect(getNextRecommendedMappingForTable(branches).targetModule).toBe('Audit Universe')
  })

  it('disables AI: NLP Document Analysis for credit_loan_portfolio (no narrative text) while enabling numeric/credit modules', () => {
    const loans = REAL_TABLES.credit_loan_portfolio
    const signals = analyzeTableSchema(loans)
    expect(signals.archetype).toBe('transactional')
    expect(signals.hasMonetaryMeasure).toBe(true)
    expect(signals.hasQuantitativeKpi).toBe(true)
    expect(signals.hasNarrativeText).toBe(false)

    expect(evaluateModuleCompatibility(loans, 'Corporate Risk Profile').compatible).toBe(true)
    expect(evaluateModuleCompatibility(loans, 'Working Paper Evidence').compatible).toBe(true)
    expect(evaluateModuleCompatibility(loans, 'AI: KPI & Loss Forecast').recommended).toBe(true)
    expect(evaluateModuleCompatibility(loans, 'AI: Anomaly Detection').recommended).toBe(true)

    const nlpEval = evaluateModuleCompatibility(loans, 'AI: NLP Document Analysis')
    expect(nlpEval.compatible).toBe(false)
    expect(nlpEval.reason).toContain('teks naratif')

    const aiOpts = getModuleOptionsForTableScope('data_analytics', loans)
    const nlpOpt = aiOpts.find(o => o.value === 'AI: NLP Document Analysis')
    expect(nlpOpt?.disabled).toBe(true)
    expect(nlpOpt?.label).toContain('Tidak cocok')
  })

  it('adapts recommendations per table and never uses Working Paper Evidence as the blind default', () => {
    for (const [name, tbl] of Object.entries(REAL_TABLES)) {
      const nextRec = getNextRecommendedMappingForTable(tbl)
      expect(nextRec.targetModule).not.toBe('Working Paper Evidence')

      const auditDefault = getDefaultModuleForTableScope('audit_features', tbl)
      expect(auditDefault).not.toBe('Working Paper Evidence')

      const recs = getRecommendedMappingsForTable(tbl)
      expect(recs.length).toBeGreaterThan(0)

      // Every scope & module option must have non-empty value (safe for Reka UI)
      const scopes = getScopeOptionsForTable(tbl)
      for (const sc of scopes) {
        expect(sc.value).not.toBe('')
        const mods = getModuleOptionsForTableScope(sc.value, tbl)
        for (const m of mods) {
          expect(m.value).not.toBe('')
        }
      }

      if (name !== 'branches') {
        expect(evaluateModuleCompatibility(tbl, 'Working Paper Evidence').compatible).toBe(true)
      }
    }
  })

  it('classifies arbitrary unknown tables by schema structure without hardcoded table names', () => {
    const purePolicyTable: TableCompatibilityInput = {
      tableName: 'internal_policy_memos',
      rowCount: 85,
      columns: [
        { name: 'memo_id', dataType: 'uuid', isPrimary: true },
        { name: 'memo_no', dataType: 'varchar(32)', isPrimary: false },
        { name: 'policy_text', dataType: 'text', isPrimary: false },
        { name: 'effective_date', dataType: 'date', isPrimary: false },
      ],
    }

    const signals = analyzeTableSchema(purePolicyTable)
    expect(signals.archetype).toBe('narrative_document')
    expect(signals.hasNarrativeText).toBe(true)
    expect(signals.hasNumericMeasure).toBe(false)

    expect(evaluateModuleCompatibility(purePolicyTable, 'AI: NLP Document Analysis').compatible).toBe(true)
    expect(evaluateModuleCompatibility(purePolicyTable, 'AI: NLP Document Analysis').recommended).toBe(true)
    expect(evaluateModuleCompatibility(purePolicyTable, 'Working Paper Evidence').compatible).toBe(true)
    expect(evaluateModuleCompatibility(purePolicyTable, 'Strategic Audit Plan').compatible).toBe(false)
    expect(evaluateModuleCompatibility(purePolicyTable, 'AI: KPI & Loss Forecast').compatible).toBe(false)
    expect(evaluateModuleCompatibility(purePolicyTable, 'AI: Entity Prediction').compatible).toBe(false)

    // Access logs with boolean override flag but no narrative text
    const accessLogsTable: TableCompatibilityInput = {
      tableName: 'user_access_audit_logs',
      rowCount: 12000,
      columns: [
        { name: 'log_id', dataType: 'bigserial', isPrimary: true },
        { name: 'user_id', dataType: 'varchar(64)', isPrimary: false },
        { name: 'action_name', dataType: 'varchar(128)', isPrimary: false },
        { name: 'is_override_auth', dataType: 'boolean', isPrimary: false },
        { name: 'logged_at', dataType: 'timestamp', isPrimary: false },
      ],
    }

    expect(evaluateModuleCompatibility(accessLogsTable, 'AI: Anomaly Detection').compatible).toBe(true)
    expect(evaluateModuleCompatibility(accessLogsTable, 'Working Paper Evidence').compatible).toBe(true)
    expect(evaluateModuleCompatibility(accessLogsTable, 'AI: NLP Document Analysis').compatible).toBe(false)
  })

  it('enforces import modes across all 26 supported modules (Annual & Strategic Plan manual via banner, Working Paper & Activity Plan pre-fill, others auto)', () => {
    expect(targetModuleToImportMode('Annual Audit Plan')).toBe('manual_banner')
    expect(targetModuleToImportMode('Strategic Audit Plan')).toBe('manual_banner')
    expect(targetModuleToImportMode('Working Paper Evidence')).toBe('prefill_form')
    expect(targetModuleToImportMode('Audit Activity Plan')).toBe('prefill_form')
    expect(targetModuleToImportMode('Corporate Risk Profile')).toBe('auto')
    expect(targetModuleToImportMode('Audit Universe')).toBe('auto')
    expect(targetModuleToImportMode('Risk Appetite Statement')).toBe('auto')
    expect(targetModuleToImportMode('Risk Factors')).toBe('auto')
    expect(targetModuleToImportMode('Audit Priority')).toBe('auto')
    expect(targetModuleToImportMode('Risk Control Matrix (RCM)')).toBe('auto')
    expect(targetModuleToImportMode('Risk Mitigation')).toBe('auto')
    expect(targetModuleToImportMode('Assignment Letter')).toBe('auto')
    expect(targetModuleToImportMode('Audit Execution Status')).toBe('auto')
    expect(targetModuleToImportMode('Audit Fieldwork')).toBe('auto')
    expect(targetModuleToImportMode('Internal Audit Performance (KPI)')).toBe('auto')
    expect(targetModuleToImportMode('Audit Result Report (LHA)')).toBe('auto')
    expect(targetModuleToImportMode('Action Taken Report (ATR)')).toBe('auto')
    expect(targetModuleToImportMode('Quality Assurance Dashboard')).toBe('auto')
    expect(targetModuleToImportMode('Periodic Self-Assessment')).toBe('auto')
    expect(targetModuleToImportMode('SAIV Validation')).toBe('auto')
    expect(targetModuleToImportMode('IACM Assessment')).toBe('auto')
    expect(targetModuleToImportMode('Consulting Service')).toBe('auto')
    expect(targetModuleToImportMode('AI: Anomaly Detection')).toBe('auto')
    expect(targetModuleToImportMode('AI: Entity Prediction')).toBe('auto')
    expect(targetModuleToImportMode('AI: NLP Document Analysis')).toBe('auto')
    expect(targetModuleToImportMode('AI: KPI & Loss Forecast')).toBe('auto')
  })

  it('keeps Banner compatibility evaluation consistent with the single source of truth matrix', () => {
    const branchesBannerRec = {
      connection_id: 'c1',
      connection_name: 'Bank DB',
      table_name: 'branches',
      target_scope: 'audit_features',
      target_module: 'Working Paper Evidence',
      audit_field: 'cabang_id',
      total_rows: 10,
      profile: {
        has_numeric_measure: false,
        dimension_column: 'branch_name',
        amount_column: null,
        columns: ['cabang_id', 'branch_code', 'branch_name', 'region', 'city', 'risk_tier', 'updated_at'],
      },
      sample_records: [{ cabang_id: 1, branch_code: 'KCU-01', branch_name: 'KCU Sudirman', risk_tier: 'High Risk' }],
    }

    const wpCompat = evaluateTableModuleCompatibility(branchesBannerRec, 'working_paper')
    expect(wpCompat.compatible).toBe(false)
    expect(wpCompat.reason).toContain('tabel referensi master')
  })

  it('builds Working Paper F-03 samples adapted to table rows with Fail status for anomalous/override records', () => {
    const wpPayload = buildWorkingPaperSamplesFromTable({
      connection_id: 'c1',
      connection_name: 'Bank DB',
      table_name: 'corporate_transactions',
      target_scope: 'audit_features',
      target_module: 'Working Paper Evidence',
      audit_field: 'trx_id',
      total_rows: 3250,
      sample_records: [
        {
          trx_id: 1,
          trx_ref_number: 'NSA-2026-000001',
          branch_name: 'KCP Menteng Raya',
          amount: 15000000,
          authorization_status: 'AUTHORIZED',
          is_suspicious: false,
        },
        {
          trx_id: 2801,
          trx_ref_number: 'NSA-AFT8-00001',
          branch_name: 'KCU Sudirman Plaza',
          amount: 850000000,
          authorization_status: 'PENDING_OVERRIDE',
          is_suspicious: true,
        },
      ],
    })

    expect(wpPayload.sampleSize).toBe(2)
    expect(wpPayload.flaggedCount).toBe(1)
    expect(wpPayload.population).toContain('3.250')
    // Flagged row is prioritized first
    expect(wpPayload.samples[0].fieldworkDocument).toBe('NSA-AFT8-00001')
    expect(wpPayload.samples[0].l1).toBe('Fail')
    expect(wpPayload.samples[0].l2).toBe('Fail')
    expect(wpPayload.samples[0].l3).toBe('Fail')
    expect(wpPayload.samples[1].fieldworkDocument).toBe('NSA-2026-000001')
    expect(wpPayload.samples[1].l1).toBe('Pass')
    expect(wpPayload.samples[1].l3).toBe('Pass')
  })

  it('wires smart per-table dropdowns, recommendations, bulk import, and Working Paper F-03 modal into UI components', () => {
    const dataSourceVue = readFileSync(
      resolve(__dirname, '../../components/settings/data-source.vue'),
      'utf8'
    )
    expect(dataSourceVue).not.toContain("const defaultModule = 'Working Paper Evidence'")
    expect(dataSourceVue).toContain('getScopesForTable(table)')
    expect(dataSourceVue).toContain('getModulesForScope(mapping.targetScope, table)')
    expect(dataSourceVue).toContain('applyRecommendedMappings(table)')
    expect(dataSourceVue).toContain('saveAndImportAllMappings')
    expect(dataSourceVue).toContain('importAllMappingsForConn(conn)')
    expect(dataSourceVue).toContain('fixIncompatibleMapping(table, mIdx)')

    const bannerVue = readFileSync(
      resolve(__dirname, '../../components/common/MappedSourceDataBanner.vue'),
      'utf8'
    )
    expect(bannerVue).toContain('buildWorkingPaperSamplesFromTable')
    expect(bannerVue).toContain('wpStore.showModalF03 = true')
    expect(bannerVue).toContain('Form F-03 Working Paper Terisi')
  })
})
