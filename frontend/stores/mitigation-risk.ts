import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'
import type { RiskMitigation, RiskMitigationForm } from '~/types/risk'
import { useRiskProfileStore } from '~/stores/risk-profile'

export const useMitigationStore = defineStore('mitigation', () => {
    // --- STATE ---
    const mitigations = ref<RiskMitigation[]>([])
    const isFormOpen = ref(false)
    const isEditing = ref(false)
    const editingId = ref<string | null>(null)
    const loading = ref(false)
    const errorMsg = ref('')

    const getRiskServiceBaseUrl = () => {
        const config = useRuntimeConfig()
        return config.public.riskServiceBaseUrl || 'http://localhost:8004/api/v1'
    }

    // Data Master untuk Dropdown
    const picOptions = ['Dimas', 'Budi', 'Caca', 'Dedi', 'Eka', 'Fahmi']
    const supervisorOptions = ['Dimas', 'Budi', 'Caca', 'Dedi', 'Eka', 'Fahmi']
    const unitInChargeOptions = ['Sales', 'Marketing', 'Product Development', 'Operasional', 'Financial']

    // Form State
    const form = reactive<RiskMitigationForm>({
        riskEvent: '',
        mitigationPlan: '',
        supervisor: '',
        pic: '',
        unitInCharge: '',
        start_date: '',
        end_date: '',
        notes: ''
    })

    // --- GETTERS ---
    const getMitigationsByRiskId = computed(() => {
        return (risk_id: string) => mitigations.value.filter(m => m.riskId === risk_id)
    })

    // --- ACTIONS ---
    const fetchMitigations = async (riskId?: string) => {
        loading.value = true
        errorMsg.value = ''
        try {
            const baseUrl = getRiskServiceBaseUrl()
            const url = riskId ? `${baseUrl}/mitigations?riskId=${riskId}` : `${baseUrl}/mitigations`
            const response: any = await $fetch(url, { method: 'GET' })
            if (response && response.success && Array.isArray(response.data)) {
                mitigations.value = response.data
            } else if (Array.isArray(response)) {
                mitigations.value = response
            }
        } catch (error: any) {
            console.error('Failed to fetch mitigations:', error)
            errorMsg.value = 'Failed to load mitigations.'
        } finally {
            loading.value = false
        }
    }

    const openForm = (data?: RiskMitigation) => {
        if (data) {
            isEditing.value = true
            editingId.value = data.id
            Object.assign(form, {
                riskEvent: data.riskEvent,
                mitigationPlan: data.mitigationPlan,
                supervisor: data.supervisor,
                pic: data.pic,
                unitInCharge: data.unitInCharge,
                start_date: data.start_date ? data.start_date.split('T')[0] : '',
                end_date: data.end_date ? data.end_date.split('T')[0] : '',
                notes: data.notes || ''
            })
        } else {
            isEditing.value = false
            editingId.value = null
            Object.assign(form, {
                riskEvent: '',
                mitigationPlan: '',
                supervisor: '',
                pic: '',
                unitInCharge: '',
                start_date: '',
                end_date: '',
                notes: ''
            })
        }
        isFormOpen.value = true
    }

    const closeForm = () => {
        isFormOpen.value = false
    }

    const handleSubmit = async (currentRiskId: string) => {
        loading.value = true
        errorMsg.value = ''
        try {
            const riskProfileStore = useRiskProfileStore()
            const risk = riskProfileStore.getRiskById(currentRiskId)
            const riskName = risk ? risk.name : 'Risk Mitigation'
            const baseUrl = getRiskServiceBaseUrl()
            const payload = {
                ...form,
                riskEvent: form.riskEvent || riskName,
                riskId: currentRiskId,
                start_date: form.start_date ? new Date(form.start_date).toISOString() : undefined,
                end_date: form.end_date ? new Date(form.end_date).toISOString() : undefined
            }

            if (isEditing.value && editingId.value) {
                await $fetch(`${baseUrl}/mitigations/${editingId.value}`, {
                    method: 'PUT',
                    body: payload
                })
            } else {
                await $fetch(`${baseUrl}/mitigations`, {
                    method: 'POST',
                    body: payload
                })
            }
            closeForm()
            await fetchMitigations(currentRiskId)
        } catch (error: any) {
            console.error('Failed to save mitigation:', error)
            errorMsg.value = 'Failed to save mitigation.'
        } finally {
            loading.value = false
        }
    }

    const updateMitigationMonitoring = async (id: string, monitoring: any[], currentRiskId: string) => {
        loading.value = true
        errorMsg.value = ''
        try {
            const baseUrl = getRiskServiceBaseUrl()
            const existing = mitigations.value.find(m => m.id === id)
            if (!existing) return
            
            const payload = {
                ...existing,
                monitoring
            }
            
            await $fetch(`${baseUrl}/mitigations/${id}`, {
                method: 'PUT',
                body: payload
            })
            await fetchMitigations(currentRiskId)
        } catch (error: any) {
            console.error('Failed to update monitoring:', error)
            errorMsg.value = 'Failed to update monitoring.'
        } finally {
            loading.value = false
        }
    }

    const deleteMitigation = async (id: string, currentRiskId?: string) => {
        if (!confirm('Apakah Anda yakin ingin menghapus rencana mitigasi ini?')) return
        loading.value = true
        errorMsg.value = ''
        try {
            const baseUrl = getRiskServiceBaseUrl()
            await $fetch(`${baseUrl}/mitigations/${id}`, {
                method: 'DELETE'
            })
            await fetchMitigations(currentRiskId)
        } catch (error: any) {
            console.error('Failed to delete mitigation:', error)
            errorMsg.value = 'Failed to delete mitigation.'
        } finally {
            loading.value = false
        }
    }

    return {
        mitigations, isFormOpen, isEditing, form, getMitigationsByRiskId,
        supervisorOptions, unitInChargeOptions, picOptions,
        openForm, closeForm, handleSubmit, deleteMitigation, fetchMitigations,
        updateMitigationMonitoring,
        loading, errorMsg
    }
})