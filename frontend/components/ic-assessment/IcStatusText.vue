<template>
  <span class="inline-flex items-center gap-1.5 font-medium" :class="visual.cls">
    <UIcon :name="visual.icon" class="size-4 shrink-0" aria-hidden="true" />
    <span>{{ label }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'
import type { ComponentStatus, PrincipleStatus } from '~/utils/icAssessment'
import { principleStatusVisual } from '~/utils/icAssessmentLabels'

// Status as icon + text (never colour alone). `short` uses the compact grid wording.
const props = defineProps<{
  status: ComponentStatus | PrincipleStatus | null
  kind?: 'component' | 'principle'
  short?: boolean
}>()

const { t } = useI18n()

const value = computed<PrincipleStatus>(() => props.status ?? 'NOT_ASSESSED')
const visual = computed(() => principleStatusVisual(value.value))
const label = computed(() => {
  if (props.kind === 'principle') return t(`icAssessment.principleStatus.${value.value}`)
  const ns = props.short ? 'componentStatusShort' : 'componentStatus'
  return t(`icAssessment.${ns}.${value.value === 'NOT_APPLICABLE' ? 'NOT_ASSESSED' : value.value}`)
})
</script>
