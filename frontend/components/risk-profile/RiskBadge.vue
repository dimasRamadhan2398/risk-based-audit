<template>
  <UTooltip 
    :ui="{ content: 'max-w-[320px] p-0' }"
    :popper="{ placement: 'top' }"
  >
    <template #content>
      <div class="p-4 space-y-4 rounded-lg shadow-xl bg-neutral-50 border border-primary-900">
        <div class="flex items-center gap-3 pb-3 border-b border-primary-900">
          <span class="text-xl">{{ categoryIcon }}</span>
          <div class="flex-1 min-w-0">
            <h4 class="text-sm font-bold truncate">
              {{ risk.name }}
            </h4>
            <div class="text-[10px] font-mono font-bold uppercase tracking-wider">
              #{{ formattedId }}
            </div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-y-3 gap-x-4">
          <div class="space-y-1">
            <div class="text-[10px] font-bold uppercase tracking-wider">Category</div>
            <div class="text-xs font-medium">{{ risk.category }}</div>
          </div>
          <div class="space-y-1">
            <div class="text-[10px] font-bold uppercase tracking-wider">Status</div>
            <div>
              <UBadge 
                :color="isPriority ? 'error' : 'success'" 
                variant="subtle" 
                size="sm"
                class="font-bold"
              >
                {{ isPriority ? 'Priority' : 'Monitored' }}
              </UBadge>
            </div>
          </div>
          <div class="space-y-1">
            <div class="text-[10px] font-bold uppercase tracking-wider">Impact</div>
            <div class="text-xs">{{ impactLabels[risk.impact] }} ({{ risk.impact }})</div>
          </div>
          <div class="space-y-1">
            <div class="text-[10px] font-bold uppercase tracking-wider">Likelihood</div>
            <div class="text-xs">{{ likelihoodLabels[risk.likelihood] }} ({{ risk.likelihood }})</div>
          </div>
        </div>

        <div class="space-y-2 pt-2 border-t border-primary-900">
          <div class="flex justify-between items-center">
            <span class="text-[10px] font-bold uppercase tracking-wider">Severity Weight</span>
            <span class="text-xs font-bold">{{ risk.severity }}%</span>
          </div>
          <UProgress 
            :value="risk.severity" 
            size="sm" 
            color="primary"
            :ui="{ progress: { background: 'bg-gradient-to-r from-success-500 via-warning-500 to-error-500' } }"
          />
        </div>

        <div v-if="risk.description" class="pt-3 border-t border-primary-900">
          <p class="text-[11px] leading-relaxed italic">
            {{ risk.description }}
          </p>
        </div>
      </div>
    </template>

    <div
      :id="`risk-badge-${risk.id}`"
      :class="[
        'relative w-9 h-9 rounded-full flex items-center justify-center shrink-0 select-none cursor-grab transition-all duration-200 ease-in-out',
        'shadow-[0_4px_12px_rgba(0,0,0,0.4)]',
        'hover:scale-[1.15] hover:shadow-[0_8px_24px_rgba(0,0,0,0.5)] hover:ring-4 hover:ring-white/15',
        'active:cursor-grabbing active:scale-105',
        { 'ring-2 ring-white/20': isPriority, 'opacity-40 scale-90': isDragging }
      ]"
      :style="{ backgroundColor: config.color, zIndex: props.zIndex }"
      :draggable="true"
      @dragstart="onDragStart"
      @dragend="onDragEnd"
    >
      <span class="text-[13px] font-extrabold text-white [text-shadow:0_2px_4px_rgba(0,0,0,0.4)] leading-none">{{ risk.id }}</span>
      <div v-if="isPriority" class="absolute -inset-1 rounded-full border-2 pointer-events-none animate-badge-pulse" :style="{ borderColor: config.color }"></div>
    </div>
  </UTooltip>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRiskProfileStore, riskLevelConfig, categoryIcons, impactLabels, likelihoodLabels } from '~/stores/risk-profile'

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

const store = useRiskProfileStore()
const isDragging = ref(false)

const riskLevel = computed(() => store.getRiskLevel(props.risk.likelihood, props.risk.impact))
const config = computed(() => riskLevelConfig[riskLevel.value])
const isPriority = computed(() => config.value.priority)
const categoryIcon = computed(() => categoryIcons[props.risk.category] || '📌')

const formattedId = computed(() => store.getFormattedId(props.risk))

function onDragStart(e) {
  isDragging.value = true
  e.dataTransfer.effectAllowed = 'move'
  e.dataTransfer.setData('application/json', JSON.stringify(props.risk))

  // Create drag ghost image
  const el = e.target.cloneNode(true)
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

<style>
/* We define the animation here because it's specific to this component.
   It could also be added to tailwind.config.js for global use. */
@keyframes badge-pulse {
  0%, 100% { opacity: 0; transform: scale(1); }
  50% { opacity: 0.6; transform: scale(1.2); }
}
.animate-badge-pulse {
  animation: badge-pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}
</style>
