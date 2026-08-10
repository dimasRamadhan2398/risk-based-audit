<template>
  <div class="p-6 max-w-7xl mx-auto space-y-8 bg-gray-50 dark:bg-gray-950 min-h-screen">
    <!-- Header -->
    <div class="flex justify-between items-center pb-4 border-b border-gray-200 dark:border-gray-800">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white uppercase tracking-tight">Risk Appetite Statement (RAS)</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">Manage organizational risk appetite statements and verify compliance with risk mitigation rules.</p>
      </div>
    </div>

    <!-- KPI Summary Section -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <UCard :ui="{ body: 'p-5' }" class="shadow-sm border border-gray-200 dark:border-gray-800 dark:bg-gray-900">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-md font-bold text-gray-500 dark:text-gray-400 uppercase">Compliance Level</p>
            <h3 class="text-2xl font-black mt-1" :class="complianceColor">
              {{ compliancePercentage }}%
            </h3>
          </div>
          <div class="p-3 rounded-full" :class="complianceBg">
            <UIcon :name="complianceIcon" class="w-6 h-6" />
          </div>
        </div>
        <p class="text-[10px] text-gray-400 mt-2 font-medium">
          {{ compliantCount }} of {{ totalRisks }} risks are appetite-compliant
        </p>
      </UCard>

      <UCard :ui="{ body: 'p-5' }" class="shadow-sm border border-gray-200 dark:border-gray-800 dark:bg-gray-900">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-md font-bold text-gray-500 dark:text-gray-400 uppercase">Total Risks</p>
            <h3 class="text-2xl font-black mt-1 text-gray-900 dark:text-white">
              {{ totalRisks }}
            </h3>
          </div>
          <div class="p-3 bg-blue-50 dark:bg-blue-900/20 text-blue-500 dark:text-blue-400 rounded-full">
            <UIcon name="i-heroicons-chart-bar" class="w-6 h-6" />
          </div>
        </div>
        <p class="text-[10px] text-gray-400 dark:text-gray-500 mt-2 font-medium">Active risks in current profile</p>
      </UCard>

      <UCard :ui="{ body: 'p-5' }" class="shadow-sm border border-gray-200 dark:border-gray-800 dark:bg-gray-900">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-md font-bold text-gray-500 dark:text-gray-400 uppercase">Acceptable Risks</p>
            <h3 class="text-2xl font-black mt-1 text-green-600 dark:text-green-500">
              {{ acceptableCount }}
            </h3>
          </div>
          <div class="p-3 bg-green-50 dark:bg-green-900/20 text-green-500 dark:text-green-400 rounded-full">
            <UIcon name="i-heroicons-check-circle" class="w-6 h-6" />
          </div>
        </div>
        <p class="text-[10px] text-gray-400 dark:text-gray-500 mt-2 font-medium">Low & Low-Moderate level risks</p>
      </UCard>

      <UCard :ui="{ body: 'p-5' }" class="shadow-sm border border-gray-200 dark:border-gray-800 dark:bg-gray-900">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-md font-bold text-gray-500 dark:text-gray-400 uppercase">Appetite Violations</p>
            <h3 class="text-2xl font-black mt-1" :class="violatingCount > 0 ? 'text-red-600 dark:text-red-500' : 'text-gray-900 dark:text-white'">
              {{ violatingCount }}
            </h3>
          </div>
          <div class="p-3 rounded-full" :class="violatingCount > 0 ? 'bg-red-50 dark:bg-red-900/20 text-red-500 dark:text-red-400' : 'bg-gray-50 dark:bg-gray-800 text-gray-400 dark:text-gray-500'">
            <UIcon name="i-heroicons-exclamation-triangle" class="w-6 h-6" :class="{ 'animate-bounce': violatingCount > 0 }" />
          </div>
        </div>
        <p class="text-[10px] text-gray-400 dark:text-gray-500 mt-2 font-medium">Moderate+ risks missing mitigations</p>
      </UCard>
    </div>

    <!-- Tabs Container -->
    <UCard :ui="{ body: 'p-0' }" class="overflow-hidden border border-gray-200 dark:border-gray-800">
      <UTabs v-model="activeTab" :items="tabs" class="w-full">
        <template #content="{ item }">
          <!-- Tab 1: Guidelines & Rules -->
          <div v-if="item.key === 'overview'" class="p-6 space-y-6">
            <div class="bg-blue-50/50 dark:bg-blue-900/20 border border-blue-100 dark:border-blue-900/50 rounded-xl p-5 space-y-3">
              <h3 class="text-base font-bold text-blue-900 dark:text-blue-100 flex items-center gap-2">
                <UIcon name="i-heroicons-information-circle" class="w-5 h-5" />
                Ketentuan Risk Appetite Statement (RAS)
              </h3>
              <p class="text-sm text-blue-800 dark:text-blue-200 leading-relaxed">
                Risk Appetite Statement menetapkan batas tingkat risiko yang dapat diterima (acceptable) oleh perusahaan untuk mencapai tujuan strategisnya, serta kewajiban tindakan perbaikan untuk risiko yang melampaui batas toleransi tersebut.
              </p>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <!-- Zone 1: Acceptable -->
              <div class="border border-green-200 dark:border-green-900/50 bg-green-50/20 dark:bg-green-900/20 rounded-xl p-6 space-y-4">
                <div class="flex items-center gap-3">
                  <div class="p-2 bg-green-100 dark:bg-green-900/50 text-green-600 dark:text-green-400 rounded-lg">
                    <UIcon name="i-heroicons-check-badge" class="w-6 h-6" />
                  </div>
                  <div>
                    <h4 class="font-extrabold text-green-900 dark:text-green-100">1. Acceptable Risk Zone</h4>
                    <p class="text-md text-green-700 dark:text-green-300">Risiko yang dapat diterima tanpa mitigasi tambahan</p>
                  </div>
                </div>

                <div class="space-y-2">
                  <div class="flex items-center gap-2 text-md font-bold text-gray-700 dark:text-gray-300">
                    <div class="w-3 h-3 rounded bg-[#4CAF50]"></div>
                    <span>Low Risk Level</span>
                  </div>
                  <div class="flex items-center gap-2 text-md font-bold text-gray-700 dark:text-gray-300">
                    <div class="w-3 h-3 rounded bg-[#8BC34A]"></div>
                    <span>Low to Moderate Risk Level</span>
                  </div>
                </div>

                <p class="text-md leading-normal text-green-800 dark:text-green-200 italic">
                  * Risiko pada area ini dinilai memiliki dampak dan kemungkinan yang kecil bagi operasional perusahaan, sehingga tidak memerlukan alokasi sumber daya mitigasi khusus.
                </p>
              </div>

              <!-- Zone 2: Mitigation Required -->
              <div class="border border-orange-200 dark:border-orange-900/50 bg-orange-50/20 dark:bg-orange-900/20 rounded-xl p-6 space-y-4">
                <div class="flex items-center gap-3">
                  <div class="p-2 bg-orange-100 dark:bg-orange-900/50 text-orange-600 dark:text-orange-400 rounded-lg">
                    <UIcon name="i-heroicons-shield-exclamation" class="w-6 h-6" />
                  </div>
                  <div>
                    <h4 class="font-extrabold text-orange-950 dark:text-orange-100">2. Action Required Zone</h4>
                    <p class="text-md text-orange-800 dark:text-orange-300">Risiko wajib dimitigasi demi meminimalkan paparan</p>
                  </div>
                </div>

                <div class="space-y-2">
                  <div class="flex items-center gap-2 text-md font-bold text-gray-700 dark:text-gray-300">
                    <div class="w-3 h-3 rounded bg-[#FFC107]"></div>
                    <span>Moderate Risk Level</span>
                  </div>
                  <div class="flex items-center gap-2 text-md font-bold text-gray-700 dark:text-gray-300">
                    <div class="w-3 h-3 rounded bg-[#FF9800]"></div>
                    <span>Moderate to High Risk Level</span>
                  </div>
                  <div class="flex items-center gap-2 text-md font-bold text-gray-700 dark:text-gray-300">
                    <div class="w-3 h-3 rounded bg-[#F44336]"></div>
                    <span>High Risk Level</span>
                  </div>
                </div>

                <p class="text-md leading-normal text-orange-900 dark:text-orange-200 italic">
                  * Risiko wajib ditangani secara aktif dengan menyusun rencana mitigasi yang terstruktur dan terukur guna menurunkan level eksposur ke batas yang aman.
                </p>
              </div>
            </div>
          </div>

          <!-- Tab 2: Compliance Validation -->
          <div v-else-if="item.key === 'compliance'" class="p-6 space-y-6">
            <!-- Filter & View Controls -->
            <div class="flex justify-between items-center pb-4 border-b border-gray-100 dark:border-gray-800">
              <div>
                <h3 class="text-lg font-bold text-gray-900 dark:text-white">Risk Profile Compliance Registry</h3>
                <p class="text-md text-gray-500 dark:text-gray-400">Checking current Risk Profile against the appetite statement rules</p>
              </div>
              <div class="flex gap-3">
                <USelect v-model="complianceFilter" :items="['All Risks', 'Mitigation Required (Moderate+)', 'Acceptable (Low/Low-Mod)', 'Non-Compliant (Violating)']" />
              </div>
            </div>

            <!-- Table -->
            <div class="overflow-x-auto border border-gray-200 dark:border-gray-800 rounded-xl bg-white dark:bg-gray-900">
              <UTable :data="filteredComplianceRisks" :columns="complianceColumns">
                <template #id-cell="{ row }">
                  <span class="font-mono text-md font-bold text-gray-500 dark:text-gray-400">
                    {{ profileStore.getFormattedId(row.original) }}
                  </span>
                </template>

                <template #name-cell="{ row }">
                  <div class="max-w-md whitespace-normal">
                    <p class="font-bold text-gray-900 dark:text-white text-md">{{ row.original.name }}</p>
                    <p class="text-[10px] text-gray-400 dark:text-gray-500 mt-0.5">{{ row.original.category }} · {{ row.original.branch || 'Head Office' }}</p>
                  </div>
                </template>

                <template #level-cell="{ row }">
                  <span 
                    class="inline-block px-2 py-0.5 rounded text-[10px] font-black text-white"
                    :style="{ backgroundColor: getRiskLevelColor(row.original) }"
                  >
                    {{ getRiskLevelLabel(row.original) }}
                  </span>
                </template>

                <template #appetite-cell="{ row }">
                  <UBadge 
                    :color="isAppetiteAcceptable(row.original) ? 'success' : 'warning'" 
                    variant="subtle"
                    class="font-black text-[9px]"
                  >
                    {{ isAppetiteAcceptable(row.original) ? 'Acceptable' : 'Mitigation Required' }}
                  </UBadge>
                </template>

                <template #mitigationStatus-cell="{ row }">
                  <div class="flex items-center gap-2">
                    <template v-if="hasMitigation(row.original.id)">
                      <UBadge v-if="isMitigationSelesai(row.original.id)" color="success" variant="solid" class="font-black text-[9px]">
                        Selesai
                      </UBadge>
                      <UBadge v-else color="success" variant="solid" class="font-black text-[9px]">
                        Active ({{ getMitigationCount(row.original.id) }} Plan)
                      </UBadge>
                    </template>
                    <template v-else-if="isAppetiteAcceptable(row.original)">
                      <span class="text-md text-gray-400 dark:text-gray-500 italic">None Required</span>
                    </template>
                    <template v-else>
                      <div class="flex items-center gap-1.5 text-red-600 dark:text-red-500 animate-pulse">
                        <UIcon name="i-heroicons-exclamation-triangle" class="w-4 h-4" />
                        <span class="text-[10px] font-bold uppercase tracking-wider">Missing Plan</span>
                      </div>
                    </template>
                  </div>
                </template>

                <template #actions-cell="{ row }">
                  <div class="flex items-center gap-2">
                    <UButton 
                      v-if="!isAppetiteAcceptable(row.original)"
                      :icon="hasMitigation(row.original.id) ? 'i-heroicons-eye' : 'i-heroicons-plus-circle'" 
                      :label="hasMitigation(row.original.id) ? 'View Mitigation' : 'Add Mitigation'" 
                      :color="hasMitigation(row.original.id) ? 'primary' : 'error'" 
                      variant="soft" 
                      size="md" 
                      class="font-bold"
                      :to="`/mitigation?id=${row.original.id}`" 
                    />
                    <span v-else class="text-md text-gray-400 dark:text-gray-500">-</span>
                  </div>
                </template>
              </UTable>
            </div>
          </div>

          <!-- Tab 3: Statements Registry (CRUD) -->
          <div v-else-if="item.key === 'statements'" class="p-6 space-y-6">
            <div class="flex justify-between items-center pb-4 border-b border-gray-100 dark:border-gray-800">
              <div>
                <h3 class="text-lg font-bold text-gray-900 dark:text-white">Appetite Statements Management</h3>
                <p class="text-md text-gray-500 dark:text-gray-400">Official written policy declarations and threshold configurations</p>
              </div>
              <UButton 
                label="New Statement" 
                icon="i-heroicons-plus" 
                color="warning" 
                class="font-bold shadow-md shadow-orange-500/10"
                @click="openAddModal" 
              />
            </div>

            <!-- Empty State -->
            <div v-if="appetiteStore.statements.length === 0" class="py-16 text-center space-y-4 bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800">
              <div class="inline-flex p-4 bg-gray-100 dark:bg-gray-800 rounded-full text-gray-400 dark:text-gray-500">
                <UIcon name="i-heroicons-clipboard-document-list" class="w-12 h-12" />
              </div>
              <div class="max-w-md mx-auto">
                <h3 class="text-sm font-bold text-gray-900 dark:text-white">No appetite statements</h3>
                <p class="text-md text-gray-500 dark:text-gray-400 mt-1">Create official policy declarations to establish limits.</p>
              </div>
            </div>

            <!-- List Grid -->
            <div v-else class="grid grid-cols-1 gap-4">
              <UCard 
                v-for="stmt in appetiteStore.statements" 
                :key="stmt.id" 
                class="border border-gray-200 dark:border-gray-800 dark:bg-gray-900 shadow-sm hover:shadow transition-shadow"
                :ui="{ body: 'p-5' }"
              >
                <div class="flex justify-between items-start gap-4">
                  <div class="space-y-2 flex-1">
                    <div class="flex items-center gap-3">
                      <UBadge 
                        :color="stmt.status === 'APPROVED' ? 'success' : stmt.status === 'SUBMITTED' ? 'warning' : 'neutral'" 
                        size="sm" 
                        variant="soft"
                        class="font-bold"
                      >
                        {{ stmt.status }}
                      </UBadge>
                      <span class="text-[10px] font-mono text-gray-400 dark:text-gray-500">Limit Threshold: {{ stmt.threshold_limit }}%</span>
                    </div>
                    <p class="text-sm font-bold text-gray-800 dark:text-gray-200 leading-relaxed">
                      "{{ stmt.statement }}"
                    </p>
                  </div>

                  <div class="flex gap-1.5 shrink-0">
                    <UButton 
                      icon="i-heroicons-pencil-square" 
                      color="primary" 
                      variant="ghost" 
                      size="sm" 
                      @click="openEditModal(stmt)" 
                    />
                    <UButton 
                      icon="i-heroicons-trash" 
                      color="error" 
                      variant="ghost" 
                      size="sm" 
                      @click="handleDelete(stmt.id)" 
                    />
                  </div>
                </div>
              </UCard>
            </div>
          </div>
        </template>
      </UTabs>
    </UCard>

    <!-- CRUD Form Modal -->
    <UModal v-model:open="isModalOpen" :title="isEditing ? 'Edit Appetite Statement' : 'New Appetite Statement'" :ui="{ content: 'sm:max-w-lg' }">
      <template #body>
        <form @submit.prevent="handleSubmit" class="space-y-4">
          <UFormField label="Statement Policy Text" required>
            <UTextarea 
              v-model="form.statement" 
              placeholder="e.g. Risiko tingkat Low dan Low to Moderate dapat diterima..." 
              :rows="4" 
              class="w-full" 
              required 
            />
          </UFormField>

          <div class="grid grid-cols-2 gap-4">
            <UFormField label="Threshold Limit (%)" required>
              <UInput 
                v-model.number="form.threshold_limit" 
                type="number" 
                step="0.01" 
                placeholder="10.00" 
                class="w-full" 
                required 
              />
            </UFormField>

            <UFormField label="Approval Status" required>
              <USelect 
                v-model="form.status" 
                :items="['DRAFT', 'SUBMITTED', 'APPROVED']" 
                class="w-full" 
                required 
              />
            </UFormField>
          </div>
        </form>
      </template>

      <template #footer>
        <div class="flex justify-end gap-3">
          <UButton label="Cancel" color="neutral" variant="ghost" @click="() => { isModalOpen = false }" />
          <UButton 
            :label="isEditing ? 'Save Changes' : 'Create'" 
            color="warning" 
            class="font-bold"
            :loading="appetiteStore.loading"
            @click="handleSubmit()" 
          />
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRiskProfileStore, riskLevelConfig } from '~/stores/risk-profile'
import { useMitigationStore } from '~/stores/mitigation-risk'
import { useRiskAppetiteStore, type RiskAppetite } from '~/stores/risk-appetite'
import { RiskLevel } from '~/types/risk'

const profileStore = useRiskProfileStore()
const mitigationStore = useMitigationStore()
const appetiteStore = useRiskAppetiteStore()

const activeTab = ref('overview')
const complianceFilter = ref('All Risks')

const isModalOpen = ref(false)
const isEditing = ref(false)
const editingId = ref('')

const form = ref({
  statement: '',
  threshold_limit: 10.00,
  status: 'DRAFT'
})

const tabs = [
  {label: 'RAS Guidelines', key: 'overview' as const, value: 'overview', icon: 'i-heroicons-information-circle'},
  {label: 'Risk Compliance', key: 'compliance' as const, value: 'compliance', icon: 'i-heroicons-shield-check'},
  {label: 'Appetite Statements', key: 'statements' as const, value: 'statements', icon: 'i-heroicons-clipboard-document-list'}
]

const complianceColumns = [
  { accessorKey: 'id', header: 'Risk ID' },
  { accessorKey: 'name', header: 'Risk Event' },
  { accessorKey: 'level', header: 'Risk Level' },
  { accessorKey: 'appetite', header: 'Appetite Rule' },
  { accessorKey: 'mitigationStatus', header: 'Mitigation Status' },
  { accessorKey: 'actions', header: 'Actions' }
]

onMounted(async () => {
  await Promise.all([
    profileStore.fetchRisks(),
    mitigationStore.fetchMitigations(),
    appetiteStore.fetchStatements()
  ])
})

// Helpers
const getRiskLevelColor = (risk: any) => {
  const level = profileStore.getRiskLevel(risk.likelihood, risk.impact)
  return riskLevelConfig[level]?.color || '#9E9E9E'
}

const getRiskLevelLabel = (risk: any) => {
  const level = profileStore.getRiskLevel(risk.likelihood, risk.impact)
  return riskLevelConfig[level]?.label || 'Unknown'
}

const isAppetiteAcceptable = (risk: any): boolean => {
  const level = profileStore.getRiskLevel(risk.likelihood, risk.impact)
  return level === RiskLevel.LOW || level === RiskLevel.LOW_MODERATE
}

const hasMitigation = (riskId: string): boolean => {
  return mitigationStore.mitigations.some(m => m.riskId === String(riskId) || m.riskId === String(riskId))
}

const getMitigationCount = (riskId: string): number => {
  return mitigationStore.mitigations.filter(m => m.riskId === String(riskId)).length
}

const isMitigationSelesai = (riskId: string): boolean => {
  const mits = mitigationStore.mitigations.filter(m => m.riskId === String(riskId))
  if (mits.length === 0) return false
  
  let allSelesai = true
  for (const row of mits) {
    if (!row.monitoring) {
      allSelesai = false
      break
    }
    
    let target = 0
    let actual = 0
    const now = new Date()
    row.monitoring.forEach((m: any) => {
      if (m.weeks) {
        target += m.weeks.filter((check: any) => new Date(check.startDate) <= now).length
        actual += m.weeks.filter((check: any) => check.checked).length
      }
    })
    
    if (target === 0 || actual < target) {
      allSelesai = false
      break
    }
  }
  return allSelesai
}

// KPI Computations
const totalRisks = computed(() => profileStore.risks.length)

const acceptableCount = computed(() => {
  return profileStore.risks.filter(r => isAppetiteAcceptable(r)).length
})

const compliantCount = computed(() => {
  return profileStore.risks.filter(r => {
    if (isAppetiteAcceptable(r)) return true
    return hasMitigation(r.id)
  }).length
})

const violatingCount = computed(() => {
  return profileStore.risks.filter(r => {
    if (isAppetiteAcceptable(r)) return false
    return !hasMitigation(r.id)
  }).length
})

const compliancePercentage = computed(() => {
  if (totalRisks.value === 0) return 0
  return Math.round((compliantCount.value / totalRisks.value) * 100)
})

// Compliance Colors
const complianceColor = computed(() => {
  const p = compliancePercentage.value
  if (p >= 90) return 'text-green-600'
  if (p >= 70) return 'text-amber-500'
  return 'text-red-600'
})

const complianceBg = computed(() => {
  const p = compliancePercentage.value
  if (p >= 90) return 'bg-green-50 text-green-500'
  if (p >= 70) return 'bg-amber-50 text-amber-500'
  return 'bg-red-50 text-red-500'
})

const complianceIcon = computed(() => {
  const p = compliancePercentage.value
  if (p >= 90) return 'i-heroicons-shield-check'
  return 'i-heroicons-shield-exclamation'
})

// Filtered Compliance Risks
const filteredComplianceRisks = computed(() => {
  return profileStore.risks.filter(r => {
    if (complianceFilter.value === 'All Risks') return true
    
    const acceptable = isAppetiteAcceptable(r)
    if (complianceFilter.value === 'Acceptable (Low/Low-Mod)') return acceptable
    if (complianceFilter.value === 'Mitigation Required (Moderate+)') return !acceptable
    
    // Non-Compliant
    if (complianceFilter.value === 'Non-Compliant (Violating)') {
      return !acceptable && !hasMitigation(r.id)
    }
    return true
  })
})

// CRUD actions
const openAddModal = () => {
  isEditing.value = false
  form.value = {
    statement: '',
    threshold_limit: 10.00,
    status: 'DRAFT'
  }
  isModalOpen.value = true
}

const openEditModal = (stmt: RiskAppetite) => {
  isEditing.value = true
  editingId.value = stmt.id
  form.value = {
    statement: stmt.statement,
    threshold_limit: stmt.threshold_limit,
    status: stmt.status
  }
  isModalOpen.value = true
}

const handleSubmit = async () => {
  try {
    if (isEditing.value) {
      await appetiteStore.updateStatement(editingId.value, form.value)
    } else {
      await appetiteStore.createStatement(form.value)
    }
    isModalOpen.value = false
  } catch (error) {
    alert('Failed to save appetite statement.')
  }
}

const handleDelete = async (id: string) => {
  if (confirm('Are you sure you want to delete this appetite statement?')) {
    try {
      await appetiteStore.deleteStatement(id)
    } catch (error) {
      alert('Failed to delete statement.')
    }
  }
}
</script>
