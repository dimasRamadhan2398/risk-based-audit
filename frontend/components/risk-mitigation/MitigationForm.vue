<template>
  <UModal v-model:open="store.isFormOpen" :dismissible="false" :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }">
    
    <div></div>

    <template #content>
      <UForm :state="store.form" @submit.prevent="() => store.handleSubmit(props.currentRiskId)">
        <div class="relative rounded-xl shadow-2xl flex flex-col max-h-[90vh]">
          
          <div class="px-6 py-4 rounded-t-xl flex justify-between items-center">
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ store.isEditing ? 'Edit Mitigation & Control Plan' : 'Tambah Mitigation & Control Plan' }}
            </h3>
            <UIcon name="close" @click="store.closeForm" class="text-[var(--text-muted)] hover:text-[var(--text-main)] text-2xl cursor-pointer">&times;</UIcon>
          </div>

          <div class="p-6 overflow-y-auto space-y-5">
            

            <UFormField label="Mitigation Plan & Controls" required>
              <UTextarea v-model="store.form.mitigationPlan" :rows="4" autoresize placeholder="Explain the mitigation plan and controls in detail..." class="w-full" required />
            </UFormField>

            <UFormField label="Supervisor" required>
              <USelectMenu 
              :items="store.supervisorOptions" 
              v-model="store.form.supervisor" 
              placeholder="Enter supervisor's name..." 
              class="w-full" 
              required />
            </UFormField>
            
            <UFormField label="PIC" required>
              <USelectMenu 
              :items="store.picOptions" 
              v-model="store.form.pic" 
              placeholder="Enter PIC's name..." 
              class="w-full" 
              required />
            </UFormField>

            <UFormField label="Unit In Charge" required>
              <USelectMenu 
              :items="store.unitInChargeOptions" 
              v-model="store.form.unitInCharge" 
              placeholder="Enter unit in charge..." 
              class="w-full" 
              required />
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField label="Start Date" required>
                    <UInput v-model="store.form.start_date" type="date" :min="todayDate" icon="i-heroicons-calendar" required />
                </UFormField>

                <UFormField label="End Date" required>
                    <UInput v-model="store.form.end_date" type="date" :min="minEndDate" icon="i-heroicons-calendar-days" required />
                </UFormField>
            </div>

            <UFormField label="Additional Notes">
              <UTextarea v-model="store.form.notes" :rows="2" placeholder="Optional notes..." class="w-full" />
            </UFormField>

          </div>
      
          <div class="px-6 py-4 rounded-b-xl flex justify-end gap-3">
            <UButton label="Cancel" color="neutral" variant="outline" @click="store.closeForm" />
            <UButton :label="store.isEditing ? 'Save Changes' : 'Save Mitigation'" color="primary" type="submit" />
          </div>
      
        </div>
      </UForm>  
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useMitigationStore } from '~/stores/mitigation-risk'

// Menerima Risk ID dari halaman induk agar mitigasi tersambung dengan risiko yang tepat
const props = defineProps<{
  currentRiskId: string
}>()

const store = useMitigationStore()

// Mendapatkan tanggal hari ini dalam format YYYY-MM-DD untuk batas minimum pemilihan tanggal
const todayDate = computed(() => {
  const today = new Date()
  // Adjust timezone offset to get the correct local date string
  const offset = today.getTimezoneOffset()
  const localDate = new Date(today.getTime() - (offset * 60 * 1000))
  return localDate.toISOString().split('T')[0]
})

// Menghitung batas minimum untuk End Date (minimal 7 hari setelah Start Date)
const minEndDate = computed(() => {
  if (store.form.start_date) {
    const startDate = new Date(store.form.start_date)
    startDate.setDate(startDate.getDate() + 7)
    const offset = startDate.getTimezoneOffset()
    const localDate = new Date(startDate.getTime() - (offset * 60 * 1000))
    return localDate.toISOString().split('T')[0]
  }
  return todayDate.value
})
</script>