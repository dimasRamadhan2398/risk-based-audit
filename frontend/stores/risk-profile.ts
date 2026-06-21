import { defineStore } from 'pinia'
import { ref } from 'vue'
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
    description: 'Terget pendapatan tidak tercapai karena kinerja tim marketing yang kurang maksimal dan strategi marketing yang tidak efektif.'
  },
  {
    id: 2,
    name: 'Target efisiensi biaya operasional dan umum tidak tercapai',
    impact: 5,
    likelihood: 4,
    severity: 95,
    category: 'Financial',
    branch: 'Head Office',
    description: 'Target efisiensi biaya operasional dan umum tidak tercapai karena kinerja tim keuangan yang kurang maksimal dan strategi keuangan yang tidak efektif.'
  },
  {
    id: 3,
    name: 'Regulatory Non-Compliance',
    impact: 4,
    likelihood: 4,
    severity: 88,
    category: 'Compliance',
    branch: 'Surabaya Branch',
    description: 'Failure to adhere to government regulations, industry standards, or legal requirements.'
  },
  {
    id: 4,
    name: 'Abuse of Power / Authority',
    impact: 5,
    likelihood: 5,
    severity: 92,
    category: 'Governance',
    branch: 'Head Office',
    description: 'Misuse of managerial or executive authority for personal gain or organizational harm.'
  },
  {
    id: 5,
    name: 'Supply Chain Disruption',
    impact: 4,
    likelihood: 3,
    severity: 72,
    category: 'Operations',
    branch: 'Bandung Branch',
    description: 'Critical interruptions in the supply chain due to vendor failures, logistics, or global events.'
  },
  {
    id: 6,
    name: 'Data Privacy Violation',
    impact: 5,
    likelihood: 3,
    severity: 85,
    category: 'Compliance',
    branch: 'Bali Branch',
    description: 'Breaches of customer or employee data privacy, violating GDPR/local data protection laws.'
  },
  {
    id: 7,
    name: 'Market Volatility Exposure',
    impact: 3,
    likelihood: 4,
    severity: 65,
    category: 'Financial',
    branch: 'Jakarta Branch',
    description: 'Financial losses due to unpredictable market fluctuations, currency risks, or commodity prices.'
  },
  {
    id: 8,
    name: 'Talent Attrition / Brain Drain',
    impact: 3,
    likelihood: 3,
    severity: 50,
    category: 'Human Resources',
    branch: 'Head Office',
    description: 'Loss of key employees and institutional knowledge affecting operational continuity.'
  },
  {
    id: 9,
    name: 'Reputational Damage',
    impact: 4,
    likelihood: 2,
    severity: 75,
    category: 'Strategic',
    branch: 'Surabaya Branch',
    description: 'Significant brand damage due to public scandals, social media crises, or product failures.'
  },
  {
    id: 10,
    name: 'Environmental Compliance Failure',
    impact: 3,
    likelihood: 2,
    severity: 55,
    category: 'Compliance',
    branch: 'Bandung Branch',
    description: 'Violations of environmental regulations leading to fines, shutdowns, or cleanup obligations.'
  },
  {
    id: 11,
    name: 'Operational System Failure',
    impact: 4,
    likelihood: 3,
    severity: 70,
    category: 'Technology',
    branch: 'Bali Branch',
    description: 'Critical failure in core business systems causing operational downtime and revenue loss.'
  },
  {
    id: 12,
    name: 'Insider Trading',
    impact: 5,
    likelihood: 2,
    severity: 90,
    category: 'Financial',
    branch: 'Head Office',
    description: 'Illegal trading of securities based on material, non-public information by employees.'
  },
  {
    id: 13,
    name: 'Workplace Safety Incident',
    impact: 3,
    likelihood: 2,
    severity: 58,
    category: 'Human Resources',
    branch: 'Surabaya Branch',
    description: 'Accidents or hazardous conditions leading to employee injury or regulatory action.'
  },
  {
    id: 14,
    name: 'Third-Party Vendor Risk',
    impact: 2,
    likelihood: 3,
    severity: 40,
    category: 'Operations',
    branch: 'Jakarta Branch',
    description: 'Risks arising from outsourced vendors failing to meet service, security, or compliance standards.'
  },
  {
    id: 15,
    name: 'Intellectual Property Theft',
    impact: 4,
    likelihood: 2,
    severity: 78,
    category: 'Strategic',
    branch: 'Bandung Branch',
    description: 'Unauthorized copying, use, or distribution of company trade secrets and proprietary technology.'
  },
  {
    id: 16,
    name: 'Natural Disaster Impact',
    impact: 5,
    likelihood: 1,
    severity: 60,
    category: 'Operations',
    branch: 'Bali Branch',
    description: 'Disruption from earthquakes, floods, hurricanes, or other catastrophic natural events.'
  },
  {
    id: 17,
    name: 'Interest Rate Fluctuation',
    impact: 2,
    likelihood: 4,
    severity: 35,
    category: 'Financial',
    branch: 'Head Office',
    description: 'Exposure to changing interest rates affecting debt servicing and investment returns.'
  },
  {
    id: 18,
    name: 'Political / Geopolitical Risk',
    impact: 3,
    likelihood: 3,
    severity: 52,
    category: 'Strategic',
    branch: 'Jakarta Branch',
    description: 'Business disruption from political instability, sanctions, trade wars, or regime changes.'
  },
  {
    id: 19,
    name: 'Product Liability',
    impact: 4,
    likelihood: 1,
    severity: 68,
    category: 'Operations',
    branch: 'Surabaya Branch',
    description: 'Legal liability from defective products causing harm to consumers or businesses.'
  },
  {
    id: 20,
    name: 'Pandemic / Health Crisis',
    impact: 5,
    likelihood: 2,
    severity: 82,
    category: 'Operations',
    branch: 'Head Office',
    description: 'Widespread health emergencies causing workforce disruption and operational shutdowns.'
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
  const risks = ref<any[]>([])
  const branchesList = ref(branches)

  // UI State
  const selectedBranch = ref('All Branches')
  const selectedRisk = ref(null)
  const isFormOpen = ref(false)
  const isDetailOpen = ref(false)
  const modalMode = ref('preview')

  const getRiskServiceBaseUrl = () => {
    return config.public.riskServiceBaseUrl || 'http://localhost:8004/api/v1'
  }

  // Load risks from backend
  const fetchRisks = async () => {
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risks`)
      if (response && response.success && Array.isArray(response.data) && response.data.length > 0) {
        risks.value = response.data.map((r: any, idx: number) => ({
          ...r,
          displayId: idx + 1
        }))
      } else {
        risks.value = initialRiskData.map((r: any, idx: number) => ({
          ...r,
          displayId: idx + 1
        }))
      }
    } catch (error) {
      console.error('Failed to fetch risks, falling back to mock data:', error)
      risks.value = initialRiskData.map((r: any, idx: number) => ({
        ...r,
        displayId: idx + 1
      }))
    }
  }

  // Fetch immediately
  fetchRisks()

  // Helper to re-map display IDs after changes
  const updateDisplayIds = () => {
    risks.value.forEach((r: any, idx: number) => {
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
      if (response && response.success) {
        const createdRisk = {
          ...response.data,
          displayId: risks.value.length + 1
        }
        risks.value.push(createdRisk)
      }
    } catch (error) {
      console.error('Failed to add risk:', error)
    }
  }

  async function updateRisk(updatedRisk: any) {
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/risks/${updatedRisk.id}`, {
        method: 'PUT',
        body: updatedRisk
      })
      if (response && response.success) {
        const idx = risks.value.findIndex(r => r.id === updatedRisk.id)
        if (idx !== -1) {
          const displayId = risks.value[idx].displayId
          risks.value[idx] = {
            ...response.data,
            displayId
          }
        }
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
        risks.value = risks.value.filter(r => r.id !== id)
        updateDisplayIds()
      }
    } catch (error) {
      console.error('Failed to delete risk:', error)
    }
  }

  return {
    risks,
    branches: branchesList,
    selectedBranch,
    selectedRisk,
    isFormOpen,
    isDetailOpen,
    modalMode,
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
