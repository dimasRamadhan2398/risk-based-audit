export interface TableColumnInput {
  name: string
  dataType?: string
  isNullable?: boolean
  isPrimary?: boolean
}

export interface TableProfileSignalInput {
  total_rows?: number
  flagged_rows?: number
  has_numeric_measure?: boolean
  dimension_column?: string | null
  amount_column?: string | null
  columns?: string[]
}

export interface TableCompatibilityInput {
  tableName: string
  columns?: TableColumnInput[]
  rowCount?: number
  profile?: TableProfileSignalInput
  sample_records?: Record<string, any>[]
  data_hub_unavailable?: boolean
}

export type TableArchetype =
  | 'transactional'
  | 'incident_risk'
  | 'timeseries_metric'
  | 'narrative_document'
  | 'reference'

export type ModuleImportMode =
  | 'auto'
  | 'manual_banner'
  | 'prefill_form'
  | 'unsupported'

export interface TableSchemaSignals {
  tableName: string
  archetype: TableArchetype
  isReferenceTable: boolean
  detectedColumns: string[]
  primaryKeyColumn: string | null
  refColumn: string | null
  hasEvidenceIdentifier: boolean
  hasRefNumber: boolean
  amountColumn: string | null
  hasMonetaryMeasure: boolean
  metricColumns: string[]
  hasQuantitativeKpi: boolean
  flagColumns: string[]
  hasAnomalyFlag: boolean
  hasNumericMeasure: boolean
  eventDateColumn: string | null
  hasEventDate: boolean
  dimensionColumn: string | null
  hasDimensionColumn: boolean
  textColumns: string[]
  hasNarrativeText: boolean
  hasContractOrAuditText: boolean
  hasRiskTierColumn: boolean
  hasIncidentOrRiskSignals: boolean
}

export interface ModuleCompatibilityEvaluation {
  targetModule: string
  targetScope: string
  actionType:
    | 'risk_profile'
    | 'audit_universe'
    | 'strategic_plan'
    | 'annual_plan'
    | 'working_paper'
    | 'ai_entity_prediction'
    | 'ai_anomaly'
    | 'ai_nlp'
    | 'ai_kpi_forecast'
    | 'unsupported'
  supported: boolean
  compatible: boolean
  recommended: boolean
  priority: number
  importMode: ModuleImportMode
  route: string
  reason: string
  shortReason: string
}

export interface ScopedModuleOption {
  label: string
  value: string
  disabled?: boolean
  recommended?: boolean
  compatible?: boolean
  reason?: string
  importMode?: ModuleImportMode
}

export interface ScopeOptionForTable {
  label: string
  value: string
  disabled?: boolean
  compatibleCount: number
}

export interface RecommendedMappingSpec {
  targetScope: string
  targetModule: string
  priority: number
  reason: string
  importMode: ModuleImportMode
}

export interface ModuleCatalogEntry {
  scope: string
  module: string
  actionType: ModuleCompatibilityEvaluation['actionType']
  supported: boolean
  importMode: ModuleImportMode
  route: string
}

export const MODULE_CATALOG: ModuleCatalogEntry[] = [
  // Risk Management
  { scope: 'risk_management', module: 'Corporate Risk Profile', actionType: 'risk_profile', supported: true, importMode: 'auto', route: '/risk-profile' },
  { scope: 'risk_management', module: 'Audit Universe', actionType: 'audit_universe', supported: true, importMode: 'auto', route: '/audit-universe' },
  { scope: 'risk_management', module: 'Risk Appetite Statement', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/risk-appetite' },
  { scope: 'risk_management', module: 'Risk Factors', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/risk-factors' },
  { scope: 'risk_management', module: 'Audit Priority', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/audit-priority' },
  { scope: 'risk_management', module: 'Risk Control Matrix (RCM)', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/rcm' },
  { scope: 'risk_management', module: 'Risk Mitigation', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/risk-mitigation' },

  // Audit Features
  { scope: 'audit_features', module: 'Strategic Audit Plan', actionType: 'strategic_plan', supported: true, importMode: 'manual_banner', route: '/strategic-audit-plan' },
  { scope: 'audit_features', module: 'Annual Audit Plan', actionType: 'annual_plan', supported: true, importMode: 'manual_banner', route: '/annual-audit' },
  { scope: 'audit_features', module: 'Working Paper Evidence', actionType: 'working_paper', supported: true, importMode: 'prefill_form', route: '/working-paper?step=f03' },
  { scope: 'audit_features', module: 'Audit Activity Plan', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/audit-activity-plan' },
  { scope: 'audit_features', module: 'Assignment Letter', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/assignment-letter' },
  { scope: 'audit_features', module: 'Audit Execution Status', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/audit-execution' },
  { scope: 'audit_features', module: 'Audit Fieldwork', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/audit-fieldwork' },
  { scope: 'audit_features', module: 'Internal Audit Performance (KPI)', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/performance-kpi' },
  { scope: 'audit_features', module: 'Audit Result Report (LHA)', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/audit-result-report' },
  { scope: 'audit_features', module: 'Action Taken Report (ATR)', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/action-taken-report' },

  // QAR Features
  { scope: 'qar_features', module: 'Quality Assurance Dashboard', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/qar' },
  { scope: 'qar_features', module: 'Periodic Self-Assessment', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/qar' },
  { scope: 'qar_features', module: 'SAIV Validation', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/qar' },
  { scope: 'qar_features', module: 'IACM Assessment', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/qar' },
  { scope: 'qar_features', module: 'Consulting Service', actionType: 'unsupported', supported: false, importMode: 'unsupported', route: '/qar' },

  // Data Analytics / AI
  { scope: 'data_analytics', module: 'AI: Entity Prediction', actionType: 'ai_entity_prediction', supported: true, importMode: 'auto', route: '/analytics/ai/risk-scoring' },
  { scope: 'data_analytics', module: 'AI: Anomaly Detection', actionType: 'ai_anomaly', supported: true, importMode: 'auto', route: '/analytics/ai/anomaly-detection' },
  { scope: 'data_analytics', module: 'AI: NLP Document Analysis', actionType: 'ai_nlp', supported: true, importMode: 'auto', route: '/analytics/ai/nlp-analysis' },
  { scope: 'data_analytics', module: 'AI: KPI & Loss Forecast', actionType: 'ai_kpi_forecast', supported: true, importMode: 'auto', route: '/analytics/ai/kpi-forecast' },
]

const SCOPE_LABELS: Record<string, string> = {
  risk_management: 'Risk Management',
  audit_features: 'Audit Features',
  qar_features: 'QAR Features',
  data_analytics: 'Data Analytics / AI',
}

const KNOWN_REFERENCE_TABLE_PATTERN = /^(branches|branch|departments|department|locations|location|employees|employee|users|roles|currencies|categories|units|divisions|master_.+|ref_.+|dim_.+)$/i

const DIMENSION_CANDIDATES = [
  'branch_name',
  'entity_unit',
  'department',
  'department_name',
  'unit_name',
  'office_name',
  'division',
  'reporting_unit',
  'region',
  'city',
  'vendor_name',
  'customer_name',
  'gl_account_name',
  'product_name',
  'risk_category',
  'category',
  'branch_code',
  'entity',
]

const AMOUNT_CANDIDATES = [
  'amount',
  'debit_amount',
  'credit_amount',
  'plafond',
  'outstanding',
  'contract_value',
  'exposure_value',
  'invoice_amount',
  'financial_loss_idr',
  'financial_loss',
  'collateral_value',
  'total_amount',
  'balance',
  'nominal',
  'nilai',
  'budget',
  'cost',
  'revenue',
  'loss',
  'limit',
]

const QUANTITATIVE_KPI_PATTERN =
  /^(severity_score|risk_score|score|collectability|days_past_due|dpd|ltv_ratio|interest_rate|tenor_months|kpi_value|actual_value|target_value|realization|ratio|rate|percentage|pct|defect_rate|sla_breach_count|loss_amount|exposure_value|amount|plafond|outstanding|debit_amount|credit_amount|contract_value|invoice_amount)$/i

const ANOMALY_OR_COMPLIANCE_FLAG_PATTERN =
  /^(is_suspicious|is_flagged|is_anomaly|is_split_contract|is_policy_breach|is_override_auth|risk_flag|authorization_status|approval_status|compliance_status|rate_within_policy|collectability|days_past_due|voucher_number)$/i

const INCIDENT_OR_RISK_PATTERN =
  /^(incident_id|incident_code|incident_title|incident_description|finding_summary|severity_score|severity|risk_category|risk_level|risk_rating|risk_score|exposure_value|root_cause|control_weakness|investigator_notes|is_suspicious|is_flagged|is_policy_breach|is_override_auth|authorization_status|approval_status|compliance_status|collectability|days_past_due)$/i

const RISK_TIER_PATTERN = /^(risk_tier|risk_level|risk_rating|severity_level|tingkat_risiko|kol_status)$/i

const REF_NUMBER_PATTERN =
  /(ref_number|trx_ref|loan_number|voucher_number|contract_no|contract_number|incident_code|invoice_no|invoice_number|po_number|journal_no|document_no|receipt_no|ticket_no|case_no|transaction_ref)/i

const EVENT_DATE_PATTERN =
  /^(trx_date|transaction_date|gl_date|journal_date|posting_date|disbursed_at|signed_at|reported_at|paid_at|logged_at|incident_date|effective_date|event_date|occurred_at|period_date|period|month|quarter|fiscal_year|year)$/i

const METADATA_TIMESTAMP_PATTERN = /^(created_at|updated_at|deleted_at|synced_at|ingested_at|modified_at)$/i

const NARRATIVE_COL_NAME_PATTERN =
  /(notes?|description|summary|finding|narrative|remarks?|comment|clause|reason|root_cause|weakness|audit_note|investigator_notes|policy_text|memo|explanation|justification|contract_title|details)/i

const CONTRACT_OR_AUDIT_TEXT_PATTERN =
  /(audit_note|audit_notes|finding_summary|incident_description|investigator_notes|root_cause|control_weakness|sla_penalty_clause|contract_title|policy_text|clause)/i

const NUMERIC_SQL_TYPE_PATTERN = /(numeric|decimal|money|float|double|real|int|integer|bigint|smallint)/i
const TEXT_SQL_TYPE_PATTERN = /(^text$|clob|longtext|mediumtext|ntext)/i
const DATE_SQL_TYPE_PATTERN = /(timestamp|date|time)/i
const ID_COL_NAME_PATTERN = /(^id$|_id$|^code$|_code$|^zip|^postal|^phone|^telp)/i

function normalizeColumnInputs(table: TableCompatibilityInput): TableColumnInput[] {
  const map = new Map<string, TableColumnInput>()

  if (Array.isArray(table.columns)) {
    for (const col of table.columns) {
      if (!col?.name) continue
      map.set(col.name.toLowerCase(), {
        name: col.name,
        dataType: col.dataType || '',
        isNullable: col.isNullable,
        isPrimary: col.isPrimary,
      })
    }
  }

  if (Array.isArray(table.profile?.columns)) {
    for (const colName of table.profile.columns) {
      if (!colName) continue
      const key = colName.toLowerCase()
      if (!map.has(key)) {
        map.set(key, { name: colName, dataType: '' })
      }
    }
  }

  if (Array.isArray(table.sample_records) && table.sample_records.length > 0) {
    const sampleLimit = Math.min(table.sample_records.length, 25)
    for (let i = 0; i < sampleLimit; i++) {
      const row = table.sample_records[i]
      if (!row || typeof row !== 'object') continue
      for (const [k, v] of Object.entries(row)) {
        const key = k.toLowerCase()
        if (!map.has(key)) {
          let inferredType = ''
          if (typeof v === 'number') inferredType = 'numeric'
          else if (typeof v === 'boolean') inferredType = 'boolean'
          else if (typeof v === 'string' && v.length >= 40) inferredType = 'text'
          else if (typeof v === 'string') inferredType = 'varchar'
          map.set(key, { name: k, dataType: inferredType })
        }
      }
    }
  }

  return Array.from(map.values())
}

export function analyzeTableSchema(table: TableCompatibilityInput): TableSchemaSignals {
  const cols = normalizeColumnInputs(table)
  const lowerMap = new Map<string, TableColumnInput>()
  for (const c of cols) {
    lowerMap.set(c.name.toLowerCase(), c)
  }

  const detectedColumns = cols.map(c => c.name)

  // 1. Primary Key & Evidence Ref
  const pkCol =
    cols.find(c => c.isPrimary)?.name ||
    cols.find(c => /^(.*_)?id$/i.test(c.name))?.name ||
    cols[0]?.name ||
    null

  const refCol =
    cols.find(c => REF_NUMBER_PATTERN.test(c.name))?.name || null

  const hasEvidenceIdentifier = Boolean(pkCol || refCol)
  const hasRefNumber = Boolean(refCol)

  // 2. Dimension / Entity column
  const dimensionColumn =
    (table.profile?.dimension_column && lowerMap.get(table.profile.dimension_column.toLowerCase())?.name) ||
    table.profile?.dimension_column ||
    DIMENSION_CANDIDATES.map(c => lowerMap.get(c)?.name).find(Boolean) ||
    cols.find(c => /(branch|unit|dept|department|entity|division|office|region|vendor|customer)/i.test(c.name) && !/_id$/i.test(c.name))?.name ||
    null

  const hasDimensionColumn = Boolean(dimensionColumn)

  // 3. Monetary / Amount column
  const amountColumn =
    (table.profile?.amount_column && lowerMap.get(table.profile.amount_column.toLowerCase())?.name) ||
    table.profile?.amount_column ||
    AMOUNT_CANDIDATES.map(c => lowerMap.get(c)?.name).find(Boolean) ||
    cols.find(c =>
      /(amount|value|plafond|outstanding|balance|nominal|nilai|exposure|loss|budget|cost|revenue)/i.test(c.name) &&
      !c.isPrimary &&
      !ID_COL_NAME_PATTERN.test(c.name)
    )?.name ||
    null

  const hasMonetaryMeasure = Boolean(amountColumn)

  // 4. Quantitative / KPI Metric columns (excluding PK and *_id foreign keys)
  const metricColumns: string[] = []
  for (const c of cols) {
    if (c.isPrimary) continue
    if (ID_COL_NAME_PATTERN.test(c.name)) continue
    if (METADATA_TIMESTAMP_PATTERN.test(c.name)) continue
    if (
      QUANTITATIVE_KPI_PATTERN.test(c.name) ||
      (c.dataType && NUMERIC_SQL_TYPE_PATTERN.test(c.dataType) && !/serial/i.test(c.dataType))
    ) {
      metricColumns.push(c.name)
    }
  }
  const hasQuantitativeKpi = metricColumns.some(
    name =>
      /(score|collectability|days_past_due|dpd|ltv|rate|tenor|ratio|pct|percent|kpi|target|actual|realization|count)/i.test(name) ||
      (metricColumns.length >= 2 && hasMonetaryMeasure)
  )

  // 5. Anomaly / Compliance Flag columns
  const flagColumns = cols
    .filter(c => ANOMALY_OR_COMPLIANCE_FLAG_PATTERN.test(c.name))
    .map(c => c.name)
  const hasAnomalyFlag = flagColumns.length > 0

  // 6. Business Event / Period Date (distinct from metadata created_at/updated_at)
  const eventDateColumn =
    cols.find(c => EVENT_DATE_PATTERN.test(c.name))?.name ||
    cols.find(
      c =>
        !METADATA_TIMESTAMP_PATTERN.test(c.name) &&
        ((c.dataType && DATE_SQL_TYPE_PATTERN.test(c.dataType)) || /(_date|_at)$/i.test(c.name))
    )?.name ||
    null
  const hasEventDate = Boolean(eventDateColumn)

  // 7. Narrative / Text columns
  const textColumns: string[] = []
  for (const c of cols) {
    const isTextType = Boolean(c.dataType && TEXT_SQL_TYPE_PATTERN.test(c.dataType))
    const isNarrativeName = NARRATIVE_COL_NAME_PATTERN.test(c.name)
    let hasLongSampleText = false
    if (!isTextType && !isNarrativeName && Array.isArray(table.sample_records) && table.sample_records.length > 0) {
      hasLongSampleText = table.sample_records.slice(0, 10).some(row => {
        const val = row?.[c.name]
        return typeof val === 'string' && val.trim().length >= 45 && val.trim().includes(' ')
      })
    }
    if (isTextType || isNarrativeName || hasLongSampleText) {
      textColumns.push(c.name)
    }
  }
  const hasNarrativeText = textColumns.length > 0
  const hasContractOrAuditText = textColumns.some(name => CONTRACT_OR_AUDIT_TEXT_PATTERN.test(name))

  // 8. Risk tier & Incident/Risk signals
  const hasRiskTierColumn = cols.some(c => RISK_TIER_PATTERN.test(c.name))
  const hasIncidentOrRiskSignals = cols.some(c => INCIDENT_OR_RISK_PATTERN.test(c.name))

  // 9. Classify Master / Reference Table vs Transactional / Incident / Metric / Narrative
  const hasProfileNumeric = table.profile?.has_numeric_measure === true
  const hasDirectNumericOrFlag = hasMonetaryMeasure || metricColumns.length > 0 || hasAnomalyFlag

  const isNamedReference = KNOWN_REFERENCE_TABLE_PATTERN.test((table.tableName || '').trim())
  const isReferenceTable =
    !hasMonetaryMeasure &&
    metricColumns.length === 0 &&
    !hasAnomalyFlag &&
    !hasEventDate &&
    !hasNarrativeText &&
    (isNamedReference || hasDimensionColumn || !hasProfileNumeric)

  const hasNumericMeasure =
    !isReferenceTable && (hasProfileNumeric || hasDirectNumericOrFlag)

  let archetype: TableArchetype = 'transactional'
  if (isReferenceTable) {
    archetype = 'reference'
  } else if (
    cols.some(c => /^(incident_id|incident_code|severity_score|finding_summary|investigator_notes|is_override_auth)$/i.test(c.name))
  ) {
    archetype = 'incident_risk'
  } else if (hasMonetaryMeasure || (hasEventDate && hasRefNumber)) {
    archetype = 'transactional'
  } else if (hasNarrativeText && !hasNumericMeasure) {
    archetype = 'narrative_document'
  } else if (hasQuantitativeKpi && hasEventDate && !hasRefNumber) {
    archetype = 'timeseries_metric'
  } else if (hasNumericMeasure) {
    archetype = 'transactional'
  } else {
    archetype = 'reference'
  }

  return {
    tableName: table.tableName,
    archetype,
    isReferenceTable,
    detectedColumns,
    primaryKeyColumn: pkCol,
    refColumn: refCol,
    hasEvidenceIdentifier,
    hasRefNumber,
    amountColumn,
    hasMonetaryMeasure,
    metricColumns,
    hasQuantitativeKpi,
    flagColumns,
    hasAnomalyFlag,
    hasNumericMeasure,
    eventDateColumn,
    hasEventDate,
    dimensionColumn,
    hasDimensionColumn,
    textColumns,
    hasNarrativeText,
    hasContractOrAuditText,
    hasRiskTierColumn,
    hasIncidentOrRiskSignals,
  }
}

export function getTableArchetypeMeta(archetype: TableArchetype): {
  label: string
  color: 'info' | 'warning' | 'neutral' | 'primary' | 'success'
  description: string
} {
  switch (archetype) {
    case 'transactional':
      return {
        label: 'Transaksional',
        color: 'info',
        description: 'Memiliki nominal/transaksi, referensi bukti, dan dimensi operasional',
      }
    case 'incident_risk':
      return {
        label: 'Insiden & Risiko',
        color: 'warning',
        description: 'Berisi catatan insiden, skor keparahan, pelanggaran, atau eksposur risiko',
      }
    case 'timeseries_metric':
      return {
        label: 'Metrik Historis',
        color: 'success',
        description: 'Berisi metrik kuantitatif dan periode waktu untuk KPI/peramalan',
      }
    case 'narrative_document':
      return {
        label: 'Dokumen & Naratif',
        color: 'primary',
        description: 'Dominan teks naratif/kontrak/catatan untuk analisis NLP & uji bukti',
      }
    case 'reference':
    default:
      return {
        label: 'Referensi Master',
        color: 'neutral',
        description: 'Tabel direktori/master entitas tanpa metrik transaksi numerik',
      }
  }
}

export function actionTypeToTargetModule(actionType: string): string | null {
  const entry = MODULE_CATALOG.find(m => m.actionType === actionType && m.supported)
  return entry ? entry.module : null
}

export function targetModuleToImportMode(targetModule?: string): ModuleImportMode {
  if (!targetModule) return 'unsupported'
  const norm = targetModule.trim().toLowerCase()
  const entry = MODULE_CATALOG.find(m => m.module.toLowerCase() === norm)
  return entry?.importMode || 'unsupported'
}

export function targetModuleToRoute(targetModule?: string): string {
  if (!targetModule) return '/'
  const norm = targetModule.trim().toLowerCase()
  const entry = MODULE_CATALOG.find(m => m.module.toLowerCase() === norm)
  return entry?.route || '/'
}

export function evaluateModuleCompatibility(
  table: TableCompatibilityInput,
  targetModule: string
): ModuleCompatibilityEvaluation {
  const normalizedTarget = (targetModule || '').trim()
  const entry = MODULE_CATALOG.find(
    m => m.module.toLowerCase() === normalizedTarget.toLowerCase()
  )

  if (!entry || !entry.supported) {
    return {
      targetModule: normalizedTarget || 'Unknown',
      targetScope: entry?.scope || 'audit_features',
      actionType: 'unsupported',
      supported: false,
      compatible: false,
      recommended: false,
      priority: 0,
      importMode: 'unsupported',
      route: entry?.route || '/',
      reason: `Modul '${normalizedTarget}' belum memiliki adapter integrasi otomatis dari tabel sumber data.`,
      shortReason: 'Belum didukung',
    }
  }

  const s = analyzeTableSchema(table)

  // Handle Data Hub unavailability when called from banner with zero rows
  const hasAnyData =
    (Array.isArray(table.sample_records) && table.sample_records.length > 0) ||
    Number(table.rowCount || table.profile?.total_rows || 0) > 0

  if (table.data_hub_unavailable && !hasAnyData) {
    return {
      targetModule: entry.module,
      targetScope: entry.scope,
      actionType: entry.actionType,
      supported: true,
      compatible: false,
      recommended: false,
      priority: 0,
      importMode: entry.importMode,
      route: entry.route,
      reason: `Data Hub sedang tidak terjangkau untuk menarik data nyata dari tabel '${table.tableName}'. Pastikan koneksi Data Hub aktif.`,
      shortReason: 'Data Hub tidak terjangkau',
    }
  }

  switch (entry.actionType) {
    case 'risk_profile': {
      const ok =
        !s.isReferenceTable &&
        (s.hasNumericMeasure || s.hasIncidentOrRiskSignals || s.hasAnomalyFlag || s.hasNarrativeText)
      if (!ok) {
        return {
          targetModule: entry.module,
          targetScope: entry.scope,
          actionType: entry.actionType,
          supported: true,
          compatible: false,
          recommended: false,
          priority: 0,
          importMode: entry.importMode,
          route: entry.route,
          reason: `Tabel '${table.tableName}' adalah tabel referensi master tanpa data transaksi/insiden atau kolom metrik risiko untuk Corporate Risk Profile.`,
          shortReason: 'Tabel referensi master tanpa data transaksi/risiko',
        }
      }
      const recommended = s.hasIncidentOrRiskSignals || s.hasMonetaryMeasure || s.hasAnomalyFlag
      const signals: string[] = []
      if (s.amountColumn) signals.push(`nominal (${s.amountColumn})`)
      if (s.flagColumns.length) signals.push(`sinyal risiko (${s.flagColumns[0]})`)
      if (s.dimensionColumn) signals.push(`unit/cabang (${s.dimensionColumn})`)
      return {
        targetModule: entry.module,
        targetScope: entry.scope,
        actionType: entry.actionType,
        supported: true,
        compatible: true,
        recommended,
        priority: s.archetype === 'incident_risk' ? 98 : 90,
        importMode: entry.importMode,
        route: entry.route,
        reason: signals.length
          ? `Cocok: terdeteksi ${signals.join(', ')} untuk membentuk profil risiko per baris/transaksi.`
          : `Cocok: baris pada tabel '${table.tableName}' dapat dipetakan menjadi profil risiko.`,
        shortReason: 'Cocok',
      }
    }

    case 'audit_universe': {
      const ok = s.hasDimensionColumn || s.hasEvidenceIdentifier
      if (!ok) {
        return {
          targetModule: entry.module,
          targetScope: entry.scope,
          actionType: entry.actionType,
          supported: true,
          compatible: false,
          recommended: false,
          priority: 0,
          importMode: entry.importMode,
          route: entry.route,
          reason: `Tabel '${table.tableName}' tidak memiliki kolom entitas/unit/cabang atau pengenal baris untuk Audit Universe.`,
          shortReason: 'Tidak memiliki kolom entitas/dimensi',
        }
      }
      const recommended = s.isReferenceTable ? s.hasDimensionColumn : s.hasDimensionColumn
      return {
        targetModule: entry.module,
        targetScope: entry.scope,
        actionType: entry.actionType,
        supported: true,
        compatible: true,
        recommended,
        priority: s.isReferenceTable ? 100 : 75,
        importMode: entry.importMode,
        route: entry.route,
        reason: s.dimensionColumn
          ? `Cocok: terdeteksi dimensi entitas (${s.dimensionColumn}) yang dapat didaftarkan ke Audit Universe.`
          : `Cocok: entitas dapat dibentuk dari pengenal baris (${s.primaryKeyColumn}).`,
        shortReason: 'Cocok',
      }
    }

    case 'strategic_plan': {
      const ok = !s.isReferenceTable && s.hasNumericMeasure
      if (!ok) {
        return {
          targetModule: entry.module,
          targetScope: entry.scope,
          actionType: entry.actionType,
          supported: true,
          compatible: false,
          recommended: false,
          priority: 0,
          importMode: entry.importMode,
          route: entry.route,
          reason: s.isReferenceTable
            ? `Tabel '${table.tableName}' adalah tabel referensi master tanpa kolom metrik/ukuran numerik (nominal, skor risiko, atau flag anomali), sehingga tidak kompatibel untuk Strategic Audit Plan.`
            : `Tabel '${table.tableName}' tidak memiliki kolom metrik/ukuran numerik untuk menurunkan indikator KPI Strategic Audit Plan.`,
          shortReason: s.isReferenceTable
            ? 'Tabel referensi master tanpa metrik numerik'
            : 'Tidak memiliki kolom metrik numerik',
        }
      }
      const recommended = s.hasQuantitativeKpi || s.hasAnomalyFlag || s.hasIncidentOrRiskSignals
      return {
        targetModule: entry.module,
        targetScope: entry.scope,
        actionType: entry.actionType,
        supported: true,
        compatible: true,
        recommended,
        priority: s.hasQuantitativeKpi ? 86 : 72,
        importMode: entry.importMode,
        route: entry.route,
        reason: `Cocok: metrik numerik (${s.amountColumn || s.metricColumns[0] || s.flagColumns[0]}) dapat diturunkan menjadi KPI Strategic Audit Plan (konfirmasi target via banner).`,
        shortReason: 'Cocok',
      }
    }

    case 'annual_plan': {
      const ok = s.hasDimensionColumn || s.hasNumericMeasure || s.hasEvidenceIdentifier
      if (!ok) {
        return {
          targetModule: entry.module,
          targetScope: entry.scope,
          actionType: entry.actionType,
          supported: true,
          compatible: false,
          recommended: false,
          priority: 0,
          importMode: entry.importMode,
          route: entry.route,
          reason: `Tabel '${table.tableName}' tidak memiliki kolom entitas/dimensi atau baris audit yang dapat dijadwalkan ke Annual Audit Plan.`,
          shortReason: 'Tidak memiliki dimensi/baris audit',
        }
      }
      const recommended = s.hasDimensionColumn && (s.hasNumericMeasure || s.hasIncidentOrRiskSignals || s.isReferenceTable)
      return {
        targetModule: entry.module,
        targetScope: entry.scope,
        actionType: entry.actionType,
        supported: true,
        compatible: true,
        recommended,
        priority: s.isReferenceTable ? 88 : 85,
        importMode: entry.importMode,
        route: entry.route,
        reason: s.isReferenceTable
          ? `Cocok sebagai referensi dimensi entitas (${s.dimensionColumn}) untuk jadwal Annual Audit Plan (konfirmasi manual via banner).`
          : `Cocok: baris/entitas (${s.dimensionColumn || s.primaryKeyColumn}) dapat dijadwalkan sebagai Draft Annual Audit Plan via banner.`,
        shortReason: 'Cocok',
      }
    }

    case 'working_paper': {
      const ok =
        !s.isReferenceTable &&
        s.hasEvidenceIdentifier &&
        (s.hasMonetaryMeasure || s.hasEventDate || s.hasIncidentOrRiskSignals || s.hasAnomalyFlag || s.hasRefNumber)
      if (!ok) {
        return {
          targetModule: entry.module,
          targetScope: entry.scope,
          actionType: entry.actionType,
          supported: true,
          compatible: false,
          recommended: false,
          priority: 0,
          importMode: entry.importMode,
          route: entry.route,
          reason: s.isReferenceTable
            ? `Tabel '${table.tableName}' adalah tabel referensi master dan tidak memiliki bukti transaksi/insiden untuk uji petik sampel Kertas Kerja (F-03).`
            : `Tabel '${table.tableName}' tidak memiliki pengenal bukti transaksi/insiden atau tanggal kejadian yang layak untuk sampel Kertas Kerja (F-03).`,
          shortReason: s.isReferenceTable
            ? 'Tabel referensi master tanpa bukti uji petik'
            : 'Tidak memiliki pengenal bukti transaksi/insiden',
        }
      }
      const recommended = s.hasRefNumber || s.hasMonetaryMeasure || s.hasIncidentOrRiskSignals
      return {
        targetModule: entry.module,
        targetScope: entry.scope,
        actionType: entry.actionType,
        supported: true,
        compatible: true,
        recommended,
        // Intentionally lower priority than Annual/Strategic Plan so Working Paper is never blindly chosen as default
        priority: s.hasRefNumber ? 70 : 62,
        importMode: entry.importMode,
        route: entry.route,
        reason: `Cocok: memiliki referensi bukti (${s.refColumn || s.primaryKeyColumn}) untuk pre-fill populasi & sampel uji petik Form F-03 Working Paper.`,
        shortReason: 'Cocok',
      }
    }

    case 'ai_entity_prediction': {
      const ok = s.hasDimensionColumn || s.hasIncidentOrRiskSignals || s.hasRiskTierColumn
      if (!ok) {
        return {
          targetModule: entry.module,
          targetScope: entry.scope,
          actionType: entry.actionType,
          supported: true,
          compatible: false,
          recommended: false,
          priority: 0,
          importMode: entry.importMode,
          route: entry.route,
          reason: `Tabel '${table.tableName}' tidak memiliki kolom entitas/cabang atau sinyal risiko untuk AI: Entity Prediction.`,
          shortReason: 'Tidak memiliki kolom entitas/cabang',
        }
      }
      const recommended = s.hasDimensionColumn && (s.hasIncidentOrRiskSignals || s.hasRiskTierColumn || s.hasMonetaryMeasure)
      return {
        targetModule: entry.module,
        targetScope: entry.scope,
        actionType: entry.actionType,
        supported: true,
        compatible: true,
        recommended,
        priority: s.isReferenceTable ? 84 : (s.archetype === 'incident_risk' ? 92 : 80),
        importMode: entry.importMode,
        route: entry.route,
        reason: `Cocok: dimensi entitas (${s.dimensionColumn || s.primaryKeyColumn})${s.hasRiskTierColumn ? ' dan tier risiko' : ''} dapat digunakan untuk prediksi skor risiko entitas.`,
        shortReason: 'Cocok',
      }
    }

    case 'ai_anomaly': {
      const ok =
        !s.isReferenceTable &&
        (s.hasNumericMeasure || s.hasAnomalyFlag || (s.hasEventDate && s.hasEvidenceIdentifier && !s.hasNarrativeText))
      if (!ok) {
        return {
          targetModule: entry.module,
          targetScope: entry.scope,
          actionType: entry.actionType,
          supported: true,
          compatible: false,
          recommended: false,
          priority: 0,
          importMode: entry.importMode,
          route: entry.route,
          reason: s.isReferenceTable
            ? `Tabel '${table.tableName}' adalah tabel referensi master tanpa kolom metrik/ukuran numerik (nominal, skor risiko, atau flag anomali), sehingga tidak kompatibel untuk AI: Anomaly Detection.`
            : `Tabel '${table.tableName}' tidak memiliki kolom nominal, metrik numerik, atau flag anomali untuk AI: Anomaly Detection.`,
          shortReason: s.isReferenceTable
            ? 'Tabel referensi master tanpa kolom nominal/anomali'
            : 'Tidak memiliki kolom nominal/flag anomali',
        }
      }
      const recommended = s.hasMonetaryMeasure || s.hasAnomalyFlag
      return {
        targetModule: entry.module,
        targetScope: entry.scope,
        actionType: entry.actionType,
        supported: true,
        compatible: true,
        recommended,
        priority: s.hasMonetaryMeasure && s.hasAnomalyFlag ? 96 : (s.hasMonetaryMeasure ? 92 : 82),
        importMode: entry.importMode,
        route: entry.route,
        reason: `Cocok: terdeteksi ${[s.amountColumn ? `nominal (${s.amountColumn})` : '', s.flagColumns[0] ? `flag (${s.flagColumns[0]})` : ''].filter(Boolean).join(' & ') || 'metrik numerik'} untuk deteksi anomali AI.`,
        shortReason: 'Cocok',
      }
    }

    case 'ai_nlp': {
      const ok = !s.isReferenceTable && s.hasNarrativeText
      if (!ok) {
        return {
          targetModule: entry.module,
          targetScope: entry.scope,
          actionType: entry.actionType,
          supported: true,
          compatible: false,
          recommended: false,
          priority: 0,
          importMode: entry.importMode,
          route: entry.route,
          reason: `Tabel '${table.tableName}' tidak memiliki kolom teks naratif/dokumen (mis. notes, description, audit_note, finding_summary) untuk AI: NLP Document Analysis.`,
          shortReason: 'Tidak memiliki kolom teks naratif/catatan',
        }
      }
      const recommended = s.hasContractOrAuditText || s.archetype === 'narrative_document' || s.archetype === 'incident_risk'
      return {
        targetModule: entry.module,
        targetScope: entry.scope,
        actionType: entry.actionType,
        supported: true,
        compatible: true,
        recommended,
        priority: s.hasContractOrAuditText || s.archetype === 'narrative_document' ? 95 : 74,
        importMode: entry.importMode,
        route: entry.route,
        reason: `Cocok: terdeteksi kolom teks naratif (${s.textColumns.slice(0, 2).join(', ')}) yang dapat dianalisis oleh model NLP.`,
        shortReason: 'Cocok',
      }
    }

    case 'ai_kpi_forecast': {
      const ok =
        !s.isReferenceTable &&
        s.hasNumericMeasure &&
        (s.hasEventDate || s.hasQuantitativeKpi || s.hasMonetaryMeasure || s.hasAnomalyFlag)
      if (!ok) {
        return {
          targetModule: entry.module,
          targetScope: entry.scope,
          actionType: entry.actionType,
          supported: true,
          compatible: false,
          recommended: false,
          priority: 0,
          importMode: entry.importMode,
          route: entry.route,
          reason: s.isReferenceTable
            ? `Tabel '${table.tableName}' adalah tabel referensi master tanpa kolom metrik/ukuran numerik atau periode historis untuk AI: KPI & Loss Forecast.`
            : `Tabel '${table.tableName}' tidak memiliki kolom metrik numerik atau nominal untuk peramalan KPI & Loss Forecast.`,
          shortReason: s.isReferenceTable
            ? 'Tabel referensi master tanpa metrik historis'
            : 'Tidak memiliki kolom metrik numerik',
        }
      }
      const recommended = s.hasQuantitativeKpi || (s.hasMonetaryMeasure && s.hasEventDate)
      return {
        targetModule: entry.module,
        targetScope: entry.scope,
        actionType: entry.actionType,
        supported: true,
        compatible: true,
        recommended,
        priority: s.hasQuantitativeKpi ? 94 : 78,
        importMode: entry.importMode,
        route: entry.route,
        reason: `Cocok: terdeteksi metrik kuantitatif (${s.metricColumns[0] || s.amountColumn || s.flagColumns[0]})${s.eventDateColumn ? ` & tanggal (${s.eventDateColumn})` : ''} untuk peramalan KPI.`,
        shortReason: 'Cocok',
      }
    }

    default:
      return {
        targetModule: entry.module,
        targetScope: entry.scope,
        actionType: 'unsupported',
        supported: false,
        compatible: false,
        recommended: false,
        priority: 0,
        importMode: 'unsupported',
        route: entry.route,
        reason: `Modul '${entry.module}' belum didukung.`,
        shortReason: 'Belum didukung',
      }
  }
}

export function evaluateAllModulesForTable(
  table: TableCompatibilityInput
): ModuleCompatibilityEvaluation[] {
  return MODULE_CATALOG.map(entry => evaluateModuleCompatibility(table, entry.module))
}

export function getRecommendedMappingsForTable(
  table: TableCompatibilityInput
): RecommendedMappingSpec[] {
  return evaluateAllModulesForTable(table)
    .filter(e => e.compatible && e.recommended)
    .sort((a, b) => b.priority - a.priority)
    .map(e => ({
      targetScope: e.targetScope,
      targetModule: e.targetModule,
      priority: e.priority,
      reason: e.reason,
      importMode: e.importMode,
    }))
}

export function getNextRecommendedMappingForTable(
  table: TableCompatibilityInput,
  existingModules: string[] = []
): RecommendedMappingSpec {
  const existingSet = new Set(existingModules.map(m => (m || '').trim().toLowerCase()))
  const evaluations = evaluateAllModulesForTable(table)
    .filter(e => e.compatible)
    .sort((a, b) => {
      if (a.recommended !== b.recommended) return a.recommended ? -1 : 1
      return b.priority - a.priority
    })

  const nextUnmapped = evaluations.find(e => !existingSet.has(e.targetModule.toLowerCase()))
  const picked = nextUnmapped || evaluations[0]

  if (picked) {
    return {
      targetScope: picked.targetScope,
      targetModule: picked.targetModule,
      priority: picked.priority,
      reason: picked.reason,
      importMode: picked.importMode,
    }
  }

  return {
    targetScope: 'risk_management',
    targetModule: 'Audit Universe',
    priority: 50,
    reason: 'Default kompatibel untuk entitas Audit Universe.',
    importMode: 'auto',
  }
}

export function getScopeOptionsForTable(table?: TableCompatibilityInput): ScopeOptionForTable[] {
  const scopes = ['risk_management', 'audit_features', 'qar_features', 'data_analytics']
  return scopes.map(scope => {
    const entries = MODULE_CATALOG.filter(m => m.scope === scope)
    if (scope === 'qar_features') {
      return {
        label: `${SCOPE_LABELS[scope]} (Belum didukung)`,
        value: scope,
        disabled: true,
        compatibleCount: 0,
      }
    }
    if (!table) {
      const supportedCount = entries.filter(e => e.supported).length
      return {
        label: SCOPE_LABELS[scope] || scope,
        value: scope,
        disabled: supportedCount === 0,
        compatibleCount: supportedCount,
      }
    }
    const compatibleCount = entries.filter(e => evaluateModuleCompatibility(table, e.module).compatible).length
    if (compatibleCount === 0) {
      return {
        label: `${SCOPE_LABELS[scope]} (Tidak ada modul cocok)`,
        value: scope,
        disabled: true,
        compatibleCount: 0,
      }
    }
    return {
      label: `${SCOPE_LABELS[scope]} (${compatibleCount} modul cocok)`,
      value: scope,
      disabled: false,
      compatibleCount,
    }
  })
}

export function getModuleOptionsForTableScope(
  scope?: string,
  table?: TableCompatibilityInput
): ScopedModuleOption[] {
  const effectiveScope = scope || 'audit_features'
  const entries = MODULE_CATALOG.filter(m => m.scope === effectiveScope)
  if (entries.length === 0) return []

  if (!table) {
    return entries.map(e => ({
      label: e.supported ? e.module : `${e.module} (Belum didukung)`,
      value: e.module,
      disabled: !e.supported ? true : undefined,
      compatible: e.supported,
      recommended: false,
      importMode: e.importMode,
    }))
  }

  const evaluated = entries.map(e => evaluateModuleCompatibility(table, e.module))

  // Sort: compatible + recommended first, then compatible, then incompatible supported, then unsupported
  evaluated.sort((a, b) => {
    const rank = (x: ModuleCompatibilityEvaluation) => {
      if (x.compatible && x.recommended) return 0
      if (x.compatible) return 1
      if (x.supported) return 2
      return 3
    }
    const rDiff = rank(a) - rank(b)
    if (rDiff !== 0) return rDiff
    return b.priority - a.priority
  })

  return evaluated.map(ev => {
    if (!ev.supported) {
      return {
        label: `${ev.targetModule} (Belum didukung)`,
        value: ev.targetModule,
        disabled: true,
        recommended: false,
        compatible: false,
        reason: ev.reason,
        importMode: ev.importMode,
      }
    }
    if (!ev.compatible) {
      return {
        label: `${ev.targetModule} — Tidak cocok (${ev.shortReason})`,
        value: ev.targetModule,
        disabled: true,
        recommended: false,
        compatible: false,
        reason: ev.reason,
        importMode: ev.importMode,
      }
    }
    return {
      label: ev.recommended ? `★ ${ev.targetModule} (Direkomendasikan)` : ev.targetModule,
      value: ev.targetModule,
      disabled: false,
      recommended: ev.recommended,
      compatible: true,
      reason: ev.reason,
      importMode: ev.importMode,
    }
  })
}

export function getDefaultModuleForTableScope(
  scope?: string,
  table?: TableCompatibilityInput
): string {
  const options = getModuleOptionsForTableScope(scope, table)
  const firstCompatible = options.find(opt => !opt.disabled)
  if (firstCompatible) return firstCompatible.value
  return options[0]?.value || 'Audit Universe'
}
