<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6 bg-gray-50 dark:bg-gray-900 min-h-screen">
    
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Audit Working Paper</h1>
        <p class="text-sm text-gray-500">Digital Working Paper System</p>
      </div>
    </div>

    <UTabs :items="store.tabs" class="w-full">
      
      <template #f01="{ item }">
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 pt-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Audit Working Paper Header</h1>
          </div>
          <UButton 
            label="Buat Kertas Kerja" 
            icon="i-heroicons-plus" 
            color="primary" 
            size="lg" 
            class="font-bold shadow-md"
            @click="store.openModalF01()" 
          />
        </div>

        <WorkingPaperHeaderTable />

        <WorkingPaperHeaderForm />

      </template>

      <template #f02="{ item }">
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 pt-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Audit Working Paper Risk</h1>
          </div>
          <UButton 
            label="Buat Kertas Kerja" 
            icon="i-heroicons-plus" 
            color="primary" 
            size="lg" 
            class="font-bold shadow-md"
            @click="store.openModalF02()" 
          />
        </div>

        <WorkingPaperRiskTable />

        <WorkingPaperRiskForm />
        
      </template>

      <template #f03="{ item }">
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 pt-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Audit Working Paper Sample</h1>
          </div>
          <UButton 
            label="Buat Kertas Kerja" 
            icon="i-heroicons-plus" 
            color="primary" 
            size="lg" 
            class="font-bold shadow-md"
            @click="store.openModalF03()" 
          />
        </div>

        <WorkingPaperSampleTable />

        <WorkingPaperSampleForm />
      </template>

      <template #f04="{ item }">
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 pt-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Audit Working Paper Cause Analysis</h1>
          </div>
          <UButton 
            label="Buat Kertas Kerja" 
            icon="i-heroicons-plus" 
            color="primary" 
            size="lg" 
            class="font-bold shadow-md"
            @click="store.openModalF04()" 
          />
        </div>

        <WorkingPaperCauseTable />

        <WorkingPaperCauseForm />
      </template>

      <template #f05="{ item }">
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 pt-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Audit Working Paper Action Plan</h1>
          </div>
          <UButton 
            label="Buat Kertas Kerja" 
            icon="i-heroicons-plus" 
            color="primary" 
            size="lg" 
            class="font-bold shadow-md"
            @click="store.openModalF05()" 
          />
        </div>

        <WorkingPaperPlanTable />

        <WorkingPaperPlanForm />
      </template>
    </UTabs>
  </div>
</template>

<script setup lang="ts">
import WorkingPaperCauseForm from '~/components/working-paper/WorkingPaperCauseForm.vue';
import WorkingPaperCauseTable from '~/components/working-paper/WorkingPaperCauseTable.vue';
import WorkingPaperHeaderForm from '~/components/working-paper/WorkingPaperHeaderForm.vue';
import WorkingPaperHeaderTable from '~/components/working-paper/WorkingPaperHeaderTable.vue';
import WorkingPaperPlanForm from '~/components/working-paper/WorkingPaperPlanForm.vue';
import WorkingPaperPlanTable from '~/components/working-paper/WorkingPaperPlanTable.vue';
import WorkingPaperRiskForm from '~/components/working-paper/WorkingPaperRiskForm.vue';
import WorkingPaperRiskTable from '~/components/working-paper/WorkingPaperRiskTable.vue';
import WorkingPaperSampleForm from '~/components/working-paper/WorkingPaperSampleForm.vue';
import WorkingPaperSampleTable from '~/components/working-paper/WorkingPaperSampleTable.vue';
import { useWorkingPaperStore } from '~/stores/working-paper'

// Panggil Store
const store = useWorkingPaperStore()

</script>