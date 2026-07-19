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

  const mockStatements: RiskAppetite[] = [
    {
      id: 'RA-001',
      statement: 'Toleransi terhadap pelanggaran kepatuhan hukum, regulasi eksternal, dan etika bisnis (Fraud / Korupsi)',
      threshold_limit: 0,
      status: 'APPROVED',
      created_at: '2026-01-05T08:00:00Z',
      updated_at: '2026-01-05T08:00:00Z'
    },
    {
      id: 'RA-002',
      statement: 'Toleransi deviasi anggaran biaya operasional tahunan departemen',
      threshold_limit: 5,
      status: 'APPROVED',
      created_at: '2026-01-05T08:00:00Z',
      updated_at: '2026-01-05T08:00:00Z'
    },
    {
      id: 'RA-003',
      statement: 'Maksimal rasio piutang tak tertagih (Bad Debt Ratio) terhadap total pendapatan',
      threshold_limit: 3,
      status: 'APPROVED',
      created_at: '2026-01-05T08:00:00Z',
      updated_at: '2026-01-05T08:00:00Z'
    },
    {
      id: 'RA-004',
      statement: 'Toleransi downtime sistem IT utama/Core System dalam sebulan',
      threshold_limit: 2,
      status: 'APPROVED',
      created_at: '2026-01-05T08:00:00Z',
      updated_at: '2026-01-05T08:00:00Z'
    }
  ]

  const fetchStatements = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risk-appetite`, { method: 'GET' })
      if (Array.isArray(response) && response.length > 0) {
        statements.value = response
      } else {
        statements.value = [...mockStatements]
      }
    } catch (error) {
      console.error('Failed to fetch risk appetite statements, falling back to mock:', error)
      errorMsg.value = 'Failed to load risk appetite statements.'
      statements.value = [...mockStatements]
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
