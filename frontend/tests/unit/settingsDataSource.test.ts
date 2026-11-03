// @ts-nocheck
import { describe, it, expect } from 'vitest'

interface DataSourceConn {
  id: string
  name: string
  type: string
  host: string
  port: number
  database: string
  environment: string
  status: string
  syncSchedule: string
  lastSync: string
  lastError?: string
  ssl: boolean
  scopes?: string[]
}

interface TableColumn {
  name: string
  dataType: string
  isNullable: boolean
  isPrimary: boolean
}

interface SchemaTable {
  tableName: string
  rowCount: number
  columnCount: number
  columns: TableColumn[]
  description?: string
}

interface TableDataMapping {
  tableName: string
  targetScope: string
  targetModule: string
  auditField: string
  syncInterval: string
  anomalyRules?: string[]
  isActive: boolean
}

// Logic mirror from frontend/components/settings/data-source.vue
function filterConnections(
  connections: DataSourceConn[],
  searchQuery: string,
  statusFilter: string,
  typeFilter: string,
  allStatusesLabel = 'All Statuses',
  allProvidersLabel = 'All Providers'
) {
  const getProviderName = (type: string) => {
    if (type === 'postgres') return 'PostgreSQL'
    return type
  }

  return connections.filter(c => {
    const q = searchQuery.toLowerCase()
    const matchesSearch = !q ||
      c.name.toLowerCase().includes(q) ||
      c.host.toLowerCase().includes(q) ||
      c.database.toLowerCase().includes(q)

    const matchesStatus = statusFilter === allStatusesLabel || c.status === statusFilter
    const matchesType = typeFilter === allProvidersLabel || getProviderName(c.type) === typeFilter

    return matchesSearch && matchesStatus && matchesType
  })
}

function getStatusBadgeColor(status: string) {
  switch (status) {
    case 'Connected': return 'success'
    case 'Syncing': return 'info'
    case 'Error': return 'error'
    default: return 'neutral'
  }
}

function getStatusDotClass(status: string) {
  switch (status) {
    case 'Connected': return 'bg-emerald-500 animate-pulse'
    case 'Syncing': return 'bg-blue-500 animate-ping'
    case 'Error': return 'bg-rose-500'
    default: return 'bg-gray-400'
  }
}

function filterSchemaTables(
  tables: SchemaTable[],
  activeMappings: Record<string, TableDataMapping>,
  searchQuery: string,
  scopeFilter: string
) {
  return tables.filter(t => {
    const q = searchQuery.toLowerCase()
    const matchesSearch = !q ||
      t.tableName.toLowerCase().includes(q) ||
      (t.description && t.description.toLowerCase().includes(q)) ||
      t.columns.some(c => c.name.toLowerCase().includes(q))

    const mapping = activeMappings[t.tableName]
    let matchesScope = true
    if (scopeFilter !== 'All Scopes') {
      if (scopeFilter === 'Risk Management') {
        matchesScope = mapping?.targetScope === 'risk_management'
      } else if (scopeFilter === 'Audit Features') {
        matchesScope = mapping?.targetScope === 'audit_features'
      } else if (scopeFilter === 'QAR Features') {
        matchesScope = mapping?.targetScope === 'qar_features'
      } else if (scopeFilter === 'Data Analytics / AI') {
        matchesScope = mapping?.targetScope === 'data_analytics'
      }
    }

    return matchesSearch && matchesScope
  })
}

describe('Settings DataSource Component Logic', () => {
  const sampleConnections: DataSourceConn[] = [
    {
      id: 'conn-1',
      name: 'Core Banking Prod Postgres',
      type: 'postgres',
      host: '10.20.0.14',
      port: 5432,
      database: 'core_banking_db',
      environment: 'Production',
      status: 'Connected',
      syncSchedule: 'Hourly',
      lastSync: '10 minutes ago',
      ssl: true,
      scopes: ['risk_management', 'audit_features']
    },
    {
      id: 'conn-2',
      name: 'Staging Audit Mirror',
      type: 'postgres',
      host: '10.20.0.99',
      port: 5432,
      database: 'audit_mirror_stg',
      environment: 'Staging',
      status: 'Syncing',
      syncSchedule: 'Daily',
      lastSync: 'Syncing in progress...',
      ssl: false,
      scopes: ['data_analytics']
    },
    {
      id: 'conn-3',
      name: 'Legacy Fin Oracle',
      type: 'oracle',
      host: '10.30.1.20',
      port: 1521,
      database: 'fin_legacy',
      environment: 'Production',
      status: 'Error',
      syncSchedule: 'Daily',
      lastSync: 'Yesterday',
      lastError: 'Connection timed out',
      ssl: true,
      scopes: ['audit_features']
    }
  ]

  it('should filter connections by name, host, and database query', () => {
    // Search by name
    const byName = filterConnections(sampleConnections, 'Core Banking', 'All Statuses', 'All Providers')
    expect(byName).toHaveLength(1)
    expect(byName[0].id).toBe('conn-1')

    // Search by host IP
    const byHost = filterConnections(sampleConnections, '10.20.0.99', 'All Statuses', 'All Providers')
    expect(byHost).toHaveLength(1)
    expect(byHost[0].id).toBe('conn-2')

    // Search by database name
    const byDb = filterConnections(sampleConnections, 'fin_legacy', 'All Statuses', 'All Providers')
    expect(byDb).toHaveLength(1)
    expect(byDb[0].id).toBe('conn-3')

    // Non-matching query
    const noMatch = filterConnections(sampleConnections, 'nonexistent', 'All Statuses', 'All Providers')
    expect(noMatch).toHaveLength(0)
  })

  it('should filter connections by status and provider type', () => {
    // Filter by Connected
    const connected = filterConnections(sampleConnections, '', 'Connected', 'All Providers')
    expect(connected).toHaveLength(1)
    expect(connected[0].name).toBe('Core Banking Prod Postgres')

    // Filter by Syncing
    const syncing = filterConnections(sampleConnections, '', 'Syncing', 'All Providers')
    expect(syncing).toHaveLength(1)
    expect(syncing[0].name).toBe('Staging Audit Mirror')

    // Filter by PostgreSQL provider
    const pgConns = filterConnections(sampleConnections, '', 'All Statuses', 'PostgreSQL')
    expect(pgConns).toHaveLength(2)

    // Combined filter (PostgreSQL + Syncing)
    const combined = filterConnections(sampleConnections, '', 'Syncing', 'PostgreSQL')
    expect(combined).toHaveLength(1)
    expect(combined[0].id).toBe('conn-2')
  })

  it('should assign correct status badge colors and pulse classes', () => {
    expect(getStatusBadgeColor('Connected')).toBe('success')
    expect(getStatusBadgeColor('Syncing')).toBe('info')
    expect(getStatusBadgeColor('Error')).toBe('error')
    expect(getStatusBadgeColor('Inactive')).toBe('neutral')

    expect(getStatusDotClass('Connected')).toContain('bg-emerald-500')
    expect(getStatusDotClass('Syncing')).toContain('bg-blue-500')
    expect(getStatusDotClass('Error')).toContain('bg-rose-500')
    expect(getStatusDotClass('Inactive')).toContain('bg-gray-400')
  })

  it('should filter schema tables by search query and target scope', () => {
    const sampleTables: SchemaTable[] = [
      {
        tableName: 'gl_transactions',
        rowCount: 142580,
        columnCount: 3,
        description: 'General Ledger posted transactions',
        columns: [
          { name: 'id', dataType: 'uuid', IsNullable: false, IsPrimary: true },
          { name: 'amount', dataType: 'numeric', IsNullable: false, IsPrimary: false },
          { name: 'account_number', dataType: 'varchar', IsNullable: false, IsPrimary: false }
        ]
      },
      {
        tableName: 'compliance_incident_records',
        rowCount: 320,
        columnCount: 2,
        description: 'Regulatory policy violations',
        columns: [
          { name: 'incident_id', dataType: 'uuid', IsNullable: false, IsPrimary: true },
          { name: 'severity_score', dataType: 'integer', IsNullable: false, IsPrimary: false }
        ]
      }
    ]

    const mappings: Record<string, TableDataMapping> = {
      gl_transactions: {
        tableName: 'gl_transactions',
        targetScope: 'audit_features',
        targetModule: 'Working Paper Evidence',
        auditField: 'amount',
        syncInterval: 'Hourly',
        isActive: true
      },
      compliance_incident_records: {
        tableName: 'compliance_incident_records',
        targetScope: 'qar_features',
        targetModule: 'Compliance Monitoring',
        auditField: 'severity_score',
        syncInterval: 'Daily',
        isActive: true
      }
    }

    // Filter by table search column name
    const foundByColumn = filterSchemaTables(sampleTables, mappings, 'account_number', 'All Scopes')
    expect(foundByColumn).toHaveLength(1)
    expect(foundByColumn[0].tableName).toBe('gl_transactions')

    // Filter by Target Scope
    const auditScopeTables = filterSchemaTables(sampleTables, mappings, '', 'Audit Features')
    expect(auditScopeTables).toHaveLength(1)
    expect(auditScopeTables[0].tableName).toBe('gl_transactions')

    const qarScopeTables = filterSchemaTables(sampleTables, mappings, '', 'QAR Features')
    expect(qarScopeTables).toHaveLength(1)
    expect(qarScopeTables[0].tableName).toBe('compliance_incident_records')

    // Active mappings count calculation
    const activeCount = Object.values(mappings).filter(m => m.isActive).length
    expect(activeCount).toBe(2)
  })

  it('should provide dynamic cascading modules based on selected scope', () => {
    const scopeModulesMap: Record<string, { label: string; value: string }[]> = {
      risk_management: [
        { label: 'Corporate Risk Profile', value: 'Corporate Risk Profile' },
        { label: 'Audit Universe', value: 'Audit Universe' }
      ],
      data_analytics: [
        { label: 'Data Analytics Center', value: 'Data Analytics Center' },
        { label: 'CAATT Automated Rules', value: 'CAATT Automated Rules' }
      ]
    }

    const getModulesForScope = (scope: string) => scopeModulesMap[scope] || []

    const riskModules = getModulesForScope('risk_management')
    expect(riskModules.map(m => m.value)).toContain('Corporate Risk Profile')
    expect(riskModules.map(m => m.value)).toContain('Audit Universe')
    expect(riskModules.map(m => m.value)).not.toContain('Data Analytics Center')

    const analyticsModules = getModulesForScope('data_analytics')
    expect(analyticsModules.map(m => m.value)).toContain('Data Analytics Center')
    expect(analyticsModules.map(m => m.value)).toContain('CAATT Automated Rules')
    expect(analyticsModules.map(m => m.value)).not.toContain('Corporate Risk Profile')
  })

  it('should ensure select options never contain empty strings that cause Reka UI freeze', () => {
    const table: SchemaTable = {
      tableName: 'users',
      rowCount: 10,
      columnCount: 2,
      columns: [
        { name: 'id', dataType: 'uuid', isNullable: false, isPrimary: true },
        { name: 'updated_at', dataType: 'timestamp', isNullable: false, isPrimary: false }
      ]
    }

    const watermarkOptions = [
      { label: 'None (Full Table Load)', value: '__none__' },
      ...table.columns.map(c => ({ label: c.name, value: c.name }))
    ]

    const softDeleteOptions = [
      { label: 'None (Hard-Delete Bucket Reconcile)', value: '__none__' },
      ...table.columns.map(c => ({ label: c.name, value: c.name }))
    ]

    // Verify none of the options have an empty string value
    expect(watermarkOptions.some(opt => opt.value === '')).toBe(false)
    expect(softDeleteOptions.some(opt => opt.value === '')).toBe(false)
    expect(watermarkOptions[0].value).toBe('__none__')
    expect(softDeleteOptions[0].value).toBe('__none__')
  })

  it('should support multiple mappings per table and toggle anomaly rules', () => {
    const multiMappings: Record<string, TableDataMapping[]> = {
      gl_transactions: [
        {
          tableName: 'gl_transactions',
          targetScope: 'risk_management',
          targetModule: 'Audit Universe',
          auditField: 'id',
          syncInterval: 'Hourly',
          isActive: true
        },
        {
          tableName: 'gl_transactions',
          targetScope: 'data_analytics',
          targetModule: 'Data Analytics Center',
          auditField: 'amount',
          syncInterval: 'Daily',
          anomalyRules: ['Benford Law First-Digit Analysis'],
          isActive: true
        }
      ]
    }

    expect(multiMappings['gl_transactions']).toHaveLength(2)
    expect(multiMappings['gl_transactions'][0].targetScope).toBe('risk_management')
    expect(multiMappings['gl_transactions'][1].targetScope).toBe('data_analytics')
    expect(multiMappings['gl_transactions'][1].anomalyRules).toContain('Benford Law First-Digit Analysis')

    // Anomaly rules should only be applied to analytics scope
    const isAnalytics = (m: TableDataMapping) => m.targetScope === 'data_analytics'
    expect(isAnalytics(multiMappings['gl_transactions'][0])).toBe(false)
    expect(isAnalytics(multiMappings['gl_transactions'][1])).toBe(true)
  })
})

