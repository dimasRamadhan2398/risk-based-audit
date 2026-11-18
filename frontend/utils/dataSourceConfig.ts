export interface DataSourceColumnMeta {
  name: string
  dataType: string
  isNullable: boolean
  isPrimary: boolean
}

export interface DataSourceTableMeta {
  tableName: string
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

export interface TargetModuleOption {
  label: string
  value: string
  disabled?: boolean
}

export const SUPPORTED_TARGET_MODULES = new Set<string>([
  'Corporate Risk Profile',
  'Audit Universe',
  'Strategic Audit Plan',
  'Annual Audit Plan',
  'Working Paper Evidence',
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
    { label: 'Risk Appetite Statement (Belum didukung)', value: 'Risk Appetite Statement', disabled: true },
    { label: 'Risk Factors (Belum didukung)', value: 'Risk Factors', disabled: true },
    { label: 'Audit Priority (Belum didukung)', value: 'Audit Priority', disabled: true },
    { label: 'Risk Control Matrix (RCM) (Belum didukung)', value: 'Risk Control Matrix (RCM)', disabled: true },
    { label: 'Risk Mitigation (Belum didukung)', value: 'Risk Mitigation', disabled: true },
  ],
  audit_features: [
    { label: 'Strategic Audit Plan', value: 'Strategic Audit Plan' },
    { label: 'Annual Audit Plan', value: 'Annual Audit Plan' },
    { label: 'Working Paper Evidence', value: 'Working Paper Evidence' },
    { label: 'Audit Activity Plan (Belum didukung)', value: 'Audit Activity Plan', disabled: true },
    { label: 'Assignment Letter (Belum didukung)', value: 'Assignment Letter', disabled: true },
    { label: 'Audit Execution Status (Belum didukung)', value: 'Audit Execution Status', disabled: true },
    { label: 'Audit Fieldwork (Belum didukung)', value: 'Audit Fieldwork', disabled: true },
    { label: 'Internal Audit Performance (KPI) (Belum didukung)', value: 'Internal Audit Performance (KPI)', disabled: true },
    { label: 'Audit Result Report (LHA) (Belum didukung)', value: 'Audit Result Report (LHA)', disabled: true },
    { label: 'Action Taken Report (ATR) (Belum didukung)', value: 'Action Taken Report (ATR)', disabled: true },
  ],
  qar_features: [
    { label: 'Quality Assurance Dashboard (Belum didukung)', value: 'Quality Assurance Dashboard', disabled: true },
    { label: 'Periodic Self-Assessment (Belum didukung)', value: 'Periodic Self-Assessment', disabled: true },
    { label: 'SAIV Validation (Belum didukung)', value: 'SAIV Validation', disabled: true },
    { label: 'IACM Assessment (Belum didukung)', value: 'IACM Assessment', disabled: true },
    { label: 'Consulting Service (Belum didukung)', value: 'Consulting Service', disabled: true },
  ],
  data_analytics: [
    { label: 'AI: Entity Prediction', value: 'AI: Entity Prediction' },
    { label: 'AI: Anomaly Detection', value: 'AI: Anomaly Detection' },
    { label: 'AI: NLP Document Analysis', value: 'AI: NLP Document Analysis' },
    { label: 'AI: KPI & Loss Forecast', value: 'AI: KPI & Loss Forecast' },
  ],
}

export function getModulesForTargetScope(scope?: string): TargetModuleOption[] {
  return SCOPE_MODULES_MAP[scope || 'audit_features'] || SCOPE_MODULES_MAP.audit_features || []
}

export function getDefaultModuleForTargetScope(scope?: string): string {
  const mods = getModulesForTargetScope(scope)
  const firstSupported = mods.find(m => !m.disabled)
  return firstSupported?.value || mods[0]?.value || 'Working Paper Evidence'
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
