<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6 min-h-screen">
    
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 ">Create Working Paper</h1>
        <p class="text-sm text-gray-500">Create and manage audit working papers step by step</p>
      </div>
    </div>

    <UStepper :items="store.workingItems" class="w-full">
      
      <template #f01>
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 pt-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 ">Audit Working Paper Header</h1>
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

      <template #f02>
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 pt-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 ">Audit Working Paper Risk</h1>
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

      <template #f03>
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 pt-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 ">Audit Working Paper Sample</h1>
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

      <template #f04>
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 pt-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 ">Audit Working Paper Cause Analysis</h1>
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

      <template #f05>
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 pt-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 ">Audit Working Paper Action Plan</h1>
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
    </UStepper>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { onMounted } from 'vue'
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
const route = useRoute()

onMounted(() => {
  const { id, action } = route.query
  
  if (action === 'create') {
    store.openModalF01()
  } else if (id && action === 'edit') {
    const wp = store.dataF01.find(item => item.id === id)
    if (wp) {
      store.handleEditF01(wp)
    }
  } else if (id) {
    // Optional: handle view mode if exists
    const wp = store.dataF01.find(item => item.id === id)
    if (wp) {
      store.handleEditF01(wp) // Use edit modal for viewing for now
    }
  }
})
</script>