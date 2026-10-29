import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getRiskServiceBaseUrl } from '~/composables/useApiUrl'
import { extractErrorMessage } from '~/utils/error'

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

  const getRiskServiceBaseUrlLocal = () => {
    return getRiskServiceBaseUrl()
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
      console.error('Failed to fetch standard risk factors, using default seeds:', error)
      standardFactors.value = [
        { id: '1', name: 'Financial Materiality', description: 'Besaran potensi kerugian finansial atau keterpaparan aset', score_guidelines: '[{"score":1,"desc":"< 100 Juta"},{"score":3,"desc":"100 Juta - 1 Miliar"},{"score":5,"desc":"> 5 Miliar"}]', created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: '2', name: 'Strategic Importance', description: 'Tingkat signifikansi terhadap pencapaian target dan sasaran strategis perusahaan', score_guidelines: '[{"score":1,"desc":"Dampak minor"},{"score":3,"desc":"Dampak moderat"},{"score":5,"desc":"Dampak kritikal"}]', created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: '3', name: 'Inherent Risk', description: 'Tingkat risiko bawaan pada proses bisnis sebelum diterapkannya sistem pengendalian internal', score_guidelines: '[{"score":1,"desc":"Rendah"},{"score":3,"desc":"Sedang"},{"score":5,"desc":"Tinggi"}]', created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: '4', name: 'Internal Control Effectiveness', description: 'Efektivitas dan keandalan sistem kontrol internal yang sedang berjalan', score_guidelines: '[{"score":1,"desc":"Sangat Efektif"},{"score":3,"desc":"Cukup Efektif"},{"score":5,"desc":"Lemah / Tidak Ada"}]', created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: '5', name: 'Regulatory & Compliance Impact', description: 'Tingkat potensi sanksi hukum, denda regulator OJK/BI, atau pencabutan izin usaha', score_guidelines: '[{"score":1,"desc":"Tanpa sanksi"},{"score":3,"desc":"Teguran tertulis"},{"score":5,"desc":"Denda material & suspensi"}]', created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: '6', name: 'Information Security & Cyber Risk', description: 'Potensi kebocoran data sensitif nasabah atau serangan siber pada sistem informasi', score_guidelines: '[{"score":1,"desc":"Low risk"},{"score":3,"desc":"Medium risk"},{"score":5,"desc":"Critical data breach"}]', created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: '7', name: 'Operational Complexity', description: 'Tingkat kerumitan operasional transaksi, volume proses, dan ketergantungan sistem otomatisasi', score_guidelines: '[{"score":1,"desc":"Sederhana"},{"score":3,"desc":"Moderat"},{"score":5,"desc":"Kompleksitas tinggi"}]', created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: '8', name: 'Fraud Exposure', description: 'Kerentanan unit kerja terhadap tindak kecurangan internal maupun eksternal', score_guidelines: '[{"score":1,"desc":"Minimal"},{"score":3,"desc":"Peluang sedang"},{"score":5,"desc":"Tinggi / Temuan berulang"}]', created_at: '2026-01-01', updated_at: '2026-01-01' }
      ]
      errorMsg.value = ''
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
      console.error('Failed to fetch corporate risk factors, using default seeds:', error)
      corporateFactors.value = [
        { id: 'c1', standard_risk_factor_id: '1', weight: 0.25, standard_risk_factor: { id: '1', name: 'Financial Materiality', description: 'Besaran potensi kerugian finansial atau keterpaparan aset', score_guidelines: '', created_at: '2026-01-01', updated_at: '2026-01-01' }, created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: 'c2', standard_risk_factor_id: '2', weight: 0.20, standard_risk_factor: { id: '2', name: 'Strategic Importance', description: 'Tingkat signifikansi terhadap pencapaian target dan sasaran strategis perusahaan', score_guidelines: '', created_at: '2026-01-01', updated_at: '2026-01-01' }, created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: 'c3', standard_risk_factor_id: '5', weight: 0.20, standard_risk_factor: { id: '5', name: 'Regulatory & Compliance Impact', description: 'Tingkat potensi sanksi hukum, denda regulator OJK/BI, atau pencabutan izin usaha', score_guidelines: '', created_at: '2026-01-01', updated_at: '2026-01-01' }, created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: 'c4', standard_risk_factor_id: '6', weight: 0.20, standard_risk_factor: { id: '6', name: 'Information Security & Cyber Risk', description: 'Potensi kebocoran data sensitif nasabah atau serangan siber pada sistem informasi', score_guidelines: '', created_at: '2026-01-01', updated_at: '2026-01-01' }, created_at: '2026-01-01', updated_at: '2026-01-01' },
        { id: 'c5', standard_risk_factor_id: '8', weight: 0.15, standard_risk_factor: { id: '8', name: 'Fraud Exposure', description: 'Kerentanan unit kerja terhadap tindak kecurangan internal maupun eksternal', score_guidelines: '', created_at: '2026-01-01', updated_at: '2026-01-01' }, created_at: '2026-01-01', updated_at: '2026-01-01' }
      ]
      errorMsg.value = ''
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
      errorMsg.value = extractErrorMessage(error, 'Failed to save corporate risk factors.')
      return false
    } finally {
      loading.value = false
    }
  }

  const createStandardFactor = async (payload: { name: string; description?: string; score_guidelines?: string }) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risk-factors/standard`, {
        method: 'POST',
        body: payload
      })
      if (response && response.success) {
        await fetchStandardFactors()
        return response.data
      }
      return null
    } catch (error: any) {
      console.error('Failed to create standard risk factor:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to create standard risk factor.')
      return null
    } finally {
      loading.value = false
    }
  }

  const updateStandardFactor = async (id: string, payload: { name: string; description?: string; score_guidelines?: string }) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risk-factors/standard/${id}`, {
        method: 'PUT',
        body: payload
      })
      if (response && response.success) {
        await fetchStandardFactors()
        await fetchCorporateFactors()
        return response.data
      }
      return null
    } catch (error: any) {
      console.error('Failed to update standard risk factor:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to update standard risk factor.')
      return null
    } finally {
      loading.value = false
    }
  }

  const deleteStandardFactor = async (id: string) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risk-factors/standard/${id}`, {
        method: 'DELETE'
      })
      if (response && response.success) {
        await fetchStandardFactors()
        await fetchCorporateFactors()
        return true
      }
      return false
    } catch (error: any) {
      console.error('Failed to delete standard risk factor:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to delete standard risk factor.')
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
    saveCorporateFactors,
    createStandardFactor,
    updateStandardFactor,
    deleteStandardFactor
  }
})
