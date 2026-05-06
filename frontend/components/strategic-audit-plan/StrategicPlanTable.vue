<template>
    <UCard variant="soft">
        <!-- Strategic Objectives Table -->
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
          </UTable>
    </UCard>
</template>

<script setup lang="ts">
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'

const store = useStrategicPlanStore()

</script>