<template>
  <div class="space-y-8 p-6 max-w-full mx-auto">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-[var(--border-main)] pb-5">
      <div>
        <h1 class="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-white font-space">
          Audit Universe Configuration
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
          Manage your Corporate Audit Universe library and establish active entities for yearly planning.
        </p>
      </div>
      <div class="flex items-center gap-3">
        <UButton
          to="/risk-profile"
          icon="i-lucide-arrow-left"
          color="neutral"
          variant="outline"
        >
          Back to Heat Map
        </UButton>
      </div>
    </div>

    <!-- Alert / Toast -->
    <Transition name="fade">
      <UAlert
        v-if="alertMessage"
        :color="alertType === 'success' ? 'success' : 'error'"
        variant="solid"
        :title="alertType === 'success' ? 'Success' : 'Error'"
        :description="alertMessage"
        icon="i-lucide-info"
        class="shadow-md"
        closable
        @close="alertMessage = ''"
      />
    </Transition>

    <!-- Tabs Navigation -->
    <UTabs :items="tabItems" class="w-full">
      
      <!-- Tab 1: Corporate Universe Builder -->
      <template #library>
        <div class="flex flex-col gap-8 mt-6">
          <!-- Left Column: Standard Library Explorer -->
          <div class="space-y-6">
            <UCard class="shadow-sm border border-[var(--border-main)]">
              <template #header>
                <div>
                  <h2 class="text-lg font-bold text-slate-800 dark:text-slate-100 font-space">
                    Standard Audit Universe Library
                  </h2>
                  <p class="text-md text-slate-500 mt-0.5">
                    Select standard entities to include them in your Corporate Audit Universe.
                  </p>
                </div>
              </template>

              <!-- Standard Library Tree -->
              <div class="max-h-[500px] overflow-y-auto pr-2 space-y-4">
                <div 
                  v-for="node in standardUniverse" 
                  :key="node.id"
                  class="border border-slate-100 dark:border-slate-800 rounded-xl p-4 bg-slate-50/30 dark:bg-slate-900/10 space-y-3"
                >
                  <div class="flex items-center justify-between">
                    <div class="flex items-center gap-2">
                      <UCheckbox
                        :model-value="isStandardNodeSelected(node.id)"
                        @update:model-value="toggleStandardSelection(node)"
                      />
                      <span class="font-bold text-sm text-slate-800 dark:text-slate-200">
                        {{ node.name }}
                      </span>
                    </div>
                  </div>

                  <!-- Children (Sub-entities) as checkboxes -->
                  <div v-if="node.children && node.children.length > 0" class="pl-6 border-l border-slate-200 dark:border-slate-800 space-y-2">
                    <div 
                      v-for="sub in node.children" 
                      :key="sub.id"
                      class="flex items-center justify-between p-1 hover:bg-slate-50 dark:hover:bg-slate-800/40 rounded transition-colors"
                    >
                      <div class="flex items-center gap-2">
                        <UCheckbox
                          :model-value="isStandardNodeSelected(sub.id)"
                          @update:model-value="toggleStandardSelection(sub)"
                        />
                        <span class="text-md font-semibold text-slate-600 dark:text-slate-400">
                          {{ sub.name }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </UCard>
          </div>

          <!-- Right Column: Corporate Universe Explorer & Editor -->
          <div>
            <UCard class="shadow-sm border border-[var(--border-main)]">
              <template #header>
                <div class="flex items-center justify-between">
                  <h2 class="text-lg font-bold text-slate-800 dark:text-slate-100 font-space">
                    Corporate Audit Universe
                  </h2>
                  <div class="flex items-center gap-2">
                    <UButton
                      size="md"
                      color="primary"
                      icon="i-lucide-plus"
                      label="Add Custom Entity"
                      @click="openAddCustomModal(null)"
                    />
                    <UBadge color="primary" variant="subtle">{{ corporateUniverse.length }} Entity</UBadge>
                  </div>
                </div>
              </template>

              <!-- Corporate Universe Tree -->
              <div v-if="corporateUniverse.length > 0" class="max-h-[500px] overflow-y-auto pr-2 space-y-3">
                <div 
                  v-for="node in corporateUniverse" 
                  :key="node.id"
                  class="border border-slate-150 dark:border-slate-800/70 rounded-xl p-3 bg-white dark:bg-slate-900/40 space-y-2"
                >
                  <div class="flex items-center justify-between">
                    <span class="font-bold text-md text-slate-800 dark:text-slate-200">{{ node.name }}</span>
                    <div class="flex items-center gap-1">
                      <UTooltip text="Add Sub-Entity">
                        <UButton
                          icon="i-lucide-plus"
                          color="primary"
                          variant="ghost"
                          size="md"
                          @click="openAddCustomModal(node.id)"
                        />
                      </UTooltip>
                      <UTooltip text="Rename">
                        <UButton
                          icon="i-lucide-edit"
                          color="warning"
                          variant="ghost"
                          size="md"
                          @click="openRenameModal(node)"
                        />
                      </UTooltip>
                      <UTooltip text="Delete">
                        <UButton
                          icon="i-lucide-trash-2"
                          color="error"
                          variant="ghost"
                          size="md"
                          @click="deleteCorporateNode(node.id)"
                        />
                      </UTooltip>
                    </div>
                  </div>

                  <!-- Children (Sub-entities) -->
                  <div v-if="node.children && node.children.length > 0" class="pl-4 border-l border-slate-100 dark:border-slate-800 space-y-1">
                    <div 
                      v-for="sub in node.children" 
                      :key="sub.id"
                      class="flex items-center justify-between p-1 hover:bg-slate-50 dark:hover:bg-slate-800/30 rounded"
                    >
                      <span class="text-md text-slate-600 dark:text-slate-400">{{ sub.name }}</span>
                      <div class="flex items-center gap-1">
                        <UTooltip text="Rename">
                          <UButton
                            icon="i-lucide-edit"
                            color="warning"
                            variant="ghost"
                            size="md"
                            @click="openRenameModal(sub)"
                          />
                        </UTooltip>
                        <UTooltip text="Delete">
                          <UButton
                            icon="i-lucide-trash-2"
                            color="error"
                            variant="ghost"
                            size="md"
                            @click="deleteCorporateNode(sub.id)"
                          />
                        </UTooltip>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div v-else class="text-center py-16 text-slate-400 text-md">
                Corporate Audit Universe is empty. Select from standard library or add custom ones.
              </div>
            </UCard>
          </div>
        </div>

        <!-- Scoped Modals inside Tab 1 library template -->
        <UModal v-model:open="renameModalOpen">
          <template #content>
            <UCard>
              <template #header>
                <h3 class="font-bold text-base text-slate-800 dark:text-slate-100">Rename Entity</h3>
              </template>
              <div class="space-y-4">
                <UFormField label="Entity Name" class="space-y-2">
                  <UInput v-model="renameNodeName" placeholder="Enter entity name" color="neutral" class="w-full" />
                </UFormField>
              </div>
              <template #footer>
                <div class="flex justify-end gap-3">
                  <UButton color="neutral" variant="outline" label="Cancel" @click="() => { renameModalOpen = false }" />
                  <UButton color="primary" label="Save" @click="saveRenameNode" />
                </div>
              </template>
            </UCard>
          </template>
        </UModal>

        <UModal v-model:open="addCustomModalOpen">
          <template #content>
            <UCard>
              <template #header>
                <h3 class="font-bold text-base text-slate-800 dark:text-slate-100">
                  {{ addCustomParentID ? 'Add Sub-Entity' : 'Add Custom Corporate Entity' }}
                </h3>
              </template>
              <div class="space-y-4">
                <UFormField label="Name" class="space-y-1">
                  <UInput v-model="addCustomNodeName" placeholder="Enter name..." color="neutral" class="w-full" />
                </UFormField>
              </div>
              <template #footer>
                <div class="flex justify-end gap-3">
                  <UButton color="neutral" variant="outline" label="Cancel" @click="() => { addCustomModalOpen = false }" />
                  <UButton color="primary" label="Add Node" @click="saveAddCustomNode" />
                </div>
              </template>
            </UCard>
          </template>
        </UModal>
      </template>

      <!-- Tab 2: Establish Universe -->
      <template #establish>
        <div class="mt-6 space-y-6">
          <UCard class="shadow-sm border border-[var(--border-main)]">
            <template #header>
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                  <h2 class="text-lg font-bold text-slate-800 dark:text-slate-100 font-space flex items-center gap-2">
                    📅 2. Establish {{ selectedYear }} Audit Universe
                  </h2>
                  <p class="text-md text-slate-500 mt-0.5">
                    Select Auditable Entities from the Corporate Audit Universe active for auditing in year {{ selectedYear }}.
                  </p>
                </div>
                <div class="flex items-center gap-3">
                  <span class="text-sm font-semibold text-slate-700">Target Year:</span>
                  <USelect
                    v-model.number="selectedYear"
                    :items="[2025, 2026, 2027, 2028]"
                    size="sm"
                    color="neutral"
                    class="w-24"
                    @update:model-value="fetchYearlyUniverse"
                  />
                  <UButton
                    size="sm"
                    color="primary"
                    variant="solid"
                    icon="i-lucide-check-circle"
                    label="Establish Active Universe"
                    @click="saveYearlyEstablishment"
                  />
                </div>
              </div>
            </template>

            <!-- Grid of Corporate Entities -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6 max-h-[460px] overflow-y-auto pr-2">
              <div 
                v-for="node in corporateUniverse" 
                :key="node.id"
                class="border border-slate-100 dark:border-slate-800 rounded-xl p-4 bg-slate-50/20 dark:bg-slate-900/10 space-y-3"
              >
                <div class="flex items-center gap-2">
                  <UCheckbox
                    :model-value="isYearlySelected(node.id)"
                    @update:model-value="toggleYearlySelection(node.id)"
                  />
                  <span class="font-bold text-sm text-slate-800 dark:text-slate-200">
                    {{ node.name }}
                  </span>
                </div>

                <!-- Children / Sub-entities as checkboxes -->
                <div v-if="node.children && node.children.length > 0" class="pl-4 border-l border-slate-200 dark:border-slate-800 space-y-2">
                  <div v-for="sub in node.children" :key="sub.id" class="flex items-center gap-2">
                    <UCheckbox
                      :model-value="isYearlySelected(sub.id)"
                      @update:model-value="toggleYearlySelection(sub.id)"
                    />
                    <span class="text-md text-slate-500">{{ sub.name }}</span>
                  </div>
                </div>
              </div>
            </div>
          </UCard>
        </div>
      </template>

      <!-- Tab 3: Rekapitulasi & Priorities (Priority) -->
      <template #priority>
        <div class="mt-6 space-y-6">
          <UCard class="shadow-sm border border-[var(--border-main)]">
            <template #header>
              <div class="flex items-center justify-between">
                <div>
                  <h2 class="text-lg font-bold text-slate-800 dark:text-slate-100 font-space">
                    Audit Priority ({{ selectedYear }})
                  </h2>
                  <p class="text-md text-slate-500 mt-0.5">
                    Annual Audit Plan priorities based on calculated risk levels.
                  </p>
                </div>
                <div class="flex items-center gap-4">
                  <UBadge color="success" variant="subtle" class="font-bold">
                    {{ prioritizedCount }} Prioritized
                  </UBadge>
                  <UBadge color="info" variant="solid" class="font-bold">
                    Audit Priority = Risk Level Medium to High or High
                  </UBadge>
                </div>
              </div>
            </template>

            <!-- Recap Table -->
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-sm">
                <thead class="bg-slate-50 dark:bg-slate-850/50">
                  <tr>
                    <th scope="col" class="px-6 py-3 text-left font-semibold text-slate-700 dark:text-slate-300">No</th>
                    <th scope="col" class="px-6 py-3 text-left font-semibold text-slate-700 dark:text-slate-300">Auditable Entity</th>
                    <th scope="col" class="px-6 py-3 text-center scope font-semibold text-slate-700 dark:text-slate-300">Risk Index</th>
                    <th scope="col" class="px-6 py-3 text-center scope font-semibold text-slate-700 dark:text-slate-300">Risk Level</th>
                    <th scope="col" class="px-6 py-3 text-center scope font-semibold text-slate-700 dark:text-slate-300">Audit Priority*</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                  <tr 
                    v-for="(ent, idx) in sortedYearlyUniverse" 
                    :key="ent.id"
                    class="hover:bg-slate-50/50 dark:hover:bg-slate-800/10"
                    :class="ent.audit_priority ? 'bg-primary-50/10' : ''"
                  >
                    <td class="px-6 py-4 text-slate-500 font-medium">{{ idx + 1 }}</td>
                    <td class="px-6 py-4 font-bold text-slate-800 dark:text-slate-200">
                      {{ ent.corporate_audit_universe?.name }}
                    </td>
                    <td class="px-6 py-4 text-center font-semibold text-slate-700 dark:text-slate-300">
                      {{ ent.risk_index?.toFixed(1) }}%
                    </td>
                    <td class="px-6 py-4 text-center">
                      <UBadge :color="getRiskLevelBadgeColor(ent.risk_level)" size="md" class="font-bold">
                        {{ ent.risk_level || 'N/A' }}
                      </UBadge>
                    </td>
                    <td class="px-6 py-4 text-center">
                      <div v-if="ent.audit_priority" class="inline-flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400 font-bold">
                        <UIcon name="i-lucide-check-circle" class="w-5 h-5 text-emerald-500" />
                        <span>√ Priority</span>
                      </div>
                      <span v-else class="text-slate-400 text-md">-</span>
                    </td>
                  </tr>
                  <tr v-if="yearlyUniverse.length === 0">
                    <td colspan="5" class="text-center py-10 text-slate-400 text-md">
                      No established auditable entities for year {{ selectedYear }}.
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <template #footer>
              <div class="flex items-center justify-between text-md text-slate-400">
                <span>Sorted by Risk Index (descending)</span>
              </div>
            </template>
          </UCard>

          <!-- Risk Index Level Info -->
          <UCard class="shadow-sm border border-[var(--border-main)] bg-slate-50/50 dark:bg-slate-900/30">
            <template #header>
              <h3 class="text-sm font-bold text-slate-800 dark:text-slate-100">
                Corporate Risk Index Level Information
              </h3>
            </template>
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-md text-left">
                <thead class="bg-slate-100 dark:bg-slate-800">
                  <tr>
                    <th class="px-4 py-2 font-semibold text-slate-700 dark:text-slate-300">Risk Index</th>
                    <th class="px-4 py-2 font-semibold text-slate-700 dark:text-slate-300">Risk Level</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                  <tr>
                    <td class="px-4 py-2 text-slate-600 dark:text-slate-400 font-medium">80 - 100%</td>
                    <td class="px-4 py-2"><UBadge size="md" class="font-bold w-28 justify-center bg-red-500/100 dark:bg-red-500/100">High</UBadge></td>
                  </tr>
                  <tr>
                    <td class="px-4 py-2 text-slate-600 dark:text-slate-400 font-medium">60 - 79%</td>
                    <td class="px-4 py-2"><UBadge size="md" class="font-bold w-28 justify-center bg-orange-500/100 dark:bg-orange-500/100">Moderate to High</UBadge></td>
                  </tr>
                  <tr>
                    <td class="px-4 py-2 text-slate-600 dark:text-slate-400 font-medium">40 - 59%</td>
                    <td class="px-4 py-2"><UBadge size="md" class="font-bold w-28 justify-center bg-yellow-500/100 dark:bg-yellow-500/100">Moderate</UBadge></td>
                  </tr>
                  <tr>
                    <td class="px-4 py-2 text-slate-600 dark:text-slate-400 font-medium">20 - 39%</td>
                    <td class="px-4 py-2"><UBadge size="md" class="font-bold w-28 justify-center bg-lime-500/100 dark:bg-lime-500/100">Low to Moderate</UBadge></td>
                  </tr>
                  <tr>
                    <td class="px-4 py-2 text-slate-600 dark:text-slate-400 font-medium">0 - 19%</td>
                    <td class="px-4 py-2"><UBadge size="md" class="font-bold w-28 justify-center bg-green-500/100 dark:bg-green-500/100">Low</UBadge></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </UCard>

        </div>
      </template>
    </UTabs>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { useRiskFactorsStore } from '~/stores/risk-factors'

const store = useAuditUniverseStore()
const riskFactorsStore = useRiskFactorsStore()

const tabItems = [
  { slot: 'library', label: '1. Corporate Universe Builder' },
  { slot: 'establish', label: '2. Yearly Establishment' },
  { slot: 'priority', label: '3. Audit Priority' }
]

// State
const selectedYear = ref(2026)
const alertMessage = ref('')
const alertType = ref('success')

// Local state for Yearly selection checkboxes
const selectedYearlyIDs = ref<string[]>([])

// Modals
const renameModalOpen = ref(false)
const renameNode = ref<any>(null)
const renameNodeName = ref('')

const addCustomModalOpen = ref(false)
const addCustomParentID = ref<string | null>(null)
const addCustomNodeName = ref('')

// Lifecycle
onMounted(async () => {
  await store.fetchStandardUniverse()
  await store.fetchCorporateUniverse()
  await riskFactorsStore.fetchCorporateFactors()
  await fetchYearlyUniverse()
})

const standardUniverse = computed(() => store.standardUniverse)
const corporateUniverse = computed(() => store.corporateUniverse)
const yearlyUniverse = computed(() => store.yearlyUniverse)

const sortedYearlyUniverse = computed(() => {
  return [...store.yearlyUniverse]
    .sort((a, b) => (b.risk_index || 0) - (a.risk_index || 0))
})

const prioritizedCount = computed(() => {
  return store.yearlyUniverse.filter(ent => ent.audit_priority).length
})

// Methods
const fetchYearlyUniverse = async () => {
  await store.fetchYearlyUniverse(selectedYear.value)
  // Populate active selection local checkbox states
  selectedYearlyIDs.value = store.yearlyUniverse.map(e => e.corporate_audit_universe_id)
}

// Check if standard node is selected
const isStandardNodeSelected = (stdID: string): boolean => {
  const checkNode = (nodes: any[]): boolean => {
    for (const n of nodes) {
      if (n.standard_audit_universe_id === stdID) return true
      if (n.children && checkNode(n.children)) return true
    }
    return false
  }
  return checkNode(store.corporateUniverse)
}

const toggleStandardSelection = async (stdNode: any) => {
  const alreadySelected = isStandardNodeSelected(stdNode.id)
  
  if (alreadySelected) {
    const findCorpNode = (nodes: any[]): any => {
      for (const n of nodes) {
        if (n.standard_audit_universe_id === stdNode.id) return n
        if (n.children) {
          const res = findCorpNode(n.children)
          if (res) return res
        }
      }
      return null
    }
    const match = findCorpNode(store.corporateUniverse)
    if (match) {
      await deleteCorporateNode(match.id)
    }
  } else {
    let corpParentID: string | undefined = undefined
    if (stdNode.parent_id) {
      const findCorpParent = (nodes: any[]): any => {
        for (const n of nodes) {
          if (n.standard_audit_universe_id === stdNode.parent_id) return n
          if (n.children) {
            const res = findCorpParent(n.children)
            if (res) return res
          }
        }
        return null
      }
      const parentMatch = findCorpParent(store.corporateUniverse)
      if (parentMatch) {
        corpParentID = parentMatch.id
      }
    }

    await store.saveCorporateNode({
      name: stdNode.name,
      standard_audit_universe_id: stdNode.id,
      parent_id: corpParentID
    })
  }
}

// Renaming
const openRenameModal = (node: any) => {
  renameNode.value = node
  renameNodeName.value = node.name
  renameModalOpen.value = true
}

const saveRenameNode = async () => {
  if (!renameNodeName.value) return
  await store.saveCorporateNode({
    id: renameNode.value.id,
    name: renameNodeName.value,
    parent_id: renameNode.value.parent_id,
    standard_audit_universe_id: renameNode.value.standard_audit_universe_id
  })
  renameModalOpen.value = false
  showAlert('Corporate entity name updated successfully.', 'success')
}

// Adding custom nodes
const openAddCustomModal = (parentID: string | null) => {
  addCustomParentID.value = parentID
  addCustomNodeName.value = ''
  addCustomModalOpen.value = true
}

const saveAddCustomNode = async () => {
  if (!addCustomNodeName.value) return

  let corpParentID: string | undefined = undefined
  if (addCustomParentID.value) {
    const findCorpParent = (nodes: any[]): any => {
      for (const n of nodes) {
        if (n.standard_audit_universe_id === addCustomParentID.value || n.id === addCustomParentID.value) return n
        if (n.children) {
          const res = findCorpParent(n.children)
          if (res) return res
        }
      }
      return null
    }
    const match = findCorpParent(store.corporateUniverse)
    if (match) {
      corpParentID = match.id
    }
  }

  await store.saveCorporateNode({
    name: addCustomNodeName.value,
    parent_id: corpParentID
  })
  addCustomModalOpen.value = false
  showAlert('Custom corporate entity added successfully.', 'success')
}

const deleteCorporateNode = async (id: string) => {
  if (!confirm('Are you sure you want to delete this entity from Corporate Audit Universe? This will delete all sub-entities and yearly entries.')) return
  await store.deleteCorporateNode(id)
  showAlert('Entity deleted successfully from corporate library.', 'success')
  await fetchYearlyUniverse()
}

// Yearly establishment
const isYearlySelected = (id: string): boolean => {
  return selectedYearlyIDs.value.includes(id)
}

const toggleYearlySelection = (id: string) => {
  const idx = selectedYearlyIDs.value.indexOf(id)
  if (idx >= 0) {
    selectedYearlyIDs.value.splice(idx, 1)
  } else {
    selectedYearlyIDs.value.push(id)
  }
}

const saveYearlyEstablishment = async () => {
  const success = await store.establishYearlyUniverse(selectedYear.value, selectedYearlyIDs.value)
  if (success) {
    showAlert(`Successfully established active Audit Universe for year ${selectedYear.value}.`, 'success')
    await fetchYearlyUniverse()
  } else {
    showAlert(store.errorMsg || 'Failed to establish yearly audit universe.', 'error')
  }
}

const getRiskLevelBadgeColor = (level?: string) => {
  if (!level) return 'neutral'
  switch (level) {
    case 'High': return 'error'
    case 'Medium to High': return 'warning'
    case 'Medium': return 'primary'
    case 'Low to Medium': return 'info'
    case 'Low': return 'success'
    default: return 'neutral'
  }
}

// Alert helper
const showAlert = (msg: string, type: string) => {
  alertMessage.value = msg
  alertType.value = type
  setTimeout(() => {
    alertMessage.value = ''
  }, 4000)
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
