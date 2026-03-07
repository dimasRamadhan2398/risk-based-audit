<template>
    <UCard class="rounded-xl shadow-sm overflow-hidden" variant="soft">
      <UTable
        :columns="store.columns"
        :data="store.assignmentLetterList"
        :empty-state="{ icon: 'i-heroicons-document-text', label: 'No assignment letters yet. Click Create Assignment Letter to start.' }"
        class="w-full text-sm"
      >
        <template #letterNumber-cell="{ row }">
          <span class="font-bold text-orange-600 dark:text-orange-400">
            {{ row.original.letterNumber }}
          </span>
        </template>

        <template #executionPeriod-cell="{ row }">
          <div class="flex items-center gap-1 text-gray-600 dark:text-gray-300">
            <UIcon name="i-heroicons-calendar" class="w-4 h-4" />
            <span>{{ row.original.executionPeriod }}</span>
          </div>
        </template>

        <template #auditTeam-cell="{ row }">
          <div class="flex flex-col">
            <span class="font-bold text-gray-800 dark:text-gray-200">{{ row.original.leader }} (PIC)</span>
            <span class="text-xs text-gray-500">{{ row.original.auditTeam }}</span>
          </div>
        </template>

        <template #status-cell="{ row }">
          <UBadge :color="row.original.status === 'Published' ? 'success' : 'neutral'" variant="subtle" size="lg" class="font-bold">
            {{ row.original.status }}
          </UBadge>
        </template>

      </UTable>
    </UCard>
</template>

<script setup lang="ts">
import { useAssignmentLetterStore } from '~/stores/assignment-letter'

const store = useAssignmentLetterStore()

</script>
