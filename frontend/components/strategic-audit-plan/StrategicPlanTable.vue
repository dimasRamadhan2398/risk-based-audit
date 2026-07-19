<template>
  <div class="space-y-4">
    <!-- Goal Selection Tabs -->
    <div v-if="vmgStore.activeVmg?.goals?.length" class="flex gap-2 border-b border-[var(--border-main)] pb-2 overflow-x-auto">
      <UButton
        v-for="tab in goalTabs"
        :key="tab.value"
        :label="tab.label"
        :variant="selectedGoalId === tab.value ? 'solid' : 'ghost'"
        :color="selectedGoalId === tab.value ? 'primary' : 'neutral'"
        size="sm"
        @click="() => { selectedGoalId = tab.value }"
      />
    </div>

    <UCard variant="soft">
      <!-- Strategic KPI Table -->
      <UTable :data="filteredObjectives" :columns="store.columns">
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useVisionMissionGoalsStore } from '~/stores/vision-mission-goals'

const store = useStrategicPlanStore()
const vmgStore = useVisionMissionGoalsStore()

const selectedGoalId = ref('ALL')

const goalTabs = computed(() => {
  const tabs = [{ label: 'All Goals', value: 'ALL' }]
  if (vmgStore.activeVmg?.goals) {
    vmgStore.activeVmg.goals.forEach(g => {
      tabs.push({
        label: `${g.goal_code} - ${g.goal_name}`,
        value: g.id || g.goal_code
      })
    })
  }
  return tabs
})

const filteredObjectives = computed(() => {
  if (selectedGoalId.value === 'ALL') {
    return store.strategicObjectives
  }
  const activeGoal = vmgStore.activeVmg?.goals?.find(g => g.id === selectedGoalId.value || g.goal_code === selectedGoalId.value)
  const activeGoalCode = activeGoal?.goal_code

  return store.strategicObjectives.filter(item => {
    return item.goalId === selectedGoalId.value || 
           (activeGoalCode && item.goalId === activeGoalCode)
  })
})
</script>