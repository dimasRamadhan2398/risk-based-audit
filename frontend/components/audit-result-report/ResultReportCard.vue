<template>
  <UCard class="hover:shadow-md transition-shadow duration-300 overflow-hidden">
    <div class="flex justify-between items-start mb-4">
      <div class="flex items-center gap-2">
        <UBadge :color="row.status === 'Final' ? 'success' : 'neutral'" variant="soft" size="sm">
          {{ row.status }}
        </UBadge>
        <span class="text-md text-gray-400">{{ row.reportDate }}</span>
      </div>
    </div>

    <h3 class="font-bold text-gray-900 mb-2 truncate" :title="row.reportTitle">
      {{ row.reportTitle }}
    </h3>
    
    <p class="text-sm text-gray-500 line-clamp-2 mb-4 h-10">
      {{ row.executiveSummary || 'No executive summary provided.' }}
    </p>

    <div class="flex items-center justify-between pt-4 border-t border-gray-100">
      <div class="flex items-center gap-1 text-md text-gray-500">
        <UIcon name="i-heroicons-magnifying-glass" class="size-4" />
        <span>{{ row.findingsCount }} Findings</span>
      </div>
      <div class="flex gap-1">
        <UButton
          color="warning"
          variant="ghost"
          icon="i-lucide-edit"
          size="md"
          @click="$emit('edit', row)"
        />
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-lucide-printer"
          size="md"
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
    case 'Very Significant': return 'error'
    case 'Significant': return 'warning'
    case 'Moderately Significant': return 'info'
    case 'Insignificant': return 'success'
    default: return 'neutral'
  }
}
</script>