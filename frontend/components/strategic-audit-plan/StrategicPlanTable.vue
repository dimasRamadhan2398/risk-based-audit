<template>
    <UCard variant="soft">
        <!-- Strategic KPI Table -->
          <UTable :data="store.strategicObjectives" :columns="store.columns">
            <template #actions-cell="{ row }">
              <UDropdownMenu
                :items="store.getRowActions(row)"
                aria-label="Actions"
              >
                <UButton
                  icon="i-lucide-ellipsis-vertical"
                  color="primary"
                  variant="ghost"
                  aria-label="Actions dropdown"
                />
              </UDropdownMenu>
            </template>

            <!-- Custom cell for status with color coding -->
            <template #status-cell="{ row }">
              <span
                :class="{
                  'text-green-600 font-semibold': row.original.status === 'Good',
                  'text-yellow-600 font-semibold': row.original.status === 'Moderate',
                  'text-red-600 font-semibold': row.original.status === 'Poor',
                }"
              >
                {{ row.original.status || '-' }}
              </span>
            </template>
          </UTable>
    </UCard>
</template>

<script setup lang="ts">
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'

const store = useStrategicPlanStore()
</script>