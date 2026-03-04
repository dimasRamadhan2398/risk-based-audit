<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6 bg-gray-50 dark:bg-gray-900 min-h-screen">
    
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Kertas Kerja Audit</h1>
        <p class="text-sm text-gray-500">Digital Working Paper System</p>
      </div>
    </div>

    <UTabs :items="store.tabs" class="w-full">
      
      <template #f01="{ item }">
        <WorkingPaperHeader />
      </template>

      <template #f02="{ item }">
        <WorkingPaperRisk />
      </template>

      <template #f03="{ item }">
        <WorkingPaperSample />
      </template>

      <template #f04="{ item }">
        <WorkingPaperCause />
      </template>

      <template #f05="{ item }">
        <WorkingPaperPlan />
      </template>
    </UTabs>
  </div>
</template>

<script setup lang="ts">
import WorkingPaperCause from '~/components/working-paper/WorkingPaperCause.vue';
import WorkingPaperHeader from '~/components/working-paper/WorkingPaperHeader.vue';
import WorkingPaperPlan from '~/components/working-paper/WorkingPaperPlan.vue';
import WorkingPaperRisk from '~/components/working-paper/WorkingPaperRisk.vue';
import WorkingPaperSample from '~/components/working-paper/WorkingPaperSample.vue';
import { useWorkingPaperStore } from '~/stores/working-paper'

// Panggil Store
const store = useWorkingPaperStore()

</script>