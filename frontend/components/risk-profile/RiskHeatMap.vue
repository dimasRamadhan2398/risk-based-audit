<template>
  <div class="heatmap-root">
    <!-- Main Heat Map Grid -->
    <div class="heatmap-container">
      <!-- Y-Axis Label (LIKELIHOOD) -->
      <div class="y-axis-label">
        <span>LIKELIHOOD</span>
      </div>

      <!-- Y-Axis Ticks -->
      <div class="y-axis">
        <div v-for="l in likelihoodLevels" :key="`y-${l}`" class="y-tick">
          <span class="tick-num">{{ l }}</span>
          <span class="tick-text">{{ likelihoodLabels[l] }}</span>
        </div>
      </div>

      <!-- Grid -->
      <div class="heatmap-grid">
        <div v-for="l in likelihoodLevels" :key="`row-${l}`" class="grid-row">
          <div
            v-for="i in impactLevels"
            :key="`cell-${l}-${i}`"
            class="grid-cell"
            :style="{ backgroundColor: getConfig(l, i).cellBg }"
            :class="{
              'grid-cell--dragover': dragOverCell === `${l}-${i}`
            }"
            @dragover.prevent="onDragOver($event, l, i)"
            @dragleave="onDragLeave"
            @drop="onDrop($event, l, i)"
          >
            <!-- Cell header: level label + score -->
            <div class="cell-header">
              <span class="cell-level-text">{{ getConfig(l, i).label }}</span>
              <span class="cell-score">{{ l * i }}</span>
            </div>

            <!-- Risk badges -->
            <div class="cell-badges">
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

      <!-- X-Axis Ticks -->
      <div class="x-axis">
        <div v-for="i in impactLevels" :key="`x-${i}`" class="x-tick">
          <span class="tick-num">{{ i }}</span>
          <span class="tick-text">{{ impactLabels[i] }}</span>
        </div>
      </div>

      <!-- X-Axis Label (IMPACT LEVEL) -->
      <div class="x-axis-label">
        <span>IMPACT LEVEL</span>
      </div>
    </div>

    <!-- Risk List Tabs -->
    <div class="mt-8 bg-[#1e2336]/40 dark:bg-gray-800/40 rounded-xl border border-gray-200 dark:border-gray-700/50 overflow-hidden">
      <UTabs :items="tabItems" class="w-full" :ui="{
        list: {
          background: 'bg-transparent dark:bg-transparent',
          rounded: 'rounded-none',
          padding: 'p-0',
          height: 'h-14',
          marker: {
            wrapper: 'absolute top-auto bottom-0 left-0 duration-200 ease-out focus:outline-none',
            base: 'w-full h-0.5 bg-primary-500 dark:bg-primary-400',
            rounded: 'rounded-none',
            shadow: 'shadow-none'
          },
          tab: {
            active: 'text-gray-900 dark:text-white',
            inactive: 'text-gray-500 dark:text-gray-400',
            height: 'h-14',
            rounded: 'rounded-none',
            font: 'font-medium'
          }
        }
      }">
        <template #default="{ item, selected }">
          <div class="flex items-center gap-2 relative truncate font-medium">
            <span>{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </div>
        </template>
        <template #priority="{ item }">
          <div class="p-4 space-y-3">
            <div v-for="risk in priorityRisks" :key="risk.id" 
              class="flex items-center justify-between p-4 bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-800 shadow-sm cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
              :style="{ borderLeftColor: getBadgeColor(risk.likelihood * risk.impact), borderLeftWidth: '4px' }"
              @click="openPreviewModal(risk)">
              
              <div class="flex items-center gap-4">
                <!-- Badge -->
                <div class="flex items-center justify-center min-w-[32px] h-8 px-2 rounded-full text-white font-bold text-xs" :style="{ backgroundColor: getBadgeColor(risk.likelihood * risk.impact) }">
                  {{ risk.id }}
                </div>
                
                <!-- Info -->
                <div class="flex flex-col">
                  <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ risk.name }}</h4>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 flex items-center gap-1.5 font-medium">
                    <span>{{ categoryIcons[risk.category] }}</span>
                    <span>{{ risk.category }}</span>
                    <span class="text-gray-400 dark:text-gray-600">&bull;</span>
                    <span>Severity: {{ risk.severity }}</span>
                  </p>
                </div>
              </div>
              
              <!-- Right side (Score & Level) -->
              <div class="text-right flex flex-col items-end justify-center">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">{{ getConfig(risk.likelihood, risk.impact).label }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white leading-none">{{ risk.likelihood * risk.impact }}</p>
              </div>
            </div>
            
            <div v-if="priorityRisks.length === 0" class="text-center py-8 text-gray-500 text-sm">
              Tidak ada risiko prioritas.
            </div>
          </div>
        </template>
        
        <template #all="{ item }">
          <div class="p-4 space-y-3">
            <div v-for="risk in sortedAllRisks" :key="risk.id" 
              class="flex items-center justify-between p-4 bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-800 shadow-sm cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
              :style="{ borderLeftColor: getBadgeColor(risk.likelihood * risk.impact), borderLeftWidth: '4px' }"
              @click="openPreviewModal(risk)">
              
              <div class="flex items-center gap-4">
                <!-- Badge -->
                <div class="flex items-center justify-center min-w-[32px] h-8 px-2 rounded-full text-white font-bold text-xs" :style="{ backgroundColor: getBadgeColor(risk.likelihood * risk.impact) }">
                  {{ risk.id }}
                </div>
                
                <!-- Info -->
                <div class="flex flex-col">
                  <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ risk.name }}</h4>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 flex items-center gap-1.5 font-medium">
                    <span>{{ categoryIcons[risk.category] }}</span>
                    <span>{{ risk.category }}</span>
                    <span class="text-gray-400 dark:text-gray-600">&bull;</span>
                    <span>Severity: {{ risk.severity }}</span>
                  </p>
                </div>
              </div>
              
              <!-- Right side (Score & Level) -->
              <div class="text-right flex flex-col items-end justify-center">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">{{ getConfig(risk.likelihood, risk.impact).label }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white leading-none">{{ risk.likelihood * risk.impact }}</p>
              </div>
            </div>
            
            <div v-if="sortedAllRisks.length === 0" class="text-center py-8 text-gray-500 text-sm">
              Tidak ada risiko.
            </div>
          </div>
        </template>
      </UTabs>
    </div>

    <!-- Risk Action Modal -->
    <RiskActionModal 
      v-model:open="isModalOpen" 
      :mode="modalMode" 
      :risk-data="selectedRisk"
      @close="isModalOpen = false"
      @submit="handleModalSubmit"
      @delete="deleteRisk"
    />
  </div>
</template>

<script setup>
import { ref, computed, defineExpose } from 'vue'

import RiskBadge from './RiskBadge.vue'
import RiskActionModal from './RiskActionModal.vue'
import {
  riskData,
  branches,
  getRiskLevel,
  impactLabels,
  likelihoodLabels,
  riskLevelConfig,
  categoryIcons
} from '~/stores/profile-risk'

const toast = useToast()

const props = defineProps({
  branch: {
    type: String,
    default: 'All Branches'
  }
})

// Reactive risk data
const risks = ref(riskData.map(r => ({ ...r })))
const dragOverCell = ref(null)
const draggingRisk = ref(null)

// UI State
const isModalOpen = ref(false)
const modalMode = ref('preview')
const selectedRisk = ref(null)

// Grid levels (reversed for Y-axis: 5 at top, 1 at bottom)
const likelihoodLevels = [5, 4, 3, 2, 1]
const impactLevels = [1, 2, 3, 4, 5]

// Safe config getter
function getConfig(likelihood, impact) {
  const level = getRiskLevel(likelihood, impact)
  return riskLevelConfig[level] || { label: 'UNKNOWN', color: '#000', bg: '#eee', cellBg: '#333', priority: false }
}

// Computed Filtered Data
const filteredRisks = computed(() => {
  if (props.branch === 'All Branches') return risks.value
  return risks.value.filter(r => r.branch === props.branch)
})

const totalRisks = computed(() => filteredRisks.value.length)

const priorityRisks = computed(() => {
  return filteredRisks.value
    .filter(r => getConfig(r.likelihood, r.impact).priority)
    .sort((a, b) => {
      const scoreA = a.likelihood * a.impact
      const scoreB = b.likelihood * b.impact
      if (scoreB !== scoreA) return scoreB - scoreA
      return b.severity - a.severity
    })
})

const priorityCount = computed(() => priorityRisks.value.length)

const sortedAllRisks = computed(() => {
  return [...filteredRisks.value].sort((a, b) => {
    const scoreA = a.likelihood * a.impact
    const scoreB = b.likelihood * b.impact
    if (scoreB !== scoreA) return scoreB - scoreA
    return b.severity - a.severity
  })
})

const tabItems = computed(() => [
  {
    slot: 'priority',
    label: `Risiko Prioritas (${priorityCount.value})`,
    icon: '🔥'
  },
  {
    slot: 'all',
    label: `Semua Risiko (${totalRisks.value})`,
    icon: '📋'
  }
])

function getBadgeColor(score) {
  if (score <= 4) return '#22c55e'       // green
  if (score <= 10) return '#84cc16'      // lime
  if (score <= 15) return '#eab308'      // yellow
  if (score <= 20) return '#f97316'      // orange
  return '#ef4444'                        // red
}

// Get risks for a specific cell
function getCellRisks(likelihood, impact) {
  return filteredRisks.value
    .filter(r => r.likelihood === likelihood && r.impact === impact)
    .sort((a, b) => b.severity - a.severity)
}

// Drag and Drop handlers
function onRiskDragStart(risk) {
  draggingRisk.value = risk
}

function onRiskDragEnd() {
  draggingRisk.value = null
  dragOverCell.value = null
}

function onDragOver(e, likelihood, impact) {
  e.dataTransfer.dropEffect = 'move'
  dragOverCell.value = `${likelihood}-${impact}`
}

function onDragLeave() {
  dragOverCell.value = null
}

function onDrop(e, newLikelihood, newImpact) {
  dragOverCell.value = null

  try {
    const riskDataStr = e.dataTransfer.getData('application/json')
    if (!riskDataStr) return

    const droppedRisk = JSON.parse(riskDataStr)
    const riskIndex = risks.value.findIndex(r => r.id === droppedRisk.id)
    if (riskIndex === -1) return

    const oldRisk = risks.value[riskIndex]

    // If dropped in the same cell, do nothing
    if (oldRisk.likelihood === newLikelihood && oldRisk.impact === newImpact) return

    risks.value[riskIndex] = {
      ...oldRisk,
      likelihood: newLikelihood,
      impact: newImpact
    }
  } catch (err) {
    console.error('Drop failed:', err)
  }
}

// Modal Handlers
function openAddModal() {
  selectedRisk.value = null
  modalMode.value = 'add'
  isModalOpen.value = true
}

function openEditModal(risk) {
  selectedRisk.value = risk
  modalMode.value = 'edit'
  isModalOpen.value = true
}

function openPreviewModal(risk) {
  selectedRisk.value = risk
  modalMode.value = 'preview'
  isModalOpen.value = true
}

function deleteRisk(id) {
  const idx = risks.value.findIndex(r => r.id === id)
  if (idx > -1) {
    const riskName = risks.value[idx].name
    risks.value.splice(idx, 1)
    toast.add({
      title: 'Data Dihapus',
      description: `Risiko "${riskName}" berhasil dihapus.`,
      icon: 'i-lucide-trash-2',
      color: 'red'
    })
  }
}

function handleModalSubmit(data, mode) {
  if (mode === 'add') {
    risks.value.push(data)
    toast.add({
      title: 'Risiko Ditambahkan',
      description: `Risiko baru "${data.name}" berhasil ditambahkan ke Heat Map.`,
      icon: 'i-lucide-check-circle',
      color: 'green'
    })
  } else if (mode === 'edit') {
    const idx = risks.value.findIndex(r => r.id === data.id)
    if (idx > -1) {
      risks.value[idx] = data
      toast.add({
        title: 'Data Diperbarui',
        description: `Risiko "${data.name}" berhasil diperbarui.`,
        icon: 'i-lucide-check-circle',
        color: 'green'
      })
    }
  }
}

defineExpose({
  openAddModal
})
</script>

<style scoped>
.heatmap-root {
  width: 100%;
}

/* ── Container layout ── */
.heatmap-container {
  display: grid;
  grid-template-columns: 28px 90px 1fr;
  grid-template-rows: 1fr auto auto;
  grid-template-areas:
    "ylabel yticks grid"
    ".      .      xticks"
    ".      .      xlabel";
  gap: 0;
  min-height: 520px;
}

/* ── Y-Axis Label (vertical text) ── */
.y-axis-label {
  grid-area: ylabel;
  display: flex;
  align-items: center;
  justify-content: center;
}

.y-axis-label span {
  writing-mode: vertical-rl;
  transform: rotate(180deg);
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.45);
}

/* ── Y-Axis Ticks ── */
.y-axis {
  grid-area: yticks;
  display: flex;
  flex-direction: column;
}

.y-tick {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding: 4px 0;
}

.tick-num {
  font-size: 1.25rem;
  font-weight: 800;
  color: rgba(255, 255, 255, 0.75);
  line-height: 1;
}

.tick-text {
  font-size: 0.65rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.45);
  text-align: center;
  line-height: 1.2;
}

/* ── Grid ── */
.heatmap-grid {
  grid-area: grid;
  display: flex;
  flex-direction: column;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 6px;
  overflow: hidden;
}

.grid-row {
  display: flex;
  flex: 1;
}

.grid-cell {
  flex: 1;
  position: relative;
  min-height: 90px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  flex-direction: column;
  transition: all 0.2s ease;
}

.grid-cell--dragover {
  outline: 2px solid rgba(99, 102, 241, 0.7);
  outline-offset: -2px;
  background-color: rgba(99, 102, 241, 0.15) !important;
}

/* ── Cell header (label + score) ── */
.cell-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 6px 8px 0;
}

.cell-level-text {
  font-size: 0.6rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.35);
  line-height: 1;
}

.cell-score {
  font-size: 0.65rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.25);
  line-height: 1;
}

/* ── Cell badges ── */
.cell-badges {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 4px 6px 6px;
}

/* ── X-Axis Ticks ── */
.x-axis {
  grid-area: xticks;
  display: flex;
  padding-top: 12px;
}

.x-tick {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

/* ── X-Axis Label ── */
.x-axis-label {
  grid-area: xlabel;
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 10px;
}

.x-axis-label span {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.45);
}
</style>