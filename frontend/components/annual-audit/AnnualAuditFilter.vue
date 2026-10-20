<template>
    <div class="flex flex-col md:flex-row gap-4 mb-4  p-4 rounded-xl shadow-sm border border-gray-100 ">
      <UInput 
        v-model="store.searchCode" 
        icon="i-heroicons-magnifying-glass" 
        :placeholder="t('annualAudit.extra.searchActivity')"
        class="flex-1"
        clearable
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
        icon="i-heroicons-x-mark" 
        color="neutral" 
        variant="ghost" 
        @click="store.clearFilters" 
        :title="t('strategicPlan.filters.reset')"
      />
    </div>
</template>

<script setup lang="ts">
const { t } = useI18n()
import { useAnnualPlanStore } from '~/stores/annual-audit'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useAnnualPlanStore()
</script>
