import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface StandardRiskFactor {
  id: string
  name: string
  description: string
  score_guidelines: string // JSON string
  created_at: string
  updated_at: string
}

export interface CorporateRiskFactor {
  id: string
  standard_risk_factor_id: string
  standard_risk_factor?: StandardRiskFactor
  weight: number // fraction in DB e.g. 0.15
  created_at: string
  updated_at: string
}

export const useRiskFactorsStore = defineStore('risk-factors', () => {
  const config = useRuntimeConfig()
  
  const standardFactors = ref<StandardRiskFactor[]>([])
  const corporateFactors = ref<CorporateRiskFactor[]>([])
  const loading = ref(false)
  const errorMsg = ref('')

  const getRiskServiceBaseUrl = () => {
    return config.public.riskServiceBaseUrl || 'http://localhost:8080/api/v1'
  }

  const fetchStandardFactors = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risk-factors/standard`)
      if (response && response.success) {
        standardFactors.value = response.data
      }
    } catch (error: any) {
      console.error('Failed to fetch standard risk factors:', error)
      errorMsg.value = error.data?.error || 'Failed to fetch standard risk factors.'
    } finally {
      loading.value = false
    }
  }

  const fetchCorporateFactors = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risk-factors/corporate`)
      if (response && response.success) {
        corporateFactors.value = response.data
      }
    } catch (error: any) {
      console.error('Failed to fetch corporate risk factors:', error)
      errorMsg.value = error.data?.error || 'Failed to fetch corporate risk factors.'
    } finally {
      loading.value = false
    }
  }

  const saveCorporateFactors = async (payload: { standard_risk_factor_id: string; weight: number }[]) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risk-factors/corporate`, {
        method: 'POST',
        body: payload
      })
      if (response && response.success) {
        corporateFactors.value = response.data
        return true
      }
      return false
    } catch (error: any) {
      console.error('Failed to save corporate risk factors:', error)
      errorMsg.value = error.data?.error || 'Failed to save corporate risk factors.'
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    standardFactors,
    corporateFactors,
    loading,
    errorMsg,
    fetchStandardFactors,
    fetchCorporateFactors,
    saveCorporateFactors
  }
})
