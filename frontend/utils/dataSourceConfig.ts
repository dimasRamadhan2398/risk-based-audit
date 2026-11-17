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
