import { defineStore } from 'pinia'
import { ref, reactive, computed, watch } from 'vue'
import { RiskLevel, ImpactLevel, PossibilityLevel } from '~/types/risk'
import { extractErrorMessage } from '~/utils/error'
import { getRiskServiceBaseUrl } from '~/composables/useApiUrl'
import { useLocationApi } from '~/composables/useLocationApi'
import { useFiscalYear } from '~/composables/useFiscalYear'
import { currentFiscalYear, currentQuarter } from '~/utils/fiscalPeriod'
import { ALL_RISK_CATEGORIES, getDistinctRiskCategories } from '~/utils/riskCategoryFilter'

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

// Lucide icon per risk category; the keys are also the category options.
export const categoryIcons: Record<string, string> = {
  'Financial': 'i-lucide-wallet',
  'Technology': 'i-lucide-lock',
  'Compliance': 'i-lucide-clipboard-list',
  'Governance': 'i-lucide-landmark',
  'Operations': 'i-lucide-settings',
  'Human Resources': 'i-lucide-users',
  'Strategic': 'i-lucide-target'
}

export const getCategoryIcon = (category?: string): string => (category && categoryIcons[category]) || 'i-lucide-tag'

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

// Rows per page in the three history tables on the Corporate Risk Profile page.
export const RISK_PAGE_SIZE = 10

// Server-paginated tables ('priority' is derived client-side, see priorityPage).
export type RiskPageKey = 'all' | 'progress'

const newRiskPage = () => ({
  items: [] as any[],
  page: 1,
  pageSize: RISK_PAGE_SIZE,
  total: 0,
  totalPages: 0,
  loading: false,
  loaded: false,
  error: ''
})

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
  // Default period = current calendar quarter (same date as the year default).
  const selectedPeriod = ref<string>(currentQuarter())
  // Defaults this tab started with, so a long-lived tab can tell "untouched".
  let initialYear = selectedYear.value
  let initialPeriod = selectedPeriod.value

  /** Move an untouched year/quarter forward when the calendar rolled over while the tab stayed open. */
  const syncDefaultPeriod = (now: Date = new Date()) => {
    const year = currentFiscalYear(now)
    const quarter = currentQuarter(now)
    if (selectedYear.value === initialYear) selectedYear.value = year
    if (selectedPeriod.value === initialPeriod) selectedPeriod.value = quarter
    initialYear = year
    initialPeriod = quarter
  }

  // Category filter ('all' keeps every row) and free-text search; both drive
  // the server-side filters of the paginated tables.
  const selectedCategory = ref<string>(ALL_RISK_CATEGORIES)
  const searchQuery = ref('')

  // Distinct category names for the dropdown (GET /risks/categories).
  const categories = ref<string[]>([])
  const categoriesError = ref('')

  // Paginated tables: 'all' (All Risks) and 'progress' (Quarterly Progress)
  // each page their own server query; 'priority' is client-side (priorityPage).
  const riskPages = reactive({ all: newRiskPage(), progress: newRiskPage() })
  const priorityPage = ref(1)
  const pageRequestIds: Record<RiskPageKey, number> = { all: 0, progress: 0 }

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

  // The selected year's CRP: only risks assessed in that year, read as they
  // stood that year (the assessment's snapshot), so a risk added or edited in
  // one year does not show up in or rewrite another year. Fields missing from
  // a snapshot fall back to the latest version.
  const mapRiskForYear = (risk: any) => {
    const periodKey = selectedPeriod.value.toLowerCase() // 'q1', 'q2', etc.
    const assessment = risk.assessments?.find((a: any) => a.year === selectedYear.value)
    if (!assessment) return null

    return {
      ...risk,
      name: assessment.name || risk.name,
      category: assessment.category || risk.category,
      description: assessment.description || risk.description,
      location_id: assessment.location_id || risk.location_id,
      branch: assessment.location_id
        ? (locationName(assessment.location_id) ?? assessment.branch)
        : risk.branch,
      impact: assessment[`impact_${periodKey}`] || risk.impact || 3,
      likelihood: assessment[`likelihood_${periodKey}`] || risk.likelihood || 3,
      riskLevel: assessment[`risk_level_${periodKey}`] || 'Low'
    }
  }

  const risks = computed(() =>
    rawRisks.value.flatMap((risk) => {
      const mapped = mapRiskForYear(risk)
      return mapped ? [mapped] : []
    })
  )

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

  /** Query params for GET /risks?page=...; empty filters are omitted. */
  const buildRiskPageQuery = (page: number) => {
    const query: Record<string, string | number> = {
      page,
      page_size: RISK_PAGE_SIZE,
      fiscal_year: selectedYear.value
    }
    // period is not sent: server-side it only matters with sort_by=score. The
    // selected Q1..Q4 view is applied client-side in mapRiskForYear.
    if (selectedCategory.value && selectedCategory.value !== ALL_RISK_CATEGORIES) query.category = selectedCategory.value
    if (selectedBranch.value === UNASSIGNED_BRANCH) query.branch_id = 'unassigned'
    else if (selectedBranch.value && selectedBranch.value !== ALL_BRANCHES) query.branch_id = selectedBranch.value
    const search = searchQuery.value.trim()
    if (search) query.search = search
    return query
  }

  /** Rows of the selected year (with the displayId of the full list) for a page of raw API rows. */
  const mapPageItems = (items: any[], page: number) =>
    items.map((item, idx) => {
      const mapped = mapRiskForYear(item) ?? item
      const known = rawRisks.value.find(r => String(r.id) === String(item.id))
      return { ...mapped, displayId: known?.displayId ?? ((page - 1) * RISK_PAGE_SIZE + idx + 1) }
    })

  /** Load one page of a paginated table. page defaults to the table's current page. */
  const fetchRiskPage = async (key: RiskPageKey, page?: number) => {
    const state = riskPages[key]
    const target = Math.max(1, page ?? state.page)
    const requestId = ++pageRequestIds[key]
    state.loading = true
    state.error = ''
    try {
      // no-store: the API response is browser-cached per query string, which
      // would serve a page from before an add/update/delete.
      const response: any = await $fetch(`${getRiskServiceBaseUrl()}/risks`, {
        query: buildRiskPageQuery(target),
        cache: 'no-store'
      })
      if (requestId !== pageRequestIds[key]) return // a newer request owns the state
      // Real shape: { data: [...], pagination: {...} }; older { data: { items, pagination } } tolerated.
      const rows: any[] = Array.isArray(response?.data)
        ? response.data
        : Array.isArray(response?.data?.items) ? response.data.items : []
      const pagination = response?.pagination ?? response?.data?.pagination ?? {}
      const total = Number(pagination.total ?? rows.length)
      const totalPages = Number(pagination.total_pages ?? Math.ceil(total / RISK_PAGE_SIZE))
      // The page vanished (last row on it deleted): go to the new last page.
      if (rows.length === 0 && total > 0 && totalPages > 0 && target > totalPages) {
        await fetchRiskPage(key, totalPages)
        return
      }
      state.page = Number(pagination.page ?? target)
      state.items = mapPageItems(rows, state.page)
      state.total = total
      state.totalPages = totalPages
      state.loaded = true
    } catch (error: any) {
      if (requestId !== pageRequestIds[key]) return
      console.error('Failed to fetch risk page:', error)
      state.error = extractErrorMessage(error, 'Failed to fetch risks.')
      state.items = []
      state.total = 0
      state.totalPages = 0
    } finally {
      if (requestId === pageRequestIds[key]) state.loading = false
    }
  }

  /** Re-read the pages already loaded, keeping their current page (after add/update/delete). */
  const refreshRiskPages = async () => {
    await Promise.all(
      (Object.keys(riskPages) as RiskPageKey[])
        .filter(key => riskPages[key].loaded)
        .map(key => fetchRiskPage(key))
    )
  }

  const resetRiskPages = () => {
    riskPages.all.page = 1
    riskPages.progress.page = 1
    priorityPage.value = 1
  }

  // Any filter/search change starts again from page 1.
  watch(
    [selectedBranch, selectedCategory, selectedYear, selectedPeriod, searchQuery],
    resetRiskPages,
    { flush: 'sync' }
  )

  /** Distinct categories for the dropdown. A failure leaves the list empty (the page falls back to loaded data). */
  const fetchCategories = async () => {
    categoriesError.value = ''
    try {
      const response: any = await $fetch(`${getRiskServiceBaseUrl()}/risks/categories`, { cache: 'no-store' })
      const data = response?.data
      const list: any[] = Array.isArray(data) ? data : Array.isArray(data?.items) ? data.items : []
      categories.value = getDistinctRiskCategories(
        list.map(c => ({ category: typeof c === 'string' ? c : c?.name ?? c?.category }))
      )
    } catch (error: any) {
      console.error('Failed to fetch risk categories:', error)
      categoriesError.value = extractErrorMessage(error, 'Failed to load risk categories.')
      categories.value = []
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

  /**
   * CRP "Priority" list: only levels flagged `priority` in riskLevelConfig, highest matrix score
   * first, then highest severity. Shared by the CRP page and the dashboard Registered Risks card
   * so both show the same risks in the same order, whatever order the API returned.
   */
  const sortPriorityRisks = <T extends { likelihood?: number, impact?: number, severity?: number }>(list: readonly T[]): T[] =>
    list
      .filter(r => riskLevelConfig[getRiskLevel(r.likelihood, r.impact)]?.priority)
      .sort((a, b) => {
        const diff = getRiskScore(b.likelihood, b.impact) - getRiskScore(a.likelihood, a.impact)
        return diff !== 0 ? diff : (b.severity ?? 0) - (a.severity ?? 0)
      })

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
          assessments: responseData.assessments ?? newRiskData.assessments,
          displayId: rawRisks.value.length + 1
        }
        rawRisks.value.push(createdRisk)
        void refreshRiskPages()
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
      // The backend writes this year's assessment only; the latest version
      // moves only when this is the most recent year the risk is assessed in.
      let payload = { ...updatedRisk, year: selectedYear.value }
      let optimisticAssessments: any[] | undefined

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
        ['q1', 'q2', 'q3', 'q4'].forEach((q) => {
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

        // This year's snapshot, so the optimistic row reads the edit back
        const locationId = riskLocationId(updatedRisk)
        Object.assign(ast, {
          name: updatedRisk.name,
          category: updatedRisk.category,
          description: updatedRisk.description,
          ...(locationId ? { location_id: locationId, branch: locationName(locationId) ?? ast.branch } : {})
        })

        payload.assessments = [ast]
        optimisticAssessments = assessments
      }
      payload = withLocationId(payload)

      // Optimistic local update, mirroring the backend: this year's assessment
      // changes, and the latest version only when this is the latest year.
      const previousRawRisks = rawRisks.value
      const idx = rawRisks.value.findIndex(r => String(r.id) === String(updatedRisk.id))
      if (idx !== -1 && optimisticAssessments) {
        const isLatestYear = optimisticAssessments.every((a: any) => a.year <= selectedYear.value)
        const { year: _year, assessments: _assessments, ...latestFields } = payload
        const newRawRisks = [...rawRisks.value]
        newRawRisks[idx] = {
          ...newRawRisks[idx],
          ...(isLatestYear ? latestFields : {}),
          assessments: optimisticAssessments,
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
          if (idx !== -1 && response.data) {
            const newRawRisks = [...rawRisks.value]
            newRawRisks[idx] = {
              ...newRawRisks[idx],
              ...response.data,
              displayId: newRawRisks[idx].displayId
            }
            rawRisks.value = newRawRisks
          }
        }
        void refreshRiskPages()
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
        void refreshRiskPages()

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
    selectedCategory,
    searchQuery,
    categories,
    categoriesError,
    riskPages,
    priorityPage,
    fetchRiskPage,
    refreshRiskPages,
    resetRiskPages,
    fetchCategories,
    buildRiskPageQuery,
    selectedRisk,
    isFormOpen,
    isDetailOpen,
    modalMode,
    selectedYear,
    selectedPeriod,
    syncDefaultPeriod,
    getRiskLevel,
    getRiskScore,
    sortPriorityRisks,
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
