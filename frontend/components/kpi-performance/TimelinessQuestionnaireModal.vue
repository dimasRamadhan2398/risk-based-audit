<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = defineProps<{
  modelValue: boolean
  year: number
  period: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'saved'): void
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const isLoading = ref(false)

interface QuestionnaireItem {
  id?: string
  questionnaire_name: string
  total_reports_planned: number
  total_reports_completed: number
  reports_completed_on_time: number
  remarks: string
  timeliness_percentage?: number
}

const formList = ref<QuestionnaireItem[]>([])

const calculatePercentage = (item: QuestionnaireItem) => {
  if (item.total_reports_completed === 0) return 0
  return Math.round((item.reports_completed_on_time / item.total_reports_completed) * 100)
}

const loadData = async () => {
  try {
    isLoading.value = true
    const res: any = await $fetch(`http://localhost:8002/api/v1/report-timeliness?year=${props.year}&period=${props.period}`)
    if (res && Array.isArray(res)) {
      formList.value = res.map((r: any) => ({
        id: r.id,
        questionnaire_name: r.questionnaire_name || '',
        total_reports_planned: r.total_reports_planned || 0,
        total_reports_completed: r.total_reports_completed || 0,
        reports_completed_on_time: r.reports_completed_on_time || 0,
        remarks: r.remarks || ''
      }))
    } else {
      formList.value = []
    }
  } catch (error: any) {
    console.error('Failed to load timeliness data:', error)
    formList.value = []
  } finally {
    isLoading.value = false
  }
}

watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    loadData()
  } else {
    formList.value = []
  }
})

const addQuestionnaire = () => {
  formList.value.push({
    questionnaire_name: '',
    total_reports_planned: 0,
    total_reports_completed: 0,
    reports_completed_on_time: 0,
    remarks: ''
  })
}

const removeQuestionnaire = (index: number) => {
  formList.value.splice(index, 1)
}

const handleSave = async () => {
  try {
    isLoading.value = true
    
    const payload = formList.value.map(item => ({
      ...item,
      timeliness_percentage: calculatePercentage(item)
    }))

    await $fetch('http://localhost:8002/api/v1/report-timeliness', {
      method: 'POST',
      body: {
        year: props.year,
        period: props.period,
        questionnaires: payload
      }
    })
    
    useToast().add({
      title: 'Success',
      description: 'Report timeliness data saved successfully',
      color: 'success'
    })
    emit('saved')
    isOpen.value = false
  } catch (error) {
    console.error('Failed to save timeliness data:', error)
    useToast().add({
      title: 'Error',
      description: 'Failed to save report timeliness data',
      color: 'error'
    })
  } finally {
    isLoading.value = false
  }
}

const columns = [
  { accessorKey: 'questionnaire_name', header: 'Quisioner' },
  { accessorKey: 'total_reports_planned', header: 'Planned' },
  { accessorKey: 'total_reports_completed', header: 'Completed' },
  { accessorKey: 'reports_completed_on_time', header: 'On Time' },
  { accessorKey: 'percentage', header: '%' },
  { accessorKey: 'remarks', header: 'Remarks' },
  { id: 'actions', header: '' }
]
</script>

<template>
  <UModal v-model="isOpen" :ui="{ width: 'sm:max-w-5xl' }">
    <UCard :ui="{ ring: '', divide: 'divide-y divide-gray-100 dark:divide-gray-800' }">
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-base font-semibold leading-6 text-gray-900 dark:text-white">
            Report Timeliness Questionnaire ({{ period }} {{ year }})
          </h3>
          <div class="flex items-center gap-3">
            <UButton icon="i-lucide-plus" size="sm" color="primary" @click="addQuestionnaire">
              Add Questionnaire
            </UButton>
            <UButton color="neutral" variant="ghost" icon="i-heroicons-x-mark-20-solid" class="-my-1" @click="isOpen = false" />
          </div>
        </div>
      </template>

      <div class="space-y-4" v-if="!isLoading">
        <UTable :columns="columns" :data="formList" :empty-state="{ icon: 'i-lucide-file-question', label: 'No questionnaires added yet.' }">
          <template #questionnaire_name-cell="{ row }">
            <UInput v-model="row.original.questionnaire_name" placeholder="Quisioner Name" size="sm" />
          </template>
          <template #total_reports_planned-cell="{ row }">
            <UInput type="number" v-model.number="row.original.total_reports_planned" min="0" size="sm" class="w-20" />
          </template>
          <template #total_reports_completed-cell="{ row }">
            <UInput type="number" v-model.number="row.original.total_reports_completed" min="0" size="sm" class="w-20" />
          </template>
          <template #reports_completed_on_time-cell="{ row }">
            <UInput type="number" v-model.number="row.original.reports_completed_on_time" min="0" :max="row.original.total_reports_completed" size="sm" class="w-20" />
          </template>
          <template #percentage-cell="{ row }">
            <div class="font-bold text-center" :class="{
              'text-green-600': calculatePercentage(row.original) >= 90,
              'text-orange-500': calculatePercentage(row.original) >= 70 && calculatePercentage(row.original) < 90,
              'text-red-500': calculatePercentage(row.original) < 70
            }">
              {{ calculatePercentage(row.original) }}%
            </div>
          </template>
          <template #remarks-cell="{ row }">
            <UInput v-model="row.original.remarks" placeholder="Remarks..." size="sm" />
          </template>
          <template #actions-cell="{ row }">
            <UButton color="error" variant="ghost" icon="i-lucide-trash-2" size="sm" @click="removeQuestionnaire(row.index)" />
          </template>
        </UTable>
      </div>
      
      <div v-else class="flex justify-center p-8">
        <UIcon name="i-lucide-loader-2" class="w-8 h-8 animate-spin text-primary" />
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <UButton color="neutral" variant="soft" @click="isOpen = false" :disabled="isLoading">Cancel</UButton>
          <UButton color="primary" @click="handleSave" :loading="isLoading" :disabled="formList.length === 0">Save Metrics</UButton>
        </div>
      </template>
    </UCard>
  </UModal>
</template>

