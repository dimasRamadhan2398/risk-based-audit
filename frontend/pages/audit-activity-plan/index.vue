<template>
  <div class="p-4 sm:p-6 space-y-6 min-w-0">
    <!-- Header Section -->
    <UCard variant="soft">
      <template #header>
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div class="flex flex-col gap-2">
            <h1 class="text-2xl sm:text-3xl font-bold text-[var(--text-main)]">
              {{ t('auditActivityPlan.title') }}
            </h1>
            <p class="text-sm text-[var(--text-muted)]">
              {{ t('auditActivityPlan.subtitle') }}
            </p>
          </div>
          <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 w-full sm:w-auto">
            <UButton
              v-if="canImportPlanDocs"
              icon="i-lucide-upload"
              :label="t('auditActivityPlan.importDocument')"
              variant="outline"
              color="neutral"
              size="sm"
              class="w-full sm:w-auto font-bold"
              to="/audit-activity-plan/upload"
            />
            <UButton
              icon="i-heroicons-plus"
              :label="t('auditActivityPlan.createPlan')"
              variant="solid"
              color="primary"
              size="sm"
              class="w-full sm:w-auto font-bold"
              @click="store.openModal()"
            />
          </div>
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
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'

const { t } = useI18n()
const { canImportPlanDocs } = useRbac()
const store = useActivityPlanStore()
store.fetchPlans()
</script>
