import { AuditCategory } from '~/types/audit'
import { getCanonicalCategory } from '~/composables/useAnomalyScatter'
import {
  normalizeAnomalyRiskLevel,
  normalizeEntityRiskLevel,
  normalizeNlpRiskLevel,
  toEntityRiskCode,
  toNlpRiskCode,
} from '~/utils/riskLevelBadge'
import { sanitizeKpiNumberInput, STRATEGIC_PLAN_CATEGORIES } from '~/utils/strategicPlanPayload'
import {
  actionTypeToTargetModule,
  analyzeTableSchema,
  evaluateModuleCompatibility,
} from '~/utils/moduleCompatibility'

export interface TableDimensionSummary {
  name: string
  row_count: number
  flagged_count: number
  total_amount: number
  max_amount: number
}

export interface TableProfileMeta {
  table_name?: string
  total_rows?: number
  flagged_rows?: number
  has_numeric_measure?: boolean
  dimension_column?: string | null
  amount_column?: string | null
  columns?: string[]
  total_amount?: number
  avg_amount?: number
  max_amount?: number
  metrics?: Record<string, number>
  dimension_breakdown?: TableDimensionSummary[]
}

export interface MappedRecordInput {
  connection_id: string
  connection_name: string
  table_name: string
  target_scope: string
  target_module: string
  audit_field: string
  total_rows?: number
  profile?: TableProfileMeta
  data_hub_unavailable?: boolean
  anomaly_rules?: string[]
  custom_rules?: any[]
  sample_records: Record<string, any>[]
}

export interface BannerModuleInfo {
  scope: string
  module: string
  slug: string
  label: string
  actionType:
    | 'risk_profile'
    | 'audit_universe'
    | 'risk_appetite'
    | 'risk_factors'
    | 'audit_priority'
    | 'rcm'
    | 'risk_mitigation'
    | 'strategic_plan'
    | 'annual_plan'
    | 'working_paper'
    | 'audit_activity_plan'
    | 'assignment_letter'
    | 'audit_execution_status'
    | 'audit_fieldwork'
    | 'internal_audit_kpi'
    | 'audit_result_report'
    | 'action_taken_report'
    | 'qar_dashboard'
    | 'periodic_self_assessment'
    | 'saiv_validation'
    | 'iacm_assessment'
    | 'consulting_service'
    | 'ai_anomaly'
    | 'ai_entity_prediction'
    | 'ai_nlp'
    | 'ai_kpi_forecast'
    | 'unsupported'
  actionLabel: string
  supported: boolean
}

export interface CompatibilityCheckResult {
  compatible: boolean
  reason?: string
  hasNumericMeasure: boolean
  dimensionColumn: string | null
  amountColumn: string | null
  detectedColumns: string[]
}

export interface StrategicPlanKpiCandidate {
  id: string
  selected: boolean
  strategicObjective: string
  kpi: string
  category: (typeof STRATEGIC_PLAN_CATEGORIES)[number]
  unit: string
  hibHig: 'HIG' | 'HIB'
  actual: string
  target: string
  internalAuditSO: string
  yearStart: number
  yearEnd: number
  rationale: string
}

export interface AnnualPlanActivityCandidate {
  selected: boolean
  dimensionName: string
  name: string
  category: AuditCategory
  departmentId: string
  departmentName: string
  auditUniverseId: string
  auditUniverseName: string
  riskLevel?: string
  supervisorName: string
  auditorCount: number
  totalMandays: number
  timelineText: string
  rowCount: number
  flaggedCount: number
}

export interface AnnualPlanDraftCandidate {
  title: string
  year: string
  selectedMonths: number[]
  notes: string
  activities: AnnualPlanActivityCandidate[]
}

export interface CorporateUniverseCandidate {
  id: string
  name: string
  department_id?: string | null
  children?: CorporateUniverseCandidate[]
}

export interface AnnualPlanAdapterContext {
  departments: Array<{ id: string; name: string }>
  corporateUniverse: CorporateUniverseCandidate[]
  supervisors: string[]
  year?: number
  mode?: 'per_row' | 'per_dimension'
}

export interface RiskProfileRowCandidate {
  refCode: string
  name: string
  category: string
  description: string
  impact: number
  likelihood: number
  severity: number
  location_id?: string
  branch?: string
  source: string
  assessments: Array<{
    year: number
    impact_q1: number
    impact_q2: number
    impact_q3: number
    impact_q4: number
    likelihood_q1: number
    likelihood_q2: number
    likelihood_q3: number
    likelihood_q4: number
  }>
}

export interface AuditUniverseNodeCandidate {
  name: string
  dimensionName: string
  refCode: string
}

const DIMENSION_COL_CANDIDATES = [
  'branch_name',
  'entity_unit',
  'department',
  'reporting_unit',
  'vendor_name',
  'gl_account_name',
  'region',
  'branch_code',
  'product_name',
  'product_type',
  'category',
  'customer_name',
]

const AMOUNT_COL_CANDIDATES = [
  'amount',
  'plafond',
  'outstanding',
  'contract_value',
  'invoice_amount',
  'exposure_value',
  'debit_amount',
  'principal_amount',
  'balance',
]

const NUMERIC_OR_FLAG_COL_CANDIDATES = new Set([
  ...AMOUNT_COL_CANDIDATES,
  'credit_amount',
  'severity_score',
  'collectability',
  'days_past_due',
  'interest_rate',
  'tenor_months',
  'is_suspicious',
  'is_flagged',
  'is_override_auth',
  'voucher_number',
])

const KNOWN_REFERENCE_ONLY_TABLES = new Set(['branches'])

export function resolveBannerModuleInfo(routePath: string): BannerModuleInfo | null {
  const path = (routePath || '').toLowerCase()
  if (!path || path.startsWith('/settings') || path.startsWith('/auth')) {
    return null
  }

  // 1. Risk Management Scope
  if (path === '/risk-profile' || path === '/risk-profile/') {
    return {
      scope: 'risk_management',
      module: 'Corporate Risk Profile',
      slug: 'corporate_risk_profile',
      label: 'Corporate Risk Profile',
      actionType: 'risk_profile',
      actionLabel: '+ Tambahkan ke Profil Risiko',
      supported: true,
    }
  }
  if (path.includes('/risk-profile/audit-universe')) {
    return {
      scope: 'risk_management',
      module: 'Audit Universe',
      slug: 'audit_universe',
      label: 'Audit Universe',
      actionType: 'audit_universe',
      actionLabel: '+ Tambahkan ke Audit Universe',
      supported: true,
    }
  }
  if (path.includes('/risk-profile/risk-factors')) {
    return {
      scope: 'risk_management',
      module: 'Risk Factors',
      slug: 'risk_factors',
      label: 'Risk Factors',
      actionType: 'risk_factors',
      actionLabel: '+ Tambahkan Faktor Risiko',
      supported: true,
    }
  }
  if (path.includes('/risk-profile/risk-control-matrix')) {
    return {
      scope: 'risk_management',
      module: 'Risk Control Matrix (RCM)',
      slug: 'risk_control_matrix',
      label: 'Risk Control Matrix (RCM)',
      actionType: 'rcm',
      actionLabel: '+ Tambahkan Kontrol RCM',
      supported: true,
    }
  }
  if (path.includes('/risk-profile/audit-priority')) {
    return {
      scope: 'risk_management',
      module: 'Audit Priority',
      slug: 'audit_priority',
      label: 'Audit Priority',
      actionType: 'audit_priority',
      actionLabel: '+ Hitung & Tetapkan Prioritas Audit',
      supported: true,
    }
  }
  if (path.includes('/risk-appetite')) {
    return {
      scope: 'risk_management',
      module: 'Risk Appetite Statement',
      slug: 'risk_appetite',
      label: 'Risk Appetite Statement',
      actionType: 'risk_appetite',
      actionLabel: '+ Buat Risk Appetite Statement',
      supported: true,
    }
  }
  if (path.includes('/mitigation')) {
    return {
      scope: 'risk_management',
      module: 'Risk Mitigation',
      slug: 'risk_mitigation',
      label: 'Risk Mitigation',
      actionType: 'risk_mitigation',
      actionLabel: '+ Buat Rencana Mitigasi Risiko',
      supported: true,
    }
  }

  // 2. Audit Features Scope
  if (path.includes('/strategic-audit-plan')) {
    return {
      scope: 'audit_features',
      module: 'Strategic Audit Plan',
      slug: 'strategic_audit_plan',
      label: 'Strategic Audit Plan',
      actionType: 'strategic_plan',
      actionLabel: '+ Buat KPI Strategic Plan',
      supported: true,
    }
  }
  if (path.includes('/annual-audit')) {
    return {
      scope: 'audit_features',
      module: 'Annual Audit Plan',
      slug: 'annual_audit',
      label: 'Annual Audit Plan',
      actionType: 'annual_plan',
      actionLabel: '+ Buat Draft Rencana Audit',
      supported: true,
    }
  }
  if (path.includes('/working-paper')) {
    return {
      scope: 'audit_features',
      module: 'Working Paper Evidence',
      slug: 'working_paper',
      label: 'Working Paper Evidence',
      actionType: 'working_paper',
      actionLabel: '+ Jadikan Bukti Sampel Kertas Kerja',
      supported: true,
    }
  }
  if (path.includes('/audit-fieldwork')) {
    return {
      scope: 'audit_features',
      module: 'Audit Fieldwork',
      slug: 'audit_fieldwork',
      label: 'Audit Fieldwork',
      actionType: 'audit_fieldwork',
      actionLabel: '+ Impor Sampel & Uji Kontrol Fieldwork',
      supported: true,
    }
  }
  if (path.includes('/audit-activity-plan')) {
    return {
      scope: 'audit_features',
      module: 'Audit Activity Plan',
      slug: 'audit_activity_plan',
      label: 'Audit Activity Plan',
      actionType: 'audit_activity_plan',
      actionLabel: '+ Isi Formulir Rencana Kegiatan',
      supported: true,
    }
  }
  if (path.includes('/assignment-letter')) {
    return {
      scope: 'audit_features',
      module: 'Assignment Letter',
      slug: 'assignment_letter',
      label: 'Assignment Letter',
      actionType: 'assignment_letter',
      actionLabel: '+ Buat Draft Surat Tugas',
      supported: true,
    }
  }
  if (path.includes('/audit-execution-status')) {
    return {
      scope: 'audit_features',
      module: 'Audit Execution Status',
      slug: 'audit_execution_status',
      label: 'Audit Execution Status',
      actionType: 'audit_execution_status',
      actionLabel: '+ Sinkronkan Status Pelaksanaan Audit',
      supported: true,
    }
  }
  if (path.includes('/audit-result-report')) {
    return {
      scope: 'audit_features',
      module: 'Audit Result Report (LHA)',
      slug: 'audit_result_report',
      label: 'Audit Result Report (LHA)',
      actionType: 'audit_result_report',
      actionLabel: '+ Buat Draft Laporan Hasil Audit (LHA)',
      supported: true,
    }
  }
  if (path.includes('/action-taken-report')) {
    return {
      scope: 'audit_features',
      module: 'Action Taken Report (ATR)',
      slug: 'action_taken_report',
      label: 'Action Taken Report (ATR)',
      actionType: 'action_taken_report',
      actionLabel: '+ Buat Tindak Lanjut Temuan (ATR)',
      supported: true,
    }
  }
  if (path.includes('/kpi-performance')) {
    return {
      scope: 'audit_features',
      module: 'Internal Audit Performance (KPI)',
      slug: 'kpi_performance',
      label: 'Internal Audit Performance (KPI)',
      actionType: 'internal_audit_kpi',
      actionLabel: '+ Tambahkan Realisasi KPI Audit',
      supported: true,
    }
  }

  // 2b. QAR Features Scope
  if (path.includes('/quality-assurance/import-periodic-self-assessment') || path.includes('type=regular')) {
    return {
      scope: 'qar_features',
      module: 'Periodic Self-Assessment',
      slug: 'periodic_self_assessment',
      label: 'Periodic Self-Assessment',
      actionType: 'periodic_self_assessment',
      actionLabel: '+ Tambahkan Laporan Periodic Self-Assessment',
      supported: true,
    }
  }
  if (path.includes('/quality-assurance/import-saiv') || path.includes('type=saiv')) {
    return {
      scope: 'qar_features',
      module: 'SAIV Validation',
      slug: 'saiv_validation',
      label: 'SAIV Validation',
      actionType: 'saiv_validation',
      actionLabel: '+ Tambahkan Laporan SAIV Validation',
      supported: true,
    }
  }
  if (path.includes('/quality-assurance/import-iacm') || path.includes('type=iacm')) {
    return {
      scope: 'qar_features',
      module: 'IACM Assessment',
      slug: 'iacm_assessment',
      label: 'IACM Assessment',
      actionType: 'iacm_assessment',
      actionLabel: '+ Tambahkan Laporan IACM Assessment',
      supported: true,
    }
  }
  if (path.includes('/quality-assurance')) {
    return {
      scope: 'qar_features',
      module: 'Quality Assurance Dashboard',
      slug: 'quality_assurance_dashboard',
      label: 'Quality Assurance Dashboard',
      actionType: 'qar_dashboard',
      actionLabel: '+ Tambahkan Laporan Evaluasi Mutu (QAR)',
      supported: true,
    }
  }
  if (path.includes('/consulting-service')) {
    return {
      scope: 'qar_features',
      module: 'Consulting Service',
      slug: 'consulting_service',
      label: 'Consulting Service',
      actionType: 'consulting_service',
      actionLabel: '+ Buat Penugasan Consulting Service',
      supported: true,
    }
  }

  // 3. Data Analytics & AI Scope
  if (path.includes('/analytics/ai/anomaly-detection') || path.includes('/ai/anomaly-detection')) {
    return {
      scope: 'data_analytics',
      module: 'AI: Anomaly Detection',
      slug: 'anomaly_detection',
      label: 'AI Anomaly Detection',
      actionType: 'ai_anomaly',
      actionLabel: '+ Jalankan AI Anomaly Detection',
      supported: true,
    }
  }
  if (path.includes('/analytics/ai/risk-scoring') || path.includes('/analytics/ai/entity-prediction')) {
    return {
      scope: 'data_analytics',
      module: 'AI: Entity Prediction',
      slug: 'entity_prediction',
      label: 'AI Entity Prediction (XGBoost)',
      actionType: 'ai_entity_prediction',
      actionLabel: '+ Prediksi Entitas AI',
      supported: true,
    }
  }
  if (path.includes('/analytics/ai/nlp')) {
    return {
      scope: 'data_analytics',
      module: 'AI: NLP Document Analysis',
      slug: 'nlp_analysis',
      label: 'AI NLP Document Analysis',
      actionType: 'ai_nlp',
      actionLabel: '+ Analisis Teks AI (NLP)',
      supported: true,
    }
  }
  if (path.includes('/analytics/ai/kpi-forecast') || path.includes('/analytics/kpi-forecast')) {
    return {
      scope: 'data_analytics',
      module: 'AI: KPI & Loss Forecast',
      slug: 'kpi_forecast',
      label: 'AI KPI & Loss Forecast',
      actionType: 'ai_kpi_forecast',
      actionLabel: '+ Prediksi Tren KPI',
      supported: true,
    }
  }
  if (path.includes('/analytics/caatt')) {
    return null
  }

  return null
}

export function extractDetectedColumns(record: MappedRecordInput): string[] {
  const colSet = new Set<string>()
  if (Array.isArray(record.profile?.columns)) {
    for (const c of record.profile.columns) {
      if (c) colSet.add(String(c))
    }
  }
  if (Array.isArray(record.sample_records)) {
    for (const row of record.sample_records) {
      if (row && typeof row === 'object') {
        for (const k of Object.keys(row)) {
          colSet.add(k)
        }
      }
    }
  }
  return Array.from(colSet)
}

export function evaluateTableModuleCompatibility(
  record: MappedRecordInput,
  actionType: BannerModuleInfo['actionType']
): CompatibilityCheckResult {
  const compatInput = {
    tableName: record.table_name,
    rowCount: record.total_rows || record.profile?.total_rows || record.sample_records?.length || 0,
    profile: record.profile,
    sample_records: record.sample_records,
    data_hub_unavailable: record.data_hub_unavailable,
  }

  const signals = analyzeTableSchema(compatInput)
  const targetModule =
    actionType === 'unsupported'
      ? record.target_module || 'Unknown'
      : actionTypeToTargetModule(actionType) || record.target_module

  const evalResult = evaluateModuleCompatibility(compatInput, targetModule)

  return {
    compatible: evalResult.compatible,
    reason: evalResult.compatible ? undefined : evalResult.reason,
    hasNumericMeasure: signals.hasNumericMeasure,
    dimensionColumn: signals.dimensionColumn,
    amountColumn: signals.amountColumn,
    detectedColumns: signals.detectedColumns,
  }
}

export function isRowAnomalousOrFlagged(
  row: Record<string, any>,
  customRules?: any[],
  presetRules?: string[]
): { flagged: boolean; reasons: string[]; severity: string } {
  if (!row || typeof row !== 'object') {
    return { flagged: false, reasons: [], severity: 'Moderate' }
  }

  const reasons: string[] = []
  let highSeverity = false

  // 1. Custom Rules evaluation
  if (Array.isArray(customRules)) {
    for (const cr of customRules) {
      if (!cr || cr.isActive === false || !cr.field) continue
      const rawVal = row[cr.field]
      if (rawVal === undefined || rawVal === null) continue
      const numVal = Number(rawVal)
      const thresh = Number(cr.threshold)
      let matched = false
      switch (cr.operator) {
        case '>':
          matched = Number.isFinite(numVal) && numVal > thresh
          break
        case '>=':
          matched = Number.isFinite(numVal) && numVal >= thresh
          break
        case '<':
          matched = Number.isFinite(numVal) && numVal < thresh
          break
        case '<=':
          matched = Number.isFinite(numVal) && numVal <= thresh
          break
        case '==':
          matched = String(rawVal) === String(cr.thresholdStr ?? cr.threshold)
          break
        case '!=':
          matched = String(rawVal) !== String(cr.thresholdStr ?? cr.threshold)
          break
        case 'contains':
          matched = String(rawVal).toLowerCase().includes(String(cr.thresholdStr ?? cr.threshold).toLowerCase())
          break
      }
      if (matched) {
        reasons.push(`[Rule: ${cr.name}] ${cr.field} (${rawVal}) ${cr.operator} ${cr.thresholdStr ?? cr.threshold}`)
        if (String(cr.severity).toUpperCase() === 'CRITICAL' || String(cr.severity).toUpperCase() === 'HIGH') {
          highSeverity = true
        }
      }
    }
  }

  // 2. Preset Rules & Database Column Evidence
  if (row.is_suspicious === true) {
    reasons.push(row.notes ? String(row.notes) : 'Ditandai is_suspicious = TRUE di database sumber')
    highSeverity = true
  }
  if (row.is_flagged === true) {
    reasons.push(row.audit_note ? String(row.audit_note) : 'Kontrak/dokumen ditandai is_flagged = TRUE')
    highSeverity = true
  }
  if (row.is_override_auth === true) {
    reasons.push(`Override otorisasi sistem (${row.action_name || 'Privileged Action'})`)
    highSeverity = true
  }
  if (
    row.authorization_status &&
    !['AUTHORIZED', 'APPROVED', 'APPROVED_DUAL'].includes(String(row.authorization_status).toUpperCase())
  ) {
    reasons.push(`Status otorisasi tidak standar: ${row.authorization_status}`)
    highSeverity = true
  }
  if (
    row.approval_status &&
    !['APPROVED'].includes(String(row.approval_status).toUpperCase())
  ) {
    reasons.push(`Status persetujuan pengadaan: ${row.approval_status}`)
  }
  if (Number(row.collectability) >= 3 || Number(row.days_past_due) > 90) {
    reasons.push(`Kredit bermasalah (Kolektibilitas ${row.collectability || '-'}, DPD ${row.days_past_due || 0} hari)`)
    highSeverity = true
  }
  if (Number(row.plafond) > 5_000_000_000) {
    reasons.push(`Plafon kredit Rp ${Number(row.plafond).toLocaleString('id-ID')} melampaui batas standar Rp 5 Miliar`)
    highSeverity = true
  }
  if (Number(row.interest_rate) > 24) {
    reasons.push(`Suku bunga kredit ${row.interest_rate}% melampaui batas kebijakan 24%`)
  }
  if (Number(row.tenor_months) > 120) {
    reasons.push(`Tenor kredit ${row.tenor_months} bulan melampaui batas kebijakan 120 bulan`)
  }
  if ('voucher_number' in row && (row.voucher_number === null || String(row.voucher_number).trim() === '')) {
    reasons.push(`Jurnal GL tanpa nomor voucher pendukung (GL Account ${row.gl_account_code || '-'})`)
    highSeverity = true
  }
  if (Number(row.severity_score) >= 70) {
    reasons.push(
      row.finding_summary
        ? `${row.finding_summary} (Severity ${row.severity_score})`
        : `Skor keparahan insiden tinggi (${row.severity_score})`
    )
    highSeverity = true
  }

  // Preset rule checks on amount/timestamp
  const amt = Number(
    row.amount ??
      row.plafond ??
      row.contract_value ??
      row.invoice_amount ??
      row.exposure_value ??
      row.debit_amount ??
      row.principal_amount ??
      0
  )
  const hasThresholdRule = (presetRules || []).some(r => r.toLowerCase().includes('threshold'))
  if ((hasThresholdRule && amt > 100_000_000) || amt >= 480_000_000) {
    if (reasons.length === 0) {
      reasons.push(`Nilai transaksi/eksposur tinggi Rp ${amt.toLocaleString('id-ID')} melampaui ambang pengawasan`)
    }
  }

  if (reasons.length === 0) {
    return { flagged: false, reasons: [], severity: 'Low' }
  }

  return {
    flagged: true,
    reasons,
    severity: highSeverity ? 'High' : 'Moderate to High',
  }
}

export function computeStrategicKpiMetrics(
  actualStr: string,
  targetStr: string,
  hibHig: 'HIG' | 'HIB'
): { calculation: string; status: 'Good' | 'Moderate' | 'Poor' | '' } {
  const actual = parseFloat(sanitizeKpiNumberInput(actualStr))
  const target = parseFloat(sanitizeKpiNumberInput(targetStr))
  if (!Number.isFinite(actual) || !Number.isFinite(target)) {
    return { calculation: '', status: '' }
  }

  let ratio = 0
  if (hibHig === 'HIG') {
    if (target === 0) return { calculation: '', status: '' }
    ratio = actual / target
  } else {
    if (actual === 0) {
      ratio = target >= 0 ? 1 : 0
    } else if (target === 0) {
      ratio = 0
    } else {
      ratio = target / actual
    }
  }

  const calculation = `${(ratio * 100).toFixed(2)}%`
  const status: 'Good' | 'Moderate' | 'Poor' = ratio >= 1 ? 'Good' : ratio >= 0.7 ? 'Moderate' : 'Poor'
  return { calculation, status }
}

export function buildStrategicPlanCandidatesFromTable(
  record: MappedRecordInput,
  currentYear = new Date().getFullYear()
): StrategicPlanKpiCandidate[] {
  const compat = evaluateTableModuleCompatibility(record, 'strategic_plan')
  if (!compat.compatible) return []

  const tbl = record.table_name
  const conn = record.connection_name
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const totalRows = Math.max(1, Number(record.profile?.total_rows || record.total_rows || rows.length || 1))
  const metrics = record.profile?.metrics || {}

  const flaggedRowsCount =
    record.profile?.flagged_rows ??
    rows.filter(r => isRowAnomalousOrFlagged(r, record.custom_rules, record.anomaly_rules).flagged).length

  const candidates: StrategicPlanKpiCandidate[] = []

  if (
    tbl === 'corporate_transactions' ||
    metrics.suspicious_count !== undefined ||
    rows.some(r => 'is_suspicious' in r || 'authorization_status' in r)
  ) {
    const suspCount =
      metrics.suspicious_count ??
      rows.filter(r => r.is_suspicious === true).length
    const unauthCount =
      metrics.unauthorized_count ??
      rows.filter(
        r =>
          r.authorization_status &&
          !['AUTHORIZED', 'APPROVED', 'APPROVED_DUAL'].includes(String(r.authorization_status).toUpperCase())
      ).length
    const denom = record.profile?.total_rows ? totalRows : Math.max(1, rows.length)
    const suspRate = +(((suspCount || flaggedRowsCount) / denom) * 100).toFixed(2)
    const authComplianceRate = +(100 - (unauthCount / denom) * 100).toFixed(2)

    candidates.push(
      {
        id: `${tbl}-suspicious-rate`,
        selected: true,
        strategicObjective: `Penguatan Pengendalian Otorisasi & Pencegahan Transaksi Mencurigakan (${tbl})`,
        kpi: `Rasio Transaksi Korporasi Mencurigakan / Anomali (${tbl})`,
        category: 'Financial',
        unit: '%',
        hibHig: 'HIB',
        actual: String(suspRate),
        target: '2.00',
        internalAuditSO: `Menurunkan insiden transaksi mencurigakan dan pemecahan nilai transaksi pada pipeline ${conn} (${tbl})`,
        yearStart: currentYear,
        yearEnd: currentYear + 4,
        rationale: `Dihitung dari ${suspCount || flaggedRowsCount} transaksi berisiko atas populasi ${denom.toLocaleString('id-ID')} baris pada tabel ${tbl}.`,
      },
      {
        id: `${tbl}-auth-compliance`,
        selected: true,
        strategicObjective: `Kepatuhan Otorisasi Berjenjang (Dual-Control) Transaksi Korporasi (${tbl})`,
        kpi: `Tingkat Kepatuhan Otorisasi Standar Transaksi (${tbl})`,
        category: 'Operational',
        unit: '%',
        hibHig: 'HIG',
        actual: String(authComplianceRate),
        target: '98.00',
        internalAuditSO: `Memastikan seluruh transaksi korporasi melalui otorisasi standar tanpa pending override (${conn} / ${tbl})`,
        yearStart: currentYear,
        yearEnd: currentYear + 4,
        rationale: `Dihitung dari tingkat transaksi terotorisasi penuh (${denom - unauthCount} dari ${denom.toLocaleString('id-ID')} baris).`,
      }
    )
    return candidates
  }

  if (
    tbl === 'credit_loan_portfolio' ||
    tbl === 'loan_accounts' ||
    metrics.npl_count !== undefined ||
    rows.some(r => 'collectability' in r || 'days_past_due' in r || 'plafond' in r)
  ) {
    const denom = record.profile?.total_rows ? totalRows : Math.max(1, rows.length)
    const nplCount =
      metrics.npl_count ??
      rows.filter(r => Number(r.collectability) >= 3 || Number(r.days_past_due) > 90).length
    const breachCount =
      metrics.policy_breach_count ??
      rows.filter(r => Number(r.plafond) > 5_000_000_000 || Number(r.interest_rate) > 24 || Number(r.tenor_months) > 120).length
    const nplRate = +((nplCount / denom) * 100).toFixed(2)
    const complianceRate = +(100 - (breachCount / denom) * 100).toFixed(2)

    candidates.push(
      {
        id: `${tbl}-npl-rate`,
        selected: true,
        strategicObjective: `Pengendalian Kualitas Aset & Mitigasi Kredit Bermasalah (${tbl})`,
        kpi: `Rasio Kredit Bermasalah / NPL Kolektibilitas 3-5 (${tbl})`,
        category: 'Financial',
        unit: '%',
        hibHig: 'HIB',
        actual: String(nplRate),
        target: '3.00',
        internalAuditSO: `Mengawasi efektivitas restrukturisasi dan early warning kredit bermasalah pada ${conn} (${tbl})`,
        yearStart: currentYear,
        yearEnd: currentYear + 4,
        rationale: `Dihitung dari ${nplCount} fasilitas kredit KOL 3-5 / DPD > 90 atas populasi ${denom.toLocaleString('id-ID')} rekening.`,
      },
      {
        id: `${tbl}-underwriting-compliance`,
        selected: true,
        strategicObjective: `Kepatuhan Kebijakan Underwriting & Batas Plafon Kredit (${tbl})`,
        kpi: `Tingkat Kepatuhan Plafon, Suku Bunga & Tenor Kredit (${tbl})`,
        category: 'Quality',
        unit: '%',
        hibHig: 'HIG',
        actual: String(complianceRate),
        target: '95.00',
        internalAuditSO: `Memastikan kepatuhan limit plafon, suku bunga, dan tenor kredit terhadap kebijakan komite (${tbl})`,
        yearStart: currentYear,
        yearEnd: currentYear + 4,
        rationale: `Dihitung dari ${denom - breachCount} rekening yang patuh kebijakan atas total ${denom.toLocaleString('id-ID')} fasilitas.`,
      }
    )
    return candidates
  }

  if (
    tbl === 'gl_journal_entries' ||
    tbl === 'gl_transactions' ||
    metrics.missing_voucher_count !== undefined ||
    rows.some(r => 'voucher_number' in r || 'gl_account_code' in r)
  ) {
    const denom = record.profile?.total_rows ? totalRows : Math.max(1, rows.length)
    const missingCount =
      metrics.missing_voucher_count ??
      rows.filter(r => r.voucher_number === null || String(r.voucher_number ?? '').trim() === '').length
    const completenessRate = +(100 - (missingCount / denom) * 100).toFixed(2)

    candidates.push({
      id: `${tbl}-voucher-completeness`,
      selected: true,
      strategicObjective: `Integritas Pembukuan Jurnal Umum & Kelengkapan Dokumen Voucher (${tbl})`,
      kpi: `Tingkat Kelengkapan Nomor Voucher & Dokumen Jurnal GL (${tbl})`,
      category: 'Financial',
      unit: '%',
      hibHig: 'HIG',
      actual: String(completenessRate),
      target: '99.50',
      internalAuditSO: `Menjamin setiap jurnal buku besar memiliki nomor voucher valid tanpa celah nomor urut (${conn} / ${tbl})`,
      yearStart: currentYear,
      yearEnd: currentYear + 4,
      rationale: `Dihitung dari ${denom - missingCount} jurnal lengkap (${missingCount} tanpa voucher) atas populasi ${denom.toLocaleString('id-ID')} baris.`,
    })
    return candidates
  }

  if (
    tbl === 'vendor_contracts' ||
    tbl === 'vendor_invoices' ||
    metrics.flagged_contract_count !== undefined ||
    rows.some(r => 'contract_no' in r || 'invoice_no' in r || 'is_flagged' in r)
  ) {
    const denom = record.profile?.total_rows ? totalRows : Math.max(1, rows.length)
    const flgCount =
      metrics.flagged_contract_count ??
      rows.filter(r => r.is_flagged === true || (r.approval_status && r.approval_status !== 'APPROVED')).length
    const flgRate = +((flgCount / denom) * 100).toFixed(2)

    candidates.push({
      id: `${tbl}-procurement-risk-rate`,
      selected: true,
      strategicObjective: `Kepatuhan Pengadaan & Pencegahan Pemecahan Kontrak Vendor (${tbl})`,
      kpi: `Rasio Kontrak Pengadaan Berisiko / Deviasi Tender (${tbl})`,
      category: 'Operational',
      unit: '%',
      hibHig: 'HIB',
      actual: String(flgRate),
      target: '5.00',
      internalAuditSO: `Mencegah split purchase order dan memastikan seluruh kontrak vendor disetujui komite (${conn} / ${tbl})`,
      yearStart: currentYear,
      yearEnd: currentYear + 4,
      rationale: `Dihitung dari ${flgCount} kontrak berisiko/override atas populasi ${denom.toLocaleString('id-ID')} kontrak.`,
    })
    return candidates
  }

  if (
    tbl === 'audit_risk_incidents' ||
    tbl === 'compliance_incident_records' ||
    metrics.high_severity_count !== undefined ||
    rows.some(r => 'severity_score' in r || 'incident_code' in r)
  ) {
    const denom = record.profile?.total_rows ? totalRows : Math.max(1, rows.length)
    const highCount =
      metrics.high_severity_count ??
      rows.filter(r => Number(r.severity_score) >= 70).length
    const highRate = +((highCount / denom) * 100).toFixed(2)

    candidates.push({
      id: `${tbl}-high-severity-rate`,
      selected: true,
      strategicObjective: `Penurunan Insiden Risiko & Kepatuhan Berdampak Tinggi (${tbl})`,
      kpi: `Rasio Insiden Risiko & Kepatuhan Berstatus High/Escalated (${tbl})`,
      category: 'Issue',
      unit: '%',
      hibHig: 'HIB',
      actual: String(highRate),
      target: '10.00',
      internalAuditSO: `Mempercepat penyelesaian tindak lanjut insiden risiko berdampak tinggi pada ${conn} (${tbl})`,
      yearStart: currentYear,
      yearEnd: currentYear + 4,
      rationale: `Dihitung dari ${highCount} insiden severity >= 70 atas populasi ${denom.toLocaleString('id-ID')} insiden.`,
    })
    return candidates
  }

  const denom = record.profile?.total_rows ? totalRows : Math.max(1, rows.length)
  const devRate = +((flaggedRowsCount / denom) * 100).toFixed(2)
  candidates.push({
    id: `${tbl}-generic-deviation-rate`,
    selected: true,
    strategicObjective: `Peningkatan Efektivitas Pengendalian Internal Pipeline Data (${tbl})`,
    kpi: `Rasio Deviasi Pengendalian Data (${tbl})`,
    category: 'Operational',
    unit: '%',
    hibHig: 'HIB',
    actual: String(devRate),
    target: '5.00',
    internalAuditSO: `Pemantauan berkelanjutan atas kualitas dan kepatuhan data pada ${conn} (${tbl})`,
    yearStart: currentYear,
    yearEnd: currentYear + 4,
    rationale: `Dihitung dari ${flaggedRowsCount} baris terindikasi deviasi atas total ${denom.toLocaleString('id-ID')} baris.`,
  })
  return candidates
}

export function calculateQuartersFromMonths(months: number[]): string[] {
  const qSet = new Set<string>()
  for (const m of months || []) {
    if (m >= 0 && m <= 2) qSet.add('Q1')
    else if (m >= 3 && m <= 5) qSet.add('Q2')
    else if (m >= 6 && m <= 8) qSet.add('Q3')
    else if (m >= 9 && m <= 11) qSet.add('Q4')
  }
  return Array.from(qSet).sort()
}

function flattenCorporateUniverse(nodes: CorporateUniverseCandidate[]): CorporateUniverseCandidate[] {
  const out: CorporateUniverseCandidate[] = []
  const visit = (list?: CorporateUniverseCandidate[]) => {
    if (!Array.isArray(list)) return
    for (const node of list) {
      if (!node) continue
      if (node.id && node.name) out.push(node)
      if (Array.isArray(node.children) && node.children.length > 0) {
        visit(node.children)
      }
    }
  }
  visit(nodes)
  return out
}

function fuzzyMatchByName<T extends { id: string; name: string }>(
  targetName: string,
  candidates: T[],
  fallbackToFirst = true
): T | undefined {
  if (!candidates.length) return undefined
  const norm = (s: string) => s.toLowerCase().replace(/[^a-z0-9]+/g, ' ').trim()
  const targetNorm = norm(targetName)
  if (!targetNorm) return fallbackToFirst ? candidates[0] : undefined

  const exact = candidates.find(c => norm(c.name) === targetNorm)
  if (exact) return exact

  const substr = candidates.find(
    c => norm(c.name).includes(targetNorm) || targetNorm.includes(norm(c.name))
  )
  if (substr) return substr

  const targetTokens = new Set(targetNorm.split(' ').filter(t => t.length > 2))
  let best: T | undefined
  let bestScore = 0
  for (const cand of candidates) {
    const tokens = norm(cand.name).split(' ').filter(t => t.length > 2)
    const overlap = tokens.filter(t => targetTokens.has(t)).length
    if (overlap > bestScore) {
      bestScore = overlap
      best = cand
    }
  }
  if (best) return best
  return fallbackToFirst ? candidates[0] : undefined
}

function extractRowRefCode(row: Record<string, any>, record: MappedRecordInput, idx: number): string {
  const raw =
    row.trx_ref_number ??
    row.loan_number ??
    row.contract_no ??
    row.contract_number ??
    row.incident_code ??
    row.voucher_number ??
    row.branch_code ??
    row.transaction_ref ??
    row.invoice_no ??
    (record.audit_field ? row[record.audit_field] : undefined) ??
    row.trx_id ??
    row.loan_id ??
    row.gl_id ??
    row.contract_id ??
    row.incident_id ??
    row.branch_id ??
    row.id
  const str = raw !== undefined && raw !== null ? String(raw).trim() : ''
  return str || `${(record.table_name || 'ROW').toUpperCase().slice(0, 4)}-${idx + 1}`
}

function extractRowDimensionName(row: Record<string, any>, record: MappedRecordInput, dimCol?: string | null): string {
  const raw =
    (dimCol ? row[dimCol] : undefined) ??
    row.branch_name ??
    row.entity_unit ??
    row.department ??
    row.reporting_unit ??
    row.vendor_name ??
    row.gl_account_name ??
    row.region ??
    row.city ??
    row.customer_name
  const str = raw !== undefined && raw !== null ? String(raw).trim() : ''
  return str || `Unit Operasional ${record.table_name}`
}

function resolveCanonicalCrpCategory(row: Record<string, any>, tableName: string): string {
  const rawCat = String(row.risk_category ?? row.category ?? '').toLowerCase()
  if (rawCat.includes('compli') || rawCat.includes('kepatuhan') || rawCat.includes('regulasi') || rawCat.includes('aml') || rawCat.includes('legal')) {
    return 'Compliance'
  }
  if (rawCat.includes('tech') || rawCat.includes('cyber') || rawCat.includes('it') || rawCat.includes('sistem') || rawCat.includes('akses')) {
    return 'Technology'
  }
  if (rawCat.includes('gov') || rawCat.includes('gcg') || rawCat.includes('tata kelola')) {
    return 'Governance'
  }
  if (rawCat.includes('hr') || rawCat.includes('sdm') || rawCat.includes('human')) {
    return 'Human Resources'
  }
  if (rawCat.includes('strateg')) {
    return 'Strategic'
  }
  if (rawCat.includes('finan') || rawCat.includes('keuangan') || rawCat.includes('kredit') || rawCat.includes('credit') || rawCat.includes('treasury')) {
    return 'Financial'
  }
  if (rawCat.includes('operas') || rawCat.includes('ops')) {
    return 'Operations'
  }

  const tbl = (tableName || '').toLowerCase()
  if (
    tbl === 'corporate_transactions' ||
    tbl === 'credit_loan_portfolio' ||
    tbl === 'loan_accounts' ||
    tbl === 'gl_journal_entries' ||
    tbl === 'gl_transactions'
  ) {
    return 'Financial'
  }
  if (
    tbl === 'vendor_contracts' ||
    tbl === 'vendor_invoices' ||
    tbl === 'audit_risk_incidents' ||
    tbl === 'compliance_incident_records'
  ) {
    return 'Compliance'
  }
  if (tbl === 'user_access_audit_logs') {
    return 'Technology'
  }
  return 'Operations'
}

const CRP_SCORE_MATRIX: Record<number, Record<number, number>> = {
  5: { 1: 7, 2: 12, 3: 17, 4: 22, 5: 25 },
  4: { 1: 4, 2: 9, 3: 14, 4: 19, 5: 24 },
  3: { 1: 3, 2: 8, 3: 13, 4: 18, 5: 23 },
  2: { 1: 2, 2: 6, 3: 11, 4: 16, 5: 21 },
  1: { 1: 1, 2: 5, 3: 10, 4: 15, 5: 20 },
}

export function buildRiskProfilePayloadsFromRows(
  record: MappedRecordInput,
  options?: {
    fiscalYear?: number
    locations?: Array<{ id: string; name: string }>
  }
): RiskProfileRowCandidate[] {
  const year = options?.fiscalYear || new Date().getFullYear()
  const locations = options?.locations || []
  const tbl = record.table_name
  const conn = record.connection_name
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const compat = evaluateTableModuleCompatibility(record, 'risk_profile')

  const matchLocation = (branchText: string): { id: string; name: string } | undefined => {
    if (!branchText || !locations.length) return undefined
    return fuzzyMatchByName(branchText, locations, false)
  }

  const makeAssessments = (likelihood: number, impact: number) => [
    {
      year,
      impact_q1: impact,
      impact_q2: impact,
      impact_q3: impact,
      impact_q4: impact,
      likelihood_q1: likelihood,
      likelihood_q2: likelihood,
      likelihood_q3: likelihood,
      likelihood_q4: likelihood,
    },
  ]

  const out: RiskProfileRowCandidate[] = []
  const seenNames = new Set<string>()

  if (rows.length > 0) {
    rows.forEach((row, idx) => {
      if (!row || typeof row !== 'object') return

      const refCode = extractRowRefCode(row, record, idx)
      const dimName = extractRowDimensionName(row, record, compat.dimensionColumn)
      const evalRes = isRowAnomalousOrFlagged(row, record.custom_rules, record.anomaly_rules)
      const category = resolveCanonicalCrpCategory(row, tbl)
      const amt = Number(
        row.amount ??
          row.plafond ??
          row.outstanding ??
          row.contract_value ??
          row.invoice_amount ??
          row.exposure_value ??
          row.debit_amount ??
          row.principal_amount ??
          0
      ) || 0

      let likelihood = 2
      let impact = 3
      if (evalRes.flagged) {
        if (evalRes.severity === 'High') {
          likelihood = amt >= 1_000_000_000 || Number(row.severity_score) >= 90 || Number(row.collectability) >= 5 ? 5 : 4
          impact = amt >= 500_000_000 || Number(row.severity_score) >= 85 || Number(row.collectability) >= 4 ? 5 : 4
        } else {
          likelihood = 3
          impact = 4
        }
      } else if (String(row.risk_tier || '').toUpperCase() === 'HIGH') {
        likelihood = 4
        impact = 4
      } else if (amt >= 250_000_000) {
        likelihood = 3
        impact = 3
      }

      const severity = CRP_SCORE_MATRIX[likelihood]?.[impact] || 13

      let baseName = ''
      if (tbl === 'corporate_transactions') {
        const sub = row.category || row.channel || dimName
        baseName = `Risiko Transaksi ${refCode} — ${sub} (${dimName})`
      } else if (tbl === 'credit_loan_portfolio' || tbl === 'loan_accounts') {
        baseName = `Risiko Kredit ${refCode} — ${dimName} (KOL ${row.collectability ?? row.kol_status ?? '-'})`
      } else if (tbl === 'gl_journal_entries' || tbl === 'gl_transactions') {
        baseName = `Risiko Jurnal GL ${refCode} — ${row.gl_account_name || row.gl_account_code || dimName}`
      } else if (tbl === 'vendor_contracts' || tbl === 'vendor_invoices') {
        baseName = `Risiko Kontrak ${refCode} — ${row.vendor_name || dimName}`
      } else if (tbl === 'audit_risk_incidents' || tbl === 'compliance_incident_records') {
        baseName = `Risiko Insiden ${refCode} — ${row.risk_category || dimName}`
      } else if (tbl === 'branches') {
        baseName = `Risiko Operasional Cabang ${refCode} — ${row.branch_name || row.city || dimName}`
      } else {
        baseName = `Risiko ${tbl} (${refCode}) — ${dimName}`
      }

      let cleanName = baseName.replace(/\s+/g, ' ').trim().slice(0, 100)
      if (seenNames.has(cleanName.toLowerCase())) {
        const suffix = ` #${idx + 1}`
        cleanName = `${cleanName.slice(0, 100 - suffix.length)}${suffix}`
      }
      seenNames.add(cleanName.toLowerCase())

      const amtText = amt > 0 ? ` Nilai/Eksposur: Rp ${amt.toLocaleString('id-ID')}.` : ''
      const reasonText = evalRes.reasons.length > 0
        ? ` Indikasi risiko: ${evalRes.reasons.join('; ')}.`
        : (row.notes || row.finding_summary || row.audit_note ? ` Catatan: ${row.notes || row.finding_summary || row.audit_note}.` : '')
      const statusText = row.authorization_status || row.approval_status || row.status
        ? ` Status: ${row.authorization_status || row.approval_status || row.status}.`
        : ''
      const description = `Risiko teridentifikasi dari baris data '${refCode}' pada unit '${dimName}' (pipeline '${conn}', tabel ${tbl}).${amtText}${statusText}${reasonText}`.trim()

      const matchedLoc = matchLocation(String(row.branch_name || row.entity_unit || dimName))

      out.push({
        refCode,
        name: cleanName,
        category,
        description,
        impact,
        likelihood,
        severity,
        ...(matchedLoc ? { location_id: matchedLoc.id, branch: matchedLoc.name } : {}),
        source: `${conn} / ${tbl}`,
        assessments: makeAssessments(likelihood, impact),
      })
    })
    return out
  }

  const dims = Array.isArray(record.profile?.dimension_breakdown) ? record.profile!.dimension_breakdown! : []
  if (dims.length > 0) {
    dims.forEach((d, idx) => {
      const refCode = `DIM-${idx + 1}`
      const likelihood = d.flagged_count >= 10 ? 4 : (d.flagged_count > 0 ? 3 : 2)
      const impact = d.max_amount >= 500_000_000 || d.flagged_count >= 5 ? 4 : 3
      const severity = CRP_SCORE_MATRIX[likelihood]?.[impact] || 13
      const category = resolveCanonicalCrpCategory({}, tbl)
      const cleanName = `Risiko ${tbl} — ${d.name}`.slice(0, 100)
      const matchedLoc = matchLocation(d.name)
      out.push({
        refCode,
        name: cleanName,
        category,
        description: `Risiko unit '${d.name}' dari pipeline '${conn}' (tabel ${tbl}): ${d.row_count.toLocaleString('id-ID')} baris (${d.flagged_count} baris berisiko).`,
        impact,
        likelihood,
        severity,
        ...(matchedLoc ? { location_id: matchedLoc.id, branch: matchedLoc.name } : {}),
        source: `${conn} / ${tbl}`,
        assessments: makeAssessments(likelihood, impact),
      })
    })
  }
  return out
}

export function buildAuditUniverseCandidatesFromTable(record: MappedRecordInput): AuditUniverseNodeCandidate[] {
  const tbl = record.table_name
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const compat = evaluateTableModuleCompatibility(record, 'audit_universe')
  const out: AuditUniverseNodeCandidate[] = []
  const seen = new Set<string>()

  if (rows.length > 0) {
    rows.forEach((row, idx) => {
      if (!row || typeof row !== 'object') return
      const refCode = extractRowRefCode(row, record, idx)
      const dimName = extractRowDimensionName(row, record, compat.dimensionColumn)
      const name = tbl === 'branches'
        ? `${row.branch_name || dimName} (${refCode})`
        : `${dimName} — ${refCode} (${tbl})`
      const key = name.trim().toLowerCase()
      if (!seen.has(key)) {
        seen.add(key)
        out.push({ name: name.trim(), dimensionName: dimName, refCode })
      }
    })
    return out
  }

  const dims = Array.isArray(record.profile?.dimension_breakdown) ? record.profile!.dimension_breakdown! : []
  dims.forEach((d, idx) => {
    if (!d?.name) return
    const name = `${d.name} (${tbl})`
    const key = name.trim().toLowerCase()
    if (!seen.has(key)) {
      seen.add(key)
      out.push({ name: name.trim(), dimensionName: d.name, refCode: `DIM-${idx + 1}` })
    }
  })
  return out
}

export function buildAnnualPlanDraftFromTable(
  record: MappedRecordInput,
  ctx: AnnualPlanAdapterContext
): AnnualPlanDraftCandidate {
  const year = ctx.year || new Date().getFullYear()
  const tbl = record.table_name
  const conn = record.connection_name
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const totalRows = Number(record.profile?.total_rows || record.total_rows || rows.length || 0)
  const compat = evaluateTableModuleCompatibility(record, 'annual_plan')
  const dimCol = compat.dimensionColumn
  const flatUniverse = flattenCorporateUniverse(ctx.corporateUniverse)

  const tableLabelMap: Record<string, string> = {
    corporate_transactions: 'Transaksi Korporasi & Otorisasi Cabang',
    credit_loan_portfolio: 'Portofolio Kredit & Kepatuhan Underwriting',
    gl_journal_entries: 'Jurnal Buku Besar (GL) & Kelengkapan Voucher',
    vendor_contracts: 'Kontrak Pengadaan & Manajemen Vendor',
    audit_risk_incidents: 'Tindak Lanjut Insiden Risiko & Kepatuhan',
    branches: 'Tata Kelola & Operasional Jaringan Cabang',
  }
  const readableScope = tableLabelMap[tbl] || `Pengendalian Data ${tbl}`

  const hasDimensionBreakdown =
    Array.isArray(record.profile?.dimension_breakdown) && record.profile!.dimension_breakdown!.length > 0
  const effectiveMode: 'per_row' | 'per_dimension' =
    ctx.mode || (rows.length > 0 ? 'per_row' : 'per_dimension')

  if (effectiveMode === 'per_row' && rows.length > 0) {
    const activities: AnnualPlanActivityCandidate[] = rows.map((r, idx) => {
      const refCode = extractRowRefCode(r, record, idx)
      const dimName = extractRowDimensionName(r, record, dimCol)
      const evalRes = isRowAnomalousOrFlagged(r, record.custom_rules, record.anomaly_rules)
      const amt = Number(r.amount ?? r.plafond ?? r.contract_value ?? r.exposure_value ?? r.debit_amount ?? 0) || 0

      const matchedUniverse = fuzzyMatchByName(dimName, flatUniverse, false)
      let matchedDept = matchedUniverse?.department_id
        ? ctx.departments.find(d => d.id === matchedUniverse.department_id)
        : undefined
      if (!matchedDept) {
        matchedDept = fuzzyMatchByName(dimName, ctx.departments, true) || ctx.departments[idx % Math.max(1, ctx.departments.length)]
      }
      const supervisor = ctx.supervisors.length > 0
        ? ctx.supervisors[idx % ctx.supervisors.length]!
        : ''

      const isHighRisk = evalRes.flagged || amt >= 1_000_000_000
      return {
        selected: true,
        dimensionName: `${dimName} (${refCode})`,
        name: `Audit ${readableScope} — ${dimName} [${refCode}]`.slice(0, 100),
        category: AuditCategory.ASSURANCE,
        departmentId: matchedDept?.id || '',
        departmentName: matchedDept?.name || '',
        auditUniverseId: matchedUniverse?.id || '',
        auditUniverseName: matchedUniverse?.name || '',
        riskLevel: evalRes.flagged ? 'High' : 'Moderate',
        supervisorName: supervisor,
        auditorCount: isHighRisk ? 3 : 2,
        totalMandays: isHighRisk ? 20 : 15,
        timelineText: 'Q1',
        rowCount: 1,
        flaggedCount: evalRes.flagged ? 1 : 0,
      }
    })

    return {
      title: `Sumber Data: ${conn} / ${tbl} (FY ${year})`,
      year: String(year),
      selectedMonths: [0, 1, 2],
      notes: `Draft PKAT otomatis dari pemetaan sumber data '${conn}' (tabel ${tbl}, total populasi ${totalRows.toLocaleString('id-ID')} baris, ${activities.length} objek audit per baris/transaksi).`,
      activities,
    }
  }

  let dimSummaries: TableDimensionSummary[] = []
  if (hasDimensionBreakdown) {
    dimSummaries = [...record.profile!.dimension_breakdown!]
  } else {
    const groups = new Map<string, TableDimensionSummary>()

    for (const r of rows) {
      if (!r || typeof r !== 'object') continue
      const dimName = extractRowDimensionName(r, record, dimCol)
      if (!groups.has(dimName)) {
        groups.set(dimName, {
          name: dimName,
          row_count: 0,
          flagged_count: 0,
          total_amount: 0,
          max_amount: 0,
        })
      }
      const g = groups.get(dimName)!
      g.row_count++
      if (isRowAnomalousOrFlagged(r, record.custom_rules, record.anomaly_rules).flagged) {
        g.flagged_count++
      }
      const amt = Number(r.amount ?? r.plafond ?? r.contract_value ?? r.exposure_value ?? r.debit_amount ?? 0) || 0
      g.total_amount += amt
      if (amt > g.max_amount) g.max_amount = amt
    }

    dimSummaries = Array.from(groups.values())
      .sort((a, b) => b.flagged_count - a.flagged_count || b.row_count - a.row_count)
  }

  if (dimSummaries.length === 0) {
    dimSummaries = [
      {
        name: `Unit Operasional ${tbl}`,
        row_count: totalRows || 1,
        flagged_count: Number(record.profile?.flagged_rows || 0),
        total_amount: Number(record.profile?.total_amount || 0),
        max_amount: Number(record.profile?.max_amount || 0),
      },
    ]
  }

  const activities: AnnualPlanActivityCandidate[] = dimSummaries.map((dim, idx) => {
    const matchedUniverse = fuzzyMatchByName(dim.name, flatUniverse, false)
    let matchedDept = matchedUniverse?.department_id
      ? ctx.departments.find(d => d.id === matchedUniverse.department_id)
      : undefined
    if (!matchedDept) {
      matchedDept = fuzzyMatchByName(dim.name, ctx.departments, true) || ctx.departments[idx % Math.max(1, ctx.departments.length)]
    }
    const supervisor = ctx.supervisors.length > 0
      ? ctx.supervisors[idx % ctx.supervisors.length]!
      : ''

    const isHighRisk = dim.flagged_count >= 5 || dim.max_amount >= 1_000_000_000
    return {
      selected: true,
      dimensionName: dim.name,
      name: `Audit ${readableScope} — ${dim.name}`,
      category: AuditCategory.ASSURANCE,
      departmentId: matchedDept?.id || '',
      departmentName: matchedDept?.name || '',
      auditUniverseId: matchedUniverse?.id || '',
      auditUniverseName: matchedUniverse?.name || '',
      riskLevel: '',
      supervisorName: supervisor,
      auditorCount: isHighRisk ? 3 : 2,
      totalMandays: isHighRisk ? 20 : 15,
      timelineText: 'Q1',
      rowCount: dim.row_count,
      flaggedCount: dim.flagged_count,
    }
  })

  return {
    title: `Sumber Data: ${conn} / ${tbl} (FY ${year})`,
    year: String(year),
    selectedMonths: [0, 1, 2],
    notes: `Draft PKAT otomatis dari pemetaan sumber data '${conn}' (tabel ${tbl}, total populasi ${totalRows.toLocaleString('id-ID')} baris, ${activities.length} objek audit).`,
    activities,
  }
}

function parseHourFromRow(row: Record<string, any>, fallbackIdx: number): { dateStr: string; timestampStr: string; hour: number } {
  const rawTs =
    row.trx_date ??
    row.timestamp ??
    row.logged_at ??
    row.reported_at ??
    row.created_at ??
    row.updated_at ??
    row.gl_date ??
    row.signed_at ??
    row.disbursed_at

  if (rawTs) {
    const str = String(rawTs).trim()
    const timeMatch = str.match(/(\d{2}):(\d{2})(?::(\d{2}))?/)
    const dateMatch = str.match(/(\d{4}-\d{2}-\d{2})/)
    const dateStr = dateMatch ? dateMatch[1]! : new Date().toISOString().slice(0, 10)
    if (timeMatch) {
      const h = parseInt(timeMatch[1]!, 10)
      const m = parseInt(timeMatch[2]!, 10)
      const hourFloat = Math.min(23.9, Math.max(0, +(h + m / 60).toFixed(1)))
      return {
        dateStr,
        timestampStr: `${dateStr} ${timeMatch[0]}`,
        hour: hourFloat,
      }
    }
    return {
      dateStr,
      timestampStr: `${dateStr} 10:00:00`,
      hour: 9 + (fallbackIdx % 8),
    }
  }

  const today = new Date().toISOString().slice(0, 10)
  return {
    dateStr: today,
    timestampStr: `${today} 10:00:00`,
    hour: 9 + (fallbackIdx % 8),
  }
}

export function buildRealTableAnomalies(record: MappedRecordInput): {
  targetCategory: string
  anomalies: any[]
  normalPoints: Array<{ x: number; y: number; type: string; isAnomaly: boolean; label: string }>
  totalPopulation: number
} {
  const targetCategory = getCanonicalCategory(record.table_name)
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const totalPopulation = Number(record.profile?.total_rows || record.total_rows || rows.length || 0)

  // Compute mean & std of amounts in sample rows for outlier fallback when no explicit flag column triggers
  const rowAmounts = rows.map(r =>
    Number(
      r?.amount ??
        r?.plafond ??
        r?.outstanding ??
        r?.contract_value ??
        r?.invoice_amount ??
        r?.exposure_value ??
        r?.debit_amount ??
        r?.principal_amount ??
        r?.severity_score ??
        0
    ) || 0
  )
  const validAmounts = rowAmounts.filter(a => a > 0)
  const meanAmt = validAmounts.length
    ? validAmounts.reduce((s, a) => s + a, 0) / validAmounts.length
    : 0
  const variance = validAmounts.length
    ? validAmounts.reduce((s, a) => s + Math.pow(a - meanAmt, 2), 0) / validAmounts.length
    : 0
  const stdAmt = Math.sqrt(variance)
  const outlierThreshold = meanAmt + 1.25 * stdAmt

  const anomalies: any[] = []
  const normalPoints: Array<{ x: number; y: number; type: string; isAnomaly: boolean; label: string }> = []

  rows.forEach((row, idx) => {
    if (!row || typeof row !== 'object') return

    const evalRes = isRowAnomalousOrFlagged(row, record.custom_rules, record.anomaly_rules)
    const rawAmt = rowAmounts[idx] || 0
    const isStatisticalOutlier = !evalRes.flagged && validAmounts.length >= 4 && rawAmt > outlierThreshold && outlierThreshold > 0

    const pkVal = String(
      row.trx_ref_number ??
        row.loan_number ??
        row.contract_no ??
        row.incident_code ??
        row.voucher_number ??
        row.transaction_ref ??
        row.invoice_no ??
        row[record.audit_field] ??
        row.trx_id ??
        row.loan_id ??
        row.gl_id ??
        row.contract_id ??
        row.incident_id ??
        row.id ??
        `${record.table_name.toUpperCase().slice(0, 4)}-${idx + 1}`
    )

    const entityVal = String(
      row.branch_name ??
        row.entity_unit ??
        row.vendor_name ??
        row.customer_name ??
        row.gl_account_name ??
        row.reporting_unit ??
        `${record.connection_name} (${record.table_name})`
    )

    const accountVal = String(
      row.account_number ??
        row.gl_account_code ??
        row.po_number ??
        row.loan_number ??
        row.account_type ??
        row.vendor_code ??
        pkVal
    )

    const xMetric = rawAmt >= 1_000_000
      ? +(rawAmt / 1_000_000).toFixed(2)
      : +(rawAmt || Number(row.severity_score) || idx + 1).toFixed(2)

    const { dateStr, timestampStr, hour } = parseHourFromRow(row, idx)

    if (evalRes.flagged || isStatisticalOutlier) {
      const reasons = evalRes.flagged
        ? evalRes.reasons
        : [`Deviasi nilai statistik Rp ${rawAmt.toLocaleString('id-ID')} melampaui rata-rata sampel pada ${record.table_name}`]
      const desc = reasons.join('; ')
      const riskLevel = normalizeAnomalyRiskLevel(
        row.risk_level ?? row.severity ?? row.severity_score ?? evalRes.severity
      )
      const score = evalRes.severity === 'High' ? 0.94 : 0.87

      anomalies.push({
        id: `DB-${record.table_name.slice(0, 4).toUpperCase()}-${pkVal}`,
        entity: entityVal,
        refNumber: pkVal,
        accountNumber: accountVal,
        amount: rawAmt,
        type: targetCategory,
        category: targetCategory,
        anomalyScore: score,
        description: `${desc} (Tabel: ${record.table_name}, Ref: ${pkVal})`,
        reason: desc,
        date: dateStr,
        timestamp: timestampStr,
        hour,
        hour_of_day: hour,
        riskLevel,
        severity: riskLevel,
        xMetric,
        isAnomaly: true,
        sourceTable: record.table_name,
      })
    } else {
      normalPoints.push({
        x: xMetric,
        y: hour,
        type: targetCategory,
        isAnomaly: false,
        label: pkVal,
      })
    }
  })

  return {
    targetCategory,
    anomalies,
    normalPoints,
    totalPopulation,
  }
}

export interface WorkingPaperSamplePayload {
  population: string
  sampleSize: number
  conclusion: string
  flaggedCount: number
  samples: Array<{
    id: number
    fieldworkDocument: string
    document: string
    step1: string
    l1: 'Pass' | 'Fail' | 'N/A'
    step2: string
    l2: 'Pass' | 'Fail' | 'N/A'
    step3: string
    l3: 'Pass' | 'Fail' | 'N/A'
  }>
}

export function buildWorkingPaperSamplesFromTable(record: MappedRecordInput): WorkingPaperSamplePayload {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const totalPop = Number(record.total_rows || record.profile?.total_rows || rows.length || 0)
  const popStr = totalPop.toLocaleString('id-ID')
  const signals = analyzeTableSchema({
    tableName: record.table_name,
    rowCount: totalPop,
    profile: record.profile,
    sample_records: rows,
  })

  // Prioritize flagged/anomalous rows first so high-risk evidence appears at the top of F-03
  const evaluatedRows = rows.map((row, idx) => ({
    row,
    idx,
    evalRes: isRowAnomalousOrFlagged(row, record.custom_rules, record.anomaly_rules),
  }))
  evaluatedRows.sort((a, b) => Number(b.evalRes.flagged) - Number(a.evalRes.flagged))

  let flaggedCount = 0
  const baseTs = Date.now()

  const samples = evaluatedRows.map(({ row, idx, evalRes }, outIdx) => {
    if (evalRes.flagged) flaggedCount++

    const refVal = String(
      row.trx_ref_number ||
        row.loan_number ||
        row.voucher_number ||
        row.contract_no ||
        row.contract_number ||
        row.incident_code ||
        row.invoice_no ||
        (signals.refColumn ? row[signals.refColumn] : '') ||
        row[record.audit_field] ||
        row.id ||
        `SMP-${String(idx + 1).padStart(3, '0')}`
    )

    const entityVal = String(
      (signals.dimensionColumn ? row[signals.dimensionColumn] : '') ||
        row.branch_name ||
        row.entity_unit ||
        row.vendor_name ||
        row.customer_name ||
        ''
    ).trim()

    const rawAmt = Number(
      (signals.amountColumn ? row[signals.amountColumn] : 0) ||
        row.amount ||
        row.plafond ||
        row.debit_amount ||
        row.contract_value ||
        row.exposure_value ||
        row.invoice_amount ||
        0
    )

    const amtPart = Number.isFinite(rawAmt) && rawAmt > 0 ? ` — Rp ${rawAmt.toLocaleString('id-ID')}` : ''
    const entityPart = entityVal ? ` [${entityVal}]` : ''
    const isAuthIssue =
      String(row.authorization_status || row.approval_status || '').toUpperCase().includes('OVERRIDE') ||
      String(row.authorization_status || row.approval_status || '').toUpperCase().includes('PENDING') ||
      row.is_override_auth === true ||
      ('voucher_number' in row && (row.voucher_number === null || String(row.voucher_number).trim() === ''))
    const isLimitOrMetricIssue =
      evalRes.flagged &&
      (rawAmt >= 450_000_000 ||
        Number(row.collectability) >= 3 ||
        Number(row.days_past_due) > 90 ||
        Number(row.severity_score) >= 75)

    return {
      id: baseTs + outIdx,
      fieldworkDocument: refVal,
      document: `Bukti ${record.table_name} (${refVal})${entityPart}${amtPart}`,
      step1: 'Verifikasi Otorisasi & Kelengkapan Dokumen Pendukung',
      l1: (isAuthIssue ? 'Fail' : 'Pass') as 'Pass' | 'Fail' | 'N/A',
      step2: 'Kesesuaian Limit, Nilai Nominal & Rekonsiliasi Saldo',
      l2: (isLimitOrMetricIssue ? 'Fail' : 'Pass') as 'Pass' | 'Fail' | 'N/A',
      step3: 'Validasi Kepatuhan & Indikator Anomali Transaksi',
      l3: (evalRes.flagged ? 'Fail' : 'Pass') as 'Pass' | 'Fail' | 'N/A',
    }
  })

  const conclusion =
    flaggedCount > 0
      ? `Pengujian substantif pada ${samples.length} sampel uji petik dari total populasi nyata ${popStr} baris tabel ${record.table_name} menemukan ${flaggedCount} sampel berindikasi deviasi/anomali yang memerlukan tindak lanjut audit.`
      : `Pengujian substantif validasi otorisasi dan kepatuhan pada ${samples.length} sampel uji petik dari total populasi nyata ${popStr} baris tabel ${record.table_name} tidak menemukan deviasi material.`

  return {
    population: `${popStr} Baris (${record.table_name})`,
    sampleSize: samples.length,
    conclusion,
    flaggedCount,
    samples,
  }
}

const CANONICAL_ENTITY_PAIRS: Record<string, [number, number]> = {
  HIGH: [4.7, 4.8],
  MODERATE_HIGH: [3.8, 4.0],
  MODERATE: [3.0, 3.2],
  LOW_MODERATE: [2.2, 2.4],
  LOW: [1.2, 1.4],
}

const CANONICAL_AUDITEE_PAIRS: Record<string, [number, number]> = {
  HIGH: [5, 4],
  MODERATE_HIGH: [4, 4],
  MODERATE: [3, 3],
  LOW_MODERATE: [2, 2],
  LOW: [1, 1],
}

function calc5TierCode(lik: number, imp: number): string {
  const m: Record<number, Record<number, string>> = {
    5: { 1: 'LOW_MODERATE', 2: 'MODERATE', 3: 'MODERATE_HIGH', 4: 'HIGH', 5: 'HIGH' },
    4: { 1: 'LOW', 2: 'LOW_MODERATE', 3: 'MODERATE', 4: 'MODERATE_HIGH', 5: 'HIGH' },
    3: { 1: 'LOW', 2: 'LOW_MODERATE', 3: 'MODERATE', 4: 'MODERATE_HIGH', 5: 'HIGH' },
    2: { 1: 'LOW', 2: 'LOW_MODERATE', 3: 'LOW_MODERATE', 4: 'MODERATE_HIGH', 5: 'HIGH' },
    1: { 1: 'LOW', 2: 'LOW', 3: 'LOW_MODERATE', 4: 'MODERATE', 5: 'HIGH' },
  }
  const lInt = Math.min(5, Math.max(1, Math.round(Number(lik) || 3)))
  const iInt = Math.min(5, Math.max(1, Math.round(Number(imp) || 3)))
  return m[lInt]?.[iInt] || 'MODERATE'
}

function extractExtRiskRaw(r: any) {
  return (
    r?.risk_level ??
    r?.riskLevel ??
    r?.risk_rating ??
    r?.risk_tier ??
    r?.severity ??
    r?.tingkat_risiko ??
    r?.severity_level ??
    r?.status_risiko ??
    r?.kolektibilitas ??
    r?.kol_status ??
    r?.rating
  )
}

export function buildEntityPredictionItemsFromTable(record: MappedRecordInput): any[] {
  const mapCategoryName = (rawCat?: string): string => {
    const c = String(rawCat || '').toLowerCase()
    if (c.includes('finan') || c.includes('keuangan') || c.includes('credit') || c.includes('kredit') || c.includes('lending') || c.includes('treasury')) return 'Financial'
    if (c.includes('tech') || c.includes('cyber') || c.includes('it') || c.includes('akses') || c.includes('sistem')) return 'Technology'
    if (c.includes('compli') || c.includes('kepatuhan') || c.includes('regulasi') || c.includes('aml') || c.includes('sop')) return 'Compliance'
    if (c.includes('govern') || c.includes('gcg') || c.includes('tata kelola') || c.includes('strateg')) return 'Governance'
    if (c.includes('hr') || c.includes('sdm') || c.includes('human')) return 'Human Resources'
    return 'Operational'
  }

  let defaultEntityName = `Unit Audit ${record.table_name.replace(/_/g, ' ')}`
  let defaultEntityType: 'Branch' | 'Auditable Entity' = 'Auditable Entity'
  let defaultCategory = 'Operational'
  let defaultLik = 3.5
  let defaultImp = 4.0

  if (record.table_name === 'audit_risk_incidents' || record.table_name === 'compliance_incident_records') {
    defaultEntityName = 'Divisi Manajemen Risiko & Kepatuhan (Audit Incidents)'
    defaultCategory = 'Compliance'
    defaultLik = 4.2
    defaultImp = 4.5
  } else if (record.table_name === 'corporate_transactions' || record.table_name === 'gl_transactions' || record.table_name === 'gl_journal_entries') {
    defaultEntityName = 'Divisi Operasional Transaksi & Treasury'
    defaultCategory = 'Financial'
    defaultLik = 3.8
    defaultImp = 4.2
  } else if (record.table_name === 'credit_loan_portfolio' || record.table_name === 'loan_accounts') {
    defaultEntityName = 'Divisi Kredit Komersial & Korporasi'
    defaultEntityType = 'Branch'
    defaultCategory = 'Financial'
    defaultLik = 4.0
    defaultImp = 4.4
  } else if (record.table_name === 'vendor_contracts' || record.table_name === 'vendor_invoices') {
    defaultEntityName = 'Divisi Pengadaan & Manajemen Vendor (Procurement)'
    defaultCategory = 'Operational'
    defaultLik = 3.5
    defaultImp = 3.8
  }

  const customItems: any[] = []
  const seenEntities = new Set<string>()
  const sampleRows = Array.isArray(record.sample_records) ? record.sample_records : []

  if (sampleRows.length > 0) {
    const entityGroups = new Map<string, any[]>()
    for (const row of sampleRows) {
      const candEntity = (
        row.entity_unit ||
        row.branch_name ||
        row.entity ||
        row.department ||
        row.reporting_unit ||
        (row.branch_code ? `Cabang ${row.branch_code}` : '') ||
        row.vendor_name ||
        row.customer_name ||
        ''
      ).toString().trim()
      if (!candEntity) continue
      const key = candEntity.toLowerCase()
      if (!entityGroups.has(key)) entityGroups.set(key, [])
      entityGroups.get(key)!.push(row)
    }

    for (const [, rows] of entityGroups.entries()) {
      const first = rows[0]
      const entName = String(
        first.entity_unit ||
        first.branch_name ||
        first.entity ||
        first.department ||
        first.reporting_unit ||
        (first.branch_code ? `Cabang ${first.branch_code}` : '') ||
        first.vendor_name ||
        first.customer_name
      ).trim()
      if (!entName || seenEntities.has(entName.toLowerCase())) continue
      seenEntities.add(entName.toLowerCase())

      const isBr = ['cabang', 'kcp', 'regional', 'kc', 'outlet', 'branch', 'kcu'].some(w => entName.toLowerCase().includes(w))
      const cat = first.risk_category ? mapCategoryName(first.risk_category) : defaultCategory

      const avgSeverity = rows.reduce((acc, r) => acc + (Number(r.severity_score) || 0), 0) / rows.length
      const maxExposure = Math.max(...rows.map(r => Number(r.exposure_value || r.financial_loss_idr || r.plafond || r.amount || r.contract_value || 0)))
      const explicitRowWithLevel = rows.find((r) => {
        const val = extractExtRiskRaw(r)
        return val !== undefined && val !== null && String(val).trim() !== ''
      })
      const explicitRawLevel = explicitRowWithLevel ? extractExtRiskRaw(explicitRowWithLevel) : null
      const normalizedExternalCode = explicitRawLevel !== null
        ? toEntityRiskCode(normalizeEntityRiskLevel(explicitRawLevel))
        : null

      let lik = normalizedExternalCode ? CANONICAL_ENTITY_PAIRS[normalizedExternalCode]![0] : defaultLik
      let imp = normalizedExternalCode ? CANONICAL_ENTITY_PAIRS[normalizedExternalCode]![1] : defaultImp
      if (avgSeverity > 0) {
        lik = Math.min(5, Math.max(1, +(avgSeverity / 20).toFixed(1)))
        imp = Math.min(5, Math.max(1, +((avgSeverity / 20) + (maxExposure > 500_000_000 ? 0.4 : 0.1)).toFixed(1)))
      } else if (maxExposure > 0 && !normalizedExternalCode) {
        lik = maxExposure > 2_000_000_000 ? 4.3 : (maxExposure > 500_000_000 ? 3.7 : 3.0)
        imp = maxExposure > 2_000_000_000 ? 4.6 : (maxExposure > 500_000_000 ? 4.0 : 3.2)
      }

      const currentRiskCode = normalizedExternalCode || calc5TierCode(lik, imp)

      customItems.push({
        entity: entName,
        entity_type: isBr ? 'Branch' : 'Auditable Entity',
        risk_category: cat,
        target_timeline: `FY ${new Date().getFullYear()}`,
        inherent_likelihood: lik,
        inherent_impact: imp,
        audit_findings_count: Math.max(2, rows.length),
        kpi_below_target_count: avgSeverity >= 75 || currentRiskCode === 'HIGH' ? 2 : 1,
        kpi_volatility: 0.18,
        previous_risk_score: +(lik * imp).toFixed(1),
        assessment_month: new Date().getMonth() + 1,
        current_risk_level: currentRiskCode,
        source_id: `external_db:${record.connection_name}:${record.table_name}`,
      })
    }
  }

  if (customItems.length === 0) {
    customItems.push({
      entity: defaultEntityName,
      entity_type: defaultEntityType,
      risk_category: defaultCategory,
      target_timeline: `FY ${new Date().getFullYear()}`,
      inherent_likelihood: defaultLik,
      inherent_impact: defaultImp,
      audit_findings_count: sampleRows.length || 3,
      kpi_below_target_count: 1,
      kpi_volatility: 0.15,
      previous_risk_score: +(defaultLik * defaultImp).toFixed(1),
      assessment_month: new Date().getMonth() + 1,
      current_risk_level: calc5TierCode(defaultLik, defaultImp),
      source_id: `external_db:${record.connection_name}:${record.table_name}`,
    })
  }

  return customItems
}

export interface NlpDocumentCandidate {
  title: string
  text: string
  source: string
  sourceType: string
  fileName: string
  fileType: string
  riskCategory?: string
  riskLevel?: string
}

export function buildNlpDocumentsFromTable(record: MappedRecordInput): NlpDocumentCandidate[] {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const sortedRows = [...rows].sort((a: any, b: any) => {
    const aRisk = (a.risk_flag || a.is_flagged || a.is_suspicious || a.is_split_contract || a.is_anomaly || a.compliance_status === 'NON_COMPLIANT' || Number(a.severity_score) >= 75) ? 1 : 0
    const bRisk = (b.risk_flag || b.is_flagged || b.is_suspicious || b.is_split_contract || b.is_anomaly || b.compliance_status === 'NON_COMPLIANT' || Number(b.severity_score) >= 75) ? 1 : 0
    return bRisk - aRisk
  })

  const riskRatingCols = new Set(['risk_level', 'risklevel', 'severity', 'risk_rating', 'tingkat_risiko', 'severity_level'])
  const docs: NlpDocumentCandidate[] = []

  for (let i = 0; i < sortedRows.length; i++) {
    const r = sortedRows[i] || {}
    let docTitle = ''
    let docText = ''
    let riskCat: string | undefined
    const rawExtRisk = extractExtRiskRaw(r) ?? (r.severity_score !== undefined && r.severity_score !== null ? r.severity_score : undefined)
    const hasExtRisk = rawExtRisk !== undefined && rawExtRisk !== null && String(rawExtRisk).trim() !== '' && !['not scored', 'none', 'nan', 'null'].includes(String(rawExtRisk).trim().toLowerCase())
    const extRiskCode = hasExtRisk ? toNlpRiskCode(rawExtRisk) : undefined

    if (record.table_name === 'vendor_contracts' || r.contract_number || r.contract_no || r.vendor_name) {
      const cNo = r.contract_number || r.contract_no || r.contract_id || `CTR-${i + 1}`
      const vName = r.vendor_name || 'Vendor Rekanan'
      const cTitle = r.contract_title || r.procurement_category || r.branch_name || 'Kontrak Pengadaan'
      const cVal = Number(r.contract_value || r.invoice_amount || 0).toLocaleString('id-ID')
      const method = r.procurement_method || r.approval_status || 'Direct Appointment'
      const sla = r.sla_penalty_clause || 'Klausul SLA dan denda keterlambatan perlu diverifikasi.'
      const notes = r.audit_notes || r.audit_note || (r.is_split_contract || r.is_flagged ? 'Ditemukan indikasi pemecahan nilai kontrak (split contract) untuk menghindari batas lelang terbuka.' : 'Tinjauan kepatuhan klausul kontrak pengadaan vendor.')
      docTitle = `Kontrak ${cNo} — ${vName} (${cTitle})`
      docText = `Temuan Audit Kontrak Pengadaan ${cNo} pada ${vName} (${cTitle}) senilai Rp ${cVal} melalui status ${method}. ${notes} Klausul SLA: ${sla} Status kepatuhan: ${r.compliance_status || r.approval_status || 'PERLU TINJAUAN'}.`
      riskCat = r.is_split_contract || r.is_flagged || String(notes).toLowerCase().includes('split') || String(notes).toLowerCase().includes('fiktif')
        ? 'Financial'
        : (String(cTitle).toLowerCase().includes('it') || String(cTitle).toLowerCase().includes('cloud') || String(cTitle).toLowerCase().includes('server') ? 'Technology' : 'Operational')
    } else if (record.table_name === 'audit_risk_incidents' || r.incident_title || r.incident_description || r.finding_summary) {
      const iCode = r.incident_code || r.incident_id || `INC-${i + 1}`
      const unit = r.entity_unit || r.branch_name || 'Unit Operasional'
      docTitle = `${iCode}: ${r.incident_title || r.risk_category || 'Insiden Risiko Audit'} (${unit})`
      docText = `Temuan Insiden ${iCode} pada ${unit}: ${r.incident_description || r.finding_summary || r.incident_title || ''}. Akar masalah: ${r.root_cause || 'Kelemahan pengendalian internal'}. Kelemahan kontrol: ${r.control_weakness || 'Verifikasi berjenjang belum optimal'}.`
      riskCat = r.risk_category || undefined
    } else {
      const pkVal = r[record.audit_field] || r.trx_ref_number || r.loan_number || r.voucher_number || r.id || `REC-${i + 1}`
      const summaryPairs = Object.entries(r)
        .map(([k, v]) => {
          const formattedVal = riskRatingCols.has(k.toLowerCase()) && v != null && String(v).trim() !== ''
            ? normalizeNlpRiskLevel(v)
            : v
          return `${k}: ${formattedVal}`
        })
        .join(', ')
      docTitle = `Rekaman Audit ${record.table_name} (${pkVal})`
      docText = `Temuan evaluasi data eksternal dari tabel ${record.table_name} (${record.connection_name}) untuk rekaman ${pkVal}. ${r.notes || r.description || ''} Detail atribut: ${summaryPairs}.`
    }

    docs.push({
      title: docTitle,
      text: docText,
      source: `${record.connection_name} (${record.table_name})`,
      sourceType: 'External Database',
      fileName: `${record.table_name}_${i + 1}.sql`,
      fileType: 'DB_RECORD',
      riskCategory: riskCat,
      ...(extRiskCode ? { riskLevel: extRiskCode } : {}),
    })
  }

  return docs
}

export function buildKpiForecastItemsFromTable(record: MappedRecordInput): any[] {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const nowTime = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
  const totalPop = Number(record.total_rows || record.profile?.total_rows || rows.length || 0)

  const resolveRisk = (candidateRows: any[], fallbackCode: string) => {
    const rowWithRisk = candidateRows.find((r) => {
      const val = extractExtRiskRaw(r)
      return val !== undefined && val !== null && String(val).trim() !== ''
    })
    const raw = rowWithRisk ? extractExtRiskRaw(rowWithRisk) : null
    const riskLevel = toEntityRiskCode(raw !== null ? raw : fallbackCode)
    const [impactScore, likelihoodScore] = CANONICAL_AUDITEE_PAIRS[riskLevel] || [3, 3]
    return { riskLevel, impactScore, likelihoodScore }
  }

  const generatedKpis: any[] = []

  if (
    record.table_name === 'credit_loan_portfolio' ||
    record.table_name === 'loan_accounts' ||
    rows.some((r: any) => r.plafond !== undefined || r.collectability !== undefined || r.days_past_due !== undefined)
  ) {
    const totalLoans = Math.max(1, rows.length)
    const nplRows = rows.filter((r: any) => Number(r.collectability) >= 3 || Number(r.days_past_due) > 90 || String(r.kol_status || '').includes('3') || String(r.kol_status || '').includes('4') || String(r.kol_status || '').includes('5'))
    const breachRows = rows.filter((r: any) => r.is_policy_breach === true || r.rate_within_policy === false || Number(r.ltv_ratio) > 85 || Number(r.plafond) > 5_000_000_000)
    const avgRate = rows.reduce((acc: number, r: any) => acc + (Number(r.interest_rate) || 9.5), 0) / totalLoans

    const nplRatio = +((nplRows.length / totalLoans) * 100).toFixed(2) || 4.85
    const foreNpl = +(nplRatio * 1.18).toFixed(2)
    const nplDelta = +(((foreNpl - nplRatio) / Math.max(0.1, nplRatio)) * 100).toFixed(1)

    const breachRatio = +((breachRows.length / totalLoans) * 100).toFixed(2) || 12.5
    const foreBreach = +(breachRatio * 1.14).toFixed(2)
    const breachDelta = +(((foreBreach - breachRatio) / Math.max(0.1, breachRatio)) * 100).toFixed(1)

    const cleanAvgRate = +avgRate.toFixed(2)
    const foreRate = +(cleanAvgRate * 0.96).toFixed(2)
    const rateDelta = +(((foreRate - cleanAvgRate) / Math.max(0.1, cleanAvgRate)) * 100).toFixed(1)

    const nplRisk = resolveRisk(nplRows.length ? nplRows : rows, foreNpl >= 5 ? 'HIGH' : (foreNpl >= 3 ? 'MODERATE_HIGH' : 'MODERATE'))
    const breachRisk = resolveRisk(breachRows, foreBreach >= 10 ? 'HIGH' : (foreBreach >= 5 ? 'MODERATE_HIGH' : 'MODERATE'))
    const rateRisk = resolveRisk(rows, 'MODERATE_HIGH')

    generatedKpis.push(
      {
        kpiName: `Rasio Kredit Bermasalah / NPL (${record.table_name})`,
        code: 'KPI-EXT-NPL',
        unit: '%',
        entity: `Portofolio Kredit (${record.connection_name})`,
        entityType: 'Branch',
        syncFrequency: `Live DB (${record.table_name})`,
        lastSyncedAt: nowTime,
        targetHorizon: 'Q4 2026',
        currentValue: nplRatio,
        forecastedValue: foreNpl,
        deltaPct: nplDelta,
        trend: 'Deteriorating',
        impactScore: nplRisk.impactScore,
        likelihoodScore: nplRisk.likelihoodScore,
        riskLevel: nplRisk.riskLevel,
        aiNarrative: `Berdasarkan analisis populasi ${totalPop || totalLoans} fasilitas kredit pada ${record.table_name}, rasio NPL (Kolektibilitas 3–5) saat ini sebesar ${nplRatio}% dan diproyeksikan naik menjadi ${foreNpl}% pada kuartal depan akibat debitur DPD > 90 hari.`,
        recommendedAction: 'Lakukan audit khusus restrukturisasi kredit pada cabang dengan konsentrasi KOL 3–5 tertinggi dan perketat early warning system koleksi kredit.',
        lastActualPeriod: '2026-Q2',
        history: [
          { period: '2025-Q3', value: +(nplRatio * 0.82).toFixed(2) },
          { period: '2025-Q4', value: +(nplRatio * 0.89).toFixed(2) },
          { period: '2026-Q1', value: +(nplRatio * 0.95).toFixed(2) },
          { period: '2026-Q2', value: nplRatio },
        ],
        forecast: [
          { period: '2026-Q3', value: +(nplRatio * 1.09).toFixed(2), lower: +(nplRatio * 1.02).toFixed(2), upper: +(nplRatio * 1.16).toFixed(2) },
          { period: '2026-Q4', value: foreNpl, lower: +(foreNpl * 0.93).toFixed(2), upper: +(foreNpl * 1.08).toFixed(2) },
        ],
      },
      {
        kpiName: `Rasio Pelanggaran Kebijakan Kredit & LTV (${record.table_name})`,
        code: 'KPI-EXT-LTV',
        unit: '%',
        entity: `Divisi Risiko Kredit (${record.connection_name})`,
        entityType: 'Department',
        syncFrequency: `Live DB (${record.table_name})`,
        lastSyncedAt: nowTime,
        targetHorizon: 'Q4 2026',
        currentValue: breachRatio,
        forecastedValue: foreBreach,
        deltaPct: breachDelta,
        trend: 'Deteriorating',
        impactScore: breachRisk.impactScore,
        likelihoodScore: breachRisk.likelihoodScore,
        riskLevel: breachRisk.riskLevel,
        aiNarrative: `Terdapat ${breachRows.length} dari ${totalLoans} sampel uji petik kredit (${breachRatio}%) yang melampaui batas plafon/LTV atau suku bunga kebijakan, diproyeksikan meningkat ke ${foreBreach}% bila kontrol otorisasi tidak diperketat.`,
        recommendedAction: 'Wajibkan persetujuan Komite Kredit berjenjang untuk setiap fasilitas dengan LTV > 80% atau deviasi suku bunga khusus.',
        lastActualPeriod: '2026-Q2',
        history: [
          { period: '2025-Q3', value: +(breachRatio * 0.85).toFixed(2) },
          { period: '2025-Q4', value: +(breachRatio * 0.90).toFixed(2) },
          { period: '2026-Q1', value: +(breachRatio * 0.96).toFixed(2) },
          { period: '2026-Q2', value: breachRatio },
        ],
        forecast: [
          { period: '2026-Q3', value: +(breachRatio * 1.07).toFixed(2), lower: +(breachRatio * 1.01).toFixed(2), upper: +(breachRatio * 1.13).toFixed(2) },
          { period: '2026-Q4', value: foreBreach, lower: +(foreBreach * 0.94).toFixed(2), upper: +(foreBreach * 1.06).toFixed(2) },
        ],
      },
      {
        kpiName: `Yield Rata-rata Suku Bunga Kredit Portofolio (${record.table_name})`,
        code: 'KPI-EXT-YLD',
        unit: '%',
        entity: `Divisi Kredit Komersial (${record.connection_name})`,
        entityType: 'Department',
        syncFrequency: `Live DB (${record.table_name})`,
        lastSyncedAt: nowTime,
        targetHorizon: 'Q4 2026',
        currentValue: cleanAvgRate,
        forecastedValue: foreRate,
        deltaPct: rateDelta,
        trend: 'Deteriorating',
        impactScore: rateRisk.impactScore,
        likelihoodScore: rateRisk.likelihoodScore,
        riskLevel: rateRisk.riskLevel,
        aiNarrative: `Rata-rata tertimbang suku bunga portofolio kredit pada ${record.table_name} berada di ${cleanAvgRate}% dan diprediksi mengalami kompresi ke ${foreRate}%.`,
        recommendedAction: 'Evaluasi kepatuhan pricing matriks suku bunga kredit terhadap biaya dana (Cost of Fund) per segmen produk.',
        lastActualPeriod: '2026-Q2',
        history: [
          { period: '2025-Q3', value: +(cleanAvgRate * 1.04).toFixed(2) },
          { period: '2025-Q4', value: +(cleanAvgRate * 1.02).toFixed(2) },
          { period: '2026-Q1', value: +(cleanAvgRate * 1.01).toFixed(2) },
          { period: '2026-Q2', value: cleanAvgRate },
        ],
        forecast: [
          { period: '2026-Q3', value: +(cleanAvgRate * 0.98).toFixed(2), lower: +(cleanAvgRate * 0.95).toFixed(2), upper: +(cleanAvgRate * 1.01).toFixed(2) },
          { period: '2026-Q4', value: foreRate, lower: +(foreRate * 0.95).toFixed(2), upper: +(foreRate * 1.02).toFixed(2) },
        ],
      }
    )
  } else {
    const totalRows = Math.max(1, rows.length)
    const flagRows = rows.filter((r: any) => r.is_suspicious || r.is_flagged || r.is_anomaly || r.risk_flag || Number(r.severity_score) >= 70)
    const anomRate = +(((flagRows.length || 1) / totalRows) * 100).toFixed(2)
    const foreRate = +(anomRate * 1.15).toFixed(2)
    const deltaPct = +(((foreRate - anomRate) / Math.max(0.1, anomRate)) * 100).toFixed(1)
    const genericFallbackCode = anomRate >= 15
      ? 'HIGH'
      : (anomRate >= 8 ? 'MODERATE_HIGH' : (anomRate >= 4 ? 'MODERATE' : (anomRate >= 1 ? 'LOW_MODERATE' : 'LOW')))
    const genericRisk = resolveRisk(flagRows.length ? flagRows : rows, genericFallbackCode)

    generatedKpis.push({
      kpiName: `Rasio Deviasi & Eksposur Risiko (${record.table_name})`,
      code: `KPI-EXT-${record.table_name.slice(0, 3).toUpperCase()}`,
      unit: '%',
      entity: `Unit Operasional (${record.connection_name})`,
      entityType: 'Department',
      syncFrequency: `Live DB (${record.table_name})`,
      lastSyncedAt: nowTime,
      targetHorizon: 'Q4 2026',
      currentValue: anomRate,
      forecastedValue: foreRate,
      deltaPct,
      trend: 'Deteriorating',
      impactScore: genericRisk.impactScore,
      likelihoodScore: genericRisk.likelihoodScore,
      riskLevel: genericRisk.riskLevel,
      aiNarrative: `Berdasarkan populasi ${totalPop || totalRows} baris pada tabel ${record.table_name}, tingkat deviasi pengendalian berada di ${anomRate}% dan diproyeksikan naik menjadi ${foreRate}%.`,
      recommendedAction: `Laksanakan pengujian substantif dan perketat monitoring otomatis pada pipeline ${record.table_name}.`,
      lastActualPeriod: '2026-Q2',
      history: [
        { period: '2025-Q4', value: +(anomRate * 0.88).toFixed(2) },
        { period: '2026-Q1', value: +(anomRate * 0.94).toFixed(2) },
        { period: '2026-Q2', value: anomRate },
      ],
      forecast: [
        { period: '2026-Q3', value: +(anomRate * 1.08).toFixed(2), lower: anomRate, upper: +(anomRate * 1.15).toFixed(2) },
        { period: '2026-Q4', value: foreRate, lower: +(foreRate * 0.92).toFixed(2), upper: +(foreRate * 1.08).toFixed(2) },
      ],
    })
  }

  return generatedKpis
}

// ============================================================================
// 17 Additional Module Adapters (Risk Management, Audit Features, QAR Features)
// ============================================================================

function isRowFlaggedBool(row: Record<string, any>, record: MappedRecordInput): boolean {
  return isRowAnomalousOrFlagged(row, record.custom_rules, record.anomaly_rules).flagged
}

function extractRowRef(row: Record<string, any>, record: MappedRecordInput, idx = 0): string {
  const candidates = [
    record.audit_field ? row?.[record.audit_field] : undefined,
    row?.trx_ref_number,
    row?.loan_number,
    row?.voucher_number,
    row?.contract_no,
    row?.contract_number,
    row?.incident_code,
    row?.memo_no,
    row?.document_no,
    row?.id,
    row?.trx_id,
    row?.loan_id,
    row?.gl_id,
    row?.contract_id,
    row?.incident_id,
    row?.cabang_id,
  ]
  for (const c of candidates) {
    if (c !== undefined && c !== null && String(c).trim() !== '') {
      return String(c).trim()
    }
  }
  return `${record.table_name.slice(0, 3).toUpperCase()}-${idx + 1}`
}

function extractRowDimension(row: Record<string, any>, idx = 0): string {
  for (const col of DIMENSION_COL_CANDIDATES) {
    const val = row?.[col]
    if (val !== undefined && val !== null && String(val).trim() !== '') {
      return String(val).trim()
    }
  }
  for (const [k, v] of Object.entries(row || {})) {
    if (/(branch|unit|dept|department|entity|division|office|region|vendor|customer)/i.test(k) && !/_id$/i.test(k)) {
      if (v !== undefined && v !== null && String(v).trim() !== '') {
        return String(v).trim()
      }
    }
  }
  return `Unit Operasional ${idx + 1}`
}

function extractRowAmount(row: Record<string, any>): number {
  for (const col of AMOUNT_COL_CANDIDATES) {
    const val = Number(row?.[col])
    if (Number.isFinite(val) && val > 0) return val
  }
  return 0
}

export interface RiskAppetiteCandidate {
  statement: string
  threshold_limit: number
  status: 'DRAFT' | 'SUBMITTED' | 'APPROVED'
}

export function buildRiskAppetiteStatementsFromTable(record: MappedRecordInput): RiskAppetiteCandidate[] {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const totalPop = Number(record.profile?.total_rows || record.total_rows || rows.length || 1)
  const flaggedRows = rows.filter(r => isRowFlaggedBool(r, record))
  const flaggedCount = Number(record.profile?.flagged_rows ?? flaggedRows.length)
  const devRate = Math.min(100, Math.max(0.5, +((flaggedCount / Math.max(1, rows.length || totalPop)) * 100).toFixed(2)))
  const thresholdPct = Math.max(1, Math.min(25, Math.ceil(devRate)))

  const candidates: RiskAppetiteCandidate[] = [
    {
      statement: `[${record.table_name}] Batas toleransi maksimum rasio deviasi/anomali pengendalian pada ${record.connection_name} (populasi ${totalPop.toLocaleString('id-ID')} rekaman) ditetapkan maksimal ${thresholdPct}%.`,
      threshold_limit: thresholdPct,
      status: 'DRAFT',
    },
  ]

  const amounts = rows
    .map(r => extractRowAmount(r))
    .filter(v => Number.isFinite(v) && v > 0)
  if (amounts.length > 0 || Number(record.profile?.max_amount) > 0) {
    const maxAmt = Math.max(Number(record.profile?.max_amount || 0), ...(amounts.length ? amounts : [0]))
    const limitJt = Math.max(10, Math.round(maxAmt / 1_000_000))
    candidates.push({
      statement: `[${record.table_name}] Batas eksposur nilai transaksi/kontrak tunggal tanpa eskalasi komite risiko pada ${record.connection_name} tidak boleh melebihi Rp ${limitJt.toLocaleString('id-ID')} Juta.`,
      threshold_limit: limitJt,
      status: 'DRAFT',
    })
  }

  return candidates
}

export interface RiskFactorCandidate {
  name: string
  description: string
  score_guidelines: string
}

export function buildRiskFactorsFromTable(record: MappedRecordInput): RiskFactorCandidate[] {
  const s = analyzeTableSchema({
    tableName: record.table_name,
    profile: record.profile,
    sample_records: record.sample_records,
  })
  const cleanTbl = record.table_name.replace(/_/g, ' ')
  const candidates: RiskFactorCandidate[] = []

  if (s.hasMonetaryMeasure) {
    candidates.push({
      name: `Materialitas Eksposur Finansial (${record.table_name})`,
      description: `Besaran eksposur finansial pada tabel ${cleanTbl} (${record.connection_name}) berdasarkan kolom ${s.amountColumn || 'nominal'}.`,
      score_guidelines: JSON.stringify([
        { score: 1, desc: 'Eksposur rendah (< Rp 100 Juta)' },
        { score: 3, desc: 'Eksposur moderat (Rp 100 Juta - Rp 1 Miliar)' },
        { score: 5, desc: 'Eksposur kritikal (> Rp 1 Miliar)' },
      ]),
    })
  }

  if (s.hasAnomalyFlag || s.hasIncidentOrRiskSignals) {
    candidates.push({
      name: `Frekuensi Deviasi & Pelanggaran Kontrol (${record.table_name})`,
      description: `Tingkat pelanggaran otorisasi, anomali, atau insiden kepatuhan pada ${cleanTbl} (${s.flagColumns[0] || 'status risiko'}).`,
      score_guidelines: JSON.stringify([
        { score: 1, desc: 'Tidak ada deviasi / patuh penuh (< 2%)' },
        { score: 3, desc: 'Deviasi moderat (2% - 10%)' },
        { score: 5, desc: 'Deviasi tinggi / override berulang (> 10%)' },
      ]),
    })
  }

  if (candidates.length === 0) {
    candidates.push({
      name: `Kompleksitas Operasional Data (${record.table_name})`,
      description: `Tingkat kerumitan operasional dan volume transaksi pada sumber data ${cleanTbl} (${record.connection_name}).`,
      score_guidelines: JSON.stringify([
        { score: 1, desc: 'Volume & kompleksitas rendah' },
        { score: 3, desc: 'Volume & kompleksitas sedang' },
        { score: 5, desc: 'Volume tinggi & berdampak langsung pada pelaporan' },
      ]),
    })
  }

  return candidates
}

export interface AuditPriorityCandidate {
  entityName: string
  dimensionName: string
  factorScore: number
  riskLevel: string
  prioritized: boolean
  rowCount: number
  flaggedCount: number
}

export function buildAuditPriorityCandidatesFromTable(record: MappedRecordInput): AuditPriorityCandidate[] {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const groups = new Map<string, { rows: Record<string, any>[]; flagged: number; maxAmt: number; maxSev: number }>()

  for (let i = 0; i < rows.length; i++) {
    const r = rows[i] || {}
    const dim = extractRowDimension(r, i)
    const g = groups.get(dim) || { rows: [], flagged: 0, maxAmt: 0, maxSev: 0 }
    g.rows.push(r)
    if (isRowFlaggedBool(r, record)) g.flagged++
    const amt = extractRowAmount(r)
    if (amt > g.maxAmt) g.maxAmt = amt
    const sev = Number(r.severity_score || r.collectability * 20 || 0)
    if (sev > g.maxSev) g.maxSev = sev
    groups.set(dim, g)
  }

  if (groups.size === 0 && Array.isArray(record.profile?.dimension_breakdown)) {
    for (const d of record.profile!.dimension_breakdown!) {
      if (!d?.name) continue
      groups.set(d.name, {
        rows: [],
        flagged: Number(d.flagged_count || 0),
        maxAmt: Number(d.max_amount || d.total_amount || 0),
        maxSev: d.flagged_count > 0 ? 80 : 45,
      })
    }
  }

  const out: AuditPriorityCandidate[] = []
  for (const [dim, g] of groups.entries()) {
    const first = g.rows[0] || {}
    const rawTier = String(first.risk_tier || first.risk_level || '').toLowerCase()
    let factorScore = 3
    if (g.maxSev >= 80 || g.flagged >= 2 || rawTier.includes('high') || g.maxAmt >= 1_000_000_000) {
      factorScore = 5
    } else if (g.maxSev >= 65 || g.flagged === 1 || rawTier.includes('mod') || g.maxAmt >= 250_000_000) {
      factorScore = 4
    } else if (rawTier.includes('low') && g.flagged === 0) {
      factorScore = 2
    }

    const riskLevel =
      factorScore >= 5
        ? 'High'
        : factorScore === 4
          ? 'Moderate to High'
          : factorScore === 3
            ? 'Moderate'
            : 'Low to Moderate'

    out.push({
      entityName: `${dim} (${record.table_name})`,
      dimensionName: dim,
      factorScore,
      riskLevel,
      prioritized: factorScore >= 4,
      rowCount: g.rows.length || 1,
      flaggedCount: g.flagged,
    })
  }

  return out
}

export interface RcmCandidate {
  risk_code: string
  risk_event: string
  control_code: string
  control_description: string
  control_type: 'Preventive' | 'Detective' | 'Corrective'
  control_owner: string
  department: string
  year: number
  design_effectiveness_weight: number
  design_effectiveness_rating: number
  operating_effectiveness_weight: number
  operating_effectiveness_rating: number
  coverage_completeness_weight: number
  coverage_completeness_rating: number
  timeliness_weight: number
  timeliness_rating: number
  automation_monitoring_weight: number
  automation_monitoring_rating: number
  notes: string
}

export function buildRcmCandidatesFromTable(record: MappedRecordInput, year = new Date().getFullYear()): RcmCandidate[] {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const targetRows = rows.length > 0 ? rows.slice(0, 15) : [{}]
  const tblPrefix = record.table_name
    .split('_')
    .map(w => w[0] || '')
    .join('')
    .toUpperCase()
    .slice(0, 4) || 'EXT'

  return targetRows.map((row, idx) => {
    const refCode = extractRowRef(row, record, idx)
    const dim = extractRowDimension(row, idx)
    const flagged = isRowFlaggedBool(row, record)
    const baseRating = flagged ? 2 : 4

    const riskEvent =
      row.incident_title ||
      row.finding_summary ||
      row.control_weakness ||
      row.audit_note ||
      row.notes ||
      `Risiko ketidakpatuhan & anomali pada ${record.table_name} (${refCode} - ${dim})`

    const controlDesc =
      row.control_weakness
        ? `Verifikasi & penguatan kontrol atas: ${row.control_weakness}`
        : `Verifikasi otorisasi berjenjang, rekonsiliasi bukti (${refCode}), dan pemantauan otomatis pada ${dim}`

    return {
      risk_code: `RSK-${tblPrefix}-${String(idx + 1).padStart(3, '0')}`,
      risk_event: String(riskEvent).slice(0, 240),
      control_code: `CTL-${tblPrefix}-${String(idx + 1).padStart(3, '0')}`,
      control_description: String(controlDesc).slice(0, 240),
      control_type: flagged ? 'Detective' : 'Preventive',
      control_owner: `Kepala Unit ${dim}`,
      department: dim,
      year,
      design_effectiveness_weight: 20,
      design_effectiveness_rating: flagged ? 3 : 4,
      operating_effectiveness_weight: 20,
      operating_effectiveness_rating: baseRating,
      coverage_completeness_weight: 20,
      coverage_completeness_rating: flagged ? 3 : 4,
      timeliness_weight: 20,
      timeliness_rating: baseRating,
      automation_monitoring_weight: 20,
      automation_monitoring_rating: flagged ? 2 : 4,
      notes: `Sumber: ${record.connection_name} -> ${record.table_name} (${refCode})`,
    }
  })
}

export interface RiskMitigationCandidate {
  riskControlId: string
  riskEvent: string
  mitigationPlan: string
  supervisor: string
  pic: string
  unitInCharge: string
  start_date: string
  end_date: string
  notes: string
}

export function buildRiskMitigationCandidatesFromTable(record: MappedRecordInput): RiskMitigationCandidate[] {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const flaggedRows = rows.filter(r => isRowFlaggedBool(r, record))
  const sourceRows = (flaggedRows.length > 0 ? flaggedRows : rows).slice(0, 12)
  const tblPrefix = record.table_name.slice(0, 3).toUpperCase()
  const today = new Date()
  const startDate = today.toISOString().split('T')[0] || '2026-01-01'
  const endObj = new Date(today.getTime() + 45 * 24 * 60 * 60 * 1000)
  const endDate = endObj.toISOString().split('T')[0] || '2026-03-01'

  return sourceRows.map((row, idx) => {
    const refCode = extractRowRef(row, record, idx)
    const dim = extractRowDimension(row, idx)
    const amt = extractRowAmount(row)
    const amtText = amt > 0 ? ` (Eksposur Rp ${amt.toLocaleString('id-ID')})` : ''

    const riskEvent =
      row.incident_title ||
      row.finding_summary ||
      row.audit_note ||
      row.notes ||
      `Temuan deviasi ${record.table_name} [${refCode}] pada ${dim}${amtText}`

    const plan =
      row.root_cause || row.control_weakness
        ? `Perbaikan akar masalah (${row.root_cause || row.control_weakness}) melalui pengetatan dual-control dan validasi dokumen pendukung ${refCode}.`
        : `Melakukan investigasi, blokir/koreksi otorisasi atas rekaman ${refCode}, serta review kepatuhan SOP pada ${dim}.`

    return {
      riskControlId: `MIT-${tblPrefix}-${String(idx + 1).padStart(3, '0')}`,
      riskEvent: String(riskEvent).slice(0, 220),
      mitigationPlan: String(plan).slice(0, 300),
      supervisor: String(row.approver_id || row.posted_by || 'Internal Audit Supervisor'),
      pic: `PIC ${dim}`,
      unitInCharge: dim,
      start_date: startDate,
      end_date: endDate,
      notes: `Diimpor dari ${record.connection_name} (${record.table_name} #${refCode})`,
    }
  })
}

export interface AuditActivityPlanDraftCandidate {
  title: string
  activityType: 'Assurance' | 'Audit Khusus' | 'Audit Tujuan Tertentu' | 'Audit Investigasi'
  auditKind: string
  assignmentCode: string
  objectives: string[]
  criteria: string[]
}

export function buildAuditActivityPlanDraftFromTable(record: MappedRecordInput): AuditActivityPlanDraftCandidate {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const flaggedCount = rows.filter(r => isRowFlaggedBool(r, record)).length
  const totalPop = Number(record.profile?.total_rows || record.total_rows || rows.length || 0)
  const tblLower = record.table_name.toLowerCase()

  const activityType: AuditActivityPlanDraftCandidate['activityType'] =
    tblLower.includes('incident') || flaggedCount >= 5 ? 'Audit Khusus' : 'Assurance'

  const auditKind =
    tblLower.includes('gl') || tblLower.includes('loan') || tblLower.includes('trx') || tblLower.includes('transact')
      ? 'Keuangan'
      : tblLower.includes('vendor') || tblLower.includes('incident')
        ? 'Kepatuhan'
        : 'Operasional'

  const cleanPrefix = record.table_name.replace(/[^A-Za-z0-9]/g, '').toUpperCase().slice(0, 6) || 'EXT'

  return {
    title: `Rencana Kegiatan Audit — ${record.table_name} (${record.connection_name})`,
    activityType,
    auditKind,
    assignmentCode: `RKA-${cleanPrefix}-${new Date().getFullYear()}`,
    objectives: [
      `Mengevaluasi keandalan pengendalian internal pada populasi ${totalPop.toLocaleString('id-ID')} baris tabel ${record.table_name}.`,
      `Memverifikasi ${flaggedCount || rows.length} transaksi/rekaman berisiko tinggi terhadap kepatuhan otorisasi dan SOP perusahaan.`,
    ],
    criteria: [
      `Kebijakan & SOP Pengendalian Internal Perusahaan terkait ${record.table_name}.`,
      `Standar Audit Intern & Ketentuan Kepatuhan Regulator.`,
    ],
  }
}

export interface AssignmentLetterCandidate {
  auditTitle: string
  leader: string
  category: AuditCategory
  auditYear: string
  auditTeam: string
  startPeriod: string
  finishPeriod: string
  workingUnit: string
  workingUnitDepartmentId: string
  auditPurpose: string
  letterDate: string
  membersList: Array<{ name: string; role: string }>
  purposeList: string[]
  scopeList: string[]
  ccList: string[]
}

export function buildAssignmentLetterCandidatesFromTable(
  record: MappedRecordInput,
  ctx?: {
    departments?: Array<{ id: string; department_name?: string; name?: string; department_code?: string }>
    employees?: Array<{ full_name: string; employee_code?: string }>
    year?: string
  }
): AssignmentLetterCandidate[] {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const depts = ctx?.departments || []
  const emps = ctx?.employees || []
  const year = ctx?.year || String(new Date().getFullYear())
  const today = new Date().toISOString().split('T')[0] || `${year}-01-15`
  const finish = new Date(Date.now() + 21 * 24 * 60 * 60 * 1000).toISOString().split('T')[0] || `${year}-02-05`

  const formatEmp = (e?: { full_name: string; employee_code?: string }, fallback = 'Auditor Senior SKAI') =>
    e ? (e.employee_code ? `${e.full_name} (${e.employee_code})` : e.full_name) : fallback

  const leaderName = formatEmp(emps[0], 'Ketua Tim Audit Internal')
  const member1 = formatEmp(emps[1], 'Anggota Auditor 1')
  const member2 = formatEmp(emps[2], 'Anggota Auditor 2')

  const distinctUnits: string[] = []
  for (let i = 0; i < rows.length; i++) {
    const d = extractRowDimension(rows[i] || {}, i)
    if (d && !distinctUnits.includes(d)) distinctUnits.push(d)
    if (distinctUnits.length >= 3) break
  }
  if (distinctUnits.length === 0) {
    distinctUnits.push(`Unit Operasional ${record.table_name}`)
  }

  return distinctUnits.map((unitName) => {
    const matchedDept =
      depts.find(d => {
        const dName = (d.department_name || d.name || '').toLowerCase()
        return dName && (unitName.toLowerCase().includes(dName) || dName.includes(unitName.toLowerCase()))
      }) || depts[0]

    const deptLabel = matchedDept
      ? matchedDept.department_code
        ? `${matchedDept.department_name || matchedDept.name} (${matchedDept.department_code})`
        : (matchedDept.department_name || matchedDept.name || unitName)
      : unitName

    return {
      auditTitle: `Audit Pemeriksaan ${record.table_name} — ${unitName}`,
      leader: leaderName,
      category: AuditCategory.ASSURANCE,
      auditYear: year,
      auditTeam: 'SKAI',
      startPeriod: today,
      finishPeriod: finish,
      workingUnit: deptLabel,
      workingUnitDepartmentId: matchedDept?.id || '',
      auditPurpose: `Melaksanakan pemeriksaan substantif dan kepatuhan pada ${unitName} berdasarkan sumber data ${record.table_name} (${record.connection_name}).`,
      letterDate: today,
      membersList: [
        { name: leaderName, role: 'Chairperson' },
        { name: member1, role: 'Member' },
        { name: member2, role: 'Member' },
      ],
      purposeList: [
        `Menguji validitas dan otorisasi transaksi pada tabel ${record.table_name} di ${unitName}.`,
        `Menilai efektivitas pengendalian internal dan tindak lanjut anomali data.`,
      ],
      scopeList: [
        `Populasi data ${record.table_name} pada koneksi ${record.connection_name}.`,
        `Dokumen bukti transaksi dan log persetujuan pada ${unitName}.`,
      ],
      ccList: ['President Director', 'Audit Committee'],
    }
  })
}

export interface FieldworkSampleCandidate {
  assignmentLetterId: string
  documentName: string
  documentNumber: string
  date: string
  description: string
}

export interface FieldworkTestControlCandidate {
  assignmentLetterId: string
  controlName: string
  controlDescription: string
  controlType: string
  unitDivision: string
  testProcedure: string
  testResult: string
  finding: string
  recommendation: string
  mitigationPlan: string
  pic: string
  dueDate: string
}

export function buildAuditFieldworkCandidatesFromTable(
  record: MappedRecordInput,
  assignmentLetterId = 'ST-001/SKAI/2026'
): {
  samples: FieldworkSampleCandidate[]
  testControls: FieldworkTestControlCandidate[]
} {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const sortedRows = [...rows].sort((a, b) => {
    const aFlag = isRowFlaggedBool(a, record) ? 1 : 0
    const bFlag = isRowFlaggedBool(b, record) ? 1 : 0
    return bFlag - aFlag
  })
  const selectedRows = sortedRows.slice(0, 12)
  const today = new Date().toISOString().split('T')[0] || '2026-01-15'
  const dueDate = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0] || '2026-02-15'

  const samples: FieldworkSampleCandidate[] = selectedRows.map((row, idx) => {
    const refCode = extractRowRef(row, record, idx)
    const dim = extractRowDimension(row, idx)
    const amt = extractRowAmount(row)
    const rawDate =
      row.trx_date || row.gl_date || row.disbursed_at || row.signed_at || row.reported_at || row.created_at || today
    const dateStr = String(rawDate).split('T')[0]?.split(' ')[0] || today

    return {
      assignmentLetterId,
      documentName: `Bukti ${record.table_name} — ${dim}`,
      documentNumber: String(refCode),
      date: dateStr,
      description: `${row.finding_summary || row.audit_note || row.notes || row.description || `Sampel uji petik ${record.table_name}`}${amt > 0 ? ` (Nominal Rp ${amt.toLocaleString('id-ID')})` : ''}`,
    }
  })

  const testControls: FieldworkTestControlCandidate[] = selectedRows.slice(0, 6).map((row, idx) => {
    const refCode = extractRowRef(row, record, idx)
    const dim = extractRowDimension(row, idx)
    const flagged = isRowFlaggedBool(row, record)

    return {
      assignmentLetterId,
      controlName: `Kontrol Verifikasi ${record.table_name} (${refCode})`,
      controlDescription: `Pengujian kepatuhan otorisasi dan kelengkapan bukti pada ${dim} untuk rekaman ${refCode}.`,
      controlType: flagged ? 'Detective' : 'Preventive',
      unitDivision: dim,
      testProcedure: `Inspeksi dokumen sumber, rekonsiliasi nilai database ${record.table_name}, dan verifikasi jejak persetujuan.`,
      testResult: flagged ? 'Ineffective' : 'Effective',
      finding: flagged
        ? String(row.finding_summary || row.audit_note || row.notes || `Ditemukan deviasi otorisasi/parameter pada rekaman ${refCode} di ${dim}.`)
        : `Tidak ditemukan deviasi material pada rekaman ${refCode}.`,
      recommendation: flagged
        ? `Wajibkan verifikasi berjenjang (dual-control) dan peninjauan ulang transaksi ${refCode} pada ${dim}.`
        : `Pertahankan prosedur pengendalian berjalan pada ${dim}.`,
      mitigationPlan: flagged
        ? `Remediasi dokumen & pemutakhiran parameter kontrol sistem pada ${dim}.`
        : 'Monitoring berkala.',
      pic: `Kepala Unit ${dim}`,
      dueDate,
    }
  })

  return { samples, testControls }
}

export interface InternalAuditKpiCandidate {
  year: number
  period: string
  kpi_name: string
  target: number
  actual: number
  achievement_rate: number
  notes: string
}

export function buildInternalAuditKpiCandidatesFromTable(
  record: MappedRecordInput,
  year = new Date().getFullYear()
): InternalAuditKpiCandidate[] {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const totalPop = Number(record.profile?.total_rows || record.total_rows || rows.length || 1)
  const flaggedRows = rows.filter(r => isRowFlaggedBool(r, record))
  const sampleTotal = Math.max(1, rows.length)
  const complianceActual = +(((sampleTotal - flaggedRows.length) / sampleTotal) * 100).toFixed(2)
  const complianceTarget = 95
  const complianceRate = Math.min(100, +((complianceActual / complianceTarget) * 100).toFixed(2))

  const coverageActual = Math.min(100, +((rows.length / Math.max(1, Math.min(totalPop, 50))) * 100).toFixed(2))

  return [
    {
      year,
      period: 'Tahunan',
      kpi_name: `Tingkat Kepatuhan Pengendalian ${record.table_name} (${record.connection_name})`,
      target: complianceTarget,
      actual: complianceActual,
      achievement_rate: complianceRate,
      notes: `Dihitung dari ${sampleTotal} sampel uji (${flaggedRows.length} deviasi) pada populasi ${totalPop.toLocaleString('id-ID')} baris.`,
    },
    {
      year,
      period: 'Tahunan',
      kpi_name: `Cakupan Pengujian Substantif Data ${record.table_name}`,
      target: 100,
      actual: coverageActual,
      achievement_rate: coverageActual,
      notes: `Realisasi penarikan & pengujian bukti otomatis dari sumber data ${record.connection_name}.`,
    },
  ]
}

export interface AuditResultReportCandidate {
  assignmentLetterId: string
  reportTitle: string
  reportDate: string
  status: 'Draft' | 'Final'
  findingsCount: number
  executiveSummary: string
  findings: Array<{
    title: string
    category: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant'
    action: string
    source: string
  }>
}

export function buildAuditResultReportCandidateFromTable(
  record: MappedRecordInput,
  assignmentLetterId = 'ST-001/SKAI/2026',
  status: 'Draft' | 'Final' = 'Draft'
): AuditResultReportCandidate {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const flaggedRows = rows.filter(r => isRowFlaggedBool(r, record))
  const targetRows = (flaggedRows.length > 0 ? flaggedRows : rows).slice(0, 8)
  const today = new Date().toISOString().split('T')[0] || '2026-01-15'
  const totalPop = Number(record.profile?.total_rows || record.total_rows || rows.length || 0)

  const findings = targetRows.map((row, idx) => {
    const refCode = extractRowRef(row, record, idx)
    const dim = extractRowDimension(row, idx)
    const amt = extractRowAmount(row)
    const sev = Number(row.severity_score || 0)
    const isHigh = sev >= 80 || amt >= 1_000_000_000 || row.is_suspicious === true || row.is_override_auth === true
    const category: AuditResultReportCandidate['findings'][number]['category'] = isHigh
      ? 'Very Significant'
      : flaggedRows.length > 0
        ? 'Significant'
        : 'Quite Significant'

    const desc =
      row.incident_title ||
      row.finding_summary ||
      row.audit_note ||
      row.notes ||
      `Deviasi pengendalian & kepatuhan pada ${record.table_name}`

    return {
      title: `[${refCode}] ${desc} — ${dim}${amt > 0 ? ` (Rp ${amt.toLocaleString('id-ID')})` : ''}`.slice(0, 240),
      category,
      action: String(
        row.control_weakness
          ? `Perbaiki kelemahan kontrol (${row.control_weakness}) dan lakukan verifikasi ulang atas ${refCode} pada ${dim}.`
          : `Lakukan tindak lanjut korektif, rekonsiliasi bukti ${refCode}, dan pengetatan otorisasi pada ${dim}.`
      ).slice(0, 240),
      source: `${record.connection_name}:${record.table_name}`,
    }
  })

  return {
    assignmentLetterId,
    reportTitle: `Laporan Hasil Audit (LHA) — Evaluasi Data ${record.table_name} (${record.connection_name})`,
    reportDate: today,
    status,
    findingsCount: findings.length,
    executiveSummary: `Berdasarkan pengujian populasi ${totalPop.toLocaleString('id-ID')} baris pada tabel ${record.table_name} (${record.connection_name}), teridentifikasi ${findings.length} temuan utama yang memerlukan tindakan korektif manajemen.`,
    findings,
  }
}

export interface QualityAssuranceCandidate {
  type: string
  period: string
  reportName: string
  assessmentTitle: string
  result: string
  status: 'Completed' | 'Verified' | 'In Progress' | 'Planned'
  conductedBy: string
  internalEvaluator: string
}

export function buildQualityAssuranceCandidatesFromTable(
  record: MappedRecordInput,
  actionType: 'qar_dashboard' | 'periodic_self_assessment' | 'saiv_validation' | 'iacm_assessment' = 'qar_dashboard',
  conductedBy = 'PT AIFL Indonesia (Internal Quality Assurance)'
): QualityAssuranceCandidate[] {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const flaggedCount = rows.filter(r => isRowFlaggedBool(r, record)).length
  const totalSample = Math.max(1, rows.length)
  const passRate = Math.round(((totalSample - flaggedCount) / totalSample) * 100)
  const year = String(new Date().getFullYear())

  const typeMap: Record<typeof actionType, { qaType: string; prefix: string; resultText: string }> = {
    qar_dashboard: {
      qaType: 'Quality Assurance Review (QAR)',
      prefix: 'QAR Review',
      resultText: passRate >= 85 ? `Generally Conforms (${passRate}%)` : `Partially Conforms (${passRate}%)`,
    },
    periodic_self_assessment: {
      qaType: 'Regular Self Assessment (RSA)',
      prefix: 'Periodic Self-Assessment',
      resultText: passRate >= 85 ? `Generally Conforms (${passRate}%)` : `Partially Conforms (${passRate}%)`,
    },
    saiv_validation: {
      qaType: 'Self Assessment w/ Independent Validation (SAIV)',
      prefix: 'SAIV Independent Validation',
      resultText: passRate >= 85 ? `Validated — Generally Conforms (${passRate}%)` : `Validated — Partially Conforms (${passRate}%)`,
    },
    iacm_assessment: {
      qaType: 'BUMN IACM Assessment',
      prefix: 'IACM Capability Assessment',
      resultText: passRate >= 90 ? `Level 4 — Managed (${passRate}%)` : passRate >= 75 ? `Level 3 — Integrated (${passRate}%)` : `Level 2 — Infrastructure (${passRate}%)`,
    },
  }

  const meta = typeMap[actionType] || typeMap.qar_dashboard
  const title = `${meta.prefix} — ${record.table_name} (${record.connection_name})`

  return [
    {
      type: meta.qaType,
      period: year,
      reportName: title,
      assessmentTitle: title,
      result: meta.resultText,
      status: 'Completed',
      conductedBy,
      internalEvaluator: 'Tim Quality Assurance Audit Internal',
    },
  ]
}

export interface ConsultingServiceCandidate {
  title: string
  category: string
  requestorDept: string
  period: string
  consultantName: string
  status: string
  notes: string
}

export function buildConsultingServiceCandidatesFromTable(
  record: MappedRecordInput,
  departments?: Array<{ department_name?: string; name?: string; department_code?: string }>
): ConsultingServiceCandidate[] {
  const rows = Array.isArray(record.sample_records) ? record.sample_records : []
  const year = new Date().getFullYear()
  const tblLower = record.table_name.toLowerCase()
  const category =
    tblLower.includes('it') || tblLower.includes('login') || tblLower.includes('access')
      ? 'IT Advisory'
      : tblLower.includes('policy') || tblLower.includes('vendor') || tblLower.includes('contract')
        ? 'Policy & SOP Review'
        : tblLower.includes('incident') || tblLower.includes('risk')
          ? 'Governance Assessment'
          : 'Operational Advisory'

  const distinctUnits: string[] = []
  for (let i = 0; i < rows.length; i++) {
    const dim = extractRowDimension(rows[i] || {}, i)
    if (dim && !distinctUnits.includes(dim)) distinctUnits.push(dim)
    if (distinctUnits.length >= 3) break
  }
  if (distinctUnits.length === 0) {
    distinctUnits.push(`Unit Operasional ${record.table_name}`)
  }

  const depts = departments || []
  return distinctUnits.map((unit) => {
    const matched = depts.find(d => {
      const n = (d.department_name || d.name || '').toLowerCase()
      return n && (unit.toLowerCase().includes(n) || n.includes(unit.toLowerCase()))
    })
    const requestorDept = matched
      ? matched.department_code
        ? `${matched.department_name || matched.name} (${matched.department_code})`
        : (matched.department_name || matched.name || unit)
      : unit

    return {
      title: `Konsultasi Penguatan Pengendalian ${record.table_name} — ${unit}`,
      category,
      requestorDept,
      period: `Q1 ${year}`,
      consultantName: 'Tim Advisory Audit Internal (SKAI)',
      status: 'In Progress',
      notes: `Penugasan konsultasi berbasis analisis sumber data ${record.connection_name} (${record.table_name}) pada unit ${unit}.`,
    }
  })
}
