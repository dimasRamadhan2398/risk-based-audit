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
                Strategic Planning for Risk-Based Audit
              </p>
            </div>
            <UButton
              icon="add"
              label="Add Strategic Plan"
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

    <!-- Vision, Mission and Goals Strategic Card -->
    <VisionMissionCard />
    <VisionMissionForm />

    <!-- Error Alert -->
    <UAlert
      v-if="store.errorMsg"
      title="Error"
      :description="store.errorMsg"
      color="error"
      variant="soft"
      icon="i-lucide-alert-circle"
      class="mb-4"
    />

    <!-- Strategic Plan Table -->
    <StrategicPlanTable />

  </div>
</template>

<script setup lang="ts">

import StrategicPlanForm from "~/components/strategic-audit-plan/StrategicPlanForm.vue";
import StrategicPlanTable from "~/components/strategic-audit-plan/StrategicPlanTable.vue";
import VisionMissionCard from "~/components/strategic-audit-plan/VisionMissionCard.vue";
import VisionMissionForm from "~/components/strategic-audit-plan/VisionMissionForm.vue";
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useVisionMissionGoalsStore } from '~/stores/vision-mission-goals'

// Inisialisasi Store
const store = useStrategicPlanStore()
store.fetchStrategicPlans()

const vmgStore = useVisionMissionGoalsStore()
vmgStore.fetchCompaniesAndVmg()

</script>