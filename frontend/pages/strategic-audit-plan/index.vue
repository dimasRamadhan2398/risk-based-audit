<template>
  <div class="space-y-6 min-w-0">
    <!-- Header Section -->
    <UCard variant="soft">
      <template #header>
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div class="flex flex-col gap-1">
            <h1 class="text-2xl sm:text-3xl font-bold text-[var(--text-main)]">
              {{ t('strategicPlan.title') }}
            </h1>
            <p class="text-xs sm:text-sm text-[var(--text-muted)]">
              {{ t('strategicPlan.subtitle') }}
            </p>
          </div>
          <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 w-full sm:w-auto">
            <UButton
              v-if="canImportPlanDocs"
              :label="t('strategicPlan.importDocument')"
              to="/strategic-audit-plan/upload"
              color="neutral"
              variant="outline"
              size="sm"
              class="w-full sm:w-auto justify-center"
              icon="i-lucide-upload"
            />
            <UButton
              v-if="canManageStrategicPlan"
              icon="add"
              :label="t('strategicPlan.addObjective')"
              variant="solid"
              color="primary"
              size="sm"
              class="w-full sm:w-auto justify-center"
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
      closable
      @close="store.errorMsg = ''"
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
const { canManageStrategicPlan, canImportPlanDocs } = useRbac()

// Inisialisasi Store
const store = useStrategicPlanStore()
store.fetchStrategicPlans()

const vmgStore = useVisionMissionGoalsStore()
vmgStore.fetchCompaniesAndVmg()

</script>
