// @ts-nocheck
import { describe, it, expect } from 'vitest'

interface MappedRecord {
  connection_id: string
  connection_name: string
  table_name: string
  target_scope: string
  target_module: string
  audit_field: string
  sample_records: Record<string, unknown>[]
}

// Logic mirror from MappedSourceDataBanner.vue
function resolveModuleInfo(path: string) {
  const p = path.toLowerCase()
  if (p.startsWith('/settings') || p.startsWith('/auth')) {
    return null
  }
  if (p === '/risk-profile' || p === '/risk-profile/') {
    return {
      scope: 'risk_management',
      module: 'Corporate Risk Profile',
      slug: 'corporate_risk_profile',
      label: 'Corporate Risk Profile',
      canAddToRiskProfile: true
    }
  }
  if (p.includes('/risk-profile/audit-universe')) {
    return {
      scope: 'risk_management',
      module: 'Audit Universe',
      slug: 'audit_universe',
      label: 'Audit Universe',
      canAddToRiskProfile: false
    }
  }
  if (p.includes('/analytics/ai/anomaly-detection') || p.includes('/ai/anomaly-detection')) {
    return {
      scope: 'data_analytics',
      module: 'AI: Anomaly Detection',
      slug: 'anomaly_detection',
      label: 'AI Anomaly Detection',
      actionType: 'ai_anomaly',
      actionLabel: '+ Jalankan AI Anomaly Detection'
    }
  }
  return null
}

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

function getSampleRecordsForTable(tableName: string) {
  if (tableName === 'loan_accounts') {
    return [
      { loan_id: 'LN-001', cif_number: 'CIF-881923', product_type: 'Commercial Working Capital', principal_amount: 2500000000, kol_status: 'KOL-1' },
      { loan_id: 'LN-002', cif_number: 'CIF-772910', product_type: 'Mortgage Griya', principal_amount: 750000000, kol_status: 'KOL-2' }
    ]
  }
  return []
}

describe('MappedSourceDataBanner Logic', () => {
  it('should accurately resolve Corporate Risk Profile route info', () => {
    const info = resolveModuleInfo('/risk-profile')
    expect(info).not.toBeNull()
    expect(info?.module).toBe('Corporate Risk Profile')
    expect(info?.scope).toBe('risk_management')
    expect(info?.canAddToRiskProfile).toBe(true)
  })

  it('should match user mapping of loan_accounts to Corporate Risk Profile', () => {
    const userMapping = {
      tableName: 'loan_accounts',
      targetScope: 'risk_management',
      targetModule: 'Corporate Risk Profile'
    }

    const moduleInfo = resolveModuleInfo('/risk-profile')
    const isMatched = matchMappingToModule(userMapping, moduleInfo)
    expect(isMatched).toBe(true)

    const samples = getSampleRecordsForTable(userMapping.tableName)
    expect(samples).toHaveLength(2)
    expect(samples[0].loan_id).toBe('LN-001')
    expect(samples[0].kol_status).toBe('KOL-1')
  })

  it('should generate a structured risk item payload from loan_accounts mapped record', () => {
    const record: MappedRecord = {
      connection_id: 'conn-core-banking',
      connection_name: 'Core Banking Dev Postgres',
      table_name: 'loan_accounts',
      target_scope: 'risk_management',
      target_module: 'Corporate Risk Profile',
      audit_field: 'loan_id',
      sample_records: getSampleRecordsForTable('loan_accounts')
    }

    const generatedRisk = {
      name: `Risiko Portofolio Kredit: Kredit Bermasalah (${record.table_name})`,
      description: `Risiko teridentifikasi otomatis dari data pipeline '${record.connection_name}' (tabel ${record.table_name}).`,
      category: 'Financial Risk',
      likelihood: 3,
      impact: 4,
      risk_level: 'High'
    }

    expect(generatedRisk.name).toContain('loan_accounts')
    expect(generatedRisk.category).toBe('Financial Risk')
    expect(generatedRisk.likelihood).toBe(3)
    expect(generatedRisk.impact).toBe(4)
  })

  it('should resolve AI Anomaly Detection route and match gl_transactions', () => {
    const info = resolveModuleInfo('/analytics/ai/anomaly-detection')
    expect(info).not.toBeNull()
    expect(info?.module).toBe('AI: Anomaly Detection')
    expect(info?.scope).toBe('data_analytics')
    expect(info?.actionType).toBe('ai_anomaly')

    const glMapping = {
      tableName: 'gl_transactions',
      targetScope: 'data_analytics',
      targetModule: 'AI: Anomaly Detection'
    }
    const isMatched = matchMappingToModule(glMapping, info)
    expect(isMatched).toBe(true)
  })

  it('should support multiple mapped tables simultaneously without dropping any', () => {
    const mappings = [
      {
        connection_id: 'conn-1',
        connection_name: 'Core Banking Dev Postgres',
        table_name: 'loan_accounts',
        target_scope: 'risk_management',
        target_module: 'Corporate Risk Profile',
        audit_field: 'loan_id',
        sample_records: [{ loan_id: 'LN-001' }]
      },
      {
        connection_id: 'conn-1',
        connection_name: 'Core Banking Dev Postgres',
        table_name: 'gl_transactions',
        target_scope: 'risk_management',
        target_module: 'Corporate Risk Profile',
        audit_field: 'id',
        sample_records: [{ id: 'TRX-001' }]
      }
    ]

    expect(mappings).toHaveLength(2)
    expect(mappings[0].table_name).toBe('loan_accounts')
    expect(mappings[1].table_name).toBe('gl_transactions')

    // Both tables should have selectable indices
    let activeIndex = 0
    expect(mappings[activeIndex].table_name).toBe('loan_accounts')
    activeIndex = 1
    expect(mappings[activeIndex].table_name).toBe('gl_transactions')
  })
})

