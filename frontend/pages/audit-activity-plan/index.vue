<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <UCard variant="soft">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex flex-col gap-2">
            <h1 class="text-3xl font-bold text-[var(--text-main)]">
              Audit Activity Plan
            </h1>
            <p class="text-sm text-[var(--text-muted)]">
              Manage and review annual audit activity plans.
            </p>
          </div>
          <UButton
            icon="i-heroicons-plus"
            label="Buat Rencana"
            variant="solid"
            color="primary"
            size="sm"
            @click="store.openModal()"
          />
        </div>
      </template>
    </UCard>

    <!-- Modals (These must be in the template to be shown) -->
    <ActivityPlanForm />
    
    <!-- Data Table -->
    <ActivityPlanTable />
  </div>
  
</template>

<script setup lang="ts">
import ActivityPlanForm from '~/components/audit-activity-plan/ActivityPlanForm.vue'
import ActivityPlanTable from '~/components/audit-activity-plan/ActivityPlanTable.vue'
import { useActivityPlanStore } from '~/stores/activity-plan'

const store = useActivityPlanStore()
store.fetchPlans()
</script>