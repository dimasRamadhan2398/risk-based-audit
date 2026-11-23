import {
  getDefaultModuleForTableScope,
  getModuleOptionsForTableScope,
  type ModuleImportMode,
  type ScopedModuleOption,
  type TableCompatibilityInput,
} from '~/utils/moduleCompatibility'

export {
  analyzeTableSchema,
  evaluateAllModulesForTable,
  evaluateModuleCompatibility,
  getDefaultModuleForTableScope,
  getModuleOptionsForTableScope,
  getNextRecommendedMappingForTable,
  getRecommendedMappingsForTable,
  getScopeOptionsForTable,
  getTableArchetypeMeta,
  targetModuleToImportMode,
  targetModuleToRoute,
  type ModuleCompatibilityEvaluation,
  type ModuleImportMode,
  type RecommendedMappingSpec,
  type ScopeOptionForTable,
  type ScopedModuleOption,
  type TableArchetype,
  type TableCompatibilityInput,
  type TableSchemaSignals,
} from '~/utils/moduleCompatibility'

export interface DataSourceColumnMeta {
  name: string
  dataType: string
  isNullable: boolean
  isPrimary: boolean
}

export interface DataSourceTableMeta {
  tableName: string
  rowCount?: number
  columns: DataSourceColumnMeta[]
}

export const SYNC_FREQUENCY_OPTIONS = [
  'Every 15 mins',
  'Hourly',
  'Daily',
  'Manual Only',
] as const

export type SyncFrequencyOption = (typeof SYNC_FREQUENCY_OPTIONS)[number]

export function normalizeSyncSchedule(schedule?: string): SyncFrequencyOption {
  const s = (schedule || '').trim().toLowerCase()
  if (
    s === 'real-time (cdc)' ||
    s === 'realtime (cdc)' ||
    s === 'real-time' ||
    s === 'realtime' ||
    s === 'every 15 mins' ||
    s === 'every_15_mins' ||
    s === '15m'
  ) {
    return 'Every 15 mins'
  }
  if (s === 'daily' || s === '24h' || s === '1d') {
    return 'Daily'
  }
  if (s === 'manual only' || s === 'manual_only' || s === 'manual') {
    return 'Manual Only'
  }
  return 'Hourly'
}

export function shouldShowAnomalyRules(targetScope?: string, targetModule?: string): boolean {
  return targetScope === 'data_analytics' && targetModule === 'AI: Anomaly Detection'
}

export interface TargetModuleOption extends ScopedModuleOption {
  label: string
  value: string
  disabled?: boolean
  recommended?: boolean
  compatible?: boolean
  reason?: string
  importMode?: ModuleImportMode
}

export const SUPPORTED_TARGET_MODULES = new Set<string>([
  'Corporate Risk Profile',
  'Audit Universe',
  'Risk Appetite Statement',
  'Risk Factors',
  'Audit Priority',
  'Risk Control Matrix (RCM)',
  'Risk Mitigation',
  'Strategic Audit Plan',
  'Annual Audit Plan',
  'Working Paper Evidence',
  'Audit Activity Plan',
  'Assignment Letter',
  'Audit Execution Status',
  'Audit Fieldwork',
  'Internal Audit Performance (KPI)',
  'Audit Result Report (LHA)',
  'Action Taken Report (ATR)',
  'Quality Assurance Dashboard',
  'Periodic Self-Assessment',
  'SAIV Validation',
  'IACM Assessment',
  'Consulting Service',
  'AI: Entity Prediction',
  'AI: Anomaly Detection',
  'AI: NLP Document Analysis',
  'AI: KPI & Loss Forecast',
])

export function isTargetModuleSupported(targetModule?: string): boolean {
  if (!targetModule) return false
  return SUPPORTED_TARGET_MODULES.has(targetModule.trim())
}

export const SCOPE_MODULES_MAP: Record<string, TargetModuleOption[]> = {
  risk_management: [
    { label: 'Corporate Risk Profile', value: 'Corporate Risk Profile' },
    { label: 'Audit Universe', value: 'Audit Universe' },
    { label: 'Risk Appetite Statement', value: 'Risk Appetite Statement' },
    { label: 'Risk Factors', value: 'Risk Factors' },
    { label: 'Audit Priority', value: 'Audit Priority' },
    { label: 'Risk Control Matrix (RCM)', value: 'Risk Control Matrix (RCM)' },
    { label: 'Risk Mitigation', value: 'Risk Mitigation' },
  ],
  audit_features: [
    { label: 'Strategic Audit Plan', value: 'Strategic Audit Plan' },
    { label: 'Annual Audit Plan', value: 'Annual Audit Plan' },
    { label: 'Working Paper Evidence', value: 'Working Paper Evidence' },
    { label: 'Audit Activity Plan', value: 'Audit Activity Plan' },
    { label: 'Assignment Letter', value: 'Assignment Letter' },
    { label: 'Audit Execution Status', value: 'Audit Execution Status' },
    { label: 'Audit Fieldwork', value: 'Audit Fieldwork' },
    { label: 'Internal Audit Performance (KPI)', value: 'Internal Audit Performance (KPI)' },
    { label: 'Audit Result Report (LHA)', value: 'Audit Result Report (LHA)' },
    { label: 'Action Taken Report (ATR)', value: 'Action Taken Report (ATR)' },
  ],
  qar_features: [
    { label: 'Quality Assurance Dashboard', value: 'Quality Assurance Dashboard' },
    { label: 'Periodic Self-Assessment', value: 'Periodic Self-Assessment' },
    { label: 'SAIV Validation', value: 'SAIV Validation' },
    { label: 'IACM Assessment', value: 'IACM Assessment' },
    { label: 'Consulting Service', value: 'Consulting Service' },
  ],
  data_analytics: [
    { label: 'AI: Entity Prediction', value: 'AI: Entity Prediction' },
    { label: 'AI: Anomaly Detection', value: 'AI: Anomaly Detection' },
    { label: 'AI: NLP Document Analysis', value: 'AI: NLP Document Analysis' },
    { label: 'AI: KPI & Loss Forecast', value: 'AI: KPI & Loss Forecast' },
  ],
}

export function getModulesForTargetScope(
  scope?: string,
  table?: TableCompatibilityInput
): TargetModuleOption[] {
  if (table) {
    return getModuleOptionsForTableScope(scope, table)
  }
  return SCOPE_MODULES_MAP[scope || 'audit_features'] || SCOPE_MODULES_MAP.audit_features || []
}

export function getDefaultModuleForTargetScope(
  scope?: string,
  table?: TableCompatibilityInput
): string {
  if (table) {
    return getDefaultModuleForTableScope(scope, table)
  }
  const mods = getModulesForTargetScope(scope)
  const firstSupported = mods.find(m => !m.disabled)
  return firstSupported?.value || mods[0]?.value || 'Annual Audit Plan'
}

export function resolveCanonicalEntityAuto(tableName: string, targetModule: string): string {
  const tbl = (tableName || '').toLowerCase()
  const mod = (targetModule || '').toLowerCase()
  if (tbl.includes('trx') || tbl.includes('transact')) return 'transactions'
  if (tbl.includes('loan') || tbl.includes('credit') || tbl.includes('kredit')) return 'loans'
  if (tbl.includes('gl') || tbl.includes('ledger') || tbl.includes('journal')) return 'gl_entries'
  if (tbl.includes('acct') || tbl.includes('account')) return 'accounts'
  if (tbl.includes('cust') || tbl.includes('customer') || tbl.includes('cif')) return 'customers'
  if (mod.includes('anomaly')) return 'ai_training_pool'
  return 'generic'
}

export function resolveAutoMappingMetadata(table: DataSourceTableMeta | undefined, targetModule: string) {
  const cols = table?.columns || []
  const pkCol = cols.find(c => c.isPrimary)?.name || cols[0]?.name || 'id'
  const wmCol = cols.find(
    c =>
      /updated_at|created_at|logged_at|disbursed_at|paid_at|reported_at|date|time|posting/i.test(c.name) ||
      /timestamp|date/i.test(c.dataType)
  )?.name || ''
  const delCol = cols.find(
    c => /^(is_deleted|is_soft_deleted|soft_deleted)$/i.test(c.name) && /bool/i.test(c.dataType)
  )?.name || ''
  return {
    canonicalEntity: resolveCanonicalEntityAuto(table?.tableName || '', targetModule),
    auditField: pkCol,
    watermarkColumn: wmCol,
    softDeleteColumn: delCol,
  }
}
