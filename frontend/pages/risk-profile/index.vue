<template>
  <UCard variant="soft">
    <template #header>
      <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 p-2">
        <!-- Title & Subtitle Section -->
        <div class="flex items-center gap-5">
          <div class="p-4 bg-primary-500/10 dark:bg-primary-500/20 text-primary-500 rounded-2xl shadow-lg border border-primary-500/20">
            <UIcon name="i-lucide-layout-grid" class="w-8 h-8" />
          </div>
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Corporate Risk Profile</h1>
            <p class="text-sm text-gray-500 dark:text-gray-400 font-medium">Interactive Risk Heat Map — Fiscal Year {{ currentYear }}</p>
          </div>
        </div>
        
        <!-- Stats & Actions Section -->
        <div class="flex flex-wrap items-center gap-4">
          <div class="flex items-center gap-3">
            <div class="stats-box">
              <span class="stats-value">{{ totalRisks }}</span>
              <span class="stats-label">TOTAL RISKS</span>
            </div>
            <div class="stats-box priority">
              <span class="stats-value text-orange-500">{{ priorityCount }}</span>
              <span class="stats-label text-orange-500/80">PRIORITY</span>
            </div>
          </div>
          
          <UButton 
            icon="i-lucide-plus" 
            color="primary" 
            size="xl" 
            class="px-6 rounded-xl shadow-lg shadow-primary-500/20 hover:scale-105 transition-transform"
            @click="onAddRisk"
          >
            Tambah Risiko
          </UButton>
        </div>
      </div>
    </template>

    <!-- Content Area: Filter, Hint, and Legend -->
    <div class="space-y-6 px-2 mb-8">
      <!-- Row 1: Hint & Filter -->
      <div class="flex flex-col lg:flex-row gap-4 items-stretch">
        <div class="flex-1 p-4 bg-gray-50 dark:bg-gray-800/40 rounded-xl border border-gray-200 dark:border-gray-700/50 flex items-start gap-3">
          <UIcon name="i-lucide-lightbulb" class="w-5 h-5 text-yellow-500 mt-0.5 shrink-0" />
          <p class="text-sm text-gray-600 dark:text-gray-300 leading-relaxed">
            <span class="font-bold text-primary-500">Hint:</span> Drag and drop risks to update their status. If multiple risks land in the same cell, they will automatically be prioritized and stacked based on their inherent severity weight.
          </p>
        </div>
        
        <div class="w-full lg:w-80 p-4 bg-gray-50 dark:bg-gray-800/40 rounded-xl border border-gray-200 dark:border-gray-700/50 flex flex-col justify-center gap-2">
          <label class="text-[10px] font-bold text-gray-500 uppercase tracking-widest px-1">Filter Branch/Dept:</label>
          <USelectMenu 
            v-model="selectedBranch" 
            :options="branchOptions" 
            class="w-full"
            size="md"
            variant="none"
            :ui="{ 
              base: 'bg-white dark:bg-gray-900 ring-1 ring-gray-200 dark:ring-gray-700 shadow-inner rounded-lg' 
            }"
          />
        </div>
      </div>

      <!-- Row 2: Legend -->
      <div class="p-4 bg-gray-50 dark:bg-gray-800/40 rounded-xl border border-gray-200 dark:border-gray-700/50">
        <div class="flex flex-wrap items-center gap-y-4 gap-x-8">
          <span class="text-[10px] font-bold text-gray-500 uppercase tracking-widest">Risk Levels</span>
          <div class="flex flex-wrap items-center gap-6">
            <div v-for="(config, key) in riskLevelConfig" :key="key" class="flex items-center gap-2 group">
              <span class="w-4 h-4 rounded shadow-sm border border-black/10 dark:border-white/10" :style="{ background: config.bg }"></span>
              <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">{{ config.label }}</span>
              <UBadge v-if="config.priority" color="primary" variant="soft" size="xs" class="ml-1 text-[9px] font-bold tracking-tighter uppercase px-1.5 py-0">Priority</UBadge>
            </div>
          </div>
        </div>
      </div>
    </div>

    <RiskHeatMap ref="heatMapRef" :branch="selectedBranch" />

  </UCard>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import RiskHeatMap from '~/components/risk-profile/RiskHeatMap.vue'
import { 
  riskData, 
  branches, 
  riskLevelConfig, 
  getRiskLevel,
  useRiskProfileStore 
} from '~/stores/profile-risk'

const currentYear = new Date().getFullYear()
const selectedBranch = ref('All Branches')
const branchOptions = computed(() => ['All Branches', ...branches])
const heatMapRef = ref<InstanceType<typeof RiskHeatMap> | null>(null)

const store = useRiskProfileStore()

// For the UI counts - showing summary of all data
const totalRisks = computed(() => riskData.length)
const priorityCount = computed(() => {
  return riskData.filter(r => {
    const level = getRiskLevel(r.likelihood, r.impact)
    return riskLevelConfig[level]?.priority || false
  }).length
})

const onAddRisk = () => {
  heatMapRef.value?.openAddModal()
}
</script>

<style scoped>
@reference "tailwindcss";

.stats-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 100px;
  height: 4rem;
  background-color: rgb(249 250 251);
  border-radius: 0.75rem;
  border-width: 1px;
  border-color: rgb(229 231 235);
  box-shadow: 0 1px 2px 0 rgb(0 0 0 / 0.05);
}

:is(.dark .stats-box) {
  background-color: rgb(31 41 55 / 0.6);
  border-color: rgb(55 65 81);
}

.stats-box.priority {
  border-color: rgb(249 115 22 / 0.3);
  background-color: rgb(249 115 22 / 0.05);
}

.stats-value {
  font-size: 1.25rem;
  font-weight: 900;
  letter-spacing: -0.025em;
  line-height: 1;
  margin-bottom: 0.25rem;
}

.stats-label {
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: rgb(107 114 128);
  text-transform: uppercase;
}
</style>
