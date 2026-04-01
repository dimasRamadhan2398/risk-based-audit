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
    //const statusOptions: MitigationStatus[] = ['planned', 'in_progress', 'completed', 'overdue']

    // Form State
    const form = reactive<RiskMitigationForm>({
        actionPlan: '',
        supervisor: '',
        pic: '',
        start_date: '',
        end_date: '',
        //status: 'planned',
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
                actionPlan: data.actionPlan,
                supervisor: data.supervisor,
                pic: data.pic,
                start_date: data.start_date,
                end_date: data.end_date,
                //status: data.status,
                notes: data.notes || ''
            })
        } else {
            isEditing.value = false
            editingId.value = null
            Object.assign(form, {
                actionPlan: '',
                supervisor: '',
                pic: '',
                start_date: '',
                end_date: '',
                //status: 'Open',
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
        mitigations, isFormOpen, isEditing, form, picOptions, getMitigationsByRiskId,
        openForm, closeForm, handleSubmit, deleteMitigation,
    }
})