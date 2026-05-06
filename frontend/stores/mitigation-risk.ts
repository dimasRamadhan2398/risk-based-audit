// stores/mitigation.ts
import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import type { RiskMitigation, RiskMitigationForm } from '~/types/risk'

export const useMitigationStore = defineStore('mitigation', () => {
    // --- STATE ---
    const mitigations = ref<RiskMitigation[]>([])
    const isFormOpen = ref(false)
    const isEditing = ref(false)
    const editingId = ref<string | null>(null)

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
    // Menambahkan fungsi untuk memfilter mitigasi berdasarkan ID risiko yang sedang dibuka
    const getMitigationsByRiskId = computed(() => {
        return (risk_id: string) => mitigations.value.filter(m => m.riskId === risk_id)
    })

    // --- ACTIONS ---
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
                start_date: data.start_date,
                end_date: data.end_date,
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

    const handleSubmit = (currentRiskId: string) => {
        if (isEditing.value && editingId.value) {
            // Update data
            const index = mitigations.value.findIndex(m => m.id === editingId.value)
            if (index !== -1) {
                mitigations.value[index] = { ...form, id: editingId.value, riskId: currentRiskId }
            }
        } else {
            // Tambah data baru (Mock ID)
            mitigations.value.push({
                ...form,
                id: `MIT-${Date.now()}`,
                riskId: currentRiskId
            })
        }
        closeForm()
    }

    const deleteMitigation = (id: string) => {
        if (confirm('Apakah Anda yakin ingin menghapus rencana mitigasi ini?')) {
            mitigations.value = mitigations.value.filter(m => m.id !== id)
        }
    }

    return {
        mitigations, isFormOpen, isEditing, form, getMitigationsByRiskId,
        supervisorOptions, unitInChargeOptions, picOptions,
        openForm, closeForm, handleSubmit, deleteMitigation,
    }
})