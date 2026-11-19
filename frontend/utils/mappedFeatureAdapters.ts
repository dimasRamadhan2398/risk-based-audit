import { AuditCategory } from '~/types/audit'
import { getCanonicalCategory } from '~/composables/useAnomalyScatter'
import { normalizeAnomalyRiskLevel } from '~/utils/riskLevelBadge'
import { sanitizeKpiNumberInput, STRATEGIC_PLAN_CATEGORIES } from '~/utils/strategicPlanPayload'

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
    | 'strategic_plan'
    | 'annual_plan'
    | 'working_paper'
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
      actionType: 'unsupported',
      actionLabel: 'Modul Belum Didukung',
      supported: false,
    }
  }
  if (path.includes('/risk-profile/risk-control-matrix')) {
    return {
      scope: 'risk_management',
      module: 'Risk Control Matrix (RCM)',
      slug: 'risk_control_matrix',
      label: 'Risk Control Matrix (RCM)',
      actionType: 'unsupported',
      actionLabel: 'Modul Belum Didukung',
      supported: false,
    }
  }
  if (path.includes('/risk-profile/audit-priority')) {
    return {
      scope: 'risk_management',
      module: 'Audit Priority',
      slug: 'audit_priority',
      label: 'Audit Priority',
      actionType: 'unsupported',
      actionLabel: 'Modul Belum Didukung',
      supported: false,
    }
  }
  if (path.includes('/risk-appetite')) {
    return {
      scope: 'risk_management',
      module: 'Risk Appetite Statement',
      slug: 'risk_appetite',
      label: 'Risk Appetite Statement',
      actionType: 'unsupported',
      actionLabel: 'Modul Belum Didukung',
      supported: false,
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
      actionType: 'unsupported',
      actionLabel: 'Modul Belum Didukung',
      supported: false,
    }
  }
  if (path.includes('/audit-activity-plan')) {
    return {
      scope: 'audit_features',
      module: 'Audit Activity Plan',
      slug: 'audit_activity_plan',
      label: 'Audit Activity Plan',
      actionType: 'unsupported',
      actionLabel: 'Modul Belum Didukung',
      supported: false,
    }
  }
  if (path.includes('/assignment-letter')) {
    return {
      scope: 'audit_features',
      module: 'Assignment Letter',
      slug: 'assignment_letter',
      label: 'Assignment Letter',
      actionType: 'unsupported',
      actionLabel: 'Modul Belum Didukung',
      supported: false,
    }
  }
  if (path.includes('/audit-execution-status')) {
    return {
      scope: 'audit_features',
      module: 'Audit Execution Status',
      slug: 'audit_execution_status',
      label: 'Audit Execution Status',
      actionType: 'unsupported',
      actionLabel: 'Modul Belum Didukung',
      supported: false,
    }
  }
  if (path.includes('/audit-result-report')) {
    return {
      scope: 'audit_features',
      module: 'Audit Result Report (LHA)',
      slug: 'audit_result_report',
      label: 'Audit Result Report (LHA)',
      actionType: 'unsupported',
      actionLabel: 'Modul Belum Didukung',
      supported: false,
    }
  }
  if (path.includes('/action-taken-report')) {
    return {
      scope: 'audit_features',
      module: 'Action Taken Report (ATR)',
      slug: 'action_taken_report',
      label: 'Action Taken Report (ATR)',
      actionType: 'unsupported',
      actionLabel: 'Modul Belum Didukung',
      supported: false,
    }
  }
  if (path.includes('/kpi-performance')) {
    return {
      scope: 'audit_features',
      module: 'Internal Audit Performance (KPI)',
      slug: 'kpi_performance',
      label: 'Internal Audit Performance (KPI)',
      actionType: 'unsupported',
      actionLabel: 'Modul Belum Didukung',
      supported: false,
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
  const detectedColumns = extractDetectedColumns(record)
  const lowerMap = new Map<string, string>()
  for (const c of detectedColumns) {
    lowerMap.set(c.toLowerCase(), c)
  }

  const dimensionColumn =
    record.profile?.dimension_column ||
    DIMENSION_COL_CANDIDATES.map(c => lowerMap.get(c)).find(Boolean) ||
    null

  const amountColumn =
    record.profile?.amount_column ||
    AMOUNT_COL_CANDIDATES.map(c => lowerMap.get(c)).find(Boolean) ||
    null

  const tblLower = (record.table_name || '').trim().toLowerCase()
  const isKnownReferenceOnly = KNOWN_REFERENCE_ONLY_TABLES.has(tblLower)

  const hasNumericMeasure =
    !isKnownReferenceOnly &&
    (record.profile?.has_numeric_measure === true ||
      Boolean(amountColumn) ||
      Array.from(lowerMap.keys()).some(k => NUMERIC_OR_FLAG_COL_CANDIDATES.has(k)))

  if (actionType === 'unsupported') {
    return {
      compatible: false,
      reason: `Modul '${record.target_module}' belum memiliki adapter integrasi otomatis dari tabel sumber data.`,
      hasNumericMeasure,
      dimensionColumn,
      amountColumn,
      detectedColumns,
    }
  }

  const hasAnyData =
    (Array.isArray(record.sample_records) && record.sample_records.length > 0) ||
    Number(record.total_rows || record.profile?.total_rows || 0) > 0

  if (record.data_hub_unavailable && !hasAnyData) {
    return {
      compatible: false,
      reason: `Data Hub sedang tidak terjangkau untuk menarik data nyata dari tabel '${record.table_name}'. Pastikan koneksi Data Hub aktif.`,
      hasNumericMeasure,
      dimensionColumn,
      amountColumn,
      detectedColumns,
    }
  }

  if (
    !hasNumericMeasure &&
    (actionType === 'strategic_plan' || actionType === 'ai_anomaly' || actionType === 'ai_kpi_forecast')
  ) {
    return {
      compatible: false,
      reason: `Tabel '${record.table_name}' adalah tabel referensi master tanpa kolom metrik/ukuran numerik (nominal, skor risiko, atau flag anomali), sehingga tidak kompatibel untuk modul ini.`,
      hasNumericMeasure,
      dimensionColumn,
      amountColumn,
      detectedColumns,
    }
  }

  return {
    compatible: true,
    hasNumericMeasure,
    dimensionColumn,
    amountColumn,
    detectedColumns,
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
