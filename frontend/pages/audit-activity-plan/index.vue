<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6">

    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 border-b border-gray-200 pb-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 ">Audit Activity Plan</h1>
      </div>
      <UButton
        label="Create Audit Activity Plan"
        icon="i-heroicons-plus"
        color="primary"
        size="lg"
        class="font-bold shadow-md"
        @click="store.openModal"
      />
    </div>

    <!-- Table Component -->
    <ActivityPlanTable />

    <!-- Form Modal Component -->
    <ActivityPlanForm />

  </div>
</template>

<script setup lang="ts">
import ActivityPlanForm from '~/components/audit-activity-plan/ActivityPlanForm.vue';
import ActivityPlanTable from '~/components/audit-activity-plan/ActivityPlanTable.vue';
import { useActivityPlanStore } from '~/stores/activity-plan'

const store = useActivityPlanStore()
</script>
