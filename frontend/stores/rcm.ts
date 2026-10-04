import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import type { TableColumn } from '@nuxt/ui'
import { useRiskProfileStore, riskLevelConfig } from '~/stores/risk-profile'
import { useMitigationStore } from '~/stores/mitigation-risk'
import { RiskLevel } from '~/types/risk'
import { extractErrorMessage } from '~/utils/error'
import { getRiskServiceBaseUrl } from '~/composables/useApiUrl'

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
  const groups = new Map<string, { inherent: number, residual: number, count: number }>()

  ;(rawRisks || []).forEach((risk: any) => {
    const department = risk?.branch || risk?.department
    if (!department) return

    const assessment = risk.assessments?.find((a: any) => a.year === year)
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
  const selectedYear = ref(2026)
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

  // RCM rows come only from the risk-service API (no dummy data, no local cache).
  // Clear the cache older builds kept in localStorage so stale rows can't resurface.
  if (typeof localStorage !== 'undefined') {
    try {
      localStorage.removeItem('rcm_items_v2')
    } catch {
      // storage unavailable (private mode etc.) — nothing to clear
    }
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

  const withScore = (item: any): RCMItem => ({
    ...item,
    total_weighted_score: calculateItemScorePercent(item)
  })

  const buildPayload = (item: Partial<RCMItem>) => {
    const payload: any = {
      risk_code: item.risk_code,
      risk_event: item.risk_event,
      control_code: item.control_code,
      control_description: item.control_description,
      control_type: item.control_type || 'Preventive',
      control_owner: item.control_owner || 'Department Lead',
      department: item.department || 'Head Office',
      year: item.year || selectedYear.value,
      design_effectiveness_weight: item.design_effectiveness_weight || 20,
      design_effectiveness_rating: item.design_effectiveness_rating || 3,
      operating_effectiveness_weight: item.operating_effectiveness_weight || 20,
      operating_effectiveness_rating: item.operating_effectiveness_rating || 3,
      coverage_completeness_weight: item.coverage_completeness_weight || 20,
      coverage_completeness_rating: item.coverage_completeness_rating || 3,
      timeliness_weight: item.timeliness_weight || 20,
      timeliness_rating: item.timeliness_rating || 3,
      automation_monitoring_weight: item.automation_monitoring_weight || 20,
      automation_monitoring_rating: item.automation_monitoring_rating || 3,
      total_weighted_score: calculateItemScorePercent(item),
      notes: item.notes || ''
    }
    if (isValidUUID(item.risk_id)) {
      payload.risk_id = item.risk_id
    }
    return payload
  }

  // Fetch RCM list from the backend. Empty response → empty list; error → empty list + errorMsg.
  const fetchRCMList = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/rcm`)
      const items = Array.isArray(response?.data) ? response.data : []
      rcmList.value = items.map(withScore)
    } catch (error: any) {
      console.warn('Failed to fetch RCM list from backend:', error)
      rcmList.value = []
      errorMsg.value = extractErrorMessage(error, 'Gagal mengambil data RCM dari server.')
    } finally {
      loading.value = false
    }
  }

  // CRUD actions — the list only changes after the server confirms the write.
  const addRCMItem = async (newItem: Omit<RCMItem, 'id' | 'total_weighted_score'> & { id?: string }) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/rcm`, {
        method: 'POST',
        body: buildPayload(newItem)
      })
      const serverItem = response?.data
      if (serverItem?.id) {
        rcmList.value.unshift(withScore(serverItem))
      } else {
        await fetchRCMList()
      }
    } catch (error: any) {
      console.warn('Backend create error:', error)
      errorMsg.value = extractErrorMessage(error, 'Gagal menyimpan data RCM ke server.')
      throw error
    } finally {
      loading.value = false
    }
  }

  const updateRCMItem = async (updatedItem: RCMItem) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/rcm/${updatedItem.id}`, {
        method: 'PUT',
        body: buildPayload(updatedItem)
      })
      const serverItem = response?.data
      const idx = rcmList.value.findIndex(item => item.id === updatedItem.id)
      if (serverItem?.id && idx !== -1) {
        rcmList.value[idx] = withScore(serverItem)
      } else {
        await fetchRCMList()
      }
    } catch (error: any) {
      console.warn('Backend update error:', error)
      errorMsg.value = extractErrorMessage(error, 'Gagal mengupdate data RCM di server.')
      throw error
    } finally {
      loading.value = false
    }
  }

  const deleteRCMItem = async (id: string) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      await $fetch(`${baseUrl}/rcm/${id}`, {
        method: 'DELETE'
      })
      rcmList.value = rcmList.value.filter(item => item.id !== id)
    } catch (error: any) {
      console.warn('Backend delete error:', error)
      errorMsg.value = extractErrorMessage(error, 'Gagal menghapus data RCM dari server.')
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
