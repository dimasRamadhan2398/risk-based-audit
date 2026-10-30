import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getRiskServiceBaseUrl } from '~/composables/useApiUrl'
import { extractErrorMessage } from '~/utils/error'

export interface StandardAuditUniverse {
  id: string
  name: string
  parent_id?: string
  category?: string
  children?: StandardAuditUniverse[]
}

export interface CorporateAuditUniverse {
  id: string
  standard_audit_universe_id?: string
  standard_audit_universe?: StandardAuditUniverse
  name: string
  parent_id?: string
  /** Owning department (master-service department UUID). null/absent = company-wide, editable by full writers only. */
  department_id?: string | null
  /** Department name snapshot sent by the server, when available. */
  department_name?: string | null
  children?: CorporateAuditUniverse[]
}

export interface AuditUniverseRiskScore {
  id: string
  audit_universe_year_id: string
  corporate_risk_factor_id: string
  corporate_risk_factor?: {
    id: string
    weight: number
    standard_risk_factor?: {
      id: string
      name: string
      description: string
    }
  }
  score: number
  weighted_score: number
}

export interface AuditUniverseYear {
  id: string
  corporate_audit_universe_id: string
  corporate_audit_universe?: CorporateAuditUniverse
  year: number
  risk_index: number
  risk_level: string
  audit_priority: boolean
  /** ISO timestamp of the last scoring; null = never scored. Missing on older backends (treated as scored). */
  scored_at?: string | null
  risk_scores?: AuditUniverseRiskScore[]
}

export const useAuditUniverseStore = defineStore('audit-universe', () => {
  const config = useRuntimeConfig()

  const standardUniverse = ref<StandardAuditUniverse[]>([])
  const corporateUniverse = ref<CorporateAuditUniverse[]>([])
  const yearlyUniverse = ref<AuditUniverseYear[]>([])
  const loading = ref(false)
  const errorMsg = ref('')

  const fetchStandardUniverse = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/audit-universe/standard`)
      if (response && response.success) {
        standardUniverse.value = response.data
      }
    } catch (error: any) {
      console.error('Failed to fetch standard audit universe, using default seeds:', error)
      standardUniverse.value = [
        { id: 'u1', name: 'Financial Management', category: 'Finance', default_risk_weight: 0.25, created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: 'u2', name: 'Information Technology Governance', category: 'IT', default_risk_weight: 0.25, created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: 'u3', name: 'Plant Operations & Supply Chain', category: 'Operations', default_risk_weight: 0.25, created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: 'u4', name: 'Regulatory Compliance & Legal', category: 'Compliance', default_risk_weight: 0.25, created_at: '2026-01-01', updated_at: '2026-01-01' }
      ]
      errorMsg.value = ''
    } finally {
      loading.value = false
    }
  }

  const fetchCorporateUniverse = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/audit-universe/corporate`)
      if (response && response.success) {
        corporateUniverse.value = response.data
      }
    } catch (error: any) {
      console.error('Failed to fetch corporate audit universe, using default seeds:', error)
      corporateUniverse.value = [
        {
          id: 'corp-root',
          name: 'PT AIFL Indonesia (Holding)',
          children: [
            {
              id: 'corp-fin',
              name: 'Divisi Keuangan & Akuntansi',
              parent_id: 'corp-root',
              children: [
                { id: 'corp-fin-1', name: 'Pengelolaan Kas & Pembayaran (Disbursement)', parent_id: 'corp-fin' },
                { id: 'corp-fin-2', name: 'Pelaporan Keuangan & Perpajakan', parent_id: 'corp-fin' }
              ]
            },
            {
              id: 'corp-it',
              name: 'Divisi Teknologi Informasi',
              parent_id: 'corp-root',
              children: [
                { id: 'corp-it-1', name: 'Keamanan ERP & Manajemen Hak Akses', parent_id: 'corp-it' },
                { id: 'corp-it-2', name: 'Pusat Data & Disaster Recovery Center', parent_id: 'corp-it' }
              ]
            },
            {
              id: 'corp-ops',
              name: 'Divisi Operasional & Rantai Pasok',
              parent_id: 'corp-root',
              children: [
                { id: 'corp-ops-1', name: 'Pengadaan Barang & Jasa (Procurement)', parent_id: 'corp-ops' },
                { id: 'corp-ops-2', name: 'Manajemen Pergudangan & Logistik', parent_id: 'corp-ops' }
              ]
            }
          ]
        }
      ]
      errorMsg.value = ''
    } finally {
      loading.value = false
    }
  }

  const saveCorporateNode = async (payload: { id?: string; name: string; parent_id?: string; standard_audit_universe_id?: string; department_id?: string | null }) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/audit-universe/corporate`, {
        method: 'POST',
        body: payload
      })
      if (response && response.success) {
        await fetchCorporateUniverse()
        return response.data
      }
      return null
    } catch (error: any) {
      console.error('Failed to save corporate audit universe node:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to save corporate audit universe node.')
      return null
    } finally {
      loading.value = false
    }
  }

  const deleteCorporateNode = async (id: string) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/audit-universe/corporate/${id}`, {
        method: 'DELETE'
      })
      if (response && response.success) {
        await fetchCorporateUniverse()
        return true
      }
      return false
    } catch (error: any) {
      console.error('Failed to delete corporate audit universe node:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to delete corporate audit universe node.')
      return false
    } finally {
      loading.value = false
    }
  }

  const syncMasterData = async (companies?: string[], locations?: string[]) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/audit-universe/corporate/sync-master`, {
        method: 'POST',
        body: {
          companies: companies || [],
          locations: locations || []
        }
      })
      if (response && response.success) {
        corporateUniverse.value = response.data
        return true
      }
      return false
    } catch (error: any) {
      console.error('Failed to sync master data to corporate audit universe:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to sync master data.')
      return false
    } finally {
      loading.value = false
    }
  }

  const fetchYearlyUniverse = async (year: number) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/audit-universe/year/${year}`)
      if (response && response.success) {
        yearlyUniverse.value = response.data
      }
    } catch (error: any) {
      console.error('Failed to fetch yearly audit universe:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to fetch yearly audit universe.')
    } finally {
      loading.value = false
    }
  }

  const establishYearlyUniverse = async (year: number, corporateAuditUniverseIDs: string[]) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/audit-universe/year/${year}`, {
        method: 'POST',
        body: { corporate_audit_universe_ids: corporateAuditUniverseIDs }
      })
      if (response && response.success) {
        yearlyUniverse.value = response.data
        return true
      }
      return false
    } catch (error: any) {
      console.error('Failed to establish yearly universe:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to establish yearly universe.')
      return false
    } finally {
      loading.value = false
    }
  }

  const scoreYearlyEntity = async (year: number, payload: { audit_universe_year_id: string; scores: { corporate_risk_factor_id: string; score: number }[] }) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/audit-universe/year/${year}/score`, {
        method: 'POST',
        body: payload
      })
      if (response && response.success) {
        // Refresh local list
        await fetchYearlyUniverse(year)
        return response.data
      }
      return null
    } catch (error: any) {
      console.error('Failed to score yearly entity:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to save risk score.')
      return null
    } finally {
      loading.value = false
    }
  }

  return {
    standardUniverse,
    corporateUniverse,
    yearlyUniverse,
    loading,
    errorMsg,
    fetchStandardUniverse,
    fetchCorporateUniverse,
    saveCorporateNode,
    deleteCorporateNode,
    syncMasterData,
    fetchYearlyUniverse,
    establishYearlyUniverse,
    scoreYearlyEntity
  }
})
