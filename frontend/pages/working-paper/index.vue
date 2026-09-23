<template>
  <div class="p-6 max-w-full mx-auto space-y-6 min-h-screen">
    
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 ">{{ t('workingPaper.index.title') }}</h1>
        <p class="text-sm text-gray-500">{{ t('workingPaper.index.subtitle') }}</p>
      </div>
      <UButton
        :label="t('workingPaper.index.importButton')"
        icon="i-lucide-upload"
        color="neutral"
        variant="outline"
        size="md"
        class="font-bold shadow"
        to="/working-paper/upload"
      />
    </div>

    <UStepper v-model="activeStep" :items="stepItems" class="w-full">
      
      <template #f01>
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 pt-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 ">{{ t('workingPaper.index.sections.headerTitle') }}</h1>
          </div>
          <UButton 
            :label="t('workingPaper.index.createButton')" 
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
            <h1 class="text-2xl font-bold text-gray-900 ">{{ t('workingPaper.index.sections.riskTitle') }}</h1>
          </div>
          <UButton 
            :label="t('workingPaper.index.createButton')" 
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
            <h1 class="text-2xl font-bold text-gray-900 ">{{ t('workingPaper.index.sections.sampleTitle') }}</h1>
          </div>
          <UButton 
            :label="t('workingPaper.index.createButton')" 
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
            <h1 class="text-2xl font-bold text-gray-900 ">{{ t('workingPaper.index.sections.causeTitle') }}</h1>
          </div>
          <UButton 
            :label="t('workingPaper.index.createButton')" 
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
            <h1 class="text-2xl font-bold text-gray-900 ">{{ t('workingPaper.index.sections.planTitle') }}</h1>
          </div>
          <UButton 
            :label="t('workingPaper.index.createButton')" 
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
import { onMounted, computed, ref, watch } from 'vue'
import type { StepperItem } from '@nuxt/ui'
import { useI18n } from '~/composables/useI18n'
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

const { t } = useI18n()

// Panggil Store
const store = useWorkingPaperStore()
const route = useRoute()

const stepItems = computed<StepperItem[]>(() => [
  {
    slot: 'f01' as const,
    title: t('workingPaper.index.steps.header.title'),
    description: t('workingPaper.index.steps.header.description'),
    icon: 'i-lucide-house'
  },
  {
    slot: 'f02' as const,
    title: t('workingPaper.index.steps.riskProfile.title'),
    description: t('workingPaper.index.steps.riskProfile.description'),
    icon: 'i-lucide-shield'
  },
  {
    slot: 'f03' as const,
    title: t('workingPaper.index.steps.testSample.title'),
    description: t('workingPaper.index.steps.testSample.description'),
    icon: 'i-lucide-table'
  },
  {
    slot: 'f04' as const,
    title: t('workingPaper.index.steps.aoiRca.title'),
    description: t('workingPaper.index.steps.aoiRca.description'),
    icon: 'i-lucide-file-search'
  },
  {
    slot: 'f05' as const,
    title: t('workingPaper.index.steps.actionPlan.title'),
    description: t('workingPaper.index.steps.actionPlan.description'),
    icon: 'i-lucide-check'
  }
])

const stepMap: Record<string, number> = {
  f01: 0,
  f02: 1,
  f03: 2,
  f04: 3,
  f05: 4
}

const activeStep = ref(
  route.query.step && typeof route.query.step === 'string' && stepMap[route.query.step] !== undefined
    ? stepMap[route.query.step]
    : 0
)

watch(() => route.query.step, (step) => {
  if (typeof step === 'string' && stepMap[step] !== undefined) {
    activeStep.value = stepMap[step]
  }
})

onMounted(() => {
  store.fetchAllData()
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
