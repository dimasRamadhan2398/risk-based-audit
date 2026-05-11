<template>
  <UTooltip 
    class="max-w-xs p-0" 
    :popper="{ placement: 'top' }"
  >
    <div
      :id="`risk-badge-${risk.id}`"
      draggable="true"
      @dragstart="onDragStart"
      @dragend="onDragEnd"
      class="risk-badge relative flex items-center justify-center rounded-full text-xs font-bold cursor-grab active:cursor-grabbing transition-transform hover:scale-125 select-none"
      :class="[
        isPriority ? 'ring-2 ring-red-400 ring-offset-1 ring-offset-gray-900' : '',
        isDragging ? 'opacity-50 scale-95' : ''
      ]"
      :style="{ 
        backgroundColor: badgeColor, 
        color: '#fff', 
        zIndex: zIndex,
        minWidth: '34px',
        padding: typeof risk.id === 'string' && risk.id.length > 2 ? '0 8px' : '0',
        height: '34px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.4)'
      }"
    >
      <span>{{ risk.id }}</span>
      
      <span 
        v-if="isPriority" 
        class="absolute inset-0 rounded-full animate-ping opacity-75 pointer-events-none" 
        :style="{ backgroundColor: badgeColor }"
      ></span>
    </div>

    <template #content>
      <div class="p-3 space-y-3 w-64 text-sm">
        <div class="flex items-center gap-2 border-b border-gray-200 dark:border-gray-700 pb-2">
          <span class="text-lg">{{ categoryIcon }}</span>
          <span class="font-bold text-gray-900 dark:text-white">{{ risk.name }}</span>
        </div>

        <div class="space-y-1.5 text-xs">
          <div class="flex justify-between items-center">
            <span class="text-gray-500 dark:text-gray-400">Risk ID</span>
            <span class="font-medium text-gray-900 dark:text-white">#{{ risk.id }}</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="text-gray-500 dark:text-gray-400">Category</span>
            <span class="font-medium text-gray-900 dark:text-white">{{ risk.category }}</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="text-gray-500 dark:text-gray-400">Impact</span>
            <span class="font-medium text-gray-900 dark:text-white">{{ impactLabels[risk.impact] }} ({{ risk.impact }})</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="text-gray-500 dark:text-gray-400">Likelihood</span>
            <span class="font-medium text-gray-900 dark:text-white">{{ likelihoodLabels[risk.likelihood] }} ({{ risk.likelihood }})</span>
          </div>
          
          <div class="flex justify-between items-center">
            <span class="text-gray-500 dark:text-gray-400">Severity</span>
            <div class="flex items-center gap-2 w-24">
              <UMeter :value="risk.severity" :max="100" size="xs" color="orange" class="flex-1" />
              <span class="font-medium text-gray-900 dark:text-white">{{ risk.severity }}</span>
            </div>
          </div>

          <div class="flex justify-between items-center mt-2 pt-2 border-t border-gray-100 dark:border-gray-800">
            <span class="text-gray-500 dark:text-gray-400">Status</span>
            <UBadge :color="isPriority ? 'error' : 'success'" variant="subtle" size="xs">
              {{ isPriority ? '⚠ Priority Risk' : '✓ Monitored' }}
            </UBadge>
          </div>
        </div>

        <p class="text-xs text-gray-600 dark:text-gray-300 mt-2 bg-gray-50 dark:bg-gray-800/50 p-2 rounded-md">
          {{ risk.description || 'Tidak ada deskripsi tersedia.' }}
        </p>
      </div>
    </template>
  </UTooltip>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { getRiskLevel, riskLevelConfig, impactLabels, likelihoodLabels, categoryIcons } from '~/stores/profile-risk'

// Definisikan tipe untuk props jika menggunakan TypeScript
const props = defineProps({
  risk: {
    type: Object,
    required: true
  },
  zIndex: {
    type: Number,
    default: 1
  }
})

const emit = defineEmits(['drag-start', 'drag-end'])

// State untuk dragging, state showTooltip sudah dihapus karena ditangani otomatis oleh UTooltip
const isDragging = ref(false)

// Computed logic
const riskLevel = computed(() => getRiskLevel(props.risk.likelihood, props.risk.impact))
const config = computed(() => {
  return riskLevelConfig[riskLevel.value] || { label: 'Unknown', color: '#000', bg: '#eee', cellBg: '#333', priority: false }
})
const isPriority = computed(() => config.value?.priority || false)
const categoryIcon = computed(() => categoryIcons[props.risk.category] || '📌')

// Badge color: vivid circle colors for visibility on dark grid
const badgeColor = computed(() => {
  const score = props.risk.likelihood * props.risk.impact
  if (score <= 4) return '#22c55e'       // green
  if (score <= 10) return '#84cc16'      // lime
  if (score <= 15) return '#eab308'      // yellow
  if (score <= 20) return '#f97316'      // orange
  return '#ef4444'                        // red
})

// Drag Handlers
function onDragStart(e: DragEvent) {
  isDragging.value = true
  if (!e.dataTransfer) return

  e.dataTransfer.effectAllowed = 'move'
  e.dataTransfer.setData('application/json', JSON.stringify(props.risk))

  // Custom drag image logic
  const target = e.target as HTMLElement
  const el = target.cloneNode(true) as HTMLElement
  el.style.position = 'absolute'
  el.style.top = '-1000px'
  el.style.opacity = '0.9'
  el.style.transform = 'scale(1.2)'
  document.body.appendChild(el)
  
  e.dataTransfer.setDragImage(el, 18, 18)
  setTimeout(() => document.body.removeChild(el), 0)

  emit('drag-start', props.risk)
}

function onDragEnd() {
  isDragging.value = false
  emit('drag-end', props.risk)
}
</script>