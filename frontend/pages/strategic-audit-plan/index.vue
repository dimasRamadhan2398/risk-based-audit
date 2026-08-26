<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <UCard variant="soft">
      <template #header>
        <div class="flex flex-col gap-2">
          <div class="flex items-center justify-between">
            <div class="flex flex-col gap-2">
              <h1 class="text-3xl font-bold text-[var(--text-main)]">
                {{ t('strategicPlan.title') }}
              </h1>
              <p class="text-sm text-[var(--text-muted)]">
                {{ t('strategicPlan.subtitle') }}
              </p>
            </div>
            <UButton
              v-if="canManageStrategicPlan"
              icon="add"
              :label="t('strategicPlan.addObjective')"
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

    <!-- View Objective Modal -->
    <StrategicObjectiveViewModal />

    <!-- Vision, Mission and Goals Strategic Card -->
    <VisionMissionCard />
    <VisionMissionForm />

    <!-- Error Alert -->
    <UAlert
      v-if="store.errorMsg"
      :title="t('strategicPlan.errorTitle')"
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
import StrategicObjectiveViewModal from "~/components/strategic-audit-plan/StrategicObjectiveViewModal.vue";
import VisionMissionCard from "~/components/strategic-audit-plan/VisionMissionCard.vue";
import VisionMissionForm from "~/components/strategic-audit-plan/VisionMissionForm.vue";
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useVisionMissionGoalsStore } from '~/stores/vision-mission-goals'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'

const { t } = useI18n()
const { canManageStrategicPlan } = useRbac()

// Inisialisasi Store
const store = useStrategicPlanStore()
store.fetchStrategicPlans()

const vmgStore = useVisionMissionGoalsStore()
vmgStore.fetchCompaniesAndVmg()

</script>
