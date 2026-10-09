<template>
  <div class="flex flex-wrap items-center gap-1.5">
    <UBadge
      :color="atrStatusBadgeColor(status)"
      variant="subtle"
      :size="size"
      :label="label"
    />
    <UBadge
      v-if="isOverdue"
      color="error"
      variant="subtle"
      :size="size"
      icon="i-lucide-alarm-clock"
      :label="overdueDays > 0 ? t('actionTakenReport.daysOverdue', { days: overdueDays }) : t('actionTakenReport.status.overdue')"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { atrStatusBadgeColor, atrStatusI18nKey } from '~/utils/actionTakenReport'

const props = withDefaults(defineProps<{
  status: string
  isOverdue?: boolean
  overdueDays?: number
  size?: 'sm' | 'md'
}>(), {
  isOverdue: false,
  overdueDays: 0,
  size: 'sm'
})

const { t } = useI18n()

const label = computed(() => {
  const key = atrStatusI18nKey(props.status)
  return key ? t(key) : (props.status || '-')
})
</script>
