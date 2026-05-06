<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <UCard variant="soft">
      <template #header>
        <div class="flex flex-col gap-2">
          <div class="flex items-center justify-between">
            <div class="flex flex-col gap-2">
              <h1 class="text-3xl font-bold text-[var(--text-main)]">
                Strategic Plan Internal Audit
              </h1>
              <p class="text-sm text-[var(--text-muted)]">
                5 Year Strategic Planning for Risk-Based Audit
              </p>
            </div>
            <UButton
              icon="add"
              label="Add Strategic Objective"
              variant="solid"
              color="primary"
              size="sm"
              @click="store.openModal()"
            />
          </div>
        </div>
      </template>
    </UCard>

    <!-- Add/Edit Modal -->
    <StrategicPlanForm />

    <!-- Strategic Plan Table -->
    <StrategicPlanTable />

  </div>
</template>

<script setup lang="ts">

import StrategicPlanForm from "~/components/strategic-audit-plan/StrategicPlanForm.vue";
import StrategicPlanTable from "~/components/strategic-audit-plan/StrategicPlanTable.vue";
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'

// Inisialisasi Store
const store = useStrategicPlanStore()

</script>