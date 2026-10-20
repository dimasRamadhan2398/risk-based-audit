<template>
  <div class="flex flex-col md:flex-row md:items-end gap-3">
    <UFormField
      :label="t('auditActivityPlan.filter.year')"
      class="w-full md:w-48"
    >
      <USelectMenu
        :model-value="store.filters.year"
        :items="yearItems"
        value-key="value"
        label-key="label"
        :placeholder="t('auditActivityPlan.filter.all')"
        class="w-full"
        data-testid="filter-year"
        @update:model-value="(value: number | undefined) => { store.filters.year = value || undefined }"
      />
    </UFormField>

    <UFormField
      :label="t('auditActivityPlan.filter.activityType')"
      class="w-full md:w-64"
    >
      <USelectMenu
        :model-value="store.filters.activityType"
        :items="store.activityTypeOptions"
        value-key="value"
        label-key="label"
        :placeholder="t('auditActivityPlan.filter.all')"
        class="w-full"
        data-testid="filter-activity-type"
        @update:model-value="(value: ActivityPlanActivityType | undefined) => { store.filters.activityType = value || undefined }"
      />
    </UFormField>

    <UButton
      v-if="store.filters.year || store.filters.activityType"
      icon="i-heroicons-x-mark"
      color="neutral"
      variant="ghost"
      :label="t('auditActivityPlan.filter.reset')"
      @click="store.clearFilters()"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useActivityPlanStore } from '~/stores/activity-plan'
import { useI18n } from '~/composables/useI18n'
import { getFiscalYears } from '~/composables/useFiscalYear'
import type { ActivityPlanActivityType } from '~/types/audit'

/** Tahun and Jenis Kegiatan Audit filters of the Audit Activity Plan list (applied by the server). */
const { t } = useI18n()
const store = useActivityPlanStore()

// Fiscal years plus any year already in the list, newest first.
const yearItems = computed(() => {
  const years = new Set<number>(getFiscalYears())
  for (const plan of store.plans) if (plan.year) years.add(plan.year)
  if (store.filters.year) years.add(store.filters.year)
  return [{ value: undefined as number | undefined, label: t('auditActivityPlan.filter.all') }]
    .concat([...years].sort((a, b) => b - a).map(year => ({ value: year as number | undefined, label: String(year) })))
})

watch(() => [store.filters.year, store.filters.activityType], () => {
  store.fetchPlans()
})
</script>
