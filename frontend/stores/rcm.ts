import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import type { TableColumn } from '@nuxt/ui'
import { useRiskProfileStore, riskLevelConfig } from '~/stores/risk-profile'
import { useMitigationStore } from '~/stores/mitigation-risk'
import { RiskLevel } from '~/types/risk'
import { extractErrorMessage } from '~/utils/error'
import { getRiskServiceBaseUrl } from '~/composables/useApiUrl'
import { useFiscalYear } from '~/composables/useFiscalYear'

export interface RCMItem {
  id: string
  risk_id?: string
  risk_code: string
  risk_event: string
  control_code: string
  control_description: string
  control_type?: string
  control_owner: string
  department: string
  year: number
  design_effectiveness_weight: number // 20%
  design_effectiveness_rating: number // 1-5 (4%, 8%, 12%, 16%, 20%)
  operating_effectiveness_weight: number
  operating_effectiveness_rating: number
  coverage_completeness_weight: number
  coverage_completeness_rating: number
  timeliness_weight: number
  timeliness_rating: number
  automation_monitoring_weight: number
  automation_monitoring_rating: number
  total_weighted_score: number // 20% - 100%
  inherent_risk?: number
  residual_risk?: number
  notes?: string
}

export const cosoDimensions = [
  { key: 'design_effectiveness', label: '1. Design Effectiveness', shortLabel: 'Design', ratingKey: 'design_effectiveness_rating' },
  { key: 'operating_effectiveness', label: '2. Operating Effectiveness', shortLabel: 'Operating', ratingKey: 'operating_effectiveness_rating' },
  { key: 'coverage_completeness', label: '3. Coverage & Completeness', shortLabel: 'Coverage', ratingKey: 'coverage_completeness_rating' },
  { key: 'timeliness', label: '4. Timeliness', shortLabel: 'Timeliness', ratingKey: 'timeliness_rating' },
  { key: 'automation_monitoring', label: '5. Automation & Monitoring', shortLabel: 'Automation', ratingKey: 'automation_monitoring_rating' }
]

export const initialRCMData: RCMItem[] = [
  {
    id: 'rcm-1',
    risk_id: '1',
    risk_code: 'FIN-001',
    risk_event: 'Target pendapatan dan laba tidak tercapai',
    control_code: 'CTL-FIN-001',
    control_description: 'Review bulanan pencapaian KPI sales dan monitoring piutang usaha secara ketat.',
    control_owner: 'Finance Manager',
    department: 'Head Office',
    year: 2026,
    design_effectiveness_weight: 20,
    design_effectiveness_rating: 4, // 16%
    operating_effectiveness_weight: 20,
    operating_effectiveness_rating: 3, // 12%
    coverage_completeness_weight: 20,
    coverage_completeness_rating: 4, // 16%
    timeliness_weight: 20,
    timeliness_rating: 3, // 12%
    automation_monitoring_weight: 20,
    automation_monitoring_rating: 2, // 8%
    total_weighted_score: 64, // 64% -> Weak
    notes: 'Internal control cukup efektif namun automasi monitoring perlu ditingkatkan.'
  },
  {
    id: 'rcm-2',
    risk_id: '3',
    risk_code: 'TEC-003',
    risk_event: 'Ancaman terhadap Cyber Security dan perlindungan data pribadi',
    control_code: 'CTL-TEC-003',
    control_description: 'Implementasi Multi-Factor Authentication (MFA) & vulnerability scanning mingguan.',
    control_owner: 'IT Security Lead',
    department: 'Head Office',
    year: 2026,
    design_effectiveness_weight: 20,
    design_effectiveness_rating: 5, // 20%
    operating_effectiveness_weight: 20,
    operating_effectiveness_rating: 4, // 16%
    coverage_completeness_weight: 20,
    coverage_completeness_rating: 4, // 16%
    timeliness_weight: 20,
    timeliness_rating: 5, // 20%
    automation_monitoring_weight: 20,
    automation_monitoring_rating: 4, // 16%
    total_weighted_score: 88, // 88% -> Effective
    notes: 'Kontrol keamanan berjalan secara otomatis dan rutin dievaluasi.'
  },
  {
    id: 'rcm-3',
    risk_id: '4',
    risk_code: 'FIN-004',
    risk_event: 'Terjadinya fraud',
    control_code: 'CTL-FIN-004',
    control_description: 'Dual approval pada sistem transaksi pembayaran di atas Rp 50 juta & audit mendadak.',
    control_owner: 'Head of Internal Audit',
    department: 'Head Office',
    year: 2026,
    design_effectiveness_weight: 20,
    design_effectiveness_rating: 4, // 16%
    operating_effectiveness_weight: 20,
    operating_effectiveness_rating: 4, // 16%
    coverage_completeness_weight: 20,
    coverage_completeness_rating: 4, // 16%
    timeliness_weight: 20,
    timeliness_rating: 4, // 16%
    automation_monitoring_weight: 20,
    automation_monitoring_rating: 3, // 12%
    total_weighted_score: 76, // 76% -> Moderately Effective
    notes: 'SOP otorisasi berjalan, perlu tindak lanjut log audit elektronik.'
  }
]

export interface DepartmentRiskExposure {
  name: string
  inherentRisk: number
  residualRisk: number
  riskCount: number
}

const roundScore = (val: number) => Math.round(val * 10) / 10

/**
 * Inherent vs Residual risk exposure per department, derived from the Corporate
 * Risk Profile.
 *
 * Inherent is the start-of-year (Q1) assessment and residual the end-of-year
 * (Q4) one — the same definitions the RCM summary uses for its risk counts.
 * Both are `impact × likelihood` averaged over the department's risks, and a
 * risk with no assessment for the year contributes its base score to both
 * series (nothing mitigated yet). Takes raw CRP risks, so the result does not
 * depend on which period the Risk Profile page happens to be showing.
 */
export const computeInherentVsResidualByDepartment = (
  rawRisks: any[] = [],
  year: number
): DepartmentRiskExposure[] => {
  const currentYear = new Date().getFullYear()
  const isFutureYear = year > currentYear
  const groups = new Map<string, { inherent: number, residual: number, count: number }>()

  ;(rawRisks || []).forEach((risk: any) => {
    const department = risk?.branch || risk?.department
    if (!department) return

    const assessment = risk.assessments?.find((a: any) => a.year === year)
    if (isFutureYear && !assessment) {
      // Future year risks must be inputted by user; exclude risks without an assessment for that year
      return
    }

    const baseImpact = risk.impact ?? 0
    const baseLikelihood = risk.likelihood ?? 0

    const inherent =
      (assessment?.impact_q1 ?? baseImpact) * (assessment?.likelihood_q1 ?? baseLikelihood)
    const residual =
      (assessment?.impact_q4 ?? baseImpact) * (assessment?.likelihood_q4 ?? baseLikelihood)

    const group = groups.get(department) || { inherent: 0, residual: 0, count: 0 }
    group.inherent += inherent
    group.residual += residual
    group.count += 1
    groups.set(department, group)
  })

  return Array.from(groups.entries()).map(([name, g]) => ({
    name,
    inherentRisk: roundScore(g.inherent / g.count),
    residualRisk: roundScore(g.residual / g.count),
    riskCount: g.count
  }))
}

export const getEffectivenessInterpretation = (scorePercent: number) => {
  if (scorePercent >= 90) {
    return {
      score: scorePercent,
      rating: 'Highly Effective',
      interpretation: 'Controls reliably mitigate risk and require only routine monitoring.',
      color: 'success',
      alertColor: 'success',
      bgClass: 'bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800',
      badgeColor: 'emerald'
    }
  } else if (scorePercent >= 80) {
    return {
      score: scorePercent,
      rating: 'Effective',
      interpretation: 'Controls function well; only minor improvements are recommended.',
      color: 'info',
      alertColor: 'info',
      bgClass: 'bg-sky-50 dark:bg-sky-950/50 text-sky-700 dark:text-sky-300 border border-sky-200 dark:border-sky-800',
      badgeColor: 'sky'
    }
  } else if (scorePercent >= 70) {
    return {
      score: scorePercent,
      rating: 'Moderately Effective',
      interpretation: 'Some weaknesses exist; corrective actions should be planned.',
      color: 'warning',
      alertColor: 'warning',
      bgClass: 'bg-amber-50 dark:bg-amber-950/50 text-amber-700 dark:text-amber-300 border border-amber-200 dark:border-amber-800',
      badgeColor: 'amber'
    }
  } else if (scorePercent >= 60) {
    return {
      score: scorePercent,
      rating: 'Weak',
      interpretation: 'Significant improvements are needed to reduce risk adequately.',
      color: 'warning',
      alertColor: 'warning',
      bgClass: 'bg-orange-50 dark:bg-orange-950/50 text-orange-700 dark:text-orange-300 border border-orange-200 dark:border-orange-800',
      badgeColor: 'orange'
    }
  } else {
    return {
      score: scorePercent,
      rating: 'Ineffective',
      interpretation: 'Controls do not provide sufficient risk mitigation and require immediate attention.',
      color: 'error',
      alertColor: 'error',
      bgClass: 'bg-red-50 dark:bg-red-950/50 text-red-700 dark:text-red-300 border border-red-200 dark:border-red-800',
      badgeColor: 'red'
    }
  }
}

export const useRCMStore = defineStore('rcm', () => {
  const config = useRuntimeConfig()
  const riskProfileStore = useRiskProfileStore()
  const rcmList = ref<RCMItem[]>([])
  const { selectedFiscalYear } = useFiscalYear()
  const selectedYear = selectedFiscalYear
  const selectedDepartment = ref('All Departments')
  const loading = ref(false)
  const errorMsg = ref('')

  const columns: (TableColumn<RCMItem> & { class?: string })[] = [
    { accessorKey: 'risk_code', id: 'risk_code', header: 'Kode / Risiko', class: 'w-[280px] min-w-[280px]' },
    { accessorKey: 'control_code', id: 'control_code', header: 'Risk Control ID & Deskripsi Mitigasi', class: 'w-[320px] min-w-[320px]' },
    { accessorKey: 'department', id: 'department', header: 'Departemen / PIC', class: 'w-[170px] min-w-[170px]' },
    { accessorKey: 'design_effectiveness_rating', id: 'design_effectiveness_rating', header: 'Design', class: 'w-[100px] min-w-[100px] text-center' },
    { accessorKey: 'operating_effectiveness_rating', id: 'operating_effectiveness_rating', header: 'Operating', class: 'w-[100px] min-w-[100px] text-center' },
    { accessorKey: 'coverage_completeness_rating', id: 'coverage_completeness_rating', header: 'Coverage', class: 'w-[100px] min-w-[100px] text-center' },
    { accessorKey: 'timeliness_rating', id: 'timeliness_rating', header: 'Timeliness', class: 'w-[100px] min-w-[100px] text-center' },
    { accessorKey: 'automation_monitoring_rating', id: 'automation_monitoring_rating', header: 'Automation', class: 'w-[110px] min-w-[110px] text-center' },
    { accessorKey: 'total_weighted_score', id: 'total_weighted_score', header: 'Total Score', class: 'w-[110px] min-w-[110px] text-center' },
    { accessorKey: 'rating', id: 'rating', header: 'Rating Efektivitas', class: 'w-[150px] min-w-[150px] text-center' },
    { accessorKey: 'actions', id: 'actions', header: 'Aksi', class: 'w-[90px] min-w-[90px] text-center' }
  ]

  const LOCAL_STORAGE_KEY = 'rcm_items_v2'

  // Initialize from LocalStorage or Fallback
  const initStoreData = () => {
    if (import.meta.client) {
      const saved = localStorage.getItem(LOCAL_STORAGE_KEY)
      if (saved) {
        try {
          const parsed = JSON.parse(saved)
          if (Array.isArray(parsed) && parsed.length > 0) {
            rcmList.value = parsed
            return
          }
        } catch (e) {
          console.error('Failed to parse saved RCM items:', e)
        }
      }
    }
    rcmList.value = JSON.parse(JSON.stringify(initialRCMData))
  }

  const saveToLocalStorage = () => {
    if (import.meta.client) {
      localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(rcmList.value))
    }
  }

  initStoreData()



  // Calculate rating (1-5) to percentage (4%, 8%, 12%, 16%, 20%)
  const ratingToPercent = (rating: number): number => {
    const validRating = Math.max(1, Math.min(5, rating || 1))
    return validRating * 4
  }

  // Calculate Total Weighted Score in % (Sum of 5 dimensions percentage)
  const calculateItemScorePercent = (item: Partial<RCMItem>): number => {
    const des = ratingToPercent(item.design_effectiveness_rating || 3)
    const op = ratingToPercent(item.operating_effectiveness_rating || 3)
    const cov = ratingToPercent(item.coverage_completeness_rating || 3)
    const time = ratingToPercent(item.timeliness_rating || 3)
    const auto = ratingToPercent(item.automation_monitoring_rating || 3)
    return Math.round(des + op + cov + time + auto)
  }

  // Filtered RCM list by selected year & department
  const filteredRCMList = computed(() => {
    const currentYear = new Date().getFullYear()
    const isFutureYear = selectedYear.value > currentYear

    return rcmList.value.filter(item => {
      const itemYear = item.year || (isFutureYear ? null : 2026)
      const matchYear = itemYear === selectedYear.value
      const matchDept = selectedDepartment.value === 'All Departments' || item.department === selectedDepartment.value
      return matchYear && matchDept
    })
  })

  // Synchronized Inherent Risk & Residual Risk from Corporate Risk Profile
  const synchronizedRiskCounts = computed(() => {
    const risks = riskProfileStore.risks || []
    
    // Filter risks by branch/department if selected
    const filteredRisks = risks.filter(r => {
      if (selectedDepartment.value === 'All Departments') return true
      return r.branch === selectedDepartment.value || r.category === selectedDepartment.value
    })

    if (filteredRisks.length === 0) {
      return {
        inherent: 0,
        residual: 0
      }
    }

    // Inherent Risks: Priority risks (Moderate, Moderate to High, High)
    const priorityRisks = filteredRisks.filter(r => {
      const lvl = riskProfileStore.getRiskLevel(r.likelihood, r.impact)
      return lvl === RiskLevel.MODERATE || lvl === RiskLevel.MODERATE_HIGH || lvl === RiskLevel.HIGH
    })

    const inherentCount = priorityRisks.length || filteredRisks.length

    // Residual Risks: Remaining high/moderate risks after mitigation
    const residualRisks = priorityRisks.filter(r => {
      const assessment = r.assessments?.find((a: any) => a.year === selectedYear.value)
      const q4Impact = assessment?.impact_q4 ?? r.impact
      const q4Likelihood = assessment?.likelihood_q4 ?? r.likelihood
      const q4Level = riskProfileStore.getRiskLevel(q4Likelihood, q4Impact)
      return q4Level === RiskLevel.MODERATE || q4Level === RiskLevel.MODERATE_HIGH || q4Level === RiskLevel.HIGH
    })

    const residualCount = residualRisks.length

    return {
      inherent: inherentCount,
      residual: residualCount
    }
  })

  const totalInherentRisk = computed(() => synchronizedRiskCounts.value.inherent)
  const totalResidualRisk = computed(() => synchronizedRiskCounts.value.residual)

  // Per-department exposure for the "Inherent vs Residual Risk by Department" chart
  const inherentVsResidualByDepartment = computed(() =>
    computeInherentVsResidualByDepartment(riskProfileStore.rawRisks, selectedYear.value)
  )

  // Formula: (1 - Residual Risk / Inherent Risk) * 100%
  const internalControlEffectiveness = computed(() => {
    const inh = totalInherentRisk.value
    const res = totalResidualRisk.value
    if (!inh || inh <= 0) return 0
    const eff = (1 - (res / inh)) * 100
    return Math.round(eff * 10) / 10
  })

  const effectivenessRating = computed(() => {
    return getEffectivenessInterpretation(internalControlEffectiveness.value)
  })

  // COSO 2013 Averages in %
  const cosoAverages = computed(() => {
    const list = filteredRCMList.value
    if (list.length === 0) {
      return {
        design: 0,
        operating: 0,
        coverage: 0,
        timeliness: 0,
        automation: 0,
        totalWeighted: 0
      }
    }

    const count = list.length
    let designSum = 0, operatingSum = 0, coverageSum = 0, timelinessSum = 0, automationSum = 0, totalSum = 0

    list.forEach(item => {
      designSum += ratingToPercent(item.design_effectiveness_rating)
      operatingSum += ratingToPercent(item.operating_effectiveness_rating)
      coverageSum += ratingToPercent(item.coverage_completeness_rating)
      timelinessSum += ratingToPercent(item.timeliness_rating)
      automationSum += ratingToPercent(item.automation_monitoring_rating)
      totalSum += item.total_weighted_score || 0
    })

    return {
      design: Math.round(designSum / count),
      operating: Math.round(operatingSum / count),
      coverage: Math.round(coverageSum / count),
      timeliness: Math.round(timelinessSum / count),
      automation: Math.round(automationSum / count),
      totalWeighted: Math.round(totalSum / count)
    }
  })

  const isValidUUID = (str?: string) => Boolean(str && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(str))

  // Fetch RCM list from backend if available
  const fetchRCMList = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/rcm`)
      if (response && response.success && Array.isArray(response.data) && response.data.length > 0) {
        rcmList.value = response.data.map((item: any) => ({
          ...item,
          total_weighted_score: calculateItemScorePercent(item)
        }))
        saveToLocalStorage()
      }
    } catch (error: any) {
      console.warn('Failed to fetch RCM list from backend:', error)
      errorMsg.value = extractErrorMessage(error, 'Gagal mengambil data RCM dari server.')
    } finally {
      loading.value = false
    }
  }

  // CRUD actions
  const addRCMItem = async (newItem: Omit<RCMItem, 'id' | 'total_weighted_score'> & { id?: string }) => {
    loading.value = true
    errorMsg.value = ''
    const scorePercent = calculateItemScorePercent(newItem)

    const requestPayload: any = {
      risk_code: newItem.risk_code,
      risk_event: newItem.risk_event,
      control_code: newItem.control_code,
      control_description: newItem.control_description,
      control_type: newItem.control_type || 'Preventive',
      control_owner: newItem.control_owner || 'Department Lead',
      department: newItem.department || 'Head Office',
      year: newItem.year || selectedYear.value,
      design_effectiveness_weight: newItem.design_effectiveness_weight || 20,
      design_effectiveness_rating: newItem.design_effectiveness_rating || 3,
      operating_effectiveness_weight: newItem.operating_effectiveness_weight || 20,
      operating_effectiveness_rating: newItem.operating_effectiveness_rating || 3,
      coverage_completeness_weight: newItem.coverage_completeness_weight || 20,
      coverage_completeness_rating: newItem.coverage_completeness_rating || 3,
      timeliness_weight: newItem.timeliness_weight || 20,
      timeliness_rating: newItem.timeliness_rating || 3,
      automation_monitoring_weight: newItem.automation_monitoring_weight || 20,
      automation_monitoring_rating: newItem.automation_monitoring_rating || 3,
      total_weighted_score: scorePercent,
      notes: newItem.notes || ''
    }

    if (isValidUUID(newItem.risk_id)) {
      requestPayload.risk_id = newItem.risk_id
    }

    // Temporary local ID in case offline
    const tempId = `rcm-${Date.now()}`
    const localItem: RCMItem = {
      ...requestPayload,
      id: tempId,
      risk_id: newItem.risk_id
    }

    rcmList.value.unshift(localItem)
    saveToLocalStorage()

    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/rcm`, {
        method: 'POST',
        body: requestPayload
      })

      if (response && (response.success || response.id)) {
        const serverItem = response.data || response
        if (serverItem && serverItem.id) {
          localItem.id = serverItem.id
          if (serverItem.risk_id) {
            localItem.risk_id = serverItem.risk_id
          }
          saveToLocalStorage()
        }
      }
    } catch (error: any) {
      console.warn('Backend create error, saved locally.', error)
      errorMsg.value = extractErrorMessage(error, 'Gagal menyimpan ke server backend, data disimpan lokal.')
      throw error
    } finally {
      loading.value = false
    }
  }

  const updateRCMItem = async (updatedItem: RCMItem) => {
    loading.value = true
    errorMsg.value = ''
    updatedItem.total_weighted_score = calculateItemScorePercent(updatedItem)
    const idx = rcmList.value.findIndex(item => item.id === updatedItem.id)
    if (idx !== -1) {
      rcmList.value[idx] = { ...updatedItem }
      saveToLocalStorage()
    }

    const requestPayload: any = {
      ...updatedItem,
      risk_id: isValidUUID(updatedItem.risk_id) ? updatedItem.risk_id : undefined,
      control_type: updatedItem.control_type || 'Preventive'
    }

    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/rcm/${updatedItem.id}`, {
        method: 'PUT',
        body: requestPayload
      })
      if (response && (response.success || response.id)) {
        const serverItem = response.data || response
        if (serverItem && idx !== -1) {
          rcmList.value[idx] = {
            ...rcmList.value[idx],
            ...serverItem,
            total_weighted_score: calculateItemScorePercent(serverItem)
          }
          saveToLocalStorage()
        }
      }
    } catch (error: any) {
      console.warn('Backend update error, updated locally.', error)
      errorMsg.value = extractErrorMessage(error, 'Gagal mengupdate ke server backend, data diupdate lokal.')
      throw error
    } finally {
      loading.value = false
    }
  }

  const deleteRCMItem = async (id: string) => {
    loading.value = true
    errorMsg.value = ''
    rcmList.value = rcmList.value.filter(item => item.id !== id)
    saveToLocalStorage()

    try {
      const baseUrl = getRiskServiceBaseUrl()
      await $fetch(`${baseUrl}/rcm/${id}`, {
        method: 'DELETE'
      })
    } catch (error: any) {
      console.warn('Backend delete error, deleted locally.', error)
      errorMsg.value = extractErrorMessage(error, 'Gagal menghapus dari server backend.')
      throw error
    } finally {
      loading.value = false
    }
  }

  return {
    rcmList,
    filteredRCMList,
    selectedYear,
    selectedDepartment,
    columns,
    totalInherentRisk,
    totalResidualRisk,
    synchronizedRiskCounts,
    inherentVsResidualByDepartment,
    internalControlEffectiveness,
    effectivenessRating,
    cosoAverages,
    ratingToPercent,
    calculateItemScorePercent,
    fetchRCMList,
    addRCMItem,
    updateRCMItem,
    deleteRCMItem,
    loading,
    errorMsg
  }
})
