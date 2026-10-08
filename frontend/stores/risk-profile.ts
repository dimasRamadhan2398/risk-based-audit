import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { RiskLevel, ImpactLevel, PossibilityLevel } from '~/types/risk'
import { extractErrorMessage } from '~/utils/error'
import { getRiskServiceBaseUrl } from '~/composables/useApiUrl'
import { useLocationApi } from '~/composables/useLocationApi'
import { useFiscalYear } from '~/composables/useFiscalYear'

// --- Constants (Exported for components) ---

export const impactLabels: Record<number, string> = {
  [ImpactLevel.VERY_LOW]: 'Very Low',
  [ImpactLevel.LOW]: 'Low',
  [ImpactLevel.MODERATE]: 'Moderate',
  [ImpactLevel.HIGH]: 'High',
  [ImpactLevel.VERY_HIGH]: 'Very High'
}

export const likelihoodLabels: Record<number, string> = {
  [PossibilityLevel.VERY_RARE]: 'Very Rare',
  [PossibilityLevel.RARE]: 'Rare',
  [PossibilityLevel.POSSIBLE]: 'Possible',
  [PossibilityLevel.LIKELY]: 'Likely',
  [PossibilityLevel.VERY_LIKELY]: 'Very Likely'
}

export const categoryIcons: Record<string, string> = {
  'Financial': '💰',
  'Technology': '🔒',
  'Compliance': '📋',
  'Governance': '🏛️',
  'Operations': '⚙️',
  'Human Resources': '👥',
  'Strategic': '🎯'
}

export const categoryPrefixes: Record<string, string> = {
  'Financial': 'FIN',
  'Technology': 'TEC',
  'Compliance': 'COM',
  'Governance': 'GOV',
  'Operations': 'OPR',
  'Human Resources': 'HUM',
  'Strategic': 'STR'
}

export const riskLevelConfig = {
  [RiskLevel.LOW]: {
    label: 'Low',
    color: '#4CAF50',
    bg: '#1B5E20',
    priority: false
  },
  [RiskLevel.LOW_MODERATE]: {
    label: 'Low to Moderate',
    color: '#8BC34A',
    bg: '#33691E',
    priority: false
  },
  [RiskLevel.MODERATE]: {
    label: 'Moderate',
    color: '#FFC107',
    bg: '#FF6F00',
    priority: true
  },
  [RiskLevel.MODERATE_HIGH]: {
    label: 'Moderate to High',
    color: '#FF9800',
    bg: '#E65100',
    priority: true
  },
  [RiskLevel.HIGH]: {
    label: 'High',
    color: '#F44336',
    bg: '#B71C1C',
    priority: true
  }
}

// Branch filter sentinels. Real options use the Location master ID as value.
export const ALL_BRANCHES = 'All Branches'
export const UNASSIGNED_BRANCH = '__unassigned__'

// --- Store Definition ---

export const useRiskProfileStore = defineStore('risk-profile', () => {
  const config = useRuntimeConfig()
  const rawRisks = ref<any[]>([])
  const masterLocations = ref<{ id: string, name: string }[]>([])
  const branchesLoading = ref(false)
  // Set when the Location master could not be loaded — the branch filter is
  // then disabled instead of offering made-up branches.
  const branchesError = ref('')
  const loading = ref(false)
  const errorMsg = ref('')
  // Set only when GET /risks fails, so the page can show an error state
  // (errorMsg is shared with add/update/delete).
  const risksLoadError = ref('')

  // UI State
  // ALL_BRANCHES, UNASSIGNED_BRANCH or a Location master ID.
  const selectedBranch = ref<string>(ALL_BRANCHES)
  const selectedRisk = ref(null)
  const isFormOpen = ref(false)
  const isDetailOpen = ref(false)
  const modalMode = ref('preview')

  // Year & Quarter State
  const { selectedFiscalYear } = useFiscalYear()
  const selectedYear = selectedFiscalYear
  const selectedPeriod = ref('Q1')



  /** Branch names, Location master only (active rows). */
  const branchesList = computed(() => masterLocations.value.map(l => l.name))

  /**
   * The Location master ID for a branch name, so a saved risk carries a real
   * reference (risk_profile.location_id) and not just a label.
   */
  const locationIdForBranch = (branch?: string | null): string | undefined => {
    if (!branch) return undefined
    return masterLocations.value.find(l => l.name === branch)?.id
  }

  const locationName = (locationId?: string | null): string | undefined => {
    if (!locationId) return undefined
    return masterLocations.value.find(l => l.id === locationId)?.name
  }

  /**
   * The Location master ID a risk belongs to. Uses location_id; the branch
   * name is matched only for rows that are not linked yet.
   */
  const riskLocationId = (risk: any): string | null => {
    if (!risk) return null
    if (risk.location_id) return String(risk.location_id)
    return locationIdForBranch(risk.branch) ?? null
  }

  /**
   * Normalise the location fields of a risk payload: send location_id plus the
   * master name for it. A branch name that is not master data is dropped (the
   * backend rejects it), leaving the risk unassigned.
   */
  const withLocationId = (payload: any) => {
    const { location_id: _locationId, branch: _branch, ...rest } = payload || {}
    const locationId = riskLocationId(payload)
    if (!locationId) return rest
    return { ...rest, location_id: locationId, branch: locationName(locationId) ?? payload?.branch }
  }

  /** Load branches from the Location master (/api/v1/locations). */
  const fetchBranches = async () => {
    branchesLoading.value = true
    branchesError.value = ''
    try {
      const locations = await useLocationApi().getLocations()
      masterLocations.value = locations
        .filter((l: any) => l.is_active !== false && l.id && l.name)
        .map((l: any) => ({ id: String(l.id), name: l.name }))
    } catch (error: any) {
      console.error('Failed to fetch branch master data:', error)
      branchesError.value = extractErrorMessage(error, 'Failed to load branch master data.')
      masterLocations.value = []
    } finally {
      branchesLoading.value = false
    }
  }

  // Dynamic mapped risks based on selectedYear and selectedPeriod
  const risks = computed(() => {
    const currentYear = new Date().getFullYear()
    const isFutureYear = selectedYear.value > currentYear

    return rawRisks.value
      .filter(risk => {
        if (isFutureYear) {
          // Data 3 tahun kedepan dikosongkan terlebih dahulu karena harus diinput oleh user
          return risk.assessments?.some((a: any) => a.year === selectedYear.value)
        }
        return true
      })
      .map(risk => {
        const assessment = risk.assessments?.find((a: any) => a.year === selectedYear.value)
        const periodKey = selectedPeriod.value.toLowerCase() // 'q1', 'q2', etc.

        const impact = (assessment && assessment[`impact_${periodKey}`]) ? assessment[`impact_${periodKey}`] : risk.impact
        const likelihood = (assessment && assessment[`likelihood_${periodKey}`]) ? assessment[`likelihood_${periodKey}`] : risk.likelihood

        const riskLevel = assessment ? assessment[`risk_level_${periodKey}`] : 'Low'

        return {
          ...risk,
          impact: impact || 3,
          likelihood: likelihood || 3,
          riskLevel: riskLevel || 'Low'
        }
      })
  })

  /** True when a risk of the selected year has no Location master branch. */
  const hasUnassignedRisks = computed(() => risks.value.some(r => !riskLocationId(r)))

  /** Risks of the selected year/period narrowed by the branch filter. */
  const filteredRisks = computed(() => {
    const selected = selectedBranch.value
    if (!selected || selected === ALL_BRANCHES) return risks.value
    if (selected === UNASSIGNED_BRANCH) return risks.value.filter(r => !riskLocationId(r))
    return risks.value.filter(r => riskLocationId(r) === selected)
  })

  // Load risks from backend. No mock fallback: an empty or failed response
  // leaves the list empty so the page shows its empty/error state.
  const fetchRisks = async () => {
    loading.value = true
    errorMsg.value = ''
    risksLoadError.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risks`)
      const data = Array.isArray(response?.data) ? response.data : []
      rawRisks.value = data.map((r: any, idx: number) => ({
        ...r,
        displayId: idx + 1
      }))
    } catch (error: any) {
      console.error('Failed to fetch risks:', error)
      risksLoadError.value = extractErrorMessage(error, 'Failed to fetch risks.')
      errorMsg.value = risksLoadError.value
      rawRisks.value = []
    } finally {
      loading.value = false
    }
  }

  // Fetch immediately
  fetchRisks()
  fetchBranches()

  // Helper to re-map display IDs after changes
  const updateDisplayIds = () => {
    rawRisks.value.forEach((r: any, idx: number) => {
      r.displayId = idx + 1
    })
  }

  // Helpers
  const getRiskLevel = (likelihood?: number, impact?: number): RiskLevel => {
    if (!likelihood || !impact) return RiskLevel.LOW
    const matrix: Record<number, Record<number, RiskLevel>> = {
      5: { 1: RiskLevel.LOW_MODERATE, 2: RiskLevel.MODERATE, 3: RiskLevel.MODERATE_HIGH, 4: RiskLevel.HIGH, 5: RiskLevel.HIGH },
      4: { 1: RiskLevel.LOW, 2: RiskLevel.LOW_MODERATE, 3: RiskLevel.MODERATE, 4: RiskLevel.MODERATE_HIGH, 5: RiskLevel.HIGH },
      3: { 1: RiskLevel.LOW, 2: RiskLevel.LOW_MODERATE, 3: RiskLevel.MODERATE, 4: RiskLevel.MODERATE_HIGH, 5: RiskLevel.HIGH },
      2: { 1: RiskLevel.LOW, 2: RiskLevel.LOW_MODERATE, 3: RiskLevel.LOW_MODERATE, 4: RiskLevel.MODERATE_HIGH, 5: RiskLevel.HIGH },
      1: { 1: RiskLevel.LOW, 2: RiskLevel.LOW, 3: RiskLevel.LOW_MODERATE, 4: RiskLevel.MODERATE, 5: RiskLevel.HIGH }
    }
    return matrix[likelihood]?.[impact] || RiskLevel.LOW
  }

  const getRiskScore = (likelihood?: number, impact?: number): number => {
    if (!likelihood || !impact) return 0
    const matrix: Record<number, Record<number, number>> = {
      5: { 1: 7, 2: 12, 3: 17, 4: 22, 5: 25 },
      4: { 1: 4, 2: 9, 3: 14, 4: 19, 5: 24 },
      3: { 1: 3, 2: 8, 3: 13, 4: 18, 5: 23 },
      2: { 1: 2, 2: 6, 3: 11, 4: 16, 5: 21 },
      1: { 1: 1, 2: 5, 3: 10, 4: 15, 5: 20 }
    }
    return matrix[likelihood]?.[impact] || 0
  }

  const getFormattedId = (risk: any): string => {
    if (!risk) return 'RSK-000'
    const prefix = categoryPrefixes[risk.category] || 'RSK'
    const idStr = String(risk.displayId || risk.id || 0).padStart(3, '0')
    return `${prefix}-${idStr}`
  }

  const getRiskById = (id: string | number) => {
    return risks.value.find(r => String(r.id) === String(id))
  }

  // Actions
  function openAddModal() {
    selectedRisk.value = null
    modalMode.value = 'add'
    isFormOpen.value = true
  }

  function openEditModal(risk: any) {
    selectedRisk.value = { ...risk }
    modalMode.value = 'edit'
    isFormOpen.value = true
  }

  function openPreviewModal(risk: any) {
    selectedRisk.value = { ...risk }
    isDetailOpen.value = true
  }

  /** Create a risk. Returns false (and sets errorMsg) when the backend rejects it. */
  async function addRisk(newRiskData: any): Promise<boolean> {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const body = withLocationId(newRiskData)
      const response: any = await $fetch(`${baseUrl}/risks`, {
        method: 'POST',
        body
      })

      const isSuccess = response && (response.success || response.id)
      const responseData = response.data || response

      if (isSuccess) {
        const createdRisk = {
          ...body, // Keep intended local data like assessments
          ...responseData,
          assessments: newRiskData.assessments, // Force keep nested assessments if backend drops it
          displayId: rawRisks.value.length + 1
        }
        rawRisks.value.push(createdRisk)
        return true
      }
      return false
    } catch (error: any) {
      // No local fallback row: a risk the backend did not store must not show up.
      console.error('Failed to add risk:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to add risk.')
      return false
    } finally {
      loading.value = false
    }
  }

  /** Update a risk. Returns false (and rolls back the optimistic change) on failure. */
  async function updateRisk(updatedRisk: any): Promise<boolean> {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()

      // Find raw risk to copy assessments
      const rawRisk = rawRisks.value.find(r => String(r.id) === String(updatedRisk.id))
      let payload = { ...updatedRisk }

      if (rawRisk) {
        // Copy raw assessments
        const assessments = rawRisk.assessments ? JSON.parse(JSON.stringify(rawRisk.assessments)) : []

        // Find or create assessment for selectedYear
        let ast = assessments.find((a: any) => a.year === selectedYear.value)
        const periodKey = selectedPeriod.value.toLowerCase() // 'q1', 'q2', 'q3', 'q4'
        if (!ast) {
          ast = {
            year: selectedYear.value,
            impact_q1: rawRisk.impact || 3, impact_q2: rawRisk.impact || 3, impact_q3: rawRisk.impact || 3, impact_q4: rawRisk.impact || 3,
            likelihood_q1: rawRisk.likelihood || 3, likelihood_q2: rawRisk.likelihood || 3, likelihood_q3: rawRisk.likelihood || 3, likelihood_q4: rawRisk.likelihood || 3,
            risk_level_q1: 'Low', risk_level_q2: 'Low', risk_level_q3: 'Low', risk_level_q4: 'Low'
          }
          assessments.push(ast)
        }

        // Copy explicitly mapped quarterly values from updatedRisk
        ['q1', 'q2', 'q3', 'q4'].forEach(q => {
          if (updatedRisk[`impact_${q}`] !== undefined) {
            ast[`impact_${q}`] = updatedRisk[`impact_${q}`]
          }
          if (updatedRisk[`likelihood_${q}`] !== undefined) {
            ast[`likelihood_${q}`] = updatedRisk[`likelihood_${q}`]
          }
          if (updatedRisk[`risk_level_${q}`] !== undefined) {
            ast[`risk_level_${q}`] = updatedRisk[`risk_level_${q}`]
          }
        })

        // Fallback for drag-and-drop which only updates base impact/likelihood
        if (updatedRisk[`impact_${periodKey}`] === undefined && updatedRisk.impact !== undefined) {
          ast[`impact_${periodKey}`] = updatedRisk.impact
        }
        if (updatedRisk[`likelihood_${periodKey}`] === undefined && updatedRisk.likelihood !== undefined) {
          ast[`likelihood_${periodKey}`] = updatedRisk.likelihood
        }

        // Attach assessments to payload
        payload.assessments = assessments
      }
      payload = withLocationId(payload)

      // Optimistic local update
      const previousRawRisks = rawRisks.value
      const idx = rawRisks.value.findIndex(r => String(r.id) === String(updatedRisk.id))
      if (idx !== -1) {
        const newRawRisks = [...rawRisks.value]
        newRawRisks[idx] = {
          ...payload,
          displayId: newRawRisks[idx].displayId
        }
        rawRisks.value = newRawRisks
      }

      try {
        const response: any = await $fetch(`${baseUrl}/risks/${updatedRisk.id}`, {
          method: 'PUT',
          body: payload
        })
        if (response && response.success) {
          if (idx !== -1) {
            const newRawRisks = [...rawRisks.value]
            newRawRisks[idx] = {
              ...response.data,
              ...payload,
              displayId: newRawRisks[idx].displayId
            }
            rawRisks.value = newRawRisks
          }
        }
        return true
      } catch (error: any) {
        console.error('Failed to update risk on backend (rolled back):', error)
        errorMsg.value = extractErrorMessage(error, 'Failed to update risk on backend.')
        rawRisks.value = previousRawRisks
        return false
      }
    } catch (error: any) {
      console.error('Failed to update risk:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to update risk.')
      return false
    } finally {
      loading.value = false
    }
  }

  async function deleteRisk(id: string | number): Promise<boolean> {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getRiskServiceBaseUrl()

      const response: any = await $fetch(`${baseUrl}/risks/${id}`, {
        method: 'DELETE'
      })

      if (response && response.success) {
        rawRisks.value = rawRisks.value.filter(
          r => String(r.id) !== String(id)
        )

        updateDisplayIds()

        return true
      }

      return false
    } catch (error: any) {
      console.error('Failed to delete risk:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to delete risk.')
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    rawRisks,
    risks,
    branches: branchesList,
    locations: masterLocations,
    branchesLoading,
    branchesError,
    hasUnassignedRisks,
    filteredRisks,
    selectedBranch,
    selectedRisk,
    isFormOpen,
    isDetailOpen,
    modalMode,
    selectedYear,
    selectedPeriod,
    getRiskLevel,
    getRiskScore,
    getFormattedId,
    getRiskById,
    openAddModal,
    openEditModal,
    openPreviewModal,
    addRisk,
    updateRisk,
    deleteRisk,
    fetchRisks,
    fetchBranches,
    locationIdForBranch,
    locationName,
    riskLocationId,
    loading,
    errorMsg,
    risksLoadError
  }
})
