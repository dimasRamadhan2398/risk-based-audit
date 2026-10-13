import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import type { TableColumn } from '@nuxt/ui'
import { extractErrorMessage } from '~/utils/error'
import { getRiskServiceBaseUrl } from '~/composables/useApiUrl'

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
  const toast = useToastNotification()

  const columns: (TableColumn<any> & { class?: string })[] = [
    {
      accessorKey: 'id',
      id: 'id',
      header: 'ID Risiko',
      class: 'w-[120px] min-w-[100px] whitespace-nowrap'
    },
    {
      accessorKey: 'name',
      id: 'name',
      header: 'Nama Risiko & Area',
      class: 'min-w-[260px] max-w-[380px]'
    },
    {
      accessorKey: 'level',
      id: 'level',
      header: 'Level Risiko',
      class: 'w-[130px] min-w-[130px] text-center whitespace-nowrap'
    },
    {
      accessorKey: 'appetite',
      id: 'appetite',
      header: 'Toleransi Appetite',
      class: 'w-[150px] min-w-[150px] text-center whitespace-nowrap'
    },
    {
      accessorKey: 'mitigationStatus',
      id: 'mitigationStatus',
      header: 'Status Mitigasi',
      class: 'w-[180px] min-w-[180px] whitespace-nowrap'
    },
    {
      accessorKey: 'actions',
      id: 'actions',
      header: 'Aksi',
      class: 'w-[140px] min-w-[140px] text-center whitespace-nowrap'
    }
  ]




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
    } catch (error: any) {
      console.error('Failed to fetch risk appetite statements:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to load risk appetite statements.')
      statements.value = []
    } finally {
      loading.value = false
    }
  }

  const createStatement = async (payload: { statement: string; threshold_limit: number; status?: 'DRAFT' | 'SUBMITTED' | 'APPROVED' | string }) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const statusValue = typeof payload.status === 'object' && payload.status !== null
        ? (payload.status as any).value
        : (payload.status || 'DRAFT')

      const response: any = await $fetch(`${baseUrl}/risk-appetite`, {
        method: 'POST',
        body: {
          statement: payload.statement,
          threshold_limit: payload.threshold_limit,
          status: statusValue
        }
      })
      toast.showSuccess('Statement berhasil dibuat')
      await fetchStatements()
      return response
    } catch (error: any) {
      console.error('Failed to create risk appetite statement:', error)
      const message = extractErrorMessage(error, 'Gagal membuat statement.')
      errorMsg.value = message
      toast.showError('Gagal membuat statement', message)
      throw error
    } finally {
      loading.value = false
    }
  }

  const updateStatement = async (id: string, payload: { statement: string; threshold_limit: number; status?: 'DRAFT' | 'SUBMITTED' | 'APPROVED' | string }) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const statusValue = typeof payload.status === 'object' && payload.status !== null
        ? (payload.status as any).value
        : (payload.status || 'DRAFT')

      const response: any = await $fetch(`${baseUrl}/risk-appetite/${id}`, {
        method: 'PUT',
        body: {
          statement: payload.statement,
          threshold_limit: payload.threshold_limit,
          status: statusValue
        }
      })
      toast.showSuccess('Statement berhasil diupdate')
      await fetchStatements()
      return response
    } catch (error: any) {
      console.error('Failed to update risk appetite statement:', error)
      const message = extractErrorMessage(error, 'Gagal mengupdate statement.')
      errorMsg.value = message
      toast.showError('Gagal mengupdate statement', message)
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
      toast.showSuccess('Statement berhasil dihapus')
      await fetchStatements()
    } catch (error: any) {
      console.error('Failed to delete risk appetite statement:', error)
      const message = extractErrorMessage(error, 'Gagal menghapus statement.')
      errorMsg.value = message
      toast.showError('Gagal menghapus statement', message)
      throw error
    } finally {
      loading.value = false
    }
  }

  return {
    statements,
    columns,
    complianceColumns: columns,
    loading,
    errorMsg,
    fetchStatements,
    createStatement,
    updateStatement,
    deleteStatement
  }
})
