import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'
import type { RiskMitigation, RiskMitigationForm } from '~/types/risk'
import { useRiskProfileStore } from '~/stores/risk-profile'

export const initialMitigationsData: RiskMitigation[] = [
  {
    id: 'mit-1',
    riskId: '1',
    riskControlId: 'CTL-FIN-001',
    riskEvent: 'Target pendapatan dan laba tidak tercapai',
    mitigationPlan: 'Review bulanan pencapaian KPI sales dan monitoring piutang usaha secara ketat.',
    supervisor: 'Dimas',
    pic: 'Finance Manager',
    unitInCharge: 'Financial',
    start_date: '2026-01-01',
    end_date: '2026-12-31',
    notes: 'Monitoring piutang dijalankan setiap minggu.'
  },
  {
    id: 'mit-1b',
    riskId: '1',
    riskControlId: 'CTL-FIN-002',
    riskEvent: 'Target pendapatan dan laba tidak tercapai',
    mitigationPlan: 'Evaluasi ulang strategi penetapan harga dan diskon produk Kuartal II.',
    supervisor: 'Dimas',
    pic: 'Sales Director',
    unitInCharge: 'Sales',
    start_date: '2026-04-01',
    end_date: '2026-06-30',
    notes: 'Penyesuaian promo produk unggulan.'
  },
  {
    id: 'mit-2',
    riskId: '3',
    riskControlId: 'CTL-TEC-003',
    riskEvent: 'Ancaman terhadap Cyber Security dan perlindungan data pribadi',
    mitigationPlan: 'Implementasi Multi-Factor Authentication (MFA) & vulnerability scanning mingguan.',
    supervisor: 'Eka',
    pic: 'IT Security Lead',
    unitInCharge: 'Product Development',
    start_date: '2026-01-15',
    end_date: '2026-12-31',
    notes: 'Vulnerability scan otomatis tiap Jumat malam.'
  },
  {
    id: 'mit-3',
    riskId: '4',
    riskControlId: 'CTL-FIN-004',
    riskEvent: 'Terjadinya fraud',
    mitigationPlan: 'Dual approval pada sistem transaksi pembayaran di atas Rp 50 juta & audit mendadak.',
    supervisor: 'Budi',
    pic: 'Head of Internal Audit',
    unitInCharge: 'Financial',
    start_date: '2026-02-01',
    end_date: '2026-12-31',
    notes: 'SOP pembatasan wewenang keuangan.'
  }
]

// Helper to generate flat monitoring checks (Weekly or Monthly) based on timeline
function generateMonitoringChecks(startDateStr: string, endDateStr: string) {
    const checks: any[] = []
    if (!startDateStr || !endDateStr) return checks

    try {
        const start = new Date(startDateStr)
        const end = new Date(endDateStr)
        if (isNaN(start.getTime()) || isNaN(end.getTime())) return checks

        const diffTime = Math.abs(end.getTime() - start.getTime())
        const durationDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))

        if (durationDays < 60) {
            // Weekly
            let numWeeks = Math.floor((durationDays + 6) / 7)
            if (numWeeks <= 0) numWeeks = 1
            
            for (let i = 1; i <= numWeeks; i++) {
                const wStart = new Date(start)
                wStart.setDate(wStart.getDate() + (i - 1) * 7)
                
                const wEnd = new Date(start)
                wEnd.setDate(wEnd.getDate() + i * 7)
                wEnd.setSeconds(wEnd.getSeconds() - 1)
                
                const finalEnd = wEnd > end ? new Date(end) : wEnd
                
                const wStartFmt = wStart.toLocaleDateString('id-ID', { day: '2-digit', month: 'short' })
                const wEndFmt = finalEnd.toLocaleDateString('id-ID', { day: '2-digit', month: 'short' })

                checks.push({
                    id: `W${i}`,
                    label: `Week ${i} (${wStartFmt} - ${wEndFmt})`,
                    checked: false,
                    notes: "",
                    startDate: wStart.toISOString(),
                    endDate: finalEnd.toISOString()
                })
            }
        } else {
            // Monthly
            const monthNames = ["Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"]
            let current = new Date(start)
            let i = 1

            while (current <= end) {
                let mStart = new Date(current.getFullYear(), current.getMonth(), 1)
                if (mStart < start) mStart = new Date(start)

                let mEnd = new Date(current.getFullYear(), current.getMonth() + 1, 0, 23, 59, 59)
                if (mEnd > end) mEnd = new Date(end)

                checks.push({
                    id: `M${i}`,
                    label: `${monthNames[current.getMonth()]} ${current.getFullYear()}`,
                    checked: false,
                    notes: "",
                    startDate: mStart.toISOString(),
                    endDate: mEnd.toISOString()
                })

                current = new Date(current.getFullYear(), current.getMonth() + 1, 1)
                i++
            }
        }
    } catch (e) {
        console.error("Error generating monitoring checks:", e)
    }
    return checks
}

// Pre-fill initial data with monitoring checks
initialMitigationsData.forEach(mit => {
    mit.monitoring = generateMonitoringChecks(mit.start_date, mit.end_date)
})

export const useMitigationStore = defineStore('mitigation', () => {
    // --- STATE ---
    const mitigations = ref<RiskMitigation[]>(initialMitigationsData)
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
    const picOptions = ['Finance Manager', 'Sales Director', 'IT Security Lead', 'Head of Internal Audit', 'Dimas', 'Budi', 'Caca', 'Dedi', 'Eka', 'Fahmi']
    const supervisorOptions = ['Dimas', 'Budi', 'Caca', 'Dedi', 'Eka', 'Fahmi']
    const unitInChargeOptions = ['Sales', 'Marketing', 'Product Development', 'Operasional', 'Financial']

    // Form State
    const form = reactive<RiskMitigationForm>({
        riskControlId: '',
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
        return (risk_id: string) => {
            const list = mitigations.value.filter(m => String(m.riskId) === String(risk_id))
            if (list.length === 0) {
                return initialMitigationsData.filter(m => String(m.riskId) === String(risk_id))
            }
            return list
        }
    })

    // --- ACTIONS ---
    const fetchMitigations = async (riskId?: string) => {
        loading.value = true
        errorMsg.value = ''
        try {
            const baseUrl = getRiskServiceBaseUrl()
            const url = riskId ? `${baseUrl}/mitigations?riskId=${riskId}` : `${baseUrl}/mitigations`
            const response: any = await $fetch(url, { method: 'GET' })
            const responseData = (response && response.success && Array.isArray(response.data)) ? response.data : (Array.isArray(response) ? response : [])
            
            if (responseData.length > 0) {
                const riskProfileStore = useRiskProfileStore()
                mitigations.value = responseData.map((m: any, idx: number) => {
                    const risk = riskProfileStore.getRiskById(m.riskId)
                    const formattedRiskCode = risk ? riskProfileStore.getFormattedId(risk) : `RSK-${String(idx + 1).padStart(3, '0')}`
                    const isLongUuid = m.riskControlId && (m.riskControlId.includes('-4') || m.riskControlId.length > 25)
                    const cleanControlId = (!m.riskControlId || isLongUuid) ? `CTL-${formattedRiskCode}` : m.riskControlId
                    
                    let safeMonitoring = m.monitoring
                    if (!safeMonitoring || safeMonitoring.length === 0) {
                        const localMit = mitigations.value.find(local => local.id === m.id)
                        if (localMit && localMit.monitoring && localMit.monitoring.length > 0) {
                            safeMonitoring = localMit.monitoring
                        } else {
                            safeMonitoring = generateMonitoringChecks(m.start_date, m.end_date)
                        }
                    }

                    return {
                        ...m,
                        riskControlId: cleanControlId,
                        monitoring: safeMonitoring
                    }
                })
            }
        } catch (error: any) {
            console.warn('Backend mitigations unavailable, using initial data.', error)
        } finally {
            loading.value = false
        }
    }

    const openForm = (data?: RiskMitigation) => {
        if (data) {
            isEditing.value = true
            editingId.value = data.id
            Object.assign(form, {
                riskControlId: data.riskControlId || '',
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
                riskControlId: '',
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
            const riskCode = risk ? riskProfileStore.getFormattedId(risk) : 'RSK-001'
            
            const generatedControlId = form.riskControlId || `CTL-${riskCode}-${mitigations.value.length + 1}`

            const baseUrl = getRiskServiceBaseUrl()
            const payload = {
                ...form,
                riskControlId: generatedControlId,
                riskEvent: form.riskEvent || riskName,
                riskId: currentRiskId,
                start_date: form.start_date ? new Date(form.start_date).toISOString() : '',
                end_date: form.end_date ? new Date(form.end_date).toISOString() : ''
            }

            const newMonitoring = generateMonitoringChecks(payload.start_date, payload.end_date)

            if (isEditing.value && editingId.value) {
                const idx = mitigations.value.findIndex(m => m.id === editingId.value)
                let payloadWithMonitoring = { ...payload }
                if (idx !== -1) {
                    const oldMit = mitigations.value[idx]
                    const dateChanged = oldMit.start_date !== payload.start_date || oldMit.end_date !== payload.end_date
                    const updatedMonitoring = dateChanged ? newMonitoring : (oldMit.monitoring || [])
                    payloadWithMonitoring = { ...payload, monitoring: updatedMonitoring }
                    mitigations.value[idx] = { ...oldMit, ...payloadWithMonitoring, id: editingId.value }
                }
                try {
                    await $fetch(`${baseUrl}/mitigations/${editingId.value}`, {
                        method: 'PUT',
                        body: payloadWithMonitoring
                    })
                    await fetchMitigations(currentRiskId)
                } catch (e) {
                    console.warn('Backend update failed, saved locally.')
                    mitigations.value = [...mitigations.value]
                }
            } else {
                const newMitigation: RiskMitigation = {
                    id: `mit-${Date.now()}`,
                    ...payload,
                    start_date: payload.start_date || new Date().toISOString(),
                    end_date: payload.end_date || new Date().toISOString(),
                    monitoring: newMonitoring
                }
                mitigations.value.push(newMitigation)
                try {
                    await $fetch(`${baseUrl}/mitigations`, {
                        method: 'POST',
                        body: { ...payload, monitoring: newMonitoring }
                    })
                    await fetchMitigations(currentRiskId)
                } catch (e) {
                    console.warn('Backend create failed, saved locally.')
                    mitigations.value = [...mitigations.value]
                }
            }
            closeForm()
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