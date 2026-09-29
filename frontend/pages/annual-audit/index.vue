<template>
  <div class="p-4 sm:p-6 max-w-full mx-auto space-y-6 min-w-0">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white"> 
        {{ t('annualAudit.title') }}
      </h1>
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 w-full sm:w-auto">
        <UButton
          v-if="canImportPlanDocs"
          :label="t('annualAudit.importPlanDocument')"
          to="/annual-audit/upload"
          color="neutral"
          variant="outline"
          class="w-full sm:w-auto justify-center px-4 font-bold shadow flex gap-2"
          icon="i-lucide-upload"
        />
        <UButton
          v-if="canManageAnnualPlan"
          :label="t('annualAudit.newAuditPlan')" 
          @click="store.openModal()"
          color="primary" 
          class="w-full sm:w-auto justify-center px-4 font-bold shadow-lg flex gap-2"
          icon="i-lucide-plus"
        />
      </div>
    </div>

    <AnnualAuditFilter />

    <AnnualAuditTable />

    <AnnualAuditForm />
  
    <AnnualAuditDetail />
          
  </div>
</template>

<script setup lang="ts">
import AnnualAuditDetail from '~/components/annual-audit/AnnualAuditDetail.vue';
import AnnualAuditFilter from '~/components/annual-audit/AnnualAuditFilter.vue';
import AnnualAuditForm from '~/components/annual-audit/AnnualAuditForm.vue';
import AnnualAuditTable from '~/components/annual-audit/AnnualAuditTable.vue';
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'

const { t } = useI18n()
const { canManageAnnualPlan, canImportPlanDocs } = useRbac()

// Inisialisasi Store
const store = useAnnualPlanStore()
store.fetchPlans()
</script>
