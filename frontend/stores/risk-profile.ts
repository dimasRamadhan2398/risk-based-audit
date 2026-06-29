import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { RiskLevel, ImpactLevel, PossibilityLevel } from '~/types/risk'

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

const initialRiskData = [
  {
    id: 1,
    name: 'Target pendapatan dan laba tidak tercapai',
    impact: 5,
    likelihood: 4,
    severity: 98,
    category: 'Financial',
    branch: 'Head Office',
    description: 'Terget pendapatan tidak tercapai karena kinerja tim marketing yang kurang maksimal dan strategi marketing yang tidak efektif.',
    assessments: [
      { year: 2026, impact_q1: 5, impact_q2: 4, impact_q3: 3, impact_q4: 2, likelihood_q1: 4, likelihood_q2: 3, likelihood_q3: 2, likelihood_q4: 2 }
    ]
  },
  {
    id: 2,
    name: 'Target efisiensi biaya operasional dan umum tidak tercapai',
    impact: 5,
    likelihood: 4,
    severity: 95,
    category: 'Financial',
    branch: 'Head Office',
    description: 'Target efisiensi biaya operasional dan umum tidak tercapai karena kinerja tim keuangan yang kurang maksimal dan strategi keuangan yang tidak efektif.',
    assessments: [
      { year: 2026, impact_q1: 5, impact_q2: 4, impact_q3: 3, impact_q4: 2, likelihood_q1: 4, likelihood_q2: 3, likelihood_q3: 3, likelihood_q4: 2 }
    ]
  },
  {
    id: 3,
    name: 'Ancaman terhadap Cyber Security dan perlindungan data pribadi',
    impact: 5,
    likelihood: 4,
    severity: 88,
    category: 'Technology',
    branch: 'Head Office',
    description: 'Ancaman terhadap cyber security dan kebocoran data pelanggan/karyawan.',
    assessments: [
      { year: 2026, impact_q1: 5, impact_q2: 4, impact_q3: 3, impact_q4: 2, likelihood_q1: 4, likelihood_q2: 3, likelihood_q3: 2, likelihood_q4: 2 }
    ]
  },
  {
    id: 4,
    name: 'Terjadinya fraud',
    impact: 4,
    likelihood: 4,
    severity: 92,
    category: 'Financial',
    branch: 'Head Office',
    description: 'Penyalahgunaan wewenang atau kecurangan keuangan di lingkungan internal.',
    assessments: [
      { year: 2026, impact_q1: 4, impact_q2: 3, impact_q3: 2, impact_q4: 1, likelihood_q1: 4, likelihood_q2: 3, likelihood_q3: 2, likelihood_q4: 1 }
    ]
  },
  {
    id: 5,
    name: 'Implementasi teknologi dan digitalisasi tidak berhasil',
    impact: 4,
    likelihood: 3,
    severity: 72,
    category: 'Technology',
    branch: 'Bali Branch',
    description: 'Kegagalan implementasi sistem baru yang menghambat operasional.',
    assessments: [
      { year: 2026, impact_q1: 4, impact_q2: 3, impact_q3: 2, impact_q4: 2, likelihood_q1: 4, likelihood_q2: 3, likelihood_q3: 2, likelihood_q4: 2 }
    ]
  },
  {
    id: 6,
    name: 'Pengembangan kompetensi karyawan tidak terlaksana sesuai rencana',
    impact: 4,
    likelihood: 3,
    severity: 58,
    category: 'Human Resources',
    branch: 'Head Office',
    description: 'Kesenjangan keahlian karyawan akibat program training tidak berjalan.',
    assessments: [
      { year: 2026, impact_q1: 4, impact_q2: 3, impact_q3: 2, impact_q4: 2, likelihood_q1: 3, likelihood_q2: 3, likelihood_q3: 2, likelihood_q4: 2 }
    ]
  },
  {
    id: 7,
    name: 'Talent Attrition / Brain Drain',
    impact: 3,
    likelihood: 3,
    severity: 50,
    category: 'Human Resources',
    branch: 'Head Office',
    description: 'Loss of key employees and institutional knowledge affecting operational continuity.',
    assessments: [
      { year: 2026, impact_q1: 3, impact_q2: 3, impact_q3: 3, impact_q4: 3, likelihood_q1: 3, likelihood_q2: 3, likelihood_q3: 3, likelihood_q4: 3 }
    ]
  },
  {
    id: 8,
    name: 'Reputational Damage',
    impact: 4,
    likelihood: 2,
    severity: 75,
    category: 'Strategic',
    branch: 'Surabaya Branch',
    description: 'Significant brand damage due to public scandals, social media crises, or product failures.',
    assessments: [
      { year: 2026, impact_q1: 4, impact_q2: 4, impact_q3: 4, impact_q4: 4, likelihood_q1: 2, likelihood_q2: 2, likelihood_q3: 2, likelihood_q4: 2 }
    ]
  },
  {
    id: 9,
    name: 'Environmental Compliance Failure',
    impact: 3,
    likelihood: 2,
    severity: 55,
    category: 'Compliance',
    branch: 'Bandung Branch',
    description: 'Violations of environmental regulations leading to fines, shutdowns, or cleanup obligations.',
    assessments: [
      { year: 2026, impact_q1: 3, impact_q2: 3, impact_q3: 3, impact_q4: 3, likelihood_q1: 2, likelihood_q2: 2, likelihood_q3: 2, likelihood_q4: 2 }
    ]
  },
  {
    id: 10,
    name: 'Operational System Failure',
    impact: 4,
    likelihood: 3,
    severity: 70,
    category: 'Technology',
    branch: 'Bali Branch',
    description: 'Critical failure in core business systems causing operational downtime and revenue loss.',
    assessments: [
      { year: 2026, impact_q1: 4, impact_q2: 4, impact_q3: 4, impact_q4: 4, likelihood_q1: 3, likelihood_q2: 3, likelihood_q3: 3, likelihood_q4: 3 }
    ]
  }
]

const branches = [
  'Head Office',
  'Jakarta Branch',
  'Surabaya Branch',
  'Bandung Branch',
  'Bali Branch'
]

// --- Store Definition ---

export const useRiskProfileStore = defineStore('risk-profile', () => {
  const config = useRuntimeConfig()
  const rawRisks = ref<any[]>([])
  const branchesList = ref(branches)

  // UI State
  const selectedBranch = ref('All Branches')
  const selectedRisk = ref(null)
  const isFormOpen = ref(false)
  const isDetailOpen = ref(false)
  const modalMode = ref('preview')

  // Year & Quarter State
  const selectedYear = ref(2026)
  const selectedPeriod = ref('Q1')

  const getRiskServiceBaseUrl = () => {
    return config.public.riskServiceBaseUrl || 'http://localhost:8004/api/v1'
  }

  // Dynamic mapped risks based on selectedYear and selectedPeriod
  const risks = computed(() => {
    return rawRisks.value.map(risk => {
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

  // Load risks from backend
  const fetchRisks = async () => {
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risks`)
      if (response && response.success && Array.isArray(response.data) && response.data.length > 0) {
        rawRisks.value = response.data.map((r: any, idx: number) => ({
          ...r,
          displayId: idx + 1
        }))
      } else {
        rawRisks.value = initialRiskData.map((r: any, idx: number) => ({
          ...r,
          displayId: idx + 1
        }))
      }
    } catch (error) {
      console.error('Failed to fetch risks, falling back to mock data:', error)
      rawRisks.value = initialRiskData.map((r: any, idx: number) => ({
        ...r,
        displayId: idx + 1
      }))
    }
  }

  // Fetch immediately
  fetchRisks()

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

  async function addRisk(newRiskData: any) {
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risks`, {
        method: 'POST',
        body: newRiskData
      })
      
      const isSuccess = response && (response.success || response.id)
      const responseData = response.data || response

      if (isSuccess) {
        const createdRisk = {
          ...newRiskData, // Keep intended local data like assessments
          ...responseData,
          assessments: newRiskData.assessments, // Force keep nested assessments if backend drops it
          displayId: rawRisks.value.length + 1
        }
        rawRisks.value.push(createdRisk)
      }
    } catch (error) {
      console.error('Failed to add risk:', error)
    }
  }

  async function updateRisk(updatedRisk: any) {
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

      // Optimistic local update
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
      } catch (error) {
        console.error('Failed to update risk on backend (using local fallback):', error)
      }
    } catch (error) {
      console.error('Failed to update risk:', error)
    }
  }

  async function deleteRisk(id: string | number) {
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risks/${id}`, {
        method: 'DELETE'
      })
      if (response && response.success) {
        rawRisks.value = rawRisks.value.filter(r => String(r.id) !== String(id))
        updateDisplayIds()
      }
    } catch (error) {
      console.error('Failed to delete risk:', error)
    }
  }

  return {
    rawRisks,
    risks,
    branches: branchesList,
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
    fetchRisks
  }
})
