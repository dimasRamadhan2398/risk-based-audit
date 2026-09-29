<script setup lang="ts">
import QAAddModal from '~/components/quality-assurance/QAAddModal.vue';

import QAFilters from '~/components/quality-assurance/QAFilters.vue';
import QASummaryCards from '~/components/quality-assurance/QASummaryCards.vue'
import QATable from '~/components/quality-assurance/QATable.vue';
import QADetailModal from '~/components/quality-assurance/QADetailModal.vue';
import { useQualityAssuranceStore } from '~/stores/quality-assurance'

const store = useQualityAssuranceStore()

</script>

<template>
  <div class="p-4 sm:p-6 space-y-6 sm:space-y-8 min-w-0">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div class="space-y-1">
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white tracking-tight">
          Internal Audit Quality Assurance - Integrated Dashboard
        </h1>
        <p class="text-sm sm:text-base text-gray-600 dark:text-gray-400 font-medium">
          Monitoring Compliance to Global Standards (IPPF 2027 & GIAS 2024)
        </p>
      </div>
      <div class="flex items-center gap-3 w-full sm:w-auto">
        <UButton
          color="primary"
          label="Add Report"
          icon="i-lucide-plus"
          class="w-full sm:w-auto px-6 py-2.5 font-bold rounded-lg shadow-lg"
          @click="store.openForm"
        />
      </div>
    </div>

    <!-- Summary Cards -->
    <QASummaryCards />

    <!-- Filters -->
    <QAFilters />

    <!-- Table -->
    <QATable />

    <!-- Add Modal -->
    <QAAddModal />

    <!-- Detail Modal -->
    <QADetailModal />

  </div>
</template>
