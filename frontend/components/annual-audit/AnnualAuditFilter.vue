<template>
  <div class="flex flex-col md:flex-row gap-4 mb-4  p-4 rounded-xl">
    <CommonSearchInput
      v-model="store.searchCode"
      :placeholder="t('annualAudit.extra.searchActivity')"
      class="flex-1"
    />

    <USelectMenu
      v-model="store.selectedDepartment"
      :items="store.departmentOptions"
      :placeholder="t('annualAudit.extra.allDepartments')"
      class="w-full md:w-60"
      clearable
    />

    <USelectMenu
      v-model="store.selectedStatus"
      :items="store.statusOptions"
      value-key="value"
      label-key="label"
      :placeholder="t('settings.dataSource.allStatuses')"
      class="w-full md:w-60"
      clearable
    />

    <UButton
      v-if="store.searchCode || store.selectedDepartment || store.selectedStatus"
      icon="i-lucide-rotate-ccw"
      color="neutral"
      variant="outline"
      :label="t('strategicPlan.filters.reset')"
      @click="store.clearFilters"
    />
  </div>
</template>

<script setup lang="ts">
import { useAnnualPlanStore } from '~/stores/annual-audit'

const { t } = useI18n()

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useAnnualPlanStore()
</script>
