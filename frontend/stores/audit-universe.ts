import { defineStore } from 'pinia'
import { ref } from 'vue'

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
  risk_scores?: AuditUniverseRiskScore[]
}

export const useAuditUniverseStore = defineStore('audit-universe', () => {
  const config = useRuntimeConfig()

  const standardUniverse = ref<StandardAuditUniverse[]>([])
  const corporateUniverse = ref<CorporateAuditUniverse[]>([])
  const yearlyUniverse = ref<AuditUniverseYear[]>([])
  const loading = ref(false)
  const errorMsg = ref('')

  const getRiskServiceBaseUrl = () => {
    return config.public.riskServiceBaseUrl || 'http://localhost:8080/api/v1'
  }

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
      console.error('Failed to fetch standard audit universe:', error)
      errorMsg.value = error.data?.error || 'Failed to fetch standard audit universe.'
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
      console.error('Failed to fetch corporate audit universe:', error)
      errorMsg.value = error.data?.error || 'Failed to fetch corporate audit universe.'
    } finally {
      loading.value = false
    }
  }

  const saveCorporateNode = async (payload: { id?: string; name: string; parent_id?: string; standard_audit_universe_id?: string }) => {
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
      errorMsg.value = error.data?.error || 'Failed to save corporate audit universe node.'
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
      errorMsg.value = error.data?.error || 'Failed to delete corporate audit universe node.'
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
      errorMsg.value = error.data?.error || 'Failed to fetch yearly audit universe.'
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
      errorMsg.value = error.data?.error || 'Failed to establish yearly universe.'
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
      errorMsg.value = error.data?.error || 'Failed to save risk score.'
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
    fetchYearlyUniverse,
    establishYearlyUniverse,
    scoreYearlyEntity
  }
})
