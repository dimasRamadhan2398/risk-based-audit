<template>
  <div class="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <!-- Header Card -->
    <UCard :ui="{ body: { padding: 'p-0' } }" class="mb-8 overflow-visible border-gray-200 dark:border-gray-800">
      <div class="p-6 flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div class="flex items-center gap-4">
          <div class="w-14 h-14 rounded-xl bg-gradient-to-br from-primary to-blue-400 flex items-center justify-center shadow-lg shadow-primary/20">
            <UIcon name="i-heroicons-chart-bar-square" class="w-8 h-8" />
          </div>
          <div>
            <h1 class="text-2xl font-extrabold tracking-tight">Corporate Risk Profile</h1>
            <p class="text-sm">Interactive Risk Heat Map — Fiscal Year {{ currentYear }}</p>
          </div>
        </div>
        
        <div class="flex items-center gap-4">
          <div class="flex gap-4">
            <div class="px-4 py-2 rounded-lg border border-gray-100 dark:border-gray-700 text-center min-w-[80px]">
              <span class="block text-xl font-black ">{{ totalRisks }}</span>
              <span class="block text-[10px] font-bold uppercase tracking-widest">Total</span>
            </div>
            <div class="px-4 py-2 bg-orange-100 dark:bg-orange-900/50 rounded-lg border border-orange-100 dark:border-orange-900/50 text-center min-w-[80px]">
              <span class="block text-xl text-warning-500 font-black ">{{ priorityCount }}</span>
              <span class="block text-[10px] font-bold text-warning-500 uppercase tracking-widest">Priority</span>
            </div>
          </div>
          
          <UModal 
            v-model:open="isAddModalOpen"
            title="Add New Corporate Risk" 
            description="Enter the details of the new risk to be added to the heatmap."
          >
            <UButton 
              icon="i-heroicons-plus"
              label="Add Risk"
              color="primary"
              size="lg"
              class="font-bold shadow-md shadow-primary/20"
            />

            <template #body>
              <form @submit.prevent="submitNewRisk" class="space-y-4">
                <UFormField label="Risk Name" required>
                  <UInput v-model="newRisk.name" placeholder="e.g. Data Breach, Financial Loss" class="w-full" />
                </UFormField>

                <UFormField label="Category">
                  <USelect v-model="newRisk.category" :items="categoryOptions" class="w-full" />
                </UFormField>

                <div class="grid grid-cols-2 gap-4">
                  <UFormField label="Impact Level">
                    <USelect v-model.number="newRisk.impact" :items="impactOptions" class="w-full" />
                  </UFormField>
                  <UFormField label="Likelihood Level">
                    <USelect v-model.number="newRisk.likelihood" :items="likelihoodOptions" class="w-full" />
                  </UFormField>
                </div>

                <UFormField :label="`Severity Weight: ${newRisk.severity}%`" help="Drag to adjust the risk priority weight">
                  <USlider v-model="newRisk.severity" :min="1" :max="100" color="primary" />
                </UFormField>

                <UFormField label="Risk Description">
                  <UTextarea v-model="newRisk.description" placeholder="Describe the risk and its potential consequences..." class="w-full" />
                </UFormField>
              </form>
            </template>

            <template #footer>
              <div class="flex justify-end gap-3">
                <UButton label="Cancel" color="neutral" variant="ghost" @click="isAddModalOpen = false" />
                <UButton label="Save Risk" color="primary" @click="submitNewRisk" />
              </div>
            </template>
          </UModal>
        </div>
      </div>
    </UCard>

    <!-- Controls & Hint -->
    <div class="grid md:grid-cols-3 gap-6 mb-8 items-start">
      <div class="md:col-span-2">
        <UAlert
          icon="i-heroicons-light-bulb"
          color="primary"
          variant="subtle"
          title="Quick Tip"
          description="Drag and drop risks to update their status. Multiple risks in the same cell are stacked by severity weight."
        />
      </div>
      <div>
        <UFormField label="Filter Branch/Dept" size="sm" class="font-bold">
          <USelect
            v-model="selectedBranch"
            :items="branchOptions"
            icon="i-heroicons-building-office"
            class="w-full"
          />
        </UFormField>
      </div>
    </div>

    <!-- Legend -->
    <div class="border border-gray-200 dark:border-gray-800 rounded-xl p-4 mb-10 flex flex-wrap items-center gap-x-8 gap-y-3">
      <span class="text-[10px] font-black uppercase tracking-[0.2em]">Risk Levels</span>
      <div class="flex flex-wrap gap-6">
        <div v-for="(config, key) in riskLevelConfig" :key="key" class="flex items-center gap-2">
          <div class="w-3.5 h-3.5 rounded-sm shadow-sm" :style="{ background: config.color }"></div>
          <span class="text-xs font-bold">{{ config.label }}</span>
          <UIcon v-if="config.priority" name="i-heroicons-fire" class="w-3.5 h-3.5 text-warning-500" />
        </div>
      </div>
    </div>

    <!-- Heat Map Grid -->
    <div class="relative pl-12 mb-16 select-none">
      <!-- Y-axis Label -->
      <div class="absolute -left-20 top-1/2 -translate-y-1/2 -rotate-90 origin-center whitespace-nowrap text-[10px] font-black uppercase tracking-[0.3em]">
        LIKELIHOOD LEVEL
      </div>

      <div class="flex gap-4">
        <!-- Y-axis Ticks -->
        <div class="flex flex-col w-28 shrink-0">
          <div v-for="l in likelihoodLevels" :key="`y-${l}`" class="flex-1 flex items-center justify-end gap-3 pr-2 min-h-[110px]">
            <span class="text-xs font-bold text-right leading-tight max-w-[70px] uppercase">{{ likelihoodLabels[l] }}</span>
            <span class="text-xl font-black">{{ l }}</span>
          </div>
        </div>

        <!-- Main Grid Area -->
        <div class="flex-1">
          <div class=" p-1 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700">
            <div class="grid grid-rows-5 gap-1 rounded-lg overflow-hidden border border-gray-200 dark:border-gray-700">
              <div v-for="l in likelihoodLevels" :key="`row-${l}`" class="grid grid-cols-5 gap-1">
                <div
                  v-for="i in impactLevels"
                  :key="`cell-${l}-${i}`"
                  :id="`cell-${l}-${i}`"
                  class="relative min-h-[110px] p-2 flex flex-col items-center justify-center transition-all duration-300 group"
                  :class="[
                    getCellBgClass(l, i),
                    dragOverCell === `${l}-${i}` ? 'ring-4 ring-primary ring-inset z-10 scale-[1.02] shadow-2xl' : ''
                  ]"
                  @dragover.prevent="onDragOver($event, l, i)"
                  @dragleave="onDragLeave"
                  @drop="onDrop($event, l, i)"
                >
                  <!-- Cell Labels -->
                  <div class="absolute inset-x-2 top-2 flex justify-between items-start pointer-events-none opacity-20 group-hover:opacity-40 transition-opacity">
                    <span class="text-[8px] font-black uppercase tracking-tighter max-w-[60%] leading-none">{{ riskLevelConfig[getRiskLevel(l, i)].label }}</span>
                    <span class="text-[10px] font-black">{{ getRiskScore(l, i) }}</span>
                  </div>

                  <!-- Badges Container -->
                  <div class="flex flex-wrap gap-1.5 justify-center items-center py-4">
                    <RiskBadge
                      v-for="(risk, idx) in getCellRisks(l, i)"
                      :key="risk.id"
                      :risk="risk"
                      :z-index="getCellRisks(l, i).length - idx"
                      @drag-start="onRiskDragStart"
                      @drag-end="onRiskDragEnd"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- X-axis Ticks -->
          <div class="flex mt-4 ml-0">
            <div v-for="i in impactLevels" :key="`x-${i}`" class="flex-1 flex flex-col items-center gap-1">
              <span class="text-xl font-black">{{ i }}</span>
              <span class="text-[10px] font-bold uppercase tracking-tighter text-center max-w-[80px] leading-tight">{{ impactLabels[i] }}</span>
            </div>
          </div>

          <!-- X-axis Label -->
          <div class="mt-8 text-center text-[10px] font-black uppercase tracking-[0.3em]">
            IMPACT LEVEL
          </div>
        </div>
      </div>
    </div>

    <!-- Risk List Panel -->
    <UCard :ui="{ body: { padding: 'p-0' } }" class="overflow-hidden border-gray-200 dark:border-gray-800">
      <UTabs v-model="activeTabIndex" :items="tabItems" class="w-full">
        <template #content="{ item }">
          <div class="p-6 max-h-[500px] overflow-y-auto">
            <TransitionGroup 
              name="list" 
              tag="div" 
              class="space-y-3"
            >
              <div
                v-for="risk in getTabRisks(item.key)"
                :key="risk.id"
                class="group flex items-center gap-4 p-4 rounded-xl border transition-all duration-200 hover:shadow-md"
                :class="getItemBorderClass(risk)"
              >
                <!-- Formatted ID Badge -->
                <div 
                  class="w-14 h-14 shrink-0 rounded-lg flex flex-col items-center justify-center shadow-lg font-mono"
                  :style="{ background: riskLevelConfig[getRiskLevel(risk.likelihood, risk.impact)].color }"
                >
                  <span class="text-[8px] font-bold opacity-70 uppercase">{{ getPrefix(risk) }}</span>
                  <span class="text-sm font-black">{{ getNumber(risk) }}</span>
                </div>

                <!-- Info -->
                <div class="flex-1 min-w-0">
                  <h4 class="text-sm font-bold truncate group-hover:text-primary transition-colors">
                    {{ risk.name }}
                  </h4>
                  <div class="flex items-center gap-3 mt-1">
                    <span class="text-[10px] font-bold flex items-center gap-1">
                      {{ categoryIcons[risk.category] }} {{ risk.category }}
                    </span>
                    <span>·</span>
                    <span class="text-[10px] font-bold">
                      Severity: {{ risk.severity }}
                    </span>
                  </div>
                </div>

                <!-- Score -->
                <div class="text-right shrink-0 px-4 border-r border-gray-100 dark:border-gray-800">
                  <div class="text-[10px] font-black uppercase tracking-tighter leading-none">
                    {{ riskLevelConfig[getRiskLevel(risk.likelihood, risk.impact)].label }}
                  </div>
                  <div class="text-2xl font-black leading-tight">
                    {{ getRiskScore(risk.likelihood, risk.impact) }}
                  </div>
                </div>

                <!-- Actions -->
                <UModal 
                  :title="`Risk Detail: # ${store.getFormattedId(risk)}`"
                  description="Complete assessment data and categorization for this risk event."
                  :ui="{ width: 'sm:max-w-lg' }"
                >
                  <UButton
                    icon="i-heroicons-eye"
                    color="neutral"
                    variant="ghost"
                    size="sm"
                  />

                  <template #body>
                    <div class="space-y-6">
                      <!-- Status Banner -->
                      <div class="flex items-center justify-between p-4 rounded-xl border border-gray-100 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-900/50">
                        <div class="flex items-center gap-3">
                          <span class="text-3xl">{{ categoryIcons[risk.category] }}</span>
                          <div>
                            <div class="text-[10px] font-black uppercase tracking-widest">Category</div>
                            <div class="text-sm font-bold">{{ risk.category }}</div>
                          </div>
                        </div>
                        <div class="text-right">
                          <div class="text-[10px] font-black uppercase tracking-widest mb-1">Risk Level</div>
                          <UBadge 
                            :style="{ backgroundColor: riskLevelConfig[getRiskLevel(risk.likelihood, risk.impact)].color, color: 'white' }"
                            size="sm"
                            class="font-black"
                          >
                            {{ riskLevelConfig[getRiskLevel(risk.likelihood, risk.impact)].label }}
                          </UBadge>
                        </div>
                      </div>

                      <!-- Basic Info -->
                      <div class="space-y-1">
                        <div class="text-[10px] font-black uppercase tracking-widest">Risk Event Name</div>
                        <h3 class="text-xl font-black leading-tight">{{ risk.name }}</h3>
                      </div>

                      <!-- Assessment Grid -->
                      <div class="grid grid-cols-2 gap-6 pt-4 border-t border-gray-100 dark:border-gray-800">
                        <div class="space-y-1.5">
                          <div class="flex items-center justify-between">
                            <span class="text-[10px] font-black uppercase tracking-widest">Impact</span>
                            <span class="text-xs font-bold">{{ risk.impact }}/5</span>
                          </div>
                          <div class="text-sm font-bold">{{ impactLabels[risk.impact] }}</div>
                        </div>
                        <div class="space-y-1.5">
                          <div class="flex items-center justify-between">
                            <span class="text-[10px] font-black uppercase tracking-widest">Likelihood</span>
                            <span class="text-xs font-bold">{{ risk.likelihood }}/5</span>
                          </div>
                          <div class="text-sm font-bold">{{ likelihoodLabels[risk.likelihood] }}</div>
                        </div>
                      </div>

                      <!-- Severity Progress -->
                      <div class="space-y-2 pt-4 border-t border-gray-100 dark:border-gray-800">
                        <div class="flex justify-between items-center">
                          <span class="text-[10px] font-black uppercase tracking-widest">Severity Weight</span>
                          <span class="text-sm font-black">{{ risk.severity }}%</span>
                        </div>
                        <UMeter :value="risk.severity" color="primary" size="md" />
                      </div>

                      <!-- Description -->
                      <div v-if="risk.description" class="space-y-1 pt-4 border-t border-gray-100 dark:border-gray-800">
                        <div class="text-[10px] font-black uppercase tracking-widest">Description</div>
                        <p class="text-sm leading-relaxed italic">
                          "{{ risk.description }}"
                        </p>
                      </div>

                      <!-- Mitigation Link -->
                      <div class="pt-6 border-t border-gray-100 dark:border-gray-800">
                        <UButton
                          icon="i-heroicons-shield-check"
                          label="Go to Mitigation Plan"
                          color="primary"
                          variant="soft"
                          class="w-full justify-center font-bold"
                          :to="`/mitigation?id=${risk.id}`"
                        />
                      </div>
                    </div>
                  </template>
                </UModal>

                <UModal 
                  :title="`Edit Risk: ${risk.name}`"
                  description="Modify the assessment parameters or categorization for this risk."
                >
                  <UButton
                    icon="i-heroicons-pencil-square"
                    color="warning"
                    variant="ghost"
                    size="sm"
                    @click="store.openEditModal(risk)"
                  />
                  <template #body>
                    <div v-if="store.selectedRisk && store.selectedRisk.id === risk.id" class="space-y-4">
                      <UFormField label="Risk Name" required>
                        <UInput v-model="store.selectedRisk.name" class="w-full" />
                      </UFormField>

                      <UFormField label="Category">
                        <USelect v-model="store.selectedRisk.category" :items="categoryOptions" class="w-full" />
                      </UFormField>

                      <div class="grid grid-cols-2 gap-4">
                        <UFormField label="Impact Level">
                          <USelect v-model.number="store.selectedRisk.impact" :items="impactOptions" class="w-full" />
                        </UFormField>
                        <UFormField label="Likelihood Level">
                          <USelect v-model.number="store.selectedRisk.likelihood" :items="likelihoodOptions" class="w-full" />
                        </UFormField>
                      </div>

                      <UFormField :label="`Severity Weight: ${store.selectedRisk.severity}%`" help="Update the risk priority weight">
                        <USlider v-model="store.selectedRisk.severity" :min="1" :max="100" color="warning" />
                      </UFormField>

                      <UFormField label="Risk Description">
                        <UTextarea v-model="store.selectedRisk.description" class="w-full" />
                      </UFormField>
                    </div>
                  </template>

                  <template #footer>
                    <div class="flex justify-end gap-3 w-full">
                      <UButton 
                        label="Cancel" 
                        color="neutral" 
                        variant="ghost" 
                        @click="(e) => e.target.closest('dialog')?.close()" 
                      />
                      <UButton 
                        label="Update Risk" 
                        color="warning" 
                        @click="(e) => { 
                          e.target.closest('dialog')?.close();
                          store.updateRisk(store.selectedRisk);
                          toast.add({ title: 'Risk Updated', description: 'Changes saved successfully.', color: 'success' });
                        }" 
                      />
                    </div>
                  </template>
                </UModal>
                
                  <UButton
                    icon="i-heroicons-trash"
                    color="error"
                    variant="ghost"
                    size="sm"
                    @click="handleDeleteRisk(risk.id)"
                  /> 
                </div>  
            </TransitionGroup>
            </div>
        </template>
      </UTabs>
    </UCard>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import RiskBadge from './RiskBadge.vue'
import { 
  useRiskProfileStore, 
  riskLevelConfig, 
  categoryIcons, 
  impactLabels, 
  likelihoodLabels 
} from '~/stores/risk-profile'

const store = useRiskProfileStore()
const { getRiskLevel, getRiskScore } = store
const toast = useToast()

// Local state for UI that doesn't need to be in the store (or mapping store state)
const currentYear = new Date().getFullYear()
const dragOverCell = ref(null)
const activeTabIndex = ref(0)
const isAddModalOpen = ref(false)

// Form state for new risk
const newRisk = ref({
  name: '',
  category: 'Strategic',
  impact: 3,
  likelihood: 3,
  severity: 50,
  description: '',
  branch: 'Head Office'
})

// Options for selects
const categoryOptions = Object.keys(categoryIcons).map(cat => ({ 
  label: `${categoryIcons[cat]} ${cat}`, 
  value: cat 
}))

const impactOptions = Object.entries(impactLabels).map(([val, label]) => ({ 
  label: `${val} - ${label}`, 
  value: Number(val) 
}))

const likelihoodOptions = Object.entries(likelihoodLabels).map(([val, label]) => ({ 
  label: `${val} - ${label}`, 
  value: Number(val) 
}))

function submitNewRisk() {
  if (!newRisk.value.name) {
    toast.add({
      title: 'Validation Error',
      description: 'Risk name is required.',
      color: 'error'
    })
    return
  }
  store.addRisk(newRisk.value)
  toast.add({
    title: 'Risk Added',
    description: `"${newRisk.value.name}" successfully added to the profile.`,
    color: 'success',
    icon: 'i-heroicons-check-circle'
  })
  
  // Reset form
  newRisk.value = {
    name: '',
    category: 'Strategic',
    impact: 3,
    likelihood: 3,
    severity: 50,
    description: '',
    branch: 'Head Office'
  }
  isAddModalOpen.value = false
}

// Grid levels
const likelihoodLevels = [5, 4, 3, 2, 1]
const impactLevels = [1, 2, 3, 4, 5]

// Map store state to local-like computed for template readability
const risks = computed(() => store.risks)
const selectedBranch = computed({
  get: () => store.selectedBranch,
  set: (val) => store.selectedBranch = val
})

const filteredRisks = computed(() => {
  if (selectedBranch.value === 'All Branches') return risks.value
  return risks.value.filter(r => r.branch === selectedBranch.value)
})

const totalRisks = computed(() => filteredRisks.value.length)

const priorityRisks = computed(() => {
  return filteredRisks.value
    .filter(r => {
      const level = store.getRiskLevel(r.likelihood, r.impact)
      return riskLevelConfig[level].priority
    })
    .sort((a, b) => {
      const scoreA = getRiskScore(a.likelihood, a.impact)
      const scoreB = getRiskScore(b.likelihood, b.impact)
      if (scoreB !== scoreA) return scoreB - scoreA
      return b.severity - a.severity
    })
})

const priorityCount = computed(() => priorityRisks.value.length)

const branchOptions = computed(() => ['All Branches', ...store.branches])

const tabItems = computed(() => [
  {
    key: 'priority',
    label: `Priority Risks (${priorityRisks.value.length})`,
    icon: 'i-heroicons-fire'
  },
  {
    key: 'all',
    label: `All Risks (${totalRisks.value})`,
    icon: 'i-heroicons-list-bullet'
  }
])

// Helpers
const getPrefix = (risk) => store.getFormattedId(risk).split('-')[0]
const getNumber = (risk) => store.getFormattedId(risk).split('-')[1]

function getCellRisks(likelihood, impact) {
  return filteredRisks.value
    .filter(r => r.likelihood === likelihood && r.impact === impact)
    .sort((a, b) => b.severity - a.severity)
}

function getTabRisks(key) {
  if (key === 'priority') return priorityRisks.value
  return [...filteredRisks.value].sort((a, b) => b.severity - a.severity)
}

const getCellBgClass = (l, i) => {
  const level = store.getRiskLevel(l, i)
  const map = {
    'low': 'bg-green-500/10 dark:bg-green-500/5 hover:bg-green-500/20',
    'low-moderate': 'bg-lime-500/15 dark:bg-lime-500/10 hover:bg-lime-500/25',
    'moderate': 'bg-yellow-500/20 dark:bg-yellow-500/15 hover:bg-yellow-500/30',
    'moderate-high': 'bg-orange-500/25 dark:bg-orange-500/20 hover:bg-orange-500/35',
    'high': 'bg-red-500/30 dark:bg-red-500/25 hover:bg-red-500/40'
  }
  return map[level]
}

const getItemBorderClass = (risk) => {
  const level = store.getRiskLevel(risk.likelihood, risk.impact)
  const map = {
    'low': 'border-green-500/20 bg-green-500/5',
    'low-moderate': 'border-lime-500/20 bg-lime-500/5',
    'moderate': 'border-yellow-500/20 bg-yellow-500/5',
    'moderate-high': 'border-orange-500/20 bg-orange-500/5',
    'high': 'border-red-500/20 bg-red-500/5'
  }
  return map[level]
}

// Event Handlers
function onRiskDragStart(risk) {
  // Store handles dragging state if needed, but local is fine for visual
}

function onRiskDragEnd() {
  dragOverCell.value = null
}

function onDragOver(e, l, i) {
  e.dataTransfer.dropEffect = 'move'
  dragOverCell.value = `${l}-${i}`
}

function onDragLeave() {
  dragOverCell.value = null
}

function onDrop(e, newLikelihood, newImpact) {
  dragOverCell.value = null
  try {
    const riskDataStr = e.dataTransfer?.getData('application/json')
    if (!riskDataStr) return

    const droppedRisk = JSON.parse(riskDataStr)
    if (droppedRisk.likelihood === newLikelihood && droppedRisk.impact === newImpact) return

    store.updateRisk({ ...droppedRisk, likelihood: newLikelihood, impact: newImpact })
    
    toast.add({
      title: 'Risk Position Updated',
      description: `"${droppedRisk.name}" moved to ${newLikelihood}×${newImpact}`,
      color: 'primary',
      icon: 'i-heroicons-arrows-right-left'
    })
  } catch (err) {
    console.error('Drop failed:', err)
  }
}

function handleDeleteRisk(id) {
  const risk = store.risks.find(r => r.id === id)
  if (!risk) return
  
  store.deleteRisk(id)
  toast.add({
    title: 'Risk Deleted',
    description: `"${risk.name}" has been removed.`,
    color: 'error',
    icon: 'i-heroicons-trash'
  })
}


</script>

<style scoped>
.list-enter-active { transition: all 0.3s ease-out; }
.list-leave-active { transition: all 0.2s ease-in; }
.list-enter-from { opacity: 0; transform: translateX(-20px); }
.list-leave-to { opacity: 0; transform: translateX(20px); }
.list-move { transition: transform 0.3s ease; }

.grid-cell:hover {
  z-index: 50;
}
</style>
