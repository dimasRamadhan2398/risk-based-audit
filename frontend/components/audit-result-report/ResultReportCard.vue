<template>
  <UCard class="hover:shadow-md transition-shadow duration-300 overflow-hidden">
    <div class="flex justify-between items-start mb-4">
      <div class="flex items-center gap-2">
        <UBadge :color="row.status === 'Final' ? 'success' : 'neutral'" variant="soft" size="sm">
          {{ row.status }}
        </UBadge>
        <span class="text-xs text-gray-400">{{ row.reportDate }}</span>
      </div>
      <UBadge :color="getRatingColor(row.overallRating)" variant="subtle" size="sm">
        {{ row.overallRating }}
      </UBadge>
    </div>

    <h3 class="font-bold text-gray-900 mb-2 truncate" :title="row.reportTitle">
      {{ row.reportTitle }}
    </h3>
    
    <p class="text-sm text-gray-500 line-clamp-2 mb-4 h-10">
      {{ row.executiveSummary || 'No executive summary provided.' }}
    </p>

    <div class="flex items-center justify-between pt-4 border-t border-gray-100">
      <div class="flex items-center gap-1 text-xs text-gray-500">
        <UIcon name="i-heroicons-magnifying-glass" class="size-4" />
        <span>{{ row.findingsCount }} Findings</span>
      </div>
      <div class="flex gap-1">
        <UButton
          color="primary"
          variant="ghost"
          icon="i-heroicons-pencil-square"
          size="xs"
          @click="$emit('edit', row)"
        />
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-heroicons-printer"
          size="xs"
          @click="$emit('print', row)"
        />
      </div>
    </div>
  </UCard>
</template>

<script setup lang="ts">
import type { AuditResultReport } from '~/stores/audit-result-report'

defineProps<{
  row: AuditResultReport
}>()

defineEmits(['edit', 'print'])

const getRatingColor = (rating: string) => {
  switch (rating) {
    case 'Satisfactory': return 'success'
    case 'Needs Improvement': return 'warning'
    case 'Unsatisfactory': return 'error'
    default: return 'neutral'
  }
}
</script>