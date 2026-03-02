// stores/annual-plan.ts
import { defineStore } from 'pinia'
import { AnnualAuditPlanStatus, type AnnualAuditPlan, type AnnualPlanForm } from '~/types/audit'

export const useAnnualPlanStore = defineStore('annual-audit', () => {

  const supervisors = ref([
    { id: 'S01', name: 'Budi Santoso (Mgr)', workload: 5 },
    { id: 'S02', name: 'Siti Aminah (Snr)', workload: 8 }, // Overloaded
    { id: 'S03', name: 'John Doe (Mgr)', workload: 2 },
  ])

  // Mock Team Capacity (F-03 Logic)
  // Available = Total Auditor (misal 10) * 220 hari * 70%
  const TEAM_CAPACITY_PER_YEAR = 10 * 220 * 0.7 // 1540 Mandays
  // F-02: Auto-grouping Quarters
  const calculateQuarters = (months: number[]) => {
    const qSet = new Set<string>()
    months.forEach(m => {
      if (m >= 0 && m <= 2) qSet.add('Q1')
      else if (m >= 3 && m <= 5) qSet.add('Q2')
      else if (m >= 6 && m <= 8) qSet.add('Q3')
      else qSet.add('Q4')
    })
    return Array.from(qSet).sort()
  }

  // F-03: Utilization Check
  // Returns: { status: 'Green'|'Yellow'|'Red', percent: number, message: string }
  const checkUtilization = (additionalMandays: number) => {
    const currentUsed = plans.value.reduce((sum, p) => sum + p.totalMandays!, 0)
    const totalAfterAdd = currentUsed + additionalMandays
    const utilization = (totalAfterAdd / TEAM_CAPACITY_PER_YEAR) * 100

    if (utilization > 95) return { color: 'red', percent: utilization, msg: '🔴 OVERLOAD (>95%)' }
    if (utilization > 80) return { color: 'yellow', percent: utilization, msg: '🟡 High Load (80-95%)' }
    return { color: 'green', percent: utilization, msg: '🟢 Optimal (60-80%)' }
  }

  // F-04: Validate Schedule Gaps
  const checkScheduleGaps = (months: number[]) => {
    if (months.length === 0) return "Wajib pilih minimal 1 bulan."
    // Logic simple: Cek apakah bulan loncat (misal Jan & Mar dipilih, Feb kosong)
    const sorted = [...months].sort((a,b) => a-b)
    for (let i = 0; i < sorted.length - 1; i++) {
      if (sorted[i+1]! - sorted[i]! > 1) {
        return "⚠️ Warning: Ada gap bulan kosong. Pastikan continuous coverage jika diperlukan."
      }
    }
    return null
  }

  const plans = ref<AnnualAuditPlan[]>([])

  // Actions
  const addPlan = (form: AnnualPlanForm) => {
    // 1. Calculate Quarters (F-02)
    const quarters = calculateQuarters(form.selectedMonths)

    // 2. Calculate Resource (F-03)
    const totalMandays = form.auditorCount * form.daysPerAuditor
    const supervisor = supervisors.value.find(s => s.id === form.supervisorId)

    const newPlan: AnnualAuditPlan = {
      id: Date.now().toString(),
      code: form.code,
      activities: [...form.activities],
      // name: form.name, 
      // type: form.type, 
      // department: form.department, 
      status: form.status,
      selectedMonths: form.selectedMonths.sort((a,b) => a-b),
      quarters: quarters,
      auditorCount: form.auditorCount,
      daysPerAuditor: form.daysPerAuditor,
      totalMandays: totalMandays,
      supervisorId: form.supervisorId,
      supervisorName: supervisor?.name || 'Unknown',
      notes: form.notes,
      year: form.year,
      isActive: form.isActive    
    }
    plans.value.unshift(newPlan)
  }

  const updatePlan = (id: string, updatedData: AnnualPlanForm) => {
    const index = plans.value.findIndex(p => p.id === id)
    const isDuplicate = plans.value.some(a => a.code === updatedData.code && a.id !== id)
    if (isDuplicate) throw new Error('Kode Kegiatan sudah digunakan data lain!')

    if (index !== -1) {
      const quarters = calculateQuarters(updatedData.selectedMonths)
      const totalMandays = updatedData.auditorCount * updatedData.daysPerAuditor
      const supervisor = supervisors.value.find(s => s.id === updatedData.supervisorId)

      plans.value[index] = { 
        ...updatedData,
        activities: [...updatedData.activities],
        quarters: quarters,
        totalMandays: totalMandays,
        supervisorName: supervisor?.name || 'Unknown'
      }
    }
  }

  const deletePlan = (id: string) => {
    const plan = plans.value.find(a => a.id === id)
    if (!plan) return

    if (plan.isUsed) {
      alert('Data ini sudah digunakan dalam RKAT. Hanya status yang akan dinonaktifkan.')
      plan.isActive = false // Soft Delete / Deactivate
    } else {
      if (confirm('Apakah Anda yakin ingin menghapus permanen?')) {
        plans.value = plans.value.filter(a => a.id !== id)
      }
    }
  }

  return {
    plans,
    supervisors,
    addPlan,
    updatePlan,
    deletePlan,
    calculateQuarters,
    checkUtilization,
    checkScheduleGaps
  }
})