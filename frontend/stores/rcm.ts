import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { useRiskProfileStore, riskLevelConfig } from '~/stores/risk-profile'
import { useMitigationStore } from '~/stores/mitigation-risk'
import { RiskLevel } from '~/types/risk'

export interface RCMItem {
  id: string
  risk_id?: string
  risk_code: string
  risk_event: string
  control_code: string
  control_description: string
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

export const getEffectivenessInterpretation = (scorePercent: number) => {
  if (scorePercent >= 90) {
    return {
      score: scorePercent,
      rating: 'Highly Effective',
      interpretation: 'Controls reliably mitigate risk and require only routine monitoring.',
      color: 'emerald',
      bgClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      badgeColor: 'emerald'
    }
  } else if (scorePercent >= 80) {
    return {
      score: scorePercent,
      rating: 'Effective',
      interpretation: 'Controls function well; only minor improvements are recommended.',
      color: 'sky',
      bgClass: 'bg-sky-50 text-sky-700 border-sky-200',
      badgeColor: 'sky'
    }
  } else if (scorePercent >= 70) {
    return {
      score: scorePercent,
      rating: 'Moderately Effective',
      interpretation: 'Some weaknesses exist; corrective actions should be planned.',
      color: 'amber',
      bgClass: 'bg-amber-50 text-amber-700 border-amber-200',
      badgeColor: 'amber'
    }
  } else if (scorePercent >= 60) {
    return {
      score: scorePercent,
      rating: 'Weak',
      interpretation: 'Significant improvements are needed to reduce risk adequately.',
      color: 'orange',
      bgClass: 'bg-orange-50 text-orange-700 border-orange-200',
      badgeColor: 'orange'
    }
  } else {
    return {
      score: scorePercent,
      rating: 'Ineffective',
      interpretation: 'Controls do not provide sufficient risk mitigation and require immediate attention.',
      color: 'red',
      bgClass: 'bg-red-50 text-red-700 border-red-200',
      badgeColor: 'red'
    }
  }
}

export const useRCMStore = defineStore('rcm', () => {
  const config = useRuntimeConfig()
  const riskProfileStore = useRiskProfileStore()
  const rcmList = ref<RCMItem[]>([])
  const selectedYear = ref(2026)
  const selectedDepartment = ref('All Departments')

  const LOCAL_STORAGE_KEY = 'rcm_items_v2'

  // Initialize from LocalStorage or Fallback
  const initStoreData = () => {
    if (import.meta.client) {
      const saved = localStorage.getItem(LOCAL_STORAGE_KEY)
      if (saved) {
        try {
          rcmList.value = JSON.parse(saved)
          return
        } catch (e) {
          console.error('Failed to parse saved RCM items:', e)
        }
      }
    }
    rcmList.value = JSON.parse(JSON.stringify(initialRCMData))
    saveToLocalStorage()
  }

  const saveToLocalStorage = () => {
    if (import.meta.client) {
      localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(rcmList.value))
    }
  }

  initStoreData()

  const getRiskServiceBaseUrl = () => {
    return config.public.riskServiceBaseUrl || 'http://localhost:8004/api/v1'
  }

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
    return rcmList.value.filter(item => {
      const matchYear = !item.year || item.year === selectedYear.value
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

    // Inherent Risks: Priority risks (Moderate, Moderate to High, High)
    const priorityRisks = filteredRisks.filter(r => {
      const lvl = riskProfileStore.getRiskLevel(r.likelihood, r.impact)
      return lvl === RiskLevel.MODERATE || lvl === RiskLevel.MODERATE_HIGH || lvl === RiskLevel.HIGH
    })

    const inherentCount = priorityRisks.length || (filteredRisks.length > 0 ? filteredRisks.length : 20)

    // Residual Risks: Remaining high/moderate risks after mitigation
    // We compute residual risk score or count from assessments (Q4 / end of year)
    const residualRisks = priorityRisks.filter(r => {
      const assessment = r.assessments?.find((a: any) => a.year === selectedYear.value)
      const q4Impact = assessment?.impact_q4 ?? r.impact
      const q4Likelihood = assessment?.likelihood_q4 ?? r.likelihood
      const q4Level = riskProfileStore.getRiskLevel(q4Likelihood, q4Impact)
      return q4Level === RiskLevel.MODERATE || q4Level === RiskLevel.MODERATE_HIGH || q4Level === RiskLevel.HIGH
    })

    // Fallback calculation: If Q4 assessment isn't updated yet, simulate mitigation reduction
    const residualCount = residualRisks.length > 0 ? residualRisks.length : Math.max(1, Math.round(inherentCount * 0.4))

    return {
      inherent: inherentCount,
      residual: residualCount
    }
  })

  const totalInherentRisk = computed(() => synchronizedRiskCounts.value.inherent)
  const totalResidualRisk = computed(() => synchronizedRiskCounts.value.residual)

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

  // Fetch RCM list from backend if available
  const fetchRCMList = async () => {
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/rcm?year=${selectedYear.value}&department=${encodeURIComponent(selectedDepartment.value)}`)
      if (response && response.success && Array.isArray(response.data) && response.data.length > 0) {
        rcmList.value = response.data.map((item: any) => ({
          ...item,
          total_weighted_score: calculateItemScorePercent(item)
        }))
        saveToLocalStorage()
      }
    } catch (error) {
      // Backend unavailable, use local state
    }
  }

  // CRUD actions
  const addRCMItem = async (newItem: Omit<RCMItem, 'id' | 'total_weighted_score'>) => {
    const scorePercent = calculateItemScorePercent(newItem)
    const payload: RCMItem = {
      ...newItem,
      id: `rcm-${Date.now()}`,
      total_weighted_score: scorePercent
    }

    rcmList.value.unshift(payload)
    saveToLocalStorage()

    try {
      const baseUrl = getRiskServiceBaseUrl()
      await $fetch(`${baseUrl}/rcm`, {
        method: 'POST',
        body: payload
      })
    } catch (error) {
      console.warn('Backend create error, saved locally.', error)
    }
  }

  const updateRCMItem = async (updatedItem: RCMItem) => {
    updatedItem.total_weighted_score = calculateItemScorePercent(updatedItem)
    const idx = rcmList.value.findIndex(item => item.id === updatedItem.id)
    if (idx !== -1) {
      rcmList.value[idx] = { ...updatedItem }
      saveToLocalStorage()
    }

    try {
      const baseUrl = getRiskServiceBaseUrl()
      await $fetch(`${baseUrl}/rcm/${updatedItem.id}`, {
        method: 'PUT',
        body: updatedItem
      })
    } catch (error) {
      console.warn('Backend update error, updated locally.', error)
    }
  }

  const deleteRCMItem = async (id: string) => {
    rcmList.value = rcmList.value.filter(item => item.id !== id)
    saveToLocalStorage()

    try {
      const baseUrl = getRiskServiceBaseUrl()
      await $fetch(`${baseUrl}/rcm/${id}`, {
        method: 'DELETE'
      })
    } catch (error) {
      console.warn('Backend delete error, deleted locally.', error)
    }
  }

  return {
    rcmList,
    filteredRCMList,
    selectedYear,
    selectedDepartment,
    totalInherentRisk,
    totalResidualRisk,
    internalControlEffectiveness,
    effectivenessRating,
    cosoAverages,
    ratingToPercent,
    calculateItemScorePercent,
    fetchRCMList,
    addRCMItem,
    updateRCMItem,
    deleteRCMItem
  }
})
