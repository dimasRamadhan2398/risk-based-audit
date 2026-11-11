<template>
  <span class="inline-flex flex-wrap items-center gap-1">
    <UBadge
      :color="followUpStatusColor(status)"
      variant="subtle"
      :size="size"
      class="font-semibold"
    >
      {{ statusLabel }}
    </UBadge>
    <UBadge
      v-if="status === 'No Action' && reasonKey"
      color="warning"
      variant="outline"
      :size="size"
      icon="i-lucide-info"
    >
      {{ t(reasonKey) }}
    </UBadge>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'
import {
  followUpStatusColor,
  followUpStatusI18nKey,
  noActionReasonI18nKey
} from '~/utils/executiveSummaryFollowUp'

// Follow-up status of a finding (ATR bucket) with the reason badge for No Action rows
const props = withDefaults(defineProps<{
  status: string
  reason?: string | null
  size?: 'xs' | 'sm' | 'md'
}>(), {
  reason: null,
  size: 'xs'
})

const { t } = useI18n()

const statusLabel = computed(() => {
  const key = followUpStatusI18nKey(props.status)
  return key ? t(key) : props.status
})
const reasonKey = computed(() => noActionReasonI18nKey(props.reason))
</script>
