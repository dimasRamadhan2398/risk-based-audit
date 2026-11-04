import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'
import type { RiskMitigation, RiskMitigationForm } from '~/types/risk'
import type { Employee, Department } from '~/types/master'
import { useEmployeeApi } from '~/composables/useEmployeeApi'
import { useDepartmentApi } from '~/composables/useDepartmentApi'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { extractErrorMessage } from '~/utils/error'
import { getRiskServiceBaseUrl } from '~/composables/useApiUrl'
import { useI18n } from '~/composables/useI18n'

export const initialMitigationsData: RiskMitigation[] = []

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

export const useMitigationStore = defineStore('mitigation', () => {
    const { t } = useI18n()
    // --- STATE ---
    const mitigations = ref<RiskMitigation[]>([])
    const isFormOpen = ref(false)
    const isEditing = ref(false)
    const editingId = ref<string | null>(null)
    const loading = ref(false)
    const errorMsg = ref('')



    // Data Master untuk Dropdown
    // Master Data Karyawan (Employee Management Sync)
    const employees = ref<Employee[]>([])
    const employeesLoading = ref(false)
    const employeesLoaded = ref(false)
    const employeesError = ref('')

    const resolveEmployeeOption = (val?: string): string => {
        if (!val) return ''
        const match = employees.value.find(
            e => e.id === val || e.employee_code === val || e.full_name === val || `${e.full_name} (${e.employee_code})` === val
        )
        if (match) {
            return match.employee_code ? `${match.full_name} (${match.employee_code})` : match.full_name
        }
        return val
    }

    const formatEmployeeName = (value?: string): string => {
        if (!value) return '-'
        const found = employees.value.find(
            e => e.id === value || e.employee_code === value
        )
        if (found) {
            return found.employee_code ? `${found.full_name} (${found.employee_code})` : found.full_name
        }
        return value
    }

    const fetchEmployees = async (force = false) => {
        if (employeesLoaded.value && !force) return
        if (employeesLoading.value) return
        employeesLoading.value = true
        employeesError.value = ''
        try {
            const api = useEmployeeApi()
            const list = await api.getAllEmployees()
            employees.value = (Array.isArray(list) ? list : [])
                .filter(e => e?.full_name && e.is_active !== false)
                .sort((a, b) => a.full_name.localeCompare(b.full_name, 'id', { sensitivity: 'base' }))
            employeesLoaded.value = true

            if (form.supervisor) {
                form.supervisor = resolveEmployeeOption(form.supervisor)
            }
            if (form.pic) {
                form.pic = resolveEmployeeOption(form.pic)
            }
        } catch (error: any) {
            console.error('Failed to fetch employees for mitigation:', error)
            employeesError.value = extractErrorMessage(error, 'Failed to fetch employees.')
            if (!employeesLoaded.value) {
                employees.value = []
            }
        } finally {
            employeesLoading.value = false
        }
    }

    const picOptions = computed(() => {
        const list = employees.value.map(e => e.employee_code ? `${e.full_name} (${e.employee_code})` : e.full_name)
        if (form.pic && !list.includes(form.pic)) {
            list.unshift(form.pic)
        }
        return list.sort((a, b) => a.localeCompare(b, 'id', { sensitivity: 'base' }))
    })

    const supervisorOptions = computed(() => {
        const list = employees.value.map(e => e.employee_code ? `${e.full_name} (${e.employee_code})` : e.full_name)
        if (form.supervisor && !list.includes(form.supervisor)) {
            list.unshift(form.supervisor)
        }
        return list.sort((a, b) => a.localeCompare(b, 'id', { sensitivity: 'base' }))
    })

    // Master Data Departemen (Department Management Sync)
    const departments = ref<Department[]>([])
    const departmentsLoading = ref(false)
    const departmentsLoaded = ref(false)
    const departmentsError = ref('')

    const resolveDepartmentOption = (val?: string): string => {
        if (!val) return ''
        const match = departments.value.find(
            d => d.id === val || d.department_code === val || d.department_name === val || `${d.department_name} (${d.department_code})` === val
        )
        if (match) {
            return match.department_code ? `${match.department_name} (${match.department_code})` : match.department_name
        }
        return val
    }

    const formatDepartmentName = (value?: string): string => {
        if (!value) return '-'
        const found = departments.value.find(
            d => d.id === value || d.department_code === value || d.department_name === value
        )
        if (found) {
            return found.department_code ? `${found.department_name} (${found.department_code})` : found.department_name
        }
        return value
    }

    const fetchDepartments = async (force = false) => {
        if (departmentsLoaded.value && !force) return
        if (departmentsLoading.value) return
        departmentsLoading.value = true
        departmentsError.value = ''
        try {
            const api = useDepartmentApi()
            const list = await api.getAllDepartments()
            departments.value = (Array.isArray(list) ? list : [])
                .filter(d => d?.department_name && d.is_active !== false)
                .sort((a, b) => a.department_name.localeCompare(b.department_name, 'id', { sensitivity: 'base' }))
            departmentsLoaded.value = true

            if (form.unitInCharge) {
                form.unitInCharge = resolveDepartmentOption(form.unitInCharge)
            }
        } catch (error: any) {
            console.error('Failed to fetch departments for mitigation:', error)
            departmentsError.value = extractErrorMessage(error, 'Failed to fetch departments.')
            if (!departmentsLoaded.value) {
                departments.value = []
            }
        } finally {
            departmentsLoading.value = false
        }
    }

    const unitInChargeOptions = computed(() => {
        const list = departments.value.map(d => d.department_code ? `${d.department_name} (${d.department_code})` : d.department_name)
        if (form.unitInCharge && !list.includes(form.unitInCharge)) {
            list.unshift(form.unitInCharge)
        }
        return list.sort((a, b) => a.localeCompare(b, 'id', { sensitivity: 'base' }))
    })


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
        return (risk_id?: string) => {
            if (!risk_id) return mitigations.value
            const target = String(risk_id).trim().toLowerCase()
            return mitigations.value.filter(m => {
                if (!m.riskId) return true
                return String(m.riskId).trim().toLowerCase() === target
            })
        }
    })

    // --- ACTIONS ---
    const fetchMitigations = async (riskId?: string) => {
        if (!employeesLoaded.value && !employeesLoading.value) {
            fetchEmployees().catch(() => {})
        }
        if (!departmentsLoaded.value && !departmentsLoading.value) {
            fetchDepartments().catch(() => {})
        }
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
            } else {
                mitigations.value = []
            }
        } catch (error: any) {
            console.error('Failed to fetch mitigations:', error)
            errorMsg.value = extractErrorMessage(error, 'Failed to fetch mitigations.')
        } finally {
            loading.value = false
        }
    }

    const openForm = (data?: RiskMitigation) => {
        if (!employeesLoaded.value) {
            fetchEmployees().catch(() => {})
        }
        if (!departmentsLoaded.value) {
            fetchDepartments().catch(() => {})
        }
        if (data) {
            isEditing.value = true
            editingId.value = data.id
            Object.assign(form, {
                riskControlId: data.riskControlId || '',
                riskEvent: data.riskEvent,
                mitigationPlan: data.mitigationPlan,
                supervisor: resolveEmployeeOption(data.supervisor),
                pic: resolveEmployeeOption(data.pic),
                unitInCharge: resolveDepartmentOption(data.unitInCharge),
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
                if (idx !== -1 && mitigations.value[idx]) {
                    const oldMit = mitigations.value[idx]!
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
                } catch (e: any) {
                    console.warn('Backend update failed, saved locally.')
                    errorMsg.value = extractErrorMessage(e, 'Backend update failed, saved locally.')
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
                } catch (e: any) {
                    console.warn('Backend create failed, saved locally.')
                    errorMsg.value = extractErrorMessage(e, 'Backend create failed, saved locally.')
                    mitigations.value = [...mitigations.value]
                }
            }
            closeForm()
        } catch (error: any) {
            console.error('Failed to save mitigation:', error)
            errorMsg.value = extractErrorMessage(error, 'Failed to save mitigation.')
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
            errorMsg.value = extractErrorMessage(error, 'Failed to update monitoring.')
        } finally {
            loading.value = false
        }
    }

    const deleteMitigation = async (id: string, currentRiskId?: string) => {
        if (!await useGlobalModalStore().confirmDelete({ description: t('riskMitigation.deleteConfirm') })) return
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
            errorMsg.value = extractErrorMessage(error, 'Failed to delete mitigation.')
        } finally {
            loading.value = false
        }
    }

    return {
        mitigations, isFormOpen, isEditing, form, getMitigationsByRiskId,
        employees, employeesLoading, employeesLoaded, employeesError, fetchEmployees, formatEmployeeName,
        departments, departmentsLoading, departmentsLoaded, departmentsError, fetchDepartments, formatDepartmentName,
        supervisorOptions, unitInChargeOptions, picOptions,
        openForm, closeForm, handleSubmit, deleteMitigation, fetchMitigations,
        updateMitigationMonitoring,
        loading, errorMsg
    }
})
