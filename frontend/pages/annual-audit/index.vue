<template>
  <div class="p-6 max-w-full mx-auto space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900 "> 
        {{ t('annualAudit.title') }}
      </h1>
      <div class="flex items-center gap-2">
        <UButton
          :label="t('annualAudit.importPlanDocument')"
          to="/annual-audit/upload"
          color="neutral"
          variant="outline"
          class="px-4 font-bold shadow flex gap-2"
          icon="i-lucide-upload"
        />
        <UButton
          :label="t('annualAudit.newAuditPlan')" 
          @click="store.openModal()"
          color="primary" 
          class="px-4 font-bold shadow-lg flex gap-2"
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

const { t } = useI18n()

// Inisialisasi Store
const store = useAnnualPlanStore()
store.fetchPlans()
</script>
