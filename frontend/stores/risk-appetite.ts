import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface RiskAppetite {
  id: string
  statement: string
  threshold_limit: number
  status: 'DRAFT' | 'SUBMITTED' | 'APPROVED'
  created_at?: string
  updated_at?: string
}

export const useRiskAppetiteStore = defineStore('risk-appetite', () => {
  const statements = ref<RiskAppetite[]>([])
  const loading = ref(false)
  const errorMsg = ref('')

  const getRiskServiceBaseUrl = () => {
    const config = useRuntimeConfig()
    return config.public.riskServiceBaseUrl || 'http://localhost:8004/api/v1'
  }

  const fetchStatements = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risk-appetite`, { method: 'GET' })
      if (Array.isArray(response)) {
        statements.value = response
      } else {
        statements.value = []
      }
    } catch (error) {
      console.error('Failed to fetch risk appetite statements:', error)
      errorMsg.value = 'Failed to load risk appetite statements.'
    } finally {
      loading.value = false
    }
  }

  const createStatement = async (payload: { statement: string; threshold_limit: number }) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risk-appetite`, {
        method: 'POST',
        body: {
          statement: payload.statement,
          threshold_limit: payload.threshold_limit,
          status: 'DRAFT'
        }
      })
      await fetchStatements()
      return response
    } catch (error: any) {
      console.error('Failed to create risk appetite statement:', error)
      errorMsg.value = error.data?.message || 'Failed to create statement.'
      throw error
    } finally {
      loading.value = false
    }
  }

  const updateStatement = async (id: string, payload: { statement: string; threshold_limit: number; status: string }) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risk-appetite/${id}`, {
        method: 'PUT',
        body: payload
      })
      await fetchStatements()
      return response
    } catch (error: any) {
      console.error('Failed to update risk appetite statement:', error)
      errorMsg.value = error.data?.message || 'Failed to update statement.'
      throw error
    } finally {
      loading.value = false
    }
  }

  const deleteStatement = async (id: string) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      await $fetch(`${baseUrl}/risk-appetite/${id}`, {
        method: 'DELETE'
      })
      await fetchStatements()
    } catch (error) {
      console.error('Failed to delete risk appetite statement:', error)
      errorMsg.value = 'Failed to delete statement.'
      throw error
    } finally {
      loading.value = false
    }
  }

  return {
    statements,
    loading,
    errorMsg,
    fetchStatements,
    createStatement,
    updateStatement,
    deleteStatement
  }
})
