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

const form = ref({
  total_reports_planned: 0,
  total_reports_completed: 0,
  reports_completed_on_time: 0,
  remarks: ''
})

const timelinessPercentage = computed(() => {
  if (form.value.total_reports_completed === 0) return 0
  return Math.round((form.value.reports_completed_on_time / form.value.total_reports_completed) * 100)
})

const loadData = async () => {
  try {
    isLoading.value = true
    const res: any = await $fetch(`http://localhost:8002/api/v1/report-timeliness?year=${props.year}&period=${props.period}`)
    if (res) {
      form.value = {
        total_reports_planned: res.total_reports_planned || 0,
        total_reports_completed: res.total_reports_completed || 0,
        reports_completed_on_time: res.reports_completed_on_time || 0,
        remarks: res.remarks || ''
      }
    }
  } catch (error: any) {
    if (error.response?.status !== 404) {
      console.error('Failed to load timeliness data:', error)
    }
  } finally {
    isLoading.value = false
  }
}

watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    loadData()
  } else {
    form.value = {
      total_reports_planned: 0,
      total_reports_completed: 0,
      reports_completed_on_time: 0,
      remarks: ''
    }
  }
})

const handleSave = async () => {
  try {
    isLoading.value = true
    await $fetch('http://localhost:8002/api/v1/report-timeliness', {
      method: 'POST',
      body: {
        year: props.year,
        period: props.period,
        ...form.value,
        timeliness_percentage: timelinessPercentage.value
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
</script>

<template>
  <UModal v-model="isOpen">
    <UCard :ui="{ ring: '', divide: 'divide-y divide-gray-100 dark:divide-gray-800' }">
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-base font-semibold leading-6 text-gray-900 dark:text-white">
            Report Timeliness Questionnaire ({{ period }} {{ year }})
          </h3>
          <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" class="-my-1" @click="isOpen = false" />
        </div>
      </template>

      <div class="space-y-4" v-if="!isLoading">
        <UFormField label="Total Reports Planned" required>
          <UInput type="number" v-model.number="form.total_reports_planned" min="0" />
        </UFormField>

        <UFormField label="Total Reports Completed" required>
          <UInput type="number" v-model.number="form.total_reports_completed" min="0" />
        </UFormField>

        <UFormField label="Reports Completed On Time" required>
          <UInput type="number" v-model.number="form.reports_completed_on_time" min="0" :max="form.total_reports_completed" />
        </UFormField>

        <UFormField label="Timeliness Percentage">
          <div class="flex items-center gap-2">
            <UProgress :value="timelinessPercentage" max="100" class="flex-1" :color="timelinessPercentage >= 90 ? 'success' : timelinessPercentage >= 70 ? 'warning' : 'error'" />
            <span class="font-bold w-12 text-right">{{ timelinessPercentage }}%</span>
          </div>
        </UFormField>

        <UFormField label="Remarks">
          <UTextarea v-model="form.remarks" rows="3" placeholder="Additional notes or context..." />
        </UFormField>
      </div>
      
      <div v-else class="flex justify-center p-8">
        <UIcon name="i-lucide-loader-2" class="w-8 h-8 animate-spin text-primary" />
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <UButton color="gray" variant="soft" @click="isOpen = false" :disabled="isLoading">Cancel</UButton>
          <UButton color="primary" @click="handleSave" :loading="isLoading">Save Metrics</UButton>
        </div>
      </template>
    </UCard>
  </UModal>
</template>
